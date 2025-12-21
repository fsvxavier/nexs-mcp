package security

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// TrustedSource represents a trusted collection source.
type TrustedSource struct {
	Name      string   // Source name (e.g., "official-collections")
	Pattern   string   // URI pattern (e.g., "github.com/nexs-mcp/*")
	PublicKey string   // GPG/SSH public key (optional)
	Required  bool     // If true, unsigned collections are rejected
	Verified  bool     // If true, this source has been verified by maintainers
	Tags      []string // Tags for categorization
}

// TrustedSourceRegistry manages trusted collection sources.
type TrustedSourceRegistry struct {
	sources map[string]*TrustedSource // name -> source
}

// NewTrustedSourceRegistry creates a new trusted source registry.
func NewTrustedSourceRegistry() *TrustedSourceRegistry {
	registry := &TrustedSourceRegistry{
		sources: make(map[string]*TrustedSource),
	}

	// Add default trusted sources
	registry.AddDefaultSources()

	return registry
}

// AddSource adds a trusted source.
func (r *TrustedSourceRegistry) AddSource(source *TrustedSource) error {
	if source.Name == "" {
		return errors.New("source name is required")
	}
	if source.Pattern == "" {
		return errors.New("source pattern is required")
	}

	r.sources[source.Name] = source
	return nil
}

// RemoveSource removes a trusted source.
func (r *TrustedSourceRegistry) RemoveSource(name string) {
	delete(r.sources, name)
}

// GetSource retrieves a trusted source by name.
func (r *TrustedSourceRegistry) GetSource(name string) (*TrustedSource, bool) {
	source, exists := r.sources[name]
	return source, exists
}

// ListSources returns all trusted sources.
func (r *TrustedSourceRegistry) ListSources() []*TrustedSource {
	sources := make([]*TrustedSource, 0, len(r.sources))
	for _, source := range r.sources {
		sources = append(sources, source)
	}
	return sources
}

// IsTrusted checks if a URI is from a trusted source.
func (r *TrustedSourceRegistry) IsTrusted(uri string) (*TrustedSource, bool) {
	for _, source := range r.sources {
		if r.matchesPattern(uri, source.Pattern) {
			return source, true
		}
	}
	return nil, false
}

// matchesPattern checks if a URI matches a pattern.
func (r *TrustedSourceRegistry) matchesPattern(uri string, pattern string) bool {
	// Convert glob-style pattern to regex
	// * -> .*
	// ? -> .
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "?", ".")
	regexPattern = "^" + regexPattern + "$"

	matched, err := regexp.MatchString(regexPattern, uri)
	if err != nil {
		return false
	}

	return matched
}

// ValidateURI validates a URI against trusted sources.
func (r *TrustedSourceRegistry) ValidateURI(uri string, requireTrusted bool) error {
	_, trusted := r.IsTrusted(uri)

	if requireTrusted && !trusted {
		return fmt.Errorf("URI is not from a trusted source: %s", uri)
	}

	// Note: If source requires signatures, that check happens elsewhere
	// This validation only checks the URI pattern

	return nil
}

// AddDefaultSources adds the default trusted sources.
func (r *TrustedSourceRegistry) AddDefaultSources() {
	// Official NEXS-MCP collections
	_ = r.AddSource(&TrustedSource{
		Name:     "nexs-official",
		Pattern:  "github.com/fsvxavier/nexs-mcp-collections/*",
		Required: false,
		Verified: true,
		Tags:     []string{"official", "verified"},
	})

	// Official NEXS-MCP organization
	_ = r.AddSource(&TrustedSource{
		Name:     "nexs-org",
		Pattern:  "github.com/nexs-mcp/*",
		Required: false,
		Verified: true,
		Tags:     []string{"official", "verified"},
	})

	// Community verified collections
	_ = r.AddSource(&TrustedSource{
		Name:     "community-verified",
		Pattern:  "github.com/*/*-verified-collection",
		Required: false,
		Verified: true,
		Tags:     []string{"community", "verified"},
	})

	// Allow local filesystem (for development)
	_ = r.AddSource(&TrustedSource{
		Name:     "local-filesystem",
		Pattern:  "file:///*",
		Required: false,
		Verified: false,
		Tags:     []string{"local", "development"},
	})
}

// SecurityConfig holds security configuration.
type SecurityConfig struct {
	RequireSignatures    bool     // Require all collections to be signed
	RequireTrustedSource bool     // Only allow collections from trusted sources
	AllowUnsigned        bool     // Allow unsigned collections (overrides RequireSignatures)
	ScanEnabled          bool     // Enable code scanning
	ScanThreshold        Severity // Minimum severity to reject (e.g., "critical", "high")
	TrustedSources       []string // Additional trusted source patterns
}

// NewSecurityConfig creates a default security configuration.
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RequireSignatures:    false,            // Default: optional signatures
		RequireTrustedSource: false,            // Default: allow any source
		AllowUnsigned:        true,             // Default: allow unsigned
		ScanEnabled:          true,             // Default: scan enabled
		ScanThreshold:        SeverityCritical, // Default: reject critical only
		TrustedSources:       []string{},
	}
}

// Validate validates the security configuration.
func (c *SecurityConfig) Validate() error {
	if c.RequireSignatures && c.AllowUnsigned {
		return errors.New("conflicting config: RequireSignatures and AllowUnsigned both enabled")
	}

	validThresholds := map[Severity]bool{
		SeverityLow:      true,
		SeverityMedium:   true,
		SeverityHigh:     true,
		SeverityCritical: true,
	}

	if !validThresholds[c.ScanThreshold] {
		return fmt.Errorf("invalid scan threshold: %s (must be: low, medium, high, critical)", c.ScanThreshold)
	}

	return nil
}

// ShouldRequireSignature determines if a signature is required for a URI.
func (c *SecurityConfig) ShouldRequireSignature(uri string, source *TrustedSource) bool {
	// If signatures explicitly allowed to be missing, return false
	if c.AllowUnsigned {
		return false
	}

	// If signatures required globally
	if c.RequireSignatures {
		return true
	}

	// If source requires signature
	if source != nil && source.Required {
		return true
	}

	return false
}
