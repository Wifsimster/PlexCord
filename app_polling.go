package main

import (
	"context"
	"log"
	"sync"
	"time"

	"plexcord/internal/plex"
)

// sessionCache holds the session currently playing so the frontend can restore
// its dashboard after a page refresh. It replaces a raw
// (*sync.RWMutex, **plex.MusicSession) pair threaded through the observer:
// the lock and the value it guards now travel together, and no caller can hold
// one without the other.
type sessionCache struct {
	mu      sync.RWMutex
	current *plex.MusicSession
}

// Set records the session now playing.
func (c *sessionCache) Set(session *plex.MusicSession) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = session
}

// Clear forgets the session, for when playback stops.
func (c *sessionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = nil
}

// Get returns the session currently playing, or nil.
func (c *sessionCache) Get() *plex.MusicSession {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current
}

// pollingController owns the session poller's lifecycle: starting it against a
// Plex client, stopping it, reporting its state, and retuning its interval
// while it runs.
//
// App previously carried the poller, its context, its cancel function and the
// mutex over all three, interleaved with Discord state and settings. Those are
// four fields that only ever change together, for one reason — whether and how
// PlexCord is watching Plex — so they live together here (SRP).
type pollingController struct {
	mu     sync.Mutex
	poller *plex.Poller
	stop   context.CancelFunc
}

// pollingConfig is everything needed to start a poller. Grouping it keeps the
// Start signature from growing a parameter every time the poller learns a new
// trick.
type pollingConfig struct {
	// Source supplies the sessions (the Plex client in production).
	Source plex.SessionSource
	// UserID is the Plex account whose playback is watched.
	UserID string
	// Interval is the poll period; the poller clamps it to [1s, 60s].
	Interval time.Duration
	// OnError fires on the transition into a failed-connection state.
	OnError func(err error)
	// OnRecovered fires when polling succeeds again after OnError.
	OnRecovered func()
}

// Start begins polling and returns the channel of session updates. It returns
// nil when a poller is already running, so the caller knows there is nothing
// new to consume.
func (p *pollingController) Start(cfg pollingConfig) <-chan *plex.MusicSession {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.poller != nil && p.poller.IsRunning() {
		log.Printf("Session polling already running")
		return nil
	}

	poller := plex.NewPoller(cfg.Source, cfg.UserID, cfg.Interval)
	poller.SetErrorCallbacks(cfg.OnError, cfg.OnRecovered)

	ctx, cancel := context.WithCancel(context.Background())
	p.poller = poller
	p.stop = cancel

	return poller.Start(ctx)
}

// Stop stops polling. Safe to call when nothing is running.
func (p *pollingController) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.poller == nil {
		return
	}
	if p.stop != nil {
		p.stop()
	}
	p.poller.Stop()
	p.poller = nil
	p.stop = nil

	log.Printf("Session polling stopped")
}

// IsRunning reports whether polling is active.
func (p *pollingController) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.poller != nil && p.poller.IsRunning()
}

// State reports whether polling is running and whether it is currently failing.
// Both are read under one lock so the dashboard never sees a torn pair.
func (p *pollingController) State() (running, inError bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.poller == nil {
		return false, false
	}
	return p.poller.IsRunning(), p.poller.IsInErrorState()
}

// SetInterval retunes a running poller. A no-op when nothing is running — the
// next Start picks the new interval up from the config.
func (p *pollingController) SetInterval(interval time.Duration) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.poller == nil || !p.poller.IsRunning() {
		return false
	}
	p.poller.SetInterval(interval)
	return true
}
