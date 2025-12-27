package application

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
)

func TestNewSemanticSearchService(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.provider == nil {
		t.Error("Expected non-nil provider")
	}
	if service.repository == nil {
		t.Error("Expected non-nil repository")
	}
	if service.store == nil {
		t.Error("Expected non-nil store")
	}
}

func TestSemanticSearchService_IndexElement(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Create memory to index
	mem := domain.NewMemory("TestMem", "This is test content", "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	err := service.IndexElement(ctx, mem)
	if err != nil {
		t.Fatalf("IndexElement failed: %v", err)
	}
}

func TestSemanticSearchService_IndexElement_Persona(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Create persona
	persona := domain.NewPersona("TestPersona", "AI assistant", "1.0.0", "test")
	repo.Create(persona)

	ctx := context.Background()
	err := service.IndexElement(ctx, persona)
	if err != nil {
		t.Fatalf("IndexElement failed for persona: %v", err)
	}
}

func TestSemanticSearchService_IndexElement_Skill(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Create skill
	skill := domain.NewSkill("Python", "Python programming", "1.0.0", "test")
	repo.Create(skill)

	ctx := context.Background()
	err := service.IndexElement(ctx, skill)
	if err != nil {
		t.Fatalf("IndexElement failed for skill: %v", err)
	}
}

func TestSemanticSearchService_IndexAllElements(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Create various elements
	mem := domain.NewMemory("Mem", "Memory content", "1.0.0", "test")
	persona := domain.NewPersona("Persona", "Description", "1.0.0", "test")
	skill := domain.NewSkill("Skill", "Skill description", "1.0.0", "test")

	repo.Create(mem)
	repo.Create(persona)
	repo.Create(skill)

	ctx := context.Background()
	err := service.IndexAllElements(ctx)
	if err != nil {
		t.Fatalf("IndexAllElements failed: %v", err)
	}
}

func TestSemanticSearchService_IndexAllElements_EmptyRepo(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	ctx := context.Background()
	err := service.IndexAllElements(ctx)
	if err != nil {
		t.Fatalf("IndexAllElements failed on empty repo: %v", err)
	}
}

func TestSemanticSearchService_Search(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index some elements
	mem1 := domain.NewMemory("Mem1", "Machine learning algorithms", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Deep learning networks", "1.0.0", "test")

	service.IndexElement(context.Background(), mem1)
	service.IndexElement(context.Background(), mem2)

	ctx := context.Background()
	results, err := service.Search(ctx, "learning", 10, "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if results == nil {
		t.Fatal("Expected non-nil results")
	}
}

func TestSemanticSearchService_Search_WithTypeFilter(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index different types
	mem := domain.NewMemory("Mem", "Memory content", "1.0.0", "test")
	persona := domain.NewPersona("Persona", "Persona description", "1.0.0", "test")

	service.IndexElement(context.Background(), mem)
	service.IndexElement(context.Background(), persona)

	ctx := context.Background()
	results, err := service.Search(ctx, "content", 10, "memory")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Results should only include memories
	for _, result := range results {
		if result.Metadata != nil {
			if typ, ok := result.Metadata["type"].(string); ok && typ != "memory" {
				t.Errorf("Expected only memory type, got: %s", typ)
			}
		}
	}
}

func TestSemanticSearchService_Search_WithLimit(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index many elements
	for i := range 20 {
		mem := domain.NewMemory("Mem"+string(rune(i)), "Content", "1.0.0", "test")
		service.IndexElement(context.Background(), mem)
	}

	ctx := context.Background()
	results, err := service.Search(ctx, "content", 5, "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) > 5 {
		t.Errorf("Expected max 5 results, got %d", len(results))
	}
}

func TestSemanticSearchService_FindSimilarMemories(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Create and index memories
	mem1 := domain.NewMemory("Mem1", "Machine learning content", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Deep learning content", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	service.IndexElement(context.Background(), mem1)
	service.IndexElement(context.Background(), mem2)

	ctx := context.Background()
	memories, err := service.FindSimilarMemories(ctx, "learning", 10)
	if err != nil {
		t.Fatalf("FindSimilarMemories failed: %v", err)
	}

	if memories == nil {
		t.Fatal("Expected non-nil memories slice")
	}
}

func TestSemanticSearchService_FindSimilarMemories_OnlyMemories(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index memory and persona
	mem := domain.NewMemory("Mem", "Memory content", "1.0.0", "test")
	persona := domain.NewPersona("Persona", "Persona content", "1.0.0", "test")

	repo.Create(mem)
	repo.Create(persona)

	service.IndexElement(context.Background(), mem)
	service.IndexElement(context.Background(), persona)

	ctx := context.Background()
	memories, err := service.FindSimilarMemories(ctx, "content", 10)
	if err != nil {
		t.Fatalf("FindSimilarMemories failed: %v", err)
	}

	// Should only return memories, not personas
	for _, memory := range memories {
		if memory.GetType() != domain.MemoryElement {
			t.Error("FindSimilarMemories returned non-memory element")
		}
	}
}

func TestSemanticSearchService_FindSimilarElements(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Create and index elements
	skill := domain.NewSkill("Python", "Python programming", "1.0.0", "test")
	repo.Create(skill)
	service.IndexElement(context.Background(), skill)

	ctx := context.Background()
	elements, err := service.FindSimilarElements(ctx, "programming", domain.SkillElement, 10)
	if err != nil {
		t.Fatalf("FindSimilarElements failed: %v", err)
	}

	if elements == nil {
		t.Fatal("Expected non-nil elements slice")
	}
}

func TestSemanticSearchService_FindSimilarElements_AnyType(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Create various elements
	mem := domain.NewMemory("Mem", "Content", "1.0.0", "test")
	persona := domain.NewPersona("Persona", "Content", "1.0.0", "test")

	repo.Create(mem)
	repo.Create(persona)

	service.IndexElement(context.Background(), mem)
	service.IndexElement(context.Background(), persona)

	ctx := context.Background()
	// Empty type filter = all types
	elements, err := service.FindSimilarElements(ctx, "content", "", 10)
	if err != nil {
		t.Fatalf("FindSimilarElements failed: %v", err)
	}

	if elements == nil {
		t.Fatal("Expected non-nil elements slice")
	}
}

func TestSemanticSearchService_GetIndexStats(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index some elements
	mem := domain.NewMemory("Mem", "Content", "1.0.0", "test")
	service.IndexElement(context.Background(), mem)

	stats := service.GetIndexStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Check expected fields
	if _, ok := stats["total_vectors"]; !ok {
		t.Error("Expected total_vectors in stats")
	}
	if _, ok := stats["provider"]; !ok {
		t.Error("Expected provider in stats")
	}
	if _, ok := stats["dimensions"]; !ok {
		t.Error("Expected dimensions in stats")
	}
}

func TestSemanticSearchService_ClearIndex(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index elements
	mem := domain.NewMemory("Mem", "Content", "1.0.0", "test")
	service.IndexElement(context.Background(), mem)

	// Clear index
	service.ClearIndex()

	// Stats should show 0 vectors
	stats := service.GetIndexStats()
	if totalVectors, ok := stats["total_vectors"].(int); ok {
		if totalVectors != 0 {
			t.Errorf("Expected 0 vectors after clear, got %d", totalVectors)
		}
	}
}

func TestSemanticSearchService_EmptyQuery(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index element
	mem := domain.NewMemory("Mem", "Content", "1.0.0", "test")
	service.IndexElement(context.Background(), mem)

	ctx := context.Background()
	results, err := service.Search(ctx, "", 10, "")
	if err != nil {
		// Empty query may return error - acceptable
		t.Logf("Empty query returned error (acceptable): %v", err)
		return
	}

	if results != nil {
		t.Logf("Empty query handled gracefully, returned %d results", len(results))
	}
}

func TestSemanticSearchService_NoResults(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Search without indexing anything
	ctx := context.Background()
	results, err := service.Search(ctx, "nonexistent", 10, "")
	if err != nil {
		// May return error if index is empty - acceptable
		t.Logf("Search on empty index returned error (acceptable): %v", err)
		return
	}

	// Results may be nil or empty slice
	if len(results) != 0 {
		t.Error("Expected 0 results for empty index")
	}
}

func TestSemanticSearchService_IndexMultipleTypes(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Create different element types
	mem := domain.NewMemory("Mem", "Memory content", "1.0.0", "test")
	persona := domain.NewPersona("Persona", "Persona description", "1.0.0", "test")
	skill := domain.NewSkill("Skill", "Skill details", "1.0.0", "test")
	agent := domain.NewAgent("Agent", "Agent info", "1.0.0", "test")

	// Index all types
	ctx := context.Background()
	if err := service.IndexElement(ctx, mem); err != nil {
		t.Fatalf("Failed to index memory: %v", err)
	}
	if err := service.IndexElement(ctx, persona); err != nil {
		t.Fatalf("Failed to index persona: %v", err)
	}
	if err := service.IndexElement(ctx, skill); err != nil {
		t.Fatalf("Failed to index skill: %v", err)
	}
	if err := service.IndexElement(ctx, agent); err != nil {
		t.Fatalf("Failed to index agent: %v", err)
	}

	// Verify all are searchable
	results, err := service.Search(ctx, "content", 10, "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected to find indexed elements")
	}
}

func TestSemanticSearchService_SearchAccuracy(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index memories with specific content
	mem1 := domain.NewMemory("ML", "Machine learning algorithms", "1.0.0", "test")
	mem2 := domain.NewMemory("DB", "Database optimization", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	service.IndexElement(context.Background(), mem1)
	service.IndexElement(context.Background(), mem2)

	ctx := context.Background()

	// Search for ML-related content
	results, err := service.Search(ctx, "machine learning", 10, "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should return results
	if len(results) == 0 {
		t.Error("Expected to find ML-related memories")
	}
}

func TestSemanticSearchService_ConcurrentIndexing(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	ctx := context.Background()

	// Index concurrently
	done := make(chan bool, 5)
	for i := range 5 {
		go func(idx int) {
			mem := domain.NewMemory("Mem"+string(rune(idx)), "Content", "1.0.0", "test")
			err := service.IndexElement(ctx, mem)
			if err != nil {
				t.Errorf("Concurrent indexing failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 5 {
		<-done
	}
}

func TestSemanticSearchService_ReindexElement(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index element
	mem := domain.NewMemory("Mem", "Original content", "1.0.0", "test")
	ctx := context.Background()
	service.IndexElement(ctx, mem)

	// Update and reindex
	mem.Content = "Updated content"
	err := service.IndexElement(ctx, mem)
	if err != nil {
		t.Fatalf("Reindexing failed: %v", err)
	}

	// Search should find updated content
	results, _ := service.Search(ctx, "updated", 10, "")
	if len(results) == 0 {
		t.Error("Expected to find reindexed element")
	}
}

func TestSemanticSearchService_LargeScale(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	ctx := context.Background()

	// Index many elements
	for i := range 100 {
		mem := domain.NewMemory("Mem"+string(rune(i)), "Content number "+string(rune(i)), "1.0.0", "test")
		err := service.IndexElement(ctx, mem)
		if err != nil {
			t.Fatalf("Failed to index element %d: %v", i, err)
		}
	}

	// Search should still work
	results, err := service.Search(ctx, "content", 10, "")
	if err != nil {
		t.Fatalf("Search failed on large index: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected to find results in large index")
	}
}

func TestSemanticSearchService_MetadataPreservation(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 128)
	repo := newMockRepoForIndex()

	service := NewSemanticSearchService(provider, repo)

	// Index with metadata
	mem := domain.NewMemory("Mem", "Content", "1.0.0", "test")
	ctx := context.Background()
	service.IndexElement(ctx, mem)

	// Search and check metadata
	results, err := service.Search(ctx, "content", 10, "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, result := range results {
		if result.Metadata == nil {
			t.Error("Expected metadata in search results")
		}
		if _, ok := result.Metadata["type"]; !ok {
			t.Error("Expected type in metadata")
		}
	}
}
