// Package vectorstore provides in-memory vector storage and similarity search.
package vectorstore

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
)

// Vector represents a stored vector with metadata.
type Vector struct {
	ID        string                 `json:"id"`
	Embedding []float32              `json:"embedding"`
	Vector    []float32              `json:"vector"` // Alias for HNSW compatibility
	Text      string                 `json:"text"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Store manages an in-memory vector database with similarity search.
type Store struct {
	vectors  map[string]*Vector
	provider embeddings.Provider
	metric   embeddings.SimilarityMetric
	mu       sync.RWMutex
}

// NewStore creates a new vector store.
func NewStore(provider embeddings.Provider) *Store {
	return &Store{
		vectors:  make(map[string]*Vector),
		provider: provider,
		metric:   embeddings.CosineSimilarity,
	}
}

// SetMetric sets the similarity metric for searches.
func (s *Store) SetMetric(metric embeddings.SimilarityMetric) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metric = metric
}

// Add stores a vector in the database.
func (s *Store) Add(ctx context.Context, id string, text string, metadata map[string]interface{}) error {
	if id == "" {
		return errors.New("vector ID cannot be empty")
	}

	if text == "" {
		return errors.New("text cannot be empty")
	}

	// Generate embedding
	embedding, err := s.provider.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.vectors[id] = &Vector{
		ID:        id,
		Embedding: embedding,
		Vector:    embedding, // Copy for HNSW compatibility
		Text:      text,
		Metadata:  metadata,
	}

	return nil
}

// AddBatch stores multiple vectors efficiently.
func (s *Store) AddBatch(ctx context.Context, items []struct {
	ID       string
	Text     string
	Metadata map[string]interface{}
}) error {
	if len(items) == 0 {
		return errors.New("empty batch")
	}

	// Extract texts for batch embedding
	texts := make([]string, len(items))
	for i, item := range items {
		if item.ID == "" {
			return fmt.Errorf("item %d: ID cannot be empty", i)
		}
		if item.Text == "" {
			return fmt.Errorf("item %d: text cannot be empty", i)
		}
		texts[i] = item.Text
	}

	// Generate embeddings in batch
	embeddings, err := s.provider.EmbedBatch(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, item := range items {
		s.vectors[item.ID] = &Vector{
			ID:        item.ID,
			Embedding: embeddings[i],
			Vector:    embeddings[i], // Copy for HNSW compatibility
			Text:      item.Text,
			Metadata:  item.Metadata,
		}
	}

	return nil
}

// Get retrieves a vector by ID.
func (s *Store) Get(id string) (*Vector, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vec, ok := s.vectors[id]
	if !ok {
		return nil, fmt.Errorf("vector not found: %s", id)
	}

	return vec, nil
}

// Delete removes a vector from the store.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.vectors[id]; !ok {
		return fmt.Errorf("vector not found: %s", id)
	}

	delete(s.vectors, id)
	return nil
}

// Search finds the k most similar vectors to a query text.
func (s *Store) Search(ctx context.Context, query string, k int, filters map[string]interface{}) ([]embeddings.Result, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	if k <= 0 {
		return nil, errors.New("k must be positive")
	}

	// Generate query embedding
	queryEmbedding, err := s.provider.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return s.SearchByVector(queryEmbedding, k, filters), nil
}

// SearchByVector finds the k most similar vectors to a query embedding.
func (s *Store) SearchByVector(queryEmbedding []float32, k int, filters map[string]interface{}) []embeddings.Result {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []embeddings.Result

	for _, vec := range s.vectors {
		// Apply filters
		if !matchesFilters(vec.Metadata, filters) {
			continue
		}

		// Calculate similarity
		similarity := s.calculateSimilarity(queryEmbedding, vec.Embedding)

		results = append(results, embeddings.Result{
			ID:         vec.ID,
			Text:       vec.Text,
			Metadata:   vec.Metadata,
			Embedding:  vec.Embedding,
			Score:      similarity,
			Similarity: similarity,
		})
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return top k
	if k < len(results) {
		results = results[:k]
	}

	return results
}

// Size returns the number of vectors in the store.
func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.vectors)
}

// Clear removes all vectors from the store.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vectors = make(map[string]*Vector)
}

// calculateSimilarity computes similarity based on the configured metric.
func (s *Store) calculateSimilarity(a, b []float32) float64 {
	switch s.metric {
	case embeddings.CosineSimilarity:
		return cosineSimilarity(a, b)
	case embeddings.EuclideanDistance:
		return 1.0 / (1.0 + euclideanDistance(a, b)) // Convert distance to similarity
	case embeddings.DotProduct:
		return float64(dotProduct(a, b))
	default:
		return cosineSimilarity(a, b)
	}
}

// cosineSimilarity calculates cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProd, normA, normB float64
	for i := range a {
		dotProd += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProd / (math.Sqrt(normA) * math.Sqrt(normB))
}

// euclideanDistance calculates Euclidean distance between two vectors.
func euclideanDistance(a, b []float32) float64 {
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

// dotProduct calculates dot product between two vectors.
func dotProduct(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}

	return sum
}

// matchesFilters checks if metadata matches all filter conditions.
func matchesFilters(metadata, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}

	for key, value := range filters {
		metaValue, ok := metadata[key]
		if !ok {
			return false
		}

		if metaValue != value {
			return false
		}
	}

	return true
}

// List returns all vectors (optionally filtered).
func (s *Store) List(filters map[string]interface{}) []*Vector {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Vector
	for _, vec := range s.vectors {
		if matchesFilters(vec.Metadata, filters) {
			results = append(results, vec)
		}
	}

	return results
}

// GetAll returns all vectors (used for HNSW reindexing).
func (s *Store) GetAll() []*Vector {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*Vector, 0, len(s.vectors))
	for _, vec := range s.vectors {
		results = append(results, vec)
	}

	return results
}

// GetByID returns a vector by ID (used for hybrid search metadata lookup).
func (s *Store) GetByID(id string) *Vector {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.vectors[id]
}
