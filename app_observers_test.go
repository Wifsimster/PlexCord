package main

import (
	"testing"

	"plexcord/internal/events"
	"plexcord/internal/plex"
)

func TestSessionCacheObserver_StoresAndClearsSession(t *testing.T) {
	cache := &sessionCache{}
	obs := newSessionCacheObserver(cache)

	session := &plex.MediaSession{MediaType: plex.MediaTypeMusic, Title: "Song", Artist: "Artist"}
	obs.OnUpdate(session)

	if got := cache.Get(); got == nil || got.Title != "Song" {
		t.Errorf("expected cached session, got %v", got)
	}

	obs.OnStop()
	if got := cache.Get(); got != nil {
		t.Errorf("expected nil after stop, got %v", got)
	}
}

func TestEventEmitterObserver_EmitsEvents(t *testing.T) {
	bus := events.NewRecordingBus()
	obs := newEventEmitterObserver(bus)

	session := &plex.MediaSession{MediaType: plex.MediaTypeMusic, Title: "Song"}
	obs.OnUpdate(session)
	obs.OnStop()

	if bus.Count(events.PlaybackUpdated) != 1 {
		t.Errorf("expected 1 PlaybackUpdated, got %d", bus.Count(events.PlaybackUpdated))
	}
	if bus.Count(events.PlaybackStopped) != 1 {
		t.Errorf("expected 1 PlaybackStopped, got %d", bus.Count(events.PlaybackStopped))
	}
}

func TestDiscordPresenceObserver_SkipsWhenManuallyPaused(t *testing.T) {
	updateCalled := false
	clearCalled := false
	obs := &discordPresenceObserver{
		update:        func(*plex.MediaSession) { updateCalled = true },
		clearOnStop:   func() { clearCalled = true },
		isManualPause: func() bool { return true },
		scheduleHide:  func() {},
		cancelHide:    func() {},
		hideOnPause:   func() bool { return false },
		log:           func(string, ...any) {},
	}

	obs.OnUpdate(&plex.MediaSession{MediaType: plex.MediaTypeMusic, State: "playing", Title: "Song"})

	if updateCalled {
		t.Error("update should not be called when manually paused")
	}
	if clearCalled {
		t.Error("clear should not be called on update")
	}
}

func TestDiscordPresenceObserver_SchedulesHideWhenPaused(t *testing.T) {
	updateCalled := false
	scheduleCalled := false
	cancelCalled := false
	obs := &discordPresenceObserver{
		update:        func(*plex.MediaSession) { updateCalled = true },
		clearOnStop:   func() {},
		isManualPause: func() bool { return false },
		scheduleHide:  func() { scheduleCalled = true },
		cancelHide:    func() { cancelCalled = true },
		hideOnPause:   func() bool { return true },
		log:           func(string, ...any) {},
	}

	obs.OnUpdate(&plex.MediaSession{MediaType: plex.MediaTypeMusic, State: "paused", Title: "Song"})

	if updateCalled {
		t.Error("update should not be called when paused + hideOnPause")
	}
	if !scheduleCalled {
		t.Error("scheduleHide should be called when paused + hideOnPause")
	}
	if cancelCalled {
		t.Error("cancelHide should NOT be called when entering paused state")
	}
}

func TestDiscordPresenceObserver_UpdatesOnPlayAfterPause(t *testing.T) {
	updateCalled := false
	cancelCalled := false
	obs := &discordPresenceObserver{
		update:        func(*plex.MediaSession) { updateCalled = true },
		clearOnStop:   func() {},
		isManualPause: func() bool { return false },
		scheduleHide:  func() {},
		cancelHide:    func() { cancelCalled = true },
		hideOnPause:   func() bool { return true },
		log:           func(string, ...any) {},
	}

	obs.OnUpdate(&plex.MediaSession{MediaType: plex.MediaTypeMusic, State: "playing", Title: "Song"})

	if !updateCalled {
		t.Error("update should be called when playing")
	}
	if !cancelCalled {
		t.Error("cancelHide should be called on play to cancel any pending hide")
	}
}

func TestRunSessionPipeline_DispatchesInOrder(t *testing.T) {
	var order []string
	recorder := func(name string) SessionObserver {
		return &fakeObserver{
			updateFn: func(*plex.MediaSession) { order = append(order, name+":update") },
			stopFn:   func() { order = append(order, name+":stop") },
		}
	}

	ch := make(chan *plex.MediaSession, 3)
	ch <- &plex.MediaSession{MediaType: plex.MediaTypeMusic, Title: "A"}
	ch <- nil
	close(ch)

	runSessionPipeline(ch, []SessionObserver{recorder("obs1"), recorder("obs2")})

	expected := []string{"obs1:update", "obs2:update", "obs1:stop", "obs2:stop"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d dispatches, got %d: %v", len(expected), len(order), order)
	}
	for i, got := range order {
		if got != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], got)
		}
	}
}

type fakeObserver struct {
	updateFn func(*plex.MediaSession)
	stopFn   func()
}

func (f *fakeObserver) OnUpdate(s *plex.MediaSession) { f.updateFn(s) }
func (f *fakeObserver) OnStop()                       { f.stopFn() }

// TestHistoryObserverRecordsMusicOnly verifies a film passing through the same
// pipeline does not land in the listening history, where it would make the
// "most played artist" statistic nonsense.
func TestHistoryObserverRecordsMusicOnly(t *testing.T) {
	store := &fakeHistory{}
	obs := newHistoryObserver(store)

	obs.OnUpdate(&plex.MediaSession{MediaType: plex.MediaTypeMusic, Title: "Song", Artist: "Artist"})
	obs.OnUpdate(&plex.MediaSession{MediaType: plex.MediaTypeMovie, Title: "Blade Runner"})
	obs.OnUpdate(&plex.MediaSession{MediaType: plex.MediaTypeTV, Title: "Good News", ShowTitle: "Severance"})

	entries := store.GetRecent(10)
	if len(entries) != 1 {
		t.Fatalf("history holds %d entries, want only the track", len(entries))
	}
	if entries[0].Track != "Song" {
		t.Errorf("history entry = %q, want the track title", entries[0].Track)
	}
}
