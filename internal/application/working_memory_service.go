package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

// WorkingMemoryService manages session-scoped working memory with auto-promotion.
// Implements two-tier memory architecture:
// - Working Memory: Session-scoped, TTL-based, high-churn, auto-promoted
// - Long-term Memory: Persistent, curated, manually managed (via existing Memory domain)
type WorkingMemoryService struct {
	store       domain.ElementRepository // For promoting to long-term memory
	sessions    map[string]*SessionMemoryCache
	mu          sync.RWMutex
	cleanupTick *time.Ticker
	stopCleanup chan struct{}
}

// SessionMemoryCache holds working memory for a single session
type SessionMemoryCache struct {
	SessionID    string
	Memories     map[string]*domain.WorkingMemory
	LastActivity time.Time
	mu           sync.RWMutex
}

// NewWorkingMemoryService creates a new working memory service
func NewWorkingMemoryService(store domain.ElementRepository) *WorkingMemoryService {
	svc := &WorkingMemoryService{
		store:       store,
		sessions:    make(map[string]*SessionMemoryCache),
		cleanupTick: time.NewTicker(5 * time.Minute), // Cleanup every 5 minutes
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go svc.backgroundCleanup()

	return svc
}

// Add creates a new working memory in the session
func (s *WorkingMemoryService) Add(ctx context.Context, sessionID, content string, priority domain.MemoryPriority, tags []string, metadata map[string]string) (*domain.WorkingMemory, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	// Create working memory
	wm := domain.NewWorkingMemory(sessionID, content, priority)
	wm.Tags = tags
	if metadata != nil {
		wm.Metadata = metadata
	}

	// Calculate initial importance
	wm.RecordAccess()

	// Validate
	if err := wm.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get or create session cache
	cache := s.getOrCreateSession(sessionID)

	// Store in session
	cache.mu.Lock()
	cache.Memories[wm.ID] = wm
	cache.LastActivity = time.Now()
	cache.mu.Unlock()

	logger.Info("Working memory created", map[string]interface{}{
		"id":         wm.ID,
		"session_id": sessionID,
		"priority":   priority,
		"expires_at": wm.ExpiresAt.Format(time.RFC3339),
	})

	// Check for auto-promotion
	if wm.ShouldPromote() {
		go s.autoPromote(ctx, wm)
	}

	return wm, nil
}

// Get retrieves a working memory by ID and records access
func (s *WorkingMemoryService) Get(ctx context.Context, sessionID, memoryID string) (*domain.WorkingMemory, error) {
	cache := s.getSession(sessionID)
	if cache == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	cache.mu.RLock()
	wm, exists := cache.Memories[memoryID]
	cache.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("working memory not found: %s", memoryID)
	}

	// Check expiration
	if wm.IsExpired() {
		return nil, fmt.Errorf("working memory expired: %s", memoryID)
	}

	// Record access
	cache.mu.Lock()
	wm.RecordAccess()
	cache.LastActivity = time.Now()
	cache.mu.Unlock()

	// Check for auto-promotion after access
	if wm.ShouldPromote() && !wm.IsPromoted() {
		go s.autoPromote(ctx, wm)
	}

	return wm, nil
}

// List retrieves all working memories for a session (optionally filtered)
func (s *WorkingMemoryService) List(ctx context.Context, sessionID string, includeExpired, includePromoted bool) ([]*domain.WorkingMemory, error) {
	cache := s.getSession(sessionID)
	if cache == nil {
		return []*domain.WorkingMemory{}, nil // Empty session
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	result := make([]*domain.WorkingMemory, 0, len(cache.Memories))
	for _, wm := range cache.Memories {
		// Filter expired
		if !includeExpired && wm.IsExpired() {
			continue
		}

		// Filter promoted
		if !includePromoted && wm.IsPromoted() {
			continue
		}

		result = append(result, wm)
	}

	return result, nil
}

// Promote manually promotes a working memory to long-term memory
func (s *WorkingMemoryService) Promote(ctx context.Context, sessionID, memoryID string) (*domain.Memory, error) {
	cache := s.getSession(sessionID)
	if cache == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	cache.mu.RLock()
	wm, exists := cache.Memories[memoryID]
	cache.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("working memory not found: %s", memoryID)
	}

	if wm.IsPromoted() {
		// Already promoted, try to retrieve existing long-term memory
		if wm.PromotedToID != "" {
			existingMem, err := s.store.GetByID(wm.PromotedToID)
			if err == nil {
				if mem, ok := existingMem.(*domain.Memory); ok {
					return mem, nil
				}
			}
		}
		return nil, fmt.Errorf("memory already promoted but long-term memory not found")
	}

	// Get all needed data from working memory in a thread-safe manner
	wmContent := wm.GetContent()
	wmTags := wm.GetTagsCopy()
	sessionID = wm.GetSessionID()
	memoryID = wm.GetID()

	// Create long-term memory using NewMemory constructor
	longTermMem := domain.NewMemory(
		fmt.Sprintf("promoted_%s", memoryID),
		fmt.Sprintf("Promoted from session %s", sessionID),
		"1.0",
		"working_memory_service",
	)
	longTermMem.Content = wmContent

	// Copy tags
	metadata := longTermMem.GetMetadata()
	metadata.Tags = append(metadata.Tags, wmTags...)
	metadata.Tags = append(metadata.Tags, "promoted", fmt.Sprintf("session:%s", sessionID))
	longTermMem.SetMetadata(metadata)

	// Copy and extend metadata (using thread-safe getters)
	if longTermMem.Metadata == nil {
		longTermMem.Metadata = make(map[string]string)
	}

	// Get a thread-safe copy of the metadata
	wmMetadata := wm.GetMetadataCopy()
	for k, v := range wmMetadata {
		longTermMem.Metadata[k] = v
	}

	// Get other fields in a thread-safe manner
	wmID := wm.GetID()
	accessCount := wm.GetAccessCount()
	importance := wm.GetImportanceScore()
	priority := wm.GetPriority()

	longTermMem.Metadata["promoted_from"] = wmID
	longTermMem.Metadata["promoted_at"] = time.Now().Format(time.RFC3339)
	longTermMem.Metadata["access_count"] = fmt.Sprintf("%d", accessCount)
	longTermMem.Metadata["importance_score"] = fmt.Sprintf("%.2f", importance)
	longTermMem.Metadata["priority"] = string(priority)

	// Compute content hash
	longTermMem.ComputeHash()

	// Save to repository using Element interface
	if err := s.store.Create(longTermMem); err != nil {
		return nil, fmt.Errorf("failed to save long-term memory: %w", err)
	}

	// Mark working memory as promoted
	cache.mu.Lock()
	wm.MarkPromoted(longTermMem.GetID())
	cache.mu.Unlock()

	// Log promotion (using thread-safe getters)
	logger.Info("Working memory promoted", map[string]interface{}{
		"working_id":   wmID,
		"long_term_id": longTermMem.GetID(),
		"session_id":   sessionID,
		"access_count": accessCount,
		"importance":   importance,
	})

	return longTermMem, nil
}

// autoPromote is called asynchronously to promote memories
func (s *WorkingMemoryService) autoPromote(ctx context.Context, wm *domain.WorkingMemory) {
	_, err := s.Promote(ctx, wm.SessionID, wm.ID)
	if err != nil {
		logger.Error("Auto-promotion failed", map[string]interface{}{
			"error":      err.Error(),
			"memory_id":  wm.ID,
			"session_id": wm.SessionID,
		})
	}
}

// ClearSession removes all working memories for a session
func (s *WorkingMemoryService) ClearSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(s.sessions, sessionID)

	logger.Info("Session cleared", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}

// ExpireMemory manually expires a working memory
func (s *WorkingMemoryService) ExpireMemory(sessionID, memoryID string) error {
	cache := s.getSession(sessionID)
	if cache == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	wm, exists := cache.Memories[memoryID]
	if !exists {
		return fmt.Errorf("working memory not found: %s", memoryID)
	}

	// Set expiration to past
	wm.ExpiresAt = time.Now().Add(-1 * time.Second)

	logger.Info("Working memory expired", map[string]interface{}{
		"memory_id":  memoryID,
		"session_id": sessionID,
	})

	return nil
}

// ExtendTTL extends the TTL of a working memory
func (s *WorkingMemoryService) ExtendTTL(sessionID, memoryID string) error {
	cache := s.getSession(sessionID)
	if cache == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	wm, exists := cache.Memories[memoryID]
	if !exists {
		return fmt.Errorf("working memory not found: %s", memoryID)
	}

	wm.ExtendTTL()

	logger.Info("Working memory TTL extended", map[string]interface{}{
		"memory_id":  memoryID,
		"session_id": sessionID,
		"expires_at": wm.ExpiresAt.Format(time.RFC3339),
	})

	return nil
}

// GetStats returns statistics for a session's working memory
func (s *WorkingMemoryService) GetStats(sessionID string) *domain.WorkingMemoryStats {
	cache := s.getSession(sessionID)
	if cache == nil {
		return &domain.WorkingMemoryStats{
			SessionID:  sessionID,
			ByPriority: make(map[domain.MemoryPriority]int),
		}
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	stats := &domain.WorkingMemoryStats{
		SessionID:  sessionID,
		TotalCount: len(cache.Memories),
		ByPriority: make(map[domain.MemoryPriority]int),
	}

	var totalAccess int
	var totalImportance float64

	for _, wm := range cache.Memories {
		// Count by state
		if wm.IsExpired() {
			stats.ExpiredCount++
		} else {
			stats.ActiveCount++
		}

		if wm.IsPromoted() {
			stats.PromotedCount++
		}

		if wm.ShouldPromote() && !wm.IsPromoted() {
			stats.PendingPromotion++
		}

		// Aggregate metrics
		totalAccess += wm.AccessCount
		totalImportance += wm.ImportanceScore
		stats.ByPriority[wm.Priority]++
	}

	// Calculate averages
	if stats.TotalCount > 0 {
		stats.AvgAccessCount = float64(totalAccess) / float64(stats.TotalCount)
		stats.AvgImportance = totalImportance / float64(stats.TotalCount)
	}

	return stats
}

// Export exports all working memories for a session (for backup/migration)
func (s *WorkingMemoryService) Export(sessionID string) ([]*domain.WorkingMemory, error) {
	cache := s.getSession(sessionID)
	if cache == nil {
		return []*domain.WorkingMemory{}, nil
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	result := make([]*domain.WorkingMemory, 0, len(cache.Memories))
	for _, wm := range cache.Memories {
		result = append(result, wm)
	}

	return result, nil
}

// getSession retrieves a session cache (read-only)
func (s *WorkingMemoryService) getSession(sessionID string) *SessionMemoryCache {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[sessionID]
}

// getOrCreateSession retrieves or creates a session cache
func (s *WorkingMemoryService) getOrCreateSession(sessionID string) *SessionMemoryCache {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cache, exists := s.sessions[sessionID]; exists {
		return cache
	}

	// Create new session
	cache := &SessionMemoryCache{
		SessionID:    sessionID,
		Memories:     make(map[string]*domain.WorkingMemory),
		LastActivity: time.Now(),
	}

	s.sessions[sessionID] = cache
	return cache
}

// backgroundCleanup periodically removes expired memories and inactive sessions
func (s *WorkingMemoryService) backgroundCleanup() {
	for {
		select {
		case <-s.cleanupTick.C:
			s.cleanup()
		case <-s.stopCleanup:
			return
		}
	}
}

// cleanup removes expired memories and inactive sessions
func (s *WorkingMemoryService) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	inactiveSessionThreshold := 2 * time.Hour

	for sessionID, cache := range s.sessions {
		cache.mu.Lock()

		// Remove expired memories
		expiredCount := 0
		for id, wm := range cache.Memories {
			if wm.IsExpired() {
				delete(cache.Memories, id)
				expiredCount++
			}
		}

		cache.mu.Unlock()

		// Remove inactive sessions (no activity for 2 hours and no active memories)
		if now.Sub(cache.LastActivity) > inactiveSessionThreshold && len(cache.Memories) == 0 {
			delete(s.sessions, sessionID)
			logger.Info("Inactive session removed", map[string]interface{}{
				"session_id": sessionID,
			})
		} else if expiredCount > 0 {
			logger.Info("Expired memories cleaned", map[string]interface{}{
				"session_id":      sessionID,
				"expired_count":   expiredCount,
				"remaining_count": len(cache.Memories),
			})
		}
	}
}

// Shutdown stops the background cleanup goroutine
func (s *WorkingMemoryService) Shutdown() {
	close(s.stopCleanup)
	s.cleanupTick.Stop()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
