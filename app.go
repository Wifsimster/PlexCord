package main

import (
	"context"
	"log"
	"sync"
	"time"

	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/history"
	"plexcord/internal/keychain"
	"plexcord/internal/platform"
	"plexcord/internal/plex"
	"plexcord/internal/retry"
)

// App struct
type App struct {
	ctx            context.Context
	config         *config.Config
	pollerCtx      context.Context
	pollerStop     context.CancelFunc
	currentSession *plex.MusicSession // Track current playback for page refresh restoration

	// Session polling
	poller *plex.Poller

	// Discord integration
	discord *discord.PresenceManager

	// Platform integration
	autostart *platform.AutoStartManager

	// Retry managers (Story 6.4)
	plexRetry    *retry.Manager
	discordRetry *retry.Manager

	// PIN authentication (maintain same client ID for PIN lifecycle)
	plexAuth *plex.Authenticator

	// Listening history
	history *history.Store

	// Presence pause state
	presencePaused bool        // Manual one-click pause toggle
	pauseTimer     *time.Timer // Timer for delayed hide-when-paused

	// Mutexes grouped together for alignment
	pollerMu   sync.Mutex
	sessionMu  sync.RWMutex // Protect currentSession access
	discordMu  sync.Mutex
	plexAuthMu sync.Mutex
	pauseMu    sync.Mutex // Protect presencePaused and pauseTimer
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

	// Initialize listening history store
	configDir := config.GetConfigDir()
	a.history = history.NewStore(configDir, 200)

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
	// Stop retry managers first to prevent post-shutdown reconnection attempts
	a.plexRetry.Stop()
	a.discordRetry.Stop()

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
