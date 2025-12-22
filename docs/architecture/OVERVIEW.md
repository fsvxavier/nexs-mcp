# NEXS MCP Architecture Overview

**Version:** 1.0.0  
**Last Updated:** December 20, 2025  
**Status:** Production

---

## Table of Contents

- [Introduction](#introduction)
- [Architecture Philosophy](#architecture-philosophy)
- [Clean Architecture Overview](#clean-architecture-overview)
- [System Architecture](#system-architecture)
- [Component Diagram](#component-diagram)
- [Data Flow](#data-flow)
- [Layer Responsibilities](#layer-responsibilities)
- [Design Principles](#design-principles)
- [Technology Stack](#technology-stack)
- [Deployment Architecture](#deployment-architecture)
- [Performance Characteristics](#performance-characteristics)
- [Security Architecture](#security-architecture)
- [Extensibility](#extensibility)

---

## Introduction

NEXS MCP Server is a **production-ready Model Context Protocol (MCP) server** built with enterprise-grade architecture principles using the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk). The system manages six types of AI elements (Personas, Skills, Templates, Agents, Memories, and Ensembles) and provides 66 MCP tools for comprehensive AI system management.

**Built with Official SDK:**
- Uses `github.com/modelcontextprotocol/go-sdk/mcp`
- Full MCP specification compliance
- Official tool and resource registration patterns
- Native stdio transport support

### Purpose

The architecture is designed to:

1. **Maintain Domain Purity** - Business logic isolated from infrastructure concerns
2. **Enable Testability** - High test coverage (72.2%) through dependency inversion
3. **Support Multiple Storage** - File-based (YAML) and in-memory implementations
4. **Ensure Performance** - Go's concurrency and efficient memory management
5. **Facilitate Evolution** - Clean boundaries allow independent layer changes

### Key Architectural Goals

- **Separation of Concerns** - Each layer has a single, well-defined responsibility
- **Dependency Inversion** - High-level modules don't depend on low-level modules
- **Testability** - All components can be tested in isolation
- **Maintainability** - Clear structure makes code easy to understand and modify
- **Extensibility** - New features can be added without modifying existing code

---

## Architecture Philosophy

NEXS MCP follows **Clean Architecture** (also known as Hexagonal or Ports & Adapters Architecture) principles, popularized by Robert C. Martin (Uncle Bob).

### Core Philosophy

```
┌─────────────────────────────────────────────────────────┐
│                    Clean Architecture                    │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │              External Interfaces               │    │
│  │  (MCP Protocol, CLI, HTTP, File System, DB)   │    │
│  └──────────────────┬─────────────────────────────┘    │
│                     │                                    │
│  ┌──────────────────▼─────────────────────────────┐    │
│  │           Infrastructure Layer                 │    │
│  │  (Adapters, Repositories, External Services)   │    │
│  └──────────────────┬─────────────────────────────┘    │
│                     │                                    │
│  ┌──────────────────▼─────────────────────────────┐    │
│  │           Application Layer                    │    │
│  │  (Use Cases, Orchestration, Services)          │    │
│  └──────────────────┬─────────────────────────────┘    │
│                     │                                    │
│  ┌──────────────────▼─────────────────────────────┐    │
│  │              Domain Layer                      │    │
│  │  (Business Logic, Entities, Rules)             │    │
│  │          [No external dependencies]            │    │
│  └────────────────────────────────────────────────┘    │
│                                                          │
└─────────────────────────────────────────────────────────┘

Dependency Flow: Outer → Inner (Never Inner → Outer)
```

### The Dependency Rule

> **Source code dependencies must point inward, toward higher-level policies.**

- **Domain Layer** has ZERO external dependencies
- **Application Layer** depends only on Domain
- **Infrastructure Layer** depends on Domain and Application
- **MCP Layer** orchestrates Infrastructure and Application

This inversion of control enables:

- **Testing** - Mock infrastructure for unit tests
- **Flexibility** - Swap implementations (file storage → database)
- **Portability** - Domain logic works anywhere

---

## Clean Architecture Overview

### The Four Layers

NEXS MCP implements four distinct architectural layers:

#### 1. Domain Layer (`internal/domain/`)

**The Heart of the System**

- **Entities**: Persona, Skill, Template, Agent, Memory, Ensemble
- **Interfaces**: ElementRepository, Element
- **Business Rules**: Validation, invariants, domain logic
- **Zero Dependencies**: No imports from other layers or external packages (except standard library)

```go
// Domain defines the contract, doesn't implement it
type ElementRepository interface {
    Create(element Element) error
    GetByID(id string) (Element, error)
    Update(element Element) error
    Delete(id string) error
    List(filter ElementFilter) ([]Element, error)
}
```

#### 2. Application Layer (`internal/application/`)

**Use Cases & Orchestration**

- **Use Cases**: Execute domain operations with cross-cutting concerns
- **Services**: MetricsCollector, StatisticsService, EnsembleExecutor
- **Orchestration**: Coordinate multiple domain entities
- **Depends On**: Domain layer only

```go
// Application orchestrates domain entities
type EnsembleExecutor struct {
    repository domain.ElementRepository  // Interface, not concrete
    logger     *slog.Logger
}

func (e *EnsembleExecutor) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
    // Load ensemble (domain entity)
    ensemble, err := e.loadEnsemble(req.EnsembleID)
    
    // Orchestrate execution (application logic)
    return e.executeSequential(ctx, ensemble, req)
}
```

#### 3. Infrastructure Layer (`internal/infrastructure/`)

**External Integrations**

- **Repositories**: FileElementRepository, InMemoryElementRepository
- **External Services**: GitHub API, OAuth, Encryption
- **Storage**: File system (YAML), memory
- **Depends On**: Domain (interfaces), Application (services)

```go
// Infrastructure implements domain interfaces
type FileElementRepository struct {
    baseDir string
    cache   map[string]*StoredElement
}

// Implements domain.ElementRepository
func (r *FileElementRepository) Create(element domain.Element) error {
    // File system implementation details
}
```

#### 4. MCP Layer (`internal/mcp/`)

**Protocol Implementation**

- **Server**: MCP Protocol server using official SDK
- **Tool Handlers**: 66 MCP tools for element management
- **Resources**: Capability index, summaries, statistics
- **Depends On**: All layers (orchestrates the entire system)

```go
// MCP layer orchestrates everything
type MCPServer struct {
    server      *sdk.Server
    repo        domain.ElementRepository    // Infrastructure
    metrics     *application.MetricsCollector  // Application
    index       *indexing.TFIDFIndex        // Indexing
}

func (s *MCPServer) handleCreateElement(ctx context.Context, ...) {
    // Validate input (MCP layer)
    // Call domain/application logic
    // Record metrics (application)
    // Return MCP response
}
```

---

## System Architecture

### High-Level Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                         MCP Clients                                │
│         (Claude Desktop, VS Code, Custom Integrations)             │
└─────────────────────────────┬──────────────────────────────────────┘
                              │ MCP Protocol (Stdio)
                              │
┌─────────────────────────────▼──────────────────────────────────────┐
│                        NEXS MCP Server                             │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    MCP Layer                             │   │
│  │  • Server (SDK v1.1.0)      • 66 Tool Handlers          │   │
│  │  • Resources Protocol       • Context Management         │   │
│  └──────────────┬───────────────────────────────────────────┘   │
│                 │                                                │
│  ┌──────────────▼──────────────┬──────────────────────────┐   │
│  │    Application Layer         │   Supporting Services     │   │
│  │  • EnsembleExecutor         │  • MetricsCollector      │   │
│  │  • EnsembleMonitor          │  • StatisticsService     │   │
│  │  • EnsembleAggregation      │  • PerformanceMetrics    │   │
│  └──────────────┬───────────────┴──────────────────────────┘   │
│                 │                                                │
│  ┌──────────────▼──────────────────────────────────────────┐   │
│  │                   Domain Layer                          │   │
│  │  • Persona  • Skill  • Template  • Agent               │   │
│  │  • Memory   • Ensemble                                 │   │
│  │  • ElementRepository Interface                         │   │
│  └──────────────┬──────────────────────────────────────────┘   │
│                 │                                                │
│  ┌──────────────▼──────────────────────────────────────────┐   │
│  │              Infrastructure Layer                       │   │
│  │  • FileRepository (YAML)    • InMemoryRepository       │   │
│  │  • GitHubClient            • GitHubOAuthClient         │   │
│  │  • TokenEncryptor          • PRTracker                 │   │
│  └──────────────┬──────────────┬──────────────────────────┘   │
│                 │              │                                │
└─────────────────┼──────────────┼────────────────────────────────┘
                  │              │
                  ▼              ▼
        ┌──────────────┐  ┌─────────────┐
        │ File System  │  │   GitHub    │
        │  (YAML/JSON) │  │     API     │
        └──────────────┘  └─────────────┘
```

### Package Structure

```
nexs-mcp/
├── cmd/
│   └── nexs-mcp/           # Entry point, CLI initialization
│       └── main.go
│
├── internal/               # Private application code
│   ├── domain/            # Domain Layer (Business Logic)
│   │   ├── element.go         # Base element interface & metadata
│   │   ├── persona.go         # Persona entity
│   │   ├── skill.go           # Skill entity
│   │   ├── template.go        # Template entity
│   │   ├── agent.go           # Agent entity
│   │   ├── memory.go          # Memory entity
│   │   ├── ensemble.go        # Ensemble entity
│   │   └── access_control.go  # Access control domain logic
│   │
│   ├── application/       # Application Layer (Use Cases)
│   │   ├── ensemble_executor.go     # Ensemble execution
│   │   ├── ensemble_monitor.go      # Execution monitoring
│   │   ├── ensemble_aggregation.go  # Result aggregation
│   │   └── statistics.go            # Metrics & statistics
│   │
│   ├── infrastructure/    # Infrastructure Layer (External)
│   │   ├── repository.go              # In-memory repository
│   │   ├── file_repository.go         # File-based repository (YAML)
│   │   ├── enhanced_file_repository.go # Enhanced with caching
│   │   ├── github_client.go           # GitHub API client
│   │   ├── github_oauth.go            # OAuth device flow
│   │   ├── github_publisher.go        # Collection publishing
│   │   ├── pr_tracker.go              # Pull request tracking
│   │   ├── crypto.go                  # Token encryption (AES-256-GCM)
│   │   ├── sync_metadata.go           # Sync state tracking
│   │   ├── sync_incremental.go        # Delta sync
│   │   └── sync_conflict_detector.go  # Conflict resolution
│   │
│   ├── mcp/              # MCP Layer (Protocol)
│   │   ├── server.go                  # MCP server
│   │   ├── tools.go                   # Element CRUD tools
│   │   ├── quick_create_tools.go      # Quick create tools
│   │   ├── github_tools.go            # GitHub sync tools
│   │   ├── github_auth_tools.go       # OAuth tools
│   │   ├── github_portfolio_tools.go  # Portfolio management
│   │   ├── collection_tools.go        # Collection tools
│   │   ├── collection_submission_tools.go # PR submission
│   │   ├── backup_tools.go            # Backup/restore tools
│   │   ├── memory_tools.go            # Memory management
│   │   ├── auto_save_tools.go         # Auto-save tools
│   │   ├── user_tools.go              # User identity
│   │   ├── log_tools.go               # Log query tools
│   │   ├── analytics_tools.go         # Analytics dashboard
│   │   ├── performance_tools.go       # Performance metrics
│   │   ├── search_tool.go             # Semantic search
│   │   ├── index_tools.go             # Index management
│   │   ├── ensemble_execution_tools.go # Ensemble execution
│   │   ├── discovery_tools.go         # Element discovery
│   │   ├── batch_tools.go             # Batch operations
│   │   ├── render_template_tools.go   # Template rendering
│   │   ├── element_validation_tools.go # Validation
│   │   ├── reload_elements_tools.go   # Cache reload
│   │   ├── publishing_tools.go        # Publishing
│   │   ├── template_tools.go          # Template tools
│   │   └── resources/                 # MCP Resources
│   │       └── capability_index.go
│   │
│   ├── config/           # Configuration Management
│   │   └── config.go         # App configuration
│   │
│   ├── logger/           # Structured Logging
│   │   ├── logger.go         # slog wrapper
│   │   └── performance.go    # Performance tracking
│   │
│   ├── validation/       # Element Validation
│   │   └── validator.go      # JSON Schema validation
│   │
│   ├── indexing/         # Semantic Search
│   │   └── tfidf.go          # TF-IDF index
│   │
│   ├── collection/       # Collection System
│   │   ├── manager.go        # Collection management
│   │   ├── registry.go       # Collection registry
│   │   ├── installer.go      # Installation
│   │   ├── manifest.go       # Manifest handling
│   │   └── validator.go      # Validation
│   │
│   ├── backup/           # Backup & Restore
│   │   ├── backup.go         # Backup creation
│   │   └── restore.go        # Backup restoration
│   │
│   ├── portfolio/        # Portfolio Management
│   │   └── manager.go        # Portfolio operations
│   │
│   └── template/         # Template Rendering
│       └── renderer.go       # Go template rendering
│
├── data/                 # Data Storage (File Mode)
│   └── elements/
│       ├── personas/
│       ├── skills/
│       ├── templates/
│       ├── agents/
│       ├── memories/
│       └── ensembles/
│
├── examples/            # Usage Examples
├── docs/                # Documentation
├── test/                # Integration Tests
└── scripts/             # Build & Deploy Scripts
```

---

## Component Diagram

### Core Components

```
                    ┌─────────────────────────────────┐
                    │         MCP Client              │
                    │   (Claude Desktop, VS Code)     │
                    └──────────────┬──────────────────┘
                                   │ stdio (JSON-RPC)
                    ┌──────────────▼──────────────────┐
                    │         MCPServer               │
                    │  ┌──────────────────────────┐  │
                    │  │  Official MCP SDK v1.1.0 │  │
                    │  └──────────────────────────┘  │
                    │  • Tool Registration            │
                    │  • Request Routing              │
                    │  • Response Formatting          │
                    └──────────────┬──────────────────┘
                                   │
            ┌──────────────────────┼──────────────────────┐
            │                      │                      │
┌───────────▼──────────┐ ┌────────▼─────────┐ ┌─────────▼────────┐
│   Tool Handlers      │ │  Resource        │ │   Supporting     │
│  • Element CRUD      │ │  Handlers        │ │   Services       │
│  • Quick Create      │ │  • Summary       │ │  • Metrics       │
│  • GitHub Sync       │ │  • Full Details  │ │  • Logging       │
│  • Collections       │ │  • Statistics    │ │  • Validation    │
│  • Backup/Restore    │ │                  │ │  • Indexing      │
│  • Memory Mgmt       │ │                  │ │                  │
│  • Analytics         │ │                  │ │                  │
└───────────┬──────────┘ └──────────────────┘ └──────────────────┘
            │
┌───────────▼──────────────────────────────────────────────┐
│              Application Services                        │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────┐ │
│  │  Ensemble      │  │   Metrics    │  │ Statistics  │ │
│  │  Executor      │  │  Collector   │  │  Service    │ │
│  └────────────────┘  └──────────────┘  └─────────────┘ │
└───────────┬──────────────────────────────────────────────┘
            │
┌───────────▼──────────────────────────────────────────────┐
│                  Domain Entities                         │
│  [Persona] [Skill] [Template] [Agent] [Memory] [Ensemble]│
│                                                           │
│             ElementRepository Interface                  │
└───────────┬──────────────────────────────────────────────┘
            │
            │ implements
            │
┌───────────▼──────────────────────────────────────────────┐
│           Infrastructure Implementations                 │
│  ┌──────────────────┐  ┌────────────────────────────┐  │
│  │  File Repository │  │  In-Memory Repository      │  │
│  │  (YAML Storage)  │  │  (RAM Storage)             │  │
│  └──────────────────┘  └────────────────────────────┘  │
│                                                          │
│  ┌──────────────────┐  ┌────────────────────────────┐  │
│  │  GitHub Client   │  │  GitHub OAuth Client       │  │
│  │  (API Wrapper)   │  │  (Device Flow Auth)        │  │
│  └──────────────────┘  └────────────────────────────┘  │
│                                                          │
│  ┌──────────────────┐  ┌────────────────────────────┐  │
│  │  Token Encryptor │  │  PR Tracker                │  │
│  │  (AES-256-GCM)   │  │  (Pull Request Tracking)   │  │
│  └──────────────────┘  └────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

### Component Interactions

```
┌──────────┐     ┌──────────────┐     ┌───────────┐     ┌─────────┐
│  Client  │────▶│  MCP Server  │────▶│ Tool      │────▶│ Domain  │
└──────────┘     └──────────────┘     │ Handler   │     │ Entity  │
                                       └───────────┘     └─────────┘
                                             │                │
                                             ▼                ▼
                                       ┌───────────┐    ┌──────────┐
                                       │Application│    │Repository│
                                       │ Service   │    │          │
                                       └───────────┘    └──────────┘
                                             │                │
                                             ▼                ▼
                                       ┌───────────┐    ┌──────────┐
                                       │ Metrics   │    │File/Mem  │
                                       │ Collector │    │ Storage  │
                                       └───────────┘    └──────────┘
```

---

## Data Flow

### Request Flow

```
1. MCP Client Request
   │
   ├─▶ JSON-RPC Message over Stdio
   │
   ├─▶ MCP Server (SDK)
   │   │
   │   ├─▶ Parse Request
   │   ├─▶ Route to Tool Handler
   │   └─▶ Validate Parameters
   │
   ├─▶ Tool Handler
   │   │
   │   ├─▶ Extract Input
   │   ├─▶ Validate Business Rules
   │   ├─▶ Record Metrics (start)
   │   │
   │   ├─▶ Call Application Service (if needed)
   │   │   │
   │   │   ├─▶ Orchestrate Multiple Operations
   │   │   ├─▶ Apply Cross-Cutting Concerns
   │   │   └─▶ Coordinate Domain Entities
   │   │
   │   ├─▶ Access Domain Entity
   │   │   │
   │   │   ├─▶ Validate Invariants
   │   │   ├─▶ Apply Business Logic
   │   │   └─▶ Return Result
   │   │
   │   ├─▶ Call Repository (Infrastructure)
   │   │   │
   │   │   ├─▶ Serialize/Deserialize
   │   │   ├─▶ File I/O or Memory Access
   │   │   └─▶ Return Entity
   │   │
   │   ├─▶ Record Metrics (end)
   │   └─▶ Format Response
   │
   └─▶ Return JSON-RPC Response
```

### Element Creation Flow

```go
// 1. MCP Client sends request
{
  "method": "tools/call",
  "params": {
    "name": "create_element",
    "arguments": {
      "type": "persona",
      "name": "AI Assistant",
      "description": "...",
      ...
    }
  }
}

// 2. MCP Server routes to handler
func (s *MCPServer) handleCreateElement(ctx context.Context, req *sdk.CallToolRequest, input CreateElementInput) {
    // 3. Record start time
    startTime := time.Now()
    
    // 4. Create domain entity
    var element domain.Element
    switch input.Type {
    case "persona":
        element = domain.NewPersona(input.Name, input.Description, input.Version, input.Author)
        // Populate fields from input.Data
    }
    
    // 5. Validate domain rules
    if err := element.Validate(); err != nil {
        return nil, output, err
    }
    
    // 6. Save via repository
    if err := s.repo.Create(element); err != nil {
        return nil, output, err
    }
    
    // 7. Update search index
    s.index.AddDocument(&indexing.Document{
        ID:      element.GetID(),
        Type:    element.GetType(),
        Name:    element.GetMetadata().Name,
        Content: extractContent(element),
    })
    
    // 8. Record metrics
    s.metrics.RecordToolCall(application.ToolCallMetric{
        ToolName:  "create_element",
        Timestamp: startTime,
        Duration:  time.Since(startTime),
        Success:   true,
    })
    
    // 9. Return response
    return &sdk.CallToolResult{
        Content: []sdk.Content{{
            Type: "text",
            Text: fmt.Sprintf("Element %s created successfully", element.GetID()),
        }},
    }, output, nil
}
```

### GitHub Sync Flow

```
┌─────────────┐
│ sync_github │ MCP Tool
└──────┬──────┘
       │
       ▼
┌─────────────────────────┐
│ GitHubClient            │
│ • Authenticate          │
│ • List Remote Files     │
│ • Get File Contents     │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│ FileRepository          │
│ • List Local Elements   │
│ • Compare Timestamps    │
│ • Detect Conflicts      │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│ SyncMetadata            │
│ • Track Sync State      │
│ • Record Last Sync      │
│ • Store Remote SHAs     │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│ Push/Pull Operations    │
│ • Create Files          │
│ • Update Files          │
│ • Delete Files          │
└─────────────────────────┘
```

---

## Layer Responsibilities

### Domain Layer Responsibilities

**Pure Business Logic - No External Dependencies**

✅ **Defining Entities**
- Element types (Persona, Skill, Template, Agent, Memory, Ensemble)
- Element metadata structure
- Value objects (BehavioralTrait, ExpertiseArea, etc.)

✅ **Business Rules**
- Validation logic (required fields, format validation)
- Invariant enforcement (consistency rules)
- Domain operations (Activate, Deactivate)

✅ **Interfaces (Ports)**
- ElementRepository interface
- Element interface
- Access control interfaces

❌ **What Domain Layer Should NOT Do**
- File I/O or database access
- Network calls or HTTP requests
- Logging or metrics collection
- Configuration reading
- External library dependencies

```go
// ✅ Good: Pure domain logic
type Persona struct {
    metadata ElementMetadata
    BehavioralTraits []BehavioralTrait
    ExpertiseAreas   []ExpertiseArea
    ResponseStyle    ResponseStyle
    SystemPrompt     string
}

func (p *Persona) Validate() error {
    if len(p.BehavioralTraits) == 0 {
        return fmt.Errorf("at least one behavioral trait is required")
    }
    // Domain validation logic
    return nil
}

// ❌ Bad: Infrastructure concerns in domain
func (p *Persona) SaveToFile(path string) error {
    // This belongs in infrastructure!
}
```

### Application Layer Responsibilities

**Use Cases & Orchestration**

✅ **Use Case Implementation**
- Execute business workflows
- Coordinate multiple domain entities
- Apply transaction boundaries

✅ **Cross-Cutting Concerns**
- Metrics collection
- Performance monitoring
- Statistics aggregation

✅ **Complex Orchestration**
- Ensemble execution (sequential, parallel, hybrid)
- Multi-step workflows
- Result aggregation

❌ **What Application Layer Should NOT Do**
- Implement storage details
- Make HTTP requests directly
- Handle MCP protocol details
- Implement UI/API concerns

```go
// ✅ Good: Application orchestration
type EnsembleExecutor struct {
    repository domain.ElementRepository  // Uses interface
    logger     *slog.Logger
}

func (e *EnsembleExecutor) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
    // Load ensemble from repository (via interface)
    ensemble, err := e.repository.GetByID(req.EnsembleID)
    
    // Orchestrate execution based on mode
    switch ensemble.ExecutionMode {
    case "sequential":
        return e.executeSequential(ctx, ensemble, req)
    case "parallel":
        return e.executeParallel(ctx, ensemble, req)
    case "hybrid":
        return e.executeHybrid(ctx, ensemble, req)
    }
}
```

### Infrastructure Layer Responsibilities

**External Systems & Storage**

✅ **Repository Implementations**
- FileElementRepository (YAML storage)
- InMemoryElementRepository (RAM storage)
- Future: DatabaseRepository, APIRepository

✅ **External Service Clients**
- GitHubClient (GitHub API wrapper)
- GitHubOAuthClient (OAuth device flow)
- TokenEncryptor (AES-256-GCM encryption)

✅ **Data Mapping**
- Domain entity ↔ Storage format
- API responses ↔ Domain entities
- File formats (YAML, JSON)

❌ **What Infrastructure Should NOT Do**
- Contain business logic
- Make business decisions
- Validate business rules

```go
// ✅ Good: Infrastructure implementation
type FileElementRepository struct {
    baseDir string
    cache   map[string]*StoredElement
}

func (r *FileElementRepository) Create(element domain.Element) error {
    // Map domain entity to storage format
    stored := r.mapToStored(element)
    
    // Serialize to YAML
    data, err := yaml.Marshal(stored)
    
    // Write to file
    path := r.getFilePath(element.GetMetadata())
    return os.WriteFile(path, data, 0644)
}

// ❌ Bad: Business logic in infrastructure
func (r *FileElementRepository) Create(element domain.Element) error {
    // Don't validate business rules here!
    if len(element.GetMetadata().Name) < 3 {
        return fmt.Errorf("name too short")
    }
    // This belongs in domain layer
}
```

### MCP Layer Responsibilities

**Protocol Implementation & API**

✅ **MCP Protocol Handling**
- Tool registration with SDK
- Request parsing and validation
- Response formatting
- Resource exposure

✅ **API Gateway Functions**
- Route requests to appropriate handlers
- Aggregate data from multiple sources
- Transform between MCP and internal formats

✅ **Cross-Layer Coordination**
- Inject dependencies
- Manage lifecycle
- Handle errors gracefully

❌ **What MCP Layer Should NOT Do**
- Implement business logic
- Directly access storage
- Bypass domain validation

```go
// ✅ Good: MCP protocol handling
func (s *MCPServer) handleCreateElement(ctx context.Context, req *sdk.CallToolRequest, input CreateElementInput) (*sdk.CallToolResult, CreateElementOutput, error) {
    // Parse MCP input
    // Create domain entity
    element := createDomainEntity(input)
    
    // Validate through domain
    if err := element.Validate(); err != nil {
        return nil, output, fmt.Errorf("validation failed: %w", err)
    }
    
    // Save via repository
    if err := s.repo.Create(element); err != nil {
        return nil, output, err
    }
    
    // Update index
    s.index.AddDocument(...)
    
    // Record metrics
    s.metrics.RecordToolCall(...)
    
    // Format MCP response
    return formatMCPResponse(element)
}
```

---

## Design Principles

### 1. SOLID Principles

#### Single Responsibility Principle (SRP)

Each component has one reason to change:

- **Domain Entities** - Business rule changes
- **Repositories** - Storage mechanism changes
- **Tool Handlers** - MCP protocol changes
- **Services** - Use case changes

```go
// ✅ Good: Single responsibility
type FileElementRepository struct {
    // Only responsible for file storage
}

type MetricsCollector struct {
    // Only responsible for metrics
}

// ❌ Bad: Multiple responsibilities
type FileRepositoryWithMetrics struct {
    // Handles both storage AND metrics - violates SRP
}
```

#### Open/Closed Principle (OCP)

Open for extension, closed for modification:

```go
// ✅ Good: Open for extension via interface
type ElementRepository interface {
    Create(element Element) error
    // Add new methods without breaking existing code
}

// New implementation doesn't modify existing code
type PostgreSQLRepository struct {
    // Implements ElementRepository
}
```

#### Liskov Substitution Principle (LSP)

Subtypes must be substitutable for their base types:

```go
// ✅ All implementations honor the contract
func NewMCPServer(repo domain.ElementRepository) *MCPServer {
    // Can use FileRepository, InMemoryRepository, or any future implementation
    // All behave correctly according to ElementRepository interface
}
```

#### Interface Segregation Principle (ISP)

Clients shouldn't depend on interfaces they don't use:

```go
// ✅ Good: Focused interfaces
type ElementReader interface {
    GetByID(id string) (Element, error)
    List(filter ElementFilter) ([]Element, error)
}

type ElementWriter interface {
    Create(element Element) error
    Update(element Element) error
}

// ❌ Bad: Fat interface forcing implementations to implement unused methods
type AllOperations interface {
    Create(element Element) error
    Update(element Element) error
    Delete(id string) error
    List(filter ElementFilter) ([]Element, error)
    // ... 20 more methods
}
```

#### Dependency Inversion Principle (DIP)

Depend on abstractions, not concretions:

```go
// ✅ Good: Depends on abstraction
type EnsembleExecutor struct {
    repository domain.ElementRepository  // Interface, not concrete type
}

// ❌ Bad: Depends on concrete implementation
type EnsembleExecutor struct {
    repository *FileElementRepository  // Tightly coupled
}
```

### 2. Dependency Injection

All dependencies injected at construction:

```go
// Constructor injection
func NewMCPServer(
    name string,
    version string,
    repo domain.ElementRepository,  // Injected
    cfg *config.Config,             // Injected
) *MCPServer {
    // All dependencies provided externally
    metrics := application.NewMetricsCollector(cfg.MetricsDir)
    index := indexing.NewTFIDFIndex()
    
    return &MCPServer{
        repo:    repo,
        metrics: metrics,
        index:   index,
        cfg:     cfg,
    }
}
```

### 3. Error Handling

Errors flow upward through layers:

```go
// Domain errors
var (
    ErrElementNotFound    = errors.New("element not found")
    ErrValidationFailed   = errors.New("validation failed")
    ErrInvalidElementType = errors.New("invalid element type")
)

// Infrastructure wraps domain errors
func (r *FileElementRepository) GetByID(id string) (domain.Element, error) {
    if _, exists := r.cache[id]; !exists {
        return nil, domain.ErrElementNotFound  // Return domain error
    }
    // ...
}

// MCP layer handles and formats errors
func (s *MCPServer) handleGetElement(...) (*sdk.CallToolResult, GetElementOutput, error) {
    element, err := s.repo.GetByID(input.ID)
    if errors.Is(err, domain.ErrElementNotFound) {
        return nil, output, fmt.Errorf("element with ID %s not found", input.ID)
    }
    // ...
}
```

### 4. Immutability & Value Objects

Prefer immutable data structures:

```go
// ✅ Value object - immutable
type BehavioralTrait struct {
    Name        string
    Description string
    Intensity   int  // 1-10 scale
}

// If you need to modify, create new instance
func (bt BehavioralTrait) WithIntensity(intensity int) BehavioralTrait {
    return BehavioralTrait{
        Name:        bt.Name,
        Description: bt.Description,
        Intensity:   intensity,
    }
}
```

### 5. Fail Fast

Validate early and clearly:

```go
// Validate at boundaries
func (p *Persona) Validate() error {
    if err := p.metadata.Validate(); err != nil {
        return fmt.Errorf("metadata validation failed: %w", err)
    }
    if len(p.BehavioralTraits) == 0 {
        return fmt.Errorf("at least one behavioral trait is required")
    }
    if p.SystemPrompt == "" {
        return fmt.Errorf("system_prompt is required")
    }
    return nil
}
```

### 6. Explicit > Implicit

Make intentions clear:

```go
// ✅ Good: Explicit operations
func (e *Ensemble) Activate() error {
    e.metadata.IsActive = true
    e.metadata.UpdatedAt = time.Now()
    return nil
}

// ❌ Bad: Side effects hidden
func (e *Ensemble) SetActive() {
    // Implicit mutation, unclear what happens
}
```

---

## Technology Stack

### Core Technologies

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.25+ | System programming language |
| **MCP SDK** | v1.1.0 | Model Context Protocol implementation |
| **slog** | stdlib | Structured logging |
| **YAML** | v3 | Element storage format |
| **JSON** | stdlib | MCP message format |

### Key Libraries

```go
import (
    // MCP Protocol
    sdk "github.com/modelcontextprotocol/go-sdk/mcp"
    
    // GitHub Integration
    "github.com/google/go-github/v57/github"
    "golang.org/x/oauth2"
    
    // Serialization
    "gopkg.in/yaml.v3"
    "encoding/json"
    
    // Cryptography
    "crypto/aes"
    "crypto/cipher"
    "golang.org/x/crypto/pbkdf2"
    
    // Validation
    "github.com/go-playground/validator/v10"
    
    // Standard Library
    "context"
    "sync"
    "time"
    "log/slog"
)
```

### External Services

- **GitHub API** - OAuth authentication, repository sync, PR management
- **File System** - YAML/JSON element storage
- **stdin/stdout** - MCP protocol communication

### Development Tools

- **Make** - Build automation
- **Go Test** - Unit and integration testing
- **Go Cover** - Test coverage analysis
- **golangci-lint** - Code quality
- **Docker** - Containerization
- **GitHub Actions** - CI/CD

---

## Deployment Architecture

### Binary Distribution

```
nexs-mcp-server
├── Linux (amd64, arm64)
├── macOS (amd64, arm64)
└── Windows (amd64, arm64)
```

Distributed via:
- **Go Install** - `go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@latest`
- **Homebrew** - `brew install fsvxavier/nexs-mcp/nexs-mcp`
- **NPM** - `npm install -g @fsvxavier/nexs-mcp-server`
- **Docker** - `docker pull fsvxavier/nexs-mcp:latest`

### Runtime Architecture

```
┌────────────────────────────────────────┐
│         Claude Desktop                 │
│    (or other MCP client)               │
└──────────────┬─────────────────────────┘
               │ stdio (JSON-RPC)
┌──────────────▼─────────────────────────┐
│       nexs-mcp Binary                  │
│  ┌──────────────────────────────────┐ │
│  │  MCP Server (Go Runtime)         │ │
│  │  • Tool Handlers                 │ │
│  │  • Resource Providers            │ │
│  │  • Domain Logic                  │ │
│  └──────────────────────────────────┘ │
└──────────────┬─────────────────────────┘
               │
    ┌──────────┼──────────┐
    │          │          │
    ▼          ▼          ▼
┌─────────┐ ┌─────────┐ ┌─────────┐
│  File   │ │ GitHub  │ │ Metrics │
│ System  │ │   API   │ │ Storage │
└─────────┘ └─────────┘ └─────────┘
```

### Configuration

```bash
# File storage (default)
nexs-mcp --data-dir=/path/to/data

# In-memory storage
nexs-mcp --storage-mode=memory

# Enable MCP Resources Protocol
nexs-mcp --resources-enabled=true

# Custom resource exposure
nexs-mcp --resources-enabled=true --resources-expose=summary,stats

# Environment variables
export NEXS_DATA_DIR=/path/to/data
export NEXS_STORAGE_MODE=file
export NEXS_RESOURCES_ENABLED=true
nexs-mcp
```

### Docker Deployment

```yaml
# docker-compose.yml
version: '3.8'

services:
  nexs-mcp:
    image: fsvxavier/nexs-mcp:latest
    volumes:
      - ./data:/app/data
      - ./metrics:/app/metrics
    environment:
      - NEXS_DATA_DIR=/app/data
      - NEXS_STORAGE_MODE=file
      - NEXS_RESOURCES_ENABLED=true
```

### Directory Structure (Runtime)

```
~/.nexs-mcp/
├── config/
│   └── config.yaml          # User configuration
├── tokens/
│   └── github.token.enc     # Encrypted GitHub token
├── metrics/
│   └── tool_metrics.json    # Usage metrics
├── performance/
│   └── performance.json     # Performance data
├── logs/
│   └── nexs-mcp.log         # Application logs
└── .salt                    # Encryption salt
```

---

## Performance Characteristics

### Benchmarks

| Operation | Performance | Notes |
|-----------|-------------|-------|
| **Element Creation** | ~1ms | In-memory |
| **Element Creation** | ~5ms | File storage (YAML) |
| **Element Retrieval** | ~0.1ms | Cached |
| **Element Retrieval** | ~2ms | File read |
| **Search (100 elements)** | ~10ms | TF-IDF index |
| **GitHub Sync** | ~50ms/file | Network dependent |
| **Ensemble Execution** | Variable | Depends on agent count |

### Memory Usage

| Storage Mode | Base Memory | Per Element | 1000 Elements |
|--------------|-------------|-------------|---------------|
| **In-Memory** | ~10MB | ~2KB | ~12MB |
| **File (Cached)** | ~8MB | ~1KB | ~9MB |
| **File (No Cache)** | ~5MB | 0 | ~5MB |

### Concurrency

- **Thread-Safe** - All operations use proper synchronization (sync.RWMutex)
- **Concurrent Reads** - Multiple readers can access simultaneously
- **Exclusive Writes** - Write operations acquire exclusive lock
- **Goroutine Pool** - Parallel ensemble execution uses bounded goroutines

```go
// Thread-safe repository
type FileElementRepository struct {
    mu      sync.RWMutex
    baseDir string
    cache   map[string]*StoredElement
}

func (r *FileElementRepository) GetByID(id string) (domain.Element, error) {
    r.mu.RLock()  // Allow concurrent reads
    defer r.mu.RUnlock()
    // ...
}

func (r *FileElementRepository) Create(element domain.Element) error {
    r.mu.Lock()  // Exclusive write
    defer r.mu.Unlock()
    // ...
}
```

### Optimization Strategies

1. **Caching** - In-memory cache for file-based storage
2. **Lazy Loading** - Load elements on-demand
3. **Indexing** - TF-IDF index for fast search
4. **Batch Operations** - Batch tool for multiple operations
5. **Connection Pooling** - HTTP client reuse for GitHub API

---

## Security Architecture

### Authentication

```
┌─────────────────────────────────────────────────────┐
│         GitHub OAuth Device Flow                    │
│                                                     │
│  1. Request device code                            │
│     github_auth_start                              │
│                                                     │
│  2. User authenticates in browser                  │
│     User visits URL and enters code                │
│                                                     │
│  3. Poll for token                                 │
│     github_auth_complete                           │
│                                                     │
│  4. Token received and encrypted                   │
│     AES-256-GCM encryption                         │
│                                                     │
│  5. Token stored securely                          │
│     ~/.nexs-mcp/tokens/github.token.enc           │
└─────────────────────────────────────────────────────┘
```

### Encryption

**Token Storage (AES-256-GCM)**

```go
type TokenEncryptor struct {
    key []byte  // 256-bit key derived from PBKDF2
}

// Key derivation
key := pbkdf2.Key(
    machineID,           // Machine identifier
    salt,                // Random 32-byte salt
    100000,              // 100,000 iterations
    32,                  // 256-bit key
    sha256.New,          // SHA-256 hash
)

// Encryption
ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
stored := base64.StdEncoding.EncodeToString(ciphertext)
```

### Privacy

| Element Type | Privacy Levels | Default | Sharing |
|--------------|----------------|---------|---------|
| **Persona** | Public, Private, Shared | Public | Owner + SharedWith |
| **Skill** | N/A | Public | All users |
| **Template** | N/A | Public | All users |
| **Agent** | N/A | Public | All users |
| **Memory** | Per-element | Private | Owner only |
| **Ensemble** | N/A | Public | All users |

### Access Control

```go
// Domain-level access control
type AccessControl struct {
    Owner      string   `json:"owner"`
    Privacy    string   `json:"privacy"`    // public, private, shared
    SharedWith []string `json:"shared_with"`
}

func (ac *AccessControl) CanAccess(userID string) bool {
    // Owner always has access
    if ac.Owner == userID {
        return true
    }
    
    // Public elements accessible to all
    if ac.Privacy == "public" {
        return true
    }
    
    // Shared elements check shared list
    if ac.Privacy == "shared" {
        for _, sharedUser := range ac.SharedWith {
            if sharedUser == userID {
                return true
            }
        }
    }
    
    return false
}
```

### File System Security

- **Permissions** - Files created with 0644 (rw-r--r--)
- **Directories** - Created with 0755 (rwxr-xr-x)
- **Sensitive Data** - Tokens encrypted and stored with 0600

---

## Extensibility

### Adding New Element Types

1. **Define Domain Entity** (`internal/domain/`)

```go
package domain

type MyNewElement struct {
    metadata ElementMetadata
    // Add custom fields
    CustomField string `json:"custom_field"`
}

func (m *MyNewElement) GetMetadata() ElementMetadata { return m.metadata }
func (m *MyNewElement) GetType() ElementType { return "mynew" }
func (m *MyNewElement) GetID() string { return m.metadata.ID }
func (m *MyNewElement) IsActive() bool { return m.metadata.IsActive }
func (m *MyNewElement) Activate() error { /* ... */ }
func (m *MyNewElement) Deactivate() error { /* ... */ }
func (m *MyNewElement) Validate() error { /* ... */ }
```

2. **Update Repository** (already supports via Element interface)

3. **Add MCP Tool Handler** (`internal/mcp/`)

```go
func (s *MCPServer) registerTools() {
    // ... existing tools
    
    sdk.AddTool(s.server, &sdk.Tool{
        Name: "create_mynew_element",
        Description: "Create a new MyNew element",
    }, s.handleCreateMyNew)
}

func (s *MCPServer) handleCreateMyNew(...) {
    // Implementation
}
```

### Adding New Storage Backends

Implement `domain.ElementRepository` interface:

```go
type PostgreSQLRepository struct {
    db *sql.DB
}

func (r *PostgreSQLRepository) Create(element domain.Element) error {
    // Convert element to SQL
    // Execute INSERT
}

func (r *PostgreSQLRepository) GetByID(id string) (domain.Element, error) {
    // Execute SELECT
    // Convert row to domain entity
}

// Implement remaining methods...
```

Then inject at startup:

```go
repo := infrastructure.NewPostgreSQLRepository(connString)
server := mcp.NewMCPServer("nexs-mcp", version, repo, cfg)
```

### Adding New MCP Tools

```go
// 1. Define input/output types
type MyCustomInput struct {
    Parameter1 string `json:"parameter1"`
    Parameter2 int    `json:"parameter2"`
}

type MyCustomOutput struct {
    Result string `json:"result"`
}

// 2. Register tool
sdk.AddTool(s.server, &sdk.Tool{
    Name:        "my_custom_tool",
    Description: "Does something custom",
}, s.handleMyCustomTool)

// 3. Implement handler
func (s *MCPServer) handleMyCustomTool(
    ctx context.Context,
    req *sdk.CallToolRequest,
    input MyCustomInput,
) (*sdk.CallToolResult, MyCustomOutput, error) {
    // Implementation
    output := MyCustomOutput{
        Result: "Success",
    }
    
    return &sdk.CallToolResult{
        Content: []sdk.Content{{
            Type: "text",
            Text: "Operation completed",
        }},
    }, output, nil
}
```

### Adding New Resources

```go
// 1. Define resource
const URIMyResource = "nexs://myresource"

// 2. Create handler
func (r *MyResource) Handler() sdk.ResourceHandler {
    return func(ctx context.Context, uri string) (*sdk.Resource, error) {
        // Generate resource content
        return &sdk.Resource{
            URI:      URIMyResource,
            Name:     "My Custom Resource",
            MIMEType: "text/markdown",
            Contents: []sdk.ResourceContent{{
                Type: "text",
                Text: "Resource content here",
            }},
        }, nil
    }
}

// 3. Register in server
s.server.AddResource(&sdk.Resource{
    URI:         URIMyResource,
    Name:        "My Custom Resource",
    Description: "Provides custom information",
    MIMEType:    "text/markdown",
}, myResource.Handler())
```

---

## Cross-References

- **[Domain Layer Documentation](./DOMAIN.md)** - Deep dive into domain entities and business logic
- **[Application Layer Documentation](./APPLICATION.md)** - Use cases and orchestration patterns
- **[Infrastructure Layer Documentation](./INFRASTRUCTURE.md)** - External integrations and storage
- **[MCP Layer Documentation](./MCP.md)** - MCP Protocol implementation and tools
- **[API Documentation](../api/)** - Complete API reference
- **[User Guide](../user-guide/)** - Getting started and usage examples

---

## Conclusion

NEXS MCP Server's architecture demonstrates how Clean Architecture principles can be applied to build a maintainable, testable, and extensible system. The clear separation of concerns, dependency inversion, and focus on domain purity enable the system to evolve without compromising stability.

**Key Takeaways:**

1. **Domain-Centric** - Business logic is independent and testable
2. **Interface-Driven** - Abstractions enable flexibility
3. **Layer Isolation** - Changes in one layer don't affect others
4. **Production-Ready** - 72.2% test coverage with comprehensive validation
5. **Extensible** - New features added without modifying existing code

This architecture provides a solid foundation for building enterprise-grade MCP servers and serves as a reference implementation for Clean Architecture in Go.

---

**Document Version:** 1.0.0  
**Total Lines:** 1147  
**Last Updated:** December 20, 2025
