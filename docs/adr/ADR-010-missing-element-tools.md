# ADR-010: Missing Element Tools Implementation

**Status:** Accepted  
**Date:** 2025-12-19  
**Deciders:** NEXS-MCP Development Team  
**Context:** M0.11 - Achieving feature parity with DollhouseMCP

## Context and Problem Statement

Analysis of NEXS-MCP vs DollhouseMCP revealed 4 missing element tools that prevent feature parity:

1. **validate_element** - Type-specific validation beyond basic schema
2. **render_template** - Direct template rendering with data
3. **reload_elements** - Hot reload without server restart
4. **search_portfolio_github** - Search GitHub repositories for portfolios

These gaps are documented in `comparing.md` and prevent NEXS-MCP from reaching full functional parity with DollhouseMCP. All 4 tools are rated as **Medium Impact** but collectively represent a **High Priority** milestone for user experience.

## Decision Drivers

* Feature parity with DollhouseMCP (from comparing.md analysis)
* User experience consistency across MCP servers
* Leverage existing NEXS-MCP infrastructure (template engine, GitHub client, repository)
* Minimal code duplication, maximum reusability
* Performance targets: <100ms per operation
* Comprehensive error reporting and validation

## Considered Options

### Option 1: Implement all 4 tools in separate files
* **Pros:** Clear separation of concerns, easy to test individually
* **Cons:** More files to maintain, potential code duplication

### Option 2: Implement all 4 tools in single file (element_tools.go)
* **Pros:** Centralized location, shared utilities, easier discovery
* **Cons:** Large file, mixed responsibilities

### Option 3: Group by functionality (validation_tools.go, rendering_tools.go, etc.)
* **Pros:** Logical grouping, moderate file sizes
* **Cons:** Less discoverable, scattered implementation

## Decision Outcome

**Chosen option:** Option 1 - Separate files for each major functionality group:

1. `internal/mcp/element_validation_tools.go` - validate_element
2. `internal/mcp/template_rendering_tools.go` - render_template  
3. `internal/mcp/element_reload_tools.go` - reload_elements
4. `internal/mcp/github_portfolio_tools.go` - search_portfolio_github

**Rationale:**
- Clear responsibility boundaries
- Easy to test in isolation
- Follows existing pattern (template_tools.go, github_tools.go, etc.)
- Better code organization for future maintenance
- Each tool can have dedicated test file

## Tool Specifications

### 1. validate_element Tool

**Purpose:** Type-specific comprehensive validation beyond basic schema checks

**Implementation File:** `internal/mcp/element_validation_tools.go`

**Schema:**
```json
{
  "name": "validate_element",
  "description": "Perform comprehensive type-specific validation on an element",
  "inputSchema": {
    "type": "object",
    "properties": {
      "element_id": {
        "type": "string",
        "description": "ID of the element to validate"
      },
      "element_type": {
        "type": "string",
        "enum": ["persona", "skill", "template", "agent", "memory", "ensemble"],
        "description": "Type of element being validated"
      },
      "validation_level": {
        "type": "string",
        "enum": ["basic", "comprehensive", "strict"],
        "default": "comprehensive",
        "description": "Level of validation to perform"
      },
      "fix_suggestions": {
        "type": "boolean",
        "default": true,
        "description": "Include suggestions for fixing validation errors"
      }
    },
    "required": ["element_id", "element_type"]
  }
}
```

**Implementation Details:**
- Leverage existing domain element Validate() methods
- Add type-specific validators for each element type:
  - **Persona:** Validate tone, expertise areas, communication style
  - **Skill:** Validate API schemas, parameters, authentication
  - **Template:** Validate Handlebars syntax, variable definitions, output format
  - **Agent:** Validate goals, actions, decision logic
  - **Memory:** Validate storage format, retention policy
  - **Ensemble:** Validate pipeline steps, dependencies, orchestration
- Return structured ValidationResult with:
  - `is_valid` (boolean)
  - `errors` (array of error objects with line numbers, fields, messages)
  - `warnings` (array of warning objects)
  - `suggestions` (array of fix suggestions if enabled)
  - `validation_time_ms` (performance metric)

**Dependencies:**
- `internal/domain` - Element types and Validate() methods
- `internal/template` - Template validator for template elements
- New package: `internal/validation` - Type-specific validators

**Performance Target:** <50ms per validation

---

### 2. render_template Tool

**Purpose:** Direct template rendering with provided data (without creating an element)

**Implementation File:** `internal/mcp/template_rendering_tools.go`

**Schema:**
```json
{
  "name": "render_template",
  "description": "Render a template directly with provided data",
  "inputSchema": {
    "type": "object",
    "properties": {
      "template_id": {
        "type": "string",
        "description": "ID of the template to render (optional if template_content provided)"
      },
      "template_content": {
        "type": "string",
        "description": "Template content to render (optional if template_id provided)"
      },
      "data": {
        "type": "object",
        "description": "Data object with variables to substitute in template",
        "additionalProperties": true
      },
      "output_format": {
        "type": "string",
        "enum": ["text", "markdown", "yaml", "json"],
        "default": "text",
        "description": "Format of the rendered output"
      },
      "validate_before_render": {
        "type": "boolean",
        "default": true,
        "description": "Validate template syntax before rendering"
      }
    },
    "required": ["data"]
  }
}
```

**Implementation Details:**
- Reuse existing InstantiationEngine from M0.9
- Support two modes:
  1. **Template ID mode:** Load template from repository, then render
  2. **Direct content mode:** Render provided template string directly
- Pre-render validation:
  - Check Handlebars syntax
  - Verify all required variables are provided
  - Validate data types match template variable definitions
- Return RenderResult:
  - `rendered_output` (string)
  - `variables_used` (array of variable names)
  - `missing_variables` (array of required but missing variables)
  - `render_time_ms` (performance metric)
  - `warnings` (array of non-fatal issues)

**Dependencies:**
- `internal/template/engine.go` - InstantiationEngine
- `internal/template/validator.go` - Template syntax validation
- `internal/domain` - Template repository access

**Performance Target:** <100ms per render (complex templates with loops)

---

### 3. reload_elements Tool

**Purpose:** Hot reload elements from disk without server restart

**Implementation File:** `internal/mcp/element_reload_tools.go`

**Schema:**
```json
{
  "name": "reload_elements",
  "description": "Reload elements from disk without restarting the server",
  "inputSchema": {
    "type": "object",
    "properties": {
      "element_types": {
        "type": "array",
        "items": {
          "type": "string",
          "enum": ["persona", "skill", "template", "agent", "memory", "ensemble", "all"]
        },
        "description": "Types of elements to reload (default: all)"
      },
      "clear_caches": {
        "type": "boolean",
        "default": true,
        "description": "Clear all related caches before reloading"
      },
      "validate_after_reload": {
        "type": "boolean",
        "default": true,
        "description": "Validate elements after reloading"
      }
    }
  }
}
```

**Implementation Details:**
- Invalidate all relevant caches:
  - Template registry cache
  - Collection registry cache
  - Any element-specific caches
- Reload elements from FileElementRepository:
  - Call repository.List() to get fresh data
  - Update in-memory indices
  - Refresh metadata index
- Validation pass (if enabled):
  - Run Validate() on each reloaded element
  - Collect and report any validation errors
- Return ReloadResult:
  - `elements_reloaded` (count by type)
  - `elements_failed` (count by type)
  - `validation_errors` (array if validation enabled)
  - `cache_stats` (before/after stats)
  - `reload_time_ms` (performance metric)

**Dependencies:**
- `internal/infrastructure` - FileElementRepository
- `internal/template` - TemplateRegistry cache invalidation
- `internal/collection` - CollectionRegistry cache invalidation
- `internal/domain` - Element types

**Performance Target:** <500ms for full reload (1000 elements)

**Edge Cases:**
- Handle file system changes during reload
- Graceful degradation if some elements fail to load
- Atomic updates to prevent partial state
- Notify connected clients about reload completion

---

### 4. search_portfolio_github Tool

**Purpose:** Search GitHub repositories for NEXS portfolios and elements

**Implementation File:** `internal/mcp/github_portfolio_tools.go`

**Schema:**
```json
{
  "name": "search_portfolio_github",
  "description": "Search GitHub repositories for NEXS portfolios and elements",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Search query (keywords, element names, etc.)"
      },
      "element_type": {
        "type": "string",
        "enum": ["persona", "skill", "template", "agent", "memory", "ensemble", "all"],
        "default": "all",
        "description": "Filter by element type"
      },
      "author": {
        "type": "string",
        "description": "Filter by GitHub username/org"
      },
      "tags": {
        "type": "array",
        "items": {"type": "string"},
        "description": "Filter by tags"
      },
      "sort_by": {
        "type": "string",
        "enum": ["stars", "updated", "created", "relevance"],
        "default": "relevance",
        "description": "Sort order for results"
      },
      "limit": {
        "type": "integer",
        "default": 20,
        "minimum": 1,
        "maximum": 100,
        "description": "Maximum number of results"
      },
      "include_archived": {
        "type": "boolean",
        "default": false,
        "description": "Include archived repositories"
      }
    },
    "required": ["query"]
  }
}
```

**Implementation Details:**
- Use GitHub API search endpoints:
  - `/search/repositories` for portfolio repos
  - `/search/code` for element files within repos
- Search strategy:
  1. **Repository search:** Find repos with "nexs-portfolio" topic or "nexs-mcp" in description
  2. **Code search:** Find YAML files matching element schemas
  3. **Content parsing:** Parse found files to extract metadata
  4. **Ranking:** Score by relevance, stars, recency
- Return SearchResult:
  - `results` (array of portfolio/element matches)
    - `repo_name` (owner/name)
    - `repo_url` (GitHub URL)
    - `stars` (integer)
    - `updated_at` (timestamp)
    - `elements_found` (array of element metadata)
    - `match_score` (relevance 0-100)
  - `total_count` (total matches available)
  - `page` (current page number)
  - `has_more` (boolean)
  - `search_time_ms` (performance metric)

**Dependencies:**
- `internal/infrastructure/github_client.go` - GitHub API integration
- Existing GitHub OAuth authentication
- Rate limiting and caching

**Performance Target:** <2000ms per search (GitHub API latency dependent)

**Edge Cases:**
- Handle GitHub API rate limits gracefully
- Cache search results (5min TTL)
- Support unauthenticated search (lower rate limits)
- Parse various portfolio formats (YAML, JSON)
- Handle malformed/invalid element files

---

## Integration Points

### Server Registration

All 4 tools will be registered in `internal/mcp/server.go`:

```go
// M0.11: Missing Element Tools
server.AddTool(NewValidateElementTool(repo))
server.AddTool(NewRenderTemplateTool(templateRegistry, templateEngine))
server.AddTool(NewReloadElementsTool(repo, templateRegistry, collectionRegistry))
server.AddTool(NewSearchPortfolioGitHubTool(githubClient))
```

### Testing Strategy

Each tool gets dedicated test file in `test/integration/`:
- `element_validation_test.go` - 15 test cases
- `template_rendering_test.go` - 12 test cases
- `element_reload_test.go` - 10 test cases
- `github_portfolio_search_test.go` - 8 test cases

**Total:** 45 integration tests

### Documentation

Update documentation:
- `NEXT_STEPS.md` - Mark M0.11 complete
- `comparing.md` - Update feature parity table
- `docs/tools/` - Add usage guides for each tool
- README.md - Update tool count (58 → 62)

## Consequences

### Positive

* ✅ **Feature parity** achieved with DollhouseMCP
* ✅ **Enhanced validation** - Type-specific validation beyond basic schema
* ✅ **Better UX** - Direct template rendering without element creation
* ✅ **Development efficiency** - Hot reload speeds up iteration
* ✅ **Discovery** - GitHub search enables portfolio sharing
* ✅ **Code reuse** - Leverages existing infrastructure (template engine, GitHub client)
* ✅ **Testing** - 45 new integration tests improve coverage
* ✅ **Documentation** - Clear specs for each tool

### Negative

* ⚠️ **Code complexity** - 4 new tool files (~1,200 LOC total)
* ⚠️ **Testing effort** - 45 integration tests require setup/teardown
* ⚠️ **GitHub rate limits** - search_portfolio_github may hit API limits
* ⚠️ **Cache invalidation** - reload_elements requires careful coordination
* ⚠️ **Maintenance** - More tools to maintain and evolve

### Neutral

* ℹ️ **Performance impact** - Minimal (<1% overhead)
* ℹ️ **Memory footprint** - Additional ~2MB for new code
* ℹ️ **Dependencies** - Reuses existing packages, no new external deps

## Implementation Plan

### Phase 1: Core Implementation (3 days)

**Day 1: Validation & Rendering**
- [ ] Create `internal/validation/` package with type-specific validators
- [ ] Implement `element_validation_tools.go` (~300 LOC)
- [ ] Implement `template_rendering_tools.go` (~200 LOC)
- [ ] Unit tests for validators

**Day 2: Reload & Search**
- [ ] Implement `element_reload_tools.go` (~250 LOC)
- [ ] Implement `github_portfolio_tools.go` (~350 LOC)
- [ ] Cache invalidation coordination
- [ ] GitHub search integration

**Day 3: Integration**
- [ ] Register all 4 tools in server.go
- [ ] Integration tests (45 tests, ~900 LOC)
- [ ] Error handling and edge cases

### Phase 2: Testing & Documentation (2 days)

**Day 4: Testing**
- [ ] Integration test suite completion
- [ ] Performance testing and optimization
- [ ] Edge case coverage
- [ ] Error scenario validation

**Day 5: Documentation**
- [ ] Update NEXT_STEPS.md with M0.11 completion
- [ ] Update comparing.md with parity status
- [ ] Create usage guides in docs/tools/
- [ ] Update README.md with new tool count

### Phase 3: Release (1 day)

**Day 6: Release**
- [ ] Code review and cleanup
- [ ] Final testing pass
- [ ] Commit with comprehensive message
- [ ] Tag v0.11.0
- [ ] Push to GitHub
- [ ] Update project board

**Total Effort:** 6 days (~48 hours)  
**Story Points:** 5 points  
**Risk Level:** Low (reuses existing infrastructure)

## Success Metrics

### Functional Metrics

* ✅ All 4 tools implemented and working
* ✅ 45 integration tests passing
* ✅ Feature parity confirmed in comparing.md
* ✅ Performance targets met:
  - validate_element: <50ms
  - render_template: <100ms
  - reload_elements: <500ms (1000 elements)
  - search_portfolio_github: <2000ms

### Quality Metrics

* ✅ Code coverage: 85%+ for new code
* ✅ Zero compilation errors
* ✅ All linter checks passing
* ✅ Documentation complete for all 4 tools

### User Experience Metrics

* ✅ Users can validate elements without manual checks
* ✅ Users can test templates without creating elements
* ✅ Developers can hot reload during iteration
* ✅ Users can discover portfolios on GitHub

## References

* [comparing.md](../../comparing.md) - Feature parity analysis
* [NEXT_STEPS.md](../../NEXT_STEPS.md#m011-missing-element-tools) - M0.11 specification
* [ADR-009: Element Template System](./ADR-009-element-template-system.md) - Template engine
* [internal/infrastructure/github_client.go](../../internal/infrastructure/github_client.go) - GitHub integration
* DollhouseMCP source code - Reference implementation

## Appendix: Type-Specific Validation Rules

### Persona Validation
- Tone must be consistent (formal/casual/technical)
- Expertise areas must be specific and actionable
- Communication style must match use case
- Example prompts should demonstrate expertise

### Skill Validation
- API endpoints must be valid URLs
- Authentication schemas must be complete
- Parameters must have types and descriptions
- Return values must be documented

### Template Validation
- Handlebars syntax must be valid
- All variables must be defined with types
- Required variables must be clearly marked
- Output format must match declared format

### Agent Validation
- Goals must be specific and measurable
- Actions must reference valid skills
- Decision logic must be deterministic
- State transitions must be well-defined

### Memory Validation
- Storage format must be valid (JSON/YAML)
- Retention policy must be reasonable
- Query patterns must be efficient
- Privacy settings must be respected

### Ensemble Validation
- Pipeline steps must form DAG (no cycles)
- All agents must exist
- Data flow between agents must be typed
- Error handling must be comprehensive

---

**Status:** Ready for implementation  
**Estimated Completion:** 2025-12-25 (6 days from start)  
**Blocked By:** None (all dependencies exist)  
**Blocks:** M0.12 Documentation & ADRs
