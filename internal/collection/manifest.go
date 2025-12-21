// Package collection provides functionality for managing collections of NEXS MCP elements.
// Collections are bundles of related elements (personas, skills, templates, etc.) that can be
// discovered, installed, and shared across different sources (GitHub, local filesystem, HTTP).
package collection

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Manifest represents the collection.yaml file structure.
// It defines metadata, dependencies, elements, and configuration for a collection.
type Manifest struct {
	// Required fields
	Name        string `json:"name"        yaml:"name"`
	Version     string `json:"version"     yaml:"version"`
	Author      string `json:"author"      yaml:"author"`
	Description string `json:"description" yaml:"description"`

	// Recommended fields
	Tags     []string `json:"tags,omitempty"     yaml:"tags,omitempty"`
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	License  string   `json:"license,omitempty"  yaml:"license,omitempty"`

	// Version requirements
	MinNEXSVersion string `json:"min_nexs_version,omitempty" yaml:"min_nexs_version,omitempty"`

	// Documentation and links
	Homepage      string `json:"homepage,omitempty"      yaml:"homepage,omitempty"`
	Documentation string `json:"documentation,omitempty" yaml:"documentation,omitempty"`
	Repository    string `json:"repository,omitempty"    yaml:"repository,omitempty"`

	// Maintainers
	Maintainers []Maintainer `json:"maintainers,omitempty" yaml:"maintainers,omitempty"`

	// Dependencies on other collections
	Dependencies []Dependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`

	// Elements included in this collection
	Elements []Element `json:"elements" yaml:"elements"`

	// Configuration options
	Config *Config `json:"config,omitempty" yaml:"config,omitempty"`

	// Installation hooks
	Hooks *Hooks `json:"hooks,omitempty" yaml:"hooks,omitempty"`

	// Statistics (auto-generated, optional in source manifest)
	Stats *Stats `json:"stats,omitempty" yaml:"stats,omitempty"`

	// Changelog entries
	Changelog []ChangelogEntry `json:"changelog,omitempty" yaml:"changelog,omitempty"`

	// Keywords for search/discovery
	Keywords []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`

	// Media (screenshots, videos)
	Media []Media `json:"media,omitempty" yaml:"media,omitempty"`
}

// Maintainer represents a collection maintainer.
type Maintainer struct {
	Name   string `json:"name"             yaml:"name"`
	Email  string `json:"email,omitempty"  yaml:"email,omitempty"`
	GitHub string `json:"github,omitempty" yaml:"github,omitempty"`
}

// Dependency represents a dependency on another collection.
type Dependency struct {
	URI         string `json:"uri"                   yaml:"uri"`                   // e.g., "github://owner/repo@^1.0.0"
	Description string `json:"description,omitempty" yaml:"description,omitempty"` // Human-readable description
	Optional    bool   `json:"optional,omitempty"    yaml:"optional,omitempty"`    // If true, installation continues even if dependency fails
	Version     string `json:"version,omitempty"     yaml:"version,omitempty"`     // Version constraint, e.g., "^1.0.0"
}

// Element represents a single element (persona, skill, template, etc.) in the collection.
type Element struct {
	Path        string `json:"path"                  yaml:"path"`                  // File path or glob pattern
	Type        string `json:"type,omitempty"        yaml:"type,omitempty"`        // persona, skill, template, agent, memory, ensemble
	Description string `json:"description,omitempty" yaml:"description,omitempty"` // Human-readable description
}

// Config holds configuration options for the collection.
type Config struct {
	DefaultPersona      string            `json:"default_persona,omitempty"       yaml:"default_persona,omitempty"`
	AutoActivateSkills  []string          `json:"auto_activate_skills,omitempty"  yaml:"auto_activate_skills,omitempty"`
	DefaultPrivacyLevel string            `json:"default_privacy_level,omitempty" yaml:"default_privacy_level,omitempty"` // public, private, shared
	Custom              map[string]string `json:"custom,omitempty"                yaml:"custom,omitempty"`                // Custom key-value settings
}

// Hooks defines lifecycle hooks for collection installation/updates.
type Hooks struct {
	PreInstall   []Hook `json:"pre_install,omitempty"   yaml:"pre_install,omitempty"`
	PostInstall  []Hook `json:"post_install,omitempty"  yaml:"post_install,omitempty"`
	PreUpdate    []Hook `json:"pre_update,omitempty"    yaml:"pre_update,omitempty"`
	PostUpdate   []Hook `json:"post_update,omitempty"   yaml:"post_update,omitempty"`
	PreUninstall []Hook `json:"pre_uninstall,omitempty" yaml:"pre_uninstall,omitempty"`
}

// Hook represents a single hook operation.
type Hook struct {
	Type        string            `json:"type"                  yaml:"type"`                  // command, validate, backup, confirm
	Command     string            `json:"command,omitempty"     yaml:"command,omitempty"`     // Shell command to run
	Description string            `json:"description,omitempty" yaml:"description,omitempty"` // Human-readable description
	Message     string            `json:"message,omitempty"     yaml:"message,omitempty"`     // For confirm type
	Checks      []ToolCheck       `json:"checks,omitempty"      yaml:"checks,omitempty"`      // For validate type
	Custom      map[string]string `json:"custom,omitempty"      yaml:"custom,omitempty"`      // Custom hook data
}

// ToolCheck represents a tool availability check.
type ToolCheck struct {
	Tool     string `json:"tool"               yaml:"tool"`
	Optional bool   `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// Stats holds collection statistics (auto-generated).
type Stats struct {
	TotalElements int       `json:"total_elements"         yaml:"total_elements"`
	Personas      int       `json:"personas"               yaml:"personas"`
	Skills        int       `json:"skills"                 yaml:"skills"`
	Templates     int       `json:"templates"              yaml:"templates"`
	Agents        int       `json:"agents"                 yaml:"agents"`
	Memories      int       `json:"memories"               yaml:"memories"`
	Ensembles     int       `json:"ensembles"              yaml:"ensembles"`
	LastUpdated   time.Time `json:"last_updated,omitempty" yaml:"last_updated,omitempty"`
}

// ChangelogEntry represents a single version changelog entry.
type ChangelogEntry struct {
	Version string   `json:"version" yaml:"version"`
	Date    string   `json:"date"    yaml:"date"` // YYYY-MM-DD
	Changes []string `json:"changes" yaml:"changes"`
}

// Media represents a screenshot, video, or other media asset.
type Media struct {
	Type        string `json:"type"                  yaml:"type"`                  // screenshot, video
	URL         string `json:"url"                   yaml:"url"`                   // URL to media
	Description string `json:"description,omitempty" yaml:"description,omitempty"` // Human-readable description
}

// ParseManifest parses a collection manifest from YAML bytes.
func ParseManifest(data []byte) (*Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	return &manifest, nil
}

// Validate validates the manifest structure and required fields.
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return errors.New("manifest validation failed: 'name' is required")
	}
	if m.Version == "" {
		return errors.New("manifest validation failed: 'version' is required")
	}
	if m.Author == "" {
		return errors.New("manifest validation failed: 'author' is required")
	}
	if m.Description == "" {
		return errors.New("manifest validation failed: 'description' is required")
	}
	if len(m.Elements) == 0 {
		return errors.New("manifest validation failed: 'elements' must contain at least one element")
	}

	// Validate element paths are not empty
	for i, elem := range m.Elements {
		if elem.Path == "" {
			return fmt.Errorf("manifest validation failed: element[%d].path is required", i)
		}
	}

	// Validate dependency URIs if present
	for i, dep := range m.Dependencies {
		if dep.URI == "" {
			return fmt.Errorf("manifest validation failed: dependency[%d].uri is required", i)
		}
	}

	return nil
}

// ToYAML serializes the manifest to YAML bytes.
func (m *Manifest) ToYAML() ([]byte, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize manifest to YAML: %w", err)
	}
	return data, nil
}

// ID returns a unique identifier for the collection (author/name).
func (m *Manifest) ID() string {
	return fmt.Sprintf("%s/%s", m.Author, m.Name)
}

// FullID returns a unique identifier including version (author/name@version).
func (m *Manifest) FullID() string {
	return fmt.Sprintf("%s/%s@%s", m.Author, m.Name, m.Version)
}
