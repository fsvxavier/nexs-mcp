package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/quality"
)

// ScoreMemoryQualityInput defines input for score_memory_quality tool.
type ScoreMemoryQualityInput struct {
	MemoryID           string                   `json:"memory_id"            jsonschema:"the memory ID to score"`
	UseImplicitSignals bool                     `json:"use_implicit_signals" jsonschema:"use implicit signals for scoring (default: false)"`
	ImplicitSignals    *quality.ImplicitSignals `json:"implicit_signals"     jsonschema:"implicit signals for quality estimation (optional)"`
}

// ScoreMemoryQualityOutput defines output for score_memory_quality tool.
type ScoreMemoryQualityOutput struct {
	QualityScore    float64                `json:"quality_score"    jsonschema:"quality score (0.0-1.0)"`
	Confidence      float64                `json:"confidence"       jsonschema:"confidence in the score (0.0-1.0)"`
	Method          string                 `json:"method"           jsonschema:"scoring method used (onnx, groq, gemini, implicit)"`
	Timestamp       string                 `json:"timestamp"        jsonschema:"timestamp of scoring"`
	Metadata        map[string]interface{} `json:"metadata"         jsonschema:"additional scoring metadata"`
	RetentionPolicy map[string]interface{} `json:"retention_policy" jsonschema:"recommended retention policy"`
}

// GetRetentionPolicyInput defines input for get_retention_policy tool.
type GetRetentionPolicyInput struct {
	QualityScore float64 `json:"quality_score" jsonschema:"quality score (0.0-1.0)"`
}

// GetRetentionPolicyOutput defines output for get_retention_policy tool.
type GetRetentionPolicyOutput struct {
	QualityScore     float64 `json:"quality_score"      jsonschema:"input quality score"`
	Tier             string  `json:"tier"               jsonschema:"policy tier (high, medium, low)"`
	RetentionDays    int     `json:"retention_days"     jsonschema:"number of days to retain"`
	ArchiveAfterDays int     `json:"archive_after_days" jsonschema:"archive threshold in days"`
	Description      string  `json:"description"        jsonschema:"policy description"`
	MinQuality       float64 `json:"min_quality"        jsonschema:"minimum quality for this tier"`
	MaxQuality       float64 `json:"max_quality"        jsonschema:"maximum quality for this tier"`
}

// GetRetentionStatsOutput defines output for get_retention_stats tool.
type GetRetentionStatsOutput struct {
	TotalScored     int            `json:"total_scored"      jsonschema:"total memories scored"`
	TotalArchived   int            `json:"total_archived"    jsonschema:"total memories archived"`
	TotalDeleted    int            `json:"total_deleted"     jsonschema:"total memories deleted"`
	LastCleanup     string         `json:"last_cleanup"      jsonschema:"last cleanup timestamp"`
	AvgQualityScore float64        `json:"avg_quality_score" jsonschema:"average quality score"`
	PolicyBreakdown map[string]int `json:"policy_breakdown"  jsonschema:"count per policy tier"`
	Running         bool           `json:"running"           jsonschema:"retention service running status"`
	AutoArchival    bool           `json:"auto_archival"     jsonschema:"auto archival enabled"`
	CleanupInterval int            `json:"cleanup_interval"  jsonschema:"cleanup interval in minutes"`
}

// RegisterQualityTools registers quality-related tools with the MCP server.
func (s *MCPServer) RegisterQualityTools() {
	if s.retentionService == nil {
		return // Retention service not available
	}

	// score_memory_quality tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "score_memory_quality",
		Description: "Score the quality of a memory using ML models or implicit signals. Returns quality score (0-1), confidence level, scoring method used, and retention policy recommendation.",
	}, s.handleScoreMemoryQuality)

	// get_retention_policy tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_retention_policy",
		Description: "Get the retention policy recommendation for a given quality score. Returns retention days, archive threshold, and policy description.",
	}, s.handleGetRetentionPolicy)

	// get_retention_stats tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_retention_stats",
		Description: "Get statistics about memory retention operations including total scored, archived, deleted, average quality, and policy breakdown.",
	}, s.handleGetRetentionStats)
}

// handleScoreMemoryQuality handles the score_memory_quality tool.
func (s *MCPServer) handleScoreMemoryQuality(ctx context.Context, req *sdk.CallToolRequest, input ScoreMemoryQualityInput) (*sdk.CallToolResult, ScoreMemoryQualityOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "score_memory_quality",
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

	if input.MemoryID == "" {
		handlerErr = errors.New("memory_id is required")
		return nil, ScoreMemoryQualityOutput{}, handlerErr
	}

	if s.retentionService == nil {
		handlerErr = errors.New("retention service not available")
		return nil, ScoreMemoryQualityOutput{}, handlerErr
	}

	var score *quality.Score
	var err error

	// Score with implicit signals if provided
	if input.UseImplicitSignals && input.ImplicitSignals != nil {
		score, err = s.retentionService.ScoreMemoryWithSignals(ctx, input.MemoryID, *input.ImplicitSignals)
	} else {
		score, err = s.retentionService.ScoreMemory(ctx, input.MemoryID)
	}

	if err != nil {
		handlerErr = fmt.Errorf("failed to score memory: %w", err)
		return nil, ScoreMemoryQualityOutput{}, handlerErr
	}

	// Get retention policy recommendation
	policy := s.retentionService.GetRetentionPolicy(score.Value)

	output := ScoreMemoryQualityOutput{
		QualityScore: score.Value,
		Confidence:   score.Confidence,
		Method:       score.Method,
		Timestamp:    score.Timestamp.Format(time.RFC3339),
		Metadata:     score.Metadata,
		RetentionPolicy: map[string]interface{}{
			"tier":               getPolicyTier(policy),
			"retention_days":     policy.RetentionDays,
			"archive_after_days": policy.ArchiveAfterDays,
			"description":        policy.Description,
		},
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "score_memory_quality", output)

	return nil, output, nil
}

// handleGetRetentionPolicy handles the get_retention_policy tool.
func (s *MCPServer) handleGetRetentionPolicy(ctx context.Context, req *sdk.CallToolRequest, input GetRetentionPolicyInput) (*sdk.CallToolResult, GetRetentionPolicyOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_retention_policy",
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

	if input.QualityScore < 0 || input.QualityScore > 1 {
		handlerErr = errors.New("quality_score must be between 0.0 and 1.0")
		return nil, GetRetentionPolicyOutput{}, handlerErr
	}

	if s.retentionService == nil {
		handlerErr = errors.New("retention service not available")
		return nil, GetRetentionPolicyOutput{}, handlerErr
	}

	policy := s.retentionService.GetRetentionPolicy(input.QualityScore)
	if policy == nil {
		handlerErr = errors.New("no retention policy found for score")
		return nil, GetRetentionPolicyOutput{}, handlerErr
	}

	output := GetRetentionPolicyOutput{
		QualityScore:     input.QualityScore,
		Tier:             getPolicyTier(policy),
		RetentionDays:    policy.RetentionDays,
		ArchiveAfterDays: policy.ArchiveAfterDays,
		Description:      policy.Description,
		MinQuality:       policy.MinQuality,
		MaxQuality:       policy.MaxQuality,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "get_retention_policy", output)

	return nil, output, nil
}

// handleGetRetentionStats handles the get_retention_stats tool.
func (s *MCPServer) handleGetRetentionStats(ctx context.Context, req *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, GetRetentionStatsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_retention_stats",
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

	if s.retentionService == nil {
		handlerErr = errors.New("retention service not available")
		return nil, GetRetentionStatsOutput{}, handlerErr
	}

	stats := s.retentionService.GetStats()

	// Safely extract values with type assertions
	lastCleanup := ""
	if lc, ok := stats["last_cleanup"].(time.Time); ok {
		lastCleanup = lc.Format(time.RFC3339)
	}

	output := GetRetentionStatsOutput{
		TotalScored:     getIntFromMap(stats, "total_scored"),
		TotalArchived:   getIntFromMap(stats, "total_archived"),
		TotalDeleted:    getIntFromMap(stats, "total_deleted"),
		LastCleanup:     lastCleanup,
		AvgQualityScore: getFloat64FromMap(stats, "avg_quality_score"),
		PolicyBreakdown: getMapIntFromMap(stats, "policy_breakdown"),
		Running:         getBoolFromMap(stats, "running"),
		AutoArchival:    getBoolFromMap(stats, "auto_archival"),
		CleanupInterval: getIntFromMap(stats, "cleanup_interval"),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "get_retention_stats", output)

	return nil, output, nil
}

// Helper functions for type-safe map extraction.
func getIntFromMap(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

func getFloat64FromMap(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0.0
}

func getBoolFromMap(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getMapIntFromMap(m map[string]interface{}, key string) map[string]int {
	if v, ok := m[key].(map[string]int); ok {
		return v
	}
	return make(map[string]int)
}

// getPolicyTier determines the tier name for a policy.
func getPolicyTier(policy *quality.RetentionPolicy) string {
	if policy == nil {
		return "unknown"
	}
	if policy.MinQuality >= 0.7 {
		return "high"
	} else if policy.MinQuality >= 0.5 {
		return "medium"
	}
	return "low"
}
