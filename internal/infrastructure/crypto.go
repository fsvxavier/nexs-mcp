package infrastructure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// KeyDerivationIterations for PBKDF2
	KeyDerivationIterations = 100000
	// SaltSize for PBKDF2
	SaltSize = 32
	// KeySize for AES-256
	KeySize = 32
)

// TokenEncryptor handles encryption and decryption of OAuth tokens
type TokenEncryptor struct {
	key []byte
}

// NewTokenEncryptor creates a new token encryptor
// The encryption key is derived from machine ID and user's home directory
// This provides reasonable security without requiring user to manage passwords
func NewTokenEncryptor() (*TokenEncryptor, error) {
	// Get machine ID (fallback to hostname if not available)
	machineID, err := getMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to get machine identifier: %w", err)
	}

	// Get salt from home directory (create if doesn't exist)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	saltPath := homeDir + "/.nexs-mcp/.salt"
	salt, err := getOrCreateSalt(saltPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create salt: %w", err)
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(machineID), salt, KeyDerivationIterations, KeySize, sha256.New)

	return &TokenEncryptor{
		key: key,
	}, nil
}

// Encrypt encrypts data using AES-256-GCM
func (e *TokenEncryptor) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data using AES-256-GCM
func (e *TokenEncryptor) Decrypt(ciphertext string) ([]byte, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// getMachineID returns a machine identifier
func getMachineID() (string, error) {
	// Try /etc/machine-id (Linux)
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		return string(data), nil
	}

	// Try /var/lib/dbus/machine-id (Linux alternative)
	if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		return string(data), nil
	}

	// Fallback to hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	// Add home directory to make it more unique
	homeDir, _ := os.UserHomeDir()
	return hostname + ":" + homeDir, nil
}

// getOrCreateSalt retrieves or creates a salt file
func getOrCreateSalt(path string) ([]byte, error) {
	// Try to read existing salt
	if data, err := os.ReadFile(path); err == nil && len(data) == SaltSize {
		return data, nil
	}

	// Generate new salt
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Ensure directory exists
	dir := path[:len(path)-len("/.salt")]
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write salt to file
	if err := os.WriteFile(path, salt, 0600); err != nil {
		return nil, fmt.Errorf("failed to write salt: %w", err)
	}

	return salt, nil
}
