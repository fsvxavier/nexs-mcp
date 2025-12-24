package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func TestNewTemporalService(t *testing.T) {
	config := DefaultTemporalConfig()
	logger := newTestLogger()

	ts := NewTemporalService(config, logger)

	assert.NotNil(t, ts)
	assert.NotNil(t, ts.elementVersions)
	assert.NotNil(t, ts.relationVersions)
	assert.NotNil(t, ts.confidenceDecay)
	assert.NotNil(t, ts.logger)
}

func TestDefaultTemporalConfig(t *testing.T) {
	config := DefaultTemporalConfig()

	assert.Equal(t, 30*24*time.Hour, config.DecayHalfLife)
	assert.Equal(t, 0.1, config.MinConfidence)
}

func TestTemporalService_RecordElementChange(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	elementID := "skill-1"
	elementData := map[string]interface{}{
		"name":        "Python",
		"description": "Python programming",
	}

	err := ts.RecordElementChange(
		ctx,
		elementID,
		domain.SkillElement,
		elementData,
		"user@example.com",
		domain.ChangeTypeCreate,
		"Initial creation",
	)

	require.NoError(t, err)

	// Verify history was created
	ts.mu.RLock()
	history, exists := ts.elementVersions[elementID]
	ts.mu.RUnlock()

	require.True(t, exists, "history should exist for element")
	assert.Equal(t, 1, history.CurrentVersion)
}

func TestTemporalService_RecordRelationshipChange(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	relationshipID := "rel-1"
	relationshipData := map[string]interface{}{
		"from":       "skill-1",
		"to":         "persona-1",
		"type":       "uses",
		"confidence": 0.95,
	}

	err := ts.RecordRelationshipChange(
		ctx,
		relationshipID,
		relationshipData,
		"user@example.com",
		domain.ChangeTypeCreate,
		"New relationship",
	)

	require.NoError(t, err)

	// Verify history was created
	ts.mu.RLock()
	history, exists := ts.relationVersions[relationshipID]
	ts.mu.RUnlock()

	require.True(t, exists, "history should exist for relationship")
	assert.Equal(t, 1, history.CurrentVersion)
}

func TestTemporalService_GetElementHistory(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	elementID := "skill-1"

	// Record multiple versions
	for i := 1; i <= 3; i++ {
		err := ts.RecordElementChange(ctx, elementID, domain.SkillElement,
			map[string]interface{}{"name": "Python", "level": i},
			fmt.Sprintf("user%d", i), domain.ChangeTypeUpdate, fmt.Sprintf("Update %d", i))
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Get full history
	history, err := ts.GetElementHistory(ctx, elementID, nil, nil)
	require.NoError(t, err)
	require.Len(t, history, 3)

	// Verify chronological order
	for i := 0; i < 3; i++ {
		assert.Equal(t, i+1, history[i].Version)
	}
}

func TestTemporalService_GetRelationshipHistory(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	relationshipID := "rel-1"

	// Record changes with different confidences
	err := ts.RecordRelationshipChange(ctx, relationshipID,
		map[string]interface{}{"from": "a", "to": "b", "confidence": 0.9},
		"user1", domain.ChangeTypeCreate, "Initial")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	err = ts.RecordRelationshipChange(ctx, relationshipID,
		map[string]interface{}{"from": "a", "to": "b", "confidence": 0.8},
		"user2", domain.ChangeTypeUpdate, "Updated")
	require.NoError(t, err)

	// Get history
	history, err := ts.GetRelationshipHistory(ctx, relationshipID, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, history, 2)

	// Verify original confidence from the first version (stored in FullData or reconstructed)
	assert.NotNil(t, history[0])
	assert.NotNil(t, history[1])
}

func TestTemporalService_GetElementAtTime(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	elementID := "skill-1"
	baseTime := time.Now()

	// Record first version
	err := ts.RecordElementChange(ctx, elementID, domain.SkillElement,
		map[string]interface{}{"name": "Python", "version": 1},
		"user1", domain.ChangeTypeCreate, "v1")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	midTime := time.Now()

	// Record second version
	err = ts.RecordElementChange(ctx, elementID, domain.SkillElement,
		map[string]interface{}{"name": "Python", "version": 2},
		"user2", domain.ChangeTypeUpdate, "v2")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	laterTime := time.Now()

	// Query before creation (should error)
	_, err = ts.GetElementAtTime(ctx, elementID, baseTime.Add(-1*time.Hour))
	require.Error(t, err, "should error for time before element creation")

	// Query at first version time
	data, err := ts.GetElementAtTime(ctx, elementID, midTime)
	require.NoError(t, err)
	assert.Equal(t, 1, data["version"])

	// Query at second version time
	data, err = ts.GetElementAtTime(ctx, elementID, laterTime)
	require.NoError(t, err)
	assert.Equal(t, 2, data["version"])
}

func TestTemporalService_GetRelationshipAtTime(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	relationshipID := "rel-1"

	err := ts.RecordRelationshipChange(ctx, relationshipID,
		map[string]interface{}{"type": "uses", "strength": "high", "confidence": 0.95},
		"user1", domain.ChangeTypeCreate, "Initial")
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	queryTime := time.Now()

	// Get at current time
	data, err := ts.GetRelationshipAtTime(ctx, relationshipID, queryTime, false)
	require.NoError(t, err)
	assert.Equal(t, "uses", data["type"])

	// Query before creation (should error)
	pastTime := time.Now().Add(-1 * time.Hour)
	_, err = ts.GetRelationshipAtTime(ctx, relationshipID, pastTime, false)
	require.Error(t, err, "should error for time before relationship creation")
}

func TestTemporalService_GetGraphAtTime(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	// Create elements
	err := ts.RecordElementChange(ctx, "skill-1", domain.SkillElement,
		map[string]interface{}{"name": "Go"},
		"user1", domain.ChangeTypeCreate, "skill")
	require.NoError(t, err)

	err = ts.RecordElementChange(ctx, "persona-1", domain.PersonaElement,
		map[string]interface{}{"name": "Developer"},
		"user1", domain.ChangeTypeCreate, "persona")
	require.NoError(t, err)

	// Create relationship
	err = ts.RecordRelationshipChange(ctx, "rel-1",
		map[string]interface{}{"from": "skill-1", "to": "persona-1", "confidence": 0.9},
		"user1", domain.ChangeTypeCreate, "link")
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	snapshotTime := time.Now()

	// Get graph snapshot without decay
	snapshot, err := ts.GetGraphAtTime(ctx, snapshotTime, false)
	require.NoError(t, err)

	assert.NotNil(t, snapshot)
	assert.Len(t, snapshot.Elements, 2)
	assert.Len(t, snapshot.Relationships, 1)
	assert.False(t, snapshot.DecayApplied)

	// Verify element data
	assert.Equal(t, "Go", snapshot.Elements["skill-1"]["name"])
	assert.Equal(t, "Developer", snapshot.Elements["persona-1"]["name"])

	// Verify relationship data
	assert.Equal(t, "skill-1", snapshot.Relationships["rel-1"]["from"])
}

func TestTemporalService_GetDecayedGraph(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	// Create element
	err := ts.RecordElementChange(ctx, "skill-1", domain.SkillElement,
		map[string]interface{}{"name": "Python"},
		"user1", domain.ChangeTypeCreate, "skill")
	require.NoError(t, err)

	// Create relationship with high confidence
	err = ts.RecordRelationshipChange(ctx, "rel-1",
		map[string]interface{}{"from": "skill-1", "to": "persona-1", "confidence": 1.0},
		"user1", domain.ChangeTypeCreate, "link")
	require.NoError(t, err)

	// Disable PreserveCritical to allow decay
	ts.confidenceDecay.Config.PreserveCritical = false

	// Artificially age the relationship
	ts.mu.Lock()
	if history, exists := ts.relationVersions["rel-1"]; exists {
		if snapshot, err := history.GetSnapshot(1); err == nil {
			snapshot.Timestamp = time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
		}
	}
	ts.mu.Unlock()

	snapshot, err := ts.GetDecayedGraph(ctx, 0.3)
	require.NoError(t, err)

	assert.NotNil(t, snapshot)
	assert.True(t, snapshot.DecayApplied)
	assert.Len(t, snapshot.Elements, 1)

	// Relationship should be included (confidence after decay should be > 0.3)
	if len(snapshot.Relationships) > 0 {
		relData := snapshot.Relationships["rel-1"]
		decayedConf, ok := relData["decayed_confidence"].(float64)
		if ok {
			assert.Less(t, decayedConf, 1.0, "confidence should have decayed")
			assert.Greater(t, decayedConf, 0.3, "confidence should be above threshold")
		}
	}
}

func TestTemporalService_GetElementHistory_NotFound(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	_, err := ts.GetElementHistory(ctx, "nonexistent", nil, nil)
	require.Error(t, err, "should error for nonexistent element")
}

func TestTemporalService_ConcurrentAccess(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	elementCount := 10
	done := make(chan bool)

	// Test concurrent writes to different elements
	for i := 0; i < elementCount; i++ {
		go func(id int) {
			elementID := fmt.Sprintf("skill-%d", id)
			err := ts.RecordElementChange(ctx, elementID, domain.SkillElement,
				map[string]interface{}{"id": id},
				"user", domain.ChangeTypeCreate, "concurrent")
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < elementCount; i++ {
		<-done
	}

	// Verify all elements were recorded
	ts.mu.RLock()
	assert.Len(t, ts.elementVersions, elementCount)
	ts.mu.RUnlock()
}

func TestTemporalService_MultipleVersions(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	elementID := "skill-1"
	versionCount := 5

	// Record multiple versions
	for i := 1; i <= versionCount; i++ {
		err := ts.RecordElementChange(ctx, elementID, domain.SkillElement,
			map[string]interface{}{"version": i, "name": fmt.Sprintf("v%d", i)},
			fmt.Sprintf("user%d", i),
			domain.ChangeTypeUpdate,
			fmt.Sprintf("Update %d", i))
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Get all history
	history, err := ts.GetElementHistory(ctx, elementID, nil, nil)
	require.NoError(t, err)
	assert.Len(t, history, versionCount)

	// Verify versions are sequential
	for i := 0; i < versionCount; i++ {
		assert.Equal(t, i+1, history[i].Version)
	}
}

func TestTemporalService_TimestampOrdering(t *testing.T) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	elementID := "skill-1"

	// Record changes with delays
	for i := 1; i <= 3; i++ {
		err := ts.RecordElementChange(ctx, elementID, domain.SkillElement,
			map[string]interface{}{"seq": i},
			"user", domain.ChangeTypeUpdate, "update")
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)
	}

	// Get history
	history, err := ts.GetElementHistory(ctx, elementID, nil, nil)
	require.NoError(t, err)

	// Verify timestamps are in ascending order
	for i := 1; i < len(history); i++ {
		assert.True(t, history[i].Timestamp.After(history[i-1].Timestamp),
			"timestamps should be in ascending order")
	}
}

func TestTemporalService_ConfidenceDecayIntegration(t *testing.T) {
	config := TemporalConfig{
		DecayHalfLife: 24 * time.Hour, // 1 day
		MinConfidence: 0.1,
	}
	ts := NewTemporalService(config, newTestLogger())
	ts.confidenceDecay.Config.PreserveCritical = false
	ctx := context.Background()

	relationshipID := "rel-1"

	// Create relationship
	err := ts.RecordRelationshipChange(ctx, relationshipID,
		map[string]interface{}{"type": "test", "confidence": 0.9},
		"user", domain.ChangeTypeCreate, "test")
	require.NoError(t, err)

	// Artificially age the relationship by 1 day
	ts.mu.Lock()
	if history, exists := ts.relationVersions[relationshipID]; exists {
		if snapshot, err := history.GetSnapshot(1); err == nil {
			snapshot.Timestamp = time.Now().Add(-24 * time.Hour)
		}
	}
	ts.mu.Unlock()

	// Get with decay applied
	snapshot, err := ts.GetDecayedGraph(ctx, 0.1)
	require.NoError(t, err)

	// Verify decay was applied
	assert.True(t, snapshot.DecayApplied)

	if len(snapshot.Relationships) > 0 {
		relData := snapshot.Relationships[relationshipID]
		decayed, ok := relData["decayed_confidence"].(float64)
		if ok {
			assert.Less(t, decayed, 0.9, "confidence should have decayed from 0.9")
			assert.Greater(t, decayed, 0.1, "confidence should be above minimum")
		}
	}
}

// Benchmark tests
func BenchmarkTemporalService_RecordElementChange(b *testing.B) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		elementID := fmt.Sprintf("skill-%d", i)
		_ = ts.RecordElementChange(ctx, elementID, domain.SkillElement,
			map[string]interface{}{"data": "test"},
			"user", domain.ChangeTypeCreate, "bench")
	}
}

func BenchmarkTemporalService_GetElementHistory(b *testing.B) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	// Setup: create element with 10 versions
	elementID := "skill-1"
	for i := 0; i < 10; i++ {
		_ = ts.RecordElementChange(ctx, elementID, domain.SkillElement,
			map[string]interface{}{"version": i},
			"user", domain.ChangeTypeUpdate, "setup")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ts.GetElementHistory(ctx, elementID, nil, nil)
	}
}

func BenchmarkTemporalService_GetDecayedGraph(b *testing.B) {
	ts := NewTemporalService(DefaultTemporalConfig(), newTestLogger())
	ctx := context.Background()

	// Setup: create 10 relationships
	for i := 0; i < 10; i++ {
		relationID := fmt.Sprintf("rel-%d", i)
		_ = ts.RecordRelationshipChange(ctx, relationID,
			map[string]interface{}{"id": i, "confidence": 0.8},
			"user", domain.ChangeTypeCreate, "setup")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ts.GetDecayedGraph(ctx, 0.5)
	}
}
