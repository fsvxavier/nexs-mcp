# Memory Consolidation - Deployment Guide

**Version:** 1.3.0  
**Last Updated:** December 26, 2025  
**Target Audience:** DevOps Engineers, System Administrators

---

## Table of Contents

1. [Overview](#overview)
2. [System Requirements](#system-requirements)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Production Deployment](#production-deployment)
6. [Monitoring](#monitoring)
7. [Performance Tuning](#performance-tuning)
8. [Backup and Recovery](#backup-and-recovery)
9. [Scaling](#scaling)
10. [Security](#security)
11. [Troubleshooting](#troubleshooting)

---

## Overview

This guide covers deploying NEXS-MCP with Memory Consolidation features in production environments. Memory Consolidation includes 7 services that require specific configuration and resources.

### Services Deployed

| Service | Resource Usage | Critical |
|---------|---------------|----------|
| DuplicateDetectionService | Medium CPU, Low Memory | No |
| ClusteringService | High CPU, Medium Memory | No |
| KnowledgeGraphExtractorService | Medium CPU, Medium Memory | No |
| MemoryConsolidationService | Low CPU, Low Memory | Yes |
| HybridSearchService | High CPU, High Memory | Yes |
| MemoryRetentionService | Low CPU, Low Memory | No |
| ContextEnrichmentService | Medium CPU, Medium Memory | Yes |

**Critical Services** must be running for core functionality. Others can be disabled if needed.

---

## System Requirements

### Minimum Requirements

```
CPU: 2 cores
RAM: 4 GB
Disk: 10 GB (SSD recommended)
OS: Linux (Ubuntu 20.04+), macOS 11+, Windows 10+
Go: 1.21+
```

### Recommended Requirements (Production)

```
CPU: 4-8 cores
RAM: 8-16 GB
Disk: 50 GB SSD
OS: Linux (Ubuntu 22.04 LTS)
Go: 1.21+
Network: 1 Gbps
```

### Resource Scaling

| Memories | Recommended RAM | Recommended CPU | Estimated Processing Time |
|----------|----------------|----------------|---------------------------|
| < 1,000 | 4 GB | 2 cores | 30 seconds |
| 1,000 - 10,000 | 8 GB | 4 cores | 2-5 minutes |
| 10,000 - 50,000 | 16 GB | 8 cores | 10-20 minutes |
| 50,000+ | 32 GB+ | 16 cores+ | 30+ minutes |

---

## Installation

### Option 1: Binary Installation (Recommended)

```bash
# Download latest release
curl -L https://github.com/fsvxavier/nexs-mcp/releases/latest/download/nexs-mcp-linux-amd64 -o nexs-mcp
chmod +x nexs-mcp
sudo mv nexs-mcp /usr/local/bin/

# Verify installation
nexs-mcp --version
```

### Option 2: Build from Source

```bash
# Clone repository
git clone https://github.com/fsvxavier/nexs-mcp.git
cd nexs-mcp

# Build with consolidation features
make build

# Install
sudo make install

# Verify
nexs-mcp --version
```

### Option 3: Docker Deployment

```bash
# Pull image
docker pull fsvxavier/nexs-mcp:1.3.0

# Run with consolidation enabled
docker run -d \
  --name nexs-mcp \
  -v /data/nexs:/data \
  -e NEXS_MEMORY_CONSOLIDATION_ENABLED=true \
  -e NEXS_HYBRID_SEARCH_PERSISTENCE=true \
  -p 8080:8080 \
  fsvxavier/nexs-mcp:1.3.0
```

### Option 4: Kubernetes Deployment

See [Kubernetes Configuration](#kubernetes-configuration) section.

---

## Configuration

### Configuration Files

**Location:** `/etc/nexs-mcp/config.yaml`

```yaml
# config.yaml
memory_consolidation:
  enabled: true
  auto: false
  interval: 24h
  min_memories: 10

duplicate_detection:
  enabled: true
  threshold: 0.95
  min_length: 20
  max_results: 100

clustering:
  enabled: true
  algorithm: dbscan
  min_size: 3
  epsilon: 0.15
  num_clusters: 10

knowledge_graph:
  enabled: true
  extract_people: true
  extract_organizations: true
  extract_keywords: true
  max_keywords: 10
  min_score: 0.3

hybrid_search:
  enabled: true
  mode: auto
  similarity_threshold: 0.7
  max_results: 10
  auto_mode_threshold: 1000
  persistence: true
  index_path: /data/hnsw-index

memory_retention:
  enabled: true
  quality_threshold: 0.5
  high_quality_days: 365
  medium_quality_days: 180
  low_quality_days: 90
  check_interval: 24h
  auto_cleanup: false

embeddings:
  provider: onnx
  model_path: /models/all-minilm-l6-v2
  cache_enabled: true
  cache_size: 10000
  cache_ttl: 24h
```

### Environment Variables

**For systemd:**

Create `/etc/systemd/system/nexs-mcp.service.d/override.conf`:

```ini
[Service]
# Memory Consolidation
Environment="NEXS_MEMORY_CONSOLIDATION_ENABLED=true"
Environment="NEXS_MEMORY_CONSOLIDATION_AUTO=false"
Environment="NEXS_MEMORY_CONSOLIDATION_INTERVAL=24h"

# Duplicate Detection
Environment="NEXS_DUPLICATE_DETECTION_ENABLED=true"
Environment="NEXS_DUPLICATE_DETECTION_THRESHOLD=0.95"

# Clustering
Environment="NEXS_CLUSTERING_ENABLED=true"
Environment="NEXS_CLUSTERING_ALGORITHM=dbscan"

# Knowledge Graph
Environment="NEXS_KNOWLEDGE_GRAPH_ENABLED=true"
Environment="NEXS_KNOWLEDGE_GRAPH_EXTRACT_PEOPLE=true"
Environment="NEXS_KNOWLEDGE_GRAPH_EXTRACT_ORGS=true"

# Hybrid Search
Environment="NEXS_HYBRID_SEARCH_ENABLED=true"
Environment="NEXS_HYBRID_SEARCH_MODE=auto"
Environment="NEXS_HYBRID_SEARCH_PERSISTENCE=true"
Environment="NEXS_HYBRID_SEARCH_INDEX_PATH=/data/hnsw-index"

# Memory Retention
Environment="NEXS_MEMORY_RETENTION_ENABLED=true"
Environment="NEXS_MEMORY_RETENTION_THRESHOLD=0.5"
Environment="NEXS_MEMORY_RETENTION_AUTO_CLEANUP=false"

# Embeddings
Environment="NEXS_EMBEDDINGS_PROVIDER=onnx"
Environment="NEXS_EMBEDDINGS_CACHE_ENABLED=true"
Environment="NEXS_EMBEDDINGS_CACHE_SIZE=10000"
```

### Configuration Profiles

**Development:**
```bash
export NEXS_MEMORY_CONSOLIDATION_AUTO=false
export NEXS_MEMORY_RETENTION_AUTO_CLEANUP=false
export NEXS_EMBEDDINGS_CACHE_SIZE=1000
```

**Staging:**
```bash
export NEXS_MEMORY_CONSOLIDATION_AUTO=true
export NEXS_MEMORY_CONSOLIDATION_INTERVAL=168h
export NEXS_MEMORY_RETENTION_AUTO_CLEANUP=false
export NEXS_EMBEDDINGS_CACHE_SIZE=5000
```

**Production:**
```bash
export NEXS_MEMORY_CONSOLIDATION_AUTO=true
export NEXS_MEMORY_CONSOLIDATION_INTERVAL=24h
export NEXS_MEMORY_RETENTION_AUTO_CLEANUP=true
export NEXS_HYBRID_SEARCH_PERSISTENCE=true
export NEXS_EMBEDDINGS_CACHE_SIZE=10000
```

---

## Production Deployment

### Systemd Service

**1. Create service file:** `/etc/systemd/system/nexs-mcp.service`

```ini
[Unit]
Description=NEXS-MCP Server with Memory Consolidation
After=network.target

[Service]
Type=simple
User=nexs
Group=nexs
WorkingDirectory=/opt/nexs-mcp
ExecStart=/usr/local/bin/nexs-mcp server --config /etc/nexs-mcp/config.yaml
Restart=always
RestartSec=10

# Resource limits
LimitNOFILE=65536
MemoryLimit=8G
CPUQuota=400%

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/data/nexs-mcp /var/log/nexs-mcp

[Install]
WantedBy=multi-user.target
```

**2. Enable and start:**

```bash
sudo systemctl daemon-reload
sudo systemctl enable nexs-mcp
sudo systemctl start nexs-mcp
sudo systemctl status nexs-mcp
```

**3. View logs:**

```bash
sudo journalctl -u nexs-mcp -f
```

### Docker Compose

**docker-compose.yml:**

```yaml
version: '3.8'

services:
  nexs-mcp:
    image: fsvxavier/nexs-mcp:1.3.0
    container_name: nexs-mcp
    restart: always
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./config.yaml:/etc/nexs-mcp/config.yaml:ro
      - ./models:/models:ro
    environment:
      # Memory Consolidation
      - NEXS_MEMORY_CONSOLIDATION_ENABLED=true
      - NEXS_MEMORY_CONSOLIDATION_AUTO=true
      - NEXS_MEMORY_CONSOLIDATION_INTERVAL=24h
      
      # Duplicate Detection
      - NEXS_DUPLICATE_DETECTION_ENABLED=true
      - NEXS_DUPLICATE_DETECTION_THRESHOLD=0.95
      
      # Clustering
      - NEXS_CLUSTERING_ENABLED=true
      - NEXS_CLUSTERING_ALGORITHM=dbscan
      
      # Knowledge Graph
      - NEXS_KNOWLEDGE_GRAPH_ENABLED=true
      - NEXS_KNOWLEDGE_GRAPH_EXTRACT_PEOPLE=true
      - NEXS_KNOWLEDGE_GRAPH_EXTRACT_ORGS=true
      
      # Hybrid Search
      - NEXS_HYBRID_SEARCH_ENABLED=true
      - NEXS_HYBRID_SEARCH_MODE=auto
      - NEXS_HYBRID_SEARCH_PERSISTENCE=true
      - NEXS_HYBRID_SEARCH_INDEX_PATH=/data/hnsw-index
      
      # Memory Retention
      - NEXS_MEMORY_RETENTION_ENABLED=true
      - NEXS_MEMORY_RETENTION_AUTO_CLEANUP=true
      
      # Embeddings
      - NEXS_EMBEDDINGS_PROVIDER=onnx
      - NEXS_EMBEDDINGS_CACHE_ENABLED=true
      - NEXS_EMBEDDINGS_CACHE_SIZE=10000
    
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G
    
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "10"
```

**Start:**

```bash
docker-compose up -d
docker-compose logs -f nexs-mcp
```

### Kubernetes Configuration

**nexs-mcp-deployment.yaml:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nexs-mcp-config
  namespace: nexs
data:
  config.yaml: |
    memory_consolidation:
      enabled: true
      auto: true
      interval: 24h
    
    hybrid_search:
      enabled: true
      mode: auto
      persistence: true
      index_path: /data/hnsw-index
    
    embeddings:
      provider: onnx
      cache_enabled: true
      cache_size: 10000

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nexs-mcp-data
  namespace: nexs
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
  storageClassName: ssd

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nexs-mcp
  namespace: nexs
spec:
  replicas: 1  # Single instance for now (stateful)
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
        image: fsvxavier/nexs-mcp:1.3.0
        ports:
        - containerPort: 8080
          name: http
        
        env:
        - name: NEXS_MEMORY_CONSOLIDATION_ENABLED
          value: "true"
        - name: NEXS_MEMORY_CONSOLIDATION_AUTO
          value: "true"
        - name: NEXS_HYBRID_SEARCH_PERSISTENCE
          value: "true"
        - name: NEXS_EMBEDDINGS_CACHE_ENABLED
          value: "true"
        
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /etc/nexs-mcp
        
        resources:
          limits:
            cpu: "4"
            memory: "8Gi"
          requests:
            cpu: "2"
            memory: "4Gi"
        
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
      
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: nexs-mcp-data
      - name: config
        configMap:
          name: nexs-mcp-config

---
apiVersion: v1
kind: Service
metadata:
  name: nexs-mcp
  namespace: nexs
spec:
  selector:
    app: nexs-mcp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

**Deploy:**

```bash
kubectl create namespace nexs
kubectl apply -f nexs-mcp-deployment.yaml
kubectl get pods -n nexs -w
```

---

## Monitoring

### Metrics to Monitor

| Metric | Alert Threshold | Description |
|--------|----------------|-------------|
| Memory Usage | > 80% | System memory consumption |
| CPU Usage | > 90% for 5m | CPU utilization |
| Consolidation Duration | > 30 minutes | Time to complete consolidation |
| Duplicate Rate | > 15% | Percentage of duplicates found |
| Average Quality Score | < 0.5 | Overall memory quality |
| HNSW Index Size | > 1 GB | Search index disk usage |
| Embedding Cache Hit Rate | < 60% | Cache effectiveness |
| Failed Consolidations | > 0 | Number of errors |

### Prometheus Metrics

**Exposed at:** `http://localhost:8080/metrics`

```promql
# Consolidation metrics
nexs_consolidation_total
nexs_consolidation_duration_seconds
nexs_consolidation_memories_processed
nexs_consolidation_duplicates_found
nexs_consolidation_clusters_created
nexs_consolidation_entities_extracted

# Search metrics
nexs_search_requests_total
nexs_search_duration_seconds
nexs_search_cache_hits_total
nexs_search_cache_misses_total

# Memory metrics
nexs_memory_count
nexs_memory_quality_avg
nexs_memory_retention_deleted_total

# System metrics
nexs_memory_usage_bytes
nexs_cpu_usage_percent
nexs_disk_usage_bytes
```

**Example Prometheus Alerts:**

```yaml
groups:
- name: nexs_mcp_alerts
  rules:
  - alert: HighMemoryUsage
    expr: nexs_memory_usage_bytes / 1024 / 1024 / 1024 > 7
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "NEXS-MCP using > 7GB RAM"
  
  - alert: LowQualityScore
    expr: nexs_memory_quality_avg < 0.5
    for: 1h
    labels:
      severity: info
    annotations:
      summary: "Average memory quality below 0.5"
  
  - alert: ConsolidationFailure
    expr: increase(nexs_consolidation_errors_total[1h]) > 0
    labels:
      severity: critical
    annotations:
      summary: "Consolidation failed"
  
  - alert: HighDuplicateRate
    expr: (nexs_consolidation_duplicates_found / nexs_consolidation_memories_processed) > 0.15
    for: 1d
    labels:
      severity: warning
    annotations:
      summary: "Duplicate rate > 15%"
```

### Grafana Dashboard

**Import dashboard:** `https://grafana.com/grafana/dashboards/nexs-mcp`

**Key Panels:**
- Memory count over time
- Consolidation duration
- Quality score distribution
- Duplicate detection rate
- Cluster distribution
- Search performance
- Cache hit rate
- Resource usage (CPU, RAM, Disk)

### Health Checks

```bash
# Basic health
curl http://localhost:8080/health

# Detailed status
curl http://localhost:8080/status

# Consolidation report
nexs-mcp get_consolidation_report \
  --element-type memory \
  --include-statistics
```

### Logging

**Log Levels:**
- **DEBUG**: Detailed operation logs
- **INFO**: Normal operations
- **WARN**: Non-critical issues
- **ERROR**: Errors requiring attention

**Configure log level:**
```bash
export NEXS_LOG_LEVEL=info
export NEXS_LOG_FORMAT=json  # or text
export NEXS_LOG_OUTPUT=/var/log/nexs-mcp/app.log
```

**Log aggregation with ELK:**

```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/nexs-mcp/*.log
  json.keys_under_root: true
  json.add_error_key: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "nexs-mcp-%{+yyyy.MM.dd}"
```

---

## Performance Tuning

### CPU Optimization

**1. Parallel Processing:**

```bash
# Enable concurrent clustering
export NEXS_CLUSTERING_WORKERS=4

# Enable concurrent duplicate detection
export NEXS_DUPLICATE_DETECTION_WORKERS=4
```

**2. CPU Affinity (Linux):**

```bash
# Pin to specific cores
taskset -c 0-3 nexs-mcp server
```

### Memory Optimization

**1. Embedding Cache:**

```bash
# Increase cache size (more RAM, faster)
export NEXS_EMBEDDINGS_CACHE_SIZE=20000

# Reduce cache size (less RAM, slower)
export NEXS_EMBEDDINGS_CACHE_SIZE=5000
```

**2. Batch Processing:**

```bash
# Process in smaller batches
export NEXS_CLUSTERING_BATCH_SIZE=100
export NEXS_DUPLICATE_DETECTION_BATCH_SIZE=100
```

**3. Memory Limits:**

```bash
# Systemd
MemoryLimit=16G

# Docker
docker run --memory="16g" ...

# Kubernetes
resources:
  limits:
    memory: "16Gi"
```

### Disk I/O Optimization

**1. SSD Storage:**

Use SSD for:
- HNSW index (`/data/hnsw-index`)
- Memory database (`/data/elements/memories`)
- Embedding cache

**2. Index Persistence:**

```bash
# Enable persistent index (faster startups)
export NEXS_HYBRID_SEARCH_PERSISTENCE=true
export NEXS_HYBRID_SEARCH_INDEX_PATH=/fast-ssd/hnsw-index
```

**3. Write-Through Cache:**

```bash
# Enable write-through for database
export NEXS_DB_WRITE_THROUGH=true
```

### Network Optimization

**1. Reduce Network Calls:**

```bash
# Cache embeddings locally
export NEXS_EMBEDDINGS_CACHE_ENABLED=true

# Use local ONNX provider (no API calls)
export NEXS_EMBEDDINGS_PROVIDER=onnx
```

**2. Connection Pooling:**

```bash
# If using external embedding service
export NEXS_EMBEDDINGS_POOL_SIZE=10
export NEXS_EMBEDDINGS_TIMEOUT=30s
```

### Algorithm Tuning

**1. DBSCAN Performance:**

```bash
# Faster but less precise
export NEXS_CLUSTERING_EPSILON=0.20
export NEXS_CLUSTERING_MIN_SIZE=5

# Slower but more precise
export NEXS_CLUSTERING_EPSILON=0.10
export NEXS_CLUSTERING_MIN_SIZE=2
```

**2. HNSW Performance:**

```bash
# Faster search (less accuracy)
export NEXS_HYBRID_SEARCH_M=8
export NEXS_HYBRID_SEARCH_EF_CONSTRUCTION=100

# Slower search (more accuracy)
export NEXS_HYBRID_SEARCH_M=16
export NEXS_HYBRID_SEARCH_EF_CONSTRUCTION=200
```

**3. Duplicate Detection:**

```bash
# Faster (may miss some duplicates)
export NEXS_DUPLICATE_DETECTION_THRESHOLD=0.98

# Slower (catches more duplicates)
export NEXS_DUPLICATE_DETECTION_THRESHOLD=0.90
```

---

## Backup and Recovery

### Backup Strategy

**1. Full Backup (daily):**

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backups/nexs-mcp"
DATE=$(date +%Y%m%d)

# Create backup
nexs-mcp backup_create --name "daily-$DATE" --path "$BACKUP_DIR"

# Verify
nexs-mcp backup_verify --name "daily-$DATE" --path "$BACKUP_DIR"

# Clean old backups (keep 30 days)
find "$BACKUP_DIR" -name "daily-*" -mtime +30 -delete
```

**2. Incremental Backup (hourly):**

```bash
#!/bin/bash
# incremental-backup.sh

BACKUP_DIR="/backups/nexs-mcp-incremental"
TIMESTAMP=$(date +%Y%m%d-%H%M)

# Backup only changed files
rsync -av --delete \
  /data/nexs-mcp/ \
  "$BACKUP_DIR/backup-$TIMESTAMP/"
```

**3. Index Backup (weekly):**

```bash
#!/bin/bash
# backup-index.sh

tar -czf /backups/hnsw-index-$(date +%Y%m%d).tar.gz \
  /data/hnsw-index
```

### Automated Backup (systemd timer)

**backup.service:**

```ini
[Unit]
Description=NEXS-MCP Daily Backup

[Service]
Type=oneshot
ExecStart=/usr/local/bin/nexs-mcp-backup.sh
User=nexs
Group=nexs
```

**backup.timer:**

```ini
[Unit]
Description=NEXS-MCP Daily Backup Timer

[Timer]
OnCalendar=daily
OnCalendar=02:00
Persistent=true

[Install]
WantedBy=timers.target
```

**Enable:**

```bash
sudo systemctl enable backup.timer
sudo systemctl start backup.timer
```

### Restore Procedures

**1. Full Restore:**

```bash
# Stop service
sudo systemctl stop nexs-mcp

# Restore data
nexs-mcp backup_restore --name "daily-20251226" --path "/backups/nexs-mcp"

# Verify
nexs-mcp verify_data

# Start service
sudo systemctl start nexs-mcp
```

**2. Partial Restore (HNSW index only):**

```bash
# Extract index
tar -xzf /backups/hnsw-index-20251226.tar.gz -C /data/

# Restart service (rebuilds if needed)
sudo systemctl restart nexs-mcp
```

**3. Disaster Recovery:**

```bash
# Reinstall NEXS-MCP
curl -L https://github.com/fsvxavier/nexs-mcp/releases/latest/download/nexs-mcp-linux-amd64 -o /usr/local/bin/nexs-mcp

# Restore latest backup
nexs-mcp backup_restore --name "latest" --path "/backups/nexs-mcp"

# Reconfigure
cp /backups/config/config.yaml /etc/nexs-mcp/

# Start service
sudo systemctl start nexs-mcp

# Rebuild indexes
nexs-mcp consolidate_memories --enable-clustering --enable-knowledge-extraction
```

---

## Scaling

### Horizontal Scaling (Multiple Instances)

**Current Status:** ⚠️ Memory consolidation is stateful - horizontal scaling not fully supported yet.

**Workaround:** Use separate instances for different element types:

```
Instance 1: Handles memories
Instance 2: Handles agents
Instance 3: Handles personas
```

**Load Balancer Configuration:**

```nginx
upstream nexs_backend {
    # Sticky sessions required
    ip_hash;
    
    server nexs-01:8080;
    server nexs-02:8080;
    server nexs-03:8080;
}

server {
    listen 80;
    
    location / {
        proxy_pass http://nexs_backend;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Vertical Scaling

**Small (< 1,000 memories):**
```
CPU: 2 cores
RAM: 4 GB
Disk: 10 GB
```

**Medium (1,000-10,000 memories):**
```
CPU: 4 cores
RAM: 8 GB
Disk: 50 GB SSD
```

**Large (10,000-50,000 memories):**
```
CPU: 8 cores
RAM: 16 GB
Disk: 100 GB SSD
```

**Extra Large (50,000+ memories):**
```
CPU: 16 cores
RAM: 32 GB
Disk: 500 GB SSD
```

### Database Sharding (Future)

**Planned for v1.4.0:**

```yaml
# config.yaml
database:
  sharding:
    enabled: true
    strategy: hash  # hash, range, or list
    shards:
      - name: shard-01
        range: [0, 10000]
        host: db-01:5432
      - name: shard-02
        range: [10001, 20000]
        host: db-02:5432
```

---

## Security

### Network Security

**1. TLS/HTTPS:**

```yaml
# config.yaml
server:
  tls:
    enabled: true
    cert_file: /etc/nexs-mcp/tls/cert.pem
    key_file: /etc/nexs-mcp/tls/key.pem
```

**2. Firewall Rules:**

```bash
# Allow only from trusted IPs
sudo ufw allow from 10.0.0.0/24 to any port 8080
sudo ufw enable
```

**3. API Authentication:**

```yaml
# config.yaml
server:
  auth:
    enabled: true
    type: jwt  # or basic, oauth2
    secret: "your-secret-key"
```

### Data Security

**1. Encryption at Rest:**

```bash
# Encrypt data directory
cryptsetup luksFormat /dev/sdb1
cryptsetup luksOpen /dev/sdb1 nexs-data
mkfs.ext4 /dev/mapper/nexs-data
mount /dev/mapper/nexs-data /data/nexs-mcp
```

**2. Encryption in Transit:**

Already covered by TLS/HTTPS.

**3. Sensitive Data Masking:**

```yaml
# config.yaml
privacy:
  enabled: true
  mask_pii: true  # Masks emails, phone numbers, SSNs
  mask_patterns:
    - email
    - phone
    - ssn
    - credit_card
```

### Access Control

**1. Role-Based Access Control (RBAC):**

```yaml
# config.yaml
rbac:
  enabled: true
  roles:
    - name: admin
      permissions: ["*"]
    - name: operator
      permissions: ["read", "consolidate"]
    - name: readonly
      permissions: ["read"]
```

**2. Audit Logging:**

```bash
# Enable audit logs
export NEXS_AUDIT_ENABLED=true
export NEXS_AUDIT_LOG=/var/log/nexs-mcp/audit.log
```

### Security Hardening

**1. Run as non-root:**

```bash
# Create dedicated user
sudo useradd -r -s /bin/false nexs
sudo chown -R nexs:nexs /data/nexs-mcp
```

**2. Limit file permissions:**

```bash
sudo chmod 600 /etc/nexs-mcp/config.yaml
sudo chmod 700 /data/nexs-mcp
```

**3. Security scanning:**

```bash
# Scan for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

---

## Troubleshooting

### High Memory Usage

**Symptoms:**
- OOM kills
- Slow performance
- Swap usage

**Solutions:**

```bash
# 1. Reduce cache size
export NEXS_EMBEDDINGS_CACHE_SIZE=5000

# 2. Enable batch processing
export NEXS_CLUSTERING_BATCH_SIZE=50

# 3. Increase system limits
echo "vm.overcommit_memory=1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# 4. Monitor and restart if needed
#!/bin/bash
while true; do
  MEM=$(ps aux | grep nexs-mcp | awk '{print $4}')
  if (( $(echo "$MEM > 80" | bc -l) )); then
    sudo systemctl restart nexs-mcp
  fi
  sleep 60
done
```

### Consolidation Timeouts

**Symptoms:**
- Consolidation takes > 30 minutes
- Timeouts in logs

**Solutions:**

```bash
# 1. Process in smaller chunks
nexs-mcp consolidate_memories --max-elements 1000

# 2. Disable expensive features temporarily
export NEXS_KNOWLEDGE_GRAPH_ENABLED=false

# 3. Increase timeout
export NEXS_CONSOLIDATION_TIMEOUT=3600s  # 1 hour

# 4. Use linear search for accuracy (slower but works)
export NEXS_HYBRID_SEARCH_MODE=linear
```

### Index Corruption

**Symptoms:**
- Search errors
- "index not found" errors

**Solutions:**

```bash
# 1. Rebuild index
rm -rf /data/hnsw-index
nexs-mcp consolidate_memories --enable-clustering

# 2. Restore from backup
tar -xzf /backups/hnsw-index-20251226.tar.gz -C /data/

# 3. Verify index integrity
nexs-mcp verify_index --path /data/hnsw-index
```

### Service Won't Start

**Symptoms:**
- systemctl start fails
- Exit code 1 or 2

**Solutions:**

```bash
# 1. Check logs
sudo journalctl -u nexs-mcp -n 100 --no-pager

# 2. Verify configuration
nexs-mcp validate_config --config /etc/nexs-mcp/config.yaml

# 3. Check file permissions
ls -la /data/nexs-mcp
sudo chown -R nexs:nexs /data/nexs-mcp

# 4. Test manually
sudo -u nexs nexs-mcp server --config /etc/nexs-mcp/config.yaml
```

---

## Maintenance Tasks

### Daily

```bash
# Check service status
sudo systemctl status nexs-mcp

# Check disk usage
df -h /data/nexs-mcp

# Quick consolidation report
nexs-mcp get_consolidation_report --element-type memory
```

### Weekly

```bash
# Full consolidation
nexs-mcp consolidate_memories --no-dry-run

# Backup
/usr/local/bin/nexs-mcp-backup.sh

# Review logs
sudo journalctl -u nexs-mcp --since "1 week ago" | grep ERROR
```

### Monthly

```bash
# Deep cleanup
nexs-mcp apply_retention_policy --no-dry-run

# Update software
sudo apt update && sudo apt upgrade nexs-mcp

# Review metrics and tune configuration
```

---

## Related Documentation

- [Memory Consolidation User Guide](../user-guide/MEMORY_CONSOLIDATION.md)
- [Memory Consolidation Developer Guide](../development/MEMORY_CONSOLIDATION.md)
- [MCP Tools Reference](../api/MCP_TOOLS.md)

---

**Version:** 1.3.0  
**Last Updated:** December 26, 2025  
**Maintainer:** NEXS-MCP Team
