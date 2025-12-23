package quality

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FallbackScorer coordinates multiple scorers with automatic fallback
type FallbackScorer struct {
	config  *Config
	scorers map[string]Scorer
	mu      sync.RWMutex
	stats   *FallbackStats
}

// FallbackStats tracks usage statistics for each scorer
type FallbackStats struct {
	mu        sync.RWMutex
	calls     map[string]int
	successes map[string]int
	failures  map[string]int
	totalCost float64
}

// NewFallbackScorer creates a fallback scorer with multiple backends
func NewFallbackScorer(config *Config) (*FallbackScorer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	fs := &FallbackScorer{
		config:  config,
		scorers: make(map[string]Scorer),
		stats: &FallbackStats{
			calls:     make(map[string]int),
			successes: make(map[string]int),
			failures:  make(map[string]int),
		},
	}

	// Initialize scorers based on fallback chain
	if err := fs.initializeScorers(); err != nil {
		return nil, fmt.Errorf("failed to initialize scorers: %w", err)
	}

	return fs, nil
}

// initializeScorers creates all configured scorer instances
func (fs *FallbackScorer) initializeScorers() error {
	for _, scorerName := range fs.config.FallbackChain {
		var scorer Scorer
		var err error

		switch scorerName {
		case "onnx":
			scorer, err = NewONNXScorer(fs.config)
			if err != nil {
				// ONNX may not be available in some builds
				continue
			}
		case "implicit":
			scorer = NewImplicitScorer(fs.config)
		case "groq", "gemini":
			// These will be implemented separately
			continue
		default:
			continue
		}

		fs.scorers[scorerName] = scorer
	}

	if len(fs.scorers) == 0 {
		return fmt.Errorf("no scorers available")
	}

	return nil
}

// Score attempts to score content using fallback chain
func (fs *FallbackScorer) Score(ctx context.Context, content string) (*Score, error) {
	var lastErr error

	for _, scorerName := range fs.config.FallbackChain {
		fs.mu.RLock()
		scorer, exists := fs.scorers[scorerName]
		fs.mu.RUnlock()

		if !exists {
			continue
		}

		// Track call
		fs.recordCall(scorerName)

		// Check if scorer is available
		if !scorer.IsAvailable(ctx) {
			fs.recordFailure(scorerName)
			lastErr = fmt.Errorf("%s scorer not available", scorerName)
			continue
		}

		// Attempt scoring
		score, err := scorer.Score(ctx, content)
		if err != nil {
			fs.recordFailure(scorerName)
			lastErr = err
			continue
		}

		// Success!
		fs.recordSuccess(scorerName, scorer.Cost())

		// Add fallback metadata
		if score.Metadata == nil {
			score.Metadata = make(map[string]interface{})
		}
		score.Metadata["fallback_used"] = scorerName
		score.Metadata["fallback_attempts"] = fs.getAttemptCount(scorerName)

		return score, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all scorers failed, last error: %w", lastErr)
	}

	return nil, fmt.Errorf("no scorers available in fallback chain")
}

// ScoreBatch scores multiple contents using fallback chain
func (fs *FallbackScorer) ScoreBatch(ctx context.Context, contents []string) ([]*Score, error) {
	scores := make([]*Score, len(contents))

	for i, content := range contents {
		score, err := fs.Score(ctx, content)
		if err != nil {
			return nil, fmt.Errorf("failed to score content %d: %w", i, err)
		}
		scores[i] = score
	}

	return scores, nil
}

// GetPreferredScorer returns the best available scorer
func (fs *FallbackScorer) GetPreferredScorer(ctx context.Context) Scorer {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for _, scorerName := range fs.config.FallbackChain {
		if scorer, exists := fs.scorers[scorerName]; exists {
			if scorer.IsAvailable(ctx) {
				return scorer
			}
		}
	}

	return nil
}

// recordCall increments call counter for a scorer
func (fs *FallbackScorer) recordCall(scorerName string) {
	fs.stats.mu.Lock()
	defer fs.stats.mu.Unlock()
	fs.stats.calls[scorerName]++
}

// recordSuccess increments success counter and adds cost
func (fs *FallbackScorer) recordSuccess(scorerName string, cost float64) {
	fs.stats.mu.Lock()
	defer fs.stats.mu.Unlock()
	fs.stats.successes[scorerName]++
	fs.stats.totalCost += cost
}

// recordFailure increments failure counter
func (fs *FallbackScorer) recordFailure(scorerName string) {
	fs.stats.mu.Lock()
	defer fs.stats.mu.Unlock()
	fs.stats.failures[scorerName]++
}

// getAttemptCount returns the number of times a scorer was called
func (fs *FallbackScorer) getAttemptCount(scorerName string) int {
	fs.stats.mu.RLock()
	defer fs.stats.mu.RUnlock()
	return fs.stats.calls[scorerName]
}

// GetStats returns current fallback statistics
func (fs *FallbackScorer) GetStats() map[string]interface{} {
	fs.stats.mu.RLock()
	defer fs.stats.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["calls"] = copyIntMap(fs.stats.calls)
	stats["successes"] = copyIntMap(fs.stats.successes)
	stats["failures"] = copyIntMap(fs.stats.failures)
	stats["total_cost"] = fs.stats.totalCost
	stats["timestamp"] = time.Now()

	return stats
}

// ResetStats clears all statistics
func (fs *FallbackScorer) ResetStats() {
	fs.stats.mu.Lock()
	defer fs.stats.mu.Unlock()

	fs.stats.calls = make(map[string]int)
	fs.stats.successes = make(map[string]int)
	fs.stats.failures = make(map[string]int)
	fs.stats.totalCost = 0
}

// Name returns the scorer identifier
func (fs *FallbackScorer) Name() string {
	return "fallback"
}

// IsAvailable checks if any scorer in the chain is available
func (fs *FallbackScorer) IsAvailable(ctx context.Context) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for _, scorerName := range fs.config.FallbackChain {
		if scorer, exists := fs.scorers[scorerName]; exists {
			if scorer.IsAvailable(ctx) {
				return true
			}
		}
	}

	return false
}

// Cost returns the average cost across all scorers
func (fs *FallbackScorer) Cost() float64 {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var totalCost float64
	count := 0

	for _, scorer := range fs.scorers {
		totalCost += scorer.Cost()
		count++
	}

	if count == 0 {
		return 0
	}

	return totalCost / float64(count)
}

// Close releases all scorer resources
func (fs *FallbackScorer) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	var lastErr error
	for name, scorer := range fs.scorers {
		if err := scorer.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close %s scorer: %w", name, err)
		}
	}

	fs.scorers = make(map[string]Scorer)
	return lastErr
}

// copyIntMap creates a copy of an int map
func copyIntMap(m map[string]int) map[string]int {
	copy := make(map[string]int, len(m))
	for k, v := range m {
		copy[k] = v
	}
	return copy
}
