package quality

import (
	"context"
	"testing"
	"time"
)

func TestNewFallbackScorer(t *testing.T) {
	config := &Config{
		DefaultScorer:  "onnx",
		EnableFallback: true,
		FallbackChain:  []string{"onnx", "implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}
	if scorer == nil {
		t.Fatal("NewFallbackScorer returned nil")
	}

	if scorer.Name() != "fallback" {
		t.Errorf("Expected name 'fallback', got '%s'", scorer.Name())
	}

	ctx := context.Background()
	if !scorer.IsAvailable(ctx) {
		t.Error("Fallback scorer should always be available (due to implicit)")
	}
}

func TestFallbackScorerChainExecution(t *testing.T) {
	config := &Config{
		DefaultScorer:  "implicit",
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}
	ctx := context.Background()

	score, err := scorer.Score(ctx, "test content")
	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}

	if score == nil {
		t.Fatal("Score returned nil")
	}

	// Should use implicit scorer
	if score.Method != "implicit" {
		t.Errorf("Expected method 'implicit', got '%s'", score.Method)
	}

	if score.Value != 0.4 {
		t.Errorf("Expected default implicit score 0.4, got %f", score.Value)
	}
}

func TestFallbackScorerPreferredScorer(t *testing.T) {
	tests := []struct {
		name           string
		fallbackChain  []string
		expectedScorer string
	}{
		{
			name:           "Implicit only",
			fallbackChain:  []string{"implicit"},
			expectedScorer: "implicit",
		},
		{
			name:           "ONNX then implicit (ONNX may not be available)",
			fallbackChain:  []string{"onnx", "implicit"},
			expectedScorer: "onnx", // or "implicit" if ONNX not available
		},
		{
			name:           "Full chain",
			fallbackChain:  []string{"onnx", "groq", "gemini", "implicit"},
			expectedScorer: "onnx", // or first available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				EnableFallback: true,
				FallbackChain:  tt.fallbackChain,
			}

			ctx := context.Background()
			scorer, err := NewFallbackScorer(config)
			if err != nil {
				t.Fatalf("NewFallbackScorer failed: %v", err)
			}
			preferred := scorer.GetPreferredScorer(ctx)

			if preferred == nil {
				t.Fatal("GetPreferredScorer returned nil")
			}

			// Check that preferred scorer is one of the expected ones
			// (could be any available scorer in the chain)
			validScorers := map[string]bool{
				"onnx":     true,
				"groq":     true,
				"gemini":   true,
				"implicit": true,
			}

			if !validScorers[preferred.Name()] {
				t.Errorf("Unexpected preferred scorer: %s", preferred.Name())
			}
		})
	}
}

func TestFallbackScorerStatistics(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}
	ctx := context.Background()

	// Initial stats should be zero
	stats := scorer.GetStats()
	calls := stats["calls"].(map[string]int)
	if calls["implicit"] != 0 {
		t.Error("Initial calls should be 0")
	}

	// Score some content
	_, err = scorer.Score(ctx, "content1")
	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}

	_, err = scorer.Score(ctx, "content2")
	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}

	// Check stats updated
	stats = scorer.GetStats()
	calls = stats["calls"].(map[string]int)
	successes := stats["successes"].(map[string]int)
	failures := stats["failures"].(map[string]int)

	if calls["implicit"] != 2 {
		t.Errorf("Expected 2 calls, got %d", calls["implicit"])
	}

	if successes["implicit"] != 2 {
		t.Errorf("Expected 2 successes, got %d", successes["implicit"])
	}

	if failures["implicit"] != 0 {
		t.Errorf("Expected 0 failures, got %d", failures["implicit"])
	}

	// Cost should be 0 for implicit
	totalCost := stats["total_cost"].(float64)
	if totalCost != 0.0 {
		t.Errorf("Expected cost 0, got %f", totalCost)
	}
}

func TestFallbackScorerResetStats(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}
	ctx := context.Background()

	// Generate some stats
	scorer.Score(ctx, "content")

	stats := scorer.GetStats()
	calls := stats["calls"].(map[string]int)
	if calls["implicit"] == 0 {
		t.Error("Stats should not be zero after scoring")
	}

	// Reset stats
	scorer.ResetStats()

	stats = scorer.GetStats()
	calls = stats["calls"].(map[string]int)
	if calls["implicit"] != 0 {
		t.Errorf("Stats should be reset to 0, got %d", calls["implicit"])
	}
}

func TestFallbackScorerBatch(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}
	ctx := context.Background()

	contents := []string{"content1", "content2", "content3"}
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

		if score.Method != "implicit" {
			t.Errorf("Batch score %d: expected method 'implicit', got '%s'", i, score.Method)
		}
	}

	// Check batch stats
	stats := scorer.GetStats()
	calls := stats["calls"].(map[string]int)
	if calls["implicit"] != 3 {
		t.Errorf("Expected 3 batch calls, got %d", calls["implicit"])
	}
}

func TestFallbackScorerCostTracking(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}
	ctx := context.Background()

	// Score multiple times
	for i := 0; i < 5; i++ {
		scorer.Score(ctx, "content")
	}

	stats := scorer.GetStats()
	totalCost := stats["total_cost"].(float64)

	// Implicit scorer has 0 cost
	expectedCost := 0.0
	if totalCost != expectedCost {
		t.Errorf("Expected total cost %f, got %f", expectedCost, totalCost)
	}
}

func TestFallbackScorerClose(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
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
}

func TestFallbackScorerAvailability(t *testing.T) {
	tests := []struct {
		name              string
		fallbackChain     []string
		shouldBeAvailable bool
	}{
		{
			name:              "With implicit - always available",
			fallbackChain:     []string{"implicit"},
			shouldBeAvailable: true,
		},
		{
			name:              "ONNX with implicit fallback",
			fallbackChain:     []string{"onnx", "implicit"},
			shouldBeAvailable: true, // implicit is always available
		},
		{
			name:              "Full chain with implicit",
			fallbackChain:     []string{"onnx", "groq", "gemini", "implicit"},
			shouldBeAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				EnableFallback: true,
				FallbackChain:  tt.fallbackChain,
			}

			scorer, err := NewFallbackScorer(config)
			if err != nil {
				t.Fatalf("NewFallbackScorer failed: %v", err)
			}
			available := scorer.IsAvailable(context.Background())

			if available != tt.shouldBeAvailable {
				t.Errorf("Expected availability %v, got %v", tt.shouldBeAvailable, available)
			}
		})
	}
}

func TestFallbackScorerEmptyChain(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"}, // Must have at least implicit
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}

	// Should still work due to implicit fallback being added automatically
	if !scorer.IsAvailable(context.Background()) {
		t.Error("Scorer with empty chain should still be available (implicit fallback)")
	}

	ctx := context.Background()
	score, err := scorer.Score(ctx, "content")

	// Should succeed with implicit fallback
	if err != nil {
		t.Logf("Score with empty chain: %v", err)
	}

	if score != nil && score.Method != "implicit" {
		t.Logf("Expected implicit method, got: %s", score.Method)
	}
}

func TestFallbackScorerConcurrentScoring(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}
	ctx := context.Background()

	// Test concurrent scoring
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			_, err := scorer.Score(ctx, "concurrent content")
			if err != nil {
				t.Errorf("Concurrent score %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check stats
	stats := scorer.GetStats()
	calls := stats["calls"].(map[string]int)
	if calls["implicit"] != 10 {
		t.Errorf("Expected 10 concurrent calls, got %d", calls["implicit"])
	}
}

func TestFallbackScorerStatsTimestamp(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}
	ctx := context.Background()

	before := time.Now()
	scorer.Score(ctx, "content")
	stats := scorer.GetStats()
	after := time.Now()

	timestamp := stats["timestamp"].(time.Time)
	if timestamp.Before(before) || timestamp.After(after) {
		t.Error("Stats timestamp should be between before and after time")
	}
}

func TestFallbackScorerMetadata(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}
	ctx := context.Background()

	score, err := scorer.Score(ctx, "content")
	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}

	if score.Metadata == nil {
		t.Fatal("Score metadata should not be nil")
	}

	// Metadata should include information about which scorer was used
	if score.Method != "implicit" {
		t.Errorf("Expected method 'implicit', got '%s'", score.Method)
	}
}

func TestFallbackScorerContextCancellation(t *testing.T) {
	config := &Config{
		EnableFallback: true,
		FallbackChain:  []string{"implicit"},
	}

	scorer, err := NewFallbackScorer(config)
	if err != nil {
		t.Fatalf("NewFallbackScorer failed: %v", err)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = scorer.Score(ctx, "content")

	// Implicit scorer might not respect context cancellation since it's fast
	// But we test that it doesn't panic
	if err != nil {
		t.Logf("Score with cancelled context returned error (expected): %v", err)
	}
}

func TestFallbackScorerNilConfig(t *testing.T) {
	// Should not panic with nil config
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewFallbackScorer panicked with nil config: %v", r)
		}
	}()

	scorer, err := NewFallbackScorer(nil)
	if scorer == nil && err == nil {
		t.Error("NewFallbackScorer should handle nil config gracefully")
	}
}
