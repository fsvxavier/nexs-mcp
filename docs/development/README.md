# Development Documentation

Complete guides for developers working on NEXS-MCP.

## Getting Started

- **[SETUP.md](./SETUP.md)** - Development environment setup
- **[CODE_TOUR.md](./CODE_TOUR.md)** - Complete code walkthrough (1,632 lines)

## Core Tutorials

### [CODE_TOUR.md](./CODE_TOUR.md) (800+ lines)
Complete architectural walkthrough covering:
- Entry point and initialization flow
- Configuration and logging systems
- Repository patterns and data persistence
- MCP Server implementation using official SDK
- Tool registration and request/response flow
- Key packages and interfaces
- Data flow diagrams
- Quick reference guide

**Read this first** to understand the codebase architecture.

### [ADDING_ELEMENT_TYPE.md](./ADDING_ELEMENT_TYPE.md) (600+ lines)
Step-by-step tutorial for adding new element types:
- Define domain model
- Implement Element interface
- Create comprehensive validator
- Update repository implementations
- Add MCP tools (create, quick_create)
- Register tools with SDK
- Write complete test suite
- Update documentation

**Complete example:** "Workflow" element type implementation.

### [ADDING_MCP_TOOL.md](./ADDING_MCP_TOOL.md) (600+ lines)
Tutorial for creating new MCP tools:
- Understanding MCP tool structure
- Define input/output schemas
- Implement handler functions
- Register with official SDK
- Business logic patterns
- Error handling strategies
- Metrics and performance tracking
- Comprehensive testing

**Complete example:** `validate_template` tool implementation.

### [EXTENDING_VALIDATION.md](./EXTENDING_VALIDATION.md) (500+ lines)
Guide for adding validation rules:
- Validation architecture overview
- Types of validation (structural, business, content quality)
- Multi-level validation (basic, comprehensive, strict)
- Implementing validation logic
- Error/warning messages best practices
- Testing validation rules
- Backward compatibility

**Examples:** Email, date range, reference, complexity, and security validation.

## Testing & Quality

- **[TESTING.md](./TESTING.md)** - Testing strategy and guidelines
- **[RELEASE.md](./RELEASE.md)** - Release process and versioning

## Quick Links

### Common Tasks

| Task | Document | Section |
|------|----------|---------|
| Add a new element type | [ADDING_ELEMENT_TYPE.md](./ADDING_ELEMENT_TYPE.md) | Full tutorial |
| Create a new MCP tool | [ADDING_MCP_TOOL.md](./ADDING_MCP_TOOL.md) | Full tutorial |
| Add validation rule | [EXTENDING_VALIDATION.md](./EXTENDING_VALIDATION.md) | Step 3 |
| Understand architecture | [CODE_TOUR.md](./CODE_TOUR.md) | Architecture Overview |
| Find specific code | [CODE_TOUR.md](./CODE_TOUR.md) | Quick Reference Guide |

### Key Concepts

- **MCP SDK Integration**: All tools use `github.com/modelcontextprotocol/go-sdk/mcp`
- **Element Interface**: Base abstraction for all capability types
- **Repository Pattern**: Data persistence abstraction layer
- **Multi-level Validation**: Basic → Comprehensive → Strict
- **Tool Handler Pattern**: Parse → Validate → Execute → Return

## Documentation Standards

All tutorials follow these standards:
- **Table of contents** for easy navigation
- **Code examples** with syntax highlighting
- **ASCII diagrams** where helpful
- **Common mistakes** section
- **Troubleshooting** tips
- **Links** to related documentation
- **Complete working examples**

## Contributing

When adding new features:

1. **Read relevant tutorial** to understand the pattern
2. **Follow existing conventions** for consistency
3. **Write comprehensive tests** (aim for 80%+ coverage)
4. **Update documentation** including API docs
5. **Add examples** to help future developers

## File Organization

```
docs/development/
├── README.md                    # This file
├── SETUP.md                     # Environment setup
├── CODE_TOUR.md                 # Complete code walkthrough ⭐
├── ADDING_ELEMENT_TYPE.md       # Element type tutorial ⭐
├── ADDING_MCP_TOOL.md           # MCP tool tutorial ⭐
├── EXTENDING_VALIDATION.md      # Validation guide ⭐
├── TESTING.md                   # Testing guidelines
└── RELEASE.md                   # Release process

⭐ = New comprehensive tutorial (6,400+ lines total)
```

## Architecture Highlights

### MCP Server Stack

```
┌─────────────────────────────────────────┐
│     MCP Client (Claude Desktop)         │
└──────────────┬──────────────────────────┘
               │ JSON-RPC over stdio
               ▼
┌─────────────────────────────────────────┐
│  Official MCP SDK (go-sdk/mcp)          │
│  - Tool registration                    │
│  - Schema generation                    │
│  - Protocol handling                    │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  NEXS-MCP Business Logic                │
│  - 55 Tools                             │
│  - 3 Resources                          │
│  - 6 Element Types                      │
│  - Multi-level Validation               │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Data Layer                             │
│  - File Repository (YAML)               │
│  - In-Memory Repository (testing)       │
│  - TF-IDF Index                         │
└─────────────────────────────────────────┘
```

### Key Packages

- `cmd/nexs-mcp/` - Entry point
- `internal/config/` - Configuration management
- `internal/domain/` - Domain models and interfaces
- `internal/validation/` - Validation rules
- `internal/infrastructure/` - Data persistence
- `internal/mcp/` - MCP server and tools
- `internal/application/` - Business logic services
- `internal/indexing/` - Search and indexing
- `internal/template/` - Template rendering

## Additional Resources

- **API Documentation**: [docs/api/](../api/)
  - [MCP_TOOLS.md](../api/MCP_TOOLS.md) - Complete tool reference
  - [MCP_RESOURCES.md](../api/MCP_RESOURCES.md) - Resource endpoints
  - [CLI.md](../api/CLI.md) - Command-line interface

- **User Guides**: [docs/user-guide/](../user-guide/)
  - [QUICK_START.md](../user-guide/QUICK_START.md) - Get started quickly
  - [GETTING_STARTED.md](../user-guide/GETTING_STARTED.md) - Detailed tutorial
  - [TROUBLESHOOTING.md](../user-guide/TROUBLESHOOTING.md) - Common issues

- **Architecture**: [docs/architecture/](../architecture/)
  - [OVERVIEW.md](../architecture/OVERVIEW.md) - System overview
  - [DOMAIN.md](../architecture/DOMAIN.md) - Domain model
  - [APPLICATION.md](../architecture/APPLICATION.md) - Application layer
  - [INFRASTRUCTURE.md](../architecture/INFRASTRUCTURE.md) - Infrastructure
  - [MCP.md](../architecture/MCP.md) - MCP integration

## Feedback

Found an issue or have suggestions? Please:
1. Check existing documentation
2. Review code examples
3. Open an issue with details
4. Contribute improvements via PR

---

**Last Updated**: December 20, 2025
**Tutorial Stats**: 6,434 lines across 4 comprehensive guides
