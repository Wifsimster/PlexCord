package plex

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"plexcord/internal/errors"
)

type transport struct {
	httpClient *http.Client
	serverURL  string
	token      string
}

func (t *transport) get(ctx context.Context, path string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s?X-Plex-Token=%s", t.serverURL, path, url.QueryEscape(t.token))
	return t.doRequest(ctx, "GET", reqURL)
}

func (t *transport) doRequest(ctx context.Context, method, reqURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create request")
	}

	req.Header.Set("User-Agent", "PlexCord/1.0")
	req.Header.Set("Accept", "application/xml")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, mapHTTPError(err, ctx)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, mapHTTPStatusCode(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to read response body")
	}
	return body, nil
}

func parseSessionsResponse(body []byte) (*SessionsResponse, error) {
	var resp SessionsResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "invalid sessions response format")
	}
	return &resp, nil
}
