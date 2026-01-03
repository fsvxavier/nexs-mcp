package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// --- SimpleElement implementation ---

// SimpleElement is a basic implementation of Element for MCP operations.
type SimpleElement struct {
	metadata domain.ElementMetadata
}

func (s *SimpleElement) GetMetadata() domain.ElementMetadata { return s.metadata }
func (s *SimpleElement) Validate() error                     { return nil }
func (s *SimpleElement) GetType() domain.ElementType         { return s.metadata.Type }
func (s *SimpleElement) GetID() string                       { return s.metadata.ID }
func (s *SimpleElement) IsActive() bool                      { return s.metadata.IsActive }
func (s *SimpleElement) Activate() error {
	s.metadata.IsActive = true
	s.metadata.UpdatedAt = time.Now()
	return nil
}

func (s *SimpleElement) Deactivate() error {
	s.metadata.IsActive = false
	s.metadata.UpdatedAt = time.Now()
	return nil
}

// --- Input/Output structures for tools ---

// ListElementsInput defines input for list_elements tool.
type ListElementsInput struct {
	Type       string `json:"type,omitempty"        jsonschema:"element type filter (persona, skill, template, agent, memory, ensemble)"`
	IsActive   *bool  `json:"is_active,omitempty"   jsonschema:"active status filter"`
	ActiveOnly bool   `json:"active_only,omitempty" jsonschema:"if true, return only active elements (shortcut for is_active=true)"`
	Tags       string `json:"tags,omitempty"        jsonschema:"comma-separated tags to filter"`
	User       string `json:"user,omitempty"        jsonschema:"authenticated username for access control (optional)"`
}

// ListElementsOutput defines output for list_elements tool.
type ListElementsOutput struct {
	Elements []map[string]interface{} `json:"elements" jsonschema:"list of elements"`
	Total    int                      `json:"total"    jsonschema:"total number of elements"`
}

// GetElementInput defines input for get_element tool.
type GetElementInput struct {
	ID   string `json:"id"             jsonschema:"the element ID"`
	User string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// GetElementOutput defines output for get_element tool.
type GetElementOutput struct {
	Element map[string]interface{} `json:"element" jsonschema:"the element details"`
}

// CreateElementInput defines input for create_element tool.
type CreateElementInput struct {
	Type        string   `json:"type"                  jsonschema:"element type (persona, skill, template, agent, memory, ensemble)"`
	Name        string   `json:"name"                  jsonschema:"element name (3-100 characters)"`
	Description string   `json:"description,omitempty" jsonschema:"element description (max 500 characters)"`
	Version     string   `json:"version"               jsonschema:"element version (semver)"`
	Author      string   `json:"author"                jsonschema:"element author"`
	Tags        []string `json:"tags,omitempty"        jsonschema:"element tags"`
	IsActive    bool     `json:"is_active,omitempty"   jsonschema:"active status (default: true)"`
	User        string   `json:"user,omitempty"        jsonschema:"authenticated username for access control (optional)"`
}

// CreateElementOutput defines output for create_element tool.
type CreateElementOutput struct {
	ID      string                 `json:"id"      jsonschema:"the created element ID"`
	Element map[string]interface{} `json:"element" jsonschema:"the created element details"`
}

// UpdateElementInput defines input for update_element tool.
type UpdateElementInput struct {
	ID          string   `json:"id"                    jsonschema:"the element ID"`
	Name        string   `json:"name,omitempty"        jsonschema:"element name"`
	Description string   `json:"description,omitempty" jsonschema:"element description"`
	Tags        []string `json:"tags,omitempty"        jsonschema:"element tags"`
	IsActive    *bool    `json:"is_active,omitempty"   jsonschema:"active status"`
	User        string   `json:"user,omitempty"        jsonschema:"authenticated username for access control (optional)"`
}

// UpdateElementOutput defines output for update_element tool.
type UpdateElementOutput struct {
	Element map[string]interface{} `json:"element" jsonschema:"the updated element details"`
}

// DeleteElementInput defines input for delete_element tool.
type DeleteElementInput struct {
	ID   string `json:"id"             jsonschema:"the element ID to delete"`
	User string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// DeleteElementOutput defines output for delete_element tool.
type DeleteElementOutput struct {
	Success bool   `json:"success" jsonschema:"deletion success status"`
	Message string `json:"message" jsonschema:"deletion result message"`
}

// DuplicateElementInput defines input for duplicate_element tool.
type DuplicateElementInput struct {
	ID      string `json:"id"                 jsonschema:"the element ID to duplicate"`
	NewName string `json:"new_name,omitempty" jsonschema:"optional new name for the duplicate (default: 'Copy of {original_name}')"`
	User    string `json:"user,omitempty"     jsonschema:"authenticated username for access control (optional)"`
}

// DuplicateElementOutput defines output for duplicate_element tool.
type DuplicateElementOutput struct {
	ID      string                 `json:"id"      jsonschema:"the duplicated element ID"`
	Element map[string]interface{} `json:"element" jsonschema:"the duplicated element details"`
	Message string                 `json:"message" jsonschema:"duplication result message"`
}

// GetUsageStatsInput defines input for get_usage_stats tool.
type GetUsageStatsInput struct {
	Period string `json:"period,omitempty" jsonschema:"time period for statistics (last_hour, last_24h, last_7_days, last_30_days, all)"`
}

// GetUsageStatsOutput defines output for get_usage_stats tool.
type GetUsageStatsOutput struct {
	TotalOperations    int                      `json:"total_operations"        jsonschema:"total number of operations"`
	SuccessfulOps      int                      `json:"successful_ops"          jsonschema:"number of successful operations"`
	FailedOps          int                      `json:"failed_ops"              jsonschema:"number of failed operations"`
	SuccessRate        float64                  `json:"success_rate"            jsonschema:"success rate percentage"`
	OperationsByTool   map[string]int           `json:"operations_by_tool"      jsonschema:"operation count by tool name"`
	ErrorsByTool       map[string]int           `json:"errors_by_tool"          jsonschema:"error count by tool name"`
	AvgDurationByTool  map[string]float64       `json:"avg_duration_by_tool_ms" jsonschema:"average duration in milliseconds by tool"`
	MostUsedTools      []map[string]interface{} `json:"most_used_tools"         jsonschema:"top 10 most used tools"`
	SlowestOperations  []map[string]interface{} `json:"slowest_operations"      jsonschema:"top 10 slowest operations"`
	RecentErrors       []map[string]interface{} `json:"recent_errors"           jsonschema:"most recent errors"`
	ActiveUsers        []string                 `json:"active_users"            jsonschema:"list of active users"`
	OperationsByPeriod map[string]int           `json:"operations_by_period"    jsonschema:"operations grouped by date"`
	Period             string                   `json:"period"                  jsonschema:"period queried"`
	StartTime          string                   `json:"start_time"              jsonschema:"period start time (ISO 8601)"`
	EndTime            string                   `json:"end_time"                jsonschema:"period end time (ISO 8601)"`
}

// GetPerformanceDashboardInput defines input for get_performance_dashboard tool.
type GetPerformanceDashboardInput struct {
	Period string `json:"period,omitempty" jsonschema:"time period for dashboard (last_hour, last_24h, last_7_days, last_30_days, all)"`
}

// GetPerformanceDashboardOutput defines output for get_performance_dashboard tool.
type GetPerformanceDashboardOutput struct {
	TotalOperations int                               `json:"total_operations" jsonschema:"total number of operations in period"`
	AvgDuration     float64                           `json:"avg_duration_ms"  jsonschema:"average operation duration in milliseconds"`
	P50Duration     float64                           `json:"p50_duration_ms"  jsonschema:"50th percentile duration (median)"`
	P95Duration     float64                           `json:"p95_duration_ms"  jsonschema:"95th percentile duration"`
	P99Duration     float64                           `json:"p99_duration_ms"  jsonschema:"99th percentile duration"`
	MaxDuration     float64                           `json:"max_duration_ms"  jsonschema:"maximum duration"`
	MinDuration     float64                           `json:"min_duration_ms"  jsonschema:"minimum duration"`
	SlowOperations  []map[string]interface{}          `json:"slow_operations"  jsonschema:"top 10 slowest operations (>p95)"`
	FastOperations  []map[string]interface{}          `json:"fast_operations"  jsonschema:"top 10 fastest operations (<p50)"`
	ByOperation     map[string]map[string]interface{} `json:"by_operation"     jsonschema:"statistics per operation"`
	Period          string                            `json:"period"           jsonschema:"period analyzed"`
}

// --- Alert Management Types ---

// AlertRule represents a configurable alert rule.
type AlertRule struct {
	ID          string  `json:"id"               jsonschema:"unique rule identifier"`
	Name        string  `json:"name"             jsonschema:"rule display name"`
	Description string  `json:"description"      jsonschema:"rule description"`
	Type        string  `json:"type"             jsonschema:"alert type: performance, error_rate, token_usage, latency"`
	Metric      string  `json:"metric"           jsonschema:"metric to monitor"`
	Operator    string  `json:"operator"         jsonschema:"comparison operator: gt, lt, eq, gte, lte"`
	Threshold   float64 `json:"threshold"        jsonschema:"threshold value for triggering alert"`
	Severity    string  `json:"severity"         jsonschema:"alert severity: critical, warning, info"`
	Enabled     bool    `json:"enabled"          jsonschema:"whether rule is active"`
	Cooldown    int     `json:"cooldown_minutes" jsonschema:"minutes before same alert can trigger again"`
}

// Alert represents a triggered alert.
type Alert struct {
	ID                string     `json:"id"                       jsonschema:"unique alert identifier"`
	RuleID            string     `json:"rule_id"                  jsonschema:"ID of the rule that triggered this alert"`
	RuleName          string     `json:"rule_name"                jsonschema:"name of the rule"`
	Severity          string     `json:"severity"                 jsonschema:"alert severity: critical, warning, info"`
	Title             string     `json:"title"                    jsonschema:"alert title"`
	Message           string     `json:"message"                  jsonschema:"detailed alert message"`
	Value             float64    `json:"value"                    jsonschema:"metric value that triggered alert"`
	Threshold         float64    `json:"threshold"                jsonschema:"threshold that was exceeded"`
	TriggeredAt       time.Time  `json:"triggered_at"             jsonschema:"when the alert was triggered"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"    jsonschema:"when the alert was resolved (if resolved)"`
	Status            string     `json:"status"                   jsonschema:"alert status: active, resolved, silenced"`
	AffectedTools     []string   `json:"affected_tools,omitempty" jsonschema:"tools affected by this issue"`
	RecommendedAction string     `json:"recommended_action"       jsonschema:"recommended action to resolve the issue"`
	Fingerprint       string     `json:"fingerprint"              jsonschema:"fingerprint for deduplication"`
}

// GetCostAnalyticsInput defines input for get_cost_analytics tool.
type GetCostAnalyticsInput struct {
	Period string `json:"period,omitempty" jsonschema:"time period for cost analysis (last_hour, last_24h, last_7_days, last_30_days, all)"`
}

// GetCostAnalyticsOutput defines output for get_cost_analytics tool.
type GetCostAnalyticsOutput struct {
	Period                    string                    `json:"period"                     jsonschema:"period analyzed"`
	StartTime                 string                    `json:"start_time"                 jsonschema:"period start time (ISO 8601)"`
	EndTime                   string                    `json:"end_time"                   jsonschema:"period end time (ISO 8601)"`
	TotalOperations           int                       `json:"total_operations"           jsonschema:"total number of operations"`
	TotalTokens               int                       `json:"total_tokens"               jsonschema:"total tokens used (original)"`
	TotalOptimizedTokens      int                       `json:"total_optimized_tokens"     jsonschema:"total tokens after optimization"`
	TokenSavings              int                       `json:"token_savings"              jsonschema:"tokens saved through optimization"`
	TokenSavingsPercent       float64                   `json:"token_savings_percent"      jsonschema:"percentage of tokens saved"`
	TotalDuration             float64                   `json:"total_duration_ms"          jsonschema:"total execution time in milliseconds"`
	AverageDuration           float64                   `json:"average_duration_ms"        jsonschema:"average execution time per operation"`
	ToolCostBreakdown         []ToolCostBreakdown       `json:"tool_cost_breakdown"        jsonschema:"cost breakdown by tool"`
	TopExpensiveTools         []ExpensiveTool           `json:"top_expensive_tools"        jsonschema:"top 10 most expensive tools"`
	CostTrends                CostTrends                `json:"cost_trends"                jsonschema:"cost trends analysis"`
	OptimizationOpportunities []OptimizationOpportunity `json:"optimization_opportunities" jsonschema:"identified optimization opportunities"`
	Recommendations           []string                  `json:"recommendations"            jsonschema:"actionable recommendations"`
	Anomalies                 []CostAnomaly             `json:"anomalies"                  jsonschema:"detected cost anomalies"`
	CostProjections           CostProjections           `json:"cost_projections"           jsonschema:"future cost projections"`
}

// ToolCostBreakdown represents cost metrics for a single tool.
type ToolCostBreakdown struct {
	ToolName         string  `json:"tool_name"         jsonschema:"tool name"`
	OperationCount   int     `json:"operation_count"   jsonschema:"number of operations"`
	AvgDuration      float64 `json:"avg_duration_ms"   jsonschema:"average duration in milliseconds"`
	TotalDuration    float64 `json:"total_duration_ms" jsonschema:"total duration in milliseconds"`
	TotalTokens      int     `json:"total_tokens"      jsonschema:"total tokens (original)"`
	OptimizedTokens  int     `json:"optimized_tokens"  jsonschema:"total optimized tokens"`
	TokenSavings     int     `json:"token_savings"     jsonschema:"tokens saved"`
	CompressionRatio float64 `json:"compression_ratio" jsonschema:"compression ratio (0.0-1.0)"`
	SuccessRate      float64 `json:"success_rate"      jsonschema:"success rate (0.0-1.0)"`
	CostScore        float64 `json:"cost_score"        jsonschema:"normalized cost score (0-100)"`
}

// ExpensiveTool represents a tool with high cost.
type ExpensiveTool struct {
	ToolName         string  `json:"tool_name"          jsonschema:"tool name"`
	TotalCost        float64 `json:"total_cost"         jsonschema:"total cost score"`
	OperationCount   int     `json:"operation_count"    jsonschema:"number of operations"`
	AvgDuration      float64 `json:"avg_duration_ms"    jsonschema:"average duration in milliseconds"`
	TotalTokens      int     `json:"total_tokens"       jsonschema:"total tokens used"`
	CostPerOperation float64 `json:"cost_per_operation" jsonschema:"cost per operation"`
}

// CostTrends represents trend analysis.
type CostTrends struct {
	Period            string  `json:"period"              jsonschema:"period analyzed"`
	OperationsChange  float64 `json:"operations_change"   jsonschema:"percentage change in operations"`
	TokenUsageChange  float64 `json:"token_usage_change"  jsonschema:"percentage change in token usage"`
	AvgDurationChange float64 `json:"avg_duration_change" jsonschema:"percentage change in average duration"`
	Trend             string  `json:"trend"               jsonschema:"trend direction: increasing, decreasing, stable"`
	TrendConfidence   float64 `json:"trend_confidence"    jsonschema:"confidence in trend analysis (0.0-1.0)"`
	PeakUsageTime     string  `json:"peak_usage_time"     jsonschema:"peak usage time window"`
	LowUsageTime      string  `json:"low_usage_time"      jsonschema:"low usage time window"`
}

// OptimizationOpportunity represents an identified optimization opportunity.
type OptimizationOpportunity struct {
	Type             string `json:"type"              jsonschema:"opportunity type: compression, performance, reliability, configuration"`
	ToolName         string `json:"tool_name"         jsonschema:"tool name (or 'system' for system-wide)"`
	Description      string `json:"description"       jsonschema:"detailed description of the opportunity"`
	Severity         string `json:"severity"          jsonschema:"severity level: low, medium, high, critical"`
	PotentialSavings string `json:"potential_savings" jsonschema:"estimated savings (tokens, time, or percentage)"`
}

// CostAnomaly represents an detected cost anomaly.
type CostAnomaly struct {
	Type        string `json:"type"        jsonschema:"anomaly type: high_error_rate, low_compression, slow_operations, unusual_pattern"`
	Description string `json:"description" jsonschema:"detailed description of the anomaly"`
	Severity    string `json:"severity"    jsonschema:"severity level: low, medium, high, critical"`
	DetectedAt  string `json:"detected_at" jsonschema:"detection timestamp (ISO 8601)"`
}

// CostProjections represents future cost projections.
type CostProjections struct {
	NextDay    string  `json:"next_day"   jsonschema:"projected cost for next day"`
	NextWeek   string  `json:"next_week"  jsonschema:"projected cost for next week"`
	NextMonth  string  `json:"next_month" jsonschema:"projected cost for next month"`
	Confidence float64 `json:"confidence" jsonschema:"projection confidence (0.0-1.0)"`
	Model      string  `json:"model"      jsonschema:"forecasting model used"`
}

// --- Alert Management Types ---

// GetActiveAlertsInput is the input for get_active_alerts tool.
type GetActiveAlertsInput struct{}

// GetActiveAlertsOutput is the output for get_active_alerts tool.
type GetActiveAlertsOutput struct {
	Count  int      `json:"count"  jsonschema:"number of active alerts"`
	Alerts []*Alert `json:"alerts" jsonschema:"list of active alerts"`
}

// GetAlertHistoryInput is the input for get_alert_history tool.
type GetAlertHistoryInput struct {
	Limit    int    `json:"limit,omitempty"    jsonschema:"maximum number of alerts to return (default: 50)"`
	Severity string `json:"severity,omitempty" jsonschema:"filter by severity: critical, warning, info"`
	Status   string `json:"status,omitempty"   jsonschema:"filter by status: active, resolved, silenced"`
}

// GetAlertHistoryOutput is the output for get_alert_history tool.
type GetAlertHistoryOutput struct {
	Count  int      `json:"count"  jsonschema:"number of alerts returned"`
	Alerts []*Alert `json:"alerts" jsonschema:"list of historical alerts"`
}

// GetAlertRulesInput is the input for get_alert_rules tool.
type GetAlertRulesInput struct{}

// GetAlertRulesOutput is the output for get_alert_rules tool.
type GetAlertRulesOutput struct {
	Count int          `json:"count" jsonschema:"number of alert rules"`
	Rules []*AlertRule `json:"rules" jsonschema:"list of alert rules"`
}

// UpdateAlertRuleInput is the input for update_alert_rule tool.
type UpdateAlertRuleInput struct {
	RuleID      string  `json:"rule_id"          jsonschema:"unique rule identifier"`
	Name        string  `json:"name"             jsonschema:"rule display name"`
	Description string  `json:"description"      jsonschema:"rule description"`
	Type        string  `json:"type"             jsonschema:"alert type: performance, error_rate, token_usage, latency"`
	Metric      string  `json:"metric"           jsonschema:"metric to monitor"`
	Operator    string  `json:"operator"         jsonschema:"comparison operator: gt, lt, eq, gte, lte"`
	Threshold   float64 `json:"threshold"        jsonschema:"threshold value for triggering alert"`
	Severity    string  `json:"severity"         jsonschema:"alert severity: critical, warning, info"`
	Enabled     bool    `json:"enabled"          jsonschema:"whether rule is active"`
	Cooldown    int     `json:"cooldown_minutes" jsonschema:"minutes before same alert can trigger again"`
}

// UpdateAlertRuleOutput is the output for update_alert_rule tool.
type UpdateAlertRuleOutput struct {
	Success bool       `json:"success" jsonschema:"whether the update was successful"`
	Message string     `json:"message" jsonschema:"status message"`
	Rule    *AlertRule `json:"rule"    jsonschema:"updated alert rule"`
}

// ResolveAlertInput is the input for resolve_alert tool.
type ResolveAlertInput struct {
	AlertID string `json:"alert_id" jsonschema:"unique alert identifier"`
}

// ResolveAlertOutput is the output for resolve_alert tool.
type ResolveAlertOutput struct {
	Success bool   `json:"success" jsonschema:"whether the resolution was successful"`
	Message string `json:"message" jsonschema:"status message"`
}

// --- Tool handlers ---

// handleListElements handles list_elements tool calls.
func (s *MCPServer) handleListElements(ctx context.Context, req *sdk.CallToolRequest, input ListElementsInput) (*sdk.CallToolResult, ListElementsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "list_elements",
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

	// Build filter
	filter := domain.ElementFilter{}

	if input.Type != "" {
		elementType := domain.ElementType(input.Type)
		filter.Type = &elementType
	}

	// Handle active_only shortcut - takes priority over is_active
	if input.ActiveOnly {
		isActive := true
		filter.IsActive = &isActive
	} else if input.IsActive != nil {
		filter.IsActive = input.IsActive
	}

	if input.Tags != "" {
		// Parse comma-separated tags
		// For simplicity, we'll add basic tag parsing here
		filter.Tags = []string{input.Tags}
	}

	// List elements
	elements, err := s.repo.List(filter)
	if err != nil {
		handlerErr = fmt.Errorf("failed to list elements: %w", err)
		return nil, ListElementsOutput{}, handlerErr
	}

	// Apply access control filtering
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()
	filteredElements := accessControl.FilterByPermissions(userCtx, elements)

	// Convert to map format
	result := make([]map[string]interface{}, 0, len(filteredElements))
	for _, elem := range filteredElements {
		result = append(result, elem.GetMetadata().ToMap())
	}

	output := ListElementsOutput{
		Elements: result,
		Total:    len(result),
	}

	// Record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "list_elements", output)

	return nil, output, nil
}

// handleGetElement handles get_element tool calls.
func (s *MCPServer) handleGetElement(ctx context.Context, req *sdk.CallToolRequest, input GetElementInput) (*sdk.CallToolResult, GetElementOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_element",
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

	if input.ID == "" {
		handlerErr = errors.New("id is required")
		return nil, GetElementOutput{}, handlerErr
	}

	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		handlerErr = fmt.Errorf("failed to get element: %w", err)
		return nil, GetElementOutput{}, handlerErr
	}

	// Check read permission
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()

	// Extract privacy fields from element (if it's a Persona, otherwise allow public access)
	owner := element.GetMetadata().Author
	privacyLevel := domain.PrivacyLevelPublic
	var sharedWith []string

	if persona, ok := element.(*domain.Persona); ok {
		privacyLevel = domain.PrivacyLevel(persona.PrivacyLevel)
		sharedWith = persona.SharedWith
	}

	if !accessControl.CheckReadPermission(userCtx, owner, privacyLevel, sharedWith) {
		handlerErr = errors.New("access denied: user does not have read permission")
		return nil, GetElementOutput{}, handlerErr
	}

	output := GetElementOutput{
		Element: element.GetMetadata().ToMap(),
	}

	// Record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "get_element", output)

	return nil, output, nil
}

// handleCreateElement handles create_element tool calls.
func (s *MCPServer) handleCreateElement(ctx context.Context, req *sdk.CallToolRequest, input CreateElementInput) (*sdk.CallToolResult, CreateElementOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "create_element",
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
	if input.Type == "" {
		handlerErr = errors.New("type is required")
		return nil, CreateElementOutput{}, handlerErr
	}
	if input.Name == "" || len(input.Name) < 3 || len(input.Name) > 100 {
		handlerErr = errors.New("name must be between 3 and 100 characters")
		return nil, CreateElementOutput{}, handlerErr
	}
	if len(input.Description) > 500 {
		handlerErr = errors.New("description must be at most 500 characters")
		return nil, CreateElementOutput{}, handlerErr
	}
	if input.Version == "" {
		handlerErr = errors.New("version is required")
		return nil, CreateElementOutput{}, handlerErr
	}
	if input.Author == "" {
		handlerErr = errors.New("author is required")
		return nil, CreateElementOutput{}, handlerErr
	}

	// Validate element type
	validTypes := map[string]bool{
		"persona":  true,
		"skill":    true,
		"template": true,
		"agent":    true,
		"memory":   true,
		"ensemble": true,
	}
	if !validTypes[input.Type] {
		handlerErr = fmt.Errorf("invalid element type: %s", input.Type)
		return nil, CreateElementOutput{}, handlerErr
	}

	// Generate ID
	id := uuid.New().String()
	now := time.Now()

	// Create metadata
	metadata := domain.ElementMetadata{
		ID:          id,
		Type:        domain.ElementType(input.Type),
		Name:        input.Name,
		Description: input.Description,
		Version:     input.Version,
		Author:      input.Author,
		Tags:        input.Tags,
		IsActive:    input.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create SimpleElement
	element := &SimpleElement{metadata: metadata}

	// Save to repository
	if err := s.repo.Create(element); err != nil {
		handlerErr = fmt.Errorf("failed to create element: %w", err)
		return nil, CreateElementOutput{}, handlerErr
	}

	output := CreateElementOutput{
		ID:      id,
		Element: metadata.ToMap(),
	}

	// Record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "create_element", output)

	return nil, output, nil
}

// handleUpdateElement handles update_element tool calls.
func (s *MCPServer) handleUpdateElement(ctx context.Context, req *sdk.CallToolRequest, input UpdateElementInput) (*sdk.CallToolResult, UpdateElementOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "update_element",
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

	if input.ID == "" {
		handlerErr = errors.New("id is required")
		return nil, UpdateElementOutput{}, handlerErr
	}

	// Get existing element
	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		handlerErr = fmt.Errorf("failed to get element: %w", err)
		return nil, UpdateElementOutput{}, handlerErr
	}

	// Check write permission
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()
	owner := element.GetMetadata().Author

	if !accessControl.CheckWritePermission(userCtx, owner) {
		handlerErr = errors.New("access denied: only the owner can update this element")
		return nil, UpdateElementOutput{}, handlerErr
	}

	metadata := element.GetMetadata()

	// Update fields
	updated := false

	if input.Name != "" && input.Name != metadata.Name {
		metadata.Name = input.Name
		updated = true
	}

	if input.Description != "" && input.Description != metadata.Description {
		metadata.Description = input.Description
		updated = true
	}

	if len(input.Tags) > 0 {
		metadata.Tags = input.Tags
		updated = true
	}

	if input.IsActive != nil && *input.IsActive != metadata.IsActive {
		metadata.IsActive = *input.IsActive
		updated = true
	}

	if updated {
		metadata.UpdatedAt = time.Now()

		// Create updated element
		updatedElement := &SimpleElement{metadata: metadata}

		if err := s.repo.Update(updatedElement); err != nil {
			handlerErr = fmt.Errorf("failed to update element: %w", err)
			return nil, UpdateElementOutput{}, handlerErr
		}
	}

	output := UpdateElementOutput{
		Element: metadata.ToMap(),
	}

	// Record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "update_element", output)

	return nil, output, nil
}

// handleDeleteElement handles delete_element tool calls.
func (s *MCPServer) handleDeleteElement(ctx context.Context, req *sdk.CallToolRequest, input DeleteElementInput) (*sdk.CallToolResult, DeleteElementOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "delete_element",
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

	if input.ID == "" {
		handlerErr = errors.New("id is required")
		return nil, DeleteElementOutput{}, handlerErr
	}

	// Get element to check permissions
	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		return nil, DeleteElementOutput{
			Success: false,
			Message: fmt.Sprintf("failed to get element: %v", err),
		}, nil
	}

	// Check delete permission
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()
	owner := element.GetMetadata().Author

	if !accessControl.CheckDeletePermission(userCtx, owner) {
		return nil, DeleteElementOutput{
			Success: false,
			Message: "access denied: only the owner can delete this element",
		}, nil
	}

	if err := s.repo.Delete(input.ID); err != nil {
		return nil, DeleteElementOutput{
			Success: false,
			Message: fmt.Sprintf("failed to delete element: %v", err),
		}, nil
	}

	output := DeleteElementOutput{
		Success: true,
		Message: fmt.Sprintf("Element %s deleted successfully", input.ID),
	}

	// Record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "delete_element", output)

	return nil, output, nil
}

// handleDuplicateElement handles duplicate_element tool calls.
func (s *MCPServer) handleDuplicateElement(ctx context.Context, req *sdk.CallToolRequest, input DuplicateElementInput) (*sdk.CallToolResult, DuplicateElementOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "duplicate_element",
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

	if input.ID == "" {
		handlerErr = errors.New("id is required")
		return nil, DuplicateElementOutput{}, handlerErr
	}

	// Get original element
	original, err := s.repo.GetByID(input.ID)
	if err != nil {
		handlerErr = fmt.Errorf("failed to get original element: %w", err)
		return nil, DuplicateElementOutput{}, handlerErr
	}

	// Check read permission on original
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()
	owner := original.GetMetadata().Author
	privacyLevel := domain.PrivacyLevelPublic
	var sharedWith []string

	if persona, ok := original.(*domain.Persona); ok {
		privacyLevel = domain.PrivacyLevel(persona.PrivacyLevel)
		sharedWith = persona.SharedWith
	}

	if !accessControl.CheckReadPermission(userCtx, owner, privacyLevel, sharedWith) {
		handlerErr = errors.New("access denied: user does not have read permission on original element")
		return nil, DuplicateElementOutput{}, handlerErr
	}

	// Create duplicate metadata
	originalMeta := original.GetMetadata()
	timestamp := time.Now().Format("20060102-150405")
	newID := fmt.Sprintf("%s-copy-%s", originalMeta.ID, timestamp)

	newName := input.NewName
	if newName == "" {
		newName = "Copy of " + originalMeta.Name
	}

	duplicateMeta := domain.ElementMetadata{
		ID:          newID,
		Type:        originalMeta.Type,
		Name:        newName,
		Description: originalMeta.Description,
		Version:     originalMeta.Version,
		Author:      originalMeta.Author,
		Tags:        append([]string{}, originalMeta.Tags...), // Copy tags
		IsActive:    originalMeta.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create duplicate element
	duplicate := &SimpleElement{metadata: duplicateMeta}

	// Save duplicate
	if err := s.repo.Create(duplicate); err != nil {
		handlerErr = fmt.Errorf("failed to create duplicate: %w", err)
		return nil, DuplicateElementOutput{}, handlerErr
	}

	output := DuplicateElementOutput{
		ID:      newID,
		Element: duplicateMeta.ToMap(),
		Message: fmt.Sprintf("Element duplicated successfully: %s -> %s", input.ID, newID),
	}

	// Record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "duplicate_element", output)

	return nil, output, nil
}
