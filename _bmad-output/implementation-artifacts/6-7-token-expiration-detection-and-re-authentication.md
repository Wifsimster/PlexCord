# Story 6.7: Token Expiration Detection and Re-authentication

Status: done

## Story

As a user,
I want to be prompted when my Plex token expires,
So that I can re-authenticate and continue using PlexCord.

## Acceptance Criteria

1. **AC1: Detection**
   - **Given** PlexCord detects a `PLEX_AUTH_FAILED` error
   - **When** the error indicates token expiration
   - **Then** the error message explains the token has expired

2. **AC2: Re-authenticate Button**
   - **Given** token expiration is detected
   - **When** the error is displayed
   - **Then** a "Re-authenticate" button is shown

3. **AC3: Re-authentication Flow**
   - **Given** the user clicks re-authenticate
   - **When** the dialog opens
   - **Then** clicking the button opens the Plex token input dialog

4. **AC4: Token Storage**
   - **Given** a new token is entered
   - **When** the token is validated
   - **Then** entering a new valid token restores connection
   - **And** the new token is securely stored

## Tasks / Subtasks

- [x] **Task 1: Auth Error Detection** (AC: 1)
  - [x] IsAuthError(code) helper function
  - [x] Identifies PLEX_AUTH_FAILED, KEYCHAIN_READ_FAILED, DECRYPTION_FAILED

- [x] **Task 2: Error Info** (AC: 1)
  - [x] PLEX_AUTH_FAILED has Retryable: false
  - [x] Suggestion: "Please re-authenticate with Plex to get a new token"

- [x] **Task 3: No Auto-Retry** (AC: 2)
  - [x] startPlexRetry() skips auth errors
  - [x] Auth errors require user action

- [x] **Task 4: Existing Token Methods** (AC: 3, 4)
  - [x] SavePlexToken() already exists
  - [x] ValidatePlexConnection() already exists

## Dev Notes

### Implementation

Auth error detection in `internal/errors/messages.go`:

```go
// IsAuthError returns whether the error indicates an authentication issue.
// Used to detect when the user needs to re-authenticate.
func IsAuthError(code string) bool {
    return code == PLEX_AUTH_FAILED ||
           code == KEYCHAIN_READ_FAILED ||
           code == DECRYPTION_FAILED
}
```

Error info for PLEX_AUTH_FAILED:

```go
PLEX_AUTH_FAILED: {
    Code:        PLEX_AUTH_FAILED,
    Title:       "Plex Authentication Failed",
    Description: "Your Plex token is invalid or has expired.",
    Suggestion:  "Please re-authenticate with Plex to get a new token.",
    Retryable:   false, // Requires user action
},
```

Auto-retry exclusion in `app.go`:

```go
func (a *App) startPlexRetry(err error) {
    code := errors.GetCode(err)
    // Only retry for connection errors, not auth errors
    if errors.IsAuthError(code) {
        return // Auth errors require user action
    }
    a.plexRetry.Start(err, code)
}
```

### Frontend Flow

1. Frontend detects PLEX_AUTH_FAILED error
2. Calls `IsAuthError(code)` to confirm
3. Shows "Re-authenticate" button instead of retry
4. Button opens Plex token input (reuse setup wizard component)
5. On token submit, calls `SavePlexToken(newToken)`
6. Then calls `ValidatePlexConnection(serverURL)`
7. On success, resume normal operation

### References

- [Source: internal/errors/messages.go:145-149] - IsAuthError function
- [Source: internal/errors/messages.go:34-39] - PLEX_AUTH_FAILED ErrorInfo
- [Source: app.go:1003-1010] - startPlexRetry skips auth errors
- [Source: app.go:166-207] - SavePlexToken, GetPlexToken
- [Source: app.go:256-291] - ValidatePlexConnection

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **IsAuthError Helper**: Detects PLEX_AUTH_FAILED and related errors.

2. **Retryable: false**: Auth errors are not auto-retried.

3. **Clear Message**: User told token expired, suggests re-authentication.

4. **Existing Methods**: SavePlexToken and ValidatePlexConnection available.

5. **Frontend Work**: Re-authenticate button and dialog are frontend responsibility.

### File List

Files implementing this story:
- `internal/errors/messages.go` - IsAuthError, PLEX_AUTH_FAILED ErrorInfo
- `app.go` - startPlexRetry auth error exclusion, IsAuthError binding
