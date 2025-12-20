# Docker Deployment Guide

Este guia explica como executar o NEXS-MCP usando Docker.

## Quick Start

```bash
# Pull da imagem
docker pull fsvxavier/nexs-mcp:latest

# Executar com configuração padrão
docker run -d \
  --name nexs-mcp \
  -v $(pwd)/data:/app/data \
  fsvxavier/nexs-mcp:latest
```

## Usando Docker Compose

### Configuração Básica

1. Crie um arquivo `docker-compose.yml`:

```yaml
version: '3.8'

services:
  nexs-mcp:
    image: fsvxavier/nexs-mcp:latest
    volumes:
      - ./data:/app/data
    environment:
      - LOG_LEVEL=info
```

2. Inicie o serviço:

```bash
docker-compose up -d
```

### Configuração Completa

O repositório inclui um `docker-compose.yml` completo com:
- Volume mounts para data, config, auth, sync e cache
- Environment variables configuráveis
- Resource limits
- Health checks
- Security hardening

```bash
# Clone o repositório (se ainda não tiver)
git clone https://github.com/fsvxavier/nexs-mcp.git
cd nexs-mcp

# Configure environment variables (opcional)
cp .env.example .env
# Edite .env com suas configurações

# Inicie com Docker Compose
docker-compose up -d

# Visualize logs
docker-compose logs -f nexs-mcp

# Pare o serviço
docker-compose down
```

## Volume Management

### Data Directory

Armazena todos os elementos criados:

```bash
docker run -v $(pwd)/data:/app/data fsvxavier/nexs-mcp:latest
```

Estrutura:
```
data/
├── agents/
├── personas/
├── skills/
├── templates/
├── memories/
└── ensembles/
```

### Configuration Directory

Configurações personalizadas (opcional):

```bash
docker run -v $(pwd)/config:/app/config:ro fsvxavier/nexs-mcp:latest
```

### Auth Directory

Tokens e credenciais (sensível):

```bash
docker run -v nexs-mcp-auth:/root/.nexs-mcp/auth fsvxavier/nexs-mcp:latest
```

**Importante:** Use named volume para persistência segura.

### Sync State

Estado de sincronização com GitHub:

```bash
docker run -v nexs-mcp-sync:/app/.nexs-sync fsvxavier/nexs-mcp:latest
```

## Environment Variables

### Logging

```bash
LOG_LEVEL=info        # debug, info, warn, error
LOG_FORMAT=json       # json, text
```

### Paths

```bash
NEXS_DATA_DIR=/app/data
NEXS_CONFIG_DIR=/app/config
```

### GitHub Integration

```bash
GITHUB_TOKEN=ghp_xxx         # OAuth token
GITHUB_OWNER=username        # Seu usuário GitHub
GITHUB_REPO=my-portfolio     # Nome do repositório
```

### Collection

```bash
COLLECTION_CACHE_TTL=86400   # Cache TTL em segundos (24h)
COLLECTION_CACHE_SIZE=100    # Tamanho máximo do cache
```

### Performance

```bash
GOMAXPROCS=4                 # Número de CPUs
GOMEMLIMIT=512MiB           # Limite de memória
```

## Docker CLI Examples

### Run com todas as configurações

```bash
docker run -d \
  --name nexs-mcp-server \
  --restart unless-stopped \
  -v $(pwd)/data:/app/data \
  -v nexs-mcp-auth:/root/.nexs-mcp/auth \
  -v nexs-mcp-sync:/app/.nexs-sync \
  -e LOG_LEVEL=info \
  -e GITHUB_TOKEN=${GITHUB_TOKEN} \
  -e GITHUB_OWNER=fsvxavier \
  -e GITHUB_REPO=my-portfolio \
  --memory=512m \
  --cpus=2 \
  fsvxavier/nexs-mcp:latest
```

### Run temporário para testes

```bash
docker run --rm -it \
  -v $(pwd)/data:/app/data \
  fsvxavier/nexs-mcp:latest \
  --help
```

### Executar comando específico

```bash
# Listar elementos
docker exec nexs-mcp-server nexs-mcp list

# Criar elemento
docker exec nexs-mcp-server nexs-mcp create persona \
  --name "DevOps Expert" \
  --description "Kubernetes specialist"

# Verificar versão
docker exec nexs-mcp-server nexs-mcp --version
```

## Image Tags

- `latest` - Última versão estável (branch main)
- `v1.0.0` - Versão específica
- `v1.0` - Última patch da minor version
- `v1` - Última minor da major version
- `main` - Development (não recomendado para produção)

```bash
# Versão específica
docker pull fsvxavier/nexs-mcp:v1.0.0

# Latest minor
docker pull fsvxavier/nexs-mcp:v1.0

# Latest major
docker pull fsvxavier/nexs-mcp:v1
```

## Multi-Architecture Support

A imagem suporta múltiplas arquiteturas:
- `linux/amd64` (x86_64)
- `linux/arm64` (ARM64/Aarch64)

Docker automaticamente seleciona a arquitetura correta:

```bash
# Em Apple Silicon (M1/M2)
docker pull fsvxavier/nexs-mcp:latest  # Usa arm64

# Em Intel/AMD
docker pull fsvxavier/nexs-mcp:latest  # Usa amd64
```

## Security Best Practices

### 1. Non-root User

A imagem executa como usuário não-privilegiado (UID 1000):

```bash
docker run --user 1000:1000 fsvxavier/nexs-mcp:latest
```

### 2. Read-only Filesystem

```bash
docker run --read-only \
  --tmpfs /tmp:size=100M \
  -v $(pwd)/data:/app/data \
  fsvxavier/nexs-mcp:latest
```

### 3. Security Options

```bash
docker run \
  --security-opt=no-new-privileges:true \
  --cap-drop=ALL \
  fsvxavier/nexs-mcp:latest
```

### 4. Network Isolation

```bash
# Criar network isolada
docker network create nexs-mcp-network

# Executar na network
docker run --network nexs-mcp-network fsvxavier/nexs-mcp:latest
```

## Health Checks

A imagem inclui health check automático:

```bash
# Verificar status
docker inspect --format='{{.State.Health.Status}}' nexs-mcp-server

# Ver últimos health checks
docker inspect --format='{{json .State.Health}}' nexs-mcp-server | jq
```

## Resource Limits

### Memory Limits

```bash
# Limite de 512MB
docker run --memory=512m fsvxavier/nexs-mcp:latest

# Limite com swap
docker run --memory=512m --memory-swap=1g fsvxavier/nexs-mcp:latest
```

### CPU Limits

```bash
# 2 CPUs
docker run --cpus=2 fsvxavier/nexs-mcp:latest

# CPU shares (relativo)
docker run --cpu-shares=512 fsvxavier/nexs-mcp:latest
```

## Troubleshooting

### Ver logs

```bash
# Logs em tempo real
docker logs -f nexs-mcp-server

# Últimas 100 linhas
docker logs --tail 100 nexs-mcp-server

# Com timestamps
docker logs -t nexs-mcp-server
```

### Debug mode

```bash
docker run -e LOG_LEVEL=debug fsvxavier/nexs-mcp:latest
```

### Shell interativo

```bash
# Executar shell no container
docker exec -it nexs-mcp-server /bin/sh

# Inspecionar filesystem
docker run -it --entrypoint /bin/sh fsvxavier/nexs-mcp:latest
```

### Verificar imagem

```bash
# Informações da imagem
docker inspect fsvxavier/nexs-mcp:latest

# Histórico de layers
docker history fsvxavier/nexs-mcp:latest

# Scan de vulnerabilidades (requer Docker Scout)
docker scout cves fsvxavier/nexs-mcp:latest
```

## Build Local

Para desenvolvimento local:

```bash
# Build simples
docker build -t nexs-mcp:local .

# Build com cache
docker buildx build --cache-from type=local,src=/tmp/buildx-cache -t nexs-mcp:local .

# Build multi-arch
docker buildx build --platform linux/amd64,linux/arm64 -t nexs-mcp:local .
```

## Integration with Claude Desktop

Se usar Claude Desktop com Docker:

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
        "${HOME}/nexs-mcp-data:/app/data",
        "fsvxavier/nexs-mcp:latest"
      ]
    }
  }
}
```

## Production Deployment

### Docker Swarm

```yaml
version: '3.8'
services:
  nexs-mcp:
    image: fsvxavier/nexs-mcp:v1.0.0
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
        max_attempts: 3
    volumes:
      - nexs-mcp-data:/app/data
    networks:
      - nexs-mcp-net

volumes:
  nexs-mcp-data:
    driver: local

networks:
  nexs-mcp-net:
    driver: overlay
```

### Kubernetes

Exemplo de deployment (veja `k8s/` directory para manifests completos):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nexs-mcp
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nexs-mcp
  template:
    metadata:
      labels:
        app: nexs-mcp
    spec:
      containers:
      - name: nexs-mcp
        image: fsvxavier/nexs-mcp:v1.0.0
        resources:
          limits:
            memory: "512Mi"
            cpu: "2"
          requests:
            memory: "128Mi"
            cpu: "0.5"
        volumeMounts:
        - name: data
          mountPath: /app/data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: nexs-mcp-data
```

## Support

- **Issues**: https://github.com/fsvxavier/nexs-mcp/issues
- **Discussions**: https://github.com/fsvxavier/nexs-mcp/discussions
- **Docker Hub**: https://hub.docker.com/r/fsvxavier/nexs-mcp
