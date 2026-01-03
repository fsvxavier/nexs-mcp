package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/common"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchElementsInput defines input parameters for search_elements tool.
type SearchElementsInput struct {
	Query     string              `json:"query"`
	Type      *domain.ElementType `json:"type,omitempty"`
	Tags      []string            `json:"tags,omitempty"`
	Author    *string             `json:"author,omitempty"`
	IsActive  *bool               `json:"is_active,omitempty"`
	DateFrom  *string             `json:"date_from,omitempty"`
	DateTo    *string             `json:"date_to,omitempty"`
	Limit     int                 `json:"limit,omitempty"`
	Offset    int                 `json:"offset,omitempty"`
	SortBy    *string             `json:"sort_by,omitempty"`
	SortOrder *string             `json:"sort_order,omitempty"`
	User      string              `json:"user,omitempty"       jsonschema:"authenticated username for access control (optional)"`
}

// SearchElementsOutput defines the output structure for search results.
type SearchElementsOutput struct {
	Results    []SearchResult `json:"results"`
	Total      int            `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	Query      string         `json:"query"`
	FilteredBy map[string]any `json:"filtered_by,omitempty"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	Tags        []string `json:"tags"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	Relevance   float64  `json:"relevance,omitempty"`
}

// handleSearchElements implements the search_elements MCP tool.
func (s *MCPServer) handleSearchElements(ctx context.Context, req *sdk.CallToolRequest, input SearchElementsInput) (*sdk.CallToolResult, SearchElementsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "search_elements",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	// Set defaults
	if input.Limit == 0 {
		input.Limit = 50
	}
	if input.Limit > 500 {
		input.Limit = 500 // Maximum limit
	}

	// Build filter
	filter := domain.ElementFilter{
		Type:     input.Type,
		Tags:     input.Tags,
		IsActive: input.IsActive,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}

	// Check if repository supports search
	var results []domain.Element
	var err error

	// Try to use enhanced repository with full-text search
	if enhancedRepo, ok := s.repo.(*infrastructure.EnhancedFileElementRepository); ok && input.Query != "" {
		results, err = enhancedRepo.Search(input.Query, filter)
	} else {
		// Fallback to regular list
		results, err = s.repo.List(filter)
	}

	if err != nil {
		handlerErr = fmt.Errorf("failed to search elements: %w", err)
		return nil, SearchElementsOutput{}, handlerErr
	}

	// Filter by author if specified
	if input.Author != nil && *input.Author != "" {
		var filtered []domain.Element
		for _, elem := range results {
			if elem.GetMetadata().Author == *input.Author {
				filtered = append(filtered, elem)
			}
		}
		results = filtered
	}

	// Filter by date range if specified
	if input.DateFrom != nil || input.DateTo != nil {
		var filtered []domain.Element
		for _, elem := range results {
			createdDate := elem.GetMetadata().CreatedAt.Format("2006-01-02")

			if input.DateFrom != nil && createdDate < *input.DateFrom {
				continue
			}

			if input.DateTo != nil && createdDate > *input.DateTo {
				continue
			}

			filtered = append(filtered, elem)
		}
		results = filtered
	}

	// Apply access control filtering
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()
	results = accessControl.FilterByPermissions(userCtx, results)

	// Calculate relevance scores (simple word matching)
	searchResults := make([]SearchResult, 0, len(results))
	for _, elem := range results {
		metadata := elem.GetMetadata()

		relevance := 0.0
		if input.Query != "" {
			relevance = calculateRelevance(input.Query, metadata)
		}

		searchResults = append(searchResults, SearchResult{
			ID:          metadata.ID,
			Type:        string(metadata.Type),
			Name:        metadata.Name,
			Description: metadata.Description,
			Author:      metadata.Author,
			Version:     metadata.Version,
			Tags:        metadata.Tags,
			IsActive:    metadata.IsActive,
			CreatedAt:   metadata.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   metadata.UpdatedAt.Format("2006-01-02 15:04:05"),
			Relevance:   relevance,
		})
	}

	// Sort results
	if input.SortBy != nil {
		sortResults(searchResults, *input.SortBy, input.SortOrder)
	}

	// Build filtered_by map
	filteredBy := make(map[string]any)
	if input.Type != nil {
		filteredBy["type"] = string(*input.Type)
	}
	if len(input.Tags) > 0 {
		filteredBy["tags"] = input.Tags
	}
	if input.Author != nil {
		filteredBy["author"] = *input.Author
	}
	if input.IsActive != nil {
		filteredBy["is_active"] = *input.IsActive
	}
	if input.DateFrom != nil {
		filteredBy["date_from"] = *input.DateFrom
	}
	if input.DateTo != nil {
		filteredBy["date_to"] = *input.DateTo
	}

	output := SearchElementsOutput{
		Results:    searchResults,
		Total:      len(searchResults),
		Limit:      input.Limit,
		Offset:     input.Offset,
		Query:      input.Query,
		FilteredBy: filteredBy,
	}

	// Record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "search_elements", output)

	return nil, output, nil
}

// calculateRelevance calculates a simple relevance score based on word matching.
func calculateRelevance(query string, metadata domain.ElementMetadata) float64 {
	queryWords := strings.Fields(strings.ToLower(query))
	if len(queryWords) == 0 {
		return 0.0
	}

	// Build searchable text
	searchText := strings.ToLower(fmt.Sprintf("%s %s %s",
		metadata.Name,
		metadata.Description,
		strings.Join(metadata.Tags, " "),
	))

	matchCount := 0
	for _, word := range queryWords {
		if strings.Contains(searchText, word) {
			matchCount++
		}
	}

	return float64(matchCount) / float64(len(queryWords))
}

// sortResults sorts search results based on the specified field and order.
func sortResults(results []SearchResult, sortBy string, sortOrder *string) {
	order := common.SortOrderDesc
	if sortOrder != nil {
		order = strings.ToLower(*sortOrder)
	}

	// Simple bubble sort for demonstration (use sort.Slice for production)
	for i := range len(results) - 1 {
		for j := i + 1; j < len(results); j++ {
			swap := false

			switch sortBy {
			case "name":
				if order == common.SortOrderAsc {
					swap = results[i].Name > results[j].Name
				} else {
					swap = results[i].Name < results[j].Name
				}
			case "created_at":
				if order == common.SortOrderAsc {
					swap = results[i].CreatedAt > results[j].CreatedAt
				} else {
					swap = results[i].CreatedAt < results[j].CreatedAt
				}
			case "updated_at":
				if order == common.SortOrderAsc {
					swap = results[i].UpdatedAt > results[j].UpdatedAt
				} else {
					swap = results[i].UpdatedAt < results[j].UpdatedAt
				}
			case "relevance":
				if order == common.SortOrderAsc {
					swap = results[i].Relevance > results[j].Relevance
				} else {
					swap = results[i].Relevance < results[j].Relevance
				}
			}

			if swap {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
