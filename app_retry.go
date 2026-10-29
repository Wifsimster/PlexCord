package main

import (
	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/retry"
)

func (a *App) setupRetryCallbacks() {
	a.plexRetry.SetCallbacks(
		func() error {
			serverURL := a.activePlexServerURL()
			if serverURL == "" || a.config.SelectedPlexUserID == "" {
				a.stopPlexRetry()
				return nil
			}

			if _, err := a.ValidatePlexConnection(serverURL); err != nil {
				code := errors.GetCode(err)
				if errors.IsAuthError(code) || (code != "" && !errors.IsRetryable(code)) {
					a.stopPlexRetry()
					return nil
				}
				return err
			}
			return a.StartSessionPolling()
		},
		func(state retry.RetryState) {
			a.bus.Emit(events.PlexRetryState, state)
		},
	)

	a.discordRetry.SetCallbacks(
		func() error {
			return a.ConnectDiscord("")
		},
		func(state retry.RetryState) {
			a.bus.Emit(events.DiscordRetryState, state)
		},
	)
}

func (a *App) GetPlexRetryState() retry.RetryState {
	return a.plexRetry.GetState()
}

func (a *App) GetDiscordRetryState() retry.RetryState {
	return a.discordRetry.GetState()
}

func (a *App) RetryPlexConnection() error {
	if a.activePlexServerURL() == "" || a.config.SelectedPlexUserID == "" {
		return errors.New(errors.PLEX_NOT_CONFIGURED, "No Plex server or user configured")
	}
	a.plexRetry.ManualRetry()
	return nil
}

func (a *App) RetryDiscordConnection() {
	a.discordRetry.ManualRetry()
}

func (a *App) startPlexRetry(err error) {
	code := errors.GetCode(err)
	if errors.IsAuthError(code) {
		return
	}
	if code != "" && !errors.IsRetryable(code) {
		return
	}
	a.plexRetry.Start(err, code)
}

func (a *App) stopPlexRetry() {
	a.plexRetry.Reset()
}

func (a *App) stopDiscordRetry() {
	a.discordRetry.Reset()
}
