package main

import (
	"log"
	"time"

	"plexcord/internal/events"
	"plexcord/internal/history"
	"plexcord/internal/plex"
)

type SessionObserver interface {
	OnUpdate(session *plex.MediaSession)
	OnStop()
}

type sessionCacheObserver struct {
	cache *sessionCache
}

func newSessionCacheObserver(cache *sessionCache) *sessionCacheObserver {
	return &sessionCacheObserver{cache: cache}
}

func (o *sessionCacheObserver) OnUpdate(session *plex.MediaSession) {
	o.cache.Set(session)
}

func (o *sessionCacheObserver) OnStop() {
	o.cache.Clear()
}

type historyObserver struct {
	store HistoryStore
}

func newHistoryObserver(store HistoryStore) *historyObserver {
	return &historyObserver{store: store}
}

func (o *historyObserver) OnUpdate(session *plex.MediaSession) {
	if o.store == nil {
		return
	}
	if session.MediaType != plex.MediaTypeMusic {
		return
	}
	o.store.Add(history.Entry{
		Track:     session.Title,
		Artist:    session.Artist,
		Album:     session.Album,
		Duration:  session.Duration,
		StartedAt: time.Now(),
		ThumbURL:  session.ThumbURL,
	})
}

func (o *historyObserver) OnStop() {
}

type eventEmitterObserver struct {
	bus events.Bus
}

func newEventEmitterObserver(bus events.Bus) *eventEmitterObserver {
	return &eventEmitterObserver{bus: bus}
}

func (o *eventEmitterObserver) OnUpdate(session *plex.MediaSession) {
	o.bus.Emit(events.PlaybackUpdated, session)
}

func (o *eventEmitterObserver) OnStop() {
	o.bus.Emit(events.PlaybackStopped, nil)
}

type discordPresenceObserver struct {
	update        func(session *plex.MediaSession)
	clearOnStop   func()
	isManualPause func() bool
	scheduleHide  func()
	cancelHide    func()
	hideOnPause   func() bool
	log           func(format string, args ...any)
}

func (o *discordPresenceObserver) OnUpdate(session *plex.MediaSession) {
	if o.isManualPause() {
		return
	}

	if session.State == "paused" && o.hideOnPause() {
		o.scheduleHide()
		return
	}

	o.cancelHide()
	o.update(session)
}

func (o *discordPresenceObserver) OnStop() {
	o.cancelHide()
	o.clearOnStop()
}

func sessionLogLine(session *plex.MediaSession) string {
	switch session.MediaType {
	case plex.MediaTypeMusic:
		return session.Title + " - " + session.Artist
	case plex.MediaTypeTV:
		return session.ShowTitle + " - " + session.Title
	default:
		return session.Title
	}
}

func runSessionPipeline(sessionCh <-chan *plex.MediaSession, observers []SessionObserver) {
	var lastSession *plex.MediaSession

	for session := range sessionCh {
		switch {
		case session != nil:
			log.Printf("Playback detected (%s): %s", session.MediaType, sessionLogLine(session))
			for _, o := range observers {
				o.OnUpdate(session)
			}
			lastSession = session
		case lastSession != nil:
			log.Printf("Playback stopped")
			for _, o := range observers {
				o.OnStop()
			}
			lastSession = nil
		}
	}

	log.Printf("Session update handler exited")
}
