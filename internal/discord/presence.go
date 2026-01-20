package discord

import (
	"log"
	"strings"
	"sync"
	"time"

	"plexcord/internal/errors"

	"github.com/hugolgst/rich-go/client"
)

// PresenceManager handles Discord Rich Presence updates.
// It manages the connection lifecycle and presence state.
type PresenceManager struct {
	mu        sync.RWMutex
	clientID  string
	connected bool
	presence  *PresenceData
}

// NewPresenceManager creates a new presence manager.
func NewPresenceManager() *PresenceManager {
	return &PresenceManager{
		clientID:  DefaultClientID,
		connected: false,
	}
}

// Connect establishes a connection to Discord using the provided Client ID.
// If clientID is empty, the default PlexCord Client ID is used.
// Returns an error if connection fails.
func (pm *PresenceManager) Connect(clientID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Use default if not provided
	if clientID == "" {
		clientID = DefaultClientID
	}

	// Validate client ID format (should be a numeric string)
	if !isValidClientID(clientID) {
		return errors.New(errors.DISCORD_CLIENT_ID_INVALID, "invalid Discord Client ID format")
	}

	// Already connected with same client ID
	if pm.connected && pm.clientID == clientID {
		log.Printf("Discord: Already connected with Client ID %s", clientID)
		return nil
	}

	// Disconnect existing connection if client ID changed
	if pm.connected && pm.clientID != clientID {
		log.Printf("Discord: Client ID changed, reconnecting...")
		client.Logout()
		pm.connected = false
	}

	log.Printf("Discord: Attempting to connect with Client ID %s", clientID)

	// Attempt to login to Discord
	err := client.Login(clientID)
	if err != nil {
		log.Printf("Discord: Connection failed: %v", err)
		return mapDiscordError(err)
	}

	pm.clientID = clientID
	pm.connected = true
	log.Printf("Discord: Successfully connected")
	return nil
}

// Disconnect closes the connection to Discord.
// It clears any active presence before disconnecting.
func (pm *PresenceManager) Disconnect() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.connected {
		return nil
	}

	log.Printf("Discord: Disconnecting...")

	// Clear presence before logout
	pm.presence = nil

	// Logout from Discord
	client.Logout()
	pm.connected = false

	log.Printf("Discord: Disconnected")
	return nil
}

// IsConnected returns whether a Discord connection is active.
func (pm *PresenceManager) IsConnected() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.connected
}

// GetClientID returns the current Discord Client ID.
func (pm *PresenceManager) GetClientID() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.clientID
}

// SetPresence updates the Discord Rich Presence with track information.
// Returns an error if not connected or if the update fails.
func (pm *PresenceManager) SetPresence(data *PresenceData) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.connected {
		return errors.New(errors.DISCORD_CONN_FAILED, "not connected to Discord")
	}

	// Build activity from presence data
	activity := buildActivity(data)

	err := client.SetActivity(activity)
	if err != nil {
		log.Printf("Discord: Failed to set presence: %v", err)
		// Check if connection was lost
		if isConnectionLostError(err) {
			pm.connected = false
			return errors.New(errors.DISCORD_NOT_RUNNING, "Discord connection lost")
		}
		return errors.Wrap(err, errors.DISCORD_CONN_FAILED, "failed to update presence")
	}

	pm.presence = data
	log.Printf("Discord: Presence updated - %s by %s", data.Track, data.Artist)
	return nil
}

// ClearPresence removes the Discord Rich Presence.
// Returns an error if not connected or if the clear fails.
func (pm *PresenceManager) ClearPresence() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.connected {
		return nil // Not connected, nothing to clear
	}

	// rich-go doesn't have a clear function, so we logout and re-login
	// This effectively clears the presence
	client.Logout()
	pm.presence = nil

	// Reconnect
	err := client.Login(pm.clientID)
	if err != nil {
		pm.connected = false
		log.Printf("Discord: Failed to reconnect after clearing presence: %v", err)
		return mapDiscordError(err)
	}

	log.Printf("Discord: Presence cleared")
	return nil
}

// GetCurrentPresence returns the current presence data, if any.
func (pm *PresenceManager) GetCurrentPresence() *PresenceData {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.presence
}

// buildActivity creates a rich-go Activity from PresenceData.
func buildActivity(data *PresenceData) client.Activity {
	activity := client.Activity{
		Details: data.Track,
	}

	// Build state line: "by Artist • Album" or just state
	if data.Artist != "" {
		if data.Album != "" {
			activity.State = "by " + data.Artist + " • " + data.Album
		} else {
			activity.State = "by " + data.Artist
		}
	}

	// Add playback state to state line if no artist
	if data.Artist == "" && data.State != "" {
		if data.State == "paused" {
			activity.State = "Paused"
		} else {
			activity.State = "Playing on Plex"
		}
	}

	// Set timestamps for elapsed time display (only when playing)
	if data.StartTime != nil && data.State == "playing" {
		activity.Timestamps = &client.Timestamps{
			Start: data.StartTime,
		}
	}

	// Use Plex logo for large image
	activity.LargeImage = "plex"
	activity.LargeText = "Plex Music"

	// Add small image to show playback state
	if data.State == "paused" {
		activity.SmallImage = "pause"
		activity.SmallText = "Paused"
	} else {
		activity.SmallImage = "play"
		activity.SmallText = "Playing"
	}

	return activity
}

// ValidateClientID checks if a Discord Client ID is valid for configuration.
// Returns nil if valid, or an error describing the validation failure.
// Special case: empty string is valid (means "use default Client ID").
func ValidateClientID(clientID string) error {
	// Empty string is valid - means use default
	if clientID == "" {
		return nil
	}

	// Must be at least 17 characters (Discord snowflake format)
	if len(clientID) < 17 {
		return errors.New(errors.DISCORD_CLIENT_ID_INVALID,
			"Discord Client ID must be at least 17 digits")
	}

	// Must be numeric only
	for _, c := range clientID {
		if c < '0' || c > '9' {
			return errors.New(errors.DISCORD_CLIENT_ID_INVALID,
				"Discord Client ID must contain only numbers")
		}
	}

	return nil
}

// isValidClientID checks if a Discord Client ID is valid.
// Discord Client IDs are numeric strings (snowflakes).
// For internal use - use ValidateClientID for user-facing validation.
func isValidClientID(clientID string) bool {
	if clientID == "" {
		return false
	}
	// Client ID should be numeric and at least 17 characters
	if len(clientID) < 17 {
		return false
	}
	for _, c := range clientID {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// mapDiscordError converts rich-go errors to PlexCord error codes.
func mapDiscordError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for common error patterns
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such file") ||
		strings.Contains(errStr, "pipe") {
		return errors.New(errors.DISCORD_NOT_RUNNING, "Discord is not running")
	}

	if strings.Contains(errStr, "invalid") {
		return errors.New(errors.DISCORD_CLIENT_ID_INVALID, "invalid Discord Client ID")
	}

	// Generic connection failure
	return errors.Wrap(err, errors.DISCORD_CONN_FAILED, "Discord connection failed")
}

// isConnectionLostError checks if an error indicates the Discord connection was lost.
func isConnectionLostError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "EOF")
}

// UpdatePresenceFromPlayback is a convenience method to update presence from playback data.
// It handles the conversion from Plex session data format to Discord presence format.
func (pm *PresenceManager) UpdatePresenceFromPlayback(track, artist, album, state string, duration, position int64) error {
	startTime := time.Now().Add(-time.Duration(position) * time.Millisecond)

	data := &PresenceData{
		Track:     track,
		Artist:    artist,
		Album:     album,
		State:     state,
		Duration:  duration,
		Position:  position,
		StartTime: &startTime,
	}

	return pm.SetPresence(data)
}
