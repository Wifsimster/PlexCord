# Story 4.10: Quit from Tray Menu

Status: done

## Story

As a user,
I want to quit PlexCord from the tray menu,
So that I can fully close the application when needed.

## Acceptance Criteria

1. **AC1: Quit Action**
   - **Given** the tray context menu is open
   - **When** the user clicks "Quit"
   - **Then** Discord presence is cleared

2. **AC2: Graceful Shutdown**
   - **Given** quit is triggered
   - **When** shutdown completes
   - **Then** all connections are closed gracefully

3. **AC3: Complete Exit**
   - **Given** quit is triggered
   - **When** the application exits
   - **Then** the application exits completely

4. **AC4: No Background Processes**
   - **Given** the application has exited
   - **When** checking system processes
   - **Then** no background processes remain running

## Tasks / Subtasks

- [x] **Task 1: QuitApp Method** (AC: 1, 2, 3, 4)
  - [x] Implement QuitApp() in app.go
  - [x] Call runtime.Quit() to terminate

- [x] **Task 2: Shutdown Hook** (AC: 1, 2)
  - [x] shutdown() callback stops session polling
  - [x] shutdown() disconnects Discord (clears presence)
  - [x] Log shutdown completion

## Dev Notes

### Implementation

The `QuitApp()` method triggers graceful shutdown via Wails:

```go
// app.go

// QuitApp terminates the application completely.
// This is called from the tray menu or when the user explicitly quits.
func (a *App) QuitApp() {
    log.Printf("Quit requested")
    runtime.Quit(a.ctx)
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
    // Stop session polling if running
    a.StopSessionPolling()

    // Disconnect Discord
    if a.discord != nil {
        a.discord.Disconnect()
    }

    log.Printf("Application shutdown complete")
}
```

### Shutdown Flow

1. User triggers quit (menu, QuitApp(), keyboard shortcut)
2. `runtime.Quit()` initiates Wails shutdown
3. `beforeClose()` callback fires (returns false to allow)
4. `shutdown()` callback fires:
   - StopSessionPolling() - stops Plex poller goroutine
   - discord.Disconnect() - clears presence and closes IPC
5. Application process terminates
6. No orphan goroutines or connections remain

### Discord Presence Clear

The `discord.Disconnect()` method (from Story 3.7) clears presence before closing:

```go
// internal/discord/presence.go
func (pm *PresenceManager) Disconnect() error {
    // ... clears presence via client.Logout()
    // ... closes IPC connection
}
```

### References

- [Source: app.go:116-120] - QuitApp method
- [Source: app.go:79-89] - shutdown callback
- [Source: internal/discord/presence.go] - Disconnect clears presence

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **QuitApp Method**: Exposes quit functionality as Wails binding.

2. **Shutdown Callback**: Wails calls shutdown() for graceful cleanup.

3. **Plex Poller Stopped**: StopSessionPolling() cancels context and stops goroutine.

4. **Discord Disconnected**: Presence cleared before IPC connection closed.

5. **Clean Exit**: No background processes or orphan connections remain.

### File List

Files modified:
- `app.go` - QuitApp method and shutdown callback
