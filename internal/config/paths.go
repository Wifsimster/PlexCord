package config

import (
	"os"
	"path/filepath"
	"runtime"

	"plexcord/internal/errors"
)

// GetConfigPath returns the platform-specific configuration file path
func GetConfigPath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\PlexCord\config.json
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", errors.New(errors.CONFIG_READ_FAILED, "APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "PlexCord")

	case "darwin":
		// macOS: ~/Library/Application Support/PlexCord/config.json
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", errors.New(errors.CONFIG_READ_FAILED, "failed to get home directory: "+err.Error())
		}
		configDir = filepath.Join(homeDir, "Library", "Application Support", "PlexCord")

	default:
		// Linux: ~/.config/plexcord/config.json
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return "", errors.New(errors.CONFIG_READ_FAILED, "failed to get config directory: "+err.Error())
		}
		configDir = filepath.Join(userConfigDir, "plexcord")
	}

	return filepath.Join(configDir, "config.json"), nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)

	// Create directory with 0700 permissions (owner read/write/execute only)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return errors.New(errors.CONFIG_WRITE_FAILED, "failed to create config directory: "+err.Error())
	}

	return nil
}
