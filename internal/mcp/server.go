package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/config"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/indexing"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
	"github.com/fsvxavier/nexs-mcp/internal/mcp/resources"
)

// MCPServer wraps the official MCP SDK server.
type MCPServer struct {
	server             *sdk.Server
	repo               domain.ElementRepository
	metrics            *application.MetricsCollector
	perfMetrics        *logger.PerformanceMetrics
	index              *indexing.TFIDFIndex
	mu                 sync.Mutex
	deviceCodes        map[string]string // Maps user codes to device codes for GitHub OAuth
	capabilityResource *resources.CapabilityIndexResource
	resourcesConfig    config.ResourcesConfig
	cfg                *config.Config // Store config for auto-save checks
}

// NewMCPServer creates a new MCP server using the official SDK.
func NewMCPServer(name, version string, repo domain.ElementRepository, cfg *config.Config) *MCPServer {
	impl := &sdk.Implementation{
		Name:    name,
		Version: version,
	}

	// Create server with default capabilities
	server := sdk.NewServer(impl, nil)

	// Create metrics collector (store in ~/.nexs-mcp/metrics)
	metricsDir := filepath.Join(os.Getenv("HOME"), ".nexs-mcp", "metrics")
	metrics := application.NewMetricsCollector(metricsDir)

	// Create performance metrics (store in ~/.nexs-mcp/performance)
	perfDir := filepath.Join(os.Getenv("HOME"), ".nexs-mcp", "performance")
	perfMetrics := logger.NewPerformanceMetrics(perfDir)

	// Create TF-IDF index
	idx := indexing.NewTFIDFIndex()

	// Create capability index resource
	capabilityResource := resources.NewCapabilityIndexResource(repo, idx, cfg.Resources.CacheTTL)

	mcpServer := &MCPServer{
		server:             server,
		repo:               repo,
		metrics:            metrics,
		perfMetrics:        perfMetrics,
		index:              idx,
		capabilityResource: capabilityResource,
		resourcesConfig:    cfg.Resources,
		cfg:                cfg, // Store config for auto-save checks
	}

	// Populate index with existing elements
	mcpServer.rebuildIndex()

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

	doc := &indexing.Document{
		ID:      metadata.ID,
		Type:    metadata.Type,
		Name:    metadata.Name,
		Content: content,
	}

	s.index.AddDocument(doc)
}

// removeFromIndex removes an element from the index.
//
//nolint:unused // Reserved for future use
func (s *MCPServer) removeFromIndex(elementID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index.RemoveDocument(elementID)
}

// Run starts the MCP server with stdio transport.
func (s *MCPServer) Run(ctx context.Context) error {
	transport := &sdk.StdioTransport{}
	return s.server.Run(ctx, transport)
}
