package plex

// This file contains pure session filtering and mapping functions — the
// domain logic that converts raw SessionEntry values from the Plex API
// into filtered, enriched MusicSession/MediaSession values ready for
// consumption by the poller and Discord presence layer.
//
// These functions are pure (no I/O, no HTTP) and take their dependencies
// as parameters so they can be unit-tested with table-driven tests.

// filterMusicSessions returns only the music sessions from the parsed
// response that belong to the given user. An empty userID matches all
// users. Fallback metadata is applied and artwork URLs are built via
// the provided thumb URL builder.
func filterMusicSessions(
	sessionsResp *SessionsResponse,
	userID string,
	buildThumbURL func(string) string,
) []MusicSession {
	result := make([]MusicSession, 0, len(sessionsResp.Tracks))
	for _, entry := range sessionsResp.Tracks {
		// Filter by user ID if specified
		if userID != "" && entry.User.ID != userID {
			continue
		}
		// Guard against non-track entries appearing under <Track>
		if entry.Type != "track" {
			continue
		}

		thumbURL := ""
		if entry.Thumb != "" && buildThumbURL != nil {
			thumbURL = buildThumbURL(entry.Thumb)
		}

		session := MusicSession{
			Session: Session{
				SessionKey: entry.SessionKey,
				UserID:     entry.User.ID,
				UserName:   entry.User.Title,
				Type:       entry.Type,
				State:      entry.Player.State,
				PlayerName: entry.Player.Title,
			},
			Track:      entry.Title,
			Artist:     entry.GrandparentTitle,
			Album:      entry.ParentTitle,
			Thumb:      entry.Thumb,
			ThumbURL:   thumbURL,
			Duration:   entry.Duration,
			ViewOffset: entry.ViewOffset,
		}

		session.ApplyFallbacks()
		result = append(result, session)
	}
	return result
}

// filterMediaSessions returns MediaSessions matching the requested media
// types (empty = all) for the given user. Applies fallbacks and builds
// artwork URLs.
func filterMediaSessions(
	sessionsResp *SessionsResponse,
	userID string,
	mediaTypes []string,
	buildThumbURL func(string) string,
) []MediaSession {
	wantType := make(map[string]bool, len(mediaTypes))
	for _, t := range mediaTypes {
		wantType[t] = true
	}
	anyType := len(wantType) == 0

	entries := sessionsResp.AllEntries()
	result := make([]MediaSession, 0, len(entries))

	for _, entry := range entries {
		// Filter by user ID if specified
		if userID != "" && entry.User.ID != userID {
			continue
		}

		thumbURL := ""
		if entry.Thumb != "" && buildThumbURL != nil {
			thumbURL = buildThumbURL(entry.Thumb)
		}
		session := NewMediaSessionFromEntry(entry, thumbURL)

		// Filter by requested media types
		if !anyType && !wantType[session.MediaType] {
			continue
		}

		session.ApplyFallbacks()
		result = append(result, session)
	}
	return result
}
