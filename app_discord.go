package main

import (
	"log"
	"time"

	"plexcord/internal/discord"
	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/plex"
)

func (a *App) ConnectDiscord(clientID string) error {
	log.Printf("Attempting Discord connection...")

	if clientID == "" {
		clientID = a.config.DiscordClientID
	}

	err := a.discord.Connect(clientID)
	if err != nil {
		log.Printf("ERROR: Discord connection failed: %v", err)
		a.bus.Emit(events.DiscordDisconnected, discord.ConnectionEvent{
			Connected: false,
			Error: &discord.Error{
				Code:    errors.GetCode(err),
				Message: err.Error(),
			},
		})
		return err
	}

	if clientID != "" && clientID != a.config.DiscordClientID {
		a.config.DiscordClientID = clientID
		if err := a.saveConfig(); err != nil {
			log.Printf("Warning: Failed to save Discord client ID to config: %v", err)
		}
	}

	a.updateDiscordConnectionTime()

	a.stopDiscordRetry()

	a.bus.Emit(events.DiscordConnected, discord.ConnectionEvent{
		Connected: true,
		ClientID:  a.discord.ClientID(),
	})

	log.Printf("Discord connected successfully")
	return nil
}

func (a *App) DisconnectDiscord() error {
	log.Printf("Disconnecting from Discord...")

	err := a.discord.Disconnect()
	if err != nil {
		log.Printf("ERROR: Discord disconnect failed: %v", err)
		return err
	}

	a.bus.Emit(events.DiscordDisconnected, discord.ConnectionEvent{
		Connected: false,
	})

	log.Printf("Discord disconnected")
	return nil
}

func (a *App) IsDiscordConnected() bool {
	return a.discord.IsConnected()
}

func (a *App) GetDefaultDiscordClientID() string {
	return discord.DefaultClientID
}

func (a *App) GetDiscordClientID() string {
	return a.effectiveDiscordClientID()
}

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

func (a *App) effectiveDiscordClientID() string {
	if a.config != nil && a.config.DiscordClientID != "" {
		return a.config.DiscordClientID
	}
	return discord.DefaultClientID
}

func (a *App) ValidateDiscordClientID(clientID string) error {
	return discord.ValidateClientID(clientID)
}

func (a *App) SaveDiscordClientID(clientID string) error {
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

func (a *App) UpdateDiscordPresence(track, artist, album, state string, duration, position int64) error {
	session := &plex.MediaSession{
		MediaType:  plex.MediaTypeMusic,
		Title:      track,
		Artist:     artist,
		Album:      album,
		State:      state,
		Duration:   duration,
		ViewOffset: position,
	}

	if !a.discord.IsConnected() {
		return errors.New(errors.DISCORD_CONN_FAILED, "not connected to Discord")
	}
	a.discord.Publish(session)
	return nil
}

func (a *App) ClearDiscordPresence() error {
	return a.discord.Clear()
}

func (a *App) TestDiscordPresence() error {
	log.Printf("Sending test presence to Discord...")

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

func (a *App) updateDiscordFromSession(session *plex.MediaSession) {
	a.discord.Publish(session)
}

func (a *App) clearDiscordOnStop() {
	if err := a.discord.Clear(); err != nil {
		log.Printf("Warning: Failed to clear Discord presence: %v", err)
	}
}

func (a *App) updateDiscordConnectionTime() {
	now := time.Now()
	a.config.DiscordLastConnected = &now
	if err := a.saveConfig(); err != nil {
		log.Printf("Warning: Failed to save Discord connection time: %v", err)
	}
}

func (a *App) TogglePresencePause() bool {
	paused := a.presence.Toggle()

	if paused {
		log.Printf("Presence manually paused")
		a.clearDiscordOnStop()
	} else {
		log.Printf("Presence manually resumed")
		if session := a.sessions.Get(); session != nil {
			a.updateDiscordFromSession(session)
		}
	}

	return paused
}

func (a *App) IsPresencePaused() bool {
	return a.presence.IsPaused()
}

func (a *App) GetHideWhenPaused() map[string]interface{} {
	return map[string]interface{}{
		"enabled":      a.config.HideWhenPaused,
		"delaySeconds": a.config.HideWhenPausedDelay,
	}
}

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

type PresenceFormatSettings struct {
	DetailsFormat string `json:"detailsFormat"`
	StateFormat   string `json:"stateFormat"`
}

func (a *App) GetPresenceFormat() PresenceFormatSettings {
	return PresenceFormatSettings{
		DetailsFormat: a.config.PresenceDetailsFormat,
		StateFormat:   a.config.PresenceStateFormat,
	}
}

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

type PresenceOptions struct {
	ActivityStyle string `json:"activityStyle"`
	StatusDisplay string `json:"statusDisplay"`
	ArtworkLookup bool   `json:"artworkLookup"`
}

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
