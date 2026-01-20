package errors

import (
	"log"
	"regexp"
	"strings"
)

// AppError represents an application error with code and message
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// New creates a new AppError with the specified code and message.
// If the message contains sensitive data, a warning is logged.
//
// Example:
//
//	err := errors.New(errors.CONFIG_READ_FAILED, "failed to read config file")
func New(code, message string) *AppError {
	if ContainsSensitiveData(message) {
		log.Printf("WARNING: Error message may contain sensitive data: %s", code)
	}

	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing Go error with an AppError, combining the messages.
// This is useful for preserving the original error context while adding
// a structured error code.
//
// Example:
//
//	file, err := os.Open(path)
//	if err != nil {
//	    return errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to open config")
//	}
func Wrap(err error, code string, message string) *AppError {
	fullMessage := message
	if err != nil {
		fullMessage = message + ": " + err.Error()
	}

	if ContainsSensitiveData(fullMessage) {
		log.Printf("WARNING: Error message may contain sensitive data: %s", code)
	}

	return &AppError{
		Code:    code,
		Message: fullMessage,
	}
}

// Is checks if the given error has the specified error code.
// Returns false if the error is not an AppError.
//
// Example:
//
//	if errors.Is(err, errors.PLEX_UNREACHABLE) {
//	    // Handle unreachable server
//	}
func Is(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// GetCode extracts the error code from an error.
// Returns empty string if the error is not an AppError.
//
// Example:
//
//	code := errors.GetCode(err)
//	if code == errors.CONFIG_WRITE_FAILED {
//	    // Handle write failure
//	}
func GetCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ""
}

// ContainsSensitiveData checks if a message contains patterns that might
// indicate sensitive data (tokens, passwords, keys, etc.).
//
// This function helps enforce NFR10 (no credentials in logs) by detecting:
// - Keywords followed by values: "token=", "password=", "secret=", etc.
// - Long hex strings (>20 chars) that are likely API tokens
// - Base64-like strings (>30 chars)
//
// Example:
//
//	if errors.ContainsSensitiveData(msg) {
//	    // Sanitize or reject the message
//	}
func ContainsSensitiveData(message string) bool {
	lowerMsg := strings.ToLower(message)

	// Check for sensitive keywords followed by assignment or colon
	sensitivePatterns := []string{
		"token=", "token:", "token ",
		"password=", "password:", "password ",
		"secret=", "secret:", "secret ",
		"key=", "key:", "key ",
		"credential=", "credential:", "credential ",
		"x-plex-token", "api_key", "apikey",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerMsg, pattern) {
			// Check if it's just a generic mention vs actual data
			// Allow phrases like "invalid token" or "token required"
			if !isGenericMention(lowerMsg, pattern) {
				return true
			}
		}
	}

	// Check for long hex strings (likely tokens)
	hexPattern := regexp.MustCompile(`[0-9a-fA-F]{20,}`)
	if hexPattern.MatchString(message) {
		return true
	}

	// Check for long base64-like strings (likely encoded tokens)
	base64Pattern := regexp.MustCompile(`[A-Za-z0-9+/]{30,}={0,2}`)
	if base64Pattern.MatchString(message) {
		return true
	}

	return false
}

// isGenericMention checks if a sensitive keyword is used in a generic context
// (e.g., "invalid token", "missing password") rather than with actual data
func isGenericMention(lowerMsg, pattern string) bool {
	genericPhrases := []string{
		"invalid ", "missing ", "required ", "failed ", "error ",
		"no ", "empty ", "bad ", "incorrect ", "expired ",
	}

	patternWord := strings.TrimSpace(pattern)

	for _, phrase := range genericPhrases {
		// Check for "phrase + pattern" (e.g., "invalid token")
		if strings.Contains(lowerMsg, phrase+patternWord) {
			return true
		}
		// Check for "pattern + phrase" (e.g., "token expired")
		if strings.Contains(lowerMsg, patternWord+" "+strings.TrimSpace(phrase)) {
			return true
		}
	}

	return false
}

// SanitizeForLogging sanitizes a message by replacing sensitive data with masked versions.
// This enforces NFR10 (no credentials in logs) by masking:
// - Plex tokens and API keys
// - Passwords and secrets
// - Long hex strings (likely tokens)
// - Long base64 strings (likely encoded secrets)
//
// For values, it shows only the first 4 and last 4 characters with "..." in between.
// For example: "X-Plex-Token: abcd1234xyz9876" becomes "X-Plex-Token: abcd...9876"
//
// Example:
//
//	safeMsg := errors.SanitizeForLogging("Token: abc123xyz789")
//	log.Printf(safeMsg) // "Token: abc1...x789"
func SanitizeForLogging(message string) string {
	if message == "" {
		return message
	}

	// Don't sanitize messages that are clearly generic/safe
	if !ContainsSensitiveData(message) {
		return message
	}

	result := message

	// Pattern 1: Sensitive keyword with assignment (token=value, password:value, etc.)
	// Replace the value with masked version
	sensitivePatterns := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		// Match "token=value" or "token: value" or "token value"
		{regexp.MustCompile(`(?i)(token[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(password[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(secret[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(key[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(credential[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(x-plex-token[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(api_key[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(apikey[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
	}

	for _, sp := range sensitivePatterns {
		result = sp.pattern.ReplaceAllString(result, sp.replacement)
	}

	// Pattern 2: Long hex strings (likely tokens) - mask the middle
	hexPattern := regexp.MustCompile(`\b[0-9a-fA-F]{20,}\b`)
	result = hexPattern.ReplaceAllStringFunc(result, func(match string) string {
		return maskMiddle(match)
	})

	// Pattern 3: Long base64-like strings - mask the middle
	base64Pattern := regexp.MustCompile(`\b[A-Za-z0-9+/]{30,}={0,2}\b`)
	result = base64Pattern.ReplaceAllStringFunc(result, func(match string) string {
		return maskMiddle(match)
	})

	return result
}

// maskMiddle masks the middle portion of a string, showing only first 4 and last 4 characters.
// For strings shorter than 12 characters, fully redacts them.
//
// Example:
//
//	maskMiddle("abcd1234xyz9876") // "abcd...9876"
//	maskMiddle("short")            // "***REDACTED***"
func maskMiddle(s string) string {
	if len(s) < 12 {
		return "***REDACTED***"
	}

	// Show first 4 and last 4 characters
	return s[:4] + "..." + s[len(s)-4:]
}
