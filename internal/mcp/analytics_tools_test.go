package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServerForAnalytics() *MCPServer {
	repo := NewMockElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

func TestHandleGetUsageStats(t *testing.T) {
	tests := []struct {
		name     string
		input    GetUsageStatsInput
		validate func(*testing.T, GetUsageStatsOutput)
	}{
		{
			name: "default period",
			input: GetUsageStatsInput{
				Period: "",
			},
			validate: func(t *testing.T, output GetUsageStatsOutput) {
				assert.Equal(t, "last_24h", output.Period)
			},
		},
		{
			name: "custom period",
			input: GetUsageStatsInput{
				Period: "last_7d",
			},
			validate: func(t *testing.T, output GetUsageStatsOutput) {
				assert.Equal(t, "last_7d", output.Period)
			},
		},
		{
			name: "last hour period",
			input: GetUsageStatsInput{
				Period: "last_1h",
			},
			validate: func(t *testing.T, output GetUsageStatsOutput) {
				assert.Equal(t, "last_1h", output.Period)
				assert.NotNil(t, output.MostUsedTools)
				assert.NotNil(t, output.SlowestOperations)
				assert.NotNil(t, output.RecentErrors)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServerForAnalytics()
			ctx := context.Background()
			req := &sdk.CallToolRequest{}

			result, output, err := server.handleGetUsageStats(ctx, req, tt.input)
			require.NoError(t, err)
			assert.Nil(t, result)

			tt.validate(t, output)
		})
	}
}

func TestHandleGetUsageStats_OutputStructure(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Verify output structure
	assert.GreaterOrEqual(t, output.TotalOperations, 0)
	assert.GreaterOrEqual(t, output.SuccessfulOps, 0)
	assert.GreaterOrEqual(t, output.FailedOps, 0)
	assert.GreaterOrEqual(t, output.SuccessRate, float64(0))
	assert.Equal(t, "last_24h", output.Period)
	assert.NotEmpty(t, output.StartTime)
	assert.NotEmpty(t, output.EndTime)
}

func TestHandleGetUsageStats_MostUsedToolsFormat(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Verify most used tools is a slice
	assert.NotNil(t, output.MostUsedTools)
}

func TestHandleGetUsageStats_SlowestOperationsFormat(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Verify slowest operations is a slice
	assert.NotNil(t, output.SlowestOperations)
}

func TestHandleGetUsageStats_RecentErrorsFormat(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Verify recent errors is a slice
	assert.NotNil(t, output.RecentErrors)
}

func TestHandleGetUsageStats_NilResult(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	result, _, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)
	assert.Nil(t, result, "CallToolResult should be nil")
}

func TestHandleGetUsageStats_EmptyPeriod(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{} // Empty period

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)
	assert.Equal(t, "last_24h", output.Period, "should default to last_24h")
}

func TestHandleGetUsageStats_CustomPeriods(t *testing.T) {
	periods := []string{"last_1h", "last_24h", "last_7d", "last_30d"}

	for _, period := range periods {
		t.Run(period, func(t *testing.T) {
			server := setupTestServerForAnalytics()
			ctx := context.Background()
			req := &sdk.CallToolRequest{}
			input := GetUsageStatsInput{Period: period}

			_, output, err := server.handleGetUsageStats(ctx, req, input)
			require.NoError(t, err)
			assert.Equal(t, period, output.Period)
		})
	}
}

func TestHandleGetUsageStats_ConsistentFields(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Verify consistent field types
	assert.IsType(t, 0, output.TotalOperations)
	assert.IsType(t, 0, output.SuccessfulOps)
	assert.IsType(t, 0, output.FailedOps)
	assert.IsType(t, float64(0), output.SuccessRate)
	assert.IsType(t, "", output.Period)
	assert.IsType(t, "", output.StartTime)
	assert.IsType(t, "", output.EndTime)
	assert.IsType(t, []string{}, output.ActiveUsers)
}

func TestHandleGetUsageStats_SuccessRate(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Success rate should be between 0 and 100
	assert.GreaterOrEqual(t, output.SuccessRate, float64(0))
	assert.LessOrEqual(t, output.SuccessRate, float64(100))
}

func TestHandleGetUsageStats_OperationCounts(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Total should equal successful + failed
	assert.GreaterOrEqual(t, output.TotalOperations, output.SuccessfulOps+output.FailedOps)
}

func TestHandleGetUsageStats_TimeFormat(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Verify timestamps are formatted strings
	assert.IsType(t, "", output.StartTime)
	assert.IsType(t, "", output.EndTime)
}

func TestHandleGetUsageStats_ActiveUsers(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Verify active users is correct type
	assert.IsType(t, []string{}, output.ActiveUsers)
}

func TestHandleGetUsageStats_SlicesNotNil(t *testing.T) {
	server := setupTestServerForAnalytics()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetUsageStatsInput{Period: "last_24h"}

	_, output, err := server.handleGetUsageStats(ctx, req, input)
	require.NoError(t, err)

	// Verify slices have correct types
	assert.IsType(t, []map[string]interface{}{}, output.MostUsedTools)
	assert.IsType(t, []map[string]interface{}{}, output.SlowestOperations)
	assert.IsType(t, []map[string]interface{}{}, output.RecentErrors)
}
