package vectorstore

import (
	"math"
	"math/rand"
	"testing"
)

// TestHNSWIndex_BasicOperations tests basic CRUD operations.
func TestHNSWIndex_BasicOperations(t *testing.T) {
	config := DefaultHNSWConfig()
	index, err := NewHNSWIndex(3, SimilarityCosine, config)
	if err != nil {
		t.Fatalf("NewHNSWIndex failed: %v", err)
	}

	// Test Add
	vec1 := []float32{1.0, 0.0, 0.0}
	err = index.Add("v1", vec1, map[string]interface{}{"label": "first"})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Test Size
	if index.Size() != 1 {
		t.Errorf("Expected size 1, got %d", index.Size())
	}

	// Test Get
	entry, exists := index.Get("v1")
	if !exists {
		t.Error("Vector v1 should exist")
	}
	if entry.ID != "v1" {
		t.Errorf("Expected ID v1, got %s", entry.ID)
	}

	// Test Delete
	err = index.Delete("v1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if index.Size() != 0 {
		t.Errorf("Expected size 0 after delete, got %d", index.Size())
	}
}

// TestHNSWIndex_DimensionMismatch tests dimension validation.
func TestHNSWIndex_DimensionMismatch(t *testing.T) {
	index, err := NewHNSWIndex(3, SimilarityCosine, nil)
	if err != nil {
		t.Fatalf("NewHNSWIndex failed: %v", err)
	}

	err = index.Add("v1", []float32{1.0, 2.0}, nil) // Wrong dimension
	if err != ErrDimensionMismatch {
		t.Errorf("Expected ErrDimensionMismatch, got %v", err)
	}
}

// TestHNSWIndex_Search tests basic search functionality.
func TestHNSWIndex_Search(t *testing.T) {
	index, err := NewHNSWIndex(3, SimilarityCosine, nil)
	if err != nil {
		t.Fatalf("NewHNSWIndex failed: %v", err)
	}

	// Add some vectors
	vectors := []struct {
		id  string
		vec []float32
	}{
		{"v1", []float32{1.0, 0.0, 0.0}},
		{"v2", []float32{0.0, 1.0, 0.0}},
		{"v3", []float32{0.0, 0.0, 1.0}},
		{"v4", []float32{0.7, 0.7, 0.0}}, // Similar to v1 and v2
	}

	for _, v := range vectors {
		err := index.Add(v.id, v.vec, nil)
		if err != nil {
			t.Fatalf("Add %s failed: %v", v.id, err)
		}
	}

	// Search for v1
	query := []float32{1.0, 0.0, 0.0}
	results, err := index.Search(query, 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// First result should be v1 (exact match)
	if results[0].ID != "v1" {
		t.Errorf("Expected first result to be v1, got %s", results[0].ID)
	}

	// Similarity should be close to 1.0
	if results[0].Similarity < 0.99 {
		t.Errorf("Expected similarity close to 1.0, got %f", results[0].Similarity)
	}
}

// TestHNSWIndex_SearchEmptyIndex tests search on empty index.
func TestHNSWIndex_SearchEmptyIndex(t *testing.T) {
	index, err := NewHNSWIndex(3, SimilarityCosine, nil)
	if err != nil {
		t.Fatalf("NewHNSWIndex failed: %v", err)
	}

	query := []float32{1.0, 0.0, 0.0}
	results, err := index.Search(query, 5)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results on empty index, got %d", len(results))
	}
}

// TestHNSWIndex_Clear tests clearing the index.
func TestHNSWIndex_Clear(t *testing.T) {
	index, err := NewHNSWIndex(3, SimilarityCosine, nil)
	if err != nil {
		t.Fatalf("NewHNSWIndex failed: %v", err)
	}

	// Add vectors
	for i := 0; i < 10; i++ {
		vec := randomVector(3)
		id := string(rune('a' + i))
		err := index.Add(id, vec, nil)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	if index.Size() != 10 {
		t.Errorf("Expected size 10, got %d", index.Size())
	}

	// Clear
	index.Clear()

	if index.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", index.Size())
	}
}

// TestHybridStore_ThresholdSwitching tests automatic switching from linear to HNSW.
func TestHybridStore_ThresholdSwitching(t *testing.T) {
	config := &HybridConfig{
		SwitchThreshold: 10, // Low threshold for testing
		HNSWConfig:      DefaultHNSWConfig(),
		Similarity:      SimilarityCosine,
		Dimension:       3,
	}

	store := NewHybridStore(config)

	// Add vectors below threshold
	for i := 0; i < 9; i++ {
		vec := randomVector(3)
		id := string(rune('a' + i))
		err := store.Add(id, vec, nil)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Should still be using linear
	if store.IsUsingHNSW() {
		t.Error("Store should be using linear mode below threshold")
	}

	// Add one more to hit threshold
	vec := randomVector(3)
	err := store.Add("z", vec, nil)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Should now be using HNSW
	if !store.IsUsingHNSW() {
		t.Error("Store should switch to HNSW mode at threshold")
	}

	// Verify all vectors were migrated
	if store.Size() != 10 {
		t.Errorf("Expected size 10 after migration, got %d", store.Size())
	}

	// Verify search works
	query := randomVector(3)
	results, err := store.Search(query, 5)
	if err != nil {
		t.Fatalf("Search after migration failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

// TestHybridStore_LinearSearch tests linear search mode.
func TestHybridStore_LinearSearch(t *testing.T) {
	config := &HybridConfig{
		SwitchThreshold: 100, // High threshold to stay in linear mode
		Similarity:      SimilarityCosine,
		Dimension:       3,
	}

	store := NewHybridStore(config)

	// Add vectors
	vectors := []struct {
		id  string
		vec []float32
	}{
		{"v1", []float32{1.0, 0.0, 0.0}},
		{"v2", []float32{0.0, 1.0, 0.0}},
		{"v3", []float32{0.0, 0.0, 1.0}},
	}

	for _, v := range vectors {
		err := store.Add(v.id, v.vec, nil)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Search
	query := []float32{1.0, 0.0, 0.0}
	results, err := store.Search(query, 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// First result should be v1
	if results[0].ID != "v1" {
		t.Errorf("Expected first result to be v1, got %s", results[0].ID)
	}
}

// TestHybridStore_CRUD tests CRUD operations in hybrid mode.
func TestHybridStore_CRUD(t *testing.T) {
	config := &HybridConfig{
		SwitchThreshold: 5,
		HNSWConfig:      DefaultHNSWConfig(),
		Similarity:      SimilarityCosine,
		Dimension:       3,
	}

	store := NewHybridStore(config)

	// Add
	vec := []float32{1.0, 0.0, 0.0}
	err := store.Add("v1", vec, map[string]interface{}{"label": "test"})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Get
	entry, err := store.Get("v1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if entry.ID != "v1" {
		t.Errorf("Expected ID v1, got %s", entry.ID)
	}

	// Delete
	err = store.Delete("v1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Get after delete
	_, err = store.Get("v1")
	if err != ErrVectorNotFound {
		t.Errorf("Expected ErrVectorNotFound, got %v", err)
	}
}

// TestHybridStore_Clear tests clearing the store.
func TestHybridStore_Clear(t *testing.T) {
	config := DefaultHybridConfig(3)
	store := NewHybridStore(config)

	// Add vectors
	for i := 0; i < 20; i++ {
		vec := randomVector(3)
		id := string(rune('a' + i))
		err := store.Add(id, vec, nil)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Clear
	store.Clear()

	if store.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", store.Size())
	}

	if store.IsUsingHNSW() {
		t.Error("Store should not be using HNSW after clear")
	}
}

// TestSimilarityMetrics tests different similarity metrics.
func TestSimilarityMetrics(t *testing.T) {
	vec1 := []float32{1.0, 0.0, 0.0}
	vec2 := []float32{1.0, 0.0, 0.0}
	vec3 := []float32{0.0, 1.0, 0.0}

	// Cosine similarity
	cos12 := CosineSimilarity(vec1, vec2)
	if math.Abs(float64(cos12)-1.0) > 0.001 {
		t.Errorf("Cosine similarity of identical vectors should be 1.0, got %f", cos12)
	}

	cos13 := CosineSimilarity(vec1, vec3)
	if math.Abs(float64(cos13)-0.0) > 0.001 {
		t.Errorf("Cosine similarity of orthogonal vectors should be 0.0, got %f", cos13)
	}

	// Euclidean distance
	eucl12 := EuclideanDistance(vec1, vec2)
	if eucl12 > 0.001 {
		t.Errorf("Euclidean distance of identical vectors should be 0.0, got %f", eucl12)
	}

	eucl13 := EuclideanDistance(vec1, vec3)
	expected := math.Sqrt(2.0)
	if math.Abs(eucl13-expected) > 0.001 {
		t.Errorf("Euclidean distance should be %.3f, got %.3f", expected, eucl13)
	}

	// Dot product
	dot12 := DotProduct(vec1, vec2)
	if math.Abs(dot12-1.0) > 0.001 {
		t.Errorf("Dot product should be 1.0, got %f", dot12)
	}

	dot13 := DotProduct(vec1, vec3)
	if math.Abs(dot13-0.0) > 0.001 {
		t.Errorf("Dot product should be 0.0, got %f", dot13)
	}
}

// --- Helper functions ---

// randomVector generates a random normalized vector of given dimension.
func randomVector(dimension int) []float32 {
	vec := make([]float32, dimension)
	var norm float32

	for i := 0; i < dimension; i++ {
		vec[i] = rand.Float32()*2 - 1 // [-1, 1]
		norm += vec[i] * vec[i]
	}

	// Normalize
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}

	return vec
}
