package infrastructure

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

func TestExtractAndRestorePersonaWithRelatedSkills(t *testing.T) {
	t.Parallel()

	// Create a persona with related skills
	persona := domain.NewPersona("Test Persona", "Test description", "1.0.0", "test-author")
	persona.SystemPrompt = "You are a test persona"

	// Add behavioral trait (required)
	_ = persona.AddBehavioralTrait(domain.BehavioralTrait{
		Name:      "analytical",
		Intensity: 8,
	})

	// Set response style (required)
	_ = persona.SetResponseStyle(domain.ResponseStyle{
		Tone:      "professional",
		Formality: "formal",
		Verbosity: "balanced",
	})

	// Add related skills
	persona.AddRelatedSkill("skill_golang_123")
	persona.AddRelatedSkill("skill_kubernetes_456")
	persona.AddRelatedSkill("skill_ci_cd_789")

	// Extract data
	data := extractElementData(persona)

	// Verify related_skills is in extracted data
	relatedSkills, ok := data["related_skills"]
	if !ok {
		t.Fatal("related_skills not found in extracted data")
	}

	relatedSkillsSlice, ok := relatedSkills.([]string)
	if !ok {
		t.Fatalf("related_skills is not []string, got %T", relatedSkills)
	}

	if len(relatedSkillsSlice) != 3 {
		t.Errorf("Expected 3 related skills, got %d", len(relatedSkillsSlice))
	}

	expectedSkills := []string{"skill_golang_123", "skill_kubernetes_456", "skill_ci_cd_789"}
	for i, expected := range expectedSkills {
		if i >= len(relatedSkillsSlice) {
			t.Errorf("Missing skill at index %d", i)
			continue
		}
		if relatedSkillsSlice[i] != expected {
			t.Errorf("Expected skill %s at index %d, got %s", expected, i, relatedSkillsSlice[i])
		}
	}

	// Create new persona and restore data
	restoredPersona := domain.NewPersona("Restored", "desc", "1.0.0", "author")
	restoreElementData(restoredPersona, data)

	// Verify related skills were restored
	if len(restoredPersona.RelatedSkills) != 3 {
		t.Errorf("Expected 3 related skills in restored persona, got %d", len(restoredPersona.RelatedSkills))
	}

	for i, expected := range expectedSkills {
		if i >= len(restoredPersona.RelatedSkills) {
			t.Errorf("Missing restored skill at index %d", i)
			continue
		}
		if restoredPersona.RelatedSkills[i] != expected {
			t.Errorf("Expected restored skill %s at index %d, got %s", expected, i, restoredPersona.RelatedSkills[i])
		}
	}
}

func TestFileRepositoryPersonaWithRelatedSkills(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repo, err := NewFileElementRepository(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create persona with related skills
	persona := domain.NewPersona("Senior Go Engineer", "Expert in Go", "1.0.0", "test")
	persona.SystemPrompt = "You are a senior Go engineer"

	_ = persona.AddBehavioralTrait(domain.BehavioralTrait{
		Name:      "analytical",
		Intensity: 8,
	})

	_ = persona.SetResponseStyle(domain.ResponseStyle{
		Tone:      "professional",
		Formality: "formal",
		Verbosity: "balanced",
	})

	// Add related skills
	persona.AddRelatedSkill("skill_golang_20260102_123456")
	persona.AddRelatedSkill("skill_kubernetes_20260102_123456")

	// Save persona
	if err := repo.Create(persona); err != nil {
		t.Fatalf("Failed to create persona: %v", err)
	}

	// Retrieve persona
	retrieved, err := repo.GetByID(persona.GetID())
	if err != nil {
		t.Fatalf("Failed to retrieve persona: %v", err)
	}

	retrievedPersona, ok := retrieved.(*domain.Persona)
	if !ok {
		t.Fatal("Retrieved element is not a persona")
	}

	// Verify related skills were persisted
	if len(retrievedPersona.RelatedSkills) != 2 {
		t.Errorf("Expected 2 related skills in retrieved persona, got %d", len(retrievedPersona.RelatedSkills))
	}

	expectedSkills := []string{"skill_golang_20260102_123456", "skill_kubernetes_20260102_123456"}
	for i, expected := range expectedSkills {
		if i >= len(retrievedPersona.RelatedSkills) {
			t.Errorf("Missing skill at index %d", i)
			continue
		}
		if retrievedPersona.RelatedSkills[i] != expected {
			t.Errorf("Expected skill %s at index %d, got %s", expected, i, retrievedPersona.RelatedSkills[i])
		}
	}
}
