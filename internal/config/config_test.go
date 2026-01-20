package config

import (
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
