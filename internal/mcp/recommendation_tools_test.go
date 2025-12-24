package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

func TestHandleSuggestRelatedElements_Success(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

	// Create persona with related skills
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill1 := domain.NewSkill("Python", "Python programming", "1.0.0", "test")
	skill2 := domain.NewSkill("Go", "Go programming", "1.0.0", "test")

	persona.RelatedSkills = []string{skill1.GetMetadata().ID, skill2.GetMetadata().ID}

	repo.Create(persona)
	repo.Create(skill1)
	repo.Create(skill2)

	ctx := context.Background()
	input := SuggestRelatedElementsInput{
		ElementID:  persona.GetMetadata().ID,
		MaxResults: 10,
	}

	_, output, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output.ElementID != persona.GetMetadata().ID {
		t.Errorf("Expected element_id=%s, got: %s", persona.GetMetadata().ID, output.ElementID)
	}

	if output.TotalFound == 0 {
		t.Error("Expected suggestions to be found")
	}

	if len(output.Suggestions) == 0 {
		t.Error("Expected non-empty suggestions array")
	}

	// Verify suggestion structure
	if len(output.Suggestions) > 0 {
		sug := output.Suggestions[0]
		if _, ok := sug["element_id"]; !ok {
			t.Error("Suggestion missing element_id")
		}
		if _, ok := sug["element_type"]; !ok {
			t.Error("Suggestion missing element_type")
		}
		if _, ok := sug["element_name"]; !ok {
			t.Error("Suggestion missing element_name")
		}
		if _, ok := sug["score"]; !ok {
			t.Error("Suggestion missing score")
		}
		if _, ok := sug["reasons"]; !ok {
			t.Error("Suggestion missing reasons")
		}
	}
}

func TestHandleSuggestRelatedElements_MissingElementID(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

	ctx := context.Background()
	input := SuggestRelatedElementsInput{
		ElementID: "",
	}

	_, _, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err == nil {
		t.Fatal("Expected error for missing element_id")
	}

	if err.Error() != "element_id is required" {
		t.Errorf("Expected 'element_id is required', got: %v", err)
	}
}

func TestHandleSuggestRelatedElements_ElementNotFound(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

	ctx := context.Background()
	input := SuggestRelatedElementsInput{
		ElementID: "nonexistent",
	}

	_, _, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err == nil {
		t.Fatal("Expected error for nonexistent element")
	}
}

func TestHandleSuggestRelatedElements_FilterByType(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

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
	input := SuggestRelatedElementsInput{
		ElementID:   persona.GetMetadata().ID,
		ElementType: "skill",
		MaxResults:  10,
	}

	_, output, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should only suggest skills
	for _, sug := range output.Suggestions {
		if sug["element_type"] != "skill" {
			t.Errorf("Expected only skills, got: %v", sug["element_type"])
		}
	}
}

func TestHandleSuggestRelatedElements_ExcludeIDs(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

	// Create persona with related skills
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill1 := domain.NewSkill("Python", "Python programming", "1.0.0", "test")
	skill2 := domain.NewSkill("Go", "Go programming", "1.0.0", "test")

	skill1ID := skill1.GetMetadata().ID
	skill2ID := skill2.GetMetadata().ID

	persona.RelatedSkills = []string{skill1ID, skill2ID}

	repo.Create(persona)
	repo.Create(skill1)
	repo.Create(skill2)

	ctx := context.Background()
	input := SuggestRelatedElementsInput{
		ElementID:  persona.GetMetadata().ID,
		ExcludeIDs: []string{skill1ID},
		MaxResults: 10,
	}

	_, output, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should not suggest excluded skill
	for _, sug := range output.Suggestions {
		if sug["element_id"] == skill1ID {
			t.Error("Should not suggest excluded element")
		}
	}
}

func TestHandleSuggestRelatedElements_MinScore(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

	// Create persona with related skill
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill := domain.NewSkill("Python", "Python programming", "1.0.0", "test")

	persona.RelatedSkills = []string{skill.GetMetadata().ID}

	repo.Create(persona)
	repo.Create(skill)

	ctx := context.Background()
	input := SuggestRelatedElementsInput{
		ElementID:  persona.GetMetadata().ID,
		MinScore:   0.5,
		MaxResults: 10,
	}

	_, output, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// All suggestions should have score >= min_score
	for _, sug := range output.Suggestions {
		score, ok := sug["score"].(float64)
		if !ok {
			t.Error("Score is not float64")
			continue
		}
		if score < input.MinScore {
			t.Errorf("Suggestion score %f below minimum %f", score, input.MinScore)
		}
	}
}

func TestHandleSuggestRelatedElements_MaxResults(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

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
	input := SuggestRelatedElementsInput{
		ElementID:  persona.GetMetadata().ID,
		MaxResults: 5,
	}

	_, output, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(output.Suggestions) > input.MaxResults {
		t.Errorf("Expected max %d suggestions, got: %d", input.MaxResults, len(output.Suggestions))
	}
}

func TestHandleSuggestRelatedElements_InvalidElementType(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)

	ctx := context.Background()
	input := SuggestRelatedElementsInput{
		ElementID:   persona.GetMetadata().ID,
		ElementType: "invalid_type",
	}

	_, _, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err == nil {
		t.Fatal("Expected error for invalid element_type")
	}
}

func TestHandleSuggestRelatedElements_JSONSerialization(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

	// Create persona with related skill
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	skill := domain.NewSkill("Python", "Python programming", "1.0.0", "test")

	persona.RelatedSkills = []string{skill.GetMetadata().ID}

	repo.Create(persona)
	repo.Create(skill)

	ctx := context.Background()
	input := SuggestRelatedElementsInput{
		ElementID:  persona.GetMetadata().ID,
		MaxResults: 10,
	}

	_, output, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal output: %v", err)
	}

	var unmarshaled SuggestRelatedElementsOutput
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if unmarshaled.ElementID != output.ElementID {
		t.Error("JSON serialization changed element_id")
	}
}

func TestHandleSuggestRelatedElements_SearchDuration(t *testing.T) {
	repo := newMockRepoForMCP()
	index := application.NewRelationshipIndex()
	server := &MCPServer{
		repo:              repo,
		relationshipIndex: index,
	}

	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)

	ctx := context.Background()
	input := SuggestRelatedElementsInput{
		ElementID:  persona.GetMetadata().ID,
		MaxResults: 10,
	}

	_, output, err := server.handleSuggestRelatedElements(ctx, nil, input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output.SearchDuration < 0 {
		t.Errorf("Expected non-negative search duration, got: %d", output.SearchDuration)
	}
}
