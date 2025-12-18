package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_handleRequest(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	repo := NewMockElementRepository()

	// Register a test tool
	tool := NewListElementsTool(repo)
	server.RegisterTool(tool)

	t.Run("initialize request", func(t *testing.T) {
		req := &JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "initialize",
			ID:      1,
		}

		response := server.handleRequest(req)

		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, 1, response.ID)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)

		result := response.Result.(map[string]interface{})
		assert.Equal(t, "2024-11-05", result["protocolVersion"])
		assert.NotNil(t, result["capabilities"])
		assert.NotNil(t, result["serverInfo"])
	})

	t.Run("tools/list request", func(t *testing.T) {
		req := &JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "tools/list",
			ID:      2,
		}

		response := server.handleRequest(req)

		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)

		result := response.Result.(map[string]interface{})
		tools := result["tools"].([]map[string]interface{})
		assert.Len(t, tools, 1)
		assert.Equal(t, "list_elements", tools[0]["name"])
	})

	t.Run("tools/call request", func(t *testing.T) {
		// Create test element
		element := createTestElement("persona", "Test Persona")
		repo.Create(element)

		params := map[string]interface{}{
			"name":      "list_elements",
			"arguments": map[string]interface{}{},
		}
		paramsJSON, _ := json.Marshal(params)

		req := &JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "tools/call",
			Params:  paramsJSON,
			ID:      3,
		}

		response := server.handleRequest(req)

		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)
	})

	t.Run("invalid method", func(t *testing.T) {
		req := &JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "invalid/method",
			ID:      4,
		}

		response := server.handleRequest(req)

		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, -32601, response.Error.Code)
		assert.Equal(t, "Method not found", response.Error.Message)
	})

	t.Run("tools/call with invalid params", func(t *testing.T) {
		req := &JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "tools/call",
			Params:  json.RawMessage(`{invalid}`),
			ID:      5,
		}

		response := server.handleRequest(req)

		assert.NotNil(t, response.Error)
		assert.Equal(t, -32602, response.Error.Code)
	})

	t.Run("tools/call with non-existent tool", func(t *testing.T) {
		params := map[string]interface{}{
			"name":      "nonexistent_tool",
			"arguments": map[string]interface{}{},
		}
		paramsJSON, _ := json.Marshal(params)

		req := &JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "tools/call",
			Params:  paramsJSON,
			ID:      6,
		}

		response := server.handleRequest(req)

		assert.NotNil(t, response.Error)
		assert.Equal(t, -32603, response.Error.Code)
	})
}

func TestServer_Start(t *testing.T) {
	t.Run("graceful shutdown", func(t *testing.T) {
		server := NewServer("test-server", "1.0.0")

		// Use strings reader/writer for testing
		input := strings.NewReader("")
		server.stdin = input
		server.stdout = &strings.Builder{}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := server.Start(ctx)
		assert.NoError(t, err)
	})

	t.Run("process initialize request", func(t *testing.T) {
		server := NewServer("test-server", "1.0.0")

		input := strings.NewReader(`{"jsonrpc":"2.0","method":"initialize","id":1}` + "\n")
		output := &strings.Builder{}
		server.stdin = input
		server.stdout = output

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := server.Start(ctx)
		assert.NoError(t, err)

		// Verify response was written
		response := output.String()
		assert.Contains(t, response, "2.0")
		assert.Contains(t, response, "2024-11-05")
	})
}

func TestJSONRPCTypes(t *testing.T) {
	t.Run("JSONRPCRequest marshaling", func(t *testing.T) {
		req := &JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "test",
			Params:  json.RawMessage(`{"key":"value"}`),
			ID:      1,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)
		assert.Contains(t, string(data), "test")
	})

	t.Run("JSONRPCResponse marshaling", func(t *testing.T) {
		resp := &JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  map[string]string{"status": "ok"},
			ID:      1,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)
		assert.Contains(t, string(data), "ok")
	})

	t.Run("JSONRPCResponse with error", func(t *testing.T) {
		resp := &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32600,
				Message: "Invalid Request",
			},
			ID: 1,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)
		assert.Contains(t, string(data), "Invalid Request")
		assert.Contains(t, string(data), "-32600")
	})
}
