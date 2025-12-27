package mcp

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

// CompressionAlgorithm defines supported compression algorithms.
type CompressionAlgorithm string

const (
	CompressionNone CompressionAlgorithm = "none"
	CompressionGzip CompressionAlgorithm = "gzip"
	CompressionZlib CompressionAlgorithm = "zlib"
	// Future: CompressionBrotli, CompressionZstd.
)

// CompressionConfig holds compression settings.
type CompressionConfig struct {
	Enabled          bool
	Algorithm        CompressionAlgorithm
	MinSize          int  // Only compress if payload > MinSize (default: 1KB)
	CompressionLevel int  // 1-9 for gzip/zlib (default: 6 = balanced)
	AdaptiveMode     bool // Auto-select algorithm based on payload
}

// ResponseCompressor handles MCP response compression.
type ResponseCompressor struct {
	config CompressionConfig
	stats  CompressionStats
	mu     sync.RWMutex
}

// CompressionStats tracks compression metrics.
type CompressionStats struct {
	TotalRequests       int64
	CompressedRequests  int64
	BytesSaved          int64
	AvgCompressionRatio float64
}

// CompressionMetadata describes compression details.
type CompressionMetadata struct {
	Algorithm        CompressionAlgorithm `json:"algorithm"`
	OriginalSize     int                  `json:"original_size"`
	CompressedSize   int                  `json:"compressed_size"`
	CompressionRatio float64              `json:"compression_ratio"`
	Enabled          bool                 `json:"enabled"`
}

// NewResponseCompressor creates a new ResponseCompressor.
func NewResponseCompressor(config CompressionConfig) *ResponseCompressor {
	// Set defaults
	if config.MinSize == 0 {
		config.MinSize = 1024 // 1KB
	}
	if config.CompressionLevel == 0 {
		config.CompressionLevel = 6 // Balanced
	}
	if config.CompressionLevel < 1 || config.CompressionLevel > 9 {
		config.CompressionLevel = 6
	}

	return &ResponseCompressor{
		config: config,
		stats: CompressionStats{
			AvgCompressionRatio: 1.0,
		},
	}
}

// CompressResponse compresses a response payload.
func (c *ResponseCompressor) CompressResponse(data interface{}) ([]byte, CompressionMetadata, error) {
	// Marshal to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, CompressionMetadata{}, fmt.Errorf("failed to marshal data: %w", err)
	}

	originalSize := len(jsonData)
	atomic.AddInt64(&c.stats.TotalRequests, 1)

	// Skip compression if disabled or below threshold
	if !c.config.Enabled || originalSize < c.config.MinSize {
		return jsonData, CompressionMetadata{
			Algorithm:      CompressionNone,
			OriginalSize:   originalSize,
			CompressedSize: originalSize,
			Enabled:        false,
		}, nil
	}

	// Select algorithm
	algorithm := c.config.Algorithm
	if c.config.AdaptiveMode {
		algorithm = c.selectOptimalAlgorithm(jsonData)
	}

	// Compress
	compressed, err := c.compress(jsonData, algorithm)
	if err != nil {
		// Fallback to uncompressed on error
		return jsonData, CompressionMetadata{
			Algorithm:      CompressionNone,
			OriginalSize:   originalSize,
			CompressedSize: originalSize,
			Enabled:        false,
		}, nil
	}

	compressedSize := len(compressed)

	// Only use compression if it actually reduces size
	if compressedSize >= originalSize {
		return jsonData, CompressionMetadata{
			Algorithm:      CompressionNone,
			OriginalSize:   originalSize,
			CompressedSize: originalSize,
			Enabled:        false,
		}, nil
	}

	compressionRatio := float64(compressedSize) / float64(originalSize)

	// Update stats
	c.updateStats(originalSize, compressedSize, compressionRatio)

	return compressed, CompressionMetadata{
		Algorithm:        algorithm,
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: compressionRatio,
		Enabled:          true,
	}, nil
}

// compress performs the actual compression.
func (c *ResponseCompressor) compress(data []byte, algorithm CompressionAlgorithm) ([]byte, error) {
	var buf bytes.Buffer

	switch algorithm {
	case CompressionGzip:
		writer, err := gzip.NewWriterLevel(&buf, c.config.CompressionLevel)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip writer: %w", err)
		}
		if _, err := writer.Write(data); err != nil {
			_ = writer.Close() // Ignore close error when write fails
			return nil, fmt.Errorf("failed to write gzip data: %w", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip writer: %w", err)
		}
		return buf.Bytes(), nil

	case CompressionZlib:
		writer, err := zlib.NewWriterLevel(&buf, c.config.CompressionLevel)
		if err != nil {
			return nil, fmt.Errorf("failed to create zlib writer: %w", err)
		}
		if _, err := writer.Write(data); err != nil {
			_ = writer.Close() // Ignore close error when write fails
			return nil, fmt.Errorf("failed to write zlib data: %w", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close zlib writer: %w", err)
		}
		return buf.Bytes(), nil

	case CompressionNone:
		return data, nil

	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}
}

// selectOptimalAlgorithm chooses the best algorithm for the payload.
func (c *ResponseCompressor) selectOptimalAlgorithm(data []byte) CompressionAlgorithm {
	// Heuristic: gzip for JSON (best compression ratio for structured data)
	// In the future, we can add benchmarks to compare gzip vs zlib vs zstd vs brotli
	return CompressionGzip
}

// updateStats updates compression statistics.
func (c *ResponseCompressor) updateStats(originalSize, compressedSize int, compressionRatio float64) {
	atomic.AddInt64(&c.stats.CompressedRequests, 1)
	atomic.AddInt64(&c.stats.BytesSaved, int64(originalSize-compressedSize))

	// Update average compression ratio using exponential moving average
	c.mu.Lock()
	alpha := 0.1 // Smoothing factor
	c.stats.AvgCompressionRatio = alpha*compressionRatio + (1-alpha)*c.stats.AvgCompressionRatio
	c.mu.Unlock()
}

// GetStats returns current compression statistics.
func (c *ResponseCompressor) GetStats() CompressionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CompressionStats{
		TotalRequests:       atomic.LoadInt64(&c.stats.TotalRequests),
		CompressedRequests:  atomic.LoadInt64(&c.stats.CompressedRequests),
		BytesSaved:          atomic.LoadInt64(&c.stats.BytesSaved),
		AvgCompressionRatio: c.stats.AvgCompressionRatio,
	}
}

// EncodeCompressed encodes compressed data as base64 for JSON transport.
func EncodeCompressed(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeCompressed decodes base64-encoded compressed data.
func DecodeCompressed(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// CompressedResponse wraps a compressed response with metadata.
type CompressedResponse struct {
	Compressed bool                `json:"compressed"`
	Data       string              `json:"data"` // Base64-encoded compressed data
	Metadata   CompressionMetadata `json:"metadata"`
}

// NewCompressedResponse creates a compressed response wrapper.
func NewCompressedResponse(compressedData []byte, metadata CompressionMetadata) CompressedResponse {
	return CompressedResponse{
		Compressed: true,
		Data:       EncodeCompressed(compressedData),
		Metadata:   metadata,
	}
}
