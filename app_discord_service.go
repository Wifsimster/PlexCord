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

type discordService struct {
	mu       sync.Mutex
	presence DiscordPresence
	artwork  ArtworkResolver

	gen atomic.Uint64

	options         func() discord.Options
	artworkEnabled  func() bool
	defaultClientID func() string
	onConnected     func()
	isPaused        func() bool
}

const artworkResolveTimeout = 6 * time.Second

func (s *discordService) Connect(clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.presence.Connect(clientID)
}

func (s *discordService) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gen.Add(1)
	return s.presence.Disconnect()
}

func (s *discordService) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.presence.IsConnected()
}

func (s *discordService) ClientID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.presence.GetClientID()
}

func (s *discordService) SetPresence(data *discord.PresenceData) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.presence.IsConnected() {
		return false, nil
	}
	return true, s.presence.SetPresence(data)
}

func (s *discordService) ConnectIfDown() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reconnectLocked()
}

func (s *discordService) reconnectLocked() bool {
	if s.presence.IsConnected() {
		return true
	}
	if err := s.presence.Connect(s.defaultClientID()); err != nil {
		return false
	}
	if s.onConnected != nil {
		s.onConnected()
	}
	log.Printf("Discord: Auto-reconnected to Discord")
	return true
}

func (s *discordService) Publish(session *plex.MediaSession) {
	s.mu.Lock()

	gen := s.gen.Add(1)

	artURL := s.cachedArtwork(session)

	if !s.reconnectLocked() {
		s.mu.Unlock()
		return
	}

	if err := s.publishLocked(session, artURL); err != nil {
		log.Printf("Warning: Failed to update Discord presence: %v", err)
	}
	s.mu.Unlock()

	if artURL == "" && s.lookupAllowed() {
		go s.resolveArtwork(session, gen)
	}
}

func (s *discordService) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gen.Add(1)
	if !s.presence.IsConnected() {
		return nil
	}
	return s.presence.ClearPresence()
}

func (s *discordService) lookupAllowed() bool {
	return s.artwork != nil && (s.artworkEnabled == nil || s.artworkEnabled())
}

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

func (s *discordService) cachedArtwork(session *plex.MediaSession) string {
	if !s.lookupAllowed() {
		return ""
	}
	if url, ok := s.artwork.Cached(artworkQuery(session)); ok {
		return url
	}
	return ""
}

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

func yearText(year int) string {
	if year <= 0 {
		return ""
	}
	return strconv.Itoa(year)
}

func (s *discordService) resolveArtwork(session *plex.MediaSession, gen uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), artworkResolveTimeout)
	defer cancel()

	url, err := s.artwork.Resolve(ctx, artworkQuery(session))
	if err != nil || url == "" {
		return
	}
	if s.gen.Load() != gen || (s.isPaused != nil && s.isPaused()) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gen.Load() != gen || !s.presence.IsConnected() {
		return
	}
	if err := s.publishLocked(session, url); err != nil {
		log.Printf("Warning: Failed to update Discord presence with artwork: %v", err)
	}
}
