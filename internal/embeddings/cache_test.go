package embeddings

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache(100, 1*time.Hour)
	assert.NotNil(t, cache)
	assert.Equal(t, 100, cache.maxSize)
	assert.Equal(t, 1*time.Hour, cache.ttl)
}

func TestLRUCache_GetSet(t *testing.T) {
	cache := NewLRUCache(100, 1*time.Hour)

	// Test Set and Get
	cache.Put("test text", []float32{1.0, 2.0, 3.0})
	embedding, ok := cache.Get("test text")
	assert.True(t, ok)
	assert.Equal(t, []float32{1.0, 2.0, 3.0}, embedding)

	// Test Get non-existent
	_, ok = cache.Get("non-existent")
	assert.False(t, ok)
}

func TestLRUCache_TTLExpiration(t *testing.T) {
	cache := NewLRUCache(100, 10*time.Millisecond)

	// Set an embedding
	cache.Put("test", []float32{1.0, 2.0, 3.0})

	// Should exist immediately
	_, ok := cache.Get("test")
	assert.True(t, ok)

	// Wait for expiration
	time.Sleep(15 * time.Millisecond)

	// Should be expired
	_, ok = cache.Get("test")
	assert.False(t, ok)
}

func TestLRUCache_EvictionOnSizeLimit(t *testing.T) {
	cache := NewLRUCache(3, 1*time.Hour)

	// Fill cache
	cache.Put("text1", []float32{1.0})
	cache.Put("text2", []float32{2.0})
	cache.Put("text3", []float32{3.0})

	// All should exist
	_, ok := cache.Get("text1")
	assert.True(t, ok)
	_, ok = cache.Get("text2")
	assert.True(t, ok)
	_, ok = cache.Get("text3")
	assert.True(t, ok)

	// Add one more, should evict oldest (text1)
	cache.Put("text4", []float32{4.0})

	// text1 should be evicted
	_, ok = cache.Get("text1")
	assert.False(t, ok)

	// Others should still exist
	_, ok = cache.Get("text2")
	assert.True(t, ok)
	_, ok = cache.Get("text3")
	assert.True(t, ok)
	_, ok = cache.Get("text4")
	assert.True(t, ok)
}

func TestLRUCache_LRUBehavior(t *testing.T) {
	cache := NewLRUCache(3, 1*time.Hour)

	// Fill cache
	cache.Put("text1", []float32{1.0})
	cache.Put("text2", []float32{2.0})
	cache.Put("text3", []float32{3.0})

	// Access text1 to make it recently used
	_, _ = cache.Get("text1")

	// Add text4, should evict text2 (least recently used)
	cache.Put("text4", []float32{4.0})

	// text2 should be evicted
	_, ok := cache.Get("text2")
	assert.False(t, ok)

	// text1 should still exist
	_, ok = cache.Get("text1")
	assert.True(t, ok)
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(100, 1*time.Hour)

	// Add some entries
	cache.Put("text1", []float32{1.0})
	cache.Put("text2", []float32{2.0})
	cache.Put("text3", []float32{3.0})

	// Clear cache
	cache.Clear()

	// All should be gone
	_, ok := cache.Get("text1")
	assert.False(t, ok)
	_, ok = cache.Get("text2")
	assert.False(t, ok)
	_, ok = cache.Get("text3")
	assert.False(t, ok)
}

func TestCachedProvider_HitAndMiss(t *testing.T) {
	mockProvider := NewMockProvider("test", 384)
	cache := NewLRUCache(100, 1*time.Hour)
	stats := &CacheStats{}
	cachedProvider := &CachedProvider{
		provider: mockProvider,
		cache:    cache,
		stats:    stats,
	}

	ctx := context.Background()

	// First call - cache miss
	embedding1, err := cachedProvider.Embed(ctx, "test text")
	require.NoError(t, err)
	assert.NotNil(t, embedding1)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, int64(0), stats.Hits)

	// Second call - cache hit
	embedding2, err := cachedProvider.Embed(ctx, "test text")
	require.NoError(t, err)
	assert.Equal(t, embedding1, embedding2)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, int64(1), stats.Hits)
}

func TestCachedProvider_Stats(t *testing.T) {
	mockProvider := NewMockProvider("test", 384)
	config := Config{
		CacheMaxSize: 3,
		CacheTTL:     1 * time.Hour,
	}
	cachedProvider := NewCachedProvider(mockProvider, config)

	ctx := context.Background()

	// Generate some cache activity
	_, _ = cachedProvider.Embed(ctx, "text1") // miss
	_, _ = cachedProvider.Embed(ctx, "text2") // miss
	_, _ = cachedProvider.Embed(ctx, "text1") // hit
	_, _ = cachedProvider.Embed(ctx, "text3") // miss
	_, _ = cachedProvider.Embed(ctx, "text4") // miss, evicts text2

	stats := cachedProvider.GetStats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(4), stats.Misses)
	// Note: Eviction count may vary based on implementation
}

func TestCachedProvider_EmbedBatch(t *testing.T) {
	mockProvider := NewMockProvider("test", 384)
	cache := NewLRUCache(100, 1*time.Hour)
	stats := &CacheStats{}
	cachedProvider := &CachedProvider{
		provider: mockProvider,
		cache:    cache,
		stats:    stats,
	}

	ctx := context.Background()
	texts := []string{"text1", "text2", "text3"}

	// First batch - all misses
	embeddings1, err := cachedProvider.EmbedBatch(ctx, texts)
	require.NoError(t, err)
	assert.Len(t, embeddings1, 3)
	assert.Equal(t, int64(3), stats.Misses)

	// Second batch - all hits
	embeddings2, err := cachedProvider.EmbedBatch(ctx, texts)
	require.NoError(t, err)
	assert.Equal(t, embeddings1, embeddings2)
	assert.Equal(t, int64(3), stats.Hits)
}

func TestHashText(t *testing.T) {
	// Test that same text produces same hash
	hash1 := hashText("test text")
	hash2 := hashText("test text")
	assert.Equal(t, hash1, hash2)

	// Test that different text produces different hash
	hash3 := hashText("different text")
	assert.NotEqual(t, hash1, hash3)

	// Test consistency
	for range 10 {
		hash := hashText("test text")
		assert.Equal(t, hash1, hash)
	}
}
