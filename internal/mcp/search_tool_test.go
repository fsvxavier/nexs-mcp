package mcp

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSearchElements(t *testing.T) {
	// Create enhanced repository with some test data
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewEnhancedFileElementRepository(tmpDir, 10)
	require.NoError(t, err)

	// Create test elements
	persona1 := domain.NewPersona("AI Expert", "Expert in artificial intelligence", "1.0.0", "john")
	metadata1 := persona1.GetMetadata()
	metadata1.Tags = []string{"ai", "expert"}
	persona1.SetMetadata(metadata1)
	require.NoError(t, repo.Create(persona1))

	persona2 := domain.NewPersona("Data Scientist", "Expert in data science", "1.0.0", "jane")
	metadata2 := persona2.GetMetadata()
	metadata2.Tags = []string{"data", "science"}
	persona2.SetMetadata(metadata2)
	require.NoError(t, repo.Create(persona2))

	skill := domain.NewSkill("Machine Learning", "ML algorithms and techniques", "1.0.0", "john")
	metadataSkill := skill.GetMetadata()
	metadataSkill.Tags = []string{"ml", "ai"}
	skill.SetMetadata(metadataSkill)
	require.NoError(t, repo.Create(skill))

	server := NewMCPServer("test", "1.0.0", repo)

	t.Run("Search with query", func(t *testing.T) {
		input := SearchElementsInput{
			Query: "expert",
			Limit: 10,
		}

		result, output, err := server.handleSearchElements(context.Background(), &sdk.CallToolRequest{}, input)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.GreaterOrEqual(t, output.Total, 2) // Should find both personas
		assert.Equal(t, "expert", output.Query)
	})

	t.Run("Search with type filter", func(t *testing.T) {
		personaType := domain.PersonaElement
		input := SearchElementsInput{
			Query: "expert",
			Type:  &personaType,
			Limit: 10,
		}

		_, output, err := server.handleSearchElements(context.Background(), &sdk.CallToolRequest{}, input)
		require.NoError(t, err)
		assert.Equal(t, 2, output.Total)
		for _, result := range output.Results {
			assert.Equal(t, "persona", result.Type)
		}
	})

	t.Run("Search with author filter", func(t *testing.T) {
		author := "john"
		input := SearchElementsInput{
			Query:  "",
			Author: &author,
			Limit:  10,
		}

		_, output, err := server.handleSearchElements(context.Background(), &sdk.CallToolRequest{}, input)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, output.Total, 2) // Should find persona1 and skill
		for _, result := range output.Results {
			assert.Equal(t, "john", result.Author)
		}
	})

	t.Run("Search with tags filter", func(t *testing.T) {
		input := SearchElementsInput{
			Query: "",
			Tags:  []string{"ai"},
			Limit: 10,
		}

		_, output, err := server.handleSearchElements(context.Background(), &sdk.CallToolRequest{}, input)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, output.Total, 2) // persona1 and skill both have "ai" tag
	})

	t.Run("Search with pagination", func(t *testing.T) {
		input := SearchElementsInput{
			Query:  "",
			Limit:  1,
			Offset: 0,
		}

		_, output, err := server.handleSearchElements(context.Background(), &sdk.CallToolRequest{}, input)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(output.Results), 1)
		assert.Equal(t, 1, output.Limit)
		assert.Equal(t, 0, output.Offset)
	})

	t.Run("Search with sort", func(t *testing.T) {
		sortBy := "name"
		sortOrder := "asc"
		input := SearchElementsInput{
			Query:     "",
			Limit:     10,
			SortBy:    &sortBy,
			SortOrder: &sortOrder,
		}

		_, output, err := server.handleSearchElements(context.Background(), &sdk.CallToolRequest{}, input)
		require.NoError(t, err)
		assert.Greater(t, output.Total, 0)

		// Verify ascending order
		for i := 1; i < len(output.Results); i++ {
			assert.LessOrEqual(t, output.Results[i-1].Name, output.Results[i].Name)
		}
	})

	t.Run("Search with maximum limit", func(t *testing.T) {
		input := SearchElementsInput{
			Query: "",
			Limit: 1000, // Should be capped at 500
		}

		_, output, err := server.handleSearchElements(context.Background(), &sdk.CallToolRequest{}, input)
		require.NoError(t, err)
		assert.Equal(t, 500, output.Limit) // Should be capped
	})

	t.Run("Search with default limit", func(t *testing.T) {
		input := SearchElementsInput{
			Query: "",
		}

		_, output, err := server.handleSearchElements(context.Background(), &sdk.CallToolRequest{}, input)
		require.NoError(t, err)
		assert.Equal(t, 50, output.Limit) // Default limit
	})
}

func TestCalculateRelevance(t *testing.T) {
	metadata := domain.ElementMetadata{
		Name:        "Machine Learning Expert",
		Description: "Expert in machine learning and AI",
		Tags:        []string{"ml", "ai", "expert"},
	}

	t.Run("Full match", func(t *testing.T) {
		score := calculateRelevance("machine learning expert", metadata)
		assert.Equal(t, 1.0, score)
	})

	t.Run("Partial match", func(t *testing.T) {
		score := calculateRelevance("machine learning database", metadata)
		assert.Equal(t, 2.0/3.0, score) // 2 out of 3 words match
	})

	t.Run("No match", func(t *testing.T) {
		score := calculateRelevance("database python", metadata)
		assert.Equal(t, 0.0, score)
	})

	t.Run("Empty query", func(t *testing.T) {
		score := calculateRelevance("", metadata)
		assert.Equal(t, 0.0, score)
	})

	t.Run("Case insensitive", func(t *testing.T) {
		score := calculateRelevance("MACHINE LEARNING", metadata)
		assert.Equal(t, 1.0, score)
	})
}

func TestSortResults(t *testing.T) {
	results := []SearchResult{
		{Name: "Zebra", CreatedAt: "2025-01-01 10:00:00", Relevance: 0.5},
		{Name: "Apple", CreatedAt: "2025-01-03 10:00:00", Relevance: 0.9},
		{Name: "Mango", CreatedAt: "2025-01-02 10:00:00", Relevance: 0.7},
	}

	t.Run("Sort by name ascending", func(t *testing.T) {
		sorted := make([]SearchResult, len(results))
		copy(sorted, results)
		order := "asc"
		sortResults(sorted, "name", &order)
		assert.Equal(t, "Apple", sorted[0].Name)
		assert.Equal(t, "Mango", sorted[1].Name)
		assert.Equal(t, "Zebra", sorted[2].Name)
	})

	t.Run("Sort by name descending", func(t *testing.T) {
		sorted := make([]SearchResult, len(results))
		copy(sorted, results)
		order := "desc"
		sortResults(sorted, "name", &order)
		assert.Equal(t, "Zebra", sorted[0].Name)
		assert.Equal(t, "Mango", sorted[1].Name)
		assert.Equal(t, "Apple", sorted[2].Name)
	})

	t.Run("Sort by created_at ascending", func(t *testing.T) {
		sorted := make([]SearchResult, len(results))
		copy(sorted, results)
		order := "asc"
		sortResults(sorted, "created_at", &order)
		assert.Equal(t, "Zebra", sorted[0].Name) // 2025-01-01
		assert.Equal(t, "Mango", sorted[1].Name) // 2025-01-02
		assert.Equal(t, "Apple", sorted[2].Name) // 2025-01-03
	})

	t.Run("Sort by relevance descending", func(t *testing.T) {
		sorted := make([]SearchResult, len(results))
		copy(sorted, results)
		order := "desc"
		sortResults(sorted, "relevance", &order)
		assert.Equal(t, 0.9, sorted[0].Relevance)
		assert.Equal(t, 0.7, sorted[1].Relevance)
		assert.Equal(t, 0.5, sorted[2].Relevance)
	})
}
