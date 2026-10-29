package main

import (
	"testing"

	"plexcord/internal/config"
	"plexcord/internal/events"
)

func newTestApp(cfg *config.Config) *App {
	store := config.NewStore(cfg, func(*config.Config) error { return nil })
	desktop := &fakeDesktop{}
	a := &App{
		config:       cfg,
		cfgStore:     store,
		configs:      &fakeConfigGateway{cfg: cfg},
		tokens:       &fakeTokenStore{},
		discovery:    &fakeDiscoverer{},
		tray:         &fakeTray{},
		autostart:    &fakeAutoStart{},
		history:      &fakeHistory{},
		bus:          events.NewRecordingBus(),
		desktop:      desktop,
		plexFactory:  func(string, string) PlexAPI { return &fakePlexAPI{} },
		polling:      &pollingController{},
		sessions:     &sessionCache{},
		plexRetry:    &fakeRetry{},
		discordRetry: &fakeRetry{},
		authFactory:  func() PlexAuthenticator { return &fakeAuthenticator{} },
		relauncher:   &fakeRelauncher{},
	}
	a.windows = newWindowManager(desktop, desktop)
	a.presence = newPresenceGate(a.clearDiscordOnStop, a.hideWhenPausedDelay)
	a.discord = a.newDiscordService(&recordingPresence{}, nil)
	return a
}

func defaultTestConfig() *config.Config { return config.DefaultConfig() }

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
	app := newTestApp(&config.Config{})

	if err := app.AddServer("proxmox-docker-plex", "http://192.168.0.237:32400", "", ""); err != nil {
		t.Fatalf("AddServer returned error: %v", err)
	}

	status := app.GetPlexConnectionStatus()
	if status.ServerURL != "http://192.168.0.237:32400" {
		t.Errorf("expected added server URL in status, got %q", status.ServerURL)
	}
	if status.Connected {
		t.Errorf("expected Connected=false with no user selected/poller running")
	}
}
