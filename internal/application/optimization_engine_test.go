package application

import (
	"testing"
)

func TestOptimizationEngine_NewEngine(t *testing.T) {
	config := DefaultOptimizationEngineConfig()
	engine := NewOptimizationEngine(config)

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}

	if engine.config.SlowToolThresholdMs != 500.0 {
		t.Errorf("Expected slow tool threshold 500, got %f", engine.config.SlowToolThresholdMs)
	}
}

func TestOptimizationEngine_DefaultConfig(t *testing.T) {
	config := DefaultOptimizationEngineConfig()

	if config.SlowToolThresholdMs != 500.0 {
		t.Errorf("Expected SlowToolThresholdMs=500, got %f", config.SlowToolThresholdMs)
	}
	if config.HighTokenUsageThreshold != 10000 {
		t.Errorf("Expected HighTokenUsageThreshold=10000, got %d", config.HighTokenUsageThreshold)
	}
	if config.LowSuccessRateThreshold != 0.95 {
		t.Errorf("Expected LowSuccessRateThreshold=0.95, got %f", config.LowSuccessRateThreshold)
	}
	if config.MinOperationsForAnalysis != 10 {
		t.Errorf("Expected MinOperationsForAnalysis=10, got %d", config.MinOperationsForAnalysis)
	}
}

func TestOptimizationEngine_GenerateRecommendations_NoData(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	// Test with nil stats
	recs := engine.GenerateRecommendations(nil, DetailedTokenStats{})
	if len(recs) > 0 {
		t.Error("Expected no recommendations with nil stats")
	}

	// Test with insufficient operations
	stats := &UsageStatistics{
		TotalOperations: 5, // Below minimum
		OperationsByTool: map[string]int{
			"test_tool": 5,
		},
		AvgDurationByTool: map[string]float64{
			"test_tool": 100,
		},
	}

	recs = engine.GenerateRecommendations(stats, DetailedTokenStats{})
	if len(recs) > 0 {
		t.Error("Expected no recommendations with insufficient operations")
	}
}

func TestOptimizationEngine_PerformanceRecommendations(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	stats := &UsageStatistics{
		TotalOperations: 100,
		OperationsByTool: map[string]int{
			"slow_tool":   50,
			"normal_tool": 50,
		},
		AvgDurationByTool: map[string]float64{
			"slow_tool":   1500, // Above threshold
			"normal_tool": 100,  // Normal
		},
		ErrorsByTool: map[string]int{
			"slow_tool":   0,
			"normal_tool": 0,
		},
		SuccessRate: 1.0,
	}

	recs := engine.GenerateRecommendations(stats, DetailedTokenStats{})

	// Debug: print generated recommendations
	if len(recs) == 0 {
		t.Log("No recommendations generated - checking if stats are correctly configured")
		t.Logf("Total operations: %d", stats.TotalOperations)
		t.Logf("Slow tool duration: %f", stats.AvgDurationByTool["slow_tool"])
		t.Logf("Threshold: %f", engine.config.SlowToolThresholdMs)
	}

	// Should generate performance recommendations
	foundPerf := false
	for _, rec := range recs {
		if rec.Type == RecommendationTypePerformance {
			foundPerf = true
			// Tool-specific recommendations (not P95) should have affected tools
			if rec.ID != "perf-high-p95" && len(rec.AffectedTools) == 0 {
				t.Errorf("Tool-specific performance recommendation should have affected tools, rec: %+v", rec)
			}
			if rec.ImpactScore < 0 || rec.ImpactScore > 100 {
				t.Errorf("Impact score should be 0-100, got %f", rec.ImpactScore)
			}
			if rec.Priority == "" {
				t.Error("Priority should be set")
			}
		}
	}

	if !foundPerf {
		t.Error("Expected performance recommendations for slow tool")
	}
}

func TestOptimizationEngine_TokenRecommendations(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	stats := &UsageStatistics{
		TotalOperations: 100,
		OperationsByTool: map[string]int{
			"token_heavy": 50,
		},
		AvgDurationByTool: map[string]float64{
			"token_heavy": 100,
		},
		ErrorsByTool: map[string]int{
			"token_heavy": 0,
		},
		SuccessRate: 1.0,
	}

	tokenStats := DetailedTokenStats{
		TotalOriginalTokens:  20000,
		TotalOptimizedTokens: 18000,
		OriginalTokensByTool: map[string]int{
			"token_heavy": 15000, // Above threshold
		},
		OptimizedTokensByTool: map[string]int{
			"token_heavy": 13500,
		},
	}

	recs := engine.GenerateRecommendations(stats, tokenStats)

	foundToken := false
	for _, rec := range recs {
		if rec.Type == RecommendationTypeTokens {
			foundToken = true
			if rec.EstimatedSavings == "" {
				t.Error("Token recommendation should have estimated savings")
			}
		}
	}

	if !foundToken {
		t.Error("Expected token recommendations for heavy token user")
	}
}

func TestOptimizationEngine_ReliabilityRecommendations(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	stats := &UsageStatistics{
		TotalOperations: 100,
		OperationsByTool: map[string]int{
			"unreliable_tool": 50,
			"reliable_tool":   50,
		},
		AvgDurationByTool: map[string]float64{
			"unreliable_tool": 100,
			"reliable_tool":   100,
		},
		ErrorsByTool: map[string]int{
			"unreliable_tool": 10, // 20% error rate
			"reliable_tool":   1,  // 2% error rate
		},
		SuccessRate: 0.89, // 89% overall
	}

	recs := engine.GenerateRecommendations(stats, DetailedTokenStats{})

	foundReliability := false
	for _, rec := range recs {
		if rec.Type == RecommendationTypeReliability {
			foundReliability = true
			if rec.Priority == PriorityLow {
				t.Error("Low success rate should have higher priority")
			}
		}
	}

	if !foundReliability {
		t.Error("Expected reliability recommendations")
	}
}

func TestOptimizationEngine_ArchitectureRecommendations(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	stats := &UsageStatistics{
		TotalOperations: 350,
		OperationsByTool: map[string]int{
			"hot_tool":    250, // Much higher: 250 vs avg 116 = 2.15x
			"normal_tool": 50,
			"low_tool":    50,
		},
		AvgDurationByTool: map[string]float64{
			"hot_tool":    100,
			"normal_tool": 100,
			"low_tool":    100,
		},
		ErrorsByTool: map[string]int{
			"hot_tool":    0,
			"normal_tool": 0,
			"low_tool":    0,
		},
		SuccessRate: 1.0,
	}

	recs := engine.GenerateRecommendations(stats, DetailedTokenStats{})

	// Debug
	t.Logf("Total recommendations generated: %d", len(recs))
	for i, rec := range recs {
		t.Logf("Rec %d: Type=%s, ID=%s, Title=%s", i, rec.Type, rec.ID, rec.Title)
	}

	foundArch := false
	for _, rec := range recs {
		if rec.Type == RecommendationTypeArchitecture {
			foundArch = true
			if len(rec.AffectedTools) == 0 {
				t.Error("Architecture recommendation should have affected tools")
			}
		}
	}

	if !foundArch {
		t.Error("Expected architecture recommendations for high-traffic tools")
	}
}

func TestOptimizationEngine_CostRecommendations(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	stats := &UsageStatistics{
		TotalOperations: 5000,
		OperationsByTool: map[string]int{
			"tool1": 2500,
			"tool2": 2500,
		},
		AvgDurationByTool: map[string]float64{
			"tool1": 200,
			"tool2": 150,
		},
		ErrorsByTool: map[string]int{
			"tool1": 0,
			"tool2": 0,
		},
		SuccessRate: 1.0,
	}

	tokenStats := DetailedTokenStats{
		TotalOriginalTokens:  50000,
		TotalOptimizedTokens: 40000,
	}

	recs := engine.GenerateRecommendations(stats, tokenStats)

	foundCost := false
	for _, rec := range recs {
		if rec.Type == RecommendationTypeCost {
			foundCost = true
			if rec.EstimatedSavings == "" {
				t.Error("Cost recommendation should have estimated savings")
			}
		}
	}

	if !foundCost {
		t.Error("Expected cost recommendations with high operations")
	}
}

func TestOptimizationEngine_PrioritySorting(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	stats := &UsageStatistics{
		TotalOperations: 100,
		OperationsByTool: map[string]int{
			"slow_critical": 50,
			"slow_medium":   30,
			"unreliable":    20,
		},
		AvgDurationByTool: map[string]float64{
			"slow_critical": 6000, // Very slow -> critical
			"slow_medium":   1500, // Slow -> high
			"unreliable":    100,
		},
		ErrorsByTool: map[string]int{
			"slow_critical": 0,
			"slow_medium":   0,
			"unreliable":    5, // 25% error -> critical
		},
		SuccessRate: 0.95,
	}

	recs := engine.GenerateRecommendations(stats, DetailedTokenStats{})

	if len(recs) == 0 {
		t.Fatal("Expected recommendations")
	}

	// Check recommendations are sorted by priority
	for i := 1; i < len(recs); i++ {
		prev := recs[i-1]
		curr := recs[i]

		priorityOrder := map[RecommendationPriority]int{
			PriorityCritical: 0,
			PriorityHigh:     1,
			PriorityMedium:   2,
			PriorityLow:      3,
		}

		if priorityOrder[curr.Priority] < priorityOrder[prev.Priority] {
			t.Errorf("Recommendations not sorted by priority: %s before %s", prev.Priority, curr.Priority)
		}
	}
}

func TestOptimizationEngine_RecommendationFields(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	stats := &UsageStatistics{
		TotalOperations: 100,
		OperationsByTool: map[string]int{
			"test_tool": 100,
		},
		AvgDurationByTool: map[string]float64{
			"test_tool": 2000, // Slow
		},
		ErrorsByTool: map[string]int{
			"test_tool": 0,
		},
		SuccessRate: 1.0,
	}

	recs := engine.GenerateRecommendations(stats, DetailedTokenStats{})

	if len(recs) == 0 {
		t.Fatal("Expected at least one recommendation")
	}

	for i, rec := range recs {
		if rec.ID == "" {
			t.Errorf("Recommendation %d missing ID", i)
		}
		if rec.Title == "" {
			t.Errorf("Recommendation %d missing title", i)
		}
		if rec.Description == "" {
			t.Errorf("Recommendation %d missing description", i)
		}
		if rec.Type == "" {
			t.Errorf("Recommendation %d missing type", i)
		}
		if rec.Priority == "" {
			t.Errorf("Recommendation %d missing priority", i)
		}
		if rec.ImplementationEffort == "" {
			t.Errorf("Recommendation %d missing implementation effort", i)
		}
		if len(rec.ActionItems) == 0 {
			t.Errorf("Recommendation %d missing action items", i)
		}
		if rec.Evidence == nil {
			t.Errorf("Recommendation %d missing evidence", i)
		}
		if rec.GeneratedAt.IsZero() {
			t.Errorf("Recommendation %d missing generated time", i)
		}
	}
}

func TestOptimizationEngine_ImpactScore(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	tests := []struct {
		name     string
		value    float64
		total    float64
		weight   float64
		expected bool // Whether score should be > 0
	}{
		{"zero contribution", 0, 100, 0.4, false},
		{"partial contribution", 30, 100, 0.4, true},
		{"full contribution", 100, 100, 0.4, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := engine.calculateImpactScore(tt.value, tt.total, tt.weight)

			if tt.expected && score == 0 {
				t.Error("Expected non-zero impact score")
			}
			if !tt.expected && score != 0 {
				t.Error("Expected zero impact score")
			}
			if score < 0 || score > 100 {
				t.Errorf("Impact score should be 0-100, got %f", score)
			}
		})
	}
}

// Test concurrent recommendation generation.
func TestOptimizationEngine_ConcurrentRecommendations(t *testing.T) {
	engine := NewOptimizationEngine(DefaultOptimizationEngineConfig())

	stats := &UsageStatistics{
		TotalOperations: 100,
		OperationsByTool: map[string]int{
			"tool1": 50,
			"tool2": 50,
		},
		AvgDurationByTool: map[string]float64{
			"tool1": 1000,
			"tool2": 500,
		},
		ErrorsByTool: map[string]int{
			"tool1": 0,
			"tool2": 0,
		},
		SuccessRate: 1.0,
	}

	done := make(chan bool)
	for range 10 {
		go func() {
			recs := engine.GenerateRecommendations(stats, DetailedTokenStats{})
			if len(recs) == 0 {
				t.Error("Expected recommendations")
			}
			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}
