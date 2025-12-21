package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// handleGetUsageStats handles get_usage_stats tool calls.
func (s *MCPServer) handleGetUsageStats(ctx context.Context, req *sdk.CallToolRequest, input GetUsageStatsInput) (*sdk.CallToolResult, GetUsageStatsOutput, error) {
	// Default period
	period := input.Period
	if period == "" {
		period = "last_24h"
	}

	// Get statistics from metrics collector
	stats, err := s.metrics.GetStatistics(period)
	if err != nil {
		return nil, GetUsageStatsOutput{}, err
	}

	// Convert most used tools
	mostUsedTools := make([]map[string]interface{}, len(stats.MostUsedTools))
	for i, tool := range stats.MostUsedTools {
		mostUsedTools[i] = map[string]interface{}{
			"tool_name":    tool.ToolName,
			"count":        tool.Count,
			"success_rate": tool.SuccessRate,
			"avg_duration": tool.AvgDuration,
		}
	}

	// Convert slowest operations
	slowestOps := make([]map[string]interface{}, len(stats.SlowestOperations))
	for i, op := range stats.SlowestOperations {
		slowestOps[i] = map[string]interface{}{
			"tool_name": op.ToolName,
			"timestamp": op.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			"duration":  op.Duration.Milliseconds(),
			"success":   op.Success,
			"user":      op.User,
		}
	}

	// Convert recent errors
	recentErrors := make([]map[string]interface{}, len(stats.RecentErrors))
	for i, err := range stats.RecentErrors {
		recentErrors[i] = map[string]interface{}{
			"tool_name":     err.ToolName,
			"timestamp":     err.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			"error_message": err.ErrorMessage,
			"user":          err.User,
		}
	}

	output := GetUsageStatsOutput{
		TotalOperations:    stats.TotalOperations,
		SuccessfulOps:      stats.SuccessfulOps,
		FailedOps:          stats.FailedOps,
		SuccessRate:        stats.SuccessRate,
		OperationsByTool:   stats.OperationsByTool,
		ErrorsByTool:       stats.ErrorsByTool,
		AvgDurationByTool:  stats.AvgDurationByTool,
		MostUsedTools:      mostUsedTools,
		SlowestOperations:  slowestOps,
		RecentErrors:       recentErrors,
		ActiveUsers:        stats.ActiveUsers,
		OperationsByPeriod: stats.OperationsByPeriod,
		Period:             stats.Period,
		StartTime:          stats.StartTime.Format("2006-01-02T15:04:05Z07:00"),
		EndTime:            stats.EndTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	return nil, output, nil
}
