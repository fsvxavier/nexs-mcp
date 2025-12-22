package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// createMCPServerWithIndex creates a test MCP server with relationship index.
func createMCPServerWithIndex(repo *mockRepoForMCP) *MCPServer {
	idx := application.NewRelationshipIndex()
	return &MCPServer{
		repo:              repo,
		relationshipIndex: idx,
	}
}

// setupTestMemoriesForSearch creates test memories for search tests.
func setupTestMemoriesForSearch(repo *mockRepoForMCP, personaID string) (string, string, string) {
	// Memory 1: References persona, tagged "work", "discussion"
	mem1 := domain.NewMemory("Work Discussion", "Discussion about work", "1.0.0", "alice")
	mem1.Metadata["related_to"] = personaID
	meta1 := mem1.GetMetadata()
	meta1.Tags = []string{"work", "discussion"}
	mem1.SetMetadata(meta1)
	repo.Create(mem1)
	mem1ID := mem1.GetMetadata().ID

	// Memory 2: References persona, tagged "personal"
	mem2 := domain.NewMemory("Personal Note", "Personal notes", "1.0.0", "bob")
	mem2.Metadata["related_to"] = personaID
	meta2 := mem2.GetMetadata()
	meta2.Tags = []string{"personal"}
	mem2.SetMetadata(meta2)
	repo.Create(mem2)
	mem2ID := mem2.GetMetadata().ID

	// Memory 3: References persona, tagged "work", "urgent"
	mem3 := domain.NewMemory("Urgent Task", "Urgent work task", "1.0.0", "alice")
	mem3.Metadata["related_to"] = personaID
	meta3 := mem3.GetMetadata()
	meta3.Tags = []string{"work", "urgent"}
	mem3.SetMetadata(meta3)
	repo.Create(mem3)
	mem3ID := mem3.GetMetadata().ID

	return mem1ID, mem2ID, mem3ID
}

func TestHandleFindRelatedMemories_Success(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	// Create persona
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Create memories referencing persona
	setupTestMemoriesForSearch(repo, personaID)

	// Rebuild index
	ctx := context.Background()
	err := server.relationshipIndex.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	// Test find related memories
	input := FindRelatedMemoriesInput{
		ElementID: personaID,
	}

	result, output, err := server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	// Verify output
	if output.ElementID != personaID {
		t.Errorf("Expected element_id=%s, got: %s", personaID, output.ElementID)
	}

	if output.ElementType != "persona" {
		t.Errorf("Expected element_type=persona, got: %s", output.ElementType)
	}

	if output.TotalMemories != 3 {
		t.Errorf("Expected 3 memories, got: %d", output.TotalMemories)
	}

	if len(output.Memories) != 3 {
		t.Errorf("Expected 3 memories in array, got: %d", len(output.Memories))
	}

	if output.SearchDuration < 0 {
		t.Errorf("Expected non-negative search duration, got: %d", output.SearchDuration)
	}
}

func TestHandleFindRelatedMemories_MissingElementID(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	input := FindRelatedMemoriesInput{
		ElementID: "",
	}

	ctx := context.Background()
	_, _, err := server.handleFindRelatedMemories(ctx, nil, input)

	if err == nil {
		t.Fatal("Expected error for missing element_id")
	}

	if err.Error() != "element_id is required" {
		t.Errorf("Expected 'element_id is required', got: %v", err)
	}
}

func TestHandleFindRelatedMemories_ElementNotFound(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	input := FindRelatedMemoriesInput{
		ElementID: "nonexistent",
	}

	ctx := context.Background()
	_, _, err := server.handleFindRelatedMemories(ctx, nil, input)

	if err == nil {
		t.Fatal("Expected error for nonexistent element")
	}
}

func TestHandleFindRelatedMemories_NoRelatedMemories(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	// Create persona with no memories
	persona := domain.NewPersona("Unused Persona", "Test", "1.0.0", "test")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Rebuild index
	ctx := context.Background()
	err := server.relationshipIndex.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	input := FindRelatedMemoriesInput{
		ElementID: personaID,
	}

	_, output, err := server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.TotalMemories != 0 {
		t.Errorf("Expected 0 memories, got: %d", output.TotalMemories)
	}

	if len(output.Memories) != 0 {
		t.Errorf("Expected empty memories array, got: %d", len(output.Memories))
	}
}

func TestHandleFindRelatedMemories_FilterByAuthor(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	// Create persona
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Create memories with different authors
	setupTestMemoriesForSearch(repo, personaID)

	// Rebuild index
	ctx := context.Background()
	err := server.relationshipIndex.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	// Filter by author "alice"
	input := FindRelatedMemoriesInput{
		ElementID: personaID,
		Author:    "alice",
	}

	_, output, err := server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only get memories by alice (2 memories)
	if output.TotalMemories != 2 {
		t.Errorf("Expected 2 memories by alice, got: %d", output.TotalMemories)
	}
}

func TestHandleFindRelatedMemories_FilterByIncludeTags(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	// Create persona
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Create memories with different tags
	setupTestMemoriesForSearch(repo, personaID)

	// Rebuild index
	ctx := context.Background()
	err := server.relationshipIndex.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	// Filter by tag "work" (should get 2 memories)
	input := FindRelatedMemoriesInput{
		ElementID:   personaID,
		IncludeTags: []string{"work"},
	}

	_, output, err := server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.TotalMemories != 2 {
		t.Errorf("Expected 2 memories with 'work' tag, got: %d", output.TotalMemories)
	}

	// Filter by tags "work" AND "urgent" (should get 1 memory)
	input.IncludeTags = []string{"work", "urgent"}

	_, output, err = server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.TotalMemories != 1 {
		t.Errorf("Expected 1 memory with both tags, got: %d", output.TotalMemories)
	}
}

func TestHandleFindRelatedMemories_FilterByExcludeTags(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	// Create persona
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Create memories with different tags
	setupTestMemoriesForSearch(repo, personaID)

	// Rebuild index
	ctx := context.Background()
	err := server.relationshipIndex.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	// Exclude tag "urgent" (should get 2 memories)
	input := FindRelatedMemoriesInput{
		ElementID:   personaID,
		ExcludeTags: []string{"urgent"},
	}

	_, output, err := server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.TotalMemories != 2 {
		t.Errorf("Expected 2 memories without 'urgent' tag, got: %d", output.TotalMemories)
	}
}

func TestHandleFindRelatedMemories_SortByName(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	// Create persona
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Create memories
	setupTestMemoriesForSearch(repo, personaID)

	// Rebuild index
	ctx := context.Background()
	err := server.relationshipIndex.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	// Sort by name ascending
	input := FindRelatedMemoriesInput{
		ElementID: personaID,
		SortBy:    "name",
		SortOrder: "asc",
	}

	_, output, err := server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify sorted order
	if len(output.Memories) >= 2 {
		name1 := output.Memories[0]["name"].(string)
		name2 := output.Memories[1]["name"].(string)
		if name1 > name2 {
			t.Errorf("Expected ascending order, got %s before %s", name1, name2)
		}
	}
}

func TestHandleFindRelatedMemories_WithLimit(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	// Create persona
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Create 5 memories with unique names (to avoid ID collision)
	for i := range 5 {
		mem := domain.NewMemory(fmt.Sprintf("Memory %d", i), "Test", "1.0.0", "test")
		mem.Metadata["related_to"] = personaID
		repo.Create(mem)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Rebuild index
	ctx := context.Background()
	err := server.relationshipIndex.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	// Limit to 2 memories
	input := FindRelatedMemoriesInput{
		ElementID: personaID,
		Limit:     2,
	}

	_, output, err := server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if output.TotalMemories != 2 {
		t.Errorf("Expected 2 memories due to limit, got: %d", output.TotalMemories)
	}
}

func TestHandleFindRelatedMemories_IndexStats(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	// Create persona
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Create memories
	setupTestMemoriesForSearch(repo, personaID)

	// Rebuild index
	ctx := context.Background()
	err := server.relationshipIndex.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	input := FindRelatedMemoriesInput{
		ElementID: personaID,
	}

	_, output, err := server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify index stats are present
	if output.IndexStats == nil {
		t.Fatal("Expected index stats in output")
	}

	if _, ok := output.IndexStats["forward_entries"]; !ok {
		t.Error("Expected forward_entries in index stats")
	}
	if _, ok := output.IndexStats["reverse_entries"]; !ok {
		t.Error("Expected reverse_entries in index stats")
	}
	if _, ok := output.IndexStats["cache_hits"]; !ok {
		t.Error("Expected cache_hits in index stats")
	}
}

func TestHandleFindRelatedMemories_JSONSerialization(t *testing.T) {
	repo := newMockRepoForMCP()
	server := createMCPServerWithIndex(repo)

	// Create persona
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	repo.Create(persona)
	personaID := persona.GetMetadata().ID

	// Create memory
	mem := domain.NewMemory("Test Memory", "Test", "1.0.0", "test")
	mem.Metadata["related_to"] = personaID
	repo.Create(mem)

	// Rebuild index
	ctx := context.Background()
	err := server.relationshipIndex.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	input := FindRelatedMemoriesInput{
		ElementID: personaID,
	}

	_, output, err := server.handleFindRelatedMemories(ctx, nil, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify JSON serialization
	jsonData, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal output: %v", err)
	}

	var unmarshaled FindRelatedMemoriesOutput
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if unmarshaled.ElementID != output.ElementID {
		t.Error("JSON serialization changed element_id")
	}
}

func TestApplyMemoryFilters(t *testing.T) {
	// Create test memories with tags using SetMetadata
	mem1 := domain.NewMemory("Memory 1", "Test", "1.0.0", "alice")
	meta1 := mem1.GetMetadata()
	meta1.Tags = []string{"work", "important"}
	mem1.SetMetadata(meta1)

	mem2 := domain.NewMemory("Memory 2", "Test", "1.0.0", "bob")
	meta2 := mem2.GetMetadata()
	meta2.Tags = []string{"personal"}
	mem2.SetMetadata(meta2)

	mem3 := domain.NewMemory("Memory 3", "Test", "1.0.0", "alice")
	meta3 := mem3.GetMetadata()
	meta3.Tags = []string{"work"}
	mem3.SetMetadata(meta3)

	memories := []*domain.Memory{mem1, mem2, mem3}

	t.Run("filter by author", func(t *testing.T) {
		input := FindRelatedMemoriesInput{
			Author: "alice",
		}
		filtered := applyMemoryFilters(memories, input)
		if len(filtered) != 2 {
			t.Errorf("Expected 2 memories by alice, got: %d", len(filtered))
		}
	})

	t.Run("filter by include tags", func(t *testing.T) {
		input := FindRelatedMemoriesInput{
			IncludeTags: []string{"work"},
		}
		filtered := applyMemoryFilters(memories, input)
		if len(filtered) != 2 {
			t.Errorf("Expected 2 memories with 'work' tag, got: %d", len(filtered))
		}
	})

	t.Run("filter by exclude tags", func(t *testing.T) {
		input := FindRelatedMemoriesInput{
			ExcludeTags: []string{"personal"},
		}
		filtered := applyMemoryFilters(memories, input)
		if len(filtered) != 2 {
			t.Errorf("Expected 2 memories without 'personal' tag, got: %d", len(filtered))
		}
	})
}

func TestSortMemories(t *testing.T) {
	// Create memories with different timestamps and names
	mem1 := domain.NewMemory("Charlie", "Test", "1.0.0", "test")
	time.Sleep(1 * time.Millisecond)
	mem2 := domain.NewMemory("Alice", "Test", "1.0.0", "test")
	time.Sleep(1 * time.Millisecond)
	mem3 := domain.NewMemory("Bob", "Test", "1.0.0", "test")

	memories := []*domain.Memory{mem1, mem2, mem3}

	t.Run("sort by name asc", func(t *testing.T) {
		sortMemories(memories, "name", "asc")
		if memories[0].GetMetadata().Name != "Alice" {
			t.Error("Expected Alice first")
		}
		if memories[2].GetMetadata().Name != "Charlie" {
			t.Error("Expected Charlie last")
		}
	})

	t.Run("sort by name desc", func(t *testing.T) {
		sortMemories(memories, "name", "desc")
		if memories[0].GetMetadata().Name != "Charlie" {
			t.Error("Expected Charlie first in desc order")
		}
	})

	t.Run("sort by created_at desc", func(t *testing.T) {
		sortMemories(memories, "created_at", "desc")
		// mem3 should be first (most recent)
		if memories[0].GetMetadata().Name != "Bob" {
			t.Error("Expected most recent memory first")
		}
	})
}

func TestHasAllTags(t *testing.T) {
	tags := []string{"work", "important", "urgent"}

	if !hasAllTags(tags, []string{"work"}) {
		t.Error("Expected true for single required tag")
	}

	if !hasAllTags(tags, []string{"work", "important"}) {
		t.Error("Expected true for multiple required tags")
	}

	if hasAllTags(tags, []string{"work", "missing"}) {
		t.Error("Expected false when missing required tag")
	}

	if !hasAllTags(tags, []string{}) {
		t.Error("Expected true for empty required tags")
	}
}

func TestHasAnyTag(t *testing.T) {
	tags := []string{"work", "important", "urgent"}

	if !hasAnyTag(tags, []string{"work"}) {
		t.Error("Expected true for matching tag")
	}

	if !hasAnyTag(tags, []string{"personal", "work"}) {
		t.Error("Expected true for any matching tag")
	}

	if hasAnyTag(tags, []string{"personal", "home"}) {
		t.Error("Expected false when no tags match")
	}

	if hasAnyTag(tags, []string{}) {
		t.Error("Expected false for empty excluded tags")
	}
}
