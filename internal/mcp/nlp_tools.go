package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ExtractEntitiesAdvancedInput represents input for extract_entities_advanced tool.
type ExtractEntitiesAdvancedInput struct {
	Text      string   `json:"text,omitempty"`
	MemoryID  string   `json:"memory_id,omitempty"`
	MemoryIDs []string `json:"memory_ids,omitempty"`
}

// AnalyzeSentimentInput represents input for analyze_sentiment tool.
type AnalyzeSentimentInput struct {
	Text      string   `json:"text,omitempty"`
	MemoryID  string   `json:"memory_id,omitempty"`
	MemoryIDs []string `json:"memory_ids,omitempty"`
}

// ExtractTopicsInput represents input for extract_topics tool.
type ExtractTopicsInput struct {
	MemoryIDs []string `json:"memory_ids"`
	NumTopics int      `json:"num_topics,omitempty"`
	Algorithm string   `json:"algorithm,omitempty"`
}

// AnalyzeSentimentTrendInput represents input for analyze_sentiment_trend tool.
type AnalyzeSentimentTrendInput struct {
	MemoryIDs []string `json:"memory_ids"`
}

// DetectEmotionalShiftsInput represents input for detect_emotional_shifts tool.
type DetectEmotionalShiftsInput struct {
	MemoryIDs []string `json:"memory_ids"`
	Threshold float64  `json:"threshold,omitempty"`
}

// SummarizeSentimentInput represents input for summarize_sentiment tool.
type SummarizeSentimentInput struct {
	MemoryIDs []string `json:"memory_ids"`
}

// RegisterNLPTools registers all NLP-related MCP tools.
func (s *MCPServer) RegisterNLPTools() {
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "extract_entities_advanced",
		Description: "Extract named entities from text or memories using transformer models. Returns entities with confidence scores and relationships. Supports PERSON, ORGANIZATION, LOCATION, DATE, EVENT, PRODUCT, TECHNOLOGY, CONCEPT, OTHER.",
	}, s.handleExtractEntitiesAdvanced)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "analyze_sentiment",
		Description: "Analyze sentiment using transformer models. Returns sentiment label (POSITIVE/NEGATIVE/NEUTRAL/MIXED), confidence, emotional tone (joy, sadness, anger, fear, surprise, disgust), and intensity.",
	}, s.handleAnalyzeSentiment)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "extract_topics",
		Description: "Extract topics from memories using LDA/NMF algorithms. Returns topics with keywords, documents, coherence and diversity scores.",
	}, s.handleExtractTopics)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "analyze_sentiment_trend",
		Description: "Analyze sentiment trends over time. Returns sentiment evolution with moving averages.",
	}, s.handleAnalyzeSentimentTrend)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "detect_emotional_shifts",
		Description: "Detect significant emotional shifts. Returns timestamps, magnitude and direction of changes.",
	}, s.handleDetectEmotionalShifts)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "summarize_sentiment",
		Description: "Create sentiment summary. Returns statistics, dominant sentiment, emotional profile and trends.",
	}, s.handleSummarizeSentiment)
}

// Tool handlers

func (s *MCPServer) handleExtractEntitiesAdvanced(ctx context.Context, req *sdk.CallToolRequest, input ExtractEntitiesAdvancedInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	if s.entityExtractor == nil {
		return nil, nil, errors.New("entity extraction not enabled (set NEXS_NLP_ENTITY_EXTRACTION_ENABLED=true)")
	}

	// Check which input was provided
	if input.Text != "" {
		// Extract from raw text
		result, err := s.entityExtractor.ExtractFromText(ctx, input.Text)
		if err != nil {
			return nil, nil, fmt.Errorf("entity extraction failed: %w", err)
		}

		return nil, map[string]interface{}{
			"entities":        result.Entities,
			"relationships":   result.Relationships,
			"entity_count":    len(result.Entities),
			"processing_time": result.ProcessingTime,
			"model_used":      result.ModelUsed,
			"confidence":      result.Confidence,
			"success":         true,
		}, nil
	}

	if input.MemoryID != "" {
		// Extract from single memory
		result, err := s.entityExtractor.ExtractFromMemory(ctx, input.MemoryID)
		if err != nil {
			return nil, nil, fmt.Errorf("entity extraction failed: %w", err)
		}

		return nil, map[string]interface{}{
			"entities":        result.Entities,
			"relationships":   result.Relationships,
			"entity_count":    len(result.Entities),
			"processing_time": result.ProcessingTime,
			"model_used":      result.ModelUsed,
			"confidence":      result.Confidence,
			"success":         true,
		}, nil
	}

	if len(input.MemoryIDs) > 0 {
		// Extract from multiple memories
		results, err := s.entityExtractor.ExtractFromMemoryBatch(ctx, input.MemoryIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("batch entity extraction failed: %w", err)
		}

		return nil, map[string]interface{}{
			"results":      results,
			"memory_count": len(input.MemoryIDs),
			"success":      true,
		}, nil
	}

	return nil, nil, errors.New("one of text, memory_id, or memory_ids is required")
}

func (s *MCPServer) handleAnalyzeSentiment(ctx context.Context, req *sdk.CallToolRequest, input AnalyzeSentimentInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	if s.sentimentAnalyzer == nil {
		return nil, nil, errors.New("sentiment analysis not enabled (set NEXS_NLP_SENTIMENT_ENABLED=true)")
	}

	// Check which input was provided
	if input.Text != "" {
		// Analyze raw text (create temporary memory)
		result, err := s.sentimentAnalyzer.AnalyzeText(ctx, input.Text)
		if err != nil {
			return nil, nil, fmt.Errorf("sentiment analysis failed: %w", err)
		}

		return nil, map[string]interface{}{
			"label":              result.Label,
			"confidence":         result.Confidence,
			"scores":             result.Scores,
			"intensity":          result.Intensity,
			"emotional_tone":     result.EmotionalTone,
			"subjectivity_score": result.SubjectivityScore,
			"processing_time":    result.ProcessingTime,
			"model_used":         result.ModelUsed,
			"success":            true,
		}, nil
	}

	if input.MemoryID != "" {
		// Analyze single memory
		result, err := s.sentimentAnalyzer.AnalyzeMemory(ctx, input.MemoryID)
		if err != nil {
			return nil, nil, fmt.Errorf("sentiment analysis failed: %w", err)
		}

		return nil, map[string]interface{}{
			"label":              result.Label,
			"confidence":         result.Confidence,
			"scores":             result.Scores,
			"intensity":          result.Intensity,
			"emotional_tone":     result.EmotionalTone,
			"subjectivity_score": result.SubjectivityScore,
			"processing_time":    result.ProcessingTime,
			"model_used":         result.ModelUsed,
			"success":            true,
		}, nil
	}

	if len(input.MemoryIDs) > 0 {
		// Analyze multiple memories
		results, err := s.sentimentAnalyzer.AnalyzeMemoryBatch(ctx, input.MemoryIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("batch sentiment analysis failed: %w", err)
		}

		return nil, map[string]interface{}{
			"results":      results,
			"memory_count": len(input.MemoryIDs),
			"success":      true,
		}, nil
	}

	return nil, nil, errors.New("one of text, memory_id, or memory_ids is required")
}

func (s *MCPServer) handleExtractTopics(ctx context.Context, req *sdk.CallToolRequest, input ExtractTopicsInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	if len(input.MemoryIDs) == 0 {
		return nil, nil, errors.New("memory_ids is required")
	}

	// Set defaults
	if input.NumTopics == 0 {
		input.NumTopics = 5
	}
	if input.Algorithm == "" {
		input.Algorithm = "lda"
	}

	config := application.TopicModelingConfig{
		Algorithm:        input.Algorithm,
		NumTopics:        input.NumTopics,
		MaxIterations:    100,
		MinWordFrequency: 2,
		MaxWordFrequency: 0.8,
		TopKeywords:      10,
		UseONNX:          false,
	}

	modeler := application.NewTopicModeler(config, s.repo, nil)

	topics, err := modeler.ExtractTopics(ctx, input.MemoryIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("topic extraction failed: %w", err)
	}

	return nil, map[string]interface{}{
		"topics":       topics,
		"topic_count":  len(topics),
		"algorithm":    input.Algorithm,
		"memory_count": len(input.MemoryIDs),
		"success":      true,
	}, nil
}

func (s *MCPServer) handleAnalyzeSentimentTrend(ctx context.Context, req *sdk.CallToolRequest, input AnalyzeSentimentTrendInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	if s.sentimentAnalyzer == nil {
		return nil, nil, errors.New("sentiment analysis not enabled")
	}

	if len(input.MemoryIDs) == 0 {
		return nil, nil, errors.New("memory_ids is required")
	}

	trend, err := s.sentimentAnalyzer.AnalyzeMemoryTrend(ctx, input.MemoryIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("trend analysis failed: %w", err)
	}

	return nil, map[string]interface{}{
		"trend":        trend,
		"memory_count": len(input.MemoryIDs),
		"success":      true,
	}, nil
}

func (s *MCPServer) handleDetectEmotionalShifts(ctx context.Context, req *sdk.CallToolRequest, input DetectEmotionalShiftsInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	if s.sentimentAnalyzer == nil {
		return nil, nil, errors.New("sentiment analysis not enabled")
	}

	if len(input.MemoryIDs) == 0 {
		return nil, nil, errors.New("memory_ids is required")
	}

	// Set default threshold
	threshold := input.Threshold
	if threshold == 0 {
		threshold = 0.3 // Default: 30% change is significant
	}

	shifts, err := s.sentimentAnalyzer.DetectEmotionalShifts(ctx, input.MemoryIDs, threshold)
	if err != nil {
		return nil, nil, fmt.Errorf("shift detection failed: %w", err)
	}

	return nil, map[string]interface{}{
		"shifts":       shifts,
		"shift_count":  len(shifts),
		"threshold":    threshold,
		"memory_count": len(input.MemoryIDs),
		"success":      true,
	}, nil
}

func (s *MCPServer) handleSummarizeSentiment(ctx context.Context, req *sdk.CallToolRequest, input SummarizeSentimentInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	if s.sentimentAnalyzer == nil {
		return nil, nil, errors.New("sentiment analysis not enabled")
	}

	if len(input.MemoryIDs) == 0 {
		return nil, nil, errors.New("memory_ids is required")
	}

	summary, err := s.sentimentAnalyzer.SummarizeMemorySentiments(ctx, input.MemoryIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("sentiment summarization failed: %w", err)
	}

	return nil, map[string]interface{}{
		"summary":      summary,
		"memory_count": len(input.MemoryIDs),
		"success":      true,
	}, nil
}
