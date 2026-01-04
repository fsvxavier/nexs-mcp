package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/config"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/quality"
)

func setupTestServer() *MCPServer {
	repo := infrastructure.NewInMemoryElementRepository()
	server := newTestServer("nexs-mcp-test", "0.2.0", repo)

	// Initialize retention service for quality tests
	qualityConfig := quality.DefaultConfig()
	var scorer quality.Scorer
	fallbackScorer, err := quality.NewFallbackScorer(qualityConfig)
	if err != nil {
		// Use implicit scorer as fallback
		scorer = quality.NewImplicitScorer(qualityConfig)
	} else {
		scorer = fallbackScorer
	}
	server.retentionService = application.NewMemoryRetentionService(
		qualityConfig,
		scorer,
		repo,
		server.workingMemory,
	)

	return server
}

// --- Persona Tests ---

func TestHandleCreatePersona_Success(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := CreatePersonaInput{
		Name:         "Test Persona",
		Description:  "A test persona",
		Version:      "1.0.0",
		Author:       "tester",
		SystemPrompt: "You are a helpful assistant",
		BehavioralTraits: []domain.BehavioralTrait{
			{Name: "friendly", Intensity: 8},
		},
		ExpertiseAreas: []domain.ExpertiseArea{
			{Domain: "testing", Level: "expert"},
		},
		ResponseStyle: &domain.ResponseStyle{
			Tone:      "professional",
			Formality: "neutral",
			Verbosity: "balanced",
		},
		PrivacyLevel: "public",
	}

	_, output, err := server.handleCreatePersona(ctx, nil, input)
	require.NoError(t, err)
	assert.NotEmpty(t, output.ID)
	assert.Equal(t, "Test Persona", output.Element["name"])
}

func TestHandleCreatePersona_ValidationErrors(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input CreatePersonaInput
		error string
	}{
		{
			name: "missing name",
			input: CreatePersonaInput{
				Version:      "1.0.0",
				Author:       "tester",
				SystemPrompt: "Test prompt",
			},
			error: "name must be between 3 and 100 characters",
		},
		{
			name: "missing system prompt",
			input: CreatePersonaInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
			},
			error: "system_prompt must be between 10 and 2000 characters",
		},
		{
			name: "invalid privacy level",
			input: CreatePersonaInput{
				Name:         "Test",
				Version:      "1.0.0",
				Author:       "tester",
				SystemPrompt: "You are helpful",
				PrivacyLevel: "invalid",
			},
			error: "invalid privacy_level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleCreatePersona(ctx, nil, tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
		})
	}
}

func TestHandleCreatePersona_AutoExtractCreatesSkills(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	// Enable skill extraction and auto-on-create
	cfg := newTestConfig()
	cfg.SkillExtraction = config.SkillExtractionConfig{
		Enabled:                   true,
		AutoExtractOnCreate:       true,
		SkipDuplicates:            true,
		MinSkillNameLength:        3,
		MaxSkillsPerPersona:       50,
		ExtractFromExpertiseAreas: true,
		ExtractFromCustomFields:   true,
		AutoUpdatePersona:         true,
	}
	server := NewMCPServer("nexs-mcp-test", "0.2.0", repo, cfg)
	ctx := context.Background()

	input := CreatePersonaInput{
		Name:             "Auto Extract Persona",
		Version:          "1.0.0",
		Author:           "tester",
		SystemPrompt:     "You are an expert",
		ResponseStyle:    &domain.ResponseStyle{Tone: "direct", Formality: "formal", Verbosity: "concise"},
		BehavioralTraits: []domain.BehavioralTrait{{Name: "hands-on", Intensity: 8}},
		ExpertiseAreas: []domain.ExpertiseArea{
			{Domain: "Golang", Level: "expert"},
			{Domain: "Observability", Level: "advanced"},
		},
	}

	_, output, err := server.handleCreatePersona(ctx, nil, input)
	require.NoError(t, err)
	personaID := output.ID

	// Wait until skills are created and persona is updated (async)
	require.Eventually(t, func() bool {
		// List skills
		skillType := domain.SkillElement
		skills, _ := repo.List(domain.ElementFilter{Type: &skillType})
		if len(skills) < 2 {
			return false
		}
		// Check persona related skills
		pElem, err := repo.GetByID(personaID)
		if err != nil {
			return false
		}
		p := pElem.(*domain.Persona)
		return len(p.RelatedSkills) >= 2
	}, 3*time.Second, 200*time.Millisecond)
}

// --- Skill Tests ---

func TestHandleCreateSkill_Success(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := CreateSkillInput{
		Name:        "Test Skill",
		Description: "A test skill",
		Version:     "1.0.0",
		Author:      "tester",
		Triggers: []domain.SkillTrigger{
			{Type: "keyword", Keywords: []string{"test"}},
		},
		Procedures: []domain.SkillProcedure{
			{Step: 1, Action: "do something"},
		},
	}

	_, output, err := server.handleCreateSkill(ctx, nil, input)
	require.NoError(t, err)
	assert.NotEmpty(t, output.ID)
	assert.Equal(t, "Test Skill", output.Element["name"])
}

func TestHandleCreateSkill_ValidationErrors(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input CreateSkillInput
		error string
	}{
		{
			name: "missing triggers",
			input: CreateSkillInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
				Procedures: []domain.SkillProcedure{
					{Step: 1, Action: "do something"},
				},
			},
			error: "at least one trigger is required",
		},
		{
			name: "missing procedures",
			input: CreateSkillInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
				Triggers: []domain.SkillTrigger{
					{Type: "keyword", Keywords: []string{"test"}},
				},
			},
			error: "at least one procedure is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleCreateSkill(ctx, nil, tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
		})
	}
}

// --- Template Tests ---

func TestHandleCreateTemplate_Success(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := CreateTemplateInput{
		Name:        "Test Template",
		Description: "A test template",
		Version:     "1.0.0",
		Author:      "tester",
		Content:     "Hello {{name}}!",
		Format:      "markdown",
		Variables: []domain.TemplateVariable{
			{Name: "name", Type: "string", Required: true},
		},
	}

	_, output, err := server.handleCreateTemplate(ctx, nil, input)
	require.NoError(t, err)
	assert.NotEmpty(t, output.ID)
	assert.Equal(t, "Test Template", output.Element["name"])
}

func TestHandleCreateTemplate_ValidationErrors(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input CreateTemplateInput
		error string
	}{
		{
			name: "missing content",
			input: CreateTemplateInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
			},
			error: "content is required",
		},
		{
			name: "invalid format",
			input: CreateTemplateInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
				Content: "Test",
				Format:  "invalid",
			},
			error: "template validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleCreateTemplate(ctx, nil, tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
		})
	}
}

// --- Agent Tests ---

func TestHandleCreateAgent_Success(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := CreateAgentInput{
		Name:        "Test Agent",
		Description: "A test agent",
		Version:     "1.0.0",
		Author:      "tester",
		Goals:       []string{"accomplish task"},
		Actions: []domain.AgentAction{
			{Name: "action1", Type: "tool"},
		},
		MaxIterations: 5,
	}

	_, output, err := server.handleCreateAgent(ctx, nil, input)
	require.NoError(t, err)
	assert.NotEmpty(t, output.ID)
	assert.Equal(t, "Test Agent", output.Element["name"])
}

func TestHandleCreateAgent_ValidationErrors(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input CreateAgentInput
		error string
	}{
		{
			name: "missing goals",
			input: CreateAgentInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
				Actions: []domain.AgentAction{{Name: "action1", Type: "tool"}},
			},
			error: "at least one goal is required",
		},
		{
			name: "missing actions",
			input: CreateAgentInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
				Goals:   []string{"goal1"},
			},
			error: "at least one action is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleCreateAgent(ctx, nil, tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
		})
	}
}

// --- Memory Tests ---

func TestHandleCreateMemory_Success(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := CreateMemoryInput{
		Name:        "Test Memory",
		Description: "A test memory",
		Version:     "1.0.0",
		Author:      "tester",
		Content:     "This is a memory content",
	}

	_, output, err := server.handleCreateMemory(ctx, nil, input)
	require.NoError(t, err)
	assert.NotEmpty(t, output.ID)
	assert.Equal(t, "Test Memory", output.Element["name"])

	// Verify memory was created with hash
	memory, err := server.repo.GetByID(output.ID)
	require.NoError(t, err)
	memoryObj, ok := memory.(*domain.Memory)
	require.True(t, ok)
	assert.NotEmpty(t, memoryObj.ContentHash)
}

func TestHandleCreateMemory_ValidationErrors(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input CreateMemoryInput
		error string
	}{
		{
			name: "missing content",
			input: CreateMemoryInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
			},
			error: "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleCreateMemory(ctx, nil, tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
		})
	}
}

// --- Ensemble Tests ---

func TestHandleCreateEnsemble_Success(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	input := CreateEnsembleInput{
		Name:        "Test Ensemble",
		Description: "A test ensemble",
		Version:     "1.0.0",
		Author:      "tester",
		Members: []domain.EnsembleMember{
			{AgentID: "agent1", Role: "leader", Priority: 1},
		},
		ExecutionMode:       "sequential",
		AggregationStrategy: "vote",
	}

	_, output, err := server.handleCreateEnsemble(ctx, nil, input)
	require.NoError(t, err)
	assert.NotEmpty(t, output.ID)
	assert.Equal(t, "Test Ensemble", output.Element["name"])
}

func TestHandleCreateEnsemble_ValidationErrors(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input CreateEnsembleInput
		error string
	}{
		{
			name: "missing members",
			input: CreateEnsembleInput{
				Name:                "Test",
				Version:             "1.0.0",
				Author:              "tester",
				AggregationStrategy: "vote",
			},
			error: "at least one member is required",
		},
		{
			name: "missing aggregation strategy",
			input: CreateEnsembleInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
				Members: []domain.EnsembleMember{
					{AgentID: "agent1", Role: "leader", Priority: 1},
				},
			},
			error: "aggregation_strategy is required",
		},
		{
			name: "invalid execution mode",
			input: CreateEnsembleInput{
				Name:    "Test",
				Version: "1.0.0",
				Author:  "tester",
				Members: []domain.EnsembleMember{
					{AgentID: "agent1", Role: "leader", Priority: 1},
				},
				ExecutionMode:       "invalid",
				AggregationStrategy: "vote",
			},
			error: "ensemble validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := server.handleCreateEnsemble(ctx, nil, tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
		})
	}
}
