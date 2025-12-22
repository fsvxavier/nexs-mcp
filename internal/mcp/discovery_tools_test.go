package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
)

func setupTestServerForDiscovery() *MCPServer {
	repo := NewMockElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

// TestHandleSearchCollections tests

func TestHandleSearchCollections_DefaultValues(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchCollectionsInput{
		Query: "test",
	}

	_, output, err := server.handleSearchCollections(ctx, req, input)
	require.NoError(t, err)
	assert.Equal(t, "test", output.Query)
	assert.NotNil(t, output.Filters)
	assert.NotNil(t, output.Timing)
}

func TestHandleSearchCollections_WithFilters(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchCollectionsInput{
		Query:     "test",
		Category:  "devops",
		Author:    "testauthor",
		Tags:      []string{"tag1", "tag2"},
		MinStars:  10,
		Source:    "github",
		SortBy:    "stars",
		SortOrder: "desc",
		Limit:     10,
		Offset:    5,
	}

	_, output, err := server.handleSearchCollections(ctx, req, input)
	require.NoError(t, err)
	assert.Equal(t, "test", output.Query)
	assert.Equal(t, "devops", output.Filters["category"])
	assert.Equal(t, "testauthor", output.Filters["author"])
	assert.Equal(t, 10, output.Filters["min_stars"])
}

func TestHandleSearchCollections_RichFormat(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchCollectionsInput{
		Query:      "test",
		RichFormat: true,
	}

	_, output, err := server.handleSearchCollections(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Collections)
}

func TestHandleSearchCollections_Timing(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchCollectionsInput{
		Query: "test",
	}

	_, output, err := server.handleSearchCollections(ctx, req, input)
	require.NoError(t, err)
	assert.Contains(t, output.Timing, "search")
	assert.Contains(t, output.Timing, "sort")
	assert.Contains(t, output.Timing, "format")
	assert.Contains(t, output.Timing, "total")
}

// TestSortCollections tests

func TestSortCollections_ByStarsAscending(t *testing.T) {
	collections := []*sources.CollectionMetadata{
		{Name: "C1", Stars: 100},
		{Name: "C2", Stars: 50},
		{Name: "C3", Stars: 200},
	}

	sortCollections(collections, "stars", "asc")

	assert.Equal(t, 50, collections[0].Stars)
	assert.Equal(t, 100, collections[1].Stars)
	assert.Equal(t, 200, collections[2].Stars)
}

func TestSortCollections_ByStarsDescending(t *testing.T) {
	collections := []*sources.CollectionMetadata{
		{Name: "C1", Stars: 100},
		{Name: "C2", Stars: 50},
		{Name: "C3", Stars: 200},
	}

	sortCollections(collections, "stars", "desc")

	assert.Equal(t, 200, collections[0].Stars)
	assert.Equal(t, 100, collections[1].Stars)
	assert.Equal(t, 50, collections[2].Stars)
}

func TestSortCollections_ByDownloads(t *testing.T) {
	collections := []*sources.CollectionMetadata{
		{Name: "C1", Downloads: 1000},
		{Name: "C2", Downloads: 500},
		{Name: "C3", Downloads: 2000},
	}

	sortCollections(collections, "downloads", "desc")

	assert.Equal(t, 2000, collections[0].Downloads)
	assert.Equal(t, 1000, collections[1].Downloads)
	assert.Equal(t, 500, collections[2].Downloads)
}

func TestSortCollections_ByName(t *testing.T) {
	collections := []*sources.CollectionMetadata{
		{Name: "Charlie"},
		{Name: "Alice"},
		{Name: "Bob"},
	}

	sortCollections(collections, "name", "asc")

	assert.Equal(t, "Alice", collections[0].Name)
	assert.Equal(t, "Bob", collections[1].Name)
	assert.Equal(t, "Charlie", collections[2].Name)
}

func TestSortCollections_ByRelevance(t *testing.T) {
	collections := []*sources.CollectionMetadata{
		{Name: "C1", Stars: 100},
		{Name: "C2", Stars: 50},
		{Name: "C3", Stars: 200},
	}

	sortCollections(collections, "relevance", "desc")

	// Relevance currently uses stars as proxy
	assert.Equal(t, 200, collections[0].Stars)
	assert.Equal(t, 100, collections[1].Stars)
	assert.Equal(t, 50, collections[2].Stars)
}

func TestSortCollections_DefaultSortBy(t *testing.T) {
	collections := []*sources.CollectionMetadata{
		{Name: "C1", Stars: 100},
		{Name: "C2", Stars: 50},
	}

	sortCollections(collections, "unknown", "desc")

	// Should default to stars
	assert.Equal(t, 100, collections[0].Stars)
	assert.Equal(t, 50, collections[1].Stars)
}

// TestCollectionMetadataToResult tests

func TestCollectionMetadataToResult_BasicConversion(t *testing.T) {
	meta := &sources.CollectionMetadata{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "testauthor",
		Category:    "devops",
		Description: "Test description",
		Tags:        []string{"tag1", "tag2"},
		Stars:       100,
		Downloads:   1000,
		Repository:  "https://github.com/test/repo",
		SourceName:  "github",
		URI:         "github://test/repo",
	}

	result := collectionMetadataToResult(meta, false)

	assert.Equal(t, "testauthor/test-collection", result.ID)
	assert.Equal(t, "test-collection", result.Name)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Equal(t, "testauthor", result.Author)
	assert.Equal(t, "devops", result.Category)
	assert.Equal(t, "Test description", result.Description)
	assert.Equal(t, []string{"tag1", "tag2"}, result.Tags)
	assert.Equal(t, 100, result.Stars)
	assert.Equal(t, 1000, result.Downloads)
	assert.Equal(t, "https://github.com/test/repo", result.Repository)
	assert.Equal(t, "github", result.Source)
	assert.Equal(t, "github://test/repo", result.URI)
	assert.Empty(t, result.Display)
}

func TestCollectionMetadataToResult_WithRichFormat(t *testing.T) {
	meta := &sources.CollectionMetadata{
		Name:        "test-collection",
		Version:     "1.0.0",
		Author:      "testauthor",
		Category:    "devops",
		Description: "Test description",
		Stars:       100,
	}

	result := collectionMetadataToResult(meta, true)

	assert.NotEmpty(t, result.Display)
	assert.Contains(t, result.Display, "test-collection")
	assert.Contains(t, result.Display, "testauthor")
}

// TestFormatCollectionRich tests

func TestFormatCollectionRich_Complete(t *testing.T) {
	meta := &sources.CollectionMetadata{
		Name:        "awesome-collection",
		Version:     "2.0.0",
		Author:      "john",
		Category:    "devops",
		Description: "An awesome collection",
		Tags:        []string{"docker", "kubernetes"},
		Stars:       250,
		Downloads:   5000,
		Repository:  "https://github.com/john/awesome",
	}

	display := formatCollectionRich(meta)

	assert.Contains(t, display, "awesome-collection")
	assert.Contains(t, display, "v2.0.0")
	assert.Contains(t, display, "john")
	assert.Contains(t, display, "250")
	assert.Contains(t, display, "devops")
	assert.Contains(t, display, "docker, kubernetes")
	assert.Contains(t, display, "An awesome collection")
	assert.Contains(t, display, "https://github.com/john/awesome")
}

func TestFormatCollectionRich_Minimal(t *testing.T) {
	meta := &sources.CollectionMetadata{
		Name:     "minimal",
		Version:  "1.0.0",
		Author:   "test",
		Category: "testing",
	}

	display := formatCollectionRich(meta)

	assert.Contains(t, display, "minimal")
	assert.Contains(t, display, "v1.0.0")
	assert.Contains(t, display, "test")
	assert.Contains(t, display, "testing")
}

// TestGetCategoryEmoji tests

func TestGetCategoryEmoji_KnownCategories(t *testing.T) {
	tests := []struct {
		category string
		expected string
	}{
		{"devops", "⚙️"},
		{"creative-writing", "✍️"},
		{"data-science", "📊"},
		{"web-development", "🌐"},
		{"mobile", "📱"},
		{"ai-ml", "🤖"},
		{"security", "🔒"},
		{"testing", "🧪"},
		{"productivity", "⚡"},
		{"education", "📚"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			emoji := getCategoryEmoji(tt.category)
			assert.Equal(t, tt.expected, emoji)
		})
	}
}

func TestGetCategoryEmoji_UnknownCategory(t *testing.T) {
	emoji := getCategoryEmoji("unknown-category")
	assert.Equal(t, "📦", emoji)
}

func TestGetCategoryEmoji_CaseInsensitive(t *testing.T) {
	emoji1 := getCategoryEmoji("DevOps")
	emoji2 := getCategoryEmoji("devops")
	assert.Equal(t, emoji1, emoji2)
	assert.Equal(t, "⚙️", emoji1)
}

// TestFormatNumber tests

func TestFormatNumber_SmallNumbers(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{100, "100"},
		{999, "999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatNumber_Thousands(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1000, "1.0K"},
		{1500, "1.5K"},
		{10000, "10.0K"},
		{999999, "1000.0K"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatNumber_Millions(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1000000, "1.0M"},
		{1500000, "1.5M"},
		{10000000, "10.0M"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatNumber_Billions(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1000000000, "1.0B"},
		{5500000000, "5.5B"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHandleListCollections tests

func TestHandleListCollections_DefaultLimit(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ListCollectionsInput{}

	_, output, err := server.handleListCollections(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Summary)
	assert.NotNil(t, output.Summary.ByCategory)
	assert.NotNil(t, output.Summary.ByAuthor)
	assert.NotNil(t, output.Summary.BySource)
}

func TestHandleListCollections_WithFilters(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ListCollectionsInput{
		Category: "devops",
		Author:   "testauthor",
		Tags:     []string{"docker"},
		Limit:    10,
		Offset:   5,
	}

	_, output, err := server.handleListCollections(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Collections)
}

func TestHandleListCollections_WithGroupBy(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}

	tests := []struct {
		name    string
		groupBy string
	}{
		{"Group by category", "category"},
		{"Group by author", "author"},
		{"Group by source", "source"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := ListCollectionsInput{
				GroupBy: tt.groupBy,
			}

			_, output, err := server.handleListCollections(ctx, req, input)
			require.NoError(t, err)
			assert.NotNil(t, output.Groups)
			assert.Nil(t, output.Collections)
		})
	}
}

func TestHandleListCollections_WithoutGroupBy(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ListCollectionsInput{}

	_, output, err := server.handleListCollections(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Collections)
	assert.Nil(t, output.Groups)
}

func TestHandleListCollections_Summary(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ListCollectionsInput{}

	_, output, err := server.handleListCollections(ctx, req, input)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, output.Summary.TotalCollections, 0)
	assert.GreaterOrEqual(t, output.Summary.TotalElements, 0)
	assert.GreaterOrEqual(t, output.Summary.TotalDownloads, 0)
	assert.GreaterOrEqual(t, output.Summary.AverageStars, float64(0))
}

func TestHandleListCollections_RichFormat(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ListCollectionsInput{
		RichFormat: true,
	}

	_, output, err := server.handleListCollections(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Collections)
}

// Edge case tests

func TestSearchCollections_EmptyQuery(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchCollectionsInput{}

	_, output, err := server.handleSearchCollections(ctx, req, input)
	require.NoError(t, err)
	assert.Empty(t, output.Query)
}

func TestSearchCollections_LargeLimit(t *testing.T) {
	server := setupTestServerForDiscovery()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := SearchCollectionsInput{
		Query: "test",
		Limit: 1000,
	}

	_, output, err := server.handleSearchCollections(ctx, req, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Collections)
}

func TestCollectionResult_ElementStats(t *testing.T) {
	stats := ElementStats{
		Total:     10,
		Personas:  3,
		Skills:    4,
		Templates: 2,
		Agents:    1,
	}

	assert.Equal(t, 10, stats.Total)
	assert.Equal(t, 3, stats.Personas)
	assert.Equal(t, 4, stats.Skills)
	assert.Equal(t, 2, stats.Templates)
	assert.Equal(t, 1, stats.Agents)
}
