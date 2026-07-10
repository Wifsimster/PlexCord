package artwork

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type mbReleaseResponse struct {
	Releases []struct {
		ID string `json:"id"`
	} `json:"releases"`
}

// resolveCoverArt is the keyless MusicBrainz + Cover Art Archive fallback: it
// finds a release MBID by artist+album, then confirms and returns the public
// Cover Art Archive front-cover URL for it. MusicBrainz requests are throttled
// to its 1 req/s policy and carry a descriptive User-Agent.
func (r *Resolver) resolveCoverArt(ctx context.Context, artist, album string) string {
	if strings.TrimSpace(album) == "" {
		return ""
	}
	r.mbLimiter.wait()

	query := fmt.Sprintf(`release:%q`, album)
	if a := strings.TrimSpace(artist); a != "" {
		query += fmt.Sprintf(` AND artist:%q`, a)
	}
	q := url.Values{}
	q.Set("query", query)
	q.Set("fmt", "json")
	q.Set("limit", "1")
	endpoint := r.mbBase + "/ws/2/release/?" + q.Encode()

	var resp mbReleaseResponse
	if !r.getJSON(ctx, endpoint, &resp) {
		return ""
	}
	if len(resp.Releases) == 0 || resp.Releases[0].ID == "" {
		return ""
	}

	caaURL := r.caaBase + "/release/" + resp.Releases[0].ID + "/front-500"
	if !r.exists(ctx, caaURL) {
		return ""
	}
	return caaURL
}
