package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Config holds the application configuration.
type Config struct {
	// StorageType can be "memory" or "file"
	StorageType string

	// DataDir is the directory for file-based storage
	DataDir string

	// ServerName is the name of the MCP server
	ServerName string

	// Version is the application version
	Version string

	// LogLevel is the logging level (debug, info, warn, error)
	LogLevel string

	// LogFormat is the log output format (json, text)
	LogFormat string

	// AutoSaveMemories enables automatic saving of conversation context as memories
	// Default: true
	AutoSaveMemories bool

	// AutoSaveInterval is the minimum time between auto-saves
	// Default: 5 minutes
	AutoSaveInterval time.Duration

	// Resources configuration
	Resources ResourcesConfig

	// Compression configuration
	Compression CompressionConfig

	// Streaming configuration
	Streaming StreamingConfig

	// Summarization configuration
	Summarization SummarizationConfig

	// AdaptiveCache configuration
	AdaptiveCache AdaptiveCacheConfig

	// PromptCompression configuration
	PromptCompression PromptCompressionConfig
}

// ResourcesConfig holds configuration for MCP Resources Protocol.
type ResourcesConfig struct {
	// Enabled controls whether resources are exposed to clients
	// Default: false (resources disabled for safety)
	Enabled bool

	// Expose lists which resource URIs to expose
	// Empty means expose all resources when Enabled=true
	// Example: ["capability-index://summary", "capability-index://stats"]
	Expose []string

	// CacheTTL is the duration to cache resource content
	// Default: 5 minutes
	CacheTTL time.Duration
}

// CompressionConfig holds configuration for response compression.
type CompressionConfig struct {
	// Enabled controls whether response compression is active
	// Default: false
	Enabled bool

	// Algorithm specifies the compression algorithm (gzip, zlib)
	// Default: gzip
	Algorithm string

	// MinSize is the minimum payload size to compress (bytes)
	// Default: 1024 (1KB)
	MinSize int

	// CompressionLevel is the compression level (1-9)
	// Default: 6 (balanced)
	CompressionLevel int

	// AdaptiveMode enables automatic algorithm selection
	// Default: true
	AdaptiveMode bool
}

// StreamingConfig holds configuration for streaming responses.
type StreamingConfig struct {
	// Enabled controls whether streaming responses are active
	// Default: false
	Enabled bool

	// ChunkSize is the number of items per chunk
	// Default: 10
	ChunkSize int

	// ThrottleRate is the delay between chunks
	// Default: 50ms
	ThrottleRate time.Duration

	// BufferSize is the channel buffer size
	// Default: 100
	BufferSize int
}

// SummarizationConfig holds configuration for automatic summarization.
type SummarizationConfig struct {
	// Enabled controls whether auto-summarization is active
	// Default: false
	Enabled bool

	// AgeBeforeSummarize is the minimum age before summarizing
	// Default: 7 days
	AgeBeforeSummarize time.Duration

	// MaxSummaryLength is the maximum summary length in characters
	// Default: 500
	MaxSummaryLength int

	// CompressionRatio is the target compression ratio (0.0-1.0)
	// Default: 0.3 (70% reduction)
	CompressionRatio float64

	// PreserveKeywords preserves extracted keywords in summary
	// Default: true
	PreserveKeywords bool

	// UseExtractiveSummary uses extractive vs abstractive summarization
	// Default: true
	UseExtractiveSummary bool
}

// AdaptiveCacheConfig holds configuration for adaptive cache TTL.
type AdaptiveCacheConfig struct {
	// Enabled controls whether adaptive cache is active
	// Default: false (uses standard LRU cache)
	Enabled bool

	// MinTTL is the minimum cache TTL
	// Default: 1 hour
	MinTTL time.Duration

	// MaxTTL is the maximum cache TTL
	// Default: 7 days
	MaxTTL time.Duration

	// BaseTTL is the baseline cache TTL
	// Default: 24 hours
	BaseTTL time.Duration
}

// PromptCompressionConfig holds configuration for prompt compression.
type PromptCompressionConfig struct {
	// Enabled controls whether prompt compression is active
	// Default: false
	Enabled bool

	// RemoveRedundancy removes syntactic redundancies
	// Default: true
	RemoveRedundancy bool

	// CompressWhitespace normalizes whitespace
	// Default: true
	CompressWhitespace bool

	// UseAliases replaces verbose phrases with aliases
	// Default: true
	UseAliases bool

	// PreserveStructure maintains JSON/YAML structure
	// Default: true
	PreserveStructure bool

	// TargetCompressionRatio is the target compression ratio
	// Default: 0.65 (35% reduction)
	TargetCompressionRatio float64

	// MinPromptLength only compresses prompts longer than this
	// Default: 500 characters
	MinPromptLength int
}

// LoadConfig loads configuration from environment variables and command-line flags.
func LoadConfig(version string) *Config {
	cfg := &Config{
		ServerName:       getEnvOrDefault("NEXS_SERVER_NAME", "nexs-mcp"),
		Version:          version,
		AutoSaveMemories: getEnvBool("NEXS_AUTO_SAVE_MEMORIES", true),
		AutoSaveInterval: getEnvDuration("NEXS_AUTO_SAVE_INTERVAL", 5*time.Minute),
		Resources: ResourcesConfig{
			Enabled:  getEnvBool("NEXS_RESOURCES_ENABLED", false),
			Expose:   []string{},
			CacheTTL: getEnvDuration("NEXS_RESOURCES_CACHE_TTL", 5*time.Minute),
		},
		Compression: CompressionConfig{
			Enabled:          getEnvBool("NEXS_COMPRESSION_ENABLED", false),
			Algorithm:        getEnvOrDefault("NEXS_COMPRESSION_ALGORITHM", "gzip"),
			MinSize:          getEnvInt("NEXS_COMPRESSION_MIN_SIZE", 1024),
			CompressionLevel: getEnvInt("NEXS_COMPRESSION_LEVEL", 6),
			AdaptiveMode:     getEnvBool("NEXS_COMPRESSION_ADAPTIVE", true),
		},
		Streaming: StreamingConfig{
			Enabled:      getEnvBool("NEXS_STREAMING_ENABLED", false),
			ChunkSize:    getEnvInt("NEXS_STREAMING_CHUNK_SIZE", 10),
			ThrottleRate: getEnvDuration("NEXS_STREAMING_THROTTLE", 50*time.Millisecond),
			BufferSize:   getEnvInt("NEXS_STREAMING_BUFFER_SIZE", 100),
		},
		Summarization: SummarizationConfig{
			Enabled:              getEnvBool("NEXS_SUMMARIZATION_ENABLED", false),
			AgeBeforeSummarize:   getEnvDuration("NEXS_SUMMARIZATION_AGE", 7*24*time.Hour),
			MaxSummaryLength:     getEnvInt("NEXS_SUMMARIZATION_MAX_LENGTH", 500),
			CompressionRatio:     getEnvFloat("NEXS_SUMMARIZATION_RATIO", 0.3),
			PreserveKeywords:     getEnvBool("NEXS_SUMMARIZATION_PRESERVE_KEYWORDS", true),
			UseExtractiveSummary: getEnvBool("NEXS_SUMMARIZATION_EXTRACTIVE", true),
		},
		AdaptiveCache: AdaptiveCacheConfig{
			Enabled: getEnvBool("NEXS_ADAPTIVE_CACHE_ENABLED", false),
			MinTTL:  getEnvDuration("NEXS_ADAPTIVE_CACHE_MIN_TTL", 1*time.Hour),
			MaxTTL:  getEnvDuration("NEXS_ADAPTIVE_CACHE_MAX_TTL", 7*24*time.Hour),
			BaseTTL: getEnvDuration("NEXS_ADAPTIVE_CACHE_BASE_TTL", 24*time.Hour),
		},
		PromptCompression: PromptCompressionConfig{
			Enabled:                getEnvBool("NEXS_PROMPT_COMPRESSION_ENABLED", false),
			RemoveRedundancy:       getEnvBool("NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY", true),
			CompressWhitespace:     getEnvBool("NEXS_PROMPT_COMPRESSION_WHITESPACE", true),
			UseAliases:             getEnvBool("NEXS_PROMPT_COMPRESSION_ALIASES", true),
			PreserveStructure:      getEnvBool("NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE", true),
			TargetCompressionRatio: getEnvFloat("NEXS_PROMPT_COMPRESSION_RATIO", 0.65),
			MinPromptLength:        getEnvInt("NEXS_PROMPT_COMPRESSION_MIN_LENGTH", 500),
		},
	}

	// Define command-line flags
	flag.StringVar(&cfg.StorageType, "storage", getEnvOrDefault("NEXS_STORAGE_TYPE", "file"),
		"Storage type: 'memory' or 'file'")
	flag.StringVar(&cfg.DataDir, "data-dir", getEnvOrDefault("NEXS_DATA_DIR", "data/elements"),
		"Directory for file-based storage")
	flag.StringVar(&cfg.LogLevel, "log-level", getEnvOrDefault("NEXS_LOG_LEVEL", "info"),
		"Log level: 'debug', 'info', 'warn', 'error'")
	flag.StringVar(&cfg.LogFormat, "log-format", getEnvOrDefault("NEXS_LOG_FORMAT", "json"),
		"Log format: 'json' or 'text'")
	flag.BoolVar(&cfg.Resources.Enabled, "resources-enabled", cfg.Resources.Enabled,
		"Enable MCP Resources Protocol (default: false)")
	flag.DurationVar(&cfg.Resources.CacheTTL, "resources-cache-ttl", cfg.Resources.CacheTTL,
		"Cache TTL for resource content")
	flag.BoolVar(&cfg.AutoSaveMemories, "auto-save-memories", cfg.AutoSaveMemories,
		"Enable automatic saving of conversation context as memories (default: true)")
	flag.DurationVar(&cfg.AutoSaveInterval, "auto-save-interval", cfg.AutoSaveInterval,
		"Minimum interval between auto-saves (default: 5m)")
	flag.BoolVar(&cfg.Compression.Enabled, "compression-enabled", cfg.Compression.Enabled,
		"Enable response compression (default: false)")
	flag.StringVar(&cfg.Compression.Algorithm, "compression-algorithm", cfg.Compression.Algorithm,
		"Compression algorithm: gzip or zlib (default: gzip)")
	flag.IntVar(&cfg.Compression.CompressionLevel, "compression-level", cfg.Compression.CompressionLevel,
		"Compression level 1-9 (default: 6)")
	flag.BoolVar(&cfg.Streaming.Enabled, "streaming-enabled", cfg.Streaming.Enabled,
		"Enable streaming responses (default: false)")
	flag.IntVar(&cfg.Streaming.ChunkSize, "streaming-chunk-size", cfg.Streaming.ChunkSize,
		"Number of items per chunk (default: 10)")
	flag.BoolVar(&cfg.Summarization.Enabled, "summarization-enabled", cfg.Summarization.Enabled,
		"Enable automatic summarization (default: false)")
	flag.BoolVar(&cfg.AdaptiveCache.Enabled, "adaptive-cache-enabled", cfg.AdaptiveCache.Enabled,
		"Enable adaptive cache TTL (default: false)")
	flag.BoolVar(&cfg.PromptCompression.Enabled, "prompt-compression-enabled", cfg.PromptCompression.Enabled,
		"Enable prompt compression (default: false)")

	flag.Parse()

	return cfg
}

// getEnvOrDefault returns an environment variable value or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns a boolean environment variable value or a default value.
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

// getEnvDuration returns a duration environment variable value or a default value.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

// getEnvInt returns an integer environment variable value or a default value.
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}

// getEnvFloat returns a float environment variable value or a default value.
func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result float64
	_, err := fmt.Sscanf(value, "%f", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
