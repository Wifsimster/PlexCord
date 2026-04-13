package main

import (
	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/retry"
)

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
			a.bus.Emit(events.PlexRetryState, state)
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
			a.bus.Emit(events.DiscordRetryState, state)
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
