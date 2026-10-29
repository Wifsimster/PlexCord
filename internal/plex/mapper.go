package plex

func filterMusicSessions(
	sessionsResp *SessionsResponse,
	userID string,
	buildThumbURL func(string) string,
) []MusicSession {
	result := make([]MusicSession, 0, len(sessionsResp.Tracks))
	for _, entry := range sessionsResp.Tracks {
		if userID != "" && entry.User.ID != userID {
			continue
		}
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
		if userID != "" && entry.User.ID != userID {
			continue
		}

		thumbURL := ""
		if entry.Thumb != "" && buildThumbURL != nil {
			thumbURL = buildThumbURL(entry.Thumb)
		}
		session := NewMediaSessionFromEntry(entry, thumbURL)

		if !anyType && !wantType[session.MediaType] {
			continue
		}

		session.ApplyFallbacks()
		result = append(result, session)
	}
	return result
}
