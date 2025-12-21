package validation

import (
	"errors"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// TemplateValidator validates Template elements.
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
		return nil, errors.New("element is not a Template type")
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
