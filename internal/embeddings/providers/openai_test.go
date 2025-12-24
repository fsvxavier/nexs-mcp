package providers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAI(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey: "test-key",
			Model:  "text-embedding-3-small",
		}

		provider, err := NewOpenAI(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "openai", provider.Name())
		assert.Equal(t, 1536, provider.Dimensions())
		assert.Equal(t, 0.00002, provider.Cost())
	})

	t.Run("missing API key", func(t *testing.T) {
		config := OpenAIConfig{
			Model: "text-embedding-3-small",
		}

		_, err := NewOpenAI(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key")
	})

	t.Run("default model", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey: "test-key",
		}

		provider, err := NewOpenAI(config)
		require.NoError(t, err)
		assert.Equal(t, 1536, provider.Dimensions())
	})

	t.Run("large model", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey: "test-key",
			Model:  "text-embedding-3-large",
		}

		provider, err := NewOpenAI(config)
		require.NoError(t, err)
		assert.Equal(t, 3072, provider.Dimensions())
		assert.Equal(t, 0.00013, provider.Cost())
	})

	t.Run("ada model", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey: "test-key",
			Model:  "text-embedding-ada-002",
		}

		provider, err := NewOpenAI(config)
		require.NoError(t, err)
		assert.Equal(t, 1536, provider.Dimensions())
		assert.Equal(t, 0.0001, provider.Cost())
	})
}

func TestGetModelSpecs(t *testing.T) {
	tests := []struct {
		name  string
		model string
		dims  int
		cost  float64
	}{
		{"small model", "text-embedding-3-small", 1536, 0.00002},
		{"large model", "text-embedding-3-large", 3072, 0.00013},
		{"ada model", "text-embedding-ada-002", 1536, 0.0001},
		{"unknown model", "unknown-model", 1536, 0.00002}, // defaults to small
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, cost := getModelSpecs(tt.model)
			assert.Equal(t, tt.dims, dims)
			assert.Equal(t, tt.cost, cost)
		})
	}
}

func TestOpenAIProvider_Interface(t *testing.T) {
	config := OpenAIConfig{
		APIKey: "test-key",
		Model:  "text-embedding-3-small",
	}

	provider, _ := NewOpenAI(config)

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "openai", provider.Name())
	})

	t.Run("Dimensions", func(t *testing.T) {
		assert.Equal(t, 1536, provider.Dimensions())
	})

	t.Run("Cost", func(t *testing.T) {
		assert.Equal(t, 0.00002, provider.Cost())
	})
}

func TestOpenAIProvider_EmbedErrors(t *testing.T) {
	config := OpenAIConfig{
		APIKey: "test-key",
		Model:  "text-embedding-3-small",
	}

	provider, _ := NewOpenAI(config)
	ctx := context.Background()

	t.Run("empty text", func(t *testing.T) {
		_, err := provider.Embed(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty text")
	})
}

func TestOpenAIProvider_EmbedBatchErrors(t *testing.T) {
	config := OpenAIConfig{
		APIKey: "test-key",
		Model:  "text-embedding-3-small",
	}

	provider, _ := NewOpenAI(config)
	ctx := context.Background()

	t.Run("empty batch", func(t *testing.T) {
		_, err := provider.EmbedBatch(ctx, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty text batch")
	})

	t.Run("nil batch", func(t *testing.T) {
		_, err := provider.EmbedBatch(ctx, nil)
		assert.Error(t, err)
	})
}

func TestOpenAIProvider_Timeout(t *testing.T) {
	t.Run("default timeout", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey: "test-key",
		}

		provider, err := NewOpenAI(config)
		require.NoError(t, err)
		assert.Equal(t, 30, provider.config.Timeout)
	})

	t.Run("custom timeout", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey:  "test-key",
			Timeout: 60,
		}

		provider, err := NewOpenAI(config)
		require.NoError(t, err)
		assert.Equal(t, 60, provider.config.Timeout)
	})
}

func TestOpenAIProvider_IsAvailable(t *testing.T) {
	config := OpenAIConfig{
		APIKey: "test-key",
	}

	provider, _ := NewOpenAI(config)
	ctx := context.Background()

	// IsAvailable tries to make actual API call
	// With invalid test key, it should fail but not panic
	available := provider.IsAvailable(ctx)
	_ = available // Can't assert true/false without real API key
}
