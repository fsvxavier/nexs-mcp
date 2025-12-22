package template

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateValidator(t *testing.T) {
	validator := NewTemplateValidator()
	require.NotNil(t, validator)
	assert.Equal(t, 1024*1024, validator.maxTemplateSize)
	assert.Equal(t, 100, validator.maxVariables)
}

func TestValidateSyntax_ValidTemplate(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template at least ten characters", "1.0", "test-author")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
	}

	err := validator.ValidateSyntax(tmpl)
	assert.NoError(t, err)
}

func TestValidateSyntax_TemplateTooBig(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template at least ten characters", "1.0", "test-author")

	// Create content larger than 1MB
	largeContent := make([]byte, 2*1024*1024)
	for i := range largeContent {
		largeContent[i] = 'x'
	}
	tmpl.Content = string(largeContent)

	err := validator.ValidateSyntax(tmpl)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum size")
}

func TestValidateSyntax_TooManyVariables(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template at least ten characters", "1.0", "test-author")
	tmpl.Content = "test"

	// Create more than 100 variables
	variables := make([]domain.TemplateVariable, 101)
	for i := range 101 {
		variables[i] = domain.TemplateVariable{
			Name: "var" + string(rune(i)),
			Type: "string",
		}
	}
	tmpl.Variables = variables

	err := validator.ValidateSyntax(tmpl)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many variables")
}

func TestCheckBalancedDelimiters_Balanced(t *testing.T) {
	validator := NewTemplateValidator()

	tests := []struct {
		name    string
		content string
	}{
		{"Simple variable", "Hello {{name}}!"},
		{"Multiple variables", "{{first}} {{last}}"},
		{"Block helper", "{{#if condition}}yes{{/if}}"},
		{"Nested blocks", "{{#each items}}{{#if active}}{{name}}{{/if}}{{/each}}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.checkBalancedDelimiters(tt.content)
			assert.NoError(t, err)
		})
	}
}

func TestCheckBalancedDelimiters_Unbalanced(t *testing.T) {
	validator := NewTemplateValidator()

	tests := []struct {
		name    string
		content string
		errMsg  string
	}{
		{"Unclosed opening", "Hello {{name", "unclosed delimiter"},
		{"Unexpected closing", "Hello }} name", "unexpected closing delimiter"},
		{"Unclosed block", "{{#if test}}content", "unclosed block helper"},
		{"Mismatched block", "{{#if test}}{{/each}}", "mismatched block helper"},
		{"Extra closing block", "{{/if}}", "unexpected closing block"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.checkBalancedDelimiters(tt.content)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestValidateComprehensive_ValidTemplate(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template at least ten characters", "1.0", "test-author")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
	}

	variables := map[string]interface{}{
		"name": "World",
	}

	result := validator.ValidateComprehensive(tmpl, variables)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateComprehensive_MissingRequiredVariable(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template at least ten characters", "1.0", "test-author")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
	}

	variables := map[string]interface{}{}

	result := validator.ValidateComprehensive(tmpl, variables)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Message, "required variable")
}

func TestValidateComprehensive_WrongVariableType(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template at least ten characters", "1.0", "test-author")
	tmpl.Content = "Count: {{count}}"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "count", Type: "number", Required: true},
	}

	variables := map[string]interface{}{
		"count": "not a number",
	}

	result := validator.ValidateComprehensive(tmpl, variables)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Message, "wrong type")
}

func TestValidateComprehensive_InvalidVariableName(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template at least ten characters", "1.0", "test-author")
	tmpl.Content = "Hello {{123invalid}}!"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "123invalid", Type: "string"},
	}

	result := validator.ValidateComprehensive(tmpl, map[string]interface{}{})
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Message, "invalid variable name")
}

func TestValidateComprehensive_WithDefault(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template at least ten characters", "1.0", "test-author")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true, Default: "Guest"},
	}

	// Even without providing the variable, it should be valid because it has default
	variables := map[string]interface{}{}

	result := validator.ValidateComprehensive(tmpl, variables)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateComprehensive_UndefinedVariableWarning(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template at least ten characters", "1.0", "test-author")
	tmpl.Content = "Hello {{name}}! Age: {{age}}"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
	}

	variables := map[string]interface{}{
		"name": "John",
	}

	result := validator.ValidateComprehensive(tmpl, variables)
	assert.True(t, result.Valid)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "age")
}

func TestCheckVariableType_String(t *testing.T) {
	validator := NewTemplateValidator()

	assert.True(t, validator.checkVariableType("hello", "string"))
	assert.False(t, validator.checkVariableType(123, "string"))
}

func TestCheckVariableType_Number(t *testing.T) {
	validator := NewTemplateValidator()

	assert.True(t, validator.checkVariableType(123, "number"))
	assert.True(t, validator.checkVariableType(int64(123), "number"))
	assert.True(t, validator.checkVariableType(float64(123.45), "number"))
	assert.False(t, validator.checkVariableType("123", "number"))
}

func TestCheckVariableType_Boolean(t *testing.T) {
	validator := NewTemplateValidator()

	assert.True(t, validator.checkVariableType(true, "boolean"))
	assert.True(t, validator.checkVariableType(false, "bool"))
	assert.False(t, validator.checkVariableType("true", "boolean"))
}

func TestCheckVariableType_Array(t *testing.T) {
	validator := NewTemplateValidator()

	assert.True(t, validator.checkVariableType([]interface{}{"a", "b"}, "array"))
	assert.True(t, validator.checkVariableType([]string{"a", "b"}, "array"))
	assert.True(t, validator.checkVariableType([]int{1, 2}, "array"))
	assert.False(t, validator.checkVariableType("not an array", "array"))
}

func TestCheckVariableType_Object(t *testing.T) {
	validator := NewTemplateValidator()

	obj := map[string]interface{}{"key": "value"}
	assert.True(t, validator.checkVariableType(obj, "object"))
	assert.False(t, validator.checkVariableType("not an object", "object"))
}

func TestCheckVariableType_UnknownType(t *testing.T) {
	validator := NewTemplateValidator()

	// Unknown types should be allowed
	assert.True(t, validator.checkVariableType("anything", "unknown"))
	assert.True(t, validator.checkVariableType(123, "custom_type"))
}

func TestIsValidVariableName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"Valid simple", "name", true},
		{"Valid with underscore", "user_name", true},
		{"Valid with number", "var1", true},
		{"Valid uppercase", "USER_NAME", true},
		{"Valid mixed case", "userName", true},
		{"Invalid starts with number", "1name", false},
		{"Invalid empty", "", false},
		{"Invalid with hyphen", "user-name", false},
		{"Invalid with space", "user name", false},
		{"Invalid with special char", "user$name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVariableName(tt.input)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsHelper(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"if", true},
		{"unless", true},
		{"each", true},
		{"eq", true},
		{"upper", true},
		{"unknown", false},
		{"custom", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHelper(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindUndefinedVariables(t *testing.T) {
	validator := NewTemplateValidator()

	tests := []struct {
		name     string
		content  string
		declared []string
		expected []string
	}{
		{
			name:     "No undefined",
			content:  "Hello {{name}}!",
			declared: []string{"name"},
			expected: []string{},
		},
		{
			name:     "One undefined",
			content:  "Hello {{name}}! Age: {{age}}",
			declared: []string{"name"},
			expected: []string{"age"},
		},
		{
			name:     "Multiple undefined",
			content:  "{{first}} {{last}} {{email}}",
			declared: []string{"first"},
			expected: []string{"last", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := domain.NewTemplate("test", "Test template", "1.0", "author")
			tmpl.Content = tt.content
			tmpl.Variables = make([]domain.TemplateVariable, len(tt.declared))
			for i, name := range tt.declared {
				tmpl.Variables[i] = domain.TemplateVariable{Name: name}
			}

			result := validator.findUndefinedVariables(tmpl)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestValidateOutput(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template", "1.0", "author")

	tests := []struct {
		format string
		output string
	}{
		{"json", `{"key": "value"}`},
		{"yaml", "key: value"},
		{"markdown", "# Title"},
		{"text", "Plain text"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			tmpl.Format = tt.format
			err := validator.ValidateOutput(tmpl, tt.output)
			assert.NoError(t, err)
		})
	}
}

func TestValidationError_Structure(t *testing.T) {
	err := ValidationError{
		Field:   "name",
		Message: "invalid value",
		Fix:     "provide valid value",
	}

	assert.Equal(t, "name", err.Field)
	assert.Equal(t, "invalid value", err.Message)
	assert.Equal(t, "provide valid value", err.Fix)
}

func TestValidationResult_Structure(t *testing.T) {
	result := ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
}

func TestCheckBalancedDelimiters_NestedBlocks(t *testing.T) {
	validator := NewTemplateValidator()

	content := `
	{{#each users}}
		{{#if active}}
			{{name}} - {{email}}
		{{/if}}
	{{/each}}
	`

	err := validator.checkBalancedDelimiters(content)
	assert.NoError(t, err)
}

func TestCheckBalancedDelimiters_ComplexNesting(t *testing.T) {
	validator := NewTemplateValidator()

	content := `
	{{#if showUsers}}
		{{#each users}}
			{{#unless deleted}}
				{{#with profile}}
					{{firstName}} {{lastName}}
				{{/with}}
			{{/unless}}
		{{/each}}
	{{/if}}
	`

	err := validator.checkBalancedDelimiters(content)
	assert.NoError(t, err)
}

func TestValidateComprehensive_MultipleErrors(t *testing.T) {
	validator := NewTemplateValidator()
	tmpl := domain.NewTemplate("test", "Test template", "1.0", "author")
	tmpl.Content = "{{123}} {{#if}}"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "123", Type: "string", Required: true},
		{Name: "valid_var", Type: "number", Required: true},
	}

	variables := map[string]interface{}{
		"valid_var": "not a number",
	}

	result := validator.ValidateComprehensive(tmpl, variables)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	// Should have multiple errors: invalid var name, wrong type, missing required
	assert.GreaterOrEqual(t, len(result.Errors), 2)
}
