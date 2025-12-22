package integration

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/indexing"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegrationTest creates an in-memory repository for testing.
func setupIntegrationTest(t *testing.T) *infrastructure.InMemoryElementRepository {
	repo := infrastructure.NewInMemoryElementRepository()
	return repo
}

// TestBidirectionalRelationships verifies bidirectional relationship indexing and retrieval.
func TestBidirectionalRelationships(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	// Create test elements
	persona := domain.NewPersona("Test Persona", "Test persona", "1.0.0", "test")
	persona.SetSystemPrompt("Test system prompt")
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 8})
	persona.AddExpertiseArea(domain.ExpertiseArea{Domain: "testing", Level: "expert"})
	persona.SetResponseStyle(domain.ResponseStyle{Tone: "professional", Formality: "formal", Verbosity: "balanced"})
	require.NoError(t, repo.Create(persona))

	skill := domain.NewSkill("Test Skill", "Test skill", "1.0.0", "test")
	skill.Triggers = []domain.SkillTrigger{{Type: "keyword", Keywords: []string{"test"}}}
	skill.Procedures = []domain.SkillProcedure{{Step: 1, Action: "test"}}
	require.NoError(t, repo.Create(skill))

	memory := domain.NewMemory("Test Memory", "Test content", "1.0.0", "test")
	memory.Metadata["related_to"] = persona.GetID() + "," + skill.GetID()
	require.NoError(t, repo.Create(memory))

	// Create relationship index
	index := application.NewRelationshipIndex()
	require.NoError(t, index.Rebuild(ctx, repo))

	// Test forward relationships (memory -> elements)
	forwardRelated := index.GetRelatedElements(memory.GetID())
	assert.Len(t, forwardRelated, 2, "Memory should have 2 forward relationships")
	assert.Contains(t, forwardRelated, persona.GetID())
	assert.Contains(t, forwardRelated, skill.GetID())

	// Test reverse relationships (element -> memories)
	reverseRelated := index.GetRelatedMemories(persona.GetID())
	assert.Len(t, reverseRelated, 1, "Persona should have 1 reverse relationship")
	assert.Contains(t, reverseRelated, memory.GetID())

	// Test bidirectional query
	bidirectional := index.GetBidirectionalRelationships(persona.GetID())
	assert.Empty(t, bidirectional.Forward, "Persona has no forward relationships")
	assert.Len(t, bidirectional.Reverse, 1, "Persona has 1 reverse relationship")

	// Test GetAllRelatedElements
	allRelated := index.GetAllRelatedElements(memory.GetID())
	assert.Len(t, allRelated, 2, "Should return all unique related elements")

	t.Log("✓ Bidirectional relationships working correctly")
}

// TestRecursiveExpansion verifies multi-level relationship expansion.
func TestRecursiveExpansion(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	// Create a chain: persona -> skill -> template
	persona := domain.NewPersona("Root Persona", "Root", "1.0.0", "test")
	persona.SetSystemPrompt("Root prompt")
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 8})
	persona.AddExpertiseArea(domain.ExpertiseArea{Domain: "testing", Level: "expert"})
	persona.SetResponseStyle(domain.ResponseStyle{Tone: "professional", Formality: "formal", Verbosity: "balanced"})
	require.NoError(t, repo.Create(persona))

	skill := domain.NewSkill("Level 1 Skill", "Skill", "1.0.0", "test")
	skill.Triggers = []domain.SkillTrigger{{Type: "keyword", Keywords: []string{"test"}}}
	skill.Procedures = []domain.SkillProcedure{{Step: 1, Action: "test"}}
	require.NoError(t, repo.Create(skill))

	template := domain.NewTemplate("Level 2 Template", "Template", "1.0.0", "test")
	template.Content = "Test content"
	template.Format = "markdown"
	require.NoError(t, repo.Create(template))

	// Create memory relationships
	memory1 := domain.NewMemory("Memory 1", "Content 1", "1.0.0", "test")
	memory1.Metadata["related_to"] = persona.GetID() + "," + skill.GetID()
	require.NoError(t, repo.Create(memory1))

	memory2 := domain.NewMemory("Memory 2", "Content 2", "1.0.0", "test")
	memory2.Metadata["related_to"] = skill.GetID() + "," + template.GetID()
	require.NoError(t, repo.Create(memory2))

	// Create and rebuild index
	index := application.NewRelationshipIndex()
	require.NoError(t, index.Rebuild(ctx, repo))

	// Test expansion with depth 1
	opts := application.RelationshipExpansionOptions{
		MaxDepth:       1,
		ExcludeVisited: true,
		FollowBothWays: true,
	}
	graph1, err := index.ExpandRelationships(ctx, memory1.GetID(), repo, opts)
	require.NoError(t, err)
	assert.Equal(t, 0, graph1.Depth)
	assert.Len(t, graph1.Children, 2, "Depth 1 should find 2 direct relationships")

	// Test expansion with depth 3
	opts.MaxDepth = 3
	graph3, err := index.ExpandRelationships(ctx, memory1.GetID(), repo, opts)
	require.NoError(t, err)
	assert.Equal(t, 0, graph3.Depth)
	assert.Greater(t, len(graph3.Children), 0, "Depth 3 should find relationships")

	t.Log("✓ Recursive expansion working correctly")
}

// TestRelationshipInference verifies automatic relationship inference.
func TestRelationshipInference(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	// Create elements with inferable relationships
	persona := domain.NewPersona("Data Scientist", "Expert in data analysis", "1.0.0", "test")
	persona.SetSystemPrompt("You are a data scientist")
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 9})
	persona.AddExpertiseArea(domain.ExpertiseArea{Domain: "data-science", Level: "expert"})
	persona.SetResponseStyle(domain.ResponseStyle{Tone: "professional", Formality: "formal", Verbosity: "balanced"})
	meta := persona.GetMetadata()
	meta.Tags = []string{"data", "science", "analysis", "python"}
	persona.SetMetadata(meta)
	require.NoError(t, repo.Create(persona))

	skill := domain.NewSkill("Data Analysis", "Analyze datasets", "1.0.0", "test")
	skill.Triggers = []domain.SkillTrigger{{Type: "keyword", Keywords: []string{"data", "analysis", "dataset"}}}
	skill.Procedures = []domain.SkillProcedure{{Step: 1, Action: "analyze"}}
	skillMeta := skill.GetMetadata()
	skillMeta.Tags = []string{"data", "analysis", "python"}
	skill.SetMetadata(skillMeta)
	require.NoError(t, repo.Create(skill))

	// Create memory that mentions the persona by name
	memory := domain.NewMemory("Analysis Task",
		"Need help from Data Scientist to analyze the dataset",
		"1.0.0", "test")
	require.NoError(t, repo.Create(memory))

	// Create inference engine
	index := application.NewRelationshipIndex()
	tfidfIndex := indexing.NewTFIDFIndex()
	inferenceEngine := application.NewRelationshipInferenceEngine(repo, index, tfidfIndex)

	// Test mention-based inference
	opts := application.InferenceOptions{
		MinConfidence:   0.5,
		Methods:         []string{"mention"},
		RequireEvidence: 1,
		AutoApply:       false,
	}
	inferences, err := inferenceEngine.InferRelationshipsForElement(ctx, memory.GetID(), opts)
	require.NoError(t, err)
	assert.Greater(t, len(inferences), 0, "Should find at least one inference")

	// Verify mention inference found the persona
	foundPersona := false
	for _, inf := range inferences {
		if inf.TargetID == persona.GetID() {
			foundPersona = true
			assert.GreaterOrEqual(t, inf.Confidence, 0.7, "Mention-based confidence should be high")
		}
	}
	assert.True(t, foundPersona, "Should infer relationship to persona via mention")

	// Test keyword-based inference
	opts.Methods = []string{"keyword"}
	inferences, err = inferenceEngine.InferRelationshipsForElement(ctx, persona.GetID(), opts)
	require.NoError(t, err)

	// Should find skill due to shared tags
	foundSkill := false
	for _, inf := range inferences {
		if inf.TargetID == skill.GetID() {
			foundSkill = true
			assert.GreaterOrEqual(t, inf.Confidence, 0.3, "Keyword confidence should be reasonable")
		}
	}
	assert.True(t, foundSkill, "Should infer relationship to skill via shared tags")

	t.Log("✓ Relationship inference working correctly")
}

// TestRecommendationEngine verifies intelligent recommendations.
func TestRecommendationEngine(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	// Create a network of related elements
	persona := domain.NewPersona("Backend Engineer", "Expert in Go", "1.0.0", "test")
	persona.SetSystemPrompt("You are a backend engineer")
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "pragmatic", Intensity: 8})
	persona.AddExpertiseArea(domain.ExpertiseArea{Domain: "golang", Level: "expert"})
	persona.SetResponseStyle(domain.ResponseStyle{Tone: "professional", Formality: "neutral", Verbosity: "concise"})
	meta := persona.GetMetadata()
	meta.Tags = []string{"golang", "backend", "api"}
	persona.SetMetadata(meta)
	require.NoError(t, repo.Create(persona))

	skill1 := domain.NewSkill("REST API Design", "Design REST APIs", "1.0.0", "test")
	skill1.Triggers = []domain.SkillTrigger{{Type: "keyword", Keywords: []string{"api", "rest"}}}
	skill1.Procedures = []domain.SkillProcedure{{Step: 1, Action: "design"}}
	skill1Meta := skill1.GetMetadata()
	skill1Meta.Tags = []string{"api", "rest", "backend"}
	skill1.SetMetadata(skill1Meta)
	require.NoError(t, repo.Create(skill1))

	// Add direct relationship
	persona.AddRelatedSkill(skill1.GetID())
	require.NoError(t, repo.Update(persona))

	// Create memories
	memory1 := domain.NewMemory("API Project", "Working on REST API", "1.0.0", "test")
	memory1.Metadata["related_to"] = persona.GetID() + "," + skill1.GetID()
	require.NoError(t, repo.Create(memory1))

	// Create recommendation engine
	index := application.NewRelationshipIndex()
	require.NoError(t, index.Rebuild(ctx, repo))
	engine := application.NewRecommendationEngine(repo, index)

	// Get recommendations for persona
	opts := application.RecommendationOptions{
		MaxResults:     5,
		MinScore:       0.1,
		IncludeReasons: true,
	}
	recommendations, err := engine.RecommendForElement(ctx, persona.GetID(), opts)
	require.NoError(t, err)
	assert.Greater(t, len(recommendations), 0, "Should generate recommendations")

	// Verify skill1 is recommended (direct relationship)
	foundSkill1 := false
	for _, rec := range recommendations {
		if rec.ElementID == skill1.GetID() {
			foundSkill1 = true
			assert.Equal(t, 1.0, rec.Score, "Direct relationship should have score 1.0")
			assert.Contains(t, rec.Reasons, "directly related")
		}
	}
	assert.True(t, foundSkill1, "Should recommend directly related skill")

	t.Log("✓ Recommendation engine working correctly")
}

// TestCacheEfficiency verifies relationship cache performance.
func TestCacheEfficiency(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	// Create test data
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	persona.SetSystemPrompt("Test prompt")
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 8})
	persona.AddExpertiseArea(domain.ExpertiseArea{Domain: "testing", Level: "expert"})
	persona.SetResponseStyle(domain.ResponseStyle{Tone: "professional", Formality: "formal", Verbosity: "balanced"})
	require.NoError(t, repo.Create(persona))

	memory := domain.NewMemory("Test Memory", "Content", "1.0.0", "test")
	memory.Metadata["related_to"] = persona.GetID()
	require.NoError(t, repo.Create(memory))

	// Create index
	index := application.NewRelationshipIndex()
	require.NoError(t, index.Rebuild(ctx, repo))

	// Get initial stats
	stats1 := index.Stats()
	initialMisses := stats1.CacheMisses

	// First call - should be cache miss
	memories1, err := application.GetMemoriesRelatedTo(ctx, persona.GetID(), repo, index)
	require.NoError(t, err)
	assert.Len(t, memories1, 1)

	stats2 := index.Stats()
	assert.Equal(t, initialMisses+1, stats2.CacheMisses, "First call should be cache miss")

	// Second call - should be cache hit
	memories2, err := application.GetMemoriesRelatedTo(ctx, persona.GetID(), repo, index)
	require.NoError(t, err)
	assert.Len(t, memories2, 1)

	stats3 := index.Stats()
	assert.Greater(t, stats3.CacheHits, stats2.CacheHits, "Second call should be cache hit")

	t.Log("✓ Cache efficiency working correctly")
}

// TestRelationshipIndexStats verifies index statistics.
func TestRelationshipIndexStats(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	// Create multiple elements with relationships
	for i := 0; i < 3; i++ {
		persona := domain.NewPersona("Persona "+string(rune('A'+i)), "Test", "1.0.0", "test")
		persona.SetSystemPrompt("Test prompt")
		persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 8})
		persona.AddExpertiseArea(domain.ExpertiseArea{Domain: "testing", Level: "expert"})
		persona.SetResponseStyle(domain.ResponseStyle{Tone: "professional", Formality: "formal", Verbosity: "balanced"})
		require.NoError(t, repo.Create(persona))
	}

	for i := 0; i < 5; i++ {
		memory := domain.NewMemory("Memory "+string(rune('1'+i)), "Content", "1.0.0", "test")
		memory.Metadata["related_to"] = "persona-A,persona-B"
		require.NoError(t, repo.Create(memory))
	}

	// Create and rebuild index
	index := application.NewRelationshipIndex()
	require.NoError(t, index.Rebuild(ctx, repo))

	// Get stats
	stats := index.Stats()
	assert.Greater(t, stats.ForwardEntries, 0, "Should have forward entries")
	assert.GreaterOrEqual(t, stats.CacheSize, 0, "Cache size should be non-negative")

	t.Log("✓ Relationship index stats working correctly")
	t.Logf("  Forward entries: %d", stats.ForwardEntries)
	t.Logf("  Reverse entries: %d", stats.ReverseEntries)
}
