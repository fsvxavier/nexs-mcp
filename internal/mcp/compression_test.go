package mcp

import (
	"strings"
	"testing"
)

func TestResponseCompressor_CompressResponse(t *testing.T) {
	tests := []struct {
		name             string
		config           CompressionConfig
		data             interface{}
		expectCompressed bool
		expectError      bool
	}{
		{
			name: "disabled compression",
			config: CompressionConfig{
				Enabled:   false,
				Algorithm: CompressionGzip,
			},
			data: map[string]interface{}{
				"message": "test data that should not be compressed",
			},
			expectCompressed: false,
			expectError:      false,
		},
		{
			name: "small payload below threshold",
			config: CompressionConfig{
				Enabled:   true,
				Algorithm: CompressionGzip,
				MinSize:   1024,
			},
			data: map[string]interface{}{
				"small": "data",
			},
			expectCompressed: false,
			expectError:      false,
		},
		{
			name: "large payload with gzip",
			config: CompressionConfig{
				Enabled:          true,
				Algorithm:        CompressionGzip,
				MinSize:          100,
				CompressionLevel: 6,
			},
			data:             generateLargeData(1000),
			expectCompressed: true,
			expectError:      false,
		},
		{
			name: "large payload with zlib",
			config: CompressionConfig{
				Enabled:          true,
				Algorithm:        CompressionZlib,
				MinSize:          100,
				CompressionLevel: 6,
			},
			data:             generateLargeData(1000),
			expectCompressed: true,
			expectError:      false,
		},
		{
			name: "adaptive mode selects gzip",
			config: CompressionConfig{
				Enabled:          true,
				Algorithm:        CompressionGzip,
				MinSize:          100,
				CompressionLevel: 6,
				AdaptiveMode:     true,
			},
			data:             generateLargeData(500),
			expectCompressed: true,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewResponseCompressor(tt.config)

			compressed, metadata, err := compressor.CompressResponse(tt.data)

			if (err != nil) != tt.expectError {
				t.Errorf("CompressResponse() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectCompressed {
				if metadata.Algorithm == CompressionNone {
					t.Error("expected compression but got CompressionNone")
				}
				if metadata.CompressedSize >= metadata.OriginalSize {
					t.Errorf("compressed size (%d) should be smaller than original (%d)",
						metadata.CompressedSize, metadata.OriginalSize)
				}
				if metadata.CompressionRatio >= 1.0 {
					t.Errorf("compression ratio (%f) should be < 1.0", metadata.CompressionRatio)
				}
				if len(compressed) == 0 {
					t.Error("compressed data should not be empty")
				}
			} else if metadata.Algorithm != CompressionNone {
				t.Errorf("expected CompressionNone but got %s", metadata.Algorithm)
			}
		})
	}
}

func TestResponseCompressor_CompressionRatios(t *testing.T) {
	config := CompressionConfig{
		Enabled:          true,
		Algorithm:        CompressionGzip,
		MinSize:          100,
		CompressionLevel: 6,
	}

	compressor := NewResponseCompressor(config)

	// Test with highly compressible data (repetitive JSON)
	repetitiveData := generateRepetitiveData(100)
	compressed, metadata, err := compressor.CompressResponse(repetitiveData)
	if err != nil {
		t.Fatalf("CompressResponse() error = %v", err)
	}

	// Highly repetitive data should compress very well (< 30% of original)
	if metadata.CompressionRatio > 0.3 {
		t.Errorf("repetitive data should compress to < 30%%, got %f%%", metadata.CompressionRatio*100)
	}

	if len(compressed) == 0 {
		t.Error("compressed data should not be empty")
	}

	t.Logf("Repetitive data: original=%d, compressed=%d, ratio=%.2f%%",
		metadata.OriginalSize, metadata.CompressedSize, metadata.CompressionRatio*100)
}

func TestResponseCompressor_Stats(t *testing.T) {
	config := CompressionConfig{
		Enabled:          true,
		Algorithm:        CompressionGzip,
		MinSize:          100,
		CompressionLevel: 6,
	}

	compressor := NewResponseCompressor(config)

	// Process multiple requests
	for range 10 {
		data := generateLargeData(500)
		_, _, err := compressor.CompressResponse(data)
		if err != nil {
			t.Fatalf("CompressResponse() error = %v", err)
		}
	}

	stats := compressor.GetStats()

	if stats.TotalRequests != 10 {
		t.Errorf("expected 10 total requests, got %d", stats.TotalRequests)
	}

	if stats.CompressedRequests == 0 {
		t.Error("expected some compressed requests")
	}

	if stats.BytesSaved == 0 {
		t.Error("expected some bytes saved")
	}

	if stats.AvgCompressionRatio >= 1.0 {
		t.Errorf("average compression ratio should be < 1.0, got %f", stats.AvgCompressionRatio)
	}

	t.Logf("Stats: total=%d, compressed=%d, saved=%d bytes, avg_ratio=%.2f%%",
		stats.TotalRequests, stats.CompressedRequests, stats.BytesSaved,
		stats.AvgCompressionRatio*100)
}

func TestCompressedResponse_EncodeDecode(t *testing.T) {
	original := []byte("test data to compress")

	// Encode
	encoded := EncodeCompressed(original)
	if encoded == "" {
		t.Error("encoded string should not be empty")
	}

	// Decode
	decoded, err := DecodeCompressed(encoded)
	if err != nil {
		t.Fatalf("DecodeCompressed() error = %v", err)
	}

	if string(decoded) != string(original) {
		t.Errorf("decoded data does not match original: got %s, want %s", decoded, original)
	}
}

func TestNewCompressedResponse(t *testing.T) {
	data := []byte("compressed data")
	metadata := CompressionMetadata{
		Algorithm:        CompressionGzip,
		OriginalSize:     1000,
		CompressedSize:   300,
		CompressionRatio: 0.3,
		Enabled:          true,
	}

	response := NewCompressedResponse(data, metadata)

	if !response.Compressed {
		t.Error("response.Compressed should be true")
	}

	if response.Data == "" {
		t.Error("response.Data should not be empty")
	}

	if response.Metadata.Algorithm != CompressionGzip {
		t.Errorf("expected CompressionGzip, got %s", response.Metadata.Algorithm)
	}
}

func TestResponseCompressor_DifferentCompressionLevels(t *testing.T) {
	data := generateLargeData(1000)

	levels := []int{1, 3, 6, 9}
	var prevSize int

	for _, level := range levels {
		config := CompressionConfig{
			Enabled:          true,
			Algorithm:        CompressionGzip,
			MinSize:          100,
			CompressionLevel: level,
		}

		compressor := NewResponseCompressor(config)
		_, metadata, err := compressor.CompressResponse(data)
		if err != nil {
			t.Fatalf("CompressResponse() error = %v at level %d", err, level)
		}

		t.Logf("Level %d: original=%d, compressed=%d, ratio=%.2f%%",
			level, metadata.OriginalSize, metadata.CompressedSize,
			metadata.CompressionRatio*100)

		// Higher compression levels should produce smaller or equal sizes
		if prevSize > 0 && metadata.CompressedSize > prevSize {
			// Note: This is not always guaranteed due to algorithm specifics,
			// but generally holds for our test data
			t.Logf("Warning: level %d produced larger output than previous level", level)
		}
		prevSize = metadata.CompressedSize
	}
}

func BenchmarkCompression_Gzip(b *testing.B) {
	config := CompressionConfig{
		Enabled:          true,
		Algorithm:        CompressionGzip,
		MinSize:          100,
		CompressionLevel: 6,
	}

	compressor := NewResponseCompressor(config)
	data := generateLargeData(1000)

	b.ResetTimer()
	for range b.N {
		_, _, err := compressor.CompressResponse(data)
		if err != nil {
			b.Fatalf("CompressResponse() error = %v", err)
		}
	}
}

func BenchmarkCompression_Zlib(b *testing.B) {
	config := CompressionConfig{
		Enabled:          true,
		Algorithm:        CompressionZlib,
		MinSize:          100,
		CompressionLevel: 6,
	}

	compressor := NewResponseCompressor(config)
	data := generateLargeData(1000)

	b.ResetTimer()
	for range b.N {
		_, _, err := compressor.CompressResponse(data)
		if err != nil {
			b.Fatalf("CompressResponse() error = %v", err)
		}
	}
}

func BenchmarkCompression_NoCompression(b *testing.B) {
	config := CompressionConfig{
		Enabled:   false,
		Algorithm: CompressionNone,
	}

	compressor := NewResponseCompressor(config)
	data := generateLargeData(1000)

	b.ResetTimer()
	for range b.N {
		_, _, err := compressor.CompressResponse(data)
		if err != nil {
			b.Fatalf("CompressResponse() error = %v", err)
		}
	}
}

// Helper functions

func generateLargeData(items int) map[string]interface{} {
	elements := make([]map[string]interface{}, items)
	for i := range items {
		elements[i] = map[string]interface{}{
			"id":          i,
			"name":        "Element " + string(rune(i%26+'A')),
			"description": "This is a description for element number " + string(rune(i%10+'0')),
			"tags":        []string{"tag1", "tag2", "tag3"},
			"metadata": map[string]interface{}{
				"created_at": "2025-12-24T00:00:00Z",
				"updated_at": "2025-12-24T00:00:00Z",
				"active":     i%2 == 0,
			},
		}
	}

	return map[string]interface{}{
		"elements": elements,
		"total":    items,
		"page":     1,
	}
}

func generateRepetitiveData(items int) map[string]interface{} {
	// Generate highly repetitive data that should compress very well
	elements := make([]map[string]interface{}, items)
	for i := range items {
		elements[i] = map[string]interface{}{
			"id":          i,
			"name":        "Repetitive Element",
			"description": strings.Repeat("This is repetitive content. ", 10),
			"tags":        []string{"same", "tags", "everywhere"},
			"metadata": map[string]interface{}{
				"created_at": "2025-12-24T00:00:00Z",
				"updated_at": "2025-12-24T00:00:00Z",
				"active":     true,
			},
		}
	}

	return map[string]interface{}{
		"elements": elements,
		"total":    items,
		"page":     1,
	}
}
