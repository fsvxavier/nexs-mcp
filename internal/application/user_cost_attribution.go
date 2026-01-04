package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// UserCostRecord tracks costs for a specific user/session.
type UserCostRecord struct {
	UserID               string             `json:"user_id"`
	SessionID            string             `json:"session_id,omitempty"`
	TotalOperations      int                `json:"total_operations"`
	TotalDuration        float64            `json:"total_duration_ms"`
	TotalTokens          int                `json:"total_tokens"`
	TotalOptimizedTokens int                `json:"total_optimized_tokens"`
	OperationsByTool     map[string]int     `json:"operations_by_tool"`
	DurationByTool       map[string]float64 `json:"duration_by_tool_ms"`
	TokensByTool         map[string]int     `json:"tokens_by_tool"`
	ErrorCount           int                `json:"error_count"`
	FirstSeen            time.Time          `json:"first_seen"`
	LastSeen             time.Time          `json:"last_seen"`
	Metadata             map[string]string  `json:"metadata,omitempty"`
}

// UserCostSummary provides aggregated cost summary for a user.
type UserCostSummary struct {
	UserID                  string         `json:"user_id"`
	TotalSessions           int            `json:"total_sessions"`
	TotalOperations         int            `json:"total_operations"`
	AvgOperationsPerSession float64        `json:"avg_operations_per_session"`
	TotalDuration           float64        `json:"total_duration_ms"`
	AvgDuration             float64        `json:"avg_duration_ms"`
	TotalTokens             int            `json:"total_tokens"`
	TokenSavings            int            `json:"token_savings"`
	TokenSavingsPercent     float64        `json:"token_savings_percent"`
	MostUsedTools           []ToolCostStat `json:"most_used_tools"`
	ErrorRate               float64        `json:"error_rate"`
	CostScore               float64        `json:"cost_score"` // Normalized cost metric
	FirstSeen               time.Time      `json:"first_seen"`
	LastSeen                time.Time      `json:"last_seen"`
	ActiveDays              int            `json:"active_days"`
}

// ToolCostStat represents cost statistics for a specific tool per user.
type ToolCostStat struct {
	ToolName       string  `json:"tool_name"`
	OperationCount int     `json:"operation_count"`
	TotalDuration  float64 `json:"total_duration_ms"`
	TotalTokens    int     `json:"total_tokens"`
	CostScore      float64 `json:"cost_score"`
}

// UserCostAttributionService tracks and reports costs per user/session.
type UserCostAttributionService struct {
	mu         sync.RWMutex
	records    map[string]*UserCostRecord // key: userID or sessionID
	storageDir string
	autosave   bool
	saveTimer  *time.Timer
}

// NewUserCostAttributionService creates a new user cost attribution service.
func NewUserCostAttributionService(storageDir string, autosave bool) *UserCostAttributionService {
	service := &UserCostAttributionService{
		records:    make(map[string]*UserCostRecord),
		storageDir: filepath.Join(storageDir, "user_costs"),
		autosave:   autosave,
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(service.storageDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create user costs directory: %v\n", err)
	}

	// Load existing records
	if err := service.loadRecords(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load user cost records: %v\n", err)
	}

	// Start autosave timer if enabled
	if autosave {
		service.startAutosave()
	}

	return service
}

// RecordUserOperation records a user's operation for cost tracking.
func (s *UserCostAttributionService) RecordUserOperation(ctx context.Context, userID, sessionID, toolName string, duration float64, tokens, optimizedTokens int, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Use userID as primary key, fallback to sessionID
	key := userID
	if key == "" {
		key = sessionID
	}
	if key == "" {
		return // Cannot track without identifier
	}

	record, exists := s.records[key]
	if !exists {
		record = &UserCostRecord{
			UserID:           userID,
			SessionID:        sessionID,
			OperationsByTool: make(map[string]int),
			DurationByTool:   make(map[string]float64),
			TokensByTool:     make(map[string]int),
			FirstSeen:        time.Now(),
			Metadata:         make(map[string]string),
		}
		s.records[key] = record
	}

	// Update metrics
	record.TotalOperations++
	record.TotalDuration += duration
	record.TotalTokens += tokens
	record.TotalOptimizedTokens += optimizedTokens
	record.OperationsByTool[toolName]++
	record.DurationByTool[toolName] += duration
	record.TokensByTool[toolName] += tokens
	record.LastSeen = time.Now()

	if !success {
		record.ErrorCount++
	}
}

// GetUserCostRecord retrieves cost record for a specific user.
func (s *UserCostAttributionService) GetUserCostRecord(userID string) (*UserCostRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, exists := s.records[userID]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	recordCopy := *record
	recordCopy.OperationsByTool = make(map[string]int)
	recordCopy.DurationByTool = make(map[string]float64)
	recordCopy.TokensByTool = make(map[string]int)

	for k, v := range record.OperationsByTool {
		recordCopy.OperationsByTool[k] = v
	}
	for k, v := range record.DurationByTool {
		recordCopy.DurationByTool[k] = v
	}
	for k, v := range record.TokensByTool {
		recordCopy.TokensByTool[k] = v
	}

	return &recordCopy, true
}

// GetUserCostSummary generates a comprehensive cost summary for a user.
func (s *UserCostAttributionService) GetUserCostSummary(userID string) (*UserCostSummary, error) {
	record, exists := s.GetUserCostRecord(userID)
	if !exists {
		return nil, fmt.Errorf("no cost data found for user: %s", userID)
	}

	// Calculate metrics
	avgDuration := 0.0
	if record.TotalOperations > 0 {
		avgDuration = record.TotalDuration / float64(record.TotalOperations)
	}

	tokenSavings := record.TotalTokens - record.TotalOptimizedTokens
	tokenSavingsPercent := 0.0
	if record.TotalTokens > 0 {
		tokenSavingsPercent = float64(tokenSavings) / float64(record.TotalTokens) * 100
	}

	errorRate := 0.0
	if record.TotalOperations > 0 {
		errorRate = float64(record.ErrorCount) / float64(record.TotalOperations) * 100
	}

	// Calculate cost score (normalized 0-100, higher = more expensive)
	// Factors: duration (40%), tokens (40%), error rate (20%)
	durationScore := (record.TotalDuration / float64(record.TotalOperations+1)) / 10.0 // Normalize to ~0-10
	tokenScore := float64(record.TotalTokens) / 1000.0                                 // Normalize to ~0-10
	errorScore := errorRate * 0.1                                                      // 0-10 scale

	costScore := (durationScore*0.4 + tokenScore*0.4 + errorScore*0.2) * 10
	if costScore > 100 {
		costScore = 100
	}

	// Build most used tools list
	mostUsed := s.buildToolCostStats(record)

	// Calculate active days
	activeDays := int(record.LastSeen.Sub(record.FirstSeen).Hours() / 24)
	if activeDays == 0 {
		activeDays = 1
	}

	summary := &UserCostSummary{
		UserID:                  userID,
		TotalSessions:           1, // Per-session tracking would need additional logic
		TotalOperations:         record.TotalOperations,
		AvgOperationsPerSession: float64(record.TotalOperations),
		TotalDuration:           record.TotalDuration,
		AvgDuration:             avgDuration,
		TotalTokens:             record.TotalTokens,
		TokenSavings:            tokenSavings,
		TokenSavingsPercent:     tokenSavingsPercent,
		MostUsedTools:           mostUsed,
		ErrorRate:               errorRate,
		CostScore:               costScore,
		FirstSeen:               record.FirstSeen,
		LastSeen:                record.LastSeen,
		ActiveDays:              activeDays,
	}

	return summary, nil
}

// buildToolCostStats builds tool-level cost statistics.
func (s *UserCostAttributionService) buildToolCostStats(record *UserCostRecord) []ToolCostStat {
	stats := []ToolCostStat{}

	for toolName, opCount := range record.OperationsByTool {
		duration := record.DurationByTool[toolName]
		tokens := record.TokensByTool[toolName]

		// Calculate tool cost score
		costScore := (duration/float64(opCount+1))*0.6 + float64(tokens/opCount+1)*0.4

		stats = append(stats, ToolCostStat{
			ToolName:       toolName,
			OperationCount: opCount,
			TotalDuration:  duration,
			TotalTokens:    tokens,
			CostScore:      costScore,
		})
	}

	// Sort by cost score descending
	sortToolCostStats(stats)

	// Return top 10
	if len(stats) > 10 {
		stats = stats[:10]
	}

	return stats
}

// sortToolCostStats sorts tool cost stats by cost score descending.
func sortToolCostStats(stats []ToolCostStat) {
	for i := range len(stats) - 1 {
		for j := i + 1; j < len(stats); j++ {
			if stats[i].CostScore < stats[j].CostScore {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
}

// GetAllUsers returns a list of all tracked user IDs.
func (s *UserCostAttributionService) GetAllUsers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]string, 0, len(s.records))
	for userID := range s.records {
		users = append(users, userID)
	}

	return users
}

// GetTopUsers returns top N users by cost score.
func (s *UserCostAttributionService) GetTopUsers(n int) ([]*UserCostSummary, error) {
	users := s.GetAllUsers()

	summaries := make([]*UserCostSummary, 0, len(users))
	for _, userID := range users {
		summary, err := s.GetUserCostSummary(userID)
		if err != nil {
			continue
		}
		summaries = append(summaries, summary)
	}

	// Sort by cost score descending
	sortUserSummaries(summaries)

	// Return top N
	if n > 0 && len(summaries) > n {
		summaries = summaries[:n]
	}

	return summaries, nil
}

// sortUserSummaries sorts user summaries by cost score descending.
// In case of equal scores, sorts by total tokens descending for consistency.
func sortUserSummaries(summaries []*UserCostSummary) {
	for i := range len(summaries) - 1 {
		for j := i + 1; j < len(summaries); j++ {
			// Primary sort: by cost score descending
			if summaries[i].CostScore < summaries[j].CostScore {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			} else if summaries[i].CostScore == summaries[j].CostScore {
				// Secondary sort: by total tokens descending (tie-breaker)
				if summaries[i].TotalTokens < summaries[j].TotalTokens {
					summaries[i], summaries[j] = summaries[j], summaries[i]
				}
			}
		}
	}
}

// GetUsersByDateRange returns users active in a date range.
func (s *UserCostAttributionService) GetUsersByDateRange(startTime, endTime time.Time) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := []string{}
	for userID, record := range s.records {
		if record.LastSeen.After(startTime) && record.FirstSeen.Before(endTime) {
			users = append(users, userID)
		}
	}

	return users
}

// ClearUser removes cost data for a specific user.
func (s *UserCostAttributionService) ClearUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.records[userID]; !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	delete(s.records, userID)

	// Save changes
	if s.autosave {
		go func() {
			if err := s.saveRecords(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save after clearing user: %v\n", err)
			}
		}()
	}

	return nil
}

// UpdateUserMetadata updates metadata for a user.
func (s *UserCostAttributionService) UpdateUserMetadata(userID string, metadata map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, exists := s.records[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	if record.Metadata == nil {
		record.Metadata = make(map[string]string)
	}

	for k, v := range metadata {
		record.Metadata[k] = v
	}

	return nil
}

// Storage methods

func (s *UserCostAttributionService) saveRecords() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recordsFile := filepath.Join(s.storageDir, "user_costs.json")

	records := make([]*UserCostRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user costs: %w", err)
	}

	return os.WriteFile(recordsFile, data, 0o644)
}

func (s *UserCostAttributionService) loadRecords() error {
	recordsFile := filepath.Join(s.storageDir, "user_costs.json")

	data, err := os.ReadFile(recordsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No records yet
		}
		return fmt.Errorf("failed to read user costs: %w", err)
	}

	var records []*UserCostRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("failed to unmarshal user costs: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, record := range records {
		key := record.UserID
		if key == "" {
			key = record.SessionID
		}
		if key != "" {
			s.records[key] = record
		}
	}

	return nil
}

func (s *UserCostAttributionService) startAutosave() {
	s.saveTimer = time.AfterFunc(5*time.Minute, func() {
		if err := s.saveRecords(); err != nil {
			fmt.Fprintf(os.Stderr, "Autosave error: %v\n", err)
		}
		s.startAutosave() // Reschedule
	})
}

// Stop gracefully shuts down the service.
func (s *UserCostAttributionService) Stop() error {
	if s.saveTimer != nil {
		s.saveTimer.Stop()
	}

	// Final save
	return s.saveRecords()
}
