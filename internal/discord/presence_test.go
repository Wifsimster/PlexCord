package discord

import (
	"fmt"
	"testing"
	"time"

	plexerrors "plexcord/internal/errors"
)

func TestNewPresenceManager(t *testing.T) {
	pm := NewPresenceManager()

	if pm == nil {
		t.Fatal("NewPresenceManager returned nil")
	}

	if pm.connected {
		t.Error("New PresenceManager should not be connected")
	}

	if pm.clientID != DefaultClientID {
		t.Errorf("Expected default client ID %s, got %s", DefaultClientID, pm.clientID)
	}

	if pm.presence != nil {
		t.Error("New PresenceManager should have nil presence")
	}
}

func TestIsConnected(t *testing.T) {
	pm := NewPresenceManager()

	if pm.IsConnected() {
		t.Error("New PresenceManager should not be connected")
	}

	// Manually set connected for testing
	pm.connected = true
	if !pm.IsConnected() {
		t.Error("IsConnected should return true when connected")
	}
}

func TestGetClientID(t *testing.T) {
	pm := NewPresenceManager()

	if pm.GetClientID() != DefaultClientID {
		t.Errorf("Expected default client ID %s, got %s", DefaultClientID, pm.GetClientID())
	}

	// Test with custom client ID
	pm.clientID = "123456789012345678"
	if pm.GetClientID() != "123456789012345678" {
		t.Error("GetClientID should return the set client ID")
	}
}

func TestIsValidClientID(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		want     bool
	}{
		{
			name:     "valid client ID",
			clientID: "12345678901234567",
			want:     true,
		},
		{
			name:     "valid long client ID",
			clientID: "123456789012345678901",
			want:     true,
		},
		{
			name:     "default client ID",
			clientID: DefaultClientID,
			want:     true,
		},
		{
			name:     "empty client ID",
			clientID: "",
			want:     false,
		},
		{
			name:     "too short",
			clientID: "1234567890123456",
			want:     false,
		},
		{
			name:     "contains letters",
			clientID: "1234567890123456a",
			want:     false,
		},
		{
			name:     "contains special characters",
			clientID: "1234567890123456!",
			want:     false,
		},
		{
			name:     "contains spaces",
			clientID: "1234567890123456 ",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidClientID(tt.clientID); got != tt.want {
				t.Errorf("isValidClientID(%q) = %v, want %v", tt.clientID, got, tt.want)
			}
		})
	}
}

func TestValidateClientID(t *testing.T) {
	tests := []struct {
		name      string
		clientID  string
		wantErr   bool
		wantCode  string
	}{
		{
			name:     "valid client ID",
			clientID: "12345678901234567",
			wantErr:  false,
		},
		{
			name:     "valid long client ID",
			clientID: "123456789012345678901",
			wantErr:  false,
		},
		{
			name:     "default client ID",
			clientID: DefaultClientID,
			wantErr:  false,
		},
		{
			name:     "empty string (use default)",
			clientID: "",
			wantErr:  false,
		},
		{
			name:     "too short",
			clientID: "1234567890123456",
			wantErr:  true,
			wantCode: plexerrors.DISCORD_CLIENT_ID_INVALID,
		},
		{
			name:     "contains letters",
			clientID: "1234567890123456a",
			wantErr:  true,
			wantCode: plexerrors.DISCORD_CLIENT_ID_INVALID,
		},
		{
			name:     "contains special characters",
			clientID: "1234567890123456!",
			wantErr:  true,
			wantCode: plexerrors.DISCORD_CLIENT_ID_INVALID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClientID(tt.clientID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClientID(%q) error = %v, wantErr %v", tt.clientID, err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantCode != "" {
				appErr, ok := err.(*plexerrors.AppError)
				if !ok {
					t.Errorf("Expected AppError, got %T", err)
					return
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("ValidateClientID(%q) error code = %s, want %s", tt.clientID, appErr.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestBuildActivity(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		data     *PresenceData
		checkFn  func(t *testing.T, activity interface{})
	}{
		{
			name: "basic playing activity",
			data: &PresenceData{
				Track:  "Test Song",
				Artist: "Test Artist",
				Album:  "Test Album",
				State:  "playing",
			},
			checkFn: func(t *testing.T, activity interface{}) {
				// Type assertion not needed since buildActivity returns client.Activity
				// We check via the returned struct fields
			},
		},
		{
			name: "paused activity",
			data: &PresenceData{
				Track:  "Paused Song",
				Artist: "Paused Artist",
				State:  "paused",
			},
			checkFn: func(t *testing.T, activity interface{}) {
				// Paused state should set SmallImage to "pause"
			},
		},
		{
			name: "activity with timestamp",
			data: &PresenceData{
				Track:     "Timed Song",
				Artist:    "Timed Artist",
				State:     "playing",
				StartTime: &now,
			},
			checkFn: func(t *testing.T, activity interface{}) {
				// Should have timestamps set
			},
		},
		{
			name: "activity without album",
			data: &PresenceData{
				Track:  "Single Song",
				Artist: "Single Artist",
				State:  "playing",
			},
			checkFn: func(t *testing.T, activity interface{}) {
				// State should be "by Artist" without album
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := buildActivity(tt.data)

			// Check basic fields are set
			if activity.Details != tt.data.Track {
				t.Errorf("Expected Details=%q, got %q", tt.data.Track, activity.Details)
			}

			if activity.LargeImage != "plex-logo" {
				t.Errorf("Expected LargeImage='plex-logo', got %q", activity.LargeImage)
			}

			if activity.LargeText != "Plex" {
				t.Errorf("Expected LargeText='Plex', got %q", activity.LargeText)
			}

			// Check state line
			if tt.data.Artist != "" {
				if tt.data.Album != "" {
					expected := "by " + tt.data.Artist + " on " + tt.data.Album
					if activity.State != expected {
						t.Errorf("Expected State=%q, got %q", expected, activity.State)
					}
				} else {
					expected := "by " + tt.data.Artist
					if activity.State != expected {
						t.Errorf("Expected State=%q, got %q", expected, activity.State)
					}
				}
			}

			// Check small image based on playback state
			if tt.data.State == "paused" {
				if activity.SmallImage != "pause" {
					t.Errorf("Expected SmallImage='pause' for paused state, got %q", activity.SmallImage)
				}
				if activity.SmallText != "Paused" {
					t.Errorf("Expected SmallText='Paused' for paused state, got %q", activity.SmallText)
				}
			} else if tt.data.State == "playing" {
				if activity.SmallImage != "play" {
					t.Errorf("Expected SmallImage='play' for playing state, got %q", activity.SmallImage)
				}
				if activity.SmallText != "Playing" {
					t.Errorf("Expected SmallText='Playing' for playing state, got %q", activity.SmallText)
				}
			}

			// Check timestamps
			if tt.data.StartTime != nil && tt.data.State == "playing" {
				if activity.Timestamps == nil {
					t.Error("Expected Timestamps to be set for playing state with StartTime")
				} else if activity.Timestamps.Start == nil {
					t.Error("Expected Timestamps.Start to be set")
				}
			}
		})
	}
}

func TestGetCurrentPresence(t *testing.T) {
	pm := NewPresenceManager()

	// Initially should be nil
	if pm.GetCurrentPresence() != nil {
		t.Error("New PresenceManager should have nil presence")
	}

	// Set presence manually for testing
	testPresence := &PresenceData{
		Track:  "Test",
		Artist: "Artist",
	}
	pm.presence = testPresence

	if pm.GetCurrentPresence() != testPresence {
		t.Error("GetCurrentPresence should return the set presence")
	}
}

func TestSetPresenceNotConnected(t *testing.T) {
	pm := NewPresenceManager()

	err := pm.SetPresence(&PresenceData{
		Track:  "Test",
		Artist: "Artist",
	})

	if err == nil {
		t.Error("SetPresence should return error when not connected")
	}
}

func TestClearPresenceNotConnected(t *testing.T) {
	pm := NewPresenceManager()

	// Should not error when not connected (nothing to clear)
	err := pm.ClearPresence()
	if err != nil {
		t.Errorf("ClearPresence should not error when not connected, got: %v", err)
	}
}

func TestDisconnectNotConnected(t *testing.T) {
	pm := NewPresenceManager()

	// Should not error when not connected
	err := pm.Disconnect()
	if err != nil {
		t.Errorf("Disconnect should not error when not connected, got: %v", err)
	}
}

func TestConnectInvalidClientID(t *testing.T) {
	pm := NewPresenceManager()

	err := pm.Connect("invalid")
	if err == nil {
		t.Error("Connect should return error for invalid client ID")
	}
}

func TestMapDiscordError(t *testing.T) {
	if mapDiscordError(nil) != nil {
		t.Error("mapDiscordError(nil) should return nil")
	}
}

func TestMapDiscordErrorNotRunning(t *testing.T) {
	tests := []struct {
		name    string
		errMsg  string
		wantNil bool
	}{
		{
			name:    "connection refused",
			errMsg:  "dial unix: connection refused",
			wantNil: false,
		},
		{
			name:    "no such file (Unix socket)",
			errMsg:  "open /run/user/1000/discord-ipc-0: no such file or directory",
			wantNil: false,
		},
		{
			name:    "pipe error (Windows)",
			errMsg:  "open \\\\.\\pipe\\discord-ipc-0: The system cannot find the file specified",
			wantNil: false,
		},
		{
			name:    "generic pipe",
			errMsg:  "pipe connection failed",
			wantNil: false,
		},
		{
			name:    "unrelated error",
			errMsg:  "some other error",
			wantNil: false, // Still returns an error, just different code
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errMsg)
			result := mapDiscordError(err)
			if result == nil {
				t.Error("mapDiscordError should not return nil for non-nil input")
			}
		})
	}
}

func TestMapDiscordErrorCodes(t *testing.T) {
	// Test that specific error patterns map to DISCORD_NOT_RUNNING
	notRunningPatterns := []string{
		"connection refused",
		"no such file",
		"pipe",
	}

	for _, pattern := range notRunningPatterns {
		t.Run(pattern, func(t *testing.T) {
			err := fmt.Errorf("error: %s", pattern)
			result := mapDiscordError(err)
			if result == nil {
				t.Fatalf("mapDiscordError should return error for pattern %q", pattern)
			}

			// Check that it's an AppError with the right code
			appErr, ok := result.(*plexerrors.AppError)
			if !ok {
				t.Fatalf("Expected AppError, got %T", result)
			}

			if appErr.Code != plexerrors.DISCORD_NOT_RUNNING {
				t.Errorf("Expected code %s, got %s for pattern %q",
					plexerrors.DISCORD_NOT_RUNNING, appErr.Code, pattern)
			}
		})
	}
}

func TestIsConnectionLostErrorPatterns(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "nil error",
			errMsg:   "",
			expected: false,
		},
		{
			name:     "broken pipe",
			errMsg:   "write: broken pipe",
			expected: true,
		},
		{
			name:     "connection reset",
			errMsg:   "read: connection reset by peer",
			expected: true,
		},
		{
			name:     "EOF",
			errMsg:   "unexpected EOF",
			expected: true,
		},
		{
			name:     "unrelated error",
			errMsg:   "some other error",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = fmt.Errorf("%s", tt.errMsg)
			}
			result := isConnectionLostError(err)
			if result != tt.expected {
				t.Errorf("isConnectionLostError(%q) = %v, want %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

// Note: Integration tests for Connect, SetPresence, and ClearPresence
// require Discord to be running and are not included in unit tests.
// These should be tested manually or in integration test suites.
