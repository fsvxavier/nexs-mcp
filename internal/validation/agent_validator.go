package validation

import (
	"errors"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// AgentValidator validates Agent elements.
type AgentValidator struct{}

func NewAgentValidator() *AgentValidator {
	return &AgentValidator{}
}

func (v *AgentValidator) SupportedType() domain.ElementType {
	return domain.AgentElement
}

func (v *AgentValidator) Validate(element domain.Element, level ValidationLevel) (*ValidationResult, error) {
	startTime := time.Now()

	agent, ok := element.(*domain.Agent)
	if !ok {
		return nil, errors.New("element is not an Agent type")
	}

	result := &ValidationResult{
		IsValid:     true,
		Errors:      []ValidationIssue{},
		Warnings:    []ValidationIssue{},
		Infos:       []ValidationIssue{},
		ElementType: string(domain.AgentElement),
		ElementID:   agent.GetID(),
	}

	v.validateBasic(agent, result)

	if level == ComprehensiveLevel || level == StrictLevel {
		v.validateComprehensive(agent, result)
	}

	result.ValidationTime = time.Since(startTime).Milliseconds()
	return result, nil
}

func (v *AgentValidator) validateBasic(agent *domain.Agent, result *ValidationResult) {
	if len(agent.Goals) == 0 {
		result.AddWarning("goals", "Agent should have at least one goal", "AGENT_NO_GOALS")
	}

	if len(agent.Actions) == 0 {
		result.AddWarning("actions", "Agent should have at least one action", "AGENT_NO_ACTIONS")
	}
}

func (v *AgentValidator) validateComprehensive(agent *domain.Agent, result *ValidationResult) {
	// Validate goals are specific and measurable
	for i, goal := range agent.Goals {
		if len(goal) < 10 {
			result.AddWarning(
				fmt.Sprintf("goals[%d]", i),
				"Goal should be more specific and detailed",
				"AGENT_VAGUE_GOAL",
			)
		}
	}

	// Validate actions are well-defined
	for i, action := range agent.Actions {
		if len(action.Name) < 3 {
			result.AddWarning(
				fmt.Sprintf("actions[%d].name", i),
				"Action name should be more descriptive",
				"AGENT_VAGUE_ACTION",
			)
		}

		// Validate action type
		validTypes := []string{"tool", "skill", "decision", "loop"}
		if issue := ValidateEnum(action.Type, fmt.Sprintf("actions[%d].type", i), validTypes); issue != nil {
			result.Errors = append(result.Errors, *issue)
			result.IsValid = false
		}
	}
}
