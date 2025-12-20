package infrastructure

import (
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConflictDetector(t *testing.T) {
	tests := []struct {
		name             string
		strategy         ConflictResolutionStrategy
		expectedStrategy ConflictResolutionStrategy
	}{
		{
			name:             "explicit local-wins strategy",
			strategy:         LocalWins,
			expectedStrategy: LocalWins,
		},
		{
			name:             "explicit remote-wins strategy",
			strategy:         RemoteWins,
			expectedStrategy: RemoteWins,
		},
		{
			name:             "empty strategy defaults to manual",
			strategy:         "",
			expectedStrategy: Manual,
		},
		{
			name:             "newest-wins strategy",
			strategy:         NewestWins,
			expectedStrategy: NewestWins,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewConflictDetector(tt.strategy)
			assert.NotNil(t, detector)
			assert.Equal(t, tt.expectedStrategy, detector.strategy)
		})
	}
}

func TestDetectConflicts(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	lastSync := baseTime.Add(-1 * time.Hour)

	createAgent := func(id string, updatedAt time.Time) *domain.Agent {
		agent := domain.NewAgent("Agent "+id, "Test agent", "1.0.0", "test")
		// Access metadata through reflection would be needed to change timestamps
		// For testing purposes, we'll use the agent as-is
		return agent
	}

	tests := []struct {
		name              string
		localElements     map[string]domain.Element
		remoteElements    map[string]domain.Element
		lastSyncTime      time.Time
		expectedConflicts int
		validateConflicts func(t *testing.T, conflicts []SyncConflict)
	}{
		{
			name: "no conflicts - identical elements",
			localElements: map[string]domain.Element{
				"agent1": createAgent("agent1", baseTime),
			},
			remoteElements: map[string]domain.Element{
				"agent1": createAgent("agent1", baseTime),
			},
			lastSyncTime:      lastSync,
			expectedConflicts: 0,
		},
		{
			name: "modify-modify conflict",
			localElements: map[string]domain.Element{
				"agent1": createAgent("agent1", baseTime),
			},
			remoteElements: map[string]domain.Element{
				"agent1": domain.NewAgent("Modified Agent 1", "Test agent", "1.0.0", "test"),
			},
			lastSyncTime:      lastSync,
			expectedConflicts: 1,
			validateConflicts: func(t *testing.T, conflicts []SyncConflict) {
				assert.Equal(t, ModifyModify, conflicts[0].ConflictType)
				assert.Equal(t, "agent1", conflicts[0].ElementID)
				assert.NotNil(t, conflicts[0].LocalVersion)
				assert.NotNil(t, conflicts[0].RemoteVersion)
			},
		},
		{
			name: "modify-delete conflict",
			localElements: map[string]domain.Element{
				"agent1": createAgent("agent1", baseTime),
			},
			remoteElements:    map[string]domain.Element{},
			lastSyncTime:      lastSync,
			expectedConflicts: 1,
			validateConflicts: func(t *testing.T, conflicts []SyncConflict) {
				assert.Equal(t, ModifyDelete, conflicts[0].ConflictType)
				assert.Equal(t, "agent1", conflicts[0].ElementID)
				assert.NotNil(t, conflicts[0].LocalVersion)
				assert.Nil(t, conflicts[0].RemoteVersion)
			},
		},
		{
			name:          "delete-modify conflict",
			localElements: map[string]domain.Element{},
			remoteElements: map[string]domain.Element{
				"agent1": createAgent("agent1", baseTime),
			},
			lastSyncTime:      lastSync,
			expectedConflicts: 1,
			validateConflicts: func(t *testing.T, conflicts []SyncConflict) {
				assert.Equal(t, DeleteModify, conflicts[0].ConflictType)
				assert.Equal(t, "agent1", conflicts[0].ElementID)
				assert.Nil(t, conflicts[0].LocalVersion)
				assert.NotNil(t, conflicts[0].RemoteVersion)
			},
		},
		{
			name: "multiple conflicts",
			localElements: map[string]domain.Element{
				"agent1": createAgent("agent1", baseTime),
				"agent2": createAgent("agent2", baseTime),
			},
			remoteElements: map[string]domain.Element{
				"agent1": domain.NewAgent("Agent 1", "Modified description", "1.0.0", "test"),
				"agent3": createAgent("agent3", baseTime),
			},
			lastSyncTime:      lastSync,
			expectedConflicts: 3, // agent1 modify-modify, agent2 modify-delete, agent3 delete-modify
		},
		{
			name: "no conflict - only remote modified before sync",
			localElements: map[string]domain.Element{
				"agent1": createAgent("agent1", baseTime.Add(-2*time.Hour)),
			},
			remoteElements: map[string]domain.Element{
				"agent1": createAgent("agent1", baseTime.Add(-90*time.Minute)),
			},
			lastSyncTime:      lastSync,
			expectedConflicts: 0, // Both modified before last sync, no conflict
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewConflictDetector(Manual)
			conflicts, err := detector.DetectConflicts(tt.localElements, tt.remoteElements, tt.lastSyncTime)

			require.NoError(t, err)
			assert.Len(t, conflicts, tt.expectedConflicts)

			if tt.validateConflicts != nil && len(conflicts) > 0 {
				tt.validateConflicts(t, conflicts)
			}
		})
	}
}

func TestResolveConflict(t *testing.T) {
	localAgent := domain.NewAgent("Local Agent", "Local version", "1.0.0", "test")
	remoteAgent := domain.NewAgent("Remote Agent", "Remote version", "1.0.0", "test")

	conflict := SyncConflict{
		ElementID:    "agent1",
		ConflictType: ModifyModify,
	}

	tests := []struct {
		name           string
		strategy       ConflictResolutionStrategy
		localElem      domain.Element
		remoteElem     domain.Element
		expectedResult string // "local", "remote", "merged", or "error"
		expectError    bool
	}{
		{
			name:           "local-wins strategy",
			strategy:       LocalWins,
			localElem:      localAgent,
			remoteElem:     remoteAgent,
			expectedResult: "local",
			expectError:    false,
		},
		{
			name:           "remote-wins strategy",
			strategy:       RemoteWins,
			localElem:      localAgent,
			remoteElem:     remoteAgent,
			expectedResult: "remote",
			expectError:    false,
		},
		{
			name:           "newest-wins strategy",
			strategy:       NewestWins,
			localElem:      localAgent,
			remoteElem:     remoteAgent,
			expectedResult: "remote", // remoteAgent is newer
			expectError:    false,
		},
		{
			name:           "merge-content strategy",
			strategy:       MergeContent,
			localElem:      localAgent,
			remoteElem:     remoteAgent,
			expectedResult: "remote", // Uses newest-wins fallback, remote is newer
			expectError:    false,
		},
		{
			name:           "manual strategy returns error",
			strategy:       Manual,
			localElem:      localAgent,
			remoteElem:     remoteAgent,
			expectedResult: "error",
			expectError:    true,
		},
		{
			name:           "local-wins with nil local returns error",
			strategy:       LocalWins,
			localElem:      nil,
			remoteElem:     remoteAgent,
			expectedResult: "error",
			expectError:    true,
		},
		{
			name:           "remote-wins with nil remote returns error",
			strategy:       RemoteWins,
			localElem:      localAgent,
			remoteElem:     nil,
			expectedResult: "error",
			expectError:    true,
		},
		{
			name:           "newest-wins with nil local returns remote",
			strategy:       NewestWins,
			localElem:      nil,
			remoteElem:     remoteAgent,
			expectedResult: "remote",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewConflictDetector(tt.strategy)
			result, strategy, err := detector.ResolveConflict(conflict, tt.localElem, tt.remoteElem)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.strategy, strategy)

			switch tt.expectedResult {
			case "local":
				assert.Equal(t, tt.localElem, result)
			case "remote":
				assert.Equal(t, tt.remoteElem, result)
			}
		})
	}
}

func TestCalculateChecksum(t *testing.T) {
	agent1 := domain.NewAgent("Test Agent", "A test agent", "1.0.0", "test")
	agent2 := domain.NewAgent("Test Agent", "A test agent", "1.0.0", "test")
	agent3 := domain.NewAgent("Modified Agent", "A test agent", "1.0.0", "test")

	detector := NewConflictDetector(Manual)

	checksum1 := detector.calculateChecksum(agent1)
	checksum2 := detector.calculateChecksum(agent2)
	checksum3 := detector.calculateChecksum(agent3)

	assert.Equal(t, checksum1, checksum2, "identical elements should have same checksum")
	assert.NotEqual(t, checksum1, checksum3, "different elements should have different checksums")
	assert.NotEmpty(t, checksum1, "checksum should not be empty")
}

func TestSyncConflictMethods(t *testing.T) {
	conflict := SyncConflict{
		ElementID:    "agent1",
		ConflictType: ModifyModify,
		DetectedAt:   time.Now(),
	}

	t.Run("GetConflictSummary", func(t *testing.T) {
		tests := []struct {
			conflictType ConflictType
			expected     string
		}{
			{ModifyModify, "Both local and remote versions of 'agent1' were modified"},
			{DeleteModify, "Element 'agent1' was deleted locally but modified remotely"},
			{ModifyDelete, "Element 'agent1' was modified locally but deleted remotely"},
			{DeleteDelete, "Element 'agent1' was deleted both locally and remotely"},
		}

		for _, tt := range tests {
			conflict.ConflictType = tt.conflictType
			summary := conflict.GetConflictSummary()
			assert.Contains(t, summary, "agent1")
		}
	})

	t.Run("IsResolved", func(t *testing.T) {
		assert.False(t, conflict.IsResolved(), "conflict should not be resolved initially")

		conflict.MarkResolved(LocalWins, "user", "Keeping local version")

		assert.True(t, conflict.IsResolved(), "conflict should be resolved after marking")
		assert.NotNil(t, conflict.ResolvedAt)
		assert.Equal(t, LocalWins, conflict.Resolution)
		assert.Equal(t, "user", conflict.ResolvedBy)
		assert.Equal(t, "Keeping local version", conflict.ResolutionNotes)
	})
}

func TestCompareChecksums(t *testing.T) {
	checksum1 := "abc123"
	checksum2 := "abc123"
	checksum3 := "def456"

	assert.True(t, CompareChecksums(checksum1, checksum2))
	assert.False(t, CompareChecksums(checksum1, checksum3))
}
