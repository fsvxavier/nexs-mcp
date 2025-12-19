package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

func setupMemoryTestServer(t *testing.T) *MCPServer {
	t.Helper()
	repo := infrastructure.NewInMemoryElementRepository()
	return NewMCPServer("nexs-mcp-test", "0.1.0", repo)
}

func createTestMemory(name, content, author, dateCreated string) *domain.Memory {
	memory := domain.NewMemory(name, "Test memory", "1.0.0", author)
	memory.Content = content
	memory.DateCreated = dateCreated
	memory.ComputeHash()
	return memory
}

func TestHandleSearchMemory(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create test memories
	mem1 := createTestMemory("Meeting Notes", "Discussed project timeline and milestones", "alice", "2025-12-15")
	mem2 := createTestMemory("Project Plan", "Timeline for Q1 2026 project delivery", "bob", "2025-12-16")
	mem3 := createTestMemory("Code Review", "Reviewed pull request #123 for authentication", "alice", "2025-12-17")
	mem4 := createTestMemory("Bug Fix", "Fixed authentication bug in login module", "charlie", "2025-12-18")

	server.repo.Create(mem1)
	server.repo.Create(mem2)
	server.repo.Create(mem3)
	server.repo.Create(mem4)

	tests := []struct {
		name          string
		query         string
		author        string
		expectedCount int
		expectedFirst string // Expected first result name
	}{
		{
			name:          "Search by timeline",
			query:         "timeline",
			expectedCount: 2,
			expectedFirst: "Project Plan", // More recent (2025-12-16 vs 2025-12-15)
		},
		{
			name:          "Search by authentication",
			query:         "authentication",
			expectedCount: 2,
			expectedFirst: "Bug Fix", // More recent
		},
		{
			name:          "Search by author",
			query:         "project",
			author:        "alice",
			expectedCount: 1,
			expectedFirst: "Meeting Notes",
		},
		{
			name:          "No results",
			query:         "nonexistent",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := SearchMemoryInput{
				Query:  tt.query,
				Author: tt.author,
				Limit:  10,
			}

			_, output, err := server.handleSearchMemory(ctx, &sdk.CallToolRequest{}, input)

			if err != nil {
				t.Fatalf("handleSearchMemory failed: %v", err)
			}

			if output.Total != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, output.Total)
			}

			if tt.expectedCount > 0 && output.Memories[0].Name != tt.expectedFirst {
				t.Errorf("Expected first result '%s', got '%s'", tt.expectedFirst, output.Memories[0].Name)
			}

			if output.Query != tt.query {
				t.Errorf("Expected query '%s', got '%s'", tt.query, output.Query)
			}
		})
	}
}

func TestHandleSearchMemory_DateFiltering(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create memories with different dates
	mem1 := createTestMemory("Old Memory", "Old content", "alice", "2025-01-15")
	mem2 := createTestMemory("Recent Memory", "Recent content", "alice", "2025-12-15")
	mem3 := createTestMemory("Latest Memory", "Latest content", "alice", "2025-12-18")

	server.repo.Create(mem1)
	server.repo.Create(mem2)
	server.repo.Create(mem3)

	// Search with date range
	input := SearchMemoryInput{
		Query:    "content",
		DateFrom: "2025-12-01",
		DateTo:   "2025-12-31",
		Limit:    10,
	}

	_, output, err := server.handleSearchMemory(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleSearchMemory failed: %v", err)
	}

	if output.Total != 2 {
		t.Errorf("Expected 2 results (December only), got %d", output.Total)
	}

	// Should get most recent first
	if output.Total > 0 && output.Memories[0].DateCreated != "2025-12-18" {
		t.Errorf("Expected most recent memory first, got date %s", output.Memories[0].DateCreated)
	}
}

func TestHandleSearchMemory_EmptyQuery(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	input := SearchMemoryInput{
		Query: "",
	}

	_, _, err := server.handleSearchMemory(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error for empty query")
	}

	if err.Error() != "query is required" {
		t.Errorf("Expected 'query is required' error, got: %v", err)
	}
}

func TestHandleSummarizeMemories(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create test memories
	mem1 := createTestMemory("Memory 1", "Content 1 with some text", "alice", "2025-12-15")
	mem2 := createTestMemory("Memory 2", "Content 2 with more text here", "bob", "2025-12-16")
	mem3 := createTestMemory("Memory 3", "Content 3", "alice", "2025-12-17")

	server.repo.Create(mem1)
	server.repo.Create(mem2)
	server.repo.Create(mem3)

	input := SummarizeMemoriesInput{
		MaxItems: 50,
	}

	_, output, err := server.handleSummarizeMemories(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleSummarizeMemories failed: %v", err)
	}

	if output.TotalCount != 3 {
		t.Errorf("Expected 3 memories, got %d", output.TotalCount)
	}

	if output.Statistics.TotalMemories != 3 {
		t.Errorf("Expected 3 total memories in stats, got %d", output.Statistics.TotalMemories)
	}

	if output.Statistics.ActiveMemories != 3 {
		t.Errorf("Expected 3 active memories in stats, got %d", output.Statistics.ActiveMemories)
	}

	if output.Statistics.TotalSize == 0 {
		t.Error("Expected non-zero total size")
	}

	if output.Statistics.AverageSize == 0 {
		t.Error("Expected non-zero average size")
	}

	if len(output.TopAuthors) == 0 {
		t.Error("Expected at least one top author")
	}

	if output.RecentMemory == nil {
		t.Error("Expected recent memory to be set")
	} else if output.RecentMemory.DateCreated != "2025-12-17" {
		t.Errorf("Expected most recent memory from 2025-12-17, got %s", output.RecentMemory.DateCreated)
	}

	if output.DateRange == "" {
		t.Error("Expected non-empty date range")
	}
}

func TestHandleSummarizeMemories_AuthorFilter(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create memories from different authors
	mem1 := createTestMemory("Alice Memory 1", "Content", "alice", "2025-12-15")
	mem2 := createTestMemory("Bob Memory", "Content", "bob", "2025-12-16")
	mem3 := createTestMemory("Alice Memory 2", "Content", "alice", "2025-12-17")

	server.repo.Create(mem1)
	server.repo.Create(mem2)
	server.repo.Create(mem3)

	input := SummarizeMemoriesInput{
		Author:   "alice",
		MaxItems: 50,
	}

	_, output, err := server.handleSummarizeMemories(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleSummarizeMemories failed: %v", err)
	}

	if output.TotalCount != 2 {
		t.Errorf("Expected 2 memories by alice, got %d", output.TotalCount)
	}

	if len(output.TopAuthors) != 1 || output.TopAuthors[0] != "alice" {
		t.Errorf("Expected top author to be alice, got %v", output.TopAuthors)
	}
}

func TestHandleUpdateMemory(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create initial memory
	memory := createTestMemory("Original Name", "Original content", "alice", "2025-12-15")
	server.repo.Create(memory)
	memoryID := memory.GetID()

	// Update memory
	input := UpdateMemoryInput{
		ID:          memoryID,
		Content:     "Updated content with new information",
		Name:        "Updated Name",
		Description: "Updated description",
		Tags:        []string{"updated", "test"},
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	_, output, err := server.handleUpdateMemory(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleUpdateMemory failed: %v", err)
	}

	if output.Memory.ID != memoryID {
		t.Errorf("Expected memory ID %s, got %s", memoryID, output.Memory.ID)
	}

	if output.Memory.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", output.Memory.Name)
	}

	if output.Memory.Content != "Updated content with new information" {
		t.Errorf("Expected updated content, got %s", output.Memory.Content)
	}

	// Verify in repository
	elem, _ := server.repo.GetByID(memoryID)
	updatedMemory := elem.(*domain.Memory)

	if updatedMemory.Content != "Updated content with new information" {
		t.Error("Content was not updated in repository")
	}

	if updatedMemory.Metadata["key1"] != "value1" {
		t.Error("Metadata was not updated in repository")
	}

	meta := updatedMemory.GetMetadata()
	if len(meta.Tags) != 2 || meta.Tags[0] != "updated" {
		t.Errorf("Tags were not updated, got %v", meta.Tags)
	}
}

func TestHandleUpdateMemory_EmptyID(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	input := UpdateMemoryInput{
		ID:      "",
		Content: "New content",
	}

	_, _, err := server.handleUpdateMemory(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error for empty ID")
	}

	if err.Error() != "id is required" {
		t.Errorf("Expected 'id is required' error, got: %v", err)
	}
}

func TestHandleUpdateMemory_NonExistent(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	input := UpdateMemoryInput{
		ID:      "non-existent-id",
		Content: "New content",
	}

	_, _, err := server.handleUpdateMemory(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error for non-existent memory")
	}
}

func TestHandleDeleteMemory(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create memory
	memory := createTestMemory("To Delete", "Content", "alice", "2025-12-15")
	server.repo.Create(memory)
	memoryID := memory.GetID()

	// Delete memory
	input := DeleteMemoryInput{
		ID: memoryID,
	}

	_, output, err := server.handleDeleteMemory(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleDeleteMemory failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	if output.ID != memoryID {
		t.Errorf("Expected ID %s, got %s", memoryID, output.ID)
	}

	if output.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Verify deletion
	_, err = server.repo.GetByID(memoryID)
	if err == nil {
		t.Error("Memory should no longer exist in repository")
	}
}

func TestHandleDeleteMemory_EmptyID(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	input := DeleteMemoryInput{
		ID: "",
	}

	_, _, err := server.handleDeleteMemory(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error for empty ID")
	}
}

func TestHandleDeleteMemory_NonMemoryElement(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create a persona (not a memory)
	persona := createTestPersona("Test Persona", "Description")
	server.repo.Create(persona)

	input := DeleteMemoryInput{
		ID: persona.GetID(),
	}

	_, _, err := server.handleDeleteMemory(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error when trying to delete non-memory element")
	}

	if err.Error() != "element is not a memory" {
		t.Errorf("Expected 'element is not a memory' error, got: %v", err)
	}
}

func TestHandleClearMemories(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create test memories
	mem1 := createTestMemory("Memory 1", "Content", "alice", "2025-12-15")
	mem2 := createTestMemory("Memory 2", "Content", "bob", "2025-12-16")
	mem3 := createTestMemory("Memory 3", "Content", "alice", "2025-12-17")

	server.repo.Create(mem1)
	server.repo.Create(mem2)
	server.repo.Create(mem3)

	// Clear all memories by alice
	input := ClearMemoriesInput{
		Author:  "alice",
		Confirm: true,
	}

	_, output, err := server.handleClearMemories(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleClearMemories failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	if output.DeletedCount != 2 {
		t.Errorf("Expected 2 memories deleted, got %d", output.DeletedCount)
	}

	// Verify only bob's memory remains
	filter := domain.ElementFilter{
		Type: func() *domain.ElementType { t := domain.MemoryElement; return &t }(),
	}
	remaining, _ := server.repo.List(filter)

	if len(remaining) != 1 {
		t.Errorf("Expected 1 remaining memory, got %d", len(remaining))
	}

	if len(remaining) > 0 && remaining[0].GetMetadata().Author != "bob" {
		t.Errorf("Expected remaining memory to be by bob, got %s", remaining[0].GetMetadata().Author)
	}
}

func TestHandleClearMemories_DateFilter(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create memories with different dates
	mem1 := createTestMemory("Old Memory", "Content", "alice", "2025-11-15")
	mem2 := createTestMemory("Recent Memory", "Content", "alice", "2025-12-15")
	mem3 := createTestMemory("Latest Memory", "Content", "alice", "2025-12-17")

	server.repo.Create(mem1)
	server.repo.Create(mem2)
	server.repo.Create(mem3)

	// Clear memories before December
	input := ClearMemoriesInput{
		DateBefore: "2025-12-01",
		Confirm:    true,
	}

	_, output, err := server.handleClearMemories(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleClearMemories failed: %v", err)
	}

	if output.DeletedCount != 1 {
		t.Errorf("Expected 1 memory deleted (November), got %d", output.DeletedCount)
	}

	// Verify December memories remain
	filter := domain.ElementFilter{
		Type: func() *domain.ElementType { t := domain.MemoryElement; return &t }(),
	}
	remaining, _ := server.repo.List(filter)

	if len(remaining) != 2 {
		t.Errorf("Expected 2 remaining memories (December), got %d", len(remaining))
	}
}

func TestHandleClearMemories_NoConfirmation(t *testing.T) {
	server := setupMemoryTestServer(t)
	ctx := context.Background()

	// Create a memory
	memory := createTestMemory("Memory", "Content", "alice", "2025-12-15")
	server.repo.Create(memory)

	// Try to clear without confirmation
	input := ClearMemoriesInput{
		Confirm: false,
	}

	_, _, err := server.handleClearMemories(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error when confirm is false")
	}

	if err.Error() != "confirmation required: set confirm=true to proceed" {
		t.Errorf("Expected confirmation error, got: %v", err)
	}

	// Verify memory still exists
	filter := domain.ElementFilter{
		Type: func() *domain.ElementType { t := domain.MemoryElement; return &t }(),
	}
	memories, _ := server.repo.List(filter)

	if len(memories) != 1 {
		t.Errorf("Expected memory to still exist, got %d memories", len(memories))
	}
}
