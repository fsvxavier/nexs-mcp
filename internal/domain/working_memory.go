package domain

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// WorkingMemory represents session-scoped, temporary memory
// with automatic promotion to long-term memory based on importance/access patterns.
//
// Working Memory vs Long-term Memory:
// - Working: Session-scoped, TTL-based, high-churn, auto-promoted
// - Long-term: Persistent, curated, manually managed, permanent.
type WorkingMemory struct {
	mu              sync.RWMutex      // Protects concurrent access to mutable fields
	ID              string            `json:"id"                       yaml:"id"`                       // Unique identifier
	SessionID       string            `json:"session_id"               yaml:"session_id"`               // Session scope identifier
	Content         string            `json:"content"                  yaml:"content"`                  // Memory content
	Context         string            `json:"context,omitempty"        yaml:"context,omitempty"`        // Additional context
	Tags            []string          `json:"tags,omitempty"           yaml:"tags,omitempty"`           // Searchable tags
	Metadata        map[string]string `json:"metadata,omitempty"       yaml:"metadata,omitempty"`       // Custom metadata
	Priority        MemoryPriority    `json:"priority"                 yaml:"priority"`                 // Priority level (affects TTL)
	AccessCount     int               `json:"access_count"             yaml:"access_count"`             // Number of accesses
	ImportanceScore float64           `json:"importance_score"         yaml:"importance_score"`         // 0.0-1.0 calculated importance
	CreatedAt       time.Time         `json:"created_at"               yaml:"created_at"`               // Creation timestamp
	LastAccessedAt  time.Time         `json:"last_accessed_at"         yaml:"last_accessed_at"`         // Last access timestamp
	ExpiresAt       time.Time         `json:"expires_at"               yaml:"expires_at"`               // Expiration timestamp (TTL)
	PromotedAt      *time.Time        `json:"promoted_at,omitempty"    yaml:"promoted_at,omitempty"`    // Promotion timestamp (nil if not promoted)
	PromotedToID    string            `json:"promoted_to_id,omitempty" yaml:"promoted_to_id,omitempty"` // ID of long-term memory after promotion
	RelatedIDs      []string          `json:"related_ids,omitempty"    yaml:"related_ids,omitempty"`    // Related working memory IDs
	Source          string            `json:"source,omitempty"         yaml:"source,omitempty"`         // Source (user, agent, system)
}

// MemoryPriority defines priority levels that affect TTL and promotion likelihood.
type MemoryPriority string

const (
	PriorityLow      MemoryPriority = "low"      // TTL: 1 hour, auto-expire quickly
	PriorityMedium   MemoryPriority = "medium"   // TTL: 4 hours, moderate retention
	PriorityHigh     MemoryPriority = "high"     // TTL: 12 hours, important info
	PriorityCritical MemoryPriority = "critical" // TTL: 24 hours, almost permanent
)

// TTL returns the default TTL duration for a priority level.
func (p MemoryPriority) TTL() time.Duration {
	switch p {
	case PriorityLow:
		return 1 * time.Hour
	case PriorityMedium:
		return 4 * time.Hour
	case PriorityHigh:
		return 12 * time.Hour
	case PriorityCritical:
		return 24 * time.Hour
	default:
		return 4 * time.Hour // Default: medium
	}
}

// PromotionThreshold returns the access count threshold for auto-promotion.
func (p MemoryPriority) PromotionThreshold() int {
	switch p {
	case PriorityLow:
		return 10 // Rarely accessed, high bar
	case PriorityMedium:
		return 7 // Moderate access needed
	case PriorityHigh:
		return 5 // Easily promoted
	case PriorityCritical:
		return 3 // Almost auto-promote
	default:
		return 7
	}
}

// NewWorkingMemory creates a new working memory with default values.
func NewWorkingMemory(sessionID, content string, priority MemoryPriority) *WorkingMemory {
	now := time.Now()
	ttl := priority.TTL()

	return &WorkingMemory{
		ID:              GenerateElementID(WorkingMemoryElement, sessionID+"-"+content[:min(20, len(content))]),
		SessionID:       sessionID,
		Content:         content,
		Tags:            []string{},
		Metadata:        make(map[string]string),
		Priority:        priority,
		AccessCount:     0,
		ImportanceScore: 0.0,
		CreatedAt:       now,
		LastAccessedAt:  now,
		ExpiresAt:       now.Add(ttl),
		RelatedIDs:      []string{},
		Source:          "user",
	}
}

// RecordAccess updates access tracking and recalculates importance.
func (wm *WorkingMemory) RecordAccess() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.AccessCount++
	wm.LastAccessedAt = time.Now()
	wm.calculateImportance()
}

// IsExpired checks if the memory has expired.
func (wm *WorkingMemory) IsExpired() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return time.Now().After(wm.ExpiresAt)
}

// IsPromoted checks if the memory has been promoted to long-term.
func (wm *WorkingMemory) IsPromoted() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.PromotedAt != nil
}

// ShouldPromote determines if memory should be auto-promoted based on access patterns.
func (wm *WorkingMemory) ShouldPromote() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if wm.PromotedAt != nil {
		return false // Already promoted
	}

	threshold := wm.Priority.PromotionThreshold()

	// Rule 1: Access count exceeds threshold
	if wm.AccessCount >= threshold {
		return true
	}

	// Rule 2: High importance score (>0.8) + moderate access (>3)
	if wm.ImportanceScore >= 0.8 && wm.AccessCount >= 3 {
		return true
	}

	// Rule 3: Critical priority + accessed at least once
	if wm.Priority == PriorityCritical && wm.AccessCount >= 1 {
		return true
	}

	// Rule 4: Long-lived + consistently accessed
	age := time.Since(wm.CreatedAt)
	if age > 6*time.Hour && wm.AccessCount >= 5 {
		return true
	}

	return false
}

// calculateImportance computes importance score based on:
// - Access frequency (weight: 0.4)
// - Age decay (weight: 0.3)
// - Priority level (weight: 0.2)
// - Content length (weight: 0.1).
func (wm *WorkingMemory) calculateImportance() {
	// Access frequency component (0-1 scale, capped at 20 accesses)
	accessScore := float64(wm.AccessCount) / 20.0
	if accessScore > 1.0 {
		accessScore = 1.0
	}

	// Age decay component (exponential decay, half-life = 2 hours)
	age := time.Since(wm.CreatedAt).Hours()
	decayScore := 1.0 / (1.0 + age/2.0)

	// Priority component
	var priorityScore float64
	switch wm.Priority {
	case PriorityLow:
		priorityScore = 0.2
	case PriorityMedium:
		priorityScore = 0.5
	case PriorityHigh:
		priorityScore = 0.8
	case PriorityCritical:
		priorityScore = 1.0
	default:
		priorityScore = 0.5
	}

	// Content length component (longer = potentially more important)
	contentScore := float64(len(wm.Content)) / 1000.0
	if contentScore > 1.0 {
		contentScore = 1.0
	}

	// Weighted combination
	wm.ImportanceScore = (accessScore * 0.4) +
		(decayScore * 0.3) +
		(priorityScore * 0.2) +
		(contentScore * 0.1)

	// Clamp to [0, 1]
	if wm.ImportanceScore > 1.0 {
		wm.ImportanceScore = 1.0
	}
}

// GetAccessCount returns the access count in a thread-safe manner.
func (wm *WorkingMemory) GetAccessCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.AccessCount
}

// GetImportanceScore returns the importance score in a thread-safe manner.
func (wm *WorkingMemory) GetImportanceScore() float64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.ImportanceScore
}

// GetPriority returns the priority in a thread-safe manner.
func (wm *WorkingMemory) GetPriority() MemoryPriority {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.Priority
}

// GetMetadataCopy returns a copy of the metadata in a thread-safe manner.
func (wm *WorkingMemory) GetMetadataCopy() map[string]string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	copy := make(map[string]string, len(wm.Metadata))
	for k, v := range wm.Metadata {
		copy[k] = v
	}
	return copy
}

// GetID returns the ID (immutable, no lock needed but consistent with other getters).
func (wm *WorkingMemory) GetID() string {
	return wm.ID
}

// GetContent returns the content in a thread-safe manner.
func (wm *WorkingMemory) GetContent() string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.Content
}

// GetTagsCopy returns a copy of the tags in a thread-safe manner.
func (wm *WorkingMemory) GetTagsCopy() []string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	copy := make([]string, len(wm.Tags))
	for i, tag := range wm.Tags {
		copy[i] = tag
	}
	return copy
}

// GetSessionID returns the session ID (immutable after creation).
func (wm *WorkingMemory) GetSessionID() string {
	return wm.SessionID
}

// MarkPromoted marks the memory as promoted to long-term.
func (wm *WorkingMemory) MarkPromoted(longTermMemoryID string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	now := time.Now()
	wm.PromotedAt = &now
	wm.PromotedToID = longTermMemoryID
}

// ExtendTTL extends the expiration time by the default TTL for this priority.
func (wm *WorkingMemory) ExtendTTL() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	ttl := wm.Priority.TTL()
	wm.ExpiresAt = time.Now().Add(ttl)
}

// AddRelation adds a related working memory ID.
func (wm *WorkingMemory) AddRelation(relatedID string) {
	// Avoid duplicates
	for _, id := range wm.RelatedIDs {
		if id == relatedID {
			return
		}
	}
	wm.RelatedIDs = append(wm.RelatedIDs, relatedID)
}

// Validate validates working memory fields.
func (wm *WorkingMemory) Validate() error {
	if wm.ID == "" {
		return errors.New("working memory ID cannot be empty")
	}

	if wm.SessionID == "" {
		return errors.New("session ID cannot be empty")
	}

	if wm.Content == "" {
		return errors.New("content cannot be empty")
	}

	if wm.Priority != PriorityLow &&
		wm.Priority != PriorityMedium &&
		wm.Priority != PriorityHigh &&
		wm.Priority != PriorityCritical {
		return fmt.Errorf("invalid priority: %s", wm.Priority)
	}

	if wm.ImportanceScore < 0.0 || wm.ImportanceScore > 1.0 {
		return fmt.Errorf("importance score must be between 0.0 and 1.0, got %f", wm.ImportanceScore)
	}

	return nil
}

// WorkingMemoryStats represents statistics for a session's working memory.
type WorkingMemoryStats struct {
	SessionID        string                 `json:"session_id"`
	TotalCount       int                    `json:"total_count"`
	ActiveCount      int                    `json:"active_count"`
	ExpiredCount     int                    `json:"expired_count"`
	PromotedCount    int                    `json:"promoted_count"`
	AvgAccessCount   float64                `json:"avg_access_count"`
	AvgImportance    float64                `json:"avg_importance"`
	PendingPromotion int                    `json:"pending_promotion"`
	ByPriority       map[MemoryPriority]int `json:"by_priority"`
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
