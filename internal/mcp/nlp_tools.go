package mcp

import (
	"context"
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
	// TODO: Implement when ONNX provider is ready
	return nil, nil, fmt.Errorf("entity extraction not yet implemented - ONNX provider integration pending")
}

func (s *MCPServer) handleAnalyzeSentiment(ctx context.Context, req *sdk.CallToolRequest, input AnalyzeSentimentInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	// TODO: Implement when ONNX provider is ready
	return nil, nil, fmt.Errorf("sentiment analysis not yet implemented - ONNX provider integration pending")
}

func (s *MCPServer) handleExtractTopics(ctx context.Context, req *sdk.CallToolRequest, input ExtractTopicsInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	if len(input.MemoryIDs) == 0 {
		return nil, nil, fmt.Errorf("memory_ids is required")
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
	// TODO: Implement when ONNX provider is ready
	return nil, nil, fmt.Errorf("sentiment trend analysis not yet implemented - ONNX provider integration pending")
}

func (s *MCPServer) handleDetectEmotionalShifts(ctx context.Context, req *sdk.CallToolRequest, input DetectEmotionalShiftsInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	// TODO: Implement when ONNX provider is ready
	return nil, nil, fmt.Errorf("emotional shift detection not yet implemented - ONNX provider integration pending")
}

func (s *MCPServer) handleSummarizeSentiment(ctx context.Context, req *sdk.CallToolRequest, input SummarizeSentimentInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	// TODO: Implement when ONNX provider is ready
	return nil, nil, fmt.Errorf("sentiment summarization not yet implemented - ONNX provider integration pending")
}
