package validation

import (
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// SkillValidator validates Skill elements
type SkillValidator struct{}

// NewSkillValidator creates a new skill validator
func NewSkillValidator() *SkillValidator {
	return &SkillValidator{}
}

// SupportedType returns the element type this validator handles
func (v *SkillValidator) SupportedType() domain.ElementType {
	return domain.SkillElement
}

// Validate performs comprehensive validation on a Skill element
func (v *SkillValidator) Validate(element domain.Element, level ValidationLevel) (*ValidationResult, error) {
	startTime := time.Now()

	skill, ok := element.(*domain.Skill)
	if !ok {
		return nil, fmt.Errorf("element is not a Skill type")
	}

	result := &ValidationResult{
		IsValid:     true,
		Errors:      []ValidationIssue{},
		Warnings:    []ValidationIssue{},
		Infos:       []ValidationIssue{},
		ElementType: string(domain.SkillElement),
		ElementID:   skill.GetID(),
	}

	v.validateBasic(skill, result)

	if level == ComprehensiveLevel || level == StrictLevel {
		v.validateComprehensive(skill, result)
	}

	if level == StrictLevel {
		v.validateStrict(skill, result)
	}

	result.ValidationTime = time.Since(startTime).Milliseconds()
	return result, nil
}

func (v *SkillValidator) validateBasic(skill *domain.Skill, result *ValidationResult) {
	metadata := skill.GetMetadata()

	if issue := ValidateNotEmpty(metadata.Name, "name"); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	if len(skill.Triggers) == 0 {
		result.AddWarning("triggers", "Skill should have at least one trigger", "SKILL_NO_TRIGGERS")
	}

	if len(skill.Procedures) == 0 {
		result.AddError("procedures", "Skill must have at least one procedure", "SKILL_NO_PROCEDURES")
	}
}

func (v *SkillValidator) validateComprehensive(skill *domain.Skill, result *ValidationResult) {
	// Validate procedures
	for i, proc := range skill.Procedures {
		if proc.Action == "" {
			result.AddError(
				fmt.Sprintf("procedures[%d].action", i),
				"Procedure action is required",
				"SKILL_PROC_NO_ACTION",
			)
		}

		if len(proc.Action) < 5 {
			result.AddWarning(
				fmt.Sprintf("procedures[%d].action", i),
				"Procedure action should be more descriptive",
				"SKILL_PROC_BRIEF_ACTION",
			)
		}
	}

	// Validate triggers
	for i, trigger := range skill.Triggers {
		validTypes := []string{"keyword", "pattern", "context", "manual"}
		if issue := ValidateEnum(trigger.Type, fmt.Sprintf("triggers[%d].type", i), validTypes); issue != nil {
			result.Errors = append(result.Errors, *issue)
			result.IsValid = false
		}
	}
}

func (v *SkillValidator) validateStrict(skill *domain.Skill, result *ValidationResult) {
	metadata := skill.GetMetadata()

	if len(metadata.Tags) == 0 {
		result.AddWarning("tags", "Skill should have tags for discoverability", "SKILL_NO_TAGS")
	}

	if len(skill.Inputs) == 0 {
		result.AddInfo("inputs", "Consider documenting expected inputs", "SKILL_NO_INPUTS")
	}

	if len(skill.Outputs) == 0 {
		result.AddInfo("outputs", "Consider documenting expected outputs", "SKILL_NO_OUTPUTS")
	}
}
