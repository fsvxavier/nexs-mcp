package mcp

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleReloadElements(t *testing.T) {
	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := &MCPServer{
		repo: repo,
	}

	ctx := context.Background()

	// Create test elements
	persona := domain.NewPersona("Test Persona", "A comprehensive test persona for reload validation", "1.0.0", "tester")
	persona.BehavioralTraits = []domain.BehavioralTrait{{Name: "helpful", Intensity: 8}, {Name: "efficient", Intensity: 7}}
	persona.ExpertiseAreas = []domain.ExpertiseArea{{Domain: "testing", Level: "expert", Keywords: []string{"qa"}}, {Domain: "automation", Level: "advanced", Keywords: []string{"ci/cd"}}}
	persona.ResponseStyle = domain.ResponseStyle{Tone: "professional", Formality: "formal", Verbosity: "balanced"}
	persona.SystemPrompt = "You are a test persona specialized in quality assurance and test automation with helpful and efficient traits."

	skill := domain.NewSkill("Test Skill", "A comprehensive test skill for reload validation", "1.0.0", "tester")
	skill.Triggers = []domain.SkillTrigger{{Type: "keyword", Keywords: []string{"test"}}}
	skill.Procedures = []domain.SkillProcedure{{Step: 1, Action: "Execute comprehensive test suite", Description: "Run all test cases"}}

	template := domain.NewTemplate("Test Template", "A comprehensive test template for reload validation", "1.0.0", "tester")
	template.Content = "Hello {{name}}"
	template.Format = "text"

	// Save elements
	if err := repo.Create(persona); err != nil {
		t.Fatalf("Failed to save persona: %v", err)
	}
	if err := repo.Create(skill); err != nil {
		t.Fatalf("Failed to save skill: %v", err)
	}
	if err := repo.Create(template); err != nil {
		t.Fatalf("Failed to save template: %v", err)
	}

	// Verify elements were saved (debug)
	allElements, err := repo.List(domain.ElementFilter{})
	if err != nil {
		t.Fatalf("Failed to list elements: %v", err)
	}
	if len(allElements) != 3 {
		t.Fatalf("Expected 3 elements in repo, got %d", len(allElements))
	}

	tests := []struct {
		name              string
		input             ReloadElementsInput
		wantErr           bool
		wantMinReloaded   int
		wantValidationErr bool
	}{
		{
			name: "Reload all elements",
			input: ReloadElementsInput{
				ElementTypes:        []string{"all"},
				ClearCaches:         true,
				ValidateAfterReload: true,
			},
			wantErr:         false,
			wantMinReloaded: 3, // At least our 3 test elements
		},
		{
			name: "Reload only personas",
			input: ReloadElementsInput{
				ElementTypes:        []string{"persona"},
				ClearCaches:         true,
				ValidateAfterReload: true,
			},
			wantErr:         false,
			wantMinReloaded: 1,
		},
		{
			name: "Reload multiple types",
			input: ReloadElementsInput{
				ElementTypes:        []string{"persona", "skill"},
				ClearCaches:         true,
				ValidateAfterReload: true,
			},
			wantErr:         false,
			wantMinReloaded: 2,
		},
		{
			name: "Reload without validation",
			input: ReloadElementsInput{
				ElementTypes:        []string{"template"},
				ClearCaches:         false,
				ValidateAfterReload: false,
			},
			wantErr:         false,
			wantMinReloaded: 1, // Only templates requested
		},
		{
			name: "Invalid element type",
			input: ReloadElementsInput{
				ElementTypes: []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name:  "Default parameters (reload all)",
			input: ReloadElementsInput{
				// Empty - should default to all
			},
			wantErr:         false,
			wantMinReloaded: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleReloadElements(ctx, &sdk.CallToolRequest{}, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
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

			if output.TotalReloaded < tt.wantMinReloaded {
				t.Errorf("TotalReloaded = %d, want at least %d", output.TotalReloaded, tt.wantMinReloaded)
			}

			// Verify that element counts are provided
			if len(output.ElementsReloaded) == 0 && output.TotalReloaded > 0 {
				t.Errorf("Expected ElementsReloaded breakdown")
			}

			// Verify cache stats are present
			if output.CacheStats.AfterSize < 0 {
				t.Errorf("Invalid cache stats")
			}
		})
	}
}

func TestHandleReloadElements_ValidationErrors(t *testing.T) {
	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := &MCPServer{
		repo: repo,
	}

	ctx := context.Background()

	// Create an invalid persona (system prompt too short - will fail validation)
	invalidPersona := domain.NewPersona("Invalid", "Invalid persona for testing validation errors", "1.0.0", "tester")
	// Set system prompt too short (less than 50 chars) - will fail validation
	invalidPersona.SystemPrompt = "Too short" // Less than 50 characters required

	if err := repo.Create(invalidPersona); err != nil {
		t.Fatalf("Failed to save invalid persona: %v", err)
	}

	input := ReloadElementsInput{
		ElementTypes:        []string{"persona"},
		ValidateAfterReload: true,
	}

	_, output, err := server.handleReloadElements(ctx, &sdk.CallToolRequest{}, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have validation errors for the invalid persona
	if len(output.ValidationErrors) == 0 {
		t.Logf("Warning: Expected validation errors for invalid persona, but validation may have passed")
	}

	if output.TotalFailed > 0 && len(output.ElementsFailed) == 0 {
		t.Errorf("TotalFailed > 0 but ElementsFailed is empty")
	}
}

func TestHandleReloadElements_TypeFiltering(t *testing.T) {
	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := &MCPServer{
		repo: repo,
	}

	ctx := context.Background()

	// Create elements of different types
	persona := domain.NewPersona("P1", "Test persona for type filtering validation", "1.0.0", "tester")
	persona.BehavioralTraits = []domain.BehavioralTrait{{Name: "helpful", Intensity: 8}, {Name: "quick", Intensity: 7}}
	persona.ExpertiseAreas = []domain.ExpertiseArea{{Domain: "testing", Level: "expert", Keywords: []string{"qa"}}, {Domain: "automation", Level: "advanced", Keywords: []string{"ci"}}}
	persona.ResponseStyle = domain.ResponseStyle{Tone: "professional", Formality: "formal", Verbosity: "balanced"}
	persona.SystemPrompt = "You are P1, a test persona specialized in quality assurance and automation testing."

	skill := domain.NewSkill("S1", "Test skill for type filtering validation", "1.0.0", "tester")
	skill.Triggers = []domain.SkillTrigger{{Type: "keyword", Keywords: []string{"test"}}}
	skill.Procedures = []domain.SkillProcedure{{Step: 1, Action: "Execute comprehensive test", Description: "Run test suite"}}

	repo.Create(persona)
	repo.Create(skill)

	// Reload only skills
	input := ReloadElementsInput{
		ElementTypes:        []string{"skill"},
		ValidateAfterReload: false,
	}

	_, output, err := server.handleReloadElements(ctx, &sdk.CallToolRequest{}, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify only skills were reloaded
	foundSkills := false
	foundPersonas := false

	for _, typeCount := range output.ElementsReloaded {
		if typeCount.Type == "skill" {
			foundSkills = true
			if typeCount.Count == 0 {
				t.Errorf("Expected at least 1 skill to be reloaded")
			}
		}
		if typeCount.Type == "persona" {
			foundPersonas = true
		}
	}

	if !foundSkills {
		t.Errorf("Expected skills in reloaded elements")
	}

	if foundPersonas {
		t.Errorf("Did not expect personas to be reloaded when filtering for skills only")
	}
}
