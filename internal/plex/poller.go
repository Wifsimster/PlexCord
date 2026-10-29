package plex

import (
	"context"
	"log"
	"sync"
	"time"
)

type Poller struct {
	lastErrorTime time.Time
	source        SessionSource
	stopCh        chan struct{}
	sessionC      chan *MusicSession
	mediaC        chan *MediaSession

	onError     func(err error)
	onRecovered func()

	userID     string
	interval   time.Duration
	mediaTypes []string

	mu           sync.RWMutex
	running      bool
	inErrorState bool
	mediaMode    bool
}

func NewPoller(source SessionSource, userID string, interval time.Duration) *Poller {
	if interval < time.Second {
		interval = time.Second
	}
	if interval > 60*time.Second {
		interval = 60 * time.Second
	}

	return &Poller{
		source:   source,
		userID:   userID,
		interval: interval,
		stopCh:   make(chan struct{}),
		sessionC: make(chan *MusicSession, 1),
		mediaC:   make(chan *MediaSession, 1),
	}
}

func (p *Poller) SetMediaTypes(types []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mediaTypes = types
}

func (p *Poller) GetMediaTypes() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.mediaTypes
}

func (p *Poller) Start(ctx context.Context) <-chan *MusicSession {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return p.sessionC
	}
	p.running = true
	p.stopCh = make(chan struct{})
	p.sessionC = make(chan *MusicSession, 1)
	p.mu.Unlock()

	go p.pollLoop(ctx)

	return p.sessionC
}

func (p *Poller) StartMedia(ctx context.Context) <-chan *MediaSession {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return p.mediaC
	}
	p.running = true
	p.mediaMode = true
	p.stopCh = make(chan struct{})
	p.mediaC = make(chan *MediaSession, 1)
	p.mu.Unlock()

	go p.mediaPollLoop(ctx)

	return p.mediaC
}

func (p *Poller) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}

	p.running = false
	p.mediaMode = false
	close(p.stopCh)
}

func (p *Poller) SetInterval(interval time.Duration) {
	if interval < time.Second {
		interval = time.Second
	}
	if interval > 60*time.Second {
		interval = 60 * time.Second
	}

	p.mu.Lock()
	p.interval = interval
	p.mu.Unlock()
}

func (p *Poller) GetInterval() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.interval
}

func (p *Poller) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

func (p *Poller) SetErrorCallbacks(onError func(err error), onRecovered func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onError = onError
	p.onRecovered = onRecovered
}

func (p *Poller) IsInErrorState() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.inErrorState
}

func (p *Poller) pollLoop(ctx context.Context) {
	p.mu.Lock()
	sessionC := p.sessionC
	stopCh := p.stopCh
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		if p.sessionC == sessionC {
			p.running = false
		}
		p.mu.Unlock()
		close(sessionC)
	}()

	runPollLoop[*MusicSession](
		ctx,
		stopCh,
		p.GetInterval,
		p.doPoll,
		sessionChanged,
		func(session *MusicSession) {
			select {
			case sessionC <- session:
			default:
				log.Printf("Session channel full, skipping update")
			}
		},
		"Music",
	)
}

func (p *Poller) doPoll() (*MusicSession, bool) {
	return pollOnce(p, "Poll", func() ([]MusicSession, error) {
		return p.source.GetMusicSessions(p.userID)
	})
}

func pollOnce[T any](p *Poller, label string, fetch func() ([]T, error)) (*T, bool) {
	sessions, err := fetch()
	if err != nil {
		log.Printf("%s error: %v", label, err)
		p.recordPollFailure(err)
		return nil, false
	}
	p.recordPollSuccess()

	if len(sessions) == 0 {
		return nil, true
	}
	return &sessions[0], true
}

func (p *Poller) recordPollFailure(err error) {
	p.mu.Lock()
	wasInErrorState := p.inErrorState
	p.inErrorState = true
	p.lastErrorTime = time.Now()
	onError := p.onError
	p.mu.Unlock()

	if !wasInErrorState && onError != nil {
		onError(err)
	}
}

func (p *Poller) recordPollSuccess() {
	p.mu.Lock()
	wasInErrorState := p.inErrorState
	p.inErrorState = false
	onRecovered := p.onRecovered
	p.mu.Unlock()

	if wasInErrorState && onRecovered != nil {
		log.Printf("Plex connection recovered")
		onRecovered()
	}
}

func (p *Poller) mediaPollLoop(ctx context.Context) {
	p.mu.Lock()
	mediaC := p.mediaC
	stopCh := p.stopCh
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		if p.mediaC == mediaC {
			p.running = false
			p.mediaMode = false
		}
		p.mu.Unlock()
		close(mediaC)
	}()

	runPollLoop[*MediaSession](
		ctx,
		stopCh,
		p.GetInterval,
		p.doMediaPoll,
		mediaSessionChanged,
		func(session *MediaSession) {
			select {
			case mediaC <- session:
			default:
				log.Printf("Media session channel full, skipping update")
			}
		},
		"Media",
	)
}

func (p *Poller) doMediaPoll() (*MediaSession, bool) {
	mediaTypes := p.GetMediaTypes()
	return pollOnce(p, "Media poll", func() ([]MediaSession, error) {
		return p.source.GetMediaSessions(p.userID, mediaTypes)
	})
}
