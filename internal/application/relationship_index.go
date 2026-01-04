package application

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// RelationshipIndex maintains a bidirectional index of element relationships
// Forward: memory_id -> [related_element_ids]
// Reverse: element_id -> [memory_ids that reference it].
type RelationshipIndex struct {
	forward map[string][]string // memory_id -> element_ids
	reverse map[string][]string // element_id -> memory_ids
	mu      sync.RWMutex
	cache   *IndexCache
}

// IndexCache provides optional caching for expensive operations.
type IndexCache struct {
	data      map[string]IndexCacheEntry
	mu        sync.RWMutex
	ttl       time.Duration
	enabled   bool
	hitCount  int64
	missCount int64
}

// IndexCacheEntry represents a cached query result.
type IndexCacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewRelationshipIndex creates a new bidirectional relationship index.
func NewRelationshipIndex() *RelationshipIndex {
	return &RelationshipIndex{
		forward: make(map[string][]string),
		reverse: make(map[string][]string),
		cache:   NewIndexCache(5 * time.Minute), // 5 min default TTL
	}
}

// NewIndexCache creates a new cache with specified TTL.
func NewIndexCache(ttl time.Duration) *IndexCache {
	return &IndexCache{
		data:    make(map[string]IndexCacheEntry),
		ttl:     ttl,
		enabled: true,
	}
}

// Add adds a relationship from a memory to related elements.
func (idx *RelationshipIndex) Add(memoryID string, relatedIDs []string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Store forward mapping
	idx.forward[memoryID] = relatedIDs

	// Update reverse mappings
	for _, elemID := range relatedIDs {
		if !contains(idx.reverse[elemID], memoryID) {
			idx.reverse[elemID] = append(idx.reverse[elemID], memoryID)
		}
	}

	// Invalidate cache for affected keys
	idx.cache.InvalidatePattern(memoryID)
	for _, elemID := range relatedIDs {
		idx.cache.InvalidatePattern(elemID)
	}
}

// Remove removes a memory from the index.
func (idx *RelationshipIndex) Remove(memoryID string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Get related elements before deletion
	relatedIDs := idx.forward[memoryID]

	// Remove forward mapping
	delete(idx.forward, memoryID)

	// Remove from reverse mappings
	for _, elemID := range relatedIDs {
		idx.reverse[elemID] = removeString(idx.reverse[elemID], memoryID)
		if len(idx.reverse[elemID]) == 0 {
			delete(idx.reverse, elemID)
		}
	}

	// Invalidate cache
	idx.cache.InvalidatePattern(memoryID)
	for _, elemID := range relatedIDs {
		idx.cache.InvalidatePattern(elemID)
	}
}

// GetRelatedElements returns elements related to a memory (forward lookup).
func (idx *RelationshipIndex) GetRelatedElements(memoryID string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if elements, ok := idx.forward[memoryID]; ok {
		return copyStrings(elements)
	}
	return nil
}

// GetRelatedMemories returns memories that reference an element (reverse lookup).
func (idx *RelationshipIndex) GetRelatedMemories(elementID string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if memories, ok := idx.reverse[elementID]; ok {
		return copyStrings(memories)
	}
	return nil
}

// Rebuild rebuilds the index from a repository.
func (idx *RelationshipIndex) Rebuild(ctx context.Context, repo domain.ElementRepository) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Clear existing index
	idx.forward = make(map[string][]string)
	idx.reverse = make(map[string][]string)
	idx.cache.Clear()

	// Get all memories
	memoryType := domain.MemoryElement
	filter := domain.ElementFilter{
		Type: &memoryType,
	}

	memories, err := repo.List(filter)
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	// Index each memory
	for _, elem := range memories {
		memory, ok := elem.(*domain.Memory)
		if !ok {
			continue
		}

		// Parse related_to metadata
		relatedStr, ok := memory.Metadata["related_to"]
		if !ok || relatedStr == "" {
			continue
		}

		relatedIDs := parseRelatedIDsFromString(relatedStr)
		if len(relatedIDs) == 0 {
			continue
		}

		memoryID := memory.GetMetadata().ID

		// Add to forward index
		idx.forward[memoryID] = relatedIDs

		// Add to reverse index
		for _, elemID := range relatedIDs {
			if !contains(idx.reverse[elemID], memoryID) {
				idx.reverse[elemID] = append(idx.reverse[elemID], memoryID)
			}
		}
	}

	return nil
}

// Stats returns index statistics.
func (idx *RelationshipIndex) Stats() IndexStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return IndexStats{
		ForwardEntries: len(idx.forward),
		ReverseEntries: len(idx.reverse),
		CacheHits:      idx.cache.hitCount,
		CacheMisses:    idx.cache.missCount,
		CacheSize:      len(idx.cache.data),
	}
}

// IndexStats contains index statistics.
type IndexStats struct {
	ForwardEntries int   // Number of memories with relationships
	ReverseEntries int   // Number of elements referenced by memories
	CacheHits      int64 // Cache hit count
	CacheMisses    int64 // Cache miss count
	CacheSize      int   // Number of cached entries
}

// Get retrieves a cached value.
func (c *IndexCache) Get(key string) (interface{}, bool) {
	if !c.enabled {
		c.missCount++
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok {
		c.missCount++
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.missCount++
		return nil, false
	}

	c.hitCount++
	return entry.Value, true
}

// Set stores a value in cache.
func (c *IndexCache) Set(key string, value interface{}) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = IndexCacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a specific key from cache.
func (c *IndexCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// InvalidatePattern removes all keys matching a pattern.
func (c *IndexCache) InvalidatePattern(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.data {
		if strings.Contains(key, pattern) {
			delete(c.data, key)
		}
	}
}

// Clear removes all cached entries.
func (c *IndexCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]IndexCacheEntry)
	c.hitCount = 0
	c.missCount = 0
}

// GetMemoriesRelatedTo retrieves all memories that reference a specific element
// This is the main function for bidirectional search.
func GetMemoriesRelatedTo(
	ctx context.Context,
	elementID string,
	repo domain.ElementRepository,
	index *RelationshipIndex,
) ([]*domain.Memory, error) {
	// Check cache first
	cacheKey := "memories_for_" + elementID
	if cached, ok := index.cache.Get(cacheKey); ok {
		return cached.([]*domain.Memory), nil
	}

	// Get memory IDs from reverse index
	memoryIDs := index.GetRelatedMemories(elementID)
	if len(memoryIDs) == 0 {
		return nil, nil
	}

	// Fetch memories from repository
	memories := make([]*domain.Memory, 0, len(memoryIDs))
	for _, memID := range memoryIDs {
		elem, err := repo.GetByID(memID)
		if err != nil {
			// Skip if memory no longer exists
			continue
		}

		memory, ok := elem.(*domain.Memory)
		if !ok {
			continue
		}

		memories = append(memories, memory)
	}

	// Cache result
	index.cache.Set(cacheKey, memories)

	return memories, nil
}

// RelationshipExpansionOptions controls recursive expansion behavior.
type RelationshipExpansionOptions struct {
	MaxDepth       int                  // Maximum recursion depth (default: 3)
	IncludeTypes   []domain.ElementType // Filter by element types
	ExcludeVisited bool                 // Prevent revisiting same elements
	FollowBothWays bool                 // Expand both forward and reverse relationships
	StopAtTypes    []domain.ElementType // Stop expansion at certain types
}

// RelationshipNode represents a node in the relationship graph.
type RelationshipNode struct {
	Element      domain.Element      // The element at this node
	Depth        int                 // Depth in the expansion tree
	Children     []*RelationshipNode // Related elements (next level)
	Relationship string              // Type of relationship (forward/reverse)
	Score        float64             // Relationship strength score
}

// ExpandRelationships performs multi-level recursive relationship expansion.
func (idx *RelationshipIndex) ExpandRelationships(
	ctx context.Context,
	rootElementID string,
	repo domain.ElementRepository,
	opts RelationshipExpansionOptions,
) (*RelationshipNode, error) {
	// Set defaults
	if opts.MaxDepth == 0 {
		opts.MaxDepth = 3
	}

	// Track visited elements to prevent cycles
	visited := make(map[string]bool)

	// Get root element
	rootElem, err := repo.GetByID(rootElementID)
	if err != nil {
		return nil, fmt.Errorf("root element not found: %w", err)
	}

	// Start recursive expansion
	return idx.expandNode(ctx, rootElem, 0, visited, repo, opts)
}

// expandNode recursively expands relationships for a single node.
func (idx *RelationshipIndex) expandNode(
	ctx context.Context,
	element domain.Element,
	currentDepth int,
	visited map[string]bool,
	repo domain.ElementRepository,
	opts RelationshipExpansionOptions,
) (*RelationshipNode, error) {
	elementID := element.GetID()

	// Check depth limit
	if currentDepth >= opts.MaxDepth {
		return &RelationshipNode{
			Element: element,
			Depth:   currentDepth,
		}, nil
	}

	// Check if already visited
	if opts.ExcludeVisited && visited[elementID] {
		return nil, nil
	}
	visited[elementID] = true

	// Check stop-at types
	for _, stopType := range opts.StopAtTypes {
		if element.GetType() == stopType {
			return &RelationshipNode{
				Element: element,
				Depth:   currentDepth,
			}, nil
		}
	}

	node := &RelationshipNode{
		Element:  element,
		Depth:    currentDepth,
		Children: []*RelationshipNode{},
	}

	// Get related element IDs
	var relatedIDs []string

	// Forward relationships (if element is a memory)
	if element.GetType() == domain.MemoryElement {
		relatedIDs = append(relatedIDs, idx.GetRelatedElements(elementID)...)
	}

	// Reverse relationships (if following both ways)
	if opts.FollowBothWays {
		relatedIDs = append(relatedIDs, idx.GetRelatedMemories(elementID)...)
	}

	// Remove duplicates
	relatedIDs = uniqueStrings(relatedIDs)

	// Expand each related element
	for _, relatedID := range relatedIDs {
		relatedElem, err := repo.GetByID(relatedID)
		if err != nil {
			continue // Skip if element doesn't exist
		}

		// Apply type filter
		if len(opts.IncludeTypes) > 0 {
			if !containsType(opts.IncludeTypes, relatedElem.GetType()) {
				continue
			}
		}

		// Recursive expansion
		childNode, err := idx.expandNode(
			ctx,
			relatedElem,
			currentDepth+1,
			visited,
			repo,
			opts,
		)
		if err != nil {
			continue
		}

		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
	}

	return node, nil
}

// GetBidirectionalRelationships returns all relationships for an element in both directions.
func (idx *RelationshipIndex) GetBidirectionalRelationships(elementID string) BidirectionalRelationships {
	return BidirectionalRelationships{
		Forward: idx.GetRelatedElements(elementID),
		Reverse: idx.GetRelatedMemories(elementID),
	}
}

// BidirectionalRelationships holds forward and reverse relationships.
type BidirectionalRelationships struct {
	Forward []string // Elements this element points to
	Reverse []string // Elements that point to this element
}

// GetAllRelatedElements returns all unique elements related to the given element (both directions).
func (idx *RelationshipIndex) GetAllRelatedElements(elementID string) []string {
	forward := idx.GetRelatedElements(elementID)
	reverse := idx.GetRelatedMemories(elementID)

	forward = append(forward, reverse...)
	return uniqueStrings(forward)
}

// Helper functions

func parseRelatedIDsFromString(relatedStr string) []string {
	if relatedStr == "" {
		return nil
	}

	parts := strings.Split(relatedStr, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id != "" {
			result = append(result, id)
		}
	}

	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeString(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func copyStrings(slice []string) []string {
	if slice == nil {
		return nil
	}
	result := make([]string, len(slice))
	copy(result, slice)
	return result
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func containsType(types []domain.ElementType, target domain.ElementType) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}
