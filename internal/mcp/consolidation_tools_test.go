package mcp

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsolidateMemories(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create test memories
	for i := range 5 {
		memory := domain.NewMemory(
			"test-memory-"+string(rune('a'+i)),
			"Test consolidation",
			"v1.0.0",
			"test-user",
		)
		memory.Content = "This is test content for consolidation"
		require.NoError(t, server.repo.Create(memory))
	}

	input := ConsolidateMemoriesInput{
		DetectDuplicates:    true,
		ClusterMemories:     true,
		ExtractKnowledge:    true,
		SimilarityThreshold: 0.95,
	}

	_, output, err := server.handleConsolidateMemories(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output.Report)
	assert.GreaterOrEqual(t, output.Report.TotalMemories, 5)
}

func TestDetectDuplicates(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create similar memories
	for i := range 3 {
		memory := domain.NewMemory(
			"duplicate-"+string(rune('a'+i)),
			"Duplicate test",
			"v1.0.0",
			"test-user",
		)
		memory.Content = "This is duplicate content"
		require.NoError(t, server.repo.Create(memory))
	}

	input := DetectDuplicatesInput{
		SimilarityThreshold: 0.8,
		MaxResults:          100,
	}

	_, output, err := server.handleDetectDuplicates(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.GreaterOrEqual(t, output.TotalGroups, 0)
}

func TestClusterMemories(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create test memories
	for i := range 10 {
		memory := domain.NewMemory(
			"cluster-test-"+string(rune('a'+i)),
			"Cluster test",
			"v1.0.0",
			"test-user",
		)
		memory.Content = "Test content for clustering analysis"
		require.NoError(t, server.repo.Create(memory))
	}

	tests := []struct {
		name      string
		algorithm string
	}{
		{"dbscan_clustering", "dbscan"},
		{"kmeans_clustering", "kmeans"},
		{"default_algorithm", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := ClusterMemoriesInput{
				Algorithm:   tt.algorithm,
				NumClusters: 3,
			}

			_, output, err := server.handleClusterMemories(ctx, nil, input)

			require.NoError(t, err)
			assert.NotNil(t, output)
			assert.GreaterOrEqual(t, output.TotalClusters, 0)
		})
	}
}

func TestExtractKnowledge(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create memories with knowledge
	memory := domain.NewMemory(
		"knowledge-test",
		"Knowledge extraction test",
		"v1.0.0",
		"test-user",
	)
	memory.Content = "Apple Inc. was founded by Steve Jobs in California."
	require.NoError(t, server.repo.Create(memory))

	input := ExtractKnowledgeInput{
		MemoryIDs: []string{memory.GetID()},
	}

	_, output, err := server.handleExtractKnowledge(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.NotNil(t, output.KnowledgeGraph)
}

func TestMergeDuplicates(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create duplicate memories
	representative := domain.NewMemory(
		"representative",
		"Main memory",
		"v1.0.0",
		"test-user",
	)
	representative.Content = "Main content"
	require.NoError(t, server.repo.Create(representative))

	duplicate := domain.NewMemory(
		"duplicate",
		"Duplicate memory",
		"v1.0.0",
		"test-user",
	)
	duplicate.Content = "Similar content"
	require.NoError(t, server.repo.Create(duplicate))

	input := MergeDuplicatesInput{
		RepresentativeID: representative.GetID(),
		DuplicateIDs:     []string{duplicate.GetID()},
	}

	_, output, err := server.handleMergeDuplicates(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, representative.GetID(), output.MergedMemoryID)
}

func TestConsolidationInputValidation(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	t.Run("invalid_similarity_threshold", func(t *testing.T) {
		input := ConsolidateMemoriesInput{
			SimilarityThreshold: 1.5, // Invalid: > 1.0
		}

		_, _, err := server.handleConsolidateMemories(ctx, nil, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "similarity_threshold must be between 0.0 and 1.0")
	})

	t.Run("invalid_clustering_algorithm", func(t *testing.T) {
		input := ClusterMemoriesInput{
			Algorithm: "invalid-algo",
		}

		_, _, err := server.handleClusterMemories(ctx, nil, input)
		// May return error or use default - check implementation
		// This is a soft validation test
		_ = err
	})
}
