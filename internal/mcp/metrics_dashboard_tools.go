package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Constants for metrics types and directions.
const (
	metricsTypeBoth          = "both"
	trendDirectionStable     = "stable"
	trendDirectionIncreasing = "increasing"
	severityCritical         = "critical"
)

// GetMetricsDashboardInput represents the input for get_metrics_dashboard tool.
type GetMetricsDashboardInput struct {
	Period           string `json:"period,omitempty"            jsonschema_description:"Time period: 24h, 7d, 30d, all (default: 24h)"`
	TopN             int    `json:"top_n,omitempty"             jsonschema_description:"Number of top tools to show (default: 10)"`
	MetricsType      string `json:"metrics_type,omitempty"      jsonschema_description:"Type of metrics: performance, token, both (default: both)"`
	IncludeAnomalies bool   `json:"include_anomalies,omitempty" jsonschema_description:"Include anomaly detection analysis (default: true)"`
	IncludeTrends    bool   `json:"include_trends,omitempty"    jsonschema_description:"Include historical trend analysis (default: true)"`
	IncludeAlerts    bool   `json:"include_alerts,omitempty"    jsonschema_description:"Include active alerts and warnings (default: true)"`
}

// GetMetricsDashboardOutput represents the output of get_metrics_dashboard tool.
type GetMetricsDashboardOutput struct {
	Period             string                 `json:"period"`
	GeneratedAt        string                 `json:"generated_at"`
	PerformanceMetrics *PerformanceMetrics    `json:"performance_metrics,omitempty"`
	TokenMetrics       *TokenMetrics          `json:"token_metrics,omitempty"`
	Summary            map[string]interface{} `json:"summary"`
	Anomalies          []DashboardAnomaly     `json:"anomalies,omitempty"`
	Trends             *DashboardTrends       `json:"trends,omitempty"`
	Alerts             []DashboardAlert       `json:"alerts,omitempty"`
	HealthScore        float64                `json:"health_score"`
	Status             string                 `json:"status"`
}

// DashboardAnomaly represents a detected anomaly in metrics.
type DashboardAnomaly struct {
	Type          string    `json:"type"`
	Severity      string    `json:"severity"`
	Description   string    `json:"description"`
	Value         float64   `json:"value"`
	Threshold     float64   `json:"threshold"`
	DetectedAt    time.Time `json:"detected_at"`
	AffectedTools []string  `json:"affected_tools,omitempty"`
}

// DashboardTrends represents historical trend analysis.
type DashboardTrends struct {
	OperationsTrend  TrendData `json:"operations_trend"`
	DurationTrend    TrendData `json:"duration_trend"`
	ErrorRateTrend   TrendData `json:"error_rate_trend"`
	TokenUsageTrend  TrendData `json:"token_usage_trend"`
	CompressionTrend TrendData `json:"compression_trend"`
}

// TrendData represents a single metric trend.
type TrendData struct {
	Direction     string  `json:"direction"`
	ChangePercent float64 `json:"change_percent"`
	Current       float64 `json:"current"`
	Previous      float64 `json:"previous"`
	Confidence    float64 `json:"confidence"`
}

// DashboardAlert represents an active alert.
type DashboardAlert struct {
	Level       string    `json:"level"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Action      string    `json:"action"`
	TriggeredAt time.Time `json:"triggered_at"`
	ToolName    string    `json:"tool_name,omitempty"`
}

// PerformanceMetrics represents performance statistics.
type PerformanceMetrics struct {
	TotalCalls         int                 `json:"total_calls"`
	SuccessRate        float64             `json:"success_rate"`
	AvgDurationMs      float64             `json:"avg_duration_ms"`
	P95DurationMs      float64             `json:"p95_duration_ms"`
	TopToolsByUsage    []ToolUsageStats    `json:"top_tools_by_usage"`
	TopToolsByDuration []ToolDurationStats `json:"top_tools_by_duration"`
	TopErrors          []ErrorStats        `json:"top_errors"`
}

// TokenMetrics represents token optimization statistics.
type TokenMetrics struct {
	TotalOriginalTokens  int                 `json:"total_original_tokens"`
	TotalOptimizedTokens int                 `json:"total_optimized_tokens"`
	TotalTokensSaved     int                 `json:"total_tokens_saved"`
	AvgCompressionRatio  float64             `json:"avg_compression_ratio"`
	TotalOptimizations   int                 `json:"total_optimizations"`
	TopToolsBySavings    []TokenSavingsStats `json:"top_tools_by_savings"`
	OptimizationTypes    map[string]int      `json:"optimization_types"`
}

// ToolUsageStats represents usage statistics for a tool.
type ToolUsageStats struct {
	ToolName    string  `json:"tool_name"`
	CallCount   int     `json:"call_count"`
	SuccessRate float64 `json:"success_rate"`
	AvgDuration float64 `json:"avg_duration_ms"`
}

// ToolDurationStats represents duration statistics for a tool.
type ToolDurationStats struct {
	ToolName    string  `json:"tool_name"`
	AvgDuration float64 `json:"avg_duration_ms"`
	P95Duration float64 `json:"p95_duration_ms"`
	CallCount   int     `json:"call_count"`
}

// ErrorStats represents error statistics.
type ErrorStats struct {
	ErrorMessage string   `json:"error_message"`
	Count        int      `json:"count"`
	ToolNames    []string `json:"tool_names"`
}

// TokenSavingsStats represents token savings statistics for a tool.
type TokenSavingsStats struct {
	ToolName          string  `json:"tool_name"`
	TotalTokensSaved  int     `json:"total_tokens_saved"`
	AvgCompression    float64 `json:"avg_compression_ratio"`
	OptimizationCount int     `json:"optimization_count"`
}

// RegisterMetricsDashboardTools registers metrics dashboard tools with the MCP server.
func (s *MCPServer) RegisterMetricsDashboardTools() {
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_metrics_dashboard",
		Description: "Get comprehensive metrics dashboard with performance and token optimization statistics",
	}, s.handleGetMetricsDashboard)
}

// handleGetMetricsDashboard handles get_metrics_dashboard tool calls.
func (s *MCPServer) handleGetMetricsDashboard(ctx context.Context, req *sdk.CallToolRequest, input GetMetricsDashboardInput) (*sdk.CallToolResult, GetMetricsDashboardOutput, error) {
	startTime := time.Now()
	var err error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_metrics_dashboard",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   err == nil,
			ErrorMessage: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		})
	}()

	// Set defaults
	if input.Period == "" {
		input.Period = "24h"
	}
	if input.TopN == 0 {
		input.TopN = 10
	}
	if input.MetricsType == "" {
		input.MetricsType = metricsTypeBoth
	}
	// Enable all enhanced features by default
	if !input.IncludeAnomalies {
		input.IncludeAnomalies = true
	}
	if !input.IncludeTrends {
		input.IncludeTrends = true
	}
	if !input.IncludeAlerts {
		input.IncludeAlerts = true
	}

	// Validate inputs
	validPeriods := map[string]bool{"24h": true, "7d": true, "30d": true, "all": true}
	if !validPeriods[input.Period] {
		err = fmt.Errorf("invalid period: %s (must be 24h, 7d, 30d, or all)", input.Period)
		return nil, GetMetricsDashboardOutput{}, err
	}

	validTypes := map[string]bool{"performance": true, "token": true, "both": true}
	if !validTypes[input.MetricsType] {
		err = fmt.Errorf("invalid metrics_type: %s (must be performance, token, or both)", input.MetricsType)
		return nil, GetMetricsDashboardOutput{}, err
	}

	// Calculate time cutoff
	var cutoff time.Time
	now := time.Now()
	switch input.Period {
	case "24h":
		cutoff = now.Add(-24 * time.Hour)
	case "7d":
		cutoff = now.Add(-7 * 24 * time.Hour)
	case "30d":
		cutoff = now.Add(-30 * 24 * time.Hour)
	case "all":
		cutoff = time.Time{} // Zero value, no filtering
	}

	output := GetMetricsDashboardOutput{
		Period:      input.Period,
		GeneratedAt: now.Format(time.RFC3339),
		Summary:     make(map[string]interface{}),
		Anomalies:   []DashboardAnomaly{},
		Alerts:      []DashboardAlert{},
	}

	// Load performance metrics
	if input.MetricsType == "performance" || input.MetricsType == metricsTypeBoth {
		perfMetrics, loadErr := s.loadPerformanceMetrics(cutoff, input.TopN)
		if loadErr != nil {
			err = fmt.Errorf("failed to load performance metrics: %w", loadErr)
			return nil, GetMetricsDashboardOutput{}, err
		}
		output.PerformanceMetrics = perfMetrics
		output.Summary["total_tool_calls"] = perfMetrics.TotalCalls
		output.Summary["overall_success_rate"] = fmt.Sprintf("%.2f%%", perfMetrics.SuccessRate*100)
		output.Summary["avg_duration_ms"] = perfMetrics.AvgDurationMs
	}

	// Load token metrics
	if input.MetricsType == "token" || input.MetricsType == metricsTypeBoth {
		tokenMetrics, loadErr := s.loadTokenMetrics(cutoff, input.TopN)
		if loadErr != nil {
			err = fmt.Errorf("failed to load token metrics: %w", loadErr)
			return nil, GetMetricsDashboardOutput{}, err
		}
		output.TokenMetrics = tokenMetrics
		output.Summary["total_tokens_saved"] = tokenMetrics.TotalTokensSaved
		output.Summary["avg_compression_ratio"] = fmt.Sprintf("%.2f%%", (1-tokenMetrics.AvgCompressionRatio)*100)
		output.Summary["total_optimizations"] = tokenMetrics.TotalOptimizations
	}

	// Detect anomalies
	if input.IncludeAnomalies && output.PerformanceMetrics != nil {
		anomalies := s.detectDashboardAnomalies(output.PerformanceMetrics, output.TokenMetrics)
		output.Anomalies = anomalies
		output.Summary["anomaly_count"] = len(anomalies)
	}

	// Analyze trends (if we have enough historical data)
	if input.IncludeTrends {
		trends, trendErr := s.analyzeDashboardTrends(cutoff, output.PerformanceMetrics, output.TokenMetrics)
		if trendErr == nil {
			output.Trends = trends
		}
	}

	// Generate alerts
	if input.IncludeAlerts {
		alerts := s.generateDashboardAlerts(output.PerformanceMetrics, output.TokenMetrics, output.Anomalies)
		output.Alerts = alerts
		output.Summary["active_alerts"] = len(alerts)
	}

	// Calculate health score and status
	healthScore, status := s.calculateSystemHealth(output.PerformanceMetrics, output.TokenMetrics, output.Anomalies)
	output.HealthScore = healthScore
	output.Status = status
	output.Summary["health_score"] = fmt.Sprintf("%.1f/100", healthScore)
	output.Summary["status"] = status

	return nil, output, nil
}

// loadPerformanceMetrics loads and analyzes performance metrics from disk.
func (s *MCPServer) loadPerformanceMetrics(cutoff time.Time, topN int) (*PerformanceMetrics, error) {
	metricsDir := filepath.Join(s.cfg.BaseDir, "metrics")
	metricsFile := filepath.Join(metricsDir, "metrics.json")

	data, err := os.ReadFile(metricsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &PerformanceMetrics{}, nil // No metrics yet
		}
		return nil, fmt.Errorf("failed to read metrics file: %w", err)
	}

	var toolMetrics []application.ToolCallMetric
	if err := json.Unmarshal(data, &toolMetrics); err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", err)
	}

	// Filter by time period
	var filtered []application.ToolCallMetric
	for _, m := range toolMetrics {
		if cutoff.IsZero() || m.Timestamp.After(cutoff) {
			filtered = append(filtered, m)
		}
	}

	if len(filtered) == 0 {
		return &PerformanceMetrics{}, nil
	}

	// Calculate statistics
	totalCalls := len(filtered)
	successCount := 0
	totalDuration := int64(0)
	durations := make([]int64, 0, totalCalls)
	toolStatsMap := make(map[string]*toolStatsData)
	errorCounts := make(map[string]*errorCount)

	for _, m := range filtered {
		if m.Success {
			successCount++
		}
		durationMs := m.Duration.Milliseconds()
		totalDuration += durationMs
		durations = append(durations, durationMs)

		// Tool stats
		if toolStatsMap[m.ToolName] == nil {
			toolStatsMap[m.ToolName] = &toolStatsData{}
		}
		stats := toolStatsMap[m.ToolName]
		stats.callCount++
		if m.Success {
			stats.successCount++
		}
		stats.totalDuration += durationMs
		stats.durations = append(stats.durations, durationMs)

		// Error stats
		if !m.Success && m.ErrorMessage != "" {
			if errorCounts[m.ErrorMessage] == nil {
				errorCounts[m.ErrorMessage] = &errorCount{toolNames: make(map[string]bool)}
			}
			errorCounts[m.ErrorMessage].count++
			errorCounts[m.ErrorMessage].toolNames[m.ToolName] = true
		}
	}

	// Calculate P95 duration
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95Index := int(float64(len(durations)) * 0.95)
	if p95Index >= len(durations) {
		p95Index = len(durations) - 1
	}

	result := &PerformanceMetrics{
		TotalCalls:    totalCalls,
		SuccessRate:   float64(successCount) / float64(totalCalls),
		AvgDurationMs: float64(totalDuration) / float64(totalCalls),
		P95DurationMs: float64(durations[p95Index]),
	}

	// Top tools by usage
	usageList := make([]ToolUsageStats, 0, len(toolStatsMap))
	for toolName, stats := range toolStatsMap {
		usageList = append(usageList, ToolUsageStats{
			ToolName:    toolName,
			CallCount:   stats.callCount,
			SuccessRate: float64(stats.successCount) / float64(stats.callCount),
			AvgDuration: float64(stats.totalDuration) / float64(stats.callCount),
		})
	}
	sort.Slice(usageList, func(i, j int) bool { return usageList[i].CallCount > usageList[j].CallCount })
	if len(usageList) > topN {
		usageList = usageList[:topN]
	}
	result.TopToolsByUsage = usageList

	// Top tools by duration
	durationList := make([]ToolDurationStats, 0, len(toolStatsMap))
	for toolName, stats := range toolStatsMap {
		sort.Slice(stats.durations, func(i, j int) bool { return stats.durations[i] < stats.durations[j] })
		p95Idx := int(float64(len(stats.durations)) * 0.95)
		if p95Idx >= len(stats.durations) {
			p95Idx = len(stats.durations) - 1
		}
		durationList = append(durationList, ToolDurationStats{
			ToolName:    toolName,
			AvgDuration: float64(stats.totalDuration) / float64(stats.callCount),
			P95Duration: float64(stats.durations[p95Idx]),
			CallCount:   stats.callCount,
		})
	}
	sort.Slice(durationList, func(i, j int) bool { return durationList[i].AvgDuration > durationList[j].AvgDuration })
	if len(durationList) > topN {
		durationList = durationList[:topN]
	}
	result.TopToolsByDuration = durationList

	// Top errors
	errorList := make([]ErrorStats, 0, len(errorCounts))
	for msg, ec := range errorCounts {
		toolNames := make([]string, 0, len(ec.toolNames))
		for tn := range ec.toolNames {
			toolNames = append(toolNames, tn)
		}
		errorList = append(errorList, ErrorStats{
			ErrorMessage: msg,
			Count:        ec.count,
			ToolNames:    toolNames,
		})
	}
	sort.Slice(errorList, func(i, j int) bool { return errorList[i].Count > errorList[j].Count })
	if len(errorList) > topN {
		errorList = errorList[:topN]
	}
	result.TopErrors = errorList

	return result, nil
}

// loadTokenMetrics loads and analyzes token metrics from disk.
func (s *MCPServer) loadTokenMetrics(cutoff time.Time, topN int) (*TokenMetrics, error) {
	metricsDir := filepath.Join(s.cfg.BaseDir, "token_metrics")
	metricsFile := filepath.Join(metricsDir, "token_metrics.json")

	data, err := os.ReadFile(metricsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &TokenMetrics{OptimizationTypes: make(map[string]int)}, nil // No metrics yet
		}
		return nil, fmt.Errorf("failed to read token metrics file: %w", err)
	}

	var tokenMetrics []application.TokenMetrics
	if err := json.Unmarshal(data, &tokenMetrics); err != nil {
		return nil, fmt.Errorf("failed to parse token metrics: %w", err)
	}

	// Filter by time period
	var filtered []application.TokenMetrics
	for _, m := range tokenMetrics {
		if cutoff.IsZero() || m.Timestamp.After(cutoff) {
			filtered = append(filtered, m)
		}
	}

	if len(filtered) == 0 {
		return &TokenMetrics{OptimizationTypes: make(map[string]int)}, nil
	}

	// Calculate statistics
	totalOriginal := int64(0)
	totalOptimized := int64(0)
	totalSaved := int64(0)
	totalRatio := 0.0
	optimizationTypes := make(map[string]int)
	toolSavings := make(map[string]*tokenToolStats)

	for _, m := range filtered {
		totalOriginal += m.OriginalTokens
		totalOptimized += m.OptimizedTokens
		totalSaved += m.TokensSaved
		totalRatio += m.CompressionRatio
		optimizationTypes[m.OptimizationType]++

		if toolSavings[m.ToolName] == nil {
			toolSavings[m.ToolName] = &tokenToolStats{}
		}
		stats := toolSavings[m.ToolName]
		stats.totalSaved += m.TokensSaved
		stats.totalRatio += m.CompressionRatio
		stats.count++
	}

	result := &TokenMetrics{
		TotalOriginalTokens:  int(totalOriginal),
		TotalOptimizedTokens: int(totalOptimized),
		TotalTokensSaved:     int(totalSaved),
		AvgCompressionRatio:  totalRatio / float64(len(filtered)),
		TotalOptimizations:   len(filtered),
		OptimizationTypes:    optimizationTypes,
	}

	// Top tools by savings
	savingsList := make([]TokenSavingsStats, 0, len(toolSavings))
	for toolName, stats := range toolSavings {
		savingsList = append(savingsList, TokenSavingsStats{
			ToolName:          toolName,
			TotalTokensSaved:  int(stats.totalSaved),
			AvgCompression:    stats.totalRatio / float64(stats.count),
			OptimizationCount: stats.count,
		})
	}
	sort.Slice(savingsList, func(i, j int) bool { return savingsList[i].TotalTokensSaved > savingsList[j].TotalTokensSaved })
	if len(savingsList) > topN {
		savingsList = savingsList[:topN]
	}
	result.TopToolsBySavings = savingsList

	return result, nil
}

// Helper structs for aggregation.
type toolStatsData struct {
	callCount     int
	successCount  int
	totalDuration int64
	durations     []int64
}

type errorCount struct {
	count     int
	toolNames map[string]bool
}

type tokenToolStats struct {
	totalSaved int64
	totalRatio float64
	count      int
}

// detectDashboardAnomalies detects anomalies in current metrics.
func (s *MCPServer) detectDashboardAnomalies(perfMetrics *PerformanceMetrics, tokenMetrics *TokenMetrics) []DashboardAnomaly {
	anomalies := []DashboardAnomaly{}
	now := time.Now()

	if perfMetrics != nil {
		// Anomaly 1: Low success rate
		if perfMetrics.SuccessRate < 0.95 {
			affectedTools := []string{}
			for _, tool := range perfMetrics.TopToolsByUsage {
				if tool.SuccessRate < 0.95 {
					affectedTools = append(affectedTools, tool.ToolName)
				}
			}
			anomalies = append(anomalies, DashboardAnomaly{
				Type:          "low_success_rate",
				Severity:      "high",
				Description:   fmt.Sprintf("System success rate is %.1f%%, below threshold of 95%%", perfMetrics.SuccessRate*100),
				Value:         perfMetrics.SuccessRate * 100,
				Threshold:     95.0,
				DetectedAt:    now,
				AffectedTools: affectedTools,
			})
		}

		// Anomaly 2: High P95 latency
		if perfMetrics.P95DurationMs > 1000 {
			affectedTools := []string{}
			for _, tool := range perfMetrics.TopToolsByDuration {
				if tool.P95Duration > 1000 {
					affectedTools = append(affectedTools, tool.ToolName)
				}
			}
			anomalies = append(anomalies, DashboardAnomaly{
				Type:          "high_p95_latency",
				Severity:      "medium",
				Description:   fmt.Sprintf("P95 latency is %.0fms, above threshold of 1000ms", perfMetrics.P95DurationMs),
				Value:         perfMetrics.P95DurationMs,
				Threshold:     1000.0,
				DetectedAt:    now,
				AffectedTools: affectedTools,
			})
		}

		// Anomaly 3: Excessive errors
		if len(perfMetrics.TopErrors) > 0 {
			totalErrors := 0
			for _, errStat := range perfMetrics.TopErrors {
				totalErrors += errStat.Count
			}
			if totalErrors > perfMetrics.TotalCalls/10 { // >10% error rate
				anomalies = append(anomalies, DashboardAnomaly{
					Type:        "excessive_errors",
					Severity:    "critical",
					Description: fmt.Sprintf("Excessive errors detected: %d errors in %d calls (%.1f%%)", totalErrors, perfMetrics.TotalCalls, float64(totalErrors)/float64(perfMetrics.TotalCalls)*100),
					Value:       float64(totalErrors) / float64(perfMetrics.TotalCalls) * 100,
					Threshold:   10.0,
					DetectedAt:  now,
				})
			}
		}
	}

	if tokenMetrics != nil {
		// Anomaly 4: Low compression efficiency
		if tokenMetrics.TotalOriginalTokens > 10000 && tokenMetrics.AvgCompressionRatio > 0.9 {
			anomalies = append(anomalies, DashboardAnomaly{
				Type:        "low_compression",
				Severity:    "medium",
				Description: fmt.Sprintf("Compression efficiency is low: %.1f%% average compression ratio", tokenMetrics.AvgCompressionRatio*100),
				Value:       tokenMetrics.AvgCompressionRatio * 100,
				Threshold:   90.0,
				DetectedAt:  now,
			})
		}
	}

	return anomalies
}

// analyzeDashboardTrends analyzes historical trends by comparing current period with previous.
func (s *MCPServer) analyzeDashboardTrends(cutoff time.Time, perfMetrics *PerformanceMetrics, tokenMetrics *TokenMetrics) (*DashboardTrends, error) {
	// Calculate previous period cutoff (same duration before current cutoff)
	now := time.Now()
	periodDuration := now.Sub(cutoff)
	previousCutoff := cutoff.Add(-periodDuration)

	// Load previous period metrics
	prevPerfMetrics, err := s.loadPerformanceMetrics(previousCutoff, 10)
	if err != nil {
		return nil, err
	}
	prevTokenMetrics, err := s.loadTokenMetrics(previousCutoff, 10)
	if err != nil {
		return nil, err
	}

	trends := &DashboardTrends{}

	// Operations trend
	if perfMetrics != nil && prevPerfMetrics != nil {
		if prevPerfMetrics.TotalCalls > 0 {
			changePercent := ((float64(perfMetrics.TotalCalls) - float64(prevPerfMetrics.TotalCalls)) / float64(prevPerfMetrics.TotalCalls)) * 100
			direction := trendDirectionStable
			if changePercent > 5 {
				direction = trendDirectionIncreasing
			} else if changePercent < -5 {
				direction = "decreasing"
			}
			trends.OperationsTrend = TrendData{
				Direction:     direction,
				ChangePercent: changePercent,
				Current:       float64(perfMetrics.TotalCalls),
				Previous:      float64(prevPerfMetrics.TotalCalls),
				Confidence:    0.85,
			}
		}

		// Duration trend
		if prevPerfMetrics.AvgDurationMs > 0 {
			changePercent := ((perfMetrics.AvgDurationMs - prevPerfMetrics.AvgDurationMs) / prevPerfMetrics.AvgDurationMs) * 100
			direction := trendDirectionStable
			if changePercent > 10 {
				direction = "degrading"
			} else if changePercent < -10 {
				direction = "improving"
			}
			trends.DurationTrend = TrendData{
				Direction:     direction,
				ChangePercent: changePercent,
				Current:       perfMetrics.AvgDurationMs,
				Previous:      prevPerfMetrics.AvgDurationMs,
				Confidence:    0.80,
			}
		}

		// Error rate trend
		currentErrorRate := 1.0 - perfMetrics.SuccessRate
		previousErrorRate := 1.0 - prevPerfMetrics.SuccessRate
		if previousErrorRate > 0 {
			changePercent := ((currentErrorRate - previousErrorRate) / previousErrorRate) * 100
			direction := trendDirectionStable
			if changePercent > 20 {
				direction = "worsening"
			} else if changePercent < -20 {
				direction = "improving"
			}
			trends.ErrorRateTrend = TrendData{
				Direction:     direction,
				ChangePercent: changePercent,
				Current:       currentErrorRate * 100,
				Previous:      previousErrorRate * 100,
				Confidence:    0.75,
			}
		}
	}

	// Token usage and compression trends
	if tokenMetrics != nil && prevTokenMetrics != nil {
		if prevTokenMetrics.TotalOriginalTokens > 0 {
			changePercent := ((float64(tokenMetrics.TotalOriginalTokens) - float64(prevTokenMetrics.TotalOriginalTokens)) / float64(prevTokenMetrics.TotalOriginalTokens)) * 100
			direction := trendDirectionStable
			if changePercent > 10 {
				direction = "increasing"
			} else if changePercent < -10 {
				direction = "decreasing"
			}
			trends.TokenUsageTrend = TrendData{
				Direction:     direction,
				ChangePercent: changePercent,
				Current:       float64(tokenMetrics.TotalOriginalTokens),
				Previous:      float64(prevTokenMetrics.TotalOriginalTokens),
				Confidence:    0.80,
			}
		}

		// Compression trend (lower is better)
		if prevTokenMetrics.AvgCompressionRatio > 0 {
			changePercent := ((tokenMetrics.AvgCompressionRatio - prevTokenMetrics.AvgCompressionRatio) / prevTokenMetrics.AvgCompressionRatio) * 100
			direction := trendDirectionStable
			if changePercent < -5 {
				direction = "improving" // Lower compression ratio = more savings
			} else if changePercent > 5 {
				direction = "degrading"
			}
			trends.CompressionTrend = TrendData{
				Direction:     direction,
				ChangePercent: changePercent,
				Current:       tokenMetrics.AvgCompressionRatio * 100,
				Previous:      prevTokenMetrics.AvgCompressionRatio * 100,
				Confidence:    0.75,
			}
		}
	}

	return trends, nil
}

// generateDashboardAlerts generates actionable alerts based on metrics and anomalies.
func (s *MCPServer) generateDashboardAlerts(perfMetrics *PerformanceMetrics, tokenMetrics *TokenMetrics, anomalies []DashboardAnomaly) []DashboardAlert {
	alerts := []DashboardAlert{}
	now := time.Now()

	// Convert critical/high severity anomalies to alerts
	for _, anomaly := range anomalies {
		if anomaly.Severity == severityCritical || anomaly.Severity == "high" {
			level := "warning"
			if anomaly.Severity == severityCritical {
				level = "critical"
			}

			action := "Review error logs and investigate affected tools"
			switch anomaly.Type {
			case "low_success_rate":
				action = "Investigate failing operations and check error messages"
			case "high_p95_latency":
				action = "Profile slow operations and consider optimization"
			}

			alerts = append(alerts, DashboardAlert{
				Level:       level,
				Title:       anomaly.Type,
				Message:     anomaly.Description,
				Action:      action,
				TriggeredAt: anomaly.DetectedAt,
			})
		}
	}

	// Additional proactive alerts
	if perfMetrics != nil {
		// Alert on very slow tools
		for _, tool := range perfMetrics.TopToolsByDuration {
			if tool.AvgDuration > 2000 { // >2 seconds
				alerts = append(alerts, DashboardAlert{
					Level:       "warning",
					Title:       "Slow Tool Detected",
					Message:     fmt.Sprintf("Tool '%s' has average duration of %.0fms", tool.ToolName, tool.AvgDuration),
					Action:      "Consider caching, optimization, or async processing",
					TriggeredAt: now,
					ToolName:    tool.ToolName,
				})
			}
		}
	}

	if tokenMetrics != nil {
		// Alert on inefficient token usage
		for _, tool := range tokenMetrics.TopToolsBySavings {
			if tool.AvgCompression > 0.8 && tool.OptimizationCount > 10 {
				alerts = append(alerts, DashboardAlert{
					Level:       "info",
					Title:       "Low Compression Efficiency",
					Message:     fmt.Sprintf("Tool '%s' has compression ratio of %.1f%%", tool.ToolName, tool.AvgCompression*100),
					Action:      "Review compression settings or data structure",
					TriggeredAt: now,
					ToolName:    tool.ToolName,
				})
			}
		}
	}

	return alerts
}

// calculateSystemHealth calculates overall system health score (0-100).
func (s *MCPServer) calculateSystemHealth(perfMetrics *PerformanceMetrics, tokenMetrics *TokenMetrics, anomalies []DashboardAnomaly) (float64, string) {
	score := 100.0

	if perfMetrics != nil {
		// Deduct for low success rate
		if perfMetrics.SuccessRate < 0.99 {
			score -= (0.99 - perfMetrics.SuccessRate) * 200 // Max -20 points
		}

		// Deduct for high latency
		if perfMetrics.AvgDurationMs > 100 {
			latencyPenalty := (perfMetrics.AvgDurationMs - 100) / 100 * 10
			if latencyPenalty > 20 {
				latencyPenalty = 20
			}
			score -= latencyPenalty
		}

		// Deduct for high P95
		if perfMetrics.P95DurationMs > 500 {
			p95Penalty := (perfMetrics.P95DurationMs - 500) / 500 * 10
			if p95Penalty > 15 {
				p95Penalty = 15
			}
			score -= p95Penalty
		}
	}

	if tokenMetrics != nil {
		// Deduct for poor compression
		if tokenMetrics.AvgCompressionRatio > 0.7 && tokenMetrics.TotalOptimizations > 0 {
			compressionPenalty := (tokenMetrics.AvgCompressionRatio - 0.7) * 50
			if compressionPenalty > 15 {
				compressionPenalty = 15
			}
			score -= compressionPenalty
		}
	}

	// Deduct for anomalies
	for _, anomaly := range anomalies {
		switch anomaly.Severity {
		case "critical":
			score -= 15
		case "high":
			score -= 10
		case "medium":
			score -= 5
		case "low":
			score -= 2
		}
	}

	// Ensure score is in range [0, 100]
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	// Determine status
	var status string
	switch {
	case score < 50:
		status = "critical"
	case score < 70:
		status = "degraded"
	case score < 90:
		status = "warning"
	default:
		status = "healthy"
	}

	return score, status
}
