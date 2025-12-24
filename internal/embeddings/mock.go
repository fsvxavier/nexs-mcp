package embeddings

import (
	"context"
)

// MockProvider is a mock embedding provider for testing.
type MockProvider struct {
	name       string
	dimensions int
	cost       float64
	available  bool
	embeddings map[string][]float32
}

// NewMockProvider creates a new mock provider.
func NewMockProvider(name string, dims int) *MockProvider {
	return &MockProvider{
		name:       name,
		dimensions: dims,
		cost:       0.0,
		available:  true,
		embeddings: make(map[string][]float32),
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Dimensions() int {
	return m.dimensions
}

func (m *MockProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// Return mock embedding
	embedding := make([]float32, m.dimensions)
	for i := range embedding {
		embedding[i] = float32(i) / float32(m.dimensions)
	}
	return embedding, nil
}

func (m *MockProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i := range texts {
		emb, err := m.Embed(ctx, texts[i])
		if err != nil {
			return nil, err
		}
		results[i] = emb
	}
	return results, nil
}

func (m *MockProvider) IsAvailable(ctx context.Context) bool {
	return m.available
}

func (m *MockProvider) Cost() float64 {
	return m.cost
}
