package validation

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// TestValidationResult tests the ValidationResult structure.
func TestValidationResult(t *testing.T) {
	vr := &ValidationResult{
		IsValid:     true,
		ElementType: "persona",
		ElementID:   "test-id",
	}

	// Test adding errors
	vr.AddError("field1", "error message", "ERR_001")
	if vr.IsValid {
		t.Error("Expected IsValid to be false after adding error")
	}
	if vr.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", vr.ErrorCount())
	}

	// Test adding warnings
	vr.AddWarning("field2", "warning message", "WARN_001")
	if vr.WarningCount() != 1 {
		t.Errorf("Expected 1 warning, got %d", vr.WarningCount())
	}

	// Test adding info
	vr.AddInfo("field3", "info message", "INFO_001")
	if len(vr.Infos) != 1 {
		t.Errorf("Expected 1 info, got %d", len(vr.Infos))
	}

	// Test total issues
	if vr.TotalIssues() != 3 {
		t.Errorf("Expected 3 total issues, got %d", vr.TotalIssues())
	}
}

// TestValidationResultWithSuggestion tests error with suggestion.
func TestValidationResultWithSuggestion(t *testing.T) {
	vr := &ValidationResult{
		IsValid:     true,
		ElementType: "skill",
		ElementID:   "test-id",
	}

	vr.AddErrorWithSuggestion("triggers", "At least one trigger required", "ERR_TRIGGER", "Add a keyword trigger")

	if vr.IsValid {
		t.Error("Expected IsValid to be false after adding error")
	}
	if len(vr.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(vr.Errors))
	}
	if vr.Errors[0].Suggestion != "Add a keyword trigger" {
		t.Errorf("Expected suggestion 'Add a keyword trigger', got %s", vr.Errors[0].Suggestion)
	}
}

// TestValidatorRegistry tests the registry functionality.
func TestValidatorRegistry(t *testing.T) {
	registry := NewValidatorRegistry()

	// Test getting validators for all types
	types := []domain.ElementType{
		domain.PersonaElement,
		domain.SkillElement,
		domain.TemplateElement,
		domain.AgentElement,
		domain.MemoryElement,
		domain.EnsembleElement,
	}

	for _, elemType := range types {
		validator, err := registry.GetValidator(elemType)
		if err != nil {
			t.Errorf("Failed to get validator for %s: %v", elemType, err)
		}
		if validator == nil {
			t.Errorf("Validator for %s is nil", elemType)
		}
		if validator.SupportedType() != elemType {
			t.Errorf("Validator type mismatch: expected %s, got %s", elemType, validator.SupportedType())
		}
	}
}

// TestValidatorRegistry_InvalidType tests error handling for invalid type.
func TestValidatorRegistry_InvalidType(t *testing.T) {
	registry := NewValidatorRegistry()

	_, err := registry.GetValidator("invalid_type")
	if err == nil {
		t.Error("Expected error for invalid element type")
	}
}

// TestValidationLevels tests validation level constants.
func TestValidationLevels(t *testing.T) {
	levels := []ValidationLevel{
		BasicLevel,
		ComprehensiveLevel,
		StrictLevel,
	}

	expected := []string{"basic", "comprehensive", "strict"}

	for i, level := range levels {
		if string(level) != expected[i] {
			t.Errorf("Expected level %s, got %s", expected[i], level)
		}
	}
}

// TestValidationSeverities tests validation severity constants.
func TestValidationSeverities(t *testing.T) {
	severities := []ValidationSeverity{
		ErrorSeverity,
		WarningSeverity,
		InfoSeverity,
	}

	expected := []string{"error", "warning", "info"}

	for i, severity := range severities {
		if string(severity) != expected[i] {
			t.Errorf("Expected severity %s, got %s", expected[i], severity)
		}
	}
}

// TestPersonaValidation_Basic tests basic persona validation.
func TestPersonaValidation_Basic(t *testing.T) {
	registry := NewValidatorRegistry()
	validator, _ := registry.GetValidator(domain.PersonaElement)

	// Valid persona
	persona := domain.NewPersona("Test Persona", "A test persona", "1.0.0", "tester")
	persona.BehavioralTraits = []domain.BehavioralTrait{
		{Name: "helpful", Intensity: 8},
	}
	persona.ExpertiseAreas = []domain.ExpertiseArea{
		{Domain: "testing", Level: "expert", Keywords: []string{"qa"}},
	}
	persona.SystemPrompt = "You are a helpful testing assistant specialized in quality assurance"

	result, err := validator.Validate(persona, BasicLevel)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
	if !result.IsValid {
		t.Errorf("Expected valid persona, got errors: %v", result.Errors)
	}
}

// TestPersonaValidation_MissingRequiredFields tests missing fields.
func TestPersonaValidation_MissingRequiredFields(t *testing.T) {
	registry := NewValidatorRegistry()
	validator, _ := registry.GetValidator(domain.PersonaElement)

	// Persona without behavioral traits
	persona := domain.NewPersona("Test Persona", "A test persona", "1.0.0", "tester")
	persona.SystemPrompt = "You are a helpful assistant"

	result, err := validator.Validate(persona, BasicLevel)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
	if result.IsValid {
		t.Error("Expected validation to fail for persona without behavioral traits")
	}
	if result.ErrorCount() == 0 {
		t.Error("Expected at least one error")
	}
}

// TestSkillValidation_Basic tests basic skill validation.
func TestSkillValidation_Basic(t *testing.T) {
	registry := NewValidatorRegistry()
	validator, _ := registry.GetValidator(domain.SkillElement)

	// Valid skill
	skill := domain.NewSkill("Test Skill", "A test skill", "1.0.0", "tester")
	skill.Triggers = []domain.SkillTrigger{
		{Type: "keyword", Keywords: []string{"test"}},
	}
	skill.Procedures = []domain.SkillProcedure{
		{Step: 1, Action: "Run test", Description: "Execute test suite"},
	}

	result, err := validator.Validate(skill, BasicLevel)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
	if !result.IsValid {
		t.Errorf("Expected valid skill, got errors: %v", result.Errors)
	}
}

// TestTemplateValidation_Basic tests basic template validation.
func TestTemplateValidation_Basic(t *testing.T) {
	registry := NewValidatorRegistry()
	validator, _ := registry.GetValidator(domain.TemplateElement)

	// Valid template
	template := domain.NewTemplate("Test Template", "A test template", "1.0.0", "tester")
	template.Content = "Hello {{name}}"
	template.Format = "text"

	result, err := validator.Validate(template, BasicLevel)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
	if !result.IsValid {
		t.Errorf("Expected valid template, got errors: %v", result.Errors)
	}
}

// TestValidationIssueStructure tests the ValidationIssue structure.
func TestValidationIssueStructure(t *testing.T) {
	issue := ValidationIssue{
		Severity:   ErrorSeverity,
		Field:      "test_field",
		Message:    "Test message",
		Code:       "TEST_001",
		Suggestion: "Fix the issue",
		Line:       42,
	}

	if issue.Severity != ErrorSeverity {
		t.Errorf("Expected severity %s, got %s", ErrorSeverity, issue.Severity)
	}
	if issue.Field != "test_field" {
		t.Errorf("Expected field 'test_field', got %s", issue.Field)
	}
	if issue.Line != 42 {
		t.Errorf("Expected line 42, got %d", issue.Line)
	}
}
