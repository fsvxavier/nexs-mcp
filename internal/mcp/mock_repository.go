package mcp

import (
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// MockElementRepository is a mock implementation of ElementRepository for testing.
type MockElementRepository struct {
	elements map[string]domain.Element
}

// NewMockElementRepository creates a new mock repository.
func NewMockElementRepository() *MockElementRepository {
	return &MockElementRepository{
		elements: make(map[string]domain.Element),
	}
}

// Create creates a new element.
func (m *MockElementRepository) Create(element domain.Element) error {
	m.elements[element.GetID()] = element
	return nil
}

// GetByID retrieves an element by its ID.
func (m *MockElementRepository) GetByID(id string) (domain.Element, error) {
	elem, exists := m.elements[id]
	if !exists {
		return nil, domain.ErrElementNotFound
	}
	return elem, nil
}

// Update updates an existing element.
func (m *MockElementRepository) Update(element domain.Element) error {
	if _, exists := m.elements[element.GetID()]; !exists {
		return domain.ErrElementNotFound
	}
	m.elements[element.GetID()] = element
	return nil
}

// Delete deletes an element by its ID.
func (m *MockElementRepository) Delete(id string) error {
	if _, exists := m.elements[id]; !exists {
		return domain.ErrElementNotFound
	}
	delete(m.elements, id)
	return nil
}

// List lists all elements with optional filtering.
func (m *MockElementRepository) List(filter domain.ElementFilter) ([]domain.Element, error) {
	elements := make([]domain.Element, 0)
	for _, elem := range m.elements {
		elements = append(elements, elem)
	}
	return elements, nil
}

// Exists checks if an element exists.
func (m *MockElementRepository) Exists(id string) (bool, error) {
	_, exists := m.elements[id]
	return exists, nil
}
