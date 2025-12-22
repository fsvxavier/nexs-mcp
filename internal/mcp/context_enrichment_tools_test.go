package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// mockRepoForMCP implements domain.ElementRepository for MCP testing.
type mockRepoForMCP struct {
	elements map[string]domain.Element
}

func newMockRepoForMCP() *mockRepoForMCP {
	return &mockRepoForMCP{
		elements: make(map[string]domain.Element),
	}
}

func (m *mockRepoForMCP) Create(elem domain.Element) error {
	m.elements[elem.GetMetadata().ID] = elem
	return nil
}

func (m *mockRepoForMCP) GetByID(id string) (domain.Element, error) {
	elem, exists := m.elements[id]
	if !exists {
		return nil, domain.ErrElementNotFound
	}
	return elem, nil
}

func (m *mockRepoForMCP) Update(elem domain.Element) error {
	m.elements[elem.GetMetadata().ID] = elem
	return nil
}

func (m *mockRepoForMCP) Delete(id string) error {
	delete(m.elements, id)
	return nil
}

func (m *mockRepoForMCP) List(filter domain.ElementFilter) ([]domain.Element, error) {
	result := make([]domain.Element, 0, len(m.elements))
	for _, elem := range m.elements {
		result = append(result, elem)
	}
	return result, nil
}

func (m *mockRepoForMCP) Exists(id string) (bool, error) {
	_, exists := m.elements[id]
	return exists, nil
}

// createMCPServerForTest creates a test MCP server.
func createMCPServerForTest(repo domain.ElementRepository) *MCPServer {
	return &MCPServer{
		repo: repo,
	}
}

// createTestMemoryWithRelations creates a memory with related elements.
func createTestMemoryWithRelations(repo *mockRepoForMCP) (string, string, string) {
	// Create persona
	persona := domain.NewPersona("Test Persona", "Test persona description", "1.0.0", "test-author")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Create skill
	skill := domain.NewSkill("Test Skill", "Test skill description", "1.0.0", "test-author")
	repo.Create(skill)
	skillID := skill.GetMetadata().ID

	// Create memory with related elements
	memory := domain.NewMemory("Test Memory", "Test memory description", "1.0.0", "test-author")
	memory.Metadata["related_to"] = personaID + "," + skillID
	repo.Create(memory)
	memoryID := memory.GetMetadata().ID

	return memoryID, personaID, skillID
}

func TestHandleExpandMemoryContext_Success(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerForTest(repo)

	memoryID, personaID, skillID := createTestMemoryWithRelations(repo)

	input := ExpandMemoryContextInput{
		MemoryID: memoryID,
	}

	ctx := context.Background()
	result, output, err := server.handleExpandMemoryContext(ctx, nil, input)

	// Verify no error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify result is nil (MCP convention)
	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	// Verify output structure
	if output.Memory == nil {
		t.Fatal("Expected memory in output")
	}

	if output.Memory["id"] != memoryID {
		t.Errorf("Expected memory ID %s, got: %v", memoryID, output.Memory["id"])
	}

	// Verify related elements
	if len(output.RelatedElements) != 2 {
		t.Errorf("Expected 2 related elements, got: %d", len(output.RelatedElements))
	}

	// Verify IDs are present
	foundPersona := false
	foundSkill := false
	for _, elem := range output.RelatedElements {
		if elem["id"] == personaID {
			foundPersona = true
		}
		if elem["id"] == skillID {
			foundSkill = true
		}
	}

	if !foundPersona {
		t.Errorf("Expected to find persona %s in related elements", personaID)
	}
	if !foundSkill {
		t.Errorf("Expected to find skill %s in related elements", skillID)
	}

	// Verify metrics
	if output.TotalElements != 2 {
		t.Errorf("Expected TotalElements=2, got: %d", output.TotalElements)
	}

	if output.TokensSaved <= 0 {
		t.Errorf("Expected positive tokens saved, got: %d", output.TokensSaved)
	}

	if output.FetchDurationMs < 0 {
		t.Errorf("Expected non-negative fetch duration, got: %d", output.FetchDurationMs)
	}
}

func TestHandleExpandMemoryContext_MissingMemoryID(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerForTest(repo)

	input := ExpandMemoryContextInput{
		MemoryID: "",
	}

	ctx := context.Background()
	_, _, err := server.handleExpandMemoryContext(ctx, nil, input)

	if err == nil {
		t.Fatal("Expected error for missing memory_id")
	}

	if err.Error() != "memory_id is required" {
		t.Errorf("Expected 'memory_id is required' error, got: %v", err)
	}
}

func TestHandleExpandMemoryContext_MemoryNotFound(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerForTest(repo)

	input := ExpandMemoryContextInput{
		MemoryID: "nonexistent",
	}

	ctx := context.Background()
	_, _, err := server.handleExpandMemoryContext(ctx, nil, input)

	if err == nil {
		t.Fatal("Expected error for nonexistent memory")
	}
}

func TestHandleExpandMemoryContext_NotAMemory(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerForTest(repo)

	// Save a persona (not a memory)
	persona := domain.NewPersona("Test Persona", "Test persona", "1.0.0", "test-author")
	repo.Create(persona)

	input := ExpandMemoryContextInput{
		MemoryID: persona.GetMetadata().ID,
	}

	ctx := context.Background()
	_, _, err := server.handleExpandMemoryContext(ctx, nil, input)

	if err == nil {
		t.Fatal("Expected error when element is not a memory")
	}
}

func TestHandleExpandMemoryContext_WithIncludeTypes(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerForTest(repo)

	memoryID, _, _ := createTestMemoryWithRelations(repo)

	input := ExpandMemoryContextInput{
		MemoryID:     memoryID,
		IncludeTypes: []string{"persona"},
	}

	ctx := context.Background()
	_, output, err := server.handleExpandMemoryContext(ctx, nil, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only have persona, not skill
	if output.TotalElements != 1 {
		t.Errorf("Expected 1 element (persona only), got: %d", output.TotalElements)
	}

	// Verify it's a persona
	if len(output.RelatedElements) > 0 {
		if output.RelatedElements[0]["type"] != "persona" {
			t.Errorf("Expected type=persona, got: %v", output.RelatedElements[0]["type"])
		}
	}
}

func TestHandleExpandMemoryContext_WithExcludeTypes(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerForTest(repo)

	memoryID, _, _ := createTestMemoryWithRelations(repo)

	input := ExpandMemoryContextInput{
		MemoryID:     memoryID,
		ExcludeTypes: []string{"skill"},
	}

	ctx := context.Background()
	_, output, err := server.handleExpandMemoryContext(ctx, nil, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only have persona, not skill
	if output.TotalElements != 1 {
		t.Errorf("Expected 1 element (no skills), got: %d", output.TotalElements)
	}

	// Verify it's a persona
	if len(output.RelatedElements) > 0 {
		if output.RelatedElements[0]["type"] != "persona" {
			t.Errorf("Expected type=persona, got: %v", output.RelatedElements[0]["type"])
		}
	}
}

func TestHandleExpandMemoryContext_InvalidElementType(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerForTest(repo)

	memoryID, _, _ := createTestMemoryWithRelations(repo)

	input := ExpandMemoryContextInput{
		MemoryID:     memoryID,
		IncludeTypes: []string{"invalid_type"},
	}

	ctx := context.Background()
	_, _, err := server.handleExpandMemoryContext(ctx, nil, input)

	if err == nil {
		t.Fatal("Expected error for invalid element type")
	}
}

func TestHandleExpandMemoryContext_WithMaxElements(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerForTest(repo)

	// Create memory with 5 related elements
	persona1 := domain.NewPersona("Persona 1", "Test", "1.0.0", "test-author")
	repo.Create(persona1)
	persona2 := domain.NewPersona("Persona 2", "Test", "1.0.0", "test-author")
	repo.Create(persona2)
	skill1 := domain.NewSkill("Skill 1", "Test", "1.0.0", "test-author")
	repo.Create(skill1)
	skill2 := domain.NewSkill("Skill 2", "Test", "1.0.0", "test-author")
	repo.Create(skill2)
	agent := domain.NewAgent("Agent 1", "Test", "1.0.0", "test-author")
	repo.Create(agent)

	memory := domain.NewMemory("Test Memory", "Test", "1.0.0", "test-author")
	memory.Metadata["related_to"] = persona1.GetMetadata().ID + "," +
		persona2.GetMetadata().ID + "," +
		skill1.GetMetadata().ID + "," +
		skill2.GetMetadata().ID + "," +
		agent.GetMetadata().ID
	repo.Create(memory)
	memoryID := memory.GetMetadata().ID

	input := ExpandMemoryContextInput{
		MemoryID:    memoryID,
		MaxElements: 3,
	}

	ctx := context.Background()
	_, output, err := server.handleExpandMemoryContext(ctx, nil, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should be limited to 3 elements
	if output.TotalElements != 3 {
		t.Errorf("Expected 3 elements (max limit), got: %d", output.TotalElements)
	}
}

func TestHandleExpandMemoryContext_WithIgnoreErrors(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerForTest(repo)

	// Create memory referencing a nonexistent element
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test-author")
	repo.Create(persona)

	memory := domain.NewMemory("Test Memory", "Test", "1.0.0", "test-author")
	memory.Metadata["related_to"] = persona.GetMetadata().ID + ",nonexistent"
	repo.Create(memory)
	memoryID := memory.GetMetadata().ID

	input := ExpandMemoryContextInput{
		MemoryID:     memoryID,
		IgnoreErrors: true,
	}

	ctx := context.Background()
	_, output, err := server.handleExpandMemoryContext(ctx, nil, input)

	// Should not return error when IgnoreErrors=true
	if err != nil {
		t.Fatalf("Expected no error with IgnoreErrors=true, got: %v", err)
	}

	// Should have 1 element (the valid one)
	if output.TotalElements != 1 {
		t.Errorf("Expected 1 element, got: %d", output.TotalElements)
	}

	// Should have errors in output
	if len(output.Errors) == 0 {
		t.Error("Expected errors in output when some elements fail")
	}
}

func TestIsValidElementType(t *testing.T) {
	tests := []struct {
		name     string
		elemType domain.ElementType
		want     bool
	}{
		{"persona", domain.PersonaElement, true},
		{"skill", domain.SkillElement, true},
		{"template", domain.TemplateElement, true},
		{"agent", domain.AgentElement, true},
		{"memory", domain.MemoryElement, true},
		{"ensemble", domain.EnsembleElement, true},
		{"invalid", domain.ElementType("invalid"), false},
		{"empty", domain.ElementType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidElementType(tt.elemType)
			if got != tt.want {
				t.Errorf("isValidElementType(%v) = %v, want %v", tt.elemType, got, tt.want)
			}
		})
	}
}

func TestConvertMemoryToMap(t *testing.T) {
	memory := domain.NewMemory("Test Memory", "Test description", "1.0.0", "test-author")
	memory.Metadata["custom_key"] = "custom_value"

	result := convertMemoryToMap(memory)

	// Verify required fields
	if result["id"] == "" {
		t.Error("Expected id in result")
	}
	if result["type"] != "memory" {
		t.Errorf("Expected type=memory, got: %v", result["type"])
	}
	if result["name"] != "Test Memory" {
		t.Errorf("Expected name='Test Memory', got: %v", result["name"])
	}

	// Verify timestamps are formatted as RFC3339
	createdAt, ok := result["created_at"].(string)
	if !ok {
		t.Fatal("Expected created_at to be string")
	}
	if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
		t.Errorf("Expected RFC3339 timestamp, got: %s", createdAt)
	}

	// Verify metadata is preserved
	metadata, ok := result["metadata"].(map[string]string)
	if !ok {
		t.Fatalf("Expected metadata to be map[string]string, got: %T", result["metadata"])
	}
	if metadata["custom_key"] != "custom_value" {
		t.Errorf("Expected custom_key=custom_value, got: %v", metadata["custom_key"])
	}
}

func TestConvertElementsToMaps(t *testing.T) {
	elements := make(map[string]domain.Element)

	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test-author")
	elements[persona.GetMetadata().ID] = persona

	skill := domain.NewSkill("Test Skill", "Test", "1.0.0", "test-author")
	elements[skill.GetMetadata().ID] = skill

	result := convertElementsToMaps(elements)

	// Verify count
	if len(result) != 2 {
		t.Fatalf("Expected 2 elements, got: %d", len(result))
	}

	// Verify structure (order not guaranteed in map)
	for _, elemMap := range result {
		if elemMap["id"] == "" {
			t.Error("Expected id in element")
		}
		if elemMap["type"] == "" {
			t.Error("Expected type in element")
		}
		if elemMap["name"] == "" {
			t.Error("Expected name in element")
		}

		// Verify timestamps are RFC3339
		createdAt, ok := elemMap["created_at"].(string)
		if !ok {
			t.Error("Expected created_at to be string")
		}
		if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
			t.Errorf("Expected RFC3339 timestamp, got: %s", createdAt)
		}
	}
}

func TestConvertRelationshipMapToStrings(t *testing.T) {
	relMap := make(domain.RelationshipMap)
	relMap.Add("elem1", domain.RelationshipRelatedTo)
	relMap.Add("elem1", domain.RelationshipDependsOn)
	relMap.Add("elem2", domain.RelationshipUses)

	result := convertRelationshipMapToStrings(relMap)

	// Verify elem1 has 2 relationships
	if len(result["elem1"]) != 2 {
		t.Errorf("Expected 2 relationships for elem1, got: %d", len(result["elem1"]))
	}

	// Verify elem2 has 1 relationship
	if len(result["elem2"]) != 1 {
		t.Errorf("Expected 1 relationship for elem2, got: %d", len(result["elem2"]))
	}

	// Verify relationship types are strings
	foundRelatedTo := false
	foundDependsOn := false
	for _, relType := range result["elem1"] {
		if relType == "related_to" {
			foundRelatedTo = true
		}
		if relType == "depends_on" {
			foundDependsOn = true
		}
	}

	if !foundRelatedTo {
		t.Error("Expected to find 'related_to' in elem1 relationships")
	}
	if !foundDependsOn {
		t.Error("Expected to find 'depends_on' in elem1 relationships")
	}
}

func TestExpandMemoryContextOutput_JSONSerialization(t *testing.T) {
	output := ExpandMemoryContextOutput{
		Memory: map[string]interface{}{
			"id":   "test-id",
			"name": "Test Memory",
		},
		RelatedElements: []map[string]interface{}{
			{
				"id":   "elem1",
				"type": "persona",
			},
		},
		RelationshipMap: map[string][]string{
			"elem1": {"related_to"},
		},
		TotalElements:   1,
		TokensSaved:     150,
		FetchDurationMs: 10,
		Errors:          []string{"test error"},
	}

	// Verify it can be marshaled to JSON
	jsonData, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal output to JSON: %v", err)
	}

	// Verify it can be unmarshaled back
	var unmarshaled ExpandMemoryContextOutput
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify key fields
	if unmarshaled.TotalElements != 1 {
		t.Errorf("Expected TotalElements=1, got: %d", unmarshaled.TotalElements)
	}
	if unmarshaled.TokensSaved != 150 {
		t.Errorf("Expected TokensSaved=150, got: %d", unmarshaled.TokensSaved)
	}
}
