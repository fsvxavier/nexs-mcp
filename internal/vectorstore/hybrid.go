package vectorstore

import (
	"errors"
	"fmt"
	"math"
	"sync"
)

// HybridConfig holds configuration for hybrid vector store.
type HybridConfig struct {
	// SwitchThreshold is the number of vectors at which to switch from linear to HNSW.
	// Default: 100 vectors.
	SwitchThreshold int

	// HNSWConfig for HNSW index when threshold is exceeded.
	HNSWConfig *HNSWConfig

	// Similarity metric to use.
	Similarity SimilarityMetric

	// Dimension of vectors.
	Dimension int
}

// DefaultHybridConfig returns hybrid config with recommended defaults.
func DefaultHybridConfig(dimension int) *HybridConfig {
	return &HybridConfig{
		SwitchThreshold: 100,
		HNSWConfig:      DefaultHNSWConfig(),
		Similarity:      SimilarityCosine,
		Dimension:       dimension,
	}
}

// HybridStore automatically switches between linear search and HNSW based on data size.
// - For <SwitchThreshold vectors: uses fast linear search (O(n))
// - For >=SwitchThreshold vectors: uses HNSW (O(log n))
// This provides optimal performance across different dataset sizes.
type HybridStore struct {
	config *HybridConfig

	// Linear storage (used when size < threshold)
	linearVectors map[string]*VectorEntry
	linearMu      sync.RWMutex

	// HNSW index (initialized when size >= threshold)
	hnswIndex *HNSWIndex
	hnswMu    sync.RWMutex

	// Current mode
	useHNSW bool
	modeMu  sync.RWMutex
}

// VectorEntry represents a stored vector with metadata.
type VectorEntry struct {
	ID       string
	Vector   []float32
	Metadata map[string]interface{}
}

// SearchResult represents a search result with similarity score.
type SearchResult struct {
	ID         string
	Vector     []float32
	Metadata   map[string]interface{}
	Score      float64 // Distance score (lower is better)
	Similarity float64 // Similarity score (higher is better)
}

// SimilarityMetric defines the similarity metric to use.
type SimilarityMetric int

const (
	SimilarityCosine SimilarityMetric = iota
	SimilarityEuclidean
	SimilarityDotProduct
)

var (
	ErrDimensionMismatch = errors.New("vector dimension mismatch")
	ErrVectorNotFound    = errors.New("vector not found")
	ErrVectorExists      = errors.New("vector already exists")
)

// NewHybridStore creates a new hybrid vector store.
func NewHybridStore(config *HybridConfig) *HybridStore {
	if config == nil {
		config = DefaultHybridConfig(384) // Default dimension for common models
	}

	return &HybridStore{
		config:        config,
		linearVectors: make(map[string]*VectorEntry),
		hnswIndex:     nil,
		useHNSW:       false,
	}
}

// Add inserts a vector into the store.
func (h *HybridStore) Add(id string, vector []float32, metadata map[string]interface{}) error {
	if len(vector) != h.config.Dimension {
		return fmt.Errorf("%w: expected %d, got %d", ErrDimensionMismatch, h.config.Dimension, len(vector))
	}

	h.modeMu.RLock()
	currentMode := h.useHNSW
	h.modeMu.RUnlock()

	if currentMode {
		// HNSW mode
		h.hnswMu.Lock()
		err := h.hnswIndex.Add(id, vector, metadata)
		h.hnswMu.Unlock()
		return err
	}

	// Linear mode
	h.linearMu.Lock()
	defer h.linearMu.Unlock()

	// Check if already exists
	if _, exists := h.linearVectors[id]; exists {
		return ErrVectorExists
	}

	h.linearVectors[id] = &VectorEntry{
		ID:       id,
		Vector:   vector,
		Metadata: metadata,
	}

	// Check if we should switch to HNSW
	if len(h.linearVectors) >= h.config.SwitchThreshold {
		h.linearMu.Unlock()
		h.migrateToHNSW()
		h.linearMu.Lock()
	}

	return nil
}

// Search performs k-NN search.
func (h *HybridStore) Search(query []float32, k int) ([]SearchResult, error) {
	if len(query) != h.config.Dimension {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrDimensionMismatch, h.config.Dimension, len(query))
	}

	h.modeMu.RLock()
	currentMode := h.useHNSW
	h.modeMu.RUnlock()

	if currentMode {
		// HNSW search
		h.hnswMu.RLock()
		results, err := h.hnswIndex.Search(query, k)
		h.hnswMu.RUnlock()
		return results, err
	}

	// Linear search
	return h.linearSearch(query, k)
}

// Get retrieves a vector by ID.
func (h *HybridStore) Get(id string) (*VectorEntry, error) {
	h.modeMu.RLock()
	currentMode := h.useHNSW
	h.modeMu.RUnlock()

	if currentMode {
		h.hnswMu.RLock()
		entry, exists := h.hnswIndex.Get(id)
		h.hnswMu.RUnlock()

		if !exists {
			return nil, ErrVectorNotFound
		}
		return &entry, nil
	}

	h.linearMu.RLock()
	defer h.linearMu.RUnlock()

	entry, exists := h.linearVectors[id]
	if !exists {
		return nil, ErrVectorNotFound
	}

	return entry, nil
}

// Delete removes a vector from the store.
func (h *HybridStore) Delete(id string) error {
	h.modeMu.RLock()
	currentMode := h.useHNSW
	h.modeMu.RUnlock()

	if currentMode {
		h.hnswMu.Lock()
		err := h.hnswIndex.Delete(id)
		h.hnswMu.Unlock()
		return err
	}

	h.linearMu.Lock()
	defer h.linearMu.Unlock()

	if _, exists := h.linearVectors[id]; !exists {
		return ErrVectorNotFound
	}

	delete(h.linearVectors, id)
	return nil
}

// Size returns the number of vectors in the store.
func (h *HybridStore) Size() int {
	h.modeMu.RLock()
	currentMode := h.useHNSW
	h.modeMu.RUnlock()

	if currentMode {
		h.hnswMu.RLock()
		size := h.hnswIndex.Size()
		h.hnswMu.RUnlock()
		return size
	}

	h.linearMu.RLock()
	size := len(h.linearVectors)
	h.linearMu.RUnlock()
	return size
}

// Clear removes all vectors from the store.
func (h *HybridStore) Clear() {
	h.modeMu.Lock()
	h.useHNSW = false
	h.modeMu.Unlock()

	h.linearMu.Lock()
	h.linearVectors = make(map[string]*VectorEntry)
	h.linearMu.Unlock()

	h.hnswMu.Lock()
	if h.hnswIndex != nil {
		h.hnswIndex.Clear()
	}
	h.hnswIndex = nil
	h.hnswMu.Unlock()
}

// IsUsingHNSW returns whether the store is currently using HNSW index.
func (h *HybridStore) IsUsingHNSW() bool {
	h.modeMu.RLock()
	defer h.modeMu.RUnlock()
	return h.useHNSW
}

// --- Internal methods ---

// migrateToHNSW switches from linear to HNSW mode by migrating all vectors.
func (h *HybridStore) migrateToHNSW() {
	h.modeMu.Lock()
	defer h.modeMu.Unlock()

	// Already using HNSW
	if h.useHNSW {
		return
	}

	h.linearMu.RLock()
	// Copy all vectors before unlocking
	vectorsCopy := make(map[string]*VectorEntry, len(h.linearVectors))
	for id, entry := range h.linearVectors {
		vectorsCopy[id] = entry
	}
	h.linearMu.RUnlock()

	// Create HNSW index
	h.hnswMu.Lock()
	index, err := NewHNSWIndex(h.config.Dimension, h.config.Similarity, h.config.HNSWConfig)
	if err != nil {
		h.hnswMu.Unlock()
		fmt.Printf("Warning: failed to create HNSW index: %v\n", err)
		return
	}
	h.hnswIndex = index

	// Migrate all vectors
	for _, entry := range vectorsCopy {
		if err := h.hnswIndex.Add(entry.ID, entry.Vector, entry.Metadata); err != nil {
			// Log error but continue migration
			fmt.Printf("Warning: failed to migrate vector %s to HNSW: %v\n", entry.ID, err)
		}
	}
	h.hnswMu.Unlock()

	// Switch mode
	h.useHNSW = true

	// Clear linear storage to free memory
	h.linearMu.Lock()
	h.linearVectors = make(map[string]*VectorEntry)
	h.linearMu.Unlock()
}

// linearSearch performs brute-force k-NN search on linear vectors.
func (h *HybridStore) linearSearch(query []float32, k int) ([]SearchResult, error) {
	h.linearMu.RLock()
	defer h.linearMu.RUnlock()

	if len(h.linearVectors) == 0 {
		return []SearchResult{}, nil
	}

	// Calculate similarities for all vectors
	type scoredResult struct {
		entry *VectorEntry
		score float64
	}

	results := make([]scoredResult, 0, len(h.linearVectors))

	for _, entry := range h.linearVectors {
		var score float64

		switch h.config.Similarity {
		case SimilarityCosine:
			score = float64(CosineSimilarity(query, entry.Vector))
		case SimilarityEuclidean:
			// For Euclidean, use negative distance as score (higher is better)
			score = -EuclideanDistance(query, entry.Vector)
		case SimilarityDotProduct:
			score = DotProduct(query, entry.Vector)
		default:
			score = float64(CosineSimilarity(query, entry.Vector))
		}

		results = append(results, scoredResult{
			entry: entry,
			score: score,
		})
	}

	// Sort by score descending (higher similarity = better match)
	for i := range len(results) - 1 {
		for j := i + 1; j < len(results); j++ {
			if results[i].score < results[j].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Return top k
	maxResults := k
	if maxResults > len(results) {
		maxResults = len(results)
	}

	finalResults := make([]SearchResult, maxResults)
	for i := range maxResults {
		finalResults[i] = SearchResult{
			ID:         results[i].entry.ID,
			Vector:     results[i].entry.Vector,
			Metadata:   results[i].entry.Metadata,
			Score:      results[i].score,
			Similarity: results[i].score,
		}
	}

	return finalResults, nil
}

// --- Similarity functions ---

// CosineSimilarity calculates cosine similarity between two vectors.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// EuclideanDistance calculates Euclidean distance between two vectors.
func EuclideanDistance(a, b []float32) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range a {
		diff := float64(a[i] - b[i])
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// DotProduct calculates dot product between two vectors.
func DotProduct(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var sum float64
	for i := range a {
		sum += float64(a[i] * b[i])
	}

	return sum
}
