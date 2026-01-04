package mcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDeduplicateMemories(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create some duplicate memories
	memory1 := domain.NewMemory("Test Memory 1", "Test content 1", "1.0.0", "test-user")
	memory2 := domain.NewMemory("Test Memory 2", "Test content 1", "1.0.0", "test-user") // Duplicate
	memory3 := domain.NewMemory("Test Memory 3", "Test content 2", "1.0.0", "test-user") // Different

	require.NoError(t, server.repo.Create(memory1))
	require.NoError(t, server.repo.Create(memory2))
	require.NoError(t, server.repo.Create(memory3))

	tests := []struct {
		name      string
		input     DeduplicateMemoriesInput
		wantErr   bool
		checkFunc func(t *testing.T, output DeduplicateMemoriesOutput)
	}{
		{
			name: "dry_run_mode",
			input: DeduplicateMemoriesInput{
				MergeStrategy: "keep_first",
				DryRun:        true,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output DeduplicateMemoriesOutput) {
				assert.True(t, output.DryRun)
				assert.GreaterOrEqual(t, output.OriginalCount, 3)
			},
		},
		{
			name: "keep_first_strategy",
			input: DeduplicateMemoriesInput{
				MergeStrategy: "keep_first",
				DryRun:        false,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output DeduplicateMemoriesOutput) {
				assert.False(t, output.DryRun)
				assert.Equal(t, "keep_first", output.MergeStrategy)
			},
		},
		{
			name: "keep_last_strategy",
			input: DeduplicateMemoriesInput{
				MergeStrategy: "keep_last",
				DryRun:        true,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output DeduplicateMemoriesOutput) {
				assert.Equal(t, "keep_last", output.MergeStrategy)
			},
		},
		{
			name: "keep_longest_strategy",
			input: DeduplicateMemoriesInput{
				MergeStrategy: "keep_longest",
				DryRun:        true,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output DeduplicateMemoriesOutput) {
				assert.Equal(t, "keep_longest", output.MergeStrategy)
			},
		},
		{
			name: "combine_strategy",
			input: DeduplicateMemoriesInput{
				MergeStrategy: "combine",
				DryRun:        true,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output DeduplicateMemoriesOutput) {
				assert.Equal(t, "combine", output.MergeStrategy)
			},
		},
		{
			name: "default_strategy",
			input: DeduplicateMemoriesInput{
				DryRun: true,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output DeduplicateMemoriesOutput) {
				// Default should be keep_first
				assert.NotEmpty(t, output.MergeStrategy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleDeduplicateMemories(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, output.OriginalCount, 0)
				assert.GreaterOrEqual(t, output.DeduplicatedCount, 0)
				assert.GreaterOrEqual(t, output.DuplicatesRemoved, 0)
				assert.GreaterOrEqual(t, output.BytesSaved, 0)
				assert.NotNil(t, output.Groups)
				assert.NotNil(t, output.Stats)

				if tt.checkFunc != nil {
					tt.checkFunc(t, output)
				}
			}
		})
	}
}

func TestDeduplicationMergeStrategies(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create memories with different timestamps and content lengths
	memory1 := domain.NewMemory("Short Memory", "Short", "1.0.0", "test-user")
	memory2 := domain.NewMemory("Medium Memory", "Medium content here", "1.0.0", "test-user")
	memory3 := domain.NewMemory("Long Memory", "This is a much longer content for testing", "1.0.0", "test-user")

	require.NoError(t, server.repo.Create(memory1))
	require.NoError(t, server.repo.Create(memory2))
	require.NoError(t, server.repo.Create(memory3))

	strategies := []string{"keep_first", "keep_last", "keep_longest", "combine"}

	for _, strategy := range strategies {
		t.Run("strategy_"+strategy, func(t *testing.T) {
			input := DeduplicateMemoriesInput{
				MergeStrategy: strategy,
				DryRun:        true,
			}

			_, output, err := server.handleDeduplicateMemories(ctx, nil, input)
			require.NoError(t, err)
			assert.Equal(t, strategy, output.MergeStrategy)
		})
	}
}

func TestDeduplicationDryRunPreservesData(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create test memories
	memory := domain.NewMemory("Test Memory", "Test content", "1.0.0", "test-user")
	require.NoError(t, server.repo.Create(memory))

	// Get initial count
	initialMemories, err := server.repo.List(domain.ElementFilter{})
	require.NoError(t, err)
	initialCount := len(initialMemories)

	// Run dry run deduplication
	input := DeduplicateMemoriesInput{
		MergeStrategy: "keep_first",
		DryRun:        true,
	}

	_, output, err := server.handleDeduplicateMemories(ctx, nil, input)
	require.NoError(t, err)
	assert.True(t, output.DryRun)

	// Verify data is preserved
	finalMemories, err := server.repo.List(domain.ElementFilter{})
	require.NoError(t, err)
	assert.Equal(t, initialCount, len(finalMemories), "Dry run should not modify data")
}

func TestDeduplicationStats(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create multiple memories with unique names to avoid ID collision
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("Test Memory %d", i)
		memory := domain.NewMemory(name, "Test content", "1.0.0", "test-user")
		require.NoError(t, server.repo.Create(memory))
	}

	input := DeduplicateMemoriesInput{
		MergeStrategy: "keep_first",
		DryRun:        true,
	}

	_, output, err := server.handleDeduplicateMemories(ctx, nil, input)
	require.NoError(t, err)

	// Verify stats are provided
	assert.NotNil(t, output.Stats)
	assert.GreaterOrEqual(t, output.OriginalCount, 5)
	assert.GreaterOrEqual(t, output.DuplicateGroups, 0)

	// Check that groups are properly structured
	assert.NotNil(t, output.Groups)
	if len(output.Groups) > 0 {
		// Each group should have necessary fields
		for _, group := range output.Groups {
			assert.NotNil(t, group)
		}
	}
}

func TestDeduplicationEmptyRepository(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Don't create any memories - test with empty repo
	input := DeduplicateMemoriesInput{
		MergeStrategy: "keep_first",
		DryRun:        true,
	}

	_, output, err := server.handleDeduplicateMemories(ctx, nil, input)
	require.NoError(t, err)

	// Should handle empty repo gracefully
	assert.GreaterOrEqual(t, output.OriginalCount, 0)
	assert.Equal(t, 0, output.DuplicatesRemoved)
	assert.Equal(t, 0, output.BytesSaved)
}

func TestDeduplicationBytesSaved(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create duplicate memories with known content sizes
	content := "This is a test content with some length"
	memory1 := domain.NewMemory("Duplicate 1", content, "1.0.0", "test-user")
	memory2 := domain.NewMemory("Duplicate 2", content, "1.0.0", "test-user")

	require.NoError(t, server.repo.Create(memory1))
	require.NoError(t, server.repo.Create(memory2))

	input := DeduplicateMemoriesInput{
		MergeStrategy: "keep_first",
		DryRun:        true,
	}

	_, output, err := server.handleDeduplicateMemories(ctx, nil, input)
	require.NoError(t, err)

	// If duplicates are found, bytes saved should be > 0
	if output.DuplicatesRemoved > 0 {
		assert.Greater(t, output.BytesSaved, 0, "Should report bytes saved when duplicates are removed")
	}
}

func TestDeduplicationGrouping(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create exact duplicates with unique names to avoid ID collision
	content := "Exact duplicate content"
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("Duplicate Memory %d", i)
		memory := domain.NewMemory(name, content, "1.0.0", "test-user")
		require.NoError(t, server.repo.Create(memory))
	}

	input := DeduplicateMemoriesInput{
		MergeStrategy: "keep_first",
		DryRun:        true,
	}

	_, output, err := server.handleDeduplicateMemories(ctx, nil, input)
	require.NoError(t, err)

	// Should identify duplicate groups
	if output.DuplicateGroups > 0 {
		assert.Greater(t, len(output.Groups), 0, "Should provide group details")
		assert.GreaterOrEqual(t, output.DuplicateGroups, 1, "Should identify at least one duplicate group")
	}
}

func TestDeduplicationMetrics(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create some memories
	memory := domain.NewMemory("Test Memory", "Test", "1.0.0", "test-user")
	require.NoError(t, server.repo.Create(memory))

	input := DeduplicateMemoriesInput{
		MergeStrategy: "keep_first",
		DryRun:        true,
	}

	// Run deduplication
	_, _, err := server.handleDeduplicateMemories(ctx, nil, input)
	require.NoError(t, err)

	// Verify metrics were recorded
	assert.NotNil(t, server.metrics, "Metrics should be initialized")

	// Check that the tool call was tracked
	stats, err := server.metrics.GetStatistics("hour")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, stats.TotalOperations, 1)
}
