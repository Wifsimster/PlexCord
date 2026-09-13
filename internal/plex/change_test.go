package plex

import "testing"

// Change detection decides whether the rest of PlexCord hears about a poll.
// ViewOffset advances on every single poll, so the one thing these tests must
// pin down is that position alone never counts as a change — otherwise Discord
// would be rewritten several times a second.

func TestSessionChanged(t *testing.T) {
	base := func() *MusicSession {
		m := &MusicSession{Track: "Song", Artist: "Artist", Album: "Album", ViewOffset: 1000}
		m.SessionKey = "key"
		m.State = "playing"
		return m
	}

	tests := []struct {
		name   string
		mutate func(*MusicSession)
		want   bool
	}{
		{"identical", func(*MusicSession) {}, false},
		{"position advanced", func(m *MusicSession) { m.ViewOffset = 45000 }, false},
		{"duration corrected", func(m *MusicSession) { m.Duration = 240000 }, false},
		{"track changed", func(m *MusicSession) { m.Track = "Other" }, true},
		{"paused", func(m *MusicSession) { m.State = "paused" }, true},
		{"new session key", func(m *MusicSession) { m.SessionKey = "other" }, true},
		{"metadata refreshed", func(m *MusicSession) { m.Artist = "Real Artist" }, true},
		{"album filled in", func(m *MusicSession) { m.Album = "Real Album" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev, curr := base(), base()
			tt.mutate(curr)
			if got := sessionChanged(prev, curr); got != tt.want {
				t.Errorf("sessionChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionChangedNilHandling(t *testing.T) {
	session := &MusicSession{Track: "Song"}

	if sessionChanged(nil, nil) {
		t.Error("nil → nil reported a change; nothing was playing and nothing is")
	}
	if !sessionChanged(nil, session) {
		t.Error("nil → session reported no change; playback started")
	}
	if !sessionChanged(session, nil) {
		t.Error("session → nil reported no change; playback stopped")
	}
}

func TestMediaSessionChanged(t *testing.T) {
	base := func() *MediaSession {
		return &MediaSession{
			SessionKey: "key",
			State:      "playing",
			MediaType:  MediaTypeTV,
			Title:      "Episode",
			ShowTitle:  "Show",
			Season:     1,
			Episode:    2,
			ViewOffset: 1000,
		}
	}

	tests := []struct {
		name   string
		mutate func(*MediaSession)
		want   bool
	}{
		{"identical", func(*MediaSession) {}, false},
		{"position advanced", func(m *MediaSession) { m.ViewOffset = 90000 }, false},
		{"next episode", func(m *MediaSession) { m.Episode = 3 }, true},
		{"next season", func(m *MediaSession) { m.Season = 2 }, true},
		{"different show", func(m *MediaSession) { m.ShowTitle = "Other Show" }, true},
		{"switched to music", func(m *MediaSession) { m.MediaType = MediaTypeMusic }, true},
		{"artist changed", func(m *MediaSession) { m.Artist = "Artist" }, true},
		{"paused", func(m *MediaSession) { m.State = "paused" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev, curr := base(), base()
			tt.mutate(curr)
			if got := mediaSessionChanged(prev, curr); got != tt.want {
				t.Errorf("mediaSessionChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMediaSessionChangedNilHandling(t *testing.T) {
	session := &MediaSession{Title: "Episode"}

	if mediaSessionChanged(nil, nil) {
		t.Error("nil → nil reported a change")
	}
	if !mediaSessionChanged(nil, session) {
		t.Error("nil → session reported no change")
	}
	if !mediaSessionChanged(session, nil) {
		t.Error("session → nil reported no change")
	}
}
