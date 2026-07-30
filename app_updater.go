package main

import (
	"fmt"
	"log"

	"plexcord/internal/platform"
	"plexcord/internal/updater"
)

// ============================================================================
// Automatic Updates
// ============================================================================

// GetAutoUpdateCheck returns whether automatic background update checks are
// enabled. Defaults to true when the setting has never been persisted.
func (a *App) GetAutoUpdateCheck() bool {
	return a.config.IsAutoUpdateCheckEnabled()
}

// SetAutoUpdateCheck enables or disables automatic background update checks
// and starts/stops the checker immediately so the change takes effect without
// a restart.
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

// GetUpdateStatus returns a snapshot of the updater state so the frontend can
// hydrate on load (for example, an update downloaded in the background while
// the user was on another page still shows "restart to apply").
func (a *App) GetUpdateStatus() updater.Status {
	if a.updater == nil {
		return updater.Status{State: updater.StateIdle}
	}
	return a.updater.GetStatus()
}

// ----------------------------------------------------------------------------
// Tray notification
// ----------------------------------------------------------------------------

// publishUpdateNotice mirrors the updater state into the system tray.
//
// The frontend toast (AppLayout) only reaches a user who has the window open,
// and PlexCord is built to run in the background: with "Minimize to tray" or
// "Start minimized" there may be no window at all when an update lands. The
// tray menu is the surface that is always there, so it carries the same news.
func (a *App) publishUpdateNotice(status updater.Status) {
	if a.tray == nil {
		return
	}
	a.tray.SetUpdateNotice(updateNotice(status))
}

// updateNotice maps an updater status onto what the tray should say. Split from
// publishUpdateNotice so the wording and the actionable/not decision are
// testable without a tray.
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
		// The one notice worth acting on from the tray: the update is on disk
		// and a restart is all that is left.
		return platform.UpdateNotice{
			Label:      fmt.Sprintf("Restart to update to %s", version),
			Tooltip:    fmt.Sprintf("PlexCord %s is installed — restart to apply", version),
			Actionable: true,
		}
	case updater.StateDownloading:
		// Informational: the percentage lives in Settings, and a click here
		// would only interrupt work already under way.
		return platform.UpdateNotice{
			Label:   fmt.Sprintf("Downloading update %s…", version),
			Tooltip: fmt.Sprintf("PlexCord is downloading %s in the background", version),
		}
	case updater.StateAvailable:
		// Either a platform without self-update, or a download about to start.
		// Clicking opens the window, where the release notes and the download
		// button are.
		return platform.UpdateNotice{
			Label:      fmt.Sprintf("Update available: %s", version),
			Tooltip:    fmt.Sprintf("PlexCord %s is available", version),
			Actionable: true,
		}
	default:
		return platform.UpdateNotice{}
	}
}

// onTrayUpdateClick handles a click on the tray's update item: restart when the
// update is already installed, otherwise bring the window up so the user can
// see the release notes and decide.
func (a *App) onTrayUpdateClick() {
	if a.updater != nil && a.updater.GetStatus().State == updater.StateReady {
		if err := a.RestartApplication(); err != nil {
			log.Printf("ERROR: Failed to restart for update from tray: %v", err)
			// Fall through to showing the window: the user asked for the
			// update and now needs a surface that can report the failure.
			a.ShowWindow()
		}
		return
	}
	a.ShowWindow()
}
