package template

import (
	"fmt"
	"strings"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// TemplateValidator performs comprehensive template validation.
type TemplateValidator struct {
	maxTemplateSize int // Maximum template size in bytes (default: 1MB)
	maxVariables    int // Maximum number of variables (default: 100)
}

// ValidationError represents a template validation error.
type ValidationError struct {
	Field   string
	Message string
	Fix     string
}

// ValidationResult contains validation results.
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []string
}

// NewTemplateValidator creates a new template validator.
func NewTemplateValidator() *TemplateValidator {
	return &TemplateValidator{
		maxTemplateSize: 1024 * 1024, // 1MB
		maxVariables:    100,
	}
}

// ValidateSyntax validates template syntax (Handlebars).
func (v *TemplateValidator) ValidateSyntax(tmpl *domain.Template) error {
	// Check template size
	if len(tmpl.Content) > v.maxTemplateSize {
		return fmt.Errorf("template exceeds maximum size of %d bytes", v.maxTemplateSize)
	}

	// Check variable count
	if len(tmpl.Variables) > v.maxVariables {
		return fmt.Errorf("template has too many variables (max: %d)", v.maxVariables)
	}

	// Basic syntax checks
	if err := v.checkBalancedDelimiters(tmpl.Content); err != nil {
		return err
	}

	if err := v.checkVariableReferences(tmpl); err != nil {
		return err
	}

	return nil
}

// ValidateComprehensive performs full validation.
func (v *TemplateValidator) ValidateComprehensive(tmpl *domain.Template, variables map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]string, 0),
	}

	// Validate syntax
	if err := v.ValidateSyntax(tmpl); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "content",
			Message: err.Error(),
			Fix:     "Fix template syntax errors",
		})
	}

	// Validate variables
	for _, templateVar := range tmpl.Variables {
		// Check if required variable is provided
		if templateVar.Required {
			if _, exists := variables[templateVar.Name]; !exists && templateVar.Default == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   templateVar.Name,
					Message: fmt.Sprintf("required variable %s not provided", templateVar.Name),
					Fix:     "provide value for variable: " + templateVar.Name,
				})
			}
		}

		// Check variable type if provided
		if value, exists := variables[templateVar.Name]; exists {
			if !v.checkVariableType(value, templateVar.Type) {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   templateVar.Name,
					Message: fmt.Sprintf("variable %s has wrong type, expected %s", templateVar.Name, templateVar.Type),
					Fix:     fmt.Sprintf("provide value of type %s for variable: %s", templateVar.Type, templateVar.Name),
				})
			}
		}

		// Check variable name format
		if !isValidVariableName(templateVar.Name) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   templateVar.Name,
				Message: "invalid variable name format",
				Fix:     "use alphanumeric characters and underscores only",
			})
		}
	}

	// Check for undefined variables in content
	undefinedVars := v.findUndefinedVariables(tmpl)
	for _, varName := range undefinedVars {
		result.Warnings = append(result.Warnings, fmt.Sprintf("variable %s used in template but not declared", varName))
	}

	return result
}

// checkVariableType validates that a value matches the expected type.
func (v *TemplateValidator) checkVariableType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return true
		default:
			return false
		}
	case "boolean", "bool":
		_, ok := value.(bool)
		return ok
	case "array":
		switch value.(type) {
		case []interface{}, []string, []int, []float64:
			return true
		default:
			return false
		}
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		// Unknown type, allow it
		return true
	}
}

// ValidateOutput validates the rendered output.
func (v *TemplateValidator) ValidateOutput(tmpl *domain.Template, output string) error {
	// Validate based on format
	switch tmpl.Format {
	case domain.FormatJSON:
		// TODO: Validate JSON format
	case domain.FormatYAML:
		// TODO: Validate YAML format
	case domain.FormatMarkdown:
		// Basic markdown validation
	case domain.FormatText:
		// No specific validation for plain text
	}

	return nil
}

// checkBalancedDelimiters ensures all {{}} are balanced and block helpers are closed.
func (v *TemplateValidator) checkBalancedDelimiters(content string) error {
	blockStack := make([]string, 0) // Stack to track open block helpers

	i := 0
	for i < len(content) {
		if i < len(content)-1 && content[i:i+2] == "{{" {
			// Find closing }}
			end := i + 2
			for end < len(content)-1 {
				if content[end:end+2] == "}}" {
					break
				}
				end++
			}

			if end >= len(content)-1 {
				return fmt.Errorf("unclosed delimiter starting at position %d", i)
			}

			// Extract the expression
			expr := strings.TrimSpace(content[i+2 : end])

			// Check for block helpers
			if strings.HasPrefix(expr, "#") {
				// Opening block helper (e.g., {{#if}}, {{#each}})
				blockName := strings.Fields(expr[1:])[0]
				blockStack = append(blockStack, blockName)
			} else if strings.HasPrefix(expr, "/") {
				// Closing block helper (e.g., {{/if}}, {{/each}})
				blockName := strings.TrimSpace(expr[1:])
				if len(blockStack) == 0 {
					return fmt.Errorf("unexpected closing block {{/%s}} at position %d", blockName, i)
				}
				lastBlock := blockStack[len(blockStack)-1]
				if lastBlock != blockName {
					return fmt.Errorf("mismatched block helper: expected {{/%s}}, got {{/%s}} at position %d", lastBlock, blockName, i)
				}
				blockStack = blockStack[:len(blockStack)-1]
			}

			i = end + 2
			continue
		}

		// Check for closing delimiter without opening
		if i < len(content)-1 && content[i:i+2] == "}}" {
			return fmt.Errorf("unexpected closing delimiter }} at position %d", i)
		}

		i++
	}

	if len(blockStack) > 0 {
		return fmt.Errorf("unclosed block helper: {{#%s}}", blockStack[len(blockStack)-1])
	}

	return nil
}

// checkVariableReferences validates variable references in template.
func (v *TemplateValidator) checkVariableReferences(tmpl *domain.Template) error {
	// For now, just a placeholder
	// In a real implementation, we'd parse the template and check all variable references
	return nil
}

// findUndefinedVariables finds variables used but not declared.
func (v *TemplateValidator) findUndefinedVariables(tmpl *domain.Template) []string {
	declared := make(map[string]bool)
	for _, v := range tmpl.Variables {
		declared[v.Name] = true
	}

	undefined := []string{}
	content := tmpl.Content

	i := 0
	for i < len(content) {
		if i < len(content)-1 && content[i:i+2] == "{{" {
			// Find closing }}
			end := i + 2
			for end < len(content)-1 {
				if content[end:end+2] == "}}" {
					varExpr := strings.TrimSpace(content[i+2 : end])

					// Skip helpers and block expressions
					if strings.HasPrefix(varExpr, "#") || strings.HasPrefix(varExpr, "/") || strings.HasPrefix(varExpr, ">") {
						break
					}

					// Extract just the variable name (before any spaces/pipes)
					varName := strings.Split(varExpr, " ")[0]
					varName = strings.Split(varName, "|")[0]
					varName = strings.TrimSpace(varName)

					if varName != "" && !declared[varName] && !isHelper(varName) {
						// Check if already in undefined list
						found := false
						for _, u := range undefined {
							if u == varName {
								found = true
								break
							}
						}
						if !found {
							undefined = append(undefined, varName)
						}
					}
					break
				}
				end++
			}
			i = end + 2
			continue
		}
		i++
	}

	return undefined
}

// isValidVariableName checks if a variable name is valid.
func isValidVariableName(name string) bool {
	if name == "" {
		return false
	}

	for i, ch := range name {
		if i == 0 {
			// First character must be letter or underscore
			//nolint:staticcheck // Keep explicit OR logic for clarity
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_') {
				return false
			}
		} else {
			// Subsequent characters can be letter, digit, or underscore
			//nolint:staticcheck // Keep explicit OR logic for clarity
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
				return false
			}
		}
	}

	return true
}

// isHelper checks if a name is a known Handlebars helper.
func isHelper(name string) bool {
	helpers := map[string]bool{
		"if": true, "unless": true, "each": true, "with": true,
		"eq": true, "ne": true, "gt": true, "lt": true, "gte": true, "lte": true,
		"and": true, "or": true, "not": true,
		"upper": true, "lower": true, "title": true, "trim": true, "replace": true,
		"json": true, "default": true, "join": true, "length": true,
	}
	return helpers[name]
}
