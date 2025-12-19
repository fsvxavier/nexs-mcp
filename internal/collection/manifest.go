// Package collection provides functionality for managing collections of NEXS MCP elements.
// Collections are bundles of related elements (personas, skills, templates, etc.) that can be
// discovered, installed, and shared across different sources (GitHub, local filesystem, HTTP).
package collection

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Manifest represents the collection.yaml file structure.
// It defines metadata, dependencies, elements, and configuration for a collection.
type Manifest struct {
	// Required fields
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Author      string `yaml:"author" json:"author"`
	Description string `yaml:"description" json:"description"`

	// Recommended fields
	Tags     []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Category string   `yaml:"category,omitempty" json:"category,omitempty"`
	License  string   `yaml:"license,omitempty" json:"license,omitempty"`

	// Version requirements
	MinNEXSVersion string `yaml:"min_nexs_version,omitempty" json:"min_nexs_version,omitempty"`

	// Documentation and links
	Homepage      string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	Documentation string `yaml:"documentation,omitempty" json:"documentation,omitempty"`
	Repository    string `yaml:"repository,omitempty" json:"repository,omitempty"`

	// Maintainers
	Maintainers []Maintainer `yaml:"maintainers,omitempty" json:"maintainers,omitempty"`

	// Dependencies on other collections
	Dependencies []Dependency `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`

	// Elements included in this collection
	Elements []Element `yaml:"elements" json:"elements"`

	// Configuration options
	Config *Config `yaml:"config,omitempty" json:"config,omitempty"`

	// Installation hooks
	Hooks *Hooks `yaml:"hooks,omitempty" json:"hooks,omitempty"`

	// Statistics (auto-generated, optional in source manifest)
	Stats *Stats `yaml:"stats,omitempty" json:"stats,omitempty"`

	// Changelog entries
	Changelog []ChangelogEntry `yaml:"changelog,omitempty" json:"changelog,omitempty"`

	// Keywords for search/discovery
	Keywords []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`

	// Media (screenshots, videos)
	Media []Media `yaml:"media,omitempty" json:"media,omitempty"`
}

// Maintainer represents a collection maintainer.
type Maintainer struct {
	Name   string `yaml:"name" json:"name"`
	Email  string `yaml:"email,omitempty" json:"email,omitempty"`
	GitHub string `yaml:"github,omitempty" json:"github,omitempty"`
}

// Dependency represents a dependency on another collection.
type Dependency struct {
	URI         string `yaml:"uri" json:"uri"`                                     // e.g., "github://owner/repo@^1.0.0"
	Description string `yaml:"description,omitempty" json:"description,omitempty"` // Human-readable description
	Optional    bool   `yaml:"optional,omitempty" json:"optional,omitempty"`       // If true, installation continues even if dependency fails
}

// Element represents a single element (persona, skill, template, etc.) in the collection.
type Element struct {
	Path        string `yaml:"path" json:"path"`                                   // File path or glob pattern
	Type        string `yaml:"type,omitempty" json:"type,omitempty"`               // persona, skill, template, agent, memory, ensemble
	Description string `yaml:"description,omitempty" json:"description,omitempty"` // Human-readable description
}

// Config holds configuration options for the collection.
type Config struct {
	DefaultPersona      string            `yaml:"default_persona,omitempty" json:"default_persona,omitempty"`
	AutoActivateSkills  []string          `yaml:"auto_activate_skills,omitempty" json:"auto_activate_skills,omitempty"`
	DefaultPrivacyLevel string            `yaml:"default_privacy_level,omitempty" json:"default_privacy_level,omitempty"` // public, private, shared
	Custom              map[string]string `yaml:"custom,omitempty" json:"custom,omitempty"`                               // Custom key-value settings
}

// Hooks defines lifecycle hooks for collection installation/updates.
type Hooks struct {
	PreInstall   []Hook `yaml:"pre_install,omitempty" json:"pre_install,omitempty"`
	PostInstall  []Hook `yaml:"post_install,omitempty" json:"post_install,omitempty"`
	PreUpdate    []Hook `yaml:"pre_update,omitempty" json:"pre_update,omitempty"`
	PostUpdate   []Hook `yaml:"post_update,omitempty" json:"post_update,omitempty"`
	PreUninstall []Hook `yaml:"pre_uninstall,omitempty" json:"pre_uninstall,omitempty"`
}

// Hook represents a single hook operation.
type Hook struct {
	Type        string            `yaml:"type" json:"type"`                                   // command, validate, backup, confirm
	Command     string            `yaml:"command,omitempty" json:"command,omitempty"`         // Shell command to run
	Description string            `yaml:"description,omitempty" json:"description,omitempty"` // Human-readable description
	Message     string            `yaml:"message,omitempty" json:"message,omitempty"`         // For confirm type
	Checks      []ToolCheck       `yaml:"checks,omitempty" json:"checks,omitempty"`           // For validate type
	Custom      map[string]string `yaml:"custom,omitempty" json:"custom,omitempty"`           // Custom hook data
}

// ToolCheck represents a tool availability check.
type ToolCheck struct {
	Tool     string `yaml:"tool" json:"tool"`
	Optional bool   `yaml:"optional,omitempty" json:"optional,omitempty"`
}

// Stats holds collection statistics (auto-generated).
type Stats struct {
	TotalElements int       `yaml:"total_elements" json:"total_elements"`
	Personas      int       `yaml:"personas" json:"personas"`
	Skills        int       `yaml:"skills" json:"skills"`
	Templates     int       `yaml:"templates" json:"templates"`
	Agents        int       `yaml:"agents" json:"agents"`
	Memories      int       `yaml:"memories" json:"memories"`
	Ensembles     int       `yaml:"ensembles" json:"ensembles"`
	LastUpdated   time.Time `yaml:"last_updated,omitempty" json:"last_updated,omitempty"`
}

// ChangelogEntry represents a single version changelog entry.
type ChangelogEntry struct {
	Version string   `yaml:"version" json:"version"`
	Date    string   `yaml:"date" json:"date"` // YYYY-MM-DD
	Changes []string `yaml:"changes" json:"changes"`
}

// Media represents a screenshot, video, or other media asset.
type Media struct {
	Type        string `yaml:"type" json:"type"`                                   // screenshot, video
	URL         string `yaml:"url" json:"url"`                                     // URL to media
	Description string `yaml:"description,omitempty" json:"description,omitempty"` // Human-readable description
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
		return fmt.Errorf("manifest validation failed: 'name' is required")
	}
	if m.Version == "" {
		return fmt.Errorf("manifest validation failed: 'version' is required")
	}
	if m.Author == "" {
		return fmt.Errorf("manifest validation failed: 'author' is required")
	}
	if m.Description == "" {
		return fmt.Errorf("manifest validation failed: 'description' is required")
	}
	if len(m.Elements) == 0 {
		return fmt.Errorf("manifest validation failed: 'elements' must contain at least one element")
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
