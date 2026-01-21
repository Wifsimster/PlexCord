package plex

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewPollerDefaults tests default poller creation
func TestNewPollerDefaults(t *testing.T) {
	client := NewClient("token", "http://localhost:32400")
	poller := NewPoller(client, "user1", 5*time.Second)

	if poller == nil {
		t.Fatal("NewPoller returned nil")
	}

	if poller.client != client {
		t.Error("Client not set correctly")
	}

	if poller.userID != "user1" {
		t.Errorf("Expected userID 'user1', got '%s'", poller.userID)
	}

	if poller.interval != 5*time.Second {
		t.Errorf("Expected interval 5s, got %v", poller.interval)
	}

	if poller.running {
		t.Error("Poller should not be running initially")
	}
}

// TestNewPollerIntervalBounds tests interval clamping (AC3)
func TestNewPollerIntervalBounds(t *testing.T) {
	client := NewClient("token", "http://localhost:32400")

	// Test minimum bound (< 1s should become 1s)
	pollerMin := NewPoller(client, "user1", 100*time.Millisecond)
	if pollerMin.interval != time.Second {
		t.Errorf("Expected minimum interval 1s, got %v", pollerMin.interval)
	}

	// Test maximum bound (> 60s should become 60s)
	pollerMax := NewPoller(client, "user1", 120*time.Second)
	if pollerMax.interval != 60*time.Second {
		t.Errorf("Expected maximum interval 60s, got %v", pollerMax.interval)
	}

	// Test valid interval passes through
	pollerValid := NewPoller(client, "user1", 10*time.Second)
	if pollerValid.interval != 10*time.Second {
		t.Errorf("Expected interval 10s, got %v", pollerValid.interval)
	}
}

// TestPollerStartStop tests basic start/stop lifecycle (AC8)
func TestPollerStartStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", time.Second)

	// Initially not running
	if poller.IsRunning() {
		t.Error("Poller should not be running before Start")
	}

	// Start poller
	ctx := context.Background()
	sessionCh := poller.Start(ctx)

	if sessionCh == nil {
		t.Error("Start should return a channel")
	}

	// Should be running now
	time.Sleep(50 * time.Millisecond) // Give goroutine time to start
	if !poller.IsRunning() {
		t.Error("Poller should be running after Start")
	}

	// Stop poller
	poller.Stop()

	// Give goroutine time to stop
	time.Sleep(50 * time.Millisecond)
	if poller.IsRunning() {
		t.Error("Poller should not be running after Stop")
	}
}

// TestPollerStopMultipleTimes tests that Stop can be called multiple times safely
func TestPollerStopMultipleTimes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", time.Second)

	ctx := context.Background()
	poller.Start(ctx)

	// Stop multiple times - should not panic
	poller.Stop()
	poller.Stop()
	poller.Stop()
}

// TestPollerEmitsSessionOnStart tests immediate first poll (AC8)
func TestPollerEmitsSessionOnStart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="test123" type="track" title="Test Song"
         grandparentTitle="Test Artist" parentTitle="Test Album">
    <User id="user1" title="TestUser"/>
    <Player state="playing" title="TestPlayer"/>
  </Track>
</MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", 5*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sessionCh := poller.Start(ctx)
	defer poller.Stop()

	// Should receive session from immediate first poll
	select {
	case session := <-sessionCh:
		if session == nil {
			t.Error("Expected session, got nil")
		} else {
			if session.Track != "Test Song" {
				t.Errorf("Expected track 'Test Song', got '%s'", session.Track)
			}
			if session.State != "playing" {
				t.Errorf("Expected state 'playing', got '%s'", session.State)
			}
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for initial session")
	}
}

// TestPollerPollsAtInterval tests that polling occurs at configured interval
func TestPollerPollsAtInterval(t *testing.T) {
	var pollCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&pollCount, 1)
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	// Use short interval for testing
	poller := NewPoller(client, "user1", time.Second)

	ctx := context.Background()
	poller.Start(ctx)

	// Wait for multiple poll cycles
	time.Sleep(2500 * time.Millisecond)
	poller.Stop()

	// Should have polled at least 2-3 times (initial + interval polls)
	count := atomic.LoadInt32(&pollCount)
	if count < 2 {
		t.Errorf("Expected at least 2 polls, got %d", count)
	}
}

// TestPollerSetInterval tests dynamic interval changes (AC3)
func TestPollerSetInterval(t *testing.T) {
	client := NewClient("token", "http://localhost:32400")
	poller := NewPoller(client, "user1", 5*time.Second)

	// Set new interval
	poller.SetInterval(10 * time.Second)
	if poller.GetInterval() != 10*time.Second {
		t.Errorf("Expected interval 10s, got %v", poller.GetInterval())
	}

	// Test clamping on SetInterval
	poller.SetInterval(500 * time.Millisecond)
	if poller.GetInterval() != time.Second {
		t.Errorf("Expected minimum interval 1s, got %v", poller.GetInterval())
	}

	poller.SetInterval(90 * time.Second)
	if poller.GetInterval() != 60*time.Second {
		t.Errorf("Expected maximum interval 60s, got %v", poller.GetInterval())
	}
}

// TestPollerContinuesOnError tests error handling (AC4)
func TestPollerContinuesOnError(t *testing.T) {
	var pollCount int32
	var errorCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&pollCount, 1)
		if count%2 == 0 {
			// Every other request fails
			atomic.AddInt32(&errorCount, 1)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", time.Second)

	ctx := context.Background()
	poller.Start(ctx)

	// Wait for multiple polls including errors
	time.Sleep(3500 * time.Millisecond)
	poller.Stop()

	// Should have continued polling despite errors
	totalPolls := atomic.LoadInt32(&pollCount)
	errors := atomic.LoadInt32(&errorCount)

	if totalPolls < 3 {
		t.Errorf("Expected at least 3 poll attempts, got %d", totalPolls)
	}

	if errors == 0 {
		t.Error("Expected some errors to occur during polling")
	}
}

// TestPollerContextCancellation tests that poller stops on context cancel
func TestPollerContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", 5*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	poller.Start(ctx)

	// Verify running
	time.Sleep(50 * time.Millisecond)
	if !poller.IsRunning() {
		t.Error("Poller should be running")
	}

	// Cancel context
	cancel()

	// Give time for goroutine to stop
	time.Sleep(100 * time.Millisecond)

	// Poller should now report not running (fixed: running state sync)
	if poller.IsRunning() {
		t.Error("Poller should not be running after context cancellation")
	}
}

// TestPollerChannelClosedOnStop tests that session channel is closed when poller stops
func TestPollerChannelClosedOnStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", 5*time.Second)

	ctx := context.Background()
	sessionCh := poller.Start(ctx)

	// Verify running
	time.Sleep(50 * time.Millisecond)

	// Stop poller
	poller.Stop()

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	// Channel should be closed - reading should return immediately with zero value
	select {
	case <-sessionCh:
		// Channel closed or got a value - both are acceptable
		// Channel closed is expected
	case <-time.After(500 * time.Millisecond):
		t.Error("Channel should be closed after stop, but read blocked")
	}
}

// TestPollerRestart tests that poller can be restarted after stopping (AC8)
func TestPollerRestart(t *testing.T) {
	var pollCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&pollCount, 1)
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", time.Second)

	// First start/stop cycle
	ctx1 := context.Background()
	poller.Start(ctx1)
	time.Sleep(100 * time.Millisecond)
	poller.Stop()

	firstCount := atomic.LoadInt32(&pollCount)
	if firstCount < 1 {
		t.Error("Expected at least 1 poll in first cycle")
	}

	// Second start/stop cycle
	time.Sleep(50 * time.Millisecond) // Brief pause between cycles

	ctx2 := context.Background()
	poller.Start(ctx2)
	time.Sleep(100 * time.Millisecond)
	poller.Stop()

	secondCount := atomic.LoadInt32(&pollCount)
	if secondCount <= firstCount {
		t.Errorf("Expected more polls after restart, first: %d, total: %d", firstCount, secondCount)
	}
}

// TestPollerFiltersUser tests that only the specified user's sessions are returned
func TestPollerFiltersUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="2">
  <Track sessionKey="session1" type="track" title="Song1">
    <User id="user1" title="User1"/>
    <Player state="playing" title="Player1"/>
  </Track>
  <Track sessionKey="session2" type="track" title="Song2">
    <User id="user2" title="User2"/>
    <Player state="playing" title="Player2"/>
  </Track>
</MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	// Poll for user2 only
	poller := NewPoller(client, "user2", 5*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sessionCh := poller.Start(ctx)
	defer poller.Stop()

	select {
	case session := <-sessionCh:
		if session == nil {
			t.Error("Expected session for user2")
		} else {
			if session.UserID != "user2" {
				t.Errorf("Expected UserID 'user2', got '%s'", session.UserID)
			}
			if session.Track != "Song2" {
				t.Errorf("Expected track 'Song2', got '%s'", session.Track)
			}
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for session")
	}
}

// TestSessionChangedLogic tests the sessionChanged helper function
func TestSessionChangedLogic(t *testing.T) {
	session1 := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Song1",
		Artist:  "Artist1",
		Album:   "Album1",
	}
	session1Same := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Song1",
		Artist:  "Artist1",
		Album:   "Album1",
	}
	session1Paused := &MusicSession{
		Session: Session{SessionKey: "key1", State: "paused"},
		Track:   "Song1",
		Artist:  "Artist1",
		Album:   "Album1",
	}
	session1DiffArtist := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Song1",
		Artist:  "Artist2", // Different artist (metadata refresh)
		Album:   "Album1",
	}
	session1DiffAlbum := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Song1",
		Artist:  "Artist1",
		Album:   "Album2", // Different album (metadata refresh)
	}
	session2 := &MusicSession{
		Session: Session{SessionKey: "key2", State: "playing"},
		Track:   "Song2",
		Artist:  "Artist2",
		Album:   "Album2",
	}

	testCases := []struct {
		prev     *MusicSession
		curr     *MusicSession
		name     string
		expected bool
	}{
		{nil, nil, "Both nil", false},
		{nil, session1, "Prev nil, curr not nil", true},
		{session1, nil, "Prev not nil, curr nil", true},
		{session1, session1Same, "Same session", false},
		{session1, session1Paused, "Different state", true},
		{session1, session2, "Different session", true},
		{session1, session1DiffArtist, "Different artist (metadata change)", true},
		{session1, session1DiffAlbum, "Different album (metadata change)", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sessionChanged(tc.prev, tc.curr)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestSessionChangedPlayToPause tests play→pause state transition detection (AC2)
func TestSessionChangedPlayToPause(t *testing.T) {
	playing := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Song1",
		Artist:  "Artist1",
		Album:   "Album1",
	}
	paused := &MusicSession{
		Session: Session{SessionKey: "key1", State: "paused"},
		Track:   "Song1",
		Artist:  "Artist1",
		Album:   "Album1",
	}

	// Play → Pause should be detected
	if !sessionChanged(playing, paused) {
		t.Error("Expected play→pause transition to be detected")
	}

	// Verify the state values are correct
	if playing.State != "playing" {
		t.Errorf("Expected playing state 'playing', got '%s'", playing.State)
	}
	if paused.State != "paused" {
		t.Errorf("Expected paused state 'paused', got '%s'", paused.State)
	}
}

// TestSessionChangedPauseToPlay tests pause→play (resume) transition detection (AC7)
func TestSessionChangedPauseToPlay(t *testing.T) {
	paused := &MusicSession{
		Session: Session{SessionKey: "key1", State: "paused"},
		Track:   "Song1",
		Artist:  "Artist1",
		Album:   "Album1",
	}
	playing := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Song1",
		Artist:  "Artist1",
		Album:   "Album1",
	}

	// Pause → Play (resume) should be detected
	if !sessionChanged(paused, playing) {
		t.Error("Expected pause→play (resume) transition to be detected")
	}
}

// TestSessionChangedPlayToStopped tests play→stopped (nil session) detection (AC3)
func TestSessionChangedPlayToStopped(t *testing.T) {
	playing := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Song1",
		Artist:  "Artist1",
		Album:   "Album1",
	}

	// Playing → nil (stopped) should be detected
	if !sessionChanged(playing, nil) {
		t.Error("Expected play→stopped (nil) transition to be detected")
	}

	// nil → nil should NOT be detected as a change
	if sessionChanged(nil, nil) {
		t.Error("Expected nil→nil to NOT be detected as a change")
	}
}

// TestSessionChangedSameSessionDifferentState tests state changes within same session key
func TestSessionChangedSameSessionDifferentState(t *testing.T) {
	sessionKey := "same-session-123"

	states := []string{"playing", "paused", "playing", "paused"}
	var prev *MusicSession

	for i, state := range states {
		curr := &MusicSession{
			Session: Session{SessionKey: sessionKey, State: state},
			Track:   "Song1",
			Artist:  "Artist1",
			Album:   "Album1",
		}

		if i > 0 && states[i] != states[i-1] {
			// State changed, should be detected
			if !sessionChanged(prev, curr) {
				t.Errorf("State change from '%s' to '%s' should be detected", prev.State, curr.State)
			}
		}

		prev = curr
	}
}

// TestSessionChangedTrackTitleChange tests track title change detection (AC4)
func TestSessionChangedTrackTitleChange(t *testing.T) {
	song1 := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "First Song",
		Artist:  "Artist1",
		Album:   "Album1",
	}
	song2 := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Second Song", // Different track title
		Artist:  "Artist1",
		Album:   "Album1",
	}

	// Track title change should be detected
	if !sessionChanged(song1, song2) {
		t.Error("Expected track title change to be detected")
	}
}

// TestSessionChangedArtistChange tests artist change detection (AC4)
func TestSessionChangedArtistChange(t *testing.T) {
	song1 := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Same Song",
		Artist:  "First Artist",
		Album:   "Album1",
	}
	song2 := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Same Song",
		Artist:  "Second Artist", // Different artist
		Album:   "Album1",
	}

	// Artist change should be detected
	if !sessionChanged(song1, song2) {
		t.Error("Expected artist change to be detected")
	}
}

// TestSessionChangedAlbumChange tests album change detection (AC4)
func TestSessionChangedAlbumChange(t *testing.T) {
	song1 := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Same Song",
		Artist:  "Artist1",
		Album:   "First Album",
	}
	song2 := &MusicSession{
		Session: Session{SessionKey: "key1", State: "playing"},
		Track:   "Same Song",
		Artist:  "Artist1",
		Album:   "Second Album", // Different album
	}

	// Album change should be detected
	if !sessionChanged(song1, song2) {
		t.Error("Expected album change to be detected")
	}
}

// TestSessionChangedViewOffsetOnly tests that viewOffset change alone does NOT trigger change
// This is expected behavior - we don't want to emit events for every second of playback
func TestSessionChangedViewOffsetOnly(t *testing.T) {
	session1 := &MusicSession{
		Session:    Session{SessionKey: "key1", State: "playing"},
		Track:      "Song1",
		Artist:     "Artist1",
		Album:      "Album1",
		ViewOffset: 10000, // 10 seconds
	}
	session2 := &MusicSession{
		Session:    Session{SessionKey: "key1", State: "playing"},
		Track:      "Song1",
		Artist:     "Artist1",
		Album:      "Album1",
		ViewOffset: 15000, // 15 seconds - only viewOffset changed
	}

	// ViewOffset-only change should NOT be detected (to avoid excessive events)
	if sessionChanged(session1, session2) {
		t.Error("Expected viewOffset-only change to NOT be detected (would cause excessive events)")
	}
}

// TestSessionChangedDurationOnly tests that duration change alone does NOT trigger change
func TestSessionChangedDurationOnly(t *testing.T) {
	session1 := &MusicSession{
		Session:  Session{SessionKey: "key1", State: "playing"},
		Track:    "Song1",
		Artist:   "Artist1",
		Album:    "Album1",
		Duration: 180000, // 3 minutes
	}
	session2 := &MusicSession{
		Session:  Session{SessionKey: "key1", State: "playing"},
		Track:    "Song1",
		Artist:   "Artist1",
		Album:    "Album1",
		Duration: 185000, // 3:05 - only duration changed (rare edge case)
	}

	// Duration-only change should NOT be detected
	if sessionChanged(session1, session2) {
		t.Error("Expected duration-only change to NOT be detected")
	}
}

// TestStateFieldParsedFromPlexAPI tests that state field is correctly parsed from Plex response
func TestStateFieldParsedFromPlexAPI(t *testing.T) {
	testCases := []struct {
		name          string
		xml           string
		expectedState string
	}{
		{
			name: "Playing state",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="test1" type="track" title="Song">
    <User id="user1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
</MediaContainer>`,
			expectedState: "playing",
		},
		{
			name: "Paused state",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="test2" type="track" title="Song">
    <User id="user1" title="User"/>
    <Player state="paused" title="Player"/>
  </Track>
</MediaContainer>`,
			expectedState: "paused",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tc.xml))
			}))
			defer server.Close()

			client := NewClient("token", server.URL)
			sessions, err := client.GetMusicSessions("user1")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(sessions) != 1 {
				t.Fatalf("Expected 1 session, got %d", len(sessions))
			}

			if sessions[0].State != tc.expectedState {
				t.Errorf("Expected state '%s', got '%s'", tc.expectedState, sessions[0].State)
			}
		})
	}
}

// TestPollerEmitsOnStateChange tests that poller emits session when state changes (AC1, AC2)
// This verifies the channel behavior that triggers PlaybackUpdated events
func TestPollerEmitsOnStateChange(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)

		// First request: playing state
		// Second request: paused state (state change should trigger emission)
		state := "playing"
		if requestCount >= 2 {
			state = "paused"
		}

		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="test" type="track" title="Song">
    <User id="user1" title="User"/>
    <Player state="` + state + `" title="Player"/>
  </Track>
</MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	sessionCh := poller.Start(ctx)
	defer poller.Stop()

	// First emission: playing state
	select {
	case session := <-sessionCh:
		if session == nil {
			t.Error("Expected playing session, got nil")
		} else if session.State != "playing" {
			t.Errorf("Expected state 'playing', got '%s'", session.State)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for playing session")
	}

	// Second emission: paused state (after interval)
	select {
	case session := <-sessionCh:
		if session == nil {
			t.Error("Expected paused session, got nil")
		} else if session.State != "paused" {
			t.Errorf("Expected state 'paused', got '%s'", session.State)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for paused session (state change)")
	}
}

// TestPollerEmitsOnTrackChange tests that poller emits session when track changes (AC4)
// This verifies the channel behavior that triggers PlaybackUpdated events on track change
func TestPollerEmitsOnTrackChange(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)

		// First request: Song1
		// Second request: Song2 (track change should trigger emission)
		track := "Song1"
		if requestCount >= 2 {
			track = "Song2"
		}

		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="test" type="track" title="` + track + `">
    <User id="user1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
</MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	sessionCh := poller.Start(ctx)
	defer poller.Stop()

	// First emission: Song1
	select {
	case session := <-sessionCh:
		if session == nil || session.Track != "Song1" {
			t.Errorf("Expected Song1, got %v", session)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for first track")
	}

	// Second emission: Song2 (track change)
	select {
	case session := <-sessionCh:
		if session == nil {
			t.Error("Expected Song2 session, got nil")
		} else if session.Track != "Song2" {
			t.Errorf("Expected track 'Song2', got '%s'", session.Track)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for track change")
	}
}

// TestPollerEmitsNilOnSessionEnd tests that poller emits nil when session ends (AC3)
// This verifies the channel behavior that triggers PlaybackStopped events
func TestPollerEmitsNilOnSessionEnd(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)

		// First request: active session
		// Second request: no session (ended)
		if requestCount == 1 {
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="test" type="track" title="Song">
    <User id="user1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
</MediaContainer>`))
		} else {
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
		}
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	sessionCh := poller.Start(ctx)
	defer poller.Stop()

	// First emission: active session
	select {
	case session := <-sessionCh:
		if session == nil {
			t.Error("Expected active session, got nil")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for active session")
	}

	// Second emission: nil (session ended - triggers PlaybackStopped)
	select {
	case session := <-sessionCh:
		if session != nil {
			t.Errorf("Expected nil session (ended), got %v", session)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for session end (nil)")
	}
}

// TestPollerNilOnNoSession tests that nil is emitted when no session exists
func TestPollerNilOnNoSession(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)

		// First request returns a session, second returns empty
		if requestCount == 1 {
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="test" type="track" title="Song">
    <User id="user1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
</MediaContainer>`))
		} else {
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
		}
	}))
	defer server.Close()

	client := NewClient("token", server.URL)
	poller := NewPoller(client, "user1", time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sessionCh := poller.Start(ctx)
	defer poller.Stop()

	// First should be a session
	select {
	case session := <-sessionCh:
		if session == nil {
			t.Error("Expected first session to not be nil")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for first session")
	}

	// Wait for second poll cycle which should return nil (no session)
	select {
	case session := <-sessionCh:
		if session != nil {
			t.Error("Expected nil session when playback stopped")
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for nil session")
	}
}
