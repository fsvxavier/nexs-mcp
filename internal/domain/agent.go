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
		Goals:         []string{},
		Actions:       []AgentAction{},
		DecisionTree:  make(map[string]interface{}),
		Context:       make(map[string]interface{}),
		MaxIterations: 10,
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
