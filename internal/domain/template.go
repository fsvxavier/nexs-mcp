package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// TemplateVariable defines a variable in a template.
type TemplateVariable struct {
	Name        string `json:"name"                  validate:"required"                                          yaml:"name"`
	Type        string `json:"type"                  validate:"required,oneof=string number boolean array object" yaml:"type"`
	Required    bool   `json:"required"              yaml:"required"`
	Default     string `json:"default,omitempty"     yaml:"default,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Template represents a reusable content template.
type Template struct {
	metadata        ElementMetadata
	Content         string             `json:"content"                    validate:"required"                               yaml:"content"`
	Format          string             `json:"format"                     validate:"required,oneof=markdown yaml json text" yaml:"format"`
	Variables       []TemplateVariable `json:"variables"                  validate:"dive"                                   yaml:"variables"`
	ValidationRules map[string]string  `json:"validation_rules,omitempty" yaml:"validation_rules,omitempty"`
	// Sprint 3: Cross-Element Relationships
	RelatedSkills   []string `json:"related_skills,omitempty"   yaml:"related_skills,omitempty"`   // Skill IDs this template requires
	RelatedMemories []string `json:"related_memories,omitempty" yaml:"related_memories,omitempty"` // Memory IDs associated with template
}

// NewTemplate creates a new Template element.
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
		RelatedSkills:   []string{},
		RelatedMemories: []string{},
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
		return errors.New("content is required")
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

// Render replaces variables in the template with provided values.
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

// AddRelatedSkill adds a skill ID to the template's related skills.
func (t *Template) AddRelatedSkill(skillID string) {
	if skillID == "" {
		return
	}
	if !containsString(t.RelatedSkills, skillID) {
		t.RelatedSkills = append(t.RelatedSkills, skillID)
		t.metadata.UpdatedAt = time.Now()
	}
}

// RemoveRelatedSkill removes a skill ID from the template's related skills.
func (t *Template) RemoveRelatedSkill(skillID string) {
	t.RelatedSkills = removeStringFromSlice(t.RelatedSkills, skillID)
	t.metadata.UpdatedAt = time.Now()
}

// GetAllRelatedIDs returns all related element IDs (skills).
func (t *Template) GetAllRelatedIDs() []string {
	return t.RelatedSkills
}
