# Memory Element

## Overview

Memories store persistent context with automatic deduplication and search indexing.

## Key Features

- Text-based YAML storage
- Date-based organization (YYYY-MM-DD)
- SHA-256 content hashing for deduplication
- Search indexing support
- Custom metadata

## Examples

### Meeting Notes
```json
{
  "name": "Team Standup 2025-12-18",
  "version": "1.0.0",
  "author": "team-lead",
  "content": "Daily standup meeting:\n\n**Completed:**\n- Implemented 6 element types\n- Created MCP handlers\n\n**In Progress:**\n- Documentation\n\n**Blockers:**\n- None",
  "date_created": "2025-12-18",
  "metadata": {
    "meeting_type": "standup",
    "attendees": "5",
    "duration": "15min"
  },
  "search_index": ["standup", "elements", "mcp", "documentation"]
}
```

### Project Decision
```json
{
  "name": "Architecture Decision: Use Go SDK",
  "version": "1.0.0",
  "author": "architect",
  "content": "Decision: Use official MCP Go SDK instead of building from scratch.\n\nRationale:\n- Maintained by MCP team\n- Better type safety\n- Community support\n\nTrade-offs:\n- Less control over internals\n- SDK update dependencies",
  "date_created": "2025-12-15",
  "metadata": {
    "decision_type": "architecture",
    "status": "approved",
    "impact": "high"
  }
}
```

## Auto-deduplication
The system automatically computes SHA-256 hashes to prevent duplicate memories with identical content.

## Usage
```javascript
{
  "tool": "create_memory",
  "arguments": {
    "name": "Research Finding",
    "content": "Key insight from research...",
    "version": "1.0.0",
    "author": "researcher"
  }
}
// Hash automatically computed on creation
```
