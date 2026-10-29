package main

import (
	"context"
	"time"

	"plexcord/internal/artwork"
	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/history"
	"plexcord/internal/platform"
	"plexcord/internal/plex"
	"plexcord/internal/retry"
	"plexcord/internal/updater"
	"plexcord/internal/version"
)

type PlexAPI interface {
	ValidateConnection() (*plex.ValidationResult, error)
	GetUsers() ([]plex.PlexUser, error)
	GetMusicSessions(userID string) ([]plex.MusicSession, error)
	GetMediaSessions(userID string, mediaTypes []string) ([]plex.MediaSession, error)
}

type PlexAPIFactory func(token, serverURL string) PlexAPI

type ServerDiscoverer interface {
	Discover(timeout time.Duration) ([]plex.Server, error)
}

type DiscordConnection interface {
	Connect(clientID string) error
	Disconnect() error
	IsConnected() bool
	GetClientID() string
}

type PresenceWriter interface {
	SetPresence(data *discord.PresenceData) error
	ClearPresence() error
	UpdatePlayback(playback discord.Playback, opts discord.Options) error
}

type DiscordPresence interface {
	DiscordConnection
	PresenceWriter
}

type ArtworkResolver interface {
	Cached(q artwork.Query) (string, bool)
	Resolve(ctx context.Context, q artwork.Query) (string, error)
}

type TokenStore interface {
	Get() (string, error)
	Set(token string) error
	Delete() error
}

type ConfigGateway interface {
	Load() (*config.Config, error)
	Delete() error
	IsSetupComplete() bool
	ConfigDir() string
}

type HistoryStore interface {
	Add(entry history.Entry)
	GetRecent(limit int) []history.Entry
	GetStats() history.Stats
	Clear()
}

type TrayController interface {
	Start()
	Stop()
	SetUpdateNotice(notice platform.UpdateNotice)
}

type AutoStartController interface {
	IsEnabled() bool
	SetEnabled(enabled bool) error
	Disable() error
	EnsureRegistered() error
}

type UpdateService interface {
	StartChecker(ctx context.Context)
	StopChecker()
	GetStatus() updater.Status
	OnStatusChange(fn func(updater.Status))
	Check() (*version.UpdateInfo, error)
	StartDownload(ctx context.Context, auto bool) (*version.UpdateInfo, error)
}

type AppRelauncher interface {
	Relaunch() error
}

type RetryManager interface {
	SetCallbacks(retry retry.RetryCallback, stateChange retry.StateChangeCallback)
	Start(err error, code string)
	Stop()
	Reset()
	ManualRetry()
	GetState() retry.RetryState
}

type PlexAuthenticator interface {
	RequestPIN(ctx context.Context) (*plex.PINResponse, error)
	CheckPIN(ctx context.Context, pinID int) (*plex.PINResponse, error)
	GetAuthURL(pinCode string) string
}

type PlexAuthenticatorFactory func() PlexAuthenticator
