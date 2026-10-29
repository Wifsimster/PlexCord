package config

import (
	"encoding/json"
	"os"
	"time"

	"plexcord/internal/errors"
)

type ServerConfig struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
	Active   bool   `json:"active"`
}

type Config struct {
	PlexLastConnected    *time.Time `json:"plexLastConnected,omitempty"`
	DiscordLastConnected *time.Time `json:"discordLastConnected,omitempty"`

	ServerURL            string `json:"serverUrl"`
	DiscordClientID      string `json:"discordClientId"`
	SelectedPlexUserID   string `json:"selectedPlexUserId"`
	SelectedPlexUserName string `json:"selectedPlexUserName"`
	PollingInterval      int    `json:"pollingInterval"`
	MinimizeToTray       bool   `json:"minimizeToTray"`
	AutoStart            bool   `json:"autoStart"`
	StartMinimized       bool   `json:"startMinimized"`
	SetupCompleted       bool   `json:"setupCompleted"`
	SetupSkipped         bool   `json:"setupSkipped"`

	HideWhenPaused      bool `json:"hideWhenPaused"`
	HideWhenPausedDelay int  `json:"hideWhenPausedDelay"`

	PresenceDetailsFormat string `json:"presenceDetailsFormat"`
	PresenceStateFormat   string `json:"presenceStateFormat"`

	PresenceActivityStyle string `json:"presenceActivityStyle"`
	PresenceStatusDisplay string `json:"presenceStatusDisplay"`

	PresenceArtworkLookup *bool `json:"presenceArtworkLookup,omitempty"`

	PresenceMediaTypes []string `json:"presenceMediaTypes,omitempty"`

	Servers []ServerConfig `json:"servers,omitempty"`

	AutoUpdateCheck *bool `json:"autoUpdateCheck,omitempty"`

	StartMinimizedOnLogin *bool `json:"startMinimizedOnLogin,omitempty"`
}

func (c *Config) IsAutoUpdateCheckEnabled() bool {
	return c.AutoUpdateCheck == nil || *c.AutoUpdateCheck
}

func boolPtr(b bool) *bool { return &b }

func (c *Config) ArtworkLookupEnabled() bool {
	return c.PresenceArtworkLookup == nil || *c.PresenceArtworkLookup
}

var AllMediaTypes = []string{"music", "movie", "tv"}

func (c *Config) EnabledMediaTypes() []string {
	enabled := make([]string, 0, len(AllMediaTypes))
	for _, known := range AllMediaTypes {
		for _, want := range c.PresenceMediaTypes {
			if want == known {
				enabled = append(enabled, known)
				break
			}
		}
	}
	if len(enabled) == 0 {
		return append([]string(nil), AllMediaTypes...)
	}
	return enabled
}

func (c *Config) IsMediaTypeEnabled(mediaType string) bool {
	for _, t := range c.EnabledMediaTypes() {
		if t == mediaType {
			return true
		}
	}
	return false
}

func (c *Config) LoginStartsMinimized() bool {
	return c.StartMinimizedOnLogin == nil || *c.StartMinimizedOnLogin
}

func DefaultConfig() *Config {
	return &Config{
		ServerURL:             "",
		PollingInterval:       2,
		MinimizeToTray:        true,
		AutoStart:             false,
		StartMinimized:        false,
		DiscordClientID:       "",
		PresenceActivityStyle: "media",
		PresenceStatusDisplay: "state",
		PresenceArtworkLookup: boolPtr(true),
		StartMinimizedOnLogin: boolPtr(true),
		PresenceMediaTypes:    append([]string(nil), AllMediaTypes...),
	}
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath) //nolint:gosec // configPath is derived from user config dir, not user input
	if err != nil {
		return nil, errors.New(errors.CONFIG_READ_FAILED, "failed to read config file: "+err.Error())
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, errors.New(errors.CONFIG_READ_FAILED, "invalid JSON in config file: "+err.Error())
	}

	cfg.MigrateToMultiServer()

	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return errors.New(errors.CONFIG_WRITE_FAILED, "failed to marshal config: "+err.Error())
	}

	if err := WriteFileAtomic(configPath, data, 0600); err != nil {
		return errors.New(errors.CONFIG_WRITE_FAILED, "failed to write config file: "+err.Error())
	}

	return nil
}

func ConfigExists() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(configPath)
	return err == nil
}

func IsSetupComplete() bool {
	if !ConfigExists() {
		return false
	}

	cfg, err := Load()
	if err != nil {
		return false
	}

	return cfg.SetupCompleted || cfg.SetupSkipped
}

func (c *Config) MigrateToMultiServer() {
	if len(c.Servers) > 0 {
		return
	}
	if c.ServerURL == "" {
		return
	}

	c.Servers = []ServerConfig{
		{
			Name:     "Default Server",
			URL:      c.ServerURL,
			UserID:   c.SelectedPlexUserID,
			UserName: c.SelectedPlexUserName,
			Active:   true,
		},
	}
}

func (c *Config) GetActiveServers() []ServerConfig {
	active := make([]ServerConfig, 0, len(c.Servers))
	for _, s := range c.Servers {
		if s.Active {
			active = append(active, s)
		}
	}
	return active
}

func Delete() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.New(errors.CONFIG_WRITE_FAILED, "failed to delete config file: "+err.Error())
	}

	return nil
}
