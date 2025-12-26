package vectorstore

import (
	"fmt"
	"sync"

	"github.com/TFMV/hnsw"
)

// HNSWConfig holds configuration for HNSW index.
type HNSWConfig struct {
	// M is the maximum number of neighbors to keep for each node.
	// Reasonable range: 8-64. Default: 16.
	M int

	// Ml is the level generation factor (fraction of nodes at each level).
	// E.g., for Ml = 0.25, each layer is 1/4 the size of the previous layer.
	// Default: 0.25.
	Ml float64

	// EfSearch is the number of nodes to consider in the search phase.
	// Higher values improve recall but increase search time. Default: 20.
	EfSearch int

	// Seed for random number generator.
	Seed int64
}

// DefaultHNSWConfig returns a HNSW config with recommended defaults.
func DefaultHNSWConfig() *HNSWConfig {
	return &HNSWConfig{
		M:        16,
		Ml:       0.25,
		EfSearch: 20,
		Seed:     42,
	}
}

// HNSWIndex implements Hierarchical Navigable Small World index for approximate NN search.
// Based on: "Efficient and robust approximate nearest neighbor search using
// Hierarchical Navigable Small World graphs" (Malkov & Yashunin, 2018).
// Uses github.com/TFMV/hnsw pure Go implementation (v0.4.0, March 2025).
type HNSWIndex struct {
	// Configuration
	config *HNSWConfig

	// Dimension of vectors
	dimension int

	// Similarity metric
	similarity SimilarityMetric

	// Underlying HNSW graph
	graph *hnsw.Graph[string]

	// Metadata storage (graph doesn't store metadata)
	metadata     map[string]map[string]interface{}
	metadataLock sync.RWMutex

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHNSWIndex creates a new HNSW index using TFMV/hnsw pure Go implementation.
func NewHNSWIndex(dimension int, similarity SimilarityMetric, config *HNSWConfig) (*HNSWIndex, error) {
	if config == nil {
		config = DefaultHNSWConfig()
	}

	// Map similarity metric to distance function
	var distanceFn hnsw.DistanceFunc

	switch similarity {
	case SimilarityCosine:
		// Use built-in cosine distance
		distanceFn = hnsw.CosineDistance
	case SimilarityEuclidean:
		// Use built-in Euclidean distance
		distanceFn = hnsw.EuclideanDistance
	case SimilarityDotProduct:
		// Custom: negative dot product (lower is more similar)
		distanceFn = func(a, b []float32) float32 {
			return -float32(DotProduct(a, b))
		}
	default:
		distanceFn = hnsw.CosineDistance
	}

	// Create HNSW graph with configuration
	graph, err := hnsw.NewGraphWithConfig[string](
		config.M,
		config.Ml,
		config.EfSearch,
		distanceFn,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HNSW graph: %w", err)
	}

	return &HNSWIndex{
		config:     config,
		dimension:  dimension,
		similarity: similarity,
		graph:      graph,
		metadata:   make(map[string]map[string]interface{}),
	}, nil
}

// Add inserts a new vector into the HNSW index.
func (h *HNSWIndex) Add(id string, vector []float32, metadata map[string]interface{}) error {
	if len(vector) != h.dimension {
		return ErrDimensionMismatch
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Create node and add to graph
	node := hnsw.MakeNode(id, vector)
	err := h.graph.Add(node)
	if err != nil {
		return fmt.Errorf("failed to add node to HNSW graph: %w", err)
	}

	// Store metadata
	if metadata != nil {
		h.metadataLock.Lock()
		h.metadata[id] = metadata
		h.metadataLock.Unlock()
	}

	return nil
}

// Search performs approximate k-NN search using HNSW algorithm.
func (h *HNSWIndex) Search(query []float32, k int) ([]SearchResult, error) {
	if len(query) != h.dimension {
		return nil, ErrDimensionMismatch
	}

	h.mu.RLock()
	size := h.graph.Len()
	h.mu.RUnlock()

	if size == 0 {
		return []SearchResult{}, nil
	}

	// Search using TFMV/hnsw
	h.mu.RLock()
	nodes, err := h.graph.Search(query, k)
	h.mu.RUnlock()

	if err != nil {
		return nil, fmt.Errorf("HNSW search failed: %w", err)
	}

	// Convert to results
	results := make([]SearchResult, 0, len(nodes))

	for _, node := range nodes {
		// Get metadata
		h.metadataLock.RLock()
		metadata := h.metadata[node.Key]
		h.metadataLock.RUnlock()

		// Calculate distance for score
		distance := h.graph.Distance(query, node.Value)

		// Distance to similarity
		similarity := h.distanceToSimilarity(float64(distance))

		results = append(results, SearchResult{
			ID:         node.Key,
			Vector:     node.Value,
			Metadata:   metadata,
			Score:      float64(distance),
			Similarity: similarity,
		})
	}

	return results, nil
}

// Get retrieves a vector by ID.
func (h *HNSWIndex) Get(id string) (VectorEntry, bool) {
	h.mu.RLock()
	vector, exists := h.graph.Lookup(id)
	h.mu.RUnlock()

	if !exists {
		return VectorEntry{}, false
	}

	h.metadataLock.RLock()
	metadata := h.metadata[id]
	h.metadataLock.RUnlock()

	return VectorEntry{
		ID:       id,
		Vector:   vector,
		Metadata: metadata,
	}, true
}

// Delete removes a vector from the index.
func (h *HNSWIndex) Delete(id string) error {
	h.mu.Lock()
	removed := h.graph.Delete(id)
	h.mu.Unlock()

	if !removed {
		return ErrVectorNotFound
	}

	// Remove metadata
	h.metadataLock.Lock()
	delete(h.metadata, id)
	h.metadataLock.Unlock()

	return nil
}

// Size returns the number of vectors in the index.
func (h *HNSWIndex) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.graph.Len()
}

// Clear removes all vectors from the index.
func (h *HNSWIndex) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Recreate graph with same configuration
	graph, err := hnsw.NewGraphWithConfig[string](
		h.config.M,
		h.config.Ml,
		h.config.EfSearch,
		h.graph.Distance,
	)
	if err != nil {
		// This should never happen since config was already validated
		panic(fmt.Sprintf("failed to recreate HNSW graph: %v", err))
	}

	h.graph = graph

	h.metadataLock.Lock()
	h.metadata = make(map[string]map[string]interface{})
	h.metadataLock.Unlock()
}

// SetEf dynamically adjusts the efSearch parameter for subsequent searches.
func (h *HNSWIndex) SetEf(ef int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config.EfSearch = ef
	h.graph.EfSearch = ef
}

// --- Internal methods ---

// distanceToSimilarity converts distance back to similarity score.
func (h *HNSWIndex) distanceToSimilarity(distance float64) float64 {
	switch h.similarity {
	case SimilarityCosine:
		// Cosine distance: 0 = identical, 2 = opposite
		// Convert to similarity: 1 = identical, 0 = orthogonal
		return 1.0 - distance
	case SimilarityEuclidean:
		// Euclidean distance: 0 = identical, larger = more different
		// Convert to similarity using 1/(1+distance)
		return 1.0 / (1.0 + distance)
	case SimilarityDotProduct:
		// Distance is negative dot product
		// Convert back to positive similarity
		return -distance
	default:
		return 1.0 - distance
	}
}
