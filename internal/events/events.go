// Package events provides a typed event bus abstraction for emitting
// frontend events. This centralizes event names and decouples the
// application core from the Wails runtime, enabling unit testing without
// a live Wails context.
package events

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Event name constants — the single source of truth for frontend event names.
// Any new event must be declared here and consumed through Bus.Emit.
const (
	PlaybackUpdated        = "PlaybackUpdated"
	PlaybackStopped        = "PlaybackStopped"
	PlexConnectionError    = "PlexConnectionError"
	PlexConnectionRestored = "PlexConnectionRestored"
	PlexRetryState         = "PlexRetryState"
	DiscordConnected       = "DiscordConnected"
	DiscordDisconnected    = "DiscordDisconnected"
	DiscordRetryState      = "DiscordRetryState"

	// Update lifecycle events emitted while an in-app update is downloaded
	// and applied. UpdateAvailable is emitted by the automatic update checker
	// when a newer release exists (before an auto-download starts, or instead
	// of one on platforms without self-update support).
	UpdateAvailable        = "UpdateAvailable"
	UpdateDownloadProgress = "UpdateDownloadProgress"
	UpdateReady            = "UpdateReady"
	UpdateError            = "UpdateError"
)

// Bus is the abstraction for emitting events to the frontend. Production code
// uses WailsBus; tests can use NewRecordingBus to assert on emissions without
// a Wails context.
type Bus interface {
	Emit(name string, payload ...interface{})
}

// WailsBus is the production implementation that forwards events to the
// Wails runtime. Construct via NewWailsBus.
type WailsBus struct {
	ctx context.Context
}

// NewWailsBus creates a Bus backed by the Wails runtime.
func NewWailsBus(ctx context.Context) *WailsBus {
	return &WailsBus{ctx: ctx}
}

// Emit forwards the event to the Wails runtime. Safe to call with nil ctx
// (becomes a no-op), which is useful during early startup before the Wails
// context is set.
func (b *WailsBus) Emit(name string, payload ...interface{}) {
	if b == nil || b.ctx == nil {
		return
	}
	runtime.EventsEmit(b.ctx, name, payload...)
}

// RecordingBus captures emitted events in memory for test assertions. It is
// safe for concurrent use: some producers (e.g. the automatic update checker)
// emit from a background goroutine while the test reads via Count/Snapshot.
type RecordingBus struct {
	mu     sync.Mutex
	Events []RecordedEvent
}

// RecordedEvent represents a single emitted event.
type RecordedEvent struct {
	Name    string
	Payload []interface{}
}

// NewRecordingBus creates a RecordingBus for use in tests.
func NewRecordingBus() *RecordingBus {
	return &RecordingBus{Events: []RecordedEvent{}}
}

// Emit records the event.
func (b *RecordingBus) Emit(name string, payload ...interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Events = append(b.Events, RecordedEvent{Name: name, Payload: payload})
}

// Reset clears all recorded events.
func (b *RecordingBus) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Events = b.Events[:0]
}

// Count returns the number of events matching the given name.
func (b *RecordingBus) Count(name string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := 0
	for _, e := range b.Events {
		if e.Name == name {
			n++
		}
	}
	return n
}

// Snapshot returns a copy of the recorded events, taken under the lock, so
// callers can iterate without racing a concurrent Emit.
func (b *RecordingBus) Snapshot() []RecordedEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]RecordedEvent, len(b.Events))
	copy(out, b.Events)
	return out
}
