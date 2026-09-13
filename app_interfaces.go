package main

import (
	"context"
	"time"

	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/history"
	"plexcord/internal/platform"
	"plexcord/internal/plex"
	"plexcord/internal/retry"
	"plexcord/internal/updater"
	"plexcord/internal/version"
)

// This file defines the interfaces that the App depends on. By depending on
// these interfaces instead of concrete types, the App becomes unit-testable
// with in-memory fakes and the concrete implementations (plex.Client,
// discord.PresenceManager, keychain package, system tray, updater, …) can be
// swapped without touching the Wails binding layer.
//
// Per Go convention, these interfaces live on the consumer side (package main)
// so the producer packages do not need to know about them. They are kept
// deliberately narrow — a caller should be able to satisfy only what it uses
// (ISP) — and composed where a consumer genuinely needs several of them.

// ----------------------------------------------------------------------------
// Plex
// ----------------------------------------------------------------------------

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

// ServerDiscoverer finds Plex servers on the local network. Abstracted so the
// setup wizard's discovery step can be exercised without multicast traffic.
type ServerDiscoverer interface {
	Discover(timeout time.Duration) ([]plex.Server, error)
}

// ----------------------------------------------------------------------------
// Discord
// ----------------------------------------------------------------------------

// DiscordConnection is the connection half of Discord Rich Presence: opening,
// closing, and reporting on the IPC link.
type DiscordConnection interface {
	Connect(clientID string) error
	Disconnect() error
	IsConnected() bool
	GetClientID() string
}

// PresenceWriter is the publishing half: what is currently shown on the user's
// Discord profile. Callers that only publish presence (the session observer)
// depend on this alone.
type PresenceWriter interface {
	// SetPresence publishes pre-built presence data, used by the connection test.
	SetPresence(data *discord.PresenceData) error
	// ClearPresence removes the presence without disconnecting.
	ClearPresence() error
	// UpdatePlayback publishes a playback snapshot with the user's display
	// options applied.
	UpdatePlayback(playback discord.Playback, opts discord.Options) error
}

// DiscordPresence is the full surface the App holds: a connection that can
// also publish presence. Narrower consumers take DiscordConnection or
// PresenceWriter instead.
type DiscordPresence interface {
	DiscordConnection
	PresenceWriter
}

// ArtworkResolver resolves a publicly reachable album-art URL for a track so
// covers render on Discord without leaking the Plex token. The production
// implementation is *artwork.Resolver; tests can inject a fake.
type ArtworkResolver interface {
	// Cached returns a previously resolved URL without any network request.
	Cached(artist, album string) (string, bool)
	// Resolve returns a public HTTPS artwork URL, or "" if none is found.
	Resolve(ctx context.Context, artist, album string) (string, error)
}

// ----------------------------------------------------------------------------
// Storage
// ----------------------------------------------------------------------------

// TokenStore abstracts credential persistence. The production implementation
// is backed by the OS keychain; tests can inject a map-based fake.
type TokenStore interface {
	Get() (string, error)
	Set(token string) error
	Delete() error
}

// ConfigGateway abstracts the configuration file itself: where it lives, and
// loading, deleting, and reading the setup state from it. Persisting changes
// to an already-loaded Config goes through config.Store, not through here.
//
// App previously called the config package's top-level functions directly,
// which hard-wired startup, reset and the setup check to a real file in the
// user's config directory.
type ConfigGateway interface {
	Load() (*config.Config, error)
	Delete() error
	IsSetupComplete() bool
	ConfigDir() string
}

// HistoryStore abstracts listening-history persistence.
type HistoryStore interface {
	Add(entry history.Entry)
	GetRecent(limit int) []history.Entry
	GetStats() history.Stats
	Clear()
}

// ----------------------------------------------------------------------------
// Platform integration
// ----------------------------------------------------------------------------

// TrayController abstracts the system tray, so the update-notice wiring can be
// verified without a desktop session or a DBus/StatusNotifierItem host.
type TrayController interface {
	Start()
	Stop()
	SetUpdateNotice(notice platform.UpdateNotice)
}

// AutoStartController abstracts launch-on-login registration, which otherwise
// writes to the Windows registry, a LaunchAgent plist, or an XDG desktop file.
type AutoStartController interface {
	IsEnabled() bool
	SetEnabled(enabled bool) error
	Disable() error
	EnsureRegistered() error
}

// UpdateService abstracts the automatic update checker and the single download
// path it shares with the user-initiated install.
type UpdateService interface {
	StartChecker(ctx context.Context)
	StopChecker()
	GetStatus() updater.Status
	OnStatusChange(fn func(updater.Status))
	Check() (*version.UpdateInfo, error)
	StartDownload(ctx context.Context, auto bool) (*version.UpdateInfo, error)
}

// AppRelauncher starts a fresh copy of the application binary, for applying an
// update that has already been written to disk. Abstracted so the restart path
// — which otherwise spawns a real process and quits this one — can be verified.
type AppRelauncher interface {
	// Relaunch spawns the updated binary and returns once it has started.
	Relaunch() error
}

// ----------------------------------------------------------------------------
// Reconnection
// ----------------------------------------------------------------------------

// RetryManager abstracts the automatic reconnection loop for one service.
// App drives two of them (Plex and Discord) and only ever uses the calls
// below, so the interface stops there rather than mirroring the whole manager.
type RetryManager interface {
	SetCallbacks(retry retry.RetryCallback, stateChange retry.StateChangeCallback)
	Start(err error, code string)
	Stop()
	Reset()
	ManualRetry()
	GetState() retry.RetryState
}

// ----------------------------------------------------------------------------
// Plex PIN authentication
// ----------------------------------------------------------------------------

// PlexAuthenticator abstracts the plex.tv PIN link flow: request a PIN, build
// the URL the user visits, then poll until they authorize it. One authenticator
// serves one PIN's lifecycle, which is why App builds them through a factory.
type PlexAuthenticator interface {
	RequestPIN(ctx context.Context) (*plex.PINResponse, error)
	CheckPIN(ctx context.Context, pinID int) (*plex.PINResponse, error)
	GetAuthURL(pinCode string) string
}

// PlexAuthenticatorFactory constructs an authenticator for a fresh PIN request.
// A new one per request is deliberate: each authentication session gets its own
// client ID.
type PlexAuthenticatorFactory func() PlexAuthenticator
