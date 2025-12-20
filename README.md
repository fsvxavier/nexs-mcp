# NEXS MCP Server

[![CI](https://github.com/fsvxavier/nexs-mcp/workflows/CI/badge.svg)](https://github.com/fsvxavier/nexs-mcp/actions)
[![Coverage](https://img.shields.io/badge/coverage-72.2%25-yellow)](./COVERAGE_REPORT.md)
[![Go Version](https://img.shields.io/badge/go-1.25-blue)](https://go.dev)
[![Release](https://img.shields.io/badge/release-v0.6.0--dev-blue)](https://github.com/fsvxavier/nexs-mcp/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)
[![MCP SDK](https://img.shields.io/badge/MCP_SDK-v1.1.0-blue)](https://github.com/modelcontextprotocol/go-sdk)
[![Tools](https://img.shields.io/badge/MCP_Tools-51-brightgreen)](#-available-tools)

**Model Context Protocol (MCP) Server implementation in Go** - A high-performance, production-ready MCP server with Clean Architecture using the official MCP Go SDK.

## 🎯 Project Overview

NEXS MCP Server is a Go implementation of the [Model Context Protocol](https://modelcontextprotocol.io/), designed to manage AI elements (Personas, Skills, Templates, Agents, Memories, and Ensembles) with enterprise-grade architecture and high test coverage.

Built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) v1.1.0 for robust and standard-compliant MCP communication.

### Key Features

#### Core Infrastructure
- ✅ **Official MCP SDK** - Built on github.com/modelcontextprotocol/go-sdk v1.1.0
- ✅ **Clean Architecture** - Domain-driven design with clear separation of concerns
- ✅ **High Test Coverage** - 72.2% overall (Logger 92.1%, Config 100%, Domain 79.2%)
- ✅ **Dual Storage Modes** - File-based YAML or in-memory
- ✅ **51 MCP Tools** - Complete portfolio, production, and analytics tooling
- ✅ **6 Element Types** - Persona, Skill, Template, Agent, Memory, Ensemble
- ✅ **Stdio Transport** - Standard MCP communication over stdin/stdout
- ✅ **Thread-Safe** - Concurrent operations with proper synchronization
- ✅ **Cross-Platform** - Binaries for Linux, macOS, Windows (amd64/arm64)

#### Production Readiness (M0.5) ✨
- ✅ **Auto-Save Feature** - Automatic conversation context preservation (default enabled)
- ✅ **Quick Create Tools** - Simplified element creation with template defaults (minimal prompts)
- ✅ **Backup & Restore** - Portfolio backup with tar.gz compression and SHA-256 checksums
- ✅ **Memory Management** - Search, summarize, update memories with relevance scoring
- ✅ **Structured Logging** - slog-based JSON/text logs with context extraction
- ✅ **Log Query Tools** - Filter and search logs by level, user, operation, tool
- ✅ **User Identity** - Session management with metadata support
- ✅ **GitHub OAuth** - Device flow authentication and token management
- ✅ **Collection System** - Install, manage, and publish element collections
- ✅ **GitHub Integration** - Sync portfolios with GitHub repositories

## 📊 Current Status

```
Version:               v0.6.0-dev (Milestone M0.6 - 72% Complete)
Logger Package:         92.1% ✓
Config Package:        100.0% ✓
Domain Layer:           79.2% ✓
Infrastructure Layer:   68.1%
MCP Layer:              66.8%
Overall Coverage:       72.2%
Lines of Code:         9,200+
Test Cases:            182+ (unit + integration)
MCP Tools:             51 (Element CRUD + Quick Create + Production + Analytics)
Element Types:         6 (Persona, Skill, Template, Agent, Memory, Ensemble)
```

**Milestone M0.6 Analytics & Convenience (19/12/2025) - 5/7 Tasks Complete:**
- ✅ list_elements active_only filter (resolves get_active_elements gap)
- ✅ duplicate_element tool (resolves duplication gap)
- ✅ get_usage_stats analytics (period filtering, top-10, success rates)
- ✅ get_performance_dashboard (p50/p95/p99 latencies, slow ops)
- 🔄 Documentation Updates (in progress)
- ⏳ Test Coverage Improvements (deferred to gradual improvement)
- ⏳ Release v0.6.0 (pending)

**Previous Milestones:**
- ✅ M0.5 Production Readiness (19/12/2025) - Backup, memory, logging, auth
- ✅ M0.4 Collection System (18/12/2025) - 10 collection tools + GitHub sync
- ✅ M0.2 Element Types (18/12/2025) - 6 element types + documentation

## 🚀 Quick Start

### Installation

#### Option 1: Go Install (Recommended)

```bash
go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@v1.0.0
```

#### Option 2: Homebrew (macOS/Linux)

```bash
# Add tap
brew tap fsvxavier/nexs-mcp

# Install
brew install nexs-mcp
```

#### Option 3: NPM

```bash
npm install -g @fsvxavier/nexs-mcp-server
```

#### Option 4: Docker

```bash
docker pull fsvxavier/nexs-mcp:latest
docker run -v $(pwd)/data:/app/data fsvxavier/nexs-mcp:latest
```

#### Option 5: Build from Source

```bash
# Clone the repository
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

### Usage

The server supports two storage modes:

**File Storage (default):**
```bash
# Default configuration (file storage, data/elements directory)
./bin/nexs-mcp

# Custom data directory
./bin/nexs-mcp -data-dir /path/to/data

# Or via environment variable
NEXS_DATA_DIR=/path/to/data ./bin/nexs-mcp
```

**In-Memory Storage:**
```bash
# Memory-only storage (data lost on restart)
./bin/nexs-mcp -storage memory

# Or via environment variable
NEXS_STORAGE_TYPE=memory ./bin/nexs-mcp
```

**Output:**
```bash
NEXS MCP Server v0.1.0
Initializing Model Context Protocol server...
Storage type: file
Data directory: data/elements
Registered 5 tools
Server ready. Listening on stdio...
```

## 🔧 Available Tools

### Element Management (11 tools)

#### Generic CRUD Operations
1. **list_elements** - List all elements with advanced filtering
2. **get_element** - Get element details by ID
3. **create_element** - Create generic element
4. **update_element** - Update existing element
5. **delete_element** - Delete element by ID

#### Type-Specific Creation
6. **create_persona** - Create Persona with behavioral traits
7. **create_skill** - Create Skill with triggers and procedures
8. **create_template** - Create Template with variable substitution
9. **create_agent** - Create Agent with goals and workflows
10. **create_memory** - Create Memory with content hashing
11. **create_ensemble** - Create Ensemble for multi-agent orchestration

### Collection System (10 tools)

12. **browse_collections** - Discover available collections (GitHub, local, HTTP)
13. **install_collection** - Install collection from URI (github://, file://, https://)
14. **uninstall_collection** - Remove installed collection
15. **list_installed_collections** - List all installed collections
16. **get_collection_info** - Get detailed collection information
17. **export_collection** - Export collection to tar.gz archive
18. **update_collection** - Update specific collection
19. **update_all_collections** - Update all installed collections
20. **check_collection_updates** - Check for available updates
21. **publish_collection** - Publish collection to GitHub

### GitHub Integration (5 tools)

22. **github_auth_start** - Initiate OAuth2 device flow authentication
23. **github_auth_status** - Check GitHub authentication status
24. **github_list_repos** - List user's GitHub repositories
25. **github_sync_push** - Push local elements to GitHub repository
26. **github_sync_pull** - Pull elements from GitHub repository

### Production Tools (18 tools) ✨ NEW

#### Backup & Restore
27. **backup_portfolio** - Create compressed backup with checksums
28. **restore_portfolio** - Restore from backup with validation
29. **activate_element** - Activate element (shortcut for update)
30. **deactivate_element** - Deactivate element (shortcut for update)

#### Memory Management
31. **search_memory** - Search memories with relevance scoring
32. **summarize_memories** - Get memory statistics and summaries
33. **update_memory** - Partial update of memory content
34. **delete_memory** - Delete specific memory
35. **clear_memories** - Bulk delete memories with filters

#### Logging & Monitoring
36. **list_logs** - Query logs with filters (level, date, user, operation, tool)

#### User Identity
37. **get_current_user** - Get current user session information
38. **set_user_context** - Set user identity with metadata
39. **clear_user_context** - Clear current user session

#### GitHub Authentication
40. **check_github_auth** - Verify GitHub token and get user info
41. **refresh_github_token** - Refresh GitHub OAuth token
42. **init_github_auth** - Initialize GitHub device flow authentication

#### Analytics & Convenience (M0.6) ✨ NEW
43. **duplicate_element** - Duplicate element with new ID and optional name
44. **get_usage_stats** - Analytics with period filtering and top-10 rankings
45. **get_performance_dashboard** - Performance metrics with p50/p95/p99 latencies

#### Context Management
46. **get_context** - Get MCP server context information
47. **search_elements** - Advanced element search with filters

**Total: 47 MCP Tools** (28 existing + 19 production + analytics tools)

## 📦 Element Types

NEXS MCP supports 6 element types for comprehensive AI system management:

| Element | Purpose | Documentation |
|---------|---------|---------------|
| **Persona** | AI behavior and personality customization | [docs/elements/PERSONA.md](docs/elements/PERSONA.md) |
| **Skill** | Reusable capabilities with triggers | [docs/elements/SKILL.md](docs/elements/SKILL.md) |
| **Template** | Content templates with variable substitution | [docs/elements/TEMPLATE.md](docs/elements/TEMPLATE.md) |
| **Agent** | Goal-oriented autonomous workflows | [docs/elements/AGENT.md](docs/elements/AGENT.md) |
| **Memory** | Content storage with deduplication | [docs/elements/MEMORY.md](docs/elements/MEMORY.md) |
| **Ensemble** | Multi-agent coordination and orchestration | [docs/elements/ENSEMBLE.md](docs/elements/ENSEMBLE.md) |

For complete element system documentation, see [docs/elements/README.md](docs/elements/README.md)

## � Usage Examples

### Production Tools (M0.5)

#### Backup & Restore
```json
// Create a backup
{
  "tool": "backup_portfolio",
  "arguments": {
    "output_path": "/backups/portfolio-2025-12-19.tar.gz",
    "compression": "best",
    "include_inactive": false
  }
}

// Restore from backup
{
  "tool": "restore_portfolio",
  "arguments": {
    "backup_path": "/backups/portfolio-2025-12-19.tar.gz",
    "strategy": "merge",
    "dry_run": false
  }
}
```

#### Memory Management
```json
// Search memories with relevance scoring
{
  "tool": "search_memory",
  "arguments": {
    "query": "machine learning optimization",
    "limit": 10,
    "min_relevance": 5
  }
}

// Summarize memories
{
  "tool": "summarize_memories",
  "arguments": {
    "author_filter": "alice",
    "type_filter": "semantic"
  }
}
```

#### User Identity & Logging
```json
// Set user context
{
  "tool": "set_user_context",
  "arguments": {
    "username": "alice",
    "metadata": {
      "team": "ml-engineering",
      "role": "senior-engineer"
    }
  }
}

// Query logs
{
  "tool": "list_logs",
  "arguments": {
    "level": "error",
    "user": "alice",
    "operation": "backup_portfolio",
    "limit": 50
  }
}
```

#### GitHub Authentication
```json
// Initialize GitHub OAuth
{
  "tool": "init_github_auth",
  "arguments": {}
}
// Returns: user_code, verification_uri, expires_in

// Check authentication status
{
  "tool": "check_github_auth",
  "arguments": {}
}
// Returns: authenticated, username, token_expiry, scopes
```

### Collection System
```json
// Browse available collections
{
  "tool": "browse_collections",
  "arguments": {
    "source": "github",
    "query": "personas"
  }
}

// Install collection
{
  "tool": "install_collection",
  "arguments": {
    "uri": "github://fsvxavier/nexs-collections/personas",
    "force": false
  }
}

// Sync with GitHub
{
  "tool": "github_sync_push",
  "arguments": {
    "repo_owner": "fsvxavier",
    "repo_name": "my-portfolio",
    "branch": "main",
    "commit_message": "Update personas"
  }
}
```

For more examples, see [examples/](./examples/) directory.

## 📁 Project Structure

```
nexs-mcp/
├── cmd/nexs-mcp/          # Application entrypoint
├── internal/
│   ├── domain/            # Business logic (79.2% coverage)
│   ├── infrastructure/    # External adapters (68.1% coverage)
│   │   ├── repository.go          # In-memory repository
│   │   ├── file_repository.go     # File-based YAML repository
│   │   ├── github_client.go       # GitHub API client
│   │   └── github_oauth.go        # OAuth2 device flow
│   ├── mcp/              # MCP protocol layer (66.8% coverage)
│   │   ├── server.go             # MCP server (44 tools)
│   │   ├── tools.go              # Element CRUD tools
│   │   ├── collection_tools.go   # Collection management
│   │   ├── github_tools.go       # GitHub integration
│   │   ├── backup_tools.go       # Backup & restore
│   │   ├── memory_tools.go       # Memory management
│   │   ├── log_tools.go          # Log querying
│   │   ├── user_tools.go         # User identity
│   │   └── github_auth_tools.go  # GitHub auth
│   ├── backup/           # Backup & restore services (56.3% coverage)
│   ├── logger/           # Structured logging (92.1% coverage)
│   ├── config/           # Configuration (100% coverage)
│   ├── collection/       # Collection system (58.6% coverage)
│   └── portfolio/        # Portfolio management (75.6% coverage)
├── data/                 # File storage (gitignored)
├── docs/                 # Complete documentation
│   ├── elements/         # Element type documentation
│   ├── plano/           # Strategic planning
│   └── next_steps/      # Roadmap and milestones
├── examples/            # Usage examples
├── CHANGELOG.md         # Version history
├── COVERAGE_REPORT.md   # Test coverage analysis
├── Makefile
└── go.mod
```

## 🛠️ Development

### Make Targets

```bash
make build             # Build binary
make test-coverage     # Run tests with coverage
make lint              # Run linters
make verify            # Run all verification steps
make ci                # Run full CI pipeline
```

## 📚 Documentation

### Quick Start
- [Installation & Usage](#-quick-start) - Get started in 5 minutes
- [Available Tools](#-available-tools) - Complete tool reference (44 tools)
- [Usage Examples](#-usage-examples) - Common workflows and patterns

### Element System
- [Element Types Overview](./docs/elements/README.md) - Quick reference and relationships
- [Persona Documentation](./docs/elements/PERSONA.md) - Behavioral traits and expertise
- [Skill Documentation](./docs/elements/SKILL.md) - Triggers and procedures
- [Template Documentation](./docs/elements/TEMPLATE.md) - Variable substitution
- [Agent Documentation](./docs/elements/AGENT.md) - Goal-oriented workflows
- [Memory Documentation](./docs/elements/MEMORY.md) - Content deduplication
- [Ensemble Documentation](./docs/elements/ENSEMBLE.md) - Multi-agent orchestration

### Production Features (M0.5)
- [Auto-Save Feature](./docs/AUTO_SAVE.md) - Automatic conversation context preservation
- [Quick Create Tools](./docs/MCP_UX_GUIDELINES.md) - Simplified element creation (minimal confirmations)
- [MCP UX Guidelines](./docs/MCP_UX_GUIDELINES.md) - Understanding client-server separation
- [Backup & Restore](./internal/backup/) - Portfolio backup with tar.gz compression
- [Structured Logging](./internal/logger/) - slog-based logging with filtering
- [User Identity](./internal/mcp/user_tools.go) - Session management
- [GitHub OAuth](./internal/mcp/github_auth_tools.go) - Device flow authentication
- [Test Coverage Report](./COVERAGE_REPORT.md) - Coverage analysis and gaps

### Project Documentation
- [CHANGELOG](./CHANGELOG.md) - Version history and release notes
- [Strategic Plan](./docs/plano/01_README.md) - Project vision and goals
- [Architecture](./docs/plano/03_ARCHITECTURE.md) - System design
- [Roadmap](./docs/next_steps/03_ROADMAP.md) - Future milestones
- [Next Steps](./NEXT_STEPS.md) - Current development status

### Contributing
- Test Coverage: See [COVERAGE_REPORT.md](./COVERAGE_REPORT.md)
- Development: Run `make test-coverage` to verify changes
- Code Style: Follow Clean Architecture principles

## 📝 License

MIT License
