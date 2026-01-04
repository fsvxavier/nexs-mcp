package mcp

import (
	"context"
	"fmt"
	"sort"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
)

// handleGetCostAnalytics handles get_cost_analytics tool calls.
// Provides comprehensive cost analytics including trends, optimization opportunities,
// and actionable recommendations.
func (s *MCPServer) handleGetCostAnalytics(ctx context.Context, req *sdk.CallToolRequest, input GetCostAnalyticsInput) (*sdk.CallToolResult, GetCostAnalyticsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_cost_analytics",
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

	// Default analysis period
	period := input.Period
	if period == "" {
		period = "last_24h"
	}

	// Gather performance metrics
	perfStats, err := s.metrics.GetStatistics(period)
	if err != nil {
		handlerErr = err
		return nil, GetCostAnalyticsOutput{}, handlerErr
	}

	// Gather token metrics
	tokenStats := s.tokenMetrics.GetDetailedStats()

	// Calculate cost breakdown by tool
	toolCosts := s.calculateToolCosts(perfStats, tokenStats)

	// Identify expensive tools (top 10 by cost)
	expensiveTools := s.identifyExpensiveTools(toolCosts, 10)

	// Analyze trends
	trends := s.analyzeCostTrends(perfStats, tokenStats, period)

	// Generate optimization opportunities
	opportunities := s.generateOptimizationOpportunities(perfStats, tokenStats, toolCosts)

	// Generate recommendations
	recommendations := s.generateCostRecommendations(perfStats, tokenStats, toolCosts, trends)

	// Detect anomalies
	anomalies := s.detectCostAnomalies(perfStats, tokenStats)

	// Calculate cost projections
	projections := s.projectFutureCosts(trends, period)

	output := GetCostAnalyticsOutput{
		Period:                    period,
		StartTime:                 perfStats.StartTime.Format("2006-01-02T15:04:05Z07:00"),
		EndTime:                   perfStats.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		TotalOperations:           perfStats.TotalOperations,
		TotalTokens:               tokenStats.TotalOriginalTokens,
		TotalOptimizedTokens:      tokenStats.TotalOptimizedTokens,
		TokenSavings:              tokenStats.TotalOriginalTokens - tokenStats.TotalOptimizedTokens,
		TokenSavingsPercent:       s.calculatePercentage(tokenStats.TotalOriginalTokens-tokenStats.TotalOptimizedTokens, tokenStats.TotalOriginalTokens),
		TotalDuration:             s.calculateTotalDuration(perfStats),
		AverageDuration:           s.calculateAverageDuration(perfStats),
		ToolCostBreakdown:         toolCosts,
		TopExpensiveTools:         expensiveTools,
		CostTrends:                trends,
		OptimizationOpportunities: opportunities,
		Recommendations:           recommendations,
		Anomalies:                 anomalies,
		CostProjections:           projections,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "get_cost_analytics", output)

	return nil, output, nil
}

// calculateToolCosts calculates the cost breakdown per tool.
func (s *MCPServer) calculateToolCosts(perfStats *application.UsageStatistics, tokenStats application.DetailedTokenStats) []ToolCostBreakdown {
	costMap := make(map[string]*ToolCostBreakdown)

	// Aggregate performance metrics
	for toolName, count := range perfStats.OperationsByTool {
		if costMap[toolName] == nil {
			costMap[toolName] = &ToolCostBreakdown{
				ToolName: toolName,
			}
		}
		costMap[toolName].OperationCount = count
		costMap[toolName].AvgDuration = perfStats.AvgDurationByTool[toolName]
		costMap[toolName].TotalDuration = float64(count) * perfStats.AvgDurationByTool[toolName]
		costMap[toolName].SuccessRate = s.calculateToolSuccessRate(toolName, perfStats)
	}

	// Aggregate token metrics
	for toolName, tokens := range tokenStats.OriginalTokensByTool {
		if costMap[toolName] == nil {
			costMap[toolName] = &ToolCostBreakdown{
				ToolName: toolName,
			}
		}
		costMap[toolName].TotalTokens = tokens
		costMap[toolName].OptimizedTokens = tokenStats.OptimizedTokensByTool[toolName]
		costMap[toolName].TokenSavings = tokens - tokenStats.OptimizedTokensByTool[toolName]
		costMap[toolName].CompressionRatio = s.calculateCompressionRatio(toolName, tokenStats)
	}

	// Calculate relative cost score (normalized)
	maxDuration := 0.0
	maxTokens := 0
	for _, cost := range costMap {
		if cost.TotalDuration > maxDuration {
			maxDuration = cost.TotalDuration
		}
		if cost.TotalTokens > maxTokens {
			maxTokens = cost.TotalTokens
		}
	}

	for _, cost := range costMap {
		// Cost score is weighted: 60% duration, 40% tokens
		durationScore := 0.0
		if maxDuration > 0 {
			durationScore = (cost.TotalDuration / maxDuration) * 60.0
		}
		tokenScore := 0.0
		if maxTokens > 0 {
			tokenScore = (float64(cost.TotalTokens) / float64(maxTokens)) * 40.0
		}
		cost.CostScore = durationScore + tokenScore
	}

	// Convert map to sorted slice
	costs := make([]ToolCostBreakdown, 0, len(costMap))
	for _, cost := range costMap {
		costs = append(costs, *cost)
	}

	// Sort by cost score descending
	sort.Slice(costs, func(i, j int) bool {
		return costs[i].CostScore > costs[j].CostScore
	})

	return costs
}

// identifyExpensiveTools returns the top N most expensive tools.
func (s *MCPServer) identifyExpensiveTools(toolCosts []ToolCostBreakdown, limit int) []ExpensiveTool {
	if len(toolCosts) < limit {
		limit = len(toolCosts)
	}

	expensive := make([]ExpensiveTool, limit)
	for i := range limit {
		cost := toolCosts[i]
		expensive[i] = ExpensiveTool{
			ToolName:         cost.ToolName,
			TotalCost:        cost.CostScore,
			OperationCount:   cost.OperationCount,
			AvgDuration:      cost.AvgDuration,
			TotalTokens:      cost.TotalTokens,
			CostPerOperation: cost.CostScore / float64(cost.OperationCount),
		}
	}

	return expensive
}

// analyzeCostTrends analyzes cost trends over time.
func (s *MCPServer) analyzeCostTrends(perfStats *application.UsageStatistics, tokenStats application.DetailedTokenStats, period string) CostTrends {
	// Calculate current metrics
	totalOps := float64(perfStats.TotalOperations)
	totalTokens := float64(tokenStats.TotalOriginalTokens)
	avgDuration := s.calculateAverageDuration(perfStats)

	// Initialize change percentages
	operationsChange := 0.0
	tokenUsageChange := 0.0
	avgDurationChange := 0.0

	// Analyze trends by comparing with historical baselines
	// For now, we'll use heuristics based on current data patterns
	// In production, this would load historical metrics from previous periods

	// Operations trend analysis
	if totalOps > 0 {
		// Estimate change based on operation distribution over time
		if len(perfStats.OperationsByPeriod) > 1 {
			// Calculate trend from period distribution
			var periodCounts []float64
			for _, count := range perfStats.OperationsByPeriod {
				periodCounts = append(periodCounts, float64(count))
			}
			if len(periodCounts) >= 2 {
				// Simple linear trend: compare last period with average of previous periods
				lastPeriod := periodCounts[len(periodCounts)-1]
				avgPrevious := 0.0
				for i := range len(periodCounts) - 1 {
					avgPrevious += periodCounts[i]
				}
				if len(periodCounts) > 1 {
					avgPrevious /= float64(len(periodCounts) - 1)
				}
				if avgPrevious > 0 {
					operationsChange = ((lastPeriod - avgPrevious) / avgPrevious) * 100
				}
			}
		}
	}

	// Token usage trend analysis
	if totalTokens > 0 {
		// Analyze token efficiency trends
		compressionRatio := 0.0
		if tokenStats.TotalOriginalTokens > 0 {
			compressionRatio = float64(tokenStats.TotalOriginalTokens-tokenStats.TotalOptimizedTokens) / float64(tokenStats.TotalOriginalTokens)
		}

		// Estimate token usage change based on compression effectiveness
		// If compression ratio is improving, token costs are decreasing
		if compressionRatio > 0.3 {
			// Good compression: tokens trending down
			tokenUsageChange = -10.0 // -10% effective cost reduction
		} else if compressionRatio < 0.1 {
			// Poor compression: tokens trending up (inefficient)
			tokenUsageChange = 5.0 // +5% effective cost increase
		}

		// Adjust based on total volume
		if totalTokens > 100000 {
			// High volume: amplify trend
			tokenUsageChange *= 1.5
		}
	}

	// Duration trend analysis
	if avgDuration > 0 {
		// Analyze performance trends based on operation mix
		slowOpsCount := 0
		fastOpsCount := 0
		totalOpsAnalyzed := 0

		for _, duration := range perfStats.AvgDurationByTool {
			totalOpsAnalyzed++
			if duration > 500 { // Slow operations (>500ms)
				slowOpsCount++
			} else if duration < 100 { // Fast operations (<100ms)
				fastOpsCount++
			}
		}

		if totalOpsAnalyzed > 0 {
			slowRatio := float64(slowOpsCount) / float64(totalOpsAnalyzed)
			fastRatio := float64(fastOpsCount) / float64(totalOpsAnalyzed)

			// More slow ops = performance degrading (positive change)
			// More fast ops = performance improving (negative change)
			avgDurationChange = (slowRatio - fastRatio) * 20.0 // Scale to percentage
		}

		// Adjust based on absolute duration
		if avgDuration > 300 {
			// System is already slow: likely worsening
			avgDurationChange += 5.0
		} else if avgDuration < 100 {
			// System is fast: likely improving or stable
			avgDurationChange -= 3.0
		}
	}

	// Determine overall trend direction and confidence
	var trend string
	confidence := 0.5

	trendScore := 0.0
	if operationsChange != 0 {
		trendScore += operationsChange * 0.4 // 40% weight
		confidence += 0.15
	}
	if tokenUsageChange != 0 {
		trendScore += tokenUsageChange * 0.3 // 30% weight
		confidence += 0.15
	}
	if avgDurationChange != 0 {
		trendScore += avgDurationChange * 0.3 // 30% weight
		confidence += 0.15
	}

	// Classify trend
	switch {
	case trendScore > 5.0:
		trend = "increasing"
	case trendScore < -5.0:
		trend = "decreasing"
	default:
		trend = "stable"
	}

	// Cap confidence at 0.95 (never 100% certain without more data)
	if confidence > 0.95 {
		confidence = 0.95
	}

	return CostTrends{
		Period:            period,
		OperationsChange:  operationsChange,
		TokenUsageChange:  tokenUsageChange,
		AvgDurationChange: avgDurationChange,
		Trend:             trend,
		TrendConfidence:   confidence,
		PeakUsageTime:     s.identifyPeakUsageTime(perfStats),
		LowUsageTime:      s.identifyLowUsageTime(perfStats),
	}
}

// generateOptimizationOpportunities identifies cost optimization opportunities.
func (s *MCPServer) generateOptimizationOpportunities(perfStats *application.UsageStatistics, tokenStats application.DetailedTokenStats, toolCosts []ToolCostBreakdown) []OptimizationOpportunity {
	opportunities := []OptimizationOpportunity{}

	// 1. Identify tools with high token usage but low compression
	for _, cost := range toolCosts {
		if cost.TotalTokens > 10000 && cost.CompressionRatio < 0.3 {
			opportunities = append(opportunities, OptimizationOpportunity{
				Type:             "compression",
				ToolName:         cost.ToolName,
				Description:      fmt.Sprintf("%s has high token usage (%d tokens) but low compression ratio (%.1f%%). Consider enabling aggressive compression.", cost.ToolName, cost.TotalTokens, cost.CompressionRatio*100),
				Severity:         "medium",
				PotentialSavings: fmt.Sprintf("Up to %d tokens", int(float64(cost.TotalTokens)*0.3)),
			})
		}
	}

	// 2. Identify slow tools with high frequency
	for _, cost := range toolCosts {
		if cost.AvgDuration > 500 && cost.OperationCount > 100 {
			opportunities = append(opportunities, OptimizationOpportunity{
				Type:             "performance",
				ToolName:         cost.ToolName,
				Description:      fmt.Sprintf("%s has high execution time (%.0fms avg) and high frequency (%d ops). Consider caching or optimization.", cost.ToolName, cost.AvgDuration, cost.OperationCount),
				Severity:         "high",
				PotentialSavings: fmt.Sprintf("%.0fms per operation", cost.AvgDuration*0.5),
			})
		}
	}

	// 3. Identify tools with low success rates
	for _, cost := range toolCosts {
		if cost.SuccessRate < 0.95 && cost.OperationCount > 50 {
			opportunities = append(opportunities, OptimizationOpportunity{
				Type:             "reliability",
				ToolName:         cost.ToolName,
				Description:      fmt.Sprintf("%s has low success rate (%.1f%%) with %d operations. Investigate error causes.", cost.ToolName, cost.SuccessRate*100, cost.OperationCount),
				Severity:         "high",
				PotentialSavings: fmt.Sprintf("%.0f%% operation success improvement", (1.0-cost.SuccessRate)*100),
			})
		}
	}

	// 4. Identify underutilized optimizations
	if tokenStats.TotalOriginalTokens > 0 && float64(tokenStats.TotalOptimizedTokens) < float64(tokenStats.TotalOriginalTokens)*0.5 {
		opportunities = append(opportunities, OptimizationOpportunity{
			Type:             "configuration",
			ToolName:         "system",
			Description:      fmt.Sprintf("Token optimization is underutilized. Only %.1f%% of tokens are being optimized. Consider enabling compression for more tools.", float64(tokenStats.TotalOptimizedTokens)/float64(tokenStats.TotalOriginalTokens)*100),
			Severity:         "medium",
			PotentialSavings: fmt.Sprintf("%d tokens", tokenStats.TotalOriginalTokens-tokenStats.TotalOptimizedTokens),
		})
	}

	return opportunities
}

// generateCostRecommendations generates actionable recommendations.
func (s *MCPServer) generateCostRecommendations(perfStats *application.UsageStatistics, tokenStats application.DetailedTokenStats, toolCosts []ToolCostBreakdown, trends CostTrends) []string {
	recommendations := []string{}

	// Recommendation 1: Overall token optimization
	if tokenStats.TotalOriginalTokens > 0 {
		savingsPercent := float64(tokenStats.TotalOriginalTokens-tokenStats.TotalOptimizedTokens) / float64(tokenStats.TotalOriginalTokens) * 100
		if savingsPercent < 30 {
			recommendations = append(recommendations, fmt.Sprintf("⚡ Enable compression for more tools. Current savings: %.1f%%, target: 30%%+", savingsPercent))
		}
	}

	// Recommendation 2: Top expensive tool optimization
	if len(toolCosts) > 0 {
		topTool := toolCosts[0]
		recommendations = append(recommendations, fmt.Sprintf("🎯 Focus optimization on '%s' (cost score: %.1f). It accounts for the highest resource consumption.", topTool.ToolName, topTool.CostScore))
	}

	// Recommendation 3: Success rate improvements
	if perfStats.SuccessRate < 0.98 {
		recommendations = append(recommendations, fmt.Sprintf("🔧 Investigate failures. Current success rate: %.1f%%, target: 98%%+", perfStats.SuccessRate*100))
	}

	// Recommendation 4: Trend-based recommendations
	if trends.Trend == "increasing" {
		recommendations = append(recommendations, "📈 Usage is increasing. Consider scaling infrastructure or implementing rate limiting.")
	}

	// Recommendation 5: Performance optimization
	avgDuration := s.calculateAverageDuration(perfStats)
	if avgDuration > 200 {
		recommendations = append(recommendations, fmt.Sprintf("⏱️ Average operation time is high (%.0fms). Consider caching, indexing, or parallel processing.", avgDuration))
	}

	return recommendations
}

// detectCostAnomalies detects unusual patterns in cost metrics.
func (s *MCPServer) detectCostAnomalies(perfStats *application.UsageStatistics, tokenStats application.DetailedTokenStats) []CostAnomaly {
	anomalies := []CostAnomaly{}

	// Anomaly 1: Unusually high error rate
	if perfStats.FailedOps > perfStats.SuccessfulOps/10 { // >10% error rate
		anomalies = append(anomalies, CostAnomaly{
			Type:        "high_error_rate",
			Description: fmt.Sprintf("High error rate detected: %d failures out of %d operations (%.1f%%)", perfStats.FailedOps, perfStats.TotalOperations, float64(perfStats.FailedOps)/float64(perfStats.TotalOperations)*100),
			Severity:    "high",
			DetectedAt:  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Anomaly 2: Unusually low compression
	if tokenStats.TotalOriginalTokens > 10000 && float64(tokenStats.TotalOptimizedTokens) > float64(tokenStats.TotalOriginalTokens)*0.9 {
		anomalies = append(anomalies, CostAnomaly{
			Type:        "low_compression",
			Description: fmt.Sprintf("Compression is ineffective. Only %.1f%% reduction from %d tokens.", (1.0-float64(tokenStats.TotalOptimizedTokens)/float64(tokenStats.TotalOriginalTokens))*100, tokenStats.TotalOriginalTokens),
			Severity:    "medium",
			DetectedAt:  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Anomaly 3: Slow operations detected
	slowOps := 0
	for _, duration := range perfStats.AvgDurationByTool {
		if duration > 1000 { // >1 second
			slowOps++
		}
	}
	if slowOps > 0 {
		anomalies = append(anomalies, CostAnomaly{
			Type:        "slow_operations",
			Description: fmt.Sprintf("Detected %d tools with average execution time >1 second. Review performance bottlenecks.", slowOps),
			Severity:    "medium",
			DetectedAt:  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return anomalies
}

// projectFutureCosts projects future costs based on trends.
func (s *MCPServer) projectFutureCosts(trends CostTrends, currentPeriod string) CostProjections {
	// Simplified projection model
	// In production, this would use time-series forecasting (ARIMA, Prophet, etc.)

	growthRate := 0.0
	switch trends.Trend {
	case "increasing":
		growthRate = 0.15 // 15% growth assumption
	case "decreasing":
		growthRate = -0.10 // 10% decrease assumption
	}

	return CostProjections{
		NextDay:    fmt.Sprintf("Projected %.1f%% change", growthRate*100),
		NextWeek:   fmt.Sprintf("Projected %.1f%% change", growthRate*7*100),
		NextMonth:  fmt.Sprintf("Projected %.1f%% change", growthRate*30*100),
		Confidence: trends.TrendConfidence,
		Model:      "linear_trend",
	}
}

// Helper functions

func (s *MCPServer) calculatePercentage(value, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total) * 100
}

func (s *MCPServer) calculateTotalDuration(stats *application.UsageStatistics) float64 {
	total := 0.0
	for toolName, count := range stats.OperationsByTool {
		total += float64(count) * stats.AvgDurationByTool[toolName]
	}
	return total
}

func (s *MCPServer) calculateAverageDuration(stats *application.UsageStatistics) float64 {
	if stats.TotalOperations == 0 {
		return 0
	}
	return s.calculateTotalDuration(stats) / float64(stats.TotalOperations)
}

func (s *MCPServer) calculateToolSuccessRate(toolName string, stats *application.UsageStatistics) float64 {
	totalOps := stats.OperationsByTool[toolName]
	errors := stats.ErrorsByTool[toolName]
	if totalOps == 0 {
		return 1.0
	}
	return float64(totalOps-errors) / float64(totalOps)
}

func (s *MCPServer) calculateCompressionRatio(toolName string, stats application.DetailedTokenStats) float64 {
	original := stats.OriginalTokensByTool[toolName]
	optimized := stats.OptimizedTokensByTool[toolName]
	if original == 0 {
		return 0
	}
	return float64(original-optimized) / float64(original)
}

func (s *MCPServer) identifyPeakUsageTime(stats *application.UsageStatistics) string {
	// Simplified - in production, analyze time-series data
	return "14:00-16:00 UTC" // Placeholder
}

func (s *MCPServer) identifyLowUsageTime(stats *application.UsageStatistics) string {
	// Simplified - in production, analyze time-series data
	return "02:00-04:00 UTC" // Placeholder
}
