package discord

import "time"

type Playback struct {
	MediaType string

	Track  string
	Artist string
	Album  string
	Year   string
	Player string

	ShowTitle string
	Season    int
	Episode   int

	State    string
	Duration int64
	Position int64

	ArtworkURL string
}

type Options struct {
	DetailsFormat string
	StateFormat   string
	ActivityStyle string
	StatusDisplay string
}

func (p Playback) toPresenceData(opts Options) *PresenceData {
	startTime := time.Now().Add(-time.Duration(p.Position) * time.Millisecond)

	data := &PresenceData{
		MediaType:     p.MediaType,
		Track:         p.Track,
		Artist:        p.Artist,
		Album:         p.Album,
		Year:          p.Year,
		Player:        p.Player,
		ShowTitle:     p.ShowTitle,
		Season:        p.Season,
		Episode:       p.Episode,
		State:         p.State,
		Duration:      p.Duration,
		Position:      p.Position,
		StartTime:     &startTime,
		ArtworkURL:    p.ArtworkURL,
		DetailsFormat: opts.DetailsFormat,
		StateFormat:   opts.StateFormat,
		ActivityStyle: opts.ActivityStyle,
		StatusDisplay: opts.StatusDisplay,
	}

	if p.Duration > 0 {
		endTime := startTime.Add(time.Duration(p.Duration) * time.Millisecond)
		data.EndTime = &endTime
	}

	return data
}
