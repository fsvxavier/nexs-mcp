package sources

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// MockOAuthProvider is a mock implementation of OAuthProvider for testing
type MockOAuthProvider struct {
	token *oauth2.Token
	err   error
}

func (m *MockOAuthProvider) GetToken(ctx context.Context) (*oauth2.Token, error) {
	return m.token, m.err
}

func TestGitHubSource_Name(t *testing.T) {
	source, err := NewGitHubSource(nil, "")
	require.NoError(t, err)
	assert.Equal(t, "github", source.Name())
}

func TestGitHubSource_Supports(t *testing.T) {
	source, err := NewGitHubSource(nil, "")
	require.NoError(t, err)

	tests := []struct {
		name     string
		uri      string
		expected bool
	}{
		{
			name:     "valid github URI",
			uri:      "github://fsvxavier/nexs-mcp-collection",
			expected: true,
		},
		{
			name:     "valid github URI with version",
			uri:      "github://fsvxavier/nexs-mcp-collection@1.0.0",
			expected: true,
		},
		{
			name:     "file URI",
			uri:      "file:///path/to/collection",
			expected: false,
		},
		{
			name:     "https URI",
			uri:      "https://example.com/collection.tar.gz",
			expected: false,
		},
		{
			name:     "invalid URI",
			uri:      "invalid://uri",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := source.Supports(tt.uri)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitHubSource_ParseURI(t *testing.T) {
	source, err := NewGitHubSource(nil, "")
	require.NoError(t, err)

	tests := []struct {
		name          string
		uri           string
		expectedOwner string
		expectedRepo  string
		expectedVer   string
		expectError   bool
	}{
		{
			name:          "valid URI without version",
			uri:           "github://fsvxavier/nexs-mcp-collection",
			expectedOwner: "fsvxavier",
			expectedRepo:  "nexs-mcp-collection",
			expectedVer:   "",
			expectError:   false,
		},
		{
			name:          "valid URI with version",
			uri:           "github://fsvxavier/nexs-mcp-collection@1.0.0",
			expectedOwner: "fsvxavier",
			expectedRepo:  "nexs-mcp-collection",
			expectedVer:   "1.0.0",
			expectError:   false,
		},
		{
			name:          "valid URI with version range",
			uri:           "github://user/repo@^2.1.0",
			expectedOwner: "user",
			expectedRepo:  "repo",
			expectedVer:   "^2.1.0",
			expectError:   false,
		},
		{
			name:        "invalid URI - wrong scheme",
			uri:         "https://github.com/user/repo",
			expectError: true,
		},
		{
			name:        "invalid URI - missing repo",
			uri:         "github://user",
			expectError: true,
		},
		{
			name:        "invalid URI - too many parts",
			uri:         "github://user/repo/extra",
			expectError: true,
		},
		{
			name:        "invalid URI - multiple @ symbols",
			uri:         "github://user/repo@1.0.0@2.0.0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, version, err := source.parseURI(tt.uri)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOwner, owner)
				assert.Equal(t, tt.expectedRepo, repo)
				assert.Equal(t, tt.expectedVer, version)
			}
		})
	}
}

func TestGitHubSource_BuildSearchQuery(t *testing.T) {
	source, err := NewGitHubSource(nil, "")
	require.NoError(t, err)

	tests := []struct {
		name     string
		filter   *BrowseFilter
		expected string
	}{
		{
			name:     "nil filter",
			filter:   nil,
			expected: "topic:nexs-mcp-collection",
		},
		{
			name:     "empty filter",
			filter:   &BrowseFilter{},
			expected: "topic:nexs-mcp-collection",
		},
		{
			name: "with author",
			filter: &BrowseFilter{
				Author: "fsvxavier",
			},
			expected: "topic:nexs-mcp-collection user:fsvxavier",
		},
		{
			name: "with query",
			filter: &BrowseFilter{
				Query: "devops",
			},
			expected: "topic:nexs-mcp-collection devops in:name,description,readme",
		},
		{
			name: "with category",
			filter: &BrowseFilter{
				Category: "development",
			},
			expected: "topic:nexs-mcp-collection topic:development",
		},
		{
			name: "with tags",
			filter: &BrowseFilter{
				Tags: []string{"docker", "kubernetes"},
			},
			expected: "topic:nexs-mcp-collection topic:docker topic:kubernetes",
		},
		{
			name: "with all filters",
			filter: &BrowseFilter{
				Author:   "fsvxavier",
				Query:    "cloud",
				Category: "devops",
				Tags:     []string{"aws", "terraform"},
			},
			expected: "topic:nexs-mcp-collection user:fsvxavier cloud in:name,description,readme topic:devops topic:aws topic:terraform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := source.buildSearchQuery(tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitHubSource_CompareVersions(t *testing.T) {
	source, err := NewGitHubSource(nil, "")
	require.NoError(t, err)

	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "equal versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "v1 > v2",
			v1:       "2.0.0",
			v2:       "1.0.0",
			expected: 1,
		},
		{
			name:     "v1 < v2",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := source.compareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitHubSource_CacheDirectory(t *testing.T) {
	t.Run("custom cache directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheDir := filepath.Join(tmpDir, "custom-cache")

		source, err := NewGitHubSource(nil, cacheDir)
		require.NoError(t, err)
		assert.Equal(t, cacheDir, source.cacheDir)

		// Verify directory was created
		_, err = os.Stat(cacheDir)
		assert.NoError(t, err)
	})

	t.Run("default cache directory", func(t *testing.T) {
		source, err := NewGitHubSource(nil, "")
		require.NoError(t, err)

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expectedDir := filepath.Join(homeDir, ".nexs", "cache", "collections")
		assert.Equal(t, expectedDir, source.cacheDir)
	})
}

func TestGitHubSource_GetClient(t *testing.T) {
	ctx := context.Background()

	t.Run("with oauth client", func(t *testing.T) {
		mockOAuth := &MockOAuthProvider{
			token: &oauth2.Token{AccessToken: "test-token"},
		}

		source, err := NewGitHubSource(mockOAuth, "")
		require.NoError(t, err)

		client, err := source.getClient(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("without oauth client", func(t *testing.T) {
		source, err := NewGitHubSource(nil, "")
		require.NoError(t, err)

		client, err := source.getClient(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with oauth error", func(t *testing.T) {
		mockOAuth := &MockOAuthProvider{
			err: fmt.Errorf("oauth error"),
		}

		source, err := NewGitHubSource(mockOAuth, "")
		require.NoError(t, err)

		// Should fall back to unauthenticated client
		client, err := source.getClient(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("client caching", func(t *testing.T) {
		source, err := NewGitHubSource(nil, "")
		require.NoError(t, err)

		client1, err := source.getClient(ctx)
		require.NoError(t, err)

		client2, err := source.getClient(ctx)
		require.NoError(t, err)

		// Should return the same client instance
		assert.Same(t, client1, client2)
	})
}

// Integration test that requires git to be installed
func TestGitHubSource_Integration(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available, skipping integration test")
	}

	// Skip if GITHUB_TOKEN is not set (for CI/CD)
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("GITHUB_TOKEN not set, skipping integration test")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	mockOAuth := &MockOAuthProvider{
		token: &oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	}

	source, err := NewGitHubSource(mockOAuth, tmpDir)
	require.NoError(t, err)

	t.Run("browse collections", func(t *testing.T) {
		filter := &BrowseFilter{
			Author: "fsvxavier",
			Limit:  10,
		}

		collections, err := source.Browse(ctx, filter)

		// If this fails due to no collections found, that's acceptable
		if err == nil {
			assert.NotNil(t, collections)
			// Verify structure of returned collections
			for _, col := range collections {
				assert.NotEmpty(t, col.Name)
				assert.NotEmpty(t, col.Author)
				assert.Equal(t, "github", col.SourceName)
			}
		}
	})

	// Note: We can't test Get() without a real collection repository
	// This would require either:
	// 1. A test repository in GitHub
	// 2. Mocking the entire GitHub API
	// For now, we'll test the URI parsing and validation logic above
}
