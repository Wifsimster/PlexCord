# Story 4.7: Minimize to System Tray

Status: done

## Story

As a user,
I want to minimize PlexCord to the system tray,
So that it runs in the background without a visible window.

## Acceptance Criteria

1. **AC1: Window Hidden on Action**
   - **Given** the PlexCord window is open
   - **When** the user clicks the minimize or close button (based on settings)
   - **Then** the window is hidden

2. **AC2: Background Operation**
   - **Given** the window is hidden
   - **When** the application continues running
   - **Then** the application continues running in the system tray

3. **AC3: Functionality Continues**
   - **Given** the window is minimized to tray
   - **When** music plays on Plex
   - **Then** Plex monitoring and Discord presence continue working

4. **AC4: Respects Preference**
   - **Given** minimize-to-tray is configured
   - **When** the user performs the configured action
   - **Then** the behavior follows the user's minimize-to-tray preference

## Tasks / Subtasks

- [x] **Task 1: HideWindowOnClose** (AC: 1, 2)
  - [x] Enable `HideWindowOnClose: true` in Wails options
  - [x] Window hides on close button click

- [x] **Task 2: Window Management Methods** (AC: 1, 3)
  - [x] Add `HideWindow()` method to app.go
  - [x] Add `MinimizeWindow()` method to app.go

- [x] **Task 3: Preference Setting** (AC: 4)
  - [x] Add `GetMinimizeToTray()` method
  - [x] Add `SetMinimizeToTray()` method
  - [x] Persist setting to config

## Dev Notes

### Implementation

Window management methods in `app.go` provide full control over window visibility:

```go
// app.go - Window Management Methods (Story 4.5-4.10)

// HideWindow hides the main application window.
// The application continues running in the background.
func (a *App) HideWindow() {
    runtime.WindowHide(a.ctx)
}

// MinimizeWindow minimizes the main application window.
func (a *App) MinimizeWindow() {
    runtime.WindowMinimise(a.ctx)
}

// GetMinimizeToTray returns whether the app should minimize to tray.
func (a *App) GetMinimizeToTray() bool {
    return a.config.MinimizeToTray
}

// SetMinimizeToTray updates the minimize to tray setting.
func (a *App) SetMinimizeToTray(enabled bool) error {
    a.config.MinimizeToTray = enabled
    if err := config.Save(a.config); err != nil {
        log.Printf("ERROR: Failed to save minimize to tray setting: %v", err)
        return err
    }
    log.Printf("Minimize to tray set to: %v", enabled)
    return nil
}
```

### Behavior Flow

1. User clicks close button (X)
2. `HideWindowOnClose: true` triggers window hide
3. App continues running in background
4. Plex poller continues monitoring
5. Discord presence updates continue
6. User can restore window via ShowWindow()

### References

- [Source: main.go:37] - `HideWindowOnClose: true`
- [Source: app.go:106-108] - HideWindow method
- [Source: app.go:111-113] - MinimizeWindow method
- [Source: app.go:123-136] - MinimizeToTray setting methods

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **HideWindowOnClose**: Wails option hides window on close instead of terminating.

2. **HideWindow Method**: Explicit method for programmatic window hiding.

3. **MinimizeWindow Method**: Standard minimize to taskbar.

4. **Preference Methods**: Get/Set for MinimizeToTray setting with config persistence.

5. **Background Continuity**: Plex polling and Discord presence continue when window hidden.

### File List

Files modified:
- `main.go` - `HideWindowOnClose: true`
- `app.go` - HideWindow, MinimizeWindow, Get/SetMinimizeToTray methods
