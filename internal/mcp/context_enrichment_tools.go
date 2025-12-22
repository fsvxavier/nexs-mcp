package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// ExpandMemoryContextInput defines input for expand_memory_context tool.
type ExpandMemoryContextInput struct {
	MemoryID     string   `json:"memory_id"               jsonschema:"required"                                                                        jsonschema_description:"Memory ID to expand with related elements"`
	IncludeTypes []string `json:"include_types,omitempty" jsonschema_description:"Filter by element types (persona, skill, agent, template, ensemble)"`
	ExcludeTypes []string `json:"exclude_types,omitempty" jsonschema_description:"Exclude specific element types"`
	MaxDepth     int      `json:"max_depth,omitempty"     jsonschema_description:"Expansion depth (default: 0 = direct relationships only)"`
	MaxElements  int      `json:"max_elements,omitempty"  jsonschema_description:"Maximum number of related elements to fetch (default: 20)"`
	IgnoreErrors bool     `json:"ignore_errors,omitempty" jsonschema_description:"Continue expansion even if some elements fail to load"`
}

// ExpandMemoryContextOutput defines output for expand_memory_context tool.
type ExpandMemoryContextOutput struct {
	Memory          map[string]interface{}   `json:"memory"            jsonschema_description:"Original memory with metadata"`
	RelatedElements []map[string]interface{} `json:"related_elements"  jsonschema_description:"Array of related elements with full details"`
	RelationshipMap map[string][]string      `json:"relationship_map"  jsonschema_description:"Map of element IDs to relationship types"`
	TotalElements   int                      `json:"total_elements"    jsonschema_description:"Number of related elements loaded"`
	TokensSaved     int                      `json:"tokens_saved"      jsonschema_description:"Estimated tokens saved vs individual requests"`
	FetchDurationMs int64                    `json:"fetch_duration_ms" jsonschema_description:"Time taken to fetch related elements (milliseconds)"`
	Errors          []string                 `json:"errors,omitempty"  jsonschema_description:"Errors encountered during expansion (if any)"`
}

// handleExpandMemoryContext handles expansion of memory context with related elements.
func (s *MCPServer) handleExpandMemoryContext(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input ExpandMemoryContextInput,
) (*sdk.CallToolResult, ExpandMemoryContextOutput, error) {
	// Validate input
	if input.MemoryID == "" {
		return nil, ExpandMemoryContextOutput{}, errors.New("memory_id is required")
	}

	// Get memory element
	elem, err := s.repo.GetByID(input.MemoryID)
	if err != nil {
		return nil, ExpandMemoryContextOutput{}, fmt.Errorf("memory not found: %w", err)
	}

	// Ensure element is a memory
	memory, ok := elem.(*domain.Memory)
	if !ok {
		return nil, ExpandMemoryContextOutput{}, fmt.Errorf("element %s is not a memory (type: %s)", input.MemoryID, elem.GetType())
	}

	// Build expand options
	options := application.DefaultExpandOptions()

	if input.MaxDepth != 0 {
		options.MaxDepth = input.MaxDepth
	}

	if input.MaxElements > 0 {
		options.MaxElements = input.MaxElements
	}

	options.IgnoreErrors = input.IgnoreErrors

	// Convert include types
	if len(input.IncludeTypes) > 0 {
		includeTypes := make([]domain.ElementType, 0, len(input.IncludeTypes))
		for _, typeStr := range input.IncludeTypes {
			elemType := domain.ElementType(typeStr)
			if isValidElementType(elemType) {
				includeTypes = append(includeTypes, elemType)
			} else {
				return nil, ExpandMemoryContextOutput{}, fmt.Errorf("invalid element type: %s", typeStr)
			}
		}
		options.IncludeTypes = includeTypes
	}

	// Convert exclude types
	if len(input.ExcludeTypes) > 0 {
		excludeTypes := make([]domain.ElementType, 0, len(input.ExcludeTypes))
		for _, typeStr := range input.ExcludeTypes {
			elemType := domain.ElementType(typeStr)
			if isValidElementType(elemType) {
				excludeTypes = append(excludeTypes, elemType)
			} else {
				return nil, ExpandMemoryContextOutput{}, fmt.Errorf("invalid element type: %s", typeStr)
			}
		}
		options.ExcludeTypes = excludeTypes
	}

	// Expand context
	enriched, err := application.ExpandMemoryContext(ctx, memory, s.repo, options)
	if err != nil && !input.IgnoreErrors {
		return nil, ExpandMemoryContextOutput{}, fmt.Errorf("context expansion failed: %w", err)
	}

	// Convert to output format
	output := ExpandMemoryContextOutput{
		Memory:          convertMemoryToMap(enriched.Memory),
		RelatedElements: convertElementsToMaps(enriched.RelatedElements),
		RelationshipMap: convertRelationshipMapToStrings(enriched.RelationshipMap),
		TotalElements:   enriched.GetElementCount(),
		TokensSaved:     enriched.TotalTokensSaved,
		FetchDurationMs: enriched.FetchDuration.Milliseconds(),
	}

	// Add errors if any
	if enriched.HasErrors() {
		output.Errors = make([]string, 0, len(enriched.FetchErrors))
		for _, fetchErr := range enriched.FetchErrors {
			output.Errors = append(output.Errors, fetchErr.Error())
		}
	}

	return nil, output, nil
}

// isValidElementType checks if element type is valid.
func isValidElementType(elemType domain.ElementType) bool {
	validTypes := []domain.ElementType{
		domain.PersonaElement,
		domain.SkillElement,
		domain.TemplateElement,
		domain.AgentElement,
		domain.MemoryElement,
		domain.EnsembleElement,
	}

	for _, validType := range validTypes {
		if elemType == validType {
			return true
		}
	}

	return false
}

// convertMemoryToMap converts Memory to map for JSON output.
func convertMemoryToMap(memory *domain.Memory) map[string]interface{} {
	metadata := memory.GetMetadata()

	return map[string]interface{}{
		"id":           metadata.ID,
		"type":         string(metadata.Type),
		"name":         metadata.Name,
		"description":  metadata.Description,
		"version":      metadata.Version,
		"author":       metadata.Author,
		"tags":         metadata.Tags,
		"is_active":    metadata.IsActive,
		"created_at":   metadata.CreatedAt.Format(time.RFC3339),
		"updated_at":   metadata.UpdatedAt.Format(time.RFC3339),
		"content":      memory.Content,
		"date_created": memory.DateCreated,
		"content_hash": memory.ContentHash,
		"search_index": memory.SearchIndex,
		"metadata":     memory.Metadata,
	}
}

// convertElementsToMaps converts map of elements to array of maps.
func convertElementsToMaps(elements map[string]domain.Element) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(elements))

	for _, elem := range elements {
		metadata := elem.GetMetadata()

		// Basic element information
		elemMap := map[string]interface{}{
			"id":          metadata.ID,
			"type":        string(metadata.Type),
			"name":        metadata.Name,
			"description": metadata.Description,
			"version":     metadata.Version,
			"author":      metadata.Author,
			"tags":        metadata.Tags,
			"is_active":   metadata.IsActive,
			"created_at":  metadata.CreatedAt.Format(time.RFC3339),
			"updated_at":  metadata.UpdatedAt.Format(time.RFC3339),
		}

		// Note: Type-specific fields are private and accessed via getters
		// For full element details, clients should use get_element tool with specific IDs
		// This maintains encapsulation while providing essential context

		result = append(result, elemMap)
	}

	return result
}

// convertRelationshipMapToStrings converts RelationshipMap to string map.
func convertRelationshipMapToStrings(relMap domain.RelationshipMap) map[string][]string {
	result := make(map[string][]string, len(relMap))

	for elemID, relTypes := range relMap {
		strTypes := make([]string, len(relTypes))
		for i, relType := range relTypes {
			strTypes[i] = string(relType)
		}
		result[elemID] = strTypes
	}

	return result
}
