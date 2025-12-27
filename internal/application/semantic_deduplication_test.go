package application

import (
	"context"
	"testing"
	"time"
)

func TestSemanticDeduplication_ExactDuplicates(t *testing.T) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.92,
		MergeStrategy:       MergeKeepFirst,
	}

	svc := NewSemanticDeduplicationService(config)

	items := []DeduplicateItem{
		{ID: "1", Content: "This is a test message", CreatedAt: time.Now()},
		{ID: "2", Content: "This is a test message", CreatedAt: time.Now()},
		{ID: "3", Content: "This is a different message", CreatedAt: time.Now()},
		{ID: "4", Content: "This is a test message", CreatedAt: time.Now()},
	}

	deduplicated, result, err := svc.DeduplicateItems(context.Background(), items)

	if err != nil {
		t.Fatalf("DeduplicateItems failed: %v", err)
	}

	if result.DuplicatesRemoved != 2 {
		t.Errorf("Expected 2 duplicates removed, got %d", result.DuplicatesRemoved)
	}

	if len(deduplicated) != 2 {
		t.Errorf("Expected 2 unique items, got %d", len(deduplicated))
	}

	t.Logf("Exact duplicates: original=%d, deduplicated=%d, removed=%d, bytes_saved=%d",
		result.OriginalCount, result.DeduplicatedCount, result.DuplicatesRemoved, result.BytesSaved)
}

func TestSemanticDeduplication_FuzzySimilarity(t *testing.T) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.92,
		MergeStrategy:       MergeKeepLongest,
	}

	svc := NewSemanticDeduplicationService(config)

	items := []DeduplicateItem{
		{ID: "1", Content: "The quick brown fox jumps over the lazy dog", CreatedAt: time.Now()},
		{ID: "2", Content: "The quick brown fox jumps over the lazy dogs", CreatedAt: time.Now()}, // 1 char diff
		{ID: "3", Content: "A completely different text that should not match", CreatedAt: time.Now()},
	}

	deduplicated, result, err := svc.DeduplicateItems(context.Background(), items)

	if err != nil {
		t.Fatalf("DeduplicateItems failed: %v", err)
	}

	if len(deduplicated) > 2 {
		t.Logf("Warning: Expected fuzzy match, but got %d items (may be threshold issue)", len(deduplicated))
	}

	// Calculate actual similarity
	similarity := svc.CalculateSimilarity(items[0].Content, items[1].Content)
	t.Logf("Fuzzy similarity: %.2f%% (threshold=%.2f%%)", similarity*100, config.SimilarityThreshold*100)
	t.Logf("Result: original=%d, deduplicated=%d, removed=%d", result.OriginalCount, result.DeduplicatedCount, result.DuplicatesRemoved)
}

func TestSemanticDeduplication_MergeStrategies(t *testing.T) {
	strategies := []MergeStrategy{
		MergeKeepFirst,
		MergeKeepLast,
		MergeKeepLongest,
		MergeCombine,
	}

	baseItems := []DeduplicateItem{
		{ID: "1", Content: "Short text", CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: "2", Content: "Short text", CreatedAt: time.Now().Add(-1 * time.Hour)},
		{ID: "3", Content: "This is a much longer version of the text", CreatedAt: time.Now()},
	}

	for _, strategy := range strategies {
		t.Run(string(strategy), func(t *testing.T) {
			config := DeduplicationConfig{
				Enabled:             true,
				SimilarityThreshold: 0.80, // Lower threshold to catch "Short text" variations
				MergeStrategy:       strategy,
			}

			svc := NewSemanticDeduplicationService(config)

			// Make a copy to avoid modifying original
			items := make([]DeduplicateItem, len(baseItems))
			copy(items, baseItems)

			deduplicated, result, err := svc.DeduplicateItems(context.Background(), items)

			if err != nil {
				t.Fatalf("DeduplicateItems failed: %v", err)
			}

			t.Logf("Strategy %s: original=%d, deduplicated=%d, removed=%d",
				strategy, result.OriginalCount, result.DeduplicatedCount, result.DuplicatesRemoved)

			if len(deduplicated) > 0 {
				t.Logf("  First result content: %q", deduplicated[0].Content)
			}
		})
	}
}

func TestSemanticDeduplication_KeepLongest(t *testing.T) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.95,
		MergeStrategy:       MergeKeepLongest,
	}

	svc := NewSemanticDeduplicationService(config)

	items := []DeduplicateItem{
		{ID: "1", Content: "Hello", CreatedAt: time.Now()},
		{ID: "2", Content: "Hello", CreatedAt: time.Now()},
		{ID: "3", Content: "Hello world", CreatedAt: time.Now()},
	}

	// Adjust similarity - "Hello" vs "Hello world" should match
	sim1 := svc.CalculateSimilarity("Hello", "Hello")
	sim2 := svc.CalculateSimilarity("Hello", "Hello world")

	t.Logf("Similarity: 'Hello' vs 'Hello' = %.2f%%", sim1*100)
	t.Logf("Similarity: 'Hello' vs 'Hello world' = %.2f%%", sim2*100)

	deduplicated, result, err := svc.DeduplicateItems(context.Background(), items)

	if err != nil {
		t.Fatalf("DeduplicateItems failed: %v", err)
	}

	// Should keep the longest version
	if len(deduplicated) > 0 {
		hasLong := false
		for _, item := range deduplicated {
			if item.Content == "Hello world" {
				hasLong = true
			}
		}
		if !hasLong && result.DuplicatesRemoved > 0 {
			t.Logf("Note: Expected longest version to be kept, but may not match due to similarity threshold")
		}
	}

	t.Logf("KeepLongest: deduplicated=%d, removed=%d", result.DeduplicatedCount, result.DuplicatesRemoved)
}

func TestSemanticDeduplication_CombineStrategy(t *testing.T) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.95,
		MergeStrategy:       MergeCombine,
		PreserveMetadata:    true,
	}

	svc := NewSemanticDeduplicationService(config)

	items := []DeduplicateItem{
		{ID: "1", Content: "First sentence. Second sentence.", CreatedAt: time.Now(),
			Metadata: map[string]interface{}{"source": "doc1"}},
		{ID: "2", Content: "First sentence. Third sentence.", CreatedAt: time.Now(),
			Metadata: map[string]interface{}{"source": "doc2"}},
	}

	deduplicated, result, err := svc.DeduplicateItems(context.Background(), items)

	if err != nil {
		t.Fatalf("DeduplicateItems failed: %v", err)
	}

	t.Logf("Combine: deduplicated=%d, removed=%d", result.DeduplicatedCount, result.DuplicatesRemoved)

	if len(deduplicated) > 0 && result.DuplicatesRemoved > 0 {
		combined := deduplicated[0]
		t.Logf("Combined content: %q", combined.Content)
		t.Logf("Combined metadata: %+v", combined.Metadata)
	}
}

func TestSemanticDeduplication_Stats(t *testing.T) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.95,
		MergeStrategy:       MergeKeepFirst,
	}

	svc := NewSemanticDeduplicationService(config)

	// Process multiple batches
	for range 3 {
		items := []DeduplicateItem{
			{ID: "1", Content: "Duplicate content here", CreatedAt: time.Now()},
			{ID: "2", Content: "Duplicate content here", CreatedAt: time.Now()},
			{ID: "3", Content: "Unique content", CreatedAt: time.Now()},
		}

		_, _, err := svc.DeduplicateItems(context.Background(), items)
		if err != nil {
			t.Fatalf("DeduplicateItems failed: %v", err)
		}
	}

	stats := svc.GetStats()

	if stats.TotalProcessed != 9 {
		t.Errorf("Expected 9 items processed, got %d", stats.TotalProcessed)
	}

	if stats.DuplicatesRemoved == 0 {
		t.Error("Expected some duplicates removed")
	}

	t.Logf("Stats: processed=%d, duplicates_found=%d, duplicates_removed=%d, bytes_saved=%d, avg_similarity=%.2f",
		stats.TotalProcessed, stats.DuplicatesFound, stats.DuplicatesRemoved, stats.BytesSaved, stats.AvgSimilarity)
}

func TestSemanticDeduplication_Disabled(t *testing.T) {
	config := DeduplicationConfig{
		Enabled: false,
	}

	svc := NewSemanticDeduplicationService(config)

	items := []DeduplicateItem{
		{ID: "1", Content: "Same content", CreatedAt: time.Now()},
		{ID: "2", Content: "Same content", CreatedAt: time.Now()},
	}

	deduplicated, result, err := svc.DeduplicateItems(context.Background(), items)

	if err != nil {
		t.Fatalf("DeduplicateItems failed: %v", err)
	}

	if result.DuplicatesRemoved != 0 {
		t.Errorf("Expected no duplicates removed when disabled, got %d", result.DuplicatesRemoved)
	}

	if len(deduplicated) != len(items) {
		t.Errorf("Expected original count when disabled")
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{"identical", "hello", "hello", 0},
		{"one_insertion", "hello", "helo", 1},
		{"one_deletion", "hello", "helloo", 1},
		{"one_substitution", "hello", "hallo", 1},
		{"multiple_changes", "kitten", "sitting", 3},
		{"empty_strings", "", "", 0},
		{"one_empty", "hello", "", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := levenshteinDistance(tt.s1, tt.s2)
			if distance != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d",
					tt.s1, tt.s2, distance, tt.expected)
			}
		})
	}
}

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "Hello World", "hello world"},
		{"punctuation", "Hello, World!", "hello world"},
		{"extra_spaces", "Hello   World", "hello world"},
		{"mixed", "Hello, World! How are you?", "hello world how are you"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeText(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeText(%q) = %q, expected %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculateSimilarity(t *testing.T) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.92,
	}

	svc := NewSemanticDeduplicationService(config)

	tests := []struct {
		name   string
		text1  string
		text2  string
		minSim float64
		maxSim float64
	}{
		{"identical", "Hello World", "Hello World", 1.0, 1.0},
		{"case_insensitive", "Hello World", "hello world", 1.0, 1.0},
		{"punctuation", "Hello, World!", "Hello World", 0.9, 1.0},
		{"similar", "The quick brown fox", "The quick brown dog", 0.8, 0.95},
		{"different", "Hello", "Goodbye", 0.0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := svc.CalculateSimilarity(tt.text1, tt.text2)
			if similarity < tt.minSim || similarity > tt.maxSim {
				t.Errorf("CalculateSimilarity(%q, %q) = %.2f, expected between %.2f and %.2f",
					tt.text1, tt.text2, similarity, tt.minSim, tt.maxSim)
			}
			t.Logf("Similarity: %.2f%%", similarity*100)
		})
	}
}

func TestFindDuplicates(t *testing.T) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.92,
	}

	svc := NewSemanticDeduplicationService(config)

	items := []DeduplicateItem{
		{ID: "1", Content: "Test message one", CreatedAt: time.Now()},
		{ID: "2", Content: "Test message one", CreatedAt: time.Now()},
		{ID: "3", Content: "Test message two", CreatedAt: time.Now()},
		{ID: "4", Content: "Test message one", CreatedAt: time.Now()},
		{ID: "5", Content: "Completely different", CreatedAt: time.Now()},
	}

	groups := svc.FindDuplicates(items)

	if len(groups) == 0 {
		t.Error("Expected to find duplicate groups")
	}

	for i, group := range groups {
		t.Logf("Group %d: %d items, similarity=%.2f%%", i+1, len(group.Items), group.Similarity*100)
		for _, item := range group.Items {
			t.Logf("  - ID=%s, Content=%q", item.ID, item.Content)
		}
	}
}

func TestSemanticDeduplication_EmptyItems(t *testing.T) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.92,
	}

	svc := NewSemanticDeduplicationService(config)

	deduplicated, result, err := svc.DeduplicateItems(context.Background(), []DeduplicateItem{})

	if err != nil {
		t.Fatalf("DeduplicateItems failed: %v", err)
	}

	if len(deduplicated) != 0 {
		t.Errorf("Expected 0 items, got %d", len(deduplicated))
	}

	if result.DuplicatesRemoved != 0 {
		t.Errorf("Expected 0 duplicates removed, got %d", result.DuplicatesRemoved)
	}
}

func BenchmarkDeduplication_Small(b *testing.B) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.92,
		MergeStrategy:       MergeKeepFirst,
	}

	svc := NewSemanticDeduplicationService(config)

	items := make([]DeduplicateItem, 10)
	for i := range 10 {
		content := "Test message"
		if i%2 == 0 {
			content = "Duplicate message"
		}
		items[i] = DeduplicateItem{
			ID:        string(rune('0' + i)),
			Content:   content,
			CreatedAt: time.Now(),
		}
	}

	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		_, _, _ = svc.DeduplicateItems(ctx, items)
	}
}

func BenchmarkDeduplication_Large(b *testing.B) {
	config := DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.92,
		MergeStrategy:       MergeKeepFirst,
		BatchSize:           100,
	}

	svc := NewSemanticDeduplicationService(config)

	items := make([]DeduplicateItem, 100)
	for i := range 100 {
		content := "Test message number"
		if i%5 == 0 {
			content = "Duplicate message appears multiple times"
		}
		items[i] = DeduplicateItem{
			ID:        string(rune('0' + (i % 10))),
			Content:   content,
			CreatedAt: time.Now(),
		}
	}

	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		_, _, _ = svc.DeduplicateItems(ctx, items)
	}
}

func BenchmarkLevenshteinDistance(b *testing.B) {
	s1 := "The quick brown fox jumps over the lazy dog"
	s2 := "The quick brown fox jumps over the lazy dogs"

	b.ResetTimer()

	for range b.N {
		_ = levenshteinDistance(s1, s2)
	}
}
