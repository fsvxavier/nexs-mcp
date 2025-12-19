package security

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// ChecksumAlgorithm represents a cryptographic hash algorithm
type ChecksumAlgorithm string

const (
	// SHA256 is the SHA-256 algorithm (recommended)
	SHA256 ChecksumAlgorithm = "sha256"
	// SHA512 is the SHA-512 algorithm (more secure but slower)
	SHA512 ChecksumAlgorithm = "sha512"
)

// ChecksumValidator validates file integrity using cryptographic hashes
type ChecksumValidator struct {
	algorithm ChecksumAlgorithm
}

// NewChecksumValidator creates a new checksum validator
func NewChecksumValidator(algorithm ChecksumAlgorithm) *ChecksumValidator {
	if algorithm == "" {
		algorithm = SHA256 // Default to SHA-256
	}
	return &ChecksumValidator{
		algorithm: algorithm,
	}
}

// Validate verifies that a file matches the expected checksum
func (c *ChecksumValidator) Validate(filePath string, expectedChecksum string) error {
	// Compute actual checksum
	actualChecksum, err := c.Compute(filePath)
	if err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}

	// Normalize checksums (lowercase, remove whitespace)
	expected := strings.ToLower(strings.TrimSpace(expectedChecksum))
	actual := strings.ToLower(strings.TrimSpace(actualChecksum))

	// Compare
	if expected != actual {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

// Compute calculates the checksum of a file
func (c *ChecksumValidator) Compute(filePath string) (string, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create hash
	var hasher hash.Hash
	switch c.algorithm {
	case SHA256:
		hasher = sha256.New()
	case SHA512:
		hasher = sha512.New()
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", c.algorithm)
	}

	// Copy file data to hasher
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	// Get checksum as hex string
	checksum := hex.EncodeToString(hasher.Sum(nil))
	return checksum, nil
}

// ComputeString calculates the checksum of a string
func (c *ChecksumValidator) ComputeString(data string) (string, error) {
	var hasher hash.Hash
	switch c.algorithm {
	case SHA256:
		hasher = sha256.New()
	case SHA512:
		hasher = sha512.New()
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", c.algorithm)
	}

	hasher.Write([]byte(data))
	checksum := hex.EncodeToString(hasher.Sum(nil))
	return checksum, nil
}

// ValidateBytes verifies that data matches the expected checksum
func (c *ChecksumValidator) ValidateBytes(data []byte, expectedChecksum string) error {
	var hasher hash.Hash
	switch c.algorithm {
	case SHA256:
		hasher = sha256.New()
	case SHA512:
		hasher = sha512.New()
	default:
		return fmt.Errorf("unsupported algorithm: %s", c.algorithm)
	}

	hasher.Write(data)
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Normalize and compare
	expected := strings.ToLower(strings.TrimSpace(expectedChecksum))
	actual := strings.ToLower(strings.TrimSpace(actualChecksum))

	if expected != actual {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

// Algorithm returns the current algorithm
func (c *ChecksumValidator) Algorithm() ChecksumAlgorithm {
	return c.algorithm
}
