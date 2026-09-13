package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
)

// windowManager owns everything about the application window's visibility and
// the process's exit: the Wails context that can drive the window, a restore
// request that arrived before that context existed, and the flag that tells
// beforeClose an exit was explicitly asked for.
//
// It is split out of App because window visibility changes for its own reasons
// — the tray, a second launch, the close button, an update restart — none of
// which have anything to do with Plex polling, Discord presence or settings
// persistence (SRP). Holding it separately also means the whole restore dance
// can be tested against a fake WindowController, with no Wails process.
type windowManager struct {
	window  WindowController
	process ProcessController

	mu sync.Mutex
	// ctx is the Wails context, published once the window can be driven.
	ctx context.Context
	// pendingShow records a restore request that arrived before ctx existed —
	// a second instance launched while PlexCord was still booting — so it can
	// be replayed instead of dropped.
	pendingShow bool

	// quitting is set when the user explicitly quits, so the close handler
	// knows to allow shutdown instead of hiding to the background.
	quitting atomic.Bool
}

// newWindowManager builds a manager over the given runtime collaborators.
func newWindowManager(window WindowController, process ProcessController) *windowManager {
	return &windowManager{window: window, process: process}
}

// MarkReady publishes the context that drives the window and reports whether a
// restore request arrived before it existed, in which case the caller should
// replay it.
func (w *windowManager) MarkReady(ctx context.Context) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.ctx = ctx
	pending := w.pendingShow
	w.pendingShow = false
	return pending
}

// Context returns the context to drive the window with, or nil when the window
// is not ready yet. In that case the request is remembered so MarkReady can
// replay it, rather than being dropped (or run against a nil context, which
// panics inside the Wails runtime).
func (w *windowManager) Context() context.Context {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ctx != nil {
		return w.ctx
	}
	w.pendingShow = true
	return nil
}

// readyContext returns the window context without recording a pending restore.
// Used by operations that are meaningless before the window exists (hide,
// minimise) and should simply no-op rather than queue.
func (w *windowManager) readyContext() context.Context {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ctx
}

// Show brings the window to the foreground. Called when restoring from the
// tray, and when PlexCord is relaunched while running in the background.
func (w *windowManager) Show() {
	ctx := w.Context()
	if ctx == nil {
		log.Printf("Show requested before the window was ready; deferred until startup completes")
		return
	}

	w.window.Show(ctx)
	// Only un-minimise when the window actually is minimised: on a window that
	// was hidden while maximised, an unconditional restore would also drop it
	// back to its normal size.
	if w.window.IsMinimised(ctx) {
		w.window.Unminimise(ctx)
	}
	w.window.SetAlwaysOnTop(ctx, true)
	w.window.SetAlwaysOnTop(ctx, false) // Trick to bring to front
}

// Hide hides the window; the application keeps running in the background.
func (w *windowManager) Hide() {
	if ctx := w.readyContext(); ctx != nil {
		w.window.Hide(ctx)
	}
}

// Minimise minimises the window.
func (w *windowManager) Minimise() {
	if ctx := w.readyContext(); ctx != nil {
		w.window.Minimise(ctx)
	}
}

// Resize sets the window size.
func (w *windowManager) Resize(width, height int) {
	if ctx := w.readyContext(); ctx != nil {
		w.window.SetSize(ctx, width, height)
		w.window.Center(ctx)
	}
}

// Quit terminates the application, flagging the exit as explicit so the close
// handler does not intercept it and merely hide the window.
func (w *windowManager) Quit() {
	w.quitting.Store(true)
	if ctx := w.readyContext(); ctx != nil {
		w.process.Quit(ctx)
	}
}

// MarkQuitting flags an explicit exit without triggering one, for callers that
// quit through another path (the update restart).
func (w *windowManager) MarkQuitting() {
	w.quitting.Store(true)
}

// IsQuitting reports whether an explicit quit is under way.
func (w *windowManager) IsQuitting() bool {
	return w.quitting.Load()
}
