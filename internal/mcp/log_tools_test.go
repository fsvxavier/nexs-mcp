package mcp

import (
	"context"
	"log/slog"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

func setupLogTestServer(t *testing.T) *MCPServer {
	t.Helper()
	repo := infrastructure.NewInMemoryElementRepository()

	// Initialize logger with buffer for testing
	logCfg := &logger.Config{
		Level:  slog.LevelDebug,
		Format: "json",
	}
	logger.InitWithBuffer(logCfg, 100)

	return NewMCPServer("nexs-mcp-test", "0.1.0", repo)
}

func TestHandleListLogs(t *testing.T) {
	server := setupLogTestServer(t)
	ctx := context.Background()

	// Generate some test logs
	logger.Info("Test info message", "user", "alice", "operation", "test_op")
	logger.Warn("Test warning message", "user", "bob")
	logger.Error("Test error message", "tool", "test_tool")

	// Give logs time to be captured
	time.Sleep(10 * time.Millisecond)

	input := ListLogsInput{
		Limit: 10,
	}

	_, output, err := server.handleListLogs(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleListLogs failed: %v", err)
	}

	if output.Total < 3 {
		t.Errorf("Expected at least 3 logs, got %d", output.Total)
	}

	if output.BufferSize == 0 {
		t.Error("Expected non-zero buffer size")
	}

	if output.Summary == "" {
		t.Error("Expected non-empty summary")
	}
}

func TestHandleListLogs_LevelFilter(t *testing.T) {
	server := setupLogTestServer(t)
	ctx := context.Background()

	// Clear buffer
	if buf := logger.GetLogBuffer(); buf != nil {
		buf.Clear()
	}

	// Generate logs with different levels
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	time.Sleep(10 * time.Millisecond)

	// Filter for warnings and above
	input := ListLogsInput{
		Level: "warn",
		Limit: 10,
	}

	_, output, err := server.handleListLogs(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleListLogs failed: %v", err)
	}

	// Should get warn and error
	if output.Total < 2 {
		t.Errorf("Expected at least 2 logs (warn+error), got %d", output.Total)
	}
}

func TestHandleListLogs_KeywordFilter(t *testing.T) {
	server := setupLogTestServer(t)
	ctx := context.Background()

	if buf := logger.GetLogBuffer(); buf != nil {
		buf.Clear()
	}

	// Generate logs with specific keywords
	logger.Info("Creating persona element")
	logger.Info("Updating memory element")

	time.Sleep(10 * time.Millisecond)

	// Search for "persona"
	input := ListLogsInput{
		Keyword: "persona",
		Limit:   10,
	}

	_, output, err := server.handleListLogs(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleListLogs failed: %v", err)
	}

	if output.Total < 1 {
		t.Errorf("Expected at least 1 log matching 'persona', got %d", output.Total)
	}
}

func TestHandleListLogs_UserFilter(t *testing.T) {
	server := setupLogTestServer(t)
	ctx := context.Background()

	if buf := logger.GetLogBuffer(); buf != nil {
		buf.Clear()
	}

	// Generate logs with different users
	logger.Info("Alice action", "user", "alice")
	logger.Info("Bob action", "user", "bob")

	time.Sleep(10 * time.Millisecond)

	// Filter for alice
	input := ListLogsInput{
		User:  "alice",
		Limit: 10,
	}

	_, output, err := server.handleListLogs(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleListLogs failed: %v", err)
	}

	if output.Total < 1 {
		t.Errorf("Expected at least 1 log for user alice, got %d", output.Total)
	}
}

func TestHandleListLogs_InvalidDateFormat(t *testing.T) {
	server := setupLogTestServer(t)
	ctx := context.Background()

	input := ListLogsInput{
		DateFrom: "invalid-date",
		Limit:    10,
	}

	_, _, err := server.handleListLogs(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error for invalid date format")
	}
}

func TestHandleListLogs_Limit(t *testing.T) {
	server := setupLogTestServer(t)
	ctx := context.Background()

	if buf := logger.GetLogBuffer(); buf != nil {
		buf.Clear()
	}

	// Generate many logs
	for i := 0; i < 10; i++ {
		logger.Info("Test log")
	}

	time.Sleep(10 * time.Millisecond)

	// Limit to 5
	input := ListLogsInput{
		Limit: 5,
	}

	_, output, err := server.handleListLogs(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleListLogs failed: %v", err)
	}

	if len(output.Logs) > 5 {
		t.Errorf("Expected max 5 logs, got %d", len(output.Logs))
	}
}
