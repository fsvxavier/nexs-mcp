package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-mcp/internal/common"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/validation"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ValidateElementInput represents the input for validate_element tool.
type ValidateElementInput struct {
	ElementID       string `json:"element_id"`
	ElementType     string `json:"element_type"`
	ValidationLevel string `json:"validation_level,omitempty"`
	FixSuggestions  bool   `json:"fix_suggestions,omitempty"`
}

// ValidationIssue represents a validation error, warning, or info.
type ValidationIssue struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Severity   string `json:"severity"`
}

// ValidateElementOutput represents the output of validate_element tool.
type ValidateElementOutput struct {
	ElementID      string            `json:"element_id"`
	ElementType    string            `json:"element_type"`
	IsValid        bool              `json:"is_valid"`
	ErrorCount     int               `json:"error_count"`
	WarningCount   int               `json:"warning_count"`
	ValidationTime string            `json:"validation_time"`
	Errors         []ValidationIssue `json:"errors,omitempty"`
	Warnings       []ValidationIssue `json:"warnings,omitempty"`
	Infos          []ValidationIssue `json:"infos,omitempty"`
	Summary        string            `json:"summary"`
}

// handleValidateElement handles validate_element tool calls.
func (s *MCPServer) handleValidateElement(ctx context.Context, req *sdk.CallToolRequest, input ValidateElementInput) (*sdk.CallToolResult, ValidateElementOutput, error) {
	// Validate required inputs
	if input.ElementID == "" {
		return nil, ValidateElementOutput{}, errors.New("element_id is required")
	}
	if input.ElementType == "" {
		return nil, ValidateElementOutput{}, errors.New("element_type is required")
	}

	// Parse element type
	var elementType domain.ElementType
	switch input.ElementType {
	case common.ElementTypePersona:
		elementType = domain.PersonaElement
	case common.ElementTypeSkill:
		elementType = domain.SkillElement
	case common.ElementTypeTemplate:
		elementType = domain.TemplateElement
	case common.ElementTypeAgent:
		elementType = domain.AgentElement
	case common.ElementTypeMemory:
		elementType = domain.MemoryElement
	case common.ElementTypeEnsemble:
		elementType = domain.EnsembleElement
	default:
		return nil, ValidateElementOutput{}, fmt.Errorf("invalid element_type: %s", input.ElementType)
	}

	// Parse validation level (default: comprehensive)
	validationLevel := validation.ComprehensiveLevel
	if input.ValidationLevel != "" {
		switch input.ValidationLevel {
		case "basic":
			validationLevel = validation.BasicLevel
		case "comprehensive":
			validationLevel = validation.ComprehensiveLevel
		case "strict":
			validationLevel = validation.StrictLevel
		default:
			return nil, ValidateElementOutput{}, fmt.Errorf("invalid validation_level: %s", input.ValidationLevel)
		}
	}

	// Retrieve element from repository
	element, err := s.repo.GetByID(input.ElementID)
	if err != nil {
		return nil, ValidateElementOutput{}, fmt.Errorf("element not found: %w", err)
	}

	// Verify element type matches
	if element.GetType() != elementType {
		return nil, ValidateElementOutput{}, fmt.Errorf("element type mismatch: expected %s, got %s", input.ElementType, element.GetType())
	}

	// Create validator registry
	registry := validation.NewValidatorRegistry()

	// Perform validation
	result, err := registry.ValidateElement(element, validationLevel)
	if err != nil {
		return nil, ValidateElementOutput{}, fmt.Errorf("validation error: %w", err)
	}

	// Convert validation issues
	errors := make([]ValidationIssue, len(result.Errors))
	for i, issue := range result.Errors {
		errors[i] = ValidationIssue{
			Field:      issue.Field,
			Message:    issue.Message,
			Code:       issue.Code,
			Suggestion: issue.Suggestion,
			Severity:   "error",
		}
		// Remove suggestions if not requested
		if !input.FixSuggestions {
			errors[i].Suggestion = ""
		}
	}

	warnings := make([]ValidationIssue, len(result.Warnings))
	for i, issue := range result.Warnings {
		warnings[i] = ValidationIssue{
			Field:      issue.Field,
			Message:    issue.Message,
			Code:       issue.Code,
			Suggestion: issue.Suggestion,
			Severity:   "warning",
		}
		// Remove suggestions if not requested
		if !input.FixSuggestions {
			warnings[i].Suggestion = ""
		}
	}

	infos := make([]ValidationIssue, len(result.Infos))
	for i, issue := range result.Infos {
		infos[i] = ValidationIssue{
			Field:      issue.Field,
			Message:    issue.Message,
			Code:       issue.Code,
			Suggestion: issue.Suggestion,
			Severity:   "info",
		}
	}

	// Generate summary
	summary := "Validation ✅ PASSED"
	if !result.IsValid {
		summary = fmt.Sprintf("Validation ❌ FAILED (%d errors, %d warnings)", result.ErrorCount(), result.WarningCount())
	} else if result.WarningCount() > 0 {
		summary = fmt.Sprintf("Validation ✅ PASSED (with %d warnings)", result.WarningCount())
	}

	output := ValidateElementOutput{
		ElementID:      input.ElementID,
		ElementType:    input.ElementType,
		IsValid:        result.IsValid,
		ErrorCount:     result.ErrorCount(),
		WarningCount:   result.WarningCount(),
		ValidationTime: fmt.Sprintf("%dms", result.ValidationTime),
		Errors:         errors,
		Warnings:       warnings,
		Infos:          infos,
		Summary:        summary,
	}

	return nil, output, nil
}

// formatValidationResultJSON formats validation result as JSON string (helper function).
//
//nolint:unused // Reserved for future use
func formatValidationResultJSON(output ValidateElementOutput) string {
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting result: %v", err)
	}
	return string(jsonBytes)
}
