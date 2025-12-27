# Memory Consolidation - Developer Guide

**Version:** 1.3.0  
**Last Updated:** December 26, 2025  
**Target Audience:** Developers and Contributors

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Services Deep Dive](#services-deep-dive)
4. [Implementation Patterns](#implementation-patterns)
5. [Adding New Features](#adding-new-features)
6. [Testing Strategies](#testing-strategies)
7. [Performance Optimization](#performance-optimization)
8. [Error Handling](#error-handling)
9. [Best Practices](#best-practices)
10. [Troubleshooting](#troubleshooting)

---

## Overview

Memory Consolidation is a suite of 7 application services that provide advanced memory management capabilities in NEXS-MCP. These services implement state-of-the-art algorithms for duplicate detection, clustering, knowledge extraction, quality scoring, and retention policies.

### Goals

1. **Reduce Duplication** - Detect and merge duplicate or highly similar memories
2. **Improve Organization** - Cluster related memories by topic/project
3. **Extract Knowledge** - Build knowledge graphs with entities and relationships
4. **Maintain Quality** - Score and retain high-quality memories
5. **Optimize Performance** - Fast search with HNSW indexing
6. **Automate Workflows** - End-to-end consolidation orchestration

### Services

| Service | Purpose | Algorithm | LOC | Tests |
|---------|---------|-----------|-----|-------|
| DuplicateDetection | Find and merge duplicates | HNSW similarity | 442 | 15 |
| Clustering | Group related memories | DBSCAN, K-means | 437 | 13 |
| KnowledgeGraphExtractor | Extract entities/relationships | NLP, regex | 518 | 20 |
| MemoryConsolidation | Orchestrate workflow | Pipeline | 583 | 20 |
| HybridSearch | Fast semantic search | HNSW + linear | 530 | 20 |
| MemoryRetention | Quality-based cleanup | Multi-factor scoring | 378 | 15 |
| SemanticSearch | Multi-type indexing | Vector search | 545 | 20 |

**Total:** 3,433 LOC, 123 tests, 76.4% coverage

---

## Architecture

### Layer Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         MCP Layer                            │
│  (mcp/consolidation_tools.go - MCP tool handlers)           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│  (internal/application/*_consolidation*.go)                 │
│                                                              │
│  ┌──────────────────────────────────────────────────┐       │
│  │  MemoryConsolidation (Orchestrator)              │       │
│  │  - Workflow coordination                         │       │
│  │  - Step execution                                │       │
│  │  - Result aggregation                            │       │
│  └──────────────────────────────────────────────────┘       │
│                       │                                      │
│    ┌──────────────────┼──────────────────┐                 │
│    │                  │                  │                  │
│    ▼                  ▼                  ▼                  │
│  ┌──────┐  ┌──────────────┐  ┌───────────────┐            │
│  │ Dup  │  │ Clustering   │  │ Knowledge     │            │
│  │ Det  │  │ DBSCAN/Kmeans│  │ Graph NLP     │            │
│  └──────┘  └──────────────┘  └───────────────┘            │
│    │                  │                  │                  │
│    └──────────────────┼──────────────────┘                 │
│                       │                                      │
│    ┌──────────────────┼──────────────────┐                 │
│    │                  │                  │                  │
│    ▼                  ▼                  ▼                  │
│  ┌──────┐  ┌──────────────┐  ┌───────────────┐            │
│  │Hybrid│  │ Semantic     │  │ Memory        │            │
│  │Search│  │ Search       │  │ Retention     │            │
│  └──────┘  └──────────────┘  └───────────────┘            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       Domain Layer                           │
│  (internal/domain - Element, Memory interfaces)             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                       │
│  - File/Memory storage (internal/infrastructure)            │
│  - HNSW indexing (internal/indexing/hnsw)                   │
│  - Embeddings (internal/embeddings)                         │
│  - Vector store (internal/vectorstore)                      │
└─────────────────────────────────────────────────────────────┘
```

### Component Dependencies

```go
// Service initialization order
embedProvider := embeddings.GetProvider()
vectorStore := vectorstore.NewVectorStore(config.Dimension)

// Core services
duplicateDetector := NewDuplicateDetectionService(repo, embedProvider, vectorStore, config, logger)
clusterer := NewClusteringService(repo, embedProvider, vectorStore, config, logger)
knowledgeExtractor := NewKnowledgeGraphExtractorService(repo, config, logger)
retention := NewMemoryRetentionService(repo, config, logger)

// Orchestrator
consolidator := NewMemoryConsolidationService(
    duplicateDetector,
    clusterer,
    knowledgeExtractor,
    retention,
    config,
    logger,
)
```

### Data Flow

```
User Request
     │
     ▼
MCP Tool Handler
     │
     ▼
MemoryConsolidation Service
     │
     ├─→ Step 1: Duplicate Detection
     │   ├─→ Load elements from repository
     │   ├─→ Generate embeddings
     │   ├─→ Build HNSW index
     │   ├─→ Find similar elements
     │   └─→ Return duplicate groups
     │
     ├─→ Step 2: Clustering
     │   ├─→ Get embeddings
     │   ├─→ Run DBSCAN/K-means
     │   ├─→ Assign cluster labels
     │   └─→ Return clusters + outliers
     │
     ├─→ Step 3: Knowledge Extraction
     │   ├─→ Parse content with NLP
     │   ├─→ Extract entities (people, orgs, URLs)
     │   ├─→ Extract keywords (TF-IDF)
     │   ├─→ Find relationships
     │   └─→ Return knowledge graph
     │
     └─→ Step 4: Quality Scoring
         ├─→ Calculate content quality
         ├─→ Calculate recency score
         ├─→ Calculate relationship score
         ├─→ Calculate access score
         └─→ Return quality scores
     │
     ▼
Aggregate Results
     │
     ▼
Return ConsolidationResult
```

---

## Services Deep Dive

### 1. DuplicateDetectionService

**File:** `internal/application/duplicate_detection.go`

**Purpose:** Detect and merge duplicate elements using HNSW-based vector similarity.

#### Structure

```go
type DuplicateDetectionService struct {
    repository        domain.ElementRepository
    embeddingProvider embeddings.Provider
    vectorStore       *vectorstore.VectorStore
    config            DuplicateDetectionConfig
    logger            *slog.Logger
}

type DuplicateDetectionConfig struct {
    Enabled             bool
    SimilarityThreshold float32  // 0.95 default
    MinContentLength    int      // 20 chars default
    MaxResults          int      // 100 default
}
```

#### Key Methods

```go
// DetectDuplicates finds duplicate groups
func (s *DuplicateDetectionService) DetectDuplicates(
    ctx context.Context,
    elementType string,
    threshold float32,
) ([]DuplicateGroup, error)

// MergeDuplicates merges duplicate elements
func (s *DuplicateDetectionService) MergeDuplicates(
    ctx context.Context,
    groups []DuplicateGroup,
    dryRun bool,
) (*MergeResult, error)

// ComputeSimilarity calculates cosine similarity
func (s *DuplicateDetectionService) ComputeSimilarity(
    vec1, vec2 []float32,
) float32
```

#### Algorithm

```go
func (s *DuplicateDetectionService) DetectDuplicates(
    ctx context.Context,
    elementType string,
    threshold float32,
) ([]DuplicateGroup, error) {
    // 1. Load elements by type
    elements, err := s.repository.List(map[string]interface{}{
        "type": elementType,
    })
    if err != nil {
        return nil, err
    }

    // 2. Filter by minimum content length
    filtered := s.filterByLength(elements, s.config.MinContentLength)

    // 3. Generate embeddings
    embeddings := make(map[string][]float32)
    for _, elem := range filtered {
        content := s.extractContent(elem)
        embedding, err := s.embeddingProvider.Embed(ctx, content)
        if err != nil {
            s.logger.Warn("Failed to generate embedding", "error", err)
            continue
        }
        embeddings[elem.GetID()] = embedding
    }

    // 4. Build HNSW index
    index := s.buildHNSWIndex(embeddings)

    // 5. Find duplicates using HNSW neighbors
    groups := make([]DuplicateGroup, 0)
    visited := make(map[string]bool)

    for id, embedding := range embeddings {
        if visited[id] {
            continue
        }

        // Find k-nearest neighbors
        neighbors, distances := index.Search(embedding, 10)

        // Filter by threshold
        duplicates := []string{id}
        for i, neighborID := range neighbors {
            if neighborID == id {
                continue
            }
            similarity := 1 - distances[i] // Convert distance to similarity
            if similarity >= threshold && !visited[neighborID] {
                duplicates = append(duplicates, neighborID)
                visited[neighborID] = true
            }
        }

        if len(duplicates) > 1 {
            group := s.createDuplicateGroup(duplicates, embeddings)
            groups = append(groups, group)
        }
        visited[id] = true
    }

    return groups, nil
}
```

#### Testing

```go
func TestDuplicateDetectionService_DetectDuplicates(t *testing.T) {
    // Setup
    mockProvider := embeddings.NewMockProvider("mock", 128)
    mockRepo := &mockRepository{elements: make(map[string]domain.Element)}
    config := DuplicateDetectionConfig{
        Enabled:             true,
        SimilarityThreshold: 0.95,
        MinContentLength:    20,
        MaxResults:          100,
    }

    // Create test data
    mem1 := createMemory("mem-1", "Machine learning implementation guide")
    mem2 := createMemory("mem-2", "Machine learning implementation tutorial")
    mem3 := createMemory("mem-3", "Completely different content")
    mockRepo.elements["mem-1"] = mem1
    mockRepo.elements["mem-2"] = mem2
    mockRepo.elements["mem-3"] = mem3

    // Initialize service
    service := NewDuplicateDetectionService(mockRepo, mockProvider, nil, config, nil)

    // Execute
    groups, err := service.DetectDuplicates(context.Background(), "memory", 0.90)

    // Assertions
    assert.NoError(t, err)
    assert.Len(t, groups, 1, "Should find 1 duplicate group")
    assert.Len(t, groups[0].Elements, 2, "Group should have 2 elements")
    assert.Greater(t, groups[0].Similarity, float32(0.90))
}
```

---

### 2. ClusteringService

**File:** `internal/application/clustering.go`

**Purpose:** Group related memories using DBSCAN or K-means clustering.

#### Algorithms

**DBSCAN (Density-Based Spatial Clustering of Applications with Noise)**

```go
func (s *ClusteringService) ClusterDBSCAN(
    ctx context.Context,
    elementType string,
    minClusterSize int,
    epsilon float32,
) ([]Cluster, []Outlier, error) {
    // 1. Get embeddings
    embeddings, elements := s.getEmbeddings(ctx, elementType)

    // 2. Initialize cluster labels (-1 = unvisited, 0 = outlier, >0 = cluster)
    labels := make(map[string]int)
    for id := range embeddings {
        labels[id] = -1
    }

    clusterID := 0

    // 3. For each point
    for id, embedding := range embeddings {
        if labels[id] != -1 {
            continue // Already visited
        }

        // Find neighbors within epsilon
        neighbors := s.findNeighbors(embedding, embeddings, epsilon)

        if len(neighbors) < minClusterSize {
            labels[id] = 0 // Mark as outlier
            continue
        }

        // Start new cluster
        clusterID++
        s.expandCluster(id, neighbors, labels, embeddings, clusterID, minClusterSize, epsilon)
    }

    // 4. Build clusters from labels
    clusters, outliers := s.buildClustersFromLabels(labels, elements, embeddings)

    return clusters, outliers, nil
}

func (s *ClusteringService) expandCluster(
    pointID string,
    neighbors []string,
    labels map[string]int,
    embeddings map[string][]float32,
    clusterID int,
    minClusterSize int,
    epsilon float32,
) {
    labels[pointID] = clusterID

    i := 0
    for i < len(neighbors) {
        neighborID := neighbors[i]

        if labels[neighborID] == -1 {
            // Unvisited - mark as part of cluster
            labels[neighborID] = clusterID

            // Find neighbors of neighbor
            newNeighbors := s.findNeighbors(embeddings[neighborID], embeddings, epsilon)
            if len(newNeighbors) >= minClusterSize {
                neighbors = append(neighbors, newNeighbors...)
            }
        } else if labels[neighborID] == 0 {
            // Outlier - add to cluster
            labels[neighborID] = clusterID
        }

        i++
    }
}
```

**K-means Clustering**

```go
func (s *ClusteringService) ClusterKMeans(
    ctx context.Context,
    elementType string,
    numClusters int,
    maxIterations int,
) ([]Cluster, error) {
    // 1. Get embeddings
    embeddings, elements := s.getEmbeddings(ctx, elementType)
    dimension := len(embeddings[firstKey(embeddings)])

    // 2. Initialize centroids randomly
    centroids := s.initializeCentroids(embeddings, numClusters)

    // 3. Iterate until convergence or max iterations
    for iter := 0; iter < maxIterations; iter++ {
        // Assign points to nearest centroid
        assignments := make(map[string]int)
        for id, embedding := range embeddings {
            nearest := s.findNearestCentroid(embedding, centroids)
            assignments[id] = nearest
        }

        // Update centroids
        newCentroids := s.updateCentroids(embeddings, assignments, numClusters, dimension)

        // Check convergence
        if s.centroidsConverged(centroids, newCentroids) {
            break
        }

        centroids = newCentroids
    }

    // 4. Build clusters
    clusters := s.buildClustersFromAssignments(assignments, elements, embeddings, centroids)

    return clusters, nil
}
```

#### Cluster Quality Metrics

```go
// Silhouette score: measures how similar an object is to its own cluster
// compared to other clusters. Range: [-1, 1], higher is better.
func (s *ClusteringService) CalculateSilhouetteScore(
    clusters []Cluster,
    embeddings map[string][]float32,
) float32 {
    totalScore := float32(0)
    totalPoints := 0

    for _, cluster := range clusters {
        for _, member := range cluster.Members {
            // a: average distance to points in same cluster
            a := s.averageIntraClusterDistance(member.ID, cluster, embeddings)

            // b: average distance to points in nearest other cluster
            b := s.averageNearestClusterDistance(member.ID, cluster, clusters, embeddings)

            // Silhouette coefficient for this point
            s_i := (b - a) / max(a, b)
            totalScore += s_i
            totalPoints++
        }
    }

    return totalScore / float32(totalPoints)
}
```

---

### 3. KnowledgeGraphExtractorService

**File:** `internal/application/knowledge_graph_extractor.go`

**Purpose:** Extract entities, relationships, and keywords using NLP.

#### Entity Extraction

```go
// Extract person names using pattern matching
func (s *KnowledgeGraphExtractorService) ExtractPeople(
    content string,
) []Person {
    people := make([]Person, 0)

    // Pattern: Capitalized words (potential names)
    // John Smith, Dr. Jane Doe, etc.
    namePattern := regexp.MustCompile(`\b[A-Z][a-z]+(?: [A-Z][a-z]+)+\b`)
    matches := namePattern.FindAllString(content, -1)

    // Filter common false positives
    for _, match := range matches {
        if s.isPotentialPersonName(match) {
            people = append(people, Person{
                Name:     match,
                Mentions: strings.Count(content, match),
            })
        }
    }

    return people
}

// Extract organizations
func (s *KnowledgeGraphExtractorService) ExtractOrganizations(
    content string,
) []Organization {
    orgs := make([]Organization, 0)

    // Patterns: Corp, Inc, Ltd, LLC, Company, etc.
    orgPatterns := []*regexp.Regexp{
        regexp.MustCompile(`\b[A-Z][a-z]+(?: [A-Z][a-z]+)* (?:Corp|Inc|Ltd|LLC|Company)\b`),
        regexp.MustCompile(`\b[A-Z]{2,}(?: [A-Z]{2,})*\b`), // Acronyms
    }

    for _, pattern := range orgPatterns {
        matches := pattern.FindAllString(content, -1)
        for _, match := range matches {
            orgs = append(orgs, Organization{
                Name:     match,
                Mentions: strings.Count(content, match),
                Type:     s.classifyOrganizationType(match),
            })
        }
    }

    return orgs
}

// Extract URLs
func (s *KnowledgeGraphExtractorService) ExtractURLs(
    content string,
) []URL {
    urlPattern := regexp.MustCompile(`https?://[^\s]+`)
    matches := urlPattern.FindAllString(content, -1)

    urls := make([]URL, 0, len(matches))
    for _, match := range matches {
        urls = append(urls, URL{
            URL:      match,
            Mentions: strings.Count(content, match),
            Context:  s.extractURLContext(content, match),
        })
    }

    return urls
}

// Extract emails
func (s *KnowledgeGraphExtractorService) ExtractEmails(
    content string,
) []Email {
    emailPattern := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
    matches := emailPattern.FindAllString(content, -1)

    emails := make([]Email, 0, len(matches))
    for _, match := range matches {
        emails = append(emails, Email{
            Email:    match,
            Mentions: strings.Count(content, match),
        })
    }

    return emails
}
```

#### Keyword Extraction (TF-IDF)

```go
func (s *KnowledgeGraphExtractorService) ExtractKeywords(
    content string,
    maxKeywords int,
) []Keyword {
    // 1. Tokenize and normalize
    words := s.tokenize(content)
    words = s.removeStopWords(words)
    words = s.stem(words)

    // 2. Calculate term frequency (TF)
    tf := make(map[string]float64)
    for _, word := range words {
        tf[word]++
    }
    totalWords := float64(len(words))
    for word := range tf {
        tf[word] = tf[word] / totalWords
    }

    // 3. Calculate inverse document frequency (IDF)
    // Using corpus statistics if available
    idf := s.getIDF(tf)

    // 4. Calculate TF-IDF score
    tfidf := make(map[string]float64)
    for word := range tf {
        tfidf[word] = tf[word] * idf[word]
    }

    // 5. Sort by score and return top N
    keywords := s.sortByScore(tfidf)
    if len(keywords) > maxKeywords {
        keywords = keywords[:maxKeywords]
    }

    return keywords
}
```

#### Relationship Extraction

```go
func (s *KnowledgeGraphExtractorService) ExtractRelationships(
    entities KnowledgeGraph,
    content string,
) []Relationship {
    relationships := make([]Relationship, 0)

    // 1. Person → Project relationships
    // Pattern: "X works on Y", "X is working on Y"
    for _, person := range entities.People {
        projects := s.findProjects(content)
        for _, project := range projects {
            if s.coOccurs(person.Name, project, content, 50) {
                relationships = append(relationships, Relationship{
                    FromEntity: person.Name,
                    ToEntity:   project,
                    Type:       "works_on",
                    Strength:   s.calculateStrength(person.Name, project, content),
                })
            }
        }
    }

    // 2. Person → Person relationships
    // Pattern: "X and Y collaborated", "X worked with Y"
    for i, person1 := range entities.People {
        for j := i + 1; j < len(entities.People); j++ {
            person2 := entities.People[j]
            if s.coOccurs(person1.Name, person2.Name, content, 30) {
                relationships = append(relationships, Relationship{
                    FromEntity: person1.Name,
                    ToEntity:   person2.Name,
                    Type:       "collaborates_with",
                    Strength:   s.calculateStrength(person1.Name, person2.Name, content),
                })
            }
        }
    }

    // 3. Person → Organization relationships
    for _, person := range entities.People {
        for _, org := range entities.Organizations {
            if s.coOccurs(person.Name, org.Name, content, 40) {
                relationships = append(relationships, Relationship{
                    FromEntity: person.Name,
                    ToEntity:   org.Name,
                    Type:       "belongs_to",
                    Strength:   s.calculateStrength(person.Name, org.Name, content),
                })
            }
        }
    }

    return relationships
}

// Check if two entities co-occur within a window
func (s *KnowledgeGraphExtractorService) coOccurs(
    entity1, entity2, content string,
    windowSize int,
) bool {
    pos1 := strings.Index(content, entity1)
    pos2 := strings.Index(content, entity2)

    if pos1 == -1 || pos2 == -1 {
        return false
    }

    distance := abs(pos1 - pos2)
    return distance <= windowSize
}
```

---

### 4. MemoryConsolidationService

**File:** `internal/application/memory_consolidation.go`

**Purpose:** Orchestrate complete consolidation workflow.

#### Workflow Execution

```go
func (s *MemoryConsolidationService) ConsolidateMemories(
    ctx context.Context,
    req ConsolidationRequest,
) (*ConsolidationResult, error) {
    startTime := time.Now()
    workflowID := fmt.Sprintf("consolidation-%s", time.Now().Format("20060102-150405"))

    result := &ConsolidationResult{
        WorkflowID: workflowID,
        StepsExecuted: make(map[string]StepResult),
        Recommendations: make([]string, 0),
    }

    // Step 1: Duplicate Detection
    if req.EnableDuplicateDetection && s.config.EnableDuplicateDetection {
        stepStart := time.Now()
        s.logger.Info("Starting duplicate detection", "workflow_id", workflowID)

        duplicates, err := s.duplicateDetection.DetectDuplicates(
            ctx,
            req.ElementType,
            s.duplicateDetection.config.SimilarityThreshold,
        )

        stepResult := StepResult{
            Status:     "completed",
            DurationMs: time.Since(stepStart).Milliseconds(),
        }

        if err != nil {
            stepResult.Status = "failed"
            stepResult.Error = err.Error()
        } else {
            stepResult.DuplicatesFound = len(duplicates)
            // Merge if not dry run
            if !req.DryRun {
                merged, _ := s.duplicateDetection.MergeDuplicates(ctx, duplicates, false)
                stepResult.Merged = merged.MergedCount
            }
        }

        result.StepsExecuted["duplicate_detection"] = stepResult
    }

    // Step 2: Clustering
    if req.EnableClustering && s.config.EnableClustering {
        stepStart := time.Now()
        s.logger.Info("Starting clustering", "workflow_id", workflowID)

        var clusters []Cluster
        var outliers []Outlier
        var err error

        if s.clustering.config.Algorithm == "dbscan" {
            clusters, outliers, err = s.clustering.ClusterDBSCAN(
                ctx,
                req.ElementType,
                s.clustering.config.MinClusterSize,
                s.clustering.config.EpsilonDistance,
            )
        } else {
            clusters, err = s.clustering.ClusterKMeans(
                ctx,
                req.ElementType,
                s.clustering.config.NumClusters,
                s.clustering.config.MaxIterations,
            )
        }

        stepResult := StepResult{
            Status:     "completed",
            Algorithm:  s.clustering.config.Algorithm,
            DurationMs: time.Since(stepStart).Milliseconds(),
        }

        if err != nil {
            stepResult.Status = "failed"
            stepResult.Error = err.Error()
        } else {
            stepResult.ClustersCreated = len(clusters)
            stepResult.Outliers = len(outliers)
            // Calculate metrics
            if len(clusters) > 0 {
                result.Summary.NewClusters = len(clusters)
            }
        }

        result.StepsExecuted["clustering"] = stepResult
    }

    // Step 3: Knowledge Extraction
    if req.EnableKnowledgeExtraction && s.config.EnableKnowledgeExtraction {
        stepStart := time.Now()
        s.logger.Info("Starting knowledge extraction", "workflow_id", workflowID)

        graph, err := s.knowledgeExtractor.ExtractKnowledgeGraph(
            ctx,
            req.ElementType,
        )

        stepResult := StepResult{
            Status:     "completed",
            DurationMs: time.Since(stepStart).Milliseconds(),
        }

        if err != nil {
            stepResult.Status = "failed"
            stepResult.Error = err.Error()
        } else {
            stepResult.EntitiesExtracted = len(graph.Entities.People) +
                len(graph.Entities.Organizations) +
                len(graph.Entities.Concepts)
            stepResult.RelationshipsCreated = len(graph.Relationships)
            result.Summary.NewRelationships = len(graph.Relationships)
        }

        result.StepsExecuted["knowledge_extraction"] = stepResult
    }

    // Step 4: Quality Scoring
    if req.EnableQualityScoring && s.config.EnableQualityScoring {
        stepStart := time.Now()
        s.logger.Info("Starting quality scoring", "workflow_id", workflowID)

        scores, err := s.qualityScorer.ScoreMemories(
            ctx,
            req.ElementType,
            req.MinQuality,
        )

        stepResult := StepResult{
            Status:     "completed",
            DurationMs: time.Since(stepStart).Milliseconds(),
        }

        if err != nil {
            stepResult.Status = "failed"
            stepResult.Error = err.Error()
        } else {
            stepResult.MemoriesScored = len(scores)
            stepResult.AvgQuality = s.calculateAvgQuality(scores)
            stepResult.HighQuality = s.countHighQuality(scores, 0.7)
            stepResult.LowQuality = s.countLowQuality(scores, req.MinQuality)
        }

        result.StepsExecuted["quality_scoring"] = stepResult
    }

    // Generate recommendations
    result.Recommendations = s.generateRecommendations(result)

    // Calculate totals
    result.Duration = time.Since(startTime)
    result.Summary = s.buildSummary(result)

    s.logger.Info("Consolidation completed",
        "workflow_id", workflowID,
        "duration_ms", result.Duration.Milliseconds())

    return result, nil
}
```

---

## Implementation Patterns

### 1. Service Initialization

```go
// Standard pattern for all consolidation services
func NewServiceName(
    repository domain.ElementRepository,
    embeddingProvider embeddings.Provider,
    config ServiceConfig,
    logger *slog.Logger,
) *ServiceName {
    if logger == nil {
        logger = logger.Get()
    }

    return &ServiceName{
        repository:        repository,
        embeddingProvider: embeddingProvider,
        config:            config,
        logger:            logger,
    }
}
```

### 2. Context Propagation

```go
// Always propagate context for cancellation
func (s *Service) Method(ctx context.Context, params Params) (*Result, error) {
    // Check context
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Pass context to downstream calls
    result, err := s.repository.Get(ctx, id)
    if err != nil {
        return nil, err
    }

    return result, nil
}
```

### 3. Error Handling

```go
// Wrap errors with context
func (s *Service) Process(ctx context.Context) error {
    elements, err := s.repository.List(ctx, filters)
    if err != nil {
        return fmt.Errorf("failed to list elements: %w", err)
    }

    // Domain errors
    if len(elements) == 0 {
        return domain.ErrElementNotFound
    }

    return nil
}
```

### 4. Logging

```go
// Structured logging with slog
func (s *Service) Execute(ctx context.Context) error {
    s.logger.Info("Starting execution",
        "service", "ServiceName",
        "timestamp", time.Now())

    result, err := s.process(ctx)
    if err != nil {
        s.logger.Error("Execution failed",
            "error", err,
            "service", "ServiceName")
        return err
    }

    s.logger.Info("Execution completed",
        "service", "ServiceName",
        "duration_ms", result.Duration.Milliseconds(),
        "items_processed", result.Count)

    return nil
}
```

### 5. Configuration

```go
// Load from environment with defaults
type ServiceConfig struct {
    Enabled   bool
    Threshold float32
    MaxItems  int
}

func LoadConfig() ServiceConfig {
    return ServiceConfig{
        Enabled:   getEnvBool("SERVICE_ENABLED", true),
        Threshold: getEnvFloat32("SERVICE_THRESHOLD", 0.7),
        MaxItems:  getEnvInt("SERVICE_MAX_ITEMS", 100),
    }
}
```

---

## Adding New Features

### Adding a New Consolidation Service

**Step 1: Define Service Structure**

```go
// internal/application/new_service.go
package application

type NewService struct {
    repository        domain.ElementRepository
    embeddingProvider embeddings.Provider
    config            NewServiceConfig
    logger            *slog.Logger
}

type NewServiceConfig struct {
    Enabled   bool
    Parameter string
}

func NewNewService(
    repository domain.ElementRepository,
    embeddingProvider embeddings.Provider,
    config NewServiceConfig,
    logger *slog.Logger,
) *NewService {
    if logger == nil {
        logger = logger.Get()
    }
    return &NewService{
        repository:        repository,
        embeddingProvider: embeddingProvider,
        config:            config,
        logger:            logger,
    }
}
```

**Step 2: Implement Core Functionality**

```go
func (s *NewService) Process(ctx context.Context, input Input) (*Output, error) {
    // Check config
    if !s.config.Enabled {
        return nil, fmt.Errorf("service disabled")
    }

    // Validate input
    if err := s.validateInput(input); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }

    // Execute business logic
    result, err := s.executeLogic(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("execution failed: %w", err)
    }

    return result, nil
}
```

**Step 3: Add Configuration**

```go
// internal/config/config.go

type Config struct {
    // ... existing configs
    NewService NewServiceConfig
}

type NewServiceConfig struct {
    Enabled   bool
    Parameter string
}

// In LoadConfig()
cfg.NewService = NewServiceConfig{
    Enabled:   getEnvBool("NEXS_NEW_SERVICE_ENABLED", true),
    Parameter: getEnvOrDefault("NEXS_NEW_SERVICE_PARAMETER", "default"),
}
```

**Step 4: Add MCP Tool**

```go
// internal/mcp/consolidation_tools.go

func (s *Server) registerNewServiceTool() {
    s.sdk.AddTool(mcp.Tool{
        Name:        "new_service_action",
        Description: "Description of what the tool does",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "param1": map[string]interface{}{
                    "type":        "string",
                    "description": "Parameter description",
                },
            },
            "required": []string{"param1"},
        },
    }, s.handleNewServiceTool)
}

func (s *Server) handleNewServiceTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Parse arguments
    var args struct {
        Param1 string `json:"param1"`
    }
    if err := parseToolArgs(req.Params.Arguments, &args); err != nil {
        return toolError(err), nil
    }

    // Call service
    result, err := s.newService.Process(ctx, args)
    if err != nil {
        return toolError(err), nil
    }

    // Return result
    return toolSuccess(result), nil
}
```

**Step 5: Write Tests**

```go
// internal/application/new_service_test.go

func TestNewService_Process(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        want    *Output
        wantErr bool
    }{
        {
            name: "valid input",
            input: Input{Param1: "value"},
            want: &Output{Result: "expected"},
            wantErr: false,
        },
        {
            name: "invalid input",
            input: Input{Param1: ""},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewNewService(mockRepo, mockProvider, config, nil)
            got, err := service.Process(context.Background(), tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

**Step 6: Update Documentation**

- Add to `docs/api/MCP_TOOLS.md`
- Add to `docs/api/CONSOLIDATION_TOOLS.md`
- Add to `docs/architecture/APPLICATION.md`
- Update this guide

---

## Testing Strategies

### Unit Testing

```go
// Use table-driven tests
func TestService_Method(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        setup   func(*mockRepository)
        want    Output
        wantErr bool
    }{
        {
            name: "success case",
            input: Input{Value: "test"},
            setup: func(repo *mockRepository) {
                repo.elements["id"] = testElement
            },
            want: Output{Result: "expected"},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := &mockRepository{elements: make(map[string]domain.Element)}
            if tt.setup != nil {
                tt.setup(repo)
            }

            service := NewService(repo, mockProvider, config, nil)
            got, err := service.Method(context.Background(), tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Mock Provider

```go
// Use consistent mock provider
mockProvider := embeddings.NewMockProvider("mock", 128)

// Mock returns deterministic embeddings
embedding, err := mockProvider.Embed(ctx, "test content")
// embedding is always same for same content
```

### Integration Testing

```go
func TestConsolidation_Integration(t *testing.T) {
    // Setup real components
    repo := infrastructure.NewFileStorage("test-data")
    provider := embeddings.GetProvider()

    // Initialize services
    detector := NewDuplicateDetectionService(repo, provider, config, nil)
    clusterer := NewClusteringService(repo, provider, config, nil)
    consolidator := NewMemoryConsolidationService(detector, clusterer, nil, nil, config, nil)

    // Execute workflow
    result, err := consolidator.ConsolidateMemories(ctx, req)

    // Verify results
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Greater(t, len(result.StepsExecuted), 0)
}
```

---

## Performance Optimization

### 1. HNSW Index Persistence

```go
// Save index to disk
func (s *HybridSearchService) SaveIndex(path string) error {
    if !s.config.IndexPersistence {
        return nil
    }

    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()

    return s.hnswIndex.Save(file)
}

// Load index from disk
func (s *HybridSearchService) LoadIndex(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    return s.hnswIndex.Load(file)
}
```

### 2. Batch Processing

```go
// Process embeddings in batches
func (s *Service) BatchEmbed(ctx context.Context, contents []string) ([][]float32, error) {
    batchSize := 32
    results := make([][]float32, 0, len(contents))

    for i := 0; i < len(contents); i += batchSize {
        end := min(i+batchSize, len(contents))
        batch := contents[i:end]

        embeddings, err := s.embeddingProvider.EmbedBatch(ctx, batch)
        if err != nil {
            return nil, err
        }

        results = append(results, embeddings...)
    }

    return results, nil
}
```

### 3. Concurrent Processing

```go
// Process elements concurrently
func (s *Service) ProcessConcurrent(ctx context.Context, elements []Element) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(elements))

    for _, elem := range elements {
        wg.Add(1)
        go func(e Element) {
            defer wg.Done()
            if err := s.processOne(ctx, e); err != nil {
                errChan <- err
            }
        }(elem)
    }

    wg.Wait()
    close(errChan)

    // Check for errors
    for err := range errChan {
        if err != nil {
            return err
        }
    }

    return nil
}
```

### 4. Caching

```go
// Cache embeddings
type EmbeddingCache struct {
    cache map[string][]float32
    mu    sync.RWMutex
    ttl   time.Duration
}

func (c *EmbeddingCache) Get(content string) ([]float32, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    embedding, ok := c.cache[content]
    return embedding, ok
}

func (c *EmbeddingCache) Set(content string, embedding []float32) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache[content] = embedding
}
```

---

## Best Practices

### 1. Always Use Context

```go
✅ func (s *Service) Method(ctx context.Context) error
❌ func (s *Service) Method() error
```

### 2. Return Errors, Don't Panic

```go
✅ return fmt.Errorf("operation failed: %w", err)
❌ panic("operation failed")
```

### 3. Use Structured Logging

```go
✅ s.logger.Info("operation completed", "duration_ms", duration, "count", count)
❌ log.Printf("operation completed in %d ms, processed %d items", duration, count)
```

### 4. Validate Input

```go
func (s *Service) Process(input Input) error {
    if err := input.Validate(); err != nil {
        return fmt.Errorf("invalid input: %w", err)
    }
    // ...
}
```

### 5. Use Dependency Injection

```go
✅ func NewService(repo Repository) *Service
❌ func NewService() *Service { repo := infrastructure.NewRepo() }
```

---

## Troubleshooting

### Issue: Slow Consolidation

**Symptoms:**
- Consolidation takes > 5 minutes
- High CPU/memory usage

**Solutions:**
1. Reduce batch size
2. Enable index persistence
3. Process in smaller chunks
4. Use linear mode for small datasets

### Issue: Low Clustering Quality

**Symptoms:**
- Too many outliers
- Clusters don't make sense
- Low silhouette score

**Solutions:**
1. Adjust epsilon parameter (DBSCAN)
2. Adjust number of clusters (K-means)
3. Try different algorithm
4. Check embedding quality

### Issue: Memory Leaks

**Symptoms:**
- Memory usage grows over time
- Out of memory errors

**Solutions:**
1. Clear caches periodically
2. Limit embedding cache size
3. Process in batches
4. Enable garbage collection

---

## Related Documentation

- [Memory Consolidation User Guide](../user-guide/MEMORY_CONSOLIDATION.md)
- [MCP Tools Reference](../api/MCP_TOOLS.md)
- [Consolidation Tools Examples](../api/CONSOLIDATION_TOOLS.md)
- [Application Architecture](../architecture/APPLICATION.md)
- [Testing Guide](./TESTING.md)

---

**Version:** 1.3.0  
**Last Updated:** December 26, 2025  
**Maintainer:** NEXS-MCP Team
