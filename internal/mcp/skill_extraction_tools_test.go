package mcp

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleExtractSkillsFromPersona(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create test persona with expertise areas
	persona := domain.NewPersona("Test Persona", "A test persona for skill extraction", "1.0.0", "test-user")
	persona.ExpertiseAreas = []domain.ExpertiseArea{
		{
			Domain:   "Python Programming",
			Level:    "expert",
			Keywords: []string{"python", "pandas", "numpy"},
		},
		{
			Domain:   "Machine Learning",
			Level:    "advanced",
			Keywords: []string{"tensorflow", "pytorch", "scikit-learn"},
		},
	}
	persona.SystemPrompt = "You are an expert in Python and machine learning."
	require.NoError(t, server.repo.Create(persona))

	tests := []struct {
		name      string
		input     ExtractSkillsFromPersonaInput
		wantErr   bool
		errString string
	}{
		{
			name: "extract_from_valid_persona",
			input: ExtractSkillsFromPersonaInput{
				PersonaID: persona.GetID(),
			},
			wantErr: false,
		},
		{
			name: "missing_persona_id",
			input: ExtractSkillsFromPersonaInput{
				PersonaID: "",
			},
			wantErr:   true,
			errString: "persona_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleExtractSkillsFromPersona(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, output.SkillsCreated, 0)
				assert.NotNil(t, output.SkillIDs)
				assert.NotEmpty(t, output.Message)
			}
		})
	}
}

func TestHandleBatchExtractSkills(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Create test personas
	persona1 := domain.NewPersona("Persona 1", "First test persona", "1.0.0", "test-user")
	persona1.ExpertiseAreas = []domain.ExpertiseArea{
		{Domain: "JavaScript", Level: "expert", Keywords: []string{"react", "node"}},
	}
	require.NoError(t, server.repo.Create(persona1))

	persona2 := domain.NewPersona("Persona 2", "Second test persona", "1.0.0", "test-user")
	persona2.ExpertiseAreas = []domain.ExpertiseArea{
		{Domain: "DevOps", Level: "advanced", Keywords: []string{"docker", "kubernetes"}},
	}
	require.NoError(t, server.repo.Create(persona2))

	tests := []struct {
		name      string
		input     BatchExtractSkillsInput
		wantErr   bool
		errString string
	}{
		{
			name: "extract_from_multiple_personas",
			input: BatchExtractSkillsInput{
				PersonaIDs: []string{persona1.GetID(), persona2.GetID()},
			},
			wantErr: false,
		},
		{
			name: "missing_persona_ids",
			input: BatchExtractSkillsInput{
				PersonaIDs: []string{},
			},
			// Should extract from all personas (not an error)
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := server.handleBatchExtractSkills(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, output.TotalSkillsCreated, 0)
				assert.GreaterOrEqual(t, output.TotalPersonasProcessed, 0)
				assert.NotNil(t, output.Results)
				assert.NotEmpty(t, output.Message)
			}
		})
	}
}
