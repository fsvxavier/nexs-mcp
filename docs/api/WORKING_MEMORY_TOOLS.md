# Working Memory Tools API

The Working Memory system provides session-scoped, time-to-live based memory management with automatic promotion to long-term storage. This document describes the 15 MCP tools available for managing working memory.

## Overview

Working memory operates with:
- **Session Scoping**: Each session has isolated memory
- **Priority-Based TTL**: 
  - Low: 1 hour
  - Medium: 4 hours
  - High: 12 hours
  - Critical: 24 hours
- **Auto-Promotion**: Based on access count, importance score, priority, and age
- **Background Cleanup**: Expired memories cleaned every 5 minutes

## Tools

### 1. working_memory_add

Creates a new working memory entry in the specified session.

**Input:**
```json
{
  "session_id": "user-session-123",
  "content": "Meeting notes from today's standup",
  "priority": "high",
  "tags": ["meeting", "standup", "team"],
  "metadata": {
    "meeting_type": "standup",
    "duration": "15min"
  }
}
```

**Parameters:**
- `session_id` (required): Unique session identifier
- `content` (required): Memory content text (min 1 char, max 100KB)
- `priority` (required): One of: "low", "medium", "high", "critical"
- `tags` (optional): Array of tag strings
- `metadata` (optional): Key-value pairs for additional context

**Returns:**
- `id`: Generated memory ID
- `session_id`: Session identifier
- `content`: Memory content
- `priority`: Priority level
- `expires_at`: ISO 8601 timestamp when memory expires
- `created_at`: ISO 8601 timestamp of creation
- `access_count`: Number of accesses (starts at 1)
- `importance_score`: Calculated importance (0.0-1.0)
- `tags`: Array of tags
- `metadata`: Metadata map

**Example Response:**
```json
{
  "id": "working_memory_user-session-123-Meeting notes fro_20241222-150000",
  "session_id": "user-session-123",
  "content": "Meeting notes from today's standup",
  "priority": "high",
  "expires_at": "2024-12-23T03:00:00Z",
  "created_at": "2024-12-22T15:00:00Z",
  "access_count": 1,
  "importance_score": 0.45,
  "tags": ["meeting", "standup", "team"],
  "metadata": {
    "meeting_type": "standup",
    "duration": "15min"
  }
}
```

---

### 2. working_memory_get

Retrieves a specific working memory by ID and records the access.

**Input:**
```json
{
  "session_id": "user-session-123",
  "memory_id": "working_memory_user-session-123-Meeting notes fro_20241222-150000"
}
```

**Parameters:**
- `session_id` (required): Session identifier
- `memory_id` (required): Memory identifier

**Returns:** Same structure as `working_memory_add` with updated `access_count`

**Behavior:**
- Increments `access_count`
- Updates `importance_score`
- Triggers auto-promotion check
- Returns error if expired

---

### 3. working_memory_list

Lists all working memories in a session with optional filters.

**Input:**
```json
{
  "session_id": "user-session-123",
  "include_expired": false,
  "include_promoted": false
}
```

**Parameters:**
- `session_id` (required): Session identifier
- `include_expired` (optional, default: false): Include expired memories
- `include_promoted` (optional, default: false): Include promoted memories

**Returns:** Array of memory objects

**Example Response:**
```json
{
  "memories": [
    {
      "id": "working_memory_...",
      "session_id": "user-session-123",
      "content": "Active memory 1",
      "priority": "medium",
      "access_count": 3,
      "expires_at": "2024-12-23T03:00:00Z"
    }
  ],
  "count": 1
}
```

---

### 4. working_memory_promote

Manually promotes a working memory to long-term storage.

**Input:**
```json
{
  "session_id": "user-session-123",
  "memory_id": "working_memory_user-session-123-Important info_20241222-150000"
}
```

**Parameters:**
- `session_id` (required): Session identifier
- `memory_id` (required): Memory identifier to promote

**Returns:** Long-term Memory object

**Example Response:**
```json
{
  "id": "memory_promoted_working_memory_user-session-123-Important info_20241222-150000_20241222-160000",
  "content": "Important info that needs long-term storage",
  "tags": ["important", "permanent"],
  "metadata": {
    "promoted_from_working": "true",
    "working_memory_id": "working_memory_user-session-123-Important info_20241222-150000"
  }
}
```

**Behavior:**
- Creates permanent Memory in repository
- Marks working memory as promoted
- Preserves all tags and metadata
- Returns existing if already promoted

---

### 5. working_memory_clear_session

Clears all working memories in a session.

**Input:**
```json
{
  "session_id": "user-session-123"
}
```

**Parameters:**
- `session_id` (required): Session identifier

**Returns:**
```json
{
  "success": true,
  "message": "Session cleared: user-session-123"
}
```

**Behavior:**
- Removes all memories from session
- Does not affect long-term promoted memories
- Session can be reused immediately

---

### 6. working_memory_stats

Gets statistics for a session's working memory.

**Input:**
```json
{
  "session_id": "user-session-123"
}
```

**Parameters:**
- `session_id` (required): Session identifier

**Returns:**
```json
{
  "session_id": "user-session-123",
  "total_count": 10,
  "active_count": 7,
  "expired_count": 2,
  "promoted_count": 3,
  "pending_promotion": 1,
  "avg_access_count": 4.2,
  "avg_importance": 0.65,
  "by_priority": {
    "low": 2,
    "medium": 5,
    "high": 2,
    "critical": 1
  }
}
```

**Fields:**
- `total_count`: Total memories in session
- `active_count`: Non-expired memories
- `expired_count`: Expired memories (awaiting cleanup)
- `promoted_count`: Memories promoted to long-term
- `pending_promotion`: Memories eligible for auto-promotion
- `avg_access_count`: Average access count across all memories
- `avg_importance`: Average importance score
- `by_priority`: Count by priority level

---

### 7. working_memory_expire

Manually expires a working memory before its TTL.

**Input:**
```json
{
  "session_id": "user-session-123",
  "memory_id": "working_memory_user-session-123-Old info_20241222-100000"
}
```

**Parameters:**
- `session_id` (required): Session identifier
- `memory_id` (required): Memory identifier

**Returns:**
```json
{
  "success": true,
  "message": "Memory expired: working_memory_..."
}
```

**Behavior:**
- Sets `expires_at` to current time
- Memory will be removed by next cleanup cycle
- Can still be retrieved until cleanup runs

---

### 8. working_memory_extend_ttl

Extends the TTL of a working memory by its priority duration.

**Input:**
```json
{
  "session_id": "user-session-123",
  "memory_id": "working_memory_user-session-123-Keep this_20241222-100000"
}
```

**Parameters:**
- `session_id` (required): Session identifier
- `memory_id` (required): Memory identifier

**Returns:**
```json
{
  "success": true,
  "new_expires_at": "2024-12-23T05:00:00Z",
  "message": "TTL extended for memory: working_memory_..."
}
```

**Behavior:**
- Adds priority TTL duration to `expires_at`
- High priority (12h) extends by 12 more hours
- Can be called multiple times
- Does not affect promotion status

---

### 9. working_memory_export

Exports all working memories from a session.

**Input:**
```json
{
  "session_id": "user-session-123"
}
```

**Parameters:**
- `session_id` (required): Session identifier

**Returns:**
```json
{
  "memories": [
    {
      "id": "working_memory_...",
      "content": "Exported memory 1",
      "priority": "high",
      "created_at": "2024-12-22T15:00:00Z",
      "expires_at": "2024-12-23T03:00:00Z",
      "access_count": 5,
      "importance_score": 0.72,
      "tags": ["export", "important"],
      "metadata": {"key": "value"}
    }
  ],
  "count": 1,
  "session_id": "user-session-123"
}
```

**Use Cases:**
- Session backup before clearing
- Migration between sessions
- Analysis and reporting
- Long-term archival

---

### 10. working_memory_list_pending

Lists memories eligible for auto-promotion.

**Input:**
```json
{
  "session_id": "user-session-123"
}
```

**Parameters:**
- `session_id` (required): Session identifier

**Returns:**
```json
{
  "memories": [
    {
      "id": "working_memory_...",
      "content": "Highly accessed memory",
      "access_count": 15,
      "importance_score": 0.85,
      "priority": "high",
      "should_promote": true
    }
  ],
  "count": 1
}
```

**Promotion Rules:**
1. Access count >= priority threshold (low:10, medium:15, high:20, critical:10)
2. OR importance score >= 0.8
3. OR critical priority AND accessed at least once
4. OR age > 6 hours AND access count >= 5

---

### 11. working_memory_list_expired

Lists expired memories awaiting cleanup.

**Input:**
```json
{
  "session_id": "user-session-123"
}
```

**Parameters:**
- `session_id` (required): Session identifier

**Returns:**
```json
{
  "memories": [
    {
      "id": "working_memory_...",
      "content": "Expired memory",
      "expires_at": "2024-12-22T10:00:00Z",
      "priority": "low"
    }
  ],
  "count": 1
}
```

---

### 12. working_memory_list_promoted

Lists memories that have been promoted to long-term storage.

**Input:**
```json
{
  "session_id": "user-session-123"
}
```

**Parameters:**
- `session_id` (required): Session identifier

**Returns:**
```json
{
  "memories": [
    {
      "id": "working_memory_...",
      "content": "Promoted memory",
      "promoted_to_id": "memory_promoted_...",
      "promoted_at": "2024-12-22T16:00:00Z",
      "access_count": 25
    }
  ],
  "count": 1
}
```

---

### 13. working_memory_bulk_promote

Promotes multiple memories in a single operation.

**Input:**
```json
{
  "session_id": "user-session-123",
  "memory_ids": [
    "working_memory_1",
    "working_memory_2",
    "working_memory_3"
  ]
}
```

**Parameters:**
- `session_id` (required): Session identifier
- `memory_ids` (required): Array of memory IDs to promote

**Returns:**
```json
{
  "success": 2,
  "failed": 1,
  "results": [
    {
      "memory_id": "working_memory_1",
      "success": true,
      "long_term_id": "memory_promoted_..."
    },
    {
      "memory_id": "working_memory_2",
      "success": true,
      "long_term_id": "memory_promoted_..."
    },
    {
      "memory_id": "working_memory_3",
      "success": false,
      "error": "memory not found"
    }
  ]
}
```

**Behavior:**
- Continues on individual failures
- Returns detailed results per memory
- Transactional per memory (not per batch)

---

### 14. working_memory_relation_add

Adds a relation between a working memory and another element.

**Input:**
```json
{
  "session_id": "user-session-123",
  "memory_id": "working_memory_...",
  "relation_type": "references",
  "target_element_id": "skill_python_expert_20241222-100000",
  "metadata": {
    "context": "Used in code generation",
    "confidence": "high"
  }
}
```

**Parameters:**
- `session_id` (required): Session identifier
- `memory_id` (required): Working memory ID
- `relation_type` (required): Type of relation (e.g., "references", "depends_on", "related_to")
- `target_element_id` (required): ID of target element
- `metadata` (optional): Additional relation metadata

**Returns:**
```json
{
  "success": true,
  "message": "Relation added successfully"
}
```

**Note:** Relations are stored in working memory metadata and preserved during promotion.

---

### 15. working_memory_search

Searches working memories by content, tags, or metadata.

**Input:**
```json
{
  "session_id": "user-session-123",
  "query": "meeting notes",
  "tags": ["standup"],
  "priority": "high",
  "include_expired": false,
  "include_promoted": true
}
```

**Parameters:**
- `session_id` (required): Session identifier
- `query` (optional): Text to search in content
- `tags` (optional): Filter by tags (matches any)
- `priority` (optional): Filter by priority
- `include_expired` (optional, default: false): Include expired
- `include_promoted` (optional, default: false): Include promoted

**Returns:**
```json
{
  "memories": [
    {
      "id": "working_memory_...",
      "content": "Meeting notes from standup",
      "relevance_score": 0.92,
      "tags": ["meeting", "standup"],
      "priority": "high"
    }
  ],
  "count": 1,
  "query": "meeting notes"
}
```

**Search Behavior:**
- Case-insensitive content matching
- Tag exact match (case-sensitive)
- Priority exact match
- Results sorted by relevance score (0.0-1.0)
- Relevance based on: query match position, access count, importance

---

## Auto-Promotion System

Working memories are automatically promoted to long-term storage when they meet promotion criteria.

### Promotion Rules

A memory is eligible for promotion if ANY of these conditions are met:

1. **High Access Count**
   - Low priority: 10+ accesses
   - Medium priority: 15+ accesses
   - High priority: 20+ accesses
   - Critical priority: 10+ accesses

2. **High Importance**
   - Importance score >= 0.8

3. **Critical and Accessed**
   - Priority is "critical"
   - AND access_count >= 1

4. **Aged with Activity**
   - Age > 6 hours
   - AND access_count >= 5

### Importance Calculation

Importance score (0.0-1.0) is calculated using:

```
importance = (0.4 × normalized_access) + 
             (0.3 × time_decay) + 
             (0.2 × priority_weight) + 
             (0.1 × content_length)
```

Where:
- `normalized_access` = min(access_count / 10, 1.0)
- `time_decay` = 1.0 - (age_minutes / ttl_minutes)
- `priority_weight` = {low: 0.25, medium: 0.5, high: 0.75, critical: 1.0}
- `content_length` = min(length / 1000, 1.0)

### Promotion Behavior

When promoted:
- Original working memory marked as `promoted`
- New long-term Memory created with ID: `memory_promoted_{working_id}_{timestamp}`
- All tags and metadata preserved
- `promoted_from_working` metadata added
- Working memory remains accessible until expired

---

## Usage Patterns

### Pattern 1: Session-Based Conversation Memory

```javascript
// Start session with context
await working_memory_add({
  session_id: "chat-123",
  content: "User prefers Python examples",
  priority: "high",
  tags: ["preference", "language"]
});

// During conversation
const memory = await working_memory_get({
  session_id: "chat-123",
  memory_id: "..."
});

// End session
await working_memory_clear_session({
  session_id: "chat-123"
});
```

### Pattern 2: Important Information Preservation

```javascript
// Create critical information
const criticalMem = await working_memory_add({
  session_id: "user-456",
  content: "User's API key for service X",
  priority: "critical",
  tags: ["credentials", "important"]
});

// Manually promote immediately
await working_memory_promote({
  session_id: "user-456",
  memory_id: criticalMem.id
});
```

### Pattern 3: Periodic Session Cleanup

```javascript
// Check what will be cleaned
const expired = await working_memory_list_expired({
  session_id: "session-789"
});

// Extend important ones
for (const mem of expired.memories) {
  if (mem.tags.includes("keep")) {
    await working_memory_extend_ttl({
      session_id: "session-789",
      memory_id: mem.id
    });
  }
}
```

### Pattern 4: Bulk Promotion Before Session End

```javascript
// Get pending promotions
const pending = await working_memory_list_pending({
  session_id: "session-999"
});

// Promote all eligible
await working_memory_bulk_promote({
  session_id: "session-999",
  memory_ids: pending.memories.map(m => m.id)
});
```

---

## Error Handling

All tools return errors in this format:

```json
{
  "error": "Error message description",
  "code": "ERROR_CODE",
  "details": {
    "session_id": "...",
    "memory_id": "..."
  }
}
```

Common Error Codes:
- `VALIDATION_ERROR`: Invalid input parameters
- `NOT_FOUND`: Session or memory not found
- `EXPIRED`: Memory has expired
- `ALREADY_PROMOTED`: Memory already promoted
- `STORAGE_ERROR`: Failed to save to repository

---

## Performance Considerations

### Memory Limits
- Maximum content size: 100 KB per memory
- Recommended session size: < 1000 memories
- Background cleanup: Every 5 minutes
- Auto-promotion check: On every Get() call

### Best Practices
1. Use appropriate priorities (don't over-use "critical")
2. Clear sessions when conversations end
3. Promote important information before clearing
4. Use tags for efficient filtering
5. Extend TTL instead of recreating memories
6. Use bulk operations for efficiency

### Monitoring
Check session health with `working_memory_stats`:
- High `pending_promotion`: Consider bulk promotion
- High `expired_count`: Cleanup may be delayed
- High `avg_access_count`: Good memory reuse
- Low `avg_importance`: Consider priority adjustments

---

## Integration with Long-Term Memory

Working memory serves as a staging area before long-term storage:

1. **Short-term context**: Session-specific information (TTL-based)
2. **Auto-promotion**: Frequently accessed → Long-term
3. **Manual promotion**: Important → Permanent
4. **Metadata preservation**: All context carries forward

**Promoted Memory ID Format:**
```
memory_promoted_{original_working_memory_id}_{promotion_timestamp}
```

**Metadata Added on Promotion:**
```json
{
  "promoted_from_working": "true",
  "working_memory_id": "original_id",
  "working_memory_session": "session_id",
  "promotion_reason": "auto|manual",
  "access_count_at_promotion": "15",
  "importance_at_promotion": "0.85"
}
```

---

## See Also

- [MCP Tools API](MCP_TOOLS.md) - Complete MCP tools documentation
- [Getting Started](../user-guide/GETTING_STARTED.md) - Basic usage guide
- [Architecture: Application Layer](../architecture/APPLICATION.md) - Working Memory service details
- [Domain Model](../architecture/DOMAIN.md) - Working Memory entity design
