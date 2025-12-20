package domain

import (
	"fmt"
	"strings"
	"time"
)

// TemplateVariable defines a variable in a template
type TemplateVariable struct {
	Name        string `json:"name" yaml:"name" validate:"required"`
	Type        string `json:"type" yaml:"type" validate:"required,oneof=string number boolean array object"`
	Required    bool   `json:"required" yaml:"required"`
	Default     string `json:"default,omitempty" yaml:"default,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Template represents a reusable content template
type Template struct {
	metadata        ElementMetadata
	Content         string             `json:"content" yaml:"content" validate:"required"`
	Format          string             `json:"format" yaml:"format" validate:"required,oneof=markdown yaml json text"`
	Variables       []TemplateVariable `json:"variables" yaml:"variables" validate:"dive"`
	ValidationRules map[string]string  `json:"validation_rules,omitempty" yaml:"validation_rules,omitempty"`
}

// NewTemplate creates a new Template element
func NewTemplate(name, description, version, author string) *Template {
	now := time.Now()
	return &Template{
		metadata: ElementMetadata{
			ID:          GenerateElementID(TemplateElement, name),
			Type:        TemplateElement,
			Name:        name,
			Description: description,
			Version:     version,
			Author:      author,
			Tags:        []string{},
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Variables:       []TemplateVariable{},
		ValidationRules: make(map[string]string),
		Format:          "markdown",
	}
}

func (t *Template) GetMetadata() ElementMetadata { return t.metadata }
func (t *Template) GetType() ElementType         { return t.metadata.Type }
func (t *Template) GetID() string                { return t.metadata.ID }
func (t *Template) IsActive() bool               { return t.metadata.IsActive }

func (t *Template) Activate() error {
	t.metadata.IsActive = true
	t.metadata.UpdatedAt = time.Now()
	return nil
}

func (t *Template) Deactivate() error {
	t.metadata.IsActive = false
	t.metadata.UpdatedAt = time.Now()
	return nil
}

func (t *Template) Validate() error {
	if err := t.metadata.Validate(); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}
	if t.Content == "" {
		return fmt.Errorf("content is required")
	}
	validFormats := map[string]bool{"markdown": true, "yaml": true, "json": true, "text": true}
	if !validFormats[t.Format] {
		return fmt.Errorf("invalid format: %s", t.Format)
	}
	return nil
}

func (t *Template) SetMetadata(metadata ElementMetadata) {
	t.metadata = metadata
	t.metadata.UpdatedAt = time.Now()
}

// Render replaces variables in the template with provided values
func (t *Template) Render(values map[string]string) (string, error) {
	result := t.Content
	for _, v := range t.Variables {
		val, ok := values[v.Name]
		if !ok {
			if v.Required {
				return "", fmt.Errorf("required variable %s not provided", v.Name)
			}
			val = v.Default
		}
		result = strings.ReplaceAll(result, "{{"+v.Name+"}}", val)
	}
	return result, nil
}
