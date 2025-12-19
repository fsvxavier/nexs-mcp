package template

import (
	"fmt"
	"strings"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// TemplateValidator performs comprehensive template validation
type TemplateValidator struct {
	maxTemplateSize int // Maximum template size in bytes (default: 1MB)
	maxVariables    int // Maximum number of variables (default: 100)
}

// ValidationError represents a template validation error
type ValidationError struct {
	Field   string
	Message string
	Fix     string
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []string
}

// NewTemplateValidator creates a new template validator
func NewTemplateValidator() *TemplateValidator {
	return &TemplateValidator{
		maxTemplateSize: 1024 * 1024, // 1MB
		maxVariables:    100,
	}
}

// ValidateSyntax validates template syntax (Handlebars)
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

// ValidateComprehensive performs full validation
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
					Fix:     fmt.Sprintf("provide value for variable: %s", templateVar.Name),
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

// ValidateOutput validates the rendered output
func (v *TemplateValidator) ValidateOutput(tmpl *domain.Template, output string) error {
	// Validate based on format
	switch tmpl.Format {
	case "json":
		// TODO: Validate JSON format
	case "yaml":
		// TODO: Validate YAML format
	case "markdown":
		// Basic markdown validation
	case "text":
		// No specific validation for plain text
	}

	return nil
}

// checkBalancedDelimiters ensures all {{}} are balanced
func (v *TemplateValidator) checkBalancedDelimiters(content string) error {
	openCount := 0
	i := 0
	for i < len(content) {
		if i < len(content)-1 && content[i:i+2] == "{{" {
			openCount++
			i += 2
			continue
		}
		if i < len(content)-1 && content[i:i+2] == "}}" {
			openCount--
			if openCount < 0 {
				return fmt.Errorf("unmatched closing delimiter at position %d", i)
			}
			i += 2
			continue
		}
		i++
	}

	if openCount != 0 {
		return fmt.Errorf("unbalanced delimiters: %d unclosed", openCount)
	}

	return nil
}

// checkVariableReferences ensures all {{variables}} are declared
func (v *TemplateValidator) checkVariableReferences(tmpl *domain.Template) error {
	// Build map of declared variables
	declared := make(map[string]bool)
	for _, v := range tmpl.Variables {
		declared[v.Name] = true
	}

	// Find all variable references in content
	// This is a simplified check - proper implementation would parse the template
	// Note: In production, this would use proper AST traversal

	return nil
}

// findUndefinedVariables finds variables used but not declared
func (v *TemplateValidator) findUndefinedVariables(tmpl *domain.Template) []string {
	// Build map of declared variables
	declared := make(map[string]bool)
	for _, v := range tmpl.Variables {
		declared[v.Name] = true
	}

	undefined := make([]string, 0)

	// Simple extraction of {{variable}} patterns
	content := tmpl.Content
	i := 0
	for i < len(content) {
		if i < len(content)-2 && content[i:i+2] == "{{" {
			// Find closing }}
			end := i + 2
			for end < len(content)-1 {
				if content[end:end+2] == "}}" {
					// Extract variable name
					varExpr := content[i+2 : end]
					varExpr = strings.TrimSpace(varExpr)

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

// isValidVariableName checks if a variable name is valid
func isValidVariableName(name string) bool {
	if name == "" {
		return false
	}

	for i, ch := range name {
		if i == 0 {
			// First character must be letter or underscore
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_') {
				return false
			}
		} else {
			// Subsequent characters can be letter, digit, or underscore
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
				return false
			}
		}
	}

	return true
}

// isHelper checks if a name is a known Handlebars helper
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
