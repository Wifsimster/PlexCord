package main

import (
	"log"

	"github.com/wailsapp/wails/v2/pkg/options"
)

func (a *App) ShowWindow() {
	a.windows.Show()
}

func (a *App) HideWindow() {
	a.windows.Hide()
}

func (a *App) MinimizeWindow() {
	a.windows.Minimise()
}

func (a *App) QuitApp() {
	log.Printf("Quit requested")
	a.windows.Quit()
}

func (a *App) onSecondInstanceLaunch(options.SecondInstanceData) {
	log.Printf("Second instance launched: restoring existing window")
	a.ShowWindow()
}

func (a *App) GetMinimizeToTray() bool {
	return a.config.MinimizeToTray
}

func (a *App) SetMinimizeToTray(enabled bool) error {
	a.config.MinimizeToTray = enabled
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save minimize to tray setting: %v", err)
		return err
	}
	log.Printf("Minimize to tray set to: %v", enabled)
	return nil
}

func (a *App) GetStartMinimized() bool {
	return a.config.StartMinimized
}

func (a *App) SetStartMinimized(enabled bool) error {
	a.config.StartMinimized = enabled
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save start minimized setting: %v", err)
		return err
	}
	log.Printf("Start minimized set to: %v", enabled)
	return nil
}

func (a *App) GetStartMinimizedOnLogin() bool {
	return a.config.LoginStartsMinimized()
}

func (a *App) SetStartMinimizedOnLogin(enabled bool) error {
	a.config.StartMinimizedOnLogin = &enabled
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save start minimized on login setting: %v", err)
		return err
	}
	log.Printf("Start minimized on login set to: %v", enabled)
	return nil
}

func (a *App) GetAutoStart() bool {
	return a.autostart.IsEnabled()
}

func (a *App) SetAutoStart(enabled bool) error {
	if err := a.autostart.SetEnabled(enabled); err != nil {
		log.Printf("ERROR: Failed to set auto-start: %v", err)
		return err
	}

	a.config.AutoStart = enabled
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save auto-start setting: %v", err)
		return err
	}

	log.Printf("Auto-start set to: %v", enabled)
	return nil
}
