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
	source        SessionSource
	stopCh        chan struct{}
	sessionC      chan *MusicSession // nil indicates no session / stopped playback (music mode)
	mediaC        chan *MediaSession // nil indicates no session / stopped playback (media mode)

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
//
// source is the abstraction the poller reads sessions from (see SessionSource);
// production callers pass a *Client, tests pass a fake.
func NewPoller(source SessionSource, userID string, interval time.Duration) *Poller {
	// Enforce interval bounds (AC3: min 1s, max 60s)
	if interval < time.Second {
		interval = time.Second
	}
	if interval > 60*time.Second {
		interval = 60 * time.Second
	}

	return &Poller{
		source:   source,
		userID:   userID,
		interval: interval,
		stopCh:   make(chan struct{}),
		sessionC: make(chan *MusicSession, 1), // Buffered to prevent blocking
		mediaC:   make(chan *MediaSession, 1), // Buffered to prevent blocking
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
//
// Channels are captured locally at goroutine start so subsequent Start/Stop
// cycles that replace the struct fields cannot cause this goroutine to send
// on (or close) the wrong channel.
func (p *Poller) pollLoop(ctx context.Context) {
	p.mu.Lock()
	sessionC := p.sessionC
	stopCh := p.stopCh
	p.mu.Unlock()

	// Ensure proper cleanup when goroutine exits
	defer func() {
		p.mu.Lock()
		// Only flip running off if this goroutine still owns the current
		// channel — otherwise a later Start has taken over.
		if p.sessionC == sessionC {
			p.running = false
		}
		p.mu.Unlock()
		close(sessionC)
	}()

	runPollLoop[*MusicSession](
		ctx,
		stopCh,
		p.GetInterval,
		p.doPoll,
		sessionChanged,
		func(session *MusicSession) {
			select {
			case sessionC <- session:
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
func (p *Poller) doPoll() (*MusicSession, bool) {
	return pollOnce(p, "Poll", func() ([]MusicSession, error) {
		return p.source.GetMusicSessions(p.userID)
	})
}

// pollOnce runs one fetch and folds the result into the poller's error state,
// returning the first session (or nil when nothing is playing) and whether the
// fetch succeeded. Both polling modes share it, so the error-state transitions
// and the recovery callback exist in exactly one place.
//
// It is a free function rather than a method because Go methods cannot take
// their own type parameters.
func pollOnce[T any](p *Poller, label string, fetch func() ([]T, error)) (*T, bool) {
	sessions, err := fetch()
	if err != nil {
		// Log and continue polling (AC4: failed polls do not stop the loop).
		log.Printf("%s error: %v", label, err)
		p.recordPollFailure(err)
		return nil, false
	}
	p.recordPollSuccess()

	if len(sessions) == 0 {
		return nil, true
	}
	// Return the first (most recent) session.
	return &sessions[0], true
}

// recordPollFailure marks the poller as being in an error state and fires the
// onError callback, but only on the transition into that state — not on every
// failing poll (Story 6.5).
func (p *Poller) recordPollFailure(err error) {
	p.mu.Lock()
	wasInErrorState := p.inErrorState
	p.inErrorState = true
	p.lastErrorTime = time.Now()
	onError := p.onError
	p.mu.Unlock()

	if !wasInErrorState && onError != nil {
		onError(err)
	}
}

// recordPollSuccess clears the error state and fires the onRecovered callback
// when this poll is the one that recovered the connection.
func (p *Poller) recordPollSuccess() {
	p.mu.Lock()
	wasInErrorState := p.inErrorState
	p.inErrorState = false
	onRecovered := p.onRecovered
	p.mu.Unlock()

	if wasInErrorState && onRecovered != nil {
		log.Printf("Plex connection recovered")
		onRecovered()
	}
}

// mediaPollLoop is the main polling goroutine for multi-media mode.
// It delegates the actual loop logic to runPollLoop (see poll_runner.go).
//
// See pollLoop for the channel-capture rationale.
func (p *Poller) mediaPollLoop(ctx context.Context) {
	p.mu.Lock()
	mediaC := p.mediaC
	stopCh := p.stopCh
	p.mu.Unlock()

	// Ensure proper cleanup when goroutine exits
	defer func() {
		p.mu.Lock()
		if p.mediaC == mediaC {
			p.running = false
			p.mediaMode = false
		}
		p.mu.Unlock()
		close(mediaC)
	}()

	runPollLoop[*MediaSession](
		ctx,
		stopCh,
		p.GetInterval,
		p.doMediaPoll,
		mediaSessionChanged,
		func(session *MediaSession) {
			select {
			case mediaC <- session:
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
	mediaTypes := p.GetMediaTypes()
	return pollOnce(p, "Media poll", func() ([]MediaSession, error) {
		return p.source.GetMediaSessions(p.userID, mediaTypes)
	})
}
