package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/plex"
)

// StartPlexPINAuth initiates PIN-based authentication with Plex.
// Returns a PIN code that the user must enter at plex.tv/link and the auth URL.
// This is the easiest and most secure way for users to authenticate.
func (a *App) StartPlexPINAuth() (map[string]interface{}, error) {
	log.Println("Starting Plex PIN authentication")

	a.plexAuthMu.Lock()
	defer a.plexAuthMu.Unlock()

	// Create a new authenticator for this PIN request
	// This ensures we get a fresh client ID for this authentication session
	a.plexAuth = plex.NewAuthenticator()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pinResp, err := a.plexAuth.RequestPIN(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to request PIN: %v", err)
		return nil, err
	}

	authURL := a.plexAuth.GetAuthURL(pinResp.Code)

	log.Printf("PIN generated: %s (ID: %d, expires in %d seconds)", pinResp.Code, pinResp.ID, pinResp.ExpiresIn)

	return map[string]interface{}{
		"pinCode":   pinResp.Code,
		"pinID":     pinResp.ID,
		"authURL":   authURL,
		"expiresIn": pinResp.ExpiresIn,
		"expiresAt": pinResp.ExpiresAt.Format(time.RFC3339),
	}, nil
}

// CheckPlexPINAuth checks if the user has authorized the PIN.
// Returns the auth token if authorized, or an error if still pending or expired.
func (a *App) CheckPlexPINAuth(pinID int) (map[string]interface{}, error) {
	a.plexAuthMu.Lock()
	auth := a.plexAuth
	a.plexAuthMu.Unlock()

	if auth == nil {
		return nil, errors.New(errors.PLEX_CONN_FAILED, "PIN authentication not started")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pinResp, err := auth.CheckPIN(ctx, pinID)
	if err != nil {
		return nil, err
	}

	if pinResp.AuthToken != "" {
		log.Printf("PIN authorized successfully, token received")

		// Clear the authenticator after successful auth
		a.plexAuthMu.Lock()
		a.plexAuth = nil
		a.plexAuthMu.Unlock()

		return map[string]interface{}{
			"authorized": true,
			"authToken":  pinResp.AuthToken,
		}, nil
	}

	// Check if expired
	if time.Now().After(pinResp.ExpiresAt) {
		// Clear the authenticator after expiration
		a.plexAuthMu.Lock()
		a.plexAuth = nil
		a.plexAuthMu.Unlock()

		return map[string]interface{}{
			"authorized": false,
			"expired":    true,
		}, nil
	}

	return map[string]interface{}{
		"authorized": false,
		"expired":    false,
	}, nil
}

// SavePlexToken stores the Plex token securely in OS keychain.
// This method is called from the setup wizard after the user enters their token.
// The token is stored using platform-specific secure storage:
// - Windows: Credential Manager
// - macOS: Keychain Access
// - Linux: Secret Service API
// If OS keychain is unavailable, it falls back to encrypted file storage.
func (a *App) SavePlexToken(token string) error {
	if token == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "token cannot be empty")
	}

	err := a.tokens.Set(token)
	if err != nil {
		log.Printf("ERROR: Failed to store Plex token: %v", err)
		return err
	}

	log.Printf("Plex token stored securely")
	return nil
}

// GetPlexToken retrieves the Plex token from OS keychain.
// This method is called during application startup to authenticate with Plex.
// Returns an error if the token cannot be retrieved or if no token has been set.
func (a *App) GetPlexToken() (string, error) {
	token, err := a.tokens.Get()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve Plex token: %v", err)
		return "", err
	}

	if token == "" {
		return "", errors.New(errors.CONFIG_READ_FAILED, "plex token not found")
	}

	log.Printf("Plex token retrieved successfully")
	return token, nil
}

// DiscoverPlexServers scans the local network for Plex servers using GDM protocol.
// This method uses Plex's Good Day Mate (GDM) multicast protocol to discover servers
// on the same subnet. Discovery completes within 5 seconds maximum.
//
// Returns a slice of discovered servers with their name, address, port, and local/remote status.
// An empty slice is returned if no servers are found (not an error condition).
func (a *App) DiscoverPlexServers() ([]plex.Server, error) {
	log.Printf("Starting Plex server discovery using GDM protocol...")

	// Discover with 5 second timeout (as per AC1)
	servers, err := plex.DiscoverServers(5 * time.Second)
	if err != nil {
		log.Printf("ERROR: Discovery failed: %v", err)
		return nil, err
	}

	log.Printf("Discovery complete: found %d server(s)", len(servers))
	return servers, nil
}

// ValidatePlexConnection validates the connection to a Plex server using the stored token.
// This method is called from the setup wizard to verify that:
// - The Plex server is reachable at the given URL
// - The stored token has valid authentication
// - The server has accessible libraries
//
// Returns validation details or an error with appropriate code.
// Validation completes within 5 seconds maximum with automatic timeout.
//
//nolint:unparam // ValidationResult is part of the public API even if not always used by callers
func (a *App) ValidatePlexConnection(serverURL string) (*plex.ValidationResult, error) {
	log.Printf("Validating Plex connection to: %s", serverURL)

	// Retrieve token from keychain
	token, err := a.tokens.Get()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve token: %v", err)
		return nil, errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to retrieve token")
	}

	if token == "" {
		log.Printf("ERROR: No Plex token found")
		return nil, errors.New(errors.CONFIG_READ_FAILED, "plex token not found")
	}

	// Create Plex client and validate connection
	client := a.plexFactory(token, serverURL)
	result, err := client.ValidateConnection()
	if err != nil {
		log.Printf("ERROR: Connection validation failed: %v", err)
		return nil, err
	}

	log.Printf("Connection validated successfully: %s v%s (%d libraries)", result.ServerName, result.ServerVersion, result.LibraryCount)

	// Update connection history
	a.updatePlexConnectionTime()

	// Stop any pending retries
	a.stopPlexRetry()

	return result, nil
}

// GetPlexUsers retrieves the list of Plex user accounts from the server.
// This method is called from the setup wizard to display available users
// that can be monitored for playback activity.
//
// Returns a slice of users with their ID, name, and avatar thumbnail URL.
// The server URL must be configured in config before calling this method.
// Retrieval completes within 5 seconds maximum with automatic timeout.
func (a *App) GetPlexUsers(serverURL string) ([]plex.PlexUser, error) {
	log.Printf("Retrieving Plex users from: %s", serverURL)

	// Retrieve token from keychain
	token, err := a.tokens.Get()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve token: %v", err)
		return nil, errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to retrieve token")
	}

	if token == "" {
		log.Printf("ERROR: No Plex token found")
		return nil, errors.New(errors.CONFIG_READ_FAILED, "plex token not found")
	}

	// Create Plex client and get users
	client := a.plexFactory(token, serverURL)
	users, err := client.GetUsers()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve users: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d user(s) from Plex server", len(users))
	return users, nil
}

// SavePlexUserSelection saves the selected Plex user to configuration.
// This method is called from the setup wizard when a user selects which
// Plex account to monitor for playback activity.
//
// The selection persists across application restarts via config.json.
func (a *App) SavePlexUserSelection(userID, userName string) error {
	log.Printf("Saving Plex user selection: ID=%s, Name=%s", userID, userName)

	if userID == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "user ID cannot be empty")
	}

	// Update config with selected user
	a.config.SelectedPlexUserID = userID
	a.config.SelectedPlexUserName = userName

	// Save config to disk
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save user selection: %v", err)
		return err
	}

	log.Printf("Plex user selection saved successfully")
	return nil
}

// SaveServerURL saves the Plex server URL to configuration.
// This method is called from the setup wizard when a user selects a server.
func (a *App) SaveServerURL(serverURL string) error {
	log.Printf("Saving Plex server URL: %s", serverURL)

	if serverURL == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL cannot be empty")
	}

	// Validate URL scheme to prevent SSRF/token exfiltration
	parsed, err := url.Parse(serverURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL must use http or https scheme")
	}

	// Update config with server URL
	a.config.ServerURL = serverURL

	// Save config to disk
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save server URL: %v", err)
		return err
	}

	log.Printf("Plex server URL saved successfully")
	return nil
}

// StartSessionPolling begins polling the Plex server for music sessions.
// This method creates a background poller that periodically checks for active
// music playback and emits Wails events when the session state changes:
//   - PlaybackUpdated: Emitted when music is playing or track changes
//   - PlaybackStopped: Emitted when music playback stops
//
// Polling uses the configured interval (default 2 seconds per NFR4).
// Each poll completes within 500ms (NFR5).
// Returns an error if required configuration is missing.
func (a *App) StartSessionPolling() error {
	a.pollerMu.Lock()
	defer a.pollerMu.Unlock()

	// Check if already polling
	if a.poller != nil && a.poller.IsRunning() {
		log.Printf("Session polling already running")
		return nil
	}

	// Validate configuration
	if a.config.ServerURL == "" {
		return errors.New(errors.CONFIG_READ_FAILED, "plex server URL not configured")
	}

	if a.config.SelectedPlexUserID == "" {
		return errors.New(errors.CONFIG_READ_FAILED, "plex user not selected")
	}

	// Retrieve token from keychain
	token, err := a.tokens.Get()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve token for polling: %v", err)
		return errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to retrieve token")
	}

	if token == "" {
		return errors.New(errors.CONFIG_READ_FAILED, "plex token not found")
	}

	// Create Plex client
	client := plex.NewClient(token, a.config.ServerURL)

	// Get polling interval from config (default 2 seconds for NFR4 compliance)
	interval := time.Duration(a.config.PollingInterval) * time.Second
	if interval < time.Second {
		interval = 2 * time.Second // Default to 2 seconds per NFR4: state changes detected within 2s
	}

	// Create poller
	a.poller = plex.NewPoller(client, a.config.SelectedPlexUserID, interval)

	// Setup error callbacks for graceful Plex unavailability handling (Story 6.5)
	a.poller.SetErrorCallbacks(
		// onError: Called when Plex connection fails
		func(err error) {
			log.Printf("Plex connection error detected, starting recovery...")

			// Clear Discord presence so it doesn't show stale data
			a.clearDiscordOnStop()

			// Emit event for frontend to show error status
			a.bus.Emit(events.PlexConnectionError, map[string]interface{}{
				"error":     err.Error(),
				"errorCode": errors.GetCode(err),
			})

			// Start automatic retry (backoff is handled by retry manager)
			a.startPlexRetry(err)
		},
		// onRecovered: Called when Plex connection recovers
		func() {
			log.Printf("Plex connection recovered")

			// Stop retry and update connection history
			a.stopPlexRetry()
			a.updatePlexConnectionTime()

			// Emit event for frontend to clear error status
			a.bus.Emit(events.PlexConnectionRestored, nil)
		},
	)

	// Create context for poller
	a.pollerCtx, a.pollerStop = context.WithCancel(context.Background())

	// Start polling
	sessionCh := a.poller.Start(a.pollerCtx)

	log.Printf("Session polling started: user=%s, interval=%v", a.config.SelectedPlexUserID, interval)

	// Start goroutine to handle session updates
	go a.handleSessionUpdates(sessionCh)

	return nil
}

// handleSessionUpdates constructs the observer pipeline and runs it.
// The actual event handling is delegated to individual observers for
// separation of concerns; see app_observers.go.
//
// Observer order matters: cache → history → discord → events. The
// discord observer is gated by the manual-pause flag and the
// hide-when-paused config, and the event emitter always fires last so
// the frontend sees the state after all side effects have run.
func (a *App) handleSessionUpdates(sessionCh <-chan *plex.MusicSession) {
	observers := []SessionObserver{
		newSessionCacheObserver(&a.sessionMu, &a.currentSession),
		newHistoryObserver(a.history),
		&discordPresenceObserver{
			update:        a.updateDiscordFromSession,
			clearOnStop:   a.clearDiscordOnStop,
			isManualPause: a.isPresencePausedLocked,
			scheduleHide:  a.scheduleHideOnPause,
			cancelHide:    a.cancelPauseTimer,
			hideOnPause:   func() bool { return a.config.HideWhenPaused },
			log:           log.Printf,
		},
		newEventEmitterObserver(a.bus),
	}
	runSessionPipeline(sessionCh, observers)
}

// isPresencePausedLocked returns the current manual pause state under lock.
func (a *App) isPresencePausedLocked() bool {
	a.pauseMu.Lock()
	defer a.pauseMu.Unlock()
	return a.presencePaused
}

// StopSessionPolling stops the background session polling.
// This method is called during application shutdown or when the user
// wants to temporarily stop monitoring playback.
// It is safe to call this method even if polling is not running.
func (a *App) StopSessionPolling() {
	a.pollerMu.Lock()
	defer a.pollerMu.Unlock()

	if a.poller == nil {
		return
	}

	// Cancel context and stop poller
	if a.pollerStop != nil {
		a.pollerStop()
	}

	a.poller.Stop()
	a.poller = nil
	a.pollerCtx = nil
	a.pollerStop = nil

	log.Printf("Session polling stopped")
}

// IsPollingActive returns whether session polling is currently running.
func (a *App) IsPollingActive() bool {
	a.pollerMu.Lock()
	defer a.pollerMu.Unlock()

	return a.poller != nil && a.poller.IsRunning()
}

// PlexConnectionStatus represents the current Plex connection status.
type PlexConnectionStatus struct {
	ServerURL    string `json:"serverUrl"`
	UserID       string `json:"userId"`
	UserName     string `json:"userName"`
	Connected    bool   `json:"connected"`
	Polling      bool   `json:"polling"`
	InErrorState bool   `json:"inErrorState"`
}

// GetPlexConnectionStatus returns the current Plex connection status (Story 6.5).
func (a *App) GetPlexConnectionStatus() PlexConnectionStatus {
	a.pollerMu.Lock()
	defer a.pollerMu.Unlock()

	status := PlexConnectionStatus{
		ServerURL: a.config.ServerURL,
		UserID:    a.config.SelectedPlexUserID,
		UserName:  a.config.SelectedPlexUserName,
	}

	if a.poller != nil {
		status.Polling = a.poller.IsRunning()
		status.InErrorState = a.poller.IsInErrorState()
		status.Connected = status.Polling && !status.InErrorState
	}

	return status
}

// SetPollingInterval updates the session polling interval dynamically.
// The interval is clamped to min 1 second, max 60 seconds (AC3).
// Changes take effect on the next polling cycle without restart.
func (a *App) SetPollingInterval(intervalSeconds int) error {
	// Validate bounds
	if intervalSeconds < 1 {
		intervalSeconds = 1
	}
	if intervalSeconds > 60 {
		intervalSeconds = 60
	}

	// Update config
	a.config.PollingInterval = intervalSeconds
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save polling interval: %v", err)
		return err
	}

	// Update running poller if active
	a.pollerMu.Lock()
	if a.poller != nil && a.poller.IsRunning() {
		a.poller.SetInterval(time.Duration(intervalSeconds) * time.Second)
		log.Printf("Polling interval updated to %d seconds", intervalSeconds)
	}
	a.pollerMu.Unlock()

	return nil
}

// GetPollingInterval returns the current polling interval in seconds.
// Returns the configured value, or 2 seconds as default per NFR4 requirement.
func (a *App) GetPollingInterval() int {
	if a.config.PollingInterval < 1 {
		return 2 // Default per NFR4: state changes detected within 2 seconds
	}
	return a.config.PollingInterval
}

// GetCurrentSession returns the current music session if music is playing.
// Returns nil if no music is currently playing.
// This is used by the frontend to restore playback state after page refresh.
func (a *App) GetCurrentSession() *plex.MusicSession {
	a.sessionMu.RLock()
	defer a.sessionMu.RUnlock()
	return a.currentSession
}

// autoConnectPlex attempts to restore the Plex connection using persisted config.
// This mirrors Discord auto-connect behavior by validating and restarting polling.
func (a *App) autoConnectPlex() {
	if a.config.ServerURL == "" {
		log.Printf("Auto-connect skipped: Plex server URL not configured")
		return
	}
	if a.config.SelectedPlexUserID == "" {
		log.Printf("Auto-connect skipped: Plex user not selected")
		return
	}

	token, err := a.tokens.Get()
	if err != nil {
		log.Printf("Warning: Failed to retrieve Plex token on startup: %v", err)
		a.startPlexRetry(err)
		return
	}
	if token == "" {
		log.Printf("Warning: No Plex token found on startup")
		return
	}

	log.Printf("Auto-connecting to Plex on startup...")
	if _, err := a.ValidatePlexConnection(a.config.ServerURL); err != nil {
		log.Printf("Warning: Failed to validate Plex connection on startup: %v", err)
		a.startPlexRetry(err)
		return
	}

	if err := a.StartSessionPolling(); err != nil {
		log.Printf("Warning: Failed to start session polling on startup: %v", err)
		a.startPlexRetry(err)
		return
	}

	log.Printf("Plex connection restored and session polling started")
}

// updatePlexConnectionTime updates the last successful Plex connection time.
// Called internally when Plex connection is successfully validated.
func (a *App) updatePlexConnectionTime() {
	now := time.Now()
	a.config.PlexLastConnected = &now
	if err := a.saveConfig(); err != nil {
		log.Printf("Warning: Failed to save Plex connection time: %v", err)
	}
}
