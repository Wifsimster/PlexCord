package keychain

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestDeriveMachineKey(t *testing.T) {
	key1 := deriveMachineKey()
	key2 := deriveMachineKey()

	if len(key1) != 32 {
		t.Errorf("Expected 32-byte key, got %d bytes", len(key1))
	}

	if string(key1) != string(key2) {
		t.Error("Machine key should be consistent across calls")
	}
}

func TestEncryptDecryptAES(t *testing.T) {
	key := deriveMachineKey()
	plaintext := []byte("test-plex-token-12345")

	ciphertext, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if string(ciphertext) == string(plaintext) {
		t.Error("Ciphertext should be different from plaintext")
	}

	decrypted, err := decryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text doesn't match. Expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptAESWithDifferentKeys(t *testing.T) {
	plaintext := []byte("test-token")
	key1 := []byte("12345678901234567890123456789012")
	key2 := []byte("abcdefghijklmnopqrstuvwxyz123456")

	ciphertext1, err := encryptAES(plaintext, key1)
	if err != nil {
		t.Fatalf("Encryption with key1 failed: %v", err)
	}

	ciphertext2, err := encryptAES(plaintext, key2)
	if err != nil {
		t.Fatalf("Encryption with key2 failed: %v", err)
	}

	if string(ciphertext1) == string(ciphertext2) {
		t.Error("Different keys should produce different ciphertexts")
	}
}

func TestDecryptAESWithWrongKey(t *testing.T) {
	plaintext := []byte("test-token")
	correctKey := []byte("12345678901234567890123456789012")
	wrongKey := []byte("abcdefghijklmnopqrstuvwxyz123456")

	ciphertext, err := encryptAES(plaintext, correctKey)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	_, err = decryptAES(ciphertext, wrongKey)
	if err == nil {
		t.Error("Decryption should fail with wrong key")
	}
}

func TestDecryptAESWithShortCiphertext(t *testing.T) {
	key := deriveMachineKey()
	shortCiphertext := []byte("short")

	_, err := decryptAES(shortCiphertext, key)
	if err == nil {
		t.Error("Decryption should fail with ciphertext that's too short")
	}
}

func TestEncryptAESNonceUniqueness(t *testing.T) {
	key := deriveMachineKey()
	plaintext := []byte("test-token")

	ciphertext1, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	ciphertext2, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	if string(ciphertext1) == string(ciphertext2) {
		t.Error("Encrypting same plaintext twice should produce different ciphertexts (different nonces)")
	}
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	key := deriveMachineKey()
	plaintext := []byte("")

	ciphertext, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption of empty string failed: %v", err)
	}

	decrypted, err := decryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != "" {
		t.Errorf("Expected empty string, got %q", decrypted)
	}
}

func TestEncryptDecryptLongString(t *testing.T) {
	key := deriveMachineKey()

	longPlaintext := make([]byte, 10240)
	for i := range longPlaintext {
		longPlaintext[i] = byte('A' + (i % 26))
	}

	ciphertext, err := encryptAES(longPlaintext, key)
	if err != nil {
		t.Fatalf("Encryption of long string failed: %v", err)
	}

	decrypted, err := decryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != string(longPlaintext) {
		t.Error("Long string not preserved through encryption/decryption")
	}
}

func TestFallbackTokenFlow(t *testing.T) {
	testToken := "test-fallback-token-12345"

	_ = deleteTokenFallback()

	err := setTokenFallback(testToken)
	if err != nil {
		t.Fatalf("setTokenFallback failed: %v", err)
	}

	retrievedToken, err := getTokenFallback()
	if err != nil {
		t.Fatalf("getTokenFallback failed: %v", err)
	}

	if retrievedToken != testToken {
		t.Errorf("Expected token %q, got %q", testToken, retrievedToken)
	}

	err = deleteTokenFallback()
	if err != nil {
		t.Fatalf("deleteTokenFallback failed: %v", err)
	}

	retrievedToken, err = getTokenFallback()
	if err != nil {
		t.Fatalf("getTokenFallback after delete failed: %v", err)
	}

	if retrievedToken != "" {
		t.Errorf("Expected empty token after deletion, got %q", retrievedToken)
	}
}

func TestFallbackGetNonExistentToken(t *testing.T) {
	_ = deleteTokenFallback()

	token, err := getTokenFallback()
	if err != nil {
		t.Errorf("getTokenFallback should not error when file doesn't exist, got: %v", err)
	}

	if token != "" {
		t.Errorf("Expected empty token, got %q", token)
	}
}

func TestFallbackFileEncryption(t *testing.T) {
	testToken := "my-secret-token"

	_ = deleteTokenFallback()

	err := setTokenFallback(testToken)
	if err != nil {
		t.Fatalf("setTokenFallback failed: %v", err)
	}

	credPath, err := getFallbackPath()
	if err != nil {
		t.Fatalf("getFallbackPath failed: %v", err)
	}

	fileContents, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("Failed to read fallback file: %v", err)
	}

	if string(fileContents) == testToken {
		t.Error("Token should be encrypted in file, not stored as plaintext")
	}

	_, err = base64.StdEncoding.DecodeString(string(fileContents))
	if err != nil {
		t.Errorf("File contents should be valid base64: %v", err)
	}

	_ = deleteTokenFallback()
}

func TestFallbackWithSpecialCharacters(t *testing.T) {
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

	_ = deleteTokenFallback()
}

func BenchmarkEncryptAES(b *testing.B) {
	key := deriveMachineKey()
	plaintext := []byte("benchmark-test-token-12345")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encryptAES(plaintext, key)
	}
}

func BenchmarkDecryptAES(b *testing.B) {
	key := deriveMachineKey()
	plaintext := []byte("benchmark-test-token-12345")
	ciphertext, _ := encryptAES(plaintext, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decryptAES(ciphertext, key)
	}
}

func BenchmarkDeriveMachineKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = deriveMachineKey()
	}
}
