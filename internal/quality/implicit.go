package quality

import (
	"context"
	"math"
	"time"
)

// ImplicitScorer calculates quality scores based on implicit signals.
type ImplicitScorer struct {
	config *Config
}

// NewImplicitScorer creates a new implicit signal scorer.
func NewImplicitScorer(config *Config) *ImplicitScorer {
	if config == nil {
		config = DefaultConfig()
	}
	return &ImplicitScorer{
		config: config,
	}
}

// Score calculates quality based on implicit signals.
func (s *ImplicitScorer) Score(ctx context.Context, content string) (*Score, error) {
	// For single content without signals, return default medium-low score
	return &Score{
		Value:      0.4,
		Confidence: 0.3,
		Method:     "implicit",
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"note": "default score - no signals provided",
		},
	}, nil
}

// ScoreWithSignals calculates quality based on provided signals.
func (s *ImplicitScorer) ScoreWithSignals(ctx context.Context, content string, signals ImplicitSignals) (*Score, error) {
	score := s.calculateScore(content, signals)
	confidence := s.calculateConfidence(signals)

	return &Score{
		Value:      score,
		Confidence: confidence,
		Method:     "implicit",
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"signals": signals,
		},
	}, nil
}

// calculateScore computes the quality score from signals.
func (s *ImplicitScorer) calculateScore(content string, signals ImplicitSignals) float64 {
	var score float64

	// Access frequency (0-0.3 points)
	accessScore := math.Min(float64(signals.AccessCount)/20.0, 1.0) * 0.3
	score += accessScore

	// Reference count (0-0.25 points)
	refScore := math.Min(float64(signals.ReferenceCount)/10.0, 1.0) * 0.25
	score += refScore

	// Recency score (0-0.2 points)
	// Newer is better, decay over time
	var recencyScore float64
	switch {
	case signals.AgeInDays <= 7:
		recencyScore = 0.2 // Very recent
	case signals.AgeInDays <= 30:
		recencyScore = 0.15 // Recent
	case signals.AgeInDays <= 90:
		recencyScore = 0.1 // Somewhat recent
	default:
		recencyScore = 0.05 // Old
	}
	score += recencyScore

	// Last access recency (0-0.15 points)
	var lastAccessScore float64
	switch {
	case signals.LastAccessDays <= 1:
		lastAccessScore = 0.15
	case signals.LastAccessDays <= 7:
		lastAccessScore = 0.1
	case signals.LastAccessDays <= 30:
		lastAccessScore = 0.05
	}
	score += lastAccessScore

	// User rating (0-0.1 points)
	if signals.UserRating >= 0 && signals.UserRating <= 1.0 {
		score += signals.UserRating * 0.1
	}

	// Content length (0-0.05 points)
	// Prefer medium-length content (not too short, not too long)
	if signals.ContentLength >= 100 && signals.ContentLength <= 2000 {
		score += 0.05
	} else if signals.ContentLength >= 50 && signals.ContentLength <= 5000 {
		score += 0.03
	}

	// Tag count (0-0.05 points)
	if signals.TagCount > 0 {
		tagScore := math.Min(float64(signals.TagCount)/5.0, 1.0) * 0.05
		score += tagScore
	}

	// Promotion bonus (0.05 points)
	if signals.IsPromoted {
		score += 0.05
	}

	// Normalize to 0-1 range
	score = math.Max(0.0, math.Min(1.0, score))

	return score
}

// calculateConfidence estimates how confident we are in the implicit score.
func (s *ImplicitScorer) calculateConfidence(signals ImplicitSignals) float64 {
	var confidence = 0.5 // Base confidence for implicit scoring

	// More signals = higher confidence
	signalCount := 0
	if signals.AccessCount > 0 {
		signalCount++
	}
	if signals.ReferenceCount > 0 {
		signalCount++
	}
	if signals.UserRating >= 0 {
		signalCount++
		confidence += 0.2 // User ratings are high-confidence signals
	}
	if signals.TagCount > 0 {
		signalCount++
	}
	if signals.IsPromoted {
		signalCount++
		confidence += 0.1 // Promotion is a strong signal
	}

	// Add confidence based on signal diversity
	confidence += float64(signalCount) * 0.05

	// Cap confidence at reasonable level (implicit never as confident as ML models)
	confidence = math.Min(0.8, confidence)

	return confidence
}

// ScoreBatch scores multiple contents with their signals.
func (s *ImplicitScorer) ScoreBatch(ctx context.Context, contents []string) ([]*Score, error) {
	scores := make([]*Score, len(contents))
	for i, content := range contents {
		score, err := s.Score(ctx, content)
		if err != nil {
			return nil, err
		}
		scores[i] = score
	}
	return scores, nil
}

// Name returns the scorer identifier.
func (s *ImplicitScorer) Name() string {
	return ScorerImplicit
}

// IsAvailable always returns true (implicit scoring always available).
func (s *ImplicitScorer) IsAvailable(ctx context.Context) bool {
	return true
}

// Cost returns 0.0 (free, no external API calls).
func (s *ImplicitScorer) Cost() float64 {
	return 0.0
}

// Close is a no-op for implicit scorer.
func (s *ImplicitScorer) Close() error {
	return nil
}
