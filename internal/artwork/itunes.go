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

// resolveITunes queries the keyless iTunes Search API for an album cover and
// upscales Apple's 100px thumbnail URL to a crisp 512px cover.
func (r *Resolver) resolveITunes(ctx context.Context, artist, album string) string {
	term := strings.TrimSpace(artist + " " + album)
	if term == "" {
		return ""
	}

	q := url.Values{}
	q.Set("term", term)
	q.Set("entity", "album")
	q.Set("limit", "1")
	endpoint := r.itunesBase + "/search?" + q.Encode()

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
