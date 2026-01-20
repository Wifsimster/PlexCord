# Story 3.5: Rich Presence Update with Track Info

Status: review

## Story

As a user,
I want my Discord status to show what I'm listening to on Plex,
So that my friends can see my music activity.

## Acceptance Criteria

1. **AC1: Track Title Display**
   - **Given** PlexCord is connected to both Plex and Discord
   - **When** music is playing on Plex
   - **Then** Discord Rich Presence shows the track title
   - **And** the title is shown in the "Details" field

2. **AC2: Artist Display**
   - **Given** music is playing
   - **When** Discord presence is updated
   - **Then** Discord Rich Presence shows the artist name
   - **And** the artist is shown in the "State" field

3. **AC3: Album Display**
   - **Given** music is playing
   - **When** Discord presence is updated
   - **Then** Discord Rich Presence shows the album name
   - **And** the album is shown in the "State" field (with artist)

4. **AC4: Playback Progress**
   - **Given** music is playing
   - **When** Discord presence is updated
   - **Then** Discord Rich Presence shows elapsed time
   - **And** the time updates in real-time via timestamp

5. **AC5: Update Timing (NFR4)**
   - **Given** a track change occurs
   - **When** the poller detects the change
   - **Then** Discord presence updates within 2 seconds
   - **And** the polling interval is 2 seconds by default

## Tasks / Subtasks

- [x] **Task 1: Wire Playback Events to Discord** (AC: 1, 2, 3, 4, 5)
  - [x] Modify `handleSessionUpdates()` to update Discord presence
  - [x] Added `updateDiscordFromSession()` helper method
  - [x] Call `discord.UpdatePresenceFromPlayback()` on playback updates
  - [x] Silently skip update if Discord not connected

- [x] **Task 2: Presence Data Mapping** (AC: 1, 2, 3)
  - [x] Map Track → Details
  - [x] Map Artist + Album → State ("by Artist on Album")
  - [x] Set LargeImage to "plex-logo"
  - [x] Set SmallImage based on playback state

- [x] **Task 3: Elapsed Time Display** (AC: 4)
  - [x] Calculate start time from ViewOffset
  - [x] Set Timestamps.Start for elapsed display
  - [x] Discord shows elapsed time automatically

- [x] **Task 4: Presence Building** (AC: 1, 2, 3, 4)
  - [x] `buildActivity()` creates rich-go Activity
  - [x] `UpdatePresenceFromPlayback()` convenience method

- [x] **Task 5: Integration Testing** (AC: 5)
  - [x] Build verification passed
  - [x] Discord tests pass (24 tests)
  - [x] NFR4 met via 2-second default polling interval

## Dev Notes

### Implementation Approach

The `handleSessionUpdates()` function in `app.go` already processes playback changes from the poller. We need to add Discord presence updates:

```go
func (a *App) handleSessionUpdates(sessionCh <-chan *plex.MusicSession) {
    var lastSession *plex.MusicSession

    for session := range sessionCh {
        if session != nil {
            // Update Discord presence with track info
            a.updateDiscordFromSession(session)

            // Emit frontend event
            runtime.EventsEmit(a.ctx, "PlaybackUpdated", session)
            lastSession = session
        } else if lastSession != nil {
            // Clear Discord presence
            a.clearDiscordPresence()

            runtime.EventsEmit(a.ctx, "PlaybackStopped", nil)
            lastSession = nil
        }
    }
}
```

### Presence Format

Discord Rich Presence will display:
- **Details**: Track title (e.g., "Highway to Hell")
- **State**: "by Artist on Album" (e.g., "by AC/DC on Highway to Hell")
- **Large Image**: "plex-logo" (configured in Discord Developer Portal)
- **Large Text**: "Plex"
- **Small Image**: "play" (playing) or "pause" (paused)
- **Small Text**: "Playing" or "Paused"
- **Timestamps**: Start time for elapsed display

### NFR4 Compliance

Default polling interval is 2 seconds. When track changes:
1. Poller detects change (within 2 seconds)
2. `handleSessionUpdates` receives new session
3. Discord presence is updated immediately
4. Total latency: ≤2 seconds

### Dependencies

- Story 3-1: Discord RPC Connection (PresenceManager)
- Story 2-8: Music Session Detection (Poller)
- Story 2-9: Track Metadata Extraction (MusicSession)

### References

- [Source: app.go:393-413] - handleSessionUpdates function
- [Source: internal/discord/presence.go] - SetPresence, buildActivity
- [Source: internal/plex/types.go:113-122] - MusicSession struct

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

(To be filled during implementation)

### Completion Notes List

1. **Wired Playback to Discord**: Modified `handleSessionUpdates()` to automatically update Discord presence when playback changes:
   - Added `updateDiscordFromSession()` - Updates Discord with track info
   - Added `clearDiscordOnStop()` - Clears Discord when playback stops

2. **Presence Data Flow**:
   - Poller detects music session → `handleSessionUpdates()` receives update
   - `updateDiscordFromSession()` calls `discord.UpdatePresenceFromPlayback()`
   - `UpdatePresenceFromPlayback()` creates `PresenceData` and calls `SetPresence()`
   - `buildActivity()` converts to rich-go Activity with track info

3. **Presence Display Format**:
   - **Details**: Track title
   - **State**: "by Artist on Album" (or "by Artist" if no album)
   - **Large Image**: "plex-logo"
   - **Small Image**: "play" (playing) or "pause" (paused)
   - **Timestamps**: Start time for elapsed display

4. **Silent Skip When Disconnected**: If Discord is not connected, presence updates are silently skipped without errors. This allows the app to function normally whether Discord is connected or not.

5. **NFR4 Compliance**: Default polling interval is 2 seconds, ensuring Discord presence updates within 2 seconds of playback state changes.

### File List

Files modified:
- `app.go` - Added `updateDiscordFromSession()` and `clearDiscordOnStop()` methods, integrated into `handleSessionUpdates()`
