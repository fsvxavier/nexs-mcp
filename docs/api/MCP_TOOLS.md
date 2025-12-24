# NEXS-MCP API Reference

**Version:** v1.2.0  
**Protocol:** Model Context Protocol (MCP)  
**SDK:** [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk) (`github.com/modelcontextprotocol/go-sdk/mcp`)  
**Last Updated:** December 23, 2025

This document provides complete reference documentation for all NEXS-MCP tools, resources, and APIs.

**Note:** NEXS-MCP is built using the official Model Context Protocol Go SDK, ensuring full compliance with the MCP specification and compatibility with all MCP clients.

---

## Table of Contents

- [MCP Tools](#mcp-tools)
  - [Element Management](#element-management)
  - [Quick Create Tools](#quick-create-tools)
  - [Element Operations](#element-operations)
  - [GitHub Integration](#github-integration)
  - [Backup & Restore](#backup--restore)
  - [Memory Management](#memory-management)
  - [Memory Quality](#memory-quality)
  - [Analytics & Performance](#analytics--performance)
  - [User Context](#user-context)
  - [Capability Index & Search](#capability-index--search)
  - [Collections](#collections)
  - [Template Operations](#template-operations)
  - [Validation](#validation)
- [MCP Resources](#mcp-resources)
- [CLI Reference](#cli-reference)
- [Error Handling](#error-handling)
- [Authentication](#authentication)

---

## MCP Tools

NEXS-MCP provides 91 MCP tools across 19 categories, implemented using the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk). All tools follow the Model Context Protocol specification and return structured JSON responses.

**SDK Integration:**
- Package: `github.com/modelcontextprotocol/go-sdk/mcp`
- Tool registration via `sdk.AddTool()`
- Request/Response types from official SDK
- Full stdio transport support

### Element Management

#### `list_elements`
List all elements with optional filtering.

**Parameters:**
```json
{
  "type": "string",           // Optional: Filter by type (persona, skill, template, agent, memory, ensemble)
  "tags": ["string"],         // Optional: Filter by tags
  "is_active": boolean,       // Optional: Filter by active status
  "author": "string",         // Optional: Filter by author
  "limit": number,            // Optional: Maximum results
  "offset": number            // Optional: Pagination offset
}
```

**Response:**
```json
{
  "elements": [
    {
      "id": "persona-001",
      "name": "Technical Writer",
      "type": "persona",
      "description": "...",
      "is_active": true,
      "created_at": "2025-12-20T10:00:00Z",
      "updated_at": "2025-12-20T10:00:00Z",
      "tags": ["writing", "technical"]
    }
  ],
  "total": 42,
  "offset": 0,
  "limit": 50
}
```

**Example:**
```json
{
  "type": "persona",
  "is_active": true,
  "limit": 10
}
```

---

#### `get_element`
Get a specific element by ID.

**Parameters:**
```json
{
  "id": "string"              // Required: Element ID
}
```

**Response:**
```json
{
  "element": {
    "id": "persona-001",
    "name": "Technical Writer",
    "type": "persona",
    "description": "...",
    // Type-specific fields...
  }
}
```

**Example:**
```json
{
  "id": "persona-001"
}
```

---

#### `search_elements`
Search elements with full-text search and advanced filtering.

**Parameters:**
```json
{
  "query": "string",          // Required: Search query
  "type": "string",           // Optional: Filter by type
  "tags": ["string"],         // Optional: Filter by tags
  "author": "string",         // Optional: Filter by author
  "date_from": "ISO8601",     // Optional: Created after date
  "date_to": "ISO8601",       // Optional: Created before date
  "limit": number,            // Optional: Max results (default: 20)
  "min_score": number         // Optional: Minimum relevance score (0-1)
}
```

**Response:**
```json
{
  "results": [
    {
      "element": { /* element object */ },
      "score": 0.92,
      "highlights": ["...matched text..."]
    }
  ],
  "total": 15,
  "query_time_ms": 45
}
```

**Example:**
```json
{
  "query": "code review best practices",
  "type": "skill",
  "limit": 5
}
```

---

#### `create_element`
Create a new element (generic - use type-specific tools for full features).

**Parameters:**
```json
{
  "type": "string",           // Required: Element type
  "name": "string",           // Required: Element name
  "description": "string",    // Required: Element description
  "tags": ["string"],         // Optional: Tags
  "metadata": {}              // Optional: Additional metadata
}
```

**Response:**
```json
{
  "id": "element-001",
  "message": "Element created successfully"
}
```

---

#### `create_persona`
Create a new Persona element with behavioral traits, expertise areas, and response styles.

**Parameters:**
```json
{
  "name": "string",                          // Required
  "description": "string",                   // Required
  "system_prompt": "string",                 // Required: Core behavior definition
  "behavioral_traits": [                     // Required
    {
      "name": "string",
      "description": "string",
      "intensity": 0.8                       // 0.0-1.0
    }
  ],
  "expertise_areas": [                       // Required
    {
      "domain": "string",
      "description": "string",
      "proficiency_level": "expert",         // beginner/intermediate/expert
      "keywords": ["string"]
    }
  ],
  "response_styles": {
    "tone": "professional",
    "formality": "moderate",
    "verbosity": "balanced"
  },
  "tags": ["string"],
  "author": "string"
}
```

**Response:**
```json
{
  "id": "persona-001",
  "preview": { /* persona object */ },
  "message": "Persona created successfully"
}
```

**Example:**
```json
{
  "name": "Senior DevOps Engineer",
  "description": "Experienced DevOps engineer focused on CI/CD and infrastructure automation",
  "system_prompt": "You are a senior DevOps engineer...",
  "behavioral_traits": [
    {
      "name": "pragmatic",
      "description": "Focuses on practical, scalable solutions",
      "intensity": 0.9
    }
  ],
  "expertise_areas": [
    {
      "domain": "CI/CD",
      "description": "Continuous integration and deployment pipelines",
      "proficiency_level": "expert",
      "keywords": ["jenkins", "github-actions", "gitlab-ci"]
    }
  ],
  "response_styles": {
    "tone": "professional",
    "formality": "moderate",
    "verbosity": "concise"
  },
  "tags": ["devops", "automation"]
}
```

---

#### `create_skill`
Create a new Skill element with triggers, procedures, and dependencies.

**Parameters:**
```json
{
  "name": "string",                          // Required
  "description": "string",                   // Required
  "triggers": [                              // Required
    {
      "pattern": "string",                   // Regex or keyword
      "context": "string",
      "keywords": ["string"]
    }
  ],
  "procedures": [                            // Required
    {
      "step": 1,
      "action": "string",
      "description": "string",
      "required": true
    }
  ],
  "parameters": [                            // Optional
    {
      "name": "string",
      "type": "string",
      "required": boolean,
      "description": "string",
      "default_value": "any"
    }
  ],
  "dependencies": ["skill-id"],              // Optional
  "output_format": {                         // Optional
    "type": "structured",
    "schema": {}
  },
  "tags": ["string"]
}
```

**Response:**
```json
{
  "id": "skill-001",
  "preview": { /* skill object */ },
  "message": "Skill created successfully"
}
```

---

#### `create_template`
Create a new Template element with variable substitution.

**Parameters:**
```json
{
  "name": "string",                          // Required
  "description": "string",                   // Required
  "content": "string",                       // Required: Template content with {{variables}}
  "variables": [                             // Required
    {
      "name": "string",
      "type": "string",                      // string/number/boolean/array/object
      "description": "string",
      "required": boolean,
      "default_value": "any"
    }
  ],
  "format": "markdown",                      // Optional: markdown/json/yaml/text
  "tags": ["string"]
}
```

**Response:**
```json
{
  "id": "template-001",
  "preview": { /* template object */ },
  "message": "Template created successfully"
}
```

---

#### `create_agent`
Create a new Agent element with goals, actions, and decision trees.

**Parameters:**
```json
{
  "name": "string",                          // Required
  "description": "string",                   // Required
  "persona_id": "string",                    // Required: Associated persona
  "skills": ["skill-id"],                    // Required: Array of skill IDs
  "goals": [                                 // Required
    {
      "description": "string",
      "priority": "high",                    // high/medium/low
      "success_criteria": "string"
    }
  ],
  "trigger_conditions": [                    // Optional
    {
      "event": "string",
      "condition": "string"
    }
  ],
  "workflow": [                              // Optional: Agent workflow steps
    {
      "step": 1,
      "name": "string",
      "description": "string",
      "actions": ["string"]
    }
  ],
  "tags": ["string"]
}
```

**Response:**
```json
{
  "id": "agent-001",
  "preview": { /* agent object */ },
  "message": "Agent created successfully"
}
```

---

#### `create_memory`
Create a new Memory element with automatic content hashing and **multilingual keyword extraction**. Supports 11 languages with automatic detection, reducing AI context usage by 70-85% through intelligent keyword indexing.

**Token Optimization:**
- Automatically extracts keywords from content using language-specific stop word filtering
- Supports: English, Portuguese, Spanish, French, German, Italian, Russian, Japanese, Chinese, Arabic, Hindi
- Reduces typical conversation memory from 1000+ tokens to 200-300 tokens when retrieving context
- Deduplicates content via SHA-256 hashing to prevent storing duplicate conversations

**Parameters:**
```json
{
  "name": "string",                          // Required
  "content": "string",                       // Required: Memory content
  "description": "string",                   // Optional
  "memory_type": "semantic",                 // semantic/episodic/procedural
  "scope": "session",                        // session/project/global
  "retention_period": "30d",                 // Optional: 7d/30d/90d/365d/permanent
  "tags": ["string"],
  "author": "string"
}
```

**Response:**
```json
{
  "id": "memory-001",
  "content_hash": "sha256:abc123...",
  "message": "Memory created successfully"
}
```

---

#### `create_ensemble`
Create a new Ensemble element for multi-agent orchestration.

**Parameters:**
```json
{
  "name": "string",                          // Required
  "description": "string",                   // Required
  "members": [                               // Required: Array of ensemble members
    {
      "id": "string",                        // Member ID
      "agent_id": "string",                  // Agent to use
      "role": "string",                      // Member role
      "weight": 0.5                          // 0.0-1.0: importance weight
    }
  ],
  "execution_mode": "sequential",            // sequential/parallel/hybrid
  "execution_order": ["member-id"],          // Required for sequential
  "aggregation_strategy": "consensus",       // first/last/consensus/voting/all/merge
  "consensus_threshold": 0.75,               // For consensus strategy
  "fallback_chain": ["member-id"],           // Optional: Fallback order
  "shared_context": {},                      // Optional: Shared data
  "tags": ["string"]
}
```

**Response:**
```json
{
  "id": "ensemble-001",
  "preview": { /* ensemble object */ },
  "message": "Ensemble created successfully"
}
```

---

### Quick Create Tools

Quick create tools provide simplified, one-shot element creation with minimal input and template defaults. No preview step needed - elements are created immediately.

#### `quick_create_persona`
**Description:** QUICK: Create persona with minimal input using template defaults (no preview needed)

**Parameters:**
```json
{
  "name": "string",                          // Required
  "description": "string",                   // Required
  "expertise": ["string"],                   // Required: List of expertise areas
  "traits": ["string"],                      // Required: Behavioral traits
  "tags": ["string"]                         // Optional
}
```

**Response:**
```json
{
  "id": "persona-quick-001",
  "message": "Persona created successfully"
}
```

**Example:**
```json
{
  "name": "Data Analyst",
  "description": "Analyzes data and creates insights",
  "expertise": ["SQL", "Python", "Data Visualization"],
  "traits": ["analytical", "detail-oriented", "communicative"]
}
```

---

#### `quick_create_skill`
**Description:** QUICK: Create skill with minimal input using template defaults (no preview needed)

**Parameters:**
```json
{
  "name": "string",                          // Required
  "description": "string",                   // Required
  "triggers": ["string"],                    // Required: Trigger keywords
  "procedure": ["string"],                   // Required: Step descriptions
  "tags": ["string"]                         // Optional
}
```

**Example:**
```json
{
  "name": "API Testing",
  "description": "Test REST APIs for functionality",
  "triggers": ["test api", "api testing"],
  "procedure": [
    "Review API documentation",
    "Create test cases",
    "Execute tests",
    "Report results"
  ]
}
```

---

#### `quick_create_memory`
**Description:** QUICK: Create memory with minimal input (no preview needed). **Automatically extracts keywords in 11 languages** and computes content hash for deduplication, reducing token usage by 70-85% when retrieving context.

**Parameters:**
```json
{
  "name": "string",                          // Required
  "content": "string",                       // Required
  "tags": ["string"]                         // Optional
}
```

---

#### `quick_create_template`
**Description:** QUICK: Create template with minimal input (no preview needed)

**Parameters:**
```json
{
  "name": "string",                          // Required
  "content": "string",                       // Required: Template with {{variables}}
  "tags": ["string"]                         // Optional
}
```

---

#### `quick_create_agent`
**Description:** QUICK: Create agent with minimal input (no preview needed)

**Parameters:**
```json
{
  "name": "string",                          // Required
  "description": "string",                   // Required
  "persona_id": "string",                    // Required
  "skills": ["skill-id"],                    // Required
  "tags": ["string"]                         // Optional
}
```

---

#### `quick_create_ensemble`
**Description:** QUICK: Create ensemble with minimal input (no preview needed)

**Parameters:**
```json
{
  "name": "string",                          // Required
  "description": "string",                   // Required
  "members": [                               // Required
    {
      "agent_id": "string",
      "role": "string"
    }
  ],
  "execution_mode": "sequential",            // Required
  "tags": ["string"]                         // Optional
}
```

---

#### `batch_create_elements`
**Description:** BATCH: Create multiple elements at once (single confirmation for all)

**Parameters:**
```json
{
  "elements": [                              // Required: Array of element specs
    {
      "type": "persona",
      "name": "string",
      "description": "string",
      // ... type-specific fields
    }
  ],
  "auto_link": boolean,                      // Optional: Auto-link related elements
  "validate": boolean                        // Optional: Validate before creation (default: true)
}
```

**Response:**
```json
{
  "created": ["id1", "id2", "id3"],
  "failed": [],
  "message": "3 elements created successfully"
}
```

---

### Element Operations

#### `update_element`
Update an existing element.

**Parameters:**
```json
{
  "id": "string",                            // Required
  "updates": {                               // Required: Fields to update
    "name": "string",
    "description": "string",
    "is_active": boolean,
    "tags": ["string"],
    // ... type-specific fields
  }
}
```

**Response:**
```json
{
  "message": "Element updated successfully",
  "updated_fields": ["name", "description"]
}
```

---

#### `delete_element`
Delete an element by ID.

**Parameters:**
```json
{
  "id": "string"                             // Required
}
```

**Response:**
```json
{
  "message": "Element deleted successfully"
}
```

---

#### `duplicate_element`
Duplicate an existing element with a new ID and optional new name.

**Parameters:**
```json
{
  "id": "string",                            // Required: Source element ID
  "new_name": "string",                      // Optional: Name for duplicate
  "deep_copy": boolean                       // Optional: Deep copy related elements (default: false)
}
```

**Response:**
```json
{
  "id": "element-002",
  "message": "Element duplicated successfully"
}
```

---

#### `activate_element`
Activate an element by ID (shortcut for updating is_active to true).

**Parameters:**
```json
{
  "id": "string"                             // Required
}
```

---

#### `deactivate_element`
Deactivate an element by ID (shortcut for updating is_active to false).

**Parameters:**
```json
{
  "id": "string"                             // Required
}
```

---

### Ensemble Operations

#### `execute_ensemble`
Execute an ensemble with specified input and options.

**Description:** Orchestrates multiple agents according to ensemble configuration (sequential/parallel/hybrid modes).

**Parameters:**
```json
{
  "ensemble_id": "string",                   // Required
  "input": {},                               // Required: Input data for ensemble
  "options": {                               // Optional
    "timeout": 300,                          // Timeout in seconds
    "dry_run": false,                        // Simulate without executing
    "override_mode": "parallel"              // Override execution mode
  }
}
```

**Response:**
```json
{
  "execution_id": "exec-001",
  "status": "completed",
  "result": {
    "aggregated_output": {},
    "individual_results": [
      {
        "member_id": "member-1",
        "output": {},
        "execution_time_ms": 1234
      }
    ],
    "consensus_level": 0.85,
    "execution_mode": "parallel"
  },
  "total_time_ms": 2345
}
```

---

#### `get_ensemble_status`
Get status and configuration of an ensemble.

**Parameters:**
```json
{
  "ensemble_id": "string"                    // Required
}
```

**Response:**
```json
{
  "ensemble": {
    "id": "ensemble-001",
    "name": "Code Review Team",
    "members": [...],
    "execution_mode": "parallel",
    "aggregation_strategy": "weighted_consensus",
    "is_active": true
  },
  "statistics": {
    "total_executions": 42,
    "success_rate": 0.95,
    "average_execution_time_ms": 2100
  }
}
```

---

### GitHub Integration

#### `github_auth_start`
Start GitHub OAuth2 device flow authentication.

**Parameters:**
```json
{}
```

**Response:**
```json
{
  "device_code": "abc123",
  "user_code": "ABCD-EFGH",
  "verification_uri": "https://github.com/login/device",
  "expires_in": 900,
  "interval": 5,
  "message": "Visit https://github.com/login/device and enter code: ABCD-EFGH"
}
```

---

#### `github_auth_status`
Check the status of GitHub authentication.

**Parameters:**
```json
{
  "device_code": "string"                    // Optional: From auth_start
}
```

**Response:**
```json
{
  "status": "authorized",                    // pending/authorized/expired/denied
  "username": "fsvxavier",
  "token_expiry": "2025-12-27T10:00:00Z"
}
```

---

#### `check_github_auth`
Check GitHub authentication status and token validity.

**Parameters:**
```json
{}
```

**Response:**
```json
{
  "authenticated": true,
  "username": "fsvxavier",
  "token_valid": true,
  "expires_at": "2025-12-27T10:00:00Z",
  "scopes": ["repo", "user"]
}
```

---

#### `refresh_github_token`
Refresh GitHub OAuth token if expired or about to expire.

**Parameters:**
```json
{}
```

**Response:**
```json
{
  "message": "Token refreshed successfully",
  "expires_at": "2025-12-28T10:00:00Z"
}
```

---

#### `init_github_auth`
Initialize GitHub device flow authentication.

**Parameters:**
```json
{}
```

**Response:**
```json
{
  "verification_uri": "https://github.com/login/device",
  "user_code": "ABCD-EFGH",
  "message": "Please visit the URL and enter the code to authenticate"
}
```

---

#### `github_list_repos`
List all repositories for the authenticated GitHub user.

**Parameters:**
```json
{
  "visibility": "all",                       // all/public/private
  "sort": "updated",                         // created/updated/pushed/full_name
  "per_page": 30
}
```

**Response:**
```json
{
  "repositories": [
    {
      "name": "nexs-mcp",
      "full_name": "fsvxavier/nexs-mcp",
      "description": "...",
      "private": false,
      "stars": 42,
      "forks": 5,
      "updated_at": "2025-12-20T10:00:00Z"
    }
  ],
  "total": 15
}
```

---

#### `github_sync_push`
Push local elements to a GitHub repository.

**Parameters:**
```json
{
  "repository": "owner/repo",                // Required
  "branch": "main",                          // Optional (default: main)
  "element_ids": ["id1", "id2"],             // Optional: Specific elements (default: all)
  "commit_message": "string",                // Optional
  "create_pr": boolean                       // Optional: Create PR instead of direct push
}
```

**Response:**
```json
{
  "commit_sha": "abc123...",
  "files_updated": 5,
  "url": "https://github.com/owner/repo/commit/abc123"
}
```

---

#### `github_sync_pull`
Pull elements from a GitHub repository to local storage.

**Parameters:**
```json
{
  "repository": "owner/repo",                // Required
  "branch": "main",                          // Optional
  "merge_strategy": "local_wins",            // local_wins/remote_wins/newest_wins/merge_content/manual
  "element_types": ["persona", "skill"]      // Optional: Filter by types
}
```

**Response:**
```json
{
  "elements_synced": 10,
  "conflicts": 2,
  "resolution_strategy": "local_wins",
  "details": [
    {
      "element_id": "persona-001",
      "action": "updated",
      "conflict": false
    }
  ]
}
```

---

#### `github_sync_bidirectional`
Perform full bidirectional sync with GitHub repository (pull then push with conflict resolution).

**Parameters:**
```json
{
  "repository": "owner/repo",                // Required
  "branch": "main",                          // Optional
  "conflict_resolution": "newest_wins",      // Strategy for conflicts
  "dry_run": false                           // Optional: Simulate without changes
}
```

**Response:**
```json
{
  "pull_summary": {
    "elements_updated": 5,
    "conflicts_resolved": 2
  },
  "push_summary": {
    "elements_pushed": 3,
    "commit_sha": "abc123"
  },
  "total_time_ms": 3456
}
```

---

#### `search_portfolio_github`
Search GitHub repositories for NEXS portfolios and elements.

**Parameters:**
```json
{
  "query": "string",                         // Required
  "element_type": "persona",                 // Optional
  "author": "string",                        // Optional
  "tags": ["string"],                        // Optional
  "sort": "stars",                           // stars/relevance/updated
  "per_page": 10
}
```

**Response:**
```json
{
  "results": [
    {
      "repository": "user/repo",
      "elements": [
        {
          "id": "persona-001",
          "name": "...",
          "type": "persona"
        }
      ],
      "stars": 42,
      "description": "..."
    }
  ],
  "total": 5
}
```

---

### Backup & Restore

#### `backup_portfolio`
Create a compressed backup of all portfolio elements with checksum validation.

**Parameters:**
```json
{
  "backup_name": "string",                   // Optional: Custom backup name
  "include_types": ["persona", "skill"],     // Optional: Specific types (default: all)
  "compression": "gzip",                     // Optional: gzip/none
  "include_metadata": true                   // Optional: Include statistics
}
```

**Response:**
```json
{
  "backup_file": "/path/to/nexs-backup-20251220-100000.tar.gz",
  "checksum": "sha256:abc123...",
  "elements_backed_up": 42,
  "size_bytes": 1048576,
  "created_at": "2025-12-20T10:00:00Z"
}
```

---

#### `restore_portfolio`
Restore portfolio from a backup file with merge strategies and optional pre-restore backup.

**Parameters:**
```json
{
  "backup_file": "string",                   // Required: Path to backup file
  "merge_strategy": "replace",               // replace/merge/skip_existing
  "validate_checksums": true,                // Optional: Verify integrity
  "create_backup_before": true,              // Optional: Backup current state first
  "element_types": ["persona"]               // Optional: Restore specific types only
}
```

**Response:**
```json
{
  "elements_restored": 42,
  "skipped": 0,
  "errors": 0,
  "pre_restore_backup": "/path/to/pre-restore-backup.tar.gz",
  "message": "Restore completed successfully"
}
```

---

### Memory Management

**Token Optimization:** All memory tools support multilingual keyword extraction (11 languages) and automatic deduplication, reducing AI context usage by 70-85%. See [CONVERSATION_HISTORY_ANALYSIS.md](../CONVERSATION_HISTORY_ANALYSIS.md) for detailed token savings strategies.

#### `search_memory`
Search memories with relevance scoring and date filtering. **Multilingual support**: Automatically detects language and searches across keyword indexes in 11 languages (EN, PT, ES, FR, DE, IT, RU, JA, ZH, AR, HI). Returns only relevant memories to minimize token usage.

**Parameters:**
```json
{
  "query": "string",                         // Required
  "memory_type": "semantic",                 // Optional: semantic/episodic/procedural
  "scope": "session",                        // Optional: session/project/global
  "date_from": "ISO8601",                    // Optional
  "date_to": "ISO8601",                      // Optional
  "limit": 10,                               // Optional
  "min_score": 0.7                           // Optional: Relevance threshold
}
```

**Response:**
```json
{
  "results": [
    {
      "memory": {
        "id": "memory-001",
        "name": "...",
        "content": "...",
        "memory_type": "semantic"
      },
      "score": 0.92,
      "highlights": ["...matched content..."]
    }
  ],
  "total": 5
}
```

---

#### `summarize_memories`
Get a summary and statistics of memories with optional filtering.

**Parameters:**
```json
{
  "memory_type": "semantic",                 // Optional
  "scope": "project",                        // Optional
  "date_from": "ISO8601",                    // Optional
  "date_to": "ISO8601"                       // Optional
}
```

**Response:**
```json
{
  "total_memories": 150,
  "by_type": {
    "semantic": 100,
    "episodic": 40,
    "procedural": 10
  },
  "by_scope": {
    "session": 50,
    "project": 80,
    "global": 20
  },
  "total_size_bytes": 2097152,
  "oldest": "2025-01-01T00:00:00Z",
  "newest": "2025-12-20T10:00:00Z"
}
```

---

#### `update_memory`
Update content, name, description, tags, or metadata of an existing memory.

**Parameters:**
```json
{
  "id": "string",                            // Required
  "updates": {
    "name": "string",
    "content": "string",
    "description": "string",
    "tags": ["string"],
    "metadata": {}
  }
}
```

---

#### `delete_memory`
Delete a specific memory by ID.

**Parameters:**
```json
{
  "id": "string"                             // Required
}
```

---

#### `clear_memories`
Clear multiple memories with optional author/date filtering (requires confirmation).

**Parameters:**
```json
{
  "memory_type": "episodic",                 // Optional: Filter by type
  "scope": "session",                        // Optional: Filter by scope
  "author": "string",                        // Optional: Filter by author
  "date_before": "ISO8601",                  // Optional: Delete memories before date
  "confirm": true                            // Required: Must be true
}
```

**Response:**
```json
{
  "deleted": 25,
  "message": "25 memories cleared"
}
```

---

#### `save_conversation_context`
Save conversation context as a memory (auto-save feature).

**Description:** Automatically stores conversation history for continuity.

**Parameters:**
```json
{
  "context": "string",                       // Required: Conversation context
  "metadata": {                              // Optional
    "session_id": "string",
    "user": "string",
    "timestamp": "ISO8601"
  }
}
```

**Response:**
```json
{
  "memory_id": "memory-conv-001",
  "message": "Conversation context saved"
}
```

---

### Memory Quality

NEXS-MCP provides ONNX-based quality scoring for intelligent memory retention and lifecycle management. Three tools enable quality assessment and retention policy management.

#### `score_memory_quality`
Score memory content quality using ONNX models with multi-tier fallback.

**Parameters:**
```json
{
  "content": "string",                       // Required: Memory content to score
  "use_fallback": boolean                    // Optional: Enable multi-tier fallback (default: true)
}
```

**Response:**
```json
{
  "score": 0.75,
  "confidence": 0.92,
  "method": "onnx",
  "timestamp": "2025-12-23T10:00:00Z",
  "metadata": {
    "model": "ms-marco-MiniLM-L-6-v2",
    "latency_ms": 61.64,
    "fallback_used": false
  }
}
```

**Scoring Methods:**
- **ONNX** (primary): Local SLM inference (MS MARCO or Paraphrase-Multilingual)
- **Groq API** (fallback 1): Fast cloud inference
- **Gemini API** (fallback 2): High-quality cloud scoring
- **Implicit** (fallback 3): Signal-based heuristics (access count, recency, etc.)

**ONNX Models:**
- **MS MARCO** (default): 61ms latency, 9 languages (no CJK)
- **Paraphrase-Multilingual** (configurable): 109ms latency, 11 languages (with CJK)

**Example:**
```json
{
  "content": "This is a comprehensive technical guide explaining ONNX model integration with detailed code examples and performance benchmarks.",
  "use_fallback": true
}
```

---

#### `get_retention_policy`
Get the appropriate retention policy for a given quality score.

**Parameters:**
```json
{
  "score": 0.75                              // Required: Quality score (0.0-1.0)
}
```

**Response:**
```json
{
  "policy": {
    "min_quality": 0.7,
    "max_quality": 1.1,
    "retention_days": 365,
    "archive_after_days": 180,
    "description": "High quality - retained for 1 year, archived after 6 months"
  }
}
```

**Retention Tiers:**
- **High Quality** (≥0.7): 365 days retention, archived after 180 days
- **Medium Quality** (0.5-0.7): 180 days retention, archived after 90 days
- **Low Quality** (<0.5): 90 days retention, archived after 30 days

**Example:**
```json
{
  "score": 0.82
}
```

---

#### `get_retention_stats`
Get memory retention statistics and quality distribution.

**Parameters:**
```json
{}
```

**Response:**
```json
{
  "stats": {
    "total_scored": 1234,
    "total_archived": 156,
    "total_deleted": 23,
    "last_cleanup": "2025-12-23T09:00:00Z",
    "avg_quality_score": 0.68,
    "policy_breakdown": {
      "high": 345,
      "medium": 567,
      "low": 322
    }
  }
}
```

**Example:**
```json
{}
```

---

### Analytics & Performance

#### `get_usage_stats`
Get usage statistics and analytics for tool calls with period filtering.

**Parameters:**
```json
{
  "period": "7d",                            // 1d/7d/30d/90d/all
  "tool_name": "string",                     // Optional: Filter by tool
  "group_by": "tool"                         // Optional: tool/date/user
}
```

**Response:**
```json
{
  "period": "7d",
  "total_calls": 1234,
  "unique_tools": 25,
  "by_tool": {
    "list_elements": 450,
    "search_elements": 230,
    "create_persona": 120
  },
  "success_rate": 0.98,
  "average_latency_ms": 45,
  "most_used_tool": "list_elements"
}
```

---

#### `get_performance_dashboard`
Get performance metrics dashboard with latency percentiles and slow operation alerts.

**Parameters:**
```json
{
  "period": "24h"                            // 1h/24h/7d/30d
}
```

**Response:**
```json
{
  "period": "24h",
  "metrics": {
    "p50_latency_ms": 25,
    "p95_latency_ms": 120,
    "p99_latency_ms": 450,
    "max_latency_ms": 2300,
    "total_operations": 5432,
    "errors": 12,
    "error_rate": 0.002
  },
  "slow_operations": [
    {
      "tool": "github_sync_bidirectional",
      "latency_ms": 2300,
      "timestamp": "2025-12-20T09:45:00Z"
    }
  ],
  "health_status": "healthy"
}
```

---

#### `list_logs`
Query and filter structured logs with date range, level, and keyword filtering.

**Parameters:**
```json
{
  "level": "info",                           // debug/info/warn/error
  "date_from": "ISO8601",                    // Optional
  "date_to": "ISO8601",                      // Optional
  "keyword": "string",                       // Optional: Search in messages
  "limit": 100                               // Optional
}
```

**Response:**
```json
{
  "logs": [
    {
      "timestamp": "2025-12-20T10:00:00Z",
      "level": "info",
      "message": "Element created successfully",
      "metadata": {
        "element_id": "persona-001",
        "tool": "create_persona"
      }
    }
  ],
  "total": 250
}
```

---

### User Context

#### `get_current_user`
Get the current authenticated user and session context.

**Parameters:**
```json
{}
```

**Response:**
```json
{
  "user": {
    "id": "user-001",
    "username": "john_doe",
    "email": "john@example.com",
    "github_connected": true
  },
  "session": {
    "id": "session-abc",
    "started_at": "2025-12-20T09:00:00Z",
    "last_activity": "2025-12-20T10:00:00Z"
  }
}
```

---

#### `set_user_context`
Set the current user context for the session with optional metadata.

**Parameters:**
```json
{
  "username": "string",                      // Required
  "email": "string",                         // Optional
  "metadata": {}                             // Optional: Custom metadata
}
```

**Response:**
```json
{
  "message": "User context set successfully"
}
```

---

#### `clear_user_context`
Clear the current user context (requires confirmation).

**Parameters:**
```json
{
  "confirm": true                            // Required
}
```

---

### Capability Index & Search

#### `search_capability_index`
Search for capabilities using semantic search across all elements.

**Description:** Uses TF-IDF indexing to find relevant personas, skills, templates, agents, memories, and ensembles based on query text. Returns ranked results with relevance scores and text highlights.

**Parameters:**
```json
{
  "query": "string",                         // Required
  "types": ["persona", "skill"],             // Optional: Filter by types
  "limit": 10,                               // Optional
  "min_score": 0.5                           // Optional
}
```

**Response:**
```json
{
  "results": [
    {
      "element_id": "persona-001",
      "name": "Technical Writer",
      "type": "persona",
      "score": 0.92,
      "highlights": ["...relevant excerpts..."]
    }
  ],
  "total": 15,
  "query_time_ms": 23
}
```

---

#### `find_similar_capabilities`
Find capabilities similar to a given element.

**Description:** Uses semantic similarity to discover related personas, skills, templates, agents, memories, or ensembles. Useful for discovering complementary capabilities or alternatives.

**Parameters:**
```json
{
  "element_id": "string",                    // Required
  "types": ["persona", "skill"],             // Optional: Filter result types
  "limit": 5,                                // Optional
  "min_similarity": 0.6                      // Optional
}
```

**Response:**
```json
{
  "source_element": {
    "id": "persona-001",
    "name": "Technical Writer",
    "type": "persona"
  },
  "similar": [
    {
      "element_id": "persona-002",
      "name": "Documentation Specialist",
      "type": "persona",
      "similarity": 0.87
    }
  ],
  "total": 3
}
```

---

#### `map_capability_relationships`
Map relationships between a capability and related elements.

**Description:** Analyzes semantic similarity to build a relationship graph showing complementary, similar, and related capabilities. Helps understand capability ecosystems.

**Parameters:**
```json
{
  "element_id": "string",                    // Required
  "depth": 2,                                // Optional: Relationship depth (1-3)
  "min_similarity": 0.5                      // Optional
}
```

**Response:**
```json
{
  "root": {
    "id": "persona-001",
    "name": "Technical Writer",
    "type": "persona"
  },
  "relationships": {
    "complementary": [
      {
        "element_id": "skill-001",
        "name": "API Documentation",
        "similarity": 0.85,
        "relationship_type": "complementary"
      }
    ],
    "similar": [...],
    "related": [...]
  },
  "graph_size": 15
}
```

---

#### `get_capability_index_stats`
Get statistics about the capability index.

**Description:** Shows total indexed documents, distribution by type, unique terms, and index health. Useful for monitoring and troubleshooting the semantic search system.

**Parameters:**
```json
{}
```

**Response:**
```json
{
  "total_documents": 42,
  "by_type": {
    "persona": 10,
    "skill": 15,
    "template": 8,
    "agent": 5,
    "memory": 3,
    "ensemble": 1
  },
  "unique_terms": 1523,
  "index_size_bytes": 524288,
  "last_updated": "2025-12-20T10:00:00Z",
  "health": "healthy"
}
```

---

### Collections

#### `search_collections`
Advanced collection search with rich formatting, filtering, sorting, and pagination.

**Parameters:**
```json
{
  "query": "string",                         // Optional: Search query
  "category": "string",                      // Optional: Filter by category
  "author": "string",                        // Optional: Filter by author
  "tags": ["string"],                        // Optional: Filter by tags
  "min_stars": 5,                            // Optional: Minimum stars
  "sort_by": "relevance",                    // relevance/stars/downloads/updated/created/name
  "page": 1,                                 // Optional: Page number
  "per_page": 20,                            // Optional: Results per page
  "format": "rich"                           // Optional: rich/plain
}
```

**Response:**
```json
{
  "collections": [
    {
      "id": "collection-001",
      "name": "AI Personas Collection",
      "description": "...",
      "author": "nexs-team",
      "category": "personas",
      "tags": ["ai", "personas"],
      "stars": 42,
      "downloads": 1234,
      "elements": {
        "total": 15,
        "by_type": {
          "persona": 10,
          "skill": 5
        }
      },
      "updated_at": "2025-12-20T10:00:00Z",
      "url": "https://github.com/nexs-mcp/collections/..."
    }
  ],
  "pagination": {
    "total": 50,
    "page": 1,
    "per_page": 20,
    "total_pages": 3
  }
}
```

---

#### `list_collections`
List available collections with optional rich formatting, grouping, and summary statistics.

**Parameters:**
```json
{
  "group_by": "category",                    // Optional: category/author/source
  "format": "rich",                          // Optional: rich/plain
  "include_stats": true                      // Optional: Include summary stats
}
```

**Response:**
```json
{
  "collections": [...],
  "total": 25,
  "summary": {
    "total_elements": 523,
    "total_downloads": 15234,
    "average_stars": 12.5,
    "by_category": {
      "personas": 10,
      "skills": 8,
      "templates": 7
    },
    "by_author": {
      "nexs-team": 15,
      "community": 10
    }
  }
}
```

---

#### `publish_collection`
Publish a collection to NEXS-MCP registry via GitHub Pull Request.

**Description:** Validates manifest with 100+ rules, scans for security issues with 50+ patterns, creates tarball with checksums, forks registry repo, creates branch, commits files, and opens PR. Supports dry-run mode for testing.

**Parameters:**
```json
{
  "manifest_file": "string",                 // Required: Path to manifest.yaml
  "dry_run": false,                          // Optional: Test without publishing
  "auto_fix": false                          // Optional: Auto-fix validation issues
}
```

**Response:**
```json
{
  "status": "submitted",
  "pr_url": "https://github.com/nexs-mcp/registry/pull/123",
  "validation_results": {
    "passed": 98,
    "warnings": 2,
    "errors": 0
  },
  "security_scan": {
    "issues_found": 0,
    "scan_time_ms": 234
  },
  "tarball": "/path/to/collection.tar.gz",
  "checksum": "sha256:abc123..."
}
```

---

#### `submit_element_to_collection`
Submit an element to a collection repository via GitHub Pull Request.

**Description:** Automatically forks the repo, creates a branch, commits the element, and opens a PR with generated description.

**Parameters:**
```json
{
  "element_id": "string",                    // Required
  "collection_repo": "owner/repo",           // Required
  "category": "string",                      // Optional: Element category
  "pr_title": "string",                      // Optional: Custom PR title
  "pr_description": "string"                 // Optional: Additional description
}
```

**Response:**
```json
{
  "pr_url": "https://github.com/owner/repo/pull/456",
  "pr_number": 456,
  "status": "open",
  "message": "Element submitted successfully"
}
```

---

### Template Operations

#### `render_template`
Render a template directly with provided data without creating an element.

**Description:** Supports both template_id (from repository) or direct template_content modes.

**Parameters:**
```json
{
  "template_id": "string",                   // Option 1: Use existing template
  "template_content": "string",              // Option 2: Provide template content directly
  "data": {},                                // Required: Data for variable substitution
  "format": "markdown"                       // Optional: Output format
}
```

**Response:**
```json
{
  "rendered": "...rendered content...",
  "variables_used": ["name", "date", "title"],
  "format": "markdown"
}
```

**Example:**
```json
{
  "template_content": "# {{title}}\n\nAuthor: {{author}}\nDate: {{date}}",
  "data": {
    "title": "My Report",
    "author": "John Doe",
    "date": "2025-12-20"
  }
}
```

**Output:**
```markdown
# My Report

Author: John Doe
Date: 2025-12-20
```

---

### Validation

#### `validate_element`
Perform comprehensive type-specific validation on an element.

**Parameters:**
```json
{
  "element_id": "string",                    // Option 1: Validate existing element
  "element_data": {},                        // Option 2: Validate data before creation
  "validation_level": "comprehensive",       // basic/comprehensive/strict
  "suggest_fixes": true                      // Optional: Include fix suggestions
}
```

**Response:**
```json
{
  "valid": false,
  "level": "comprehensive",
  "errors": [
    {
      "field": "behavioral_traits",
      "message": "Required field missing",
      "severity": "error"
    }
  ],
  "warnings": [
    {
      "field": "expertise_areas",
      "message": "Recommended to have at least 3 areas",
      "severity": "warning"
    }
  ],
  "suggestions": [
    {
      "field": "tags",
      "suggestion": "Add descriptive tags for better discoverability"
    }
  ],
  "score": 0.75
}
```

---

#### `reload_elements`
Hot reload elements from disk without server restart.

**Description:** Supports selective reload by element type with optional cache clearing and validation.

**Parameters:**
```json
{
  "element_types": ["persona", "skill"],     // Optional: Specific types (default: all)
  "clear_cache": true,                       // Optional: Clear index cache
  "validate": true                           // Optional: Validate after reload
}
```

**Response:**
```json
{
  "reloaded": 42,
  "by_type": {
    "persona": 10,
    "skill": 15,
    "template": 8,
    "agent": 5,
    "memory": 3,
    "ensemble": 1
  },
  "validation_errors": 0,
  "message": "Elements reloaded successfully"
}
```

---

## MCP Resources

NEXS-MCP implements the MCP Resources Protocol for exposing capability indices to MCP clients. Resources can be enabled/disabled via configuration.

### Configuration

```bash
nexs-mcp --resources-enabled=true --resources-expose=summary,stats
```

Or via config file:
```yaml
resources:
  enabled: true
  expose:
    - summary
    - stats
  cache_ttl: 3600  # 1 hour
```

### Available Resources

#### `capability://nexs-mcp/index/summary`
**Name:** Capability Index Summary  
**MIME Type:** text/markdown  
**Size:** ~3K tokens  
**Description:** A concise summary of the capability index including element counts, top keywords, and recent elements.

**Content:**
```markdown
# NEXS-MCP Capability Index Summary

## Overview
- **Total Elements:** 42
- **Active Elements:** 38
- **Last Updated:** 2025-12-20T10:00:00Z

## Element Distribution
- Personas: 10 (24%)
- Skills: 15 (36%)
- Templates: 8 (19%)
- Agents: 5 (12%)
- Memories: 3 (7%)
- Ensembles: 1 (2%)

## Top Keywords
1. technical (15 elements)
2. automation (12 elements)
3. documentation (10 elements)
...

## Recently Added
- Technical Architect (persona) - 2025-12-20
- Code Review Expert (skill) - 2025-12-19
...
```

---

#### `capability://nexs-mcp/index/full`
**Name:** Capability Index Full Details  
**MIME Type:** text/markdown  
**Size:** ~40K tokens  
**Description:** Complete detailed view of the capability index with all elements, metadata, relationships, and vocabulary.

**Content:**
```markdown
# NEXS-MCP Capability Index - Full Details

## All Elements

### Personas (10)

#### Technical Architect
- **ID:** technical-architect-001
- **Description:** An experienced technical architect...
- **Expertise:** System architecture, microservices, cloud architecture
- **Tags:** technical, architecture, system-design, enterprise
- **Created:** 2025-12-20T00:00:00Z

...

### Skills (15)

#### Code Review Expert
- **ID:** code-review-expert-001
- **Description:** Expert-level code review...
- **Triggers:** code review request, pull request submission
- **Tags:** code-review, quality-assurance
...

### Relationships

#### Technical Architect → Code Review Expert
- **Type:** Complementary
- **Similarity:** 0.85
- **Reason:** Both focus on code quality and best practices

...

### Vocabulary Index
- **Total Terms:** 1523
- **By Category:**
  - Technical: 423
  - Business: 234
  - Domain-specific: 866

...
```

---

#### `capability://nexs-mcp/index/stats`
**Name:** Capability Index Statistics  
**MIME Type:** application/json  
**Description:** Statistical data about the capability index in JSON format including element counts, index statistics, and cache metrics.

**Content:**
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

---

## CLI Reference

### Installation

```bash
# Go install
go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@latest

# Homebrew
brew install nexs-mcp

# NPM
npm install -g @fsvxavier/nexs-mcp-server

# Docker
docker pull fsvxavier/nexs-mcp:latest
```

### Usage

```bash
nexs-mcp [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--data-dir` | string | `~/.nexs-mcp/data` | Directory for element storage |
| `--storage-type` | string | `yaml` | Storage type (yaml/json) |
| `--log-level` | string | `info` | Log level (debug/info/warn/error) |
| `--log-format` | string | `text` | Log format (text/json) |
| `--log-file` | string | - | Log file path (default: stdout) |
| `--resources-enabled` | bool | `false` | Enable MCP Resources Protocol |
| `--resources-expose` | []string | all | Resource URIs to expose (summary/full/stats) |
| `--resources-cache-ttl` | int | `3600` | Resource cache TTL in seconds |
| `--github-client-id` | string | - | GitHub OAuth client ID |
| `--config` | string | - | Config file path |
| `--version` | bool | `false` | Print version and exit |
| `--help` | bool | `false` | Print help and exit |

### Environment Variables

Environment variables override default values but are overridden by command-line flags.

| Variable | Description | Example |
|----------|-------------|---------|
| `NEXS_DATA_DIR` | Data directory path | `/path/to/data` |
| `NEXS_STORAGE_TYPE` | Storage type | `yaml` |
| `NEXS_LOG_LEVEL` | Logging level | `debug` |
| `NEXS_LOG_FORMAT` | Log format | `json` |
| `NEXS_LOG_FILE` | Log file path | `/var/log/nexs-mcp.log` |
| `NEXS_RESOURCES_ENABLED` | Enable resources | `true` |
| `NEXS_RESOURCES_EXPOSE` | Resource URIs | `summary,stats` |
| `NEXS_RESOURCES_CACHE_TTL` | Cache TTL (seconds) | `7200` |
| `NEXS_GITHUB_CLIENT_ID` | GitHub OAuth client ID | `your-client-id` |

### Configuration File

NEXS-MCP can be configured via YAML file:

```yaml
# config.yaml
data_dir: ~/.nexs-mcp/data
storage_type: yaml

logging:
  level: info
  format: text
  file: /var/log/nexs-mcp.log

resources:
  enabled: true
  expose:
    - summary
    - stats
  cache_ttl: 3600

github:
  client_id: your-client-id-here

# Auto-save settings
auto_save:
  enabled: true
  interval: 300  # 5 minutes
  on_exit: true
```

Load config file:
```bash
nexs-mcp --config=/path/to/config.yaml
```

### Examples

**Start with debug logging:**
```bash
nexs-mcp --log-level=debug
```

**Use custom data directory:**
```bash
nexs-mcp --data-dir=/custom/path
```

**Enable MCP Resources:**
```bash
nexs-mcp --resources-enabled=true --resources-expose=summary,stats
```

**JSON storage with file logging:**
```bash
nexs-mcp --storage-type=json --log-format=json --log-file=/var/log/nexs.log
```

**Docker with custom config:**
```bash
docker run -v $(pwd)/config.yaml:/config.yaml \
  -v $(pwd)/data:/data \
  fsvxavier/nexs-mcp:latest \
  --config=/config.yaml
```

---

## Error Handling

All tools return errors in a consistent format:

```json
{
  "error": {
    "code": "ELEMENT_NOT_FOUND",
    "message": "Element with ID 'persona-999' not found",
    "details": {
      "element_id": "persona-999",
      "element_type": "persona"
    },
    "suggestion": "Use list_elements to see available elements"
  }
}
```

### Common Error Codes

| Code | Description | HTTP Equivalent |
|------|-------------|-----------------|
| `INVALID_INPUT` | Invalid parameters provided | 400 Bad Request |
| `ELEMENT_NOT_FOUND` | Element does not exist | 404 Not Found |
| `VALIDATION_ERROR` | Element failed validation | 422 Unprocessable Entity |
| `DUPLICATE_ELEMENT` | Element with ID already exists | 409 Conflict |
| `UNAUTHORIZED` | GitHub authentication required | 401 Unauthorized |
| `FORBIDDEN` | Insufficient permissions | 403 Forbidden |
| `RATE_LIMITED` | GitHub API rate limit exceeded | 429 Too Many Requests |
| `INTERNAL_ERROR` | Server internal error | 500 Internal Server Error |
| `SERVICE_UNAVAILABLE` | External service unavailable | 503 Service Unavailable |

---

## Authentication

### GitHub OAuth

NEXS-MCP uses GitHub Device Flow for authentication:

1. **Start authentication:**
   ```json
   {
     "tool": "github_auth_start"
   }
   ```

2. **Visit URL and enter code:**
   ```
   https://github.com/login/device
   Code: ABCD-EFGH
   ```

3. **Check status:**
   ```json
   {
     "tool": "github_auth_status",
     "parameters": {
       "device_code": "abc123"
     }
   }
   ```

4. **Token stored securely:**
   - Location: `~/.nexs-mcp/auth/github_token.enc`
   - Encryption: AES-256-GCM
   - Key derivation: PBKDF2 (100,000 iterations)

### Token Management

**Check authentication:**
```json
{
  "tool": "check_github_auth"
}
```

**Refresh token:**
```json
{
  "tool": "refresh_github_token"
}
```

**Scopes required:**
- `repo`: Repository access
- `user`: User information

---

## Rate Limits

### GitHub API
- **Unauthenticated:** 60 requests/hour
- **Authenticated:** 5,000 requests/hour
- **Search API:** 30 requests/minute

NEXS-MCP automatically handles rate limiting with exponential backoff.

### Local Operations
No rate limits for local element operations.

---

## Versioning

API version: **v1.0.0**

NEXS-MCP follows [Semantic Versioning](https://semver.org/):
- **Major:** Breaking API changes
- **Minor:** New features, backward compatible
- **Patch:** Bug fixes, backward compatible

---

## Support

- **Documentation:** https://github.com/fsvxavier/nexs-mcp/docs
- **Issues:** https://github.com/fsvxavier/nexs-mcp/issues
- **Discussions:** https://github.com/fsvxavier/nexs-mcp/discussions

---

**Last Updated:** December 20, 2025  
**API Version:** v1.0.0  
**MCP Protocol Version:** 2024-11-05
