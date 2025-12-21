package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSyncState(t *testing.T) {
	state := NewSyncState("https://github.com/user/repo.git", "main")

	assert.Equal(t, "1.0.0", state.Version)
	assert.Equal(t, "https://github.com/user/repo.git", state.RemoteURL)
	assert.Equal(t, "main", state.RemoteBranch)
	assert.True(t, state.LastSyncAt.IsZero())
	assert.NotNil(t, state.Files)
	assert.NotNil(t, state.History)
	assert.NotNil(t, state.Conflicts)
	assert.NotNil(t, state.Metadata)
	assert.Empty(t, state.Files)
	assert.Empty(t, state.History)
}

func TestSyncMetadataManager_SaveAndLoad(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "sync-metadata-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := NewSyncMetadataManager(tempDir)

	// Create a sync state
	state := NewSyncState("https://github.com/user/repo.git", "main")
	state.LastSyncAt = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	state.Files["test.yaml"] = FileState{
		FilePath:     "test.yaml",
		Checksum:     "abc123",
		LastSyncedAt: time.Now(),
		SyncStatus:   "synced",
		ElementID:    "agent1",
		ElementType:  "agent",
	}

	// Save state
	err = manager.SaveState(state)
	require.NoError(t, err)

	// Verify file exists
	statePath := manager.GetStatePath()
	assert.FileExists(t, statePath)

	// Load state
	loadedState, err := manager.LoadState()
	require.NoError(t, err)

	assert.Equal(t, state.Version, loadedState.Version)
	assert.Equal(t, state.RemoteURL, loadedState.RemoteURL)
	assert.Equal(t, state.RemoteBranch, loadedState.RemoteBranch)
	assert.Equal(t, state.LastSyncAt.Unix(), loadedState.LastSyncAt.Unix())
	assert.Len(t, loadedState.Files, 1)
	assert.Contains(t, loadedState.Files, "test.yaml")
}

func TestSyncMetadataManager_LoadNonExistent(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "sync-metadata-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := NewSyncMetadataManager(tempDir)

	// Load state from non-existent file
	state, err := manager.LoadState()
	require.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, "1.0.0", state.Version)
	assert.Empty(t, state.Files)
	assert.Empty(t, state.History)
}

func TestSyncMetadataManager_TrackFile(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	err := manager.TrackFile(state, "agent1.yaml", "checksum123", "agent1", "agent")
	require.NoError(t, err)

	assert.Len(t, state.Files, 1)
	assert.Contains(t, state.Files, "agent1.yaml")

	fileState := state.Files["agent1.yaml"]
	assert.Equal(t, "agent1.yaml", fileState.FilePath)
	assert.Equal(t, "checksum123", fileState.Checksum)
	assert.Equal(t, "synced", fileState.SyncStatus)
	assert.Equal(t, "agent1", fileState.ElementID)
	assert.Equal(t, "agent", fileState.ElementType)
}

func TestSyncMetadataManager_UntrackFile(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	// Track a file
	_ = manager.TrackFile(state, "agent1.yaml", "checksum123", "agent1", "agent")
	assert.Len(t, state.Files, 1)

	// Untrack the file
	err := manager.UntrackFile(state, "agent1.yaml")
	require.NoError(t, err)
	assert.Empty(t, state.Files)
}

func TestSyncMetadataManager_MarkFileModified(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	// Track a file
	_ = manager.TrackFile(state, "agent1.yaml", "checksum123", "agent1", "agent")

	// Mark as modified
	err := manager.MarkFileModified(state, "agent1.yaml")
	require.NoError(t, err)

	fileState := state.Files["agent1.yaml"]
	assert.Equal(t, "modified", fileState.SyncStatus)

	// Try to mark non-existent file
	err = manager.MarkFileModified(state, "nonexistent.yaml")
	assert.Error(t, err)
}

func TestSyncMetadataManager_MarkFileConflicted(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	// Track a file
	_ = manager.TrackFile(state, "agent1.yaml", "checksum123", "agent1", "agent")

	// Mark as conflicted
	err := manager.MarkFileConflicted(state, "agent1.yaml")
	require.NoError(t, err)

	fileState := state.Files["agent1.yaml"]
	assert.Equal(t, "conflicted", fileState.SyncStatus)

	// Try to mark non-existent file
	err = manager.MarkFileConflicted(state, "nonexistent.yaml")
	assert.Error(t, err)
}

func TestSyncMetadataManager_AddHistory(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	history := SyncHistory{
		Timestamp:    time.Now(),
		Direction:    "push",
		FilesChanged: 5,
		Status:       "success",
	}

	manager.AddHistory(state, history)
	assert.Len(t, state.History, 1)
	assert.Equal(t, "push", state.History[0].Direction)
	assert.Equal(t, 5, state.History[0].FilesChanged)

	// Test history limit (100 entries)
	for i := range 105 {
		manager.AddHistory(state, SyncHistory{
			Timestamp:    time.Now(),
			Direction:    "pull",
			FilesChanged: i,
			Status:       "success",
		})
	}
	assert.Len(t, state.History, 100, "history should be limited to 100 entries")
}

func TestSyncMetadataManager_GetFileState(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	// Track a file
	_ = manager.TrackFile(state, "agent1.yaml", "checksum123", "agent1", "agent")

	// Get existing file state
	fileState, exists := manager.GetFileState(state, "agent1.yaml")
	assert.True(t, exists)
	assert.Equal(t, "agent1.yaml", fileState.FilePath)

	// Get non-existent file state
	_, exists = manager.GetFileState(state, "nonexistent.yaml")
	assert.False(t, exists)
}

func TestSyncMetadataManager_GetModifiedFiles(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	// Track files
	_ = manager.TrackFile(state, "agent1.yaml", "checksum1", "agent1", "agent")
	_ = manager.TrackFile(state, "agent2.yaml", "checksum2", "agent2", "agent")
	_ = manager.TrackFile(state, "agent3.yaml", "checksum3", "agent3", "agent")

	// Mark some as modified
	_ = manager.MarkFileModified(state, "agent1.yaml")
	_ = manager.MarkFileModified(state, "agent3.yaml")

	modified := manager.GetModifiedFiles(state)
	assert.Len(t, modified, 2)
	assert.Contains(t, modified, "agent1.yaml")
	assert.Contains(t, modified, "agent3.yaml")
}

func TestSyncMetadataManager_GetConflictedFiles(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	// Track files
	_ = manager.TrackFile(state, "agent1.yaml", "checksum1", "agent1", "agent")
	_ = manager.TrackFile(state, "agent2.yaml", "checksum2", "agent2", "agent")

	// Mark one as conflicted
	_ = manager.MarkFileConflicted(state, "agent1.yaml")

	conflicted := manager.GetConflictedFiles(state)
	assert.Len(t, conflicted, 1)
	assert.Contains(t, conflicted, "agent1.yaml")
}

func TestSyncMetadataManager_UpdateLastSync(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	assert.True(t, state.LastSyncAt.IsZero())

	manager.UpdateLastSync(state)
	assert.False(t, state.LastSyncAt.IsZero())
}

func TestSyncMetadataManager_ConflictManagement(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	conflict1 := SyncConflict{
		ElementID:    "agent1",
		ConflictType: ModifyModify,
		DetectedAt:   time.Now(),
	}

	conflict2 := SyncConflict{
		ElementID:    "agent2",
		ConflictType: DeleteModify,
		DetectedAt:   time.Now(),
	}

	// Add conflicts
	manager.AddConflict(state, conflict1)
	manager.AddConflict(state, conflict2)
	assert.Len(t, state.Conflicts, 2)

	// Get unresolved conflicts
	unresolved := manager.GetUnresolvedConflicts(state)
	assert.Len(t, unresolved, 2)

	// Resolve a conflict
	state.Conflicts[0].MarkResolved(LocalWins, "user", "keeping local")
	unresolved = manager.GetUnresolvedConflicts(state)
	assert.Len(t, unresolved, 1)

	// Remove conflict
	manager.ResolveConflict(state, "agent1")
	assert.Len(t, state.Conflicts, 1)
	assert.Equal(t, "agent2", state.Conflicts[0].ElementID)
}

func TestSyncMetadataManager_IsFileTracked(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	assert.False(t, manager.IsFileTracked(state, "agent1.yaml"))

	_ = manager.TrackFile(state, "agent1.yaml", "checksum123", "agent1", "agent")
	assert.True(t, manager.IsFileTracked(state, "agent1.yaml"))
}

func TestSyncMetadataManager_GetTrackedFilesCount(t *testing.T) {
	manager := NewSyncMetadataManager("")
	state := NewSyncState("https://github.com/user/repo.git", "main")

	assert.Equal(t, 0, manager.GetTrackedFilesCount(state))

	_ = manager.TrackFile(state, "agent1.yaml", "checksum1", "agent1", "agent")
	_ = manager.TrackFile(state, "agent2.yaml", "checksum2", "agent2", "agent")
	assert.Equal(t, 2, manager.GetTrackedFilesCount(state))
}

func TestSyncMetadataManager_Clear(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "sync-metadata-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := NewSyncMetadataManager(tempDir)

	// Create and save a state
	state := NewSyncState("https://github.com/user/repo.git", "main")
	err = manager.SaveState(state)
	require.NoError(t, err)

	statePath := manager.GetStatePath()
	assert.FileExists(t, statePath)

	// Clear state
	err = manager.Clear()
	require.NoError(t, err)

	_, err = os.Stat(statePath)
	assert.True(t, os.IsNotExist(err))

	// Clear again should not error
	err = manager.Clear()
	require.NoError(t, err)
}

func TestSyncMetadataManager_GetSyncDir(t *testing.T) {
	manager := NewSyncMetadataManager("/test/dir")

	expected := filepath.Join("/test/dir", SyncStateDir)
	assert.Equal(t, expected, manager.GetSyncDir())
}

func TestSyncMetadataManager_GetStatePath(t *testing.T) {
	manager := NewSyncMetadataManager("/test/dir")

	expected := filepath.Join("/test/dir", SyncStateDir, SyncStateFile)
	assert.Equal(t, expected, manager.GetStatePath())
}
