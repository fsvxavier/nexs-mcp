package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServerForPerformance() *MCPServer {
	repo := NewMockElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

func TestHandleGetPerformanceDashboard(t *testing.T) {
	tests := []struct {
		name     string
		input    GetPerformanceDashboardInput
		validate func(*testing.T, GetPerformanceDashboardOutput)
	}{
		{
			name: "default period",
			input: GetPerformanceDashboardInput{
				Period: "",
			},
			validate: func(t *testing.T, output GetPerformanceDashboardOutput) {
				assert.Equal(t, "last_24h", output.Period)
			},
		},
		{
			name: "custom period",
			input: GetPerformanceDashboardInput{
				Period: "last_7d",
			},
			validate: func(t *testing.T, output GetPerformanceDashboardOutput) {
				assert.Equal(t, "last_7d", output.Period)
			},
		},
		{
			name: "last hour period",
			input: GetPerformanceDashboardInput{
				Period: "last_1h",
			},
			validate: func(t *testing.T, output GetPerformanceDashboardOutput) {
				assert.Equal(t, "last_1h", output.Period)
				assert.NotNil(t, output.SlowOperations)
				assert.NotNil(t, output.FastOperations)
				assert.NotNil(t, output.ByOperation)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServerForPerformance()
			ctx := context.Background()
			req := &sdk.CallToolRequest{}

			result, output, err := server.handleGetPerformanceDashboard(ctx, req, tt.input)
			require.NoError(t, err)
			assert.Nil(t, result)

			tt.validate(t, output)
		})
	}
}

func TestHandleGetPerformanceDashboard_OutputStructure(t *testing.T) {
	server := setupTestServerForPerformance()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetPerformanceDashboardInput{Period: "last_24h"}

	_, output, err := server.handleGetPerformanceDashboard(ctx, req, input)
	require.NoError(t, err)

	// Verify output structure
	assert.GreaterOrEqual(t, output.TotalOperations, 0)
	assert.GreaterOrEqual(t, output.AvgDuration, float64(0))
	assert.GreaterOrEqual(t, output.P50Duration, float64(0))
	assert.GreaterOrEqual(t, output.P95Duration, float64(0))
	assert.GreaterOrEqual(t, output.P99Duration, float64(0))
	assert.GreaterOrEqual(t, output.MaxDuration, float64(0))
	assert.GreaterOrEqual(t, output.MinDuration, float64(0))
	assert.NotNil(t, output.SlowOperations)
	assert.NotNil(t, output.FastOperations)
	assert.NotNil(t, output.ByOperation)
	assert.Equal(t, "last_24h", output.Period)
}

func TestHandleGetPerformanceDashboard_SlowOperationsFormat(t *testing.T) {
	server := setupTestServerForPerformance()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetPerformanceDashboardInput{Period: "last_24h"}

	_, output, err := server.handleGetPerformanceDashboard(ctx, req, input)
	require.NoError(t, err)

	// Verify slow operations is a slice
	assert.NotNil(t, output.SlowOperations)
}

func TestHandleGetPerformanceDashboard_FastOperationsFormat(t *testing.T) {
	server := setupTestServerForPerformance()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetPerformanceDashboardInput{Period: "last_24h"}

	_, output, err := server.handleGetPerformanceDashboard(ctx, req, input)
	require.NoError(t, err)

	// Verify fast operations is a slice
	assert.NotNil(t, output.FastOperations)
}

func TestHandleGetPerformanceDashboard_ByOperationFormat(t *testing.T) {
	server := setupTestServerForPerformance()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetPerformanceDashboardInput{Period: "last_24h"}

	_, output, err := server.handleGetPerformanceDashboard(ctx, req, input)
	require.NoError(t, err)

	// Verify by-operation stats is a map
	assert.NotNil(t, output.ByOperation)
}

func TestHandleGetPerformanceDashboard_NilResult(t *testing.T) {
	server := setupTestServerForPerformance()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetPerformanceDashboardInput{Period: "last_24h"}

	result, _, err := server.handleGetPerformanceDashboard(ctx, req, input)
	require.NoError(t, err)
	assert.Nil(t, result, "CallToolResult should be nil")
}

func TestHandleGetPerformanceDashboard_EmptyPeriod(t *testing.T) {
	server := setupTestServerForPerformance()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetPerformanceDashboardInput{} // Empty period

	_, output, err := server.handleGetPerformanceDashboard(ctx, req, input)
	require.NoError(t, err)
	assert.Equal(t, "last_24h", output.Period, "should default to last_24h")
}

func TestHandleGetPerformanceDashboard_CustomPeriods(t *testing.T) {
	periods := []string{"last_1h", "last_24h", "last_7d", "last_30d"}

	for _, period := range periods {
		t.Run(period, func(t *testing.T) {
			server := setupTestServerForPerformance()
			ctx := context.Background()
			req := &sdk.CallToolRequest{}
			input := GetPerformanceDashboardInput{Period: period}

			_, output, err := server.handleGetPerformanceDashboard(ctx, req, input)
			require.NoError(t, err)
			assert.Equal(t, period, output.Period)
		})
	}
}

func TestHandleGetPerformanceDashboard_ConsistentFields(t *testing.T) {
	server := setupTestServerForPerformance()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetPerformanceDashboardInput{Period: "last_24h"}

	_, output, err := server.handleGetPerformanceDashboard(ctx, req, input)
	require.NoError(t, err)

	// Verify consistent field types
	assert.IsType(t, 0, output.TotalOperations)
	assert.IsType(t, float64(0), output.AvgDuration)
	assert.IsType(t, float64(0), output.P50Duration)
	assert.IsType(t, float64(0), output.P95Duration)
	assert.IsType(t, float64(0), output.P99Duration)
	assert.IsType(t, float64(0), output.MaxDuration)
	assert.IsType(t, float64(0), output.MinDuration)
	assert.IsType(t, "", output.Period)
}
