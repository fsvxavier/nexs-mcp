package mcp

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleValidateElement(t *testing.T) {
	// Setup test repository
	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	// Create test server
	server := &MCPServer{
		repo: repo,
	}

	ctx := context.Background()

	// Create test persona with all required fields for validation
	persona := domain.NewPersona("Test Persona", "A comprehensive test persona for validation testing", "1.0.0", "tester")
	persona.BehavioralTraits = []domain.BehavioralTrait{
		{Name: "helpful", Intensity: 8},
		{Name: "analytical", Intensity: 7},
	}
	persona.ExpertiseAreas = []domain.ExpertiseArea{
		{Domain: "testing", Level: "expert", Keywords: []string{"unit tests", "integration tests"}},
		{Domain: "validation", Level: "advanced", Keywords: []string{"quality assurance"}},
	}
	persona.ResponseStyle = domain.ResponseStyle{
		Tone:      "professional",
		Formality: "formal",
		Verbosity: "balanced",
	}
	persona.SystemPrompt = "You are Test Persona, an expert in testing and validation with a helpful and analytical attitude. You specialize in quality assurance and comprehensive testing strategies."

	err := repo.Create(persona)
	if err != nil {
		t.Fatalf("Failed to save test persona: %v", err)
	}

	tests := []struct {
		name           string
		input          ValidateElementInput
		wantErr        bool
		wantValid      bool
		wantErrorCount int
	}{
		{
			name: "Valid persona - comprehensive level",
			input: ValidateElementInput{
				ElementID:       persona.GetID(),
				ElementType:     "persona",
				ValidationLevel: "comprehensive",
				FixSuggestions:  true,
			},
			wantErr:        false,
			wantValid:      true,
			wantErrorCount: 0,
		},
		{
			name: "Valid persona - basic level",
			input: ValidateElementInput{
				ElementID:       persona.GetID(),
				ElementType:     "persona",
				ValidationLevel: "basic",
				FixSuggestions:  false,
			},
			wantErr:        false,
			wantValid:      true,
			wantErrorCount: 0,
		},
		{
			name: "Invalid element ID",
			input: ValidateElementInput{
				ElementID:   "non-existent-id",
				ElementType: "persona",
			},
			wantErr: true,
		},
		{
			name: "Missing element ID",
			input: ValidateElementInput{
				ElementType: "persona",
			},
			wantErr: true,
		},
		{
			name: "Invalid element type",
			input: ValidateElementInput{
				ElementID:   persona.GetID(),
				ElementType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "Invalid validation level",
			input: ValidateElementInput{
				ElementID:       persona.GetID(),
				ElementType:     "persona",
				ValidationLevel: "invalid",
			},
			wantErr: true,
		},
		{
			name: "Type mismatch",
			input: ValidateElementInput{
				ElementID:   persona.GetID(),
				ElementType: "skill",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleValidateElement(ctx, &sdk.CallToolRequest{}, tt.input)

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

			if output.IsValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", output.IsValid, tt.wantValid)
				// Log actual errors for debugging
				for _, err := range output.Errors {
					t.Logf("Validation Error: Field=%s, Code=%s, Message=%s", err.Field, err.Code, err.Message)
				}
			}

			if output.ErrorCount != tt.wantErrorCount {
				t.Errorf("ErrorCount = %d, want %d", output.ErrorCount, tt.wantErrorCount)
				// Log actual errors for debugging
				for _, err := range output.Errors {
					t.Logf("Validation Error: Field=%s, Code=%s, Message=%s", err.Field, err.Code, err.Message)
				}
			}

			if output.ElementID != tt.input.ElementID {
				t.Errorf("ElementID = %s, want %s", output.ElementID, tt.input.ElementID)
			}

			if output.ElementType != tt.input.ElementType {
				t.Errorf("ElementType = %s, want %s", output.ElementType, tt.input.ElementType)
			}

			// Verify suggestions are removed when not requested
			if !tt.input.FixSuggestions {
				for _, issue := range output.Errors {
					if issue.Suggestion != "" {
						t.Errorf("Expected no suggestions, but found: %s", issue.Suggestion)
					}
				}
			}
		})
	}
}

func TestHandleValidateElement_StrictLevel(t *testing.T) {
	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := &MCPServer{
		repo: repo,
	}

	ctx := context.Background()

	// Create persona without tags (should fail strict validation)
	persona := domain.NewPersona("Test Persona", "A comprehensive test persona for strict validation", "1.0.0", "tester")
	persona.BehavioralTraits = []domain.BehavioralTrait{
		{Name: "helpful", Intensity: 8},
		{Name: "precise", Intensity: 9},
	}
	persona.ExpertiseAreas = []domain.ExpertiseArea{
		{Domain: "testing", Level: "expert", Keywords: []string{"qa", "validation"}},
		{Domain: "analysis", Level: "advanced", Keywords: []string{"metrics"}},
	}
	persona.ResponseStyle = domain.ResponseStyle{
		Tone:      "professional",
		Formality: "formal",
		Verbosity: "balanced",
	}
	persona.SystemPrompt = "You are Test Persona, an expert in testing and analysis with precision and helpfulness as core traits."
	// No tags set - should fail strict validation

	err := repo.Create(persona)
	if err != nil {
		t.Fatalf("Failed to save test persona: %v", err)
	}

	input := ValidateElementInput{
		ElementID:       persona.GetID(),
		ElementType:     "persona",
		ValidationLevel: "strict",
		FixSuggestions:  true,
	}

	_, output, err := server.handleValidateElement(ctx, &sdk.CallToolRequest{}, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Strict validation should fail without tags
	if output.IsValid {
		t.Errorf("Expected validation to fail in strict mode without tags")
	}

	if output.ErrorCount == 0 && output.WarningCount == 0 {
		t.Errorf("Expected errors or warnings in strict mode")
	}
}

// Helper function to setup test repository.
func setupTestRepository(t *testing.T) domain.ElementRepository {
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewEnhancedFileElementRepository(tmpDir, 100)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	return repo
}

// Helper function to cleanup test repository.
func cleanupTestRepository(t *testing.T, repo domain.ElementRepository) {
	// TempDir is automatically cleaned up by testing framework
}
