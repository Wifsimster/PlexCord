package main

import (
	"log"

	"plexcord/internal/config"
)

func (a *App) CheckSetupComplete() bool {
	return a.configs.IsSetupComplete()
}

func (a *App) CompleteSetup() error {
	log.Printf("Completing setup wizard...")

	a.config.SetupCompleted = true

	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save setup completion: %v", err)
		return err
	}

	log.Printf("Setup wizard completed successfully")

	if !a.discord.IsConnected() {
		log.Printf("Discord not connected during setup completion, attempting to connect...")
		if a.discord.ConnectIfDown() {
			log.Printf("Discord connected successfully after setup completion")
		} else {
			log.Printf("Warning: Failed to connect to Discord after setup")
		}
	}

	if a.config.ServerURL != "" && a.config.SelectedPlexUserID != "" {
		if err := a.StartSessionPolling(); err != nil {
			log.Printf("Warning: Failed to start session polling after setup: %v", err)
		}
	}

	return nil
}

func (a *App) SkipSetup() error {
	log.Printf("Skipping setup wizard...")

	a.config.SetupSkipped = true
	a.config.SetupCompleted = false

	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save setup skip: %v", err)
		return err
	}

	log.Printf("Setup wizard skipped - user can complete later from settings")
	return nil
}

func (a *App) ResetApplication() {
	log.Printf("Resetting application to initial state...")

	a.StopSessionPolling()

	if a.discord.IsConnected() {
		if err := a.discord.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect Discord: %v", err)
		}
	}

	if err := a.tokens.Delete(); err != nil {
		log.Printf("Warning: Failed to delete Plex token: %v", err)
	} else {
		log.Printf("Plex token removed from secure storage")
	}

	if err := a.autostart.Disable(); err != nil {
		log.Printf("Warning: Failed to disable auto-start: %v", err)
	}

	if err := a.configs.Delete(); err != nil {
		log.Printf("Warning: Failed to delete config file: %v", err)
	} else {
		log.Printf("Configuration file deleted")
	}

	a.cfgStore.UpdateNoSave(func(c *config.Config) { *c = *config.DefaultConfig() })
	log.Printf("In-memory configuration reset to defaults")

	log.Printf("Application reset complete - setup wizard will show on next launch")
}
