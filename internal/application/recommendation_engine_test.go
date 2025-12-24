package application

import (
	"context"
	"fmt"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

func TestNewRecommendationEngine(t *testing.T) {
	repo := newMockRepoForIndex()
	index := NewRelationshipIndex()

	engine := NewRecommendationEngine(repo, index)

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}
	if engine.repo == nil {
		t.Error("Expected non-nil repo")
	}
	if engine.index == nil {
		t.Error("Expected non-nil index")
	}
}

func TestRecommendForElement_DirectRelationships(t *testing.T) {
	repo := newMockRepoForIndex()
	index := NewRelationshipIndex()
	engine := NewRecommendationEngine(repo, index)

	// Create persona with related skills
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill1 := domain.NewSkill("Python", "Python programming", "1.0.0", "test")
	skill2 := domain.NewSkill("Go", "Go programming", "1.0.0", "test")

	persona.RelatedSkills = []string{skill1.GetMetadata().ID, skill2.GetMetadata().ID}

	repo.Create(persona)
	repo.Create(skill1)
	repo.Create(skill2)

	ctx := context.Background()
	options := RecommendationOptions{
		MaxResults: 10,
		MinScore:   0.1,
	}

	recommendations, err := engine.RecommendForElement(ctx, persona.GetMetadata().ID, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(recommendations) != 2 {
		t.Errorf("Expected 2 recommendations, got: %d", len(recommendations))
	}

	// Check scores for direct relationships
	for _, rec := range recommendations {
		if rec.Score != 1.0 {
			t.Errorf("Expected score 1.0 for direct relationship, got: %f", rec.Score)
		}
		if !containsString(rec.Reasons, "directly related") {
			t.Error("Expected 'directly related' reason")
		}
	}
}

func TestRecommendForElement_CoOccurrence(t *testing.T) {
	repo := newMockRepoForIndex()
	index := NewRelationshipIndex()
	engine := NewRecommendationEngine(repo, index)

	// Create persona and skill
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill := domain.NewSkill("Python", "Python programming", "1.0.0", "test")

	repo.Create(persona)
	repo.Create(skill)

	personaID := persona.GetMetadata().ID
	skillID := skill.GetMetadata().ID

	// Create memories that reference both - using index properly
	for i := range 3 {
		mem := domain.NewMemory(fmt.Sprintf("Memory%d", i), "Test", "1.0.0", "test")
		mem.Metadata["related_to"] = personaID + "," + skillID
		repo.Create(mem)
		memID := mem.GetMetadata().ID
		// Add to index - forward direction (memory -> related elements)
		index.Add(memID, []string{personaID, skillID})
	}

	ctx := context.Background()
	options := RecommendationOptions{
		MaxResults: 10,
		MinScore:   0.1,
	}

	recommendations, err := engine.RecommendForElement(ctx, personaID, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should recommend skill due to co-occurrence
	found := false
	for _, rec := range recommendations {
		if rec.ElementID == skillID {
			found = true
			if !containsString(rec.Reasons, "frequently co-occurs") {
				t.Errorf("Expected 'frequently co-occurs' reason, got: %v", rec.Reasons)
			}
		}
	}

	if !found {
		t.Error("Expected skill to be recommended due to co-occurrence")
	}
}

func TestRecommendForElement_TagSimilarity(t *testing.T) {
	repo := newMockRepoForIndex()
	index := NewRelationshipIndex()
	engine := NewRecommendationEngine(repo, index)

	// Create persona with tags
	persona1 := domain.NewPersona("Persona 1", "Test", "1.0.0", "test")
	meta1 := persona1.GetMetadata()
	meta1.Tags = []string{"python", "backend", "api"}
	persona1.SetMetadata(meta1)

	// Create another persona with similar tags
	persona2 := domain.NewPersona("Persona 2", "Test", "1.0.0", "test")
	meta2 := persona2.GetMetadata()
	meta2.Tags = []string{"python", "backend", "web"}
	persona2.SetMetadata(meta2)

	repo.Create(persona1)
	repo.Create(persona2)

	ctx := context.Background()
	options := RecommendationOptions{
		MaxResults: 10,
		MinScore:   0.1,
	}

	recommendations, err := engine.RecommendForElement(ctx, persona1.GetMetadata().ID, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should recommend persona2 due to tag similarity
	found := false
	for _, rec := range recommendations {
		if rec.ElementID == persona2.GetMetadata().ID {
			found = true
			if !containsString(rec.Reasons, "similar tags") {
				t.Error("Expected 'similar tags' reason")
			}
		}
	}

	if !found {
		t.Error("Expected persona2 to be recommended due to tag similarity")
	}
}

func TestRecommendForElement_TypeBasedRecommendations(t *testing.T) {
	repo := newMockRepoForIndex()
	index := NewRelationshipIndex()
	engine := NewRecommendationEngine(repo, index)

	// Create persona (should recommend skills)
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill := domain.NewSkill("Python", "Python programming", "1.0.0", "test")

	repo.Create(persona)
	repo.Create(skill)

	ctx := context.Background()
	skillType := domain.SkillElement
	options := RecommendationOptions{
		ElementType: &skillType,
		MaxResults:  10,
		MinScore:    0.1,
	}

	recommendations, err := engine.RecommendForElement(ctx, persona.GetMetadata().ID, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should recommend skill based on type pattern
	found := false
	for _, rec := range recommendations {
		if rec.ElementType == domain.SkillElement {
			found = true
		}
	}

	if !found {
		t.Error("Expected skill to be recommended based on type pattern")
	}
}

func TestRecommendForElement_FilterByType(t *testing.T) {
	repo := newMockRepoForIndex()
	index := NewRelationshipIndex()
	engine := NewRecommendationEngine(repo, index)

	// Create persona with related skills and templates
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill := domain.NewSkill("Python", "Python programming", "1.0.0", "test")
	template := domain.NewTemplate("Test Template", "Test", "1.0.0", "test")

	persona.RelatedSkills = []string{skill.GetMetadata().ID}
	persona.RelatedTemplates = []string{template.GetMetadata().ID}

	repo.Create(persona)
	repo.Create(skill)
	repo.Create(template)

	ctx := context.Background()
	skillType := domain.SkillElement
	options := RecommendationOptions{
		ElementType: &skillType,
		MaxResults:  10,
		MinScore:    0.1,
	}

	recommendations, err := engine.RecommendForElement(ctx, persona.GetMetadata().ID, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should only recommend skills, not templates
	for _, rec := range recommendations {
		if rec.ElementType != domain.SkillElement {
			t.Errorf("Expected only skills, got: %s", rec.ElementType)
		}
	}
}

func TestRecommendForElement_ExcludeIDs(t *testing.T) {
	repo := newMockRepoForIndex()
	index := NewRelationshipIndex()
	engine := NewRecommendationEngine(repo, index)

	// Create persona with related skills
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill1 := domain.NewSkill("Python", "Python programming", "1.0.0", "test")
	skill2 := domain.NewSkill("Go", "Go programming", "1.0.0", "test")

	persona.RelatedSkills = []string{skill1.GetMetadata().ID, skill2.GetMetadata().ID}

	repo.Create(persona)
	repo.Create(skill1)
	repo.Create(skill2)

	ctx := context.Background()
	options := RecommendationOptions{
		ExcludeIDs: []string{skill1.GetMetadata().ID},
		MaxResults: 10,
		MinScore:   0.1,
	}

	recommendations, err := engine.RecommendForElement(ctx, persona.GetMetadata().ID, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should not recommend excluded skill
	for _, rec := range recommendations {
		if rec.ElementID == skill1.GetMetadata().ID {
			t.Error("Should not recommend excluded element")
		}
	}
}

func TestRecommendForElement_MinScore(t *testing.T) {
	repo := newMockRepoForIndex()
	index := NewRelationshipIndex()
	engine := NewRecommendationEngine(repo, index)

	// Create persona with related skill
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill := domain.NewSkill("Python", "Python programming", "1.0.0", "test")

	persona.RelatedSkills = []string{skill.GetMetadata().ID}

	repo.Create(persona)
	repo.Create(skill)

	ctx := context.Background()
	options := RecommendationOptions{
		MinScore:   0.5, // Direct relationships have score 1.0
		MaxResults: 10,
	}

	recommendations, err := engine.RecommendForElement(ctx, persona.GetMetadata().ID, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should include skill with score 1.0
	if len(recommendations) == 0 {
		t.Error("Expected at least one recommendation above min score")
	}

	for _, rec := range recommendations {
		if rec.Score < options.MinScore {
			t.Errorf("Recommendation score %f below minimum %f", rec.Score, options.MinScore)
		}
	}
}

func TestRecommendForElement_MaxResults(t *testing.T) {
	repo := newMockRepoForIndex()
	index := NewRelationshipIndex()
	engine := NewRecommendationEngine(repo, index)

	// Create persona with many related skills
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")

	for range 20 {
		skill := domain.NewSkill("Skill", "Test", "1.0.0", "test")
		skillID := skill.GetMetadata().ID
		persona.RelatedSkills = append(persona.RelatedSkills, skillID)
		repo.Create(skill)
	}

	repo.Create(persona)

	ctx := context.Background()
	options := RecommendationOptions{
		MaxResults: 5,
		MinScore:   0.1,
	}

	recommendations, err := engine.RecommendForElement(ctx, persona.GetMetadata().ID, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(recommendations) > options.MaxResults {
		t.Errorf("Expected max %d results, got: %d", options.MaxResults, len(recommendations))
	}
}

func TestCalculateTagSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		tags1    []string
		tags2    []string
		expected float64
	}{
		{
			name:     "identical tags",
			tags1:    []string{"python", "backend"},
			tags2:    []string{"python", "backend"},
			expected: 1.0,
		},
		{
			name:     "no overlap",
			tags1:    []string{"python", "backend"},
			tags2:    []string{"java", "frontend"},
			expected: 0.0,
		},
		{
			name:     "partial overlap",
			tags1:    []string{"python", "backend", "api"},
			tags2:    []string{"python", "frontend", "web"},
			expected: 0.2, // 1 intersection / 5 union
		},
		{
			name:     "empty tags",
			tags1:    []string{},
			tags2:    []string{"python"},
			expected: 0.0,
		},
		{
			name:     "case insensitive",
			tags1:    []string{"Python", "Backend"},
			tags2:    []string{"python", "backend"},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := calculateTagSimilarity(tt.tags1, tt.tags2)
			if similarity != tt.expected {
				t.Errorf("Expected similarity %f, got: %f", tt.expected, similarity)
			}
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", "d"}
	result := uniqueStrings(input)

	if len(result) != 4 {
		t.Errorf("Expected 4 unique strings, got: %d", len(result))
	}

	// Check no duplicates
	seen := make(map[string]bool)
	for _, s := range result {
		if seen[s] {
			t.Errorf("Found duplicate: %s", s)
		}
		seen[s] = true
	}
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
