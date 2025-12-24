// Package embeddings provides vector embedding generation with multiple provider support.
package embeddings

import (
	"context"
	"time"
)

// Provider defines the interface for embedding generation services.
// Implementations include OpenAI, Local Transformers, Sentence Transformers, and ONNX.
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "transformers")
	Name() string

	// Dimensions returns the vector dimensionality for this provider
	Dimensions() int

	// Embed generates embeddings for a single text
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts efficiently
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// IsAvailable checks if the provider is ready to use
	IsAvailable(ctx context.Context) bool

	// Cost returns the estimated cost per 1000 tokens (0.0 for local providers)
	Cost() float64
}

// Config holds configuration for embedding providers.
type Config struct {
	// Provider selects which embedding provider to use
	// Options: "openai", "transformers", "sentence", "onnx", "auto"
	Provider string `json:"provider"`

	// OpenAI specific
	OpenAIKey     string `json:"openai_key,omitempty"`
	OpenAIModel   string `json:"openai_model,omitempty"` // default: text-embedding-3-small
	OpenAITimeout int    `json:"openai_timeout,omitempty"`

	// Local Transformers specific
	TransformersModel string `json:"transformers_model,omitempty"` // default: all-MiniLM-L6-v2
	TransformersCache string `json:"transformers_cache,omitempty"`

	// Sentence Transformers specific
	SentenceModel string `json:"sentence_model,omitempty"` // default: paraphrase-multilingual
	SentenceCache string `json:"sentence_cache,omitempty"`

	// ONNX specific
	ONNXModel string `json:"onnx_model,omitempty"` // default: ms-marco-MiniLM
	ONNXCache string `json:"onnx_cache,omitempty"`
	UseGPU    bool   `json:"use_gpu,omitempty"`

	// Cache configuration
	EnableCache  bool          `json:"enable_cache"`
	CacheTTL     time.Duration `json:"cache_ttl"`
	CacheMaxSize int           `json:"cache_max_size"` // Max cached embeddings

	// Fallback configuration
	EnableFallback   bool     `json:"enable_fallback"`
	FallbackPriority []string `json:"fallback_priority"` // e.g., ["openai", "transformers", "onnx"]
}

// DefaultConfig returns sensible defaults for embedding configuration.
func DefaultConfig() Config {
	return Config{
		Provider:          "transformers", // Default to free, offline-capable provider
		OpenAIModel:       "text-embedding-3-small",
		OpenAITimeout:     30,
		TransformersModel: "all-MiniLM-L6-v2",
		SentenceModel:     "paraphrase-multilingual-MiniLM-L12-v2",
		ONNXModel:         "ms-marco-MiniLM-L-12-v2",
		EnableCache:       true,
		CacheTTL:          24 * time.Hour,
		CacheMaxSize:      10000,
		EnableFallback:    true,
		FallbackPriority:  []string{"transformers", "sentence", "onnx", "openai"},
	}
}

// Stats tracks provider usage statistics.
type Stats struct {
	Provider        string        `json:"provider"`
	TotalEmbeddings int64         `json:"total_embeddings"`
	TotalTokens     int64         `json:"total_tokens"`
	TotalCost       float64       `json:"total_cost"`
	AvgLatency      time.Duration `json:"avg_latency"`
	CacheHits       int64         `json:"cache_hits"`
	CacheMisses     int64         `json:"cache_misses"`
	Errors          int64         `json:"errors"`
	LastUsed        time.Time     `json:"last_used"`
}

// SimilarityMetric defines how to calculate vector similarity.
type SimilarityMetric string

const (
	// CosineSimilarity measures cosine similarity (default, normalized).
	CosineSimilarity SimilarityMetric = "cosine"

	// EuclideanDistance measures L2 distance.
	EuclideanDistance SimilarityMetric = "euclidean"

	// DotProduct measures dot product similarity.
	DotProduct SimilarityMetric = "dotproduct"
)

// Result represents a semantic search result.
type Result struct {
	ID         string                 `json:"id"`
	Score      float64                `json:"score"`
	Text       string                 `json:"text"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Embedding  []float32              `json:"-"` // Not serialized
	Distance   float64                `json:"distance,omitempty"`
	Similarity float64                `json:"similarity"`
}
