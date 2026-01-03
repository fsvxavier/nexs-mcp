package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

// WorkingMemoryService manages session-scoped working memory with auto-promotion.
// Implements two-tier memory architecture:
// - Working Memory: Session-scoped, TTL-based, high-churn, auto-promoted
// - Long-term Memory: Persistent, curated, manually managed (via existing Memory domain).
type WorkingMemoryService struct {
	store              domain.ElementRepository // For promoting to long-term memory
	sessions           map[string]*SessionMemoryCache
	mu                 sync.RWMutex
	cleanupTick        *time.Ticker
	stopCleanup        chan struct{}
	persistenceDir     string // Directory for persisting working memories to disk
	persistenceEnabled bool
}

// SessionMemoryCache holds working memory for a single session.
type SessionMemoryCache struct {
	SessionID    string
	Memories     map[string]*domain.WorkingMemory
	LastActivity time.Time
	mu           sync.RWMutex
}

// NewWorkingMemoryService creates a new working memory service.
func NewWorkingMemoryService(store domain.ElementRepository) *WorkingMemoryService {
	return NewWorkingMemoryServiceWithPersistence(store, "", false)
}

// NewWorkingMemoryServiceWithPersistence creates a new working memory service with file persistence.
func NewWorkingMemoryServiceWithPersistence(store domain.ElementRepository, persistenceDir string, enablePersistence bool) *WorkingMemoryService {
	svc := &WorkingMemoryService{
		store:              store,
		sessions:           make(map[string]*SessionMemoryCache),
		cleanupTick:        time.NewTicker(5 * time.Minute), // Cleanup every 5 minutes
		stopCleanup:        make(chan struct{}),
		persistenceDir:     persistenceDir,
		persistenceEnabled: enablePersistence,
	}

	// Create persistence directory if enabled
	if enablePersistence && persistenceDir != "" {
		if err := os.MkdirAll(persistenceDir, 0o755); err != nil {
			logger.Error("Failed to create working memory persistence directory", map[string]interface{}{
				"dir":   persistenceDir,
				"error": err.Error(),
			})
			svc.persistenceEnabled = false
		} else {
			logger.Info("Working memory file persistence enabled", map[string]interface{}{
				"dir": persistenceDir,
			})
			// Load existing working memories from disk on startup
			svc.loadFromDisk()
		}
	}

	// Start background cleanup goroutine
	go svc.backgroundCleanup()

	return svc
}

// Add creates a new working memory in the session.
func (s *WorkingMemoryService) Add(ctx context.Context, sessionID, content string, priority domain.MemoryPriority, tags []string, metadata map[string]string) (*domain.WorkingMemory, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	if content == "" {
		return nil, errors.New("content cannot be empty")
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
		"expires_at": wm.GetExpiresAt().Format(time.RFC3339),
	})

	// Persist to disk if enabled
	if s.persistenceEnabled {
		if err := s.persistToDisk(wm); err != nil {
			logger.Warn("Failed to persist working memory to disk", map[string]interface{}{
				"id":    wm.ID,
				"error": err.Error(),
			})
		}
	}

	// Check for auto-promotion
	if wm.ShouldPromote() {
		go s.autoPromote(ctx, wm)
	}

	return wm, nil
}

// Get retrieves a working memory by ID and records access.
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

// List retrieves all working memories for a session (optionally filtered).
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

// Promote manually promotes a working memory to long-term memory.
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

	// Attempt to become the promoting goroutine. If another goroutine is already
	// promoting, wait briefly for it to finish and then return the created long-term
	// memory if available.
	if !wm.TryStartPromotion() {
		// Wait for the ongoing promotion to complete (bounded wait)
		for range 200 {
			if wm.IsPromoted() {
				promotedID := wm.GetPromotedToID()
				if promotedID != "" {
					existingMem, err := s.store.GetByID(promotedID)
					if err == nil {
						if mem, ok := existingMem.(*domain.Memory); ok {
							return mem, nil
						}
					}
				}
				return nil, errors.New("memory already promoted but long-term memory not found")
			}
			// Sleep a bit and retry
			time.Sleep(5 * time.Millisecond)
		}

		// If still not promoted, give up trying to wait and return an error to avoid
		// blocking indefinitely. The caller may retry.
		return nil, errors.New("concurrent promotion in progress")
	}

	// Get all needed data from working memory in a thread-safe manner
	wmContent := wm.GetContent()
	wmTags := wm.GetTagsCopy()
	sessionID = wm.GetSessionID()
	memoryID = wm.GetID()

	// Create long-term memory using NewMemory constructor
	longTermMem := domain.NewMemory(
		"promoted_"+memoryID,
		"Promoted from session "+sessionID,
		"1.0",
		"working_memory_service",
	)
	longTermMem.Content = wmContent

	// Copy tags
	metadata := longTermMem.GetMetadata()
	metadata.Tags = append(metadata.Tags, wmTags...)
	metadata.Tags = append(metadata.Tags, "promoted", "session:"+sessionID)
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
	longTermMem.Metadata["access_count"] = strconv.Itoa(accessCount)
	longTermMem.Metadata["importance_score"] = fmt.Sprintf("%.2f", importance)
	longTermMem.Metadata["priority"] = string(priority)

	// Compute content hash
	longTermMem.ComputeHash()

	// Save to repository using Element interface
	if err := s.store.Create(longTermMem); err != nil {
		// Clear the promoting flag so another goroutine can retry
		wm.CancelPromotion()
		return nil, fmt.Errorf("failed to save long-term memory: %w", err)
	}

	// Mark working memory as promoted (FinishPromotion is thread-safe)
	cache.mu.Lock()
	wm.FinishPromotion(longTermMem.GetID())
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

// autoPromote is called asynchronously to promote memories.
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

// ClearSession removes all working memories for a session.
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

// ExpireMemory manually expires a working memory.
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

	// Expire using thread-safe method on WorkingMemory
	wm.Expire()

	logger.Info("Working memory expired", map[string]interface{}{
		"memory_id":  memoryID,
		"session_id": sessionID,
	})

	return nil
}

// ExtendTTL extends the TTL of a working memory.
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
		"expires_at": wm.GetExpiresAt().Format(time.RFC3339),
	})

	return nil
}

// GetStats returns statistics for a session's working memory.
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

		// Aggregate metrics (use thread-safe getters)
		totalAccess += wm.GetAccessCount()
		totalImportance += wm.GetImportanceScore()
		p := wm.GetPriority()
		stats.ByPriority[p]++
	}

	// Calculate averages
	if stats.TotalCount > 0 {
		stats.AvgAccessCount = float64(totalAccess) / float64(stats.TotalCount)
		stats.AvgImportance = totalImportance / float64(stats.TotalCount)
	}

	return stats
}

// Export exports all working memories for a session (for backup/migration).
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

// getSession retrieves a session cache (read-only).
func (s *WorkingMemoryService) getSession(sessionID string) *SessionMemoryCache {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[sessionID]
}

// getOrCreateSession retrieves or creates a session cache.
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

// backgroundCleanup periodically removes expired memories and inactive sessions.
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

// cleanup removes expired memories and inactive sessions.
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

		// Capture stats under lock to avoid races
		lastActivity := cache.LastActivity
		remaining := len(cache.Memories)

		cache.mu.Unlock()

		// Remove inactive sessions (no activity for 2 hours and no active memories)
		if now.Sub(lastActivity) > inactiveSessionThreshold && remaining == 0 {
			delete(s.sessions, sessionID)
			logger.Info("Inactive session removed", map[string]interface{}{
				"session_id": sessionID,
			})
		} else if expiredCount > 0 {
			logger.Info("Expired memories cleaned", map[string]interface{}{
				"session_id":      sessionID,
				"expired_count":   expiredCount,
				"remaining_count": remaining,
			})
		}
	}
}

// Shutdown stops the background cleanup goroutine.
func (s *WorkingMemoryService) Shutdown() {
	close(s.stopCleanup)
	s.cleanupTick.Stop()
}

// persistToDisk saves a working memory to disk as JSON.
func (s *WorkingMemoryService) persistToDisk(wm *domain.WorkingMemory) error {
	if !s.persistenceEnabled || s.persistenceDir == "" {
		return nil
	}

	// Create filename based on session_id and timestamp
	filename := fmt.Sprintf("%s_%d.json", wm.SessionID, time.Now().Unix())
	filepath := filepath.Join(s.persistenceDir, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(wm, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal working memory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write working memory file: %w", err)
	}

	logger.Debug("Working memory persisted to disk", map[string]interface{}{
		"id":   wm.ID,
		"file": filepath,
	})

	return nil
}

// loadFromDisk loads all working memories from disk on service startup.
func (s *WorkingMemoryService) loadFromDisk() {
	if !s.persistenceEnabled || s.persistenceDir == "" {
		return
	}

	files, err := os.ReadDir(s.persistenceDir)
	if err != nil {
		logger.Warn("Failed to read working memory persistence directory", map[string]interface{}{
			"dir":   s.persistenceDir,
			"error": err.Error(),
		})
		return
	}

	loadedCount := 0
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(s.persistenceDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("Failed to read working memory file", map[string]interface{}{
				"file":  filePath,
				"error": err.Error(),
			})
			continue
		}

		var wm domain.WorkingMemory
		if err := json.Unmarshal(data, &wm); err != nil {
			logger.Warn("Failed to unmarshal working memory file", map[string]interface{}{
				"file":  filePath,
				"error": err.Error(),
			})
			continue
		}

		// Skip expired memories
		if wm.IsExpired() {
			if err := os.Remove(filePath); err != nil {
				logger.Warn("Failed to remove expired working memory file", map[string]interface{}{
					"file":  filePath,
					"error": err.Error(),
				})
			}
			continue
		}

		// Add to session cache
		cache := s.getOrCreateSession(wm.SessionID)
		cache.mu.Lock()
		cache.Memories[wm.ID] = &wm
		cache.mu.Unlock()

		loadedCount++
	}

	if loadedCount > 0 {
		logger.Info("Working memories loaded from disk", map[string]interface{}{
			"count": loadedCount,
			"dir":   s.persistenceDir,
		})
	}
}
