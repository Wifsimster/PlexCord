package plex

// SessionSource is the abstraction the Poller depends on to read playback
// sessions. Depending on this interface rather than the concrete *Client
// inverts the dependency (DIP): the poller is driven by whatever can supply
// sessions — the production HTTP client, a fake in tests, or a future
// websocket/event-based source — without the poller changing.
//
// Per Go convention the interface lives beside its consumer rather than its
// implementation, and is kept to exactly the two calls the poller makes (ISP).
type SessionSource interface {
	// GetMusicSessions returns the active music sessions for the user.
	GetMusicSessions(userID string) ([]MusicSession, error)
	// GetMediaSessions returns the active sessions of the given media types
	// for the user. A nil or empty mediaTypes means "all types".
	GetMediaSessions(userID string, mediaTypes []string) ([]MediaSession, error)
}

// Compile-time assertion that the production client satisfies the contract.
var _ SessionSource = (*Client)(nil)
