package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// SentimentLabel represents the sentiment classification.
type SentimentLabel string

const (
	SentimentPositive SentimentLabel = "POSITIVE"
	SentimentNegative SentimentLabel = "NEGATIVE"
	SentimentNeutral  SentimentLabel = "NEUTRAL"
	SentimentMixed    SentimentLabel = "MIXED"
)

// SentimentResult represents the result of sentiment analysis.
type SentimentResult struct {
	Label             SentimentLabel  `json:"label"`
	Confidence        float64         `json:"confidence"`         // 0.0-1.0 confidence in the label
	Scores            SentimentScores `json:"scores"`             // Individual scores for each sentiment
	Intensity         float64         `json:"intensity"`          // Emotional intensity (0.0-1.0)
	ProcessingTime    float64         `json:"processing_time_ms"` // Processing time in milliseconds
	ModelUsed         string          `json:"model_used"`         // Model identifier
	EmotionalTone     EmotionalTone   `json:"emotional_tone"`     // Fine-grained emotional analysis
	SubjectivityScore float64         `json:"subjectivity_score"` // How subjective vs objective (0.0-1.0)
}

// SentimentScores contains individual sentiment scores.
type SentimentScores struct {
	Positive float64 `json:"positive"` // 0.0-1.0
	Negative float64 `json:"negative"` // 0.0-1.0
	Neutral  float64 `json:"neutral"`  // 0.0-1.0
}

// EmotionalTone represents fine-grained emotional analysis.
type EmotionalTone struct {
	Joy      float64 `json:"joy"`      // 0.0-1.0
	Sadness  float64 `json:"sadness"`  // 0.0-1.0
	Anger    float64 `json:"anger"`    // 0.0-1.0
	Fear     float64 `json:"fear"`     // 0.0-1.0
	Surprise float64 `json:"surprise"` // 0.0-1.0
	Disgust  float64 `json:"disgust"`  // 0.0-1.0
}

// SentimentTrend represents sentiment evolution over time.
type SentimentTrend struct {
	Timestamp  time.Time      `json:"timestamp"`
	Sentiment  SentimentLabel `json:"sentiment"`
	Confidence float64        `json:"confidence"`
	MovingAvg  float64        `json:"moving_average"` // Moving average of sentiment scores
}

// SentimentSummary contains aggregated sentiment statistics.
type SentimentSummary struct {
	TotalMemories     int              `json:"total_memories"`
	PositiveCount     int              `json:"positive_count"`
	NegativeCount     int              `json:"negative_count"`
	NeutralCount      int              `json:"neutral_count"`
	AverageIntensity  float64          `json:"average_intensity"`
	DominantSentiment SentimentLabel   `json:"dominant_sentiment"`
	SentimentScore    float64          `json:"sentiment_score"` // -1.0 (very negative) to +1.0 (very positive)
	Trend             []SentimentTrend `json:"trend"`
	EmotionalProfile  EmotionalTone    `json:"emotional_profile"` // Aggregated emotional profile
}

// SentimentAnalyzer provides sentiment analysis using transformer models.
type SentimentAnalyzer struct {
	config          EnhancedNLPConfig
	repository      ElementRepository
	modelProvider   ONNXModelProvider
	fallbackEnabled bool
}

// NewSentimentAnalyzer creates a new sentiment analyzer.
func NewSentimentAnalyzer(
	config EnhancedNLPConfig,
	repository ElementRepository,
	modelProvider ONNXModelProvider,
) *SentimentAnalyzer {
	return &SentimentAnalyzer{
		config:          config,
		repository:      repository,
		modelProvider:   modelProvider,
		fallbackEnabled: config.EnableFallback,
	}
}

// AnalyzeMemory analyzes sentiment of a single memory.
func (s *SentimentAnalyzer) AnalyzeMemory(ctx context.Context, memoryID string) (*SentimentResult, error) {
	// Get memory
	element, err := s.repository.GetByID(memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	memory, ok := element.(*domain.Memory)
	if !ok {
		return nil, fmt.Errorf("element %s is not a memory", memoryID)
	}

	return s.AnalyzeText(ctx, memory.Content)
}

// AnalyzeText analyzes sentiment of raw text.
func (s *SentimentAnalyzer) AnalyzeText(ctx context.Context, text string) (*SentimentResult, error) {
	startTime := time.Now()

	// Check if ONNX is available
	if !s.modelProvider.IsAvailable() {
		if s.fallbackEnabled {
			return s.fallbackAnalysis(text)
		}
		return nil, fmt.Errorf("ONNX runtime not available and fallback disabled")
	}

	// Analyze using transformer model
	result, err := s.modelProvider.AnalyzeSentiment(ctx, text)
	if err != nil {
		if s.fallbackEnabled {
			return s.fallbackAnalysis(text)
		}
		return nil, fmt.Errorf("sentiment analysis failed: %w", err)
	}

	// Calculate processing time
	result.ProcessingTime = float64(time.Since(startTime).Milliseconds())

	// Filter by threshold
	if result.Confidence < s.config.SentimentThreshold {
		result.Label = SentimentNeutral
	}

	return result, nil
}

// AnalyzeMemoryBatch analyzes sentiment of multiple memories in batch.
func (s *SentimentAnalyzer) AnalyzeMemoryBatch(ctx context.Context, memoryIDs []string) ([]*SentimentResult, error) {
	results := make([]*SentimentResult, 0, len(memoryIDs))

	for _, memoryID := range memoryIDs {
		result, err := s.AnalyzeMemory(ctx, memoryID)
		if err != nil {
			continue // Skip failed analyses
		}
		results = append(results, result)
	}

	return results, nil
}

// AnalyzeMemoryTrend analyzes sentiment trend over a series of memories.
func (s *SentimentAnalyzer) AnalyzeMemoryTrend(ctx context.Context, memoryIDs []string) ([]SentimentTrend, error) {
	trends := make([]SentimentTrend, 0, len(memoryIDs))
	movingWindow := make([]float64, 0, 5) // 5-point moving average

	for _, memoryID := range memoryIDs {
		element, err := s.repository.GetByID(memoryID)
		if err != nil {
			continue
		}

		memory, ok := element.(*domain.Memory)
		if !ok {
			continue
		}

		// Analyze sentiment
		result, err := s.AnalyzeText(ctx, memory.Content)
		if err != nil {
			continue
		}

		// Convert sentiment to score (-1 to +1)
		sentimentScore := s.sentimentToScore(result)

		// Update moving window
		movingWindow = append(movingWindow, sentimentScore)
		if len(movingWindow) > 5 {
			movingWindow = movingWindow[1:]
		}

		// Calculate moving average
		movingAvg := 0.0
		for _, score := range movingWindow {
			movingAvg += score
		}
		movingAvg /= float64(len(movingWindow))

		trends = append(trends, SentimentTrend{
			Timestamp:  memory.GetMetadata().CreatedAt,
			Sentiment:  result.Label,
			Confidence: result.Confidence,
			MovingAvg:  movingAvg,
		})
	}

	return trends, nil
}

// SummarizeMemorySentiments creates a sentiment summary for a collection of memories.
func (s *SentimentAnalyzer) SummarizeMemorySentiments(ctx context.Context, memoryIDs []string) (*SentimentSummary, error) {
	summary := &SentimentSummary{
		TotalMemories: len(memoryIDs),
		Trend:         make([]SentimentTrend, 0),
	}

	var totalIntensity float64
	var totalSentimentScore float64
	aggregatedEmotions := EmotionalTone{}

	// Analyze each memory
	for _, memoryID := range memoryIDs {
		result, err := s.AnalyzeMemory(ctx, memoryID)
		if err != nil {
			continue
		}

		// Count by sentiment
		switch result.Label {
		case SentimentPositive:
			summary.PositiveCount++
		case SentimentNegative:
			summary.NegativeCount++
		case SentimentNeutral:
			summary.NeutralCount++
		}

		// Accumulate intensity and sentiment score
		totalIntensity += result.Intensity
		totalSentimentScore += s.sentimentToScore(result)

		// Aggregate emotional tone
		aggregatedEmotions.Joy += result.EmotionalTone.Joy
		aggregatedEmotions.Sadness += result.EmotionalTone.Sadness
		aggregatedEmotions.Anger += result.EmotionalTone.Anger
		aggregatedEmotions.Fear += result.EmotionalTone.Fear
		aggregatedEmotions.Surprise += result.EmotionalTone.Surprise
		aggregatedEmotions.Disgust += result.EmotionalTone.Disgust
	}

	// Calculate averages
	analyzedCount := summary.PositiveCount + summary.NegativeCount + summary.NeutralCount
	if analyzedCount > 0 {
		summary.AverageIntensity = totalIntensity / float64(analyzedCount)
		summary.SentimentScore = totalSentimentScore / float64(analyzedCount)

		// Average emotional profile
		summary.EmotionalProfile = EmotionalTone{
			Joy:      aggregatedEmotions.Joy / float64(analyzedCount),
			Sadness:  aggregatedEmotions.Sadness / float64(analyzedCount),
			Anger:    aggregatedEmotions.Anger / float64(analyzedCount),
			Fear:     aggregatedEmotions.Fear / float64(analyzedCount),
			Surprise: aggregatedEmotions.Surprise / float64(analyzedCount),
			Disgust:  aggregatedEmotions.Disgust / float64(analyzedCount),
		}
	}

	// Determine dominant sentiment
	if summary.PositiveCount > summary.NegativeCount && summary.PositiveCount > summary.NeutralCount {
		summary.DominantSentiment = SentimentPositive
	} else if summary.NegativeCount > summary.PositiveCount && summary.NegativeCount > summary.NeutralCount {
		summary.DominantSentiment = SentimentNegative
	} else {
		summary.DominantSentiment = SentimentNeutral
	}

	// Get trend
	trend, err := s.AnalyzeMemoryTrend(ctx, memoryIDs)
	if err == nil {
		summary.Trend = trend
	}

	return summary, nil
}

// sentimentToScore converts sentiment label and scores to a single score (-1 to +1).
func (s *SentimentAnalyzer) sentimentToScore(result *SentimentResult) float64 {
	// Weight by confidence
	score := (result.Scores.Positive - result.Scores.Negative) * result.Confidence

	// Clamp to [-1, 1]
	if score > 1.0 {
		score = 1.0
	} else if score < -1.0 {
		score = -1.0
	}

	return score
}

// fallbackAnalysis provides rule-based sentiment analysis as fallback.
func (s *SentimentAnalyzer) fallbackAnalysis(text string) (*SentimentResult, error) {
	// Simple lexicon-based approach
	positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "fantastic", "love", "happy", "joy"}
	negativeWords := []string{"bad", "terrible", "awful", "horrible", "hate", "sad", "angry", "fear", "disgust"}

	text = strings.ToLower(text)
	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(text, word)
	}

	for _, word := range negativeWords {
		negativeCount += strings.Count(text, word)
	}

	// Determine sentiment
	var label SentimentLabel
	var scores SentimentScores

	total := float64(positiveCount + negativeCount)
	if total == 0 {
		label = SentimentNeutral
		scores = SentimentScores{
			Positive: 0.33,
			Negative: 0.33,
			Neutral:  0.34,
		}
	} else {
		posScore := float64(positiveCount) / total
		negScore := float64(negativeCount) / total

		scores = SentimentScores{
			Positive: posScore,
			Negative: negScore,
			Neutral:  0.0,
		}

		if posScore > negScore {
			label = SentimentPositive
		} else {
			label = SentimentNegative
		}
	}

	// Calculate confidence (lower for fallback)
	confidence := 0.4
	if total > 5 {
		confidence = 0.5
	}

	return &SentimentResult{
		Label:      label,
		Confidence: confidence,
		Scores:     scores,
		Intensity:  0.5,
		ModelUsed:  "lexicon-fallback",
		EmotionalTone: EmotionalTone{
			Joy:      scores.Positive,
			Sadness:  scores.Negative * 0.5,
			Anger:    scores.Negative * 0.3,
			Fear:     scores.Negative * 0.2,
			Surprise: 0.1,
			Disgust:  0.0,
		},
		SubjectivityScore: 0.5,
	}, nil
}

// DetectEmotionalShifts detects significant changes in sentiment over time.
func (s *SentimentAnalyzer) DetectEmotionalShifts(ctx context.Context, memoryIDs []string, threshold float64) ([]EmotionalShift, error) {
	trends, err := s.AnalyzeMemoryTrend(ctx, memoryIDs)
	if err != nil {
		return nil, err
	}

	shifts := make([]EmotionalShift, 0)

	for i := 1; i < len(trends); i++ {
		prev := trends[i-1]
		curr := trends[i]

		// Calculate change in moving average
		change := curr.MovingAvg - prev.MovingAvg

		if abs(change) >= threshold {
			direction := "positive"
			if change < 0 {
				direction = "negative"
			}

			shifts = append(shifts, EmotionalShift{
				Timestamp:     curr.Timestamp,
				FromSentiment: prev.Sentiment,
				ToSentiment:   curr.Sentiment,
				Magnitude:     abs(change),
				Direction:     direction,
			})
		}
	}

	return shifts, nil
}

// EmotionalShift represents a significant change in sentiment.
type EmotionalShift struct {
	Timestamp     time.Time      `json:"timestamp"`
	FromSentiment SentimentLabel `json:"from_sentiment"`
	ToSentiment   SentimentLabel `json:"to_sentiment"`
	Magnitude     float64        `json:"magnitude"` // How large the shift was
	Direction     string         `json:"direction"` // "positive" or "negative"
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
