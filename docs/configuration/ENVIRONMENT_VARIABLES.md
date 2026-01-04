# NEXS MCP - Environment Variables Reference

**Version:** v1.4.0
**Date:** January 4, 2026
**Status:** Complete Configuration Reference

---

## Overview

NEXS MCP can be configured entirely through environment variables. This document provides a comprehensive reference of all available configuration options.

## Table of Contents

1. [Core Configuration](#core-configuration)
2. [Storage Configuration](#storage-configuration)
3. [Logging Configuration](#logging-configuration)
4. [NLP Configuration](#nlp-configuration)
5. [Memory Management](#memory-management)
6. [Vector Store & Embeddings](#vector-store--embeddings)
7. [Optimization Features](#optimization-features)
8. [Resources & Caching](#resources--caching)
9. [GitHub Integration](#github-integration)

---

## Core Configuration

### NEXS_DATA_DIR
- **Type:** string
- **Default:** `./.nexs-mcp`
- **Description:** Base directory for all persistent data storage
- **Example:** `NEXS_DATA_DIR=/var/lib/nexs-mcp`

### NEXS_SERVER_NAME
- **Type:** string
- **Default:** `nexs-mcp`
- **Description:** Server identification name shown in logs and metrics
- **Example:** `NEXS_SERVER_NAME=production-mcp-01`

---

## Storage Configuration

### NEXS_STORAGE_TYPE
- **Type:** string
- **Default:** `file`
- **Options:** `file`, `memory`
- **Description:** Storage backend type
- **Example:** `NEXS_STORAGE_TYPE=file`

---

## Logging Configuration

### NEXS_LOG_LEVEL
- **Type:** string
- **Default:** `info`
- **Options:** `debug`, `info`, `warn`, `error`
- **Description:** Minimum log level to output
- **Example:** `NEXS_LOG_LEVEL=debug`

### NEXS_LOG_FORMAT
- **Type:** string
- **Default:** `json`
- **Options:** `json`, `text`
- **Description:** Log output format
- **Example:** `NEXS_LOG_FORMAT=json`

---

## NLP Configuration

### Entity Extraction

#### NEXS_NLP_ENTITY_EXTRACTION_ENABLED
- **Type:** boolean
- **Default:** `false`
- **Description:** Enable ONNX-based advanced entity extraction
- **Example:** `NEXS_NLP_ENTITY_EXTRACTION_ENABLED=true`
- **Note:** Requires ONNX models downloaded

#### NEXS_NLP_ENTITY_EXTRACTION_MODEL_PATH
- **Type:** string
- **Default:** `./models/bert-base-NER`
- **Description:** Path to BERT NER model directory
- **Example:** `NEXS_NLP_ENTITY_EXTRACTION_MODEL_PATH=/opt/models/ner`

#### NEXS_NLP_ENTITY_MIN_CONFIDENCE
- **Type:** float
- **Default:** `0.5`
- **Range:** `0.0` - `1.0`
- **Description:** Minimum confidence threshold for entity extraction
- **Example:** `NEXS_NLP_ENTITY_MIN_CONFIDENCE=0.7`

#### NEXS_NLP_ENTITY_MAX_ENTITIES
- **Type:** integer
- **Default:** `50`
- **Description:** Maximum entities to extract per text
- **Example:** `NEXS_NLP_ENTITY_MAX_ENTITIES=100`

### Sentiment Analysis

#### NEXS_NLP_SENTIMENT_ENABLED
- **Type:** boolean
- **Default:** `false`
- **Description:** Enable ONNX-based sentiment analysis
- **Example:** `NEXS_NLP_SENTIMENT_ENABLED=true`
- **Note:** Requires ONNX models downloaded

#### NEXS_NLP_SENTIMENT_MODEL_PATH
- **Type:** string
- **Default:** `./models/distilbert-base-uncased-finetuned-sst-2-english`
- **Description:** Path to sentiment analysis model directory
- **Example:** `NEXS_NLP_SENTIMENT_MODEL_PATH=/opt/models/sentiment`

#### NEXS_NLP_SENTIMENT_MIN_CONFIDENCE
- **Type:** float
- **Default:** `0.5`
- **Range:** `0.0` - `1.0`
- **Description:** Minimum confidence for sentiment classification
- **Example:** `NEXS_NLP_SENTIMENT_MIN_CONFIDENCE=0.6`

#### NEXS_NLP_SENTIMENT_ANALYZE_EMOTIONS
- **Type:** boolean
- **Default:** `true`
- **Description:** Extract emotional dimensions (joy, anger, sadness, fear, surprise)
- **Example:** `NEXS_NLP_SENTIMENT_ANALYZE_EMOTIONS=true`

### Topic Modeling

#### NEXS_NLP_TOPIC_MODELING_ENABLED
- **Type:** boolean
- **Default:** `false`
- **Description:** Enable LDA/NMF topic modeling
- **Example:** `NEXS_NLP_TOPIC_MODELING_ENABLED=true`

#### NEXS_NLP_TOPIC_MODELING_ALGORITHM
- **Type:** string
- **Default:** `lda`
- **Options:** `lda`, `nmf`
- **Description:** Topic modeling algorithm
- **Example:** `NEXS_NLP_TOPIC_MODELING_ALGORITHM=lda`

#### NEXS_NLP_TOPIC_MIN_TOPICS
- **Type:** integer
- **Default:** `2`
- **Description:** Minimum number of topics to extract
- **Example:** `NEXS_NLP_TOPIC_MIN_TOPICS=3`

#### NEXS_NLP_TOPIC_MAX_TOPICS
- **Type:** integer
- **Default:** `10`
- **Description:** Maximum number of topics to extract
- **Example:** `NEXS_NLP_TOPIC_MAX_TOPICS=15`

#### NEXS_NLP_TOPIC_MAX_KEYWORDS
- **Type:** integer
- **Default:** `10`
- **Description:** Maximum keywords per topic
- **Example:** `NEXS_NLP_TOPIC_MAX_KEYWORDS=15`

#### NEXS_NLP_TOPIC_MIN_DF
- **Type:** integer
- **Default:** `2`
- **Description:** Minimum document frequency for terms
- **Example:** `NEXS_NLP_TOPIC_MIN_DF=3`

---

## Memory Management

### Auto-Save

#### NEXS_AUTO_SAVE_MEMORIES
- **Type:** boolean
- **Default:** `true`
- **Description:** Automatically save memories to disk
- **Example:** `NEXS_AUTO_SAVE_MEMORIES=true`

#### NEXS_AUTO_SAVE_INTERVAL
- **Type:** duration
- **Default:** `5m`
- **Description:** Interval between auto-save operations
- **Example:** `NEXS_AUTO_SAVE_INTERVAL=10m`

### Working Memory

#### NEXS_WORKING_MEMORY_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable two-tier memory system
- **Example:** `NEXS_WORKING_MEMORY_ENABLED=true`

#### NEXS_WORKING_MEMORY_MAX_SIZE
- **Type:** integer
- **Default:** `100`
- **Description:** Maximum items in working memory
- **Example:** `NEXS_WORKING_MEMORY_MAX_SIZE=200`

### Memory Consolidation

#### NEXS_CONSOLIDATION_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable memory consolidation service
- **Example:** `NEXS_CONSOLIDATION_ENABLED=true`

#### NEXS_CONSOLIDATION_WINDOW
- **Type:** duration
- **Default:** `24h`
- **Description:** Time window for consolidation analysis
- **Example:** `NEXS_CONSOLIDATION_WINDOW=48h`

#### NEXS_CONSOLIDATION_MIN_CLUSTER_SIZE
- **Type:** integer
- **Default:** `3`
- **Description:** Minimum memories per cluster
- **Example:** `NEXS_CONSOLIDATION_MIN_CLUSTER_SIZE=5`

#### NEXS_CONSOLIDATION_MAX_CLUSTERS
- **Type:** integer
- **Default:** `10`
- **Description:** Maximum clusters per consolidation
- **Example:** `NEXS_CONSOLIDATION_MAX_CLUSTERS=15`

#### NEXS_CONSOLIDATION_SIMILARITY_THRESHOLD
- **Type:** float
- **Default:** `0.7`
- **Range:** `0.0` - `1.0`
- **Description:** Similarity threshold for clustering
- **Example:** `NEXS_CONSOLIDATION_SIMILARITY_THRESHOLD=0.75`

#### NEXS_CONSOLIDATION_USE_TEMPORAL
- **Type:** boolean
- **Default:** `true`
- **Description:** Consider temporal proximity in clustering
- **Example:** `NEXS_CONSOLIDATION_USE_TEMPORAL=true`

#### NEXS_CONSOLIDATION_TEMPORAL_WEIGHT
- **Type:** float
- **Default:** `0.3`
- **Range:** `0.0` - `1.0`
- **Description:** Weight of temporal similarity
- **Example:** `NEXS_CONSOLIDATION_TEMPORAL_WEIGHT=0.4`

#### NEXS_CONSOLIDATION_AUTO_PROMOTE
- **Type:** boolean
- **Default:** `true`
- **Description:** Automatically promote consolidated memories
- **Example:** `NEXS_CONSOLIDATION_AUTO_PROMOTE=true`

#### NEXS_CONSOLIDATION_PRESERVE_ORIGINALS
- **Type:** boolean
- **Default:** `false`
- **Description:** Keep original memories after consolidation
- **Example:** `NEXS_CONSOLIDATION_PRESERVE_ORIGINALS=true`

### Memory Retention

#### NEXS_RETENTION_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable memory retention policies
- **Example:** `NEXS_RETENTION_ENABLED=true`

#### NEXS_RETENTION_MAX_AGE
- **Type:** duration
- **Default:** `2160h` (90 days)
- **Description:** Maximum age for memories
- **Example:** `NEXS_RETENTION_MAX_AGE=720h`

#### NEXS_RETENTION_MIN_IMPORTANCE
- **Type:** float
- **Default:** `0.3`
- **Range:** `0.0` - `1.0`
- **Description:** Minimum importance score to retain
- **Example:** `NEXS_RETENTION_MIN_IMPORTANCE=0.5`

#### NEXS_RETENTION_MIN_ACCESS_COUNT
- **Type:** integer
- **Default:** `1`
- **Description:** Minimum access count to retain
- **Example:** `NEXS_RETENTION_MIN_ACCESS_COUNT=2`

#### NEXS_RETENTION_PROTECT_RECENT
- **Type:** duration
- **Default:** `168h` (7 days)
- **Description:** Protect memories created within this period
- **Example:** `NEXS_RETENTION_PROTECT_RECENT=336h`

#### NEXS_RETENTION_PROTECT_IMPORTANT
- **Type:** float
- **Default:** `0.8`
- **Range:** `0.0` - `1.0`
- **Description:** Protect memories above this importance
- **Example:** `NEXS_RETENTION_PROTECT_IMPORTANT=0.9`

#### NEXS_RETENTION_DRY_RUN
- **Type:** boolean
- **Default:** `false`
- **Description:** Simulate retention without deleting
- **Example:** `NEXS_RETENTION_DRY_RUN=true`

#### NEXS_RETENTION_BATCH_SIZE
- **Type:** integer
- **Default:** `100`
- **Description:** Number of memories to process per batch
- **Example:** `NEXS_RETENTION_BATCH_SIZE=200`

### Duplicate Detection

#### NEXS_DUPLICATE_DETECTION_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable duplicate memory detection
- **Example:** `NEXS_DUPLICATE_DETECTION_ENABLED=true`

#### NEXS_DUPLICATE_SIMILARITY_THRESHOLD
- **Type:** float
- **Default:** `0.95`
- **Range:** `0.0` - `1.0`
- **Description:** Similarity threshold to consider duplicates
- **Example:** `NEXS_DUPLICATE_SIMILARITY_THRESHOLD=0.92`

#### NEXS_DUPLICATE_MIN_CONTENT_LENGTH
- **Type:** integer
- **Default:** `20`
- **Description:** Minimum content length to check
- **Example:** `NEXS_DUPLICATE_MIN_CONTENT_LENGTH=50`

#### NEXS_DUPLICATE_MAX_RESULTS
- **Type:** integer
- **Default:** `100`
- **Description:** Maximum duplicate results to return
- **Example:** `NEXS_DUPLICATE_MAX_RESULTS=50`

### Clustering

#### NEXS_CLUSTERING_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable memory clustering
- **Example:** `NEXS_CLUSTERING_ENABLED=true`

#### NEXS_CLUSTERING_ALGORITHM
- **Type:** string
- **Default:** `kmeans`
- **Options:** `kmeans`, `dbscan`, `hierarchical`
- **Description:** Clustering algorithm
- **Example:** `NEXS_CLUSTERING_ALGORITHM=dbscan`

#### NEXS_CLUSTERING_MIN_CLUSTER_SIZE
- **Type:** integer
- **Default:** `3`
- **Description:** Minimum cluster size
- **Example:** `NEXS_CLUSTERING_MIN_CLUSTER_SIZE=5`

#### NEXS_CLUSTERING_MAX_CLUSTERS
- **Type:** integer
- **Default:** `10`
- **Description:** Maximum number of clusters
- **Example:** `NEXS_CLUSTERING_MAX_CLUSTERS=20`

#### NEXS_CLUSTERING_DISTANCE_THRESHOLD
- **Type:** float
- **Default:** `0.5`
- **Range:** `0.0` - `1.0`
- **Description:** Distance threshold for DBSCAN
- **Example:** `NEXS_CLUSTERING_DISTANCE_THRESHOLD=0.4`

#### NEXS_CLUSTERING_USE_KEYWORDS
- **Type:** boolean
- **Default:** `true`
- **Description:** Use keywords for clustering
- **Example:** `NEXS_CLUSTERING_USE_KEYWORDS=true`

### Knowledge Graph

#### NEXS_KNOWLEDGE_GRAPH_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable knowledge graph extraction
- **Example:** `NEXS_KNOWLEDGE_GRAPH_ENABLED=true`

#### NEXS_KNOWLEDGE_GRAPH_MIN_CONFIDENCE
- **Type:** float
- **Default:** `0.5`
- **Range:** `0.0` - `1.0`
- **Description:** Minimum confidence for relationships
- **Example:** `NEXS_KNOWLEDGE_GRAPH_MIN_CONFIDENCE=0.6`

#### NEXS_KNOWLEDGE_GRAPH_MAX_ENTITIES
- **Type:** integer
- **Default:** `50`
- **Description:** Maximum entities to extract
- **Example:** `NEXS_KNOWLEDGE_GRAPH_MAX_ENTITIES=100`

#### NEXS_KNOWLEDGE_GRAPH_MAX_RELATIONS
- **Type:** integer
- **Default:** `100`
- **Description:** Maximum relations to extract
- **Example:** `NEXS_KNOWLEDGE_GRAPH_MAX_RELATIONS=200`

#### NEXS_KNOWLEDGE_GRAPH_EXTRACT_CONCEPTS
- **Type:** boolean
- **Default:** `true`
- **Description:** Extract concepts/topics
- **Example:** `NEXS_KNOWLEDGE_GRAPH_EXTRACT_CONCEPTS=true`

#### NEXS_KNOWLEDGE_GRAPH_INFER_RELATIONS
- **Type:** boolean
- **Default:** `true`
- **Description:** Infer implicit relationships
- **Example:** `NEXS_KNOWLEDGE_GRAPH_INFER_RELATIONS=true`

#### NEXS_KNOWLEDGE_GRAPH_USE_COREFERENCE
- **Type:** boolean
- **Default:** `false`
- **Description:** Resolve coreferences (experimental)
- **Example:** `NEXS_KNOWLEDGE_GRAPH_USE_COREFERENCE=true`

#### NEXS_KNOWLEDGE_GRAPH_MIN_ENTITY_LENGTH
- **Type:** integer
- **Default:** `2`
- **Description:** Minimum entity name length
- **Example:** `NEXS_KNOWLEDGE_GRAPH_MIN_ENTITY_LENGTH=3`

#### NEXS_KNOWLEDGE_GRAPH_MAX_DEPTH
- **Type:** integer
- **Default:** `3`
- **Description:** Maximum graph traversal depth
- **Example:** `NEXS_KNOWLEDGE_GRAPH_MAX_DEPTH=5`

#### NEXS_KNOWLEDGE_GRAPH_TEMPORAL_DECAY
- **Type:** boolean
- **Default:** `true`
- **Description:** Apply temporal decay to relationships
- **Example:** `NEXS_KNOWLEDGE_GRAPH_TEMPORAL_DECAY=true`

---

## Vector Store & Embeddings

### Vector Store

#### NEXS_VECTOR_STORE_TYPE
- **Type:** string
- **Default:** `memory`
- **Options:** `memory`, `disk`
- **Description:** Vector store backend
- **Example:** `NEXS_VECTOR_STORE_TYPE=disk`

#### NEXS_VECTOR_STORE_DIMENSION
- **Type:** integer
- **Default:** `384`
- **Description:** Embedding dimension size
- **Example:** `NEXS_VECTOR_STORE_DIMENSION=768`

#### NEXS_VECTOR_STORE_PATH
- **Type:** string
- **Default:** `./data/vectors`
- **Description:** Path for disk-based vector storage
- **Example:** `NEXS_VECTOR_STORE_PATH=/var/lib/nexs-mcp/vectors`

### HNSW Index

#### NEXS_HNSW_M
- **Type:** integer
- **Default:** `16`
- **Description:** Number of bi-directional links per element
- **Example:** `NEXS_HNSW_M=32`
- **Note:** Higher values improve recall but increase memory

#### NEXS_HNSW_EF_CONSTRUCTION
- **Type:** integer
- **Default:** `200`
- **Description:** Size of dynamic candidate list during construction
- **Example:** `NEXS_HNSW_EF_CONSTRUCTION=400`

#### NEXS_HNSW_EF_SEARCH
- **Type:** integer
- **Default:** `50`
- **Description:** Size of dynamic candidate list during search
- **Example:** `NEXS_HNSW_EF_SEARCH=100`

#### NEXS_HNSW_MAX_ELEMENTS
- **Type:** integer
- **Default:** `10000`
- **Description:** Maximum elements in index
- **Example:** `NEXS_HNSW_MAX_ELEMENTS=100000`

#### NEXS_HNSW_SEED
- **Type:** integer
- **Default:** `100`
- **Description:** Random seed for reproducibility
- **Example:** `NEXS_HNSW_SEED=42`

### Embeddings

#### NEXS_EMBEDDINGS_PROVIDER
- **Type:** string
- **Default:** `onnx`
- **Options:** `onnx`, `openai`, `local`
- **Description:** Embedding provider
- **Example:** `NEXS_EMBEDDINGS_PROVIDER=onnx`

#### NEXS_EMBEDDINGS_MODEL
- **Type:** string
- **Default:** `paraphrase-multilingual-MiniLM-L12-v2`
- **Description:** Embedding model name
- **Example:** `NEXS_EMBEDDINGS_MODEL=all-MiniLM-L6-v2`

#### NEXS_EMBEDDINGS_DIMENSION
- **Type:** integer
- **Default:** `384`
- **Description:** Embedding vector dimension
- **Example:** `NEXS_EMBEDDINGS_DIMENSION=768`

#### NEXS_EMBEDDINGS_BATCH_SIZE
- **Type:** integer
- **Default:** `32`
- **Description:** Batch size for embedding generation
- **Example:** `NEXS_EMBEDDINGS_BATCH_SIZE=64`

#### NEXS_EMBEDDINGS_MAX_LENGTH
- **Type:** integer
- **Default:** `512`
- **Description:** Maximum token length for embeddings
- **Example:** `NEXS_EMBEDDINGS_MAX_LENGTH=1024`

#### NEXS_EMBEDDINGS_NORMALIZE
- **Type:** boolean
- **Default:** `true`
- **Description:** Normalize embedding vectors
- **Example:** `NEXS_EMBEDDINGS_NORMALIZE=true`

### Hybrid Search

#### NEXS_HYBRID_SEARCH_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable hybrid semantic + keyword search
- **Example:** `NEXS_HYBRID_SEARCH_ENABLED=true`

#### NEXS_HYBRID_SEARCH_SEMANTIC_WEIGHT
- **Type:** float
- **Default:** `0.7`
- **Range:** `0.0` - `1.0`
- **Description:** Weight for semantic search results
- **Example:** `NEXS_HYBRID_SEARCH_SEMANTIC_WEIGHT=0.6`

#### NEXS_HYBRID_SEARCH_KEYWORD_WEIGHT
- **Type:** float
- **Default:** `0.3`
- **Range:** `0.0` - `1.0`
- **Description:** Weight for keyword search results
- **Example:** `NEXS_HYBRID_SEARCH_KEYWORD_WEIGHT=0.4`

#### NEXS_HYBRID_SEARCH_MIN_SCORE
- **Type:** float
- **Default:** `0.3`
- **Range:** `0.0` - `1.0`
- **Description:** Minimum hybrid score threshold
- **Example:** `NEXS_HYBRID_SEARCH_MIN_SCORE=0.4`

#### NEXS_HYBRID_SEARCH_RERANK
- **Type:** boolean
- **Default:** `true`
- **Description:** Rerank results using hybrid scoring
- **Example:** `NEXS_HYBRID_SEARCH_RERANK=true`

#### NEXS_HYBRID_SEARCH_BOOST_RECENT
- **Type:** boolean
- **Default:** `true`
- **Description:** Boost recent results in ranking
- **Example:** `NEXS_HYBRID_SEARCH_BOOST_RECENT=true`

#### NEXS_HYBRID_SEARCH_RECENCY_WEIGHT
- **Type:** float
- **Default:** `0.1`
- **Range:** `0.0` - `1.0`
- **Description:** Weight for recency in ranking
- **Example:** `NEXS_HYBRID_SEARCH_RECENCY_WEIGHT=0.15`

---

## Optimization Features

### Response Compression

#### NEXS_COMPRESSION_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable response compression
- **Example:** `NEXS_COMPRESSION_ENABLED=true`

#### NEXS_COMPRESSION_ALGORITHM
- **Type:** string
- **Default:** `gzip`
- **Options:** `gzip`, `zlib`
- **Description:** Compression algorithm
- **Example:** `NEXS_COMPRESSION_ALGORITHM=gzip`

#### NEXS_COMPRESSION_MIN_SIZE
- **Type:** integer
- **Default:** `1024` (1KB)
- **Description:** Minimum response size to compress (bytes)
- **Example:** `NEXS_COMPRESSION_MIN_SIZE=2048`

#### NEXS_COMPRESSION_LEVEL
- **Type:** integer
- **Default:** `6`
- **Range:** `1` - `9`
- **Description:** Compression level (1=fast, 9=best)
- **Example:** `NEXS_COMPRESSION_LEVEL=9`

#### NEXS_COMPRESSION_ADAPTIVE
- **Type:** boolean
- **Default:** `true`
- **Description:** Automatically select best algorithm
- **Example:** `NEXS_COMPRESSION_ADAPTIVE=true`

### Streaming

#### NEXS_STREAMING_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable streaming responses
- **Example:** `NEXS_STREAMING_ENABLED=true`

#### NEXS_STREAMING_CHUNK_SIZE
- **Type:** integer
- **Default:** `10`
- **Description:** Items per streaming chunk
- **Example:** `NEXS_STREAMING_CHUNK_SIZE=20`

#### NEXS_STREAMING_THROTTLE
- **Type:** duration
- **Default:** `50ms`
- **Description:** Delay between chunks
- **Example:** `NEXS_STREAMING_THROTTLE=100ms`

#### NEXS_STREAMING_BUFFER_SIZE
- **Type:** integer
- **Default:** `100`
- **Description:** Stream buffer size
- **Example:** `NEXS_STREAMING_BUFFER_SIZE=200`

### Summarization

#### NEXS_SUMMARIZATION_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable automatic summarization
- **Example:** `NEXS_SUMMARIZATION_ENABLED=true`

#### NEXS_SUMMARIZATION_AGE
- **Type:** duration
- **Default:** `168h` (7 days)
- **Description:** Age threshold for summarization
- **Example:** `NEXS_SUMMARIZATION_AGE=336h`

#### NEXS_SUMMARIZATION_MAX_LENGTH
- **Type:** integer
- **Default:** `500`
- **Description:** Maximum summary length (characters)
- **Example:** `NEXS_SUMMARIZATION_MAX_LENGTH=300`

#### NEXS_SUMMARIZATION_RATIO
- **Type:** float
- **Default:** `0.3`
- **Range:** `0.0` - `1.0`
- **Description:** Target compression ratio
- **Example:** `NEXS_SUMMARIZATION_RATIO=0.2`

#### NEXS_SUMMARIZATION_PRESERVE_KEYWORDS
- **Type:** boolean
- **Default:** `true`
- **Description:** Preserve important keywords
- **Example:** `NEXS_SUMMARIZATION_PRESERVE_KEYWORDS=true`

#### NEXS_SUMMARIZATION_EXTRACTIVE
- **Type:** boolean
- **Default:** `true`
- **Description:** Use extractive summarization (vs abstractive)
- **Example:** `NEXS_SUMMARIZATION_EXTRACTIVE=true`

### Prompt Compression

#### NEXS_PROMPT_COMPRESSION_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable prompt compression
- **Example:** `NEXS_PROMPT_COMPRESSION_ENABLED=true`

#### NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY
- **Type:** boolean
- **Default:** `true`
- **Description:** Remove redundant phrases
- **Example:** `NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY=true`

#### NEXS_PROMPT_COMPRESSION_WHITESPACE
- **Type:** boolean
- **Default:** `true`
- **Description:** Compress whitespace
- **Example:** `NEXS_PROMPT_COMPRESSION_WHITESPACE=true`

#### NEXS_PROMPT_COMPRESSION_ALIASES
- **Type:** boolean
- **Default:** `true`
- **Description:** Replace terms with shorter aliases
- **Example:** `NEXS_PROMPT_COMPRESSION_ALIASES=true`

#### NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE
- **Type:** boolean
- **Default:** `true`
- **Description:** Preserve semantic structure
- **Example:** `NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE=true`

#### NEXS_PROMPT_COMPRESSION_RATIO
- **Type:** float
- **Default:** `0.65`
- **Range:** `0.0` - `1.0`
- **Description:** Target compression ratio
- **Example:** `NEXS_PROMPT_COMPRESSION_RATIO=0.5`

#### NEXS_PROMPT_COMPRESSION_MIN_LENGTH
- **Type:** integer
- **Default:** `500`
- **Description:** Minimum prompt length to compress
- **Example:** `NEXS_PROMPT_COMPRESSION_MIN_LENGTH=1000`

### Context Enrichment

#### NEXS_CONTEXT_ENRICHMENT_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable context enrichment
- **Example:** `NEXS_CONTEXT_ENRICHMENT_ENABLED=true`

#### NEXS_CONTEXT_ENRICHMENT_MAX_DEPTH
- **Type:** integer
- **Default:** `2`
- **Description:** Maximum relationship depth
- **Example:** `NEXS_CONTEXT_ENRICHMENT_MAX_DEPTH=3`

#### NEXS_CONTEXT_ENRICHMENT_MAX_ITEMS
- **Type:** integer
- **Default:** `50`
- **Description:** Maximum enriched items
- **Example:** `NEXS_CONTEXT_ENRICHMENT_MAX_ITEMS=100`

#### NEXS_CONTEXT_ENRICHMENT_INCLUDE_RELATED
- **Type:** boolean
- **Default:** `true`
- **Description:** Include related elements
- **Example:** `NEXS_CONTEXT_ENRICHMENT_INCLUDE_RELATED=true`

#### NEXS_CONTEXT_ENRICHMENT_INCLUDE_METADATA
- **Type:** boolean
- **Default:** `true`
- **Description:** Include element metadata
- **Example:** `NEXS_CONTEXT_ENRICHMENT_INCLUDE_METADATA=true`

#### NEXS_CONTEXT_ENRICHMENT_BATCH_SIZE
- **Type:** integer
- **Default:** `20`
- **Description:** Batch size for enrichment
- **Example:** `NEXS_CONTEXT_ENRICHMENT_BATCH_SIZE=50`

---

## Resources & Caching

### Resources

#### NEXS_RESOURCES_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable MCP resources endpoint
- **Example:** `NEXS_RESOURCES_ENABLED=true`

#### NEXS_RESOURCES_CACHE_TTL
- **Type:** duration
- **Default:** `5m`
- **Description:** Resource cache time-to-live
- **Example:** `NEXS_RESOURCES_CACHE_TTL=10m`

### Adaptive Cache

#### NEXS_ADAPTIVE_CACHE_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable adaptive caching
- **Example:** `NEXS_ADAPTIVE_CACHE_ENABLED=true`

#### NEXS_ADAPTIVE_CACHE_MIN_TTL
- **Type:** duration
- **Default:** `1h`
- **Description:** Minimum cache TTL
- **Example:** `NEXS_ADAPTIVE_CACHE_MIN_TTL=30m`

#### NEXS_ADAPTIVE_CACHE_MAX_TTL
- **Type:** duration
- **Default:** `168h` (7 days)
- **Description:** Maximum cache TTL
- **Example:** `NEXS_ADAPTIVE_CACHE_MAX_TTL=720h`

#### NEXS_ADAPTIVE_CACHE_BASE_TTL
- **Type:** duration
- **Default:** `24h`
- **Description:** Base cache TTL
- **Example:** `NEXS_ADAPTIVE_CACHE_BASE_TTL=12h`

---

## GitHub Integration

### Authentication

#### NEXS_GITHUB_TOKEN
- **Type:** string
- **Default:** (none)
- **Description:** GitHub personal access token
- **Example:** `NEXS_GITHUB_TOKEN=ghp_xxxxxxxxxxxxx`
- **Required for:** GitHub sync, publish, search

#### NEXS_GITHUB_CLIENT_ID
- **Type:** string
- **Default:** (none)
- **Description:** GitHub OAuth2 client ID
- **Example:** `NEXS_GITHUB_CLIENT_ID=Iv1.abc123def456`

#### NEXS_GITHUB_CLIENT_SECRET
- **Type:** string
- **Default:** (none)
- **Description:** GitHub OAuth2 client secret
- **Example:** `NEXS_GITHUB_CLIENT_SECRET=xxxxxxxxxxxxx`

### Sync Configuration

#### NEXS_GITHUB_SYNC_ENABLED
- **Type:** boolean
- **Default:** `false`
- **Description:** Enable GitHub synchronization
- **Example:** `NEXS_GITHUB_SYNC_ENABLED=true`

#### NEXS_GITHUB_SYNC_INTERVAL
- **Type:** duration
- **Default:** `1h`
- **Description:** Auto-sync interval
- **Example:** `NEXS_GITHUB_SYNC_INTERVAL=30m`

#### NEXS_GITHUB_SYNC_AUTO_PUSH
- **Type:** boolean
- **Default:** `false`
- **Description:** Automatically push changes
- **Example:** `NEXS_GITHUB_SYNC_AUTO_PUSH=true`

---

## Skill Extraction

### NEXS_SKILL_EXTRACTION_ENABLED
- **Type:** boolean
- **Default:** `true`
- **Description:** Enable automatic skill extraction from personas
- **Example:** `NEXS_SKILL_EXTRACTION_ENABLED=true`

### NEXS_SKILL_EXTRACTION_AUTO_ON_CREATE
- **Type:** boolean
- **Default:** `true`
- **Description:** Auto-extract skills when creating personas
- **Example:** `NEXS_SKILL_EXTRACTION_AUTO_ON_CREATE=false`

### NEXS_SKILL_EXTRACTION_SKIP_DUPLICATES
- **Type:** boolean
- **Default:** `true`
- **Description:** Skip duplicate skill creation
- **Example:** `NEXS_SKILL_EXTRACTION_SKIP_DUPLICATES=true`

### NEXS_SKILL_EXTRACTION_MIN_NAME_LENGTH
- **Type:** integer
- **Default:** `3`
- **Description:** Minimum skill name length
- **Example:** `NEXS_SKILL_EXTRACTION_MIN_NAME_LENGTH=5`

### NEXS_SKILL_EXTRACTION_MAX_PER_PERSONA
- **Type:** integer
- **Default:** `50`
- **Description:** Maximum skills per persona
- **Example:** `NEXS_SKILL_EXTRACTION_MAX_PER_PERSONA=100`

### NEXS_SKILL_EXTRACTION_FROM_EXPERTISE
- **Type:** boolean
- **Default:** `true`
- **Description:** Extract from expertise field
- **Example:** `NEXS_SKILL_EXTRACTION_FROM_EXPERTISE=true`

### NEXS_SKILL_EXTRACTION_FROM_CUSTOM
- **Type:** boolean
- **Default:** `true`
- **Description:** Extract from custom fields
- **Example:** `NEXS_SKILL_EXTRACTION_FROM_CUSTOM=true`

### NEXS_SKILL_EXTRACTION_AUTO_UPDATE
- **Type:** boolean
- **Default:** `true`
- **Description:** Auto-update skills when persona changes
- **Example:** `NEXS_SKILL_EXTRACTION_AUTO_UPDATE=false`

---

## Configuration Examples

### Minimal Production Setup
```bash
export NEXS_DATA_DIR=/var/lib/nexs-mcp
export NEXS_STORAGE_TYPE=file
export NEXS_LOG_LEVEL=info
export NEXS_LOG_FORMAT=json
```

### NLP-Enabled Setup
```bash
export NEXS_NLP_ENTITY_EXTRACTION_ENABLED=true
export NEXS_NLP_SENTIMENT_ENABLED=true
export NEXS_NLP_TOPIC_MODELING_ENABLED=true
export NEXS_NLP_ENTITY_EXTRACTION_MODEL_PATH=/opt/models/bert-ner
export NEXS_NLP_SENTIMENT_MODEL_PATH=/opt/models/distilbert-sentiment
```

### High-Performance Setup
```bash
export NEXS_VECTOR_STORE_TYPE=disk
export NEXS_HNSW_M=32
export NEXS_HNSW_EF_CONSTRUCTION=400
export NEXS_HNSW_EF_SEARCH=100
export NEXS_EMBEDDINGS_BATCH_SIZE=64
export NEXS_COMPRESSION_LEVEL=9
export NEXS_STREAMING_ENABLED=true
```

### Memory Optimization
```bash
export NEXS_CONSOLIDATION_ENABLED=true
export NEXS_CONSOLIDATION_WINDOW=24h
export NEXS_RETENTION_ENABLED=true
export NEXS_RETENTION_MAX_AGE=720h
export NEXS_DUPLICATE_DETECTION_ENABLED=true
export NEXS_CLUSTERING_ENABLED=true
```

### GitHub Integration
```bash
export NEXS_GITHUB_TOKEN=ghp_xxxxxxxxxxxxx
export NEXS_GITHUB_SYNC_ENABLED=true
export NEXS_GITHUB_SYNC_INTERVAL=30m
export NEXS_GITHUB_SYNC_AUTO_PUSH=true
```

---

## Notes

1. **Duration Format**: Uses Go duration format: `1h`, `30m`, `24h`, `168h`
2. **Boolean Values**: Use `true` or `false` (case-insensitive)
3. **Float Values**: Use decimal notation: `0.5`, `0.95`, `0.3`
4. **Validation**: Invalid values will use defaults with warning logs
5. **Priority**: Environment variables override config file settings

For more information, see:
- [Configuration Guide](../user-guide/CONFIGURATION.md)
- [NLP Features](../NLP_FEATURES.md)
- [Token Optimization](../analysis/TOKEN_OPTIMIZATION_GAPS.md)
