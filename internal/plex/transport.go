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

// transport encapsulates the raw HTTP concerns of talking to a Plex server:
// URL + auth, user agent, timeouts, error classification, and body reading.
// This separates transport from parsing and domain filtering logic, so each
// layer can be tested in isolation.
//
// Methods on *Client that need to talk to the server should delegate here
// rather than duplicating the same http.NewRequestWithContext boilerplate.
type transport struct {
	httpClient *http.Client
	serverURL  string
	token      string
}

// get executes a GET request against the given path with standard headers
// and returns the response body bytes. The path should start with "/".
// Authentication is always appended via X-Plex-Token query parameter.
// The response body is fully read and closed before returning.
func (t *transport) get(ctx context.Context, path string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s?X-Plex-Token=%s", t.serverURL, path, url.QueryEscape(t.token))
	return t.doRequest(ctx, "GET", reqURL)
}

// doRequest is the shared HTTP execution path — standard headers, error
// mapping, body read, and close. All fetch operations go through here.
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

// ----------------------------------------------------------------------------
// Parser functions — pure XML → domain-object transformations with no I/O.
// These are trivially unit-testable with in-memory byte slices.
// ----------------------------------------------------------------------------

// parseSessionsResponse deserializes the /status/sessions XML payload into
// a SessionsResponse. Returns a wrapped error on invalid XML.
func parseSessionsResponse(body []byte) (*SessionsResponse, error) {
	var resp SessionsResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "invalid sessions response format")
	}
	return &resp, nil
}
