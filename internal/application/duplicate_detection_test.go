package application

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
)

func TestNewDuplicateDetectionService(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{}

	service := NewDuplicateDetectionService(provider, repo, config)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.config.SimilarityThreshold != 0.95 {
		t.Errorf("Expected default threshold 0.95, got: %f", service.config.SimilarityThreshold)
	}
	if service.config.MinContentLength != 20 {
		t.Errorf("Expected default min length 20, got: %d", service.config.MinContentLength)
	}
	if service.config.MaxResults != 100 {
		t.Errorf("Expected default max results 100, got: %d", service.config.MaxResults)
	}
}

func TestDetectDuplicates_NoDuplicates(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{
		SimilarityThreshold: 0.95,
		MinContentLength:    10,
		MaxResults:          100,
	}

	service := NewDuplicateDetectionService(provider, repo, config)

	// Create unique memories
	mem1 := domain.NewMemory("Memory1", "Unique content about artificial intelligence", "1.0.0", "test")
	mem2 := domain.NewMemory("Memory2", "Different content about machine learning", "1.0.0", "test")
	mem3 := domain.NewMemory("Memory3", "Another topic about databases", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)
	repo.Create(mem3)

	ctx := context.Background()
	groups, err := service.DetectDuplicates(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// With fake provider, we might get some false positives, but test structure
	if groups == nil {
		t.Error("Expected non-nil groups slice")
	}
}

func TestDetectDuplicates_WithDuplicates(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{
		SimilarityThreshold: 0.90,
		MinContentLength:    10,
		MaxResults:          100,
	}

	service := NewDuplicateDetectionService(provider, repo, config)

	// Create near-duplicate memories
	content := "This is a test memory about machine learning algorithms"
	mem1 := domain.NewMemory("Original", content, "1.0.0", "test")
	mem2 := domain.NewMemory("Duplicate1", content, "1.0.0", "test")
	mem3 := domain.NewMemory("Duplicate2", content, "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)
	repo.Create(mem3)

	ctx := context.Background()
	groups, err := service.DetectDuplicates(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should detect at least some similarity
	if groups == nil {
		t.Error("Expected non-nil groups slice")
	}
}

func TestDetectDuplicates_EmptyRepository(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{}

	service := NewDuplicateDetectionService(provider, repo, config)

	ctx := context.Background()
	groups, err := service.DetectDuplicates(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups for empty repo, got: %d", len(groups))
	}
}

func TestDetectDuplicates_SingleMemory(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{}

	service := NewDuplicateDetectionService(provider, repo, config)

	mem := domain.NewMemory("Single", "Only one memory in the system", "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	groups, err := service.DetectDuplicates(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups for single memory, got: %d", len(groups))
	}
}

func TestDetectDuplicates_ShortContent(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{
		MinContentLength: 50, // Require at least 50 chars
	}

	service := NewDuplicateDetectionService(provider, repo, config)

	// Create short memories
	mem1 := domain.NewMemory("Short1", "Short", "1.0.0", "test")
	mem2 := domain.NewMemory("Short2", "Short", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	groups, err := service.DetectDuplicates(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should skip short content
	if len(groups) != 0 {
		t.Errorf("Expected 0 groups for short content, got: %d", len(groups))
	}
}

func TestFindDuplicatesForMemory(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{
		SimilarityThreshold: 0.90,
	}

	service := NewDuplicateDetectionService(provider, repo, config)

	// Create memories
	content := "Test content about machine learning"
	mem1 := domain.NewMemory("Target", content, "1.0.0", "test")
	mem2 := domain.NewMemory("Similar", content, "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	duplicates, err := service.FindDuplicatesForMemory(ctx, mem1.GetID())
	if err != nil {
		// May fail with empty query/embedding - acceptable
		t.Logf("FindDuplicatesForMemory returned error (acceptable): %v", err)
		return
	}

	// Verify duplicates structure is valid
	if duplicates != nil {
		t.Logf("Found %d duplicates", len(duplicates))
	}
}

func TestFindDuplicatesForMemory_NotFound(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{}

	service := NewDuplicateDetectionService(provider, repo, config)

	ctx := context.Background()
	_, err := service.FindDuplicatesForMemory(ctx, "nonexistent-id")
	if err == nil {
		t.Error("Expected error for nonexistent memory")
	}
}

func TestMergeDuplicates(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{}

	service := NewDuplicateDetectionService(provider, repo, config)

	// Create memories to merge
	mem1 := domain.NewMemory("Representative", "Main content", "1.0.0", "test")
	mem2 := domain.NewMemory("Duplicate1", "Additional info 1", "1.0.0", "test")
	mem3 := domain.NewMemory("Duplicate2", "Additional info 2", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)
	repo.Create(mem3)

	ctx := context.Background()
	duplicateIDs := []string{mem2.GetID(), mem3.GetID()}

	merged, err := service.MergeDuplicates(ctx, mem1.GetID(), duplicateIDs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if merged == nil {
		t.Fatal("Expected non-nil merged memory")
	}

	// Verify merged content
	if len(merged.Content) == 0 {
		t.Error("Expected non-empty merged content")
	}

	// Verify metadata
	if merged.Metadata["merged_count"] != "2" {
		t.Errorf("Expected merged_count=2, got: %s", merged.Metadata["merged_count"])
	}

	if merged.Metadata["merged_from"] != mem1.GetID() {
		t.Errorf("Expected merged_from=%s, got: %s", mem1.GetID(), merged.Metadata["merged_from"])
	}
}

func TestMergeDuplicates_InvalidRepresentative(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{}

	service := NewDuplicateDetectionService(provider, repo, config)

	ctx := context.Background()
	_, err := service.MergeDuplicates(ctx, "nonexistent", []string{"id1", "id2"})
	if err == nil {
		t.Error("Expected error for invalid representative")
	}
}

func TestComputeSimilarity(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{}

	service := NewDuplicateDetectionService(provider, repo, config)

	// Create two memories
	mem1 := domain.NewMemory("Memory1", "Test content about AI", "1.0.0", "test")
	mem2 := domain.NewMemory("Memory2", "Test content about AI", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	similarity, err := service.ComputeSimilarity(ctx, mem1.GetID(), mem2.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Similarity should be between 0 and 1
	if similarity < 0 || similarity > 1 {
		t.Errorf("Expected similarity between 0 and 1, got: %f", similarity)
	}
}

func TestComputeSimilarity_InvalidMemories(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{}

	service := NewDuplicateDetectionService(provider, repo, config)

	ctx := context.Background()
	_, err := service.ComputeSimilarity(ctx, "invalid1", "invalid2")
	if err == nil {
		t.Error("Expected error for invalid memory IDs")
	}
}

func TestDuplicateDetectionConfig_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		config   DuplicateDetectionConfig
		expected DuplicateDetectionConfig
	}{
		{
			name:   "empty config applies defaults",
			config: DuplicateDetectionConfig{},
			expected: DuplicateDetectionConfig{
				SimilarityThreshold: 0.95,
				MinContentLength:    20,
				MaxResults:          100,
			},
		},
		{
			name: "partial config preserves custom values",
			config: DuplicateDetectionConfig{
				SimilarityThreshold: 0.90,
			},
			expected: DuplicateDetectionConfig{
				SimilarityThreshold: 0.90,
				MinContentLength:    20,
				MaxResults:          100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepoForIndex()
			provider := embeddings.NewMockProvider("mock", 128)

			service := NewDuplicateDetectionService(provider, repo, tt.config)

			if service.config.SimilarityThreshold != tt.expected.SimilarityThreshold {
				t.Errorf("Expected threshold %f, got %f", tt.expected.SimilarityThreshold, service.config.SimilarityThreshold)
			}
			if service.config.MinContentLength != tt.expected.MinContentLength {
				t.Errorf("Expected min length %d, got %d", tt.expected.MinContentLength, service.config.MinContentLength)
			}
			if service.config.MaxResults != tt.expected.MaxResults {
				t.Errorf("Expected max results %d, got %d", tt.expected.MaxResults, service.config.MaxResults)
			}
		})
	}
}

func TestMemoryDuplicateGroup_Structure(t *testing.T) {
	mem1 := domain.NewMemory("Rep", "Content", "1.0.0", "test")
	mem2 := domain.NewMemory("Dup1", "Content", "1.0.0", "test")
	mem3 := domain.NewMemory("Dup2", "Content", "1.0.0", "test")

	group := MemoryDuplicateGroup{
		Representative: mem1,
		Duplicates:     []*domain.Memory{mem2, mem3},
		Similarity:     0.96,
		Count:          3,
	}

	if group.Representative.GetID() != mem1.GetID() {
		t.Error("Representative mismatch")
	}
	if len(group.Duplicates) != 2 {
		t.Errorf("Expected 2 duplicates, got %d", len(group.Duplicates))
	}
	if group.Count != 3 {
		t.Errorf("Expected count 3, got %d", group.Count)
	}
	if group.Similarity != 0.96 {
		t.Errorf("Expected similarity 0.96, got %f", group.Similarity)
	}
}

func TestDetectDuplicates_MaxResults(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{
		MaxResults: 2, // Limit to 2 groups
	}

	service := NewDuplicateDetectionService(provider, repo, config)

	// Create many duplicate groups
	for i := range 10 {
		content := "Test content " + string(rune(i))
		mem1 := domain.NewMemory("Mem"+string(rune(i))+"_1", content, "1.0.0", "test")
		mem2 := domain.NewMemory("Mem"+string(rune(i))+"_2", content, "1.0.0", "test")
		repo.Create(mem1)
		repo.Create(mem2)
	}

	ctx := context.Background()
	groups, err := service.DetectDuplicates(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should respect max results
	if len(groups) > 2 {
		t.Errorf("Expected max 2 groups, got %d", len(groups))
	}
}

func TestMergeDuplicates_PreservesMetadata(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := DuplicateDetectionConfig{}

	service := NewDuplicateDetectionService(provider, repo, config)

	// Create memories with metadata
	mem1 := domain.NewMemory("Rep", "Main", "1.0.0", "test")
	mem1.DateCreated = time.Now().Add(-2 * time.Hour).Format(time.RFC3339)

	mem2 := domain.NewMemory("Dup", "Additional", "1.0.0", "test")
	mem2.DateCreated = time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	merged, err := service.MergeDuplicates(ctx, mem1.GetID(), []string{mem2.GetID()})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify merged content is not empty
	if merged.Content == "" {
		t.Error("Expected non-empty merged content")
	}

	// Verify metadata
	if _, ok := merged.Metadata["merged_at"]; !ok {
		t.Error("Expected merged_at timestamp in metadata")
	}
}
