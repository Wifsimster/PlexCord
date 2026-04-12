package retry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// testBackoffSchedule overrides the package-level BackoffSchedule for fast tests.
func withShortBackoff(t *testing.T, schedule []time.Duration) {
	t.Helper()
	orig := make([]time.Duration, len(BackoffSchedule))
	copy(orig, BackoffSchedule)
	BackoffSchedule = schedule
	t.Cleanup(func() {
		BackoffSchedule = orig
	})
}

func TestNewManager(t *testing.T) {
	m := NewManager("test")
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.name != "test" {
		t.Errorf("expected name 'test', got %q", m.name)
	}
	if m.running {
		t.Error("expected running to be false")
	}
	if m.attemptNumber != 0 {
		t.Errorf("expected attemptNumber 0, got %d", m.attemptNumber)
	}
	if m.lastError != nil {
		t.Error("expected lastError to be nil")
	}

	state := m.GetState()
	if state.IsRetrying {
		t.Error("expected IsRetrying to be false for new manager")
	}
	if state.AttemptNumber != 0 {
		t.Error("expected AttemptNumber 0 for new manager")
	}
	if state.LastError != "" {
		t.Error("expected empty LastError for new manager")
	}
}

func TestStartAndSchedule(t *testing.T) {
	withShortBackoff(t, []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
	})

	m := NewManager("test-start")

	// Use a channel to block the callback so we can inspect state while running
	callbackCh := make(chan struct{}, 1)
	m.SetCallbacks(func() error {
		<-callbackCh
		return errors.New("keep retrying")
	}, nil)

	testErr := errors.New("connection failed")
	m.Start(testErr, "CONN_FAIL")

	state := m.GetState()
	if !state.IsRetrying {
		t.Error("expected IsRetrying to be true after Start")
	}
	if state.LastError != "connection failed" {
		t.Errorf("expected LastError 'connection failed', got %q", state.LastError)
	}
	if state.LastErrorCode != "CONN_FAIL" {
		t.Errorf("expected LastErrorCode 'CONN_FAIL', got %q", state.LastErrorCode)
	}

	// Unblock callback and cleanup
	callbackCh <- struct{}{}
	// Give the retry goroutine time to fire and re-schedule
	time.Sleep(50 * time.Millisecond)
	m.Stop()
}

func TestStopCancels(t *testing.T) {
	withShortBackoff(t, []time.Duration{
		500 * time.Millisecond, // long enough that it won't fire during the test
	})

	callbackCalled := make(chan struct{}, 1)
	m := NewManager("test-stop")
	m.SetCallbacks(func() error {
		callbackCalled <- struct{}{}
		return nil
	}, nil)

	m.Start(errors.New("err"), "")

	// Stop immediately before the timer fires
	m.Stop()

	// Verify state
	state := m.GetState()
	if state.IsRetrying {
		t.Error("expected IsRetrying to be false after Stop")
	}

	// Wait to make sure the callback doesn't fire
	select {
	case <-callbackCalled:
		t.Error("callback should not have been called after Stop")
	case <-time.After(100 * time.Millisecond):
		// good
	}
}

func TestResetClearsState(t *testing.T) {
	withShortBackoff(t, []time.Duration{500 * time.Millisecond})

	stateEmitted := make(chan RetryState, 1)
	m := NewManager("test-reset")
	m.SetCallbacks(func() error {
		return errors.New("fail")
	}, func(state RetryState) {
		// Non-blocking send; we only care about the last emission from Reset
		select {
		case stateEmitted <- state:
		default:
			// drain and re-send to keep latest
			<-stateEmitted
			stateEmitted <- state
		}
	})

	m.Start(errors.New("some error"), "SOME_CODE")

	// Drain the state emitted by Start's scheduleNextRetry
	select {
	case <-stateEmitted:
	case <-time.After(100 * time.Millisecond):
	}

	m.Reset()

	// Check the emitted state from Reset
	select {
	case s := <-stateEmitted:
		if s.IsRetrying {
			t.Error("expected IsRetrying false in reset state emission")
		}
		if s.AttemptNumber != 0 {
			t.Error("expected AttemptNumber 0 in reset state emission")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected state callback to be called on Reset")
	}

	// Verify internal state
	state := m.GetState()
	if state.IsRetrying {
		t.Error("expected IsRetrying false after Reset")
	}
	if state.LastError != "" {
		t.Errorf("expected empty LastError after Reset, got %q", state.LastError)
	}
	if state.LastErrorCode != "" {
		t.Errorf("expected empty LastErrorCode after Reset, got %q", state.LastErrorCode)
	}
}

func TestManualRetryResetsBackoff(t *testing.T) {
	withShortBackoff(t, []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
	})

	attempts := make(chan int, 10)
	m := NewManager("test-manual")
	m.SetCallbacks(func() error {
		return errors.New("still failing")
	}, func(state RetryState) {
		select {
		case attempts <- state.AttemptNumber:
		default:
		}
	})

	// Start, which will begin at attempt 0
	m.Start(errors.New("initial"), "")

	// Wait for a couple of retries to increase the attempt number
	time.Sleep(80 * time.Millisecond)

	// ManualRetry should reset attempt number to 0
	m.ManualRetry()

	// Wait for state callback from the immediate retry's reschedule
	time.Sleep(40 * time.Millisecond)

	// After ManualRetry, the doRetry will increment attemptNumber to 1,
	// then scheduleNextRetry emits AttemptNumber = attemptNumber + 1 = 1 + 1 = 2.
	// But the key thing is attempt was reset from a higher value.
	state := m.GetState()
	// The attempt number should be low (reset happened)
	if state.AttemptNumber > 3 {
		t.Errorf("expected low attempt number after ManualRetry, got %d", state.AttemptNumber)
	}

	m.Stop()
}

func TestBackoffSchedule(t *testing.T) {
	// Test the default backoff schedule intervals
	m := NewManager("test-backoff")

	expected := []time.Duration{
		5 * time.Second,
		10 * time.Second,
		30 * time.Second,
		60 * time.Second,
	}

	for i, exp := range expected {
		got := m.getInterval(i)
		if got != exp {
			t.Errorf("attempt %d: expected %v, got %v", i, exp, got)
		}
	}

	// Past the schedule length, it should cap at the last value
	got := m.getInterval(10)
	if got != 60*time.Second {
		t.Errorf("attempt 10: expected 60s cap, got %v", got)
	}

	got = m.getInterval(100)
	if got != 60*time.Second {
		t.Errorf("attempt 100: expected 60s cap, got %v", got)
	}
}

func TestCallbackSuccess(t *testing.T) {
	withShortBackoff(t, []time.Duration{10 * time.Millisecond})

	successEmitted := make(chan struct{}, 1)
	m := NewManager("test-success")
	m.SetCallbacks(func() error {
		return nil // success on first try
	}, func(state RetryState) {
		if !state.IsRetrying && state.AttemptNumber == 0 {
			select {
			case successEmitted <- struct{}{}:
			default:
			}
		}
	})

	m.Start(errors.New("trigger"), "")

	select {
	case <-successEmitted:
		// good
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for success state emission")
	}

	state := m.GetState()
	if state.IsRetrying {
		t.Error("expected IsRetrying to be false after successful callback")
	}
	if state.LastError != "" {
		t.Errorf("expected empty LastError after success, got %q", state.LastError)
	}
}

func TestCallbackFailure(t *testing.T) {
	withShortBackoff(t, []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
	})

	attemptCount := 0
	var mu sync.Mutex
	reachedSecondAttempt := make(chan struct{}, 1)

	m := NewManager("test-failure")
	m.SetCallbacks(func() error {
		mu.Lock()
		attemptCount++
		current := attemptCount
		mu.Unlock()
		if current >= 2 {
			select {
			case reachedSecondAttempt <- struct{}{}:
			default:
			}
		}
		return errors.New("still failing")
	}, nil)

	m.Start(errors.New("initial failure"), "")

	select {
	case <-reachedSecondAttempt:
		// good - retry scheduled and executed at least twice
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for second retry attempt")
	}

	state := m.GetState()
	if !state.IsRetrying {
		t.Error("expected IsRetrying to be true while callback keeps failing")
	}

	m.Stop()
}

func TestConcurrentManualRetry(t *testing.T) {
	withShortBackoff(t, []time.Duration{10 * time.Millisecond})

	m := NewManager("test-concurrent")
	m.SetCallbacks(func() error {
		return errors.New("fail")
	}, nil)

	m.Start(errors.New("err"), "")

	// Hammer ManualRetry from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.ManualRetry()
		}()
	}
	wg.Wait()

	// If we get here without panic, test passes
	m.Stop()
}

func TestStartWhileRunning(t *testing.T) {
	withShortBackoff(t, []time.Duration{500 * time.Millisecond})

	m := NewManager("test-start-running")
	m.SetCallbacks(func() error {
		return errors.New("fail")
	}, nil)

	m.Start(errors.New("first error"), "CODE1")

	state1 := m.GetState()
	if !state1.IsRetrying {
		t.Fatal("expected IsRetrying after first Start")
	}
	if state1.LastError != "first error" {
		t.Errorf("expected 'first error', got %q", state1.LastError)
	}

	// Start again while running - should update error but not restart
	m.Start(errors.New("second error"), "CODE2")

	state2 := m.GetState()
	if !state2.IsRetrying {
		t.Error("expected still IsRetrying after second Start")
	}
	if state2.LastError != "second error" {
		t.Errorf("expected 'second error', got %q", state2.LastError)
	}
	if state2.LastErrorCode != "CODE2" {
		t.Errorf("expected 'CODE2', got %q", state2.LastErrorCode)
	}

	// Attempt number should not have reset (no restart)
	if state2.AttemptNumber != state1.AttemptNumber {
		t.Errorf("attempt number should not change on re-start: was %d, now %d",
			state1.AttemptNumber, state2.AttemptNumber)
	}

	m.Stop()
}

func TestGetStateAccuracy(t *testing.T) {
	withShortBackoff(t, []time.Duration{200 * time.Millisecond})

	m := NewManager("test-state-accuracy")
	m.SetCallbacks(func() error {
		return errors.New("fail")
	}, nil)

	m.Start(errors.New("err"), "")

	state := m.GetState()
	if state.NextRetryAt.IsZero() {
		t.Fatal("expected NextRetryAt to be set")
	}

	// NextRetryAt should be stored, not recalculated. Verify it stays the same
	// across multiple GetState calls.
	state2 := m.GetState()
	if !state.NextRetryAt.Equal(state2.NextRetryAt) {
		t.Errorf("NextRetryAt changed between calls: %v vs %v", state.NextRetryAt, state2.NextRetryAt)
	}

	// NextRetryIn should be approximately 200ms minus elapsed time
	if state.NextRetryIn <= 0 || state.NextRetryIn > 200*time.Millisecond {
		t.Errorf("NextRetryIn out of range: %v", state.NextRetryIn)
	}

	m.Stop()
}

func TestStateCallback(t *testing.T) {
	withShortBackoff(t, []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
	})

	states := make(chan RetryState, 10)
	m := NewManager("test-state-cb")
	m.SetCallbacks(func() error {
		return errors.New("keep failing")
	}, func(state RetryState) {
		states <- state
	})

	m.Start(errors.New("trigger error"), "TRIGGER")

	// First state emission from scheduleNextRetry in Start
	select {
	case s := <-states:
		if !s.IsRetrying {
			t.Error("expected IsRetrying true in first state emission")
		}
		if s.AttemptNumber != 1 {
			t.Errorf("expected AttemptNumber 1, got %d", s.AttemptNumber)
		}
		if s.LastError != "trigger error" {
			t.Errorf("expected 'trigger error', got %q", s.LastError)
		}
		if s.LastErrorCode != "TRIGGER" {
			t.Errorf("expected 'TRIGGER', got %q", s.LastErrorCode)
		}
		if s.NextRetryAt.IsZero() {
			t.Error("expected NextRetryAt to be set")
		}
		if s.NextRetryIn != 10*time.Millisecond {
			t.Errorf("expected NextRetryIn 10ms, got %v", s.NextRetryIn)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for first state emission")
	}

	// After the first retry fails, a second state emission happens
	select {
	case s := <-states:
		if s.AttemptNumber != 2 {
			t.Errorf("expected AttemptNumber 2 in second emission, got %d", s.AttemptNumber)
		}
		if s.LastError != "keep failing" {
			t.Errorf("expected 'keep failing' in second emission, got %q", s.LastError)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for second state emission")
	}

	m.Stop()
}

func TestNilLastError(t *testing.T) {
	withShortBackoff(t, []time.Duration{10 * time.Millisecond})

	states := make(chan RetryState, 5)
	m := NewManager("test-nil-error")
	m.SetCallbacks(func() error {
		return nil
	}, func(state RetryState) {
		states <- state
	})

	// Manually set up the manager with nil lastError and call scheduleNextRetry
	// to exercise the nil-check path in scheduleNextRetry
	m.mu.Lock()
	m.lastError = nil
	m.lastErrorCode = ""
	m.running = true
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.scheduleNextRetry()
	m.mu.Unlock()

	// Should emit state with empty LastError, not panic
	select {
	case s := <-states:
		if s.LastError != "" {
			t.Errorf("expected empty LastError for nil error, got %q", s.LastError)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for state emission")
	}

	m.Stop()
}
