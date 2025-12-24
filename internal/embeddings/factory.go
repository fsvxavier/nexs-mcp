package embeddings

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fsvxavier/nexs-mcp/internal/embeddings/providers"
)

// Factory creates embedding providers based on configuration.
type Factory struct {
	config Config
}

// NewFactory creates a new provider factory.
func NewFactory(config Config) *Factory {
	return &Factory{config: config}
}

// NewFactoryFromEnv creates a factory from environment variables.
func NewFactoryFromEnv() *Factory {
	config := DefaultConfig()

	// Override from environment
	if provider := os.Getenv("NEXS_EMBEDDING_PROVIDER"); provider != "" {
		config.Provider = provider
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		config.OpenAIKey = key
	}
	if model := os.Getenv("NEXS_OPENAI_MODEL"); model != "" {
		config.OpenAIModel = model
	}
	if cache := os.Getenv("NEXS_TRANSFORMERS_CACHE"); cache != "" {
		config.TransformersCache = cache
	}
	if os.Getenv("NEXS_USE_GPU") == "true" {
		config.UseGPU = true
	}

	return NewFactory(config)
}

// Create creates a provider based on configuration.
func (f *Factory) Create(ctx context.Context) (Provider, error) {
	providerName := f.config.Provider
	if providerName == "auto" {
		return f.createAuto(ctx)
	}

	return f.createProvider(ctx, providerName)
}

// CreateWithFallback creates a provider with automatic fallback support.
func (f *Factory) CreateWithFallback(ctx context.Context) (Provider, error) {
	if !f.config.EnableFallback {
		return f.Create(ctx)
	}

	var lastErr error
	priority := f.config.FallbackPriority
	if len(priority) == 0 {
		priority = []string{"transformers", "sentence", "onnx", "openai"}
	}

	for _, providerName := range priority {
		provider, err := f.createProvider(ctx, providerName)
		if err != nil {
			lastErr = err
			continue
		}

		if provider.IsAvailable(ctx) {
			return &FallbackProvider{
				primary:  provider,
				factory:  f,
				priority: priority,
			}, nil
		}
	}

	return nil, fmt.Errorf("no available embedding providers: %w", lastErr)
}

func (f *Factory) createAuto(ctx context.Context) (Provider, error) {
	// Try providers in order of preference: free first, paid last
	tryOrder := []string{"transformers", "sentence", "onnx", "openai"}

	for _, name := range tryOrder {
		provider, err := f.createProvider(ctx, name)
		if err != nil {
			continue
		}
		if provider.IsAvailable(ctx) {
			return provider, nil
		}
	}

	return nil, errors.New("no embedding providers available")
}

func (f *Factory) createProvider(ctx context.Context, name string) (Provider, error) {
	switch name {
	case "openai":
		return providers.NewOpenAI(providers.OpenAIConfig{
			APIKey:  f.config.OpenAIKey,
			Model:   f.config.OpenAIModel,
			Timeout: f.config.OpenAITimeout,
		})

	case "transformers":
		return providers.NewTransformers(providers.TransformersConfig{
			Model:    f.config.TransformersModel,
			CacheDir: f.config.TransformersCache,
			UseGPU:   f.config.UseGPU,
		})

	case "sentence":
		return providers.NewSentenceTransformers(providers.SentenceConfig{
			Model:    f.config.SentenceModel,
			CacheDir: f.config.SentenceCache,
			UseGPU:   f.config.UseGPU,
		})

	case "onnx":
		return providers.NewONNX(providers.ONNXConfig{
			Model:    f.config.ONNXModel,
			CacheDir: f.config.ONNXCache,
			UseGPU:   f.config.UseGPU,
		})

	default:
		return nil, fmt.Errorf("unknown embedding provider: %s", name)
	}
}

// FallbackProvider wraps a provider with automatic fallback.
type FallbackProvider struct {
	primary  Provider
	factory  *Factory
	priority []string
}

func (f *FallbackProvider) Name() string {
	return f.primary.Name() + "-fallback"
}

func (f *FallbackProvider) Dimensions() int {
	return f.primary.Dimensions()
}

func (f *FallbackProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	result, err := f.primary.Embed(ctx, text)
	if err == nil {
		return result, nil
	}

	// Try fallback providers
	for _, name := range f.priority {
		if name == f.primary.Name() {
			continue
		}

		provider, err := f.factory.createProvider(ctx, name)
		if err != nil {
			continue
		}

		if !provider.IsAvailable(ctx) {
			continue
		}

		// Check dimensions match
		if provider.Dimensions() != f.primary.Dimensions() {
			continue
		}

		result, err := provider.Embed(ctx, text)
		if err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("all embedding providers failed: %w", err)
}

func (f *FallbackProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result, err := f.primary.EmbedBatch(ctx, texts)
	if err == nil {
		return result, nil
	}

	// Try fallback providers
	for _, name := range f.priority {
		if name == f.primary.Name() {
			continue
		}

		provider, err := f.factory.createProvider(ctx, name)
		if err != nil {
			continue
		}

		if !provider.IsAvailable(ctx) {
			continue
		}

		if provider.Dimensions() != f.primary.Dimensions() {
			continue
		}

		result, err := provider.EmbedBatch(ctx, texts)
		if err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("all embedding providers failed for batch: %w", err)
}

func (f *FallbackProvider) IsAvailable(ctx context.Context) bool {
	if f.primary.IsAvailable(ctx) {
		return true
	}

	for _, name := range f.priority {
		if name == f.primary.Name() {
			continue
		}

		provider, err := f.factory.createProvider(ctx, name)
		if err != nil {
			continue
		}

		if provider.IsAvailable(ctx) {
			return true
		}
	}

	return false
}

func (f *FallbackProvider) Cost() float64 {
	return f.primary.Cost()
}
