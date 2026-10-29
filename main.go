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

//go:embed build/appicon-update.png
var iconUpdate []byte

//go:embed build/windows/icon-update.ico
var iconWindowsUpdate []byte

func main() {
	isUpdateRelaunch := waitForPreviousInstanceExit()

	launch := resolveWindowLaunchState(loadLaunchConfig(config.Load), launchContext{
		IsUpdateRelaunch: isUpdateRelaunch,
		IsAutoStart:      isAutoStartLaunch(os.Args[1:]),
	})

	app := NewApp()

	app.trayIcons = platform.TrayIcons{
		PNG:       icon,
		ICO:       iconWindows,
		UpdatePNG: iconUpdate,
		UpdateICO: iconWindowsUpdate,
	}

	err := wails.Run(&options.App{
		Title:             "plexcord",
		Width:             preferredWindowWidth,
		Height:            preferredWindowHeight,
		MinWidth:          minWindowWidth,
		MinHeight:         minWindowHeight,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         true,
		StartHidden:       launch.StartHidden,
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
		WindowStartState: launch.StartState,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "com.plexcord.app",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			WebviewUserDataPath:  "",
			ZoomFactor:           1.0,
		},
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
