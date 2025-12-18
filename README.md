# NEXS MCP Server

[![CI](https://github.com/fsvxavier/nexs-mcp/workflows/CI/badge.svg)](https://github.com/fsvxavier/nexs-mcp/actions)
[![Coverage](https://img.shields.io/badge/coverage-80.7%25-green)](./coverage.html)
[![Go Version](https://img.shields.io/badge/go-1.25-blue)](https://go.dev)
[![Release](https://img.shields.io/badge/release-v0.1.0-blue)](https://github.com/fsvxavier/nexs-mcp/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

**Model Context Protocol (MCP) Server implementation in Go** - A high-performance, production-ready MCP server with Clean Architecture.

## 🎯 Project Overview

NEXS MCP Server is a Go implementation of the [Model Context Protocol](https://modelcontextprotocol.io/), designed to manage AI elements (Personas, Skills, Templates, Agents, Memories, and Ensembles) with enterprise-grade architecture and high test coverage.

### Key Features

- ✅ **Clean Architecture** - Domain-driven design with clear separation of concerns
- ✅ **High Test Coverage** - 80.7% overall (Domain 100%, Infrastructure 87.7%, MCP 94%)
- ✅ **Dual Storage Modes** - File-based YAML or in-memory
- ✅ **5 MCP Tools** - Complete CRUD operations
- ✅ **6 Element Types** - Comprehensive element management
- ✅ **Configurable** - Environment variables and command-line flags
- ✅ **Thread-Safe** - Concurrent operations with proper synchronization
- ✅ **Cross-Platform** - Binaries for Linux, macOS, Windows (amd64/arm64)
- ✅ **Production Ready** - Graceful shutdown, error handling, JSON-RPC protocol

## 📊 Current Status

```
Version:               v0.1.0 (Production Ready)
Domain Layer:          100.0% ✓
Infrastructure Layer:   87.7% ✓
MCP Layer:              94.0% ✓
Overall Coverage:       80.7%
Lines of Code:         3,155
Test Cases:            100+
```

**Implemented:**
- ✅ MCP Server with JSON-RPC 2.0
- ✅ 5 CRUD tools (list, get, create, update, delete)
- ✅ File-based persistence (YAML)
- ✅ In-memory repository
- ✅ Configuration system
- ✅ Element type system (6 types)
- ✅ Thread-safe operations
- ✅ Graceful shutdown
- ✅ Cross-platform binaries
- ✅ Docker support

**Ready for Release:**
- 🎯 Version 0.1.0 complete
- 🎯 Production ready
- 🎯 Comprehensive documentation
- 🎯 Claude Desktop integration

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

1. **list_elements** - List all elements with filtering
2. **get_element** - Get element by ID
3. **create_element** - Create new element
4. **update_element** - Update existing element
5. **delete_element** - Delete element by ID

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

- [Strategic Plan](./docs/plano/01_README.md)
- [Architecture](./docs/plano/03_ARCHITECTURE.md)
- [Roadmap](./docs/next_steps/03_ROADMAP.md)

## 📝 License

MIT License
