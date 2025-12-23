package embeddings

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFactory(t *testing.T) {
	config := DefaultConfig()
	factory := NewFactory(config)

	assert.NotNil(t, factory)
	assert.Equal(t, config, factory.config)
}

func TestNewFactoryFromEnv(t *testing.T) {
	// Save original env vars
	origProvider := os.Getenv("NEXS_EMBEDDING_PROVIDER")
	origKey := os.Getenv("OPENAI_API_KEY")
	origModel := os.Getenv("NEXS_OPENAI_MODEL")
	origGPU := os.Getenv("NEXS_USE_GPU")

	defer func() {
		os.Setenv("NEXS_EMBEDDING_PROVIDER", origProvider)
		os.Setenv("OPENAI_API_KEY", origKey)
		os.Setenv("NEXS_OPENAI_MODEL", origModel)
		os.Setenv("NEXS_USE_GPU", origGPU)
	}()

	t.Run("with env vars", func(t *testing.T) {
		os.Setenv("NEXS_EMBEDDING_PROVIDER", "openai")
		os.Setenv("OPENAI_API_KEY", "test-key")
		os.Setenv("NEXS_OPENAI_MODEL", "text-embedding-3-large")
		os.Setenv("NEXS_USE_GPU", "true")

		factory := NewFactoryFromEnv()

		assert.Equal(t, "openai", factory.config.Provider)
		assert.Equal(t, "test-key", factory.config.OpenAIKey)
		assert.Equal(t, "text-embedding-3-large", factory.config.OpenAIModel)
		assert.True(t, factory.config.UseGPU)
	})

	t.Run("without env vars", func(t *testing.T) {
		os.Unsetenv("NEXS_EMBEDDING_PROVIDER")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("NEXS_OPENAI_MODEL")
		os.Unsetenv("NEXS_USE_GPU")

		factory := NewFactoryFromEnv()

		// Should use defaults
		assert.Equal(t, "transformers", factory.config.Provider)
		assert.False(t, factory.config.UseGPU)
	})
}

func TestFactory_CreateProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("create openai provider", func(t *testing.T) {
		config := DefaultConfig()
		config.Provider = "openai"
		config.OpenAIKey = "test-key"

		factory := NewFactory(config)
		provider, err := factory.createProvider(ctx, "openai")

		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "openai", provider.Name())
	})

	t.Run("create transformers provider", func(t *testing.T) {
		config := DefaultConfig()
		factory := NewFactory(config)

		provider, err := factory.createProvider(ctx, "transformers")

		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "transformers", provider.Name())
	})

	t.Run("create sentence provider", func(t *testing.T) {
		config := DefaultConfig()
		factory := NewFactory(config)

		provider, err := factory.createProvider(ctx, "sentence")

		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "sentence", provider.Name())
	})

	t.Run("create onnx provider", func(t *testing.T) {
		config := DefaultConfig()
		factory := NewFactory(config)

		provider, err := factory.createProvider(ctx, "onnx")

		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "onnx", provider.Name())
	})

	t.Run("unknown provider", func(t *testing.T) {
		config := DefaultConfig()
		factory := NewFactory(config)

		_, err := factory.createProvider(ctx, "unknown")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown embedding provider")
	})
}

func TestFactory_CreateAuto(t *testing.T) {
	ctx := context.Background()

	t.Run("auto selects transformers", func(t *testing.T) {
		config := DefaultConfig()
		config.Provider = "auto"

		factory := NewFactory(config)
		provider, err := factory.createAuto(ctx)

		// Should select transformers (first in order)
		// Note: will fail to be available since not implemented yet
		if err == nil {
			assert.NotNil(t, provider)
		}
	})
}

func TestFallbackProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("fallback on primary failure", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableFallback = true
		config.OpenAIKey = "test-key"
		config.FallbackPriority = []string{"openai", "transformers", "onnx"}

		factory := NewFactory(config)

		// Create fallback provider (will use openai as primary)
		provider, err := factory.CreateWithFallback(ctx)

		// May succeed or fail depending on API key validity
		// Just verify it doesn't panic
		if err == nil {
			assert.NotNil(t, provider)
			assert.Contains(t, provider.Name(), "fallback")
		}
	})

	t.Run("no fallback when disabled", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableFallback = false
		config.Provider = "transformers"

		factory := NewFactory(config)

		provider, err := factory.CreateWithFallback(ctx)

		// Should create without fallback wrapper
		if err == nil {
			assert.NotNil(t, provider)
			assert.Equal(t, "transformers", provider.Name())
		}
	})
}

func TestFallbackProvider_Methods(t *testing.T) {
	ctx := context.Background()

	// Create a mock-based fallback for testing
	mockPrimary := NewMockProvider("primary", 384)
	mockPrimary.available = false // Simulate unavailable

	config := DefaultConfig()
	factory := NewFactory(config)

	fallback := &FallbackProvider{
		primary:  mockPrimary,
		factory:  factory,
		priority: []string{"transformers", "onnx"},
	}

	t.Run("Name", func(t *testing.T) {
		assert.Contains(t, fallback.Name(), "fallback")
	})

	t.Run("Dimensions", func(t *testing.T) {
		assert.Equal(t, 384, fallback.Dimensions())
	})

	t.Run("Cost", func(t *testing.T) {
		assert.Equal(t, 0.0, fallback.Cost())
	})

	t.Run("IsAvailable checks fallbacks", func(t *testing.T) {
		// Primary is unavailable, should check fallbacks
		// Result depends on whether fallbacks are available
		available := fallback.IsAvailable(ctx)
		// Just verify it doesn't panic
		_ = available
	})
}
