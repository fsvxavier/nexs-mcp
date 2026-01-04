package mcp

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/quality"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleScoreMemoryQuality(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create test memory
	memory := domain.NewMemory("Test Memory", "Test content for quality scoring", "1.0.0", "test-user")
	require.NoError(t, server.repo.Create(memory))

	tests := []struct {
		name    string
		input   ScoreMemoryQualityInput
		wantErr bool
	}{
		{
			name: "score_existing_memory",
			input: ScoreMemoryQualityInput{
				MemoryID: memory.GetID(),
			},
			wantErr: false,
		},
		{
			name: "score_with_implicit_signals",
			input: ScoreMemoryQualityInput{
				MemoryID:           memory.GetID(),
				UseImplicitSignals: true,
				ImplicitSignals: &quality.ImplicitSignals{
					AccessCount:    10,
					LastAccessDays: 1,
					ContentLength:  500,
					TagCount:       3,
				},
			},
			wantErr: false,
		},
		{
			name: "missing_memory_id",
			input: ScoreMemoryQualityInput{
				MemoryID: "",
			},
			wantErr: true,
		},
		{
			name: "non_existent_memory",
			input: ScoreMemoryQualityInput{
				MemoryID: "non_existent_id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleScoreMemoryQuality(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, output.QualityScore, 0.0)
				assert.LessOrEqual(t, output.QualityScore, 1.0)
				assert.GreaterOrEqual(t, output.Confidence, 0.0)
				assert.LessOrEqual(t, output.Confidence, 1.0)
			}
		})
	}
}

func TestHandleGetRetentionPolicy(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name            string
		input           GetRetentionPolicyInput
		expectedTier    string
		expectedMinDays int
		wantErr         bool
	}{
		{
			name:            "high_quality",
			input:           GetRetentionPolicyInput{QualityScore: 0.8},
			expectedTier:    "high",
			expectedMinDays: 300,
			wantErr:         false,
		},
		{
			name:            "medium_quality",
			input:           GetRetentionPolicyInput{QualityScore: 0.6},
			expectedTier:    "medium",
			expectedMinDays: 100,
			wantErr:         false,
		},
		{
			name:            "low_quality",
			input:           GetRetentionPolicyInput{QualityScore: 0.3},
			expectedTier:    "low",
			expectedMinDays: 30,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleGetRetentionPolicy(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, output.RetentionDays, tt.expectedMinDays)
			}
		})
	}
}

func TestHandleGetRetentionStats(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	_, output, err := server.handleGetRetentionStats(ctx, nil, struct{}{})

	require.NoError(t, err)
	assert.GreaterOrEqual(t, output.TotalScored, 0)
}
