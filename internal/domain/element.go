package domain

import (
	"errors"
	"time"
)

// ElementType represents the type of an element.
type ElementType string

const (
	PersonaElement  ElementType = "persona"
	SkillElement    ElementType = "skill"
	TemplateElement ElementType = "template"
	AgentElement    ElementType = "agent"
	MemoryElement   ElementType = "memory"
	EnsembleElement ElementType = "ensemble"
)

// Common errors.
var (
	ErrInvalidElementType = errors.New("invalid element type")
	ErrInvalidElementID   = errors.New("invalid element ID")
	ErrElementNotFound    = errors.New("element not found")
	ErrValidationFailed   = errors.New("validation failed")
)

// ElementMetadata contains common metadata for all elements.
type ElementMetadata struct {
	ID          string                 `json:"id"               validate:"required"`
	Type        ElementType            `json:"type"             validate:"required,oneof=persona skill template agent memory ensemble"`
	Name        string                 `json:"name"             validate:"required,min=3,max=100"`
	Description string                 `json:"description"      validate:"max=500"`
	Version     string                 `json:"version"          validate:"required,semver"`
	Author      string                 `json:"author"           validate:"required"`
	Tags        []string               `json:"tags,omitempty"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Custom      map[string]interface{} `json:"custom,omitempty"`
}

// ToMap converts ElementMetadata to a map for JSON serialization.
func (m ElementMetadata) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          m.ID,
		"type":        string(m.Type),
		"name":        m.Name,
		"description": m.Description,
		"version":     m.Version,
		"author":      m.Author,
		"tags":        m.Tags,
		"is_active":   m.IsActive,
		"created_at":  m.CreatedAt,
		"updated_at":  m.UpdatedAt,
		"custom":      m.Custom,
	}
}

// Validate validates the element metadata.
func (m ElementMetadata) Validate() error {
	if m.ID == "" {
		return ErrInvalidElementID
	}
	if !ValidateElementType(m.Type) {
		return ErrInvalidElementType
	}
	if len(m.Name) < 3 || len(m.Name) > 100 {
		return errors.New("name must be between 3 and 100 characters")
	}
	if len(m.Description) > 500 {
		return errors.New("description must not exceed 500 characters")
	}
	if m.Version == "" {
		return errors.New("version is required")
	}
	if m.Author == "" {
		return errors.New("author is required")
	}
	return nil
}

// Element is the base interface for all element types.
type Element interface {
	// GetMetadata returns the element's metadata
	GetMetadata() ElementMetadata

	// Validate checks if the element is valid
	Validate() error

	// GetType returns the element type
	GetType() ElementType

	// GetID returns the element ID
	GetID() string

	// IsActive returns whether the element is active
	IsActive() bool

	// Activate activates the element
	Activate() error

	// Deactivate deactivates the element
	Deactivate() error
}

// ElementRepository defines the interface for element storage operations.
type ElementRepository interface {
	// Create creates a new element
	Create(element Element) error

	// GetByID retrieves an element by its ID
	GetByID(id string) (Element, error)

	// Update updates an existing element
	Update(element Element) error

	// Delete deletes an element by its ID
	Delete(id string) error

	// List lists all elements with optional filtering
	List(filter ElementFilter) ([]Element, error)

	// Exists checks if an element exists
	Exists(id string) (bool, error)
}

// ElementFilter defines filtering options for listing elements.
type ElementFilter struct {
	Type     *ElementType `json:"type,omitempty"`
	IsActive *bool        `json:"is_active,omitempty"`
	Tags     []string     `json:"tags,omitempty"`
	Limit    int          `json:"limit,omitempty"`
	Offset   int          `json:"offset,omitempty"`
}

// ValidateElementType checks if an element type is valid.
func ValidateElementType(t ElementType) bool {
	switch t {
	case PersonaElement, SkillElement, TemplateElement,
		AgentElement, MemoryElement, EnsembleElement:
		return true
	default:
		return false
	}
}

// GenerateElementID generates a unique ID for an element.
func GenerateElementID(elementType ElementType, name string) string {
	timestamp := time.Now().Format("20060102-150405")
	return string(elementType) + "_" + name + "_" + timestamp
}
