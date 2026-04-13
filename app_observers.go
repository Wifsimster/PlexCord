package main

import (
	"log"
	"sync"
	"time"

	"plexcord/internal/events"
	"plexcord/internal/history"
	"plexcord/internal/plex"
)

// SessionObserver is a component that reacts to session changes from the
// poller. The session update handler runs each observer's OnUpdate method
// when a session starts/changes, and OnStop when playback stops.
//
// By decomposing the previous god function handleSessionUpdates into a
// pipeline of small observers, each responsibility becomes independently
// testable and new observers (Last.fm scrobbling, OS notifications, etc.)
// can be added without modifying existing code (OCP).
type SessionObserver interface {
	// OnUpdate is called when a new or changed session is received.
	OnUpdate(session *plex.MusicSession)
	// OnStop is called when playback stops (session becomes nil).
	OnStop()
}

// ----------------------------------------------------------------------------
// sessionCacheObserver stores the current session for page refresh restoration
// ----------------------------------------------------------------------------
type sessionCacheObserver struct {
	mu      *sync.RWMutex
	current **plex.MusicSession
}

func newSessionCacheObserver(mu *sync.RWMutex, current **plex.MusicSession) *sessionCacheObserver {
	return &sessionCacheObserver{mu: mu, current: current}
}

func (o *sessionCacheObserver) OnUpdate(session *plex.MusicSession) {
	o.mu.Lock()
	defer o.mu.Unlock()
	*o.current = session
}

func (o *sessionCacheObserver) OnStop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	*o.current = nil
}

// ----------------------------------------------------------------------------
// historyObserver records played tracks to the listening history
// ----------------------------------------------------------------------------
type historyObserver struct {
	store *history.Store
}

func newHistoryObserver(store *history.Store) *historyObserver {
	return &historyObserver{store: store}
}

func (o *historyObserver) OnUpdate(session *plex.MusicSession) {
	if o.store == nil {
		return
	}
	o.store.Add(history.Entry{
		Track:     session.Track,
		Artist:    session.Artist,
		Album:     session.Album,
		Duration:  session.Duration,
		StartedAt: time.Now(),
		ThumbURL:  session.ThumbURL,
	})
}

func (o *historyObserver) OnStop() {
	// History doesn't react to stops — entries are added on play start.
}

// ----------------------------------------------------------------------------
// eventEmitterObserver broadcasts Playback events to the frontend
// ----------------------------------------------------------------------------
type eventEmitterObserver struct {
	bus events.Bus
}

func newEventEmitterObserver(bus events.Bus) *eventEmitterObserver {
	return &eventEmitterObserver{bus: bus}
}

func (o *eventEmitterObserver) OnUpdate(session *plex.MusicSession) {
	o.bus.Emit(events.PlaybackUpdated, session)
}

func (o *eventEmitterObserver) OnStop() {
	o.bus.Emit(events.PlaybackStopped, nil)
}

// ----------------------------------------------------------------------------
// discordPresenceObserver updates Discord Rich Presence (with pause-gating)
// ----------------------------------------------------------------------------
//
// This observer is gated by the App's manual-pause flag and the
// hide-when-paused timer. Rather than embedding that logic inline,
// we delegate to small hook functions the App provides, keeping the
// observer testable without the App.
type discordPresenceObserver struct {
	update        func(session *plex.MusicSession) // wraps updateDiscordFromSession
	clearOnStop   func()                           // wraps clearDiscordOnStop
	isManualPause func() bool                      // returns true when presence paused
	scheduleHide  func()                           // schedules hide-when-paused timer
	cancelHide    func()                           // cancels hide-when-paused timer
	hideOnPause   func() bool                      // returns config.HideWhenPaused
	log           func(format string, args ...any)
}

func (o *discordPresenceObserver) OnUpdate(session *plex.MusicSession) {
	if o.isManualPause() {
		// Manually paused — skip presence updates entirely
		return
	}

	if session.State == "paused" && o.hideOnPause() {
		// Schedule hide-when-paused; leave presence as-is for now
		o.scheduleHide()
		return
	}

	// Cancel any pending hide timer since we're actively playing
	o.cancelHide()
	o.update(session)
}

func (o *discordPresenceObserver) OnStop() {
	o.cancelHide()
	o.clearOnStop()
}

// ----------------------------------------------------------------------------
// Pipeline runner
// ----------------------------------------------------------------------------

// runSessionPipeline consumes session updates from the channel and dispatches
// each to the ordered list of observers. This replaces the previous
// handleSessionUpdates god function. The loop exits when the channel closes.
func runSessionPipeline(sessionCh <-chan *plex.MusicSession, observers []SessionObserver) {
	var lastSession *plex.MusicSession

	for session := range sessionCh {
		switch {
		case session != nil:
			log.Printf("Playback detected: %s - %s", session.Track, session.Artist)
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
