package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

func setupIndexTestServer(t *testing.T) *MCPServer {
	t.Helper()
	repo := infrastructure.NewInMemoryElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

func TestHandleSearchCapabilityIndex(t *testing.T) {
	// Create test server
	server := setupIndexTestServer(t)

	tests := []struct {
		name        string
		input       SearchCapabilityIndexInput
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid search query",
			input: SearchCapabilityIndexInput{
				Query:      "Go programming",
				MaxResults: 10,
			},
			expectError: false,
		},
		{
			name: "Empty query",
			input: SearchCapabilityIndexInput{
				Query: "",
			},
			expectError: true,
			errorMsg:    "query cannot be empty",
		},
		{
			name: "Query with type filter",
			input: SearchCapabilityIndexInput{
				Query:      "programming",
				MaxResults: 5,
				Types:      []string{"persona", "skill"},
			},
			expectError: false,
		},
		{
			name: "Large max results gets capped",
			input: SearchCapabilityIndexInput{
				Query:      "test",
				MaxResults: 1000,
			},
			expectError: false,
		},
		{
			name: "Zero max results uses default",
			input: SearchCapabilityIndexInput{
				Query:      "test",
				MaxResults: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleSearchCapabilityIndex(
				context.Background(),
				&sdk.CallToolRequest{},
				tt.input,
			)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, output)
			assert.Equal(t, tt.input.Query, output.Query)
			assert.GreaterOrEqual(t, output.Total, 0)
			assert.NotNil(t, output.Results)

			// Verify max results cap
			if tt.input.MaxResults > 100 {
				assert.LessOrEqual(t, output.Total, 100)
			}

			// Result should be nil or have content
			if result != nil {
				assert.NotNil(t, result.Content)
			}
		})
	}
}

func TestHandleFindSimilarCapabilities(t *testing.T) {
	// Create test server and add a test element
	server := setupIndexTestServer(t)

	// Create test persona
	persona := domain.NewPersona("Test Persona", "Test description", "1.0.0", "test-author")
	err := server.repo.Create(persona)
	assert.NoError(t, err)
	elementID := persona.GetMetadata().ID

	tests := []struct {
		name        string
		input       FindSimilarCapabilitiesInput
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid element ID",
			input: FindSimilarCapabilitiesInput{
				ElementID:  elementID,
				MaxResults: 5,
			},
			expectError: false,
		},
		{
			name: "Empty element ID",
			input: FindSimilarCapabilitiesInput{
				ElementID: "",
			},
			expectError: true,
			errorMsg:    "element_id cannot be empty",
		},
		{
			name: "Non-existent element",
			input: FindSimilarCapabilitiesInput{
				ElementID: "non-existent-id",
			},
			expectError: true,
			errorMsg:    "element not found",
		},
		{
			name: "Zero max results uses default",
			input: FindSimilarCapabilitiesInput{
				ElementID:  elementID,
				MaxResults: 0,
			},
			expectError: false,
		},
		{
			name: "Large max results gets capped",
			input: FindSimilarCapabilitiesInput{
				ElementID:  elementID,
				MaxResults: 100,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleFindSimilarCapabilities(
				context.Background(),
				&sdk.CallToolRequest{},
				tt.input,
			)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, output)
			assert.Equal(t, tt.input.ElementID, output.ElementID)
			assert.GreaterOrEqual(t, output.Total, 0)
			assert.NotNil(t, output.Similar)

			// Verify max results cap
			if tt.input.MaxResults > 50 {
				assert.LessOrEqual(t, output.Total, 50)
			}

			if result != nil {
				assert.NotNil(t, result.Content)
			}
		})
	}
}

func TestHandleMapCapabilityRelationships(t *testing.T) {
	// Create test server and add a test element
	server := setupIndexTestServer(t)

	// Create test persona
	persona := domain.NewPersona("Test Persona", "Test description", "1.0.0", "test-author")
	err := server.repo.Create(persona)
	assert.NoError(t, err)
	elementID := persona.GetMetadata().ID

	tests := []struct {
		name        string
		input       MapCapabilityRelationshipsInput
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid element with default threshold",
			input: MapCapabilityRelationshipsInput{
				ElementID: elementID,
			},
			expectError: false,
		},
		{
			name: "Valid element with custom threshold",
			input: MapCapabilityRelationshipsInput{
				ElementID: elementID,
				Threshold: 0.5,
			},
			expectError: false,
		},
		{
			name: "Empty element ID",
			input: MapCapabilityRelationshipsInput{
				ElementID: "",
			},
			expectError: true,
			errorMsg:    "element_id cannot be empty",
		},
		{
			name: "Non-existent element",
			input: MapCapabilityRelationshipsInput{
				ElementID: "non-existent",
			},
			expectError: true,
			errorMsg:    "element not found",
		},
		{
			name: "Zero threshold uses default",
			input: MapCapabilityRelationshipsInput{
				ElementID: elementID,
				Threshold: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleMapCapabilityRelationships(
				context.Background(),
				&sdk.CallToolRequest{},
				tt.input,
			)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, output)
			assert.Equal(t, tt.input.ElementID, output.ElementID)
			assert.NotNil(t, output.Relationships)
			assert.NotNil(t, output.Graph)

			// Verify graph structure
			nodes, ok := output.Graph["nodes"].([]map[string]interface{})
			assert.True(t, ok)
			assert.GreaterOrEqual(t, len(nodes), 1) // At least the source node

			edges, ok := output.Graph["edges"].([]map[string]interface{})
			assert.True(t, ok)
			assert.GreaterOrEqual(t, len(edges), 0)

			// Verify relationships
			for _, rel := range output.Relationships {
				assert.NotEmpty(t, rel.TargetID)
				assert.NotEmpty(t, rel.TargetType)
				assert.GreaterOrEqual(t, rel.Similarity, 0.0)
				assert.LessOrEqual(t, rel.Similarity, 1.0)
				assert.Contains(t, []string{"similar", "complementary", "related"}, rel.RelationshipType)
			}

			if result != nil {
				assert.NotNil(t, result.Content)
			}
		})
	}
}

func TestHandleGetCapabilityIndexStats(t *testing.T) {
	server := setupIndexTestServer(t)

	input := GetCapabilityIndexStatsInput{}

	result, output, err := server.handleGetCapabilityIndexStats(
		context.Background(),
		&sdk.CallToolRequest{},
		input,
	)

	assert.NoError(t, err)
	assert.NotNil(t, output)

	// Verify output structure
	assert.GreaterOrEqual(t, output.TotalDocuments, 0)
	assert.NotNil(t, output.DocumentsByType)
	assert.GreaterOrEqual(t, output.UniqueTerms, 0)
	assert.GreaterOrEqual(t, output.AverageTermsPerDoc, 0.0)
	assert.NotEmpty(t, output.IndexHealth)
	assert.Contains(t, []string{"empty", "healthy", "degraded"}, output.IndexHealth)
	assert.NotEmpty(t, output.LastUpdated)

	if result != nil {
		assert.NotNil(t, result.Content)
	}
}

func TestSearchCapabilityIndexInputValidation(t *testing.T) {
	server := setupIndexTestServer(t)

	tests := []struct {
		name  string
		input SearchCapabilityIndexInput
		check func(t *testing.T, output SearchCapabilityIndexOutput, err error)
	}{
		{
			name: "Whitespace-only query",
			input: SearchCapabilityIndexInput{
				Query: "   ",
			},
			check: func(t *testing.T, output SearchCapabilityIndexOutput, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "Negative max results",
			input: SearchCapabilityIndexInput{
				Query:      "test",
				MaxResults: -1,
			},
			check: func(t *testing.T, output SearchCapabilityIndexOutput, err error) {
				assert.NoError(t, err)
				// Should use default (10)
			},
		},
		{
			name: "Valid types array",
			input: SearchCapabilityIndexInput{
				Query: "test",
				Types: []string{"persona", "skill", "template"},
			},
			check: func(t *testing.T, output SearchCapabilityIndexOutput, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleSearchCapabilityIndex(
				context.Background(),
				&sdk.CallToolRequest{},
				tt.input,
			)
			tt.check(t, output, err)
		})
	}
}

func TestFindSimilarCapabilitiesEdgeCases(t *testing.T) {
	server := setupIndexTestServer(t)

	// Create test element
	persona := domain.NewPersona("Edge Case Test", "Test description", "1.0.0", "test-author")
	err := server.repo.Create(persona)
	assert.NoError(t, err)
	elementID := persona.GetMetadata().ID

	t.Run("Max results boundary", func(t *testing.T) {
		_, output, err := server.handleFindSimilarCapabilities(
			context.Background(),
			&sdk.CallToolRequest{},
			FindSimilarCapabilitiesInput{
				ElementID:  elementID,
				MaxResults: 51, // Above cap
			},
		)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(output.Similar), 50)
	})

	t.Run("Whitespace element ID", func(t *testing.T) {
		_, _, err := server.handleFindSimilarCapabilities(
			context.Background(),
			&sdk.CallToolRequest{},
			FindSimilarCapabilitiesInput{
				ElementID: "   ",
			},
		)
		assert.Error(t, err)
	})
}

func TestMapCapabilityRelationshipsThresholds(t *testing.T) {
	server := setupIndexTestServer(t)

	// Create test element
	persona := domain.NewPersona("Threshold Test", "Test description", "1.0.0", "test-author")
	err := server.repo.Create(persona)
	assert.NoError(t, err)
	elementID := persona.GetMetadata().ID

	tests := []struct {
		name      string
		threshold float64
		wantMin   float64
	}{
		{"Low threshold", 0.1, 0.1},
		{"Medium threshold", 0.5, 0.5},
		{"High threshold", 0.9, 0.9},
		{"Zero uses default", 0.0, 0.3},
		{"Negative uses default", -0.1, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleMapCapabilityRelationships(
				context.Background(),
				&sdk.CallToolRequest{},
				MapCapabilityRelationshipsInput{
					ElementID: elementID,
					Threshold: tt.threshold,
				},
			)
			assert.NoError(t, err)

			// All relationships should meet minimum threshold
			for _, rel := range output.Relationships {
				assert.GreaterOrEqual(t, rel.Similarity, tt.wantMin)
			}
		})
	}
}

func TestOutputStructures(t *testing.T) {
	t.Run("SearchCapabilityIndexOutput", func(t *testing.T) {
		output := SearchCapabilityIndexOutput{
			Results: []SearchResultItem{
				{
					DocumentID: "doc-1",
					Type:       "persona",
					Name:       "Test",
					Score:      0.95,
					Highlights: []string{"highlight 1"},
				},
			},
			Query: "test query",
			Total: 1,
		}

		assert.Len(t, output.Results, 1)
		assert.Equal(t, "test query", output.Query)
		assert.Equal(t, 1, output.Total)
		assert.Equal(t, "doc-1", output.Results[0].DocumentID)
	})

	t.Run("FindSimilarCapabilitiesOutput", func(t *testing.T) {
		output := FindSimilarCapabilitiesOutput{
			Similar: []SimilarCapabilityItem{
				{
					DocumentID: "similar-1",
					Type:       "skill",
					Name:       "Similar Skill",
					Similarity: 0.85,
				},
			},
			ElementID: "source-id",
			Total:     1,
		}

		assert.Len(t, output.Similar, 1)
		assert.Equal(t, "source-id", output.ElementID)
		assert.Equal(t, 1, output.Total)
	})

	t.Run("MapCapabilityRelationshipsOutput", func(t *testing.T) {
		output := MapCapabilityRelationshipsOutput{
			ElementID: "test-id",
			Relationships: []RelationshipItem{
				{
					TargetID:         "target-1",
					TargetType:       "persona",
					TargetName:       "Target",
					Similarity:       0.9,
					RelationshipType: "similar",
				},
			},
			Graph: map[string]interface{}{
				"nodes": []map[string]interface{}{},
				"edges": []map[string]interface{}{},
			},
		}

		assert.Equal(t, "test-id", output.ElementID)
		assert.Len(t, output.Relationships, 1)
		assert.NotNil(t, output.Graph["nodes"])
		assert.NotNil(t, output.Graph["edges"])
	})

	t.Run("GetCapabilityIndexStatsOutput", func(t *testing.T) {
		output := GetCapabilityIndexStatsOutput{
			TotalDocuments:     100,
			DocumentsByType:    map[string]int{"persona": 50, "skill": 50},
			UniqueTerms:        1000,
			AverageTermsPerDoc: 10.5,
			IndexHealth:        "healthy",
			LastUpdated:        "2025-12-19",
		}

		assert.Equal(t, 100, output.TotalDocuments)
		assert.Equal(t, 50, output.DocumentsByType["persona"])
		assert.Equal(t, 1000, output.UniqueTerms)
		assert.Equal(t, "healthy", output.IndexHealth)
	})
}
