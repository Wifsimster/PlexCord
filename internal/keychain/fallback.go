package keychain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"

	"plexcord/internal/config"
	"plexcord/internal/errors"
)

const fallbackFilename = ".credentials"

// setTokenFallback encrypts and stores the token when OS keychain is unavailable
func setTokenFallback(token string) error {
	// Derive encryption key from machine-specific data
	key := deriveMachineKey()

	// Encrypt the token
	encrypted, err := encryptAES([]byte(token), key)
	if err != nil {
		return errors.Wrap(err, errors.ENCRYPTION_FAILED, "failed to encrypt token")
	}

	// Encode to base64 for safe file storage
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	// Get fallback file path
	credPath, err := getFallbackPath()
	if err != nil {
		return err
	}

	// Write to file with restricted permissions (0600 = user read/write only)
	err = os.WriteFile(credPath, []byte(encoded), 0600)
	if err != nil {
		return errors.Wrap(err, errors.ENCRYPTION_FAILED, "failed to write encrypted token")
	}

	return nil
}

// getTokenFallback retrieves and decrypts the token from fallback storage
func getTokenFallback() (string, error) {
	credPath, err := getFallbackPath()
	if err != nil {
		return "", err
	}

	// Check if fallback file exists
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		// Token not set yet
		return "", nil
	}

	// Read encrypted token
	encoded, err := os.ReadFile(credPath)
	if err != nil {
		return "", errors.Wrap(err, errors.DECRYPTION_FAILED, "failed to read encrypted token")
	}

	// Decode from base64
	encrypted, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return "", errors.Wrap(err, errors.DECRYPTION_FAILED, "failed to decode encrypted token")
	}

	// Derive same encryption key
	key := deriveMachineKey()

	// Decrypt the token
	decrypted, err := decryptAES(encrypted, key)
	if err != nil {
		return "", errors.Wrap(err, errors.DECRYPTION_FAILED, "failed to decrypt token")
	}

	return string(decrypted), nil
}

// deleteTokenFallback removes the encrypted token file
func deleteTokenFallback() error {
	credPath, err := getFallbackPath()
	if err != nil {
		return err
	}

	// Remove file if it exists
	err = os.Remove(credPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, errors.KEYCHAIN_READ_FAILED, "failed to delete encrypted token")
	}

	return nil
}

// getFallbackPath returns the path to the fallback credentials file
func getFallbackPath() (string, error) {
	// Get config directory
	configPath, err := config.GetConfigPath()
	if err != nil {
		return "", err
	}

	// Fallback file is in same directory as config
	credPath := filepath.Join(filepath.Dir(configPath), fallbackFilename)
	return credPath, nil
}

// deriveMachineKey creates an encryption key from machine-specific data
// This provides basic protection for the fallback encryption.
// The key is derived from hostname + username to make it machine-specific.
func deriveMachineKey() []byte {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	// Get username
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // Windows
	}
	if username == "" {
		username = "unknown"
	}

	// Combine with salt
	data := hostname + ":" + username + ":plexcord-v1-salt"

	// Hash to create 32-byte key for AES-256
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// encryptAES encrypts plaintext using AES-256-GCM
func encryptAES(plaintext []byte, key []byte) ([]byte, error) {
	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM (Galois/Counter Mode) cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and append nonce to the beginning
	// Format: [nonce][encrypted data]
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAES decrypts ciphertext using AES-256-GCM
func decryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce from beginning
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New(errors.DECRYPTION_FAILED, "ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
