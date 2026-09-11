package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"plexcord/internal/artwork"
	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/events"
	"plexcord/internal/history"
	"plexcord/internal/platform"
	"plexcord/internal/plex"
	"plexcord/internal/retry"
	"plexcord/internal/updater"
	"plexcord/internal/version"
)

// App struct
type App struct {
	ctx context.Context
	// config holds a pointer to the current Config for direct reads; all
	// writes go through cfgStore which handles atomic mutation + persistence.
	config         *config.Config
	cfgStore       *config.Store
	pollerCtx      context.Context
	pollerStop     context.CancelFunc
	currentSession *plex.MusicSession // Track current playback for page refresh restoration

	// Session polling
	poller *plex.Poller

	// Discord integration (production type, accessed via DiscordPresence interface)
	discord DiscordPresence

	// artwork resolves public album-art URLs so covers render on Discord
	// without leaking the Plex token; accessed via the ArtworkResolver interface.
	artwork ArtworkResolver

	// artworkGen debounces async artwork re-issues: each session change bumps
	// it, and a late resolve only re-issues presence if its generation is current.
	artworkGen atomic.Uint64

	// Plex client factory for constructing clients on demand (per-server).
	// Using a factory instead of a singleton reflects that the server URL/token
	// can change at runtime and enables tests to inject fakes.
	plexFactory PlexAPIFactory

	// Token store abstracts credential persistence (OS keychain in production)
	tokens TokenStore

	// Platform integration
	autostart *platform.AutoStartManager
	tray      *platform.TrayManager

	// Tray icon data, injected from main so the platform layer stays
	// asset-agnostic. iconPNG is used on macOS/Linux, iconICO on Windows.
	trayIconPNG       []byte
	trayIconICO       []byte
	trayIconUpdatePNG []byte // badged variants, shown while an update is pending
	trayIconUpdateICO []byte

	// Retry managers (Story 6.4)
	plexRetry    *retry.Manager
	discordRetry *retry.Manager

	// Automatic update checker (constructed in startup — it needs the bus)
	updater *updater.Updater

	// PIN authentication (maintain same client ID for PIN lifecycle)
	plexAuth *plex.Authenticator

	// Listening history
	history *history.Store

	// Event bus for emitting events to the frontend (abstracts Wails runtime)
	bus events.Bus

	// Presence pause state
	presencePaused bool        // Manual one-click pause toggle
	pauseTimer     *time.Timer // Timer for delayed hide-when-paused
	pauseTimerGen  uint64      // Incremented every schedule/cancel so a fired-but-cancelled callback bails out

	// quitting is set when the user explicitly quits (e.g. via QuitApp) so
	// beforeClose knows to allow shutdown instead of hiding to the background.
	quitting atomic.Bool

	// windowCtx is the Wails context published once the window can be driven;
	// pendingShow records a restore request that arrived before that — a second
	// instance launched while PlexCord was still booting — so it can be
	// replayed instead of dropped. Both are guarded by windowMu.
	windowCtx   context.Context
	pendingShow bool

	// Mutexes grouped together for alignment
	windowMu   sync.Mutex // Protect windowCtx and pendingShow
	pollerMu   sync.Mutex
	sessionMu  sync.RWMutex // Protect currentSession access
	discordMu  sync.Mutex
	plexAuthMu sync.Mutex
	pauseMu    sync.Mutex // Protect presencePaused and pauseTimer
}

// saveConfig persists the current in-memory config via the ConfigStore.
// This is the single path for all config writes — callers that need to
// mutate the config should set fields on a.config then call this method.
// Future enhancements (debouncing, schema migration, atomic writes) can
// hook in here without changing call sites.
func (a *App) saveConfig() error {
	if a.cfgStore == nil {
		// Fallback for code paths that run before startup (should not happen
		// in practice, but keeps tests that bypass startup working).
		return config.Save(a.config)
	}
	// No-op mutator: the caller already updated a.config directly; this
	// just triggers the store's atomic save path.
	return a.cfgStore.Update(func(*config.Config) {})
}

// NewApp creates a new App application struct with production dependencies.
// For tests, construct an App directly with injected fakes for bus,
// plexFactory, tokens, and discord.
func NewApp() *App {
	return &App{
		discord:      discord.NewPresenceManager(),
		plexFactory:  newPlexClientFactory(),
		tokens:       newKeychainTokenStore(),
		autostart:    platform.NewAutoStartManager(),
		plexRetry:    retry.NewManager("Plex"),
		discordRetry: retry.NewManager("Discord"),
		artwork:      artwork.NewResolver(artwork.WithUserAgent("PlexCord/" + version.Version)),
	}
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
	if a.markWindowReady(ctx) {
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
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: failed to load config, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}
	a.config = cfg
	a.cfgStore = config.NewStore(cfg, config.Save)
	log.Printf("Configuration loaded successfully")

	// Initialize listening history store
	configDir := config.GetConfigDir()
	a.history = history.NewStore(configDir, 200)

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
	a.tray.Start()

	// Setup retry callbacks for automatic reconnection
	a.setupRetryCallbacks()

	// Start the automatic update checker (startup check + periodic re-check).
	// No-op for dev builds; can be toggled at runtime via SetAutoUpdateCheck.
	a.updater = updater.New(a.bus, 6*time.Hour)
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
	if !a.quitting.Load() && a.config != nil && a.config.MinimizeToTray {
		log.Printf("Close requested: hiding window, PlexCord keeps running in the background")
		runtime.WindowHide(ctx)
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
