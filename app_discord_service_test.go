package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/plex"
)

// blockingResolver returns a cover only after release is closed, so a test can
// control exactly when a background artwork lookup lands.
type blockingResolver struct {
	url     string
	release chan struct{}
}

func (r *blockingResolver) Cached(string, string) (string, bool) { return "", false }

func (r *blockingResolver) Resolve(ctx context.Context, _, _ string) (string, error) {
	select {
	case <-r.release:
		return r.url, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func playingSession(track string) *plex.MusicSession {
	// A distinct album per track, so an artwork lookup can be keyed on which
	// track asked for it.
	s := &plex.MusicSession{Track: track, Artist: "Artist", Album: track + " Album", Duration: 1000}
	s.State = "playing"
	return s
}

// TestDiscordServiceReconnectsBeforePublishing covers Discord restarting
// mid-playback: the link is down when a session arrives, and the presence must
// still go out once the silent reconnect succeeds.
func TestDiscordServiceReconnectsBeforePublishing(t *testing.T) {
	presence := &recordingPresence{connected: false}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, nil)

	app.discord.Publish(playingSession("Song"))

	if presence.updateCount() != 1 {
		t.Fatalf("presence updates = %d, want 1 after a reconnect", presence.updateCount())
	}
	if len(presence.connectCalls) != 1 {
		t.Fatalf("connect calls = %d, want 1", len(presence.connectCalls))
	}
	if presence.connectCalls[0] != discord.DefaultClientID {
		t.Errorf("reconnected with client ID %q, want the default", presence.connectCalls[0])
	}
}

// TestDiscordServiceSkipsPublishWhenDiscordIsDown verifies a failed reconnect
// is silent rather than an error storm: Discord simply not running is the
// expected case on every poll.
func TestDiscordServiceSkipsPublishWhenDiscordIsDown(t *testing.T) {
	presence := &recordingPresence{connected: false, connectErr: errors.New("discord not running")}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, nil)

	app.discord.Publish(playingSession("Song"))

	if presence.updateCount() != 0 {
		t.Errorf("presence updates = %d, want 0 when Discord cannot be reached", presence.updateCount())
	}
}

// TestDiscordServicePublishesUserDisplayOptions verifies the persisted
// preferences reach Discord on the presence path.
func TestDiscordServicePublishesUserDisplayOptions(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PresenceDetailsFormat = "{track}"
	cfg.PresenceStateFormat = "by {artist}"
	cfg.PresenceActivityStyle = discord.ActivityStyleGame
	cfg.PresenceStatusDisplay = discord.StatusDisplayDetails

	presence := &recordingPresence{connected: true}
	app := newTestApp(cfg)
	app.discord = app.newDiscordService(presence, nil)

	app.discord.Publish(playingSession("Song"))

	want := discord.Options{
		DetailsFormat: "{track}",
		StateFormat:   "by {artist}",
		ActivityStyle: discord.ActivityStyleGame,
		StatusDisplay: discord.StatusDisplayDetails,
	}
	if presence.lastOptions != want {
		t.Errorf("options sent = %+v, want %+v", presence.lastOptions, want)
	}
}

// TestDiscordServiceLateArtworkReissuesPresence covers the happy path of the
// background cover lookup: the presence goes out with the Plex logo, and is
// re-issued once the real cover arrives.
func TestDiscordServiceLateArtworkReissuesPresence(t *testing.T) {
	resolver := &blockingResolver{url: "https://cdn/cover.jpg", release: make(chan struct{})}
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, resolver)

	app.discord.Publish(playingSession("Song"))

	first, _ := presence.lastPlayback()
	if first.ArtworkURL != "" {
		t.Fatalf("first presence carried artwork %q, want the Plex logo fallback", first.ArtworkURL)
	}

	close(resolver.release)
	waitFor(t, func() bool { return presence.updateCount() == 2 }, "the resolved cover never re-issued the presence")

	second, _ := presence.lastPlayback()
	if second.ArtworkURL != "https://cdn/cover.jpg" {
		t.Errorf("re-issued presence artwork = %q, want the resolved cover", second.ArtworkURL)
	}
}

// TestDiscordServiceDropsStaleArtwork is the generation guard: the track
// changed while its cover was being looked up, so that cover belongs to a track
// that is no longer playing and must not be published over the current one.
func TestDiscordServiceDropsStaleArtwork(t *testing.T) {
	// Only the superseded track's lookup blocks; the current track misses
	// immediately, so any third presence update can only have come from the
	// stale lookup. Keying on the album rather than on call order keeps the
	// test deterministic when the two lookups race.
	resolver := &albumBlockingResolver{
		album:   "First Album",
		url:     "https://cdn/old-cover.jpg",
		release: make(chan struct{}),
	}
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, resolver)

	app.discord.Publish(playingSession("First"))
	app.discord.Publish(playingSession("Second"))

	close(resolver.release)
	time.Sleep(80 * time.Millisecond)

	if presence.updateCount() != 2 {
		t.Fatalf("presence updates = %d, want 2 — the stale cover re-issued a superseded track", presence.updateCount())
	}
	last, _ := presence.lastPlayback()
	if last.Track != "Second" {
		t.Errorf("last presence track = %q, want Second", last.Track)
	}
	if last.ArtworkURL == "https://cdn/old-cover.jpg" {
		t.Error("a cover for the previous track was published against the current one")
	}
}

// albumBlockingResolver blocks the lookup for one specific album until
// released, and reports an immediate miss for every other album.
type albumBlockingResolver struct {
	album   string
	url     string
	release chan struct{}
}

func (r *albumBlockingResolver) Cached(string, string) (string, bool) { return "", false }

func (r *albumBlockingResolver) Resolve(ctx context.Context, _, album string) (string, error) {
	if album != r.album {
		return "", nil
	}
	select {
	case <-r.release:
		return r.url, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// TestDiscordServiceSkipsArtworkWhileManuallyPaused verifies a cover arriving
// after the user paused presence does not un-hide it.
func TestDiscordServiceSkipsArtworkWhileManuallyPaused(t *testing.T) {
	resolver := &blockingResolver{url: "https://cdn/cover.jpg", release: make(chan struct{})}
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, resolver)

	app.discord.Publish(playingSession("Song"))
	app.presence.Toggle() // user pauses presence

	close(resolver.release)
	time.Sleep(80 * time.Millisecond)

	if presence.updateCount() != 1 {
		t.Errorf("presence updates = %d, want 1 — a late cover resurrected a paused presence", presence.updateCount())
	}
}

// TestDiscordServiceNeverLooksUpArtworkWhenDisabled verifies the privacy
// promise: with lookup off, the artist and album never leave the machine.
func TestDiscordServiceNeverLooksUpArtworkWhenDisabled(t *testing.T) {
	resolver := &blockingResolver{url: "https://cdn/cover.jpg", release: make(chan struct{})}
	close(resolver.release) // would answer instantly if it were ever asked

	cfg := config.DefaultConfig()
	disabled := false
	cfg.PresenceArtworkLookup = &disabled

	presence := &recordingPresence{connected: true}
	app := newTestApp(cfg)
	app.discord = app.newDiscordService(presence, resolver)

	app.discord.Publish(playingSession("Song"))
	time.Sleep(50 * time.Millisecond)

	if presence.updateCount() != 1 {
		t.Errorf("presence updates = %d, want 1 — artwork was looked up despite being disabled", presence.updateCount())
	}
}

// TestDiscordServiceClearIsIdempotentWhenDisconnected verifies clearing a
// presence on a closed link is a no-op rather than an error.
func TestDiscordServiceClearIsIdempotentWhenDisconnected(t *testing.T) {
	presence := &recordingPresence{connected: false}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, nil)

	app.discord.Clear()

	if presence.clearCount() != 0 {
		t.Errorf("clear calls = %d, want 0 on a closed link", presence.clearCount())
	}
}
