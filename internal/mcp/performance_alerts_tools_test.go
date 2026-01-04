package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetActiveAlerts(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetActiveAlertsInput{}

	_, output, err := server.handleGetActiveAlerts(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.GreaterOrEqual(t, len(output.Alerts), 0)
}

func TestGetAlertHistory(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetAlertHistoryInput{Limit: 50}

	_, output, err := server.handleGetAlertHistory(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
}

func TestGetAlertRules(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := GetAlertRulesInput{}

	_, output, err := server.handleGetAlertRules(ctx, nil, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.GreaterOrEqual(t, len(output.Rules), 0)
}
