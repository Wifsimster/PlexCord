package history

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a single listening history entry.
type Entry struct {
	Track     string    `json:"track"`
	Artist    string    `json:"artist"`
	Album     string    `json:"album"`
	Duration  int64     `json:"duration"`
	StartedAt time.Time `json:"startedAt"`
	ThumbURL  string    `json:"thumbUrl,omitempty"`
}

// Stats contains aggregate listening statistics.
type Stats struct {
	TotalTracks      int    `json:"totalTracks"`
	UniqueArtists    int    `json:"uniqueArtists"`
	MostPlayedArtist string `json:"mostPlayedArtist"`
}

// Store manages listening history persistence and retrieval.
type Store struct {
	mu         sync.RWMutex
	entries    []Entry
	path       string
	maxEntries int
}

// NewStore creates a new history store.
// configDir is the directory where the history JSON file will be saved.
// maxEntries controls how many entries are retained (oldest are trimmed).
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

// Add inserts a new entry at the front of the history.
// It deduplicates against the most recent entry to prevent poll-driven duplicates
// (same track+artist+album as the last entry is skipped).
// The list is trimmed to maxEntries and auto-saved.
func (s *Store) Add(entry Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deduplicate: skip if identical to the last entry
	if len(s.entries) > 0 {
		last := s.entries[0]
		if last.Track == entry.Track && last.Artist == entry.Artist && last.Album == entry.Album {
			return
		}
	}

	// Prepend
	s.entries = append([]Entry{entry}, s.entries...)

	// Trim to max
	if len(s.entries) > s.maxEntries {
		s.entries = s.entries[:s.maxEntries]
	}

	// Auto-save (best-effort)
	if err := s.saveLocked(); err != nil {
		log.Printf("Warning: failed to save listening history: %v", err)
	}
}

// GetRecent returns the most recent N entries.
// If limit <= 0 or exceeds the number of entries, all entries are returned.
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

// GetStats returns aggregate listening statistics.
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

	// Find most played artist
	maxCount := 0
	for artist, count := range artistCounts {
		if count > maxCount {
			maxCount = count
			stats.MostPlayedArtist = artist
		}
	}

	return stats
}

// Clear removes all history entries and saves the empty state.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = nil
	if err := s.saveLocked(); err != nil {
		log.Printf("Warning: failed to save cleared listening history: %v", err)
	}
}

// Load reads the history from disk. It is called automatically by NewStore.
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

// Save writes the current history to disk.
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.saveLocked()
}

// saveLocked performs the actual file write. Caller must hold at least a read lock.
func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0600)
}
