package plex

import (
	"encoding/xml"
	"fmt"
)

const (
	FallbackTrackTitle = "Unknown Track"
	FallbackArtist     = "Unknown Artist"
	FallbackAlbum      = "Unknown Album"
	FallbackTitle      = "Unknown Title"
	FallbackShowTitle  = "Unknown Show"
)

const (
	MediaTypeMusic = "music"
	MediaTypeMovie = "movie"
	MediaTypeTV    = "tv"
	MediaTypePhoto = "photo"
)

type Server struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    string `json:"port"`
	Version string `json:"version"`
	IsLocal bool   `json:"isLocal"`
}

func (s *Server) URL() string {
	return fmt.Sprintf("http://%s:%s", s.Address, s.Port)
}

type ValidationResult struct {
	ServerName        string `json:"serverName"`
	ServerVersion     string `json:"serverVersion"`
	MachineIdentifier string `json:"machineIdentifier"`
	LibraryCount      int    `json:"libraryCount"`
	Success           bool   `json:"success"`
}

type PlexUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Thumb string `json:"thumb"`
}

type AccountsResponse struct {
	XMLName  xml.Name       `xml:"MediaContainer"`
	Accounts []AccountEntry `xml:"Account"`
	Size     int            `xml:"size,attr"`
}

type AccountEntry struct {
	ID    string `xml:"id,attr"`
	Name  string `xml:"name,attr"`
	Thumb string `xml:"thumb,attr"`
}

type SessionsResponse struct {
	XMLName xml.Name       `xml:"MediaContainer"`
	Tracks  []SessionEntry `xml:"Track"`
	Videos  []SessionEntry `xml:"Video"`
	Photos  []SessionEntry `xml:"Photo"`
	Size    int            `xml:"size,attr"`
}

func (sr *SessionsResponse) AllEntries() []SessionEntry {
	total := len(sr.Tracks) + len(sr.Videos) + len(sr.Photos)
	entries := make([]SessionEntry, 0, total)
	entries = append(entries, sr.Tracks...)
	entries = append(entries, sr.Videos...)
	entries = append(entries, sr.Photos...)
	return entries
}

type SessionEntry struct {
	User   SessionUser   `xml:"User"`
	Player SessionPlayer `xml:"Player"`

	SessionKey string `xml:"sessionKey,attr"`
	Key        string `xml:"key,attr"`
	Type       string `xml:"type,attr"`

	Title            string `xml:"title,attr"`
	GrandparentTitle string `xml:"grandparentTitle,attr"`
	ParentTitle      string `xml:"parentTitle,attr"`
	Thumb            string `xml:"thumb,attr"`
	Duration         int64  `xml:"duration,attr"`
	ViewOffset       int64  `xml:"viewOffset,attr"`

	Year        int `xml:"year,attr"`
	ParentIndex int `xml:"parentIndex,attr"`
	Index       int `xml:"index,attr"`
}

type SessionUser struct {
	ID    string `xml:"id,attr"`
	Title string `xml:"title,attr"`
	Thumb string `xml:"thumb,attr"`
}

type SessionPlayer struct {
	State   string `xml:"state,attr"`
	Title   string `xml:"title,attr"`
	Product string `xml:"product,attr"`
}

type Session struct {
	SessionKey string `json:"sessionKey"`
	UserID     string `json:"userId"`
	UserName   string `json:"userName"`
	Type       string `json:"type"`
	State      string `json:"state"`
	PlayerName string `json:"playerName"`
}

type MusicSession struct {
	Session
	Track      string `json:"track"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	Thumb      string `json:"thumb"`
	ThumbURL   string `json:"thumbUrl"`
	Duration   int64  `json:"duration"`
	ViewOffset int64  `json:"viewOffset"`
}

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
}

type MediaSession struct {
	SessionKey string `json:"sessionKey"`
	Type       string `json:"type"`
	MediaType  string `json:"mediaType"`
	State      string `json:"state"`

	Title      string `json:"title"`
	Thumb      string `json:"thumb"`
	ThumbURL   string `json:"thumbUrl"`
	Duration   int64  `json:"duration"`
	ViewOffset int64  `json:"viewOffset"`
	Year       int    `json:"year"`

	Artist string `json:"artist"`
	Album  string `json:"album"`

	ShowTitle string `json:"showTitle"`
	Season    int    `json:"season"`
	Episode   int    `json:"episode"`

	UserID     string `json:"userId"`
	UserName   string `json:"userName"`
	PlayerName string `json:"playerName"`
}

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
