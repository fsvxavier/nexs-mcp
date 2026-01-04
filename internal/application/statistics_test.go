package application

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMetricsCollector_RecordAndGetStatistics(t *testing.T) {
	// Create temporary directory for test
	tmpDir := filepath.Join(os.TempDir(), "nexs-mcp-test-metrics")
	defer os.RemoveAll(tmpDir)

	mc := NewMetricsCollector(tmpDir, 30*time.Second)

	// Record some metrics
	mc.RecordToolCall(ToolCallMetric{
		ToolName:  "list_elements",
		Timestamp: time.Now(),
		Duration:  50 * time.Millisecond,
		Success:   true,
		User:      "alice",
	})

	mc.RecordToolCall(ToolCallMetric{
		ToolName:  "get_element",
		Timestamp: time.Now(),
		Duration:  30 * time.Millisecond,
		Success:   true,
		User:      "bob",
	})

	mc.RecordToolCall(ToolCallMetric{
		ToolName:     "create_element",
		Timestamp:    time.Now(),
		Duration:     100 * time.Millisecond,
		Success:      false,
		ErrorMessage: "validation error",
		User:         "alice",
	})

	// Get statistics
	stats, err := mc.GetStatistics("all")
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	// Verify statistics
	if stats.TotalOperations != 3 {
		t.Errorf("Expected 3 total operations, got %d", stats.TotalOperations)
	}

	if stats.SuccessfulOps != 2 {
		t.Errorf("Expected 2 successful operations, got %d", stats.SuccessfulOps)
	}

	if stats.FailedOps != 1 {
		t.Errorf("Expected 1 failed operation, got %d", stats.FailedOps)
	}

	expectedRate := 66.66666666666666
	if stats.SuccessRate < expectedRate-0.1 || stats.SuccessRate > expectedRate+0.1 {
		t.Errorf("Expected success rate around %.2f%%, got %.2f%%", expectedRate, stats.SuccessRate)
	}

	if len(stats.ActiveUsers) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(stats.ActiveUsers))
	}

	if stats.OperationsByTool["list_elements"] != 1 {
		t.Errorf("Expected 1 list_elements operation, got %d", stats.OperationsByTool["list_elements"])
	}

	if stats.ErrorsByTool["create_element"] != 1 {
		t.Errorf("Expected 1 create_element error, got %d", stats.ErrorsByTool["create_element"])
	}
}

func TestMetricsCollector_PeriodFiltering(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "nexs-mcp-test-metrics-period")
	defer os.RemoveAll(tmpDir)

	mc := NewMetricsCollector(tmpDir, 30*time.Second)

	// Record metrics with different timestamps
	now := time.Now()

	// Old metric (25 hours ago)
	mc.RecordToolCall(ToolCallMetric{
		ToolName:  "old_tool",
		Timestamp: now.Add(-25 * time.Hour),
		Duration:  10 * time.Millisecond,
		Success:   true,
	})

	// Recent metric (1 hour ago)
	mc.RecordToolCall(ToolCallMetric{
		ToolName:  "recent_tool",
		Timestamp: now.Add(-1 * time.Hour),
		Duration:  20 * time.Millisecond,
		Success:   true,
	})

	// Current metric
	mc.RecordToolCall(ToolCallMetric{
		ToolName:  "current_tool",
		Timestamp: now,
		Duration:  30 * time.Millisecond,
		Success:   true,
	})

	// Test last_24h period
	stats24h, err := mc.GetStatistics("last_24h")
	if err != nil {
		t.Fatalf("Failed to get last_24h statistics: %v", err)
	}

	if stats24h.TotalOperations != 2 {
		t.Errorf("Expected 2 operations in last_24h, got %d", stats24h.TotalOperations)
	}

	// Test all period
	statsAll, err := mc.GetStatistics("all")
	if err != nil {
		t.Fatalf("Failed to get all statistics: %v", err)
	}

	if statsAll.TotalOperations != 3 {
		t.Errorf("Expected 3 operations in all period, got %d", statsAll.TotalOperations)
	}
}

func TestMetricsCollector_SaveAndLoad(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "nexs-mcp-test-metrics-persistence")
	defer os.RemoveAll(tmpDir)

	// Create collector and add metrics
	mc1 := NewMetricsCollector(tmpDir, 30*time.Second)

	mc1.RecordToolCall(ToolCallMetric{
		ToolName:  "test_tool",
		Timestamp: time.Now(),
		Duration:  100 * time.Millisecond,
		Success:   true,
	})

	// Save metrics
	if err := mc1.SaveMetrics(); err != nil {
		t.Fatalf("Failed to save metrics: %v", err)
	}

	// Create new collector (should load existing metrics)
	mc2 := NewMetricsCollector(tmpDir, 30*time.Second)

	if mc2.GetMetricsCount() != 1 {
		t.Errorf("Expected 1 metric after loading, got %d", mc2.GetMetricsCount())
	}

	stats, err := mc2.GetStatistics("all")
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	if stats.TotalOperations != 1 {
		t.Errorf("Expected 1 operation in loaded metrics, got %d", stats.TotalOperations)
	}
}

func TestMetricsCollector_MostUsedTools(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "nexs-mcp-test-metrics-most-used")
	defer os.RemoveAll(tmpDir)

	mc := NewMetricsCollector(tmpDir, 30*time.Second)

	// Record multiple calls to different tools
	for range 10 {
		mc.RecordToolCall(ToolCallMetric{
			ToolName:  "popular_tool",
			Timestamp: time.Now(),
			Duration:  50 * time.Millisecond,
			Success:   true,
		})
	}

	for range 5 {
		mc.RecordToolCall(ToolCallMetric{
			ToolName:  "medium_tool",
			Timestamp: time.Now(),
			Duration:  30 * time.Millisecond,
			Success:   true,
		})
	}

	mc.RecordToolCall(ToolCallMetric{
		ToolName:  "rare_tool",
		Timestamp: time.Now(),
		Duration:  20 * time.Millisecond,
		Success:   true,
	})

	stats, err := mc.GetStatistics("all")
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	// Verify most used tools are sorted correctly
	if len(stats.MostUsedTools) < 3 {
		t.Fatalf("Expected at least 3 tools in MostUsedTools, got %d", len(stats.MostUsedTools))
	}

	if stats.MostUsedTools[0].ToolName != "popular_tool" {
		t.Errorf("Expected most used tool to be 'popular_tool', got '%s'", stats.MostUsedTools[0].ToolName)
	}

	if stats.MostUsedTools[0].Count != 10 {
		t.Errorf("Expected popular_tool count to be 10, got %d", stats.MostUsedTools[0].Count)
	}

	if stats.MostUsedTools[1].ToolName != "medium_tool" {
		t.Errorf("Expected second most used tool to be 'medium_tool', got '%s'", stats.MostUsedTools[1].ToolName)
	}
}

func TestMetricsCollector_ClearMetrics(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "nexs-mcp-test-metrics-clear")
	defer os.RemoveAll(tmpDir)

	mc := NewMetricsCollector(tmpDir, 30*time.Second)

	// Add some metrics
	mc.RecordToolCall(ToolCallMetric{
		ToolName:  "test_tool",
		Timestamp: time.Now(),
		Duration:  100 * time.Millisecond,
		Success:   true,
	})

	if mc.GetMetricsCount() != 1 {
		t.Errorf("Expected 1 metric before clear, got %d", mc.GetMetricsCount())
	}

	// Clear metrics
	if err := mc.ClearMetrics(); err != nil {
		t.Fatalf("Failed to clear metrics: %v", err)
	}

	if mc.GetMetricsCount() != 0 {
		t.Errorf("Expected 0 metrics after clear, got %d", mc.GetMetricsCount())
	}

	stats, err := mc.GetStatistics("all")
	if err != nil {
		t.Fatalf("Failed to get statistics after clear: %v", err)
	}

	if stats.TotalOperations != 0 {
		t.Errorf("Expected 0 operations after clear, got %d", stats.TotalOperations)
	}
}
