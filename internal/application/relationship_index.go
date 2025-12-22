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
// Reverse: element_id -> [memory_ids that reference it]
type RelationshipIndex struct {
	forward map[string][]string // memory_id -> element_ids
	reverse map[string][]string // element_id -> memory_ids
	mu      sync.RWMutex
	cache   *IndexCache
}

// IndexCache provides optional caching for expensive operations
type IndexCache struct {
	data      map[string]CacheEntry
	mu        sync.RWMutex
	ttl       time.Duration
	enabled   bool
	hitCount  int64
	missCount int64
}

// CacheEntry represents a cached query result
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewRelationshipIndex creates a new bidirectional relationship index
func NewRelationshipIndex() *RelationshipIndex {
	return &RelationshipIndex{
		forward: make(map[string][]string),
		reverse: make(map[string][]string),
		cache:   NewIndexCache(5 * time.Minute), // 5 min default TTL
	}
}

// NewIndexCache creates a new cache with specified TTL
func NewIndexCache(ttl time.Duration) *IndexCache {
	return &IndexCache{
		data:    make(map[string]CacheEntry),
		ttl:     ttl,
		enabled: true,
	}
}

// Add adds a relationship from a memory to related elements
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

// Remove removes a memory from the index
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

// GetRelatedElements returns elements related to a memory (forward lookup)
func (idx *RelationshipIndex) GetRelatedElements(memoryID string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if elements, ok := idx.forward[memoryID]; ok {
		return copyStrings(elements)
	}
	return nil
}

// GetRelatedMemories returns memories that reference an element (reverse lookup)
func (idx *RelationshipIndex) GetRelatedMemories(elementID string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if memories, ok := idx.reverse[elementID]; ok {
		return copyStrings(memories)
	}
	return nil
}

// Rebuild rebuilds the index from a repository
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

// Stats returns index statistics
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

// IndexStats contains index statistics
type IndexStats struct {
	ForwardEntries int   // Number of memories with relationships
	ReverseEntries int   // Number of elements referenced by memories
	CacheHits      int64 // Cache hit count
	CacheMisses    int64 // Cache miss count
	CacheSize      int   // Number of cached entries
}

// Get retrieves a cached value
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

// Set stores a value in cache
func (c *IndexCache) Set(key string, value interface{}) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a specific key from cache
func (c *IndexCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// InvalidatePattern removes all keys matching a pattern
func (c *IndexCache) InvalidatePattern(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.data {
		if strings.Contains(key, pattern) {
			delete(c.data, key)
		}
	}
}

// Clear removes all cached entries
func (c *IndexCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]CacheEntry)
	c.hitCount = 0
	c.missCount = 0
}

// GetMemoriesRelatedTo retrieves all memories that reference a specific element
// This is the main function for bidirectional search
func GetMemoriesRelatedTo(
	ctx context.Context,
	elementID string,
	repo domain.ElementRepository,
	index *RelationshipIndex,
) ([]*domain.Memory, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("memories_for_%s", elementID)
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
