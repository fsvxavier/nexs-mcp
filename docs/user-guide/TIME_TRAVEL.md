# Time Travel User Guide

## Introduction

The Time Travel feature allows you to explore the evolution of your knowledge graph over time, understanding how elements and relationships changed, and analyzing confidence decay patterns.

## Use Cases

### 1. Audit and Compliance
Track who made changes, when, and why for compliance and audit purposes.

### 2. Debugging
Understand when a bug was introduced by examining historical states.

### 3. Analysis
Analyze how skills, relationships, and confidence have evolved over time.

### 4. Recovery
Restore previous states if recent changes caused issues.

### 5. Machine Learning
Use historical data to train models on relationship evolution.

## Getting Started

### Prerequisites

- nexs-mcp v1.2.0 or later
- MCP client (Claude Desktop, VS Code, etc.)
- Basic understanding of MCP tools

### Quick Start

1. **View Element History**

```json
// Request
{
  "tool": "get_element_history",
  "arguments": {
    "element_id": "skill-python"
  }
}

// Response shows all versions
{
  "element_id": "skill-python",
  "history": [
    {
      "version": 1,
      "timestamp": "2024-06-15T10:30:00Z",
      "author": "alice@example.com",
      "change_type": "create",
      "element_data": {
        "name": "Python",
        "level": 1,
        "category": "Programming"
      }
    },
    {
      "version": 2,
      "timestamp": "2024-09-20T15:45:00Z",
      "author": "bob@example.com",
      "change_type": "update",
      "element_data": {
        "name": "Python",
        "level": 3,
        "category": "Programming"
      },
      "changes": {
        "level": 3
      }
    }
  ],
  "total": 2
}
```

2. **Time Travel to Specific Date**

```json
// Request - see graph state on June 15, 2024
{
  "tool": "get_graph_at_time",
  "arguments": {
    "target_time": "2024-06-15T14:30:00Z"
  }
}

// Response shows graph as it was on that date
{
  "timestamp": "2024-06-15T14:30:00Z",
  "elements": {
    "skill-python": {
      "name": "Python",
      "level": 1,
      "category": "Programming"
    }
  },
  "relationships": {},
  "element_count": 1,
  "relationship_count": 0
}
```

3. **Analyze Confidence Decay**

```json
// Request - show relationships with decayed confidence
{
  "tool": "get_decayed_graph",
  "arguments": {
    "confidence_threshold": 0.6
  }
}

// Response shows current state with decay applied
{
  "timestamp": "2024-12-24T10:00:00Z",
  "relationships": {
    "rel-skill-persona": {
      "from": "skill-python",
      "to": "persona-developer",
      "original_confidence": 0.95,
      "decayed_confidence": 0.73,
      "created": "2024-03-15T10:00:00Z",
      "last_accessed": "2024-11-10T08:30:00Z"
    }
  },
  "confidence_threshold": 0.6,
  "total_relationships": 10,
  "filtered_out": 8
}
```

## Common Workflows

### Workflow 1: Tracking Skill Development

**Goal**: See how a developer's skill level progressed over time.

```json
// Step 1: Get skill history
{
  "tool": "get_element_history",
  "arguments": {
    "element_id": "skill-python-alice",
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-12-31T23:59:59Z"
  }
}

// Step 2: Analyze progression
// Look at level field in each version:
// - Version 1 (Jan): level 1
// - Version 2 (Mar): level 2
// - Version 3 (Jun): level 3
// - Version 4 (Sep): level 4
// - Version 5 (Dec): level 5

// Insight: Steady progression, ~1 level per quarter
```

### Workflow 2: Investigating Relationship Changes

**Goal**: Understand why a relationship confidence decreased.

```json
// Step 1: Get relationship history with decay
{
  "tool": "get_relation_history",
  "arguments": {
    "relationship_id": "rel-skill-project",
    "apply_decay": true
  }
}

// Step 2: Review output
{
  "history": [
    {
      "version": 1,
      "timestamp": "2024-01-15T10:00:00Z",
      "original_confidence": 0.95,
      "decayed_confidence": 0.65  // Significant decay
    }
  ]
}

// Step 3: Check last access
// If last_accessed is 6+ months ago, consider:
// - Relationship may be obsolete
// - Refresh if still relevant
// - Archive if no longer needed
```

### Workflow 3: Comparing States

**Goal**: Compare graph state between two dates.

```json
// Step 1: Get state at date 1
{
  "tool": "get_graph_at_time",
  "arguments": {
    "target_time": "2024-06-01T00:00:00Z"
  }
}
// Save result as state_june

// Step 2: Get state at date 2
{
  "tool": "get_graph_at_time",
  "arguments": {
    "target_time": "2024-12-01T00:00:00Z"
  }
}
// Save result as state_december

// Step 3: Compare
// - New elements: state_december.elements NOT IN state_june.elements
// - Removed elements: state_june.elements NOT IN state_december.elements
// - New relationships: Compare relationship counts
// - Confidence changes: Compare relationship confidence values
```

### Workflow 4: Confidence Maintenance

**Goal**: Identify relationships that need attention due to decay.

```json
// Step 1: Get highly decayed relationships
{
  "tool": "get_decayed_graph",
  "arguments": {
    "confidence_threshold": 0.3
  }
}

// Step 2: Review filtered_out count
{
  "total_relationships": 50,
  "filtered_out": 35,  // 70% below threshold!
  "confidence_threshold": 0.3
}

// Step 3: Take action
// For each relationship with low confidence:
// - Verify if still relevant
// - Update/reinforce if active
// - Archive if obsolete
// - Document decision
```

## Advanced Features

### Time Range Queries

Limit history to specific time periods:

```json
{
  "tool": "get_element_history",
  "arguments": {
    "element_id": "skill-python",
    "start_time": "2024-06-01T00:00:00Z",
    "end_time": "2024-09-30T23:59:59Z"
  }
}
```

**Benefits:**
- Faster queries
- Focused analysis
- Reduced data transfer

### Confidence Decay Strategies

Different decay functions suit different scenarios:

1. **Exponential (Default)**
```
Fast initial decay, slower over time
Best for: General relationships
Example: skill-to-project links
```

2. **Linear**
```
Constant decay rate
Best for: Time-sensitive data
Example: temporary collaborations
```

3. **Logarithmic**
```
Slow initial decay, faster later
Best for: Core competencies
Example: fundamental skills
```

4. **Step-based**
```
Discrete confidence levels
Best for: Categorical confidence
Example: certification levels
```

### Reinforcement Learning

Relationships gain confidence through use:

- **Access tracking**: Each query reinforces confidence
- **Bonus per access**: +10% by default (configurable)
- **Maximum boost**: +30% total reinforcement
- **Automatic**: No manual intervention needed

**Example:**
```
Initial confidence: 0.70
After 6 months: 0.50 (decay)
But with 5 accesses: 0.50 + (5 × 0.10) = 0.70 (maintained!)
```

### Critical Relationship Preservation

Protect important relationships from decay:

**Configuration:**
- `PreserveCritical`: true
- `CriticalThreshold`: 0.9 (default)

**Effect:**
Relationships with confidence ≥ 0.9 don't decay.

**Use cases:**
- Core competencies
- Primary relationships
- Certified skills
- Permanent assignments

## Interpreting Results

### History Entries

```json
{
  "version": 2,
  "timestamp": "2024-09-20T15:45:00Z",
  "author": "bob@example.com",
  "change_type": "update",
  "element_data": {...},    // Complete data at this version
  "changes": {"level": 3}   // Only changed fields
}
```

**Fields:**
- `version`: Sequential number, starts at 1
- `timestamp`: When change occurred (UTC)
- `author`: Who made the change
- `change_type`: create, update, activate, deactivate, major
- `element_data`: Full state at this version
- `changes`: Delta from previous version (null for full snapshots)

### Decay Metrics

```json
{
  "original_confidence": 0.95,      // Initial value
  "decayed_confidence": 0.73,       // After time decay
  "created": "2024-03-15T10:00:00Z",
  "last_accessed": "2024-11-10T08:30:00Z",
  "access_count": 12                // Times accessed
}
```

**Interpretation:**
- **23% decay** (0.95 → 0.73) over 9 months
- **Last accessed** 1.5 months ago (relatively recent)
- **12 accesses** suggests active use
- **Verdict**: Healthy relationship, still relevant

### Graph Snapshots

```json
{
  "timestamp": "2024-12-24T10:00:00Z",
  "element_count": 150,
  "relationship_count": 45,
  "total_relationships": 250,
  "filtered_out": 205
}
```

**Metrics:**
- **Element count**: Total elements at this time
- **Relationship count**: Relationships above threshold
- **Total relationships**: All relationships (before filtering)
- **Filtered out**: Relationships below threshold (82% in this case)

## Tips and Best Practices

### 1. Regular Audits

Schedule monthly/quarterly reviews:
- Check decay statistics
- Identify neglected relationships
- Reinforce active connections
- Archive obsolete data

### 2. Meaningful Change Messages

When recording changes, use descriptive messages:
- ✅ "Upgraded to senior level after Q3 review"
- ❌ "update"

### 3. Author Attribution

Always record who made changes:
- Accountability
- Contact for questions
- Historical context

### 4. Time Zone Awareness

All timestamps are in UTC. Convert to local time when:
- Displaying to users
- Comparing with local events
- Generating reports

### 5. Performance Optimization

For large datasets:
- Use time ranges to limit queries
- Cache frequently accessed snapshots
- Batch operations when possible
- Monitor query performance

### 6. Confidence Threshold Selection

Choose thresholds based on use case:
- **0.3-0.4**: Keep most relationships
- **0.5-0.6**: Moderate filtering
- **0.7-0.8**: Only high-confidence
- **0.9+**: Critical relationships only

## Troubleshooting

### "No history found for element"

**Cause**: Element hasn't been tracked yet.

**Solution**: History starts from first recorded change.

### "No snapshot found at time"

**Cause**: Querying before element was created.

**Solution**: Use later timestamp or check creation time.

### "Too many versions"

**Cause**: History growing too large.

**Solution**: Configure retention policy to limit versions.

### "Decay calculation error"

**Cause**: Invalid confidence or timestamp.

**Solution**: Verify confidence is 0.0-1.0 and timestamp is valid.

## Examples by Role

### For Developers

**Track bug introduction:**
```json
// Find when bug was introduced
{
  "tool": "get_element_history",
  "arguments": {
    "element_id": "component-auth",
    "start_time": "2024-11-01T00:00:00Z"
  }
}
// Review changes until bug appears
```

### For Project Managers

**Track team skill evolution:**
```json
// See skill progression
{
  "tool": "get_graph_at_time",
  "arguments": {
    "target_time": "2024-01-01T00:00:00Z"
  }
}
// Compare with current state
```

### For Data Scientists

**Analyze relationship patterns:**
```json
// Get all relationships with decay
{
  "tool": "get_decayed_graph",
  "arguments": {
    "confidence_threshold": 0.0
  }
}
// Analyze decay patterns, predict future states
```

### For HR/Talent Management

**Audit certifications and skills:**
```json
// Check when skills were added/verified
{
  "tool": "get_element_history",
  "arguments": {
    "element_id": "skill-aws-certified",
    "author": "hr@example.com"
  }
}
```

## Next Steps

- [API Reference](../api/TEMPORAL_FEATURES.md) - Detailed API documentation
- [Code Examples](../../examples/workflows/temporal_examples.sh) - Ready-to-use scripts
- [Architecture](../architecture/TEMPORAL.md) - System design details

## Support

For questions or issues:
- GitHub Issues: [nexs-mcp/issues](https://github.com/fsvxavier/nexs-mcp/issues)
- Documentation: [docs/](https://github.com/fsvxavier/nexs-mcp/tree/main/docs)
