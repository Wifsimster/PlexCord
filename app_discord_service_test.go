package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"plexcord/internal/artwork"
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

func (r *blockingResolver) Cached(artwork.Query) (string, bool) { return "", false }

func (r *blockingResolver) Resolve(ctx context.Context, _ artwork.Query) (string, error) {
	select {
	case <-r.release:
		return r.url, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func playingSession(track string) *plex.MediaSession {
	// A distinct album per track, so an artwork lookup can be keyed on which
	// track asked for it.
	return &plex.MediaSession{
		MediaType: plex.MediaTypeMusic,
		Title:     track,
		Artist:    "Artist",
		Album:     track + " Album",
		Duration:  1000,
		State:     "playing",
	}
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

// TestDiscordServiceDropsArtworkAfterClear covers playback stopping (or the
// hide-when-paused timer firing) while a cover is being looked up: the cover
// must not republish the presence that was just cleared.
func TestDiscordServiceDropsArtworkAfterClear(t *testing.T) {
	resolver := &blockingResolver{url: "https://cdn/cover.jpg", release: make(chan struct{})}
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, resolver)

	app.discord.Publish(playingSession("Song"))
	if err := app.discord.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	close(resolver.release)
	time.Sleep(80 * time.Millisecond)

	if presence.updateCount() != 1 {
		t.Fatalf("presence updates = %d, want 1 — the late cover resurrected a cleared presence", presence.updateCount())
	}
}

// albumBlockingResolver blocks the lookup for one specific album until
// released, and reports an immediate miss for every other album.
type albumBlockingResolver struct {
	album   string
	url     string
	release chan struct{}
}

func (r *albumBlockingResolver) Cached(artwork.Query) (string, bool) { return "", false }

func (r *albumBlockingResolver) Resolve(ctx context.Context, q artwork.Query) (string, error) {
	if q.Album != r.album {
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

// recordingResolver answers every lookup instantly and remembers what it was
// asked for, so a test can assert which picture a session goes looking for.
type recordingResolver struct {
	url     string
	queries []artwork.Query
}

func (r *recordingResolver) Cached(q artwork.Query) (string, bool) {
	r.queries = append(r.queries, q)
	return r.url, r.url != ""
}

func (r *recordingResolver) Resolve(_ context.Context, q artwork.Query) (string, error) {
	r.queries = append(r.queries, q)
	return r.url, nil
}

// TestDiscordServicePublishesAMovie verifies a film reaches Discord as a film:
// the movie builder's fields, not a track's.
func TestDiscordServicePublishesAMovie(t *testing.T) {
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	resolver := &recordingResolver{url: "https://cdn/poster.jpg"}
	app.discord = app.newDiscordService(presence, resolver)

	app.discord.Publish(&plex.MediaSession{
		MediaType: plex.MediaTypeMovie,
		Title:     "Blade Runner",
		Year:      1982,
		State:     "playing",
		Duration:  7_000_000,
	})

	playback, ok := presence.lastPlayback()
	if !ok {
		t.Fatal("expected a presence update")
	}
	if playback.MediaType != discord.MediaTypeMovie {
		t.Errorf("media type = %q, want %q", playback.MediaType, discord.MediaTypeMovie)
	}
	if playback.Track != "Blade Runner" {
		t.Errorf("title = %q, want the film title", playback.Track)
	}
	if playback.Year != "1982" {
		t.Errorf("year = %q, want %q", playback.Year, "1982")
	}
	if playback.ArtworkURL != "https://cdn/poster.jpg" {
		t.Errorf("artwork = %q, want the resolved poster", playback.ArtworkURL)
	}
	if len(resolver.queries) == 0 || resolver.queries[0].MediaType != artwork.MediaTypeMovie {
		t.Errorf("artwork was looked up as %+v, want a movie query", resolver.queries)
	}
}

// TestDiscordServicePublishesAnEpisode verifies an episode carries its show and
// its season/episode numbers, and that the poster is looked up by show — the
// episode's own title is in no poster database.
func TestDiscordServicePublishesAnEpisode(t *testing.T) {
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	resolver := &recordingResolver{url: "https://cdn/show.jpg"}
	app.discord = app.newDiscordService(presence, resolver)

	app.discord.Publish(&plex.MediaSession{
		MediaType: plex.MediaTypeTV,
		Title:     "Good News",
		ShowTitle: "Severance",
		Season:    1,
		Episode:   2,
		State:     "playing",
		Duration:  3_000_000,
	})

	playback, ok := presence.lastPlayback()
	if !ok {
		t.Fatal("expected a presence update")
	}
	if playback.MediaType != discord.MediaTypeTV {
		t.Errorf("media type = %q, want %q", playback.MediaType, discord.MediaTypeTV)
	}
	if playback.ShowTitle != "Severance" || playback.Season != 1 || playback.Episode != 2 {
		t.Errorf("episode fields = (%q, S%d, E%d), want (Severance, S1, E2)",
			playback.ShowTitle, playback.Season, playback.Episode)
	}
	if len(resolver.queries) == 0 || resolver.queries[0].Title != "Severance" {
		t.Errorf("artwork was looked up as %+v, want the show's art", resolver.queries)
	}
}

// TestDiscordServiceOmitsAnUnknownYear verifies Plex's missing-year zero never
// reaches Discord as the year zero.
func TestDiscordServiceOmitsAnUnknownYear(t *testing.T) {
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, nil)

	app.discord.Publish(&plex.MediaSession{
		MediaType: plex.MediaTypeMovie,
		Title:     "A Film With No Year",
		State:     "playing",
	})

	playback, _ := presence.lastPlayback()
	if playback.Year != "" {
		t.Errorf("year = %q, want it omitted when Plex does not know one", playback.Year)
	}
}
