package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// NewGetElementTool creates the get_element tool
func NewGetElementTool(repo domain.ElementRepository) *Tool {
	return &Tool{
		Name:        "get_element",
		Description: "Get a specific element by ID",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The element ID",
				},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var input struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			if input.ID == "" {
				return nil, fmt.Errorf("id is required")
			}

			element, err := repo.GetByID(input.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get element: %w", err)
			}

			return map[string]interface{}{
				"element": element.GetMetadata(),
			}, nil
		},
	}
}

// NewCreateElementTool creates the create_element tool
func NewCreateElementTool(repo domain.ElementRepository) *Tool {
	return &Tool{
		Name:        "create_element",
		Description: "Create a new element",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"required": []string{"type", "name", "version", "author"},
			"properties": map[string]interface{}{
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Element type",
					"enum":        []string{"persona", "skill", "template", "agent", "memory", "ensemble"},
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Element name",
					"minLength":   3,
					"maxLength":   100,
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Element description",
					"maxLength":   500,
				},
				"version": map[string]interface{}{
					"type":        "string",
					"description": "Element version (semver)",
				},
				"author": map[string]interface{}{
					"type":        "string",
					"description": "Element author",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "Element tags",
				},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var input struct {
				Type        string   `json:"type"`
				Name        string   `json:"name"`
				Description string   `json:"description"`
				Version     string   `json:"version"`
				Author      string   `json:"author"`
				Tags        []string `json:"tags"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			elemType := domain.ElementType(input.Type)
			if !domain.ValidateElementType(elemType) {
				return nil, fmt.Errorf("invalid element type: %s", input.Type)
			}

			id := domain.GenerateElementID(elemType, input.Name)
			now := time.Now()

			metadata := domain.ElementMetadata{
				ID:          id,
				Type:        elemType,
				Name:        input.Name,
				Description: input.Description,
				Version:     input.Version,
				Author:      input.Author,
				Tags:        input.Tags,
				IsActive:    true,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			element := &SimpleElement{metadata: metadata}

			if err := repo.Create(element); err != nil {
				return nil, fmt.Errorf("failed to create element: %w", err)
			}

			return map[string]interface{}{
				"id":      id,
				"element": metadata,
			}, nil
		},
	}
}

// NewUpdateElementTool creates the update_element tool
func NewUpdateElementTool(repo domain.ElementRepository) *Tool {
	return &Tool{
		Name:        "update_element",
		Description: "Update an existing element",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The element ID",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Element name",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Element description",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "Element tags",
				},
				"is_active": map[string]interface{}{
					"type":        "boolean",
					"description": "Active status",
				},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var input struct {
				ID          string   `json:"id"`
				Name        string   `json:"name,omitempty"`
				Description string   `json:"description,omitempty"`
				Tags        []string `json:"tags,omitempty"`
				IsActive    *bool    `json:"is_active,omitempty"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			if input.ID == "" {
				return nil, fmt.Errorf("id is required")
			}

			element, err := repo.GetByID(input.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get element: %w", err)
			}

			metadata := element.GetMetadata()

			if input.Name != "" {
				metadata.Name = input.Name
			}
			if input.Description != "" {
				metadata.Description = input.Description
			}
			if len(input.Tags) > 0 {
				metadata.Tags = input.Tags
			}
			if input.IsActive != nil {
				metadata.IsActive = *input.IsActive
			}
			metadata.UpdatedAt = time.Now()

			updated := &SimpleElement{metadata: metadata}

			if err := repo.Update(updated); err != nil {
				return nil, fmt.Errorf("failed to update element: %w", err)
			}

			return map[string]interface{}{
				"id":      input.ID,
				"element": metadata,
			}, nil
		},
	}
}

// NewDeleteElementTool creates the delete_element tool
func NewDeleteElementTool(repo domain.ElementRepository) *Tool {
	return &Tool{
		Name:        "delete_element",
		Description: "Delete an element by ID",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The element ID to delete",
				},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var input struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			if input.ID == "" {
				return nil, fmt.Errorf("id is required")
			}

			if err := repo.Delete(input.ID); err != nil {
				return nil, fmt.Errorf("failed to delete element: %w", err)
			}

			return map[string]interface{}{
				"id":      input.ID,
				"deleted": true,
			}, nil
		},
	}
}

// SimpleElement is a basic implementation of Element for CRUD operations
type SimpleElement struct {
	metadata domain.ElementMetadata
}

func (s *SimpleElement) GetMetadata() domain.ElementMetadata { return s.metadata }
func (s *SimpleElement) Validate() error                     { return nil }
func (s *SimpleElement) GetType() domain.ElementType         { return s.metadata.Type }
func (s *SimpleElement) GetID() string                       { return s.metadata.ID }
func (s *SimpleElement) IsActive() bool                      { return s.metadata.IsActive }
func (s *SimpleElement) Activate() error {
	s.metadata.IsActive = true
	return nil
}
func (s *SimpleElement) Deactivate() error {
	s.metadata.IsActive = false
	return nil
}
