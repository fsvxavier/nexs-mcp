package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
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

	// Get GitHub client if available (prefer injected mock for tests)
	var githubClient infrastructure.GitHubClientInterface
	if s.githubClient != nil {
		githubClient = s.githubClient
	} else {
		githubOAuthClient, err := s.getGitHubOAuthClient()
		if err != nil {
			return nil, SearchPortfolioGitHubOutput{}, fmt.Errorf("GitHub OAuth not configured: %w", err)
		}
		githubClient = infrastructure.NewGitHubClient(githubOAuthClient)
	}

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
		// Parse repository contents to find NEXS elements (extracted to helper to reduce complexity)
		elements, perr := s.parseRepoForElements(ctx, githubClient, repo, input)
		if perr != nil {
			logger.Error("Failed to parse repository contents", "repo", repo.FullName, "error", perr)
		}

		result := PortfolioSearchResult{
			RepoName:      repo.FullName,
			RepoURL:       repo.URL,
			Description:   repo.Description,
			Stars:         repo.Stars,
			UpdatedAt:     repo.UpdatedAt,
			ElementsFound: elements,
			MatchScore:    1.0, // Basic scoring for now
		}

		// Only include repo if element type is 'all' or elements were found
		if input.ElementType == "all" || len(elements) > 0 {
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

// parseRepoForElements inspects repository files and extracts elements that match filters.
func (s *MCPServer) parseRepoForElements(ctx context.Context, githubClient infrastructure.GitHubClientInterface, repo *infrastructure.Repository, input SearchPortfolioGitHubInput) ([]ElementMatch, error) {
	elements := make([]ElementMatch, 0)

	owner, repoName, perr := infrastructure.ParseRepoURL(repo.URL)
	if perr != nil {
		return elements, perr
	}

	branch := repo.DefaultBranch
	if branch == "" {
		branch = "main"
	}

	files, ferr := githubClient.ListAllFiles(ctx, owner, repoName, branch)
	if ferr != nil {
		return elements, ferr
	}

	for _, fpath := range files {
		lf := strings.ToLower(fpath)
		if !strings.HasSuffix(lf, ".yaml") && !strings.HasSuffix(lf, ".yml") && !strings.HasSuffix(lf, ".json") {
			continue
		}

		fileContent, gerr := githubClient.GetFile(ctx, owner, repoName, fpath, branch)
		if gerr != nil || fileContent == nil {
			continue
		}

		var stored infrastructure.StoredElement
		if err := yaml.Unmarshal([]byte(fileContent.Content), &stored); err != nil {
			continue
		}

		// Check element type filter
		elemType := string(stored.Metadata.Type)
		if input.ElementType != "all" && input.ElementType != "" && input.ElementType != elemType {
			continue
		}

		// Tags filter (all tags must be present)
		if len(input.Tags) > 0 {
			ok := true
			for _, t := range input.Tags {
				found := false
				for _, st := range stored.Metadata.Tags {
					if t == st {
						found = true
						break
					}
				}
				if !found {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
		}

		// Author filter
		if input.Author != "" && stored.Metadata.Author != input.Author {
			continue
		}

		match := ElementMatch{
			ID:          stored.Metadata.ID,
			Name:        stored.Metadata.Name,
			Type:        elemType,
			Description: stored.Metadata.Description,
			Version:     stored.Metadata.Version,
			Author:      stored.Metadata.Author,
			Tags:        stored.Metadata.Tags,
			FilePath:    fileContent.Path,
		}
		elements = append(elements, match)
	}

	return elements, nil
}
