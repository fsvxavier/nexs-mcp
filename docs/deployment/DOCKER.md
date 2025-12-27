# Docker Deployment Guide

This guide explains how to build, run, and deploy NEXS-MCP using Docker with full ONNX support.

## Overview

The NEXS-MCP Docker image includes:

- **ONNX Runtime v1.23.2**: High-performance inference engine
- **Pre-loaded Models**: 
  - MS MARCO MiniLM-L-6-v2 (87MB, 9 languages)
  - Paraphrase-Multilingual-MiniLM-L12-v2 (449MB, 11 languages including CJK)
- **All Features Enabled**: Complete configuration support
- **Production Ready**: Optimized for performance and security

## Quick Start

### 1. Pull and Run

```bash
# Pull from Docker Hub
docker pull fsvxavier/nexs-mcp:latest

# Run with default settings
docker run -d \
  --name nexs-mcp \
  -v $(pwd)/data:/app/data \
  fsvxavier/nexs-mcp:latest
```

### 2. Docker Compose (Recommended)

```bash
# Copy environment file
cp .env.example .env

# Edit .env with your settings
nano .env

# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Building from Source

### Build Docker Image

```bash
# Build with ONNX support (downloads ~500MB of models)
make docker-build

# Or manually
docker build -t nexs-mcp:local .
```

**Note**: The build process:
- Downloads ONNX Runtime v1.23.2 (~10MB)
- Downloads MS MARCO model (87MB)
- Downloads Paraphrase-Multilingual model (449MB)
- Total download size: ~550MB
- Build time: 5-10 minutes on first build

### Publish to Docker Hub

```bash
# Requires .env with DOCKER_USER and DOCKER_TOKEN
make docker-publish
```

## Configuration

### Environment Variables

All NEXS-MCP configurations can be set via environment variables. See [.env.example](../../.env.example) for complete list.

#### Core Settings

```bash
# Storage and data
NEXS_STORAGE_TYPE=filesystem
NEXS_DATA_DIR=/app/data

# Logging
NEXS_LOG_LEVEL=info
NEXS_LOG_FORMAT=json
```

#### Embedding Provider

```bash
# Use local ONNX models (default, included in image)
NEXS_EMBEDDING_PROVIDER=onnx
NEXS_TRANSFORMERS_CACHE=/app/models
NEXS_USE_GPU=false
LD_LIBRARY_PATH=/usr/local/lib

# Or use OpenAI API
NEXS_EMBEDDING_PROVIDER=openai
OPENAI_API_KEY=your_api_key_here
NEXS_OPENAI_MODEL=text-embedding-3-small
```

#### Working Memory

```bash
# Auto-save memories to disk
NEXS_AUTO_SAVE_MEMORIES=true
NEXS_AUTO_SAVE_INTERVAL=30s
```

#### Resources Protocol

```bash
# Enable resources protocol (use with caution)
NEXS_RESOURCES_ENABLED=false
NEXS_RESOURCES_CACHE_TTL=5m
```

#### Performance Tuning

```bash
# Go runtime settings
GOMAXPROCS=0           # 0 = use all CPUs
GOMEMLIMIT=2GiB        # Memory limit
```

#### GitHub Integration

```bash
# Optional GitHub integration
GITHUB_TOKEN=your_token_here
GITHUB_CLIENT_ID=your_client_id
```

### Volume Mounts

The Docker image uses several volumes:

```yaml
volumes:
  # Data directory (required)
  - ./data:/app/data:rw
  
  # Models cache (optional, models pre-loaded in image)
  - nexs-mcp-models:/app/models:rw
  
  # Transformers cache (optional)
  - nexs-mcp-cache:/root/.cache:rw
```

## Advanced Usage

### Custom Configuration

Create a custom docker-compose.yml:

```yaml
version: '3.8'

services:
  nexs-mcp:
    image: fsvxavier/nexs-mcp:latest
    container_name: my-nexs-mcp
    restart: unless-stopped
    
    volumes:
      - ./my-data:/app/data:rw
    
    environment:
      - NEXS_LOG_LEVEL=debug
      - NEXS_EMBEDDING_PROVIDER=onnx
      - NEXS_AUTO_SAVE_MEMORIES=true
      - GOMAXPROCS=0
      - GOMEMLIMIT=2GiB
    
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 2GiB
```

### GPU Support

To enable GPU acceleration for ONNX:

1. Install NVIDIA Docker runtime
2. Update docker-compose.yml:

```yaml
services:
  nexs-mcp:
    runtime: nvidia
    environment:
      - NEXS_USE_GPU=true
      - NVIDIA_VISIBLE_DEVICES=all
```

### Development Mode

For development with live reload:

```bash
# Mount source code
docker run -it \
  --name nexs-mcp-dev \
  -v $(pwd):/app \
  -v $(pwd)/data:/app/data \
  -e NEXS_LOG_LEVEL=debug \
  nexs-mcp:local \
  /bin/bash
```

## Image Architecture

### Multi-Stage Build

```
Stage 1: Builder
├── Ubuntu 22.04
├── Go 1.25
├── ONNX Runtime v1.23.2
└── Build with CGO_ENABLED=1

Stage 2: Runtime
├── Ubuntu 22.04
├── ONNX Runtime libraries
├── Downloaded models
└── Compiled binary
```

### Image Size

- Base image: ~100MB
- ONNX Runtime: ~10MB
- Models: ~550MB
- Binary: ~30MB
- **Total**: ~690MB

### Layers

The image is optimized for layer caching:

1. Base OS and dependencies
2. ONNX Runtime installation
3. Go dependencies
4. Source code and build
5. Model downloads
6. Configuration and entrypoint

## Troubleshooting

### Build Issues

**Problem**: Build fails with "cannot download models"

```bash
# Check network connectivity
docker build --progress=plain -t nexs-mcp:debug .

# Use proxy if needed
docker build \
  --build-arg HTTP_PROXY=http://proxy:port \
  --build-arg HTTPS_PROXY=http://proxy:port \
  -t nexs-mcp:local .
```

**Problem**: Build is slow

```bash
# Use BuildKit for faster builds
DOCKER_BUILDKIT=1 docker build -t nexs-mcp:local .
```

### Runtime Issues

**Problem**: ONNX Runtime not found

```bash
# Check library path
docker run --rm nexs-mcp:latest ldconfig -p | grep onnx

# Should output:
# libonnxruntime.so.1.23.2 (libc6,x86-64) => /usr/local/lib/libonnxruntime.so.1.23.2
```

**Problem**: Models not found

```bash
# List models
docker run --rm nexs-mcp:latest ls -la /app/models/

# Should show:
# ms-marco-MiniLM-L-6-v2/
# paraphrase-multilingual-MiniLM-L12-v2/
```

**Problem**: Memory issues

```bash
# Increase memory limit
docker run \
  -e GOMEMLIMIT=4GiB \
  --memory=4g \
  nexs-mcp:latest
```

### Performance Issues

**Problem**: Slow embeddings

```bash
# Enable all CPUs
docker run -e GOMAXPROCS=0 nexs-mcp:latest

# Or limit to specific count
docker run -e GOMAXPROCS=4 nexs-mcp:latest
```

**Problem**: High memory usage

```bash
# Reduce memory limit
docker run -e GOMEMLIMIT=1GiB nexs-mcp:latest

# Monitor memory
docker stats nexs-mcp
```

## Security Considerations

### Best Practices

1. **Use Named Volumes**: Protect data persistence

```yaml
volumes:
  nexs-mcp-data:
    driver: local
```

2. **Limit Resources**: Prevent resource exhaustion

```yaml
deploy:
  resources:
    limits:
      cpus: '4'
      memory: 2GiB
```

3. **Security Options**: Enable security features

```yaml
security_opt:
  - no-new-privileges:true
```

4. **Sensitive Data**: Use secrets for tokens

```bash
# Create secrets
echo "$GITHUB_TOKEN" | docker secret create github_token -

# Use in compose
secrets:
  - github_token
```

5. **Network Isolation**: Use custom networks

```yaml
networks:
  nexs-mcp-network:
    driver: bridge
    internal: true
```

## Monitoring

### Health Checks

Built-in health check:

```yaml
healthcheck:
  test: ["CMD", "/app/nexs-mcp", "--version"]
  interval: 30s
  timeout: 10s
  retries: 3
```

### Logging

View logs:

```bash
# Follow logs
docker logs -f nexs-mcp

# Last 100 lines
docker logs --tail 100 nexs-mcp

# With timestamps
docker logs -t nexs-mcp
```

Configure log driver:

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### Metrics

Monitor resources:

```bash
# Real-time stats
docker stats nexs-mcp

# Inspect container
docker inspect nexs-mcp
```

## Production Deployment

### Recommendations

1. **Use Docker Compose**: Easier configuration management
2. **Enable Auto-restart**: `restart: unless-stopped`
3. **Configure Logging**: Set max size and rotation
4. **Monitor Resources**: Set memory and CPU limits
5. **Backup Data**: Regular backups of data volume
6. **Update Regularly**: Keep image up to date
7. **Use Secrets**: Never commit tokens to git

### Example Production Setup

```yaml
version: '3.8'

services:
  nexs-mcp:
    image: fsvxavier/nexs-mcp:v1.3.0  # Pin version
    container_name: nexs-mcp-prod
    restart: unless-stopped
    
    volumes:
      - nexs-data:/app/data:rw
      - nexs-models:/app/models:rw
    
    environment:
      - NEXS_LOG_LEVEL=info
      - NEXS_LOG_FORMAT=json
      - NEXS_EMBEDDING_PROVIDER=onnx
      - NEXS_AUTO_SAVE_MEMORIES=true
      - NEXS_RESOURCES_ENABLED=false
      - GOMAXPROCS=0
      - GOMEMLIMIT=2GiB
    
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 2GiB
        reservations:
          cpus: '1'
          memory: 512M
    
    healthcheck:
      test: ["CMD", "/app/nexs-mcp", "--version"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    
    security_opt:
      - no-new-privileges:true
    
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "5"

volumes:
  nexs-data:
    driver: local
  nexs-models:
    driver: local

networks:
  default:
    name: nexs-mcp-network
    driver: bridge
```

## References

- [ONNX Runtime Documentation](https://onnxruntime.ai/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [NEXS-MCP Documentation](../README.md)
