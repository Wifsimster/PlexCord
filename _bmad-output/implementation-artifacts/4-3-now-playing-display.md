# Story 4.3: Now Playing Display

Status: done

## Story

As a user,
I want to see what track is currently playing,
So that I can verify PlexCord is detecting my music.

## Acceptance Criteria

1. **AC1: Track Title**
   - **Given** music is playing on Plex
   - **When** viewing the dashboard
   - **Then** the current track title is displayed

2. **AC2: Artist Name**
   - **Given** music is playing on Plex
   - **When** viewing the dashboard
   - **Then** the artist name is displayed

3. **AC3: Album Name**
   - **Given** music is playing on Plex
   - **When** viewing the dashboard
   - **Then** the album name is displayed

4. **AC4: Album Artwork**
   - **Given** music is playing on Plex
   - **When** viewing the dashboard
   - **Then** album artwork is displayed (if available)

5. **AC5: Playback State**
   - **Given** music is playing or paused
   - **When** viewing the dashboard
   - **Then** playback state is indicated (playing/paused)

6. **AC6: Real-time Updates**
   - **Given** tracks change during playback
   - **When** viewing the dashboard
   - **Then** the display updates in real-time

7. **AC7: Not Playing State**
   - **Given** no music is playing
   - **When** viewing the dashboard
   - **Then** a "Not playing" state is shown

## Tasks / Subtasks

- [x] **Task 1: Playback Store** (AC: 6)
  - [x] Create Pinia store for playback state
  - [x] Subscribe to PlaybackUpdated/PlaybackStopped events
  - [x] Format duration and progress

- [x] **Task 2: NowPlaying Component** (AC: 1, 2, 3, 4, 5, 7)
  - [x] Track, artist, album display
  - [x] Album artwork with fallback
  - [x] Progress bar with position/duration
  - [x] Playback state indicator (icon + text)
  - [x] "No music playing" empty state

## Dev Notes

### Implementation

Playback store in `frontend/src/stores/playback.js`:
- Subscribes to PlaybackUpdated and PlaybackStopped events
- Stores current track info (title, artist, album, artwork, duration, position)
- Provides formatted getters (formattedPosition, formattedDuration, progressPercent)

NowPlaying component in `frontend/src/components/NowPlaying.vue`:
- 80x80px album artwork with placeholder fallback
- Track title (font-semibold text-lg)
- Artist name (text-muted-color)
- Album name (text-sm text-muted-color)
- Progress bar with time display
- Playback state icon (play/pause/stop)
- Player name display

### Empty State

When no music is playing:
- Volume-off icon
- "No music playing" text
- Instruction text

### References

- [Source: frontend/src/stores/playback.js] - Playback store
- [Source: frontend/src/components/NowPlaying.vue] - Now playing component

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Event-driven**: Updates via Wails events for real-time sync.

2. **Album Artwork**: Uses thumbUrl from MusicSession, falls back to placeholder SVG.

3. **Progress Bar**: Shows position/duration with animated progress.

4. **State Icons**: Green play, yellow pause icons indicate state.

5. **Empty State**: Clear messaging when no music is playing.

### File List

Files implementing this story:
- `frontend/src/stores/playback.js` - Playback store
- `frontend/src/components/NowPlaying.vue` - Now playing component
