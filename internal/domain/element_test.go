package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateElementType(t *testing.T) {
	tests := []struct {
		name     string
		elemType ElementType
		want     bool
	}{
		{"valid persona type", PersonaElement, true},
		{"valid skill type", SkillElement, true},
		{"valid template type", TemplateElement, true},
		{"valid agent type", AgentElement, true},
		{"valid memory type", MemoryElement, true},
		{"valid ensemble type", EnsembleElement, true},
		{"invalid type - empty", ElementType(""), false},
		{"invalid type - unknown", ElementType("unknown"), false},
		{"invalid type - wrong case", ElementType("Persona"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateElementType(tt.elemType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateElementID(t *testing.T) {
	tests := []struct {
		name        string
		elementType ElementType
		elementName string
		wantPrefix  string
	}{
		{"persona element", PersonaElement, "alice", "persona_alice_"},
		{"skill element", SkillElement, "code_review", "skill_code_review_"},
		{"template element", TemplateElement, "api_doc", "template_api_doc_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateElementID(tt.elementType, tt.elementName)
			assert.Contains(t, got, tt.wantPrefix)
			assert.Regexp(t, `^[a-z]+_[a-z_]+_\d{8}-\d{6}$`, got)

			// Test uniqueness
			time.Sleep(1 * time.Second)
			got2 := GenerateElementID(tt.elementType, tt.elementName)
			assert.NotEqual(t, got, got2, "IDs should be unique")
		})
	}
}

func TestElementMetadata(t *testing.T) {
	t.Run("valid metadata creation", func(t *testing.T) {
		meta := ElementMetadata{
			ID:          "persona_test_20250101-120000",
			Type:        PersonaElement,
			Name:        "test-persona",
			Description: "Test description",
			Version:     "1.0.0",
			Author:      "test@example.com",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NotEmpty(t, meta.ID)
		assert.Equal(t, PersonaElement, meta.Type)
		assert.Equal(t, "test-persona", meta.Name)
	})
}

func TestElementType_String(t *testing.T) {
	tests := []struct {
		name     string
		elemType ElementType
		want     string
	}{
		{"persona type", PersonaElement, "persona"},
		{"skill type", SkillElement, "skill"},
		{"template type", TemplateElement, "template"},
		{"agent type", AgentElement, "agent"},
		{"memory type", MemoryElement, "memory"},
		{"ensemble type", EnsembleElement, "ensemble"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.elemType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestElementFilter(t *testing.T) {
	t.Run("filter with type", func(t *testing.T) {
		elemType := PersonaElement
		filter := ElementFilter{Type: &elemType}
		require.NotNil(t, filter.Type)
		assert.Equal(t, PersonaElement, *filter.Type)
	})

	t.Run("filter with active status", func(t *testing.T) {
		isActive := true
		filter := ElementFilter{IsActive: &isActive}
		require.NotNil(t, filter.IsActive)
		assert.True(t, *filter.IsActive)
	})

	t.Run("filter with tags", func(t *testing.T) {
		filter := ElementFilter{Tags: []string{"tag1", "tag2"}}
		assert.Len(t, filter.Tags, 2)
		assert.Contains(t, filter.Tags, "tag1")
		assert.Contains(t, filter.Tags, "tag2")
	})

	t.Run("filter with pagination", func(t *testing.T) {
		filter := ElementFilter{Limit: 10, Offset: 20}
		assert.Equal(t, 10, filter.Limit)
		assert.Equal(t, 20, filter.Offset)
	})
}
