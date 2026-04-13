package discord

import (
	"testing"

	"github.com/hugolgst/rich-go/client"
)

func TestMusicBuilder_DefaultFormat(t *testing.T) {
	data := &PresenceData{
		MediaType: MediaTypeMusic,
		Track:     "Bohemian Rhapsody",
		Artist:    "Queen",
		Album:     "A Night at the Opera",
		State:     "playing",
	}

	activity := (musicBuilder{}).Build(data)

	if activity.Details != "Bohemian Rhapsody" {
		t.Errorf("expected Details 'Bohemian Rhapsody', got %q", activity.Details)
	}
	if activity.State != "by Queen • A Night at the Opera" {
		t.Errorf("unexpected State: %q", activity.State)
	}
}

func TestMusicBuilder_CustomFormat(t *testing.T) {
	data := &PresenceData{
		MediaType:     MediaTypeMusic,
		Track:         "Song",
		Artist:        "Artist",
		Album:         "Album",
		Year:          "2024",
		State:         "playing",
		DetailsFormat: "{track} ({year})",
		StateFormat:   "{artist} - {album}",
	}

	activity := (musicBuilder{}).Build(data)

	if activity.Details != "Song (2024)" {
		t.Errorf("unexpected Details: %q", activity.Details)
	}
	if activity.State != "Artist - Album" {
		t.Errorf("unexpected State: %q", activity.State)
	}
}

func TestMovieBuilder_WithYear(t *testing.T) {
	data := &PresenceData{
		MediaType: MediaTypeMovie,
		Track:     "Inception",
		Year:      "2010",
		State:     "playing",
	}

	activity := (movieBuilder{}).Build(data)

	if activity.Details != "Inception" {
		t.Errorf("expected Details 'Inception', got %q", activity.Details)
	}
	if activity.State != "Movie • 2010" {
		t.Errorf("expected 'Movie • 2010', got %q", activity.State)
	}
}

func TestTVBuilder_WithSeasonAndEpisode(t *testing.T) {
	data := &PresenceData{
		MediaType: MediaTypeTV,
		Track:     "The Rains of Castamere",
		ShowTitle: "Game of Thrones",
		Season:    3,
		Episode:   9,
		State:     "playing",
	}

	activity := (tvBuilder{}).Build(data)

	if activity.Details != "The Rains of Castamere" {
		t.Errorf("expected episode title as Details, got %q", activity.Details)
	}
	if activity.State != "Game of Thrones • S03E09" {
		t.Errorf("unexpected State: %q", activity.State)
	}
}

func TestTVBuilder_CustomFormatWithTokens(t *testing.T) {
	data := &PresenceData{
		MediaType:     MediaTypeTV,
		Track:         "Episode Title",
		ShowTitle:     "Show",
		Season:        2,
		Episode:       5,
		State:         "playing",
		DetailsFormat: "{show}",
		StateFormat:   "S{season}E{episode}: {track}",
	}

	activity := (tvBuilder{}).Build(data)

	if activity.Details != "Show" {
		t.Errorf("unexpected Details: %q", activity.Details)
	}
	if activity.State != "S2E5: Episode Title" {
		t.Errorf("unexpected State: %q", activity.State)
	}
}

func TestBuildActivityForMediaType_DefaultsToMusic(t *testing.T) {
	data := &PresenceData{
		// No MediaType set
		Track:  "Song",
		Artist: "Artist",
		State:  "playing",
	}

	activity := buildActivityForMediaType(data)

	// Should fall through to music builder
	if activity.Details != "Song" {
		t.Errorf("expected music builder to be used, got Details %q", activity.Details)
	}
}

func TestBuildActivityForMediaType_UnknownFallsBackToMusic(t *testing.T) {
	data := &PresenceData{
		MediaType: "audiobook", // unregistered
		Track:     "Chapter 1",
		Artist:    "Author",
		State:     "playing",
	}

	activity := buildActivityForMediaType(data)

	if activity.Details != "Chapter 1" {
		t.Errorf("expected fallback to music builder, got Details %q", activity.Details)
	}
}

func TestBuildActivityForMediaType_DispatchesToMovie(t *testing.T) {
	data := &PresenceData{
		MediaType: MediaTypeMovie,
		Track:     "Film",
		Year:      "2023",
	}

	activity := buildActivityForMediaType(data)

	if activity.State != "Movie • 2023" {
		t.Errorf("expected movie builder, got State %q", activity.State)
	}
}

func TestRegisterPresenceBuilder_Custom(t *testing.T) {
	// Save and restore state
	orig := builderRegistry["custom"]
	defer func() {
		if orig == nil {
			delete(builderRegistry, "custom")
		} else {
			builderRegistry["custom"] = orig
		}
	}()

	RegisterPresenceBuilder("custom", &customBuilder{marker: "CUSTOM"})

	data := &PresenceData{MediaType: "custom"}
	activity := buildActivityForMediaType(data)

	if activity.Details != "CUSTOM" {
		t.Errorf("expected custom builder, got %q", activity.Details)
	}
}

type customBuilder struct {
	marker string
}

func (c *customBuilder) Build(*PresenceData) client.Activity {
	return client.Activity{Details: c.marker}
}
