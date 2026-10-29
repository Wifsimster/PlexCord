package main

import (
	"log"
	"sync"
	"time"
)

type presenceGate struct {
	clear     func()
	hideDelay func() time.Duration

	mu       sync.Mutex
	paused   bool
	timer    *time.Timer
	timerGen uint64
}

func newPresenceGate(clear func(), hideDelay func() time.Duration) *presenceGate {
	return &presenceGate{clear: clear, hideDelay: hideDelay}
}

func (g *presenceGate) IsPaused() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.paused
}

func (g *presenceGate) Toggle() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.paused = !g.paused
	return g.paused
}

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
	if gen == g.timerGen {
		g.timer = timer
	} else {
		timer.Stop()
	}
	g.mu.Unlock()
}

func (g *presenceGate) CancelHide() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.stopTimerLocked()
	g.timerGen++
}

func (g *presenceGate) stopTimerLocked() {
	if g.timer != nil {
		g.timer.Stop()
		g.timer = nil
	}
}
