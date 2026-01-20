# Story 3.7: Clear Presence on Playback Stop

Status: review

## Story

As a user,
I want my Discord status to be cleared when I stop listening,
So that my friends don't see outdated activity.

## Acceptance Criteria

1. **AC1: Presence Cleared on Stop**
   - **Given** music was playing and Discord presence was active
   - **When** playback stops on Plex
   - **Then** Discord Rich Presence is cleared
   - **And** no music activity is shown in Discord

2. **AC2: Graceful Clearing**
   - **Given** Discord connection is active
   - **When** presence is cleared
   - **Then** the clearing happens without errors
   - **And** the connection remains active for future updates

3. **AC3: Update Timing (NFR4)**
   - **Given** playback stops
   - **When** the poller detects no active session
   - **Then** presence is cleared within 2 seconds

## Tasks / Subtasks

- [x] **Task 1: Stop Detection** (AC: 1, 3)
  - [x] Poller detects when session ends (returns nil)
  - [x] `handleSessionUpdates()` calls `clearDiscordOnStop()`
  - [x] Presence cleared within polling interval

- [x] **Task 2: Clear Implementation** (AC: 2)
  - [x] `clearDiscordOnStop()` helper method in app.go
  - [x] Calls `discord.ClearPresence()` to remove presence
  - [x] Silent skip if Discord not connected

- [x] **Task 3: Connection Maintained** (AC: 2)
  - [x] `ClearPresence()` re-logins after logout to maintain connection
  - [x] Connection available for future presence updates

## Dev Notes

### Implementation (Completed in Story 3-5)

This story was implemented as part of Story 3-5's playback-to-Discord integration. When playback stops:

```go
// In handleSessionUpdates()
} else if lastSession != nil {
    // Music stopped - clear Discord presence and emit frontend event
    log.Printf("Playback stopped")

    // Clear Discord Rich Presence
    a.clearDiscordOnStop()

    runtime.EventsEmit(a.ctx, "PlaybackStopped", nil)
    lastSession = nil
}

// clearDiscordOnStop clears Discord Rich Presence when playback stops.
func (a *App) clearDiscordOnStop() {
    a.discordMu.Lock()
    defer a.discordMu.Unlock()

    if !a.discord.IsConnected() {
        return
    }

    err := a.discord.ClearPresence()
    if err != nil {
        log.Printf("Warning: Failed to clear Discord presence: %v", err)
    }
}
```

### Clear Presence Mechanism

The `ClearPresence()` method in `internal/discord/presence.go` handles clearing by:
1. Logging out of Discord
2. Re-logging in to maintain connection
3. Presence is cleared but connection remains active

```go
func (pm *PresenceManager) ClearPresence() error {
    // rich-go doesn't have a clear function, so we logout and re-login
    client.Logout()
    pm.presence = nil

    // Reconnect
    err := client.Login(pm.clientID)
    if err != nil {
        pm.connected = false
        return mapDiscordError(err)
    }

    log.Printf("Discord: Presence cleared")
    return nil
}
```

### References

- [Source: Story 3-5] - Playback integration
- [Source: app.go] - `clearDiscordOnStop()` method
- [Source: internal/discord/presence.go:143-166] - ClearPresence implementation

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Implemented in Story 3-5**: Stop detection and presence clearing were implemented as part of the playback integration.

2. **`clearDiscordOnStop()` Method**: Added to app.go, called when `handleSessionUpdates()` receives nil session after a previous session existed.

3. **Connection Maintained**: `ClearPresence()` uses logout/re-login pattern to clear presence while maintaining connection for future updates.

4. **Silent Skip**: If Discord is not connected, clearing is silently skipped.

5. **NFR4 Compliance**: Stop detected within 2-second polling interval.

### File List

No new files - implemented in Story 3-5:
- `app.go` - `clearDiscordOnStop()` method
- `internal/discord/presence.go` - `ClearPresence()` implementation
