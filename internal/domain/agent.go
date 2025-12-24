package domain

import (
	"errors"
	"fmt"
	"time"
)

// AgentAction defines an action an agent can perform.
type AgentAction struct {
	Name       string            `json:"name"                 validate:"required"                                yaml:"name"`
	Type       string            `json:"type"                 validate:"required,oneof=tool skill decision loop" yaml:"type"`
	Parameters map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	OnSuccess  string            `json:"on_success,omitempty" yaml:"on_success,omitempty"`
	OnFailure  string            `json:"on_failure,omitempty" yaml:"on_failure,omitempty"`
}

// Agent represents an autonomous task executor.
type Agent struct {
	metadata         ElementMetadata
	Goals            []string               `json:"goals"                       validate:"required,min=1"          yaml:"goals"`
	Actions          []AgentAction          `json:"actions"                     validate:"required,min=1,dive"     yaml:"actions"`
	DecisionTree     map[string]interface{} `json:"decision_tree,omitempty"     yaml:"decision_tree,omitempty"`
	FallbackStrategy string                 `json:"fallback_strategy,omitempty" yaml:"fallback_strategy,omitempty"`
	MaxIterations    int                    `json:"max_iterations"              validate:"min=1,max=100"           yaml:"max_iterations"`
	Context          map[string]interface{} `json:"context,omitempty"           yaml:"context,omitempty"`
	// Sprint 3: Cross-Element Relationships
	PersonaID        string   `json:"persona_id,omitempty"        yaml:"persona_id,omitempty"`        // Persona this agent uses
	RelatedSkills    []string `json:"related_skills,omitempty"    yaml:"related_skills,omitempty"`    // Skill IDs this agent uses
	RelatedTemplates []string `json:"related_templates,omitempty" yaml:"related_templates,omitempty"` // Template IDs this agent uses
	RelatedMemories  []string `json:"related_memories,omitempty"  yaml:"related_memories,omitempty"`  // Memory IDs associated with agent
}

// NewAgent creates a new Agent element.
func NewAgent(name, description, version, author string) *Agent {
	now := time.Now()
	return &Agent{
		metadata: ElementMetadata{
			ID:          GenerateElementID(AgentElement, name),
			Type:        AgentElement,
			Name:        name,
			Description: description,
			Version:     version,
			Author:      author,
			Tags:        []string{},
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Goals:            []string{},
		Actions:          []AgentAction{},
		DecisionTree:     make(map[string]interface{}),
		Context:          make(map[string]interface{}),
		MaxIterations:    10,
		RelatedSkills:    []string{},
		RelatedTemplates: []string{},
		RelatedMemories:  []string{},
	}
}

func (a *Agent) GetMetadata() ElementMetadata { return a.metadata }
func (a *Agent) GetType() ElementType         { return a.metadata.Type }
func (a *Agent) GetID() string                { return a.metadata.ID }
func (a *Agent) IsActive() bool               { return a.metadata.IsActive }

func (a *Agent) Activate() error {
	a.metadata.IsActive = true
	a.metadata.UpdatedAt = time.Now()
	return nil
}

func (a *Agent) Deactivate() error {
	a.metadata.IsActive = false
	a.metadata.UpdatedAt = time.Now()
	return nil
}

func (a *Agent) Validate() error {
	if err := a.metadata.Validate(); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}
	if len(a.Goals) == 0 {
		return errors.New("at least one goal is required")
	}
	if len(a.Actions) == 0 {
		return errors.New("at least one action is required")
	}
	if a.MaxIterations < 1 || a.MaxIterations > 100 {
		return errors.New("max_iterations must be between 1 and 100")
	}
	return nil
}

func (a *Agent) SetMetadata(metadata ElementMetadata) {
	a.metadata = metadata
	a.metadata.UpdatedAt = time.Now()
}

// AddRelatedSkill adds a skill ID to the agent's related skills.
func (a *Agent) AddRelatedSkill(skillID string) {
	if skillID == "" {
		return
	}
	if !containsString(a.RelatedSkills, skillID) {
		a.RelatedSkills = append(a.RelatedSkills, skillID)
		a.metadata.UpdatedAt = time.Now()
	}
}

// RemoveRelatedSkill removes a skill ID from the agent's related skills.
func (a *Agent) RemoveRelatedSkill(skillID string) {
	a.RelatedSkills = removeStringFromSlice(a.RelatedSkills, skillID)
	a.metadata.UpdatedAt = time.Now()
}

// AddRelatedTemplate adds a template ID to the agent's related templates.
func (a *Agent) AddRelatedTemplate(templateID string) {
	if templateID == "" {
		return
	}
	if !containsString(a.RelatedTemplates, templateID) {
		a.RelatedTemplates = append(a.RelatedTemplates, templateID)
		a.metadata.UpdatedAt = time.Now()
	}
}

// RemoveRelatedTemplate removes a template ID from the agent's related templates.
func (a *Agent) RemoveRelatedTemplate(templateID string) {
	a.RelatedTemplates = removeStringFromSlice(a.RelatedTemplates, templateID)
	a.metadata.UpdatedAt = time.Now()
}

// AddRelatedMemory adds a memory ID to the agent's related memories.
func (a *Agent) AddRelatedMemory(memoryID string) {
	if memoryID == "" {
		return
	}
	if !containsString(a.RelatedMemories, memoryID) {
		a.RelatedMemories = append(a.RelatedMemories, memoryID)
		a.metadata.UpdatedAt = time.Now()
	}
}

// RemoveRelatedMemory removes a memory ID from the agent's related memories.
func (a *Agent) RemoveRelatedMemory(memoryID string) {
	a.RelatedMemories = removeStringFromSlice(a.RelatedMemories, memoryID)
	a.metadata.UpdatedAt = time.Now()
}

// SetPersona sets the persona ID for this agent.
func (a *Agent) SetPersona(personaID string) {
	a.PersonaID = personaID
	a.metadata.UpdatedAt = time.Now()
}

// GetAllRelatedIDs returns all related element IDs (skills, templates, memories, persona).
func (a *Agent) GetAllRelatedIDs() []string {
	allIDs := make([]string, 0)

	if a.PersonaID != "" {
		allIDs = append(allIDs, a.PersonaID)
	}

	allIDs = append(allIDs, a.RelatedSkills...)
	allIDs = append(allIDs, a.RelatedTemplates...)
	allIDs = append(allIDs, a.RelatedMemories...)

	return allIDs
}
