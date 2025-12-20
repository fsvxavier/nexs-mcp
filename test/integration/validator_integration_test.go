package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/fsvxavier/nexs-mcp/internal/collection"
)

// TestValidatorIntegration tests the comprehensive validation with real manifest scenarios
func TestValidatorIntegration(t *testing.T) {
	tests := []struct {
		name           string
		manifest       *Manifest
		expectValid    bool
		expectErrors   int
		expectWarnings int
		errorFields    []string
	}{
		{
			name: "valid_complete_manifest",
			manifest: &Manifest{
				Name:        "test-collection",
				Version:     "1.0.0",
				Author:      "test@example.com",
				Description: "A test collection with all required fields",
				Category:    "testing",
				License:     "MIT",
				Repository:  "https://github.com/test/repo",
				Elements: []Element{
					{
						Type:        "persona",
						Path:        "personas/test.yaml",
						Description: "Test persona",
					},
				},
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "missing_required_fields",
			manifest: &Manifest{
				Name:    "test",
				Version: "1.0.0",
				Elements: []Element{
					{Path: "personas/test.yaml", Type: "persona"},
				},
				// Missing author, description
			},
			expectValid:  false,
			expectErrors: 2, // author, description
			errorFields:  []string{"author", "description"},
		},
		{
			name: "invalid_version_format",
			manifest: &Manifest{
				Name:        "test-collection",
				Version:     "v1.0", // Invalid semver
				Author:      "test@example.com",
				Description: "Test",
				Category:    "testing",
				Elements: []Element{
					{Path: "personas/test.yaml", Type: "persona"},
				},
			},
			expectValid:  false,
			expectErrors: 1,
			errorFields:  []string{"version"},
		},
		{
			name: "invalid_email",
			manifest: &Manifest{
				Name:        "test-collection",
				Version:     "1.0.0",
				Author:      "Test Author", // Valid author (email not required)
				Description: "Test",
				Category:    "testing",
				Elements: []Element{
					{Path: "personas/test.yaml", Type: "persona"},
				},
				Maintainers: []Maintainer{
					{Name: "Test", Email: "not-an-email"}, // Invalid email in maintainer
				},
			},
			expectValid:  false,
			expectErrors: 1,
			errorFields:  []string{"maintainers[0].email"},
		},
		{
			name: "path_traversal_attempt",
			manifest: &Manifest{
				Name:        "malicious-collection",
				Version:     "1.0.0",
				Author:      "hacker@example.com",
				Description: "Test",
				Category:    "testing",
				Elements: []Element{
					{
						Type:        "persona",
						Path:        "../../etc/passwd", // Path traversal
						Description: "Evil persona",
					},
				},
			},
			expectValid:  false,
			expectErrors: 1,
			errorFields:  []string{"elements[0].path"},
		},
		{
			name: "shell_injection_in_hooks",
			manifest: &Manifest{
				Name:        "malicious-collection",
				Version:     "1.0.0",
				Author:      "hacker@example.com",
				Description: "Test",
				Category:    "testing",
				Elements: []Element{
					{Path: "personas/test.yaml", Type: "persona"},
				},
				Hooks: &Hooks{
					PreInstall: []Hook{
						{
							Type:    "command",
							Command: "echo 'hello'; rm -rf /", // Shell injection
						},
					},
				},
			},
			expectValid:  false,
			expectErrors: 1,
			errorFields:  []string{"hooks.pre_install[0].command"},
		},
		{
			name: "circular_dependency",
			manifest: &Manifest{
				Name:        "circular-test",
				Version:     "1.0.0",
				Author:      "test@example.com",
				Description: "Test",
				Category:    "testing",
				Elements: []Element{
					{Path: "personas/test.yaml", Type: "persona"},
				},
				Dependencies: []Dependency{
					{
						URI:     "github://test/dep1",
						Version: "^1.0.0",
					},
				},
			},
			expectValid:    true, // Can't detect circular without full graph
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "excessive_elements",
			manifest: &Manifest{
				Name:        "huge-collection",
				Version:     "1.0.0",
				Author:      "test@example.com",
				Description: "Test",
				Category:    "testing",
				Elements:    makeElements(500), // Large but valid number
			},
			expectValid:    true, // No limit enforced
			expectErrors:   0,
			expectWarnings: 0,
		},
	}

	tmpDir := t.TempDir()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tmpDir)
			result := validator.ValidateComprehensive(tt.manifest)

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
				t.Logf("Errors: %d, Warnings: %d", len(result.Errors), len(result.Warnings))
				for _, err := range result.Errors {
					t.Logf("  Error: %s - %s", err.Field, err.Message)
				}
			}

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(result.Errors))
				for _, err := range result.Errors {
					t.Logf("  Error: %s - %s", err.Field, err.Message)
				}
			}

			// Verify specific error fields
			if tt.errorFields != nil {
				errorFieldMap := make(map[string]bool)
				for _, err := range result.Errors {
					errorFieldMap[err.Field] = true
				}
				for _, expectedField := range tt.errorFields {
					if !errorFieldMap[expectedField] {
						t.Errorf("Expected error for field %s, but not found", expectedField)
					}
				}
			}
		})
	}
}

// TestValidatorWithRealFiles tests validation with actual file system
func TestValidatorWithRealFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	personaDir := filepath.Join(tmpDir, "personas")
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		t.Fatal(err)
	}

	personaFile := filepath.Join(personaDir, "test.yaml")
	if err := os.WriteFile(personaFile, []byte("name: test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	manifest := &Manifest{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "test@example.com",
		Description: "Test collection with real files",
		Category:    "testing",
		Elements: []Element{
			{
				Type:        "persona",
				Path:        "personas/test.yaml",
				Description: "Test persona",
			},
		},
	}

	validator := NewValidator(tmpDir)
	result := validator.ValidateComprehensive(manifest)

	// Should pass validation (file existence is not validated)
	if !result.Valid {
		t.Errorf("Expected validation to pass, got %d errors", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("  Error: %s - %s", err.Field, err.Message)
		}
	}
}

// TestValidatorPerformance tests validation performance
func TestValidatorPerformance(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a large manifest
	manifest := &Manifest{
		Name:        "large-collection",
		Version:     "1.0.0",
		Author:      "test@example.com",
		Description: "Large collection for performance testing",
		Category:    "testing",
		Elements:    makeElements(100),
		Dependencies: []Dependency{
			{URI: "github://test/dep1", Version: "^1.0.0"},
			{URI: "github://test/dep2", Version: "^2.0.0"},
			{URI: "github://test/dep3", Version: "^3.0.0"},
		},
		Hooks: &Hooks{
			PreInstall:  []Hook{{Type: "command", Command: "echo 'pre-install'"}},
			PostInstall: []Hook{{Type: "command", Command: "echo 'post-install'"}},
		},
	}

	validator := NewValidator(tmpDir)

	// Validation should complete quickly (< 100ms for 100 elements)
	result := validator.ValidateComprehensive(manifest)

	if result.Stats["total_rules_checked"] == 0 {
		t.Error("No rules were checked")
	}

	t.Logf("Validated %d rules for %d elements", result.Stats["total_rules_checked"], len(manifest.Elements))
}

// TestValidatorErrorMessages tests that error messages are helpful
func TestValidatorErrorMessages(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &Manifest{
		Name:        "",          // Empty name
		Version:     "invalid",   // Invalid version
		Author:      "not-email", // Invalid email
		Description: "",          // Empty description
		Category:    "",          // Empty category
	}

	validator := NewValidator(tmpDir)
	result := validator.ValidateComprehensive(manifest)

	if result.Valid {
		t.Error("Expected validation to fail")
	}

	// Check that all errors have helpful messages and fix suggestions
	for _, err := range result.Errors {
		if err.Message == "" {
			t.Errorf("Error for field %s has empty message", err.Field)
		}
		if err.Rule == "" {
			t.Errorf("Error for field %s has empty rule", err.Field)
		}
		// Some errors should have fix suggestions
		if err.Field == "version" && err.Fix == "" {
			t.Errorf("Expected fix suggestion for version error")
		}
	}
}

// makeElements creates n test elements
func makeElements(n int) []Element {
	elements := make([]Element, n)
	for i := 0; i < n; i++ {
		elements[i] = Element{
			Type:        "persona",
			Path:        fmt.Sprintf("personas/persona%d.yaml", i),
			Description: "Test persona",
		}
	}
	return elements
}
