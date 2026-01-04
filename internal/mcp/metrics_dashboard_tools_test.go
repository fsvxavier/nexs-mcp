package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetricsDashboard(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetMetricsDashboardInput{}

	_, output, err := server.handleGetMetricsDashboard(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.NotNil(t, output.PerformanceMetrics)
	assert.NotNil(t, output.TokenMetrics)
}

func TestMetricsDashboardSections(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetMetricsDashboardInput{}

	_, output, err := server.handleGetMetricsDashboard(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output.PerformanceMetrics)
	assert.NotNil(t, output.TokenMetrics)
	assert.NotNil(t, output.Summary)
	assert.GreaterOrEqual(t, output.HealthScore, float64(0))
}

func TestMetricsLatencyPercentiles(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetMetricsDashboardInput{}

	_, output, err := server.handleGetMetricsDashboard(ctx, nil, input)

	require.NoError(t, err)
	perf := output.PerformanceMetrics

	// Validate duration metrics
	assert.GreaterOrEqual(t, perf.P95DurationMs, float64(0))
	assert.GreaterOrEqual(t, perf.AvgDurationMs, float64(0))
}

func TestMetricsSuccessRate(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetMetricsDashboardInput{}

	_, output, err := server.handleGetMetricsDashboard(ctx, nil, input)

	require.NoError(t, err)
	perf := output.PerformanceMetrics

	assert.GreaterOrEqual(t, perf.SuccessRate, float64(0))
	assert.LessOrEqual(t, perf.SuccessRate, float64(100))
}

func TestMetricsTopTools(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetMetricsDashboardInput{
		TopN: 10,
	}

	_, output, err := server.handleGetMetricsDashboard(ctx, nil, input)

	require.NoError(t, err)
	perf := output.PerformanceMetrics
	assert.LessOrEqual(t, len(perf.TopToolsByUsage), 10)
}

func TestMetricsSlowestOperations(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetMetricsDashboardInput{}

	_, output, err := server.handleGetMetricsDashboard(ctx, nil, input)

	require.NoError(t, err)
	perf := output.PerformanceMetrics
	// TopToolsByDuration can be empty slice or nil depending on metrics availability
	// Just check it doesn't panic and has reasonable length
	if perf.TopToolsByDuration != nil {
		assert.LessOrEqual(t, len(perf.TopToolsByDuration), 10)
	}
}
