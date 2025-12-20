# Collection Registry

The Collection Registry is a high-performance, multi-source aggregation system for discovering, caching, and indexing NEXS MCP collections.

## Overview

The registry provides:
- **Multi-source support**: GitHub, local filesystem, HTTP registries
- **Intelligent caching**: <10ms cached lookups (actual: 343ns)
- **Rich metadata indexing**: Fast search by author, category, tags, keywords
- **Dependency management**: Cycle detection, topological sorting
- **Thread-safe operations**: Concurrent access support

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Registry                             │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐ │
│  │   Cache     │  │ Metadata     │  │  Dependency   │ │
│  │   (15min)   │  │   Index      │  │    Graph      │ │
│  │  <10ms hit  │  │ (4 indices)  │  │ (cycle detect)│ │
│  └─────────────┘  └──────────────┘  └───────────────┘ │
└─────────────────────────────────────────────────────────┘
         │                    │                   │
    ┌────┴────┐         ┌────┴────┐        ┌────┴────┐
    │ GitHub  │         │  Local  │        │  HTTP   │
    │ Source  │         │  Source │        │ Source  │
    └─────────┘         └─────────┘        └─────────┘
```

## Components

### 1. Registry Cache

High-performance in-memory cache with TTL support.

**Features:**
- Configurable TTL (default: 15 minutes)
- Automatic expiration
- Access tracking
- Cache statistics

**Performance:**
- Cache miss: ~5µs
- Cache hit: **343ns** (29,000x faster than 10ms target)
- Memory efficient

**API:**
```go
registry := collection.NewRegistryWithTTL(15 * time.Minute)

// Get with cache
collection, found := registry.GetCached(ctx, uri)

// Invalidate cache
registry.GetCache().Invalidate(uri)

// Clear all
registry.GetCache().Clear()

// Get statistics
stats := registry.GetCache().Stats()
// Returns: total_cached, expired, total_access, ttl_minutes, enabled
```

### 2. Metadata Index

Fast search across collection metadata with 4 specialized indices.

**Indices:**
1. **By Author**: `map[string][]*CollectionMetadata`
2. **By Category**: `map[string][]*CollectionMetadata`
3. **By Tag**: `map[string][]*CollectionMetadata`
4. **By Keyword**: `map[string][]*CollectionMetadata`

**Search Capabilities:**
- Author filtering
- Category filtering
- Tag matching (AND/OR)
- Keyword text search
- Pagination (limit/offset)
- Statistics

**API:**
```go
index := registry.GetMetadataIndex()

// Index a collection
index.Index(metadata)

// Search by category
results := index.Search("devops", &sources.BrowseFilter{
    Category: "devops",
    Limit:    10,
})

// Search by author
results = index.Search("", &sources.BrowseFilter{
    Author: "john@example.com",
})

// Search by tags
results = index.Search("", &sources.BrowseFilter{
    Tags: []string{"automation", "ci-cd"},
})

// Full-text search
results = index.Search("keyword search", &sources.BrowseFilter{
    Query: "kubernetes deployment",
})

// Get statistics
stats := index.Stats()
// Returns: total_collections, authors, categories, tags, keywords
```

### 3. Dependency Graph

Manages collection dependencies with cycle detection.

**Features:**
- Dependency tracking
- Circular dependency detection
- Topological sorting (install order)
- Diamond dependency resolution

**API:**
```go
graph := registry.GetDependencyGraph()

// Add nodes
graph.AddNode("github.com/user/collection-a", "Collection A", "1.0.0")
graph.AddNode("github.com/user/collection-b", "Collection B", "1.0.0")

// Add dependency: A depends on B
err := graph.AddDependency(
    "github.com/user/collection-a",
    "github.com/user/collection-b",
)
// Returns error if circular dependency detected

// Get install order
order, err := graph.TopologicalSort()
// Returns nodes in dependency-first order
```

**Cycle Detection:**
```go
// Example: A → B → C → A (circular)
graph.AddNode("A", "Collection A", "1.0.0")
graph.AddNode("B", "Collection B", "1.0.0")
graph.AddNode("C", "Collection C", "1.0.0")

graph.AddDependency("A", "B") // OK
graph.AddDependency("B", "C") // OK
err := graph.AddDependency("C", "A") // ERROR: circular dependency
```

## Usage Examples

### Basic Collection Retrieval

```go
import "github.com/fsvxavier/nexs-mcp/internal/collection"

// Create registry
registry := collection.NewRegistry()

// Add sources
registry.AddSource(sources.NewGitHubSource(token))
registry.AddSource(sources.NewLocalSource("/path/to/collections"))

// Get collection (with caching)
ctx := context.Background()
collection, err := registry.Get(ctx, "github.com/user/my-collection")
```

### Browse with Filters

```go
// Search by category
results, err := registry.Browse(ctx, &sources.BrowseFilter{
    Category: "devops",
    Limit:    10,
})

// Search by author
results, err = registry.Browse(ctx, &sources.BrowseFilter{
    Author: "john@example.com",
})

// Combined filters
results, err = registry.Browse(ctx, &sources.BrowseFilter{
    Category: "ai-ml",
    Tags:     []string{"nlp", "transformers"},
    Query:    "text generation",
    Limit:    20,
    Offset:   0,
})
```

### Cache Management

```go
// Enable/disable cache
cache := registry.GetCache()
cache.SetEnabled(true)

// Configure TTL
registry = collection.NewRegistryWithTTL(30 * time.Minute)

// Manual cache control
cache.Invalidate(uri)  // Remove specific entry
cache.Clear()          // Clear all

// Monitor cache performance
stats := cache.Stats()
fmt.Printf("Cached: %d, Expired: %d, Access: %d\n",
    stats["total_cached"],
    stats["expired"],
    stats["total_access"],
)
```

### Dependency Resolution

```go
// Build dependency graph
graph := registry.GetDependencyGraph()

// Add collections and their dependencies
for _, manifest := range manifests {
    graph.AddNode(manifest.Repository, manifest.Name, manifest.Version)
    
    for _, dep := range manifest.Dependencies {
        graph.AddDependency(manifest.Repository, dep.URI)
    }
}

// Get install order
installOrder, err := graph.TopologicalSort()
if err != nil {
    // Handle circular dependency
    log.Fatalf("Circular dependency: %v", err)
}

// Install in order
for _, node := range installOrder {
    installCollection(node.URI)
}
```

## Performance Metrics

| Operation | Target | Actual | Notes |
|-----------|--------|--------|-------|
| Cache hit | <10ms | **343ns** | 29,000x faster than target |
| Cache miss | - | ~5µs | First fetch from source |
| Index search | - | <1ms | In-memory hash lookup |
| Dependency sort | - | O(n+e) | Topological sort complexity |

## Configuration

```go
// Registry with custom TTL
registry := collection.NewRegistryWithTTL(45 * time.Minute)

// Access internal components
cache := registry.GetCache()
index := registry.GetMetadataIndex()
graph := registry.GetDependencyGraph()

// Get registry statistics
stats := registry.GetStats()
// Returns:
// - sources: number of registered sources
// - cache: cache statistics
// - metadata_index: index statistics
// - dependency_graph: graph statistics
```

## Thread Safety

All registry operations are thread-safe:
- Registry uses `sync.RWMutex` for source management
- Cache uses `sync.RWMutex` for concurrent access
- Index uses `sync.RWMutex` for search operations
- Graph uses `sync.RWMutex` for dependency tracking

## Best Practices

1. **Cache TTL**: Use 15-30 minutes for production
2. **Pagination**: Always use `Limit` for large result sets
3. **Error Handling**: Check for circular dependencies before installation
4. **Cache Warming**: Pre-populate cache for frequently accessed collections
5. **Index Rebuilding**: Call `RebuildIndex()` after bulk updates

## Troubleshooting

### Cache Not Working
- Check if cache is enabled: `cache.Stats()["enabled"]`
- Verify TTL is not too short
- Monitor expiration: `cache.Stats()["expired"]`

### Slow Search
- Use specific filters (category, author) instead of full-text
- Implement pagination for large result sets
- Consider rebuilding index: `index.Clear()` + re-index

### Circular Dependency
- Use `TopologicalSort()` to detect cycles
- Check dependency chain before adding
- Review manifest dependency declarations

## See Also

- [Publishing Guide](PUBLISHING.md)
- [Security Validation](SECURITY.md)
- [Collection Manifest](../collection-manifest-example.yaml)
- [ADR-008: Production Architecture](../adr/ADR-008-collection-registry-production.md)
