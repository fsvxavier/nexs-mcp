package mcp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test elements
func createTestElement(elementType, name string) *SimpleElement {
	elemType := domain.ElementType(elementType)
	id := domain.GenerateElementID(elemType, name)
	now := time.Now()

	return &SimpleElement{
		metadata: domain.ElementMetadata{
			ID:          id,
			Type:        elemType,
			Name:        name,
			Description: "Test description",
			Version:     "1.0.0",
			Author:      "Test Author",
			Tags:        []string{"test"},
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

func TestGetElementTool(t *testing.T) {
	repo := NewMockElementRepository()
	tool := NewGetElementTool(repo)

	assert.Equal(t, "get_element", tool.Name)
	assert.NotNil(t, tool.InputSchema)
	assert.NotNil(t, tool.Handler)

	t.Run("Success", func(t *testing.T) {
		// Create element first
		element := createTestElement("persona", "Test Persona")
		err := repo.Create(element)
		require.NoError(t, err)

		args := json.RawMessage(`{"id":"` + element.GetID() + `"}`)
		result, err := tool.Handler(args)

		require.NoError(t, err)
		assert.NotNil(t, result)

		resultMap := result.(map[string]interface{})
		assert.Contains(t, resultMap, "element")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		args := json.RawMessage(`{invalid}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid arguments")
	})

	t.Run("MissingID", func(t *testing.T) {
		args := json.RawMessage(`{}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("ElementNotFound", func(t *testing.T) {
		args := json.RawMessage(`{"id":"nonexistent"}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get element")
	})
}

func TestCreateElementTool(t *testing.T) {
	repo := NewMockElementRepository()
	tool := NewCreateElementTool(repo)

	assert.Equal(t, "create_element", tool.Name)
	assert.NotNil(t, tool.InputSchema)
	assert.NotNil(t, tool.Handler)

	t.Run("Success", func(t *testing.T) {
		args := json.RawMessage(`{
			"type": "persona",
			"name": "Test Persona",
			"description": "Test description",
			"version": "1.0.0",
			"author": "Test Author",
			"tags": ["test"]
		}`)

		result, err := tool.Handler(args)

		require.NoError(t, err)
		assert.NotNil(t, result)

		resultMap := result.(map[string]interface{})
		assert.Contains(t, resultMap, "id")
		assert.Contains(t, resultMap, "element")

		id := resultMap["id"].(string)
		assert.NotEmpty(t, id)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		args := json.RawMessage(`{invalid}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid arguments")
	})

	t.Run("InvalidType", func(t *testing.T) {
		args := json.RawMessage(`{
			"type": "invalid_type",
			"name": "Test",
			"version": "1.0.0",
			"author": "Test"
		}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid element type")
	})
}

func TestUpdateElementTool(t *testing.T) {
	repo := NewMockElementRepository()
	tool := NewUpdateElementTool(repo)

	assert.Equal(t, "update_element", tool.Name)
	assert.NotNil(t, tool.InputSchema)
	assert.NotNil(t, tool.Handler)

	t.Run("Success", func(t *testing.T) {
		// Create element first
		element := createTestElement("persona", "Original Name")
		err := repo.Create(element)
		require.NoError(t, err)

		args := json.RawMessage(`{
			"id": "` + element.GetID() + `",
			"name": "Updated Name",
			"description": "Updated description"
		}`)

		result, err := tool.Handler(args)

		require.NoError(t, err)
		assert.NotNil(t, result)

		resultMap := result.(map[string]interface{})
		assert.Contains(t, resultMap, "id")
		assert.Contains(t, resultMap, "element")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		args := json.RawMessage(`{invalid}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid arguments")
	})

	t.Run("MissingID", func(t *testing.T) {
		args := json.RawMessage(`{"name":"Test"}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("ElementNotFound", func(t *testing.T) {
		args := json.RawMessage(`{"id":"nonexistent","name":"Test"}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get element")
	})

	t.Run("UpdateIsActive", func(t *testing.T) {
		element := createTestElement("skill", "Test Skill")
		err := repo.Create(element)
		require.NoError(t, err)

		args := json.RawMessage(`{
			"id": "` + element.GetID() + `",
			"is_active": false
		}`)

		result, err := tool.Handler(args)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestDeleteElementTool(t *testing.T) {
	repo := NewMockElementRepository()
	tool := NewDeleteElementTool(repo)

	assert.Equal(t, "delete_element", tool.Name)
	assert.NotNil(t, tool.InputSchema)
	assert.NotNil(t, tool.Handler)

	t.Run("Success", func(t *testing.T) {
		// Create element first
		element := createTestElement("template", "Test Template")
		err := repo.Create(element)
		require.NoError(t, err)

		args := json.RawMessage(`{"id":"` + element.GetID() + `"}`)
		result, err := tool.Handler(args)

		require.NoError(t, err)
		assert.NotNil(t, result)

		resultMap := result.(map[string]interface{})
		assert.Equal(t, element.GetID(), resultMap["id"])
		assert.True(t, resultMap["deleted"].(bool))
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		args := json.RawMessage(`{invalid}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid arguments")
	})

	t.Run("MissingID", func(t *testing.T) {
		args := json.RawMessage(`{}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("ElementNotFound", func(t *testing.T) {
		args := json.RawMessage(`{"id":"nonexistent"}`)
		_, err := tool.Handler(args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete element")
	})
}

func TestSimpleElement(t *testing.T) {
	element := createTestElement("agent", "Test Agent")

	t.Run("GetMetadata", func(t *testing.T) {
		metadata := element.GetMetadata()
		assert.Equal(t, "agent", string(metadata.Type))
		assert.Equal(t, "Test Agent", metadata.Name)
		assert.True(t, metadata.IsActive)
	})

	t.Run("Validate", func(t *testing.T) {
		err := element.Validate()
		assert.NoError(t, err)
	})

	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, "agent", string(element.GetType()))
	})

	t.Run("GetID", func(t *testing.T) {
		assert.NotEmpty(t, element.GetID())
	})

	t.Run("IsActive", func(t *testing.T) {
		assert.True(t, element.IsActive())
	})

	t.Run("Activate", func(t *testing.T) {
		err := element.Activate()
		assert.NoError(t, err)
		assert.True(t, element.IsActive())
	})

	t.Run("Deactivate", func(t *testing.T) {
		err := element.Deactivate()
		assert.NoError(t, err)
		assert.False(t, element.IsActive())
	})
}
