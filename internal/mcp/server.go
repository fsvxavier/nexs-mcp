package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// Server represents the MCP server instance
type Server struct {
	name         string
	version      string
	tools        map[string]*Tool
	capabilities ServerCapabilities
	stdin        io.Reader
	stdout       io.Writer
}

// ServerCapabilities defines what the server can do
type ServerCapabilities struct {
	Tools     bool `json:"tools"`
	Resources bool `json:"resources"`
	Prompts   bool `json:"prompts"`
}

// NewServer creates a new MCP server instance
func NewServer(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
		tools:   make(map[string]*Tool),
		capabilities: ServerCapabilities{
			Tools:     true,
			Resources: false,
			Prompts:   false,
		},
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
}

// RegisterTool registers a new tool with the server
func (s *Server) RegisterTool(tool *Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if _, exists := s.tools[tool.Name]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name)
	}
	s.tools[tool.Name] = tool
	return nil
}

// GetTool retrieves a tool by name
func (s *Server) GetTool(name string) (*Tool, error) {
	tool, exists := s.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return tool, nil
}

// ListTools returns all registered tools
func (s *Server) ListTools() []*Tool {
	tools := make([]*Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Start starts the MCP server with stdio transport
func (s *Server) Start(ctx context.Context) error {
	decoder := json.NewDecoder(s.stdin)
	encoder := json.NewEncoder(s.stdout)

	// Channel for processing messages
	msgChan := make(chan *JSONRPCRequest, 10)
	errChan := make(chan error, 1)

	// Start message reader goroutine
	go func() {
		for {
			var req JSONRPCRequest
			if err := decoder.Decode(&req); err != nil {
				if err == io.EOF {
					close(msgChan)
					return
				}
				errChan <- fmt.Errorf("failed to decode request: %w", err)
				return
			}
			msgChan <- &req
		}
	}()

	// Process messages
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			return err
		case req, ok := <-msgChan:
			if !ok {
				return nil
			}

			response := s.handleRequest(req)
			if err := encoder.Encode(response); err != nil {
				return fmt.Errorf("failed to encode response: %w", err)
			}
		}
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize() (map[string]interface{}, error) {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    s.capabilities,
		"serverInfo": map[string]string{
			"name":    s.name,
			"version": s.version,
		},
	}, nil
}

// handleListTools handles the tools/list request
func (s *Server) handleListTools() (map[string]interface{}, error) {
	tools := make([]map[string]interface{}, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}
	return map[string]interface{}{
		"tools": tools,
	}, nil
}

// handleCallTool handles the tools/call request
func (s *Server) handleCallTool(name string, args json.RawMessage) (interface{}, error) {
	tool, err := s.GetTool(name)
	if err != nil {
		return nil, err
	}
	return tool.Execute(args)
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     ToolHandler
}

// ToolHandler is the function signature for tool execution
type ToolHandler func(args json.RawMessage) (interface{}, error)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// handleRequest processes a JSON-RPC request
func (s *Server) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		result, err := s.handleInitialize()
		if err != nil {
			response.Error = &RPCError{Code: -32603, Message: err.Error()}
		} else {
			response.Result = result
		}

	case "tools/list":
		result, err := s.handleListTools()
		if err != nil {
			response.Error = &RPCError{Code: -32603, Message: err.Error()}
		} else {
			response.Result = result
		}

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			response.Error = &RPCError{Code: -32602, Message: "Invalid params"}
		} else {
			result, err := s.handleCallTool(params.Name, params.Arguments)
			if err != nil {
				response.Error = &RPCError{Code: -32603, Message: err.Error()}
			} else {
				response.Result = map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": fmt.Sprintf("%v", result),
						},
					},
				}
			}
		}

	default:
		response.Error = &RPCError{Code: -32601, Message: "Method not found"}
	}

	return response
}

// Execute executes the tool with the given arguments
func (t *Tool) Execute(args json.RawMessage) (interface{}, error) {
	if t.Handler == nil {
		return nil, fmt.Errorf("tool %s has no handler", t.Name)
	}
	return t.Handler(args)
}

// NewListElementsTool creates the list_elements tool
func NewListElementsTool(repo domain.ElementRepository) *Tool {
	return &Tool{
		Name:        "list_elements",
		Description: "List all elements with optional filtering by type, status, and tags",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Filter by element type",
					"enum":        []string{"persona", "skill", "template", "agent", "memory", "ensemble"},
				},
				"is_active": map[string]interface{}{
					"type":        "boolean",
					"description": "Filter by active status",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "Filter by tags",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results",
					"minimum":     1,
					"maximum":     100,
					"default":     10,
				},
				"offset": map[string]interface{}{
					"type":        "integer",
					"description": "Number of results to skip",
					"minimum":     0,
					"default":     0,
				},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var filter domain.ElementFilter
			if len(args) > 0 {
				if err := json.Unmarshal(args, &filter); err != nil {
					return nil, fmt.Errorf("invalid arguments: %w", err)
				}
			}

			elements, err := repo.List(filter)
			if err != nil {
				return nil, fmt.Errorf("failed to list elements: %w", err)
			}

			return map[string]interface{}{
				"elements": elements,
				"count":    len(elements),
			}, nil
		},
	}
}
