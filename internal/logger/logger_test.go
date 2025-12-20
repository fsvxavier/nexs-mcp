package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != slog.LevelInfo {
		t.Errorf("Expected level Info, got %v", cfg.Level)
	}

	if cfg.Format != "json" {
		t.Errorf("Expected format json, got %s", cfg.Format)
	}

	if cfg.AddSource {
		t.Error("Expected AddSource to be false")
	}
}

func TestInit_JSONHandler(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:     slog.LevelDebug,
		Format:    "json",
		Output:    &buf,
		AddSource: false,
	}

	Init(cfg)

	logger := Get()
	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Expected log to contain message")
	}

	if !strings.Contains(output, "key") {
		t.Error("Expected log to contain key")
	}

	// Verify it's valid JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}
}

func TestInit_TextHandler(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &buf,
	}

	Init(cfg)

	logger := Get()
	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Expected log to contain message")
	}

	if !strings.Contains(output, "key=value") {
		t.Error("Expected log to contain key=value")
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:  slog.LevelDebug,
		Format: "json",
		Output: &buf,
	}

	Init(cfg)

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()

	if !strings.Contains(output, "debug message") {
		t.Error("Expected debug message")
	}

	if !strings.Contains(output, "info message") {
		t.Error("Expected info message")
	}

	if !strings.Contains(output, "warn message") {
		t.Error("Expected warn message")
	}

	if !strings.Contains(output, "error message") {
		t.Error("Expected error message")
	}
}

func TestWithContext_RequestID(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: &buf,
	}

	Init(cfg)

	ctx := context.WithValue(context.Background(), RequestIDKey, "req-123")
	InfoContext(ctx, "test message")

	output := buf.String()

	if !strings.Contains(output, "req-123") {
		t.Error("Expected log to contain request ID")
	}

	if !strings.Contains(output, "request_id") {
		t.Error("Expected log to have request_id field")
	}
}

func TestWithContext_User(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: &buf,
	}

	Init(cfg)

	ctx := context.WithValue(context.Background(), UserKey, "alice")
	InfoContext(ctx, "test message")

	output := buf.String()

	if !strings.Contains(output, "alice") {
		t.Error("Expected log to contain user")
	}

	if !strings.Contains(output, "user") {
		t.Error("Expected log to have user field")
	}
}

func TestWithContext_Operation(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: &buf,
	}

	Init(cfg)

	ctx := context.WithValue(context.Background(), OperationKey, "create_element")
	InfoContext(ctx, "test message")

	output := buf.String()

	if !strings.Contains(output, "create_element") {
		t.Error("Expected log to contain operation")
	}

	if !strings.Contains(output, "operation") {
		t.Error("Expected log to have operation field")
	}
}

func TestWithContext_Tool(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: &buf,
	}

	Init(cfg)

	ctx := context.WithValue(context.Background(), ToolKey, "search_memory")
	InfoContext(ctx, "test message")

	output := buf.String()

	if !strings.Contains(output, "search_memory") {
		t.Error("Expected log to contain tool name")
	}

	if !strings.Contains(output, "tool") {
		t.Error("Expected log to have tool field")
	}
}

func TestWithContext_MultipleFields(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: &buf,
	}

	Init(cfg)

	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "req-456")
	ctx = context.WithValue(ctx, UserKey, "bob")
	ctx = context.WithValue(ctx, OperationKey, "update_element")
	ctx = context.WithValue(ctx, ToolKey, "update_memory")

	InfoContext(ctx, "test message")

	output := buf.String()

	expectedFields := []string{"req-456", "bob", "update_element", "update_memory"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected log to contain %s", field)
		}
	}
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: &buf,
	}

	Init(cfg)

	logger := With("service", "mcp-server", "version", "0.1.0")
	logger.Info("test message")

	output := buf.String()

	if !strings.Contains(output, "mcp-server") {
		t.Error("Expected log to contain service")
	}

	if !strings.Contains(output, "0.1.0") {
		t.Error("Expected log to contain version")
	}
}

func TestContextLogging(t *testing.T) {
	var buf bytes.Buffer

	cfg := &Config{
		Level:  slog.LevelDebug,
		Format: "json",
		Output: &buf,
	}

	Init(cfg)

	ctx := context.WithValue(context.Background(), UserKey, "charlie")

	DebugContext(ctx, "debug with context", "detail", "value1")
	buf.Reset()

	InfoContext(ctx, "info with context", "detail", "value2")
	output := buf.String()
	if !strings.Contains(output, "charlie") || !strings.Contains(output, "value2") {
		t.Error("Expected info context to include user and detail")
	}
	buf.Reset()

	WarnContext(ctx, "warn with context", "detail", "value3")
	output = buf.String()
	if !strings.Contains(output, "charlie") || !strings.Contains(output, "value3") {
		t.Error("Expected warn context to include user and detail")
	}
	buf.Reset()

	ErrorContext(ctx, "error with context", "detail", "value4")
	output = buf.String()
	if !strings.Contains(output, "charlie") || !strings.Contains(output, "value4") {
		t.Error("Expected error context to include user and detail")
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer

	// Set level to Warn
	cfg := &Config{
		Level:  slog.LevelWarn,
		Format: "json",
		Output: &buf,
	}

	Init(cfg)

	Debug("debug message - should not appear")
	Info("info message - should not appear")
	Warn("warn message - should appear")
	Error("error message - should appear")

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}

	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}

	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be present")
	}

	if !strings.Contains(output, "error message") {
		t.Error("Error message should be present")
	}
}
