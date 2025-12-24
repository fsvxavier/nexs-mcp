package embeddings

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// CachedProvider wraps a provider with LRU caching.
type CachedProvider struct {
	provider Provider
	cache    *LRUCache
	stats    *CacheStats
	mu       sync.RWMutex
}

// CacheStats tracks cache performance metrics.
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Evictions   int64     `json:"evictions"`
	TotalCached int       `json:"total_cached"`
	LastCleared time.Time `json:"last_cleared"`
}

// cacheEntry represents a cached embedding with TTL.
type cacheEntry struct {
	embedding []float32
	createdAt time.Time
	ttl       time.Duration
}

func (e *cacheEntry) isExpired() bool {
	return time.Since(e.createdAt) > e.ttl
}

// LRUCache implements an LRU cache for embeddings.
type LRUCache struct {
	maxSize int
	ttl     time.Duration
	items   map[string]*list.Element
	order   *list.List
	mu      sync.RWMutex
}

type lruItem struct {
	key   string
	entry *cacheEntry
}

// NewLRUCache creates a new LRU cache.
func NewLRUCache(maxSize int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		maxSize: maxSize,
		ttl:     ttl,
		items:   make(map[string]*list.Element),
		order:   list.New(),
	}
}

// Get retrieves an embedding from cache.
func (c *LRUCache) Get(text string) ([]float32, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := hashText(text)
	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}

	item := elem.Value.(*lruItem)
	if item.entry.isExpired() {
		c.removeElement(elem)
		return nil, false
	}

	// Move to front (most recently used)
	c.order.MoveToFront(elem)
	return item.entry.embedding, true
}

// Put stores an embedding in cache.
func (c *LRUCache) Put(text string, embedding []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := hashText(text)

	// Update existing entry
	if elem, ok := c.items[key]; ok {
		item := elem.Value.(*lruItem)
		item.entry.embedding = embedding
		item.entry.createdAt = time.Now()
		c.order.MoveToFront(elem)
		return
	}

	// Add new entry
	entry := &cacheEntry{
		embedding: embedding,
		createdAt: time.Now(),
		ttl:       c.ttl,
	}

	item := &lruItem{
		key:   key,
		entry: entry,
	}

	elem := c.order.PushFront(item)
	c.items[key] = elem

	// Evict oldest if over capacity
	if c.order.Len() > c.maxSize {
		oldest := c.order.Back()
		c.removeElement(oldest)
	}
}

// Clear removes all cached entries.
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.order.Init()
}

// Size returns the current cache size.
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// RemoveExpired removes all expired entries.
func (c *LRUCache) RemoveExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var toRemove []*list.Element
	for elem := c.order.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(*lruItem)
		if item.entry.isExpired() {
			toRemove = append(toRemove, elem)
		}
	}

	for _, elem := range toRemove {
		c.removeElement(elem)
	}

	return len(toRemove)
}

func (c *LRUCache) removeElement(elem *list.Element) {
	item := elem.Value.(*lruItem)
	delete(c.items, item.key)
	c.order.Remove(elem)
}

func hashText(text string) string {
	h := sha256.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

// NewCachedProvider wraps a provider with caching.
func NewCachedProvider(provider Provider, config Config) *CachedProvider {
	cache := NewLRUCache(config.CacheMaxSize, config.CacheTTL)

	return &CachedProvider{
		provider: provider,
		cache:    cache,
		stats: &CacheStats{
			LastCleared: time.Now(),
		},
	}
}

func (c *CachedProvider) Name() string {
	return c.provider.Name() + "-cached"
}

func (c *CachedProvider) Dimensions() int {
	return c.provider.Dimensions()
}

func (c *CachedProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// Try cache first
	if embedding, ok := c.cache.Get(text); ok {
		c.recordHit()
		return embedding, nil
	}

	c.recordMiss()

	// Generate embedding
	embedding, err := c.provider.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	// Cache the result
	c.cache.Put(text, embedding)
	c.updateTotalCached()

	return embedding, nil
}

func (c *CachedProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	var uncachedIndices []int
	var uncachedTexts []string

	// Check cache for each text
	for i, text := range texts {
		if embedding, ok := c.cache.Get(text); ok {
			c.recordHit()
			results[i] = embedding
		} else {
			c.recordMiss()
			uncachedIndices = append(uncachedIndices, i)
			uncachedTexts = append(uncachedTexts, text)
		}
	}

	// Generate embeddings for uncached texts
	if len(uncachedTexts) > 0 {
		embeddings, err := c.provider.EmbedBatch(ctx, uncachedTexts)
		if err != nil {
			return nil, err
		}

		// Store in cache and results
		for i, embedding := range embeddings {
			idx := uncachedIndices[i]
			results[idx] = embedding
			c.cache.Put(uncachedTexts[i], embedding)
		}

		c.updateTotalCached()
	}

	return results, nil
}

func (c *CachedProvider) IsAvailable(ctx context.Context) bool {
	return c.provider.IsAvailable(ctx)
}

func (c *CachedProvider) Cost() float64 {
	return c.provider.Cost()
}

// GetStats returns cache statistics.
func (c *CachedProvider) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := *c.stats
	stats.TotalCached = c.cache.Size()
	return stats
}

// ClearCache removes all cached embeddings.
func (c *CachedProvider) ClearCache() {
	c.cache.Clear()
	c.mu.Lock()
	c.stats.LastCleared = time.Now()
	c.mu.Unlock()
}

// CleanExpired removes expired cache entries.
func (c *CachedProvider) CleanExpired() int {
	removed := c.cache.RemoveExpired()
	c.mu.Lock()
	c.stats.Evictions += int64(removed)
	c.mu.Unlock()
	return removed
}

func (c *CachedProvider) recordHit() {
	c.mu.Lock()
	c.stats.Hits++
	c.mu.Unlock()
}

func (c *CachedProvider) recordMiss() {
	c.mu.Lock()
	c.stats.Misses++
	c.mu.Unlock()
}

func (c *CachedProvider) updateTotalCached() {
	c.mu.Lock()
	c.stats.TotalCached = c.cache.Size()
	c.mu.Unlock()
}

// HitRate returns the cache hit rate (0.0 - 1.0).
func (c *CachedProvider) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.stats.Hits + c.stats.Misses
	if total == 0 {
		return 0.0
	}

	return float64(c.stats.Hits) / float64(total)
}
