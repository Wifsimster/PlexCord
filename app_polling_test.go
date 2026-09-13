package main

import (
	"errors"
	"testing"
	"time"

	"plexcord/internal/config"
	"plexcord/internal/plex"
)

// countingSource is a SessionSource that answers from memory, so the polling
// lifecycle can be driven with no Plex server anywhere.
type countingSource struct {
	sessions []plex.MusicSession
	err      error
}

func (s *countingSource) GetMusicSessions(string) ([]plex.MusicSession, error) {
	return s.sessions, s.err
}

func (s *countingSource) GetMediaSessions(string, []string) ([]plex.MediaSession, error) {
	return nil, s.err
}

func musicSession(track string) plex.MusicSession {
	m := plex.MusicSession{Track: track, Artist: "Artist", Album: "Album"}
	m.State = "playing"
	m.SessionKey = track
	return m
}

func TestPollingControllerStartsStopsAndReportsState(t *testing.T) {
	ctrl := &pollingController{}

	if ctrl.IsRunning() {
		t.Fatal("a fresh controller reports polling as running")
	}
	if running, inError := ctrl.State(); running || inError {
		t.Fatalf("State() = (%v, %v) before start, want (false, false)", running, inError)
	}

	ch := ctrl.Start(pollingConfig{
		Source:   &countingSource{sessions: []plex.MusicSession{musicSession("Song")}},
		UserID:   "user1",
		Interval: time.Second,
	})
	if ch == nil {
		t.Fatal("Start() returned no session channel")
	}
	defer ctrl.Stop()

	// The poller performs an immediate first poll, so the session arrives
	// without waiting out an interval.
	select {
	case session := <-ch:
		if session == nil || session.Track != "Song" {
			t.Fatalf("first session = %v, want the playing track", session)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no session arrived from the poller")
	}

	if !ctrl.IsRunning() {
		t.Error("IsRunning() = false while polling")
	}
}

func TestPollingControllerRefusesConcurrentStart(t *testing.T) {
	ctrl := &pollingController{}
	cfg := pollingConfig{Source: &countingSource{}, UserID: "user1", Interval: time.Second}

	if ch := ctrl.Start(cfg); ch == nil {
		t.Fatal("first Start() returned no channel")
	}
	defer ctrl.Stop()

	// A second start must report that nothing new was created, so the caller
	// does not spawn a second consumer goroutine on a channel nobody feeds.
	if ch := ctrl.Start(cfg); ch != nil {
		t.Error("second Start() created a second poller")
	}
}

func TestPollingControllerStopIsSafeWhenIdle(t *testing.T) {
	ctrl := &pollingController{}
	ctrl.Stop() // must not panic
	ctrl.Stop()
}

func TestPollingControllerSetIntervalOnlyWhileRunning(t *testing.T) {
	ctrl := &pollingController{}

	if ctrl.SetInterval(5 * time.Second) {
		t.Error("SetInterval() reported a retune with no poller running")
	}

	ctrl.Start(pollingConfig{Source: &countingSource{}, UserID: "user1", Interval: time.Second})
	defer ctrl.Stop()

	if !ctrl.SetInterval(5 * time.Second) {
		t.Error("SetInterval() = false while polling, want the running poller retuned")
	}
}

// TestPollingControllerSurfacesErrorState verifies a failing Plex server puts
// the controller into the error state the dashboard reads.
func TestPollingControllerSurfacesErrorState(t *testing.T) {
	ctrl := &pollingController{}
	failed := make(chan error, 1)

	ctrl.Start(pollingConfig{
		Source:   &countingSource{err: errors.New("plex unreachable")},
		UserID:   "user1",
		Interval: time.Second,
		OnError:  func(err error) { failed <- err },
	})
	defer ctrl.Stop()

	select {
	case <-failed:
	case <-time.After(2 * time.Second):
		t.Fatal("a failing poll never reported an error")
	}

	waitFor(t, func() bool {
		_, inError := ctrl.State()
		return inError
	}, "State() never reported the error state after a failing poll")
}

// TestStartSessionPollingUsesInjectedFactory is the regression this refactor
// was aimed at: session polling used to build a concrete plex.Client of its
// own, so it was the one Plex path that ignored the injected factory and could
// not be driven in a test.
func TestStartSessionPollingUsesInjectedFactory(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ServerURL = "http://plex.test:32400"
	cfg.SelectedPlexUserID = "user1"

	app := newTestApp(cfg)
	app.tokens = &fakeTokenStore{token: "a-token"}

	var gotToken, gotURL string
	app.plexFactory = func(token, serverURL string) PlexAPI {
		gotToken, gotURL = token, serverURL
		return &fakePlexAPI{music: []plex.MusicSession{musicSession("Song")}}
	}

	if err := app.StartSessionPolling(); err != nil {
		t.Fatalf("StartSessionPolling() error: %v", err)
	}
	defer app.StopSessionPolling()

	if gotToken != "a-token" || gotURL != "http://plex.test:32400" {
		t.Errorf("factory called with (%q, %q), want the stored token and configured server", gotToken, gotURL)
	}

	// The session reaches the cache the frontend restores from.
	waitFor(t, func() bool {
		s := app.GetCurrentSession()
		return s != nil && s.Track == "Song"
	}, "the polled session never reached the session cache")
}

func TestStartSessionPollingRejectsIncompleteConfiguration(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*App)
	}{
		{
			name:  "no server configured",
			setup: func(a *App) { a.config.SelectedPlexUserID = "user1" },
		},
		{
			name:  "no user selected",
			setup: func(a *App) { a.config.ServerURL = "http://plex.test:32400" },
		},
		{
			name: "no token stored",
			setup: func(a *App) {
				a.config.ServerURL = "http://plex.test:32400"
				a.config.SelectedPlexUserID = "user1"
				a.tokens = &fakeTokenStore{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(config.DefaultConfig())
			app.tokens = &fakeTokenStore{token: "a-token"}
			tt.setup(app)

			if err := app.StartSessionPolling(); err == nil {
				app.StopSessionPolling()
				t.Error("StartSessionPolling() = nil, want an error for an incomplete configuration")
			}
		})
	}
}
