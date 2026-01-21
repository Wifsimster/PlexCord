package discord

import "time"

// DefaultClientID is the official PlexCord Discord Application Client ID.
// Users can override this with their own application ID if desired.
// This is a public identifier, not a secret.
const DefaultClientID = "1463211692656689172"

// ConnectionStatus represents the current Discord connection state
type ConnectionStatus string

const (
	// StatusDisconnected indicates no active Discord connection
	StatusDisconnected ConnectionStatus = "disconnected"
	// StatusConnecting indicates a connection attempt is in progress
	StatusConnecting ConnectionStatus = "connecting"
	// StatusConnected indicates an active Discord connection
	StatusConnected ConnectionStatus = "connected"
)

// PresenceData represents the information to display in Discord Rich Presence
type PresenceData struct {
	// Timestamps for elapsed time display
	StartTime *time.Time `json:"startTime,omitempty"`

	// Track information
	Track  string `json:"track"`
	Artist string `json:"artist"`
	Album  string `json:"album"`

	// Artwork URL (for large image)
	ArtworkURL string `json:"artworkUrl"`

	// Playback state
	State    string `json:"state"` // "playing", "paused"
	Duration int64  `json:"duration"`
	Position int64  `json:"position"`
}

// ConnectionEvent represents a Discord connection state change event
type ConnectionEvent struct {
	Error     *Error `json:"error,omitempty"`
	ClientID  string `json:"clientId,omitempty"`
	Connected bool   `json:"connected"`
}

// Error represents a Discord error for frontend consumption
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PresenceState represents the current presence state for frontend display
type PresenceState struct {
	Track    string `json:"track,omitempty"`
	Artist   string `json:"artist,omitempty"`
	Album    string `json:"album,omitempty"`
	State    string `json:"state,omitempty"` // "playing", "paused"
	Duration int64  `json:"duration,omitempty"`
	Position int64  `json:"position,omitempty"`
	Active   bool   `json:"active"`
}
