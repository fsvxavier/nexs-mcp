//go:build noonnx
// +build noonnx

package providers

import (
	"context"
	"errors"
)

// TransformersConfig stub for builds without ONNX support
type TransformersConfig struct {
	Model    string
	CacheDir string
	UseGPU   bool
}

// SentenceConfig stub for builds without ONNX support
type SentenceConfig struct {
	Model    string
	CacheDir string
	UseGPU   bool
}

// TransformersProvider stub for builds without ONNX support
type TransformersProvider struct{}

// NewTransformers returns an error when ONNX is disabled
func NewTransformers(config TransformersConfig) (*TransformersProvider, error) {
	return nil, errors.New("Transformers/ONNX support disabled in this build (compiled with noonnx tag)")
}

// NewTransformersProvider returns an error when ONNX is disabled
func NewTransformersProvider(modelPath string) (*TransformersProvider, error) {
	return nil, errors.New("Transformers/ONNX support disabled in this build (compiled with noonnx tag)")
}

// NewSentenceTransformers returns an error when ONNX is disabled
func NewSentenceTransformers(config SentenceConfig) (*TransformersProvider, error) {
	return nil, errors.New("SentenceTransformers/ONNX support disabled in this build (compiled with noonnx tag)")
}

// GenerateEmbedding returns an error when ONNX is disabled
func (p *TransformersProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, errors.New("Transformers/ONNX support disabled in this build")
}

// GenerateBatchEmbeddings returns an error when ONNX is disabled
func (p *TransformersProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, errors.New("Transformers/ONNX support disabled in this build")
}

// Embed returns an error when ONNX is disabled (Provider interface)
func (p *TransformersProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, errors.New("Transformers/ONNX support disabled in this build")
}

// EmbedBatch returns an error when ONNX is disabled (Provider interface)
func (p *TransformersProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, errors.New("Transformers/ONNX support disabled in this build")
}

// IsAvailable always returns false for stub
func (p *TransformersProvider) IsAvailable(ctx context.Context) bool {
	return false
}

// Dimensions returns 0 when ONNX is disabled
func (p *TransformersProvider) Dimensions() int {
	return 0
}

// Name returns the provider name
func (p *TransformersProvider) Name() string {
	return "transformers-stub"
}

// Cost returns 0 for stub
func (p *TransformersProvider) Cost() float64 {
	return 0.0
}

// Close is a no-op
func (p *TransformersProvider) Close() error {
	return nil
}
