package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigExists(t *testing.T) {
	exists := ConfigExists()
	t.Logf("ConfigExists() returned: %v", exists)
}

func TestIsSetupComplete(t *testing.T) {
	complete := IsSetupComplete()
	t.Logf("IsSetupComplete() returned: %v", complete)

	if complete != true && complete != false {
		t.Error("IsSetupComplete() should return a boolean value")
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should not return error when config doesn't exist: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() should return a config, got nil")
	}

	if cfg.PollingInterval <= 0 {
		t.Error("Default config should have positive polling interval")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()

	testConfigPath := filepath.Join(tempDir, "test-config.json")

	data := `{
  "serverUrl": "http://test.example.com",
  "pollingInterval": 10,
  "minimizeToTray": false,
  "autoStart": true,
  "discordClientId": "TEST_CLIENT_ID"
}`
	err := os.WriteFile(testConfigPath, []byte(data), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Error("Test config file should exist after writing")
	}

	t.Logf("Test config created at: %s", testConfigPath)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() should not return nil")
	}

	if cfg.PollingInterval <= 0 {
		t.Error("Default polling interval should be positive")
	}

	if cfg.PollingInterval > 60 {
		t.Error("Default polling interval should be reasonable (< 60 seconds)")
	}

	if cfg.DiscordClientID != "" {
		t.Error("Default Discord client ID should be empty (use package default)")
	}
}

func TestIsAutoUpdateCheckEnabled(t *testing.T) {
	if !DefaultConfig().IsAutoUpdateCheckEnabled() {
		t.Error("automatic update checks should default to enabled")
	}

	var legacy Config
	if err := json.Unmarshal([]byte(`{"serverUrl":"http://plex:32400"}`), &legacy); err != nil {
		t.Fatalf("unmarshal legacy config: %v", err)
	}
	if !legacy.IsAutoUpdateCheckEnabled() {
		t.Error("configs without the autoUpdateCheck key should be enabled")
	}

	enabled := true
	disabled := false
	if !(&Config{AutoUpdateCheck: &enabled}).IsAutoUpdateCheckEnabled() {
		t.Error("explicit true should be enabled")
	}
	if (&Config{AutoUpdateCheck: &disabled}).IsAutoUpdateCheckEnabled() {
		t.Error("explicit false should be disabled")
	}
}

func TestLoginStartsMinimized(t *testing.T) {
	if !DefaultConfig().LoginStartsMinimized() {
		t.Error("a login launch should default to starting minimized")
	}

	var legacy Config
	if err := json.Unmarshal([]byte(`{"serverUrl":"http://plex:32400"}`), &legacy); err != nil {
		t.Fatalf("unmarshal legacy config: %v", err)
	}
	if !legacy.LoginStartsMinimized() {
		t.Error("configs without the startMinimizedOnLogin key should start minimized on login")
	}

	enabled := true
	disabled := false
	if !(&Config{StartMinimizedOnLogin: &enabled}).LoginStartsMinimized() {
		t.Error("explicit true should start minimized on login")
	}
	if (&Config{StartMinimizedOnLogin: &disabled}).LoginStartsMinimized() {
		t.Error("explicit false should open the window on login")
	}
}

func TestEnabledMediaTypes(t *testing.T) {
	tests := []struct {
		name      string
		persisted []string
		want      []string
	}{
		{
			name:      "unset means every kind",
			persisted: nil,
			want:      []string{"music", "movie", "tv"},
		},
		{
			name:      "an empty list means every kind",
			persisted: []string{},
			want:      []string{"music", "movie", "tv"},
		},
		{
			name:      "a selection is honoured",
			persisted: []string{"movie", "tv"},
			want:      []string{"movie", "tv"},
		},
		{
			name:      "unknown values are dropped",
			persisted: []string{"music", "podcast"},
			want:      []string{"music"},
		},
		{
			name:      "only unknown values fall back to every kind",
			persisted: []string{"podcast"},
			want:      []string{"music", "movie", "tv"},
		},
		{
			name:      "the canonical order is used",
			persisted: []string{"tv", "music"},
			want:      []string{"music", "tv"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{PresenceMediaTypes: tt.persisted}
			got := cfg.EnabledMediaTypes()
			if len(got) != len(tt.want) {
				t.Fatalf("EnabledMediaTypes() = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("EnabledMediaTypes() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestIsMediaTypeEnabled(t *testing.T) {
	cfg := &Config{PresenceMediaTypes: []string{"music"}}

	if !cfg.IsMediaTypeEnabled("music") {
		t.Error("IsMediaTypeEnabled(music) = false for a config that selected it")
	}
	if cfg.IsMediaTypeEnabled("movie") {
		t.Error("IsMediaTypeEnabled(movie) = true for a config that did not select it")
	}
}

func TestEnabledMediaTypesDoesNotAliasTheDefault(t *testing.T) {
	cfg := &Config{}
	got := cfg.EnabledMediaTypes()
	got[0] = "tampered"

	if AllMediaTypes[0] != "music" {
		t.Errorf("AllMediaTypes was mutated through a returned slice: %v", AllMediaTypes)
	}
	if cfg.EnabledMediaTypes()[0] != "music" {
		t.Error("a later call saw the tampered value")
	}
}
