package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ExecuteEnsembleInput defines input for execute_ensemble tool
type ExecuteEnsembleInput struct {
	EnsembleID       string                 `json:"ensemble_id" jsonschema:"required,description=ID of the ensemble to execute"`
	Input            map[string]interface{} `json:"input" jsonschema:"description=Input data for ensemble execution"`
	TimeoutSeconds   int                    `json:"timeout_seconds,omitempty" jsonschema:"description=Execution timeout in seconds (default: 300)"`
	MaxRetries       int                    `json:"max_retries,omitempty" jsonschema:"description=Max retries per agent (default: 1)"`
	FailFast         bool                   `json:"fail_fast,omitempty" jsonschema:"description=Stop on first agent failure (default: false)"`
	EnableMonitoring bool                   `json:"enable_monitoring,omitempty" jsonschema:"description=Enable execution monitoring (default: true)"`
}

// ExecuteEnsembleOutput defines output for execute_ensemble tool
type ExecuteEnsembleOutput struct {
	EnsembleID       string                    `json:"ensemble_id"`
	Status           string                    `json:"status"` // success, partial_success, failed
	Results          []application.AgentResult `json:"results"`
	AggregatedResult interface{}               `json:"aggregated_result,omitempty"`
	ExecutionTime    string                    `json:"execution_time"` // formatted duration
	StartedAt        string                    `json:"started_at"`
	FinishedAt       string                    `json:"finished_at"`
	Errors           []string                  `json:"errors,omitempty"`
	Summary          string                    `json:"summary"`
}

// handleExecuteEnsemble handles execute_ensemble tool calls
func (s *MCPServer) handleExecuteEnsemble(ctx context.Context, req *sdk.CallToolRequest, input ExecuteEnsembleInput) (*sdk.CallToolResult, ExecuteEnsembleOutput, error) {
	// Validate required inputs
	if input.EnsembleID == "" {
		return nil, ExecuteEnsembleOutput{}, fmt.Errorf("ensemble_id is required")
	}

	// Set defaults
	if input.TimeoutSeconds <= 0 {
		input.TimeoutSeconds = 300 // 5 minutes default
	}
	if input.MaxRetries <= 0 {
		input.MaxRetries = 1
	}
	if input.Input == nil {
		input.Input = make(map[string]interface{})
	}

	// Create ensemble executor
	executor := application.NewEnsembleExecutor(s.repo, logger.Get())

	// Prepare execution request
	execReq := application.ExecutionRequest{
		EnsembleID: input.EnsembleID,
		Input:      input.Input,
		Options: application.ExecutionOptions{
			Timeout:          time.Duration(input.TimeoutSeconds) * time.Second,
			MaxRetries:       input.MaxRetries,
			FailFast:         input.FailFast,
			EnableMonitoring: input.EnableMonitoring,
		},
	}

	// Execute ensemble
	result, err := executor.Execute(ctx, execReq)
	if err != nil {
		return nil, ExecuteEnsembleOutput{}, fmt.Errorf("ensemble execution failed: %w", err)
	}

	// Build output
	output := ExecuteEnsembleOutput{
		EnsembleID:       result.EnsembleID,
		Status:           result.Status,
		Results:          result.Results,
		AggregatedResult: result.AggregatedResult,
		ExecutionTime:    result.ExecutionTime.String(),
		StartedAt:        result.StartedAt.Format(time.RFC3339),
		FinishedAt:       result.FinishedAt.Format(time.RFC3339),
		Errors:           result.Errors,
	}

	// Generate summary
	output.Summary = s.generateEnsembleSummary(result)

	// Return nil for CallToolResult (SDK will create it from output)
	return nil, output, nil
}

// GetEnsembleStatusInput defines input for get_ensemble_status tool
type GetEnsembleStatusInput struct {
	EnsembleID string `json:"ensemble_id" jsonschema:"required,description=ID of the ensemble to check"`
}

// GetEnsembleStatusOutput defines output for get_ensemble_status tool
type GetEnsembleStatusOutput struct {
	EnsembleID          string   `json:"ensemble_id"`
	Name                string   `json:"name"`
	ExecutionMode       string   `json:"execution_mode"`
	MemberCount         int      `json:"member_count"`
	Members             []string `json:"members"`
	AggregationStrategy string   `json:"aggregation_strategy"`
	FallbackChain       []string `json:"fallback_chain,omitempty"`
	IsActive            bool     `json:"is_active"`
}

// handleGetEnsembleStatus handles get_ensemble_status tool calls
func (s *MCPServer) handleGetEnsembleStatus(ctx context.Context, req *sdk.CallToolRequest, input GetEnsembleStatusInput) (*sdk.CallToolResult, GetEnsembleStatusOutput, error) {
	// Validate required inputs
	if input.EnsembleID == "" {
		return nil, GetEnsembleStatusOutput{}, fmt.Errorf("ensemble_id is required")
	}

	// Load ensemble
	element, err := s.repo.GetByID(input.EnsembleID)
	if err != nil {
		return nil, GetEnsembleStatusOutput{}, fmt.Errorf("ensemble not found: %w", err)
	}

	ensemble, ok := element.(*domain.Ensemble)
	if !ok {
		return nil, GetEnsembleStatusOutput{}, fmt.Errorf("element %s is not an ensemble", input.EnsembleID)
	}

	// Extract member IDs
	memberIDs := make([]string, len(ensemble.Members))
	for i, member := range ensemble.Members {
		memberIDs[i] = member.AgentID
	}

	// Build output
	output := GetEnsembleStatusOutput{
		EnsembleID:          ensemble.GetID(),
		Name:                ensemble.GetMetadata().Name,
		ExecutionMode:       ensemble.ExecutionMode,
		MemberCount:         len(ensemble.Members),
		Members:             memberIDs,
		AggregationStrategy: ensemble.AggregationStrategy,
		FallbackChain:       ensemble.FallbackChain,
		IsActive:            ensemble.GetMetadata().IsActive,
	}

	// Return nil for CallToolResult (SDK will create it from output)
	return nil, output, nil
}

// Helper functions

func (s *MCPServer) generateEnsembleSummary(result *application.ExecutionResult) string {
	if result.Status == "success" {
		return "All agents executed successfully. Results aggregated."
	}

	if result.Status == "partial_success" {
		successCount := countSuccessfulAgents(result.Results)
		return fmt.Sprintf("%d/%d agents succeeded. Partial results available.", successCount, len(result.Results))
	}

	return "Ensemble execution failed. Check errors for details."
}

func countSuccessfulAgents(results []application.AgentResult) int {
	count := 0
	for _, result := range results {
		if result.Status == "success" {
			count++
		}
	}
	return count
}

func formatMemberList(members []domain.EnsembleMember) string {
	if len(members) == 0 {
		return "  (no members)"
	}

	result := ""
	for i, member := range members {
		result += fmt.Sprintf("  %d. **%s** (role: %s, priority: %d)\n",
			i+1, member.AgentID, member.Role, member.Priority)
	}
	return result
}
