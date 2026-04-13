package plex

import (
	"context"
	"log"
	"sync"
	"time"
)

// Poller manages periodic polling of Plex sessions for a specific user.
// It uses time.Ticker for accurate interval-based polling (not busy-waiting)
// to meet CPU efficiency requirements (NFR3: <1% average CPU).
//
// The poller supports two modes:
//   - Music-only mode (default): Uses Start() and emits *MusicSession on the channel.
//   - Multi-media mode: Uses StartMedia() and emits *MediaSession on the media channel.
//     Enabled by setting MediaTypes before calling StartMedia().
type Poller struct {
	lastErrorTime time.Time // Track when last error occurred
	client        *Client
	stopCh        chan struct{}
	sessionC      chan *MusicSession  // nil indicates no session / stopped playback (music mode)
	mediaC        chan *MediaSession  // nil indicates no session / stopped playback (media mode)

	// Error handling (Story 6.5)
	onError     func(err error) // Called when poll errors occur
	onRecovered func()          // Called when connection recovers after error

	userID     string
	interval   time.Duration
	mediaTypes []string // Media types to poll for (e.g., ["music", "movie", "tv"]). Empty = music only.

	// Synchronization
	mu           sync.RWMutex
	running      bool
	inErrorState bool // Whether currently in error state
	mediaMode    bool // Whether polling in multi-media mode
}

// NewPoller creates a new session poller for the specified user.
// The interval parameter controls how frequently sessions are polled.
// Minimum interval is 1 second, maximum is 60 seconds.
func NewPoller(client *Client, userID string, interval time.Duration) *Poller {
	// Enforce interval bounds (AC3: min 1s, max 60s)
	if interval < time.Second {
		interval = time.Second
	}
	if interval > 60*time.Second {
		interval = 60 * time.Second
	}

	return &Poller{
		client:   client,
		userID:   userID,
		interval: interval,
		stopCh:   make(chan struct{}),
		sessionC: make(chan *MusicSession, 1),  // Buffered to prevent blocking
		mediaC:   make(chan *MediaSession, 1),   // Buffered to prevent blocking
	}
}

// SetMediaTypes sets the media types that the poller should monitor.
// Valid types: "music", "movie", "tv", "photo".
// An empty or nil slice defaults to music-only polling (backward compatible).
// Must be called before StartMedia().
func (p *Poller) SetMediaTypes(types []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mediaTypes = types
}

// GetMediaTypes returns the currently configured media types.
func (p *Poller) GetMediaTypes() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.mediaTypes
}

// Start begins polling for music sessions.
// Returns a channel that receives MusicSession updates when the session state changes.
// A nil value on the channel indicates no music session is active (playback stopped).
// The poller performs an immediate first poll, then continues at the configured interval.
// The channel is closed when the poller stops - consumers should handle this gracefully.
func (p *Poller) Start(ctx context.Context) <-chan *MusicSession {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return p.sessionC
	}
	p.running = true
	p.stopCh = make(chan struct{})
	// Create new session channel for this run (previous one was closed on stop)
	p.sessionC = make(chan *MusicSession, 1)
	p.mu.Unlock()

	go p.pollLoop(ctx)

	return p.sessionC
}

// StartMedia begins polling for media sessions of the configured types.
// Returns a channel that receives MediaSession updates when the session state changes.
// A nil value on the channel indicates no matching session is active (playback stopped).
// Use SetMediaTypes() before calling this to configure which media types to poll.
// The channel is closed when the poller stops - consumers should handle this gracefully.
func (p *Poller) StartMedia(ctx context.Context) <-chan *MediaSession {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return p.mediaC
	}
	p.running = true
	p.mediaMode = true
	p.stopCh = make(chan struct{})
	// Create new media channel for this run (previous one was closed on stop)
	p.mediaC = make(chan *MediaSession, 1)
	p.mu.Unlock()

	go p.mediaPollLoop(ctx)

	return p.mediaC
}

// Stop gracefully stops the poller and cleans up resources.
// It is safe to call Stop multiple times.
func (p *Poller) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}

	p.running = false
	p.mediaMode = false
	close(p.stopCh)
}

// SetInterval updates the polling interval dynamically.
// Changes take effect on the next polling cycle.
// Interval is clamped to min 1s, max 60s.
func (p *Poller) SetInterval(interval time.Duration) {
	// Enforce interval bounds
	if interval < time.Second {
		interval = time.Second
	}
	if interval > 60*time.Second {
		interval = 60 * time.Second
	}

	p.mu.Lock()
	p.interval = interval
	p.mu.Unlock()
}

// GetInterval returns the current polling interval.
func (p *Poller) GetInterval() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.interval
}

// IsRunning returns whether the poller is currently running.
func (p *Poller) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// SetErrorCallbacks sets callbacks for error handling (Story 6.5).
// onError is called when a poll fails (for starting retry, clearing presence).
// onRecovered is called when connection recovers after errors.
func (p *Poller) SetErrorCallbacks(onError func(err error), onRecovered func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onError = onError
	p.onRecovered = onRecovered
}

// IsInErrorState returns whether the poller is currently in an error state.
func (p *Poller) IsInErrorState() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.inErrorState
}

// pollLoop is the main polling goroutine for music-only mode.
// It delegates the actual loop logic to runPollLoop (see poll_runner.go).
func (p *Poller) pollLoop(ctx context.Context) {
	// Ensure proper cleanup when goroutine exits
	defer func() {
		p.mu.Lock()
		p.running = false
		close(p.sessionC)
		p.mu.Unlock()
	}()

	runPollLoop[*MusicSession](
		ctx,
		p.stopCh,
		p.GetInterval,
		p.doPoll,
		sessionChanged,
		func(session *MusicSession) {
			select {
			case p.sessionC <- session:
			default:
				log.Printf("Session channel full, skipping update")
			}
		},
		"Music",
	)
}

// doPoll performs a single poll for music sessions.
// Returns the current music session, or nil if no music is playing.
// The second return value indicates whether the result is valid (not an error).
// Also handles error state transitions and callbacks (Story 6.5).
func (p *Poller) doPoll() (*MusicSession, bool) {
	sessions, err := p.client.GetMusicSessions(p.userID)
	if err != nil {
		// Log error but continue polling (AC4: failed polls continue polling)
		log.Printf("Poll error: %v", err)

		// Handle error state transition (Story 6.5)
		p.mu.Lock()
		wasInErrorState := p.inErrorState
		p.inErrorState = true
		p.lastErrorTime = time.Now()
		onError := p.onError
		p.mu.Unlock()

		// Call error callback only on first error (not every poll)
		if !wasInErrorState && onError != nil {
			onError(err)
		}

		return nil, false
	}

	// Connection successful - check if recovering from error state
	p.mu.Lock()
	wasInErrorState := p.inErrorState
	p.inErrorState = false
	onRecovered := p.onRecovered
	p.mu.Unlock()

	if wasInErrorState && onRecovered != nil {
		log.Printf("Plex connection recovered")
		onRecovered()
	}

	if len(sessions) == 0 {
		return nil, true
	}

	// Return the first (most recent) music session
	return &sessions[0], true
}

// mediaPollLoop is the main polling goroutine for multi-media mode.
// It delegates the actual loop logic to runPollLoop (see poll_runner.go).
func (p *Poller) mediaPollLoop(ctx context.Context) {
	// Ensure proper cleanup when goroutine exits
	defer func() {
		p.mu.Lock()
		p.running = false
		p.mediaMode = false
		close(p.mediaC)
		p.mu.Unlock()
	}()

	runPollLoop[*MediaSession](
		ctx,
		p.stopCh,
		p.GetInterval,
		p.doMediaPoll,
		mediaSessionChanged,
		func(session *MediaSession) {
			select {
			case p.mediaC <- session:
			default:
				log.Printf("Media session channel full, skipping update")
			}
		},
		"Media",
	)
}

// doMediaPoll performs a single poll for media sessions.
// Returns the current media session, or nil if no matching media is playing.
// The second return value indicates whether the result is valid (not an error).
func (p *Poller) doMediaPoll() (*MediaSession, bool) {
	p.mu.RLock()
	mediaTypes := p.mediaTypes
	p.mu.RUnlock()

	sessions, err := p.client.GetMediaSessions(p.userID, mediaTypes)
	if err != nil {
		log.Printf("Media poll error: %v", err)

		// Handle error state transition (Story 6.5)
		p.mu.Lock()
		wasInErrorState := p.inErrorState
		p.inErrorState = true
		p.lastErrorTime = time.Now()
		onError := p.onError
		p.mu.Unlock()

		if !wasInErrorState && onError != nil {
			onError(err)
		}

		return nil, false
	}

	// Connection successful - check if recovering from error state
	p.mu.Lock()
	wasInErrorState := p.inErrorState
	p.inErrorState = false
	onRecovered := p.onRecovered
	p.mu.Unlock()

	if wasInErrorState && onRecovered != nil {
		log.Printf("Plex connection recovered")
		onRecovered()
	}

	if len(sessions) == 0 {
		return nil, true
	}

	// Return the first (most recent) session
	return &sessions[0], true
}

// mediaSessionChanged determines if the media session state has meaningfully changed.
// Used to avoid emitting duplicate updates in multi-media mode.
func mediaSessionChanged(prev, curr *MediaSession) bool {
	// Both nil - no change
	if prev == nil && curr == nil {
		return false
	}

	// One nil, one not - definite change
	if prev == nil || curr == nil {
		return true
	}

	// Compare key session attributes
	if prev.SessionKey != curr.SessionKey {
		return true
	}

	if prev.State != curr.State {
		return true
	}

	if prev.Title != curr.Title {
		return true
	}

	if prev.MediaType != curr.MediaType {
		return true
	}

	// Music-specific changes
	if prev.Artist != curr.Artist {
		return true
	}

	if prev.Album != curr.Album {
		return true
	}

	// TV-specific changes
	if prev.ShowTitle != curr.ShowTitle {
		return true
	}

	if prev.Season != curr.Season {
		return true
	}

	if prev.Episode != curr.Episode {
		return true
	}

	return false
}

// sessionChanged determines if the session state has meaningfully changed.
// Used to avoid emitting duplicate updates.
func sessionChanged(prev, curr *MusicSession) bool {
	// Both nil - no change
	if prev == nil && curr == nil {
		return false
	}

	// One nil, one not - definite change
	if prev == nil || curr == nil {
		return true
	}

	// Compare key session attributes
	if prev.SessionKey != curr.SessionKey {
		return true
	}

	if prev.State != curr.State {
		return true
	}

	if prev.Track != curr.Track {
		return true
	}

	// Also detect metadata changes (e.g., if Plex refreshes metadata during playback)
	if prev.Artist != curr.Artist {
		return true
	}

	if prev.Album != curr.Album {
		return true
	}

	// ViewOffset changes are expected during playback, don't emit for every update
	// Only emit if track, state, or metadata changed

	return false
}
