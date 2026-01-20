# Story 5.3: Auto-Start on Login

Status: done

## Story

As a user,
I want PlexCord to start automatically when I log in,
So that my Discord presence works without manual intervention.

## Acceptance Criteria

1. **AC1: Toggle Setting**
   - **Given** the user is in the settings view
   - **When** toggling the auto-start option
   - **Then** a toggle/checkbox enables or disables auto-start

2. **AC2: Enable Auto-Start**
   - **Given** auto-start is toggled on
   - **When** the setting is saved
   - **Then** PlexCord is added to system startup

3. **AC3: Disable Auto-Start**
   - **Given** auto-start is toggled off
   - **When** the setting is saved
   - **Then** PlexCord is removed from system startup

4. **AC4: State Accuracy**
   - **Given** auto-start setting is displayed
   - **When** viewing the toggle
   - **Then** the current state is accurately reflected

5. **AC5: Cross-Platform**
   - **Given** PlexCord runs on any supported platform
   - **When** auto-start is enabled/disabled
   - **Then** the setting works correctly on all three platforms

## Tasks / Subtasks

- [x] **Task 1: AutoStartManager** (AC: 2, 3, 5)
  - [x] Create `internal/platform/autostart.go` base
  - [x] Platform-specific implementations

- [x] **Task 2: Windows Implementation** (AC: 5)
  - [x] Registry-based auto-start (HKCU\Software\Microsoft\Windows\CurrentVersion\Run)
  - [x] `autostart_windows.go` with build tag

- [x] **Task 3: macOS Implementation** (AC: 5)
  - [x] LaunchAgent plist creation
  - [x] `autostart_darwin.go` with build tag

- [x] **Task 4: Linux Implementation** (AC: 5)
  - [x] XDG .desktop file in ~/.config/autostart/
  - [x] `autostart_linux.go` with build tag

- [x] **Task 5: Wails Bindings** (AC: 1, 4)
  - [x] `GetAutoStart()` method
  - [x] `SetAutoStart()` method

## Dev Notes

### Implementation

Platform-specific auto-start using native OS mechanisms:

**Windows** (`autostart_windows.go`):
- Uses `golang.org/x/sys/windows/registry`
- Adds/removes entry in `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
- Stores quoted executable path

**macOS** (`autostart_darwin.go`):
- Creates LaunchAgent plist at `~/Library/LaunchAgents/com.plexcord.app.plist`
- Sets `RunAtLoad: true` for login startup

**Linux** (`autostart_linux.go`):
- Creates XDG .desktop file at `~/.config/autostart/plexcord.desktop`
- Uses `X-GNOME-Autostart-enabled=true`

### Wails Bindings

```go
// GetAutoStart returns whether auto-start on login is enabled.
// This checks the actual OS registration, not just the config value.
func (a *App) GetAutoStart() bool {
    return a.autostart.IsEnabled()
}

// SetAutoStart enables or disables auto-start on login.
func (a *App) SetAutoStart(enabled bool) error {
    if err := a.autostart.SetEnabled(enabled); err != nil {
        return err
    }
    a.config.AutoStart = enabled
    return config.Save(a.config)
}
```

### References

- [Source: internal/platform/autostart.go] - Base AutoStartManager
- [Source: internal/platform/autostart_windows.go] - Windows registry
- [Source: internal/platform/autostart_darwin.go] - macOS LaunchAgent
- [Source: internal/platform/autostart_linux.go] - Linux XDG
- [Source: app.go:143-171] - GetAutoStart, SetAutoStart

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **AutoStartManager**: Cross-platform abstraction with platform-specific implementations.

2. **Windows**: Registry key in HKCU Run key.

3. **macOS**: LaunchAgent plist with RunAtLoad.

4. **Linux**: XDG autostart .desktop file.

5. **Wails Bindings**: GetAutoStart checks actual OS state, SetAutoStart updates both OS and config.

### File List

Files created/modified:
- `internal/platform/autostart.go` - Base manager
- `internal/platform/autostart_windows.go` - Windows implementation
- `internal/platform/autostart_darwin.go` - macOS implementation
- `internal/platform/autostart_linux.go` - Linux implementation
- `app.go` - GetAutoStart, SetAutoStart bindings
