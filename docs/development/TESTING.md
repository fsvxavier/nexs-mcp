# Testing Guide

**Version:** 1.3.0  
**Last Updated:** December 26, 2025  
**Target Audience:** Contributors and Developers

---

## Table of Contents

- [Overview](#overview)
- [Testing Philosophy](#testing-philosophy)
- [Test Organization](#test-organization)
- [Unit Testing](#unit-testing)
- [Integration Testing](#integration-testing)
- [MCP Protocol Testing](#mcp-protocol-testing)
- [Writing Tests](#writing-tests)
- [Running Tests](#running-tests)
- [Coverage Requirements](#coverage-requirements)
- [Mocking Strategies](#mocking-strategies)
- [Test Fixtures](#test-fixtures)
- [Performance Testing](#performance-testing)
- [CI/CD Pipeline](#cicd-pipeline)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

NEXS MCP maintains high code quality through comprehensive testing. Built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) (`github.com/modelcontextprotocol/go-sdk/mcp v1.1.0`), our testing strategy ensures reliability, maintainability, and MCP protocol compliance.

**Current Test Coverage:** 76.4% (application layer), 63.2% (overall)  
**Target Coverage:** 70%+  
**Test Count:** 730+ tests across all packages (123 new tests in Sprint 14)  
**Quality Metrics:** Zero race conditions, Zero linter issues

### Sprint 14 Test Additions (v1.3.0)

**New Test Files (7):**
- `duplicate_detection_test.go` - 15 tests, 442 lines
- `clustering_test.go` - 13 tests, 437 lines
- `knowledge_graph_extractor_test.go` - 20 tests, 518 lines
- `memory_consolidation_test.go` - 20 tests, 583 lines
- `hybrid_search_test.go` - 20 tests, 530 lines
- `memory_retention_test.go` - 15 tests, 378 lines
- `semantic_search_test.go` - 20 tests, 545 lines

**Total New Tests:** 123 tests, 3,433 lines of code  
**Pass Rate:** 100% (295/295 application tests passing)  
**Coverage Improvement:** +13.2% (63.2% → 76.4%)

### Test Structure

```
nexs-mcp/
├── internal/
│   ├── domain/              # Domain tests (95%+ coverage)
│   │   ├── element_test.go
│   │   ├── agent_test.go
│   │   └── ...
│   ├── application/         # Application tests (85%+ coverage)
│   │   ├── ensemble_executor_test.go
│   │   └── ...
│   ├── infrastructure/      # Infrastructure tests (70%+ coverage)
│   │   └── ...
│   └── mcp/                 # MCP server tests
│       └── server_test.go
└── test/
    └── integration/         # Integration tests
        ├── basic_test.go
        └── mcp_test.go
```

---

## Testing Philosophy

### Principles

1. **Test Behavior, Not Implementation** - Focus on what code does, not how
2. **Test Isolation** - Each test should be independent
3. **Clear Test Names** - Test names should describe the scenario
4. **Comprehensive Coverage** - Cover happy paths, edge cases, and errors
5. **Fast Execution** - Unit tests should run in milliseconds
6. **Maintainable Tests** - Tests should be easy to understand and modify

### Test Pyramid

```
        /\
       /  \        E2E Tests (Few)
      /____\       
     /      \      Integration Tests (Some)
    /________\     
   /          \    Unit Tests (Many)
  /__________  \   
```

**Distribution:**
- 70% Unit Tests - Fast, isolated, test individual functions
- 25% Integration Tests - Test component interactions
- 5% E2E Tests - Test complete workflows

### What to Test

**✅ DO Test:**
- Public APIs and exported functions
- Business logic in domain layer
- Error handling and edge cases
- MCP tool implementations
- Data validation and transformations
- Integration between components

**❌ DON'T Test:**
- Private implementation details
- Third-party libraries (trust but verify integration)
- Generated code
- Trivial getters/setters without logic

---

## Test Organization

### File Naming

Test files follow Go conventions:

```
element.go       → element_test.go
agent.go         → agent_test.go
server.go        → server_test.go
```

### Package Structure

**Same Package Testing** (preferred for unit tests):

```go
package domain

// element.go
type Element struct { ... }

// element_test.go
package domain

func TestElement_Validate(t *testing.T) { ... }
```

**External Package Testing** (for integration tests):

```go
// domain_test package
package domain_test

import "github.com/fsvxavier/nexs-mcp/internal/domain"

func TestDomainIntegration(t *testing.T) { ... }
```

### Test Organization Within File

```go
package domain

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// 1. Test fixtures and helpers at top
var validElement = Element{
    ID:   "test-1",
    Type: PersonaType,
    Name: "Test Persona",
}

func createTestElement() *Element { ... }

// 2. Table-driven tests
func TestElement_Validate(t *testing.T) {
    tests := []struct {
        name    string
        element Element
        wantErr bool
        errMsg  string
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}

// 3. Individual test functions
func TestElement_SpecialCase(t *testing.T) { ... }

// 4. Benchmark tests at end
func BenchmarkElement_Validate(b *testing.B) { ... }
```

---

## Unit Testing

### Sprint 14: Memory Consolidation Tests

Sprint 14 introduced 7 comprehensive test files (3,433 lines, 123 tests) for advanced memory management services:

#### 1. Duplicate Detection Tests (`duplicate_detection_test.go`)

**Coverage:** 15 tests, 442 lines, 78.5% coverage

**Key Test Scenarios:**
```go
// Basic duplicate detection
func TestDuplicateDetectionService_DetectDuplicates(t *testing.T)

// Similarity threshold testing
func TestDuplicateDetectionService_SimilarityThreshold(t *testing.T)

// Merging functionality
func TestDuplicateDetectionService_MergeDuplicates(t *testing.T)

// Edge cases
func TestDuplicateDetectionService_EmptyElements(t *testing.T)
func TestDuplicateDetectionService_SingleElement(t *testing.T)
func TestDuplicateDetectionService_NoDuplicates(t *testing.T)
```

**Example Test:**
```go
func TestDuplicateDetectionService_DetectDuplicates(t *testing.T) {
    // Setup mock provider and repository
    mockProvider := embeddings.NewMockProvider("mock", 128)
    mockRepo := &mockRepository{elements: make(map[string]domain.Element)}
    
    // Create test memories with similar content
    memory1 := createTestMemory("mem-1", "Machine learning implementation guide")
    memory2 := createTestMemory("mem-2", "Machine learning implementation tutorial")
    
    // Initialize service
    service := NewDuplicateDetectionService(mockRepo, mockProvider, config, logger)
    
    // Detect duplicates
    groups, err := service.DetectDuplicates(ctx, "memory", 0.90)
    
    // Assertions
    assert.NoError(t, err)
    assert.Len(t, groups, 1)
    assert.Len(t, groups[0].Elements, 2)
    assert.Greater(t, groups[0].Similarity, float32(0.90))
}
```

#### 2. Clustering Tests (`clustering_test.go`)

**Coverage:** 13 tests, 437 lines, 72.1% coverage

**Key Test Scenarios:**
```go
// DBSCAN clustering
func TestClusteringService_DBSCAN(t *testing.T)
func TestClusteringService_DBSCAN_Parameters(t *testing.T)
func TestClusteringService_DBSCAN_Outliers(t *testing.T)

// K-means clustering
func TestClusteringService_KMeans(t *testing.T)
func TestClusteringService_KMeans_Convergence(t *testing.T)

// Quality metrics
func TestClusteringService_SilhouetteScore(t *testing.T)
func TestClusteringService_ClusterQuality(t *testing.T)
```

**Algorithm Testing:**
```go
func TestClusteringService_DBSCAN(t *testing.T) {
    tests := []struct {
        name           string
        minClusterSize int
        epsilon        float32
        expectedClusters int
        expectedOutliers int
    }{
        {
            name:           "standard parameters",
            minClusterSize: 3,
            epsilon:        0.15,
            expectedClusters: 3,
            expectedOutliers: 2,
        },
        {
            name:           "tight clustering",
            minClusterSize: 5,
            epsilon:        0.10,
            expectedClusters: 2,
            expectedOutliers: 5,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewClusteringService(repo, provider, config, logger)
            clusters, outliers, err := service.ClusterDBSCAN(
                ctx, "memory", tt.minClusterSize, tt.epsilon)
            
            assert.NoError(t, err)
            assert.Len(t, clusters, tt.expectedClusters)
            assert.Len(t, outliers, tt.expectedOutliers)
        })
    }
}
```

#### 3. Knowledge Graph Tests (`knowledge_graph_extractor_test.go`)

**Coverage:** 20 tests, 518 lines, 81.3% coverage

**Entity Extraction Tests:**
```go
// Person name extraction
func TestKnowledgeGraphExtractor_ExtractPeople(t *testing.T)

// Organization extraction
func TestKnowledgeGraphExtractor_ExtractOrganizations(t *testing.T)

// URL and email extraction
func TestKnowledgeGraphExtractor_ExtractURLs(t *testing.T)
func TestKnowledgeGraphExtractor_ExtractEmails(t *testing.T)

// Keyword extraction with TF-IDF
func TestKnowledgeGraphExtractor_ExtractKeywords(t *testing.T)

// Relationship extraction
func TestKnowledgeGraphExtractor_ExtractRelationships(t *testing.T)
```

**NLP Testing Example:**
```go
func TestKnowledgeGraphExtractor_ExtractPeople(t *testing.T) {
    content := `
        John Smith and Jane Doe worked together on Project Alpha.
        They collaborated with Dr. Robert Johnson from MIT.
    `
    
    extractor := NewKnowledgeGraphExtractorService(repo, config, logger)
    graph, err := extractor.ExtractFromContent(ctx, content)
    
    assert.NoError(t, err)
    assert.Len(t, graph.Entities.People, 3)
    
    people := graph.Entities.People
    assert.Contains(t, people, Person{Name: "John Smith"})
    assert.Contains(t, people, Person{Name: "Jane Doe"})
    assert.Contains(t, people, Person{Name: "Dr. Robert Johnson"})
}
```

#### 4. Memory Consolidation Tests (`memory_consolidation_test.go`)

**Coverage:** 20 tests, 583 lines, 85.2% coverage

**Workflow Orchestration Tests:**
```go
// Full consolidation workflow
func TestMemoryConsolidation_CompleteWorkflow(t *testing.T)

// Individual step testing
func TestMemoryConsolidation_DuplicateDetectionStep(t *testing.T)
func TestMemoryConsolidation_ClusteringStep(t *testing.T)
func TestMemoryConsolidation_KnowledgeExtractionStep(t *testing.T)
func TestMemoryConsolidation_QualityScoringStep(t *testing.T)

// Error handling
func TestMemoryConsolidation_ErrorRecovery(t *testing.T)
func TestMemoryConsolidation_PartialFailure(t *testing.T)

// Dry run mode
func TestMemoryConsolidation_DryRun(t *testing.T)
```

#### 5. Hybrid Search Tests (`hybrid_search_test.go`)

**Coverage:** 20 tests, 530 lines, 79.8% coverage

**Mode Switching Tests:**
```go
// Auto mode selection
func TestHybridSearch_AutoMode(t *testing.T)
func TestHybridSearch_ModeSwitch(t *testing.T)

// HNSW mode
func TestHybridSearch_HNSWMode(t *testing.T)
func TestHybridSearch_HNSWPerformance(t *testing.T)

// Linear mode
func TestHybridSearch_LinearMode(t *testing.T)
func TestHybridSearch_LinearAccuracy(t *testing.T)

// Index persistence
func TestHybridSearch_IndexPersistence(t *testing.T)
```

#### 6. Memory Retention Tests (`memory_retention_test.go`)

**Coverage:** 15 tests, 378 lines, 76.9% coverage

**Quality Scoring Tests:**
```go
// Quality score calculation
func TestMemoryRetention_ScoreMemories(t *testing.T)
func TestMemoryRetention_QualityFactors(t *testing.T)

// Retention policies
func TestMemoryRetention_ApplyPolicy(t *testing.T)
func TestMemoryRetention_RetentionPeriods(t *testing.T)

// Cleanup
func TestMemoryRetention_AutoCleanup(t *testing.T)
func TestMemoryRetention_DryRun(t *testing.T)
```

#### 7. Semantic Search Tests (`semantic_search_test.go`)

**Coverage:** 20 tests, 545 lines, 82.4% coverage

**Indexing Tests:**
```go
// Index management
func TestSemanticSearch_IndexElements(t *testing.T)
func TestSemanticSearch_ReindexElement(t *testing.T)

// Search functionality
func TestSemanticSearch_Search(t *testing.T)
func TestSemanticSearch_MultiTypeSearch(t *testing.T)

// Filtering
func TestSemanticSearch_TagFiltering(t *testing.T)
func TestSemanticSearch_DateRangeFiltering(t *testing.T)
```

### Test Quality Metrics (Sprint 14)

**Code Quality:**
- ✅ 0 race conditions (verified with `go test -race`)
- ✅ 0 lint issues (golangci-lint)
- ✅ 100% pass rate (295/295 tests)
- ✅ Table-driven tests for comprehensive coverage
- ✅ Mock providers for deterministic testing

**Performance:**
- Average test execution: 1.2s for all 295 application tests
- No flaky tests detected
- Consistent results across multiple runs

### Basic Unit Test

```go
package domain

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewPersona(t *testing.T) {
    // Arrange
    name := "assistant"
    profile := PersonaProfile{
        Traits: []string{"helpful", "concise"},
        Tone:   "professional",
    }

    // Act
    persona, err := NewPersona(name, profile)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, persona)
    assert.Equal(t, name, persona.Name)
    assert.NotEmpty(t, persona.ID)
    assert.Equal(t, profile.Traits, persona.Profile.Traits)
}
```

### Table-Driven Tests

```go
func TestElement_Validate(t *testing.T) {
    tests := []struct {
        name    string
        element Element
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid element",
            element: Element{
                ID:   "test-1",
                Type: PersonaType,
                Name: "Test",
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            element: Element{
                Type: PersonaType,
                Name: "Test",
            },
            wantErr: true,
            errMsg:  "ID is required",
        },
        {
            name: "invalid type",
            element: Element{
                ID:   "test-1",
                Type: "invalid",
                Name: "Test",
            },
            wantErr: true,
            errMsg:  "invalid element type",
        },
        {
            name: "empty name",
            element: Element{
                ID:   "test-1",
                Type: PersonaType,
                Name: "",
            },
            wantErr: true,
            errMsg:  "name is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.element.Validate()
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Testing Error Cases

```go
func TestCreatePersona_Errors(t *testing.T) {
    t.Run("empty name", func(t *testing.T) {
        _, err := NewPersona("", validProfile)
        
        assert.Error(t, err)
        assert.IsType(t, &ValidationError{}, err)
        
        validErr := err.(*ValidationError)
        assert.Equal(t, "name", validErr.Field)
    })

    t.Run("invalid profile", func(t *testing.T) {
        invalidProfile := PersonaProfile{}
        _, err := NewPersona("test", invalidProfile)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "profile")
    })

    t.Run("repository failure", func(t *testing.T) {
        mockRepo := new(MockRepository)
        mockRepo.On("Save", mock.Anything).Return(errors.New("db error"))
        
        svc := NewService(mockRepo)
        _, err := svc.CreatePersona("test", validProfile)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "db error")
    })
}
```

### Testing with Context

```go
func TestService_CreateWithTimeout(t *testing.T) {
    t.Run("successful creation", func(t *testing.T) {
        ctx := context.Background()
        svc := NewService(mockRepo)
        
        persona, err := svc.Create(ctx, "test", validProfile)
        
        assert.NoError(t, err)
        assert.NotNil(t, persona)
    })

    t.Run("context timeout", func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
        defer cancel()
        
        time.Sleep(10 * time.Millisecond) // Force timeout
        
        svc := NewService(slowRepo)
        _, err := svc.Create(ctx, "test", validProfile)
        
        assert.Error(t, err)
        assert.ErrorIs(t, err, context.DeadlineExceeded)
    })

    t.Run("context cancellation", func(t *testing.T) {
        ctx, cancel := context.WithCancel(context.Background())
        
        go func() {
            time.Sleep(5 * time.Millisecond)
            cancel()
        }()
        
        svc := NewService(slowRepo)
        _, err := svc.Create(ctx, "test", validProfile)
        
        assert.Error(t, err)
        assert.ErrorIs(t, err, context.Canceled)
    })
}
```

---

## Integration Testing

### Basic Integration Test

```go
// test/integration/element_integration_test.go
package integration

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/fsvxavier/nexs-mcp/internal/domain"
    "github.com/fsvxavier/nexs-mcp/internal/infrastructure/storage"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestElementStorageIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup
    tempDir := t.TempDir()
    repo := storage.NewFileRepository(tempDir)

    ctx := context.Background()

    // Test create
    element := domain.Element{
        ID:   "test-1",
        Type: domain.PersonaType,
        Name: "Test Persona",
    }

    err := repo.Save(ctx, element)
    require.NoError(t, err)

    // Test retrieve
    retrieved, err := repo.Get(ctx, "test-1")
    require.NoError(t, err)
    assert.Equal(t, element.ID, retrieved.ID)
    assert.Equal(t, element.Name, retrieved.Name)

    // Test list
    elements, err := repo.List(ctx, domain.PersonaType)
    require.NoError(t, err)
    assert.Len(t, elements, 1)

    // Test update
    element.Name = "Updated Name"
    err = repo.Save(ctx, element)
    require.NoError(t, err)

    retrieved, err = repo.Get(ctx, "test-1")
    require.NoError(t, err)
    assert.Equal(t, "Updated Name", retrieved.Name)

    // Test delete
    err = repo.Delete(ctx, "test-1")
    require.NoError(t, err)

    _, err = repo.Get(ctx, "test-1")
    assert.Error(t, err)
}
```

### Multi-Component Integration

```go
func TestEnsembleExecutionIntegration(t *testing.T) {
    // Setup complete system
    ctx := context.Background()
    tempDir := t.TempDir()
    
    repo := storage.NewFileRepository(tempDir)
    executor := application.NewEnsembleExecutor(repo)
    monitor := application.NewEnsembleMonitor()

    // Create test data
    agent1 := createTestAgent("agent-1", "analyzer")
    agent2 := createTestAgent("agent-2", "executor")
    
    require.NoError(t, repo.Save(ctx, agent1))
    require.NoError(t, repo.Save(ctx, agent2))

    ensemble := domain.NewEnsemble("test-ensemble")
    ensemble.AddMember(agent1.ID)
    ensemble.AddMember(agent2.ID)
    
    require.NoError(t, repo.Save(ctx, ensemble))

    // Execute ensemble
    input := "Test input for processing"
    results, err := executor.Execute(ctx, ensemble.ID, input)
    
    require.NoError(t, err)
    assert.Len(t, results, 2)
    
    // Verify monitoring
    stats := monitor.GetStats(ensemble.ID)
    assert.Equal(t, 1, stats.ExecutionCount)
    assert.True(t, stats.LastExecutionTime > 0)
}
```

### Database Integration Tests

```go
func TestDatabaseIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database integration test")
    }

    // Setup test database
    db, cleanup := setupTestDB(t)
    defer cleanup()

    repo := NewDatabaseRepository(db)
    ctx := context.Background()

    t.Run("transaction rollback on error", func(t *testing.T) {
        // Begin transaction
        tx, err := db.Begin()
        require.NoError(t, err)

        // Save element
        element := createTestElement()
        err = repo.SaveWithTx(ctx, tx, element)
        require.NoError(t, err)

        // Rollback
        tx.Rollback()

        // Verify element not persisted
        _, err = repo.Get(ctx, element.ID)
        assert.Error(t, err)
    })

    t.Run("transaction commit", func(t *testing.T) {
        tx, err := db.Begin()
        require.NoError(t, err)

        element := createTestElement()
        err = repo.SaveWithTx(ctx, tx, element)
        require.NoError(t, err)

        // Commit
        err = tx.Commit()
        require.NoError(t, err)

        // Verify element persisted
        retrieved, err := repo.Get(ctx, element.ID)
        require.NoError(t, err)
        assert.Equal(t, element.ID, retrieved.ID)
    })
}
```

---

## MCP Protocol Testing

### Testing MCP Tool Registration

```go
package mcp

import (
    "testing"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestServer_RegisterTools(t *testing.T) {
    // Create server with official MCP SDK
    serverInfo := mcp.ServerInfo{
        Name:    "nexs-mcp-test",
        Version: "0.1.0",
    }
    
    mcpServer := mcp.NewServer(serverInfo)
    server := NewServer(mcpServer, mockRepo)

    // Register all tools
    err := server.RegisterTools()
    require.NoError(t, err)

    // Verify expected tools are registered
    expectedTools := []string{
        "create_persona",
        "list_personas",
        "get_persona",
        "update_persona",
        "delete_persona",
        // ... more tools
    }

    for _, toolName := range expectedTools {
        // Verify tool can be called
        _, err := server.CallTool(toolName, nil)
        assert.NoError(t, err, "Tool %s should be registered", toolName)
    }
}
```

### Testing MCP Tool Handlers

```go
func TestServer_HandleCreatePersona(t *testing.T) {
    server := setupTestServer(t)
    ctx := context.Background()

    tests := []struct {
        name    string
        args    map[string]interface{}
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid creation",
            args: map[string]interface{}{
                "name": "assistant",
                "profile": map[string]interface{}{
                    "traits": []string{"helpful"},
                    "tone":   "professional",
                },
            },
            wantErr: false,
        },
        {
            name: "missing name",
            args: map[string]interface{}{
                "profile": map[string]interface{}{},
            },
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name: "invalid profile",
            args: map[string]interface{}{
                "name":    "test",
                "profile": "invalid",
            },
            wantErr: true,
            errMsg:  "profile must be an object",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            response, err := server.handleCreatePersona(ctx, tt.args)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
                assert.NotNil(t, response)
                
                // Verify response structure
                content := response.Content
                require.NotEmpty(t, content)
                
                firstContent := content[0].(map[string]interface{})
                assert.Equal(t, "text", firstContent["type"])
                assert.NotEmpty(t, firstContent["text"])
            }
        })
    }
}
```

### Testing MCP Resources

```go
func TestServer_Resources(t *testing.T) {
    server := setupTestServer(t)
    ctx := context.Background()

    // Seed test data
    persona := createTestPersona("test-persona")
    require.NoError(t, server.repo.Save(ctx, persona))

    t.Run("list resources", func(t *testing.T) {
        resources, err := server.ListResources(ctx)
        
        require.NoError(t, err)
        assert.NotEmpty(t, resources)
        
        // Find persona resource
        found := false
        for _, res := range resources {
            if res.URI == "persona://test-persona" {
                found = true
                assert.Equal(t, "Test Persona", res.Name)
                break
            }
        }
        assert.True(t, found, "Persona resource should be listed")
    })

    t.Run("read resource", func(t *testing.T) {
        uri := "persona://test-persona"
        content, err := server.ReadResource(ctx, uri)
        
        require.NoError(t, err)
        assert.NotEmpty(t, content)
        assert.Contains(t, content, "test-persona")
    })

    t.Run("resource not found", func(t *testing.T) {
        uri := "persona://nonexistent"
        _, err := server.ReadResource(ctx, uri)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "not found")
    })
}
```

### Testing MCP Protocol Flow

```go
func TestMCPProtocolFlow(t *testing.T) {
    // Simulate complete MCP interaction
    server := setupTestServer(t)
    ctx := context.Background()

    // 1. Initialize
    err := server.Initialize(ctx)
    require.NoError(t, err)

    // 2. List available tools
    tools, err := server.ListTools(ctx)
    require.NoError(t, err)
    assert.NotEmpty(t, tools)

    // 3. Call a tool
    createArgs := map[string]interface{}{
        "name": "test-persona",
        "profile": map[string]interface{}{
            "traits": []string{"helpful"},
        },
    }
    
    createResp, err := server.CallTool(ctx, "create_persona", createArgs)
    require.NoError(t, err)
    assert.NotNil(t, createResp)

    // 4. List resources
    resources, err := server.ListResources(ctx)
    require.NoError(t, err)
    
    // Verify created persona is in resources
    found := false
    for _, res := range resources {
        if res.URI == "persona://test-persona" {
            found = true
            break
        }
    }
    assert.True(t, found)

    // 5. Read resource
    content, err := server.ReadResource(ctx, "persona://test-persona")
    require.NoError(t, err)
    assert.Contains(t, content, "test-persona")

    // 6. Update via tool
    updateArgs := map[string]interface{}{
        "name": "test-persona",
        "profile": map[string]interface{}{
            "traits": []string{"helpful", "concise"},
        },
    }
    
    updateResp, err := server.CallTool(ctx, "update_persona", updateArgs)
    require.NoError(t, err)
    assert.NotNil(t, updateResp)

    // 7. Delete via tool
    deleteArgs := map[string]interface{}{
        "name": "test-persona",
    }
    
    deleteResp, err := server.CallTool(ctx, "delete_persona", deleteArgs)
    require.NoError(t, err)
    assert.NotNil(t, deleteResp)

    // 8. Verify deletion
    resources, err = server.ListResources(ctx)
    require.NoError(t, err)
    
    for _, res := range resources {
        assert.NotEqual(t, "persona://test-persona", res.URI)
    }
}
```

---

## Writing Tests

### Test Structure (AAA Pattern)

```go
func TestFeature(t *testing.T) {
    // Arrange - Setup test data and dependencies
    repo := setupTestRepository(t)
    svc := NewService(repo)
    input := "test input"

    // Act - Execute the functionality being tested
    result, err := svc.Process(input)

    // Assert - Verify the results
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, expectedValue, result.Value)
}
```

### Test Naming Conventions

```go
// Pattern: Test<Function>_<Scenario>_<ExpectedResult>

func TestCreatePersona_ValidInput_Success(t *testing.T) { }
func TestCreatePersona_EmptyName_ReturnsError(t *testing.T) { }
func TestCreatePersona_DuplicateName_ReturnsError(t *testing.T) { }

// Pattern: Test<Type>_<Method>_<Scenario>
func TestElement_Validate_MissingID(t *testing.T) { }
func TestElement_Validate_ValidData(t *testing.T) { }
```

### Using Assertions

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAssertions(t *testing.T) {
    // Use assert for non-critical checks (test continues)
    assert.Equal(t, expected, actual)
    assert.NoError(t, err)
    assert.True(t, condition)

    // Use require for critical checks (test stops)
    require.NoError(t, err)  // Stop if error
    require.NotNil(t, obj)   // Stop if nil

    // Specific assertions
    assert.Contains(t, "hello world", "world")
    assert.Greater(t, value, 10)
    assert.Len(t, slice, 5)
    assert.IsType(t, &MyType{}, value)
    assert.JSONEq(t, expectedJSON, actualJSON)
}
```

### Subtests

```go
func TestMultipleScenarios(t *testing.T) {
    setup := func() *Service {
        return NewService(mockRepo)
    }

    t.Run("scenario 1", func(t *testing.T) {
        svc := setup()
        // Test scenario 1
    })

    t.Run("scenario 2", func(t *testing.T) {
        svc := setup()
        // Test scenario 2
    })

    t.Run("error cases", func(t *testing.T) {
        t.Run("invalid input", func(t *testing.T) {
            svc := setup()
            // Test invalid input
        })

        t.Run("timeout", func(t *testing.T) {
            svc := setup()
            // Test timeout
        })
    })
}
```

---

## Running Tests

### Basic Test Execution

```bash
# Run all tests
go test ./...

# Run tests in specific package
go test ./internal/domain

# Run specific test
go test -run TestCreatePersona ./internal/domain

# Run tests with verbose output
go test -v ./...

# Run tests matching pattern
go test -run "TestElement_.*" ./internal/domain
```

### Using Make Targets

```bash
# Run all tests
make test

# Run with race detector
make test-race

# Generate coverage report
make test-coverage

# Run only short tests
make test-short
```

### Race Detection

```bash
# Run with race detector (detects data races)
go test -race ./...

# Or use Makefile
make test-race
```

### Running Specific Tests

```bash
# Run single test function
go test -run TestCreatePersona

# Run tests in subtests
go test -run TestElement/valid_element

# Run multiple patterns
go test -run "Test(Create|Update|Delete)"
```

### Short Mode

```bash
# Skip slow integration tests
go test -short ./...

# In test code
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    // Integration test logic
}
```

### Parallel Execution

```bash
# Run tests in parallel
go test -parallel 4 ./...

# In test code
func TestParallel(t *testing.T) {
    t.Parallel()  // Mark test as parallelizable
    // Test logic
}
```

### Timeout Configuration

```bash
# Set global timeout
go test -timeout 30s ./...

# Set timeout per test
go test -timeout 5m ./test/integration/...
```

---

## Coverage Requirements

### Coverage Targets

| Layer               | Target Coverage | Current (v1.3.0) | Status |
|---------------------|-----------------|------------------|--------|
| Domain              | 95%+            | 95.7%           | ✅     |
| Application         | 85%+            | 76.4%           | ⚠️     |
| Infrastructure      | 70%+            | 72.1%           | ✅     |
| MCP Server          | 80%+            | 82.3%           | ✅     |
| HNSW Indexing       | 85%+            | 91.7%           | ✅     |
| TF-IDF              | 85%+            | 96.7%           | ✅     |
| Embeddings          | 80%+            | 78.2%           | ⚠️     |
| **Overall Project** | **80%+**        | **76.4%**       | ⚠️     |

**Note:** Application layer coverage increased from 63.2% to 76.4% (+13.2%) in Sprint 14 with the addition of 123 new tests for memory consolidation services.

### Coverage by Service (Sprint 14)

| Service                    | Tests | Coverage | LOC   |
|----------------------------|-------|----------|-------|
| DuplicateDetection         | 15    | 78.5%    | 442   |
| Clustering                 | 13    | 72.1%    | 437   |
| KnowledgeGraphExtractor    | 20    | 81.3%    | 518   |
| MemoryConsolidation        | 20    | 85.2%    | 583   |
| HybridSearch               | 20    | 79.8%    | 530   |
| MemoryRetention            | 15    | 76.9%    | 378   |
| SemanticSearch             | 20    | 82.4%    | 545   |
| **Total Sprint 14**        | **123** | **79.5%** | **3,433** |

### Generating Coverage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Coverage Analysis

```bash
# Show coverage by function
go tool cover -func=coverage.out

# Sample output:
# github.com/fsvxavier/nexs-mcp/internal/domain/element.go:15:    NewElement      100.0%
# github.com/fsvxavier/nexs-mcp/internal/domain/element.go:25:    Validate        95.5%
# github.com/fsvxavier/nexs-mcp/internal/domain/agent.go:10:      NewAgent        100.0%
# total:                                                          (statements)    85.2%

# View coverage per package
go test -cover ./...

# Output with coverage percentage per package
```

### Coverage in CI/CD

```yaml
# .github/workflows/test.yml
- name: Run tests with coverage
  run: |
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

- name: Check coverage threshold
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    threshold=80
    if (( $(echo "$coverage < $threshold" | bc -l) )); then
      echo "Coverage $coverage% is below threshold $threshold%"
      exit 1
    fi
```

---

## Mocking Strategies

### Interface Mocking with testify/mock

```go
package domain

import (
    "github.com/stretchr/testify/mock"
)

// Define mock
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, elem Element) error {
    args := m.Called(ctx, elem)
    return args.Error(0)
}

func (m *MockRepository) Get(ctx context.Context, id string) (*Element, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Element), args.Error(1)
}

// Use in tests
func TestServiceWithMock(t *testing.T) {
    mockRepo := new(MockRepository)
    
    // Setup expectations
    mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
    mockRepo.On("Get", mock.Anything, "test-1").Return(&testElement, nil)
    
    // Create service with mock
    svc := NewService(mockRepo)
    
    // Test service
    err := svc.CreateElement(ctx, testElement)
    assert.NoError(t, err)
    
    // Verify expectations were met
    mockRepo.AssertExpectations(t)
}
```

### Advanced Mock Expectations

```go
func TestAdvancedMocking(t *testing.T) {
    mockRepo := new(MockRepository)

    // Match specific arguments
    mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(e Element) bool {
        return e.Type == PersonaType
    })).Return(nil)

    // Return different values based on input
    mockRepo.On("Get", mock.Anything, "found").Return(&testElement, nil)
    mockRepo.On("Get", mock.Anything, "notfound").Return(nil, ErrNotFound)

    // Track call count
    mockRepo.On("List", mock.Anything).Return([]Element{}, nil).Times(3)

    // Setup once expectation
    mockRepo.On("Initialize").Return(nil).Once()

    // Run with mock
    svc := NewService(mockRepo)
    // ... test logic ...

    // Verify specific calls
    mockRepo.AssertCalled(t, "Get", mock.Anything, "found")
    mockRepo.AssertNumberOfCalls(t, "List", 3)
}
```

### Test Doubles (Stubs, Fakes)

```go
// Stub - Returns predetermined values
type StubRepository struct {
    ReturnElement *Element
    ReturnError   error
}

func (s *StubRepository) Get(ctx context.Context, id string) (*Element, error) {
    return s.ReturnElement, s.ReturnError
}

// Fake - Working implementation with shortcuts
type FakeRepository struct {
    elements map[string]Element
}

func NewFakeRepository() *FakeRepository {
    return &FakeRepository{
        elements: make(map[string]Element),
    }
}

func (f *FakeRepository) Save(ctx context.Context, elem Element) error {
    f.elements[elem.ID] = elem
    return nil
}

func (f *FakeRepository) Get(ctx context.Context, id string) (*Element, error) {
    elem, ok := f.elements[id]
    if !ok {
        return nil, ErrNotFound
    }
    return &elem, nil
}

// Use in tests
func TestWithFake(t *testing.T) {
    repo := NewFakeRepository()
    svc := NewService(repo)
    
    // Fake persists data across calls
    elem := createTestElement()
    err := svc.Create(ctx, elem)
    require.NoError(t, err)
    
    retrieved, err := svc.Get(ctx, elem.ID)
    require.NoError(t, err)
    assert.Equal(t, elem.ID, retrieved.ID)
}
```

---

## Test Fixtures

### Setup and Teardown

```go
func TestMain(m *testing.M) {
    // Global setup
    setup()
    
    // Run tests
    code := m.Run()
    
    // Global teardown
    teardown()
    
    os.Exit(code)
}

func setup() {
    // Initialize test resources
    initTestDatabase()
    createTestDirectories()
}

func teardown() {
    // Cleanup test resources
    cleanupTestDatabase()
    removeTestDirectories()
}
```

### Per-Test Setup

```go
func setupTest(t *testing.T) (*Service, func()) {
    // Setup
    tempDir := t.TempDir()  // Automatically cleaned up
    repo := NewFileRepository(tempDir)
    svc := NewService(repo)
    
    // Return cleanup function
    cleanup := func() {
        // Additional cleanup if needed
    }
    
    return svc, cleanup
}

func TestWithSetup(t *testing.T) {
    svc, cleanup := setupTest(t)
    defer cleanup()
    
    // Test logic using svc
}
```

### Test Data Builders

```go
// Builder pattern for test data
type PersonaBuilder struct {
    persona Persona
}

func NewPersonaBuilder() *PersonaBuilder {
    return &PersonaBuilder{
        persona: Persona{
            ID:   uuid.NewString(),
            Name: "default-persona",
            Profile: PersonaProfile{
                Traits: []string{"helpful"},
                Tone:   "professional",
            },
        },
    }
}

func (b *PersonaBuilder) WithID(id string) *PersonaBuilder {
    b.persona.ID = id
    return b
}

func (b *PersonaBuilder) WithName(name string) *PersonaBuilder {
    b.persona.Name = name
    return b
}

func (b *PersonaBuilder) WithTraits(traits ...string) *PersonaBuilder {
    b.persona.Profile.Traits = traits
    return b
}

func (b *PersonaBuilder) Build() Persona {
    return b.persona
}

// Use in tests
func TestWithBuilder(t *testing.T) {
    persona := NewPersonaBuilder().
        WithName("test-persona").
        WithTraits("helpful", "concise").
        Build()
    
    err := svc.Create(ctx, persona)
    assert.NoError(t, err)
}
```

### Fixture Files

```go
// Load test data from files
func loadTestFixture(t *testing.T, filename string) []byte {
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    require.NoError(t, err)
    return data
}

func TestWithFixture(t *testing.T) {
    data := loadTestFixture(t, "persona.yaml")
    
    var persona Persona
    err := yaml.Unmarshal(data, &persona)
    require.NoError(t, err)
    
    // Test with loaded data
}

// testdata/ directory structure
// testdata/
// ├── persona.yaml
// ├── agent.yaml
// └── ensemble.yaml
```

---

## Performance Testing

### Benchmark Tests

```go
func BenchmarkElement_Validate(b *testing.B) {
    element := createTestElement()
    
    b.ResetTimer()  // Reset timer after setup
    
    for i := 0; i < b.N; i++ {
        _ = element.Validate()
    }
}

func BenchmarkRepository_Save(b *testing.B) {
    repo := setupBenchmarkRepo(b)
    element := createTestElement()
    ctx := context.Background()
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _ = repo.Save(ctx, element)
    }
}

// Run benchmarks
// go test -bench=. ./internal/domain
// go test -bench=BenchmarkElement ./...
```

### Benchmark Comparison

```bash
# Save baseline
go test -bench=. -benchmem ./... > old.txt

# Make changes

# Compare
go test -bench=. -benchmem ./... > new.txt
benchcmp old.txt new.txt
```

### Performance Tests

```go
func TestPerformance_LargeDataset(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }

    repo := setupTestRepository(t)
    ctx := context.Background()

    // Create large dataset
    const numElements = 10000
    for i := 0; i < numElements; i++ {
        elem := createTestElementWithID(fmt.Sprintf("elem-%d", i))
        require.NoError(t, repo.Save(ctx, elem))
    }

    // Measure list performance
    start := time.Now()
    elements, err := repo.List(ctx, PersonaType)
    duration := time.Since(start)

    require.NoError(t, err)
    assert.Len(t, elements, numElements)
    
    // Assert performance requirement
    maxDuration := 100 * time.Millisecond
    assert.Less(t, duration, maxDuration,
        "List operation took %v, expected less than %v", duration, maxDuration)
}
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        go-version: [1.21, 1.25]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Download dependencies
        run: go mod download

      - name: Verify dependencies
        run: go mod verify

      - name: Run tests
        run: make test-race

      - name: Generate coverage
        run: make test-coverage

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          flags: unittests
          name: codecov-umbrella

      - name: Check coverage threshold
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $coverage%"
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage below 80%"
            exit 1
          fi
```

### Pre-commit Hooks

```bash
# .git/hooks/pre-commit
#!/bin/bash

echo "Running pre-commit tests..."

# Run tests
if ! make test; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

# Check coverage
coverage=$(go test -cover ./... 2>&1 | grep -o 'coverage: [0-9.]*%' | awk '{sum+=$2; count++} END {print sum/count}')
threshold=80

if (( $(echo "$coverage < $threshold" | bc -l) )); then
    echo "Coverage $coverage% is below $threshold%. Commit aborted."
    exit 1
fi

echo "All checks passed!"
```

---

## Best Practices

### 1. Write Tests First (TDD)

```go
// 1. Write the test
func TestCreatePersona_ValidInput_Success(t *testing.T) {
    svc := NewService(mockRepo)
    persona, err := svc.CreatePersona("test", validProfile)
    
    assert.NoError(t, err)
    assert.NotNil(t, persona)
}

// 2. Run test (fails)
// 3. Implement minimal code to pass
// 4. Refactor
// 5. Repeat
```

### 2. Keep Tests Independent

```go
// ❌ Bad - Tests depend on execution order
func TestA(t *testing.T) {
    globalVar = "value"
}

func TestB(t *testing.T) {
    assert.Equal(t, "value", globalVar)  // Depends on TestA
}

// ✅ Good - Each test is independent
func TestA(t *testing.T) {
    localVar := "value"
    // Test with localVar
}

func TestB(t *testing.T) {
    localVar := "value"
    // Test with localVar
}
```

### 3. Test One Thing Per Test

```go
// ❌ Bad - Testing multiple concerns
func TestPersona(t *testing.T) {
    persona := createPersona()
    assert.NotNil(t, persona)
    assert.Equal(t, "name", persona.Name)
    
    err := persona.Validate()
    assert.NoError(t, err)
    
    updated := persona.Update(newData)
    assert.True(t, updated)
}

// ✅ Good - One concern per test
func TestCreatePersona_ReturnsNonNil(t *testing.T) {
    persona := createPersona()
    assert.NotNil(t, persona)
}

func TestPersona_Validate_ValidData(t *testing.T) {
    persona := createValidPersona()
    err := persona.Validate()
    assert.NoError(t, err)
}
```

### 4. Use Descriptive Names

```go
// ❌ Bad
func TestPersona1(t *testing.T) { }
func TestError(t *testing.T) { }

// ✅ Good
func TestCreatePersona_EmptyName_ReturnsValidationError(t *testing.T) { }
func TestPersona_Validate_MissingRequiredFields_ReturnsError(t *testing.T) { }
```

### 5. Don't Test Implementation Details

```go
// ❌ Bad - Testing internal implementation
func TestPersona_InternalCacheUpdated(t *testing.T) {
    persona := createPersona()
    assert.Equal(t, expectedValue, persona.internalCache)
}

// ✅ Good - Testing behavior
func TestPersona_Get_ReturnsCachedValue(t *testing.T) {
    persona := createPersona()
    value := persona.Get("key")
    assert.Equal(t, expectedValue, value)
}
```

---

## Troubleshooting

### Tests Fail Intermittently

**Cause:** Race conditions, timing issues, or test dependencies

**Solution:**
```bash
# Run with race detector
go test -race ./...

# Run multiple times
go test -count=100 ./...

# Check for parallel test issues
go test -parallel 1 ./...
```

### Coverage Not Increasing

**Cause:** Untested edge cases or error paths

**Solution:**
```bash
# View uncovered lines
go tool cover -html=coverage.out

# Focus on red (uncovered) sections
```

### Tests Run Slowly

**Solution:**
```bash
# Use -short flag for quick tests
go test -short ./...

# Run specific package
go test ./internal/domain

# Use parallel execution
go test -parallel 8 ./...
```

### Mock Expectations Not Met

**Solution:**
```go
// Add debug output
mockRepo.On("Save", mock.Anything, mock.Anything).
    Return(nil).
    Run(func(args mock.Arguments) {
        fmt.Printf("Save called with: %+v\n", args)
    })

// Verify expectations
mockRepo.AssertExpectations(t)
mockRepo.AssertNumberOfCalls(t, "Save", 1)
```

---

## Additional Resources

### Documentation
- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)

### Articles
- [Table Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Go Testing Best Practices](https://go.dev/wiki/TestComments)

---

**Remember:** Good tests are your safety net. They give you confidence to refactor and evolve the codebase. Invest time in writing quality tests!
