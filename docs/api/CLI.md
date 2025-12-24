# NEXS-MCP CLI Reference

**Version:** v1.2.0  
**Last Updated:** December 22, 2025

Complete command-line interface reference for NEXS-MCP.

---

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#usage)
- [Command-Line Flags](#command-line-flags)
- [Environment Variables](#environment-variables)
- [Configuration File](#configuration-file)
- [Examples](#examples)
- [Exit Codes](#exit-codes)
- [Logging](#logging)
- [Data Management](#data-management)
- [Troubleshooting](#troubleshooting)

---

## Installation

### Go Install

```bash
go install github.com/fsvxavier/nexs-mcp/cmd/nexs-mcp@latest
```

Installs to `$GOPATH/bin/nexs-mcp` (usually `~/go/bin/nexs-mcp`)

### Homebrew

```bash
brew install nexs-mcp
```

### NPM

```bash
# Global installation
npm install -g @fsvxavier/nexs-mcp-server

# Run with npx (no installation)
npx @fsvxavier/nexs-mcp-server
```

### Docker

```bash
# Pull from Docker Hub
docker pull fsvxavier/nexs-mcp:latest

# Or specific version
docker pull fsvxavier/nexs-mcp:v1.2.0

# Run container
docker run -v $(pwd)/data:/app/data fsvxavier/nexs-mcp:latest
```

🐳 **Docker Hub:** https://hub.docker.com/r/fsvxavier/nexs-mcp  
📦 **Image Size:** 14.5 MB (compressed), 53.7 MB (uncompressed)

### From Source

```bash
git clone https://github.com/fsvxavier/nexs-mcp.git
cd nexs-mcp
make build
./bin/nexs-mcp
```

---

## Usage

```bash
nexs-mcp [flags]
```

### Basic Startup

```bash
# Start with default settings
nexs-mcp

# Start with custom data directory
nexs-mcp --data-dir=/path/to/data

# Start with debug logging
nexs-mcp --log-level=debug

# Start with resources enabled
nexs-mcp --resources-enabled=true
```

---

## Command-Line Flags

### Data Storage

#### `--data-dir <path>`
**Type:** string  
**Default:** `~/.nexs-mcp/data`  
**Description:** Directory for storing element files

**Example:**
```bash
nexs-mcp --data-dir=/custom/path/elements
```

#### `--storage-type <type>`
**Type:** string  
**Default:** `yaml`  
**Options:** `yaml`, `json`  
**Description:** File format for element storage

**Example:**
```bash
# Use JSON format
nexs-mcp --storage-type=json
```

---

### Logging

#### `--log-level <level>`
**Type:** string  
**Default:** `info`  
**Options:** `debug`, `info`, `warn`, `error`  
**Description:** Logging verbosity level

**Example:**
```bash
# Debug logging for troubleshooting
nexs-mcp --log-level=debug

# Only errors
nexs-mcp --log-level=error
```

#### `--log-format <format>`
**Type:** string  
**Default:** `text`  
**Options:** `text`, `json`  
**Description:** Log output format

**Example:**
```bash
# JSON logs for log aggregation
nexs-mcp --log-format=json
```

#### `--log-file <path>`
**Type:** string  
**Default:** stdout  
**Description:** Write logs to file instead of stdout

**Example:**
```bash
# Log to file
nexs-mcp --log-file=/var/log/nexs-mcp.log

# Log to file with rotation handled externally
nexs-mcp --log-file=/var/log/nexs-mcp.log --log-format=json
```

---

### MCP Resources

#### `--resources-enabled`
**Type:** boolean  
**Default:** `false`  
**Description:** Enable MCP Resources Protocol

**Example:**
```bash
nexs-mcp --resources-enabled=true
```

#### `--resources-expose <uris>`
**Type:** comma-separated list  
**Default:** all resources  
**Options:** `summary`, `full`, `stats`  
**Description:** Specific resource URIs to expose

**Example:**
```bash
# Expose only summary and stats
nexs-mcp --resources-enabled=true --resources-expose=summary,stats
```

#### `--resources-cache-ttl <seconds>`
**Type:** integer  
**Default:** `3600` (1 hour)  
**Description:** Resource cache Time-To-Live in seconds

**Example:**
```bash
# 2-hour cache
nexs-mcp --resources-cache-ttl=7200

# No caching (always fresh)
nexs-mcp --resources-cache-ttl=0
```

---

### GitHub Integration

#### `--github-client-id <id>`
**Type:** string  
**Default:** built-in client ID  
**Description:** GitHub OAuth application client ID

**Example:**
```bash
nexs-mcp --github-client-id=your_client_id_here
```

**Note:** Only needed if using a custom GitHub OAuth app

---

### Configuration

#### `--config <path>`
**Type:** string  
**Default:** none  
**Description:** Path to YAML configuration file

**Example:**
```bash
nexs-mcp --config=/etc/nexs-mcp/config.yaml
```

**Precedence:** CLI flags > Environment variables > Config file > Defaults

---

### Utility Flags

#### `--version`
**Type:** boolean  
**Description:** Print version information and exit

**Example:**
```bash
nexs-mcp --version
# Output: NEXS-MCP v1.0.0 (build: abc123, date: 2025-12-20)
```

#### `--help`
**Type:** boolean  
**Description:** Print help message and exit

**Example:**
```bash
nexs-mcp --help
```

---

## Environment Variables

Environment variables provide an alternative to command-line flags. They are overridden by CLI flags if both are specified.

### Variable List

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `NEXS_DATA_DIR` | string | `~/.nexs-mcp/data` | Data directory path |
| `NEXS_STORAGE_TYPE` | string | `yaml` | Storage format (yaml/json) |
| `NEXS_LOG_LEVEL` | string | `info` | Log level (debug/info/warn/error) |
| `NEXS_LOG_FORMAT` | string | `text` | Log format (text/json) |
| `NEXS_LOG_FILE` | string | stdout | Log file path |
| `NEXS_RESOURCES_ENABLED` | boolean | `false` | Enable MCP Resources |
| `NEXS_RESOURCES_EXPOSE` | string | all | Resource URIs (comma-separated) |
| `NEXS_RESOURCES_CACHE_TTL` | integer | `3600` | Cache TTL in seconds |
| `NEXS_GITHUB_CLIENT_ID` | string | built-in | GitHub OAuth client ID |

### Examples

```bash
# Set data directory
export NEXS_DATA_DIR=/custom/path/data
nexs-mcp

# Enable debug logging
export NEXS_LOG_LEVEL=debug
export NEXS_LOG_FORMAT=json
nexs-mcp

# Enable resources with 2-hour cache
export NEXS_RESOURCES_ENABLED=true
export NEXS_RESOURCES_CACHE_TTL=7200
nexs-mcp

# Complete environment configuration
export NEXS_DATA_DIR=/data
export NEXS_STORAGE_TYPE=json
export NEXS_LOG_LEVEL=info
export NEXS_LOG_FORMAT=json
export NEXS_LOG_FILE=/var/log/nexs.log
export NEXS_RESOURCES_ENABLED=true
nexs-mcp
```

---

## Configuration File

NEXS-MCP supports YAML configuration files for complex setups.

### File Location

Specify with `--config` flag:
```bash
nexs-mcp --config=/path/to/config.yaml
```

Or place in default locations (checked in order):
1. `./nexs-mcp.yaml`
2. `~/.nexs-mcp/config.yaml`
3. `/etc/nexs-mcp/config.yaml`

### Configuration Schema

```yaml
# Data storage configuration
data_dir: ~/.nexs-mcp/data
storage_type: yaml  # yaml or json

# Logging configuration
logging:
  level: info       # debug, info, warn, error
  format: text      # text or json
  file: ""          # Empty string = stdout

# MCP Resources configuration
resources:
  enabled: false
  expose:           # List of resources to expose (empty = all)
    - summary
    - full
    - stats
  cache_ttl: 3600   # Cache TTL in seconds

# GitHub integration
github:
  client_id: ""     # Empty = use built-in client ID

# Auto-save configuration
auto_save:
  enabled: true
  interval: 300     # Auto-save interval in seconds (5 minutes)
  on_exit: true     # Save on graceful exit

# Performance tuning
performance:
  max_concurrent_operations: 10
  index_rebuild_interval: 3600  # Seconds (1 hour)
  
# Feature flags
features:
  enable_analytics: true
  enable_performance_tracking: true
```

### Example Configurations

#### Development Configuration

```yaml
# dev-config.yaml
data_dir: ./dev-data
storage_type: yaml

logging:
  level: debug
  format: text

resources:
  enabled: true
  cache_ttl: 60  # 1 minute for faster testing

auto_save:
  enabled: true
  interval: 30  # 30 seconds for dev
```

**Usage:**
```bash
nexs-mcp --config=dev-config.yaml
```

---

#### Production Configuration

```yaml
# prod-config.yaml
data_dir: /var/lib/nexs-mcp/data
storage_type: json

logging:
  level: info
  format: json
  file: /var/log/nexs-mcp/server.log

resources:
  enabled: true
  expose:
    - summary
    - stats
  cache_ttl: 7200  # 2 hours

github:
  client_id: your_production_client_id

auto_save:
  enabled: true
  interval: 300
  on_exit: true

performance:
  max_concurrent_operations: 20
  index_rebuild_interval: 1800  # 30 minutes

features:
  enable_analytics: true
  enable_performance_tracking: true
```

**Usage:**
```bash
nexs-mcp --config=/etc/nexs-mcp/prod-config.yaml
```

---

#### Docker Configuration

```yaml
# docker-config.yaml
data_dir: /data
storage_type: yaml

logging:
  level: info
  format: json

resources:
  enabled: true
  cache_ttl: 3600

auto_save:
  enabled: true
  interval: 300
```

**Usage with Docker:**
```bash
docker run -v $(pwd)/docker-config.yaml:/config.yaml \
  -v $(pwd)/data:/data \
  fsvxavier/nexs-mcp:latest \
  --config=/config.yaml
```

---

## Examples

### Basic Examples

#### Start with Default Settings
```bash
nexs-mcp
```

#### Custom Data Directory
```bash
nexs-mcp --data-dir=~/my-nexs-data
```

#### Enable Debug Logging
```bash
nexs-mcp --log-level=debug
```

#### JSON Storage
```bash
nexs-mcp --storage-type=json
```

---

### Advanced Examples

#### Production Setup
```bash
nexs-mcp \
  --data-dir=/var/lib/nexs-mcp/data \
  --storage-type=json \
  --log-level=info \
  --log-format=json \
  --log-file=/var/log/nexs-mcp.log \
  --resources-enabled=true \
  --resources-cache-ttl=7200
```

#### Development Setup
```bash
nexs-mcp \
  --data-dir=./dev-data \
  --log-level=debug \
  --resources-enabled=true \
  --resources-cache-ttl=60
```

#### Docker Compose
```yaml
version: '3.8'

services:
  nexs-mcp:
    image: fsvxavier/nexs-mcp:latest
    command:
      - --data-dir=/data
      - --log-level=info
      - --log-format=json
      - --resources-enabled=true
    volumes:
      - ./data:/data
      - ./logs:/logs
    environment:
      - NEXS_LOG_FILE=/logs/nexs-mcp.log
      - NEXS_RESOURCES_CACHE_TTL=3600
    restart: unless-stopped
```

---

### Claude Desktop Integration

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "nexs-mcp",
      "args": [
        "--resources-enabled=true",
        "--log-level=info"
      ],
      "env": {
        "NEXS_DATA_DIR": "~/.nexs-mcp/data",
        "NEXS_LOG_FILE": "~/.nexs-mcp/logs/server.log"
      }
    }
  }
}
```

**Linux:** `~/.config/Claude/claude_desktop_config.json`

**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

---

## Exit Codes

NEXS-MCP uses standard exit codes:

| Code | Meaning | Description |
|------|---------|-------------|
| 0 | Success | Normal exit |
| 1 | General Error | Unspecified error |
| 2 | Misuse | Invalid command-line arguments |
| 3 | Config Error | Configuration file error |
| 4 | Data Error | Data directory or storage error |
| 5 | Permission Error | Insufficient permissions |
| 130 | Interrupted | Terminated by Ctrl+C (SIGINT) |
| 143 | Terminated | Terminated by SIGTERM |

**Examples:**
```bash
# Check exit code
nexs-mcp --invalid-flag
echo $?  # Output: 2

# Graceful shutdown
nexs-mcp &
PID=$!
kill -TERM $PID
wait $PID
echo $?  # Output: 143
```

---

## Logging

### Log Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| `debug` | Detailed diagnostic information | Development, troubleshooting |
| `info` | General informational messages | Production, monitoring |
| `warn` | Warning messages (non-critical) | Important events |
| `error` | Error messages | Failures, issues |

### Log Formats

#### Text Format (default)

```
2025-12-20 10:00:00 INFO  Server started port=stdio version=v1.0.0
2025-12-20 10:00:01 DEBUG Element loaded id=persona-001 type=persona
2025-12-20 10:00:02 INFO  Tool called tool=list_elements duration_ms=23
2025-12-20 10:00:03 WARN  Rate limit approaching remaining=100 limit=5000
2025-12-20 10:00:04 ERROR Failed to sync error="connection timeout"
```

#### JSON Format

```json
{"time":"2025-12-20T10:00:00Z","level":"info","msg":"Server started","port":"stdio","version":"v1.0.0"}
{"time":"2025-12-20T10:00:01Z","level":"debug","msg":"Element loaded","id":"persona-001","type":"persona"}
{"time":"2025-12-20T10:00:02Z","level":"info","msg":"Tool called","tool":"list_elements","duration_ms":23}
{"time":"2025-12-20T10:00:03Z","level":"warn","msg":"Rate limit approaching","remaining":100,"limit":5000}
{"time":"2025-12-20T10:00:04Z","level":"error","msg":"Failed to sync","error":"connection timeout"}
```

### Log Rotation

NEXS-MCP does not implement built-in log rotation. Use external tools:

#### Linux/macOS with logrotate

```bash
# /etc/logrotate.d/nexs-mcp
/var/log/nexs-mcp.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 nexs nexs
    postrotate
        pkill -HUP nexs-mcp
    endscript
}
```

#### Docker with log drivers

```yaml
services:
  nexs-mcp:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

---

## Data Management

### Data Directory Structure

```
~/.nexs-mcp/
├── data/                    # Element storage
│   ├── personas/
│   │   ├── persona-001.yaml
│   │   └── persona-002.yaml
│   ├── skills/
│   ├── templates/
│   ├── agents/
│   ├── memories/
│   └── ensembles/
├── auth/                    # Authentication tokens
│   └── github_token.enc
├── cache/                   # Resource cache
│   └── registry-cache.json
├── metrics/                 # Usage statistics
│   └── usage-stats.json
├── performance/             # Performance metrics
│   └── perf-metrics.json
├── backups/                 # Backup files
│   └── nexs-backup-*.tar.gz
└── logs/                    # Log files (if configured)
    └── server.log
```

### Backup

**Create backup:**
```bash
# Use MCP tool (recommended)
# Call backup_portfolio tool via MCP client

# Manual backup
tar -czf nexs-backup-$(date +%Y%m%d).tar.gz ~/.nexs-mcp/data
```

**Restore from backup:**
```bash
# Use MCP tool (recommended)
# Call restore_portfolio tool via MCP client

# Manual restore
tar -xzf nexs-backup-20251220.tar.gz -C ~/.nexs-mcp/data
```

### Migration

**YAML to JSON:**
```bash
# Start with JSON storage
nexs-mcp --data-dir=./new-data --storage-type=json

# Use sync tools to migrate
# Elements are automatically converted during sync
```

**Move to new location:**
```bash
# Copy data
cp -r ~/.nexs-mcp/data /new/location/

# Update config or use flag
nexs-mcp --data-dir=/new/location
```

---

## Troubleshooting

### Common Issues

#### Server Won't Start

**Problem:** Server fails to start

**Solutions:**
```bash
# Check port availability (not applicable for stdio, but useful for future)
# Verify data directory permissions
ls -la ~/.nexs-mcp/data

# Check logs
nexs-mcp --log-level=debug

# Verify configuration
nexs-mcp --config=/path/to/config.yaml --log-level=debug
```

#### Elements Not Loading

**Problem:** Elements don't appear in list

**Solutions:**
```bash
# Check data directory
ls ~/.nexs-mcp/data/personas/

# Verify storage type matches files
nexs-mcp --storage-type=yaml  # or json

# Reload elements
# Use reload_elements MCP tool
```

#### GitHub Authentication Fails

**Problem:** GitHub OAuth not working

**Solutions:**
```bash
# Check GitHub status
curl -I https://api.github.com

# Try re-authenticating
# Use github_auth_start tool

# Check token file
ls -la ~/.nexs-mcp/auth/github_token.enc
```

#### Performance Issues

**Problem:** Slow response times

**Solutions:**
```bash
# Enable performance metrics
nexs-mcp --log-level=info

# Check index stats
# Use get_capability_index_stats tool

# Rebuild index
# Use reload_elements tool with clear_cache=true

# Increase cache TTL
nexs-mcp --resources-cache-ttl=7200
```

---

## systemd Service (Linux)

**Create service file:** `/etc/systemd/system/nexs-mcp.service`

```ini
[Unit]
Description=NEXS-MCP Server
After=network.target

[Service]
Type=simple
User=nexs
Group=nexs
WorkingDirectory=/var/lib/nexs-mcp
ExecStart=/usr/local/bin/nexs-mcp \
    --data-dir=/var/lib/nexs-mcp/data \
    --log-file=/var/log/nexs-mcp/server.log \
    --log-level=info \
    --log-format=json \
    --resources-enabled=true

Restart=always
RestartSec=10

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/nexs-mcp /var/log/nexs-mcp

[Install]
WantedBy=multi-user.target
```

**Enable and start:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable nexs-mcp
sudo systemctl start nexs-mcp
sudo systemctl status nexs-mcp
```

---

## See Also

- [MCP Tools Reference](./MCP_TOOLS.md)
- [MCP Resources Reference](./MCP_RESOURCES.md)
- [Getting Started Guide](../user-guide/GETTING_STARTED.md)
- [Troubleshooting Guide](../user-guide/TROUBLESHOOTING.md)

---

**Last Updated:** December 20, 2025  
**NEXS-MCP Version:** v1.0.0
