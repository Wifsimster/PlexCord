package keychain

import (
	"encoding/base64"
	"os"
	"testing"
)

// TestDeriveMachineKey tests that machine key derivation is consistent
func TestDeriveMachineKey(t *testing.T) {
	// Get key twice and verify they're identical
	key1 := deriveMachineKey()
	key2 := deriveMachineKey()

	if len(key1) != 32 {
		t.Errorf("Expected 32-byte key, got %d bytes", len(key1))
	}

	if string(key1) != string(key2) {
		t.Error("Machine key should be consistent across calls")
	}
}

// TestEncryptDecryptAES tests AES encryption/decryption round-trip
func TestEncryptDecryptAES(t *testing.T) {
	key := deriveMachineKey()
	plaintext := []byte("test-plex-token-12345")

	// Encrypt
	ciphertext, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Verify ciphertext is different from plaintext
	if string(ciphertext) == string(plaintext) {
		t.Error("Ciphertext should be different from plaintext")
	}

	// Decrypt
	decrypted, err := decryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify decrypted matches original
	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text doesn't match. Expected %q, got %q", plaintext, decrypted)
	}
}

// TestEncryptAESWithDifferentKeys tests that different keys produce different ciphertexts
func TestEncryptAESWithDifferentKeys(t *testing.T) {
	plaintext := []byte("test-token")
	key1 := []byte("12345678901234567890123456789012") // 32 bytes
	key2 := []byte("abcdefghijklmnopqrstuvwxyz123456") // 32 bytes

	// Encrypt with key1
	ciphertext1, err := encryptAES(plaintext, key1)
	if err != nil {
		t.Fatalf("Encryption with key1 failed: %v", err)
	}

	// Encrypt with key2
	ciphertext2, err := encryptAES(plaintext, key2)
	if err != nil {
		t.Fatalf("Encryption with key2 failed: %v", err)
	}

	// Ciphertexts should be different
	if string(ciphertext1) == string(ciphertext2) {
		t.Error("Different keys should produce different ciphertexts")
	}
}

// TestDecryptAESWithWrongKey tests that decryption fails with wrong key
func TestDecryptAESWithWrongKey(t *testing.T) {
	plaintext := []byte("test-token")
	correctKey := []byte("12345678901234567890123456789012")
	wrongKey := []byte("abcdefghijklmnopqrstuvwxyz123456")

	// Encrypt with correct key
	ciphertext, err := encryptAES(plaintext, correctKey)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Try to decrypt with wrong key
	_, err = decryptAES(ciphertext, wrongKey)
	if err == nil {
		t.Error("Decryption should fail with wrong key")
	}
}

// TestDecryptAESWithShortCiphertext tests that decryption fails with too-short ciphertext
func TestDecryptAESWithShortCiphertext(t *testing.T) {
	key := deriveMachineKey()
	shortCiphertext := []byte("short")

	_, err := decryptAES(shortCiphertext, key)
	if err == nil {
		t.Error("Decryption should fail with ciphertext that's too short")
	}
}

// TestEncryptAESNonceUniqueness tests that encryption produces different nonces
func TestEncryptAESNonceUniqueness(t *testing.T) {
	key := deriveMachineKey()
	plaintext := []byte("test-token")

	// Encrypt same plaintext twice
	ciphertext1, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	ciphertext2, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	// Ciphertexts should be different due to different nonces
	if string(ciphertext1) == string(ciphertext2) {
		t.Error("Encrypting same plaintext twice should produce different ciphertexts (different nonces)")
	}
}

// TestEncryptDecryptEmptyString tests encrypting/decrypting empty string
func TestEncryptDecryptEmptyString(t *testing.T) {
	key := deriveMachineKey()
	plaintext := []byte("")

	// Encrypt
	ciphertext, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption of empty string failed: %v", err)
	}

	// Decrypt
	decrypted, err := decryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify empty string is preserved
	if string(decrypted) != "" {
		t.Errorf("Expected empty string, got %q", decrypted)
	}
}

// TestEncryptDecryptLongString tests encrypting/decrypting very long strings
func TestEncryptDecryptLongString(t *testing.T) {
	key := deriveMachineKey()

	// Create a 10KB string
	longPlaintext := make([]byte, 10240)
	for i := range longPlaintext {
		longPlaintext[i] = byte('A' + (i % 26))
	}

	// Encrypt
	ciphertext, err := encryptAES(longPlaintext, key)
	if err != nil {
		t.Fatalf("Encryption of long string failed: %v", err)
	}

	// Decrypt
	decrypted, err := decryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify match
	if string(decrypted) != string(longPlaintext) {
		t.Error("Long string not preserved through encryption/decryption")
	}
}

// TestFallbackTokenFlow tests the complete fallback storage flow
func TestFallbackTokenFlow(t *testing.T) {
	testToken := "test-fallback-token-12345"

	// Clean up before test
	_ = deleteTokenFallback()

	// Set token using fallback
	err := setTokenFallback(testToken)
	if err != nil {
		t.Fatalf("setTokenFallback failed: %v", err)
	}

	// Get token using fallback
	retrievedToken, err := getTokenFallback()
	if err != nil {
		t.Fatalf("getTokenFallback failed: %v", err)
	}

	// Verify token matches
	if retrievedToken != testToken {
		t.Errorf("Expected token %q, got %q", testToken, retrievedToken)
	}

	// Delete token
	err = deleteTokenFallback()
	if err != nil {
		t.Fatalf("deleteTokenFallback failed: %v", err)
	}

	// Verify token is gone
	retrievedToken, err = getTokenFallback()
	if err != nil {
		t.Fatalf("getTokenFallback after delete failed: %v", err)
	}

	if retrievedToken != "" {
		t.Errorf("Expected empty token after deletion, got %q", retrievedToken)
	}
}

// TestFallbackGetNonExistentToken tests getting a token that doesn't exist
func TestFallbackGetNonExistentToken(t *testing.T) {
	// Clean up to ensure no token exists
	_ = deleteTokenFallback()

	token, err := getTokenFallback()
	if err != nil {
		t.Errorf("getTokenFallback should not error when file doesn't exist, got: %v", err)
	}

	if token != "" {
		t.Errorf("Expected empty token, got %q", token)
	}
}

// TestFallbackFileEncryption tests that the fallback file is actually encrypted
func TestFallbackFileEncryption(t *testing.T) {
	testToken := "my-secret-token"

	// Clean up before test
	_ = deleteTokenFallback()

	// Set token
	err := setTokenFallback(testToken)
	if err != nil {
		t.Fatalf("setTokenFallback failed: %v", err)
	}

	// Read the file directly
	credPath, err := getFallbackPath()
	if err != nil {
		t.Fatalf("getFallbackPath failed: %v", err)
	}

	fileContents, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("Failed to read fallback file: %v", err)
	}

	// Verify the file contents don't contain the plaintext token
	if string(fileContents) == testToken {
		t.Error("Token should be encrypted in file, not stored as plaintext")
	}

	// Verify it's base64 encoded (should decode without error)
	_, err = base64.StdEncoding.DecodeString(string(fileContents))
	if err != nil {
		t.Errorf("File contents should be valid base64: %v", err)
	}

	// Clean up
	_ = deleteTokenFallback()
}

// TestFallbackWithSpecialCharacters tests fallback with special characters
func TestFallbackWithSpecialCharacters(t *testing.T) {
	// Clean up before test
	_ = deleteTokenFallback()

	specialToken := "token!@#$%^&*()=+[]{}|;:',.<>?/~`"

	err := setTokenFallback(specialToken)
	if err != nil {
		t.Fatalf("setTokenFallback with special chars failed: %v", err)
	}

	retrievedToken, err := getTokenFallback()
	if err != nil {
		t.Fatalf("getTokenFallback failed: %v", err)
	}

	if retrievedToken != specialToken {
		t.Errorf("Special characters not preserved. Expected %q, got %q", specialToken, retrievedToken)
	}

	// Clean up
	_ = deleteTokenFallback()
}

// BenchmarkEncryptAES benchmarks AES encryption performance
func BenchmarkEncryptAES(b *testing.B) {
	key := deriveMachineKey()
	plaintext := []byte("benchmark-test-token-12345")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encryptAES(plaintext, key)
	}
}

// BenchmarkDecryptAES benchmarks AES decryption performance
func BenchmarkDecryptAES(b *testing.B) {
	key := deriveMachineKey()
	plaintext := []byte("benchmark-test-token-12345")
	ciphertext, _ := encryptAES(plaintext, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decryptAES(ciphertext, key)
	}
}

// BenchmarkDeriveMachineKey benchmarks machine key derivation
func BenchmarkDeriveMachineKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = deriveMachineKey()
	}
}
