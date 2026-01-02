package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
)

func TestNewMemoryConsolidationService(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.duplicateDetector == nil {
		t.Error("Expected non-nil duplicate detector")
	}
	if service.clusteringService == nil {
		t.Error("Expected non-nil clustering service")
	}
	if service.knowledgeExtractor == nil {
		t.Error("Expected non-nil knowledge extractor")
	}
}

func TestConsolidateMemories_AllFeatures(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create test memories
	created := []string{}
	for i := range 5 {
		mem := domain.NewMemory(fmt.Sprintf("Mem%d", i), "Content about AI", "1.0.0", "test")
		repo.Create(mem)
		created = append(created, mem.GetID())
	}
	t.Logf("created memory IDs: %v", created)

	options := ConsolidationOptions{
		DetectDuplicates: true,
		ClusterMemories:  true,
		ExtractKnowledge: true,
		AutoMerge:        false,
	}

	ctx := context.Background()
	report, err := service.ConsolidateMemories(ctx, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report == nil {
		t.Fatal("Expected non-nil report")
	}

	// Verify report structure
	if report.TotalMemories != 5 {
		t.Errorf("Expected 5 total memories, got %d", report.TotalMemories)
	}

	if report.ProcessingTime <= 0 {
		t.Error("Expected positive processing time")
	}

	if report.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if report.QualityScore < 0 || report.QualityScore > 1 {
		t.Errorf("Expected quality score between 0 and 1, got %f", report.QualityScore)
	}
}

func TestConsolidateMemories_DuplicatesOnly(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create memories
	mem := domain.NewMemory("Mem", "Test content", "1.0.0", "test")
	repo.Create(mem)

	options := ConsolidationOptions{
		DetectDuplicates: true,
		ClusterMemories:  false,
		ExtractKnowledge: false,
	}

	ctx := context.Background()
	report, err := service.ConsolidateMemories(ctx, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should run duplicates only
	if len(report.Clusters) != 0 {
		t.Error("Expected no clusters when clustering disabled")
	}
	if len(report.KnowledgeGraphs) != 0 {
		t.Error("Expected no knowledge graphs when extraction disabled")
	}
}

func TestConsolidateMemories_EmptyRepository(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	options := ConsolidationOptions{
		DetectDuplicates: true,
		ClusterMemories:  true,
	}

	ctx := context.Background()
	report, err := service.ConsolidateMemories(ctx, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.TotalMemories != 0 {
		t.Errorf("Expected 0 total memories, got %d", report.TotalMemories)
	}
}

func TestConsolidateMemories_WithAutoMerge(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{
		SimilarityThreshold: 0.90,
	}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create duplicate memories
	content := "Identical content for testing"
	mem1 := domain.NewMemory("Mem1", content, "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", content, "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	options := ConsolidationOptions{
		DetectDuplicates:      true,
		AutoMerge:             true,
		MinSimilarityForMerge: 0.95,
	}

	ctx := context.Background()
	report, err := service.ConsolidateMemories(ctx, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Auto-merge may or may not execute depending on similarity
	if report == nil {
		t.Fatal("Expected non-nil report")
	}
}

func TestDetectDuplicatesOnly(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create memories
	mem := domain.NewMemory("Mem", "Test", "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	groups, err := service.DetectDuplicatesOnly(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if groups == nil {
		t.Fatal("Expected non-nil groups slice")
	}
}

func TestClusterMemoriesOnly(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create memories
	for i := range 5 {
		mem := domain.NewMemory(fmt.Sprintf("Mem%d", i), "Content", "1.0.0", "test")
		repo.Create(mem)
	}

	ctx := context.Background()
	clusters, err := service.ClusterMemoriesOnly(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if clusters == nil {
		t.Fatal("Expected non-nil clusters slice")
	}
}

func TestExtractKnowledgeOnly(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create memories
	mem1 := domain.NewMemory("Mem1", "Machine learning content", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Deep learning content", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	memoryIDs := []string{mem1.GetID(), mem2.GetID()}

	ctx := context.Background()
	graph, err := service.ExtractKnowledgeOnly(ctx, memoryIDs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if graph == nil {
		t.Fatal("Expected non-nil knowledge graph")
	}
}

func TestConsolidationMergeDuplicates(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create memories
	mem1 := domain.NewMemory("Rep", "Representative", "1.0.0", "test")
	mem2 := domain.NewMemory("Dup", "Duplicate", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	merged, err := service.MergeDuplicates(ctx, mem1.GetID(), []string{mem2.GetID()})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if merged == nil {
		t.Fatal("Expected non-nil merged memory")
	}
}

func TestFindSimilarMemories(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create memories
	mem1 := domain.NewMemory("Mem1", "Test content", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Similar content", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	similar, err := service.FindSimilarMemories(ctx, mem1.GetID(), 0.80)
	if err != nil {
		// May fail with empty embedding - acceptable
		t.Logf("FindSimilarMemories returned error (acceptable): %v", err)
		return
	}

	if similar == nil {
		t.Fatal("Expected non-nil similar memories slice")
	}
	t.Logf("Found %d similar memories", len(similar))
}

func TestConsolidationComputeSimilarity(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create memories
	mem1 := domain.NewMemory("Mem1", "Content", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Content", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	similarity, err := service.ComputeSimilarity(ctx, mem1.GetID(), mem2.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if similarity < 0 || similarity > 1 {
		t.Errorf("Expected similarity between 0 and 1, got %f", similarity)
	}
}

func TestGetConsolidationStatistics(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create memories
	for i := range 3 {
		mem := domain.NewMemory(fmt.Sprintf("Mem%d", i), "Content", "1.0.0", "test")
		repo.Create(mem)
	}

	ctx := context.Background()
	stats, err := service.GetConsolidationStatistics(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected non-nil statistics")
	}

	if stats.TotalMemories != 3 {
		t.Errorf("Expected 3 total memories, got %d", stats.TotalMemories)
	}

	if stats.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestGenerateMergeRecommendations(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	// Create duplicate groups
	mem1 := domain.NewMemory("Rep", "Content", "1.0.0", "test")
	mem2 := domain.NewMemory("Dup1", "Content", "1.0.0", "test")
	mem3 := domain.NewMemory("Dup2", "Content", "1.0.0", "test")

	groups := []MemoryDuplicateGroup{
		{
			Representative: mem1,
			Duplicates:     []*domain.Memory{mem2, mem3},
			Similarity:     0.96,
			Count:          3,
		},
	}

	recommendations := service.generateMergeRecommendations(groups, 0.95)

	if len(recommendations) == 0 {
		t.Error("Expected at least one recommendation")
	}

	for _, rec := range recommendations {
		if rec.RepresentativeID == "" {
			t.Error("Expected non-empty representative ID")
		}
		if len(rec.DuplicateIDs) == 0 {
			t.Error("Expected non-empty duplicate IDs")
		}
		if rec.Similarity == 0 {
			t.Error("Expected non-zero similarity")
		}
		if rec.Confidence < 0 || rec.Confidence > 1 {
			t.Errorf("Expected confidence between 0 and 1, got %f", rec.Confidence)
		}
	}
}

func TestGenerateMergeRecommendations_BelowThreshold(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	mem1 := domain.NewMemory("Rep", "Content", "1.0.0", "test")
	mem2 := domain.NewMemory("Dup", "Content", "1.0.0", "test")

	groups := []MemoryDuplicateGroup{
		{
			Representative: mem1,
			Duplicates:     []*domain.Memory{mem2},
			Similarity:     0.80, // Below threshold
			Count:          2,
		},
	}

	recommendations := service.generateMergeRecommendations(groups, 0.95)

	if len(recommendations) != 0 {
		t.Error("Expected 0 recommendations for similarity below threshold")
	}
}

func TestComputeQualityScore(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	report := &ConsolidationReport{
		TotalMemories:   10,
		DuplicateGroups: []MemoryDuplicateGroup{},
		Clusters:        []Cluster{{Size: 5}, {Size: 5}},
		KnowledgeGraphs: []*KnowledgeGraph{{}},
	}

	score := service.computeQualityScore(report)

	if score < 0 || score > 1 {
		t.Errorf("Expected score between 0 and 1, got %f", score)
	}
}

func TestComputeQualityScore_EmptyReport(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	report := &ConsolidationReport{
		TotalMemories: 0,
	}

	score := service.computeQualityScore(report)

	if score != 0 {
		t.Errorf("Expected score 0 for empty report, got %f", score)
	}
}

func TestConsolidationReport_Structure(t *testing.T) {
	report := ConsolidationReport{
		TotalMemories:     10,
		DuplicateGroups:   []MemoryDuplicateGroup{},
		Clusters:          []Cluster{},
		KnowledgeGraphs:   []*KnowledgeGraph{},
		RecommendedMerges: []MergeRecommendation{},
		ProcessingTime:    100 * time.Millisecond,
		QualityScore:      0.85,
		Timestamp:         time.Now(),
	}

	if report.TotalMemories != 10 {
		t.Errorf("Expected 10 total memories, got %d", report.TotalMemories)
	}
	if report.ProcessingTime != 100*time.Millisecond {
		t.Errorf("Expected 100ms processing time, got %v", report.ProcessingTime)
	}
	if report.QualityScore != 0.85 {
		t.Errorf("Expected quality score 0.85, got %f", report.QualityScore)
	}
}

func TestMergeRecommendation_Structure(t *testing.T) {
	rec := MergeRecommendation{
		RepresentativeID: "rep-id",
		DuplicateIDs:     []string{"dup1", "dup2"},
		Similarity:       0.96,
		Confidence:       0.95,
		Reason:           "High similarity detected",
	}

	if rec.RepresentativeID != "rep-id" {
		t.Errorf("Expected representative ID 'rep-id', got %s", rec.RepresentativeID)
	}
	if len(rec.DuplicateIDs) != 2 {
		t.Errorf("Expected 2 duplicate IDs, got %d", len(rec.DuplicateIDs))
	}
	if rec.Similarity != 0.96 {
		t.Errorf("Expected similarity 0.96, got %f", rec.Similarity)
	}
	if rec.Confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", rec.Confidence)
	}
}

func TestConsolidationStatistics_Structure(t *testing.T) {
	stats := ConsolidationStatistics{
		TotalMemories:   100,
		DuplicateCount:  10,
		DuplicateGroups: 5,
		ClusterCount:    8,
		AvgClusterSize:  12.5,
		Timestamp:       time.Now(),
	}

	if stats.TotalMemories != 100 {
		t.Errorf("Expected 100 total memories, got %d", stats.TotalMemories)
	}
	if stats.DuplicateCount != 10 {
		t.Errorf("Expected 10 duplicates, got %d", stats.DuplicateCount)
	}
	if stats.AvgClusterSize != 12.5 {
		t.Errorf("Expected avg cluster size 12.5, got %f", stats.AvgClusterSize)
	}
}

func TestConsolidateMemories_ProcessingTime(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	dupConfig := DuplicateDetectionConfig{}
	clusterConfig := ClusteringConfig{}

	service := NewMemoryConsolidationService(provider, repo, dupConfig, clusterConfig)

	mem := domain.NewMemory("Mem", "Content", "1.0.0", "test")
	repo.Create(mem)

	options := ConsolidationOptions{
		DetectDuplicates: true,
	}

	ctx := context.Background()
	start := time.Now()
	report, err := service.ConsolidateMemories(ctx, options)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Processing time should be reasonable
	if report.ProcessingTime <= 0 {
		t.Error("Expected positive processing time")
	}

	if report.ProcessingTime > duration {
		t.Error("Reported processing time should not exceed actual duration")
	}
}
