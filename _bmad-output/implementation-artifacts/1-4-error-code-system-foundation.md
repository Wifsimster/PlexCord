# Story 1.4: Error Code System Foundation

Status: done

## Story

As a developer,
I want a structured error code system,
So that errors can be consistently identified and handled across the application.

## Acceptance Criteria

1. **AC1: Comprehensive Error Code Set**
   - **Given** the error code system is implemented in `/internal/errors/`
   - **When** an error occurs in any module
   - **Then** errors include a code from the defined set:
     - `PLEX_UNREACHABLE` - Server not responding
     - `PLEX_AUTH_FAILED` - Invalid token
     - `DISCORD_NOT_RUNNING` - Discord client not detected
     - `DISCORD_CONN_FAILED` - RPC connection error
     - `CONFIG_READ_FAILED` - Cannot read config file
     - `CONFIG_WRITE_FAILED` - Cannot write config file
   - **And** additional codes are added as needed for future functionality

2. **AC2: Human-Readable Error Messages**
   - **Given** an AppError is created
   - **When** the error is returned or logged
   - **Then** it includes a human-readable message
   - **And** the message provides actionable context (what failed, why)
   - **And** the message does NOT contain sensitive data (tokens, credentials)

3. **AC3: Frontend Serialization**
   - **Given** an AppError needs to be sent to the frontend
   - **When** the error is serialized to JSON
   - **Then** both `code` and `message` fields are included
   - **And** JSON uses camelCase field names
   - **And** the error conforms to the AppError interface

4. **AC4: Error Safety (NFR10)**
   - **Given** an error message is constructed
   - **When** the message includes dynamic data
   - **Then** no sensitive data (tokens, passwords, credentials) appears in the message
   - **And** helper functions validate messages don't contain sensitive patterns
   - **And** all existing errors comply with this requirement

5. **AC5: Error System Documentation**
   - **Given** the error system is implemented
   - **When** developers need to use it
   - **Then** documentation exists explaining:
     - How to create errors
     - When to use each error code
     - How to add new error codes
     - Best practices for error messages
   - **And** code examples demonstrate proper usage

## Tasks / Subtasks

- [x] **Task 1: Add Missing Error Codes** (AC: 1)
  - [x] Review existing error codes in `errors.go`
  - [x] Add any missing error codes from architecture/epics
  - [x] Group error codes by domain (Plex, Discord, Config, General)
  - [x] Add comments documenting each error code's purpose
  - [x] Verify all codes follow naming convention (UPPERCASE_UNDERSCORE)

- [x] **Task 2: Add Helper Functions** (AC: 2)
  - [x] Create `Wrap(err error, code string, message string)` function for wrapping Go errors
  - [x] Create `Is(err error, code string)` function to check error codes
  - [x] Create `GetCode(err error)` function to extract error code
  - [x] Add examples in code comments
  - [x] Test helper functions compile and work correctly

- [x] **Task 3: Implement Sensitive Data Validation** (AC: 4)
  - [x] Create `ContainsSensitiveData(message string)` function
  - [x] Check for patterns: "token", "password", "secret", "key", "credential"
  - [x] Check for hex strings >20 chars (likely tokens)
  - [x] Add validation to `New()` function (log warning if sensitive data detected)
  - [x] Add unit tests for sensitive data detection

- [x] **Task 4: Add Error Code Documentation** (AC: 5)
  - [x] Create `codes.go` file with error code constants and documentation
  - [x] Move error code constants from `errors.go` to `codes.go`
  - [x] Add godoc comments for each error code explaining:
     - What triggers this error
     - How to handle it
     - Example scenarios
  - [x] Keep `AppError` and helper functions in `errors.go`

- [x] **Task 5: Create Usage Examples** (AC: 5)
  - [x] Add package-level godoc with usage examples
  - [x] Show how to create a basic error
  - [x] Show how to wrap a Go error
  - [x] Show how to check error codes
  - [x] Show proper message construction (without sensitive data)

- [x] **Task 6: Write Unit Tests** (AC: 1-4)
  - [x] Create `errors_test.go` file
  - [x] Test `New()` function creates errors correctly
  - [x] Test `Error()` method returns message
  - [x] Test `Wrap()` function preserves original error
  - [x] Test `Is()` function identifies error codes correctly
  - [x] Test `ContainsSensitiveData()` detects sensitive patterns
  - [x] Test JSON serialization produces correct format
  - [x] Run tests: `go test ./internal/errors`

- [x] **Task 7: Verify Existing Usage** (AC: 2, 4)
  - [x] Review config package usage of errors
  - [x] Verify error messages don't contain sensitive data
  - [x] Verify error codes are used correctly
  - [x] Update any non-compliant error messages

- [x] **Task 8: Test Error Flow to Frontend** (AC: 3)
  - [x] Create test that serializes AppError to JSON
  - [x] Verify JSON has correct structure: `{"code": "...", "message": "..."}`
  - [x] Verify camelCase field names
  - [x] Verify no extra fields leak into JSON
  - [x] Document JSON format for frontend developers

- [x] **Task 9: Add Error Mapping Documentation** (AC: 5)
  - [x] Create error code mapping table
  - [x] List each error code with:
     - User-friendly message template
     - Technical cause
     - Recommended user action
  - [x] Add to story file or separate docs

- [x] **Task 10: Compile and Integration Test** (AC: 1-5)
  - [x] Run `go build ./internal/errors` to verify compilation
  - [x] Run `go test ./internal/errors` to verify all tests pass
  - [x] Run full project build with `wails build`
  - [x] Verify no compilation errors
  - [x] Verify application still launches successfully

## Dev Notes

### Critical Architecture Compliance

**This story SOLIDIFIES the error system foundation established in Story 1.2.**

Per Architecture Document (architecture.md):

**Error Handling Strategy:**
- Structured error codes for frontend display
- Error codes map to user-friendly messages
- NO sensitive data in error messages (NFR10)
- Errors serializable to JSON for Wails bindings

### Current Implementation (From Story 1.2)

**Existing in `/internal/errors/errors.go`:**
```go
// Error code constants
const (
    PLEX_UNREACHABLE      = "PLEX_UNREACHABLE"
    PLEX_AUTH_FAILED      = "PLEX_AUTH_FAILED"
    DISCORD_NOT_RUNNING   = "DISCORD_NOT_RUNNING"
    DISCORD_CONN_FAILED   = "DISCORD_CONN_FAILED"
    CONFIG_READ_FAILED    = "CONFIG_READ_FAILED"
    CONFIG_WRITE_FAILED   = "CONFIG_WRITE_FAILED"
)

// AppError represents an application error with code and message
type AppError struct {
    Code    string `json:"code"`     // camelCase JSON tag ✓
    Message string `json:"message"`  // camelCase JSON tag ✓
}

// Error implements the error interface
func (e *AppError) Error() string {
    return e.Message
}

// New creates a new AppError
func New(code, message string) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
    }
}
```

**This story ENHANCES with:**
- Helper functions (Wrap, Is, GetCode)
- Sensitive data validation
- Better organization (codes.go vs errors.go)
- Comprehensive unit tests
- Documentation and examples

### Error Code Design Principles

**From Architecture:**

| Error Code | User Message | Technical Meaning |
|------------|--------------|-------------------|
| `PLEX_UNREACHABLE` | "Cannot reach Plex server" | Server not responding to HTTP requests |
| `PLEX_AUTH_FAILED` | "Plex authentication failed" | Invalid or expired X-Plex-Token |
| `DISCORD_NOT_RUNNING` | "Discord is not running" | Discord client process not detected |
| `DISCORD_CONN_FAILED` | "Cannot connect to Discord" | Discord RPC IPC connection failed |
| `CONFIG_READ_FAILED` | "Cannot read configuration" | Config file missing or malformed |
| `CONFIG_WRITE_FAILED` | "Cannot save configuration" | Config file write permission denied |

### Naming Conventions (CRITICAL)

**Error Codes:**
- Format: `UPPERCASE_UNDERSCORE`
- Examples: `PLEX_UNREACHABLE`, `CONFIG_READ_FAILED`
- NOT: `PlexUnreachable`, `plex-unreachable`, `PLEX-UNREACHABLE`

**JSON Fields:**
- Format: `camelCase`
- Example: `{"code": "...", "message": "..."}`
- NOT: `{"Code": "...", "Message": "..."}` (PascalCase)
- NOT: `{"error_code": "...", "error_message": "..."}` (snake_case)

### Helper Functions to Add

**Wrap() - For wrapping Go errors:**
```go
func Wrap(err error, code string, message string) *AppError {
    fullMessage := message
    if err != nil {
        fullMessage = message + ": " + err.Error()
    }
    return &AppError{
        Code:    code,
        Message: fullMessage,
    }
}

// Usage:
err := os.ReadFile(path)
if err != nil {
    return errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to read config file")
}
```

**Is() - For checking error codes:**
```go
func Is(err error, code string) bool {
    if appErr, ok := err.(*AppError); ok {
        return appErr.Code == code
    }
    return false
}

// Usage:
if errors.Is(err, errors.PLEX_UNREACHABLE) {
    // Handle unreachable server
}
```

**GetCode() - For extracting error code:**
```go
func GetCode(err error) string {
    if appErr, ok := err.(*AppError); ok {
        return appErr.Code
    }
    return ""
}
```

### Sensitive Data Detection

**Patterns to detect:**
- Strings containing: "token", "password", "secret", "key", "credential", "auth"
- Hex strings > 20 characters (likely API tokens)
- Base64 strings > 30 characters
- UUID-like patterns with actual values

**Safe to include:**
- Generic descriptions: "invalid token", "authentication failed"
- File paths: "/path/to/config.json"
- Server URLs: "http://plex.example.com"
- Error codes: "PLEX_AUTH_FAILED"

**NOT safe:**
- Actual tokens: "X-Plex-Token: abc123xyz..."
- Actual passwords: "password: mypass123"
- Full error details with tokens

```go
func ContainsSensitiveData(message string) bool {
    lowerMsg := strings.ToLower(message)

    // Check for sensitive keywords followed by values
    sensitivePatterns := []string{
        "token=", "password=", "secret=", "key=", "credential=",
    }

    for _, pattern := range sensitivePatterns {
        if strings.Contains(lowerMsg, pattern) {
            return true
        }
    }

    // Check for long hex strings (likely tokens)
    hexPattern := regexp.MustCompile(`[0-9a-fA-F]{20,}`)
    if hexPattern.MatchString(message) {
        return true
    }

    return false
}
```

### File Organization

**After this story:**
```
internal/errors/
├── codes.go        # Error code constants with documentation
├── errors.go       # AppError struct and helper functions
└── errors_test.go  # Unit tests
```

### Unit Test Coverage

**Required tests:**
1. `TestNew()` - Verify New() creates errors correctly
2. `TestError()` - Verify Error() returns message
3. `TestWrap()` - Verify Wrap() includes original error
4. `TestIs()` - Verify Is() checks codes correctly
5. `TestGetCode()` - Verify GetCode() extracts codes
6. `TestContainsSensitiveData()` - Verify sensitive data detection
7. `TestJSONSerialization()` - Verify JSON format
8. `TestJSONFieldNames()` - Verify camelCase JSON tags

### Previous Story Intelligence

**Story 1.2 established:**
- `/internal/errors/errors.go` with AppError struct
- 6 error code constants
- New() constructor
- Error() method for error interface

**Story 1.3 used the error system:**
- `errors.New(errors.CONFIG_READ_FAILED, "...")`
- `errors.New(errors.CONFIG_WRITE_FAILED, "...")`
- Proper error messages without sensitive data

**This story COMPLETES the foundation:**
- Helper functions for easier usage
- Sensitive data validation (NFR10)
- Comprehensive documentation
- Unit tests for reliability
- Better code organization

### NFR Considerations

- **NFR10 (No credentials in logs):** Sensitive data validation enforces this ✓
- **NFR28 (Binary <20MB):** Error system adds ~5KB ✓
- **NFR29 (Single file):** All error code compiles into binary ✓

### Testing Strategy

**Unit Tests:**
- Test each error code can be created
- Test error messages are formatted correctly
- Test JSON serialization produces correct output
- Test sensitive data detection catches common patterns
- Test helper functions work as expected

**Integration Tests:**
- Verify errors flow from backend to frontend correctly
- Verify Wails bindings serialize errors properly
- Verify frontend can parse error JSON

### Common Pitfalls to Avoid

| Pitfall | How to Avoid |
|---------|--------------|
| Including actual tokens in messages | Use ContainsSensitiveData() validation |
| Using snake_case JSON tags | Stick to camelCase per architecture |
| Not documenting error codes | Add godoc comments for each code |
| Inconsistent error code naming | Use UPPERCASE_UNDERSCORE format |
| Missing error tests | Write comprehensive unit tests |

### Integration with Future Stories

**Story 2.x (Plex Integration) will use:**
- `PLEX_UNREACHABLE` when server doesn't respond
- `PLEX_AUTH_FAILED` when token is invalid

**Story 3.x (Discord Integration) will use:**
- `DISCORD_NOT_RUNNING` when client isn't detected
- `DISCORD_CONN_FAILED` when RPC connection fails

**Story 5.x (Settings UI) will:**
- Display error messages to users
- Map error codes to user-friendly UI messages

### References

- [Source: architecture.md#Error Handling Strategy]
- [Source: architecture.md#Implementation Patterns & Consistency Rules]
- [Source: epics.md#Story 1.4]
- [Source: NFR10 from PRD]
- [Source: Story 1.2 - errors package created]
- [Source: Story 1.3 - errors package usage]

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### Debug Log References

No debug logs required - implementation straightforward with clear requirements from architecture.

### Completion Notes List

**Implementation Summary:**

Story 1.4 successfully enhanced the error system foundation established in Story 1.2 with comprehensive helper functions, sensitive data validation, and thorough documentation.

**Files Created:**
1. **internal/errors/codes.go** (NEW)
   - Moved all error code constants from errors.go to dedicated codes.go file
   - Added comprehensive godoc documentation for each error code
   - Organized codes by domain: Plex, Discord, Config, General
   - Each constant documents: what triggers it, recommended action, example scenarios
   - 7 error codes: PLEX_UNREACHABLE, PLEX_AUTH_FAILED, DISCORD_NOT_RUNNING, DISCORD_CONN_FAILED, CONFIG_READ_FAILED, CONFIG_WRITE_FAILED, UNKNOWN_ERROR
   - Package-level godoc with complete usage examples

2. **internal/errors/errors_test.go** (NEW)
   - 13 comprehensive unit tests covering all functionality
   - Tests for New(), Error(), Wrap(), Is(), GetCode()
   - 11 sub-tests for ContainsSensitiveData() edge cases
   - JSON serialization tests verifying camelCase field names
   - Error code validation tests
   - All tests passing (13/13, 0 failures)

**Files Modified:**
1. **internal/errors/errors.go** (MODIFIED)
   - Added Wrap() function for wrapping Go errors with AppError
   - Added Is() function for checking error codes (replaces type assertions)
   - Added GetCode() function for extracting error codes
   - Added ContainsSensitiveData() function implementing NFR10 compliance
   - Added isGenericMention() helper to distinguish actual sensitive data from generic mentions
   - Enhanced New() function with sensitive data validation warning
   - All functions include comprehensive godoc with examples

**Key Implementation Details:**

1. **Sensitive Data Detection (NFR10 Compliance):**
   - Detects patterns: "token=", "password=", "secret=", "key=", "credential="
   - Detects long hex strings (>20 chars) likely to be API tokens
   - Detects base64 strings (>30 chars) likely to be encoded credentials
   - Allows generic mentions: "invalid token", "missing password", "token expired"
   - Bidirectional pattern matching: checks both "phrase + pattern" and "pattern + phrase"
   - Logs warning when sensitive data detected in error messages

2. **Helper Functions:**
   - `Wrap(err, code, msg)` - Preserves original error context while adding structured code
   - `Is(err, code)` - Type-safe error code checking without manual type assertions
   - `GetCode(err)` - Safe extraction of error code from any error type
   - All functions handle nil errors and non-AppError types gracefully

3. **JSON Serialization:**
   - Verified camelCase field names: `{"code": "...", "message": "..."}`
   - NOT PascalCase (Code, Message) or snake_case (error_code, error_message)
   - JSON round-trip tested to ensure data integrity
   - Compatible with Wails frontend bindings

4. **Code Organization:**
   - codes.go: Error code constants with extensive documentation
   - errors.go: AppError type and helper functions
   - errors_test.go: Comprehensive test suite
   - Clean separation of concerns for maintainability

**Testing Results:**
- Unit tests: 13/13 passing, 0 failures
- Test coverage includes: basic functionality, edge cases, JSON serialization, error code validation
- Full project build successful: `wails build` completed in 10.328s
- Application launch verified: User confirmed "it works"

**Architecture Compliance:**
- ✅ Error codes use UPPERCASE_UNDERSCORE format
- ✅ JSON fields use camelCase per architecture requirements
- ✅ NFR10 enforced: No sensitive data in error messages
- ✅ Errors serializable to JSON for Wails bindings
- ✅ User-friendly messages with actionable context

**Integration with Existing Code:**
- config package (Story 1.3) already uses error system correctly
- No changes required to existing error usages
- Error messages verified safe (no sensitive data)

**Test Fixes Applied:**
- Fixed isGenericMention() to check bidirectional patterns ("token expired" vs "expired token")
- All ContainsSensitiveData() tests now passing

**Ready for Integration:**
- Error system ready for use in upcoming Plex integration (Epic 2)
- Error system ready for use in upcoming Discord integration (Epic 3)
- Settings UI (Epic 5) can display error codes and messages to users
- Foundation solid for error recovery strategies in Epic 6

### File List

Files created/modified:
- `internal/errors/codes.go` (NEW - 95 lines, error code constants with comprehensive godoc)
- `internal/errors/errors.go` (MODIFIED - 168 lines, added Wrap/Is/GetCode/ContainsSensitiveData helper functions)
- `internal/errors/errors_test.go` (NEW - 216 lines, 13 comprehensive unit tests)
