package plex

import (
	"context"
	"log"
	"time"
)

// runPollLoop runs a generic polling loop parameterized by a fetch function,
// a change-detection function, and an emit function. This consolidates the
// previously duplicated logic between pollLoop (music) and mediaPollLoop
// (multi-media) into a single implementation (OCP + DRY).
//
// Type parameter T is the session type (*MusicSession or *MediaSession).
// The nil-check of the session goes through fetch's second return value
// (true = valid result including "no session"; false = error, skip update).
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

	// Perform immediate first poll
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
			// Refresh ticker if interval changed at runtime
			newInterval := getInterval()
			if newInterval != interval {
				ticker.Reset(newInterval)
				interval = newInterval
				log.Printf("%s poller interval changed to %v", label, interval)
			}

			session, ok := fetch()
			// Only emit if poll succeeded and session state changed.
			// On error, keep lastSession unchanged to avoid false "stopped" events.
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
