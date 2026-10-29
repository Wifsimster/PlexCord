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

func (r *Resolver) resolveCoverArt(ctx context.Context, q Query) string {
	if q.MediaType != MediaTypeMusic || strings.TrimSpace(q.Album) == "" {
		return ""
	}
	r.mbLimiter.wait()

	query := fmt.Sprintf(`release:%q`, q.Album)
	if a := strings.TrimSpace(q.Artist); a != "" {
		query += fmt.Sprintf(` AND artist:%q`, a)
	}
	v := url.Values{}
	v.Set("query", query)
	v.Set("fmt", "json")
	v.Set("limit", "1")
	endpoint := r.mbBase + "/ws/2/release/?" + v.Encode()

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
