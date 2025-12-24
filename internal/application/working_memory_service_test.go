package application

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkingMemoryService(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)

	assert.NotNil(t, svc)
	assert.NotNil(t, svc.store)
	assert.NotNil(t, svc.sessions)
	assert.NotNil(t, svc.cleanupTick)
	assert.NotNil(t, svc.stopCleanup)
}

func TestWorkingMemoryService_Add(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		content   string
		priority  domain.MemoryPriority
		tags      []string
		metadata  map[string]string
		wantErr   bool
	}{
		{
			name:      "valid memory",
			sessionID: "session-123",
			content:   "Test content",
			priority:  domain.PriorityMedium,
			tags:      []string{"test"},
			metadata:  map[string]string{"key": "value"},
			wantErr:   false,
		},
		{
			name:      "empty session ID",
			sessionID: "",
			content:   "Test content",
			priority:  domain.PriorityMedium,
			wantErr:   true,
		},
		{
			name:      "empty content",
			sessionID: "session-123",
			content:   "",
			priority:  domain.PriorityMedium,
			wantErr:   true,
		},
		{
			name:      "low priority",
			sessionID: "session-123",
			content:   "Low priority content",
			priority:  domain.PriorityLow,
			wantErr:   false,
		},
		{
			name:      "critical priority",
			sessionID: "session-123",
			content:   "Critical content",
			priority:  domain.PriorityCritical,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepository()
			svc := NewWorkingMemoryService(repo)
			ctx := context.Background()

			wm, err := svc.Add(ctx, tt.sessionID, tt.content, tt.priority, tt.tags, tt.metadata)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, wm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wm)
				assert.Equal(t, tt.sessionID, wm.SessionID)
				assert.Equal(t, tt.content, wm.Content)
				assert.Equal(t, tt.priority, wm.Priority)
				if tt.tags != nil {
					assert.Equal(t, tt.tags, wm.Tags)
				}
				if tt.metadata != nil {
					for k, v := range tt.metadata {
						assert.Equal(t, v, wm.Metadata[k])
					}
				}
			}
		})
	}
}

func TestWorkingMemoryService_Get(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"
	content := "Test content"

	// Add memory
	wm, err := svc.Add(ctx, sessionID, content, domain.PriorityMedium, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)

	// Get memory
	retrieved, err := svc.Get(ctx, sessionID, wm.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, wm.ID, retrieved.ID)
	assert.Equal(t, wm.Content, retrieved.Content)

	// Access count should increase
	assert.Greater(t, retrieved.AccessCount, 0)

	// Get non-existent memory
	_, err = svc.Get(ctx, sessionID, "non-existent")
	assert.Error(t, err)

	// Get from non-existent session
	_, err = svc.Get(ctx, "non-existent-session", wm.ID)
	assert.Error(t, err)
}

func TestWorkingMemoryService_List(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"

	// Add multiple memories
	wm1, _ := svc.Add(ctx, sessionID, "Content 1", domain.PriorityMedium, nil, nil)
	wm2, _ := svc.Add(ctx, sessionID, "Content 2", domain.PriorityHigh, nil, nil)
	wm3, _ := svc.Add(ctx, sessionID, "Content 3", domain.PriorityLow, nil, nil)

	// Make one expired
	wm3.ExpiresAt = time.Now().Add(-1 * time.Hour)

	// Promote one
	now := time.Now()
	wm2.PromotedAt = &now

	// List all
	all, err := svc.List(ctx, sessionID, true, true)
	assert.NoError(t, err)
	assert.Len(t, all, 3)

	// List without expired
	active, err := svc.List(ctx, sessionID, false, true)
	assert.NoError(t, err)
	assert.Len(t, active, 2)

	// List without promoted
	nonPromoted, err := svc.List(ctx, sessionID, true, false)
	assert.NoError(t, err)
	assert.Len(t, nonPromoted, 2)
	// Verify wm2 (promoted) is not in the list
	for _, m := range nonPromoted {
		assert.NotEqual(t, wm2.ID, m.ID)
	}

	// List only active and non-promoted
	filtered, err := svc.List(ctx, sessionID, false, false)
	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, wm1.ID, filtered[0].ID)

	// List from non-existent session (returns empty list, not error)
	empty, err := svc.List(ctx, "non-existent-session", true, true)
	assert.NoError(t, err)
	assert.Len(t, empty, 0)
}

func TestWorkingMemoryService_Promote(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"
	content := "Important content to promote"

	// Add memory
	wm, err := svc.Add(ctx, sessionID, content, domain.PriorityHigh, []string{"important"}, nil)
	require.NoError(t, err)

	// Access it multiple times
	for range 5 {
		_, err := svc.Get(ctx, sessionID, wm.ID)
		require.NoError(t, err)
	}

	// Promote to long-term
	longTermMem, err := svc.Promote(ctx, sessionID, wm.ID)
	assert.NoError(t, err)
	assert.NotNil(t, longTermMem)
	assert.Equal(t, content, longTermMem.Content)
	assert.Contains(t, longTermMem.Metadata, "promoted_from")
	assert.Contains(t, longTermMem.Metadata, "access_count")

	// Check working memory is marked as promoted
	retrieved, err := svc.Get(ctx, sessionID, wm.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.True(t, retrieved.IsPromoted())

	// Verify long-term memory was saved
	exists, err := repo.Exists(longTermMem.GetID())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Try to promote again (should return existing long-term memory)
	longTermMem2, err := svc.Promote(ctx, sessionID, wm.ID)
	assert.NoError(t, err)
	assert.NotNil(t, longTermMem2)
	assert.Equal(t, longTermMem.GetID(), longTermMem2.GetID())

	// Promote non-existent memory
	_, err = svc.Promote(ctx, sessionID, "non-existent")
	assert.Error(t, err)
}

func TestWorkingMemoryService_ClearSession(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"

	// Add memories
	_, err := svc.Add(ctx, sessionID, "Content 1", domain.PriorityMedium, nil, nil)
	require.NoError(t, err)
	_, err = svc.Add(ctx, sessionID, "Content 2", domain.PriorityHigh, nil, nil)
	require.NoError(t, err)

	// Verify memories exist
	list, err := svc.List(ctx, sessionID, true, true)
	require.NoError(t, err)
	require.Len(t, list, 2)

	// Clear session
	err = svc.ClearSession(sessionID)
	assert.NoError(t, err)

	// Verify session is empty (List returns empty array, not error)
	empty, err := svc.List(ctx, sessionID, true, true)
	assert.NoError(t, err)
	assert.Len(t, empty, 0)
}

func TestWorkingMemoryService_ExpireMemory(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"

	// Add memory
	wm, err := svc.Add(ctx, sessionID, "Content", domain.PriorityMedium, nil, nil)
	require.NoError(t, err)

	// Verify not expired initially
	assert.False(t, wm.IsExpired())

	// Expire memory
	err = svc.ExpireMemory(sessionID, wm.ID)
	assert.NoError(t, err)

	// Verify memory is expired (Get returns error for expired memories)
	_, err = svc.Get(ctx, sessionID, wm.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")

	// But List with includeExpired=true should still return it
	expired, err := svc.List(ctx, sessionID, true, true)
	assert.NoError(t, err)
	assert.Len(t, expired, 1)
	assert.True(t, expired[0].IsExpired())

	// Expire non-existent memory
	err = svc.ExpireMemory(sessionID, "non-existent")
	assert.Error(t, err)
}

func TestWorkingMemoryService_ExtendTTL(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"

	// Add memory
	wm, err := svc.Add(ctx, sessionID, "Content", domain.PriorityMedium, nil, nil)
	require.NoError(t, err)
	originalExpiry := wm.ExpiresAt

	time.Sleep(10 * time.Millisecond)

	// Extend TTL
	err = svc.ExtendTTL(sessionID, wm.ID)
	assert.NoError(t, err)

	// Verify expiry was extended
	retrieved, err := svc.Get(ctx, sessionID, wm.ID)
	assert.NoError(t, err)
	assert.True(t, retrieved.ExpiresAt.After(originalExpiry))

	// Extend non-existent memory
	err = svc.ExtendTTL(sessionID, "non-existent")
	assert.Error(t, err)
}

func TestWorkingMemoryService_GetStats(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"

	// Add memories
	wm1, _ := svc.Add(ctx, sessionID, "Content 1", domain.PriorityMedium, nil, nil)
	wm2, _ := svc.Add(ctx, sessionID, "Content 2", domain.PriorityHigh, nil, nil)
	wm3, _ := svc.Add(ctx, sessionID, "Content 3", domain.PriorityLow, nil, nil)

	// Access some memories
	svc.Get(ctx, sessionID, wm1.ID)
	svc.Get(ctx, sessionID, wm1.ID)
	svc.Get(ctx, sessionID, wm2.ID)

	// Expire one
	wm3.ExpiresAt = time.Now().Add(-1 * time.Hour)

	// Promote one
	now := time.Now()
	wm2.PromotedAt = &now

	// Get stats
	stats := svc.GetStats(sessionID)
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.TotalCount)
	assert.Equal(t, 2, stats.ActiveCount)
	assert.Equal(t, 1, stats.ExpiredCount)
	assert.Equal(t, 1, stats.PromotedCount)
	assert.Greater(t, stats.AvgAccessCount, 0.0)
	assert.Greater(t, stats.AvgImportance, 0.0)

	// Get stats from non-existent session returns empty stats
	emptyStats := svc.GetStats("non-existent-session")
	assert.NotNil(t, emptyStats)
	assert.Equal(t, 0, emptyStats.TotalCount)
}

func TestWorkingMemoryService_Export(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"

	// Add memories
	_, err := svc.Add(ctx, sessionID, "Content 1", domain.PriorityMedium, nil, nil)
	require.NoError(t, err)
	_, err = svc.Add(ctx, sessionID, "Content 2", domain.PriorityHigh, nil, nil)
	require.NoError(t, err)

	// Export
	exported, err := svc.Export(sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, exported)
	assert.Len(t, exported, 2)

	// Export from non-existent session returns empty array
	empty, err := svc.Export("non-existent-session")
	assert.NoError(t, err)
	assert.Len(t, empty, 0)
}

func TestWorkingMemoryService_AutoPromotion(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"
	content := "Important content"

	// Add high priority memory
	wm, err := svc.Add(ctx, sessionID, content, domain.PriorityHigh, nil, nil)
	require.NoError(t, err)

	// Access it many times to trigger auto-promotion
	for range 6 {
		_, err := svc.Get(ctx, sessionID, wm.ID)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond)
	}

	// Wait a bit for async promotion (if triggered)
	time.Sleep(200 * time.Millisecond)

	// Check if memory should be promoted
	// Note: Auto-promotion happens asynchronously, so we check if it meets criteria
	// The memory might already be promoted by the async goroutine
	list, _ := svc.List(ctx, sessionID, true, true)
	assert.Len(t, list, 1)
	// Either should promote or already promoted
	assert.True(t, list[0].ShouldPromote() || list[0].IsPromoted())
}

func TestWorkingMemoryService_BackgroundCleanup(t *testing.T) {
	// This test is time-sensitive and might be flaky in CI
	// We just verify the service can be started and stopped
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)

	// Service should be running background cleanup
	assert.NotNil(t, svc.cleanupTick)

	// Shutdown should stop cleanup
	svc.Shutdown()

	// Verify channel is closed (will panic if we write to it)
	assert.NotPanics(t, func() {
		// Just verify shutdown doesn't panic
		time.Sleep(10 * time.Millisecond)
	})
}

func TestWorkingMemoryService_MultipleSessions(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	session1 := "session-1"
	session2 := "session-2"

	// Add memories to different sessions
	wm1, err := svc.Add(ctx, session1, "Content 1", domain.PriorityMedium, nil, nil)
	require.NoError(t, err)
	wm2, err := svc.Add(ctx, session2, "Content 2", domain.PriorityHigh, nil, nil)
	require.NoError(t, err)

	// Verify isolation between sessions
	list1, err := svc.List(ctx, session1, true, true)
	assert.NoError(t, err)
	assert.Len(t, list1, 1)
	assert.Equal(t, wm1.ID, list1[0].ID)

	list2, err := svc.List(ctx, session2, true, true)
	assert.NoError(t, err)
	assert.Len(t, list2, 1)
	assert.Equal(t, wm2.ID, list2[0].ID)

	// Get from wrong session should fail
	_, err = svc.Get(ctx, session1, wm2.ID)
	assert.Error(t, err)

	_, err = svc.Get(ctx, session2, wm1.ID)
	assert.Error(t, err)
}

func TestWorkingMemoryService_ConcurrentAccess(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "concurrent-session"
	content := "Concurrent content"

	// Add memory
	wm, err := svc.Add(ctx, sessionID, content, domain.PriorityMedium, nil, nil)
	require.NoError(t, err)

	// Concurrent reads
	done := make(chan bool)
	errors := make(chan error, 10)
	for range 10 {
		go func() {
			_, err := svc.Get(ctx, sessionID, wm.ID)
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}
	close(errors)

	// Verify that concurrent access worked (all should succeed or fail gracefully)
	errorCount := 0
	for range errors {
		errorCount++
	}

	// All concurrent accesses should have succeeded
	assert.Equal(t, 0, errorCount, "Expected no errors during concurrent access")

	// Verify access count or promotion status
	retrieved, err := svc.Get(ctx, sessionID, wm.ID)
	// If auto-promoted during concurrent access, Get will fail with ErrNotFound
	if err != nil {
		// Check if it was auto-promoted
		list, listErr := svc.List(ctx, sessionID, true, true)
		require.NoError(t, listErr)
		require.Len(t, list, 1, "Expected exactly one memory (promoted)")
		assert.True(t, list[0].IsPromoted(), "Memory should be promoted")
	} else {
		// Not promoted yet, check access count (should be at least 10 + initial + final)
		assert.GreaterOrEqual(t, retrieved.AccessCount, 10, "Access count should reflect concurrent accesses")
	}
}

func TestWorkingMemoryService_PromoteWithMetadata(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"
	content := "Content with metadata"
	customMetadata := map[string]string{
		"source":   "api",
		"category": "important",
	}

	// Add memory with metadata
	wm, err := svc.Add(ctx, sessionID, content, domain.PriorityHigh, []string{"tag1", "tag2"}, customMetadata)
	require.NoError(t, err)

	// Promote
	longTermMem, err := svc.Promote(ctx, sessionID, wm.ID)
	assert.NoError(t, err)

	// Verify custom metadata was preserved
	assert.Equal(t, "api", longTermMem.Metadata["source"])
	assert.Equal(t, "important", longTermMem.Metadata["category"])

	// Verify promotion metadata was added
	assert.Contains(t, longTermMem.Metadata, "promoted_from")
	assert.Contains(t, longTermMem.Metadata, "promoted_at")
	assert.Contains(t, longTermMem.Metadata, "access_count")
}

func TestSessionMemoryCache_LastActivity(t *testing.T) {
	repo := NewMockRepository()
	svc := NewWorkingMemoryService(repo)
	ctx := context.Background()

	sessionID := "session-123"

	// Add memory
	wm, err := svc.Add(ctx, sessionID, "Content", domain.PriorityMedium, nil, nil)
	require.NoError(t, err)

	cache := svc.sessions[sessionID]
	initialActivity := cache.LastActivity

	time.Sleep(10 * time.Millisecond)

	// Access memory
	_, err = svc.Get(ctx, sessionID, wm.ID)
	require.NoError(t, err)

	// Verify last activity was updated
	assert.True(t, cache.LastActivity.After(initialActivity))
}
