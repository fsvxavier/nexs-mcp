package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// MCPServer wraps the official MCP SDK server
type MCPServer struct {
	server *sdk.Server
	repo   domain.ElementRepository
}

// NewMCPServer creates a new MCP server using the official SDK
func NewMCPServer(name, version string, repo domain.ElementRepository) *MCPServer {
	impl := &sdk.Implementation{
		Name:    name,
		Version: version,
	}

	// Create server with default capabilities
	server := sdk.NewServer(impl, nil)

	mcpServer := &MCPServer{
		server: server,
		repo:   repo,
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

	// Register search_elements tool
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "search_elements",
		Description: "Search elements with full-text search and advanced filtering (type, tags, author, date range)",
	}, s.handleSearchElements)
}

// Run starts the MCP server with stdio transport
func (s *MCPServer) Run(ctx context.Context) error {
	transport := &sdk.StdioTransport{}
	return s.server.Run(ctx, transport)
}
