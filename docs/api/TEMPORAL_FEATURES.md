# Temporal Features API Reference

## Overview

The Temporal Features provide time travel capabilities for the nexs-mcp system, allowing you to:
- Track complete version history of elements and relationships
- Query historical states at any point in time
- Apply confidence decay to relationships over time
- Reconstruct graph snapshots from the past

## Architecture

### Domain Layer

#### VersionHistory

Tracks all versions of an element or relationship over time using a snapshot/diff compression strategy.

**Key Features:**
- Full snapshots at configurable intervals (default: every 10 versions)
- Delta compression for intermediate versions
- Time-based and version-based queries
- Retention policies for automatic cleanup
- Multiple change types: create, update, activate, deactivate, major

**Usage:**
```go
history := domain.NewVersionHistory("skill-python", domain.SkillElement)

snapshot := &domain.VersionSnapshot{
    Version:    1,
    Timestamp:  time.Now(),
    Author:     "user@example.com",
    ChangeType: domain.ChangeTypeCreate,
    Message:    "Initial creation",
    FullData:   map[string]interface{}{"name": "Python", "level": 5},
}

err := history.AddSnapshot(snapshot)
```

**Methods:**
- `AddSnapshot(snapshot *VersionSnapshot) error` - Record a new version
- `GetSnapshot(version int) (*VersionSnapshot, error)` - Get specific version
- `GetSnapshotAtTime(t time.Time) (*VersionSnapshot, error)` - Get version at time
- `GetVersionRange(start, end int) ([]*VersionSnapshot, error)` - Get version range
- `GetTimeRange(start, end time.Time) ([]*VersionSnapshot, error)` - Get time range
- `ReconstructAtVersion(version int) (map[string]interface{}, error)` - Reconstruct data

#### ConfidenceDecay

Implements time-based confidence decay for relationships with reinforcement learning.

**Key Features:**
- 4 decay functions: exponential, linear, logarithmic, step-based
- Critical relationship preservation (confidence >= threshold)
- Reinforcement learning with access tracking
- Batch processing for performance
- Future confidence projection

**Decay Functions:**

1. **Exponential** (default): `confidence * exp(-elapsed/halfLife)`
   - Natural decay, fast initially, slower over time
   - Best for general-purpose decay

2. **Linear**: `confidence * max(0, 1 - elapsed/halfLife)`
   - Constant decay rate
   - Predictable, easy to understand

3. **Logarithmic**: `confidence * (1 - log(1 + elapsed)/log(1 + halfLife))`
   - Slow initial decay, faster later
   - Good for critical relationships

4. **Step**: Discrete confidence levels at intervals
   - Predictable thresholds
   - Good for categorization

**Usage:**
```go
decay := domain.NewConfidenceDecay(30*24*time.Hour, 0.1) // 30 days, min 0.1
decay.Config.PreserveCritical = true
decay.Config.CriticalThreshold = 0.9

// Basic decay
decayed, err := decay.CalculateDecay(0.8, createdTime)

// With reinforcement
decay.Reinforce("rel-1", time.Now(), 0.1) // +10% boost
decayed, err := decay.CalculateDecayWithReinforcement("rel-1", 0.8, createdTime)

// Batch processing
results := decay.BatchCalculateDecay([]domain.ConfidenceItem{...})

// Project future
future, err := decay.ProjectFutureConfidence(0.8, createdTime, 90*24*time.Hour)
```

**Configuration:**
```go
type ConfidenceDecayConfig struct {
    DecayFunction      string        // "exponential", "linear", "logarithmic", "step"
    PreserveCritical   bool          // Protect high-confidence relationships
    CriticalThreshold  float64       // Threshold for preservation (default: 0.9)
    ReinforcementBonus float64       // Boost per access (default: 0.1)
    MaxReinforcement   float64       // Max reinforcement effect (default: 0.3)
    StepIntervals      []time.Duration // For step decay
    StepValues         []float64       // Confidence at each step
}
```

### Application Layer

#### TemporalService

High-level service for temporal operations, integrating version history and confidence decay.

**Key Features:**
- Unified API for temporal operations
- Element and relationship tracking
- Time travel queries
- Graph snapshot reconstruction
- Confidence decay integration
- Thread-safe operations

**Configuration:**
```go
config := application.DefaultTemporalConfig()
// config.DecayHalfLife = 30 * 24 * time.Hour
// config.MinConfidence = 0.1

service := application.NewTemporalService(config, logger)
```

**Core Methods:**

1. **Recording Changes:**
```go
// Record element change
err := service.RecordElementChange(
    ctx,
    "skill-python",
    domain.SkillElement,
    map[string]interface{}{"name": "Python", "level": 5},
    "user@example.com",
    domain.ChangeTypeCreate,
    "Initial creation",
)

// Record relationship change
err := service.RecordRelationshipChange(
    ctx,
    "rel-skill-persona",
    map[string]interface{}{
        "from": "skill-python",
        "to": "persona-dev",
        "type": "uses",
        "confidence": 0.95,
    },
    "user@example.com",
    domain.ChangeTypeCreate,
    "Link skill to persona",
)
```

2. **Querying History:**
```go
// Get element history
history, err := service.GetElementHistory(ctx, "skill-python", nil, nil)
// With time range
startTime := time.Now().Add(-30 * 24 * time.Hour)
endTime := time.Now()
history, err := service.GetElementHistory(ctx, "skill-python", &startTime, &endTime)

// Get relationship history (with optional decay)
history, err := service.GetRelationshipHistory(ctx, "rel-1", nil, nil, true)
```

3. **Time Travel:**
```go
// Get element at specific time
targetTime := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
data, err := service.GetElementAtTime(ctx, "skill-python", targetTime)

// Get relationship at specific time
data, err := service.GetRelationshipAtTime(ctx, "rel-1", targetTime, true)
```

4. **Graph Snapshots:**
```go
// Get graph at specific time
snapshot, err := service.GetGraphAtTime(ctx, targetTime, false)
// With decay applied
snapshot, err := service.GetGraphAtTime(ctx, targetTime, true)

// Get current graph with decay filtering
threshold := 0.5 // Only relationships with confidence >= 0.5
snapshot, err := service.GetDecayedGraph(ctx, threshold)
```

### MCP Layer

Four new MCP tools for temporal operations accessible via the Model Context Protocol.

#### 1. get_element_history

Retrieve complete version history of an element.

**Input:**
```json
{
  "element_id": "skill-python",
  "start_time": "2024-01-01T00:00:00Z",  // optional
  "end_time": "2024-12-31T23:59:59Z"     // optional
}
```

**Output:**
```json
{
  "element_id": "skill-python",
  "history": [
    {
      "version": 1,
      "timestamp": "2024-06-15T10:30:00Z",
      "author": "user@example.com",
      "change_type": "create",
      "element_data": {"name": "Python", "level": 1},
      "changes": null
    },
    {
      "version": 2,
      "timestamp": "2024-07-20T14:15:00Z",
      "author": "admin@example.com",
      "change_type": "update",
      "element_data": {"name": "Python", "level": 3},
      "changes": {"level": 3}
    }
  ],
  "total": 2
}
```

#### 2. get_relation_history

Retrieve complete version history of a relationship.

**Input:**
```json
{
  "relationship_id": "rel-skill-persona",
  "start_time": "2024-01-01T00:00:00Z",  // optional
  "end_time": "2024-12-31T23:59:59Z",    // optional
  "apply_decay": true                      // optional, default: false
}
```

**Output:**
```json
{
  "relationship_id": "rel-skill-persona",
  "history": [
    {
      "version": 1,
      "timestamp": "2024-06-15T10:30:00Z",
      "author": "user@example.com",
      "change_type": "create",
      "relationship_data": {
        "from": "skill-python",
        "to": "persona-dev",
        "type": "uses",
        "confidence": 0.95
      },
      "original_confidence": 0.95,
      "decayed_confidence": 0.87
    }
  ],
  "total": 1
}
```

#### 3. get_graph_at_time

Reconstruct the entire graph state at a specific point in time.

**Input:**
```json
{
  "target_time": "2024-06-15T14:30:00Z",
  "apply_decay": false  // optional, default: false
}
```

**Output:**
```json
{
  "timestamp": "2024-06-15T14:30:00Z",
  "elements": {
    "skill-python": {"name": "Python", "level": 3},
    "persona-dev": {"name": "Developer", "role": "Backend"}
  },
  "relationships": {
    "rel-1": {
      "from": "skill-python",
      "to": "persona-dev",
      "type": "uses",
      "confidence": 0.95
    }
  },
  "element_count": 2,
  "relationship_count": 1,
  "decay_applied": false
}
```

#### 4. get_decayed_graph

Get current graph with confidence decay applied and filtering.

**Input:**
```json
{
  "confidence_threshold": 0.5  // optional, default: 0.5, range: 0.0-1.0
}
```

**Output:**
```json
{
  "timestamp": "2024-12-24T10:00:00Z",
  "elements": {
    "skill-python": {"name": "Python", "level": 5}
  },
  "relationships": {
    "rel-1": {
      "from": "skill-python",
      "to": "persona-dev",
      "confidence": 0.85,
      "original_confidence": 0.95,
      "decayed_confidence": 0.85
    }
  },
  "element_count": 1,
  "relationship_count": 1,
  "confidence_threshold": 0.5,
  "total_relationships": 5,
  "filtered_out": 4
}
```

## Performance Considerations

### Snapshot Strategy

- **Full snapshots**: Every 10th version by default (configurable)
- **Diffs**: Lightweight for intermediate versions
- **Major changes**: Always create full snapshot for breaking changes

### Retention Policies

Configure automatic cleanup of old versions:

```go
policy := &domain.RetentionPolicy{
    MaxVersions:      100,           // Keep last 100 versions
    MaxAge:           365 * 24 * time.Hour, // Keep 1 year
    MinVersions:      10,            // Always keep at least 10
    CompactAfter:     30 * 24 * time.Hour,  // Compact old diffs
}

history.RetentionPolicy = policy
```

### Batch Operations

Use batch methods for better performance:

```go
// Batch decay calculation
items := []domain.ConfidenceItem{
    {ID: "rel-1", InitialConfidence: 0.9, CreatedAt: time1},
    {ID: "rel-2", InitialConfidence: 0.8, CreatedAt: time2},
}
results := decay.BatchCalculateDecay(items)
```

### Benchmarks

Performance metrics from test suite:

- `RecordElementChange`: ~5,766 ns/op
- `GetElementHistory`: ~23,335 ns/op (10 versions)
- `GetDecayedGraph`: ~13,789 ns/op (10 relationships)

All operations are thread-safe with minimal lock contention.

## Best Practices

### 1. Recording Changes

- Always provide meaningful commit messages
- Use appropriate change types (create, update, major)
- Record changes immediately after operations
- Include author information for audit trails

### 2. Confidence Decay

- Choose decay function based on use case:
  - **Exponential**: General purpose, natural decay
  - **Linear**: Predictable, uniform decay
  - **Logarithmic**: Preserve important relationships longer
  - **Step**: Categorical confidence levels

- Enable `PreserveCritical` for important relationships
- Use reinforcement learning for actively used relationships
- Set appropriate `MinConfidence` threshold (default: 0.1)

### 3. Time Travel Queries

- Use time ranges to limit history retrieval
- Cache frequently accessed snapshots
- Consider decay when querying relationship confidence
- Use `GetGraphAtTime` sparingly (expensive operation)

### 4. Graph Snapshots

- Filter by confidence threshold to reduce payload
- Apply decay only when needed (computational cost)
- Consider using versioning for complete graph snapshots
- Monitor filtered relationship counts

## Error Handling

All temporal operations return descriptive errors:

```go
// Element not found
_, err := service.GetElementHistory(ctx, "nonexistent", nil, nil)
// Returns: "no history found for element: nonexistent"

// Invalid time range
past := time.Now().Add(-1 * time.Hour)
_, err := service.GetElementAtTime(ctx, "skill-1", past)
// Returns: "no snapshot found for element skill-1 at time ..."

// Invalid confidence threshold
_, err := service.GetDecayedGraph(ctx, 1.5)
// Returns: "confidence threshold must be between 0.0 and 1.0"
```

## Thread Safety

All temporal services are thread-safe:

- Version histories use internal locking
- TemporalService uses RWMutex for concurrent access
- Batch operations are atomic per item
- Safe for use in concurrent MCP tool handlers

## Migration Guide

### From Non-Temporal System

1. Enable temporal tracking in your application:
```go
temporalService := application.NewTemporalService(
    application.DefaultTemporalConfig(),
    logger,
)
```

2. Start recording changes:
```go
// After creating/updating elements
err := temporalService.RecordElementChange(ctx, id, elementType, data, author, changeType, message)

// After creating/updating relationships
err := temporalService.RecordRelationshipChange(ctx, id, data, author, changeType, message)
```

3. Access MCP tools via your MCP client
4. Existing data won't have history until first change is recorded

### Backward Compatibility

- Temporal features are opt-in
- No impact on existing non-temporal operations
- Can be enabled per-element/relationship
- Historical queries return empty for untracked items

## See Also

- [User Guide: Time Travel](../user-guide/TIME_TRAVEL.md)
- [Architecture: Temporal System](../architecture/TEMPORAL.md)
- [Code Tour: Temporal Implementation](../development/CODE_TOUR.md#temporal-features)
