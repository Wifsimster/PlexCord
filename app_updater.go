package main

import (
	"fmt"
	"log"

	"plexcord/internal/platform"
	"plexcord/internal/updater"
)

func (a *App) GetAutoUpdateCheck() bool {
	return a.config.IsAutoUpdateCheckEnabled()
}

func (a *App) SetAutoUpdateCheck(enabled bool) error {
	a.config.AutoUpdateCheck = &enabled
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save auto update check setting: %v", err)
		return err
	}

	if a.updater != nil {
		if enabled {
			a.updater.StartChecker(a.ctx)
		} else {
			a.updater.StopChecker()
		}
	}

	log.Printf("Automatic update checks set to: %v", enabled)
	return nil
}

func (a *App) GetUpdateStatus() updater.Status {
	if a.updater == nil {
		return updater.Status{State: updater.StateIdle}
	}
	return a.updater.GetStatus()
}

func (a *App) publishUpdateNotice(status updater.Status) {
	if a.tray == nil {
		return
	}
	a.tray.SetUpdateNotice(updateNotice(status))
}

func updateNotice(status updater.Status) platform.UpdateNotice {
	version := ""
	if status.Info != nil {
		version = status.Info.LatestVersion
	}
	if version == "" {
		return platform.UpdateNotice{}
	}

	switch status.State {
	case updater.StateReady:
		return platform.UpdateNotice{
			Label:      fmt.Sprintf("Restart to update to %s", version),
			Tooltip:    fmt.Sprintf("PlexCord %s is installed — restart to apply", version),
			Actionable: true,
		}
	case updater.StateDownloading:
		return platform.UpdateNotice{
			Label:   fmt.Sprintf("Downloading update %s…", version),
			Tooltip: fmt.Sprintf("PlexCord is downloading %s in the background", version),
		}
	case updater.StateAvailable:
		return platform.UpdateNotice{
			Label:      fmt.Sprintf("Update available: %s", version),
			Tooltip:    fmt.Sprintf("PlexCord %s is available", version),
			Actionable: true,
		}
	default:
		return platform.UpdateNotice{}
	}
}

func (a *App) onTrayUpdateClick() {
	if a.updater != nil && a.updater.GetStatus().State == updater.StateReady {
		if err := a.RestartApplication(); err != nil {
			log.Printf("ERROR: Failed to restart for update from tray: %v", err)
			a.ShowWindow()
		}
		return
	}
	a.ShowWindow()
}
