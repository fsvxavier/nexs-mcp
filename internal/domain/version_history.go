package domain

import (
	"encoding/json"
	"errors"
	"time"
)

// VersionSnapshot represents a point-in-time snapshot of an element.
type VersionSnapshot struct {
	Version     int                    `json:"version"`      // Sequential version number
	Timestamp   time.Time              `json:"timestamp"`    // When this version was created
	Author      string                 `json:"author"`       // Who made this change
	ChangeType  ChangeType             `json:"change_type"`  // Type of change
	Changes     map[string]interface{} `json:"changes"`      // Field-level changes (diff)
	FullData    map[string]interface{} `json:"full_data"`    // Complete snapshot (optional, for major versions)
	Message     string                 `json:"message"`      // Change description
	ChecksumSHA string                 `json:"checksum_sha"` // SHA-256 of full_data
}

// ChangeType categorizes the type of change.
type ChangeType string

const (
	// ChangeTypeCreate indicates element creation.
	ChangeTypeCreate ChangeType = "create"
	// ChangeTypeUpdate indicates field updates.
	ChangeTypeUpdate ChangeType = "update"
	// ChangeTypeActivate indicates activation state change.
	ChangeTypeActivate ChangeType = "activate"
	// ChangeTypeDeactivate indicates deactivation.
	ChangeTypeDeactivate ChangeType = "deactivate"
	// ChangeTypeMajor indicates major version bump (breaking change).
	ChangeTypeMajor ChangeType = "major"
)

// VersionHistory tracks all versions of an element over time.
type VersionHistory struct {
	ElementID       string             `json:"element_id"`
	ElementType     ElementType        `json:"element_type"`
	CurrentVersion  int                `json:"current_version"`
	Snapshots       []*VersionSnapshot `json:"snapshots"`
	FirstCreated    time.Time          `json:"first_created"`
	LastModified    time.Time          `json:"last_modified"`
	TotalVersions   int                `json:"total_versions"`
	SnapshotPolicy  SnapshotPolicy     `json:"snapshot_policy"`
	RetentionPolicy *RetentionPolicy   `json:"retention_policy,omitempty"`
}

// SnapshotPolicy defines when to store full snapshots vs diffs.
type SnapshotPolicy struct {
	// FullSnapshotInterval - store full snapshot every N versions (e.g., every 10th version)
	FullSnapshotInterval int `json:"full_snapshot_interval"`
	// MaxDiffChain - max number of diffs before forcing full snapshot
	MaxDiffChain int `json:"max_diff_chain"`
	// MajorVersionFullSnapshot - always store full snapshot on major version changes
	MajorVersionFullSnapshot bool `json:"major_version_full_snapshot"`
}

// RetentionPolicy defines how long to keep historical versions.
type RetentionPolicy struct {
	// MaxVersions - maximum number of versions to retain (0 = unlimited)
	MaxVersions int `json:"max_versions"`
	// MaxAge - maximum age of versions to retain (0 = unlimited)
	MaxAge time.Duration `json:"max_age"`
	// KeepMilestones - always keep major version snapshots
	KeepMilestones bool `json:"keep_milestones"`
}

// DefaultSnapshotPolicy returns the default snapshot policy.
func DefaultSnapshotPolicy() SnapshotPolicy {
	return SnapshotPolicy{
		FullSnapshotInterval:     10,   // Full snapshot every 10 versions
		MaxDiffChain:             20,   // Max 20 diffs before forcing full snapshot
		MajorVersionFullSnapshot: true, // Always full snapshot on major changes
	}
}

// DefaultRetentionPolicy returns the default retention policy.
func DefaultRetentionPolicy() *RetentionPolicy {
	return &RetentionPolicy{
		MaxVersions:    100,                  // Keep last 100 versions
		MaxAge:         365 * 24 * time.Hour, // Keep 1 year
		KeepMilestones: true,                 // Always keep major versions
	}
}

// NewVersionHistory creates a new version history for an element.
func NewVersionHistory(elementID string, elementType ElementType) *VersionHistory {
	now := time.Now()
	return &VersionHistory{
		ElementID:       elementID,
		ElementType:     elementType,
		CurrentVersion:  0,
		Snapshots:       make([]*VersionSnapshot, 0),
		FirstCreated:    now,
		LastModified:    now,
		TotalVersions:   0,
		SnapshotPolicy:  DefaultSnapshotPolicy(),
		RetentionPolicy: DefaultRetentionPolicy(),
	}
}

// AddSnapshot adds a new version snapshot to the history.
func (vh *VersionHistory) AddSnapshot(snapshot *VersionSnapshot) error {
	if snapshot == nil {
		return errors.New("snapshot cannot be nil")
	}

	// Validate sequential versioning
	if snapshot.Version != vh.CurrentVersion+1 {
		return errors.New("version must be sequential")
	}

	// Apply retention policy before adding new snapshot
	if err := vh.applyRetentionPolicy(); err != nil {
		return err
	}

	vh.Snapshots = append(vh.Snapshots, snapshot)
	vh.CurrentVersion = snapshot.Version
	vh.LastModified = snapshot.Timestamp
	vh.TotalVersions++

	return nil
}

// GetSnapshot retrieves a specific version snapshot.
func (vh *VersionHistory) GetSnapshot(version int) (*VersionSnapshot, error) {
	if version < 1 || version > vh.CurrentVersion {
		return nil, errors.New("version out of range")
	}

	for _, snapshot := range vh.Snapshots {
		if snapshot.Version == version {
			return snapshot, nil
		}
	}

	return nil, errors.New("snapshot not found")
}

// GetSnapshotAtTime retrieves the snapshot closest to (but not after) the given time.
func (vh *VersionHistory) GetSnapshotAtTime(t time.Time) (*VersionSnapshot, error) {
	if t.Before(vh.FirstCreated) {
		return nil, errors.New("time is before element creation")
	}

	var closest *VersionSnapshot
	for _, snapshot := range vh.Snapshots {
		if snapshot.Timestamp.After(t) {
			break
		}
		closest = snapshot
	}

	if closest == nil {
		return nil, errors.New("no snapshot found for given time")
	}

	return closest, nil
}

// GetVersionRange retrieves snapshots within a version range (inclusive).
func (vh *VersionHistory) GetVersionRange(startVersion, endVersion int) ([]*VersionSnapshot, error) {
	if startVersion < 1 || endVersion > vh.CurrentVersion || startVersion > endVersion {
		return nil, errors.New("invalid version range")
	}

	result := make([]*VersionSnapshot, 0)
	for _, snapshot := range vh.Snapshots {
		if snapshot.Version >= startVersion && snapshot.Version <= endVersion {
			result = append(result, snapshot)
		}
	}

	return result, nil
}

// GetTimeRange retrieves snapshots within a time range (inclusive).
func (vh *VersionHistory) GetTimeRange(startTime, endTime time.Time) ([]*VersionSnapshot, error) {
	if startTime.After(endTime) {
		return nil, errors.New("start time must be before end time")
	}

	result := make([]*VersionSnapshot, 0)
	for _, snapshot := range vh.Snapshots {
		if (snapshot.Timestamp.Equal(startTime) || snapshot.Timestamp.After(startTime)) &&
			(snapshot.Timestamp.Equal(endTime) || snapshot.Timestamp.Before(endTime)) {
			result = append(result, snapshot)
		}
	}

	return result, nil
}

// ReconstructAtVersion reconstructs the full element state at a specific version
// by applying diffs from the most recent full snapshot.
func (vh *VersionHistory) ReconstructAtVersion(version int) (map[string]interface{}, error) {
	if version < 1 || version > vh.CurrentVersion {
		return nil, errors.New("version out of range")
	}

	// Find the most recent full snapshot at or before the target version
	var baseSnapshot *VersionSnapshot
	diffsToApply := make([]*VersionSnapshot, 0)

	for i := len(vh.Snapshots) - 1; i >= 0; i-- {
		snapshot := vh.Snapshots[i]
		if snapshot.Version > version {
			continue
		}
		if snapshot.Version == version {
			// If target is a full snapshot, return it directly
			if len(snapshot.FullData) > 0 {
				return snapshot.FullData, nil
			}
			diffsToApply = append([]*VersionSnapshot{snapshot}, diffsToApply...)
		} else if snapshot.Version < version {
			if len(snapshot.FullData) > 0 {
				baseSnapshot = snapshot
				break
			}
			diffsToApply = append([]*VersionSnapshot{snapshot}, diffsToApply...)
		}
	}

	if baseSnapshot == nil {
		return nil, errors.New("no base snapshot found for reconstruction")
	}

	// Start with base snapshot data
	result := make(map[string]interface{})
	for k, v := range baseSnapshot.FullData {
		result[k] = v
	}

	// Apply diffs in order
	for _, diff := range diffsToApply {
		for k, v := range diff.Changes {
			result[k] = v
		}
	}

	return result, nil
}

// applyRetentionPolicy removes old snapshots according to retention policy.
func (vh *VersionHistory) applyRetentionPolicy() error {
	if vh.RetentionPolicy == nil {
		return nil
	}

	policy := vh.RetentionPolicy
	toKeep := make([]*VersionSnapshot, 0)
	now := time.Now()

	for _, snapshot := range vh.Snapshots {
		keep := false

		// Always keep milestones (major versions) if policy says so
		if policy.KeepMilestones && snapshot.ChangeType == ChangeTypeMajor {
			keep = true
		}

		// Check age
		if policy.MaxAge > 0 {
			age := now.Sub(snapshot.Timestamp)
			if age <= policy.MaxAge {
				keep = true
			}
		}

		// Check version count (keep most recent N)
		if policy.MaxVersions > 0 {
			if len(vh.Snapshots)-len(toKeep) <= policy.MaxVersions {
				keep = true
			}
		}

		if keep {
			toKeep = append(toKeep, snapshot)
		}
	}

	vh.Snapshots = toKeep
	return nil
}

// ComputeDiff computes the difference between two element states.
func ComputeDiff(oldState, newState map[string]interface{}) map[string]interface{} {
	diff := make(map[string]interface{})

	// Find changed and new fields
	for key, newValue := range newState {
		if oldValue, exists := oldState[key]; !exists || !deepEqual(oldValue, newValue) {
			diff[key] = newValue
		}
	}

	// Mark deleted fields
	for key := range oldState {
		if _, exists := newState[key]; !exists {
			diff[key] = nil // nil indicates deletion
		}
	}

	return diff
}

// deepEqual checks if two values are deeply equal (handles nested structures).
func deepEqual(a, b interface{}) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aJSON) == string(bJSON)
}

// GetStats returns statistics about the version history.
func (vh *VersionHistory) GetStats() map[string]interface{} {
	fullSnapshots := 0
	diffSnapshots := 0
	totalSize := 0

	for _, snapshot := range vh.Snapshots {
		if len(snapshot.FullData) > 0 {
			fullSnapshots++
		} else {
			diffSnapshots++
		}
		// Estimate size
		data, _ := json.Marshal(snapshot)
		totalSize += len(data)
	}

	return map[string]interface{}{
		"total_versions":   vh.TotalVersions,
		"full_snapshots":   fullSnapshots,
		"diff_snapshots":   diffSnapshots,
		"estimated_size":   totalSize,
		"first_created":    vh.FirstCreated,
		"last_modified":    vh.LastModified,
		"retention_policy": vh.RetentionPolicy,
	}
}
