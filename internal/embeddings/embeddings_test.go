package embeddings

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "transformers", config.Provider)
	assert.Equal(t, "text-embedding-3-small", config.OpenAIModel)
	assert.Equal(t, true, config.EnableCache)
	assert.Equal(t, 24*time.Hour, config.CacheTTL)
	assert.Equal(t, 10000, config.CacheMaxSize)
	assert.Equal(t, true, config.EnableFallback)
	assert.Contains(t, config.FallbackPriority, "transformers")
}

func TestLRUCache(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		cache := NewLRUCache(3, time.Hour)

		// Put and Get
		embedding := []float32{1.0, 2.0, 3.0}
		cache.Put("test", embedding)

		retrieved, found := cache.Get("test")
		require.True(t, found)
		assert.Equal(t, embedding, retrieved)
	})

	t.Run("LRU eviction", func(t *testing.T) {
		cache := NewLRUCache(2, time.Hour)

		cache.Put("a", []float32{1.0})
		cache.Put("b", []float32{2.0})
		cache.Put("c", []float32{3.0}) // Should evict "a"

		_, found := cache.Get("a")
		assert.False(t, found, "oldest item should be evicted")

		_, found = cache.Get("b")
		assert.True(t, found)

		_, found = cache.Get("c")
		assert.True(t, found)
	})

	t.Run("TTL expiration", func(t *testing.T) {
		cache := NewLRUCache(10, 10*time.Millisecond)

		cache.Put("test", []float32{1.0})

		// Should be found immediately
		_, found := cache.Get("test")
		require.True(t, found)

		// Wait for expiration
		time.Sleep(15 * time.Millisecond)

		// Should not be found after TTL
		_, found = cache.Get("test")
		assert.False(t, found)
	})

	t.Run("clear cache", func(t *testing.T) {
		cache := NewLRUCache(10, time.Hour)

		cache.Put("a", []float32{1.0})
		cache.Put("b", []float32{2.0})

		assert.Equal(t, 2, cache.Size())

		cache.Clear()

		assert.Equal(t, 0, cache.Size())
		_, found := cache.Get("a")
		assert.False(t, found)
	})

	t.Run("remove expired", func(t *testing.T) {
		cache := NewLRUCache(10, 10*time.Millisecond)

		cache.Put("a", []float32{1.0})
		cache.Put("b", []float32{2.0})

		time.Sleep(15 * time.Millisecond)

		removed := cache.RemoveExpired()
		assert.Equal(t, 2, removed)
		assert.Equal(t, 0, cache.Size())
	})
}

func TestCachedProvider(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		mock := NewMockProvider("mock", 384)
		config := Config{
			EnableCache:  true,
			CacheTTL:     time.Hour,
			CacheMaxSize: 100,
		}
		cached := NewCachedProvider(mock, config)

		ctx := context.Background()

		// First call - cache miss
		emb1, err := cached.Embed(ctx, "test")
		require.NoError(t, err)
		require.NotNil(t, emb1)

		stats := cached.GetStats()
		assert.Equal(t, int64(0), stats.Hits)
		assert.Equal(t, int64(1), stats.Misses)

		// Second call - cache hit
		emb2, err := cached.Embed(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, emb1, emb2)

		stats = cached.GetStats()
		assert.Equal(t, int64(1), stats.Hits)
		assert.Equal(t, int64(1), stats.Misses)
	})

	t.Run("batch with partial cache", func(t *testing.T) {
		mock := NewMockProvider("mock", 384)
		config := Config{
			EnableCache:  true,
			CacheTTL:     time.Hour,
			CacheMaxSize: 100,
		}
		cached := NewCachedProvider(mock, config)

		ctx := context.Background()

		// Cache one item
		_, err := cached.Embed(ctx, "cached")
		require.NoError(t, err)

		// Batch request with one cached and one new
		embeddings, err := cached.EmbedBatch(ctx, []string{"cached", "new"})
		require.NoError(t, err)
		assert.Len(t, embeddings, 2)

		stats := cached.GetStats()
		assert.Equal(t, int64(1), stats.Hits)   // "cached" was hit
		assert.Equal(t, int64(2), stats.Misses) // both were misses initially
	})

	t.Run("hit rate calculation", func(t *testing.T) {
		mock := NewMockProvider("mock", 384)
		config := Config{
			EnableCache:  true,
			CacheTTL:     time.Hour,
			CacheMaxSize: 100,
		}
		cached := NewCachedProvider(mock, config)

		ctx := context.Background()

		// 10 unique calls (all misses)
		for i := 0; i < 10; i++ {
			_, _ = cached.Embed(ctx, string(rune('a'+i)))
		}

		// 10 repeated calls (all hits)
		for i := 0; i < 10; i++ {
			_, _ = cached.Embed(ctx, string(rune('a'+i)))
		}

		hitRate := cached.HitRate()
		assert.InDelta(t, 0.5, hitRate, 0.01) // 10 hits / 20 total = 50%
	})

	t.Run("clear cache", func(t *testing.T) {
		mock := NewMockProvider("mock", 384)
		config := Config{
			EnableCache:  true,
			CacheTTL:     time.Hour,
			CacheMaxSize: 100,
		}
		cached := NewCachedProvider(mock, config)

		ctx := context.Background()

		_, _ = cached.Embed(ctx, "test")

		stats := cached.GetStats()
		assert.Equal(t, 1, stats.TotalCached)

		cached.ClearCache()

		stats = cached.GetStats()
		assert.Equal(t, 0, stats.TotalCached)
	})
}

func TestSimilarityMetrics(t *testing.T) {
	t.Run("similarity metric constants", func(t *testing.T) {
		assert.Equal(t, SimilarityMetric("cosine"), CosineSimilarity)
		assert.Equal(t, SimilarityMetric("euclidean"), EuclideanDistance)
		assert.Equal(t, SimilarityMetric("dotproduct"), DotProduct)
	})
}

func TestResult(t *testing.T) {
	t.Run("result structure", func(t *testing.T) {
		result := Result{
			ID:         "test-id",
			Score:      0.95,
			Text:       "test text",
			Metadata:   map[string]interface{}{"type": "memory"},
			Embedding:  []float32{1.0, 2.0, 3.0},
			Similarity: 0.95,
		}

		assert.Equal(t, "test-id", result.ID)
		assert.Equal(t, 0.95, result.Score)
		assert.Equal(t, "test text", result.Text)
		assert.Len(t, result.Embedding, 3)
	})
}
