package main

import (
	"context"
	"log"
	"sync"
	"time"

	"plexcord/internal/artwork"
	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/history"
	"plexcord/internal/platform"
	"plexcord/internal/retry"
	"plexcord/internal/updater"
	"plexcord/internal/version"
)

type App struct {
	ctx      context.Context
	config   *config.Config
	cfgStore *config.Store
	configs  ConfigGateway

	polling  *pollingController
	sessions *sessionCache

	discord *discordService

	plexFactory PlexAPIFactory

	discovery ServerDiscoverer

	tokens TokenStore

	autostart AutoStartController
	tray      TrayController

	desktop Desktop

	windows *windowManager

	presence *presenceGate

	trayIcons platform.TrayIcons

	plexRetry    RetryManager
	discordRetry RetryManager

	updater UpdateService

	relauncher AppRelauncher

	authFactory PlexAuthenticatorFactory
	plexAuth    PlexAuthenticator

	history HistoryStore

	bus events.Bus

	plexAuthMu sync.Mutex
}

const historyEntryLimit = 200

func (a *App) saveConfig() error {
	if a.cfgStore == nil {
		return errors.New(errors.CONFIG_WRITE_FAILED, "configuration store is not initialised")
	}
	return a.cfgStore.Update(func(*config.Config) {})
}

func NewApp() *App {
	return newAppWithDesktop(newDesktop())
}

func newAppWithDesktop(desktop Desktop) *App {
	a := &App{
		plexFactory:  newPlexClientFactory(),
		discovery:    newServerDiscoverer(),
		tokens:       newKeychainTokenStore(),
		configs:      newConfigGateway(),
		authFactory:  newPlexAuthenticatorFactory(),
		autostart:    platform.NewAutoStartManager(),
		plexRetry:    retry.NewManager("Plex"),
		discordRetry: retry.NewManager("Discord"),
		desktop:      desktop,
		polling:      &pollingController{},
		sessions:     &sessionCache{},
		relauncher:   newAppRelauncher(),
	}
	a.windows = newWindowManager(desktop, desktop)
	a.presence = newPresenceGate(a.clearDiscordOnStop, a.hideWhenPausedDelay)
	a.discord = a.newDiscordService(
		discord.NewPresenceManager(),
		artwork.NewResolver(artwork.WithUserAgent("PlexCord/"+version.Version)),
	)
	return a
}

func (a *App) newDiscordService(presence DiscordPresence, resolver ArtworkResolver) *discordService {
	return &discordService{
		presence:        presence,
		artwork:         resolver,
		options:         a.presenceDisplayOptions,
		artworkEnabled:  func() bool { return a.config != nil && a.config.ArtworkLookupEnabled() },
		defaultClientID: a.effectiveDiscordClientID,
		onConnected:     a.updateDiscordConnectionTime,
		isPaused:        func() bool { return a.presence.IsPaused() },
	}
}

func (a *App) hideWhenPausedDelay() time.Duration {
	if a.config == nil {
		return 0
	}
	return time.Duration(a.config.HideWhenPausedDelay) * time.Second
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.bus = events.NewWailsBus(ctx)

	if a.windows.MarkReady(ctx) {
		log.Printf("Replaying window restore requested before startup completed")
		a.ShowWindow()
	}

	a.adaptWindowToScreen(ctx)

	version.CaptureLaunchPath()

	cfg, err := a.configs.Load()
	if err != nil {
		log.Printf("Warning: failed to load config, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}
	a.config = cfg
	a.cfgStore = config.NewStore(cfg, config.Save)
	log.Printf("Configuration loaded successfully")

	if a.history == nil {
		a.history = history.NewStore(a.configs.ConfigDir(), historyEntryLimit)
	}

	if err := a.autostart.EnsureRegistered(); err != nil {
		log.Printf("Warning: failed to refresh auto-start registration: %v", err)
	}

	if !version.IsDevBuild() {
		if err := platform.SyncInstalledVersion(version.GetInfo().Version); err != nil {
			log.Printf("Warning: failed to refresh the installed version registration: %v", err)
		}
	}

	if a.tray == nil {
		a.tray = platform.NewTrayManager(platform.TrayCallbacks{
			OnShow:   a.ShowWindow,
			OnQuit:   a.QuitApp,
			OnUpdate: a.onTrayUpdateClick,
		}, a.trayIcons)
	}
	a.tray.Start()

	a.setupRetryCallbacks()

	if a.updater == nil {
		a.updater = updater.New(a.bus, 6*time.Hour)
	}
	a.updater.OnStatusChange(a.publishUpdateNotice)
	if a.config.IsAutoUpdateCheckEnabled() {
		a.updater.StartChecker(ctx)
	}

	token, err := a.tokens.Get()
	switch {
	case err != nil:
		log.Printf("Warning: failed to retrieve Plex token: %v", err)
	case token != "":
		log.Printf("Plex token retrieved successfully from secure storage")
	default:
		log.Printf("No Plex token found - user needs to complete setup")
	}

	if a.configs.IsSetupComplete() {
		go func() {
			time.Sleep(500 * time.Millisecond)

			log.Printf("Auto-connecting to Discord on startup...")
			err := a.ConnectDiscord("")
			if err != nil {
				log.Printf("Warning: Failed to auto-connect Discord: %v", err)
			}

			a.autoConnectPlex()
		}()
	}
}

func (a *App) domReady(ctx context.Context) {
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	if !a.windows.IsQuitting() && a.config != nil && a.config.MinimizeToTray {
		log.Printf("Close requested: hiding window, PlexCord keeps running in the background")
		a.desktop.Hide(ctx)
		return true
	}
	return false
}

func (a *App) shutdown(ctx context.Context) {
	a.plexRetry.Stop()
	a.discordRetry.Stop()

	if a.updater != nil {
		a.updater.StopChecker()
	}

	a.StopSessionPolling()

	if a.discord != nil {
		if err := a.discord.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect Discord: %v", err)
		}
	}

	if a.tray != nil {
		a.tray.Stop()
	}

	log.Printf("Application shutdown complete")
}
