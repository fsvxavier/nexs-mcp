package infrastructure

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// ConflictResolutionStrategy defines how to handle sync conflicts
type ConflictResolutionStrategy string

const (
	// LocalWins keeps the local version in case of conflict
	LocalWins ConflictResolutionStrategy = "local-wins"
	// RemoteWins keeps the remote version in case of conflict
	RemoteWins ConflictResolutionStrategy = "remote-wins"
	// Manual requires manual resolution of conflicts
	Manual ConflictResolutionStrategy = "manual"
	// NewestWins keeps the version with the most recent timestamp
	NewestWins ConflictResolutionStrategy = "newest-wins"
	// MergeContent attempts to merge non-conflicting changes
	MergeContent ConflictResolutionStrategy = "merge-content"
)

// SyncConflict represents a conflict between local and remote versions
type SyncConflict struct {
	ElementID       string                     `json:"element_id"`
	FilePath        string                     `json:"file_path"`
	ConflictType    ConflictType               `json:"conflict_type"`
	LocalVersion    *domain.ElementMetadata    `json:"local_version"`
	RemoteVersion   *domain.ElementMetadata    `json:"remote_version"`
	LocalChecksum   string                     `json:"local_checksum"`
	RemoteChecksum  string                     `json:"remote_checksum"`
	DetectedAt      time.Time                  `json:"detected_at"`
	Resolution      ConflictResolutionStrategy `json:"resolution,omitempty"`
	ResolvedAt      *time.Time                 `json:"resolved_at,omitempty"`
	ResolvedBy      string                     `json:"resolved_by,omitempty"`
	ResolutionNotes string                     `json:"resolution_notes,omitempty"`
}

// ConflictType categorizes the type of conflict
type ConflictType string

const (
	// ModifyModify both local and remote were modified
	ModifyModify ConflictType = "modify-modify"
	// DeleteModify local deleted, remote modified
	DeleteModify ConflictType = "delete-modify"
	// ModifyDelete local modified, remote deleted
	ModifyDelete ConflictType = "modify-delete"
	// DeleteDelete both deleted (not a real conflict)
	DeleteDelete ConflictType = "delete-delete"
)

// ConflictDetector detects and resolves sync conflicts
type ConflictDetector struct {
	strategy ConflictResolutionStrategy
}

// NewConflictDetector creates a new conflict detector
func NewConflictDetector(strategy ConflictResolutionStrategy) *ConflictDetector {
	if strategy == "" {
		strategy = Manual // Default to manual resolution
	}
	return &ConflictDetector{
		strategy: strategy,
	}
}

// DetectConflicts compares local and remote elements to find conflicts
func (cd *ConflictDetector) DetectConflicts(
	localElements map[string]domain.Element,
	remoteElements map[string]domain.Element,
	lastSyncTime time.Time,
) ([]SyncConflict, error) {
	conflicts := make([]SyncConflict, 0)

	// Check for conflicts in elements present in both
	for id, localElem := range localElements {
		remoteElem, existsRemote := remoteElements[id]

		if !existsRemote {
			// Element deleted remotely
			localMeta := localElem.GetMetadata()
			if localMeta.UpdatedAt.After(lastSyncTime) {
				// Local was modified after last sync, remote deleted
				conflicts = append(conflicts, SyncConflict{
					ElementID:     id,
					FilePath:      "", // Will be filled by caller
					ConflictType:  ModifyDelete,
					LocalVersion:  &localMeta,
					RemoteVersion: nil,
					LocalChecksum: cd.calculateChecksum(localElem),
					DetectedAt:    time.Now(),
				})
			}
			continue
		}

		// Both exist - check for modifications
		localMeta := localElem.GetMetadata()
		remoteMeta := remoteElem.GetMetadata()

		localModified := localMeta.UpdatedAt.After(lastSyncTime)
		remoteModified := remoteMeta.UpdatedAt.After(lastSyncTime)

		if localModified && remoteModified {
			// Both modified - check if they're actually different
			localChecksum := cd.calculateChecksum(localElem)
			remoteChecksum := cd.calculateChecksum(remoteElem)

			if localChecksum != remoteChecksum {
				conflicts = append(conflicts, SyncConflict{
					ElementID:      id,
					FilePath:       "", // Will be filled by caller
					ConflictType:   ModifyModify,
					LocalVersion:   &localMeta,
					RemoteVersion:  &remoteMeta,
					LocalChecksum:  localChecksum,
					RemoteChecksum: remoteChecksum,
					DetectedAt:     time.Now(),
				})
			}
		}
	}

	// Check for elements deleted locally but modified remotely
	for id, remoteElem := range remoteElements {
		if _, existsLocal := localElements[id]; !existsLocal {
			remoteMeta := remoteElem.GetMetadata()
			if remoteMeta.UpdatedAt.After(lastSyncTime) {
				// Remote was modified after last sync, local deleted
				conflicts = append(conflicts, SyncConflict{
					ElementID:      id,
					FilePath:       "", // Will be filled by caller
					ConflictType:   DeleteModify,
					LocalVersion:   nil,
					RemoteVersion:  &remoteMeta,
					RemoteChecksum: cd.calculateChecksum(remoteElem),
					DetectedAt:     time.Now(),
				})
			}
		}
	}

	return conflicts, nil
}

// ResolveConflict resolves a single conflict based on the configured strategy
func (cd *ConflictDetector) ResolveConflict(
	conflict SyncConflict,
	localElem, remoteElem domain.Element,
) (domain.Element, ConflictResolutionStrategy, error) {
	strategy := cd.strategy
	if conflict.Resolution != "" {
		// Use conflict-specific resolution if set
		strategy = conflict.Resolution
	}

	switch strategy {
	case LocalWins:
		if localElem == nil {
			return nil, strategy, fmt.Errorf("local element is nil, cannot use local-wins strategy")
		}
		return localElem, strategy, nil

	case RemoteWins:
		if remoteElem == nil {
			return nil, strategy, fmt.Errorf("remote element is nil, cannot use remote-wins strategy")
		}
		return remoteElem, strategy, nil

	case NewestWins:
		if localElem == nil {
			return remoteElem, strategy, nil
		}
		if remoteElem == nil {
			return localElem, strategy, nil
		}
		localMeta := localElem.GetMetadata()
		remoteMeta := remoteElem.GetMetadata()
		if localMeta.UpdatedAt.After(remoteMeta.UpdatedAt) {
			return localElem, strategy, nil
		}
		return remoteElem, strategy, nil

	case MergeContent:
		// For merge, we need both elements
		if localElem == nil || remoteElem == nil {
			return cd.ResolveConflict(conflict, localElem, remoteElem)
		}
		// For now, use newest-wins as merge strategy
		// More sophisticated merging would require element-specific logic
		localMeta := localElem.GetMetadata()
		remoteMeta := remoteElem.GetMetadata()
		if localMeta.UpdatedAt.After(remoteMeta.UpdatedAt) {
			return localElem, strategy, nil
		}
		return remoteElem, strategy, nil

	case Manual:
		return nil, strategy, fmt.Errorf("manual resolution required for conflict: %s", conflict.ElementID)

	default:
		return nil, strategy, fmt.Errorf("unknown resolution strategy: %s", strategy)
	}
}

// calculateChecksum computes a SHA256 checksum of the element
func (cd *ConflictDetector) calculateChecksum(elem domain.Element) string {
	if elem == nil {
		return ""
	}

	// Use metadata as basis for checksum
	meta := elem.GetMetadata()
	data := fmt.Sprintf("%s|%s|%s|%v|%s",
		meta.ID,
		meta.Name,
		meta.Description,
		meta.UpdatedAt.Unix(),
		meta.Version,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CalculateFileChecksum computes SHA256 checksum of a file
func CalculateFileChecksum(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// CompareChecksums compares two checksums
func CompareChecksums(checksum1, checksum2 string) bool {
	return checksum1 == checksum2
}

// GetConflictSummary returns a human-readable summary of the conflict
func (c *SyncConflict) GetConflictSummary() string {
	switch c.ConflictType {
	case ModifyModify:
		return fmt.Sprintf("Both local and remote versions of '%s' were modified", c.ElementID)
	case DeleteModify:
		return fmt.Sprintf("Element '%s' was deleted locally but modified remotely", c.ElementID)
	case ModifyDelete:
		return fmt.Sprintf("Element '%s' was modified locally but deleted remotely", c.ElementID)
	case DeleteDelete:
		return fmt.Sprintf("Element '%s' was deleted both locally and remotely", c.ElementID)
	default:
		return fmt.Sprintf("Unknown conflict type for element '%s'", c.ElementID)
	}
}

// IsResolved returns true if the conflict has been resolved
func (c *SyncConflict) IsResolved() bool {
	return c.ResolvedAt != nil
}

// MarkResolved marks the conflict as resolved
func (c *SyncConflict) MarkResolved(strategy ConflictResolutionStrategy, resolvedBy, notes string) {
	now := time.Now()
	c.ResolvedAt = &now
	c.Resolution = strategy
	c.ResolvedBy = resolvedBy
	c.ResolutionNotes = notes
}
