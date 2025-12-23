package hnsw

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRandomVector(dim int) []float32 {
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float32()
	}
	return vec
}

func TestNewGraph(t *testing.T) {
	graph := NewGraph(CosineSimilarity)
	assert.NotNil(t, graph)
	assert.Equal(t, 0, graph.Size())
}

func TestGraphInsert(t *testing.T) {
	graph := NewGraph(CosineSimilarity)
	vec := generateRandomVector(128)
	err := graph.Insert("node1", vec)
	require.NoError(t, err)
	assert.Equal(t, 1, graph.Size())
}

func TestGraphSearch(t *testing.T) {
	graph := NewGraph(CosineSimilarity)

	// Insert 50 vectors
	for i := 0; i < 50; i++ {
		vec := generateRandomVector(128)
		err := graph.Insert(fmt.Sprintf("node%d", i), vec)
		require.NoError(t, err)
	}

	query := generateRandomVector(128)
	results, err := graph.SearchKNN(query, 5)
	require.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestGraphDelete(t *testing.T) {
	graph := NewGraph(CosineSimilarity)

	vec := generateRandomVector(128)
	err := graph.Insert("node1", vec)
	require.NoError(t, err)

	err = graph.Delete("node1")
	require.NoError(t, err)
	assert.Equal(t, 0, graph.Size())
}

func TestGraphStatistics(t *testing.T) {
	graph := NewGraph(CosineSimilarity)

	for i := 0; i < 100; i++ {
		vec := generateRandomVector(128)
		err := graph.Insert(fmt.Sprintf("node%d", i), vec)
		require.NoError(t, err)
	}

	stats := graph.GetStatistics()
	assert.Equal(t, 100, stats.NodeCount)
	assert.GreaterOrEqual(t, stats.MaxLevel, 0)
}

func BenchmarkInsert(b *testing.B) {
	graph := NewGraph(CosineSimilarity)
	for i := 0; i < b.N; i++ {
		vec := generateRandomVector(384)
		_ = graph.Insert(fmt.Sprintf("node%d", i), vec)
	}
}

func BenchmarkSearch(b *testing.B) {
	graph := NewGraph(CosineSimilarity)
	for i := 0; i < 1000; i++ {
		vec := generateRandomVector(384)
		_ = graph.Insert(fmt.Sprintf("node%d", i), vec)
	}

	query := generateRandomVector(384)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = graph.SearchKNN(query, 10)
	}
}
