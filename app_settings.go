package main

import (
	"context"
	"log"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ============================================================================
// Window Management Methods (Story 4.5-4.10)
// ============================================================================

// ShowWindow shows and focuses the main application window.
// This is used when restoring from minimized/hidden state: from the tray, and
// from a relaunch of PlexCord while it is already running in the background.
func (a *App) ShowWindow() {
	ctx := a.windowContext()
	if ctx == nil {
		log.Printf("Show requested before the window was ready; deferred until startup completes")
		return
	}

	runtime.WindowShow(ctx)
	// Only un-minimise when the window actually is minimised: on a window that
	// was hidden while maximised, an unconditional restore would also drop it
	// back to its normal size.
	if runtime.WindowIsMinimised(ctx) {
		runtime.WindowUnminimise(ctx)
	}
	runtime.WindowSetAlwaysOnTop(ctx, true)
	runtime.WindowSetAlwaysOnTop(ctx, false) // Trick to bring to front
}

// windowContext returns the context to drive the window with, or nil when the
// window is not ready yet — a restore arriving before OnStartup handed us the
// Wails context. In that case the request is remembered so markWindowReady can
// replay it, rather than being dropped (or run against a nil context, which
// panics inside the Wails runtime).
func (a *App) windowContext() context.Context {
	a.windowMu.Lock()
	defer a.windowMu.Unlock()

	if a.windowCtx != nil {
		return a.windowCtx
	}
	a.pendingShow = true
	return nil
}

// markWindowReady publishes the Wails context that drives the window and
// reports whether a restore request arrived before it existed, in which case
// the caller should replay it. Called from startup.
func (a *App) markWindowReady(ctx context.Context) bool {
	a.windowMu.Lock()
	defer a.windowMu.Unlock()

	a.windowCtx = ctx
	pending := a.pendingShow
	a.pendingShow = false
	return pending
}

// HideWindow hides the main application window.
// The application continues running in the background.
func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

// MinimizeWindow minimizes the main application window.
func (a *App) MinimizeWindow() {
	runtime.WindowMinimise(a.ctx)
}

// QuitApp terminates the application completely.
// This is called from the tray menu or when the user explicitly quits.
// It flags an explicit quit so beforeClose allows shutdown instead of
// hiding the window when "Minimize to tray" is enabled.
func (a *App) QuitApp() {
	log.Printf("Quit requested")
	a.quitting.Store(true)
	runtime.Quit(a.ctx)
}

// onSecondInstanceLaunch is invoked (via SingleInstanceLock) when the user
// launches PlexCord again while an instance is already running in the
// background. Alongside the system tray, relaunching is a restore path:
// bring the existing window back to the foreground instead of starting a copy.
//
// This is the path that makes "Start minimized" recoverable on Windows: with
// no window and (when minimizing to the tray) no taskbar button, re-running
// PlexCord — from the Start menu, a shortcut, or a double-clicked exe — is
// what the user reaches for, and it must reopen the running instance. The
// second instance can land before OnStartup has run, so ShowWindow parks the
// request until the window exists rather than dropping it.
func (a *App) onSecondInstanceLaunch(options.SecondInstanceData) {
	log.Printf("Second instance launched: restoring existing window")
	a.ShowWindow()
}

// GetMinimizeToTray returns whether the app should minimize to tray.
func (a *App) GetMinimizeToTray() bool {
	return a.config.MinimizeToTray
}

// SetMinimizeToTray updates the minimize to tray setting.
func (a *App) SetMinimizeToTray(enabled bool) error {
	a.config.MinimizeToTray = enabled
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save minimize to tray setting: %v", err)
		return err
	}
	log.Printf("Minimize to tray set to: %v", enabled)
	return nil
}

// GetStartMinimized returns whether PlexCord should launch in the background
// instead of showing its window.
func (a *App) GetStartMinimized() bool {
	return a.config.StartMinimized
}

// SetStartMinimized updates the start-minimized setting. It takes effect on
// the next launch — the window state is decided before Wails starts, in
// resolveWindowLaunchState.
func (a *App) SetStartMinimized(enabled bool) error {
	a.config.StartMinimized = enabled
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save start minimized setting: %v", err)
		return err
	}
	log.Printf("Start minimized set to: %v", enabled)
	return nil
}

// GetAutoStart returns whether auto-start on login is enabled.
// This checks the actual OS registration, not just the config value.
func (a *App) GetAutoStart() bool {
	return a.autostart.IsEnabled()
}

// SetAutoStart enables or disables auto-start on login.
// On Windows: Adds/removes from HKCU\Software\Microsoft\Windows\CurrentVersion\Run
// On macOS: Creates/removes LaunchAgent plist
// On Linux: Creates/removes XDG .desktop file in ~/.config/autostart/
func (a *App) SetAutoStart(enabled bool) error {
	// Update OS auto-start registration
	if err := a.autostart.SetEnabled(enabled); err != nil {
		log.Printf("ERROR: Failed to set auto-start: %v", err)
		return err
	}

	// Update config to match
	a.config.AutoStart = enabled
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save auto-start setting: %v", err)
		// Note: OS registration succeeded but config save failed
		// The actual auto-start behavior will work, but config may be out of sync
		return err
	}

	log.Printf("Auto-start set to: %v", enabled)
	return nil
}
