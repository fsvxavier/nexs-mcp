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

	fmt.Printf("NEXS MCP Server v%s\n", cfg.Version)
	fmt.Println("Initializing Model Context Protocol server...")
	fmt.Printf("Storage type: %s\n", cfg.StorageType)

	// Create repository based on configuration
	var repo domain.ElementRepository
	var err error

	switch cfg.StorageType {
	case "file":
		fmt.Printf("Data directory: %s\n", cfg.DataDir)
		repo, err = infrastructure.NewFileElementRepository(cfg.DataDir)
		if err != nil {
			return fmt.Errorf("failed to create file repository: %w", err)
		}
	case "memory":
		repo = infrastructure.NewInMemoryElementRepository()
	default:
		return fmt.Errorf("invalid storage type: %s (must be 'memory' or 'file')", cfg.StorageType)
	}

	// Create MCP server
	server := mcp.NewServer(cfg.ServerName, cfg.Version)

	// Register tools
	if err := server.RegisterTool(mcp.NewListElementsTool(repo)); err != nil {
		return fmt.Errorf("failed to register list_elements tool: %w", err)
	}
	if err := server.RegisterTool(mcp.NewGetElementTool(repo)); err != nil {
		return fmt.Errorf("failed to register get_element tool: %w", err)
	}
	if err := server.RegisterTool(mcp.NewCreateElementTool(repo)); err != nil {
		return fmt.Errorf("failed to register create_element tool: %w", err)
	}
	if err := server.RegisterTool(mcp.NewUpdateElementTool(repo)); err != nil {
		return fmt.Errorf("failed to register update_element tool: %w", err)
	}
	if err := server.RegisterTool(mcp.NewDeleteElementTool(repo)); err != nil {
		return fmt.Errorf("failed to register delete_element tool: %w", err)
	}

	fmt.Printf("Registered %d tools\n", len(server.ListTools()))
	fmt.Println("Server ready. Listening on stdio...")

	// Start server
	return server.Start(ctx)
}
