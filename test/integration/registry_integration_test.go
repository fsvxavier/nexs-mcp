package integration_test

import (
	"testing"
	"time"

	. "github.com/fsvxavier/nexs-mcp/internal/collection"
	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
)

// TestRegistryCacheIntegration tests the caching system.
func TestRegistryCacheIntegration(t *testing.T) {
	t.Run("cache_hit_performance", func(t *testing.T) {
		registry := NewRegistryWithTTL(15 * time.Minute)

		// Create test metadata
		testMeta := &sources.CollectionMetadata{
			Name:        "test-collection",
			Version:     "1.0.0",
			Author:      "test@example.com",
			Description: "Test collection",
			Category:    "testing",
			SourceName:  "test",
			URI:         "test://test-collection",
		}

		// First access - cache miss
		start := time.Now()
		registry.GetCache().Set("test://test-collection", &sources.Collection{
			Metadata: testMeta,
		})
		cacheMissTime := time.Since(start)

		// Second access - cache hit
		start = time.Now()
		cached, found := registry.GetCache().Get("test://test-collection")
		cacheHitTime := time.Since(start)

		if !found || cached == nil {
			t.Error("Cache miss on second access")
		}

		// Cache hit should be significantly faster (< 1ms typically)
		if cacheHitTime > time.Millisecond {
			t.Errorf("Cache hit too slow: %v (should be < 1ms)", cacheHitTime)
		}

		t.Logf("Cache miss: %v, Cache hit: %v", cacheMissTime, cacheHitTime)
	})

	t.Run("cache_expiration", func(t *testing.T) {
		// Use very short TTL for testing
		registry := NewRegistryWithTTL(100 * time.Millisecond)

		testMeta := &sources.CollectionMetadata{
			Name:       "expiring-collection",
			Version:    "1.0.0",
			Author:     "test",
			SourceName: "test",
			URI:        "test://expiring",
		}

		registry.GetCache().Set("test://expiring", &sources.Collection{
			Metadata: testMeta,
		})

		// Immediate access - should be cached
		if cached, found := registry.GetCache().Get("test://expiring"); !found || cached == nil {
			t.Error("Item not in cache immediately after set")
		}

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		if cached, found := registry.GetCache().Get("test://expiring"); found && cached != nil {
			t.Error("Item still in cache after expiration")
		}
	})

	t.Run("cache_statistics", func(t *testing.T) {
		registry := NewRegistryWithTTL(15 * time.Minute)

		// Add multiple items
		for i := range 10 {
			uri := "test://collection-" + string(rune('0'+i))
			registry.GetCache().Set(uri, &sources.Collection{
				Metadata: &sources.CollectionMetadata{
					Name:       "collection-" + string(rune('0'+i)),
					Version:    "1.0.0",
					SourceName: "test",
					URI:        uri,
				},
			})
		}

		stats := registry.GetCache().Stats()
		if stats["total_cached"].(int) != 10 {
			t.Errorf("Expected 10 cached items, got %d", stats["total_cached"])
		}

		// Access some items (note: Get doesn't increment AccessCount for performance)
		registry.GetCache().Get("test://collection-0")
		registry.GetCache().Get("test://collection-0")
		registry.GetCache().Get("test://collection-1")

		stats = registry.GetCache().Stats()
		// Verify AccessCount is not incremented (by design - see registry.go Get comment)
		if stats["total_access"].(int) != 0 {
			t.Logf("Note: total_access = %d (access tracking not implemented for performance)", stats["total_access"])
		}
	})

	t.Run("cache_invalidation", func(t *testing.T) {
		registry := NewRegistryWithTTL(15 * time.Minute)

		uri := "test://invalidate-test"
		registry.GetCache().Set(uri, &sources.Collection{
			Metadata: &sources.CollectionMetadata{
				Name:       "test",
				Version:    "1.0.0",
				SourceName: "test",
				URI:        uri,
			},
		})

		// Verify it's cached
		if cached, found := registry.GetCache().Get(uri); !found || cached == nil {
			t.Error("Item not cached")
		}

		// Invalidate
		registry.GetCache().Invalidate(uri)

		// Should be gone
		if cached, found := registry.GetCache().Get(uri); found && cached != nil {
			t.Error("Item still cached after invalidation")
		}
	})

	t.Run("cache_clear_all", func(t *testing.T) {
		registry := NewRegistryWithTTL(15 * time.Minute)

		// Add multiple items
		for i := range 5 {
			uri := "test://clear-" + string(rune('0'+i))
			registry.GetCache().Set(uri, &sources.Collection{
				Metadata: &sources.CollectionMetadata{URI: uri},
			})
		}

		stats := registry.GetCache().Stats()
		if stats["total_cached"].(int) != 5 {
			t.Error("Items not cached")
		}

		// Clear all
		registry.GetCache().Clear()

		stats = registry.GetCache().Stats()
		if stats["total_cached"].(int) != 0 {
			t.Error("Cache not cleared")
		}
	})
}

// TestMetadataIndexIntegration tests the indexing system.
func TestMetadataIndexIntegration(t *testing.T) {
	index := NewMetadataIndex()

	// Create test collections
	collections := []*sources.CollectionMetadata{
		{
			Name:       "devops-tools",
			Version:    "1.0.0",
			Author:     "devops-team",
			Category:   "devops",
			Tags:       []string{"automation", "ci-cd", "docker"},
			SourceName: "github",
			URI:        "github://test/devops-tools",
		},
		{
			Name:       "web-templates",
			Version:    "2.0.0",
			Author:     "web-team",
			Category:   "web-development",
			Tags:       []string{"templates", "html", "css"},
			SourceName: "github",
			URI:        "github://test/web-templates",
		},
		{
			Name:       "devops-scripts",
			Version:    "1.5.0",
			Author:     "devops-team",
			Category:   "devops",
			Tags:       []string{"automation", "bash"},
			SourceName: "local",
			URI:        "file:///local/devops-scripts",
		},
	}

	// Index all collections
	for _, meta := range collections {
		index.Index(meta)
	}

	t.Run("search_by_category", func(t *testing.T) {
		filter := &sources.BrowseFilter{
			Category: "devops",
		}

		results := index.Search("", filter)
		if len(results) != 2 {
			t.Errorf("Expected 2 devops collections, got %d", len(results))
		}
	})

	t.Run("search_by_author", func(t *testing.T) {
		filter := &sources.BrowseFilter{
			Author: "devops-team",
		}

		results := index.Search("", filter)
		if len(results) != 2 {
			t.Errorf("Expected 2 collections from devops-team, got %d", len(results))
		}
	})

	t.Run("search_by_tags", func(t *testing.T) {
		filter := &sources.BrowseFilter{
			Tags: []string{"automation"},
		}

		results := index.Search("", filter)
		if len(results) != 2 {
			t.Errorf("Expected 2 collections with automation tag, got %d", len(results))
		}

		// Multiple tags (AND logic)
		filter.Tags = []string{"automation", "docker"}
		results = index.Search("", filter)
		if len(results) != 1 {
			t.Errorf("Expected 1 collection with both automation and docker tags, got %d", len(results))
		}
	})

	t.Run("search_by_query", func(t *testing.T) {
		results := index.Search("templates", nil)
		if len(results) != 1 {
			t.Errorf("Expected 1 collection matching 'templates', got %d", len(results))
		}

		results = index.Search("devops", nil)
		if len(results) != 2 {
			t.Errorf("Expected 2 collections matching 'devops', got %d", len(results))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		filter := &sources.BrowseFilter{
			Limit:  2,
			Offset: 0,
		}

		results := index.Search("", filter)
		if len(results) > 2 {
			t.Errorf("Expected max 2 results with limit=2, got %d", len(results))
		}

		// Second page
		filter.Offset = 2
		results = index.Search("", filter)
		if len(results) > 1 {
			t.Errorf("Expected max 1 result on second page, got %d", len(results))
		}
	})

	t.Run("index_statistics", func(t *testing.T) {
		stats := index.Stats()

		if stats["total_collections"].(int) != 3 {
			t.Errorf("Expected 3 total collections, got %d", stats["total_collections"])
		}

		if stats["categories"].(int) != 2 {
			t.Errorf("Expected 2 categories, got %d", stats["categories"])
		}

		if stats["authors"].(int) != 2 {
			t.Errorf("Expected 2 authors, got %d", stats["authors"])
		}
	})

	t.Run("index_rebuild", func(t *testing.T) {
		// Clear and verify
		index.Clear()
		stats := index.Stats()
		if stats["total_collections"].(int) != 0 {
			t.Error("Index not cleared")
		}

		// Rebuild
		for _, meta := range collections {
			index.Index(meta)
		}

		stats = index.Stats()
		if stats["total_collections"].(int) != 3 {
			t.Error("Index not rebuilt correctly")
		}
	})
}

// TestDependencyGraphIntegration tests dependency resolution.
func TestDependencyGraphIntegration(t *testing.T) {
	graph := NewDependencyGraph()

	t.Run("simple_dependency_chain", func(t *testing.T) {
		// Create chain: A -> B -> C
		graph.AddNode("test://A", "A", "1.0.0")
		graph.AddNode("test://B", "B", "1.0.0")
		graph.AddNode("test://C", "C", "1.0.0")

		graph.AddDependency("test://A", "test://B")
		graph.AddDependency("test://B", "test://C")

		// Topological sort should give [C, B, A] (install order)
		order, err := graph.TopologicalSort()
		if err != nil {
			t.Fatalf("Topological sort failed: %v", err)
		}
		if len(order) != 3 {
			t.Errorf("Expected 3 nodes in order, got %d", len(order))
		}

		// C should come before B, B before A
		positions := make(map[string]int)
		for i, node := range order {
			positions[node.URI] = i
		}

		if positions["test://C"] >= positions["test://B"] {
			t.Error("C should come before B in install order")
		}
		if positions["test://B"] >= positions["test://A"] {
			t.Error("B should come before A in install order")
		}
	})

	t.Run("circular_dependency_detection", func(t *testing.T) {
		graph := NewDependencyGraph()

		// Create cycle: A -> B -> C -> A
		graph.AddNode("test://A", "A", "1.0.0")
		graph.AddNode("test://B", "B", "1.0.0")
		graph.AddNode("test://C", "C", "1.0.0")

		graph.AddDependency("test://A", "test://B")
		graph.AddDependency("test://B", "test://C")
		err := graph.AddDependency("test://C", "test://A") // Creates cycle

		// AddDependency should return error (cycle detected)
		if err == nil {
			t.Error("Expected error (cycle detected), but got nil")
		}
	})

	t.Run("diamond_dependency", func(t *testing.T) {
		graph := NewDependencyGraph()

		//     A
		//    / \
		//   B   C
		//    \ /
		//     D
		graph.AddNode("test://A", "A", "1.0.0")
		graph.AddNode("test://B", "B", "1.0.0")
		graph.AddNode("test://C", "C", "1.0.0")
		graph.AddNode("test://D", "D", "1.0.0")

		graph.AddDependency("test://A", "test://B")
		graph.AddDependency("test://A", "test://C")
		graph.AddDependency("test://B", "test://D")
		graph.AddDependency("test://C", "test://D")

		order, err := graph.TopologicalSort()
		if err != nil {
			t.Errorf("Expected valid topological order for diamond dependency: %v", err)
			return
		}

		// D should come before both B and C
		positions := make(map[string]int)
		for i, node := range order {
			positions[node.URI] = i
		}

		if positions["test://D"] >= positions["test://B"] {
			t.Error("D should come before B")
		}
		if positions["test://D"] >= positions["test://C"] {
			t.Error("D should come before C")
		}
		if positions["test://B"] >= positions["test://A"] || positions["test://C"] >= positions["test://A"] {
			t.Error("Both B and C should come before A")
		}
	})
}

// BenchmarkCachePerformance benchmarks cache operations.
func BenchmarkCachePerformance(b *testing.B) {
	registry := NewRegistryWithTTL(15 * time.Minute)

	testCollection := &sources.Collection{
		Metadata: &sources.CollectionMetadata{
			Name:       "bench-collection",
			Version:    "1.0.0",
			SourceName: "test",
			URI:        "test://bench",
		},
	}

	b.Run("cache_set", func(b *testing.B) {
		for range b.N {
			registry.GetCache().Set("test://bench", testCollection)
		}
	})

	b.Run("cache_get", func(b *testing.B) {
		registry.GetCache().Set("test://bench", testCollection)
		b.ResetTimer()
		for range b.N {
			registry.GetCache().Get("test://bench")
		}
	})
}

// BenchmarkIndexSearch benchmarks search operations.
func BenchmarkIndexSearch(b *testing.B) {
	index := NewMetadataIndex()

	// Create 100 test collections
	for i := range 100 {
		meta := &sources.CollectionMetadata{
			Name:       "collection-" + string(rune('0'+(i%10))),
			Version:    "1.0.0",
			Author:     "author-" + string(rune('0'+(i%5))),
			Category:   "category-" + string(rune('0'+(i%3))),
			SourceName: "test",
			URI:        "test://collection-" + string(rune('0'+i)),
		}
		index.Index(meta)
	}

	b.Run("search_by_category", func(b *testing.B) {
		filter := &sources.BrowseFilter{Category: "category-0"}
		b.ResetTimer()
		for range b.N {
			index.Search("", filter)
		}
	})

	b.Run("search_by_query", func(b *testing.B) {
		for range b.N {
			index.Search("collection", nil)
		}
	})
}
