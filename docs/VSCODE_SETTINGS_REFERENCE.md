# VSCode Settings Reference

Complete reference for NEXS MCP configuration in Visual Studio Code and Cursor.

## Table of Contents

- [Quick Start](#quick-start)
- [Core Configuration](#core-configuration)
- [Advanced Features](#advanced-features)
- [Production Settings](#production-settings)
- [Future Features (ONNX/Vector Search)](#future-features-onnxvector-search)
- [Configuration Priority](#configuration-priority)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Minimal Configuration

Create `.vscode/settings.json` in your workspace:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  
  "terminal.integrated.env.linux": {
    "NEXS_STORAGE_TYPE": "file",
    "NEXS_DATA_DIR": "data/elements",
    "NEXS_LOG_LEVEL": "info"
  }
}
```

### Development Configuration

```json
{
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
  
  "terminal.integrated.env.linux": {
    "NEXS_SERVER_NAME": "nexs-mcp-dev",
    "NEXS_STORAGE_TYPE": "file",
    "NEXS_DATA_DIR": "data/elements",
    "NEXS_LOG_LEVEL": "debug",
    "NEXS_LOG_FORMAT": "text"
  }
}
```

## Core Configuration

### Server Settings

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_SERVER_NAME` | `"nexs-mcp"` | Server instance name |
| `NEXS_STORAGE_TYPE` | `"file"` | Storage backend: `"file"` or `"memory"` |
| `NEXS_DATA_DIR` | `"data/elements"` | Directory for file-based storage |
| `NEXS_LOG_LEVEL` | `"info"` | Log level: `"debug"`, `"info"`, `"warn"`, `"error"` |
| `NEXS_LOG_FORMAT` | `"json"` | Log format: `"json"` or `"text"` |

### Auto-Save Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_AUTO_SAVE_MEMORIES` | `true` | Automatically save conversation context as memories |
| `NEXS_AUTO_SAVE_INTERVAL` | `"5m"` | Minimum interval between auto-saves (e.g., `"3m"`, `"10m"`) |

**Example:**

```json
{
  "terminal.integrated.env.linux": {
    "NEXS_AUTO_SAVE_MEMORIES": "true",
    "NEXS_AUTO_SAVE_INTERVAL": "3m"
  }
}
```

## Advanced Features

### Resources Protocol (MCP Resources)

Expose elements as MCP resources for better integration with MCP clients.

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_RESOURCES_ENABLED` | `false` | Enable MCP Resources Protocol |
| `NEXS_RESOURCES_CACHE_TTL` | `"5m"` | Cache TTL for resource content |

**Example:**

```json
{
  "terminal.integrated.env.linux": {
    "NEXS_RESOURCES_ENABLED": "true",
    "NEXS_RESOURCES_CACHE_TTL": "10m"
  }
}
```

### Response Compression

Compress large responses to reduce bandwidth usage.

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_COMPRESSION_ENABLED` | `false` | Enable response compression |
| `NEXS_COMPRESSION_ALGORITHM` | `"gzip"` | Algorithm: `"gzip"` or `"zlib"` |
| `NEXS_COMPRESSION_MIN_SIZE` | `1024` | Minimum response size to compress (bytes) |
| `NEXS_COMPRESSION_LEVEL` | `6` | Compression level (1-9, higher = more compression) |
| `NEXS_COMPRESSION_ADAPTIVE` | `true` | Dynamically adjust compression based on content |

**Example:**

```json
{
  "terminal.integrated.env.linux": {
    "NEXS_COMPRESSION_ENABLED": "true",
    "NEXS_COMPRESSION_ALGORITHM": "gzip",
    "NEXS_COMPRESSION_MIN_SIZE": "512",
    "NEXS_COMPRESSION_LEVEL": "9",
    "NEXS_COMPRESSION_ADAPTIVE": "true"
  }
}
```

### Streaming Responses

Stream large results incrementally for better user experience.

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_STREAMING_ENABLED` | `false` | Enable streaming responses |
| `NEXS_STREAMING_CHUNK_SIZE` | `10` | Number of items per chunk |
| `NEXS_STREAMING_THROTTLE` | `"50ms"` | Delay between chunks |
| `NEXS_STREAMING_BUFFER_SIZE` | `100` | Maximum buffered items |

**Example:**

```json
{
  "terminal.integrated.env.linux": {
    "NEXS_STREAMING_ENABLED": "true",
    "NEXS_STREAMING_CHUNK_SIZE": "20",
    "NEXS_STREAMING_THROTTLE": "25ms",
    "NEXS_STREAMING_BUFFER_SIZE": "200"
  }
}
```

### Automatic Summarization

Automatically summarize old memories to reduce storage and improve performance.

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_SUMMARIZATION_ENABLED` | `false` | Enable automatic summarization |
| `NEXS_SUMMARIZATION_AGE` | `"168h"` | Age before summarizing (e.g., `"72h"` = 3 days) |
| `NEXS_SUMMARIZATION_MAX_LENGTH` | `500` | Maximum summary length (characters) |
| `NEXS_SUMMARIZATION_RATIO` | `0.3` | Target compression ratio (0.3 = 70% reduction) |
| `NEXS_SUMMARIZATION_PRESERVE_KEYWORDS` | `true` | Preserve important keywords in summaries |
| `NEXS_SUMMARIZATION_EXTRACTIVE` | `true` | Use extractive summarization (vs. abstractive) |

**Example:**

```json
{
  "terminal.integrated.env.linux": {
    "NEXS_SUMMARIZATION_ENABLED": "true",
    "NEXS_SUMMARIZATION_AGE": "72h",
    "NEXS_SUMMARIZATION_MAX_LENGTH": "300",
    "NEXS_SUMMARIZATION_RATIO": "0.25",
    "NEXS_SUMMARIZATION_PRESERVE_KEYWORDS": "true",
    "NEXS_SUMMARIZATION_EXTRACTIVE": "true"
  }
}
```

### Adaptive Cache TTL

Dynamically adjust cache TTL based on access patterns.

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_ADAPTIVE_CACHE_ENABLED` | `false` | Enable adaptive cache TTL |
| `NEXS_ADAPTIVE_CACHE_MIN_TTL` | `"1h"` | Minimum cache TTL |
| `NEXS_ADAPTIVE_CACHE_MAX_TTL` | `"168h"` | Maximum cache TTL (7 days) |
| `NEXS_ADAPTIVE_CACHE_BASE_TTL` | `"24h"` | Baseline cache TTL |

**Example:**

```json
{
  "terminal.integrated.env.linux": {
    "NEXS_ADAPTIVE_CACHE_ENABLED": "true",
    "NEXS_ADAPTIVE_CACHE_MIN_TTL": "30m",
    "NEXS_ADAPTIVE_CACHE_MAX_TTL": "336h",
    "NEXS_ADAPTIVE_CACHE_BASE_TTL": "48h"
  }
}
```

### Prompt Compression

Compress prompts to reduce token usage and costs.

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_PROMPT_COMPRESSION_ENABLED` | `false` | Enable prompt compression |
| `NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY` | `true` | Remove syntactic redundancies |
| `NEXS_PROMPT_COMPRESSION_WHITESPACE` | `true` | Normalize whitespace |
| `NEXS_PROMPT_COMPRESSION_ALIASES` | `true` | Replace verbose phrases with aliases |
| `NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE` | `true` | Maintain JSON/YAML structure |
| `NEXS_PROMPT_COMPRESSION_RATIO` | `0.65` | Target compression ratio (0.65 = 35% reduction) |
| `NEXS_PROMPT_COMPRESSION_MIN_LENGTH` | `500` | Only compress prompts longer than this |

**Example:**

```json
{
  "terminal.integrated.env.linux": {
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

## Production Settings

### Complete Production Configuration

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  
  "terminal.integrated.env.linux": {
    // Core Configuration
    "NEXS_SERVER_NAME": "nexs-mcp-prod",
    "NEXS_STORAGE_TYPE": "file",
    "NEXS_DATA_DIR": "/var/lib/nexs-mcp/data",
    "NEXS_LOG_LEVEL": "warn",
    "NEXS_LOG_FORMAT": "json",
    
    // Auto-Save (Recommended)
    "NEXS_AUTO_SAVE_MEMORIES": "true",
    "NEXS_AUTO_SAVE_INTERVAL": "3m",
    
    // Resources Protocol
    "NEXS_RESOURCES_ENABLED": "true",
    "NEXS_RESOURCES_CACHE_TTL": "10m",
    
    // Compression (Reduces Bandwidth)
    "NEXS_COMPRESSION_ENABLED": "true",
    "NEXS_COMPRESSION_ALGORITHM": "gzip",
    "NEXS_COMPRESSION_MIN_SIZE": "512",
    "NEXS_COMPRESSION_LEVEL": "9",
    "NEXS_COMPRESSION_ADAPTIVE": "true",
    
    // Streaming (Better UX)
    "NEXS_STREAMING_ENABLED": "true",
    "NEXS_STREAMING_CHUNK_SIZE": "20",
    "NEXS_STREAMING_THROTTLE": "25ms",
    "NEXS_STREAMING_BUFFER_SIZE": "200",
    
    // Summarization (Memory Optimization)
    "NEXS_SUMMARIZATION_ENABLED": "true",
    "NEXS_SUMMARIZATION_AGE": "72h",
    "NEXS_SUMMARIZATION_MAX_LENGTH": "300",
    "NEXS_SUMMARIZATION_RATIO": "0.25",
    "NEXS_SUMMARIZATION_PRESERVE_KEYWORDS": "true",
    "NEXS_SUMMARIZATION_EXTRACTIVE": "true",
    
    // Adaptive Cache (Performance)
    "NEXS_ADAPTIVE_CACHE_ENABLED": "true",
    "NEXS_ADAPTIVE_CACHE_MIN_TTL": "30m",
    "NEXS_ADAPTIVE_CACHE_MAX_TTL": "336h",
    "NEXS_ADAPTIVE_CACHE_BASE_TTL": "48h",
    
    // Prompt Compression (Token Optimization)
    "NEXS_PROMPT_COMPRESSION_ENABLED": "true",
    "NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY": "true",
    "NEXS_PROMPT_COMPRESSION_WHITESPACE": "true",
    "NEXS_PROMPT_COMPRESSION_ALIASES": "true",
    "NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE": "true",
    "NEXS_PROMPT_COMPRESSION_RATIO": "0.60",
    "NEXS_PROMPT_COMPRESSION_MIN_LENGTH": "300"
  },
  
  "terminal.integrated.env.osx": {
    // Same as linux for macOS
  },
  
  "terminal.integrated.env.windows": {
    // Same as linux for Windows
  }
}
```

### Performance Recommendations

**Low Resource Environments:**
```json
{
  "terminal.integrated.env.linux": {
    "NEXS_LOG_LEVEL": "warn",
    "NEXS_COMPRESSION_ENABLED": "true",
    "NEXS_COMPRESSION_LEVEL": "6",
    "NEXS_STREAMING_ENABLED": "true",
    "NEXS_STREAMING_CHUNK_SIZE": "10",
    "NEXS_ADAPTIVE_CACHE_ENABLED": "false"
  }
}
```

**High Performance:**
```json
{
  "terminal.integrated.env.linux": {
    "NEXS_LOG_LEVEL": "error",
    "NEXS_COMPRESSION_ENABLED": "true",
    "NEXS_COMPRESSION_LEVEL": "9",
    "NEXS_STREAMING_ENABLED": "true",
    "NEXS_STREAMING_CHUNK_SIZE": "50",
    "NEXS_STREAMING_BUFFER_SIZE": "500",
    "NEXS_ADAPTIVE_CACHE_ENABLED": "true",
    "NEXS_ADAPTIVE_CACHE_MAX_TTL": "720h"
  }
}
```

## Future Features (ONNX/Vector Search)

**Status:** Planned for Sprints 5-7 (see [Competitive Analysis](./analysis/COMPETITIVE_ANALYSIS_MEMORY_MCP.md))

### ONNX Runtime Configuration

When ONNX support is implemented:

```json
{
  "terminal.integrated.env.linux": {
    // ONNX Runtime
    "NEXS_ONNX_ENABLED": "true",
    "NEXS_ONNX_MODEL_PATH": "./models/all-MiniLM-L6-v2.onnx",
    "NEXS_ONNX_TOKENIZER_PATH": "./models/tokenizer.json",
    "NEXS_ONNX_EMBEDDING_DIM": "384",
    "NEXS_ONNX_MAX_SEQ_LENGTH": "256",
    "NEXS_ONNX_DEVICE": "cpu",
    "NEXS_ONNX_NUM_THREADS": "4"
  }
}
```

**ONNX Configuration Reference:**

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_ONNX_ENABLED` | `false` | Enable ONNX embeddings |
| `NEXS_ONNX_MODEL_PATH` | `"./models/model.onnx"` | Path to ONNX model file |
| `NEXS_ONNX_TOKENIZER_PATH` | `"./models/tokenizer.json"` | Path to tokenizer |
| `NEXS_ONNX_EMBEDDING_DIM` | `384` | Embedding dimensions |
| `NEXS_ONNX_MAX_SEQ_LENGTH` | `256` | Maximum sequence length |
| `NEXS_ONNX_DEVICE` | `"cpu"` | Device: `"cpu"`, `"cuda"`, `"rocm"` |
| `NEXS_ONNX_NUM_THREADS` | `4` | Number of CPU threads |

### Vector Search Configuration

```json
{
  "terminal.integrated.env.linux": {
    // Vector Index
    "NEXS_VECTOR_ENABLED": "true",
    "NEXS_VECTOR_INDEX_TYPE": "hnsw",
    "NEXS_VECTOR_M": "16",
    "NEXS_VECTOR_EF_CONSTRUCTION": "200",
    "NEXS_VECTOR_EF_SEARCH": "50",
    "NEXS_VECTOR_DISTANCE_METRIC": "cosine"
  }
}
```

**Vector Search Configuration Reference:**

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_VECTOR_ENABLED` | `false` | Enable vector search |
| `NEXS_VECTOR_INDEX_TYPE` | `"hnsw"` | Index type: `"hnsw"`, `"flat"` |
| `NEXS_VECTOR_M` | `16` | HNSW M parameter (connectivity) |
| `NEXS_VECTOR_EF_CONSTRUCTION` | `200` | HNSW efConstruction (build quality) |
| `NEXS_VECTOR_EF_SEARCH` | `50` | HNSW efSearch (search quality) |
| `NEXS_VECTOR_DISTANCE_METRIC` | `"cosine"` | Distance metric: `"cosine"`, `"euclidean"`, `"dot"` |

### Hybrid Search Configuration

```json
{
  "terminal.integrated.env.linux": {
    // Hybrid Search (Vector + BM25)
    "NEXS_HYBRID_SEARCH_ENABLED": "true",
    "NEXS_HYBRID_ALPHA": "0.7",
    "NEXS_HYBRID_MIN_SCORE": "0.5"
  }
}
```

**Hybrid Search Configuration Reference:**

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `NEXS_HYBRID_SEARCH_ENABLED` | `false` | Enable hybrid search |
| `NEXS_HYBRID_ALPHA` | `0.7` | Weight for vector search (0-1, higher = more vector) |
| `NEXS_HYBRID_MIN_SCORE` | `0.5` | Minimum similarity score |

### Setting Up ONNX Models

**1. Download Pre-trained Model:**

```bash
mkdir -p models

# all-MiniLM-L6-v2 (384 dimensions, recommended)
wget https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/onnx/model.onnx \
  -O models/all-MiniLM-L6-v2.onnx
wget https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/tokenizer.json \
  -O models/tokenizer.json
```

**2. Install ONNX Runtime (Go):**

```bash
go get github.com/yalue/onnxruntime_go
```

**3. System Dependencies:**

- **Linux:** `libonnxruntime.so` (install via package manager)
- **macOS:** `libonnxruntime.dylib` (via Homebrew)
- **Windows:** `onnxruntime.dll` (download from ONNX releases)

**4. GPU Acceleration (Optional):**

For CUDA support:

```bash
# Install CUDA toolkit
sudo apt install nvidia-cuda-toolkit

# Configure settings
{
  "NEXS_ONNX_DEVICE": "cuda",
  "NEXS_ONNX_NUM_THREADS": "1"
}
```

### Recommended Models

| Model | Dimensions | Speed | Quality | Use Case |
|-------|-----------|-------|---------|----------|
| all-MiniLM-L6-v2 | 384 | Fast | Good | General purpose, low resource |
| all-mpnet-base-v2 | 768 | Medium | Better | Higher quality, more resources |
| multi-qa-MiniLM-L6-cos-v1 | 384 | Fast | Good | Question-answering |
| paraphrase-multilingual | 384 | Fast | Good | Multilingual support |

## Configuration Priority

NEXS MCP loads configuration in this order (later overrides earlier):

1. **Default values** (hardcoded in `internal/config/config.go`)
2. **Environment variables** (from shell or `settings.json`)
3. **Command-line flags** (when launching manually)

### Using Command-Line Flags

Override any setting via flags:

```bash
./nexs-mcp \
  --storage=file \
  --data-dir=/custom/path \
  --log-level=debug \
  --resources-enabled=true \
  --compression-enabled=true \
  --streaming-enabled=true \
  --adaptive-cache-enabled=true
```

**Available Flags:**

```
--storage                       Storage type (file/memory)
--data-dir                      Data directory path
--log-level                     Log level (debug/info/warn/error)
--log-format                    Log format (json/text)
--resources-enabled             Enable MCP Resources
--resources-cache-ttl           Resource cache TTL
--auto-save-memories            Enable auto-save
--auto-save-interval            Auto-save interval
--compression-enabled           Enable compression
--compression-algorithm         Compression algorithm
--compression-level             Compression level
--streaming-enabled             Enable streaming
--streaming-chunk-size          Streaming chunk size
--summarization-enabled         Enable summarization
--adaptive-cache-enabled        Enable adaptive cache
--prompt-compression-enabled    Enable prompt compression
```

## Troubleshooting

### Configuration Not Applied

**Problem:** Environment variables not taking effect

**Solutions:**

1. **Reload VSCode/Cursor** after changing `settings.json`
2. **Restart terminal** if running server manually
3. **Check OS-specific env block**: Use correct `terminal.integrated.env.{linux|osx|windows}`
4. **Verify syntax**: JSON must be valid (no trailing commas, proper quotes)

### Logs Not Appearing

**Problem:** No log output or wrong format

**Solutions:**

```json
{
  "terminal.integrated.env.linux": {
    "NEXS_LOG_LEVEL": "debug",
    "NEXS_LOG_FORMAT": "text"
  }
}
```

### Storage Issues

**Problem:** Cannot read/write files

**Solutions:**

1. **Check permissions:**
   ```bash
   chmod -R 755 data/elements
   ```

2. **Verify path exists:**
   ```json
   {
     "NEXS_DATA_DIR": "/absolute/path/to/data"
   }
   ```

3. **Use memory storage for testing:**
   ```json
   {
     "NEXS_STORAGE_TYPE": "memory"
   }
   ```

### Performance Issues

**Problem:** Server feels slow

**Solutions:**

1. **Enable caching:**
   ```json
   {
     "NEXS_ADAPTIVE_CACHE_ENABLED": "true",
     "NEXS_RESOURCES_CACHE_TTL": "10m"
   }
   ```

2. **Enable compression:**
   ```json
   {
     "NEXS_COMPRESSION_ENABLED": "true",
     "NEXS_COMPRESSION_LEVEL": "6"
   }
   ```

3. **Enable streaming:**
   ```json
   {
     "NEXS_STREAMING_ENABLED": "true",
     "NEXS_STREAMING_CHUNK_SIZE": "20"
   }
   ```

### Memory Usage High

**Problem:** Server using too much RAM

**Solutions:**

1. **Enable summarization:**
   ```json
   {
     "NEXS_SUMMARIZATION_ENABLED": "true",
     "NEXS_SUMMARIZATION_AGE": "48h"
   }
   ```

2. **Reduce cache TTL:**
   ```json
   {
     "NEXS_RESOURCES_CACHE_TTL": "2m",
     "NEXS_ADAPTIVE_CACHE_BASE_TTL": "12h"
   }
   ```

3. **Lower streaming buffer:**
   ```json
   {
     "NEXS_STREAMING_BUFFER_SIZE": "50"
   }
   ```

## See Also

- [Setup Guide](./development/SETUP.md) - Complete development environment setup
- [Getting Started](./user-guide/GETTING_STARTED.md) - First steps with NEXS MCP
- [Configuration Reference](../internal/config/config.go) - Source code documentation
- [Competitive Analysis](./analysis/COMPETITIVE_ANALYSIS_MEMORY_MCP.md) - Roadmap and future features
- [MCP Tools](./api/MCP_TOOLS.md) - All 66 available MCP tools
