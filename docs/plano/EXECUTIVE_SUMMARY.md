# Executive Summary - DollhouseMCP Go Port

**Data:** 18 de Dezembro de 2025  
**Versão:** 1.0  
**Status:** Planejamento Inicial

## Visão Geral

Este documento apresenta o plano executivo para o desenvolvimento de um servidor MCP (Model Context Protocol) em **Go 1.25**, replicando todas as funcionalidades do [DollhouseMCP](https://github.com/DollhouseMCP/mcp-server) original (TypeScript/Node.js), com melhorias em performance, segurança e arquitetura.

## Objetivos Estratégicos

### 1. Paridade Funcional Completa
- Implementar 100% das funcionalidades do DollhouseMCP v1.9.27+
- 41+ ferramentas MCP para gerenciamento de elementos
- Suporte a 6 tipos de elementos: Personas, Skills, Templates, Agents, Memories, Ensembles
- Sistema de portfolio local e remoto (GitHub)
- Collection community-driven

### 2. Superioridade Técnica
- **Performance:** 10-50x mais rápido que Node.js em operações I/O
- **Concorrência:** Goroutines nativas para operações paralelas
- **Memória:** Footprint 3-5x menor
- **Segurança:** Type safety, memory safety, sem runtime dependencies

### 3. Arquitetura de Excelência
- Clean Architecture + Hexagonal Architecture
- SOLID principles em todo código
- Cobertura de testes mínima: 98%
- Zero dependencies runtime (stdlib only quando possível)

## Escopo do Projeto

### Funcionalidades Core (Fase 1 - 8 semanas)

#### 1. MCP Server Foundation
- **SDK Integration:** Uso do `modelcontextprotocol/go-sdk` oficial
- **Schema Auto-generation:** Geração automática de JSON Schema via reflection (`invopop/jsonschema`)
- **Transport Layer:** 3 transportes prontos do SDK:
  - **Stdio** (padrão) - comunicação via stdin/stdout para Claude Desktop
  - **SSE** (Server-Sent Events) - HTTP streaming para web clients
  - **HTTP** - REST API tradicional para integrações
- Tool registry com auto-discovery e schema automático
- Resource management via SDK
- Error handling padronizado (SDK + custom domain errors)
- Validation tags para input validation automática (`go-playground/validator`)

#### 2. Element System
**Objetivo:** Sistema completo de elementos com 6 tipos distintos

**BaseElement Abstraction:**
- Interface comum para todos os elementos
- Metadata padronizado (ID, version, author, timestamps)
- Lifecycle management (create, activate, deactivate, delete)
- Validation hooks

**Element Types (6 tipos implementados):**

1. **Personas** – Shape how AI acts and responds
   - Behavioral traits configuration
   - Domain expertise definition
   - Response style customization
   - Context-aware behavior switching
   - Hot-swap capability (change persona sem restart)
   - **Private Personas:** User-specific directories (`personas/private-{username}/`)
   - **Collaboration:** Persona sharing, forking, versioning
   - **Bulk Operations:** Mass import/export, template application
   - **Advanced Search:** Multi-criteria filtering, tag-based discovery
   - Metadata: `behavioral_traits`, `expertise_areas`, `tone`, `style`, `privacy_level`, `owner`

2. **Skills** – Add specialized capabilities
   - Discrete procedural knowledge
   - Trigger-based activation (keywords, patterns)
   - Step-by-step procedures
   - Tool integration hooks
   - Composable (skills podem chamar outras skills)
   - Metadata: `triggers`, `procedures`, `dependencies`, `tools_required`
   - Claude Skills compatibility (bidirectional converter)

3. **Templates** – Ensure consistent outputs
   - Variable substitution system
   - Output format standardization
   - Reusable content structures
   - Validation of required variables
   - Multiple format support (Markdown, YAML, JSON)
   - Metadata: `variables`, `format`, `validation_rules`

4. **Agents** – Autonomous task completion
   - Goal-oriented execution
   - Multi-step workflow orchestration
   - Smart decision-making (escolha de tools/skills)
   - Context accumulation across steps
   - Error recovery strategies
   - Metadata: `goals`, `actions`, `decision_tree`, `fallback_strategies`

5. **Memory** – Persistent context storage
   - Text-based storage (YAML format)
   - Date-based folder organization (YYYY-MM-DD)
   - Smart deduplication (SHA-256 hashing)
   - Search indexing (full-text + metadata)
   - Retention policies (permanent, temporary, TTL)
   - Auto-load baseline memories (token budget aware)
   - Metadata: `retention_policy`, `tags`, `privacy_level`, `auto_load`

6. **Ensembles** – Combined element orchestration
   - Multi-element composition (persona + skills + templates)
   - Unified activation/deactivation
   - Dependency resolution
   - Token budget optimization
   - Metadata: `composition`, `activation_order`, `dependencies`

**Common Features (todos os tipos):**
- CRUD operations completas
- Validation engine com 300+ regras de segurança
- Version control integrado (auto-increment)
- Search indexing (inverted index + NLP scoring)
- Export/Import (portability)
- GitHub sync (portfolio integration)

**Private Personas Advanced Features:**
- **User Isolation:** Private directories per user (`personas/private-{username}/`)
- **Access Control:** Owner-only read/write, optional sharing
- **Templates:** Persona templates for quick creation
- **Bulk Operations:** 
  - Batch create from CSV/JSON
  - Mass update (tags, metadata)
  - Bulk export/backup
- **Advanced Search/Filtering:**
  - Multi-criteria queries (author, tags, date range)
  - Fuzzy search (Levenshtein distance)
  - Regex pattern matching
- **Collaboration Tools:**
  - Share personas (read-only or fork)
  - Fork & customize (creates private copy)
  - Versioning (track changes, rollback)
  - Diff viewer (compare versions)
- **Privacy Controls:**
  - Public (community visible)
  - Private (owner only)
  - Shared (specific users/groups)

#### 3. Portfolio System
- Local storage (filesystem hierarchy)
- GitHub integration (OAuth2 device flow)
- Sync bidirectional
- Search indexing (inverted index + NLP scoring)
- Backup/restore automation

#### 4. Collection System
- Community collection browser
- Content installation
- Sharing e submission
- Rating e validation

---

## Modelos de IA Suportados

O servidor MCP suporta integração com os seguintes modelos de IA:

### Modelos Primários

1. **auto** - Seleção automática do melhor modelo com base na capacidade e desempenho
2. **claude-sonnet-4.5** - Claude Sonnet 4.5 (recomendado para uso geral)
3. **claude-haiku-4.5** - Claude Haiku 4.5 (rápido e eficiente)
4. **claude-opus-4.5** - Claude Opus 4.5 (máxima capacidade)
5. **claude-sonnet-4** - Claude Sonnet 4 (versão anterior estável)

### Modelos Google Gemini

6. **gemini-2.5-pro** - Gemini 2.5 Pro
7. **gemini-3-flash-preview** - Gemini 3 Flash Preview (velocidade)
8. **gemini-3-pro-preview** - Gemini 3 Pro Preview

### Modelos OpenAI GPT

9. **gpt-4.1** - GPT-4.1
10. **gpt-4o** - GPT-4o
11. **gpt-5** - GPT-5
12. **gpt-5-mini** - GPT-5 Mini (eficiente)
13. **gpt-5-codex** - GPT-5 Codex (otimizado para código)
14. **gpt-5.1** - GPT-5.1
15. **gpt-5.1-codex** - GPT-5.1 Codex
16. **gpt-5.1-codex-max** - GPT-5.1 Codex Max (máxima capacidade)
17. **gpt-5.1-codex-mini** - GPT-5.1 Codex Mini (eficiente)
18. **gpt-5.2** - GPT-5.2

### Modelos Especializados

19. **grok-code-fast-1** - Grok Code Fast 1 (otimizado para código)
20. **oswe-vscode-prim** - OSWE VSCode Prim (integração VSCode)

**Total:** 20 modelos suportados

**Configuração:** Os modelos são configurados via arquivo de configuração ou variáveis de ambiente. O modo `auto` seleciona automaticamente o modelo mais adequado para cada solicitação.

---

### Funcionalidades Avançadas (Fase 2 - 6 semanas)

#### 5. Security Layer
- Input sanitization (path traversal, injection)
- YAML bomb detection
- Prototype pollution protection
- Rate limiting
- Audit logging
- Encryption (AES-256-GCM)

#### 6. Enhanced Capabilities
- Capability Index com NLP scoring
- Relationship mapping (GraphRAG-style)
- Auto-load baseline memories
- Background validation
- Telemetry (opt-in)

#### 6.1 Private Personas Management
- **User-specific directories:** `personas/private-{username}/`
- **Access control layer:** Owner-based permissions
- **Persona templates:** Reusable base personas
- **Bulk operations API:** Mass CRUD operations
- **Advanced filtering:** Multi-dimensional search
- **Collaboration workflow:**
  - Share URL generation
  - Fork with attribution
  - Version history (Git-like)
  - Merge conflict resolution

#### 7. Advanced Features
- Skills converter (bidirectional Claude Skills)
- Hot-swap elements
- Memory auto-repair
- Unified search (3-tier index)
- Source priority system

### Integrações (Fase 3 - 4 semanas)

#### 8. External Integrations
- GitHub API v3/v4
- OAuth2 device flow
- MCP Registry publishing
- PostHog telemetry (optional)

## Stack Tecnológico

### Core Dependencies
```go
// MCP Protocol SDK (Official)
github.com/modelcontextprotocol/go-sdk/mcp  // MCP protocol types
github.com/modelcontextprotocol/go-sdk/server // MCP Server implementation
github.com/modelcontextprotocol/go-sdk/transport // Transports: stdio, SSE, HTTP

// Schema Generation (Automatic)
github.com/invopop/jsonschema  // JSON Schema from Go structs
github.com/go-playground/validator/v10 // Struct validation tags

// Stdlib-first approach
encoding/json
net/http
os/exec
crypto/*
path/filepath
reflect  // Para schema generation

// Essentials only
gopkg.in/yaml.v3          // YAML parsing
github.com/google/uuid    // UUID generation
golang.org/x/crypto       // Advanced crypto
golang.org/x/oauth2       // OAuth2 flows
```

### Testing & Quality
```go
testing                   // Stdlib testing
github.com/stretchr/testify // Assertions
github.com/golangci/golangci-lint // Linting
```

### Optional (feature-specific)
```go
github.com/dgraph-io/badger/v4 // Embedded KV store (capability index)
github.com/blevesearch/bleve/v2 // Full-text search
```

## Estrutura do Projeto

```
mcp-server/
├── cmd/
│   └── mcp-server/         # Main entry point
│       └── main.go
├── internal/
│   ├── mcp/                # MCP SDK integration
│   │   ├── server/         # SDK server wrapper
│   │   ├── schema/         # Auto schema generation
│   │   └── tools/          # Tool registration
│   ├── elements/           # Element system
│   │   ├── base/
│   │   ├── persona/
│   │   │   ├── private/    # Private personas logic
│   │   │   ├── templates/  # Persona templates
│   │   │   └── collab/     # Collaboration (share/fork)
│   │   ├── skill/
│   │   ├── template/
│   │   ├── agent/
│   │   ├── memory/
│   │   └── ensemble/
│   ├── portfolio/          # Portfolio management
│   │   ├── local/
│   │   ├── github/
│   │   └── sync/
│   ├── collection/         # Community collection
│   ├── security/           # Security layer
│   │   ├── validation/
│   │   ├── sanitization/
│   │   └── encryption/
│   ├── capability/         # Capability index
│   │   ├── nlp/
│   │   └── graph/
│   └── telemetry/          # Optional telemetry
├── pkg/                    # Public APIs
│   ├── client/             # MCP client library
│   └── types/              # Shared types
├── api/                    # OpenAPI specs
│   └── openapi.yaml
├── docs/                   # Documentation
│   ├── plano/              # This planning doc
│   ├── architecture/       # ADRs
│   └── guides/             # User guides
├── examples/               # Usage examples
├── test/
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
├── scripts/                # Automation scripts
├── .github/
│   └── workflows/          # CI/CD
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
└── README.md
```

## Cronograma Executivo

### Fase 1: Foundation (Semanas 1-8)
**Objetivo:** MCP Server funcional com elementos básicos

| Semana | Entregas |
|--------|----------|
| 1-2 | MCP SDK integration + schema auto-gen framework + tool registry |
| 3 | Element system base: BaseElement, common interfaces, validation framework |
| 4 | **Persona** + **Skill** + **Template** elements (3 tipos core) |
| 5 | Portfolio local storage + filesystem adapter + basic search |
| 6 | GitHub integration (OAuth2 device flow) + bidirectional sync |
| 7 | Collection browser + content installation |
| 8 | Integration tests + cobertura 95%+ + documentação |

### Fase 2: Advanced Features (Semanas 9-14)
**Objetivo:** Paridade funcional completa

| Semana | Entregas |
|--------|----------|
| 9 | **Agent** element (goal-oriented execution + workflow orchestration) |
| 10 | **Memory** element (YAML storage + date folders + retention policies) |
| 11 | **Ensemble** element (composition + dependency resolution) |
| 12 | Security layer completa (300+ validation rules + encryption) |
| 13 | **Private Personas:** User directories + access control + templates |
| 14 | **Collaboration Tools:** Sharing + forking + versioning + bulk ops |
| 15 | Capability index (NLP scoring: Jaccard + Shannon Entropy) |
| 16 | Relationship graph (GraphRAG-style) + auto-load memories |

### Fase 3: Polish & Production (Semanas 17-20)
**Objetivo:** Production-ready release

| Semana | Entregas |
|--------|----------|
| 17-18 | Skills converter + telemetry + advanced search/filtering |
| 19 | Performance tuning + security audit |
| 20 | Documentation + examples + v1.0.0 release |

## Métricas de Sucesso

### Performance Targets
- Startup time: < 50ms (vs. ~500ms Node.js)
- Memory footprint: < 50MB (vs. ~150MB Node.js)
- Element load: < 1ms per element (vs. ~5ms Node.js)
- Search query: < 10ms for 1000 elements (vs. ~50ms Node.js)

### Quality Targets
- Test coverage: ≥ 98%
- Linting: Zero issues (golangci-lint)
- Security: Zero vulnerabilities (Snyk/Dependabot)
- Documentation: 100% public APIs documented

### Compatibility Targets
- MCP Protocol: 100% compliant
- DollhouseMCP elements: 100% compatible
- Claude Desktop: Full integration
- Cross-platform: Linux, macOS, Windows

## Features Detalhadas por Tipo de Elemento

### 1. Personas (Semana 4) + Private Personas (Semanas 13-14)
**Capacidades:**
- Define behavioral traits (curiosity, precision, creativity, etc.)
- Configura expertise areas (coding, writing, analysis, etc.)
- Ajusta response style (formal, casual, technical, etc.)
- Hot-swap entre personas sem restart do servidor
- Ativação condicional baseada em contexto

**Private Personas Advanced Features:**
- **User Isolation:**
  - Directory structure: `personas/private-{username}/`
  - Automatic user detection from context
  - Owner-only access by default
  - Shared personas in `personas/shared/`
  
- **Templates System:**
  - Base templates: `personas/templates/`
  - Quick creation: `create_from_template(template_id, customizations)`
  - Template marketplace (community templates)
  
- **Bulk Operations:**
  - Batch import: `bulk_import_personas(csv_file)` → creates multiple
  - Mass update: `bulk_update_personas(filter, updates)` → updates matching
  - Bulk export: `bulk_export_personas(filter)` → CSV/JSON download
  - Duplicate detection: SHA-256 content hashing
  
- **Advanced Search/Filtering:**
  - Multi-criteria: `search_personas(author="alice", tags=["technical"], date_range="2025-12")`
  - Fuzzy search: Levenshtein distance ≤ 2 for typo tolerance
  - Regex patterns: `search_personas(name_pattern="^dev-.*")`
  - NLP scoring: Relevance ranking (TF-IDF + semantic similarity)
  
- **Collaboration Tools:**
  - **Share:** Generate read-only link → `share_persona(id, permissions={"read": true})`
  - **Fork:** Create private copy → `fork_persona(id)` → `personas/private-bob/forked-from-alice-persona`
  - **Versioning:** Git-like history → `list_versions(id)`, `rollback(id, version)`
  - **Diff:** Compare versions → `diff_personas(id, v1, v2)` → shows changes
  - **Merge:** Conflict resolution for collaborative edits

**Implementação:**
```go
type Persona struct {
    Element
    BehavioralTraits []string
    ExpertiseAreas   []string
    ResponseStyle    string
    Tone             string
    
    // Private Personas fields
    PrivacyLevel     PrivacyLevel  // public, private, shared
    Owner            string        // username
    SharedWith       []string      // usernames with access
    TemplateID       string        // if created from template
    ForkedFrom       string        // if forked from another persona
    VersionHistory   []Version     // Git-like history
}

type PrivacyLevel string
const (
    PrivacyPublic  PrivacyLevel = "public"   // Community visible
    PrivacyPrivate PrivacyLevel = "private"  // Owner only
    PrivacyShared  PrivacyLevel = "shared"   // Specific users
)

type Version struct {
    Number    int
    Timestamp time.Time
    Author    string
    Message   string
    ContentHash string // SHA-256 of content
}
```

**Storage Structure:**
```
portfolio/
├── personas/
│   ├── public/              # Community personas
│   │   └── creative-writer.md
│   ├── private-alice/       # Alice's private personas
│   │   ├── work-persona.md
│   │   └── personal-assistant.md
│   ├── private-bob/         # Bob's private personas
│   │   └── forked-from-alice-creative-writer.md
│   ├── shared/              # Explicitly shared
│   │   └── team-persona.md
│   └── templates/           # Persona templates
│       ├── developer-template.md
│       └── writer-template.md
```

### 2. Skills (Semana 4)
**Capacidades:**
- Define procedimentos passo a passo
- Trigger-based activation (keywords detectam skill relevante)
- Composabilidade (skills podem chamar outras skills)
- Tool integration (skills podem usar ferramentas externas)
- **Claude Skills Converter** (bidirectional compatibility)

**Implementação:**
```go
type Skill struct {
    Element
    Triggers     []string
    Procedures   []Procedure
    Dependencies []string
    ToolsRequired []string
}
```

### 3. Templates (Semana 4)
**Capacidades:**
- Variable substitution com validação
- Multiple format support (Markdown, YAML, JSON)
- Required vs optional variables
- Nested template support
- Output validation contra schema

**Implementação:**
```go
type Template struct {
    Element
    Variables      []TemplateVariable
    Format         string
    ValidationRules []ValidationRule
}
```

### 4. Agents (Semana 9)
**Capacidades:**
- Goal-oriented autonomous execution
- Multi-step workflow orchestration
- Smart decision-making (escolhe tools/skills dinamicamente)
- Context accumulation (memória entre steps)
- Error recovery com fallback strategies
- Parallel task execution

**Implementação:**
```go
type Agent struct {
    Element
    Goals              []string
    Actions            []Action
    DecisionTree       *DecisionTree
    FallbackStrategies []Strategy
    MaxIterations      int
}
```

### 5. Memory (Semana 10)
**Capacidades:**
- Persistent context storage (survive sessions)
- Date-based organization (YYYY-MM-DD folders)
- Smart deduplication (SHA-256 content hashing)
- Full-text search indexing
- Retention policies (permanent, 30 days, 90 days, TTL)
- Privacy levels (public, private, sensitive)
- **Auto-load baseline memories** (token budget aware)
- Background validation (trust levels)

**Implementação:**
```go
type Memory struct {
    Element
    Entries         []MemoryEntry
    RetentionPolicy string
    Tags            []string
    PrivacyLevel    string
    AutoLoad        bool
    TrustLevel      string
}
```

### 6. Ensembles (Semana 11)
**Capacidades:**
- Multi-element composition (persona + skills + templates)
- Unified activation (ativa todos de uma vez)
- Dependency resolution (ordem correta)
- Token budget optimization (distribui tokens entre elementos)
- Conflict detection (evita incompatibilidades)

**Implementação:**
```go
type Ensemble struct {
    Element
    Composition     []ElementReference
    ActivationOrder []string
    Dependencies    map[string][]string
    TokenBudget     int
}
```

## Diferenciais Competitivos

### vs. DollhouseMCP Original (TypeScript)

| Aspecto | DollhouseMCP (TS) | Go Port | Ganho |
|---------|-------------------|---------|-------|
| Startup | ~500ms | <50ms | 10x |
| Memory | ~150MB | <50MB | 3x |
| Performance | Baseline | 10-50x | 10-50x |
| Concurrency | Event loop | Goroutines | Superior |
| Type Safety | TypeScript | Go | Compilado |
| Dependencies | 50+ npm packages | <10 Go modules | Minimal |
| Binary Size | N/A (runtime) | ~15MB | Portable |
| Cross-compile | Não | Sim | Superior |

### Recursos Exclusivos do Go Port

1. **Single Binary Distribution**
   - Zero runtime dependencies
   - Cross-compilation nativa
   - Deploy simplificado

2. **Superior Concurrency**
   - Parallel element loading
   - Concurrent GitHub sync
   - Background validation sem blocking

3. **Multiple Transports**
   - Stdio (Claude Desktop)
   - SSE (Web clients)
   - HTTP (REST integrations)
   - Pluggable architecture

4. **Enhanced Security**
   - Memory safety garantida
   - No prototype pollution
   - Compile-time type checks

5. **Better Resource Management**
   - Automatic garbage collection
   - Efficient memory pooling
   - Lower CPU utilization

## Riscos e Mitigações

### Riscos Técnicos

| Risco | Probabilidade | Impacto | Mitigação |
|-------|--------------|---------|-----------|
| MCP protocol changes | Média | Alto | Testes contínuos contra spec oficial |
| GitHub API rate limits | Alta | Médio | Cache agressivo + exponential backoff |
| YAML parsing vulnerabilities | Baixa | Alto | Input validation + size limits |
| Cross-platform filesystem issues | Média | Médio | Extensive testing em 3 OS |

### Riscos de Projeto

| Risco | Probabilidade | Impacto | Mitigação |
|-------|--------------|---------|-----------|
| Scope creep | Alta | Alto | MVP iterativo + feature freeze dates |
| Underestimated complexity | Média | Alto | 20% buffer em cada fase |
| Dependency vulnerabilities | Baixa | Médio | Dependabot + manual audits |

## Próximos Passos

### Imediato (Próximas 48h)
1. ✅ Criar estrutura de diretórios do projeto
2. ✅ Inicializar go.mod com SDK oficial: `github.com/modelcontextprotocol/go-sdk`
3. ✅ Configurar CI/CD (GitHub Actions)
4. ✅ Setup MCP Server usando SDK + schema auto-generation

### Curto Prazo (Semana 1)
1. MCP Server setup com SDK oficial (stdio transport como padrão)
2. Schema auto-generation framework:
   - Reflection engine com `reflect` package
   - JSON Schema generation via `invopop/jsonschema`
   - Struct tags para validação (`validate`, `jsonschema`)
3. Tool registration com auto-discovery
4. Primeiro tool: `list_elements` com schema 100% automático
5. Testes unitários (cobertura 95%+)
6. Suporte para os 3 transportes (stdio, SSE, HTTP)

### Médio Prazo (Semanas 2-4)
1. Element system completo (Persona, Skill, Template)
2. Portfolio local storage
3. Basic validation
4. Integration tests

## Recursos Necessários

### Equipe
- 1 Senior Go Engineer (full-time)
- 1 DevOps Engineer (25% allocation para CI/CD)
- 1 QA Engineer (25% allocation para testes)

### Infraestrutura
- GitHub Actions (CI/CD)
- SonarCloud (code quality)
- Dependabot (security)
- Docker Hub (container registry)

### Ferramentas
- golangci-lint
- go test + testify
- Swagger/OpenAPI
- MCP Inspector (debugging)

## Conclusão

Este projeto representa uma oportunidade única de criar uma implementação **superior** do DollhouseMCP, aproveitando os pontos fortes do Go:

- **Performance incomparável** para operações I/O intensivas
- **Segurança nativa** com type safety e memory safety
- **Deploy simplificado** com single binary
- **Melhor experiência de desenvolvimento** com tooling moderno

Com um cronograma realista de 18 semanas e foco em qualidade (98% coverage), o resultado será um servidor MCP **production-ready**, **performante** e **maintainável** que estabelece um novo padrão para servidores MCP em Go.

---

**Próximo Documento:** [Arquitetura Técnica Detalhada](./ARCHITECTURE.md)
