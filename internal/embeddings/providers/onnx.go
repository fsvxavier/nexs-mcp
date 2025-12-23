package providers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ONNXConfig holds configuration for ONNX Runtime inference
type ONNXConfig struct {
	Model    string // Model name (e.g., ms-marco-MiniLM-L-12-v2)
	CacheDir string // Local cache directory for ONNX models
	UseGPU   bool   // Enable GPU acceleration via CUDA/TensorRT
}

// ONNXProvider implements embeddings using ONNX Runtime
// Provides fast inference with CPU/GPU acceleration
type ONNXProvider struct {
	config ONNXConfig
	dims   int
}

// NewONNX creates a new ONNX embedding provider
func NewONNX(config ONNXConfig) (*ONNXProvider, error) {
	if config.Model == "" {
		config.Model = "ms-marco-MiniLM-L-12-v2"
	}

	if config.CacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		config.CacheDir = filepath.Join(homeDir, ".cache", "nexs-mcp", "onnx-models")
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	dims := getONNXDimensions(config.Model)

	return &ONNXProvider{
		config: config,
		dims:   dims,
	}, nil
}

func getONNXDimensions(model string) int {
	switch model {
	case "ms-marco-MiniLM-L-12-v2":
		return 384
	case "ms-marco-MiniLM-L-6-v2":
		return 384
	case "all-MiniLM-L6-v2":
		return 384
	default:
		return 384
	}
}

func (o *ONNXProvider) Name() string {
	return "onnx"
}

func (o *ONNXProvider) Dimensions() int {
	return o.dims
}

func (o *ONNXProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, errors.New("empty text provided")
	}

	// ONNX Runtime implementation requires:
	// 1. Install onnxruntime shared library: https://github.com/microsoft/onnxruntime/releases
	// 2. Add dependency: github.com/yalue/onnxruntime_go
	// 3. Export sentence-transformer model to ONNX format
	// 4. Implement tokenization (BPE/WordPiece)
	//
	// Example implementation:
	// - Load ONNX model with ort.NewSession(modelPath)
	// - Tokenize text to input_ids, attention_mask tensors
	// - Run session.Run() to get embeddings
	// - Extract and normalize output vector
	//
	// Performance:
	// - CPU: 50-100ms per embedding
	// - GPU (CUDA): 10-20ms per embedding
	// - Model size: ~23MB for MiniLM

	return nil, fmt.Errorf("ONNX provider requires setup: install onnxruntime library and export model to ONNX format (see docs)")
}

func (o *ONNXProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("empty text batch")
	}

	// Batch inference with ONNX Runtime:
	// - Tokenize all texts into single batch tensors
	// - Run single inference call with batch
	// - Extract embeddings for each text from output
	//
	// Benefits:
	// - 10x faster than sequential for batches of 10+
	// - Better GPU utilization (if using CUDA)
	// - Amortized tokenization and model loading overhead

	// For now, delegate to individual calls
	results := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := o.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		results[i] = emb
	}
	return results, nil
}

func (o *ONNXProvider) IsAvailable(ctx context.Context) bool {
	// Check if ONNX model exists in cache
	modelPath := filepath.Join(o.config.CacheDir, o.config.Model+".onnx")
	_, err := os.Stat(modelPath)

	// For now, return false until ONNX Runtime is integrated
	return err == nil && false
}

func (o *ONNXProvider) Cost() float64 {
	return 0.0 // Local ONNX inference is free
}

// GetExecutionProvider returns the optimal ONNX execution provider
func (o *ONNXProvider) GetExecutionProvider() string {
	if o.config.UseGPU {
		// Try CUDA first, fallback to CPU
		return "CUDAExecutionProvider"
	}
	return "CPUExecutionProvider"
}

// EstimateLatency returns expected latency for this configuration
func (o *ONNXProvider) EstimateLatency() string {
	if o.config.UseGPU {
		return "10-20ms per embedding (GPU)"
	}
	return "50-100ms per embedding (CPU)"
}

// ModelSize returns the approximate model size
func (o *ONNXProvider) ModelSize() string {
	switch o.config.Model {
	case "ms-marco-MiniLM-L-12-v2":
		return "23MB"
	case "ms-marco-MiniLM-L-6-v2":
		return "12MB"
	default:
		return "~20MB"
	}
}
