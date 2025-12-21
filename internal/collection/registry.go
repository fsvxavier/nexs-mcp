package collection

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
)

// CachedCollection represents a cached collection.
type CachedCollection struct {
	Collection  *sources.Collection
	CachedAt    time.Time
	ExpiresAt   time.Time
	AccessCount int
	LastAccess  time.Time
}

// RegistryCache manages in-memory collection cache.
type RegistryCache struct {
	collections map[string]*CachedCollection // URI -> cached collection
	mu          sync.RWMutex
	ttl         time.Duration // Time to live
	enabled     bool
}

// NewRegistryCache creates a new registry cache.
func NewRegistryCache(ttl time.Duration) *RegistryCache {
	if ttl == 0 {
		ttl = 15 * time.Minute // Default: 15 minutes
	}
	return &RegistryCache{
		collections: make(map[string]*CachedCollection),
		ttl:         ttl,
		enabled:     true,
	}
}

// Get retrieves a cached collection if available and not expired.
func (c *RegistryCache) Get(uri string) (*sources.Collection, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.collections[uri]
	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(cached.ExpiresAt) {
		return nil, false // Expired
	}

	// Update access stats (requires write lock, so we skip for performance)
	// In production, could use atomic counters
	return cached.Collection, true
}

// Set stores a collection in cache.
func (c *RegistryCache) Set(uri string, collection *sources.Collection) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.collections[uri] = &CachedCollection{
		Collection:  collection,
		CachedAt:    now,
		ExpiresAt:   now.Add(c.ttl),
		AccessCount: 0,
		LastAccess:  now,
	}
}

// Invalidate removes a collection from cache.
func (c *RegistryCache) Invalidate(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.collections, uri)
}

// Clear removes all cached collections.
func (c *RegistryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collections = make(map[string]*CachedCollection)
}

// Stats returns cache statistics.
func (c *RegistryCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalAccess := 0
	expired := 0
	now := time.Now()

	for _, cached := range c.collections {
		totalAccess += cached.AccessCount
		if now.After(cached.ExpiresAt) {
			expired++
		}
	}

	return map[string]interface{}{
		"total_cached": len(c.collections),
		"expired":      expired,
		"total_access": totalAccess,
		"ttl_minutes":  int(c.ttl.Minutes()),
		"enabled":      c.enabled,
	}
}

// MetadataIndex indexes collection metadata for fast search.
type MetadataIndex struct {
	byAuthor   map[string][]*sources.CollectionMetadata // author -> collections
	byCategory map[string][]*sources.CollectionMetadata // category -> collections
	byTag      map[string][]*sources.CollectionMetadata // tag -> collections
	byKeyword  map[string][]*sources.CollectionMetadata // keyword -> collections
	all        []*sources.CollectionMetadata            // all collections
	mu         sync.RWMutex
}

// NewMetadataIndex creates a new metadata index.
func NewMetadataIndex() *MetadataIndex {
	return &MetadataIndex{
		byAuthor:   make(map[string][]*sources.CollectionMetadata),
		byCategory: make(map[string][]*sources.CollectionMetadata),
		byTag:      make(map[string][]*sources.CollectionMetadata),
		byKeyword:  make(map[string][]*sources.CollectionMetadata),
		all:        make([]*sources.CollectionMetadata, 0),
	}
}

// Index indexes a collection's metadata.
func (idx *MetadataIndex) Index(metadata *sources.CollectionMetadata) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Add to all
	idx.all = append(idx.all, metadata)

	// Index by author
	author := strings.ToLower(metadata.Author)
	idx.byAuthor[author] = append(idx.byAuthor[author], metadata)

	// Index by category
	if metadata.Category != "" {
		category := strings.ToLower(metadata.Category)
		idx.byCategory[category] = append(idx.byCategory[category], metadata)
	}

	// Index by tags
	for _, tag := range metadata.Tags {
		tag = strings.ToLower(tag)
		idx.byTag[tag] = append(idx.byTag[tag], metadata)
	}

	// Index by name keywords
	nameWords := strings.Fields(strings.ToLower(metadata.Name))
	for _, word := range nameWords {
		idx.byKeyword[word] = append(idx.byKeyword[word], metadata)
	}

	// Index by description keywords
	descWords := strings.Fields(strings.ToLower(metadata.Description))
	for _, word := range descWords {
		if len(word) > 3 { // Skip short words
			idx.byKeyword[word] = append(idx.byKeyword[word], metadata)
		}
	}
}

// Search performs a search across indexed metadata.
func (idx *MetadataIndex) Search(query string, filters *sources.BrowseFilter) []*sources.CollectionMetadata {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var results []*sources.CollectionMetadata

	// Start with all collections
	candidates := idx.all

	// Filter by category if specified
	if filters != nil && filters.Category != "" {
		category := strings.ToLower(filters.Category)
		if collections, exists := idx.byCategory[category]; exists {
			candidates = collections
		} else {
			return []*sources.CollectionMetadata{} // No matches
		}
	}

	// Filter by author if specified
	if filters != nil && filters.Author != "" {
		author := strings.ToLower(filters.Author)
		if collections, exists := idx.byAuthor[author]; exists {
			candidates = idx.intersect(candidates, collections)
		} else {
			return []*sources.CollectionMetadata{} // No matches
		}
	}

	// Filter by tags if specified
	if filters != nil && len(filters.Tags) > 0 {
		for _, tag := range filters.Tags {
			tag = strings.ToLower(tag)
			if collections, exists := idx.byTag[tag]; exists {
				candidates = idx.intersect(candidates, collections)
			} else {
				return []*sources.CollectionMetadata{} // No matches (AND logic)
			}
		}
	}

	// Full-text search if query specified
	if query != "" {
		queryLower := strings.ToLower(query)
		queryWords := strings.Fields(queryLower)

		for _, collection := range candidates {
			score := idx.calculateRelevance(collection, queryWords)
			if score > 0 {
				results = append(results, collection)
			}
		}
	} else {
		results = candidates
	}

	// Apply limit and offset
	if filters != nil {
		if filters.Offset > 0 && filters.Offset < len(results) {
			results = results[filters.Offset:]
		}
		if filters.Limit > 0 && filters.Limit < len(results) {
			results = results[:filters.Limit]
		}
	}

	return results
}

// calculateRelevance scores a collection against query words.
func (idx *MetadataIndex) calculateRelevance(collection *sources.CollectionMetadata, queryWords []string) int {
	score := 0

	nameLower := strings.ToLower(collection.Name)
	descLower := strings.ToLower(collection.Description)

	for _, word := range queryWords {
		// Exact name match: +10
		if strings.Contains(nameLower, word) {
			score += 10
		}
		// Description match: +3
		if strings.Contains(descLower, word) {
			score += 3
		}
		// Tag match: +5
		for _, tag := range collection.Tags {
			if strings.ToLower(tag) == word {
				score += 5
			}
		}
	}

	return score
}

// intersect returns the intersection of two metadata slices.
func (idx *MetadataIndex) intersect(a, b []*sources.CollectionMetadata) []*sources.CollectionMetadata {
	uriMap := make(map[string]bool)
	for _, item := range a {
		uriMap[item.URI] = true
	}

	result := make([]*sources.CollectionMetadata, 0)
	for _, item := range b {
		if uriMap[item.URI] {
			result = append(result, item)
		}
	}

	return result
}

// Clear clears the index.
func (idx *MetadataIndex) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.byAuthor = make(map[string][]*sources.CollectionMetadata)
	idx.byCategory = make(map[string][]*sources.CollectionMetadata)
	idx.byTag = make(map[string][]*sources.CollectionMetadata)
	idx.byKeyword = make(map[string][]*sources.CollectionMetadata)
	idx.all = make([]*sources.CollectionMetadata, 0)
}

// Stats returns index statistics.
func (idx *MetadataIndex) Stats() map[string]interface{} {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return map[string]interface{}{
		"total_collections": len(idx.all),
		"authors":           len(idx.byAuthor),
		"categories":        len(idx.byCategory),
		"tags":              len(idx.byTag),
		"keywords":          len(idx.byKeyword),
	}
}

// DependencyNode represents a node in the dependency graph.
type DependencyNode struct {
	URI          string
	Name         string
	Version      string
	Dependencies []*DependencyNode
	Dependents   []*DependencyNode
	Depth        int
	Visited      bool
}

// DependencyGraph manages collection dependencies.
type DependencyGraph struct {
	nodes map[string]*DependencyNode
	mu    sync.RWMutex
}

// NewDependencyGraph creates a new dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*DependencyNode),
	}
}

// AddNode adds a node to the graph.
func (g *DependencyGraph) AddNode(uri string, name string, version string) *DependencyNode {
	g.mu.Lock()
	defer g.mu.Unlock()

	if node, exists := g.nodes[uri]; exists {
		return node
	}

	node := &DependencyNode{
		URI:          uri,
		Name:         name,
		Version:      version,
		Dependencies: make([]*DependencyNode, 0),
		Dependents:   make([]*DependencyNode, 0),
		Depth:        0,
	}

	g.nodes[uri] = node
	return node
}

// AddDependency adds a dependency relationship.
func (g *DependencyGraph) AddDependency(fromURI string, toURI string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	from, fromExists := g.nodes[fromURI]
	to, toExists := g.nodes[toURI]

	if !fromExists {
		return fmt.Errorf("source node not found: %s", fromURI)
	}
	if !toExists {
		return fmt.Errorf("target node not found: %s", toURI)
	}

	// Check for circular dependency
	if g.hasPath(to, from) {
		return fmt.Errorf("circular dependency detected: %s -> %s", fromURI, toURI)
	}

	from.Dependencies = append(from.Dependencies, to)
	to.Dependents = append(to.Dependents, from)

	return nil
}

// hasPath checks if there's a path from 'from' to 'to' (for cycle detection).
func (g *DependencyGraph) hasPath(from *DependencyNode, to *DependencyNode) bool {
	if from.URI == to.URI {
		return true
	}

	for _, dep := range from.Dependencies {
		if g.hasPath(dep, to) {
			return true
		}
	}

	return false
}

// TopologicalSort returns nodes in dependency order (dependencies first).
func (g *DependencyGraph) TopologicalSort() ([]*DependencyNode, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*DependencyNode, 0, len(g.nodes))
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(*DependencyNode) error
	visit = func(node *DependencyNode) error {
		if visited[node.URI] {
			return nil
		}

		if visiting[node.URI] {
			return fmt.Errorf("circular dependency detected at: %s", node.URI)
		}

		visiting[node.URI] = true

		for _, dep := range node.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[node.URI] = false
		visited[node.URI] = true
		result = append(result, node)

		return nil
	}

	for _, node := range g.nodes {
		if err := visit(node); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Registry coordinates multiple collection sources for discovery and retrieval.
// It aggregates results from GitHub, local filesystem, and other configured sources.
type Registry struct {
	sources         []sources.CollectionSource
	cache           *RegistryCache
	metadataIndex   *MetadataIndex
	dependencyGraph *DependencyGraph
	mu              sync.RWMutex
}

// NewRegistry creates a new collection registry.
func NewRegistry() *Registry {
	return &Registry{
		sources:         make([]sources.CollectionSource, 0),
		cache:           NewRegistryCache(15 * time.Minute),
		metadataIndex:   NewMetadataIndex(),
		dependencyGraph: NewDependencyGraph(),
	}
}

// NewRegistryWithTTL creates a registry with custom cache TTL.
func NewRegistryWithTTL(cacheTTL time.Duration) *Registry {
	return &Registry{
		sources:         make([]sources.CollectionSource, 0),
		cache:           NewRegistryCache(cacheTTL),
		metadataIndex:   NewMetadataIndex(),
		dependencyGraph: NewDependencyGraph(),
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
// - https://example.com/collection.tar.gz -> HTTP source.
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

// GetCached retrieves a collection with caching support.
func (r *Registry) GetCached(ctx context.Context, uri string) (*sources.Collection, error) {
	// Try cache first
	if cached, hit := r.cache.Get(uri); hit {
		return cached, nil
	}

	// Cache miss - fetch from source
	collection, err := r.Get(ctx, uri)
	if err != nil {
		return nil, err
	}

	// Store in cache
	r.cache.Set(uri, collection)

	return collection, nil
}

// InvalidateCache invalidates a specific URI or all cache.
func (r *Registry) InvalidateCache(uri string) {
	if uri == "" {
		r.cache.Clear()
	} else {
		r.cache.Invalidate(uri)
	}
}

// Search performs an indexed search across collections.
func (r *Registry) Search(query string, filters *sources.BrowseFilter) []*sources.CollectionMetadata {
	return r.metadataIndex.Search(query, filters)
}

// IndexMetadata indexes collection metadata for search.
func (r *Registry) IndexMetadata(metadata *sources.CollectionMetadata) {
	r.metadataIndex.Index(metadata)
}

// RebuildIndex rebuilds the metadata index from all sources.
func (r *Registry) RebuildIndex(ctx context.Context) error {
	r.metadataIndex.Clear()

	// Browse all sources and index metadata
	metadata, err := r.Browse(ctx, nil, "")
	if err != nil {
		return fmt.Errorf("failed to browse for indexing: %w", err)
	}

	for _, meta := range metadata {
		r.metadataIndex.Index(meta)
	}

	return nil
}

// GetStats returns registry statistics.
func (r *Registry) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"sources":        len(r.sources),
		"cache":          r.cache.Stats(),
		"metadata_index": r.metadataIndex.Stats(),
		"dependency_graph": map[string]int{
			"nodes": len(r.dependencyGraph.nodes),
		},
	}
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

// GetCache returns the registry cache (for testing).
func (r *Registry) GetCache() *RegistryCache {
	return r.cache
}

// GetMetadataIndex returns the metadata index (for testing).
func (r *Registry) GetMetadataIndex() *MetadataIndex {
	return r.metadataIndex
}

// GetDependencyGraph returns the dependency graph (for testing).
func (r *Registry) GetDependencyGraph() *DependencyGraph {
	return r.dependencyGraph
}
