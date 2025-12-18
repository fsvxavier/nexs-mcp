# Arquitetura Técnica - MCP Server Go

**Versão:** 1.0  
**Data:** 18 de Dezembro de 2025

## Índice
1. [Visão Geral Arquitetural](#visão-geral-arquitetural)
2. [Modelos de IA Suportados](#modelos-de-ia-suportados)
3. [Camadas da Arquitetura](#camadas-da-arquitetura)
4. [Domain Model](#domain-model)
5. [Padrões de Design](#padrões-de-design)
6. [Fluxo de Dados](#fluxo-de-dados)
7. [Decisões Arquiteturais](#decisões-arquiteturais)

---

## Visão Geral Arquitetural

### Princípios Fundamentais

1. **Clean Architecture + Hexagonal Architecture**
   - Core business logic independente de frameworks
   - Dependency inversion em todas as camadas
   - Testabilidade máxima

2. **Domain-Driven Design (DDD)**
   - Ubiquitous language
   - Bounded contexts claros
   - Aggregates e Entities bem definidos

3. **SOLID Principles**
   - Single Responsibility
   - Open/Closed
   - Liskov Substitution
   - Interface Segregation
   - Dependency Inversion

---

## Modelos de IA Suportados

O servidor MCP foi projetado para ser compatível com múltiplos modelos de IA, permitindo flexibilidade na escolha do modelo mais adequado para cada tarefa.

### Lista de Modelos Suportados

#### Seleção Automática
- **auto** - Seleciona automaticamente o melhor modelo baseado na solicitação

#### Família Claude (Anthropic)
- **claude-sonnet-4.5** - Equilíbrio ideal entre capacidade e velocidade
- **claude-haiku-4.5** - Respostas rápidas e eficientes
- **claude-opus-4.5** - Máxima capacidade de raciocínio
- **claude-sonnet-4** - Versão estável anterior

#### Família Gemini (Google)
- **gemini-2.5-pro** - Gemini Pro 2.5
- **gemini-3-flash-preview** - Preview de alta velocidade
- **gemini-3-pro-preview** - Preview avançado

#### Família GPT (OpenAI)
- **gpt-4.1** - GPT-4.1 base
- **gpt-4o** - GPT-4 otimizado
- **gpt-5** - GPT-5 base
- **gpt-5-mini** - Versão compacta e eficiente
- **gpt-5-codex** - Especializado em código
- **gpt-5.1** - GPT-5.1 base
- **gpt-5.1-codex** - Codex versão 5.1
- **gpt-5.1-codex-max** - Máxima capacidade para código
- **gpt-5.1-codex-mini** - Versão compacta do Codex
- **gpt-5.2** - Última versão GPT-5

#### Modelos Especializados
- **grok-code-fast-1** - Otimizado para geração rápida de código
- **oswe-vscode-prim** - Integração especializada para VSCode

### Configuração de Modelos

```go
// internal/mcp/config/models.go
type ModelConfig struct {
    DefaultModel string   `json:"default_model" yaml:"default_model"`
    EnabledModels []string `json:"enabled_models" yaml:"enabled_models"`
}

var SupportedModels = []string{
    "auto",
    "claude-sonnet-4.5",
    "claude-haiku-4.5",
    "claude-opus-4.5",
    "claude-sonnet-4",
    "gemini-2.5-pro",
    "gemini-3-flash-preview",
    "gemini-3-pro-preview",
    "gpt-4.1",
    "gpt-4o",
    "gpt-5",
    "gpt-5-mini",
    "gpt-5-codex",
    "gpt-5.1",
    "gpt-5.1-codex",
    "gpt-5.1-codex-max",
    "gpt-5.1-codex-mini",
    "gpt-5.2",
    "grok-code-fast-1",
    "oswe-vscode-prim",
}
```

### Seleção Automática (modo "auto")

O modo `auto` analisa a solicitação e seleciona o modelo ideal baseado em:

1. **Complexidade da Tarefa:**
   - Tarefas simples → Modelos rápidos (haiku, mini)
   - Tarefas complexas → Modelos avançados (opus, codex-max)

2. **Tipo de Conteúdo:**
   - Geração de código → Modelos Codex/Grok
   - Raciocínio complexo → Opus/Sonnet
   - Respostas rápidas → Haiku/Flash

3. **Disponibilidade e Performance:**
   - Verifica disponibilidade dos modelos
   - Considera latência e custo
   - Fallback automático se modelo indisponível

---

### Arquitetura em Camadas

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Stdio        │  │ HTTP/SSE     │  │ WebSocket    │      │
│  │ Transport    │  │ Transport    │  │ Transport    │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
└─────────┼──────────────────┼──────────────────┼─────────────┘
          │                  │                  │
┌─────────┼──────────────────┼──────────────────┼─────────────┐
│         │      Application Layer (MCP Server)│              │
│         ▼                  ▼                  ▼              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           MCP Protocol Handler (JSON-RPC 2.0)        │   │
│  └──────────────────────────────────────────────────────┘   │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Tool Registry & Router                   │   │
│  └──────────────────────────────────────────────────────┘   │
└──────────────────────────────┬───────────────────────────────┘
                               │
┌──────────────────────────────┼───────────────────────────────┐
│              Domain Layer (Business Logic)                   │
│                              │                                │
│  ┌───────────────────────────┼────────────────────────────┐  │
│  │         Use Cases (Interactors)                        │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │  │
│  │  │ Element CRUD │  │ Portfolio    │  │ Collection  │ │  │
│  │  │ Operations   │  │ Management   │  │ Browser     │ │  │
│  │  └──────────────┘  └──────────────┘  └─────────────┘ │  │
│  └────────────────────────────────────────────────────────┘  │
│                              │                                │
│  ┌───────────────────────────┼────────────────────────────┐  │
│  │            Domain Entities & Aggregates                │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐            │  │
│  │  │ Element  │  │Portfolio │  │Collection│            │  │
│  │  │ (Root)   │  │(Root)    │  │(Root)    │            │  │
│  │  └──────────┘  └──────────┘  └──────────┘            │  │
│  └────────────────────────────────────────────────────────┘  │
│                              │                                │
│  ┌───────────────────────────┼────────────────────────────┐  │
│  │           Domain Services                              │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │  │
│  │  │ Validation   │  │ Search       │  │ Security    │ │  │
│  │  │ Engine       │  │ Indexer      │  │ Scanner     │ │  │
│  │  └──────────────┘  └──────────────┘  └─────────────┘ │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────┬───────────────────────────────┘
                               │
┌──────────────────────────────┼───────────────────────────────┐
│         Infrastructure Layer (Ports & Adapters)              │
│                              │                                │
│  ┌───────────────────────────┼────────────────────────────┐  │
│  │              Repositories (Interfaces)                 │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │  │
│  │  │ Element      │  │ Portfolio    │  │ Collection  │ │  │
│  │  │ Repository   │  │ Repository   │  │ Repository  │ │  │
│  │  └──────────────┘  └──────────────┘  └─────────────┘ │  │
│  └────────────────────────────────────────────────────────┘  │
│                              │                                │
│  ┌───────────────────────────┼────────────────────────────┐  │
│  │       Concrete Implementations (Adapters)              │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │  │
│  │  │ Filesystem   │  │ GitHub API   │  │ BadgerDB    │ │  │
│  │  │ Adapter      │  │ Adapter      │  │ Adapter     │ │  │
│  │  └──────────────┘  └──────────────┘  └─────────────┘ │  │
│  └────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

---

## Camadas da Arquitetura

### 1. Presentation Layer (Transport)

**Responsabilidade:** Comunicação com clientes MCP via SDK oficial

#### Transportes Suportados

O servidor suporta **3 tipos de transportes** via SDK oficial:

##### 1. Stdio Transport (Padrão)
**Uso:** Integração com Claude Desktop e CLIs  
**Protocolo:** JSON-RPC 2.0 via stdin/stdout

```go
package transport

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/server"
    "github.com/modelcontextprotocol/go-sdk/transport"
)

// NewStdioServer creates MCP server with stdio transport
func NewStdioServer(name, version string) (*server.MCPServer, error) {
    // Stdio transport lê de stdin e escreve em stdout
    transport := transport.NewStdioTransport()
    
    srv := server.NewMCPServer(
        server.WithName(name),
        server.WithVersion(version),
        server.WithTransport(transport),
    )
    
    return srv, nil
}

// Uso típico com Claude Desktop:
// claude-desktop executa: ./mcp-server
// Comunicação bidirecional via stdin/stdout
```

**Características:**
- Ideal para Claude Desktop integration
- Zero configuração de rede
- Baixa latência (IPC local)
- Suporta streaming de respostas

---

##### 2. SSE Transport (Server-Sent Events)
**Uso:** Web clients, dashboards, monitoring  
**Protocolo:** HTTP + Server-Sent Events

```go
// NewSSEServer creates MCP server with SSE transport
func NewSSEServer(name, version string, addr string) (*server.MCPServer, error) {
    // SSE transport via HTTP streaming
    transport := transport.NewSSETransport(addr)
    
    srv := server.NewMCPServer(
        server.WithName(name),
        server.WithVersion(version),
        server.WithTransport(transport),
    )
    
    return srv, nil
}

// Servidor HTTP escuta em :8080
// Cliente conecta: GET http://localhost:8080/sse
// Recebe eventos: data: {"jsonrpc":"2.0",...}
```

**Características:**
- Unidirecional (server → client)
- Auto-reconnect nativo (browsers)
- Ideal para dashboards real-time
- CORS support built-in

**Endpoints:**
- `GET /sse` - Estabelece conexão SSE
- `POST /message` - Envia mensagens do cliente

---

##### 3. HTTP Transport (REST)
**Uso:** Integrações REST, webhooks, APIs  
**Protocolo:** HTTP POST com JSON-RPC 2.0

```go
// NewHTTPServer creates MCP server with HTTP transport
func NewHTTPServer(name, version string, addr string) (*server.MCPServer, error) {
    // HTTP transport tradicional
    transport := transport.NewHTTPTransport(addr)
    
    srv := server.NewMCPServer(
        server.WithName(name),
        server.WithVersion(version),
        server.WithTransport(transport),
    )
    
    return srv, nil
}

// Servidor HTTP escuta em :8080
// Cliente envia: POST http://localhost:8080/rpc
// Body: {"jsonrpc":"2.0","method":"tools/list","id":1}
```

**Características:**
- Request/Response síncrono
- Compatible com qualquer HTTP client
- Suporta batch requests
- CORS configurável

**Endpoints:**
- `POST /rpc` - JSON-RPC endpoint
- `GET /health` - Health check
- `GET /capabilities` - List server capabilities

---

#### Múltiplos Transportes Simultâneos

É possível rodar múltiplos transportes ao mesmo tempo:

```go
package main

import (
    "context"
    "log"
    "sync"
    
    "github.com/fsvxavier/mcp-server/internal/mcp"
)

func main() {
    ctx := context.Background()
    var wg sync.WaitGroup
    
    // 1. Stdio transport (Claude Desktop)
    stdioSrv, _ := mcp.NewStdioServer("mcp-server", "1.0.0")
    wg.Add(1)
    go func() {
        defer wg.Done()
        stdioSrv.Start(ctx)
    }()
    
    // 2. SSE transport (Web dashboard)
    sseSrv, _ := mcp.NewSSEServer("mcp-server", "1.0.0", ":8080")
    wg.Add(1)
    go func() {
        defer wg.Done()
        sseSrv.Start(ctx)
    }()
    
    // 3. HTTP transport (REST API)
    httpSrv, _ := mcp.NewHTTPServer("mcp-server", "1.0.0", ":8081")
    wg.Add(1)
    go func() {
        defer wg.Done()
        httpSrv.Start(ctx)
    }()
    
    log.Println("MCP Server running on 3 transports:")
    log.Println("  - Stdio (Claude Desktop)")
    log.Println("  - SSE: http://localhost:8080/sse")
    log.Println("  - HTTP: http://localhost:8081/rpc")
    
    wg.Wait()
}
```

**Benefícios:**
- Stdio para Claude Desktop (produção)
- SSE para monitoring dashboard
- HTTP para integrações e testes

---

**Common Transport Features:**
- Protocol compliance garantido pelo SDK
- Graceful shutdown nativo
- Context-based lifecycle
- Error handling padronizado
- Automatic JSON-RPC 2.0 parsing
- Request/response logging

### 2. Application Layer (MCP Server)

**Responsabilidade:** Registro de tools e schema auto-generation

```go
package mcp

import (
    "reflect"
    "github.com/modelcontextprotocol/go-sdk/server"
    "github.com/invopop/jsonschema"
)

// ToolRegistry manages tool registration with auto schema generation
type ToolRegistry struct {
    server *server.MCPServer
    reflector *jsonschema.Reflector
}

// RegisterTool registers a tool with automatic schema generation
func (r *ToolRegistry) RegisterTool(name, description string, handler interface{}) error {
    // 1. Generate JSON Schema from handler input struct using reflection
    schema := r.generateSchema(handler)
    
    // 2. Register tool with SDK
    return r.server.RegisterTool(server.Tool{
        Name:        name,
        Description: description,
        InputSchema: schema,
        Handler:     r.wrapHandler(handler),
    })
}

// generateSchema generates JSON Schema from Go struct via reflection
func (r *ToolRegistry) generateSchema(handler interface{}) map[string]interface{} {
    // Get handler input type via reflection
    handlerType := reflect.TypeOf(handler)
    inputType := handlerType.In(1) // assume func(ctx, input)
    
    // Generate schema using jsonschema library
    schema := r.reflector.Reflect(inputType)
    
    return schema
}

// Example tool with validation tags
type ListElementsInput struct {
    Type string `json:"type" jsonschema:"required,enum=personas|skills|templates" validate:"required,oneof=personas skills templates"`
}

func (h *ElementHandler) ListElements(ctx context.Context, input ListElementsInput) (*ListElementsOutput, error) {
    // Implementation
}
```

**Componentes:**
- SDK server wrapper
- **Auto schema generation** via reflection
- Tool registry com type-safe handlers
- Validation automática (struct tags)
- Error handler integrado ao SDK

### 3. Domain Layer (Business Logic)

**Responsabilidade:** Regras de negócio puras, independentes de frameworks

#### 3.1 Domain Entities

```go
package domain

// Element is the root aggregate for all element types
type Element struct {
    ID          string
    Type        ElementType
    Name        string
    Version     string
    Author      string
    Description string
    Content     string
    Metadata    ElementMetadata
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// ElementMetadata contains type-specific metadata
type ElementMetadata struct {
    // Persona-specific
    BehavioralTraits []string `yaml:"behavioral_traits,omitempty"`
    
    // Private Personas fields
    PrivacyLevel     PrivacyLevel `yaml:"privacy_level,omitempty"`
    Owner            string       `yaml:"owner,omitempty"`
    SharedWith       []string     `yaml:"shared_with,omitempty"`
    TemplateID       string       `yaml:"template_id,omitempty"`
    ForkedFrom       string       `yaml:"forked_from,omitempty"`
    VersionHistory   []Version    `yaml:"version_history,omitempty"`
    
    // Skill-specific
    Triggers []string `yaml:"triggers,omitempty"`
    
    // Memory-specific
    RetentionPolicy string `yaml:"retention_policy,omitempty"`
    Tags            []string `yaml:"tags,omitempty"`
    
    // Agent-specific
    Goals []string `yaml:"goals,omitempty"`
}

type PrivacyLevel string
const (
    PrivacyPublic  PrivacyLevel = "public"   // Community visible
    PrivacyPrivate PrivacyLevel = "private"  // Owner only
    PrivacyShared  PrivacyLevel = "shared"   // Specific users
)

type Version struct {
    Number      int       `yaml:"number"`
    Timestamp   time.Time `yaml:"timestamp"`
    Author      string    `yaml:"author"`
    Message     string    `yaml:"message"`
    ContentHash string    `yaml:"content_hash"` // SHA-256
}
    
    // Template-specific
    Variables []TemplateVariable `yaml:"variables,omitempty"`
}

// ElementType represents the type of element
type ElementType string

const (
    PersonaElement  ElementType = "personas"
    SkillElement    ElementType = "skills"
    TemplateElement ElementType = "templates"
    AgentElement    ElementType = "agents"
    MemoryElement   ElementType = "memories"
    EnsembleElement ElementType = "ensembles"
)
```

#### 3.2 Value Objects

```go
package domain

// ElementID is a value object representing element identification
type ElementID struct {
    Type      ElementType
    Name      string
    Author    string
    Timestamp time.Time
}

// Format: {type}_{name}_{author}_{YYYYMMDD}-{HHMMSS}
func (id ElementID) String() string {
    return fmt.Sprintf("%s_%s_%s_%s",
        id.Type,
        id.Name,
        id.Author,
        id.Timestamp.Format("20060102-150405"),
    )
}

// ValidationResult represents validation outcome
type ValidationResult struct {
    Valid    bool
    Errors   []ValidationError
    Warnings []ValidationWarning
}
```

#### 3.3 Domain Services

```go
package domain

// ValidationService validates elements according to business rules
type ValidationService struct {
    rules []ValidationRule
}

func (s *ValidationService) Validate(elem *Element) ValidationResult {
    // 300+ validation rules
    result := ValidationResult{Valid: true}
    
    for _, rule := range s.rules {
        if err := rule.Check(elem); err != nil {
            result.Valid = false
            result.Errors = append(result.Errors, err)
        }
    }
    
    return result
}

// SearchIndexer builds inverted index for elements
type SearchIndexer struct {
    index *InvertedIndex
    nlp   *NLPScorer
}

func (s *SearchIndexer) Index(elem *Element) error {
    // Build inverted index with NLP scoring
}

func (s *SearchIndexer) Search(query string) []SearchResult {
    // Jaccard similarity + Shannon entropy scoring
}

// PrivatePersonaService handles private persona operations
type PrivatePersonaService struct {
    repo          ElementRepository
    authz         AuthorizationService
    versionCtrl   VersionControlService
}

func (s *PrivatePersonaService) CreatePrivatePersona(ctx context.Context, owner string, input CreatePersonaInput) (*Persona, error) {
    // Create in personas/private-{owner}/ directory
    // Set PrivacyLevel = Private, Owner = owner
    // Initialize version control
}

func (s *PrivatePersonaService) SharePersona(ctx context.Context, personaID string, sharedWith []string) error {
    // Verify owner permissions
    // Update SharedWith list
    // Generate share URLs
}

func (s *PrivatePersonaService) ForkPersona(ctx context.Context, sourceID, newOwner string, customizations map[string]interface{}) (*Persona, error) {
    // Check fork permissions
    // Clone source, apply customizations
    // Save to private-{newOwner}/, set ForkedFrom
}

// VersionControlService manages Git-like versioning
type VersionControlService struct {
    storage VersionStorage
}

func (s *VersionControlService) CreateVersion(personaID, message string) (*Version, error) {
    // Snapshot with SHA-256 hash, increment version
}

func (s *VersionControlService) DiffVersions(personaID string, v1, v2 int) (*Diff, error) {
    // Field-by-field diff between versions
}
```

#### 3.4 Use Cases (Interactors)

```go
package usecase

// CreateElementUseCase handles element creation
type CreateElementUseCase struct {
    repo       ElementRepository
    validator  *ValidationService
    indexer    *SearchIndexer
}

func (uc *CreateElementUseCase) Execute(ctx context.Context, input CreateElementInput) (*Element, error) {
    // 1. Validate input
    if err := uc.validator.ValidateInput(input); err != nil {
        return nil, err
    }
    
    // 2. Create element
    elem := &Element{
        ID:   generateID(input),
        Type: input.Type,
        Name: input.Name,
        // ...
    }
    
    // 3. Validate element
    if result := uc.validator.Validate(elem); !result.Valid {
        return nil, errors.New("validation failed")
    }
    
    // 4. Save to repository
    if err := uc.repo.Save(ctx, elem); err != nil {
        return nil, err
    }
    
    // 5. Index for search
    if err := uc.indexer.Index(elem); err != nil {
        // Log but don't fail
        log.Warn("failed to index element", "error", err)
    }
    
    return elem, nil
}
```

### 4. Infrastructure Layer (Adapters)

**Responsabilidade:** Implementações concretas de interfaces de domínio

```go
package infrastructure

// FilesystemElementRepository implements ElementRepository using filesystem
type FilesystemElementRepository struct {
    basePath string
}

func (r *FilesystemElementRepository) Save(ctx context.Context, elem *domain.Element) error {
    // Implement filesystem persistence
    path := r.getElementPath(elem)
    
    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }
    
    // Serialize to YAML/Markdown
    data, err := r.serialize(elem)
    if err != nil {
        return err
    }
    
    // Atomic write
    return atomicWrite(path, data)
}

func (r *FilesystemElementRepository) FindByID(ctx context.Context, id string) (*domain.Element, error) {
    // Implement filesystem read
}

// GitHubPortfolioRepository implements PortfolioRepository using GitHub API
type GitHubPortfolioRepository struct {
    client *github.Client
    oauth  *oauth2.Config
}

func (r *GitHubPortfolioRepository) Sync(ctx context.Context, local *Portfolio) error {
    // Implement bidirectional sync with GitHub
}
```

---

## Domain Model

### Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        Element                               │
│  (Abstract Aggregate Root)                                   │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ - ID: string                                            │ │
│ │ - Type: ElementType                                     │ │
│ │ - Name: string                                          │ │
│ │ - Version: string                                       │ │
│ │ - Author: string                                        │ │
│ │ - Description: string                                   │ │
│ │ - Content: string                                       │ │
│ │ - Metadata: ElementMetadata                             │ │
│ │ - CreatedAt: time.Time                                  │ │
│ │ - UpdatedAt: time.Time                                  │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────┬───────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│   Persona     │ │    Skill      │ │   Template    │
├───────────────┤ ├───────────────┤ ├───────────────┤
│ + Traits      │ │ + Triggers    │ │ + Variables   │
│ + Expertise   │ │ + Procedures  │ │ + Format      │
└───────────────┘ └───────────────┘ └───────────────┘

        │                 │                 │
        ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│    Agent      │ │    Memory     │ │   Ensemble    │
├───────────────┤ ├───────────────┤ ├───────────────┤
│ + Goals       │ │ + Entries     │ │ + Elements    │
│ + Actions     │ │ + Retention   │ │ + Composition │
└───────────────┘ └───────────────┘ └───────────────┘
```

### Portfolio Model

```
┌─────────────────────────────────────────────────────────────┐
│                      Portfolio                               │
│  (Aggregate Root)                                            │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ - ID: string                                            │ │
│ │ - Owner: User                                           │ │
│ │ - Location: PortfolioLocation                           │ │
│ │ - Elements: []Element                                   │ │
│ │ - SyncStatus: SyncStatus                                │ │
│ │ - LastSync: time.Time                                   │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ contains
                          ▼
                    ┌──────────┐
                    │ Element  │
                    └──────────┘
```

---

## Padrões de Design

### 1. Repository Pattern
```go
// Domain layer defines interface
type ElementRepository interface {
    Save(ctx context.Context, elem *Element) error
    FindByID(ctx context.Context, id string) (*Element, error)
    FindByType(ctx context.Context, typ ElementType) ([]*Element, error)
    Delete(ctx context.Context, id string) error
}

// Infrastructure layer implements
type FilesystemElementRepository struct {
    basePath string
}
```

### 2. Factory Pattern
```go
type ElementFactory struct {}

func (f *ElementFactory) CreateElement(typ ElementType, input CreateInput) (*Element, error) {
    switch typ {
    case PersonaElement:
        return f.createPersona(input)
    case SkillElement:
        return f.createSkill(input)
    // ...
    }
}
```

### 3. Strategy Pattern (Validation)
```go
type ValidationRule interface {
    Check(elem *Element) error
}

type SizeValidationRule struct {
    MaxSize int
}

func (r *SizeValidationRule) Check(elem *Element) error {
    if len(elem.Content) > r.MaxSize {
        return errors.New("content too large")
    }
    return nil
}
```

### 4. Observer Pattern (Events)
```go
type ElementEvent struct {
    Type      EventType
    Element   *Element
    Timestamp time.Time
}

type EventBus struct {
    subscribers map[EventType][]chan ElementEvent
}

func (b *EventBus) Publish(event ElementEvent) {
    for _, ch := range b.subscribers[event.Type] {
        ch <- event
    }
}
```

### 5. Decorator Pattern (Middleware)
```go
type Middleware func(Handler) Handler

func LoggingMiddleware(next Handler) Handler {
    return HandlerFunc(func(ctx context.Context, req *Request) (*Response, error) {
        log.Info("handling request", "method", req.Method)
        resp, err := next.Handle(ctx, req)
        log.Info("request handled", "error", err)
        return resp, err
    })
}
```

### 6. Builder Pattern (Element Construction)
```go
type ElementBuilder struct {
    element *Element
}

func NewElementBuilder(typ ElementType) *ElementBuilder {
    return &ElementBuilder{
        element: &Element{Type: typ},
    }
}

func (b *ElementBuilder) WithName(name string) *ElementBuilder {
    b.element.Name = name
    return b
}

func (b *ElementBuilder) Build() (*Element, error) {
    // Validate and return
}
```

---

## Fluxo de Dados

### Criação de Elemento

```
Client (Claude Desktop)
    │
    │ JSON-RPC Request
    │ {"method": "tools/call", "params": {"name": "create_element", ...}}
    ▼
┌─────────────────────┐
│  Stdio Transport    │
│  - Read from stdin  │
└──────────┬──────────┘
           │
           │ Deserialize
           ▼
┌─────────────────────┐
│   MCP Server        │
│  - Route to handler │
└──────────┬──────────┘
           │
           │ Dispatch
           ▼
┌─────────────────────┐
│  Tool Registry      │
│  - Find handler     │
└──────────┬──────────┘
           │
           │ Execute
           ▼
┌───────────────────────────────┐
│  CreateElementUseCase         │
│  1. Validate input            │
│  2. Create element entity     │
│  3. Validate element          │
│  4. Save to repository        │
│  5. Index for search          │
│  6. Publish event             │
└──────────┬────────────────────┘
           │
           ├─────────────────┐
           │                 │
           ▼                 ▼
┌────────────MCP SDK Oficial + Clean Architecture

**Status:** Aceito  
**Contexto:** Precisamos de protocol compliance garantido e alta testabilidade  
**Decisão:** Usar `modelcontextprotocol/go-sdk` + Clean Architecture com Hexagonal Architecture  
**Consequências:**
- ✅ Protocol compliance garantido pelo SDK oficial
- ✅ Menos código para manter (SDK cuida do protocolo)
- ✅ Updates automáticos quando MCP spec evolui
- ✅ Testabilidade máxima (domain layer isolado)
- ✅ Fácil substituição de adapters (FS → S3, GitHub → GitLab)
- ❌ Dependência do SDK (mas é oficial e mantido)
│  - Element created  │
└──────────┬──────────┘
           │
           │ Serialize
           ▼
┌─────────────────────┐
│  Stdio Transport    │
│  - Write to stdout  │
└──────────┬──────────┘
           │
           │ JSON-RPC Response
           ▼
Client receives confirmation
```

---

## Decisões Arquiteturais

### ADR-001: Clean Architecture + Hexagonal

**Status:** Aceito  
**Contexto:** Precisamos de alta testabilidade e independência de frameworks  
**Decisão:** Adotar Clean Architecture com Hexagonal Architecture  
**Consequências:**
- ✅ Testabilidade máxima (domain layer isolado)
- ✅ Fácil substituição de adapters (FS → S3, GitHub → GitLab)
- ❌ Mais código boilerplate
- ❌ Curva de aprendizado para novos desenvolvedores

### ADR-002: Repository Pattern

**Status:** Aceito  
**Contexto:** Precisamos desacoplar persistência de lógica de negócio  
**Decisão:** Usar Repository Pattern com interfaces no domain layer  
**Consequências:**
- ✅ Testes sem I/O real
- ✅ Fácil migração entre backends (FS, DB, S3)
- ❌ Overhead de abstração

### ADR-003: Context-Based Lifecycle

**Status:** Aceito  
**Contexto:** Go idiomático requer context.Context para cancelamento  
**Decisão:** Todos os métodos públicos recebem `context.Context`  
**Consequências:**
- ✅ Graceful shutdown
- ✅ Timeout control
- ✅ Request tracing
- ❌ Context propagation em todo código

### ADR-004: Error Wrapping

**Status:** Aceito  
**Contexto:** Precisamos de stack traces e context em erros  
**Decisão:** Usar `fmt.Errorf` com `%w` para wrapping  
**Consequências:**
- ✅ Error chains com context
- ✅ CompatívelDK Oficial + Stdlib-First

**Status:** Aceito  
**Contexto:** Minimizar dependencies externas mas usar SDK oficial para protocol compliance  
**Decisão:** Usar MCP SDK oficial + priorizar stdlib Go para resto  
**Consequências:**
- ✅ Protocol compliance via SDK oficial
- ✅ Menor superfície de ataque (poucas deps além do SDK)
- ✅ Builds rápidos
- ✅ Menos dependências quebradas
- ❌ Dependência crítica no SDK (mas é oficial)

### ADR-006: Auto Schema Generation

**Status:** Aceito  
**Contexto:** Evitar duplicação entre structs Go e JSON Schema manual  
**Decisão:** Gerar JSON Schema automaticamente via reflection usando `invopop/jsonschema`  
**Consequências:**
- ✅ DRY: Schema gerado automaticamente de structs Go
- ✅ Type safety: Schema sempre sincronizado com código
- ✅ Validation tags reutilizadas (jsonschema + validate)
- ✅ Menos código para manter
- ❌ Overhead mínimo de reflection (apenas na inicialização)
- ❌ Schemas complexos podem precisar customização

**Exemplo:**
```go
type CreateElementInput struct {
    Type        string `json:"type" jsonschema:"required,enum=personas|skills" validate:"required"`
    Name        string `json:"name" jsonschema:"required,minLength=3,maxLength=100" validate:"required,min=3,max=100"`
    Description string `json:"description" jsonschema:"required" validate:"required"`
}

// Schema gerado automaticamente:
// {
//   "type": "object",
//   "properties": {
//     "type": {"type": "string", "enum": ["personas", "skills"]},
//     "name": {"type": "string", "minLength": 3, "maxLength": 100},
//     "description": {"type": "string"}
//   },
//   "required": ["type", "name", "description"]
// }
```
- ✅ Builds mais rápidos
- ✅ Menos dependências quebradas
- ❌ Mais código manual (sem libs convenientes)

---

**Próximo Documento:** [Especificação das Ferramentas MCP](./TOOLS_SPEC.md)
