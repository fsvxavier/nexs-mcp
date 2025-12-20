# ADR-009: Element Template System Architecture

**Status:** Approved  
**Date:** 2025-12-19  
**Milestone:** M0.9 - Element Templates  
**Complexity:** Medium (10 tasks, ~3,500 LOC)

## Context

NEXS MCP currently has a basic Template domain model with simple `{{variable}}` substitution via the `Render()` method. However, the system lacks:

1. **Template Discovery**: No registry or search capabilities for finding reusable templates
2. **Advanced Instantiation**: Limited to simple string replacement, no conditionals or iterations
3. **Template Validation**: No syntax checking or variable validation before instantiation
4. **Standard Library**: No built-in templates for common element types
5. **Template Publishing**: No workflow for sharing templates in collections

### Current State

**Existing Template Implementation:**
```go
// internal/domain/template.go
type Template struct {
    metadata        ElementMetadata
    Content         string             // Template content with {{variables}}
    Format          string             // markdown, yaml, json, text
    Variables       []TemplateVariable // Variable definitions
    ValidationRules map[string]string  // Validation constraints
}

func (t *Template) Render(values map[string]string) (string, error) {
    // Simple string replacement only
    result := t.Content
    for _, v := range t.Variables {
        val := values[v.Name] // or v.Default
        result = strings.ReplaceAll(result, "{{"+v.Name+"}}", val)
    }
    return result, nil
}
```

**Gaps:**
- No template registry or caching
- No advanced syntax (conditionals, loops)
- No template validation before instantiation
- No MCP tools for template operations
- No standard template library

## Decision

Build a comprehensive **Element Template System** with 5 core components:

### 1. Template Discovery System
- Template registry with caching (reuse M0.8 registry architecture)
- Search/filter by category, tags, element type
- Template metadata indexing

### 2. Advanced Instantiation Engine
- Handlebars-style syntax for rich templating
- Conditional blocks: `{{#if variable}}...{{/if}}`
- Iteration support: `{{#each items}}...{{/each}}`
- Nested templates and partials
- Type-safe variable substitution

### 3. Template Validation System
- Syntax validation (parse template before save)
- Variable presence checking
- Output structure validation
- Integration with collection validator

### 4. Standard Template Library
- Built-in templates for all 6 element types
- Persona templates (assistant, analyst, developer, etc.)
- Skill templates (API integration, data processing, etc.)
- Agent templates (research, automation, orchestration)
- Memory templates (conversation, knowledge, cache)
- Ensemble templates (multi-agent workflows)

### 5. Template MCP Tools
- `list_templates`: Browse available templates
- `get_template`: Retrieve template details
- `instantiate_template`: Create element from template
- `validate_template`: Check template syntax

## Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP Template Tools                        │
│  list_templates | get_template | instantiate_template       │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Template Registry                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Cache        │  │ Index        │  │ Standard Lib │      │
│  │ (15min TTL)  │  │ (Category)   │  │ (30+ built-in)│     │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                 Instantiation Engine                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Parser       │  │ Validator    │  │ Renderer     │      │
│  │ (Handlebars) │  │ (Syntax)     │  │ (Output)     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Element Factory                           │
│  Create typed elements (Persona, Skill, Agent, etc.)        │
└─────────────────────────────────────────────────────────────┘
```

### 1. Template Registry

**Package:** `internal/template/registry.go`

```go
// TemplateRegistry manages template discovery and caching
type TemplateRegistry struct {
    cache          *TemplateCache          // In-memory cache (15min TTL)
    index          *TemplateIndex          // Category/tag indexing
    standardLib    *StandardLibrary        // Built-in templates
    repo           ElementRepository       // Persistence
}

// TemplateCache provides fast template lookup
type TemplateCache struct {
    templates   map[string]*domain.Template
    expires     map[string]time.Time
    ttl         time.Duration // 15 minutes
    mu          sync.RWMutex
}

// TemplateIndex enables rich filtering
type TemplateIndex struct {
    byCategory   map[string][]string // persona, skill, agent, etc.
    byTag        map[string][]string
    byElementType map[string][]string
}

// StandardLibrary contains built-in templates
type StandardLibrary struct {
    templates map[string]*domain.Template // 30+ built-in
}
```

**Performance Targets:**
- Cache hits: <1µs (similar to M0.8 343ns)
- Search operations: <10ms
- Index rebuild: <100ms

### 2. Instantiation Engine

**Package:** `internal/template/engine.go`

```go
// InstantiationEngine renders templates with advanced syntax
type InstantiationEngine struct {
    parser    *TemplateParser
    validator *TemplateValidator
    renderer  *TemplateRenderer
}

// TemplateParser handles Handlebars-style syntax
type TemplateParser struct {
    // Parses template content into AST
}

// Supported Syntax:
// - Variables: {{name}}
// - Conditionals: {{#if condition}}...{{else}}...{{/if}}
// - Iterations: {{#each items}}{{this}}{{/each}}
// - Partials: {{> partial_name}}
// - Helpers: {{upper name}}, {{lower text}}

// TemplateValidator checks syntax and variables
type TemplateValidator struct {
    // Validates template before instantiation
}

// TemplateRenderer produces output
type TemplateRenderer struct {
    // Generates final content
}
```

**Example Advanced Template:**
```handlebars
# {{title}}

{{#if author}}
**Author:** {{author}}
{{/if}}

## Features

{{#each features}}
- {{this}}
{{/each}}

{{#if has_code}}
```{{language}}
{{code_content}}
```
{{/if}}
```

### 3. Template Validation

**Package:** `internal/template/validator.go`

```go
// TemplateValidator performs comprehensive validation
type TemplateValidator struct {
    syntaxRules    []ValidationRule
    variableRules  []ValidationRule
    outputRules    []ValidationRule
}

// Validation Checks:
// 1. Syntax Validation
//    - Balanced {{#if}}...{{/if}} blocks
//    - Valid helper functions
//    - Proper variable references
//
// 2. Variable Validation
//    - All {{variables}} declared
//    - Required variables have values
//    - Type constraints satisfied
//
// 3. Output Validation
//    - Valid YAML/JSON if format specified
//    - Required element fields present
//    - Schema compliance
```

### 4. Standard Template Library

**Package:** `internal/template/stdlib/`

```
stdlib/
├── personas/
│   ├── assistant.yaml         # Helpful assistant persona
│   ├── analyst.yaml           # Data analyst persona
│   ├── developer.yaml         # Software developer persona
│   ├── writer.yaml            # Content writer persona
│   └── researcher.yaml        # Research assistant persona
├── skills/
│   ├── api_integration.yaml   # REST API caller skill
│   ├── data_processing.yaml   # Data transformation skill
│   ├── file_operations.yaml   # File I/O skill
│   ├── web_scraping.yaml      # Web scraper skill
│   └── notification.yaml      # Notification sender skill
├── agents/
│   ├── research_agent.yaml    # Research workflow agent
│   ├── automation_agent.yaml  # Task automation agent
│   ├── monitoring_agent.yaml  # System monitoring agent
│   └── reporting_agent.yaml   # Report generation agent
├── memories/
│   ├── conversation.yaml      # Conversation memory
│   ├── knowledge_base.yaml    # Knowledge storage
│   ├── cache.yaml             # Temporary cache
│   └── session.yaml           # Session state
├── ensembles/
│   ├── multi_agent.yaml       # Multi-agent coordination
│   ├── pipeline.yaml          # Sequential workflow
│   └── parallel.yaml          # Parallel execution
└── templates/
    ├── email.yaml             # Email template
    ├── report.yaml            # Report template
    └── api_response.yaml      # API response template
```

**Total:** 30+ built-in templates

### 5. Template MCP Tools

**Package:** `internal/mcp/template_tools.go`

```go
// Tool 1: list_templates
type ListTemplatesInput struct {
    Category    string   // Filter by category (persona, skill, etc.)
    Tags        []string // Filter by tags
    ElementType string   // Filter by target element type
    IncludeBuiltIn bool  // Include standard library (default: true)
    Page        int      // Pagination
    PerPage     int      // Results per page (default: 20)
}

// Tool 2: get_template
type GetTemplateInput struct {
    ID string // Template ID
}

// Tool 3: instantiate_template
type InstantiateTemplateInput struct {
    TemplateID string            // Template to instantiate
    Variables  map[string]string // Variable values
    SaveAs     string            // Save to repository (optional)
    DryRun     bool              // Preview only (default: false)
}

// Tool 4: validate_template
type ValidateTemplateInput struct {
    TemplateID string            // Template to validate
    Variables  map[string]string // Test values (optional)
}
```

## Implementation Plan

### Task Breakdown (10 tasks, ~3,500 LOC)

**Task 1: Analyze Requirements (2h)**
- Review existing Template domain model
- Study Handlebars syntax specification
- Identify enhancement points
- Document gaps

**Task 2: Create ADR-009 (3h)**
- Architecture design (5 components)
- Sequence diagrams
- API specifications
- Performance targets

**Task 3: Template Registry (4h)**
- `internal/template/registry.go` (~400 LOC)
- TemplateCache with TTL
- TemplateIndex (category/tag/type)
- Integration with existing repository

**Task 4: Instantiation Engine (6h)**
- `internal/template/engine.go` (~500 LOC)
- Handlebars parser (use library: raymond/mustache)
- Variable substitution
- Conditional blocks (`{{#if}}`)
- Iteration support (`{{#each}}`)

**Task 5: Template Validator (4h)**
- `internal/template/validator.go` (~300 LOC)
- Syntax validation
- Variable checking
- Output validation
- Integration with collection validator

**Task 6: Standard Template Library (8h)**
- Create 30+ YAML templates (~1,500 LOC)
- 5 persona templates
- 5 skill templates
- 4 agent templates
- 4 memory templates
- 3 ensemble templates
- 3 template templates
- Loader: `internal/template/stdlib/loader.go` (~200 LOC)

**Task 7: Template MCP Tools (5h)**
- `internal/mcp/template_tools.go` (~400 LOC)
- list_templates implementation
- get_template implementation
- instantiate_template implementation
- validate_template implementation

**Task 8: Integration Tests (6h)**
- `test/integration/template_integration_test.go` (~500 LOC)
- Registry caching tests
- Instantiation tests (conditionals, loops)
- Validation tests
- Standard library tests
- All tools tests

**Task 9: Documentation (4h)**
- `docs/templates/TEMPLATES.md` (~8KB)
- Template authoring guide
- Variable reference
- Syntax examples
- Standard library catalog
- Best practices

**Task 10: Finalize M0.9 (2h)**
- Update NEXT_STEPS.md
- Commit changes
- Tag v0.9.0
- Generate release notes

**Total Effort:** ~44 hours (~1 week)

## Alternatives Considered

### Alternative 1: Simple String Replacement Only
**Rejected** - Too limited for complex templates. No conditionals or iterations.

### Alternative 2: Full-Featured Template Engine (Jinja2/Liquid)
**Rejected** - Over-engineered for our use case. Handlebars provides sufficient power with simpler syntax.

### Alternative 3: Go text/template Package
**Considered** - Good native option, but Handlebars has better JSON/YAML integration and is more familiar to users.

**Decision:** Use Handlebars-compatible library (raymond or mustache) for balance of power and simplicity.

## Consequences

### Positive

1. **Enhanced Productivity**: Users can quickly create elements from templates
2. **Consistency**: Standard templates enforce best practices
3. **Flexibility**: Advanced syntax supports complex use cases
4. **Discovery**: Easy to find and browse available templates
5. **Quality**: Validation prevents errors before instantiation

### Negative

1. **Complexity**: More code to maintain (~3,500 LOC)
2. **Learning Curve**: Users need to learn Handlebars syntax
3. **Performance**: Template parsing adds overhead (mitigated by caching)

### Mitigation

- Provide comprehensive documentation
- Include 30+ examples in standard library
- Cache parsed templates
- Validate early to prevent runtime errors

## Performance Targets

| Metric | Target | Strategy |
|--------|--------|----------|
| Cache Hits | <1µs | In-memory map with RWMutex |
| Template Search | <10ms | Indexed by category/tag/type |
| Template Parse | <5ms | Cache parsed AST |
| Instantiation | <10ms | Fast variable substitution |
| Validation | <5ms | Syntax-only checks |

## Testing Strategy

### Unit Tests (20 tests)
- Template registry caching
- Index operations
- Parser edge cases
- Validator rules
- Renderer output

### Integration Tests (15 tests)
- End-to-end instantiation
- Standard library loading
- MCP tool invocations
- Multi-template workflows
- Error scenarios

### Performance Tests (5 tests)
- Cache hit latency
- Search performance
- Parse/render benchmarks
- Concurrent access
- Memory usage

**Total Tests:** 40 tests, target >95% coverage

## Security Considerations

1. **Template Injection**: Validate all user inputs, escape dangerous characters
2. **Resource Limits**: Limit template size (<10MB), depth (<10 levels), iterations (<1000)
3. **Sandboxing**: No code execution in templates, pure data substitution
4. **Validation**: All templates validated before storage

## Success Criteria

1. ✅ 30+ standard templates available
2. ✅ Handlebars conditionals and iterations working
3. ✅ Template validation catches 95%+ errors
4. ✅ Cache hits <1µs latency
5. ✅ All 4 MCP tools functional
6. ✅ 40+ tests passing (>95% coverage)
7. ✅ Comprehensive documentation published

## Future Enhancements (Post-M0.9)

1. **Template Marketplace**: Community template sharing
2. **Visual Template Editor**: GUI for template creation
3. **Template Versioning**: Track template changes over time
4. **Template Composition**: Combine multiple templates
5. **AI-Generated Templates**: Use LLM to create templates from descriptions

## References

- **Handlebars Specification:** https://handlebarsjs.com/guide/
- **Raymond (Go Handlebars):** https://github.com/aymerick/raymond
- **Mustache Spec:** https://mustache.github.io/
- **NEXS Template Docs:** [docs/elements/TEMPLATE.md](../elements/TEMPLATE.md)
- **M0.8 Registry Architecture:** [ADR-008-collection-registry-production.md](ADR-008-collection-registry-production.md)

---

**Approved By:** @fsvxavier  
**Implementation Start:** 2025-12-19  
**Target Completion:** 2025-12-26 (1 week)
