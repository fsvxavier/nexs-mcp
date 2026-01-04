//go:build noonnx
// +build noonnx

package application

import (
	"context"
	"errors"
)

// ONNXBERTProvider stub for builds without ONNX support.
type ONNXBERTProvider struct{}

// NewONNXBERTProvider returns a stub provider that always reports unavailable.
func NewONNXBERTProvider(config EnhancedNLPConfig) (*ONNXBERTProvider, error) {
	return &ONNXBERTProvider{}, nil
}

func (p *ONNXBERTProvider) ExtractEntities(ctx context.Context, text string) ([]EnhancedEntity, error) {
	return nil, errors.New("ONNX support not enabled (built with noonnx tag)")
}

func (p *ONNXBERTProvider) ExtractEntitiesBatch(ctx context.Context, texts []string) ([][]EnhancedEntity, error) {
	return nil, errors.New("ONNX support not enabled (built with noonnx tag)")
}

func (p *ONNXBERTProvider) AnalyzeSentiment(ctx context.Context, text string) (*SentimentResult, error) {
	return nil, errors.New("ONNX support not enabled (built with noonnx tag)")
}

func (p *ONNXBERTProvider) ExtractTopics(ctx context.Context, texts []string, numTopics int) ([]Topic, error) {
	return nil, errors.New("ONNX support not enabled (built with noonnx tag)")
}

func (p *ONNXBERTProvider) IsAvailable() bool {
	return false
}

func (p *ONNXBERTProvider) Close() error {
	return nil
}
