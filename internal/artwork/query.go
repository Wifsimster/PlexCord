package artwork

import "strings"

const (
	MediaTypeMusic = "music"
	MediaTypeMovie = "movie"
	MediaTypeTV    = "tv"
)

type Query struct {
	MediaType string

	Artist string
	Album  string

	Title string

	Year int
}

func MusicQuery(artist, album string) Query {
	return Query{MediaType: MediaTypeMusic, Artist: artist, Album: album}
}

func MovieQuery(title string, year int) Query {
	return Query{MediaType: MediaTypeMovie, Title: title, Year: year}
}

func ShowQuery(showTitle string, year int) Query {
	return Query{MediaType: MediaTypeTV, Title: showTitle, Year: year}
}

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

func (q Query) isEmpty() bool {
	switch q.MediaType {
	case MediaTypeMovie, MediaTypeTV:
		return q.Title == ""
	default:
		return q.Artist == "" && q.Album == ""
	}
}

func (q Query) cacheKey() string {
	return strings.ToLower(strings.Join([]string{q.MediaType, q.Artist, q.Album, q.Title}, "\x00"))
}
