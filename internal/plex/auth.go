package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"plexcord/internal/errors"
)

const (
	plexPinURL  = "https://plex.tv/api/v2/pins"
	plexAuthURL = "https://app.plex.tv/auth#"
	clientID    = "plexcord-" // Will be appended with generated UUID
	productName = "PlexCord"
	version     = "1.0.0"
)

// PINResponse represents the response from creating a PIN
type PINResponse struct {
	ID               int    `json:"id"`
	Code             string `json:"code"` // The 4-digit PIN
	Product          string `json:"product"`
	Trusted          bool   `json:"trusted"`
	ClientIdentifier string `json:"clientIdentifier"`
	Location         struct {
		Code string `json:"code"`
		City string `json:"city"`
	} `json:"location"`
	ExpiresIn int       `json:"expiresIn"` // Seconds until expiration
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	AuthToken string    `json:"authToken"` // Present after user authorizes
}

// Authenticatorstores PIN authentication state
type Authenticator struct {
	httpClient *http.Client
	clientID   string
}

// NewAuthenticator creates a new PIN-based authenticator
func NewAuthenticator() *Authenticator {
	// Generate a unique client ID for this installation
	clientID := fmt.Sprintf("%s%d", clientID, time.Now().Unix())

	return &Authenticator{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		clientID: clientID,
	}
}

// RequestPIN requests a new PIN from plex.tv
func (a *Authenticator) RequestPIN(ctx context.Context) (*PINResponse, error) {
	// Build form data
	data := url.Values{}
	data.Set("strong", "true") // Request a 4-digit PIN

	req, err := http.NewRequestWithContext(ctx, "POST", plexPinURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create PIN request")
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Product", productName)
	req.Header.Set("X-Plex-Version", version)
	req.Header.Set("X-Plex-Client-Identifier", a.clientID)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to request PIN")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New(errors.PLEX_CONN_FAILED,
			fmt.Sprintf("PIN request failed with status %d: %s", resp.StatusCode, string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to read PIN response")
	}

	var pinResp PINResponse
	if err := json.Unmarshal(body, &pinResp); err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to parse PIN response")
	}

	return &pinResp, nil
}

// CheckPIN checks if the user has authorized the PIN and returns the auth token if available
func (a *Authenticator) CheckPIN(ctx context.Context, pinID int) (*PINResponse, error) {
	checkURL := fmt.Sprintf("%s/%d", plexPinURL, pinID)

	req, err := http.NewRequestWithContext(ctx, "GET", checkURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create PIN check request")
	}

	// Set required headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Client-Identifier", a.clientID)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to check PIN status")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New(errors.PLEX_CONN_FAILED,
			fmt.Sprintf("PIN check failed with status %d: %s", resp.StatusCode, string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to read PIN check response")
	}

	var pinResp PINResponse
	if err := json.Unmarshal(body, &pinResp); err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to parse PIN check response")
	}

	return &pinResp, nil
}

// GetAuthURL returns the URL the user should visit to authorize the PIN
func (a *Authenticator) GetAuthURL(pinCode string) string {
	params := url.Values{}
	params.Set("clientID", a.clientID)
	params.Set("code", pinCode)
	params.Set("context[device][product]", productName)

	return fmt.Sprintf("%s?%s", plexAuthURL, params.Encode())
}

// WaitForAuth polls for PIN authorization and returns the auth token when available
// This blocks until either the user authorizes, the PIN expires, or the context is canceled
func (a *Authenticator) WaitForAuth(ctx context.Context, pinID int, pinCode string) (string, error) {
	ticker := time.NewTicker(1 * time.Second) // Poll every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", errors.New(errors.PLEX_CONN_FAILED, "authentication canceled")
		case <-ticker.C:
			pinResp, err := a.CheckPIN(ctx, pinID)
			if err != nil {
				return "", err
			}

			// Check if we have an auth token
			if pinResp.AuthToken != "" {
				return pinResp.AuthToken, nil
			}

			// Check if PIN has expired
			if time.Now().After(pinResp.ExpiresAt) {
				return "", errors.New(errors.PLEX_CONN_FAILED, "PIN expired before authorization")
			}
		}
	}
}
