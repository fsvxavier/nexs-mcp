package embeddings

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// AdaptiveCacheEntry extends cache entry with access tracking.
type AdaptiveCacheEntry struct {
	embedding       []float32
	createdAt       time.Time
	expiresAt       time.Time
	accessCount     int64
	lastAccessedAt  time.Time
	accessFrequency float64 // Accesses per hour
}

// AdaptiveCache implements cache with adaptive TTL based on access patterns.
type AdaptiveCache struct {
	maxSize int
	minTTL  time.Duration
	maxTTL  time.Duration
	baseTTL time.Duration
	items   map[string]*list.Element
	order   *list.List
	stats   *AdaptiveCacheStats
	mu      sync.RWMutex
}

// AdaptiveCacheStats tracks cache behavior with adaptive metrics.
type AdaptiveCacheStats struct {
	Hits           int64         `json:"hits"`
	Misses         int64         `json:"misses"`
	Evictions      int64         `json:"evictions"`
	TotalCached    int           `json:"total_cached"`
	HotEntries     int64         `json:"hot_entries"`  // High frequency
	ColdEntries    int64         `json:"cold_entries"` // Low frequency
	AvgTTL         time.Duration `json:"avg_ttl"`
	TTLAdjustments int64         `json:"ttl_adjustments"`
	LastCleared    time.Time     `json:"last_cleared"`
}

type adaptiveLRUItem struct {
	key   string
	entry *AdaptiveCacheEntry
}

// NewAdaptiveCache creates a new adaptive cache.
func NewAdaptiveCache(maxSize int, minTTL, maxTTL, baseTTL time.Duration) *AdaptiveCache {
	return &AdaptiveCache{
		maxSize: maxSize,
		minTTL:  minTTL,
		maxTTL:  maxTTL,
		baseTTL: baseTTL,
		items:   make(map[string]*list.Element),
		order:   list.New(),
		stats: &AdaptiveCacheStats{
			LastCleared: time.Now(),
			AvgTTL:      baseTTL,
		},
	}
}

// Get retrieves an embedding and updates access tracking.
func (c *AdaptiveCache) Get(text string) ([]float32, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := hashText(text)
	elem, ok := c.items[key]
	if !ok {
		c.stats.Misses++
		return nil, false
	}

	item := elem.Value.(*adaptiveLRUItem)
	entry := item.entry

	// Check expiration
	if time.Now().After(entry.expiresAt) {
		c.removeElement(elem)
		c.stats.Misses++
		return nil, false
	}

	// Update access tracking
	entry.accessCount++
	entry.lastAccessedAt = time.Now()
	c.updateAccessFrequency(entry)

	// Adaptively adjust TTL based on access patterns
	c.adjustTTL(entry)

	// Move to front (most recently used)
	c.order.MoveToFront(elem)

	c.stats.Hits++
	return entry.embedding, true
}

// Put stores an embedding with initial TTL.
func (c *AdaptiveCache) Put(text string, embedding []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := hashText(text)
	now := time.Now()

	// Check if entry already exists
	if elem, ok := c.items[key]; ok {
		item := elem.Value.(*adaptiveLRUItem)
		item.entry.embedding = embedding
		item.entry.createdAt = now
		item.entry.lastAccessedAt = now
		c.adjustTTL(item.entry)
		c.order.MoveToFront(elem)
		return
	}

	// Add new entry
	entry := &AdaptiveCacheEntry{
		embedding:       embedding,
		createdAt:       now,
		expiresAt:       now.Add(c.baseTTL),
		accessCount:     1,
		lastAccessedAt:  now,
		accessFrequency: 0.0,
	}

	item := &adaptiveLRUItem{
		key:   key,
		entry: entry,
	}

	elem := c.order.PushFront(item)
	c.items[key] = elem

	// Evict oldest if over capacity
	if c.order.Len() > c.maxSize {
		oldest := c.order.Back()
		c.removeElement(oldest)
		c.stats.Evictions++
	}

	c.stats.TotalCached = len(c.items)
}

// updateAccessFrequency calculates accesses per hour.
func (c *AdaptiveCache) updateAccessFrequency(entry *AdaptiveCacheEntry) {
	age := time.Since(entry.createdAt).Hours()
	if age < 1.0 {
		age = 1.0 // Minimum 1 hour to avoid division issues
	}
	entry.accessFrequency = float64(entry.accessCount) / age
}

// adjustTTL dynamically adjusts TTL based on access patterns.
func (c *AdaptiveCache) adjustTTL(entry *AdaptiveCacheEntry) {
	// High frequency -> longer TTL
	// Low frequency -> shorter TTL

	var multiplier float64

	switch {
	case entry.accessFrequency > 10.0:
		// Very hot: max TTL
		multiplier = float64(c.maxTTL) / float64(c.baseTTL)
		c.stats.HotEntries++
	case entry.accessFrequency > 1.0:
		// Hot: extend TTL proportionally
		multiplier = 1.0 + (entry.accessFrequency / 10.0)
	case entry.accessFrequency < 0.1:
		// Very cold: min TTL
		multiplier = float64(c.minTTL) / float64(c.baseTTL)
		c.stats.ColdEntries++
	default:
		// Cold: reduce TTL proportionally
		multiplier = 0.5 + (entry.accessFrequency * 0.5)
	}

	newTTL := time.Duration(float64(c.baseTTL) * multiplier)

	// Clamp to min/max
	if newTTL < c.minTTL {
		newTTL = c.minTTL
	} else if newTTL > c.maxTTL {
		newTTL = c.maxTTL
	}

	// Update expiration
	entry.expiresAt = time.Now().Add(newTTL)
	c.stats.TTLAdjustments++

	// Update average TTL using exponential moving average
	alpha := 0.1
	c.stats.AvgTTL = time.Duration(alpha*float64(newTTL) + (1-alpha)*float64(c.stats.AvgTTL))
}

// Clear removes all cached entries.
func (c *AdaptiveCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.order.Init()
	c.stats.LastCleared = time.Now()
	c.stats.TotalCached = 0
}

// Size returns the current cache size.
func (c *AdaptiveCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// RemoveExpired removes all expired entries.
func (c *AdaptiveCache) RemoveExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var toRemove []*list.Element

	for elem := c.order.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(*adaptiveLRUItem)
		if now.After(item.entry.expiresAt) {
			toRemove = append(toRemove, elem)
		}
	}

	for _, elem := range toRemove {
		c.removeElement(elem)
		c.stats.Evictions++
	}

	c.stats.TotalCached = len(c.items)
	return len(toRemove)
}

// GetStats returns current cache statistics.
func (c *AdaptiveCache) GetStats() AdaptiveCacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := *c.stats
	stats.TotalCached = len(c.items)
	return stats
}

// GetHitRate returns the cache hit rate (0.0-1.0).
func (c *AdaptiveCache) GetHitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.stats.Hits + c.stats.Misses
	if total == 0 {
		return 0.0
	}
	return float64(c.stats.Hits) / float64(total)
}

func (c *AdaptiveCache) removeElement(elem *list.Element) {
	item := elem.Value.(*adaptiveLRUItem)
	delete(c.items, item.key)
	c.order.Remove(elem)
}

// AdaptiveCachedProvider wraps a provider with adaptive caching.
type AdaptiveCachedProvider struct {
	provider Provider
	cache    *AdaptiveCache
	stats    *CacheStats
	mu       sync.RWMutex
}

// NewAdaptiveCachedProvider wraps a provider with adaptive caching.
func NewAdaptiveCachedProvider(provider Provider, config Config, minTTL, maxTTL time.Duration) *AdaptiveCachedProvider {
	cache := NewAdaptiveCache(config.CacheMaxSize, minTTL, maxTTL, config.CacheTTL)

	return &AdaptiveCachedProvider{
		provider: provider,
		cache:    cache,
		stats: &CacheStats{
			LastCleared: time.Now(),
		},
	}
}

func (c *AdaptiveCachedProvider) Name() string {
	return c.provider.Name() + "-adaptive-cached"
}

func (c *AdaptiveCachedProvider) Dimensions() int {
	return c.provider.Dimensions()
}

func (c *AdaptiveCachedProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// Try cache first
	if embedding, ok := c.cache.Get(text); ok {
		c.recordHit()
		return embedding, nil
	}

	c.recordMiss()

	// Cache miss - generate embedding
	embedding, err := c.provider.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Put(text, embedding)

	return embedding, nil
}

func (c *AdaptiveCachedProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	var misses []int
	var missTexts []string

	// Check cache for each text
	for i, text := range texts {
		if embedding, ok := c.cache.Get(text); ok {
			embeddings[i] = embedding
			c.recordHit()
		} else {
			misses = append(misses, i)
			missTexts = append(missTexts, text)
			c.recordMiss()
		}
	}

	// Generate embeddings for misses
	if len(missTexts) > 0 {
		missEmbeddings, err := c.provider.EmbedBatch(ctx, missTexts)
		if err != nil {
			return nil, err
		}

		// Store in cache and fill results
		for i, missIdx := range misses {
			embeddings[missIdx] = missEmbeddings[i]
			c.cache.Put(missTexts[i], missEmbeddings[i])
		}
	}

	return embeddings, nil
}

func (c *AdaptiveCachedProvider) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := *c.stats
	stats.TotalCached = c.cache.Size()
	return stats
}

func (c *AdaptiveCachedProvider) GetAdaptiveStats() AdaptiveCacheStats {
	return c.cache.GetStats()
}

func (c *AdaptiveCachedProvider) ClearCache() {
	c.cache.Clear()
	c.mu.Lock()
	c.stats.LastCleared = time.Now()
	c.mu.Unlock()
}

func (c *AdaptiveCachedProvider) recordHit() {
	c.mu.Lock()
	c.stats.Hits++
	c.mu.Unlock()
}

func (c *AdaptiveCachedProvider) recordMiss() {
	c.mu.Lock()
	c.stats.Misses++
	c.mu.Unlock()
}
