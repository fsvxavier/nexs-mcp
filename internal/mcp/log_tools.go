package mcp

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

// --- List Logs Input/Output structures ---

// ListLogsInput defines input for list_logs tool
type ListLogsInput struct {
	Level     string `json:"level,omitempty" jsonschema:"minimum log level (debug, info, warn, error)"`
	DateFrom  string `json:"date_from,omitempty" jsonschema:"filter logs from this date/time (RFC3339 format)"`
	DateTo    string `json:"date_to,omitempty" jsonschema:"filter logs to this date/time (RFC3339 format)"`
	Keyword   string `json:"keyword,omitempty" jsonschema:"search keyword in log messages"`
	Limit     int    `json:"limit,omitempty" jsonschema:"maximum number of log entries to return (default: 100)"`
	User      string `json:"user,omitempty" jsonschema:"filter by user attribute"`
	Operation string `json:"operation,omitempty" jsonschema:"filter by operation attribute"`
	Tool      string `json:"tool,omitempty" jsonschema:"filter by MCP tool name"`
}

// ListLogsOutput defines output for list_logs tool
type ListLogsOutput struct {
	Logs       []LogEntrySummary `json:"logs" jsonschema:"list of log entries"`
	Total      int               `json:"total" jsonschema:"total number of logs returned"`
	Summary    string            `json:"summary" jsonschema:"summary of log query results"`
	BufferSize int               `json:"buffer_size" jsonschema:"total logs in buffer"`
}

// LogEntrySummary represents a log entry in the output
type LogEntrySummary struct {
	Time       string            `json:"time" jsonschema:"log timestamp (RFC3339)"`
	Level      string            `json:"level" jsonschema:"log level"`
	Message    string            `json:"message" jsonschema:"log message"`
	Attributes map[string]string `json:"attributes,omitempty" jsonschema:"log attributes"`
}

// handleListLogs handles the list_logs tool
func (s *MCPServer) handleListLogs(ctx context.Context, req *sdk.CallToolRequest, input ListLogsInput) (*sdk.CallToolResult, ListLogsOutput, error) {
	// Get log buffer
	buffer := logger.GetLogBuffer()
	if buffer == nil {
		return nil, ListLogsOutput{}, fmt.Errorf("log buffer not initialized")
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 100
	}

	// Build filter
	filter := logger.LogFilter{
		Level:     input.Level,
		Keyword:   input.Keyword,
		Limit:     input.Limit,
		User:      input.User,
		Operation: input.Operation,
		Tool:      input.Tool,
	}

	// Parse date filters
	var err error
	if input.DateFrom != "" {
		filter.DateFrom, err = time.Parse(time.RFC3339, input.DateFrom)
		if err != nil {
			return nil, ListLogsOutput{}, fmt.Errorf("invalid date_from format (use RFC3339): %w", err)
		}
	}

	if input.DateTo != "" {
		filter.DateTo, err = time.Parse(time.RFC3339, input.DateTo)
		if err != nil {
			return nil, ListLogsOutput{}, fmt.Errorf("invalid date_to format (use RFC3339): %w", err)
		}
	}

	// Query logs
	entries := buffer.Query(filter)

	// Convert to output format
	logs := make([]LogEntrySummary, len(entries))
	for i, entry := range entries {
		logs[i] = LogEntrySummary{
			Time:       entry.Time.Format(time.RFC3339),
			Level:      entry.Level,
			Message:    entry.Message,
			Attributes: entry.Attributes,
		}
	}

	// Build summary
	summary := fmt.Sprintf("Found %d log entries", len(logs))
	if input.Level != "" {
		summary += fmt.Sprintf(" (level >= %s)", input.Level)
	}
	if input.Keyword != "" {
		summary += fmt.Sprintf(" matching '%s'", input.Keyword)
	}

	output := ListLogsOutput{
		Logs:       logs,
		Total:      len(logs),
		Summary:    summary,
		BufferSize: buffer.Size(),
	}

	return nil, output, nil
}
