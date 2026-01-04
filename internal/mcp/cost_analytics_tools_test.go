package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCostAnalytics(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetCostAnalyticsInput{
		Period: "last_24h",
	}

	_, output, err := server.handleGetCostAnalytics(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.GreaterOrEqual(t, output.TotalOperations, 0)
	assert.GreaterOrEqual(t, output.TotalTokens, 0)
}

func TestCostAnalyticsPeriods(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	periods := []string{"last_1h", "last_24h", "last_7d", "last_30d"}

	for _, period := range periods {
		t.Run(period, func(t *testing.T) {
			input := GetCostAnalyticsInput{
				Period: period,
			}

			_, output, err := server.handleGetCostAnalytics(ctx, nil, input)

			require.NoError(t, err)
			assert.NotNil(t, output)
			assert.GreaterOrEqual(t, output.TotalOperations, 0)
		})
	}
}

func TestCostTrendsAnalysis(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetCostAnalyticsInput{
		Period: "last_7d",
	}

	_, output, err := server.handleGetCostAnalytics(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.NotNil(t, output.CostTrends)
}

func TestExpensiveToolsIdentification(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetCostAnalyticsInput{
		Period: "last_24h",
	}

	_, output, err := server.handleGetCostAnalytics(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.LessOrEqual(t, len(output.TopExpensiveTools), 10)
}

func TestCostProjections(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetCostAnalyticsInput{
		Period: "last_30d",
	}

	_, output, err := server.handleGetCostAnalytics(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.NotNil(t, output.CostProjections)
}

func TestTokenSavingsCalculation(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetCostAnalyticsInput{
		Period: "last_24h",
	}

	_, output, err := server.handleGetCostAnalytics(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.GreaterOrEqual(t, output.TokenSavingsPercent, float64(0))
	assert.LessOrEqual(t, output.TokenSavingsPercent, float64(100))
}

func TestCostAnomalyDetection(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetCostAnalyticsInput{
		Period: "last_7d",
	}

	_, output, err := server.handleGetCostAnalytics(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	// Anomalies may be empty if none detected
	assert.NotNil(t, output.Anomalies)
}
