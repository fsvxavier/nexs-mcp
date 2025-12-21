package domain

import (
	"errors"
	"fmt"
	"time"
)

// SkillTrigger defines when a skill should be activated.
type SkillTrigger struct {
	Type     string   `json:"type"               validate:"required,oneof=keyword pattern context manual" yaml:"type"`
	Keywords []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`
	Pattern  string   `json:"pattern,omitempty"  yaml:"pattern,omitempty"`
	Context  string   `json:"context,omitempty"  yaml:"context,omitempty"`
}

// SkillProcedure defines a step in the skill execution.
type SkillProcedure struct {
	Step        int      `json:"step"                  validate:"required,min=1"    yaml:"step"`
	Action      string   `json:"action"                validate:"required"          yaml:"action"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	ToolsUsed   []string `json:"tools_used,omitempty"  yaml:"tools_used,omitempty"`
	Validation  string   `json:"validation,omitempty"  yaml:"validation,omitempty"`
}

// SkillDependency defines a dependency on another skill.
type SkillDependency struct {
	SkillID  string `json:"skill_id"          validate:"required"      yaml:"skill_id"`
	Required bool   `json:"required"          yaml:"required"`
	Version  string `json:"version,omitempty" yaml:"version,omitempty"`
}

// Skill represents a specialized capability.
type Skill struct {
	metadata      ElementMetadata
	Triggers      []SkillTrigger    `json:"triggers"                 validate:"required,min=1,dive"  yaml:"triggers"`
	Procedures    []SkillProcedure  `json:"procedures"               validate:"required,min=1,dive"  yaml:"procedures"`
	Dependencies  []SkillDependency `json:"dependencies,omitempty"   yaml:"dependencies,omitempty"`
	ToolsRequired []string          `json:"tools_required,omitempty" yaml:"tools_required,omitempty"`
	Inputs        map[string]string `json:"inputs,omitempty"         yaml:"inputs,omitempty"`
	Outputs       map[string]string `json:"outputs,omitempty"        yaml:"outputs,omitempty"`
	Composable    bool              `json:"composable"               yaml:"composable"`
}

// NewSkill creates a new Skill element.
func NewSkill(name, description, version, author string) *Skill {
	now := time.Now()
	return &Skill{
		metadata: ElementMetadata{
			ID:          GenerateElementID(SkillElement, name),
			Type:        SkillElement,
			Name:        name,
			Description: description,
			Version:     version,
			Author:      author,
			Tags:        []string{},
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Triggers:      []SkillTrigger{},
		Procedures:    []SkillProcedure{},
		Dependencies:  []SkillDependency{},
		ToolsRequired: []string{},
		Inputs:        make(map[string]string),
		Outputs:       make(map[string]string),
		Composable:    true,
	}
}

// GetMetadata returns the element metadata.
func (s *Skill) GetMetadata() ElementMetadata {
	return s.metadata
}

// GetType returns the element type.
func (s *Skill) GetType() ElementType {
	return s.metadata.Type
}

// GetID returns the element ID.
func (s *Skill) GetID() string {
	return s.metadata.ID
}

// IsActive returns whether the element is active.
func (s *Skill) IsActive() bool {
	return s.metadata.IsActive
}

// Activate activates the skill.
func (s *Skill) Activate() error {
	s.metadata.IsActive = true
	s.metadata.UpdatedAt = time.Now()
	return nil
}

// Deactivate deactivates the skill.
func (s *Skill) Deactivate() error {
	s.metadata.IsActive = false
	s.metadata.UpdatedAt = time.Now()
	return nil
}

// Validate validates the skill structure.
func (s *Skill) Validate() error {
	if err := s.metadata.Validate(); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	if len(s.Triggers) == 0 {
		return errors.New("at least one trigger is required")
	}

	for i, trigger := range s.Triggers {
		validTypes := map[string]bool{"keyword": true, "pattern": true, "context": true, "manual": true}
		if !validTypes[trigger.Type] {
			return fmt.Errorf("trigger %d: invalid type %s", i, trigger.Type)
		}
		if trigger.Type == "keyword" && len(trigger.Keywords) == 0 {
			return fmt.Errorf("trigger %d: keywords required for keyword type", i)
		}
		if trigger.Type == "pattern" && trigger.Pattern == "" {
			return fmt.Errorf("trigger %d: pattern required for pattern type", i)
		}
	}

	if len(s.Procedures) == 0 {
		return errors.New("at least one procedure is required")
	}

	for i, proc := range s.Procedures {
		if proc.Step < 1 {
			return fmt.Errorf("procedure %d: step must be >= 1", i)
		}
		if proc.Action == "" {
			return fmt.Errorf("procedure %d: action is required", i)
		}
	}

	return nil
}

// SetMetadata updates the skill metadata.
func (s *Skill) SetMetadata(metadata ElementMetadata) {
	s.metadata = metadata
	s.metadata.UpdatedAt = time.Now()
}

// AddTrigger adds a trigger to the skill.
func (s *Skill) AddTrigger(trigger SkillTrigger) error {
	validTypes := map[string]bool{"keyword": true, "pattern": true, "context": true, "manual": true}
	if !validTypes[trigger.Type] {
		return fmt.Errorf("invalid trigger type: %s", trigger.Type)
	}
	s.Triggers = append(s.Triggers, trigger)
	s.metadata.UpdatedAt = time.Now()
	return nil
}

// AddProcedure adds a procedure step to the skill.
func (s *Skill) AddProcedure(procedure SkillProcedure) error {
	if procedure.Action == "" {
		return errors.New("action is required")
	}
	s.Procedures = append(s.Procedures, procedure)
	s.metadata.UpdatedAt = time.Now()
	return nil
}

// AddDependency adds a dependency to another skill.
func (s *Skill) AddDependency(dep SkillDependency) error {
	if dep.SkillID == "" {
		return errors.New("skill_id is required")
	}
	s.Dependencies = append(s.Dependencies, dep)
	s.metadata.UpdatedAt = time.Now()
	return nil
}
