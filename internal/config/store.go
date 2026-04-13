package config

import "sync"

// Store wraps a Config with atomic update semantics. Instead of having
// 15+ call sites that mutate the in-memory config and separately call
// Save(), callers use Store.Update(func) which mutates and persists in
// one operation under a single lock.
//
// This centralizes persistence so future changes (debouncing, atomic
// writes, schema migration) touch one file instead of many.
type Store struct {
	mu     sync.RWMutex
	cfg    *Config
	saveFn func(*Config) error // injectable for tests
}

// NewStore wraps an existing Config in a Store. The saveFn is called on
// every Update; pass config.Save for production use.
func NewStore(cfg *Config, saveFn func(*Config) error) *Store {
	if saveFn == nil {
		saveFn = Save
	}
	return &Store{cfg: cfg, saveFn: saveFn}
}

// Get returns a pointer to the current Config. Callers MUST NOT mutate
// the returned value; for mutations, use Update.
// Returned pointer is safe to read without locking because writes go
// through Update which holds the lock and the Config is never reassigned.
func (s *Store) Get() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// Update atomically mutates the config via the supplied function and
// persists the result to disk. Returns an error if the save fails.
// The mutation function runs under a write lock so concurrent updates
// are serialized.
func (s *Store) Update(mutate func(*Config)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	mutate(s.cfg)
	return s.saveFn(s.cfg)
}

// UpdateNoSave atomically mutates the config without persisting. Use for
// in-memory-only changes (e.g., runtime state that shouldn't survive restart).
func (s *Store) UpdateNoSave(mutate func(*Config)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mutate(s.cfg)
}
