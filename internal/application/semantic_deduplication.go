package application

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
)

// SemanticDeduplicationService handles duplicate detection and merging.
type SemanticDeduplicationService struct {
	config DeduplicationConfig
	stats  DeduplicationStats
	mu     sync.RWMutex
}

// DeduplicationConfig configures deduplication behavior.
type DeduplicationConfig struct {
	Enabled             bool
	SimilarityThreshold float64       // Default: 0.92 (92% similarity)
	MergeStrategy       MergeStrategy // How to merge duplicates
	PreserveMetadata    bool          // Keep metadata from all duplicates
	BatchSize           int           // Process in batches
}

// MergeStrategy defines how to merge duplicate items.
type MergeStrategy string

const (
	MergeKeepFirst   MergeStrategy = "keep_first"   // Keep first occurrence
	MergeKeepLast    MergeStrategy = "keep_last"    // Keep last occurrence
	MergeKeepLongest MergeStrategy = "keep_longest" // Keep longest content
	MergeCombine     MergeStrategy = "combine"      // Merge all content
)

// DeduplicationStats tracks deduplication metrics.
type DeduplicationStats struct {
	TotalProcessed    int64
	DuplicatesFound   int64
	DuplicatesRemoved int64
	BytesSaved        int64
	AvgSimilarity     float64
}

// DuplicateGroup represents a group of similar items.
type DuplicateGroup struct {
	Items      []DeduplicateItem
	Similarity float64
	MergedItem *DeduplicateItem
}

// DeduplicateItem represents an item to deduplicate.
type DeduplicateItem struct {
	ID          string
	Content     string
	CreatedAt   time.Time
	Metadata    map[string]interface{}
	Fingerprint string // Hash or fingerprint for quick comparison
}

// DeduplicationResult describes the outcome of deduplication.
type DeduplicationResult struct {
	OriginalCount     int
	DeduplicatedCount int
	DuplicatesRemoved int
	BytesSaved        int
	Groups            []DuplicateGroup
}

// NewSemanticDeduplicationService creates a new deduplication service.
func NewSemanticDeduplicationService(config DeduplicationConfig) *SemanticDeduplicationService {
	// Set defaults
	if config.SimilarityThreshold == 0 {
		config.SimilarityThreshold = 0.92
	}
	if config.MergeStrategy == "" {
		config.MergeStrategy = MergeKeepFirst
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}

	return &SemanticDeduplicationService{
		config: config,
		stats: DeduplicationStats{
			AvgSimilarity: 0.0,
		},
	}
}

// DeduplicateItems finds and merges duplicate items.
func (s *SemanticDeduplicationService) DeduplicateItems(ctx context.Context, items []DeduplicateItem) ([]DeduplicateItem, DeduplicationResult, error) {
	if !s.config.Enabled || len(items) == 0 {
		return items, DeduplicationResult{
			OriginalCount:     len(items),
			DeduplicatedCount: len(items),
			DuplicatesRemoved: 0,
			BytesSaved:        0,
			Groups:            []DuplicateGroup{},
		}, nil
	}

	originalCount := len(items)
	originalBytes := s.calculateTotalBytes(items)

	// Find duplicate groups
	groups := s.findDuplicateGroups(items)

	// Merge duplicates according to strategy
	deduplicated := s.mergeDuplicateGroups(items, groups)

	deduplicatedBytes := s.calculateTotalBytes(deduplicated)
	bytesSaved := originalBytes - deduplicatedBytes

	result := DeduplicationResult{
		OriginalCount:     originalCount,
		DeduplicatedCount: len(deduplicated),
		DuplicatesRemoved: originalCount - len(deduplicated),
		BytesSaved:        bytesSaved,
		Groups:            groups,
	}

	// Update stats
	s.updateStats(result)

	return deduplicated, result, nil
}

// findDuplicateGroups identifies groups of similar items.
func (s *SemanticDeduplicationService) findDuplicateGroups(items []DeduplicateItem) []DuplicateGroup {
	groups := []DuplicateGroup{}
	processed := make(map[string]bool)

	for i := range items {
		if processed[items[i].ID] {
			continue
		}

		group := []DeduplicateItem{items[i]}
		similarities := []float64{}

		for j := i + 1; j < len(items); j++ {
			if processed[items[j].ID] {
				continue
			}

			similarity := s.calculateSimilarity(items[i].Content, items[j].Content)
			if similarity >= s.config.SimilarityThreshold {
				group = append(group, items[j])
				similarities = append(similarities, similarity)
				processed[items[j].ID] = true
			}
		}

		if len(group) > 1 {
			avgSimilarity := 0.0
			if len(similarities) > 0 {
				for _, sim := range similarities {
					avgSimilarity += sim
				}
				avgSimilarity /= float64(len(similarities))
			}

			groups = append(groups, DuplicateGroup{
				Items:      group,
				Similarity: avgSimilarity,
			})
		}

		processed[items[i].ID] = true
	}

	return groups
}

// mergeDuplicateGroups merges duplicate groups according to strategy.
func (s *SemanticDeduplicationService) mergeDuplicateGroups(items []DeduplicateItem, groups []DuplicateGroup) []DeduplicateItem {
	// Create a set of IDs to remove
	toRemove := make(map[string]bool)
	merged := make(map[string]DeduplicateItem)

	for _, group := range groups {
		mergedItem := s.mergeGroup(group)
		merged[mergedItem.ID] = mergedItem

		// Mark others in group for removal
		for i := 1; i < len(group.Items); i++ {
			toRemove[group.Items[i].ID] = true
		}
	}

	// Build deduplicated list
	result := []DeduplicateItem{}
	for _, item := range items {
		if toRemove[item.ID] {
			continue
		}

		// Use merged version if available
		if mergedItem, exists := merged[item.ID]; exists {
			result = append(result, mergedItem)
		} else {
			result = append(result, item)
		}
	}

	return result
}

// mergeGroup merges items in a duplicate group according to strategy.
func (s *SemanticDeduplicationService) mergeGroup(group DuplicateGroup) DeduplicateItem {
	if len(group.Items) == 0 {
		return DeduplicateItem{}
	}

	switch s.config.MergeStrategy {
	case MergeKeepFirst:
		return group.Items[0]

	case MergeKeepLast:
		return group.Items[len(group.Items)-1]

	case MergeKeepLongest:
		longest := group.Items[0]
		for _, item := range group.Items {
			if len(item.Content) > len(longest.Content) {
				longest = item
			}
		}
		return longest

	case MergeCombine:
		return s.combineItems(group.Items)

	default:
		return group.Items[0]
	}
}

// combineItems combines multiple items into one.
func (s *SemanticDeduplicationService) combineItems(items []DeduplicateItem) DeduplicateItem {
	if len(items) == 0 {
		return DeduplicateItem{}
	}

	// Use first item as base
	merged := items[0]

	// Combine content (deduplicate sentences)
	allContent := []string{merged.Content}
	for i := 1; i < len(items); i++ {
		allContent = append(allContent, items[i].Content)
	}
	merged.Content = s.deduplicateSentences(allContent)

	// Merge metadata if enabled
	if s.config.PreserveMetadata {
		merged.Metadata = s.mergeMetadata(items)
	}

	return merged
}

// deduplicateSentences removes duplicate sentences from combined content.
func (s *SemanticDeduplicationService) deduplicateSentences(contents []string) string {
	seen := make(map[string]bool)
	unique := []string{}

	for _, content := range contents {
		sentences := splitIntoSentences(content)
		for _, sentence := range sentences {
			normalized := normalizeText(sentence)
			if !seen[normalized] && len(normalized) > 5 {
				seen[normalized] = true
				unique = append(unique, sentence)
			}
		}
	}

	return strings.Join(unique, " ")
}

// mergeMetadata combines metadata from multiple items.
func (s *SemanticDeduplicationService) mergeMetadata(items []DeduplicateItem) map[string]interface{} {
	merged := make(map[string]interface{})

	for _, item := range items {
		for key, value := range item.Metadata {
			// For duplicate keys, store as array
			if existing, exists := merged[key]; exists {
				switch v := existing.(type) {
				case []interface{}:
					merged[key] = append(v, value)
				default:
					merged[key] = []interface{}{existing, value}
				}
			} else {
				merged[key] = value
			}
		}
	}

	return merged
}

// calculateSimilarity calculates fuzzy similarity between two texts (0.0-1.0).
func (s *SemanticDeduplicationService) calculateSimilarity(text1, text2 string) float64 {
	// Normalize texts
	norm1 := normalizeText(text1)
	norm2 := normalizeText(text2)

	// Quick exact match
	if norm1 == norm2 {
		return 1.0
	}

	// Levenshtein distance-based similarity
	distance := levenshteinDistance(norm1, norm2)
	maxLen := max(len(norm1), len(norm2))

	if maxLen == 0 {
		return 1.0
	}

	similarity := 1.0 - (float64(distance) / float64(maxLen))
	return maxSimilarity(0.0, similarity)
}

// normalizeText normalizes text for comparison.
func normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove extra whitespace
	text = strings.Join(strings.Fields(text), " ")

	// Remove punctuation (keep letters, numbers, spaces)
	result := strings.Builder{}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}

	return strings.TrimSpace(result.String())
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1, // deletion
				min(
					matrix[i][j-1]+1,      // insertion
					matrix[i-1][j-1]+cost, // substitution
				),
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// splitIntoSentences splits text into sentences.
func splitIntoSentences(text string) []string {
	// Simple sentence splitter
	text = strings.ReplaceAll(text, "? ", "?\n")
	text = strings.ReplaceAll(text, "! ", "!\n")
	text = strings.ReplaceAll(text, ". ", ".\n")

	sentences := strings.Split(text, "\n")
	result := []string{}

	for _, s := range sentences {
		trimmed := strings.TrimSpace(s)
		if len(trimmed) > 5 {
			result = append(result, trimmed)
		}
	}

	return result
}

// calculateTotalBytes calculates total bytes in items.
func (s *SemanticDeduplicationService) calculateTotalBytes(items []DeduplicateItem) int {
	total := 0
	for _, item := range items {
		total += len(item.Content)
	}
	return total
}

// updateStats updates deduplication statistics.
func (s *SemanticDeduplicationService) updateStats(result DeduplicationResult) {
	atomic.AddInt64(&s.stats.TotalProcessed, int64(result.OriginalCount))
	atomic.AddInt64(&s.stats.DuplicatesRemoved, int64(result.DuplicatesRemoved))
	atomic.AddInt64(&s.stats.BytesSaved, int64(result.BytesSaved))

	// Update average similarity
	if len(result.Groups) > 0 {
		avgSim := 0.0
		for _, group := range result.Groups {
			avgSim += group.Similarity
		}
		avgSim /= float64(len(result.Groups))

		s.mu.Lock()
		alpha := 0.1
		s.stats.AvgSimilarity = alpha*avgSim + (1-alpha)*s.stats.AvgSimilarity
		s.mu.Unlock()
	}

	if result.DuplicatesRemoved > 0 {
		atomic.AddInt64(&s.stats.DuplicatesFound, int64(len(result.Groups)))
	}
}

// GetStats returns current deduplication statistics.
func (s *SemanticDeduplicationService) GetStats() DeduplicationStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return DeduplicationStats{
		TotalProcessed:    atomic.LoadInt64(&s.stats.TotalProcessed),
		DuplicatesFound:   atomic.LoadInt64(&s.stats.DuplicatesFound),
		DuplicatesRemoved: atomic.LoadInt64(&s.stats.DuplicatesRemoved),
		BytesSaved:        atomic.LoadInt64(&s.stats.BytesSaved),
		AvgSimilarity:     s.stats.AvgSimilarity,
	}
}

// FindDuplicates finds duplicate items without merging them.
func (s *SemanticDeduplicationService) FindDuplicates(items []DeduplicateItem) []DuplicateGroup {
	return s.findDuplicateGroups(items)
}

// CalculateSimilarity exposes similarity calculation for external use.
func (s *SemanticDeduplicationService) CalculateSimilarity(text1, text2 string) float64 {
	return s.calculateSimilarity(text1, text2)
}

// Helper functions.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxSimilarity(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
