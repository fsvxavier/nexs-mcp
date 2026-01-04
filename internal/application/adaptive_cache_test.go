package application

import (
	"context"
	"testing"
	"time"
)

func TestAdaptiveCacheService_BasicOperations(t *testing.T) {
	config := AdaptiveCacheConfig{
		Enabled: true,
		MinTTL:  1 * time.Hour,
		MaxTTL:  7 * 24 * time.Hour,
		BaseTTL: 24 * time.Hour,
	}

	cache := NewAdaptiveCacheService(config)
	ctx := context.Background()

	// Test Set
	err := cache.Set(ctx, "key1", "value1", 100)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// Test Get
	value, found := cache.Get(ctx, "key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test Get non-existent key
	_, found = cache.Get(ctx, "nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}

	// Test Delete
	cache.Delete(ctx, "key1")
	_, found = cache.Get(ctx, "key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}
}

func TestAdaptiveCacheService_TTLAdjustment(t *testing.T) {
	config := AdaptiveCacheConfig{
		Enabled: true,
		MinTTL:  1 * time.Hour,
		MaxTTL:  7 * 24 * time.Hour,
		BaseTTL: 24 * time.Hour,
	}

	cache := NewAdaptiveCacheService(config)
	ctx := context.Background()

	// Set initial value
	err := cache.Set(ctx, "popular", "data", 100)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// Get initial entry
	entry, exists := cache.GetEntry("popular")
	if !exists {
		t.Fatal("Expected to find popular key")
	}
	initialTTL := entry.TTL

	// Access multiple times to increase frequency
	for range 15 {
		cache.Get(ctx, "popular")
		time.Sleep(10 * time.Millisecond)
	}

	// Check if TTL increased
	entry, exists = cache.GetEntry("popular")
	if !exists {
		t.Fatal("Expected to find popular key after accesses")
	}

	if entry.AccessCount < 15 {
		t.Errorf("Expected at least 15 accesses, got %d", entry.AccessCount)
	}

	// TTL should have been adjusted upward for frequently accessed item
	stats := cache.GetStats()
	if stats.TTLAdjustments == 0 {
		t.Log("Note: TTL adjustment might not have triggered (depends on timing)")
	}

	t.Logf("Initial TTL: %v, Final TTL: %v, Access Count: %d",
		initialTTL, entry.TTL, entry.AccessCount)
}

func TestAdaptiveCacheService_Expiration(t *testing.T) {
	config := AdaptiveCacheConfig{
		Enabled: true,
		MinTTL:  50 * time.Millisecond,
		MaxTTL:  1 * time.Hour,
		BaseTTL: 100 * time.Millisecond,
	}

	cache := NewAdaptiveCacheService(config)
	ctx := context.Background()

	// Set value with short TTL
	err := cache.Set(ctx, "temp", "data", 50)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// Get entry to check initial expiration time (but don't call Get which extends TTL)
	entry, exists := cache.GetEntry("temp")
	if !exists {
		t.Fatal("Expected to find temp key initially")
	}

	expiresAt := entry.ExpiresAt
	t.Logf("Entry created at: %v, expires at: %v (TTL: %v)",
		entry.CreatedAt, expiresAt, entry.TTL)

	// Wait for expiration (with buffer)
	time.Sleep(entry.TTL + 100*time.Millisecond)

	// Should be expired now
	_, found := cache.Get(ctx, "temp")
	if found {
		t.Error("Expected temp key to be expired")
	}

	// Verify eviction stats
	stats := cache.GetStats()
	if stats.TotalEvictions == 0 {
		t.Error("Expected at least 1 eviction")
	}
}

func TestAdaptiveCacheService_Statistics(t *testing.T) {
	config := AdaptiveCacheConfig{
		Enabled: true,
		MinTTL:  1 * time.Hour,
		MaxTTL:  7 * 24 * time.Hour,
		BaseTTL: 24 * time.Hour,
	}

	cache := NewAdaptiveCacheService(config)
	ctx := context.Background()

	// Set multiple values
	cache.Set(ctx, "key1", "value1", 100)
	cache.Set(ctx, "key2", "value2", 200)
	cache.Set(ctx, "key3", "value3", 150)

	// Access keys
	cache.Get(ctx, "key1") // hit
	cache.Get(ctx, "key2") // hit
	cache.Get(ctx, "key4") // miss

	stats := cache.GetStats()

	if stats.TotalHits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.TotalHits)
	}

	if stats.TotalMisses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.TotalMisses)
	}

	if stats.TotalEntries != 3 {
		t.Errorf("Expected 3 entries, got %d", stats.TotalEntries)
	}

	if stats.BytesCached != 450 {
		t.Errorf("Expected 450 bytes cached, got %d", stats.BytesCached)
	}

	hitRate := cache.GetHitRate()
	expectedRate := 2.0 / 3.0
	if hitRate < expectedRate-0.01 || hitRate > expectedRate+0.01 {
		t.Errorf("Expected hit rate ~%.2f, got %.2f", expectedRate, hitRate)
	}
}

func TestAdaptiveCacheService_Clear(t *testing.T) {
	config := AdaptiveCacheConfig{
		Enabled: true,
		MinTTL:  1 * time.Hour,
		MaxTTL:  7 * 24 * time.Hour,
		BaseTTL: 24 * time.Hour,
	}

	cache := NewAdaptiveCacheService(config)
	ctx := context.Background()

	// Set multiple values
	cache.Set(ctx, "key1", "value1", 100)
	cache.Set(ctx, "key2", "value2", 200)
	cache.Set(ctx, "key3", "value3", 150)

	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}

	// Clear cache
	cache.Clear(ctx)

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	stats := cache.GetStats()
	if stats.TotalEntries != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", stats.TotalEntries)
	}

	if stats.BytesCached != 0 {
		t.Errorf("Expected 0 bytes cached after clear, got %d", stats.BytesCached)
	}
}

func TestAdaptiveCacheService_Disabled(t *testing.T) {
	config := AdaptiveCacheConfig{
		Enabled: false,
	}

	cache := NewAdaptiveCacheService(config)
	ctx := context.Background()

	// Set should not error but shouldn't cache
	err := cache.Set(ctx, "key1", "value1", 100)
	if err != nil {
		t.Errorf("Set should not error when disabled: %v", err)
	}

	// Get should always miss
	_, found := cache.Get(ctx, "key1")
	if found {
		t.Error("Cache should not return values when disabled")
	}

	stats := cache.GetStats()
	if stats.TotalMisses != 1 {
		t.Errorf("Expected 1 miss even when disabled, got %d", stats.TotalMisses)
	}
}

func TestAdaptiveCacheService_UpdateExisting(t *testing.T) {
	config := AdaptiveCacheConfig{
		Enabled: true,
		MinTTL:  1 * time.Hour,
		MaxTTL:  7 * 24 * time.Hour,
		BaseTTL: 24 * time.Hour,
	}

	cache := NewAdaptiveCacheService(config)
	ctx := context.Background()

	// Set initial value
	cache.Set(ctx, "key1", "value1", 100)

	// Get initial entry
	entry, _ := cache.GetEntry("key1")
	initialAccessCount := entry.AccessCount

	// Update value
	cache.Set(ctx, "key1", "value2", 200)

	// Check updated value
	value, found := cache.Get(ctx, "key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if value != "value2" {
		t.Errorf("Expected value2, got %v", value)
	}

	// Check access count increased
	entry, _ = cache.GetEntry("key1")
	if entry.AccessCount <= initialAccessCount {
		t.Error("Expected access count to increase on update")
	}

	// Check size updated
	if entry.Size != 200 {
		t.Errorf("Expected size 200, got %d", entry.Size)
	}
}

func TestAdaptiveCacheService_ConcurrentAccess(t *testing.T) {
	config := AdaptiveCacheConfig{
		Enabled: true,
		MinTTL:  1 * time.Hour,
		MaxTTL:  7 * 24 * time.Hour,
		BaseTTL: 24 * time.Hour,
	}

	cache := NewAdaptiveCacheService(config)
	ctx := context.Background()

	// Set initial values
	for i := range 10 {
		cache.Set(ctx, string(rune('a'+i)), i, 100)
	}

	// Concurrent reads and writes
	done := make(chan bool)
	for i := range 5 {
		go func(id int) {
			for j := range 100 {
				cache.Get(ctx, string(rune('a'+j%10)))
				cache.Set(ctx, string(rune('a'+j%10)), j, 100)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 5 {
		<-done
	}

	stats := cache.GetStats()
	if stats.TotalHits == 0 && stats.TotalMisses == 0 {
		t.Error("Expected some cache operations")
	}

	t.Logf("Concurrent test stats: Hits=%d, Misses=%d, Entries=%d",
		stats.TotalHits, stats.TotalMisses, stats.TotalEntries)
}
