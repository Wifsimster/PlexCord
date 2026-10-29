package events

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	PlaybackUpdated        = "PlaybackUpdated"
	PlaybackStopped        = "PlaybackStopped"
	PlexConnectionError    = "PlexConnectionError"
	PlexConnectionRestored = "PlexConnectionRestored"
	PlexRetryState         = "PlexRetryState"
	DiscordConnected       = "DiscordConnected"
	DiscordDisconnected    = "DiscordDisconnected"
	DiscordRetryState      = "DiscordRetryState"

	UpdateAvailable        = "UpdateAvailable"
	UpdateDownloadProgress = "UpdateDownloadProgress"
	UpdateReady            = "UpdateReady"
	UpdateError            = "UpdateError"
)

type Bus interface {
	Emit(name string, payload ...interface{})
}

type WailsBus struct {
	ctx context.Context
}

func NewWailsBus(ctx context.Context) *WailsBus {
	return &WailsBus{ctx: ctx}
}

func (b *WailsBus) Emit(name string, payload ...interface{}) {
	if b == nil || b.ctx == nil {
		return
	}
	runtime.EventsEmit(b.ctx, name, payload...)
}

type RecordingBus struct {
	mu     sync.Mutex
	Events []RecordedEvent
}

type RecordedEvent struct {
	Name    string
	Payload []interface{}
}

func NewRecordingBus() *RecordingBus {
	return &RecordingBus{Events: []RecordedEvent{}}
}

func (b *RecordingBus) Emit(name string, payload ...interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Events = append(b.Events, RecordedEvent{Name: name, Payload: payload})
}

func (b *RecordingBus) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Events = b.Events[:0]
}

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

func (b *RecordingBus) Snapshot() []RecordedEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]RecordedEvent, len(b.Events))
	copy(out, b.Events)
	return out
}
