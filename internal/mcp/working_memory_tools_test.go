package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Working memory tools are registered via RegisterWorkingMemoryTools with a WorkingMemoryService
// These tests verify the basic structures and would require a full integration test setup

func TestWorkingMemoryStructures(t *testing.T) {
	t.Run("working_memory_input_structures", func(t *testing.T) {
		// Test that input structs are properly defined
		addInput := WorkingMemoryAddInput{
			SessionID: "test-session",
			Content:   "test content",
			Priority:  "medium",
			Tags:      []string{"test"},
			Metadata:  map[string]string{"key": "value"},
			Context:   "test context",
		}
		assert.Equal(t, "test-session", addInput.SessionID)
		assert.Equal(t, "test content", addInput.Content)
		assert.Equal(t, "medium", addInput.Priority)
		assert.NotEmpty(t, addInput.Tags)
		assert.NotEmpty(t, addInput.Metadata)

		getInput := WorkingMemoryGetInput{
			SessionID: "test-session",
			MemoryID:  "test-memory",
		}
		assert.Equal(t, "test-session", getInput.SessionID)
		assert.Equal(t, "test-memory", getInput.MemoryID)

		listInput := WorkingMemoryListInput{
			SessionID:       "test-session",
			IncludeExpired:  true,
			IncludePromoted: false,
		}
		assert.Equal(t, "test-session", listInput.SessionID)
		assert.True(t, listInput.IncludeExpired)
		assert.False(t, listInput.IncludePromoted)
	})

	t.Run("working_memory_priority_values", func(t *testing.T) {
		// Test valid priority values
		validPriorities := []string{"low", "medium", "high", "critical"}
		for _, priority := range validPriorities {
			input := WorkingMemoryAddInput{
				SessionID: "test",
				Content:   "test",
				Priority:  priority,
			}
			assert.Equal(t, priority, input.Priority)
		}
	})

	t.Run("working_memory_session_operations", func(t *testing.T) {
		// Test session-related input structures
		clearInput := WorkingMemoryClearSessionInput{
			SessionID: "test-session",
		}
		assert.Equal(t, "test-session", clearInput.SessionID)

		statsInput := WorkingMemoryStatsInput{
			SessionID: "test-session",
		}
		assert.Equal(t, "test-session", statsInput.SessionID)

		exportInput := WorkingMemoryExportInput{
			SessionID: "test-session",
		}
		assert.Equal(t, "test-session", exportInput.SessionID)
	})

	t.Run("working_memory_lifecycle_operations", func(t *testing.T) {
		// Test lifecycle operation structures
		promoteInput := WorkingMemoryPromoteInput{
			SessionID: "test-session",
			MemoryID:  "test-memory",
		}
		assert.Equal(t, "test-session", promoteInput.SessionID)
		assert.Equal(t, "test-memory", promoteInput.MemoryID)

		expireInput := WorkingMemoryExpireInput{
			SessionID: "test-session",
			MemoryID:  "test-memory",
		}
		assert.Equal(t, "test-session", expireInput.SessionID)
		assert.Equal(t, "test-memory", expireInput.MemoryID)

		extendInput := WorkingMemoryExtendTTLInput{
			SessionID: "test-session",
			MemoryID:  "test-memory",
		}
		assert.Equal(t, "test-session", extendInput.SessionID)
		assert.Equal(t, "test-memory", extendInput.MemoryID)
	})

	t.Run("working_memory_bulk_operations", func(t *testing.T) {
		// Test bulk operation structures
		bulkPromoteInput := WorkingMemoryBulkPromoteInput{
			SessionID: "test-session",
		}
		assert.Equal(t, "test-session", bulkPromoteInput.SessionID)
	})

	t.Run("working_memory_relation_operations", func(t *testing.T) {
		// Test relation operation structures
		relationInput := WorkingMemoryRelationAddInput{
			SessionID:       "test-session",
			MemoryID:        "memory-1",
			RelatedMemoryID: "memory-2",
		}
		assert.Equal(t, "test-session", relationInput.SessionID)
		assert.Equal(t, "memory-1", relationInput.MemoryID)
		assert.Equal(t, "memory-2", relationInput.RelatedMemoryID)
	})

	t.Run("working_memory_search_operations", func(t *testing.T) {
		// Test search operation structures
		searchInput := WorkingMemorySearchInput{
			SessionID: "test-session",
			Query:     "test query",
		}
		assert.Equal(t, "test-session", searchInput.SessionID)
		assert.Equal(t, "test query", searchInput.Query)
	})
}

func TestRegisterWorkingMemoryTools(t *testing.T) {
	t.Run("register_with_nil_service", func(t *testing.T) {
		server := setupTestServer()

		// RegisterWorkingMemoryTools should handle nil service gracefully
		RegisterWorkingMemoryTools(server, nil)

		// Server should still be functional
		assert.NotNil(t, server)
		assert.NotNil(t, server.repo)
	})
}
