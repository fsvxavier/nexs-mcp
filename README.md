# NEXS MCP Server

[![CI](https://github.com/fsvxavier/nexs-mcp/workflows/CI/badge.svg)](https://github.com/fsvxavier/nexs-mcp/actions)
[![Coverage](https://img.shields.io/badge/coverage-80.7%25-green)](./coverage.html)
[![Go Version](https://img.shields.io/badge/go-1.25-blue)](https://go.dev)
[![Release](https://img.shields.io/badge/release-v0.1.0-blue)](https://github.com/fsvxavier/nexs-mcp/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)
[![MCP SDK](https://img.shields.io/badge/MCP_SDK-Official-blue)](https://github.com/modelcontextprotocol/go-sdk)

**Model Context Protocol (MCP) Server implementation in Go** - A high-performance, production-ready MCP server with Clean Architecture using the official MCP Go SDK.

## 🎯 Project Overview

NEXS MCP Server is a Go implementation of the [Model Context Protocol](https://modelcontextprotocol.io/), designed to manage AI elements (Personas, Skills, Templates, Agents, Memories, and Ensembles) with enterprise-grade architecture and high test coverage.

Built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) v1.1.0 for robust and standard-compliant MCP communication.

### Key Features

- ✅ **Official MCP SDK** - Built on github.com/modelcontextprotocol/go-sdk v1.1.0
- ✅ **Clean Architecture** - Domain-driven design with clear separation of concerns
- ✅ **High Test Coverage** - 80.7% overall (Domain 76.4%, Infrastructure 87.7%, MCP 79.0%)
- ✅ **Dual Storage Modes** - File-based YAML or in-memory
- ✅ **11 MCP Tools** - 5 generic CRUD + 6 type-specific creation tools
- ✅ **6 Element Types** - Persona, Skill, Template, Agent, Memory, Ensemble
- ✅ **Stdio Transport** - Standard MCP communication over stdin/stdout
- ✅ **Configurable** - Environment variables and command-line flags
- ✅ **Thread-Safe** - Concurrent operations with proper synchronization
- ✅ **Cross-Platform** - Binaries for Linux, macOS, Windows (amd64/arm64)
- ✅ **Production Ready** - Graceful shutdown, error handling, full MCP protocol support

## 📊 Current Status

```
Version:               v0.2.0-dev (Milestone M0.2 Complete)
Domain Layer:           76.4% ✓
Infrastructure Layer:   87.7% ✓
MCP Layer:              79.0% ✓
Overall Coverage:       80.7%
Lines of Code:         4,800+
Test Cases:            124+ (unit + integration)
MCP Tools:             11 (5 generic + 6 type-specific)
Element Types:         6 (Persona, Skill, Template, Agent, Memory, Ensemble)
```

**Milestone M0.2 Completed (18/12/2025):**
- ✅ 6 element types fully implemented with domain logic
- ✅ 6 type-specific MCP handlers (create_persona, create_skill, etc.)
- ✅ Complete documentation (~800 lines) for all element types
- ✅ Integration tests (6 test functions) demonstrating element interactions
- ✅ Unit test coverage for all handlers (18 test functions, 100% passing)

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- Make

### Installation

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

### Generic CRUD Operations
1. **list_elements** - List all elements with filtering
2. **get_element** - Get element by ID
3. **create_element** - Create new element (generic)
4. **update_element** - Update existing element
5. **delete_element** - Delete element by ID

### Type-Specific Element Creation
6. **create_persona** - Create Persona with behavioral traits and expertise
7. **create_skill** - Create Skill with triggers and procedures
8. **create_template** - Create Template with variable substitution
9. **create_agent** - Create Agent with goals and action workflows
10. **create_memory** - Create Memory with content hashing
11. **create_ensemble** - Create Ensemble for multi-agent orchestration

**Total:** 11 MCP Tools

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

## 📁 Project Structure

```
nexs-mcp/
├── cmd/nexs-mcp/          # Application entrypoint
├── internal/
│   ├── domain/            # Business logic (100% coverage)
│   ├── infrastructure/    # External adapters (98.5% coverage)
│   │   ├── repository.go          # In-memory repository
│   │   └── file_repository.go     # File-based repository
│   ├── mcp/              # MCP protocol layer (96.8% coverage)
│   ├── config/           # Configuration management
│   └── application/      # Use cases (planned)
├── data/                 # File storage (gitignored)
├── docs/                 # Complete documentation
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

### Element System
- [Element Types Overview](./docs/elements/README.md) - Quick reference and relationships
- [Persona Documentation](./docs/elements/PERSONA.md) - Behavioral traits and expertise
- [Skill Documentation](./docs/elements/SKILL.md) - Triggers and procedures
- [Template Documentation](./docs/elements/TEMPLATE.md) - Variable substitution
- [Agent Documentation](./docs/elements/AGENT.md) - Goal-oriented workflows
- [Memory Documentation](./docs/elements/MEMORY.md) - Content deduplication
- [Ensemble Documentation](./docs/elements/ENSEMBLE.md) - Multi-agent orchestration

### Project Documentation
- [Strategic Plan](./docs/plano/01_README.md)
- [Architecture](./docs/plano/03_ARCHITECTURE.md)
- [Roadmap](./docs/next_steps/03_ROADMAP.md)
- [Next Steps](./NEXT_STEPS.md)

## 📝 License

MIT License
