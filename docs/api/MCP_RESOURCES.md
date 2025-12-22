# MCP Resources Reference

**NEXS-MCP v1.1.0**  
**SDK:** [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk) (`github.com/modelcontextprotocol/go-sdk/mcp`)  
**Last Updated:** December 22, 2025

## Overview

NEXS-MCP implements the [Model Context Protocol (MCP) Resources specification](https://spec.modelcontextprotocol.io/specification/2024-11-05/server/resources/) using the official MCP Go SDK, allowing MCP clients to access capability indices and element information without explicit tool calls.

**SDK Implementation:**
- Resources registered via `server.AddResource()`
- URI-based resource identification
- MIME type support from SDK
- Standard resource handler interface

Resources provide read-only access to structured information about the NEXS-MCP portfolio, enabling clients to understand available capabilities and make informed decisions about tool usage.

---

## Table of Contents

- [What are MCP Resources?](#what-are-mcp-resources)
- [Enabling Resources](#enabling-resources)
- [Resource URIs](#resource-uris)
- [Resource Details](#resource-details)
- [Accessing Resources](#accessing-resources)
- [Caching](#caching)
- [Use Cases](#use-cases)
- [Performance Considerations](#performance-considerations)

---

## What are MCP Resources?

MCP Resources are server-provided data sources that clients can read without invoking tools. They are identified by URIs and can represent:

- Static information (documentation, schemas)
- Dynamic data (capability indices, statistics)
- Large datasets that would be inefficient to pass through tool responses

### Benefits

- **Efficient:** No tool call overhead for accessing information
- **Cacheable:** Clients can cache resources with TTL
- **Discoverable:** Clients can list available resources
- **Structured:** Resources have well-defined MIME types and schemas

---

## Enabling Resources

Resources are **disabled by default**. Enable them via:

### Command Line

```bash
nexs-mcp --resources-enabled=true
```

### Environment Variable

```bash
export NEXS_RESOURCES_ENABLED=true
nexs-mcp
```

### Configuration File

```yaml
# config.yaml
resources:
  enabled: true
  expose:
    - summary
    - full
    - stats
  cache_ttl: 3600  # 1 hour in seconds
```

### Selective Exposure

Expose only specific resources:

```bash
# Expose only summary and stats
nexs-mcp --resources-enabled=true --resources-expose=summary,stats
```

```yaml
# config.yaml
resources:
  enabled: true
  expose:
    - summary
    - stats
```

---

## Resource URIs

NEXS-MCP provides three capability index resources:

| URI | Name | MIME Type | Size | Description |
|-----|------|-----------|------|-------------|
| `capability://nexs-mcp/index/summary` | Capability Index Summary | `text/markdown` | ~3K tokens | Concise overview |
| `capability://nexs-mcp/index/full` | Capability Index Full Details | `text/markdown` | ~40K tokens | Complete details |
| `capability://nexs-mcp/index/stats` | Capability Index Statistics | `application/json` | ~500 bytes | Statistical data |

### URI Scheme

```
capability://nexs-mcp/{category}/{resource-name}
```

- **Scheme:** `capability://` - Indicates capability-related resource
- **Authority:** `nexs-mcp` - Server identifier
- **Path:** `/{category}/{resource-name}` - Resource location

---

## Resource Details

### 1. Capability Index Summary

**URI:** `capability://nexs-mcp/index/summary`

#### Description
A concise, human-readable summary of the capability index including:
- Element counts by type
- Top keywords and tags
- Recently added elements
- Quick statistics

#### MIME Type
`text/markdown`

#### Size
Approximately 3,000 tokens (~12KB)

#### Update Frequency
Updates when elements are added, updated, or deleted

#### Example Content

```markdown
# NEXS-MCP Capability Index Summary

## Overview
- **Total Elements:** 42
- **Active Elements:** 38
- **Last Updated:** 2025-12-20T10:00:00Z
- **Index Health:** ✅ Healthy

## Element Distribution
| Type | Count | Percentage |
|------|-------|------------|
| Personas | 10 | 24% |
| Skills | 15 | 36% |
| Templates | 8 | 19% |
| Agents | 5 | 12% |
| Memories | 3 | 7% |
| Ensembles | 1 | 2% |

## Top Keywords
1. **technical** (15 elements) - Architecture, system design, engineering
2. **automation** (12 elements) - CI/CD, workflows, agents
3. **documentation** (10 elements) - Writing, reports, guides
4. **analysis** (8 elements) - Data, business intelligence
5. **code-review** (7 elements) - Quality assurance, best practices

## Popular Tags
- `engineering` (18 elements)
- `writing` (12 elements)
- `automation` (10 elements)
- `analytics` (8 elements)

## Recently Added
- **Technical Architect** (persona) - Added 2025-12-20
  - Expert in system architecture and design patterns
  - Tags: technical, architecture, system-design, enterprise
  
- **Code Review Expert** (skill) - Added 2025-12-19
  - Comprehensive code review with security analysis
  - Tags: code-review, quality-assurance, best-practices

- **CI/CD Automation Agent** (agent) - Added 2025-12-18
  - Automated testing and deployment workflows
  - Tags: ci-cd, automation, devops, testing

## Most Used Elements
1. **Technical Writer** (persona) - Used in 12 agents
2. **Data Analysis** (skill) - Used in 8 workflows
3. **Technical Report** (template) - Rendered 45 times

## Index Statistics
- **Unique Terms:** 1,523
- **Index Size:** 512 KB
- **Last Rebuild:** 2025-12-20T09:00:00Z
- **Average Search Time:** 23ms

## Quick Links
- Use `search_capability_index` to find specific capabilities
- Use `find_similar_capabilities` to discover related elements
- Use `map_capability_relationships` to explore connections
```

#### Use Cases
- Quick overview of available capabilities
- Understanding element distribution
- Discovering popular elements
- Monitoring index health

---

### 2. Capability Index Full Details

**URI:** `capability://nexs-mcp/index/full`

#### Description
Complete, detailed view of the capability index including:
- All elements with full metadata
- Relationships and connections
- Complete vocabulary index
- Detailed statistics

#### MIME Type
`text/markdown`

#### Size
Approximately 40,000 tokens (~160KB)

#### Update Frequency
Updates when elements are added, updated, or deleted

#### Example Content Structure

```markdown
# NEXS-MCP Capability Index - Full Details

## Table of Contents
1. [All Elements](#all-elements)
   - [Personas](#personas)
   - [Skills](#skills)
   - [Templates](#templates)
   - [Agents](#agents)
   - [Memories](#memories)
   - [Ensembles](#ensembles)
2. [Relationships](#relationships)
3. [Vocabulary Index](#vocabulary-index)
4. [Detailed Statistics](#detailed-statistics)

---

## All Elements

### Personas (10 elements)

#### 1. Technical Architect
- **ID:** `technical-architect-001`
- **Status:** Active
- **Version:** 1.0.0
- **Description:** An experienced technical architect focused on system design, scalability, and best practices for enterprise-grade software development.
- **Created:** 2025-12-20T00:00:00Z
- **Updated:** 2025-12-20T00:00:00Z
- **Author:** NEXS-MCP Team
- **Tags:** technical, architecture, system-design, enterprise

**Expertise Areas:**
1. System architecture and design patterns
2. Microservices and distributed systems
3. Cloud architecture (AWS, GCP, Azure)
4. Database design and optimization
5. API design and integration

**Behavioral Traits:**
- analytical (0.9)
- strategic (0.85)
- detail-oriented (0.8)
- pragmatic (0.9)

**Response Style:**
- Tone: professional and technical
- Formality: high
- Verbosity: balanced

**Related Elements:**
- Used in: `ci-automation-agent-001`, `monitoring-agent-001`
- Similar to: `senior-engineer-001` (similarity: 0.82)
- Complementary with: `code-review-expert-001` (similarity: 0.75)

---

#### 2. Creative Content Writer
[Similar detailed structure...]

---

### Skills (15 elements)

#### 1. Code Review Expert
- **ID:** `code-review-expert-001`
- **Status:** Active
- **Version:** 1.0.0
- **Description:** Expert-level code review skill that analyzes code quality, security, performance, and maintainability with actionable feedback.

**Triggers:**
- code review request
- pull request submission
- quality check needed
- security audit

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| code | string | Yes | The code to review |
| language | string | Yes | Programming language |
| context | string | No | Additional context |
| focus_areas | array | No | Specific areas to focus on |

**Procedure (7 steps):**
1. Parse code and identify language-specific patterns
2. Check code quality and style
3. Analyze architecture and design patterns
4. Identify security vulnerabilities
5. Assess performance implications
6. Review test coverage and quality
7. Generate comprehensive feedback

**Output Format:**
```json
{
  "overall_rating": 1-10,
  "critical_issues": [],
  "warnings": [],
  "suggestions": [],
  "positive_notes": [],
  "security_score": 1-10,
  "performance_score": 1-10,
  "maintainability_score": 1-10
}
```

**Dependencies:** None
**Used By:** `code-review-team-001`, `ci-automation-agent-001`

---

[Continue for all elements...]

---

## Relationships

### Semantic Similarity Network

#### Technical Architect (persona)
**High Similarity (>0.8):**
- Senior Engineer (persona) - 0.82
  - Shared focus on architecture and best practices
  
- Code Review Expert (skill) - 0.79
  - Both emphasize quality and maintainability

**Medium Similarity (0.6-0.8):**
- System Monitoring Agent (agent) - 0.72
- CI/CD Automation Agent (agent) - 0.68

**Complementary Relationships:**
- Code Review Expert (skill) - Works well together for quality assurance
- Technical Report (template) - Useful for documenting decisions

---

[Continue for all elements...]

---

## Vocabulary Index

### Statistics
- **Total Unique Terms:** 1,523
- **Most Common Terms:** technical (89 occurrences), system (67), data (56)
- **Average Terms per Element:** 36

### Terms by Category

#### Technical Terms (423)
- architecture: 45 elements
- system: 42 elements
- design: 38 elements
- pattern: 32 elements
- api: 28 elements
[...]

#### Domain Terms (234)
- automation: 25 elements
- testing: 22 elements
- deployment: 18 elements
[...]

#### Business Terms (156)
- project: 15 elements
- team: 14 elements
- workflow: 12 elements
[...]

---

## Detailed Statistics

### Element Statistics
- **Total:** 42 elements
- **Active:** 38 (90.5%)
- **Inactive:** 4 (9.5%)

### By Author
- NEXS-MCP Team: 30 (71%)
- Community: 12 (29%)

### By Creation Date
- Last 7 days: 5
- Last 30 days: 12
- Last 90 days: 25
- Older: 17

### Usage Statistics
- **Most Referenced:** Technical Writer (12 references)
- **Most Executed:** Data Analysis Skill (45 executions)
- **Most Rendered:** Technical Report Template (67 renders)

### Index Performance
- **Last Rebuild:** 2025-12-20T09:00:00Z
- **Rebuild Duration:** 234ms
- **Index Size:** 512 KB
- **Average Search Time:** 23ms
- **P95 Search Time:** 67ms
- **P99 Search Time:** 120ms

### Cache Statistics
- **Hit Rate:** 95.6%
- **Total Hits:** 1,234
- **Total Misses:** 56
- **Cache Size:** 100 KB
- **Cached Entries:** 42
```

#### Use Cases
- Deep exploration of capabilities
- Understanding element relationships
- Building capability maps
- Analyzing vocabulary and taxonomy

---

### 3. Capability Index Statistics

**URI:** `capability://nexs-mcp/index/stats`

#### Description
Machine-readable statistical data about the capability index in JSON format.

#### MIME Type
`application/json`

#### Size
Approximately 500 bytes

#### Update Frequency
Updates in real-time as elements change

#### JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["timestamp", "elements", "index", "cache", "performance"],
  "properties": {
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "When statistics were generated"
    },
    "elements": {
      "type": "object",
      "properties": {
        "total": {"type": "integer"},
        "active": {"type": "integer"},
        "by_type": {
          "type": "object",
          "properties": {
            "persona": {"type": "integer"},
            "skill": {"type": "integer"},
            "template": {"type": "integer"},
            "agent": {"type": "integer"},
            "memory": {"type": "integer"},
            "ensemble": {"type": "integer"}
          }
        },
        "by_author": {
          "type": "object",
          "additionalProperties": {"type": "integer"}
        }
      }
    },
    "index": {
      "type": "object",
      "properties": {
        "total_documents": {"type": "integer"},
        "unique_terms": {"type": "integer"},
        "index_size_bytes": {"type": "integer"},
        "last_rebuild": {"type": "string", "format": "date-time"}
      }
    },
    "cache": {
      "type": "object",
      "properties": {
        "hits": {"type": "integer"},
        "misses": {"type": "integer"},
        "hit_rate": {"type": "number", "minimum": 0, "maximum": 1},
        "size_bytes": {"type": "integer"},
        "entries": {"type": "integer"}
      }
    },
    "performance": {
      "type": "object",
      "properties": {
        "avg_search_time_ms": {"type": "number"},
        "p95_search_time_ms": {"type": "number"},
        "p99_search_time_ms": {"type": "number"}
      }
    }
  }
}
```

#### Example Content

```json
{
  "timestamp": "2025-12-20T10:00:00Z",
  "elements": {
    "total": 42,
    "active": 38,
    "by_type": {
      "persona": 10,
      "skill": 15,
      "template": 8,
      "agent": 5,
      "memory": 3,
      "ensemble": 1
    },
    "by_author": {
      "nexs-team": 30,
      "community": 12
    }
  },
  "index": {
    "total_documents": 42,
    "unique_terms": 1523,
    "index_size_bytes": 524288,
    "last_rebuild": "2025-12-20T09:00:00Z"
  },
  "cache": {
    "hits": 1234,
    "misses": 56,
    "hit_rate": 0.956,
    "size_bytes": 102400,
    "entries": 42
  },
  "performance": {
    "avg_search_time_ms": 23,
    "p95_search_time_ms": 67,
    "p99_search_time_ms": 120
  }
}
```

#### Use Cases
- Programmatic monitoring
- Dashboard integration
- Performance tracking
- Automated alerts

---

## Accessing Resources

### From MCP Clients

MCP clients can access resources using the resources list and read operations:

```typescript
// List available resources
const resources = await client.listResources();

// Read a specific resource
const summary = await client.readResource("capability://nexs-mcp/index/summary");
console.log(summary.contents[0].text);
```

### Claude Desktop

Resources are automatically available to Claude when NEXS-MCP is configured:

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "nexs-mcp",
      "args": ["--resources-enabled=true"],
      "env": {
        "NEXS_DATA_DIR": "~/.nexs-mcp/data"
      }
    }
  }
}
```

Claude can then reference resources implicitly when answering questions about capabilities.

---

## Caching

### Server-Side Caching

NEXS-MCP implements intelligent caching for resources:

- **Cache TTL:** Configurable (default: 1 hour)
- **Cache Invalidation:** Automatic on element changes
- **Cache Strategy:** LRU (Least Recently Used)

**Configuration:**

```yaml
resources:
  cache_ttl: 3600  # 1 hour
```

```bash
nexs-mcp --resources-cache-ttl=7200  # 2 hours
```

### Client-Side Caching

Clients should respect cache headers and implement their own caching:

```typescript
// Example cache implementation
const cache = new Map();

async function getCachedResource(uri, ttl = 3600000) {
  const cached = cache.get(uri);
  if (cached && Date.now() - cached.timestamp < ttl) {
    return cached.data;
  }
  
  const data = await client.readResource(uri);
  cache.set(uri, { data, timestamp: Date.now() });
  return data;
}
```

---

## Use Cases

### 1. Capability Discovery

AI assistants can read the summary resource to understand what capabilities are available without querying the server multiple times.

**Example:**
```
User: "What personas do you have?"
Assistant: [Reads capability://nexs-mcp/index/summary]
"I have 10 personas including Technical Architect, Creative Writer, and Data Analyst..."
```

### 2. Performance Monitoring

Monitoring systems can periodically read the stats resource to track index health.

**Example:**
```javascript
setInterval(async () => {
  const stats = await client.readResource("capability://nexs-mcp/index/stats");
  if (stats.performance.p95_search_time_ms > 100) {
    alert("Search performance degraded");
  }
}, 60000);
```

### 3. Recommendation Engines

Clients can use the full details resource to build recommendation systems based on element relationships.

**Example:**
```
User: "I'm using Technical Architect persona. What skills should I add?"
Assistant: [Reads capability://nexs-mcp/index/full, analyzes relationships]
"Based on relationships, I recommend: Code Review Expert (similarity: 0.79), System Monitoring (0.72)..."
```

### 4. Documentation Generation

Tools can read resources to automatically generate documentation.

**Example:**
```bash
# Generate capability catalog
nexs-doc-generator --source=capability://nexs-mcp/index/full --output=catalog.md
```

---

## Performance Considerations

### Resource Size

| Resource | Size | Tokens | Load Time (typical) |
|----------|------|--------|---------------------|
| Summary | ~12 KB | ~3,000 | <50ms |
| Full | ~160 KB | ~40,000 | <200ms |
| Stats | ~500 B | ~150 | <10ms |

### Best Practices

1. **Use Summary First:** Start with the summary resource for quick overview
2. **Cache Aggressively:** Resources change infrequently, cache on client side
3. **Conditional Access:** Only read full details when necessary
4. **Monitor Stats:** Use stats resource for health checks and monitoring
5. **Respect TTL:** Honor cache TTL to reduce server load

### Scaling

- **Small portfolios (<100 elements):** All resources load quickly
- **Medium portfolios (100-1000 elements):** Summary remains fast, full resource may take 1-2 seconds
- **Large portfolios (>1000 elements):** Consider pagination or filtering strategies

---

## Troubleshooting

### Resource Not Available

**Problem:** `Resource 'capability://nexs-mcp/index/summary' not found`

**Solutions:**
1. Ensure resources are enabled: `--resources-enabled=true`
2. Check exposed resources: `--resources-expose=summary`
3. Verify server is running with resources support

### Empty or Outdated Content

**Problem:** Resource shows no elements or old data

**Solutions:**
1. Trigger index rebuild: use `reload_elements` tool
2. Clear cache: restart server or use `get_capability_index_stats` tool
3. Check if elements exist: use `list_elements` tool

### Slow Performance

**Problem:** Reading resources takes too long

**Solutions:**
1. Enable caching: set appropriate `cache_ttl`
2. Use summary instead of full resource
3. Implement client-side caching
4. Check index health with stats resource

---

## Future Enhancements

Planned improvements to the Resources API:

- **Resource Templates:** `capability://nexs-mcp/index/{type}` for type-specific views
- **Query Parameters:** `capability://nexs-mcp/index/summary?tags=technical`
- **Webhooks:** Notify clients when resources change
- **Versioning:** Track resource version for client cache invalidation
- **Compression:** GZIP compression for large resources
- **Pagination:** Paginated full resource for very large portfolios

---

## References

- [MCP Resources Specification](https://spec.modelcontextprotocol.io/specification/2024-11-05/server/resources/)
- [NEXS-MCP API Reference](./MCP_TOOLS.md)
- [NEXS-MCP Architecture](../architecture/MCP.md)

---

**Last Updated:** December 20, 2025  
**NEXS-MCP Version:** v1.0.0  
**MCP Protocol Version:** 2024-11-05
