# Code Tour: Complete Walkthrough of NEXS-MCP

## Table of Contents

- [Introduction](#introduction)
- [Entry Point: cmd/nexs-mcp/main.go](#entry-point-cmdnexs-mcpmaingo)
- [Configuration: internal/config/](#configuration-internalconfig)
- [Logger: internal/logger/](#logger-internallogger)
- [Repository Setup: internal/infrastructure/](#repository-setup-internalinfrastructure)
- [MCP Server: internal/mcp/](#mcp-server-internalmcp)
- [Domain Models: internal/domain/](#domain-models-internaldomain)
- [Validation: internal/validation/](#validation-internalvalidation)
- [Application Services: internal/application/](#application-services-internalapplication)
- [Indexing: internal/indexing/](#indexing-internalindexing)
- [Templates: internal/template/](#templates-internaltemplate)
- [Portfolio: internal/portfolio/](#portfolio-internalportfolio)
- [Collections: internal/collection/](#collections-internalcollection)
- [Backup: internal/backup/](#backup-internalbackup)
- [Data Flow Diagrams](#data-flow-diagrams)
- [Key Interfaces](#key-interfaces)
- [Request/Response Flow](#requestresponse-flow)
- [Quick Reference Guide](#quick-reference-guide)

---

## Introduction

NEXS-MCP (Next-generation Extensible System for Model Context Protocol) is a sophisticated MCP server implementation in Go that manages reusable AI capability elements. This code tour provides a comprehensive walkthrough of the entire codebase.

**Key Technologies:**
- Go 1.25+
- Official MCP Go SDK (`github.com/modelcontextprotocol/go-sdk/mcp`)
- YAML for element storage
- TF-IDF for semantic search
- Structured logging with slog

**Architecture Overview:**
```
┌─────────────────────────────────────────────────────┐
│              MCP Client (Claude Desktop)            │
└────────────────────┬────────────────────────────────┘
                     │ JSON-RPC over stdio
                     ▼
┌─────────────────────────────────────────────────────┐
│           MCP Server (Official SDK)                 │
│  ┌──────────────────────────────────────────────┐  │
│  │    66 Tools + 3 Resources Registered         │  │
│  └──────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│              Business Logic Layer                   │
│  ┌──────────┬──────────┬──────────┬──────────┐     │
│  │ Domain   │Validation│Templates │  Index   │     │
│  │ Models   │  Rules   │ Engine   │  Search  │     │
│  └──────────┴──────────┴──────────┴──────────┘     │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│           Data Persistence Layer                    │
│  ┌──────────────────┬──────────────────────────┐   │
│  │ File Repository  │  In-Memory Repository    │   │
│  │  (YAML files)    │     (for testing)        │   │
│  └──────────────────┴──────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

---

## Entry Point: cmd/nexs-mcp/main.go

**Location:** `cmd/nexs-mcp/main.go` (123 lines)

### Purpose
The entry point initializes all components and starts the MCP server.

### Key Functions

#### `main()`
```go
func main() {
    // Setup context with cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        logger.Info("Shutdown signal received, gracefully shutting down...")
        cancel()
    }()

    // Initialize and run server
    if err := run(ctx); err != nil {
        logger.Error("Server error", "error", err)
        os.Exit(1)
    }
}
```

**What it does:**
1. Creates a context for graceful shutdown
2. Sets up signal handlers for CTRL+C and SIGTERM
3. Calls `run()` to initialize and start the server
4. Handles shutdown gracefully

#### `run(ctx context.Context)`
```go
func run(ctx context.Context) error {
    // 1. Load configuration
    cfg := config.LoadConfig(version)

    // 2. Initialize logger with buffer
    logCfg := &logger.Config{
        Level:     parseLogLevel(cfg.LogLevel),
        Format:    cfg.LogFormat,
        Output:    os.Stderr,
        AddSource: false,
    }
    logger.InitWithBuffer(logCfg, 1000)

    // 3. Create repository based on configuration
    var repo domain.ElementRepository
    switch cfg.StorageType {
    case "file":
        repo, err = infrastructure.NewFileElementRepository(cfg.DataDir)
    case "memory":
        repo = infrastructure.NewInMemoryElementRepository()
    }

    // 4. Create MCP server using official SDK
    server := mcp.NewMCPServer(cfg.ServerName, cfg.Version, repo, cfg)

    // 5. Start server with stdio transport
    return server.Run(ctx)
}
```

**Initialization Flow:**
```
LoadConfig → InitLogger → CreateRepository → CreateMCPServer → Run
```

### Configuration Sources
1. Environment variables (`NEXS_*`)
2. Command-line flags (override env vars)
3. Defaults (fallback values)

**Example:**
```bash
# Environment variable
export NEXS_STORAGE_TYPE=file
export NEXS_DATA_DIR=/custom/path
export NEXS_LOG_LEVEL=debug

# Or command-line flags
./nexs-mcp --storage=file --data-dir=/custom/path --log-level=debug
```

---

## Configuration: internal/config/

**Location:** `internal/config/config.go` (123 lines)

### Purpose
Manages all server configuration from environment variables and command-line flags.

### Config Structure

```go
type Config struct {
    // Storage settings
    StorageType string  // "memory" or "file"
    DataDir     string  // Directory for file storage

    // Server identity
    ServerName string  // MCP server name
    Version    string  // Application version

    // Logging
    LogLevel  string  // "debug", "info", "warn", "error"
    LogFormat string  // "json" or "text"

    // Auto-save feature
    AutoSaveMemories bool           // Enable conversation auto-save
    AutoSaveInterval time.Duration  // Minimum time between saves

    // MCP Resources Protocol
    Resources ResourcesConfig
}

type ResourcesConfig struct {
    Enabled  bool          // Enable/disable resources
    Expose   []string      // Which resources to expose
    CacheTTL time.Duration // Cache duration
}
```

### Loading Process

```go
func LoadConfig(version string) *Config {
    cfg := &Config{
        ServerName:       getEnvOrDefault("NEXS_SERVER_NAME", "nexs-mcp"),
        Version:          version,
        AutoSaveMemories: getEnvBool("NEXS_AUTO_SAVE_MEMORIES", true),
        AutoSaveInterval: getEnvDuration("NEXS_AUTO_SAVE_INTERVAL", 5*time.Minute),
        Resources: ResourcesConfig{
            Enabled:  getEnvBool("NEXS_RESOURCES_ENABLED", false),
            Expose:   []string{},
            CacheTTL: getEnvDuration("NEXS_RESOURCES_CACHE_TTL", 5*time.Minute),
        },
    }

    // Define command-line flags
    flag.StringVar(&cfg.StorageType, "storage", getEnvOrDefault("NEXS_STORAGE_TYPE", "file"), ...)
    flag.StringVar(&cfg.DataDir, "data-dir", getEnvOrDefault("NEXS_DATA_DIR", "data/elements"), ...)
    // ... more flags

    flag.Parse()
    return cfg
}
```

**Priority:** Command-line flags > Environment variables > Defaults

### Environment Variables Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `NEXS_STORAGE_TYPE` | string | `file` | Storage backend (memory/file) |
| `NEXS_DATA_DIR` | string | `data/elements` | Directory for file storage |
| `NEXS_SERVER_NAME` | string | `nexs-mcp` | MCP server name |
| `NEXS_LOG_LEVEL` | string | `info` | Logging level |
| `NEXS_LOG_FORMAT` | string | `json` | Log format (json/text) |
| `NEXS_AUTO_SAVE_MEMORIES` | bool | `true` | Enable auto-save |
| `NEXS_AUTO_SAVE_INTERVAL` | duration | `5m` | Auto-save interval |
| `NEXS_RESOURCES_ENABLED` | bool | `false` | Enable MCP Resources |
| `NEXS_RESOURCES_CACHE_TTL` | duration | `5m` | Resource cache TTL |

---

## Logger: internal/logger/

**Files:**
- `logger.go` - Core logger implementation
- `buffer.go` - Circular buffer for log history
- `performance.go` - Performance metrics tracking

### Purpose
Provides structured logging with buffering and performance tracking.

### Logger Structure

```go
var (
    defaultLogger *slog.Logger
    logBuffer     *CircularBuffer
    perfMetrics   *PerformanceMetrics
)

type Config struct {
    Level     slog.Level
    Format    string        // "json" or "text"
    Output    io.Writer
    AddSource bool
}
```

### Initialization

```go
func InitWithBuffer(cfg *Config, bufferSize int) {
    // Create handler based on format
    var handler slog.Handler
    if cfg.Format == "text" {
        handler = slog.NewTextHandler(cfg.Output, &slog.HandlerOptions{
            Level:     cfg.Level,
            AddSource: cfg.AddSource,
        })
    } else {
        handler = slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
            Level:     cfg.Level,
            AddSource: cfg.AddSource,
        })
    }

    // Create logger
    defaultLogger = slog.New(handler)
    slog.SetDefault(defaultLogger)

    // Initialize circular buffer
    logBuffer = NewCircularBuffer(bufferSize)
}
```

### Usage Examples

```go
// Simple logging
logger.Info("Server started", "port", 8080)
logger.Error("Failed to connect", "error", err)
logger.Debug("Processing request", "id", requestID)

// With context
logger.InfoContext(ctx, "User authenticated", "user", username)

// Structured fields
logger.Info("Operation completed",
    "operation", "create_element",
    "type", "persona",
    "duration_ms", duration,
    "success", true)
```

### Circular Buffer

The circular buffer stores the last N log entries for retrieval via the `get_logs` tool:

```go
type CircularBuffer struct {
    entries []LogEntry
    size    int
    index   int
    mu      sync.RWMutex
}

type LogEntry struct {
    Level     string
    Message   string
    Time      time.Time
    Attrs     map[string]interface{}
}
```

**Retrieval:**
```go
// Get last 100 logs at info level or higher
logs := logBuffer.GetLogs(100, "info")
```

### Performance Metrics

```go
type PerformanceMetrics struct {
    dir     string
    mu      sync.RWMutex
    entries []PerformanceEntry
}

type PerformanceEntry struct {
    Timestamp   time.Time
    Operation   string
    Duration    time.Duration
    Success     bool
    ErrorMsg    string
    Metadata    map[string]interface{}
}
```

**Recording metrics:**
```go
start := time.Now()
// ... perform operation ...
duration := time.Since(start)

perfMetrics.Record(PerformanceEntry{
    Timestamp: start,
    Operation: "create_element",
    Duration:  duration,
    Success:   err == nil,
    ErrorMsg:  errString,
    Metadata: map[string]interface{}{
        "element_type": "persona",
        "element_id": id,
    },
})
```

---

## Repository Setup: internal/infrastructure/

**Files:**
- `file_repository.go` - File-based YAML storage (359 lines)
- `memory_repository.go` - In-memory storage for testing (241 lines)

### Purpose
Implements the `domain.ElementRepository` interface for data persistence.

### Repository Interface

```go
type ElementRepository interface {
    Create(element Element) error
    Get(id string) (Element, error)
    Update(element Element) error
    Delete(id string) error
    List(filter ElementFilter) ([]Element, error)
    
    // Batch operations
    CreateBatch(elements []Element) error
    GetBatch(ids []string) ([]Element, error)
    
    // Special operations
    Duplicate(id string, newName string) (Element, error)
    GetByType(elementType ElementType) ([]Element, error)
}
```

### File Repository Implementation

**Storage Structure:**
```
data/elements/
├── persona/
│   ├── 2025-12-01/
│   │   ├── persona-001.yaml
│   │   └── persona-002.yaml
│   └── 2025-12-02/
│       └── persona-003.yaml
├── skill/
│   └── 2025-12-01/
│       ├── skill-001.yaml
│       └── skill-002.yaml
└── template/
    └── 2025-12-01/
        └── template-001.yaml
```

**File Format (YAML):**
```yaml
metadata:
  id: persona-001
  type: persona
  name: "Senior Software Engineer"
  description: "Expert in system design and architecture"
  version: "1.0.0"
  author: "team@example.com"
  tags:
    - engineering
    - architecture
  is_active: true
  created_at: 2025-12-01T10:00:00Z
  updated_at: 2025-12-01T10:00:00Z
data:
  role: "Senior Software Engineer"
  tone: "professional"
  expertise:
    - "system design"
    - "microservices"
  context: "You are an experienced software engineer..."
```

**Key Implementation Details:**

```go
type FileElementRepository struct {
    mu      sync.RWMutex
    baseDir string
    cache   map[string]*StoredElement // In-memory cache
}

func (r *FileElementRepository) Create(element domain.Element) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    metadata := element.GetMetadata()
    
    // Check cache for duplicates
    if _, exists := r.cache[metadata.ID]; exists {
        return fmt.Errorf("element with ID %s already exists", metadata.ID)
    }

    // Serialize to YAML
    stored := &StoredElement{
        Metadata: metadata,
        Data:     elementToMap(element),
    }
    
    data, err := yaml.Marshal(stored)
    if err != nil {
        return fmt.Errorf("failed to marshal element: %w", err)
    }

    // Write to file
    filePath := r.getFilePath(metadata)
    if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
        return err
    }
    
    if err := os.WriteFile(filePath, data, 0644); err != nil {
        return err
    }

    // Update cache
    r.cache[metadata.ID] = stored
    return nil
}
```

**Cache Strategy:**
- All elements loaded into memory on startup
- Cache updated on all write operations
- Read operations served from cache (fast)
- File system is source of truth

### Memory Repository Implementation

Used for testing and development:

```go
type InMemoryElementRepository struct {
    mu       sync.RWMutex
    elements map[string]domain.Element
}

func (r *InMemoryElementRepository) Create(element domain.Element) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    id := element.GetID()
    if _, exists := r.elements[id]; exists {
        return fmt.Errorf("element with ID %s already exists", id)
    }

    r.elements[id] = element
    return nil
}
```

**Benefits:**
- Fast (no disk I/O)
- No cleanup needed
- Perfect for unit tests
- Simple implementation

---

## MCP Server: internal/mcp/

**Core Files:**
- `server.go` - Server initialization and tool registration (548 lines)
- `tools.go` - Tool input/output schemas (494 lines)
- `type_specific_handlers.go` - Element-specific CRUD (651 lines)
- `quick_create_tools.go` - Quick creation tools (339 lines)
- `search_tool.go` - Semantic search (241 lines)
- `memory_tools.go` - Memory management (323 lines)
- `template_tools.go` - Template operations (267 lines)
- `ensemble_execution_tools.go` - Ensemble execution (418 lines)
- `batch_tools.go` - Batch operations (312 lines)
- `index_tools.go` - Index management (189 lines)
- `log_tools.go` - Log retrieval (152 lines)
- `performance_tools.go` - Performance metrics (198 lines)
- `github_tools.go` - GitHub integration (489 lines)
- `github_auth_tools.go` - GitHub OAuth (387 lines)
- `collection_tools.go` - Collection management (623 lines)
- `backup_tools.go` - Backup/restore (298 lines)

### Server Structure

```go
type MCPServer struct {
    server             *sdk.Server  // Official MCP SDK server
    repo               domain.ElementRepository
    metrics            *application.MetricsCollector
    perfMetrics        *logger.PerformanceMetrics
    index              *indexing.TFIDFIndex
    mu                 sync.Mutex
    deviceCodes        map[string]string // GitHub OAuth
    capabilityResource *resources.CapabilityIndexResource
    resourcesConfig    config.ResourcesConfig
    cfg                *config.Config
}
```

### Initialization Flow

```go
func NewMCPServer(name, version string, repo domain.ElementRepository, cfg *config.Config) *MCPServer {
    // 1. Create SDK implementation
    impl := &sdk.Implementation{
        Name:    name,
        Version: version,
    }

    // 2. Create server with default capabilities
    server := sdk.NewServer(impl, nil)

    // 3. Create metrics collector
    metricsDir := filepath.Join(os.Getenv("HOME"), ".nexs-mcp", "metrics")
    metrics := application.NewMetricsCollector(metricsDir)

    // 4. Create performance metrics
    perfDir := filepath.Join(os.Getenv("HOME"), ".nexs-mcp", "performance")
    perfMetrics := logger.NewPerformanceMetrics(perfDir)

    // 5. Create TF-IDF index
    idx := indexing.NewTFIDFIndex()

    // 6. Create capability index resource
    capabilityResource := resources.NewCapabilityIndexResource(repo, idx, cfg.Resources.CacheTTL)

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

    // 7. Populate index with existing elements
    mcpServer.rebuildIndex()

    // 8. Register all tools
    mcpServer.registerTools()

    // 9. Register resources if enabled
    if cfg.Resources.Enabled {
        mcpServer.registerResources()
    }

    return mcpServer
}
```

### Tool Registration

NEXS-MCP uses the official MCP Go SDK for tool registration:

```go
func (s *MCPServer) registerTools() {
    // Basic CRUD operations
    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "list_elements",
        Description: "List all elements with optional filtering",
    }, s.handleListElements)

    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "get_element",
        Description: "Get a specific element by ID",
    }, s.handleGetElement)

    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "create_element",
        Description: "Create a new element",
    }, s.handleCreateElement)

    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "update_element",
        Description: "Update an existing element",
    }, s.handleUpdateElement)

    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "delete_element",
        Description: "Delete an element",
    }, s.handleDeleteElement)

    // Type-specific quick creation tools
    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "quick_create_persona",
        Description: "Quickly create a persona with minimal input",
    }, s.handleQuickCreatePersona)

    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "quick_create_skill",
        Description: "Quickly create a skill with minimal input",
    }, s.handleQuickCreateSkill)

    // ... 48 more tools registered similarly
}
```

**SDK Tool Structure:**
```go
type Tool struct {
    Name        string      // Tool identifier
    Description string      // Human-readable description
    InputSchema interface{} // JSON Schema (auto-generated from input struct)
}
```

### Handler Pattern

All tool handlers follow a consistent pattern:

```go
func (s *MCPServer) handleListElements(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    // 1. Start timing
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        s.metrics.RecordOperation("list_elements", duration, err == nil)
    }()

    // 2. Parse arguments
    var input ListElementsInput
    if err := parseArguments(arguments, &input); err != nil {
        return errorResult("Invalid arguments: %v", err), nil
    }

    // 3. Build filter
    filter := domain.ElementFilter{
        Type:     domain.ElementType(input.Type),
        IsActive: input.IsActive,
        Tags:     parseTags(input.Tags),
    }

    // 4. Query repository
    elements, err := s.repo.List(filter)
    if err != nil {
        return errorResult("Failed to list elements: %v", err), nil
    }

    // 5. Check access control
    if input.User != "" {
        elements = filterByAccess(elements, input.User)
    }

    // 6. Convert to maps
    result := make([]map[string]interface{}, len(elements))
    for i, elem := range elements {
        result[i] = elementToMap(elem)
    }

    // 7. Build response
    output := ListElementsOutput{
        Elements: result,
        Total:    len(result),
    }

    // 8. Return success
    return successResult(output)
}
```

**Error Handling:**
```go
func errorResult(format string, args ...interface{}) *sdk.CallToolResult {
    return &sdk.CallToolResult{
        IsError: true,
        Content: []sdk.Content{{
            Type: "text",
            Text: fmt.Sprintf(format, args...),
        }},
    }
}

func successResult(data interface{}) (*sdk.CallToolResult, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return errorResult("Failed to marshal result: %v", err), nil
    }

    return &sdk.CallToolResult{
        Content: []sdk.Content{{
            Type: "text",
            Text: string(jsonData),
        }},
    }, nil
}
```

### Resource Registration

MCP Resources provide read-only context to clients:

```go
func (s *MCPServer) registerResources() {
    handler := s.capabilityResource.Handler()

    // Summary resource (~3K tokens)
    s.server.AddResource(&sdk.Resource{
        URI:         "capability-index://summary",
        Name:        "Capability Index Summary",
        Description: "Concise summary of available capabilities",
        MIMEType:    "text/markdown",
    }, handler)

    // Full details resource (~40K tokens)
    s.server.AddResource(&sdk.Resource{
        URI:         "capability-index://full",
        Name:        "Capability Index Full Details",
        Description: "Complete capability index with all details",
        MIMEType:    "text/markdown",
    }, handler)

    // Statistics resource (JSON)
    s.server.AddResource(&sdk.Resource{
        URI:         "capability-index://stats",
        Name:        "Capability Index Statistics",
        Description: "Statistical data about the capability index",
        MIMEType:    "application/json",
    }, handler)
}
```

### Running the Server

```go
func (s *MCPServer) Run(ctx context.Context) error {
    // Use stdio transport (standard for MCP)
    transport := &sdk.StdioTransport{}
    
    // Start server
    return s.server.Run(ctx, transport)
}
```

**Communication Flow:**
```
Client ──JSON-RPC──> Stdio Transport ──> SDK Server ──> Tool Handler ──> Business Logic
   ^                                                                              │
   │                                                                              │
   └────────────────────────────── JSON Response ────────────────────────────────┘
```

---

## Domain Models: internal/domain/

**Files:**
- `element.go` - Base element interface and metadata (154 lines)
- `persona.go` - Persona element (157 lines)
- `skill.go` - Skill element (143 lines)
- `template.go` - Template element (185 lines)
- `agent.go` - Agent element (189 lines)
- `memory.go` - Memory element (167 lines)
- `ensemble.go` - Ensemble element (245 lines)
- `access_control.go` - Access control system (198 lines)

### Element Type Hierarchy

```
┌─────────────────────────────────────────────────────┐
│                   Element (interface)               │
│  - GetMetadata() ElementMetadata                    │
│  - Validate() error                                 │
│  - GetType() ElementType                            │
│  - GetID() string                                   │
│  - IsActive() bool                                  │
│  - Activate() error                                 │
│  - Deactivate() error                               │
└───────────────────┬─────────────────────────────────┘
                    │
        ┌───────────┴───────────┬────────────────┬─────────────┐
        │                       │                │             │
    ┌───▼────┐            ┌────▼────┐      ┌────▼────┐   ┌───▼────┐
    │Persona │            │ Skill   │      │Template │   │ Agent  │
    └────────┘            └─────────┘      └─────────┘   └────────┘
        │                       │                │             │
    ┌───▼────┐            ┌────▼────────────────▼─────────────▼────┐
    │ Memory │            │           Ensemble                      │
    └────────┘            │  (Orchestrates multiple elements)       │
                          └─────────────────────────────────────────┘
```

### Base Element Interface

```go
type Element interface {
    GetMetadata() ElementMetadata
    Validate() error
    GetType() ElementType
    GetID() string
    IsActive() bool
    Activate() error
    Deactivate() error
}
```

### ElementMetadata

Common metadata for all elements:

```go
type ElementMetadata struct {
    ID          string                 `json:"id" validate:"required"`
    Type        ElementType            `json:"type" validate:"required,oneof=persona skill template agent memory ensemble"`
    Name        string                 `json:"name" validate:"required,min=3,max=100"`
    Description string                 `json:"description" validate:"max=500"`
    Version     string                 `json:"version" validate:"required,semver"`
    Author      string                 `json:"author" validate:"required"`
    Tags        []string               `json:"tags,omitempty"`
    IsActive    bool                   `json:"is_active"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
    Custom      map[string]interface{} `json:"custom,omitempty"`
}
```

### Persona Element

Represents an AI persona/character:

```go
type Persona struct {
    Metadata ElementMetadata `json:"metadata"`
    
    // Persona-specific fields
    Role           string            `json:"role" validate:"required"`
    Tone           string            `json:"tone" validate:"required"`
    Expertise      []string          `json:"expertise" validate:"required,min=1"`
    Context        string            `json:"context" validate:"required"`
    Constraints    []string          `json:"constraints,omitempty"`
    Examples       []PersonaExample  `json:"examples,omitempty"`
    AccessControl  *AccessControl    `json:"access_control,omitempty"`
}

type PersonaExample struct {
    UserMessage      string `json:"user_message"`
    AssistantMessage string `json:"assistant_message"`
}

func (p *Persona) Validate() error {
    if err := p.Metadata.Validate(); err != nil {
        return err
    }
    
    if p.Role == "" {
        return fmt.Errorf("role is required")
    }
    
    if p.Tone == "" {
        return fmt.Errorf("tone is required")
    }
    
    if len(p.Expertise) == 0 {
        return fmt.Errorf("at least one expertise area is required")
    }
    
    if p.Context == "" {
        return fmt.Errorf("context is required")
    }
    
    return nil
}
```

### Skill Element

Represents a specific capability:

```go
type Skill struct {
    Metadata ElementMetadata `json:"metadata"`
    
    // Skill-specific fields
    Category      string            `json:"category" validate:"required"`
    InputSchema   map[string]interface{} `json:"input_schema" validate:"required"`
    OutputSchema  map[string]interface{} `json:"output_schema" validate:"required"`
    Instructions  string            `json:"instructions" validate:"required"`
    Examples      []SkillExample    `json:"examples" validate:"min=1"`
    Dependencies  []string          `json:"dependencies,omitempty"`
    Complexity    string            `json:"complexity" validate:"oneof=simple medium complex"`
    AccessControl *AccessControl    `json:"access_control,omitempty"`
}

type SkillExample struct {
    Input       map[string]interface{} `json:"input"`
    Output      map[string]interface{} `json:"output"`
    Description string                 `json:"description"`
}
```

### Template Element

Represents a content template:

```go
type Template struct {
    Metadata ElementMetadata `json:"metadata"`
    
    // Template-specific fields
    TemplateType string                 `json:"template_type" validate:"required,oneof=handlebars jinja2 go-template"`
    Content      string                 `json:"content" validate:"required"`
    Variables    []TemplateVariable     `json:"variables" validate:"required"`
    Partials     map[string]string      `json:"partials,omitempty"`
    Helpers      []string               `json:"helpers,omitempty"`
    Examples     []TemplateExample      `json:"examples" validate:"min=1"`
    AccessControl *AccessControl        `json:"access_control,omitempty"`
}

type TemplateVariable struct {
    Name        string `json:"name" validate:"required"`
    Type        string `json:"type" validate:"required,oneof=string number boolean array object"`
    Required    bool   `json:"required"`
    Default     interface{} `json:"default,omitempty"`
    Description string `json:"description"`
}
```

### Agent Element

Represents an autonomous agent:

```go
type Agent struct {
    Metadata ElementMetadata `json:"metadata"`
    
    // Agent-specific fields
    PersonaID     string            `json:"persona_id" validate:"required"`
    SkillIDs      []string          `json:"skill_ids" validate:"required,min=1"`
    WorkflowSteps []WorkflowStep    `json:"workflow_steps" validate:"required,min=1"`
    MaxIterations int               `json:"max_iterations" validate:"required,min=1,max=100"`
    Timeout       int               `json:"timeout" validate:"required,min=1"`
    AccessControl *AccessControl    `json:"access_control,omitempty"`
}

type WorkflowStep struct {
    Name        string                 `json:"name" validate:"required"`
    SkillID     string                 `json:"skill_id" validate:"required"`
    InputMap    map[string]string      `json:"input_map"`
    Condition   string                 `json:"condition,omitempty"`
    OnSuccess   string                 `json:"on_success,omitempty"`
    OnFailure   string                 `json:"on_failure,omitempty"`
}
```

### Memory Element

Represents stored context/knowledge:

```go
type Memory struct {
    Metadata ElementMetadata `json:"metadata"`
    
    // Memory-specific fields
    ContentType   string            `json:"content_type" validate:"required,oneof=conversation fact knowledge insight"`
    Content       string            `json:"content" validate:"required"`
    Source        string            `json:"source"`
    Importance    int               `json:"importance" validate:"min=1,max=10"`
    Timestamp     time.Time         `json:"timestamp"`
    References    []string          `json:"references,omitempty"`
    Keywords      []string          `json:"keywords,omitempty"`
    AccessControl *AccessControl    `json:"access_control,omitempty"`
}
```

### Ensemble Element

Orchestrates multiple elements:

```go
type Ensemble struct {
    Metadata ElementMetadata `json:"metadata"`
    
    // Ensemble-specific fields
    Strategy      string            `json:"strategy" validate:"required,oneof=sequential parallel consensus voting"`
    ElementIDs    []string          `json:"element_ids" validate:"required,min=2"`
    Configuration map[string]interface{} `json:"configuration"`
    AggregationMethod string        `json:"aggregation_method,omitempty" validate:"omitempty,oneof=merge select best consensus weighted"`
    MinConsensus  float64           `json:"min_consensus,omitempty" validate:"omitempty,min=0,max=1"`
    Timeout       int               `json:"timeout" validate:"required,min=1"`
    AccessControl *AccessControl    `json:"access_control,omitempty"`
}
```

### Access Control

```go
type AccessControl struct {
    Owner       string   `json:"owner"`
    Permissions []string `json:"permissions" validate:"required"`
    AllowedUsers []string `json:"allowed_users,omitempty"`
    AllowedGroups []string `json:"allowed_groups,omitempty"`
}

func (ac *AccessControl) CanAccess(user string, action string) bool {
    // Check if user is owner
    if ac.Owner == user {
        return true
    }
    
    // Check if action is allowed
    for _, perm := range ac.Permissions {
        if perm == action || perm == "*" {
            // Check if user is in allowed list
            for _, allowed := range ac.AllowedUsers {
                if allowed == user || allowed == "*" {
                    return true
                }
            }
        }
    }
    
    return false
}
```

---

## Validation: internal/validation/

**Purpose:** Comprehensive validation of all element types.

### Validation Architecture

```
┌────────────────────────────────────────────────────┐
│            Validator Interface                     │
│  - Validate(element, level) ValidationResult       │
└────────────┬───────────────────────────────────────┘
             │
   ┌─────────┴─────────┬──────────────┬───────────┐
   │                   │              │           │
┌──▼──────────┐  ┌────▼────────┐  ┌──▼────────┐  │
│PersonaValid │  │SkillValid   │  │TemplateVal│  ...
│ator         │  │ator         │  │idator     │
└─────────────┘  └─────────────┘  └───────────┘
```

### ValidationResult Structure

```go
type ValidationResult struct {
    IsValid        bool
    Errors         []ValidationIssue
    Warnings       []ValidationIssue
    Infos          []ValidationIssue
    ValidationTime int64
    ElementType    string
    ElementID      string
}

type ValidationIssue struct {
    Severity   ValidationSeverity  // error, warning, info
    Field      string
    Message    string
    Line       int
    Suggestion string
    Code       string  // e.g., "PERSONA_TONE_INCONSISTENT"
}
```

### Validation Levels

```go
type ValidationLevel string

const (
    BasicLevel         ValidationLevel = "basic"          // Structure only
    ComprehensiveLevel ValidationLevel = "comprehensive"  // + Business rules
    StrictLevel        ValidationLevel = "strict"         // + Best practices
)
```

### Example: PersonaValidator

```go
type PersonaValidator struct{}

func (v *PersonaValidator) Validate(element domain.Element, level ValidationLevel) ValidationResult {
    persona, ok := element.(*domain.Persona)
    if !ok {
        return invalidTypeResult()
    }
    
    result := ValidationResult{
        IsValid:     true,
        ElementType: "persona",
        ElementID:   persona.GetID(),
    }
    
    start := time.Now()
    defer func() {
        result.ValidationTime = time.Since(start).Milliseconds()
    }()
    
    // Basic validation
    if err := persona.Metadata.Validate(); err != nil {
        result.AddError("metadata", err.Error(), "METADATA_INVALID")
    }
    
    // Role validation
    if persona.Role == "" {
        result.AddError("role", "Role is required", "ROLE_REQUIRED")
    } else if len(persona.Role) < 5 {
        result.AddWarning("role", "Role is very short, consider being more descriptive", "ROLE_TOO_SHORT")
    }
    
    // Comprehensive validation
    if level == ComprehensiveLevel || level == StrictLevel {
        v.validateToneConsistency(persona, &result)
        v.validateExpertiseRelevance(persona, &result)
        v.validateExamples(persona, &result)
    }
    
    // Strict validation
    if level == StrictLevel {
        v.validateBestPractices(persona, &result)
    }
    
    return result
}
```

---

## Application Services: internal/application/

**Files:**
- `statistics.go` - Usage statistics (287 lines)
- `ensemble_aggregation.go` - Result aggregation (412 lines)
- `ensemble_executor.go` - Ensemble execution (389 lines)
- `ensemble_monitor.go` - Execution monitoring (234 lines)

### Metrics Collection

```go
type MetricsCollector struct {
    mu          sync.RWMutex
    metricsFile string
    operations  []OperationMetric
}

type OperationMetric struct {
    Timestamp time.Time
    ToolName  string
    Duration  time.Duration
    Success   bool
    Error     string
    User      string
    Metadata  map[string]interface{}
}

func (mc *MetricsCollector) RecordOperation(tool string, duration time.Duration, success bool) {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    
    metric := OperationMetric{
        Timestamp: time.Now(),
        ToolName:  tool,
        Duration:  duration,
        Success:   success,
    }
    
    mc.operations = append(mc.operations, metric)
    
    // Persist every 100 operations
    if len(mc.operations)%100 == 0 {
        mc.persist()
    }
}
```

---

## Indexing: internal/indexing/

**Purpose:** TF-IDF indexing for semantic search.

### Index Structure

```go
type TFIDFIndex struct {
    mu         sync.RWMutex
    documents  map[string]*Document
    vocabulary map[string]float64  // IDF scores
    totalDocs  int
}

type Document struct {
    ID       string
    Content  string
    Type     string
    TF       map[string]float64  // Term frequency
    Vector   map[string]float64  // TF-IDF vector
}
```

### Indexing Process

```go
func (idx *TFIDFIndex) AddDocument(id, content, docType string) {
    // 1. Tokenize content
    tokens := tokenize(content)
    
    // 2. Calculate term frequency
    tf := calculateTF(tokens)
    
    // 3. Store document
    doc := &Document{
        ID:      id,
        Content: content,
        Type:    docType,
        TF:      tf,
    }
    
    idx.documents[id] = doc
    idx.totalDocs++
    
    // 4. Update vocabulary
    idx.updateVocabulary()
    
    // 5. Calculate TF-IDF vectors
    idx.calculateVectors()
}

func (idx *TFIDFIndex) Search(query string, limit int) []SearchResult {
    // 1. Tokenize query
    queryTokens := tokenize(query)
    queryTF := calculateTF(queryTokens)
    
    // 2. Calculate query vector
    queryVector := idx.calculateQueryVector(queryTF)
    
    // 3. Calculate cosine similarity with all documents
    scores := make([]SearchResult, 0)
    for id, doc := range idx.documents {
        score := cosineSimilarity(queryVector, doc.Vector)
        if score > 0 {
            scores = append(scores, SearchResult{
                DocumentID: id,
                Score:      score,
                Type:       doc.Type,
            })
        }
    }
    
    // 4. Sort by score
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].Score > scores[j].Score
    })
    
    // 5. Return top N results
    if limit > 0 && limit < len(scores) {
        return scores[:limit]
    }
    return scores
}
```

---

## Templates: internal/template/

**Purpose:** Template rendering with Handlebars.

```go
type TemplateEngine struct {
    renderer *raymond.Engine
}

func (te *TemplateEngine) Render(template domain.Template, data map[string]interface{}) (string, error) {
    // Register partials
    for name, content := range template.Partials {
        raymond.RegisterPartial(name, content)
    }
    
    // Render template
    result, err := raymond.Render(template.Content, data)
    if err != nil {
        return "", fmt.Errorf("failed to render template: %w", err)
    }
    
    return result, nil
}
```

---

## Data Flow Diagrams

### Create Element Flow

```
Client                MCP Server              Repository           Validation
  │                        │                        │                   │
  │──create_element─────> │                        │                   │
  │                        │                        │                   │
  │                        │──Parse Input──────────>│                   │
  │                        │                        │                   │
  │                        │──Validate──────────────┼──────────────────>│
  │                        │                        │                   │
  │                        │<─────────────────────────Valid/Invalid─────│
  │                        │                        │                   │
  │                        │──Create(element)─────> │                   │
  │                        │                        │                   │
  │                        │                        │──Write YAML File  │
  │                        │                        │                   │
  │                        │                        │──Update Cache     │
  │                        │                        │                   │
  │                        │<───Element Created─────│                   │
  │                        │                        │                   │
  │                        │──Update Index          │                   │
  │                        │                        │                   │
  │                        │──Record Metrics        │                   │
  │                        │                        │                   │
  │<────Success Response───│                        │                   │
```

### Search Flow

```
Client          MCP Server        TF-IDF Index      Repository
  │                  │                  │                │
  │──search_────────>│                  │                │
  │  elements        │                  │                │
  │                  │                  │                │
  │                  │──Tokenize Query─>│                │
  │                  │                  │                │
  │                  │                  │──Calculate     │
  │                  │                  │  Similarity    │
  │                  │                  │                │
  │                  │<─Result IDs──────│                │
  │                  │                  │                │
  │                  │──Get Elements────┼───────────────>│
  │                  │  by IDs          │                │
  │                  │                  │                │
  │                  │<────Elements─────┼────────────────│
  │                  │                  │                │
  │<─Search Results──│                  │                │
```

---

## Key Interfaces

### ElementRepository

```go
type ElementRepository interface {
    // CRUD
    Create(element Element) error
    Get(id string) (Element, error)
    Update(element Element) error
    Delete(id string) error
    List(filter ElementFilter) ([]Element, error)
    
    // Batch
    CreateBatch(elements []Element) error
    GetBatch(ids []string) ([]Element, error)
    
    // Advanced
    Duplicate(id string, newName string) (Element, error)
    GetByType(elementType ElementType) ([]Element, error)
}
```

### Validator

```go
type Validator interface {
    Validate(element Element, level ValidationLevel) ValidationResult
}
```

### Index

```go
type Index interface {
    AddDocument(id, content, docType string)
    RemoveDocument(id string)
    Search(query string, limit int) []SearchResult
    Rebuild(documents []Document) error
}
```

---

## Request/Response Flow

### Tool Invocation Flow

1. **Client sends JSON-RPC request over stdio:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "create_element",
    "arguments": {
      "type": "persona",
      "name": "Senior Engineer",
      "version": "1.0.0",
      "author": "team@example.com"
    }
  }
}
```

2. **SDK routes to handler:**
```go
s.handleCreateElement(ctx, arguments)
```

3. **Handler processes:**
```go
func (s *MCPServer) handleCreateElement(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    // Parse input
    var input CreateElementInput
    parseArguments(arguments, &input)
    
    // Create element
    element := constructElement(input)
    
    // Validate
    if err := element.Validate(); err != nil {
        return errorResult("Validation failed: %v", err), nil
    }
    
    // Store
    if err := s.repo.Create(element); err != nil {
        return errorResult("Failed to create: %v", err), nil
    }
    
    // Update index
    s.index.AddDocument(element.GetID(), elementContent, string(element.GetType()))
    
    // Return success
    return successResult(CreateElementOutput{
        ID: element.GetID(),
        Element: elementToMap(element),
    })
}
```

4. **SDK returns JSON-RPC response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"id\":\"persona-001\",\"element\":{...}}"
    }]
  }
}
```

---

## Quick Reference Guide

### Where to Find Things

| What | Where |
|------|-------|
| **Entry point** | `cmd/nexs-mcp/main.go` |
| **Configuration** | `internal/config/config.go` |
| **Logging** | `internal/logger/` |
| **Domain models** | `internal/domain/*.go` |
| **Validation** | `internal/validation/*_validator.go` |
| **Repository** | `internal/infrastructure/*_repository.go` |
| **MCP tools** | `internal/mcp/*_tools.go` |
| **Tool registration** | `internal/mcp/server.go:registerTools()` |
| **Indexing** | `internal/indexing/tfidf.go` |
| **Templates** | `internal/template/engine.go` |
| **Metrics** | `internal/application/statistics.go` |
| **Tests** | `*_test.go` files throughout |

### Common Tasks

#### Add a new tool
1. Define input/output structs in `internal/mcp/tools.go`
2. Implement handler in `internal/mcp/new_tool.go`
3. Register in `server.go:registerTools()`
4. Write tests in `new_tool_test.go`
5. Update docs in `docs/api/MCP_TOOLS.md`

#### Add a new element type
1. Define in `internal/domain/new_type.go`
2. Add to `ElementType` enum
3. Implement `Element` interface
4. Create validator in `internal/validation/new_type_validator.go`
5. Update repository implementations
6. Add MCP tools (create, quick_create)
7. Register tools
8. Write comprehensive tests

#### Add validation rule
1. Open appropriate validator (e.g., `persona_validator.go`)
2. Add validation logic in `Validate()` method
3. Add error/warning/info using `result.AddError()` etc.
4. Write test cases in `*_validator_test.go`
5. Update validation documentation

### Key Design Patterns

1. **Repository Pattern**: Abstracts data storage
2. **Dependency Injection**: Pass dependencies explicitly
3. **Interface Segregation**: Small, focused interfaces
4. **Error Wrapping**: Use `fmt.Errorf("...: %w", err)` 
5. **Structured Logging**: Use key-value pairs with slog
6. **Immutable Configuration**: Load once at startup
7. **Resource Cleanup**: Use `defer` for cleanup

### Performance Considerations

1. **Caching**: File repository uses in-memory cache
2. **Batch Operations**: Reduce I/O with batch tools
3. **Indexing**: Rebuild only when needed
4. **Connection Pooling**: Not applicable (stdio transport)
5. **Metrics**: Tracked but not blocking

### Testing Strategy

1. **Unit Tests**: Test individual functions
2. **Integration Tests**: Test tool handlers end-to-end
3. **Mock Repository**: Use for testing without I/O
4. **Table-Driven Tests**: Common Go pattern
5. **Coverage**: Aim for >80%

---

## Conclusion

This code tour provides a comprehensive overview of the NEXS-MCP codebase. Key takeaways:

1. **Clean Architecture**: Separation of concerns with clear layers
2. **Official MCP SDK**: Uses `github.com/modelcontextprotocol/go-sdk/mcp`
3. **Extensible Design**: Easy to add new element types and tools
4. **Comprehensive Validation**: Multi-level validation system
5. **Production Ready**: Metrics, logging, error handling
6. **Well Tested**: Extensive test coverage

For specific implementation guides, see:
- [ADDING_ELEMENT_TYPE.md](./ADDING_ELEMENT_TYPE.md)
- [ADDING_MCP_TOOL.md](./ADDING_MCP_TOOL.md)
- [EXTENDING_VALIDATION.md](./EXTENDING_VALIDATION.md)
