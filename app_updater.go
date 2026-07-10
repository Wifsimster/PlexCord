package main

import (
	"log"

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
