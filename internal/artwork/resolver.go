package artwork

import (
	"context"
	"net/http"
	"time"
)

type Resolver struct {
	http      *http.Client
	cache     *lruCache
	userAgent string

	itunesBase string
	mbBase     string
	caaBase    string

	mbLimiter *rateLimiter

	sources []Source
}

type Source interface {
	Lookup(ctx context.Context, q Query) string
	Name() string
}

type SourceFunc struct {
	SourceName string
	Fn         func(ctx context.Context, q Query) string
}

func (f SourceFunc) Lookup(ctx context.Context, q Query) string {
	return f.Fn(ctx, q)
}

func (f SourceFunc) Name() string { return f.SourceName }

type Option func(*Resolver)

func WithHTTPClient(c *http.Client) Option { return func(r *Resolver) { r.http = c } }

func WithUserAgent(ua string) Option { return func(r *Resolver) { r.userAgent = ua } }

func WithBaseURLs(itunes, musicbrainz, coverart string) Option {
	return func(r *Resolver) {
		r.itunesBase = itunes
		r.mbBase = musicbrainz
		r.caaBase = coverart
	}
}

func WithMusicBrainzInterval(d time.Duration) Option {
	return func(r *Resolver) { r.mbLimiter = newRateLimiter(d) }
}

func WithSources(sources ...Source) Option {
	return func(r *Resolver) { r.sources = sources }
}

func AppendSource(source Source) Option {
	return func(r *Resolver) {
		if source != nil {
			r.sources = append(r.sources, source)
		}
	}
}

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
	r.sources = []Source{
		SourceFunc{SourceName: "itunes", Fn: r.resolveITunes},
		SourceFunc{SourceName: "coverart", Fn: r.resolveCoverArt},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Resolver) Cached(q Query) (string, bool) {
	q = q.normalized()
	if q.isEmpty() {
		return "", false
	}
	return r.cache.get(q.cacheKey())
}

func (r *Resolver) Resolve(ctx context.Context, q Query) (string, error) {
	q = q.normalized()
	if q.isEmpty() {
		return "", nil
	}
	key := q.cacheKey()
	if url, ok := r.cache.get(key); ok {
		return url, nil
	}

	for _, source := range r.sources {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if url := source.Lookup(ctx, q); url != "" {
			r.cache.put(key, url)
			return url, nil
		}
	}

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	r.cache.put(key, "")
	return "", nil
}
