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

// Media type constants used to select a PresenceBuilder.
const (
	MediaTypeMusic = "music"
	MediaTypeMovie = "movie"
	MediaTypeTV    = "tv"
	MediaTypePhoto = "photo"
)

// Activity style constants control the Discord activity type.
//   - ActivityStyleMedia: music → "Listening to", movie/TV → "Watching".
//   - ActivityStyleGame:  everything → classic "Playing" (pre-2024 behavior).
const (
	ActivityStyleMedia = "media"
	ActivityStyleGame  = "game"
)

// Status-display constants control which line Discord surfaces in the member
// list (status_display_type). Empty leaves it to Discord's default.
const (
	StatusDisplayApp     = "app"     // application name (e.g. "PlexCord")
	StatusDisplayState   = "state"   // the state line (e.g. "by Def Leppard")
	StatusDisplayDetails = "details" // the details line (e.g. the track name)
)

// PresenceData represents the information to display in Discord Rich Presence.
// Different builders use different fields depending on MediaType; unused
// fields for a given media type are simply ignored.
type PresenceData struct {
	// Timestamps for elapsed time / progress bar display.
	// StartTime is when playback began (now − position); EndTime is
	// StartTime + duration and is only set when the duration is known, so
	// Discord can render a live progress bar.
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`

	// MediaType selects which PresenceBuilder will format this data.
	// Empty string defaults to MediaTypeMusic for backward compatibility.
	MediaType string `json:"mediaType,omitempty"`

	// Track information (primary fields for music)
	Track  string `json:"track"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Year   string `json:"year"`
	Player string `json:"player"`

	// Video/TV fields (ignored for music)
	ShowTitle string `json:"showTitle,omitempty"`
	Season    int    `json:"season,omitempty"`
	Episode   int    `json:"episode,omitempty"`

	// Artwork URL (for large image)
	ArtworkURL string `json:"artworkUrl"`

	// Playback state
	State    string `json:"state"` // "playing", "paused"
	Duration int64  `json:"duration"`
	Position int64  `json:"position"`

	// Custom format strings for presence display
	DetailsFormat string `json:"detailsFormat,omitempty"`
	StateFormat   string `json:"stateFormat,omitempty"`

	// Presence display options. ActivityStyle selects "media" (Listening/
	// Watching) vs "game" (classic Playing); StatusDisplay selects which line
	// Discord shows in the member list. Empty values fall back to defaults.
	ActivityStyle string `json:"activityStyle,omitempty"`
	StatusDisplay string `json:"statusDisplay,omitempty"`
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
