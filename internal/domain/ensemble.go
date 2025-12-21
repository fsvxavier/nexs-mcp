package domain

import (
	"errors"
	"fmt"
	"time"
)

// EnsembleMember represents a member agent in the ensemble.
type EnsembleMember struct {
	AgentID  string `json:"agent_id" validate:"required"     yaml:"agent_id"`
	Role     string `json:"role"     validate:"required"     yaml:"role"`
	Priority int    `json:"priority" validate:"min=1,max=10" yaml:"priority"`
}

// Ensemble represents multi-agent orchestration.
type Ensemble struct {
	metadata            ElementMetadata
	Members             []EnsembleMember       `json:"members"                  validate:"required,min=1,dive"                       yaml:"members"`
	ExecutionMode       string                 `json:"execution_mode"           validate:"required,oneof=sequential parallel hybrid" yaml:"execution_mode"`
	AggregationStrategy string                 `json:"aggregation_strategy"     validate:"required"                                  yaml:"aggregation_strategy"`
	FallbackChain       []string               `json:"fallback_chain,omitempty" yaml:"fallback_chain,omitempty"`
	SharedContext       map[string]interface{} `json:"shared_context,omitempty" yaml:"shared_context,omitempty"`
}

// NewEnsemble creates a new Ensemble element.
func NewEnsemble(name, description, version, author string) *Ensemble {
	now := time.Now()
	return &Ensemble{
		metadata: ElementMetadata{
			ID:          GenerateElementID(EnsembleElement, name),
			Type:        EnsembleElement,
			Name:        name,
			Description: description,
			Version:     version,
			Author:      author,
			Tags:        []string{},
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Members:       []EnsembleMember{},
		ExecutionMode: "sequential",
		SharedContext: make(map[string]interface{}),
		FallbackChain: []string{},
	}
}

func (e *Ensemble) GetMetadata() ElementMetadata { return e.metadata }
func (e *Ensemble) GetType() ElementType         { return e.metadata.Type }
func (e *Ensemble) GetID() string                { return e.metadata.ID }
func (e *Ensemble) IsActive() bool               { return e.metadata.IsActive }

func (e *Ensemble) Activate() error {
	e.metadata.IsActive = true
	e.metadata.UpdatedAt = time.Now()
	return nil
}

func (e *Ensemble) Deactivate() error {
	e.metadata.IsActive = false
	e.metadata.UpdatedAt = time.Now()
	return nil
}

func (e *Ensemble) Validate() error {
	if err := e.metadata.Validate(); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}
	if len(e.Members) == 0 {
		return errors.New("at least one member is required")
	}
	validModes := map[string]bool{"sequential": true, "parallel": true, "hybrid": true}
	if !validModes[e.ExecutionMode] {
		return fmt.Errorf("invalid execution_mode: %s", e.ExecutionMode)
	}
	if e.AggregationStrategy == "" {
		return errors.New("aggregation_strategy is required")
	}
	return nil
}

func (e *Ensemble) SetMetadata(metadata ElementMetadata) {
	e.metadata = metadata
	e.metadata.UpdatedAt = time.Now()
}
