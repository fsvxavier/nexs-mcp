package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusConstants(t *testing.T) {
	assert.Equal(t, "success", StatusSuccess)
	assert.Equal(t, "failed", StatusFailed)
	assert.Equal(t, "error", StatusError)

	// Ensure they're distinct
	assert.NotEqual(t, StatusSuccess, StatusFailed)
	assert.NotEqual(t, StatusSuccess, StatusError)
	assert.NotEqual(t, StatusFailed, StatusError)
}

func TestElementTypeConstants(t *testing.T) {
	elementTypes := []string{
		ElementTypePersona,
		ElementTypeSkill,
		ElementTypeTemplate,
		ElementTypeAgent,
		ElementTypeMemory,
		ElementTypeEnsemble,
	}

	// Verify values
	assert.Equal(t, "persona", ElementTypePersona)
	assert.Equal(t, "skill", ElementTypeSkill)
	assert.Equal(t, "template", ElementTypeTemplate)
	assert.Equal(t, "agent", ElementTypeAgent)
	assert.Equal(t, "memory", ElementTypeMemory)
	assert.Equal(t, "ensemble", ElementTypeEnsemble)

	// Ensure all are unique
	uniqueTypes := make(map[string]bool)
	for _, et := range elementTypes {
		assert.False(t, uniqueTypes[et], "Element type %s is duplicated", et)
		uniqueTypes[et] = true
	}
	assert.Len(t, uniqueTypes, 6)
}

func TestMethodConstants(t *testing.T) {
	methods := []string{
		MethodMention,
		MethodKeyword,
		MethodSemantic,
		MethodPattern,
	}

	// Verify values
	assert.Equal(t, "mention", MethodMention)
	assert.Equal(t, "keyword", MethodKeyword)
	assert.Equal(t, "semantic", MethodSemantic)
	assert.Equal(t, "pattern", MethodPattern)

	// Ensure all are unique
	uniqueMethods := make(map[string]bool)
	for _, m := range methods {
		assert.False(t, uniqueMethods[m], "Method %s is duplicated", m)
		uniqueMethods[m] = true
	}
	assert.Len(t, uniqueMethods, 4)
}

func TestSortOrderConstants(t *testing.T) {
	assert.Equal(t, "asc", SortOrderAsc)
	assert.Equal(t, "desc", SortOrderDesc)

	// Ensure they're distinct
	assert.NotEqual(t, SortOrderAsc, SortOrderDesc)
}

func TestCommonConstants(t *testing.T) {
	assert.Equal(t, "all", SelectorAll)
	assert.Equal(t, "main", BranchMain)

	// Ensure they're distinct
	assert.NotEqual(t, SelectorAll, BranchMain)
}

func TestConstantImmutability(t *testing.T) {
	// Verify constants can be used in comparisons and switch statements
	status := StatusSuccess
	switch status {
	case StatusSuccess:
		// Expected path
	case StatusFailed, StatusError:
		t.Error("Unexpected status")
	default:
		t.Error("Unknown status")
	}

	elementType := ElementTypePersona
	switch elementType {
	case ElementTypePersona:
		// Expected path
	case ElementTypeSkill, ElementTypeTemplate, ElementTypeAgent, ElementTypeMemory, ElementTypeEnsemble:
		t.Error("Unexpected element type")
	default:
		t.Error("Unknown element type")
	}
}
