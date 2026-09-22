package main

import (
	"testing"

	"plexcord/internal/config"
)

// TestResetApplicationKeepsStoreInSync covers the wizard re-run after a reset:
// the reset defaults must be what later saves persist, not the settings the
// user just reset away.
func TestResetApplicationKeepsStoreInSync(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ServerURL = "http://old:32400"
	cfg.SetupCompleted = true
	app := newTestApp(cfg)

	app.ResetApplication()

	if app.config != app.cfgStore.Get() {
		t.Fatal("a.config and the store diverged after reset")
	}
	stored := app.cfgStore.Get()
	if stored.ServerURL != "" || stored.SetupCompleted {
		t.Errorf("store still holds pre-reset settings: ServerURL=%q SetupCompleted=%v",
			stored.ServerURL, stored.SetupCompleted)
	}
}
