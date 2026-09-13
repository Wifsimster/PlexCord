package discord

import (
	stderrors "errors"
	"log"
	"strings"
	"sync"

	"plexcord/internal/discord/ipc"
	"plexcord/internal/errors"
)

// PresenceManager handles Discord Rich Presence updates.
// It manages the connection lifecycle and presence state.
type PresenceManager struct {
	presence *PresenceData
	conn     Conn
	// dial constructs a fresh Conn for each login. Injected so the manager can
	// be driven without a running Discord (see WithDialer).
	dial      Dialer
	clientID  string
	mu        sync.RWMutex
	connected bool
}

// ManagerOption configures a PresenceManager at construction.
type ManagerOption func(*PresenceManager)

// WithDialer replaces the Discord IPC dialer, so tests (or an alternative
// transport) can supply their own Conn.
func WithDialer(d Dialer) ManagerOption {
	return func(pm *PresenceManager) {
		if d != nil {
			pm.dial = d
		}
	}
}

// NewPresenceManager creates a new presence manager. Without options it talks
// to the local Discord IPC socket.
func NewPresenceManager(opts ...ManagerOption) *PresenceManager {
	pm := &PresenceManager{
		clientID:  DefaultClientID,
		connected: false,
		dial:      defaultDialer,
	}
	for _, opt := range opts {
		opt(pm)
	}
	return pm
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
		if pm.conn != nil {
			if err := pm.conn.Close(); err != nil {
				log.Printf("Discord: error closing previous IPC connection: %v", err)
			}
			pm.conn = nil
		}
		pm.connected = false
	}

	log.Printf("Discord: Attempting to connect with Client ID %s", clientID)

	// Attempt to login to Discord over the injected connection.
	c := pm.dial()
	if err := c.Login(clientID); err != nil {
		log.Printf("Discord: Connection failed: %v", err)
		return mapDiscordError(err)
	}

	pm.conn = c
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

	// Close the IPC connection
	if pm.conn != nil {
		if err := pm.conn.Close(); err != nil {
			log.Printf("Discord: error closing IPC connection: %v", err)
		}
		pm.conn = nil
	}
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

	err := pm.conn.SetActivity(activity)
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

// ClearPresence removes the Discord Rich Presence without disconnecting.
// Sends an empty activity to clear the presence display while keeping the
// Discord IPC connection alive so subsequent presence updates work immediately.
// Returns nil if not connected (idempotent).
func (pm *PresenceManager) ClearPresence() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.connected {
		return nil // Not connected, nothing to clear
	}

	// Send an empty activity to clear the presence display.
	// This avoids the logout/login cycle that would briefly disconnect us
	// and risk leaving the manager in an inconsistent state.
	if err := pm.conn.SetActivity(ipc.Activity{}); err != nil {
		// If the upstream rejects the empty activity, log but don't disconnect.
		// The previous presence data will still be showing until the next update.
		log.Printf("Discord: Failed to clear presence (non-fatal): %v", err)
		return mapDiscordError(err)
	}

	pm.presence = nil
	log.Printf("Discord: Presence cleared")
	return nil
}

// GetCurrentPresence returns the current presence data, if any.
func (pm *PresenceManager) GetCurrentPresence() *PresenceData {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.presence
}

// buildActivity creates an ipc.Activity from PresenceData by dispatching
// to the appropriate PresenceBuilder for the data's MediaType. The actual
// formatting logic lives in builder.go — this function is kept as a thin
// alias to preserve the old call sites in presence.go.
func buildActivity(data *PresenceData) ipc.Activity {
	return buildActivityForMediaType(data)
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

// mapDiscordError converts IPC client errors to PlexCord error codes.
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
// It prefers the typed *ipc.ClosedError from the internal IPC client and falls
// back to string matching for transport-level errors (broken pipe, EOF, ...).
func isConnectionLostError(err error) bool {
	if err == nil {
		return false
	}
	var closed *ipc.ClosedError
	if stderrors.As(err, &closed) {
		return true
	}
	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "EOF")
}

// UpdatePlayback updates the presence from a playback snapshot and the user's
// display options. It is the method the application uses on every session
// change; SetPresence remains available for callers that have already built
// their own PresenceData (the connection test, for example).
func (pm *PresenceManager) UpdatePlayback(playback Playback, opts Options) error {
	return pm.SetPresence(playback.toPresenceData(opts))
}
