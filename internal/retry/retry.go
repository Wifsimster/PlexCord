package retry

import (
	"context"
	"log"
	"sync"
	"time"
)

var BackoffSchedule = []time.Duration{
	5 * time.Second,
	10 * time.Second,
	30 * time.Second,
	60 * time.Second,
}

type RetryState struct {
	NextRetryAt        time.Time     `json:"nextRetryAt"`
	LastError          string        `json:"lastError,omitempty"`
	LastErrorCode      string        `json:"lastErrorCode,omitempty"`
	NextRetryIn        time.Duration `json:"nextRetryIn"`
	AttemptNumber      int           `json:"attemptNumber"`
	IsRetrying         bool          `json:"isRetrying"`
	MaxIntervalReached bool          `json:"maxIntervalReached"`
}

type RetryCallback func() error

type StateChangeCallback func(state RetryState)

type Manager struct {
	lastError     error
	timer         *time.Timer
	ctx           context.Context
	cancel        context.CancelFunc
	retryCallback RetryCallback
	stateCallback StateChangeCallback
	name          string
	lastErrorCode string
	nextRetryAt   time.Time
	mu            sync.Mutex
	attemptNumber int
	running       bool
}

func NewManager(name string) *Manager {
	return &Manager{
		name: name,
	}
}

func (m *Manager) SetCallbacks(retry RetryCallback, stateChange StateChangeCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retryCallback = retry
	m.stateCallback = stateChange
}

func (m *Manager) Start(err error, code string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		m.lastError = err
		m.lastErrorCode = code
		return
	}

	m.lastError = err
	m.lastErrorCode = code
	m.attemptNumber = 0
	m.running = true

	m.ctx, m.cancel = context.WithCancel(context.Background())

	m.scheduleNextRetry()
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopInternal()
}

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

func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopInternal()
	m.lastError = nil
	m.lastErrorCode = ""

	if m.stateCallback != nil {
		m.stateCallback(RetryState{
			AttemptNumber: 0,
			IsRetrying:    false,
		})
	}
}

func (m *Manager) ManualRetry() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ctx == nil || m.ctx.Err() != nil {
		m.ctx, m.cancel = context.WithCancel(context.Background())
		m.running = true
	}

	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}

	m.attemptNumber = 0

	go m.doRetry()
}

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
		state.NextRetryIn = time.Until(m.nextRetryAt)
		state.NextRetryAt = m.nextRetryAt
	}

	return state
}

func (m *Manager) IsRetrying() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *Manager) getInterval(attempt int) time.Duration {
	if attempt >= len(BackoffSchedule) {
		return BackoffSchedule[len(BackoffSchedule)-1]
	}
	return BackoffSchedule[attempt]
}

func (m *Manager) scheduleNextRetry() {
	interval := m.getInterval(m.attemptNumber)
	m.nextRetryAt = time.Now().Add(interval)

	log.Printf("[%s] Retry scheduled in %v (attempt %d)", m.name, interval, m.attemptNumber+1)

	if m.stateCallback != nil {
		lastErr := ""
		if m.lastError != nil {
			lastErr = m.lastError.Error()
		}
		m.stateCallback(RetryState{
			AttemptNumber:      m.attemptNumber + 1,
			NextRetryIn:        interval,
			NextRetryAt:        m.nextRetryAt,
			LastError:          lastErr,
			LastErrorCode:      m.lastErrorCode,
			IsRetrying:         true,
			MaxIntervalReached: m.attemptNumber >= len(BackoffSchedule)-1,
		})
	}

	m.timer = time.AfterFunc(interval, func() {
		m.doRetry()
	})
}

func (m *Manager) doRetry() {
	m.mu.Lock()

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

	if m.ctx == nil || m.ctx.Err() != nil {
		return
	}

	if err == nil {
		log.Printf("[%s] Retry succeeded", m.name)
		m.stopInternal()
		m.lastError = nil
		m.lastErrorCode = ""

		if m.stateCallback != nil {
			m.stateCallback(RetryState{
				AttemptNumber: 0,
				IsRetrying:    false,
			})
		}
		return
	}

	log.Printf("[%s] Retry failed: %v", m.name, err)
	m.lastError = err
	m.attemptNumber++
	m.scheduleNextRetry()
}
