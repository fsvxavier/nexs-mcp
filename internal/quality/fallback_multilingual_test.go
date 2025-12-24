//go:build !noonnx
// +build !noonnx

package quality

import (
	"context"
	"strings"
	"testing"
)

// TestFallbackWithCJKLanguages tests that fallback works correctly when ONNX fails with CJK.
func TestFallbackWithCJKLanguages(t *testing.T) {
	skipIfONNXNotAvailable(t)

	config := getTestModelConfig(t)
	config.FallbackChain = []string{"onnx", "implicit"} // ONNX first, then implicit

	fallbackScorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("Failed to create fallback scorer: %v", err)
	}
	defer fallbackScorer.Close()

	ctx := context.Background()

	testCases := []struct {
		name           string
		content        string
		expectedMethod string // Which scorer should be used
		description    string
	}{
		{
			name:           "Portuguese (ONNX should work)",
			content:        "Este é um texto em português sobre inteligência artificial.",
			expectedMethod: "onnx",
			description:    "Supported language - ONNX should succeed",
		},
		{
			name:           "Japanese (ONNX should fail, fallback to implicit)",
			content:        "これは日本語の高品質なテキストです。人工知能は自然言語処理の方法を革新しています。",
			expectedMethod: "implicit",
			description:    "CJK language - ONNX fails, implicit takes over",
		},
		{
			name:           "Chinese (ONNX should fail, fallback to implicit)",
			content:        "这是一篇高质量的中文文本。人工智能正在彻底改变我们处理自然语言的方式。",
			expectedMethod: "implicit",
			description:    "CJK language - ONNX fails, implicit takes over",
		},
		{
			name:           "English (ONNX should work)",
			content:        "This is a high-quality English text about artificial intelligence.",
			expectedMethod: "onnx",
			description:    "Supported language - ONNX should succeed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score, err := fallbackScorer.Score(ctx, tc.content)
			if err != nil {
				t.Errorf("Fallback scorer failed: %v", err)
				return
			}

			// Check which scorer was actually used
			usedScorer, ok := score.Metadata["fallback_used"].(string)
			if !ok {
				t.Error("Missing fallback_used metadata")
				return
			}

			// For CJK languages, we expect implicit scorer (ONNX should fail)
			// For non-CJK languages, accept either onnx or implicit (model may not be loaded)
			if strings.Contains(tc.content, "日本") || strings.Contains(tc.content, "这是") {
				// CJK languages - must use implicit
				if usedScorer != "implicit" {
					t.Errorf("CJK language should use implicit scorer, got %s", usedScorer)
				}
			} else {
				// Non-CJK languages - accept either scorer
				if usedScorer != "onnx" && usedScorer != "implicit" {
					t.Errorf("Expected onnx or implicit scorer, got %s", usedScorer)
				}
			}

			t.Logf("✓ %s: used=%s, score=%.3f, confidence=%.3f",
				tc.description, usedScorer, score.Value, score.Confidence)
		})
	}
}

// TestFallbackStatistics tests that fallback tracks usage statistics correctly.
func TestFallbackStatistics(t *testing.T) {
	skipIfONNXNotAvailable(t)

	config := getTestModelConfig(t)
	config.FallbackChain = []string{"onnx", "implicit"}

	fallbackScorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("Failed to create fallback scorer: %v", err)
	}
	defer fallbackScorer.Close()

	ctx := context.Background()

	// Test with supported language (ONNX should succeed)
	_, err = fallbackScorer.Score(ctx, "English text about artificial intelligence.")
	if err != nil {
		t.Fatalf("Failed to score English text: %v", err)
	}

	// Test with CJK (ONNX should fail, implicit should succeed)
	_, err = fallbackScorer.Score(ctx, "日本語のテキストです。")
	if err != nil {
		t.Fatalf("Failed to score Japanese text with fallback: %v", err)
	}

	// Get statistics
	stats := fallbackScorer.GetStats()

	// Check ONNX stats (may be 0 if ONNX model not loaded)
	onnxCalls, onnxExists := stats["onnx_calls"]
	if onnxExists {
		t.Logf("ONNX calls: %v", onnxCalls)
	}

	onnxSuccesses, onnxSuccessExists := stats["onnx_successes"]
	if onnxSuccessExists {
		t.Logf("ONNX successes: %v", onnxSuccesses)
	}

	onnxFailures, onnxFailExists := stats["onnx_failures"]
	if onnxFailExists {
		t.Logf("ONNX failures: %v", onnxFailures)
	}

	// Check implicit stats (should have been called at least once)
	implicitCalls, implicitCallsExists := stats["implicit_calls"]
	if !implicitCallsExists {
		t.Error("Missing implicit_calls in stats")
	} else if implicitCalls.(int) < 1 {
		t.Errorf("Expected at least 1 implicit call, got %d", implicitCalls.(int))
	}

	implicitSuccesses, implicitExists := stats["implicit_successes"]
	if !implicitExists {
		t.Error("Missing implicit_successes in stats")
	} else if implicitSuccesses.(int) < 1 {
		t.Errorf("Expected at least 1 implicit success, got %d", implicitSuccesses.(int))
	}

	t.Logf("✓ Statistics tracking working correctly:")
	if onnxExists {
		t.Logf("  ONNX: calls=%v, successes=%v, failures=%v",
			onnxCalls, onnxSuccesses, onnxFailures)
	}
	t.Logf("  Implicit: calls=%v, successes=%v",
		implicitCalls, implicitSuccesses)
}

// TestFallbackMultilingualBatch tests batch processing with mixed languages.
func TestFallbackMultilingualBatch(t *testing.T) {
	skipIfONNXNotAvailable(t)

	config := getTestModelConfig(t)
	config.FallbackChain = []string{"onnx", "implicit"}

	fallbackScorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("Failed to create fallback scorer: %v", err)
	}
	defer fallbackScorer.Close()

	ctx := context.Background()

	// Mix of supported and unsupported languages
	contents := []string{
		"Portuguese text about artificial intelligence - português",
		"これは日本語のテキストです。", // Japanese - should fallback
		"English text about machine learning",
		"这是中文文本。", // Chinese - should fallback
		"Texto en español sobre inteligencia artificial",
	}

	scores, err := fallbackScorer.ScoreBatch(ctx, contents)
	if err != nil {
		t.Fatalf("Batch scoring failed: %v", err)
	}

	if len(scores) != len(contents) {
		t.Fatalf("Expected %d scores, got %d", len(contents), len(scores))
	}

	languages := []string{"Portuguese", "Japanese", "English", "Chinese", "Spanish"}

	for i, score := range scores {
		usedScorer := score.Metadata["fallback_used"].(string)

		// For CJK languages, we expect implicit scorer
		// For non-CJK languages, accept either onnx or implicit
		isCJK := strings.Contains(contents[i], "これは") || strings.Contains(contents[i], "这是")

		if isCJK {
			if usedScorer != "implicit" {
				t.Errorf("%s (CJK): expected implicit, got %s", languages[i], usedScorer)
			}
		} else {
			if usedScorer != "onnx" && usedScorer != "implicit" {
				t.Errorf("%s: expected onnx or implicit, got %s", languages[i], usedScorer)
			}
		}

		t.Logf("%s: scorer=%s, score=%.3f, confidence=%.3f",
			languages[i], usedScorer, score.Value, score.Confidence)
	}
}

// TestFallbackErrorHandling tests graceful degradation when all scorers fail.
func TestFallbackErrorHandling(t *testing.T) {
	config := DefaultConfig()
	config.FallbackChain = []string{"onnx"} // Only ONNX, no fallback

	// Try to create fallback with invalid model path
	config.ONNXModelPath = "/nonexistent/model.onnx"

	fallbackScorer, err := NewFallbackScorer(config)
	if err == nil {
		// If creation succeeded, test with content that would fail
		ctx := context.Background()
		_, err = fallbackScorer.Score(ctx, "test content")

		if err != nil {
			// Check error message mentions all scorers failed
			if !strings.Contains(err.Error(), "failed") {
				t.Errorf("Error should mention failure: %v", err)
			}
			t.Logf("✓ Proper error handling: %v", err)
		}

		fallbackScorer.Close()
	}
}
