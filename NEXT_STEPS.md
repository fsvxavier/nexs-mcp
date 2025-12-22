# NEXS-MCP - Roadmap de Desenvolvimento

**Data de Atualização:** 22 de dezembro de 2025  
**Versão Atual:** v1.0.5  
**Próxima Meta:** v2.0.0 - Enterprise Features + Vector Search + Advanced Memory Management

---

## 📊 Status Atual

### ✅ Base Implementada (v1.0.5 + Relationships + Tests)
- 6 tipos de elementos (Persona, Skill, Agent, Memory, Template, Ensemble)
- 71 MCP Tools (66 base + 5 relacionamentos)
- Arquitetura Limpa Go
- GitHub Integration (OAuth, sync, PR)
- Collection System (registry, cache)
- Ensembles (monitoring, voting, consensus)
- Context Enrichment System
- **Sistema Avançado de Relacionamentos** ✨ NOVO
  - Busca bidirecional com índice invertido O(1)
  - Inferência automática (4 métodos: mention, keyword, semantic, pattern)
  - Expansão recursiva multi-nível (depth 1-5)
  - Recommendation engine (4 estratégias de scoring)
  - Cache LRU com métricas (hits/misses)
- **Cobertura de Testes Abrangente** ✅ COMPLETO
  - **63.2% cobertura total** do projeto
  - **425+ testes novos** em 17 arquivos
  - Zero race conditions (race detector ✓)
  - Zero linter issues (golangci-lint ✓)
  - Timeout otimizado (120s) para race detection
- Multilíngue (11 idiomas)
- NPM Distribution (@fsvxavier/nexs-mcp-server)

### ✨ Sistema Avançado de Relacionamentos (Implementado - 22/12/2025)

**Arquivos Criados/Modificados:**
- `internal/application/relationship_index.go` - Expansão recursiva e busca bidirecional
- `internal/application/relationship_inference.go` - Motor de inferência (566 linhas)
- `internal/domain/agent.go` - Métodos helper para relacionamentos
- `internal/domain/persona.go` - Métodos helper para relacionamentos
- `internal/domain/template.go` - Métodos helper para relacionamentos
- `internal/mcp/relationship_tools.go` - 5 novos MCP tools
- `test/integration/relationships_integration_test.go` - 6 testes (100% passando)

**MCP Tools Adicionados:**
1. `get_related_elements` - Busca bidirecional com filtros (forward/reverse/both)
2. `expand_relationships` - Expansão recursiva até 5 níveis
3. `infer_relationships` - Inferência automática multi-método
4. `get_recommendations` - Recomendações inteligentes com scoring
5. `get_relationship_stats` - Estatísticas do índice

**Funcionalidades Implementadas:**
- ✅ Busca bidirecional (GetMemoriesRelatedTo) com O(1) lookups
- ✅ Índice invertido para relacionamentos
- ✅ Cross-element relationships (Persona → Skills, Agent → Persona)
- ✅ Relationship inference from content (4 métodos: mention, keyword, semantic, pattern)
- ✅ Multi-level depth expansion (recursive, depth 1-5)
- ✅ Context caching (LRU, TTL 5min, auto-invalidation)
- ✅ Recommendation engine (4 estratégias de scoring)

**Performance & Qualidade:**
- O(1) lookups com índice invertido
- Cache LRU com métricas (hits/misses/hit rate)
- 6 testes de integração (100% passando)
- Zero erros de compilação
- Suporta grafos profundos sem degradação

### 🎯 Objetivos v2.0.0

**Meta:** Paridade enterprise com competidores + Diferenciais técnicos únicos  
**Timeline:** Janeiro 2026 - Junho 2026 (24 semanas)

---

---

## 📜 Histórico de Implementações

### Release v1.0.6 - 22 de dezembro de 2025

#### Cobertura de Testes Abrangente
- ✅ **17 Arquivos de Teste Criados**: 425+ testes novos em internal/
- ✅ **Cobertura Total**: 63.2% do código (aumento de ~30%)
- ✅ **Pacotes Testados**:
  - `internal/backup/restore_test.go` - 14 testes
  - `internal/infrastructure/github_publisher_test.go` - 21 testes
  - `internal/mcp/relationship_tools_test.go` - 26 testes
  - `internal/common/constants_test.go` - 6 testes
  - `internal/collection/security/` - 102 testes (checksum, scanner, signature, sources)
  - `internal/template/validator_test.go` - 35 testes
  - `internal/template/stdlib/loader_test.go` - 27 testes
  - `internal/collection/validator_test.go` - 25 testes
  - `internal/mcp/*_tools_test.go` - 200+ testes (9 arquivos)
- ✅ **Qualidade de Código**:
  - Zero race conditions (race detector com -race flag)
  - Zero linter issues (goconst corrigido)
  - Template format constants criados (FormatMarkdown, FormatYAML, FormatJSON, FormatText)
  - Timeout aumentado de 30s → 120s para suportar race detection em test-coverage
- ✅ **Performance**:
  - internal/mcp: 46.5s com race detection (62.5% coverage)
  - internal/template: 87.0% coverage
  - internal/portfolio: 75.6% coverage
  - internal/validation: 66.3% coverage

### Release v1.0.5 - 21 de dezembro de 2025

#### Automação de Release e Distribuição NPM
- ✅ **Pacote NPM Publicado**: [@fsvxavier/nexs-mcp-server@1.0.5](https://www.npmjs.com/package/@fsvxavier/nexs-mcp-server)
- ✅ **GitHub Release Automation**: Comando `make github-publish` criado e funcional
- ✅ **Stop Words Portuguesas**: Lista expandida para melhor extração de keywords
- ✅ **Makefile**: Comandos npm-publish e github-publish com verificação

### Release v1.0.2 - 21 de dezembro de 2025

#### Correções de Qualidade de Código
- ✅ **Linter Issues**: 69 issues → 0 (goconst, gocritic, usetesting, staticcheck, ineffassign, gocyclo)
- ✅ **Complexidade Ciclomática**: Reduzida de 91 para < 35 em todas as funções
- ✅ **Test Patterns**: Modernizados (t.TempDir, require.NoError)
- ✅ **Type-Safe Context Keys**: Custom type para prevenir colisões

### Implementações Anteriores (v1.0.0 - v1.0.1)

#### GitHub Integration ✅ COMPLETO
- Token storage persistente com criptografia AES-256-GCM
- Portfolio sync (push/pull) com detecção de conflitos
- PR submission workflow com template automático
- Tracking de PRs com 4 status (pending, merged, rejected, draft)
- Sync incremental com metadata tracking

**Arquivos:**
- `internal/infrastructure/github_oauth.go` (220 lines)
- `internal/infrastructure/crypto.go` (166 lines)
- `internal/infrastructure/sync_conflict_detector.go` (248 lines)
- `internal/infrastructure/sync_metadata.go` (318 lines)
- `internal/infrastructure/sync_incremental.go` (412 lines)
- `internal/infrastructure/pr_tracker.go` (384 lines)
- `docs/templates/pr_template.md` (102 lines)

#### Collection System ✅ COMPLETO
- Browse/search robusto com filtros avançados
- Cache de collection com TTL configurável (24h default)
- Offline mode com fallback para cache
- Registry com RegistryCache struct
- Installer e validator completos

**Arquivos:**
- `internal/collection/manager.go` (browser functionality)
- `internal/collection/registry.go` (cache functionality)
- `internal/collection/installer.go`
- `internal/collection/validator.go`
- `internal/mcp/collection_tools.go`

#### Ensembles ✅ COMPLETO
- Execution engine com 3 modos (sequential, parallel, hybrid)
- 6 estratégias de agregação (first, last, consensus, voting, all, merge)
- Monitoring real-time com progress tracking
- Voting strategies completos (weighted, threshold, confidence-based)
- 5 MCP tools de ensemble

**Arquivos:**
- `internal/application/ensemble_executor.go` (509 lines)
- `internal/application/ensemble_monitor.go` (250 lines)
- `internal/application/ensemble_aggregation.go` (420 lines)
- `internal/mcp/ensemble_execution_tools.go` (218 lines)
- **Total:** 75 testes passando no pacote application

#### Distribution ✅ COMPLETO
- **Go Module**: v1.0.5 publicado, disponível via `go install`
- **Docker**: Imagem 14.5 MB no Docker Hub (fsvxavier/nexs-mcp)
- **NPM**: @fsvxavier/nexs-mcp-server@1.0.5 com binários multi-plataforma
- **Homebrew**: Formula disponível no tap fsvxavier/nexs-mcp
- **CI/CD**: Workflows completos (release, docker, npm, homebrew)

#### Documentation ✅ COMPLETO
- User Guide: Getting Started, Quick Start, Troubleshooting (2,000+ lines)
- Developer Docs: Code Tour, Testing, Setup, Release
- API Docs: CLI, Context Enrichment, MCP Resources/Tools
- Architecture: Domain, Application, Infrastructure, MCP
- 10+ ADRs (Architecture Decision Records)

#### Context Enrichment System ✅ IMPLEMENTADO (Sprint 1-4)
- Bidirectional search e índice invertido
- Cross-element relationships
- Relationship inference (4 métodos)
- Multi-level expansion recursiva (depth 1-5)
- Context caching (LRU, TTL 5min)
- Recommendation engine (4 estratégias)
- TF-IDF indexing para semantic similarity
- Statistics tracking

**Arquivos:**
- `internal/application/relationship_index.go`
- `internal/application/relationship_inference.go` (566 lines)
- `internal/application/recommendation_engine.go`
- `internal/application/context_enrichment.go`
- `internal/mcp/relationship_tools.go` (5 MCP tools)
- `test/integration/relationships_integration_test.go` (6 tests, 100% passing)

---

## 1. Análise de Gaps Competitivos

**Referência:** [docs/analysis/COMPETITIVE_ANALYSIS_MEMORY_MCP.md](docs/analysis/COMPETITIVE_ANALYSIS_MEMORY_MCP.md)

### 1.1 Projetos Competidores Analisados

1. **Memento MCP** (TypeScript/Neo4j) - Vector search + Temporal features complete
2. **Zero-Vector v3** (JavaScript) - HNSW + Memory-efficient vector storage
3. **Agent Memory** (Python/Redis) - Two-tier memory + Enterprise auth
4. **simple-memory-mcp** (JavaScript) - Obsidian integration + One-click install
5. **mcp-memory-service** (Python/SQLite) - Hybrid backend + Memory quality (ONNX)

### 1.2 Gaps Críticos Identificados

#### Features que TODOS os competidores enterprise têm:

❌ **Vector Embeddings + Semantic Search**
- Competidores: Memento, Zero-Vector, Agent Memory, MCP Memory Service
- Impacto: CRÍTICO - Diferencial competitivo essencial
- Status: Não implementado

❌ **HNSW Index (Approximate NN)**
- Competidores: Zero-Vector, Agent Memory, MCP Memory Service
- Impacto: ALTO - Performance em escala (sub-100ms queries)
- Status: Atualmente usando TF-IDF (lento em >10k memories)

❌ **Memory Quality System**
- Competidores: MCP Memory Service (ONNX local)
- Impacto: ALTO - Gestão inteligente de retenção
- Status: Não implementado

❌ **Two-Tier Memory Architecture**
- Competidores: Agent Memory
- Impacto: ALTO - Working (session) + Long-term (persistent)
- Status: Apenas memória persistente única

❌ **Temporal Features Complete**
- Competidores: Memento (complete cycle)
- Impacto: MÉDIO - Version history + Time-travel + Decay
- Status: Apenas timestamps básicos

❌ **Confidence Decay System**
- Competidores: Memento, MCP Memory Service
- Impacto: MÉDIO - Time-based scoring automático
- Status: Não implementado

❌ **OAuth2/JWT Authentication**
- Competidores: Agent Memory, MCP Memory Service
- Impacto: ALTO - Enterprise adoption blocker
- Status: Não implementado

❌ **Hybrid Backend**
- Competidores: MCP Memory Service
- Impacto: MÉDIO - Local performance + Cloud backup
- Status: SQLite local apenas

❌ **Background Task System**
- Competidores: Agent Memory, MCP Memory Service
- Impacto: MÉDIO - Async processing (consolidation, cleanup)
- Status: Não implementado

❌ **Obsidian Export**
- Competidores: simple-memory-mcp
- Impacto: BAIXO - Convenience feature
- Status: Não implementado

❌ **One-Click Install**
- Competidores: simple-memory-mcp
- Impacto: MÉDIO - User onboarding
- Status: Manual installation apenas

❌ **Web Dashboard**
- Competidores: MCP Memory Service
- Impacto: MÉDIO - Visual management
- Status: CLI apenas

---

## 2. Roadmap de Implementação

### Timeline Geral: 24 semanas (Janeiro - Junho 2026)

**Prioridades:**
- **P0 (Sprints 5-8):** Features críticas para paridade enterprise
- **P1 (Sprints 9-12):** Features importantes para competitividade
- **P2 (Sprints 13-17):** Features de diferenciação e UX

---

## 3. Sprint 5 (Semanas 9-10): Vector Embeddings Foundation

**Duração:** 12 dias úteis  
**Prioridade:** P0 - CRÍTICO  
**Objetivo:** Implementar múltiplos providers de embeddings com semantic search

### 3.1 Features a Desenvolver

#### 3.1.1 Multiple Embedding Providers (8 dias)

**Provider 1: OpenAI** (2 dias)
- [ ] Integração OpenAI API (text-embedding-3-small)
- [ ] Dimensões: 1536
- [ ] Rate limiting e retry logic
- [ ] Error handling robusto
- **Arquivos:** `internal/embeddings/providers/openai.go`

**Provider 2: Local Transformers - DEFAULT** (3 dias)
- [ ] Integração all-MiniLM-L6-v2
- [ ] Dimensões: 384
- [ ] Zero custo, full privacy
- [ ] Offline-capable
- **Arquivos:** `internal/embeddings/providers/transformers.go`

**Provider 3: Sentence Transformers** (2 dias)
- [ ] Integração paraphrase-multilingual
- [ ] Support para 50+ idiomas
- [ ] Compatível com 11 idiomas do NEXS
- **Arquivos:** `internal/embeddings/providers/sentence.go`

**Provider 4: ONNX Runtime** (1 dia)
- [ ] Integração ms-marco-MiniLM (23MB)
- [ ] CPU/GPU acceleration
- [ ] 50-100ms latency (CPU), 10-20ms (GPU)
- **Arquivos:** `internal/embeddings/providers/onnx.go`

**Provider Abstraction** (incluído acima)
- [ ] Factory pattern para criar providers
- [ ] Fallback automático: OpenAI → Transformers → Sentence → ONNX
- [ ] Configuration via env vars
- **Arquivos:** `internal/embeddings/factory.go`, `internal/embeddings/provider.go`

#### 3.1.2 Semantic Search API (4 dias)

- [ ] Vector similarity search (cosine/euclidean/dot product)
- [ ] Batch embedding generation
- [ ] Embedding cache (TTL configurável)
- [ ] Integration com todos providers
- [ ] MCP tools: `semantic_search`, `find_similar_memories`
- **Arquivos:** `internal/application/semantic_search.go`, `internal/vectorstore/store.go`

### 3.2 Entregáveis

- [ ] `internal/embeddings/` - Package completo com 4 providers
- [ ] `internal/vectorstore/` - Vector storage abstraction
- [ ] `internal/application/semantic_search.go` - Semantic search service
- [ ] 2+ MCP tools novos
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests

### 3.3 Dependências Necessárias

```go
// go.mod additions
require (
    github.com/sashabaranov/go-openai v1.17.9          // OpenAI embeddings
    github.com/nlpodyssey/spago v1.1.0                 // Local Transformers
    github.com/james-bowman/nlp v0.0.0                 // Sentence Transformers
    github.com/yalue/onnxruntime_go v1.8.0             // ONNX Runtime
)
```

### 3.4 Métricas de Sucesso

- [ ] 4 providers funcionais com testes
- [ ] Semantic search accuracy >85% vs TF-IDF
- [ ] Latência <500ms para embedding generation (batch de 10)
- [ ] Zero breaking changes em APIs existentes

---

## 4. Sprint 6 (Semanas 11-12): HNSW Performance

**Duração:** 12 dias úteis  
**Prioridade:** P0 - CRÍTICO  
**Objetivo:** Implementar HNSW index para queries sub-100ms em escala

### 4.1 Features a Desenvolver

#### 4.1.1 HNSW Index Implementation (7 dias)

**Hierarchical Navigable Small World Algorithm:**
- [ ] HNSW graph construction
- [ ] Parâmetros: M=16 connections, efConstruction=200, efSearch=50
- [ ] Approximate nearest neighbor search
- [ ] Sub-50ms queries para 10k+ vectors
- [ ] Support 349k+ vectors capacity
- [ ] Incremental index updates (add/remove vectors)
- **Arquivos:** `internal/indexing/hnsw/graph.go`, `internal/indexing/hnsw/search.go`

#### 4.1.2 Integration com Semantic Search (3 dias)

- [ ] Hybrid search: HNSW + metadata filtering
- [ ] Index persistence (save/load from disk)
- [ ] Automatic reindexing triggers
- [ ] Threshold: 100 vectors para criar índice
- [ ] Fallback para linear search (<100 vectors)
- **Arquivos:** `internal/application/hybrid_search.go`

#### 4.1.3 Benchmark Suite (2 dias)

- [ ] TF-IDF vs Vector Search vs HNSW comparison
- [ ] Latency benchmarks (1k, 10k, 100k vectors)
- [ ] Memory usage profiling
- [ ] Accuracy vs speed trade-off analysis
- **Arquivos:** `benchmark/vector_search_test.go`

### 4.2 Entregáveis

- [ ] `internal/indexing/hnsw/` - HNSW implementation completa
- [ ] Integration tests com 10k+ vectors
- [ ] Performance benchmarks com relatório
- [ ] Documentation: HNSW parameter tuning guide

### 4.3 Dependências Necessárias

```go
require (
    github.com/Bithack/go-hnsw v0.0.0-20211102081019   // HNSW index
)
```

### 4.4 Métricas de Sucesso

- [ ] <50ms queries para 10k vectors
- [ ] <200ms queries para 100k vectors
- [ ] Accuracy >95% vs linear search
- [ ] Memory overhead <50MB para 10k vectors (384 dims)

---

## 5. Sprint 7 (Semanas 13-14): Two-Tier Memory

**Duração:** 10 dias úteis  
**Prioridade:** P0 - CRÍTICO  
**Objetivo:** Separar working memory (session) de long-term memory (persistent)

### 5.1 Features a Desenvolver

#### 5.1.1 Working Memory Model (5 dias)

**Working Memory:**
- [ ] Session-scoped storage (in-memory)
- [ ] TTL configurável (default: 1 hora)
- [ ] Automatic expiration
- [ ] Fast access (<1ms)
- [ ] Context: messages, structured memories, metadata
- **Arquivos:** `internal/domain/working_memory.go`, `internal/infrastructure/working_memory_store.go`

**Long-Term Memory:**
- [ ] Persistent storage (SQLite)
- [ ] Semantic indexing
- [ ] Topic modeling (opcional)
- [ ] Entity recognition (opcional)
- **Nota:** Já implementado, apenas refatorar interface

#### 5.1.2 Memory Promotion Logic (3 dias)

- [ ] Automatic promotion rules (working → long-term):
  - Importance score ≥0.7
  - Referenced multiple times
  - Explicitly marked by user
- [ ] Manual promotion MCP tool
- [ ] Batch promotion background task
- **Arquivos:** `internal/application/memory_promotion.go`

#### 5.1.3 MCP Tools Integration (2 dias)

**Novos MCP Tools (15+):**
- [ ] `store_working_memory` - Adicionar à working memory
- [ ] `get_working_memory` - Buscar working memory
- [ ] `list_working_memories` - Listar todas working memories
- [ ] `promote_to_longterm` - Promover manualmente
- [ ] `clear_working_memory` - Limpar session
- [ ] Atualizar tools existentes para suportar tier selection
- **Arquivos:** `internal/mcp/working_memory_tools.go`

### 5.2 Entregáveis

- [ ] Two-tier architecture completa
- [ ] 15+ new MCP tools
- [ ] Migration guide (single-tier → two-tier)
- [ ] Unit + integration tests

### 5.3 Métricas de Sucesso

- [ ] <1ms access para working memory
- [ ] Automatic promotion >90% accuracy
- [ ] Zero data loss durante promotion
- [ ] Backward compatibility com single-tier

---

## 6. Sprint 8 (Semanas 15-16): Memory Quality (ONNX)

**Duração:** 12 dias úteis  
**Prioridade:** P0 - CRÍTICO  
**Objetivo:** Sistema de quality scoring com ONNX local + Multi-tier fallback

### 6.1 Features a Desenvolver

#### 6.1.1 Local ONNX Quality Scoring (5 dias)

**ONNX Runtime Integration:**
- [ ] ms-marco-MiniLM-L-6-v2 model (23MB download)
- [ ] Quality score prediction (0.0-1.0)
- [ ] CPU optimization (50-100ms latency)
- [ ] GPU acceleration support (10-20ms latency)
- [ ] Zero cost, full privacy, offline-capable
- **Arquivos:** `internal/quality/onnx.go`, `models/ms-marco-MiniLM-L-6-v2.onnx`

#### 6.1.2 Multi-Tier Fallback System (5 dias)

**Fallback Chain:**
1. [ ] **ONNX** (local SLM, default)
2. [ ] **Groq API** (fast cloud inference)
3. [ ] **Gemini API** (high-quality fallback)
4. [ ] **Implicit Signals** (fallback of last resort)

**Implicit Signals:**
- [ ] Recency (age of memory)
- [ ] Access frequency
- [ ] Reference count
- [ ] User ratings (if available)
- **Arquivos:** `internal/quality/fallback.go`, `internal/quality/implicit.go`

#### 6.1.3 Quality-Based Retention Policies (2 dias)

**Retention Rules:**
- [ ] High quality (≥0.7): 365 days retention
- [ ] Medium quality (0.5-0.7): 180 days retention
- [ ] Low quality (<0.5): 30-90 days retention
- [ ] Automatic archival (não deletion)
- [ ] Background cleanup task (scheduled)
- **Arquivos:** `internal/application/memory_retention.go`

### 6.2 Entregáveis

- [ ] `internal/quality/` - Quality system completo
- [ ] ONNX model integration
- [ ] Multi-tier fallback working
- [ ] MCP tool: `score_memory_quality`
- [ ] Retention policy engine

### 6.3 Dependências Necessárias

```go
require (
    github.com/yalue/onnxruntime_go v1.8.0             // ONNX Runtime (já adicionado Sprint 5)
)
```

### 6.4 Métricas de Sucesso

- [ ] ONNX scoring accuracy >85% vs Groq
- [ ] 50-100ms latency (CPU)
- [ ] <1% fallback rate para Groq/Gemini
- [ ] Quality distribution curve saudável (bell curve)

---

## 7. Sprint 9 (Semanas 17-18): Enterprise Auth

**Duração:** 15 dias úteis  
**Prioridade:** P1 - IMPORTANTE  
**Objetivo:** OAuth2/JWT authentication para enterprise adoption

### 7.1 Features a Desenvolver

#### 7.1.1 OAuth2 Multi-Provider (10 dias)

**Supported Providers:**
- [ ] Auth0
- [ ] AWS Cognito
- [ ] Okta
- [ ] Azure AD
- [ ] Google Workspace (opcional)

**OAuth2 Features:**
- [ ] Dynamic Client Registration (RFC 7591)
- [ ] OpenID Connect Discovery (RFC 8414)
- [ ] Token refresh automático
- [ ] Session management
- **Arquivos:** `internal/infrastructure/auth/oauth2.go`

#### 7.1.2 JWT Authentication (3 dias)

- [ ] JWT token generation
- [ ] Token validation middleware
- [ ] Claims-based authorization
- [ ] Role-based access control (RBAC)
- **Arquivos:** `internal/infrastructure/auth/jwt.go`

#### 7.1.3 Security Features (2 dias)

- [ ] Token storage (encrypted)
- [ ] Token rotation
- [ ] Audit logging
- [ ] Rate limiting per user
- **Arquivos:** `internal/infrastructure/auth/security.go`

### 7.2 Entregáveis

- [ ] `internal/infrastructure/auth/` - Auth system completo
- [ ] Multi-provider support (4+)
- [ ] Documentation: Auth setup guide
- [ ] Migration path (no auth → auth)

### 7.3 Dependências Necessárias

```go
require (
    golang.org/x/oauth2 v0.15.0                         // OAuth2
    github.com/go-chi/jwtauth/v5 v5.3.0                // JWT
)
```

### 7.4 Métricas de Sucesso

- [ ] 4+ OAuth providers working
- [ ] <100ms token validation
- [ ] Zero security vulnerabilities (OWASP scan)
- [ ] Enterprise-ready docs

---

## 8. Sprint 10 (Semanas 19-20): Hybrid Backend

**Duração:** 15 dias úteis  
**Prioridade:** P1 - IMPORTANTE  
**Objetivo:** Local SQLite (fast) + Cloud sync (backup)

### 8.1 Features a Desenvolver

#### 8.1.1 Cloudflare Integration (10 dias)

**Cloudflare Services:**
- [ ] D1 Database (SQL)
- [ ] Vectorize (vector storage)
- [ ] R2 (object storage para backups)

**Sync Logic:**
- [ ] Local-first architecture (5ms reads)
- [ ] Background sync (writes)
- [ ] Conflict resolution (last-write-wins)
- [ ] Offline-capable
- **Arquivos:** `internal/infrastructure/hybrid/cloudflare.go`

#### 8.1.2 Sync Engine (5 dias)

- [ ] Bidirectional sync
- [ ] Delta sync (apenas mudanças)
- [ ] Sync status tracking
- [ ] Error handling e retry
- [ ] Manual sync trigger
- **Arquivos:** `internal/sync/engine.go`

### 8.2 Entregáveis

- [ ] `internal/infrastructure/hybrid/` - Hybrid backend
- [ ] `internal/sync/` - Sync engine
- [ ] MCP tools: `sync_now`, `get_sync_status`
- [ ] Migration from local-only

### 8.3 Dependências Necessárias

```go
require (
    github.com/cloudflare/cloudflare-go v0.82.0        // Cloudflare API
)
```

### 8.4 Métricas de Sucesso

- [ ] <10ms local reads
- [ ] Background sync <5min latency
- [ ] 99.9% sync success rate
- [ ] Zero data loss

---

## 9. Sprint 11 (Semanas 21-22): Temporal Features COMPLETE

**Duração:** 12 dias úteis  
**Prioridade:** P1 - IMPORTANTE  
**Objetivo:** Ciclo completo - Criação → Versionamento → Decay → Análise histórica

### 9.1 Features a Desenvolver

#### 9.1.1 Background Task System (5 dias)

**Task Queue:**
- [ ] Goroutine pool (configurable size)
- [ ] Job queue (priority-based)
- [ ] Task scheduling (cron-like)
- [ ] Error handling e retry
- **Arquivos:** `internal/infrastructure/taskqueue/pool.go`

#### 9.1.2 Temporal Features (7 dias)

**1. Criação** (já implementado)
- ✅ Timestamps automáticos em todos elementos
- [ ] Melhorar precisão (nanoseconds)

**2. Versionamento** (3 dias)
- [ ] Version history tracking para cada elemento
- [ ] Snapshot storage (diffs, não full copies)
- [ ] MCP tool: `get_element_history(id, limit)`
- **Arquivos:** `internal/domain/version_history.go`

**3. Confidence Decay** (2 dias)
- [ ] Half-life configurável (default: 30 dias)
- [ ] Exponential decay function
- [ ] Minimum confidence floors (não decai abaixo de X)
- [ ] Reinforcement learning: relações ganham confidence quando reforçadas
- [ ] MCP tool: `get_decayed_graph(reference_time)`
- **Arquivos:** `internal/domain/confidence_decay.go`

**4. Análise Histórica - Time Travel** (2 dias)
- [ ] `get_graph_at_time(timestamp)` - Estado do grafo em momento específico
- [ ] `get_relation_history(id)` - Histórico de relacionamento
- [ ] Reference time flexibility
- **Arquivos:** `internal/application/temporal.go`

### 9.2 Novos MCP Tools

- [ ] `get_element_history` - Version history de elemento
- [ ] `get_relation_history` - Histórico de relacionamento
- [ ] `get_graph_at_time` - Time-travel query
- [ ] `get_decayed_graph` - Graph com confidence decay aplicado

### 9.3 Entregáveis

- [ ] `internal/infrastructure/taskqueue/` - Task system
- [ ] `internal/application/temporal.go` - Temporal queries
- [ ] `internal/domain/version_history.go` - Versioning
- [ ] `internal/domain/confidence_decay.go` - Decay logic
- [ ] 4+ new MCP tools

### 9.4 Dependências Necessárias

```go
require (
    github.com/panjf2000/ants/v2 v2.9.0                // Goroutine pool
    github.com/RichardKnop/machinery/v2 v2.0.13        // Task queue (opcional)
)
```

### 9.5 Métricas de Sucesso

- [ ] Version history <10% storage overhead
- [ ] Time-travel queries <100ms
- [ ] Decay calculations <50ms
- [ ] Background tasks sem impacto em foreground

---

## 10. Sprint 12 (Semanas 23-24): UX & Installation

**Duração:** 8 dias úteis  
**Prioridade:** P1 - IMPORTANTE  
**Objetivo:** Melhorar onboarding e integrações

### 10.1 Features a Desenvolver

#### 10.1.1 One-Click Installer (3 dias)

**NPX-Based Setup:**
- [ ] `npx @fsvxavier/nexs-mcp-server init` command
- [ ] Auto-detect environment (Claude Desktop, VS Code, etc.)
- [ ] Generate config files automaticamente
- [ ] Download binaries se necessário
- [ ] Setup wizard interativo
- **Arquivos:** `scripts/install.js`

#### 10.1.2 Obsidian Export (3 dias)

**Export Formats:**
- [ ] Markdown (basic)
- [ ] Dataview format (with frontmatter)
- [ ] Canvas format (mindmaps)
- [ ] Auto-export option (após create)
- [ ] Batch export command

**MCP Tools:**
- [ ] `export_to_obsidian` - Export single element
- [ ] `batch_export_to_obsidian` - Export multiple
- **Arquivos:** `internal/export/obsidian.go`

#### 10.1.3 CLI Improvements (2 dias)

- [ ] Better help messages
- [ ] Interactive prompts
- [ ] Progress bars para long operations
- [ ] Colored output
- [ ] Auto-completion scripts (bash/zsh)

### 10.2 Entregáveis

- [ ] `scripts/install.js` - One-click installer
- [ ] `internal/export/obsidian.go` - Obsidian integration
- [ ] Enhanced CLI
- [ ] User onboarding guide

### 10.3 Dependências Necessárias

```go
require (
    github.com/yuin/goldmark v1.6.0                     // Markdown export
)
```

### 10.4 Métricas de Sucesso

- [ ] <2min setup time (fresh install)
- [ ] Obsidian export compatibility >95%
- [ ] User satisfaction >4.5/5 (surveys)

---

## 11. Features P2 - Roadmap Futuro (Q2 2026)

**Timeline:** Abril-Junho 2026 (Sprints 13-17)  
**Prioridade:** P2 - Nice-to-have

### 11.1 Sprint 13-14: Web Dashboard (20 dias)

**Objetivo:** Interface web React para visualização e gestão

**Features:**
- [ ] React 18 + TypeScript frontend
- [ ] Real-time statistics dashboard (SSE)
- [ ] Memory distribution charts (Recharts)
- [ ] Graph visualization (React Flow)
- [ ] Element browser com filtros avançados
- [ ] Search interface com preview
- [ ] Quality score analytics
- [ ] Responsive design (mobile-friendly)

**Arquivos:**
- `web/dashboard/` - Frontend React app
- `internal/infrastructure/httpserver/` - HTTP/SSE server
- `internal/application/dashboard_stats.go` - Statistics API

**Métricas:**
- [ ] <2s load time
- [ ] Support 100k+ elements
- [ ] WCAG 2.1 AA accessibility

### 11.2 Sprint 15: Memory Consolidation (15 dias)

**Objetivo:** Dream-inspired memory consolidation automática

**Features:**
- [ ] Decay scoring (time-based importance)
- [ ] Association discovery automática
- [ ] Semantic clustering (K-means)
- [ ] Memory compression (merge duplicates)
- [ ] Scheduled consolidation (nightly, 24/7)
- [ ] Archival de low-quality memories

**Arquivos:**
- `internal/application/consolidation.go`
- `internal/infrastructure/scheduler/`

**Métricas:**
- [ ] 30-50% memory reduction após consolidation
- [ ] <5min processing (10k memories)
- [ ] Zero data loss

### 11.3 Sprint 16: Graph Database + Export (15 dias)

**Graph Database Native (10 dias):**
- [ ] SQLite recursive CTEs para graph traversal
- [ ] Shortest path queries (A*, Dijkstra)
- [ ] Connected components detection
- [ ] Relationship strength scoring
- [ ] MCP tools: `find_path`, `get_connected`

**Advanced Export Formats (5 dias):**
- [ ] JSON Schema
- [ ] CSV/Excel (tabular)
- [ ] Graphviz DOT (graph viz)
- [ ] Neo4j Cypher (import)
- [ ] OPML (outliner)

**Métricas:**
- [ ] <50ms queries (10k nodes)
- [ ] Path finding accuracy >99%

### 11.4 Sprint 17: Advanced Analytics + Plugins (12 dias)

**Advanced Analytics (7 dias):**
- [ ] Usage statistics (most accessed)
- [ ] Relationship analytics (centrality, clustering)
- [ ] Quality trends over time
- [ ] Language/type distribution
- [ ] Topic modeling (BERTopic opcional)
- [ ] MCP tool: `get_analytics`

**Plugin System (5 dias):**
- [ ] Plugin interface definition
- [ ] Plugin loader (Go plugins ou gRPC)
- [ ] Plugin lifecycle management
- [ ] Custom element types via plugins
- [ ] Custom MCP tools via plugins

**Métricas:**
- [ ] 15+ analytics metrics
- [ ] Plugin hot-reload <1s

---

---

## Priority Matrix

### 🔴 Critical (Sprints 5-8) - P0
1. ❌ **Vector Embeddings Foundation** - 4 providers + semantic search
2. ❌ **HNSW Performance** - Sub-50ms queries, approximate NN
3. ❌ **Two-Tier Memory** - Working memory + Long-term separation
4. ❌ **Memory Quality (ONNX)** - Local SLM scoring + Multi-tier fallback

### 🟡 High Priority (Sprints 9-10) - P1
5. ❌ **Enterprise Auth** - OAuth2/JWT (Auth0, Cognito, Okta, Azure AD)
6. ❌ **Hybrid Backend** - Cloudflare D1/Vectorize/R2 sync
7. ❌ **Temporal Features** - Version history, confidence decay, time-travel
8. ❌ **UX & Installation** - One-click installer, Obsidian export

### 🟢 Medium Priority (Sprints 13-15) - P2
9. ❌ **Web Dashboard** - React UI com real-time statistics
10. ❌ **Memory Consolidation** - Dream-inspired algorithms
11. ❌ **Graph Database Native** - CTEs, path finding, advanced queries
12. ❌ **Advanced Export** - JSON, CSV, Graphviz, Neo4j, OPML

### 🔵 Low Priority (Sprints 16-17) - P2
13. ❌ **Advanced Analytics** - Usage stats, topic modeling, centrality
14. ❌ **Plugin System** - Hot-reload, custom elements/tools
15. **Enhanced CLI** - Auto-completion, progress bars, colored output
16. **Mobile Support** - Progressive Web App

---

## Success Metrics

### Technical Metrics (v2.0.0 Targets)
- [ ] Test Coverage: 85%+ (atual: ~75%)
- [ ] Zero critical security issues (OWASP scan)
- [ ] Vector search <100ms (10k vectors)
- [ ] HNSW queries <50ms (10k vectors)
- [ ] Working memory access <1ms
- [ ] Quality scoring <100ms (ONNX CPU)
- [ ] Support 100k+ elements
- [ ] Support 1M+ relationships
- [ ] 99.9% uptime

### Feature Parity Metrics
- ✅ GitHub Integration: 100% (COMPLETO v1.0.x)
- ✅ Collection System: 100% (COMPLETO v1.0.x)
- ✅ Ensembles: 100% (COMPLETO v1.0.x)
- ✅ Context Enrichment: 100% (COMPLETO v1.0.x)
- ❌ Vector Embeddings: 0%
- ❌ HNSW Index: 0%
- ❌ Two-Tier Memory: 0%
- ❌ Memory Quality: 0%
- ❌ Enterprise Auth: 0%

### Distribution Metrics
- ✅ Go install available (v1.0.5)
- ✅ Docker Hub published (v0.1.0, 14.5 MB)
- ✅ NPM published (@fsvxavier/nexs-mcp-server@1.0.5)
- [ ] Homebrew installs: 50+
- [ ] GitHub stars: 500+
- [ ] Docker pulls: 1000+

### Documentation Metrics
- ✅ User guide complete (2,000+ lines)
- ✅ API reference complete
- ✅ Developer documentation (15+ files)
- ✅ Architecture docs (5 files)
- ✅ 10+ ADRs
- [ ] Tutorial videos (3+)

### Community Metrics
- [ ] GitHub Discussions active
- [ ] 10+ external contributors
- [ ] 50+ collection submissions
- [ ] Active Slack/Discord
- [ ] Monthly releases

---

## Timeline v2.0.0

### Q1 2026 (Janeiro - Março)
- **Sprints 5-8 (8 semanas):** P0 Features críticas
  - Vector Embeddings (2 semanas)
  - HNSW Performance (2 semanas)
  - Two-Tier Memory (2 semanas)
  - Memory Quality (2 semanas)

### Q2 2026 (Abril - Junho)
- **Sprints 9-12 (8 semanas):** P1 Features importantes
  - Enterprise Auth (3 semanas)
  - Hybrid Backend (3 semanas)
  - Temporal Complete (2 semanas)
  - UX & Installation (1 semana)
- **Sprints 13-17 (8 semanas):** P2 Features diferenciação
  - Web Dashboard (4 semanas)
  - Memory Consolidation (3 semanas)
  - Graph Database (3 semanas)
  - Analytics + Plugins (2 semanas)

### Milestones
- **v2.0.0-alpha (Fim Sprint 8):** Core enterprise features
- **v2.0.0-beta (Fim Sprint 12):** Production-ready
- **v2.0.0-rc (Fim Sprint 15):** Release candidate
- **v2.0.0 GA (Junho 2026):** General availability

---

## Riscos e Mitigações

### Risco 1: Performance Degradation
**Probabilidade:** Média | **Impacto:** Alto  
**Mitigação:**
- Extensive benchmarking em cada sprint
- Performance budgets definidos (Vector <100ms, HNSW <50ms)
- Profiling contínuo com pprof
- Fallback para approaches mais leves

### Risco 2: Breaking Changes
**Probabilidade:** Média | **Impacto:** Alto  
**Mitigação:**
- API versioning desde início (v2 namespace)
- Migration guides para cada sprint
- Backward compatibility tests automáticos
- Deprecation warnings (2 releases antes de remoção)

### Risco 3: Dependency Hell
**Probabilidade:** Baixa | **Impacto:** Médio  
**Mitigação:**
- Dependências mínimas necessárias (15 novas libs)
- Vendor quando crítico (ONNX models)
- Abstractions para trocar libs facilmente
- Regular dependency audits (Dependabot)

### Risco 4: Scope Creep
**Probabilidade:** Alta | **Impacto:** Médio  
**Mitigação:**
- P0/P1/P2 priorization rígida
- Sprint goals bem definidos (3-4 features max)
- Weekly checkpoints com review
- Defer para P2 quando necessário
- Feature freeze antes de cada release

### Risco 5: ONNX Compatibility Issues
**Probabilidade:** Média | **Impacto:** Médio  
**Mitigação:**
- Multi-tier fallback (ONNX → Groq → Gemini → Implicit)
- Extensive testing em múltiplas plataformas
- Documentação clara de requirements
- Community feedback early (alpha releases)

---

## Próximos Passos Imediatos

### Esta Semana (23-27 Dezembro 2025)
1. [ ] Review e aprovação deste roadmap v2.0.0
2. [ ] Setup environment para Sprint 5
3. [ ] Research aprofundado em embedding providers
4. [ ] Criar issues no GitHub para cada feature Sprint 5
5. [ ] Definir métricas de success detalhadas

### Próxima Semana (30 Dez - 3 Jan 2026)
1. [ ] Iniciar Sprint 5 (Vector Embeddings)
2. [ ] Implementar OpenAI provider
3. [ ] Implementar Local Transformers provider (default)
4. [ ] Setup CI/CD para novos tests
5. [ ] Documentar decisões arquiteturais (ADRs)

### Janeiro 2026 (Semanas 1-4)
1. [ ] Completar Sprint 5 (Vector Embeddings)
2. [ ] Iniciar Sprint 6 (HNSW Performance)
3. [ ] Publicar v2.0.0-alpha1 com vector search
4. [ ] Community feedback round 1

---

## 16. Checklist Completo de Desenvolvimento

### Sprint 5: Vector Embeddings ✅ = 0/12
- [ ] OpenAI provider
- [ ] Local Transformers provider (default)
- [ ] Sentence Transformers provider
- [ ] ONNX provider
- [ ] Provider factory + fallback
- [ ] Semantic search API
- [ ] Vector store abstraction
- [ ] Embedding cache
- [ ] 2+ MCP tools
- [ ] Unit tests
- [ ] Integration tests
- [ ] Documentation

### Sprint 6: HNSW Index ✅ = 0/8
- [ ] HNSW graph construction
- [ ] Approximate NN search
- [ ] Index persistence
- [ ] Hybrid search (HNSW + filters)
- [ ] Benchmark suite
- [ ] Integration tests
- [ ] Parameter tuning guide
- [ ] Performance report

### Sprint 7: Two-Tier Memory ✅ = 0/10
- [ ] Working memory model
- [ ] Long-term memory refactor
- [ ] TTL + expiration
- [ ] Promotion rules
- [ ] Manual promotion tool
- [ ] 15+ MCP tools
- [ ] Migration guide
- [ ] Unit tests
- [ ] Integration tests
- [ ] Documentation

### Sprint 8: Memory Quality ✅ = 0/9
- [ ] ONNX integration
- [ ] Quality scoring
- [ ] Multi-tier fallback
- [ ] Implicit signals
- [ ] Retention policies
- [ ] Archival system
- [ ] Background cleanup
- [ ] MCP tool
- [ ] Tests

### Sprint 9: Enterprise Auth ✅ = 0/10
- [ ] OAuth2 (Auth0)
- [ ] OAuth2 (AWS Cognito)
- [ ] OAuth2 (Okta)
- [ ] OAuth2 (Azure AD)
- [ ] JWT generation
- [ ] JWT validation
- [ ] RBAC
- [ ] Token storage
- [ ] Audit logging
- [ ] Documentation

### Sprint 10: Hybrid Backend ✅ = 0/8
- [ ] Cloudflare D1 integration
- [ ] Cloudflare Vectorize
- [ ] Cloudflare R2
- [ ] Sync engine
- [ ] Conflict resolution
- [ ] Delta sync
- [ ] MCP tools
- [ ] Tests

### Sprint 11: Temporal Complete ✅ = 0/9
- [ ] Task queue system
- [ ] Version history
- [ ] Snapshot storage
- [ ] Confidence decay
- [ ] Reinforcement learning
- [ ] Time-travel queries
- [ ] 4+ MCP tools
- [ ] Tests
- [ ] Documentation

### Sprint 12: UX & Installation ✅ = 0/7
- [ ] One-click installer
- [ ] Setup wizard
- [ ] Obsidian Markdown export
- [ ] Obsidian Dataview export
- [ ] Obsidian Canvas export
- [ ] CLI improvements
- [ ] User guide

### Sprints 13-17: P2 Features ✅ = 0/20
- [ ] Web Dashboard (React)
- [ ] Real-time statistics
- [ ] Graph visualization
- [ ] Memory consolidation
- [ ] Dream-inspired algorithms
- [ ] Semantic clustering
- [ ] Graph database CTEs
- [ ] Path finding
- [ ] Advanced export (5 formats)
- [ ] Advanced analytics
- [ ] Topic modeling
- [ ] Plugin system
- [ ] Plugin loader
- [ ] Hot-reload
- [ ] Auto-completion
- [ ] Progress bars
- [ ] Colored output
- [ ] Accessibility (WCAG)
- [ ] Mobile responsive
- [ ] Documentation completa

---

---

## 17. Dependências Consolidadas

### Sprint 5-8 Dependencies (P0)

```go
// go.mod additions
require (
    // Sprint 5: Vector Embeddings
    github.com/sashabaranov/go-openai v1.17.9          // OpenAI embeddings
    github.com/nlpodyssey/spago v1.1.0                 // Local Transformers
    github.com/james-bowman/nlp v0.0.0                 // Sentence Transformers
    github.com/yalue/onnxruntime_go v1.8.0             // ONNX Runtime
    
    // Sprint 6: HNSW
    github.com/Bithack/go-hnsw v0.0.0-20211102081019   // HNSW index
    
    // Sprint 8: Memory Quality (ONNX já incluído acima)
)
```

### Sprint 9-12 Dependencies (P1)

```go
require (
    // Sprint 9: Auth
    golang.org/x/oauth2 v0.15.0                         // OAuth2
    github.com/go-chi/jwtauth/v5 v5.3.0                // JWT
    
    // Sprint 10: Hybrid Backend
    github.com/cloudflare/cloudflare-go v0.82.0        // Cloudflare API
    
    // Sprint 11: Temporal
    github.com/panjf2000/ants/v2 v2.9.0                // Goroutine pool
    github.com/RichardKnop/machinery/v2 v2.0.13        // Task queue (opcional)
    
    // Sprint 12: Export
    github.com/yuin/goldmark v1.6.0                     // Markdown
)
```

### Sprint 13-17 Dependencies (P2)

```go
require (
    // Web Dashboard
    github.com/go-echarts/go-echarts/v2 v2.3.3         // Charts (opcional)
    
    // Export Formats
    github.com/jung-kurt/gofpdf v1.16.2                // PDF
    github.com/tealeg/xlsx v1.0.5                      // Excel
    github.com/emicklei/dot v1.6.0                     // Graphviz
    
    // Plugin System
    github.com/hashicorp/go-plugin v1.6.0              // Plugins
)
```

### Dependências Existentes (v1.0.x)

```go
// Já instaladas
require (
    github.com/modelcontextprotocol/go-sdk v1.1.0     // MCP SDK
    github.com/google/go-github/v57 v57.0.0           // GitHub API
    golang.org/x/oauth2 v0.15.0                        // OAuth2 (GitHub)
    modernc.org/sqlite v1.28.0                         // SQLite
    github.com/spf13/cobra v1.8.0                     // CLI
    gopkg.in/yaml.v3 v3.0.1                           // YAML parsing
    github.com/stretchr/testify v1.8.4                // Testing
)
```

---

## 18. Métricas de Sucesso Globais v2.0.0

### Performance Targets
- [ ] Vector search <100ms (10k vectors)
- [ ] HNSW queries <50ms (10k vectors)
- [ ] Working memory access <1ms
- [ ] Long-term memory access <10ms
- [ ] Quality scoring <100ms (ONNX CPU)
- [ ] Time-travel queries <100ms
- [ ] Graph queries <50ms (10k nodes)

### Quality Targets
- [ ] Test coverage >80% (all new code)
- [ ] Zero security vulnerabilities
- [ ] API backward compatibility 100%
- [ ] Documentation coverage 100%
- [ ] User satisfaction >4.5/5

### Scale Targets
- [ ] Support 100k+ elements
- [ ] Support 1M+ relationships
- [ ] 99.9% uptime
- [ ] <1% error rate
- [ ] Memory usage <500MB (100k elements)

---

**Última Atualização:** 22 de dezembro de 2025  
**Próxima Revisão:** 27 de dezembro de 2025  
**Status:** 📋 PLANEJAMENTO - Aguardando aprovação para início Sprint 5
