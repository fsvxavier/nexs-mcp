package mcp

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

// MCPServer wraps the official MCP SDK server
type MCPServer struct {
	server      *sdk.Server
	repo        domain.ElementRepository
	metrics     *application.MetricsCollector
	perfMetrics *logger.PerformanceMetrics
	mu          sync.Mutex
	deviceCodes map[string]string // Maps user codes to device codes for GitHub OAuth
}

// NewMCPServer creates a new MCP server using the official SDK
func NewMCPServer(name, version string, repo domain.ElementRepository) *MCPServer {
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

	mcpServer := &MCPServer{
		server:      server,
		repo:        repo,
		metrics:     metrics,
		perfMetrics: perfMetrics,
	}

	// Register all tools
	mcpServer.registerTools()

	return mcpServer
}

// registerTools registers all NEXS MCP tools
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

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "create_memory",
		Description: "Create a new Memory element with automatic content hashing",
	}, s.handleCreateMemory)

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "create_ensemble",
		Description: "Create a new Ensemble element for multi-agent orchestration",
	}, s.handleCreateEnsemble)

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
}

// Run starts the MCP server with stdio transport
func (s *MCPServer) Run(ctx context.Context) error {
	transport := &sdk.StdioTransport{}
	return s.server.Run(ctx, transport)
}
