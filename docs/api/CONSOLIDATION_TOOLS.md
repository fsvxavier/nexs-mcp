# Memory Consolidation Tools Reference

**Version:** v1.3.0  
**Category:** Memory Management  
**Last Updated:** December 26, 2025

## Overview

Memory Consolidation tools provide advanced capabilities for managing, organizing, and optimizing memories in NEXS-MCP. These tools implement state-of-the-art algorithms for duplicate detection, clustering, knowledge extraction, quality scoring, and retention policies.

**Key Features:**
- 🔍 **Duplicate Detection**: HNSW-based similarity search with configurable thresholds
- 📊 **Clustering**: DBSCAN and K-means algorithms for grouping related memories
- 🧠 **Knowledge Graphs**: NLP-based entity and relationship extraction
- ⚡ **Hybrid Search**: Automatic HNSW/linear mode selection
- 🎯 **Quality Scoring**: Multi-factor memory quality assessment
- 🗑️ **Retention Policies**: Quality-based cleanup with configurable retention periods
- 🔗 **Context Enrichment**: Relationship traversal and temporal analysis
- 🤖 **Automated Workflows**: End-to-end consolidation orchestration

## Table of Contents

1. [Configuration](#configuration)
2. [Tool Reference](#tool-reference)
   - [consolidate_memories](#consolidate_memories)
   - [detect_duplicates](#detect_duplicates)
   - [cluster_memories](#cluster_memories)
   - [extract_knowledge_graph](#extract_knowledge_graph)
   - [get_consolidation_report](#get_consolidation_report)
   - [hybrid_search](#hybrid_search)
   - [score_memory_quality](#score_memory_quality)
   - [apply_retention_policy](#apply_retention_policy)
   - [enrich_context](#enrich_context)
3. [Workflow Examples](#workflow-examples)
4. [Best Practices](#best-practices)
5. [Troubleshooting](#troubleshooting)

---

## Configuration

### Environment Variables

```bash
# Memory Consolidation
NEXS_MEMORY_CONSOLIDATION_ENABLED=true
NEXS_MEMORY_CONSOLIDATION_AUTO=false
NEXS_MEMORY_CONSOLIDATION_INTERVAL=24h
NEXS_MEMORY_CONSOLIDATION_MIN_MEMORIES=10
NEXS_MEMORY_CONSOLIDATION_DUPLICATES=true
NEXS_MEMORY_CONSOLIDATION_CLUSTERING=true
NEXS_MEMORY_CONSOLIDATION_KNOWLEDGE=true
NEXS_MEMORY_CONSOLIDATION_QUALITY=true

# Duplicate Detection
NEXS_DUPLICATE_DETECTION_ENABLED=true
NEXS_DUPLICATE_DETECTION_THRESHOLD=0.95
NEXS_DUPLICATE_DETECTION_MIN_LENGTH=20
NEXS_DUPLICATE_DETECTION_MAX_RESULTS=100

# Clustering
NEXS_CLUSTERING_ENABLED=true
NEXS_CLUSTERING_ALGORITHM=dbscan
NEXS_CLUSTERING_MIN_SIZE=3
NEXS_CLUSTERING_EPSILON=0.15
NEXS_CLUSTERING_NUM_CLUSTERS=10
NEXS_CLUSTERING_MAX_ITERATIONS=100

# Knowledge Graph
NEXS_KNOWLEDGE_GRAPH_ENABLED=true
NEXS_KNOWLEDGE_GRAPH_EXTRACT_PEOPLE=true
NEXS_KNOWLEDGE_GRAPH_EXTRACT_ORGS=true
NEXS_KNOWLEDGE_GRAPH_EXTRACT_URLS=true
NEXS_KNOWLEDGE_GRAPH_EXTRACT_EMAILS=true
NEXS_KNOWLEDGE_GRAPH_EXTRACT_CONCEPTS=true
NEXS_KNOWLEDGE_GRAPH_EXTRACT_KEYWORDS=true
NEXS_KNOWLEDGE_GRAPH_MAX_KEYWORDS=10
NEXS_KNOWLEDGE_GRAPH_EXTRACT_RELATIONSHIPS=true
NEXS_KNOWLEDGE_GRAPH_MAX_RELATIONSHIPS=20

# Hybrid Search
NEXS_HYBRID_SEARCH_ENABLED=true
NEXS_HYBRID_SEARCH_MODE=auto
NEXS_HYBRID_SEARCH_THRESHOLD=0.7
NEXS_HYBRID_SEARCH_MAX_RESULTS=10
NEXS_HYBRID_SEARCH_AUTO_SWITCH=100
NEXS_HYBRID_SEARCH_PERSISTENCE=true
NEXS_HYBRID_SEARCH_INDEX_PATH=data/hnsw-index

# Memory Retention
NEXS_MEMORY_RETENTION_ENABLED=true
NEXS_MEMORY_RETENTION_THRESHOLD=0.5
NEXS_MEMORY_RETENTION_HIGH_DAYS=365
NEXS_MEMORY_RETENTION_MEDIUM_DAYS=180
NEXS_MEMORY_RETENTION_LOW_DAYS=90
NEXS_MEMORY_RETENTION_AUTO_CLEANUP=false
NEXS_MEMORY_RETENTION_CLEANUP_INTERVAL=24h

# Context Enrichment
NEXS_CONTEXT_ENRICHMENT_ENABLED=true
NEXS_CONTEXT_ENRICHMENT_MAX_MEMORIES=5
NEXS_CONTEXT_ENRICHMENT_MAX_DEPTH=2
NEXS_CONTEXT_ENRICHMENT_RELATIONSHIPS=true
NEXS_CONTEXT_ENRICHMENT_TIMESTAMPS=true
NEXS_CONTEXT_ENRICHMENT_THRESHOLD=0.6

# Embeddings
NEXS_EMBEDDINGS_PROVIDER=onnx
NEXS_EMBEDDINGS_DIMENSION=384
NEXS_EMBEDDINGS_CACHE_ENABLED=true
NEXS_EMBEDDINGS_CACHE_TTL=24h
NEXS_EMBEDDINGS_CACHE_SIZE=10000
NEXS_EMBEDDINGS_BATCH_SIZE=32
```

### CLI Flags

```bash
# Enable/disable features
--memory-consolidation-enabled=true
--memory-consolidation-auto=false
--memory-consolidation-interval=24h
--duplicate-detection-enabled=true
--clustering-enabled=true
--clustering-algorithm=dbscan
--knowledge-graph-enabled=true
--hybrid-search-enabled=true
--hybrid-search-mode=auto
--memory-retention-enabled=true
--memory-retention-auto-cleanup=false
--context-enrichment-enabled=true
--context-enrichment-max-memories=5
--embeddings-provider=onnx
--embeddings-cache-enabled=true
```

---

## Tool Reference

### consolidate_memories

Execute complete memory consolidation workflow with all features enabled.

**Use Cases:**
- Regular maintenance (weekly/monthly)
- After importing large datasets
- Before archiving or exporting
- Quality improvement initiatives

**Parameters:**
```json
{
  "element_type": "memory",
  "min_quality": 0.3,
  "enable_duplicate_detection": true,
  "enable_clustering": true,
  "enable_knowledge_extraction": true,
  "enable_quality_scoring": true,
  "dry_run": false
}
```

**Example 1: Full Consolidation**
```bash
curl -X POST http://localhost:3000/mcp/tools/consolidate_memories \
  -H "Content-Type: application/json" \
  -d '{
    "element_type": "memory",
    "min_quality": 0.5,
    "enable_duplicate_detection": true,
    "enable_clustering": true,
    "enable_knowledge_extraction": true,
    "enable_quality_scoring": true,
    "dry_run": false
  }'
```

**Response:**
```json
{
  "workflow_id": "consolidation-20251226-001",
  "duration_ms": 3456,
  "steps_executed": {
    "duplicate_detection": {
      "status": "completed",
      "duplicates_found": 15,
      "groups": 7,
      "merged": 6,
      "duration_ms": 890
    },
    "clustering": {
      "status": "completed",
      "algorithm": "dbscan",
      "clusters_created": 12,
      "memories_clustered": 145,
      "outliers": 8,
      "duration_ms": 720
    },
    "knowledge_extraction": {
      "status": "completed",
      "entities_extracted": 234,
      "relationships_created": 156,
      "keywords_tagged": 89,
      "duration_ms": 1120
    },
    "quality_scoring": {
      "status": "completed",
      "memories_scored": 153,
      "avg_quality": 0.72,
      "high_quality": 45,
      "low_quality": 12,
      "duration_ms": 726
    }
  },
  "recommendations": [
    "Consider removing 12 low-quality memories (quality < 0.3)",
    "Cluster 'project-alpha' contains 23 memories, consider summarization",
    "Entity 'John Smith' appears in 15 memories, strong relationship detected"
  ],
  "summary": {
    "total_memories_processed": 153,
    "duplicates_removed": 6,
    "new_clusters": 12,
    "new_relationships": 156,
    "quality_improved": 0.08
  }
}
```

**Example 2: Dry Run (Preview Only)**
```json
{
  "element_type": "memory",
  "min_quality": 0.5,
  "enable_duplicate_detection": true,
  "enable_clustering": false,
  "enable_knowledge_extraction": false,
  "enable_quality_scoring": true,
  "dry_run": true
}
```

---

### detect_duplicates

Find and optionally merge duplicate or highly similar elements.

**Use Cases:**
- Before batch imports
- Regular duplicate cleanup
- Data quality audits
- Storage optimization

**Example 1: Find Duplicates (Manual Review)**
```json
{
  "element_type": "memory",
  "similarity_threshold": 0.95,
  "min_content_length": 50,
  "max_results": 50,
  "auto_merge": false
}
```

**Response:**
```json
{
  "duplicate_groups": [
    {
      "group_id": "dup-001",
      "similarity": 0.98,
      "elements": [
        {
          "id": "memory-123",
          "name": "Meeting Notes - Q4 Planning",
          "type": "memory",
          "content_preview": "Discussed Q4 objectives and key results...",
          "created_at": "2025-12-20T10:00:00Z",
          "updated_at": "2025-12-20T10:00:00Z",
          "size_bytes": 1250
        },
        {
          "id": "memory-456",
          "name": "Q4 Planning Meeting Notes",
          "type": "memory",
          "content_preview": "Q4 objectives discussion and KR definition...",
          "created_at": "2025-12-20T11:30:00Z",
          "updated_at": "2025-12-20T11:30:00Z",
          "size_bytes": 1180
        }
      ],
      "recommended_action": "merge",
      "keep_element_id": "memory-123",
      "merge_reason": "Earlier timestamp, more comprehensive content"
    }
  ],
  "total_groups": 7,
  "total_duplicates": 15,
  "potential_space_saved_kb": 45
}
```

**Example 2: Auto-Merge Duplicates**
```json
{
  "element_type": "memory",
  "similarity_threshold": 0.98,
  "auto_merge": true
}
```

**Example 3: Lower Threshold (Near-Duplicates)**
```json
{
  "element_type": "memory",
  "similarity_threshold": 0.85,
  "min_content_length": 100,
  "auto_merge": false
}
```

---

### cluster_memories

Group related memories using clustering algorithms.

**Use Cases:**
- Topic organization
- Project grouping
- Knowledge discovery
- Archive preparation

**Example 1: DBSCAN Clustering (Density-Based)**
```json
{
  "algorithm": "dbscan",
  "min_cluster_size": 5,
  "epsilon_distance": 0.12,
  "element_type": "memory"
}
```

**Response:**
```json
{
  "algorithm": "dbscan",
  "clusters": [
    {
      "cluster_id": "cluster-001",
      "name": "Project Alpha Discussions",
      "size": 23,
      "members": [
        {
          "id": "memory-101",
          "name": "Sprint Planning Meeting",
          "distance_to_centroid": 0.08,
          "is_representative": true
        },
        {
          "id": "memory-102",
          "name": "Sprint Review Notes",
          "distance_to_centroid": 0.11,
          "is_representative": false
        }
      ],
      "centroid_embedding": [0.123, -0.456, 0.789, ...],
      "keywords": ["project", "alpha", "sprint", "planning", "goals"],
      "date_range": {
        "earliest": "2025-11-15T10:00:00Z",
        "latest": "2025-12-20T15:00:00Z"
      },
      "avg_quality_score": 0.78,
      "total_content_bytes": 28450
    }
  ],
  "outliers": [
    {
      "id": "memory-999",
      "name": "Random Personal Note",
      "reason": "No similar memories found within epsilon distance"
    }
  ],
  "statistics": {
    "total_memories": 153,
    "clustered": 145,
    "outliers": 8,
    "num_clusters": 12,
    "avg_cluster_size": 12.08,
    "silhouette_score": 0.73,
    "clustering_quality": "good"
  }
}
```

**Example 2: K-Means Clustering (Fixed Number)**
```json
{
  "algorithm": "kmeans",
  "num_clusters": 8,
  "max_iterations": 150,
  "element_type": "memory"
}
```

**Example 3: Aggressive Clustering (Larger Groups)**
```json
{
  "algorithm": "dbscan",
  "min_cluster_size": 10,
  "epsilon_distance": 0.20,
  "element_type": "memory"
}
```

---

### extract_knowledge_graph

Extract entities, relationships, and keywords from content.

**Use Cases:**
- Building knowledge bases
- Relationship discovery
- Entity tracking
- Semantic enrichment

**Example 1: Extract All Entity Types**
```json
{
  "element_type": "memory",
  "extract_people": true,
  "extract_organizations": true,
  "extract_urls": true,
  "extract_emails": true,
  "extract_concepts": true,
  "extract_keywords": true,
  "max_keywords": 10,
  "extract_relationships": true,
  "max_relationships": 20
}
```

**Response:**
```json
{
  "knowledge_graph": {
    "entities": {
      "people": [
        {
          "name": "John Smith",
          "mentions": 15,
          "contexts": ["project-alpha", "sprint-planning", "code-review"],
          "first_seen": "2025-11-15T10:00:00Z",
          "last_seen": "2025-12-20T15:00:00Z",
          "related_memories": ["memory-101", "memory-102", "memory-115"]
        },
        {
          "name": "Jane Doe",
          "mentions": 8,
          "contexts": ["project-beta", "architecture"],
          "first_seen": "2025-12-01T09:00:00Z",
          "last_seen": "2025-12-18T16:00:00Z",
          "related_memories": ["memory-203", "memory-210"]
        }
      ],
      "organizations": [
        {
          "name": "Acme Corp",
          "mentions": 12,
          "type": "company",
          "contexts": ["client", "partnership"]
        }
      ],
      "urls": [
        {
          "url": "https://github.com/example/repo",
          "mentions": 5,
          "context": "code repository",
          "first_seen": "2025-12-05T10:00:00Z"
        }
      ],
      "emails": [
        {
          "email": "john@example.com",
          "mentions": 3,
          "associated_person": "John Smith"
        }
      ],
      "concepts": [
        {
          "concept": "machine learning",
          "mentions": 12,
          "related_concepts": ["neural networks", "deep learning", "AI"],
          "confidence": 0.92
        }
      ]
    },
    "relationships": [
      {
        "from_entity": "John Smith",
        "to_entity": "Project Alpha",
        "relationship_type": "works_on",
        "strength": 0.87,
        "evidence_count": 15,
        "first_observed": "2025-11-15T10:00:00Z",
        "last_observed": "2025-12-20T15:00:00Z"
      },
      {
        "from_entity": "Jane Doe",
        "to_entity": "John Smith",
        "relationship_type": "collaborates_with",
        "strength": 0.65,
        "evidence_count": 4
      }
    ],
    "keywords": {
      "memory-001": ["planning", "sprint", "goals", "Q4"],
      "memory-002": ["review", "retrospective", "improvements", "team"]
    }
  },
  "statistics": {
    "elements_processed": 153,
    "entities_extracted": 234,
    "relationships_created": 156,
    "keywords_tagged": 89,
    "processing_time_ms": 1120
  }
}
```

**Example 2: Extract Only People and Keywords**
```json
{
  "element_type": "memory",
  "extract_people": true,
  "extract_organizations": false,
  "extract_keywords": true,
  "max_keywords": 5
}
```

**Example 3: Process Specific Elements**
```json
{
  "element_ids": ["memory-001", "memory-002", "memory-003"],
  "extract_people": true,
  "extract_relationships": true,
  "max_relationships": 10
}
```

---

### get_consolidation_report

Get comprehensive consolidation status and actionable recommendations.

**Use Cases:**
- Regular health checks
- Planning maintenance
- Performance monitoring
- Executive summaries

**Example:**
```json
{
  "element_type": "memory",
  "include_statistics": true,
  "include_recommendations": true
}
```

**Response:**
```json
{
  "report_id": "consolidation-report-20251226",
  "generated_at": "2025-12-26T10:00:00Z",
  "element_type": "memory",
  "statistics": {
    "total_elements": 153,
    "duplicate_groups": 7,
    "total_duplicates": 15,
    "clusters": 12,
    "outliers": 8,
    "entities_extracted": 234,
    "relationships": 156,
    "avg_quality_score": 0.72,
    "high_quality_elements": 45,
    "medium_quality_elements": 96,
    "low_quality_elements": 12,
    "total_storage_mb": 12.5,
    "avg_access_frequency": 3.2
  },
  "health_metrics": {
    "duplication_rate": 0.098,
    "clustering_effectiveness": 0.95,
    "knowledge_extraction_coverage": 0.87,
    "avg_retention_days": 142,
    "storage_efficiency": 0.92,
    "search_performance_score": 0.88
  },
  "recommendations": [
    {
      "priority": "high",
      "category": "quality",
      "issue": "12 memories below quality threshold (0.5)",
      "action": "Review and remove low-quality memories",
      "impact": "Improve avg quality from 0.72 to 0.79, save 1.2MB",
      "estimated_time": "15 minutes",
      "command": "apply_retention_policy"
    },
    {
      "priority": "high",
      "category": "duplication",
      "issue": "7 duplicate groups with 15 total duplicates",
      "action": "Merge duplicate memories",
      "impact": "Save 45KB storage, improve search accuracy by 5%",
      "estimated_time": "10 minutes",
      "command": "detect_duplicates --auto-merge=true"
    },
    {
      "priority": "medium",
      "category": "organization",
      "issue": "8 outlier memories without clusters",
      "action": "Review outliers for relevance or recategorization",
      "impact": "Better memory organization and search results",
      "estimated_time": "5 minutes",
      "command": "cluster_memories"
    },
    {
      "priority": "medium",
      "category": "knowledge",
      "issue": "Entity 'John Smith' highly connected (15 relationships)",
      "action": "Consider creating dedicated agent or persona profile",
      "impact": "Better context tracking and relationship management",
      "estimated_time": "20 minutes",
      "command": "create_persona"
    },
    {
      "priority": "low",
      "category": "performance",
      "issue": "HNSW index not persisted",
      "action": "Enable index persistence for faster searches",
      "impact": "Reduce search latency by 40%",
      "estimated_time": "5 minutes",
      "config": "NEXS_HYBRID_SEARCH_PERSISTENCE=true"
    }
  ],
  "trends": {
    "quality_trend_7d": 0.05,
    "quality_trend_30d": 0.12,
    "duplicate_rate_trend_7d": -0.02,
    "storage_growth_rate_7d_mb": 2.3,
    "clustering_effectiveness_trend_7d": 0.03
  },
  "next_scheduled_consolidation": "2025-12-27T10:00:00Z"
}
```

---

### hybrid_search

Perform intelligent search with automatic HNSW/linear selection.

**Use Cases:**
- Fast semantic search
- Content discovery
- Similarity finding
- Context retrieval

**Example 1: Auto Mode (Recommended)**
```json
{
  "query": "machine learning implementation",
  "element_type": "memory",
  "mode": "auto",
  "similarity_threshold": 0.7,
  "max_results": 10
}
```

**Response:**
```json
{
  "search_mode": "hnsw",
  "query": "machine learning implementation",
  "results": [
    {
      "id": "memory-123",
      "name": "ML Project Implementation Notes",
      "type": "memory",
      "similarity": 0.92,
      "content_preview": "Discussed machine learning approach for user recommendation system...",
      "metadata": {
        "tags": ["ml", "project", "implementation"],
        "created_at": "2025-12-20T10:00:00Z",
        "author": "John Smith",
        "cluster_id": "cluster-003"
      },
      "knowledge_context": {
        "entities": ["machine learning", "recommendation system"],
        "relationships": ["related_to:memory-124"]
      }
    }
  ],
  "total_results": 15,
  "search_time_ms": 12,
  "index_stats": {
    "total_vectors": 1245,
    "index_size_mb": 18,
    "last_updated": "2025-12-26T09:00:00Z",
    "mode_switch_threshold": 100
  }
}
```

**Example 2: Force HNSW Mode**
```json
{
  "query": "project planning",
  "mode": "hnsw",
  "max_results": 20
}
```

**Example 3: With Filters**
```json
{
  "query": "sprint retrospective",
  "element_type": "memory",
  "mode": "auto",
  "similarity_threshold": 0.6,
  "filter_tags": ["sprint", "retrospective"],
  "date_from": "2025-12-01T00:00:00Z",
  "date_to": "2025-12-26T23:59:59Z"
}
```

---

### score_memory_quality

Calculate quality scores based on multiple factors.

**Use Cases:**
- Quality audits
- Cleanup planning
- Retention decisions
- Performance analysis

**Example:**
```json
{
  "element_type": "memory",
  "min_threshold": 0.3,
  "include_details": true
}
```

**Response:**
```json
{
  "scored_memories": [
    {
      "id": "memory-123",
      "name": "Project Planning Meeting",
      "quality_score": 0.87,
      "components": {
        "content_quality": 0.92,
        "structure_score": 0.85,
        "recency_score": 0.88,
        "relationship_score": 0.83,
        "access_score": 0.89
      },
      "factors": {
        "length": 1450,
        "has_title": true,
        "has_tags": true,
        "tag_count": 5,
        "num_relationships": 8,
        "age_days": 5,
        "access_count": 23,
        "last_accessed": "2025-12-25T14:30:00Z",
        "has_keywords": true,
        "keyword_count": 7
      },
      "classification": "high-quality",
      "retention_recommendation": "keep-365-days",
      "strengths": [
        "High access frequency",
        "Strong relationships with other memories",
        "Well-structured content with tags and keywords"
      ],
      "improvements": [
        "Consider adding more detailed tags for better categorization"
      ]
    }
  ],
  "statistics": {
    "total_scored": 153,
    "avg_quality": 0.72,
    "median_quality": 0.74,
    "high_quality": 45,
    "medium_quality": 96,
    "low_quality": 12,
    "quality_distribution": {
      "0.0-0.3": 5,
      "0.3-0.5": 7,
      "0.5-0.7": 96,
      "0.7-0.9": 38,
      "0.9-1.0": 7
    }
  }
}
```

---

### apply_retention_policy

Apply retention policies based on quality and age.

**Use Cases:**
- Regular cleanup
- Storage optimization
- Compliance requirements
- Performance improvement

**Example 1: Dry Run (Preview)**
```json
{
  "quality_threshold": 0.5,
  "high_quality_days": 365,
  "medium_quality_days": 180,
  "low_quality_days": 90,
  "dry_run": true,
  "element_type": "memory"
}
```

**Response:**
```json
{
  "policy_applied": {
    "quality_threshold": 0.5,
    "high_quality_days": 365,
    "medium_quality_days": 180,
    "low_quality_days": 90
  },
  "actions_taken": [
    {
      "action": "delete",
      "element_id": "memory-999",
      "element_name": "Old Random Note",
      "reason": "Quality 0.25 below threshold 0.5",
      "quality_score": 0.25,
      "age_days": 120,
      "last_accessed": "2025-09-15T10:00:00Z",
      "size_bytes": 450
    },
    {
      "action": "delete",
      "element_id": "memory-888",
      "element_name": "Outdated Meeting Notes",
      "reason": "Low quality (0.45), exceeded 90-day retention",
      "quality_score": 0.45,
      "age_days": 95,
      "last_accessed": "2025-11-20T14:00:00Z",
      "size_bytes": 1250
    }
  ],
  "summary": {
    "elements_reviewed": 153,
    "high_quality_kept": 45,
    "medium_quality_kept": 96,
    "low_quality_removed": 12,
    "space_freed_kb": 78,
    "dry_run": true
  },
  "recommendations": [
    "12 elements will be deleted if policy is applied",
    "Consider backing up before running with dry_run=false",
    "Review elements with quality < 0.5 manually before deletion"
  ]
}
```

**Example 2: Apply Policy (Actual Cleanup)**
```json
{
  "quality_threshold": 0.5,
  "high_quality_days": 365,
  "medium_quality_days": 180,
  "low_quality_days": 90,
  "dry_run": false,
  "element_type": "memory"
}
```

---

### enrich_context

Enrich element context by traversing relationships.

**Use Cases:**
- Contextual retrieval
- Relationship exploration
- Knowledge discovery
- AI context building

**Example:**
```json
{
  "element_id": "memory-123",
  "max_related": 5,
  "max_depth": 2,
  "include_relationships": true,
  "include_timestamps": true,
  "similarity_threshold": 0.6
}
```

**Response:**
```json
{
  "element": {
    "id": "memory-123",
    "name": "Project Planning Meeting",
    "type": "memory",
    "content": "Discussed Q4 objectives and key results...",
    "created_at": "2025-12-20T10:00:00Z"
  },
  "enriched_context": {
    "related_memories": [
      {
        "id": "memory-124",
        "name": "Q4 Goals Summary",
        "similarity": 0.89,
        "relationship_type": "related_to",
        "relationship_strength": 0.85,
        "distance": 1,
        "content_preview": "Summary of Q4 goals discussed in planning meeting..."
      },
      {
        "id": "memory-125",
        "name": "Sprint 1 Planning",
        "similarity": 0.76,
        "relationship_type": "follows_from",
        "relationship_strength": 0.72,
        "distance": 2,
        "content_preview": "Sprint 1 planning based on Q4 objectives..."
      }
    ],
    "temporal_context": {
      "created_at": "2025-12-20T10:00:00Z",
      "updated_at": "2025-12-20T15:00:00Z",
      "last_accessed": "2025-12-25T14:30:00Z",
      "related_timeframe_days": 7,
      "temporal_cluster": "Q4-2025"
    },
    "knowledge_context": {
      "entities": ["John Smith", "Project Alpha", "Q4 Goals"],
      "keywords": ["planning", "sprint", "goals", "objectives"],
      "relationships": [
        {
          "from": "John Smith",
          "to": "Project Alpha",
          "type": "works_on",
          "strength": 0.87
        }
      ]
    },
    "cluster_context": {
      "cluster_id": "cluster-001",
      "cluster_name": "Project Alpha Discussions",
      "cluster_size": 23,
      "position_in_cluster": "central"
    }
  },
  "depth_traversed": 2,
  "total_related_found": 12,
  "enrichment_quality_score": 0.91
}
```

---

## Workflow Examples

### Workflow 1: Weekly Maintenance

Complete maintenance workflow for regular memory health.

```bash
# Step 1: Get current status
curl -X POST http://localhost:3000/mcp/tools/get_consolidation_report \
  -H "Content-Type: application/json" \
  -d '{
    "element_type": "memory",
    "include_statistics": true,
    "include_recommendations": true
  }'

# Step 2: Detect and merge duplicates
curl -X POST http://localhost:3000/mcp/tools/detect_duplicates \
  -H "Content-Type: application/json" \
  -d '{
    "element_type": "memory",
    "similarity_threshold": 0.95,
    "auto_merge": true
  }'

# Step 3: Cluster memories
curl -X POST http://localhost:3000/mcp/tools/cluster_memories \
  -H "Content-Type: application/json" \
  -d '{
    "algorithm": "dbscan",
    "min_cluster_size": 3,
    "epsilon_distance": 0.15
  }'

# Step 4: Apply retention policy (dry run first)
curl -X POST http://localhost:3000/mcp/tools/apply_retention_policy \
  -H "Content-Type: application/json" \
  -d '{
    "quality_threshold": 0.5,
    "high_quality_days": 365,
    "medium_quality_days": 180,
    "low_quality_days": 90,
    "dry_run": true
  }'

# Step 5: Apply if satisfied with preview
curl -X POST http://localhost:3000/mcp/tools/apply_retention_policy \
  -H "Content-Type: application/json" \
  -d '{
    "quality_threshold": 0.5,
    "high_quality_days": 365,
    "medium_quality_days": 180,
    "low_quality_days": 90,
    "dry_run": false
  }'
```

### Workflow 2: Knowledge Base Building

Extract knowledge and build relationship graph.

```bash
# Step 1: Extract knowledge graph
curl -X POST http://localhost:3000/mcp/tools/extract_knowledge_graph \
  -H "Content-Type: application/json" \
  -d '{
    "element_type": "memory",
    "extract_people": true,
    "extract_organizations": true,
    "extract_keywords": true,
    "extract_relationships": true,
    "max_keywords": 10,
    "max_relationships": 20
  }'

# Step 2: Cluster by topic
curl -X POST http://localhost:3000/mcp/tools/cluster_memories \
  -H "Content-Type: application/json" \
  -d '{
    "algorithm": "kmeans",
    "num_clusters": 8
  }'

# Step 3: Score quality
curl -X POST http://localhost:3000/mcp/tools/score_memory_quality \
  -H "Content-Type: application/json" \
  -d '{
    "element_type": "memory",
    "include_details": true
  }'
```

### Workflow 3: One-Command Full Consolidation

Execute complete consolidation in single step.

```bash
curl -X POST http://localhost:3000/mcp/tools/consolidate_memories \
  -H "Content-Type: application/json" \
  -d '{
    "element_type": "memory",
    "min_quality": 0.5,
    "enable_duplicate_detection": true,
    "enable_clustering": true,
    "enable_knowledge_extraction": true,
    "enable_quality_scoring": true,
    "dry_run": false
  }'
```

---

## Best Practices

### 1. Regular Maintenance Schedule

```bash
# Daily: Quick duplicate check
SCHEDULE="0 2 * * *"  # 2 AM daily
detect_duplicates --auto-merge=false

# Weekly: Full consolidation
SCHEDULE="0 3 * * 0"  # 3 AM Sunday
consolidate_memories --min-quality=0.5

# Monthly: Quality review and cleanup
SCHEDULE="0 4 1 * *"  # 4 AM 1st of month
apply_retention_policy --dry-run=false
```

### 2. Optimal Configuration

```bash
# For small datasets (< 100 memories)
NEXS_CLUSTERING_ALGORITHM=kmeans
NEXS_CLUSTERING_NUM_CLUSTERS=5
NEXS_HYBRID_SEARCH_MODE=linear

# For medium datasets (100-1000 memories)
NEXS_CLUSTERING_ALGORITHM=dbscan
NEXS_CLUSTERING_MIN_SIZE=3
NEXS_CLUSTERING_EPSILON=0.15
NEXS_HYBRID_SEARCH_MODE=auto

# For large datasets (> 1000 memories)
NEXS_CLUSTERING_ALGORITHM=dbscan
NEXS_CLUSTERING_MIN_SIZE=5
NEXS_CLUSTERING_EPSILON=0.12
NEXS_HYBRID_SEARCH_MODE=hnsw
NEXS_HYBRID_SEARCH_PERSISTENCE=true
```

### 3. Quality Thresholds

```bash
# Conservative (keep more)
NEXS_MEMORY_RETENTION_THRESHOLD=0.3
NEXS_DUPLICATE_DETECTION_THRESHOLD=0.98

# Balanced (recommended)
NEXS_MEMORY_RETENTION_THRESHOLD=0.5
NEXS_DUPLICATE_DETECTION_THRESHOLD=0.95

# Aggressive (keep less)
NEXS_MEMORY_RETENTION_THRESHOLD=0.7
NEXS_DUPLICATE_DETECTION_THRESHOLD=0.90
```

### 4. Performance Optimization

```bash
# Enable caching
NEXS_EMBEDDINGS_CACHE_ENABLED=true
NEXS_EMBEDDINGS_CACHE_SIZE=10000

# Enable index persistence
NEXS_HYBRID_SEARCH_PERSISTENCE=true
NEXS_HYBRID_SEARCH_INDEX_PATH=data/hnsw-index

# Batch processing
NEXS_EMBEDDINGS_BATCH_SIZE=64

# Auto-switch threshold
NEXS_HYBRID_SEARCH_AUTO_SWITCH=100
```

---

## Troubleshooting

### Issue: Slow Consolidation

**Symptoms:**
- Consolidation takes > 5 minutes
- High CPU usage
- Timeouts

**Solutions:**
```bash
# Reduce batch sizes
NEXS_EMBEDDINGS_BATCH_SIZE=16

# Enable caching
NEXS_EMBEDDINGS_CACHE_ENABLED=true

# Process in smaller batches
consolidate_memories --element-type=memory --max-elements=100

# Use linear search for small datasets
NEXS_HYBRID_SEARCH_MODE=linear
```

### Issue: Low Clustering Quality

**Symptoms:**
- Too many outliers
- Clusters don't make sense
- Silhouette score < 0.5

**Solutions:**
```bash
# Adjust DBSCAN parameters
NEXS_CLUSTERING_EPSILON=0.20  # Increase for larger clusters
NEXS_CLUSTERING_MIN_SIZE=2    # Decrease for smaller clusters

# Try K-means instead
NEXS_CLUSTERING_ALGORITHM=kmeans
NEXS_CLUSTERING_NUM_CLUSTERS=10

# Re-run after knowledge extraction
extract_knowledge_graph
cluster_memories
```

### Issue: Too Many False Duplicates

**Symptoms:**
- Unrelated memories marked as duplicates
- Important content being merged

**Solutions:**
```bash
# Increase threshold
NEXS_DUPLICATE_DETECTION_THRESHOLD=0.98

# Increase minimum length
NEXS_DUPLICATE_DETECTION_MIN_LENGTH=50

# Disable auto-merge
detect_duplicates --auto-merge=false

# Review manually before merging
```

### Issue: Low Quality Scores

**Symptoms:**
- Most memories scored < 0.5
- Quality distribution skewed low

**Solutions:**
```bash
# Extract knowledge to improve scores
extract_knowledge_graph

# Add tags and metadata
# (use element update tools)

# Cluster to build relationships
cluster_memories

# Review scoring factors
score_memory_quality --include-details=true
```

---

## Related Documentation

- [MCP Tools Reference](./MCP_TOOLS.md)
- [Memory Consolidation Developer Guide](../development/MEMORY_CONSOLIDATION.md)
- [Memory Consolidation User Guide](../user-guide/MEMORY_CONSOLIDATION.md)
- [Application Architecture](../architecture/APPLICATION.md)
- [Configuration Reference](../api/CLI.md)
- [Testing Guide](../development/TESTING.md)

---

**Last Updated:** December 26, 2025  
**Version:** v1.3.0  
**Maintainer:** NEXS-MCP Team
