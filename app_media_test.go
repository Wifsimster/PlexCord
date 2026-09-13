package main

import (
	"testing"

	"plexcord/internal/config"
	"plexcord/internal/plex"
)

func TestGetMediaSyncDefaultsToEverything(t *testing.T) {
	// A config written before PlexCord handled video carries no media types at
	// all; the feature must arrive switched on rather than hidden.
	cfg := config.DefaultConfig()
	cfg.PresenceMediaTypes = nil
	app := newTestApp(cfg)

	got := app.GetMediaSync()
	if !got.Music || !got.Movies || !got.TV {
		t.Errorf("GetMediaSync() = %+v, want every kind enabled for an unset config", got)
	}
}

func TestSetMediaSyncPersistsTheSelection(t *testing.T) {
	app := newTestApp(config.DefaultConfig())

	if err := app.SetMediaSync(MediaSyncSettings{Music: true, TV: true}); err != nil {
		t.Fatalf("SetMediaSync() error: %v", err)
	}

	want := []string{plex.MediaTypeMusic, plex.MediaTypeTV}
	got := app.config.EnabledMediaTypes()
	if len(got) != len(want) {
		t.Fatalf("enabled media types = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("enabled media types = %v, want %v", got, want)
		}
	}
	if app.GetMediaSync().Movies {
		t.Error("GetMediaSync() still reports movies enabled after they were switched off")
	}
}

func TestSetMediaSyncRejectsSwitchingEverythingOff(t *testing.T) {
	// An empty list means "all of them" everywhere else, so storing one would
	// show three switches off while still relaying all three.
	app := newTestApp(config.DefaultConfig())

	if err := app.SetMediaSync(MediaSyncSettings{}); err == nil {
		t.Fatal("SetMediaSync() accepted a selection with nothing enabled")
	}
	if got := app.GetMediaSync(); !got.Music {
		t.Errorf("a rejected call still changed the stored selection: %+v", got)
	}
}

func TestStartSessionPollingAsksPlexForTheConfiguredMediaTypes(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ServerURL = "http://plex.test:32400"
	cfg.SelectedPlexUserID = "user1"
	cfg.PresenceMediaTypes = []string{plex.MediaTypeMovie, plex.MediaTypeTV}

	app := newTestApp(cfg)
	app.tokens = &fakeTokenStore{token: "a-token"}

	api := &fakePlexAPI{media: []plex.MediaSession{movieSession("Blade Runner")}}
	app.plexFactory = func(string, string) PlexAPI { return api }

	if err := app.StartSessionPolling(); err != nil {
		t.Fatalf("StartSessionPolling() error: %v", err)
	}
	defer app.StopSessionPolling()

	waitFor(t, func() bool {
		s := app.GetCurrentSession()
		return s != nil && s.MediaType == plex.MediaTypeMovie
	}, "the polled film never reached the session cache")

	got := api.lastTypes
	if len(got) != 2 || got[0] != plex.MediaTypeMovie || got[1] != plex.MediaTypeTV {
		t.Errorf("Plex was asked for media types %v, want the configured [movie tv]", got)
	}
}

// movieSession is a film playing, as the Plex client would report it.
func movieSession(title string) plex.MediaSession {
	return plex.MediaSession{
		SessionKey: title,
		Type:       "movie",
		MediaType:  plex.MediaTypeMovie,
		State:      "playing",
		Title:      title,
		Year:       1982,
		Duration:   7_000_000,
	}
}

// TestSetMediaSyncWithdrawsAPresenceOfANowDisabledKind covers the gap a plain
// poller restart leaves: the restarted poller has no previous session, so
// "nothing matching is playing" reads as no change rather than as a stop, and
// the film the user just excluded would sit on their Discord profile.
func TestSetMediaSyncWithdrawsAPresenceOfANowDisabledKind(t *testing.T) {
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, nil)

	film := movieSession("Blade Runner")
	app.sessions.Set(&film)

	if err := app.SetMediaSync(MediaSyncSettings{Music: true}); err != nil {
		t.Fatalf("SetMediaSync() error: %v", err)
	}

	if presence.clearCount() != 1 {
		t.Errorf("Discord clears = %d, want 1 — the film stayed on the profile", presence.clearCount())
	}
	if app.GetCurrentSession() != nil {
		t.Error("the withdrawn session is still cached for the dashboard to restore")
	}
}

// TestSetMediaSyncLeavesAStillEnabledPresenceAlone is the other half: turning
// films off must not interrupt the track that is playing.
func TestSetMediaSyncLeavesAStillEnabledPresenceAlone(t *testing.T) {
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, nil)

	track := musicSession("Song")
	app.sessions.Set(&track)

	if err := app.SetMediaSync(MediaSyncSettings{Music: true}); err != nil {
		t.Fatalf("SetMediaSync() error: %v", err)
	}

	if presence.clearCount() != 0 {
		t.Errorf("Discord clears = %d, want 0 — a still-synced track was interrupted", presence.clearCount())
	}
	if app.GetCurrentSession() == nil {
		t.Error("the playing track was dropped from the session cache")
	}
}
