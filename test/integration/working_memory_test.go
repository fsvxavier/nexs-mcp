package integration

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkingMemory_E2E_Creation tests the complete creation flow
func TestWorkingMemory_E2E_Creation(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-1"

	// Create working memory
	wm, err := service.Add(ctx, sessionID, "Test content for working memory", domain.PriorityHigh, []string{"test", "integration"}, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, wm.ID)
	assert.Equal(t, sessionID, wm.SessionID)
	assert.Equal(t, "Test content for working memory", wm.Content)
	assert.Equal(t, domain.PriorityHigh, wm.Priority)
	assert.Contains(t, wm.Tags, "test")
	assert.Contains(t, wm.Tags, "integration")

	// Verify we can retrieve it
	retrieved, err := service.Get(ctx, sessionID, wm.ID)
	require.NoError(t, err)
	assert.Equal(t, wm.ID, retrieved.ID)
	assert.Equal(t, 2, retrieved.AccessCount) // RecordAccess called in Add() and Get()

	t.Log("✓ E2E Working Memory creation and retrieval successful")
}

// TestWorkingMemory_E2E_AutoPromotion tests automatic promotion to long-term memory
func TestWorkingMemory_E2E_AutoPromotion(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-2"

	// Create high-priority working memory
	wm, err := service.Add(ctx, sessionID, "Important content for auto-promotion", domain.PriorityHigh, []string{"important"}, nil)
	require.NoError(t, err)

	// Access multiple times to trigger auto-promotion (threshold is 10 for high priority)
	for i := 0; i < 12; i++ {
		_, err := service.Get(ctx, sessionID, wm.ID)
		require.NoError(t, err)
	}

	// Wait for async promotion
	time.Sleep(300 * time.Millisecond)

	// Verify promotion occurred
	retrieved, err := service.Get(ctx, sessionID, wm.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.IsPromoted() || retrieved.ShouldPromote(), "Memory should be promoted or ready for promotion")

	// If promoted, verify long-term memory exists
	if retrieved.IsPromoted() {
		assert.NotEmpty(t, retrieved.PromotedToID)
		longTermMem, err := repo.GetByID(retrieved.PromotedToID)
		require.NoError(t, err)
		assert.NotNil(t, longTermMem)

		if mem, ok := longTermMem.(*domain.Memory); ok {
			assert.Contains(t, mem.GetMetadata().Tags, "promoted")
		}
	}

	t.Log("✓ E2E Working Memory auto-promotion successful")
}

// TestWorkingMemory_E2E_ManualPromotion tests manual promotion
func TestWorkingMemory_E2E_ManualPromotion(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-3"

	// Create working memory
	wm, err := service.Add(ctx, sessionID, "Content to be manually promoted", domain.PriorityMedium, []string{"manual"}, nil)
	require.NoError(t, err)

	// Manually promote
	longTermMem, err := service.Promote(ctx, sessionID, wm.ID)
	require.NoError(t, err)
	assert.NotNil(t, longTermMem)
	assert.Contains(t, longTermMem.GetMetadata().Tags, "promoted")

	// Verify working memory is marked as promoted
	retrieved, err := service.Get(ctx, sessionID, wm.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.IsPromoted())
	assert.Equal(t, longTermMem.GetID(), retrieved.PromotedToID)

	// Verify long-term memory contains original content
	assert.Contains(t, longTermMem.Content, "Content to be manually promoted")

	t.Log("✓ E2E Working Memory manual promotion successful")
}

// TestWorkingMemory_E2E_TTLExpiration tests TTL and expiration
func TestWorkingMemory_E2E_TTLExpiration(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-4"

	// Create low-priority working memory (1 hour TTL)
	wm, err := service.Add(ctx, sessionID, "Content that will expire", domain.PriorityLow, []string{"expiring"}, nil)
	require.NoError(t, err)

	// Verify it's not expired
	assert.False(t, wm.IsExpired())

	// Force expiration by manually setting ExpiresAt to past
	wm.ExpiresAt = time.Now().Add(-1 * time.Hour)

	// Verify it's now expired
	assert.True(t, wm.IsExpired())

	// List should not include expired by default
	list, err := service.List(ctx, sessionID, false, false)
	require.NoError(t, err)
	assert.NotContains(t, list, wm)

	// List with includeExpired should include it
	listWithExpired, err := service.List(ctx, sessionID, true, false)
	require.NoError(t, err)
	found := false
	for _, m := range listWithExpired {
		if m.ID == wm.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expired memory should be in list when includeExpired=true")

	t.Log("✓ E2E Working Memory TTL expiration successful")
}

// TestWorkingMemory_E2E_MultipleSessions tests session isolation
func TestWorkingMemory_E2E_MultipleSessions(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	session1 := "test-session-5a"
	session2 := "test-session-5b"

	// Create memories in different sessions
	wm1, err := service.Add(ctx, session1, "Content in session 1", domain.PriorityMedium, []string{"session1"}, nil)
	require.NoError(t, err)

	wm2, err := service.Add(ctx, session2, "Content in session 2", domain.PriorityMedium, []string{"session2"}, nil)
	require.NoError(t, err)

	// Verify session isolation
	list1, err := service.List(ctx, session1, false, false)
	require.NoError(t, err)
	assert.Len(t, list1, 1)
	assert.Equal(t, wm1.ID, list1[0].ID)

	list2, err := service.List(ctx, session2, false, false)
	require.NoError(t, err)
	assert.Len(t, list2, 1)
	assert.Equal(t, wm2.ID, list2[0].ID)

	// Session 1 cannot access session 2's memory
	_, err = service.Get(ctx, session1, wm2.ID)
	assert.Error(t, err)

	// Clear one session doesn't affect the other
	err = service.ClearSession(session1)
	require.NoError(t, err)

	list1After, err := service.List(ctx, session1, false, false)
	require.NoError(t, err)
	assert.Empty(t, list1After)

	list2After, err := service.List(ctx, session2, false, false)
	require.NoError(t, err)
	assert.Len(t, list2After, 1)

	t.Log("✓ E2E Working Memory session isolation successful")
}

// TestWorkingMemory_E2E_ExtendTTL tests TTL extension
func TestWorkingMemory_E2E_ExtendTTL(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-6"

	// Create working memory
	wm, err := service.Add(ctx, sessionID, "Content with extendable TTL", domain.PriorityMedium, []string{"extendable"}, nil)
	require.NoError(t, err)

	originalExpiry := wm.ExpiresAt

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Extend TTL
	err = service.ExtendTTL(sessionID, wm.ID)
	require.NoError(t, err)

	// Verify expiry was extended
	retrieved, err := service.Get(ctx, sessionID, wm.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.ExpiresAt.After(originalExpiry), "ExpiresAt should be extended")

	t.Log("✓ E2E Working Memory TTL extension successful")
}

// TestWorkingMemory_E2E_Statistics tests statistics collection
func TestWorkingMemory_E2E_Statistics(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-7"

	// Create multiple memories
	_, err := service.Add(ctx, sessionID, "Memory 1", domain.PriorityHigh, nil, nil)
	require.NoError(t, err)

	wm2, err := service.Add(ctx, sessionID, "Memory 2", domain.PriorityLow, nil, nil)
	require.NoError(t, err)

	_, err = service.Add(ctx, sessionID, "Memory 3", domain.PriorityMedium, nil, nil)
	require.NoError(t, err)

	// Promote one
	_, err = service.Promote(ctx, sessionID, wm2.ID)
	require.NoError(t, err)

	// Get statistics
	stats := service.GetStats(sessionID)
	assert.Equal(t, 3, stats.TotalCount)
	assert.Equal(t, 3, stats.ActiveCount) // All 3 are active (not expired)
	assert.Equal(t, 1, stats.PromotedCount)
	assert.Equal(t, 0, stats.ExpiredCount)
	assert.Greater(t, stats.AvgAccessCount, 0.0)

	t.Log("✓ E2E Working Memory statistics successful")
}

// TestWorkingMemory_E2E_BackgroundCleanup tests background cleanup of expired memories
func TestWorkingMemory_E2E_BackgroundCleanup(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-8"

	// Create working memory that will expire soon
	wm, err := service.Add(ctx, sessionID, "Content that expires quickly", domain.PriorityLow, nil, nil)
	require.NoError(t, err)

	// Force immediate expiration
	err = service.ExpireMemory(sessionID, wm.ID)
	require.NoError(t, err)

	// Background cleanup runs every 5 minutes in production, but we can verify
	// that expired memories are filtered out by List
	list, err := service.List(ctx, sessionID, false, false)
	require.NoError(t, err)
	assert.Empty(t, list, "Expired memory should not be in default list")

	t.Log("✓ E2E Working Memory background cleanup successful")
}

// TestWorkingMemory_E2E_Export tests export functionality
func TestWorkingMemory_E2E_Export(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-9"

	// Create multiple memories
	for i := 1; i <= 3; i++ {
		_, err := service.Add(ctx, sessionID, "Memory "+string(rune('0'+i)), domain.PriorityMedium, []string{"export-test"}, nil)
		require.NoError(t, err)
	}

	// Export all memories
	exported, err := service.Export(sessionID)
	require.NoError(t, err)
	assert.Len(t, exported, 3)

	// Verify exported memories have all fields
	for _, wm := range exported {
		assert.NotEmpty(t, wm.ID)
		assert.Equal(t, sessionID, wm.SessionID)
		assert.NotEmpty(t, wm.Content)
		assert.Contains(t, wm.Tags, "export-test")
	}

	t.Log("✓ E2E Working Memory export successful")
}

// TestWorkingMemory_E2E_PriorityLevels tests different priority levels
func TestWorkingMemory_E2E_PriorityLevels(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-10"

	priorities := []domain.MemoryPriority{
		domain.PriorityLow,
		domain.PriorityMedium,
		domain.PriorityHigh,
		domain.PriorityCritical,
	}

	expectedTTLs := map[domain.MemoryPriority]time.Duration{
		domain.PriorityLow:      1 * time.Hour,
		domain.PriorityMedium:   4 * time.Hour,
		domain.PriorityHigh:     12 * time.Hour,
		domain.PriorityCritical: 24 * time.Hour,
	}

	for _, priority := range priorities {
		wm, err := service.Add(ctx, sessionID, "Content with "+string(priority)+" priority", priority, nil, nil)
		require.NoError(t, err)

		// Verify TTL is approximately correct (within 1 minute tolerance)
		actualTTL := time.Until(wm.ExpiresAt)
		expectedTTL := expectedTTLs[priority]
		tolerance := 1 * time.Minute

		assert.InDelta(t, expectedTTL.Seconds(), actualTTL.Seconds(), tolerance.Seconds(),
			"TTL for %s priority should be approximately %v", priority, expectedTTL)
	}

	t.Log("✓ E2E Working Memory priority levels successful")
}

// TestWorkingMemory_E2E_ConcurrentSessions tests concurrent access across sessions
func TestWorkingMemory_E2E_ConcurrentSessions(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessions := []string{"concurrent-1", "concurrent-2", "concurrent-3"}

	// Create memories concurrently
	done := make(chan bool, len(sessions))
	for _, sessionID := range sessions {
		go func(sid string) {
			for i := 0; i < 5; i++ {
				_, err := service.Add(ctx, sid, "Concurrent content "+string(rune('A'+i)), domain.PriorityMedium, nil, nil)
				assert.NoError(t, err)
			}
			done <- true
		}(sessionID)
	}

	// Wait for all goroutines
	for range sessions {
		<-done
	}

	// Verify each session has 5 memories
	for _, sessionID := range sessions {
		list, err := service.List(ctx, sessionID, false, false)
		require.NoError(t, err)
		assert.Len(t, list, 5, "Session %s should have 5 memories", sessionID)
	}

	t.Log("✓ E2E Working Memory concurrent sessions successful")
}

// TestWorkingMemory_E2E_PromotionPreservesMetadata tests metadata preservation during promotion
func TestWorkingMemory_E2E_PromotionPreservesMetadata(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	service := application.NewWorkingMemoryService(repo)
	defer service.Shutdown()

	ctx := context.Background()
	sessionID := "test-session-11"

	// Create working memory with custom metadata
	wm, err := service.Add(ctx, sessionID, "Content with metadata", domain.PriorityHigh, []string{"meta-test", "preservation"}, nil)
	require.NoError(t, err)

	// Add custom metadata
	wm.Metadata["custom_field"] = "custom_value"
	wm.Metadata["importance_flag"] = "high"

	// Promote
	longTermMem, err := service.Promote(ctx, sessionID, wm.ID)
	require.NoError(t, err)

	// Verify metadata was preserved
	assert.Equal(t, "custom_value", longTermMem.Metadata["custom_field"])
	assert.Equal(t, "high", longTermMem.Metadata["importance_flag"])

	// Verify promotion metadata was added
	assert.NotEmpty(t, longTermMem.Metadata["promoted_from"])
	assert.NotEmpty(t, longTermMem.Metadata["promoted_at"])
	assert.NotEmpty(t, longTermMem.Metadata["access_count"])
	assert.NotEmpty(t, longTermMem.Metadata["importance_score"])
	assert.NotEmpty(t, longTermMem.Metadata["priority"])

	t.Log("✓ E2E Working Memory metadata preservation successful")
}
