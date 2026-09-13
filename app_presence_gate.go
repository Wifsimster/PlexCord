package main

import (
	"log"
	"sync"
	"time"
)

// presenceGate decides whether a presence update reaches Discord, and owns the
// state behind that decision: the manual pause toggle and the hide-when-paused
// timer.
//
// It is separate from App because pausing has its own reasons to change (the
// tray toggle, the delay setting, the generation guard against a stale timer)
// and none of them are about connecting to Discord or reading Plex (SRP). It
// takes the clear action and the delay as collaborators rather than reaching
// back into App, so it can be tested on its own.
type presenceGate struct {
	// clear hides the presence currently showing.
	clear func()
	// hideDelay reports the configured hide-when-paused delay.
	hideDelay func() time.Duration

	mu     sync.Mutex
	paused bool
	timer  *time.Timer
	// timerGen is bumped on every schedule and cancel. timer.Stop() does not
	// wait for an already-firing callback, so a stale callback could race a
	// subsequent play-resume and clear the presence just restored; the callback
	// compares the generation it was armed with and bails out if it moved.
	timerGen uint64
}

// newPresenceGate builds a gate over the given clear action and delay source.
func newPresenceGate(clear func(), hideDelay func() time.Duration) *presenceGate {
	return &presenceGate{clear: clear, hideDelay: hideDelay}
}

// IsPaused reports whether presence updates are manually paused.
func (g *presenceGate) IsPaused() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.paused
}

// Toggle flips the manual pause state and returns the new value. It does not
// itself clear or restore the presence — the caller decides what that means.
func (g *presenceGate) Toggle() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.paused = !g.paused
	return g.paused
}

// ScheduleHide arms the hide-when-paused timer, replacing any pending one. A
// delay of zero or less clears immediately.
func (g *presenceGate) ScheduleHide() {
	g.mu.Lock()
	g.stopTimerLocked()
	g.timerGen++
	gen := g.timerGen
	g.mu.Unlock()

	delay := g.hideDelay()
	if delay <= 0 {
		go g.clear()
		return
	}

	timer := time.AfterFunc(delay, func() {
		g.mu.Lock()
		stale := gen != g.timerGen
		g.mu.Unlock()
		if stale {
			return
		}
		log.Printf("Hide-when-paused delay elapsed, clearing presence")
		g.clear()
	})

	g.mu.Lock()
	// A cancel that landed between arming and here already moved the
	// generation; drop this timer rather than publishing it.
	if gen == g.timerGen {
		g.timer = timer
	} else {
		timer.Stop()
	}
	g.mu.Unlock()
}

// CancelHide cancels any pending hide-when-paused timer and bumps the
// generation so an already-firing callback returns without clearing presence.
func (g *presenceGate) CancelHide() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.stopTimerLocked()
	g.timerGen++
}

// stopTimerLocked stops and drops the pending timer. Caller holds mu.
func (g *presenceGate) stopTimerLocked() {
	if g.timer != nil {
		g.timer.Stop()
		g.timer = nil
	}
}
