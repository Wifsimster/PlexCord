package discord

import (
	"fmt"
	"strings"

	"github.com/hugolgst/rich-go/client"
)

// PresenceBuilder builds a rich-go Activity from PresenceData for a single
// media type. Adding a new media type (photo slideshow, audiobook, etc.)
// means implementing this interface and registering it — no changes to
// existing builders (OCP).
type PresenceBuilder interface {
	// Build constructs a Discord Activity for the given presence data.
	// The builder should only use fields relevant to its media type.
	Build(data *PresenceData) client.Activity
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
func buildActivityForMediaType(data *PresenceData) client.Activity {
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

// applyTimestamps sets the elapsed-time display when playing.
func applyTimestamps(activity *client.Activity, data *PresenceData) {
	if data.StartTime != nil && data.State == "playing" {
		activity.Timestamps = &client.Timestamps{Start: data.StartTime}
	}
}

// applyPlaybackIcon sets the small image/text based on play state.
func applyPlaybackIcon(activity *client.Activity, data *PresenceData) {
	if data.State == "paused" {
		activity.SmallImage = "pause"
		activity.SmallText = "Paused"
	} else {
		activity.SmallImage = "play"
		activity.SmallText = "Playing"
	}
}

// applyArtwork sets the large image to the artwork URL or falls back.
func applyArtwork(activity *client.Activity, data *PresenceData, fallbackText string) {
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

func (musicBuilder) Build(data *PresenceData) client.Activity {
	activity := client.Activity{}

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

	applyTimestamps(&activity, data)
	applyArtwork(&activity, data, "Plex Music")
	applyPlaybackIcon(&activity, data)
	return activity
}

// ----------------------------------------------------------------------------
// movieBuilder — formats a movie session
// ----------------------------------------------------------------------------

type movieBuilder struct{}

func (movieBuilder) Build(data *PresenceData) client.Activity {
	activity := client.Activity{}

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

	applyTimestamps(&activity, data)
	applyArtwork(&activity, data, "Plex")
	applyPlaybackIcon(&activity, data)
	return activity
}

// ----------------------------------------------------------------------------
// tvBuilder — formats a TV episode session
// ----------------------------------------------------------------------------

type tvBuilder struct{}

func (tvBuilder) Build(data *PresenceData) client.Activity {
	activity := client.Activity{}

	if data.DetailsFormat != "" || data.StateFormat != "" {
		activity.Details = applyFormatTokens(data.DetailsFormat, data)
		activity.State = applyFormatTokens(data.StateFormat, data)
	} else {
		// Episode title as details, show + S/E as state
		activity.Details = data.Track
		if data.ShowTitle != "" && data.Season > 0 && data.Episode > 0 {
			activity.State = fmt.Sprintf("%s • S%02dE%02d", data.ShowTitle, data.Season, data.Episode)
		} else if data.ShowTitle != "" {
			activity.State = data.ShowTitle
		} else {
			activity.State = "TV Episode"
		}
	}

	applyTimestamps(&activity, data)
	applyArtwork(&activity, data, "Plex")
	applyPlaybackIcon(&activity, data)
	return activity
}
