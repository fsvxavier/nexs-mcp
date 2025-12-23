package hnsw

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphSaveLoad(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test_index.json")

	// Create and populate graph
	graph1 := NewGraph(CosineSimilarity)
	graph1.SetParameters(8, 100)

	for i := 0; i < 50; i++ {
		vec := generateRandomVector(128)
		err := graph1.Insert(string(rune('A'+i)), vec)
		require.NoError(t, err)
	}

	// Save graph
	err := graph1.Save(indexPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(indexPath)
	require.NoError(t, err)

	// Load into new graph
	graph2 := NewGraph(CosineSimilarity)
	err = graph2.Load(indexPath)
	require.NoError(t, err)

	// Verify parameters
	assert.Equal(t, graph1.m, graph2.m)
	assert.Equal(t, graph1.efConstruction, graph2.efConstruction)
	assert.Equal(t, graph1.maxLevel, graph2.maxLevel)
	assert.Equal(t, graph1.Size(), graph2.Size())

	// Verify search results match
	query := generateRandomVector(128)
	results1, err := graph1.SearchKNN(query, 5)
	require.NoError(t, err)

	results2, err := graph2.SearchKNN(query, 5)
	require.NoError(t, err)

	// Results should be identical
	assert.Equal(t, len(results1), len(results2))
	for i := range results1 {
		assert.Equal(t, results1[i].ID, results2[i].ID)
	}
}

func TestGraphSaveEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "empty_index.json")

	graph := NewGraph(CosineSimilarity)
	err := graph.Save(indexPath)
	require.NoError(t, err)

	// Load empty graph
	graph2 := NewGraph(CosineSimilarity)
	err = graph2.Load(indexPath)
	require.NoError(t, err)
	assert.Equal(t, 0, graph2.Size())
}

func TestGraphLoadNonExistent(t *testing.T) {
	graph := NewGraph(CosineSimilarity)
	err := graph.Load("/nonexistent/path/index.json")
	assert.Error(t, err)
}
