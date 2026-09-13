package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// This file isolates every call into the Wails runtime behind narrow
// interfaces. Before, App called runtime.WindowShow / runtime.Quit /
// runtime.BrowserOpenURL directly, which meant none of the window, restore,
// quit or open-in-browser behaviour could be exercised without a live Wails
// context — and a nil context made those paths panic rather than fail.
//
// The interfaces are split by concern rather than lumped into one "runtime"
// interface, so a consumer depends only on the calls it makes (ISP): the
// window manager takes a WindowController, the screen-fitting code takes a
// ScreenProvider, and the two methods that open a browser take a BrowserOpener.

// WindowController drives the application window.
type WindowController interface {
	Show(ctx context.Context)
	Hide(ctx context.Context)
	Minimise(ctx context.Context)
	Unminimise(ctx context.Context)
	IsMinimised(ctx context.Context) bool
	SetAlwaysOnTop(ctx context.Context, onTop bool)
	SetSize(ctx context.Context, width, height int)
	Center(ctx context.Context)
}

// ProcessController terminates the application.
type ProcessController interface {
	Quit(ctx context.Context)
}

// BrowserOpener opens a URL in the user's default browser.
type BrowserOpener interface {
	OpenURL(ctx context.Context, url string)
}

// ScreenProvider reports the displays the window could be shown on.
type ScreenProvider interface {
	Screens(ctx context.Context) ([]runtime.Screen, error)
}

// Desktop is the union the production adapter implements. App holds its
// collaborators as the narrow interfaces above; this alias exists only so a
// single value can be injected for all of them.
type Desktop interface {
	WindowController
	ProcessController
	BrowserOpener
	ScreenProvider
}

// wailsDesktop is the production adapter, forwarding to the Wails runtime.
type wailsDesktop struct{}

func (wailsDesktop) Show(ctx context.Context)       { runtime.WindowShow(ctx) }
func (wailsDesktop) Hide(ctx context.Context)       { runtime.WindowHide(ctx) }
func (wailsDesktop) Minimise(ctx context.Context)   { runtime.WindowMinimise(ctx) }
func (wailsDesktop) Unminimise(ctx context.Context) { runtime.WindowUnminimise(ctx) }
func (wailsDesktop) IsMinimised(ctx context.Context) bool {
	return runtime.WindowIsMinimised(ctx)
}

func (wailsDesktop) SetAlwaysOnTop(ctx context.Context, onTop bool) {
	runtime.WindowSetAlwaysOnTop(ctx, onTop)
}

func (wailsDesktop) SetSize(ctx context.Context, width, height int) {
	runtime.WindowSetSize(ctx, width, height)
}

func (wailsDesktop) Center(ctx context.Context) { runtime.WindowCenter(ctx) }
func (wailsDesktop) Quit(ctx context.Context)   { runtime.Quit(ctx) }

func (wailsDesktop) OpenURL(ctx context.Context, url string) {
	runtime.BrowserOpenURL(ctx, url)
}

func (wailsDesktop) Screens(ctx context.Context) ([]runtime.Screen, error) {
	return runtime.ScreenGetAll(ctx)
}

// newDesktop returns the production Wails-backed Desktop.
func newDesktop() Desktop { return wailsDesktop{} }
