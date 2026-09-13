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
		ClientID:  a.discord.ClientID(),
	})

	log.Printf("Discord connected successfully")
	return nil
}

// DisconnectDiscord closes the connection to Discord.
// Clears any active presence before disconnecting.
// Emits DiscordDisconnected Wails event.
func (a *App) DisconnectDiscord() error {
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
	return a.discord.IsConnected()
}

// GetDefaultDiscordClientID returns the default PlexCord Discord Application Client ID.
func (a *App) GetDefaultDiscordClientID() string {
	return discord.DefaultClientID
}

// GetDiscordClientID returns the currently configured Discord Client ID.
// Returns the config value if set, otherwise the default.
func (a *App) GetDiscordClientID() string {
	return a.effectiveDiscordClientID()
}

// presenceDisplayOptions projects the user's persisted presence preferences
// onto the options the Discord layer consumes. Having one place build them
// keeps every presence path — session updates, artwork re-issues, the manual
// binding — telling Discord the same story.
func (a *App) presenceDisplayOptions() discord.Options {
	if a.config == nil {
		return discord.Options{}
	}
	return discord.Options{
		DetailsFormat: a.config.PresenceDetailsFormat,
		StateFormat:   a.config.PresenceStateFormat,
		ActivityStyle: a.config.PresenceActivityStyle,
		StatusDisplay: a.config.PresenceStatusDisplay,
	}
}

// effectiveDiscordClientID is the Client ID PlexCord actually connects with:
// the configured one, or the built-in PlexCord application. It is the single
// answer to that question, shared by the binding and by the silent reconnect.
func (a *App) effectiveDiscordClientID() string {
	if a.config != nil && a.config.DiscordClientID != "" {
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
	session := &plex.MusicSession{Track: track, Artist: artist, Album: album, Duration: duration, ViewOffset: position}
	session.State = state

	if !a.discord.IsConnected() {
		return errors.New(errors.DISCORD_CONN_FAILED, "not connected to Discord")
	}
	a.discord.Publish(session)
	return nil
}

// ClearDiscordPresence removes the Discord Rich Presence.
// Called when playback stops.
func (a *App) ClearDiscordPresence() error {
	return a.discord.Clear()
}

// TestDiscordPresence sends a test presence message to Discord to verify the connection.
// This displays a sample "Now Playing" message on the user's Discord profile.
// Returns an error if not connected or if the test fails.
func (a *App) TestDiscordPresence() error {
	log.Printf("Sending test presence to Discord...")

	// Create test presence data
	testPresence := &discord.PresenceData{
		Track:  "Test Song - PlexCord",
		Artist: "PlexCord Test",
		Album:  "Connection Test",
		State:  "playing",
	}

	connected, err := a.discord.SetPresence(testPresence)
	if !connected {
		return errors.New(errors.DISCORD_CONN_FAILED, "not connected to Discord")
	}
	if err != nil {
		log.Printf("ERROR: Failed to send test presence: %v", err)
		return err
	}

	log.Printf("Test presence sent successfully")
	return nil
}

// updateDiscordFromSession publishes a music session to Discord. The
// mechanics — cached-vs-background artwork, silent reconnect, the generation
// guard on a late cover — live in discordService.
func (a *App) updateDiscordFromSession(session *plex.MusicSession) {
	a.discord.Publish(session)
}

// clearDiscordOnStop clears Discord Rich Presence when playback stops. It runs
// on the poll path, where there is nobody to report a failure to, so a failed
// clear is logged and the next update overwrites the stale presence anyway.
func (a *App) clearDiscordOnStop() {
	if err := a.discord.Clear(); err != nil {
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
	paused := a.presence.Toggle()

	if paused {
		log.Printf("Presence manually paused")
		// Clear current presence immediately
		a.clearDiscordOnStop()
	} else {
		log.Printf("Presence manually resumed")
		// Restore presence from current session if available
		if session := a.sessions.Get(); session != nil {
			a.updateDiscordFromSession(session)
		}
	}

	return paused
}

// IsPresencePaused returns whether presence updates are manually paused.
func (a *App) IsPresencePaused() bool {
	return a.presence.IsPaused()
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

// ============================================================================
// Presence Display Options (activity style, member-list line, artwork lookup)
// ============================================================================

// PresenceOptions represents the presence display configuration for the frontend.
type PresenceOptions struct {
	ActivityStyle string `json:"activityStyle"` // "media" | "game"
	StatusDisplay string `json:"statusDisplay"` // "app" | "state" | "details"
	ArtworkLookup bool   `json:"artworkLookup"`
}

// GetPresenceOptions returns the current presence display options, normalizing
// unset values to their defaults so the frontend always has a concrete choice.
func (a *App) GetPresenceOptions() PresenceOptions {
	style := a.config.PresenceActivityStyle
	if style == "" {
		style = discord.ActivityStyleMedia
	}
	display := a.config.PresenceStatusDisplay
	if display == "" {
		display = discord.StatusDisplayState
	}
	return PresenceOptions{
		ActivityStyle: style,
		StatusDisplay: display,
		ArtworkLookup: a.config.ArtworkLookupEnabled(),
	}
}

// SetPresenceOptions updates the presence display options. Invalid values are
// rejected so a bad frontend value cannot corrupt the presence output. The new
// options take effect on the next presence update — no reconnect required.
func (a *App) SetPresenceOptions(opts PresenceOptions) error {
	switch opts.ActivityStyle {
	case discord.ActivityStyleMedia, discord.ActivityStyleGame:
	default:
		return errors.New(errors.CONFIG_WRITE_FAILED, "invalid activity style: "+opts.ActivityStyle)
	}
	switch opts.StatusDisplay {
	case discord.StatusDisplayApp, discord.StatusDisplayState, discord.StatusDisplayDetails:
	default:
		return errors.New(errors.CONFIG_WRITE_FAILED, "invalid status display: "+opts.StatusDisplay)
	}

	a.config.PresenceActivityStyle = opts.ActivityStyle
	a.config.PresenceStatusDisplay = opts.StatusDisplay
	lookup := opts.ArtworkLookup
	a.config.PresenceArtworkLookup = &lookup

	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save presence options: %v", err)
		return err
	}
	log.Printf("Presence options updated: style=%s, display=%s, artwork=%v",
		opts.ActivityStyle, opts.StatusDisplay, opts.ArtworkLookup)
	return nil
}
