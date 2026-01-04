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

// SuggestRelatedElementsInput defines input for suggest_related_elements tool.
type SuggestRelatedElementsInput struct {
	ElementID   string   `json:"element_id"             jsonschema:"required"                                                                               jsonschema_description:"Element ID to get suggestions for"`
	ElementType string   `json:"element_type,omitempty" jsonschema_description:"Filter by element type (persona, skill, agent, template, ensemble, memory)"`
	ExcludeIDs  []string `json:"exclude_ids,omitempty"  jsonschema_description:"Element IDs to exclude from suggestions"`
	MinScore    float64  `json:"min_score,omitempty"    jsonschema_description:"Minimum recommendation score (0-1, default: 0.1)"`
	MaxResults  int      `json:"max_results,omitempty"  jsonschema_description:"Maximum number of suggestions (default: 10)"`
}

// SuggestRelatedElementsOutput defines output for suggest_related_elements tool.
type SuggestRelatedElementsOutput struct {
	ElementID      string                   `json:"element_id"      jsonschema_description:"Element ID that was analyzed"`
	ElementType    string                   `json:"element_type"    jsonschema_description:"Type of the element"`
	ElementName    string                   `json:"element_name"    jsonschema_description:"Name of the element"`
	Suggestions    []map[string]interface{} `json:"suggestions"     jsonschema_description:"Recommended related elements"`
	TotalFound     int                      `json:"total_found"     jsonschema_description:"Number of suggestions found"`
	SearchDuration int64                    `json:"search_duration" jsonschema_description:"Time taken to generate suggestions (milliseconds)"`
}

// handleSuggestRelatedElements handles recommendation requests.
func (s *MCPServer) handleSuggestRelatedElements(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input SuggestRelatedElementsInput,
) (*sdk.CallToolResult, SuggestRelatedElementsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "suggest_related_elements",
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

	// Validate input
	if input.ElementID == "" {
		handlerErr = errors.New("element_id is required")
		return nil, SuggestRelatedElementsOutput{}, handlerErr
	}

	// Get element to verify it exists
	elem, err := s.repo.GetByID(input.ElementID)
	if err != nil {
		handlerErr = fmt.Errorf("element not found: %w", err)
		return nil, SuggestRelatedElementsOutput{}, handlerErr
	}

	metadata := elem.GetMetadata()

	// Parse element type filter
	var elementType *domain.ElementType
	if input.ElementType != "" {
		et := domain.ElementType(input.ElementType)
		if !isValidElementType(et) {
			handlerErr = fmt.Errorf("invalid element_type: %s", input.ElementType)
			return nil, SuggestRelatedElementsOutput{}, handlerErr
		}
		elementType = &et
	}

	// Set defaults
	if input.MinScore == 0 {
		input.MinScore = 0.1
	}
	if input.MaxResults == 0 {
		input.MaxResults = 10
	}

	// Create recommendation engine
	engine := application.NewRecommendationEngine(s.repo, s.relationshipIndex)

	// Get recommendations
	options := application.RecommendationOptions{
		ElementType:    elementType,
		ExcludeIDs:     input.ExcludeIDs,
		MinScore:       input.MinScore,
		MaxResults:     input.MaxResults,
		IncludeReasons: true,
	}

	recommendations, err := engine.RecommendForElement(ctx, input.ElementID, options)
	if err != nil {
		handlerErr = fmt.Errorf("failed to generate recommendations: %w", err)
		return nil, SuggestRelatedElementsOutput{}, handlerErr
	}

	// Convert recommendations to maps
	suggestions := make([]map[string]interface{}, len(recommendations))
	for i, rec := range recommendations {
		suggestions[i] = map[string]interface{}{
			"element_id":   rec.ElementID,
			"element_type": string(rec.ElementType),
			"element_name": rec.ElementName,
			"score":        rec.Score,
			"reasons":      rec.Reasons,
		}
	}

	// Build output
	output := SuggestRelatedElementsOutput{
		ElementID:      input.ElementID,
		ElementType:    string(metadata.Type),
		ElementName:    metadata.Name,
		Suggestions:    suggestions,
		TotalFound:     len(suggestions),
		SearchDuration: time.Since(startTime).Milliseconds(),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "suggest_related_elements", output)

	return nil, output, nil
}
