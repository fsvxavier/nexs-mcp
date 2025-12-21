package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// handleGetPerformanceDashboard handles the get_performance_dashboard tool call.
func (s *MCPServer) handleGetPerformanceDashboard(ctx context.Context, req *sdk.CallToolRequest, input GetPerformanceDashboardInput) (*sdk.CallToolResult, GetPerformanceDashboardOutput, error) {
	// Default to last 24 hours if no period specified
	period := input.Period
	if period == "" {
		period = "last_24h"
	}

	// Get dashboard from performance metrics
	dashboard := s.perfMetrics.GetDashboard(period)

	// Convert slow operations to map format
	slowOps := make([]map[string]interface{}, 0, len(dashboard.SlowOperations))
	for _, op := range dashboard.SlowOperations {
		slowOps = append(slowOps, map[string]interface{}{
			"operation":   op.Operation,
			"duration_ms": op.Duration,
			"timestamp":   op.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Convert fast operations to map format
	fastOps := make([]map[string]interface{}, 0, len(dashboard.FastOperations))
	for _, op := range dashboard.FastOperations {
		fastOps = append(fastOps, map[string]interface{}{
			"operation":   op.Operation,
			"duration_ms": op.Duration,
			"timestamp":   op.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Convert per-operation stats to map format
	byOperation := make(map[string]map[string]interface{})
	for op, stats := range dashboard.ByOperation {
		byOperation[op] = map[string]interface{}{
			"count":        stats.Count,
			"avg_duration": stats.AvgDuration,
			"max_duration": stats.MaxDuration,
			"min_duration": stats.MinDuration,
		}
	}

	output := GetPerformanceDashboardOutput{
		TotalOperations: dashboard.TotalOperations,
		AvgDuration:     dashboard.AvgDuration,
		P50Duration:     dashboard.P50Duration,
		P95Duration:     dashboard.P95Duration,
		P99Duration:     dashboard.P99Duration,
		MaxDuration:     dashboard.MaxDuration,
		MinDuration:     dashboard.MinDuration,
		SlowOperations:  slowOps,
		FastOperations:  fastOps,
		ByOperation:     byOperation,
		Period:          period,
	}

	return nil, output, nil
}
