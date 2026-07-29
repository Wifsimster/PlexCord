package main

import (
	"log"

	"github.com/wailsapp/wails/v2/pkg/options"

	"plexcord/internal/config"
)

// windowLaunchState describes how the main window should come up, expressed in
// the two Wails options that decide it.
type windowLaunchState struct {
	StartState  options.WindowStartState
	StartHidden bool
}

// resolveWindowLaunchState maps the user's preferences onto those options.
//
// "Start minimized" has two flavors, mirroring what closing the window already
// does:
//   - with "Minimize to tray" on, PlexCord starts hidden — no window and no
//     taskbar button, just the tray icon. Restoring goes through the tray or
//     through relaunching PlexCord, which the single-instance lock turns into
//     a "show the running instance" request (see App.onSecondInstanceLaunch).
//   - with it off, the window starts minimized to the taskbar instead. Hiding
//     it outright would leave the user no way back: closing to the tray is
//     disabled, so the taskbar button is the only restore affordance.
//
// An update relaunch always comes up normally, whatever the setting says: the
// user just clicked "restart to apply", and a restart that vanished into the
// tray would read as a crash.
func resolveWindowLaunchState(cfg *config.Config, isUpdateRelaunch bool) windowLaunchState {
	if cfg == nil || !cfg.StartMinimized || isUpdateRelaunch {
		return windowLaunchState{StartState: options.Normal}
	}
	if cfg.MinimizeToTray {
		return windowLaunchState{StartState: options.Normal, StartHidden: true}
	}
	return windowLaunchState{StartState: options.Minimised}
}

// loadLaunchConfig reads the persisted configuration for the sole purpose of
// deciding the initial window state, which Wails needs up front — before
// OnStartup runs and the app loads the config it actually works with.
//
// A config that cannot be read must never keep the window from opening, so
// failures fall back to defaults (which start the window normally).
func loadLaunchConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: failed to load config for window start state, using defaults: %v", err)
		return config.DefaultConfig()
	}
	return cfg
}
