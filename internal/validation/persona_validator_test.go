package validation

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonaValidator_SupportedType(t *testing.T) {
	validator := NewPersonaValidator()
	assert.Equal(t, domain.PersonaElement, validator.SupportedType())
}

func TestPersonaValidator_Validate_Success(t *testing.T) {
	validator := NewPersonaValidator()

	persona := domain.NewPersona("Test Persona", "A test persona for validation", "1.0.0", "tester")

	result, err := validator.Validate(persona, BasicLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(domain.PersonaElement), result.ElementType)
	assert.Equal(t, persona.GetID(), result.ElementID)
}

func TestPersonaValidator_Validate_ComprehensiveLevel(t *testing.T) {
	validator := NewPersonaValidator()

	persona := domain.NewPersona("Test Persona", "Short description", "1.0.0", "tester")

	result, err := validator.Validate(persona, ComprehensiveLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestPersonaValidator_Validate_StrictLevel(t *testing.T) {
	validator := NewPersonaValidator()

	persona := domain.NewPersona("Test Persona", "A comprehensive test persona", "1.0.0", "tester")

	result, err := validator.Validate(persona, StrictLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestPersonaValidator_Validate_InvalidType(t *testing.T) {
	validator := NewPersonaValidator()

	skill := domain.NewSkill("Test Skill", "A test skill", "1.0.0", "tester")

	result, err := validator.Validate(skill, BasicLevel)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a Persona")
}
