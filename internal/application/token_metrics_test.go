package application

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewTokenMetricsCollector(t *testing.T) {
	tempDir := t.TempDir()

	tmc := NewTokenMetricsCollector(tempDir, 5*time.Second)

	if tmc == nil {
		t.Fatal("NewTokenMetricsCollector returned nil")
	}

	if tmc.maxMetrics != 10000 {
		t.Errorf("Expected maxMetrics=10000, got %d", tmc.maxMetrics)
	}

	if !tmc.autoSave {
		t.Error("Expected autoSave=true")
	}
}

func TestRecordTokenOptimization(t *testing.T) {
	tempDir := t.TempDir()
	tmc := NewTokenMetricsCollector(tempDir, 5*time.Second)

	metric := TokenMetrics{
		OriginalTokens:   1000,
		OptimizedTokens:  650,
		TokensSaved:      350,
		CompressionRatio: 0.65,
		OptimizationType: "response_compression",
		ToolName:         "test_tool",
		Timestamp:        time.Now(),
	}

	tmc.RecordTokenOptimization(metric)

	stats := tmc.GetStats()

	if stats.TotalOriginalTokens != 1000 {
		t.Errorf("Expected TotalOriginalTokens=1000, got %d", stats.TotalOriginalTokens)
	}

	if stats.TotalOptimizedTokens != 650 {
		t.Errorf("Expected TotalOptimizedTokens=650, got %d", stats.TotalOptimizedTokens)
	}

	if stats.TotalTokensSaved != 350 {
		t.Errorf("Expected TotalTokensSaved=350, got %d", stats.TotalTokensSaved)
	}

	if stats.OptimizationCount != 1 {
		t.Errorf("Expected OptimizationCount=1, got %d", stats.OptimizationCount)
	}
}

func TestRecordTokenOptimizationAutoCalculation(t *testing.T) {
	tempDir := t.TempDir()
	tmc := NewTokenMetricsCollector(tempDir, 5*time.Second)

	// Test auto-calculation of TokensSaved and CompressionRatio
	metric := TokenMetrics{
		OriginalTokens:   1000,
		OptimizedTokens:  600,
		OptimizationType: "prompt_compression",
		ToolName:         "test_tool",
	}

	tmc.RecordTokenOptimization(metric)

	recent := tmc.GetRecentMetrics(1)
	if len(recent) != 1 {
		t.Fatal("Expected 1 recent metric")
	}

	if recent[0].TokensSaved != 400 {
		t.Errorf("Expected auto-calculated TokensSaved=400, got %d", recent[0].TokensSaved)
	}

	expectedRatio := 0.6
	if recent[0].CompressionRatio != expectedRatio {
		t.Errorf("Expected auto-calculated CompressionRatio=%.2f, got %.2f",
			expectedRatio, recent[0].CompressionRatio)
	}
}

func TestGetStatsByType(t *testing.T) {
	tempDir := t.TempDir()
	tmc := NewTokenMetricsCollector(tempDir, 5*time.Second)

	metrics := []TokenMetrics{
		{
			OriginalTokens:   1000,
			OptimizedTokens:  600,
			TokensSaved:      400,
			OptimizationType: "response_compression",
			ToolName:         "tool1",
		},
		{
			OriginalTokens:   2000,
			OptimizedTokens:  1300,
			TokensSaved:      700,
			OptimizationType: "prompt_compression",
			ToolName:         "tool2",
		},
		{
			OriginalTokens:   500,
			OptimizedTokens:  300,
			TokensSaved:      200,
			OptimizationType: "response_compression",
			ToolName:         "tool3",
		},
	}

	for _, m := range metrics {
		tmc.RecordTokenOptimization(m)
	}

	stats := tmc.GetStats()

	// Check aggregation by type
	if stats.TokensSavedByType["response_compression"] != 600 {
		t.Errorf("Expected 600 tokens saved for response_compression, got %d",
			stats.TokensSavedByType["response_compression"])
	}

	if stats.TokensSavedByType["prompt_compression"] != 700 {
		t.Errorf("Expected 700 tokens saved for prompt_compression, got %d",
			stats.TokensSavedByType["prompt_compression"])
	}

	// Check aggregation by tool
	if stats.TokensSavedByTool["tool1"] != 400 {
		t.Errorf("Expected 400 tokens saved for tool1, got %d",
			stats.TokensSavedByTool["tool1"])
	}
}

func TestSaveAndLoadMetrics(t *testing.T) {
	tempDir := t.TempDir()

	// Create collector and add metrics
	tmc1 := NewTokenMetricsCollector(tempDir, 5*time.Second)

	metrics := []TokenMetrics{
		{
			OriginalTokens:   1000,
			OptimizedTokens:  650,
			TokensSaved:      350,
			OptimizationType: "response_compression",
			ToolName:         "tool1",
		},
		{
			OriginalTokens:   2000,
			OptimizedTokens:  1400,
			TokensSaved:      600,
			OptimizationType: "prompt_compression",
			ToolName:         "tool2",
		},
	}

	for _, m := range metrics {
		tmc1.RecordTokenOptimization(m)
	}

	// Save metrics
	if err := tmc1.SaveMetrics(); err != nil {
		t.Fatalf("Failed to save metrics: %v", err)
	}

	// Create new collector and load metrics
	tmc2 := NewTokenMetricsCollector(tempDir, 5*time.Second)

	stats := tmc2.GetStats()

	if stats.TotalOriginalTokens != 3000 {
		t.Errorf("Expected TotalOriginalTokens=3000 after load, got %d", stats.TotalOriginalTokens)
	}

	if stats.TotalTokensSaved != 950 {
		t.Errorf("Expected TotalTokensSaved=950 after load, got %d", stats.TotalTokensSaved)
	}

	if stats.OptimizationCount != 2 {
		t.Errorf("Expected OptimizationCount=2 after load, got %d", stats.OptimizationCount)
	}
}

func TestGetRecentMetrics(t *testing.T) {
	tempDir := t.TempDir()
	tmc := NewTokenMetricsCollector(tempDir, 5*time.Second)

	// Add 10 metrics
	for range 10 {
		tmc.RecordTokenOptimization(TokenMetrics{
			OriginalTokens:   1000,
			OptimizedTokens:  600,
			OptimizationType: "test",
			ToolName:         "tool",
		})
	}

	// Get last 5
	recent := tmc.GetRecentMetrics(5)
	if len(recent) != 5 {
		t.Errorf("Expected 5 recent metrics, got %d", len(recent))
	}

	// Get more than available
	recent = tmc.GetRecentMetrics(20)
	if len(recent) != 10 {
		t.Errorf("Expected 10 metrics when requesting 20, got %d", len(recent))
	}
}

func TestMaxMetricsLimit(t *testing.T) {
	tempDir := t.TempDir()
	tmc := NewTokenMetricsCollector(tempDir, 5*time.Second)
	tmc.maxMetrics = 5 // Set low limit for testing

	// Add more metrics than limit
	for range 10 {
		tmc.RecordTokenOptimization(TokenMetrics{
			OriginalTokens:   1000,
			OptimizedTokens:  600,
			OptimizationType: "test",
			ToolName:         "tool",
		})
	}

	// Should only keep last 5
	recent := tmc.GetRecentMetrics(100)
	if len(recent) != 5 {
		t.Errorf("Expected maxMetrics limit of 5, got %d metrics", len(recent))
	}
}

func TestEstimateTokenCount(t *testing.T) {
	tests := []struct {
		text     string
		expected int64
	}{
		{"Hello, World!", 3},             // 13 chars / 4 = 3
		{"A", 0},                         // 1 char / 4 = 0
		{"12345678", 2},                  // 8 chars / 4 = 2
		{string(make([]byte, 400)), 100}, // 400 chars / 4 = 100
	}

	for _, tt := range tests {
		result := EstimateTokenCount(tt.text)
		if result != tt.expected {
			t.Errorf("EstimateTokenCount(%q) = %d, expected %d",
				tt.text, result, tt.expected)
		}
	}
}

func TestEstimateTokenCountFromBytes(t *testing.T) {
	tests := []struct {
		size     int
		expected int64
	}{
		{1000, 250},
		{4000, 1000},
		{100, 25},
		{0, 0},
	}

	for _, tt := range tests {
		result := EstimateTokenCountFromBytes(tt.size)
		if result != tt.expected {
			t.Errorf("EstimateTokenCountFromBytes(%d) = %d, expected %d",
				tt.size, result, tt.expected)
		}
	}
}

func TestAvgCompressionRatio(t *testing.T) {
	tempDir := t.TempDir()
	tmc := NewTokenMetricsCollector(tempDir, 5*time.Second)

	// Add metrics with known compression ratios
	tmc.RecordTokenOptimization(TokenMetrics{
		OriginalTokens:   1000,
		OptimizedTokens:  500, // 0.5 ratio
		OptimizationType: "test",
		ToolName:         "tool",
	})

	tmc.RecordTokenOptimization(TokenMetrics{
		OriginalTokens:   1000,
		OptimizedTokens:  700, // 0.7 ratio
		OptimizationType: "test",
		ToolName:         "tool",
	})

	stats := tmc.GetStats()

	// Overall ratio should be (500+700)/(1000+1000) = 1200/2000 = 0.6
	expectedRatio := 0.6
	if stats.AvgCompressionRatio != expectedRatio {
		t.Errorf("Expected AvgCompressionRatio=%.2f, got %.2f",
			expectedRatio, stats.AvgCompressionRatio)
	}
}

func TestMetricsFilePath(t *testing.T) {
	tempDir := t.TempDir()
	tmc := NewTokenMetricsCollector(tempDir, 5*time.Second)

	if err := tmc.SaveMetrics(); err != nil {
		t.Fatalf("Failed to save metrics: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "token_metrics.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected metrics file to exist at %s", expectedPath)
	}
}
