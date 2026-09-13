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

	// sources is the ordered lookup chain, tried until one returns a URL.
	// Holding the chain as data rather than as a hard-coded sequence of calls
	// is what lets a new provider be added — or the order changed, or a
	// provider dropped — without editing Resolve (OCP).
	sources []Source
}

// Source is one place artwork can be looked up. Implementations must return an
// empty string on a miss (not an error) and must respect ctx cancellation. A
// source that does not cover the query's media type misses, which is how the
// music-only providers sit in the same chain as the video-capable ones.
//
// The interface is a single method so a plain function can be a Source; see
// SourceFunc.
type Source interface {
	// Lookup returns a public HTTPS artwork URL for the query, or "".
	Lookup(ctx context.Context, q Query) string
	// Name identifies the source in logs and tests.
	Name() string
}

// SourceFunc adapts a plain function to the Source interface.
type SourceFunc struct {
	SourceName string
	Fn         func(ctx context.Context, q Query) string
}

// Lookup calls the wrapped function.
func (f SourceFunc) Lookup(ctx context.Context, q Query) string {
	return f.Fn(ctx, q)
}

// Name returns the source's name.
func (f SourceFunc) Name() string { return f.SourceName }

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

// WithSources replaces the lookup chain. Sources are tried in order until one
// returns a URL; passing none disables lookups entirely (every pair misses).
func WithSources(sources ...Source) Option {
	return func(r *Resolver) { r.sources = sources }
}

// AppendSource adds a source to the end of the lookup chain, so it is consulted
// only after the built-in ones miss.
func AppendSource(source Source) Option {
	return func(r *Resolver) {
		if source != nil {
			r.sources = append(r.sources, source)
		}
	}
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
	// The default chain: iTunes first (fast, high coverage for mainstream
	// music), then MusicBrainz + Cover Art Archive as a keyless fallback.
	r.sources = []Source{
		SourceFunc{SourceName: "itunes", Fn: r.resolveITunes},
		SourceFunc{SourceName: "coverart", Fn: r.resolveCoverArt},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Cached returns a previously resolved URL for the query without performing any
// network request. The second result reports whether the query was cached.
// It is used for the synchronous fast path so a known cover shows instantly.
func (r *Resolver) Cached(q Query) (string, bool) {
	q = q.normalized()
	if q.isEmpty() {
		return "", false
	}
	return r.cache.get(q.cacheKey())
}

// Resolve returns a public HTTPS artwork URL for the given query — an album
// cover, a film poster, a show's art — or an empty string if none is found.
// Results (including misses) are cached. The returned URL is never a Plex URL
// and never contains a Plex token.
func (r *Resolver) Resolve(ctx context.Context, q Query) (string, error) {
	q = q.normalized()
	if q.isEmpty() {
		return "", nil
	}
	key := q.cacheKey()
	if url, ok := r.cache.get(key); ok {
		return url, nil
	}

	// Walk the chain in order; the first source with an answer wins.
	for _, source := range r.sources {
		if ctx.Err() != nil {
			// Caller gave up (or timed out): do not cache a miss we never
			// actually established, or the next poll would skip the lookup.
			return "", ctx.Err()
		}
		if url := source.Lookup(ctx, q); url != "" {
			r.cache.put(key, url)
			return url, nil
		}
	}

	// Every source missed — cache the negative result so we don't re-query
	// on every poll.
	r.cache.put(key, "")
	return "", nil
}
