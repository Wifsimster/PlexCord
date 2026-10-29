package plex

type SessionSource interface {
	GetMusicSessions(userID string) ([]MusicSession, error)
	GetMediaSessions(userID string, mediaTypes []string) ([]MediaSession, error)
}

var _ SessionSource = (*Client)(nil)
