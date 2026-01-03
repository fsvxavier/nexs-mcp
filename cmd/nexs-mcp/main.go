package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsvxavier/nexs-mcp/internal/config"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
	"github.com/fsvxavier/nexs-mcp/internal/mcp"
)

const version = "1.4.0"

func main() {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received, gracefully shutting down...")
		cancel()
	}()

	// Initialize and run server
	if err := run(ctx); err != nil {
		logger.Error("Server error", "error", err)
		cancel() // Ensure cleanup before exit
		//nolint:gocritic // exitAfterDefer is intentional - cancel() must run before exit
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
}

func run(ctx context.Context) error {
	// Load configuration
	cfg := config.LoadConfig(version)

	// Initialize structured logger with buffer (1000 entries)
	logCfg := &logger.Config{
		Level:     parseLogLevel(cfg.LogLevel),
		Format:    cfg.LogFormat,
		Output:    os.Stderr,
		AddSource: false,
	}
	logger.InitWithBuffer(logCfg, 1000)

	// Log startup information
	logger.Info("Starting NEXS MCP Server",
		"version", cfg.Version,
		"storage_type", cfg.StorageType,
		"log_level", cfg.LogLevel,
		"log_format", cfg.LogFormat,
		"onnx_support", getONNXStatus())

	// Create repository based on configuration
	var repo domain.ElementRepository
	var err error

	switch cfg.StorageType {
	case "file":
		logger.Info("Initializing file-based storage", "data_dir", cfg.DataDir)
		repo, err = infrastructure.NewFileElementRepository(cfg.DataDir)
		if err != nil {
			return fmt.Errorf("failed to create file repository: %w", err)
		}
	case "memory":
		logger.Info("Initializing in-memory storage")
		repo = infrastructure.NewInMemoryElementRepository()
	default:
		return fmt.Errorf("invalid storage type: %s (must be 'memory' or 'file')", cfg.StorageType)
	}

	// Create MCP server using official SDK
	server := mcp.NewMCPServer(cfg.ServerName, cfg.Version, repo, cfg)

	logger.Info("MCP Server initialized",
		"server_name", cfg.ServerName,
		"tools_registered", "104",
		"resources_enabled", cfg.Resources.Enabled)
	logger.Info("Server ready. Listening on stdio...")

	// Start server with stdio transport
	return server.Run(ctx)
}

// parseLogLevel converts string log level to slog.Level.
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
