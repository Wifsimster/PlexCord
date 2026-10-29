package artwork

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

const maxBodyBytes = 1 << 20

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

func closeBody(rc io.Closer) {
	if err := rc.Close(); err != nil {
		_ = err
	}
}
