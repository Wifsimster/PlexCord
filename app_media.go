package main

import (
	"log"

	"plexcord/internal/config"
	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/plex"
)

// MediaSyncSettings is which kinds of Plex playback reach Discord. It is a
// struct of named booleans rather than the config's string slice because that
// is the shape the settings UI actually has — three switches — and it keeps the
// frontend from having to know the media-type spellings the poller uses.
type MediaSyncSettings struct {
	Music  bool `json:"music"`
	Movies bool `json:"movies"`
	TV     bool `json:"tv"`
}

// mediaTypes projects the settings onto the media-type list the poller takes.
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

// GetMediaSync returns which kinds of playback are relayed to Discord.
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

// SetMediaSync updates which kinds of playback are relayed to Discord.
//
// Turning everything off is rejected rather than silently stored: an empty list
// means "the default, all of them" everywhere else, so accepting it would show
// the user three switches off and then keep relaying all three. Pausing the
// presence entirely is what the pause toggle is for.
//
// The media types are fixed when a poller starts, so a running poller is
// restarted here, and whatever is showing on Discord is withdrawn first if the
// user just switched its kind off.
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

// dropPresenceForDisabledMedia takes down a presence whose kind the user has
// just switched off.
//
// The restarted poller cannot do this itself: it starts with no previous
// session, so "nothing matching is playing" reads as no change rather than as a
// stop, and the film the user just excluded would sit on their profile until
// they played something else.
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

// restartPollingForMediaTypes cycles a running poller so it picks up the new
// media types. A poller that is not running needs nothing: the next start reads
// the config.
func (a *App) restartPollingForMediaTypes() {
	if !a.polling.IsRunning() {
		return
	}
	a.StopSessionPolling()
	if err := a.StartSessionPolling(); err != nil {
		log.Printf("Warning: Failed to restart session polling after a media sync change: %v", err)
	}
}
