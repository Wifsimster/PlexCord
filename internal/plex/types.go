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
)

// Server represents a discovered Plex Media Server
type Server struct {
	ID      string `json:"id"`       // Unique resource identifier
	Name    string `json:"name"`     // Server display name
	Address string `json:"address"`  // IP address
	Port    string `json:"port"`     // Port (typically 32400)
	IsLocal bool   `json:"isLocal"`  // True if on local network
	Version string `json:"version"`  // Server version (optional)
}

// URL returns the full server URL
func (s *Server) URL() string {
	return fmt.Sprintf("http://%s:%s", s.Address, s.Port)
}

// ValidationResult represents the outcome of validating a Plex connection
type ValidationResult struct {
	Success           bool   `json:"success"`           // Whether validation passed
	ServerName        string `json:"serverName"`        // Plex server friendly name
	ServerVersion     string `json:"serverVersion"`     // Plex Media Server version
	LibraryCount      int    `json:"libraryCount"`      // Number of media libraries
	MachineIdentifier string `json:"machineIdentifier"` // Unique server ID
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
	Size     int            `xml:"size,attr"`
	Accounts []AccountEntry `xml:"Account"`
}

// AccountEntry represents a single account in the accounts response
type AccountEntry struct {
	ID    string `xml:"id,attr"`
	Name  string `xml:"name,attr"`
	Thumb string `xml:"thumb,attr"`
}

// SessionsResponse represents the XML response from /status/sessions endpoint
type SessionsResponse struct {
	XMLName  xml.Name       `xml:"MediaContainer"`
	Size     int            `xml:"size,attr"`
	Sessions []SessionEntry `xml:"Track"`
}

// SessionEntry represents a single session entry from the sessions response
// This captures both music (Track) and other media types
type SessionEntry struct {
	// Core session identifiers
	SessionKey string `xml:"sessionKey,attr"`
	Key        string `xml:"key,attr"`
	Type       string `xml:"type,attr"` // "track", "episode", "movie", "photo"

	// Track metadata (for music sessions)
	Title            string `xml:"title,attr"`            // Track title
	GrandparentTitle string `xml:"grandparentTitle,attr"` // Artist name
	ParentTitle      string `xml:"parentTitle,attr"`      // Album name
	Thumb            string `xml:"thumb,attr"`            // Album art URL
	Duration         int64  `xml:"duration,attr"`         // Duration in milliseconds
	ViewOffset       int64  `xml:"viewOffset,attr"`       // Current position in milliseconds

	// Nested elements
	User   SessionUser   `xml:"User"`
	Player SessionPlayer `xml:"Player"`
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
	Session                  // Embedded session info
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
