package main

import (
	"context"
	"strings"
	"testing"

	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/plex"
)

// fakeDiscordPresence records the arguments of the last presence update so
// tests can assert what PlexCord sends to Discord.
type fakeDiscordPresence struct {
	connected      bool
	updateCount    int
	lastArtworkURL string
	lastTrack      string
}

func (f *fakeDiscordPresence) Connect(string) error { return nil }
func (f *fakeDiscordPresence) Disconnect() error    { return nil }
func (f *fakeDiscordPresence) IsConnected() bool    { return f.connected }
func (f *fakeDiscordPresence) GetClientID() string  { return "" }
func (f *fakeDiscordPresence) SetPresence(*discord.PresenceData) error {
	return nil
}
func (f *fakeDiscordPresence) ClearPresence() error { return nil }
func (f *fakeDiscordPresence) UpdatePresenceFromPlayback(track, artist, album, state string, duration, position int64, artworkURL, player, detailsFormat, stateFormat, activityStyle, statusDisplay string) error {
	f.updateCount++
	f.lastArtworkURL = artworkURL
	f.lastTrack = track
	return nil
}

// fakeArtworkResolver returns a preset cached URL.
type fakeArtworkResolver struct {
	cached string
	ok     bool
}

func (f *fakeArtworkResolver) Cached(string, string) (string, bool) { return f.cached, f.ok }
func (f *fakeArtworkResolver) Resolve(context.Context, string, string) (string, error) {
	return f.cached, nil
}

func newTokenedSession() *plex.MusicSession {
	s := &plex.MusicSession{
		Track:    "Song",
		Artist:   "Artist",
		Album:    "Album",
		ThumbURL: "http://192.168.1.5:32400/library/metadata/1/thumb/1?X-Plex-Token=secret-token",
		Duration: 240000,
	}
	s.State = "playing"
	return s
}

func TestUpdateDiscordFromSession_NeverSendsPlexToken(t *testing.T) {
	fake := &fakeDiscordPresence{connected: true}
	a := &App{
		discord: fake,
		config:  config.DefaultConfig(),
		// No resolver: artwork falls back to the Plex logo asset, never the URL.
		artwork: nil,
	}

	a.updateDiscordFromSession(newTokenedSession())

	if fake.updateCount == 0 {
		t.Fatal("expected a presence update")
	}
	if strings.Contains(fake.lastArtworkURL, "X-Plex-Token") {
		t.Errorf("presence artwork URL leaked the Plex token: %q", fake.lastArtworkURL)
	}
	if fake.lastArtworkURL != "" {
		t.Errorf("expected empty artwork URL (Plex logo fallback), got %q", fake.lastArtworkURL)
	}
}

func TestUpdateDiscordFromSession_UsesCachedPublicArtwork(t *testing.T) {
	fake := &fakeDiscordPresence{connected: true}
	a := &App{
		discord: fake,
		config:  config.DefaultConfig(),
		artwork: &fakeArtworkResolver{cached: "https://cdn/cover-512.jpg", ok: true},
	}

	a.updateDiscordFromSession(newTokenedSession())

	if fake.lastArtworkURL != "https://cdn/cover-512.jpg" {
		t.Errorf("expected cached public cover URL, got %q", fake.lastArtworkURL)
	}
}

func TestUpdateDiscordFromSession_ArtworkLookupDisabled(t *testing.T) {
	fake := &fakeDiscordPresence{connected: true}
	cfg := config.DefaultConfig()
	disabled := false
	cfg.PresenceArtworkLookup = &disabled
	a := &App{
		discord: fake,
		config:  cfg,
		// Even with a resolver that has a cached cover, lookup-disabled sends none.
		artwork: &fakeArtworkResolver{cached: "https://cdn/cover.jpg", ok: true},
	}

	a.updateDiscordFromSession(newTokenedSession())

	if fake.lastArtworkURL != "" {
		t.Errorf("artwork lookup disabled should send no URL, got %q", fake.lastArtworkURL)
	}
}
