package validation

import (
	"errors"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// EnsembleValidator validates Ensemble elements.
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
		return nil, errors.New("element is not an Ensemble type")
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
