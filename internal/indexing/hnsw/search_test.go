package hnsw

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchKNN(t *testing.T) {
	graph := NewGraph(CosineSimilarity)

	for i := range 50 {
		vec := generateRandomVector(128)
		err := graph.Insert(fmt.Sprintf("node%d", i), vec)
		require.NoError(t, err)
	}

	query := generateRandomVector(128)
	results, err := graph.SearchKNN(query, 10)

	require.NoError(t, err)
	assert.Len(t, results, 10)

	for _, result := range results {
		assert.GreaterOrEqual(t, result.Distance, float32(0.0))
		assert.NotEmpty(t, result.ID)
	}
}

func TestSearchKNNEmptyGraph(t *testing.T) {
	graph := NewGraph(CosineSimilarity)

	query := generateRandomVector(128)
	results, err := graph.SearchKNN(query, 5)

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestSearchWithCustomEfSearch(t *testing.T) {
	graph := NewGraph(CosineSimilarity)

	for i := range 100 {
		vec := generateRandomVector(128)
		err := graph.Insert(fmt.Sprintf("node%d", i), vec)
		require.NoError(t, err)
	}

	query := generateRandomVector(128)

	for _, ef := range []int{10, 50, 100, 200} {
		results, err := graph.Search(query, 10, ef)
		require.NoError(t, err)
		assert.Len(t, results, 10)
	}
}

func TestRangeSearch(t *testing.T) {
	graph := NewGraph(CosineSimilarity)

	for i := range 50 {
		vec := generateRandomVector(128)
		err := graph.Insert(fmt.Sprintf("node%d", i), vec)
		require.NoError(t, err)
	}

	query := generateRandomVector(128)
	maxDistance := float32(0.5)

	results, err := graph.RangeSearch(query, maxDistance, 100)
	require.NoError(t, err)

	for _, result := range results {
		assert.LessOrEqual(t, result.Distance, maxDistance)
	}
}

func TestBatchSearch(t *testing.T) {
	graph := NewGraph(CosineSimilarity)

	for i := range 100 {
		vec := generateRandomVector(384)
		err := graph.Insert(fmt.Sprintf("node%d", i), vec)
		require.NoError(t, err)
	}

	queries := make([][]float32, 5)
	for i := range 5 {
		queries[i] = generateRandomVector(384)
	}

	results, err := graph.BatchSearch(queries, 10, 50)
	require.NoError(t, err)
	assert.Len(t, results, 5)

	for _, resultSet := range results {
		assert.Len(t, resultSet, 10)
	}
}

func TestDeleteAndSearch(t *testing.T) {
	graph := NewGraph(CosineSimilarity)

	for i := range 20 {
		vec := generateRandomVector(128)
		err := graph.Insert(fmt.Sprintf("node%d", i), vec)
		require.NoError(t, err)
	}

	err := graph.Delete("node5")
	require.NoError(t, err)

	query := generateRandomVector(128)
	results, err := graph.SearchKNN(query, 20)
	require.NoError(t, err)

	for _, result := range results {
		assert.NotEqual(t, "node5", result.ID)
	}
}

func BenchmarkSearchKNN(b *testing.B) {
	graph := NewGraph(CosineSimilarity)

	for i := range 1000 {
		vec := generateRandomVector(384)
		_ = graph.Insert(fmt.Sprintf("node%d", i), vec)
	}

	query := generateRandomVector(384)

	b.ResetTimer()
	for range b.N {
		_, _ = graph.SearchKNN(query, 10)
	}
}
