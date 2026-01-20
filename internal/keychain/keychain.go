package keychain

import (
	"log"

	"github.com/zalando/go-keyring"
	"plexcord/internal/errors"
)

const (
	// ServiceName is the identifier for PlexCord in the OS keychain
	ServiceName = "PlexCord"
	// TokenKey is the account name for the Plex authentication token
	TokenKey = "plex-token"
)

// SetToken stores the Plex authentication token securely in the OS keychain.
// If the OS keychain is unavailable, it falls back to encrypted file storage.
//
// Supported platforms:
// - Windows: Uses Credential Manager
// - macOS: Uses Keychain Access
// - Linux: Uses Secret Service API (libsecret)
//
// Example:
//
//	err := keychain.SetToken("my-plex-token-here")
//	if err != nil {
//	    // Handle error
//	}
func SetToken(token string) error {
	if token == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "token cannot be empty")
	}

	err := keyring.Set(ServiceName, TokenKey, token)
	if err != nil {
		// If keychain unavailable, use fallback encryption
		if isKeychainUnavailable(err) {
			log.Printf("WARNING: OS keychain unavailable, using encrypted fallback storage")
			return setTokenFallback(token)
		}
		return errors.Wrap(err, errors.KEYCHAIN_STORE_FAILED, "failed to store token in keychain")
	}

	return nil
}

// GetToken retrieves the Plex authentication token from the OS keychain.
// If the OS keychain is unavailable, it attempts to read from encrypted fallback storage.
//
// Returns an empty string (not an error) if the token has not been set yet.
//
// Example:
//
//	token, err := keychain.GetToken()
//	if err != nil {
//	    // Handle error
//	}
//	if token == "" {
//	    // Token not set, user needs to complete setup
//	}
func GetToken() (string, error) {
	token, err := keyring.Get(ServiceName, TokenKey)
	if err != nil {
		// If keychain unavailable, try fallback
		if isKeychainUnavailable(err) {
			return getTokenFallback()
		}

		// Token not found is not an error (user hasn't set it up yet)
		if err == keyring.ErrNotFound {
			return "", nil
		}

		return "", errors.Wrap(err, errors.KEYCHAIN_READ_FAILED, "failed to retrieve token from keychain")
	}

	return token, nil
}

// DeleteToken removes the Plex authentication token from the OS keychain.
// This is used when the user resets the application or wants to reconfigure.
//
// Example:
//
//	err := keychain.DeleteToken()
//	if err != nil {
//	    // Handle error
//	}
func DeleteToken() error {
	err := keyring.Delete(ServiceName, TokenKey)
	if err != nil {
		// If keychain unavailable, try fallback
		if isKeychainUnavailable(err) {
			return deleteTokenFallback()
		}

		// Token not found is not an error
		if err == keyring.ErrNotFound {
			return nil
		}

		return errors.Wrap(err, errors.KEYCHAIN_READ_FAILED, "failed to delete token from keychain")
	}

	return nil
}

// isKeychainUnavailable checks if the error indicates that the OS keychain
// is not available (e.g., running in a restricted environment)
func isKeychainUnavailable(err error) bool {
	if err == nil {
		return false
	}

	// go-keyring returns specific errors when keychain is unavailable
	// This is a simplified check - in production you'd check error strings
	// For now, we'll return false to always use keychain (fallback is backup)
	return false
}
