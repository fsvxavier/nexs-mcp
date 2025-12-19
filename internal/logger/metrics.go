package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// PerformanceMetrics tracks timing and latency for operations
type PerformanceMetrics struct {
	mu          sync.RWMutex
	metrics     []OperationMetric
	metricsPath string
}

// OperationMetric represents a single operation timing
type OperationMetric struct {
	Operation string    `json:"operation"`
	Duration  float64   `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
}

// PerformanceDashboard provides performance analysis
type PerformanceDashboard struct {
	TotalOperations int                `json:"total_operations"`
	AvgDuration     float64            `json:"avg_duration_ms"`
	P50Duration     float64            `json:"p50_duration_ms"`
	P95Duration     float64            `json:"p95_duration_ms"`
	P99Duration     float64            `json:"p99_duration_ms"`
	MaxDuration     float64            `json:"max_duration_ms"`
	MinDuration     float64            `json:"min_duration_ms"`
	SlowOperations  []SlowOperation    `json:"slow_operations"`
	FastOperations  []SlowOperation    `json:"fast_operations"`
	ByOperation     map[string]OpStats `json:"by_operation"`
	Period          string             `json:"period"`
}

// SlowOperation represents a slow operation entry
type SlowOperation struct {
	Operation string    `json:"operation"`
	Duration  float64   `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
}

// OpStats represents statistics for a specific operation
type OpStats struct {
	Count       int     `json:"count"`
	AvgDuration float64 `json:"avg_duration_ms"`
	MaxDuration float64 `json:"max_duration_ms"`
	MinDuration float64 `json:"min_duration_ms"`
}

// NewPerformanceMetrics creates a new performance metrics tracker
func NewPerformanceMetrics(dataDir string) *PerformanceMetrics {
	metricsPath := filepath.Join(dataDir, "performance_metrics.json")

	pm := &PerformanceMetrics{
		metrics:     make([]OperationMetric, 0, 10000),
		metricsPath: metricsPath,
	}

	// Load existing metrics
	pm.loadMetrics()

	return pm
}

// RecordOperation records an operation with its duration in milliseconds
func (pm *PerformanceMetrics) RecordOperation(operation string, durationMs float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metric := OperationMetric{
		Operation: operation,
		Duration:  durationMs,
		Timestamp: time.Now(),
	}

	pm.metrics = append(pm.metrics, metric)

	// Keep only last 10000 metrics (circular buffer)
	if len(pm.metrics) > 10000 {
		pm.metrics = pm.metrics[len(pm.metrics)-10000:]
	}
}

// GetDashboard returns a performance dashboard for the specified period
func (pm *PerformanceMetrics) GetDashboard(period string) *PerformanceDashboard {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Calculate period cutoff
	cutoff := pm.calculatePeriod(period)

	// Filter metrics by period
	filtered := make([]OperationMetric, 0)
	for _, m := range pm.metrics {
		if m.Timestamp.After(cutoff) || m.Timestamp.Equal(cutoff) {
			filtered = append(filtered, m)
		}
	}

	if len(filtered) == 0 {
		return &PerformanceDashboard{
			Period:      period,
			ByOperation: make(map[string]OpStats),
		}
	}

	// Calculate statistics
	dashboard := &PerformanceDashboard{
		TotalOperations: len(filtered),
		Period:          period,
		ByOperation:     make(map[string]OpStats),
	}

	// Extract durations for percentile calculation
	durations := make([]float64, len(filtered))
	totalDuration := 0.0
	minDur := filtered[0].Duration
	maxDur := filtered[0].Duration

	for i, m := range filtered {
		durations[i] = m.Duration
		totalDuration += m.Duration

		if m.Duration < minDur {
			minDur = m.Duration
		}
		if m.Duration > maxDur {
			maxDur = m.Duration
		}
	}

	dashboard.AvgDuration = totalDuration / float64(len(filtered))
	dashboard.MinDuration = minDur
	dashboard.MaxDuration = maxDur

	// Calculate percentiles
	sort.Float64s(durations)
	dashboard.P50Duration = pm.percentile(durations, 50)
	dashboard.P95Duration = pm.percentile(durations, 95)
	dashboard.P99Duration = pm.percentile(durations, 99)

	// Identify slow and fast operations
	slowOps := make([]SlowOperation, 0)
	fastOps := make([]SlowOperation, 0)

	for _, m := range filtered {
		if m.Duration > dashboard.P95Duration {
			slowOps = append(slowOps, SlowOperation{
				Operation: m.Operation,
				Duration:  m.Duration,
				Timestamp: m.Timestamp,
			})
		} else if m.Duration < dashboard.P50Duration {
			fastOps = append(fastOps, SlowOperation{
				Operation: m.Operation,
				Duration:  m.Duration,
				Timestamp: m.Timestamp,
			})
		}
	}

	// Sort slow operations by duration (descending)
	sort.Slice(slowOps, func(i, j int) bool {
		return slowOps[i].Duration > slowOps[j].Duration
	})

	// Sort fast operations by duration (ascending)
	sort.Slice(fastOps, func(i, j int) bool {
		return fastOps[i].Duration < fastOps[j].Duration
	})

	// Keep top 10 slow and fast
	if len(slowOps) > 10 {
		slowOps = slowOps[:10]
	}
	if len(fastOps) > 10 {
		fastOps = fastOps[:10]
	}

	dashboard.SlowOperations = slowOps
	dashboard.FastOperations = fastOps

	// Calculate per-operation stats
	opStats := make(map[string]*OpStats)
	for _, m := range filtered {
		if _, exists := opStats[m.Operation]; !exists {
			opStats[m.Operation] = &OpStats{
				MinDuration: m.Duration,
				MaxDuration: m.Duration,
			}
		}

		stats := opStats[m.Operation]
		stats.Count++
		stats.AvgDuration += m.Duration

		if m.Duration < stats.MinDuration {
			stats.MinDuration = m.Duration
		}
		if m.Duration > stats.MaxDuration {
			stats.MaxDuration = m.Duration
		}
	}

	// Finalize averages
	for op, stats := range opStats {
		stats.AvgDuration /= float64(stats.Count)
		dashboard.ByOperation[op] = *stats
	}

	return dashboard
}

// TimedOperation executes a function and records its duration
func (pm *PerformanceMetrics) TimedOperation(operation string, fn func() interface{}) interface{} {
	start := time.Now()
	result := fn()
	duration := float64(time.Since(start).Milliseconds())

	pm.RecordOperation(operation, duration)

	return result
}

// AlertSlowOperations returns operations slower than threshold (ms)
func (pm *PerformanceMetrics) AlertSlowOperations(thresholdMs float64) []SlowOperation {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	slowOps := make([]SlowOperation, 0)
	for _, m := range pm.metrics {
		if m.Duration > thresholdMs {
			slowOps = append(slowOps, SlowOperation{
				Operation: m.Operation,
				Duration:  m.Duration,
				Timestamp: m.Timestamp,
			})
		}
	}

	// Sort by duration (descending)
	sort.Slice(slowOps, func(i, j int) bool {
		return slowOps[i].Duration > slowOps[j].Duration
	})

	return slowOps
}

// SaveMetrics persists metrics to disk
func (pm *PerformanceMetrics) SaveMetrics() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(pm.metricsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal metrics to JSON
	data, err := json.MarshalIndent(pm.metrics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pm.metricsPath, data, 0644)
}

// loadMetrics loads metrics from disk
func (pm *PerformanceMetrics) loadMetrics() error {
	data, err := os.ReadFile(pm.metricsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No metrics file yet
		}
		return err
	}

	return json.Unmarshal(data, &pm.metrics)
}

// calculatePeriod calculates the cutoff time for a given period
func (pm *PerformanceMetrics) calculatePeriod(period string) time.Time {
	now := time.Now()

	switch period {
	case "last_hour":
		return now.Add(-1 * time.Hour)
	case "last_24h":
		return now.Add(-24 * time.Hour)
	case "last_7_days":
		return now.Add(-7 * 24 * time.Hour)
	case "last_30_days":
		return now.Add(-30 * 24 * time.Hour)
	default: // "all"
		return time.Time{} // Zero time (beginning of time)
	}
}

// percentile calculates the nth percentile of sorted values
func (pm *PerformanceMetrics) percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}

	// Calculate index
	index := float64(len(sorted)-1) * float64(p) / 100.0
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}
