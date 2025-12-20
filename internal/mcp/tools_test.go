package mcp

import (
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestSimpleElement_GetMetadata(t *testing.T) {
	now := time.Now()
	metadata := domain.ElementMetadata{
		ID:          "test-id",
		Type:        domain.PersonaElement,
		Name:        "Test Element",
		Description: "Test Description",
		Version:     "1.0.0",
		Author:      "Test Author",
		Tags:        []string{"tag1", "tag2"},
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	elem := &SimpleElement{metadata: metadata}

	assert.Equal(t, metadata, elem.GetMetadata())
}

func TestSimpleElement_Validate(t *testing.T) {
	elem := &SimpleElement{}
	assert.NoError(t, elem.Validate())
}

func TestSimpleElement_GetType(t *testing.T) {
	elem := &SimpleElement{
		metadata: domain.ElementMetadata{
			Type: domain.SkillElement,
		},
	}
	assert.Equal(t, domain.SkillElement, elem.GetType())
}

func TestSimpleElement_GetID(t *testing.T) {
	elem := &SimpleElement{
		metadata: domain.ElementMetadata{
			ID: "test-id-123",
		},
	}
	assert.Equal(t, "test-id-123", elem.GetID())
}

func TestSimpleElement_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
	}{
		{"active element", true},
		{"inactive element", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem := &SimpleElement{
				metadata: domain.ElementMetadata{
					IsActive: tt.isActive,
				},
			}
			assert.Equal(t, tt.isActive, elem.IsActive())
		})
	}
}

func TestSimpleElement_Activate(t *testing.T) {
	now := time.Now()
	elem := &SimpleElement{
		metadata: domain.ElementMetadata{
			IsActive:  false,
			UpdatedAt: now,
		},
	}

	err := elem.Activate()
	assert.NoError(t, err)
	assert.True(t, elem.IsActive())
	assert.True(t, elem.GetMetadata().UpdatedAt.After(now))
}

func TestSimpleElement_Deactivate(t *testing.T) {
	now := time.Now()
	elem := &SimpleElement{
		metadata: domain.ElementMetadata{
			IsActive:  true,
			UpdatedAt: now,
		},
	}

	err := elem.Deactivate()
	assert.NoError(t, err)
	assert.False(t, elem.IsActive())
	assert.True(t, elem.GetMetadata().UpdatedAt.After(now))
}

func TestInputOutputStructures(t *testing.T) {
	t.Run("ListElementsInput", func(t *testing.T) {
		isActive := true
		input := ListElementsInput{
			Type:     "persona",
			IsActive: &isActive,
			Tags:     "tag1,tag2",
		}
		assert.Equal(t, "persona", input.Type)
		assert.Equal(t, true, *input.IsActive)
		assert.Equal(t, "tag1,tag2", input.Tags)
	})

	t.Run("ListElementsOutput", func(t *testing.T) {
		output := ListElementsOutput{
			Elements: []map[string]interface{}{
				{"id": "1", "name": "Test"},
			},
			Total: 1,
		}
		assert.Equal(t, 1, output.Total)
		assert.Len(t, output.Elements, 1)
	})

	t.Run("GetElementInput", func(t *testing.T) {
		input := GetElementInput{ID: "test-id"}
		assert.Equal(t, "test-id", input.ID)
	})

	t.Run("GetElementOutput", func(t *testing.T) {
		output := GetElementOutput{
			Element: map[string]interface{}{"id": "test-id"},
		}
		assert.NotNil(t, output.Element)
	})

	t.Run("CreateElementInput", func(t *testing.T) {
		input := CreateElementInput{
			Type:        "skill",
			Name:        "Test Skill",
			Description: "Test Description",
			Version:     "1.0.0",
			Author:      "Test Author",
			Tags:        []string{"tag1"},
			IsActive:    true,
		}
		assert.Equal(t, "skill", input.Type)
		assert.Equal(t, "Test Skill", input.Name)
		assert.True(t, input.IsActive)
	})

	t.Run("CreateElementOutput", func(t *testing.T) {
		output := CreateElementOutput{
			ID:      "new-id",
			Element: map[string]interface{}{"id": "new-id"},
		}
		assert.Equal(t, "new-id", output.ID)
	})

	t.Run("UpdateElementInput", func(t *testing.T) {
		isActive := false
		input := UpdateElementInput{
			ID:          "update-id",
			Name:        "Updated Name",
			Description: "Updated Description",
			Tags:        []string{"newtag"},
			IsActive:    &isActive,
		}
		assert.Equal(t, "update-id", input.ID)
		assert.False(t, *input.IsActive)
	})

	t.Run("UpdateElementOutput", func(t *testing.T) {
		output := UpdateElementOutput{
			Element: map[string]interface{}{"id": "updated"},
		}
		assert.NotNil(t, output.Element)
	})

	t.Run("DeleteElementInput", func(t *testing.T) {
		input := DeleteElementInput{ID: "delete-id"}
		assert.Equal(t, "delete-id", input.ID)
	})

	t.Run("DeleteElementOutput", func(t *testing.T) {
		output := DeleteElementOutput{
			Success: true,
			Message: "Deleted successfully",
		}
		assert.True(t, output.Success)
		assert.Contains(t, output.Message, "successfully")
	})
}
