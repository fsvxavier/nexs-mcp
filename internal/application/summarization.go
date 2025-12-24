package application

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
)

// SummarizationService handles automatic content summarization.
type SummarizationService struct {
	config SummarizationConfig
	stats  SummarizationStats
	mu     sync.RWMutex
}

// SummarizationConfig configures summarization behavior.
type SummarizationConfig struct {
	Enabled              bool
	AgeBeforeSummarize   time.Duration // Default: 7 days
	MaxSummaryLength     int           // Default: 500 chars
	CompressionRatio     float64       // Target: 0.3 (70% reduction)
	PreserveKeywords     bool          // Preserve extracted keywords
	UseExtractiveSummary bool          // Extract key sentences vs abstractive
}

// SummarizationStats tracks summarization metrics.
type SummarizationStats struct {
	TotalSummarized     int64
	BytesSaved          int64
	AvgCompressionRatio float64
	QualityScore        float64
}

// SummarizationMetadata describes summarization results.
type SummarizationMetadata struct {
	OriginalLength    int      `json:"original_length"`
	SummaryLength     int      `json:"summary_length"`
	CompressionRatio  float64  `json:"compression_ratio"`
	Method            string   `json:"method"` // "extractive", "truncation", "none"
	KeywordsPreserved []string `json:"keywords_preserved,omitempty"`
}

// NewSummarizationService creates a new summarization service.
func NewSummarizationService(config SummarizationConfig) *SummarizationService {
	// Set defaults
	if config.MaxSummaryLength == 0 {
		config.MaxSummaryLength = 500
	}
	if config.CompressionRatio == 0 {
		config.CompressionRatio = 0.3
	}
	if config.AgeBeforeSummarize == 0 {
		config.AgeBeforeSummarize = 7 * 24 * time.Hour
	}

	return &SummarizationService{
		config: config,
		stats: SummarizationStats{
			AvgCompressionRatio: 1.0,
			QualityScore:        1.0,
		},
	}
}

// SummarizeText creates a concise summary of text content.
func (s *SummarizationService) SummarizeText(ctx context.Context, content string, createdAt time.Time) (string, SummarizationMetadata, error) {
	if !s.config.Enabled {
		return content, SummarizationMetadata{
			OriginalLength:   len(content),
			SummaryLength:    len(content),
			CompressionRatio: 1.0,
			Method:           "none",
		}, nil
	}

	// Check if content is old enough to summarize
	age := time.Since(createdAt)
	if age < s.config.AgeBeforeSummarize {
		return content, SummarizationMetadata{
			OriginalLength:   len(content),
			SummaryLength:    len(content),
			CompressionRatio: 1.0,
			Method:           "none",
		}, nil
	}

	originalLength := len(content)

	// Don't summarize if already short enough
	if originalLength <= s.config.MaxSummaryLength {
		return content, SummarizationMetadata{
			OriginalLength:   originalLength,
			SummaryLength:    originalLength,
			CompressionRatio: 1.0,
			Method:           "none",
		}, nil
	}

	// Choose summarization method
	var summary string
	var method string
	var keywords []string

	if s.config.UseExtractiveSummary {
		summary, keywords = s.extractiveSummarize(content)
		method = "extractive"
	} else {
		summary = s.truncateSummary(content)
		method = "truncation"
	}

	summaryLength := len(summary)
	compressionRatio := float64(summaryLength) / float64(originalLength)

	// Update stats
	s.updateStats(originalLength, summaryLength, compressionRatio)

	return summary, SummarizationMetadata{
		OriginalLength:    originalLength,
		SummaryLength:     summaryLength,
		CompressionRatio:  compressionRatio,
		Method:            method,
		KeywordsPreserved: keywords,
	}, nil
}

// extractiveSummarize extracts key sentences using TF-IDF scoring.
func (s *SummarizationService) extractiveSummarize(content string) (string, []string) {
	sentences := splitSentences(content)
	if len(sentences) <= 3 {
		return content, nil // Too short to summarize
	}

	// Extract keywords
	keywords := s.extractKeywords(content, 10)

	// Score sentences by keyword density and position
	scores := make(map[int]float64)

	for i, sentence := range sentences {
		score := 0.0

		// Keyword density score
		sentenceLower := strings.ToLower(sentence)
		for _, keyword := range keywords {
			if strings.Contains(sentenceLower, strings.ToLower(keyword)) {
				score += 1.0
			}
		}

		// Position bias (first and last sentences are important)
		if i == 0 {
			score += 2.0 // First sentence is very important
		} else if i == len(sentences)-1 {
			score += 1.0 // Last sentence is important
		}

		// Length bias (prefer medium-length sentences)
		wordCount := len(strings.Fields(sentence))
		if wordCount >= 10 && wordCount <= 30 {
			score += 0.5
		}

		scores[i] = score
	}

	// Select top sentences
	selectedIndices := s.selectTopSentences(scores, s.config.MaxSummaryLength, sentences)
	sort.Ints(selectedIndices) // Maintain original order

	// Build summary
	var summaryParts []string
	for _, idx := range selectedIndices {
		summaryParts = append(summaryParts, sentences[idx])
	}

	summary := strings.Join(summaryParts, " ")

	// If still too long, truncate
	if len(summary) > s.config.MaxSummaryLength {
		summary = summary[:s.config.MaxSummaryLength] + "..."
	}

	return summary, keywords
}

// selectTopSentences selects sentences with highest scores up to max length.
func (s *SummarizationService) selectTopSentences(scores map[int]float64, maxLength int, sentences []string) []int {
	type scoredIndex struct {
		index int
		score float64
	}

	// Create sorted list of sentence indices by score
	scored := make([]scoredIndex, 0, len(scores))
	for idx, score := range scores {
		scored = append(scored, scoredIndex{idx, score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Select sentences until we reach max length
	selected := []int{}
	currentLength := 0

	for _, item := range scored {
		sentenceLen := len(sentences[item.index])
		if currentLength+sentenceLen <= maxLength {
			selected = append(selected, item.index)
			currentLength += sentenceLen + 1 // +1 for space
		}
		if currentLength >= maxLength {
			break
		}
	}

	// Ensure we have at least first sentence if nothing selected
	if len(selected) == 0 && len(sentences) > 0 {
		selected = append(selected, 0)
	}

	return selected
}

// truncateSummary creates a simple summary (first + last sentences).
func (s *SummarizationService) truncateSummary(content string) string {
	sentences := splitSentences(content)
	if len(sentences) <= 2 {
		return content
	}

	// Take first 2 and last sentence
	summary := sentences[0]
	if len(sentences) > 1 {
		summary += " " + sentences[1]
	}
	if len(sentences) > 3 {
		summary += " ... " + sentences[len(sentences)-1]
	}

	if len(summary) > s.config.MaxSummaryLength {
		summary = summary[:s.config.MaxSummaryLength] + "..."
	}

	return summary
}

// extractKeywords extracts important keywords using simple frequency analysis.
func (s *SummarizationService) extractKeywords(content string, maxKeywords int) []string {
	// Tokenize
	words := strings.FieldsFunc(content, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	// Count frequency (excluding common stop words)
	frequency := make(map[string]int)
	for _, word := range words {
		word = strings.ToLower(word)
		if len(word) > 3 && !isStopWord(word) { // Min 4 chars
			frequency[word]++
		}
	}

	// Sort by frequency
	type wordFreq struct {
		word  string
		count int
	}

	freqs := make([]wordFreq, 0, len(frequency))
	for word, count := range frequency {
		freqs = append(freqs, wordFreq{word, count})
	}

	sort.Slice(freqs, func(i, j int) bool {
		return freqs[i].count > freqs[j].count
	})

	// Take top N
	keywords := make([]string, 0, maxKeywords)
	for i := 0; i < len(freqs) && i < maxKeywords; i++ {
		keywords = append(keywords, freqs[i].word)
	}

	return keywords
}

// splitSentences splits text into sentences.
func splitSentences(text string) []string {
	// Simple sentence splitter (can be improved with NLP)
	text = strings.ReplaceAll(text, "? ", "?\n")
	text = strings.ReplaceAll(text, "! ", "!\n")
	text = strings.ReplaceAll(text, ". ", ".\n")

	// Handle end of text
	text = strings.ReplaceAll(text, "?", "?\n")
	text = strings.ReplaceAll(text, "!", "!\n")

	// Handle period at end but not duplicating if already has newline
	if !strings.HasSuffix(text, "\n") && strings.HasSuffix(text, ".") {
		text += "\n"
	}

	sentences := strings.Split(text, "\n")
	result := make([]string, 0, len(sentences))

	for _, s := range sentences {
		trimmed := strings.TrimSpace(s)
		// Accept sentences with at least 3 chars
		if len(trimmed) >= 3 {
			result = append(result, trimmed)
		}
	}

	return result
}

// isStopWord checks if a word is a common stop word.
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "be": true, "to": true, "of": true, "and": true,
		"a": true, "in": true, "that": true, "have": true, "i": true,
		"it": true, "for": true, "not": true, "on": true, "with": true,
		"he": true, "as": true, "you": true, "do": true, "at": true,
		"this": true, "but": true, "his": true, "by": true, "from": true,
		"they": true, "we": true, "say": true, "her": true, "she": true,
		"or": true, "an": true, "will": true, "my": true, "one": true,
		"all": true, "would": true, "there": true, "their": true,
	}
	return stopWords[word]
}

// updateStats updates summarization statistics.
func (s *SummarizationService) updateStats(originalLength, summaryLength int, compressionRatio float64) {
	atomic.AddInt64(&s.stats.TotalSummarized, 1)
	atomic.AddInt64(&s.stats.BytesSaved, int64(originalLength-summaryLength))

	// Update average compression ratio using exponential moving average
	s.mu.Lock()
	alpha := 0.1 // Smoothing factor
	s.stats.AvgCompressionRatio = alpha*compressionRatio + (1-alpha)*s.stats.AvgCompressionRatio
	s.mu.Unlock()
}

// GetStats returns current summarization statistics.
func (s *SummarizationService) GetStats() SummarizationStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return SummarizationStats{
		TotalSummarized:     atomic.LoadInt64(&s.stats.TotalSummarized),
		BytesSaved:          atomic.LoadInt64(&s.stats.BytesSaved),
		AvgCompressionRatio: s.stats.AvgCompressionRatio,
		QualityScore:        s.stats.QualityScore,
	}
}

// CalculateCompressionRatio calculates the target compression based on content length.
func (s *SummarizationService) CalculateCompressionRatio(contentLength int) float64 {
	// Adaptive compression: longer content gets higher compression
	switch {
	case contentLength < 500:
		return 0.8 // 20% reduction for short content
	case contentLength < 2000:
		return 0.5 // 50% reduction for medium content
	default:
		return s.config.CompressionRatio // 70% reduction for long content
	}
}

// ShouldSummarize determines if content should be summarized based on age and length.
func (s *SummarizationService) ShouldSummarize(contentLength int, age time.Duration) bool {
	if !s.config.Enabled {
		return false
	}

	// Check age threshold
	if age < s.config.AgeBeforeSummarize {
		return false
	}

	// Only summarize if content is long enough to benefit
	return contentLength > s.config.MaxSummaryLength
}

// EstimateSavings estimates bytes saved from summarization.
func (s *SummarizationService) EstimateSavings(contentLength int) int {
	if !s.config.Enabled || contentLength <= s.config.MaxSummaryLength {
		return 0
	}

	targetLength := int(math.Min(float64(s.config.MaxSummaryLength), float64(contentLength)*s.config.CompressionRatio))
	return contentLength - targetLength
}
