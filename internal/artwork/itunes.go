package artwork

import (
	"context"
	"net/url"
	"strings"
)

type itunesResponse struct {
	Results []struct {
		ArtworkURL100 string `json:"artworkUrl100"`
	} `json:"results"`
}

// itunesSearch picks the search term and the iTunes entity for a query. The
// entity is what makes the same endpoint answer with an album cover, a film
// poster or a show's season art, so it — not the term alone — is what the
// media type decides.
func itunesSearch(q Query) (term, entity string) {
	switch q.MediaType {
	case MediaTypeMovie:
		return q.Title, "movie"
	case MediaTypeTV:
		return q.Title, "tvSeason"
	default:
		return strings.TrimSpace(q.Artist + " " + q.Album), "album"
	}
}

// resolveITunes queries the keyless iTunes Search API for cover or poster art
// and upscales Apple's 100px thumbnail URL to a crisp 512px image.
func (r *Resolver) resolveITunes(ctx context.Context, q Query) string {
	term, entity := itunesSearch(q)
	if term == "" {
		return ""
	}

	v := url.Values{}
	v.Set("term", term)
	v.Set("entity", entity)
	v.Set("limit", "1")
	endpoint := r.itunesBase + "/search?" + v.Encode()

	var resp itunesResponse
	if !r.getJSON(ctx, endpoint, &resp) {
		return ""
	}
	if len(resp.Results) == 0 {
		return ""
	}
	art := resp.Results[0].ArtworkURL100
	if art == "" {
		return ""
	}
	return strings.Replace(art, "100x100bb", "512x512bb", 1)
}
