package discord

import (
	"testing"
	"time"

	"plexcord/internal/discord/ipc"
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

func (c *customBuilder) Build(*PresenceData) ipc.Activity {
	return ipc.Activity{Details: c.marker}
}

// ----------------------------------------------------------------------------
// Activity type & status-display (US-002)
// ----------------------------------------------------------------------------

func TestActivityType_PerMediaType(t *testing.T) {
	tests := []struct {
		name      string
		mediaType string
		want      ipc.ActivityType
	}{
		{"music is Listening", MediaTypeMusic, ipc.ActivityListening},
		{"movie is Watching", MediaTypeMovie, ipc.ActivityWatching},
		{"tv is Watching", MediaTypeTV, ipc.ActivityWatching},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := buildActivityForMediaType(&PresenceData{
				MediaType: tt.mediaType,
				Track:     "x",
				State:     "playing",
			})
			if activity.Type != tt.want {
				t.Errorf("Type = %d, want %d", activity.Type, tt.want)
			}
		})
	}
}

func TestActivityType_GameStyleForcesPlaying(t *testing.T) {
	for _, mt := range []string{MediaTypeMusic, MediaTypeMovie, MediaTypeTV} {
		activity := buildActivityForMediaType(&PresenceData{
			MediaType:     mt,
			Track:         "x",
			State:         "playing",
			ActivityStyle: ActivityStyleGame,
		})
		if activity.Type != ipc.ActivityPlaying {
			t.Errorf("mediaType %q with game style: Type = %d, want 0 (Playing)", mt, activity.Type)
		}
		if activity.StatusDisplayType != nil {
			t.Errorf("mediaType %q with game style should not set status_display_type", mt)
		}
	}
}

func TestStatusDisplay_MapsToType(t *testing.T) {
	tests := []struct {
		display string
		want    ipc.StatusDisplayType
	}{
		{StatusDisplayApp, ipc.StatusDisplayName},
		{StatusDisplayState, ipc.StatusDisplayState},
		{StatusDisplayDetails, ipc.StatusDisplayDetails},
	}
	for _, tt := range tests {
		t.Run(tt.display, func(t *testing.T) {
			activity := (musicBuilder{}).Build(&PresenceData{
				Track:         "x",
				State:         "playing",
				StatusDisplay: tt.display,
			})
			if activity.StatusDisplayType == nil {
				t.Fatalf("StatusDisplayType is nil for %q", tt.display)
			}
			if *activity.StatusDisplayType != tt.want {
				t.Errorf("StatusDisplayType = %d, want %d", *activity.StatusDisplayType, tt.want)
			}
		})
	}
}

func TestStatusDisplay_EmptyLeavesNil(t *testing.T) {
	activity := (musicBuilder{}).Build(&PresenceData{Track: "x", State: "playing"})
	if activity.StatusDisplayType != nil {
		t.Error("empty StatusDisplay should leave StatusDisplayType nil (Discord default)")
	}
}

// ----------------------------------------------------------------------------
// Progress-bar timestamps (US-003)
// ----------------------------------------------------------------------------

func TestApplyTimestamps_PlayingWithDurationSendsStartAndEnd(t *testing.T) {
	start := time.Unix(1000, 0)
	end := start.Add(240 * time.Second)
	activity := (musicBuilder{}).Build(&PresenceData{
		Track:     "x",
		State:     "playing",
		Duration:  240_000,
		StartTime: &start,
		EndTime:   &end,
	})
	if activity.Timestamps == nil || activity.Timestamps.Start == nil {
		t.Fatal("expected start timestamp")
	}
	if activity.Timestamps.End == nil {
		t.Error("expected end timestamp for a known duration (progress bar)")
	}
}

func TestApplyTimestamps_ZeroDurationSendsStartOnly(t *testing.T) {
	start := time.Unix(1000, 0)
	activity := (musicBuilder{}).Build(&PresenceData{
		Track:     "x",
		State:     "playing",
		Duration:  0,
		StartTime: &start,
		// EndTime intentionally nil, mirroring UpdatePresenceFromPlayback.
	})
	if activity.Timestamps == nil || activity.Timestamps.Start == nil {
		t.Fatal("expected start timestamp")
	}
	if activity.Timestamps.End != nil {
		t.Error("duration == 0 should send start-only timestamps")
	}
}

func TestApplyTimestamps_PausedSendsNone(t *testing.T) {
	start := time.Unix(1000, 0)
	end := start.Add(240 * time.Second)
	activity := (musicBuilder{}).Build(&PresenceData{
		Track:     "x",
		State:     "paused",
		Duration:  240_000,
		StartTime: &start,
		EndTime:   &end,
	})
	if activity.Timestamps != nil {
		t.Error("paused sessions should send no timestamps")
	}
}
