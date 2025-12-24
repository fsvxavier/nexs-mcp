package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// TemporalService provides time travel and historical query capabilities.
type TemporalService struct {
	elementVersions  map[string]*domain.VersionHistory // elementID -> version history
	relationVersions map[string]*domain.VersionHistory // relationshipID -> version history
	confidenceDecay  *domain.ConfidenceDecay
	mu               sync.RWMutex
	logger           *slog.Logger
}

// TemporalConfig configures the temporal service behavior.
type TemporalConfig struct {
	DecayHalfLife time.Duration
	MinConfidence float64
}

// DefaultTemporalConfig returns a default configuration.
func DefaultTemporalConfig() TemporalConfig {
	return TemporalConfig{
		DecayHalfLife: 30 * 24 * time.Hour, // 30 days
		MinConfidence: 0.1,
	}
}

// NewTemporalService creates a new temporal service.
func NewTemporalService(config TemporalConfig, logger *slog.Logger) *TemporalService {
	return &TemporalService{
		elementVersions:  make(map[string]*domain.VersionHistory),
		relationVersions: make(map[string]*domain.VersionHistory),
		confidenceDecay:  domain.NewConfidenceDecay(config.DecayHalfLife, config.MinConfidence),
		logger:           logger,
	}
}

// ElementHistoryEntry represents a single historical record of an element.
type ElementHistoryEntry struct {
	Version     int                    `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Author      string                 `json:"author"`
	ChangeType  string                 `json:"change_type"`
	ElementData map[string]interface{} `json:"element_data"`
	Changes     map[string]interface{} `json:"changes,omitempty"` // nil for full snapshots
}

// RelationHistoryEntry represents a single historical record of a relationship.
type RelationHistoryEntry struct {
	Version            int                    `json:"version"`
	Timestamp          time.Time              `json:"timestamp"`
	Author             string                 `json:"author"`
	ChangeType         string                 `json:"change_type"`
	RelationshipData   map[string]interface{} `json:"relationship_data"`
	Changes            map[string]interface{} `json:"changes,omitempty"`
	OriginalConfidence float64                `json:"original_confidence"`
	DecayedConfidence  float64                `json:"decayed_confidence,omitempty"`
}

// GraphSnapshot represents the state of the entire graph at a point in time.
type GraphSnapshot struct {
	Timestamp     time.Time                         `json:"timestamp"`
	Elements      map[string]map[string]interface{} `json:"elements"`
	Relationships map[string]map[string]interface{} `json:"relationships"`
	DecayApplied  bool                              `json:"decay_applied"`
}

// RecordElementChange records a change to an element.
func (ts *TemporalService) RecordElementChange(
	ctx context.Context,
	elementID string,
	elementType domain.ElementType,
	elementData map[string]interface{},
	author string,
	changeType domain.ChangeType,
	message string,
) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	history, exists := ts.elementVersions[elementID]
	if !exists {
		history = domain.NewVersionHistory(elementID, elementType)
		ts.elementVersions[elementID] = history
	}

	// Create snapshot
	snapshot := &domain.VersionSnapshot{
		Version:    history.CurrentVersion + 1,
		Timestamp:  time.Now(),
		Author:     author,
		ChangeType: changeType,
		Message:    message,
	}

	// Determine if this should be a full snapshot
	shouldBeFullSnapshot := snapshot.Version == 1 ||
		snapshot.Version%history.SnapshotPolicy.FullSnapshotInterval == 0 ||
		changeType == domain.ChangeTypeMajor

	if shouldBeFullSnapshot {
		snapshot.FullData = elementData
	} else {
		// Calculate diff from previous version
		if history.CurrentVersion > 0 {
			prevData, err := history.ReconstructAtVersion(history.CurrentVersion)
			if err != nil {
				return fmt.Errorf("failed to reconstruct previous version: %w", err)
			}
			snapshot.Changes = computeDiff(prevData, elementData)
		} else {
			snapshot.FullData = elementData
		}
	}

	if err := history.AddSnapshot(snapshot); err != nil {
		ts.logger.Error("Failed to record element change",
			"error", err,
			"element_id", elementID,
		)
		return fmt.Errorf("failed to record element change: %w", err)
	}

	ts.logger.Debug("Recorded element change",
		"element_id", elementID,
		"version", history.CurrentVersion,
		"change_type", changeType,
	)

	return nil
}

// RecordRelationshipChange records a change to a relationship.
func (ts *TemporalService) RecordRelationshipChange(
	ctx context.Context,
	relationshipID string,
	relationshipData map[string]interface{},
	author string,
	changeType domain.ChangeType,
	message string,
) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	history, exists := ts.relationVersions[relationshipID]
	if !exists {
		// Use a generic type for relationships since they're not elements
		history = domain.NewVersionHistory(relationshipID, "relationship")
		ts.relationVersions[relationshipID] = history
	}

	// Create snapshot
	snapshot := &domain.VersionSnapshot{
		Version:    history.CurrentVersion + 1,
		Timestamp:  time.Now(),
		Author:     author,
		ChangeType: changeType,
		Message:    message,
	}

	// Determine if this should be a full snapshot
	shouldBeFullSnapshot := snapshot.Version == 1 ||
		snapshot.Version%history.SnapshotPolicy.FullSnapshotInterval == 0 ||
		changeType == domain.ChangeTypeMajor

	if shouldBeFullSnapshot {
		snapshot.FullData = relationshipData
	} else {
		// Calculate diff from previous version
		if history.CurrentVersion > 0 {
			prevData, err := history.ReconstructAtVersion(history.CurrentVersion)
			if err != nil {
				return fmt.Errorf("failed to reconstruct previous version: %w", err)
			}
			snapshot.Changes = computeDiff(prevData, relationshipData)
		} else {
			snapshot.FullData = relationshipData
		}
	}

	if err := history.AddSnapshot(snapshot); err != nil {
		ts.logger.Error("Failed to record relationship change",
			"error", err,
			"relationship_id", relationshipID,
		)
		return fmt.Errorf("failed to record relationship change: %w", err)
	}

	ts.logger.Debug("Recorded relationship change",
		"relationship_id", relationshipID,
		"version", history.CurrentVersion,
		"change_type", changeType,
	)

	return nil
}

// GetElementHistory retrieves the complete history of an element.
func (ts *TemporalService) GetElementHistory(
	ctx context.Context,
	elementID string,
	startTime *time.Time,
	endTime *time.Time,
) ([]ElementHistoryEntry, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	history, exists := ts.elementVersions[elementID]
	if !exists {
		return nil, fmt.Errorf("no history found for element: %s", elementID)
	}

	var snapshots []*domain.VersionSnapshot
	var err error

	if startTime != nil && endTime != nil {
		snapshots, err = history.GetTimeRange(*startTime, *endTime)
	} else {
		// Get all snapshots
		snapshots, err = history.GetVersionRange(1, history.CurrentVersion)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get version range: %w", err)
	}

	entries := make([]ElementHistoryEntry, 0, len(snapshots))
	for _, snapshot := range snapshots {
		entry := ElementHistoryEntry{
			Version:    snapshot.Version,
			Timestamp:  snapshot.Timestamp,
			Author:     snapshot.Author,
			ChangeType: string(snapshot.ChangeType),
			Changes:    snapshot.Changes,
		}

		// Reconstruct full data at this version
		fullData, err := history.ReconstructAtVersion(snapshot.Version)
		if err != nil {
			ts.logger.Warn("Failed to reconstruct element at version",
				"error", err,
				"element_id", elementID,
				"version", snapshot.Version,
			)
			continue
		}
		entry.ElementData = fullData

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetRelationshipHistory retrieves the complete history of a relationship with decay applied.
func (ts *TemporalService) GetRelationshipHistory(
	ctx context.Context,
	relationshipID string,
	startTime *time.Time,
	endTime *time.Time,
	applyDecay bool,
) ([]RelationHistoryEntry, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	history, exists := ts.relationVersions[relationshipID]
	if !exists {
		return nil, fmt.Errorf("no history found for relationship: %s", relationshipID)
	}

	var snapshots []*domain.VersionSnapshot
	var err error

	if startTime != nil && endTime != nil {
		snapshots, err = history.GetTimeRange(*startTime, *endTime)
	} else {
		// Get all snapshots
		snapshots, err = history.GetVersionRange(1, history.CurrentVersion)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get version range: %w", err)
	}

	entries := make([]RelationHistoryEntry, 0, len(snapshots))
	for _, snapshot := range snapshots {
		entry := RelationHistoryEntry{
			Version:    snapshot.Version,
			Timestamp:  snapshot.Timestamp,
			Author:     snapshot.Author,
			ChangeType: string(snapshot.ChangeType),
			Changes:    snapshot.Changes,
		}

		// Reconstruct full data at this version
		fullData, err := history.ReconstructAtVersion(snapshot.Version)
		if err != nil {
			ts.logger.Warn("Failed to reconstruct relationship at version",
				"error", err,
				"relationship_id", relationshipID,
				"version", snapshot.Version,
			)
			continue
		}
		entry.RelationshipData = fullData

		// Extract original confidence if available
		if confidence, ok := fullData["confidence"].(float64); ok {
			entry.OriginalConfidence = confidence

			// Apply decay if requested
			if applyDecay {
				decayed, err := ts.confidenceDecay.CalculateDecayWithReinforcement(
					relationshipID,
					confidence,
					snapshot.Timestamp,
				)
				if err == nil {
					entry.DecayedConfidence = decayed
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetElementAtTime retrieves an element's state at a specific point in time.
func (ts *TemporalService) GetElementAtTime(
	ctx context.Context,
	elementID string,
	targetTime time.Time,
) (map[string]interface{}, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	history, exists := ts.elementVersions[elementID]
	if !exists {
		return nil, fmt.Errorf("no history found for element: %s", elementID)
	}

	snapshot, err := history.GetSnapshotAtTime(targetTime)
	if err != nil {
		return nil, fmt.Errorf("no snapshot found for element %s at time %v: %w", elementID, targetTime, err)
	}

	// Reconstruct full data at this version
	fullData, err := history.ReconstructAtVersion(snapshot.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct element at time: %w", err)
	}

	return fullData, nil
}

// GetRelationshipAtTime retrieves a relationship's state at a specific point in time.
func (ts *TemporalService) GetRelationshipAtTime(
	ctx context.Context,
	relationshipID string,
	targetTime time.Time,
	applyDecay bool,
) (map[string]interface{}, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	history, exists := ts.relationVersions[relationshipID]
	if !exists {
		return nil, fmt.Errorf("no history found for relationship: %s", relationshipID)
	}

	snapshot, err := history.GetSnapshotAtTime(targetTime)
	if err != nil {
		return nil, fmt.Errorf("no snapshot found for relationship %s at time %v: %w", relationshipID, targetTime, err)
	}

	// Reconstruct full data at this version
	fullData, err := history.ReconstructAtVersion(snapshot.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct relationship at time: %w", err)
	}

	// Apply decay if requested
	if applyDecay {
		if confidence, ok := fullData["confidence"].(float64); ok {
			decayed, err := ts.confidenceDecay.CalculateDecayWithReinforcement(
				relationshipID,
				confidence,
				snapshot.Timestamp,
			)
			if err == nil {
				fullData["decayed_confidence"] = decayed
				fullData["original_confidence"] = confidence
			}
		}
	}

	return fullData, nil
}

// GetGraphAtTime reconstructs the entire graph state at a specific point in time.
func (ts *TemporalService) GetGraphAtTime(
	ctx context.Context,
	targetTime time.Time,
	applyDecay bool,
) (*GraphSnapshot, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	snapshot := &GraphSnapshot{
		Timestamp:     targetTime,
		Elements:      make(map[string]map[string]interface{}),
		Relationships: make(map[string]map[string]interface{}),
		DecayApplied:  applyDecay,
	}

	// Reconstruct all elements at target time
	for elementID, history := range ts.elementVersions {
		historySnapshot, err := history.GetSnapshotAtTime(targetTime)
		if err != nil {
			continue // Element didn't exist at this time
		}

		fullData, err := history.ReconstructAtVersion(historySnapshot.Version)
		if err != nil {
			ts.logger.Warn("Failed to reconstruct element for graph snapshot",
				"error", err,
				"element_id", elementID,
				"time", targetTime,
			)
			continue
		}

		snapshot.Elements[elementID] = fullData
	}

	// Reconstruct all relationships at target time
	for relationshipID, history := range ts.relationVersions {
		historySnapshot, err := history.GetSnapshotAtTime(targetTime)
		if err != nil {
			continue // Relationship didn't exist at this time
		}

		fullData, err := history.ReconstructAtVersion(historySnapshot.Version)
		if err != nil {
			ts.logger.Warn("Failed to reconstruct relationship for graph snapshot",
				"error", err,
				"relationship_id", relationshipID,
				"time", targetTime,
			)
			continue
		}

		// Apply decay if requested
		if applyDecay {
			if confidence, ok := fullData["confidence"].(float64); ok {
				decayed, err := ts.confidenceDecay.CalculateDecayWithReinforcement(
					relationshipID,
					confidence,
					historySnapshot.Timestamp,
				)
				if err == nil {
					fullData["decayed_confidence"] = decayed
					fullData["original_confidence"] = confidence
				}
			}
		}

		snapshot.Relationships[relationshipID] = fullData
	}

	ts.logger.Info("Reconstructed graph snapshot",
		"time", targetTime,
		"elements", len(snapshot.Elements),
		"relationships", len(snapshot.Relationships),
		"decay_applied", applyDecay,
	)

	return snapshot, nil
}

// GetDecayedGraph returns the current graph with confidence decay applied to all relationships.
func (ts *TemporalService) GetDecayedGraph(
	ctx context.Context,
	confidenceThreshold float64,
) (*GraphSnapshot, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	now := time.Now()
	snapshot := &GraphSnapshot{
		Timestamp:     now,
		Elements:      make(map[string]map[string]interface{}),
		Relationships: make(map[string]map[string]interface{}),
		DecayApplied:  true,
	}

	// Get current state of all elements
	for elementID, history := range ts.elementVersions {
		if history.CurrentVersion == 0 {
			continue
		}

		fullData, err := history.ReconstructAtVersion(history.CurrentVersion)
		if err != nil {
			ts.logger.Warn("Failed to reconstruct current element",
				"error", err,
				"element_id", elementID,
			)
			continue
		}

		snapshot.Elements[elementID] = fullData
	}

	// Prepare batch decay calculation for all relationships
	decayInputs := make([]domain.DecayInput, 0, len(ts.relationVersions))
	relationshipData := make(map[string]map[string]interface{})
	relationshipOrder := make([]string, 0, len(ts.relationVersions))

	for relationshipID, history := range ts.relationVersions {
		if history.CurrentVersion == 0 {
			continue
		}

		fullData, err := history.ReconstructAtVersion(history.CurrentVersion)
		if err != nil {
			ts.logger.Warn("Failed to reconstruct current relationship",
				"error", err,
				"relationship_id", relationshipID,
			)
			continue
		}

		confidence, ok := fullData["confidence"].(float64)
		if !ok {
			continue
		}

		// Get the timestamp of the current version
		currentSnapshot, err := history.GetSnapshot(history.CurrentVersion)
		if err != nil {
			continue
		}

		decayInputs = append(decayInputs, domain.DecayInput{
			RelationshipID:    relationshipID,
			InitialConfidence: confidence,
			CreatedAt:         currentSnapshot.Timestamp,
		})
		relationshipData[relationshipID] = fullData
		relationshipOrder = append(relationshipOrder, relationshipID)
	}

	// Batch calculate decay for all relationships
	decayOutputs, err := ts.confidenceDecay.BatchCalculateDecay(decayInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch calculate decay: %w", err)
	}

	// Apply decayed confidences and filter by threshold
	for i, output := range decayOutputs {
		relationshipID := relationshipOrder[i]
		if output.DecayedConfidence < confidenceThreshold {
			continue // Filter out relationships below threshold
		}

		data := relationshipData[relationshipID]
		data["decayed_confidence"] = output.DecayedConfidence
		data["original_confidence"] = output.InitialConfidence
		snapshot.Relationships[relationshipID] = data
	}

	ts.logger.Info("Generated decayed graph",
		"total_relationships", len(decayInputs),
		"filtered_relationships", len(snapshot.Relationships),
		"threshold", confidenceThreshold,
	)

	return snapshot, nil
}

// ReinforceRelationship adds a reinforcement event to a relationship (e.g., when accessed or used).
func (ts *TemporalService) ReinforceRelationship(ctx context.Context, relationshipID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.confidenceDecay.Reinforce(relationshipID)
	ts.logger.Debug("Reinforced relationship", "relationship_id", relationshipID)
	return nil
}

// GetVersionStats returns statistics about version storage.
func (ts *TemporalService) GetVersionStats(ctx context.Context) map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	totalElementVersions := 0
	totalRelationshipVersions := 0

	for _, history := range ts.elementVersions {
		totalElementVersions += history.TotalVersions
	}

	for _, history := range ts.relationVersions {
		totalRelationshipVersions += history.TotalVersions
	}

	return map[string]interface{}{
		"tracked_elements":            len(ts.elementVersions),
		"tracked_relationships":       len(ts.relationVersions),
		"total_element_versions":      totalElementVersions,
		"total_relationship_versions": totalRelationshipVersions,
		"total_versions":              totalElementVersions + totalRelationshipVersions,
		"decay_stats":                 ts.confidenceDecay.GetStats(),
	}
}

// ExportElementToJSON exports an element as JSON at a specific version.
func (ts *TemporalService) ExportElementToJSON(
	ctx context.Context,
	elementID string,
	version int,
) ([]byte, error) {
	ts.mu.RLock()
	history, exists := ts.elementVersions[elementID]
	ts.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no history found for element: %s", elementID)
	}

	data, err := history.ReconstructAtVersion(version)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct element: %w", err)
	}

	return json.MarshalIndent(data, "", "  ")
}

// computeDiff computes the difference between two maps.
func computeDiff(old, new map[string]interface{}) map[string]interface{} {
	diff := make(map[string]interface{})

	// Find changed and new fields
	for key, newValue := range new {
		oldValue, exists := old[key]
		if !exists || !deepEqual(oldValue, newValue) {
			diff[key] = newValue
		}
	}

	// Find deleted fields (represented as nil)
	for key := range old {
		if _, exists := new[key]; !exists {
			diff[key] = nil
		}
	}

	return diff
}

// deepEqual performs deep equality check for interface{} values.
func deepEqual(a, b interface{}) bool {
	aJSON, err1 := json.Marshal(a)
	bJSON, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(aJSON) == string(bJSON)
}
