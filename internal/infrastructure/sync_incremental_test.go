package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIncrementalSync(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	assert.NotNil(t, sync)
	assert.NotNil(t, sync.metadataManager)
	assert.NotNil(t, sync.conflictDetector)
	assert.NotNil(t, sync.repository)
	assert.Equal(t, tempDir, sync.baseDir)
}

func TestIncrementalSync_GetChangedFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Create some test elements
	agent1 := domain.NewAgent("Test Agent 1", "Description", "1.0.0", "test")
	err = repo.Create(agent1)
	require.NoError(t, err)

	// Create sync state
	state := NewSyncState("https://github.com/user/repo.git", "main")
	state.LastSyncAt = time.Now().Add(-1 * time.Hour)

	// Get changed files
	changedFiles, err := sync.GetChangedFiles(state)
	require.NoError(t, err)

	// Should have 1 changed file (agent1 is new and not tracked)
	assert.Greater(t, len(changedFiles), 0)
}

func TestIncrementalSync_SyncLocal_FirstSync(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Create a test element
	agent := domain.NewAgent("Test Agent", "Description", "1.0.0", "test")
	err = repo.Create(agent)
	require.NoError(t, err)

	// Perform first sync
	options := SyncOptions{
		Direction: SyncDirectionPush,
		DryRun:    false,
	}

	report, err := sync.SyncLocal(
		context.Background(),
		"https://github.com/user/repo.git",
		"main",
		options,
	)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Greater(t, report.FilesChanged, 0)

	// Verify state was saved
	state, err := sync.GetSyncState()
	require.NoError(t, err)
	assert.False(t, state.LastSyncAt.IsZero())
}

func TestIncrementalSync_SyncLocal_IncrementalSync(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Create initial state with tracked files
	state := NewSyncState("https://github.com/user/repo.git", "main")
	state.LastSyncAt = time.Now().Add(-1 * time.Hour)

	err = sync.metadataManager.SaveState(state)
	require.NoError(t, err)

	// Create a test element
	agent := domain.NewAgent("Test Agent", "Description", "1.0.0", "test")
	err = repo.Create(agent)
	require.NoError(t, err)

	// Perform incremental sync
	options := SyncOptions{
		Direction:     SyncDirectionPush,
		ForceFullSync: false,
		DryRun:        false,
	}

	report, err := sync.SyncLocal(
		context.Background(),
		"https://github.com/user/repo.git",
		"main",
		options,
	)
	require.NoError(t, err)
	assert.NotNil(t, report)
}

func TestIncrementalSync_SyncLocal_DryRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Create a test element
	agent := domain.NewAgent("Test Agent", "Description", "1.0.0", "test")
	err = repo.Create(agent)
	require.NoError(t, err)

	// Perform dry run
	options := SyncOptions{
		Direction: SyncDirectionPush,
		DryRun:    true,
	}

	report, err := sync.SyncLocal(
		context.Background(),
		"https://github.com/user/repo.git",
		"main",
		options,
	)
	require.NoError(t, err)
	assert.NotNil(t, report)

	// Verify state was NOT saved (dry run)
	state, err := sync.GetSyncState()
	require.NoError(t, err)
	assert.True(t, state.LastSyncAt.IsZero())
}

func TestIncrementalSync_SyncLocal_WithTypeFilters(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Create elements of different types
	agent := domain.NewAgent("Test Agent", "Description", "1.0.0", "test")
	err = repo.Create(agent)
	require.NoError(t, err)

	persona := domain.NewPersona("Test Persona", "Description", "1.0.0", "test")
	err = repo.Create(persona)
	require.NoError(t, err)

	// Sync only agents
	options := SyncOptions{
		Direction:    SyncDirectionPush,
		DryRun:       true,
		IncludeTypes: []domain.ElementType{domain.AgentElement},
	}

	report, err := sync.SyncLocal(
		context.Background(),
		"https://github.com/user/repo.git",
		"main",
		options,
	)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Greater(t, report.FilesSkipped, 0, "persona should be skipped")
}

func TestIncrementalSync_SyncLocal_WithProgressCallback(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Create a test element
	agent := domain.NewAgent("Test Agent", "Description", "1.0.0", "test")
	err = repo.Create(agent)
	require.NoError(t, err)

	// Track progress
	progressCalls := 0
	callback := func(message string, current, total int) {
		progressCalls++
		assert.NotEmpty(t, message)
	}

	options := SyncOptions{
		Direction:        SyncDirectionPush,
		DryRun:           true,
		ProgressCallback: callback,
	}

	_, err = sync.SyncLocal(
		context.Background(),
		"https://github.com/user/repo.git",
		"main",
		options,
	)
	require.NoError(t, err)
	assert.Greater(t, progressCalls, 0, "progress callback should be called")
}

func TestIncrementalSync_DetectConflicts(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Create sync state
	state := NewSyncState("https://github.com/user/repo.git", "main")
	state.LastSyncAt = time.Now().Add(-1 * time.Hour)
	err = sync.metadataManager.SaveState(state)
	require.NoError(t, err)

	// Create test elements
	localAgent := domain.NewAgent("Local Agent", "Local", "1.0.0", "test")
	remoteAgent := domain.NewAgent("Remote Agent", "Remote", "1.0.0", "test")

	localElements := map[string]domain.Element{
		"agent1": localAgent,
	}
	remoteElements := map[string]domain.Element{
		"agent1": remoteAgent,
	}

	// Detect conflicts
	conflicts, err := sync.DetectConflicts(localElements, remoteElements)
	require.NoError(t, err)

	// Should detect modify-modify conflict
	assert.Greater(t, len(conflicts), 0)
}

func TestIncrementalSync_ResolveConflicts(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Create test elements
	localAgent := domain.NewAgent("Local Agent", "Local", "1.0.0", "test")
	remoteAgent := domain.NewAgent("Remote Agent", "Remote", "1.0.0", "test")

	localElements := map[string]domain.Element{
		"agent1": localAgent,
	}
	remoteElements := map[string]domain.Element{
		"agent1": remoteAgent,
	}

	// Create conflicts
	conflicts := []SyncConflict{
		{
			ElementID:    "agent1",
			ConflictType: ModifyModify,
			DetectedAt:   time.Now(),
		},
	}

	// Resolve conflicts with local-wins strategy
	resolved, err := sync.ResolveConflicts(conflicts, localElements, remoteElements, LocalWins)
	require.NoError(t, err)
	assert.Len(t, resolved, 1)
	assert.Equal(t, localAgent, resolved["agent1"])
}

func TestIncrementalSync_ExtractElementIDFromPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	tests := []struct {
		path     string
		expected string
	}{
		{"agent/2025-01-01/agent_test_123.yaml", "agent_test_123"},
		{"persona/2025-01-01/persona_abc.yaml", "persona_abc"},
		{"test.yaml", "test"},
		{"path/to/element.json", "element"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := sync.extractElementIDFromPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIncrementalSync_ExtractElementTypeFromPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	tests := []struct {
		path     string
		expected domain.ElementType
	}{
		{"agent/2025-01-01/agent_test_123.yaml", domain.AgentElement},
		{"persona/2025-01-01/persona_abc.yaml", domain.PersonaElement},
		{"skill/2025-01-01/skill_xyz.yaml", domain.SkillElement},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := sync.extractElementTypeFromPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIncrementalSync_GetSyncState(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Get initial state (should be empty)
	state, err := sync.GetSyncState()
	require.NoError(t, err)
	assert.NotNil(t, state)
	assert.True(t, state.LastSyncAt.IsZero())

	// Create and save a state
	state.LastSyncAt = time.Now()
	err = sync.metadataManager.SaveState(state)
	require.NoError(t, err)

	// Get state again
	loadedState, err := sync.GetSyncState()
	require.NoError(t, err)
	assert.False(t, loadedState.LastSyncAt.IsZero())
}

func TestIncrementalSync_ClearSyncState(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := NewEnhancedFileElementRepository(tempDir, 100)
	require.NoError(t, err)
	sync := NewIncrementalSync(tempDir, repo)

	// Create and save a state
	state := NewSyncState("https://github.com/user/repo.git", "main")
	state.LastSyncAt = time.Now()
	err = sync.metadataManager.SaveState(state)
	require.NoError(t, err)

	// Clear state
	err = sync.ClearSyncState()
	require.NoError(t, err)

	// Verify state is cleared
	loadedState, err := sync.GetSyncState()
	require.NoError(t, err)
	assert.True(t, loadedState.LastSyncAt.IsZero())
}
