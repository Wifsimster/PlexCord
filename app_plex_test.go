package main

import (
	"errors"
	"testing"
	"time"

	"plexcord/internal/config"
	plexerrors "plexcord/internal/errors"
	"plexcord/internal/plex"
)

func TestStartPlexPINAuthReturnsCodeAndURL(t *testing.T) {
	expires := time.Now().Add(15 * time.Minute)
	auth := &fakeAuthenticator{
		requestResp: &plex.PINResponse{ID: 314159, Code: "ABCD", ExpiresIn: 900, ExpiresAt: expires},
		authURL:     "https://plex.tv/link?pin=ABCD",
	}
	app := newTestApp(config.DefaultConfig())
	app.authFactory = func() PlexAuthenticator { return auth }

	result, err := app.StartPlexPINAuth()
	if err != nil {
		t.Fatalf("StartPlexPINAuth() error: %v", err)
	}

	if result["pinCode"] != "ABCD" {
		t.Errorf("pinCode = %v, want ABCD", result["pinCode"])
	}
	if result["pinID"] != 314159 {
		t.Errorf("pinID = %v, want 314159", result["pinID"])
	}
	if result["authURL"] != "https://plex.tv/link?pin=ABCD" {
		t.Errorf("authURL = %v, want the plex.tv link URL", result["authURL"])
	}
}

func TestCheckPlexPINAuthWithoutStart(t *testing.T) {
	app := newTestApp(config.DefaultConfig())

	if _, err := app.CheckPlexPINAuth(1); err == nil {
		t.Error("CheckPlexPINAuth() = nil before a PIN was requested, want an error")
	}
}

func TestCheckPlexPINAuthAuthorized(t *testing.T) {
	auth := &fakeAuthenticator{
		requestResp: &plex.PINResponse{ID: 1, Code: "ABCD", ExpiresAt: time.Now().Add(time.Minute)},
		checkResp:   &plex.PINResponse{ID: 1, AuthToken: "a-plex-token", ExpiresAt: time.Now().Add(time.Minute)},
	}
	app := newTestApp(config.DefaultConfig())
	app.authFactory = func() PlexAuthenticator { return auth }

	if _, err := app.StartPlexPINAuth(); err != nil {
		t.Fatalf("StartPlexPINAuth() error: %v", err)
	}

	result, err := app.CheckPlexPINAuth(1)
	if err != nil {
		t.Fatalf("CheckPlexPINAuth() error: %v", err)
	}
	if result["authorized"] != true {
		t.Errorf("authorized = %v, want true", result["authorized"])
	}
	if result["authToken"] != "a-plex-token" {
		t.Errorf("authToken = %v, want the issued token", result["authToken"])
	}

	if _, err := app.CheckPlexPINAuth(1); err == nil {
		t.Error("CheckPlexPINAuth() succeeded after the PIN was consumed")
	}
}

func TestCheckPlexPINAuthExpired(t *testing.T) {
	auth := &fakeAuthenticator{
		requestResp: &plex.PINResponse{ID: 1, Code: "ABCD", ExpiresAt: time.Now().Add(time.Minute)},
		checkResp:   &plex.PINResponse{ID: 1, ExpiresAt: time.Now().Add(-time.Minute)},
	}
	app := newTestApp(config.DefaultConfig())
	app.authFactory = func() PlexAuthenticator { return auth }

	if _, err := app.StartPlexPINAuth(); err != nil {
		t.Fatalf("StartPlexPINAuth() error: %v", err)
	}

	result, err := app.CheckPlexPINAuth(1)
	if err != nil {
		t.Fatalf("CheckPlexPINAuth() error: %v", err)
	}
	if result["authorized"] != false || result["expired"] != true {
		t.Errorf("result = %v, want an expired, unauthorized PIN", result)
	}
}

func TestStartPlexRetryOnlyForRecoverableErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantRetry bool
	}{
		{
			name:      "connection failure retries",
			err:       plexerrors.New(plexerrors.PLEX_UNREACHABLE, "cannot reach server"),
			wantRetry: true,
		},
		{
			name:      "auth failure does not retry",
			err:       plexerrors.New(plexerrors.PLEX_AUTH_FAILED, "invalid token"),
			wantRetry: false,
		},
		{
			name:      "an unknown error is treated as retryable",
			err:       errors.New("something went wrong"),
			wantRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(config.DefaultConfig())
			fake := app.plexRetry.(*fakeRetry)

			app.startPlexRetry(tt.err)

			if got := fake.startCount() > 0; got != tt.wantRetry {
				t.Errorf("retry started = %v, want %v", got, tt.wantRetry)
			}
		})
	}
}

func TestRetryPlexConnectionRequiresConfiguration(t *testing.T) {
	app := newTestApp(config.DefaultConfig())

	if err := app.RetryPlexConnection(); err == nil {
		t.Error("RetryPlexConnection() = nil with nothing configured, want an error")
	}

	app.config.ServerURL = "http://plex.test:32400"
	app.config.SelectedPlexUserID = "user1"
	if err := app.RetryPlexConnection(); err != nil {
		t.Fatalf("RetryPlexConnection() error with a configured server: %v", err)
	}
	if got := app.plexRetry.(*fakeRetry).manualRetries; got != 1 {
		t.Errorf("manual retries = %d, want 1", got)
	}
}

func TestValidatePlexConnectionStopsRetries(t *testing.T) {
	app := newTestApp(config.DefaultConfig())
	app.tokens = &fakeTokenStore{token: "a-token"}

	if _, err := app.ValidatePlexConnection("http://plex.test:32400"); err != nil {
		t.Fatalf("ValidatePlexConnection() error: %v", err)
	}

	if got := app.plexRetry.(*fakeRetry).resets; got != 1 {
		t.Errorf("retry resets = %d, want 1 after a successful validation", got)
	}
	if app.config.PlexLastConnected == nil {
		t.Error("a successful validation did not record the connection time")
	}
}

func TestValidatePlexConnectionRequiresToken(t *testing.T) {
	app := newTestApp(config.DefaultConfig())
	app.tokens = &fakeTokenStore{}

	if _, err := app.ValidatePlexConnection("http://plex.test:32400"); err == nil {
		t.Error("ValidatePlexConnection() = nil with no stored token, want an error")
	}
}

func TestDiscoverPlexServersUsesInjectedDiscoverer(t *testing.T) {
	app := newTestApp(config.DefaultConfig())
	app.discovery = &fakeDiscoverer{servers: []plex.Server{
		{Name: "Living Room", Address: "192.168.0.10", Port: "32400", IsLocal: true},
	}}

	servers, err := app.DiscoverPlexServers()
	if err != nil {
		t.Fatalf("DiscoverPlexServers() error: %v", err)
	}
	if len(servers) != 1 || servers[0].Name != "Living Room" {
		t.Errorf("servers = %+v, want the discovered server", servers)
	}
}

func TestSaveServerURLRejectsNonHTTPSchemes(t *testing.T) {
	app := newTestApp(config.DefaultConfig())

	for _, bad := range []string{"", "file:///etc/passwd", "javascript:alert(1)", "ftp://plex.test"} {
		if err := app.SaveServerURL(bad); err == nil {
			t.Errorf("SaveServerURL(%q) = nil, want the URL rejected", bad)
		}
	}

	if err := app.SaveServerURL("http://plex.test:32400"); err != nil {
		t.Errorf("SaveServerURL() rejected a valid http URL: %v", err)
	}
}

func TestPlexRecoveryRestoresPresence(t *testing.T) {
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, nil)

	app.sessions.Set(playingSession("Song"))
	app.onPlexPollError(errors.New("timeout"))
	app.onPlexPollRecovered()

	if presence.updateCount() != 1 {
		t.Fatalf("presence updates = %d, want 1 after recovery", presence.updateCount())
	}
}

func TestPlexRecoveryKeepsManualPause(t *testing.T) {
	presence := &recordingPresence{connected: true}
	app := newTestApp(config.DefaultConfig())
	app.discord = app.newDiscordService(presence, nil)

	app.sessions.Set(playingSession("Song"))
	app.presence.Toggle()
	app.onPlexPollRecovered()

	if presence.updateCount() != 0 {
		t.Errorf("presence updates = %d, want 0 while manually paused", presence.updateCount())
	}
}
