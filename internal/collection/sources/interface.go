package sources

import (
	"context"
)

// CollectionSource represents a source for discovering and retrieving collections.
// Implementations can support GitHub repositories, local filesystems, HTTP registries, etc.
type CollectionSource interface {
	// Name returns the unique name of this source (e.g., "github", "local", "http")
	Name() string

	// Browse discovers collections from this source with optional filters.
	// Returns a list of collection metadata (not full manifests).
	Browse(ctx context.Context, filter *BrowseFilter) ([]*CollectionMetadata, error)

	// Get retrieves a specific collection by URI.
	// The URI format is source-specific (e.g., "github://owner/repo", "file:///path/to/collection")
	Get(ctx context.Context, uri string) (*Collection, error)

	// Supports returns true if this source can handle the given URI.
	Supports(uri string) bool
}

// BrowseFilter defines optional filters for browsing collections.
type BrowseFilter struct {
	// Category filters by collection category (e.g., "devops", "creative-writing")
	Category string

	// Tags filters collections that have ALL specified tags
	Tags []string

	// Author filters by collection author
	Author string

	// Query performs a text search across name, description, and keywords
	Query string

	// Limit restricts the number of results (0 = no limit)
	Limit int

	// Offset skips the first N results (for pagination)
	Offset int
}

// CollectionMetadata contains lightweight metadata about a collection (for browsing).
// Full manifest is loaded only when needed via Get().
type CollectionMetadata struct {
	// Source information
	SourceName string `json:"source_name"` // Name of the source (e.g., "github", "local")
	URI        string `json:"uri"`         // Full URI to retrieve this collection

	// Basic manifest fields (subset for performance)
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Category    string   `json:"category,omitempty"`

	// Discovery metadata
	Repository string `json:"repository,omitempty"` // Link to source repository/directory
	Stars      int    `json:"stars,omitempty"`      // GitHub stars or popularity metric
	Downloads  int    `json:"downloads,omitempty"`  // Download count if available
}

// Collection represents a fully loaded collection with manifest and metadata.
// The Manifest field will be populated with *collection.Manifest at runtime.
type Collection struct {
	// Metadata
	Metadata *CollectionMetadata `json:"metadata"`

	// Full manifest (pointer to avoid import cycle, will be *collection.Manifest)
	Manifest interface{} `json:"manifest"`

	// Source-specific data (e.g., Git commit SHA, local path)
	SourceData map[string]interface{} `json:"source_data,omitempty"`
}

// InstallationState represents the state of an installed collection.
type InstallationState struct {
	// Collection identity
	ID      string `json:"id"`      // author/name
	Version string `json:"version"` // Installed version
	URI     string `json:"uri"`     // URI used for installation
	Source  string `json:"source"`  // Source name (github, local, http)

	// Installation metadata
	InstalledAt     string `json:"installed_at"`     // ISO 8601 timestamp
	InstalledBy     string `json:"installed_by"`     // Username
	InstallLocation string `json:"install_location"` // Path where collection is installed

	// Update tracking
	LatestVersion string `json:"latest_version,omitempty"`  // Latest available version (if known)
	UpdateCheckAt string `json:"update_check_at,omitempty"` // Last update check timestamp

	// Configuration
	AutoUpdate bool `json:"auto_update"` // Whether to auto-update this collection

	// Statistics
	ElementCount int                    `json:"element_count"`    // Number of elements installed
	Stats        map[string]int         `json:"stats,omitempty"`  // Element counts by type
	Custom       map[string]interface{} `json:"custom,omitempty"` // Custom installation data
}
