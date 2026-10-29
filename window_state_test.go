package main

import (
	"context"
	"errors"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"plexcord/internal/config"
	"plexcord/internal/platform"
)

func TestResolveWindowLaunchState(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.Config
		launch     launchContext
		wantHidden bool
		wantState  options.WindowStartState
	}{
		{
			name:       "start minimized off shows the window",
			cfg:        &config.Config{StartMinimized: false, MinimizeToTray: true},
			wantHidden: false,
			wantState:  options.Normal,
		},
		{
			name:       "start minimized with tray starts hidden",
			cfg:        &config.Config{StartMinimized: true, MinimizeToTray: true},
			wantHidden: true,
			wantState:  options.Normal,
		},
		{
			name:       "start minimized without tray minimizes to the taskbar",
			cfg:        &config.Config{StartMinimized: true, MinimizeToTray: false},
			wantHidden: false,
			wantState:  options.Minimised,
		},
		{
			name:       "login launch starts in the tray even without start minimized",
			cfg:        &config.Config{StartMinimized: false, MinimizeToTray: true},
			launch:     launchContext{IsAutoStart: true},
			wantHidden: true,
			wantState:  options.Normal,
		},
		{
			name:       "login launch without tray minimizes to the taskbar",
			cfg:        &config.Config{StartMinimized: false, MinimizeToTray: false},
			launch:     launchContext{IsAutoStart: true},
			wantHidden: false,
			wantState:  options.Minimised,
		},
		{
			name: "login launch shows the window when the toggle is off",
			cfg: &config.Config{
				StartMinimized:        false,
				MinimizeToTray:        true,
				StartMinimizedOnLogin: boolPtr(false),
			},
			launch:     launchContext{IsAutoStart: true},
			wantHidden: false,
			wantState:  options.Normal,
		},
		{
			name: "start minimized still applies with the login toggle off",
			cfg: &config.Config{
				StartMinimized:        true,
				MinimizeToTray:        true,
				StartMinimizedOnLogin: boolPtr(false),
			},
			launch:     launchContext{IsAutoStart: true},
			wantHidden: true,
			wantState:  options.Normal,
		},
		{
			name:       "update relaunch always shows the window",
			cfg:        &config.Config{StartMinimized: true, MinimizeToTray: true},
			launch:     launchContext{IsUpdateRelaunch: true},
			wantHidden: false,
			wantState:  options.Normal,
		},
		{
			name:       "update relaunch wins over a login launch",
			cfg:        &config.Config{StartMinimized: false, MinimizeToTray: true},
			launch:     launchContext{IsUpdateRelaunch: true, IsAutoStart: true},
			wantHidden: false,
			wantState:  options.Normal,
		},
		{
			name:       "nil config falls back to a normal window",
			cfg:        nil,
			wantHidden: false,
			wantState:  options.Normal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveWindowLaunchState(tt.cfg, tt.launch)
			if got.StartHidden != tt.wantHidden {
				t.Errorf("StartHidden = %v, want %v", got.StartHidden, tt.wantHidden)
			}
			if got.StartState != tt.wantState {
				t.Errorf("StartState = %v, want %v", got.StartState, tt.wantState)
			}
		})
	}
}

func TestDefaultConfigStartsVisible(t *testing.T) {
	if got := resolveWindowLaunchState(config.DefaultConfig(), launchContext{}); got.StartHidden || got.StartState != options.Normal {
		t.Fatalf("default config launch state = %+v, want a normal visible window", got)
	}
}

func boolPtr(b bool) *bool { return &b }

func TestIsAutoStartLaunch(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no arguments", args: nil, want: false},
		{name: "double dash", args: []string{platform.AutoStartFlag}, want: true},
		{name: "single dash", args: []string{"-autostart"}, want: true},
		{name: "alongside other arguments", args: []string{"--verbose", platform.AutoStartFlag}, want: true},
		{name: "unrelated argument", args: []string{"--autostartle"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAutoStartLaunch(tt.args); got != tt.want {
				t.Errorf("isAutoStartLaunch(%q) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestFitWindowToScreen(t *testing.T) {
	preferred := windowSize{Width: preferredWindowWidth, Height: preferredWindowHeight}

	tests := []struct {
		name   string
		screen windowSize
		want   windowSize
	}{
		{
			name:   "roomy screen keeps the preferred size",
			screen: windowSize{Width: 2560, Height: 1440},
			want:   preferred,
		},
		{
			name:   "1080p fits the preferred size",
			screen: windowSize{Width: 1920, Height: 1080},
			want:   preferred,
		},
		{
			name:   "1366x768 laptop leaves room for the taskbar",
			screen: windowSize{Width: 1366, Height: 768},
			want:   windowSize{Width: 1248, Height: 691},
		},
		{
			name:   "small screen shrinks both dimensions",
			screen: windowSize{Width: 1280, Height: 720},
			want:   windowSize{Width: 1216, Height: 648},
		},
		{
			name:   "minimum size wins over the screen fraction",
			screen: windowSize{Width: 900, Height: 640},
			want:   windowSize{Width: minWindowWidth, Height: minWindowHeight},
		},
		{
			name:   "never larger than the screen itself",
			screen: windowSize{Width: 800, Height: 560},
			want:   windowSize{Width: 800, Height: 560},
		},
		{
			name:   "unknown screen keeps the preferred size",
			screen: windowSize{},
			want:   preferred,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fitWindowToScreen(preferred, tt.screen); got != tt.want {
				t.Errorf("fitWindowToScreen(%+v, %+v) = %+v, want %+v", preferred, tt.screen, got, tt.want)
			}
		})
	}
}

func TestPreferredWindowSizeFitsItsContent(t *testing.T) {
	const dashboardHeight = 663
	const widestViewWidth = 1248

	if preferredWindowHeight < dashboardHeight {
		t.Errorf("preferredWindowHeight = %d, want >= %d so the Dashboard fits without scrolling", preferredWindowHeight, dashboardHeight)
	}
	if preferredWindowWidth < widestViewWidth {
		t.Errorf("preferredWindowWidth = %d, want >= %d so the Dashboard grid reaches its full width", preferredWindowWidth, widestViewWidth)
	}
	const wizardMinWidth = 848
	if minWindowWidth < wizardMinWidth {
		t.Errorf("minWindowWidth = %d, want >= %d so the wizard step column is not squeezed", minWindowWidth, wizardMinWidth)
	}
	if minWindowHeight > 691 {
		t.Errorf("minWindowHeight = %d, want <= 691 so the window fits a 1366x768 display", minWindowHeight)
	}
}

func screen(width, height int) runtime.Screen {
	var s runtime.Screen
	s.Size.Width = width
	s.Size.Height = height
	return s
}

func currentScreen(width, height int) runtime.Screen {
	s := screen(width, height)
	s.IsCurrent = true
	return s
}

func primaryScreen(width, height int) runtime.Screen {
	s := screen(width, height)
	s.IsPrimary = true
	return s
}

func TestScreenForWindow(t *testing.T) {
	tests := []struct {
		name    string
		screens []runtime.Screen
		want    windowSize
	}{
		{
			name:    "no screens reported",
			screens: nil,
			want:    windowSize{},
		},
		{
			name:    "prefers the screen the window is on",
			screens: []runtime.Screen{primaryScreen(1920, 1080), currentScreen(1366, 768)},
			want:    windowSize{Width: 1366, Height: 768},
		},
		{
			name:    "falls back to the primary screen",
			screens: []runtime.Screen{screen(1024, 600), primaryScreen(1920, 1080)},
			want:    windowSize{Width: 1920, Height: 1080},
		},
		{
			name:    "falls back to the first screen",
			screens: []runtime.Screen{screen(1600, 900)},
			want:    windowSize{Width: 1600, Height: 900},
		},
		{
			name:    "skips a screen with no dimensions",
			screens: []runtime.Screen{{IsCurrent: true}, primaryScreen(1600, 900)},
			want:    windowSize{Width: 1600, Height: 900},
		},
		{
			name:    "no screen reports dimensions",
			screens: []runtime.Screen{{IsCurrent: true}, {IsPrimary: true}},
			want:    windowSize{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := screenForWindow(tt.screens); got != tt.want {
				t.Errorf("screenForWindow() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestShowWindowBeforeReadyIsDeferred(t *testing.T) {
	desktop := &fakeDesktop{}
	app := newTestApp(config.DefaultConfig())
	app.desktop = desktop
	app.windows = newWindowManager(desktop, desktop)

	app.ShowWindow()

	if shown, _, _ := desktop.counts(); shown != 0 {
		t.Fatalf("ShowWindow before startup drove the window %d time(s)", shown)
	}

	if !app.windows.MarkReady(context.Background()) {
		t.Fatal("MarkReady() = false, want true to replay the parked request")
	}
	if app.windows.MarkReady(context.Background()) {
		t.Fatal("MarkReady() replayed the same request twice")
	}

	app.ShowWindow()
	if shown, _, _ := desktop.counts(); shown != 1 {
		t.Fatalf("ShowWindow after ready drove the window %d time(s), want 1", shown)
	}
}

func TestWindowContextAfterReady(t *testing.T) {
	ctx := context.Background()
	desktop := &fakeDesktop{}
	windows := newWindowManager(desktop, desktop)

	if windows.MarkReady(ctx) {
		t.Fatal("MarkReady() = true with no request pending")
	}
	if windows.Context() == nil {
		t.Fatal("Context() = nil after the window became ready")
	}
	if windows.MarkReady(ctx) {
		t.Fatal("Context() parked a request even though the window was ready")
	}
}

func TestLoadLaunchConfigFallsBackToDefaults(t *testing.T) {
	cfg := loadLaunchConfig(func() (*config.Config, error) {
		return nil, errors.New("config file is corrupt")
	})

	if cfg == nil {
		t.Fatal("loadLaunchConfig() = nil on a read failure; the window would never open")
	}
	if cfg.StartMinimized {
		t.Error("the fallback config starts minimized; a broken config must open the window")
	}
}

func TestLoadLaunchConfigUsesLoadedConfig(t *testing.T) {
	want := &config.Config{StartMinimized: true}

	cfg := loadLaunchConfig(func() (*config.Config, error) { return want, nil })

	if cfg != want {
		t.Fatalf("loadLaunchConfig() = %+v, want the loaded config", cfg)
	}
}
