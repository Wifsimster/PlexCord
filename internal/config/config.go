package config

import (
	"encoding/json"
	"os"
	"time"

	"plexcord/internal/errors"
)

// Config holds application configuration
type Config struct {
	// Connection history (Story 6.8)
	PlexLastConnected    *time.Time `json:"plexLastConnected,omitempty"`    // Last successful Plex connection
	DiscordLastConnected *time.Time `json:"discordLastConnected,omitempty"` // Last successful Discord connection

	ServerURL            string `json:"serverUrl"`
	DiscordClientID      string `json:"discordClientId"`
	SelectedPlexUserID   string `json:"selectedPlexUserId"`   // ID of the Plex user to monitor
	SelectedPlexUserName string `json:"selectedPlexUserName"` // Display name for UI purposes
	PollingInterval      int    `json:"pollingInterval"`      // seconds
	MinimizeToTray       bool   `json:"minimizeToTray"`
	AutoStart            bool   `json:"autoStart"`
	SetupCompleted       bool   `json:"setupCompleted"` // True when setup wizard is done
	SetupSkipped         bool   `json:"setupSkipped"`   // True when user skipped setup
}

// DefaultConfig returns a configuration with default values.
// Default PollingInterval is 2 seconds to meet NFR4 requirement:
// "Discord presence updates shall occur within 2 seconds of playback state change"
func DefaultConfig() *Config {
	return &Config{
		ServerURL:       "",
		PollingInterval: 2, // 2 seconds for NFR4 compliance
		MinimizeToTray:  true,
		AutoStart:       false,
		DiscordClientID: "", // Empty means use default from discord package
	}
}

// Load loads configuration from file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist - return default config (not an error)
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath) //nolint:gosec // configPath is derived from user config dir, not user input
	if err != nil {
		return nil, errors.New(errors.CONFIG_READ_FAILED, "failed to read config file: "+err.Error())
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, errors.New(errors.CONFIG_READ_FAILED, "invalid JSON in config file: "+err.Error())
	}

	return &cfg, nil
}

// Save saves configuration to file
func Save(cfg *Config) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	// Get config file path
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Marshal config to JSON with indentation
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return errors.New(errors.CONFIG_WRITE_FAILED, "failed to marshal config: "+err.Error())
	}

	// Write to file with 0600 permissions (owner read/write only)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return errors.New(errors.CONFIG_WRITE_FAILED, "failed to write config file: "+err.Error())
	}

	return nil
}

// ConfigExists checks if the configuration file exists on disk
func ConfigExists() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(configPath)
	return err == nil
}

// IsSetupComplete checks if the setup wizard has been completed or skipped
// Setup is considered complete if:
// 1. Config file exists
// 2. Config file can be loaded successfully
// 3. SetupCompleted OR SetupSkipped flag is true
func IsSetupComplete() bool {
	// First check if config file exists
	if !ConfigExists() {
		return false
	}

	// Try to load config
	cfg, err := Load()
	if err != nil {
		return false
	}

	// Check if setup was explicitly completed or skipped
	return cfg.SetupCompleted || cfg.SetupSkipped
}

// Delete removes the configuration file from disk.
// This is used during application reset to clear all settings.
// Returns nil if the file doesn't exist (idempotent).
func Delete() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Remove the config file
	if err := os.Remove(configPath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - that's fine
			return nil
		}
		return errors.New(errors.CONFIG_WRITE_FAILED, "failed to delete config file: "+err.Error())
	}

	return nil
}
