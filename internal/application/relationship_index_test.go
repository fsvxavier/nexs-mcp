package application

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// mockRepoForIndex implements domain.ElementRepository for index testing.
type mockRepoForIndex struct {
	elements map[string]domain.Element
}

func newMockRepoForIndex() *mockRepoForIndex {
	return &mockRepoForIndex{
		elements: make(map[string]domain.Element),
	}
}

func (m *mockRepoForIndex) Create(elem domain.Element) error {
	m.elements[elem.GetMetadata().ID] = elem
	return nil
}

func (m *mockRepoForIndex) GetByID(id string) (domain.Element, error) {
	if elem, ok := m.elements[id]; ok {
		return elem, nil
	}
	return nil, domain.ErrElementNotFound
}

func (m *mockRepoForIndex) Update(elem domain.Element) error {
	m.elements[elem.GetMetadata().ID] = elem
	return nil
}

func (m *mockRepoForIndex) Delete(id string) error {
	delete(m.elements, id)
	return nil
}

func (m *mockRepoForIndex) List(filter domain.ElementFilter) ([]domain.Element, error) {
	result := make([]domain.Element, 0, len(m.elements))
	for _, elem := range m.elements {
		// Filter by type if specified
		if filter.Type != nil && elem.GetMetadata().Type != *filter.Type {
			continue
		}
		result = append(result, elem)
	}
	return result, nil
}

func (m *mockRepoForIndex) Exists(id string) (bool, error) {
	_, exists := m.elements[id]
	return exists, nil
}

func TestRelationshipIndex_Add(t *testing.T) {
	idx := NewRelationshipIndex()

	// Add relationship
	idx.Add("memory-001", []string{"persona-001", "skill-001"})

	// Verify forward lookup
	related := idx.GetRelatedElements("memory-001")
	if len(related) != 2 {
		t.Fatalf("Expected 2 related elements, got %d", len(related))
	}

	// Verify reverse lookup
	memories := idx.GetRelatedMemories("persona-001")
	if len(memories) != 1 || memories[0] != "memory-001" {
		t.Errorf("Expected memory-001 in reverse lookup for persona-001")
	}

	memories = idx.GetRelatedMemories("skill-001")
	if len(memories) != 1 || memories[0] != "memory-001" {
		t.Errorf("Expected memory-001 in reverse lookup for skill-001")
	}
}

func TestRelationshipIndex_AddMultipleMemories(t *testing.T) {
	idx := NewRelationshipIndex()

	// Multiple memories referencing same element
	idx.Add("memory-001", []string{"persona-001"})
	idx.Add("memory-002", []string{"persona-001"})
	idx.Add("memory-003", []string{"persona-001"})

	// Verify reverse lookup returns all memories
	memories := idx.GetRelatedMemories("persona-001")
	if len(memories) != 3 {
		t.Fatalf("Expected 3 memories, got %d", len(memories))
	}

	expected := map[string]bool{
		"memory-001": false,
		"memory-002": false,
		"memory-003": false,
	}

	for _, memID := range memories {
		if _, ok := expected[memID]; !ok {
			t.Errorf("Unexpected memory ID: %s", memID)
		}
		expected[memID] = true
	}

	for memID, found := range expected {
		if !found {
			t.Errorf("Missing memory ID: %s", memID)
		}
	}
}

func TestRelationshipIndex_Remove(t *testing.T) {
	idx := NewRelationshipIndex()

	// Add relationships
	idx.Add("memory-001", []string{"persona-001", "skill-001"})
	idx.Add("memory-002", []string{"persona-001"})

	// Remove memory-001
	idx.Remove("memory-001")

	// Verify forward lookup is gone
	related := idx.GetRelatedElements("memory-001")
	if related != nil {
		t.Errorf("Expected nil for removed memory, got %v", related)
	}

	// Verify reverse lookup updated
	memories := idx.GetRelatedMemories("persona-001")
	if len(memories) != 1 || memories[0] != "memory-002" {
		t.Errorf("Expected only memory-002 for persona-001, got %v", memories)
	}

	memories = idx.GetRelatedMemories("skill-001")
	if memories != nil {
		t.Errorf("Expected nil for skill-001, got %v", memories)
	}
}

func TestRelationshipIndex_GetRelatedElements(t *testing.T) {
	idx := NewRelationshipIndex()

	// Test empty index
	related := idx.GetRelatedElements("nonexistent")
	if related != nil {
		t.Errorf("Expected nil for nonexistent memory, got %v", related)
	}

	// Add and test
	idx.Add("memory-001", []string{"elem-001", "elem-002", "elem-003"})
	related = idx.GetRelatedElements("memory-001")

	if len(related) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(related))
	}

	// Verify immutability (returned slice is a copy)
	related[0] = "modified"
	original := idx.GetRelatedElements("memory-001")
	if original[0] == "modified" {
		t.Error("Index was modified through returned slice")
	}
}

func TestRelationshipIndex_GetRelatedMemories(t *testing.T) {
	idx := NewRelationshipIndex()

	// Test empty index
	memories := idx.GetRelatedMemories("nonexistent")
	if memories != nil {
		t.Errorf("Expected nil for nonexistent element, got %v", memories)
	}

	// Add relationships
	idx.Add("memory-001", []string{"elem-001"})
	idx.Add("memory-002", []string{"elem-001"})
	idx.Add("memory-003", []string{"elem-001"})

	memories = idx.GetRelatedMemories("elem-001")
	if len(memories) != 3 {
		t.Fatalf("Expected 3 memories, got %d", len(memories))
	}

	// Verify immutability
	memories[0] = "modified"
	original := idx.GetRelatedMemories("elem-001")
	if original[0] == "modified" {
		t.Error("Index was modified through returned slice")
	}
}

func TestRelationshipIndex_Rebuild(t *testing.T) {
	repo := newMockRepoForIndex()
	idx := NewRelationshipIndex()

	// Create test memories
	mem1 := domain.NewMemory("Memory 1", "Test", "1.0.0", "test")
	mem1.Metadata["related_to"] = "persona-001,skill-001"
	repo.Create(mem1)

	mem2 := domain.NewMemory("Memory 2", "Test", "1.0.0", "test")
	mem2.Metadata["related_to"] = "persona-001"
	repo.Create(mem2)

	mem3 := domain.NewMemory("Memory 3", "Test", "1.0.0", "test")
	mem3.Metadata["related_to"] = "agent-001"
	repo.Create(mem3)

	// Memory with no relationships
	mem4 := domain.NewMemory("Memory 4", "Test", "1.0.0", "test")
	repo.Create(mem4)

	// Rebuild index
	ctx := context.Background()
	err := idx.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Verify forward lookups
	mem1ID := mem1.GetMetadata().ID
	related := idx.GetRelatedElements(mem1ID)
	if len(related) != 2 {
		t.Errorf("Expected 2 elements for memory-1, got %d", len(related))
	}

	// Verify reverse lookups
	memories := idx.GetRelatedMemories("persona-001")
	if len(memories) != 2 {
		t.Errorf("Expected 2 memories for persona-001, got %d", len(memories))
	}

	memories = idx.GetRelatedMemories("skill-001")
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory for skill-001, got %d", len(memories))
	}

	memories = idx.GetRelatedMemories("agent-001")
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory for agent-001, got %d", len(memories))
	}
}

func TestRelationshipIndex_RebuildClearsExisting(t *testing.T) {
	repo := newMockRepoForIndex()
	idx := NewRelationshipIndex()

	// Add some initial data
	idx.Add("old-memory", []string{"old-elem"})

	// Create new memory in repo
	mem := domain.NewMemory("New Memory", "Test", "1.0.0", "test")
	mem.Metadata["related_to"] = "new-elem"
	repo.Create(mem)

	// Rebuild
	ctx := context.Background()
	err := idx.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Old data should be gone
	related := idx.GetRelatedElements("old-memory")
	if related != nil {
		t.Error("Old memory should not exist after rebuild")
	}

	memories := idx.GetRelatedMemories("old-elem")
	if memories != nil {
		t.Error("Old element should not exist after rebuild")
	}

	// New data should exist
	memories = idx.GetRelatedMemories("new-elem")
	if len(memories) != 1 {
		t.Error("New element should have 1 memory")
	}
}

func TestRelationshipIndex_Stats(t *testing.T) {
	idx := NewRelationshipIndex()

	// Initial stats
	stats := idx.Stats()
	if stats.ForwardEntries != 0 || stats.ReverseEntries != 0 {
		t.Error("Expected empty index initially")
	}

	// Add relationships
	idx.Add("memory-001", []string{"elem-001", "elem-002"})
	idx.Add("memory-002", []string{"elem-001"})

	stats = idx.Stats()
	if stats.ForwardEntries != 2 {
		t.Errorf("Expected 2 forward entries, got %d", stats.ForwardEntries)
	}
	if stats.ReverseEntries != 2 {
		t.Errorf("Expected 2 reverse entries (elem-001, elem-002), got %d", stats.ReverseEntries)
	}
}

func TestIndexCache_GetSet(t *testing.T) {
	cache := NewIndexCache(1 * time.Second)

	// Test miss
	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected cache miss for non-existent key")
	}

	// Set and get
	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected cache hit after Set")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
}

func TestIndexCache_Expiration(t *testing.T) {
	cache := NewIndexCache(100 * time.Millisecond)

	cache.Set("key1", "value1")

	// Should be available immediately
	_, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected cache hit immediately after Set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Expected cache miss after expiration")
	}
}

func TestIndexCache_Invalidate(t *testing.T) {
	cache := NewIndexCache(1 * time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Invalidate key1
	cache.Invalidate("key1")

	// key1 should be gone
	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected cache miss after invalidate")
	}

	// key2 should still exist
	_, ok = cache.Get("key2")
	if !ok {
		t.Error("Expected key2 to still exist")
	}
}

func TestIndexCache_InvalidatePattern(t *testing.T) {
	cache := NewIndexCache(1 * time.Minute)

	cache.Set("memory-001", "value1")
	cache.Set("memory-002", "value2")
	cache.Set("persona-001", "value3")

	// Invalidate all memory keys
	cache.InvalidatePattern("memory")

	// Memory keys should be gone
	_, ok := cache.Get("memory-001")
	if ok {
		t.Error("Expected memory-001 to be invalidated")
	}
	_, ok = cache.Get("memory-002")
	if ok {
		t.Error("Expected memory-002 to be invalidated")
	}

	// Persona key should still exist
	_, ok = cache.Get("persona-001")
	if !ok {
		t.Error("Expected persona-001 to still exist")
	}
}

func TestIndexCache_Clear(t *testing.T) {
	cache := NewIndexCache(1 * time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Clear all
	cache.Clear()

	// All keys should be gone
	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be cleared")
	}
	_, ok = cache.Get("key2")
	if ok {
		t.Error("Expected key2 to be cleared")
	}
	_, ok = cache.Get("key3")
	if ok {
		t.Error("Expected key3 to be cleared")
	}
}

func TestIndexCache_Stats(t *testing.T) {
	cache := NewIndexCache(1 * time.Minute)

	// Initial stats
	if cache.hitCount != 0 || cache.missCount != 0 {
		t.Error("Expected zero stats initially")
	}

	// Miss
	cache.Get("nonexistent")
	if cache.missCount != 1 {
		t.Errorf("Expected 1 miss, got %d", cache.missCount)
	}

	// Hit
	cache.Set("key1", "value1")
	cache.Get("key1")
	if cache.hitCount != 1 {
		t.Errorf("Expected 1 hit, got %d", cache.hitCount)
	}

	// Multiple operations
	cache.Get("nonexistent2")
	cache.Get("key1")
	cache.Get("nonexistent3")

	if cache.hitCount != 2 {
		t.Errorf("Expected 2 hits, got %d", cache.hitCount)
	}
	if cache.missCount != 3 {
		t.Errorf("Expected 3 misses, got %d", cache.missCount)
	}
}

func TestGetMemoriesRelatedTo(t *testing.T) {
	repo := newMockRepoForIndex()
	idx := NewRelationshipIndex()

	// Create memories
	mem1 := domain.NewMemory("Memory 1", "Test", "1.0.0", "test")
	mem1.Metadata["related_to"] = "persona-001"
	repo.Create(mem1)

	mem2 := domain.NewMemory("Memory 2", "Test", "1.0.0", "test")
	mem2.Metadata["related_to"] = "persona-001,skill-001"
	repo.Create(mem2)

	mem3 := domain.NewMemory("Memory 3", "Test", "1.0.0", "test")
	mem3.Metadata["related_to"] = "skill-001"
	repo.Create(mem3)

	// Rebuild index
	ctx := context.Background()
	err := idx.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// Get memories related to persona-001
	memories, err := GetMemoriesRelatedTo(ctx, "persona-001", repo, idx)
	if err != nil {
		t.Fatalf("GetMemoriesRelatedTo failed: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("Expected 2 memories for persona-001, got %d", len(memories))
	}

	// Get memories related to skill-001
	memories, err = GetMemoriesRelatedTo(ctx, "skill-001", repo, idx)
	if err != nil {
		t.Fatalf("GetMemoriesRelatedTo failed: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("Expected 2 memories for skill-001, got %d", len(memories))
	}

	// Get memories for non-existent element
	memories, err = GetMemoriesRelatedTo(ctx, "nonexistent", repo, idx)
	if err != nil {
		t.Fatalf("GetMemoriesRelatedTo failed: %v", err)
	}

	if memories != nil {
		t.Errorf("Expected nil for nonexistent element, got %v", memories)
	}
}

func TestGetMemoriesRelatedTo_WithCache(t *testing.T) {
	repo := newMockRepoForIndex()
	idx := NewRelationshipIndex()

	// Create memory
	mem := domain.NewMemory("Memory 1", "Test", "1.0.0", "test")
	mem.Metadata["related_to"] = "persona-001"
	repo.Create(mem)

	// Rebuild index
	ctx := context.Background()
	err := idx.Rebuild(ctx, repo)
	if err != nil {
		t.Fatalf("Rebuild failed: %v", err)
	}

	// First call - should cache
	memories1, err := GetMemoriesRelatedTo(ctx, "persona-001", repo, idx)
	if err != nil {
		t.Fatalf("GetMemoriesRelatedTo failed: %v", err)
	}

	// Second call - should use cache
	memories2, err := GetMemoriesRelatedTo(ctx, "persona-001", repo, idx)
	if err != nil {
		t.Fatalf("GetMemoriesRelatedTo failed: %v", err)
	}

	// Results should be equal
	if len(memories1) != len(memories2) {
		t.Error("Cached result differs from original")
	}

	// Check cache stats
	stats := idx.Stats()
	if stats.CacheHits == 0 {
		t.Error("Expected cache hit on second call")
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("parseRelatedIDsFromString", func(t *testing.T) {
		tests := []struct {
			input string
			want  int
		}{
			{"elem-001,elem-002,elem-003", 3},
			{"elem-001, elem-002 , elem-003", 3},
			{"", 0},
			{"elem-001", 1},
			{"elem-001,,elem-002", 2},
			{"  elem-001  ,  elem-002  ", 2},
		}

		for _, tt := range tests {
			result := parseRelatedIDsFromString(tt.input)
			if len(result) != tt.want {
				t.Errorf("parseRelatedIDsFromString(%q) = %d elements, want %d", tt.input, len(result), tt.want)
			}
		}
	})

	t.Run("contains", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		if !contains(slice, "b") {
			t.Error("Expected contains to return true for 'b'")
		}
		if contains(slice, "d") {
			t.Error("Expected contains to return false for 'd'")
		}
	})

	t.Run("removeString", func(t *testing.T) {
		slice := []string{"a", "b", "c", "b"}
		result := removeString(slice, "b")
		if len(result) != 2 {
			t.Errorf("Expected 2 elements after removing 'b', got %d", len(result))
		}
		if contains(result, "b") {
			t.Error("Result should not contain 'b'")
		}
	})

	t.Run("copyStrings", func(t *testing.T) {
		original := []string{"a", "b", "c"}
		copy := copyStrings(original)
		if len(copy) != len(original) {
			t.Error("Copy length differs from original")
		}
		copy[0] = "modified"
		if original[0] == "modified" {
			t.Error("Original was modified through copy")
		}
	})
}
