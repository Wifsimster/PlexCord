package main

import (
	"context"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/options"

	"plexcord/internal/config"
)

// TestResolveWindowLaunchState covers how the "Start minimized" and "Minimize
// to tray" settings combine into Wails' window start options.
func TestResolveWindowLaunchState(t *testing.T) {
	tests := []struct {
		name             string
		cfg              *config.Config
		isUpdateRelaunch bool
		wantHidden       bool
		wantState        options.WindowStartState
	}{
		{
			name:       "start minimized off shows the window",
			cfg:        &config.Config{StartMinimized: false, MinimizeToTray: true},
			wantHidden: false,
			wantState:  options.Normal,
		},
		{
			name:       "start minimized with tray starts hidden",
			cfg:        &config.Config{StartMinimized: true, MinimizeToTray: true},
			wantHidden: true,
			wantState:  options.Normal,
		},
		{
			name: "start minimized without tray minimizes to the taskbar",
			// Hiding outright would leave no restore affordance: closing to
			// the tray is off, so the taskbar button is the way back.
			cfg:        &config.Config{StartMinimized: true, MinimizeToTray: false},
			wantHidden: false,
			wantState:  options.Minimised,
		},
		{
			name:             "update relaunch always shows the window",
			cfg:              &config.Config{StartMinimized: true, MinimizeToTray: true},
			isUpdateRelaunch: true,
			wantHidden:       false,
			wantState:        options.Normal,
		},
		{
			name:       "nil config falls back to a normal window",
			cfg:        nil,
			wantHidden: false,
			wantState:  options.Normal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveWindowLaunchState(tt.cfg, tt.isUpdateRelaunch)
			if got.StartHidden != tt.wantHidden {
				t.Errorf("StartHidden = %v, want %v", got.StartHidden, tt.wantHidden)
			}
			if got.StartState != tt.wantState {
				t.Errorf("StartState = %v, want %v", got.StartState, tt.wantState)
			}
		})
	}
}

// TestDefaultConfigStartsVisible verifies PlexCord opens its window by default:
// starting in the background is opt-in.
func TestDefaultConfigStartsVisible(t *testing.T) {
	if got := resolveWindowLaunchState(config.DefaultConfig(), false); got.StartHidden || got.StartState != options.Normal {
		t.Fatalf("default config launch state = %+v, want a normal visible window", got)
	}
}

// TestShowWindowBeforeReadyIsDeferred verifies a restore request arriving
// before startup (a second instance launched while PlexCord is still booting)
// is parked rather than dropped — and does not run against a nil context.
func TestShowWindowBeforeReadyIsDeferred(t *testing.T) {
	app := &App{}

	app.ShowWindow()

	app.windowMu.Lock()
	pending := app.pendingShow
	app.windowMu.Unlock()
	if !pending {
		t.Fatal("ShowWindow before startup did not record a pending restore request")
	}

	if !app.markWindowReady(context.Background()) {
		t.Fatal("markWindowReady() = false, want true to replay the parked request")
	}
	if app.markWindowReady(context.Background()) {
		t.Fatal("markWindowReady() replayed the same request twice")
	}
}

// TestWindowContextAfterReady verifies that once the window is ready, show
// requests run against the published context instead of being parked.
func TestWindowContextAfterReady(t *testing.T) {
	ctx := context.Background()
	app := &App{}

	if app.markWindowReady(ctx) {
		t.Fatal("markWindowReady() = true with no request pending")
	}
	if app.windowContext() == nil {
		t.Fatal("windowContext() = nil after the window became ready")
	}

	app.windowMu.Lock()
	pending := app.pendingShow
	app.windowMu.Unlock()
	if pending {
		t.Fatal("windowContext() parked a request even though the window was ready")
	}
}
