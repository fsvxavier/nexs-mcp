package validation

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillValidator_SupportedType(t *testing.T) {
	validator := NewSkillValidator()
	assert.Equal(t, domain.SkillElement, validator.SupportedType())
}

func TestSkillValidator_Validate_Success(t *testing.T) {
	validator := NewSkillValidator()

	skill := domain.NewSkill("Test Skill", "A comprehensive test skill for validation", "1.0.0", "tester")

	result, err := validator.Validate(skill, BasicLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(domain.SkillElement), result.ElementType)
	assert.Equal(t, skill.GetID(), result.ElementID)
}

func TestSkillValidator_Validate_ComprehensiveLevel(t *testing.T) {
	validator := NewSkillValidator()

	skill := domain.NewSkill("Test Skill", "A test skill", "1.0.0", "tester")

	result, err := validator.Validate(skill, ComprehensiveLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSkillValidator_Validate_StrictLevel(t *testing.T) {
	validator := NewSkillValidator()

	skill := domain.NewSkill("Test Skill", "A comprehensive test skill", "1.0.0", "tester")

	result, err := validator.Validate(skill, StrictLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSkillValidator_Validate_InvalidType(t *testing.T) {
	validator := NewSkillValidator()

	persona := domain.NewPersona("Test Persona", "A test persona", "1.0.0", "tester")

	result, err := validator.Validate(persona, BasicLevel)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a Skill")
}
