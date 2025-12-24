package mcp

import (
	"context"
	"sync"
	"time"
)

// StreamingHandler handles chunked streaming responses for large datasets.
type StreamingHandler struct {
	config StreamingConfig
	stats  StreamingStats
	mu     sync.RWMutex
}

// StreamingConfig holds configuration for streaming responses.
type StreamingConfig struct {
	Enabled      bool
	ChunkSize    int           // Items per chunk
	ThrottleRate time.Duration // Delay between chunks
	BufferSize   int           // Channel buffer size
	MaxChunks    int           // Maximum chunks (0 = unlimited)
}

// StreamingStats tracks streaming performance metrics.
type StreamingStats struct {
	TotalStreams    int64
	TotalChunks     int64
	TotalItems      int64
	AvgChunkTime    time.Duration
	MemoryPeakBytes int64
}

// StreamChunk represents a chunk of data in the stream.
type StreamChunk struct {
	ChunkID   int
	Items     []interface{}
	HasMore   bool
	TotalSize int
	Error     error
}

// NewStreamingHandler creates a new streaming handler.
func NewStreamingHandler(config StreamingConfig) *StreamingHandler {
	// Set defaults
	if config.ChunkSize == 0 {
		config.ChunkSize = 10
	}
	if config.ThrottleRate == 0 {
		config.ThrottleRate = 50 * time.Millisecond
	}
	if config.BufferSize == 0 {
		config.BufferSize = 100
	}

	return &StreamingHandler{
		config: config,
		stats:  StreamingStats{},
	}
}

// StreamItems streams items in chunks through a channel.
func (h *StreamingHandler) StreamItems(ctx context.Context, items []interface{}) <-chan StreamChunk {
	chunks := make(chan StreamChunk, h.config.BufferSize)

	if !h.config.Enabled {
		// Send all items in one chunk if streaming disabled
		go func() {
			defer close(chunks)
			chunks <- StreamChunk{
				ChunkID:   0,
				Items:     items,
				HasMore:   false,
				TotalSize: len(items),
			}
		}()
		return chunks
	}

	// Stream chunks asynchronously
	go h.streamWorker(ctx, items, chunks)

	return chunks
}

// streamWorker processes items and sends chunks.
func (h *StreamingHandler) streamWorker(ctx context.Context, items []interface{}, chunks chan<- StreamChunk) {
	defer close(chunks)

	startTime := time.Now()
	totalSize := len(items)
	chunkID := 0

	h.updateStats(func(s *StreamingStats) {
		s.TotalStreams++
	})

	for i := 0; i < len(items); i += h.config.ChunkSize {
		// Check context cancellation
		select {
		case <-ctx.Done():
			chunks <- StreamChunk{
				ChunkID: chunkID,
				Error:   ctx.Err(),
			}
			return
		default:
		}

		// Check max chunks limit
		if h.config.MaxChunks > 0 && chunkID >= h.config.MaxChunks {
			break
		}

		// Calculate chunk bounds
		end := i + h.config.ChunkSize
		if end > len(items) {
			end = len(items)
		}

		// Create chunk
		chunk := StreamChunk{
			ChunkID:   chunkID,
			Items:     items[i:end],
			HasMore:   end < len(items),
			TotalSize: totalSize,
		}

		chunkStartTime := time.Now()

		// Send chunk
		select {
		case chunks <- chunk:
			chunkDuration := time.Since(chunkStartTime)

			// Update stats
			h.updateStats(func(s *StreamingStats) {
				s.TotalChunks++
				s.TotalItems += int64(len(chunk.Items))

				// Update average chunk time (exponential moving average)
				alpha := 0.1
				if s.AvgChunkTime == 0 {
					s.AvgChunkTime = chunkDuration
				} else {
					s.AvgChunkTime = time.Duration(float64(s.AvgChunkTime)*(1-alpha) + float64(chunkDuration)*alpha)
				}
			})

		case <-ctx.Done():
			chunks <- StreamChunk{
				ChunkID: chunkID,
				Error:   ctx.Err(),
			}
			return
		}

		chunkID++

		// Throttle between chunks
		if chunk.HasMore && h.config.ThrottleRate > 0 {
			select {
			case <-time.After(h.config.ThrottleRate):
			case <-ctx.Done():
				return
			}
		}
	}

	// Log stream completion
	duration := time.Since(startTime)
	_ = duration // Avoid unused variable warning
}

// CollectChunks collects all chunks from a stream into a single slice.
// Useful for testing or when you need all items at once.
func (h *StreamingHandler) CollectChunks(ctx context.Context, chunks <-chan StreamChunk) ([]interface{}, error) {
	var allItems []interface{}

	for chunk := range chunks {
		if chunk.Error != nil {
			return allItems, chunk.Error
		}

		allItems = append(allItems, chunk.Items...)
	}

	return allItems, nil
}

// StreamWithTransform streams items with a transformation function applied to each chunk.
func (h *StreamingHandler) StreamWithTransform(
	ctx context.Context,
	items []interface{},
	transform func([]interface{}) ([]interface{}, error),
) <-chan StreamChunk {
	chunks := make(chan StreamChunk, h.config.BufferSize)

	go func() {
		defer close(chunks)

		sourceChunks := h.StreamItems(ctx, items)
		chunkID := 0

		for chunk := range sourceChunks {
			if chunk.Error != nil {
				chunks <- chunk
				return
			}

			// Apply transformation
			transformed, err := transform(chunk.Items)
			if err != nil {
				chunks <- StreamChunk{
					ChunkID: chunkID,
					Error:   err,
				}
				return
			}

			// Send transformed chunk
			chunks <- StreamChunk{
				ChunkID:   chunkID,
				Items:     transformed,
				HasMore:   chunk.HasMore,
				TotalSize: chunk.TotalSize,
			}

			chunkID++
		}
	}()

	return chunks
}

// StreamWithFilter streams items with a filter applied.
func (h *StreamingHandler) StreamWithFilter(
	ctx context.Context,
	items []interface{},
	filter func(interface{}) bool,
) <-chan StreamChunk {
	return h.StreamWithTransform(ctx, items, func(items []interface{}) ([]interface{}, error) {
		filtered := make([]interface{}, 0, len(items))
		for _, item := range items {
			if filter(item) {
				filtered = append(filtered, item)
			}
		}
		return filtered, nil
	})
}

// GetStats returns current streaming statistics.
func (h *StreamingHandler) GetStats() StreamingStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.stats
}

// updateStats updates statistics with a mutation function.
func (h *StreamingHandler) updateStats(fn func(*StreamingStats)) {
	h.mu.Lock()
	defer h.mu.Unlock()

	fn(&h.stats)
}

// EstimateMemorySavings estimates memory savings from streaming vs loading all at once.
func (h *StreamingHandler) EstimateMemorySavings(totalItems int, itemSize int) int64 {
	if !h.config.Enabled {
		return 0
	}

	// Memory used without streaming: all items in memory
	totalMemory := int64(totalItems * itemSize)

	// Memory used with streaming: only one chunk in memory
	streamMemory := int64(h.config.ChunkSize * itemSize)

	return totalMemory - streamMemory
}

// CalculateTTFB calculates estimated Time To First Byte improvement.
func (h *StreamingHandler) CalculateTTFB(totalItems int) time.Duration {
	if !h.config.Enabled || totalItems == 0 {
		return 0
	}

	// Assume baseline: loading all items takes 100ms per 100 items
	baselinePerItem := 1 * time.Millisecond
	baselineTTFB := time.Duration(totalItems) * baselinePerItem

	// With streaming: TTFB is time to produce first chunk
	streamingTTFB := time.Duration(h.config.ChunkSize) * baselinePerItem

	return baselineTTFB - streamingTTFB
}

// ResetStats resets streaming statistics.
func (h *StreamingHandler) ResetStats() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stats = StreamingStats{}
}
