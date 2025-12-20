# Auto-Save Feature

## Overview

The auto-save feature automatically saves conversation context as memories, ensuring continuity across sessions and preventing loss of important context.

## Configuration

### Default Behavior

Auto-save is **enabled by default** with the following settings:
- Auto-save enabled: `true`
- Auto-save interval: `5 minutes`

### Command-Line Flags

```bash
# Disable auto-save
./bin/nexs-mcp --auto-save-memories=false

# Change auto-save interval to 10 minutes
./bin/nexs-mcp --auto-save-interval=10m

# Combine settings
./bin/nexs-mcp --auto-save-memories=true --auto-save-interval=3m
```

### Environment Variables

```bash
# Disable auto-save
export NEXS_AUTO_SAVE_MEMORIES=false

# Set custom interval (Go duration format)
export NEXS_AUTO_SAVE_INTERVAL=10m

# Start server
./bin/nexs-mcp
```

## MCP Tool: `save_conversation_context`

The `save_conversation_context` tool allows manual or automatic saving of conversation context as memories.

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `context` | string | Yes | The conversation context to save (min 10 characters) |
| `summary` | string | No | Brief summary of the context (used as memory name) |
| `tags` | array[string] | No | Tags for categorization (auto-adds "auto-save", "conversation") |
| `importance` | string | No | Importance level: "low", "medium", "high", "critical" |
| `related_to` | array[string] | No | IDs of related elements |

### Output

```json
{
  "memory_id": "memory-abc123",
  "saved": true,
  "message": "Conversation context saved successfully"
}
```

### Examples

#### Basic Usage

```json
{
  "tool": "save_conversation_context",
  "input": {
    "context": "User asked about creating personas. Agent explained the process and demonstrated with examples.",
    "summary": "Persona creation tutorial"
  }
}
```

Result:
- Memory created with auto-generated ID
- Tags: `["auto-save", "conversation"]`
- Name: "Persona creation tutorial"

#### With Importance Level

```json
{
  "tool": "save_conversation_context",
  "input": {
    "context": "Critical bug discovered in memory persistence. Root cause: using SimpleElement instead of typed elements. Solution implemented and tested.",
    "summary": "Memory persistence bug fix",
    "importance": "critical",
    "tags": ["bug", "fix", "persistence"]
  }
}
```

Result:
- Tags: `["bug", "fix", "persistence", "importance:critical", "auto-save", "conversation"]`
- Metadata includes `"importance": "critical"`

#### With Related Elements

```json
{
  "tool": "save_conversation_context",
  "input": {
    "context": "Discussed ensemble configuration based on previous agent setup. User requested example YAML.",
    "summary": "Ensemble configuration discussion",
    "related_to": ["agent-web-scraper", "ensemble-data-pipeline"],
    "tags": ["ensemble", "configuration"]
  }
}
```

Result:
- Metadata includes `"related_to": "agent-web-scraper,ensemble-data-pipeline"`
- Links conversation to specific elements

## How It Works

### Storage Location

Memories are saved in the directory structure:
```
data/elements/memory/YYYY-MM-DD/memory-<id>.yaml
```

Example:
```
data/elements/memory/2025-12-20/memory-abc123.yaml
```

### Automatic Context Preservation

When auto-save is enabled:
1. The AI can call `save_conversation_context` at appropriate times
2. Context is saved with relevant metadata and tags
3. Memories include keyword extraction for search indexing
4. Related elements are linked automatically

### Memory Structure

```yaml
id: memory-abc123
type: memory
name: "Conversation Context - 2025-12-20 10:30"
description: "Brief summary of the context"
version: 1.0.0
author: auto-save
created_at: "2025-12-20T10:30:00Z"
updated_at: "2025-12-20T10:30:00Z"
is_active: true
tags:
  - auto-save
  - conversation
  - importance:high
metadata:
  auto_saved: "true"
  saved_at: "2025-12-20T10:30:00Z"
  importance: high
  related_to: "agent-123,persona-456"
data:
  content: |
    Full conversation context here...
    Multiple lines preserved...
  search_index:
    - keyword1
    - keyword2
    - keyword3
  hash: sha256:abc123...
```

## Keyword Extraction

The tool automatically extracts keywords from the context for search indexing:

- Filters out common stop words (English and Portuguese)
- Excludes words shorter than 3 characters
- Ranks by frequency
- Returns top 10 keywords by default

This enables semantic search via:
```json
{
  "tool": "search_memory",
  "input": {
    "query": "persona creation"
  }
}
```

## Best Practices

### When to Save Context

- After completing a major task or feature
- Before context switching to a different topic
- After bug investigations or problem-solving
- When documenting important decisions
- After tutorial/teaching sessions

### Importance Levels

- **`low`**: General conversation, routine questions
- **`medium`**: Feature discussions, minor issues (default if omitted)
- **`high`**: Important decisions, significant features
- **`critical`**: Security issues, data loss, breaking changes

### Effective Summaries

Good summaries are:
- Concise (< 100 characters)
- Descriptive of the main topic
- Searchable (include key terms)

Examples:
- ✅ "Memory persistence bug fix - root cause analysis"
- ✅ "Persona creation tutorial with examples"
- ❌ "Discussion"
- ❌ "Context"

### Using Tags

Effective tagging helps future searches:
- **Type**: `tutorial`, `bug`, `feature`, `documentation`
- **Domain**: `persistence`, `mcp`, `configuration`, `testing`
- **Status**: `completed`, `in-progress`, `blocked`

Example:
```json
{
  "tags": ["bug", "persistence", "critical", "completed"]
}
```

## Searching Saved Context

Use memory search tools to find saved context:

```json
{
  "tool": "search_memory",
  "input": {
    "query": "persistence bug",
    "filters": {
      "tags": ["bug", "critical"]
    }
  }
}
```

Or list by date:
```json
{
  "tool": "list_elements",
  "input": {
    "type": "memory",
    "created_after": "2025-12-20T00:00:00Z"
  }
}
```

## Troubleshooting

### Auto-save not working

Check configuration:
```bash
./bin/nexs-mcp --help | grep auto
```

Verify settings:
```bash
# Should show auto-save-memories=true
echo $NEXS_AUTO_SAVE_MEMORIES
```

### Context not being saved

Ensure:
1. Context is at least 10 characters
2. Auto-save is enabled in config
3. Permissions allow writing to `data/elements/memory/`

### Finding old context

Use date-based search:
```bash
ls data/elements/memory/2025-12-*/
```

Or search by content:
```json
{
  "tool": "search_memory",
  "input": {
    "query": "your search terms",
    "date_from": "2025-12-01"
  }
}
```

## VS Code Integration

When using with GitHub Copilot Chat in VS Code:

1. Auto-save will preserve context automatically
2. Use natural language to trigger saves:
   - "Save this conversation for later"
   - "Remember this bug fix process"
   - "Store this configuration example"

3. Retrieve context later:
   - "What did we discuss about personas yesterday?"
   - "Show me the bug fix from earlier"
   - "Find conversations about configuration"

## Performance Considerations

- **Interval**: Minimum 5 minutes recommended to avoid excessive I/O
- **Storage**: Memories use ~1-5KB per conversation context
- **Indexing**: Keyword extraction is lightweight (< 1ms per save)
- **Search**: File-based search is fast for < 10,000 memories

## Future Enhancements

Planned features:
- [ ] Automatic context detection (AI decides when to save)
- [ ] Context compression for long conversations
- [ ] Integration with vector databases for semantic search
- [ ] Context summaries across multiple sessions
- [ ] Automatic tagging based on content analysis
