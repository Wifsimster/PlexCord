package main

import (
	"log"
	"net/url"

	"plexcord/internal/config"
	"plexcord/internal/errors"
)

// GetServers returns the list of configured Plex servers from config.
// Returns an empty slice (never nil) so the frontend can iterate safely.
func (a *App) GetServers() []config.ServerConfig {
	if a.config.Servers == nil {
		return []config.ServerConfig{}
	}
	return a.config.Servers
}

// AddServer appends a new server to the configuration. The URL must use
// http or https and must be unique within the existing server list.
// userID and userName are optional and may be filled in later via the
// per-server user-selection flow.
func (a *App) AddServer(name, serverURL, userID, userName string) error {
	if name == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server name cannot be empty")
	}
	if serverURL == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL cannot be empty")
	}
	parsed, err := url.Parse(serverURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL must use http or https scheme")
	}
	for _, s := range a.config.Servers {
		if s.URL == serverURL {
			return errors.New(errors.CONFIG_WRITE_FAILED, "server with this URL already exists")
		}
	}

	a.config.Servers = append(a.config.Servers, config.ServerConfig{
		Name:     name,
		URL:      serverURL,
		UserID:   userID,
		UserName: userName,
		Active:   true,
	})
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save server: %v", err)
		return err
	}
	log.Printf("Server added: %s (%s)", name, serverURL)
	return nil
}

// RemoveServer removes the server with the given URL from the configuration.
// Returns an error if no server with that URL is configured.
func (a *App) RemoveServer(serverURL string) error {
	if serverURL == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL cannot be empty")
	}
	idx := -1
	for i, s := range a.config.Servers {
		if s.URL == serverURL {
			idx = i
			break
		}
	}
	if idx < 0 {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server not found")
	}
	a.config.Servers = append(a.config.Servers[:idx], a.config.Servers[idx+1:]...)
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save servers after removal: %v", err)
		return err
	}
	log.Printf("Server removed: %s", serverURL)
	return nil
}

// SetServerActive toggles a server's Active flag.
// Returns an error if no server with that URL is configured.
func (a *App) SetServerActive(serverURL string, active bool) error {
	if serverURL == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL cannot be empty")
	}
	for i := range a.config.Servers {
		if a.config.Servers[i].URL == serverURL {
			a.config.Servers[i].Active = active
			if err := a.saveConfig(); err != nil {
				log.Printf("ERROR: Failed to save server active state: %v", err)
				return err
			}
			log.Printf("Server %s active=%v", serverURL, active)
			return nil
		}
	}
	return errors.New(errors.CONFIG_WRITE_FAILED, "server not found")
}
