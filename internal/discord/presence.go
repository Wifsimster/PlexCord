package discord

import (
	stderrors "errors"
	"log"
	"net"
	"strings"
	"sync"

	"plexcord/internal/discord/ipc"
	"plexcord/internal/errors"
)

type PresenceManager struct {
	presence  *PresenceData
	conn      Conn
	dial      Dialer
	clientID  string
	mu        sync.RWMutex
	connected bool
}

type ManagerOption func(*PresenceManager)

func WithDialer(d Dialer) ManagerOption {
	return func(pm *PresenceManager) {
		if d != nil {
			pm.dial = d
		}
	}
}

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

func (pm *PresenceManager) Connect(clientID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if clientID == "" {
		clientID = DefaultClientID
	}

	if !isValidClientID(clientID) {
		return errors.New(errors.DISCORD_CLIENT_ID_INVALID, "invalid Discord Client ID format")
	}

	if pm.connected && pm.clientID == clientID {
		log.Printf("Discord: Already connected with Client ID %s", clientID)
		return nil
	}

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

func (pm *PresenceManager) Disconnect() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.connected {
		return nil
	}

	log.Printf("Discord: Disconnecting...")

	pm.presence = nil

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

func (pm *PresenceManager) IsConnected() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.connected
}

func (pm *PresenceManager) GetClientID() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.clientID
}

func (pm *PresenceManager) SetPresence(data *PresenceData) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.connected {
		return errors.New(errors.DISCORD_CONN_FAILED, "not connected to Discord")
	}

	activity := buildActivity(data)

	err := pm.conn.SetActivity(activity)
	if err != nil {
		log.Printf("Discord: Failed to set presence: %v", err)
		if isConnectionLostError(err) {
			pm.dropConnectionLocked()
			return errors.New(errors.DISCORD_NOT_RUNNING, "Discord connection lost")
		}
		return errors.Wrap(err, errors.DISCORD_CONN_FAILED, "failed to update presence")
	}

	pm.presence = data
	log.Printf("Discord: Presence updated - %s by %s", data.Track, data.Artist)
	return nil
}

func (pm *PresenceManager) ClearPresence() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.connected {
		return nil
	}

	if err := pm.conn.SetActivity(ipc.Activity{}); err != nil {
		log.Printf("Discord: Failed to clear presence (non-fatal): %v", err)
		if isConnectionLostError(err) {
			pm.dropConnectionLocked()
		}
		return mapDiscordError(err)
	}

	pm.presence = nil
	log.Printf("Discord: Presence cleared")
	return nil
}

func (pm *PresenceManager) GetCurrentPresence() *PresenceData {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.presence
}

func buildActivity(data *PresenceData) ipc.Activity {
	return buildActivityForMediaType(data)
}

func ValidateClientID(clientID string) error {
	if clientID == "" {
		return nil
	}

	if len(clientID) < 17 {
		return errors.New(errors.DISCORD_CLIENT_ID_INVALID,
			"Discord Client ID must be at least 17 digits")
	}

	for _, c := range clientID {
		if c < '0' || c > '9' {
			return errors.New(errors.DISCORD_CLIENT_ID_INVALID,
				"Discord Client ID must contain only numbers")
		}
	}

	return nil
}

func isValidClientID(clientID string) bool {
	if clientID == "" {
		return false
	}
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

func mapDiscordError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such file") ||
		strings.Contains(errStr, "pipe") {
		return errors.New(errors.DISCORD_NOT_RUNNING, "Discord is not running")
	}

	if strings.Contains(errStr, "invalid") {
		return errors.New(errors.DISCORD_CLIENT_ID_INVALID, "invalid Discord Client ID")
	}

	return errors.Wrap(err, errors.DISCORD_CONN_FAILED, "Discord connection failed")
}

func (pm *PresenceManager) dropConnectionLocked() {
	if pm.conn != nil {
		if err := pm.conn.Close(); err != nil {
			log.Printf("Discord: error closing lost IPC connection: %v", err)
		}
		pm.conn = nil
	}
	pm.connected = false
}

func isConnectionLostError(err error) bool {
	if err == nil {
		return false
	}
	var closed *ipc.ClosedError
	if stderrors.As(err, &closed) {
		return true
	}
	var netErr net.Error
	if stderrors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "EOF")
}

func (pm *PresenceManager) UpdatePlayback(playback Playback, opts Options) error {
	return pm.SetPresence(playback.toPresenceData(opts))
}
