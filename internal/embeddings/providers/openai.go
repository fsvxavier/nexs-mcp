// Package providers implements concrete embedding providers.
package providers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

// OpenAIConfig holds OpenAI-specific configuration
type OpenAIConfig struct {
	APIKey  string
	Model   string // text-embedding-3-small, text-embedding-3-large, text-embedding-ada-002
	Timeout int    // seconds
}

// OpenAIProvider implements embeddings using OpenAI API
type OpenAIProvider struct {
	client *openai.Client
	config OpenAIConfig
	dims   int
	cost   float64
}

// NewOpenAI creates a new OpenAI embedding provider
func NewOpenAI(config OpenAIConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	if config.Model == "" {
		config.Model = "text-embedding-3-small"
	}

	if config.Timeout == 0 {
		config.Timeout = 30
	}

	client := openai.NewClient(config.APIKey)

	// Set dimensions and cost based on model
	dims, cost := getModelSpecs(config.Model)

	return &OpenAIProvider{
		client: client,
		config: config,
		dims:   dims,
		cost:   cost,
	}, nil
}

func getModelSpecs(model string) (dims int, cost float64) {
	switch model {
	case "text-embedding-3-small":
		return 1536, 0.00002 // $0.02 per 1M tokens
	case "text-embedding-3-large":
		return 3072, 0.00013 // $0.13 per 1M tokens
	case "text-embedding-ada-002":
		return 1536, 0.0001 // $0.10 per 1M tokens
	default:
		return 1536, 0.00002
	}
}

func (o *OpenAIProvider) Name() string {
	return "openai"
}

func (o *OpenAIProvider) Dimensions() int {
	return o.dims
}

func (o *OpenAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, errors.New("empty text provided")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(o.config.Timeout)*time.Second)
	defer cancel()

	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(o.config.Model),
	}

	resp, err := o.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no embeddings returned from OpenAI")
	}

	return resp.Data[0].Embedding, nil
}

func (o *OpenAIProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("empty text batch")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(o.config.Timeout)*time.Second)
	defer cancel()

	req := openai.EmbeddingRequest{
		Input: texts,
		Model: openai.EmbeddingModel(o.config.Model),
	}

	resp, err := o.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API batch error: %w", err)
	}

	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(resp.Data))
	}

	result := make([][]float32, len(texts))
	for i, data := range resp.Data {
		result[i] = data.Embedding
	}

	return result, nil
}

func (o *OpenAIProvider) IsAvailable(ctx context.Context) bool {
	if o.config.APIKey == "" {
		return false
	}

	// Quick health check: try to embed a single word
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := o.Embed(ctx, "test")
	return err == nil
}

func (o *OpenAIProvider) Cost() float64 {
	return o.cost
}
