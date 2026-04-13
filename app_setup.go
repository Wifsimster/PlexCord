package main

import (
	"log"

	"plexcord/internal/config"
	"plexcord/internal/discord"
)

// CheckSetupComplete checks if the setup wizard has been completed
// This method is called from the frontend router to determine if
// the user should be redirected to the setup wizard on first launch
func (a *App) CheckSetupComplete() bool {
	return config.IsSetupComplete()
}

// CompleteSetup marks the setup wizard as complete and saves the configuration.
// This method is called from the frontend when the user finishes the setup wizard.
// It marks SetupCompleted as true in config.json so that subsequent app launches
// go directly to the dashboard instead of the setup wizard.
// Also starts session polling if server and user are configured.
// Also ensures Discord connection is active.
func (a *App) CompleteSetup() error {
	log.Printf("Completing setup wizard...")

	// Mark setup as complete
	a.config.SetupCompleted = true

	// Save config to disk
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save setup completion: %v", err)
		return err
	}

	log.Printf("Setup wizard completed successfully")

	// Ensure Discord is connected for Rich Presence
	a.discordMu.Lock()
	if !a.discord.IsConnected() {
		log.Printf("Discord not connected during setup completion, attempting to connect...")
		clientID := a.config.DiscordClientID
		if clientID == "" {
			clientID = discord.DefaultClientID
		}
		if err := a.discord.Connect(clientID); err != nil {
			log.Printf("Warning: Failed to connect to Discord after setup: %v", err)
			// Don't fail setup - user might not have Discord running
		} else {
			a.updateDiscordConnectionTime()
			log.Printf("Discord connected successfully after setup completion")
		}
	}
	a.discordMu.Unlock()

	// Try to start session polling if configured
	// This is non-blocking - errors are logged but don't fail setup completion
	if a.config.ServerURL != "" && a.config.SelectedPlexUserID != "" {
		if err := a.StartSessionPolling(); err != nil {
			log.Printf("Warning: Failed to start session polling after setup: %v", err)
			// Don't return error - setup is still complete
		}
	}

	return nil
}

// SkipSetup marks the setup wizard as skipped and saves partial configuration.
// This method is called from the frontend when the user chooses to skip setup.
// The user can complete setup later from the Settings page.
// Partial progress (token, server URL) is preserved for later completion.
func (a *App) SkipSetup() error {
	log.Printf("Skipping setup wizard...")

	// Mark setup as skipped (not completed)
	a.config.SetupSkipped = true
	a.config.SetupCompleted = false

	// Save config to disk
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save setup skip: %v", err)
		return err
	}

	log.Printf("Setup wizard skipped - user can complete later from settings")
	return nil
}

// ResetApplication resets PlexCord to its initial state.
// This method:
// - Stops session polling
// - Disconnects from Discord
// - Removes Plex token from secure storage
// - Removes auto-start registration
// - Deletes the configuration file
// - Resets in-memory config to defaults
// After calling this method, the setup wizard will be shown on next app launch.
// The application does NOT exit - the user can continue or restart.
func (a *App) ResetApplication() {
	log.Printf("Resetting application to initial state...")

	// 1. Stop session polling
	a.StopSessionPolling()

	// 2. Disconnect from Discord (clears presence)
	a.discordMu.Lock()
	if a.discord.IsConnected() {
		if err := a.discord.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect Discord: %v", err)
		}
	}
	a.discordMu.Unlock()

	// 3. Remove Plex token from secure storage
	if err := a.tokens.Delete(); err != nil {
		log.Printf("Warning: Failed to delete Plex token: %v", err)
		// Continue with reset - token deletion failure is not critical
	} else {
		log.Printf("Plex token removed from secure storage")
	}

	// 4. Remove auto-start registration
	if err := a.autostart.Disable(); err != nil {
		log.Printf("Warning: Failed to disable auto-start: %v", err)
		// Continue with reset - auto-start failure is not critical
	}

	// 5. Delete configuration file
	if err := config.Delete(); err != nil {
		log.Printf("Warning: Failed to delete config file: %v", err)
		// Continue with reset - we'll create fresh defaults
	} else {
		log.Printf("Configuration file deleted")
	}

	// 6. Reset in-memory config to defaults
	a.config = config.DefaultConfig()
	log.Printf("In-memory configuration reset to defaults")

	log.Printf("Application reset complete - setup wizard will show on next launch")
}
