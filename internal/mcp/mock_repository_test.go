package mcp

import (
	"encoding/json"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockElementRepository_Create(t *testing.T) {
	repo := NewMockElementRepository()
	elem := newMockElement("test-1", domain.PersonaElement)

	err := repo.Create(elem)
	require.NoError(t, err)

	retrieved, err := repo.GetByID("test-1")
	require.NoError(t, err)
	assert.Equal(t, "test-1", retrieved.GetID())
}

func TestMockElementRepository_GetByID(t *testing.T) {
	repo := NewMockElementRepository()

	t.Run("get non-existing element", func(t *testing.T) {
		_, err := repo.GetByID("non-existing")
		require.Error(t, err)
		assert.Equal(t, domain.ErrElementNotFound, err)
	})

	t.Run("get existing element", func(t *testing.T) {
		elem := newMockElement("test-1", domain.PersonaElement)
		require.NoError(t, repo.Create(elem))

		retrieved, err := repo.GetByID("test-1")
		require.NoError(t, err)
		assert.Equal(t, "test-1", retrieved.GetID())
	})
}

func TestMockElementRepository_Update(t *testing.T) {
	repo := NewMockElementRepository()
	elem := newMockElement("test-1", domain.PersonaElement)

	t.Run("update non-existing element", func(t *testing.T) {
		err := repo.Update(elem)
		require.Error(t, err)
		assert.Equal(t, domain.ErrElementNotFound, err)
	})

	t.Run("update existing element", func(t *testing.T) {
		require.NoError(t, repo.Create(elem))
		err := repo.Update(elem)
		require.NoError(t, err)
	})
}

func TestMockElementRepository_Delete(t *testing.T) {
	repo := NewMockElementRepository()

	t.Run("delete non-existing element", func(t *testing.T) {
		err := repo.Delete("non-existing")
		require.Error(t, err)
		assert.Equal(t, domain.ErrElementNotFound, err)
	})

	t.Run("delete existing element", func(t *testing.T) {
		elem := newMockElement("test-1", domain.PersonaElement)
		require.NoError(t, repo.Create(elem))

		err := repo.Delete("test-1")
		require.NoError(t, err)

		_, err = repo.GetByID("test-1")
		require.Error(t, err)
	})
}

func TestMockElementRepository_List(t *testing.T) {
	repo := NewMockElementRepository()

	t.Run("list empty repository", func(t *testing.T) {
		elements, err := repo.List(domain.ElementFilter{})
		require.NoError(t, err)
		assert.Empty(t, elements)
	})

	t.Run("list with elements", func(t *testing.T) {
		elem1 := newMockElement("test-1", domain.PersonaElement)
		elem2 := newMockElement("test-2", domain.SkillElement)
		require.NoError(t, repo.Create(elem1))
		require.NoError(t, repo.Create(elem2))

		elements, err := repo.List(domain.ElementFilter{})
		require.NoError(t, err)
		assert.Len(t, elements, 2)
	})
}

func TestMockElementRepository_Exists(t *testing.T) {
	repo := NewMockElementRepository()

	t.Run("non-existing element", func(t *testing.T) {
		exists, err := repo.Exists("non-existing")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("existing element", func(t *testing.T) {
		elem := newMockElement("test-1", domain.PersonaElement)
		require.NoError(t, repo.Create(elem))

		exists, err := repo.Exists("test-1")
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestListElementsTool_EdgeCases(t *testing.T) {
	repo := NewMockElementRepository()
	tool := NewListElementsTool(repo)

	t.Run("execute with nil args", func(t *testing.T) {
		result, err := tool.Execute(nil)
		require.NoError(t, err)

		resultMap := result.(map[string]interface{})
		assert.Equal(t, 0, resultMap["count"])
	})

	t.Run("execute with complex filter", func(t *testing.T) {
		// The mock repository doesn't implement filtering
		// Just verify the tool executes without error
		args := json.RawMessage(`{"type": "persona"}`)
		result, err := tool.Execute(args)
		require.NoError(t, err)

		resultMap := result.(map[string]interface{})
		assert.NotNil(t, resultMap["elements"])
		assert.NotNil(t, resultMap["count"])
	})

	t.Run("tool error handling", func(t *testing.T) {
		// Test with malformed JSON
		_, err := tool.Execute(json.RawMessage(`{`))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid arguments")
	})
}

func BenchmarkServer_RegisterTool(b *testing.B) {
	server := NewServer("bench", "1.0.0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool := &Tool{
			Name:    "tool_" + string(rune(i)),
			Handler: func(args json.RawMessage) (interface{}, error) { return nil, nil },
		}
		server.RegisterTool(tool)
	}
}

func BenchmarkServer_GetTool(b *testing.B) {
	server := NewServer("bench", "1.0.0")
	tool := &Tool{
		Name:    "test_tool",
		Handler: func(args json.RawMessage) (interface{}, error) { return nil, nil },
	}
	server.RegisterTool(tool)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.GetTool("test_tool")
	}
}

func BenchmarkListElementsTool_Execute(b *testing.B) {
	repo := NewMockElementRepository()

	// Populate repository
	for i := 0; i < 100; i++ {
		elem := newMockElement("elem-"+string(rune(i)), domain.PersonaElement)
		repo.Create(elem)
	}

	tool := NewListElementsTool(repo)
	args := json.RawMessage(`{"limit": 10}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(args)
	}
}

// Helper function for creating mock elements
func newMockElement(id string, elemType domain.ElementType) domain.Element {
	return &mockElement{
		metadata: domain.ElementMetadata{
			ID:       id,
			Type:     elemType,
			Name:     "test",
			Version:  "1.0.0",
			Author:   "test",
			IsActive: true,
		},
	}
}

type mockElement struct {
	metadata domain.ElementMetadata
}

func (m *mockElement) GetMetadata() domain.ElementMetadata { return m.metadata }
func (m *mockElement) Validate() error                     { return nil }
func (m *mockElement) GetType() domain.ElementType         { return m.metadata.Type }
func (m *mockElement) GetID() string                       { return m.metadata.ID }
func (m *mockElement) IsActive() bool                      { return m.metadata.IsActive }
func (m *mockElement) Activate() error {
	m.metadata.IsActive = true
	return nil
}
func (m *mockElement) Deactivate() error {
	m.metadata.IsActive = false
	return nil
}
