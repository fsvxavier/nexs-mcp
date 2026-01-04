package mcp

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleRenderTemplate(t *testing.T) {
	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := newTestServer("test-server", "1.0.0", repo)

	ctx := context.Background()

	// Create test template
	tmpl := domain.NewTemplate("Greeting Template", "A simple greeting template", "1.0.0", "tester")
	tmpl.Content = "Hello {{name}}! Welcome to {{place}}."
	tmpl.Format = "text"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
		{Name: "place", Type: "string", Required: true},
	}

	err := repo.Create(tmpl)
	if err != nil {
		t.Fatalf("Failed to save test template: %v", err)
	}

	tests := []struct {
		name            string
		input           RenderTemplateInput
		wantErr         bool
		wantOutput      string
		wantMissingVars bool
	}{
		{
			name: "Render with template_id",
			input: RenderTemplateInput{
				TemplateID: tmpl.GetID(),
				Data: map[string]interface{}{
					"name":  "Alice",
					"place": "Wonderland",
				},
				OutputFormat:         "text",
				ValidateBeforeRender: true,
			},
			wantErr:    false,
			wantOutput: "Hello Alice! Welcome to Wonderland.",
		},
		{
			name: "Render with template_content",
			input: RenderTemplateInput{
				TemplateContent: "Hi {{user}}, you have {{count}} messages.",
				Data: map[string]interface{}{
					"user":  "Bob",
					"count": 5,
				},
				OutputFormat:         "text",
				ValidateBeforeRender: false,
			},
			wantErr:    false,
			wantOutput: "Hi Bob, you have 5 messages.",
		},
		{
			name: "Missing required field - no template_id or content",
			input: RenderTemplateInput{
				Data: map[string]interface{}{
					"name": "Test",
				},
			},
			wantErr: true,
		},
		{
			name: "Both template_id and content provided",
			input: RenderTemplateInput{
				TemplateID:      tmpl.GetID(),
				TemplateContent: "Test",
				Data:            map[string]interface{}{"name": "Test"},
			},
			wantErr: true,
		},
		{
			name: "No data provided",
			input: RenderTemplateInput{
				TemplateID: tmpl.GetID(),
			},
			wantErr: true,
		},
		{
			name: "Invalid output format",
			input: RenderTemplateInput{
				TemplateContent: "Test",
				Data:            map[string]interface{}{"name": "Test"},
				OutputFormat:    "invalid",
			},
			wantErr: true,
		},
		{
			name: "Template not found",
			input: RenderTemplateInput{
				TemplateID: "non-existent-id",
				Data:       map[string]interface{}{"name": "Test"},
			},
			wantErr: true,
		},
		{
			name: "Missing required variables with validation",
			input: RenderTemplateInput{
				TemplateID: tmpl.GetID(),
				Data: map[string]interface{}{
					"name": "Alice",
					// "place" is missing
				},
				ValidateBeforeRender: true,
			},
			wantErr:         true,
			wantMissingVars: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleRenderTemplate(ctx, &sdk.CallToolRequest{}, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.wantMissingVars && len(output.MissingVariables) == 0 {
					t.Errorf("Expected missing variables in output")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != nil {
				t.Errorf("Expected nil CallToolResult, got %v", result)
			}

			if output.RenderedOutput != tt.wantOutput {
				t.Errorf("RenderedOutput = %q, want %q", output.RenderedOutput, tt.wantOutput)
			}

			if output.RenderTimeMs < 0 {
				t.Errorf("Expected non-negative RenderTimeMs, got %d", output.RenderTimeMs)
			}

			if output.Format == "" {
				t.Errorf("Expected Format to be set")
			}
		})
	}
}

func TestHandleRenderTemplate_WithHelpers(t *testing.T) {
	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := newTestServer("test-server", "1.0.0", repo)

	ctx := context.Background()

	input := RenderTemplateInput{
		TemplateContent: "{{upper name}}",
		Data: map[string]interface{}{
			"name": "alice",
		},
		OutputFormat: "text",
	}

	_, output, err := server.handleRenderTemplate(ctx, &sdk.CallToolRequest{}, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output.RenderedOutput != "ALICE" {
		t.Errorf("RenderedOutput = %q, want %q", output.RenderedOutput, "ALICE")
	}

	// Verify that helpers are tracked
	if len(output.VariablesUsed) == 0 {
		t.Errorf("Expected variables to be tracked")
	}
}

func TestHandleRenderTemplate_OutputFormats(t *testing.T) {
	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := newTestServer("test-server", "1.0.0", repo)

	ctx := context.Background()

	formats := []string{"text", "markdown", "yaml", "json"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			input := RenderTemplateInput{
				TemplateContent: "Test content",
				Data:            map[string]interface{}{"key": "value"},
				OutputFormat:    format,
			}

			_, output, err := server.handleRenderTemplate(ctx, &sdk.CallToolRequest{}, input)
			if err != nil {
				t.Errorf("Unexpected error for format %s: %v", format, err)
			}

			if output.Format != format {
				t.Errorf("Format = %s, want %s", output.Format, format)
			}
		})
	}
}
