//go:build noonnx
// +build noonnx

package quality

import (
	"context"
	"fmt"
)

// ONNXScorer stub for builds without ONNX support
type ONNXScorer struct {
	config *Config
}

// NewONNXScorer creates a stub ONNX scorer that always returns unavailable
func NewONNXScorer(config *Config) (*ONNXScorer, error) {
	return nil, fmt.Errorf("ONNX support not available in this build")
}

// Score always returns an error
func (s *ONNXScorer) Score(ctx context.Context, content string) (*Score, error) {
	return nil, fmt.Errorf("ONNX support not available in this build")
}

// ScoreBatch always returns an error
func (s *ONNXScorer) ScoreBatch(ctx context.Context, contents []string) ([]*Score, error) {
	return nil, fmt.Errorf("ONNX support not available in this build")
}

// Name returns the scorer identifier
func (s *ONNXScorer) Name() string {
	return "onnx"
}

// IsAvailable always returns false
func (s *ONNXScorer) IsAvailable(ctx context.Context) bool {
	return false
}

// Cost returns 0
func (s *ONNXScorer) Cost() float64 {
	return 0.0
}

// Close is a no-op
func (s *ONNXScorer) Close() error {
	return nil
}
