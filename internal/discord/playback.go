package discord

import "time"

// Playback is one snapshot of what the user is playing, as the presence layer
// needs it. It replaces a twelve-parameter positional call: adding or removing
// a field no longer forces every caller to re-count arguments, and a caller
// that only knows a track and an artist can leave the rest zero.
//
// Options carries the user's display preferences, which change independently of
// what is playing — separating them keeps each caller free to supply only the
// half it actually owns (ISP).
type Playback struct {
	// MediaType selects the PresenceBuilder; empty means music.
	MediaType string

	Track  string
	Artist string
	Album  string
	Year   string
	Player string

	// Video/TV fields, ignored for music.
	ShowTitle string
	Season    int
	Episode   int

	// State is "playing" or "paused".
	State string
	// Duration and Position are in milliseconds; Duration 0 means unknown
	// (a live stream), which renders an elapsed timer instead of a progress bar.
	Duration int64
	Position int64

	// ArtworkURL is a publicly reachable cover URL, or "" for the Plex logo.
	// It must never be a tokened Plex URL.
	ArtworkURL string
}

// Options is the user's presence display configuration. Every field may be
// empty, in which case the builder's default applies.
type Options struct {
	// DetailsFormat and StateFormat are token templates ("{track}", "{artist}", …).
	DetailsFormat string
	StateFormat   string
	// ActivityStyle is ActivityStyleMedia or ActivityStyleGame.
	ActivityStyle string
	// StatusDisplay is StatusDisplayApp, StatusDisplayState or StatusDisplayDetails.
	StatusDisplay string
}

// toPresenceData converts the playback snapshot and display options into the
// PresenceData the builders consume, deriving the timestamps Discord needs to
// render elapsed time or a progress bar.
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

	// Compute the end timestamp for the progress bar when the duration is known.
	// Streams / unknown durations fall back to an elapsed-only timer.
	if p.Duration > 0 {
		endTime := startTime.Add(time.Duration(p.Duration) * time.Millisecond)
		data.EndTime = &endTime
	}

	return data
}
