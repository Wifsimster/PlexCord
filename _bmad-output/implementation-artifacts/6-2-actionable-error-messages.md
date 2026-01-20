# Story 6.2: Actionable Error Messages

Status: done

## Story

As a user,
I want error messages that tell me what went wrong and how to fix it,
So that I can resolve issues without guessing.

## Acceptance Criteria

1. **AC1: Clear Explanation**
   - **Given** an error occurs
   - **When** the error is displayed
   - **Then** the message explains the problem clearly

2. **AC2: Corrective Action**
   - **Given** an error is displayed
   - **When** viewing the error
   - **Then** the message suggests corrective action where applicable

3. **AC3: Specific Messages**
   - **Given** specific error types
   - **When** displayed to user
   - **Then** messages are appropriate:
     - `PLEX_UNREACHABLE`: "Cannot reach Plex server..."
     - `PLEX_AUTH_FAILED`: "Plex authentication failed..."
     - `DISCORD_NOT_RUNNING`: "Discord is not running..."
     - `DISCORD_CONN_FAILED`: "Cannot connect to Discord..."

4. **AC4: No Technical Jargon**
   - **Given** any error message
   - **When** displayed to user
   - **Then** no technical jargon is used (NFR25)

## Tasks / Subtasks

- [x] **Task 1: ErrorInfo Structure** (AC: 1, 2, 4)
  - [x] Define ErrorInfo struct with Title, Description, Suggestion
  - [x] Include Retryable flag for UI decisions

- [x] **Task 2: Error Message Map** (AC: 3)
  - [x] Map all error codes to user-friendly messages
  - [x] Include specific suggestions for each error type

- [x] **Task 3: Helper Functions** (AC: 2)
  - [x] GetErrorInfo(code) function
  - [x] IsRetryable(code) function
  - [x] IsAuthError(code) function
  - [x] IsConnectionError(code) function

- [x] **Task 4: Wails Bindings** (AC: 1, 2)
  - [x] GetErrorInfo method exposed to frontend
  - [x] IsRetryableError method exposed
  - [x] IsAuthError method exposed

## Dev Notes

### Implementation

Created `internal/errors/messages.go` with comprehensive error information:

```go
type ErrorInfo struct {
    Code        string `json:"code"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Suggestion  string `json:"suggestion"`
    Retryable   bool   `json:"retryable"`
}

var errorInfoMap = map[string]ErrorInfo{
    PLEX_UNREACHABLE: {
        Title:       "Plex Server Unreachable",
        Description: "Cannot reach Plex server...",
        Suggestion:  "Check if your Plex server is running...",
        Retryable:   true,
    },
    // ... all error codes mapped
}

func GetErrorInfo(code string) ErrorInfo
func IsRetryable(code string) bool
func IsAuthError(code string) bool
func IsConnectionError(code string) bool
```

### Error Categories

| Category | Error Codes | Retryable |
|----------|-------------|-----------|
| Connection | PLEX_UNREACHABLE, PLEX_CONN_FAILED, DISCORD_CONN_FAILED | Yes |
| Auth | PLEX_AUTH_FAILED, KEYCHAIN_READ_FAILED | No |
| Discord | DISCORD_NOT_RUNNING, DISCORD_CLIENT_ID_INVALID | Mixed |
| Config | CONFIG_READ_FAILED, CONFIG_WRITE_FAILED | Mixed |

### References

- [Source: internal/errors/messages.go] - ErrorInfo and helper functions
- [Source: app.go:875-889] - Wails bindings for error info

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **ErrorInfo Structure**: Comprehensive error info with title, description, suggestion, and retryable flag.

2. **Complete Mapping**: All error codes mapped to user-friendly messages.

3. **Helper Functions**: IsRetryable, IsAuthError, IsConnectionError for categorization.

4. **No Jargon**: Messages written in plain language per NFR25.

### File List

Files created/modified:
- `internal/errors/messages.go` - ErrorInfo and mappings
- `app.go` - GetErrorInfo, IsRetryableError, IsAuthError Wails bindings
