# Memory Consolidation - Integration Guide

**Version:** 1.3.0  
**Last Updated:** December 26, 2025  
**Target Audience:** Integration Engineers, Solution Architects

---

## Table of Contents

1. [Overview](#overview)
2. [Integration Patterns](#integration-patterns)
3. [MCP Protocol Integration](#mcp-protocol-integration)
4. [API Integration](#api-integration)
5. [SDK Integration](#sdk-integration)
6. [Event-Driven Integration](#event-driven-integration)
7. [Third-Party Integrations](#third-party-integrations)
8. [Testing Integration](#testing-integration)
9. [Examples](#examples)
10. [Best Practices](#best-practices)

---

## Overview

This guide explains how to integrate NEXS-MCP Memory Consolidation services with your existing systems, applications, and workflows.

### Integration Approaches

| Approach | Use Case | Complexity | Performance |
|----------|----------|------------|-------------|
| **MCP Protocol** | AI agents (Claude, GPT) | Low | High |
| **REST API** | Web applications | Low | Medium |
| **Go SDK** | Go applications | Medium | High |
| **CLI Wrapper** | Scripts, automation | Low | Medium |
| **Events** | Async workflows | High | High |
| **Webhooks** | External notifications | Medium | Medium |

---

## Integration Patterns

### Pattern 1: Request-Response

**Best for:** Interactive queries, on-demand consolidation

```
Client → Request → NEXS-MCP → Response → Client
```

**Example:**
```bash
# Client makes request
curl -X POST http://nexs-mcp:8080/api/v1/consolidate \
  -H "Content-Type: application/json" \
  -d '{"element_type": "memory"}'

# NEXS-MCP processes
# Returns result immediately
```

### Pattern 2: Async Task Queue

**Best for:** Long-running consolidation, background processing

```
Client → Submit Task → Queue → Worker → Result Store
                                   ↓
Client ← Poll Status ← Status API
```

**Example:**
```bash
# Submit task
task_id=$(curl -X POST http://nexs-mcp:8080/api/v1/consolidate/async \
  -d '{"element_type": "memory"}' | jq -r '.task_id')

# Poll status
while true; do
  status=$(curl http://nexs-mcp:8080/api/v1/tasks/$task_id | jq -r '.status')
  [[ "$status" == "completed" ]] && break
  sleep 5
done
```

### Pattern 3: Event-Driven

**Best for:** Real-time updates, microservices

```
NEXS-MCP → Event Bus → Subscribers
             ↓
        (Kafka, NATS, RabbitMQ)
```

**Example:**
```go
// Subscribe to consolidation events
consumer.Subscribe("nexs.consolidation.completed", func(event Event) {
    log.Printf("Consolidation completed: %v", event.Data)
})
```

### Pattern 4: Batch Processing

**Best for:** Bulk operations, scheduled tasks

```
Scheduler → Batch Job → NEXS-MCP (multiple requests) → Results
```

**Example:**
```bash
# Process 1000 memories in batches of 100
for i in {1..10}; do
  nexs-mcp consolidate_memories \
    --offset $((i*100)) \
    --limit 100 \
    --no-dry-run &
done
wait
```

---

## MCP Protocol Integration

### Claude Desktop Integration

**1. Configure Claude Desktop:**

`~/.config/claude/config.json`:

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "/usr/local/bin/nexs-mcp",
      "args": ["serve"],
      "env": {
        "NEXS_MEMORY_CONSOLIDATION_ENABLED": "true",
        "NEXS_HYBRID_SEARCH_ENABLED": "true",
        "NEXS_EMBEDDINGS_PROVIDER": "onnx"
      }
    }
  }
}
```

**2. Use in Claude:**

```
You: Consolidate my memories and show me the report

Claude: [Uses consolidate_memories tool]
        [Uses get_consolidation_report tool]
        
        Here's your consolidation report:
        - 45 memories processed
        - 7 duplicates found
        - 8 clusters created
        - Average quality: 0.72
```

### Custom MCP Client

**Go Implementation:**

```go
package main

import (
    "context"
    "encoding/json"
    "github.com/fsvxavier/nexs-mcp/pkg/mcp"
)

func main() {
    // Create MCP client
    client := mcp.NewClient("http://localhost:8080")
    
    // Call consolidate_memories tool
    result, err := client.CallTool(context.Background(), mcp.ToolRequest{
        Name: "consolidate_memories",
        Arguments: map[string]interface{}{
            "element_type":                 "memory",
            "enable_duplicate_detection":   true,
            "enable_clustering":            true,
            "enable_knowledge_extraction":  true,
            "dry_run":                      false,
        },
    })
    
    if err != nil {
        panic(err)
    }
    
    // Parse result
    var report ConsolidationReport
    json.Unmarshal(result.Content, &report)
    
    fmt.Printf("Processed %d memories in %dms\n", 
        report.MemoriesProcessed, 
        report.DurationMS)
}
```

**Python Implementation:**

```python
import requests

class NexsMCPClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
    
    def consolidate_memories(self, dry_run=True):
        """Call consolidate_memories MCP tool"""
        response = requests.post(
            f"{self.base_url}/mcp/tools/call",
            json={
                "name": "consolidate_memories",
                "arguments": {
                    "element_type": "memory",
                    "enable_duplicate_detection": True,
                    "enable_clustering": True,
                    "dry_run": dry_run
                }
            }
        )
        return response.json()
    
    def get_report(self):
        """Get consolidation report"""
        response = requests.post(
            f"{self.base_url}/mcp/tools/call",
            json={
                "name": "get_consolidation_report",
                "arguments": {
                    "element_type": "memory",
                    "include_statistics": True
                }
            }
        )
        return response.json()

# Usage
client = NexsMCPClient()
result = client.consolidate_memories(dry_run=False)
print(f"Duplicates found: {result['duplicates_found']}")

report = client.get_report()
print(f"Average quality: {report['statistics']['avg_quality']}")
```

---

## API Integration

### REST API

**Base URL:** `http://localhost:8080/api/v1`

**Authentication:**

```bash
# API Key
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/consolidate

# JWT
curl -H "Authorization: Bearer your-jwt-token" \
  http://localhost:8080/api/v1/consolidate
```

### Endpoints

#### POST /api/v1/consolidate

**Request:**
```json
{
  "element_type": "memory",
  "options": {
    "enable_duplicate_detection": true,
    "enable_clustering": true,
    "enable_knowledge_extraction": true,
    "enable_quality_scoring": true,
    "dry_run": false
  },
  "filters": {
    "tags": ["project-alpha"],
    "date_from": "2025-11-01",
    "date_to": "2025-12-26"
  }
}
```

**Response:**
```json
{
  "workflow_id": "consolidation-20251226-001",
  "status": "completed",
  "duration_ms": 3456,
  "results": {
    "memories_processed": 45,
    "duplicates_found": 7,
    "clusters_created": 8,
    "entities_extracted": 52,
    "avg_quality": 0.72
  },
  "recommendations": [
    "7 duplicate groups found - consider merging",
    "5 low-quality memories - consider removing"
  ]
}
```

#### GET /api/v1/consolidation/report

**Response:**
```json
{
  "total_memories": 45,
  "duplicate_rate": 0.156,
  "clustering": {
    "num_clusters": 8,
    "outliers": 3,
    "silhouette_score": 0.67
  },
  "quality": {
    "average": 0.72,
    "high_count": 25,
    "medium_count": 15,
    "low_count": 5
  },
  "knowledge_graph": {
    "entities": 52,
    "relationships": 28,
    "people": 12,
    "organizations": 8
  }
}
```

#### POST /api/v1/duplicates/detect

**Request:**
```json
{
  "element_type": "memory",
  "similarity_threshold": 0.95,
  "auto_merge": false
}
```

#### POST /api/v1/clusters/create

**Request:**
```json
{
  "algorithm": "dbscan",
  "epsilon": 0.15,
  "min_cluster_size": 3
}
```

#### POST /api/v1/search/hybrid

**Request:**
```json
{
  "query": "machine learning implementation",
  "mode": "auto",
  "similarity_threshold": 0.7,
  "max_results": 10
}
```

### Client Libraries

**JavaScript/TypeScript:**

```typescript
import { NexsMCPClient } from '@nexs-mcp/client';

const client = new NexsMCPClient({
  baseURL: 'http://localhost:8080',
  apiKey: process.env.NEXS_API_KEY
});

// Consolidate memories
const result = await client.consolidate({
  elementType: 'memory',
  enableDuplicateDetection: true,
  enableClustering: true,
  dryRun: false
});

console.log(`Found ${result.duplicates_found} duplicates`);
console.log(`Created ${result.clusters_created} clusters`);

// Search with hybrid mode
const searchResults = await client.hybridSearch({
  query: 'project planning',
  mode: 'auto',
  maxResults: 10
});

searchResults.forEach(result => {
  console.log(`${result.name}: ${result.similarity}`);
});
```

**Python:**

```python
from nexs_mcp import Client

client = Client(
    base_url='http://localhost:8080',
    api_key=os.getenv('NEXS_API_KEY')
)

# Consolidate memories
result = client.consolidate(
    element_type='memory',
    enable_duplicate_detection=True,
    enable_clustering=True,
    dry_run=False
)

print(f"Found {result.duplicates_found} duplicates")
print(f"Created {result.clusters_created} clusters")

# Search
results = client.hybrid_search(
    query='project planning',
    mode='auto',
    max_results=10
)

for r in results:
    print(f"{r.name}: {r.similarity}")
```

---

## SDK Integration

### Go SDK

**Installation:**

```bash
go get github.com/fsvxavier/nexs-mcp/pkg/client
```

**Usage:**

```go
package main

import (
    "context"
    "github.com/fsvxavier/nexs-mcp/pkg/client"
)

func main() {
    // Create client
    c := client.New(client.Config{
        BaseURL: "http://localhost:8080",
        APIKey:  "your-api-key",
    })
    
    ctx := context.Background()
    
    // Consolidate memories
    result, err := c.ConsolidateMemories(ctx, &client.ConsolidateRequest{
        ElementType:              "memory",
        EnableDuplicateDetection: true,
        EnableClustering:         true,
        EnableKnowledgeExtraction: true,
        DryRun:                   false,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Workflow ID: %s", result.WorkflowID)
    log.Printf("Duplicates: %d", result.DuplicatesFound)
    log.Printf("Clusters: %d", result.ClustersCreated)
    
    // Get report
    report, err := c.GetConsolidationReport(ctx, &client.ReportRequest{
        ElementType:        "memory",
        IncludeStatistics:  true,
        IncludeRecommendations: true,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Total memories: %d", report.TotalMemories)
    log.Printf("Avg quality: %.2f", report.Quality.Average)
    
    // Hybrid search
    searchResults, err := c.HybridSearch(ctx, &client.SearchRequest{
        Query:               "machine learning",
        Mode:                "auto",
        SimilarityThreshold: 0.7,
        MaxResults:          10,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    for _, result := range searchResults {
        log.Printf("%s: %.2f", result.Name, result.Similarity)
    }
}
```

### Advanced SDK Usage

**Batch Processing:**

```go
func consolidateBatch(c *client.Client, elementIDs []string) error {
    ctx := context.Background()
    
    // Process in batches of 100
    batchSize := 100
    for i := 0; i < len(elementIDs); i += batchSize {
        end := i + batchSize
        if end > len(elementIDs) {
            end = len(elementIDs)
        }
        
        batch := elementIDs[i:end]
        
        _, err := c.ConsolidateMemories(ctx, &client.ConsolidateRequest{
            ElementType:    "memory",
            ElementIDs:     batch,
            EnableClustering: true,
            DryRun:         false,
        })
        
        if err != nil {
            return fmt.Errorf("batch %d failed: %w", i/batchSize, err)
        }
        
        log.Printf("Processed batch %d/%d", i/batchSize+1, len(elementIDs)/batchSize)
    }
    
    return nil
}
```

**Async Processing:**

```go
func consolidateAsync(c *client.Client) error {
    ctx := context.Background()
    
    // Start async consolidation
    task, err := c.ConsolidateMemoriesAsync(ctx, &client.ConsolidateRequest{
        ElementType: "memory",
        EnableDuplicateDetection: true,
        EnableClustering: true,
        DryRun: false,
    })
    
    if err != nil {
        return err
    }
    
    log.Printf("Task ID: %s", task.ID)
    
    // Poll for completion
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            status, err := c.GetTaskStatus(ctx, task.ID)
            if err != nil {
                return err
            }
            
            log.Printf("Status: %s (%d%%)", status.State, status.Progress)
            
            if status.State == "completed" {
                log.Printf("Completed in %dms", status.DurationMS)
                return nil
            }
            
            if status.State == "failed" {
                return fmt.Errorf("task failed: %s", status.Error)
            }
        
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

---

## Event-Driven Integration

### Event Types

```go
const (
    // Consolidation events
    EventConsolidationStarted   = "nexs.consolidation.started"
    EventConsolidationCompleted = "nexs.consolidation.completed"
    EventConsolidationFailed    = "nexs.consolidation.failed"
    
    // Duplicate detection events
    EventDuplicatesDetected = "nexs.duplicates.detected"
    EventDuplicatesMerged   = "nexs.duplicates.merged"
    
    // Clustering events
    EventClusteringCompleted = "nexs.clustering.completed"
    EventClusterCreated      = "nexs.cluster.created"
    
    // Knowledge graph events
    EventEntitiesExtracted = "nexs.knowledge.entities_extracted"
    EventRelationshipsCreated = "nexs.knowledge.relationships_created"
    
    // Retention events
    EventMemoriesDeleted = "nexs.retention.memories_deleted"
)
```

### Kafka Integration

**Producer (NEXS-MCP):**

```go
// internal/infrastructure/events/kafka_producer.go
package events

import (
    "github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
    writer *kafka.Writer
}

func NewKafkaProducer(brokers []string) *KafkaProducer {
    return &KafkaProducer{
        writer: &kafka.Writer{
            Addr:     kafka.TCP(brokers...),
            Topic:    "nexs-events",
            Balancer: &kafka.LeastBytes{},
        },
    }
}

func (p *KafkaProducer) PublishConsolidationCompleted(ctx context.Context, data ConsolidationData) error {
    event := Event{
        Type:      EventConsolidationCompleted,
        Timestamp: time.Now(),
        Data:      data,
    }
    
    payload, _ := json.Marshal(event)
    
    return p.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(data.WorkflowID),
        Value: payload,
    })
}
```

**Consumer (Your Application):**

```go
package main

import (
    "github.com/segmentio/kafka-go"
)

func main() {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{"localhost:9092"},
        Topic:   "nexs-events",
        GroupID: "my-app",
    })
    
    for {
        msg, err := reader.ReadMessage(context.Background())
        if err != nil {
            log.Fatal(err)
        }
        
        var event Event
        json.Unmarshal(msg.Value, &event)
        
        switch event.Type {
        case "nexs.consolidation.completed":
            handleConsolidationCompleted(event.Data)
        case "nexs.duplicates.detected":
            handleDuplicatesDetected(event.Data)
        case "nexs.cluster.created":
            handleClusterCreated(event.Data)
        }
    }
}

func handleConsolidationCompleted(data interface{}) {
    log.Printf("Consolidation completed: %v", data)
    // Send notification, update dashboard, etc.
}
```

### NATS Integration

**Publisher:**

```go
func publishEvent(nc *nats.Conn, eventType string, data interface{}) error {
    payload, _ := json.Marshal(Event{
        Type:      eventType,
        Timestamp: time.Now(),
        Data:      data,
    })
    
    return nc.Publish("nexs.events."+eventType, payload)
}
```

**Subscriber:**

```go
func subscribeToEvents(nc *nats.Conn) {
    // Subscribe to all consolidation events
    nc.Subscribe("nexs.events.consolidation.*", func(msg *nats.Msg) {
        var event Event
        json.Unmarshal(msg.Data, &event)
        
        log.Printf("Received event: %s", event.Type)
        handleEvent(event)
    })
    
    // Subscribe to specific events
    nc.Subscribe("nexs.events.duplicates.detected", handleDuplicates)
    nc.Subscribe("nexs.events.cluster.created", handleCluster)
}
```

---

## Third-Party Integrations

### Slack Notifications

```go
package integrations

import (
    "github.com/slack-go/slack"
)

type SlackNotifier struct {
    client *slack.Client
    channel string
}

func (n *SlackNotifier) NotifyConsolidationCompleted(report ConsolidationReport) error {
    attachment := slack.Attachment{
        Color: "good",
        Title: "Memory Consolidation Completed",
        Fields: []slack.AttachmentField{
            {
                Title: "Memories Processed",
                Value: fmt.Sprintf("%d", report.MemoriesProcessed),
                Short: true,
            },
            {
                Title: "Duplicates Found",
                Value: fmt.Sprintf("%d", report.DuplicatesFound),
                Short: true,
            },
            {
                Title: "Clusters Created",
                Value: fmt.Sprintf("%d", report.ClustersCreated),
                Short: true,
            },
            {
                Title: "Average Quality",
                Value: fmt.Sprintf("%.2f", report.AvgQuality),
                Short: true,
            },
        },
    }
    
    _, _, err := n.client.PostMessage(
        n.channel,
        slack.MsgOptionAttachments(attachment),
    )
    
    return err
}
```

### Prometheus Metrics Export

```go
package integrations

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    consolidationDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "nexs_consolidation_duration_seconds",
        Help: "Time taken to consolidate memories",
    })
    
    duplicatesFound = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "nexs_duplicates_found_total",
        Help: "Total number of duplicates found",
    })
    
    avgQuality = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "nexs_memory_quality_avg",
        Help: "Average memory quality score",
    })
)

func init() {
    prometheus.MustRegister(consolidationDuration)
    prometheus.MustRegister(duplicatesFound)
    prometheus.MustRegister(avgQuality)
}

func RecordConsolidationMetrics(report ConsolidationReport) {
    consolidationDuration.Observe(float64(report.DurationMS) / 1000.0)
    duplicatesFound.Add(float64(report.DuplicatesFound))
    avgQuality.Set(report.AvgQuality)
}
```

### Grafana Dashboard Integration

```bash
# Export metrics to Grafana
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -H "Content-Type: application/json" \
  -d @grafana-dashboard.json
```

### Elasticsearch Integration

```go
package integrations

import (
    "github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchIndexer struct {
    client *elasticsearch.Client
}

func (i *ElasticsearchIndexer) IndexConsolidationReport(report ConsolidationReport) error {
    doc := map[string]interface{}{
        "timestamp":          time.Now(),
        "workflow_id":        report.WorkflowID,
        "memories_processed": report.MemoriesProcessed,
        "duplicates_found":   report.DuplicatesFound,
        "clusters_created":   report.ClustersCreated,
        "avg_quality":        report.AvgQuality,
    }
    
    body, _ := json.Marshal(doc)
    
    _, err := i.client.Index(
        "nexs-consolidation-reports",
        bytes.NewReader(body),
    )
    
    return err
}
```

---

## Testing Integration

### Integration Tests

```go
// integration_test.go
package integration_test

import (
    "testing"
    "github.com/fsvxavier/nexs-mcp/pkg/client"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestConsolidationIntegration(t *testing.T) {
    // Setup test client
    c := client.New(client.Config{
        BaseURL: "http://localhost:8080",
        APIKey:  "test-api-key",
    })
    
    ctx := context.Background()
    
    // Test consolidation
    t.Run("Consolidate Memories", func(t *testing.T) {
        result, err := c.ConsolidateMemories(ctx, &client.ConsolidateRequest{
            ElementType:               "memory",
            EnableDuplicateDetection:  true,
            EnableClustering:          true,
            DryRun:                    true,
        })
        
        require.NoError(t, err)
        assert.NotEmpty(t, result.WorkflowID)
        assert.GreaterOrEqual(t, result.MemoriesProcessed, 0)
    })
    
    // Test duplicate detection
    t.Run("Detect Duplicates", func(t *testing.T) {
        result, err := c.DetectDuplicates(ctx, &client.DuplicateRequest{
            ElementType:         "memory",
            SimilarityThreshold: 0.95,
            AutoMerge:           false,
        })
        
        require.NoError(t, err)
        assert.GreaterOrEqual(t, len(result.DuplicateGroups), 0)
    })
    
    // Test search
    t.Run("Hybrid Search", func(t *testing.T) {
        results, err := c.HybridSearch(ctx, &client.SearchRequest{
            Query:      "test query",
            Mode:       "auto",
            MaxResults: 10,
        })
        
        require.NoError(t, err)
        assert.LessOrEqual(t, len(results), 10)
    })
}
```

### Mock Server for Testing

```go
// mock_server_test.go
package integration_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func newMockServer(t *testing.T) *httptest.Server {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/api/v1/consolidate":
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{
                "workflow_id": "test-workflow-123",
                "status": "completed",
                "memories_processed": 100,
                "duplicates_found": 5,
                "clusters_created": 8
            }`))
        
        case "/api/v1/duplicates/detect":
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{
                "duplicate_groups": [
                    {
                        "elements": ["mem-1", "mem-2"],
                        "similarity": 0.98
                    }
                ]
            }`))
        
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    })
    
    return httptest.NewServer(handler)
}

func TestWithMockServer(t *testing.T) {
    server := newMockServer(t)
    defer server.Close()
    
    c := client.New(client.Config{
        BaseURL: server.URL,
    })
    
    // Test against mock server
    result, err := c.ConsolidateMemories(context.Background(), &client.ConsolidateRequest{
        ElementType: "memory",
        DryRun:      true,
    })
    
    require.NoError(t, err)
    assert.Equal(t, "test-workflow-123", result.WorkflowID)
}
```

---

## Examples

### Example 1: Automated Weekly Consolidation

```bash
#!/bin/bash
# weekly-consolidation.sh

set -e

# Configuration
NEXS_URL="http://localhost:8080"
SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK"

# Run consolidation
echo "Starting weekly consolidation..."
result=$(curl -s -X POST "$NEXS_URL/api/v1/consolidate" \
  -H "Content-Type: application/json" \
  -d '{
    "element_type": "memory",
    "options": {
      "enable_duplicate_detection": true,
      "enable_clustering": true,
      "enable_knowledge_extraction": true,
      "dry_run": false
    }
  }')

# Extract metrics
workflow_id=$(echo "$result" | jq -r '.workflow_id')
duplicates=$(echo "$result" | jq -r '.results.duplicates_found')
clusters=$(echo "$result" | jq -r '.results.clusters_created')
quality=$(echo "$result" | jq -r '.results.avg_quality')

# Send notification to Slack
curl -X POST "$SLACK_WEBHOOK" \
  -H "Content-Type: application/json" \
  -d "{
    \"text\": \"Weekly Consolidation Completed\",
    \"attachments\": [{
      \"color\": \"good\",
      \"fields\": [
        {\"title\": \"Workflow ID\", \"value\": \"$workflow_id\", \"short\": true},
        {\"title\": \"Duplicates Found\", \"value\": \"$duplicates\", \"short\": true},
        {\"title\": \"Clusters Created\", \"value\": \"$clusters\", \"short\": true},
        {\"title\": \"Avg Quality\", \"value\": \"$quality\", \"short\": true}
      ]
    }]
  }"

echo "Consolidation completed: workflow_id=$workflow_id"
```

### Example 2: Real-Time Duplicate Detection

```python
# real_time_duplicate_detector.py

import time
import requests
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

class MemoryWatcher(FileSystemEventHandler):
    def __init__(self, nexs_url):
        self.nexs_url = nexs_url
    
    def on_created(self, event):
        if event.is_directory:
            return
        
        # New memory created, check for duplicates
        self.check_duplicates()
    
    def check_duplicates(self):
        response = requests.post(
            f"{self.nexs_url}/api/v1/duplicates/detect",
            json={
                "element_type": "memory",
                "similarity_threshold": 0.95,
                "auto_merge": False
            }
        )
        
        result = response.json()
        
        if result.get('duplicate_groups'):
            print(f"Warning: {len(result['duplicate_groups'])} duplicate groups found!")
            for group in result['duplicate_groups']:
                print(f"  - {group['elements']} (similarity: {group['similarity']})")

# Watch memory directory
observer = Observer()
observer.schedule(
    MemoryWatcher("http://localhost:8080"),
    path="/data/nexs-mcp/elements/memories",
    recursive=False
)
observer.start()

try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    observer.stop()

observer.join()
```

### Example 3: Knowledge Graph Export

```python
# export_knowledge_graph.py

import requests
import networkx as nx
import matplotlib.pyplot as plt

def export_knowledge_graph():
    # Get knowledge graph from NEXS-MCP
    response = requests.post(
        "http://localhost:8080/mcp/tools/call",
        json={
            "name": "extract_knowledge_graph",
            "arguments": {
                "element_type": "memory",
                "extract_relationships": True
            }
        }
    )
    
    data = response.json()
    
    # Create NetworkX graph
    G = nx.Graph()
    
    # Add nodes (entities)
    for entity in data['entities']:
        G.add_node(entity['id'], 
                   type=entity['type'],
                   name=entity['name'])
    
    # Add edges (relationships)
    for rel in data['relationships']:
        G.add_edge(rel['source'], 
                   rel['target'],
                   type=rel['type'],
                   weight=rel['strength'])
    
    # Export to various formats
    nx.write_gexf(G, "knowledge_graph.gexf")  # For Gephi
    nx.write_graphml(G, "knowledge_graph.graphml")  # For yEd
    
    # Visualize
    plt.figure(figsize=(15, 15))
    pos = nx.spring_layout(G)
    nx.draw(G, pos, with_labels=True, node_color='lightblue', 
            node_size=1500, font_size=10, font_weight='bold')
    plt.savefig("knowledge_graph.png", dpi=300, bbox_inches='tight')
    
    print(f"Exported graph with {G.number_of_nodes()} nodes and {G.number_of_edges()} edges")

if __name__ == "__main__":
    export_knowledge_graph()
```

---

## Best Practices

### 1. Use Dry Run First

Always preview changes:

```python
# Dry run
result = client.consolidate(dry_run=True)
print(f"Would process {result.memories_processed} memories")
print(f"Would find {result.duplicates_found} duplicates")

# Confirm with user
if input("Apply changes? (y/n): ").lower() == 'y':
    # Real run
    result = client.consolidate(dry_run=False)
```

### 2. Handle Errors Gracefully

```go
result, err := client.ConsolidateMemories(ctx, request)
if err != nil {
    if errors.Is(err, client.ErrTimeout) {
        // Retry with smaller batch
        return retryWithSmallerBatch(request)
    } else if errors.Is(err, client.ErrRateLimit) {
        // Wait and retry
        time.Sleep(time.Minute)
        return retryConsolidation(request)
    } else {
        // Log and alert
        log.Error("Consolidation failed", "error", err)
        alertOps(err)
        return err
    }
}
```

### 3. Implement Retry Logic

```python
import tenacity

@tenacity.retry(
    wait=tenacity.wait_exponential(multiplier=1, min=4, max=60),
    stop=tenacity.stop_after_attempt(3),
    retry=tenacity.retry_if_exception_type(requests.exceptions.Timeout)
)
def consolidate_with_retry(client):
    return client.consolidate(dry_run=False)
```

### 4. Monitor Integration Health

```go
func checkIntegrationHealth() error {
    // Check NEXS-MCP availability
    resp, err := http.Get("http://localhost:8080/health")
    if err != nil || resp.StatusCode != 200 {
        return fmt.Errorf("NEXS-MCP unhealthy")
    }
    
    // Check consolidation status
    report, err := client.GetConsolidationReport(ctx, &client.ReportRequest{
        ElementType: "memory",
    })
    if err != nil {
        return fmt.Errorf("failed to get report: %w", err)
    }
    
    // Alert if quality drops
    if report.Quality.Average < 0.5 {
        alertOps("Memory quality below threshold")
    }
    
    return nil
}
```

### 5. Use Circuit Breakers

```go
import "github.com/sony/gobreaker"

var breaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "nexs-mcp",
    MaxRequests: 3,
    Interval:    time.Minute,
    Timeout:     time.Minute * 5,
})

func consolidateWithCircuitBreaker(client *client.Client) error {
    _, err := breaker.Execute(func() (interface{}, error) {
        return client.ConsolidateMemories(ctx, request)
    })
    return err
}
```

---

## Related Documentation

- [Memory Consolidation User Guide](../user-guide/MEMORY_CONSOLIDATION.md)
- [Memory Consolidation Developer Guide](../development/MEMORY_CONSOLIDATION.md)
- [Deployment Guide](DEPLOYMENT.md)
- [MCP Tools Reference](../api/MCP_TOOLS.md)

---

**Version:** 1.3.0  
**Last Updated:** December 26, 2025  
**Maintainer:** NEXS-MCP Team
