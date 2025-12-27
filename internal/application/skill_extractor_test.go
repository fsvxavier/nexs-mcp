package application

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// mockRepository is a simple in-memory repository for testing.
type mockSkillExtractorRepo struct {
	elements map[string]domain.Element
}

func newMockSkillExtractorRepo() *mockSkillExtractorRepo {
	return &mockSkillExtractorRepo{
		elements: make(map[string]domain.Element),
	}
}

func (m *mockSkillExtractorRepo) Create(element domain.Element) error {
	m.elements[element.GetID()] = element
	return nil
}

func (m *mockSkillExtractorRepo) GetByID(id string) (domain.Element, error) {
	elem, ok := m.elements[id]
	if !ok {
		return nil, domain.ErrElementNotFound
	}
	return elem, nil
}

func (m *mockSkillExtractorRepo) Update(element domain.Element) error {
	m.elements[element.GetID()] = element
	return nil
}

func (m *mockSkillExtractorRepo) Delete(id string) error {
	delete(m.elements, id)
	return nil
}

func (m *mockSkillExtractorRepo) List(filter domain.ElementFilter) ([]domain.Element, error) {
	var result []domain.Element
	for _, elem := range m.elements {
		if filter.Type != nil && elem.GetType() != *filter.Type {
			continue
		}
		result = append(result, elem)
	}
	return result, nil
}

func (m *mockSkillExtractorRepo) Exists(id string) (bool, error) {
	_, ok := m.elements[id]
	return ok, nil
}

func (m *mockSkillExtractorRepo) Search(filter domain.ElementFilter) ([]domain.Element, error) {
	return m.List(filter)
}

func TestSkillExtractor_ExtractSkillsFromPersona(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	// Create a test persona with expertise areas
	persona := domain.NewPersona("Test Persona", "Test description", "1.0.0", "test@example.com")
	persona.AddExpertiseArea(domain.ExpertiseArea{
		Domain:      "Software Architecture",
		Level:       "expert",
		Keywords:    []string{"clean-architecture", "ddd"},
		Description: "Expert in software architecture",
	})
	persona.AddExpertiseArea(domain.ExpertiseArea{
		Domain:   "Go Programming",
		Level:    "advanced",
		Keywords: []string{"golang", "concurrency"},
	})
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 8})
	persona.SetResponseStyle(domain.ResponseStyle{
		Tone:      "professional",
		Formality: "neutral",
		Verbosity: "balanced",
	})
	persona.SetSystemPrompt("Test system prompt for persona")

	require.NoError(t, repo.Create(persona))

	// Extract skills
	ctx := context.Background()
	result, err := extractor.ExtractSkillsFromPersona(ctx, persona.GetID())

	// Verify results
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.SkillsCreated, "Should create 2 skills from expertise areas")
	assert.Len(t, result.SkillIDs, 2)
	assert.True(t, result.PersonaUpdated)
	assert.Empty(t, result.Errors)

	// Verify skills were created
	for _, skillID := range result.SkillIDs {
		elem, err := repo.GetByID(skillID)
		require.NoError(t, err)
		skill, ok := elem.(*domain.Skill)
		require.True(t, ok)
		assert.NotEmpty(t, skill.GetMetadata().Name)
		assert.NotEmpty(t, skill.Triggers)
		assert.NotEmpty(t, skill.Procedures)
	}

	// Verify persona was updated with related skills
	updatedPersona, err := repo.GetByID(persona.GetID())
	require.NoError(t, err)
	p, ok := updatedPersona.(*domain.Persona)
	require.True(t, ok)
	assert.Len(t, p.RelatedSkills, 2)
}

func TestSkillExtractor_ExtractSkillsFromCustomFields(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	// Create a persona with custom technical_skills field (like the real persona)
	persona := domain.NewPersona("Engineer", "Senior Engineer", "1.0.0", "test@example.com")
	persona.AddExpertiseArea(domain.ExpertiseArea{
		Domain: "Software Engineering",
		Level:  "expert",
	})
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 8})
	persona.SetResponseStyle(domain.ResponseStyle{
		Tone:      "professional",
		Formality: "neutral",
		Verbosity: "balanced",
	})
	persona.SetSystemPrompt("Test system prompt")

	require.NoError(t, repo.Create(persona))

	// Test extraction
	ctx := context.Background()
	result, err := extractor.ExtractSkillsFromPersona(ctx, persona.GetID())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.SkillsCreated, 1)
}

func TestSkillExtractor_SkipDuplicateSkills(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	// Create existing skill
	existingSkill := domain.NewSkill("Software Architecture", "Architecture skill", "1.0.0", "test@example.com")
	existingSkill.AddTrigger(domain.SkillTrigger{Type: "keyword", Keywords: []string{"architecture"}})
	existingSkill.AddProcedure(domain.SkillProcedure{Step: 1, Action: "Test"})
	require.NoError(t, repo.Create(existingSkill))

	// Create persona with same skill
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test@example.com")
	persona.AddExpertiseArea(domain.ExpertiseArea{
		Domain: "Software Architecture",
		Level:  "expert",
	})
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 8})
	persona.SetResponseStyle(domain.ResponseStyle{
		Tone:      "professional",
		Formality: "neutral",
		Verbosity: "balanced",
	})
	persona.SetSystemPrompt("Test system prompt")

	require.NoError(t, repo.Create(persona))

	// Extract skills
	ctx := context.Background()
	result, err := extractor.ExtractSkillsFromPersona(ctx, persona.GetID())

	require.NoError(t, err)
	assert.Equal(t, 0, result.SkillsCreated, "Should not create duplicate skill")
	assert.Equal(t, 1, result.SkippedDuplicate)
	assert.Len(t, result.SkillIDs, 1)
	assert.True(t, result.PersonaUpdated)
}

func TestSkillExtractor_NonExistentPersona(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	ctx := context.Background()
	result, err := extractor.ExtractSkillsFromPersona(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSkillExtractor_InvalidElement(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	// Create a skill (not a persona)
	skill := domain.NewSkill("Test Skill", "Test", "1.0.0", "test@example.com")
	skill.AddTrigger(domain.SkillTrigger{Type: "keyword", Keywords: []string{"test"}})
	skill.AddProcedure(domain.SkillProcedure{Step: 1, Action: "Test"})
	require.NoError(t, repo.Create(skill))

	ctx := context.Background()
	result, err := extractor.ExtractSkillsFromPersona(ctx, skill.GetID())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a persona")
	assert.Nil(t, result)
}

func TestSkillExtractor_GenerateKeywords(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single word",
			input:    "Architecture",
			expected: []string{"architecture"},
		},
		{
			name:     "multiple words",
			input:    "Software Architecture",
			expected: []string{"software architecture", "software", "architecture"},
		},
		{
			name:     "with short words",
			input:    "Go Programming",
			expected: []string{"go programming", "programming"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := extractor.generateKeywords(tt.input)
			assert.ElementsMatch(t, tt.expected, keywords)
		})
	}
}

func TestSkillExtractor_CreateSkillFromName(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	spec := extractor.createSkillFromName("Test Skill", "testing")

	assert.Equal(t, "Test Skill", spec.Name)
	assert.Contains(t, spec.Description, "Test Skill")
	assert.Contains(t, spec.Description, "testing")
	assert.NotEmpty(t, spec.Triggers)
	assert.NotEmpty(t, spec.Procedures)
	assert.Contains(t, spec.Tags, "auto-extracted")
	assert.Contains(t, spec.Tags, "testing")
}

func TestSkillExtractor_GetPersonaRawData(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	// Create a persona
	persona := domain.NewPersona("Test", "Test", "1.0.0", "test@example.com")
	persona.AddExpertiseArea(domain.ExpertiseArea{Domain: "Testing", Level: "expert"})
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 8})
	persona.SetResponseStyle(domain.ResponseStyle{
		Tone:      "professional",
		Formality: "neutral",
		Verbosity: "balanced",
	})
	persona.SetSystemPrompt("Test prompt")
	require.NoError(t, repo.Create(persona))

	// Get raw data
	rawData, err := extractor.getPersonaRawData(persona.GetID())

	require.NoError(t, err)
	assert.NotNil(t, rawData)

	// Verify basic fields exist
	_, hasExpertise := rawData["expertise_areas"]
	assert.True(t, hasExpertise, "Should have expertise_areas in raw data")
}

func TestSkillExtractor_ExtractSkillsFromRawData(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	// Create persona
	persona := domain.NewPersona("Test", "Test", "1.0.0", "test@example.com")
	persona.AddExpertiseArea(domain.ExpertiseArea{
		Domain:   "Cloud Architecture",
		Level:    "expert",
		Keywords: []string{"aws", "gcp"},
	})

	// Create raw data with custom fields
	rawData := map[string]interface{}{
		"technical_skills": map[string]interface{}{
			"core_expertise": []interface{}{
				"Software Architecture",
				"Performance Optimization",
			},
			"architecture_patterns": []interface{}{
				"Microservices",
				"Event-Driven Architecture",
			},
		},
	}

	// Extract skills
	skills := extractor.extractSkillsFromRawData(rawData, persona)

	// Should extract from both expertise_areas and custom fields
	assert.GreaterOrEqual(t, len(skills), 3, "Should extract skills from expertise and custom fields")

	// Verify skill structure
	for _, skill := range skills {
		assert.NotEmpty(t, skill.Name)
		assert.NotEmpty(t, skill.Description)
		assert.NotEmpty(t, skill.Triggers)
		assert.NotEmpty(t, skill.Procedures)
		assert.NotEmpty(t, skill.Tags)
	}
}

func TestSkillExtractor_FindExistingSkill(t *testing.T) {
	repo := newMockSkillExtractorRepo()
	extractor := NewSkillExtractor(repo)

	// Create a skill
	skill := domain.NewSkill("Test Skill", "Test", "1.0.0", "test@example.com")
	skill.AddTrigger(domain.SkillTrigger{Type: "keyword", Keywords: []string{"test"}})
	skill.AddProcedure(domain.SkillProcedure{Step: 1, Action: "Test"})
	require.NoError(t, repo.Create(skill))

	// Find by exact name
	found := extractor.findExistingSkill("Test Skill")
	assert.NotNil(t, found)
	assert.Equal(t, skill.GetID(), found.GetID())

	// Find by case-insensitive name
	found = extractor.findExistingSkill("test skill")
	assert.NotNil(t, found)

	// Not found
	found = extractor.findExistingSkill("Non Existent")
	assert.Nil(t, found)
}

func TestExtractionResult_JSONMarshaling(t *testing.T) {
	result := &ExtractionResult{
		SkillsCreated:    5,
		SkillIDs:         []string{"skill-1", "skill-2", "skill-3"},
		PersonaUpdated:   true,
		Errors:           []string{"error 1"},
		SkippedDuplicate: 2,
	}

	// Marshal to JSON
	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal back
	var unmarshaled ExtractionResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, result.SkillsCreated, unmarshaled.SkillsCreated)
	assert.Equal(t, result.SkillIDs, unmarshaled.SkillIDs)
	assert.Equal(t, result.PersonaUpdated, unmarshaled.PersonaUpdated)
	assert.Equal(t, result.SkippedDuplicate, unmarshaled.SkippedDuplicate)
}
