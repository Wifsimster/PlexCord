package main

import (
	"context"
	"log"
	goruntime "runtime"
	"sync"
	"time"

	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/errors"
	"plexcord/internal/keychain"
	"plexcord/internal/platform"
	"plexcord/internal/plex"
	"plexcord/internal/retry"
	"plexcord/internal/version"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	config *config.Config

	// Session polling
	poller         *plex.Poller
	pollerMu       sync.Mutex
	pollerCtx      context.Context
	pollerStop     context.CancelFunc
	currentSession *plex.MusicSession // Track current playback for page refresh restoration
	sessionMu      sync.RWMutex       // Protect currentSession access

	// Discord integration
	discord   *discord.PresenceManager
	discordMu sync.Mutex

	// Platform integration
	autostart *platform.AutoStartManager

	// Retry managers (Story 6.4)
	plexRetry    *retry.Manager
	discordRetry *retry.Manager

	// PIN authentication (maintain same client ID for PIN lifecycle)
	plexAuth   *plex.Authenticator
	plexAuthMu sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		discord:      discord.NewPresenceManager(),
		autostart:    platform.NewAutoStartManager(),
		plexRetry:    retry.NewManager("Plex"),
		discordRetry: retry.NewManager("Discord"),
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: failed to load config, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}
	a.config = cfg
	log.Printf("Configuration loaded successfully")

	// Setup retry callbacks for automatic reconnection
	a.setupRetryCallbacks()

	// Check if Plex token is available in keychain
	// The token will be used in later stories for Plex connection
	token, err := keychain.GetToken()
	switch {
	case err != nil:
		log.Printf("Warning: failed to retrieve Plex token: %v", err)
	case token != "":
		log.Printf("Plex token retrieved successfully from secure storage")
	default:
		log.Printf("No Plex token found - user needs to complete setup")
	}

	// Auto-connect to Discord and Plex if setup is complete
	if config.IsSetupComplete() {
		go func() {
			// Small delay to allow UI to initialize
			time.Sleep(500 * time.Millisecond)

			// Auto-connect to Discord
			log.Printf("Auto-connecting to Discord on startup...")
			err := a.ConnectDiscord("")
			if err != nil {
				log.Printf("Warning: Failed to auto-connect Discord: %v", err)
				// This is not critical - user can manually connect if needed
			}

			// Auto-connect to Plex and start session polling
			a.autoConnectPlex()
		}()
	}
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

	token, err := keychain.GetToken()
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

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Stop session polling if running
	a.StopSessionPolling()

	// Disconnect Discord
	if a.discord != nil {
		if err := a.discord.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect Discord: %v", err)
		}
	}

	log.Printf("Application shutdown complete")
}

// ============================================================================
// Window Management Methods (Story 4.5-4.10)
// ============================================================================

// ShowWindow shows and focuses the main application window.
// This is used when restoring from minimized/hidden state.
func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false) // Trick to bring to front
}

// HideWindow hides the main application window.
// The application continues running in the background.
func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

// MinimizeWindow minimizes the main application window.
func (a *App) MinimizeWindow() {
	runtime.WindowMinimise(a.ctx)
}

// QuitApp terminates the application completely.
// This is called from the tray menu or when the user explicitly quits.
func (a *App) QuitApp() {
	log.Printf("Quit requested")
	runtime.Quit(a.ctx)
}

// GetMinimizeToTray returns whether the app should minimize to tray.
func (a *App) GetMinimizeToTray() bool {
	return a.config.MinimizeToTray
}

// SetMinimizeToTray updates the minimize to tray setting.
func (a *App) SetMinimizeToTray(enabled bool) error {
	a.config.MinimizeToTray = enabled
	if err := config.Save(a.config); err != nil {
		log.Printf("ERROR: Failed to save minimize to tray setting: %v", err)
		return err
	}
	log.Printf("Minimize to tray set to: %v", enabled)
	return nil
}

// GetAutoStart returns whether auto-start on login is enabled.
// This checks the actual OS registration, not just the config value.
func (a *App) GetAutoStart() bool {
	return a.autostart.IsEnabled()
}

// SetAutoStart enables or disables auto-start on login.
// On Windows: Adds/removes from HKCU\Software\Microsoft\Windows\CurrentVersion\Run
// On macOS: Creates/removes LaunchAgent plist
// On Linux: Creates/removes XDG .desktop file in ~/.config/autostart/
func (a *App) SetAutoStart(enabled bool) error {
	// Update OS auto-start registration
	if err := a.autostart.SetEnabled(enabled); err != nil {
		log.Printf("ERROR: Failed to set auto-start: %v", err)
		return err
	}

	// Update config to match
	a.config.AutoStart = enabled
	if err := config.Save(a.config); err != nil {
		log.Printf("ERROR: Failed to save auto-start setting: %v", err)
		// Note: OS registration succeeded but config save failed
		// The actual auto-start behavior will work, but config may be out of sync
		return err
	}

	log.Printf("Auto-start set to: %v", enabled)
	return nil
}

// CheckSetupComplete checks if the setup wizard has been completed
// This method is called from the frontend router to determine if
// the user should be redirected to the setup wizard on first launch
func (a *App) CheckSetupComplete() bool {
	return config.IsSetupComplete()
}

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

	err := keychain.SetToken(token)
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
	token, err := keychain.GetToken()
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
func (a *App) ValidatePlexConnection(serverURL string) (*plex.ValidationResult, error) {
	log.Printf("Validating Plex connection to: %s", serverURL)

	// Retrieve token from keychain
	token, err := keychain.GetToken()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve token: %v", err)
		return nil, errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to retrieve token")
	}

	if token == "" {
		log.Printf("ERROR: No Plex token found")
		return nil, errors.New(errors.CONFIG_READ_FAILED, "plex token not found")
	}

	// Create Plex client and validate connection
	client := plex.NewClient(token, serverURL)
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
	token, err := keychain.GetToken()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve token: %v", err)
		return nil, errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to retrieve token")
	}

	if token == "" {
		log.Printf("ERROR: No Plex token found")
		return nil, errors.New(errors.CONFIG_READ_FAILED, "plex token not found")
	}

	// Create Plex client and get users
	client := plex.NewClient(token, serverURL)
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
	if err := config.Save(a.config); err != nil {
		log.Printf("ERROR: Failed to save user selection: %v", err)
		return err
	}

	log.Printf("Plex user selection saved successfully")
	return nil
}

// CompleteSetup marks the setup wizard as complete and saves the configuration.
// This method is called from the frontend when the user finishes the setup wizard.
// It marks SetupCompleted as true in config.json so that subsequent app launches
// go directly to the dashboard instead of the setup wizard.
// Also starts session polling if server and user are configured.
// Also ensures Discord connection is active.
func (a *App) CompleteSetup() error {
	log.Printf("Completing setup wizard...")

	// Mark setup as complete
	a.config.SetupCompleted = true

	// Save config to disk
	if err := config.Save(a.config); err != nil {
		log.Printf("ERROR: Failed to save setup completion: %v", err)
		return err
	}

	log.Printf("Setup wizard completed successfully")

	// Ensure Discord is connected for Rich Presence
	a.discordMu.Lock()
	if !a.discord.IsConnected() {
		log.Printf("Discord not connected during setup completion, attempting to connect...")
		clientID := a.config.DiscordClientID
		if clientID == "" {
			clientID = discord.DefaultClientID
		}
		if err := a.discord.Connect(clientID); err != nil {
			log.Printf("Warning: Failed to connect to Discord after setup: %v", err)
			// Don't fail setup - user might not have Discord running
		} else {
			a.updateDiscordConnectionTime()
			log.Printf("Discord connected successfully after setup completion")
		}
	}
	a.discordMu.Unlock()

	// Try to start session polling if configured
	// This is non-blocking - errors are logged but don't fail setup completion
	if a.config.ServerURL != "" && a.config.SelectedPlexUserID != "" {
		if err := a.StartSessionPolling(); err != nil {
			log.Printf("Warning: Failed to start session polling after setup: %v", err)
			// Don't return error - setup is still complete
		}
	}

	return nil
}

// SkipSetup marks the setup wizard as skipped and saves partial configuration.
// This method is called from the frontend when the user chooses to skip setup.
// The user can complete setup later from the Settings page.
// Partial progress (token, server URL) is preserved for later completion.
func (a *App) SkipSetup() error {
	log.Printf("Skipping setup wizard...")

	// Mark setup as skipped (not completed)
	a.config.SetupSkipped = true
	a.config.SetupCompleted = false

	// Save config to disk
	if err := config.Save(a.config); err != nil {
		log.Printf("ERROR: Failed to save setup skip: %v", err)
		return err
	}

	log.Printf("Setup wizard skipped - user can complete later from settings")
	return nil
}

// SaveServerURL saves the Plex server URL to configuration.
// This method is called from the setup wizard when a user selects a server.
func (a *App) SaveServerURL(serverURL string) error {
	log.Printf("Saving Plex server URL: %s", serverURL)

	if serverURL == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL cannot be empty")
	}

	// Update config with server URL
	a.config.ServerURL = serverURL

	// Save config to disk
	if err := config.Save(a.config); err != nil {
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
	token, err := keychain.GetToken()
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
			runtime.EventsEmit(a.ctx, "PlexConnectionError", map[string]interface{}{
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
			runtime.EventsEmit(a.ctx, "PlexConnectionRestored", nil)
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

// handleSessionUpdates processes session updates from the poller
// and emits appropriate Wails events to the frontend.
// Also updates Discord Rich Presence when connected.
func (a *App) handleSessionUpdates(sessionCh <-chan *plex.MusicSession) {
	var lastSession *plex.MusicSession

	for session := range sessionCh {
		if session != nil {
			// Music is playing - update Discord presence and emit frontend event
			log.Printf("Playback detected: %s - %s", session.Track, session.Artist)

			// Store current session for page refresh restoration
			a.sessionMu.Lock()
			a.currentSession = session
			a.sessionMu.Unlock()

			// Update Discord Rich Presence if connected
			a.updateDiscordFromSession(session)

			runtime.EventsEmit(a.ctx, "PlaybackUpdated", session)
			lastSession = session
		} else if lastSession != nil {
			// Music stopped - clear Discord presence and emit frontend event
			log.Printf("Playback stopped")

			// Clear current session
			a.sessionMu.Lock()
			a.currentSession = nil
			a.sessionMu.Unlock()

			// Clear Discord Rich Presence
			a.clearDiscordOnStop()

			runtime.EventsEmit(a.ctx, "PlaybackStopped", nil)
			lastSession = nil
		}
	}

	log.Printf("Session update handler exited")
}

// updateDiscordFromSession updates Discord Rich Presence with music session info.
// This is called automatically when playback is detected by the poller.
// If Discord is not connected, attempts to reconnect automatically (Story 3-8).
func (a *App) updateDiscordFromSession(session *plex.MusicSession) {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	// If not connected, try to reconnect (auto-recovery for Discord restart)
	if !a.discord.IsConnected() {
		a.tryDiscordReconnect()
		// If still not connected after reconnect attempt, skip update
		if !a.discord.IsConnected() {
			return
		}
		log.Printf("Discord: Reconnected - restoring presence")
	}

	// Update presence with session data
	err := a.discord.UpdatePresenceFromPlayback(
		session.Track,
		session.Artist,
		session.Album,
		session.State,
		session.Duration,
		session.ViewOffset,
	)
	if err != nil {
		log.Printf("Warning: Failed to update Discord presence: %v", err)
	}
}

// tryDiscordReconnect attempts to reconnect to Discord.
// This is called automatically when music is playing but Discord is not connected.
// Uses the configured or default Client ID for reconnection.
// Does not emit events - this is a silent background reconnection.
func (a *App) tryDiscordReconnect() {
	// Get the client ID to use
	clientID := a.config.DiscordClientID
	if clientID == "" {
		clientID = discord.DefaultClientID
	}

	// Attempt to reconnect (silently - no event emission)
	err := a.discord.Connect(clientID)
	if err != nil {
		// Failed to reconnect - Discord probably still not running
		// This is expected, don't log as error
		return
	}

	// Successfully reconnected - update connection history
	a.updateDiscordConnectionTime()
	log.Printf("Discord: Auto-reconnected to Discord")
}

// clearDiscordOnStop clears Discord Rich Presence when playback stops.
// This is called automatically when playback ends.
// If Discord is not connected, the clear is silently skipped.
func (a *App) clearDiscordOnStop() {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	if !a.discord.IsConnected() {
		// Not connected to Discord - nothing to clear
		return
	}

	err := a.discord.ClearPresence()
	if err != nil {
		log.Printf("Warning: Failed to clear Discord presence: %v", err)
	}
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
	Connected    bool   `json:"connected"`
	Polling      bool   `json:"polling"`
	InErrorState bool   `json:"inErrorState"`
	ServerURL    string `json:"serverUrl"`
	UserID       string `json:"userId"`
	UserName     string `json:"userName"`
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
	if err := config.Save(a.config); err != nil {
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

// ConnectDiscord establishes a connection to Discord using the provided Client ID.
// If clientID is empty, the default PlexCord Client ID is used.
// Returns an error if connection fails (e.g., Discord not running).
// Emits DiscordConnected or DiscordDisconnected Wails event based on result.
func (a *App) ConnectDiscord(clientID string) error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	log.Printf("Attempting Discord connection...")

	// Use configured client ID if not provided
	if clientID == "" {
		clientID = a.config.DiscordClientID
	}

	err := a.discord.Connect(clientID)
	if err != nil {
		log.Printf("ERROR: Discord connection failed: %v", err)
		// Emit disconnected event with error
		runtime.EventsEmit(a.ctx, "DiscordDisconnected", discord.ConnectionEvent{
			Connected: false,
			Error: &discord.Error{
				Code:    errors.GetCode(err),
				Message: err.Error(),
			},
		})
		return err
	}

	// Save client ID to config if connection successful
	if clientID != "" && clientID != a.config.DiscordClientID {
		a.config.DiscordClientID = clientID
		if err := config.Save(a.config); err != nil {
			log.Printf("Warning: Failed to save Discord client ID to config: %v", err)
		}
	}

	// Update connection history
	a.updateDiscordConnectionTime()

	// Stop any pending retries
	a.stopDiscordRetry()

	// Emit connected event
	runtime.EventsEmit(a.ctx, "DiscordConnected", discord.ConnectionEvent{
		Connected: true,
		ClientID:  a.discord.GetClientID(),
	})

	log.Printf("Discord connected successfully")
	return nil
}

// DisconnectDiscord closes the connection to Discord.
// Clears any active presence before disconnecting.
// Emits DiscordDisconnected Wails event.
func (a *App) DisconnectDiscord() error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	log.Printf("Disconnecting from Discord...")

	err := a.discord.Disconnect()
	if err != nil {
		log.Printf("ERROR: Discord disconnect failed: %v", err)
		return err
	}

	// Emit disconnected event
	runtime.EventsEmit(a.ctx, "DiscordDisconnected", discord.ConnectionEvent{
		Connected: false,
	})

	log.Printf("Discord disconnected")
	return nil
}

// IsDiscordConnected returns whether a Discord connection is active.
func (a *App) IsDiscordConnected() bool {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()
	return a.discord.IsConnected()
}

// GetDefaultDiscordClientID returns the default PlexCord Discord Application Client ID.
func (a *App) GetDefaultDiscordClientID() string {
	return discord.DefaultClientID
}

// GetDiscordClientID returns the currently configured Discord Client ID.
// Returns the config value if set, otherwise the default.
func (a *App) GetDiscordClientID() string {
	if a.config.DiscordClientID != "" {
		return a.config.DiscordClientID
	}
	return discord.DefaultClientID
}

// ValidateDiscordClientID validates a Discord Client ID format.
// Returns nil if valid, or an error with code DISCORD_CLIENT_ID_INVALID.
// Empty string is valid (means "use default Client ID").
func (a *App) ValidateDiscordClientID(clientID string) error {
	return discord.ValidateClientID(clientID)
}

// SaveDiscordClientID saves a custom Discord Client ID to configuration.
// Pass empty string to reset to default.
// Returns an error if the Client ID format is invalid.
func (a *App) SaveDiscordClientID(clientID string) error {
	// Validate before saving
	if err := discord.ValidateClientID(clientID); err != nil {
		log.Printf("ERROR: Invalid Discord client ID: %v", err)
		return err
	}

	a.config.DiscordClientID = clientID
	if err := config.Save(a.config); err != nil {
		log.Printf("ERROR: Failed to save Discord client ID: %v", err)
		return err
	}
	log.Printf("Discord Client ID saved")
	return nil
}

// UpdateDiscordPresence updates the Discord Rich Presence with current playback info.
// This is called internally when playback state changes.
func (a *App) UpdateDiscordPresence(track, artist, album, state string, duration, position int64) error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	if !a.discord.IsConnected() {
		return errors.New(errors.DISCORD_CONN_FAILED, "not connected to Discord")
	}

	return a.discord.UpdatePresenceFromPlayback(track, artist, album, state, duration, position)
}

// ClearDiscordPresence removes the Discord Rich Presence.
// Called when playback stops.
func (a *App) ClearDiscordPresence() error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	if !a.discord.IsConnected() {
		return nil // Not connected, nothing to clear
	}

	return a.discord.ClearPresence()
}

// TestDiscordPresence sends a test presence message to Discord to verify the connection.
// This displays a sample "Now Playing" message on the user's Discord profile.
// Returns an error if not connected or if the test fails.
func (a *App) TestDiscordPresence() error {
	a.discordMu.Lock()
	defer a.discordMu.Unlock()

	if !a.discord.IsConnected() {
		return errors.New(errors.DISCORD_CONN_FAILED, "not connected to Discord")
	}

	log.Printf("Sending test presence to Discord...")

	// Create test presence data
	testPresence := &discord.PresenceData{
		Track:  "Test Song - PlexCord",
		Artist: "PlexCord Test",
		Album:  "Connection Test",
		State:  "playing",
	}

	err := a.discord.SetPresence(testPresence)
	if err != nil {
		log.Printf("ERROR: Failed to send test presence: %v", err)
		return err
	}

	log.Printf("Test presence sent successfully")
	return nil
}

// ============================================================================
// Application Reset (Story 5.7)
// ============================================================================

// ResetApplication resets PlexCord to its initial state.
// This method:
// - Stops session polling
// - Disconnects from Discord
// - Removes Plex token from secure storage
// - Removes auto-start registration
// - Deletes the configuration file
// - Resets in-memory config to defaults
// After calling this method, the setup wizard will be shown on next app launch.
// The application does NOT exit - the user can continue or restart.
func (a *App) ResetApplication() {
	log.Printf("Resetting application to initial state...")

	// 1. Stop session polling
	a.StopSessionPolling()

	// 2. Disconnect from Discord (clears presence)
	a.discordMu.Lock()
	if a.discord.IsConnected() {
		if err := a.discord.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect Discord: %v", err)
		}
	}
	a.discordMu.Unlock()

	// 3. Remove Plex token from secure storage
	if err := keychain.DeleteToken(); err != nil {
		log.Printf("Warning: Failed to delete Plex token: %v", err)
		// Continue with reset - token deletion failure is not critical
	} else {
		log.Printf("Plex token removed from secure storage")
	}

	// 4. Remove auto-start registration
	if err := a.autostart.Disable(); err != nil {
		log.Printf("Warning: Failed to disable auto-start: %v", err)
		// Continue with reset - auto-start failure is not critical
	}

	// 5. Delete configuration file
	if err := config.Delete(); err != nil {
		log.Printf("Warning: Failed to delete config file: %v", err)
		// Continue with reset - we'll create fresh defaults
	} else {
		log.Printf("Configuration file deleted")
	}

	// 6. Reset in-memory config to defaults
	a.config = config.DefaultConfig()
	log.Printf("In-memory configuration reset to defaults")

	log.Printf("Application reset complete - setup wizard will show on next launch")
}

// ============================================================================
// Error Information (Story 6.2)
// ============================================================================

// GetErrorInfo returns user-friendly error information for an error code.
// This is used by the frontend to display actionable error messages.
func (a *App) GetErrorInfo(code string) errors.ErrorInfo {
	return errors.GetErrorInfo(code)
}

// IsRetryableError returns whether an error with the given code can be retried.
func (a *App) IsRetryableError(code string) bool {
	return errors.IsRetryable(code)
}

// IsAuthError returns whether the error indicates authentication is needed.
func (a *App) IsAuthError(code string) bool {
	return errors.IsAuthError(code)
}

// ============================================================================
// Connection History (Story 6.8)
// ============================================================================

// ConnectionHistory contains timestamps for last successful connections.
type ConnectionHistory struct {
	PlexLastConnected    *time.Time `json:"plexLastConnected"`
	DiscordLastConnected *time.Time `json:"discordLastConnected"`
}

// GetConnectionHistory returns the last successful connection timestamps.
func (a *App) GetConnectionHistory() ConnectionHistory {
	return ConnectionHistory{
		PlexLastConnected:    a.config.PlexLastConnected,
		DiscordLastConnected: a.config.DiscordLastConnected,
	}
}

// updatePlexConnectionTime updates the last successful Plex connection time.
// Called internally when Plex connection is successfully validated.
func (a *App) updatePlexConnectionTime() {
	now := time.Now()
	a.config.PlexLastConnected = &now
	if err := config.Save(a.config); err != nil {
		log.Printf("Warning: Failed to save Plex connection time: %v", err)
	}
}

// updateDiscordConnectionTime updates the last successful Discord connection time.
// Called internally when Discord connection is established.
func (a *App) updateDiscordConnectionTime() {
	now := time.Now()
	a.config.DiscordLastConnected = &now
	if err := config.Save(a.config); err != nil {
		log.Printf("Warning: Failed to save Discord connection time: %v", err)
	}
}

// ============================================================================
// Automatic Retry (Story 6.4)
// ============================================================================

// setupRetryCallbacks configures the retry managers with callbacks.
// Called during startup after ctx is set.
func (a *App) setupRetryCallbacks() {
	// Plex retry callback
	a.plexRetry.SetCallbacks(
		func() error {
			// Try to reconnect to Plex
			_, err := a.ValidatePlexConnection(a.config.ServerURL)
			if err != nil {
				return err
			}
			// Also restart polling if it was running
			return a.StartSessionPolling()
		},
		func(state retry.RetryState) {
			// Emit retry state change event
			runtime.EventsEmit(a.ctx, "PlexRetryState", state)
		},
	)

	// Discord retry callback
	a.discordRetry.SetCallbacks(
		func() error {
			// Try to reconnect to Discord
			return a.ConnectDiscord("")
		},
		func(state retry.RetryState) {
			// Emit retry state change event
			runtime.EventsEmit(a.ctx, "DiscordRetryState", state)
		},
	)
}

// GetPlexRetryState returns the current Plex retry state.
func (a *App) GetPlexRetryState() retry.RetryState {
	return a.plexRetry.GetState()
}

// GetDiscordRetryState returns the current Discord retry state.
func (a *App) GetDiscordRetryState() retry.RetryState {
	return a.discordRetry.GetState()
}

// RetryPlexConnection manually triggers a Plex connection retry.
// Resets the backoff schedule and attempts immediately.
func (a *App) RetryPlexConnection() {
	a.plexRetry.ManualRetry()
}

// RetryDiscordConnection manually triggers a Discord connection retry.
// Resets the backoff schedule and attempts immediately.
func (a *App) RetryDiscordConnection() {
	a.discordRetry.ManualRetry()
}

// startPlexRetry begins automatic retry for Plex connection failures.
func (a *App) startPlexRetry(err error) {
	code := errors.GetCode(err)
	// Only retry for connection errors, not auth errors
	if errors.IsAuthError(code) {
		return // Auth errors require user action
	}
	if code != "" && !errors.IsRetryable(code) {
		return
	}
	a.plexRetry.Start(err, code)
}

// stopPlexRetry stops automatic Plex retry on success.
func (a *App) stopPlexRetry() {
	a.plexRetry.Reset()
}

// stopDiscordRetry stops automatic Discord retry on success.
func (a *App) stopDiscordRetry() {
	a.discordRetry.Reset()
}

// ============================================================================
// Version & Updates (Epic 7)
// ============================================================================

// GetVersion returns the current application version information.
// Version is set at build time via -ldflags.
// Example: go build -ldflags "-X plexcord/internal/version.Version=v1.0.0"
func (a *App) GetVersion() version.Info {
	return version.GetInfo()
}

// CheckForUpdate checks GitHub releases for a newer version.
// Returns update info including availability, latest version, and download URL.
// A loading indicator should be shown during the check (typically 1-5 seconds).
func (a *App) CheckForUpdate() (*version.UpdateInfo, error) {
	log.Printf("Checking for updates...")

	info, err := version.CheckForUpdate()
	if err != nil {
		log.Printf("ERROR: Update check failed: %v", err)
		return nil, errors.Wrap(err, errors.TIMEOUT, "failed to check for updates")
	}

	if info.Available {
		log.Printf("Update available: %s -> %s", info.CurrentVersion, info.LatestVersion)
	} else {
		log.Printf("Application is up to date: %s", info.CurrentVersion)
	}

	return info, nil
}

// OpenReleasesPage opens the GitHub releases page in the default browser.
// Used for viewing changelog and downloading updates.
func (a *App) OpenReleasesPage() error {
	url := version.GetReleasesURL()
	log.Printf("Opening releases page: %s", url)
	runtime.BrowserOpenURL(a.ctx, url)
	return nil
}

// OpenReleaseURL opens a specific release URL in the default browser.
// Used when an update notification provides a direct link to the new release.
func (a *App) OpenReleaseURL(url string) error {
	if url == "" {
		return a.OpenReleasesPage()
	}
	log.Printf("Opening release URL: %s", url)
	runtime.BrowserOpenURL(a.ctx, url)
	return nil
}

// ============================================================================
// Resource Monitoring (Story 6.9)
// ============================================================================

// ResourceStats contains runtime statistics for monitoring long-running stability.
type ResourceStats struct {
	MemoryAllocMB  float64 `json:"memoryAllocMB"`
	MemoryTotalMB  float64 `json:"memoryTotalMB"`
	GoroutineCount int     `json:"goroutineCount"`
	Timestamp      string  `json:"timestamp"`
}

// GetResourceStats returns current resource usage statistics.
// This is useful for debugging memory leaks and verifying long-running stability.
// Primarily used for development and troubleshooting - not displayed to users normally.
func (a *App) GetResourceStats() ResourceStats {
	var m goruntime.MemStats
	goruntime.ReadMemStats(&m)

	return ResourceStats{
		MemoryAllocMB:  float64(m.Alloc) / 1024 / 1024,
		MemoryTotalMB:  float64(m.TotalAlloc) / 1024 / 1024,
		GoroutineCount: goruntime.NumGoroutine(),
		Timestamp:      time.Now().Format(time.RFC3339),
	}
}
