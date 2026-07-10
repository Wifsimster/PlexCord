package main

import (
	"testing"

	"plexcord/internal/config"
)

// newTestApp builds an App backed by an in-memory config store that never
// touches disk, so server/connection helpers can be exercised in isolation.
func newTestApp(cfg *config.Config) *App {
	store := config.NewStore(cfg, func(*config.Config) error { return nil })
	return &App{config: cfg, cfgStore: store}
}

func TestActivePlexServerURL_PrefersActiveMultiServerEntry(t *testing.T) {
	app := newTestApp(&config.Config{
		ServerURL: "http://legacy:32400",
		Servers: []config.ServerConfig{
			{Name: "Disabled", URL: "http://disabled:32400", Active: false},
			{Name: "Active", URL: "http://active:32400", Active: true},
		},
	})

	if got := app.activePlexServerURL(); got != "http://active:32400" {
		t.Errorf("expected active server URL, got %q", got)
	}
}

func TestActivePlexServerURL_FallsBackToLegacyServerURL(t *testing.T) {
	// No servers, or only inactive ones, should fall back to the legacy field.
	app := newTestApp(&config.Config{
		ServerURL: "http://legacy:32400",
		Servers: []config.ServerConfig{
			{Name: "Disabled", URL: "http://disabled:32400", Active: false},
		},
	})

	if got := app.activePlexServerURL(); got != "http://legacy:32400" {
		t.Errorf("expected legacy server URL fallback, got %q", got)
	}
}

func TestActivePlexServerURL_EmptyWhenNothingConfigured(t *testing.T) {
	app := newTestApp(&config.Config{})

	if got := app.activePlexServerURL(); got != "" {
		t.Errorf("expected empty URL when nothing configured, got %q", got)
	}
}

func TestAddServer_SurfacesThroughConnectionStatus(t *testing.T) {
	// A server added via the Settings dialog (config.Servers only, no legacy
	// ServerURL) must become the URL the dashboard reports, so the connection
	// path targets it instead of an empty legacy field.
	app := newTestApp(&config.Config{})

	if err := app.AddServer("proxmox-docker-plex", "http://192.168.0.237:32400", "", ""); err != nil {
		t.Fatalf("AddServer returned error: %v", err)
	}

	status := app.GetPlexConnectionStatus()
	if status.ServerURL != "http://192.168.0.237:32400" {
		t.Errorf("expected added server URL in status, got %q", status.ServerURL)
	}
	// Without a selected user the dashboard must not report a live connection.
	if status.Connected {
		t.Errorf("expected Connected=false with no user selected/poller running")
	}
}
