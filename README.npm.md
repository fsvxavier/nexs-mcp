# NEXS MCP Server

**Model Context Protocol server for managing AI development elements: agents, personas, skills, templates, ensembles, and memories.**

[![npm version](https://img.shields.io/npm/v/@nexs-mcp/server.svg)](https://www.npmjs.com/package/@nexs-mcp/server)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🚀 Quick Start

Install globally:

```bash
npm install -g @nexs-mcp/server
```

Or use with npx:

```bash
npx @nexs-mcp/server --help
```

## 📦 What is NEXS MCP?

NEXS MCP is a powerful server implementing the **Model Context Protocol (MCP)** for managing AI development elements. It provides a structured approach to organizing and reusing AI components like agents, personas, skills, and templates.

### Key Features

- **🤖 Agent Management**: Define and manage AI agents with specific roles and capabilities
- **👤 Persona System**: Create reusable personality profiles for consistent AI behavior
- **🛠 Skills Library**: Organize and share executable skills across agents
- **📝 Template Engine**: Maintain prompt templates with variable interpolation
- **🎭 Ensembles**: Coordinate multiple agents for complex tasks
- **💾 Memory Management**: Persistent storage for agent experiences
- **🔍 GitHub Portfolio Search**: Discover and import community elements
- **✅ Element Validation**: Built-in validation for element structure
- **📊 Analytics**: Statistics and insights on your element library

## 🔧 Installation

### Prerequisites

- Node.js 16+ (for NPM installation)
- OR use as a Claude Desktop integration

### Global Installation

```bash
npm install -g @nexs-mcp/server
nexs-mcp --version
```

### Local Installation

```bash
npm install @nexs-mcp/server
npx nexs-mcp --version
```

### Claude Desktop Integration

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "npx",
      "args": ["@nexs-mcp/server"],
      "env": {
        "NEXS_DATA_DIR": "/path/to/your/nexs/elements"
      }
    }
  }
}
```

## 📚 Usage

### Command Line

```bash
# Show version
nexs-mcp --version

# Show help
nexs-mcp --help

# Start MCP server (stdio mode)
nexs-mcp
```

### MCP Tools

When connected via MCP client (like Claude Desktop), you have access to 30+ tools:

#### Element Management
- `create_element` - Create new elements (agents, personas, skills, etc.)
- `list_elements` - List all elements or filter by type
- `get_element` - Retrieve element details
- `update_element` - Modify existing elements
- `delete_element` - Remove elements
- `search_elements` - Search by name, tags, or content

#### Validation & Templates
- `validate_element` - Validate element structure and content
- `render_template` - Render templates with variable substitution
- `reload_elements` - Reload element library without restart

#### GitHub Portfolio
- `search_portfolio_github` - Discover community elements on GitHub

#### Analytics
- `get_statistics` - View statistics about your element library

#### Backup & Restore
- `backup_elements` - Create backups of your elements
- `restore_backup` - Restore from backup

## 🗂 Directory Structure

NEXS MCP organizes elements in a structured directory:

```
$NEXS_DATA_DIR/
├── agents/          # AI agent definitions
├── personas/        # Personality profiles
├── skills/          # Executable skills
├── templates/       # Prompt templates
├── ensembles/       # Multi-agent coordination
└── memories/        # Persistent agent memory
```

## 📖 Element Examples

### Agent

```yaml
name: code-reviewer
type: agent
description: Reviews code for quality and best practices
persona: senior-developer
skills:
  - code-analysis
  - security-audit
tags:
  - code-review
  - quality
```

### Persona

```yaml
name: senior-developer
type: persona
description: Experienced software engineer with 10+ years
traits:
  - analytical
  - detail-oriented
  - pragmatic
communication_style: technical, direct, concise
```

### Skill

```yaml
name: code-analysis
type: skill
description: Analyzes code structure and quality
parameters:
  - name: language
    type: string
    required: true
  - name: focus
    type: string
    default: "all"
implementation: |
  # Code analysis logic
  analyze_code(language, focus)
```

## 🌐 Environment Variables

- `NEXS_DATA_DIR`: Base directory for NEXS elements (default: `~/.nexs/elements`)
- `NEXS_LOG_LEVEL`: Logging level: `debug`, `info`, `warn`, `error` (default: `info`)
- `NEXS_PORT`: Server port for HTTP mode (default: `8080`)
- `GITHUB_TOKEN`: GitHub token for portfolio search (optional)

## 🛠 Development

### Building from Source

```bash
git clone https://github.com/fsvxavier/nexs-mcp.git
cd nexs-mcp
make build
```

### Running Tests

```bash
make test
```

## 📝 Documentation

- [Full Documentation](https://github.com/fsvxavier/nexs-mcp/tree/main/docs)
- [Element Guide](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/elements/README.md)
- [User Guide](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/USER_GUIDE.md)
- [API Reference](https://github.com/fsvxavier/nexs-mcp/blob/main/docs/API.md)

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](https://github.com/fsvxavier/nexs-mcp/blob/main/CONTRIBUTING.md).

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🔗 Links

- [GitHub Repository](https://github.com/fsvxavier/nexs-mcp)
- [NPM Package](https://www.npmjs.com/package/@nexs-mcp/server)
- [Issue Tracker](https://github.com/fsvxavier/nexs-mcp/issues)
- [Model Context Protocol](https://modelcontextprotocol.io)

## 🙏 Acknowledgments

Built with the [Model Context Protocol](https://modelcontextprotocol.io) by Anthropic.

---

**Made with ❤️ by the NEXS MCP Team**
