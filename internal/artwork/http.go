package artwork

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// maxBodyBytes caps how much of an API response we read into memory.
const maxBodyBytes = 1 << 20 // 1 MiB

// getJSON performs a GET and decodes a JSON body into out. It returns false on
// any transport, status, or decode failure — artwork lookup is best-effort and
// must never surface an error into the presence path.
func (r *Resolver) getJSON(ctx context.Context, endpoint string, out any) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", r.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := r.http.Do(req)
	if err != nil {
		return false
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return false
	}
	return json.Unmarshal(body, out) == nil
}

// exists reports whether u resolves to a fetchable resource (following
// redirects), used to confirm a Cover Art Archive image is present.
func (r *Resolver) exists(ctx context.Context, u string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", r.userAgent)

	resp, err := r.http.Do(req)
	if err != nil {
		return false
	}
	defer closeBody(resp.Body)
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// closeBody closes a response body best-effort. A close error on a read-only
// GET/HEAD is not actionable; it is handled here so the linters see it checked.
func closeBody(rc io.Closer) {
	if err := rc.Close(); err != nil {
		_ = err
	}
}
