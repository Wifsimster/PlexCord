package main

import (
	"context"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"plexcord/internal/config"
	"plexcord/internal/platform"
)

// TestResolveWindowLaunchState covers how the "Start minimized" and "Minimize
// to tray" settings, and how PlexCord was launched, combine into Wails' window
// start options.
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
			name: "start minimized without tray minimizes to the taskbar",
			// Hiding outright would leave no restore affordance: closing to
			// the tray is off, so the taskbar button is the way back.
			cfg:        &config.Config{StartMinimized: true, MinimizeToTray: false},
			wantHidden: false,
			wantState:  options.Minimised,
		},
		{
			name: "login launch starts in the tray even without start minimized",
			// Nobody asked for a window at boot: the OS started PlexCord, and
			// an unset StartMinimizedOnLogin means the default, which is on.
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
			// The two settings cover different launches; turning off the login
			// one does not undo "Start minimized".
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
			name: "update relaunch wins over a login launch",
			// An update that restarted into the tray would read as a crash,
			// even if the flag came along from the entry that started it.
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

// TestDefaultConfigStartsVisible verifies PlexCord opens its window when the
// user launches it themselves: starting in the background is opt-in, except at
// login (covered above).
func TestDefaultConfigStartsVisible(t *testing.T) {
	if got := resolveWindowLaunchState(config.DefaultConfig(), launchContext{}); got.StartHidden || got.StartState != options.Normal {
		t.Fatalf("default config launch state = %+v, want a normal visible window", got)
	}
}

// boolPtr mirrors the config package's helper for the optional bool fields.
func boolPtr(b bool) *bool { return &b }

// TestIsAutoStartLaunch covers recognizing the flag the auto-start entries are
// registered with — the only thing that tells a login launch apart from the
// user opening PlexCord.
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

// TestFitWindowToScreen covers shrinking the content-derived window size onto
// the display it actually opens on.
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
			// The case the old 1024×768 minimum made impossible: a window
			// taller than the screen it opens on.
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
			// Even the minimum must not push the window off a tiny display:
			// unreachable window controls are worse than a cramped layout.
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

// TestPreferredWindowSizeFitsItsContent guards the numbers against drifting
// below what the views measure (docs/design-system.md §5.2/§5.4).
func TestPreferredWindowSizeFitsItsContent(t *testing.T) {
	// Dashboard in its tallest (idle) state: 64px top gutter + 525px of panels
	// + 46px footer + 24px bottom gutter.
	const dashboardHeight = 663
	// Dashboard grid: 1200px content cap + 2×24px page gutter.
	const widestViewWidth = 1248

	if preferredWindowHeight < dashboardHeight {
		t.Errorf("preferredWindowHeight = %d, want >= %d so the Dashboard fits without scrolling", preferredWindowHeight, dashboardHeight)
	}
	if preferredWindowWidth < widestViewWidth {
		t.Errorf("preferredWindowWidth = %d, want >= %d so the Dashboard grid reaches its full width", preferredWindowWidth, widestViewWidth)
	}
	// Wizard rail (240) + step column (560) + 2×24px gutter.
	const wizardMinWidth = 848
	if minWindowWidth < wizardMinWidth {
		t.Errorf("minWindowWidth = %d, want >= %d so the wizard step column is not squeezed", minWindowWidth, wizardMinWidth)
	}
	// A minimum taller than this cannot fit a 1366×768 screen with a taskbar.
	if minWindowHeight > 691 {
		t.Errorf("minWindowHeight = %d, want <= 691 so the window fits a 1366x768 display", minWindowHeight)
	}
}

// screen builds a runtime.Screen with a logical size. Wails does not export
// the ScreenSize type from pkg/runtime (only the Screen alias), so the field is
// filled by assignment rather than in a composite literal.
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

// TestScreenForWindow covers which display the window is sized against.
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
			// The Windows backend reports a zeroed Screen for a monitor whose
			// DPI it could not read; a working monitor's numbers are better.
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

// TestShowWindowBeforeReadyIsDeferred verifies a restore request arriving
// before startup (a second instance launched while PlexCord is still booting)
// is parked rather than dropped — and does not run against a nil context.
func TestShowWindowBeforeReadyIsDeferred(t *testing.T) {
	app := &App{}

	app.ShowWindow()

	app.windowMu.Lock()
	pending := app.pendingShow
	app.windowMu.Unlock()
	if !pending {
		t.Fatal("ShowWindow before startup did not record a pending restore request")
	}

	if !app.markWindowReady(context.Background()) {
		t.Fatal("markWindowReady() = false, want true to replay the parked request")
	}
	if app.markWindowReady(context.Background()) {
		t.Fatal("markWindowReady() replayed the same request twice")
	}
}

// TestWindowContextAfterReady verifies that once the window is ready, show
// requests run against the published context instead of being parked.
func TestWindowContextAfterReady(t *testing.T) {
	ctx := context.Background()
	app := &App{}

	if app.markWindowReady(ctx) {
		t.Fatal("markWindowReady() = true with no request pending")
	}
	if app.windowContext() == nil {
		t.Fatal("windowContext() = nil after the window became ready")
	}

	app.windowMu.Lock()
	pending := app.pendingShow
	app.windowMu.Unlock()
	if pending {
		t.Fatal("windowContext() parked a request even though the window was ready")
	}
}
