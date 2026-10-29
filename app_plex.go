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

func (a *App) StartPlexPINAuth() (map[string]interface{}, error) {
	log.Println("Starting Plex PIN authentication")

	a.plexAuthMu.Lock()
	defer a.plexAuthMu.Unlock()

	a.plexAuth = a.authFactory()

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

		a.plexAuthMu.Lock()
		a.plexAuth = nil
		a.plexAuthMu.Unlock()

		return map[string]interface{}{
			"authorized": true,
			"authToken":  pinResp.AuthToken,
		}, nil
	}

	if time.Now().After(pinResp.ExpiresAt) {
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

func (a *App) DiscoverPlexServers() ([]plex.Server, error) {
	log.Printf("Starting Plex server discovery using GDM protocol...")

	servers, err := a.discovery.Discover(5 * time.Second)
	if err != nil {
		log.Printf("ERROR: Discovery failed: %v", err)
		return nil, err
	}

	log.Printf("Discovery complete: found %d server(s)", len(servers))
	return servers, nil
}

//nolint:unparam // ValidationResult is part of the public API even if not always used by callers
func (a *App) ValidatePlexConnection(serverURL string) (*plex.ValidationResult, error) {
	log.Printf("Validating Plex connection to: %s", serverURL)

	token, err := a.tokens.Get()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve token: %v", err)
		return nil, errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to retrieve token")
	}

	if token == "" {
		log.Printf("ERROR: No Plex token found")
		return nil, errors.New(errors.CONFIG_READ_FAILED, "plex token not found")
	}

	client := a.plexFactory(token, serverURL)
	result, err := client.ValidateConnection()
	if err != nil {
		log.Printf("ERROR: Connection validation failed: %v", err)
		return nil, err
	}

	log.Printf("Connection validated successfully: %s v%s (%d libraries)", result.ServerName, result.ServerVersion, result.LibraryCount)

	a.updatePlexConnectionTime()

	a.stopPlexRetry()

	return result, nil
}

func (a *App) GetPlexUsers(serverURL string) ([]plex.PlexUser, error) {
	log.Printf("Retrieving Plex users from: %s", serverURL)

	token, err := a.tokens.Get()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve token: %v", err)
		return nil, errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to retrieve token")
	}

	if token == "" {
		log.Printf("ERROR: No Plex token found")
		return nil, errors.New(errors.CONFIG_READ_FAILED, "plex token not found")
	}

	client := a.plexFactory(token, serverURL)
	users, err := client.GetUsers()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve users: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d user(s) from Plex server", len(users))
	return users, nil
}

func (a *App) SavePlexUserSelection(userID, userName string) error {
	log.Printf("Saving Plex user selection: ID=%s, Name=%s", userID, userName)

	if userID == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "user ID cannot be empty")
	}

	a.config.SelectedPlexUserID = userID
	a.config.SelectedPlexUserName = userName

	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save user selection: %v", err)
		return err
	}

	log.Printf("Plex user selection saved successfully")
	return nil
}

func (a *App) SaveServerURL(serverURL string) error {
	log.Printf("Saving Plex server URL: %s", serverURL)

	if serverURL == "" {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL cannot be empty")
	}

	parsed, err := url.Parse(serverURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New(errors.CONFIG_WRITE_FAILED, "server URL must use http or https scheme")
	}

	a.config.ServerURL = serverURL

	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save server URL: %v", err)
		return err
	}

	log.Printf("Plex server URL saved successfully")
	return nil
}

func (a *App) StartSessionPolling() error {
	serverURL := a.activePlexServerURL()
	if serverURL == "" {
		return errors.New(errors.CONFIG_READ_FAILED, "plex server URL not configured")
	}

	if a.config.SelectedPlexUserID == "" {
		return errors.New(errors.CONFIG_READ_FAILED, "plex user not selected")
	}

	token, err := a.tokens.Get()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve token for polling: %v", err)
		return errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to retrieve token")
	}

	if token == "" {
		return errors.New(errors.CONFIG_READ_FAILED, "plex token not found")
	}

	client, ok := a.plexFactory(token, serverURL).(plex.SessionSource)
	if !ok {
		return errors.New(errors.PLEX_CONN_FAILED, "plex client cannot supply sessions")
	}

	mediaTypes := a.config.EnabledMediaTypes()
	sessionCh := a.polling.Start(pollingConfig{
		Source:      client,
		UserID:      a.config.SelectedPlexUserID,
		MediaTypes:  mediaTypes,
		Interval:    a.pollingInterval(),
		OnError:     a.onPlexPollError,
		OnRecovered: a.onPlexPollRecovered,
	})
	if sessionCh == nil {
		return nil
	}

	log.Printf("Session polling started: user=%s, interval=%v, media=%v",
		a.config.SelectedPlexUserID, a.pollingInterval(), mediaTypes)

	go a.handleSessionUpdates(sessionCh)

	return nil
}

func (a *App) pollingInterval() time.Duration {
	interval := time.Duration(a.config.PollingInterval) * time.Second
	if interval < time.Second {
		return 2 * time.Second
	}
	return interval
}

func (a *App) onPlexPollError(err error) {
	log.Printf("Plex connection error detected, starting recovery...")

	a.clearDiscordOnStop()

	a.bus.Emit(events.PlexConnectionError, map[string]interface{}{
		"error":     err.Error(),
		"errorCode": errors.GetCode(err),
	})

	a.startPlexRetry(err)
}

func (a *App) onPlexPollRecovered() {
	log.Printf("Plex connection recovered")

	a.stopPlexRetry()
	a.updatePlexConnectionTime()

	a.bus.Emit(events.PlexConnectionRestored, nil)

	a.restorePresenceAfterOutage()
}

func (a *App) restorePresenceAfterOutage() {
	session := a.sessions.Get()
	if session == nil || a.presence.IsPaused() {
		return
	}
	if session.State == "paused" && a.config.HideWhenPaused {
		return
	}
	a.updateDiscordFromSession(session)
}

func (a *App) handleSessionUpdates(sessionCh <-chan *plex.MediaSession) {
	observers := []SessionObserver{
		newSessionCacheObserver(a.sessions),
		newHistoryObserver(a.history),
		&discordPresenceObserver{
			update:        a.updateDiscordFromSession,
			clearOnStop:   a.clearDiscordOnStop,
			isManualPause: a.presence.IsPaused,
			scheduleHide:  a.presence.ScheduleHide,
			cancelHide:    a.presence.CancelHide,
			hideOnPause:   func() bool { return a.config.HideWhenPaused },
			log:           log.Printf,
		},
		newEventEmitterObserver(a.bus),
	}
	runSessionPipeline(sessionCh, observers)
}

func (a *App) StopSessionPolling() {
	a.polling.Stop()
}

func (a *App) IsPollingActive() bool {
	return a.polling.IsRunning()
}

type PlexConnectionStatus struct {
	ServerURL    string `json:"serverUrl"`
	UserID       string `json:"userId"`
	UserName     string `json:"userName"`
	Connected    bool   `json:"connected"`
	Polling      bool   `json:"polling"`
	InErrorState bool   `json:"inErrorState"`
}

func (a *App) GetPlexConnectionStatus() PlexConnectionStatus {
	polling, inError := a.polling.State()

	return PlexConnectionStatus{
		ServerURL:    a.activePlexServerURL(),
		UserID:       a.config.SelectedPlexUserID,
		UserName:     a.config.SelectedPlexUserName,
		Polling:      polling,
		InErrorState: inError,
		Connected:    polling && !inError,
	}
}

func (a *App) SetPollingInterval(intervalSeconds int) error {
	if intervalSeconds < 1 {
		intervalSeconds = 1
	}
	if intervalSeconds > 60 {
		intervalSeconds = 60
	}

	a.config.PollingInterval = intervalSeconds
	if err := a.saveConfig(); err != nil {
		log.Printf("ERROR: Failed to save polling interval: %v", err)
		return err
	}

	if a.polling.SetInterval(time.Duration(intervalSeconds) * time.Second) {
		log.Printf("Polling interval updated to %d seconds", intervalSeconds)
	}

	return nil
}

func (a *App) GetPollingInterval() int {
	if a.config.PollingInterval < 1 {
		return 2
	}
	return a.config.PollingInterval
}

func (a *App) GetCurrentSession() *plex.MediaSession {
	return a.sessions.Get()
}

func (a *App) autoConnectPlex() {
	serverURL := a.activePlexServerURL()
	if serverURL == "" {
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
	if _, err := a.ValidatePlexConnection(serverURL); err != nil {
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

func (a *App) updatePlexConnectionTime() {
	now := time.Now()
	a.config.PlexLastConnected = &now
	if err := a.saveConfig(); err != nil {
		log.Printf("Warning: Failed to save Plex connection time: %v", err)
	}
}
