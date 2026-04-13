package main

import (
	"log"
	"time"

	"plexcord/internal/discord"
	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/plex"
)

// ConnectDiscord establishes a connection to Discord using the provided Client ID.
// If clientID is empty, the default PlexCord Client ID is used.
// Returns an error if connection fails (e.g., Discord not running).
// Emits DiscordConnected or DiscordDisconnected Wails event based on result.
func (a *App) ConnectDiscord(clientID string) error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	log.Printf("Attempting Discord connection...")

	// Use configured client ID if not provided
	if clientID == "" {
		clientID = a.config.DiscordClientID
	}

	err := a.discord.Connect(clientID)
	if err != nil {
		log.Printf("ERROR: Discord connection failed: %v", err)
		// Emit disconnected event with error
		a.bus.Emit(events.DiscordDisconnected, discord.ConnectionEvent{
			Connected: false,
			Error: &discord.Error{
				Code:    errors.GetCode(err),
				Message: err.Error(),
			},
		})
		return err
	}

	// Save client ID to config if connection successful
	if clientID != "" && clientID != a.config.DiscordClientID {
		a.config.DiscordClientID = clientID
		if err := a.saveConfig(); err != nil {
			log.Printf("Warning: Failed to save Discord client ID to config: %v", err)
		}
	}

	// Update connection history
	a.updateDiscordConnectionTime()

	// Stop any pending retries
	a.stopDiscordRetry()

	// Emit connected event
	a.bus.Emit(events.DiscordConnected, discord.ConnectionEvent{
		Connected: true,
		ClientID:  a.discord.GetClientID(),
	})

	log.Printf("Discord connected successfully")
	return nil
}

// DisconnectDiscord closes the connection to Discord.
// Clears any active presence before disconnecting.
// Emits DiscordDisconnected Wails event.
func (a *App) DisconnectDiscord() error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	log.Printf("Disconnecting from Discord...")

	err := a.discord.Disconnect()
	if err != nil {
		log.Printf("ERROR: Discord disconnect failed: %v", err)
		return err
	}

	// Emit disconnected event
	a.bus.Emit(events.DiscordDisconnected, discord.ConnectionEvent{
		Connected: false,
	})

	log.Printf("Discord disconnected")
	return nil
}

// IsDiscordConnected returns whether a Discord connection is active.
func (a *App) IsDiscordConnected() bool {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()
	return a.discord.IsConnected()
}

// GetDefaultDiscordClientID returns the default PlexCord Discord Application Client ID.
func (a *App) GetDefaultDiscordClientID() string {
	return discord.DefaultClientID
}

// GetDiscordClientID returns the currently configured Discord Client ID.
// Returns the config value if set, otherwise the default.
func (a *App) GetDiscordClientID() string {
	if a.config.DiscordClientID != "" {
		return a.config.DiscordClientID
	}
	return discord.DefaultClientID
}

// ValidateDiscordClientID validates a Discord Client ID format.
// Returns nil if valid, or an error with code DISCORD_CLIENT_ID_INVALID.
// Empty string is valid (means "use default Client ID").
func (a *App) ValidateDiscordClientID(clientID string) error {
	return discord.ValidateClientID(clientID)
}

// SaveDiscordClientID saves a custom Discord Client ID to configuration.
// Pass empty string to reset to default.
// Returns an error if the Client ID format is invalid.
func (a *App) SaveDiscordClientID(clientID string) error {
	// Validate before saving
	if err := discord.ValidateClientID(clientID); err != nil {
		log.Printf("ERROR: Invalid Discord client ID: %v", err)
		return err
	}

	a.config.DiscordClientID = clientID
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save Discord client ID: %v", err)
		return err
	}
	log.Printf("Discord Client ID saved")
	return nil
}

// UpdateDiscordPresence updates the Discord Rich Presence with current playback info.
// This is called internally when playback state changes.
func (a *App) UpdateDiscordPresence(track, artist, album, state string, duration, position int64) error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	if !a.discord.IsConnected() {
		return errors.New(errors.DISCORD_CONN_FAILED, "not connected to Discord")
	}

	return a.discord.UpdatePresenceFromPlayback(track, artist, album, state, duration, position, "", "", a.config.PresenceDetailsFormat, a.config.PresenceStateFormat)
}

// ClearDiscordPresence removes the Discord Rich Presence.
// Called when playback stops.
func (a *App) ClearDiscordPresence() error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	if !a.discord.IsConnected() {
		return nil // Not connected, nothing to clear
	}

	return a.discord.ClearPresence()
}

// TestDiscordPresence sends a test presence message to Discord to verify the connection.
// This displays a sample "Now Playing" message on the user's Discord profile.
// Returns an error if not connected or if the test fails.
func (a *App) TestDiscordPresence() error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	if !a.discord.IsConnected() {
		return errors.New(errors.DISCORD_CONN_FAILED, "not connected to Discord")
	}

	log.Printf("Sending test presence to Discord...")

	// Create test presence data
	testPresence := &discord.PresenceData{
		Track:  "Test Song - PlexCord",
		Artist: "PlexCord Test",
		Album:  "Connection Test",
		State:  "playing",
	}

	err := a.discord.SetPresence(testPresence)
	if err != nil {
		log.Printf("ERROR: Failed to send test presence: %v", err)
		return err
	}

	log.Printf("Test presence sent successfully")
	return nil
}

// updateDiscordFromSession updates Discord Rich Presence with music session info.
// This is called automatically when playback is detected by the poller.
// If Discord is not connected, attempts to reconnect automatically (Story 3-8).
func (a *App) updateDiscordFromSession(session *plex.MusicSession) {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	// If not connected, try to reconnect (auto-recovery for Discord restart)
	if !a.discord.IsConnected() {
		a.tryDiscordReconnect()
		// If still not connected after reconnect attempt, skip update
		if !a.discord.IsConnected() {
			return
		}
		log.Printf("Discord: Reconnected - restoring presence")
	}

	// Update presence with session data, including artwork URL and format strings
	err := a.discord.UpdatePresenceFromPlayback(
		session.Track,
		session.Artist,
		session.Album,
		session.State,
		session.Duration,
		session.ViewOffset,
		session.ThumbURL,
		session.PlayerName,
		a.config.PresenceDetailsFormat,
		a.config.PresenceStateFormat,
	)
	if err != nil {
		log.Printf("Warning: Failed to update Discord presence: %v", err)
	}
}

// tryDiscordReconnect attempts to reconnect to Discord.
// This is called automatically when music is playing but Discord is not connected.
// Uses the configured or default Client ID for reconnection.
// Does not emit events - this is a silent background reconnection.
func (a *App) tryDiscordReconnect() {
	// Get the client ID to use
	clientID := a.config.DiscordClientID
	if clientID == "" {
		clientID = discord.DefaultClientID
	}

	// Attempt to reconnect (silently - no event emission)
	err := a.discord.Connect(clientID)
	if err != nil {
		// Failed to reconnect - Discord probably still not running
		// This is expected, don't log as error
		return
	}

	// Successfully reconnected - update connection history
	a.updateDiscordConnectionTime()
	log.Printf("Discord: Auto-reconnected to Discord")
}

// clearDiscordOnStop clears Discord Rich Presence when playback stops.
// This is called automatically when playback ends.
// If Discord is not connected, the clear is silently skipped.
func (a *App) clearDiscordOnStop() {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	if !a.discord.IsConnected() {
		// Not connected to Discord - nothing to clear
		return
	}

	err := a.discord.ClearPresence()
	if err != nil {
		log.Printf("Warning: Failed to clear Discord presence: %v", err)
	}
}

// updateDiscordConnectionTime updates the last successful Discord connection time.
// Called internally when Discord connection is established.
func (a *App) updateDiscordConnectionTime() {
	now := time.Now()
	a.config.DiscordLastConnected = &now
	if err := a.saveConfig(); err != nil {
		log.Printf("Warning: Failed to save Discord connection time: %v", err)
	}
}

// ============================================================================
// Presence Pause Toggle & Hide When Paused
// ============================================================================

// TogglePresencePause toggles the manual presence pause state.
// When paused, all Discord presence updates are skipped.
// Returns the new paused state.
func (a *App) TogglePresencePause() bool {
	a.pauseMu.Lock()
	a.presencePaused = !a.presencePaused
	paused := a.presencePaused
	a.pauseMu.Unlock()

	if paused {
		log.Printf("Presence manually paused")
		// Clear current presence immediately
		a.clearDiscordOnStop()
	} else {
		log.Printf("Presence manually resumed")
		// Restore presence from current session if available
		a.sessionMu.RLock()
		session := a.currentSession
		a.sessionMu.RUnlock()
		if session != nil {
			a.updateDiscordFromSession(session)
		}
	}

	return paused
}

// IsPresencePaused returns whether presence updates are manually paused.
func (a *App) IsPresencePaused() bool {
	a.pauseMu.Lock()
	defer a.pauseMu.Unlock()
	return a.presencePaused
}

// GetHideWhenPaused returns the hide-when-paused settings.
func (a *App) GetHideWhenPaused() map[string]interface{} {
	return map[string]interface{}{
		"enabled":      a.config.HideWhenPaused,
		"delaySeconds": a.config.HideWhenPausedDelay,
	}
}

// SetHideWhenPaused updates the hide-when-paused settings.
func (a *App) SetHideWhenPaused(enabled bool, delaySeconds int) error {
	if delaySeconds < 0 {
		delaySeconds = 0
	}
	a.config.HideWhenPaused = enabled
	a.config.HideWhenPausedDelay = delaySeconds
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save hide-when-paused settings: %v", err)
		return err
	}
	log.Printf("Hide when paused set to: enabled=%v, delay=%d seconds", enabled, delaySeconds)
	return nil
}

// scheduleHideOnPause schedules clearing Discord presence after the configured delay.
func (a *App) scheduleHideOnPause() {
	a.pauseMu.Lock()
	defer a.pauseMu.Unlock()

	// Cancel existing timer if any
	if a.pauseTimer != nil {
		a.pauseTimer.Stop()
		a.pauseTimer = nil
	}

	delay := time.Duration(a.config.HideWhenPausedDelay) * time.Second
	if delay <= 0 {
		// Immediate clear
		go a.clearDiscordOnStop()
		return
	}

	a.pauseTimer = time.AfterFunc(delay, func() {
		log.Printf("Hide-when-paused delay elapsed, clearing presence")
		a.clearDiscordOnStop()
	})
}

// cancelPauseTimer cancels any pending hide-when-paused timer.
func (a *App) cancelPauseTimer() {
	a.pauseMu.Lock()
	defer a.pauseMu.Unlock()

	if a.pauseTimer != nil {
		a.pauseTimer.Stop()
		a.pauseTimer = nil
	}
}

// ============================================================================
// Custom Presence Format Strings
// ============================================================================

// PresenceFormatSettings represents the presence format configuration for the frontend.
type PresenceFormatSettings struct {
	DetailsFormat string `json:"detailsFormat"`
	StateFormat   string `json:"stateFormat"`
}

// GetPresenceFormat returns the current presence format strings.
func (a *App) GetPresenceFormat() PresenceFormatSettings {
	return PresenceFormatSettings{
		DetailsFormat: a.config.PresenceDetailsFormat,
		StateFormat:   a.config.PresenceStateFormat,
	}
}

// SetPresenceFormat updates the presence format strings.
// Pass empty strings to reset to defaults.
func (a *App) SetPresenceFormat(details, state string) error {
	a.config.PresenceDetailsFormat = details
	a.config.PresenceStateFormat = state
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save presence format: %v", err)
		return err
	}
	log.Printf("Presence format updated: details=%q, state=%q", details, state)
	return nil
}
