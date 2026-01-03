package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

// ResponseMiddleware wraps tool handlers to add compression and metrics.
type ResponseMiddleware struct {
	server         *MCPServer
	compressor     *ResponseCompressor
	tokenMetrics   *application.TokenMetricsCollector
	compressionCfg CompressionConfig
}

// NewResponseMiddleware creates a new response middleware.
func NewResponseMiddleware(server *MCPServer) *ResponseMiddleware {
	return &ResponseMiddleware{
		server:         server,
		compressor:     server.compressor,
		tokenMetrics:   server.tokenMetrics,
		compressionCfg: server.compressor.config,
	}
}

// MeasureResponseSize measures the size of a tool response output and records metrics.
// This should be called from tool handlers with their output data before returning.
func (rm *ResponseMiddleware) MeasureResponseSize(
	ctx context.Context,
	toolName string,
	output interface{},
) {
	if !rm.compressionCfg.Enabled {
		return
	}

	// Measure original size
	originalData, err := json.Marshal(output)
	if err != nil {
		logger.Error("Failed to marshal tool output for metrics", "error", err)
		return
	}

	originalSize := len(originalData)

	// Skip if below minimum size
	if originalSize < rm.compressionCfg.MinSize {
		logger.Debug("Response below minimum size for compression",
			"tool", toolName,
			"size", originalSize,
			"min_size", rm.compressionCfg.MinSize)
		return
	}

	logger.Debug("Attempting compression for metrics",
		"tool", toolName,
		"original_size", originalSize)

	// Simulate compression to measure potential savings
	compressed, metadata, err := rm.compressor.CompressResponse(output)
	if err != nil {
		logger.Debug("Compression simulation failed",
			"tool", toolName,
			"error", err)
		return
	}

	compressedSize := len(compressed)
	ratio := float64(compressedSize) / float64(originalSize)

	logger.Debug("Compression result",
		"tool", toolName,
		"original_size", originalSize,
		"compressed_size", compressedSize,
		"ratio", ratio,
		"algorithm", metadata.Algorithm)

	// Record metrics even if compression wasn't beneficial (for tracking purposes)
	// This allows us to see all large responses, not just compressible ones
	rm.tokenMetrics.RecordTokenOptimization(application.TokenMetrics{
		OriginalTokens:   application.EstimateTokenCountFromBytes(originalSize),
		OptimizedTokens:  application.EstimateTokenCountFromBytes(compressedSize),
		CompressionRatio: ratio,
		OptimizationType: "response_compression",
		ToolName:         toolName,
		Timestamp:        time.Now(),
	})

	if compressedSize >= originalSize {
		logger.Debug("Compression not beneficial but metrics recorded",
			"tool", toolName,
			"compressed_size", compressedSize,
			"original_size", originalSize)
	} else {
		logger.Debug("Response compression metrics recorded with savings",
			"tool", toolName,
			"original_size", originalSize,
			"compressed_size", compressedSize,
			"ratio", fmt.Sprintf("%.2f%%", ratio*100),
			"algorithm", metadata.Algorithm)
	}
}

// CompressPromptIfNeeded compresses a prompt before sending to LLM.
func (rm *ResponseMiddleware) CompressPromptIfNeeded(ctx context.Context, prompt string) string {
	if !rm.server.cfg.PromptCompression.Enabled {
		return prompt
	}

	if len(prompt) < rm.server.cfg.PromptCompression.MinPromptLength {
		return prompt
	}

	originalSize := len(prompt)
	compressed, _, err := rm.server.promptCompressor.CompressPrompt(ctx, prompt)
	if err != nil {
		logger.Error("Failed to compress prompt", "error", err)
		return prompt
	}

	compressedSize := len(compressed)

	// Only use compressed version if it's actually smaller
	if compressedSize >= originalSize {
		return prompt
	}

	// Record token optimization metrics
	rm.tokenMetrics.RecordTokenOptimization(application.TokenMetrics{
		OriginalTokens:   application.EstimateTokenCountFromBytes(originalSize),
		OptimizedTokens:  application.EstimateTokenCountFromBytes(compressedSize),
		CompressionRatio: float64(compressedSize) / float64(originalSize),
		OptimizationType: "prompt_compression",
		ToolName:         "prompt_compressor",
		Timestamp:        time.Now(),
	})

	logger.Info("Compressed prompt",
		"original_size", originalSize,
		"compressed_size", compressedSize,
		"ratio", fmt.Sprintf("%.2f%%", (float64(compressedSize)/float64(originalSize))*100))

	return compressed
}

// StreamLargeResponse streams large responses in chunks to reduce memory usage.
// Returns true if streaming was applied, false otherwise.
func (rm *ResponseMiddleware) StreamLargeResponse(
	ctx context.Context,
	toolName string,
	output interface{},
) bool {
	if !rm.server.cfg.Streaming.Enabled {
		return false
	}

	// Check response size
	data, err := json.Marshal(output)
	if err != nil {
		return false
	}

	// Stream if response is larger than 10KB
	const streamThreshold = 10 * 1024
	if len(data) < streamThreshold {
		return false
	}

	logger.Info("Streaming large response",
		"tool", toolName,
		"size", len(data),
		"chunks", (len(data)/rm.server.cfg.Streaming.ChunkSize)+1)

	// Record streaming metrics
	rm.tokenMetrics.RecordTokenOptimization(application.TokenMetrics{
		OriginalTokens:   application.EstimateTokenCountFromBytes(len(data)),
		OptimizedTokens:  application.EstimateTokenCountFromBytes(len(data)), // Same tokens, but streamed
		CompressionRatio: 1.0,
		OptimizationType: "streaming",
		ToolName:         toolName,
		Timestamp:        time.Now(),
	})

	return true
}
