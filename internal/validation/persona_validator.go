package validation

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// PersonaValidator validates Persona elements.
type PersonaValidator struct{}

// NewPersonaValidator creates a new persona validator.
func NewPersonaValidator() *PersonaValidator {
	return &PersonaValidator{}
}

// SupportedType returns the element type this validator handles.
func (v *PersonaValidator) SupportedType() domain.ElementType {
	return domain.PersonaElement
}

// Validate performs comprehensive validation on a Persona element.
func (v *PersonaValidator) Validate(element domain.Element, level ValidationLevel) (*ValidationResult, error) {
	startTime := time.Now()

	persona, ok := element.(*domain.Persona)
	if !ok {
		return nil, errors.New("element is not a Persona type")
	}

	result := &ValidationResult{
		IsValid:     true,
		Errors:      []ValidationIssue{},
		Warnings:    []ValidationIssue{},
		Infos:       []ValidationIssue{},
		ElementType: string(domain.PersonaElement),
		ElementID:   persona.GetID(),
	}

	// Basic validation (always performed)
	v.validateBasic(persona, result)

	// Comprehensive validation
	if level == ComprehensiveLevel || level == StrictLevel {
		v.validateComprehensive(persona, result)
	}

	// Strict validation
	if level == StrictLevel {
		v.validateStrict(persona, result)
	}

	result.ValidationTime = time.Since(startTime).Milliseconds()
	return result, nil
}

func (v *PersonaValidator) validateBasic(persona *domain.Persona, result *ValidationResult) {
	metadata := persona.GetMetadata()

	// Validate required fields
	if issue := ValidateNotEmpty(metadata.Name, "name"); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	if issue := ValidateNotEmpty(metadata.Description, "description"); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	if issue := ValidateNotEmpty(persona.SystemPrompt, "system_prompt"); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	// Validate minimum lengths
	if issue := ValidateLength(metadata.Name, "name", 3, 100); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	if issue := ValidateLength(metadata.Description, "description", 10, 500); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	if issue := ValidateLength(persona.SystemPrompt, "system_prompt", 50, 2000); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}
}

func (v *PersonaValidator) validateComprehensive(persona *domain.Persona, result *ValidationResult) {
	// Validate behavioral traits
	v.validateBehavioralTraits(persona, result)

	// Validate expertise areas
	v.validateExpertiseAreas(persona, result)

	// Validate response style
	v.validateResponseStyle(persona, result)

	// Validate system prompt quality
	v.validateSystemPrompt(persona, result)
}

func (v *PersonaValidator) validateStrict(persona *domain.Persona, result *ValidationResult) {
	metadata := persona.GetMetadata()

	// Strict: Require tags
	if len(metadata.Tags) == 0 {
		result.AddErrorWithSuggestion(
			"tags",
			"Persona must have at least one tag in strict mode",
			"PERSONA_NO_TAGS",
			"Add relevant tags to improve discoverability (e.g., 'assistant', 'technical', 'creative')",
		)
	}

	// Strict: Require version
	if metadata.Version == "" {
		result.AddWarning(
			"version",
			"Persona should have a version number",
			"PERSONA_NO_VERSION",
		)
	}

	// Strict: Validate author information
	if metadata.Author == "" {
		result.AddWarning(
			"author",
			"Persona should have author information",
			"PERSONA_NO_AUTHOR",
		)
	}

	// Strict: Validate expertise depth
	if len(persona.ExpertiseAreas) > 0 && len(persona.ExpertiseAreas) < 2 {
		result.AddWarning(
			"expertise_areas",
			"Persona should have at least 2 expertise areas for better context",
			"PERSONA_LIMITED_EXPERTISE",
		)
	}
}

func (v *PersonaValidator) validateBehavioralTraits(persona *domain.Persona, result *ValidationResult) {
	if len(persona.BehavioralTraits) == 0 {
		result.AddWarning(
			"behavioral_traits",
			"Persona should define at least one behavioral trait",
			"PERSONA_NO_TRAITS",
		)
		return
	}

	// Check trait quality
	for i, trait := range persona.BehavioralTraits {
		if len(strings.TrimSpace(trait.Name)) < 3 {
			result.AddError(
				fmt.Sprintf("behavioral_traits[%d].name", i),
				"Behavioral trait name is too short",
				"PERSONA_TRAIT_NAME_SHORT",
			)
		}

		if trait.Intensity < 1 || trait.Intensity > 10 {
			result.AddError(
				fmt.Sprintf("behavioral_traits[%d].intensity", i),
				"Intensity must be between 1 and 10",
				"PERSONA_TRAIT_INVALID_INTENSITY",
			)
		}
	}
}

func (v *PersonaValidator) validateExpertiseAreas(persona *domain.Persona, result *ValidationResult) {
	if len(persona.ExpertiseAreas) == 0 {
		result.AddWarning(
			"expertise_areas",
			"Persona should define at least one area of expertise",
			"PERSONA_NO_EXPERTISE",
		)
		return
	}

	// Check expertise quality
	for i, exp := range persona.ExpertiseAreas {
		domainTrimmed := strings.TrimSpace(exp.Domain)

		// Too vague
		if len(domainTrimmed) < 3 {
			result.AddError(
				fmt.Sprintf("expertise_areas[%d].domain", i),
				"Expertise domain is too vague",
				"PERSONA_VAGUE_EXPERTISE",
			)
		}

		// Validate level
		validLevels := []string{"beginner", "intermediate", "advanced", "expert"}
		if issue := ValidateEnum(exp.Level, fmt.Sprintf("expertise_areas[%d].level", i), validLevels); issue != nil {
			result.Errors = append(result.Errors, *issue)
			result.IsValid = false
		}

		// Too generic
		genericTerms := []string{"general", "various", "multiple", "diverse"}
		for _, generic := range genericTerms {
			if strings.Contains(strings.ToLower(domainTrimmed), generic) {
				result.AddWarning(
					fmt.Sprintf("expertise_areas[%d].domain", i),
					"Expertise domain '"+domainTrimmed+"' is too generic - be more specific",
					"PERSONA_GENERIC_EXPERTISE",
				)
				break
			}
		}
	}
}

func (v *PersonaValidator) validateResponseStyle(persona *domain.Persona, result *ValidationResult) {
	style := persona.ResponseStyle

	if style.Tone == "" {
		result.AddError(
			"response_style.tone",
			"Response style tone is required",
			"PERSONA_NO_TONE",
		)
		return
	}

	// Validate formality
	validFormality := []string{"casual", "neutral", "formal"}
	if issue := ValidateEnum(style.Formality, "response_style.formality", validFormality); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	// Validate verbosity
	validVerbosity := []string{"concise", "balanced", "verbose"}
	if issue := ValidateEnum(style.Verbosity, "response_style.verbosity", validVerbosity); issue != nil {
		result.Errors = append(result.Errors, *issue)
		result.IsValid = false
	}

	// Check tone length
	if len(style.Tone) < 3 {
		result.AddWarning(
			"response_style.tone",
			"Tone description is too brief",
			"PERSONA_BRIEF_TONE",
		)
	}
}

func (v *PersonaValidator) validateSystemPrompt(persona *domain.Persona, result *ValidationResult) {
	if persona.SystemPrompt == "" {
		result.AddError(
			"system_prompt",
			"System prompt is required for AI assistant integration",
			"PERSONA_NO_SYSTEM_PROMPT",
		)
		return
	}

	// Validate minimum length
	if len(persona.SystemPrompt) < 50 {
		result.AddError(
			"system_prompt",
			"System prompt is too short to be effective (minimum 50 characters)",
			"PERSONA_SHORT_SYSTEM_PROMPT",
		)
	}

	// Check for persona reference in prompt
	metadata := persona.GetMetadata()
	if !strings.Contains(strings.ToLower(persona.SystemPrompt), strings.ToLower(metadata.Name)) {
		result.AddInfo(
			"system_prompt",
			"System prompt doesn't reference the persona name - consider adding it for clarity",
			"PERSONA_PROMPT_NO_NAME",
		)
	}

	// Check for expertise integration
	if len(persona.ExpertiseAreas) > 0 {
		promptLower := strings.ToLower(persona.SystemPrompt)
		expertiseFound := false

		for _, exp := range persona.ExpertiseAreas {
			if strings.Contains(promptLower, strings.ToLower(exp.Domain)) {
				expertiseFound = true
				break
			}
		}

		if !expertiseFound {
			result.AddInfo(
				"system_prompt",
				"System prompt doesn't mention expertise areas - consider incorporating them",
				"PERSONA_PROMPT_NO_EXPERTISE",
			)
		}
	}
}
