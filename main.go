package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

//go:embed build/windows/icon.ico
var iconWindows []byte

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Provide the tray icon assets (embedded above) to the app so the platform
	// layer can render the system tray without importing embedded assets.
	app.trayIconPNG = icon
	app.trayIconICO = iconWindows

	// Create application with options
	// Note: Removed MaxWidth and MaxHeight
	err := wails.Run(&options.App{
		Title:         "plexcord",
		Width:         1100,
		Height:        1000,
		MinWidth:      1024,
		MinHeight:     768,
		DisableResize: false,
		Fullscreen:    false,
		Frameless:     true, // Custom in-app title bar (single merged header)
		StartHidden:   false,
		// Close behavior is handled dynamically in app.beforeClose so it can
		// honor the user's "Minimize to tray" setting: hide to the background
		// when enabled, quit when disabled.
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:             nil,
		Logger:           nil,
		LogLevel:         logger.DEBUG,
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		WindowStartState: options.Normal,
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
