package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/template"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RenderTemplateInput represents the input for render_template tool.
type RenderTemplateInput struct {
	TemplateID           string                 `json:"template_id,omitempty"`
	TemplateContent      string                 `json:"template_content,omitempty"`
	Data                 map[string]interface{} `json:"data"`
	OutputFormat         string                 `json:"output_format,omitempty"`
	ValidateBeforeRender bool                   `json:"validate_before_render,omitempty"`
}

// RenderTemplateOutput represents the output of render_template tool.
type RenderTemplateOutput struct {
	RenderedOutput   string   `json:"rendered_output"`
	VariablesUsed    []string `json:"variables_used,omitempty"`
	MissingVariables []string `json:"missing_variables,omitempty"`
	RenderTimeMs     int64    `json:"render_time_ms"`
	Warnings         []string `json:"warnings,omitempty"`
	Format           string   `json:"format"`
}

// handleRenderTemplate handles render_template tool calls.
func (s *MCPServer) handleRenderTemplate(ctx context.Context, req *sdk.CallToolRequest, input RenderTemplateInput) (*sdk.CallToolResult, RenderTemplateOutput, error) {
	startTime := time.Now()

	// Validate input: must have either template_id or template_content
	if input.TemplateID == "" && input.TemplateContent == "" {
		return nil, RenderTemplateOutput{}, errors.New("either template_id or template_content is required")
	}

	if input.TemplateID != "" && input.TemplateContent != "" {
		return nil, RenderTemplateOutput{}, errors.New("cannot specify both template_id and template_content")
	}

	// Validate data is provided
	if input.Data == nil {
		return nil, RenderTemplateOutput{}, errors.New("data is required for template rendering")
	}

	// Set defaults
	if input.OutputFormat == "" {
		input.OutputFormat = "text"
	}

	// Validate output format
	validFormats := map[string]bool{
		"text":     true,
		"markdown": true,
		"yaml":     true,
		"json":     true,
	}
	if !validFormats[input.OutputFormat] {
		return nil, RenderTemplateOutput{}, fmt.Errorf("invalid output_format: %s (must be text, markdown, yaml, or json)", input.OutputFormat)
	}

	var tmpl *domain.Template
	var err error

	// Mode 1: Load template from repository by ID
	if input.TemplateID != "" {
		// Create registry
		registry := template.NewTemplateRegistry(s.repo, 0)

		// Load standard library
		if err := registry.LoadStandardLibrary(); err != nil {
			return nil, RenderTemplateOutput{}, fmt.Errorf("failed to load standard library: %w", err)
		}

		// Get template
		tmpl, err = registry.GetTemplate(ctx, input.TemplateID)
		if err != nil {
			return nil, RenderTemplateOutput{}, fmt.Errorf("template not found: %w", err)
		}
	} else {
		// Mode 2: Create template from provided content
		tmpl = &domain.Template{
			Content: input.TemplateContent,
			Format:  input.OutputFormat,
		}
	}

	// Pre-render validation if enabled
	var warnings []string
	if input.ValidateBeforeRender {
		validator := template.NewTemplateValidator()
		validationErr := validator.ValidateSyntax(tmpl)
		if validationErr != nil {
			return nil, RenderTemplateOutput{}, fmt.Errorf("template validation failed: %w", validationErr)
		}

		// Check for missing required variables
		if tmpl.Variables != nil {
			missingVars := []string{}
			for _, variable := range tmpl.Variables {
				if variable.Required {
					if _, exists := input.Data[variable.Name]; !exists {
						missingVars = append(missingVars, variable.Name)
					}
				}
			}
			if len(missingVars) > 0 {
				return nil, RenderTemplateOutput{
					MissingVariables: missingVars,
					Warnings:         []string{fmt.Sprintf("Missing required variables: %v", missingVars)},
				}, fmt.Errorf("missing required variables: %v", missingVars)
			}
		}
	}

	// Create instantiation engine
	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	// Render template
	result, err := engine.Instantiate(tmpl, input.Data)
	if err != nil {
		return nil, RenderTemplateOutput{}, fmt.Errorf("template rendering failed: %w", err)
	}

	// Add result warnings to our warnings
	warnings = append(warnings, result.Warnings...)

	// Calculate render time
	renderTime := time.Since(startTime).Milliseconds()

	// Extract variables used (from result.Variables map keys)
	variablesUsed := []string{}
	for varName := range result.Variables {
		variablesUsed = append(variablesUsed, varName)
	}

	output := RenderTemplateOutput{
		RenderedOutput:   result.Output,
		VariablesUsed:    variablesUsed,
		MissingVariables: []string{}, // Empty if we got here
		RenderTimeMs:     renderTime,
		Warnings:         warnings,
		Format:           input.OutputFormat,
	}

	return nil, output, nil
}
