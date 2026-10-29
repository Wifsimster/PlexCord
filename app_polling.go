package main

import (
	"context"
	"log"
	"sync"
	"time"

	"plexcord/internal/plex"
)

type sessionCache struct {
	mu      sync.RWMutex
	current *plex.MediaSession
}

func (c *sessionCache) Set(session *plex.MediaSession) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = session
}

func (c *sessionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = nil
}

func (c *sessionCache) Get() *plex.MediaSession {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current
}

type pollingController struct {
	mu     sync.Mutex
	poller *plex.Poller
	stop   context.CancelFunc
}

type pollingConfig struct {
	Source      plex.SessionSource
	UserID      string
	MediaTypes  []string
	Interval    time.Duration
	OnError     func(err error)
	OnRecovered func()
}

func (p *pollingController) Start(cfg pollingConfig) <-chan *plex.MediaSession {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.poller != nil && p.poller.IsRunning() {
		log.Printf("Session polling already running")
		return nil
	}

	poller := plex.NewPoller(cfg.Source, cfg.UserID, cfg.Interval)
	poller.SetErrorCallbacks(cfg.OnError, cfg.OnRecovered)
	poller.SetMediaTypes(cfg.MediaTypes)

	ctx, cancel := context.WithCancel(context.Background())
	p.poller = poller
	p.stop = cancel

	return poller.StartMedia(ctx)
}

func (p *pollingController) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.poller == nil {
		return
	}
	if p.stop != nil {
		p.stop()
	}
	p.poller.Stop()
	p.poller = nil
	p.stop = nil

	log.Printf("Session polling stopped")
}

func (p *pollingController) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.poller != nil && p.poller.IsRunning()
}

func (p *pollingController) State() (running, inError bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.poller == nil {
		return false, false
	}
	return p.poller.IsRunning(), p.poller.IsInErrorState()
}

func (p *pollingController) SetInterval(interval time.Duration) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.poller == nil || !p.poller.IsRunning() {
		return false
	}
	p.poller.SetInterval(interval)
	return true
}
