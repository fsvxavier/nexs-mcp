package application

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepoForInference implements domain.ElementRepository for inference testing.
type mockRepoForInference struct {
	elements map[string]domain.Element
}

func newMockRepoForInference() *mockRepoForInference {
	return &mockRepoForInference{
		elements: make(map[string]domain.Element),
	}
}

func (m *mockRepoForInference) Create(elem domain.Element) error {
	m.elements[elem.GetMetadata().ID] = elem
	return nil
}

func (m *mockRepoForInference) GetByID(id string) (domain.Element, error) {
	if elem, ok := m.elements[id]; ok {
		return elem, nil
	}
	return nil, domain.ErrElementNotFound
}

func (m *mockRepoForInference) Update(elem domain.Element) error {
	m.elements[elem.GetMetadata().ID] = elem
	return nil
}

func (m *mockRepoForInference) Delete(id string) error {
	delete(m.elements, id)
	return nil
}

func (m *mockRepoForInference) List(filter domain.ElementFilter) ([]domain.Element, error) {
	result := make([]domain.Element, 0, len(m.elements))
	for _, elem := range m.elements {
		if filter.Type != nil && elem.GetMetadata().Type != *filter.Type {
			continue
		}
		result = append(result, elem)
	}
	return result, nil
}

func (m *mockRepoForInference) Exists(id string) (bool, error) {
	_, exists := m.elements[id]
	return exists, nil
}

func setupInferenceEngine(t *testing.T) (*RelationshipInferenceEngine, *mockRepoForInference) {
	t.Helper()

	repo := newMockRepoForInference()
	index := NewRelationshipIndex()

	// Create hybrid search with mock provider
	provider := embeddings.NewMockProvider("mock", 384)
	hybridSearch := NewHybridSearchService(HybridSearchConfig{
		Provider:    provider,
		AutoReindex: false, // Disable for tests
	})

	engine := NewRelationshipInferenceEngine(repo, index, hybridSearch)
	return engine, repo
}

func TestNewRelationshipInferenceEngine(t *testing.T) {
	engine, _ := setupInferenceEngine(t)
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.repo)
	assert.NotNil(t, engine.index)
	assert.NotNil(t, engine.hybridSearch)
}

func TestInferRelationshipsForElement_BasicFunctionality(t *testing.T) {
	engine, repo := setupInferenceEngine(t)
	ctx := context.Background()

	// Create a simple memory
	memory := domain.NewMemory("Test Memory", "A test memory", "1.0.0", "testuser")
	memory.Content = "This is a test memory"
	err := repo.Create(memory)
	require.NoError(t, err)

	// Test with default options
	opts := InferenceOptions{
		MinConfidence: 0.5,
		Methods:       []string{"mention"},
	}

	inferences, err := engine.InferRelationshipsForElement(ctx, memory.GetID(), opts)
	require.NoError(t, err)
	assert.NotNil(t, inferences)
}

func TestInferRelationships_NonExistentElement(t *testing.T) {
	engine, _ := setupInferenceEngine(t)
	ctx := context.Background()

	opts := InferenceOptions{
		MinConfidence: 0.5,
		Methods:       []string{"mention"},
	}

	// Test with non-existent element
	inferences, err := engine.InferRelationshipsForElement(ctx, "nonexistent_id", opts)
	assert.Error(t, err, "Should error when source element not found")
	assert.Nil(t, inferences)
}

func TestInferenceOptions_Defaults(t *testing.T) {
	engine, repo := setupInferenceEngine(t)
	ctx := context.Background()

	// Create a simple memory
	memory := domain.NewMemory("Test Memory", "Test", "1.0.0", "testuser")
	memory.Content = "Test content"
	err := repo.Create(memory)
	require.NoError(t, err)

	// Test with empty options (should use defaults)
	opts := InferenceOptions{}

	inferences, err := engine.InferRelationshipsForElement(ctx, memory.GetID(), opts)
	require.NoError(t, err)
	assert.NotNil(t, inferences, "Should not error with default options")
}

func TestInferRelationships_WithMultipleMethods(t *testing.T) {
	engine, repo := setupInferenceEngine(t)
	ctx := context.Background()

	// Create elements
	skill := domain.NewSkill("Test Skill", "A test skill", "1.0.0", "testuser")
	err := skill.AddTrigger(domain.SkillTrigger{Pattern: "test", Type: "keyword"})
	require.NoError(t, err)
	err = repo.Create(skill)
	require.NoError(t, err)

	memory := domain.NewMemory("Test Memory", "Test", "1.0.0", "testuser")
	memory.Content = "Using test skill"
	meta := memory.GetMetadata()
	meta.Tags = append(meta.Tags, "test")
	memory.SetMetadata(meta)
	err = repo.Create(memory)
	require.NoError(t, err)

	// Test with multiple methods
	opts := InferenceOptions{
		MinConfidence: 0.3,
		Methods:       []string{"mention", "keyword"},
		AutoApply:     false,
	}

	inferences, err := engine.InferRelationshipsForElement(ctx, memory.GetID(), opts)
	require.NoError(t, err)
	assert.NotNil(t, inferences)
}

func TestInferRelationships_WithHighConfidenceThreshold(t *testing.T) {
	engine, repo := setupInferenceEngine(t)
	ctx := context.Background()

	// Create simple elements with weak relationship
	skill := domain.NewSkill("Weak Skill", "Unrelated", "1.0.0", "testuser")
	err := skill.AddTrigger(domain.SkillTrigger{Pattern: "unrelated", Type: "keyword"})
	require.NoError(t, err)
	skillMeta := skill.GetMetadata()
	skillMeta.Tags = append(skillMeta.Tags, "different")
	skill.SetMetadata(skillMeta)
	err = repo.Create(skill)
	require.NoError(t, err)

	memory := domain.NewMemory("Weak Memory", "Unrelated", "1.0.0", "testuser")
	memory.Content = "Different content"
	memMeta := memory.GetMetadata()
	memMeta.Tags = append(memMeta.Tags, "other")
	memory.SetMetadata(memMeta)
	err = repo.Create(memory)
	require.NoError(t, err)

	// Test with very high confidence threshold
	opts := InferenceOptions{
		MinConfidence: 0.95, // Very high threshold
		Methods:       []string{"keyword"},
		AutoApply:     false,
	}

	inferences, err := engine.InferRelationshipsForElement(ctx, memory.GetID(), opts)
	require.NoError(t, err)
	// Should work without errors, even if no inferences are found
	assert.NotNil(t, inferences)
}

func TestInferRelationships_WithTargetTypeFilter(t *testing.T) {
	engine, repo := setupInferenceEngine(t)
	ctx := context.Background()

	// Create multiple element types
	skill := domain.NewSkill("Filter Skill", "Test", "1.0.0", "testuser")
	err := skill.AddTrigger(domain.SkillTrigger{Pattern: "test", Type: "keyword"})
	require.NoError(t, err)
	err = repo.Create(skill)
	require.NoError(t, err)

	persona := domain.NewPersona("Filter Persona", "Test", "1.0.0", "testuser")
	err = persona.SetSystemPrompt("Test prompt")
	require.NoError(t, err)
	err = repo.Create(persona)
	require.NoError(t, err)

	memory := domain.NewMemory("Filter Memory", "Test", "1.0.0", "testuser")
	memory.Content = "Test content"
	err = repo.Create(memory)
	require.NoError(t, err)

	// Test with target type filter (only skills)
	skillType := domain.SkillElement
	opts := InferenceOptions{
		MinConfidence: 0.3,
		Methods:       []string{"keyword"},
		TargetTypes:   []domain.ElementType{skillType},
		AutoApply:     false,
	}

	inferences, err := engine.InferRelationshipsForElement(ctx, memory.GetID(), opts)
	require.NoError(t, err)

	// Verify only skill relationships are returned (if any)
	for _, inf := range inferences {
		assert.Equal(t, skillType, inf.TargetType,
			"Should only infer relationships to skills")
	}
}

func TestInferRelationships_AutoApply(t *testing.T) {
	engine, repo := setupInferenceEngine(t)
	ctx := context.Background()

	// Create a memory
	memory := domain.NewMemory("Auto Memory", "Test", "1.0.0", "testuser")
	memory.Content = "Auto apply test"
	err := repo.Create(memory)
	require.NoError(t, err)

	// Create a skill
	skill := domain.NewSkill("Auto Skill", "Test", "1.0.0", "testuser")
	err = skill.AddTrigger(domain.SkillTrigger{Pattern: "apply", Type: "keyword"})
	require.NoError(t, err)
	err = repo.Create(skill)
	require.NoError(t, err)

	// Test with auto-apply enabled
	opts := InferenceOptions{
		MinConfidence: 0.3,
		Methods:       []string{"keyword"},
		AutoApply:     true,
	}

	inferences, err := engine.InferRelationshipsForElement(ctx, memory.GetID(), opts)
	require.NoError(t, err)
	assert.NotNil(t, inferences)
	// Auto-apply should work without errors
}

func TestInferredRelationship_Structure(t *testing.T) {
	inference := &InferredRelationship{
		SourceID:   "source_123",
		TargetID:   "target_456",
		SourceType: domain.MemoryElement,
		TargetType: domain.SkillElement,
		Confidence: 0.85,
		Evidence:   []string{"ID mention in content"},
		InferredBy: "mention",
	}

	assert.Equal(t, "source_123", inference.SourceID)
	assert.Equal(t, "target_456", inference.TargetID)
	assert.Equal(t, domain.MemoryElement, inference.SourceType)
	assert.Equal(t, domain.SkillElement, inference.TargetType)
	assert.Equal(t, 0.85, inference.Confidence)
	assert.NotEmpty(t, inference.Evidence)
	assert.Equal(t, "mention", inference.InferredBy)
}

func TestInferenceOptions_RequireEvidence(t *testing.T) {
	engine, repo := setupInferenceEngine(t)
	ctx := context.Background()

	// Create a memory
	memory := domain.NewMemory("Evidence Memory", "Test", "1.0.0", "testuser")
	memory.Content = "Test evidence"
	err := repo.Create(memory)
	require.NoError(t, err)

	// Test with evidence requirement
	opts := InferenceOptions{
		MinConfidence:   0.5,
		Methods:         []string{"mention"},
		RequireEvidence: 2, // Require at least 2 pieces of evidence
		AutoApply:       false,
	}

	inferences, err := engine.InferRelationshipsForElement(ctx, memory.GetID(), opts)
	require.NoError(t, err)

	// Verify evidence requirement is enforced (if any inferences found)
	for _, inf := range inferences {
		assert.GreaterOrEqual(t, len(inf.Evidence), 2,
			"Should have at least required evidence count")
	}
}

func TestInferRelationships_InvalidMethod(t *testing.T) {
	engine, repo := setupInferenceEngine(t)
	ctx := context.Background()

	// Create a memory
	memory := domain.NewMemory("Invalid Method Memory", "Test", "1.0.0", "testuser")
	memory.Content = "Test content"
	err := repo.Create(memory)
	require.NoError(t, err)

	// Test with invalid method (should be skipped)
	opts := InferenceOptions{
		MinConfidence: 0.5,
		Methods:       []string{"invalid_method", "mention"},
		AutoApply:     false,
	}

	inferences, err := engine.InferRelationshipsForElement(ctx, memory.GetID(), opts)
	require.NoError(t, err)
	// Should continue with valid methods
	assert.NotNil(t, inferences)
}
