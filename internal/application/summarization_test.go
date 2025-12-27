package application

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSummarizationService_Extractive(t *testing.T) {
	config := SummarizationConfig{
		Enabled:              true,
		AgeBeforeSummarize:   1 * time.Hour,
		MaxSummaryLength:     200,
		CompressionRatio:     0.3,
		PreserveKeywords:     true,
		UseExtractiveSummary: true,
	}

	svc := NewSummarizationService(config)

	tests := []struct {
		name                string
		content             string
		age                 time.Duration
		expectSummarization bool
		maxExpectedLength   int
	}{
		{
			name: "long_technical_content",
			content: `The Go programming language was designed at Google in 2007 by Robert Griesemer, Rob Pike, and Ken Thompson. ` +
				`Go is statically typed and compiled. It provides built-in support for concurrent programming. ` +
				`The language has a simple syntax that is easy to learn. Go programs compile to native machine code. ` +
				`Memory management is automatic through garbage collection. The standard library is comprehensive. ` +
				`Go is widely used for cloud infrastructure, microservices, and DevOps tools. ` +
				`Many companies including Google, Uber, and Netflix use Go in production.`,
			age:                 2 * time.Hour,
			expectSummarization: true,
			maxExpectedLength:   250,
		},
		{
			name:                "short_content_no_summarization",
			content:             `Go is a programming language. It was created by Google.`,
			age:                 2 * time.Hour,
			expectSummarization: false,
			maxExpectedLength:   100,
		},
		{
			name:                "recent_content_no_summarization",
			content:             strings.Repeat("This is a very long piece of content that should not be summarized because it is too recent. ", 10),
			age:                 30 * time.Minute,
			expectSummarization: false,
			maxExpectedLength:   1000,
		},
		{
			name: "multiple_sentences_with_keywords",
			content: `Machine learning is transforming software development. Deep learning models can understand natural language. ` +
				`Neural networks require significant computational resources. Training large models takes days or weeks. ` +
				`Transfer learning allows using pre-trained models. Fine-tuning adapts models to specific tasks. ` +
				`Model deployment requires optimization for inference. Edge devices benefit from quantized models. ` +
				`The field is rapidly evolving with new architectures. Research papers are published daily on arXiv.`,
			age:                 24 * time.Hour,
			expectSummarization: true,
			maxExpectedLength:   220,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdAt := time.Now().Add(-tt.age)
			summary, metadata, err := svc.SummarizeText(context.Background(), tt.content, createdAt)

			if err != nil {
				t.Fatalf("SummarizeText failed: %v", err)
			}

			if tt.expectSummarization {
				if metadata.Method == "none" {
					t.Errorf("Expected summarization but method=none")
				}
				if len(summary) >= len(tt.content) {
					t.Errorf("Expected summary (%d) to be shorter than original (%d)", len(summary), len(tt.content))
				}
				if len(summary) > tt.maxExpectedLength {
					t.Errorf("Summary length %d exceeds max expected %d", len(summary), tt.maxExpectedLength)
				}
				t.Logf("Compression ratio: %.2f%% (original=%d, summary=%d, method=%s)",
					metadata.CompressionRatio*100, metadata.OriginalLength, metadata.SummaryLength, metadata.Method)

				if len(metadata.KeywordsPreserved) == 0 {
					t.Logf("Warning: No keywords preserved")
				} else {
					t.Logf("Keywords: %v", metadata.KeywordsPreserved)
				}
			} else {
				if metadata.Method != "none" {
					t.Errorf("Expected no summarization but method=%s", metadata.Method)
				}
				if summary != tt.content {
					t.Errorf("Expected original content when not summarizing")
				}
			}
		})
	}
}

func TestSummarizationService_Truncation(t *testing.T) {
	config := SummarizationConfig{
		Enabled:              true,
		AgeBeforeSummarize:   1 * time.Hour,
		MaxSummaryLength:     150,
		CompressionRatio:     0.4,
		UseExtractiveSummary: false, // Use truncation
	}

	svc := NewSummarizationService(config)

	content := `First sentence introduces the topic. Second sentence provides context. ` +
		`Third sentence gives details. Fourth sentence expands on details. ` +
		`Fifth sentence adds more information. Sixth sentence concludes the paragraph.`

	createdAt := time.Now().Add(-2 * time.Hour)
	summary, metadata, err := svc.SummarizeText(context.Background(), content, createdAt)

	if err != nil {
		t.Fatalf("SummarizeText failed: %v", err)
	}

	if metadata.Method != "truncation" {
		t.Errorf("Expected truncation method, got %s", metadata.Method)
	}

	if len(summary) >= len(content) {
		t.Errorf("Expected summary to be shorter than original")
	}

	// Should contain first and last sentences
	if !strings.Contains(summary, "First sentence") {
		t.Errorf("Expected first sentence in summary")
	}
	if !strings.Contains(summary, "concludes") {
		t.Errorf("Expected last sentence in summary")
	}

	t.Logf("Truncation summary: %s", summary)
	t.Logf("Compression ratio: %.2f%%", metadata.CompressionRatio*100)
}

func TestSummarizationService_Stats(t *testing.T) {
	config := SummarizationConfig{
		Enabled:              true,
		AgeBeforeSummarize:   1 * time.Hour,
		MaxSummaryLength:     100,
		CompressionRatio:     0.3,
		UseExtractiveSummary: true,
	}

	svc := NewSummarizationService(config)

	// Summarize multiple texts
	texts := []string{
		strings.Repeat("Test content number one with enough text to be summarized properly. ", 5),
		strings.Repeat("Different test content for the second summarization attempt. ", 5),
		strings.Repeat("Third piece of content that will also be summarized. ", 5),
	}

	createdAt := time.Now().Add(-2 * time.Hour)

	for _, text := range texts {
		_, _, err := svc.SummarizeText(context.Background(), text, createdAt)
		if err != nil {
			t.Fatalf("SummarizeText failed: %v", err)
		}
	}

	stats := svc.GetStats()

	if stats.TotalSummarized != 3 {
		t.Errorf("Expected 3 summarizations, got %d", stats.TotalSummarized)
	}

	if stats.BytesSaved <= 0 {
		t.Errorf("Expected positive bytes saved, got %d", stats.BytesSaved)
	}

	if stats.AvgCompressionRatio <= 0 || stats.AvgCompressionRatio >= 1 {
		t.Errorf("Expected compression ratio between 0 and 1, got %.2f", stats.AvgCompressionRatio)
	}

	t.Logf("Stats: summarized=%d, bytes_saved=%d, avg_ratio=%.2f",
		stats.TotalSummarized, stats.BytesSaved, stats.AvgCompressionRatio)
}

func TestSummarizationService_ShouldSummarize(t *testing.T) {
	config := SummarizationConfig{
		Enabled:            true,
		AgeBeforeSummarize: 24 * time.Hour,
		MaxSummaryLength:   200,
	}

	svc := NewSummarizationService(config)

	tests := []struct {
		name          string
		contentLength int
		age           time.Duration
		expected      bool
	}{
		{"old_long_content", 1000, 48 * time.Hour, true},
		{"old_short_content", 100, 48 * time.Hour, false},
		{"recent_long_content", 1000, 12 * time.Hour, false},
		{"recent_short_content", 100, 12 * time.Hour, false},
		{"exact_threshold_age", 500, 24 * time.Hour, true},
		{"exact_threshold_length", 200, 48 * time.Hour, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.ShouldSummarize(tt.contentLength, tt.age)
			if result != tt.expected {
				t.Errorf("ShouldSummarize(%d, %v) = %v, expected %v",
					tt.contentLength, tt.age, result, tt.expected)
			}
		})
	}
}

func TestSummarizationService_EstimateSavings(t *testing.T) {
	config := SummarizationConfig{
		Enabled:          true,
		MaxSummaryLength: 200,
		CompressionRatio: 0.3,
	}

	svc := NewSummarizationService(config)

	tests := []struct {
		name          string
		contentLength int
		minSavings    int
	}{
		{"very_long_content", 2000, 1000},
		{"medium_content", 500, 100},
		{"short_content", 150, 0}, // Below threshold, no savings
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savings := svc.EstimateSavings(tt.contentLength)
			if savings < tt.minSavings {
				t.Errorf("EstimateSavings(%d) = %d, expected at least %d",
					tt.contentLength, savings, tt.minSavings)
			}
			t.Logf("Content length %d -> estimated savings %d bytes", tt.contentLength, savings)
		})
	}
}

func TestExtractKeywords(t *testing.T) {
	config := SummarizationConfig{
		Enabled:              true,
		UseExtractiveSummary: true,
	}

	svc := NewSummarizationService(config)

	content := `Machine learning algorithms can process vast amounts of data. ` +
		`Neural networks are particularly effective for pattern recognition. ` +
		`Training neural networks requires significant computational resources. ` +
		`Deep learning techniques have revolutionized computer vision and natural language processing.`

	keywords := svc.extractKeywords(content, 5)

	if len(keywords) == 0 {
		t.Error("Expected non-empty keywords list")
	}

	t.Logf("Extracted keywords: %v", keywords)

	// Check that keywords are meaningful (not stop words)
	for _, kw := range keywords {
		if isStopWord(kw) {
			t.Errorf("Keyword '%s' is a stop word", kw)
		}
	}

	// Check for expected technical terms
	expectedTerms := []string{"neural", "learning", "networks", "data"}
	foundCount := 0
	for _, expected := range expectedTerms {
		for _, kw := range keywords {
			if strings.Contains(kw, expected) {
				foundCount++
				break
			}
		}
	}

	if foundCount < 2 {
		t.Errorf("Expected at least 2 technical terms in keywords, found %d", foundCount)
	}
}

func TestSplitSentences(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"simple", "First. Second. Third.", 3},
		{"with_question", "What is this? It is a test. Done.", 3},
		{"with_exclamation", "Hello! How are you? I am fine. Thanks!", 4},
		{"short_fragments", "Hi. No. Ok.", 3}, // All short but valid
		{"mixed_length", "This is a proper sentence. No. Another good sentence here.", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentences := splitSentences(tt.text)
			if len(sentences) != tt.expected {
				t.Errorf("splitSentences(%q) returned %d sentences, expected %d: %v",
					tt.text, len(sentences), tt.expected, sentences)
			}
		})
	}
}

func TestIsStopWord(t *testing.T) {
	stopWords := []string{"the", "be", "to", "of", "and", "a"}
	contentWords := []string{"algorithm", "data", "network", "learning", "system"}

	for _, word := range stopWords {
		if !isStopWord(word) {
			t.Errorf("Expected '%s' to be a stop word", word)
		}
	}

	for _, word := range contentWords {
		if isStopWord(word) {
			t.Errorf("Expected '%s' to NOT be a stop word", word)
		}
	}
}

func TestSummarizationService_Disabled(t *testing.T) {
	config := SummarizationConfig{
		Enabled: false, // Disabled
	}

	svc := NewSummarizationService(config)

	content := strings.Repeat("This content should not be summarized because the service is disabled. ", 20)
	createdAt := time.Now().Add(-48 * time.Hour)

	summary, metadata, err := svc.SummarizeText(context.Background(), content, createdAt)

	if err != nil {
		t.Fatalf("SummarizeText failed: %v", err)
	}

	if summary != content {
		t.Errorf("Expected original content when disabled")
	}

	if metadata.Method != "none" {
		t.Errorf("Expected method='none' when disabled, got %s", metadata.Method)
	}

	if metadata.CompressionRatio != 1.0 {
		t.Errorf("Expected compression ratio 1.0 when disabled, got %.2f", metadata.CompressionRatio)
	}
}

func BenchmarkSummarization_Extractive(b *testing.B) {
	config := SummarizationConfig{
		Enabled:              true,
		AgeBeforeSummarize:   1 * time.Hour,
		MaxSummaryLength:     300,
		UseExtractiveSummary: true,
	}

	svc := NewSummarizationService(config)

	content := strings.Repeat(`The field of artificial intelligence is advancing rapidly. `+
		`Machine learning models are becoming more sophisticated. `+
		`Natural language processing enables human-computer interaction. `+
		`Computer vision systems can recognize objects and faces. `+
		`Reinforcement learning allows agents to learn from experience. `, 10)

	createdAt := time.Now().Add(-2 * time.Hour)
	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		_, _, _ = svc.SummarizeText(ctx, content, createdAt)
	}
}

func BenchmarkSummarization_Truncation(b *testing.B) {
	config := SummarizationConfig{
		Enabled:              true,
		AgeBeforeSummarize:   1 * time.Hour,
		MaxSummaryLength:     300,
		UseExtractiveSummary: false,
	}

	svc := NewSummarizationService(config)

	content := strings.Repeat(`Test sentence for benchmarking truncation performance. `, 50)
	createdAt := time.Now().Add(-2 * time.Hour)
	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		_, _, _ = svc.SummarizeText(ctx, content, createdAt)
	}
}
