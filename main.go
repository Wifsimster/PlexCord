package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"plexcord/internal/config"
	"plexcord/internal/platform"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

//go:embed build/windows/icon.ico
var iconWindows []byte

// Badged variants, shown in the tray while an update waits to be applied.
//
//go:embed build/appicon-update.png
var iconUpdate []byte

//go:embed build/windows/icon-update.ico
var iconWindowsUpdate []byte

func main() {
	// When this process was spawned by an in-app update relaunch, wait for the
	// old instance to fully exit before wails.Run acquires the single-instance
	// lock. Otherwise the lock is still held by the outgoing process and this
	// (updated) instance would be treated as a second instance and exit,
	// leaving the previous version running. No-op for normal launches.
	isUpdateRelaunch := waitForPreviousInstanceExit()

	// Decide how the window comes up before handing the options to Wails: the
	// "Start minimized" setting lives in the config file, which the app itself
	// only loads later, in OnStartup. A login launch starts in the background
	// whatever that setting says, and an update relaunch always shows the
	// window (see resolveWindowLaunchState).
	launch := resolveWindowLaunchState(loadLaunchConfig(config.Load), launchContext{
		IsUpdateRelaunch: isUpdateRelaunch,
		IsAutoStart:      isAutoStartLaunch(os.Args[1:]),
	})

	// Create an instance of the app structure
	app := NewApp()

	// Provide the tray icon assets (embedded above) to the app so the platform
	// layer can render the system tray without importing embedded assets.
	app.trayIcons = platform.TrayIcons{
		PNG:       icon,
		ICO:       iconWindows,
		UpdatePNG: iconUpdate,
		UpdateICO: iconWindowsUpdate,
	}

	// Create application with options
	// Note: Removed MaxWidth and MaxHeight
	err := wails.Run(&options.App{
		Title: "plexcord",
		// Sized to the content, not to round numbers: the Dashboard grid sets
		// the width and the setup wizard's tallest step sets the height (see
		// window_state.go). Wails needs a size before any screen is known, so
		// this is the preferred size — app.startup shrinks it to fit the actual
		// display via adaptWindowToScreen.
		Width:         preferredWindowWidth,
		Height:        preferredWindowHeight,
		MinWidth:      minWindowWidth,
		MinHeight:     minWindowHeight,
		DisableResize: false,
		Fullscreen:    false,
		Frameless:     true, // Custom in-app title bar (single merged header)
		// Start hidden when PlexCord runs straight into the tray ("Start
		// minimized", or a launch the OS performed at login); the tray icon and
		// relaunching the app both restore it.
		StartHidden: launch.StartHidden,
		// Close behavior is handled dynamically in app.beforeClose so it can
		// honor the user's "Minimize to tray" setting: hide to the background
		// when enabled, quit when disabled.
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:          nil,
		Logger:        nil,
		LogLevel:      logger.DEBUG,
		OnStartup:     app.startup,
		OnDomReady:    app.domReady,
		OnBeforeClose: app.beforeClose,
		OnShutdown:    app.shutdown,
		// Normal, or Minimised when PlexCord starts in the background without
		// "Minimize to tray" (see resolveWindowLaunchState).
		WindowStartState: launch.StartState,
		// A single-instance lock complements the system tray: it prevents
		// stacking up background copies and, when PlexCord is relaunched while
		// already running, restores the existing window instead of starting
		// another instance.
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "com.plexcord.app",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		Bind: []interface{}{
			app,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
			ZoomFactor:          1.0,
		},
		// Mac platform specific options
		// NOTE: Changed TitlebarAppearsTransparent to false
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "plexcord",
				Message: "",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
