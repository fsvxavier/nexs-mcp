# NEXS MCP Server

<div align="center">

[![CI](https://github.com/fsvxavier/nexs-mcp/workflows/CI/badge.svg)](https://github.com/fsvxavier/nexs-mcp/actions)
[![Coverage](https://img.shields.io/badge/coverage-76.4%25-green)](./COVERAGE_REPORT.md)
[![Go Version](https://img.shields.io/badge/go-1.25-blue)](https://go.dev)
[![Release](https://img.shields.io/badge/release-v1.4.0-blue)](https://github.com/fsvxavier/nexs-mcp/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)
[![MCP SDK](https://img.shields.io/badge/MCP_SDK-v1.2.0-blue)](https://github.com/modelcontextprotocol/go-sdk)
[![Tools](https://img.shields.io/badge/MCP_Tools-121-brightgreen)](#-available-tools)
[![NPM Package](https://img.shields.io/npm/v/@fsvxavier/nexs-mcp-server?label=npm)](https://www.npmjs.com/package/@fsvxavier/nexs-mcp-server)
[![Docker Hub](https://img.shields.io/docker/pulls/fsvxavier/nexs-mcp?label=docker%20pulls)](https://hub.docker.com/r/fsvxavier/nexs-mcp)

**A production-ready Model Context Protocol (MCP) server built in Go**

*Manage AI elements (Personas, Skills, Templates, Agents, Memories, and Ensembles) with enterprise-grade architecture, high performance, comprehensive tooling, and **intelligent token optimization** that reduces AI context usage by 81-95% through 8 advanced optimization services including compression, streaming, deduplication, summarization, and adaptive caching.*

[📚 Documentation](#-documentation) • [🚀 Quick Start](#-quick-start) • [🔧 Tools](#-available-tools) • [📦 Element Types](#-element-types) • [💡 Examples](#-usage-examples) • [⚡ Token Optimization](#-token-optimization-system)

</div>

---

## 🎯 What is NEXS MCP?

NEXS MCP Server is a high-performance implementation of the [Model Context Protocol](https://modelcontextprotocol.io/), designed to manage AI elements with enterprise-grade architecture. Built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) v1.1.0, it provides a robust foundation for AI system management.

### Why NEXS MCP?

- **� Token Economy** - Reduces AI context usage by 70-85% through intelligent conversation memory and keyword extraction
- **🌍 Multilingual Support** - 11 languages supported (EN, PT, ES, FR, DE, IT, RU, JA, ZH, AR, HI) with automatic detection
- **�🚀 High Performance** - Built in Go for speed and efficiency
- **🏗️ Clean Architecture** - Domain-driven design with clear separation of concerns
- **✅ Production Ready** - 70% test coverage with 500+ tests, zero race conditions, zero linter issues
- **🔧 96 MCP Tools** - Complete portfolio (74 base + 8 optimization + 15 working memory + 3 quality scoring + others)
- **📦 6 Element Types** - Personas, Skills, Templates, Agents, Memories, Ensembles
- **🔄 Dual Storage** - File-based (YAML) or in-memory storage modes
- **🌐 Cross-Platform** - Binaries for Linux, macOS, Windows (amd64/arm64)
- **🐳 Docker Ready** - Multi-arch Docker images with security hardening
- **📊 Analytics** - Built-in performance monitoring and usage statistics

### Use Cases

- **Token Optimization** - Reduce AI API costs by 81-95% with 8 intelligent optimization services: response compression (gzip/zlib), streaming responses, semantic deduplication, TF-IDF summarization, context window management, adaptive caching, batch processing, and prompt compression
- **Memory Consolidation** - HNSW-based duplicate detection, DBSCAN/K-means clustering, knowledge graph extraction with NLP entities & relationships (Sprint 14)
- **Quality Scoring** - Built-in ONNX models for content quality assessment (MS MARCO for speed, Paraphrase-Multilingual for quality)
- **AI System Management** - Centralized management of AI personas, skills, and workflows
- **Portfolio Organization** - Organize and version control AI elements with GitHub integration
- **Team Collaboration** - Share collections of elements across teams via GitHub
- **Development Workflows** - Automate AI element creation and deployment
- **Context Management** - Store and retrieve conversation memories with deduplication and automatic language detection
- **Multi-Agent Systems** - Orchestrate ensembles of agents with sophisticated execution strategies
- **Multilingual Applications** - Support conversations in 11 languages with automatic detection and optimized stop word filtering

---

## ✨ Key Features

### NLP & Analytics Features (Sprint 18)
- ✅ **Enhanced Entity Extraction** - ONNX BERT-based NER with 9 entity types (PERSON, ORGANIZATION, LOCATION, DATE, EVENT, PRODUCT, TECHNOLOGY, CONCEPT, OTHER)
  - Transformer models: protectai/bert-base-NER-onnx (411 MB)
  - CoNLL-2003 BIO format labels (B-PER, I-PER, B-ORG, I-ORG, B-LOC, I-LOC, B-MISC, I-MISC, O)
  - Performance: 100-200ms (CPU), 15-30ms (GPU), 93%+ accuracy
  - Confidence scoring: 0.0-1.0 with configurable threshold (default: 0.7)
  - Fallback: Rule-based regex extraction (confidence=0.5)
  - [API Documentation](docs/api/MCP_TOOLS.md#nlp-tools) | [NLP Features Guide](docs/NLP_FEATURES.md)
- ✅ **Relationship Detection** - 10 relationship types with evidence tracking
  - Types: WORKS_AT, FOUNDED, LOCATED_IN, BORN_IN, LIVES_IN, HEADQUARTERED_IN, DEVELOPED_BY, USED_BY, AFFILIATED_WITH, RELATED_TO
  - Co-occurrence-based relationship inference
  - Bidirectional relationship storage
  - Confidence scores and evidence text
- ✅ **Sentiment Analysis** - DistilBERT-based multilingual sentiment with emotional dimensions
  - Model: lxyuan/distilbert-base-multilingual-cased-sentiments-student (516 MB)
  - Labels: POSITIVE, NEGATIVE, NEUTRAL, MIXED (threshold: 0.6)
  - Emotional dimensions: joy, sadness, anger, fear, surprise, disgust (0.0-1.0 scores)
  - Performance: 50-100ms (CPU), 10-20ms (GPU), 91%+ accuracy
  - Trend analysis: 5-point moving average for sentiment tracking
  - Shift detection: Configurable threshold for emotional changes
  - Fallback: Lexicon-based sentiment (positive/negative word lists)
- ✅ **Topic Modeling** - Classical algorithms with coherence scoring
  - Algorithms: LDA (Latent Dirichlet Allocation), NMF (Non-negative Matrix Factorization)
  - LDA: Gibbs sampling, configurable iterations/topics
  - NMF: Multiplicative updates, faster than LDA (0.5-2s vs 1-5s for 100 docs)
  - Quality metrics: Coherence (keyword co-occurrence), Diversity (keyword uniqueness)
  - Performance: 1-5s for 100 documents (LDA, CPU)
  - Pure Go implementation: No ONNX dependency
- ✅ **6 NLP MCP Tools** - Advanced NLP capabilities accessible via MCP protocol
  - `extract_entities_advanced` - Entity extraction with transformer models
  - `analyze_sentiment` - Sentiment analysis with emotional tone
  - `extract_topics` - Topic modeling with LDA/NMF
  - `analyze_sentiment_trend` - Sentiment trend analysis with moving averages
  - `detect_emotional_shifts` - Emotional change detection
  - `summarize_sentiment` - Aggregate sentiment statistics
- ✅ **ONNXBERTProvider** - Unified ONNX provider for BERT/DistilBERT models
  - Thread-safe with sync.RWMutex protection
  - Dual model support: BERT NER (3 inputs), DistilBERT Sentiment (2 inputs)
  - BIO format tokenization with space-based fallback
  - Batch processing: Configurable batch size (default: 16)
  - GPU acceleration: CUDA/ROCm support via NEXS_NLP_USE_GPU=true
  - Build tags: Portable builds without ONNX (noonnx tag)
  - [Model Download Guide](docs/DOWNLOAD_NLP_MODELS.md) | [Configuration Reference](docs/user-guide/ONNX_MODEL_CONFIGURATION.md)

### Core Infrastructure
- ✅ **Official MCP SDK** - Built on github.com/modelcontextprotocol/go-sdk v1.2.0
- ✅ **Clean Architecture** - Domain-driven design with clear separation of concerns
- ✅ **High Test Coverage** - 76.4% application, 91.7% HNSW, 96.7% TF-IDF with 295 tests, zero race conditions, zero linter issues
- ✅ **Dual Storage Modes** - File-based YAML or in-memory
- ✅ **121 MCP Tools** - Complete portfolio with NLP, consolidation, optimization, temporal features, and task scheduling
- ✅ **8 Token Optimization Services** - Compression, streaming, deduplication, summarization, context management, adaptive caching, batch processing, prompt compression
- ✅ **6 Element Types** - Persona, Skill, Template, Agent, Memory, Ensemble
- ✅ **24 Application Services** - Including 3 NLP services + 1 ONNX provider (Sprint 18)
- ✅ **Stdio Transport** - Standard MCP communication over stdin/stdout
- ✅ **Thread-Safe** - Concurrent operations with proper synchronization
- ✅ **Cross-Platform** - Binaries for Linux, macOS, Windows (amd64/arm64)

### Production Features
- ✅ **Official MCP SDK** - Built on github.com/modelcontextprotocol/go-sdk v1.2.0
- ✅ **Clean Architecture** - Domain-driven design with clear separation of concerns
- ✅ **High Test Coverage** - 76.4% application, 91.7% HNSW, 96.7% TF-IDF with 295 tests, zero race conditions, zero linter issues
- ✅ **Dual Storage Modes** - File-based YAML or in-memory
- ✅ **104 MCP Tools** - Complete portfolio with consolidation, optimization, temporal features, and task scheduling
- ✅ **8 Token Optimization Services** - Compression, streaming, deduplication, summarization, context management, adaptive caching, batch processing, prompt compression
- ✅ **6 Element Types** - Persona, Skill, Template, Agent, Memory, Ensemble
- ✅ **21 Application Services** - Including 4 new consolidation services (Sprint 14)
- ✅ **Stdio Transport** - Standard MCP communication over stdin/stdout
- ✅ **Thread-Safe** - Concurrent operations with proper synchronization
- ✅ **Cross-Platform** - Binaries for Linux, macOS, Windows (amd64/arm64)

### GitHub Integration
- ✅ **OAuth Authentication** - Secure device flow authentication
- ✅ **Portfolio Sync** - Push/pull elements to/from GitHub repositories
- ✅ **Collection System** - Install, manage, and publish element collections
- ✅ **PR Submission** - Submit elements to collections via automated PRs
- ✅ **Conflict Detection** - Smart conflict resolution with multiple strategies
- ✅ **Incremental Sync** - Efficient delta-based synchronization

### Production Features
- ✅ **Auto-Save** - Automatic conversation context preservation with multilingual keyword extraction (11 languages)
- ✅ **Token Optimization** - 81-95% reduction in AI context usage through 8 optimization services: compression (gzip/zlib 70-75%), streaming (prevent overflow), semantic deduplication (92%+ similarity), TF-IDF summarization (70% reduction), context window management (smart truncation), adaptive cache (dynamic TTL 1h-7d), batch processing (10x faster), and prompt compression (35% reduction)
- ✅ **ONNX Quality Scoring** - Built-in models for content quality assessment
  - **MS MARCO MiniLM-L-6-v2** (default): 61.64ms latency, 9 languages (non-CJK), ~16 inf/s throughput
  - **Paraphrase-Multilingual-MiniLM-L12-v2** (configurable): 109.41ms latency, 11 languages including CJK, 71% more effective
  - Multi-tier fallback: ONNX → Groq API → Gemini API → Implicit Signals
  - Quality-based retention policies (High: 365d, Medium: 180d, Low: 90d)
  - [Configuration Guide](docs/user-guide/ONNX_MODEL_CONFIGURATION.md) | [Benchmarks](BENCHMARK_RESULTS.md)
- ✅ **Working Memory System** - Session-scoped memory with priority-based TTL (15 tools)
  - Priority levels: Low (1h), Medium (4h), High (12h), Critical (24h)
  - Auto-promotion to long-term storage based on access patterns
  - Background cleanup every 5 minutes
  - [API Documentation](docs/api/WORKING_MEMORY_TOOLS.md)
- ✅ **Background Task Scheduler** - Robust scheduling system (Sprint 11)
  - Cron-like expressions: wildcards, ranges, steps, lists
  - Priority-based execution: Low/Medium/High
  - Task dependencies with validation
  - Persistent storage with JSON and atomic writes
  - Auto-retry with configurable delays
  - [API Documentation](docs/api/TASK_SCHEDULER.md)
- ✅ **Temporal Features** - Time travel and version history (Sprint 11 - 4 tools)
  - Version history with snapshot/diff compression
  - Confidence decay: exponential, linear, logarithmic, step
  - Time travel queries: reconstruct graph at any point in time
  - Critical relationship preservation
  - [API Documentation](docs/api/TEMPORAL_FEATURES.md) | [User Guide](docs/user-guide/TIME_TRAVEL.md)
- ✅ **Multilingual Memory** - Automatic language detection (EN, PT, ES, FR, DE, IT, RU, JA, ZH, AR, HI) with language-specific stop word filtering
- ✅ **Quick Create Tools** - Simplified element creation with template defaults
- ✅ **Backup & Restore** - Portfolio backup with tar.gz compression and SHA-256 checksums
- ✅ **Memory Management** - Search, summarize, update memories with relevance scoring
- ✅ **Structured Logging** - slog-based JSON/text logs with context extraction
- ✅ **Log Query Tools** - Filter and search logs by level, user, operation, tool
- ✅ **User Identity** - Session management with metadata support
- ✅ **Analytics Dashboard** - Usage statistics and performance metrics (p50/p95/p99)

### Ensemble Capabilities
- ✅ **Sequential Execution** - Run agents in order with context sharing
- ✅ **Parallel Execution** - Run agents concurrently for speed
- ✅ **Hybrid Execution** - Mix sequential and parallel strategies
- ✅ **Aggregation Strategies** - First, last, consensus, voting, all, merge
- ✅ **Monitoring** - Real-time progress tracking and callbacks
- ✅ **Fallback Chains** - Automatic failover to backup agents

---

## 📊 Project Status

```
Version:               v1.4.0 (Enhanced NLP & Analytics)
Application Coverage:   76.4% ✓ (+13.2%)
HNSW Index:            91.7% ✓
TF-IDF Index:          96.7% ✓
Template Layer:        87.0% ✓
Portfolio Layer:       75.6% ✓
Lines of Code:         ~86,925 (42,739 production + 44,186 tests)
Test Cases:            295 tests (100% passing, 0 race conditions)
MCP Tools:             121 (26 element + 9 memory + 15 working + 10 consolidation + 6 NLP + others)
Application Services:  24 (3 new NLP services + 1 ONNX provider)
Element Types:         6 (Persona, Skill, Template, Agent, Memory, Ensemble)
ONNX Models:           4 (MS MARCO, Paraphrase-Multilingual, BERT NER, DistilBERT Sentiment)
Quality:               Zero race conditions, Zero linter issues
Token Optimization:    81-95% economy (8 services integrated)
NLP Performance:       100-200ms entity extraction, 50-100ms sentiment (CPU)
```

**Recent Milestones:**
- ✅ **Sprint 18 Complete** (04/01/2026) - v1.4.0: Enhanced NLP & Analytics (ONNX BERT/DistilBERT integration, 6 NLP tools, 4,849 LOC, 93%+ accuracy)
- ✅ **Sprint 14 Complete** (26/12/2025) - Advanced Application Services Test Coverage (295 tests, 76.4% coverage, 10 consolidation tools)
- ✅ **v1.3.0 Release** (24/12/2025) - Token Optimization (8 services: compression, streaming, deduplication, summarization, context, cache, batch, prompt compression)
- ✅ **v1.2.0 Release** (24/12/2025) - Task Scheduler + Temporal Features (Sprint 11 complete)
- ✅ **v1.1.0 Release** (23/12/2025) - ONNX Quality Scoring + Working Memory System + 91 MCP Tools
- ✅ **v1.0.1 Release** (20/12/2025) - Community infrastructure, benchmarks, template validator enhancements
- ✅ **v1.0.0 Release** (19/12/2025) - Production release with 66 MCP tools, GitHub integration, NPM distribution

---

## 🚀 Quick Start

### Installation

Choose your preferred installation method:

#### Option 1: NPM (Recommended - Cross-platform)

```bash
# Install globally
npm install -g @fsvxavier/nexs-mcp-server

# Verify installation
nexs-mcp --version
```

📦 **NPM Package:** https://www.npmjs.com/package/@fsvxavier/nexs-mcp-server

#### Option 2: Go Install (For Go developers)

```bash
go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@v1.4.0
```

#### Option 3: Homebrew (macOS/Linux)

```bash
# Add tap
brew tap fsvxavier/nexs-mcp

# Install
brew install nexs-mcp

# Verify installation
nexs-mcp --version
```

#### Option 4: Docker (Containerized)

```bash
# Pull image from Docker Hub
docker pull fsvxavier/nexs-mcp:latest

# Or pull specific version
docker pull fsvxavier/nexs-mcp:v1.4.0

# Run with volume mount
docker run -v $(pwd)/data:/app/data fsvxavier/nexs-mcp:latest

# Or use Docker Compose
docker-compose up -d
```

🐳 **Docker Hub:** https://hub.docker.com/r/fsvxavier/nexs-mcp
📦 **Image Size:** 14.5 MB (compressed), 53.7 MB (uncompressed)

#### Option 5: Build from Source

```bash
# Clone repository
git clone https://github.com/fsvxavier/nexs-mcp.git
cd nexs-mcp

# Install dependencies
go mod download

# Build
make build

# Run tests
make test-coverage

# Run server
./bin/nexs-mcp
```

### First Run

**File Storage (default):**
```bash
# Default configuration (file storage in data/elements)
nexs-mcp

# Custom data directory
nexs-mcp -data-dir /path/to/data

# Or via environment variable
NEXS_DATA_DIR=/path/to/data nexs-mcp
```

**In-Memory Storage:**
```bash
# Memory-only storage (data lost on restart)
nexs-mcp -storage memory

# Or via environment variable
NEXS_STORAGE_TYPE=memory nexs-mcp
```

**Output:**
```
NEXS MCP Server v1.0.0
Initializing Model Context Protocol server...
Storage type: file
Data directory: data/elements
Registered 66 tools
Server ready. Listening on stdio...
```

### Integration with Claude Desktop

Add to your Claude Desktop configuration:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`

**Linux:** `~/.config/Claude/claude_desktop_config.json`

**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "nexs-mcp",
      "args": [],
      "env": {
        "NEXS_DATA_DIR": "/path/to/your/elements",
        "NEXS_STORAGE_TYPE": "file"
      }
    }
  }
}
```

Restart Claude Desktop and you'll see NEXS MCP tools available!

For detailed setup instructions, see [docs/user-guide/GETTING_STARTED.md](docs/user-guide/GETTING_STARTED.md)

---

## 🔧 Available Tools

NEXS MCP provides **96 MCP tools** organized into categories:

### 🗂️ Element Management (11 tools)

**Generic CRUD Operations:**
1. **list_elements** - List all elements with advanced filtering (type, active_only, tags)
2. **get_element** - Get element details by ID
3. **create_element** - Create generic element
4. **update_element** - Update existing element
5. **delete_element** - Delete element by ID

**Type-Specific Creation:**
6. **create_persona** - Create Persona with behavioral traits
7. **create_skill** - Create Skill with triggers and procedures
8. **create_template** - Create Template with variable substitution
9. **create_agent** - Create Agent with goals and workflows
10. **create_memory** - Create Memory with content hashing
11. **create_ensemble** - Create Ensemble for multi-agent orchestration

### ⚡ Quick Create Tools (6 tools)

12. **quick_create_persona** - Simplified persona creation with minimal prompts
13. **quick_create_skill** - Simplified skill creation
14. **quick_create_template** - Simplified template creation
15. **quick_create_agent** - Simplified agent creation
16. **quick_create_memory** - Simplified memory creation
17. **quick_create_ensemble** - Simplified ensemble creation

### 📚 Collection System (10 tools)

18. **browse_collections** - Discover available collections (GitHub, local, HTTP)
19. **install_collection** - Install collection from URI (github://, file://, https://)
20. **uninstall_collection** - Remove installed collection
21. **list_installed_collections** - List all installed collections
22. **get_collection_info** - Get detailed collection information
23. **export_collection** - Export collection to tar.gz archive
24. **update_collection** - Update specific collection
25. **update_all_collections** - Update all installed collections
26. **check_collection_updates** - Check for available updates
27. **publish_collection** - Publish collection to GitHub

### 🐙 GitHub Integration (8 tools)

28. **github_auth_start** - Initiate OAuth2 device flow authentication
29. **github_auth_status** - Check GitHub authentication status
30. **github_list_repos** - List user's GitHub repositories
31. **github_sync_push** - Push local elements to GitHub repository
32. **github_sync_pull** - Pull elements from GitHub repository
33. **github_sync_bidirectional** - Two-way sync with conflict resolution
34. **submit_element_to_collection** - Submit element via automated PR
35. **track_pr_status** - Track PR submission status

### 💾 Backup & Restore (4 tools)

36. **backup_portfolio** - Create compressed backup with checksums
37. **restore_portfolio** - Restore from backup with validation
38. **activate_element** - Activate element (shortcut for update)
39. **deactivate_element** - Deactivate element (shortcut for update)

### 🧠 Memory Management (5 tools)

40. **search_memory** - Search memories with relevance scoring
41. **summarize_memories** - Get memory statistics and summaries
42. **update_memory** - Partial update of memory content
43. **delete_memory** - Delete specific memory
44. **clear_memories** - Bulk delete memories with filters

### 🎯 Memory Quality System (3 tools)

45. **score_memory_quality** - ONNX-based quality scoring with multi-tier fallback
46. **get_retention_policy** - Get retention policy for quality score
47. **get_retention_stats** - Memory retention statistics and quality distribution

### 📊 Analytics & Monitoring (11 tools)

48. **duplicate_element** - Duplicate element with new ID and optional name
49. **get_usage_stats** - Analytics with period filtering and top-10 rankings
50. **get_performance_dashboard** - Performance metrics with p50/p95/p99 latencies
51. **list_logs** - Query logs with filters (level, date, user, operation, tool)
52. **get_current_user** - Get current user session information
53. **set_user_context** - Set user identity with metadata
54. **clear_user_context** - Clear current user session
55. **get_context** - Get MCP server context information
56. **search_elements** - Advanced element search with filters
57. **execute_ensemble** - Execute ensemble with monitoring
58. **get_ensemble_status** - Get ensemble execution status

### 🔍 Context Enrichment System (3 tools)

59. **expand_memory_context** - Expand memory context by fetching related elements
60. **find_related_memories** - Find memories that reference a specific element (reverse search)
61. **suggest_related_elements** - Get intelligent recommendations based on relationships and patterns

### 🔗 Relationship System (5 tools)

62. **get_related_elements** - Bidirectional search with O(1) lookups (forward/reverse/both)
63. **expand_relationships** - Recursive expansion up to 5 levels with depth control
64. **infer_relationships** - Automatic inference (mention, keyword, semantic, pattern)
65. **get_recommendations** - Intelligent recommendations with 4 scoring strategies
66. **get_relationship_stats** - Index statistics (entries, cache hit rate)

### 🎨 Template System (4 tools)

67. **list_templates** - List available templates with filtering
68. **get_template** - Retrieve complete template details
69. **instantiate_template** - Instantiate template with variables (Handlebars)
70. **validate_template** - Validate template syntax and variables

### ✅ Validation & Rendering (2 tools)

71. **validate_element** - Type-specific validation (basic/comprehensive/strict)
72. **render_template** - Render template directly without creating element

### 🔄 Operations (2 tools)

73. **reload_elements** - Hot reload elements without server restart
74. **search_portfolio_github** - Search GitHub repositories for NEXS portfolios

### 🧠 Working Memory System (15 tools)

75. **working_memory_add** - Add entry to working memory with session scoping
76. **working_memory_get** - Retrieve working memory and record access
77. **working_memory_list** - List all memories in session with filters
78. **working_memory_promote** - Manually promote to long-term storage
79. **working_memory_clear_session** - Clear all memories in session
80. **working_memory_update** - Update existing working memory
81. **working_memory_delete** - Delete specific working memory
82. **working_memory_search** - Search within session memories
83. **working_memory_stats** - Get session statistics
84. **working_memory_extend_ttl** - Extend TTL of specific memory
85. **working_memory_set_priority** - Change memory priority
86. **working_memory_add_tags** - Add tags to existing memory
87. **working_memory_remove_tags** - Remove tags from memory
88. **working_memory_get_promoted** - List promoted memories
89. **working_memory_cleanup** - Manual cleanup trigger

**Features:**
- Session-scoped isolation
- Priority-based TTL (Low: 1h, Medium: 4h, High: 12h, Critical: 24h)
- Auto-promotion based on access patterns
- Background cleanup every 5 minutes
- Full metadata and tag support

**Documentation:** [Working Memory Tools API](docs/api/WORKING_MEMORY_TOOLS.md)

### ⚡ Token Optimization System (8 tools) **NEW in v1.3.0**

90. **deduplicate_memories** - Find and merge semantically similar memories (92%+ similarity threshold)
91. **optimize_context** - Optimize conversation context for token efficiency using all optimization services
92. **get_optimization_stats** - Comprehensive statistics for all 8 optimization services
93. **summarize_memory** - Summarize specific memory using TF-IDF extractive summarization
94. **compress_response** - Manually compress response using gzip/zlib (70-75% reduction)
95. **stream_large_list** - Stream large element lists in chunks (prevent memory overflow)
96. **batch_create_elements** - Create multiple elements in parallel (10x faster)
97. **get_cache_stats** - Adaptive cache statistics (access patterns, TTL distribution)

**8 Optimization Services:**
1. **Response Compression** - Gzip/zlib compression (70-75% size reduction)
2. **Streaming Handler** - Chunked streaming (prevent memory overflow)
3. **Semantic Deduplication** - Similarity-based duplicate detection (92%+ threshold)
4. **Auto-Summarization** - TF-IDF extractive summarization (70% compression)
5. **Context Window Manager** - Smart truncation strategies (preserve recent + relevant)
6. **Adaptive Cache** - Dynamic TTL based on access patterns (1h-7d)
7. **Batch Processing** - Parallel execution for bulk operations (10x faster)
8. **Prompt Compression** - Remove redundancies and fillers (35% reduction)

**Performance:**
- Overall token reduction: **81-95% in production workloads** (target: 90-95%)
- Zero additional latency overhead
- Configurable per-service via environment variables
- Comprehensive metrics and monitoring

**Documentation:** [Token Optimization System](docs/analysis/TOKEN_OPTIMIZATION_GAPS.md)

### 🎯 Memory Quality System (3 tools)

98. **score_memory_quality** - ONNX-based quality scoring with multi-tier fallback (ONNX → Groq → Gemini → Implicit)
99. **get_retention_policy** - Get retention policy for quality score (High: 365d, Medium: 180d, Low: 90d)
100. **get_retention_stats** - Memory retention statistics and quality distribution

**Features:**
- 2 ONNX models: MS MARCO (default, 61.64ms) and Paraphrase-Multilingual (configurable, 109.41ms)
- Multi-tier fallback system for reliability
- Automatic quality-based retention policies
- Zero cost, full privacy, offline-capable

**Documentation:** [ONNX Model Configuration](docs/user-guide/ONNX_MODEL_CONFIGURATION.md) | [Benchmarks](BENCHMARK_RESULTS.md)

For semantic search tools (73-74), see relationship system above.

For detailed tool documentation, see [docs/user-guide/QUICK_START.md](docs/user-guide/QUICK_START.md)

---

## 📦 Element Types

NEXS MCP supports **6 element types** for comprehensive AI system management:

| Element | Purpose | Key Features | Documentation |
|---------|---------|--------------|---------------|
| **Persona** | AI behavior and personality | Behavioral traits, expertise areas, communication style | [PERSONA.md](docs/elements/PERSONA.md) |
| **Skill** | Reusable capabilities | Triggers, procedures, execution strategies | [SKILL.md](docs/elements/SKILL.md) |
| **Template** | Content generation | Variable substitution, dynamic rendering | [TEMPLATE.md](docs/elements/TEMPLATE.md) |
| **Agent** | Autonomous workflows | Goals, planning, execution | [AGENT.md](docs/elements/AGENT.md) |
| **Memory** | Context persistence | Content storage, deduplication, search | [MEMORY.md](docs/elements/MEMORY.md) |
| **Ensemble** | Multi-agent orchestration | Sequential/parallel execution, voting, consensus | [ENSEMBLE.md](docs/elements/ENSEMBLE.md) |

### Quick Element Creation Examples

**Create a Persona:**
```json
{
  "tool": "quick_create_persona",
  "arguments": {
    "name": "Technical Writer",
    "description": "Expert in writing clear technical documentation",
    "expertise": ["documentation", "technical writing", "API design"],
    "traits": ["clear", "concise", "thorough"]
  }
}
```

**Create a Skill:**
```json
{
  "tool": "quick_create_skill",
  "arguments": {
    "name": "Code Review",
    "description": "Review code for best practices and bugs",
    "triggers": ["code review", "pr review"],
    "procedure": "1. Check code style\n2. Verify logic\n3. Suggest improvements"
  }
}
```

**Create an Ensemble:**
```json
{
  "tool": "quick_create_ensemble",
  "arguments": {
    "name": "Documentation Team",
    "description": "Multi-agent documentation generation",
    "members": ["persona:technical-writer", "agent:proofreader"],
    "execution_mode": "sequential",
    "aggregation_strategy": "merge"
  }
}
```

For complete element documentation, see [docs/elements/README.md](docs/elements/README.md)

---

## 💡 Usage Examples

### Basic Element Operations

**List all elements:**
```json
{
  "tool": "list_elements",
  "arguments": {
    "type": "persona",
    "active_only": true
  }
}
```

**Get element details:**
```json
{
  "tool": "get_element",
  "arguments": {
    "id": "persona-technical-writer"
  }
}
```

**Update element:**
```json
{
  "tool": "update_element",
  "arguments": {
    "id": "persona-technical-writer",
    "updates": {
      "expertise": ["documentation", "technical writing", "API design", "Markdown"]
    }
  }
}
```

### GitHub Integration

**Authenticate with GitHub:**
```json
{
  "tool": "github_auth_start",
  "arguments": {}
}
// Returns: user_code, verification_uri, expires_in
// Visit https://github.com/login/device and enter the code
```

**Sync portfolio to GitHub:**
```json
{
  "tool": "github_sync_push",
  "arguments": {
    "repo_owner": "yourusername",
    "repo_name": "my-ai-portfolio",
    "branch": "main",
    "commit_message": "Update personas and skills"
  }
}
```

**Pull elements from GitHub:**
```json
{
  "tool": "github_sync_pull",
  "arguments": {
    "repo_owner": "yourusername",
    "repo_name": "my-ai-portfolio",
    "branch": "main",
    "strategy": "newest-wins"
  }
}
```

### Collection Management

**Browse available collections:**
```json
{
  "tool": "browse_collections",
  "arguments": {
    "source": "github",
    "query": "technical writing"
  }
}
```

**Install a collection:**
```json
{
  "tool": "install_collection",
  "arguments": {
    "uri": "github://fsvxavier/nexs-collections/technical-writing",
    "force": false
  }
}
```

**Submit element to collection:**
```json
{
  "tool": "submit_element_to_collection",
  "arguments": {
    "element_id": "persona-technical-writer",
    "collection_repo": "fsvxavier/nexs-collections",
    "category": "personas"
  }
}
```

### Backup & Restore

**Create backup:**
```json
{
  "tool": "backup_portfolio",
  "arguments": {
    "output_path": "/backups/portfolio-2025-12-20.tar.gz",
    "compression": "best",
    "include_inactive": false
  }
}
```

**Restore from backup:**
```json
{
  "tool": "restore_portfolio",
  "arguments": {
    "backup_path": "/backups/portfolio-2025-12-20.tar.gz",
    "strategy": "merge",
    "dry_run": false
  }
}
```

### Memory Management

**Search memories:**
```json
{
  "tool": "search_memory",
  "arguments": {
    "query": "machine learning optimization techniques",
    "limit": 10,
    "min_relevance": 5
  }
}
```

**Summarize memories:**
```json
{
  "tool": "summarize_memories",
  "arguments": {
    "author_filter": "alice",
    "type_filter": "semantic"
  }
}
```

**Add to working memory:**
```json
{
  "tool": "working_memory_add",
  "arguments": {
    "session_id": "user-session-123",
    "content": "Meeting notes from today's standup",
    "priority": "high",
    "tags": ["meeting", "standup"]
  }
}
```

**Promote to long-term:**
```json
{
  "tool": "working_memory_promote",
  "arguments": {
    "session_id": "user-session-123",
    "memory_id": "working_memory_..."
  }
}
```

**Score memory quality:**
```json
{
  "tool": "score_memory_quality",
  "arguments": {
    "memory_id": "memory-xyz",
    "context": "technical documentation"
  }
}
```

### Analytics

**Get usage statistics:**
```json
{
  "tool": "get_usage_stats",
  "arguments": {
    "period": "30d",
    "include_top_n": 10
  }
}
```

**Performance dashboard:**
```json
{
  "tool": "get_performance_dashboard",
  "arguments": {
    "period": "7d"
  }
}
// Returns p50/p95/p99 latencies, slow operations, error rates
```

### Ensemble Execution

**Execute ensemble:**
```json
{
  "tool": "execute_ensemble",
  "arguments": {
    "ensemble_id": "documentation-team",
    "input": "Write API documentation for the /users endpoint",
    "context": {
      "api_version": "v2.0",
      "format": "OpenAPI"
    }
  }
}
```

For more examples, see:
- [Quick Start Guide](docs/user-guide/QUICK_START.md) - 10 hands-on tutorials
- [Examples Directory](examples/) - Complete workflows and integration examples

---

## 📁 Project Structure

```
nexs-mcp/
├── cmd/nexs-mcp/          # Application entrypoint
├── internal/
│   ├── domain/            # Business logic (79.2% coverage)
│   │   ├── element.go            # Base element interface
│   │   ├── persona.go            # Persona domain model
│   │   ├── skill.go              # Skill domain model
│   │   ├── template.go           # Template domain model
│   │   ├── agent.go              # Agent domain model
│   │   ├── memory.go             # Memory domain model
│   │   └── ensemble.go           # Ensemble domain model
│   ├── application/       # Use cases and services
│   │   ├── ensemble_executor.go  # Ensemble execution engine
│   │   ├── ensemble_monitor.go   # Real-time monitoring
│   │   ├── ensemble_aggregation.go # Voting & consensus
│   │   └── statistics.go         # Analytics service
│   ├── infrastructure/    # External adapters (68.1% coverage)
│   │   ├── repository.go          # In-memory repository
│   │   ├── file_repository.go     # File-based YAML repository
│   │   ├── github_client.go       # GitHub API client
│   │   ├── github_oauth.go        # OAuth2 device flow
│   │   ├── sync_conflict_detector.go  # Conflict resolution
│   │   ├── sync_metadata.go       # Sync state tracking
│   │   ├── sync_incremental.go    # Incremental sync
│   │   └── pr_tracker.go          # PR submission tracking
│   ├── mcp/              # MCP protocol layer (66.8% coverage)
│   │   ├── server.go             # MCP server (66 tools)
│   │   ├── tools.go              # Element CRUD tools
│   │   ├── quick_create_tools.go # Quick create tools
│   │   ├── collection_tools.go   # Collection management
│   │   ├── github_tools.go       # GitHub integration
│   │   ├── github_portfolio_tools.go # Portfolio sync
│   │   ├── backup_tools.go       # Backup & restore
│   │   ├── memory_tools.go       # Memory management
│   │   ├── log_tools.go          # Log querying
│   │   ├── user_tools.go         # User identity
│   │   ├── analytics_tools.go    # Usage & performance stats
│   │   └── ensemble_execution_tools.go # Ensemble execution
│   ├── backup/           # Backup & restore services (56.3% coverage)
│   ├── logger/           # Structured logging (92.1% coverage)
│   ├── config/           # Configuration (100% coverage)
│   ├── collection/       # Collection system (58.6% coverage)
│   ├── validation/       # Validation logic
│   └── portfolio/        # Portfolio management (75.6% coverage)
├── data/                 # File storage (gitignored)
│   └── elements/         # YAML element storage
├── docs/                 # Complete documentation
│   ├── user-guide/       # User documentation
│   │   ├── GETTING_STARTED.md   # Onboarding guide
│   │   ├── QUICK_START.md       # 10 tutorials
│   │   └── TROUBLESHOOTING.md   # Common issues
│   ├── elements/         # Element type documentation
│   ├── deployment/       # Deployment guides
│   ├── adr/             # Architecture Decision Records
│   └── README.md        # Documentation index
├── examples/            # Usage examples
│   ├── basic/           # Basic examples
│   ├── integration/     # Integration examples
│   └── workflows/       # Complete workflows
├── homebrew/            # Homebrew formula
├── .github/workflows/   # CI/CD pipelines
├── CHANGELOG.md         # Version history
├── COVERAGE_REPORT.md   # Test coverage analysis
├── NEXT_STEPS.md        # Development roadmap
├── docker-compose.yml   # Docker Compose config
├── Dockerfile           # Multi-stage Docker build
├── Makefile            # Build targets
└── go.mod              # Go module definition
```

---

## 🛠️ Development

### Prerequisites

- Go 1.25+
- Make (optional, for convenience targets)
- Docker (optional, for containerized deployment)

### Building

```bash
# Clone repository
git clone https://github.com/fsvxavier/nexs-mcp.git
cd nexs-mcp

# Install dependencies
go mod download

# Build binary
make build
# or
go build -o bin/nexs-mcp ./cmd/nexs-mcp

# Run tests
make test-coverage
# or
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

### Make Targets

```bash
make build             # Build binary
make test              # Run tests
make test-coverage     # Run tests with coverage report
make lint              # Run linters (golangci-lint)
make verify            # Run all verification steps
make ci                # Run full CI pipeline
make clean             # Clean build artifacts
```

### Running Locally

```bash
# Run with default settings (file storage)
./bin/nexs-mcp

# Run with custom data directory
./bin/nexs-mcp -data-dir ./my-elements

# Run in memory mode
./bin/nexs-mcp -storage memory

# Enable debug logging
./bin/nexs-mcp -log-level debug

# Run with environment variables
NEXS_DATA_DIR=./my-elements \
NEXS_STORAGE_TYPE=file \
NEXS_LOG_LEVEL=debug \
./bin/nexs-mcp
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/domain/...

# Run specific test
go test -run TestPersonaValidation ./internal/domain/

# Run with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 📚 Documentation

### User Documentation
- [Getting Started Guide](docs/user-guide/GETTING_STARTED.md) - Installation, first run, Claude Desktop integration
- [Quick Start Tutorial](docs/user-guide/QUICK_START.md) - 10 hands-on tutorials (2-5 min each)
- [ONNX Model Configuration](docs/user-guide/ONNX_MODEL_CONFIGURATION.md) - Quality scoring models (MS MARCO vs Paraphrase-Multilingual)
- [Troubleshooting Guide](docs/user-guide/TROUBLESHOOTING.md) - Common issues, FAQ, error codes
- [Documentation Index](docs/README.md) - Complete documentation navigation

### Element Types
- [Elements Overview](docs/elements/README.md) - Quick reference and relationships
- [Persona Documentation](docs/elements/PERSONA.md) - Behavioral traits and expertise
- [Skill Documentation](docs/elements/SKILL.md) - Triggers and procedures
- [Template Documentation](docs/elements/TEMPLATE.md) - Variable substitution
- [Agent Documentation](docs/elements/AGENT.md) - Goal-oriented workflows
- [Memory Documentation](docs/elements/MEMORY.md) - Content deduplication
- [Ensemble Documentation](docs/elements/ENSEMBLE.md) - Multi-agent orchestration

### Deployment
- [Docker Deployment](docs/deployment/DOCKER.md) - Complete Docker guide (600+ lines)
- [NPM Installation](README.npm.md) - NPM package usage
- [Homebrew Installation](homebrew/README.md) - Homebrew tap setup

### Architecture & Development
- [ADR-001: Hybrid Collection Architecture](docs/adr/ADR-001-hybrid-collection-architecture.md)
- [ADR-007: MCP Resources Implementation](docs/adr/ADR-007-mcp-resources-implementation.md)
- [ADR-008: Collection Registry Production](docs/adr/ADR-008-collection-registry-production.md)
- [ADR-009: Element Template System](docs/adr/ADR-009-element-template-system.md)
- [ADR-010: Missing Element Tools](docs/adr/ADR-010-missing-element-tools.md)
- [Test Coverage Report](COVERAGE_REPORT.md) - Coverage analysis and gaps

### Benchmarks & Quality
- [ONNX Benchmark Results](BENCHMARK_RESULTS.md) - Performance comparison of MS MARCO vs Paraphrase-Multilingual models
- [ONNX Quality Audit](ONNX_QUALITY_AUDIT.md) - Technical audit of quality system (80% conforme)
- [Quality Usage Analysis](QUALITY_USAGE_ANALYSIS.md) - Internal usage analysis (100% conforme)

### Project Planning
- [Roadmap](docs/next_steps/03_ROADMAP.md) - Future milestones
- [Next Steps](NEXT_STEPS.md) - Current development status
- [Changelog](CHANGELOG.md) - Version history and release notes

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test-coverage`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Standards

- Follow Clean Architecture principles
- Maintain test coverage (aim for 80%+)
- Use meaningful commit messages
- Document public APIs with godoc comments
- Run `make verify` before submitting PRs

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- Inspired by the [Model Context Protocol](https://modelcontextprotocol.io/) specification
- Thanks to all [contributors](https://github.com/fsvxavier/nexs-mcp/graphs/contributors)

---

## 📧 Support

- **Documentation**: [docs/README.md](docs/README.md)
- **Issues**: [GitHub Issues](https://github.com/fsvxavier/nexs-mcp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fsvxavier/nexs-mcp/discussions)

---

<div align="center">

**[⬆ Back to Top](#nexs-mcp-server)**

Made with ❤️ by the NEXS MCP team

</div>
