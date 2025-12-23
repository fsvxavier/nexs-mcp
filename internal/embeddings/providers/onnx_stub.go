//go:build noonnx
// +build noonnx

package providers

import (
	"context"
	"errors"
)

// ONNXConfig stub for builds without ONNX support
type ONNXConfig struct {
	Model    string
	CacheDir string
	UseGPU   bool
}

// ONNXProvider stub for builds without ONNX support
type ONNXProvider struct{}

// NewONNX returns an error when ONNX is disabled
func NewONNX(config ONNXConfig) (*ONNXProvider, error) {
	return nil, errors.New("ONNX support disabled in this build (compiled with noonnx tag)")
}

// NewONNXProvider returns an error when ONNX is disabled
func NewONNXProvider(modelPath string) (*ONNXProvider, error) {
	return nil, errors.New("ONNX support disabled in this build (compiled with noonnx tag)")
}

// GenerateEmbedding returns an error when ONNX is disabled
func (p *ONNXProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, errors.New("ONNX support disabled in this build")
}

// GenerateBatchEmbeddings returns an error when ONNX is disabled
func (p *ONNXProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, errors.New("ONNX support disabled in this build")
}

// Embed returns an error when ONNX is disabled (Provider interface)
func (p *ONNXProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, errors.New("ONNX support disabled in this build")
}

// EmbedBatch returns an error when ONNX is disabled (Provider interface)
func (p *ONNXProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, errors.New("ONNX support disabled in this build")
}

// IsAvailable always returns false for stub
func (p *ONNXProvider) IsAvailable(ctx context.Context) bool {
	return false
}

// Dimensions returns 0 when ONNX is disabled
func (p *ONNXProvider) Dimensions() int {
	return 0
}

// Name returns the provider name
func (p *ONNXProvider) Name() string {
	return "onnx-stub"
}

// Cost returns 0 for stub
func (p *ONNXProvider) Cost() float64 {
	return 0.0
}

// Close is a no-op
func (p *ONNXProvider) Close() error {
	return nil
}
