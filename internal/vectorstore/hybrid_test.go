package vectorstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultHybridConfig(t *testing.T) {
	config := DefaultHybridConfig(384)
	require.NotNil(t, config)
	assert.Equal(t, 384, config.Dimension)
	assert.Equal(t, 100, config.SwitchThreshold)
	assert.Equal(t, SimilarityCosine, config.Similarity)
	assert.NotNil(t, config.HNSWConfig)
}

func TestNewHybridStore(t *testing.T) {
	config := DefaultHybridConfig(384)
	store := NewHybridStore(config)

	require.NotNil(t, store)
	assert.NotNil(t, store.linearVectors)
	assert.False(t, store.useHNSW)
	assert.Equal(t, config, store.config)
}

func TestHybridStore_AddGetLinear(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Add a vector
	vector := []float32{1.0, 2.0, 3.0}
	err := store.Add("test-1", vector, map[string]interface{}{"key": "value"})
	require.NoError(t, err)

	// Get the vector
	entry, err := store.Get("test-1")
	require.NoError(t, err)
	assert.Equal(t, "test-1", entry.ID)
	assert.Equal(t, vector, entry.Vector)
	assert.Equal(t, "value", entry.Metadata["key"])
}

func TestHybridStore_Delete(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Add a vector
	vector := []float32{1.0, 2.0, 3.0}
	err := store.Add("test-1", vector, nil)
	require.NoError(t, err)

	// Delete it
	err = store.Delete("test-1")
	require.NoError(t, err)

	// Should not be found
	_, err = store.Get("test-1")
	assert.ErrorIs(t, err, ErrVectorNotFound)
}

func TestHybridStore_SearchLinear(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Add some vectors
	vectors := []struct {
		id     string
		vector []float32
	}{
		{"v1", []float32{1.0, 0.0, 0.0}},
		{"v2", []float32{0.0, 1.0, 0.0}},
		{"v3", []float32{0.0, 0.0, 1.0}},
	}

	for _, v := range vectors {
		err := store.Add(v.id, v.vector, nil)
		require.NoError(t, err)
	}

	// Search for vector similar to v1
	query := []float32{0.9, 0.1, 0.0}
	results, err := store.Search(query, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	// The closest should be v1
	assert.Equal(t, "v1", results[0].ID)
}

func TestHybridStore_Size(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Initially empty
	assert.Equal(t, 0, store.Size())

	// Add vectors
	for i := 0; i < 5; i++ {
		vector := []float32{float32(i), float32(i + 1), float32(i + 2)}
		err := store.Add(string(rune('a'+i)), vector, nil)
		require.NoError(t, err)
	}

	// Should have 5 vectors
	assert.Equal(t, 5, store.Size())
}

func TestHybridStore_Clear(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Add vectors
	for i := 0; i < 5; i++ {
		vector := []float32{float32(i), float32(i + 1), float32(i + 2)}
		err := store.Add(string(rune('a'+i)), vector, nil)
		require.NoError(t, err)
	}

	// Clear
	store.Clear()

	// Should be empty
	assert.Equal(t, 0, store.Size())
	assert.False(t, store.useHNSW)
}

func TestHybridStore_DimensionMismatch(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Try to add vector with wrong dimension
	wrongVector := []float32{1.0, 2.0, 3.0, 4.0} // 4D instead of 3D
	err := store.Add("test", wrongVector, nil)
	assert.ErrorIs(t, err, ErrDimensionMismatch)
}

func TestHybridStore_EmptyID(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Add with empty ID should work (no validation in current implementation)
	vector := []float32{1.0, 2.0, 3.0}
	err := store.Add("", vector, nil)
	assert.NoError(t, err)
}

func TestHybridStore_VectorExists(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Add a vector
	vector := []float32{1.0, 2.0, 3.0}
	err := store.Add("test", vector, nil)
	require.NoError(t, err)

	// Try to add again with same ID
	err = store.Add("test", vector, nil)
	assert.ErrorIs(t, err, ErrVectorExists)
}

func TestHybridStore_SearchEmptyStore(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Search in empty store
	query := []float32{1.0, 2.0, 3.0}
	results, err := store.Search(query, 5)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestHybridStore_SearchWithMetadata(t *testing.T) {
	store := NewHybridStore(DefaultHybridConfig(3))

	// Add vector with metadata
	vector := []float32{1.0, 2.0, 3.0}
	metadata := map[string]interface{}{
		"category": "test",
		"score":    0.95,
	}
	err := store.Add("test-1", vector, metadata)
	require.NoError(t, err)

	// Search and verify metadata is preserved
	results, err := store.Search(vector, 1)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test-1", results[0].ID)
	assert.Equal(t, metadata, results[0].Metadata)
}

func TestHybridStore_SwitchToHNSW(t *testing.T) {
	// Create store with low threshold for testing
	config := DefaultHybridConfig(3)
	config.SwitchThreshold = 10
	store := NewHybridStore(config)

	// Add vectors below threshold
	for i := 0; i < 9; i++ {
		vector := []float32{float32(i), float32(i + 1), float32(i + 2)}
		err := store.Add(string(rune('a'+i)), vector, nil)
		require.NoError(t, err)
	}

	// Should still be in linear mode
	assert.False(t, store.useHNSW)

	// Add one more to trigger switch
	vector := []float32{9.0, 10.0, 11.0}
	err := store.Add("trigger", vector, nil)
	require.NoError(t, err)

	// Should now be in HNSW mode
	assert.True(t, store.useHNSW)

	// Verify we can still search
	results, err := store.Search(vector, 2)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}
