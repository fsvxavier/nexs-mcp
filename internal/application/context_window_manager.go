package application

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ContextWindowManager manages context window to prevent overflow and optimize relevance.
type ContextWindowManager struct {
	config ContextWindowConfig
	stats  ContextWindowStats
	mu     sync.RWMutex
}

// ContextWindowConfig configures context window management.
type ContextWindowConfig struct {
	MaxTokens          int              // Maximum tokens allowed
	PriorityStrategy   PriorityStrategy // Strategy for prioritizing content
	TruncationMethod   TruncationMethod // Method for truncating overflow
	PreserveRecent     int              // Number of recent items to always keep
	RelevanceThreshold float64          // Minimum relevance score (0.0-1.0)
}

// PriorityStrategy defines how to prioritize context items.
type PriorityStrategy string

const (
	PriorityRecency    PriorityStrategy = "recency"    // Most recent first
	PriorityRelevance  PriorityStrategy = "relevance"  // Highest relevance score first
	PriorityHybrid     PriorityStrategy = "hybrid"     // Balanced recency + relevance
	PriorityImportance PriorityStrategy = "importance" // User-marked importance
)

// TruncationMethod defines how to truncate overflowing content.
type TruncationMethod string

const (
	TruncationHead   TruncationMethod = "head"   // Remove oldest items
	TruncationTail   TruncationMethod = "tail"   // Remove newest items (unusual)
	TruncationMiddle TruncationMethod = "middle" // Keep first and last, remove middle
)

// ContextItem represents an item in the context window.
type ContextItem struct {
	ID          string
	Content     string
	TokenCount  int
	CreatedAt   time.Time
	LastAccess  time.Time
	AccessCount int
	Relevance   float64 // 0.0-1.0
	Importance  int     // User-defined priority (0-10)
	Metadata    map[string]interface{}
}

// ContextWindowStats tracks optimization statistics.
type ContextWindowStats struct {
	TotalOptimizations int64
	OverflowsPrevented int64
	TokensSaved        int64
	AvgRelevanceGain   float64
}

// OptimizationResult describes the outcome of context optimization.
type OptimizationResult struct {
	OriginalTokenCount  int
	OptimizedTokenCount int
	ItemsRemoved        int
	ItemsRetained       int
	RelevanceScore      float64
	Method              string
}

// NewContextWindowManager creates a new context window manager.
func NewContextWindowManager(config ContextWindowConfig) *ContextWindowManager {
	// Set defaults
	if config.MaxTokens == 0 {
		config.MaxTokens = 8000 // Claude 3 Sonnet: ~8k context
	}
	if config.PreserveRecent == 0 {
		config.PreserveRecent = 5
	}
	if config.RelevanceThreshold == 0 {
		config.RelevanceThreshold = 0.3
	}
	if config.PriorityStrategy == "" {
		config.PriorityStrategy = PriorityHybrid
	}
	if config.TruncationMethod == "" {
		config.TruncationMethod = TruncationHead
	}

	return &ContextWindowManager{
		config: config,
		stats: ContextWindowStats{
			AvgRelevanceGain: 0.0,
		},
	}
}

// OptimizeContext optimizes the context window to fit within token limits.
func (m *ContextWindowManager) OptimizeContext(ctx context.Context, items []ContextItem) ([]ContextItem, OptimizationResult, error) {
	if len(items) == 0 {
		return items, OptimizationResult{
			OriginalTokenCount:  0,
			OptimizedTokenCount: 0,
			ItemsRemoved:        0,
			ItemsRetained:       0,
			RelevanceScore:      0.0,
			Method:              "none",
		}, nil
	}

	// Calculate total tokens
	totalTokens := 0
	for _, item := range items {
		totalTokens += item.TokenCount
	}

	// Check if optimization is needed
	if totalTokens <= m.config.MaxTokens {
		avgRelevance := m.calculateAverageRelevance(items)
		return items, OptimizationResult{
			OriginalTokenCount:  totalTokens,
			OptimizedTokenCount: totalTokens,
			ItemsRemoved:        0,
			ItemsRetained:       len(items),
			RelevanceScore:      avgRelevance,
			Method:              "none",
		}, nil
	}

	// Apply prioritization strategy
	prioritized := m.prioritizeItems(items)

	// Apply truncation method
	optimized := m.truncateItems(prioritized, m.config.MaxTokens)

	// Calculate metrics
	optimizedTokens := 0
	for _, item := range optimized {
		optimizedTokens += item.TokenCount
	}

	avgRelevance := m.calculateAverageRelevance(optimized)

	result := OptimizationResult{
		OriginalTokenCount:  totalTokens,
		OptimizedTokenCount: optimizedTokens,
		ItemsRemoved:        len(items) - len(optimized),
		ItemsRetained:       len(optimized),
		RelevanceScore:      avgRelevance,
		Method:              fmt.Sprintf("%s+%s", m.config.PriorityStrategy, m.config.TruncationMethod),
	}

	// Update stats
	m.updateStats(result)

	return optimized, result, nil
}

// prioritizeItems applies the priority strategy to sort items.
func (m *ContextWindowManager) prioritizeItems(items []ContextItem) []ContextItem {
	// Create a copy to avoid modifying original
	sorted := make([]ContextItem, len(items))
	copy(sorted, items)

	switch m.config.PriorityStrategy {
	case PriorityRecency:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
		})

	case PriorityRelevance:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Relevance > sorted[j].Relevance
		})

	case PriorityImportance:
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Importance != sorted[j].Importance {
				return sorted[i].Importance > sorted[j].Importance
			}
			// Tie-breaker: recency
			return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
		})

	case PriorityHybrid:
		// Hybrid: 40% relevance + 40% recency + 20% importance
		now := time.Now()
		sort.Slice(sorted, func(i, j int) bool {
			scoreI := m.calculateHybridScore(sorted[i], now)
			scoreJ := m.calculateHybridScore(sorted[j], now)
			return scoreI > scoreJ
		})

	default:
		// Default to recency
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
		})
	}

	return sorted
}

// calculateHybridScore calculates a hybrid priority score.
func (m *ContextWindowManager) calculateHybridScore(item ContextItem, now time.Time) float64 {
	// Relevance component (40%)
	relevanceScore := item.Relevance * 0.4

	// Recency component (40%) - decay over 30 days
	age := now.Sub(item.CreatedAt)
	maxAge := 30 * 24 * time.Hour
	recencyScore := 0.0
	if age < maxAge {
		recencyScore = (1.0 - float64(age)/float64(maxAge)) * 0.4
	}

	// Importance component (20%)
	importanceScore := (float64(item.Importance) / 10.0) * 0.2

	return relevanceScore + recencyScore + importanceScore
}

// truncateItems truncates items to fit within max tokens.
func (m *ContextWindowManager) truncateItems(items []ContextItem, maxTokens int) []ContextItem {
	if len(items) == 0 {
		return items
	}

	switch m.config.TruncationMethod {
	case TruncationHead:
		return m.truncateHead(items, maxTokens)

	case TruncationTail:
		return m.truncateTail(items, maxTokens)

	case TruncationMiddle:
		return m.truncateMiddle(items, maxTokens)

	default:
		return m.truncateHead(items, maxTokens)
	}
}

// truncateHead keeps items from the beginning until maxTokens is reached.
func (m *ContextWindowManager) truncateHead(items []ContextItem, maxTokens int) []ContextItem {
	result := []ContextItem{}
	currentTokens := 0

	for _, item := range items {
		if currentTokens+item.TokenCount <= maxTokens {
			result = append(result, item)
			currentTokens += item.TokenCount
		} else {
			break
		}
	}

	// Ensure we preserve minimum recent items
	if len(result) < m.config.PreserveRecent && len(items) > len(result) {
		result = items[:minPreserveRecent(m.config.PreserveRecent, len(items))]
	}

	return result
}

// truncateTail keeps items from the end (unusual but available).
func (m *ContextWindowManager) truncateTail(items []ContextItem, maxTokens int) []ContextItem {
	result := []ContextItem{}
	currentTokens := 0

	// Reverse iteration
	for i := len(items) - 1; i >= 0; i-- {
		if currentTokens+items[i].TokenCount <= maxTokens {
			result = append([]ContextItem{items[i]}, result...) // Prepend
			currentTokens += items[i].TokenCount
		} else {
			break
		}
	}

	return result
}

// truncateMiddle keeps first and last items, removes middle ones.
func (m *ContextWindowManager) truncateMiddle(items []ContextItem, maxTokens int) []ContextItem {
	if len(items) <= 2*m.config.PreserveRecent {
		return m.truncateHead(items, maxTokens)
	}

	// Always preserve first and last N items
	preserve := m.config.PreserveRecent
	firstItems := items[:preserve]
	lastItems := items[len(items)-preserve:]

	// Calculate tokens for preserved items
	preservedTokens := 0
	for _, item := range firstItems {
		preservedTokens += item.TokenCount
	}
	for _, item := range lastItems {
		preservedTokens += item.TokenCount
	}

	// If preserved items already exceed max, truncate head
	if preservedTokens >= maxTokens {
		return m.truncateHead(items, maxTokens)
	}

	// Fill remaining space with middle items (prioritized)
	middleItems := items[preserve : len(items)-preserve]
	remainingTokens := maxTokens - preservedTokens

	selectedMiddle := []ContextItem{}
	currentTokens := 0

	for _, item := range middleItems {
		if currentTokens+item.TokenCount <= remainingTokens {
			selectedMiddle = append(selectedMiddle, item)
			currentTokens += item.TokenCount
		}
	}

	// Combine: first + selected middle + last
	result := append([]ContextItem{}, firstItems...)
	result = append(result, selectedMiddle...)
	result = append(result, lastItems...)

	return result
}

// calculateAverageRelevance calculates average relevance of items.
func (m *ContextWindowManager) calculateAverageRelevance(items []ContextItem) float64 {
	if len(items) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, item := range items {
		sum += item.Relevance
	}

	return sum / float64(len(items))
}

// FilterByRelevance filters items below relevance threshold.
func (m *ContextWindowManager) FilterByRelevance(items []ContextItem) []ContextItem {
	filtered := []ContextItem{}

	for _, item := range items {
		if item.Relevance >= m.config.RelevanceThreshold {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// updateStats updates context optimization statistics.
func (m *ContextWindowManager) updateStats(result OptimizationResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.TotalOptimizations++
	if result.ItemsRemoved > 0 {
		m.stats.OverflowsPrevented++
		m.stats.TokensSaved += int64(result.OriginalTokenCount - result.OptimizedTokenCount)
	}

	// Update average relevance gain using exponential moving average
	alpha := 0.1
	m.stats.AvgRelevanceGain = alpha*result.RelevanceScore + (1-alpha)*m.stats.AvgRelevanceGain
}

// GetStats returns current context optimization statistics.
func (m *ContextWindowManager) GetStats() ContextWindowStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.stats
}

// EstimateTokens estimates token count from text (rough approximation).
func EstimateTokens(text string) int {
	// Rough estimate: 1 token ≈ 4 characters for English
	// This is a simplified approximation
	return len(text) / 4
}

// min helper function.
func minPreserveRecent(a, b int) int {
	if a < b {
		return a
	}
	return b
}
