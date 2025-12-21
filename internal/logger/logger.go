package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// ContextKey is the type for context keys used in logging.
type ContextKey string

const (
	// RequestIDKey is the context key for request ID.
	RequestIDKey ContextKey = "request_id"
	// UserKey is the context key for user.
	UserKey ContextKey = "user"
	// OperationKey is the context key for operation name.
	OperationKey ContextKey = "operation"
	// ToolKey is the context key for MCP tool name.
	ToolKey ContextKey = "tool"
)

var (
	// defaultLogger is the default slog logger instance.
	defaultLogger *slog.Logger
)

// Config holds logger configuration.
type Config struct {
	// Level is the minimum log level (debug, info, warn, error)
	Level slog.Level

	// Format is the output format ("json" or "text")
	Format string

	// Output is where logs should be written (defaults to stderr)
	Output io.Writer

	// AddSource adds source file and line number to log records
	AddSource bool
}

// DefaultConfig returns default logger configuration.
func DefaultConfig() *Config {
	return &Config{
		Level:     slog.LevelInfo,
		Format:    "json",
		Output:    os.Stderr,
		AddSource: false,
	}
}

// Init initializes the global logger with the given configuration.
func Init(cfg *Config) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Get returns the default logger.
func Get() *slog.Logger {
	if defaultLogger == nil {
		Init(DefaultConfig())
	}
	return defaultLogger
}

// WithContext extracts logging fields from context and returns a logger with those fields.
func WithContext(ctx context.Context) *slog.Logger {
	logger := Get()

	// Extract context values and add as attributes
	attrs := make([]any, 0, 8) // pairs of key-value

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}

	if user, ok := ctx.Value(UserKey).(string); ok && user != "" {
		attrs = append(attrs, "user", user)
	}

	if operation, ok := ctx.Value(OperationKey).(string); ok && operation != "" {
		attrs = append(attrs, "operation", operation)
	}

	if tool, ok := ctx.Value(ToolKey).(string); ok && tool != "" {
		attrs = append(attrs, "tool", tool)
	}

	if len(attrs) > 0 {
		logger = logger.With(attrs...)
	}

	return logger
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info message.
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// DebugContext logs a debug message with context.
func DebugContext(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Debug(msg, args...)
}

// InfoContext logs an info message with context.
func InfoContext(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Info(msg, args...)
}

// WarnContext logs a warning message with context.
func WarnContext(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Warn(msg, args...)
}

// ErrorContext logs an error message with context.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Error(msg, args...)
}

// With returns a logger with the given attributes.
func With(args ...any) *slog.Logger {
	return Get().With(args...)
}
