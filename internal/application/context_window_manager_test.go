package application

import (
	"context"
	"testing"
	"time"
)

func TestContextWindowManager_NoOptimizationNeeded(t *testing.T) {
	config := ContextWindowConfig{
		MaxTokens:        1000,
		PriorityStrategy: PriorityRecency,
		TruncationMethod: TruncationHead,
		PreserveRecent:   5,
	}

	manager := NewContextWindowManager(config)

	items := []ContextItem{
		{ID: "1", Content: "Short content", TokenCount: 100, CreatedAt: time.Now(), Relevance: 0.8},
		{ID: "2", Content: "Another short", TokenCount: 100, CreatedAt: time.Now(), Relevance: 0.7},
		{ID: "3", Content: "Third item", TokenCount: 100, CreatedAt: time.Now(), Relevance: 0.6},
	}

	optimized, result, err := manager.OptimizeContext(context.Background(), items)

	if err != nil {
		t.Fatalf("OptimizeContext failed: %v", err)
	}

	if result.ItemsRemoved != 0 {
		t.Errorf("Expected no items removed, got %d", result.ItemsRemoved)
	}

	if len(optimized) != len(items) {
		t.Errorf("Expected %d items, got %d", len(items), len(optimized))
	}

	if result.Method != "none" {
		t.Errorf("Expected method='none', got %s", result.Method)
	}

	t.Logf("Result: tokens=%d, relevance=%.2f", result.OptimizedTokenCount, result.RelevanceScore)
}

func TestContextWindowManager_TruncationHead(t *testing.T) {
	config := ContextWindowConfig{
		MaxTokens:        500,
		PriorityStrategy: PriorityRecency,
		TruncationMethod: TruncationHead,
		PreserveRecent:   2,
	}

	manager := NewContextWindowManager(config)

	now := time.Now()
	items := []ContextItem{
		{ID: "1", Content: "Recent 1", TokenCount: 200, CreatedAt: now, Relevance: 0.9},
		{ID: "2", Content: "Recent 2", TokenCount: 200, CreatedAt: now.Add(-1 * time.Hour), Relevance: 0.8},
		{ID: "3", Content: "Old 1", TokenCount: 200, CreatedAt: now.Add(-2 * time.Hour), Relevance: 0.7},
		{ID: "4", Content: "Old 2", TokenCount: 200, CreatedAt: now.Add(-3 * time.Hour), Relevance: 0.6},
	}

	optimized, result, err := manager.OptimizeContext(context.Background(), items)

	if err != nil {
		t.Fatalf("OptimizeContext failed: %v", err)
	}

	if result.ItemsRemoved == 0 {
		t.Error("Expected items to be removed")
	}

	if result.OptimizedTokenCount > config.MaxTokens {
		t.Errorf("Optimized tokens %d exceeds max %d", result.OptimizedTokenCount, config.MaxTokens)
	}

	// Should keep most recent items (after sorting by recency)
	if len(optimized) < 2 {
		t.Errorf("Expected at least 2 items preserved, got %d", len(optimized))
	}

	t.Logf("TruncationHead: removed=%d, retained=%d, tokens=%d/%d",
		result.ItemsRemoved, result.ItemsRetained, result.OptimizedTokenCount, result.OriginalTokenCount)
}

func TestContextWindowManager_PriorityRelevance(t *testing.T) {
	config := ContextWindowConfig{
		MaxTokens:        400,
		PriorityStrategy: PriorityRelevance,
		TruncationMethod: TruncationHead,
		PreserveRecent:   1,
	}

	manager := NewContextWindowManager(config)

	items := []ContextItem{
		{ID: "low", Content: "Low relevance", TokenCount: 150, CreatedAt: time.Now(), Relevance: 0.2},
		{ID: "high", Content: "High relevance", TokenCount: 150, CreatedAt: time.Now(), Relevance: 0.9},
		{ID: "medium", Content: "Medium relevance", TokenCount: 150, CreatedAt: time.Now(), Relevance: 0.5},
		{ID: "highest", Content: "Highest relevance", TokenCount: 150, CreatedAt: time.Now(), Relevance: 0.95},
	}

	optimized, result, err := manager.OptimizeContext(context.Background(), items)

	if err != nil {
		t.Fatalf("OptimizeContext failed: %v", err)
	}

	// Should keep items with highest relevance
	if len(optimized) == 0 {
		t.Fatal("Expected at least one item retained")
	}

	// First item should be highest relevance (0.95)
	if optimized[0].ID != "highest" {
		t.Errorf("Expected first item to be 'highest', got '%s'", optimized[0].ID)
	}

	// Should have high average relevance
	if result.RelevanceScore < 0.7 {
		t.Errorf("Expected high avg relevance, got %.2f", result.RelevanceScore)
	}

	t.Logf("PriorityRelevance: avg_relevance=%.2f, items=%d", result.RelevanceScore, result.ItemsRetained)
}

func TestContextWindowManager_HybridStrategy(t *testing.T) {
	config := ContextWindowConfig{
		MaxTokens:        600,
		PriorityStrategy: PriorityHybrid,
		TruncationMethod: TruncationHead,
		PreserveRecent:   2,
	}

	manager := NewContextWindowManager(config)

	now := time.Now()
	items := []ContextItem{
		{ID: "recent_low", Content: "Recent low relevance", TokenCount: 200,
			CreatedAt: now, Relevance: 0.3, Importance: 3},
		{ID: "old_high", Content: "Old high relevance", TokenCount: 200,
			CreatedAt: now.Add(-10 * 24 * time.Hour), Relevance: 0.9, Importance: 7},
		{ID: "medium_both", Content: "Medium both", TokenCount: 200,
			CreatedAt: now.Add(-5 * 24 * time.Hour), Relevance: 0.6, Importance: 5},
		{ID: "recent_high", Content: "Recent high relevance", TokenCount: 200,
			CreatedAt: now.Add(-1 * time.Hour), Relevance: 0.85, Importance: 8},
	}

	optimized, result, err := manager.OptimizeContext(context.Background(), items)

	if err != nil {
		t.Fatalf("OptimizeContext failed: %v", err)
	}

	if len(optimized) == 0 {
		t.Fatal("Expected at least one item retained")
	}

	// Hybrid strategy should balance recency, relevance, and importance
	t.Logf("Hybrid: retained=%d, avg_relevance=%.2f", result.ItemsRetained, result.RelevanceScore)

	for i, item := range optimized {
		t.Logf("  Item %d: id=%s, relevance=%.2f, importance=%d", i, item.ID, item.Relevance, item.Importance)
	}

	// recent_high should score very high (recent + high relevance + high importance)
	found := false
	for _, item := range optimized {
		if item.ID == "recent_high" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'recent_high' to be retained with hybrid strategy")
	}
}

func TestContextWindowManager_TruncationMiddle(t *testing.T) {
	config := ContextWindowConfig{
		MaxTokens:        500,
		PriorityStrategy: PriorityRecency,
		TruncationMethod: TruncationMiddle,
		PreserveRecent:   2,
	}

	manager := NewContextWindowManager(config)

	now := time.Now()
	items := []ContextItem{
		{ID: "first", Content: "First", TokenCount: 100, CreatedAt: now, Relevance: 0.9},
		{ID: "middle1", Content: "Middle 1", TokenCount: 100, CreatedAt: now.Add(-1 * time.Hour), Relevance: 0.5},
		{ID: "middle2", Content: "Middle 2", TokenCount: 100, CreatedAt: now.Add(-2 * time.Hour), Relevance: 0.4},
		{ID: "middle3", Content: "Middle 3", TokenCount: 100, CreatedAt: now.Add(-3 * time.Hour), Relevance: 0.3},
		{ID: "last", Content: "Last", TokenCount: 100, CreatedAt: now.Add(-4 * time.Hour), Relevance: 0.8},
	}

	optimized, result, err := manager.OptimizeContext(context.Background(), items)

	if err != nil {
		t.Fatalf("OptimizeContext failed: %v", err)
	}

	// Should preserve first 2 and last 2 items
	ids := make([]string, len(optimized))
	for i, item := range optimized {
		ids[i] = item.ID
	}

	// Check that we have first items
	hasFirst := false
	for _, id := range ids[:min(2, len(ids))] {
		if id == "first" {
			hasFirst = true
		}
	}

	// Check that we have last items
	hasLast := false
	for _, id := range ids[max(0, len(ids)-2):] {
		if id == "last" {
			hasLast = true
		}
	}

	if !hasFirst {
		t.Error("Expected to preserve first items")
	}
	if !hasLast {
		t.Error("Expected to preserve last items")
	}

	t.Logf("TruncationMiddle: retained_ids=%v, tokens=%d", ids, result.OptimizedTokenCount)
}

func TestContextWindowManager_FilterByRelevance(t *testing.T) {
	config := ContextWindowConfig{
		MaxTokens:          1000,
		RelevanceThreshold: 0.5,
	}

	manager := NewContextWindowManager(config)

	items := []ContextItem{
		{ID: "high1", Content: "High", TokenCount: 100, Relevance: 0.9},
		{ID: "low1", Content: "Low", TokenCount: 100, Relevance: 0.2},
		{ID: "medium", Content: "Medium", TokenCount: 100, Relevance: 0.6},
		{ID: "low2", Content: "Low", TokenCount: 100, Relevance: 0.3},
		{ID: "high2", Content: "High", TokenCount: 100, Relevance: 0.8},
	}

	filtered := manager.FilterByRelevance(items)

	if len(filtered) != 3 {
		t.Errorf("Expected 3 items above threshold, got %d", len(filtered))
	}

	for _, item := range filtered {
		if item.Relevance < config.RelevanceThreshold {
			t.Errorf("Item %s has relevance %.2f below threshold %.2f",
				item.ID, item.Relevance, config.RelevanceThreshold)
		}
	}

	t.Logf("Filtered: %d/%d items (threshold=%.2f)", len(filtered), len(items), config.RelevanceThreshold)
}

func TestContextWindowManager_Stats(t *testing.T) {
	config := ContextWindowConfig{
		MaxTokens:        300,
		PriorityStrategy: PriorityRecency,
		TruncationMethod: TruncationHead,
		PreserveRecent:   2, // Only preserve 2 recent items
	}

	manager := NewContextWindowManager(config)

	// Perform multiple optimizations
	for range 5 {
		items := []ContextItem{
			{ID: "1", Content: "Item 1", TokenCount: 120, CreatedAt: time.Now(), Relevance: 0.8},
			{ID: "2", Content: "Item 2", TokenCount: 120, CreatedAt: time.Now().Add(-1 * time.Hour), Relevance: 0.7},
			{ID: "3", Content: "Item 3", TokenCount: 120, CreatedAt: time.Now().Add(-2 * time.Hour), Relevance: 0.6},
			{ID: "4", Content: "Item 4", TokenCount: 120, CreatedAt: time.Now().Add(-3 * time.Hour), Relevance: 0.5},
		}

		_, _, err := manager.OptimizeContext(context.Background(), items)
		if err != nil {
			t.Fatalf("OptimizeContext failed: %v", err)
		}
	}

	stats := manager.GetStats()

	if stats.TotalOptimizations != 5 {
		t.Errorf("Expected 5 optimizations, got %d", stats.TotalOptimizations)
	}

	if stats.OverflowsPrevented == 0 {
		t.Error("Expected overflows to be prevented")
	}

	if stats.TokensSaved <= 0 {
		t.Errorf("Expected positive tokens saved, got %d", stats.TokensSaved)
	}

	t.Logf("Stats: optimizations=%d, overflows_prevented=%d, tokens_saved=%d, avg_relevance=%.2f",
		stats.TotalOptimizations, stats.OverflowsPrevented, stats.TokensSaved, stats.AvgRelevanceGain)
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		minTokens int
		maxTokens int
	}{
		{"short", "Hello", 1, 2},
		{"medium", "This is a medium length sentence.", 6, 10},
		{"long", "This is a much longer piece of text that should result in a higher token count estimate.", 15, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := EstimateTokens(tt.text)
			if tokens < tt.minTokens || tokens > tt.maxTokens {
				t.Errorf("EstimateTokens(%q) = %d, expected between %d and %d",
					tt.text, tokens, tt.minTokens, tt.maxTokens)
			}
			t.Logf("Text length=%d, estimated_tokens=%d", len(tt.text), tokens)
		})
	}
}

func TestContextWindowManager_EmptyItems(t *testing.T) {
	config := ContextWindowConfig{
		MaxTokens: 1000,
	}

	manager := NewContextWindowManager(config)

	optimized, result, err := manager.OptimizeContext(context.Background(), []ContextItem{})

	if err != nil {
		t.Fatalf("OptimizeContext failed: %v", err)
	}

	if len(optimized) != 0 {
		t.Errorf("Expected 0 items, got %d", len(optimized))
	}

	if result.Method != "none" {
		t.Errorf("Expected method='none', got %s", result.Method)
	}

	t.Logf("Empty items result: method=%s, tokens=%d", result.Method, result.OptimizedTokenCount)
}

func TestContextWindowManager_PriorityImportance(t *testing.T) {
	config := ContextWindowConfig{
		MaxTokens:        400,
		PriorityStrategy: PriorityImportance,
		TruncationMethod: TruncationHead,
	}

	manager := NewContextWindowManager(config)

	now := time.Now()
	items := []ContextItem{
		{ID: "low_imp", Content: "Low importance", TokenCount: 150,
			CreatedAt: now, Relevance: 0.9, Importance: 2},
		{ID: "high_imp", Content: "High importance", TokenCount: 150,
			CreatedAt: now.Add(-1 * time.Hour), Relevance: 0.5, Importance: 9},
		{ID: "medium_imp", Content: "Medium importance", TokenCount: 150,
			CreatedAt: now.Add(-2 * time.Hour), Relevance: 0.7, Importance: 5},
	}

	optimized, result, err := manager.OptimizeContext(context.Background(), items)

	if err != nil {
		t.Fatalf("OptimizeContext failed: %v", err)
	}

	// Should keep high importance item first
	if len(optimized) > 0 && optimized[0].ID != "high_imp" {
		t.Errorf("Expected first item to be 'high_imp', got '%s'", optimized[0].ID)
	}

	t.Logf("PriorityImportance: retained=%d, first_id=%s, tokens=%d/%d",
		len(optimized), optimized[0].ID, result.OptimizedTokenCount, result.OriginalTokenCount)
}

// Helper functions.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func BenchmarkOptimizeContext_Small(b *testing.B) {
	config := ContextWindowConfig{
		MaxTokens:        500,
		PriorityStrategy: PriorityHybrid,
		TruncationMethod: TruncationHead,
	}

	manager := NewContextWindowManager(config)

	items := make([]ContextItem, 10)
	for i := range 10 {
		items[i] = ContextItem{
			ID:         string(rune('0' + i)),
			Content:    "Test content",
			TokenCount: 100,
			CreatedAt:  time.Now().Add(-time.Duration(i) * time.Hour),
			Relevance:  float64(10-i) / 10.0,
			Importance: 10 - i,
		}
	}

	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		_, _, _ = manager.OptimizeContext(ctx, items)
	}
}

func BenchmarkOptimizeContext_Large(b *testing.B) {
	config := ContextWindowConfig{
		MaxTokens:        5000,
		PriorityStrategy: PriorityHybrid,
		TruncationMethod: TruncationMiddle,
	}

	manager := NewContextWindowManager(config)

	items := make([]ContextItem, 100)
	for i := range 100 {
		items[i] = ContextItem{
			ID:         string(rune('0' + (i % 10))),
			Content:    "Test content with more text",
			TokenCount: 150,
			CreatedAt:  time.Now().Add(-time.Duration(i) * time.Hour),
			Relevance:  float64(100-i) / 100.0,
			Importance: (100 - i) / 10,
		}
	}

	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		_, _, _ = manager.OptimizeContext(ctx, items)
	}
}
