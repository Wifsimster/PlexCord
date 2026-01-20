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
type Poller struct {
	client   *Client
	userID   string
	interval time.Duration

	// Synchronization
	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}
	sessionC chan *MusicSession // nil indicates no session / stopped playback

	// Error handling (Story 6.5)
	onError       func(err error) // Called when poll errors occur
	onRecovered   func()          // Called when connection recovers after error
	lastErrorTime time.Time       // Track when last error occurred
	inErrorState  bool            // Whether currently in error state
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
		sessionC: make(chan *MusicSession, 1), // Buffered to prevent blocking
	}
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

// Stop gracefully stops the poller and cleans up resources.
// It is safe to call Stop multiple times.
func (p *Poller) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}

	p.running = false
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

// pollLoop is the main polling goroutine.
// It uses time.Ticker for efficient interval-based polling.
func (p *Poller) pollLoop(ctx context.Context) {
	// Ensure proper cleanup when goroutine exits (fixes running state sync)
	defer func() {
		p.mu.Lock()
		p.running = false
		// Close session channel to unblock any consumers (fixes goroutine leak)
		close(p.sessionC)
		p.mu.Unlock()
	}()

	// Create ticker for interval-based polling
	p.mu.RLock()
	interval := p.interval
	p.mu.RUnlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastSession *MusicSession

	// Perform immediate first poll (AC8: polling begins immediately)
	session := p.doPoll()
	if session != nil {
		select {
		case p.sessionC <- session:
			// Successfully sent initial session
		default:
			log.Printf("Session channel full, skipping initial update")
		}
		lastSession = session
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Poller stopped: context cancelled")
			return
		case <-p.stopCh:
			log.Printf("Poller stopped: stop signal received")
			return
		case <-ticker.C:
			// Check if interval changed and reset ticker if needed
			p.mu.RLock()
			newInterval := p.interval
			p.mu.RUnlock()

			if newInterval != interval {
				ticker.Reset(newInterval)
				interval = newInterval
				log.Printf("Poller interval changed to %v", interval)
			}

			session := p.doPoll()

			// Only emit if session state changed
			if sessionChanged(lastSession, session) {
				select {
				case p.sessionC <- session:
					// Successfully sent
				default:
					// Channel full, skip this update (prevents blocking)
					log.Printf("Session channel full, skipping update")
				}
				lastSession = session
			}
		}
	}
}

// doPoll performs a single poll for music sessions.
// Returns the current music session, or nil if no music is playing.
// Also handles error state transitions and callbacks (Story 6.5).
func (p *Poller) doPoll() *MusicSession {
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

		return nil
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
		return nil
	}

	// Return the first (most recent) music session
	return &sessions[0]
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
