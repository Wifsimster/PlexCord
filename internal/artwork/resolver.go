// Package artwork resolves publicly reachable album/poster artwork URLs for a
// media session using keyless public APIs (iTunes Search, MusicBrainz + Cover
// Art Archive).
//
// It exists so PlexCord can show real cover art on a Discord profile without
// ever sending Discord the LAN Plex URL — which Discord's media proxy cannot
// fetch and which embeds the Plex token (a credential leak). Every URL this
// package returns is a public HTTPS URL; it never returns a Plex URL.
package artwork

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// Resolver resolves artwork URLs through an ordered chain: cache → iTunes →
// MusicBrainz/Cover Art Archive → miss. It is safe for concurrent use.
type Resolver struct {
	http      *http.Client
	cache     *lruCache
	userAgent string

	// Base URLs are fields so tests can point them at httptest servers.
	itunesBase string
	mbBase     string
	caaBase    string

	// mbLimiter throttles MusicBrainz requests to respect its 1 req/s policy.
	mbLimiter *rateLimiter
}

// Option configures a Resolver.
type Option func(*Resolver)

// WithHTTPClient overrides the HTTP client (e.g. to inject a test transport).
func WithHTTPClient(c *http.Client) Option { return func(r *Resolver) { r.http = c } }

// WithUserAgent overrides the User-Agent sent to external APIs.
func WithUserAgent(ua string) Option { return func(r *Resolver) { r.userAgent = ua } }

// WithBaseURLs overrides the external API base URLs (used by tests).
func WithBaseURLs(itunes, musicbrainz, coverart string) Option {
	return func(r *Resolver) {
		r.itunesBase = itunes
		r.mbBase = musicbrainz
		r.caaBase = coverart
	}
}

// WithMusicBrainzInterval sets the minimum spacing between MusicBrainz requests.
// Tests pass 0 to disable throttling.
func WithMusicBrainzInterval(d time.Duration) Option {
	return func(r *Resolver) { r.mbLimiter = newRateLimiter(d) }
}

// NewResolver builds a Resolver with sensible production defaults.
func NewResolver(opts ...Option) *Resolver {
	r := &Resolver{
		http:       &http.Client{Timeout: 5 * time.Second},
		cache:      newLRUCache(512),
		userAgent:  "PlexCord",
		itunesBase: "https://itunes.apple.com",
		mbBase:     "https://musicbrainz.org",
		caaBase:    "https://coverartarchive.org",
		mbLimiter:  newRateLimiter(time.Second),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// cacheKey builds a stable, case-insensitive key for an artist/album pair.
func cacheKey(artist, album string) string {
	return strings.ToLower(strings.TrimSpace(artist)) + "\x00" + strings.ToLower(strings.TrimSpace(album))
}

// Cached returns a previously resolved URL for artist/album without performing
// any network request. The second result reports whether the pair was cached.
// It is used for the synchronous fast path so a known cover shows instantly.
func (r *Resolver) Cached(artist, album string) (string, bool) {
	if artist == "" && album == "" {
		return "", false
	}
	return r.cache.get(cacheKey(artist, album))
}

// Resolve returns a public HTTPS artwork URL for the given artist/album, or an
// empty string if none is found. Results (including misses) are cached. The
// returned URL is never a Plex URL and never contains a Plex token.
func (r *Resolver) Resolve(ctx context.Context, artist, album string) (string, error) {
	if strings.TrimSpace(artist) == "" && strings.TrimSpace(album) == "" {
		return "", nil
	}
	key := cacheKey(artist, album)
	if url, ok := r.cache.get(key); ok {
		return url, nil
	}

	// 1) iTunes Search — fast, high coverage for mainstream music.
	if url := r.resolveITunes(ctx, artist, album); url != "" {
		r.cache.put(key, url)
		return url, nil
	}

	// 2) MusicBrainz + Cover Art Archive — keyless fallback.
	if url := r.resolveCoverArt(ctx, artist, album); url != "" {
		r.cache.put(key, url)
		return url, nil
	}

	// 3) Miss — cache the negative result so we don't re-query every poll.
	r.cache.put(key, "")
	return "", nil
}
