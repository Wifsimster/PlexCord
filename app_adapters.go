package main

import (
	"time"

	"plexcord/internal/config"
	"plexcord/internal/keychain"
	"plexcord/internal/platform"
	"plexcord/internal/plex"
)

// This file contains adapters that wrap the concrete internal packages
// to satisfy the interfaces defined in app_interfaces.go. They are the
// production implementations injected into App at startup.
//
// Each adapter is deliberately trivial: it does nothing but forward. That is
// what keeps the interface boundary honest — there is no logic here that a
// test replacing the adapter would lose.

// keychainTokenStore adapts the package-level keychain functions to
// the TokenStore interface for dependency injection.
type keychainTokenStore struct{}

func (keychainTokenStore) Get() (string, error)   { return keychain.GetToken() }
func (keychainTokenStore) Set(token string) error { return keychain.SetToken(token) }
func (keychainTokenStore) Delete() error          { return keychain.DeleteToken() }

// newKeychainTokenStore returns the default OS-keychain-backed TokenStore.
func newKeychainTokenStore() TokenStore {
	return keychainTokenStore{}
}

// newPlexClientFactory returns the default PlexAPIFactory that constructs
// concrete *plex.Client instances. Tests can replace this with a factory
// that returns fakes.
func newPlexClientFactory() PlexAPIFactory {
	return func(token, serverURL string) PlexAPI {
		return plex.NewClient(token, serverURL)
	}
}

// gdmDiscoverer adapts the package-level GDM discovery to ServerDiscoverer.
type gdmDiscoverer struct{}

func (gdmDiscoverer) Discover(timeout time.Duration) ([]plex.Server, error) {
	return plex.DiscoverServers(timeout)
}

// newServerDiscoverer returns the default GDM-backed ServerDiscoverer.
func newServerDiscoverer() ServerDiscoverer { return gdmDiscoverer{} }

// fileConfigGateway adapts the config package's file-backed top-level
// functions to the ConfigGateway interface.
type fileConfigGateway struct{}

func (fileConfigGateway) Load() (*config.Config, error) { return config.Load() }
func (fileConfigGateway) Delete() error                 { return config.Delete() }
func (fileConfigGateway) IsSetupComplete() bool         { return config.IsSetupComplete() }
func (fileConfigGateway) ConfigDir() string             { return config.GetConfigDir() }

// newConfigGateway returns the default file-backed ConfigGateway.
func newConfigGateway() ConfigGateway { return fileConfigGateway{} }

// Compile-time assertions that the production types satisfy the interfaces
// App depends on. These keep a signature change in an internal package from
// silently becoming a runtime surprise at startup.
var (
	_ TrayController      = (*platform.TrayManager)(nil)
	_ AutoStartController = (*platform.AutoStartManager)(nil)
)
