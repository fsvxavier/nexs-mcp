//go:build windows

package vectorstore

import (
	"math"
	"sync"

	"github.com/fogfish/hnsw"
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

// vectorEntry represents a vector with ID for fogfish/hnsw storage
type vectorEntry struct {
	id     string
	vector []float32
}

// cosineSurface implements vector.Surface for fogfish/hnsw
type cosineSurface struct{}

func (cosineSurface) Distance(a, b vectorEntry) float32 {
	var dot, normA, normB float32
	for i := range a.vector {
		dot += a.vector[i] * b.vector[i]
		normA += a.vector[i] * a.vector[i]
		normB += b.vector[i] * b.vector[i]
	}
	if normA == 0 || normB == 0 {
		return 1.0
	}
	similarity := dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
	return 1.0 - similarity
}

func (cosineSurface) Equal(a, b vectorEntry) bool {
	if a.id != b.id {
		return false
	}
	if len(a.vector) != len(b.vector) {
		return false
	}
	for i := range a.vector {
		if a.vector[i] != b.vector[i] {
			return false
		}
	}
	return true
}

// euclideanSurface implements vector.Surface for fogfish/hnsw
type euclideanSurface struct{}

func (euclideanSurface) Distance(a, b vectorEntry) float32 {
	var sum float32
	for i := range a.vector {
		diff := a.vector[i] - b.vector[i]
		sum += diff * diff
	}
	return float32(math.Sqrt(float64(sum)))
}

func (euclideanSurface) Equal(a, b vectorEntry) bool {
	if a.id != b.id {
		return false
	}
	if len(a.vector) != len(b.vector) {
		return false
	}
	for i := range a.vector {
		if a.vector[i] != b.vector[i] {
			return false
		}
	}
	return true
}

// dotProductSurface implements vector.Surface for fogfish/hnsw
type dotProductSurface struct{}

func (dotProductSurface) Distance(a, b vectorEntry) float32 {
	var dot float32
	for i := range a.vector {
		dot += a.vector[i] * b.vector[i]
	}
	return -dot // Negative because higher dot product = more similar
}

func (dotProductSurface) Equal(a, b vectorEntry) bool {
	if a.id != b.id {
		return false
	}
	if len(a.vector) != len(b.vector) {
		return false
	}
	for i := range a.vector {
		if a.vector[i] != b.vector[i] {
			return false
		}
	}
	return true
}

// HNSWIndex implements Hierarchical Navigable Small World index for approximate NN search.
// Based on: "Efficient and robust approximate nearest neighbor search using
// Hierarchical Navigable Small World graphs" (Malkov & Yashunin, 2018).
// Uses github.com/fogfish/hnsw pure Go implementation (v0.0.5) on Windows.
type HNSWIndex struct {
	// Configuration
	config *HNSWConfig

	// Dimension of vectors
	dimension int

	// Similarity metric
	similarity SimilarityMetric

	// Underlying HNSW index
	index *hnsw.HNSW[vectorEntry]

	// Metadata storage
	metadata     map[string]map[string]interface{}
	metadataLock sync.RWMutex

	// Vector storage (fogfish stores values, not IDs separately)
	vectors     map[string][]float32
	vectorsLock sync.RWMutex

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHNSWIndex creates a new HNSW index using fogfish/hnsw pure Go implementation (Windows).
func NewHNSWIndex(dimension int, similarity SimilarityMetric, config *HNSWConfig) (*HNSWIndex, error) {
	if config == nil {
		config = DefaultHNSWConfig()
	}

	// Map similarity metric to surface implementation
	var index *hnsw.HNSW[vectorEntry]

	switch similarity {
	case SimilarityCosine:
		index = hnsw.New[vectorEntry](
			cosineSurface{},
			hnsw.WithM(config.M),
			hnsw.WithEfConstruction(config.EfSearch*5), // ef_construction typically higher
		)
	case SimilarityEuclidean:
		index = hnsw.New[vectorEntry](
			euclideanSurface{},
			hnsw.WithM(config.M),
			hnsw.WithEfConstruction(config.EfSearch*5),
		)
	case SimilarityDotProduct:
		index = hnsw.New[vectorEntry](
			dotProductSurface{},
			hnsw.WithM(config.M),
			hnsw.WithEfConstruction(config.EfSearch*5),
		)
	default:
		index = hnsw.New[vectorEntry](
			cosineSurface{},
			hnsw.WithM(config.M),
			hnsw.WithEfConstruction(config.EfSearch*5),
		)
	}

	return &HNSWIndex{
		config:     config,
		dimension:  dimension,
		similarity: similarity,
		index:      index,
		metadata:   make(map[string]map[string]interface{}),
		vectors:    make(map[string][]float32),
	}, nil
}

// Add inserts a new vector into the HNSW index.
func (h *HNSWIndex) Add(id string, vector []float32, metadata map[string]interface{}) error {
	if len(vector) != h.dimension {
		return ErrDimensionMismatch
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Create entry
	entry := vectorEntry{
		id:     id,
		vector: make([]float32, len(vector)),
	}
	copy(entry.vector, vector)

	// Insert into index
	h.index.Insert(entry)

	// Store vector for retrieval
	h.vectorsLock.Lock()
	h.vectors[id] = make([]float32, len(vector))
	copy(h.vectors[id], vector)
	h.vectorsLock.Unlock()

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
	size := len(h.vectors)
	h.mu.RUnlock()

	if size == 0 {
		return []SearchResult{}, nil
	}

	// Create query entry
	queryEntry := vectorEntry{
		id:     "",
		vector: query,
	}

	// Search using fogfish/hnsw
	h.mu.RLock()
	results := h.index.Search(queryEntry, k, h.config.EfSearch)
	h.mu.RUnlock()

	// Convert to SearchResult
	searchResults := make([]SearchResult, 0, len(results))

	for _, entry := range results {
		// Get metadata
		h.metadataLock.RLock()
		metadata := h.metadata[entry.id]
		h.metadataLock.RUnlock()

		// Calculate distance for score
		distance := h.calculateDistance(query, entry.vector)

		// Distance to similarity
		similarity := h.distanceToSimilarity(float64(distance))

		searchResults = append(searchResults, SearchResult{
			ID:         entry.id,
			Vector:     entry.vector,
			Metadata:   metadata,
			Score:      float64(distance),
			Similarity: similarity,
		})
	}

	return searchResults, nil
}

// Get retrieves a vector by ID.
func (h *HNSWIndex) Get(id string) (VectorEntry, bool) {
	h.vectorsLock.RLock()
	vector, exists := h.vectors[id]
	h.vectorsLock.RUnlock()

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
	h.vectorsLock.Lock()
	_, exists := h.vectors[id]
	if !exists {
		h.vectorsLock.Unlock()
		return ErrVectorNotFound
	}
	delete(h.vectors, id)
	h.vectorsLock.Unlock()

	// Remove metadata
	h.metadataLock.Lock()
	delete(h.metadata, id)
	h.metadataLock.Unlock()

	// Note: fogfish/hnsw doesn't have a Delete method,
	// so we only remove from our tracking maps.
	// The vector remains in the index but becomes unreachable.

	return nil
}

// Size returns the number of vectors in the index.
func (h *HNSWIndex) Size() int {
	h.vectorsLock.RLock()
	defer h.vectorsLock.RUnlock()
	return len(h.vectors)
}

// Clear removes all vectors from the index.
func (h *HNSWIndex) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Recreate index with same configuration
	var index *hnsw.HNSW[vectorEntry]

	switch h.similarity {
	case SimilarityCosine:
		index = hnsw.New[vectorEntry](
			cosineSurface{},
			hnsw.WithM(h.config.M),
			hnsw.WithEfConstruction(h.config.EfSearch*5),
		)
	case SimilarityEuclidean:
		index = hnsw.New[vectorEntry](
			euclideanSurface{},
			hnsw.WithM(h.config.M),
			hnsw.WithEfConstruction(h.config.EfSearch*5),
		)
	case SimilarityDotProduct:
		index = hnsw.New[vectorEntry](
			dotProductSurface{},
			hnsw.WithM(h.config.M),
			hnsw.WithEfConstruction(h.config.EfSearch*5),
		)
	default:
		index = hnsw.New[vectorEntry](
			cosineSurface{},
			hnsw.WithM(h.config.M),
			hnsw.WithEfConstruction(h.config.EfSearch*5),
		)
	}

	h.index = index

	h.vectorsLock.Lock()
	h.vectors = make(map[string][]float32)
	h.vectorsLock.Unlock()

	h.metadataLock.Lock()
	h.metadata = make(map[string]map[string]interface{})
	h.metadataLock.Unlock()
}

// SetEf dynamically adjusts the efSearch parameter for subsequent searches.
func (h *HNSWIndex) SetEf(ef int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config.EfSearch = ef
}

// --- Internal methods ---

// calculateDistance calculates distance based on similarity metric
func (h *HNSWIndex) calculateDistance(a, b []float32) float32 {
	switch h.similarity {
	case SimilarityCosine:
		return cosineSurface{}.Distance(
			vectorEntry{vector: a},
			vectorEntry{vector: b},
		)
	case SimilarityEuclidean:
		return euclideanSurface{}.Distance(
			vectorEntry{vector: a},
			vectorEntry{vector: b},
		)
	case SimilarityDotProduct:
		return dotProductSurface{}.Distance(
			vectorEntry{vector: a},
			vectorEntry{vector: b},
		)
	default:
		return cosineSurface{}.Distance(
			vectorEntry{vector: a},
			vectorEntry{vector: b},
		)
	}
}

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
