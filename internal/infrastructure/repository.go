package infrastructure

import (
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// InMemoryElementRepository implements ElementRepository using in-memory storage
type InMemoryElementRepository struct {
	mu       sync.RWMutex
	elements map[string]domain.Element
}

// NewInMemoryElementRepository creates a new in-memory repository
func NewInMemoryElementRepository() *InMemoryElementRepository {
	return &InMemoryElementRepository{
		elements: make(map[string]domain.Element),
	}
}

// Create creates a new element
func (r *InMemoryElementRepository) Create(element domain.Element) error {
	if element == nil {
		return fmt.Errorf("element cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := element.GetID()
	if _, exists := r.elements[id]; exists {
		return fmt.Errorf("element with ID %s already exists", id)
	}

	r.elements[id] = element
	return nil
}

// GetByID retrieves an element by its ID
func (r *InMemoryElementRepository) GetByID(id string) (domain.Element, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	element, exists := r.elements[id]
	if !exists {
		return nil, domain.ErrElementNotFound
	}

	return element, nil
}

// Update updates an existing element
func (r *InMemoryElementRepository) Update(element domain.Element) error {
	if element == nil {
		return fmt.Errorf("element cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := element.GetID()
	if _, exists := r.elements[id]; !exists {
		return domain.ErrElementNotFound
	}

	r.elements[id] = element
	return nil
}

// Delete deletes an element by its ID
func (r *InMemoryElementRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.elements[id]; !exists {
		return domain.ErrElementNotFound
	}

	delete(r.elements, id)
	return nil
}

// List lists all elements with optional filtering
func (r *InMemoryElementRepository) List(filter domain.ElementFilter) ([]domain.Element, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	elements := make([]domain.Element, 0)

	for _, elem := range r.elements {
		// Apply type filter
		if filter.Type != nil && elem.GetType() != *filter.Type {
			continue
		}

		// Apply active status filter
		if filter.IsActive != nil && elem.IsActive() != *filter.IsActive {
			continue
		}

		// Apply tags filter (if element has all required tags)
		if len(filter.Tags) > 0 {
			meta := elem.GetMetadata()
			hasAllTags := true
			for _, requiredTag := range filter.Tags {
				found := false
				for _, elemTag := range meta.Tags {
					if elemTag == requiredTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		elements = append(elements, elem)
	}

	// Apply pagination
	start := filter.Offset
	if start > len(elements) {
		return []domain.Element{}, nil
	}

	end := start + filter.Limit
	if filter.Limit > 0 && end < len(elements) {
		elements = elements[start:end]
	} else if start > 0 {
		elements = elements[start:]
	}

	return elements, nil
}

// Exists checks if an element exists
func (r *InMemoryElementRepository) Exists(id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.elements[id]
	return exists, nil
}

// Count returns the total number of elements
func (r *InMemoryElementRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.elements)
}

// Clear removes all elements (useful for testing)
func (r *InMemoryElementRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.elements = make(map[string]domain.Element)
}
