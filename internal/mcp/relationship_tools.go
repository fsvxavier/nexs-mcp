package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// --- MCP Tool Inputs/Outputs ---

// GetRelatedElementsInput defines input for get_related_elements tool.
type GetRelatedElementsInput struct {
	ElementID       string   `json:"element_id"              jsonschema:"the element ID to find relationships for"`
	Direction       string   `json:"direction,omitempty"     jsonschema:"relationship direction: 'forward', 'reverse', or 'both' (default: both)"`
	ElementTypes    []string `json:"element_types,omitempty" jsonschema:"filter by element types"`
	IncludeInactive bool     `json:"include_inactive,omitempty" jsonschema:"include inactive elements"`
}

// GetRelatedElementsOutput defines output for get_related_elements tool.
type GetRelatedElementsOutput struct {
	ElementID string                   `json:"element_id"`
	Forward   []map[string]interface{} `json:"forward"   jsonschema:"elements this element points to"`
	Reverse   []map[string]interface{} `json:"reverse"   jsonschema:"elements that point to this element"`
	Total     int                      `json:"total"`
}

// ExpandRelationshipsInput defines input for expand_relationships tool.
type ExpandRelationshipsInput struct {
	ElementID      string   `json:"element_id"                 jsonschema:"the root element ID"`
	MaxDepth       int      `json:"max_depth,omitempty"        jsonschema:"maximum recursion depth (default: 3, max: 5)"`
	IncludeTypes   []string `json:"include_types,omitempty"    jsonschema:"include only these element types"`
	ExcludeVisited bool     `json:"exclude_visited,omitempty"  jsonschema:"prevent revisiting same elements (default: true)"`
	FollowBothWays bool     `json:"follow_both_ways,omitempty" jsonschema:"expand both forward and reverse relationships"`
}

// ExpandRelationshipsOutput defines output for expand_relationships tool.
type ExpandRelationshipsOutput struct {
	RootID        string                 `json:"root_id"`
	TotalElements int                    `json:"total_elements"`
	MaxDepth      int                    `json:"max_depth"`
	Graph         map[string]interface{} `json:"graph" jsonschema:"hierarchical relationship graph"`
}

// InferRelationshipsInput defines input for infer_relationships tool.
type InferRelationshipsInput struct {
	ElementID       string   `json:"element_id"                 jsonschema:"the element ID to infer relationships for"`
	Methods         []string `json:"methods,omitempty"          jsonschema:"inference methods: 'mention', 'keyword', 'semantic', 'pattern' (default: all)"`
	MinConfidence   float64  `json:"min_confidence,omitempty"   jsonschema:"minimum confidence threshold 0.0-1.0 (default: 0.5)"`
	TargetTypes     []string `json:"target_types,omitempty"     jsonschema:"limit inference to these element types"`
	AutoApply       bool     `json:"auto_apply,omitempty"       jsonschema:"automatically apply inferred relationships"`
	RequireEvidence int      `json:"require_evidence,omitempty" jsonschema:"minimum evidence count (default: 1)"`
}

// InferRelationshipsOutput defines output for infer_relationships tool.
type InferRelationshipsOutput struct {
	ElementID     string                   `json:"element_id"`
	TotalInferred int                      `json:"total_inferred"`
	AutoApplied   bool                     `json:"auto_applied"`
	Inferences    []map[string]interface{} `json:"inferences"`
}

// GetRecommendationsInput defines input for get_recommendations tool.
type GetRecommendationsInput struct {
	ElementID      string  `json:"element_id"                jsonschema:"the element ID to get recommendations for"`
	ElementType    string  `json:"element_type,omitempty"    jsonschema:"filter recommendations by type"`
	MinScore       float64 `json:"min_score,omitempty"       jsonschema:"minimum recommendation score 0.0-1.0 (default: 0.1)"`
	MaxResults     int     `json:"max_results,omitempty"     jsonschema:"maximum number of recommendations (default: 10)"`
	IncludeReasons bool    `json:"include_reasons,omitempty" jsonschema:"include explanation of recommendations"`
}

// GetRecommendationsOutput defines output for get_recommendations tool.
type GetRecommendationsOutput struct {
	ElementID       string                   `json:"element_id"`
	TotalFound      int                      `json:"total_found"`
	Recommendations []map[string]interface{} `json:"recommendations"`
}

// GetRelationshipStatsInput defines input for get_relationship_stats tool.
type GetRelationshipStatsInput struct {
	ElementID string `json:"element_id,omitempty" jsonschema:"optional element ID for detailed stats"`
}

// GetRelationshipStatsOutput defines output for get_relationship_stats tool.
type GetRelationshipStatsOutput struct {
	ForwardEntries int     `json:"forward_entries"  jsonschema:"number of memories with relationships"`
	ReverseEntries int     `json:"reverse_entries"  jsonschema:"number of elements referenced by memories"`
	CacheHits      int64   `json:"cache_hits"       jsonschema:"cache hit count"`
	CacheMisses    int64   `json:"cache_misses"     jsonschema:"cache miss count"`
	CacheHitRate   float64 `json:"cache_hit_rate"   jsonschema:"cache hit rate percentage"`
	CacheSize      int     `json:"cache_size"       jsonschema:"number of cached entries"`
	ElementDetails *struct {
		ForwardCount int `json:"forward_count"`
		ReverseCount int `json:"reverse_count"`
	} `json:"element_details,omitempty"`
}

// --- MCP Tool Handlers ---

// handleGetRelatedElements retrieves related elements bidirectionally.
func (s *MCPServer) handleGetRelatedElements(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input GetRelatedElementsInput,
) (*sdk.CallToolResult, GetRelatedElementsOutput, error) {
	if input.Direction == "" {
		input.Direction = "both"
	}

	// Get bidirectional relationships
	relationships := s.relationshipIndex.GetBidirectionalRelationships(input.ElementID)

	output := GetRelatedElementsOutput{
		ElementID: input.ElementID,
		Forward:   []map[string]interface{}{},
		Reverse:   []map[string]interface{}{},
	}

	// Fetch forward elements
	if input.Direction == "forward" || input.Direction == "both" {
		for _, id := range relationships.Forward {
			elem, err := s.repo.GetByID(id)
			if err != nil {
				continue
			}

			if !input.IncludeInactive && !elem.IsActive() {
				continue
			}

			meta := elem.GetMetadata()
			if len(input.ElementTypes) > 0 && !containsString(input.ElementTypes, string(meta.Type)) {
				continue
			}

			output.Forward = append(output.Forward, meta.ToMap())
		}
	}

	// Fetch reverse elements
	if input.Direction == "reverse" || input.Direction == "both" {
		for _, id := range relationships.Reverse {
			elem, err := s.repo.GetByID(id)
			if err != nil {
				continue
			}

			if !input.IncludeInactive && !elem.IsActive() {
				continue
			}

			meta := elem.GetMetadata()
			if len(input.ElementTypes) > 0 && !containsString(input.ElementTypes, string(meta.Type)) {
				continue
			}

			output.Reverse = append(output.Reverse, meta.ToMap())
		}
	}

	output.Total = len(output.Forward) + len(output.Reverse)

	return nil, output, nil
}

// handleExpandRelationships performs recursive relationship expansion.
func (s *MCPServer) handleExpandRelationships(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input ExpandRelationshipsInput,
) (*sdk.CallToolResult, ExpandRelationshipsOutput, error) {
	// Set defaults and limits
	if input.MaxDepth == 0 {
		input.MaxDepth = 3
	}
	if input.MaxDepth > 5 {
		input.MaxDepth = 5 // Hard limit
	}

	// Convert string types to ElementType
	var includeTypes []domain.ElementType
	for _, t := range input.IncludeTypes {
		includeTypes = append(includeTypes, domain.ElementType(t))
	}

	// Build expansion options
	opts := application.RelationshipExpansionOptions{
		MaxDepth:       input.MaxDepth,
		IncludeTypes:   includeTypes,
		ExcludeVisited: input.ExcludeVisited,
		FollowBothWays: input.FollowBothWays,
	}

	// Perform expansion
	rootNode, err := s.relationshipIndex.ExpandRelationships(ctx, input.ElementID, s.repo, opts)
	if err != nil {
		return nil, ExpandRelationshipsOutput{}, fmt.Errorf("expansion failed: %w", err)
	}

	// Convert to output format
	graph := convertNodeToMap(rootNode)
	totalElements := countNodesInTree(rootNode)

	output := ExpandRelationshipsOutput{
		RootID:        input.ElementID,
		TotalElements: totalElements,
		MaxDepth:      input.MaxDepth,
		Graph:         graph,
	}

	return nil, output, nil
}

// handleInferRelationships infers relationships from content analysis.
func (s *MCPServer) handleInferRelationships(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input InferRelationshipsInput,
) (*sdk.CallToolResult, InferRelationshipsOutput, error) {
	// Build inference options
	var targetTypes []domain.ElementType
	for _, t := range input.TargetTypes {
		targetTypes = append(targetTypes, domain.ElementType(t))
	}

	opts := application.InferenceOptions{
		MinConfidence:   input.MinConfidence,
		Methods:         input.Methods,
		TargetTypes:     targetTypes,
		AutoApply:       input.AutoApply,
		RequireEvidence: input.RequireEvidence,
	}

	// Perform inference
	inferences, err := s.inferenceEngine.InferRelationshipsForElement(ctx, input.ElementID, opts)
	if err != nil {
		return nil, InferRelationshipsOutput{}, fmt.Errorf("inference failed: %w", err)
	}

	// Convert to output format
	output := InferRelationshipsOutput{
		ElementID:     input.ElementID,
		TotalInferred: len(inferences),
		AutoApplied:   input.AutoApply,
		Inferences:    []map[string]interface{}{},
	}

	for _, inf := range inferences {
		output.Inferences = append(output.Inferences, map[string]interface{}{
			"source_id":   inf.SourceID,
			"target_id":   inf.TargetID,
			"source_type": string(inf.SourceType),
			"target_type": string(inf.TargetType),
			"confidence":  inf.Confidence,
			"evidence":    inf.Evidence,
			"inferred_by": inf.InferredBy,
		})
	}

	return nil, output, nil
}

// handleGetRecommendations gets intelligent recommendations for an element.
func (s *MCPServer) handleGetRecommendations(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input GetRecommendationsInput,
) (*sdk.CallToolResult, GetRecommendationsOutput, error) {
	// Build recommendation options
	opts := application.RecommendationOptions{
		MinScore:       input.MinScore,
		MaxResults:     input.MaxResults,
		IncludeReasons: input.IncludeReasons,
	}

	if input.ElementType != "" {
		elemType := domain.ElementType(input.ElementType)
		opts.ElementType = &elemType
	}

	// Get recommendations
	recommendations, err := s.recommendationEngine.RecommendForElement(ctx, input.ElementID, opts)
	if err != nil {
		return nil, GetRecommendationsOutput{}, fmt.Errorf("recommendation failed: %w", err)
	}

	// Convert to output format
	output := GetRecommendationsOutput{
		ElementID:       input.ElementID,
		TotalFound:      len(recommendations),
		Recommendations: []map[string]interface{}{},
	}

	for _, rec := range recommendations {
		recMap := map[string]interface{}{
			"element_id":   rec.ElementID,
			"element_type": string(rec.ElementType),
			"element_name": rec.ElementName,
			"score":        rec.Score,
		}
		if input.IncludeReasons {
			recMap["reasons"] = rec.Reasons
		}
		output.Recommendations = append(output.Recommendations, recMap)
	}

	return nil, output, nil
}

// handleGetRelationshipStats gets relationship index statistics.
func (s *MCPServer) handleGetRelationshipStats(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input GetRelationshipStatsInput,
) (*sdk.CallToolResult, GetRelationshipStatsOutput, error) {
	stats := s.relationshipIndex.Stats()

	output := GetRelationshipStatsOutput{
		ForwardEntries: stats.ForwardEntries,
		ReverseEntries: stats.ReverseEntries,
		CacheHits:      stats.CacheHits,
		CacheMisses:    stats.CacheMisses,
		CacheSize:      stats.CacheSize,
	}

	// Calculate cache hit rate
	totalRequests := stats.CacheHits + stats.CacheMisses
	if totalRequests > 0 {
		output.CacheHitRate = float64(stats.CacheHits) / float64(totalRequests) * 100
	}

	// Get element-specific details if requested
	if input.ElementID != "" {
		forwardCount := len(s.relationshipIndex.GetRelatedElements(input.ElementID))
		reverseCount := len(s.relationshipIndex.GetRelatedMemories(input.ElementID))

		output.ElementDetails = &struct {
			ForwardCount int `json:"forward_count"`
			ReverseCount int `json:"reverse_count"`
		}{
			ForwardCount: forwardCount,
			ReverseCount: reverseCount,
		}
	}

	return nil, output, nil
}

// --- Helper functions ---

func convertNodeToMap(node *application.RelationshipNode) map[string]interface{} {
	if node == nil {
		return nil
	}

	meta := node.Element.GetMetadata()
	result := map[string]interface{}{
		"id":           meta.ID,
		"type":         string(meta.Type),
		"name":         meta.Name,
		"depth":        node.Depth,
		"relationship": node.Relationship,
		"score":        node.Score,
	}

	if len(node.Children) > 0 {
		children := make([]map[string]interface{}, 0, len(node.Children))
		for _, child := range node.Children {
			children = append(children, convertNodeToMap(child))
		}
		result["children"] = children
	}

	return result
}

func countNodesInTree(node *application.RelationshipNode) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.Children {
		count += countNodesInTree(child)
	}

	return count
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
