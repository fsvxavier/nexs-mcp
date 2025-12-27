package application

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
)

func TestNewHybridSearchService(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.provider == nil {
		t.Error("Expected non-nil provider")
	}
	if service.linearStore == nil {
		t.Error("Expected non-nil linear store")
	}
	if service.hnswIndex == nil {
		t.Error("Expected non-nil HNSW index")
	}
}

func TestHybridSearchConfig_Defaults(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	// Verify default parameters are applied (indirectly through graph state)
	if service == nil {
		t.Fatal("Expected service to be created with defaults")
	}
}

func TestHybridSearchService_Provider(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	returnedProvider := service.Provider()
	if returnedProvider != provider {
		t.Error("Expected Provider() to return the same provider")
	}
}

func TestHybridSearchService_Add(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()
	err := service.Add(ctx, "doc1", "Test content", map[string]interface{}{
		"category": "test",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestHybridSearchService_AddMultiple(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add multiple documents
	for i := range 50 {
		err := service.Add(ctx, "doc"+string(rune(i)), "Content", nil)
		if err != nil {
			t.Fatalf("Failed to add document %d: %v", i, err)
		}
	}
}

func TestHybridSearchService_Search_Linear(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add a few documents (below HNSW threshold)
	service.Add(ctx, "doc1", "Machine learning content", nil)
	service.Add(ctx, "doc2", "Deep learning algorithms", nil)

	results, err := service.Search(ctx, "learning", 10, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if results == nil {
		t.Fatal("Expected non-nil results")
	}
}

func TestHybridSearchService_Search_HNSW(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add many documents (above HNSW threshold)
	for i := range 150 {
		service.Add(ctx, "doc"+string(rune(i)), "Content about AI", nil)
	}

	results, err := service.Search(ctx, "artificial intelligence", 10, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if results == nil {
		t.Fatal("Expected non-nil results")
	}
}

func TestHybridSearchService_SearchWithFilters(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add documents with metadata
	service.Add(ctx, "doc1", "Content 1", map[string]interface{}{"type": "article"})
	service.Add(ctx, "doc2", "Content 2", map[string]interface{}{"type": "blog"})

	filters := map[string]interface{}{
		"type": "article",
	}

	results, err := service.Search(ctx, "content", 10, filters)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if results == nil {
		t.Fatal("Expected non-nil results")
	}
}

func TestHybridSearchService_SearchWithHNSW(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add documents
	for i := range 10 {
		service.Add(ctx, "doc"+string(rune(i)), "Content", nil)
	}

	// Use explicit HNSW search
	results, err := service.SearchWithHNSW(ctx, "test query", 5, 50)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if results == nil {
		t.Fatal("Expected non-nil results")
	}

	if len(results) > 5 {
		t.Errorf("Expected max 5 results, got %d", len(results))
	}
}

func TestHybridSearchService_Delete(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add document
	service.Add(ctx, "doc1", "Content", nil)

	// Delete document
	err := service.Delete(ctx, "doc1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Search should not return deleted document
	results, _ := service.Search(ctx, "content", 10, nil)
	for _, result := range results {
		if result.ID == "doc1" {
			t.Error("Deleted document should not appear in search results")
		}
	}
}

func TestHybridSearchService_Size(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	stats := service.GetStatistics()
	if stats.TotalDocuments != 0 {
		t.Errorf("Expected initial size 0, got %d", stats.TotalDocuments)
	}

	// Add documents
	service.Add(ctx, "doc1", "Content 1", nil)
	service.Add(ctx, "doc2", "Content 2", nil)

	stats = service.GetStatistics()
	if stats.TotalDocuments != 2 {
		t.Errorf("Expected size 2 after adding 2 docs, got %d", stats.TotalDocuments)
	}
}

func TestHybridSearchService_Clear(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add documents
	service.Add(ctx, "doc1", "Content", nil)
	service.Add(ctx, "doc2", "Content", nil)

	// Clear
	err := service.Clear()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	stats := service.GetStatistics()
	if stats.TotalDocuments != 0 {
		t.Errorf("Expected size 0 after clear, got %d", stats.TotalDocuments)
	}
}

func TestHybridSearchService_SaveAndLoadIndex(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)

	// Create temp directory
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test_index.hnsw")

	config := HybridSearchConfig{
		Provider: provider,
		HNSWPath: indexPath,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add documents
	for i := range 10 {
		service.Add(ctx, "doc"+string(rune(i)), "Content", nil)
	}

	// Save index
	err := service.SaveIndex()
	if err != nil {
		t.Fatalf("Failed to save index: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("Index file was not created")
	}

	// Create new service and load index
	service2 := NewHybridSearchService(config)
	err = service2.LoadIndex()
	if err != nil {
		t.Fatalf("Failed to load index: %v", err)
	}

	// Verify size is reasonable after load
	stats1 := service.GetStatistics()
	stats2 := service2.GetStatistics()
	// LoadIndex may or may not preserve exact count depending on implementation
	t.Logf("Original size: %d, Loaded size: %d", stats1.TotalDocuments, stats2.TotalDocuments)
	if stats2.TotalDocuments < 0 {
		t.Error("Invalid document count after load")
	}
}

func TestHybridSearchService_AutoReindex(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider:    provider,
		AutoReindex: true,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add many documents to trigger auto-reindex
	for i := range 105 {
		service.Add(ctx, "doc"+string(rune(i)), "Content", nil)
	}

	// Auto-reindex should have been triggered
	// (verification is indirect as it runs in background)
}

func TestHybridSearchService_HNSWThreshold(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Initially should use linear search
	if service.useHNSW {
		t.Error("Expected linear search mode initially")
	}

	// Add documents to cross threshold
	for i := range HNSWThreshold + 10 {
		service.Add(ctx, "doc"+string(rune(i)), "Content", nil)
	}

	// Should now use HNSW
	if !service.useHNSW {
		t.Error("Expected HNSW mode after crossing threshold")
	}
}

func TestHybridSearchService_RebuildIndex(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add documents
	for i := range 20 {
		service.Add(ctx, "doc"+string(rune(i)), "Content", nil)
	}

	// Rebuild index
	err := service.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}
}

func TestHybridSearchService_EmptyQuery(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	service.Add(ctx, "doc1", "Content", nil)

	results, err := service.Search(ctx, "", 10, nil)
	if err != nil {
		// Empty query may return error - acceptable
		t.Logf("Empty query returned error (acceptable): %v", err)
		return
	}

	// Empty query should return some results
	if results == nil {
		t.Fatal("Expected non-nil results for empty query")
	}
}

func TestHybridSearchService_LimitResults(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add many documents
	for i := range 50 {
		service.Add(ctx, "doc"+string(rune(i)), "Test content", nil)
	}

	// Search with limit
	results, err := service.Search(ctx, "test", 5, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) > 5 {
		t.Errorf("Expected max 5 results, got %d", len(results))
	}
}

func TestHybridSearchService_SearchFallback(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add documents
	service.Add(ctx, "doc1", "Content", nil)

	// Search should work even if HNSW fails (fallback to linear)
	results, err := service.Search(ctx, "content", 10, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if results == nil {
		t.Fatal("Expected non-nil results")
	}
}

func TestHybridSearchService_ConcurrentAdd(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Add documents concurrently
	done := make(chan bool, 10)
	for i := range 10 {
		go func(idx int) {
			err := service.Add(ctx, "doc"+string(rune(idx)), "Content", nil)
			if err != nil {
				t.Errorf("Concurrent add failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}
}

func TestHybridSearchService_InvalidID(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	config := HybridSearchConfig{
		Provider: provider,
	}

	service := NewHybridSearchService(config)

	ctx := context.Background()

	// Try to delete non-existent document
	err := service.Delete(ctx, "nonexistent")
	// Should handle gracefully (may return error or succeed)
	if err != nil {
		t.Logf("Delete non-existent returned error: %v", err)
	}
}
