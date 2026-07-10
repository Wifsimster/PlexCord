package artwork

import (
	"sync"
	"time"
)

// rateLimiter enforces a minimum interval between successive calls to wait().
// A zero interval makes wait() a no-op (used in tests).
type rateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

func newRateLimiter(interval time.Duration) *rateLimiter {
	return &rateLimiter{interval: interval}
}

// wait blocks until at least `interval` has elapsed since the previous call,
// or until ctx-independent time passes. Callers hold no locks across wait.
func (l *rateLimiter) wait() {
	if l == nil || l.interval <= 0 {
		return
	}
	l.mu.Lock()
	now := time.Now()
	if !l.last.IsZero() {
		if d := l.interval - now.Sub(l.last); d > 0 {
			l.last = now.Add(d)
			l.mu.Unlock()
			time.Sleep(d)
			return
		}
	}
	l.last = now
	l.mu.Unlock()
}
