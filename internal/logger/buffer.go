package logger

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

// LogEntry represents a single log entry.
type LogEntry struct {
	Time       time.Time         `json:"time"`
	Level      string            `json:"level"`
	Message    string            `json:"message"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// LogBuffer is a circular buffer that stores recent log entries.
type LogBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
	index   int
}

// NewLogBuffer creates a new log buffer with the specified maximum size.
func NewLogBuffer(maxSize int) *LogBuffer {
	if maxSize <= 0 {
		maxSize = 1000 // default size
	}
	return &LogBuffer{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds a log entry to the buffer.
func (lb *LogBuffer) Add(entry LogEntry) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.entries) < lb.maxSize {
		lb.entries = append(lb.entries, entry)
	} else {
		lb.entries[lb.index] = entry
		lb.index = (lb.index + 1) % lb.maxSize
	}
}

// Query retrieves log entries matching the given criteria.
func (lb *LogBuffer) Query(filter LogFilter) []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var results []LogEntry

	for _, entry := range lb.entries {
		if filter.Matches(entry) {
			results = append(results, entry)
		}
	}

	// Sort by time descending (newest first)
	for i := range len(results) / 2 {
		j := len(results) - 1 - i
		results[i], results[j] = results[j], results[i]
	}

	// Apply limit
	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return results
}

// Clear removes all log entries from the buffer.
func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.entries = make([]LogEntry, 0, lb.maxSize)
	lb.index = 0
}

// Size returns the current number of entries in the buffer.
func (lb *LogBuffer) Size() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return len(lb.entries)
}

// LogFilter defines criteria for filtering log entries.
type LogFilter struct {
	Level     string    // Minimum log level (debug, info, warn, error)
	DateFrom  time.Time // Filter entries after this time
	DateTo    time.Time // Filter entries before this time
	Keyword   string    // Search keyword in message
	Limit     int       // Maximum number of results
	User      string    // Filter by user attribute
	Operation string    // Filter by operation attribute
	Tool      string    // Filter by tool attribute
}

// Matches checks if a log entry matches the filter criteria.
func (f LogFilter) Matches(entry LogEntry) bool {
	// Level filtering
	if f.Level != "" {
		entryLevel := parseLevelValue(entry.Level)
		filterLevel := parseLevelValue(f.Level)
		if entryLevel < filterLevel {
			return false
		}
	}

	// Date range filtering
	if !f.DateFrom.IsZero() && entry.Time.Before(f.DateFrom) {
		return false
	}
	if !f.DateTo.IsZero() && entry.Time.After(f.DateTo) {
		return false
	}

	// Keyword filtering
	if f.Keyword != "" {
		found := false
		if contains(entry.Message, f.Keyword) {
			found = true
		}
		for _, v := range entry.Attributes {
			if contains(v, f.Keyword) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Attribute filtering
	if f.User != "" && entry.Attributes["user"] != f.User {
		return false
	}
	if f.Operation != "" && entry.Attributes["operation"] != f.Operation {
		return false
	}
	if f.Tool != "" && entry.Attributes["tool"] != f.Tool {
		return false
	}

	return true
}

// parseLevelValue converts level string to numeric value for comparison.
func parseLevelValue(level string) int {
	switch level {
	case "DEBUG", "debug":
		return 0
	case "INFO", "info":
		return 1
	case "WARN", "warn":
		return 2
	case "ERROR", "error":
		return 3
	default:
		return 1 // default to INFO
	}
}

// contains checks if s contains substr (case-insensitive).
func contains(s, substr string) bool {
	// Simple case-insensitive contains
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	// Convert both to lowercase for comparison
	sLower := toLower(s)
	substrLower := toLower(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

// toLower converts string to lowercase.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range len(s) {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// BufferedHandler wraps an slog.Handler and captures logs to a buffer.
type BufferedHandler struct {
	handler slog.Handler
	buffer  *LogBuffer
}

// NewBufferedHandler creates a new buffered handler.
func NewBufferedHandler(handler slog.Handler, buffer *LogBuffer) *BufferedHandler {
	return &BufferedHandler{
		handler: handler,
		buffer:  buffer,
	}
}

// Enabled implements slog.Handler.
func (h *BufferedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (h *BufferedHandler) Handle(ctx context.Context, record slog.Record) error {
	// Capture to buffer
	entry := LogEntry{
		Time:       record.Time,
		Level:      record.Level.String(),
		Message:    record.Message,
		Attributes: make(map[string]string),
	}

	record.Attrs(func(attr slog.Attr) bool {
		entry.Attributes[attr.Key] = attr.Value.String()
		return true
	})

	h.buffer.Add(entry)

	// Forward to wrapped handler
	return h.handler.Handle(ctx, record)
}

// WithAttrs implements slog.Handler.
func (h *BufferedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &BufferedHandler{
		handler: h.handler.WithAttrs(attrs),
		buffer:  h.buffer,
	}
}

// WithGroup implements slog.Handler.
func (h *BufferedHandler) WithGroup(name string) slog.Handler {
	return &BufferedHandler{
		handler: h.handler.WithGroup(name),
		buffer:  h.buffer,
	}
}

// globalLogBuffer is the global log buffer instance.
var globalLogBuffer *LogBuffer

// InitWithBuffer initializes the logger with a buffered handler.
func InitWithBuffer(cfg *Config, bufferSize int) *LogBuffer {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	var baseHandler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	if cfg.Format == "json" {
		baseHandler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		baseHandler = slog.NewTextHandler(cfg.Output, opts)
	}

	// Create buffer and buffered handler
	globalLogBuffer = NewLogBuffer(bufferSize)
	bufferedHandler := NewBufferedHandler(baseHandler, globalLogBuffer)

	defaultLogger = slog.New(bufferedHandler)
	slog.SetDefault(defaultLogger)

	// Mark initOnce as done so subsequent calls to Get() don't override our buffered logger.
	initOnce.Do(func() {})

	return globalLogBuffer
}

// GetLogBuffer returns the global log buffer.
func GetLogBuffer() *LogBuffer {
	return globalLogBuffer
}
