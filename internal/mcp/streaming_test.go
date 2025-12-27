package mcp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStreamingHandler_BasicStreaming(t *testing.T) {
	config := StreamingConfig{
		Enabled:      true,
		ChunkSize:    5,
		ThrottleRate: 10 * time.Millisecond,
		BufferSize:   10,
	}

	handler := NewStreamingHandler(config)

	items := make([]interface{}, 20)
	for i := range 20 {
		items[i] = i
	}

	ctx := context.Background()
	chunks := handler.StreamItems(ctx, items)

	receivedChunks := 0
	totalItems := 0

	for chunk := range chunks {
		if chunk.Error != nil {
			t.Fatalf("Unexpected error: %v", chunk.Error)
		}

		receivedChunks++
		totalItems += len(chunk.Items)

		t.Logf("Chunk %d: %d items, hasMore=%v", chunk.ChunkID, len(chunk.Items), chunk.HasMore)

		if len(chunk.Items) > config.ChunkSize {
			t.Errorf("Chunk size %d exceeds max %d", len(chunk.Items), config.ChunkSize)
		}
	}

	expectedChunks := 4 // 20 items / 5 per chunk = 4 chunks
	if receivedChunks != expectedChunks {
		t.Errorf("Expected %d chunks, got %d", expectedChunks, receivedChunks)
	}

	if totalItems != len(items) {
		t.Errorf("Expected %d total items, got %d", len(items), totalItems)
	}
}

func TestStreamingHandler_Disabled(t *testing.T) {
	config := StreamingConfig{
		Enabled:   false,
		ChunkSize: 5,
	}

	handler := NewStreamingHandler(config)

	items := make([]interface{}, 20)
	for i := range 20 {
		items[i] = i
	}

	ctx := context.Background()
	chunks := handler.StreamItems(ctx, items)

	receivedChunks := 0
	for chunk := range chunks {
		receivedChunks++

		if chunk.Error != nil {
			t.Fatalf("Unexpected error: %v", chunk.Error)
		}

		// Should receive all items in one chunk
		if len(chunk.Items) != len(items) {
			t.Errorf("Expected all %d items in one chunk, got %d", len(items), len(chunk.Items))
		}
	}

	if receivedChunks != 1 {
		t.Errorf("Expected 1 chunk when disabled, got %d", receivedChunks)
	}
}

func TestStreamingHandler_ContextCancellation(t *testing.T) {
	config := StreamingConfig{
		Enabled:      true,
		ChunkSize:    5,
		ThrottleRate: 50 * time.Millisecond,
		BufferSize:   5,
	}

	handler := NewStreamingHandler(config)

	items := make([]interface{}, 100)
	for i := range 100 {
		items[i] = i
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	chunks := handler.StreamItems(ctx, items)

	receivedChunks := 0

	// Cancel after receiving 2 chunks
	for chunk := range chunks {
		receivedChunks++

		if receivedChunks == 2 {
			cancel()
		}

		// Should eventually get cancellation error
		if chunk.Error != nil {
			if !errors.Is(chunk.Error, context.Canceled) {
				t.Errorf("Expected context.Canceled, got %v", chunk.Error)
			}
			break
		}
	}

	if receivedChunks < 2 {
		t.Errorf("Expected at least 2 chunks before cancellation, got %d", receivedChunks)
	}

	t.Logf("Received %d chunks before cancellation", receivedChunks)
}

func TestStreamingHandler_CollectChunks(t *testing.T) {
	config := StreamingConfig{
		Enabled:   true,
		ChunkSize: 3,
	}

	handler := NewStreamingHandler(config)

	items := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9}

	ctx := context.Background()
	chunks := handler.StreamItems(ctx, items)

	collected, err := handler.CollectChunks(ctx, chunks)

	if err != nil {
		t.Fatalf("CollectChunks failed: %v", err)
	}

	if len(collected) != len(items) {
		t.Errorf("Expected %d items, got %d", len(items), len(collected))
	}

	// Verify order is preserved
	for i, item := range collected {
		if item != items[i] {
			t.Errorf("Item %d mismatch: expected %v, got %v", i, items[i], item)
		}
	}
}

func TestStreamingHandler_StreamWithTransform(t *testing.T) {
	config := StreamingConfig{
		Enabled:   true,
		ChunkSize: 5,
	}

	handler := NewStreamingHandler(config)

	items := make([]interface{}, 10)
	for i := range 10 {
		items[i] = i
	}

	// Transform: multiply each number by 2
	transform := func(chunk []interface{}) ([]interface{}, error) {
		transformed := make([]interface{}, len(chunk))
		for i, item := range chunk {
			if num, ok := item.(int); ok {
				transformed[i] = num * 2
			}
		}
		return transformed, nil
	}

	ctx := context.Background()
	chunks := handler.StreamWithTransform(ctx, items, transform)

	collected, err := handler.CollectChunks(ctx, chunks)

	if err != nil {
		t.Fatalf("CollectChunks failed: %v", err)
	}

	// Verify transformation
	for i, item := range collected {
		expected := i * 2
		if item != expected {
			t.Errorf("Item %d: expected %v, got %v", i, expected, item)
		}
	}

	t.Logf("Transform test: %d items transformed successfully", len(collected))
}

func TestStreamingHandler_StreamWithFilter(t *testing.T) {
	config := StreamingConfig{
		Enabled:   true,
		ChunkSize: 5,
	}

	handler := NewStreamingHandler(config)

	items := make([]interface{}, 20)
	for i := range 20 {
		items[i] = i
	}

	// Filter: keep only even numbers
	filter := func(item interface{}) bool {
		if num, ok := item.(int); ok {
			return num%2 == 0
		}
		return false
	}

	ctx := context.Background()
	chunks := handler.StreamWithFilter(ctx, items, filter)

	collected, err := handler.CollectChunks(ctx, chunks)

	if err != nil {
		t.Fatalf("CollectChunks failed: %v", err)
	}

	expected := 10 // 0, 2, 4, 6, 8, 10, 12, 14, 16, 18
	if len(collected) != expected {
		t.Errorf("Expected %d items after filter, got %d", expected, len(collected))
	}

	// Verify all are even
	for i, item := range collected {
		if num, ok := item.(int); ok {
			if num%2 != 0 {
				t.Errorf("Item %d (%v) is not even", i, item)
			}
		}
	}

	t.Logf("Filter test: %d/%d items kept", len(collected), len(items))
}

func TestStreamingHandler_Stats(t *testing.T) {
	config := StreamingConfig{
		Enabled:   true,
		ChunkSize: 10,
	}

	handler := NewStreamingHandler(config)

	// Reset stats before test
	handler.ResetStats()

	items := make([]interface{}, 50)
	for i := range 50 {
		items[i] = i
	}

	ctx := context.Background()
	chunks := handler.StreamItems(ctx, items)

	// Consume all chunks
	for range chunks {
	}

	stats := handler.GetStats()

	if stats.TotalStreams != 1 {
		t.Errorf("Expected 1 stream, got %d", stats.TotalStreams)
	}

	expectedChunks := int64(5) // 50 items / 10 per chunk
	if stats.TotalChunks != expectedChunks {
		t.Errorf("Expected %d chunks, got %d", expectedChunks, stats.TotalChunks)
	}

	if stats.TotalItems != 50 {
		t.Errorf("Expected 50 items, got %d", stats.TotalItems)
	}

	t.Logf("Stats: streams=%d, chunks=%d, items=%d, avg_chunk_time=%v",
		stats.TotalStreams, stats.TotalChunks, stats.TotalItems, stats.AvgChunkTime)
}

func TestStreamingHandler_MaxChunks(t *testing.T) {
	config := StreamingConfig{
		Enabled:   true,
		ChunkSize: 5,
		MaxChunks: 3,
	}

	handler := NewStreamingHandler(config)

	items := make([]interface{}, 50)
	for i := range 50 {
		items[i] = i
	}

	ctx := context.Background()
	chunks := handler.StreamItems(ctx, items)

	receivedChunks := 0
	for chunk := range chunks {
		receivedChunks++

		if chunk.Error != nil {
			t.Fatalf("Unexpected error: %v", chunk.Error)
		}
	}

	if receivedChunks != config.MaxChunks {
		t.Errorf("Expected %d chunks (max), got %d", config.MaxChunks, receivedChunks)
	}

	t.Logf("MaxChunks limit: received %d/%d chunks", receivedChunks, config.MaxChunks)
}

func TestStreamingHandler_EstimateMemorySavings(t *testing.T) {
	config := StreamingConfig{
		Enabled:   true,
		ChunkSize: 10,
	}

	handler := NewStreamingHandler(config)

	totalItems := 1000
	itemSize := 1024 // 1KB per item

	savings := handler.EstimateMemorySavings(totalItems, itemSize)

	expectedSavings := int64((totalItems - config.ChunkSize) * itemSize)
	if savings != expectedSavings {
		t.Errorf("Expected savings %d bytes, got %d", expectedSavings, savings)
	}

	t.Logf("Memory savings: %d bytes (%.2f MB)", savings, float64(savings)/(1024*1024))
}

func TestStreamingHandler_CalculateTTFB(t *testing.T) {
	config := StreamingConfig{
		Enabled:   true,
		ChunkSize: 10,
	}

	handler := NewStreamingHandler(config)

	totalItems := 1000

	ttfbImprovement := handler.CalculateTTFB(totalItems)

	if ttfbImprovement <= 0 {
		t.Error("Expected positive TTFB improvement")
	}

	t.Logf("TTFB improvement: %v", ttfbImprovement)
}

func TestStreamingHandler_EmptyItems(t *testing.T) {
	config := StreamingConfig{
		Enabled:   true,
		ChunkSize: 10,
	}

	handler := NewStreamingHandler(config)

	ctx := context.Background()
	chunks := handler.StreamItems(ctx, []interface{}{})

	receivedChunks := 0
	for chunk := range chunks {
		receivedChunks++

		if chunk.Error != nil {
			t.Fatalf("Unexpected error: %v", chunk.Error)
		}

		if len(chunk.Items) != 0 {
			t.Errorf("Expected empty chunk, got %d items", len(chunk.Items))
		}
	}

	if receivedChunks > 1 {
		t.Errorf("Expected at most 1 chunk for empty items, got %d", receivedChunks)
	}
}

func BenchmarkStreaming_Small(b *testing.B) {
	config := StreamingConfig{
		Enabled:      true,
		ChunkSize:    10,
		ThrottleRate: 0, // No throttle for benchmark
	}

	handler := NewStreamingHandler(config)

	items := make([]interface{}, 100)
	for i := range 100 {
		items[i] = i
	}

	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		chunks := handler.StreamItems(ctx, items)
		for range chunks {
			// Consume chunks
		}
	}
}

func BenchmarkStreaming_Large(b *testing.B) {
	config := StreamingConfig{
		Enabled:      true,
		ChunkSize:    100,
		ThrottleRate: 0,
		BufferSize:   200,
	}

	handler := NewStreamingHandler(config)

	items := make([]interface{}, 10000)
	for i := range 10000 {
		items[i] = i
	}

	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		chunks := handler.StreamItems(ctx, items)
		for range chunks {
			// Consume chunks
		}
	}
}

func BenchmarkStreaming_WithTransform(b *testing.B) {
	config := StreamingConfig{
		Enabled:      true,
		ChunkSize:    50,
		ThrottleRate: 0,
	}

	handler := NewStreamingHandler(config)

	items := make([]interface{}, 1000)
	for i := range 1000 {
		items[i] = i
	}

	transform := func(chunk []interface{}) ([]interface{}, error) {
		transformed := make([]interface{}, len(chunk))
		for i, item := range chunk {
			if num, ok := item.(int); ok {
				transformed[i] = num * 2
			}
		}
		return transformed, nil
	}

	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		chunks := handler.StreamWithTransform(ctx, items, transform)
		for range chunks {
			// Consume chunks
		}
	}
}
