package keychain

import (
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestSetAndGetToken(t *testing.T) {
	testToken := "test-plex-token-12345"

	_ = DeleteToken()

	err := SetToken(testToken)
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	retrievedToken, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if retrievedToken != testToken {
		t.Errorf("Expected token %q, got %q", testToken, retrievedToken)
	}

	_ = DeleteToken()
}

func TestSetEmptyToken(t *testing.T) {
	err := SetToken("")
	if err == nil {
		t.Error("Expected error when setting empty token, got nil")
	}
}

func TestGetTokenWhenNotSet(t *testing.T) {
	_ = DeleteToken()

	token, err := GetToken()
	if err != nil {
		t.Errorf("GetToken should not error when token not set, got: %v", err)
	}

	if token != "" {
		t.Errorf("Expected empty token, got %q", token)
	}
}

func TestDeleteToken(t *testing.T) {
	testToken := "test-token-to-delete"

	err := SetToken(testToken)
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	err = DeleteToken()
	if err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "" {
		t.Errorf("Expected empty token after deletion, got %q", token)
	}
}

func TestDeleteTokenWhenNotSet(t *testing.T) {
	_ = DeleteToken()

	err := DeleteToken()
	if err != nil {
		t.Errorf("DeleteToken should not error when token doesn't exist, got: %v", err)
	}
}

func TestMultipleSetCalls(t *testing.T) {
	_ = DeleteToken()

	token1 := "first-token"
	err := SetToken(token1)
	if err != nil {
		t.Fatalf("First SetToken failed: %v", err)
	}

	token2 := "second-token"
	err = SetToken(token2)
	if err != nil {
		t.Fatalf("Second SetToken failed: %v", err)
	}

	retrievedToken, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if retrievedToken != token2 {
		t.Errorf("Expected token %q, got %q", token2, retrievedToken)
	}

	_ = DeleteToken()
}

func TestTokenWithSpecialCharacters(t *testing.T) {
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

	_ = DeleteToken()
}

func TestLongToken(t *testing.T) {
	_ = DeleteToken()

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

	_ = DeleteToken()
}

func BenchmarkSetToken(b *testing.B) {
	token := "benchmark-test-token-12345"
	_ = DeleteToken()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SetToken(token)
	}

	_ = DeleteToken()
}

func BenchmarkGetToken(b *testing.B) {
	token := "benchmark-test-token-12345"
	_ = SetToken(token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetToken()
	}

	_ = DeleteToken()
}

func TestMain(m *testing.M) {
	code := m.Run()

	_ = DeleteToken()

	os.Exit(code)
}

func TestGetTokenReadsFallbackWhenKeychainIsEmpty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	t.Setenv("APPDATA", dir)
	keyring.MockInit()

	if err := setTokenFallback("fallback-token"); err != nil {
		t.Fatalf("setTokenFallback: %v", err)
	}
	got, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if got != "fallback-token" {
		t.Errorf("GetToken() = %q, want the fallback token", got)
	}
}
