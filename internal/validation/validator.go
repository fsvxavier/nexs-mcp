package validation

import (
	"fmt"
	"strings"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// ValidationLevel defines the depth of validation.
type ValidationLevel string

const (
	BasicLevel         ValidationLevel = "basic"
	ComprehensiveLevel ValidationLevel = "comprehensive"
	StrictLevel        ValidationLevel = "strict"
)

// ValidationSeverity indicates the importance of a validation issue.
type ValidationSeverity string

const (
	ErrorSeverity   ValidationSeverity = "error"
	WarningSeverity ValidationSeverity = "warning"
	InfoSeverity    ValidationSeverity = "info"
)

// ValidationIssue represents a single validation problem.
type ValidationIssue struct {
	Severity   ValidationSeverity `json:"severity"`
	Field      string             `json:"field"`
	Message    string             `json:"message"`
	Line       int                `json:"line,omitempty"`
	Suggestion string             `json:"suggestion,omitempty"`
	Code       string             `json:"code"` // e.g., "PERSONA_TONE_INCONSISTENT"
}

// ValidationResult contains the outcome of element validation.
type ValidationResult struct {
	IsValid        bool              `json:"is_valid"`
	Errors         []ValidationIssue `json:"errors"`
	Warnings       []ValidationIssue `json:"warnings"`
	Infos          []ValidationIssue `json:"infos"`
	ValidationTime int64             `json:"validation_time_ms"`
	ElementType    string            `json:"element_type"`
	ElementID      string            `json:"element_id"`
}

// AddError adds an error-level validation issue.
func (vr *ValidationResult) AddError(field, message, code string) {
	vr.IsValid = false
	vr.Errors = append(vr.Errors, ValidationIssue{
		Severity: ErrorSeverity,
		Field:    field,
		Message:  message,
		Code:     code,
	})
}

// AddErrorWithSuggestion adds an error with a fix suggestion.
func (vr *ValidationResult) AddErrorWithSuggestion(field, message, code, suggestion string) {
	vr.IsValid = false
	vr.Errors = append(vr.Errors, ValidationIssue{
		Severity:   ErrorSeverity,
		Field:      field,
		Message:    message,
		Code:       code,
		Suggestion: suggestion,
	})
}

// AddWarning adds a warning-level validation issue.
func (vr *ValidationResult) AddWarning(field, message, code string) {
	vr.Warnings = append(vr.Warnings, ValidationIssue{
		Severity: WarningSeverity,
		Field:    field,
		Message:  message,
		Code:     code,
	})
}

// AddInfo adds an info-level validation issue.
func (vr *ValidationResult) AddInfo(field, message, code string) {
	vr.Infos = append(vr.Infos, ValidationIssue{
		Severity: InfoSeverity,
		Field:    field,
		Message:  message,
		Code:     code,
	})
}

// ErrorCount returns the total number of errors.
func (vr *ValidationResult) ErrorCount() int {
	return len(vr.Errors)
}

// WarningCount returns the total number of warnings.
func (vr *ValidationResult) WarningCount() int {
	return len(vr.Warnings)
}

// TotalIssues returns the total number of all issues.
func (vr *ValidationResult) TotalIssues() int {
	return len(vr.Errors) + len(vr.Warnings) + len(vr.Infos)
}

// ElementValidator interface for type-specific validation.
type ElementValidator interface {
	// Validate performs validation at the specified level
	Validate(element domain.Element, level ValidationLevel) (*ValidationResult, error)

	// SupportedType returns the element type this validator handles
	SupportedType() domain.ElementType
}

// ValidatorRegistry manages type-specific validators.
type ValidatorRegistry struct {
	validators map[domain.ElementType]ElementValidator
}

// NewValidatorRegistry creates a new validator registry.
func NewValidatorRegistry() *ValidatorRegistry {
	registry := &ValidatorRegistry{
		validators: make(map[domain.ElementType]ElementValidator),
	}

	// Register all type-specific validators
	registry.Register(NewPersonaValidator())
	registry.Register(NewSkillValidator())
	registry.Register(NewTemplateValidator())
	registry.Register(NewAgentValidator())
	registry.Register(NewMemoryValidator())
	registry.Register(NewEnsembleValidator())

	return registry
}

// Register adds a validator to the registry.
func (r *ValidatorRegistry) Register(validator ElementValidator) {
	r.validators[validator.SupportedType()] = validator
}

// GetValidator retrieves a validator for the given element type.
func (r *ValidatorRegistry) GetValidator(elementType domain.ElementType) (ElementValidator, error) {
	validator, exists := r.validators[elementType]
	if !exists {
		return nil, fmt.Errorf("no validator registered for element type: %s", elementType)
	}
	return validator, nil
}

// ValidateElement performs validation using the appropriate type-specific validator.
func (r *ValidatorRegistry) ValidateElement(element domain.Element, level ValidationLevel) (*ValidationResult, error) {
	validator, err := r.GetValidator(element.GetType())
	if err != nil {
		return nil, err
	}

	return validator.Validate(element, level)
}

// Helper functions for common validation patterns

// ValidateNotEmpty checks if a string field is not empty.
func ValidateNotEmpty(value, fieldName string) *ValidationIssue {
	if strings.TrimSpace(value) == "" {
		return &ValidationIssue{
			Severity:   ErrorSeverity,
			Field:      fieldName,
			Message:    fieldName + " cannot be empty",
			Code:       strings.ToUpper(fieldName) + "_EMPTY",
			Suggestion: "Provide a meaningful " + fieldName,
		}
	}
	return nil
}

// ValidateLength checks if a string meets minimum and maximum length requirements.
func ValidateLength(value, fieldName string, min, max int) *ValidationIssue {
	length := len(strings.TrimSpace(value))
	if length < min {
		return &ValidationIssue{
			Severity:   ErrorSeverity,
			Field:      fieldName,
			Message:    fmt.Sprintf("%s is too short (minimum %d characters)", fieldName, min),
			Code:       strings.ToUpper(fieldName) + "_TOO_SHORT",
			Suggestion: fmt.Sprintf("Expand %s to at least %d characters", fieldName, min),
		}
	}
	if max > 0 && length > max {
		return &ValidationIssue{
			Severity:   ErrorSeverity,
			Field:      fieldName,
			Message:    fmt.Sprintf("%s is too long (maximum %d characters)", fieldName, max),
			Code:       strings.ToUpper(fieldName) + "_TOO_LONG",
			Suggestion: fmt.Sprintf("Reduce %s to at most %d characters", fieldName, max),
		}
	}
	return nil
}

// ValidateURL checks if a string is a valid URL.
func ValidateURL(value, fieldName string) *ValidationIssue {
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		return &ValidationIssue{
			Severity:   ErrorSeverity,
			Field:      fieldName,
			Message:    fieldName + " must be a valid HTTP(S) URL",
			Code:       strings.ToUpper(fieldName) + "_INVALID_URL",
			Suggestion: "Use a valid URL starting with http:// or https://",
		}
	}
	return nil
}

// ValidateEnum checks if a value is in a list of allowed values.
func ValidateEnum(value, fieldName string, allowedValues []string) *ValidationIssue {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return &ValidationIssue{
		Severity:   ErrorSeverity,
		Field:      fieldName,
		Message:    fmt.Sprintf("%s must be one of: %s", fieldName, strings.Join(allowedValues, ", ")),
		Code:       strings.ToUpper(fieldName) + "_INVALID_VALUE",
		Suggestion: "Choose from: " + strings.Join(allowedValues, ", "),
	}
}

// ValidateArray checks if an array is not empty.
func ValidateArrayNotEmpty(arr []string, fieldName string) *ValidationIssue {
	if len(arr) == 0 {
		return &ValidationIssue{
			Severity:   WarningSeverity,
			Field:      fieldName,
			Message:    fieldName + " should not be empty",
			Code:       strings.ToUpper(fieldName) + "_EMPTY",
			Suggestion: fmt.Sprintf("Add at least one %s entry", fieldName),
		}
	}
	return nil
}
