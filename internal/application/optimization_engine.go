package application

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// RecommendationType categorizes different types of recommendations.
type RecommendationType string

const (
	RecommendationTypePerformance   RecommendationType = "performance"
	RecommendationTypeTokens        RecommendationType = "tokens"
	RecommendationTypeReliability   RecommendationType = "reliability"
	RecommendationTypeArchitecture  RecommendationType = "architecture"
	RecommendationTypeCost          RecommendationType = "cost"
	RecommendationTypeConfiguration RecommendationType = "configuration"
)

// RecommendationPriority indicates urgency of a recommendation.
type RecommendationPriority string

const (
	PriorityCritical RecommendationPriority = "critical"
	PriorityHigh     RecommendationPriority = "high"
	PriorityMedium   RecommendationPriority = "medium"
	PriorityLow      RecommendationPriority = "low"
)

// OptimizationRecommendation represents an optimization suggestion.
type OptimizationRecommendation struct {
	ID                   string                 `json:"id"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	Type                 RecommendationType     `json:"type"`
	Priority             RecommendationPriority `json:"priority"`
	ImpactScore          float64                `json:"impact_score"`          // 0-100, higher is more impactful
	ImplementationEffort string                 `json:"implementation_effort"` // low, medium, high
	AffectedTools        []string               `json:"affected_tools,omitempty"`
	EstimatedSavings     string                 `json:"estimated_savings,omitempty"`
	ActionItems          []string               `json:"action_items"`
	Evidence             map[string]interface{} `json:"evidence"`
	GeneratedAt          time.Time              `json:"generated_at"`
}

// OptimizationEngineConfig configures the recommendation engine.
type OptimizationEngineConfig struct {
	// Thresholds for generating recommendations
	SlowToolThresholdMs      float64
	HighTokenUsageThreshold  int
	LowSuccessRateThreshold  float64
	PoorCompressionThreshold float64
	MinOperationsForAnalysis int

	// Weights for impact scoring
	PerformanceWeight float64
	TokenWeight       float64
	ReliabilityWeight float64
}

// DefaultOptimizationEngineConfig returns default configuration.
func DefaultOptimizationEngineConfig() OptimizationEngineConfig {
	return OptimizationEngineConfig{
		SlowToolThresholdMs:      500.0,
		HighTokenUsageThreshold:  10000,
		LowSuccessRateThreshold:  0.95,
		PoorCompressionThreshold: 0.7,
		MinOperationsForAnalysis: 10,
		PerformanceWeight:        0.4,
		TokenWeight:              0.3,
		ReliabilityWeight:        0.3,
	}
}

// OptimizationEngine generates intelligent optimization recommendations.
type OptimizationEngine struct {
	config OptimizationEngineConfig
}

// NewOptimizationEngine creates a new optimization engine.
func NewOptimizationEngine(config OptimizationEngineConfig) *OptimizationEngine {
	return &OptimizationEngine{
		config: config,
	}
}

// GenerateRecommendations analyzes metrics and generates prioritized recommendations.
func (e *OptimizationEngine) GenerateRecommendations(perfStats *UsageStatistics, tokenStats DetailedTokenStats) []OptimizationRecommendation {
	recommendations := []OptimizationRecommendation{}

	// Performance recommendations
	recommendations = append(recommendations, e.analyzePerformance(perfStats)...)

	// Token optimization recommendations
	recommendations = append(recommendations, e.analyzeTokenUsage(tokenStats)...)

	// Reliability recommendations
	recommendations = append(recommendations, e.analyzeReliability(perfStats)...)

	// Architecture recommendations
	recommendations = append(recommendations, e.analyzeArchitecture(perfStats, tokenStats)...)

	// Cost optimization recommendations
	recommendations = append(recommendations, e.analyzeCostOptimization(perfStats, tokenStats)...)

	// Sort by priority and impact score
	sort.Slice(recommendations, func(i, j int) bool {
		priorityOrder := map[RecommendationPriority]int{
			PriorityCritical: 0,
			PriorityHigh:     1,
			PriorityMedium:   2,
			PriorityLow:      3,
		}

		if priorityOrder[recommendations[i].Priority] != priorityOrder[recommendations[j].Priority] {
			return priorityOrder[recommendations[i].Priority] < priorityOrder[recommendations[j].Priority]
		}

		return recommendations[i].ImpactScore > recommendations[j].ImpactScore
	})

	return recommendations
}

// analyzePerformance generates performance optimization recommendations.
func (e *OptimizationEngine) analyzePerformance(stats *UsageStatistics) []OptimizationRecommendation {
	recommendations := []OptimizationRecommendation{}

	if stats == nil || stats.TotalOperations < e.config.MinOperationsForAnalysis {
		return recommendations
	}

	// Find slow tools
	slowTools := []struct {
		name     string
		duration float64
		count    int
	}{}

	for toolName, avgDuration := range stats.AvgDurationByTool {
		if avgDuration > e.config.SlowToolThresholdMs {
			slowTools = append(slowTools, struct {
				name     string
				duration float64
				count    int
			}{
				name:     toolName,
				duration: avgDuration,
				count:    stats.OperationsByTool[toolName],
			})
		}
	}

	// Sort by impact (duration * frequency)
	sort.Slice(slowTools, func(i, j int) bool {
		impactI := slowTools[i].duration * float64(slowTools[i].count)
		impactJ := slowTools[j].duration * float64(slowTools[j].count)
		return impactI > impactJ
	})

	// Generate recommendations for top slow tools
	for i, tool := range slowTools {
		if i >= 5 { // Limit to top 5
			break
		}

		totalTime := tool.duration * float64(tool.count)
		// Calculate total duration across all tools
		totalDuration := 0.0
		for _, avgDur := range stats.AvgDurationByTool {
			totalDuration += avgDur
		}
		impactScore := e.calculateImpactScore(totalTime, totalDuration, e.config.PerformanceWeight)
		priority := PriorityMedium
		if tool.duration > 2000 { // >2s
			priority = PriorityHigh
		}
		if tool.duration > 5000 { // >5s
			priority = PriorityCritical
		}

		rec := OptimizationRecommendation{
			ID:                   fmt.Sprintf("perf-slow-tool-%d", i+1),
			Title:                fmt.Sprintf("Optimize '%s' Performance", tool.name),
			Description:          fmt.Sprintf("Tool '%s' has an average latency of %.0fms across %d operations, significantly impacting overall performance.", tool.name, tool.duration, tool.count),
			Type:                 RecommendationTypePerformance,
			Priority:             priority,
			ImpactScore:          impactScore,
			ImplementationEffort: e.estimateEffort(tool.duration),
			AffectedTools:        []string{tool.name},
			EstimatedSavings:     fmt.Sprintf("%.0fms total latency reduction", totalTime*0.5), // Assume 50% improvement
			ActionItems: []string{
				"Profile tool execution to identify bottlenecks",
				"Add caching for frequently accessed data",
				"Optimize database queries or API calls",
				"Consider parallel processing where applicable",
				"Review algorithm complexity",
			},
			Evidence: map[string]interface{}{
				"avg_duration_ms": tool.duration,
				"operation_count": tool.count,
				"total_time_ms":   totalTime,
				"threshold_ms":    e.config.SlowToolThresholdMs,
			},
			GeneratedAt: time.Now(),
		}

		recommendations = append(recommendations, rec)
	}

	// Check for overall high P95 latency
	if len(stats.AvgDurationByTool) > 0 {
		maxDuration := 0.0
		for _, dur := range stats.AvgDurationByTool {
			if dur > maxDuration {
				maxDuration = dur
			}
		}

		if maxDuration > 1000 { // P95 > 1s
			rec := OptimizationRecommendation{
				ID:                   "perf-high-p95",
				Title:                "High P95 Latency Detected",
				Description:          fmt.Sprintf("The 95th percentile latency is approximately %.0fms, indicating performance issues affecting user experience.", maxDuration),
				Type:                 RecommendationTypePerformance,
				Priority:             PriorityHigh,
				ImpactScore:          75.0,
				ImplementationEffort: "medium",
				EstimatedSavings:     "30-50% latency reduction",
				ActionItems: []string{
					"Implement comprehensive performance monitoring",
					"Add request timeout policies",
					"Optimize slow endpoints identified above",
					"Consider adding circuit breakers for failing operations",
					"Review infrastructure scaling policies",
				},
				Evidence: map[string]interface{}{
					"p95_duration_ms": maxDuration,
					"threshold_ms":    1000.0,
				},
				GeneratedAt: time.Now(),
			}
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations
}

// analyzeTokenUsage generates token optimization recommendations.
func (e *OptimizationEngine) analyzeTokenUsage(tokenStats DetailedTokenStats) []OptimizationRecommendation {
	recommendations := []OptimizationRecommendation{}

	if tokenStats.TotalOriginalTokens == 0 {
		return recommendations
	}

	// Find tools with high token usage
	type tokenUser struct {
		tool     string
		original int
		savings  int
		ratio    float64
	}

	heavyUsers := []tokenUser{}
	for tool, originalTokens := range tokenStats.OriginalTokensByTool {
		if originalTokens > e.config.HighTokenUsageThreshold {
			optimizedTokens := tokenStats.OptimizedTokensByTool[tool]
			savings := originalTokens - optimizedTokens
			ratio := float64(optimizedTokens) / float64(originalTokens)

			heavyUsers = append(heavyUsers, tokenUser{
				tool:     tool,
				original: originalTokens,
				savings:  savings,
				ratio:    ratio,
			})
		}
	}

	// Sort by original tokens (highest first)
	sort.Slice(heavyUsers, func(i, j int) bool {
		return heavyUsers[i].original > heavyUsers[j].original
	})

	// Generate recommendations for high token users
	for i, user := range heavyUsers {
		if i >= 5 { // Limit to top 5
			break
		}

		impactScore := e.calculateImpactScore(float64(user.original), float64(tokenStats.TotalOriginalTokens), e.config.TokenWeight)

		priority := PriorityMedium
		if user.original > 50000 {
			priority = PriorityHigh
		}

		rec := OptimizationRecommendation{
			ID:                   fmt.Sprintf("token-heavy-user-%d", i+1),
			Title:                fmt.Sprintf("Reduce Token Usage in '%s'", user.tool),
			Description:          fmt.Sprintf("Tool '%s' consumes %d tokens (%.1f%% of total), with %.1f%% compression ratio.", user.tool, user.original, float64(user.original)/float64(tokenStats.TotalOriginalTokens)*100, user.ratio*100),
			Type:                 RecommendationTypeTokens,
			Priority:             priority,
			ImpactScore:          impactScore,
			ImplementationEffort: "medium",
			AffectedTools:        []string{user.tool},
			EstimatedSavings:     fmt.Sprintf("%d tokens (%.1f%% reduction)", int(float64(user.original)*0.3), 30.0),
			ActionItems: []string{
				"Implement aggressive response compression",
				"Use summarization for large responses",
				"Paginate large result sets",
				"Remove unnecessary fields from responses",
				"Enable adaptive compression based on content size",
			},
			Evidence: map[string]interface{}{
				"original_tokens":   user.original,
				"optimized_tokens":  user.original - user.savings,
				"compression_ratio": user.ratio,
				"threshold":         e.config.HighTokenUsageThreshold,
			},
			GeneratedAt: time.Now(),
		}

		recommendations = append(recommendations, rec)
	}

	// Check overall compression effectiveness
	overallRatio := float64(tokenStats.TotalOptimizedTokens) / float64(tokenStats.TotalOriginalTokens)
	if overallRatio > e.config.PoorCompressionThreshold {
		rec := OptimizationRecommendation{
			ID:                   "token-poor-compression",
			Title:                "Improve Overall Token Compression",
			Description:          fmt.Sprintf("Current compression ratio is %.1f%%, indicating opportunities for better optimization. Target: <70%%.", overallRatio*100),
			Type:                 RecommendationTypeTokens,
			Priority:             PriorityHigh,
			ImpactScore:          80.0,
			ImplementationEffort: "low",
			EstimatedSavings:     fmt.Sprintf("%d tokens annually", int(float64(tokenStats.TotalOriginalTokens)*0.15)),
			ActionItems: []string{
				"Enable advanced compression algorithms (gzip, brotli)",
				"Implement semantic deduplication",
				"Use prompt compression for large contexts",
				"Configure adaptive cache with longer TTLs",
				"Review compression settings and increase compression level",
			},
			Evidence: map[string]interface{}{
				"compression_ratio": overallRatio,
				"total_original":    tokenStats.TotalOriginalTokens,
				"total_optimized":   tokenStats.TotalOptimizedTokens,
				"threshold":         e.config.PoorCompressionThreshold,
			},
			GeneratedAt: time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// analyzeReliability generates reliability improvement recommendations.
func (e *OptimizationEngine) analyzeReliability(stats *UsageStatistics) []OptimizationRecommendation {
	recommendations := []OptimizationRecommendation{}

	if stats == nil || stats.TotalOperations < e.config.MinOperationsForAnalysis {
		return recommendations
	}

	// Find unreliable tools
	unreliableTools := []struct {
		name        string
		successRate float64
		errorCount  int
		totalCount  int
	}{}

	for toolName, count := range stats.OperationsByTool {
		errors := stats.ErrorsByTool[toolName]
		successRate := float64(count-errors) / float64(count)

		if successRate < e.config.LowSuccessRateThreshold && count >= e.config.MinOperationsForAnalysis {
			unreliableTools = append(unreliableTools, struct {
				name        string
				successRate float64
				errorCount  int
				totalCount  int
			}{
				name:        toolName,
				successRate: successRate,
				errorCount:  errors,
				totalCount:  count,
			})
		}
	}

	// Sort by error count (highest first)
	sort.Slice(unreliableTools, func(i, j int) bool {
		return unreliableTools[i].errorCount > unreliableTools[j].errorCount
	})

	// Generate recommendations for unreliable tools
	for i, tool := range unreliableTools {
		if i >= 5 { // Limit to top 5
			break
		}

		impactScore := e.calculateImpactScore(float64(tool.errorCount), float64(stats.TotalOperations), e.config.ReliabilityWeight)

		priority := PriorityMedium
		if tool.successRate < 0.90 { // <90%
			priority = PriorityHigh
		}
		if tool.successRate < 0.80 { // <80%
			priority = PriorityCritical
		}

		rec := OptimizationRecommendation{
			ID:                   fmt.Sprintf("reliability-low-success-%d", i+1),
			Title:                fmt.Sprintf("Improve Reliability of '%s'", tool.name),
			Description:          fmt.Sprintf("Tool '%s' has a success rate of %.1f%% (%d errors out of %d operations), impacting overall system reliability.", tool.name, tool.successRate*100, tool.errorCount, tool.totalCount),
			Type:                 RecommendationTypeReliability,
			Priority:             priority,
			ImpactScore:          impactScore,
			ImplementationEffort: "medium",
			AffectedTools:        []string{tool.name},
			EstimatedSavings:     fmt.Sprintf("%.0f%% error reduction", (1.0-tool.successRate)*100*0.7), // Assume 70% improvement
			ActionItems: []string{
				"Review error logs to identify root causes",
				"Add comprehensive error handling and retry logic",
				"Implement circuit breakers for external dependencies",
				"Add input validation to prevent common errors",
				"Set up alerting for elevated error rates",
				"Consider fallback strategies for critical operations",
			},
			Evidence: map[string]interface{}{
				"success_rate":     tool.successRate,
				"error_count":      tool.errorCount,
				"total_operations": tool.totalCount,
				"threshold":        e.config.LowSuccessRateThreshold,
			},
			GeneratedAt: time.Now(),
		}

		recommendations = append(recommendations, rec)
	}

	// Check overall success rate
	if stats.SuccessRate < e.config.LowSuccessRateThreshold {
		rec := OptimizationRecommendation{
			ID:                   "reliability-low-overall",
			Title:                "Improve Overall System Reliability",
			Description:          fmt.Sprintf("Overall success rate is %.1f%%, below the target of %.0f%%. This affects user trust and system stability.", stats.SuccessRate*100, e.config.LowSuccessRateThreshold*100),
			Type:                 RecommendationTypeReliability,
			Priority:             PriorityCritical,
			ImpactScore:          90.0,
			ImplementationEffort: "high",
			EstimatedSavings:     "Improved user experience and reduced support costs",
			ActionItems: []string{
				"Implement comprehensive monitoring and alerting",
				"Add health checks for all critical dependencies",
				"Implement graceful degradation strategies",
				"Set up automated recovery procedures",
				"Review and improve error handling across all tools",
				"Conduct root cause analysis on recent failures",
			},
			Evidence: map[string]interface{}{
				"success_rate":     stats.SuccessRate,
				"total_operations": stats.TotalOperations,
				"threshold":        e.config.LowSuccessRateThreshold,
			},
			GeneratedAt: time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// analyzeArchitecture generates architecture improvement recommendations.
func (e *OptimizationEngine) analyzeArchitecture(perfStats *UsageStatistics, tokenStats DetailedTokenStats) []OptimizationRecommendation {
	recommendations := []OptimizationRecommendation{}

	if perfStats == nil || perfStats.TotalOperations < e.config.MinOperationsForAnalysis {
		return recommendations
	}

	// Identify high-traffic tools for caching opportunities
	type toolTraffic struct {
		name  string
		count int
	}

	highTraffic := []toolTraffic{}
	avgOps := perfStats.TotalOperations / len(perfStats.OperationsByTool)

	for toolName, count := range perfStats.OperationsByTool {
		if count > avgOps*2 { // More than 2x average
			highTraffic = append(highTraffic, toolTraffic{
				name:  toolName,
				count: count,
			})
		}
	}

	sort.Slice(highTraffic, func(i, j int) bool {
		return highTraffic[i].count > highTraffic[j].count
	})

	// Recommend caching for high-traffic tools
	if len(highTraffic) > 0 {
		affectedTools := []string{}
		for i, ht := range highTraffic {
			if i >= 3 {
				break
			}
			affectedTools = append(affectedTools, ht.name)
		}

		rec := OptimizationRecommendation{
			ID:                   "arch-implement-caching",
			Title:                "Implement Caching for High-Traffic Tools",
			Description:          fmt.Sprintf("%d tools have significantly higher traffic than average. Implementing caching could reduce load and improve response times.", len(highTraffic)),
			Type:                 RecommendationTypeArchitecture,
			Priority:             PriorityHigh,
			ImpactScore:          85.0,
			ImplementationEffort: "medium",
			AffectedTools:        affectedTools,
			EstimatedSavings:     fmt.Sprintf("30-50%% latency reduction on %d operations", highTraffic[0].count),
			ActionItems: []string{
				"Implement Redis or in-memory caching layer",
				"Define appropriate cache TTL policies",
				"Add cache invalidation strategies",
				"Monitor cache hit rates",
				"Consider CDN for static responses",
			},
			Evidence: map[string]interface{}{
				"high_traffic_tools": len(highTraffic),
				"top_tool":           highTraffic[0].name,
				"top_tool_ops":       highTraffic[0].count,
				"avg_ops":            avgOps,
			},
			GeneratedAt: time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// analyzeCostOptimization generates cost-focused recommendations.
func (e *OptimizationEngine) analyzeCostOptimization(perfStats *UsageStatistics, tokenStats DetailedTokenStats) []OptimizationRecommendation {
	recommendations := []OptimizationRecommendation{}

	// Overall cost optimization recommendation
	if perfStats != nil && perfStats.TotalOperations > 1000 {
		// Calculate average duration across all tools
		avgDuration := 0.0
		if len(perfStats.AvgDurationByTool) > 0 {
			for _, dur := range perfStats.AvgDurationByTool {
				avgDuration += dur
			}
			avgDuration /= float64(len(perfStats.AvgDurationByTool))
		}
		totalCostScore := (avgDuration * float64(perfStats.TotalOperations)) + float64(tokenStats.TotalOriginalTokens)

		rec := OptimizationRecommendation{
			ID:                   "cost-overall-optimization",
			Title:                "Comprehensive Cost Optimization Strategy",
			Description:          "Implement a holistic approach to reduce operational costs by optimizing performance, token usage, and infrastructure efficiency.",
			Type:                 RecommendationTypeCost,
			Priority:             PriorityHigh,
			ImpactScore:          88.0,
			ImplementationEffort: "high",
			EstimatedSavings:     "20-30% total operational cost reduction",
			ActionItems: []string{
				"Implement request batching to reduce operation count",
				"Use connection pooling for database/API calls",
				"Enable all token optimization features",
				"Review and optimize resource allocation",
				"Implement auto-scaling policies",
				"Monitor and eliminate unused resources",
				"Consider serverless architecture for variable loads",
			},
			Evidence: map[string]interface{}{
				"total_operations": perfStats.TotalOperations,
				"total_tokens":     tokenStats.TotalOriginalTokens,
				"cost_score":       totalCostScore,
			},
			GeneratedAt: time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// calculateImpactScore calculates impact score (0-100) based on contribution to total.
func (e *OptimizationEngine) calculateImpactScore(value, total, weight float64) float64 {
	if total == 0 {
		return 0
	}

	ratio := value / total
	score := ratio * 100 * weight * 2.5 // Amplify for 0-100 scale

	return math.Min(score, 100.0)
}

// estimateEffort estimates implementation effort based on performance impact.
func (e *OptimizationEngine) estimateEffort(durationMs float64) string {
	if durationMs > 5000 {
		return "high" // Severe performance issues likely require architectural changes
	} else if durationMs > 1000 {
		return "medium"
	}
	return "low"
}
