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

// Client handles Plex Media Server communication
type Client struct {
	httpClient *http.Client
	serverURL  string
	token      string
}

// NewClient creates a new Plex client with the given token and server URL
func NewClient(token, serverURL string) *Client {
	return &Client{
		token:     token,
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// IdentityResponse represents the XML response from /identity endpoint
type IdentityResponse struct {
	XMLName           xml.Name `xml:"MediaContainer"`
	MachineIdentifier string   `xml:"machineIdentifier,attr"`
	Version           string   `xml:"version,attr"`
	Claimed           string   `xml:"claimed,attr"`
	FriendlyName      string   `xml:"friendlyName,attr"` // May be empty on some servers
	Size              int      `xml:"size,attr"`
}

// LibraryResponse represents the XML response from /library/sections endpoint
type LibraryResponse struct {
	XMLName xml.Name `xml:"MediaContainer"`
	Size    int      `xml:"size,attr"`
}

// ValidateConnection validates the Plex server connection by querying server info and library count
func (c *Client) ValidateConnection() (*ValidationResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if this is a claim token (not valid for API authentication)
	if strings.HasPrefix(c.token, "claim-") {
		return nil, errors.New(errors.PLEX_AUTH_FAILED,
			"Invalid token type: You provided a claim token (starts with 'claim-'). "+
				"Claim tokens are only used to link servers to your Plex account. "+
				"Please get your authentication token from plex.tv instead. "+
				"Visit: https://www.plex.tv/claim and sign in, then go to Settings > Account to get your X-Plex-Token.")
	}

	// Step 1: Get server identity using /identity endpoint
	// Note: This endpoint doesn't require authentication
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

	// Step 2: Get library count
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
		// Read the response body for debugging
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

	// Step 3: Return validation result
	// Use a default server name if friendlyName is not provided
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

// GetUsers retrieves the list of Plex users/accounts that can be monitored
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
		// Use fallback name if account name is empty
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

// GetSessions retrieves active playback sessions from the Plex server.
// It filters sessions to only include those belonging to the specified userID.
// Returns an empty slice (not error) when no sessions are active.
// Uses 500ms timeout per NFR5 performance requirement.
func (c *Client) GetSessions(userID string) ([]Session, error) {
	// Use 500ms timeout for polling performance (NFR5)
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

	// Filter and convert sessions for the specified user
	sessions := make([]Session, 0, len(sessionsResp.Sessions))
	for _, entry := range sessionsResp.Sessions {
		// Filter by user ID if specified
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

	// Return empty slice (not nil) when no sessions - this is not an error
	if sessions == nil {
		sessions = []Session{}
	}

	return sessions, nil
}

// GetMusicSessions retrieves active music sessions for the specified user.
// Convenience method that calls GetSessions and filters for music (type="track").
// Applies fallback values for missing metadata and builds absolute artwork URLs.
func (c *Client) GetMusicSessions(userID string) ([]MusicSession, error) {
	// Use 500ms timeout for polling performance (NFR5)
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

	// Filter for music sessions (type="track") belonging to specified user
	musicSessions := make([]MusicSession, 0, len(sessionsResp.Sessions))
	for _, entry := range sessionsResp.Sessions {
		// Filter by user ID if specified
		if userID != "" && entry.User.ID != userID {
			continue
		}

		// Only include music sessions (type="track")
		if entry.Type != "track" {
			continue
		}

		// Build absolute artwork URL if thumb path exists (AC4)
		thumbURL := ""
		if entry.Thumb != "" {
			thumbURL = c.buildArtworkURL(entry.Thumb)
		}

		session := MusicSession{
			Session: Session{
				SessionKey: entry.SessionKey,
				UserID:     entry.User.ID,
				UserName:   entry.User.Title,
				Type:       entry.Type,
				State:      entry.Player.State,
				PlayerName: entry.Player.Title,
			},
			Track:      entry.Title,
			Artist:     entry.GrandparentTitle,
			Album:      entry.ParentTitle,
			Thumb:      entry.Thumb,
			ThumbURL:   thumbURL,
			Duration:   entry.Duration,
			ViewOffset: entry.ViewOffset,
		}

		// Apply fallback values for missing metadata (AC1, AC2, AC3, AC7)
		session.ApplyFallbacks()

		musicSessions = append(musicSessions, session)
	}

	// Return empty slice (not nil) when no sessions - this is not an error
	if musicSessions == nil {
		musicSessions = []MusicSession{}
	}

	return musicSessions, nil
}

// buildArtworkURL constructs an absolute URL for album artwork.
// Combines server URL, thumb path, and authentication token.
// Returns empty string if thumbPath is empty.
func (c *Client) buildArtworkURL(thumbPath string) string {
	if thumbPath == "" {
		return ""
	}
	// Build absolute URL: {serverURL}{thumb}?X-Plex-Token={token}
	return fmt.Sprintf("%s%s?X-Plex-Token=%s", c.serverURL, thumbPath, url.QueryEscape(c.token))
}

// mapHTTPError maps HTTP client errors to appropriate error codes
func mapHTTPError(err error, ctx context.Context) error {
	// Check for timeout first
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New(errors.TIMEOUT, "connection timed out after 5 seconds")
	}

	// Check for network/connection errors
	if err != nil {
		errStr := strings.ToLower(err.Error())
		// URL errors (DNS, connection refused, etc.)
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

// mapHTTPStatusCode maps HTTP status codes to appropriate error codes
func mapHTTPStatusCode(statusCode int) error {
	switch statusCode {
	case http.StatusBadRequest:
		// 400 Bad Request - often indicates token format issues
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
