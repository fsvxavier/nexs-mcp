package security

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGPGVerifier(t *testing.T) {
	verifier, err := NewGPGVerifier()

	// Check if gpg is available
	_, lookupErr := exec.LookPath("gpg")

	if lookupErr != nil {
		// GPG not available - should return error
		assert.Error(t, err)
		assert.Nil(t, verifier)
		assert.Contains(t, err.Error(), "gpg not found")
	} else {
		// GPG available - should succeed
		require.NoError(t, err)
		require.NotNil(t, verifier)
		assert.NotEmpty(t, verifier.gpgPath)
		assert.Equal(t, "GPG", verifier.Name())
	}
}

func TestGPGVerifier_Name(t *testing.T) {
	verifier, err := NewGPGVerifier()
	if err != nil {
		t.Skip("GPG not available")
	}

	assert.Equal(t, "GPG", verifier.Name())
}

func TestNewSSHVerifier(t *testing.T) {
	verifier, err := NewSSHVerifier()

	// Check if ssh-keygen is available
	_, lookupErr := exec.LookPath("ssh-keygen")

	if lookupErr != nil {
		// ssh-keygen not available - should return error
		assert.Error(t, err)
		assert.Nil(t, verifier)
		assert.Contains(t, err.Error(), "ssh-keygen not found")
	} else {
		// ssh-keygen available - should succeed
		require.NoError(t, err)
		require.NotNil(t, verifier)
		assert.NotEmpty(t, verifier.sshKeygenPath)
		assert.Equal(t, "SSH", verifier.Name())
	}
}

func TestSSHVerifier_Name(t *testing.T) {
	verifier, err := NewSSHVerifier()
	if err != nil {
		t.Skip("SSH verifier not available")
	}

	assert.Equal(t, "SSH", verifier.Name())
}

func TestSSHVerifier_Verify_MissingPublicKey(t *testing.T) {
	verifier, err := NewSSHVerifier()
	if err != nil {
		t.Skip("SSH verifier not available")
	}

	tmpDir := t.TempDir()

	// Create dummy files
	dataFile := filepath.Join(tmpDir, "data.txt")
	sigFile := filepath.Join(tmpDir, "data.txt.sig")

	err = os.WriteFile(dataFile, []byte("test data"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(sigFile, []byte("fake signature"), 0644)
	require.NoError(t, err)

	// Verify with empty public key - should error
	err = verifier.Verify(dataFile, sigFile, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public key path is required")
}

func TestNewSignatureManager(t *testing.T) {
	manager := NewSignatureManager()
	require.NotNil(t, manager)
	assert.NotNil(t, manager.verifiers)
	assert.Empty(t, manager.verifiers)
}

func TestSignatureManager_RegisterVerifier(t *testing.T) {
	manager := NewSignatureManager()

	// Create a mock verifier
	verifier, err := NewGPGVerifier()
	if err != nil {
		t.Skip("GPG not available")
	}

	manager.RegisterVerifier("gpg", verifier)

	assert.Len(t, manager.verifiers, 1)
	assert.Contains(t, manager.verifiers, "gpg")
	assert.Equal(t, verifier, manager.verifiers["gpg"])
}

func TestSignatureManager_RegisterMultipleVerifiers(t *testing.T) {
	manager := NewSignatureManager()

	gpgVerifier, gpgErr := NewGPGVerifier()
	sshVerifier, sshErr := NewSSHVerifier()

	count := 0
	if gpgErr == nil {
		manager.RegisterVerifier("gpg", gpgVerifier)
		count++
	}
	if sshErr == nil {
		manager.RegisterVerifier("ssh", sshVerifier)
		count++
	}

	if count == 0 {
		t.Skip("No verifiers available")
	}

	assert.Len(t, manager.verifiers, count)
}

func TestSignatureManager_Verify_VerifierNotFound(t *testing.T) {
	manager := NewSignatureManager()

	err := manager.Verify("nonexistent", "/path/to/file", "/path/to/sig", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verifier not found")
}

func TestSignatureManager_Verify_WithRegisteredVerifier(t *testing.T) {
	manager := NewSignatureManager()

	verifier, err := NewGPGVerifier()
	if err != nil {
		t.Skip("GPG not available")
	}

	manager.RegisterVerifier("gpg", verifier)

	tmpDir := t.TempDir()

	dataFile := filepath.Join(tmpDir, "data.txt")
	sigFile := filepath.Join(tmpDir, "data.txt.sig")

	err = os.WriteFile(dataFile, []byte("test"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(sigFile, []byte("fake"), 0644)
	require.NoError(t, err)

	// This will fail because we don't have valid signature
	// but we're testing that the verifier is called
	err = manager.Verify("gpg", dataFile, sigFile, "")
	assert.Error(t, err, "Should fail with invalid signature")
}

func TestSignatureManager_VerifyAny_NoVerifiers(t *testing.T) {
	manager := NewSignatureManager()

	err := manager.VerifyAny("/path/to/file", "/path/to/sig", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no verifiers registered")
}

func TestSignatureManager_VerifyAny_AllFail(t *testing.T) {
	manager := NewSignatureManager()

	gpgVerifier, gpgErr := NewGPGVerifier()
	sshVerifier, sshErr := NewSSHVerifier()

	if gpgErr != nil && sshErr != nil {
		t.Skip("No verifiers available")
	}

	if gpgErr == nil {
		manager.RegisterVerifier("gpg", gpgVerifier)
	}
	if sshErr == nil {
		manager.RegisterVerifier("ssh", sshVerifier)
	}

	tmpDir := t.TempDir()

	dataFile := filepath.Join(tmpDir, "data.txt")
	sigFile := filepath.Join(tmpDir, "data.txt.sig")

	err := os.WriteFile(dataFile, []byte("test"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(sigFile, []byte("fake"), 0644)
	require.NoError(t, err)

	// All verifiers should fail with invalid signature
	err = manager.VerifyAny(dataFile, sigFile, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all verifiers failed")
}

func TestSignatureManager_InitializeDefaultVerifiers(t *testing.T) {
	manager := NewSignatureManager()

	err := manager.InitializeDefaultVerifiers()

	// Check if any verifiers are available
	_, gpgErr := exec.LookPath("gpg")
	_, sshErr := exec.LookPath("ssh-keygen")

	if gpgErr != nil && sshErr != nil {
		// No verifiers available
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no signature verifiers available")
		assert.Empty(t, manager.verifiers)
	} else {
		// At least one verifier available
		require.NoError(t, err)
		assert.NotEmpty(t, manager.verifiers)

		if gpgErr == nil {
			assert.Contains(t, manager.verifiers, "gpg")
		}
		if sshErr == nil {
			assert.Contains(t, manager.verifiers, "ssh")
		}
	}
}

func TestGPGVerifier_Verify_NonExistentFile(t *testing.T) {
	verifier, err := NewGPGVerifier()
	if err != nil {
		t.Skip("GPG not available")
	}

	err = verifier.Verify("/nonexistent/file", "/nonexistent/sig", "")
	assert.Error(t, err)
}

func TestGPGVerifier_Verify_InvalidSignature(t *testing.T) {
	verifier, err := NewGPGVerifier()
	if err != nil {
		t.Skip("GPG not available")
	}

	tmpDir := t.TempDir()

	dataFile := filepath.Join(tmpDir, "data.txt")
	sigFile := filepath.Join(tmpDir, "data.txt.sig")

	err = os.WriteFile(dataFile, []byte("test data"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(sigFile, []byte("invalid signature"), 0644)
	require.NoError(t, err)

	err = verifier.Verify(dataFile, sigFile, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")
}

func TestSSHVerifier_Verify_InvalidSignature(t *testing.T) {
	verifier, err := NewSSHVerifier()
	if err != nil {
		t.Skip("SSH verifier not available")
	}

	tmpDir := t.TempDir()

	dataFile := filepath.Join(tmpDir, "data.txt")
	sigFile := filepath.Join(tmpDir, "data.txt.sig")
	keyFile := filepath.Join(tmpDir, "allowed_signers")

	err = os.WriteFile(dataFile, []byte("test data"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(sigFile, []byte("invalid signature"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(keyFile, []byte("fake key"), 0644)
	require.NoError(t, err)

	err = verifier.Verify(dataFile, sigFile, keyFile)
	assert.Error(t, err)
}

func TestSignatureVerifier_Interface(t *testing.T) {
	// Test that GPGVerifier implements SignatureVerifier
	var _ SignatureVerifier = (*GPGVerifier)(nil)

	// Test that SSHVerifier implements SignatureVerifier
	var _ SignatureVerifier = (*SSHVerifier)(nil)
}

func TestGPGVerifier_ImportPublicKey_InvalidKey(t *testing.T) {
	verifier, err := NewGPGVerifier()
	if err != nil {
		t.Skip("GPG not available")
	}

	// Try to import invalid key
	err = verifier.importPublicKey("invalid key data")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key import failed")
}

func TestSignatureManager_OverwriteVerifier(t *testing.T) {
	manager := NewSignatureManager()

	verifier1, err := NewGPGVerifier()
	if err != nil {
		t.Skip("GPG not available")
	}

	manager.RegisterVerifier("test", verifier1)
	assert.Len(t, manager.verifiers, 1)

	// Register another verifier with same name
	verifier2, err := NewGPGVerifier()
	require.NoError(t, err)

	manager.RegisterVerifier("test", verifier2)
	assert.Len(t, manager.verifiers, 1, "Should overwrite, not add")
	assert.Equal(t, verifier2, manager.verifiers["test"])
}

func TestSignatureManager_VerifierTypes(t *testing.T) {
	manager := NewSignatureManager()

	gpgVerifier, gpgErr := NewGPGVerifier()
	sshVerifier, sshErr := NewSSHVerifier()

	if gpgErr == nil {
		manager.RegisterVerifier("gpg", gpgVerifier)
		assert.IsType(t, &GPGVerifier{}, manager.verifiers["gpg"])
	}

	if sshErr == nil {
		manager.RegisterVerifier("ssh", sshVerifier)
		assert.IsType(t, &SSHVerifier{}, manager.verifiers["ssh"])
	}

	if gpgErr != nil && sshErr != nil {
		t.Skip("No verifiers available")
	}
}

func TestGPGVerifier_VerifyWithPublicKey(t *testing.T) {
	verifier, err := NewGPGVerifier()
	if err != nil {
		t.Skip("GPG not available")
	}

	tmpDir := t.TempDir()

	dataFile := filepath.Join(tmpDir, "data.txt")
	sigFile := filepath.Join(tmpDir, "data.txt.sig")

	err = os.WriteFile(dataFile, []byte("test"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(sigFile, []byte("fake sig"), 0644)
	require.NoError(t, err)

	// Verify with invalid public key
	err = verifier.Verify(dataFile, sigFile, "invalid public key")
	assert.Error(t, err, "Should fail with invalid key/signature")
}
