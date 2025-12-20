# NEXS MCP Protocol Layer

**Version:** 1.0.0  
**SDK:** [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk) v1.1.0 (`github.com/modelcontextprotocol/go-sdk/mcp`)  
**Last Updated:** December 20, 2025  
**Status:** Production

---

## Table of Contents

- [Introduction](#introduction)
- [Official MCP Go SDK](#official-mcp-go-sdk)
- [MCP Protocol Overview](#mcp-protocol-overview)
- [Server Implementation](#server-implementation)
- [Tool Registration](#tool-registration)
- [Element CRUD Tools](#element-crud-tools)
- [Quick Create Tools](#quick-create-tools)
- [GitHub Tools](#github-tools)
- [Collection Tools](#collection-tools)
- [Memory Tools](#memory-tools)
- [Analytics Tools](#analytics-tools)
- [Utility Tools](#utility-tools)
- [Resources Protocol](#resources-protocol)
- [Tool Handler Patterns](#tool-handler-patterns)
- [Error Handling](#error-handling)
- [Performance Optimization](#performance-optimization)
- [Best Practices](#best-practices)

---

## Introduction

The **MCP Layer** implements the Model Context Protocol (MCP) using the **official Model Context Protocol Go SDK** (`github.com/modelcontextprotocol/go-sdk/mcp`). It provides 55 tools for comprehensive element management, GitHub integration, analytics, and production features.

**Key Point:** NEXS-MCP is built entirely on the official MCP Go SDK, ensuring full specification compliance and compatibility with all MCP clients including Claude Desktop, continue.dev, and other MCP-compatible applications.

### MCP Layer Location

```
internal/mcp/
├── server.go                      # MCP server core
├── tools.go                       # Element CRUD tools (6)
├── type_specific_handlers.go      # Type-specific create tools (5)
├── quick_create_tools.go          # Quick create tools (6)
├── github_tools.go                # GitHub sync tools (3)
├── github_auth_tools.go           # OAuth tools (3)
├── github_portfolio_tools.go      # Portfolio management (4)
├── collection_tools.go            # Collection management (10)
├── collection_submission_tools.go # PR submission (2)
├── backup_tools.go                # Backup/restore (2)
├── memory_tools.go                # Memory management (4)
├── auto_save_tools.go             # Auto-save (2)
├── user_tools.go                  # User identity (2)
├── log_tools.go                   # Log query (1)
├── analytics_tools.go             # Analytics dashboard (1)
├── performance_tools.go           # Performance metrics (1)
├── search_tool.go                 # Semantic search (1)
├── index_tools.go                 # Index management (2)
├── ensemble_execution_tools.go    # Ensemble execution (2)
├── discovery_tools.go             # Element discovery (1)
├── batch_tools.go                 # Batch operations (1)
├── render_template_tools.go       # Template rendering (1)
├── element_validation_tools.go    # Validation (1)
├── reload_elements_tools.go       # Cache reload (1)
├── template_tools.go              # Template operations (4)
└── resources/                     # MCP Resources
    └── capability_index.go        # Capability indexing
```

### Total: 55 MCP Tools

---

## Official MCP Go SDK

### SDK Information

**Package:** `github.com/modelcontextprotocol/go-sdk/mcp`  
**Version:** v1.1.0  
**Repository:** https://github.com/modelcontextprotocol/go-sdk  
**Specification:** https://spec.modelcontextprotocol.io/

### Why the Official SDK?

NEXS-MCP uses the official Model Context Protocol Go SDK for several critical reasons:

1. **Specification Compliance** - Guarantees compatibility with MCP specification
2. **Client Compatibility** - Works with all MCP clients (Claude Desktop, continue.dev, etc.)
3. **Protocol Updates** - Automatic support for new protocol features
4. **Type Safety** - Strong typing for all MCP types and methods
5. **Maintained** - Active development and community support
6. **Standard Patterns** - Follows established MCP patterns

### SDK Core Types

```go
import sdk "github.com/modelcontextprotocol/go-sdk/mcp"

// Server creation
server := sdk.NewServer(impl, nil)

// Tool registration
sdk.AddTool(server, &sdk.Tool{
    Name:        "tool_name",
    Description: "Tool description",
}, handlerFunc)

// Resource registration
server.AddResource(&sdk.Resource{
    URI:         "capability://nexs-mcp/index/summary",
    Name:        "Capability Index Summary",
    Description: "Summary of available capabilities",
    MIMEType:    "text/markdown",
}, resourceHandler)

// Request/Response types
type CallToolRequest struct {
    Method string
    Params CallToolParams
}

type CallToolResult struct {
    Content []Content
    IsError bool
}
```

### SDK Integration Points

```go
// File: internal/mcp/server.go

import (
    sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// 1. Server initialization
func NewMCPServer(...) *MCPServer {
    impl := &sdk.Implementation{
        Name:    "NEXS-MCP",
        Version: version,
    }
    server := sdk.NewServer(impl, nil)
    // ...
}

// 2. Tool registration
func (s *MCPServer) registerTools() {
    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "create_persona",
        Description: "Create a new Persona element",
    }, s.handleCreatePersona)
}

// 3. Tool handler signature
func (s *MCPServer) handleCreatePersona(
    ctx context.Context,
    req *sdk.CallToolRequest,
    input CreatePersonaInput,
) (*sdk.CallToolResult, CreatePersonaOutput, error) {
    // Implementation
}

// 4. Server execution
func (s *MCPServer) Run(ctx context.Context) error {
    transport := &sdk.StdioTransport{}
    return s.server.Run(ctx, transport)
}
```

### SDK Features Used

| Feature | Usage in NEXS-MCP | File |
|---------|-------------------|------|
| **Server Creation** | `sdk.NewServer()` | `server.go:32` |
| **Tool Registration** | `sdk.AddTool()` | `server.go:145-548` |
| **Resource Registration** | `server.AddResource()` | `server.go:92-134` |
| **Stdio Transport** | `sdk.StdioTransport{}` | `server.go:546` |
| **Type Definitions** | `sdk.CallToolRequest`, `sdk.CallToolResult` | All tool handlers |
| **Content Types** | `sdk.TextContent`, `sdk.ImageContent` | Response building |

---

## MCP Protocol Overview

### What is MCP?

**Model Context Protocol (MCP)** is a standard protocol for LLMs to interact with external tools and resources. It provides:

- **Tools Protocol**: Register and invoke functions
- **Resources Protocol**: Expose data for context
- **Stdio Transport**: JSON-RPC over stdin/stdout
- **Type Safety**: Strongly-typed request/response

### MCP Message Flow

```
┌─────────────┐                     ┌─────────────┐
│ MCP Client  │                     │ MCP Server  │
│ (Claude)    │                     │ (NEXS)      │
└──────┬──────┘                     └──────┬──────┘
       │                                   │
       │  tools/call                       │
       │  {                                │
       │    "name": "create_element",      │
       │    "arguments": {...}             │
       │  }                                │
       ├──────────────────────────────────▶│
       │                                   │
       │                                   ├─ Parse Request
       │                                   ├─ Route to Handler
       │                                   ├─ Execute Logic
       │                                   ├─ Record Metrics
       │                                   │
       │  tools/call response              │
       │  {                                │
       │    "content": [...],              │
       │    "isError": false               │
       │  }                                │
       │◀──────────────────────────────────┤
       │                                   │
```

### JSON-RPC Format

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "create_element",
    "arguments": {
      "type": "persona",
      "name": "Technical Expert",
      "data": {...}
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Element created successfully: persona-technical-expert-1703088000"
      }
    ],
    "isError": false
  }
}
```

---

## Server Implementation

### MCPServer Structure

```go
type MCPServer struct {
    server             *sdk.Server
    repo               domain.ElementRepository
    metrics            *application.MetricsCollector
    perfMetrics        *logger.PerformanceMetrics
    index              *indexing.TFIDFIndex
    mu                 sync.Mutex
    deviceCodes        map[string]string
    capabilityResource *resources.CapabilityIndexResource
    resourcesConfig    config.ResourcesConfig
    cfg                *config.Config
}
```

### Server Initialization

```go
func NewMCPServer(
    name, version string,
    repo domain.ElementRepository,
    cfg *config.Config,
) *MCPServer {
    // Create SDK implementation
    impl := &sdk.Implementation{
        Name:    name,
        Version: version,
    }
    
    // Create server with stdio transport
    server := sdk.NewServer(impl, nil)
    
    // Create metrics collector
    metricsDir := filepath.Join(os.Getenv("HOME"), ".nexs-mcp", "metrics")
    metrics := application.NewMetricsCollector(metricsDir)
    
    // Create performance metrics
    perfDir := filepath.Join(os.Getenv("HOME"), ".nexs-mcp", "performance")
    perfMetrics := logger.NewPerformanceMetrics(perfDir)
    
    // Create TF-IDF index
    idx := indexing.NewTFIDFIndex()
    
    // Create capability resource
    capabilityResource := resources.NewCapabilityIndexResource(
        repo, idx, cfg.Resources.CacheTTL,
    )
    
    mcpServer := &MCPServer{
        server:             server,
        repo:               repo,
        metrics:            metrics,
        perfMetrics:        perfMetrics,
        index:              idx,
        capabilityResource: capabilityResource,
        resourcesConfig:    cfg.Resources,
        cfg:                cfg,
    }
    
    // Populate index with existing elements
    mcpServer.rebuildIndex()
    
    // Register all tools
    mcpServer.registerTools()
    
    // Register resources if enabled
    if cfg.Resources.Enabled {
        mcpServer.registerResources()
    }
    
    return mcpServer
}
```

### Server Lifecycle

```go
// Start server
func (s *MCPServer) Start(ctx context.Context) error {
    logger.Info("Starting NEXS MCP Server",
        "version", version,
        "tools", 55,
        "resources", s.resourcesConfig.Enabled)
    
    // Start the server (blocks until context cancelled)
    return s.server.Serve(ctx)
}

// Shutdown server
func (s *MCPServer) Shutdown(ctx context.Context) error {
    logger.Info("Shutting down NEXS MCP Server")
    
    // Save metrics
    if err := s.metrics.SaveMetrics(); err != nil {
        logger.Error("Failed to save metrics", "error", err)
    }
    
    // Save performance data
    if err := s.perfMetrics.SaveMetrics(); err != nil {
        logger.Error("Failed to save performance metrics", "error", err)
    }
    
    return nil
}
```

---

## Tool Registration

### Registration Pattern

Tools are registered using the official SDK:

```go
func (s *MCPServer) registerTools() {
    // Element CRUD
    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "create_element",
        Description: "Create a new element",
    }, s.handleCreateElement)
    
    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "get_element",
        Description: "Get a specific element by ID",
    }, s.handleGetElement)
    
    // ... 53 more tools
}
```

### Tool Handler Signature

```go
func (s *MCPServer) handleCreateElement(
    ctx context.Context,
    req *sdk.CallToolRequest,
    input CreateElementInput,
) (*sdk.CallToolResult, CreateElementOutput, error) {
    // Implementation
}
```

**Components:**
- **ctx**: Context for cancellation and deadlines
- **req**: Full MCP request with metadata
- **input**: Strongly-typed input parameters
- **Returns**:
  - `*sdk.CallToolResult`: MCP-formatted response
  - `Output`: Structured output data
  - `error`: Error if operation failed

---

## Element CRUD Tools

### 1. list_elements

List all elements with optional filtering:

```go
type ListElementsInput struct {
    Type     string   `json:"type,omitempty"`
    IsActive *bool    `json:"is_active,omitempty"`
    Tags     []string `json:"tags,omitempty"`
    Limit    int      `json:"limit,omitempty"`
    Offset   int      `json:"offset,omitempty"`
}

type ListElementsOutput struct {
    Elements []ElementSummary `json:"elements"`
    Total    int              `json:"total"`
    Filtered int              `json:"filtered"`
}

// Example usage
{
  "type": "persona",
  "is_active": true,
  "tags": ["expert"],
  "limit": 10
}
```

### 2. get_element

Retrieve element by ID:

```go
type GetElementInput struct {
    ID string `json:"id"`
}

type GetElementOutput struct {
    Element map[string]interface{} `json:"element"`
}
```

### 3. create_element

Generic element creation (use type-specific for full features):

```go
type CreateElementInput struct {
    Type        string                 `json:"type"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Version     string                 `json:"version"`
    Author      string                 `json:"author"`
    Tags        []string               `json:"tags,omitempty"`
    Data        map[string]interface{} `json:"data"`
}
```

### 4. update_element

Update existing element:

```go
type UpdateElementInput struct {
    ID   string                 `json:"id"`
    Data map[string]interface{} `json:"data"`
}
```

### 5. delete_element

Delete element by ID:

```go
type DeleteElementInput struct {
    ID string `json:"id"`
}
```

### 6. duplicate_element

Duplicate existing element with new ID:

```go
type DuplicateElementInput struct {
    ID      string `json:"id"`
    NewName string `json:"new_name,omitempty"`
}
```

---

## Quick Create Tools

Simplified creation with template defaults (no preview needed):

### 1. quick_create_persona

```go
type QuickCreatePersonaInput struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Expertise   []string `json:"expertise"`    // ["Go", "Architecture"]
    Tone        string   `json:"tone"`         // "Professional"
    Author      string   `json:"author"`
}

// Example
{
  "name": "Go Expert",
  "description": "Expert Go developer",
  "expertise": ["Go", "Concurrency", "Performance"],
  "tone": "Professional",
  "author": "team@company.com"
}
```

### 2. quick_create_skill

```go
type QuickCreateSkillInput struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Actions     []string `json:"actions"`      // List of action descriptions
    Author      string   `json:"author"`
}
```

### 3. quick_create_memory

```go
type QuickCreateMemoryInput struct {
    Name        string `json:"name"`
    Content     string `json:"content"`
    Description string `json:"description,omitempty"`
    Author      string `json:"author"`
}
```

### 4. quick_create_template

```go
type QuickCreateTemplateInput struct {
    Name        string   `json:"name"`
    Content     string   `json:"content"`
    Variables   []string `json:"variables"`    // ["name", "project"]
    Description string   `json:"description"`
    Author      string   `json:"author"`
}
```

### 5. quick_create_agent

```go
type QuickCreateAgentInput struct {
    Name        string   `json:"name"`
    Goals       []string `json:"goals"`
    Actions     []string `json:"actions"`
    Description string   `json:"description"`
    Author      string   `json:"author"`
}
```

### 6. quick_create_ensemble

```go
type QuickCreateEnsembleInput struct {
    Name              string   `json:"name"`
    AgentIDs          []string `json:"agent_ids"`
    ExecutionMode     string   `json:"execution_mode"`     // sequential/parallel/hybrid
    AggregationStrategy string `json:"aggregation_strategy"` // first/last/consensus
    Description       string   `json:"description"`
    Author            string   `json:"author"`
}
```

---

## GitHub Tools

### GitHub Sync Tools

#### 1. sync_github_portfolio

Full bidirectional sync:

```go
type SyncGitHubPortfolioInput struct {
    Owner      string `json:"owner"`       // GitHub username
    Repository string `json:"repository"`  // Repo name
    Branch     string `json:"branch"`      // Branch name (default: main)
    Direction  string `json:"direction"`   // push, pull, bidirectional
    DryRun     bool   `json:"dry_run"`     // Preview without executing
}

type SyncGitHubPortfolioOutput struct {
    Pushed    []string `json:"pushed"`
    Pulled    []string `json:"pulled"`
    Conflicts []string `json:"conflicts"`
    Skipped   []string `json:"skipped"`
    Summary   string   `json:"summary"`
}
```

#### 2. push_to_github

Push local elements to GitHub:

```go
type PushToGitHubInput struct {
    Owner       string   `json:"owner"`
    Repository  string   `json:"repository"`
    Branch      string   `json:"branch"`
    ElementIDs  []string `json:"element_ids,omitempty"` // Specific elements
    CommitMessage string `json:"commit_message,omitempty"`
}
```

#### 3. pull_from_github

Pull remote elements from GitHub:

```go
type PullFromGitHubInput struct {
    Owner      string `json:"owner"`
    Repository string `json:"repository"`
    Branch     string `json:"branch"`
    Pattern    string `json:"pattern,omitempty"` // File pattern to match
}
```

### GitHub Auth Tools

#### 4. github_auth_start

Initiate OAuth device flow:

```go
type GitHubAuthStartInput struct {
    // No parameters needed
}

type GitHubAuthStartOutput struct {
    DeviceCode      string `json:"device_code"`
    UserCode        string `json:"user_code"`
    VerificationURI string `json:"verification_uri"`
    ExpiresIn       int    `json:"expires_in"`
    Instructions    string `json:"instructions"`
}

// Response example
{
  "user_code": "ABCD-1234",
  "verification_uri": "https://github.com/login/device",
  "expires_in": 600,
  "instructions": "Visit https://github.com/login/device and enter code ABCD-1234"
}
```

#### 5. github_auth_complete

Complete OAuth authentication:

```go
type GitHubAuthCompleteInput struct {
    DeviceCode string `json:"device_code"`
}

type GitHubAuthCompleteOutput struct {
    Success  bool   `json:"success"`
    Username string `json:"username,omitempty"`
    Message  string `json:"message"`
}
```

#### 6. github_auth_status

Check authentication status:

```go
type GitHubAuthStatusInput struct {
    // No parameters
}

type GitHubAuthStatusOutput struct {
    Authenticated bool   `json:"authenticated"`
    Username      string `json:"username,omitempty"`
}
```

### GitHub Portfolio Tools

#### 7. create_github_portfolio

Create new portfolio repository:

```go
type CreateGitHubPortfolioInput struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Private     bool   `json:"private"`
}
```

#### 8. list_github_repositories

List user's repositories:

```go
type ListGitHubRepositoriesInput struct {
    // No parameters - lists all repos for authenticated user
}

type ListGitHubRepositoriesOutput struct {
    Repositories []GitHubRepo `json:"repositories"`
    Total        int          `json:"total"`
}
```

#### 9. clone_github_portfolio

Clone portfolio from GitHub:

```go
type CloneGitHubPortfolioInput struct {
    Owner      string `json:"owner"`
    Repository string `json:"repository"`
    Branch     string `json:"branch,omitempty"`
}
```

#### 10. delete_github_portfolio

Delete portfolio repository:

```go
type DeleteGitHubPortfolioInput struct {
    Owner      string `json:"owner"`
    Repository string `json:"repository"`
    Confirm    bool   `json:"confirm"` // Safety check
}
```

---

## Collection Tools

### Collection Management

#### 11. install_collection

Install collection from GitHub or local:

```go
type InstallCollectionInput struct {
    Source     string `json:"source"`      // github:owner/repo or local:path
    Version    string `json:"version,omitempty"`
    Namespace  string `json:"namespace,omitempty"`
    OverrideIDs bool  `json:"override_ids,omitempty"`
}

type InstallCollectionOutput struct {
    Installed  []string `json:"installed"`
    Skipped    []string `json:"skipped"`
    Errors     []string `json:"errors"`
    Total      int      `json:"total"`
    Collection string   `json:"collection"`
}
```

#### 12. list_installed_collections

```go
type ListInstalledCollectionsOutput struct {
    Collections []CollectionInfo `json:"collections"`
    Total       int              `json:"total"`
}

type CollectionInfo struct {
    Name        string    `json:"name"`
    Version     string    `json:"version"`
    Source      string    `json:"source"`
    ElementCount int      `json:"element_count"`
    InstalledAt time.Time `json:"installed_at"`
}
```

#### 13. uninstall_collection

```go
type UninstallCollectionInput struct {
    Name    string `json:"name"`
    Confirm bool   `json:"confirm"`
}
```

#### 14. update_collection

```go
type UpdateCollectionInput struct {
    Name    string `json:"name"`
    Version string `json:"version,omitempty"` // Latest if not specified
}
```

#### 15. publish_collection

```go
type PublishCollectionInput struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description"`
    ElementIDs  []string `json:"element_ids"`
    Author      string   `json:"author"`
    License     string   `json:"license,omitempty"`
    Repository  string   `json:"repository"`  // owner/repo
}
```

#### 16-20. More Collection Tools

- `search_collections` - Search available collections
- `validate_collection` - Validate collection manifest
- `export_collection` - Export as standalone package
- `import_collection` - Import from package
- `collection_dependencies` - Show dependency tree

---

## Memory Tools

### Memory Management

#### 21. search_memories

Search memories by content:

```go
type SearchMemoriesInput struct {
    Query      string   `json:"query"`
    Tags       []string `json:"tags,omitempty"`
    Limit      int      `json:"limit,omitempty"`
    DateFrom   string   `json:"date_from,omitempty"`
    DateTo     string   `json:"date_to,omitempty"`
}

type SearchMemoriesOutput struct {
    Memories []MemoryMatch `json:"memories"`
    Total    int           `json:"total"`
}

type MemoryMatch struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Content     string   `json:"content"`
    Relevance   float64  `json:"relevance"`
    Highlights  []string `json:"highlights"`
    DateCreated string   `json:"date_created"`
}
```

#### 22. summarize_memories

Summarize multiple memories:

```go
type SummarizeMemoriesInput struct {
    MemoryIDs   []string `json:"memory_ids,omitempty"`
    Tags        []string `json:"tags,omitempty"`
    DateFrom    string   `json:"date_from,omitempty"`
    DateTo      string   `json:"date_to,omitempty"`
    MaxLength   int      `json:"max_length,omitempty"`
}

type SummarizeMemoriesOutput struct {
    Summary     string `json:"summary"`
    MemoryCount int    `json:"memory_count"`
    Keywords    []string `json:"keywords"`
}
```

#### 23. deduplicate_memories

Find and remove duplicate memories:

```go
type DeduplicateMemoriesInput struct {
    DryRun bool `json:"dry_run"` // Preview only
}

type DeduplicateMemoriesOutput struct {
    Duplicates []DuplicateGroup `json:"duplicates"`
    Removed    []string         `json:"removed"`
    Kept       []string         `json:"kept"`
}
```

#### 24. update_memory

Update memory content:

```go
type UpdateMemoryInput struct {
    ID      string `json:"id"`
    Content string `json:"content"`
}
```

---

## Analytics Tools

### 25. get_analytics_dashboard

Comprehensive analytics:

```go
type GetAnalyticsDashboardInput struct {
    Period string `json:"period"` // hour, day, week, month, all
}

type GetAnalyticsDashboardOutput struct {
    UsageStatistics    UsageStatistics    `json:"usage_statistics"`
    PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
    ElementStatistics  ElementStatistics  `json:"element_statistics"`
    TopTools           []ToolUsageStat    `json:"top_tools"`
    RecentErrors       []ErrorSummary     `json:"recent_errors"`
    Trends             TrendData          `json:"trends"`
}

type UsageStatistics struct {
    TotalOperations int                `json:"total_operations"`
    SuccessRate     float64            `json:"success_rate"`
    OperationsByTool map[string]int    `json:"operations_by_tool"`
    ActiveUsers     []string           `json:"active_users"`
}

type PerformanceMetrics struct {
    P50Latency time.Duration `json:"p50_latency_ms"`
    P95Latency time.Duration `json:"p95_latency_ms"`
    P99Latency time.Duration `json:"p99_latency_ms"`
    AvgLatency time.Duration `json:"avg_latency_ms"`
}
```

---

## Utility Tools

### 26. auto_save_enable

Enable auto-save for conversation context:

```go
type AutoSaveEnableInput struct {
    Interval int `json:"interval"` // Save interval in seconds
}
```

### 27. auto_save_disable

Disable auto-save.

### 28. set_user_identity

Set current user identity:

```go
type SetUserIdentityInput struct {
    UserID   string                 `json:"user_id"`
    Name     string                 `json:"name,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

### 29. get_user_identity

Get current user identity.

### 30. query_logs

Query structured logs:

```go
type QueryLogsInput struct {
    Level     string `json:"level,omitempty"`      // info, warn, error
    Operation string `json:"operation,omitempty"`  // create, update, delete
    Tool      string `json:"tool,omitempty"`
    User      string `json:"user,omitempty"`
    FromTime  string `json:"from_time,omitempty"`
    ToTime    string `json:"to_time,omitempty"`
    Limit     int    `json:"limit,omitempty"`
}
```

### 31. get_performance_metrics

Get detailed performance metrics:

```go
type GetPerformanceMetricsInput struct {
    Tool   string `json:"tool,omitempty"`
    Period string `json:"period,omitempty"`
}

type GetPerformanceMetricsOutput struct {
    Metrics []ToolPerformanceMetric `json:"metrics"`
    Summary PerformanceSummary      `json:"summary"`
}
```

### 32. search

TF-IDF semantic search:

```go
type SearchInput struct {
    Query string `json:"query"`
    Limit int    `json:"limit,omitempty"`
}

type SearchOutput struct {
    Results []SearchResult `json:"results"`
    Total   int            `json:"total"`
}
```

### 33. rebuild_index

Rebuild search index from scratch.

### 34. get_index_stats

Get index statistics:

```go
type GetIndexStatsOutput struct {
    TotalDocuments int               `json:"total_documents"`
    DocumentsByType map[string]int   `json:"documents_by_type"`
    VocabularySize int               `json:"vocabulary_size"`
    TopTerms       []TermFrequency   `json:"top_terms"`
}
```

### 35-36. Ensemble Tools

- `execute_ensemble` - Execute ensemble with options
- `get_ensemble_status` - Get ensemble configuration and status

### 37. discover_element_types

Discover available element types and their schemas.

### 38. batch_create_elements

Create multiple elements in one operation:

```go
type BatchCreateElementsInput struct {
    Elements []CreateElementInput `json:"elements"`
}
```

### 39. render_template

Render template with values:

```go
type RenderTemplateInput struct {
    TemplateID string            `json:"template_id"`
    Values     map[string]string `json:"values"`
}
```

### 40. validate_element

Validate element without saving:

```go
type ValidateElementInput struct {
    Type string                 `json:"type"`
    Data map[string]interface{} `json:"data"`
}
```

### 41. reload_elements_cache

Reload file repository cache.

### 42-45. Template Tools

- `list_templates` - List all templates
- `get_template_variables` - Get template variables
- `preview_template` - Preview rendered template
- `validate_template` - Validate template syntax

### 46-47. Backup Tools

- `create_backup` - Create portfolio backup (tar.gz with SHA-256)
- `restore_backup` - Restore from backup

### 48-50. Collection Submission Tools

- `submit_to_collection` - Submit element via PR
- `track_submission_status` - Track PR status
- `list_submissions` - List all submissions

### 51-55. Additional Tools

- `activate_element` - Activate element
- `deactivate_element` - Deactivate element
- `get_element_history` - Get element change history
- `export_element` - Export element to file
- `import_element` - Import element from file

---

## Resources Protocol

### Capability Index Resources

NEXS MCP exposes three resources via the MCP Resources Protocol:

#### 1. nexs://capability-index/summary

Concise summary (~3K tokens):

```markdown
# NEXS Capability Index Summary

## Overview
Total Elements: 42
- Personas: 8
- Skills: 15
- Templates: 7
- Agents: 6
- Memories: 4
- Ensembles: 2

## Recent Elements
- persona-technical-expert-1703088000 (active)
- skill-code-review-1703088100 (active)
...

## Top Keywords
architecture, go, clean code, testing, performance...
```

#### 2. nexs://capability-index/full

Complete detailed view (~40K tokens):

```markdown
# NEXS Capability Index - Full Details

## Personas (8 elements)

### persona-technical-expert-1703088000
- **Name**: Technical Expert
- **Description**: Expert in software architecture
- **Tags**: architecture, expert, mentor
- **Expertise**: Software Architecture (expert), Go Programming (expert)
- **System Prompt**: [full prompt]
...
```

#### 3. nexs://capability-index/stats

Statistical data (JSON):

```json
{
  "total_elements": 42,
  "by_type": {
    "persona": 8,
    "skill": 15,
    ...
  },
  "index_stats": {
    "vocabulary_size": 1247,
    "total_documents": 42
  },
  "cache_stats": {
    "hits": 1523,
    "misses": 42,
    "hit_rate": 0.973
  }
}
```

### Resource Configuration

```bash
# Enable resources
nexs-mcp --resources-enabled=true

# Expose specific resources only
nexs-mcp --resources-enabled=true --resources-expose=summary,stats

# Disable resources (default)
nexs-mcp --resources-enabled=false
```

---

## Tool Handler Patterns

### Standard Tool Handler

```go
func (s *MCPServer) handleCreateElement(
    ctx context.Context,
    req *sdk.CallToolRequest,
    input CreateElementInput,
) (*sdk.CallToolResult, CreateElementOutput, error) {
    // 1. Record start time
    startTime := time.Now()
    
    // 2. Extract context
    userID := ExtractUserID(ctx)
    
    // 3. Validate input
    if input.Name == "" {
        return nil, output, fmt.Errorf("name is required")
    }
    
    // 4. Create domain entity
    var element domain.Element
    switch input.Type {
    case "persona":
        element = domain.NewPersona(input.Name, input.Description, input.Version, input.Author)
        // Populate fields from input.Data
    case "skill":
        element = domain.NewSkill(input.Name, input.Description, input.Version, input.Author)
    // ... other types
    }
    
    // 5. Validate domain rules
    if err := element.Validate(); err != nil {
        s.recordToolCall(startTime, "create_element", false, err)
        return nil, output, fmt.Errorf("validation failed: %w", err)
    }
    
    // 6. Save via repository
    if err := s.repo.Create(element); err != nil {
        s.recordToolCall(startTime, "create_element", false, err)
        return nil, output, fmt.Errorf("failed to create element: %w", err)
    }
    
    // 7. Update search index
    s.index.AddDocument(&indexing.Document{
        ID:      element.GetID(),
        Type:    element.GetType(),
        Name:    element.GetMetadata().Name,
        Content: extractContent(element),
    })
    
    // 8. Record metrics
    s.recordToolCall(startTime, "create_element", true, nil)
    
    // 9. Log operation
    logger.Info("Element created",
        "element_id", element.GetID(),
        "type", element.GetType(),
        "user", userID)
    
    // 10. Format output
    output = CreateElementOutput{
        ID:      element.GetID(),
        Type:    string(element.GetType()),
        Name:    element.GetMetadata().Name,
        Success: true,
    }
    
    // 11. Return MCP result
    return &sdk.CallToolResult{
        Content: []sdk.Content{{
            Type: "text",
            Text: fmt.Sprintf("Element %s created successfully", element.GetID()),
        }},
    }, output, nil
}
```

### Metrics Recording

```go
func (s *MCPServer) recordToolCall(
    startTime time.Time,
    toolName string,
    success bool,
    err error,
) {
    metric := application.ToolCallMetric{
        ToolName:  toolName,
        Timestamp: startTime,
        Duration:  time.Since(startTime),
        Success:   success,
    }
    
    if err != nil {
        metric.ErrorMessage = err.Error()
    }
    
    s.metrics.RecordToolCall(metric)
    
    // Also record performance metrics
    s.perfMetrics.RecordOperation(toolName, time.Since(startTime))
}
```

### Context Extraction

```go
func ExtractUserID(ctx context.Context) string {
    if userID, ok := ctx.Value("user_id").(string); ok {
        return userID
    }
    return "anonymous"
}

func ExtractRequestMetadata(req *sdk.CallToolRequest) map[string]interface{} {
    return map[string]interface{}{
        "tool_name": req.Params.Name,
        "timestamp": time.Now(),
    }
}
```

---

## Error Handling

### Error Response Pattern

```go
func (s *MCPServer) handleGetElement(
    ctx context.Context,
    req *sdk.CallToolRequest,
    input GetElementInput,
) (*sdk.CallToolResult, GetElementOutput, error) {
    element, err := s.repo.GetByID(input.ID)
    
    if errors.Is(err, domain.ErrElementNotFound) {
        return &sdk.CallToolResult{
            Content: []sdk.Content{{
                Type: "text",
                Text: fmt.Sprintf("Element with ID %s not found", input.ID),
            }},
            IsError: true,
        }, output, nil // Return nil error, mark as MCP error
    }
    
    if err != nil {
        return nil, output, fmt.Errorf("failed to get element: %w", err)
    }
    
    // Success path...
}
```

### Error Categories

| Category | HTTP Status | MCP Handling |
|----------|-------------|--------------|
| **Not Found** | 404 | IsError=true, nil error |
| **Validation** | 400 | IsError=true, nil error |
| **Unauthorized** | 401 | IsError=true, nil error |
| **Server Error** | 500 | Return error |
| **Network Error** | 503 | Return error |

---

## Performance Optimization

### Caching Strategy

```go
// Cache frequently accessed elements
type ElementCache struct {
    mu    sync.RWMutex
    cache map[string]domain.Element
    ttl   time.Duration
}

func (c *ElementCache) Get(id string) (domain.Element, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    elem, ok := c.cache[id]
    return elem, ok
}
```

### Batch Operations

```go
func (s *MCPServer) handleBatchCreateElements(
    ctx context.Context,
    req *sdk.CallToolRequest,
    input BatchCreateElementsInput,
) (*sdk.CallToolResult, BatchCreateElementsOutput, error) {
    // Process all elements in single transaction
    created := make([]string, 0, len(input.Elements))
    failed := make([]string, 0)
    
    for _, elemInput := range input.Elements {
        element, err := s.createElement(elemInput)
        if err != nil {
            failed = append(failed, elemInput.Name)
            continue
        }
        
        if err := s.repo.Create(element); err != nil {
            failed = append(failed, elemInput.Name)
            continue
        }
        
        created = append(created, element.GetID())
    }
    
    // Single index rebuild for all
    s.rebuildIndex()
    
    output := BatchCreateElementsOutput{
        Created: created,
        Failed:  failed,
        Total:   len(input.Elements),
    }
    
    return formatMCPResponse(output), output, nil
}
```

### Connection Pooling

```go
// Reuse HTTP clients
var githubClientPool = sync.Pool{
    New: func() interface{} {
        return &http.Client{
            Timeout: 30 * time.Second,
        }
    },
}
```

---

## Best Practices

### 1. Validate Early

```go
// ✅ Good: Validate at entry point
func (s *MCPServer) handleCreateElement(...) {
    if input.Name == "" {
        return nil, output, fmt.Errorf("name is required")
    }
    
    // Domain validation
    if err := element.Validate(); err != nil {
        return nil, output, err
    }
    
    // Continue processing
}
```

### 2. Record All Operations

```go
// ✅ Good: Record metrics for every tool call
defer s.recordToolCall(startTime, toolName, success, err)
```

### 3. Structured Logging

```go
// ✅ Good: Structured context
logger.Info("Element created",
    "element_id", element.GetID(),
    "type", element.GetType(),
    "user", userID,
    "duration_ms", time.Since(startTime).Milliseconds())
```

### 4. Graceful Degradation

```go
// ✅ Good: Degrade gracefully on non-critical errors
if err := s.metrics.RecordToolCall(metric); err != nil {
    logger.Warn("Failed to record metrics", "error", err)
    // Continue processing
}
```

### 5. Context Propagation

```go
// ✅ Good: Pass context through layers
func (s *MCPServer) handleCreateElement(ctx context.Context, ...) {
    element, err := s.repo.Create(ctx, element)
}
```

---

## Conclusion

The MCP Layer provides a comprehensive interface to NEXS MCP Server with 55 tools covering element management, GitHub integration, analytics, and production features. Built on the official MCP SDK v1.1.0, it provides type-safe, high-performance access to all system capabilities.

**Key Features:**
- 55 production-ready tools
- Type-safe request/response
- Comprehensive metrics collection
- MCP Resources Protocol support
- Error handling and logging
- Performance optimization

**Tool Categories:**
- Element CRUD (6 tools)
- Quick Create (6 tools)
- GitHub Integration (10 tools)
- Collection Management (10 tools)
- Memory Management (4 tools)
- Analytics (2 tools)
- Utilities (17 tools)

---

**Document Version:** 1.0.0  
**Total Lines:** 1267  
**Last Updated:** December 20, 2025
