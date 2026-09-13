package artwork

import "strings"

// Media type constants naming what a Query is asking for. They mirror the
// values plex.MediaSession.MediaType carries, so a caller can hand the session's
// type straight to a Query without translating it.
const (
	MediaTypeMusic = "music"
	MediaTypeMovie = "movie"
	MediaTypeTV    = "tv"
)

// Query describes the artwork being looked for. It replaces the artist/album
// pair the resolver originally took: a movie has no artist and a TV episode's
// poster belongs to its show, so the lookup needs to know which kind of thing
// it is searching for before it can pick the right search term — and the right
// provider.
//
// Fields not relevant to a media type are simply left zero.
type Query struct {
	// MediaType is MediaTypeMusic, MediaTypeMovie or MediaTypeTV. Empty means
	// music, which keeps a zero-value Query behaving like the original
	// artist/album lookup.
	MediaType string

	// Music fields.
	Artist string
	Album  string

	// Video fields: the movie's title, or the show's name for a TV episode —
	// never the episode's own title, which no poster database indexes.
	Title string

	// Year disambiguates remakes and reboots when the provider supports it.
	Year int
}

// MusicQuery builds the music lookup, the shape every caller used before the
// resolver learned about video.
func MusicQuery(artist, album string) Query {
	return Query{MediaType: MediaTypeMusic, Artist: artist, Album: album}
}

// MovieQuery builds the lookup for a film.
func MovieQuery(title string, year int) Query {
	return Query{MediaType: MediaTypeMovie, Title: title, Year: year}
}

// ShowQuery builds the lookup for a TV episode, which is keyed on the show
// rather than the episode: posters are published per show/season.
func ShowQuery(showTitle string, year int) Query {
	return Query{MediaType: MediaTypeTV, Title: showTitle, Year: year}
}

// normalized trims the text fields and fills in the default media type, so a
// caller's stray whitespace cannot produce two cache entries for one album.
func (q Query) normalized() Query {
	q.MediaType = strings.TrimSpace(strings.ToLower(q.MediaType))
	if q.MediaType == "" {
		q.MediaType = MediaTypeMusic
	}
	q.Artist = strings.TrimSpace(q.Artist)
	q.Album = strings.TrimSpace(q.Album)
	q.Title = strings.TrimSpace(q.Title)
	return q
}

// isEmpty reports whether the query carries nothing to search on, in which case
// every provider would miss and the round trip is not worth making.
func (q Query) isEmpty() bool {
	switch q.MediaType {
	case MediaTypeMovie, MediaTypeTV:
		return q.Title == ""
	default:
		return q.Artist == "" && q.Album == ""
	}
}

// cacheKey is the stable, case-insensitive identity of the query. The media
// type is part of it so a film and an album that happen to share a name cannot
// collide in the cache.
func (q Query) cacheKey() string {
	return strings.ToLower(strings.Join([]string{q.MediaType, q.Artist, q.Album, q.Title}, "\x00"))
}
