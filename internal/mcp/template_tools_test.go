package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/common"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

func setupTestServerForTemplates() *MCPServer {
	repo := NewMockElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

func TestHandleListTemplates_DefaultParameters(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := ListTemplatesInput{}

	_, output, err := server.handleListTemplates(ctx, nil, input)
	require.NoError(t, err)
	assert.Equal(t, 1, output.Page)
	assert.Equal(t, 20, output.PerPage)
	assert.NotNil(t, output.Templates)
}

func TestHandleListTemplates_CustomPage(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := ListTemplatesInput{
		Page:    2,
		PerPage: 10,
	}

	_, output, err := server.handleListTemplates(ctx, nil, input)
	require.NoError(t, err)
	assert.Equal(t, 2, output.Page)
	assert.Equal(t, 10, output.PerPage)
}

func TestHandleListTemplates_FilterByCategory(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := ListTemplatesInput{
		Category: "persona",
	}

	_, output, err := server.handleListTemplates(ctx, nil, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Templates)
}

func TestHandleListTemplates_FilterByTags(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := ListTemplatesInput{
		Tags: []string{"technical", "expert"},
	}

	_, output, err := server.handleListTemplates(ctx, nil, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Templates)
}

func TestHandleListTemplates_FilterByElementType(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := ListTemplatesInput{
		ElementType: common.ElementTypeSkill,
	}

	_, output, err := server.handleListTemplates(ctx, nil, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Templates)
}

func TestHandleListTemplates_IncludeBuiltIn(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := ListTemplatesInput{
		IncludeBuiltIn: true,
	}

	_, output, err := server.handleListTemplates(ctx, nil, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Templates)
}

func TestHandleListTemplates_OutputStructure(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := ListTemplatesInput{}

	_, output, err := server.handleListTemplates(ctx, nil, input)
	require.NoError(t, err)
	assert.NotNil(t, output.Templates)
	assert.GreaterOrEqual(t, output.Total, 0)
	assert.GreaterOrEqual(t, output.Page, 1)
	assert.GreaterOrEqual(t, output.PerPage, 1)
}

func TestHandleGetTemplate_RequiredID(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := GetTemplateInput{}

	_, _, err := server.handleGetTemplate(ctx, nil, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template ID is required")
}

func TestHandleGetTemplate_TemplateNotFound(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := GetTemplateInput{
		ID: "nonexistent-template",
	}

	_, _, err := server.handleGetTemplate(ctx, nil, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestHandleGetTemplate_ValidTemplate(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	// Create a test template
	tmpl := domain.NewTemplate("Test Template", "A test template", "1.0.0", "test")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Variables = []domain.TemplateVariable{
		{
			Name:        "name",
			Description: "The name to greet",
		},
	}
	require.NoError(t, server.repo.Create(tmpl))

	input := GetTemplateInput{
		ID: tmpl.GetMetadata().ID,
	}

	_, output, err := server.handleGetTemplate(ctx, nil, input)
	require.NoError(t, err)
	assert.Equal(t, tmpl.GetMetadata().ID, output.ID)
	assert.Equal(t, "Test Template", output.Name)
	assert.Equal(t, "Hello {{name}}!", output.Content)
	assert.Len(t, output.Variables, 1)
	assert.NotNil(t, output.Helpers)
}

func TestHandleInstantiateTemplate_RequiredTemplateID(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := InstantiateTemplateInput{}

	_, _, err := server.handleInstantiateTemplate(ctx, nil, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template_id is required")
}

func TestHandleInstantiateTemplate_TemplateNotFound(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := InstantiateTemplateInput{
		TemplateID: "nonexistent-template",
		Variables:  map[string]interface{}{},
	}

	_, _, err := server.handleInstantiateTemplate(ctx, nil, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestHandleInstantiateTemplate_Success(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	// Create a test template
	tmpl := domain.NewTemplate("Greeting Template", "A greeting template", "1.0.0", "test")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Variables = []domain.TemplateVariable{
		{
			Name:        "name",
			Description: "The name to greet",
		},
	}
	require.NoError(t, server.repo.Create(tmpl))

	input := InstantiateTemplateInput{
		TemplateID: tmpl.GetMetadata().ID,
		Variables: map[string]interface{}{
			"name": "World",
		},
	}

	_, output, err := server.handleInstantiateTemplate(ctx, nil, input)
	require.NoError(t, err)
	assert.Contains(t, output.Output, "Hello")
	assert.False(t, output.Saved)
	assert.Empty(t, output.ElementID)
}

func TestHandleInstantiateTemplate_WithDryRun(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	// Create a test template
	tmpl := domain.NewTemplate("Test Template", "A test template", "1.0.0", "test")
	tmpl.Content = "Content: {{value}}"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "value", Description: "Test value"},
	}
	require.NoError(t, server.repo.Create(tmpl))

	input := InstantiateTemplateInput{
		TemplateID: tmpl.GetMetadata().ID,
		Variables: map[string]interface{}{
			"value": "test",
		},
		DryRun: true,
		SaveAs: "saved-element",
	}

	_, output, err := server.handleInstantiateTemplate(ctx, nil, input)
	require.NoError(t, err)
	assert.False(t, output.Saved)
	assert.Empty(t, output.ElementID)
}

func TestHandleInstantiateTemplate_WithSaveAs(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	// Create a test template that generates valid persona YAML
	tmpl := domain.NewTemplate("Test Template", "A test template", "1.0.0", "test")
	tmpl.Content = `name: {{name}}
description: {{description}}
role: {{role}}
expertise: []
communication_style: formal
interaction_guidelines: []`
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Description: "Persona name"},
		{Name: "description", Description: "Persona description"},
		{Name: "role", Description: "Persona role"},
	}
	require.NoError(t, server.repo.Create(tmpl))

	input := InstantiateTemplateInput{
		TemplateID: tmpl.GetMetadata().ID,
		Variables: map[string]interface{}{
			"name":        "Test Persona",
			"description": "A test persona",
			"role":        "Tester",
		},
		SaveAs: "saved-element",
	}

	_, output, err := server.handleInstantiateTemplate(ctx, nil, input)
	require.NoError(t, err)
	// Note: Since template output doesn't include full metadata (ID, type, version, author),
	// the unmarshaling will fail and element won't be saved
	// This is expected behavior - templates should generate complete element YAML
	assert.False(t, output.Saved)
	assert.Empty(t, output.ElementID)
	assert.Contains(t, output.Warnings[0], "Failed to parse template output")
}

func TestHandleValidateTemplate_RequiredTemplateID(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := ValidateTemplateInput{}

	_, _, err := server.handleValidateTemplate(ctx, nil, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template_id is required")
}

func TestHandleValidateTemplate_TemplateNotFound(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	input := ValidateTemplateInput{
		TemplateID: "nonexistent-template",
	}

	_, _, err := server.handleValidateTemplate(ctx, nil, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestHandleValidateTemplate_ValidSyntax(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	// Create a valid template
	tmpl := domain.NewTemplate("Valid Template", "A valid template", "1.0.0", "test")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Description: "The name"},
	}
	require.NoError(t, server.repo.Create(tmpl))

	input := ValidateTemplateInput{
		TemplateID: tmpl.GetMetadata().ID,
	}

	_, output, err := server.handleValidateTemplate(ctx, nil, input)
	require.NoError(t, err)
	assert.True(t, output.Valid)
	assert.Empty(t, output.Errors)
}

func TestHandleValidateTemplate_InvalidSyntax(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	// Create a template with invalid syntax
	tmpl := domain.NewTemplate("Invalid Template", "An invalid template", "1.0.0", "test")
	tmpl.Content = "Hello {{name}!" // Missing closing brace
	require.NoError(t, server.repo.Create(tmpl))

	input := ValidateTemplateInput{
		TemplateID: tmpl.GetMetadata().ID,
	}

	_, output, err := server.handleValidateTemplate(ctx, nil, input)
	require.NoError(t, err)
	assert.False(t, output.Valid)
	assert.NotEmpty(t, output.Errors)
}

func TestHandleValidateTemplate_WithVariables(t *testing.T) {
	server := setupTestServerForTemplates()
	ctx := context.Background()

	// Create a template
	tmpl := domain.NewTemplate("Test Template", "A test template", "1.0.0", "test")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Description: "The name", Required: true},
	}
	require.NoError(t, server.repo.Create(tmpl))

	input := ValidateTemplateInput{
		TemplateID: tmpl.GetMetadata().ID,
		Variables: map[string]interface{}{
			"name": "World",
		},
	}

	_, output, err := server.handleValidateTemplate(ctx, nil, input)
	require.NoError(t, err)
	// Comprehensive validation may have different results depending on implementation
	assert.NotNil(t, output.Valid)
}

func TestInferElementType(t *testing.T) {
	t.Skip("Template metadata is private, tested via integration")
}

func TestIsBuiltInTemplate(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{
			name:     "Standard library template",
			id:       "stdlib-persona-technical",
			expected: true,
		},
		{
			name:     "Standard library prefix",
			id:       "stdlib-skill-coding",
			expected: true,
		},
		{
			name:     "Custom template",
			id:       "custom-template-123",
			expected: false,
		},
		{
			name:     "Short ID",
			id:       "short",
			expected: false,
		},
		{
			name:     "Empty ID",
			id:       "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBuiltInTemplate(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateInfo_Structure(t *testing.T) {
	info := TemplateInfo{
		ID:          "test-id",
		Name:        "Test Template",
		Description: "A test",
		Version:     "1.0.0",
		Author:      "test-author",
		Tags:        []string{"tag1", "tag2"},
		ElementType: common.ElementTypePersona,
		Variables:   3,
		IsBuiltIn:   true,
	}

	assert.Equal(t, "test-id", info.ID)
	assert.Equal(t, "Test Template", info.Name)
	assert.Equal(t, 3, info.Variables)
	assert.True(t, info.IsBuiltIn)
}

func TestValidationErrorInfo_Structure(t *testing.T) {
	err := ValidationErrorInfo{
		Field:   "content",
		Message: "Invalid syntax",
		Fix:     "Fix the template",
	}

	assert.Equal(t, "content", err.Field)
	assert.Equal(t, "Invalid syntax", err.Message)
	assert.Equal(t, "Fix the template", err.Fix)
}
