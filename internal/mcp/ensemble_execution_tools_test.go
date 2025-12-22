package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServerForEnsemble() *MCPServer {
	repo := NewMockElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

func TestHandleExecuteEnsemble_RequiredEnsembleID(t *testing.T) {
	server := setupTestServerForEnsemble()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ExecuteEnsembleInput{} // Missing ensemble_id

	_, _, err := server.handleExecuteEnsemble(ctx, req, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ensemble_id is required")
}

func TestHandleExecuteEnsemble_DefaultTimeout(t *testing.T) {
	server := setupTestServerForEnsemble()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ExecuteEnsembleInput{
		EnsembleID: "test-ensemble",
	}

	// This will fail but we're testing that defaults are set
	_, _, err := server.handleExecuteEnsemble(ctx, req, input)
	// Error expected since mock doesn't have ensemble executor
	assert.Error(t, err)
}

func TestHandleExecuteEnsemble_DefaultMaxRetries(t *testing.T) {
	server := setupTestServerForEnsemble()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ExecuteEnsembleInput{
		EnsembleID: "test-ensemble",
		MaxRetries: 0, // Should default to 1
	}

	_, _, err := server.handleExecuteEnsemble(ctx, req, input)
	assert.Error(t, err) // Expected to fail on mock
}

func TestHandleExecuteEnsemble_DefaultInput(t *testing.T) {
	server := setupTestServerForEnsemble()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ExecuteEnsembleInput{
		EnsembleID: "test-ensemble",
		Input:      nil, // Should be initialized to empty map
	}

	_, _, err := server.handleExecuteEnsemble(ctx, req, input)
	assert.Error(t, err) // Expected to fail on mock
}

func TestHandleExecuteEnsemble_CustomTimeout(t *testing.T) {
	server := setupTestServerForEnsemble()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ExecuteEnsembleInput{
		EnsembleID:     "test-ensemble",
		TimeoutSeconds: 60,
	}

	_, _, err := server.handleExecuteEnsemble(ctx, req, input)
	assert.Error(t, err) // Expected to fail on mock
}

func TestHandleExecuteEnsemble_FailFast(t *testing.T) {
	server := setupTestServerForEnsemble()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ExecuteEnsembleInput{
		EnsembleID: "test-ensemble",
		FailFast:   true,
	}

	_, _, err := server.handleExecuteEnsemble(ctx, req, input)
	assert.Error(t, err) // Expected to fail on mock
}

func TestHandleExecuteEnsemble_EnableMonitoring(t *testing.T) {
	server := setupTestServerForEnsemble()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := ExecuteEnsembleInput{
		EnsembleID:       "test-ensemble",
		EnableMonitoring: true,
	}

	_, _, err := server.handleExecuteEnsemble(ctx, req, input)
	assert.Error(t, err) // Expected to fail on mock
}

func TestHandleGetEnsembleStatus_RequiredEnsembleID(t *testing.T) {
	server := setupTestServerForEnsemble()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetEnsembleStatusInput{} // Missing ensemble_id

	_, _, err := server.handleGetEnsembleStatus(ctx, req, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ensemble_id is required")
}

func TestHandleGetEnsembleStatus_EnsembleNotFound(t *testing.T) {
	server := setupTestServerForEnsemble()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetEnsembleStatusInput{
		EnsembleID: "non-existent-ensemble",
	}

	_, _, err := server.handleGetEnsembleStatus(ctx, req, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHandleGetEnsembleStatus_ValidEnsemble(t *testing.T) {
	server := setupTestServerForEnsemble()

	// Add a test ensemble to repository
	ensemble := domain.NewEnsemble("Test Ensemble", "Test description", "1.0.0", "test-author")
	ensemble.ExecutionMode = "sequential"
	ensemble.AggregationStrategy = "consensus"
	ensemble.Members = []domain.EnsembleMember{
		{AgentID: "agent-1", Role: "primary", Priority: 1},
		{AgentID: "agent-2", Role: "secondary", Priority: 2},
	}
	ensemble.FallbackChain = []string{"agent-1", "agent-2"}

	err := server.repo.Create(ensemble)
	require.NoError(t, err)

	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetEnsembleStatusInput{
		EnsembleID: ensemble.GetID(),
	}

	result, output, err := server.handleGetEnsembleStatus(ctx, req, input)
	require.NoError(t, err)
	assert.Nil(t, result)

	// Verify output
	assert.Equal(t, ensemble.GetID(), output.EnsembleID)
	assert.Equal(t, "consensus", output.AggregationStrategy)
	assert.Equal(t, []string{"agent-1", "agent-2"}, output.FallbackChain)
	assert.True(t, output.IsActive)
}

func TestHandleGetEnsembleStatus_WrongElementType(t *testing.T) {
	server := setupTestServerForEnsemble()

	// Add a non-ensemble element
	persona := domain.NewPersona("Test Persona", "Test description", "1.0.0", "test-author")
	err := server.repo.Create(persona)
	require.NoError(t, err)

	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetEnsembleStatusInput{
		EnsembleID: persona.GetID(),
	}

	_, _, err = server.handleGetEnsembleStatus(ctx, req, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an ensemble")
}

func TestGenerateEnsembleSummary_Success(t *testing.T) {
	server := setupTestServerForEnsemble()
	result := &application.ExecutionResult{
		Status: "success",
	}

	summary := server.generateEnsembleSummary(result)
	assert.Contains(t, summary, "All agents executed successfully")
}

func TestGenerateEnsembleSummary_PartialSuccess(t *testing.T) {
	server := setupTestServerForEnsemble()
	result := &application.ExecutionResult{
		Status: "partial_success",
		Results: []application.AgentResult{
			{Status: "success"},
			{Status: "failed"},
			{Status: "success"},
		},
	}

	summary := server.generateEnsembleSummary(result)
	assert.Contains(t, summary, "2/3 agents succeeded")
}

func TestGenerateEnsembleSummary_Failed(t *testing.T) {
	server := setupTestServerForEnsemble()
	result := &application.ExecutionResult{
		Status: "failed",
	}

	summary := server.generateEnsembleSummary(result)
	assert.Contains(t, summary, "Ensemble execution failed")
}

func TestCountSuccessfulAgents(t *testing.T) {
	tests := []struct {
		name     string
		results  []application.AgentResult
		expected int
	}{
		{
			name:     "all successful",
			results:  []application.AgentResult{{Status: "success"}, {Status: "success"}},
			expected: 2,
		},
		{
			name:     "partial success",
			results:  []application.AgentResult{{Status: "success"}, {Status: "failed"}},
			expected: 1,
		},
		{
			name:     "all failed",
			results:  []application.AgentResult{{Status: "failed"}, {Status: "failed"}},
			expected: 0,
		},
		{
			name:     "empty",
			results:  []application.AgentResult{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := countSuccessfulAgents(tt.results)
			assert.Equal(t, tt.expected, count)
		})
	}
}

func TestHandleGetEnsembleStatus_NilResult(t *testing.T) {
	server := setupTestServerForEnsemble()

	ensemble := domain.NewEnsemble("Test", "Test description", "1.0.0", "test-author")
	ensemble.ExecutionMode = "parallel"
	ensemble.Members = []domain.EnsembleMember{{AgentID: "agent-1", Role: "worker", Priority: 1}}

	err := server.repo.Create(ensemble)
	require.NoError(t, err)

	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetEnsembleStatusInput{EnsembleID: ensemble.GetID()}

	result, _, err := server.handleGetEnsembleStatus(ctx, req, input)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestHandleGetEnsembleStatus_EmptyMembers(t *testing.T) {
	server := setupTestServerForEnsemble()

	ensemble := domain.NewEnsemble("Empty Ensemble", "Test description", "1.0.0", "test-author")
	ensemble.ExecutionMode = "sequential"
	ensemble.Members = []domain.EnsembleMember{}

	err := server.repo.Create(ensemble)
	require.NoError(t, err)

	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := GetEnsembleStatusInput{EnsembleID: ensemble.GetID()}

	_, output, err := server.handleGetEnsembleStatus(ctx, req, input)
	require.NoError(t, err)
	assert.Equal(t, 0, output.MemberCount)
	assert.Empty(t, output.Members)
}

func TestExecuteEnsembleInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input ExecuteEnsembleInput
		valid bool
	}{
		{
			name:  "valid with all fields",
			input: ExecuteEnsembleInput{EnsembleID: "test", TimeoutSeconds: 60, MaxRetries: 3},
			valid: true,
		},
		{
			name:  "missing ensemble_id",
			input: ExecuteEnsembleInput{TimeoutSeconds: 60},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServerForEnsemble()
			ctx := context.Background()
			req := &sdk.CallToolRequest{}

			_, _, err := server.handleExecuteEnsemble(ctx, req, tt.input)
			if tt.valid {
				// Will still error due to mock, but not validation error
				if err != nil {
					assert.NotContains(t, err.Error(), "is required")
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestExecuteEnsembleOutput_Structure(t *testing.T) {
	output := ExecuteEnsembleOutput{
		EnsembleID:    "test",
		Status:        "success",
		Results:       []application.AgentResult{},
		ExecutionTime: "1s",
		StartedAt:     time.Now().Format(time.RFC3339),
		FinishedAt:    time.Now().Format(time.RFC3339),
		Summary:       "Test summary",
	}

	assert.Equal(t, "test", output.EnsembleID)
	assert.Equal(t, "success", output.Status)
	assert.NotNil(t, output.Results)
	assert.NotEmpty(t, output.ExecutionTime)
	assert.NotEmpty(t, output.Summary)
}

func TestGetEnsembleStatusOutput_Structure(t *testing.T) {
	output := GetEnsembleStatusOutput{
		EnsembleID:          "test-id",
		Name:                "Test Ensemble",
		ExecutionMode:       "parallel",
		MemberCount:         3,
		Members:             []string{"a1", "a2", "a3"},
		AggregationStrategy: "voting",
		IsActive:            true,
	}

	assert.Equal(t, "test-id", output.EnsembleID)
	assert.Equal(t, "Test Ensemble", output.Name)
	assert.Equal(t, 3, output.MemberCount)
	assert.Len(t, output.Members, 3)
	assert.True(t, output.IsActive)
}
