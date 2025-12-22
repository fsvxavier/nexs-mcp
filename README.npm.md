# NEXS MCP Server - NPM Package

[![npm version](https://img.shields.io/npm/v/@fsvxavier/nexs-mcp-server.svg)](https://www.npmjs.com/package/@fsvxavier/nexs-mcp-server)
[![npm downloads](https://img.shields.io/npm/dm/@fsvxavier/nexs-mcp-server.svg)](https://www.npmjs.com/package/@fsvxavier/nexs-mcp-server)
[![License](https://img.shields.io/badge/license-MIT-green)](https://github.com/fsvxavier/nexs-mcp/blob/main/LICENSE)

A production-ready Model Context Protocol (MCP) server for managing AI elements (Personas, Skills, Templates, Agents, Memories, and Ensembles). **Features intelligent token optimization that reduces AI context usage by 70-85%** through multilingual keyword extraction and conversation memory management across 11 languages.

---

## 🚀 Quick Start

### Installation

Install globally via NPM:

```bash
npm install -g @fsvxavier/nexs-mcp-server
```

Or use with `npx` without installation:

```bash
npx @fsvxavier/nexs-mcp-server
```

### Verify Installation

```bash
nexs-mcp --version
# Output: NEXS MCP Server v1.0.5
```

### First Run

```bash
# Run with default configuration (file storage in data/elements)
nexs-mcp

# Run with custom data directory
nexs-mcp -data-dir /path/to/data

# Run in memory mode
nexs-mcp -storage memory
```

**Output:**
```
NEXS MCP Server v1.0.5
Initializing Model Context Protocol server...
Storage type: file
Data directory: data/elements
Registered 66 tools
Server ready. Listening on stdio...
```

---

## 📦 What's Included

This NPM package includes:

- **Cross-platform binaries** for:
  - macOS (Intel and Apple Silicon)
  - Linux (amd64 and arm64)
  - Windows (amd64)
- **Automatic platform detection** and binary selection
- **66 MCP tools** for comprehensive AI element management
- **6 element types**: Persona, Skill, Template, Agent, Memory, Ensemble
- **Dual storage modes**: File-based (YAML) or in-memory
- **💰 Token optimization**: 70-85% reduction in AI context usage
- **🌍 Multilingual support**: 11 languages (EN, PT, ES, FR, DE, IT, RU, JA, ZH, AR, HI) with automatic detection

---

## 🔧 Usage

### Command Line

```bash
# Basic usage
nexs-mcp

# Custom data directory
nexs-mcp -data-dir ./my-elements

# Memory-only storage
nexs-mcp -storage memory

# Enable debug logging
nexs-mcp -log-level debug

# Show help
nexs-mcp --help
```

### Environment Variables

You can also configure via environment variables:

```bash
# Set data directory
export NEXS_DATA_DIR=/path/to/data

# Set storage type
export NEXS_STORAGE_TYPE=file  # or 'memory'

# Set log level
export NEXS_LOG_LEVEL=debug  # or 'info', 'warn', 'error'

# Run
nexs-mcp
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

**Or use npx if you don't want global installation:**

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "npx",
      "args": ["-y", "@fsvxavier/nexs-mcp-server"],
      "env": {
        "NEXS_DATA_DIR": "/path/to/your/elements",
        "NEXS_STORAGE_TYPE": "file"
      }
    }
  }
}
```

Restart Claude Desktop and you'll see NEXS MCP tools available!

---

## ✨ Features

### Core Capabilities

- **💰 Token Optimization** - 70-85% reduction in AI context usage through intelligent conversation memory
- **🌍 Multilingual Support** - 11 languages (EN, PT, ES, FR, DE, IT, RU, JA, ZH, AR, HI) with automatic detection
- **66 MCP Tools** - Complete portfolio management, GitHub integration, analytics
- **6 Element Types** - Persona, Skill, Template, Agent, Memory, Ensemble
- **Dual Storage** - File-based (YAML) or in-memory
- **GitHub Integration** - OAuth, portfolio sync, collection management, PR submission
- **Production Features** - Backup/restore, memory management, logging, analytics
- **Ensemble Execution** - Sequential/parallel/hybrid with voting and consensus

### Element Types

| Element | Purpose | Key Features |
|---------|---------|--------------|
| **Persona** | AI behavior and personality | Traits, expertise, communication style |
| **Skill** | Reusable capabilities | Triggers, procedures, execution strategies |
| **Template** | Content generation | Variable substitution, dynamic rendering |
| **Agent** | Autonomous workflows | Goals, planning, execution |
| **Memory** | Context persistence | Content storage, deduplication, search, multilingual keyword extraction, 70-85% token savings |
| **Ensemble** | Multi-agent orchestration | Sequential/parallel execution, consensus |

### GitHub Integration

- **OAuth Authentication** - Secure device flow authentication
- **Portfolio Sync** - Push/pull elements to/from GitHub repositories
- **Collection System** - Install, manage, and publish element collections
- **PR Submission** - Submit elements to collections via automated PRs
- **Conflict Resolution** - Smart conflict resolution with multiple strategies

---

## 📚 Documentation

### User Guides

- [Getting Started](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/user-guide/GETTING_STARTED.md) - Installation, first run, Claude Desktop integration
- [Quick Start Tutorial](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/user-guide/QUICK_START.md) - 10 hands-on tutorials (2-5 min each)
- [Troubleshooting](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/user-guide/TROUBLESHOOTING.md) - Common issues, FAQ, error codes

### Element Documentation

- [Persona](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/elements/PERSONA.md) - Behavioral traits and expertise
- [Skill](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/elements/SKILL.md) - Triggers and procedures
- [Template](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/elements/TEMPLATE.md) - Variable substitution
- [Agent](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/elements/AGENT.md) - Goal-oriented workflows
- [Memory](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/elements/MEMORY.md) - Content deduplication
- [Ensemble](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/elements/ENSEMBLE.md) - Multi-agent orchestration

### API Reference

See [Main README](https://github.com/fsvxavier/nexs-mcp#-available-tools) for complete tool reference.

---

## 💡 Examples

### Create a Persona

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

### Sync with GitHub

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

### Install a Collection

```json
{
  "tool": "install_collection",
  "arguments": {
    "uri": "github://fsvxavier/nexs-collections/technical-writing",
    "force": false
  }
}
```

### Create Backup

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

For more examples, see [Examples Directory](https://github.com/fsvxavier/nexs-mcp/tree/main/examples).

---

## 🔧 Troubleshooting

### Binary Not Found

If you get "command not found" after installation:

```bash
# Check npm global bin directory
npm config get prefix

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH="$PATH:$(npm config get prefix)/bin"

# Reload shell
source ~/.bashrc  # or source ~/.zshrc
```

### Platform Not Supported

If your platform is not supported, you can:

1. **Use Docker**: `docker pull fsvxavier/nexs-mcp:latest`
2. **Build from source**: See [build instructions](https://github.com/fsvxavier/nexs-mcp#building)

### Permission Denied

On Linux/macOS, if you get permission errors:

```bash
# Install without sudo (recommended)
npm install -g --unsafe-perm @fsvxavier/nexs-mcp-server

# Or use npx instead
npx @fsvxavier/nexs-mcp-server
```

### Connection Issues with Claude Desktop

If Claude Desktop can't connect:

1. Verify the binary path: `which nexs-mcp`
2. Test the binary: `nexs-mcp --version`
3. Check Claude Desktop logs (see [troubleshooting guide](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/user-guide/TROUBLESHOOTING.md))
4. Ensure correct configuration in `claude_desktop_config.json`

For more troubleshooting, see [Troubleshooting Guide](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/user-guide/TROUBLESHOOTING.md).

---

## 🏗️ Project Information

### About NEXS MCP

NEXS MCP Server is a high-performance Model Context Protocol implementation built in Go with Clean Architecture. It provides enterprise-grade AI element management with comprehensive tooling.

- **GitHub Repository**: https://github.com/fsvxavier/nexs-mcp
- **NPM Package**: https://www.npmjs.com/package/@fsvxavier/nexs-mcp-server
- **Docker Hub**: https://hub.docker.com/r/fsvxavier/nexs-mcp
- **Documentation**: https://github.com/fsvxavier/nexs-mcp/tree/main/docs

### Technology Stack

- **Language**: Go 1.25
- **MCP SDK**: github.com/modelcontextprotocol/go-sdk v1.1.0
- **Architecture**: Clean Architecture with Domain-Driven Design
- **Storage**: File-based (YAML) and in-memory
- **Test Coverage**: 72.2% (Logger 92.1%, Config 100%, Domain 79.2%)

### Version History

See [CHANGELOG](https://github.com/fsvxavier/nexs-mcp/blob/main/CHANGELOG.md) for release history.

---

## 🤝 Contributing

Contributions are welcome! Please visit the [GitHub repository](https://github.com/fsvxavier/nexs-mcp) to:

- Report bugs
- Request features
- Submit pull requests
- Join discussions

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/fsvxavier/nexs-mcp/blob/main/LICENSE) file for details.

---

## 📧 Support

- **Documentation**: https://github.com/fsvxavier/nexs-mcp/tree/main/docs
- **Issues**: https://github.com/fsvxavier/nexs-mcp/issues
- **Discussions**: https://github.com/fsvxavier/nexs-mcp/discussions

---

## 🔗 Alternative Installation Methods

If NPM doesn't work for you, try these alternatives:

### Go Install

```bash
go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@v1.0.0
```

### Homebrew (macOS/Linux)

```bash
brew tap fsvxavier/nexs-mcp
brew install nexs-mcp
```

### Docker

```bash
docker pull fsvxavier/nexs-mcp:latest
docker run -v $(pwd)/data:/app/data fsvxavier/nexs-mcp:latest
```

### Build from Source

```bash
git clone https://github.com/fsvxavier/nexs-mcp.git
cd nexs-mcp
make build
./bin/nexs-mcp
```

---

<div align="center">

**[⬆ Back to Top](#nexs-mcp-server---npm-package)**

Made with ❤️ by the NEXS MCP team

</div>
