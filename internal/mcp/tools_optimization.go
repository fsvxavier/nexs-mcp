package mcp

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// DeduplicateMemoriesInput defines input for deduplicate_memories tool.
type DeduplicateMemoriesInput struct {
	MergeStrategy string `json:"merge_strategy,omitempty"` // keep_first, keep_last, keep_longest, combine
	DryRun        bool   `json:"dry_run,omitempty"`        // Preview without applying
}

// DeduplicateMemoriesOutput defines output for deduplicate_memories tool.
type DeduplicateMemoriesOutput struct {
	OriginalCount     int                      `json:"original_count"`
	DeduplicatedCount int                      `json:"deduplicated_count"`
	DuplicatesRemoved int                      `json:"duplicates_removed"`
	BytesSaved        int                      `json:"bytes_saved"`
	MergeStrategy     string                   `json:"merge_strategy"`
	DryRun            bool                     `json:"dry_run"`
	DuplicateGroups   int                      `json:"duplicate_groups"`
	Groups            []map[string]interface{} `json:"groups"`
	Stats             map[string]interface{}   `json:"stats"`
}

// handleDeduplicateMemories finds and merges duplicate memories.
func (s *MCPServer) handleDeduplicateMemories(ctx context.Context, req *sdk.CallToolRequest, input DeduplicateMemoriesInput) (*sdk.CallToolResult, DeduplicateMemoriesOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "deduplicate_memories",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	// List all memories (use empty filter to get all elements)
	memories, err := s.repo.List(domain.ElementFilter{})
	if err != nil {
		handlerErr = fmt.Errorf("failed to list memories: %w", err)
		return nil, DeduplicateMemoriesOutput{}, handlerErr
	}

	// Convert to deduplication items
	items := make([]application.DeduplicateItem, len(memories))
	for i, mem := range memories {
		metadata := mem.GetMetadata()
		// Convert Tags to metadata map
		metaMap := make(map[string]interface{})
		if len(metadata.Tags) > 0 {
			metaMap["tags"] = metadata.Tags
		}
		items[i] = application.DeduplicateItem{
			ID:        metadata.ID,
			Content:   metadata.Description,
			CreatedAt: metadata.CreatedAt,
			Metadata:  metaMap,
		}
	}

	// Set merge strategy
	strategy := application.MergeKeepFirst
	switch input.MergeStrategy {
	case "keep_last":
		strategy = application.MergeKeepLast
	case "keep_longest":
		strategy = application.MergeKeepLongest
	case "combine":
		strategy = application.MergeCombine
	}

	// Configure deduplication
	deduplicationService := application.NewSemanticDeduplicationService(application.DeduplicationConfig{
		Enabled:             true,
		SimilarityThreshold: 0.92,
		MergeStrategy:       strategy,
		PreserveMetadata:    true,
		BatchSize:           100,
	})

	// Deduplicate
	deduplicated, result, err := deduplicationService.DeduplicateItems(ctx, items)
	if err != nil {
		handlerErr = fmt.Errorf("deduplication failed: %w", err)
		return nil, DeduplicateMemoriesOutput{}, handlerErr
	}

	// Apply changes if not dry run
	if !input.DryRun && result.DuplicatesRemoved > 0 {
		// Delete duplicates
		for _, group := range result.Groups {
			// Keep first, delete rest
			for i := 1; i < len(group.Items); i++ {
				if err := s.repo.Delete(group.Items[i].ID); err != nil {
					handlerErr = fmt.Errorf("failed to delete duplicate %s: %w", group.Items[i].ID, err)
					return nil, DeduplicateMemoriesOutput{}, handlerErr
				}
			}
		}

		// Update merged items if using combine strategy
		if strategy == application.MergeCombine {
			for _, item := range deduplicated {
				elem, err := s.repo.GetByID(item.ID)
				if err != nil {
					continue
				}
				// Update element content through metadata
				metadata := elem.GetMetadata()
				metadata.Description = item.Content
				if err := s.repo.Update(elem); err != nil {
					handlerErr = fmt.Errorf("failed to update merged item %s: %w", item.ID, err)
					return nil, DeduplicateMemoriesOutput{}, handlerErr
				}
			}
		}
	}

	// Add group details
	groups := make([]map[string]interface{}, len(result.Groups))
	for i, group := range result.Groups {
		ids := make([]string, len(group.Items))
		for j, item := range group.Items {
			ids[j] = item.ID
		}
		groups[i] = map[string]interface{}{
			"similarity": fmt.Sprintf("%.1f%%", group.Similarity*100),
			"items":      ids,
			"count":      len(group.Items),
		}
	}

	// Get stats
	stats := deduplicationService.GetStats()

	output := DeduplicateMemoriesOutput{
		OriginalCount:     result.OriginalCount,
		DeduplicatedCount: result.DeduplicatedCount,
		DuplicatesRemoved: result.DuplicatesRemoved,
		BytesSaved:        result.BytesSaved,
		MergeStrategy:     input.MergeStrategy,
		DryRun:            input.DryRun,
		DuplicateGroups:   len(result.Groups),
		Groups:            groups,
		Stats: map[string]interface{}{
			"total_processed":    stats.TotalProcessed,
			"duplicates_found":   stats.DuplicatesFound,
			"duplicates_removed": stats.DuplicatesRemoved,
			"bytes_saved":        stats.BytesSaved,
			"avg_similarity":     fmt.Sprintf("%.1f%%", stats.AvgSimilarity*100),
		},
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "deduplicate_memories", output)

	return nil, output, nil
}

// OptimizeContextInput represents input for context optimization.
type OptimizeContextInput struct {
	Items       []map[string]interface{} `json:"items"`
	MaxTokens   int                      `json:"max_tokens"`
	Strategy    string                   `json:"strategy"`
	Truncation  string                   `json:"truncation"`
	PreserveKey bool                     `json:"preserve_key_items"`
}

// OptimizeContextOutput represents output from context optimization.
type OptimizeContextOutput struct {
	OriginalCount   int                      `json:"original_count"`
	OptimizedCount  int                      `json:"optimized_count"`
	ItemsRemoved    int                      `json:"items_removed"`
	TokensSaved     int                      `json:"tokens_saved"`
	Strategy        string                   `json:"strategy"`
	Truncation      string                   `json:"truncation"`
	PreserveKey     bool                     `json:"preserve_key"`
	OriginalTokens  int                      `json:"original_tokens"`
	OptimizedTokens int                      `json:"optimized_tokens"`
	OptimizedItems  []map[string]interface{} `json:"optimized_items"`
	Stats           map[string]interface{}   `json:"stats"`
}

// handleOptimizeContext optimizes context window.
func (s *MCPServer) handleOptimizeContext(ctx context.Context, req *sdk.CallToolRequest, input OptimizeContextInput) (*sdk.CallToolResult, OptimizeContextOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "optimize_context",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	// Convert to context items
	items := make([]application.ContextItem, len(input.Items))
	for i, item := range input.Items {
		id, _ := item["id"].(string)
		content, _ := item["content"].(string)
		createdAt := time.Now()
		if ts, ok := item["created_at"].(string); ok {
			createdAt, _ = time.Parse(time.RFC3339, ts)
		}
		lastAccess := time.Now()
		if ts, ok := item["last_access"].(string); ok {
			lastAccess, _ = time.Parse(time.RFC3339, ts)
		}
		relevance := 0.0
		if rel, ok := item["relevance"].(float64); ok {
			relevance = rel
		}
		importance := 5
		if imp, ok := item["importance"].(float64); ok {
			importance = int(imp)
		}

		items[i] = application.ContextItem{
			ID:         id,
			Content:    content,
			CreatedAt:  createdAt,
			LastAccess: lastAccess,
			Relevance:  relevance,
			Importance: importance,
		}
	}

	// Configure manager
	strategy := application.PriorityRecency
	switch input.Strategy {
	case "relevance":
		strategy = application.PriorityRelevance
	case "importance":
		strategy = application.PriorityImportance
	case "hybrid":
		strategy = application.PriorityHybrid
	}

	truncation := application.TruncationTail
	switch input.Truncation {
	case "head":
		truncation = application.TruncationHead
	case "middle":
		truncation = application.TruncationMiddle
	}

	contextWindowManager := application.NewContextWindowManager(application.ContextWindowConfig{
		MaxTokens:          input.MaxTokens,
		PriorityStrategy:   strategy,
		TruncationMethod:   truncation,
		PreserveRecent:     5,
		RelevanceThreshold: 0.3,
	})

	// Optimize
	optimized, result, err := contextWindowManager.OptimizeContext(ctx, items)
	if err != nil {
		handlerErr = fmt.Errorf("optimization failed: %w", err)
		return nil, OptimizeContextOutput{}, handlerErr
	}

	// Add optimized items
	optimizedItems := make([]map[string]interface{}, len(optimized))
	for i, item := range optimized {
		optimizedItems[i] = map[string]interface{}{
			"id":          item.ID,
			"content":     item.Content,
			"created_at":  item.CreatedAt.Format(time.RFC3339),
			"last_access": item.LastAccess.Format(time.RFC3339),
			"relevance":   item.Relevance,
			"importance":  item.Importance,
		}
	}

	// Get stats
	stats := contextWindowManager.GetStats()

	output := OptimizeContextOutput{
		OriginalCount:   len(items),
		OptimizedCount:  len(optimized),
		ItemsRemoved:    result.ItemsRemoved,
		TokensSaved:     result.OriginalTokenCount - result.OptimizedTokenCount,
		Strategy:        input.Strategy,
		Truncation:      input.Truncation,
		PreserveKey:     input.PreserveKey,
		OriginalTokens:  result.OriginalTokenCount,
		OptimizedTokens: result.OptimizedTokenCount,
		OptimizedItems:  optimizedItems,
		Stats: map[string]interface{}{
			"total_optimizations": stats.TotalOptimizations,
			"items_removed":       result.ItemsRemoved,
			"tokens_saved":        stats.TokensSaved,
			"relevance_gain":      fmt.Sprintf("%.2f", stats.AvgRelevanceGain),
			"strategy_used":       input.Strategy,
			"truncation_used":     input.Truncation,
			"preserve_key_used":   input.PreserveKey,
		},
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "optimize_context", output)

	return nil, output, nil
}

// GetOptimizationStatsInput represents input for getting optimization stats.
type GetOptimizationStatsInput struct {
	Detailed bool `json:"detailed"`
}

// GetOptimizationStatsOutput represents optimization statistics.
type GetOptimizationStatsOutput struct {
	Compression          map[string]interface{} `json:"compression,omitempty"`
	Streaming            map[string]interface{} `json:"streaming,omitempty"`
	Summarization        map[string]interface{} `json:"summarization,omitempty"`
	Deduplication        map[string]interface{} `json:"deduplication,omitempty"`
	ContextWindow        map[string]interface{} `json:"context_window,omitempty"`
	PromptCompression    map[string]interface{} `json:"prompt_compression,omitempty"`
	AdaptiveCache        map[string]interface{} `json:"adaptive_cache,omitempty"`
	TokenMetrics         map[string]interface{} `json:"token_metrics,omitempty"`
	TotalBytesSaved      int64                  `json:"total_bytes_saved"`
	TotalMBSaved         string                 `json:"total_mb_saved"`
	TotalTokensSaved     int64                  `json:"total_tokens_saved"`
	EstimatedCostSavings string                 `json:"estimated_cost_savings"`
}

// handleGetOptimizationStats returns comprehensive optimization statistics.
func (s *MCPServer) handleGetOptimizationStats(ctx context.Context, req *sdk.CallToolRequest, input GetOptimizationStatsInput) (*sdk.CallToolResult, GetOptimizationStatsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_optimization_stats",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	output := GetOptimizationStatsOutput{}
	totalBytesSaved := int64(0)

	// Compression stats
	if s.compressor != nil {
		compStats := s.compressor.GetStats()
		output.Compression = map[string]interface{}{
			"enabled":               s.cfg.Compression.Enabled,
			"total_requests":        compStats.TotalRequests,
			"compressed_requests":   compStats.CompressedRequests,
			"bytes_saved":           compStats.BytesSaved,
			"avg_compression_ratio": fmt.Sprintf("%.1f%%", compStats.AvgCompressionRatio*100),
			"algorithm":             s.cfg.Compression.Algorithm,
		}
		totalBytesSaved += compStats.BytesSaved
	}

	// Streaming stats
	if s.streamingHandler != nil {
		streamStats := s.streamingHandler.GetStats()
		output.Streaming = map[string]interface{}{
			"enabled":        s.cfg.Streaming.Enabled,
			"total_streams":  streamStats.TotalStreams,
			"total_chunks":   streamStats.TotalChunks,
			"total_items":    streamStats.TotalItems,
			"avg_chunk_time": streamStats.AvgChunkTime.String(),
			"memory_peak":    streamStats.MemoryPeakBytes,
			"chunk_size":     s.cfg.Streaming.ChunkSize,
		}
	}

	// Summarization stats
	if s.summarizationService != nil {
		sumStats := s.summarizationService.GetStats()
		output.Summarization = map[string]interface{}{
			"enabled":               s.cfg.Summarization.Enabled,
			"total_summarized":      sumStats.TotalSummarized,
			"bytes_saved":           sumStats.BytesSaved,
			"avg_compression_ratio": fmt.Sprintf("%.1f%%", sumStats.AvgCompressionRatio*100),
			"quality_score":         fmt.Sprintf("%.2f", sumStats.QualityScore),
		}
		totalBytesSaved += sumStats.BytesSaved
	}

	// Deduplication stats
	if s.deduplicationService != nil {
		dedupStats := s.deduplicationService.GetStats()
		output.Deduplication = map[string]interface{}{
			"total_processed":    dedupStats.TotalProcessed,
			"duplicates_found":   dedupStats.DuplicatesFound,
			"duplicates_removed": dedupStats.DuplicatesRemoved,
			"bytes_saved":        dedupStats.BytesSaved,
			"avg_similarity":     fmt.Sprintf("%.1f%%", dedupStats.AvgSimilarity*100),
		}
		totalBytesSaved += dedupStats.BytesSaved
	}

	// Context window stats
	if s.contextWindowManager != nil {
		ctxStats := s.contextWindowManager.GetStats()
		output.ContextWindow = map[string]interface{}{
			"total_optimizations": ctxStats.TotalOptimizations,
			"overflows_prevented": ctxStats.OverflowsPrevented,
			"tokens_saved":        ctxStats.TokensSaved,
			"avg_relevance_gain":  fmt.Sprintf("%.2f", ctxStats.AvgRelevanceGain),
		}
	}

	// Prompt compression stats
	if s.promptCompressor != nil {
		promptStats := s.promptCompressor.GetStats()
		output.PromptCompression = map[string]interface{}{
			"enabled":               s.cfg.PromptCompression.Enabled,
			"total_compressed":      promptStats.TotalCompressed,
			"bytes_saved":           promptStats.BytesSaved,
			"avg_compression_ratio": fmt.Sprintf("%.1f%%", promptStats.AvgCompressionRatio*100),
		}
		totalBytesSaved += promptStats.BytesSaved
	}

	// Adaptive cache stats
	if s.adaptiveCache != nil {
		cacheStats := s.adaptiveCache.GetStats()
		hitRate := s.adaptiveCache.GetHitRate()
		output.AdaptiveCache = map[string]interface{}{
			"enabled":          s.cfg.AdaptiveCache.Enabled,
			"total_hits":       cacheStats.TotalHits,
			"total_misses":     cacheStats.TotalMisses,
			"hit_rate":         fmt.Sprintf("%.1f%%", hitRate*100),
			"total_entries":    cacheStats.TotalEntries,
			"total_evictions":  cacheStats.TotalEvictions,
			"bytes_cached":     cacheStats.BytesCached,
			"avg_ttl":          cacheStats.AvgTTL.String(),
			"ttl_adjustments":  cacheStats.TTLAdjustments,
			"avg_access_count": fmt.Sprintf("%.1f", cacheStats.AvgAccessCount),
		}
	}

	// Token metrics (real production data)
	if s.tokenMetrics != nil {
		tokenStats := s.tokenMetrics.GetStats()
		output.TokenMetrics = map[string]interface{}{
			"total_original_tokens":  tokenStats.TotalOriginalTokens,
			"total_optimized_tokens": tokenStats.TotalOptimizedTokens,
			"total_tokens_saved":     tokenStats.TotalTokensSaved,
			"avg_compression_ratio":  fmt.Sprintf("%.1f%%", tokenStats.AvgCompressionRatio*100),
			"optimization_count":     tokenStats.OptimizationCount,
			"tokens_saved_by_type":   tokenStats.TokensSavedByType,
			"tokens_saved_by_tool":   tokenStats.TokensSavedByTool,
			"last_optimization":      tokenStats.LastOptimizationTime.Format("2006-01-02 15:04:05"),
		}
		output.TotalTokensSaved = tokenStats.TotalTokensSaved

		// Estimate cost savings (assuming ~$0.01 per 1000 tokens for input, $0.03 per 1000 tokens for output)
		// Using average of $0.02 per 1000 tokens
		costSavings := float64(tokenStats.TotalTokensSaved) / 1000.0 * 0.02
		output.EstimatedCostSavings = fmt.Sprintf("$%.2f", costSavings)
	}

	output.TotalBytesSaved = totalBytesSaved
	output.TotalMBSaved = fmt.Sprintf("%.2f MB", float64(totalBytesSaved)/(1024*1024))

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "get_optimization_stats", output)

	return nil, output, nil
}
