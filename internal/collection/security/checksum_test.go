package security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChecksumValidator_DefaultAlgorithm(t *testing.T) {
	validator := NewChecksumValidator("")
	assert.Equal(t, SHA256, validator.Algorithm())
}

func TestNewChecksumValidator_WithAlgorithm(t *testing.T) {
	tests := []struct {
		name      string
		algorithm ChecksumAlgorithm
	}{
		{"SHA256", SHA256},
		{"SHA512", SHA512},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewChecksumValidator(tt.algorithm)
			assert.Equal(t, tt.algorithm, validator.Algorithm())
		})
	}
}

func TestComputeString_SHA256(t *testing.T) {
	validator := NewChecksumValidator(SHA256)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "Simple text",
			input:    "hello world",
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:     "JSON data",
			input:    `{"name":"test","version":"1.0.0"}`,
			expected: "55996d4e16775131502bf8df7cb40d6ee00ab82aae087e45ad23c72e641ff399",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checksum, err := validator.ComputeString(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, checksum)
		})
	}
}

func TestComputeString_SHA512(t *testing.T) {
	validator := NewChecksumValidator(SHA512)

	checksum, err := validator.ComputeString("hello world")
	require.NoError(t, err)

	// SHA-512 produces a 128-character hex string (64 bytes)
	assert.Len(t, checksum, 128)
	assert.Equal(t, "309ecc489c12d6eb4cc40f50c902f2b4d0ed77ee511a7c7a9bcd3ca86d4cd86f989dd35bc5ff499670da34255b45b0cfd830e81f605dcf7dc5542e93ae9cd76f", checksum)
}

func TestCompute_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create test file
	content := "test file content"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	validator := NewChecksumValidator(SHA256)
	checksum, err := validator.Compute(testFile)

	require.NoError(t, err)
	assert.NotEmpty(t, checksum)
	assert.Len(t, checksum, 64) // SHA-256 produces 64 hex chars
}

func TestCompute_NonExistentFile(t *testing.T) {
	validator := NewChecksumValidator(SHA256)
	_, err := validator.Compute("/nonexistent/file.txt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestCompute_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	emptyFile := filepath.Join(tempDir, "empty.txt")

	require.NoError(t, os.WriteFile(emptyFile, []byte{}, 0644))

	validator := NewChecksumValidator(SHA256)
	checksum, err := validator.Compute(emptyFile)

	require.NoError(t, err)
	// SHA-256 of empty file
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", checksum)
}

func TestCompute_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	largeFile := filepath.Join(tempDir, "large.bin")

	// Create 1MB file
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	require.NoError(t, os.WriteFile(largeFile, data, 0644))

	validator := NewChecksumValidator(SHA256)
	checksum, err := validator.Compute(largeFile)

	require.NoError(t, err)
	assert.NotEmpty(t, checksum)
	assert.Len(t, checksum, 64)
}

func TestValidate_Success(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	content := "hello world"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	validator := NewChecksumValidator(SHA256)
	expectedChecksum := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	err := validator.Validate(testFile, expectedChecksum)
	assert.NoError(t, err)
}

func TestValidate_ChecksumMismatch(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	content := "hello world"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	validator := NewChecksumValidator(SHA256)
	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"

	err := validator.Validate(testFile, wrongChecksum)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestValidate_CaseInsensitive(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	content := "hello"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	validator := NewChecksumValidator(SHA256)

	// Test with uppercase checksum
	uppercaseChecksum := "2CF24DBA5FB0A30E26E83B2AC5B9E29E1B161E5C1FA7425E73043362938B9824"
	err := validator.Validate(testFile, uppercaseChecksum)
	assert.NoError(t, err)

	// Test with mixed case
	mixedChecksum := "2Cf24DbA5Fb0A30E26E83B2Ac5B9E29E1B161E5C1Fa7425E73043362938B9824"
	err = validator.Validate(testFile, mixedChecksum)
	assert.NoError(t, err)
}

func TestValidate_WithWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	content := "hello"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	validator := NewChecksumValidator(SHA256)
	checksumWithSpaces := "  2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824  "

	err := validator.Validate(testFile, checksumWithSpaces)
	assert.NoError(t, err)
}

func TestValidateBytes_Success(t *testing.T) {
	validator := NewChecksumValidator(SHA256)
	data := []byte("hello world")
	expectedChecksum := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	err := validator.ValidateBytes(data, expectedChecksum)
	assert.NoError(t, err)
}

func TestValidateBytes_Mismatch(t *testing.T) {
	validator := NewChecksumValidator(SHA256)
	data := []byte("hello world")
	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"

	err := validator.ValidateBytes(data, wrongChecksum)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestValidateBytes_EmptyData(t *testing.T) {
	validator := NewChecksumValidator(SHA256)
	data := []byte{}
	expectedChecksum := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	err := validator.ValidateBytes(data, expectedChecksum)
	assert.NoError(t, err)
}

func TestUnsupportedAlgorithm_ComputeString(t *testing.T) {
	validator := &ChecksumValidator{algorithm: "md5"} // Unsupported

	_, err := validator.ComputeString("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported algorithm")
}

func TestUnsupportedAlgorithm_Compute(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	validator := &ChecksumValidator{algorithm: "md5"} // Unsupported

	_, err := validator.Compute(testFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported algorithm")
}

func TestUnsupportedAlgorithm_ValidateBytes(t *testing.T) {
	validator := &ChecksumValidator{algorithm: "md5"} // Unsupported

	err := validator.ValidateBytes([]byte("test"), "abc123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported algorithm")
}

func TestChecksumAlgorithm_Constants(t *testing.T) {
	assert.Equal(t, ChecksumAlgorithm("sha256"), SHA256)
	assert.Equal(t, ChecksumAlgorithm("sha512"), SHA512)
}

func TestCompute_SHA256_vs_SHA512_Different(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test data"), 0644))

	validator256 := NewChecksumValidator(SHA256)
	checksum256, err := validator256.Compute(testFile)
	require.NoError(t, err)

	validator512 := NewChecksumValidator(SHA512)
	checksum512, err := validator512.Compute(testFile)
	require.NoError(t, err)

	// Different algorithms should produce different checksums
	assert.NotEqual(t, checksum256, checksum512)
	// SHA-256 produces 64 hex chars, SHA-512 produces 128
	assert.Len(t, checksum256, 64)
	assert.Len(t, checksum512, 128)
}

func TestComputeString_Deterministic(t *testing.T) {
	validator := NewChecksumValidator(SHA256)
	input := "deterministic test"

	// Compute multiple times
	checksum1, err1 := validator.ComputeString(input)
	checksum2, err2 := validator.ComputeString(input)
	checksum3, err3 := validator.ComputeString(input)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	// Should always produce the same result
	assert.Equal(t, checksum1, checksum2)
	assert.Equal(t, checksum2, checksum3)
}

func TestValidate_FileNotFound(t *testing.T) {
	validator := NewChecksumValidator(SHA256)

	err := validator.Validate("/nonexistent/file.txt", "abc123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to compute checksum")
}

func TestAlgorithm_Getter(t *testing.T) {
	validator256 := NewChecksumValidator(SHA256)
	assert.Equal(t, SHA256, validator256.Algorithm())

	validator512 := NewChecksumValidator(SHA512)
	assert.Equal(t, SHA512, validator512.Algorithm())
}
