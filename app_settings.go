package main

import (
	"log"

	"github.com/wailsapp/wails/v2/pkg/options"
)

// ============================================================================
// Window Management Methods (Story 4.5-4.10)
// ============================================================================

// ShowWindow shows and focuses the main application window.
// This is used when restoring from minimized/hidden state: from the tray, and
// from a relaunch of PlexCord while it is already running in the background.
//
// The window mechanics live in windowManager; App only forwards, so the Wails
// binding surface stays a translation layer.
func (a *App) ShowWindow() {
	a.windows.Show()
}

// HideWindow hides the main application window.
// The application continues running in the background.
func (a *App) HideWindow() {
	a.windows.Hide()
}

// MinimizeWindow minimizes the main application window.
func (a *App) MinimizeWindow() {
	a.windows.Minimise()
}

// QuitApp terminates the application completely.
// This is called from the tray menu or when the user explicitly quits.
// It flags an explicit quit so beforeClose allows shutdown instead of
// hiding the window when "Minimize to tray" is enabled.
func (a *App) QuitApp() {
	log.Printf("Quit requested")
	a.windows.Quit()
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

// GetStartMinimizedOnLogin returns whether a launch the OS performs at login
// should come up in the background instead of opening the window.
func (a *App) GetStartMinimizedOnLogin() bool {
	return a.config.LoginStartsMinimized()
}

// SetStartMinimizedOnLogin updates that setting. Like SetStartMinimized it
// takes effect on the next launch: the window state is decided before Wails
// starts, in resolveWindowLaunchState.
func (a *App) SetStartMinimizedOnLogin(enabled bool) error {
	a.config.StartMinimizedOnLogin = &enabled
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save start minimized on login setting: %v", err)
		return err
	}
	log.Printf("Start minimized on login set to: %v", enabled)
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
