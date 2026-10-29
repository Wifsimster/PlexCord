package main

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"time"

	"plexcord/internal/artwork"
	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/errors"
	"plexcord/internal/history"
	"plexcord/internal/keychain"
	"plexcord/internal/platform"
	"plexcord/internal/plex"
	"plexcord/internal/retry"
	"plexcord/internal/updater"
	"plexcord/internal/version"
)

type keychainTokenStore struct{}

func (keychainTokenStore) Get() (string, error)   { return keychain.GetToken() }
func (keychainTokenStore) Set(token string) error { return keychain.SetToken(token) }
func (keychainTokenStore) Delete() error          { return keychain.DeleteToken() }

func newKeychainTokenStore() TokenStore {
	return keychainTokenStore{}
}

func newPlexClientFactory() PlexAPIFactory {
	return func(token, serverURL string) PlexAPI {
		return plex.NewClient(token, serverURL)
	}
}

type gdmDiscoverer struct{}

func (gdmDiscoverer) Discover(timeout time.Duration) ([]plex.Server, error) {
	return plex.DiscoverServers(timeout)
}

func newServerDiscoverer() ServerDiscoverer { return gdmDiscoverer{} }

type fileConfigGateway struct{}

func (fileConfigGateway) Load() (*config.Config, error) { return config.Load() }
func (fileConfigGateway) Delete() error                 { return config.Delete() }
func (fileConfigGateway) IsSetupComplete() bool         { return config.IsSetupComplete() }
func (fileConfigGateway) ConfigDir() string             { return config.GetConfigDir() }

func newConfigGateway() ConfigGateway { return fileConfigGateway{} }

type execRelauncher struct{}

func (execRelauncher) Relaunch() error {
	exe := version.LaunchPath()
	if exe == "" {
		return errors.New(errors.UNKNOWN_ERROR, "failed to locate executable")
	}

	cmd := exec.CommandContext(context.Background(), exe) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), relaunchPIDEnv+"="+strconv.Itoa(os.Getpid()))
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, errors.UNKNOWN_ERROR, "failed to relaunch application")
	}
	return nil
}

func newAppRelauncher() AppRelauncher { return execRelauncher{} }

func newPlexAuthenticatorFactory() PlexAuthenticatorFactory {
	return func() PlexAuthenticator { return plex.NewAuthenticator() }
}

var (
	_ TrayController      = (*platform.TrayManager)(nil)
	_ AutoStartController = (*platform.AutoStartManager)(nil)
	_ RetryManager        = (*retry.Manager)(nil)
	_ PlexAuthenticator   = (*plex.Authenticator)(nil)
	_ UpdateService       = (*updater.Updater)(nil)
	_ HistoryStore        = (*history.Store)(nil)
	_ ArtworkResolver     = (*artwork.Resolver)(nil)
	_ DiscordPresence     = (*discord.PresenceManager)(nil)
	_ Desktop             = wailsDesktop{}
)
