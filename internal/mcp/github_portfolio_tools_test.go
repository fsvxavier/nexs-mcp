package mcp

import (
	"context"
	"os"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServerForGitHubPortfolio() *MCPServer {
	repo := NewMockElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

// skipIfNoGitHubToken skips the test if GitHub token is not configured.
func skipIfNoGitHubToken(t *testing.T) {
	// Check if token file exists or if GITHUB_TOKEN env is set
	tokenPath := "data/github_token.json"
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) && os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("Skipping test: GitHub OAuth token not configured. Set GITHUB_TOKEN env var or authenticate with GitHub.")
	}
}

func TestHandleSearchPortfolioGitHub_RequiredQuery(t *testing.T) {
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{} // Missing query

	_, _, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query is required")
}

func TestHandleSearchPortfolioGitHub_ValidQuery(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query: "test query",
	}

	result, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.Nil(t, result)
	assert.NotNil(t, output.Results)
}

func TestHandleSearchPortfolioGitHub_DefaultElementType(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query: "test",
	}

	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output)
}

func TestHandleSearchPortfolioGitHub_DefaultSortBy(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query: "test",
	}

	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output)
}

func TestHandleSearchPortfolioGitHub_LimitDefaults(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{"zero limit", 0},
		{"negative limit", -1},
		{"default applied", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipIfNoGitHubToken(t)
			server := setupTestServerForGitHubPortfolio()
			ctx := context.Background()
			req := &sdk.CallToolRequest{}
			input := SearchPortfolioGitHubInput{
				Query: "test",
				Limit: tt.limit,
			}

			_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
			require.NoError(t, err)
			assert.NotNil(t, output)
		})
	}
}

func TestHandleSearchPortfolioGitHub_LimitMaxCap(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query: "test",
		Limit: 200, // Exceeds max of 100
	}

	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output)
}

func TestHandleSearchPortfolioGitHub_InvalidElementType(t *testing.T) {
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query:       "test",
		ElementType: "invalid_type",
	}

	_, _, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid element_type")
}

func TestHandleSearchPortfolioGitHub_ValidElementTypes(t *testing.T) {
	elementTypes := []string{"all", "persona", "skill", "template", "agent", "memory", "ensemble"}

	for _, elemType := range elementTypes {
		t.Run(elemType, func(t *testing.T) {
			skipIfNoGitHubToken(t)
			server := setupTestServerForGitHubPortfolio()
			ctx := context.Background()
			req := &sdk.CallToolRequest{}
			input := SearchPortfolioGitHubInput{
				Query:       "test",
				ElementType: elemType,
			}

			_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
			require.NoError(t, err)
			assert.NotNil(t, output)
		})
	}
}

func TestHandleSearchPortfolioGitHub_InvalidSortBy(t *testing.T) {
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query:  "test",
		SortBy: "invalid_sort",
	}

	_, _, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort_by")
}

func TestHandleSearchPortfolioGitHub_ValidSortByOptions(t *testing.T) {
	sortOptions := []string{"stars", "updated", "created", "relevance"}

	for _, sortBy := range sortOptions {
		t.Run(sortBy, func(t *testing.T) {
			skipIfNoGitHubToken(t)
			server := setupTestServerForGitHubPortfolio()
			ctx := context.Background()
			req := &sdk.CallToolRequest{}
			input := SearchPortfolioGitHubInput{
				Query:  "test",
				SortBy: sortBy,
			}

			_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
			require.NoError(t, err)
			assert.NotNil(t, output)
		})
	}
}

func TestHandleSearchPortfolioGitHub_WithTags(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query: "test",
		Tags:  []string{"tag1", "tag2"},
	}

	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output)
}

func TestHandleSearchPortfolioGitHub_WithAuthor(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query:  "test",
		Author: "test_author",
	}

	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output)
}

func TestHandleSearchPortfolioGitHub_IncludeArchived(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query:           "test",
		IncludeArchived: true,
	}

	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output)
}

func TestHandleSearchPortfolioGitHub_OutputStructure(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query: "test",
	}

	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)

	// Verify output structure
	assert.NotNil(t, output.Results)
	assert.GreaterOrEqual(t, output.TotalCount, 0)
	assert.GreaterOrEqual(t, output.Page, 0)
	assert.IsType(t, false, output.HasMore)
	assert.GreaterOrEqual(t, output.SearchTimeMs, int64(0))
}

func TestHandleSearchPortfolioGitHub_SearchTime(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query: "test",
	}

	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)

	// Search time should be recorded
	assert.GreaterOrEqual(t, output.SearchTimeMs, int64(0))
}

func TestHandleSearchPortfolioGitHub_NilResult(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query: "test",
	}

	result, _, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestHandleSearchPortfolioGitHub_CompleteInput(t *testing.T) {
	skipIfNoGitHubToken(t)
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchPortfolioGitHubInput{
		Query:           "test query",
		ElementType:     "persona",
		Author:          "author123",
		Tags:            []string{"ai", "automation"},
		SortBy:          "stars",
		Limit:           50,
		IncludeArchived: true,
	}

	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.GreaterOrEqual(t, output.SearchTimeMs, int64(0))
}
