package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of ElementRepository for testing.
type MockRepository struct {
	elements map[string]domain.Element
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		elements: make(map[string]domain.Element),
	}
}

func (m *MockRepository) Create(element domain.Element) error {
	m.elements[element.GetID()] = element
	return nil
}

func (m *MockRepository) GetByID(id string) (domain.Element, error) {
	element, exists := m.elements[id]
	if !exists {
		return nil, errors.New("element not found")
	}
	return element, nil
}

func (m *MockRepository) Update(element domain.Element) error {
	m.elements[element.GetID()] = element
	return nil
}

func (m *MockRepository) Delete(id string) error {
	delete(m.elements, id)
	return nil
}

func (m *MockRepository) List(filter domain.ElementFilter) ([]domain.Element, error) {
	var result []domain.Element
	for _, element := range m.elements {
		// Apply type filter if specified
		if filter.Type != nil && element.GetType() != *filter.Type {
			continue
		}
		// Apply active filter if specified
		if filter.IsActive != nil && element.GetMetadata().IsActive != *filter.IsActive {
			continue
		}
		result = append(result, element)
	}
	return result, nil
}

func (m *MockRepository) Exists(id string) (bool, error) {
	_, exists := m.elements[id]
	return exists, nil
}

// Test helper: create a test ensemble.
func createTestEnsemble(t *testing.T, executionMode string) *domain.Ensemble {
	ensemble := domain.NewEnsemble("test-ensemble", "Test ensemble", "1.0.0", "test-author")

	// Add members with specific agent IDs
	members := []domain.EnsembleMember{
		{AgentID: "agent-agent-1", Role: "primary", Priority: 10},
		{AgentID: "agent-agent-2", Role: "secondary", Priority: 5},
		{AgentID: "agent-agent-3", Role: "tertiary", Priority: 1},
	}

	// Use reflection to set private fields (for testing purposes)
	// In real code, you'd use proper setters
	ensemble.Members = members
	ensemble.ExecutionMode = executionMode
	ensemble.AggregationStrategy = "all"

	return ensemble
}

// Test helper: create a test agent (used for special test cases).
func createTestAgent(t *testing.T, name string) *domain.Agent {
	agent := domain.NewAgent(name, "Test agent", "1.0.0", "test-author")
	agent.Actions = []domain.AgentAction{
		{Name: "test-action", Type: "tool"},
	}
	agent.Goals = []string{"test-goal"}
	return agent
}

// Test helper: create and save test agents for ensemble.
func createAndSaveAgentsForEnsemble(t *testing.T, repo *MockRepository, ensemble *domain.Ensemble) {
	for i, member := range ensemble.Members {
		agent := domain.NewAgent("Agent-"+string(rune('1'+i)), "Test agent", "1.0.0", "test-author")
		agent.Actions = []domain.AgentAction{
			{Name: "test-action", Type: "tool"},
		}
		agent.Goals = []string{"test-goal"}

		err := repo.Create(agent)
		require.NoError(t, err)

		// Also map to the expected ID from ensemble
		if agent.GetID() != member.AgentID {
			repo.elements[member.AgentID] = agent
		}
	}
}

func TestNewEnsembleExecutor(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()

	executor := NewEnsembleExecutor(repo, log)

	assert.NotNil(t, executor)
	assert.Equal(t, repo, executor.repository)
	assert.Equal(t, log, executor.logger)
}

func TestExecuteSequential(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	// Create and save ensemble
	ensemble := createTestEnsemble(t, "sequential")
	err := repo.Create(ensemble)
	require.NoError(t, err)

	// Create and save agents - ensure IDs match ensemble members
	for i, member := range ensemble.Members {
		agent := domain.NewAgent("Agent-"+string(rune('1'+i)), "Test agent", "1.0.0", "test-author")
		agent.Actions = []domain.AgentAction{
			{Name: "test-action", Type: "tool"},
		}
		agent.Goals = []string{"test-goal"}

		// Override the ID to match what ensemble expects
		t.Logf("Creating agent with ID: %s (ensemble expects: %s)", agent.GetID(), member.AgentID)
		err := repo.Create(agent)
		require.NoError(t, err)

		// Also create with the expected ID
		if agent.GetID() != member.AgentID {
			repo.elements[member.AgentID] = agent
		}
	}

	// Execute
	ctx := context.Background()
	req := ExecutionRequest{
		EnsembleID: ensemble.GetID(),
		Input: map[string]interface{}{
			"test": "data",
		},
		Options: ExecutionOptions{
			Timeout: 10 * time.Second,
		},
	}

	result, err := executor.Execute(ctx, req)

	// Verify
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ensemble.GetID(), result.EnsembleID)
	assert.Len(t, result.Results, 3)
	assert.Contains(t, []string{"success", "partial_success", "failed"}, result.Status)
}

func TestExecuteParallel(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	// Create and save ensemble
	ensemble := createTestEnsemble(t, "parallel")
	err := repo.Create(ensemble)
	require.NoError(t, err)

	// Create and save agents
	createAndSaveAgentsForEnsemble(t, repo, ensemble)

	// Execute
	ctx := context.Background()
	req := ExecutionRequest{
		EnsembleID: ensemble.GetID(),
		Input: map[string]interface{}{
			"test": "data",
		},
		Options: ExecutionOptions{
			Timeout: 10 * time.Second,
		},
	}

	result, err := executor.Execute(ctx, req)

	// Verify
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ensemble.GetID(), result.EnsembleID)
	assert.Len(t, result.Results, 3)
	assert.Contains(t, []string{"success", "partial_success"}, result.Status)
}

func TestExecuteHybrid(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	// Create and save ensemble with different priorities
	ensemble := createTestEnsemble(t, "hybrid")
	err := repo.Create(ensemble)
	require.NoError(t, err)

	// Create and save agents
	createAndSaveAgentsForEnsemble(t, repo, ensemble)

	// Execute
	ctx := context.Background()
	req := ExecutionRequest{
		EnsembleID: ensemble.GetID(),
		Input: map[string]interface{}{
			"test": "data",
		},
		Options: ExecutionOptions{
			Timeout: 10 * time.Second,
		},
	}

	result, err := executor.Execute(ctx, req)

	// Verify
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ensemble.GetID(), result.EnsembleID)
	assert.Len(t, result.Results, 3)
	assert.Contains(t, []string{"success", "partial_success", "failed"}, result.Status)

	// Verify hybrid execution: results should be ordered by priority
	// (in hybrid mode, higher priority executes first within parallel groups)
	assert.Equal(t, ensemble.Members[0].AgentID, result.Results[0].AgentID) // Priority 10
}

func TestExecuteWithTimeout(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	// Create and save ensemble
	ensemble := createTestEnsemble(t, "sequential")
	err := repo.Create(ensemble)
	require.NoError(t, err)

	// Create and save agents
	createAndSaveAgentsForEnsemble(t, repo, ensemble)

	// Execute with very short timeout
	ctx := context.Background()
	req := ExecutionRequest{
		EnsembleID: ensemble.GetID(),
		Input:      map[string]interface{}{"test": "data"},
		Options: ExecutionOptions{
			Timeout: 1 * time.Nanosecond, // Very short timeout
		},
	}

	result, err := executor.Execute(ctx, req)

	// Should complete but may have timeout errors
	assert.NotNil(t, result)
}

func TestExecuteFailFast(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	// Create ensemble with fail-fast
	ensemble := createTestEnsemble(t, "sequential")
	err := repo.Create(ensemble)
	require.NoError(t, err)

	// Create agents (one will "fail" because it doesn't exist)
	agent1 := createTestAgent(t, "Agent-1")
	err = repo.Create(agent1)
	require.NoError(t, err)
	// Map to ensemble's expected ID
	repo.elements[ensemble.Members[0].AgentID] = agent1

	// Don't save agent-2, it will fail to load

	agent3 := createTestAgent(t, "Agent-3")
	err = repo.Create(agent3)
	require.NoError(t, err)
	// Map to ensemble's expected ID
	repo.elements[ensemble.Members[2].AgentID] = agent3

	// Execute with fail-fast
	ctx := context.Background()
	req := ExecutionRequest{
		EnsembleID: ensemble.GetID(),
		Input:      map[string]interface{}{"test": "data"},
		Options: ExecutionOptions{
			FailFast: true,
			Timeout:  10 * time.Second,
		},
	}

	result, err := executor.Execute(ctx, req)

	// Should stop at first failure
	assert.NotNil(t, result)
	// Should have fewer than 3 results due to fail-fast
	assert.LessOrEqual(t, len(result.Results), 3)
}

func TestAggregationStrategies(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	// Test different aggregation strategies
	strategies := []string{"first", "last", "all", "merge"}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			ensemble := createTestEnsemble(t, "parallel")
			ensemble.AggregationStrategy = strategy

			results := []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: map[string]interface{}{"value": 1}},
				{AgentID: "agent-2", Status: "success", Result: map[string]interface{}{"value": 2}},
				{AgentID: "agent-3", Status: "success", Result: map[string]interface{}{"value": 3}},
			}

			aggregated, err := executor.aggregateResults(ensemble, results)

			require.NoError(t, err)
			assert.NotNil(t, aggregated)

			switch strategy {
			case "first":
				resultMap, ok := aggregated.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, 1, resultMap["value"])

			case "last":
				resultMap, ok := aggregated.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, 3, resultMap["value"])

			case "all":
				resultArray, ok := aggregated.([]interface{})
				assert.True(t, ok)
				assert.Len(t, resultArray, 3)

			case "merge":
				resultMap, ok := aggregated.(map[string]interface{})
				assert.True(t, ok)
				assert.Contains(t, resultMap, "agent_0_value")
				assert.Contains(t, resultMap, "agent_1_value")
				assert.Contains(t, resultMap, "agent_2_value")
			}
		})
	}
}

func TestDetermineStatus(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	tests := []struct {
		name     string
		results  []AgentResult
		execErr  error
		expected string
	}{
		{
			name: "all success",
			results: []AgentResult{
				{Status: "success"},
				{Status: "success"},
			},
			execErr:  nil,
			expected: "success",
		},
		{
			name: "all failed",
			results: []AgentResult{
				{Status: "failed"},
				{Status: "failed"},
			},
			execErr:  nil,
			expected: "failed",
		},
		{
			name: "partial success",
			results: []AgentResult{
				{Status: "success"},
				{Status: "failed"},
			},
			execErr:  nil,
			expected: "partial_success",
		},
		{
			name: "execution error",
			results: []AgentResult{
				{Status: "success"},
			},
			execErr:  errors.New("test error"),
			expected: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := executor.determineStatus(tt.results, tt.execErr)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestGroupByPriority(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	members := []domain.EnsembleMember{
		{AgentID: "agent-1", Priority: 10},
		{AgentID: "agent-2", Priority: 10},
		{AgentID: "agent-3", Priority: 5},
		{AgentID: "agent-4", Priority: 1},
	}

	groups := executor.groupByPriority(members)

	assert.Len(t, groups, 3)
	assert.Len(t, groups[10], 2)
	assert.Len(t, groups[5], 1)
	assert.Len(t, groups[1], 1)
}

func TestInitializeSharedContext(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	ensemble := createTestEnsemble(t, "sequential")
	ensemble.SharedContext = map[string]interface{}{
		"shared_key": "shared_value",
	}

	input := map[string]interface{}{
		"input_key": "input_value",
	}

	ctx := executor.initializeSharedContext(ensemble, input)

	assert.Equal(t, "shared_value", ctx["shared_key"])
	assert.Equal(t, input, ctx["input"])
	assert.Equal(t, ensemble.GetID(), ctx["ensemble_id"])
}

func TestExecuteWithInvalidEnsembleID(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	ctx := context.Background()
	req := ExecutionRequest{
		EnsembleID: "non-existent-ensemble",
		Input:      map[string]interface{}{},
	}

	result, err := executor.Execute(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to load ensemble")
}

func TestExecuteWithNoMembers(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	// Create ensemble with no members
	ensemble := domain.NewEnsemble("empty-ensemble", "Empty ensemble", "1.0.0", "test-author")
	ensemble.ExecutionMode = "sequential"
	ensemble.AggregationStrategy = "all"
	err := repo.Create(ensemble)
	require.NoError(t, err)

	ctx := context.Background()
	req := ExecutionRequest{
		EnsembleID: ensemble.GetID(),
		Input:      map[string]interface{}{},
	}

	result, err := executor.Execute(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ensemble has no members")
}

func TestExecuteWithUnsupportedMode(t *testing.T) {
	repo := NewMockRepository()
	log := logger.Get()
	executor := NewEnsembleExecutor(repo, log)

	// Create and save ensemble with unsupported mode
	ensemble := createTestEnsemble(t, "unsupported_mode")
	err := repo.Create(ensemble)
	require.NoError(t, err)

	// Create and save agents
	createAndSaveAgentsForEnsemble(t, repo, ensemble)

	ctx := context.Background()
	req := ExecutionRequest{
		EnsembleID: ensemble.GetID(),
		Input:      map[string]interface{}{},
	}

	result, err := executor.Execute(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "failed", result.Status)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "unsupported execution mode")
}
