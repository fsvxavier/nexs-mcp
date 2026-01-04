package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/config"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

// TestPersonaCreationWithSyncSkillExtraction tests the complete flow of
// persona creation with synchronous skill extraction enabled.
func TestPersonaCreationWithSyncSkillExtraction(t *testing.T) {
	t.Parallel()

	// Setup server with skill extraction enabled
	cfg := &config.Config{
		SkillExtraction: config.SkillExtractionConfig{
			Enabled:                   true,
			AutoExtractOnCreate:       true,
			SkipDuplicates:            true,
			MinSkillNameLength:        3,
			MaxSkillsPerPersona:       50,
			ExtractFromExpertiseAreas: true,
			ExtractFromCustomFields:   true,
			AutoUpdatePersona:         true,
		},
	}

	server := setupTestServerWithConfig("test-server", "1.0.0", cfg)
	ctx := context.Background()

	// Create persona with expertise areas that should generate skills
	input := CreatePersonaInput{
		Name:         "Senior Go Engineer",
		Description:  "Expert in Go programming and system design",
		Version:      "1.0.0",
		Author:       "test",
		SystemPrompt: "You are a senior Go engineer with expertise in distributed systems.",
		BehavioralTraits: []domain.BehavioralTrait{
			{
				Name:      "analytical",
				Intensity: 8,
			},
		},
		ResponseStyle: &domain.ResponseStyle{
			Tone:      "professional",
			Formality: "formal",
			Verbosity: "balanced",
		},
		ExpertiseAreas: []domain.ExpertiseArea{
			{
				Domain:      "Go Programming",
				Level:       "expert",
				Description: "Deep knowledge of Go language, concurrency, and best practices",
				Keywords:    []string{"golang", "concurrency", "goroutines"},
			},
			{
				Domain:      "System Design",
				Level:       "expert",
				Description: "Expertise in designing scalable distributed systems",
				Keywords:    []string{"architecture", "microservices", "scalability"},
			},
		},
	}

	// Create persona - skills should be extracted synchronously
	_, output, err := server.handleCreatePersona(ctx, nil, input)
	if err != nil {
		t.Fatalf("Failed to create persona: %v", err)
	}

	personaID := output.ID
	if personaID == "" {
		t.Fatal("Persona ID is empty")
	}

	// Verify skills were created (synchronously)
	skillsCreated, ok := output.Element["skills_created"].(int)
	if !ok {
		t.Error("skills_created field missing or wrong type in output")
	}
	if skillsCreated < 2 {
		t.Errorf("Expected at least 2 skills to be created, got %d", skillsCreated)
	}

	skillIDs, ok := output.Element["skill_ids"].([]string)
	if !ok {
		t.Error("skill_ids field missing or wrong type in output")
	}
	if len(skillIDs) != skillsCreated {
		t.Errorf("Mismatch: skills_created=%d but skill_ids has %d entries", skillsCreated, len(skillIDs))
	}

	// Verify skills actually exist in repository
	for _, skillID := range skillIDs {
		elem, err := server.repo.GetByID(skillID)
		if err != nil {
			t.Errorf("Skill %s not found in repository: %v", skillID, err)
			continue
		}

		skill, ok := elem.(*domain.Skill)
		if !ok {
			t.Errorf("Element %s is not a skill", skillID)
			continue
		}

		// Verify skill has related_personas metadata
		metadata := skill.GetMetadata()
		if metadata.Custom == nil {
			t.Errorf("Skill %s has no custom metadata", skillID)
			continue
		}

		relatedPersonas, ok := metadata.Custom["related_personas"].([]string)
		if !ok {
			t.Errorf("Skill %s missing related_personas in custom metadata", skillID)
			continue
		}

		// Verify persona is linked to skill
		found := false
		for _, pid := range relatedPersonas {
			if pid == personaID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Skill %s not linked to persona %s", skillID, personaID)
		}
	}

	// Verify persona has related_skills
	personaElem, err := server.repo.GetByID(personaID)
	if err != nil {
		t.Fatalf("Failed to retrieve persona: %v", err)
	}

	persona, ok := personaElem.(*domain.Persona)
	if !ok {
		t.Fatal("Element is not a persona")
	}

	if len(persona.RelatedSkills) != len(skillIDs) {
		t.Errorf("Persona has %d related skills, expected %d", len(persona.RelatedSkills), len(skillIDs))
	}

	// Verify bidirectional relationship
	for _, skillID := range skillIDs {
		found := false
		for _, relSkillID := range persona.RelatedSkills {
			if relSkillID == skillID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Persona missing link to skill %s", skillID)
		}
	}
}

// TestPersonaCreationSkillExtractionDisabled verifies that no skills are created
// when skill extraction is disabled.
func TestPersonaCreationSkillExtractionDisabled(t *testing.T) {
	t.Parallel()

	// Setup server with skill extraction DISABLED
	cfg := &config.Config{
		SkillExtraction: config.SkillExtractionConfig{
			Enabled:             false,
			AutoExtractOnCreate: false,
		},
	}

	server := setupTestServerWithConfig("test-server", "1.0.0", cfg)
	ctx := context.Background()

	input := CreatePersonaInput{
		Name:         "Test Persona",
		Description:  "Test persona for skill extraction disabled",
		Version:      "1.0.0",
		Author:       "test",
		SystemPrompt: "You are a test persona.",
		BehavioralTraits: []domain.BehavioralTrait{
			{
				Name:      "methodical",
				Intensity: 7,
			},
		},
		ResponseStyle: &domain.ResponseStyle{
			Tone:      "friendly",
			Formality: "casual",
			Verbosity: "concise",
		},
		ExpertiseAreas: []domain.ExpertiseArea{
			{
				Domain:      "Testing",
				Level:       "expert",
				Description: "Expert in testing",
				Keywords:    []string{"testing", "qa"},
			},
		},
	}

	_, output, err := server.handleCreatePersona(ctx, nil, input)
	if err != nil {
		t.Fatalf("Failed to create persona: %v", err)
	}

	// Verify NO skills were created
	if skillsCreated, ok := output.Element["skills_created"].(int); ok {
		t.Errorf("Expected no skills_created field when extraction is disabled, but got %d", skillsCreated)
	}

	if skillIDs, ok := output.Element["skill_ids"].([]string); ok && len(skillIDs) > 0 {
		t.Errorf("Expected no skill_ids when extraction is disabled, but got %v", skillIDs)
	}
}

// setupTestServerWithConfig creates a test server with custom configuration.
func setupTestServerWithConfig(name, version string, cfg *config.Config) *MCPServer {
	repo := infrastructure.NewInMemoryElementRepository()
	if cfg.ServerName == "" {
		cfg.ServerName = name
	}
	if cfg.Version == "" {
		cfg.Version = version
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "/tmp/nexs-test"
	}

	// Create server with config
	server := NewMCPServer(name, version, repo, cfg)

	// Give time for initialization
	time.Sleep(100 * time.Millisecond)

	return server
}
