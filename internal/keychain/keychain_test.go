package keychain

import (
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestSetAndGetToken tests the basic SetToken and GetToken flow
func TestSetAndGetToken(t *testing.T) {
	testToken := "test-plex-token-12345"

	// Clean up before test
	_ = DeleteToken()

	// Set the token
	err := SetToken(testToken)
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Get the token
	retrievedToken, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	// Verify token matches
	if retrievedToken != testToken {
		t.Errorf("Expected token %q, got %q", testToken, retrievedToken)
	}

	// Clean up after test
	_ = DeleteToken()
}

// TestSetEmptyToken tests that setting an empty token returns an error
func TestSetEmptyToken(t *testing.T) {
	err := SetToken("")
	if err == nil {
		t.Error("Expected error when setting empty token, got nil")
	}
}

// TestGetTokenWhenNotSet tests that GetToken returns empty string when no token is set
func TestGetTokenWhenNotSet(t *testing.T) {
	// Clean up to ensure no token exists
	_ = DeleteToken()

	token, err := GetToken()
	if err != nil {
		t.Errorf("GetToken should not error when token not set, got: %v", err)
	}

	if token != "" {
		t.Errorf("Expected empty token, got %q", token)
	}
}

// TestDeleteToken tests token deletion
func TestDeleteToken(t *testing.T) {
	testToken := "test-token-to-delete"

	// Set a token first
	err := SetToken(testToken)
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	// Delete the token
	err = DeleteToken()
	if err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	// Verify token is gone
	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "" {
		t.Errorf("Expected empty token after deletion, got %q", token)
	}
}

// TestDeleteTokenWhenNotSet tests that deleting a non-existent token doesn't error
func TestDeleteTokenWhenNotSet(t *testing.T) {
	// Ensure no token exists
	_ = DeleteToken()

	// Try to delete again
	err := DeleteToken()
	if err != nil {
		t.Errorf("DeleteToken should not error when token doesn't exist, got: %v", err)
	}
}

// TestMultipleSetCalls tests that setting a token multiple times updates it
func TestMultipleSetCalls(t *testing.T) {
	// Clean up before test
	_ = DeleteToken()

	// Set first token
	token1 := "first-token"
	err := SetToken(token1)
	if err != nil {
		t.Fatalf("First SetToken failed: %v", err)
	}

	// Set second token (should overwrite)
	token2 := "second-token"
	err = SetToken(token2)
	if err != nil {
		t.Fatalf("Second SetToken failed: %v", err)
	}

	// Verify we get the second token
	retrievedToken, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if retrievedToken != token2 {
		t.Errorf("Expected token %q, got %q", token2, retrievedToken)
	}

	// Clean up after test
	_ = DeleteToken()
}

// TestTokenWithSpecialCharacters tests that tokens with special characters are handled correctly
func TestTokenWithSpecialCharacters(t *testing.T) {
	// Clean up before test
	_ = DeleteToken()

	specialToken := "token-with-special-chars!@#$%^&*()=+[]{}|;:',.<>?/~`"

	err := SetToken(specialToken)
	if err != nil {
		t.Fatalf("SetToken with special chars failed: %v", err)
	}

	retrievedToken, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if retrievedToken != specialToken {
		t.Errorf("Special characters not preserved. Expected %q, got %q", specialToken, retrievedToken)
	}

	// Clean up after test
	_ = DeleteToken()
}

// TestLongToken tests that very long tokens are handled correctly
func TestLongToken(t *testing.T) {
	// Clean up before test
	_ = DeleteToken()

	// Create a very long token (1000 characters)
	longToken := ""
	for i := 0; i < 100; i++ {
		longToken += "0123456789"
	}

	err := SetToken(longToken)
	if err != nil {
		t.Fatalf("SetToken with long token failed: %v", err)
	}

	retrievedToken, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if retrievedToken != longToken {
		t.Errorf("Long token not preserved. Length expected %d, got %d", len(longToken), len(retrievedToken))
	}

	// Clean up after test
	_ = DeleteToken()
}

// TestIsKeychainUnavailable tests the keychain availability check
func TestIsKeychainUnavailable(t *testing.T) {
	// Test with nil error
	if isKeychainUnavailable(nil) {
		t.Error("nil error should not be considered keychain unavailable")
	}

	// Test with normal keyring error
	normalErr := keyring.ErrNotFound
	if isKeychainUnavailable(normalErr) {
		t.Error("ErrNotFound should not be considered keychain unavailable")
	}
}

// BenchmarkSetToken benchmarks token storage performance
func BenchmarkSetToken(b *testing.B) {
	token := "benchmark-test-token-12345"
	_ = DeleteToken()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SetToken(token)
	}

	// Clean up
	_ = DeleteToken()
}

// BenchmarkGetToken benchmarks token retrieval performance
func BenchmarkGetToken(b *testing.B) {
	token := "benchmark-test-token-12345"
	_ = SetToken(token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetToken()
	}

	// Clean up
	_ = DeleteToken()
}

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Clean up any leftover test data
	_ = DeleteToken()

	os.Exit(code)
}
