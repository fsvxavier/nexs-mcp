package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkingMemory(t *testing.T) {
	sessionID := "test-session-123"
	content := "Test content for working memory"
	priority := PriorityMedium

	wm := NewWorkingMemory(sessionID, content, priority)

	assert.NotNil(t, wm)
	assert.NotEmpty(t, wm.ID)
	assert.Equal(t, sessionID, wm.SessionID)
	assert.Equal(t, content, wm.Content)
	assert.Equal(t, priority, wm.Priority)
	assert.NotZero(t, wm.CreatedAt)
	assert.NotZero(t, wm.LastAccessedAt)
	assert.NotZero(t, wm.ExpiresAt)
	assert.NotNil(t, wm.Tags)
	assert.NotNil(t, wm.Metadata)
	assert.Zero(t, wm.AccessCount)
	assert.Zero(t, wm.ImportanceScore)
	assert.Nil(t, wm.PromotedAt)
	assert.Empty(t, wm.PromotedToID)
}

func TestMemoryPriority_TTL(t *testing.T) {
	tests := []struct {
		name     string
		priority MemoryPriority
		expected time.Duration
	}{
		{"low priority", PriorityLow, 1 * time.Hour},
		{"medium priority", PriorityMedium, 4 * time.Hour},
		{"high priority", PriorityHigh, 12 * time.Hour},
		{"critical priority", PriorityCritical, 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttl := tt.priority.TTL()
			assert.Equal(t, tt.expected, ttl)
		})
	}
}

func TestMemoryPriority_PromotionThreshold(t *testing.T) {
	tests := []struct {
		name      string
		priority  MemoryPriority
		threshold int
	}{
		{"low priority", PriorityLow, 10},
		{"medium priority", PriorityMedium, 7},
		{"high priority", PriorityHigh, 5},
		{"critical priority", PriorityCritical, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threshold := tt.priority.PromotionThreshold()
			assert.Equal(t, tt.threshold, threshold)
		})
	}
}

func TestWorkingMemory_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *WorkingMemory
		expected bool
	}{
		{
			name: "not expired",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityMedium)
				return wm
			},
			expected: false,
		},
		{
			name: "expired",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityMedium)
				wm.ExpiresAt = time.Now().Add(-1 * time.Hour)
				return wm
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := tt.setup()
			assert.Equal(t, tt.expected, wm.IsExpired())
		})
	}
}

func TestWorkingMemory_IsPromoted(t *testing.T) {
	wm := NewWorkingMemory("session", "content", PriorityMedium)
	assert.False(t, wm.IsPromoted())

	now := time.Now()
	wm.PromotedAt = &now
	assert.True(t, wm.IsPromoted())
}

func TestWorkingMemory_MarkPromoted(t *testing.T) {
	wm := NewWorkingMemory("session", "content", PriorityMedium)
	longTermID := "mem-long-term-123"

	wm.MarkPromoted(longTermID)

	assert.NotNil(t, wm.PromotedAt)
	assert.Equal(t, longTermID, wm.PromotedToID)
}

func TestWorkingMemory_RecordAccess(t *testing.T) {
	wm := NewWorkingMemory("session", "content", PriorityMedium)
	initialAccessCount := wm.AccessCount
	initialLastAccessed := wm.LastAccessedAt

	time.Sleep(10 * time.Millisecond)
	wm.RecordAccess()

	assert.Equal(t, initialAccessCount+1, wm.AccessCount)
	assert.True(t, wm.LastAccessedAt.After(initialLastAccessed))
	assert.Greater(t, wm.ImportanceScore, 0.0)
}

func TestWorkingMemory_ShouldPromote(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *WorkingMemory
		should bool
	}{
		{
			name: "should promote - high access count",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityMedium)
				wm.AccessCount = 10 // Threshold is 7 for medium
				wm.RecordAccess()
				return wm
			},
			should: true,
		},
		{
			name: "should promote - high importance",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityMedium)
				// Manually set high importance and exceed access threshold
				wm.ImportanceScore = 0.85
				wm.AccessCount = 10 // Exceeds medium priority threshold of 7
				return wm
			},
			should: true,
		},
		{
			name: "should promote - critical priority and accessed",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityCritical)
				wm.AccessCount = 3
				wm.RecordAccess()
				return wm
			},
			should: true,
		},
		{
			name: "should not promote - low access",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityLow)
				wm.AccessCount = 2
				return wm
			},
			should: false,
		},
		{
			name: "should not promote - already promoted",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityMedium)
				wm.AccessCount = 10
				now := time.Now()
				wm.PromotedAt = &now
				return wm
			},
			should: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := tt.setup()
			assert.Equal(t, tt.should, wm.ShouldPromote())
		})
	}
}

func TestWorkingMemory_calculateImportance(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *WorkingMemory
		minImportance float64
		maxImportance float64
	}{
		{
			name: "new memory - low importance",
			setup: func() *WorkingMemory {
				return NewWorkingMemory("session", "short", PriorityLow)
			},
			minImportance: 0.0,
			maxImportance: 0.3,
		},
		{
			name: "frequently accessed - high importance",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "longer content here", PriorityHigh)
				for range 10 {
					wm.RecordAccess()
					time.Sleep(1 * time.Millisecond)
				}
				return wm
			},
			minImportance: 0.5,
			maxImportance: 1.0,
		},
		{
			name: "critical priority - higher importance",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "critical content", PriorityCritical)
				wm.RecordAccess()
				return wm
			},
			minImportance: 0.3,
			maxImportance: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := tt.setup()
			importance := wm.ImportanceScore
			assert.GreaterOrEqual(t, importance, tt.minImportance)
			assert.LessOrEqual(t, importance, tt.maxImportance)
		})
	}
}

func TestWorkingMemory_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *WorkingMemory
		wantErr bool
	}{
		{
			name: "valid memory",
			setup: func() *WorkingMemory {
				return NewWorkingMemory("session", "content", PriorityMedium)
			},
			wantErr: false,
		},
		{
			name: "empty session ID",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityMedium)
				wm.SessionID = ""
				return wm
			},
			wantErr: true,
		},
		{
			name: "empty content",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityMedium)
				wm.Content = ""
				return wm
			},
			wantErr: true,
		},
		{
			name: "invalid priority",
			setup: func() *WorkingMemory {
				wm := NewWorkingMemory("session", "content", PriorityMedium)
				wm.Priority = MemoryPriority("invalid")
				return wm
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := tt.setup()
			err := wm.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWorkingMemoryStats(t *testing.T) {
	stats := WorkingMemoryStats{
		SessionID:        "test-session",
		TotalCount:       10,
		ActiveCount:      7,
		ExpiredCount:     2,
		PromotedCount:    1,
		AvgAccessCount:   5.5,
		AvgImportance:    0.65,
		PendingPromotion: 2,
		ByPriority: map[MemoryPriority]int{
			PriorityLow:    2,
			PriorityMedium: 5,
			PriorityHigh:   3,
		},
	}

	assert.Equal(t, "test-session", stats.SessionID)
	assert.Equal(t, 10, stats.TotalCount)
	assert.Equal(t, 7, stats.ActiveCount)
	assert.Equal(t, 2, stats.ExpiredCount)
	assert.Equal(t, 1, stats.PromotedCount)
	assert.Equal(t, 5.5, stats.AvgAccessCount)
	assert.Equal(t, 0.65, stats.AvgImportance)
	assert.Equal(t, 2, stats.PendingPromotion)
	assert.Len(t, stats.ByPriority, 3)
}

func TestWorkingMemory_ExtendTTL(t *testing.T) {
	wm := NewWorkingMemory("session", "content", PriorityMedium)
	originalExpiry := wm.ExpiresAt

	time.Sleep(10 * time.Millisecond)
	wm.ExpiresAt = time.Now().Add(wm.Priority.TTL())

	assert.True(t, wm.ExpiresAt.After(originalExpiry))
}

func TestWorkingMemory_Concurrency(t *testing.T) {
	wm := NewWorkingMemory("session", "content", PriorityMedium)

	// Test concurrent access recording
	done := make(chan bool)
	for range 10 {
		go func() {
			wm.RecordAccess()
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	assert.Equal(t, 10, wm.AccessCount)
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a smaller", 5, 10, 5},
		{"b smaller", 10, 5, 5},
		{"equal", 5, 5, 5},
		{"negative", -5, 5, -5},
		{"both negative", -10, -5, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkingMemory_IDGeneration(t *testing.T) {
	wm1 := NewWorkingMemory("session", "test content one", PriorityMedium)
	wm2 := NewWorkingMemory("session", "test content two", PriorityMedium)

	// IDs should be unique with different content
	assert.NotEqual(t, wm1.ID, wm2.ID)

	// IDs should have correct prefix (working_memory)
	assert.Contains(t, wm1.ID, "working_memory")
	assert.Contains(t, wm2.ID, "working_memory")
}

func TestWorkingMemory_TagsAndMetadata(t *testing.T) {
	wm := NewWorkingMemory("session", "content", PriorityMedium)

	// Test tags
	wm.Tags = append(wm.Tags, "custom-tag")
	assert.Contains(t, wm.Tags, "custom-tag")

	// Test metadata
	wm.Metadata["custom_key"] = "custom_value"
	assert.Equal(t, "custom_value", wm.Metadata["custom_key"])
}

func TestWorkingMemory_FullLifecycle(t *testing.T) {
	// Create
	wm := NewWorkingMemory("test-session", "Important information", PriorityHigh)
	require.NotNil(t, wm)
	require.NoError(t, wm.Validate())

	// Access multiple times
	for range 6 {
		wm.RecordAccess()
		time.Sleep(1 * time.Millisecond)
	}
	assert.Equal(t, 6, wm.AccessCount)
	assert.Greater(t, wm.ImportanceScore, 0.0)

	// Should promote (high priority threshold is 5)
	assert.True(t, wm.ShouldPromote())

	// Promote
	wm.MarkPromoted("mem-long-term-xyz")
	assert.True(t, wm.IsPromoted())
	assert.Equal(t, "mem-long-term-xyz", wm.PromotedToID)

	// Should not promote again
	assert.False(t, wm.ShouldPromote())
}
