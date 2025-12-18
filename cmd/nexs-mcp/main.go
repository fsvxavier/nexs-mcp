package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsvxavier/nexs-mcp/internal/config"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/mcp"
)

const version = "0.1.0"

func main() {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, gracefully shutting down...")
		cancel()
	}()

	// Initialize and run server
	if err := run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server shutdown complete")
}

func run(ctx context.Context) error {
	// Load configuration
	cfg := config.LoadConfig(version)

	// Use stderr for all logs to keep stdout clean for JSON-RPC
	fmt.Fprintf(os.Stderr, "NEXS MCP Server v%s\n", cfg.Version)
	fmt.Fprintln(os.Stderr, "Initializing Model Context Protocol server...")
	fmt.Fprintf(os.Stderr, "Storage type: %s\n", cfg.StorageType)

	// Create repository based on configuration
	var repo domain.ElementRepository
	var err error

	switch cfg.StorageType {
	case "file":
		fmt.Fprintf(os.Stderr, "Data directory: %s\n", cfg.DataDir)
		repo, err = infrastructure.NewFileElementRepository(cfg.DataDir)
		if err != nil {
			return fmt.Errorf("failed to create file repository: %w", err)
		}
	case "memory":
		repo = infrastructure.NewInMemoryElementRepository()
	default:
		return fmt.Errorf("invalid storage type: %s (must be 'memory' or 'file')", cfg.StorageType)
	}

	// Create MCP server using official SDK
	server := mcp.NewMCPServer(cfg.ServerName, cfg.Version, repo)

	fmt.Fprintln(os.Stderr, "Registered 5 tools:")
	fmt.Fprintln(os.Stderr, "  - list_elements: List all elements with optional filtering")
	fmt.Fprintln(os.Stderr, "  - get_element: Get a specific element by ID")
	fmt.Fprintln(os.Stderr, "  - create_element: Create a new element")
	fmt.Fprintln(os.Stderr, "  - update_element: Update an existing element")
	fmt.Fprintln(os.Stderr, "  - delete_element: Delete an element by ID")
	fmt.Fprintln(os.Stderr, "Server ready. Listening on stdio...")

	// Start server with stdio transport
	return server.Run(ctx)
}
