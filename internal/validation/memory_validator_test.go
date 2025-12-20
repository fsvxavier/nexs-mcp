package validation

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryValidator_SupportedType(t *testing.T) {
	validator := NewMemoryValidator()
	assert.Equal(t, domain.MemoryElement, validator.SupportedType())
}

func TestMemoryValidator_Validate_Success(t *testing.T) {
	validator := NewMemoryValidator()

	memory := domain.NewMemory("Test Memory", "A comprehensive test memory", "1.0.0", "tester")

	result, err := validator.Validate(memory, BasicLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(domain.MemoryElement), result.ElementType)
	assert.Equal(t, memory.GetID(), result.ElementID)
}

func TestMemoryValidator_Validate_ComprehensiveLevel(t *testing.T) {
	validator := NewMemoryValidator()

	memory := domain.NewMemory("Test Memory", "A test memory", "1.0.0", "tester")

	result, err := validator.Validate(memory, ComprehensiveLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMemoryValidator_Validate_StrictLevel(t *testing.T) {
	validator := NewMemoryValidator()

	memory := domain.NewMemory("Test Memory", "A comprehensive test memory", "1.0.0", "tester")

	result, err := validator.Validate(memory, StrictLevel)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMemoryValidator_Validate_InvalidType(t *testing.T) {
	validator := NewMemoryValidator()

	skill := domain.NewSkill("Test Skill", "A test skill", "1.0.0", "tester")

	result, err := validator.Validate(skill, BasicLevel)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a Memory")
}
