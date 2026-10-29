package config

import "sync"

type Store struct {
	mu     sync.RWMutex
	cfg    *Config
	saveFn func(*Config) error
}

func NewStore(cfg *Config, saveFn func(*Config) error) *Store {
	if saveFn == nil {
		saveFn = Save
	}
	return &Store{cfg: cfg, saveFn: saveFn}
}

func (s *Store) Get() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *Store) Update(mutate func(*Config)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	mutate(s.cfg)
	return s.saveFn(s.cfg)
}

func (s *Store) UpdateNoSave(mutate func(*Config)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mutate(s.cfg)
}
