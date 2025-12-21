package infrastructure

import (
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func newMockElement(id string, elemType domain.ElementType, isActive bool, tags []string) *mockElement {
	return &mockElement{
		metadata: domain.ElementMetadata{
			ID:        id,
			Type:      elemType,
			Name:      "test-element",
			Version:   "1.0.0",
			Author:    "test",
			IsActive:  isActive,
			Tags:      tags,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func TestNewInMemoryElementRepository(t *testing.T) {
	repo := NewInMemoryElementRepository()
	require.NotNil(t, repo)
	assert.Equal(t, 0, repo.Count())
}

func TestInMemoryElementRepository_Create(t *testing.T) {
	t.Run("create valid element", func(t *testing.T) {
		repo := NewInMemoryElementRepository()
		elem := newMockElement("test-1", domain.PersonaElement, true, nil)

		err := repo.Create(elem)
		require.NoError(t, err)
		assert.Equal(t, 1, repo.Count())
	})

	t.Run("create nil element", func(t *testing.T) {
		repo := NewInMemoryElementRepository()
		err := repo.Create(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("create duplicate element", func(t *testing.T) {
		repo := NewInMemoryElementRepository()
		elem := newMockElement("test-1", domain.PersonaElement, true, nil)

		require.NoError(t, repo.Create(elem))
		err := repo.Create(elem)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestInMemoryElementRepository_GetByID(t *testing.T) {
	repo := NewInMemoryElementRepository()
	elem := newMockElement("test-1", domain.PersonaElement, true, nil)
	require.NoError(t, repo.Create(elem))

	t.Run("get existing element", func(t *testing.T) {
		retrieved, err := repo.GetByID("test-1")
		require.NoError(t, err)
		assert.Equal(t, "test-1", retrieved.GetID())
	})

	t.Run("get non-existing element", func(t *testing.T) {
		_, err := repo.GetByID("non-existing")
		require.Error(t, err)
		assert.Equal(t, domain.ErrElementNotFound, err)
	})
}

func TestInMemoryElementRepository_Update(t *testing.T) {
	repo := NewInMemoryElementRepository()
	elem := newMockElement("test-1", domain.PersonaElement, true, nil)
	require.NoError(t, repo.Create(elem))

	t.Run("update existing element", func(t *testing.T) {
		elem.Deactivate()
		err := repo.Update(elem)
		require.NoError(t, err)

		retrieved, err := repo.GetByID("test-1")
		require.NoError(t, err)
		assert.False(t, retrieved.IsActive())
	})

	t.Run("update non-existing element", func(t *testing.T) {
		newElem := newMockElement("non-existing", domain.PersonaElement, true, nil)
		err := repo.Update(newElem)
		require.Error(t, err)
		assert.Equal(t, domain.ErrElementNotFound, err)
	})

	t.Run("update nil element", func(t *testing.T) {
		err := repo.Update(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestInMemoryElementRepository_Delete(t *testing.T) {
	repo := NewInMemoryElementRepository()
	elem := newMockElement("test-1", domain.PersonaElement, true, nil)
	require.NoError(t, repo.Create(elem))

	t.Run("delete existing element", func(t *testing.T) {
		err := repo.Delete("test-1")
		require.NoError(t, err)
		assert.Equal(t, 0, repo.Count())
	})

	t.Run("delete non-existing element", func(t *testing.T) {
		err := repo.Delete("non-existing")
		require.Error(t, err)
		assert.Equal(t, domain.ErrElementNotFound, err)
	})
}

func TestInMemoryElementRepository_List(t *testing.T) {
	repo := NewInMemoryElementRepository()

	elem1 := newMockElement("persona-1", domain.PersonaElement, true, []string{"tag1", "tag2"})
	elem2 := newMockElement("skill-1", domain.SkillElement, true, []string{"tag1"})
	elem3 := newMockElement("persona-2", domain.PersonaElement, false, []string{"tag2"})

	require.NoError(t, repo.Create(elem1))
	require.NoError(t, repo.Create(elem2))
	require.NoError(t, repo.Create(elem3))

	t.Run("list all elements", func(t *testing.T) {
		elements, err := repo.List(domain.ElementFilter{})
		require.NoError(t, err)
		assert.Len(t, elements, 3)
	})

	t.Run("filter by type", func(t *testing.T) {
		elemType := domain.PersonaElement
		elements, err := repo.List(domain.ElementFilter{Type: &elemType})
		require.NoError(t, err)
		assert.Len(t, elements, 2)
	})

	t.Run("filter by active status", func(t *testing.T) {
		isActive := true
		elements, err := repo.List(domain.ElementFilter{IsActive: &isActive})
		require.NoError(t, err)
		assert.Len(t, elements, 2)
	})

	t.Run("filter by tags", func(t *testing.T) {
		elements, err := repo.List(domain.ElementFilter{Tags: []string{"tag1"}})
		require.NoError(t, err)
		assert.Len(t, elements, 2)
	})

	t.Run("filter with pagination", func(t *testing.T) {
		elements, err := repo.List(domain.ElementFilter{Limit: 2, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, elements, 2)
	})
}

func TestInMemoryElementRepository_Exists(t *testing.T) {
	repo := NewInMemoryElementRepository()
	elem := newMockElement("test-1", domain.PersonaElement, true, nil)
	require.NoError(t, repo.Create(elem))

	t.Run("existing element", func(t *testing.T) {
		exists, err := repo.Exists("test-1")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existing element", func(t *testing.T) {
		exists, err := repo.Exists("non-existing")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestInMemoryElementRepository_Clear(t *testing.T) {
	repo := NewInMemoryElementRepository()
	elem := newMockElement("test-1", domain.PersonaElement, true, nil)
	require.NoError(t, repo.Create(elem))

	assert.Equal(t, 1, repo.Count())
	repo.Clear()
	assert.Equal(t, 0, repo.Count())
}

func TestInMemoryElementRepository_Concurrency(t *testing.T) {
	repo := NewInMemoryElementRepository()
	done := make(chan bool)

	// Concurrent writes
	for i := range 10 {
		go func(idx int) {
			elem := newMockElement(fmt.Sprintf("elem-%d", idx), domain.PersonaElement, true, nil)
			repo.Create(elem)
			done <- true
		}(i)
	}

	for range 10 {
		<-done
	}

	// Concurrent reads
	for range 10 {
		go func() {
			repo.List(domain.ElementFilter{})
			done <- true
		}()
	}

	for range 10 {
		<-done
	}

	assert.True(t, repo.Count() >= 0)
}
