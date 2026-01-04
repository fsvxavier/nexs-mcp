//go:build !noonnx && integration
// +build !noonnx,integration

package integration_test

import (
	"sync"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// mockElementRepository is a mock implementation of ElementRepository for testing.
type mockElementRepository struct {
	elements map[string]domain.Element
	mu       sync.RWMutex
}

// newMockElementRepository creates a new mock repository.
func newMockElementRepository() *mockElementRepository {
	return &mockElementRepository{
		elements: make(map[string]domain.Element),
	}
}

// Create creates a new element.
func (m *mockElementRepository) Create(element domain.Element) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.elements[element.GetID()] = element
	return nil
}

// GetByID retrieves an element by its ID.
func (m *mockElementRepository) GetByID(id string) (domain.Element, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	elem, exists := m.elements[id]
	if !exists {
		return nil, domain.ErrElementNotFound
	}
	return elem, nil
}

// Update updates an existing element.
func (m *mockElementRepository) Update(element domain.Element) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.elements[element.GetID()]; !exists {
		return domain.ErrElementNotFound
	}
	m.elements[element.GetID()] = element
	return nil
}

// Delete deletes an element by its ID.
func (m *mockElementRepository) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.elements[id]; !exists {
		return domain.ErrElementNotFound
	}
	delete(m.elements, id)
	return nil
}

// List lists all elements with optional filters.
func (m *mockElementRepository) List(filters domain.ElementFilter) ([]domain.Element, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	elements := make([]domain.Element, 0, len(m.elements))
	for _, elem := range m.elements {
		elements = append(elements, elem)
	}
	return elements, nil
}
