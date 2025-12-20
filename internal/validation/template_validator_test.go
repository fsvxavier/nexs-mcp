package validation

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateValidator_SupportedType(t *testing.T) {
	validator := NewTemplateValidator()
	assert.Equal(t, domain.TemplateElement, validator.SupportedType())
}

func TestTemplateValidator_Validate_Success(t *testing.T) {
	validator := NewTemplateValidator()

	template := domain.NewTemplate("Test Template", "A comprehensive test template", "1.0.0", "tester")

	result, err := validator.Validate(template, BasicLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(domain.TemplateElement), result.ElementType)
	assert.Equal(t, template.GetID(), result.ElementID)
}

func TestTemplateValidator_Validate_ComprehensiveLevel(t *testing.T) {
	validator := NewTemplateValidator()

	template := domain.NewTemplate("Test Template", "A test template", "1.0.0", "tester")

	result, err := validator.Validate(template, ComprehensiveLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestTemplateValidator_Validate_StrictLevel(t *testing.T) {
	validator := NewTemplateValidator()

	template := domain.NewTemplate("Test Template", "A comprehensive test template", "1.0.0", "tester")

	result, err := validator.Validate(template, StrictLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestTemplateValidator_Validate_InvalidType(t *testing.T) {
	validator := NewTemplateValidator()

	skill := domain.NewSkill("Test Skill", "A test skill", "1.0.0", "tester")

	result, err := validator.Validate(skill, BasicLevel)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a Template")
}
