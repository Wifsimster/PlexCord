package main

import (
	"sync/atomic"
	"testing"
	"time"
)

// clearRecorder counts presence clears for the gate tests.
type clearRecorder struct{ n atomic.Int64 }

func (c *clearRecorder) clear()       { c.n.Add(1) }
func (c *clearRecorder) count() int64 { return c.n.Load() }

func TestPresenceGateTogglesPause(t *testing.T) {
	rec := &clearRecorder{}
	gate := newPresenceGate(rec.clear, func() time.Duration { return 0 })

	if gate.IsPaused() {
		t.Fatal("a fresh gate reports presence as paused")
	}
	if !gate.Toggle() {
		t.Fatal("Toggle() = false on the first call, want true")
	}
	if !gate.IsPaused() {
		t.Error("IsPaused() = false after toggling on")
	}
	if gate.Toggle() {
		t.Error("Toggle() = true on the second call, want false")
	}
}

func TestPresenceGateHidesImmediatelyWithZeroDelay(t *testing.T) {
	rec := &clearRecorder{}
	gate := newPresenceGate(rec.clear, func() time.Duration { return 0 })

	gate.ScheduleHide()

	// A zero delay clears on its own goroutine; give it a moment to land.
	waitFor(t, func() bool { return rec.count() == 1 }, "presence was never cleared with a zero delay")
}

func TestPresenceGateHidesAfterDelay(t *testing.T) {
	rec := &clearRecorder{}
	gate := newPresenceGate(rec.clear, func() time.Duration { return 30 * time.Millisecond })

	gate.ScheduleHide()

	if got := rec.count(); got != 0 {
		t.Fatalf("presence cleared before the delay elapsed (%d clears)", got)
	}
	waitFor(t, func() bool { return rec.count() == 1 }, "presence was never cleared after the delay")
}

// TestPresenceGateCancelBeatsFiringTimer is the race the generation counter
// exists for: playback resumes just as the hide timer fires, and the stale
// callback must not clear the presence that was just restored.
func TestPresenceGateCancelBeatsFiringTimer(t *testing.T) {
	rec := &clearRecorder{}
	gate := newPresenceGate(rec.clear, func() time.Duration { return 20 * time.Millisecond })

	gate.ScheduleHide()
	gate.CancelHide()

	time.Sleep(60 * time.Millisecond)
	if got := rec.count(); got != 0 {
		t.Errorf("a cancelled hide still cleared the presence (%d clears)", got)
	}
}

// TestPresenceGateRescheduleSupersedesPrevious verifies a second schedule
// replaces the first rather than stacking a second clear behind it.
func TestPresenceGateRescheduleSupersedesPrevious(t *testing.T) {
	rec := &clearRecorder{}
	gate := newPresenceGate(rec.clear, func() time.Duration { return 25 * time.Millisecond })

	gate.ScheduleHide()
	gate.ScheduleHide()

	waitFor(t, func() bool { return rec.count() >= 1 }, "presence was never cleared")
	time.Sleep(40 * time.Millisecond)
	if got := rec.count(); got != 1 {
		t.Errorf("clears = %d, want exactly 1 — the first schedule should have been superseded", got)
	}
}

// waitFor polls cond until it holds or the deadline passes.
func waitFor(t *testing.T, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal(msg)
}
