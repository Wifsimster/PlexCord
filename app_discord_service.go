package main

import (
	"context"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"plexcord/internal/artwork"
	"plexcord/internal/discord"
	"plexcord/internal/plex"
)

// discordService owns the Discord link and everything that has to be
// serialized against it: the connection itself, the presence currently
// showing, and the artwork generation counter that keeps a late cover lookup
// from resurrecting a presence the user has moved on from.
//
// It is separate from App because all of that changes for Discord's reasons —
// Discord restarting, a cover arriving seconds after the track did, a client ID
// change — and not for the app's (SRP). App keeps what is genuinely its own:
// persisting the client ID, recording connection times, and emitting frontend
// events.
//
// One mutex guards the whole service. Connection state and presence publishing
// are the same resource; splitting the lock between App and a helper would have
// left two locks over one Discord socket.
type discordService struct {
	mu       sync.Mutex
	presence DiscordPresence
	artwork  ArtworkResolver

	// gen debounces async artwork re-issues: each session change bumps it, and
	// a late resolve only re-issues presence if its generation is still current.
	gen atomic.Uint64

	// options reports the user's presence display preferences.
	options func() discord.Options
	// artworkEnabled reports whether public artwork lookup is allowed. When it
	// is off, PlexCord never sends external services the artist/album names.
	artworkEnabled func() bool
	// defaultClientID is the client ID a silent reconnect should use.
	defaultClientID func() string
	// onConnected records a successful connection (the connection history).
	onConnected func()
	// isPaused reports the manual pause state, so a late artwork resolve does
	// not un-hide a presence the user paused in the meantime.
	isPaused func() bool
}

// artworkResolveTimeout bounds a background cover lookup. It is generous
// because the lookup is off the presence path: the presence already went out
// with the Plex logo, and this only upgrades it.
const artworkResolveTimeout = 6 * time.Second

// Connect opens the Discord link with the given client ID.
func (s *discordService) Connect(clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.presence.Connect(clientID)
}

// Disconnect closes the Discord link.
func (s *discordService) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.presence.Disconnect()
}

// IsConnected reports whether the Discord link is up.
func (s *discordService) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.presence.IsConnected()
}

// ClientID returns the client ID the link was opened with.
func (s *discordService) ClientID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.presence.GetClientID()
}

// SetPresence publishes pre-built presence data. Returns false when there is no
// connection to publish on.
func (s *discordService) SetPresence(data *discord.PresenceData) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.presence.IsConnected() {
		return false, nil
	}
	return true, s.presence.SetPresence(data)
}

// ConnectIfDown opens the link when it is down, using the default client ID,
// and reports whether it is up afterwards. Silent: no events are emitted, since
// Discord simply not running is the expected case, not a failure to report.
func (s *discordService) ConnectIfDown() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reconnectLocked()
}

// reconnectLocked is ConnectIfDown's body. Caller holds mu.
func (s *discordService) reconnectLocked() bool {
	if s.presence.IsConnected() {
		return true
	}
	if err := s.presence.Connect(s.defaultClientID()); err != nil {
		// Discord is probably not running; not an error worth logging on every
		// poll.
		return false
	}
	if s.onConnected != nil {
		s.onConnected()
	}
	log.Printf("Discord: Auto-reconnected to Discord")
	return true
}

// Publish sends the session to Discord, reconnecting first if the link dropped
// (Discord restarting mid-playback is the common case).
//
// Artwork is resolved on the fast path only from cache, so a known album shows
// instantly; an unknown one goes out with the Plex logo and the real cover is
// looked up in the background and re-issued when it lands.
func (s *discordService) Publish(session *plex.MediaSession) {
	s.mu.Lock()

	// Each session update supersedes any in-flight async artwork resolve.
	gen := s.gen.Add(1)

	// Never send Discord the tokened Plex ThumbURL (credential leak): resolve a
	// public URL instead, or fall back to the Plex logo asset.
	artURL := s.cachedArtwork(session)

	if !s.reconnectLocked() {
		s.mu.Unlock()
		return
	}

	if err := s.publishLocked(session, artURL); err != nil {
		log.Printf("Warning: Failed to update Discord presence: %v", err)
	}
	s.mu.Unlock()

	// If we have no cover yet, resolve one off the presence path and re-issue
	// when it lands (dropped if the session has since changed).
	if artURL == "" && s.lookupAllowed() {
		go s.resolveArtwork(session, gen)
	}
}

// Clear removes the presence currently showing. A closed link is not an error:
// there is nothing to clear.
func (s *discordService) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.presence.IsConnected() {
		return nil
	}
	return s.presence.ClearPresence()
}

// lookupAllowed reports whether a background artwork lookup may run.
func (s *discordService) lookupAllowed() bool {
	return s.artwork != nil && (s.artworkEnabled == nil || s.artworkEnabled())
}

// artworkQuery describes what picture this session wants: an album cover, a
// film poster, or the art of the show an episode belongs to. The episode's own
// title is deliberately not searched — no poster database indexes it.
func artworkQuery(session *plex.MediaSession) artwork.Query {
	switch session.MediaType {
	case plex.MediaTypeMovie:
		return artwork.MovieQuery(session.Title, session.Year)
	case plex.MediaTypeTV:
		return artwork.ShowQuery(session.ShowTitle, session.Year)
	default:
		return artwork.MusicQuery(session.Artist, session.Album)
	}
}

// cachedArtwork returns a public artwork URL for the session if one is already
// cached (no network), or "" to use the Plex logo fallback. It never returns
// the tokened Plex ThumbURL.
func (s *discordService) cachedArtwork(session *plex.MediaSession) string {
	if !s.lookupAllowed() {
		return ""
	}
	if url, ok := s.artwork.Cached(artworkQuery(session)); ok {
		return url
	}
	return ""
}

// publishLocked issues a presence update for the session with the given public
// artwork URL. Caller holds mu.
func (s *discordService) publishLocked(session *plex.MediaSession, artURL string) error {
	return s.presence.UpdatePlayback(discord.Playback{
		MediaType:  session.MediaType,
		Track:      session.Title,
		Artist:     session.Artist,
		Album:      session.Album,
		Year:       yearText(session.Year),
		ShowTitle:  session.ShowTitle,
		Season:     session.Season,
		Episode:    session.Episode,
		State:      session.State,
		Duration:   session.Duration,
		Position:   session.ViewOffset,
		ArtworkURL: artURL,
		Player:     session.PlayerName,
	}, s.options())
}

// yearText renders a release year for the presence layer, which carries it as
// text so an unknown year (Plex sends 0) renders as nothing at all rather than
// as the year zero.
func yearText(year int) string {
	if year <= 0 {
		return ""
	}
	return strconv.Itoa(year)
}

// resolveArtwork looks a cover up off the presence path and, if the session is
// still current (generation unchanged) and not paused, re-issues the presence
// with it. Runs in its own goroutine.
func (s *discordService) resolveArtwork(session *plex.MediaSession, gen uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), artworkResolveTimeout)
	defer cancel()

	url, err := s.artwork.Resolve(ctx, artworkQuery(session))
	if err != nil || url == "" {
		return
	}
	// Drop stale resolves (a newer session update superseded this one) and skip
	// while manually paused, so we don't resurrect a hidden presence.
	if s.gen.Load() != gen || (s.isPaused != nil && s.isPaused()) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// Re-check under the lock to avoid racing a concurrent session update.
	if s.gen.Load() != gen || !s.presence.IsConnected() {
		return
	}
	if err := s.publishLocked(session, url); err != nil {
		log.Printf("Warning: Failed to update Discord presence with artwork: %v", err)
	}
}
