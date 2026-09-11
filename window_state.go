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

// Window geometry, derived from what the UI actually lays out rather than from
// round numbers. All figures are measured against the design system contracts
// in docs/design-system.md §5.1–5.4 with the shell chrome added back.
//
// Width — the Dashboard is the widest surface: its grid caps at 1200px (§5.2)
// inside the 24px page gutter on both sides, so 1248px is where it stops
// growing. Settings caps at 1040 (1088 with gutters) and the wizard needs
// 240px rail + 560px step column + gutters (848); both center inside the
// Dashboard's width, so the Dashboard sets it.
//
// Height — the Dashboard is the view that stays open, so it sets the height:
// 64px top gutter + 525px of panels in its tallest (idle) state + 46px footer
// + 24px ≈ 663px, and 700 gives that a little air. Sizing for the tallest view
// instead would mean ~800px, which the Dashboard would spend most of its life
// padding out with empty canvas.
//
// The two views that need more scroll rather than resist it: Settings is
// ~2170px of content and no window height would change that, and the setup
// wizard's Complete step (796px) scrolls inside its own pane with the footer
// bar pinned — a one-time flow, and one that already scrolls on a 1366×768
// laptop once fitWindowToScreen has had its say.
const (
	preferredWindowWidth  = 1248
	preferredWindowHeight = 700
)

// Minimum size — the floor where every layout still resolves to its designed
// form rather than a squeezed one.
//
// 848px is where the wizard's step column reaches its full 560px (240 + 2×24 +
// 560); 880 rounds that up with a little air. Below it the wizard narrows,
// while the Dashboard (single column under 992px) and Settings (rail folds
// into a row under 860px) already have responsive fallbacks — the point of the
// floor is that nothing is *forced* into them.
//
// 600px in height is deliberately below any view's natural height: pages
// scroll, the wizard has its own scroller with a pinned footer, and a floor
// taller than that is what stops the window fitting on small or scaled
// displays. The previous 768px minimum was the entire height of a 1366×768
// laptop screen, leaving no room for a taskbar.
const (
	minWindowWidth  = 880
	minWindowHeight = 600
)

// Fraction of the screen the initial window may occupy. The remainder absorbs
// the taskbar/dock and window shadows, which no cross-platform Wails API
// reports (v2 exposes screen size, not work area).
const (
	maxScreenWidthFraction  = 0.95
	maxScreenHeightFraction = 0.90
)

// windowSize is a plain width/height pair, used to keep the sizing arithmetic
// independent of the Wails runtime types.
type windowSize struct {
	Width  int
	Height int
}

// windowLaunchState describes how the main window should come up, expressed in
// the two Wails options that decide it.
type windowLaunchState struct {
	StartState  options.WindowStartState
	StartHidden bool
}

// launchContext describes how this process came to be running. Both cases
// override what the settings alone would decide, so they travel together.
type launchContext struct {
	// IsUpdateRelaunch marks the instance the in-app updater spawned after
	// installing a new version.
	IsUpdateRelaunch bool
	// IsAutoStart marks a launch the OS performed at login, recognized by the
	// flag PlexCord registers its auto-start entry with.
	IsAutoStart bool
}

// resolveWindowLaunchState maps the user's preferences, and how PlexCord was
// launched, onto those options.
//
// Starting in the background has two flavors, mirroring what closing the
// window already does:
//   - with "Minimize to tray" on, PlexCord starts hidden — no window and no
//     taskbar button, just the tray icon. Restoring goes through the tray or
//     through relaunching PlexCord, which the single-instance lock turns into
//     a "show the running instance" request (see App.onSecondInstanceLaunch).
//   - with it off, the window starts minimized to the taskbar instead. Hiding
//     it outright would leave the user no way back: closing to the tray is
//     disabled, so the taskbar button is the only restore affordance.
//
// Two things put PlexCord in the background. "Start minimized" does it for
// every launch. A login launch does it on its own, whatever that setting says:
// nobody asked for a window at that moment — the OS started PlexCord, not the
// user — and a presence bridge that pops a window over the desktop on every
// boot is the reason people turn "Start on login" back off.
//
// An update relaunch is the mirror image and wins over both: the user just
// clicked "restart to apply", and a restart that vanished into the tray would
// read as a crash.
func resolveWindowLaunchState(cfg *config.Config, launch launchContext) windowLaunchState {
	if cfg == nil || launch.IsUpdateRelaunch {
		return windowLaunchState{StartState: options.Normal}
	}
	if !cfg.StartMinimized && !launch.IsAutoStart {
		return windowLaunchState{StartState: options.Normal}
	}
	if cfg.MinimizeToTray {
		return windowLaunchState{StartState: options.Normal, StartHidden: true}
	}
	return windowLaunchState{StartState: options.Minimised}
}

// isAutoStartLaunch reports whether the process was started by the OS at
// login, which every auto-start registration marks with
// platform.AutoStartFlag (see internal/platform/autostart.go). Both the
// double- and single-dash spellings are accepted, since Go's own flag package
// treats them as the same flag.
//
// Nothing else passes this flag: the update relaunch spawns the executable
// with no arguments at all, so a login launch is the only way it arrives.
func isAutoStartLaunch(args []string) bool {
	for _, arg := range args {
		if arg == platform.AutoStartFlag || arg == strings.TrimPrefix(platform.AutoStartFlag, "-") {
			return true
		}
	}
	return false
}

// fitWindowToScreen shrinks the preferred window size to something that fits
// the screen it will open on.
//
// The preferred size is derived from the content, not from any particular
// display, so on a small or heavily scaled screen it can be larger than the
// desktop — a 820px-tall window does not fit a 1366×768 laptop once the
// taskbar takes its share. Clamping keeps the window on-screen with its
// controls reachable.
//
// Order matters: the screen fraction wins over the preferred size, and the
// minimum size wins over the fraction (a window below its minimum is one Wails
// would immediately grow back), but nothing is allowed to exceed the screen
// itself. An unknown screen (no dimensions reported) leaves the preferred size
// untouched.
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

// screenForWindow picks the screen the window should be sized against: the one
// it is currently on, else the primary, else the first reported. Size is the
// logical pixel space Wails sizes windows in, which is the space the preferred
// size is expressed in too.
//
// Screens without usable dimensions are skipped rather than chosen: the Windows
// backend reports a zeroed Screen for a monitor whose DPI it could not read, and
// picking that one would discard the information a working monitor did report.
//
// The zero windowSize means "no usable screen information", which
// fitWindowToScreen reads as "leave the preferred size alone".
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

// adaptWindowToScreen resizes the window to fit the display it opens on.
//
// Wails takes its initial size before any screen is known (wails.Run receives
// static options), so the content-derived preferred size is applied up front
// and corrected here, on startup, once the runtime can report the screen. The
// window is re-centered afterwards because Wails centered it against the
// previous size.
//
// Resizing is safe while the window is hidden or minimized ("Start minimized"):
// every platform backend moves the window without showing it, so a tray start
// stays in the tray and comes back at the right size.
func (a *App) adaptWindowToScreen(ctx context.Context) {
	screens, err := runtime.ScreenGetAll(ctx)
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
	runtime.WindowSetSize(ctx, fitted.Width, fitted.Height)
	runtime.WindowCenter(ctx)
}

// loadLaunchConfig reads the persisted configuration for the sole purpose of
// deciding the initial window state, which Wails needs up front — before
// OnStartup runs and the app loads the config it actually works with.
//
// A config that cannot be read must never keep the window from opening, so
// failures fall back to defaults (which start the window normally).
func loadLaunchConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: failed to load config for window start state, using defaults: %v", err)
		return config.DefaultConfig()
	}
	return cfg
}
