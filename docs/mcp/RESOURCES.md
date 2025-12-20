# MCP Resources Protocol

**Status:** ✅ Implemented (M0.7)  
**Default:** Disabled (opt-in for security)  
**Version:** NEXS-MCP 0.1.0+

## Overview

NEXS-MCP implements the MCP Resources Protocol to expose capability index data as readable resources. This enables AI assistants to efficiently access project context without requiring multiple tool calls.

**Key Features:**
- 3 resource variants (summary, full, stats)
- Configurable caching with TTL
- Whitelist-based exposure control
- Zero overhead when disabled
- Thread-safe concurrent access

## Resource Variants

### 1. Summary Resource (`capability-index://summary`)

**Size:** ~3K tokens  
**Format:** Markdown  
**Purpose:** Quick overview for AI context

**Contents:**
- Element counts by type (Persona, Skill, Template, Agent, Ensemble, Memory)
- Index statistics (documents, vocabulary size, avg terms/doc)
- Top keywords (most significant terms)
- Recent elements (last 7 days, max 10)
- Available resources list

**Use Case:** Initial context loading, quick capability assessment

### 2. Full Resource (`capability-index://full`)

**Size:** ~40K tokens  
**Format:** Markdown  
**Purpose:** Complete capability index details

**Contents:**
- Comprehensive index statistics
- Element distribution by type
- All elements with full metadata (ID, version, created/updated dates, tags)
- Vocabulary breakdown (top 100 terms)
- Relationship graph (agent goals and actions)

**Use Case:** Deep analysis, detailed planning, comprehensive context

### 3. Stats Resource (`capability-index://stats`)

**Size:** <1K tokens  
**Format:** JSON  
**Purpose:** Machine-readable statistics

**Contents:**
```json
{
  "generated_at": "2025-12-19T10:30:00Z",
  "element_counts": {
    "persona": 5,
    "skill": 12,
    "template": 3,
    "agent": 2,
    "ensemble": 1,
    "memory": 8
  },
  "index_statistics": {
    "document_count": 31,
    "vocabulary_size": 487,
    "avg_document_length": 42,
    "memory_usage_kb": 156
  },
  "cache_statistics": {
    "entries": 3,
    "ttl_seconds": 300
  },
  "resources": {
    "summary": "capability-index://summary",
    "full": "capability-index://full",
    "stats": "capability-index://stats"
  }
}
```

**Use Case:** Programmatic access, monitoring, dashboards

## Configuration

### Command Line Flags

```bash
# Enable all resources
nexs-mcp --resources-enabled=true

# Custom cache TTL
nexs-mcp --resources-enabled=true --resources-cache-ttl=10m
```

### Environment Variables

```bash
# Enable resources
export NEXS_RESOURCES_ENABLED=true

# Expose specific resources only (comma-separated)
export NEXS_RESOURCES_EXPOSE="capability-index://summary,capability-index://stats"

# Set cache TTL (Go duration format)
export NEXS_RESOURCES_CACHE_TTL=5m
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `resources.enabled` | bool | `false` | Enable MCP Resources Protocol |
| `resources.expose` | []string | `[]` | Whitelist of resource URIs to expose (empty = all) |
| `resources.cache_ttl` | duration | `5m` | Cache TTL for resource content |

## Usage Examples

### Claude Desktop

1. **Enable resources:**
   ```bash
   nexs-mcp --resources-enabled=true
   ```

2. **Attach resource manually:**
   In Claude Desktop, use the "Attach Resource" feature to select:
   - `capability-index://summary` for quick context
   - `capability-index://full` for comprehensive analysis

3. **Resource will be cached** for 5 minutes by default

### VS Code with MCP Extension

1. **Configure server** with resources enabled in `.vscode/mcp.json`:
   ```json
   {
     "servers": {
       "nexs-mcp": {
         "command": "nexs-mcp",
         "args": ["--resources-enabled=true"],
         "env": {
           "NEXS_RESOURCES_CACHE_TTL": "10m"
         }
       }
     }
   }
   ```

2. **Discover resources** using MCP client:
   ```javascript
   const resources = await client.listResources();
   // Returns: 3 resources (summary, full, stats)
   ```

3. **Read resource:**
   ```javascript
   const summary = await client.readResource("capability-index://summary");
   console.log(summary.contents[0].text);
   ```

### Selective Exposure (Security Best Practice)

Expose only summary and stats (not full details):

```bash
export NEXS_RESOURCES_ENABLED=true
export NEXS_RESOURCES_EXPOSE="capability-index://summary,capability-index://stats"
nexs-mcp
```

This limits token usage and prevents exposing all element details.

## Caching Behavior

**Cache Key:** Resource URI  
**Cache Entry:** `{ content: string, timestamp: time.Time }`  
**Cache TTL:** Configurable (default: 5 minutes)

**Cache Invalidation:**
- Automatic: After TTL expires
- Manual: Restart server (clears all cache)

**Cache Benefits:**
- Reduces index computation overhead
- Consistent snapshots during TTL window
- Thread-safe concurrent reads

**Cache Trade-offs:**
- May serve stale data (max age = TTL)
- Memory usage grows with cache entries
- No cross-session persistence

## Limitations & Known Issues

### Claude Code (October 2025)
**Status:** ⚠️ Partial Support

Resources are discoverable via `resources/list` but **cannot be read** via `resources/read`. This is a known limitation of the Claude Code client implementation as of October 2025.

**Workaround:** Use Claude Desktop or VS Code with MCP extension for full resource support.

### Token Limits

- **Summary:** ~3K tokens (safe for most contexts)
- **Full:** ~40K tokens (may exceed some model context limits)
- **Stats:** <1K tokens (always safe)

**Recommendation:** Start with summary resource, upgrade to full only when needed.

### Performance

Resource generation is **synchronous** and blocks the MCP call:
- Summary: ~10-50ms (typical)
- Full: ~50-200ms (with many elements)
- Stats: ~5-20ms (fastest)

**Optimization:** Cache significantly reduces subsequent calls to <1ms.

## Security Considerations

### Default Disabled

Resources are **disabled by default** to:
1. Prevent unintended data exposure
2. Minimize attack surface
3. Require explicit opt-in

### Whitelist Control

Use `resources.expose` to limit which resources are available:

```bash
# Only expose stats (machine-readable, no sensitive data)
export NEXS_RESOURCES_EXPOSE="capability-index://stats"
```

### Data Exposure

Resources may contain:
- ✅ Element metadata (names, descriptions, tags)
- ✅ Relationship information (agent goals, actions)
- ⚠️ Full element details (in `full` variant)
- ❌ Element content/implementation (not exposed)

**Best Practice:** Use `summary` for general context, reserve `full` for trusted environments.

## Troubleshooting

### Resources Not Listed

**Problem:** `resources/list` returns empty array

**Solutions:**
1. Check `--resources-enabled=true` flag
2. Verify environment variable: `echo $NEXS_RESOURCES_ENABLED`
3. Check server logs for "MCP Resources Protocol enabled"

### Resources Listed But Not Readable

**Problem:** `resources/list` shows resources, but `resources/read` fails

**Possible Causes:**
1. **Claude Code limitation** (see Known Issues above)
2. **Whitelist mismatch:** Resource URI not in `resources.expose`
3. **Cache issue:** Try restarting server

**Debug:**
```bash
# Enable all resources explicitly
nexs-mcp --resources-enabled=true --resources-cache-ttl=0s
```

### Stale Data

**Problem:** Resource content doesn't reflect recent changes

**Cause:** Cache TTL hasn't expired yet

**Solutions:**
1. Wait for cache expiration (check `cache_ttl`)
2. Restart server (clears all cache)
3. Reduce cache TTL: `--resources-cache-ttl=1m`

## Migration from DollhouseMCP

If migrating from DollhouseMCP, note these differences:

| Feature | DollhouseMCP | NEXS-MCP |
|---------|--------------|----------|
| Resource Types | 3 (summary, full, stats) | ✅ Same |
| Default State | Disabled | ✅ Same |
| Configuration | `resources.enabled`, `resources.expose`, `resources.cache_ttl` | ✅ Same pattern |
| URI Scheme | `dollhouse://...` | `capability-index://...` |
| Format | Markdown + JSON | ✅ Same |

**Key Difference:** URI scheme changed from `dollhouse://` to `capability-index://` to reflect NEXS-MCP's focus.

## Related Documentation

- [MCP Specification - Resources](https://spec.modelcontextprotocol.io/specification/server/resources/)
- [ADR-007: MCP Resources Implementation](../adr/ADR-007-mcp-resources-implementation.md)
- [TF-IDF Index Documentation](../elements/README.md#semantic-search)
- [Configuration Reference](../configuration/README.md)

## Future Enhancements

Planned improvements (see [NEXT_STEPS.md](../../NEXT_STEPS.md)):

1. **Resource Templates** (M1.3)
   - Dynamic filtering: `capability-index://elements/{type}`
   - Time-based queries: `capability-index://recent?days=7`

2. **Resource Subscriptions** (M1.4)
   - Real-time updates when index changes
   - WebSocket-based notifications

3. **Custom Resources** (M2.0)
   - User-defined resource generators
   - Plugin system for resource providers

4. **Performance Monitoring** (M1.5)
   - Resource generation metrics
   - Cache hit/miss statistics
   - Token usage tracking

---

**Version:** 1.0.0  
**Last Updated:** 2025-12-19  
**Author:** NEXS-MCP Team
