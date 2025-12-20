package domain

import (
	"fmt"
	"time"
)

// EnsembleMember represents a member agent in the ensemble
type EnsembleMember struct {
	AgentID  string `json:"agent_id" yaml:"agent_id" validate:"required"`
	Role     string `json:"role" yaml:"role" validate:"required"`
	Priority int    `json:"priority" yaml:"priority" validate:"min=1,max=10"`
}

// Ensemble represents multi-agent orchestration
type Ensemble struct {
	metadata            ElementMetadata
	Members             []EnsembleMember       `json:"members" yaml:"members" validate:"required,min=1,dive"`
	ExecutionMode       string                 `json:"execution_mode" yaml:"execution_mode" validate:"required,oneof=sequential parallel hybrid"`
	AggregationStrategy string                 `json:"aggregation_strategy" yaml:"aggregation_strategy" validate:"required"`
	FallbackChain       []string               `json:"fallback_chain,omitempty" yaml:"fallback_chain,omitempty"`
	SharedContext       map[string]interface{} `json:"shared_context,omitempty" yaml:"shared_context,omitempty"`
}

// NewEnsemble creates a new Ensemble element
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
		return fmt.Errorf("at least one member is required")
	}
	validModes := map[string]bool{"sequential": true, "parallel": true, "hybrid": true}
	if !validModes[e.ExecutionMode] {
		return fmt.Errorf("invalid execution_mode: %s", e.ExecutionMode)
	}
	if e.AggregationStrategy == "" {
		return fmt.Errorf("aggregation_strategy is required")
	}
	return nil
}

func (e *Ensemble) SetMetadata(metadata ElementMetadata) {
	e.metadata = metadata
	e.metadata.UpdatedAt = time.Now()
}
