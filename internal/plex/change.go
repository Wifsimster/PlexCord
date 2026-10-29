package plex

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

func (m *MusicSession) changeKey() changeKey {
	return changeKey{
		sessionKey: m.SessionKey,
		state:      m.State,
		title:      m.Track,
		artist:     m.Artist,
		album:      m.Album,
	}
}

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

func sessionChanged(prev, curr *MusicSession) bool {
	return nilAwareChanged(prev, curr, (*MusicSession).changeKey)
}

func mediaSessionChanged(prev, curr *MediaSession) bool {
	return nilAwareChanged(prev, curr, (*MediaSession).changeKey)
}
