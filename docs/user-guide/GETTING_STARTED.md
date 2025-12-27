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

- **� Token Optimization**: 81-95% token reduction through 8 integrated services (v1.3.0)
- **⚡ Performance**: 10x throughput with batch processing, streaming, and adaptive caching (v1.3.0)
- **🌍 Multilingual Support**: 11 languages (EN, PT, ES, FR, DE, IT, RU, JA, ZH, AR, HI) with automatic detection
- **6 Element Types**: Personas, Skills, Templates, Agents, Memories, and Ensembles
- **96 MCP Tools**: Complete CRUD, GitHub, collections, working memory, token optimization, quality management
- **⏱️ Task Scheduler**: Cron-like scheduling with priorities and dependencies (v1.2.0)
- **🕰️ Time Travel**: Version history and confidence decay for temporal analysis (v1.2.0)
- **Clean Architecture**: Domain-driven design with high test coverage (70%+ in new modules)
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
go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@v1.3.0
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
docker pull fsvxavier/nexs-mcp:v1.3.0
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
NEXS MCP Server v1.3.0
Initializing Model Context Protocol server...
Storage type: file
Data directory: data/elements
Registered 96 tools
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

## Token Optimization (v1.3.0) ⭐

NEXS-MCP v1.3.0 introduces a comprehensive **Token Optimization System** that achieves **81-95% token reduction** across all operations through 8 integrated services. This dramatically reduces AI API costs and improves response times.

### Quick Start with Optimization

All optimization features are **enabled by default** - just use NEXS-MCP normally and benefit from automatic token savings!

```bash
# Run with default optimizations (recommended)
nexs-mcp

# Check optimization statistics
# In Claude:
"Get optimization stats"
```

**Expected output:**
```json
{
  "total_requests": 1250,
  "tokens_original": 150000,
  "tokens_final": 22500,
  "reduction_percent": 85.0,
  "cache_hit_rate": 0.92,
  "deduplication_rate": 0.38,
  "services_active": 8
}
```

### 8 Optimization Services

#### 1. **Prompt Compression** (35% reduction)
Automatically removes redundant words and simplifies syntax in prompts.

```yaml
# In Claude:
"Create a memory about machine learning best practices for neural networks"

# Before: 15 tokens
# After: 10 tokens (33% reduction)
```

**Configuration:**
```bash
# Disable if needed (not recommended)
nexs-mcp -optimize-prompt=false

# Adjust compression ratio (default: 0.65 = 35% reduction)
nexs-mcp -prompt-compression-ratio=0.70
```

#### 2. **Streaming Handler** (Chunked Delivery)
Delivers large responses progressively for better UX.

```yaml
# In Claude:
"Search for all my memories"

# Automatically streams if response > 10KB
# Shows results as they arrive, not all at once
```

#### 3. **Semantic Deduplication** (92%+ similarity)
Prevents storing duplicate memories/elements based on semantic similarity.

```yaml
# In Claude:
"Create a memory: Machine learning is useful for predictions"
"Create a memory: ML is great for making predictions"

# Second attempt blocked:
Error: Duplicate memory found (94% similar): mem_12345
```

**Configuration:**
```bash
# Adjust similarity threshold (default: 0.92 = 92%)
nexs-mcp -dedup-threshold=0.95
```

#### 4. **Automatic Summarization** (70% compression)
Summarizes large content automatically using TF-IDF algorithm.

```yaml
# In Claude:
"Get memory mem_12345"

# If memory content > 500 chars, auto-summarizes to 150 chars
# Preserves key information using extractive summarization
```

#### 5. **Context Window Manager** (Smart Truncation)
Manages token limits intelligently with LRU and priority strategies.

```yaml
# In Claude:
"Search memories with full context"

# Automatically truncates to 8192 tokens
# Keeps most recent + high-priority items
```

**Configuration:**
```bash
# Adjust max context tokens (default: 8192)
nexs-mcp -max-context-tokens=16384

# Change truncation strategy (default: hybrid)
nexs-mcp -truncation-mode=priority
```

#### 6. **Adaptive Cache** (L1/L2 with 1h-7d TTL)
Two-tier caching with adaptive TTL based on access patterns.

```yaml
# In Claude:
"Get element persona_dev_senior"

# First call: 50ms (database)
# Second call: <1ms (L1 cache)
# Cache TTL adapts: frequently accessed = longer TTL (up to 7 days)
```

**Performance:**
- L1 (in-memory) hit rate: 85-90%
- L2 (Redis) hit rate: 10-12%
- Combined: 95-98% hit rate

#### 7. **Batch Processing** (10x throughput)
Processes multiple operations in single transaction.

```yaml
# In Claude:
"Batch create 50 memories from this list"

# Before: 50 separate DB calls + 50 index rebuilds
# After: 1 transaction + 1 index rebuild
# Result: 10x faster
```

#### 8. **Response Compression** (70-75% reduction)
Compresses responses using gzip encoding.

```yaml
# Automatic for all responses
# Before: 100KB response
# After: 25KB compressed (75% reduction)
```

### Configuration File

Create `config.yaml` for fine-tuned control:

```yaml
# Token Optimization Configuration
optimization:
  enabled: true  # Master switch
  
  # Prompt Compression
  prompt_compression:
    enabled: true
    min_length: 100
    max_ratio: 0.65
    remove_redundancies: true
  
  # Streaming
  streaming:
    enabled: true
    chunk_size: 4096
    threshold: 10240  # 10KB
  
  # Semantic Deduplication
  semantic_dedup:
    enabled: true
    similarity_threshold: 0.92
    embedding_model: "distiluse-base-multilingual-cased-v2"
  
  # Summarization
  summarization:
    enabled: true
    algorithm: "tf-idf"
    compression_ratio: 0.3
    min_content_length: 500
  
  # Context Window
  context_window:
    enabled: true
    max_tokens: 8192
    truncation_mode: "hybrid"
    preserve_recent: 5
  
  # Adaptive Cache
  adaptive_cache:
    enabled: true
    l1_max_entries: 1000
    l2_redis_addr: "localhost:6379"
    default_ttl: "1h"
    max_ttl: "168h"  # 7 days
  
  # Batch Processing
  batch_processing:
    enabled: true
    max_batch_size: 100
    concurrent_workers: 10
  
  # Response Compression
  response_compression:
    enabled: true
    algorithm: "gzip"
    level: 6
    min_size: 1024
```

### Monitoring Optimization

```yaml
# In Claude:
"Get optimization stats"

# Returns:
{
  "total_requests": 5000,
  "tokens_original": 500000,
  "tokens_final": 75000,
  "reduction_percent": 85.0,
  "cache_hit_rate": 0.94,
  "deduplication_rate": 0.35,
  "avg_compression_time_ms": 5,
  "services_active": 8,
  "service_metrics": {
    "prompt_compression": {
      "requests": 5000,
      "avg_reduction": 0.35,
      "avg_time_ms": 3
    },
    "cache": {
      "hits": 4700,
      "misses": 300,
      "hit_rate": 0.94
    },
    "deduplication": {
      "duplicates_found": 175,
      "rate": 0.35
    }
  }
}
```

### Performance Comparison

| Operation | Without Optimization | With Optimization | Reduction |
|---|---|---|---|
| Simple query | 150 tokens | 45 tokens | 70% |
| Memory search | 2,500 tokens | 380 tokens | 85% |
| Batch create (50 items) | 15,000 tokens | 750 tokens | 95% |
| Large content retrieval | 8,000 tokens | 1,200 tokens | 85% |
| GitHub analysis | 12,000 tokens | 1,800 tokens | 85% |

### Cost Savings Example

**Without optimization:**
- 1,000 requests/day
- Avg 500 tokens/request
- Total: 500,000 tokens/day
- Cost: $10/day (at $0.02/1K tokens)
- Monthly: **$300**

**With optimization (85% reduction):**
- Same 1,000 requests/day
- Avg 75 tokens/request
- Total: 75,000 tokens/day
- Cost: $1.50/day
- Monthly: **$45**

💰 **Savings: $255/month (85% cost reduction)**

### Troubleshooting

**Q: Optimization stats show 0% reduction**

```bash
# Check if optimization is enabled
nexs-mcp -log-level=debug

# Look for:
"Optimization services initialized: 8"
```

**Q: Deduplication blocking valid elements**

```bash
# Lower similarity threshold
nexs-mcp -dedup-threshold=0.95  # Default: 0.92

# Or disable temporarily
nexs-mcp -optimize-dedup=false
```

**Q: Cache not working**

```bash
# Check Redis connection (L2 cache)
redis-cli ping

# Or run without Redis (L1 only)
nexs-mcp -cache-l2-enabled=false
```

### Best Practices

1. **Use batch operations** for multiple elements:
   ```yaml
   "Batch create 20 memories from this data"
   ```

2. **Let deduplication work** - it prevents wasted storage and API calls

3. **Monitor optimization stats** regularly:
   ```yaml
   "Get optimization stats"
   ```

4. **Use summarization** for large content:
   ```yaml
   "Summarize this 5000-word memory to 500 words"
   ```

5. **Configure per environment**:
   - Development: Lower thresholds for testing
   - Production: Default settings (optimal balance)

---

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
3. **[MCP Tools Reference](../api/MCP_TOOLS.md)** - All 96 tools explained
4. **[Troubleshooting Guide](./TROUBLESHOOTING.md)** - Common issues and solutions

### Advanced Features

1. **Token Optimization (v1.3.0)**: Reduce costs by 81-95% ⭐
   ```
   "Get optimization stats"
   "Enable/disable specific optimization services"
   "Configure cache TTL and deduplication threshold"
   ```

2. **Collections**: Install pre-built element packages
   ```
   "Browse available collections"
   "Install collection from github://fsvxavier/nexs-collection-starter"
   ```

3. **Ensembles**: Orchestrate multiple agents
   ```
   "Create an ensemble with security-expert and performance-analyst"
   "Execute the ensemble on this code: [code]"
   ```

4. **Working Memory (v1.3.0)**: Context-aware conversations ⭐
   ```
   "Start a conversation about Python optimization"
   "Add this code to working memory"
   "Search conversation history"
   ```

5. **Analytics**: Monitor your usage
   ```
   "Show usage statistics for the last 7 days"
   "Display performance dashboard"
   ```

6. **Templates**: Advanced templating with Handlebars
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
