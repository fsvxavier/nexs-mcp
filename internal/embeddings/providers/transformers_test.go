//go:build !noonnx
// +build !noonnx

package providers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransformers(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := TransformersConfig{
			Model: "all-MiniLM-L6-v2",
		}

		provider, err := NewTransformers(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "transformers", provider.Name())
		assert.Equal(t, 384, provider.Dimensions())
		assert.Equal(t, 0.0, provider.Cost())
	})

	t.Run("default model", func(t *testing.T) {
		config := TransformersConfig{}

		provider, err := NewTransformers(config)
		require.NoError(t, err)
		assert.Equal(t, "all-MiniLM-L6-v2", provider.config.Model)
		assert.Equal(t, 384, provider.Dimensions())
	})

	t.Run("custom cache dir", func(t *testing.T) {
		config := TransformersConfig{
			Model:    "all-MiniLM-L6-v2",
			CacheDir: "/tmp/models",
		}

		provider, err := NewTransformers(config)
		require.NoError(t, err)
		assert.Equal(t, "/tmp/models", provider.config.CacheDir)
	})
}

func TestTransformersProvider_Interface(t *testing.T) {
	config := TransformersConfig{}
	provider, _ := NewTransformers(config)

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "transformers", provider.Name())
	})

	t.Run("Dimensions", func(t *testing.T) {
		assert.Equal(t, 384, provider.Dimensions())
	})

	t.Run("Cost", func(t *testing.T) {
		assert.Equal(t, 0.0, provider.Cost()) // Local inference is free
	})
}

func TestTransformersProvider_IsAvailable(t *testing.T) {
	config := TransformersConfig{}
	provider, _ := NewTransformers(config)
	ctx := context.Background()

	// Currently returns false as model needs to be downloaded
	available := provider.IsAvailable(ctx)
	assert.False(t, available)
}

func TestTransformersProvider_Embed(t *testing.T) {
	config := TransformersConfig{}
	provider, _ := NewTransformers(config)
	ctx := context.Background()

	t.Run("empty text error", func(t *testing.T) {
		_, err := provider.Embed(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty text")
	})

	t.Run("requires external setup", func(t *testing.T) {
		_, err := provider.Embed(ctx, "test text")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model file not found")
	})
}

func TestTransformersProvider_EmbedBatch(t *testing.T) {
	config := TransformersConfig{}
	provider, _ := NewTransformers(config)
	ctx := context.Background()

	t.Run("empty batch error", func(t *testing.T) {
		_, err := provider.EmbedBatch(ctx, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty text batch")
	})

	t.Run("requires external setup", func(t *testing.T) {
		_, err := provider.EmbedBatch(ctx, []string{"text1", "text2"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model file not found")
	})
}

func TestNewSentenceTransformers(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := SentenceConfig{
			Model: "paraphrase-multilingual-MiniLM-L12-v2",
		}

		provider, err := NewSentenceTransformers(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "sentence", provider.Name())
		assert.Equal(t, 384, provider.Dimensions())
		assert.Equal(t, 0.0, provider.Cost())
	})

	t.Run("default model", func(t *testing.T) {
		config := SentenceConfig{}

		provider, err := NewSentenceTransformers(config)
		require.NoError(t, err)
		assert.Equal(t, "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2", provider.config.Model)
	})
}

func TestSentenceTransformersProvider_Interface(t *testing.T) {
	config := SentenceConfig{}
	provider, _ := NewSentenceTransformers(config)

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "sentence", provider.Name())
	})

	t.Run("Dimensions", func(t *testing.T) {
		assert.Equal(t, 384, provider.Dimensions())
	})

	t.Run("Cost", func(t *testing.T) {
		assert.Equal(t, 0.0, provider.Cost())
	})
}

func TestSentenceTransformersProvider_IsAvailable(t *testing.T) {
	config := SentenceConfig{}
	provider, _ := NewSentenceTransformers(config)
	ctx := context.Background()

	// Currently returns false as implementation is pending
	available := provider.IsAvailable(ctx)
	assert.False(t, available)
}

func TestSentenceTransformersProvider_Embed(t *testing.T) {
	config := SentenceConfig{}
	provider, _ := NewSentenceTransformers(config)
	ctx := context.Background()

	t.Run("empty text error", func(t *testing.T) {
		_, err := provider.Embed(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty text")
	})

	t.Run("requires external setup", func(t *testing.T) {
		_, err := provider.Embed(ctx, "test text")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires external service")
	})
}

func TestSentenceTransformersProvider_EmbedBatch(t *testing.T) {
	config := SentenceConfig{}
	provider, _ := NewSentenceTransformers(config)
	ctx := context.Background()

	t.Run("empty batch error", func(t *testing.T) {
		_, err := provider.EmbedBatch(ctx, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty text batch")
	})

	t.Run("delegates to individual calls", func(t *testing.T) {
		_, err := provider.EmbedBatch(ctx, []string{"text1"})
		assert.Error(t, err) // Will fail because Embed returns error
	})
}
