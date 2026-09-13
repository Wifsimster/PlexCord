package plex

// Change detection for poll results.
//
// The poller must only emit an update when something the user can perceive has
// changed; ViewOffset advances on every poll and must not trigger an emit.
// Rather than one hand-written comparison per session type, each type projects
// itself onto a comparable changeKey and a single generic helper does the
// nil-handling and the comparison. Adding a session type means adding one
// changeKey method, not another comparison ladder (OCP).

// changeKey is the comparable projection of a session: the fields whose change
// is worth telling the rest of the app about. Unused fields stay at their zero
// value for session types that do not have them.
type changeKey struct {
	sessionKey string
	state      string
	title      string
	mediaType  string
	artist     string
	album      string
	showTitle  string
	season     int
	episode    int
}

// changeKey projects a music session. Metadata is included so a Plex metadata
// refresh mid-playback still reaches Discord.
func (m *MusicSession) changeKey() changeKey {
	return changeKey{
		sessionKey: m.SessionKey,
		state:      m.State,
		title:      m.Track,
		artist:     m.Artist,
		album:      m.Album,
	}
}

// changeKey projects a media session across every media type it can carry.
func (m *MediaSession) changeKey() changeKey {
	return changeKey{
		sessionKey: m.SessionKey,
		state:      m.State,
		title:      m.Title,
		mediaType:  m.MediaType,
		artist:     m.Artist,
		album:      m.Album,
		showTitle:  m.ShowTitle,
		season:     m.Season,
		episode:    m.Episode,
	}
}

// nilAwareChanged reports whether prev and curr differ, treating a nil pointer
// as "no session": nil→nil is no change, nil→session (or the reverse) always is.
func nilAwareChanged[T any](prev, curr *T, key func(*T) changeKey) bool {
	switch {
	case prev == nil && curr == nil:
		return false
	case prev == nil || curr == nil:
		return true
	default:
		return key(prev) != key(curr)
	}
}

// sessionChanged reports whether the music session state meaningfully changed.
func sessionChanged(prev, curr *MusicSession) bool {
	return nilAwareChanged(prev, curr, (*MusicSession).changeKey)
}

// mediaSessionChanged reports whether the media session state meaningfully changed.
func mediaSessionChanged(prev, curr *MediaSession) bool {
	return nilAwareChanged(prev, curr, (*MediaSession).changeKey)
}
