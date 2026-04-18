package plex

import (
	"encoding/xml"
	"fmt"
)

// Fallback constants for missing metadata (AC1, AC2, AC3, AC7)
const (
	FallbackTrackTitle = "Unknown Track"
	FallbackArtist     = "Unknown Artist"
	FallbackAlbum      = "Unknown Album"
	FallbackTitle      = "Unknown Title"
	FallbackShowTitle  = "Unknown Show"
)

// Media type constants for MediaSession.MediaType
const (
	MediaTypeMusic = "music"
	MediaTypeMovie = "movie"
	MediaTypeTV    = "tv"
	MediaTypePhoto = "photo"
)

// Server represents a discovered Plex Media Server
type Server struct {
	ID      string `json:"id"`      // Unique resource identifier
	Name    string `json:"name"`    // Server display name
	Address string `json:"address"` // IP address
	Port    string `json:"port"`    // Port (typically 32400)
	Version string `json:"version"` // Server version (optional)
	IsLocal bool   `json:"isLocal"` // True if on local network
}

// URL returns the full server URL
func (s *Server) URL() string {
	return fmt.Sprintf("http://%s:%s", s.Address, s.Port)
}

// ValidationResult represents the outcome of validating a Plex connection
type ValidationResult struct {
	ServerName        string `json:"serverName"`        // Plex server friendly name
	ServerVersion     string `json:"serverVersion"`     // Plex Media Server version
	MachineIdentifier string `json:"machineIdentifier"` // Unique server ID
	LibraryCount      int    `json:"libraryCount"`      // Number of media libraries
	Success           bool   `json:"success"`           // Whether validation passed
}

// PlexUser represents a Plex account that can be monitored for playback
type PlexUser struct {
	ID    string `json:"id"`    // Unique user identifier
	Name  string `json:"name"`  // Display name
	Thumb string `json:"thumb"` // Avatar URL (optional)
}

// AccountsResponse represents the XML response from /accounts endpoint
type AccountsResponse struct {
	XMLName  xml.Name       `xml:"MediaContainer"`
	Accounts []AccountEntry `xml:"Account"`
	Size     int            `xml:"size,attr"`
}

// AccountEntry represents a single account in the accounts response
type AccountEntry struct {
	ID    string `xml:"id,attr"`
	Name  string `xml:"name,attr"`
	Thumb string `xml:"thumb,attr"`
}

// SessionsResponse represents the XML response from /status/sessions endpoint.
// Plex returns different XML element types for different media:
//   - <Track> for music
//   - <Video> for movies and TV episodes
//   - <Photo> for photos
type SessionsResponse struct {
	XMLName xml.Name       `xml:"MediaContainer"`
	Tracks  []SessionEntry `xml:"Track"`
	Videos  []SessionEntry `xml:"Video"`
	Photos  []SessionEntry `xml:"Photo"`
	Size    int            `xml:"size,attr"`
}

// AllEntries returns all session entries merged from Tracks, Videos, and Photos.
func (sr *SessionsResponse) AllEntries() []SessionEntry {
	total := len(sr.Tracks) + len(sr.Videos) + len(sr.Photos)
	entries := make([]SessionEntry, 0, total)
	entries = append(entries, sr.Tracks...)
	entries = append(entries, sr.Videos...)
	entries = append(entries, sr.Photos...)
	return entries
}

// SessionEntry represents a single session entry from the sessions response.
// This captures music (Track), video (Video), and photo (Photo) media types.
// Field semantics vary by type:
//   - Track:   GrandparentTitle=Artist, ParentTitle=Album, Title=TrackName
//   - Episode: GrandparentTitle=ShowName, ParentTitle=SeasonName, Title=EpisodeName
//   - Movie:   Title=MovieName, Year=ReleaseYear
//   - Photo:   Title=PhotoName
type SessionEntry struct {
	// Nested elements
	User   SessionUser   `xml:"User"`
	Player SessionPlayer `xml:"Player"`

	// Core session identifiers
	SessionKey string `xml:"sessionKey,attr"`
	Key        string `xml:"key,attr"`
	Type       string `xml:"type,attr"` // "track", "episode", "movie", "photo"

	// Common metadata
	Title            string `xml:"title,attr"`            // Track/episode/movie title
	GrandparentTitle string `xml:"grandparentTitle,attr"` // Artist (music) or Show name (TV)
	ParentTitle      string `xml:"parentTitle,attr"`      // Album (music) or Season name (TV)
	Thumb            string `xml:"thumb,attr"`            // Artwork URL
	Duration         int64  `xml:"duration,attr"`         // Duration in milliseconds
	ViewOffset       int64  `xml:"viewOffset,attr"`       // Current position in milliseconds

	// Video-specific metadata
	Year        int `xml:"year,attr"`        // Release year (movies, episodes)
	ParentIndex int `xml:"parentIndex,attr"` // Season number (TV episodes)
	Index       int `xml:"index,attr"`       // Episode number (TV) or track number (music)
}

// SessionUser represents the user associated with a session
type SessionUser struct {
	ID    string `xml:"id,attr"`
	Title string `xml:"title,attr"` // Username
	Thumb string `xml:"thumb,attr"` // Avatar URL
}

// SessionPlayer represents the player/client for a session
type SessionPlayer struct {
	State   string `xml:"state,attr"`   // "playing", "paused", "stopped"
	Title   string `xml:"title,attr"`   // Player name (e.g., "Chrome")
	Product string `xml:"product,attr"` // Product name (e.g., "Plex Web")
}

// Session represents a parsed Plex playback session
type Session struct {
	SessionKey string `json:"sessionKey"` // Unique session identifier
	UserID     string `json:"userId"`     // User ID for this session
	UserName   string `json:"userName"`   // User display name
	Type       string `json:"type"`       // Media type: "track", "episode", "movie", "photo"
	State      string `json:"state"`      // Playback state: "playing", "paused", "stopped"
	PlayerName string `json:"playerName"` // Player/client name
}

// MusicSession represents a music playback session with track metadata
type MusicSession struct {
	Session           // Embedded session info
	Track      string `json:"track"`      // Track title
	Artist     string `json:"artist"`     // Artist name
	Album      string `json:"album"`      // Album name
	Thumb      string `json:"thumb"`      // Relative album artwork path from Plex
	ThumbURL   string `json:"thumbUrl"`   // Absolute album artwork URL (includes server URL and token)
	Duration   int64  `json:"duration"`   // Track duration in milliseconds
	ViewOffset int64  `json:"viewOffset"` // Current playback position in milliseconds
}

// ApplyFallbacks replaces empty metadata fields with appropriate fallback values.
// This ensures the UI always has displayable content even with incomplete metadata.
func (m *MusicSession) ApplyFallbacks() {
	if m.Track == "" {
		m.Track = FallbackTrackTitle
	}
	if m.Artist == "" {
		m.Artist = FallbackArtist
	}
	if m.Album == "" {
		m.Album = FallbackAlbum
	}
	// Duration and ViewOffset default to 0 (Go zero value) - no fallback needed
	// Thumb/ThumbURL empty string is acceptable - no fallback needed (AC4)
}

// MediaSession represents a generic media playback session that works for all media types
// (music, movies, TV episodes, photos). This is the unified session type for multi-media support.
type MediaSession struct {
	SessionKey string `json:"sessionKey"`
	Type       string `json:"type"`      // Plex type: "track", "movie", "episode", "photo"
	MediaType  string `json:"mediaType"` // Simplified: "music", "movie", "tv", "photo"
	State      string `json:"state"`     // "playing", "paused", "stopped"

	// Common metadata
	Title      string `json:"title"`      // Track name, movie title, or episode title
	Thumb      string `json:"thumb"`      // Relative artwork path from Plex
	ThumbURL   string `json:"thumbUrl"`   // Absolute artwork URL (includes server URL and token)
	Duration   int64  `json:"duration"`   // Duration in milliseconds
	ViewOffset int64  `json:"viewOffset"` // Current playback position in milliseconds
	Year       int    `json:"year"`       // Release year

	// Music-specific
	Artist string `json:"artist"` // Artist name (music only)
	Album  string `json:"album"`  // Album name (music only)

	// TV-specific
	ShowTitle string `json:"showTitle"` // Show name (TV episodes only)
	Season    int    `json:"season"`    // Season number (TV episodes only)
	Episode   int    `json:"episode"`   // Episode number (TV episodes only)

	// Session context
	UserID     string `json:"userId"`
	UserName   string `json:"userName"`
	PlayerName string `json:"playerName"`
}

// ApplyFallbacks replaces empty metadata fields with appropriate fallback values
// based on the media type.
func (m *MediaSession) ApplyFallbacks() {
	switch m.MediaType {
	case MediaTypeMusic:
		if m.Title == "" {
			m.Title = FallbackTrackTitle
		}
		if m.Artist == "" {
			m.Artist = FallbackArtist
		}
		if m.Album == "" {
			m.Album = FallbackAlbum
		}
	case MediaTypeTV:
		if m.Title == "" {
			m.Title = FallbackTitle
		}
		if m.ShowTitle == "" {
			m.ShowTitle = FallbackShowTitle
		}
	case MediaTypeMovie, MediaTypePhoto:
		if m.Title == "" {
			m.Title = FallbackTitle
		}
	}
}

// mediaTypeFromPlexType converts a Plex type string to a simplified media type.
func mediaTypeFromPlexType(plexType string) string {
	switch plexType {
	case "track":
		return MediaTypeMusic
	case "movie":
		return MediaTypeMovie
	case "episode":
		return MediaTypeTV
	case "photo":
		return MediaTypePhoto
	default:
		return plexType
	}
}

// NewMediaSessionFromEntry creates a MediaSession from a SessionEntry and absolute thumb URL.
func NewMediaSessionFromEntry(entry SessionEntry, thumbURL string) MediaSession {
	mediaType := mediaTypeFromPlexType(entry.Type)

	ms := MediaSession{
		SessionKey: entry.SessionKey,
		Type:       entry.Type,
		MediaType:  mediaType,
		State:      entry.Player.State,
		Title:      entry.Title,
		Thumb:      entry.Thumb,
		ThumbURL:   thumbURL,
		Duration:   entry.Duration,
		ViewOffset: entry.ViewOffset,
		Year:       entry.Year,
		UserID:     entry.User.ID,
		UserName:   entry.User.Title,
		PlayerName: entry.Player.Title,
	}

	// Populate type-specific fields based on Plex's metadata hierarchy
	switch mediaType {
	case MediaTypeMusic:
		ms.Artist = entry.GrandparentTitle
		ms.Album = entry.ParentTitle
	case MediaTypeTV:
		ms.ShowTitle = entry.GrandparentTitle
		ms.Season = entry.ParentIndex
		ms.Episode = entry.Index
	}

	return ms
}
