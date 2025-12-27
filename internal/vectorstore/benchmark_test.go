package vectorstore

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// Benchmark helpers.
func generateBenchmarkVectors(n, dimension int) [][]float32 {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	vectors := make([][]float32, n)
	for i := range n {
		vec := make([]float32, dimension)
		for j := range dimension {
			vec[j] = rng.Float32()*2 - 1
		}
		vectors[i] = vec
	}
	return vectors
}

func populateHybridStore(store *HybridStore, n, dimension int) {
	vectors := generateBenchmarkVectors(n, dimension)
	for i := range n {
		id := fmt.Sprintf("vec_%d", i)
		_ = store.Add(id, vectors[i], nil)
	}
}

func populateHNSWIndex(index *HNSWIndex, n, dimension int) {
	vectors := generateBenchmarkVectors(n, dimension)
	for i := range n {
		id := fmt.Sprintf("vec_%d", i)
		_ = index.Add(id, vectors[i], nil)
	}
}

// Linear Search Benchmarks (using HybridStore below threshold).
func BenchmarkLinearSearch_1k(b *testing.B) {
	config := &HybridConfig{
		Dimension:       384,
		Similarity:      SimilarityCosine,
		SwitchThreshold: 10000, // Keep in linear mode
		HNSWConfig:      DefaultHNSWConfig(),
	}
	store := NewHybridStore(config)
	populateHybridStore(store, 1000, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = store.Search(query, 10)
	}
}

func BenchmarkLinearSearch_10k(b *testing.B) {
	config := &HybridConfig{
		Dimension:       384,
		Similarity:      SimilarityCosine,
		SwitchThreshold: 20000, // Keep in linear mode
		HNSWConfig:      DefaultHNSWConfig(),
	}
	store := NewHybridStore(config)
	populateHybridStore(store, 10000, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = store.Search(query, 10)
	}
}

// HNSW Search Benchmarks (direct HNSW index).
func BenchmarkHNSWSearch_1k(b *testing.B) {
	config := DefaultHNSWConfig()
	index, err := NewHNSWIndex(384, SimilarityCosine, config)
	if err != nil {
		b.Fatalf("Failed to create HNSW index: %v", err)
	}
	populateHNSWIndex(index, 1000, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = index.Search(query, 10)
	}
}

func BenchmarkHNSWSearch_10k(b *testing.B) {
	config := DefaultHNSWConfig()
	index, err := NewHNSWIndex(384, SimilarityCosine, config)
	if err != nil {
		b.Fatalf("Failed to create HNSW index: %v", err)
	}
	populateHNSWIndex(index, 10000, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = index.Search(query, 10)
	}
}

func BenchmarkHNSWSearch_100k(b *testing.B) {
	config := DefaultHNSWConfig()
	index, err := NewHNSWIndex(384, SimilarityCosine, config)
	if err != nil {
		b.Fatalf("Failed to create HNSW index: %v", err)
	}
	populateHNSWIndex(index, 100000, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = index.Search(query, 10)
	}
}

// HNSW with different EfSearch values.
func BenchmarkHNSWSearch_10k_Ef10(b *testing.B) {
	config := &HNSWConfig{
		M:        16,
		Ml:       0.25,
		EfSearch: 10,
		Seed:     42,
	}
	index, err := NewHNSWIndex(384, SimilarityCosine, config)
	if err != nil {
		b.Fatalf("Failed to create HNSW index: %v", err)
	}
	populateHNSWIndex(index, 10000, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = index.Search(query, 10)
	}
}

func BenchmarkHNSWSearch_10k_Ef50(b *testing.B) {
	config := &HNSWConfig{
		M:        16,
		Ml:       0.25,
		EfSearch: 50,
		Seed:     42,
	}
	index, err := NewHNSWIndex(384, SimilarityCosine, config)
	if err != nil {
		b.Fatalf("Failed to create HNSW index: %v", err)
	}
	populateHNSWIndex(index, 10000, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = index.Search(query, 10)
	}
}

func BenchmarkHNSWSearch_10k_Ef100(b *testing.B) {
	config := &HNSWConfig{
		M:        16,
		Ml:       0.25,
		EfSearch: 100,
		Seed:     42,
	}
	index, err := NewHNSWIndex(384, SimilarityCosine, config)
	if err != nil {
		b.Fatalf("Failed to create HNSW index: %v", err)
	}
	populateHNSWIndex(index, 10000, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = index.Search(query, 10)
	}
}

// Hybrid Store Benchmarks.
func BenchmarkHybridStore_Below_Threshold(b *testing.B) {
	config := &HybridConfig{
		Dimension:       384,
		Similarity:      SimilarityCosine,
		SwitchThreshold: 100,
		HNSWConfig:      DefaultHNSWConfig(),
	}
	store := NewHybridStore(config)

	// Add 50 vectors (below threshold)
	populateHybridStore(store, 50, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = store.Search(query, 10)
	}
}

func BenchmarkHybridStore_Above_Threshold(b *testing.B) {
	config := &HybridConfig{
		Dimension:       384,
		Similarity:      SimilarityCosine,
		SwitchThreshold: 100,
		HNSWConfig:      DefaultHNSWConfig(),
	}
	store := NewHybridStore(config)

	// Add 10k vectors (above threshold)
	populateHybridStore(store, 10000, 384)
	query := randomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = store.Search(query, 10)
	}
}

// Add operation benchmarks.
func BenchmarkHNSWAdd_Sequential(b *testing.B) {
	config := DefaultHNSWConfig()
	index, err := NewHNSWIndex(384, SimilarityCosine, config)
	if err != nil {
		b.Fatalf("Failed to create HNSW index: %v", err)
	}

	vectors := generateBenchmarkVectors(b.N, 384)

	b.ResetTimer()
	for i := range b.N {
		id := fmt.Sprintf("vec_%d", i)
		_ = index.Add(id, vectors[i], nil)
	}
}

func BenchmarkHybridAdd_Sequential(b *testing.B) {
	config := &HybridConfig{
		Dimension:       384,
		Similarity:      SimilarityCosine,
		SwitchThreshold: 100,
		HNSWConfig:      DefaultHNSWConfig(),
	}
	store := NewHybridStore(config)
	vectors := generateBenchmarkVectors(b.N, 384)

	b.ResetTimer()
	for i := range b.N {
		id := fmt.Sprintf("vec_%d", i)
		_ = store.Add(id, vectors[i], nil)
	}
}

// Memory benchmarks.
func BenchmarkHNSWMemory_10k(b *testing.B) {
	config := DefaultHNSWConfig()

	b.ReportAllocs()
	for range b.N {
		index, _ := NewHNSWIndex(384, SimilarityCosine, config)
		populateHNSWIndex(index, 10000, 384)
	}
}

func BenchmarkHybridMemory_10k(b *testing.B) {
	config := &HybridConfig{
		Dimension:       384,
		Similarity:      SimilarityCosine,
		SwitchThreshold: 100,
		HNSWConfig:      DefaultHNSWConfig(),
	}

	b.ReportAllocs()
	for range b.N {
		store := NewHybridStore(config)
		populateHybridStore(store, 10000, 384)
	}
}
