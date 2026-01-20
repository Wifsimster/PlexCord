# Story 3.6: Presence State for Paused Playback

Status: review

## Story

As a user,
I want my Discord status to reflect when music is paused,
So that my status accurately shows my activity.

## Acceptance Criteria

1. **AC1: Paused State Display**
   - **Given** music was playing and Discord presence was active
   - **When** the user pauses playback on Plex
   - **Then** Discord Rich Presence updates to show "Paused" state
   - **And** the SmallImage changes to "pause"

2. **AC2: Track Info Remains**
   - **Given** playback is paused
   - **When** Discord presence is updated
   - **Then** the track information remains visible
   - **And** track title, artist, album are still displayed

3. **AC3: Elapsed Time Stops**
   - **Given** playback is paused
   - **When** Discord presence is updated
   - **Then** the elapsed time stops incrementing
   - **And** timestamps are not set for paused state

4. **AC4: Update Timing (NFR4)**
   - **Given** playback state changes to paused
   - **When** the poller detects the change
   - **Then** the update occurs within 2 seconds

## Tasks / Subtasks

- [x] **Task 1: Paused State Detection** (AC: 1, 4)
  - [x] Poller detects `state: "paused"` from Plex
  - [x] `handleSessionUpdates()` processes paused sessions
  - [x] Presence updated with paused state

- [x] **Task 2: Paused Visual Indicator** (AC: 1)
  - [x] `buildActivity()` sets SmallImage to "pause" for paused state
  - [x] SmallText set to "Paused"

- [x] **Task 3: Timestamps Handling** (AC: 3)
  - [x] Timestamps only set when `state == "playing"`
  - [x] Paused state doesn't show elapsed time incrementing

## Dev Notes

### Implementation (Completed in Story 3-5)

This story was implemented as part of Story 3-5's playback-to-Discord integration. The `buildActivity()` function in `internal/discord/presence.go` already handles paused state:

```go
// Add small image for playback state
if data.State == "paused" {
    activity.SmallImage = "pause"
    activity.SmallText = "Paused"
} else if data.State == "playing" {
    activity.SmallImage = "play"
    activity.SmallText = "Playing"
}

// Set timestamps for elapsed time display (only for playing)
if data.StartTime != nil && data.State == "playing" {
    activity.Timestamps = &client.Timestamps{
        Start: data.StartTime,
    }
}
```

### References

- [Source: Story 3-5] - Playback integration
- [Source: internal/discord/presence.go:193-206] - buildActivity paused handling

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Implemented in Story 3-5**: Paused state handling was already implemented in the `buildActivity()` function when Story 3-5's integration was completed.

2. **Paused State Visual**: SmallImage = "pause", SmallText = "Paused"

3. **Timestamps Not Set**: When state is "paused", timestamps are not set, so Discord doesn't show incrementing elapsed time.

4. **NFR4 Compliance**: Paused state changes detected within 2-second polling interval.

### File List

No new files - implemented in Story 3-5:
- `internal/discord/presence.go` - Paused state handling in `buildActivity()`
- `app.go` - `updateDiscordFromSession()` passes state to Discord
