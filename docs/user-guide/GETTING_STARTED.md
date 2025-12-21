# Getting Started with NEXS-MCP

Welcome to NEXS-MCP! This guide will help you get up and running with the NEXS Model Context Protocol server.

## Table of Contents

- [What is NEXS-MCP?](#what-is-nexs-mcp)
- [Installation](#installation)
- [First Run](#first-run)
- [Claude Desktop Integration](#claude-desktop-integration)
- [Creating Your First Element](#creating-your-first-element)
- [Understanding Element Types](#understanding-element-types)
- [Common Workflows](#common-workflows)
- [Next Steps](#next-steps)

## What is NEXS-MCP?

NEXS-MCP is a high-performance Model Context Protocol (MCP) server that helps you manage AI capabilities through a structured portfolio system. It provides:

- **6 Element Types**: Personas, Skills, Templates, Agents, Memories, and Ensembles
- **55+ MCP Tools**: Complete CRUD operations, GitHub integration, collections, analytics
- **Clean Architecture**: Domain-driven design with high test coverage (72%+)
- **Dual Storage**: File-based (persistent) or in-memory (temporary)
- **Production Ready**: Backup/restore, logging, monitoring, and analytics

## Installation

### Option 1: NPM (Recommended - Cross-platform)

```bash
npm install -g @fsvxavier/nexs-mcp-server
```

📦 **NPM Package:** https://www.npmjs.com/package/@fsvxavier/nexs-mcp-server

Verify installation:
```bash
nexs-mcp --version
```

### Option 2: Go Install

```bash
go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@v1.0.5
```

### Option 3: Homebrew (macOS/Linux)

```bash
brew tap fsvxavier/nexs-mcp
brew install nexs-mcp
```

### Option 4: Docker

```bash
# Pull from Docker Hub
docker pull fsvxavier/nexs-mcp:latest

# Or specific version
docker pull fsvxavier/nexs-mcp:v1.0.5
```

🐳 **Docker Hub:** https://hub.docker.com/r/fsvxavier/nexs-mcp

### Option 5: Build from Source

```bash
git clone https://github.com/fsvxavier/nexs-mcp.git
cd nexs-mcp
make build
./bin/nexs-mcp
```

## First Run

### Standalone Mode

Run the server directly to test it:

```bash
# Default: file storage in ./data/elements
nexs-mcp

# Custom data directory
nexs-mcp -data-dir ~/my-portfolio

# In-memory storage (data not persisted)
nexs-mcp -storage memory

# Debug logging
nexs-mcp -log-level debug
```

**Output:**
```
NEXS MCP Server v1.0.0
Initializing Model Context Protocol server...
Storage type: file
Data directory: data/elements
Registered 55 tools
Server ready. Listening on stdio...
```

### Understanding Storage Modes

**File Storage (Default):**
- Elements persist in YAML files
- Organized by type: `data/elements/{agents,personas,skills,templates,memories,ensembles}`
- Survives server restarts
- Recommended for production

**In-Memory Storage:**
- Elements stored in RAM only
- Fast for testing
- Lost on server restart
- Good for CI/CD pipelines

## Claude Desktop Integration

NEXS-MCP is designed to work with Claude Desktop through the Model Context Protocol.

### Step 1: Locate Configuration File

**macOS:**
```bash
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Linux:**
```bash
~/.config/Claude/claude_desktop_config.json
```

**Windows:**
```
%APPDATA%\Claude\claude_desktop_config.json
```

### Step 2: Add NEXS-MCP Configuration

Create or edit `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "nexs-mcp",
      "args": []
    }
  }
}
```

**With custom data directory:**
```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "nexs-mcp",
      "args": ["-data-dir", "/path/to/your/portfolio"]
    }
  }
}
```

**With Docker:**
```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-v",
        "${HOME}/nexs-data:/app/data",
        "fsvxavier/nexs-mcp:latest"
      ]
    }
  }
}
```

### Step 3: Restart Claude Desktop

1. Quit Claude Desktop completely
2. Relaunch Claude Desktop
3. Look for "NEXS-MCP" in the available MCP servers

### Step 4: Verify Connection

In Claude, try:
```
Can you list all my elements?
```

Claude should respond using the `list_elements` tool showing an empty portfolio or your existing elements.

## Creating Your First Element

Let's create a simple Persona using Claude Desktop.

### Example 1: Tech Expert Persona

In Claude, say:
```
Create a persona called "Tech Expert" who is knowledgeable about software development, 
DevOps, and cloud technologies. They should have a professional but friendly communication style.
```

Claude will use the `quick_create_persona` tool to create:

**Result:**
```yaml
id: persona-tech-expert-001
type: persona
name: Tech Expert
description: A software development expert specializing in DevOps and cloud technologies
traits:
  - analytical
  - helpful
  - knowledgeable
  - professional
expertise_areas:
  - Software Development
  - DevOps
  - Cloud Technologies
communication_style: Professional yet friendly
```

### Example 2: Code Review Skill

```
Create a skill for reviewing code that checks for best practices, security issues, 
and performance problems.
```

Claude will create a Skill element with appropriate triggers and procedures.

### Example 3: Summary Template

```
Create a template for generating executive summaries with variables for title, 
key points, and recommendations.
```

## Understanding Element Types

NEXS-MCP supports 6 element types, each with specific purposes:

### 1. **Persona**
**Purpose:** Define behavioral characteristics and expertise  
**Use Cases:**
- Role-based AI assistants (e.g., "Senior Developer", "Business Analyst")
- Specialized domain experts
- Multi-personality chatbots

**Key Fields:**
- `traits`: Behavioral characteristics
- `expertise_areas`: Knowledge domains
- `communication_style`: How they express themselves

**Example:**
```yaml
type: persona
name: Senior DevOps Engineer
traits: [analytical, proactive, security-conscious]
expertise_areas: [Kubernetes, CI/CD, Infrastructure as Code]
```

### 2. **Skill**
**Purpose:** Reusable capabilities and procedures  
**Use Cases:**
- Code review processes
- Data analysis workflows
- Troubleshooting procedures

**Key Fields:**
- `category`: Skill classification
- `triggers`: When to activate
- `procedures`: Step-by-step actions
- `dependencies`: Required tools or skills

**Example:**
```yaml
type: skill
name: API Security Audit
category: security
triggers: [api_deployment, security_review]
procedures:
  - Check authentication mechanisms
  - Validate input sanitization
  - Review rate limiting
```

### 3. **Template**
**Purpose:** Reusable content patterns with variable substitution  
**Use Cases:**
- Document generation
- Email templates
- Report formats

**Key Fields:**
- `template_type`: Format (handlebars, go_template)
- `variables`: Required parameters
- `content`: Template body with {{variables}}

**Example:**
```yaml
type: template
name: Bug Report
template_type: handlebars
variables:
  - name: bug_id
  - name: severity
  - name: description
content: |
  Bug ID: {{bug_id}}
  Severity: {{severity}}
  Description: {{description}}
```

### 4. **Agent**
**Purpose:** Autonomous task execution with goals and actions  
**Use Cases:**
- Automated monitoring
- CI/CD workflows
- Data processing pipelines

**Key Fields:**
- `goals`: What the agent aims to achieve
- `actions`: Available capabilities
- `max_iterations`: Execution limit

**Example:**
```yaml
type: agent
name: Deploy Monitor
goals: [Monitor deployment health, Alert on failures]
actions:
  - check_pod_status
  - verify_endpoints
  - send_notifications
max_iterations: 10
```

### 5. **Memory**
**Purpose:** Store context and conversation history  
**Use Cases:**
- Conversation continuity
- Context preservation
- Knowledge retention

**Key Fields:**
- `memory_type`: episodic, semantic, procedural
- `content`: Stored information
- `timestamp`: When created
- `content_hash`: SHA-256 for deduplication

**Example:**
```yaml
type: memory
name: Last Deployment Context
memory_type: episodic
content: Deployed v2.1.0 to production with 3 new features
timestamp: 2025-12-20T10:30:00Z
```

### 6. **Ensemble**
**Purpose:** Orchestrate multiple agents for complex tasks  
**Use Cases:**
- Multi-agent workflows
- Consensus decision making
- Parallel task processing

**Key Fields:**
- `members`: Agent IDs in the ensemble
- `execution_mode`: sequential, parallel, hybrid
- `aggregation_strategy`: first, last, consensus, voting, merge

**Example:**
```yaml
type: ensemble
name: Code Review Team
members: [security-expert, performance-analyst, style-checker]
execution_mode: parallel
aggregation_strategy: consensus
```

## Common Workflows

### Workflow 1: Create and Activate a Persona

```
# In Claude
1. "Create a persona called 'Data Scientist' who specializes in ML and statistics"
2. "Activate the Data Scientist persona"
3. "As Data Scientist, analyze this dataset: [your data]"
```

### Workflow 2: Use a Template

```
# In Claude
1. "Create a template for daily standup notes with fields: tasks_completed, blockers, next_steps"
2. "Instantiate the standup template with:
    - tasks_completed: Fixed bug #123
    - blockers: Waiting for API access
    - next_steps: Implement feature X"
```

### Workflow 3: Search Your Portfolio

```
# In Claude
1. "Search for all personas related to development"
2. "Find skills about security"
3. "List all my active elements"
```

### Workflow 4: Backup and Restore

```
# In Claude
1. "Create a backup of my portfolio"
   → Creates nexs-portfolio-backup-20251220.tar.gz

2. If needed later:
   "Restore my portfolio from backup nexs-portfolio-backup-20251220.tar.gz"
```

### Workflow 5: GitHub Integration

```
# In Claude
1. "Start GitHub authentication"
   → Follow the device flow instructions

2. "Sync my portfolio to GitHub repository my-portfolio"
   → Pushes elements to GitHub

3. Later: "Pull latest changes from GitHub"
   → Syncs remote changes locally
```

## Configuration Options

### Command-Line Flags

```bash
# Storage configuration
-storage string         Storage type: "file" or "memory" (default "file")
-data-dir string        Data directory path (default "data/elements")

# Logging
-log-level string       Log level: debug, info, warn, error (default "info")
-log-format string      Log format: json, text (default "text")

# Server
-server-name string     MCP server name (default "nexs-mcp")

# Resources
-resources-enabled      Enable MCP Resources Protocol (default false)
```

### Environment Variables

```bash
# Override command-line flags
export NEXS_STORAGE_TYPE=file
export NEXS_DATA_DIR=/path/to/data
export NEXS_LOG_LEVEL=debug
export NEXS_LOG_FORMAT=json
export NEXS_RESOURCES_ENABLED=true

# GitHub integration
export GITHUB_TOKEN=ghp_your_token_here
export GITHUB_OWNER=your_username
export GITHUB_REPO=your_portfolio_repo
```

## File Structure

When using file storage, NEXS-MCP creates this structure:

```
data/
└── elements/
    ├── agents/
    │   └── agent-deploy-monitor-001.yaml
    ├── personas/
    │   ├── persona-tech-expert-001.yaml
    │   └── persona-data-scientist-002.yaml
    ├── skills/
    │   └── skill-code-review-001.yaml
    ├── templates/
    │   └── template-bug-report-001.yaml
    ├── memories/
    │   └── memory-deployment-context-001.yaml
    └── ensembles/
        └── ensemble-review-team-001.yaml
```

Each element is stored as a YAML file with:
- Human-readable format
- Git-friendly (easy to diff)
- Can be edited manually
- Validated on load

## Next Steps

### Learning More

1. **[Quick Start Guide](./QUICK_START.md)** - 5-minute tutorials
2. **[Element Types Deep Dive](../elements/README.md)** - Detailed element documentation
3. **[MCP Tools Reference](../api/MCP_TOOLS.md)** - All 55 tools explained
4. **[Troubleshooting Guide](./TROUBLESHOOTING.md)** - Common issues and solutions

### Advanced Features

1. **Collections**: Install pre-built element packages
   ```
   "Browse available collections"
   "Install collection from github://fsvxavier/nexs-collection-starter"
   ```

2. **Ensembles**: Orchestrate multiple agents
   ```
   "Create an ensemble with security-expert and performance-analyst"
   "Execute the ensemble on this code: [code]"
   ```

3. **Analytics**: Monitor your usage
   ```
   "Show usage statistics for the last 7 days"
   "Display performance dashboard"
   ```

4. **Templates**: Advanced templating with Handlebars
   ```
   "Create a template with conditionals and loops"
   "List all available template helpers"
   ```

### Community and Support

- **Documentation**: https://github.com/fsvxavier/nexs-mcp/tree/main/docs
- **Issues**: https://github.com/fsvxavier/nexs-mcp/issues
- **Discussions**: https://github.com/fsvxavier/nexs-mcp/discussions

### Tips for Success

1. **Start Simple**: Begin with Personas and Skills before moving to Agents and Ensembles
2. **Use Quick Create**: The `quick_create_*` tools need minimal input
3. **Leverage Search**: Use semantic search to find relevant elements
4. **Backup Regularly**: Create backups before major changes
5. **Use Tags**: Tag elements for easy filtering and discovery

## Getting Help

If you encounter issues:

1. Check the [Troubleshooting Guide](./TROUBLESHOOTING.md)
2. Enable debug logging: `nexs-mcp -log-level debug`
3. Review logs: Claude Desktop > View > Developer > Developer Tools
4. Ask in [GitHub Discussions](https://github.com/fsvxavier/nexs-mcp/discussions)
5. Report bugs in [GitHub Issues](https://github.com/fsvxavier/nexs-mcp/issues)

---

**Ready to dive deeper?** Check out the [Quick Start Guide](./QUICK_START.md) for hands-on tutorials!
