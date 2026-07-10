package main

import (
	"log"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ============================================================================
// Window Management Methods (Story 4.5-4.10)
// ============================================================================

// ShowWindow shows and focuses the main application window.
// This is used when restoring from minimized/hidden state.
func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false) // Trick to bring to front
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
// background. Since Wails v2 has no native system tray, relaunching the app
// is the restore path: bring the existing window back to the foreground.
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
