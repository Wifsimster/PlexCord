package plex

import (
	"context"
	"log"
	"time"
)

func runPollLoop[T any](
	ctx context.Context,
	stopCh <-chan struct{},
	getInterval func() time.Duration,
	fetch func() (T, bool),
	changed func(prev, curr T) bool,
	emit func(T),
	label string,
) {
	interval := getInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastSession T
	var isZero = true

	if session, ok := fetch(); ok {
		emit(session)
		lastSession = session
		isZero = false
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s poller stopped: context cancelled", label)
			return
		case <-stopCh:
			log.Printf("%s poller stopped: stop signal received", label)
			return
		case <-ticker.C:
			newInterval := getInterval()
			if newInterval != interval {
				ticker.Reset(newInterval)
				interval = newInterval
				log.Printf("%s poller interval changed to %v", label, interval)
			}

			session, ok := fetch()
			if !ok {
				continue
			}
			if isZero || changed(lastSession, session) {
				emit(session)
				lastSession = session
				isZero = false
			}
		}
	}
}
