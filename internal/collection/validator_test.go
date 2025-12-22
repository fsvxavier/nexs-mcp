package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator("/test/path")
	require.NotNil(t, validator)
	assert.Equal(t, "/test/path", validator.basePath)
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"Valid simple", "user@example.com", true},
		{"Valid with subdomain", "user@mail.example.com", true},
		{"Valid with plus", "user+tag@example.com", true},
		{"Valid with dots", "first.last@example.com", true},
		{"Invalid no @", "userexample.com", false},
		{"Invalid no domain", "user@", false},
		{"Invalid no user", "@example.com", false},
		{"Invalid empty", "", false},
		{"Invalid spaces", "user @example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{"Valid HTTP", "http://example.com", true},
		{"Valid HTTPS", "https://example.com", true},
		{"Valid with path", "https://example.com/path", true},
		{"Valid with query", "https://example.com?key=value", true},
		{"Invalid no scheme", "example.com", false},
		{"Invalid empty", "", false},
		{"Invalid invalid scheme", "ftp://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidURL(tt.url)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{"Valid simple", "1.0.0", true},
		{"Valid with patch", "1.2.3", true},
		{"Valid with prerelease", "1.0.0-beta", true},
		{"Valid with prerelease number", "2.0.0-rc.1", true},
		{"Valid with prerelease complex", "1.0.0-alpha.beta", true},
		{"Invalid no dots", "100", false},
		{"Invalid one dot", "1.0", false},
		{"Invalid letters", "v1.0.0", false},
		{"Invalid empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVersion(tt.version)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidCollectionName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"Valid simple", "my-collection", true},
		{"Valid with numbers", "collection-123", true},
		{"Valid all lowercase", "mycollection", true},
		{"Invalid uppercase", "My-Collection", false},
		{"Invalid spaces", "my collection", false},
		{"Invalid underscore", "my_collection", false},
		{"Invalid empty", "", false},
		{"Invalid starts with hyphen", "-collection", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidCollectionName(tt.input)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidElementType(t *testing.T) {
	tests := []struct {
		name    string
		eleType string
		valid   bool
	}{
		{"Valid persona", "persona", true},
		{"Valid skill", "skill", true},
		{"Valid template", "template", true},
		{"Valid agent", "agent", true},
		{"Valid memory", "memory", true},
		{"Valid ensemble", "ensemble", true},
		{"Invalid unknown", "unknown", false},
		{"Invalid empty", "", false},
		{"Invalid uppercase", "PERSONA", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidElementType(tt.eleType)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidDependencyURI(t *testing.T) {
	tests := []struct {
		name  string
		uri   string
		valid bool
	}{
		{"Valid GitHub", "github://owner/repo", true},
		{"Valid file", "file:///path/to/collection", true},
		{"Valid HTTP", "http://example.com/collection", true},
		{"Valid HTTPS", "https://example.com/collection", true},
		{"Invalid no scheme", "owner/repo", false},
		{"Invalid empty", "", false},
		{"Invalid local scheme", "local://path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidDependencyURI(tt.uri)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidHookType(t *testing.T) {
	tests := []struct {
		name     string
		hookType string
		valid    bool
	}{
		{"Valid command", "command", true},
		{"Valid validate", "validate", true},
		{"Valid backup", "backup", true},
		{"Valid confirm", "confirm", true},
		{"Invalid unknown", "unknown-hook", false},
		{"Invalid empty", "", false},
		{"Invalid pre-install", "pre-install", false}, // These are hook groups, not types
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidHookType(tt.hookType)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidVersionConstraint(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		valid      bool
	}{
		{"Valid exact", "1.0.0", true},
		{"Valid greater equal", ">=1.0.0", true},
		{"Valid less", "<2.0.0", true},
		{"Valid caret", "^1.0.0", true},
		{"Valid tilde", "~1.0.0", true},
		{"Valid wildcard", "*", true},
		{"Invalid empty", "", false},
		{"Invalid bad version", ">abc", false},
		{"Invalid range", ">=1.0.0 <2.0.0", false}, // Not supported by pattern
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVersionConstraint(tt.constraint)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestValidateManifest_ValidMinimal(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test description for collection validation testing",
		Elements: []Element{
			{
				Type: "persona",
				Path: "personas/test.yaml",
			},
		},
	}

	err := validator.ValidateManifest(manifest)
	assert.NoError(t, err)
}

func TestValidateManifest_MissingName(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test description for validation",
		Elements:    []Element{{Type: "persona", Path: "test.yaml"}},
	}

	err := validator.ValidateManifest(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestValidateManifest_MissingVersion(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "test-collection",
		Author:      "Test Author",
		Description: "Test description for validation",
		Elements:    []Element{{Type: "persona", Path: "test.yaml"}},
	}

	err := validator.ValidateManifest(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestValidateManifest_InvalidVersion(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "test-collection",
		Version:     "invalid",
		Author:      "Test Author",
		Description: "Test description for validation",
		Elements:    []Element{{Type: "persona", Path: "test.yaml"}},
	}

	err := validator.ValidateManifest(manifest)
	assert.Error(t, err)
}

func TestValidateManifest_NoElements(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test description for validation",
		Elements:    []Element{},
	}

	err := validator.ValidateManifest(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "element")
}

func TestValidateComprehensive_Valid(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test description for comprehensive validation testing",
		Elements: []Element{
			{Type: "persona", Path: "personas/test.yaml"},
		},
	}

	result := validator.ValidateComprehensive(manifest)
	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateComprehensive_MultipleErrors(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "",        // Missing
		Version:     "invalid", // Invalid format
		Author:      "",
		Description: "",
		Elements:    []Element{},
	}

	result := validator.ValidateComprehensive(manifest)
	require.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.GreaterOrEqual(t, len(result.Errors), 4)
}

func TestValidationError_Structure(t *testing.T) {
	err := &ValidationError{
		Field:    "name",
		Rule:     "required",
		Message:  "name is required",
		Severity: "error",
		Path:     "$.name",
		Fix:      "Add a collection name",
	}

	assert.Equal(t, "name", err.Field)
	assert.Equal(t, "required", err.Rule)
	assert.Equal(t, "error", err.Severity)
	assert.NotEmpty(t, err.Message)
}

func TestValidationResult_Structure(t *testing.T) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []*ValidationError{},
		Warnings: []*ValidationError{},
		Stats:    map[string]int{"total": 10},
	}

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.NotNil(t, result.Stats)
}

func TestValidateComprehensive_WithWarnings(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Short", // Too short - should generate warning
		Elements: []Element{
			{Type: "persona", Path: "test.yaml"},
		},
	}

	result := validator.ValidateComprehensive(manifest)
	require.NotNil(t, result)
	// Should be valid (warnings don't invalidate)
	assert.True(t, result.Valid)
	// But should have warnings
	assert.NotEmpty(t, result.Warnings)
}

func TestIsValidCollectionName_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"Too short", "ab", true}, // Pattern allows 2 chars
		{"Min length", "abc", true},
		{"Max length 64", "a-very-long-collection-name-that-is-exactly-sixty-four-chars-x", true},       // No length check in regex
		{"Too long", "a-very-long-collection-name-that-exceeds-the-maximum-allowed-length-limit", true}, // No length check in regex
		{"Special chars", "my@collection", false},
		{"Ends with hyphen", "collection-", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidCollectionName(tt.input)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidEmail_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"Multiple @", "user@@example.com", false},
		{"Dot at end", "user@example.", false},
		{"Consecutive dots", "user..name@example.com", true}, // Pattern allows consecutive dots
		{"Valid with numbers", "user123@example456.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidVersion_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{"Zero version", "0.0.0", true},
		{"Large numbers", "999.999.999", true},
		{"Complex prerelease", "1.0.0-alpha.1", true},
		{"Leading zeros", "01.02.03", true}, // Pattern allows leading zeros
		{"Negative", "-1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVersion(tt.version)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestValidateManifest_WithMaintainers(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test description for validation with maintainers",
		Elements:    []Element{{Type: "persona", Path: "test.yaml"}},
		Maintainers: []Maintainer{
			{Name: "Maintainer 1", Email: "valid@example.com"},
		},
	}

	err := validator.ValidateManifest(manifest)
	assert.NoError(t, err)
}

func TestValidateManifest_WithInvalidMaintainerEmail(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test description for validation",
		Elements:    []Element{{Type: "persona", Path: "test.yaml"}},
		Maintainers: []Maintainer{
			{Name: "Maintainer 1", Email: "invalid-email"},
		},
	}

	result := validator.ValidateComprehensive(manifest)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestValidateComprehensive_Stats(t *testing.T) {
	validator := NewValidator("/test")
	manifest := &Manifest{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test description for stats validation",
		Elements:    []Element{{Type: "persona", Path: "test.yaml"}},
	}

	result := validator.ValidateComprehensive(manifest)
	require.NotNil(t, result)
	assert.NotNil(t, result.Stats)
	assert.Contains(t, result.Stats, "schema")
	assert.Contains(t, result.Stats, "errors")
	assert.Contains(t, result.Stats, "warnings")
}
