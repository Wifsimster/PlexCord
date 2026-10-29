package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"plexcord/internal/errors"
	"plexcord/internal/version"
)

const (
	plexPinURL  = "https://plex.tv/api/v2/pins"
	plexAuthURL = "https://app.plex.tv/auth#"
	clientID    = "plexcord-"
	productName = "PlexCord"
)

type PINResponse struct {
	Location struct {
		Code string `json:"code"`
		City string `json:"city"`
	} `json:"location"`
	CreatedAt        time.Time `json:"createdAt"`
	ExpiresAt        time.Time `json:"expiresAt"`
	Code             string    `json:"code"`
	Product          string    `json:"product"`
	ClientIdentifier string    `json:"clientIdentifier"`
	AuthToken        string    `json:"authToken"`
	ID               int       `json:"id"`
	ExpiresIn        int       `json:"expiresIn"`
	Trusted          bool      `json:"trusted"`
}

type Authenticator struct {
	httpClient *http.Client
	clientID   string
}

func NewAuthenticator() *Authenticator {
	clientID := fmt.Sprintf("%s%d", clientID, time.Now().Unix())

	return &Authenticator{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		clientID: clientID,
	}
}

func (a *Authenticator) RequestPIN(ctx context.Context) (*PINResponse, error) {
	data := url.Values{}
	data.Set("strong", "true")

	req, err := http.NewRequestWithContext(ctx, "POST", plexPinURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create PIN request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Product", productName)
	req.Header.Set("X-Plex-Version", version.Version)
	req.Header.Set("X-Plex-Client-Identifier", a.clientID)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to request PIN")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("unable to read error response")
		}
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

func (a *Authenticator) CheckPIN(ctx context.Context, pinID int) (*PINResponse, error) {
	checkURL := fmt.Sprintf("%s/%d", plexPinURL, pinID)

	req, err := http.NewRequestWithContext(ctx, "GET", checkURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create PIN check request")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Client-Identifier", a.clientID)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to check PIN status")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("unable to read error response")
		}
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

func (a *Authenticator) GetAuthURL(pinCode string) string {
	params := url.Values{}
	params.Set("clientID", a.clientID)
	params.Set("code", pinCode)
	params.Set("context[device][product]", productName)

	return fmt.Sprintf("%s?%s", plexAuthURL, params.Encode())
}

func (a *Authenticator) WaitForAuth(ctx context.Context, pinID int, pinCode string) (string, error) {
	ticker := time.NewTicker(1 * time.Second)
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

			if pinResp.AuthToken != "" {
				return pinResp.AuthToken, nil
			}

			if time.Now().After(pinResp.ExpiresAt) {
				return "", errors.New(errors.PLEX_CONN_FAILED, "PIN expired before authorization")
			}
		}
	}
}
