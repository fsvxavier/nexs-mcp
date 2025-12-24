package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// Default sort order for portfolio search.
	defaultSortRelevance = "relevance"
)

// SearchPortfolioGitHubInput represents the input for search_portfolio_github tool.
type SearchPortfolioGitHubInput struct {
	Query           string   `json:"query"`
	ElementType     string   `json:"element_type,omitempty"`
	Author          string   `json:"author,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	SortBy          string   `json:"sort_by,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	IncludeArchived bool     `json:"include_archived,omitempty"`
}

// ElementMatch represents an element found in a portfolio.
type ElementMatch struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Author      string   `json:"author,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	FilePath    string   `json:"file_path"`
}

// PortfolioSearchResult represents a portfolio repository match.
type PortfolioSearchResult struct {
	RepoName      string         `json:"repo_name"`
	RepoURL       string         `json:"repo_url"`
	Description   string         `json:"description,omitempty"`
	Stars         int            `json:"stars"`
	UpdatedAt     string         `json:"updated_at"`
	ElementsFound []ElementMatch `json:"elements_found,omitempty"`
	MatchScore    float64        `json:"match_score"`
}

// SearchPortfolioGitHubOutput represents the output of search_portfolio_github tool.
type SearchPortfolioGitHubOutput struct {
	Results      []PortfolioSearchResult `json:"results"`
	TotalCount   int                     `json:"total_count"`
	Page         int                     `json:"page"`
	HasMore      bool                    `json:"has_more"`
	SearchTimeMs int64                   `json:"search_time_ms"`
}

// handleSearchPortfolioGitHub handles search_portfolio_github tool calls.
func (s *MCPServer) handleSearchPortfolioGitHub(ctx context.Context, req *sdk.CallToolRequest, input SearchPortfolioGitHubInput) (*sdk.CallToolResult, SearchPortfolioGitHubOutput, error) {
	startTime := time.Now()

	// Validate required inputs
	if input.Query == "" {
		return nil, SearchPortfolioGitHubOutput{}, errors.New("query is required")
	}

	// Set defaults
	if input.ElementType == "" {
		input.ElementType = "all"
	}
	if input.SortBy == "" {
		input.SortBy = defaultSortRelevance
	}
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	// Validate element_type
	validTypes := map[string]bool{
		"all": true, "persona": true, "skill": true,
		"template": true, "agent": true, "memory": true, "ensemble": true,
	}
	if !validTypes[input.ElementType] {
		return nil, SearchPortfolioGitHubOutput{}, fmt.Errorf("invalid element_type: %s", input.ElementType)
	}

	// Validate sort_by
	validSortBy := map[string]bool{
		"stars": true, "updated": true, "created": true, defaultSortRelevance: true,
	}
	if !validSortBy[input.SortBy] {
		return nil, SearchPortfolioGitHubOutput{}, fmt.Errorf("invalid sort_by: %s", input.SortBy)
	}

	// Note: This is a placeholder implementation
	// In a real implementation, we would:
	// 1. Check if GitHub client is authenticated
	// 2. Use s.githubClient to search repositories
	// 3. Search for repos with "nexs-portfolio" topic or "nexs-mcp" in description
	// 4. Parse repository contents to find elements matching the query
	// 5. Score and rank results by relevance
	// 6. Apply filters (author, tags, element_type)
	// 7. Sort by requested sort_by field

	// Get GitHub client if available
	githubOAuthClient, err := s.getGitHubOAuthClient()
	if err != nil {
		return nil, SearchPortfolioGitHubOutput{}, fmt.Errorf("GitHub OAuth not configured: %w", err)
	}

	githubClient := infrastructure.NewGitHubClient(githubOAuthClient)

	// Build search query
	searchQuery := input.Query + " topic:nexs-portfolio"
	if input.Author != "" {
		searchQuery += " user:" + input.Author
	}

	// Search repositories
	searchOpts := &infrastructure.SearchOptions{
		SortBy:          input.SortBy,
		IncludeArchived: input.IncludeArchived,
		Limit:           input.Limit,
	}

	searchResult, err := githubClient.SearchRepositories(ctx, searchQuery, searchOpts)
	if err != nil {
		return nil, SearchPortfolioGitHubOutput{}, fmt.Errorf("GitHub search failed: %w", err)
	}

	// Convert repositories to results
	results := make([]PortfolioSearchResult, 0, len(searchResult.Repositories))
	for _, repo := range searchResult.Repositories {
		// TODO: Parse repository contents to find actual NEXS elements
		// For now, return repository info only
		result := PortfolioSearchResult{
			RepoName:      repo.FullName,
			RepoURL:       repo.URL,
			Description:   repo.Description,
			Stars:         repo.Stars,
			UpdatedAt:     repo.UpdatedAt,
			ElementsFound: []ElementMatch{}, // Would be populated by parsing repo
			MatchScore:    1.0,              // Would be calculated based on relevance
		}

		// Filter by element type if specified and not "all"
		// In a complete implementation, we would check repo contents
		if input.ElementType == "all" || input.ElementType == "" {
			results = append(results, result)
		}
	}

	// Calculate search time
	searchTime := time.Since(startTime).Milliseconds()

	output := SearchPortfolioGitHubOutput{
		Results:      results,
		TotalCount:   searchResult.TotalCount,
		Page:         1,
		HasMore:      len(results) < searchResult.TotalCount,
		SearchTimeMs: searchTime,
	}

	return nil, output, nil
}
