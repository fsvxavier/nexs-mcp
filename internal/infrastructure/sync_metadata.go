package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SyncStateDir is the directory where sync state is stored
const SyncStateDir = ".nexs-sync"

// SyncStateFile is the filename for the sync state
const SyncStateFile = "state.json"

// FileState represents the sync state of a single file
type FileState struct {
	FilePath     string    `json:"file_path"`
	Checksum     string    `json:"checksum"`
	LastSyncedAt time.Time `json:"last_synced_at"`
	SyncStatus   string    `json:"sync_status"` // "synced", "modified", "conflicted", "pending"
	ElementID    string    `json:"element_id"`
	ElementType  string    `json:"element_type"`
}

// SyncHistory represents a single sync operation
type SyncHistory struct {
	Timestamp    time.Time         `json:"timestamp"`
	Direction    string            `json:"direction"` // "push", "pull"
	FilesChanged int               `json:"files_changed"`
	Status       string            `json:"status"` // "success", "partial", "failed"
	Error        string            `json:"error,omitempty"`
	Details      map[string]string `json:"details,omitempty"`
}

// SyncState represents the complete sync state for a repository
type SyncState struct {
	Version      string                 `json:"version"`
	LastSyncAt   time.Time              `json:"last_sync_at"`
	RemoteURL    string                 `json:"remote_url"`
	RemoteBranch string                 `json:"remote_branch"`
	Files        map[string]FileState   `json:"files"` // key is file path
	History      []SyncHistory          `json:"history"`
	Conflicts    []SyncConflict         `json:"conflicts,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NewSyncState creates a new SyncState
func NewSyncState(remoteURL, remoteBranch string) *SyncState {
	return &SyncState{
		Version:      "1.0.0",
		LastSyncAt:   time.Time{}, // Zero time indicates never synced
		RemoteURL:    remoteURL,
		RemoteBranch: remoteBranch,
		Files:        make(map[string]FileState),
		History:      []SyncHistory{},
		Conflicts:    []SyncConflict{},
		Metadata:     make(map[string]interface{}),
	}
}

// SyncMetadataManager manages sync metadata persistence
type SyncMetadataManager struct {
	baseDir string // Base directory for the repository
}

// NewSyncMetadataManager creates a new sync metadata manager
func NewSyncMetadataManager(baseDir string) *SyncMetadataManager {
	return &SyncMetadataManager{
		baseDir: baseDir,
	}
}

// GetStatePath returns the path to the sync state file
func (m *SyncMetadataManager) GetStatePath() string {
	return filepath.Join(m.baseDir, SyncStateDir, SyncStateFile)
}

// GetSyncDir returns the path to the sync directory
func (m *SyncMetadataManager) GetSyncDir() string {
	return filepath.Join(m.baseDir, SyncStateDir)
}

// SaveState saves the sync state to disk
func (m *SyncMetadataManager) SaveState(state *SyncState) error {
	// Create sync directory if it doesn't exist
	syncDir := m.GetSyncDir()
	if err := os.MkdirAll(syncDir, 0755); err != nil {
		return fmt.Errorf("failed to create sync directory: %w", err)
	}

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sync state: %w", err)
	}

	// Write to file
	statePath := m.GetStatePath()
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write sync state file: %w", err)
	}

	return nil
}

// LoadState loads the sync state from disk
func (m *SyncMetadataManager) LoadState() (*SyncState, error) {
	statePath := m.GetStatePath()

	// Check if file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		// Return empty state if file doesn't exist
		return &SyncState{
			Version:   "1.0.0",
			Files:     make(map[string]FileState),
			History:   []SyncHistory{},
			Conflicts: []SyncConflict{},
			Metadata:  make(map[string]interface{}),
		}, nil
	}

	// Read file
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sync state file: %w", err)
	}

	// Unmarshal JSON
	var state SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sync state: %w", err)
	}

	// Initialize maps if nil
	if state.Files == nil {
		state.Files = make(map[string]FileState)
	}
	if state.History == nil {
		state.History = []SyncHistory{}
	}
	if state.Conflicts == nil {
		state.Conflicts = []SyncConflict{}
	}
	if state.Metadata == nil {
		state.Metadata = make(map[string]interface{})
	}

	return &state, nil
}

// TrackFile adds or updates a file in the sync state
func (m *SyncMetadataManager) TrackFile(state *SyncState, filePath, checksum, elementID, elementType string) error {
	fileState := FileState{
		FilePath:     filePath,
		Checksum:     checksum,
		LastSyncedAt: time.Now(),
		SyncStatus:   "synced",
		ElementID:    elementID,
		ElementType:  elementType,
	}

	state.Files[filePath] = fileState
	return nil
}

// UntrackFile removes a file from the sync state
func (m *SyncMetadataManager) UntrackFile(state *SyncState, filePath string) error {
	delete(state.Files, filePath)
	return nil
}

// MarkFileModified marks a file as modified in the sync state
func (m *SyncMetadataManager) MarkFileModified(state *SyncState, filePath string) error {
	fileState, exists := state.Files[filePath]
	if !exists {
		return fmt.Errorf("file not tracked: %s", filePath)
	}

	fileState.SyncStatus = "modified"
	state.Files[filePath] = fileState
	return nil
}

// MarkFileConflicted marks a file as conflicted in the sync state
func (m *SyncMetadataManager) MarkFileConflicted(state *SyncState, filePath string) error {
	fileState, exists := state.Files[filePath]
	if !exists {
		return fmt.Errorf("file not tracked: %s", filePath)
	}

	fileState.SyncStatus = "conflicted"
	state.Files[filePath] = fileState
	return nil
}

// AddHistory adds a sync operation to the history
func (m *SyncMetadataManager) AddHistory(state *SyncState, history SyncHistory) {
	state.History = append(state.History, history)

	// Keep only last 100 history entries
	if len(state.History) > 100 {
		state.History = state.History[len(state.History)-100:]
	}
}

// GetFileState returns the state of a specific file
func (m *SyncMetadataManager) GetFileState(state *SyncState, filePath string) (FileState, bool) {
	fileState, exists := state.Files[filePath]
	return fileState, exists
}

// GetModifiedFiles returns all files that have been modified since last sync
func (m *SyncMetadataManager) GetModifiedFiles(state *SyncState) []string {
	modified := []string{}
	for filePath, fileState := range state.Files {
		if fileState.SyncStatus == "modified" {
			modified = append(modified, filePath)
		}
	}
	return modified
}

// GetConflictedFiles returns all files that have conflicts
func (m *SyncMetadataManager) GetConflictedFiles(state *SyncState) []string {
	conflicted := []string{}
	for filePath, fileState := range state.Files {
		if fileState.SyncStatus == "conflicted" {
			conflicted = append(conflicted, filePath)
		}
	}
	return conflicted
}

// UpdateLastSync updates the last sync timestamp
func (m *SyncMetadataManager) UpdateLastSync(state *SyncState) {
	state.LastSyncAt = time.Now()
}

// AddConflict adds a conflict to the state
func (m *SyncMetadataManager) AddConflict(state *SyncState, conflict SyncConflict) {
	state.Conflicts = append(state.Conflicts, conflict)
}

// ResolveConflict removes a conflict from the state
func (m *SyncMetadataManager) ResolveConflict(state *SyncState, elementID string) {
	filtered := []SyncConflict{}
	for _, c := range state.Conflicts {
		if c.ElementID != elementID {
			filtered = append(filtered, c)
		}
	}
	state.Conflicts = filtered
}

// GetUnresolvedConflicts returns all unresolved conflicts
func (m *SyncMetadataManager) GetUnresolvedConflicts(state *SyncState) []SyncConflict {
	unresolved := []SyncConflict{}
	for _, c := range state.Conflicts {
		if !c.IsResolved() {
			unresolved = append(unresolved, c)
		}
	}
	return unresolved
}

// IsFileTracked checks if a file is tracked in the sync state
func (m *SyncMetadataManager) IsFileTracked(state *SyncState, filePath string) bool {
	_, exists := state.Files[filePath]
	return exists
}

// GetTrackedFilesCount returns the number of tracked files
func (m *SyncMetadataManager) GetTrackedFilesCount(state *SyncState) int {
	return len(state.Files)
}

// Clear clears all sync state
func (m *SyncMetadataManager) Clear() error {
	statePath := m.GetStatePath()

	// Check if file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return nil // Nothing to clear
	}

	// Remove the file
	if err := os.Remove(statePath); err != nil {
		return fmt.Errorf("failed to remove sync state file: %w", err)
	}

	return nil
}
