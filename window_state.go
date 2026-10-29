package main

import (
	"context"
	"log"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"plexcord/internal/config"
	"plexcord/internal/platform"
)

const (
	preferredWindowWidth  = 1248
	preferredWindowHeight = 700
)

const (
	minWindowWidth  = 880
	minWindowHeight = 600
)

const (
	maxScreenWidthFraction  = 0.95
	maxScreenHeightFraction = 0.90
)

type windowSize struct {
	Width  int
	Height int
}

type windowLaunchState struct {
	StartState  options.WindowStartState
	StartHidden bool
}

type launchContext struct {
	IsUpdateRelaunch bool
	IsAutoStart      bool
}

func resolveWindowLaunchState(cfg *config.Config, launch launchContext) windowLaunchState {
	if cfg == nil || launch.IsUpdateRelaunch {
		return windowLaunchState{StartState: options.Normal}
	}
	background := cfg.StartMinimized || (launch.IsAutoStart && cfg.LoginStartsMinimized())
	if !background {
		return windowLaunchState{StartState: options.Normal}
	}
	if cfg.MinimizeToTray {
		return windowLaunchState{StartState: options.Normal, StartHidden: true}
	}
	return windowLaunchState{StartState: options.Minimised}
}

func isAutoStartLaunch(args []string) bool {
	for _, arg := range args {
		if arg == platform.AutoStartFlag || arg == strings.TrimPrefix(platform.AutoStartFlag, "-") {
			return true
		}
	}
	return false
}

func fitWindowToScreen(preferred windowSize, screen windowSize) windowSize {
	fit := func(want, minimum, available int, fraction float64) int {
		if available <= 0 {
			return want
		}
		size := want
		if usable := int(float64(available) * fraction); size > usable {
			size = usable
		}
		if size < minimum {
			size = minimum
		}
		if size > available {
			size = available
		}
		return size
	}

	return windowSize{
		Width:  fit(preferred.Width, minWindowWidth, screen.Width, maxScreenWidthFraction),
		Height: fit(preferred.Height, minWindowHeight, screen.Height, maxScreenHeightFraction),
	}
}

func screenForWindow(screens []runtime.Screen) windowSize {
	var current, primary, first windowSize
	for i := range screens {
		size := windowSize{Width: screens[i].Size.Width, Height: screens[i].Size.Height}
		if size.Width <= 0 || size.Height <= 0 {
			continue
		}
		if screens[i].IsCurrent && current == (windowSize{}) {
			current = size
		}
		if screens[i].IsPrimary && primary == (windowSize{}) {
			primary = size
		}
		if first == (windowSize{}) {
			first = size
		}
	}

	switch {
	case current != windowSize{}:
		return current
	case primary != windowSize{}:
		return primary
	default:
		return first
	}
}

func (a *App) adaptWindowToScreen(ctx context.Context) {
	screens, err := a.desktop.Screens(ctx)
	if err != nil {
		log.Printf("Warning: failed to read screen size, keeping the default window size: %v", err)
		return
	}

	screen := screenForWindow(screens)
	preferred := windowSize{Width: preferredWindowWidth, Height: preferredWindowHeight}
	fitted := fitWindowToScreen(preferred, screen)
	if fitted == preferred {
		return
	}

	log.Printf("Window resized to %dx%d to fit the %dx%d screen", fitted.Width, fitted.Height, screen.Width, screen.Height)
	a.windows.Resize(fitted.Width, fitted.Height)
}

func loadLaunchConfig(load func() (*config.Config, error)) *config.Config {
	cfg, err := load()
	if err != nil {
		log.Printf("Warning: failed to load config for window start state, using defaults: %v", err)
		return config.DefaultConfig()
	}
	return cfg
}
