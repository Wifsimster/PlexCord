package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
)

type windowManager struct {
	window  WindowController
	process ProcessController

	mu          sync.Mutex
	ctx         context.Context
	pendingShow bool

	quitting atomic.Bool
}

func newWindowManager(window WindowController, process ProcessController) *windowManager {
	return &windowManager{window: window, process: process}
}

func (w *windowManager) MarkReady(ctx context.Context) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.ctx = ctx
	pending := w.pendingShow
	w.pendingShow = false
	return pending
}

func (w *windowManager) Context() context.Context {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ctx != nil {
		return w.ctx
	}
	w.pendingShow = true
	return nil
}

func (w *windowManager) readyContext() context.Context {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ctx
}

func (w *windowManager) Show() {
	ctx := w.Context()
	if ctx == nil {
		log.Printf("Show requested before the window was ready; deferred until startup completes")
		return
	}

	w.window.Show(ctx)
	if w.window.IsMinimised(ctx) {
		w.window.Unminimise(ctx)
	}
	w.window.SetAlwaysOnTop(ctx, true)
	w.window.SetAlwaysOnTop(ctx, false)
}

func (w *windowManager) Hide() {
	if ctx := w.readyContext(); ctx != nil {
		w.window.Hide(ctx)
	}
}

func (w *windowManager) Minimise() {
	if ctx := w.readyContext(); ctx != nil {
		w.window.Minimise(ctx)
	}
}

func (w *windowManager) Resize(width, height int) {
	if ctx := w.readyContext(); ctx != nil {
		w.window.SetSize(ctx, width, height)
		w.window.Center(ctx)
	}
}

func (w *windowManager) Quit() {
	w.quitting.Store(true)
	if ctx := w.readyContext(); ctx != nil {
		w.process.Quit(ctx)
	}
}

func (w *windowManager) MarkQuitting() {
	w.quitting.Store(true)
}

func (w *windowManager) IsQuitting() bool {
	return w.quitting.Load()
}
