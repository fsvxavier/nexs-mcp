# Memory Consolidation - User Guide

**Version:** 1.3.0  
**Last Updated:** December 26, 2025  
**Target Audience:** End Users and AI Agents

---

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Features Overview](#features-overview)
4. [Using Memory Consolidation](#using-memory-consolidation)
5. [Common Workflows](#common-workflows)
6. [Configuration Guide](#configuration-guide)
7. [Best Practices](#best-practices)
8. [Tips and Tricks](#tips-and-tricks)
9. [Troubleshooting](#troubleshooting)
10. [FAQ](#faq)

---

## Introduction

Memory Consolidation is a powerful feature in NEXS-MCP that helps you organize, clean up, and optimize your memories. Think of it as a smart assistant that automatically:

- 🔍 **Finds duplicates** - Detects similar memories you've accidentally created multiple times
- 📊 **Organizes by topic** - Groups related memories into clusters (like "Project Alpha", "Team Meetings")
- 🧠 **Builds knowledge** - Extracts people, organizations, and relationships from your memories
- ⚡ **Improves search** - Makes finding memories faster and more accurate
- 🎯 **Maintains quality** - Scores memories and suggests which ones to keep or remove
- 🗑️ **Cleans up** - Automatically removes old, low-quality memories (if enabled)

### Why Use Memory Consolidation?

**Before Consolidation:**
- 50+ memories, many duplicates
- Hard to find specific information
- Cluttered and disorganized
- Slow searches
- No relationships between memories

**After Consolidation:**
- 35 unique, high-quality memories
- Organized into 8 topic clusters
- 50+ entities and relationships extracted
- Fast, accurate searches
- Clear knowledge graph

---

## Getting Started

### Prerequisites

- NEXS-MCP v1.3.0 or later
- At least 10 memories created
- Basic familiarity with MCP tools

### Quick Start (5 minutes)

**Step 1: Check Your Current Status**

```bash
# Get consolidation report
nexs-mcp get_consolidation_report \
  --element-type memory \
  --include-statistics \
  --include-recommendations
```

This shows you:
- How many memories you have
- Duplicate rate
- Clustering effectiveness
- Quality scores
- Actionable recommendations

**Step 2: Run Full Consolidation**

```bash
# One-command consolidation (preview mode)
nexs-mcp consolidate_memories \
  --element-type memory \
  --min-quality 0.5 \
  --enable-duplicate-detection \
  --enable-clustering \
  --enable-knowledge-extraction \
  --enable-quality-scoring \
  --dry-run
```

Review the preview, then run again with `--no-dry-run` to apply changes.

**Step 3: Explore Results**

```bash
# View clusters
nexs-mcp cluster_memories --algorithm dbscan

# View knowledge graph
nexs-mcp extract_knowledge_graph --element-type memory

# Search with improved index
nexs-mcp hybrid_search --query "project planning" --mode auto
```

---

## Features Overview

### 1. Duplicate Detection

**What it does:** Finds memories that are too similar (like duplicates).

**Example:**
```
Memory 1: "Meeting notes for Q4 planning - discussed goals"
Memory 2: "Q4 planning meeting - discussed goals and objectives"
Similarity: 98% → Marked as duplicates
```

**When to use:**
- After importing memories from multiple sources
- Regular cleanup (monthly)
- When you notice similar memories

**Options:**
- `similarity_threshold`: How similar to consider duplicates (0.90-0.98)
- `auto_merge`: Automatically merge duplicates (use carefully!)

### 2. Clustering

**What it does:** Groups related memories by topic.

**Example:**
```
Cluster 1: "Project Alpha" (15 memories)
  - Sprint planning meetings
  - Code review notes
  - Architecture discussions

Cluster 2: "Team Onboarding" (8 memories)
  - New hire documentation
  - Training materials
  - Setup guides
```

**Algorithms:**
- **DBSCAN** (recommended): Finds natural groupings, identifies outliers
- **K-means**: Creates fixed number of groups

**When to use:**
- Organize large collections (50+ memories)
- Discover hidden patterns
- Prepare for archiving

### 3. Knowledge Graph Extraction

**What it does:** Finds people, organizations, and relationships in your memories.

**Example:**
```
Entities Found:
  People: John Smith, Jane Doe, Dr. Robert Johnson
  Organizations: Acme Corp, MIT, Google
  Concepts: machine learning, neural networks

Relationships:
  John Smith → works_on → Project Alpha
  Jane Doe → collaborates_with → John Smith
  Dr. Robert Johnson → belongs_to → MIT
```

**When to use:**
- Build team knowledge base
- Track project relationships
- Understand collaboration patterns

### 4. Hybrid Search

**What it does:** Super-fast search that automatically switches between HNSW (fast) and linear (accurate) modes.

**Modes:**
- **Auto (recommended)**: Automatically picks best mode
- **HNSW**: Fast approximate search for large datasets
- **Linear**: Exhaustive accurate search for small datasets

**Example:**
```bash
# Search with auto mode
nexs-mcp hybrid_search \
  --query "machine learning implementation" \
  --mode auto \
  --similarity-threshold 0.7 \
  --max-results 10
```

### 5. Quality Scoring

**What it does:** Scores each memory's quality (0.0-1.0) based on:
- Content quality (length, structure, formatting)
- Recency (how old is it)
- Relationships (connections to other memories)
- Access patterns (how often it's used)

**Quality Tiers:**
- **High (0.7-1.0)**: Keep for 365 days
- **Medium (0.5-0.7)**: Keep for 180 days
- **Low (< 0.5)**: Review or delete after 90 days

### 6. Retention Policies

**What it does:** Automatically removes old, low-quality memories based on rules you set.

**Safety Features:**
- Dry-run mode (preview before deleting)
- Quality thresholds
- Age-based retention
- Manual override

**Example Policy:**
```
High quality (≥0.7): Keep 365 days
Medium quality (0.5-0.7): Keep 180 days
Low quality (<0.5): Keep 90 days
Below threshold (0.3): Delete immediately
```

---

## Using Memory Consolidation

### One-Command Consolidation

The easiest way to consolidate all your memories:

```bash
nexs-mcp consolidate_memories \
  --element-type memory \
  --min-quality 0.5 \
  --enable-duplicate-detection \
  --enable-clustering \
  --enable-knowledge-extraction \
  --enable-quality-scoring \
  --dry-run
```

**Output:**
```json
{
  "workflow_id": "consolidation-20251226-001",
  "duration_ms": 3456,
  "steps_executed": {
    "duplicate_detection": {
      "status": "completed",
      "duplicates_found": 7,
      "merged": 0
    },
    "clustering": {
      "status": "completed",
      "clusters_created": 8,
      "outliers": 3
    },
    "knowledge_extraction": {
      "status": "completed",
      "entities_extracted": 45,
      "relationships_created": 23
    },
    "quality_scoring": {
      "status": "completed",
      "avg_quality": 0.72,
      "low_quality": 5
    }
  },
  "recommendations": [
    "7 duplicate groups found - consider merging",
    "5 low-quality memories - consider removing",
    "Cluster 'Project Alpha' has 15 memories - well organized"
  ]
}
```

### Individual Tools

Use specific tools for targeted actions:

#### Find Duplicates Only

```bash
nexs-mcp detect_duplicates \
  --element-type memory \
  --similarity-threshold 0.95 \
  --auto-merge false
```

#### Cluster Memories Only

```bash
# DBSCAN (automatic topic discovery)
nexs-mcp cluster_memories \
  --algorithm dbscan \
  --min-cluster-size 3 \
  --epsilon-distance 0.15

# K-means (fixed number of topics)
nexs-mcp cluster_memories \
  --algorithm kmeans \
  --num-clusters 8
```

#### Extract Knowledge Only

```bash
nexs-mcp extract_knowledge_graph \
  --element-type memory \
  --extract-people \
  --extract-keywords \
  --max-keywords 10
```

#### Search Only

```bash
nexs-mcp hybrid_search \
  --query "sprint retrospective" \
  --mode auto \
  --max-results 10
```

#### Score Quality Only

```bash
nexs-mcp score_memory_quality \
  --element-type memory \
  --min-threshold 0.3 \
  --include-details
```

#### Apply Retention Only

```bash
# Dry run first (preview)
nexs-mcp apply_retention_policy \
  --quality-threshold 0.5 \
  --high-quality-days 365 \
  --medium-quality-days 180 \
  --low-quality-days 90 \
  --dry-run

# Apply if satisfied
nexs-mcp apply_retention_policy \
  --quality-threshold 0.5 \
  --high-quality-days 365 \
  --medium-quality-days 180 \
  --low-quality-days 90 \
  --no-dry-run
```

---

## Common Workflows

### Workflow 1: First-Time Setup

**Goal:** Organize existing memories and set up consolidation.

```bash
# Step 1: Get baseline report
nexs-mcp get_consolidation_report \
  --element-type memory \
  --include-statistics \
  --include-recommendations

# Step 2: Run full consolidation (dry run)
nexs-mcp consolidate_memories \
  --element-type memory \
  --min-quality 0.5 \
  --enable-duplicate-detection \
  --enable-clustering \
  --enable-knowledge-extraction \
  --enable-quality-scoring \
  --dry-run

# Step 3: Review recommendations and apply
# (run again with --no-dry-run if satisfied)

# Step 4: Enable auto-consolidation (optional)
export NEXS_MEMORY_CONSOLIDATION_AUTO=true
export NEXS_MEMORY_CONSOLIDATION_INTERVAL=168h  # Weekly
```

### Workflow 2: Weekly Maintenance

**Goal:** Keep memories organized with minimal effort.

```bash
# Monday morning routine
nexs-mcp consolidate_memories \
  --element-type memory \
  --min-quality 0.5 \
  --enable-duplicate-detection \
  --enable-clustering \
  --no-dry-run
```

### Workflow 3: Before Important Search

**Goal:** Improve search accuracy before looking for specific information.

```bash
# Step 1: Update clusters
nexs-mcp cluster_memories --algorithm dbscan

# Step 2: Refresh knowledge graph
nexs-mcp extract_knowledge_graph --element-type memory

# Step 3: Search
nexs-mcp hybrid_search \
  --query "your search query" \
  --mode auto \
  --similarity-threshold 0.7
```

### Workflow 4: Spring Cleaning

**Goal:** Deep cleanup of old, low-quality memories.

```bash
# Step 1: Score all memories
nexs-mcp score_memory_quality \
  --element-type memory \
  --include-details

# Step 2: Preview cleanup
nexs-mcp apply_retention_policy \
  --quality-threshold 0.5 \
  --high-quality-days 365 \
  --medium-quality-days 180 \
  --low-quality-days 90 \
  --dry-run

# Step 3: Review list of memories to be deleted

# Step 4: Apply if satisfied
nexs-mcp apply_retention_policy \
  --quality-threshold 0.5 \
  --high-quality-days 365 \
  --medium-quality-days 180 \
  --low-quality-days 90 \
  --no-dry-run

# Step 5: Merge duplicates
nexs-mcp detect_duplicates \
  --similarity-threshold 0.95 \
  --auto-merge true
```

### Workflow 5: Project Archive Preparation

**Goal:** Organize project memories before archiving.

```bash
# Step 1: Cluster project memories
nexs-mcp cluster_memories \
  --algorithm dbscan \
  --element-type memory

# Step 2: Extract project knowledge
nexs-mcp extract_knowledge_graph \
  --element-type memory \
  --extract-people \
  --extract-organizations \
  --extract-relationships

# Step 3: Generate report
nexs-mcp get_consolidation_report \
  --element-type memory \
  --include-statistics

# Step 4: Export (use regular export tools)
```

---

## Configuration Guide

### Environment Variables

```bash
# Enable/disable consolidation
export NEXS_MEMORY_CONSOLIDATION_ENABLED=true

# Auto-consolidation (runs on schedule)
export NEXS_MEMORY_CONSOLIDATION_AUTO=false
export NEXS_MEMORY_CONSOLIDATION_INTERVAL=24h
export NEXS_MEMORY_CONSOLIDATION_MIN_MEMORIES=10

# Duplicate detection
export NEXS_DUPLICATE_DETECTION_ENABLED=true
export NEXS_DUPLICATE_DETECTION_THRESHOLD=0.95
export NEXS_DUPLICATE_DETECTION_MIN_LENGTH=20
export NEXS_DUPLICATE_DETECTION_MAX_RESULTS=100

# Clustering
export NEXS_CLUSTERING_ENABLED=true
export NEXS_CLUSTERING_ALGORITHM=dbscan
export NEXS_CLUSTERING_MIN_SIZE=3
export NEXS_CLUSTERING_EPSILON=0.15
export NEXS_CLUSTERING_NUM_CLUSTERS=10

# Knowledge graph
export NEXS_KNOWLEDGE_GRAPH_ENABLED=true
export NEXS_KNOWLEDGE_GRAPH_EXTRACT_PEOPLE=true
export NEXS_KNOWLEDGE_GRAPH_EXTRACT_ORGS=true
export NEXS_KNOWLEDGE_GRAPH_EXTRACT_KEYWORDS=true
export NEXS_KNOWLEDGE_GRAPH_MAX_KEYWORDS=10

# Hybrid search
export NEXS_HYBRID_SEARCH_ENABLED=true
export NEXS_HYBRID_SEARCH_MODE=auto
export NEXS_HYBRID_SEARCH_THRESHOLD=0.7
export NEXS_HYBRID_SEARCH_MAX_RESULTS=10
export NEXS_HYBRID_SEARCH_PERSISTENCE=true
export NEXS_HYBRID_SEARCH_INDEX_PATH=data/hnsw-index

# Memory retention
export NEXS_MEMORY_RETENTION_ENABLED=true
export NEXS_MEMORY_RETENTION_THRESHOLD=0.5
export NEXS_MEMORY_RETENTION_HIGH_DAYS=365
export NEXS_MEMORY_RETENTION_MEDIUM_DAYS=180
export NEXS_MEMORY_RETENTION_LOW_DAYS=90
export NEXS_MEMORY_RETENTION_AUTO_CLEANUP=false

# Embeddings
export NEXS_EMBEDDINGS_PROVIDER=onnx
export NEXS_EMBEDDINGS_CACHE_ENABLED=true
export NEXS_EMBEDDINGS_CACHE_SIZE=10000
```

### CLI Flags

```bash
# Enable/disable features
nexs-mcp --memory-consolidation-enabled=true \
         --duplicate-detection-enabled=true \
         --clustering-enabled=true \
         --knowledge-graph-enabled=true \
         --hybrid-search-enabled=true

# Algorithm selection
nexs-mcp --clustering-algorithm=dbscan \
         --hybrid-search-mode=auto

# Quality thresholds
nexs-mcp --memory-retention-threshold=0.5

# Auto features
nexs-mcp --memory-consolidation-auto=false \
         --memory-retention-auto-cleanup=false
```

### Configuration Profiles

**Profile 1: Conservative (keep more)**
```bash
export NEXS_DUPLICATE_DETECTION_THRESHOLD=0.98
export NEXS_MEMORY_RETENTION_THRESHOLD=0.3
export NEXS_MEMORY_RETENTION_LOW_DAYS=180
```

**Profile 2: Balanced (recommended)**
```bash
export NEXS_DUPLICATE_DETECTION_THRESHOLD=0.95
export NEXS_MEMORY_RETENTION_THRESHOLD=0.5
export NEXS_MEMORY_RETENTION_LOW_DAYS=90
```

**Profile 3: Aggressive (keep less)**
```bash
export NEXS_DUPLICATE_DETECTION_THRESHOLD=0.90
export NEXS_MEMORY_RETENTION_THRESHOLD=0.7
export NEXS_MEMORY_RETENTION_LOW_DAYS=60
```

---

## Best Practices

### 1. Start with Dry Runs

Always preview changes before applying:

```bash
# Good ✅
nexs-mcp consolidate_memories --dry-run
# Review output, then run with --no-dry-run

# Risky ❌
nexs-mcp consolidate_memories --no-dry-run
# No preview, changes applied immediately
```

### 2. Run Regular Maintenance

Set up a schedule:

- **Daily:** Quick duplicate check (if creating many memories)
- **Weekly:** Full consolidation (recommended)
- **Monthly:** Deep cleanup with retention policies

### 3. Use Appropriate Thresholds

**Duplicate Detection:**
- 0.98: Very conservative (only exact duplicates)
- 0.95: Balanced (recommended)
- 0.90: Aggressive (catches near-duplicates)

**Quality Scoring:**
- 0.3: Keep almost everything
- 0.5: Balanced (recommended)
- 0.7: Keep only high quality

### 4. Choose Right Clustering Algorithm

**Use DBSCAN when:**
- You don't know how many topics you have
- You want automatic outlier detection
- Memories have natural groupings

**Use K-means when:**
- You want specific number of clusters
- Memories are evenly distributed
- You need consistent cluster sizes

### 5. Monitor Your Metrics

Check consolidation report regularly:

```bash
nexs-mcp get_consolidation_report \
  --element-type memory \
  --include-statistics
```

Watch for:
- Increasing duplicate rate (> 10%)
- Decreasing quality scores
- Growing outlier count
- Storage growth rate

### 6. Backup Before Major Cleanup

```bash
# Backup first
nexs-mcp backup_create --name "before-cleanup"

# Then cleanup
nexs-mcp apply_retention_policy --no-dry-run

# Restore if needed
nexs-mcp backup_restore --name "before-cleanup"
```

### 7. Use Tags for Better Organization

Well-tagged memories cluster better:

```bash
# Good ✅
create_memory --name "Sprint Planning" \
              --tags "sprint,planning,agile,team"

# Less effective ❌
create_memory --name "Notes"
```

---

## Tips and Tricks

### Tip 1: Find Specific Duplicate Groups

```bash
# Find duplicates of specific memory
nexs-mcp detect_duplicates \
  --element-ids "memory-123" \
  --similarity-threshold 0.90
```

### Tip 2: Cluster Specific Time Range

```bash
# Cluster only recent memories
nexs-mcp cluster_memories \
  --element-type memory \
  --date-from "2025-11-01" \
  --date-to "2025-12-26"
```

### Tip 3: Extract Knowledge from Specific Cluster

```bash
# First, get cluster members
# Then extract knowledge from those IDs
nexs-mcp extract_knowledge_graph \
  --element-ids "mem-1,mem-2,mem-3"
```

### Tip 4: Search Within Cluster

```bash
nexs-mcp hybrid_search \
  --query "your query" \
  --filter-tags "cluster-001"
```

### Tip 5: Quality Score Specific Memories

```bash
nexs-mcp score_memory_quality \
  --memory-ids "memory-1,memory-2,memory-3" \
  --include-details
```

### Tip 6: Enrich Context for AI

Before feeding memories to AI, enrich them:

```bash
nexs-mcp enrich_context \
  --element-id "memory-123" \
  --max-related 5 \
  --max-depth 2 \
  --include-relationships
```

### Tip 7: Monitor Consolidation Progress

Use workflow ID to track long-running consolidations:

```bash
# Start consolidation
result=$(nexs-mcp consolidate_memories --no-dry-run)
workflow_id=$(echo $result | jq -r '.workflow_id')

# Check status (if needed)
nexs-mcp get_workflow_status --workflow-id $workflow_id
```

---

## Troubleshooting

### Problem: Consolidation is Slow

**Symptoms:**
- Takes more than 5 minutes
- High CPU usage
- Process hangs

**Solutions:**
1. Process fewer memories at a time
```bash
nexs-mcp consolidate_memories --max-elements 100
```

2. Disable expensive features
```bash
nexs-mcp consolidate_memories \
  --enable-knowledge-extraction false
```

3. Use linear mode for small datasets
```bash
export NEXS_HYBRID_SEARCH_MODE=linear
```

4. Enable caching
```bash
export NEXS_EMBEDDINGS_CACHE_ENABLED=true
export NEXS_EMBEDDINGS_CACHE_SIZE=10000
```

### Problem: Too Many Duplicates Found

**Symptoms:**
- Memories that aren't really duplicates
- Threshold seems too low

**Solutions:**
1. Increase threshold
```bash
nexs-mcp detect_duplicates --similarity-threshold 0.98
```

2. Increase minimum content length
```bash
export NEXS_DUPLICATE_DETECTION_MIN_LENGTH=50
```

3. Review and merge manually
```bash
nexs-mcp detect_duplicates --auto-merge false
```

### Problem: Poor Clustering Results

**Symptoms:**
- Too many outliers
- Clusters don't make sense
- All memories in one cluster

**Solutions:**
1. Adjust DBSCAN parameters
```bash
# Larger clusters
nexs-mcp cluster_memories \
  --epsilon-distance 0.20 \
  --min-cluster-size 2

# Smaller, tighter clusters
nexs-mcp cluster_memories \
  --epsilon-distance 0.10 \
  --min-cluster-size 5
```

2. Try K-means instead
```bash
nexs-mcp cluster_memories \
  --algorithm kmeans \
  --num-clusters 8
```

3. Extract knowledge first
```bash
# This improves embeddings
nexs-mcp extract_knowledge_graph --element-type memory
nexs-mcp cluster_memories --algorithm dbscan
```

### Problem: Low Quality Scores

**Symptoms:**
- Most memories scored < 0.5
- Good memories have low scores

**Solutions:**
1. Add more metadata (tags, relationships)
2. Access memories more frequently
3. Extract knowledge to build relationships
```bash
nexs-mcp extract_knowledge_graph --element-type memory
```
4. Adjust quality factors (advanced)

### Problem: Search Results Not Relevant

**Symptoms:**
- Search returns unrelated results
- Missing expected results

**Solutions:**
1. Lower similarity threshold
```bash
nexs-mcp hybrid_search \
  --query "your query" \
  --similarity-threshold 0.6
```

2. Use different search mode
```bash
# Try linear for accuracy
nexs-mcp hybrid_search --query "your query" --mode linear
```

3. Rebuild index
```bash
# Delete index and let it rebuild
rm -rf data/hnsw-index
nexs-mcp hybrid_search --query "your query" --mode auto
```

4. Add more filters
```bash
nexs-mcp hybrid_search \
  --query "your query" \
  --filter-tags "project-alpha" \
  --date-from "2025-11-01"
```

---

## FAQ

### Q: How often should I run consolidation?

**A:** Weekly for most users. Daily if you create 20+ memories per day.

### Q: Will consolidation delete my memories?

**A:** Only if you:
1. Enable retention policies AND
2. Set auto-cleanup=true OR
3. Manually run apply_retention_policy with --no-dry-run

Always use --dry-run first to preview deletions.

### Q: Can I undo consolidation changes?

**A:** 
- **Merges:** No, create backup first
- **Deletions:** No, create backup first
- **Clusters/Knowledge:** Yes, these are metadata only

Always backup before major changes:
```bash
nexs-mcp backup_create --name "before-consolidation"
```

### Q: Which clustering algorithm is better?

**A:**
- **DBSCAN**: Better for most cases, finds natural groups, detects outliers
- **K-means**: Better when you know exactly how many clusters you want

Start with DBSCAN.

### Q: What's a good duplicate threshold?

**A:** 
- 0.98: Very conservative, only catches exact copies
- 0.95: **Recommended**, catches similar content
- 0.90: Aggressive, may find false positives

### Q: How do quality scores work?

**A:** Scores are calculated from:
- 40% content quality (length, structure, formatting)
- 20% recency (how old/recently accessed)
- 20% relationships (connections to other memories)
- 20% access patterns (usage frequency)

### Q: Can I automate everything?

**A:** Yes:
```bash
export NEXS_MEMORY_CONSOLIDATION_AUTO=true
export NEXS_MEMORY_CONSOLIDATION_INTERVAL=168h  # Weekly
export NEXS_MEMORY_RETENTION_AUTO_CLEANUP=true
```

But start manually first to understand the behavior.

### Q: How much does consolidation improve search?

**A:** Typical improvements:
- 40-60% faster search with HNSW
- 20-30% better relevance with clustering
- 50%+ better context with knowledge graphs

### Q: What happens to outliers?

**A:** Outliers are memories that don't fit any cluster. They're kept but marked as outliers. Review them manually - they might be:
- Unique important memories
- Low-quality noise
- Memories needing better tags

### Q: Can I consolidate other element types?

**A:** Yes! Works with:
- Memories (most common)
- Agents
- Personas
- Skills

Use `--element-type` parameter.

### Q: How much storage does consolidation save?

**A:** Typical savings:
- 5-15% from duplicate removal
- 10-20% from retention policies
- Improved search performance (faster = less CPU)

---

## Related Documentation

- [Memory Consolidation Developer Guide](../development/MEMORY_CONSOLIDATION.md)
- [MCP Tools Reference](../api/MCP_TOOLS.md)
- [Consolidation Tools Examples](../api/CONSOLIDATION_TOOLS.md)
- [Configuration Reference](../api/CLI.md)

---

**Version:** 1.3.0  
**Last Updated:** December 26, 2025  
**Maintainer:** NEXS-MCP Team

**Need Help?** See [Troubleshooting](#troubleshooting) or [FAQ](#faq)
