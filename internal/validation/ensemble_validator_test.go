package validation

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsembleValidator_SupportedType(t *testing.T) {
	validator := NewEnsembleValidator()
	assert.Equal(t, domain.EnsembleElement, validator.SupportedType())
}

func TestEnsembleValidator_Validate_Success(t *testing.T) {
	validator := NewEnsembleValidator()

	ensemble := domain.NewEnsemble("Test Ensemble", "A comprehensive test ensemble", "1.0.0", "tester")
	ensemble.Members = []domain.EnsembleMember{
		{AgentID: "agent-1", Role: "leader", Priority: 1},
		{AgentID: "agent-2", Role: "worker", Priority: 2},
	}
	ensemble.ExecutionMode = "sequential"
	ensemble.AggregationStrategy = "first"

	result, err := validator.Validate(ensemble, BasicLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(domain.EnsembleElement), result.ElementType)
	assert.Equal(t, ensemble.GetID(), result.ElementID)
}

func TestEnsembleValidator_Validate_ComprehensiveLevel(t *testing.T) {
	validator := NewEnsembleValidator()

	ensemble := domain.NewEnsemble("Test Ensemble", "A test ensemble", "1.0.0", "tester")
	ensemble.Members = []domain.EnsembleMember{
		{AgentID: "agent-1", Role: "leader", Priority: 1},
	}
	ensemble.ExecutionMode = "sequential"
	ensemble.AggregationStrategy = "first"

	result, err := validator.Validate(ensemble, ComprehensiveLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestEnsembleValidator_Validate_StrictLevel(t *testing.T) {
	validator := NewEnsembleValidator()

	ensemble := domain.NewEnsemble("Test Ensemble", "A comprehensive test ensemble", "1.0.0", "tester")
	ensemble.Members = []domain.EnsembleMember{
		{AgentID: "agent-1", Role: "coordinator", Priority: 1},
		{AgentID: "agent-2", Role: "executor", Priority: 2},
		{AgentID: "agent-3", Role: "validator", Priority: 3},
	}
	ensemble.ExecutionMode = "parallel"
	ensemble.AggregationStrategy = "consensus"

	result, err := validator.Validate(ensemble, StrictLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestEnsembleValidator_Validate_InvalidType(t *testing.T) {
	validator := NewEnsembleValidator()

	skill := domain.NewSkill("Test Skill", "A test skill", "1.0.0", "tester")

	result, err := validator.Validate(skill, BasicLevel)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an Ensemble")
}
