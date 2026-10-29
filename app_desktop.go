package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

type ProcessController interface {
	Quit(ctx context.Context)
}

type BrowserOpener interface {
	OpenURL(ctx context.Context, url string)
}

type ScreenProvider interface {
	Screens(ctx context.Context) ([]runtime.Screen, error)
}

type Desktop interface {
	WindowController
	ProcessController
	BrowserOpener
	ScreenProvider
}

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

func newDesktop() Desktop { return wailsDesktop{} }
