package validation

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentValidator_SupportedType(t *testing.T) {
	validator := NewAgentValidator()
	assert.Equal(t, domain.AgentElement, validator.SupportedType())
}

func TestAgentValidator_Validate_Success(t *testing.T) {
	validator := NewAgentValidator()

	agent := domain.NewAgent("Test Agent", "A comprehensive test agent", "1.0.0", "tester")
	agent.Goals = []string{"achieve testing excellence"}
	agent.Actions = []domain.AgentAction{
		{Name: "test-action", Type: "tool"},
	}

	result, err := validator.Validate(agent, BasicLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(domain.AgentElement), result.ElementType)
	assert.Equal(t, agent.GetID(), result.ElementID)
}

func TestAgentValidator_Validate_ComprehensiveLevel(t *testing.T) {
	validator := NewAgentValidator()

	agent := domain.NewAgent("Test Agent", "A test agent", "1.0.0", "tester")
	agent.Goals = []string{"test goal"}

	result, err := validator.Validate(agent, ComprehensiveLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAgentValidator_Validate_StrictLevel(t *testing.T) {
	validator := NewAgentValidator()

	agent := domain.NewAgent("Test Agent", "A comprehensive test agent", "1.0.0", "tester")
	agent.Goals = []string{"achieve comprehensive testing"}
	agent.Actions = []domain.AgentAction{
		{Name: "comprehensive-test-action", Type: "tool"},
	}

	result, err := validator.Validate(agent, StrictLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAgentValidator_Validate_InvalidType(t *testing.T) {
	validator := NewAgentValidator()

	persona := domain.NewPersona("Test Persona", "A test persona", "1.0.0", "tester")

	result, err := validator.Validate(persona, BasicLevel)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an Agent")
}
