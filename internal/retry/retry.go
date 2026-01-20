// Package retry provides exponential backoff retry functionality.
// This implements NFR18: automatic retry with 5s → 10s → 30s → 60s max backoff.
package retry

import (
	"context"
	"log"
	"sync"
	"time"
)

// BackoffSchedule defines the retry intervals (NFR18).
var BackoffSchedule = []time.Duration{
	5 * time.Second,
	10 * time.Second,
	30 * time.Second,
	60 * time.Second, // Max interval - continues at this rate
}

// RetryState represents the current state of retry attempts.
type RetryState struct {
	AttemptNumber      int           `json:"attemptNumber"`
	NextRetryIn        time.Duration `json:"nextRetryIn"`
	NextRetryAt        time.Time     `json:"nextRetryAt"`
	LastError          string        `json:"lastError,omitempty"`
	LastErrorCode      string        `json:"lastErrorCode,omitempty"`
	IsRetrying         bool          `json:"isRetrying"`
	MaxIntervalReached bool          `json:"maxIntervalReached"`
}

// RetryCallback is called when a retry should be attempted.
// Returns nil on success, or error to continue retrying.
type RetryCallback func() error

// StateChangeCallback is called when the retry state changes.
type StateChangeCallback func(state RetryState)

// Manager handles automatic retry with exponential backoff.
type Manager struct {
	mu            sync.Mutex
	name          string
	attemptNumber int
	lastError     error
	lastErrorCode string
	timer         *time.Timer
	ctx           context.Context
	cancel        context.CancelFunc
	running       bool
	retryCallback RetryCallback
	stateCallback StateChangeCallback
}

// NewManager creates a new retry manager.
// name is used for logging identification.
func NewManager(name string) *Manager {
	return &Manager{
		name: name,
	}
}

// SetCallbacks configures the retry and state change callbacks.
func (m *Manager) SetCallbacks(retry RetryCallback, stateChange StateChangeCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retryCallback = retry
	m.stateCallback = stateChange
}

// Start begins the retry cycle after a failure.
// err is the error that triggered the retry, code is the error code (optional).
func (m *Manager) Start(err error, code string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If already retrying, just update the error
	if m.running {
		m.lastError = err
		m.lastErrorCode = code
		return
	}

	m.lastError = err
	m.lastErrorCode = code
	m.attemptNumber = 0
	m.running = true

	// Create cancellation context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Schedule first retry
	m.scheduleNextRetry()
}

// Stop cancels any pending retries.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopInternal()
}

// stopInternal stops the retry cycle (must be called with lock held).
func (m *Manager) stopInternal() {
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.running = false
	m.attemptNumber = 0
}

// Reset stops retrying and clears state (call on success).
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopInternal()
	m.lastError = nil
	m.lastErrorCode = ""

	// Emit cleared state
	if m.stateCallback != nil {
		m.stateCallback(RetryState{
			AttemptNumber: 0,
			IsRetrying:    false,
		})
	}
}

// ManualRetry triggers an immediate retry and resets the backoff schedule.
func (m *Manager) ManualRetry() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure retry context exists so manual retry works even if not already running
	if m.ctx == nil || m.ctx.Err() != nil {
		m.ctx, m.cancel = context.WithCancel(context.Background())
		m.running = true
	}

	// Stop any pending timer
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}

	// Reset attempt number for fresh backoff schedule
	m.attemptNumber = 0

	// Attempt retry immediately
	go m.doRetry()
}

// GetState returns the current retry state.
func (m *Manager) GetState() RetryState {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := RetryState{
		AttemptNumber:      m.attemptNumber,
		IsRetrying:         m.running,
		MaxIntervalReached: m.attemptNumber >= len(BackoffSchedule)-1,
	}

	if m.lastError != nil {
		state.LastError = m.lastError.Error()
		state.LastErrorCode = m.lastErrorCode
	}

	if m.running && m.timer != nil {
		interval := m.getInterval(m.attemptNumber)
		state.NextRetryIn = interval
		state.NextRetryAt = time.Now().Add(interval)
	}

	return state
}

// IsRetrying returns whether a retry cycle is active.
func (m *Manager) IsRetrying() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// getInterval returns the backoff interval for the given attempt number.
func (m *Manager) getInterval(attempt int) time.Duration {
	if attempt >= len(BackoffSchedule) {
		return BackoffSchedule[len(BackoffSchedule)-1] // Max interval
	}
	return BackoffSchedule[attempt]
}

// scheduleNextRetry schedules the next retry attempt (must be called with lock held).
func (m *Manager) scheduleNextRetry() {
	interval := m.getInterval(m.attemptNumber)

	log.Printf("[%s] Retry scheduled in %v (attempt %d)", m.name, interval, m.attemptNumber+1)

	// Emit state change
	if m.stateCallback != nil {
		m.stateCallback(RetryState{
			AttemptNumber:      m.attemptNumber + 1,
			NextRetryIn:        interval,
			NextRetryAt:        time.Now().Add(interval),
			LastError:          m.lastError.Error(),
			LastErrorCode:      m.lastErrorCode,
			IsRetrying:         true,
			MaxIntervalReached: m.attemptNumber >= len(BackoffSchedule)-1,
		})
	}

	m.timer = time.AfterFunc(interval, func() {
		m.doRetry()
	})
}

// doRetry performs the actual retry attempt.
func (m *Manager) doRetry() {
	m.mu.Lock()

	// Check if cancelled
	if m.ctx == nil || m.ctx.Err() != nil {
		m.mu.Unlock()
		return
	}

	callback := m.retryCallback
	m.mu.Unlock()

	if callback == nil {
		log.Printf("[%s] No retry callback configured", m.name)
		return
	}

	log.Printf("[%s] Attempting retry...", m.name)
	err := callback()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if cancelled during callback
	if m.ctx == nil || m.ctx.Err() != nil {
		return
	}

	if err == nil {
		// Success! Reset everything
		log.Printf("[%s] Retry succeeded", m.name)
		m.stopInternal()
		m.lastError = nil
		m.lastErrorCode = ""

		// Emit success state
		if m.stateCallback != nil {
			m.stateCallback(RetryState{
				AttemptNumber: 0,
				IsRetrying:    false,
			})
		}
		return
	}

	// Failed - schedule next retry
	log.Printf("[%s] Retry failed: %v", m.name, err)
	m.lastError = err
	m.attemptNumber++
	m.scheduleNextRetry()
}
