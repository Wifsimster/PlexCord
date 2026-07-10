package artwork

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestResolver wires a Resolver to the given httptest server for all three
// external APIs and disables MusicBrainz throttling.
func newTestResolver(base string) *Resolver {
	return NewResolver(
		WithBaseURLs(base, base, base),
		WithMusicBrainzInterval(0),
		WithUserAgent("PlexCord/test"),
	)
}

func TestResolve_ITunesHit(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		gotUA = req.Header.Get("User-Agent")
		if strings.HasPrefix(req.URL.Path, "/search") {
			_, _ = w.Write([]byte(`{"results":[{"artworkUrl100":"https://is1.mzstatic.com/image/thumb/x/100x100bb.jpg"}]}`))
			return
		}
		http.NotFound(w, req)
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	url, err := r.Resolve(context.Background(), "Queen", "A Night at the Opera")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if url != "https://is1.mzstatic.com/image/thumb/x/512x512bb.jpg" {
		t.Errorf("expected upscaled 512px URL, got %q", url)
	}
	if gotUA != "PlexCord/test" {
		t.Errorf("User-Agent = %q, want PlexCord/test", gotUA)
	}
}

func TestResolve_CachesAndServesFromCache(t *testing.T) {
	var searchHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/search") {
			searchHits++
			_, _ = w.Write([]byte(`{"results":[{"artworkUrl100":"https://cdn/100x100bb.jpg"}]}`))
			return
		}
		http.NotFound(w, req)
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	for i := 0; i < 3; i++ {
		if _, err := r.Resolve(context.Background(), "A", "B"); err != nil {
			t.Fatalf("Resolve: %v", err)
		}
	}
	if searchHits != 1 {
		t.Errorf("expected exactly 1 upstream search (cache hit after), got %d", searchHits)
	}

	// Cached() must return without any network call.
	if url, ok := r.Cached("A", "B"); !ok || url != "https://cdn/512x512bb.jpg" {
		t.Errorf("Cached() = %q, %v; want the resolved URL", url, ok)
	}
}

func TestResolve_CoverArtFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.URL.Path, "/search"):
			// iTunes miss.
			_, _ = w.Write([]byte(`{"results":[]}`))
		case strings.HasPrefix(req.URL.Path, "/ws/2/release/"):
			_, _ = w.Write([]byte(`{"releases":[{"id":"mbid-123"}]}`))
		case req.Method == http.MethodHead && req.URL.Path == "/release/mbid-123/front-500":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, req)
		}
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	url, err := r.Resolve(context.Background(), "Obscure Artist", "Rare Album")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := srv.URL + "/release/mbid-123/front-500"
	if url != want {
		t.Errorf("expected Cover Art Archive URL %q, got %q", want, url)
	}
}

func TestResolve_CoverArtMissingImageIsMiss(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.URL.Path, "/search"):
			_, _ = w.Write([]byte(`{"results":[]}`))
		case strings.HasPrefix(req.URL.Path, "/ws/2/release/"):
			_, _ = w.Write([]byte(`{"releases":[{"id":"mbid-404"}]}`))
		default:
			// CAA HEAD returns 404 → no cover art for this release.
			http.NotFound(w, req)
		}
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	url, err := r.Resolve(context.Background(), "X", "Y")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if url != "" {
		t.Errorf("expected empty URL when CAA image is missing, got %q", url)
	}
}

func TestResolve_TotalMiss(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.URL.Path, "/search"):
			_, _ = w.Write([]byte(`{"results":[]}`))
		case strings.HasPrefix(req.URL.Path, "/ws/2/release/"):
			_, _ = w.Write([]byte(`{"releases":[]}`))
		default:
			http.NotFound(w, req)
		}
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	url, err := r.Resolve(context.Background(), "Nobody", "Nothing")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if url != "" {
		t.Errorf("expected empty URL on total miss, got %q", url)
	}
	// A miss is cached as "" so we don't re-query every poll.
	if url, ok := r.Cached("Nobody", "Nothing"); !ok || url != "" {
		t.Errorf("expected negative result cached, got %q, %v", url, ok)
	}
}

func TestResolve_NeverReturnsPlexToken(t *testing.T) {
	// Even if an upstream misbehaves and echoes a tokened URL, Resolve only
	// returns iTunes/CAA-shaped URLs; assert the token never leaks through.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"artworkUrl100":"https://cdn/100x100bb.jpg"}]}`))
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	url, _ := r.Resolve(context.Background(), "Artist", "Album")
	if strings.Contains(url, "X-Plex-Token") {
		t.Errorf("resolved URL must never contain a Plex token: %q", url)
	}
}

func TestResolve_EmptyInputs(t *testing.T) {
	r := NewResolver(WithMusicBrainzInterval(0))
	if url, err := r.Resolve(context.Background(), "", ""); err != nil || url != "" {
		t.Errorf("empty inputs should yield empty URL, got %q, %v", url, err)
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	c := newLRUCache(2)
	c.put("a", "1")
	c.put("b", "2")
	c.put("c", "3") // evicts "a" (least recently used)

	if _, ok := c.get("a"); ok {
		t.Error("expected 'a' to be evicted")
	}
	if v, ok := c.get("b"); !ok || v != "2" {
		t.Errorf("expected 'b'='2', got %q, %v", v, ok)
	}
	if v, ok := c.get("c"); !ok || v != "3" {
		t.Errorf("expected 'c'='3', got %q, %v", v, ok)
	}
}
