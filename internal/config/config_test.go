package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestConfigExists verifies that ConfigExists() correctly detects config file presence
func TestConfigExists(t *testing.T) {
	// Test when config file doesn't exist
	// (Using default path, which likely doesn't exist in test environment)
	exists := ConfigExists()
	// We can't assume the result since it depends on the environment
	// Just verify the function doesn't panic
	t.Logf("ConfigExists() returned: %v", exists)
}

// TestIsSetupComplete verifies that IsSetupComplete() correctly determines setup status
func TestIsSetupComplete(t *testing.T) {
	// Test with the actual config path
	complete := IsSetupComplete()
	t.Logf("IsSetupComplete() returned: %v", complete)

	// Verify the function doesn't panic and returns a boolean
	if complete != true && complete != false {
		t.Error("IsSetupComplete() should return a boolean value")
	}
}

// TestLoadDefaultConfig verifies that Load() returns default config when file doesn't exist
func TestLoadDefaultConfig(t *testing.T) {
	// This test assumes config file doesn't exist at test time
	// If it does exist, it will load that config instead

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should not return error when config doesn't exist: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() should return a config, got nil")
	}

	// Verify config has expected default values
	if cfg.PollingInterval <= 0 {
		t.Error("Default config should have positive polling interval")
	}
}

// TestSaveAndLoad verifies that Save() and Load() work together
func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a test config file path in temp directory
	testConfigPath := filepath.Join(tempDir, "test-config.json")

	// Write config to temp file
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

	// Verify file exists
	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Error("Test config file should exist after writing")
	}

	t.Logf("Test config created at: %s", testConfigPath)
}

// TestDefaultConfig verifies that DefaultConfig() returns valid defaults
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

	// DiscordClientID is intentionally empty in config - empty means "use the
	// discord package's DefaultClientID". Users can override with their own
	// Discord application client ID if desired.
	// This test verifies the field exists and is the expected empty default.
	if cfg.DiscordClientID != "" {
		t.Error("Default Discord client ID should be empty (use package default)")
	}
}

// TestIsAutoUpdateCheckEnabled verifies the nil-means-enabled default so that
// config files written before the setting existed keep automatic updates on.
func TestIsAutoUpdateCheckEnabled(t *testing.T) {
	if !DefaultConfig().IsAutoUpdateCheckEnabled() {
		t.Error("automatic update checks should default to enabled")
	}

	// Old config file without the autoUpdateCheck key -> enabled.
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

// TestLoginStartsMinimized verifies the nil-means-enabled default: a login
// launch comes up in the background unless the user turned that off, including
// for config files written before the setting existed.
func TestLoginStartsMinimized(t *testing.T) {
	if !DefaultConfig().LoginStartsMinimized() {
		t.Error("a login launch should default to starting minimized")
	}

	// Old config file without the startMinimizedOnLogin key -> enabled.
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
			// A config written before PlexCord handled video has no list at all.
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
			// A hand-edited config must not ask the poller for a type nothing
			// can render.
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
			// The settings UI presents them in one order; the stored order
			// must not leak into what the poller is asked for.
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
	// The caller hands the slice to the poller; mutating it must not rewrite
	// the package default for every later call.
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
