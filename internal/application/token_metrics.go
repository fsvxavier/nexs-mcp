package application

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// TokenMetrics tracks token usage and optimization savings.
type TokenMetrics struct {
	OriginalTokens   int64     `json:"original_tokens"`
	OptimizedTokens  int64     `json:"optimized_tokens"`
	TokensSaved      int64     `json:"tokens_saved"`
	CompressionRatio float64   `json:"compression_ratio"`
	OptimizationType string    `json:"optimization_type"` // "compression", "dedup", "summary", "context"
	ToolName         string    `json:"tool_name"`
	Timestamp        time.Time `json:"timestamp"`
}

// TokenOptimizationStats aggregates token optimization metrics.
type TokenOptimizationStats struct {
	TotalOriginalTokens  int64            `json:"total_original_tokens"`
	TotalOptimizedTokens int64            `json:"total_optimized_tokens"`
	TotalTokensSaved     int64            `json:"total_tokens_saved"`
	AvgCompressionRatio  float64          `json:"avg_compression_ratio"`
	TokensSavedByType    map[string]int64 `json:"tokens_saved_by_type"`
	TokensSavedByTool    map[string]int64 `json:"tokens_saved_by_tool"`
	OptimizationCount    int64            `json:"optimization_count"`
	LastOptimizationTime time.Time        `json:"last_optimization_time"`
}

// TokenMetricsCollector collects and aggregates token optimization metrics.
type TokenMetricsCollector struct {
	mu           sync.RWMutex
	metrics      []TokenMetrics
	maxMetrics   int
	storageDir   string
	autoSave     bool
	saveInterval time.Duration
	lastSaveTime time.Time

	// Atomic counters for thread-safe updates
	totalOriginal   int64
	totalOptimized  int64
	totalSaved      int64
	optimizationCnt int64
}

// NewTokenMetricsCollector creates a new token metrics collector.
func NewTokenMetricsCollector(storageDir string) *TokenMetricsCollector {
	tmc := &TokenMetricsCollector{
		metrics:      make([]TokenMetrics, 0, 10000),
		maxMetrics:   10000,
		storageDir:   storageDir,
		autoSave:     true,
		saveInterval: 5 * time.Minute,
		lastSaveTime: time.Now(),
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create token metrics directory: %v\n", err)
	}

	// Load existing metrics
	if err := tmc.loadMetrics(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load existing token metrics: %v\n", err)
	}

	return tmc
}

// RecordTokenOptimization records a token optimization event.
func (tmc *TokenMetricsCollector) RecordTokenOptimization(metric TokenMetrics) {
	tmc.mu.Lock()
	defer tmc.mu.Unlock()

	// Set timestamp if not already set
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	// Calculate tokens saved if not provided
	if metric.TokensSaved == 0 && metric.OriginalTokens > 0 {
		metric.TokensSaved = metric.OriginalTokens - metric.OptimizedTokens
	}

	// Calculate compression ratio if not provided
	if metric.CompressionRatio == 0 && metric.OriginalTokens > 0 {
		metric.CompressionRatio = float64(metric.OptimizedTokens) / float64(metric.OriginalTokens)
	}

	// Add metric
	tmc.metrics = append(tmc.metrics, metric)

	// Update atomic counters
	atomic.AddInt64(&tmc.totalOriginal, metric.OriginalTokens)
	atomic.AddInt64(&tmc.totalOptimized, metric.OptimizedTokens)
	atomic.AddInt64(&tmc.totalSaved, metric.TokensSaved)
	atomic.AddInt64(&tmc.optimizationCnt, 1)

	// Trim if exceeds max
	if len(tmc.metrics) > tmc.maxMetrics {
		tmc.metrics = tmc.metrics[len(tmc.metrics)-tmc.maxMetrics:]
	}

	// Auto-save if interval elapsed
	if tmc.autoSave && time.Since(tmc.lastSaveTime) > tmc.saveInterval {
		go func() {
			if err := tmc.SaveMetrics(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to auto-save token metrics: %v\n", err)
			}
		}()
	}
}

// GetStats returns aggregated token optimization statistics.
func (tmc *TokenMetricsCollector) GetStats() TokenOptimizationStats {
	tmc.mu.RLock()
	defer tmc.mu.RUnlock()

	stats := TokenOptimizationStats{
		TotalOriginalTokens:  atomic.LoadInt64(&tmc.totalOriginal),
		TotalOptimizedTokens: atomic.LoadInt64(&tmc.totalOptimized),
		TotalTokensSaved:     atomic.LoadInt64(&tmc.totalSaved),
		OptimizationCount:    atomic.LoadInt64(&tmc.optimizationCnt),
		TokensSavedByType:    make(map[string]int64),
		TokensSavedByTool:    make(map[string]int64),
	}

	// Calculate average compression ratio
	if stats.TotalOriginalTokens > 0 {
		stats.AvgCompressionRatio = float64(stats.TotalOptimizedTokens) / float64(stats.TotalOriginalTokens)
	}

	// Aggregate by type and tool
	for _, m := range tmc.metrics {
		if m.OptimizationType != "" {
			stats.TokensSavedByType[m.OptimizationType] += m.TokensSaved
		}
		if m.ToolName != "" {
			stats.TokensSavedByTool[m.ToolName] += m.TokensSaved
		}
		if m.Timestamp.After(stats.LastOptimizationTime) {
			stats.LastOptimizationTime = m.Timestamp
		}
	}

	return stats
}

// GetRecentMetrics returns the N most recent token optimization events.
func (tmc *TokenMetricsCollector) GetRecentMetrics(n int) []TokenMetrics {
	tmc.mu.RLock()
	defer tmc.mu.RUnlock()

	if n > len(tmc.metrics) {
		n = len(tmc.metrics)
	}

	// Return last N metrics
	start := len(tmc.metrics) - n
	if start < 0 {
		start = 0
	}

	result := make([]TokenMetrics, n)
	copy(result, tmc.metrics[start:])
	return result
}

// SaveMetrics saves metrics to disk.
func (tmc *TokenMetricsCollector) SaveMetrics() error {
	tmc.mu.Lock()
	defer tmc.mu.Unlock()

	metricsPath := filepath.Join(tmc.storageDir, "token_metrics.json")

	// Marshal metrics
	data, err := json.MarshalIndent(tmc.metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token metrics: %w", err)
	}

	// Write atomically using temp file
	tempPath := metricsPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write token metrics: %w", err)
	}

	if err := os.Rename(tempPath, metricsPath); err != nil {
		return fmt.Errorf("failed to rename token metrics file: %w", err)
	}

	tmc.lastSaveTime = time.Now()
	return nil
}

// loadMetrics loads metrics from disk.
func (tmc *TokenMetricsCollector) loadMetrics() error {
	metricsPath := filepath.Join(tmc.storageDir, "token_metrics.json")

	data, err := os.ReadFile(metricsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing metrics, that's ok
		}
		return fmt.Errorf("failed to read token metrics: %w", err)
	}

	var metrics []TokenMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("failed to unmarshal token metrics: %w", err)
	}

	tmc.metrics = metrics

	// Rebuild atomic counters
	var totalOrig, totalOpt, totalSaved, count int64
	for _, m := range metrics {
		totalOrig += m.OriginalTokens
		totalOpt += m.OptimizedTokens
		totalSaved += m.TokensSaved
		count++
	}

	atomic.StoreInt64(&tmc.totalOriginal, totalOrig)
	atomic.StoreInt64(&tmc.totalOptimized, totalOpt)
	atomic.StoreInt64(&tmc.totalSaved, totalSaved)
	atomic.StoreInt64(&tmc.optimizationCnt, count)

	return nil
}

// EstimateTokenCount estimates token count from text length.
// This is a rough estimate: ~4 characters per token on average.
func EstimateTokenCount(text string) int64 {
	return int64(len(text) / 4)
}

// EstimateTokenCountFromBytes estimates token count from byte size.
func EstimateTokenCountFromBytes(size int) int64 {
	return int64(size / 4)
}
