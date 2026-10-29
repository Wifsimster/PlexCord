package discord

import "time"

const DefaultClientID = "1463211692656689172"

type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusConnected    ConnectionStatus = "connected"
)

const (
	MediaTypeMusic = "music"
	MediaTypeMovie = "movie"
	MediaTypeTV    = "tv"
	MediaTypePhoto = "photo"
)

const (
	ActivityStyleMedia = "media"
	ActivityStyleGame  = "game"
)

const (
	StatusDisplayApp     = "app"
	StatusDisplayState   = "state"
	StatusDisplayDetails = "details"
)

type PresenceData struct {
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`

	MediaType string `json:"mediaType,omitempty"`

	Track  string `json:"track"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Year   string `json:"year"`
	Player string `json:"player"`

	ShowTitle string `json:"showTitle,omitempty"`
	Season    int    `json:"season,omitempty"`
	Episode   int    `json:"episode,omitempty"`

	ArtworkURL string `json:"artworkUrl"`

	State    string `json:"state"`
	Duration int64  `json:"duration"`
	Position int64  `json:"position"`

	DetailsFormat string `json:"detailsFormat,omitempty"`
	StateFormat   string `json:"stateFormat,omitempty"`

	ActivityStyle string `json:"activityStyle,omitempty"`
	StatusDisplay string `json:"statusDisplay,omitempty"`
}

type ConnectionEvent struct {
	Error     *Error `json:"error,omitempty"`
	ClientID  string `json:"clientId,omitempty"`
	Connected bool   `json:"connected"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PresenceState struct {
	Track    string `json:"track,omitempty"`
	Artist   string `json:"artist,omitempty"`
	Album    string `json:"album,omitempty"`
	State    string `json:"state,omitempty"`
	Duration int64  `json:"duration,omitempty"`
	Position int64  `json:"position,omitempty"`
	Active   bool   `json:"active"`
}
