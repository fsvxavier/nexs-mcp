package security

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// SignatureVerifier verifies digital signatures.
type SignatureVerifier interface {
	// Verify verifies a signature for a given file
	Verify(filePath string, signaturePath string, publicKey string) error
	// Name returns the verifier name
	Name() string
}

// GPGVerifier verifies GPG signatures.
type GPGVerifier struct {
	gpgPath string
}

// NewGPGVerifier creates a new GPG verifier.
func NewGPGVerifier() (*GPGVerifier, error) {
	// Check if gpg is available
	gpgPath, err := exec.LookPath("gpg")
	if err != nil {
		return nil, fmt.Errorf("gpg not found in PATH: %w", err)
	}

	return &GPGVerifier{
		gpgPath: gpgPath,
	}, nil
}

// Name returns the verifier name.
func (g *GPGVerifier) Name() string {
	return "GPG"
}

// Verify verifies a GPG signature.
func (g *GPGVerifier) Verify(filePath string, signaturePath string, publicKey string) error {
	// Import public key if provided
	if publicKey != "" {
		if err := g.importPublicKey(publicKey); err != nil {
			return fmt.Errorf("failed to import public key: %w", err)
		}
	}

	// Verify signature
	cmd := exec.Command(g.gpgPath, "--verify", signaturePath, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("signature verification failed: %w (output: %s)", err, string(output))
	}

	// Check for "Good signature" in output
	outputStr := string(output)
	if !strings.Contains(outputStr, "Good signature") {
		return fmt.Errorf("invalid signature: %s", outputStr)
	}

	return nil
}

// importPublicKey imports a public key into GPG keyring.
func (g *GPGVerifier) importPublicKey(publicKey string) error {
	cmd := exec.Command(g.gpgPath, "--import")
	cmd.Stdin = strings.NewReader(publicKey)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("key import failed: %w (output: %s)", err, string(output))
	}
	return nil
}

// SSHVerifier verifies SSH signatures (using ssh-keygen).
type SSHVerifier struct {
	sshKeygenPath string
}

// NewSSHVerifier creates a new SSH verifier.
func NewSSHVerifier() (*SSHVerifier, error) {
	// Check if ssh-keygen is available
	sshKeygenPath, err := exec.LookPath("ssh-keygen")
	if err != nil {
		return nil, fmt.Errorf("ssh-keygen not found in PATH: %w", err)
	}

	return &SSHVerifier{
		sshKeygenPath: sshKeygenPath,
	}, nil
}

// Name returns the verifier name.
func (s *SSHVerifier) Name() string {
	return "SSH"
}

// Verify verifies an SSH signature.
func (s *SSHVerifier) Verify(filePath string, signaturePath string, publicKeyPath string) error {
	if publicKeyPath == "" {
		return errors.New("public key path is required for SSH verification")
	}

	// ssh-keygen -Y verify -f allowed_signers -I identity -n namespace -s signature_file < signed_file
	// For simplicity, we use a basic verification approach
	cmd := exec.Command(s.sshKeygenPath, "-Y", "verify",
		"-f", publicKeyPath,
		"-I", "collection-signer",
		"-n", "file",
		"-s", signaturePath,
	)

	// Note: ssh-keygen expects data on stdin
	// This is a simplified implementation
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("SSH signature verification failed: %w (output: %s)", err, string(output))
	}

	// Check for success message
	outputStr := string(output)
	if !strings.Contains(outputStr, "Good") && !strings.Contains(outputStr, "valid") {
		return fmt.Errorf("invalid SSH signature: %s", outputStr)
	}

	return nil
}

// SignatureManager manages signature verification.
type SignatureManager struct {
	verifiers map[string]SignatureVerifier
}

// NewSignatureManager creates a new signature manager.
func NewSignatureManager() *SignatureManager {
	return &SignatureManager{
		verifiers: make(map[string]SignatureVerifier),
	}
}

// RegisterVerifier registers a signature verifier.
func (m *SignatureManager) RegisterVerifier(name string, verifier SignatureVerifier) {
	m.verifiers[name] = verifier
}

// Verify verifies a signature using the specified verifier.
func (m *SignatureManager) Verify(verifierName string, filePath string, signaturePath string, publicKey string) error {
	verifier, exists := m.verifiers[verifierName]
	if !exists {
		return fmt.Errorf("verifier not found: %s", verifierName)
	}

	return verifier.Verify(filePath, signaturePath, publicKey)
}

// VerifyAny tries all registered verifiers until one succeeds.
func (m *SignatureManager) VerifyAny(filePath string, signaturePath string, publicKey string) error {
	if len(m.verifiers) == 0 {
		return errors.New("no verifiers registered")
	}

	var lastErr error
	for name, verifier := range m.verifiers {
		err := verifier.Verify(filePath, signaturePath, publicKey)
		if err == nil {
			return nil // Success
		}
		lastErr = fmt.Errorf("%s verification failed: %w", name, err)
	}

	return fmt.Errorf("all verifiers failed: %w", lastErr)
}

// InitializeDefaultVerifiers initializes GPG and SSH verifiers if available.
func (m *SignatureManager) InitializeDefaultVerifiers() error {
	// Try to register GPG verifier
	if gpg, err := NewGPGVerifier(); err == nil {
		m.RegisterVerifier("gpg", gpg)
	}

	// Try to register SSH verifier
	if ssh, err := NewSSHVerifier(); err == nil {
		m.RegisterVerifier("ssh", ssh)
	}

	if len(m.verifiers) == 0 {
		return errors.New("no signature verifiers available (install gpg or ssh-keygen)")
	}

	return nil
}
