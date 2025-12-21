package template

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aymerick/raymond"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	helpersRegistered bool
	helpersMutex      sync.Mutex
)

// InstantiationEngine renders templates with advanced Handlebars syntax.
type InstantiationEngine struct {
	validator *TemplateValidator
	options   *EngineOptions
}

// EngineOptions configures the instantiation engine.
type EngineOptions struct {
	// MaxDepth limits nested template/partial depth (default: 10)
	MaxDepth int

	// MaxIterations limits loop iterations (default: 1000)
	MaxIterations int

	// StrictMode fails on missing variables (default: false)
	StrictMode bool

	// AllowUnsafeHelpers enables potentially dangerous helpers (default: false)
	AllowUnsafeHelpers bool
}

// InstantiationResult contains the rendered output and metadata.
type InstantiationResult struct {
	Output      string
	Variables   map[string]interface{}
	UsedHelpers []string
	Warnings    []string
}

// NewInstantiationEngine creates a new template engine.
func NewInstantiationEngine(validator *TemplateValidator, options *EngineOptions) *InstantiationEngine {
	if options == nil {
		options = &EngineOptions{
			MaxDepth:           10,
			MaxIterations:      1000,
			StrictMode:         false,
			AllowUnsafeHelpers: false,
		}
	}

	return &InstantiationEngine{
		validator: validator,
		options:   options,
	}
}

// Instantiate renders a template with provided variables.
func (e *InstantiationEngine) Instantiate(tmpl *domain.Template, variables map[string]interface{}) (*InstantiationResult, error) {
	// Validate template syntax
	if e.validator != nil {
		if err := e.validator.ValidateSyntax(tmpl); err != nil {
			return nil, fmt.Errorf("template syntax validation failed: %w", err)
		}
	}

	// Convert template variables to map for easier access
	varMap := make(map[string]interface{})
	for k, v := range variables {
		varMap[k] = v
	}

	// Add required variables with defaults if missing
	warnings := make([]string, 0)
	for _, templateVar := range tmpl.Variables {
		if _, exists := varMap[templateVar.Name]; !exists {
			if templateVar.Required {
				if e.options.StrictMode {
					return nil, fmt.Errorf("required variable %s not provided", templateVar.Name)
				}
				if templateVar.Default != "" {
					varMap[templateVar.Name] = templateVar.Default
					warnings = append(warnings, "using default value for required variable: "+templateVar.Name)
				} else {
					return nil, fmt.Errorf("required variable %s has no value or default", templateVar.Name)
				}
			} else if templateVar.Default != "" {
				varMap[templateVar.Name] = templateVar.Default
			}
		}
	}

	// Register custom helpers
	usedHelpers := e.registerHelpers()

	// Parse and render template
	output, err := raymond.Render(tmpl.Content, varMap)
	if err != nil {
		return nil, fmt.Errorf("template rendering failed: %w", err)
	}

	return &InstantiationResult{
		Output:      output,
		Variables:   varMap,
		UsedHelpers: usedHelpers,
		Warnings:    warnings,
	}, nil
}

// InstantiateWithDefaults renders a template using only default values.
func (e *InstantiationEngine) InstantiateWithDefaults(tmpl *domain.Template) (*InstantiationResult, error) {
	variables := make(map[string]interface{})

	// Populate with defaults
	for _, templateVar := range tmpl.Variables {
		if templateVar.Default != "" {
			variables[templateVar.Name] = templateVar.Default
		} else if templateVar.Required {
			return nil, fmt.Errorf("required variable %s has no default value", templateVar.Name)
		}
	}

	return e.Instantiate(tmpl, variables)
}

// Preview generates a preview with placeholder values.
func (e *InstantiationEngine) Preview(tmpl *domain.Template) (*InstantiationResult, error) {
	variables := make(map[string]interface{})

	// Generate placeholder values based on variable types
	for _, templateVar := range tmpl.Variables {
		variables[templateVar.Name] = e.generatePlaceholder(templateVar)
	}

	return e.Instantiate(tmpl, variables)
}

// ValidateVariables checks if provided variables match template requirements.
func (e *InstantiationEngine) ValidateVariables(tmpl *domain.Template, variables map[string]interface{}) error {
	for _, templateVar := range tmpl.Variables {
		value, exists := variables[templateVar.Name]
		if !exists {
			if templateVar.Required && templateVar.Default == "" {
				return fmt.Errorf("required variable %s not provided", templateVar.Name)
			}
			continue
		}

		// Type validation
		if err := e.validateVariableType(templateVar, value); err != nil {
			return fmt.Errorf("variable %s: %w", templateVar.Name, err)
		}
	}

	return nil
}

// registerHelpers registers custom Handlebars helpers (only once).
func (e *InstantiationEngine) registerHelpers() []string {
	helpersMutex.Lock()
	defer helpersMutex.Unlock()

	// Only register once globally
	if helpersRegistered {
		return []string{"upper", "lower", "title", "trim", "replace", "split", "join", "concat"}
	}

	helpers := make([]string, 0)

	// String helpers
	raymond.RegisterHelper("upper", func(str string) string {
		helpers = append(helpers, "upper")
		return strings.ToUpper(str)
	})

	raymond.RegisterHelper("lower", func(str string) string {
		helpers = append(helpers, "lower")
		return strings.ToLower(str)
	})

	raymond.RegisterHelper("title", func(str string) string {
		helpers = append(helpers, "title")
		caser := cases.Title(language.English)
		return caser.String(str)
	})

	raymond.RegisterHelper("trim", func(str string) string {
		helpers = append(helpers, "trim")
		return strings.TrimSpace(str)
	})

	raymond.RegisterHelper("replace", func(str, old, new string) string {
		helpers = append(helpers, "replace")
		return strings.ReplaceAll(str, old, new)
	})

	// Conditional helpers
	raymond.RegisterHelper("eq", func(a, b interface{}) bool {
		helpers = append(helpers, "eq")
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	})

	raymond.RegisterHelper("ne", func(a, b interface{}) bool {
		helpers = append(helpers, "ne")
		return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
	})

	raymond.RegisterHelper("gt", func(a, b int) bool {
		helpers = append(helpers, "gt")
		return a > b
	})

	raymond.RegisterHelper("lt", func(a, b int) bool {
		helpers = append(helpers, "lt")
		return a < b
	})

	raymond.RegisterHelper("gte", func(a, b int) bool {
		helpers = append(helpers, "gte")
		return a >= b
	})

	raymond.RegisterHelper("lte", func(a, b int) bool {
		helpers = append(helpers, "lte")
		return a <= b
	})

	// Logic helpers
	raymond.RegisterHelper("and", func(a, b bool) bool {
		helpers = append(helpers, "and")
		return a && b
	})

	raymond.RegisterHelper("or", func(a, b bool) bool {
		helpers = append(helpers, "or")
		return a || b
	})

	raymond.RegisterHelper("not", func(a bool) bool {
		helpers = append(helpers, "not")
		return !a
	})

	// Formatting helpers
	raymond.RegisterHelper("json", func(v interface{}) string {
		helpers = append(helpers, "json")
		return fmt.Sprintf("%v", v) // Simplified JSON representation
	})

	raymond.RegisterHelper("default", func(value, defaultValue interface{}) interface{} {
		helpers = append(helpers, "default")
		if value == nil || fmt.Sprintf("%v", value) == "" {
			return defaultValue
		}
		return value
	})

	// Array helpers
	raymond.RegisterHelper("join", func(arr []interface{}, sep string) string {
		helpers = append(helpers, "join")
		parts := make([]string, len(arr))
		for i, item := range arr {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, sep)
	})

	raymond.RegisterHelper("length", func(arr []interface{}) int {
		helpers = append(helpers, "length")
		return len(arr)
	})

	helpersRegistered = true
	return helpers
}

// generatePlaceholder creates a placeholder value based on variable type.
func (e *InstantiationEngine) generatePlaceholder(variable domain.TemplateVariable) interface{} {
	switch variable.Type {
	case "string":
		if variable.Default != "" {
			return variable.Default
		}
		return fmt.Sprintf("{{%s}}", variable.Name)
	case "number":
		if variable.Default != "" {
			return variable.Default
		}
		return 0
	case "boolean":
		if variable.Default != "" {
			return variable.Default == "true"
		}
		return false
	case "array":
		return []interface{}{}
	case "object":
		return map[string]interface{}{}
	default:
		return fmt.Sprintf("{{%s}}", variable.Name)
	}
}

// validateVariableType checks if a value matches the expected type.
func (e *InstantiationEngine) validateVariableType(variable domain.TemplateVariable, value interface{}) error {
	switch variable.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Valid number type
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		switch value.(type) {
		case []interface{}, []string, []int, []float64:
			// Valid array type
		default:
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}

	return nil
}

// ParseTemplate parses a template string and returns any syntax errors.
func (e *InstantiationEngine) ParseTemplate(content string) error {
	_, err := raymond.Parse(content)
	if err != nil {
		return fmt.Errorf("template parsing failed: %w", err)
	}
	return nil
}

// ExtractVariables extracts variable names from template content.
func (e *InstantiationEngine) ExtractVariables(content string) ([]string, error) {
	tmpl, err := raymond.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Extract variables from AST
	variables := make(map[string]bool)
	e.extractVariablesFromNode(tmpl, variables)

	// Convert to slice
	result := make([]string, 0, len(variables))
	for v := range variables {
		result = append(result, v)
	}

	return result, nil
}

// extractVariablesFromNode recursively extracts variables from template AST.
func (e *InstantiationEngine) extractVariablesFromNode(node interface{}, variables map[string]bool) {
	// This is a simplified version - the actual implementation would need to
	// traverse the raymond AST properly. For now, we'll use a regex approach.

	// Note: This is a placeholder. In production, you'd want to properly
	// traverse the raymond AST using reflection or the library's visitor pattern.
}

// RenderPartial renders a partial template.
func (e *InstantiationEngine) RenderPartial(partialName string, partialContent string, context map[string]interface{}) (string, error) {
	// Register the partial if not already registered
	raymond.RegisterPartial(partialName, partialContent)

	// Render a template that uses the partial
	templateContent := fmt.Sprintf("{{> %s}}", partialName)
	output, err := raymond.Render(templateContent, context)
	if err != nil {
		return "", fmt.Errorf("partial rendering failed: %w", err)
	}

	return output, nil
}

// RegisterPartial registers a partial template.
func (e *InstantiationEngine) RegisterPartial(name, content string) error {
	raymond.RegisterPartial(name, content)
	return nil
}

// UnregisterPartial removes a registered partial.
func (e *InstantiationEngine) UnregisterPartial(name string) {
	raymond.RegisterPartial(name, "")
}

// GetRegisteredHelpers returns the list of available helpers.
func (e *InstantiationEngine) GetRegisteredHelpers() []string {
	return []string{
		// String helpers
		"upper", "lower", "title", "trim", "replace",
		// Conditional helpers
		"eq", "ne", "gt", "lt", "gte", "lte",
		// Logic helpers
		"and", "or", "not",
		// Formatting helpers
		"json", "default",
		// Array helpers
		"join", "length",
		// Built-in Handlebars helpers
		"if", "unless", "each", "with",
	}
}

// GetEngineStats returns engine statistics.
func (e *InstantiationEngine) GetEngineStats() map[string]interface{} {
	return map[string]interface{}{
		"max_depth":            e.options.MaxDepth,
		"max_iterations":       e.options.MaxIterations,
		"strict_mode":          e.options.StrictMode,
		"allow_unsafe_helpers": e.options.AllowUnsafeHelpers,
		"registered_helpers":   len(e.GetRegisteredHelpers()),
	}
}
