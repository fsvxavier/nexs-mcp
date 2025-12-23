package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
)

// --- Input/Output structures for index tools ---

// SearchCapabilityIndexInput defines input for search_capability_index tool.
type SearchCapabilityIndexInput struct {
	Query      string   `json:"query"                 jsonschema:"search query for capabilities"`
	MaxResults int      `json:"max_results,omitempty" jsonschema:"maximum number of results (default: 10)"`
	Types      []string `json:"types,omitempty"       jsonschema:"filter by element types (persona, skill, template, etc)"`
	User       string   `json:"user,omitempty"        jsonschema:"authenticated username for access control (optional)"`
}

// SearchCapabilityIndexOutput defines output for search_capability_index tool.
type SearchCapabilityIndexOutput struct {
	Results []SearchResultItem `json:"results" jsonschema:"search results with scores and highlights"`
	Query   string             `json:"query"   jsonschema:"original search query"`
	Total   int                `json:"total"   jsonschema:"total number of results"`
}

// SearchResultItem represents a single search result.
type SearchResultItem struct {
	DocumentID string   `json:"document_id" jsonschema:"element ID"`
	Type       string   `json:"type"        jsonschema:"element type"`
	Name       string   `json:"name"        jsonschema:"element name"`
	Score      float64  `json:"score"       jsonschema:"relevance score (0-1)"`
	Highlights []string `json:"highlights"  jsonschema:"relevant text snippets"`
}

// FindSimilarCapabilitiesInput defines input for find_similar_capabilities tool.
type FindSimilarCapabilitiesInput struct {
	ElementID  string `json:"element_id"            jsonschema:"element ID to find similar capabilities for"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"maximum number of results (default: 5)"`
	User       string `json:"user,omitempty"        jsonschema:"authenticated username for access control (optional)"`
}

// FindSimilarCapabilitiesOutput defines output for find_similar_capabilities tool.
type FindSimilarCapabilitiesOutput struct {
	Similar   []SimilarCapabilityItem `json:"similar"    jsonschema:"similar capabilities"`
	ElementID string                  `json:"element_id" jsonschema:"original element ID"`
	Total     int                     `json:"total"      jsonschema:"total number of similar items"`
}

// SimilarCapabilityItem represents a similar capability.
type SimilarCapabilityItem struct {
	DocumentID string  `json:"document_id" jsonschema:"similar element ID"`
	Type       string  `json:"type"        jsonschema:"element type"`
	Name       string  `json:"name"        jsonschema:"element name"`
	Similarity float64 `json:"similarity"  jsonschema:"similarity score (0-1)"`
}

// MapCapabilityRelationshipsInput defines input for map_capability_relationships tool.
type MapCapabilityRelationshipsInput struct {
	ElementID string  `json:"element_id"          jsonschema:"element ID to map relationships for"`
	Threshold float64 `json:"threshold,omitempty" jsonschema:"minimum similarity threshold (0-1, default: 0.3)"`
	User      string  `json:"user,omitempty"      jsonschema:"authenticated username for access control (optional)"`
}

// MapCapabilityRelationshipsOutput defines output for map_capability_relationships tool.
type MapCapabilityRelationshipsOutput struct {
	ElementID     string                 `json:"element_id"    jsonschema:"original element ID"`
	Relationships []RelationshipItem     `json:"relationships" jsonschema:"capability relationships"`
	Graph         map[string]interface{} `json:"graph"         jsonschema:"relationship graph data"`
}

// RelationshipItem represents a capability relationship.
type RelationshipItem struct {
	TargetID         string  `json:"target_id"         jsonschema:"related element ID"`
	TargetType       string  `json:"target_type"       jsonschema:"related element type"`
	TargetName       string  `json:"target_name"       jsonschema:"related element name"`
	Similarity       float64 `json:"similarity"        jsonschema:"similarity score"`
	RelationshipType string  `json:"relationship_type" jsonschema:"type of relationship (complementary, similar, related)"`
}

// GetCapabilityIndexStatsInput defines input for get_capability_index_stats tool.
type GetCapabilityIndexStatsInput struct {
	User string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// GetCapabilityIndexStatsOutput defines output for get_capability_index_stats tool.
type GetCapabilityIndexStatsOutput struct {
	TotalDocuments     int            `json:"total_documents"       jsonschema:"total indexed documents"`
	DocumentsByType    map[string]int `json:"documents_by_type"     jsonschema:"documents grouped by type"`
	UniqueTerms        int            `json:"unique_terms"          jsonschema:"total unique terms in index"`
	AverageTermsPerDoc float64        `json:"average_terms_per_doc" jsonschema:"average terms per document"`
	IndexHealth        string         `json:"index_health"          jsonschema:"index health status (healthy, degraded, empty)"`
	LastUpdated        string         `json:"last_updated"          jsonschema:"last index update time"`
}

// --- Tool implementations ---

// handleSearchCapabilityIndex handles the search_capability_index tool call.
func (s *MCPServer) handleSearchCapabilityIndex(ctx context.Context, req *sdk.CallToolRequest, input SearchCapabilityIndexInput) (*sdk.CallToolResult, SearchCapabilityIndexOutput, error) {
	// Validate input
	if strings.TrimSpace(input.Query) == "" {
		return nil, SearchCapabilityIndexOutput{}, errors.New("query cannot be empty")
	}

	// Set default max results
	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 100 {
		maxResults = 100
	}

	// Perform search using HNSW-backed hybrid search
	searchResults, err := s.hybridSearch.Search(ctx, input.Query, maxResults, nil)
	if err != nil {
		return nil, SearchCapabilityIndexOutput{}, fmt.Errorf("search failed: %w", err)
	}

	// Filter by types if specified
	var filteredResults []embeddings.Result
	if len(input.Types) > 0 {
		typeMap := make(map[domain.ElementType]bool)
		for _, t := range input.Types {
			typeMap[domain.ElementType(t)] = true
		}

		for _, result := range searchResults {
			if typeStr, ok := result.Metadata["type"].(string); ok {
				if typeMap[domain.ElementType(typeStr)] {
					filteredResults = append(filteredResults, result)
				}
			}
		}
	} else {
		filteredResults = searchResults
	}

	// Limit results
	if len(filteredResults) > maxResults {
		filteredResults = filteredResults[:maxResults]
	}

	// Convert to output format
	results := make([]SearchResultItem, len(filteredResults))
	for i, r := range filteredResults {
		name := "Unknown"
		if nameVal, ok := r.Metadata["name"].(string); ok {
			name = nameVal
		}
		typeStr := "unknown"
		if typeVal, ok := r.Metadata["type"].(string); ok {
			typeStr = typeVal
		}
		results[i] = SearchResultItem{
			DocumentID: r.ID,
			Type:       typeStr,
			Name:       name,
			Score:      r.Score,
			Highlights: []string{}, // HNSW doesn't provide highlights
		}
	}

	output := SearchCapabilityIndexOutput{
		Results: results,
		Query:   input.Query,
		Total:   len(results),
	}

	return nil, output, nil
}

// handleFindSimilarCapabilities handles the find_similar_capabilities tool call.
func (s *MCPServer) handleFindSimilarCapabilities(ctx context.Context, req *sdk.CallToolRequest, input FindSimilarCapabilitiesInput) (*sdk.CallToolResult, FindSimilarCapabilitiesOutput, error) {
	// Validate input
	if strings.TrimSpace(input.ElementID) == "" {
		return nil, FindSimilarCapabilitiesOutput{}, errors.New("element_id cannot be empty")
	}

	// Verify element exists
	_, err := s.repo.GetByID(input.ElementID)
	if err != nil {
		return nil, FindSimilarCapabilitiesOutput{}, fmt.Errorf("element not found: %w", err)
	}

	// Set default max results
	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}
	if maxResults > 50 {
		maxResults = 50
	}

	// Find similar documents using HNSW-backed hybrid search
	// We need to get the element's text content first
	element, err := s.repo.GetByID(input.ElementID)
	if err != nil {
		return nil, FindSimilarCapabilitiesOutput{}, fmt.Errorf("element not found: %w", err)
	}

	// Create searchable text from element
	text := s.createSearchableText(element)
	similarResults, err := s.hybridSearch.Search(ctx, text, maxResults+1, nil) // +1 to exclude self
	if err != nil {
		return nil, FindSimilarCapabilitiesOutput{}, fmt.Errorf("similarity search failed: %w", err)
	}

	// Remove self from results
	var filtered []embeddings.Result
	for _, r := range similarResults {
		if r.ID != input.ElementID {
			filtered = append(filtered, r)
		}
	}
	similarResults = filtered

	// Limit to maxResults
	if len(similarResults) > maxResults {
		similarResults = similarResults[:maxResults]
	}

	// Convert to output format
	similar := make([]SimilarCapabilityItem, len(similarResults))
	for i, r := range similarResults {
		name := "Unknown"
		if nameVal, ok := r.Metadata["name"].(string); ok {
			name = nameVal
		}
		typeStr := "unknown"
		if typeVal, ok := r.Metadata["type"].(string); ok {
			typeStr = typeVal
		}
		similar[i] = SimilarCapabilityItem{
			DocumentID: r.ID,
			Type:       typeStr,
			Name:       name,
			Similarity: r.Score,
		}
	}

	output := FindSimilarCapabilitiesOutput{
		Similar:   similar,
		ElementID: input.ElementID,
		Total:     len(similar),
	}

	return nil, output, nil
}

// handleMapCapabilityRelationships handles the map_capability_relationships tool call.
func (s *MCPServer) handleMapCapabilityRelationships(ctx context.Context, req *sdk.CallToolRequest, input MapCapabilityRelationshipsInput) (*sdk.CallToolResult, MapCapabilityRelationshipsOutput, error) {
	// Validate input
	if strings.TrimSpace(input.ElementID) == "" {
		return nil, MapCapabilityRelationshipsOutput{}, errors.New("element_id cannot be empty")
	}

	// Verify element exists
	element, err := s.repo.GetByID(input.ElementID)
	if err != nil {
		return nil, MapCapabilityRelationshipsOutput{}, fmt.Errorf("element not found: %w", err)
	}

	// Set default threshold
	threshold := input.Threshold
	if threshold <= 0 {
		threshold = 0.3
	}

	// Find similar elements using HNSW-backed hybrid search
	text := s.createSearchableText(element)
	similarResults, err := s.hybridSearch.Search(ctx, text, 50, nil)
	if err != nil {
		return nil, MapCapabilityRelationshipsOutput{}, fmt.Errorf("similarity search failed: %w", err)
	}

	// Build relationships (initialize as empty slice, not nil)
	relationships := make([]RelationshipItem, 0)
	for _, r := range similarResults {
		if r.ID == input.ElementID {
			continue // Skip self
		}
		if r.Score < threshold {
			continue
		}

		// Get element type from metadata
		typeStr := "unknown"
		if typeVal, ok := r.Metadata["type"].(string); ok {
			typeStr = typeVal
		}
		name := "Unknown"
		if nameVal, ok := r.Metadata["name"].(string); ok {
			name = nameVal
		}

		// Determine relationship type based on similarity
		relType := "related"
		if r.Score >= 0.8 {
			relType = "similar"
		} else if r.Score >= 0.5 && domain.ElementType(typeStr) != element.GetType() {
			relType = "complementary"
		}

		relationships = append(relationships, RelationshipItem{
			TargetID:         r.ID,
			TargetType:       typeStr,
			TargetName:       name,
			Similarity:       r.Score,
			RelationshipType: relType,
		})
	}

	// Build graph structure
	graph := map[string]interface{}{
		"nodes": []map[string]interface{}{
			{
				"id":   element.GetID(),
				"type": string(element.GetType()),
				"name": element.GetMetadata().Name,
			},
		},
		"edges": []map[string]interface{}{},
	}

	// Add nodes and edges
	for _, rel := range relationships {
		graph["nodes"] = append(graph["nodes"].([]map[string]interface{}), map[string]interface{}{
			"id":   rel.TargetID,
			"type": rel.TargetType,
			"name": rel.TargetName,
		})

		graph["edges"] = append(graph["edges"].([]map[string]interface{}), map[string]interface{}{
			"source": element.GetID(),
			"target": rel.TargetID,
			"weight": rel.Similarity,
			"type":   rel.RelationshipType,
		})
	}

	output := MapCapabilityRelationshipsOutput{
		ElementID:     input.ElementID,
		Relationships: relationships,
		Graph:         graph,
	}

	return nil, output, nil
}

// handleGetCapabilityIndexStats handles the get_capability_index_stats tool call.
func (s *MCPServer) handleGetCapabilityIndexStats(ctx context.Context, req *sdk.CallToolRequest, input GetCapabilityIndexStatsInput) (*sdk.CallToolResult, GetCapabilityIndexStatsOutput, error) {
	// Get stats from HNSW-backed hybrid search
	stats := s.hybridSearch.GetStatistics()

	// Extract stats
	totalDocs := stats.TotalDocuments
	docsByType := make(map[domain.ElementType]int)
	uniqueTerms := 0 // HNSW doesn't track terms

	// Count documents by type from repository
	filter := domain.ElementFilter{}
	elements, err := s.repo.List(filter)
	if err == nil {
		for _, elem := range elements {
			docsByType[elem.GetType()]++
		}
	}

	// Convert type map to string keys
	docsByTypeStr := make(map[string]int)
	for k, v := range docsByType {
		docsByTypeStr[string(k)] = v
	}

	// Determine health
	health := "empty"
	if totalDocs > 0 {
		health = "healthy"
	}

	output := GetCapabilityIndexStatsOutput{
		TotalDocuments:     totalDocs,
		DocumentsByType:    docsByTypeStr,
		UniqueTerms:        uniqueTerms,
		AverageTermsPerDoc: 0.0, // HNSW doesn't track this
		IndexHealth:        health,
		LastUpdated:        "real-time",
	}

	return nil, output, nil
}
