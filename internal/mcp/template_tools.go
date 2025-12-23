package mcp

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert/yaml"

	"github.com/fsvxavier/nexs-mcp/internal/common"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/template"
)

// Template Tool Definitions

// ListTemplatesInput defines input for list_templates tool.
type ListTemplatesInput struct {
	Category       string   `json:"category,omitempty"     jsonschema:"filter by category (persona, skill, agent, etc.)"`
	Tags           []string `json:"tags,omitempty"         jsonschema:"filter by tags"`
	ElementType    string   `json:"element_type,omitempty" jsonschema:"filter by target element type"`
	IncludeBuiltIn bool     `json:"include_builtin"        jsonschema:"include standard library templates (default: true)"`
	Page           int      `json:"page,omitempty"         jsonschema:"page number (default: 1)"`
	PerPage        int      `json:"per_page,omitempty"     jsonschema:"results per page (default: 20)"`
}

// ListTemplatesOutput defines output for list_templates tool.
type ListTemplatesOutput struct {
	Templates []TemplateInfo `json:"templates"`
	Total     int            `json:"total"`
	Page      int            `json:"page"`
	PerPage   int            `json:"per_page"`
	HasMore   bool           `json:"has_more"`
}

// TemplateInfo contains template metadata.
type TemplateInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	ElementType string   `json:"element_type"`
	Variables   int      `json:"variables"`
	IsBuiltIn   bool     `json:"is_builtin"`
}

// GetTemplateInput defines input for get_template tool.
type GetTemplateInput struct {
	ID string `json:"id" jsonschema:"template ID to retrieve"`
}

// GetTemplateOutput defines output for get_template tool.
type GetTemplateOutput struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Version     string                    `json:"version"`
	Author      string                    `json:"author"`
	Tags        []string                  `json:"tags"`
	Content     string                    `json:"content"`
	Format      string                    `json:"format"`
	Variables   []domain.TemplateVariable `json:"variables"`
	Helpers     []string                  `json:"helpers"`
	IsBuiltIn   bool                      `json:"is_builtin"`
}

// InstantiateTemplateInput defines input for instantiate_template tool.
type InstantiateTemplateInput struct {
	TemplateID string                 `json:"template_id"       jsonschema:"template ID to instantiate"`
	Variables  map[string]interface{} `json:"variables"         jsonschema:"variable values for instantiation"`
	SaveAs     string                 `json:"save_as,omitempty" jsonschema:"save instantiated element with this ID"`
	DryRun     bool                   `json:"dry_run,omitempty" jsonschema:"preview only, don't save (default: false)"`
}

// InstantiateTemplateOutput defines output for instantiate_template tool.
type InstantiateTemplateOutput struct {
	Output      string                 `json:"output"`
	ElementID   string                 `json:"element_id,omitempty"`
	Variables   map[string]interface{} `json:"variables"`
	Warnings    []string               `json:"warnings,omitempty"`
	UsedHelpers []string               `json:"used_helpers,omitempty"`
	Saved       bool                   `json:"saved"`
}

// ValidateTemplateInput defines input for validate_template tool.
type ValidateTemplateInput struct {
	TemplateID string                 `json:"template_id"         jsonschema:"template ID to validate"`
	Variables  map[string]interface{} `json:"variables,omitempty" jsonschema:"test variables (optional)"`
}

// ValidateTemplateOutput defines output for validate_template tool.
type ValidateTemplateOutput struct {
	Valid    bool                  `json:"valid"`
	Errors   []ValidationErrorInfo `json:"errors,omitempty"`
	Warnings []string              `json:"warnings,omitempty"`
}

// ValidationErrorInfo contains validation error details.
type ValidationErrorInfo struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Fix     string `json:"fix"`
}

// Template Tool Handlers

// handleListTemplates handles list_templates tool calls.
func (s *MCPServer) handleListTemplates(ctx context.Context, req *sdk.CallToolRequest, input ListTemplatesInput) (*sdk.CallToolResult, ListTemplatesOutput, error) {
	// Default values
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 {
		input.PerPage = 20
	}
	if !input.IncludeBuiltIn {
		input.IncludeBuiltIn = true // Default: include built-in templates
	}

	// Create registry if not exists
	registry := template.NewTemplateRegistry(s.repo, 0)

	// Load standard library
	if err := registry.LoadStandardLibrary(); err != nil {
		return nil, ListTemplatesOutput{}, fmt.Errorf("failed to load standard library: %w", err)
	}

	// Search templates
	filter := template.TemplateSearchFilter{
		Category:       input.Category,
		Tags:           input.Tags,
		ElementType:    input.ElementType,
		IncludeBuiltIn: input.IncludeBuiltIn,
		Page:           input.Page,
		PerPage:        input.PerPage,
	}

	result, err := registry.SearchTemplates(ctx, filter)
	if err != nil {
		return nil, ListTemplatesOutput{}, fmt.Errorf("template search failed: %w", err)
	}

	// Convert to output format
	templates := make([]TemplateInfo, len(result.Templates))
	for i, tmpl := range result.Templates {
		metadata := tmpl.GetMetadata()
		templates[i] = TemplateInfo{
			ID:          metadata.ID,
			Name:        metadata.Name,
			Description: metadata.Description,
			Version:     metadata.Version,
			Author:      metadata.Author,
			Tags:        metadata.Tags,
			ElementType: inferElementType(tmpl),
			Variables:   len(tmpl.Variables),
			IsBuiltIn:   isBuiltInTemplate(metadata.ID),
		}
	}

	output := ListTemplatesOutput{
		Templates: templates,
		Total:     result.Total,
		Page:      result.Page,
		PerPage:   result.PerPage,
		HasMore:   result.HasMore,
	}

	return nil, output, nil
}

// handleGetTemplate handles get_template tool calls.
func (s *MCPServer) handleGetTemplate(ctx context.Context, req *sdk.CallToolRequest, input GetTemplateInput) (*sdk.CallToolResult, GetTemplateOutput, error) {
	if input.ID == "" {
		return nil, GetTemplateOutput{}, errors.New("template ID is required")
	}

	// Create registry
	registry := template.NewTemplateRegistry(s.repo, 0)

	// Load standard library
	if err := registry.LoadStandardLibrary(); err != nil {
		return nil, GetTemplateOutput{}, fmt.Errorf("failed to load standard library: %w", err)
	}

	// Get template
	tmpl, err := registry.GetTemplate(ctx, input.ID)
	if err != nil {
		return nil, GetTemplateOutput{}, fmt.Errorf("template not found: %w", err)
	}

	// Create engine to get helpers list
	engine := template.NewInstantiationEngine(nil, nil)
	helpers := engine.GetRegisteredHelpers()

	metadata := tmpl.GetMetadata()
	output := GetTemplateOutput{
		ID:          metadata.ID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Version:     metadata.Version,
		Author:      metadata.Author,
		Tags:        metadata.Tags,
		Content:     tmpl.Content,
		Format:      tmpl.Format,
		Variables:   tmpl.Variables,
		Helpers:     helpers,
		IsBuiltIn:   isBuiltInTemplate(metadata.ID),
	}

	return nil, output, nil
}

// handleInstantiateTemplate handles instantiate_template tool calls.
func (s *MCPServer) handleInstantiateTemplate(ctx context.Context, req *sdk.CallToolRequest, input InstantiateTemplateInput) (*sdk.CallToolResult, InstantiateTemplateOutput, error) {
	if input.TemplateID == "" {
		return nil, InstantiateTemplateOutput{}, errors.New("template_id is required")
	}

	// Create registry
	registry := template.NewTemplateRegistry(s.repo, 0)

	// Load standard library
	if err := registry.LoadStandardLibrary(); err != nil {
		return nil, InstantiateTemplateOutput{}, fmt.Errorf("failed to load standard library: %w", err)
	}

	// Get template
	tmpl, err := registry.GetTemplate(ctx, input.TemplateID)
	if err != nil {
		return nil, InstantiateTemplateOutput{}, fmt.Errorf("template not found: %w", err)
	}

	// Create engine
	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	// Instantiate template
	result, err := engine.Instantiate(tmpl, input.Variables)
	if err != nil {
		return nil, InstantiateTemplateOutput{}, fmt.Errorf("instantiation failed: %w", err)
	}

	output := InstantiateTemplateOutput{
		Output:      result.Output,
		Variables:   result.Variables,
		Warnings:    result.Warnings,
		UsedHelpers: result.UsedHelpers,
		Saved:       false,
	}

	// Save element if requested and not dry-run
	if input.SaveAs != "" && !input.DryRun {
		// Try to parse the output as different element types
		// Templates can generate any type of element, so we try all types
		var element domain.Element
		
		// Try each element type in order
		types := []struct {
			name string
			parse func() domain.Element
		}{
			{"persona", func() domain.Element {
				persona := &domain.Persona{}
				if err := yaml.Unmarshal([]byte(result.Output), persona); err == nil && persona.GetID() != "" {
					return persona
				}
				return nil
			}},
			{"skill", func() domain.Element {
				skill := &domain.Skill{}
				if err := yaml.Unmarshal([]byte(result.Output), skill); err == nil && skill.GetID() != "" {
					return skill
				}
				return nil
			}},
			{"agent", func() domain.Element {
				agent := &domain.Agent{}
				if err := yaml.Unmarshal([]byte(result.Output), agent); err == nil && agent.GetID() != "" {
					return agent
				}
				return nil
			}},
			{"memory", func() domain.Element {
				memory := &domain.Memory{}
				if err := yaml.Unmarshal([]byte(result.Output), memory); err == nil && memory.GetID() != "" {
					return memory
				}
				return nil
			}},
			{"ensemble", func() domain.Element {
				ensemble := &domain.Ensemble{}
				if err := yaml.Unmarshal([]byte(result.Output), ensemble); err == nil && ensemble.GetID() != "" {
					return ensemble
				}
				return nil
			}},
		}

		// Try each type until one succeeds
		for _, t := range types {
			if elem := t.parse(); elem != nil {
				element = elem
				break
			}
		}

		// Save if element was created successfully
		if element != nil {
			// Check if element exists, update if it does, create if not
			exists, _ := s.repo.Exists(element.GetID())
			var err error
			if exists {
				err = s.repo.Update(element)
			} else {
				err = s.repo.Create(element)
			}

			if err != nil {
				return nil, output, fmt.Errorf("failed to save element: %w", err)
			}
			output.ElementID = element.GetID()
			output.Saved = true
		} else {
			output.Warnings = append(output.Warnings, "Failed to parse template output into element - ensure template generates valid element YAML")
		}
	}

	return nil, output, nil
}

// handleValidateTemplate handles validate_template tool calls.
func (s *MCPServer) handleValidateTemplate(ctx context.Context, req *sdk.CallToolRequest, input ValidateTemplateInput) (*sdk.CallToolResult, ValidateTemplateOutput, error) {
	if input.TemplateID == "" {
		return nil, ValidateTemplateOutput{}, errors.New("template_id is required")
	}

	// Create registry
	registry := template.NewTemplateRegistry(s.repo, 0)

	// Load standard library
	if err := registry.LoadStandardLibrary(); err != nil {
		return nil, ValidateTemplateOutput{}, fmt.Errorf("failed to load standard library: %w", err)
	}

	// Get template
	tmpl, err := registry.GetTemplate(ctx, input.TemplateID)
	if err != nil {
		return nil, ValidateTemplateOutput{}, fmt.Errorf("template not found: %w", err)
	}

	// Create validator
	validator := template.NewTemplateValidator()

	// If variables provided, do comprehensive validation
	if input.Variables != nil {
		result := validator.ValidateComprehensive(tmpl, input.Variables)

		errors := make([]ValidationErrorInfo, len(result.Errors))
		for i, e := range result.Errors {
			errors[i] = ValidationErrorInfo{
				Field:   e.Field,
				Message: e.Message,
				Fix:     e.Fix,
			}
		}

		output := ValidateTemplateOutput{
			Valid:    result.Valid,
			Errors:   errors,
			Warnings: result.Warnings,
		}

		return nil, output, nil
	}

	// Otherwise, just validate syntax
	if err := validator.ValidateSyntax(tmpl); err != nil {
		output := ValidateTemplateOutput{
			Valid: false,
			Errors: []ValidationErrorInfo{
				{
					Field:   "content",
					Message: err.Error(),
					Fix:     "Fix template syntax errors",
				},
			},
		}
		return nil, output, nil
	}

	output := ValidateTemplateOutput{
		Valid:    true,
		Errors:   []ValidationErrorInfo{},
		Warnings: []string{},
	}

	return nil, output, nil
}

// Helper functions

// inferElementType infers the element type from template.
func inferElementType(tmpl *domain.Template) string {
	metadata := tmpl.GetMetadata()

	// Check tags for hints
	for _, tag := range metadata.Tags {
		switch tag {
		case common.ElementTypePersona, common.ElementTypeSkill, "agent", "memory", "ensemble", common.ElementTypeTemplate:
			return tag
		}
	}

	// Default
	return common.ElementTypeTemplate
}

// isBuiltInTemplate checks if a template ID belongs to standard library.
func isBuiltInTemplate(id string) bool {
	// Standard library templates have specific prefixes
	// This is a simplified check - actual implementation would query the stdlib
	return len(id) > 7 && id[:7] == "stdlib-"
}

// registerTemplateTools registers all template-related MCP tools.
func (s *MCPServer) registerTemplateTools() {
	// Tool 1: list_templates
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "list_templates",
		Description: "List available templates with filtering by category, tags, and element type. Returns paginated results including both built-in and custom templates.",
	}, s.handleListTemplates)

	// Tool 2: get_template
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_template",
		Description: "Retrieve complete template details including content, variables, and metadata. Returns the full template definition.",
	}, s.handleGetTemplate)

	// Tool 3: instantiate_template
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "instantiate_template",
		Description: "Instantiate a template with provided variables. Supports Handlebars syntax with conditionals, loops, and 20+ helpers. Can optionally save the result as a new element.",
	}, s.handleInstantiateTemplate)

	// Tool 4: validate_template
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "validate_template",
		Description: "Validate template syntax and variables. Checks for balanced delimiters, required variables, and proper structure. Optionally test with variable values.",
	}, s.handleValidateTemplate)
}
