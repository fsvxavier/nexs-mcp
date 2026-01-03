package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds the application configuration.
type Config struct {
	// StorageType can be "memory" or "file"
	StorageType string

	// DataDir is the directory for file-based storage
	DataDir string

	// BaseDir is the base directory for all nexs-mcp data
	// Metrics, performance logs, HNSW index, etc. will be stored here
	// Default: derived from DataDir parent (e.g., if DataDir=~/.nexs-mcp/elements, BaseDir=~/.nexs-mcp)
	BaseDir string

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

	// TokenMetricsSaveInterval is the interval for auto-saving token metrics
	// Default: 5 minutes
	TokenMetricsSaveInterval time.Duration

	// MetricsSaveInterval is the interval for auto-saving performance metrics
	// Default: 5 minutes
	MetricsSaveInterval time.Duration

	// WorkingMemory configuration
	WorkingMemory WorkingMemoryConfig

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

	// VectorStore configuration
	VectorStore VectorStoreConfig

	// DuplicateDetection configuration
	DuplicateDetection DuplicateDetectionConfig

	// Clustering configuration
	Clustering ClusteringConfig

	// KnowledgeGraph configuration
	KnowledgeGraph KnowledgeGraphConfig

	// MemoryConsolidation configuration
	MemoryConsolidation MemoryConsolidationConfig

	// HybridSearch configuration
	HybridSearch HybridSearchConfig

	// MemoryRetention configuration
	MemoryRetention MemoryRetentionConfig

	// ContextEnrichment configuration
	ContextEnrichment ContextEnrichmentConfig

	// Embeddings configuration
	Embeddings EmbeddingsConfig

	// SkillExtraction configuration
	SkillExtraction SkillExtractionConfig
}

// WorkingMemoryConfig holds configuration for working memory persistence.
type WorkingMemoryConfig struct {
	// PersistenceEnabled controls whether working memories are persisted to disk
	// Default: true
	PersistenceEnabled bool

	// PersistenceDir is the directory where working memories are stored
	// Default: ~/.nexs-mcp/working_memory
	PersistenceDir string
}

// SkillExtractionConfig holds configuration for skill extraction from personas.
type SkillExtractionConfig struct {
	// Enabled controls whether automatic skill extraction is active
	// Default: true
	Enabled bool

	// AutoExtractOnCreate automatically extracts skills when creating a persona
	// Default: false
	AutoExtractOnCreate bool

	// SkipDuplicates skips creating skills that already exist
	// Default: true
	SkipDuplicates bool

	// MinSkillNameLength is the minimum skill name length to extract
	// Default: 3 characters
	MinSkillNameLength int

	// MaxSkillsPerPersona is the maximum skills to extract per persona
	// Default: 50 (0 = unlimited)
	MaxSkillsPerPersona int

	// ExtractFromExpertiseAreas enables extraction from expertise_areas field
	// Default: true
	ExtractFromExpertiseAreas bool

	// ExtractFromCustomFields enables extraction from custom technical_skills fields
	// Default: true
	ExtractFromCustomFields bool

	// AutoUpdatePersona automatically updates persona with skill references
	// Default: true
	AutoUpdatePersona bool
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

// VectorStoreConfig holds configuration for vector storage and search.
type VectorStoreConfig struct {
	// Dimension is the vector embedding dimension
	// Default: 384 (common for MiniLM, MPNet models)
	Dimension int

	// Similarity metric: "cosine", "euclidean", "dotproduct"
	// Default: cosine
	Similarity string

	// HybridThreshold is the vector count to switch from linear to HNSW
	// Default: 100 vectors
	HybridThreshold int

	// HNSW configuration
	HNSW HNSWConfig
}

// HNSWConfig holds configuration for HNSW index.
type HNSWConfig struct {
	// Enabled controls whether HNSW is used (vs linear-only)
	// Default: true
	Enabled bool

	// M is the number of bi-directional links per node
	// Range: 8-64, Default: 16
	// Higher M = better recall, more memory
	M int

	// Ml is the level generation multiplier
	// Range: 0.1-0.5, Default: 0.25
	// Controls hierarchical layer structure
	Ml float64

	// EfSearch is the search candidate list size
	// Range: 10-200, Default: 20
	// Higher EfSearch = better recall, slower search
	EfSearch int

	// Seed for random number generation
	// Default: 42
	Seed int64
}

// DuplicateDetectionConfig holds configuration for duplicate detection.
type DuplicateDetectionConfig struct {
	// Enabled controls whether duplicate detection is active
	// Default: true
	Enabled bool

	// SimilarityThreshold is the minimum similarity to consider duplicates (0.0-1.0)
	// Default: 0.95 (95% similar)
	SimilarityThreshold float32

	// MinContentLength is the minimum content length to check for duplicates
	// Default: 20 characters
	MinContentLength int

	// MaxResults is the maximum number of duplicate groups to return
	// Default: 100
	MaxResults int
}

// ClusteringConfig holds configuration for clustering algorithm.
type ClusteringConfig struct {
	// Enabled controls whether clustering is active
	// Default: true
	Enabled bool

	// Algorithm specifies the clustering algorithm: "dbscan" or "kmeans"
	// Default: dbscan
	Algorithm string

	// MinClusterSize is the minimum memories per cluster (DBSCAN)
	// Default: 3
	MinClusterSize int

	// EpsilonDistance is the distance threshold for DBSCAN (0.0-1.0)
	// Default: 0.15 (15% distance)
	EpsilonDistance float32

	// NumClusters is the number of clusters for K-means
	// Default: 10
	NumClusters int

	// MaxIterations is the max iterations for K-means
	// Default: 100
	MaxIterations int
}

// KnowledgeGraphConfig holds configuration for knowledge graph extraction.
type KnowledgeGraphConfig struct {
	// Enabled controls whether knowledge graph extraction is active
	// Default: true
	Enabled bool

	// ExtractPeople enables person name extraction
	// Default: true
	ExtractPeople bool

	// ExtractOrganizations enables organization extraction
	// Default: true
	ExtractOrganizations bool

	// ExtractURLs enables URL extraction
	// Default: true
	ExtractURLs bool

	// ExtractEmails enables email extraction
	// Default: true
	ExtractEmails bool

	// ExtractConcepts enables concept/entity extraction
	// Default: true
	ExtractConcepts bool

	// ExtractKeywords enables keyword extraction
	// Default: true
	ExtractKeywords bool

	// MaxKeywords is the maximum keywords to extract per memory
	// Default: 10
	MaxKeywords int

	// ExtractRelationships enables relationship extraction
	// Default: true
	ExtractRelationships bool

	// MaxRelationships is the maximum relationships to extract
	// Default: 20
	MaxRelationships int
}

// MemoryConsolidationConfig holds configuration for memory consolidation.
type MemoryConsolidationConfig struct {
	// Enabled controls whether memory consolidation is active
	// Default: true
	Enabled bool

	// AutoConsolidate enables automatic consolidation on schedule
	// Default: false
	AutoConsolidate bool

	// ConsolidationInterval is the time between auto-consolidations
	// Default: 24 hours
	ConsolidationInterval time.Duration

	// MinMemoriesForConsolidation is the minimum memories to trigger consolidation
	// Default: 10
	MinMemoriesForConsolidation int

	// EnableDuplicateDetection includes duplicate detection in workflow
	// Default: true
	EnableDuplicateDetection bool

	// EnableClustering includes clustering in workflow
	// Default: true
	EnableClustering bool

	// EnableKnowledgeExtraction includes knowledge graph extraction
	// Default: true
	EnableKnowledgeExtraction bool

	// EnableQualityScoring includes quality scoring in workflow
	// Default: true
	EnableQualityScoring bool
}

// HybridSearchConfig holds configuration for hybrid search.
type HybridSearchConfig struct {
	// Enabled controls whether hybrid search is active
	// Default: true
	Enabled bool

	// Mode specifies search mode: "hnsw", "linear", "auto"
	// Default: auto
	Mode string

	// SimilarityThreshold is the minimum similarity for results (0.0-1.0)
	// Default: 0.7 (70% similar)
	SimilarityThreshold float32

	// MaxResults is the maximum search results to return
	// Default: 10
	MaxResults int

	// AutoSwitchThreshold is vector count to switch from linear to HNSW
	// Default: 100 vectors
	AutoSwitchThreshold int

	// IndexPersistence enables saving HNSW index to disk
	// Default: true
	IndexPersistence bool

	// IndexPath is the directory to save HNSW index
	// Default: data/hnsw-index
	IndexPath string
}

// MemoryRetentionConfig holds configuration for memory retention.
type MemoryRetentionConfig struct {
	// Enabled controls whether memory retention is active
	// Default: true
	Enabled bool

	// QualityThreshold is the minimum quality score to retain (0.0-1.0)
	// Default: 0.5
	QualityThreshold float32

	// HighQualityRetentionDays is retention for high-quality memories
	// Default: 365 days
	HighQualityRetentionDays int

	// MediumQualityRetentionDays is retention for medium-quality memories
	// Default: 180 days
	MediumQualityRetentionDays int

	// LowQualityRetentionDays is retention for low-quality memories
	// Default: 90 days
	LowQualityRetentionDays int

	// AutoCleanup enables automatic cleanup of old memories
	// Default: false
	AutoCleanup bool

	// CleanupInterval is the time between cleanup cycles
	// Default: 24 hours
	CleanupInterval time.Duration
}

// ContextEnrichmentConfig holds configuration for context enrichment.
type ContextEnrichmentConfig struct {
	// Enabled controls whether context enrichment is active
	// Default: true
	Enabled bool

	// MaxRelatedMemories is the max related memories to include
	// Default: 5
	MaxRelatedMemories int

	// MaxDepth is the maximum relationship depth to traverse
	// Default: 2
	MaxDepth int

	// IncludeRelationships includes relationship metadata
	// Default: true
	IncludeRelationships bool

	// IncludeTimestamps includes temporal information
	// Default: true
	IncludeTimestamps bool

	// SimilarityThreshold for related memory inclusion (0.0-1.0)
	// Default: 0.6
	SimilarityThreshold float32
}

// EmbeddingsConfig holds configuration for embeddings.
type EmbeddingsConfig struct {
	// Provider specifies the embedding provider: "openai", "transformers", "onnx", "sentence"
	// Default: onnx
	Provider string

	// Dimension is the embedding vector dimension
	// Default: 384
	Dimension int

	// CacheEnabled enables embedding cache
	// Default: true
	CacheEnabled bool

	// CacheTTL is the cache time-to-live
	// Default: 24 hours
	CacheTTL time.Duration

	// CacheSize is the maximum cache entries
	// Default: 10000
	CacheSize int

	// BatchSize for batch embedding operations
	// Default: 32
	BatchSize int

	// OpenAI configuration
	OpenAI OpenAIEmbeddingConfig

	// Transformers configuration
	Transformers TransformersEmbeddingConfig

	// ONNX configuration
	ONNX ONNXEmbeddingConfig
}

// OpenAIEmbeddingConfig holds OpenAI embedding configuration.
type OpenAIEmbeddingConfig struct {
	// APIKey is the OpenAI API key
	APIKey string

	// Model is the embedding model to use
	// Default: text-embedding-3-small
	Model string
}

// TransformersEmbeddingConfig holds Transformers embedding configuration.
type TransformersEmbeddingConfig struct {
	// ModelPath is the path to the ONNX model
	ModelPath string
}

// ONNXEmbeddingConfig holds ONNX embedding configuration.
type ONNXEmbeddingConfig struct {
	// ModelPath is the path to the ONNX model
	ModelPath string
}

// LoadConfig loads configuration from environment variables and command-line flags.
func LoadConfig(version string) *Config {
	cfg := &Config{
		ServerName:               getEnvOrDefault("NEXS_SERVER_NAME", "nexs-mcp"),
		Version:                  version,
		AutoSaveMemories:         getEnvBool("NEXS_AUTO_SAVE_MEMORIES", true),
		AutoSaveInterval:         getEnvDuration("NEXS_AUTO_SAVE_INTERVAL", 30*time.Second),
		TokenMetricsSaveInterval: getEnvDuration("NEXS_TOKEN_METRICS_SAVE_INTERVAL", 30*time.Second),
		MetricsSaveInterval:      getEnvDuration("NEXS_METRICS_SAVE_INTERVAL", 5*time.Minute),
		WorkingMemory: WorkingMemoryConfig{
			PersistenceEnabled: getEnvBool("NEXS_WORKING_MEMORY_PERSISTENCE", true),
			PersistenceDir:     getEnvOrDefault("NEXS_WORKING_MEMORY_DIR", "~/.nexs-mcp/working_memory"),
		},
		Resources: ResourcesConfig{
			Enabled:  getEnvBool("NEXS_RESOURCES_ENABLED", true),
			Expose:   []string{},
			CacheTTL: getEnvDuration("NEXS_RESOURCES_CACHE_TTL", 5*time.Minute),
		},
		Compression: CompressionConfig{
			Enabled:          getEnvBool("NEXS_COMPRESSION_ENABLED", true),
			Algorithm:        getEnvOrDefault("NEXS_COMPRESSION_ALGORITHM", "gzip"),
			MinSize:          getEnvInt("NEXS_COMPRESSION_MIN_SIZE", 1024),
			CompressionLevel: getEnvInt("NEXS_COMPRESSION_LEVEL", 6),
			AdaptiveMode:     getEnvBool("NEXS_COMPRESSION_ADAPTIVE", true),
		},
		Streaming: StreamingConfig{
			Enabled:      getEnvBool("NEXS_STREAMING_ENABLED", true),
			ChunkSize:    getEnvInt("NEXS_STREAMING_CHUNK_SIZE", 10),
			ThrottleRate: getEnvDuration("NEXS_STREAMING_THROTTLE", 50*time.Millisecond),
			BufferSize:   getEnvInt("NEXS_STREAMING_BUFFER_SIZE", 100),
		},
		Summarization: SummarizationConfig{
			Enabled:              getEnvBool("NEXS_SUMMARIZATION_ENABLED", true),
			AgeBeforeSummarize:   getEnvDuration("NEXS_SUMMARIZATION_AGE", 7*24*time.Hour),
			MaxSummaryLength:     getEnvInt("NEXS_SUMMARIZATION_MAX_LENGTH", 500),
			CompressionRatio:     getEnvFloat("NEXS_SUMMARIZATION_RATIO", 0.3),
			PreserveKeywords:     getEnvBool("NEXS_SUMMARIZATION_PRESERVE_KEYWORDS", true),
			UseExtractiveSummary: getEnvBool("NEXS_SUMMARIZATION_EXTRACTIVE", true),
		},
		AdaptiveCache: AdaptiveCacheConfig{
			Enabled: getEnvBool("NEXS_ADAPTIVE_CACHE_ENABLED", true),
			MinTTL:  getEnvDuration("NEXS_ADAPTIVE_CACHE_MIN_TTL", 1*time.Hour),
			MaxTTL:  getEnvDuration("NEXS_ADAPTIVE_CACHE_MAX_TTL", 7*24*time.Hour),
			BaseTTL: getEnvDuration("NEXS_ADAPTIVE_CACHE_BASE_TTL", 24*time.Hour),
		},
		PromptCompression: PromptCompressionConfig{
			Enabled:                getEnvBool("NEXS_PROMPT_COMPRESSION_ENABLED", true),
			RemoveRedundancy:       getEnvBool("NEXS_PROMPT_COMPRESSION_REMOVE_REDUNDANCY", true),
			CompressWhitespace:     getEnvBool("NEXS_PROMPT_COMPRESSION_WHITESPACE", true),
			UseAliases:             getEnvBool("NEXS_PROMPT_COMPRESSION_ALIASES", true),
			PreserveStructure:      getEnvBool("NEXS_PROMPT_COMPRESSION_PRESERVE_STRUCTURE", true),
			TargetCompressionRatio: getEnvFloat("NEXS_PROMPT_COMPRESSION_RATIO", 0.65),
			MinPromptLength:        getEnvInt("NEXS_PROMPT_COMPRESSION_MIN_LENGTH", 500),
		},
		VectorStore: VectorStoreConfig{
			Dimension:       getEnvInt("NEXS_VECTOR_DIMENSION", 384),
			Similarity:      getEnvOrDefault("NEXS_VECTOR_SIMILARITY", "cosine"),
			HybridThreshold: getEnvInt("NEXS_VECTOR_HYBRID_THRESHOLD", 100),
			HNSW: HNSWConfig{
				Enabled:  getEnvBool("NEXS_HNSW_ENABLED", true),
				M:        getEnvInt("NEXS_HNSW_M", 16),
				Ml:       getEnvFloat("NEXS_HNSW_ML", 0.25),
				EfSearch: getEnvInt("NEXS_HNSW_EF_SEARCH", 20),
				Seed:     getEnvInt64("NEXS_HNSW_SEED", 42),
			},
		},
		DuplicateDetection: DuplicateDetectionConfig{
			Enabled:             getEnvBool("NEXS_DUPLICATE_DETECTION_ENABLED", true),
			SimilarityThreshold: getEnvFloat32("NEXS_DUPLICATE_DETECTION_THRESHOLD", 0.95),
			MinContentLength:    getEnvInt("NEXS_DUPLICATE_DETECTION_MIN_LENGTH", 20),
			MaxResults:          getEnvInt("NEXS_DUPLICATE_DETECTION_MAX_RESULTS", 100),
		},
		Clustering: ClusteringConfig{
			Enabled:         getEnvBool("NEXS_CLUSTERING_ENABLED", true),
			Algorithm:       getEnvOrDefault("NEXS_CLUSTERING_ALGORITHM", "dbscan"),
			MinClusterSize:  getEnvInt("NEXS_CLUSTERING_MIN_SIZE", 3),
			EpsilonDistance: getEnvFloat32("NEXS_CLUSTERING_EPSILON", 0.15),
			NumClusters:     getEnvInt("NEXS_CLUSTERING_NUM_CLUSTERS", 10),
			MaxIterations:   getEnvInt("NEXS_CLUSTERING_MAX_ITERATIONS", 100),
		},
		KnowledgeGraph: KnowledgeGraphConfig{
			Enabled:              getEnvBool("NEXS_KNOWLEDGE_GRAPH_ENABLED", true),
			ExtractPeople:        getEnvBool("NEXS_KNOWLEDGE_GRAPH_EXTRACT_PEOPLE", true),
			ExtractOrganizations: getEnvBool("NEXS_KNOWLEDGE_GRAPH_EXTRACT_ORGS", true),
			ExtractURLs:          getEnvBool("NEXS_KNOWLEDGE_GRAPH_EXTRACT_URLS", true),
			ExtractEmails:        getEnvBool("NEXS_KNOWLEDGE_GRAPH_EXTRACT_EMAILS", true),
			ExtractConcepts:      getEnvBool("NEXS_KNOWLEDGE_GRAPH_EXTRACT_CONCEPTS", true),
			ExtractKeywords:      getEnvBool("NEXS_KNOWLEDGE_GRAPH_EXTRACT_KEYWORDS", true),
			MaxKeywords:          getEnvInt("NEXS_KNOWLEDGE_GRAPH_MAX_KEYWORDS", 10),
			ExtractRelationships: getEnvBool("NEXS_KNOWLEDGE_GRAPH_EXTRACT_RELATIONSHIPS", true),
			MaxRelationships:     getEnvInt("NEXS_KNOWLEDGE_GRAPH_MAX_RELATIONSHIPS", 20),
		},
		MemoryConsolidation: MemoryConsolidationConfig{
			Enabled:                     getEnvBool("NEXS_MEMORY_CONSOLIDATION_ENABLED", true),
			AutoConsolidate:             getEnvBool("NEXS_MEMORY_CONSOLIDATION_AUTO", true),
			ConsolidationInterval:       getEnvDuration("NEXS_MEMORY_CONSOLIDATION_INTERVAL", 24*time.Hour),
			MinMemoriesForConsolidation: getEnvInt("NEXS_MEMORY_CONSOLIDATION_MIN_MEMORIES", 10),
			EnableDuplicateDetection:    getEnvBool("NEXS_MEMORY_CONSOLIDATION_DUPLICATES", true),
			EnableClustering:            getEnvBool("NEXS_MEMORY_CONSOLIDATION_CLUSTERING", true),
			EnableKnowledgeExtraction:   getEnvBool("NEXS_MEMORY_CONSOLIDATION_KNOWLEDGE", true),
			EnableQualityScoring:        getEnvBool("NEXS_MEMORY_CONSOLIDATION_QUALITY", true),
		},
		HybridSearch: HybridSearchConfig{
			Enabled:             getEnvBool("NEXS_HYBRID_SEARCH_ENABLED", true),
			Mode:                getEnvOrDefault("NEXS_HYBRID_SEARCH_MODE", "auto"),
			SimilarityThreshold: getEnvFloat32("NEXS_HYBRID_SEARCH_THRESHOLD", 0.7),
			MaxResults:          getEnvInt("NEXS_HYBRID_SEARCH_MAX_RESULTS", 10),
			AutoSwitchThreshold: getEnvInt("NEXS_HYBRID_SEARCH_AUTO_SWITCH", 100),
			IndexPersistence:    getEnvBool("NEXS_HYBRID_SEARCH_PERSISTENCE", true),
			IndexPath:           getEnvOrDefault("NEXS_HYBRID_SEARCH_INDEX_PATH", "data/hnsw-index"),
		},
		MemoryRetention: MemoryRetentionConfig{
			Enabled:                    getEnvBool("NEXS_MEMORY_RETENTION_ENABLED", true),
			QualityThreshold:           getEnvFloat32("NEXS_MEMORY_RETENTION_THRESHOLD", 0.5),
			HighQualityRetentionDays:   getEnvInt("NEXS_MEMORY_RETENTION_HIGH_DAYS", 365),
			MediumQualityRetentionDays: getEnvInt("NEXS_MEMORY_RETENTION_MEDIUM_DAYS", 180),
			LowQualityRetentionDays:    getEnvInt("NEXS_MEMORY_RETENTION_LOW_DAYS", 90),
			AutoCleanup:                getEnvBool("NEXS_MEMORY_RETENTION_AUTO_CLEANUP", false),
			CleanupInterval:            getEnvDuration("NEXS_MEMORY_RETENTION_CLEANUP_INTERVAL", 24*time.Hour),
		},
		ContextEnrichment: ContextEnrichmentConfig{
			Enabled:              getEnvBool("NEXS_CONTEXT_ENRICHMENT_ENABLED", true),
			MaxRelatedMemories:   getEnvInt("NEXS_CONTEXT_ENRICHMENT_MAX_MEMORIES", 5),
			MaxDepth:             getEnvInt("NEXS_CONTEXT_ENRICHMENT_MAX_DEPTH", 2),
			IncludeRelationships: getEnvBool("NEXS_CONTEXT_ENRICHMENT_RELATIONSHIPS", true),
			IncludeTimestamps:    getEnvBool("NEXS_CONTEXT_ENRICHMENT_TIMESTAMPS", true),
			SimilarityThreshold:  getEnvFloat32("NEXS_CONTEXT_ENRICHMENT_THRESHOLD", 0.6),
		},
		Embeddings: EmbeddingsConfig{
			Provider:     getEnvOrDefault("NEXS_EMBEDDINGS_PROVIDER", "onnx"),
			Dimension:    getEnvInt("NEXS_EMBEDDINGS_DIMENSION", 384),
			CacheEnabled: getEnvBool("NEXS_EMBEDDINGS_CACHE_ENABLED", true),
			CacheTTL:     getEnvDuration("NEXS_EMBEDDINGS_CACHE_TTL", 24*time.Hour),
			CacheSize:    getEnvInt("NEXS_EMBEDDINGS_CACHE_SIZE", 10000),
			BatchSize:    getEnvInt("NEXS_EMBEDDINGS_BATCH_SIZE", 32),
			OpenAI: OpenAIEmbeddingConfig{
				APIKey: getEnvOrDefault("NEXS_EMBEDDINGS_OPENAI_API_KEY", ""),
				Model:  getEnvOrDefault("NEXS_EMBEDDINGS_OPENAI_MODEL", "text-embedding-3-small"),
			},
			Transformers: TransformersEmbeddingConfig{
				ModelPath: getEnvOrDefault("NEXS_EMBEDDINGS_TRANSFORMERS_MODEL_PATH", ""),
			},
			ONNX: ONNXEmbeddingConfig{
				ModelPath: getEnvOrDefault("NEXS_EMBEDDINGS_ONNX_MODEL_PATH", ""),
			},
		},
		SkillExtraction: SkillExtractionConfig{
			Enabled:                   getEnvBool("NEXS_SKILL_EXTRACTION_ENABLED", true),
			AutoExtractOnCreate:       getEnvBool("NEXS_SKILL_EXTRACTION_AUTO_ON_CREATE", true),
			SkipDuplicates:            getEnvBool("NEXS_SKILL_EXTRACTION_SKIP_DUPLICATES", true),
			MinSkillNameLength:        getEnvInt("NEXS_SKILL_EXTRACTION_MIN_NAME_LENGTH", 3),
			MaxSkillsPerPersona:       getEnvInt("NEXS_SKILL_EXTRACTION_MAX_PER_PERSONA", 50),
			ExtractFromExpertiseAreas: getEnvBool("NEXS_SKILL_EXTRACTION_FROM_EXPERTISE", true),
			ExtractFromCustomFields:   getEnvBool("NEXS_SKILL_EXTRACTION_FROM_CUSTOM", true),
			AutoUpdatePersona:         getEnvBool("NEXS_SKILL_EXTRACTION_AUTO_UPDATE", true),
		},
	}

	// Define command-line flags
	flag.StringVar(&cfg.StorageType, "storage", getEnvOrDefault("NEXS_STORAGE_TYPE", "file"),
		"Storage type: 'memory' or 'file'")
	flag.StringVar(&cfg.DataDir, "data-dir", getEnvOrDefault("NEXS_DATA_DIR", "~/.nexs-mcp/elements"),
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
	flag.BoolVar(&cfg.WorkingMemory.PersistenceEnabled, "working-memory-persistence", cfg.WorkingMemory.PersistenceEnabled,
		"Enable working memory persistence to disk (default: true)")
	flag.StringVar(&cfg.WorkingMemory.PersistenceDir, "working-memory-dir", cfg.WorkingMemory.PersistenceDir,
		"Directory for working memory persistence (default: ~/.nexs-mcp/working_memory)")
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
	flag.BoolVar(&cfg.DuplicateDetection.Enabled, "duplicate-detection-enabled", cfg.DuplicateDetection.Enabled,
		"Enable duplicate detection (default: true)")
	flag.BoolVar(&cfg.Clustering.Enabled, "clustering-enabled", cfg.Clustering.Enabled,
		"Enable memory clustering (default: true)")
	flag.StringVar(&cfg.Clustering.Algorithm, "clustering-algorithm", cfg.Clustering.Algorithm,
		"Clustering algorithm: dbscan or kmeans (default: dbscan)")
	flag.BoolVar(&cfg.KnowledgeGraph.Enabled, "knowledge-graph-enabled", cfg.KnowledgeGraph.Enabled,
		"Enable knowledge graph extraction (default: true)")
	flag.BoolVar(&cfg.MemoryConsolidation.Enabled, "memory-consolidation-enabled", cfg.MemoryConsolidation.Enabled,
		"Enable memory consolidation (default: true)")
	flag.BoolVar(&cfg.MemoryConsolidation.AutoConsolidate, "memory-consolidation-auto", cfg.MemoryConsolidation.AutoConsolidate,
		"Enable automatic consolidation on schedule (default: false)")
	flag.DurationVar(&cfg.MemoryConsolidation.ConsolidationInterval, "memory-consolidation-interval", cfg.MemoryConsolidation.ConsolidationInterval,
		"Interval between auto-consolidations (default: 24h)")
	flag.BoolVar(&cfg.HybridSearch.Enabled, "hybrid-search-enabled", cfg.HybridSearch.Enabled,
		"Enable hybrid search (default: true)")
	flag.StringVar(&cfg.HybridSearch.Mode, "hybrid-search-mode", cfg.HybridSearch.Mode,
		"Search mode: hnsw, linear, or auto (default: auto)")
	flag.BoolVar(&cfg.MemoryRetention.Enabled, "memory-retention-enabled", cfg.MemoryRetention.Enabled,
		"Enable memory retention (default: true)")
	flag.BoolVar(&cfg.SkillExtraction.Enabled, "skill-extraction-enabled", cfg.SkillExtraction.Enabled,
		"Enable skill extraction from personas (default: true)")
	flag.BoolVar(&cfg.SkillExtraction.AutoExtractOnCreate, "skill-extraction-auto-on-create", cfg.SkillExtraction.AutoExtractOnCreate,
		"Automatically extract skills when creating a persona (default: false)")
	flag.BoolVar(&cfg.MemoryRetention.AutoCleanup, "memory-retention-auto-cleanup", cfg.MemoryRetention.AutoCleanup,
		"Enable automatic cleanup of old memories (default: false)")
	flag.BoolVar(&cfg.ContextEnrichment.Enabled, "context-enrichment-enabled", cfg.ContextEnrichment.Enabled,
		"Enable context enrichment (default: true)")
	flag.IntVar(&cfg.ContextEnrichment.MaxRelatedMemories, "context-enrichment-max-memories", cfg.ContextEnrichment.MaxRelatedMemories,
		"Maximum related memories to include (default: 5)")
	flag.BoolVar(&cfg.Embeddings.CacheEnabled, "embeddings-cache-enabled", cfg.Embeddings.CacheEnabled,
		"Enable embeddings cache (default: true)")
	flag.StringVar(&cfg.Embeddings.Provider, "embeddings-provider", cfg.Embeddings.Provider,
		"Embeddings provider: openai, transformers, onnx, or sentence (default: onnx)")

	flag.Parse()

	// Derive BaseDir from DataDir if not explicitly set
	// If DataDir = ~/.nexs-mcp/elements -> BaseDir = ~/.nexs-mcp
	// If DataDir = /app/data -> BaseDir = /app/data (parent or same)
	cfg.BaseDir = getEnvOrDefault("NEXS_BASE_DIR", "")
	if cfg.BaseDir == "" {
		cfg.BaseDir = deriveBaseDir(cfg.DataDir)
	}

	return cfg
}

// deriveBaseDir derives the base directory from the data directory.
// If DataDir ends with "/elements", removes it. Otherwise uses the parent directory.
func deriveBaseDir(dataDir string) string {
	if dataDir == "" {
		return "~/.nexs-mcp"
	}

	// Expand ~ if present
	if len(dataDir) > 0 && dataDir[0] == '~' {
		if home := os.Getenv("HOME"); home != "" {
			dataDir = home + dataDir[1:]
		}
	}

	// If ends with /elements, use parent
	if strings.HasSuffix(dataDir, "/elements") || strings.HasSuffix(dataDir, "\\elements") {
		return filepath.Dir(dataDir)
	}

	// Otherwise use parent directory
	return filepath.Dir(dataDir)
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

// getEnvInt64 returns an int64 environment variable value or a default value.
func getEnvInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result int64
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}

// getEnvFloat32 returns a float32 environment variable value or a default value.
func getEnvFloat32(key string, defaultValue float32) float32 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result float32
	_, err := fmt.Sscanf(value, "%f", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
