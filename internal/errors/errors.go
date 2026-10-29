package errors

import (
	"log"
	"regexp"
	"strings"
)

var (
	hexPattern    = regexp.MustCompile(`[0-9a-fA-F]{20,}`)
	base64Pattern = regexp.MustCompile(`[A-Za-z0-9+/]{30,}={0,2}`)

	sanitizePatterns = []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`(?i)(token[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(password[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(secret[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(key[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(credential[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(x-plex-token[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(api_key[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
		{regexp.MustCompile(`(?i)(apikey[=:\s]+)([^\s,;]+)`), "${1}***REDACTED***"},
	}

	sanitizeHexPattern    = regexp.MustCompile(`\b[0-9a-fA-F]{20,}\b`)
	sanitizeBase64Pattern = regexp.MustCompile(`\b[A-Za-z0-9+/]{30,}={0,2}\b`)
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string) *AppError {
	if ContainsSensitiveData(message) {
		log.Printf("WARNING: Error message may contain sensitive data: %s", code)
	}

	return &AppError{
		Code:    code,
		Message: message,
	}
}

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

func Is(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

func GetCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ""
}

func ContainsSensitiveData(message string) bool {
	lowerMsg := strings.ToLower(message)

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
			if !isGenericMention(lowerMsg, pattern) {
				return true
			}
		}
	}

	if hexPattern.MatchString(message) {
		return true
	}

	return base64Pattern.MatchString(message)
}

func isGenericMention(lowerMsg, pattern string) bool {
	genericPhrases := []string{
		"invalid ", "missing ", "required ", "failed ", "error ",
		"no ", "empty ", "bad ", "incorrect ", "expired ",
	}

	patternWord := strings.TrimSpace(pattern)

	for _, phrase := range genericPhrases {
		if strings.Contains(lowerMsg, phrase+patternWord) {
			return true
		}
		if strings.Contains(lowerMsg, patternWord+" "+strings.TrimSpace(phrase)) {
			return true
		}
	}

	return false
}

func SanitizeForLogging(message string) string {
	if message == "" {
		return message
	}

	if !ContainsSensitiveData(message) {
		return message
	}

	result := message

	for _, sp := range sanitizePatterns {
		result = sp.pattern.ReplaceAllString(result, sp.replacement)
	}

	result = sanitizeHexPattern.ReplaceAllStringFunc(result, maskMiddle)

	result = sanitizeBase64Pattern.ReplaceAllStringFunc(result, maskMiddle)

	return result
}

func maskMiddle(s string) string {
	if len(s) < 12 {
		return "***REDACTED***"
	}

	return s[:4] + "..." + s[len(s)-4:]
}
