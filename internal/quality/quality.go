package quality

import (
	"context"
	"time"
)

// Score represents a quality score for a memory
type Score struct {
	Value      float64                `json:"value"`
	Confidence float64                `json:"confidence"`
	Method     string                 `json:"method"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Scorer is the interface for quality scoring implementations
type Scorer interface {
	Score(ctx context.Context, content string) (*Score, error)
	ScoreBatch(ctx context.Context, contents []string) ([]*Score, error)
	Name() string
	IsAvailable(ctx context.Context) bool
	Cost() float64
	Close() error
}

// RetentionPolicy defines how long memories should be retained based on quality
type RetentionPolicy struct {
	MinQuality       float64 `json:"min_quality"`
	MaxQuality       float64 `json:"max_quality"`
	RetentionDays    int     `json:"retention_days"`
	ArchiveAfterDays int     `json:"archive_after_days"`
	Description      string  `json:"description"`
}

// ImplicitSignals represents signals that can be used for quality estimation
type ImplicitSignals struct {
	AccessCount    int     `json:"access_count"`
	ReferenceCount int     `json:"reference_count"`
	AgeInDays      int     `json:"age_in_days"`
	LastAccessDays int     `json:"last_access_days"`
	UserRating     float64 `json:"user_rating"`
	ContentLength  int     `json:"content_length"`
	TagCount       int     `json:"tag_count"`
	IsPromoted     bool    `json:"is_promoted"`
}

// Config holds configuration for the quality system
type Config struct {
	DefaultScorer          string            `json:"default_scorer"`
	EnableFallback         bool              `json:"enable_fallback"`
	FallbackChain          []string          `json:"fallback_chain"`
	ONNXModelPath          string            `json:"onnx_model_path"`
	RequiresTokenTypeIds   bool              `json:"requires_token_type_ids"` // true for BERT, false for DistilBERT/RoBERTa
	ONNXModelType          string            `json:"onnx_model_type"`         // "reranker" or "embedder"
	ONNXOutputName         string            `json:"onnx_output_name"`        // "logits", "last_hidden_state", etc.
	ONNXOutputShape        []int64           `json:"onnx_output_shape"`       // [1, 1] for reranker, [1, 384/512/768] for embedder
	GroqAPIKey             string            `json:"groq_api_key"`
	GeminiAPIKey           string            `json:"gemini_api_key"`
	RetentionPolicies      []RetentionPolicy `json:"retention_policies"`
	EnableAutoArchival     bool              `json:"enable_auto_archival"`
	CleanupIntervalMinutes int               `json:"cleanup_interval_minutes"`
}

// DefaultRetentionPolicies returns the standard retention policies
func DefaultRetentionPolicies() []RetentionPolicy {
	return []RetentionPolicy{
		{
			MinQuality:       0.7,
			MaxQuality:       1.1,
			RetentionDays:    365,
			ArchiveAfterDays: 180,
			Description:      "High quality - retained for 1 year, archived after 6 months",
		},
		{
			MinQuality:       0.5,
			MaxQuality:       0.7,
			RetentionDays:    180,
			ArchiveAfterDays: 90,
			Description:      "Medium quality - retained for 6 months, archived after 3 months",
		},
		{
			MinQuality:       0.0,
			MaxQuality:       0.5,
			RetentionDays:    90,
			ArchiveAfterDays: 30,
			Description:      "Low quality - retained for 3 months, archived after 1 month",
		},
	}
}

// GetRetentionPolicy returns the appropriate policy for a quality score
func GetRetentionPolicy(score float64, policies []RetentionPolicy) *RetentionPolicy {
	for i := range policies {
		p := &policies[i]
		if score >= p.MinQuality && score < p.MaxQuality {
			return p
		}
	}
	if len(policies) > 0 {
		return &policies[len(policies)-1]
	}
	return nil
}

// DefaultConfig returns the default quality system configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultScorer:          "onnx",
		EnableFallback:         true,
		FallbackChain:          []string{"onnx", "groq", "gemini", "implicit"},
		ONNXModelPath:          "models/ms-marco-MiniLM-L-6-v2.onnx",
		RetentionPolicies:      DefaultRetentionPolicies(),
		EnableAutoArchival:     true,
		CleanupIntervalMinutes: 60,
	}
}
