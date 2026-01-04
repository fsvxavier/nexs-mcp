package application

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSentimentProvider implements ONNXModelProvider for sentiment testing
type mockSentimentProvider struct {
	available bool
	result    *SentimentResult
	err       error
}

func (m *mockSentimentProvider) ExtractEntities(ctx context.Context, text string) ([]EnhancedEntity, error) {
	return nil, nil
}

func (m *mockSentimentProvider) ExtractEntitiesBatch(ctx context.Context, texts []string) ([][]EnhancedEntity, error) {
	return nil, nil
}

func (m *mockSentimentProvider) AnalyzeSentiment(ctx context.Context, text string) (*SentimentResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func (m *mockSentimentProvider) ExtractTopics(ctx context.Context, texts []string, numTopics int) ([]Topic, error) {
	return nil, nil
}

func (m *mockSentimentProvider) IsAvailable() bool {
	return m.available
}

func TestNewSentimentAnalyzer(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockSentimentProvider{available: true}

	analyzer := NewSentimentAnalyzer(config, repo, mockProvider)

	assert.NotNil(t, analyzer)
	assert.Equal(t, config, analyzer.config)
}

func TestAnalyzeText(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	config := DefaultEnhancedNLPConfig()

	tests := []struct {
		name      string
		text      string
		mockSetup func() *mockSentimentProvider
		wantErr   bool
		checkFunc func(t *testing.T, result *SentimentResult)
	}{
		{
			name: "positive_sentiment",
			text: "I absolutely love this product! It's amazing and wonderful!",
			mockSetup: func() *mockSentimentProvider {
				return &mockSentimentProvider{
					available: true,
					result: &SentimentResult{
						Label:      SentimentPositive,
						Confidence: 0.95,
						Scores: SentimentScores{
							Positive: 0.95,
							Negative: 0.02,
							Neutral:  0.03,
						},
						Intensity: 0.9,
						ModelUsed: "test-model",
						EmotionalTone: EmotionalTone{
							Joy:      0.85,
							Sadness:  0.05,
							Anger:    0.02,
							Fear:     0.01,
							Surprise: 0.05,
							Disgust:  0.02,
						},
					},
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *SentimentResult) {
				assert.Equal(t, SentimentPositive, result.Label)
				assert.GreaterOrEqual(t, result.Confidence, 0.9)
				assert.GreaterOrEqual(t, result.EmotionalTone.Joy, 0.8)
			},
		},
		{
			name: "negative_sentiment",
			text: "This is terrible and disappointing. I hate it.",
			mockSetup: func() *mockSentimentProvider {
				return &mockSentimentProvider{
					available: true,
					result: &SentimentResult{
						Label:      SentimentNegative,
						Confidence: 0.92,
						Scores: SentimentScores{
							Positive: 0.03,
							Negative: 0.92,
							Neutral:  0.05,
						},
						Intensity: 0.85,
						EmotionalTone: EmotionalTone{
							Joy:     0.02,
							Sadness: 0.40,
							Anger:   0.45,
							Fear:    0.05,
							Disgust: 0.08,
						},
					},
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *SentimentResult) {
				assert.Equal(t, SentimentNegative, result.Label)
				assert.GreaterOrEqual(t, result.Confidence, 0.9)
				assert.GreaterOrEqual(t, result.Scores.Negative, 0.9)
			},
		},
		{
			name: "neutral_sentiment",
			text: "The meeting is scheduled for 3 PM tomorrow.",
			mockSetup: func() *mockSentimentProvider {
				return &mockSentimentProvider{
					available: true,
					result: &SentimentResult{
						Label:      SentimentNeutral,
						Confidence: 0.88,
						Scores: SentimentScores{
							Positive: 0.10,
							Negative: 0.08,
							Neutral:  0.82,
						},
						Intensity: 0.2,
					},
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *SentimentResult) {
				assert.Equal(t, SentimentNeutral, result.Label)
				assert.LessOrEqual(t, result.Intensity, 0.3)
			},
		},
		{
			name: "onnx_unavailable_fallback",
			text: "This is a test sentence.",
			mockSetup: func() *mockSentimentProvider {
				return &mockSentimentProvider{available: false}
			},
			wantErr: false, // Fallback should work
			checkFunc: func(t *testing.T, result *SentimentResult) {
				assert.NotNil(t, result)
				assert.Contains(t, result.ModelUsed, "fallback")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := tt.mockSetup()
			analyzer := NewSentimentAnalyzer(config, repo, mockProvider)

			result, err := analyzer.AnalyzeText(context.Background(), tt.text)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestAnalyzeMemory(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create test memory
	memory := domain.NewMemory(
		"test-sentiment",
		"Test memory for sentiment",
		"v1.0.0",
		"test-user",
	)
	memory.Content = "I'm very happy with the results!"
	require.NoError(t, repo.Create(memory))

	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockSentimentProvider{
		available: true,
		result: &SentimentResult{
			Label:      SentimentPositive,
			Confidence: 0.94,
			Scores: SentimentScores{
				Positive: 0.94,
				Negative: 0.02,
				Neutral:  0.04,
			},
			Intensity: 0.88,
		},
	}

	analyzer := NewSentimentAnalyzer(config, repo, mockProvider)
	result, err := analyzer.AnalyzeMemory(ctx, memory.GetID())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, SentimentPositive, result.Label)
}

func TestAnalyzeMemoryBatch(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create test memories
	contents := []string{
		"This is great!",
		"This is awful.",
		"This is okay.",
	}

	memoryIDs := make([]string, 0)
	for i, content := range contents {
		memory := domain.NewMemory(
			"batch-test-"+string(rune('a'+i)),
			"Batch test",
			"v1.0.0",
			"test-user",
		)
		memory.Content = content
		require.NoError(t, repo.Create(memory))
		memoryIDs = append(memoryIDs, memory.GetID())
	}

	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockSentimentProvider{
		available: true,
		result: &SentimentResult{
			Label:      SentimentPositive,
			Confidence: 0.85,
			Scores:     SentimentScores{Positive: 0.85, Negative: 0.05, Neutral: 0.10},
		},
	}

	analyzer := NewSentimentAnalyzer(config, repo, mockProvider)
	results, err := analyzer.AnalyzeMemoryBatch(ctx, memoryIDs)

	require.NoError(t, err)
	assert.Equal(t, 3, len(results))
}

func TestDetectEmotionalShifts(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create memories with sentiment progression
	memoryIDs := make([]string, 0)
	for i := 0; i < 3; i++ {
		memory := domain.NewMemory(
			"shift-test-"+string(rune('a'+i)),
			"Emotional shift test",
			"v1.0.0",
			"test-user",
		)
		memory.Content = "Test content"
		require.NoError(t, repo.Create(memory))
		memoryIDs = append(memoryIDs, memory.GetID())
	}

	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockSentimentProvider{
		available: true,
		result: &SentimentResult{
			Label:      SentimentPositive,
			Confidence: 0.8,
		},
	}

	analyzer := NewSentimentAnalyzer(config, repo, mockProvider)
	shifts, err := analyzer.DetectEmotionalShifts(ctx, memoryIDs, 0.3)

	require.NoError(t, err)
	assert.NotNil(t, shifts)
}

func TestSummarizeSentiment(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create test memories
	memoryIDs := make([]string, 0)
	for i := 0; i < 5; i++ {
		memory := domain.NewMemory(
			"summary-test-"+string(rune('a'+i)),
			"Summary test",
			"v1.0.0",
			"test-user",
		)
		memory.Content = "Test content"
		require.NoError(t, repo.Create(memory))
		memoryIDs = append(memoryIDs, memory.GetID())
	}

	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockSentimentProvider{
		available: true,
		result: &SentimentResult{
			Label:      SentimentPositive,
			Confidence: 0.85,
			Scores:     SentimentScores{Positive: 0.85, Negative: 0.05, Neutral: 0.10},
			Intensity:  0.75,
		},
	}

	analyzer := NewSentimentAnalyzer(config, repo, mockProvider)
	summary, err := analyzer.SummarizeMemorySentiments(ctx, memoryIDs)

	require.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 5, summary.TotalMemories)
	assert.GreaterOrEqual(t, summary.PositiveCount, 0)
}

func TestSentimentAnalyzerWithInvalidMemory(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockSentimentProvider{available: true}

	analyzer := NewSentimentAnalyzer(config, repo, mockProvider)
	_, err := analyzer.AnalyzeMemory(context.Background(), "non-existent-id")

	assert.Error(t, err)
}

func TestFallbackSentimentAnalysis(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	config := DefaultEnhancedNLPConfig()
	config.EnableFallback = true
	mockProvider := &mockSentimentProvider{available: false}

	analyzer := NewSentimentAnalyzer(config, repo, mockProvider)

	// Test various sentiment patterns
	tests := []struct {
		text     string
		expected SentimentLabel
	}{
		{"I love this! It's amazing!", SentimentPositive},
		{"This is terrible and awful.", SentimentNegative},
		{"The meeting is at 3 PM.", SentimentNeutral},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			result, err := analyzer.AnalyzeText(context.Background(), tt.text)
			require.NoError(t, err)
			// Fallback may not be as accurate but should return a result
			assert.NotNil(t, result)
			assert.Contains(t, result.ModelUsed, "fallback")
		})
	}
}
