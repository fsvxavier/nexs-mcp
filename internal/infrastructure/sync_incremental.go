package infrastructure

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// SyncDirection indicates the direction of synchronization.
type SyncDirection string

const (
	// SyncDirectionPush syncs from local to remote.
	SyncDirectionPush SyncDirection = "push"
	// SyncDirectionPull syncs from remote to local.
	SyncDirectionPull SyncDirection = "pull"
	// SyncDirectionBidirectional syncs in both directions.
	SyncDirectionBidirectional SyncDirection = "bidirectional"
)

// SyncProgressCallback is called during sync to report progress.
type SyncProgressCallback func(message string, current, total int)

// SyncOptions contains options for incremental sync.
type SyncOptions struct {
	Direction            SyncDirection
	ConflictStrategy     ConflictResolutionStrategy
	DryRun               bool
	ProgressCallback     SyncProgressCallback
	IncludeTypes         []domain.ElementType // If empty, sync all types
	ExcludeTypes         []domain.ElementType
	ForceFullSync        bool // If true, ignore metadata and sync everything
	AutoResolveConflicts bool // If true, automatically resolve using strategy
}

// SyncReport contains the results of a sync operation.
type SyncReport struct {
	Direction         SyncDirection
	FilesScanned      int
	FilesChanged      int
	FilesAdded        int
	FilesDeleted      int
	FilesSkipped      int
	ConflictsFound    int
	ConflictsResolved int
	Errors            []string
	Conflicts         []SyncConflict
	StartTime         string
	EndTime           string
	Duration          string
}

// IncrementalSync manages incremental synchronization with conflict detection.
type IncrementalSync struct {
	metadataManager  *SyncMetadataManager
	conflictDetector *ConflictDetector
	repository       *EnhancedFileElementRepository
	baseDir          string
}

// NewIncrementalSync creates a new incremental sync manager.
func NewIncrementalSync(
	baseDir string,
	repository *EnhancedFileElementRepository,
) *IncrementalSync {
	return &IncrementalSync{
		metadataManager:  NewSyncMetadataManager(baseDir),
		conflictDetector: NewConflictDetector(Manual), // Default to manual resolution
		repository:       repository,
		baseDir:          baseDir,
	}
}

// GetChangedFiles returns files that have changed since last sync.
func (is *IncrementalSync) GetChangedFiles(state *SyncState) ([]string, error) {
	changedFiles := []string{}

	// Get all elements from repository
	filter := domain.ElementFilter{
		Offset: 0,
		Limit:  10000,
	}
	elements, err := is.repository.List(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list elements: %w", err)
	}

	for _, elem := range elements {
		meta := elem.GetMetadata()

		// Build file path (this should match repository's file path logic)
		filename := meta.ID + ".yaml"
		relPath := filepath.Join(string(meta.Type),
			meta.UpdatedAt.Format("2006-01-02"),
			filename)

		// Check if file is tracked
		fileState, tracked := is.metadataManager.GetFileState(state, relPath)

		if !tracked {
			// New file, not tracked yet
			changedFiles = append(changedFiles, relPath)
			continue
		}

		// Check if file has been modified since last sync
		if meta.UpdatedAt.After(fileState.LastSyncedAt) {
			changedFiles = append(changedFiles, relPath)
		}
	}

	return changedFiles, nil
}

// SyncLocal performs incremental sync of local files.
func (is *IncrementalSync) SyncLocal(
	ctx context.Context,
	remoteURL, remoteBranch string,
	options SyncOptions,
) (*SyncReport, error) {
	report := &SyncReport{
		Direction: options.Direction,
		Errors:    []string{},
		Conflicts: []SyncConflict{},
	}

	// Load sync state
	state, err := is.metadataManager.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load sync state: %w", err)
	}

	// Initialize state if empty
	if state.RemoteURL == "" {
		state.RemoteURL = remoteURL
		state.RemoteBranch = remoteBranch
	}

	// Force full sync if requested or if it's first sync
	if options.ForceFullSync || state.LastSyncAt.IsZero() {
		return is.performFullSync(ctx, state, options, report)
	}

	// Get changed files
	changedFiles, err := is.GetChangedFiles(state)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	report.FilesScanned = len(state.Files)
	report.FilesChanged = len(changedFiles)

	// Report progress
	if options.ProgressCallback != nil {
		options.ProgressCallback("Scanning for changes...", 0, len(changedFiles))
	}

	// Process each changed file
	for i, filePath := range changedFiles {
		if options.ProgressCallback != nil {
			options.ProgressCallback(
				fmt.Sprintf("Processing %s...", filepath.Base(filePath)),
				i+1,
				len(changedFiles),
			)
		}

		// Get element from repository
		// Note: This is simplified - in real implementation, we'd need to
		// extract element ID from path and load the element
		elementID := is.extractElementIDFromPath(filePath)
		if elementID == "" {
			report.Errors = append(report.Errors,
				"could not extract element ID from path: "+filePath)
			report.FilesSkipped++
			continue
		}

		// Load element (we'll need to determine type from path)
		// elementType := is.extractElementTypeFromPath(filePath)
		element, err := is.repository.GetByID(elementID)
		if err != nil {
			report.Errors = append(report.Errors,
				fmt.Sprintf("failed to load element %s: %v", elementID, err))
			report.FilesSkipped++
			continue
		}

		// Calculate checksum
		checksum := is.calculateElementChecksum(element)

		// Update tracking
		err = is.metadataManager.TrackFile(state, filePath, checksum, elementID, string(element.GetType()))
		if err != nil {
			report.Errors = append(report.Errors,
				fmt.Sprintf("failed to track file %s: %v", filePath, err))
		}
	}

	// Update last sync time
	if !options.DryRun {
		is.metadataManager.UpdateLastSync(state)

		// Save state
		if err := is.metadataManager.SaveState(state); err != nil {
			return nil, fmt.Errorf("failed to save sync state: %w", err)
		}
	}

	return report, nil
}

// performFullSync performs a complete synchronization.
func (is *IncrementalSync) performFullSync(
	ctx context.Context,
	state *SyncState,
	options SyncOptions,
	report *SyncReport,
) (*SyncReport, error) {
	// Get all elements from repository
	filter := domain.ElementFilter{
		Offset: 0,
		Limit:  10000,
	}
	elements, err := is.repository.List(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list elements: %w", err)
	}

	report.FilesScanned = len(elements)

	if options.ProgressCallback != nil {
		options.ProgressCallback("Performing full sync...", 0, len(elements))
	}

	for i, elem := range elements {
		meta := elem.GetMetadata()

		// Apply type filters
		if len(options.IncludeTypes) > 0 {
			included := false
			for _, t := range options.IncludeTypes {
				if meta.Type == t {
					included = true
					break
				}
			}
			if !included {
				report.FilesSkipped++
				continue
			}
		}

		for _, t := range options.ExcludeTypes {
			if meta.Type == t {
				report.FilesSkipped++
				continue
			}
		}

		// Build file path
		filename := meta.ID + ".yaml"
		relPath := filepath.Join(string(meta.Type),
			meta.UpdatedAt.Format("2006-01-02"),
			filename)

		if options.ProgressCallback != nil {
			options.ProgressCallback(
				fmt.Sprintf("Syncing %s...", meta.Name),
				i+1,
				len(elements),
			)
		}

		// Calculate checksum
		checksum := is.calculateElementChecksum(elem)

		// Track file
		if !options.DryRun {
			err := is.metadataManager.TrackFile(state, relPath, checksum, meta.ID, string(meta.Type))
			if err != nil {
				report.Errors = append(report.Errors,
					fmt.Sprintf("failed to track file %s: %v", relPath, err))
				continue
			}
		}

		report.FilesChanged++
	}

	if !options.DryRun {
		is.metadataManager.UpdateLastSync(state)
		if err := is.metadataManager.SaveState(state); err != nil {
			return nil, fmt.Errorf("failed to save sync state: %w", err)
		}
	}

	return report, nil
}

// DetectConflicts checks for conflicts between local and remote elements.
func (is *IncrementalSync) DetectConflicts(
	localElements map[string]domain.Element,
	remoteElements map[string]domain.Element,
) ([]SyncConflict, error) {
	// Load sync state to get last sync time
	state, err := is.metadataManager.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load sync state: %w", err)
	}

	// Use conflict detector
	conflicts, err := is.conflictDetector.DetectConflicts(
		localElements,
		remoteElements,
		state.LastSyncAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to detect conflicts: %w", err)
	}

	return conflicts, nil
}

// ResolveConflicts resolves conflicts using the specified strategy.
func (is *IncrementalSync) ResolveConflicts(
	conflicts []SyncConflict,
	localElements, remoteElements map[string]domain.Element,
	strategy ConflictResolutionStrategy,
) (map[string]domain.Element, error) {
	resolved := make(map[string]domain.Element)

	// Update detector strategy
	detector := NewConflictDetector(strategy)

	for _, conflict := range conflicts {
		localElem := localElements[conflict.ElementID]
		remoteElem := remoteElements[conflict.ElementID]

		result, _, err := detector.ResolveConflict(conflict, localElem, remoteElem)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve conflict for %s: %w", conflict.ElementID, err)
		}

		resolved[conflict.ElementID] = result
	}

	return resolved, nil
}

// extractElementIDFromPath extracts element ID from file path.
func (is *IncrementalSync) extractElementIDFromPath(path string) string {
	// Extract filename
	filename := filepath.Base(path)

	// Remove extension
	ext := filepath.Ext(filename)
	if ext != "" {
		filename = filename[:len(filename)-len(ext)]
	}

	return filename
}

// extractElementTypeFromPath extracts element type from file path.
func (is *IncrementalSync) extractElementTypeFromPath(path string) domain.ElementType {
	// Get first directory component
	dir := filepath.Dir(path)
	for dir != "." && dir != "/" {
		base := filepath.Base(dir)
		parent := filepath.Dir(dir)

		// Check if this is a type directory (no more parents or parent is date)
		if parent == "." || parent == "/" {
			return domain.ElementType(base)
		}

		dir = parent
	}

	return ""
}

// calculateElementChecksum calculates checksum for an element.
func (is *IncrementalSync) calculateElementChecksum(elem domain.Element) string {
	// Use the conflict detector's checksum method
	detector := NewConflictDetector(Manual)
	return detector.calculateChecksum(elem)
}

// GetSyncState returns the current sync state.
func (is *IncrementalSync) GetSyncState() (*SyncState, error) {
	return is.metadataManager.LoadState()
}

// ClearSyncState clears all sync metadata.
func (is *IncrementalSync) ClearSyncState() error {
	return is.metadataManager.Clear()
}
