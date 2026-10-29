package keychain

import (
	"log"

	"plexcord/internal/errors"

	"github.com/zalando/go-keyring"
)

const (
	ServiceName = "PlexCord"
	TokenKey    = "plex-token"
)

func SetToken(token string) error {
	if token == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "token cannot be empty")
	}

	err := keyring.Set(ServiceName, TokenKey, token)
	if err != nil {
		if fallbackErr := setTokenFallback(token); fallbackErr != nil {
			return errors.Wrap(fallbackErr, errors.KEYCHAIN_STORE_FAILED, "failed to store token in keychain or fallback")
		}
		return nil
	}

	if err := deleteTokenFallback(); err != nil {
		log.Printf("Warning: failed to remove stale token fallback: %v", err)
	}
	return nil
}

func GetToken() (string, error) {
	token, err := keyring.Get(ServiceName, TokenKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			fallbackToken, fallbackErr := getTokenFallback()
			if fallbackErr != nil {
				return "", nil
			}
			return fallbackToken, nil
		}

		fallbackToken, fallbackErr := getTokenFallback()
		if fallbackErr != nil {
			return "", errors.Wrap(fallbackErr, errors.KEYCHAIN_READ_FAILED, "failed to retrieve token from keychain or fallback")
		}
		return fallbackToken, nil
	}

	return token, nil
}

func DeleteToken() error {
	err := keyring.Delete(ServiceName, TokenKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			if fallbackErr := deleteTokenFallback(); fallbackErr != nil {
				return errors.Wrap(fallbackErr, errors.KEYCHAIN_READ_FAILED, "failed to delete token fallback")
			}
			return nil
		}

		if fallbackErr := deleteTokenFallback(); fallbackErr != nil {
			return errors.Wrap(fallbackErr, errors.KEYCHAIN_READ_FAILED, "failed to delete token from keychain or fallback")
		}
		return nil
	}

	if fallbackErr := deleteTokenFallback(); fallbackErr != nil {
		return errors.Wrap(fallbackErr, errors.KEYCHAIN_READ_FAILED, "failed to delete token fallback")
	}
	return nil
}
