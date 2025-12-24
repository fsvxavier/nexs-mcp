package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
)

// --- Temporal Tool Input/Output structures ---

// GetElementHistoryInput defines input for get_element_history tool.
type GetElementHistoryInput struct {
	ElementID string `json:"element_id"           jsonschema:"element ID to retrieve history for"`
	StartTime string `json:"start_time,omitempty" jsonschema:"start time for history range (RFC3339 format, e.g. 2024-01-01T00:00:00Z)"`
	EndTime   string `json:"end_time,omitempty"   jsonschema:"end time for history range (RFC3339 format, e.g. 2024-12-31T23:59:59Z)"`
}

// GetElementHistoryOutput defines output for get_element_history tool.
type GetElementHistoryOutput struct {
	ElementID string                            `json:"element_id" jsonschema:"element ID"`
	History   []application.ElementHistoryEntry `json:"history"    jsonschema:"list of historical versions"`
	Total     int                               `json:"total"      jsonschema:"total number of versions"`
}

// GetRelationHistoryInput defines input for get_relation_history tool.
type GetRelationHistoryInput struct {
	RelationshipID string `json:"relationship_id"       jsonschema:"relationship ID to retrieve history for"`
	StartTime      string `json:"start_time,omitempty"  jsonschema:"start time for history range (RFC3339 format)"`
	EndTime        string `json:"end_time,omitempty"    jsonschema:"end time for history range (RFC3339 format)"`
	ApplyDecay     bool   `json:"apply_decay,omitempty" jsonschema:"apply confidence decay to historical values (default: false)"`
}

// GetRelationHistoryOutput defines output for get_relation_history tool.
type GetRelationHistoryOutput struct {
	RelationshipID string                             `json:"relationship_id" jsonschema:"relationship ID"`
	History        []application.RelationHistoryEntry `json:"history"         jsonschema:"list of historical versions"`
	Total          int                                `json:"total"           jsonschema:"total number of versions"`
}

// GetGraphAtTimeInput defines input for get_graph_at_time tool.
type GetGraphAtTimeInput struct {
	TargetTime string `json:"target_time"           jsonschema:"point in time to reconstruct graph (RFC3339 format, e.g. 2024-06-15T14:30:00Z)"`
	ApplyDecay bool   `json:"apply_decay,omitempty" jsonschema:"apply confidence decay to relationships (default: false)"`
}

// GetGraphAtTimeOutput defines output for get_graph_at_time tool.
type GetGraphAtTimeOutput struct {
	Timestamp         string                            `json:"timestamp"          jsonschema:"the target timestamp"`
	Elements          map[string]map[string]interface{} `json:"elements"           jsonschema:"elements at this point in time"`
	Relationships     map[string]map[string]interface{} `json:"relationships"      jsonschema:"relationships at this point in time"`
	ElementCount      int                               `json:"element_count"      jsonschema:"number of elements"`
	RelationshipCount int                               `json:"relationship_count" jsonschema:"number of relationships"`
	DecayApplied      bool                              `json:"decay_applied"      jsonschema:"whether decay was applied"`
}

// GetDecayedGraphInput defines input for get_decayed_graph tool.
type GetDecayedGraphInput struct {
	ConfidenceThreshold float64 `json:"confidence_threshold,omitempty" jsonschema:"minimum confidence threshold for relationships (default: 0.5, range: 0.0-1.0)"`
}

// GetDecayedGraphOutput defines output for get_decayed_graph tool.
type GetDecayedGraphOutput struct {
	Timestamp           string                            `json:"timestamp"            jsonschema:"current timestamp"`
	Elements            map[string]map[string]interface{} `json:"elements"             jsonschema:"all current elements"`
	Relationships       map[string]map[string]interface{} `json:"relationships"        jsonschema:"relationships above threshold with decayed confidence"`
	ElementCount        int                               `json:"element_count"        jsonschema:"number of elements"`
	RelationshipCount   int                               `json:"relationship_count"   jsonschema:"number of relationships (after filtering)"`
	ConfidenceThreshold float64                           `json:"confidence_threshold" jsonschema:"threshold used for filtering"`
	TotalRelationships  int                               `json:"total_relationships"  jsonschema:"total relationships before filtering"`
	FilteredOut         int                               `json:"filtered_out"         jsonschema:"number of relationships filtered due to low confidence"`
}

// --- Temporal Tool Handlers ---

// handleGetElementHistory retrieves the complete version history of an element.
func (s *MCPServer) handleGetElementHistory(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input GetElementHistoryInput,
) (*sdk.CallToolResult, GetElementHistoryOutput, error) {
	if s.temporalService == nil {
		return nil, GetElementHistoryOutput{}, errors.New("temporal service not available")
	}

	var startTime, endTime *time.Time

	// Parse start time if provided
	if input.StartTime != "" {
		t, err := time.Parse(time.RFC3339, input.StartTime)
		if err != nil {
			return nil, GetElementHistoryOutput{}, fmt.Errorf("invalid start_time format (use RFC3339): %w", err)
		}
		startTime = &t
	}

	// Parse end time if provided
	if input.EndTime != "" {
		t, err := time.Parse(time.RFC3339, input.EndTime)
		if err != nil {
			return nil, GetElementHistoryOutput{}, fmt.Errorf("invalid end_time format (use RFC3339): %w", err)
		}
		endTime = &t
	}

	history, err := s.temporalService.GetElementHistory(ctx, input.ElementID, startTime, endTime)
	if err != nil {
		return nil, GetElementHistoryOutput{}, fmt.Errorf("failed to get element history: %w", err)
	}

	output := GetElementHistoryOutput{
		ElementID: input.ElementID,
		History:   history,
		Total:     len(history),
	}

	return nil, output, nil
}

// handleGetRelationHistory retrieves the complete version history of a relationship.
func (s *MCPServer) handleGetRelationHistory(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input GetRelationHistoryInput,
) (*sdk.CallToolResult, GetRelationHistoryOutput, error) {
	if s.temporalService == nil {
		return nil, GetRelationHistoryOutput{}, errors.New("temporal service not available")
	}

	var startTime, endTime *time.Time

	// Parse start time if provided
	if input.StartTime != "" {
		t, err := time.Parse(time.RFC3339, input.StartTime)
		if err != nil {
			return nil, GetRelationHistoryOutput{}, fmt.Errorf("invalid start_time format (use RFC3339): %w", err)
		}
		startTime = &t
	}

	// Parse end time if provided
	if input.EndTime != "" {
		t, err := time.Parse(time.RFC3339, input.EndTime)
		if err != nil {
			return nil, GetRelationHistoryOutput{}, fmt.Errorf("invalid end_time format (use RFC3339): %w", err)
		}
		endTime = &t
	}

	history, err := s.temporalService.GetRelationshipHistory(ctx, input.RelationshipID, startTime, endTime, input.ApplyDecay)
	if err != nil {
		return nil, GetRelationHistoryOutput{}, fmt.Errorf("failed to get relationship history: %w", err)
	}

	output := GetRelationHistoryOutput{
		RelationshipID: input.RelationshipID,
		History:        history,
		Total:          len(history),
	}

	return nil, output, nil
}

// handleGetGraphAtTime reconstructs the graph state at a specific point in time.
func (s *MCPServer) handleGetGraphAtTime(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input GetGraphAtTimeInput,
) (*sdk.CallToolResult, GetGraphAtTimeOutput, error) {
	if s.temporalService == nil {
		return nil, GetGraphAtTimeOutput{}, errors.New("temporal service not available")
	}

	// Parse target time
	targetTime, err := time.Parse(time.RFC3339, input.TargetTime)
	if err != nil {
		return nil, GetGraphAtTimeOutput{}, fmt.Errorf("invalid target_time format (use RFC3339): %w", err)
	}

	snapshot, err := s.temporalService.GetGraphAtTime(ctx, targetTime, input.ApplyDecay)
	if err != nil {
		return nil, GetGraphAtTimeOutput{}, fmt.Errorf("failed to get graph at time: %w", err)
	}

	output := GetGraphAtTimeOutput{
		Timestamp:         snapshot.Timestamp.Format(time.RFC3339),
		Elements:          snapshot.Elements,
		Relationships:     snapshot.Relationships,
		ElementCount:      len(snapshot.Elements),
		RelationshipCount: len(snapshot.Relationships),
		DecayApplied:      snapshot.DecayApplied,
	}

	return nil, output, nil
}

// handleGetDecayedGraph returns the current graph with confidence decay applied.
func (s *MCPServer) handleGetDecayedGraph(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input GetDecayedGraphInput,
) (*sdk.CallToolResult, GetDecayedGraphOutput, error) {
	if s.temporalService == nil {
		return nil, GetDecayedGraphOutput{}, errors.New("temporal service not available")
	}

	// Default threshold
	threshold := input.ConfidenceThreshold
	if threshold == 0 {
		threshold = 0.5
	}

	// Validate threshold
	if threshold < 0 || threshold > 1 {
		return nil, GetDecayedGraphOutput{}, errors.New("confidence_threshold must be between 0.0 and 1.0")
	}

	// Get version stats to calculate total relationships before filtering
	stats := s.temporalService.GetVersionStats(ctx)
	totalRelationships := 0
	if tr, ok := stats["tracked_relationships"].(int); ok {
		totalRelationships = tr
	}

	snapshot, err := s.temporalService.GetDecayedGraph(ctx, threshold)
	if err != nil {
		return nil, GetDecayedGraphOutput{}, fmt.Errorf("failed to get decayed graph: %w", err)
	}

	filteredOut := totalRelationships - len(snapshot.Relationships)

	output := GetDecayedGraphOutput{
		Timestamp:           snapshot.Timestamp.Format(time.RFC3339),
		Elements:            snapshot.Elements,
		Relationships:       snapshot.Relationships,
		ElementCount:        len(snapshot.Elements),
		RelationshipCount:   len(snapshot.Relationships),
		ConfidenceThreshold: threshold,
		TotalRelationships:  totalRelationships,
		FilteredOut:         filteredOut,
	}

	return nil, output, nil
}

// --- Output Formatters ---

// formatElementHistoryOutput formats element history output for display.
