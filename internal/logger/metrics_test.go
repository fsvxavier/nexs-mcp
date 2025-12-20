package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPerformanceMetrics_RecordAndGetDashboard(t *testing.T) {
	// Create temp dir for test metrics
	tmpDir := t.TempDir()
	metrics := NewPerformanceMetrics(tmpDir)

	// Record some operations
	metrics.RecordOperation("list_elements", 45.5)
	metrics.RecordOperation("get_element", 12.3)
	metrics.RecordOperation("list_elements", 67.8)
	metrics.RecordOperation("create_element", 234.5)
	metrics.RecordOperation("get_element", 8.9)

	// Get dashboard for all time
	dashboard := metrics.GetDashboard("all")

	// Verify total operations
	if dashboard.TotalOperations != 5 {
		t.Errorf("Expected 5 total operations, got %d", dashboard.TotalOperations)
	}

	// Verify by-operation stats
	if listStats, ok := dashboard.ByOperation["list_elements"]; ok {
		if listStats.Count != 2 {
			t.Errorf("Expected 2 list_elements operations, got %d", listStats.Count)
		}
	} else {
		t.Error("Expected list_elements in ByOperation stats")
	}

	// Verify percentiles are calculated
	if dashboard.P50Duration <= 0 {
		t.Error("Expected P50 duration > 0")
	}
	if dashboard.P95Duration <= 0 {
		t.Error("Expected P95 duration > 0")
	}
	if dashboard.P99Duration <= 0 {
		t.Error("Expected P99 duration > 0")
	}

	// Verify min/max
	if dashboard.MinDuration != 8.9 {
		t.Errorf("Expected min duration 8.9, got %.1f", dashboard.MinDuration)
	}
	if dashboard.MaxDuration != 234.5 {
		t.Errorf("Expected max duration 234.5, got %.1f", dashboard.MaxDuration)
	}
}

func TestPerformanceMetrics_PeriodFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	metrics := NewPerformanceMetrics(tmpDir)

	now := time.Now()

	// Record operations with specific timestamps
	metrics.metrics = []OperationMetric{
		{Operation: "op1", Duration: 10.0, Timestamp: now.Add(-2 * time.Hour)},
		{Operation: "op2", Duration: 20.0, Timestamp: now.Add(-30 * time.Minute)},
		{Operation: "op3", Duration: 30.0, Timestamp: now.Add(-5 * time.Minute)},
	}

	// Test last_hour filter
	dashboard := metrics.GetDashboard("last_hour")
	if dashboard.TotalOperations != 2 {
		t.Errorf("Expected 2 operations in last hour, got %d", dashboard.TotalOperations)
	}

	// Test last_24h filter
	dashboard = metrics.GetDashboard("last_24h")
	if dashboard.TotalOperations != 3 {
		t.Errorf("Expected 3 operations in last 24 hours, got %d", dashboard.TotalOperations)
	}

	// Test "all" period includes everything
	dashboard = metrics.GetDashboard("all")
	if dashboard.TotalOperations != 3 {
		t.Errorf("Expected 3 operations with 'all' period, got %d", dashboard.TotalOperations)
	}
}

func TestPerformanceMetrics_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	metricsPath := filepath.Join(tmpDir, "performance_metrics.json")

	// Create metrics and record operations
	metrics1 := NewPerformanceMetrics(tmpDir)
	metrics1.RecordOperation("test_op", 123.45)
	metrics1.RecordOperation("another_op", 67.89)

	// Save metrics
	if err := metrics1.SaveMetrics(); err != nil {
		t.Fatalf("Failed to save metrics: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		t.Fatal("Metrics file was not created")
	}

	// Load metrics in new instance
	metrics2 := NewPerformanceMetrics(tmpDir)

	// Verify metrics were loaded
	if len(metrics2.metrics) != 2 {
		t.Errorf("Expected 2 loaded metrics, got %d", len(metrics2.metrics))
	}

	// Verify dashboard shows loaded metrics
	dashboard := metrics2.GetDashboard("all")
	if dashboard.TotalOperations != 2 {
		t.Errorf("Expected 2 total operations in loaded metrics, got %d", dashboard.TotalOperations)
	}
}

func TestPerformanceMetrics_TimedOperation(t *testing.T) {
	tmpDir := t.TempDir()
	metrics := NewPerformanceMetrics(tmpDir)

	// Execute a timed operation
	result := metrics.TimedOperation("test_operation", func() interface{} {
		time.Sleep(10 * time.Millisecond)
		return "success"
	})

	// Verify result
	if result != "success" {
		t.Errorf("Expected result 'success', got %v", result)
	}

	// Verify metric was recorded
	dashboard := metrics.GetDashboard("all")
	if dashboard.TotalOperations != 1 {
		t.Errorf("Expected 1 operation recorded, got %d", dashboard.TotalOperations)
	}

	// Verify duration is reasonable (>10ms due to sleep)
	if dashboard.AvgDuration < 10.0 {
		t.Errorf("Expected duration >= 10ms, got %.2fms", dashboard.AvgDuration)
	}
}

func TestPerformanceMetrics_SlowOperationAlerts(t *testing.T) {
	tmpDir := t.TempDir()
	metrics := NewPerformanceMetrics(tmpDir)

	// Record mix of fast and slow operations
	metrics.RecordOperation("fast1", 5.0)
	metrics.RecordOperation("fast2", 8.0)
	metrics.RecordOperation("fast3", 12.0)
	metrics.RecordOperation("slow1", 150.0)
	metrics.RecordOperation("slow2", 200.0)

	// Get slow operation alerts (threshold 100ms)
	slowOps := metrics.AlertSlowOperations(100.0)

	// Should have 2 slow operations
	if len(slowOps) != 2 {
		t.Errorf("Expected 2 slow operations, got %d", len(slowOps))
	}

	// Verify slowest operation is first
	if len(slowOps) > 0 && slowOps[0].Duration != 200.0 {
		t.Errorf("Expected slowest operation with 200ms, got %.1fms", slowOps[0].Duration)
	}

	// Test with very high threshold
	slowOps = metrics.AlertSlowOperations(500.0)
	if len(slowOps) != 0 {
		t.Errorf("Expected 0 slow operations with 500ms threshold, got %d", len(slowOps))
	}
}

func TestPerformanceMetrics_Percentiles(t *testing.T) {
	tmpDir := t.TempDir()
	metrics := NewPerformanceMetrics(tmpDir)

	// Record 100 operations with known distribution
	for i := 1; i <= 100; i++ {
		metrics.RecordOperation("test_op", float64(i))
	}

	dashboard := metrics.GetDashboard("all")

	// Verify P50 (median) is around 50
	if dashboard.P50Duration < 48.0 || dashboard.P50Duration > 52.0 {
		t.Errorf("Expected P50 around 50, got %.1f", dashboard.P50Duration)
	}

	// Verify P95 is around 95
	if dashboard.P95Duration < 93.0 || dashboard.P95Duration > 97.0 {
		t.Errorf("Expected P95 around 95, got %.1f", dashboard.P95Duration)
	}

	// Verify P99 is around 99
	if dashboard.P99Duration < 97.0 || dashboard.P99Duration > 101.0 {
		t.Errorf("Expected P99 around 99, got %.1f", dashboard.P99Duration)
	}
}

func TestPerformanceMetrics_CircularBuffer(t *testing.T) {
	tmpDir := t.TempDir()
	metrics := NewPerformanceMetrics(tmpDir)

	// Record more than max metrics (10000)
	for i := 0; i < 10500; i++ {
		metrics.RecordOperation("test_op", float64(i))
	}

	// Should keep only last 10000
	if len(metrics.metrics) != 10000 {
		t.Errorf("Expected 10000 metrics (circular buffer), got %d", len(metrics.metrics))
	}

	// First metric should be from iteration 500 (oldest kept)
	if metrics.metrics[0].Duration < 500.0 {
		t.Errorf("Expected oldest metric >= 500, got %.0f", metrics.metrics[0].Duration)
	}
}
