package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
	"github.com/fsvxavier/nexs-mcp/internal/indexing/hnsw"
	"github.com/fsvxavier/nexs-mcp/internal/vectorstore"
)

const (
	// HNSWThreshold defines minimum vectors to use HNSW (vs linear search).
	HNSWThreshold = 100

	// DefaultHNSWM is the default M parameter for HNSW.
	DefaultHNSWM = 16

	// DefaultHNSWEfConstruction is the default efConstruction for HNSW.
	DefaultHNSWEfConstruction = 200

	// DefaultHNSWEfSearch is the default efSearch for HNSW queries.
	DefaultHNSWEfSearch = 50
)

// HybridSearchService combines linear and HNSW search with automatic fallback.
type HybridSearchService struct {
	linearStore    *vectorstore.Store
	hnswIndex      *hnsw.Graph
	provider       embeddings.Provider
	useHNSW        bool
	hnswPath       string
	mu             sync.RWMutex
	autoReindex    bool
	reindexCounter int
	adaptiveCache  domain.CacheService // Cache for search results and embeddings
}

// HybridSearchConfig holds configuration for hybrid search.
type HybridSearchConfig struct {
	Provider        embeddings.Provider
	HNSWPath        string
	M               int
	EfConstruction  int
	EfSearch        int
	AutoReindex     bool
	ReindexInterval int // Reindex every N insertions
}

// NewHybridSearchService creates a new hybrid search service.
func NewHybridSearchService(config HybridSearchConfig) *HybridSearchService {
	if config.M == 0 {
		config.M = DefaultHNSWM
	}
	if config.EfConstruction == 0 {
		config.EfConstruction = DefaultHNSWEfConstruction
	}
	if config.EfSearch == 0 {
		config.EfSearch = DefaultHNSWEfSearch
	}
	if config.ReindexInterval == 0 {
		config.ReindexInterval = 100
	}

	// Determine distance function from provider config
	distFunc := hnsw.CosineSimilarity // Default

	hnswGraph := hnsw.NewGraph(distFunc)
	hnswGraph.SetParameters(config.M, config.EfConstruction)

	return &HybridSearchService{
		linearStore:    vectorstore.NewStore(config.Provider),
		hnswIndex:      hnswGraph,
		provider:       config.Provider,
		hnswPath:       config.HNSWPath,
		autoReindex:    config.AutoReindex,
		reindexCounter: 0,
	}
}

// Provider returns the embedding provider.
func (h *HybridSearchService) Provider() embeddings.Provider {
	return h.provider
}

// SetAdaptiveCache sets the adaptive cache for this search service.
func (h *HybridSearchService) SetAdaptiveCache(cache domain.CacheService) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.adaptiveCache = cache
}

// Add adds a document to the hybrid search index.
func (h *HybridSearchService) Add(ctx context.Context, id, text string, metadata map[string]interface{}) error {
	// Always add to linear store
	if err := h.linearStore.Add(ctx, id, text, metadata); err != nil {
		return err
	}

	// Get embedding for HNSW
	embedding, err := h.provider.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to HNSW index
	if err := h.hnswIndex.Insert(id, embedding); err != nil {
		return fmt.Errorf("failed to insert into HNSW: %w", err)
	}

	h.reindexCounter++

	// Check if we should switch to HNSW mode
	if !h.useHNSW && h.hnswIndex.Size() >= HNSWThreshold {
		h.useHNSW = true
	}

	// Auto-save HNSW index periodically
	if h.autoReindex && h.hnswPath != "" && h.reindexCounter%100 == 0 {
		go func() { _ = h.SaveIndex() }() // Non-blocking save
	}

	return nil
}

// Search performs hybrid search with automatic fallback.
func (h *HybridSearchService) Search(ctx context.Context, query string, limit int, filters map[string]interface{}) ([]embeddings.Result, error) {
	h.mu.RLock()
	useHNSW := h.useHNSW
	cache := h.adaptiveCache
	h.mu.RUnlock()

	// Try cache first (for repeated searches)
	if cache != nil {
		cacheKey := fmt.Sprintf("search:%s:limit=%d", query, limit)
		if cached, found := cache.Get(ctx, cacheKey); found {
			return cached.([]embeddings.Result), nil
		}
	}

	// If below threshold, use linear search
	if !useHNSW {
		results, err := h.linearStore.Search(ctx, query, limit, filters)
		if err == nil && cache != nil {
			// Cache successful search results (estimate ~500 bytes per result)
			_ = cache.Set(ctx, fmt.Sprintf("search:%s:limit=%d", query, limit), results, len(results)*500)
		}
		return results, err
	}

	// Try to get cached embedding for query
	var embedding []float32
	var err error
	if cache != nil {
		embedCacheKey := fmt.Sprintf("embedding:%s", query)
		if cached, found := cache.Get(ctx, embedCacheKey); found {
			embedding = cached.([]float32)
		} else {
			// Generate and cache embedding
			embedding, err = h.provider.Embed(ctx, query)
			if err == nil {
				// Cache embedding (4 bytes per dimension * dimensions)
				_ = cache.Set(ctx, embedCacheKey, embedding, len(embedding)*4)
			}
		}
	} else {
		// No cache available, generate directly
		embedding, err = h.provider.Embed(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search HNSW index
	searchResults, err := h.hnswIndex.SearchKNN(embedding, limit)
	if err != nil {
		// Fallback to linear search on error
		return h.linearStore.Search(ctx, query, limit, filters)
	}

	// Convert HNSW results to embeddings.Result format
	results := make([]embeddings.Result, 0, len(searchResults))
	for _, sr := range searchResults {
		// Apply filters if needed
		if !h.matchesFilters(sr.ID, filters) {
			continue
		}

		results = append(results, embeddings.Result{
			ID:       sr.ID,
			Score:    float64(1.0 - sr.Distance), // Convert distance to similarity score
			Metadata: h.getMetadata(sr.ID),
		})
	}

	// Cache successful search results
	if cache != nil {
		cacheKey := fmt.Sprintf("search:%s:limit=%d", query, limit)
		_ = cache.Set(ctx, cacheKey, results, len(results)*500) // ~500 bytes per result
	}

	return results, nil
}

// SearchWithHNSW explicitly uses HNSW search with custom efSearch.
func (h *HybridSearchService) SearchWithHNSW(ctx context.Context, query string, limit int, efSearch int) ([]embeddings.Result, error) {
	embedding, err := h.provider.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	searchResults, err := h.hnswIndex.Search(embedding, limit, efSearch)
	if err != nil {
		return nil, err
	}

	results := make([]embeddings.Result, 0, len(searchResults))
	for _, sr := range searchResults {
		results = append(results, embeddings.Result{
			ID:       sr.ID,
			Score:    float64(1.0 - sr.Distance),
			Metadata: h.getMetadata(sr.ID),
		})
	}

	return results, nil
}

// RebuildIndex rebuilds the HNSW index from linear store.
func (h *HybridSearchService) RebuildIndex(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Create new HNSW graph
	distFunc := hnsw.CosineSimilarity
	newGraph := hnsw.NewGraph(distFunc)
	newGraph.SetParameters(DefaultHNSWM, DefaultHNSWEfConstruction)

	// Get all documents from linear store
	allDocs := h.linearStore.GetAll()

	// Re-index all documents
	for _, doc := range allDocs {
		if err := newGraph.Insert(doc.ID, doc.Vector); err != nil {
			return fmt.Errorf("failed to insert %s into new index: %w", doc.ID, err)
		}
	}

	// Replace old index
	h.hnswIndex = newGraph
	h.useHNSW = newGraph.Size() >= HNSWThreshold
	h.reindexCounter = 0

	return nil
}

// SaveIndex saves the HNSW index to disk.
func (h *HybridSearchService) SaveIndex() error {
	if h.hnswPath == "" {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.hnswIndex.Save(h.hnswPath)
}

// LoadIndex loads the HNSW index from disk.
func (h *HybridSearchService) LoadIndex() error {
	if h.hnswPath == "" {
		return errors.New("no HNSW path configured")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.hnswIndex.Load(h.hnswPath); err != nil {
		return err
	}

	h.useHNSW = h.hnswIndex.Size() >= HNSWThreshold
	return nil
}

// Delete removes a document from both indexes.
func (h *HybridSearchService) Delete(ctx context.Context, id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from linear store
	if err := h.linearStore.Delete(id); err != nil {
		return err
	}

	// Remove from HNSW index
	if err := h.hnswIndex.Delete(id); err != nil {
		return fmt.Errorf("failed to delete from HNSW: %w", err)
	}

	// Check if we should switch back to linear search
	if h.useHNSW && h.hnswIndex.Size() < HNSWThreshold {
		h.useHNSW = false
	}

	return nil
}

// Clear removes all documents from both indexes.
func (h *HybridSearchService) Clear() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.linearStore.Clear()
	h.hnswIndex.Clear()
	h.useHNSW = false
	h.reindexCounter = 0

	return nil
}

// GetStatistics returns statistics about both indexes.
func (h *HybridSearchService) GetStatistics() HybridSearchStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	hnswStats := h.hnswIndex.GetStatistics()

	return HybridSearchStats{
		TotalDocuments: h.linearStore.Size(),
		UseHNSW:        h.useHNSW,
		HNSWThreshold:  HNSWThreshold,
		HNSWNodeCount:  hnswStats.NodeCount,
		HNSWMaxLevel:   hnswStats.MaxLevel,
		CacheSize:      0,
		CacheHits:      0,
		CacheMisses:    0,
	}
}

// HybridSearchStats holds statistics for hybrid search.
type HybridSearchStats struct {
	TotalDocuments int
	UseHNSW        bool
	HNSWThreshold  int
	HNSWNodeCount  int
	HNSWMaxLevel   int
	CacheSize      int
	CacheHits      int64
	CacheMisses    int64
}

// matchesFilters checks if a document ID matches the given filters.
func (h *HybridSearchService) matchesFilters(id string, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}

	metadata := h.getMetadata(id)
	if metadata == nil {
		return false
	}

	for key, expectedValue := range filters {
		actualValue, exists := metadata[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

// getMetadata retrieves metadata for a document ID from linear store.
func (h *HybridSearchService) getMetadata(id string) map[string]interface{} {
	doc := h.linearStore.GetByID(id)
	if doc == nil {
		return nil
	}
	return doc.Metadata
}

// AutoSave starts a background goroutine to periodically save the index.
func (h *HybridSearchService) AutoSave(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := h.SaveIndex(); err != nil {
				// Log error but continue
				fmt.Printf("Failed to auto-save HNSW index: %v\n", err)
			}
		case <-stopCh:
			// Final save before stopping
			_ = h.SaveIndex()
			return
		}
	}
}
