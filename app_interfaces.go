package main

import (
	"plexcord/internal/discord"
	"plexcord/internal/plex"
)

// This file defines the interfaces that the App depends on. By depending on
// these interfaces instead of concrete types, the App becomes unit-testable
// with in-memory fakes and the concrete implementations (plex.Client,
// discord.PresenceManager, keychain package) can be swapped without touching
// the Wails binding layer.
//
// Per Go convention, these interfaces live on the consumer side (package main)
// so the producer packages do not need to know about them.

// PlexAPI abstracts the Plex server client used by App methods. It covers
// the operations needed by validation, user selection, and session polling.
type PlexAPI interface {
	ValidateConnection() (*plex.ValidationResult, error)
	GetUsers() ([]plex.PlexUser, error)
	GetMusicSessions(userID string) ([]plex.MusicSession, error)
	GetMediaSessions(userID string, mediaTypes []string) ([]plex.MediaSession, error)
}

// PlexAPIFactory constructs a PlexAPI for a given token and server URL.
// Using a factory (instead of a singleton PlexAPI) reflects the reality
// that the Plex client is per-server and may change at runtime when the
// user switches servers.
type PlexAPIFactory func(token, serverURL string) PlexAPI

// DiscordPresence abstracts the Discord Rich Presence manager.
type DiscordPresence interface {
	Connect(clientID string) error
	Disconnect() error
	IsConnected() bool
	GetClientID() string
	SetPresence(data *discord.PresenceData) error
	ClearPresence() error
	UpdatePresenceFromPlayback(
		track, artist, album, state string,
		duration, position int64,
		artworkURL, player, detailsFormat, stateFormat string,
	) error
}

// TokenStore abstracts credential persistence. The production implementation
// is backed by the OS keychain; tests can inject a map-based fake.
type TokenStore interface {
	Get() (string, error)
	Set(token string) error
	Delete() error
}
