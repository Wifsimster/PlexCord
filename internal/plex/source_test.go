package plex

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// scriptedSource is a SessionSource backed by a script of responses, so the
// poller can be driven with no HTTP server and no wall-clock dependency on a
// real Plex instance.
type scriptedSource struct {
	mu       chan struct{} // used as a 1-slot mutex so calls serialize
	steps    [][]MusicSession
	errs     []error
	calls    atomic.Int64
	fallback []MusicSession
}

func newScriptedSource(steps [][]MusicSession, errs []error) *scriptedSource {
	s := &scriptedSource{mu: make(chan struct{}, 1), steps: steps, errs: errs}
	return s
}

func (s *scriptedSource) GetMusicSessions(string) ([]MusicSession, error) {
	i := int(s.calls.Add(1)) - 1
	if i < len(s.errs) && s.errs[i] != nil {
		return nil, s.errs[i]
	}
	if i < len(s.steps) {
		return s.steps[i], nil
	}
	return s.fallback, nil
}

func (s *scriptedSource) GetMediaSessions(string, []string) ([]MediaSession, error) {
	return nil, nil
}

func playing(track string) MusicSession {
	m := MusicSession{Track: track, Artist: "Artist", Album: "Album"}
	m.SessionKey = "key-" + track
	m.State = "playing"
	return m
}

// TestPollerDrivenByFakeSource is the point of the SessionSource seam: the
// poller's behaviour is verified without a Plex server or an httptest stand-in.
func TestPollerDrivenByFakeSource(t *testing.T) {
	source := newScriptedSource([][]MusicSession{{playing("Song")}}, nil)
	source.fallback = []MusicSession{playing("Song")}

	poller := NewPoller(source, "user1", time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := poller.Start(ctx)
	defer poller.Stop()

	select {
	case session := <-ch:
		if session == nil || session.Track != "Song" {
			t.Fatalf("first emission = %v, want the playing track", session)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the poller never emitted the first session")
	}
}

// TestPollerErrorAndRecoveryCallbacks verifies the error-state transitions
// both polling modes now share: onError fires once on the way in, onRecovered
// once on the way out.
func TestPollerErrorAndRecoveryCallbacks(t *testing.T) {
	source := newScriptedSource(nil, []error{errors.New("unreachable"), errors.New("unreachable")})
	source.fallback = []MusicSession{playing("Song")}

	var errorCalls, recoveredCalls atomic.Int64
	poller := NewPoller(source, "user1", time.Second)
	poller.SetErrorCallbacks(
		func(error) { errorCalls.Add(1) },
		func() { recoveredCalls.Add(1) },
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	poller.Start(ctx)
	defer poller.Stop()

	// Two consecutive failures, then success.
	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) && recoveredCalls.Load() == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	if got := errorCalls.Load(); got != 1 {
		t.Errorf("onError calls = %d, want exactly 1 — it must fire on the transition, not on every failing poll", got)
	}
	if got := recoveredCalls.Load(); got != 1 {
		t.Errorf("onRecovered calls = %d, want 1", got)
	}
	if poller.IsInErrorState() {
		t.Error("IsInErrorState() = true after recovery")
	}
}

// TestPollerReportsErrorState verifies a failing source is visible to the
// dashboard while it keeps failing.
func TestPollerReportsErrorState(t *testing.T) {
	poller := NewPoller(alwaysFailingSource{}, "user1", time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	poller.Start(ctx)
	defer poller.Stop()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && !poller.IsInErrorState() {
		time.Sleep(10 * time.Millisecond)
	}
	if !poller.IsInErrorState() {
		t.Error("IsInErrorState() = false while every poll is failing")
	}
}

type alwaysFailingSource struct{}

func (alwaysFailingSource) GetMusicSessions(string) ([]MusicSession, error) {
	return nil, errors.New("unreachable")
}

func (alwaysFailingSource) GetMediaSessions(string, []string) ([]MediaSession, error) {
	return nil, errors.New("unreachable")
}

// TestClientSatisfiesSessionSource guards the production wiring: the poller is
// constructed with a *Client at runtime, so a *Client that stopped satisfying
// SessionSource would break the app while every fake-driven test still passed.
func TestClientSatisfiesSessionSource(t *testing.T) {
	var source SessionSource = NewClient("token", "http://localhost:32400")

	// A poller built over the real client type must construct cleanly.
	if poller := NewPoller(source, "user1", time.Second); poller == nil {
		t.Fatal("NewPoller() over a *Client returned nil")
	}
}
