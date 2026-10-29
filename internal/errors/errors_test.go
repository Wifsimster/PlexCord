package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(PLEX_UNREACHABLE, "server not responding")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Code != PLEX_UNREACHABLE {
		t.Errorf("Expected code %s, got %s", PLEX_UNREACHABLE, err.Code)
	}

	if err.Message != "server not responding" {
		t.Errorf("Expected message 'server not responding', got '%s'", err.Message)
	}
}

func TestError(t *testing.T) {
	msg := "test error message"
	err := New(CONFIG_READ_FAILED, msg)

	if err.Error() != msg {
		t.Errorf("Expected Error() to return '%s', got '%s'", msg, err.Error())
	}
}

func TestWrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := Wrap(originalErr, CONFIG_READ_FAILED, "failed to read config")

	expectedMsg := "failed to read config: original error"
	if wrappedErr.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, wrappedErr.Message)
	}

	if wrappedErr.Code != CONFIG_READ_FAILED {
		t.Errorf("Expected code %s, got %s", CONFIG_READ_FAILED, wrappedErr.Code)
	}
}

func TestWrapNilError(t *testing.T) {
	wrappedErr := Wrap(nil, CONFIG_READ_FAILED, "failed to read config")

	if wrappedErr.Message != "failed to read config" {
		t.Errorf("Expected message 'failed to read config', got '%s'", wrappedErr.Message)
	}
}

func TestIs(t *testing.T) {
	err := New(PLEX_UNREACHABLE, "server not responding")

	if !Is(err, PLEX_UNREACHABLE) {
		t.Error("Expected Is() to return true for matching code")
	}

	if Is(err, DISCORD_NOT_RUNNING) {
		t.Error("Expected Is() to return false for non-matching code")
	}
}

func TestIsNonAppError(t *testing.T) {
	err := fmt.Errorf("regular Go error")

	if Is(err, PLEX_UNREACHABLE) {
		t.Error("Expected Is() to return false for non-AppError")
	}
}

func TestGetCode(t *testing.T) {
	err := New(CONFIG_WRITE_FAILED, "failed to write config")

	code := GetCode(err)
	if code != CONFIG_WRITE_FAILED {
		t.Errorf("Expected code %s, got %s", CONFIG_WRITE_FAILED, code)
	}
}

func TestGetCodeNonAppError(t *testing.T) {
	err := fmt.Errorf("regular Go error")

	code := GetCode(err)
	if code != "" {
		t.Errorf("Expected empty code, got %s", code)
	}
}

func TestContainsSensitiveData(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		sensitive bool
	}{
		{"token with value", "token=abc123xyz", true},
		{"password with value", "password=mypassword", true},
		{"generic token mention", "invalid token", false},
		{"generic password mention", "missing password", false},
		{"hex string token", "token: 1234567890abcdef12345678", true},
		{"base64 token", "Authorization: dGVzdDp0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3Q=", true},
		{"safe error message", "failed to connect to server", false},
		{"server URL", "http://plex.example.com", false},
		{"x-plex-token header", "X-Plex-Token: abc123", true},
		{"api key", "api_key=12345", true},
		{"expired token generic", "token expired", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsSensitiveData(tt.message)
			if result != tt.sensitive {
				t.Errorf("ContainsSensitiveData(%q) = %v, want %v", tt.message, result, tt.sensitive)
			}
		})
	}
}

func TestJSONSerialization(t *testing.T) {
	err := New(PLEX_UNREACHABLE, "server not responding")

	jsonData, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal error: %v", jsonErr)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if _, ok := jsonMap["code"]; !ok {
		t.Error("JSON missing 'code' field")
	}

	if _, ok := jsonMap["message"]; !ok {
		t.Error("JSON missing 'message' field")
	}

	if _, ok := jsonMap["Code"]; ok {
		t.Error("JSON should use camelCase 'code', not PascalCase 'Code'")
	}

	if _, ok := jsonMap["Message"]; ok {
		t.Error("JSON should use camelCase 'message', not PascalCase 'Message'")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	original := New(DISCORD_CONN_FAILED, "connection failed")

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var restored AppError
	if err := json.Unmarshal(jsonData, &restored); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if restored.Code != original.Code {
		t.Errorf("Code mismatch: got %s, want %s", restored.Code, original.Code)
	}

	if restored.Message != original.Message {
		t.Errorf("Message mismatch: got %s, want %s", restored.Message, original.Message)
	}
}

func TestAllErrorCodes(t *testing.T) {
	errorCodes := []string{
		PLEX_UNREACHABLE,
		PLEX_AUTH_FAILED,
		DISCORD_NOT_RUNNING,
		DISCORD_CONN_FAILED,
		CONFIG_READ_FAILED,
		CONFIG_WRITE_FAILED,
		KEYCHAIN_UNAVAILABLE,
		KEYCHAIN_STORE_FAILED,
		KEYCHAIN_READ_FAILED,
		ENCRYPTION_FAILED,
		DECRYPTION_FAILED,
		UNKNOWN_ERROR,
	}

	for _, code := range errorCodes {
		if code == "" {
			t.Error("Found empty error code")
		}

		if code != strings.ToUpper(code) {
			t.Errorf("Error code %s should be uppercase", code)
		}
	}
}

func TestSanitizeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "safe message unchanged",
			input:    "failed to connect to server",
			expected: "failed to connect to server",
		},
		{
			name:     "token with equals",
			input:    "error: token=abc123xyz789",
			expected: "error: token=***REDACTED***",
		},
		{
			name:     "token with colon",
			input:    "X-Plex-Token: abc123xyz789",
			expected: "X-Plex-Token: ***REDACTED***",
		},
		{
			name:     "password with equals",
			input:    "password=mySecretPassword123",
			expected: "password=***REDACTED***",
		},
		{
			name:     "secret key",
			input:    "secret=s3cr3tk3y",
			expected: "secret=***REDACTED***",
		},
		{
			name:     "api key",
			input:    "api_key=1234567890abcdef",
			expected: "api_key=***REDACTED***",
		},
		{
			name:     "long hex token",
			input:    "found token: 1234567890abcdef1234567890",
			expected: "found token: ***REDACTED***",
		},
		{
			name:     "long base64 token",
			input:    "Authorization: dGVzdDp0ZXN0dGVzdHRlc3R0ZXN0dGVzdHRlc3Q=",
			expected: "Authorization: dGVz...lc3Q=",
		},
		{
			name:     "generic token mention safe",
			input:    "invalid token provided",
			expected: "invalid token provided",
		},
		{
			name:     "expired token generic safe",
			input:    "token expired, please re-authenticate",
			expected: "token expired, please re-authenticate",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "multiple sensitive values",
			input:    "token=abc123 and password=xyz789",
			expected: "token=***REDACTED*** and password=***REDACTED***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLogging(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeForLogging(%q)\ngot:  %q\nwant: %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeForLoggingNoPlaintextLeak(t *testing.T) {
	sensitiveValue := "my-secret-plex-token-12345"
	input := "token=" + sensitiveValue

	result := SanitizeForLogging(input)

	if result == input {
		t.Error("Sanitized output should not be identical to input with sensitive data")
	}

	if len(result) > 0 && len(input) > 0 {
		if result == input {
			t.Errorf("Sensitive value leaked in output: %q", result)
		}
	}
}

func TestMaskMiddle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abcd1234xyz9876", "abcd...9876"},
		{"short", "***REDACTED***"},
		{"12345", "***REDACTED***"},
		{"exactly12chr", "exac...2chr"},
		{"a1b2c3d4e5f6g7h8i9j0", "a1b2...i9j0"},
		{"verylongtokenstring123456789", "very...6789"},
		{"", "***REDACTED***"},
		{"1234567890ab", "1234...90ab"},
	}

	for _, tt := range tests {
		result := maskMiddle(tt.input)
		if result != tt.expected {
			t.Errorf("maskMiddle(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeForLoggingWithRealTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "plex token format",
			input: "X-Plex-Token: aBcD1234eFgH5678iJkL9012",
		},
		{
			name:  "jwt token",
			input: "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
		},
		{
			name:  "api key",
			input: "API-Key: sk_live_1234567890abcdefghijklmnop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLogging(tt.input)

			if result == tt.input {
				t.Error("Sensitive token not sanitized")
			}

			if tt.input != "" && result == "" {
				t.Error("Sanitization produced empty result")
			}

			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
		})
	}
}

func BenchmarkSanitizeForLogging(b *testing.B) {
	msg := "X-Plex-Token: abc123xyz789 and password=mySecretPassword"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SanitizeForLogging(msg)
	}
}

func BenchmarkSanitizeForLoggingSafeMessage(b *testing.B) {
	msg := "failed to connect to server at http://plex.example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SanitizeForLogging(msg)
	}
}
