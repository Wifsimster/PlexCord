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

func setTokenFallback(token string) error {
	key := deriveMachineKey()

	encrypted, err := encryptAES([]byte(token), key)
	if err != nil {
		return errors.Wrap(err, errors.ENCRYPTION_FAILED, "failed to encrypt token")
	}

	encoded := base64.StdEncoding.EncodeToString(encrypted)

	credPath, err := getFallbackPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(credPath), 0700); err != nil {
		return errors.Wrap(err, errors.ENCRYPTION_FAILED, "failed to create credential directory")
	}

	err = os.WriteFile(credPath, []byte(encoded), 0600)
	if err != nil {
		return errors.Wrap(err, errors.ENCRYPTION_FAILED, "failed to write encrypted token")
	}

	return nil
}

func getTokenFallback() (string, error) {
	credPath, err := getFallbackPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		return "", nil
	}

	encoded, err := os.ReadFile(credPath) //nolint:gosec // credPath is derived from user config dir, not user input
	if err != nil {
		return "", errors.Wrap(err, errors.DECRYPTION_FAILED, "failed to read encrypted token")
	}

	encrypted, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return "", errors.Wrap(err, errors.DECRYPTION_FAILED, "failed to decode encrypted token")
	}

	key := deriveMachineKey()

	decrypted, err := decryptAES(encrypted, key)
	if err != nil {
		return "", errors.Wrap(err, errors.DECRYPTION_FAILED, "failed to decrypt token")
	}

	return string(decrypted), nil
}

func deleteTokenFallback() error {
	credPath, err := getFallbackPath()
	if err != nil {
		return err
	}

	err = os.Remove(credPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, errors.KEYCHAIN_READ_FAILED, "failed to delete encrypted token")
	}

	return nil
}

func getFallbackPath() (string, error) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return "", err
	}

	credPath := filepath.Join(filepath.Dir(configPath), fallbackFilename)
	return credPath, nil
}

func deriveMachineKey() []byte {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if username == "" {
		username = "unknown"
	}

	data := hostname + ":" + username + ":plexcord-v1-salt"

	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

func encryptAES(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func decryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New(errors.DECRYPTION_FAILED, "ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
