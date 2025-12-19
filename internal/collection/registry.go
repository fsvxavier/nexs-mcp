package collection

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
)

// Registry coordinates multiple collection sources for discovery and retrieval.
// It aggregates results from GitHub, local filesystem, and other configured sources.
type Registry struct {
	sources []sources.CollectionSource
	mu      sync.RWMutex
}

// NewRegistry creates a new collection registry.
func NewRegistry() *Registry {
	return &Registry{
		sources: make([]sources.CollectionSource, 0),
	}
}

// AddSource registers a new collection source.
func (r *Registry) AddSource(source sources.CollectionSource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources = append(r.sources, source)
}

// GetSources returns all registered sources.
func (r *Registry) GetSources() []sources.CollectionSource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Return a copy to prevent external modification
	sourcesCopy := make([]sources.CollectionSource, len(r.sources))
	copy(sourcesCopy, r.sources)
	return sourcesCopy
}

// Browse discovers collections from all sources or a specific source.
// Results are aggregated and deduplicated by collection ID (author/name).
func (r *Registry) Browse(ctx context.Context, filter *sources.BrowseFilter, sourceName string) ([]*sources.CollectionMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.sources) == 0 {
		return []*sources.CollectionMetadata{}, nil
	}

	// Filter sources if sourceName is specified
	sourcesToQuery := r.sources
	if sourceName != "" {
		sourcesToQuery = make([]sources.CollectionSource, 0)
		for _, src := range r.sources {
			if src.Name() == sourceName {
				sourcesToQuery = append(sourcesToQuery, src)
			}
		}
		if len(sourcesToQuery) == 0 {
			return nil, fmt.Errorf("source not found: %s", sourceName)
		}
	}

	// Query all sources in parallel
	type result struct {
		metadata []*sources.CollectionMetadata
		err      error
	}

	results := make(chan result, len(sourcesToQuery))
	var wg sync.WaitGroup

	for _, src := range sourcesToQuery {
		wg.Add(1)
		go func(source sources.CollectionSource) {
			defer wg.Done()
			metadata, err := source.Browse(ctx, filter)
			results <- result{metadata: metadata, err: err}
		}(src)
	}

	// Wait for all queries to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregate results
	allMetadata := make([]*sources.CollectionMetadata, 0)
	var errs []error

	for res := range results {
		if res.err != nil {
			errs = append(errs, res.err)
			continue
		}
		allMetadata = append(allMetadata, res.metadata...)
	}

	// If all sources failed, return error
	if len(errs) > 0 && len(allMetadata) == 0 {
		return nil, fmt.Errorf("all sources failed: %v", errs)
	}

	// Deduplicate by collection ID (author/name@version)
	seen := make(map[string]bool)
	deduplicated := make([]*sources.CollectionMetadata, 0, len(allMetadata))

	for _, meta := range allMetadata {
		id := fmt.Sprintf("%s/%s@%s", meta.Author, meta.Name, meta.Version)
		if !seen[id] {
			seen[id] = true
			deduplicated = append(deduplicated, meta)
		}
	}

	return deduplicated, nil
}

// Get retrieves a specific collection by URI.
// The URI format determines which source will handle the request:
// - github://owner/repo[@version] -> GitHub source
// - file:///path/to/collection -> Local source
// - https://example.com/collection.tar.gz -> HTTP source
func (r *Registry) Get(ctx context.Context, uri string) (*sources.Collection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find a source that supports this URI
	for _, src := range r.sources {
		if src.Supports(uri) {
			return src.Get(ctx, uri)
		}
	}

	return nil, fmt.Errorf("no source supports URI: %s", uri)
}

// FindSource returns the source that handles a given URI.
func (r *Registry) FindSource(uri string) sources.CollectionSource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, src := range r.sources {
		if src.Supports(uri) {
			return src
		}
	}
	return nil
}
