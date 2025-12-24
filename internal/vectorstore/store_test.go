package vectorstore

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_AddAndGet(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 384)
	store := NewStore(provider)

	ctx := context.Background()

	// Add a vector
	err := store.Add(ctx, "test-1", "hello world", map[string]interface{}{
		"type": "test",
	})
	require.NoError(t, err)

	// Get the vector
	vec, err := store.Get("test-1")
	require.NoError(t, err)
	assert.Equal(t, "test-1", vec.ID)
	assert.Equal(t, "hello world", vec.Text)
	assert.Equal(t, 384, len(vec.Embedding))
}

func TestStore_AddBatch(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 384)
	store := NewStore(provider)

	ctx := context.Background()

	items := []struct {
		ID       string
		Text     string
		Metadata map[string]interface{}
	}{
		{"id-1", "text 1", map[string]interface{}{"type": "a"}},
		{"id-2", "text 2", map[string]interface{}{"type": "b"}},
		{"id-3", "text 3", map[string]interface{}{"type": "a"}},
	}

	err := store.AddBatch(ctx, items)
	require.NoError(t, err)

	assert.Equal(t, 3, store.Size())
}

func TestStore_Delete(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 384)
	store := NewStore(provider)

	ctx := context.Background()

	// Add and delete
	_ = store.Add(ctx, "test-1", "hello", nil)
	assert.Equal(t, 1, store.Size())

	err := store.Delete("test-1")
	require.NoError(t, err)
	assert.Equal(t, 0, store.Size())

	// Try to get deleted vector
	_, err = store.Get("test-1")
	assert.Error(t, err)
}

func TestStore_Search(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 384)
	store := NewStore(provider)

	ctx := context.Background()

	// Add some vectors
	items := []struct {
		ID       string
		Text     string
		Metadata map[string]interface{}
	}{
		{"id-1", "machine learning", map[string]interface{}{"category": "ai"}},
		{"id-2", "deep learning", map[string]interface{}{"category": "ai"}},
		{"id-3", "cooking recipes", map[string]interface{}{"category": "food"}},
	}

	for _, item := range items {
		_ = store.Add(ctx, item.ID, item.Text, item.Metadata)
	}

	// Search without filters
	results, err := store.Search(ctx, "artificial intelligence", 2, nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Search with filters
	results, err = store.Search(ctx, "learning", 10, map[string]interface{}{"category": "ai"})
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Verify results have expected fields
	for _, r := range results {
		assert.NotEmpty(t, r.ID)
		assert.NotEmpty(t, r.Text)
		assert.NotZero(t, r.Score)
	}
}

func TestStore_SearchByVector(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 384)
	store := NewStore(provider)

	ctx := context.Background()

	// Add vectors
	_ = store.Add(ctx, "id-1", "test 1", nil)
	_ = store.Add(ctx, "id-2", "test 2", nil)

	// Create a query embedding
	queryEmb, _ := provider.Embed(ctx, "query")

	// Search by vector
	results := store.SearchByVector(queryEmb, 2, nil)
	assert.Len(t, results, 2)
}

func TestStore_Clear(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 384)
	store := NewStore(provider)

	ctx := context.Background()

	// Add vectors
	_ = store.Add(ctx, "id-1", "test 1", nil)
	_ = store.Add(ctx, "id-2", "test 2", nil)
	assert.Equal(t, 2, store.Size())

	// Clear
	store.Clear()
	assert.Equal(t, 0, store.Size())
}

func TestStore_List(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 384)
	store := NewStore(provider)

	ctx := context.Background()

	// Add vectors with different metadata
	_ = store.Add(ctx, "id-1", "test 1", map[string]interface{}{"type": "a"})
	_ = store.Add(ctx, "id-2", "test 2", map[string]interface{}{"type": "b"})
	_ = store.Add(ctx, "id-3", "test 3", map[string]interface{}{"type": "a"})

	// List all
	all := store.List(nil)
	assert.Len(t, all, 3)

	// List with filter
	filtered := store.List(map[string]interface{}{"type": "a"})
	assert.Len(t, filtered, 2)
}

func TestStore_SimilarityMetrics(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 384)
	store := NewStore(provider)

	ctx := context.Background()

	_ = store.Add(ctx, "id-1", "test", nil)

	t.Run("cosine similarity", func(t *testing.T) {
		store.SetMetric(embeddings.CosineSimilarity)
		results, _ := store.Search(ctx, "test", 1, nil)
		assert.Len(t, results, 1)
		assert.Greater(t, results[0].Score, 0.0)
	})

	t.Run("euclidean distance", func(t *testing.T) {
		store.SetMetric(embeddings.EuclideanDistance)
		results, _ := store.Search(ctx, "test", 1, nil)
		assert.Len(t, results, 1)
	})

	t.Run("dot product", func(t *testing.T) {
		store.SetMetric(embeddings.DotProduct)
		results, _ := store.Search(ctx, "test", 1, nil)
		assert.Len(t, results, 1)
	})
}

func TestStore_ErrorCases(t *testing.T) {
	provider := embeddings.NewMockProvider("mock", 384)
	store := NewStore(provider)

	ctx := context.Background()

	t.Run("empty ID", func(t *testing.T) {
		err := store.Add(ctx, "", "text", nil)
		assert.Error(t, err)
	})

	t.Run("empty text", func(t *testing.T) {
		err := store.Add(ctx, "id", "", nil)
		assert.Error(t, err)
	})

	t.Run("empty query", func(t *testing.T) {
		_, err := store.Search(ctx, "", 10, nil)
		assert.Error(t, err)
	})

	t.Run("invalid k", func(t *testing.T) {
		_, err := store.Search(ctx, "query", 0, nil)
		assert.Error(t, err)
	})

	t.Run("get non-existent", func(t *testing.T) {
		_, err := store.Get("non-existent")
		assert.Error(t, err)
	})

	t.Run("delete non-existent", func(t *testing.T) {
		err := store.Delete("non-existent")
		assert.Error(t, err)
	})

	t.Run("empty batch", func(t *testing.T) {
		err := store.AddBatch(ctx, nil)
		assert.Error(t, err)
	})
}

func TestCosineSimilarity(t *testing.T) {
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{1.0, 0.0, 0.0}

	sim := cosineSimilarity(a, b)
	assert.InDelta(t, 1.0, sim, 0.01)

	c := []float32{0.0, 1.0, 0.0}
	sim = cosineSimilarity(a, c)
	assert.InDelta(t, 0.0, sim, 0.01)
}

func TestEuclideanDistance(t *testing.T) {
	a := []float32{0.0, 0.0}
	b := []float32{3.0, 4.0}

	dist := euclideanDistance(a, b)
	assert.InDelta(t, 5.0, dist, 0.01)
}

func TestDotProduct(t *testing.T) {
	a := []float32{1.0, 2.0, 3.0}
	b := []float32{4.0, 5.0, 6.0}

	dot := dotProduct(a, b)
	assert.Equal(t, float32(32.0), dot) // 1*4 + 2*5 + 3*6 = 32
}
