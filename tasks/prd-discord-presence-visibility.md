# PRD: Discord Presence Visibility — "Listening to" Activity & Public Album Art

## Introduction

PlexCord's Discord Rich Presence currently renders as a generic game activity: the
profile header reads "Playing PlexCord" ("Joue à PlexCord"), the album cover shows a
"?" placeholder, and there is no progress bar. This PRD upgrades the presence to a
Spotify-style music card: **"Listening to PlexCord"** header, a **visible album
cover**, artist/track in the **member list**, and a **live progress bar** — while
also fixing a token-leak issue in the current artwork URL.

### Root causes identified

| Symptom | Root cause | Where |
|---|---|---|
| Album cover shows "?" | Discord's media proxy fetches `large_image` URLs from Discord's servers. We send `http://<LAN-IP>:32400/library/...?X-Plex-Token=...`, which is unreachable from the internet. | `internal/plex/client.go` → `buildArtworkURL()` |
| Plex token exposure | The `X-Plex-Token` is embedded in the presence payload, which any Discord user who can see the activity can inspect. | same |
| "Playing" instead of "Listening to" | `github.com/hugolgst/rich-go` hardcodes activity type 0 (Playing). Discord IPC supports `type: 2` (Listening) and `type: 3` (Watching) since mid-2024. | `internal/discord/presence.go`, rich-go `client/types.go` |
| Member list shows "PlexCord" instead of the artist/track | `status_display_type` is not sent (rich-go has no field for it). | same |
| No progress bar | Only the `start` timestamp is sent (and only while playing). Listening activities render a progress bar when both `start` and `end` are present. | `internal/discord/builder.go` → `applyTimestamps()` |

## Goals

- Music sessions display as **"Listening to PlexCord"** with track title, artist, album, album cover, and a live progress bar.
- Movie/TV sessions display as **"Watching PlexCord"** with poster artwork.
- Member list / status line can show the track or artist (e.g. "Listening to **Def Leppard**") instead of "PlexCord".
- Album covers resolve to **publicly reachable URLs** so they actually render — without ever leaking the Plex token.
- All new behavior is configurable from Settings, with safe defaults, and previewable in the existing `DiscordSpecimen` component.
- Remove the unmaintained `rich-go` dependency in favor of a small internal IPC client we control.

## Non-Goals

- No Discord bot / OAuth features (join/spectate secrets, parties).
- No artwork uploads to third-party image hosts requiring accounts or API keys (Imgur, etc.).
- No change to the Plex polling architecture or event bus.

## User Stories

### US-001: Internal Discord IPC client with activity-type support
**Description:** As a developer, I want PlexCord to own the Discord IPC layer so we can send fields rich-go doesn't support (`type`, `status_display_type`) and parse responses instead of ignoring them.

**Implementation notes:**
- New package `internal/discord/ipc`: handshake (`op 0`) + `SET_ACTIVITY` frame (`op 1`) over the platform socket (unix socket `$XDG_RUNTIME_DIR|$TMPDIR/discord-ipc-{0..9}`, Windows named pipe `\\.\pipe\discord-ipc-{0..9}`). The rich-go implementation being replaced is ~150 lines total; port it, then extend the payload:
  ```go
  type PayloadActivity struct {
      Type              int  `json:"type"`                          // 0 Playing, 2 Listening, 3 Watching
      StatusDisplayType *int `json:"status_display_type,omitempty"` // 0 name, 1 state, 2 details
      // ... existing details/state/assets/timestamps/buttons fields
  }
  ```
- Parse the SET_ACTIVITY response frame and surface `evt: ERROR` payloads as Go errors (rich-go leaves `// TODO: Response should be parsed`), so connection-lost detection in `PresenceManager` stops relying on string matching alone.
- `internal/discord/presence.go` switches imports from `rich-go/client` to the new package; `go.mod` drops `hugolgst/rich-go`.

**Acceptance Criteria:**
- [ ] `internal/discord/ipc` connects, handshakes, and sets activity on Linux/macOS (unix socket) and Windows (named pipe)
- [ ] Activity payload supports `type` and `status_display_type`
- [ ] Frame (de)serialization covered by unit tests (no live Discord needed)
- [ ] `github.com/hugolgst/rich-go` removed from `go.mod`
- [ ] Existing `PresenceManager` tests still pass

### US-002: "Listening to" / "Watching" activity types per media type
**Description:** As a user, I want my Discord profile to say "Listening to PlexCord" when I play music and "Watching PlexCord" for movies/TV, so the presence reads like a media activity instead of a game.

**Implementation notes:**
- `internal/discord/builder.go`: each `PresenceBuilder` sets the activity type — `musicBuilder` → Listening (2), `movieBuilder`/`tvBuilder` → Watching (3).
- Config: `activityStyle: "media" | "game"` (default `media`) to preserve today's look for users who prefer it.
- `status_display_type` from config `statusDisplay: "app" | "state" | "details"` (default `state`, so the member list shows "Listening to *by Def Leppard • Hysteria*" line → recommend default `details` = track name; decide during implementation with real Discord rendering, both behind one setting).

**Acceptance Criteria:**
- [ ] Music sessions produce activity `type = 2`, movie/TV produce `type = 3`
- [ ] Setting to fall back to classic "Playing" (type 0) style
- [ ] Member-list line configurable (app name / track / artist line)
- [ ] Builder unit tests assert type + display fields per media type

### US-003: Progress bar via start + end timestamps
**Description:** As a user, I want a live progress bar on the music card like Spotify shows.

**Implementation notes:**
- `applyTimestamps()` sends both timestamps while playing:
  `start = now − position`, `end = start + duration` (fields already on `PresenceData`).
- When paused: keep the pause icon behavior, omit timestamps (Discord cannot freeze a progress bar; today's elapsed-timer omission logic stays).
- Guard: only send `end` when `duration > 0` (streams/unknown durations fall back to elapsed-only).

**Acceptance Criteria:**
- [ ] Playing music session shows progress bar in Discord (manual verification)
- [ ] `duration == 0` sends start-only timestamps
- [ ] Paused sessions send no timestamps (unchanged)
- [ ] Unit tests for the three timestamp cases

### US-004: Public album-art resolver (fixes "?" cover and token leak)
**Description:** As a user, I want the actual album cover to appear on my Discord profile, and as a security-conscious user I never want my Plex token broadcast in the presence payload.

**Implementation notes:**
- New package `internal/artwork` with a resolver chain returning a public HTTPS URL:
  1. **Cache** — LRU keyed by `artist|album` (music) / `title|year` (video), persisted to the config dir so lookups survive restarts.
  2. **iTunes Search API** — keyless: `https://itunes.apple.com/search?term={artist}+{album}&entity=album&limit=1`, take `artworkUrl100` and rewrite to `512x512bb.jpg`. Fast, high coverage for mainstream music.
  3. **MusicBrainz + Cover Art Archive** — keyless fallback: release search by artist+album, then `https://coverartarchive.org/release/{mbid}/front-500`. Respect the 1 req/s rate limit and send a proper `User-Agent` (`PlexCord/{version}`).
  4. **Fallback** — the existing `plex` uploaded asset (current behavior when no artwork).
- For movies/TV, iTunes `entity=movie|tvSeason` works the same way; ship music first, video posters as a stretch within this story.
- **The tokened Plex URL is never sent to Discord again** (security fix). It remains in `MediaSession.ThumbURL` for the local dashboard preview only. `updateDiscordFromSession` passes the resolved public URL instead.
- Async flow: on track change, set presence immediately with the fallback asset; resolve artwork in a goroutine (~sub-second for cache/iTunes) and re-issue `SetPresence` with the cover once resolved. Debounce so a resolve landing after the track changed again is dropped (compare session key).
- Privacy: config `artworkLookup: true|false` (default `true`); when off, skip external lookups entirely — Settings copy explains that artist/album names are sent to iTunes/MusicBrainz.

**Acceptance Criteria:**
- [ ] Album cover renders on Discord profile for a mainstream album (manual verification)
- [ ] Presence payload never contains `X-Plex-Token` (assert in unit test on built activity)
- [ ] Resolver chain covered by `httptest`-based unit tests, including cache hit, iTunes hit, CAA fallback, and total miss
- [ ] Stale resolutions (track changed mid-lookup) are discarded
- [ ] Lookup can be disabled in Settings; disabled mode sends the `plex` asset
- [ ] Large-image hover text remains the album name

### US-005: Settings & Dashboard preview
**Description:** As a user, I want to control the new presence style from Settings and see an accurate preview of what Discord will show.

**Implementation notes:**
- `Settings.vue` (Discord section): activity style (Listening/Watching vs classic Playing), member-list line selector, artwork lookup toggle. New Wails bindings: `GetPresenceOptions` / `SetPresenceOptions` in `app_discord.go`, persisted via `internal/config`.
- `DiscordSpecimen.vue`: render the Listening-style card — "Écoute/Listening to PlexCord" header, cover art, track/artist/album lines, progress bar — driven by the same store (`stores/presence.js`), so the preview matches the new default.
- `SetupDiscord.vue` wizard copy updated to mention the Listening display.

**Acceptance Criteria:**
- [ ] New settings persist across restarts and take effect on the next presence update (no reconnect required)
- [ ] Dashboard specimen shows progress bar + cover matching the active session
- [ ] `presenceFormat` custom format strings keep working with the new builders
- [ ] Frontend unit tests for the store/format changes

## Technical Considerations

- **Rate limits:** Discord IPC allows presence updates every ~15 s per client but tolerates the current poll cadence; the artwork re-issue adds at most one extra `SET_ACTIVITY` per track change. Keep the existing debounce in the poll runner.
- **Discord client compatibility:** `type` 2/3 over IPC works on desktop stable since 2024; older clients silently render as Playing — graceful degradation, no version gate needed. `status_display_type` is newer (2025); unknown-field behavior is also silent fallback.
- **Failure isolation:** artwork resolver failures must never block or fail a presence update; log at debug level and use fallback asset.
- **Testing without Discord:** all IPC framing is tested against an in-process fake socket; only two manual verification points (card rendering, cover rendering) remain.

## Suggested Delivery Order

1. **PR 1 — IPC ownership + activity types + progress bar** (US-001, US-002, US-003): biggest visual win, no new external services.
2. **PR 2 — artwork resolver + token-leak fix** (US-004).
3. **PR 3 — settings & preview polish** (US-005).

## Success Metrics

- Discord profile shows "Listening to PlexCord" with cover art and progress bar for a playing track (screenshot parity with Spotify-style cards).
- Zero occurrences of `X-Plex-Token` in outgoing presence payloads.
- No regression in existing presence, pause/hide, and reconnect test suites.
