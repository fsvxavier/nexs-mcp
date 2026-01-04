package mcp

import (
	"context"
	"os"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
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

// Mock GitHub client.
type mockClient struct{}

func (m *mockClient) ListRepositories(ctx context.Context) ([]*infrastructure.Repository, error) {
	return nil, nil
}

func (m *mockClient) SearchRepositories(ctx context.Context, query string, options *infrastructure.SearchOptions) (*infrastructure.SearchResult, error) {
	return &infrastructure.SearchResult{
		Repositories: []*infrastructure.Repository{{
			FullName:      "testuser/test-repo",
			URL:           "https://github.com/testuser/test-repo",
			Description:   "Test repo",
			Stars:         10,
			UpdatedAt:     "2025-01-01T00:00:00Z",
			DefaultBranch: "main",
		}},
		TotalCount: 1,
	}, nil
}

func (m *mockClient) GetFile(ctx context.Context, owner, repo, path, branch string) (*infrastructure.FileContent, error) {
	if path == "elements/testuser/template/2025-01-01/template_Test_20250101-120000.yaml" {
		yamlContent := `metadata:
  id: template_Test_20250101-120000
  type: template
  name: Test Template
  description: Test template
  version: 1.0.0
  author: testuser
  tags: ["ai"]
  is_active: true
  created_at: 2025-01-01T00:00:00Z
  updated_at: 2025-01-01T00:00:00Z
`
		return &infrastructure.FileContent{Path: path, Content: yamlContent, SHA: "sha"}, nil
	}
	return nil, nil
}

func (m *mockClient) CreateFile(ctx context.Context, owner, repo, path, message, content, branch string) (*infrastructure.CommitInfo, error) {
	return nil, nil
}

func (m *mockClient) UpdateFile(ctx context.Context, owner, repo, path, message, content, sha, branch string) (*infrastructure.CommitInfo, error) {
	return nil, nil
}

func (m *mockClient) DeleteFile(ctx context.Context, owner, repo, path, message, sha, branch string) error {
	return nil
}

func (m *mockClient) ListFilesInDirectory(ctx context.Context, owner, repo, path, branch string) ([]string, error) {
	return []string{"elements/testuser/template/2025-01-01/template_Test_20250101-120000.yaml"}, nil
}

func (m *mockClient) ListAllFiles(ctx context.Context, owner, repo, branch string) ([]string, error) {
	return m.ListFilesInDirectory(ctx, owner, repo, "", branch)
}
func (m *mockClient) GetUser(ctx context.Context) (string, error) { return "testuser", nil }
func (m *mockClient) CreateRepository(ctx context.Context, name, description string, private bool) (*infrastructure.Repository, error) {
	return nil, nil
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

func TestHandleSearchPortfolioGitHub_ParseRepoContents(t *testing.T) {
	// Unit test that does not require real GitHub access
	server := setupTestServerForGitHubPortfolio()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}

	mock := &mockClient{}
	server.githubClient = mock

	input := SearchPortfolioGitHubInput{Query: "test"}
	_, output, err := server.handleSearchPortfolioGitHub(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.GreaterOrEqual(t, len(output.Results), 1)
	// Expect the first repo to have one element found
	assert.GreaterOrEqual(t, len(output.Results[0].ElementsFound), 1)
	match := output.Results[0].ElementsFound[0]
	assert.Equal(t, "template_Test_20250101-120000", match.ID)
	assert.Equal(t, "template", match.Type)
	assert.Equal(t, "elements/testuser/template/2025-01-01/template_Test_20250101-120000.yaml", match.FilePath)
}
