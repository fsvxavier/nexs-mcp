package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/collection"
	"github.com/fsvxavier/nexs-mcp/internal/config"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
	"github.com/fsvxavier/nexs-mcp/internal/mcp/resources"
)

// MCPServer wraps the official MCP SDK server.
type MCPServer struct {
	server               *sdk.Server
	repo                 domain.ElementRepository
	metrics              *application.MetricsCollector
	perfMetrics          *logger.PerformanceMetrics
	hybridSearch         *application.HybridSearchService
	relationshipIndex    *application.RelationshipIndex
	recommendationEngine *application.RecommendationEngine
	inferenceEngine      *application.RelationshipInferenceEngine
	workingMemory        *application.WorkingMemoryService
	retentionService     *application.MemoryRetentionService
	temporalService      *application.TemporalService
	registry             *collection.Registry
	mu                   sync.Mutex
	deviceCodes          map[string]string // Maps user codes to device codes for GitHub OAuth
	capabilityResource   *resources.CapabilityIndexResource
	resourcesConfig      config.ResourcesConfig
	cfg                  *config.Config // Store config for auto-save checks
	// Token optimization services
	compressor           *ResponseCompressor
	streamingHandler     *StreamingHandler
	summarizationService *application.SummarizationService
	deduplicationService *application.SemanticDeduplicationService
	contextWindowManager *application.ContextWindowManager
	promptCompressor     *application.PromptCompressor
	adaptiveCache        *application.AdaptiveCacheService
	tokenMetrics         *application.TokenMetricsCollector
	responseMiddleware   *ResponseMiddleware
	githubClient         infrastructure.GitHubClientInterface
	// Auto-save state
	autoSaveTicker   *time.Ticker
	autoSaveStopChan chan struct{}
	lastAutoSave     time.Time
	// Alert management
	alertManager *AlertManager
	// Analytics and optimization services
	forecastingService  *application.CostForecastingService
	optimizationEngine  *application.OptimizationEngine
	userCostAttribution *application.UserCostAttributionService
	// NLP services (Sprint 18)
	entityExtractor   *application.EnhancedEntityExtractor
	sentimentAnalyzer *application.SentimentAnalyzer
	topicModeler      *application.TopicModeler
}

// NewMCPServer creates a new MCP server using the official SDK.
func NewMCPServer(name, version string, repo domain.ElementRepository, cfg *config.Config) *MCPServer {
	impl := &sdk.Implementation{
		Name:    name,
		Version: version,
	}

	// Create server with default capabilities
	server := sdk.NewServer(impl, nil)

	// Create metrics collector - use BaseDir (global or workspace)
	metricsDir := filepath.Join(cfg.BaseDir, "metrics")
	metrics := application.NewMetricsCollector(metricsDir, cfg.MetricsSaveInterval)

	// Create performance metrics - use BaseDir (global or workspace)
	perfDir := filepath.Join(cfg.BaseDir, "performance")
	perfMetrics := logger.NewPerformanceMetrics(perfDir)

	// Create embedding provider (default: transformers)
	// Check if we're in test mode
	var provider embeddings.Provider
	var err error
	if os.Getenv("NEXS_TEST_MODE") == "1" {
		// Use mock provider for tests
		provider = embeddings.NewMockProvider("mock-test", 384)
	} else {
		// Use real provider in production
		factoryConfig := embeddings.Config{
			Provider: "transformers", // Default to local transformers
		}
		factory := embeddings.NewFactory(factoryConfig)
		provider, err = factory.Create(context.Background())
		if err != nil || provider == nil {
			// Fallback to mock if initialization fails
			provider = embeddings.NewMockProvider("mock", 384)
		}
	}

	// Create HNSW-based hybrid search service - use BaseDir (global or workspace)
	hnswPath := filepath.Join(cfg.BaseDir, "hnsw_index.json")
	hybridSearch := application.NewHybridSearchService(application.HybridSearchConfig{
		Provider:        provider,
		HNSWPath:        hnswPath,
		AutoReindex:     true,
		ReindexInterval: 100,
	})

	// Try to load existing HNSW index
	_ = hybridSearch.LoadIndex() // Ignore error if index doesn't exist yet

	// Create relationship index for bidirectional search
	relationshipIndex := application.NewRelationshipIndex()

	// Create recommendation engine
	recommendationEngine := application.NewRecommendationEngine(repo, relationshipIndex)

	// Create relationship inference engine with hybrid search
	inferenceEngine := application.NewRelationshipInferenceEngine(repo, relationshipIndex, hybridSearch)

	// Create working memory service for two-tier memory architecture
	// Use configured persistence directory or default to BaseDir/working_memory
	workingMemoryDir := cfg.WorkingMemory.PersistenceDir
	if workingMemoryDir == "" {
		// Use BaseDir - same location as metrics, performance, hnsw
		workingMemoryDir = filepath.Join(cfg.BaseDir, "working_memory")
	}

	workingMemory := application.NewWorkingMemoryServiceWithPersistence(
		repo,
		workingMemoryDir,
		cfg.WorkingMemory.PersistenceEnabled,
	)

	// Create temporal service for version history and time travel
	temporalConfig := application.DefaultTemporalConfig()
	temporalService := application.NewTemporalService(temporalConfig, logger.Get())

	// Create capability index resource (compatibility wrapper)
	capabilityResource := resources.NewCapabilityIndexResource(repo, nil, cfg.Resources.CacheTTL)

	// Create collection registry
	registry := collection.NewRegistry()

	// Create token optimization services
	var algo CompressionAlgorithm
	if cfg.Compression.Algorithm == "zlib" {
		algo = CompressionZlib
	} else {
		algo = CompressionGzip
	}

	compressor := NewResponseCompressor(CompressionConfig{
		Enabled:          cfg.Compression.Enabled,
		Algorithm:        algo,
		MinSize:          cfg.Compression.MinSize,
		CompressionLevel: cfg.Compression.CompressionLevel,
	})

	streamingHandler := NewStreamingHandler(StreamingConfig{
		Enabled:      cfg.Streaming.Enabled,
		ChunkSize:    cfg.Streaming.ChunkSize,
		ThrottleRate: cfg.Streaming.ThrottleRate,
		BufferSize:   cfg.Streaming.BufferSize,
		MaxChunks:    10,
	})

	summarizationService := application.NewSummarizationService(application.SummarizationConfig{
		Enabled:              cfg.Summarization.Enabled,
		AgeBeforeSummarize:   cfg.Summarization.AgeBeforeSummarize,
		MaxSummaryLength:     cfg.Summarization.MaxSummaryLength,
		CompressionRatio:     cfg.Summarization.CompressionRatio,
		PreserveKeywords:     cfg.Summarization.PreserveKeywords,
		UseExtractiveSummary: cfg.Summarization.UseExtractiveSummary,
	})

	deduplicationService := application.NewSemanticDeduplicationService(application.DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.92,
		MergeStrategy:       application.MergeKeepFirst,
		PreserveMetadata:    true,
		BatchSize:           100,
	})

	contextWindowManager := application.NewContextWindowManager(application.ContextWindowConfig{
		MaxTokens:          8000,
		PriorityStrategy:   application.PriorityHybrid,
		TruncationMethod:   application.TruncationHead,
		PreserveRecent:     5,
		RelevanceThreshold: 0.3,
	})

	promptCompressor := application.NewPromptCompressor(application.PromptCompressionConfig{
		Enabled:                cfg.PromptCompression.Enabled,
		RemoveRedundancy:       cfg.PromptCompression.RemoveRedundancy,
		CompressWhitespace:     cfg.PromptCompression.CompressWhitespace,
		UseAliases:             cfg.PromptCompression.UseAliases,
		PreserveStructure:      cfg.PromptCompression.PreserveStructure,
		TargetCompressionRatio: cfg.PromptCompression.TargetCompressionRatio,
		MinPromptLength:        cfg.PromptCompression.MinPromptLength,
	})

	// Create token metrics collector - use BaseDir (global or workspace)
	tokenMetricsDir := filepath.Join(cfg.BaseDir, "token_metrics")
	tokenMetrics := application.NewTokenMetricsCollector(tokenMetricsDir, cfg.TokenMetricsSaveInterval)

	// Create adaptive cache service
	adaptiveCache := application.NewAdaptiveCacheService(application.AdaptiveCacheConfig{
		Enabled: cfg.AdaptiveCache.Enabled,
		MinTTL:  cfg.AdaptiveCache.MinTTL,
		MaxTTL:  cfg.AdaptiveCache.MaxTTL,
		BaseTTL: cfg.AdaptiveCache.BaseTTL,
	})

	// Inject adaptive cache into services
	hybridSearch.SetAdaptiveCache(adaptiveCache)
	if fileRepo, ok := repo.(*infrastructure.FileElementRepository); ok {
		fileRepo.SetAdaptiveCache(adaptiveCache)
	}

	// Create analytics and optimization services
	forecastingService := application.NewCostForecastingService()
	optimizationEngine := application.NewOptimizationEngine(application.DefaultOptimizationEngineConfig())
	userCostAttribution := application.NewUserCostAttributionService(
		filepath.Join(cfg.BaseDir, "user_costs"),
		true, // autosave enabled
	)

	mcpServer := &MCPServer{
		server:               server,
		repo:                 repo,
		metrics:              metrics,
		perfMetrics:          perfMetrics,
		hybridSearch:         hybridSearch,
		relationshipIndex:    relationshipIndex,
		recommendationEngine: recommendationEngine,
		inferenceEngine:      inferenceEngine,
		workingMemory:        workingMemory,
		temporalService:      temporalService,
		registry:             registry,
		capabilityResource:   capabilityResource,
		resourcesConfig:      cfg.Resources,
		cfg:                  cfg, // Store config for auto-save checks
		// Token optimization services
		compressor:           compressor,
		streamingHandler:     streamingHandler,
		summarizationService: summarizationService,
		deduplicationService: deduplicationService,
		contextWindowManager: contextWindowManager,
		promptCompressor:     promptCompressor,
		adaptiveCache:        adaptiveCache,
		tokenMetrics:         tokenMetrics,
		autoSaveStopChan:     make(chan struct{}),
		lastAutoSave:         time.Now(),
		// Analytics and optimization
		forecastingService:  forecastingService,
		optimizationEngine:  optimizationEngine,
		userCostAttribution: userCostAttribution,
	}

	// Initialize NLP services (Sprint 18)
	if cfg.NLP.TopicModelingEnabled {
		// Topic modeling works without ONNX (classical LDA/NMF)
		topicConfig := application.TopicModelingConfig{
			Algorithm:        cfg.NLP.TopicAlgorithm,
			NumTopics:        cfg.NLP.TopicCount,
			MaxIterations:    100,
			MinWordFrequency: 2,
			MaxWordFrequency: 0.8,
			TopKeywords:      10,
			RandomSeed:       42,
			Alpha:            0.1,
			Beta:             0.01,
			UseONNX:          false, // Classical algorithms only for now
		}
		mcpServer.topicModeler = application.NewTopicModeler(topicConfig, repo, nil)
		logger.Info("Topic modeling service initialized", "algorithm", cfg.NLP.TopicAlgorithm, "topics", cfg.NLP.TopicCount)
	}

	if cfg.NLP.EntityExtractionEnabled || cfg.NLP.SentimentAnalysisEnabled {
		// Entity extraction and sentiment analysis require ONNX provider
		nlpConfig := application.EnhancedNLPConfig{
			EntityModel:         cfg.NLP.EntityModel,
			EntityConfidenceMin: cfg.NLP.EntityConfidenceMin,
			EntityMaxPerDoc:     cfg.NLP.EntityMaxPerDoc,
			SentimentModel:      cfg.NLP.SentimentModel,
			SentimentThreshold:  cfg.NLP.SentimentThreshold,
			TopicCount:          cfg.NLP.TopicCount,
			BatchSize:           cfg.NLP.BatchSize,
			MaxLength:           cfg.NLP.MaxLength,
			UseGPU:              cfg.NLP.UseGPU,
			EnableFallback:      cfg.NLP.EnableFallback,
		}

		// Create ONNX BERT provider
		onnxProvider, err := application.NewONNXBERTProvider(nlpConfig)
		if err != nil {
			logger.Warn("Failed to create ONNX provider, will use fallback methods", "error", err)
		}

		// Log ONNX availability
		onnxAvailable := onnxProvider != nil && onnxProvider.IsAvailable()
		if onnxAvailable {
			logger.Info("✅ ONNX Runtime initialized successfully",
				"status", "enabled",
				"gpu", cfg.NLP.UseGPU,
				"batch_size", cfg.NLP.BatchSize,
				"max_length", cfg.NLP.MaxLength)
		} else {
			logger.Warn("⚠️  ONNX Runtime not available, using fallback methods",
				"status", "fallback",
				"fallback_enabled", cfg.NLP.EnableFallback)
		}

		if cfg.NLP.EntityExtractionEnabled {
			mcpServer.entityExtractor = application.NewEnhancedEntityExtractor(nlpConfig, repo, onnxProvider)
			logger.Info("✅ Entity Extraction enabled",
				"status", "enabled",
				"model", cfg.NLP.EntityModel,
				"onnx", onnxAvailable,
				"confidence_min", cfg.NLP.EntityConfidenceMin,
				"max_per_doc", cfg.NLP.EntityMaxPerDoc)
		} else {
			logger.Info("⚠️  Entity Extraction disabled", "status", "disabled")
		}

		if cfg.NLP.SentimentAnalysisEnabled {
			mcpServer.sentimentAnalyzer = application.NewSentimentAnalyzer(nlpConfig, repo, onnxProvider)
			logger.Info("✅ Sentiment Analysis enabled",
				"status", "enabled",
				"model", cfg.NLP.SentimentModel,
				"onnx", onnxAvailable,
				"threshold", cfg.NLP.SentimentThreshold)
		} else {
			logger.Info("⚠️  Sentiment Analysis disabled", "status", "disabled")
		}
	}

	// Create response middleware for compression
	mcpServer.responseMiddleware = NewResponseMiddleware(mcpServer)

	// Populate index with existing elements
	mcpServer.rebuildIndex()

	// Rebuild relationship index
	ctx := context.Background()
	if err := relationshipIndex.Rebuild(ctx, repo); err != nil {
		logger.Error("Failed to rebuild relationship index", "error", err)
	}

	// Register all tools
	mcpServer.registerTools()

	// Register resources if enabled
	if cfg.Resources.Enabled {
		mcpServer.registerResources()
		logger.Info("MCP Resources Protocol enabled",
			"expose", cfg.Resources.Expose,
			"cache_ttl", cfg.Resources.CacheTTL)
	} else {
		logger.Info("MCP Resources Protocol disabled (set --resources-enabled=true to enable)")
	}

	return mcpServer
}

// registerResources registers all MCP resources.
func (s *MCPServer) registerResources() {
	handler := s.capabilityResource.Handler()

	// Check if we should expose specific resources or all
	expose := s.resourcesConfig.Expose
	shouldExposeAll := len(expose) == 0

	shouldExpose := func(uri string) bool {
		if shouldExposeAll {
			return true
		}
		for _, allowed := range expose {
			if allowed == uri {
				return true
			}
		}
		return false
	}

	// Register summary resource
	if shouldExpose(resources.URISummary) {
		s.server.AddResource(&sdk.Resource{
			URI:         resources.URISummary,
			Name:        "Capability Index Summary",
			Description: "A concise summary (~3K tokens) of the capability index including element counts, top keywords, and recent elements",
			MIMEType:    "text/markdown",
		}, handler)
		logger.Info("Registered resource", "uri", resources.URISummary)
	}

	// Register full resource
	if shouldExpose(resources.URIFull) {
		s.server.AddResource(&sdk.Resource{
			URI:         resources.URIFull,
			Name:        "Capability Index Full Details",
			Description: "Complete detailed view (~40K tokens) of the capability index with all elements, metadata, relationships, and vocabulary",
			MIMEType:    "text/markdown",
		}, handler)
		logger.Info("Registered resource", "uri", resources.URIFull)
	}

	// Register stats resource
	if shouldExpose(resources.URIStats) {
		s.server.AddResource(&sdk.Resource{
			URI:         resources.URIStats,
			Name:        "Capability Index Statistics",
			Description: "Statistical data about the capability index in JSON format including element counts, index statistics, and cache metrics",
			MIMEType:    "application/json",
		}, handler)
		logger.Info("Registered resource", "uri", resources.URIStats)
	}
}

// registerTools registers all NEXS MCP tools.
func (s *MCPServer) registerTools() {
	// Register list_elements tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "list_elements",
		Description: "List all elements with optional filtering",
	}, s.handleListElements)

	// Register get_element tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_element",
		Description: "Get a specific element by ID",
	}, s.handleGetElement)

	// Register create_element tool (generic)
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "create_element",
		Description: "Create a new element (generic - use type-specific tools for full features)",
	}, s.handleCreateElement)

	// Register type-specific create tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "create_persona",
		Description: "Create a new Persona element with behavioral traits, expertise areas, and response styles",
	}, s.handleCreatePersona)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "create_skill",
		Description: "Create a new Skill element with triggers, procedures, and dependencies",
	}, s.handleCreateSkill)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "create_template",
		Description: "Create a new Template element with variable substitution",
	}, s.handleCreateTemplate)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "create_agent",
		Description: "Create a new Agent element with goals, actions, and decision trees",
	}, s.handleCreateAgent)

	// Register quick create tools (simplified, minimal input, no preview needed)
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "quick_create_persona",
		Description: "QUICK: Create persona with minimal input using template defaults (no preview needed)",
	}, s.handleQuickCreatePersona)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "quick_create_skill",
		Description: "QUICK: Create skill with minimal input using template defaults (no preview needed)",
	}, s.handleQuickCreateSkill)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "quick_create_memory",
		Description: "QUICK: Create memory with minimal input (no preview needed)",
	}, s.handleQuickCreateMemory)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "quick_create_template",
		Description: "QUICK: Create template with minimal input (no preview needed)",
	}, s.handleQuickCreateTemplate)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "quick_create_agent",
		Description: "QUICK: Create agent with minimal input (no preview needed)",
	}, s.handleQuickCreateAgent)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "quick_create_ensemble",
		Description: "QUICK: Create ensemble with minimal input (no preview needed)",
	}, s.handleQuickCreateEnsemble)

	// Register batch creation tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "batch_create_elements",
		Description: "BATCH: Create multiple elements at once (single confirmation for all)",
	}, s.handleBatchCreateElements)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "create_memory",
		Description: "Create a new Memory element with automatic content hashing",
	}, s.handleCreateMemory)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "create_ensemble",
		Description: "Create a new Ensemble element for multi-agent orchestration",
	}, s.handleCreateEnsemble)

	// Register ensemble execution tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "execute_ensemble",
		Description: "Execute an ensemble with specified input and options. Orchestrates multiple agents according to ensemble configuration (sequential/parallel/hybrid modes).",
	}, s.handleExecuteEnsemble)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_ensemble_status",
		Description: "Get status and configuration of an ensemble including members, execution mode, and aggregation strategy",
	}, s.handleGetEnsembleStatus)

	// Register update_element tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "update_element",
		Description: "Update an existing element",
	}, s.handleUpdateElement)

	// Register delete_element tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "delete_element",
		Description: "Delete an element by ID",
	}, s.handleDeleteElement)

	// Register duplicate_element tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "duplicate_element",
		Description: "Duplicate an existing element with a new ID and optional new name",
	}, s.handleDuplicateElement)

	// Register search_elements tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "search_elements",
		Description: "Search elements with full-text search and advanced filtering (type, tags, author, date range)",
	}, s.handleSearchElements)

	// Register GitHub integration tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "github_auth_start",
		Description: "Start GitHub OAuth2 device flow authentication",
	}, s.handleGitHubAuthStart)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "github_auth_status",
		Description: "Check the status of GitHub authentication",
	}, s.handleGitHubAuthStatus)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "github_list_repos",
		Description: "List all repositories for the authenticated GitHub user",
	}, s.handleGitHubListRepos)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "github_sync_push",
		Description: "Push local elements to a GitHub repository",
	}, s.handleGitHubSyncPush)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "github_sync_pull",
		Description: "Pull elements from a GitHub repository to local storage",
	}, s.handleGitHubSyncPull)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "github_sync_bidirectional",
		Description: "Perform full bidirectional sync with GitHub repository (pull then push with conflict resolution)",
	}, s.handleGitHubSyncBidirectional)

	// Register backup/restore tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "backup_portfolio",
		Description: "Create a compressed backup of all portfolio elements with checksum validation",
	}, s.handleBackupPortfolio)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "restore_portfolio",
		Description: "Restore portfolio from a backup file with merge strategies and optional pre-restore backup",
	}, s.handleRestorePortfolio)

	// Register element activation tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "activate_element",
		Description: "Activate an element by ID (shortcut for updating is_active to true)",
	}, s.handleActivateElement)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "deactivate_element",
		Description: "Deactivate an element by ID (shortcut for updating is_active to false)",
	}, s.handleDeactivateElement)

	// Register memory management tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "search_memory",
		Description: "Search memories with relevance scoring and date filtering",
	}, s.handleSearchMemory)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "summarize_memories",
		Description: "Get a summary and statistics of memories with optional filtering",
	}, s.handleSummarizeMemories)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "update_memory",
		Description: "Update content, name, description, tags, or metadata of an existing memory",
	}, s.handleUpdateMemory)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "delete_memory",
		Description: "Delete a specific memory by ID",
	}, s.handleDeleteMemory)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "clear_memories",
		Description: "Clear multiple memories with optional author/date filtering (requires confirmation)",
	}, s.handleClearMemories)

	// Register context enrichment tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "expand_memory_context",
		Description: "Expand memory context by fetching related elements (personas, skills, agents, etc.). Supports type filtering, parallel/sequential fetch, and provides token savings estimation.",
	}, s.handleExpandMemoryContext)

	// Register bidirectional relationship search tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "find_related_memories",
		Description: "Find all memories that reference a specific element (reverse relationship search). Supports filtering by tags, author, date range, and sorting.",
	}, s.handleFindRelatedMemories)

	// Register recommendation tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "suggest_related_elements",
		Description: "Get intelligent recommendations for related elements based on relationships, co-occurrence patterns, and tag similarity. Returns scored suggestions with explanations.",
	}, s.handleSuggestRelatedElements)

	// Register auto-save tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "save_conversation_context",
		Description: "Save conversation context as a memory (auto-save feature). Automatically stores conversation history for continuity.",
	}, s.handleSaveConversationContext)

	// Register logging tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "list_logs",
		Description: "Query and filter structured logs with date range, level, and keyword filtering",
	}, s.handleListLogs)

	// Register analytics and performance tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_usage_stats",
		Description: "Get usage statistics and analytics for tool calls with period filtering",
	}, s.handleGetUsageStats)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_performance_dashboard",
		Description: "Get performance metrics dashboard with latency percentiles and slow operation alerts",
	}, s.handleGetPerformanceDashboard)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_cost_analytics",
		Description: "Get comprehensive cost analytics including trends, optimization opportunities, anomaly detection, and cost projections. Analyzes both performance metrics (duration, operations) and token metrics (usage, compression) to provide actionable recommendations for cost reduction.",
	}, s.handleGetCostAnalytics)

	// Register alert management tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_active_alerts",
		Description: "Get all currently active performance and cost alerts with severity levels, affected tools, and recommended actions",
	}, s.handleGetActiveAlerts)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_alert_history",
		Description: "Get historical alert data with optional filtering by severity, status, and time range",
	}, s.handleGetAlertHistory)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_alert_rules",
		Description: "Get all configured alert rules with thresholds, conditions, and cooldown periods",
	}, s.handleGetAlertRules)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "update_alert_rule",
		Description: "Create or update an alert rule with custom thresholds, metrics, and severity levels",
	}, s.handleUpdateAlertRule)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "resolve_alert",
		Description: "Mark an alert as resolved, removing it from the active alerts list",
	}, s.handleResolveAlert)

	// Register user identity tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_current_user",
		Description: "Get the current authenticated user and session context",
	}, s.handleGetCurrentUser)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "set_user_context",
		Description: "Set the current user context for the session with optional metadata",
	}, s.handleSetUserContext)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "clear_user_context",
		Description: "Clear the current user context (requires confirmation)",
	}, s.handleClearUserContext)

	// Register GitHub authentication tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "check_github_auth",
		Description: "Check GitHub authentication status and token validity",
	}, s.handleCheckGitHubAuth)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "refresh_github_token",
		Description: "Refresh GitHub OAuth token if expired or about to expire",
	}, s.handleRefreshGitHubToken)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "init_github_auth",
		Description: "Initialize GitHub device flow authentication",
	}, s.handleInitGitHubAuth)

	// Register index/search tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "search_capability_index",
		Description: "Search for capabilities using semantic search across all elements. Uses TF-IDF indexing to find relevant personas, skills, templates, agents, memories, and ensembles based on query text. Returns ranked results with relevance scores and text highlights.",
	}, s.handleSearchCapabilityIndex)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "find_similar_capabilities",
		Description: "Find capabilities similar to a given element. Uses semantic similarity to discover related personas, skills, templates, agents, memories, or ensembles. Useful for discovering complementary capabilities or alternatives.",
	}, s.handleFindSimilarCapabilities)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "map_capability_relationships",
		Description: "Map relationships between a capability and related elements. Analyzes semantic similarity to build a relationship graph showing complementary, similar, and related capabilities. Helps understand capability ecosystems.",
	}, s.handleMapCapabilityRelationships)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_capability_index_stats",
		Description: "Get statistics about the capability index. Shows total indexed documents, distribution by type, unique terms, and index health. Useful for monitoring and troubleshooting the semantic search system.",
	}, s.handleGetCapabilityIndexStats)

	// Register publishing tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "publish_collection",
		Description: "Publish a collection to NEXS-MCP registry via GitHub Pull Request. Validates manifest with 100+ rules, scans for security issues with 50+ patterns, creates tarball with checksums, forks registry repo, creates branch, commits files, and opens PR. Supports dry-run mode for testing.",
	}, s.handlePublishCollection)

	// Register enhanced discovery tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "search_collections",
		Description: "Advanced collection search with rich formatting, filtering (category, author, tags, min_stars), sorting (relevance, stars, downloads, updated, created, name), and pagination. Returns detailed results with element statistics, links, and optional emoji-rich display format.",
	}, s.handleSearchCollections)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "list_collections",
		Description: "List available collections with optional rich formatting, grouping (by category, author, source), and comprehensive summary statistics. Includes total elements, downloads, average stars, and breakdowns by category/author/source.",
	}, s.handleListCollections)

	// Register collection submission tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "submit_element_to_collection",
		Description: "Submit an element to a collection repository via GitHub Pull Request. Automatically forks the repo, creates a branch, commits the element, and opens a PR with generated description.",
	}, s.handleSubmitElementToCollection)

	// Register template tools
	s.registerTemplateTools()

	// Register validation tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "validate_element",
		Description: "Perform comprehensive type-specific validation on an element with configurable validation levels (basic, comprehensive, strict) and optional fix suggestions",
	}, s.handleValidateElement)

	// Register rendering tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "render_template",
		Description: "Render a template directly with provided data without creating an element. Supports both template_id (from repository) or direct template_content modes",
	}, s.handleRenderTemplate)

	// Register reload tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "reload_elements",
		Description: "Hot reload elements from disk without server restart. Supports selective reload by element type with optional cache clearing and validation",
	}, s.handleReloadElements)

	// Register GitHub portfolio search tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "search_portfolio_github",
		Description: "Search GitHub repositories for NEXS portfolios and elements. Requires GitHub authentication. Supports filtering by element type, author, tags, and sorting by stars/relevance/date",
	}, s.handleSearchPortfolioGitHub)

	// Register advanced relationship tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_related_elements",
		Description: "Get related elements bidirectionally (forward and reverse relationships). Supports direction filtering ('forward', 'reverse', 'both'), element type filtering, and active/inactive filtering",
	}, s.handleGetRelatedElements)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "expand_relationships",
		Description: "Perform multi-level recursive relationship expansion to discover deep connections. Supports max depth control (1-5), type filtering, cycle prevention, and bidirectional traversal",
	}, s.handleExpandRelationships)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "infer_relationships",
		Description: "Automatically infer relationships from content using multiple methods (mention detection, keyword matching, semantic similarity, pattern recognition). Returns confidence scores and evidence with optional auto-apply",
	}, s.handleInferRelationships)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_recommendations",
		Description: "Get intelligent element recommendations based on relationships, co-occurrence patterns, and similarity. Returns scored recommendations with optional reasoning explanations",
	}, s.handleGetRecommendations)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_relationship_stats",
		Description: "Get relationship index statistics including forward/reverse entry counts, cache hit rates, and optional element-specific relationship counts",
	}, s.handleGetRelationshipStats)

	// Register temporal/versioning tools (Sprint 11)
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_element_history",
		Description: "Retrieve complete version history of an element with timestamps, authors, change types, and diffs. Supports optional time range filtering (RFC3339 format)",
	}, s.handleGetElementHistory)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_relation_history",
		Description: "Retrieve complete version history of a relationship with confidence tracking. Supports optional time range filtering and confidence decay calculation",
	}, s.handleGetRelationHistory)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_graph_at_time",
		Description: "Reconstruct the entire graph state (elements + relationships) at a specific point in time. Supports optional confidence decay application. Time travel query for historical analysis",
	}, s.handleGetGraphAtTime)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_decayed_graph",
		Description: "Get current graph with time-based confidence decay applied to all relationships. Filters relationships below threshold. Useful for finding stale connections",
	}, s.handleGetDecayedGraph)

	// Register working memory tools (Two-Tier Memory Architecture)
	RegisterWorkingMemoryTools(s, s.workingMemory)

	// Register quality and retention tools (Sprint 8)
	s.RegisterQualityTools()

	// Register token optimization tools
	s.registerOptimizationTools()

	// Register memory consolidation tools (Sprint 14)
	s.RegisterConsolidationTools()

	// Register metrics dashboard tools
	s.RegisterMetricsDashboardTools()

	// Register NLP tools (Sprint 18)
	s.RegisterNLPTools()

	// Register skill extraction tools
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "extract_skills_from_persona",
		Description: "Extract skills from a persona's expertise areas and custom fields, creating them as separate skill elements and linking them to the persona",
	}, s.handleExtractSkillsFromPersona)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "batch_extract_skills",
		Description: "Extract skills from multiple personas in batch. If no persona IDs provided, processes all personas in the system",
	}, s.handleBatchExtractSkills)
}

// registerOptimizationTools registers token optimization tools.
func (s *MCPServer) registerOptimizationTools() {
	// Semantic Deduplication
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "deduplicate_memories",
		Description: "Find and merge duplicate memories using semantic similarity (92%+ threshold). Supports multiple merge strategies: keep_first, keep_last, keep_longest, combine. Returns groups of duplicates with similarity scores and merged results.",
	}, s.handleDeduplicateMemories)

	// Context Window Optimization
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "optimize_context",
		Description: "Optimize context window to prevent overflow and maximize relevance. Supports 4 priority strategies (recency, relevance, hybrid, importance) and 3 truncation methods (head, tail, middle). Returns optimized items with compression metrics.",
	}, s.handleOptimizeContext)

	// Compression Statistics
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_optimization_stats",
		Description: "Get comprehensive statistics for all token optimization services: compression ratios, streaming performance, summarization savings, deduplication metrics, context window optimizations, and prompt compression rates.",
	}, s.handleGetOptimizationStats)
}

// rebuildIndex populates the TF-IDF index with all elements from the repository.
func (s *MCPServer) rebuildIndex() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// List all elements
	elements, err := s.repo.List(domain.ElementFilter{})
	if err != nil {
		return // Silently skip if repository not ready
	}

	// Index each element
	for _, elem := range elements {
		s.indexElement(elem)
	}
}

// indexElement adds or updates an element in the index.
func (s *MCPServer) indexElement(elem domain.Element) {
	metadata := elem.GetMetadata()

	// Build content string from element metadata and type-specific fields
	content := metadata.Name + " " + metadata.Description

	// Add type-specific content
	switch e := elem.(type) {
	case *domain.Persona:
		var contentSb497 strings.Builder
		for _, trait := range e.BehavioralTraits {
			contentSb497.WriteString(" " + trait.Name + " " + trait.Description)
		}
		content += contentSb497.String()
		var contentSb500 strings.Builder
		var contentSb504 strings.Builder
		for _, area := range e.ExpertiseAreas {
			contentSb500.WriteString(" " + area.Domain + " " + area.Description)
			var contentSb502 strings.Builder
			for _, keyword := range area.Keywords {
				contentSb502.WriteString(" " + keyword)
			}
			contentSb504.WriteString(contentSb502.String())
		}
		content += contentSb504.String()
		content += contentSb500.String()
		content += " " + e.SystemPrompt
	case *domain.Skill:
		var contentSb508 strings.Builder
		var contentSb516 strings.Builder
		for _, trigger := range e.Triggers {
			contentSb508.WriteString(" " + trigger.Pattern + " " + trigger.Context)
			var contentSb510 strings.Builder
			for _, keyword := range trigger.Keywords {
				contentSb510.WriteString(" " + keyword)
			}
			contentSb516.WriteString(contentSb510.String())
		}
		content += contentSb516.String()
		content += contentSb508.String()
		var contentSb514 strings.Builder
		for _, proc := range e.Procedures {
			contentSb514.WriteString(" " + proc.Action + " " + proc.Description)
		}
		content += contentSb514.String()
	case *domain.Template:
		content += " " + e.Content
	}

	// Add tags
	var contentSb522 strings.Builder
	for _, tag := range metadata.Tags {
		contentSb522.WriteString(" " + tag)
	}
	content += contentSb522.String()

	// Index using HNSW-backed hybrid search
	metadataMap := map[string]interface{}{
		"type": string(metadata.Type),
		"name": metadata.Name,
	}
	ctx := context.Background()
	_ = s.hybridSearch.Add(ctx, metadata.ID, content, metadataMap)
}

// removeFromIndex removes an element from the index.
//
//nolint:unused // Reserved for future use
func (s *MCPServer) removeFromIndex(elementID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from HNSW-backed hybrid search
	ctx := context.Background()
	_ = s.hybridSearch.Delete(ctx, elementID)
}

// createSearchableText creates searchable text from an element for semantic search.
func (s *MCPServer) createSearchableText(elem domain.Element) string {
	metadata := elem.GetMetadata()
	content := metadata.Name + " " + metadata.Description

	// Add type-specific content
	switch e := elem.(type) {
	case *domain.Persona:
		var contentSb686 strings.Builder
		for _, trait := range e.BehavioralTraits {
			contentSb686.WriteString(" " + trait.Name + " " + trait.Description)
		}
		content += contentSb686.String()
		var contentSb689 strings.Builder
		var contentSb692 strings.Builder
		for _, area := range e.ExpertiseAreas {
			contentSb689.WriteString(" " + area.Domain + " " + area.Description)
			var contentSb691 strings.Builder
			for _, keyword := range area.Keywords {
				contentSb691.WriteString(" " + keyword)
			}
			contentSb692.WriteString(contentSb691.String())
		}
		content += contentSb692.String()
		content += contentSb689.String()
		content += " " + e.SystemPrompt
	case *domain.Skill:
		var contentSb697 strings.Builder
		var contentSb704 strings.Builder
		for _, trigger := range e.Triggers {
			contentSb697.WriteString(" " + trigger.Pattern + " " + trigger.Context)
			var contentSb699 strings.Builder
			for _, keyword := range trigger.Keywords {
				contentSb699.WriteString(" " + keyword)
			}
			contentSb704.WriteString(contentSb699.String())
		}
		content += contentSb704.String()
		content += contentSb697.String()
		var contentSb703 strings.Builder
		for _, proc := range e.Procedures {
			contentSb703.WriteString(" " + proc.Action + " " + proc.Description)
		}
		content += contentSb703.String()
	case *domain.Template:
		content += " " + e.Content
	case *domain.Memory:
		content += " " + e.Content
	case *domain.Agent:
		if len(e.Goals) > 0 {
			content += " " + e.Goals[0] // Use first goal
		}
	}

	// Add tags
	var contentSb717 strings.Builder
	for _, tag := range metadata.Tags {
		contentSb717.WriteString(" " + tag)
	}
	content += contentSb717.String()

	return content
}

// Run starts the MCP server with stdio transport.
func (s *MCPServer) Run(ctx context.Context) error {
	// Start auto-save worker if enabled
	if s.cfg.AutoSaveMemories {
		s.startAutoSaveWorker(ctx)
	}

	// Create context with cancel for cleanup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Cleanup auto-save on exit
	defer s.stopAutoSaveWorker()

	transport := &sdk.StdioTransport{}
	return s.server.Run(ctx, transport)
}

// startAutoSaveWorker starts the background auto-save worker.
func (s *MCPServer) startAutoSaveWorker(ctx context.Context) {
	interval := s.cfg.AutoSaveInterval
	if interval <= 0 {
		interval = 5 * time.Minute // Default to 5 minutes
	}

	s.autoSaveTicker = time.NewTicker(interval)
	logger.Info("Auto-save worker started", "interval", interval.String())

	go func() {
		for {
			select {
			case <-s.autoSaveTicker.C:
				s.performAutoSave(ctx)
			case <-s.autoSaveStopChan:
				logger.Info("Auto-save worker stopped")
				return
			case <-ctx.Done():
				logger.Info("Auto-save worker stopped due to context cancellation")
				return
			}
		}
	}()
}

// stopAutoSaveWorker stops the auto-save worker.
func (s *MCPServer) stopAutoSaveWorker() {
	if s.autoSaveTicker != nil {
		s.autoSaveTicker.Stop()
	}
	close(s.autoSaveStopChan)
}

// performAutoSave executes the auto-save logic.
func (s *MCPServer) performAutoSave(ctx context.Context) {
	// Check if enough time has passed since last save
	if time.Since(s.lastAutoSave) < s.cfg.AutoSaveInterval {
		return
	}

	logger.Debug("Auto-save triggered", "interval", s.cfg.AutoSaveInterval.String())

	// Get current user context
	currentUser, _ := globalUserSession.GetUser()
	if currentUser == "" {
		logger.Debug("Auto-save skipped: No user context set. Use 'set_user_context' tool to define user.")
		return
	}

	// Get recent working memories for the current session
	sessionID := "auto-save-" + currentUser
	memories, err := s.workingMemory.List(ctx, sessionID, false, false)
	if err != nil {
		logger.Error("Failed to list working memories", "error", err)
		return
	}

	if len(memories) == 0 {
		logger.Debug("Auto-save skipped: No working memories to save",
			"user", currentUser,
			"session_id", sessionID,
			"tip", "Working memories are created when you use MCP tools or call 'working_memory_add'")
		return
	}

	logger.Info("Performing auto-save of conversation context",
		"user", currentUser,
		"memory_count", len(memories))

	// Build conversation context from working memories
	var contextParts []string
	for _, mem := range memories {
		contextParts = append(contextParts, mem.Content)
	}
	conversationContext := strings.Join(contextParts, "\n\n")

	// Step 1: Apply semantic deduplication to remove duplicate memories
	if s.deduplicationService != nil && len(memories) > 1 {
		logger.Debug("Applying semantic deduplication", "memory_count", len(memories))

		// Convert memories to deduplication items
		dedupeItems := make([]application.DeduplicateItem, len(memories))
		for i, mem := range memories {
			dedupeItems[i] = application.DeduplicateItem{
				ID:      mem.ID,
				Content: mem.Content,
			}
		}

		// Deduplicate (returns 3 values: items, result, error)
		deduplicated, result, err := s.deduplicationService.DeduplicateItems(ctx, dedupeItems)
		if err != nil {
			logger.Warn("Deduplication failed, continuing with original memories", "error", err)
		} else {
			originalCount := len(memories)
			deduplicatedCount := len(deduplicated)

			if deduplicatedCount < originalCount {
				logger.Info("Deduplication reduced memory count",
					"original", originalCount,
					"deduplicated", deduplicatedCount,
					"removed", result.DuplicatesRemoved,
					"bytes_saved", result.BytesSaved)

				// Rebuild context from deduplicated items
				contextParts = make([]string, len(deduplicated))
				for i, item := range deduplicated {
					contextParts[i] = item.Content
				}
				conversationContext = strings.Join(contextParts, "\n\n")

				// Record deduplication metrics
				originalTokens := application.EstimateTokenCount(strings.Repeat("x", result.BytesSaved+len(conversationContext)))
				optimizedTokens := application.EstimateTokenCount(conversationContext)
				s.tokenMetrics.RecordTokenOptimization(application.TokenMetrics{
					OriginalTokens:   originalTokens,
					OptimizedTokens:  optimizedTokens,
					OptimizationType: "deduplication",
					ToolName:         "auto_save_worker",
					Timestamp:        time.Now(),
				})
			}
		}
	}

	// Step 2: Apply context window optimization if context is too large
	if s.contextWindowManager != nil {
		contextTokens := application.EstimateTokenCount(conversationContext)
		maxTokens := int64(s.cfg.Streaming.BufferSize * 100) // Use buffer size as proxy for max context

		if contextTokens > maxTokens {
			logger.Debug("Context too large, applying context window optimization",
				"tokens", contextTokens,
				"max_tokens", maxTokens)

			// Convert context to items for optimization
			items := make([]application.ContextItem, len(contextParts))
			for i, part := range contextParts {
				tokenCount := int(application.EstimateTokenCount(part))
				items[i] = application.ContextItem{
					ID:         fmt.Sprintf("mem_%d", i),
					Content:    part,
					TokenCount: tokenCount,
					CreatedAt:  time.Now().Add(-time.Duration(len(contextParts)-i) * time.Minute),
					Relevance:  1.0 - (float64(len(contextParts)-i) / float64(len(contextParts))),
					Importance: 5,
				}
			}

			// Optimize context (returns 3 values: items, result, error)
			optimized, result, err := s.contextWindowManager.OptimizeContext(ctx, items)
			if err != nil {
				logger.Warn("Context window optimization failed", "error", err)
			} else {
				originalTokens := contextTokens

				// Rebuild context from optimized items
				optimizedParts := make([]string, len(optimized))
				for i, item := range optimized {
					optimizedParts[i] = item.Content
				}
				conversationContext = strings.Join(optimizedParts, "\n\n")
				optimizedTokens := application.EstimateTokenCount(conversationContext)

				logger.Info("Context window optimized",
					"original_tokens", originalTokens,
					"optimized_tokens", optimizedTokens,
					"saved_tokens", originalTokens-optimizedTokens,
					"items_removed", result.ItemsRemoved)

				// Record optimization metrics
				s.tokenMetrics.RecordTokenOptimization(application.TokenMetrics{
					OriginalTokens:   originalTokens,
					OptimizedTokens:  optimizedTokens,
					OptimizationType: "context_window",
					ToolName:         "auto_save_worker",
					Timestamp:        time.Now(),
				})
			}
		}
	}

	// Step 3: Apply summarization for old memories (> 7 days)
	if s.summarizationService != nil && len(conversationContext) > 1000 {
		age := time.Since(time.Now().AddDate(0, 0, -7)) // Simulate 7 days old
		contentLength := len(conversationContext)

		if s.summarizationService.ShouldSummarize(contentLength, age) {
			logger.Debug("Applying TF-IDF summarization", "content_length", contentLength)

			summarized, metadata, err := s.summarizationService.SummarizeText(ctx, conversationContext, time.Now())
			if err != nil {
				logger.Warn("Summarization failed", "error", err)
			} else {
				originalTokens := application.EstimateTokenCount(conversationContext)
				summarizedTokens := application.EstimateTokenCount(summarized)

				logger.Info("Content summarized",
					"original_tokens", originalTokens,
					"summarized_tokens", summarizedTokens,
					"compression_ratio", metadata.CompressionRatio)

				conversationContext = summarized

				// Record summarization metrics
				s.tokenMetrics.RecordTokenOptimization(application.TokenMetrics{
					OriginalTokens:   originalTokens,
					OptimizedTokens:  summarizedTokens,
					CompressionRatio: metadata.CompressionRatio,
					OptimizationType: "summarization",
					ToolName:         "auto_save_worker",
					Timestamp:        time.Now(),
				})
			}
		}
	}

	// Step 4: Compress the context if enabled
	if s.cfg.PromptCompression.Enabled && len(conversationContext) > s.cfg.PromptCompression.MinPromptLength {
		originalSize := len(conversationContext)
		compressed, _, err := s.promptCompressor.CompressPrompt(ctx, conversationContext)
		if err != nil {
			logger.Error("Failed to compress conversation context", "error", err)
		} else {
			compressedSize := len(compressed)
			conversationContext = compressed

			// Record token savings
			s.tokenMetrics.RecordTokenOptimization(application.TokenMetrics{
				OriginalTokens:   application.EstimateTokenCount(conversationContext[:originalSize]),
				OptimizedTokens:  application.EstimateTokenCount(conversationContext),
				OptimizationType: "auto_save_compression",
				ToolName:         "auto_save_worker",
				Timestamp:        time.Now(),
			})

			logger.Info("Compressed conversation context",
				"original_size", originalSize,
				"compressed_size", compressedSize,
				"ratio", float64(compressedSize)/float64(originalSize))
		}
	}

	// Create memory element using NewMemory factory
	now := time.Now()
	memory := domain.NewMemory(
		"Auto-saved conversation context - "+now.Format("2006-01-02 15:04:05"),
		"Automatically saved conversation context",
		"1",
		currentUser,
	)
	memory.Content = conversationContext
	memory.DateCreated = now.Format("2006-01-02")
	memory.Metadata = map[string]string{
		"source":       "auto_save_worker",
		"session_id":   sessionID,
		"memory_count": strconv.Itoa(len(memories)),
		"user":         currentUser,
		"timestamp":    now.Format(time.RFC3339),
	}

	// Save to repository - cast Memory to Element interface
	if err := s.repo.Create(memory); err != nil {
		logger.Error("Failed to auto-save conversation context", "error", err)
		return
	}

	logger.Info("Successfully auto-saved conversation context",
		"memory_id", memory.GetMetadata().ID,
		"memory_count", len(memories),
		"user", currentUser)

	// Update last auto-save time
	s.lastAutoSave = time.Now()

	// Clear working memories after successful save
	if err := s.workingMemory.ClearSession(sessionID); err != nil {
		logger.Warn("Failed to clear working memories after auto-save", "error", err)
	}
}
