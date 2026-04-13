package plex

import "testing"

// These tests exercise the pure filter/map functions without any HTTP.
// They demonstrate the testability win from splitting Client into
// transport + mapper layers.

func TestFilterMusicSessions_FiltersByUser(t *testing.T) {
	resp := &SessionsResponse{
		Tracks: []SessionEntry{
			{
				SessionKey: "1",
				Type:       "track",
				Title:      "Song A",
				User:       SessionUser{ID: "alice"},
				Player:     SessionPlayer{State: "playing"},
			},
			{
				SessionKey: "2",
				Type:       "track",
				Title:      "Song B",
				User:       SessionUser{ID: "bob"},
				Player:     SessionPlayer{State: "playing"},
			},
		},
	}

	got := filterMusicSessions(resp, "alice", nil)

	if len(got) != 1 {
		t.Fatalf("expected 1 session for alice, got %d", len(got))
	}
	if got[0].Track != "Song A" {
		t.Errorf("expected Song A, got %s", got[0].Track)
	}
}

func TestFilterMusicSessions_EmptyUserIDReturnsAll(t *testing.T) {
	resp := &SessionsResponse{
		Tracks: []SessionEntry{
			{Type: "track", Title: "A", User: SessionUser{ID: "alice"}},
			{Type: "track", Title: "B", User: SessionUser{ID: "bob"}},
		},
	}

	got := filterMusicSessions(resp, "", nil)

	if len(got) != 2 {
		t.Errorf("expected 2 sessions when userID empty, got %d", len(got))
	}
}

func TestFilterMusicSessions_AppliesFallbacks(t *testing.T) {
	resp := &SessionsResponse{
		Tracks: []SessionEntry{
			{Type: "track", User: SessionUser{ID: "alice"}},
		},
	}

	got := filterMusicSessions(resp, "alice", nil)

	if len(got) != 1 {
		t.Fatalf("expected 1 session, got %d", len(got))
	}
	if got[0].Track == "" || got[0].Artist == "" || got[0].Album == "" {
		t.Errorf("fallbacks not applied: track=%q artist=%q album=%q",
			got[0].Track, got[0].Artist, got[0].Album)
	}
}

func TestFilterMusicSessions_BuildsArtworkURL(t *testing.T) {
	resp := &SessionsResponse{
		Tracks: []SessionEntry{
			{Type: "track", Title: "X", Thumb: "/library/thumb/123", User: SessionUser{ID: "u"}},
		},
	}
	builder := func(thumb string) string { return "http://server" + thumb + "?token=abc" }

	got := filterMusicSessions(resp, "u", builder)

	if got[0].ThumbURL != "http://server/library/thumb/123?token=abc" {
		t.Errorf("unexpected ThumbURL: %s", got[0].ThumbURL)
	}
}

func TestFilterMediaSessions_FiltersByMediaType(t *testing.T) {
	resp := &SessionsResponse{
		Tracks: []SessionEntry{
			{Type: "track", Title: "Song", User: SessionUser{ID: "u"}},
		},
		Videos: []SessionEntry{
			{Type: "movie", Title: "Movie", User: SessionUser{ID: "u"}},
			{Type: "episode", Title: "Episode", User: SessionUser{ID: "u"}},
		},
		Photos: []SessionEntry{
			{Type: "photo", Title: "Photo", User: SessionUser{ID: "u"}},
		},
	}

	// Only music
	got := filterMediaSessions(resp, "u", []string{"music"}, nil)
	if len(got) != 1 {
		t.Errorf("expected 1 music session, got %d", len(got))
	}

	// Music + movie
	got = filterMediaSessions(resp, "u", []string{"music", "movie"}, nil)
	if len(got) != 2 {
		t.Errorf("expected 2 sessions (music + movie), got %d", len(got))
	}

	// Empty filter = all
	got = filterMediaSessions(resp, "u", nil, nil)
	if len(got) != 4 {
		t.Errorf("expected 4 sessions with empty filter, got %d", len(got))
	}
}

func TestFilterMediaSessions_FiltersByUser(t *testing.T) {
	resp := &SessionsResponse{
		Tracks: []SessionEntry{
			{Type: "track", Title: "A", User: SessionUser{ID: "alice"}},
			{Type: "track", Title: "B", User: SessionUser{ID: "bob"}},
		},
	}

	got := filterMediaSessions(resp, "alice", nil, nil)

	if len(got) != 1 {
		t.Errorf("expected 1 session for alice, got %d", len(got))
	}
}

func TestParseSessionsResponse_ValidXML(t *testing.T) {
	xml := []byte(`<?xml version="1.0"?>
<MediaContainer size="1">
  <Track sessionKey="42" type="track" title="Test Song">
    <User id="alice" title="Alice"/>
    <Player state="playing" title="Plexamp"/>
  </Track>
</MediaContainer>`)

	resp, err := parseSessionsResponse(xml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(resp.Tracks))
	}
	if resp.Tracks[0].Title != "Test Song" {
		t.Errorf("expected 'Test Song', got %s", resp.Tracks[0].Title)
	}
	if resp.Tracks[0].User.ID != "alice" {
		t.Errorf("expected user alice, got %s", resp.Tracks[0].User.ID)
	}
}

func TestParseSessionsResponse_InvalidXML(t *testing.T) {
	_, err := parseSessionsResponse([]byte("not xml"))
	if err == nil {
		t.Error("expected error parsing invalid XML")
	}
}
