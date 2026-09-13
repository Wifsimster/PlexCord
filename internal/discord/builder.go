package discord

import (
	"fmt"
	"strings"
	"sync"

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

// BuilderRegistry resolves a PresenceBuilder for a media type. It is the
// extension point for new media types: register a builder and every presence
// path picks it up, with no change to the existing builders or to the dispatch
// (OCP).
//
// Registration is guarded by a mutex because a presence update can be built on
// the poller goroutine while another goroutine registers a builder.
type BuilderRegistry struct {
	mu       sync.RWMutex
	builders map[string]PresenceBuilder
	// fallback formats media types nothing is registered for, so an unknown
	// type degrades to a sensible presence rather than none at all.
	fallback PresenceBuilder
}

// NewBuilderRegistry returns a registry preloaded with the built-in builders.
func NewBuilderRegistry() *BuilderRegistry {
	music := &musicBuilder{}
	return &BuilderRegistry{
		builders: map[string]PresenceBuilder{
			MediaTypeMusic: music,
			MediaTypeMovie: &movieBuilder{},
			MediaTypeTV:    &tvBuilder{},
		},
		fallback: music,
	}
}

// Register adds or replaces the builder for a media type.
func (r *BuilderRegistry) Register(mediaType string, builder PresenceBuilder) {
	if builder == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builders[mediaType] = builder
}

// Builder returns the builder for a media type, falling back to the music
// builder for an empty or unregistered type (backward compatibility).
func (r *BuilderRegistry) Builder(mediaType string) PresenceBuilder {
	if mediaType == "" {
		mediaType = MediaTypeMusic
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if b, ok := r.builders[mediaType]; ok {
		return b
	}
	return r.fallback
}

// Build formats the presence data with the builder for its media type.
func (r *BuilderRegistry) Build(data *PresenceData) ipc.Activity {
	return r.Builder(data.MediaType).Build(data)
}

// defaultRegistry backs the package-level helpers used by PresenceManager.
var defaultRegistry = NewBuilderRegistry()

// RegisterPresenceBuilder registers a builder for a given media type on the
// default registry. Intended for tests and future extensions.
func RegisterPresenceBuilder(mediaType string, builder PresenceBuilder) {
	defaultRegistry.Register(mediaType, builder)
}

// buildActivityForMediaType dispatches to the appropriate PresenceBuilder
// based on data.MediaType, via the default registry.
func buildActivityForMediaType(data *PresenceData) ipc.Activity {
	return defaultRegistry.Build(data)
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
