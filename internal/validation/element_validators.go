package validation

import (
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// TemplateValidator validates Template elements
type TemplateValidator struct{}

func NewTemplateValidator() *TemplateValidator {
	return &TemplateValidator{}
}

func (v *TemplateValidator) SupportedType() domain.ElementType {
	return domain.TemplateElement
}

func (v *TemplateValidator) Validate(element domain.Element, level ValidationLevel) (*ValidationResult, error) {
	startTime := time.Now()

	template, ok := element.(*domain.Template)
	if !ok {
		return nil, fmt.Errorf("element is not a Template type")
	}

	result := &ValidationResult{
		IsValid:     true,
		Errors:      []ValidationIssue{},
		Warnings:    []ValidationIssue{},
		Infos:       []ValidationIssue{},
		ElementType: string(domain.TemplateElement),
		ElementID:   template.GetID(),
	}

	v.validateBasic(template, result)

	if level == ComprehensiveLevel || level == StrictLevel {
		v.validateComprehensive(template, result)
	}

	result.ValidationTime = time.Since(startTime).Milliseconds()
	return result, nil
}

func (v *TemplateValidator) validateBasic(template *domain.Template, result *ValidationResult) {
	if issue := ValidateNotEmpty(template.Content, "content"); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	if issue := ValidateEnum(template.Format, "format", []string{"markdown", "yaml", "json", "text"}); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}
}

func (v *TemplateValidator) validateComprehensive(template *domain.Template, result *ValidationResult) {
	// Validate variables are well-defined
	for i, variable := range template.Variables {
		if variable.Name == "" {
			result.AddError(
				fmt.Sprintf("variables[%d].name", i),
				"Variable must have a name",
				"TEMPLATE_VAR_NO_NAME",
			)
		}

		if variable.Type == "" {
			result.AddError(
				fmt.Sprintf("variables[%d].type", i),
				"Variable must have a type",
				"TEMPLATE_VAR_NO_TYPE",
			)
		}

		if variable.Required && variable.Default != "" {
			result.AddWarning(
				fmt.Sprintf("variables[%d]", i),
				"Required variable has a default value - this may be confusing",
				"TEMPLATE_VAR_REQ_WITH_DEFAULT",
			)
		}
	}
}

// AgentValidator validates Agent elements
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
		return nil, fmt.Errorf("element is not an Agent type")
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

// MemoryValidator validates Memory elements
type MemoryValidator struct{}

func NewMemoryValidator() *MemoryValidator {
	return &MemoryValidator{}
}

func (v *MemoryValidator) SupportedType() domain.ElementType {
	return domain.MemoryElement
}

func (v *MemoryValidator) Validate(element domain.Element, level ValidationLevel) (*ValidationResult, error) {
	startTime := time.Now()

	memory, ok := element.(*domain.Memory)
	if !ok {
		return nil, fmt.Errorf("element is not a Memory type")
	}

	result := &ValidationResult{
		IsValid:     true,
		Errors:      []ValidationIssue{},
		Warnings:    []ValidationIssue{},
		Infos:       []ValidationIssue{},
		ElementType: string(domain.MemoryElement),
		ElementID:   memory.GetID(),
	}

	v.validateBasic(memory, result)

	result.ValidationTime = time.Since(startTime).Milliseconds()
	return result, nil
}

func (v *MemoryValidator) validateBasic(memory *domain.Memory, result *ValidationResult) {
	// Validate content hash exists
	if memory.ContentHash == "" {
		result.AddInfo("content_hash", "Consider computing content hash for deduplication", "MEMORY_NO_HASH")
	}

	// Validate date format
	if memory.DateCreated == "" {
		result.AddWarning("date_created", "DateCreated should be set", "MEMORY_NO_DATE")
	}

	// Validate search index
	if len(memory.SearchIndex) == 0 {
		result.AddInfo("search_index", "Consider adding search index entries for better discoverability", "MEMORY_NO_INDEX")
	}
}

// EnsembleValidator validates Ensemble elements
type EnsembleValidator struct{}

func NewEnsembleValidator() *EnsembleValidator {
	return &EnsembleValidator{}
}

func (v *EnsembleValidator) SupportedType() domain.ElementType {
	return domain.EnsembleElement
}

func (v *EnsembleValidator) Validate(element domain.Element, level ValidationLevel) (*ValidationResult, error) {
	startTime := time.Now()

	ensemble, ok := element.(*domain.Ensemble)
	if !ok {
		return nil, fmt.Errorf("element is not an Ensemble type")
	}

	result := &ValidationResult{
		IsValid:     true,
		Errors:      []ValidationIssue{},
		Warnings:    []ValidationIssue{},
		Infos:       []ValidationIssue{},
		ElementType: string(domain.EnsembleElement),
		ElementID:   ensemble.GetID(),
	}

	v.validateBasic(ensemble, result)

	if level == ComprehensiveLevel || level == StrictLevel {
		v.validateComprehensive(ensemble, result)
	}

	result.ValidationTime = time.Since(startTime).Milliseconds()
	return result, nil
}

func (v *EnsembleValidator) validateBasic(ensemble *domain.Ensemble, result *ValidationResult) {
	if len(ensemble.Members) < 2 {
		result.AddWarning("members", "Ensemble should have at least 2 members for effective orchestration", "ENSEMBLE_FEW_MEMBERS")
	}

	if issue := ValidateEnum(ensemble.ExecutionMode, "execution_mode", []string{"sequential", "parallel", "hybrid"}); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	if ensemble.AggregationStrategy == "" {
		result.AddError("aggregation_strategy", "Aggregation strategy is required", "ENSEMBLE_NO_AGGREGATION")
	}
}

func (v *EnsembleValidator) validateComprehensive(ensemble *domain.Ensemble, result *ValidationResult) {
	// Validate member roles are unique and meaningful
	roles := make(map[string]bool)
	for i, member := range ensemble.Members {
		if member.Role == "" {
			result.AddError(
				fmt.Sprintf("members[%d].role", i),
				"Member role is required",
				"ENSEMBLE_MEMBER_NO_ROLE",
			)
		}

		if roles[member.Role] {
			result.AddWarning(
				fmt.Sprintf("members[%d].role", i),
				"Duplicate role detected: "+member.Role,
				"ENSEMBLE_DUPLICATE_ROLE",
			)
		}
		roles[member.Role] = true

		// Validate priority
		if member.Priority < 1 || member.Priority > 10 {
			result.AddError(
				fmt.Sprintf("members[%d].priority", i),
				"Priority must be between 1 and 10",
				"ENSEMBLE_INVALID_PRIORITY",
			)
		}
	}

	// Check for cyclic dependencies in sequential orchestration
	if ensemble.ExecutionMode == "sequential" && len(ensemble.Members) > 1 {
		result.AddInfo(
			"execution_mode",
			"Ensure agents in sequential ensemble don't have circular dependencies",
			"ENSEMBLE_CHECK_CYCLES",
		)
	}
}
