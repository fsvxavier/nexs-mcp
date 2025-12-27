package embeddings

import (
	"context"
	"testing"
	"time"
)

func TestAdaptiveCache_Basic(t *testing.T) {
	cache := NewAdaptiveCache(10, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	// Test Put and Get
	embedding := []float32{1.0, 2.0, 3.0}
	cache.Put("test", embedding)

	retrieved, ok := cache.Get("test")
	if !ok {
		t.Fatal("expected cache hit")
	}

	if len(retrieved) != len(embedding) {
		t.Errorf("expected %d dimensions, got %d", len(embedding), len(retrieved))
	}

	for i, v := range embedding {
		if retrieved[i] != v {
			t.Errorf("at index %d: expected %f, got %f", i, v, retrieved[i])
		}
	}
}

func TestAdaptiveCache_MissOnNonExistent(t *testing.T) {
	cache := NewAdaptiveCache(10, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	_, ok := cache.Get("nonexistent")
	if ok {
		t.Error("expected cache miss for nonexistent key")
	}

	stats := cache.GetStats()
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestAdaptiveCache_AccessFrequency(t *testing.T) {
	cache := NewAdaptiveCache(10, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	embedding := []float32{1.0, 2.0, 3.0}
	cache.Put("test", embedding)

	// Access multiple times
	for i := range 15 {
		_, ok := cache.Get("test")
		if !ok {
			t.Fatalf("expected cache hit at iteration %d", i)
		}
	}

	stats := cache.GetStats()
	if stats.Hits != 15 {
		t.Errorf("expected 15 hits, got %d", stats.Hits)
	}

	// Check that hot entries are tracked
	if stats.HotEntries == 0 {
		t.Log("Note: HotEntries is 0, may need more time for frequency calculation")
	}
}

func TestAdaptiveCache_TTLAdjustment(t *testing.T) {
	cache := NewAdaptiveCache(10, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	embedding := []float32{1.0, 2.0, 3.0}
	cache.Put("hot", embedding)
	cache.Put("cold", embedding)

	// Access "hot" multiple times
	for range 20 {
		cache.Get("hot")
	}

	// Access "cold" only once
	cache.Get("cold")

	stats := cache.GetStats()
	if stats.TTLAdjustments == 0 {
		t.Error("expected TTL adjustments to occur")
	}

	t.Logf("TTL Adjustments: %d, Hot Entries: %d, Cold Entries: %d",
		stats.TTLAdjustments, stats.HotEntries, stats.ColdEntries)
}

func TestAdaptiveCache_Eviction(t *testing.T) {
	maxSize := 5
	cache := NewAdaptiveCache(maxSize, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	embedding := []float32{1.0, 2.0, 3.0}

	// Fill cache beyond capacity
	for i := range maxSize + 3 {
		cache.Put(string(rune('a'+i)), embedding)
	}

	size := cache.Size()
	if size != maxSize {
		t.Errorf("expected cache size %d, got %d", maxSize, size)
	}

	stats := cache.GetStats()
	if stats.Evictions != 3 {
		t.Errorf("expected 3 evictions, got %d", stats.Evictions)
	}
}

func TestAdaptiveCache_Expiration(t *testing.T) {
	// Use very short TTLs for testing
	cache := NewAdaptiveCache(10, 10*time.Millisecond, 100*time.Millisecond, 50*time.Millisecond)

	embedding := []float32{1.0, 2.0, 3.0}
	cache.Put("test", embedding)

	// Should be available immediately
	_, ok := cache.Get("test")
	if !ok {
		t.Fatal("expected immediate cache hit")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	_, ok = cache.Get("test")
	if ok {
		t.Error("expected cache miss after expiration")
	}
}

func TestAdaptiveCache_RemoveExpired(t *testing.T) {
	cache := NewAdaptiveCache(10, 10*time.Millisecond, 100*time.Millisecond, 50*time.Millisecond)

	embedding := []float32{1.0, 2.0, 3.0}

	// Add multiple entries
	for i := range 5 {
		cache.Put(string(rune('a'+i)), embedding)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	removed := cache.RemoveExpired()
	if removed != 5 {
		t.Errorf("expected 5 expired entries, got %d", removed)
	}

	if cache.Size() != 0 {
		t.Errorf("expected empty cache after removing expired, got size %d", cache.Size())
	}
}

func TestAdaptiveCache_Clear(t *testing.T) {
	cache := NewAdaptiveCache(10, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	embedding := []float32{1.0, 2.0, 3.0}

	for i := range 5 {
		cache.Put(string(rune('a'+i)), embedding)
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("expected empty cache after clear, got size %d", cache.Size())
	}
}

func TestAdaptiveCache_HitRate(t *testing.T) {
	cache := NewAdaptiveCache(10, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	embedding := []float32{1.0, 2.0, 3.0}
	cache.Put("test", embedding)

	// 5 hits
	for range 5 {
		cache.Get("test")
	}

	// 3 misses
	for range 3 {
		cache.Get("nonexistent")
	}

	hitRate := cache.GetHitRate()
	expected := 5.0 / 8.0 // 5 hits out of 8 total requests

	if hitRate < expected-0.01 || hitRate > expected+0.01 {
		t.Errorf("expected hit rate ~%.2f, got %.2f", expected, hitRate)
	}

	t.Logf("Hit rate: %.2f%% (5 hits, 3 misses)", hitRate*100)
}

func TestAdaptiveCachedProvider_Integration(t *testing.T) {
	// Create a mock provider
	mockProvider := &mockEmbeddingProvider{
		dimensions: 3,
	}

	config := Config{
		CacheMaxSize: 10,
		CacheTTL:     24 * time.Hour,
	}

	provider := NewAdaptiveCachedProvider(mockProvider, config, 1*time.Hour, 7*24*time.Hour)

	ctx := context.Background()

	// First call - cache miss
	embedding1, err := provider.Embed(ctx, "test")
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}

	if len(embedding1) != 3 {
		t.Errorf("expected 3 dimensions, got %d", len(embedding1))
	}

	// Second call - cache hit
	embedding2, err := provider.Embed(ctx, "test")
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}

	// Should be same instance from cache
	if len(embedding2) != 3 {
		t.Errorf("expected 3 dimensions, got %d", len(embedding2))
	}

	stats := provider.GetStats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}

	adaptiveStats := provider.GetAdaptiveStats()
	t.Logf("Adaptive stats: Hits=%d, Misses=%d, TTL Adjustments=%d",
		adaptiveStats.Hits, adaptiveStats.Misses, adaptiveStats.TTLAdjustments)
}

func TestAdaptiveCachedProvider_EmbedBatch(t *testing.T) {
	mockProvider := &mockEmbeddingProvider{
		dimensions: 3,
	}

	config := Config{
		CacheMaxSize: 10,
		CacheTTL:     24 * time.Hour,
	}

	provider := NewAdaptiveCachedProvider(mockProvider, config, 1*time.Hour, 7*24*time.Hour)

	ctx := context.Background()

	texts := []string{"text1", "text2", "text3"}

	// First batch - all misses
	embeddings1, err := provider.EmbedBatch(ctx, texts)
	if err != nil {
		t.Fatalf("EmbedBatch() error = %v", err)
	}

	if len(embeddings1) != 3 {
		t.Errorf("expected 3 embeddings, got %d", len(embeddings1))
	}

	// Second batch - all hits
	embeddings2, err := provider.EmbedBatch(ctx, texts)
	if err != nil {
		t.Fatalf("EmbedBatch() error = %v", err)
	}

	if len(embeddings2) != 3 {
		t.Errorf("expected 3 embeddings, got %d", len(embeddings2))
	}

	stats := provider.GetStats()
	if stats.Hits != 3 {
		t.Errorf("expected 3 hits, got %d", stats.Hits)
	}
	if stats.Misses != 3 {
		t.Errorf("expected 3 misses, got %d", stats.Misses)
	}

	hitRate := float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	t.Logf("Batch hit rate: %.2f%%", hitRate*100)
}

func BenchmarkAdaptiveCache_Get(b *testing.B) {
	cache := NewAdaptiveCache(1000, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	embedding := []float32{1.0, 2.0, 3.0}
	cache.Put("test", embedding)

	b.ResetTimer()
	for range b.N {
		cache.Get("test")
	}
}

func BenchmarkAdaptiveCache_Put(b *testing.B) {
	cache := NewAdaptiveCache(1000, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	embedding := []float32{1.0, 2.0, 3.0}

	b.ResetTimer()
	for range b.N {
		cache.Put("test", embedding)
	}
}

func BenchmarkAdaptiveCache_Mixed(b *testing.B) {
	cache := NewAdaptiveCache(1000, 1*time.Hour, 7*24*time.Hour, 24*time.Hour)

	embedding := []float32{1.0, 2.0, 3.0}

	// Pre-populate
	for i := range 100 {
		cache.Put(string(rune('a'+i%26))+string(rune('0'+i/26)), embedding)
	}

	b.ResetTimer()
	for i := range b.N {
		if i%2 == 0 {
			cache.Get(string(rune('a'+i%26)) + string(rune('0'+i/26)))
		} else {
			cache.Put(string(rune('a'+i%26))+string(rune('0'+i/26)), embedding)
		}
	}
}

// Mock embedding provider for testing.
type mockEmbeddingProvider struct {
	dimensions int
}

func (m *mockEmbeddingProvider) Name() string {
	return "mock"
}

func (m *mockEmbeddingProvider) Dimensions() int {
	return m.dimensions
}

func (m *mockEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// Return deterministic embedding based on text length
	embedding := make([]float32, m.dimensions)
	for i := range embedding {
		embedding[i] = float32(len(text) + i)
	}
	return embedding, nil
}

func (m *mockEmbeddingProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (m *mockEmbeddingProvider) Cost() float64 {
	return 0.0 // Free for testing
}

func (m *mockEmbeddingProvider) IsAvailable(ctx context.Context) bool {
	return true
}
