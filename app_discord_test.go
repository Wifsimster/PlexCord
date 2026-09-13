package main

import (
	"context"
	"strings"
	"testing"

	"plexcord/internal/config"
	"plexcord/internal/plex"
)

// fakeArtworkResolver returns a preset cached URL.
type fakeArtworkResolver struct {
	cached string
	ok     bool
}

func (f *fakeArtworkResolver) Cached(string, string) (string, bool) { return f.cached, f.ok }
func (f *fakeArtworkResolver) Resolve(context.Context, string, string) (string, error) {
	return f.cached, nil
}

// newPresenceTestApp wires an App around a presence fake with everything the
// presence path touches, and nothing it does not: no Wails window, no tray, no
// keychain, no Plex server.
func newPresenceTestApp(presence DiscordPresence, cfg *config.Config) *App {
	a := newTestApp(cfg)
	a.discord = a.newDiscordService(presence, nil)
	return a
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
	fake := &recordingPresence{connected: true}
	// No resolver: artwork falls back to the Plex logo asset, never the URL.
	a := newPresenceTestApp(fake, config.DefaultConfig())

	a.updateDiscordFromSession(newTokenedSession())

	playback, ok := fake.lastPlayback()
	if !ok {
		t.Fatal("expected a presence update")
	}
	if strings.Contains(playback.ArtworkURL, "X-Plex-Token") {
		t.Errorf("presence artwork URL leaked the Plex token: %q", playback.ArtworkURL)
	}
	if playback.ArtworkURL != "" {
		t.Errorf("expected empty artwork URL (Plex logo fallback), got %q", playback.ArtworkURL)
	}
}

func TestUpdateDiscordFromSession_UsesCachedPublicArtwork(t *testing.T) {
	fake := &recordingPresence{connected: true}
	a := newTestApp(config.DefaultConfig())
	a.discord = a.newDiscordService(fake, &fakeArtworkResolver{cached: "https://cdn/cover-512.jpg", ok: true})

	a.updateDiscordFromSession(newTokenedSession())

	playback, _ := fake.lastPlayback()
	if playback.ArtworkURL != "https://cdn/cover-512.jpg" {
		t.Errorf("expected cached public cover URL, got %q", playback.ArtworkURL)
	}
}

func TestGetPresenceOptions_NormalizesDefaults(t *testing.T) {
	// A legacy config with empty presence options should read back as the
	// media/state/artwork-on defaults.
	a := newTestApp(&config.Config{})
	opts := a.GetPresenceOptions()
	if opts.ActivityStyle != "media" {
		t.Errorf("ActivityStyle = %q, want media", opts.ActivityStyle)
	}
	if opts.StatusDisplay != "state" {
		t.Errorf("StatusDisplay = %q, want state", opts.StatusDisplay)
	}
	if !opts.ArtworkLookup {
		t.Error("ArtworkLookup should default to true for a legacy config")
	}
}

func TestSetPresenceOptions_RejectsInvalidValues(t *testing.T) {
	a := newTestApp(config.DefaultConfig())

	if err := a.SetPresenceOptions(PresenceOptions{ActivityStyle: "bogus", StatusDisplay: "state"}); err == nil {
		t.Error("expected error for invalid activity style")
	}
	if err := a.SetPresenceOptions(PresenceOptions{ActivityStyle: "media", StatusDisplay: "bogus"}); err == nil {
		t.Error("expected error for invalid status display")
	}
	// Config must be untouched after a rejected update.
	if a.config.PresenceActivityStyle != "media" {
		t.Errorf("config mutated on rejected update: %q", a.config.PresenceActivityStyle)
	}
}

func TestUpdateDiscordFromSession_ArtworkLookupDisabled(t *testing.T) {
	fake := &recordingPresence{connected: true}
	cfg := config.DefaultConfig()
	disabled := false
	cfg.PresenceArtworkLookup = &disabled
	// Even with a resolver that has a cached cover, lookup-disabled sends none.
	a := newTestApp(cfg)
	a.discord = a.newDiscordService(fake, &fakeArtworkResolver{cached: "https://cdn/cover.jpg", ok: true})

	a.updateDiscordFromSession(newTokenedSession())

	playback, _ := fake.lastPlayback()
	if playback.ArtworkURL != "" {
		t.Errorf("artwork lookup disabled should send no URL, got %q", playback.ArtworkURL)
	}
}
