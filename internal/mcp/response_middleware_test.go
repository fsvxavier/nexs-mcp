package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/config"
)

func TestResponseMiddleware_Basic(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Compression: config.CompressionConfig{
			Enabled:          true,
			Algorithm:        "gzip",
			MinSize:          100,
			CompressionLevel: 6,
		},
		PromptCompression: config.PromptCompressionConfig{
			Enabled:         true,
			MinPromptLength: 50,
		},
	}

	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := newTestServer("test-server", "1.0.0", repo)
	server.compressor = NewResponseCompressor(CompressionConfig{
		Enabled:          true,
		Algorithm:        CompressionGzip,
		MinSize:          100,
		CompressionLevel: 6,
	})
	server.tokenMetrics = application.NewTokenMetricsCollector(tempDir, 5*time.Second)
	server.promptCompressor = application.NewPromptCompressor(application.PromptCompressionConfig{
		Enabled:                true,
		RemoveRedundancy:       true,
		CompressWhitespace:     true,
		PreserveStructure:      true,
		TargetCompressionRatio: 0.7,
		MinPromptLength:        50,
	})
	server.cfg = cfg

	middleware := NewResponseMiddleware(server)

	if middleware == nil {
		t.Fatal("NewResponseMiddleware returned nil")
	}

	t.Log("✓ ResponseMiddleware created successfully")
}

func TestMeasureResponseSize(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Compression: config.CompressionConfig{
			Enabled:          true,
			Algorithm:        "gzip",
			MinSize:          100,
			CompressionLevel: 6,
		},
	}

	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := newTestServer("test-server", "1.0.0", repo)
	server.compressor = NewResponseCompressor(CompressionConfig{
		Enabled:          true,
		Algorithm:        CompressionGzip,
		MinSize:          100,
		CompressionLevel: 6,
	})
	server.tokenMetrics = application.NewTokenMetricsCollector(tempDir, 5*time.Second)
	server.cfg = cfg

	middleware := NewResponseMiddleware(server)
	ctx := context.Background()

	// Create a large output
	output := map[string]interface{}{
		"data":    string(make([]byte, 2000)),
		"status":  "success",
		"message": "Test message",
	}

	middleware.MeasureResponseSize(ctx, "test_tool", output)

	stats := server.tokenMetrics.GetStats()
	t.Logf("Optimizations recorded: %d", stats.OptimizationCount)
	t.Logf("Total tokens saved: %d", stats.TotalTokensSaved)

	t.Log("✓ MeasureResponseSize executed without errors")
}

func TestCompressPromptIfNeeded(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Compression: config.CompressionConfig{
			Enabled: true,
		},
		PromptCompression: config.PromptCompressionConfig{
			Enabled:         true,
			MinPromptLength: 50,
		},
	}

	repo := setupTestRepository(t)
	defer cleanupTestRepository(t, repo)

	server := newTestServer("test-server", "1.0.0", repo)
	server.compressor = NewResponseCompressor(CompressionConfig{Enabled: true})
	server.cfg = cfg
	server.tokenMetrics = application.NewTokenMetricsCollector(tempDir, 5*time.Second)
	server.promptCompressor = application.NewPromptCompressor(application.PromptCompressionConfig{
		Enabled:                true,
		RemoveRedundancy:       true,
		CompressWhitespace:     true,
		PreserveStructure:      true,
		TargetCompressionRatio: 0.7,
		MinPromptLength:        50,
	})

	middleware := NewResponseMiddleware(server)
	ctx := context.Background()

	prompt := "Hello     world.     This     is     a     test     with     spaces."
	result := middleware.CompressPromptIfNeeded(ctx, prompt)

	t.Logf("Original length: %d", len(prompt))
	t.Logf("Result length: %d", len(result))

	if result == "" {
		t.Error("Result should not be empty")
	}

	t.Log("✓ CompressPromptIfNeeded executed successfully")
}
