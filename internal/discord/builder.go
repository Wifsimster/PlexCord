package discord

import (
	"fmt"
	"strings"
	"sync"

	"plexcord/internal/discord/ipc"
)

type PresenceBuilder interface {
	Build(data *PresenceData) ipc.Activity
}

type BuilderRegistry struct {
	mu       sync.RWMutex
	builders map[string]PresenceBuilder
	fallback PresenceBuilder
}

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

func (r *BuilderRegistry) Register(mediaType string, builder PresenceBuilder) {
	if builder == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builders[mediaType] = builder
}

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

func (r *BuilderRegistry) Build(data *PresenceData) ipc.Activity {
	return r.Builder(data.MediaType).Build(data)
}

var defaultRegistry = NewBuilderRegistry()

func RegisterPresenceBuilder(mediaType string, builder PresenceBuilder) {
	defaultRegistry.Register(mediaType, builder)
}

func buildActivityForMediaType(data *PresenceData) ipc.Activity {
	return defaultRegistry.Build(data)
}

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

func applyPlaybackIcon(activity *ipc.Activity, data *PresenceData) {
	if data.State == "paused" {
		activity.SmallImage = "pause"
		activity.SmallText = "Paused"
	} else {
		activity.SmallImage = "play"
		activity.SmallText = "Playing"
	}
}

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

func orPlex(text string) string {
	if text == "" {
		return "Plex"
	}
	return text
}

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

type movieBuilder struct{}

func (movieBuilder) Build(data *PresenceData) ipc.Activity {
	activity := ipc.Activity{}

	if data.DetailsFormat != "" || data.StateFormat != "" {
		activity.Details = applyFormatTokens(data.DetailsFormat, data)
		activity.State = applyFormatTokens(data.StateFormat, data)
	} else {
		activity.Details = data.Track
		if data.Year != "" {
			activity.State = fmt.Sprintf("Movie • %s", data.Year)
		} else {
			activity.State = "Movie"
		}
	}

	applyActivityType(&activity, data, ipc.ActivityWatching)
	applyTimestamps(&activity, data)
	applyArtwork(&activity, data, orPlex(data.Track))
	applyPlaybackIcon(&activity, data)
	return activity
}

type tvBuilder struct{}

func (tvBuilder) Build(data *PresenceData) ipc.Activity {
	activity := ipc.Activity{}

	if data.DetailsFormat != "" || data.StateFormat != "" {
		activity.Details = applyFormatTokens(data.DetailsFormat, data)
		activity.State = applyFormatTokens(data.StateFormat, data)
	} else {
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
	applyArtwork(&activity, data, orPlex(data.ShowTitle))
	applyPlaybackIcon(&activity, data)
	return activity
}
