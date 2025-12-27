//go:build !noonnx
// +build !noonnx

package providers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewONNX(t *testing.T) {
	t.Run("valid config CPU", func(t *testing.T) {
		config := ONNXConfig{
			Model: "ms-marco-MiniLM-L-12-v2",
		}

		provider, err := NewONNX(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "onnx", provider.Name())
		assert.Equal(t, 384, provider.Dimensions())
		assert.Equal(t, 0.0, provider.Cost())
	})

	t.Run("valid config GPU", func(t *testing.T) {
		config := ONNXConfig{
			Model:  "ms-marco-MiniLM-L-12-v2",
			UseGPU: true,
		}

		provider, err := NewONNX(config)
		require.NoError(t, err)
		assert.True(t, provider.config.UseGPU)
	})

	t.Run("default model", func(t *testing.T) {
		config := ONNXConfig{}

		provider, err := NewONNX(config)
		require.NoError(t, err)
		assert.Equal(t, "ms-marco-MiniLM-L-12-v2", provider.config.Model)
	})

	t.Run("custom cache dir", func(t *testing.T) {
		config := ONNXConfig{
			CacheDir: "/tmp/onnx-models",
		}

		provider, err := NewONNX(config)
		require.NoError(t, err)
		assert.Equal(t, "/tmp/onnx-models", provider.config.CacheDir)
	})
}

func TestONNXProvider_Interface(t *testing.T) {
	config := ONNXConfig{
		Model: "ms-marco-MiniLM-L-12-v2",
	}
	provider, _ := NewONNX(config)

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "onnx", provider.Name())
	})

	t.Run("Dimensions", func(t *testing.T) {
		assert.Equal(t, 384, provider.Dimensions())
	})

	t.Run("Cost", func(t *testing.T) {
		assert.Equal(t, 0.0, provider.Cost()) // Local inference is free
	})
}

func TestONNXProvider_IsAvailable(t *testing.T) {
	config := ONNXConfig{
		Model: "ms-marco-MiniLM-L-12-v2",
	}
	provider, _ := NewONNX(config)
	ctx := context.Background()

	// Currently returns false as model needs to be downloaded
	available := provider.IsAvailable(ctx)
	assert.False(t, available)
}

func TestONNXProvider_Embed(t *testing.T) {
	config := ONNXConfig{}
	provider, _ := NewONNX(config)
	ctx := context.Background()

	t.Run("empty text error", func(t *testing.T) {
		_, err := provider.Embed(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty text")
	})

	t.Run("requires setup", func(t *testing.T) {
		_, err := provider.Embed(ctx, "test text")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires setup")
	})
}

func TestONNXProvider_EmbedBatch(t *testing.T) {
	config := ONNXConfig{}
	provider, _ := NewONNX(config)
	ctx := context.Background()

	t.Run("empty batch error", func(t *testing.T) {
		_, err := provider.EmbedBatch(ctx, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty text batch")
	})

	t.Run("requires setup", func(t *testing.T) {
		_, err := provider.EmbedBatch(ctx, []string{"text1", "text2"})
		assert.Error(t, err)
		// Will fail because Embed returns error
	})
}

func TestONNXProvider_GetExecutionProvider(t *testing.T) {
	t.Run("CPU mode", func(t *testing.T) {
		config := ONNXConfig{
			UseGPU: false,
		}
		provider, _ := NewONNX(config)

		ep := provider.GetExecutionProvider()
		assert.Equal(t, "CPUExecutionProvider", ep)
	})

	t.Run("GPU mode", func(t *testing.T) {
		config := ONNXConfig{
			UseGPU: true,
		}
		provider, _ := NewONNX(config)

		ep := provider.GetExecutionProvider()
		assert.Equal(t, "CUDAExecutionProvider", ep)
	})
}

func TestONNXProvider_EstimateLatency(t *testing.T) {
	t.Run("CPU latency", func(t *testing.T) {
		config := ONNXConfig{
			UseGPU: false,
		}
		provider, _ := NewONNX(config)

		latency := provider.EstimateLatency()
		assert.Contains(t, latency, "50-100ms")
		assert.Contains(t, latency, "CPU")
	})

	t.Run("GPU latency", func(t *testing.T) {
		config := ONNXConfig{
			UseGPU: true,
		}
		provider, _ := NewONNX(config)

		latency := provider.EstimateLatency()
		assert.Contains(t, latency, "10-20ms")
		assert.Contains(t, latency, "GPU")
	})
}

func TestONNXProvider_ModelSize(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		wantSize string
	}{
		{"ms-marco-12", "ms-marco-MiniLM-L-12-v2", "23MB"},
		{"ms-marco-6", "ms-marco-MiniLM-L-6-v2", "12MB"},
		{"unknown", "unknown-model", "~20MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ONNXConfig{
				Model: tt.model,
			}
			provider, _ := NewONNX(config)

			size := provider.ModelSize()
			assert.Equal(t, tt.wantSize, size)
		})
	}
}

func TestONNXProvider_DimensionsByModel(t *testing.T) {
	tests := []struct {
		model string
		dims  int
	}{
		{"ms-marco-MiniLM-L-12-v2", 384},
		{"ms-marco-MiniLM-L-6-v2", 384},
		{"all-MiniLM-L6-v2", 384},
		{"unknown-model", 384},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			config := ONNXConfig{
				Model: tt.model,
			}
			provider, _ := NewONNX(config)
			assert.Equal(t, tt.dims, provider.Dimensions())
		})
	}
}
