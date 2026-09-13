package artwork

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

// recordingSource is a lookup source that answers from memory and counts how
// often it was consulted.
type recordingSource struct {
	name  string
	url   string
	calls atomic.Int64
}

func (s *recordingSource) Lookup(context.Context, Query) string {
	s.calls.Add(1)
	return s.url
}

func (s *recordingSource) Name() string { return s.name }

// TestResolverWalksSourceChainInOrder verifies the chain is consulted in order
// and stops at the first hit — the later sources cost a network round trip
// each, so they must not run once an answer exists.
func TestResolverWalksSourceChainInOrder(t *testing.T) {
	first := &recordingSource{name: "first", url: "https://cdn/first.jpg"}
	second := &recordingSource{name: "second", url: "https://cdn/second.jpg"}

	r := NewResolver(WithSources(first, second))

	got, err := r.Resolve(context.Background(), MusicQuery("Artist", "Album"))
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if got != "https://cdn/first.jpg" {
		t.Errorf("Resolve() = %q, want the first source's answer", got)
	}
	if second.calls.Load() != 0 {
		t.Errorf("the second source ran %d time(s) despite the first answering", second.calls.Load())
	}
}

// TestResolverFallsThroughToLaterSources is the reason the chain exists: a
// source that misses hands off rather than ending the lookup.
func TestResolverFallsThroughToLaterSources(t *testing.T) {
	miss := &recordingSource{name: "miss"}
	hit := &recordingSource{name: "hit", url: "https://cdn/hit.jpg"}

	r := NewResolver(WithSources(miss, hit))

	got, err := r.Resolve(context.Background(), MusicQuery("Artist", "Album"))
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if got != "https://cdn/hit.jpg" {
		t.Errorf("Resolve() = %q, want the second source's answer", got)
	}
	if miss.calls.Load() != 1 {
		t.Errorf("the first source ran %d time(s), want 1", miss.calls.Load())
	}
}

// TestAppendSourceExtendsDefaultChain verifies a new provider can be added
// without touching the existing ones (OCP).
func TestAppendSourceExtendsDefaultChain(t *testing.T) {
	extra := &recordingSource{name: "extra", url: "https://cdn/extra.jpg"}

	// Replace the built-in chain with a miss so the appended source is reached
	// without any network traffic.
	r := NewResolver(
		WithSources(&recordingSource{name: "builtin-miss"}),
		AppendSource(extra),
	)

	got, err := r.Resolve(context.Background(), MusicQuery("Artist", "Album"))
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if got != "https://cdn/extra.jpg" {
		t.Errorf("Resolve() = %q, want the appended source's answer", got)
	}
}

// TestResolverCachesMisses verifies a total miss is remembered, so a track
// nobody has art for does not re-query every provider on every poll.
func TestResolverCachesMisses(t *testing.T) {
	miss := &recordingSource{name: "miss"}
	r := NewResolver(WithSources(miss))

	for range 3 {
		if _, err := r.Resolve(context.Background(), MusicQuery("Artist", "Album")); err != nil {
			t.Fatalf("Resolve() error: %v", err)
		}
	}

	if got := miss.calls.Load(); got != 1 {
		t.Errorf("the source was consulted %d times, want 1 — the miss should be cached", got)
	}
}

// TestResolverCachesHits verifies a resolved cover is served from memory, and
// that Cached sees it without any lookup.
func TestResolverCachesHits(t *testing.T) {
	hit := &recordingSource{name: "hit", url: "https://cdn/cover.jpg"}
	r := NewResolver(WithSources(hit))

	if _, ok := r.Cached(MusicQuery("Artist", "Album")); ok {
		t.Fatal("Cached() reported a hit before anything was resolved")
	}
	if _, err := r.Resolve(context.Background(), MusicQuery("Artist", "Album")); err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}

	url, ok := r.Cached(MusicQuery("Artist", "Album"))
	if !ok || url != "https://cdn/cover.jpg" {
		t.Errorf("Cached() = (%q, %v), want the resolved cover", url, ok)
	}
	if got := hit.calls.Load(); got != 1 {
		t.Errorf("the source was consulted %d times, want 1", got)
	}
}

// TestResolverStopsOnCancelledContext verifies a caller that gave up does not
// have a miss cached on its behalf — the lookup never actually concluded.
func TestResolverStopsOnCancelledContext(t *testing.T) {
	source := &recordingSource{name: "never-reached", url: "https://cdn/cover.jpg"}
	r := NewResolver(WithSources(source))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := r.Resolve(ctx, MusicQuery("Artist", "Album")); !errors.Is(err, context.Canceled) {
		t.Errorf("Resolve() error = %v, want context.Canceled", err)
	}
	if source.calls.Load() != 0 {
		t.Errorf("a cancelled lookup still consulted the source %d time(s)", source.calls.Load())
	}
	if _, ok := r.Cached(MusicQuery("Artist", "Album")); ok {
		t.Error("a cancelled lookup cached a miss it never established")
	}
}

// TestResolverWithNoSourcesMisses verifies an empty chain is a clean miss
// rather than a panic.
func TestResolverWithNoSourcesMisses(t *testing.T) {
	r := NewResolver(WithSources())

	got, err := r.Resolve(context.Background(), MusicQuery("Artist", "Album"))
	if err != nil || got != "" {
		t.Errorf("Resolve() = (%q, %v), want a clean miss", got, err)
	}
}
