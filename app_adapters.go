package main

import (
	"plexcord/internal/keychain"
	"plexcord/internal/plex"
)

// This file contains adapters that wrap the concrete internal packages
// to satisfy the interfaces defined in app_interfaces.go. They are the
// production implementations injected into App at startup.

// keychainTokenStore adapts the package-level keychain functions to
// the TokenStore interface for dependency injection.
type keychainTokenStore struct{}

func (keychainTokenStore) Get() (string, error)     { return keychain.GetToken() }
func (keychainTokenStore) Set(token string) error   { return keychain.SetToken(token) }
func (keychainTokenStore) Delete() error            { return keychain.DeleteToken() }

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
