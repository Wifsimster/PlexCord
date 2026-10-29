package plex

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type scriptedSource struct {
	mu       chan struct{}
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

func TestClientSatisfiesSessionSource(t *testing.T) {
	var source SessionSource = NewClient("token", "http://localhost:32400")

	if poller := NewPoller(source, "user1", time.Second); poller == nil {
		t.Fatal("NewPoller() over a *Client returned nil")
	}
}
