//go:build !noonnx
// +build !noonnx

package quality

import (
	"context"
	"strings"
	"testing"
)

func TestNewONNXScorer(t *testing.T) {
	config := getTestModelConfig(t)

	scorer, err := NewONNXScorer(config)
	if err != nil {
		// Skip if ONNX Runtime is not available
		t.Skipf("ONNX Runtime not available: %v\n"+
			"Hint: Run 'make install-onnx' or see docs/development/ONNX_SETUP.md", err)
	}

	defer scorer.Close()

	if scorer.Name() != "onnx" {
		t.Errorf("Expected name 'onnx', got '%s'", scorer.Name())
	}

	// Cost should be non-zero (inference time)
	if scorer.Cost() <= 0 {
		t.Errorf("Expected positive cost, got %f", scorer.Cost())
	}
}

func TestONNXScorerNotAvailableWithoutModel(t *testing.T) {
	config := &Config{
		ONNXModelPath: "/nonexistent/model.onnx",
	}

	scorer, err := NewONNXScorer(config)

	// Should fail without model file
	if err == nil {
		if scorer != nil {
			scorer.Close()
		}
		t.Error("Expected error when model file doesn't exist")
	}
}

func TestONNXScorerIsAvailable(t *testing.T) {
	config := getTestModelConfig(t)

	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()
	if !scorer.IsAvailable(ctx) {
		t.Error("ONNX scorer should be available after successful initialization")
	}
}

func TestONNXScorerScore(t *testing.T) {
	config := getTestModelConfig(t)

	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()
	score, err := scorer.Score(ctx, "This is a test content for quality scoring")

	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}

	assertValidScore(t, score, "onnx")

	// Should have metadata with inference time
	if score.Metadata == nil {
		t.Error("Metadata should not be nil")
	}
}

func TestONNXScorerEmptyContent(t *testing.T) {
	config := getTestModelConfig(t)

	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()
	score, err := scorer.Score(ctx, "")

	// Should handle empty content gracefully
	if err != nil {
		t.Logf("Score with empty content returned error: %v", err)
	}

	if score != nil {
		if score.Value < 0 || score.Value > 1 {
			t.Errorf("Score value %f out of valid range", score.Value)
		}
	}
}

func TestONNXScorerBatch(t *testing.T) {
	config := getTestModelConfig(t)

	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()
	contents := []string{
		"First test content",
		"Second test content",
		"Third test content",
	}

	scores, err := scorer.ScoreBatch(ctx, contents)
	if err != nil {
		t.Fatalf("ScoreBatch failed: %v", err)
	}

	if len(scores) != len(contents) {
		t.Errorf("Expected %d scores, got %d", len(contents), len(scores))
	}

	for i, score := range scores {
		if score == nil {
			t.Errorf("Score %d is nil", i)
			continue
		}

		assertValidScore(t, score, "onnx")
	}
}

func TestONNXScorerClose(t *testing.T) {
	config := getTestModelConfig(t)

	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}

	err = scorer.Close()
	if err != nil {
		t.Errorf("Close should not return error, got: %v", err)
	}

	// Should be safe to close multiple times
	err = scorer.Close()
	if err != nil {
		t.Errorf("Second Close should not return error, got: %v", err)
	}

	// After close, scorer should not be available
	ctx := context.Background()
	if scorer.IsAvailable(ctx) {
		t.Error("Scorer should not be available after Close")
	}
}

func TestONNXScorerConcurrentScoring(t *testing.T) {
	config := getTestModelConfig(t)

	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()

	// Test concurrent scoring
	done := make(chan bool)
	for i := range 5 {
		go func(id int) {
			_, err := scorer.Score(ctx, "concurrent test content")
			if err != nil {
				t.Errorf("Concurrent score %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 5 {
		<-done
	}
}

func TestONNXScorerLongContent(t *testing.T) {
	config := getTestModelConfig(t)

	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()

	// Test with content longer than 512 tokens
	longContent := ""
	var longContentSb215 strings.Builder
	for range 1000 {
		longContentSb215.WriteString("This is a very long content that exceeds the maximum input length. ")
	}
	longContent += longContentSb215.String()

	score, err := scorer.Score(ctx, longContent)
	if err != nil {
		t.Fatalf("Score with long content failed: %v", err)
	}

	if score == nil {
		t.Fatal("Score returned nil for long content")
	}

	// Should truncate and still produce valid score
	assertValidScore(t, score, "onnx")
}
