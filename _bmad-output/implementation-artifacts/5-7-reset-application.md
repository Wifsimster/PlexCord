# Story 5.7: Reset Application

Status: done

## Story

As a user,
I want to reset PlexCord to its initial state,
So that I can start fresh or troubleshoot issues.

## Acceptance Criteria

1. **AC1: Confirmation Dialog**
   - **Given** the user is in the settings view
   - **When** selecting the reset option
   - **Then** a confirmation dialog warns about data loss

2. **AC2: Clear Config**
   - **Given** reset is confirmed
   - **When** the reset executes
   - **Then** reset clears all configuration settings

3. **AC3: Remove Token**
   - **Given** reset is confirmed
   - **When** the reset executes
   - **Then** reset removes the Plex token from secure storage

4. **AC4: Remove Auto-Start**
   - **Given** reset is confirmed
   - **When** the reset executes
   - **Then** reset removes auto-start registration

5. **AC5: Setup Wizard**
   - **Given** reset is complete
   - **When** the app is next launched
   - **Then** the setup wizard is shown

6. **AC6: No Auto-Exit**
   - **Given** reset is complete
   - **When** the reset finishes
   - **Then** the application does not exit automatically (user decides)

## Tasks / Subtasks

- [x] **Task 1: Config Delete** (AC: 2)
  - [x] Add `config.Delete()` function
  - [x] Remove config.json file

- [x] **Task 2: ResetApplication Method** (AC: 2, 3, 4, 5, 6)
  - [x] Stop session polling
  - [x] Disconnect Discord
  - [x] Delete Plex token from keychain
  - [x] Disable auto-start
  - [x] Delete config file
  - [x] Reset in-memory config to defaults
  - [x] Do NOT exit application

## Dev Notes

### Implementation

**Config Delete Function:**
```go
// internal/config/config.go
func Delete() error {
    configPath, err := GetConfigPath()
    if err != nil {
        return err
    }
    if err := os.Remove(configPath); err != nil {
        if os.IsNotExist(err) {
            return nil // Idempotent
        }
        return errors.New(errors.CONFIG_WRITE_FAILED, "failed to delete config file")
    }
    return nil
}
```

**ResetApplication Method:**
```go
// app.go
func (a *App) ResetApplication() error {
    log.Printf("Resetting application to initial state...")

    // 1. Stop session polling
    a.StopSessionPolling()

    // 2. Disconnect from Discord (clears presence)
    a.discordMu.Lock()
    if a.discord.IsConnected() {
        a.discord.Disconnect()
    }
    a.discordMu.Unlock()

    // 3. Remove Plex token from secure storage
    keychain.DeleteToken()

    // 4. Remove auto-start registration
    a.autostart.Disable()

    // 5. Delete configuration file
    config.Delete()

    // 6. Reset in-memory config to defaults
    a.config = config.DefaultConfig()

    log.Printf("Application reset complete - setup wizard will show on next launch")
    return nil
}
```

### Reset Flow

1. Frontend shows confirmation dialog (AC1)
2. User confirms reset
3. Frontend calls `ResetApplication()`
4. Backend performs cleanup (AC2-4)
5. In-memory config reset to defaults
6. App stays running (AC6)
7. Frontend navigates to setup wizard or user restarts app
8. Next launch shows setup wizard (AC5) because `SetupCompleted` is false

### Error Handling

The method continues even if individual steps fail:
- Token deletion failure: Logged, continues
- Auto-start failure: Logged, continues
- Config deletion failure: Logged, creates fresh defaults

### References

- [Source: internal/config/config.go:123-142] - Delete function
- [Source: app.go:802-853] - ResetApplication method
- [Source: internal/keychain/keychain.go:91] - DeleteToken

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Config Delete**: New `config.Delete()` function removes config.json.

2. **ResetApplication**: Comprehensive reset method in app.go.

3. **Graceful Cleanup**: Stops polling, disconnects Discord, clears presence.

4. **Secure Deletion**: Token removed from OS keychain.

5. **Auto-Start Cleanup**: Removes startup registration.

6. **No Auto-Exit**: App stays running for user control.

### File List

Files created/modified:
- `internal/config/config.go` - Delete function
- `app.go` - ResetApplication method
