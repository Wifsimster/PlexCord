# Story 2.9: Track Metadata Extraction

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want PlexCord to extract track information from my Plex session,
So that accurate details can be displayed.

## Acceptance Criteria

1. **AC1: Track Title Extraction**
   - **Given** an active music session is detected
   - **When** metadata is extracted
   - **Then** the track title is captured from the Plex session
   - **And** the title is available in the MusicSession event payload
   - **And** missing titles show "Unknown Track" fallback

2. **AC2: Artist Name Extraction**
   - **Given** an active music session is detected
   - **When** metadata is extracted
   - **Then** the artist name is captured from `grandparentTitle` field
   - **And** the artist is available in the MusicSession event payload
   - **And** missing artist shows "Unknown Artist" fallback

3. **AC3: Album Name Extraction**
   - **Given** an active music session is detected
   - **When** metadata is extracted
   - **Then** the album name is captured from `parentTitle` field
   - **And** the album is available in the MusicSession event payload
   - **And** missing album shows "Unknown Album" fallback

4. **AC4: Album Artwork URL**
   - **Given** an active music session with album artwork
   - **When** metadata is extracted
   - **Then** the album artwork URL is captured from `thumb` attribute
   - **And** the URL is converted to an absolute URL (includes server base URL)
   - **And** missing artwork returns empty string (no fallback needed)
   - **And** artwork URL is accessible for display in frontend

5. **AC5: Track Duration**
   - **Given** an active music session is detected
   - **When** metadata is extracted
   - **Then** the track duration is captured in milliseconds
   - **And** duration is available for progress display
   - **And** missing duration defaults to 0

6. **AC6: Playback Position**
   - **Given** an active music session is detected
   - **When** metadata is extracted
   - **Then** the current playback position (viewOffset) is captured in milliseconds
   - **And** position updates with each poll
   - **And** position can be used to calculate elapsed time
   - **And** missing position defaults to 0

7. **AC7: Metadata Fallback Behavior**
   - **Given** a music session with incomplete metadata
   - **When** metadata fields are missing or empty
   - **Then** appropriate fallback values are used
   - **And** the application does not crash or error
   - **And** events are still emitted with available data

## Tasks / Subtasks

- [x] **Task 1: Add Fallback Constants** (AC: 1, 2, 3, 7)
  - [x] Add fallback constants to `internal/plex/types.go`:
    - `FallbackTrackTitle = "Unknown Track"`
    - `FallbackArtist = "Unknown Artist"`
    - `FallbackAlbum = "Unknown Album"`
  - [x] Add helper method `ApplyFallbacks()` to MusicSession struct
  - [x] Ensure fallbacks are applied after parsing Plex response

- [x] **Task 2: Implement Absolute Artwork URL** (AC: 4)
  - [x] Update `GetMusicSessions()` to build absolute artwork URLs
  - [x] Combine server base URL with thumb path: `{serverURL}{thumb}?X-Plex-Token={token}`
  - [x] Handle missing thumb gracefully (empty string)
  - [x] Add `ThumbURL` field to MusicSession for the complete URL

- [x] **Task 3: Update MusicSession Extraction** (AC: 1, 2, 3, 5, 6, 7)
  - [x] Verify Track field maps from `title` attribute
  - [x] Verify Artist field maps from `grandparentTitle` attribute
  - [x] Verify Album field maps from `parentTitle` attribute
  - [x] Verify Duration field maps from `duration` attribute (milliseconds)
  - [x] Verify ViewOffset field maps from `viewOffset` attribute (milliseconds)
  - [x] Apply fallbacks for any empty string fields

- [x] **Task 4: Add Metadata Extraction Tests** (AC: 1, 2, 3, 4, 5, 6, 7)
  - [x] Add test case for complete metadata extraction
  - [x] Add test case for missing track title (verify fallback)
  - [x] Add test case for missing artist (verify fallback)
  - [x] Add test case for missing album (verify fallback)
  - [x] Add test case for missing artwork (verify empty string)
  - [x] Add test case for missing duration/viewOffset (verify 0 default)
  - [x] Add test case for absolute artwork URL generation

- [x] **Task 5: Create NowPlaying Component** (AC: 1, 2, 3, 4, 5, 6)
  - [x] Create `frontend/src/components/NowPlaying.vue`
  - [x] Display track title, artist, album
  - [x] Display album artwork (or placeholder if missing)
  - [x] Display playback progress (elapsed/total duration)
  - [x] Display playback state indicator (playing/paused)
  - [x] Subscribe to `PlaybackUpdated` and `PlaybackStopped` events
  - [x] Show "Not playing" state when no active session

- [x] **Task 6: Add Playback Store** (AC: 1, 2, 3, 4, 5, 6)
  - [x] Create `frontend/src/stores/playback.js` Pinia store
  - [x] State: `currentTrack`, `isPlaying`, `isPaused`
  - [x] Actions: `setTrack`, `clearTrack`
  - [x] Initialize event listeners for Wails events
  - [x] Format duration/position for display (mm:ss)

- [x] **Task 7: Integrate NowPlaying into Dashboard** (AC: 1, 2, 3, 4)
  - [x] Import and use NowPlaying component in Dashboard view
  - [x] Position appropriately in dashboard layout
  - [x] Ensure real-time updates work correctly
  - [x] Test with actual Plex playback

## Dev Notes

### Architecture Patterns

- **Go Package Location**: `internal/plex/` per architecture document
- **Vue Components**: `frontend/src/components/NowPlaying.vue` per architecture
- **Pinia Stores**: `frontend/src/stores/playback.ts` per architecture
- **Event Naming**: Use `PascalCase` for Wails events (`PlaybackUpdated`)

### Plex API Metadata Fields

From Plex `/status/sessions` response:
```xml
<Track key="/library/metadata/12345"
       title="Song Name"                    <!-- Track title -->
       grandparentTitle="Artist Name"       <!-- Artist -->
       parentTitle="Album Name"             <!-- Album -->
       thumb="/library/metadata/12345/thumb/1234567890"  <!-- Relative URL -->
       duration="180000"                    <!-- Duration in milliseconds -->
       viewOffset="45000">                  <!-- Position in milliseconds -->
```

### Artwork URL Construction

The `thumb` attribute from Plex is a relative URL. To display artwork:
1. Build absolute URL: `{serverURL}{thumb}?X-Plex-Token={token}`
2. Example: `http://192.168.1.100:32400/library/metadata/12345/thumb/1234?X-Plex-Token=abc123`

**Security Note**: The token is required to access artwork. Since this is displayed only locally in the app, this is acceptable per architecture security guidelines.

### Frontend Duration Formatting

Convert milliseconds to display format:
```typescript
function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
}
```

### Dependencies from Previous Stories

- **Story 2.8**: Provides `MusicSession` type, `GetMusicSessions()`, `Poller`, and Wails events
- **Story 1.1-1.5**: Provides project structure, config, error codes

### Existing Implementation (From Story 2.8)

The following are already implemented:
- `MusicSession` struct with `Track`, `Artist`, `Album`, `Thumb`, `Duration`, `ViewOffset`
- `GetMusicSessions()` method that extracts from Plex API
- `PlaybackUpdated` and `PlaybackStopped` Wails events
- `sessionChanged()` helper that detects metadata changes

### Testing Standards

- Unit tests for fallback value application
- Unit tests for artwork URL generation
- Frontend component tests (optional, Vue Test Utils)
- Manual testing with actual Plex playback

### References

- [Source: _bmad-output/planning-artifacts/architecture.md#Frontend Architecture] - Vue component patterns
- [Source: _bmad-output/planning-artifacts/architecture.md#Communication Patterns] - Wails events
- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.9] - Acceptance criteria
- [Source: 2-8-music-session-detection.md] - Previous story implementation
- [Source: internal/plex/types.go] - Existing MusicSession type
- [Source: internal/plex/client.go] - Existing GetMusicSessions method

## Dev Agent Record

### Agent Model Used
Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References
- All 95 tests pass in internal/plex package
- Frontend builds successfully with no errors

### Completion Notes List
- **Task 1**: Added FallbackTrackTitle, FallbackArtist, FallbackAlbum constants and ApplyFallbacks() method to MusicSession struct in types.go
- **Task 2**: Added ThumbURL field to MusicSession and buildArtworkURL() helper method in client.go. URL escapes token for security.
- **Task 3**: GetMusicSessions() now applies fallbacks after creating each session and builds absolute artwork URLs
- **Task 4**: Added 15 new tests covering fallback behavior, artwork URL generation, and complete metadata extraction
- **Task 5**: Created NowPlaying.vue with album artwork, track info, progress bar, and playback state indicator
- **Task 6**: Created playback.js Pinia store with event listeners for PlaybackUpdated/PlaybackStopped, formatDuration helper
- **Task 7**: Integrated NowPlaying component into Dashboard.vue at prominent position

### File List

**Created:**
- `frontend/src/components/NowPlaying.vue` - Track display component with artwork, metadata, progress bar
- `frontend/src/stores/playback.js` - Pinia store for playback state management and Wails event handling

**Modified:**
- `internal/plex/types.go` - Added fallback constants, ThumbURL field, ApplyFallbacks() method
- `internal/plex/client.go` - Added buildArtworkURL(), apply fallbacks in GetMusicSessions()
- `internal/plex/client_test.go` - Added 15 new tests for metadata extraction and fallbacks
- `frontend/src/views/pages/Dashboard.vue` - Integrated NowPlaying component
- `frontend/src/types/events.js` - Updated MusicSession typedef with thumbUrl field

### Code Review Record

**Reviewed by:** claude-opus-4-5-20251101
**Review Date:** 2026-01-20

**Issues Found & Fixed:**
1. [MEDIUM] Console.log statements in playback.js - Fixed: Removed 4 debug console.log statements

**Accepted as-is:**
- Redundant frontend fallbacks (defensive coding, acceptable)
- No frontend unit tests (marked as optional in story)

**Verification:**
- All 16 metadata extraction tests pass
- Frontend builds successfully
- All 7 ACs verified as implemented
