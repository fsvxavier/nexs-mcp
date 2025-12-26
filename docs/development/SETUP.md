# Development Setup Guide

**Version:** 1.0.0  
**Last Updated:** December 20, 2025  
**Target Audience:** Contributors and Developers

---

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation Steps](#installation-steps)
- [Building from Source](#building-from-source)
- [Running Locally](#running-locally)
- [IDE Setup](#ide-setup)
- [Development Tools](#development-tools)
- [Configuration](#configuration)
- [Debugging](#debugging)
- [Common Issues](#common-issues)
- [Docker Development](#docker-development)
- [Advanced Setup](#advanced-setup)

---

## Overview

This guide walks you through setting up a complete development environment for NEXS MCP. NEXS MCP is built with Go and uses the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) (`github.com/modelcontextprotocol/go-sdk/mcp v1.1.0`).

**What you'll set up:**

- Go development environment (1.21+)
- NEXS MCP source code
- Development dependencies
- Testing infrastructure
- IDE configuration
- Debug tools

**Time estimate:** 15-30 minutes

---

## Prerequisites

### Required Software

#### 1. Go (1.21 or later)

NEXS MCP requires Go 1.21 or later for generics and modern Go features.

**Installation:**

**macOS:**
```bash
# Using Homebrew
brew install go

# Verify installation
go version  # Should show 1.21 or later
```

**Linux (Ubuntu/Debian):**
```bash
# Remove old Go version if exists
sudo apt remove golang-go
sudo rm -rf /usr/local/go

# Download and install (update version as needed)
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Reload shell configuration
source ~/.bashrc  # or source ~/.zshrc

# Verify installation
go version
```

**Windows:**
```powershell
# Download installer from https://go.dev/dl/
# Run installer and follow prompts

# Verify installation (in PowerShell)
go version
```

#### 2. Git

**macOS:**
```bash
# Usually pre-installed, or use Homebrew
brew install git
```

**Linux:**
```bash
sudo apt update
sudo apt install git
```

**Windows:**
```powershell
# Download from https://git-scm.com/download/win
# Run installer with default options
```

**Verify:**
```bash
git --version
```

#### 3. Make

**macOS:**
```bash
# Pre-installed with Xcode Command Line Tools
xcode-select --install
```

**Linux:**
```bash
sudo apt install build-essential
```

**Windows:**
```powershell
# Install via chocolatey
choco install make

# Or use WSL (Windows Subsystem for Linux)
```

**Verify:**
```bash
make --version
```

### Optional Software

#### 1. golangci-lint (Recommended)

Comprehensive linter for Go code.

```bash
# Install via go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verify installation
golangci-lint --version
```

#### 2. Docker (Optional)

For containerized development and testing.

**macOS:**
```bash
# Download Docker Desktop from https://www.docker.com/products/docker-desktop
# Install and start Docker Desktop
```

**Linux:**
```bash
# Install Docker Engine
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker --version
```

**Windows:**
```powershell
# Download Docker Desktop from https://www.docker.com/products/docker-desktop
# Install and enable WSL 2 backend
```

#### 3. Visual Studio Code or GoLand

See [IDE Setup](#ide-setup) section for configuration.

### System Requirements

- **Operating System:** macOS, Linux, or Windows
- **RAM:** 4GB minimum, 8GB recommended
- **Disk Space:** 2GB free space
- **Network:** Internet connection for downloading dependencies

---

## Installation Steps

### Step 1: Fork and Clone Repository

#### Fork on GitHub

1. Navigate to https://github.com/fsvxavier/nexs-mcp
2. Click **Fork** button (top right)
3. Wait for fork to complete

#### Clone Your Fork

```bash
# Clone repository
git clone https://github.com/YOUR_USERNAME/nexs-mcp.git

# Navigate to directory
cd nexs-mcp

# Verify structure
ls -la
```

Expected output:
```
drwxr-xr-x  cmd/
drwxr-xr-x  internal/
drwxr-xr-x  docs/
drwxr-xr-x  examples/
-rw-r--r--  go.mod
-rw-r--r--  go.sum
-rw-r--r--  Makefile
-rw-r--r--  README.md
...
```

### Step 2: Add Upstream Remote

```bash
# Add upstream remote
git remote add upstream https://github.com/fsvxavier/nexs-mcp.git

# Verify remotes
git remote -v
```

Expected output:
```
origin    https://github.com/YOUR_USERNAME/nexs-mcp.git (fetch)
origin    https://github.com/YOUR_USERNAME/nexs-mcp.git (push)
upstream  https://github.com/fsvxavier/nexs-mcp.git (fetch)
upstream  https://github.com/fsvxavier/nexs-mcp.git (push)
```

### Step 3: Install Go Dependencies

```bash
# Download all dependencies (including MCP Go SDK)
go mod download

# Verify dependencies
go mod verify
```

Expected output:
```
all modules verified
```

**Key dependencies installed:**

- `github.com/modelcontextprotocol/go-sdk/mcp v1.1.0` - Official MCP SDK
- `github.com/stretchr/testify v1.11.1` - Testing framework
- `github.com/google/uuid v1.6.0` - UUID generation
- `gopkg.in/yaml.v3 v3.0.1` - YAML parsing

### Step 4: Verify Installation

```bash
# Run all tests
make test

# Build the binary
make build

# Check binary was created
ls -lh bin/nexs-mcp
```

Expected output:
```
-rwxr-xr-x 1 user user 15M Dec 20 10:00 bin/nexs-mcp
```

If you see errors, check [Common Issues](#common-issues) section.

---

## Building from Source

### Basic Build

```bash
# Build for current platform
make build

# Output: bin/nexs-mcp
```

### Build with Make

NEXS MCP includes a comprehensive Makefile with these targets:

```bash
# Show all available targets
make help

# Build the binary
make build

# Build and run
make run

# Clean build artifacts
make clean

# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Run tests with race detector
make test-race

# Generate coverage report
make test-coverage
```

### Build for Multiple Platforms

```bash
# Build for all supported platforms
make build-all

# Builds created in dist/ directory:
# - nexs-mcp-linux-amd64
# - nexs-mcp-linux-arm64
# - nexs-mcp-darwin-amd64
# - nexs-mcp-darwin-arm64
# - nexs-mcp-windows-amd64.exe
```

### Manual Build with Go

```bash
# Basic build
go build -o bin/nexs-mcp ./cmd/nexs-mcp

# Build with optimizations (smaller binary)
go build -ldflags="-w -s" -o bin/nexs-mcp ./cmd/nexs-mcp

# Build with version info
VERSION=0.1.0
go build -ldflags="-w -s -X main.version=$VERSION" -o bin/nexs-mcp ./cmd/nexs-mcp

# Cross-compile for Linux from macOS
GOOS=linux GOARCH=amd64 go build -o bin/nexs-mcp-linux ./cmd/nexs-mcp

# Cross-compile for Windows from macOS/Linux
GOOS=windows GOARCH=amd64 go build -o bin/nexs-mcp.exe ./cmd/nexs-mcp
```

### Build Flags Explained

- `-ldflags="-w -s"` - Strip debug info (reduces binary size)
- `-X main.version=...` - Set version at compile time
- `-race` - Enable race detector (for testing)
- `-o` - Output file path

### Verify Build

```bash
# Check binary size
ls -lh bin/nexs-mcp

# Run help command
./bin/nexs-mcp --help

# Check version (if implemented)
./bin/nexs-mcp --version
```

---

## Running Locally

### Basic Execution

```bash
# Build and run
make run

# Or run directly
./bin/nexs-mcp
```

The server starts in stdio mode, ready to receive MCP protocol messages.

### Configuration

Create a configuration file at `~/.nexs-mcp/config.yaml`:

```yaml
# NEXS MCP Configuration

# Data directory for storing elements
data_dir: ~/.nexs-mcp/data

# Log level: debug, info, warn, error
log_level: info

# Enable/disable features
features:
  indexing: true
  collections: true
  portfolio: true

# Performance tuning
performance:
  cache_size: 1000
  max_concurrent_operations: 10
  batch_size: 100

# Element directories
elements:
  personas_dir: personas
  skills_dir: skills
  templates_dir: templates
  agents_dir: agents
  memories_dir: memories
  ensembles_dir: ensembles
```

### Environment Variables

NEXS MCP supports environment variable configuration:

```bash
# Set data directory
export NEXS_DATA_DIR=~/.nexs-mcp/data

# Set log level
export NEXS_LOG_LEVEL=debug

# Enable debug mode
export NEXS_DEBUG=true

# Run with environment variables
./bin/nexs-mcp
```

### Running with MCP Clients

#### Claude Desktop

Add to Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "nexs": {
      "command": "/path/to/nexs-mcp/bin/nexs-mcp",
      "args": [],
      "env": {
        "NEXS_DATA_DIR": "/Users/username/.nexs-mcp/data",
        "NEXS_LOG_LEVEL": "info"
      }
    }
  }
}
```

#### Custom MCP Client

```go
package main

import (
    "context"
    "os/exec"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    // Start NEXS MCP server
    cmd := exec.Command("./bin/nexs-mcp")
    
    // Create MCP client
    client := mcp.NewStdioClient(cmd)
    
    // Initialize connection
    ctx := context.Background()
    if err := client.Initialize(ctx); err != nil {
        panic(err)
    }
    
    // Call tools
    result, err := client.CallTool(ctx, "list_personas", nil)
    if err != nil {
        panic(err)
    }
    
    println(result)
}
```

### Development Mode

Run with enhanced logging and debugging:

```bash
# Enable debug logging
NEXS_LOG_LEVEL=debug ./bin/nexs-mcp

# Enable verbose output
NEXS_VERBOSE=true ./bin/nexs-mcp

# Enable profiling
NEXS_PROFILE=true ./bin/nexs-mcp
```

---

## IDE Setup

### Visual Studio Code

#### Extensions

Install these recommended extensions:

1. **Go** (golang.go) - Official Go extension
2. **Go Test Explorer** (premparihar.gotestexplorer) - Test UI
3. **Go Outliner** (766b.go-outliner) - Code outline
4. **MCP Protocol** (if available) - MCP syntax support

```bash
# Install extensions via CLI
code --install-extension golang.go
code --install-extension premparihar.gotestexplorer
```

#### Workspace Settings

Create `.vscode/settings.json`:

```json
{
  // Go Language Settings
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"],
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "gutter"
  },
  "go.testTimeout": "30s",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "files.exclude": {
    "**/.git": true,
    "**/bin": true,
    "**/dist": true,
    "**/*.out": true
  },

  // NEXS MCP Configuration
  // All environment variables are optional and shown with their default values
  "terminal.integrated.env.linux": {
    // Core Server Settings
    "NEXS_SERVER_NAME": "nexs-mcp",
    "NEXS_STORAGE_TYPE": "file",
    "NEXS_DATA_DIR": "data/elements",
    "NEXS_LOG_LEVEL": "info",
    "NEXS_LOG_FORMAT": "json",

    // Auto-Save Configuration
    "NEXS_AUTO_SAVE_MEMORIES": "true",
    "NEXS_AUTO_SAVE_INTERVAL": "5m",

    // Resources Protocol (MCP Resources)
    "NEXS_RESOURCES_ENABLED": "false",
    "NEXS_RESOURCES_CACHE_TTL": "5m",

    // Response Compression
    "NEXS_COMPRESSION_ENABLED": "false",
    "NEXS_COMPRESSION_ALGORITHM": "gzip",
    "NEXS_COMPRESSION_MIN_SIZE": "1024",
    "NEXS_COMPRESSION_LEVEL": "6",
    "NEXS_COMPRESSION_ADAPTIVE": "true",

    // Streaming Responses
    "NEXS_STREAMING_ENABLED": "false",
    "NEXS_STREAMING_CHUNK_SIZE": "10",
    "NEXS_STREAMING_THROTTLE": "50ms",
    "NEXS_STREAMING_BUFFER_SIZE": "100",

    // Automatic Summarization
    "NEXS_SUMMARIZATION_ENABLED": "false",
    "NEXS_SUMMARIZATION_AGE": "168h",
    "NEXS_SUMMARIZATION_MAX_LENGTH": "500",
    "NEXS_SUMMARIZATION_RATIO": "0.3",
    "NEXS_SUMMARIZATION_PRESERVE_KEYWORDS": "true",
    "NEXS_SUMMARIZATION_EXTRACTIVE": "true",

    // Adaptive Cache TTL
    "NEXS_ADAPTIVE_CACHE_ENABLED": "false",
    "NEXS_ADAPTIVE_CACHE_MIN_TTL": "1h",
    "NEXS_ADAPTIVE_CACHE_MAX_TTL": "168h",
    "NEXS_ADAPTIVE_CACHE_BASE_TTL": "24h",

    // Prompt Compression
    "NEXS_PROMPT_COMPRESSION_ENABLED": "false",
    "NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY": "true",
    "NEXS_PROMPT_COMPRESSION_WHITESPACE": "true",
    "NEXS_PROMPT_COMPRESSION_ALIASES": "true",
    "NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE": "true",
    "NEXS_PROMPT_COMPRESSION_RATIO": "0.65",
    "NEXS_PROMPT_COMPRESSION_MIN_LENGTH": "500"
  },
  "terminal.integrated.env.osx": {
    // Same as linux - macOS configuration
    "NEXS_SERVER_NAME": "nexs-mcp",
    "NEXS_STORAGE_TYPE": "file",
    "NEXS_DATA_DIR": "data/elements",
    "NEXS_LOG_LEVEL": "info",
    "NEXS_LOG_FORMAT": "json",
    "NEXS_AUTO_SAVE_MEMORIES": "true",
    "NEXS_AUTO_SAVE_INTERVAL": "5m",
    "NEXS_RESOURCES_ENABLED": "false",
    "NEXS_RESOURCES_CACHE_TTL": "5m",
    "NEXS_COMPRESSION_ENABLED": "false",
    "NEXS_COMPRESSION_ALGORITHM": "gzip",
    "NEXS_COMPRESSION_MIN_SIZE": "1024",
    "NEXS_COMPRESSION_LEVEL": "6",
    "NEXS_COMPRESSION_ADAPTIVE": "true",
    "NEXS_STREAMING_ENABLED": "false",
    "NEXS_STREAMING_CHUNK_SIZE": "10",
    "NEXS_STREAMING_THROTTLE": "50ms",
    "NEXS_STREAMING_BUFFER_SIZE": "100",
    "NEXS_SUMMARIZATION_ENABLED": "false",
    "NEXS_SUMMARIZATION_AGE": "168h",
    "NEXS_SUMMARIZATION_MAX_LENGTH": "500",
    "NEXS_SUMMARIZATION_RATIO": "0.3",
    "NEXS_SUMMARIZATION_PRESERVE_KEYWORDS": "true",
    "NEXS_SUMMARIZATION_EXTRACTIVE": "true",
    "NEXS_ADAPTIVE_CACHE_ENABLED": "false",
    "NEXS_ADAPTIVE_CACHE_MIN_TTL": "1h",
    "NEXS_ADAPTIVE_CACHE_MAX_TTL": "168h",
    "NEXS_ADAPTIVE_CACHE_BASE_TTL": "24h",
    "NEXS_PROMPT_COMPRESSION_ENABLED": "false",
    "NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY": "true",
    "NEXS_PROMPT_COMPRESSION_WHITESPACE": "true",
    "NEXS_PROMPT_COMPRESSION_ALIASES": "true",
    "NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE": "true",
    "NEXS_PROMPT_COMPRESSION_RATIO": "0.65",
    "NEXS_PROMPT_COMPRESSION_MIN_LENGTH": "500"
  },
  "terminal.integrated.env.windows": {
    // Same as linux/osx - Windows configuration
    "NEXS_SERVER_NAME": "nexs-mcp",
    "NEXS_STORAGE_TYPE": "file",
    "NEXS_DATA_DIR": "data/elements",
    "NEXS_LOG_LEVEL": "info",
    "NEXS_LOG_FORMAT": "json",
    "NEXS_AUTO_SAVE_MEMORIES": "true",
    "NEXS_AUTO_SAVE_INTERVAL": "5m",
    "NEXS_RESOURCES_ENABLED": "false",
    "NEXS_RESOURCES_CACHE_TTL": "5m",
    "NEXS_COMPRESSION_ENABLED": "false",
    "NEXS_COMPRESSION_ALGORITHM": "gzip",
    "NEXS_COMPRESSION_MIN_SIZE": "1024",
    "NEXS_COMPRESSION_LEVEL": "6",
    "NEXS_COMPRESSION_ADAPTIVE": "true",
    "NEXS_STREAMING_ENABLED": "false",
    "NEXS_STREAMING_CHUNK_SIZE": "10",
    "NEXS_STREAMING_THROTTLE": "50ms",
    "NEXS_STREAMING_BUFFER_SIZE": "100",
    "NEXS_SUMMARIZATION_ENABLED": "false",
    "NEXS_SUMMARIZATION_AGE": "168h",
    "NEXS_SUMMARIZATION_MAX_LENGTH": "500",
    "NEXS_SUMMARIZATION_RATIO": "0.3",
    "NEXS_SUMMARIZATION_PRESERVE_KEYWORDS": "true",
    "NEXS_SUMMARIZATION_EXTRACTIVE": "true",
    "NEXS_ADAPTIVE_CACHE_ENABLED": "false",
    "NEXS_ADAPTIVE_CACHE_MIN_TTL": "1h",
    "NEXS_ADAPTIVE_CACHE_MAX_TTL": "168h",
    "NEXS_ADAPTIVE_CACHE_BASE_TTL": "24h",
    "NEXS_PROMPT_COMPRESSION_ENABLED": "false",
    "NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY": "true",
    "NEXS_PROMPT_COMPRESSION_WHITESPACE": "true",
    "NEXS_PROMPT_COMPRESSION_ALIASES": "true",
    "NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE": "true",
    "NEXS_PROMPT_COMPRESSION_RATIO": "0.65",
    "NEXS_PROMPT_COMPRESSION_MIN_LENGTH": "500"
  }
}
```

#### Production-Ready Settings Example

For production deployments with all features enabled:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "terminal.integrated.env.linux": {
    // Core Configuration
    "NEXS_SERVER_NAME": "nexs-mcp-prod",
    "NEXS_STORAGE_TYPE": "file",
    "NEXS_DATA_DIR": "/var/lib/nexs-mcp/data",
    "NEXS_LOG_LEVEL": "warn",
    "NEXS_LOG_FORMAT": "json",

    // Auto-Save Enabled (recommended)
    "NEXS_AUTO_SAVE_MEMORIES": "true",
    "NEXS_AUTO_SAVE_INTERVAL": "3m",

    // Resources Protocol Enabled
    "NEXS_RESOURCES_ENABLED": "true",
    "NEXS_RESOURCES_CACHE_TTL": "10m",

    // Compression Enabled (reduces bandwidth)
    "NEXS_COMPRESSION_ENABLED": "true",
    "NEXS_COMPRESSION_ALGORITHM": "gzip",
    "NEXS_COMPRESSION_MIN_SIZE": "512",
    "NEXS_COMPRESSION_LEVEL": "9",
    "NEXS_COMPRESSION_ADAPTIVE": "true",

    // Streaming Enabled (better UX for large responses)
    "NEXS_STREAMING_ENABLED": "true",
    "NEXS_STREAMING_CHUNK_SIZE": "20",
    "NEXS_STREAMING_THROTTLE": "25ms",
    "NEXS_STREAMING_BUFFER_SIZE": "200",

    // Summarization Enabled (memory optimization)
    "NEXS_SUMMARIZATION_ENABLED": "true",
    "NEXS_SUMMARIZATION_AGE": "72h",
    "NEXS_SUMMARIZATION_MAX_LENGTH": "300",
    "NEXS_SUMMARIZATION_RATIO": "0.25",
    "NEXS_SUMMARIZATION_PRESERVE_KEYWORDS": "true",
    "NEXS_SUMMARIZATION_EXTRACTIVE": "true",

    // Adaptive Cache Enabled (performance optimization)
    "NEXS_ADAPTIVE_CACHE_ENABLED": "true",
    "NEXS_ADAPTIVE_CACHE_MIN_TTL": "30m",
    "NEXS_ADAPTIVE_CACHE_MAX_TTL": "336h",
    "NEXS_ADAPTIVE_CACHE_BASE_TTL": "48h",

    // Prompt Compression Enabled (token optimization)
    "NEXS_PROMPT_COMPRESSION_ENABLED": "true",
    "NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY": "true",
    "NEXS_PROMPT_COMPRESSION_WHITESPACE": "true",
    "NEXS_PROMPT_COMPRESSION_ALIASES": "true",
    "NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE": "true",
    "NEXS_PROMPT_COMPRESSION_RATIO": "0.60",
    "NEXS_PROMPT_COMPRESSION_MIN_LENGTH": "300"
  }
}
```

#### ONNX Configuration (Future: Vector Embeddings)

**Note:** ONNX support for local embeddings is planned for Sprint 5-6 (see [COMPETITIVE_ANALYSIS](../analysis/COMPETITIVE_ANALYSIS_MEMORY_MCP.md)).

When ONNX support is implemented, configuration will include:

```json
{
  "terminal.integrated.env.linux": {
    // ... existing NEXS configuration ...

    // ONNX Runtime Configuration (Coming in Sprint 5)
    "NEXS_ONNX_ENABLED": "true",
    "NEXS_ONNX_MODEL_PATH": "./models/all-MiniLM-L6-v2.onnx",
    "NEXS_ONNX_TOKENIZER_PATH": "./models/tokenizer.json",
    "NEXS_ONNX_EMBEDDING_DIM": "384",
    "NEXS_ONNX_MAX_SEQ_LENGTH": "256",
    "NEXS_ONNX_DEVICE": "cpu",
    "NEXS_ONNX_NUM_THREADS": "4",

    // Vector Search Configuration (Coming in Sprint 5-6)
    "NEXS_VECTOR_ENABLED": "true",
    "NEXS_VECTOR_INDEX_TYPE": "hnsw",
    "NEXS_VECTOR_M": "16",
    "NEXS_VECTOR_EF_CONSTRUCTION": "200",
    "NEXS_VECTOR_EF_SEARCH": "50",
    "NEXS_VECTOR_DISTANCE_METRIC": "cosine",

    // Hybrid Search Configuration (Coming in Sprint 7)
    "NEXS_HYBRID_SEARCH_ENABLED": "true",
    "NEXS_HYBRID_ALPHA": "0.7",
    "NEXS_HYBRID_MIN_SCORE": "0.5"
  }
}
```

To prepare for ONNX/vector search:

1. **Download ONNX Model** (when feature is available):
   ```bash
   mkdir -p models
   # Example: all-MiniLM-L6-v2 model (384 dimensions)
   wget https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/onnx/model.onnx \
     -O models/all-MiniLM-L6-v2.onnx
   wget https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/tokenizer.json \
     -O models/tokenizer.json
   ```

2. **Install ONNX Runtime** (Go bindings):
   ```bash
   # Will be added to go.mod in Sprint 5
   go get github.com/yalue/onnxruntime_go
   ```

3. **Performance Recommendations**:
   - **CPU**: Use 2-4 threads, smaller models (all-MiniLM-L6-v2)
   - **GPU**: Configure `NEXS_ONNX_DEVICE=cuda` with CUDA provider
   - **Memory**: ~500MB RAM per model loaded

#### Configuration Priority

NEXS MCP loads configuration in this order (later sources override earlier):

1. **Default values** (hardcoded in config.go)
2. **Environment variables** (from shell or settings.json)
3. **Command-line flags** (when launching manually)

Example using flags:

```bash
./nexs-mcp \
  --storage=file \
  --data-dir=/custom/path \
  --log-level=debug \
  --resources-enabled=true \
  --compression-enabled=true \
  --streaming-enabled=true
```

#### Launch Configuration

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch NEXS MCP",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/nexs-mcp",
      "env": {
        "NEXS_LOG_LEVEL": "debug",
        "NEXS_DATA_DIR": "${workspaceFolder}/data"
      },
      "args": []
    },
    {
      "name": "Test Current File",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${file}"
    },
    {
      "name": "Test Current Package",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}"
    }
  ]
}
```

#### Tasks

Create `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build",
      "type": "shell",
      "command": "make build",
      "group": {
        "kind": "build",
        "isDefault": true
      }
    },
    {
      "label": "Test",
      "type": "shell",
      "command": "make test",
      "group": {
        "kind": "test",
        "isDefault": true
      }
    },
    {
      "label": "Lint",
      "type": "shell",
      "command": "make lint"
    }
  ]
}
```

#### Snippets

Create `.vscode/go.code-snippets`:

```json
{
  "MCP Tool Handler": {
    "prefix": "mcptool",
    "body": [
      "func (s *Server) handle${1:ToolName}(ctx context.Context, args map[string]interface{}) (*mcp.ToolResponse, error) {",
      "\t// Validate input",
      "\t${2:paramName}, ok := args[\"${3:param_name}\"].(string)",
      "\tif !ok {",
      "\t\treturn nil, fmt.Errorf(\"${3:param_name} is required\")",
      "\t}",
      "",
      "\t// Implementation",
      "\t${0}",
      "",
      "\t// Return response",
      "\treturn &mcp.ToolResponse{",
      "\t\tContent: []interface{}{",
      "\t\t\tmap[string]interface{}{",
      "\t\t\t\t\"type\": \"text\",",
      "\t\t\t\t\"text\": result,",
      "\t\t\t},",
      "\t\t},",
      "\t}, nil",
      "}"
    ],
    "description": "MCP tool handler function"
  }
}
```

### GoLand / IntelliJ IDEA

#### Configuration

1. **Open project** - File → Open → Select nexs-mcp directory
2. **Configure Go SDK** - File → Settings → Go → GOROOT
3. **Enable Go Modules** - Settings → Go → Go Modules → Enable

#### Run Configurations

**Build and Run:**

1. Run → Edit Configurations
2. Add New Configuration → Go Build
3. Configure:
   - Name: "Build NEXS MCP"
   - Run kind: Package
   - Package path: `github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp`
   - Output directory: `bin/`

**Test Configuration:**

1. Run → Edit Configurations
2. Add New Configuration → Go Test
3. Configure:
   - Name: "All Tests"
   - Test kind: All packages
   - Package path: `./...`

#### Code Style

1. Settings → Editor → Code Style → Go
2. Import scheme → Select "gofmt"
3. Enable "Organize imports on save"

---

## Development Tools

### golangci-lint Configuration

Create `.golangci.yml`:

```yaml
run:
  timeout: 5m
  tests: true
  skip-dirs:
    - vendor
    - _old

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - misspell
    - gocritic
    - gocyclo
    - dupl

linters-settings:
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - dupl
        - gosec
```

### Git Hooks

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

# Pre-commit hook for NEXS MCP

echo "Running pre-commit checks..."

# Format code
make fmt

# Run linter
if ! make lint; then
    echo "Linting failed. Please fix errors before committing."
    exit 1
fi

# Run tests
if ! make test; then
    echo "Tests failed. Please fix before committing."
    exit 1
fi

echo "Pre-commit checks passed!"
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

### Code Generation

NEXS MCP uses Go generate for code generation:

```bash
# Generate mocks
go generate ./...

# Generate specific package
go generate ./internal/domain
```

---

## Configuration

### Data Directory Structure

```
~/.nexs-mcp/
├── config.yaml           # Configuration file
├── data/                 # Element storage
│   ├── personas/
│   ├── skills/
│   ├── templates/
│   ├── agents/
│   ├── memories/
│   └── ensembles/
├── cache/               # Cache files
├── logs/                # Log files
└── backups/             # Backup storage
```

### Creating Data Directory

```bash
# Create directory structure
mkdir -p ~/.nexs-mcp/{data,cache,logs,backups}
mkdir -p ~/.nexs-mcp/data/{personas,skills,templates,agents,memories,ensembles}

# Set permissions
chmod -R 755 ~/.nexs-mcp
```

### Sample Configuration

```yaml
# ~/.nexs-mcp/config.yaml

server:
  name: nexs-mcp
  version: 0.1.0

storage:
  backend: file
  data_dir: ~/.nexs-mcp/data
  backup_enabled: true
  backup_interval: 24h

logging:
  level: info
  format: json
  output: file
  file: ~/.nexs-mcp/logs/nexs-mcp.log

performance:
  cache_enabled: true
  cache_size: 1000
  cache_ttl: 1h
  max_workers: 10

features:
  indexing:
    enabled: true
    backend: memory
  collections:
    enabled: true
    registry_url: https://registry.nexs.dev
  portfolio:
    enabled: true
    auto_optimize: true
```

---

## Debugging

### Debug with VS Code

1. Set breakpoints in code
2. Press F5 or Run → Start Debugging
3. Use Debug Console to inspect variables

### Debug with Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start debugging
dlv debug ./cmd/nexs-mcp

# In Delve console
break main.main
continue
print variableName
```

### Verbose Logging

```bash
# Enable debug logging
export NEXS_LOG_LEVEL=debug
export NEXS_VERBOSE=true
./bin/nexs-mcp 2>&1 | tee debug.log
```

### Profiling

#### CPU Profile

```go
// Add to main.go
import (
    "os"
    "runtime/pprof"
)

func main() {
    // CPU profiling
    if os.Getenv("NEXS_PROFILE") == "true" {
        f, err := os.Create("cpu.prof")
        if err != nil {
            panic(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }
    
    // Rest of main...
}
```

Run with profiling:

```bash
NEXS_PROFILE=true ./bin/nexs-mcp

# Analyze profile
go tool pprof cpu.prof
```

#### Memory Profile

```bash
# Enable memory profiling
export NEXS_MEM_PROFILE=true
./bin/nexs-mcp

# Analyze memory
go tool pprof mem.prof
```

### Trace Analysis

```bash
# Record trace
NEXS_TRACE=true ./bin/nexs-mcp

# Analyze trace
go tool trace trace.out
```

---

## Common Issues

### Issue: "command not found: go"

**Solution:**

Ensure Go is installed and in PATH:

```bash
# Check Go installation
which go

# If not found, add to PATH
export PATH=$PATH:/usr/local/go/bin

# Add to shell config permanently
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Issue: "cannot find package"

**Solution:**

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Verify modules
go mod verify
```

### Issue: "permission denied"

**Solution:**

```bash
# Make binary executable
chmod +x bin/nexs-mcp

# Or rebuild
make clean
make build
```

### Issue: Tests fail with "timeout"

**Solution:**

```bash
# Increase timeout
go test -timeout 60s ./...

# Or use Makefile
make test
```

### Issue: "golangci-lint: command not found"

**Solution:**

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verify installation
which golangci-lint

# If still not found, add GOPATH/bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### Issue: Build fails with MCP SDK errors

**Solution:**

```bash
# Ensure MCP SDK is at correct version
go get github.com/modelcontextprotocol/go-sdk@v1.1.0

# Update go.mod
go mod tidy

# Rebuild
make clean
make build
```

---

## Docker Development

### Build Docker Image

```bash
# Build image
docker build -t nexs-mcp:dev .

# Or use Makefile
make docker-build
```

### Run in Docker

```bash
# Run container
docker run -it --rm \
  -v ~/.nexs-mcp/data:/data \
  -e NEXS_DATA_DIR=/data \
  nexs-mcp:dev

# Or use Makefile
make docker-run
```

### Docker Compose

Create `docker-compose.dev.yml`:

```yaml
version: '3.8'

services:
  nexs-mcp:
    build: .
    volumes:
      - ~/.nexs-mcp/data:/data
      - ./:/app
    environment:
      - NEXS_DATA_DIR=/data
      - NEXS_LOG_LEVEL=debug
    command: /app/bin/nexs-mcp
```

Run with docker-compose:

```bash
docker-compose -f docker-compose.dev.yml up
```

---

## Advanced Setup

### Multiple Go Versions

Use `gvm` (Go Version Manager):

```bash
# Install gvm
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)

# Install Go versions
gvm install go1.21.5
gvm install go1.25.0

# Use specific version
gvm use go1.21.5
```

### Custom GOPATH

```bash
# Set custom GOPATH
export GOPATH=/custom/path/go
export PATH=$PATH:$GOPATH/bin

# Clone project
cd $GOPATH/src/github.com/fsvxavier
git clone https://github.com/fsvxavier/nexs-mcp.git
```

### Development on Remote Machine

Using VS Code Remote SSH:

```bash
# On local machine
code --install-extension ms-vscode-remote.remote-ssh

# Connect to remote
# Cmd+Shift+P → "Remote-SSH: Connect to Host"

# Open folder on remote
# File → Open Folder → /path/to/nexs-mcp
```

### Air (Live Reload)

Install Air for automatic rebuilds:

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Create .air.toml
air init

# Run with live reload
air
```

---

## Next Steps

Now that your environment is set up:

1. **Read the architecture docs** - [Architecture Overview](../architecture/OVERVIEW.md)
2. **Explore the codebase** - Start with `cmd/nexs-mcp/main.go`
3. **Run tests** - `make test-coverage`
4. **Try examples** - See `examples/` directory
5. **Read contributing guide** - [CONTRIBUTING.md](../../CONTRIBUTING.md)
6. **Join discussions** - GitHub Discussions

**Ready to contribute?**

- Check open issues for "good first issue" labels
- Read the [Testing Guide](TESTING.md)
- Review [Coding Standards](../../CONTRIBUTING.md#coding-standards)

---

## Support

If you encounter issues:

1. Check [Common Issues](#common-issues)
2. Search existing GitHub issues
3. Ask in GitHub Discussions
4. Review [Troubleshooting Guide](../user-guide/TROUBLESHOOTING.md)

**Happy coding! 🚀**
