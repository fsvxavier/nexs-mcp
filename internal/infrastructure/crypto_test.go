package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenEncryptor_EncryptDecrypt(t *testing.T) {
	encryptor, err := NewTokenEncryptor()
	require.NoError(t, err)
	require.NotNil(t, encryptor)

	testData := []byte("sensitive OAuth token data")

	// Encrypt
	encrypted, err := encryptor.Encrypt(testData)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, string(testData), encrypted)

	// Decrypt
	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, testData, decrypted)
}

func TestTokenEncryptor_EncryptDecryptDifferentData(t *testing.T) {
	encryptor, err := NewTokenEncryptor()
	require.NoError(t, err)

	testCases := []string{
		"",
		"a",
		"short string",
		"This is a longer string with special characters: !@#$%^&*()",
		`{"access_token":"gho_test123","token_type":"bearer","expires_in":3600}`,
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			encrypted, err := encryptor.Encrypt([]byte(testCase))
			require.NoError(t, err)

			decrypted, err := encryptor.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, testCase, string(decrypted))
		})
	}
}

func TestTokenEncryptor_DifferentEncryptionsAreDifferent(t *testing.T) {
	encryptor, err := NewTokenEncryptor()
	require.NoError(t, err)

	data := []byte("same data")

	encrypted1, err := encryptor.Encrypt(data)
	require.NoError(t, err)

	encrypted2, err := encryptor.Encrypt(data)
	require.NoError(t, err)

	// Same data should produce different ciphertexts (due to random nonce)
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to the same plaintext
	decrypted1, err := encryptor.Decrypt(encrypted1)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted1)

	decrypted2, err := encryptor.Decrypt(encrypted2)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted2)
}

func TestTokenEncryptor_DecryptInvalidData(t *testing.T) {
	encryptor, err := NewTokenEncryptor()
	require.NoError(t, err)

	invalidTests := []struct {
		name string
		data string
	}{
		{"empty string", ""},
		{"invalid base64", "not-valid-base64!@#$"},
		{"valid base64 but too short", "YWJj"},
		{"valid base64 but invalid ciphertext", "dGhpcyBpcyBub3QgdmFsaWQgY2lwaGVydGV4dA=="},
	}

	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.data)
			assert.Error(t, err)
		})
	}
}

func TestTokenEncryptor_ConsistentKeyDerivation(t *testing.T) {
	// Create two encryptors - they should use the same key
	encryptor1, err := NewTokenEncryptor()
	require.NoError(t, err)

	encryptor2, err := NewTokenEncryptor()
	require.NoError(t, err)

	testData := []byte("test data")

	// Encrypt with first encryptor
	encrypted, err := encryptor1.Encrypt(testData)
	require.NoError(t, err)

	// Decrypt with second encryptor (should work if keys are the same)
	decrypted, err := encryptor2.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, testData, decrypted)
}

func TestGetOrCreateSalt(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	saltPath := filepath.Join(tmpDir, ".salt")

	// First call should create salt
	salt1, err := getOrCreateSalt(saltPath)
	require.NoError(t, err)
	assert.Len(t, salt1, SaltSize)

	// File should exist
	_, err = os.Stat(saltPath)
	require.NoError(t, err)

	// Second call should return same salt
	salt2, err := getOrCreateSalt(saltPath)
	require.NoError(t, err)
	assert.Equal(t, salt1, salt2)
}

func TestGetMachineID(t *testing.T) {
	// Machine ID should not be empty
	machineID, err := getMachineID()
	require.NoError(t, err)
	assert.NotEmpty(t, machineID)
}

func TestTokenEncryptor_LargeData(t *testing.T) {
	encryptor, err := NewTokenEncryptor()
	require.NoError(t, err)

	// Test with large data (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, err := encryptor.Encrypt(largeData)
	require.NoError(t, err)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, largeData, decrypted)
}
