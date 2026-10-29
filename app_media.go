package main

import (
	"log"

	"plexcord/internal/config"
	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/plex"
)

type MediaSyncSettings struct {
	Music  bool `json:"music"`
	Movies bool `json:"movies"`
	TV     bool `json:"tv"`
}

func (s MediaSyncSettings) mediaTypes() []string {
	types := make([]string, 0, len(config.AllMediaTypes))
	if s.Music {
		types = append(types, plex.MediaTypeMusic)
	}
	if s.Movies {
		types = append(types, plex.MediaTypeMovie)
	}
	if s.TV {
		types = append(types, plex.MediaTypeTV)
	}
	return types
}

func (a *App) GetMediaSync() MediaSyncSettings {
	if a.config == nil {
		return MediaSyncSettings{Music: true, Movies: true, TV: true}
	}
	return MediaSyncSettings{
		Music:  a.config.IsMediaTypeEnabled(plex.MediaTypeMusic),
		Movies: a.config.IsMediaTypeEnabled(plex.MediaTypeMovie),
		TV:     a.config.IsMediaTypeEnabled(plex.MediaTypeTV),
	}
}

func (a *App) SetMediaSync(settings MediaSyncSettings) error {
	types := settings.mediaTypes()
	if len(types) == 0 {
		return errors.New(errors.CONFIG_WRITE_FAILED, "at least one media type must be enabled")
	}

	a.config.PresenceMediaTypes = types
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save media sync settings: %v", err)
		return err
	}
	log.Printf("Media sync set to: %v", types)

	a.dropPresenceForDisabledMedia()
	a.restartPollingForMediaTypes()
	return nil
}

func (a *App) dropPresenceForDisabledMedia() {
	current := a.sessions.Get()
	if current == nil || a.config.IsMediaTypeEnabled(current.MediaType) {
		return
	}

	log.Printf("Withdrawing the %s presence: that kind is no longer synced", current.MediaType)
	a.sessions.Clear()
	a.presence.CancelHide()
	a.clearDiscordOnStop()
	a.bus.Emit(events.PlaybackStopped, nil)
}

func (a *App) restartPollingForMediaTypes() {
	if !a.polling.IsRunning() {
		return
	}
	a.StopSessionPolling()
	if err := a.StartSessionPolling(); err != nil {
		log.Printf("Warning: Failed to restart session polling after a media sync change: %v", err)
	}
}
