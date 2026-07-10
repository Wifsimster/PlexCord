package discord

import (
	"fmt"
	"strings"

	"plexcord/internal/discord/ipc"
)

// PresenceBuilder builds an ipc.Activity from PresenceData for a single
// media type. Adding a new media type (photo slideshow, audiobook, etc.)
// means implementing this interface and registering it — no changes to
// existing builders (OCP).
type PresenceBuilder interface {
	// Build constructs a Discord Activity for the given presence data.
	// The builder should only use fields relevant to its media type.
	Build(data *PresenceData) ipc.Activity
}

// builderRegistry maps media type strings to their PresenceBuilder.
var builderRegistry = map[string]PresenceBuilder{
	MediaTypeMusic: &musicBuilder{},
	MediaTypeMovie: &movieBuilder{},
	MediaTypeTV:    &tvBuilder{},
}

// RegisterPresenceBuilder registers a builder for a given media type.
// Intended for tests and future extensions.
func RegisterPresenceBuilder(mediaType string, builder PresenceBuilder) {
	builderRegistry[mediaType] = builder
}

// buildActivityForMediaType dispatches to the appropriate PresenceBuilder
// based on data.MediaType. Falls back to the music builder for empty or
// unknown media types to preserve backward compatibility.
func buildActivityForMediaType(data *PresenceData) ipc.Activity {
	mt := data.MediaType
	if mt == "" {
		mt = MediaTypeMusic
	}
	builder, ok := builderRegistry[mt]
	if !ok {
		builder = builderRegistry[MediaTypeMusic]
	}
	return builder.Build(data)
}

// ----------------------------------------------------------------------------
// Common helpers shared by builders
// ----------------------------------------------------------------------------

// applyTimestamps sets the elapsed-time / progress-bar display when playing.
//
// Discord renders a live progress bar when both start and end are present, and
// a plain elapsed timer with only a start. We therefore send:
//   - playing + known duration → start + end (progress bar)
//   - playing + unknown duration (streams) → start only (elapsed timer)
//   - paused → no timestamps (Discord cannot freeze a bar)
func applyTimestamps(activity *ipc.Activity, data *PresenceData) {
	if data.StartTime == nil || data.State != "playing" {
		return
	}
	ts := &ipc.Timestamps{Start: data.StartTime}
	if data.EndTime != nil && data.Duration > 0 {
		ts.End = data.EndTime
	}
	activity.Timestamps = ts
}

// applyActivityType sets the Discord activity type and status-display line.
// base is the media-appropriate type (Listening for music, Watching for video);
// the "game" style overrides it back to classic Playing.
func applyActivityType(activity *ipc.Activity, data *PresenceData, base ipc.ActivityType) {
	if data.ActivityStyle == ActivityStyleGame {
		activity.Type = ipc.ActivityPlaying
		return
	}
	activity.Type = base

	switch data.StatusDisplay {
	case StatusDisplayApp:
		sd := ipc.StatusDisplayName
		activity.StatusDisplayType = &sd
	case StatusDisplayState:
		sd := ipc.StatusDisplayState
		activity.StatusDisplayType = &sd
	case StatusDisplayDetails:
		sd := ipc.StatusDisplayDetails
		activity.StatusDisplayType = &sd
	}
}

// applyPlaybackIcon sets the small image/text based on play state.
func applyPlaybackIcon(activity *ipc.Activity, data *PresenceData) {
	if data.State == "paused" {
		activity.SmallImage = "pause"
		activity.SmallText = "Paused"
	} else {
		activity.SmallImage = "play"
		activity.SmallText = "Playing"
	}
}

// applyArtwork sets the large image to the artwork URL or falls back.
func applyArtwork(activity *ipc.Activity, data *PresenceData, fallbackText string) {
	if data.ArtworkURL != "" {
		activity.LargeImage = data.ArtworkURL
		activity.LargeText = data.Album
		if activity.LargeText == "" {
			activity.LargeText = fallbackText
		}
	} else {
		activity.LargeImage = "plex"
		activity.LargeText = fallbackText
	}
}

// applyFormatTokens applies custom format strings with token replacement.
// Supported tokens: {track}, {artist}, {album}, {year}, {player},
// {show}, {season}, {episode}.
func applyFormatTokens(format string, data *PresenceData) string {
	if format == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"{track}", data.Track,
		"{artist}", data.Artist,
		"{album}", data.Album,
		"{year}", data.Year,
		"{player}", data.Player,
		"{show}", data.ShowTitle,
		"{season}", fmt.Sprintf("%d", data.Season),
		"{episode}", fmt.Sprintf("%d", data.Episode),
	)
	return replacer.Replace(format)
}

// ----------------------------------------------------------------------------
// musicBuilder — the default builder, matches previous buildActivity behavior
// ----------------------------------------------------------------------------

type musicBuilder struct{}

func (musicBuilder) Build(data *PresenceData) ipc.Activity {
	activity := ipc.Activity{}

	if data.DetailsFormat != "" || data.StateFormat != "" {
		activity.Details = applyFormatTokens(data.DetailsFormat, data)
		activity.State = applyFormatTokens(data.StateFormat, data)
	} else {
		activity.Details = data.Track
		if data.Artist != "" {
			if data.Album != "" {
				activity.State = "by " + data.Artist + " • " + data.Album
			} else {
				activity.State = "by " + data.Artist
			}
		}
		if data.Artist == "" && data.State != "" {
			if data.State == "paused" {
				activity.State = "Paused"
			} else {
				activity.State = "Playing on Plex"
			}
		}
	}

	applyActivityType(&activity, data, ipc.ActivityListening)
	applyTimestamps(&activity, data)
	applyArtwork(&activity, data, "Plex Music")
	applyPlaybackIcon(&activity, data)
	return activity
}

// ----------------------------------------------------------------------------
// movieBuilder — formats a movie session
// ----------------------------------------------------------------------------

type movieBuilder struct{}

func (movieBuilder) Build(data *PresenceData) ipc.Activity {
	activity := ipc.Activity{}

	if data.DetailsFormat != "" || data.StateFormat != "" {
		activity.Details = applyFormatTokens(data.DetailsFormat, data)
		activity.State = applyFormatTokens(data.StateFormat, data)
	} else {
		activity.Details = data.Track // Movie title stored in Track field
		if data.Year != "" {
			activity.State = fmt.Sprintf("Movie • %s", data.Year)
		} else {
			activity.State = "Movie"
		}
	}

	applyActivityType(&activity, data, ipc.ActivityWatching)
	applyTimestamps(&activity, data)
	applyArtwork(&activity, data, "Plex")
	applyPlaybackIcon(&activity, data)
	return activity
}

// ----------------------------------------------------------------------------
// tvBuilder — formats a TV episode session
// ----------------------------------------------------------------------------

type tvBuilder struct{}

func (tvBuilder) Build(data *PresenceData) ipc.Activity {
	activity := ipc.Activity{}

	if data.DetailsFormat != "" || data.StateFormat != "" {
		activity.Details = applyFormatTokens(data.DetailsFormat, data)
		activity.State = applyFormatTokens(data.StateFormat, data)
	} else {
		// Episode title as details, show + S/E as state
		activity.Details = data.Track
		switch {
		case data.ShowTitle != "" && data.Season > 0 && data.Episode > 0:
			activity.State = fmt.Sprintf("%s • S%02dE%02d", data.ShowTitle, data.Season, data.Episode)
		case data.ShowTitle != "":
			activity.State = data.ShowTitle
		default:
			activity.State = "TV Episode"
		}
	}

	applyActivityType(&activity, data, ipc.ActivityWatching)
	applyTimestamps(&activity, data)
	applyArtwork(&activity, data, "Plex")
	applyPlaybackIcon(&activity, data)
	return activity
}
