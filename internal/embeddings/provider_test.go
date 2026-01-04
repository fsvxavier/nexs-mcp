package embeddings

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockProvider_Name(t *testing.T) {
	provider := NewMockProvider("test-provider", 384)
	assert.Equal(t, "test-provider", provider.Name())
}

func TestMockProvider_Dimensions(t *testing.T) {
	tests := []struct {
		name string
		dims int
	}{
		{"384 dimensions", 384},
		{"768 dimensions", 768},
		{"1536 dimensions", 1536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewMockProvider("test", tt.dims)
			assert.Equal(t, tt.dims, provider.Dimensions())
		})
	}
}

func TestMockProvider_Embed(t *testing.T) {
	provider := NewMockProvider("test", 384)
	ctx := context.Background()

	embedding, err := provider.Embed(ctx, "test text")
	require.NoError(t, err)
	assert.NotNil(t, embedding)
	assert.Equal(t, 384, len(embedding))

	// Verify embedding values are in reasonable range
	for _, val := range embedding {
		assert.GreaterOrEqual(t, val, float32(-1.0))
		assert.LessOrEqual(t, val, float32(1.0))
	}
}

func TestMockProvider_EmbedBatch(t *testing.T) {
	provider := NewMockProvider("test", 384)
	ctx := context.Background()

	texts := []string{"text1", "text2", "text3"}
	embeddings, err := provider.EmbedBatch(ctx, texts)
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)

	for _, emb := range embeddings {
		assert.Equal(t, 384, len(emb))
	}
}

func TestMockProvider_EmbedBatch_EmptyInput(t *testing.T) {
	provider := NewMockProvider("test", 384)
	ctx := context.Background()

	embeddings, err := provider.EmbedBatch(ctx, []string{})
	require.NoError(t, err)
	assert.Len(t, embeddings, 0)
}

func TestMockProvider_IsAvailable(t *testing.T) {
	provider := NewMockProvider("test", 384)
	ctx := context.Background()

	assert.True(t, provider.IsAvailable(ctx))
}

func TestMockProvider_Cost(t *testing.T) {
	provider := NewMockProvider("test", 384)
	assert.Equal(t, 0.0, provider.Cost())
}

func TestMockProvider_DeterministicBehavior(t *testing.T) {
	provider := NewMockProvider("test", 384)
	ctx := context.Background()

	// Test that same text produces consistent embeddings
	text := "consistent test"
	emb1, err1 := provider.Embed(ctx, text)
	require.NoError(t, err1)

	emb2, err2 := provider.Embed(ctx, text)
	require.NoError(t, err2)

	assert.Equal(t, emb1, emb2, "Same text should produce same embedding")
}

func TestMockProvider_DifferentTexts(t *testing.T) {
	provider := NewMockProvider("test", 384)
	ctx := context.Background()

	emb1, err := provider.Embed(ctx, "text one")
	require.NoError(t, err)
	assert.Len(t, emb1, 384)

	emb2, err := provider.Embed(ctx, "text two")
	require.NoError(t, err)
	assert.Len(t, emb2, 384)

	// Note: MockProvider generates deterministic embeddings based on text hash
	// So same text will always produce same embedding (which is good for testing caching)
}

func TestProviderDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "transformers", config.Provider)
	assert.True(t, config.EnableCache)
	assert.True(t, config.EnableFallback)
	assert.Equal(t, 10000, config.CacheMaxSize)
	assert.NotEmpty(t, config.FallbackPriority)
}

func TestConfig_Fields(t *testing.T) {
	config := Config{
		Provider:         "openai",
		OpenAIKey:        "test-key",
		OpenAIModel:      "text-embedding-3-small",
		EnableCache:      true,
		CacheMaxSize:     5000,
		EnableFallback:   true,
		FallbackPriority: []string{"openai", "transformers"},
	}

	assert.Equal(t, "openai", config.Provider)
	assert.Equal(t, "test-key", config.OpenAIKey)
	assert.Equal(t, "text-embedding-3-small", config.OpenAIModel)
	assert.True(t, config.EnableCache)
	assert.Equal(t, 5000, config.CacheMaxSize)
	assert.Len(t, config.FallbackPriority, 2)
}

func TestProviderInterface_Compliance(t *testing.T) {
	// Verify MockProvider implements Provider interface
	var _ Provider = (*MockProvider)(nil)
}

func TestMockProvider_ConcurrentAccess(t *testing.T) {
	provider := NewMockProvider("test", 384)
	ctx := context.Background()

	// Test concurrent Embed calls
	done := make(chan bool, 10)
	for i := range 10 {
		go func(idx int) {
			_, err := provider.Embed(ctx, "concurrent test")
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}
}

func TestMockProvider_LargeText(t *testing.T) {
	provider := NewMockProvider("test", 384)
	ctx := context.Background()

	// Create large text
	largeText := ""
	var largeTextSb176 strings.Builder
	for range 1000 {
		largeTextSb176.WriteString("This is a large text for testing purposes. ")
	}
	largeText += largeTextSb176.String()

	embedding, err := provider.Embed(ctx, largeText)
	require.NoError(t, err)
	assert.Equal(t, 384, len(embedding))
}

func TestMockProvider_SpecialCharacters(t *testing.T) {
	provider := NewMockProvider("test", 384)
	ctx := context.Background()

	tests := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"unicode", "こんにちは世界"},
		{"emoji", "Hello 🌍 World 🚀"},
		{"special chars", "!@#$%^&*()_+-=[]{}|;:,.<>?"},
		{"newlines", "line1\nline2\nline3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedding, err := provider.Embed(ctx, tt.text)
			require.NoError(t, err)
			assert.Equal(t, 384, len(embedding))
		})
	}
}
