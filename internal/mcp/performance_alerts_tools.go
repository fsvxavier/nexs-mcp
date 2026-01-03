package mcp

import (
	"context"
	"crypto/md5" // #nosec G501 -- MD5 used for alert ID generation, not cryptography
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
)

// AlertManager manages alert state and history.
type AlertManager struct {
	mu              sync.RWMutex
	alerts          map[string]*Alert
	rules           map[string]*AlertRule
	lastTriggerTime map[string]time.Time
	storageDir      string
}

// NewAlertManager creates a new alert manager.
func NewAlertManager(storageDir string) *AlertManager {
	am := &AlertManager{
		alerts:          make(map[string]*Alert),
		rules:           make(map[string]*AlertRule),
		lastTriggerTime: make(map[string]time.Time),
		storageDir:      filepath.Join(storageDir, "alerts"),
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(am.storageDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create alerts directory: %v\n", err)
	}

	// Load existing alerts and rules
	if err := am.loadAlerts(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load alerts: %v\n", err)
	}
	if err := am.loadRules(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load rules: %v\n", err)
	}

	// Initialize default rules if none exist
	if len(am.rules) == 0 {
		am.initializeDefaultRules()
	}

	return am
}

// initializeDefaultRules sets up default alert rules.
func (am *AlertManager) initializeDefaultRules() {
	defaultRules := []*AlertRule{
		{
			ID:          "high_error_rate",
			Name:        "High Error Rate",
			Description: "Triggers when error rate exceeds threshold",
			Type:        "error_rate",
			Metric:      "error_rate_percent",
			Operator:    "gt",
			Threshold:   10.0,
			Severity:    "critical",
			Enabled:     true,
			Cooldown:    15,
		},
		{
			ID:          "low_success_rate",
			Name:        "Low Success Rate",
			Description: "Triggers when success rate drops below threshold",
			Type:        "performance",
			Metric:      "success_rate_percent",
			Operator:    "lt",
			Threshold:   95.0,
			Severity:    "warning",
			Enabled:     true,
			Cooldown:    30,
		},
		{
			ID:          "high_p95_latency",
			Name:        "High P95 Latency",
			Description: "Triggers when P95 latency exceeds threshold",
			Type:        "latency",
			Metric:      "p95_duration_ms",
			Operator:    "gt",
			Threshold:   1000.0,
			Severity:    "warning",
			Enabled:     true,
			Cooldown:    30,
		},
		{
			ID:          "slow_average_latency",
			Name:        "Slow Average Latency",
			Description: "Triggers when average latency is too high",
			Type:        "latency",
			Metric:      "avg_duration_ms",
			Operator:    "gt",
			Threshold:   500.0,
			Severity:    "warning",
			Enabled:     true,
			Cooldown:    60,
		},
		{
			ID:          "poor_token_compression",
			Name:        "Poor Token Compression",
			Description: "Triggers when token compression ratio is too low",
			Type:        "token_usage",
			Metric:      "compression_ratio",
			Operator:    "gt",
			Threshold:   0.8,
			Severity:    "info",
			Enabled:     true,
			Cooldown:    120,
		},
		{
			ID:          "excessive_token_usage",
			Name:        "Excessive Token Usage",
			Description: "Triggers when token usage is unusually high",
			Type:        "token_usage",
			Metric:      "total_tokens",
			Operator:    "gt",
			Threshold:   100000.0,
			Severity:    "warning",
			Enabled:     true,
			Cooldown:    60,
		},
	}

	for _, rule := range defaultRules {
		am.rules[rule.ID] = rule
	}
	if err := am.saveRules(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save default rules: %v\n", err)
	}
}

// EvaluateMetrics evaluates current metrics against alert rules.
func (am *AlertManager) EvaluateMetrics(perfStats *application.UsageStatistics, tokenStats application.DetailedTokenStats) []*Alert {
	am.mu.Lock()
	defer am.mu.Unlock()

	triggeredAlerts := []*Alert{}
	now := time.Now()

	// Extract metrics
	metrics := am.extractMetrics(perfStats, tokenStats)

	// Evaluate each enabled rule
	for _, rule := range am.rules {
		if !rule.Enabled {
			continue
		}

		// Check cooldown
		if lastTrigger, exists := am.lastTriggerTime[rule.ID]; exists {
			if now.Sub(lastTrigger).Minutes() < float64(rule.Cooldown) {
				continue // Still in cooldown
			}
		}

		// Get metric value
		metricValue, exists := metrics[rule.Metric]
		if !exists {
			continue
		}

		// Evaluate condition
		triggered := false
		switch rule.Operator {
		case "gt":
			triggered = metricValue > rule.Threshold
		case "lt":
			triggered = metricValue < rule.Threshold
		case "gte":
			triggered = metricValue >= rule.Threshold
		case "lte":
			triggered = metricValue <= rule.Threshold
		case "eq":
			triggered = metricValue == rule.Threshold
		}

		if triggered {
			// Create alert
			alert := &Alert{
				ID:                fmt.Sprintf("%s_%d", rule.ID, now.Unix()),
				RuleID:            rule.ID,
				RuleName:          rule.Name,
				Severity:          rule.Severity,
				Title:             rule.Name,
				Message:           am.formatAlertMessage(rule, metricValue),
				Value:             metricValue,
				Threshold:         rule.Threshold,
				TriggeredAt:       now,
				Status:            "active",
				RecommendedAction: am.getRecommendedAction(rule),
				Fingerprint:       am.generateFingerprint(rule.ID, metricValue),
			}

			// Add affected tools for performance/error alerts
			if rule.Type == "performance" || rule.Type == "error_rate" {
				alert.AffectedTools = am.identifyAffectedTools(perfStats, rule)
			}

			am.alerts[alert.ID] = alert
			am.lastTriggerTime[rule.ID] = now
			triggeredAlerts = append(triggeredAlerts, alert)
		}
	}

	// Save updated alerts
	if len(triggeredAlerts) > 0 {
		if err := am.saveAlerts(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save alerts: %v\n", err)
		}
	}

	return triggeredAlerts
}

// extractMetrics extracts relevant metrics from stats.
func (am *AlertManager) extractMetrics(perfStats *application.UsageStatistics, tokenStats application.DetailedTokenStats) map[string]float64 {
	metrics := make(map[string]float64)

	if perfStats != nil {
		metrics["total_operations"] = float64(perfStats.TotalOperations)
		metrics["success_rate_percent"] = perfStats.SuccessRate * 100
		metrics["error_rate_percent"] = (1.0 - perfStats.SuccessRate) * 100

		// Calculate average duration
		totalDuration := 0.0
		for toolName, count := range perfStats.OperationsByTool {
			totalDuration += float64(count) * perfStats.AvgDurationByTool[toolName]
		}
		if perfStats.TotalOperations > 0 {
			metrics["avg_duration_ms"] = totalDuration / float64(perfStats.TotalOperations)
		}

		// Calculate P95 (simplified - use max avg duration as proxy)
		maxDuration := 0.0
		for _, duration := range perfStats.AvgDurationByTool {
			if duration > maxDuration {
				maxDuration = duration
			}
		}
		metrics["p95_duration_ms"] = maxDuration
	}

	if tokenStats.TotalOriginalTokens > 0 {
		metrics["total_tokens"] = float64(tokenStats.TotalOriginalTokens)
		metrics["optimized_tokens"] = float64(tokenStats.TotalOptimizedTokens)
		metrics["tokens_saved"] = float64(tokenStats.TotalOriginalTokens - tokenStats.TotalOptimizedTokens)
		metrics["compression_ratio"] = float64(tokenStats.TotalOptimizedTokens) / float64(tokenStats.TotalOriginalTokens)
	}

	return metrics
}

// formatAlertMessage formats a human-readable alert message.
func (am *AlertManager) formatAlertMessage(rule *AlertRule, value float64) string {
	switch rule.Type {
	case "error_rate":
		return fmt.Sprintf("Error rate is %.1f%%, exceeding threshold of %.1f%%", value, rule.Threshold)
	case "performance":
		return fmt.Sprintf("Success rate is %.1f%%, below threshold of %.1f%%", value, rule.Threshold)
	case "latency":
		return fmt.Sprintf("Latency is %.0fms, exceeding threshold of %.0fms", value, rule.Threshold)
	case "token_usage":
		if rule.Metric == "compression_ratio" {
			return fmt.Sprintf("Compression ratio is %.1f%%, indicating poor efficiency (threshold: %.1f%%)", value*100, rule.Threshold*100)
		}
		return fmt.Sprintf("Token usage is %.0f, exceeding threshold of %.0f", value, rule.Threshold)
	default:
		return fmt.Sprintf("%s: %.2f (threshold: %.2f)", rule.Name, value, rule.Threshold)
	}
}

// getRecommendedAction returns recommended action for a rule.
func (am *AlertManager) getRecommendedAction(rule *AlertRule) string {
	actions := map[string]string{
		"high_error_rate":        "Investigate error logs, check affected tools, verify external dependencies",
		"low_success_rate":       "Review recent changes, check system resources, analyze error patterns",
		"high_p95_latency":       "Profile slow operations, optimize database queries, add caching",
		"slow_average_latency":   "Review tool implementations, consider parallel processing, optimize algorithms",
		"poor_token_compression": "Review compression settings, analyze data structures, enable aggressive compression",
		"excessive_token_usage":  "Implement token limits, optimize response sizes, use summarization",
	}

	if action, exists := actions[rule.ID]; exists {
		return action
	}
	return "Review metrics and investigate root cause"
}

// identifyAffectedTools identifies which tools are contributing to the alert.
func (am *AlertManager) identifyAffectedTools(perfStats *application.UsageStatistics, rule *AlertRule) []string {
	affected := []string{}

	switch rule.Type {
	case "error_rate", "performance":
		for toolName := range perfStats.OperationsByTool {
			totalOps := perfStats.OperationsByTool[toolName]
			errors := perfStats.ErrorsByTool[toolName]

			if totalOps > 0 {
				errorRate := float64(errors) / float64(totalOps)
				if errorRate > 0.1 { // >10% error rate
					affected = append(affected, toolName)
				}
			}
		}
	case "latency":
		for toolName, duration := range perfStats.AvgDurationByTool {
			if duration > rule.Threshold {
				affected = append(affected, toolName)
			}
		}
	}

	return affected
}

// generateFingerprint creates a fingerprint for deduplication.
func (am *AlertManager) generateFingerprint(ruleID string, value float64) string {
	data := fmt.Sprintf("%s_%.2f", ruleID, value)
	hash := md5.Sum([]byte(data)) // #nosec G401 -- MD5 used for alert ID generation, not cryptography
	return hex.EncodeToString(hash[:8])
}

// ResolveAlert marks an alert as resolved.
func (am *AlertManager) ResolveAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.ResolvedAt = &now
	alert.Status = "resolved"

	if err := am.saveAlerts(); err != nil {
		return fmt.Errorf("failed to save alerts: %w", err)
	}
	return nil
}

// SilenceAlert silences an alert temporarily.
func (am *AlertManager) SilenceAlert(alertID string, duration time.Duration) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	alert.Status = "silenced"

	// Extend cooldown for the rule
	if rule, ruleExists := am.rules[alert.RuleID]; ruleExists {
		am.lastTriggerTime[rule.ID] = time.Now().Add(duration)
	}

	if err := am.saveAlerts(); err != nil {
		return fmt.Errorf("failed to save alerts: %w", err)
	}
	return nil
}

// GetActiveAlerts returns all active alerts.
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	active := []*Alert{}
	for _, alert := range am.alerts {
		if alert.Status == "active" {
			active = append(active, alert)
		}
	}

	// Sort by severity and time
	sort.Slice(active, func(i, j int) bool {
		severityOrder := map[string]int{"critical": 0, "warning": 1, "info": 2}
		if severityOrder[active[i].Severity] != severityOrder[active[j].Severity] {
			return severityOrder[active[i].Severity] < severityOrder[active[j].Severity]
		}
		return active[i].TriggeredAt.After(active[j].TriggeredAt)
	})

	return active
}

// GetAlertHistory returns alert history with optional filters.
func (am *AlertManager) GetAlertHistory(limit int, severity string, status string) []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	history := []*Alert{}
	for _, alert := range am.alerts {
		if severity != "" && alert.Severity != severity {
			continue
		}
		if status != "" && alert.Status != status {
			continue
		}
		history = append(history, alert)
	}

	// Sort by time descending
	sort.Slice(history, func(i, j int) bool {
		return history[i].TriggeredAt.After(history[j].TriggeredAt)
	})

	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	return history
}

// GetAlertRules returns all alert rules.
func (am *AlertManager) GetAlertRules() []*AlertRule {
	am.mu.RLock()
	defer am.mu.RUnlock()

	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		rules = append(rules, rule)
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Name < rules[j].Name
	})

	return rules
}

// UpdateAlertRule updates an alert rule.
func (am *AlertManager) UpdateAlertRule(rule *AlertRule) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if rule.ID == "" {
		return errors.New("rule ID is required")
	}

	am.rules[rule.ID] = rule
	if err := am.saveRules(); err != nil {
		return fmt.Errorf("failed to save rules: %w", err)
	}
	return nil
}

// Storage methods

func (am *AlertManager) saveAlerts() error {
	alertsFile := filepath.Join(am.storageDir, "alerts.json")

	alerts := make([]*Alert, 0, len(am.alerts))
	for _, alert := range am.alerts {
		alerts = append(alerts, alert)
	}

	data, err := json.MarshalIndent(alerts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal alerts: %w", err)
	}

	return os.WriteFile(alertsFile, data, 0o644)
}

func (am *AlertManager) loadAlerts() error {
	alertsFile := filepath.Join(am.storageDir, "alerts.json")

	data, err := os.ReadFile(alertsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No alerts yet
		}
		return fmt.Errorf("failed to read alerts: %w", err)
	}

	var alerts []*Alert
	if err := json.Unmarshal(data, &alerts); err != nil {
		return fmt.Errorf("failed to unmarshal alerts: %w", err)
	}

	for _, alert := range alerts {
		am.alerts[alert.ID] = alert
	}

	return nil
}

func (am *AlertManager) saveRules() error {
	rulesFile := filepath.Join(am.storageDir, "rules.json")

	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		rules = append(rules, rule)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	return os.WriteFile(rulesFile, data, 0o644)
}

func (am *AlertManager) loadRules() error {
	rulesFile := filepath.Join(am.storageDir, "rules.json")

	data, err := os.ReadFile(rulesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No rules yet
		}
		return fmt.Errorf("failed to read rules: %w", err)
	}

	var rules []*AlertRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	for _, rule := range rules {
		am.rules[rule.ID] = rule
	}

	return nil
}

// MCP Tool Handlers

// handleGetActiveAlerts handles get_active_alerts tool calls.
func (s *MCPServer) handleGetActiveAlerts(ctx context.Context, req *sdk.CallToolRequest, input GetActiveAlertsInput) (*sdk.CallToolResult, GetActiveAlertsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_active_alerts",
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

	// Initialize alert manager if not already done
	if s.alertManager == nil {
		s.alertManager = NewAlertManager(s.cfg.BaseDir)
	}

	// Evaluate current metrics against rules
	perfStats, err := s.metrics.GetStatistics("last_24h")
	if err == nil {
		tokenStats := s.tokenMetrics.GetDetailedStats()
		s.alertManager.EvaluateMetrics(perfStats, tokenStats)
	}

	// Get active alerts
	alerts := s.alertManager.GetActiveAlerts()

	output := GetActiveAlertsOutput{
		Count:  len(alerts),
		Alerts: alerts,
	}

	s.responseMiddleware.MeasureResponseSize(ctx, "get_active_alerts", output)
	return nil, output, nil
}

// handleGetAlertHistory handles get_alert_history tool calls.
func (s *MCPServer) handleGetAlertHistory(ctx context.Context, req *sdk.CallToolRequest, input GetAlertHistoryInput) (*sdk.CallToolResult, GetAlertHistoryOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_alert_history",
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

	if s.alertManager == nil {
		s.alertManager = NewAlertManager(s.cfg.BaseDir)
	}

	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	alerts := s.alertManager.GetAlertHistory(limit, input.Severity, input.Status)

	output := GetAlertHistoryOutput{
		Count:  len(alerts),
		Alerts: alerts,
	}

	s.responseMiddleware.MeasureResponseSize(ctx, "get_alert_history", output)
	return nil, output, nil
}

// handleGetAlertRules handles get_alert_rules tool calls.
func (s *MCPServer) handleGetAlertRules(ctx context.Context, req *sdk.CallToolRequest, input GetAlertRulesInput) (*sdk.CallToolResult, GetAlertRulesOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_alert_rules",
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

	if s.alertManager == nil {
		s.alertManager = NewAlertManager(s.cfg.BaseDir)
	}

	rules := s.alertManager.GetAlertRules()

	output := GetAlertRulesOutput{
		Count: len(rules),
		Rules: rules,
	}

	s.responseMiddleware.MeasureResponseSize(ctx, "get_alert_rules", output)
	return nil, output, nil
}

// handleUpdateAlertRule handles update_alert_rule tool calls.
func (s *MCPServer) handleUpdateAlertRule(ctx context.Context, req *sdk.CallToolRequest, input UpdateAlertRuleInput) (*sdk.CallToolResult, UpdateAlertRuleOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "update_alert_rule",
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

	if s.alertManager == nil {
		s.alertManager = NewAlertManager(s.cfg.BaseDir)
	}

	rule := &AlertRule{
		ID:          input.RuleID,
		Name:        input.Name,
		Description: input.Description,
		Type:        input.Type,
		Metric:      input.Metric,
		Operator:    input.Operator,
		Threshold:   input.Threshold,
		Severity:    input.Severity,
		Enabled:     input.Enabled,
		Cooldown:    input.Cooldown,
	}

	if err := s.alertManager.UpdateAlertRule(rule); err != nil {
		handlerErr = err
		return nil, UpdateAlertRuleOutput{}, handlerErr
	}

	output := UpdateAlertRuleOutput{
		Success: true,
		Message: fmt.Sprintf("Alert rule '%s' updated successfully", rule.Name),
		Rule:    rule,
	}

	s.responseMiddleware.MeasureResponseSize(ctx, "update_alert_rule", output)
	return nil, output, nil
}

// handleResolveAlert handles resolve_alert tool calls.
func (s *MCPServer) handleResolveAlert(ctx context.Context, req *sdk.CallToolRequest, input ResolveAlertInput) (*sdk.CallToolResult, ResolveAlertOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "resolve_alert",
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

	if s.alertManager == nil {
		s.alertManager = NewAlertManager(s.cfg.BaseDir)
	}

	if err := s.alertManager.ResolveAlert(input.AlertID); err != nil {
		handlerErr = err
		return nil, ResolveAlertOutput{}, handlerErr
	}

	output := ResolveAlertOutput{
		Success: true,
		Message: fmt.Sprintf("Alert %s resolved successfully", input.AlertID),
	}

	s.responseMiddleware.MeasureResponseSize(ctx, "resolve_alert", output)
	return nil, output, nil
}
