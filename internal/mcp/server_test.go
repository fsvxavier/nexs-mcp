package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	server := NewServer("test-server", "1.0.0")

	require.NotNil(t, server)
	assert.Equal(t, "test-server", server.name)
	assert.Equal(t, "1.0.0", server.version)
	assert.NotNil(t, server.tools)
	assert.Empty(t, server.tools)
	assert.True(t, server.capabilities.Tools)
}

func TestServer_RegisterTool(t *testing.T) {
	t.Run("register valid tool", func(t *testing.T) {
		server := NewServer("test", "1.0.0")
		tool := &Tool{
			Name:        "test_tool",
			Description: "Test tool",
			Handler:     func(args json.RawMessage) (interface{}, error) { return nil, nil },
		}

		err := server.RegisterTool(tool)
		require.NoError(t, err)
		assert.Len(t, server.tools, 1)
	})

	t.Run("register nil tool", func(t *testing.T) {
		server := NewServer("test", "1.0.0")
		err := server.RegisterTool(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("register tool with empty name", func(t *testing.T) {
		server := NewServer("test", "1.0.0")
		tool := &Tool{Description: "Test"}
		err := server.RegisterTool(tool)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("register duplicate tool", func(t *testing.T) {
		server := NewServer("test", "1.0.0")
		tool := &Tool{
			Name:    "duplicate",
			Handler: func(args json.RawMessage) (interface{}, error) { return nil, nil },
		}

		require.NoError(t, server.RegisterTool(tool))
		err := server.RegisterTool(tool)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestServer_GetTool(t *testing.T) {
	server := NewServer("test", "1.0.0")
	tool := &Tool{
		Name:    "existing",
		Handler: func(args json.RawMessage) (interface{}, error) { return nil, nil },
	}
	require.NoError(t, server.RegisterTool(tool))

	t.Run("get existing tool", func(t *testing.T) {
		retrieved, err := server.GetTool("existing")
		require.NoError(t, err)
		assert.Equal(t, "existing", retrieved.Name)
	})

	t.Run("get non-existing tool", func(t *testing.T) {
		_, err := server.GetTool("non_existing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestServer_ListTools(t *testing.T) {
	server := NewServer("test", "1.0.0")

	t.Run("empty tool list", func(t *testing.T) {
		tools := server.ListTools()
		assert.Empty(t, tools)
	})

	t.Run("list multiple tools", func(t *testing.T) {
		tool1 := &Tool{Name: "tool1", Handler: func(args json.RawMessage) (interface{}, error) { return nil, nil }}
		tool2 := &Tool{Name: "tool2", Handler: func(args json.RawMessage) (interface{}, error) { return nil, nil }}

		require.NoError(t, server.RegisterTool(tool1))
		require.NoError(t, server.RegisterTool(tool2))

		tools := server.ListTools()
		assert.Len(t, tools, 2)
	})
}

func TestServer_HandleInitialize(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	result, err := server.handleInitialize()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "2024-11-05", result["protocolVersion"])

	serverInfo := result["serverInfo"].(map[string]string)
	assert.Equal(t, "test-server", serverInfo["name"])
	assert.Equal(t, "1.0.0", serverInfo["version"])
}

func TestServer_HandleListTools(t *testing.T) {
	server := NewServer("test", "1.0.0")
	tool := &Tool{
		Name:        "test_tool",
		Description: "Test description",
		InputSchema: map[string]interface{}{"type": "object"},
		Handler:     func(args json.RawMessage) (interface{}, error) { return nil, nil },
	}
	require.NoError(t, server.RegisterTool(tool))

	result, err := server.handleListTools()
	require.NoError(t, err)

	tools := result["tools"].([]map[string]interface{})
	assert.Len(t, tools, 1)
	assert.Equal(t, "test_tool", tools[0]["name"])
}

func TestServer_StartShutdown(t *testing.T) {
	server := NewServer("test", "1.0.0")
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- server.Start(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestTool_Execute(t *testing.T) {
	t.Run("execute with handler", func(t *testing.T) {
		tool := &Tool{
			Name: "test",
			Handler: func(args json.RawMessage) (interface{}, error) {
				return map[string]string{"status": "ok"}, nil
			},
		}

		result, err := tool.Execute(json.RawMessage(`{}`))
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("execute without handler", func(t *testing.T) {
		tool := &Tool{Name: "no_handler"}
		_, err := tool.Execute(json.RawMessage(`{}`))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has no handler")
	})
}

func TestNewListElementsTool(t *testing.T) {
	repo := NewMockElementRepository()
	tool := NewListElementsTool(repo)

	require.NotNil(t, tool)
	assert.Equal(t, "list_elements", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.NotNil(t, tool.InputSchema)
	assert.NotNil(t, tool.Handler)
}

func TestServer_HandleCallTool(t *testing.T) {
	server := NewServer("test", "1.0.0")
	repo := NewMockElementRepository()
	tool := NewListElementsTool(repo)
	require.NoError(t, server.RegisterTool(tool))

	t.Run("call existing tool", func(t *testing.T) {
		args := json.RawMessage(`{}`)
		result, err := server.handleCallTool("list_elements", args)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("call non-existing tool", func(t *testing.T) {
		args := json.RawMessage(`{}`)
		_, err := server.handleCallTool("non_existing", args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestListElementsTool_Execute(t *testing.T) {
	repo := NewMockElementRepository()
	tool := NewListElementsTool(repo)

	t.Run("execute with empty args", func(t *testing.T) {
		result, err := tool.Execute(json.RawMessage(`{}`))
		require.NoError(t, err)

		resultMap := result.(map[string]interface{})
		assert.Equal(t, 0, resultMap["count"])
	})

	t.Run("execute with invalid json", func(t *testing.T) {
		_, err := tool.Execute(json.RawMessage(`{invalid json`))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid arguments")
	})

	t.Run("execute with valid filter", func(t *testing.T) {
		args := json.RawMessage(`{"limit": 10, "offset": 0}`)
		result, err := tool.Execute(args)
		require.NoError(t, err)

		resultMap := result.(map[string]interface{})
		assert.NotNil(t, resultMap["elements"])
		assert.NotNil(t, resultMap["count"])
	})
}

func TestServerCapabilities(t *testing.T) {
	server := NewServer("test", "1.0.0")

	assert.True(t, server.capabilities.Tools)
	assert.False(t, server.capabilities.Resources)
	assert.False(t, server.capabilities.Prompts)
}
