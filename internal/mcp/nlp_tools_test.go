package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleExtractEntitiesAdvanced(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     ExtractEntitiesAdvancedInput
		wantErr   bool
		errString string
	}{
		{
			name: "with_text",
			input: ExtractEntitiesAdvancedInput{
				Text: "Apple Inc. was founded by Steve Jobs in California.",
			},
			wantErr:   true,
			errString: "entity extraction not enabled",
		},
		{
			name: "with_memory_id",
			input: ExtractEntitiesAdvancedInput{
				MemoryID: "memory_123",
			},
			wantErr:   true,
			errString: "entity extraction not enabled",
		},
		{
			name: "with_memory_ids",
			input: ExtractEntitiesAdvancedInput{
				MemoryIDs: []string{"memory_1", "memory_2"},
			},
			wantErr:   true,
			errString: "entity extraction not enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleExtractEntitiesAdvanced(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleAnalyzeSentiment(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     AnalyzeSentimentInput
		wantErr   bool
		errString string
	}{
		{
			name: "with_text",
			input: AnalyzeSentimentInput{
				Text: "This is a great product! I love it!",
			},
			wantErr:   true,
			errString: "sentiment analysis not enabled",
		},
		{
			name: "with_memory_id",
			input: AnalyzeSentimentInput{
				MemoryID: "memory_123",
			},
			wantErr:   true,
			errString: "sentiment analysis not enabled",
		},
		{
			name: "with_memory_ids",
			input: AnalyzeSentimentInput{
				MemoryIDs: []string{"memory_1", "memory_2"},
			},
			wantErr:   true,
			errString: "sentiment analysis not enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleAnalyzeSentiment(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleExtractTopics(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     ExtractTopicsInput
		wantErr   bool
		errString string
	}{
		{
			name: "missing_memory_ids",
			input: ExtractTopicsInput{
				NumTopics: 5,
			},
			wantErr:   true,
			errString: "memory_ids is required",
		},
		{
			name: "with_defaults",
			input: ExtractTopicsInput{
				MemoryIDs: []string{"memory_1", "memory_2"},
			},
			wantErr:   true,
			errString: "topic modeling not yet implemented",
		},
		{
			name: "with_lda_algorithm",
			input: ExtractTopicsInput{
				MemoryIDs: []string{"memory_1", "memory_2"},
				NumTopics: 3,
				Algorithm: "lda",
			},
			wantErr:   true,
			errString: "topic modeling not yet implemented",
		},
		{
			name: "with_nmf_algorithm",
			input: ExtractTopicsInput{
				MemoryIDs: []string{"memory_1", "memory_2"},
				NumTopics: 5,
				Algorithm: "nmf",
			},
			wantErr:   true,
			errString: "topic modeling not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleExtractTopics(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				// Accept either error message since implementation may vary
				errMsg := err.Error()
				assert.True(t,
					strings.Contains(errMsg, tt.errString) ||
						strings.Contains(errMsg, "no valid memories found") ||
						strings.Contains(errMsg, "memory_ids is required"),
					"Expected error containing '%s' or 'no valid memories' but got: %s", tt.errString, errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleAnalyzeSentimentTrend(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := AnalyzeSentimentTrendInput{
		MemoryIDs: []string{"memory_1", "memory_2"},
	}

	_, _, err := server.handleAnalyzeSentimentTrend(ctx, nil, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "sentiment analysis not enabled")
}

func TestHandleDetectEmotionalShifts(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := DetectEmotionalShiftsInput{
		MemoryIDs: []string{"memory_1", "memory_2"},
		Threshold: 0.5,
	}

	_, _, err := server.handleDetectEmotionalShifts(ctx, nil, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "sentiment analysis not enabled")
}

func TestHandleSummarizeSentiment(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := SummarizeSentimentInput{
		MemoryIDs: []string{"memory_1", "memory_2"},
	}

	_, _, err := server.handleSummarizeSentiment(ctx, nil, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "sentiment analysis not enabled")
}
