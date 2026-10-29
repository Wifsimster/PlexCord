package plex

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"plexcord/internal/errors"
)

type Client struct {
	httpClient *http.Client
	serverURL  string
	token      string
}

func NewClient(token, serverURL string) *Client {
	return &Client{
		token:     token,
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type IdentityResponse struct {
	XMLName           xml.Name `xml:"MediaContainer"`
	MachineIdentifier string   `xml:"machineIdentifier,attr"`
	Version           string   `xml:"version,attr"`
	Claimed           string   `xml:"claimed,attr"`
	FriendlyName      string   `xml:"friendlyName,attr"`
	Size              int      `xml:"size,attr"`
}

type LibraryResponse struct {
	XMLName xml.Name `xml:"MediaContainer"`
	Size    int      `xml:"size,attr"`
}

func (c *Client) ValidateConnection() (*ValidationResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if strings.HasPrefix(c.token, "claim-") {
		return nil, errors.New(errors.PLEX_AUTH_FAILED,
			"Invalid token type: You provided a claim token (starts with 'claim-'). "+
				"Claim tokens are only used to link servers to your Plex account. "+
				"Please get your authentication token from plex.tv instead. "+
				"Visit: https://www.plex.tv/claim and sign in, then go to Settings > Account to get your X-Plex-Token.")
	}

	identityURL := fmt.Sprintf("%s/identity", c.serverURL)
	identityReq, err := http.NewRequestWithContext(ctx, "GET", identityURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create identity request")
	}

	identityReq.Header.Set("User-Agent", "PlexCord/1.0")
	identityReq.Header.Set("Accept", "application/xml")

	identityResp, err := c.httpClient.Do(identityReq)
	if err != nil {
		return nil, mapHTTPError(err, ctx)
	}
	defer func() {
		if err := identityResp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close response body: %v", err)
		}
	}()

	if identityResp.StatusCode != http.StatusOK {
		return nil, mapHTTPStatusCode(identityResp.StatusCode)
	}

	identityBody, err := io.ReadAll(identityResp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to read identity response")
	}

	var identity IdentityResponse
	if err := xml.Unmarshal(identityBody, &identity); err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "invalid server response format")
	}

	libraryURL := fmt.Sprintf("%s/library/sections/?X-Plex-Token=%s", c.serverURL, url.QueryEscape(c.token))

	libraryReq, err := http.NewRequestWithContext(ctx, "GET", libraryURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create library request")
	}

	libraryReq.Header.Set("User-Agent", "PlexCord/1.0")
	libraryReq.Header.Set("Accept", "application/xml")

	libraryResp, err := c.httpClient.Do(libraryReq)
	if err != nil {
		return nil, mapHTTPError(err, ctx)
	}
	defer func() {
		if err := libraryResp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close response body: %v", err)
		}
	}()

	log.Printf("Library validation response status: %d %s", libraryResp.StatusCode, libraryResp.Status)

	if libraryResp.StatusCode != http.StatusOK {
		if bodyBytes, err := io.ReadAll(libraryResp.Body); err == nil {
			log.Printf("Library validation failed with status %d, body: %s", libraryResp.StatusCode, string(bodyBytes))
		}
		return nil, mapHTTPStatusCode(libraryResp.StatusCode)
	}

	libraryBody, err := io.ReadAll(libraryResp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to read library response")
	}

	var library LibraryResponse
	if err := xml.Unmarshal(libraryBody, &library); err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "invalid library response format")
	}

	serverName := identity.FriendlyName
	if serverName == "" {
		serverName = "Plex Media Server"
	}

	return &ValidationResult{
		Success:           true,
		ServerName:        serverName,
		ServerVersion:     identity.Version,
		LibraryCount:      library.Size,
		MachineIdentifier: identity.MachineIdentifier,
	}, nil
}

func (c *Client) GetUsers() ([]PlexUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accountsURL := fmt.Sprintf("%s/accounts?X-Plex-Token=%s", c.serverURL, url.QueryEscape(c.token))
	req, err := http.NewRequestWithContext(ctx, "GET", accountsURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create accounts request")
	}

	req.Header.Set("User-Agent", "PlexCord/1.0")
	req.Header.Set("Accept", "application/xml")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, mapHTTPError(err, ctx)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, mapHTTPStatusCode(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to read accounts response")
	}

	var accounts AccountsResponse
	if err := xml.Unmarshal(body, &accounts); err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "invalid accounts response format")
	}

	users := make([]PlexUser, len(accounts.Accounts))
	for i, acc := range accounts.Accounts {
		name := acc.Name
		if name == "" {
			name = fmt.Sprintf("User %s", acc.ID)
		}

		users[i] = PlexUser{
			ID:    acc.ID,
			Name:  name,
			Thumb: acc.Thumb,
		}
	}

	return users, nil
}

func (c *Client) GetSessions(userID string) ([]Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	sessionsURL := fmt.Sprintf("%s/status/sessions?X-Plex-Token=%s", c.serverURL, url.QueryEscape(c.token))
	req, err := http.NewRequestWithContext(ctx, "GET", sessionsURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create sessions request")
	}

	req.Header.Set("User-Agent", "PlexCord/1.0")
	req.Header.Set("Accept", "application/xml")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, mapHTTPError(err, ctx)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, mapHTTPStatusCode(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to read sessions response")
	}

	var sessionsResp SessionsResponse
	if err := xml.Unmarshal(body, &sessionsResp); err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "invalid sessions response format")
	}

	allEntries := sessionsResp.AllEntries()
	sessions := make([]Session, 0, len(allEntries))
	for _, entry := range allEntries {
		if userID != "" && entry.User.ID != userID {
			continue
		}

		sessions = append(sessions, Session{
			SessionKey: entry.SessionKey,
			UserID:     entry.User.ID,
			UserName:   entry.User.Title,
			Type:       entry.Type,
			State:      entry.Player.State,
			PlayerName: entry.Player.Title,
		})
	}

	if sessions == nil {
		sessions = []Session{}
	}

	return sessions, nil
}

func (c *Client) GetMusicSessions(userID string) ([]MusicSession, error) {
	sessionsResp, err := c.fetchSessions()
	if err != nil {
		return nil, err
	}
	return filterMusicSessions(sessionsResp, userID, c.buildArtworkURL), nil
}

func (c *Client) GetMediaSessions(userID string, mediaTypes []string) ([]MediaSession, error) {
	sessionsResp, err := c.fetchSessions()
	if err != nil {
		return nil, err
	}
	return filterMediaSessions(sessionsResp, userID, mediaTypes, c.buildArtworkURL), nil
}

func (c *Client) fetchSessions() (*SessionsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	body, err := c.transport().get(ctx, "/status/sessions")
	if err != nil {
		return nil, err
	}
	return parseSessionsResponse(body)
}

func (c *Client) transport() *transport {
	return &transport{
		httpClient: c.httpClient,
		serverURL:  c.serverURL,
		token:      c.token,
	}
}

func (c *Client) buildArtworkURL(thumbPath string) string {
	if thumbPath == "" {
		return ""
	}
	return fmt.Sprintf("%s%s?X-Plex-Token=%s", c.serverURL, thumbPath, url.QueryEscape(c.token))
}

func mapHTTPError(err error, ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New(errors.TIMEOUT, "connection timed out after 5 seconds")
	}

	if err != nil {
		errStr := strings.ToLower(err.Error())
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				return errors.New(errors.TIMEOUT, "connection timed out")
			}
			if strings.Contains(errStr, "connection refused") ||
				strings.Contains(errStr, "no such host") ||
				strings.Contains(errStr, "network is unreachable") {
				return errors.New(errors.PLEX_UNREACHABLE, "cannot reach Plex server - check URL and network")
			}
		}
		return errors.Wrap(err, errors.PLEX_UNREACHABLE, "failed to connect to server")
	}

	return errors.New(errors.PLEX_CONN_FAILED, "connection failed")
}

func mapHTTPStatusCode(statusCode int) error {
	switch statusCode {
	case http.StatusBadRequest:
		return errors.New(errors.PLEX_AUTH_FAILED, "invalid request - check token format")
	case http.StatusUnauthorized, http.StatusForbidden:
		return errors.New(errors.PLEX_AUTH_FAILED, "invalid Plex token - authentication failed")
	case http.StatusNotFound:
		return errors.New(errors.PLEX_UNREACHABLE, "server endpoint not found")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return errors.New(errors.PLEX_UNREACHABLE, "Plex server error")
	default:
		return errors.New(errors.PLEX_CONN_FAILED, fmt.Sprintf("server returned status %d", statusCode))
	}
}
