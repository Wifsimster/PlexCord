package main

import (
	"log"
	"net/url"

	"plexcord/internal/config"
	"plexcord/internal/errors"
)

func (a *App) GetServers() []config.ServerConfig {
	if a.cfgStore == nil {
		if a.config.Servers == nil {
			return []config.ServerConfig{}
		}
		return a.config.Servers
	}
	cfg := a.cfgStore.Get()
	if len(cfg.Servers) == 0 {
		return []config.ServerConfig{}
	}
	out := make([]config.ServerConfig, len(cfg.Servers))
	copy(out, cfg.Servers)
	return out
}

func (a *App) activePlexServerURL() string {
	for _, s := range a.config.Servers {
		if s.Active && s.URL != "" {
			return s.URL
		}
	}
	return a.config.ServerURL
}

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

	var dup bool
	if err := a.cfgStore.Update(func(c *config.Config) {
		for _, s := range c.Servers {
			if s.URL == serverURL {
				dup = true
				return
			}
		}
		c.Servers = append(c.Servers, config.ServerConfig{
			Name:     name,
			URL:      serverURL,
			UserID:   userID,
			UserName: userName,
			Active:   true,
		})
	}); err != nil {
		log.Printf("ERROR: Failed to save server: %v", err)
		return err
	}
	if dup {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server with this URL already exists")
	}
	log.Printf("Server added: %s (%s)", name, serverURL)
	return nil
}

func (a *App) RemoveServer(serverURL string) error {
	if serverURL == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL cannot be empty")
	}

	var found bool
	if err := a.cfgStore.Update(func(c *config.Config) {
		for i, s := range c.Servers {
			if s.URL == serverURL {
				c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
				found = true
				return
			}
		}
	}); err != nil {
		log.Printf("ERROR: Failed to save servers after removal: %v", err)
		return err
	}
	if !found {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server not found")
	}
	log.Printf("Server removed: %s", serverURL)
	return nil
}

func (a *App) SetServerActive(serverURL string, active bool) error {
	if serverURL == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL cannot be empty")
	}

	var found bool
	if err := a.cfgStore.Update(func(c *config.Config) {
		for i := range c.Servers {
			if c.Servers[i].URL == serverURL {
				c.Servers[i].Active = active
				found = true
				return
			}
		}
	}); err != nil {
		log.Printf("ERROR: Failed to save server active state: %v", err)
		return err
	}
	if !found {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server not found")
	}
	log.Printf("Server %s active=%v", serverURL, active)
	return nil
}
