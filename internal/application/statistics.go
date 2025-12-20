package application

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// ToolCallMetric represents a single tool call metric
type ToolCallMetric struct {
	ToolName      string        `json:"tool_name"`
	Timestamp     time.Time     `json:"timestamp"`
	Duration      time.Duration `json:"duration_ms"`
	Success       bool          `json:"success"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	User          string        `json:"user,omitempty"`
	RequestParams interface{}   `json:"request_params,omitempty"`
}

// UsageStatistics represents aggregated statistics
type UsageStatistics struct {
	TotalOperations    int                `json:"total_operations"`
	SuccessfulOps      int                `json:"successful_ops"`
	FailedOps          int                `json:"failed_ops"`
	SuccessRate        float64            `json:"success_rate"`
	OperationsByTool   map[string]int     `json:"operations_by_tool"`
	ErrorsByTool       map[string]int     `json:"errors_by_tool"`
	AvgDurationByTool  map[string]float64 `json:"avg_duration_by_tool_ms"`
	MostUsedTools      []ToolUsageStat    `json:"most_used_tools"`
	SlowestOperations  []ToolCallMetric   `json:"slowest_operations"`
	RecentErrors       []ToolCallMetric   `json:"recent_errors"`
	ActiveUsers        []string           `json:"active_users"`
	OperationsByPeriod map[string]int     `json:"operations_by_period"`
	Period             string             `json:"period"`
	StartTime          time.Time          `json:"start_time"`
	EndTime            time.Time          `json:"end_time"`
}

// ToolUsageStat represents usage statistics for a specific tool
type ToolUsageStat struct {
	ToolName    string  `json:"tool_name"`
	Count       int     `json:"count"`
	SuccessRate float64 `json:"success_rate"`
	AvgDuration float64 `json:"avg_duration_ms"`
}

// MetricsCollector collects and aggregates metrics
type MetricsCollector struct {
	mu           sync.RWMutex
	metrics      []ToolCallMetric
	maxMetrics   int
	storageDir   string
	autoSave     bool
	saveInterval time.Duration
	lastSaveTime time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(storageDir string) *MetricsCollector {
	mc := &MetricsCollector{
		metrics:      make([]ToolCallMetric, 0, 10000),
		maxMetrics:   10000,
		storageDir:   storageDir,
		autoSave:     true,
		saveInterval: 5 * time.Minute,
		lastSaveTime: time.Now(),
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		// Log error but continue
		fmt.Fprintf(os.Stderr, "Warning: failed to create metrics directory: %v\n", err)
	}

	// Load existing metrics
	if err := mc.loadMetrics(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load existing metrics: %v\n", err)
	}

	return mc
}

// RecordToolCall records a tool call metric
func (mc *MetricsCollector) RecordToolCall(metric ToolCallMetric) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Add metric
	mc.metrics = append(mc.metrics, metric)

	// Trim if exceeds max
	if len(mc.metrics) > mc.maxMetrics {
		// Keep most recent metrics
		mc.metrics = mc.metrics[len(mc.metrics)-mc.maxMetrics:]
	}

	// Auto-save if interval elapsed
	if mc.autoSave && time.Since(mc.lastSaveTime) > mc.saveInterval {
		go func() {
			if err := mc.SaveMetrics(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to auto-save metrics: %v\n", err)
			}
		}()
	}
}

// GetStatistics returns aggregated statistics for a given period
func (mc *MetricsCollector) GetStatistics(period string) (*UsageStatistics, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	startTime, endTime := mc.calculatePeriod(period)

	// Filter metrics by period
	filteredMetrics := make([]ToolCallMetric, 0)
	for _, m := range mc.metrics {
		if (m.Timestamp.Equal(startTime) || m.Timestamp.After(startTime)) && (m.Timestamp.Equal(endTime) || m.Timestamp.Before(endTime)) {
			filteredMetrics = append(filteredMetrics, m)
		}
	}

	if len(filteredMetrics) == 0 {
		return &UsageStatistics{
			Period:    period,
			StartTime: startTime,
			EndTime:   endTime,
		}, nil
	}

	stats := &UsageStatistics{
		OperationsByTool:   make(map[string]int),
		ErrorsByTool:       make(map[string]int),
		AvgDurationByTool:  make(map[string]float64),
		OperationsByPeriod: make(map[string]int),
		Period:             period,
		StartTime:          startTime,
		EndTime:            endTime,
		ActiveUsers:        make([]string, 0),
	}

	// Aggregate statistics
	durationSumByTool := make(map[string]float64)
	countByTool := make(map[string]int)
	userSet := make(map[string]bool)
	slowestOps := make([]ToolCallMetric, 0)
	recentErrors := make([]ToolCallMetric, 0)

	for _, m := range filteredMetrics {
		stats.TotalOperations++

		if m.Success {
			stats.SuccessfulOps++
		} else {
			stats.FailedOps++
			stats.ErrorsByTool[m.ToolName]++
			if len(recentErrors) < 10 {
				recentErrors = append(recentErrors, m)
			}
		}

		stats.OperationsByTool[m.ToolName]++
		durationSumByTool[m.ToolName] += float64(m.Duration.Milliseconds())
		countByTool[m.ToolName]++

		if m.User != "" {
			userSet[m.User] = true
		}

		// Track slowest operations
		slowestOps = append(slowestOps, m)

		// Group by period
		periodKey := m.Timestamp.Format("2006-01-02")
		stats.OperationsByPeriod[periodKey]++
	}

	// Calculate success rate
	if stats.TotalOperations > 0 {
		stats.SuccessRate = float64(stats.SuccessfulOps) / float64(stats.TotalOperations) * 100
	}

	// Calculate average durations
	for tool, sum := range durationSumByTool {
		if count := countByTool[tool]; count > 0 {
			stats.AvgDurationByTool[tool] = sum / float64(count)
		}
	}

	// Build most used tools list
	mostUsed := make([]ToolUsageStat, 0, len(stats.OperationsByTool))
	for tool, count := range stats.OperationsByTool {
		successCount := count - stats.ErrorsByTool[tool]
		successRate := 0.0
		if count > 0 {
			successRate = float64(successCount) / float64(count) * 100
		}

		mostUsed = append(mostUsed, ToolUsageStat{
			ToolName:    tool,
			Count:       count,
			SuccessRate: successRate,
			AvgDuration: stats.AvgDurationByTool[tool],
		})
	}

	// Sort by count descending
	sort.Slice(mostUsed, func(i, j int) bool {
		return mostUsed[i].Count > mostUsed[j].Count
	})

	// Take top 10
	if len(mostUsed) > 10 {
		mostUsed = mostUsed[:10]
	}
	stats.MostUsedTools = mostUsed

	// Sort slowest operations by duration descending
	sort.Slice(slowestOps, func(i, j int) bool {
		return slowestOps[i].Duration > slowestOps[j].Duration
	})

	// Take top 10 slowest
	if len(slowestOps) > 10 {
		slowestOps = slowestOps[:10]
	}
	stats.SlowestOperations = slowestOps

	// Sort recent errors by timestamp descending
	sort.Slice(recentErrors, func(i, j int) bool {
		return recentErrors[i].Timestamp.After(recentErrors[j].Timestamp)
	})
	stats.RecentErrors = recentErrors

	// Extract active users
	for user := range userSet {
		stats.ActiveUsers = append(stats.ActiveUsers, user)
	}
	sort.Strings(stats.ActiveUsers)

	return stats, nil
}

// calculatePeriod calculates start and end time for a given period string
func (mc *MetricsCollector) calculatePeriod(period string) (time.Time, time.Time) {
	now := time.Now()
	var startTime time.Time

	switch period {
	case "last_hour":
		startTime = now.Add(-1 * time.Hour)
	case "last_24h", "today":
		startTime = now.Add(-24 * time.Hour)
	case "last_7_days", "week":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "last_30_days", "month":
		startTime = now.Add(-30 * 24 * time.Hour)
	case "all", "":
		// All time - use earliest metric timestamp if available
		if len(mc.metrics) > 0 {
			startTime = mc.metrics[0].Timestamp
		} else {
			startTime = now.Add(-365 * 24 * time.Hour) // 1 year back as default
		}
	default:
		// Default to last 24 hours
		startTime = now.Add(-24 * time.Hour)
	}

	return startTime, now
}

// SaveMetrics persists metrics to disk
func (mc *MetricsCollector) SaveMetrics() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	filename := filepath.Join(mc.storageDir, "metrics.json")
	data, err := json.MarshalIndent(mc.metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	mc.lastSaveTime = time.Now()
	return nil
}

// loadMetrics loads persisted metrics from disk
func (mc *MetricsCollector) loadMetrics() error {
	filename := filepath.Join(mc.storageDir, "metrics.json")

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing metrics, not an error
		}
		return fmt.Errorf("failed to read metrics file: %w", err)
	}

	if err := json.Unmarshal(data, &mc.metrics); err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return nil
}

// ClearMetrics removes all collected metrics
func (mc *MetricsCollector) ClearMetrics() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics = make([]ToolCallMetric, 0, 10000)

	filename := filepath.Join(mc.storageDir, "metrics.json")
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metrics file: %w", err)
	}

	return nil
}

// GetMetricsCount returns the current number of stored metrics
func (mc *MetricsCollector) GetMetricsCount() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return len(mc.metrics)
}
