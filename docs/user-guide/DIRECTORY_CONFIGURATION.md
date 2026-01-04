# Directory Configuration Guide

## Overview

NEXS-MCP can be configured to store data either globally (user-wide) or per-workspace (project-specific). All directories follow a consistent pattern based on the `NEXS_DATA_DIR` configuration.

## Directory Structure

```
{BASE_DIR}/
├── elements/          # Elements (personas, agents, skills, etc.) - from NEXS_DATA_DIR
├── working_memory/    # Working memory files (session-based)
├── metrics/           # Metrics collection data
├── performance/       # Performance logs and benchmarks
├── token_metrics/     # Token usage tracking
└── hnsw_index.json    # HNSW vector search index
```

## Configuration Patterns

### 1. Global Configuration (User-wide)

**Use Case**: Share data across all projects for the same user.

```bash
# .env
NEXS_DATA_DIR=~/.nexs-mcp/elements
```

**Result**:
- **BaseDir**: `~/.nexs-mcp/`
- **Elements**: `~/.nexs-mcp/elements/`
- **Working Memory**: `~/.nexs-mcp/working_memory/`
- **Metrics**: `~/.nexs-mcp/metrics/`
- **HNSW Index**: `~/.nexs-mcp/hnsw_index.json`

### 2. Workspace Configuration (Project-specific)

**Use Case**: Isolate data per project/workspace.

```bash
# .env
NEXS_DATA_DIR=.nexs-mcp/elements
```

**Result**:
- **BaseDir**: `./.nexs-mcp/`
- **Elements**: `./.nexs-mcp/elements/`
- **Working Memory**: `./.nexs-mcp/working_memory/`
- **Metrics**: `./.nexs-mcp/metrics/`
- **HNSW Index**: `./.nexs-mcp/hnsw_index.json`

### 3. Custom Location

**Use Case**: Store data in a specific location (e.g., shared network drive).

```bash
# .env
NEXS_DATA_DIR=/data/nexs/elements
NEXS_BASE_DIR=/data/nexs
```

**Result**:
- **BaseDir**: `/data/nexs/`
- **Elements**: `/data/nexs/elements/`
- **Working Memory**: `/data/nexs/working_memory/`
- **Metrics**: `/data/nexs/metrics/`

## Environment Variables

### Primary Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXS_DATA_DIR` | `~/.nexs-mcp/elements` | Directory for element storage |
| `NEXS_BASE_DIR` | *derived from DATA_DIR* | Base directory for all data |

### Directory Overrides

If you need fine-grained control, you can override specific directories:

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXS_WORKING_MEMORY_DIR` | `{BASE_DIR}/working_memory` | Working memory persistence |

## Automatic BaseDir Derivation

The `NEXS_BASE_DIR` is automatically derived from `NEXS_DATA_DIR`:

1. **If DATA_DIR ends with `/elements`**: BaseDir = parent directory
   - `~/.nexs-mcp/elements` → BaseDir = `~/.nexs-mcp`
   - `/app/data/elements` → BaseDir = `/app/data`

2. **Otherwise**: BaseDir = parent directory
   - `./data` → BaseDir = `.`
   - `/var/nexs` → BaseDir = `/var`

## Best Practices

### Development
```bash
# Use workspace-specific directories during development
NEXS_DATA_DIR=./.nexs-mcp/elements
```

**Benefits**:
- Each project has isolated data
- Easy to version control .gitignore for `.nexs-mcp/`
- No conflicts between projects

### Production
```bash
# Use global or persistent directories in production
NEXS_DATA_DIR=/app/data/elements
```

**Benefits**:
- Data persists between deployments
- Centralized data management
- Better for containerized environments

### Docker
```yaml
# docker-compose.yml
services:
  nexs-mcp:
    volumes:
      - nexs-data:/app/data
    environment:
      NEXS_DATA_DIR: /app/data/elements

volumes:
  nexs-data:
```

## Migration

### From Global to Workspace

```bash
# Copy global data to workspace
cp -r ~/.nexs-mcp ./.nexs-mcp

# Update .env
NEXS_DATA_DIR=./.nexs-mcp/elements
```

### From Workspace to Global

```bash
# Copy workspace data to global
mkdir -p ~/.nexs-mcp
cp -r ./.nexs-mcp/* ~/.nexs-mcp/

# Update .env
NEXS_DATA_DIR=~/.nexs-mcp/elements
```

## Verification

Check your current configuration:

```bash
# View effective directories
./bin/nexs-mcp --help | grep -A 5 "data-dir"

# Check actual paths in logs
grep -E "data_dir|BaseDir|working_memory" /tmp/nexs-mcp.log
```

## Troubleshooting

### Issue: Data not found after changing DataDir

**Solution**: Verify BaseDir derivation
```bash
# Check current DataDir
echo $NEXS_DATA_DIR

# Verify BaseDir is correct
# If DataDir = ~/.nexs-mcp/elements, BaseDir should be ~/.nexs-mcp
```

### Issue: Working memory not persisting

**Solution**: Check persistence settings
```bash
# Verify persistence enabled
echo $NEXS_WORKING_MEMORY_PERSISTENCE  # should be "true"

# Check directory exists and is writable
ls -ld $(dirname $NEXS_DATA_DIR)/working_memory
```

### Issue: Different directories in different locations

**Solution**: Use NEXS_BASE_DIR explicitly
```bash
# Force all directories to same base
NEXS_BASE_DIR=/app/data
NEXS_DATA_DIR=/app/data/elements
```

## Summary

All NEXS-MCP directories now follow a **unified pattern**:
- ✅ Configurable between global and workspace
- ✅ Consistent BaseDir derivation
- ✅ All auxiliary data (metrics, performance, HNSW, working memory) in same location
- ✅ Easy migration between configurations
- ✅ Docker-friendly with volume mounts

Choose global (`~/.nexs-mcp`) for user-wide data sharing, or workspace (`.nexs-mcp`) for project isolation.
