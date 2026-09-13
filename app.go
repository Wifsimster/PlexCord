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

// App is the Wails binding surface: the set of methods the frontend can call.
// Its single job is to translate those calls into work on the collaborators
// below and to marshal the results back — the behaviour itself lives in the
// internal packages and in the small helpers alongside this file
// (windowManager, presenceGate, the session observers).
//
// Every collaborator is held as an interface so the whole surface can be
// exercised with fakes: no Plex server, no Discord socket, no keychain, no
// system tray, no Wails window.
type App struct {
	ctx context.Context
	// config holds a pointer to the current Config for direct reads; all
	// writes go through cfgStore which handles atomic mutation + persistence.
	config   *config.Config
	cfgStore *config.Store
	configs  ConfigGateway // loading, deleting, and locating the config file

	// polling owns the session poller's lifecycle; sessions caches what is
	// playing so the frontend can restore its dashboard after a page refresh.
	polling  *pollingController
	sessions *sessionCache

	// discord owns the Discord link and everything serialized against it:
	// connecting, publishing presence, and the artwork generation guard.
	discord *discordService

	// Plex client factory for constructing clients on demand (per-server).
	// Using a factory instead of a singleton reflects that the server URL/token
	// can change at runtime and enables tests to inject fakes.
	plexFactory PlexAPIFactory

	// discovery finds Plex servers on the local network (GDM in production).
	discovery ServerDiscoverer

	// Token store abstracts credential persistence (OS keychain in production)
	tokens TokenStore

	// Platform integration
	autostart AutoStartController
	tray      TrayController

	// desktop is the Wails runtime: window, quit, browser, screens. Held as
	// the union so the narrow pieces can be handed to their consumers.
	desktop Desktop

	// windows owns window visibility, the deferred-restore dance, and the
	// explicit-quit flag.
	windows *windowManager

	// presence gates Discord updates behind the manual pause toggle and the
	// hide-when-paused timer.
	presence *presenceGate

	// Tray icon data, injected from main so the platform layer stays
	// asset-agnostic. iconPNG is used on macOS/Linux, iconICO on Windows.
	trayIconPNG       []byte
	trayIconICO       []byte
	trayIconUpdatePNG []byte // badged variants, shown while an update is pending
	trayIconUpdateICO []byte

	// Retry managers (Story 6.4)
	plexRetry    RetryManager
	discordRetry RetryManager

	// Automatic update checker (constructed in startup — it needs the bus)
	updater UpdateService

	// relauncher spawns the updated binary when applying an update.
	relauncher AppRelauncher

	// PIN authentication: authFactory builds one authenticator per PIN request
	// (each gets its own client ID); plexAuth is the one in flight.
	authFactory PlexAuthenticatorFactory
	plexAuth    PlexAuthenticator

	// Listening history
	history HistoryStore

	// Event bus for emitting events to the frontend (abstracts Wails runtime)
	bus events.Bus

	// plexAuthMu guards the PIN authenticator across the request/poll cycle.
	plexAuthMu sync.Mutex
}

// historyEntryLimit is how many listening-history entries are retained; older
// ones are trimmed.
const historyEntryLimit = 200

// saveConfig persists the current in-memory config via the ConfigStore.
// This is the single path for all config writes — callers that need to
// mutate the config should set fields on a.config then call this method.
// Future enhancements (debouncing, schema migration, atomic writes) can
// hook in here without changing call sites.
func (a *App) saveConfig() error {
	if a.cfgStore == nil {
		// Every production path runs after startup, which builds the store.
		// Failing loudly beats the old fallback, which wrote straight to the
		// real config file and so behaved differently from every other save.
		return errors.New(errors.CONFIG_WRITE_FAILED, "configuration store is not initialised")
	}
	// No-op mutator: the caller already updated a.config directly; this
	// just triggers the store's atomic save path.
	return a.cfgStore.Update(func(*config.Config) {})
}

// NewApp creates a new App application struct with production dependencies.
// For tests, construct an App directly with injected fakes for bus,
// plexFactory, tokens, discord, desktop, tray, autostart and updater.
func NewApp() *App {
	return newAppWithDesktop(newDesktop())
}

// newAppWithDesktop builds an App over the given runtime. Split from NewApp so
// tests can drive the full window/quit surface against a fake Desktop.
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

// newDiscordService wires a discordService to the parts of App it needs to
// consult — the persisted presence options, the artwork-lookup toggle, the
// configured client ID, the connection history, and the pause state. Passing
// them as functions keeps the service from holding a reference back to App.
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

// hideWhenPausedDelay reports the configured hide-when-paused delay. It is the
// presence gate's window onto the config, so the gate never reads it directly.
func (a *App) hideWhenPausedDelay() time.Duration {
	if a.config == nil {
		return 0
	}
	return time.Duration(a.config.HideWhenPausedDelay) * time.Second
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
	a.bus = events.NewWailsBus(ctx)

	// The window can be driven from here on. A restore request that came in
	// earlier — PlexCord started minimized and the user relaunched it while it
	// was still booting — was parked rather than run against a nil context, so
	// replay it now.
	if a.windows.MarkReady(ctx) {
		log.Printf("Replaying window restore requested before startup completed")
		a.ShowWindow()
	}

	// The window opened at the content-derived preferred size, which Wails had
	// to take before it could report a screen. Now that it can, shrink the
	// window if that size does not fit the display it opened on.
	a.adaptWindowToScreen(ctx)

	// Capture the executable's launch path now, while the running binary still
	// has its original name. A self-update later renames it in place, so this
	// pre-update snapshot is what a restart must relaunch to run the new version.
	version.CaptureLaunchPath()

	// Load configuration
	cfg, err := a.configs.Load()
	if err != nil {
		log.Printf("Warning: failed to load config, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}
	a.config = cfg
	a.cfgStore = config.NewStore(cfg, config.Save)
	log.Printf("Configuration loaded successfully")

	// Initialize listening history store. Guarded like the tray and the
	// updater below, so a test (or a future alternative backing store) that
	// injected one before startup keeps it.
	if a.history == nil {
		a.history = history.NewStore(a.configs.ConfigDir(), historyEntryLimit)
	}

	// Bring an existing auto-start registration up to date with what this build
	// registers: entries written by older versions launch the bare executable,
	// with no flag marking the launch as one the OS performed at login, so
	// PlexCord would open its window on every boot (see EnsureRegistered).
	if err := a.autostart.EnsureRegistered(); err != nil {
		log.Printf("Warning: failed to refresh auto-start registration: %v", err)
	}

	// Same idea for an installed copy on Windows: a self-update replaces the
	// executable without running the installer, so the version shown in
	// "Apps & features" is refreshed here. No-op for portable and non-Windows
	// builds, and for dev builds, which have no release version to advertise.
	if !version.IsDevBuild() {
		if err := platform.SyncInstalledVersion(version.GetInfo().Version); err != nil {
			log.Printf("Warning: failed to refresh the installed version registration: %v", err)
		}
	}

	// Start the system tray. This is the visible affordance for restoring the
	// window (or quitting) once the app is running in the background, so it
	// runs regardless of the "Minimize to tray" setting.
	if a.tray == nil {
		a.tray = platform.NewTrayManager(platform.TrayCallbacks{
			OnShow:   a.ShowWindow,
			OnQuit:   a.QuitApp,
			OnUpdate: a.onTrayUpdateClick,
		}, platform.TrayIcons{
			PNG:       a.trayIconPNG,
			ICO:       a.trayIconICO,
			UpdatePNG: a.trayIconUpdatePNG,
			UpdateICO: a.trayIconUpdateICO,
		})
	}
	a.tray.Start()

	// Setup retry callbacks for automatic reconnection
	a.setupRetryCallbacks()

	// Start the automatic update checker (startup check + periodic re-check).
	// No-op for dev builds; can be toggled at runtime via SetAutoUpdateCheck.
	if a.updater == nil {
		a.updater = updater.New(a.bus, 6*time.Hour)
	}
	// Mirror update state into the tray menu: the frontend toast needs an open
	// window, and PlexCord is built to run without one.
	a.updater.OnStatusChange(a.publishUpdateNotice)
	if a.config.IsAutoUpdateCheckEnabled() {
		a.updater.StartChecker(ctx)
	}

	// Check if Plex token is available in keychain
	// The token will be used in later stories for Plex connection
	token, err := a.tokens.Get()
	switch {
	case err != nil:
		log.Printf("Warning: failed to retrieve Plex token: %v", err)
	case token != "":
		log.Printf("Plex token retrieved successfully from secure storage")
	default:
		log.Printf("No Plex token found - user needs to complete setup")
	}

	// Auto-connect to Discord and Plex if setup is complete
	if a.configs.IsSetupComplete() {
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

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
//
// When "Minimize to tray" is enabled, clicking the window close button hides
// the window and keeps PlexCord running in the background instead of quitting.
// Explicit quits (QuitApp) set the quitting flag so this path is bypassed.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	if !a.windows.IsQuitting() && a.config != nil && a.config.MinimizeToTray {
		log.Printf("Close requested: hiding window, PlexCord keeps running in the background")
		a.desktop.Hide(ctx)
		return true
	}
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Stop retry managers first to prevent post-shutdown reconnection attempts
	a.plexRetry.Stop()
	a.discordRetry.Stop()

	// Stop the automatic update checker
	if a.updater != nil {
		a.updater.StopChecker()
	}

	// Stop session polling if running
	a.StopSessionPolling()

	// Disconnect Discord
	if a.discord != nil {
		if err := a.discord.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect Discord: %v", err)
		}
	}

	// Remove the system tray icon
	if a.tray != nil {
		a.tray.Stop()
	}

	log.Printf("Application shutdown complete")
}
