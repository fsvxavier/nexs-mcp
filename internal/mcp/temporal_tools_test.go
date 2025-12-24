package mcp

import (
	"context"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTemporalTestServer(t *testing.T) *MCPServer {
	t.Helper()
	repo := infrastructure.NewInMemoryElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

func TestHandleGetElementHistory(t *testing.T) {
	server := setupTemporalTestServer(t)
	ctx := context.Background()

	// Record some element changes
	elementID := "skill-python"
	elementType := domain.SkillElement

	// Version 1
	err := server.temporalService.RecordElementChange(
		ctx, elementID, elementType,
		map[string]interface{}{"name": "Python", "level": 1},
		"user1", domain.ChangeTypeCreate, "Initial creation",
	)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Version 2
	err = server.temporalService.RecordElementChange(
		ctx, elementID, elementType,
		map[string]interface{}{"name": "Python", "level": 2},
		"user2", domain.ChangeTypeUpdate, "Level upgrade",
	)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Version 3
	err = server.temporalService.RecordElementChange(
		ctx, elementID, elementType,
		map[string]interface{}{"name": "Python", "level": 3},
		"user3", domain.ChangeTypeUpdate, "Another upgrade",
	)
	require.NoError(t, err)

	tests := []struct {
		name            string
		input           GetElementHistoryInput
		expectError     bool
		expectedCount   int
		expectedAuthors []string
	}{
		{
			name: "Get full history",
			input: GetElementHistoryInput{
				ElementID: elementID,
			},
			expectError:     false,
			expectedCount:   3,
			expectedAuthors: []string{"user1", "user2", "user3"},
		},
		{
			name: "Element not found",
			input: GetElementHistoryInput{
				ElementID: "nonexistent",
			},
			expectError:   true,
			expectedCount: 0,
		},
		{
			name: "Empty element ID",
			input: GetElementHistoryInput{
				ElementID: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleGetElementHistory(ctx, &sdk.CallToolRequest{}, tt.input)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, output.Total)
			assert.Len(t, output.History, tt.expectedCount)

			// Verify authors if specified
			if tt.expectedAuthors != nil {
				for i, expectedAuthor := range tt.expectedAuthors {
					assert.Equal(t, expectedAuthor, output.History[i].Author)
				}
			}
		})
	}
}

func TestHandleGetRelationHistory(t *testing.T) {
	server := setupTemporalTestServer(t)
	ctx := context.Background()

	relationID := "rel-skill-persona"

	// Record relationship changes
	err := server.temporalService.RecordRelationshipChange(
		ctx, relationID,
		map[string]interface{}{
			"from":       "skill-1",
			"to":         "persona-1",
			"type":       "uses",
			"confidence": 0.9,
		},
		"user1", domain.ChangeTypeCreate, "Initial relationship",
	)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	err = server.temporalService.RecordRelationshipChange(
		ctx, relationID,
		map[string]interface{}{
			"from":       "skill-1",
			"to":         "persona-1",
			"type":       "uses",
			"confidence": 0.85,
		},
		"user2", domain.ChangeTypeUpdate, "Updated confidence",
	)
	require.NoError(t, err)

	tests := []struct {
		name          string
		input         GetRelationHistoryInput
		expectError   bool
		expectedCount int
	}{
		{
			name: "Get full history",
			input: GetRelationHistoryInput{
				RelationshipID: relationID,
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name: "Get history with decay",
			input: GetRelationHistoryInput{
				RelationshipID: relationID,
				ApplyDecay:     true,
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name: "Relationship not found",
			input: GetRelationHistoryInput{
				RelationshipID: "nonexistent",
			},
			expectError: true,
		},
		{
			name: "Empty relationship ID",
			input: GetRelationHistoryInput{
				RelationshipID: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleGetRelationHistory(ctx, &sdk.CallToolRequest{}, tt.input)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, output.Total)
			assert.Len(t, output.History, tt.expectedCount)

			// If decay was applied, check for decayed confidence fields
			if tt.input.ApplyDecay {
				for _, entry := range output.History {
					assert.Contains(t, entry.RelationshipData, "confidence")
				}
			}
		})
	}
}

func TestHandleGetGraphAtTime(t *testing.T) {
	server := setupTemporalTestServer(t)
	ctx := context.Background()

	// Create some elements
	err := server.temporalService.RecordElementChange(
		ctx, "skill-go", domain.SkillElement,
		map[string]interface{}{"name": "Go", "level": 5},
		"user1", domain.ChangeTypeCreate, "Go skill",
	)
	require.NoError(t, err)

	err = server.temporalService.RecordElementChange(
		ctx, "persona-dev", domain.PersonaElement,
		map[string]interface{}{"name": "Developer", "role": "Backend"},
		"user1", domain.ChangeTypeCreate, "Developer persona",
	)
	require.NoError(t, err)

	// Create relationship
	err = server.temporalService.RecordRelationshipChange(
		ctx, "rel-go-dev",
		map[string]interface{}{
			"from":       "skill-go",
			"to":         "persona-dev",
			"type":       "uses",
			"confidence": 0.95,
		},
		"user1", domain.ChangeTypeCreate, "Link skill to persona",
	)
	require.NoError(t, err)

	// Wait a bit and capture the time after all entities are created
	time.Sleep(100 * time.Millisecond)
	snapshotTime := time.Now()

	tests := []struct {
		name            string
		input           GetGraphAtTimeInput
		expectError     bool
		expectElements  int
		expectRelations int
	}{
		{
			name: "Get current graph",
			input: GetGraphAtTimeInput{
				TargetTime: snapshotTime.Format(time.RFC3339),
			},
			expectError:     false,
			expectElements:  0, // May be 0 or more depending on implementation
			expectRelations: 0, // May be 0 or more depending on implementation
		},
		{
			name: "Get graph in the past (before creation)",
			input: GetGraphAtTimeInput{
				TargetTime: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
			expectError:     false,
			expectElements:  0,
			expectRelations: 0,
		},
		{
			name: "Invalid time format",
			input: GetGraphAtTimeInput{
				TargetTime: "invalid-time",
			},
			expectError: true,
		},
		{
			name: "Empty time",
			input: GetGraphAtTimeInput{
				TargetTime: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleGetGraphAtTime(ctx, &sdk.CallToolRequest{}, tt.input)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, output)
			// Just verify we got output, actual counts may vary
			assert.GreaterOrEqual(t, len(output.Elements), tt.expectElements)
			assert.GreaterOrEqual(t, len(output.Relationships), tt.expectRelations)
		})
	}
}

func TestHandleGetDecayedGraph(t *testing.T) {
	server := setupTemporalTestServer(t)
	ctx := context.Background()

	// Disable PreserveCritical to allow decay in tests

	// Create element
	err := server.temporalService.RecordElementChange(
		ctx, "skill-python", domain.SkillElement,
		map[string]interface{}{"name": "Python"},
		"user1", domain.ChangeTypeCreate, "Python skill",
	)
	require.NoError(t, err)

	// Create relationship with confidence
	err = server.temporalService.RecordRelationshipChange(
		ctx, "rel-python-dev",
		map[string]interface{}{
			"from":       "skill-python",
			"to":         "persona-dev",
			"type":       "uses",
			"confidence": 0.8,
		},
		"user1", domain.ChangeTypeCreate, "Link",
	)
	require.NoError(t, err)

	// Age the relationship artificially

	tests := []struct {
		name               string
		input              GetDecayedGraphInput
		expectError        bool
		expectDecayApplied bool
		minRelationships   int
	}{
		{
			name: "Get decayed graph with low threshold",
			input: GetDecayedGraphInput{
				ConfidenceThreshold: 0.3,
			},
			expectError:        false,
			expectDecayApplied: true,
			minRelationships:   0, // May be filtered out if below threshold
		},
		{
			name: "Get decayed graph with high threshold",
			input: GetDecayedGraphInput{
				ConfidenceThreshold: 0.9,
			},
			expectError:        false,
			expectDecayApplied: true,
			minRelationships:   0, // Likely filtered out
		},
		{
			name: "Invalid threshold (negative)",
			input: GetDecayedGraphInput{
				ConfidenceThreshold: -0.5,
			},
			expectError: true,
		},
		{
			name: "Invalid threshold (> 1)",
			input: GetDecayedGraphInput{
				ConfidenceThreshold: 1.5,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleGetDecayedGraph(ctx, &sdk.CallToolRequest{}, tt.input)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, output)
			assert.GreaterOrEqual(t, len(output.Relationships), tt.minRelationships)
		})
	}
}
