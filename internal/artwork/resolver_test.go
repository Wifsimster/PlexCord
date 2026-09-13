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
	url, err := r.Resolve(context.Background(), MusicQuery("Queen", "A Night at the Opera"))
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
		if _, err := r.Resolve(context.Background(), MusicQuery("A", "B")); err != nil {
			t.Fatalf("Resolve: %v", err)
		}
	}
	if searchHits != 1 {
		t.Errorf("expected exactly 1 upstream search (cache hit after), got %d", searchHits)
	}

	// Cached() must return without any network call.
	if url, ok := r.Cached(MusicQuery("A", "B")); !ok || url != "https://cdn/512x512bb.jpg" {
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
	url, err := r.Resolve(context.Background(), MusicQuery("Obscure Artist", "Rare Album"))
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
	url, err := r.Resolve(context.Background(), MusicQuery("X", "Y"))
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
	url, err := r.Resolve(context.Background(), MusicQuery("Nobody", "Nothing"))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if url != "" {
		t.Errorf("expected empty URL on total miss, got %q", url)
	}
	// A miss is cached as "" so we don't re-query every poll.
	if url, ok := r.Cached(MusicQuery("Nobody", "Nothing")); !ok || url != "" {
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
	url, _ := r.Resolve(context.Background(), MusicQuery("Artist", "Album"))
	if strings.Contains(url, "X-Plex-Token") {
		t.Errorf("resolved URL must never contain a Plex token: %q", url)
	}
}

func TestResolve_EmptyInputs(t *testing.T) {
	r := NewResolver(WithMusicBrainzInterval(0))
	if url, err := r.Resolve(context.Background(), MusicQuery("", "")); err != nil || url != "" {
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

func TestResolve_MoviePoster(t *testing.T) {
	var gotEntity, gotTerm string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		gotEntity = req.URL.Query().Get("entity")
		gotTerm = req.URL.Query().Get("term")
		_, _ = w.Write([]byte(`{"results":[{"artworkUrl100":"https://cdn/poster/100x100bb.jpg"}]}`))
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	url, err := r.Resolve(context.Background(), MovieQuery("Blade Runner", 1982))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if url != "https://cdn/poster/512x512bb.jpg" {
		t.Errorf("Resolve() = %q, want the upscaled poster", url)
	}
	if gotEntity != "movie" {
		t.Errorf("entity = %q, want movie — an album search returns the wrong artwork", gotEntity)
	}
	if gotTerm != "Blade Runner" {
		t.Errorf("term = %q, want the film title", gotTerm)
	}
}

func TestResolve_ShowPosterSearchesTheShowNotTheEpisode(t *testing.T) {
	var gotEntity, gotTerm string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		gotEntity = req.URL.Query().Get("entity")
		gotTerm = req.URL.Query().Get("term")
		_, _ = w.Write([]byte(`{"results":[{"artworkUrl100":"https://cdn/show/100x100bb.jpg"}]}`))
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	url, err := r.Resolve(context.Background(), ShowQuery("Severance", 2022))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if url != "https://cdn/show/512x512bb.jpg" {
		t.Errorf("Resolve() = %q, want the upscaled show art", url)
	}
	if gotEntity != "tvSeason" {
		t.Errorf("entity = %q, want tvSeason", gotEntity)
	}
	if gotTerm != "Severance" {
		t.Errorf("term = %q, want the show title", gotTerm)
	}
}

func TestResolve_VideoSkipsCoverArtArchive(t *testing.T) {
	// Cover Art Archive indexes music releases only. A film that iTunes misses
	// must not cost a MusicBrainz round trip that can never answer.
	var mbHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/ws/2/release/") {
			mbHits++
		}
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	if _, err := r.Resolve(context.Background(), MovieQuery("An Unlisted Film", 0)); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if mbHits != 0 {
		t.Errorf("a movie lookup queried MusicBrainz %d time(s); it indexes music only", mbHits)
	}
}

func TestResolve_MediaTypeSeparatesCacheEntries(t *testing.T) {
	// A film and an album that share a name are different pictures.
	var searches int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		searches++
		switch req.URL.Query().Get("entity") {
		case "movie":
			_, _ = w.Write([]byte(`{"results":[{"artworkUrl100":"https://cdn/film/100x100bb.jpg"}]}`))
		default:
			_, _ = w.Write([]byte(`{"results":[{"artworkUrl100":"https://cdn/album/100x100bb.jpg"}]}`))
		}
	}))
	defer srv.Close()

	r := newTestResolver(srv.URL)
	album, _ := r.Resolve(context.Background(), MusicQuery("", "Purple Rain"))
	film, _ := r.Resolve(context.Background(), MovieQuery("Purple Rain", 1984))

	if album == film {
		t.Errorf("album and film resolved to the same URL %q — the cache key ignores the media type", album)
	}
	if searches != 2 {
		t.Errorf("upstream searched %d time(s), want 2 — one per media type", searches)
	}
}

func TestResolve_EmptyVideoTitleIsAMiss(t *testing.T) {
	r := NewResolver(WithMusicBrainzInterval(0))
	if url, err := r.Resolve(context.Background(), MovieQuery("", 0)); err != nil || url != "" {
		t.Errorf("Resolve() = (%q, %v), want a clean miss with nothing to search on", url, err)
	}
}
