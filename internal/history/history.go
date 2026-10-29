package history

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"plexcord/internal/config"
)

type Entry struct {
	Track     string    `json:"track"`
	Artist    string    `json:"artist"`
	Album     string    `json:"album"`
	Duration  int64     `json:"duration"`
	StartedAt time.Time `json:"startedAt"`
	ThumbURL  string    `json:"thumbUrl,omitempty"`
}

type Stats struct {
	TotalTracks      int    `json:"totalTracks"`
	UniqueArtists    int    `json:"uniqueArtists"`
	MostPlayedArtist string `json:"mostPlayedArtist"`
}

type Store struct {
	mu         sync.RWMutex
	entries    []Entry
	path       string
	maxEntries int
}

func NewStore(configDir string, maxEntries int) *Store {
	s := &Store{
		path:       filepath.Join(configDir, "history.json"),
		maxEntries: maxEntries,
	}
	if err := s.Load(); err != nil {
		log.Printf("Warning: failed to load listening history: %v", err)
	}
	return s
}

func (s *Store) Add(entry Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.entries) > 0 {
		last := s.entries[0]
		if last.Track == entry.Track && last.Artist == entry.Artist && last.Album == entry.Album {
			return
		}
	}

	s.entries = append([]Entry{entry}, s.entries...)

	if len(s.entries) > s.maxEntries {
		s.entries = s.entries[:s.maxEntries]
	}

	if err := s.saveLocked(); err != nil {
		log.Printf("Warning: failed to save listening history: %v", err)
	}
}

func (s *Store) GetRecent(limit int) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.entries) {
		limit = len(s.entries)
	}

	result := make([]Entry, limit)
	copy(result, s.entries[:limit])
	return result
}

func (s *Store) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := Stats{
		TotalTracks: len(s.entries),
	}

	if len(s.entries) == 0 {
		return stats
	}

	artistCounts := make(map[string]int)
	for _, e := range s.entries {
		artistCounts[e.Artist]++
	}

	stats.UniqueArtists = len(artistCounts)

	maxCount := 0
	for artist, count := range artistCounts {
		if count > maxCount {
			maxCount = count
			stats.MostPlayedArtist = artist
		}
	}

	return stats
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = nil
	if err := s.saveLocked(); err != nil {
		log.Printf("Warning: failed to save cleared listening history: %v", err)
	}
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.entries = nil
			return nil
		}
		return err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	s.entries = entries
	return nil
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return config.WriteFileAtomic(s.path, data, 0600)
}
