package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrustedSourceRegistry(t *testing.T) {
	registry := NewTrustedSourceRegistry()
	require.NotNil(t, registry)
	assert.NotNil(t, registry.sources)
	assert.NotEmpty(t, registry.sources, "Should have default sources")
}

func TestTrustedSourceRegistry_AddSource(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	source := &TrustedSource{
		Name:     "test-source",
		Pattern:  "github.com/test/*",
		Required: false,
		Verified: true,
		Tags:     []string{"test"},
	}

	err := registry.AddSource(source)
	require.NoError(t, err)

	retrieved, exists := registry.GetSource("test-source")
	assert.True(t, exists)
	assert.Equal(t, source, retrieved)
}

func TestTrustedSourceRegistry_AddSource_EmptyName(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	source := &TrustedSource{
		Name:    "",
		Pattern: "github.com/test/*",
	}

	err := registry.AddSource(source)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source name is required")
}

func TestTrustedSourceRegistry_AddSource_EmptyPattern(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	source := &TrustedSource{
		Name:    "test",
		Pattern: "",
	}

	err := registry.AddSource(source)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source pattern is required")
}

func TestTrustedSourceRegistry_RemoveSource(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	source := &TrustedSource{
		Name:    "removable",
		Pattern: "test/*",
	}

	err := registry.AddSource(source)
	require.NoError(t, err)

	_, exists := registry.GetSource("removable")
	assert.True(t, exists)

	registry.RemoveSource("removable")

	_, exists = registry.GetSource("removable")
	assert.False(t, exists)
}

func TestTrustedSourceRegistry_RemoveNonExistent(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	// Should not panic
	registry.RemoveSource("nonexistent")
}

func TestTrustedSourceRegistry_GetSource(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	source := &TrustedSource{
		Name:     "get-test",
		Pattern:  "github.com/get/*",
		Required: true,
		Verified: false,
	}

	err := registry.AddSource(source)
	require.NoError(t, err)

	retrieved, exists := registry.GetSource("get-test")
	assert.True(t, exists)
	assert.Equal(t, "get-test", retrieved.Name)
	assert.Equal(t, "github.com/get/*", retrieved.Pattern)
	assert.True(t, retrieved.Required)
	assert.False(t, retrieved.Verified)
}

func TestTrustedSourceRegistry_GetNonExistent(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	source, exists := registry.GetSource("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, source)
}

func TestTrustedSourceRegistry_ListSources(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	sources := registry.ListSources()
	assert.NotEmpty(t, sources, "Should have default sources")

	initialCount := len(sources)

	// Add a new source
	newSource := &TrustedSource{
		Name:    "new-source",
		Pattern: "test/*",
	}
	err := registry.AddSource(newSource)
	require.NoError(t, err)

	sources = registry.ListSources()
	assert.Len(t, sources, initialCount+1)
}

func TestTrustedSourceRegistry_IsTrusted(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		uri     string
		trusted bool
	}{
		{
			name:    "Exact match",
			pattern: "github.com/owner/repo",
			uri:     "github.com/owner/repo",
			trusted: true,
		},
		{
			name:    "Wildcard match",
			pattern: "github.com/owner/*",
			uri:     "github.com/owner/repo",
			trusted: true,
		},
		{
			name:    "Deep wildcard match",
			pattern: "github.com/*/*",
			uri:     "github.com/owner/repo",
			trusted: true,
		},
		{
			name:    "No match",
			pattern: "github.com/owner/*",
			uri:     "gitlab.com/owner/repo",
			trusted: false,
		},
		{
			name:    "Prefix not matched",
			pattern: "github.com/owner/*",
			uri:     "github.com/other/repo",
			trusted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewTrustedSourceRegistry()

			source := &TrustedSource{
				Name:    "test",
				Pattern: tt.pattern,
			}
			err := registry.AddSource(source)
			require.NoError(t, err)

			matchedSource, trusted := registry.IsTrusted(tt.uri)
			assert.Equal(t, tt.trusted, trusted)

			if tt.trusted {
				assert.NotNil(t, matchedSource)
				assert.Equal(t, "test", matchedSource.Name)
			} else {
				assert.Nil(t, matchedSource)
			}
		})
	}
}

func TestTrustedSourceRegistry_MatchesPattern(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	tests := []struct {
		pattern string
		uri     string
		matches bool
	}{
		{"github.com/owner/repo", "github.com/owner/repo", true},
		{"github.com/owner/*", "github.com/owner/repo", true},
		{"github.com/owner/*", "github.com/owner/repo/sub", true},
		{"github.com/*", "github.com/anything", true},
		{"*/repo", "github.com/repo", true},
		{"*", "anything", true},
		{"github.com/owner/repo", "github.com/owner/other", false},
		{"github.com/owner/?", "github.com/owner/a", true},
		{"github.com/owner/?", "github.com/owner/ab", false},
		{"file:///*", "file:///path/to/file", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_vs_"+tt.uri, func(t *testing.T) {
			matches := registry.matchesPattern(tt.uri, tt.pattern)
			assert.Equal(t, tt.matches, matches)
		})
	}
}

func TestTrustedSourceRegistry_ValidateURI_NotRequired(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	// URI not in trusted sources, but not required
	err := registry.ValidateURI("untrusted.com/repo", false)
	assert.NoError(t, err)
}

func TestTrustedSourceRegistry_ValidateURI_Required(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	// URI not in trusted sources, but required
	err := registry.ValidateURI("untrusted.com/repo", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not from a trusted source")
}

func TestTrustedSourceRegistry_ValidateURI_TrustedNotRequired(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	source := &TrustedSource{
		Name:    "test",
		Pattern: "trusted.com/*",
	}
	err := registry.AddSource(source)
	require.NoError(t, err)

	// URI in trusted sources
	err = registry.ValidateURI("trusted.com/repo", false)
	assert.NoError(t, err)
}

func TestTrustedSourceRegistry_ValidateURI_TrustedRequired(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	source := &TrustedSource{
		Name:    "test",
		Pattern: "trusted.com/*",
	}
	err := registry.AddSource(source)
	require.NoError(t, err)

	// URI in trusted sources, required check passes
	err = registry.ValidateURI("trusted.com/repo", true)
	assert.NoError(t, err)
}

func TestTrustedSourceRegistry_AddDefaultSources(t *testing.T) {
	registry := &TrustedSourceRegistry{
		sources: make(map[string]*TrustedSource),
	}

	// Initially empty
	assert.Empty(t, registry.sources)

	registry.AddDefaultSources()

	// Should have default sources
	assert.NotEmpty(t, registry.sources)

	// Check for specific default sources
	_, exists := registry.GetSource("nexs-official")
	assert.True(t, exists, "Should have nexs-official source")

	_, exists = registry.GetSource("nexs-org")
	assert.True(t, exists, "Should have nexs-org source")

	_, exists = registry.GetSource("local-filesystem")
	assert.True(t, exists, "Should have local-filesystem source")
}

func TestTrustedSourceRegistry_DefaultSourcesProperties(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	// Test nexs-official
	official, exists := registry.GetSource("nexs-official")
	require.True(t, exists)
	assert.Equal(t, "github.com/fsvxavier/nexs-mcp-collections/*", official.Pattern)
	assert.True(t, official.Verified)
	assert.Contains(t, official.Tags, "official")
	assert.Contains(t, official.Tags, "verified")

	// Test nexs-org
	org, exists := registry.GetSource("nexs-org")
	require.True(t, exists)
	assert.Equal(t, "github.com/nexs-mcp/*", org.Pattern)
	assert.True(t, org.Verified)

	// Test local-filesystem
	local, exists := registry.GetSource("local-filesystem")
	require.True(t, exists)
	assert.Equal(t, "file:///*", local.Pattern)
	assert.False(t, local.Verified)
	assert.Contains(t, local.Tags, "local")
	assert.Contains(t, local.Tags, "development")
}

func TestTrustedSource_Structure(t *testing.T) {
	source := &TrustedSource{
		Name:      "test-source",
		Pattern:   "github.com/test/*",
		PublicKey: "test-key",
		Required:  true,
		Verified:  true,
		Tags:      []string{"tag1", "tag2"},
	}

	assert.Equal(t, "test-source", source.Name)
	assert.Equal(t, "github.com/test/*", source.Pattern)
	assert.Equal(t, "test-key", source.PublicKey)
	assert.True(t, source.Required)
	assert.True(t, source.Verified)
	assert.Len(t, source.Tags, 2)
	assert.Contains(t, source.Tags, "tag1")
	assert.Contains(t, source.Tags, "tag2")
}

func TestNewSecurityConfig(t *testing.T) {
	config := NewSecurityConfig()
	require.NotNil(t, config)

	// Check defaults
	assert.False(t, config.RequireSignatures)
	assert.False(t, config.RequireTrustedSource)
	assert.True(t, config.AllowUnsigned)
	assert.True(t, config.ScanEnabled)
	assert.Equal(t, SeverityCritical, config.ScanThreshold)
	assert.Empty(t, config.TrustedSources)
}

func TestSecurityConfig_Validate_Valid(t *testing.T) {
	config := NewSecurityConfig()

	err := config.Validate()
	assert.NoError(t, err)
}

func TestSecurityConfig_Validate_Conflicting(t *testing.T) {
	config := &SecurityConfig{
		RequireSignatures: true,
		AllowUnsigned:     true,
		ScanThreshold:     SeverityCritical,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicting config")
}

func TestSecurityConfig_Validate_InvalidThreshold(t *testing.T) {
	config := &SecurityConfig{
		RequireSignatures: false,
		AllowUnsigned:     true,
		ScanThreshold:     Severity("invalid"),
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scan threshold")
}

func TestSecurityConfig_Validate_AllThresholds(t *testing.T) {
	thresholds := []Severity{
		SeverityLow,
		SeverityMedium,
		SeverityHigh,
		SeverityCritical,
	}

	for _, threshold := range thresholds {
		t.Run(string(threshold), func(t *testing.T) {
			config := &SecurityConfig{
				ScanThreshold: threshold,
			}

			err := config.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestSecurityConfig_ShouldRequireSignature_AllowUnsigned(t *testing.T) {
	config := &SecurityConfig{
		AllowUnsigned:     true,
		RequireSignatures: false,
	}

	// Should always return false when AllowUnsigned is true
	assert.False(t, config.ShouldRequireSignature("test", nil))
	assert.False(t, config.ShouldRequireSignature("test", &TrustedSource{Required: true}))
}

func TestSecurityConfig_ShouldRequireSignature_RequireGlobally(t *testing.T) {
	config := &SecurityConfig{
		AllowUnsigned:     false,
		RequireSignatures: true,
	}

	// Should always return true when RequireSignatures is true
	assert.True(t, config.ShouldRequireSignature("test", nil))
	assert.True(t, config.ShouldRequireSignature("test", &TrustedSource{Required: false}))
}

func TestSecurityConfig_ShouldRequireSignature_SourceRequired(t *testing.T) {
	config := &SecurityConfig{
		AllowUnsigned:     false,
		RequireSignatures: false,
	}

	// Source not required
	source := &TrustedSource{Required: false}
	assert.False(t, config.ShouldRequireSignature("test", source))

	// Source required
	source = &TrustedSource{Required: true}
	assert.True(t, config.ShouldRequireSignature("test", source))
}

func TestSecurityConfig_ShouldRequireSignature_NoSource(t *testing.T) {
	config := &SecurityConfig{
		AllowUnsigned:     false,
		RequireSignatures: false,
	}

	// No source provided
	assert.False(t, config.ShouldRequireSignature("test", nil))
}

func TestSecurityConfig_CustomTrustedSources(t *testing.T) {
	config := &SecurityConfig{
		TrustedSources: []string{
			"custom.com/*",
			"another.com/*",
		},
	}

	assert.Len(t, config.TrustedSources, 2)
	assert.Contains(t, config.TrustedSources, "custom.com/*")
	assert.Contains(t, config.TrustedSources, "another.com/*")
}

func TestTrustedSourceRegistry_MultipleMatches(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	// Add multiple sources that could match
	source1 := &TrustedSource{
		Name:    "specific",
		Pattern: "github.com/owner/repo",
	}
	source2 := &TrustedSource{
		Name:    "wildcard",
		Pattern: "github.com/owner/*",
	}

	err := registry.AddSource(source1)
	require.NoError(t, err)
	err = registry.AddSource(source2)
	require.NoError(t, err)

	// Should match one of them
	matchedSource, trusted := registry.IsTrusted("github.com/owner/repo")
	assert.True(t, trusted)
	assert.NotNil(t, matchedSource)
	// Either source1 or source2 could match (implementation dependent)
	assert.Contains(t, []string{"specific", "wildcard"}, matchedSource.Name)
}

func TestTrustedSourceRegistry_OverwriteSource(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	source1 := &TrustedSource{
		Name:     "test",
		Pattern:  "pattern1",
		Verified: false,
	}

	err := registry.AddSource(source1)
	require.NoError(t, err)

	source2 := &TrustedSource{
		Name:     "test",
		Pattern:  "pattern2",
		Verified: true,
	}

	err = registry.AddSource(source2)
	require.NoError(t, err)

	// Should overwrite
	retrieved, exists := registry.GetSource("test")
	assert.True(t, exists)
	assert.Equal(t, "pattern2", retrieved.Pattern)
	assert.True(t, retrieved.Verified)
}

func TestTrustedSourceRegistry_PatternEdgeCases(t *testing.T) {
	registry := NewTrustedSourceRegistry()

	tests := []struct {
		pattern string
		uri     string
		matches bool
	}{
		// Empty pattern
		{"", "anything", false},
		// Special characters
		{"github.com/owner/repo.git", "github.com/owner/repo.git", true},
		// Multiple wildcards
		{"*/*/repo", "github.com/owner/repo", true},
		// Question mark
		{"github.com/?", "github.com/a", true},
		{"github.com/?", "github.com/ab", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.uri, func(t *testing.T) {
			matches := registry.matchesPattern(tt.uri, tt.pattern)
			assert.Equal(t, tt.matches, matches)
		})
	}
}
