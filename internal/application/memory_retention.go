package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/quality"
)

// MemoryRetentionService manages memory lifecycle based on quality scores.
type MemoryRetentionService struct {
	config            *quality.Config
	scorer            quality.Scorer
	memoryRepo        domain.ElementRepository
	workingMemService *WorkingMemoryService
	mu                sync.RWMutex
	stopChan          chan struct{}
	running           bool
	stats             *RetentionStats
}

// RetentionStats tracks retention operations.
type RetentionStats struct {
	mu              sync.RWMutex
	totalScored     int
	totalArchived   int
	totalDeleted    int
	lastCleanup     time.Time
	avgQualityScore float64
	policyBreakdown map[string]int // high/medium/low counts
}

// NewMemoryRetentionService creates a new retention service.
func NewMemoryRetentionService(
	config *quality.Config,
	scorer quality.Scorer,
	memoryRepo domain.ElementRepository,
	workingMemService *WorkingMemoryService,
) *MemoryRetentionService {
	if config == nil {
		config = quality.DefaultConfig()
	}

	return &MemoryRetentionService{
		config:            config,
		scorer:            scorer,
		memoryRepo:        memoryRepo,
		workingMemService: workingMemService,
		stopChan:          make(chan struct{}),
		stats: &RetentionStats{
			policyBreakdown: make(map[string]int),
		},
	}
}

// Start begins the background retention cleanup process.
func (rs *MemoryRetentionService) Start(ctx context.Context) error {
	rs.mu.Lock()
	if rs.running {
		rs.mu.Unlock()
		return errors.New("retention service already running")
	}
	rs.running = true
	rs.mu.Unlock()

	if !rs.config.EnableAutoArchival {
		return nil
	}

	// Start background cleanup goroutine
	go rs.cleanupLoop(ctx)

	return nil
}

// Stop halts the background retention process.
func (rs *MemoryRetentionService) Stop() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if !rs.running {
		return errors.New("retention service not running")
	}

	close(rs.stopChan)
	rs.running = false

	return nil
}

// cleanupLoop runs periodic cleanup operations.
func (rs *MemoryRetentionService) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(rs.config.CleanupIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rs.stopChan:
			return
		case <-ticker.C:
			if err := rs.RunCleanup(ctx); err != nil {
				// Log error but continue running
				fmt.Printf("Retention cleanup error: %v\n", err)
			}
		}
	}
}

// RunCleanup executes a full retention cleanup cycle.
func (rs *MemoryRetentionService) RunCleanup(ctx context.Context) error {
	rs.stats.mu.Lock()
	rs.stats.lastCleanup = time.Now()
	rs.stats.mu.Unlock()

	// Get all long-term memories
	memoryType := domain.MemoryElement
	elements, err := rs.memoryRepo.List(domain.ElementFilter{
		Type: &memoryType,
	})
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	// Process each memory
	for _, elem := range elements {
		memory, ok := elem.(*domain.Memory)
		if !ok {
			continue
		}
		if err := rs.processMemory(ctx, memory); err != nil {
			// Log error but continue with other memories
			fmt.Printf("Failed to process memory %s: %v\n", memory.GetID(), err)
		}
	}

	return nil
}

// processMemory evaluates and applies retention policy to a single memory.
func (rs *MemoryRetentionService) processMemory(ctx context.Context, memory *domain.Memory) error {
	// Score the memory
	score, err := rs.scorer.Score(ctx, memory.Content)
	if err != nil {
		return fmt.Errorf("failed to score memory: %w", err)
	}

	rs.recordScore(score.Value)

	// Get appropriate retention policy
	policy := quality.GetRetentionPolicy(score.Value, rs.config.RetentionPolicies)
	if policy == nil {
		return fmt.Errorf("no retention policy found for score %.2f", score.Value)
	}

	// Calculate memory age
	age := time.Since(memory.GetMetadata().CreatedAt)
	ageDays := int(age.Hours() / 24)

	// Determine action based on policy
	switch {
	case ageDays >= policy.RetentionDays:
		// Delete old memories
		if err := rs.deleteMemory(ctx, memory); err != nil {
			return err
		}
		rs.recordDeletion()
		rs.recordPolicyUsage("deleted")
	case ageDays >= policy.ArchiveAfterDays:
		// Archive memories past archive threshold
		if err := rs.archiveMemory(ctx, memory); err != nil {
			return err
		}
		rs.recordArchival()
		rs.recordPolicyUsage(rs.getPolicyTier(policy))
	default:
		// Memory is still within active period
		rs.recordPolicyUsage(rs.getPolicyTier(policy))
	}

	return nil
}

// ScoreMemory scores a specific memory and returns the result.
func (rs *MemoryRetentionService) ScoreMemory(ctx context.Context, memoryID string) (*quality.Score, error) {
	elem, err := rs.memoryRepo.GetByID(memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	memory, ok := elem.(*domain.Memory)
	if !ok {
		return nil, errors.New("element is not a memory")
	}

	score, err := rs.scorer.Score(ctx, memory.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to score memory: %w", err)
	}

	rs.recordScore(score.Value)

	return score, nil
}

// ScoreMemoryWithSignals scores using implicit signals.
func (rs *MemoryRetentionService) ScoreMemoryWithSignals(
	ctx context.Context,
	memoryID string,
	signals quality.ImplicitSignals,
) (*quality.Score, error) {
	elem, err := rs.memoryRepo.GetByID(memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	memory, ok := elem.(*domain.Memory)
	if !ok {
		return nil, errors.New("element is not a memory")
	}

	// Check if scorer supports implicit signals
	if implicitScorer, ok := rs.scorer.(*quality.ImplicitScorer); ok {
		return implicitScorer.ScoreWithSignals(ctx, memory.Content, signals)
	}

	// Fallback to regular scoring
	return rs.scorer.Score(ctx, memory.Content)
}

// GetRetentionPolicy returns the policy for a given quality score.
func (rs *MemoryRetentionService) GetRetentionPolicy(score float64) *quality.RetentionPolicy {
	return quality.GetRetentionPolicy(score, rs.config.RetentionPolicies)
}

// archiveMemory marks a memory as archived.
func (rs *MemoryRetentionService) archiveMemory(ctx context.Context, memory *domain.Memory) error {
	// Update memory metadata to mark as archived
	if memory.Metadata == nil {
		memory.Metadata = make(map[string]string)
	}
	memory.Metadata["archived"] = "true"
	memory.Metadata["archived_at"] = time.Now().Format(time.RFC3339)

	// Update the memory
	if err := rs.memoryRepo.Update(memory); err != nil {
		return fmt.Errorf("failed to archive memory: %w", err)
	}

	return nil
}

// deleteMemory removes a memory from the system.
func (rs *MemoryRetentionService) deleteMemory(ctx context.Context, memory *domain.Memory) error {
	if err := rs.memoryRepo.Delete(memory.GetID()); err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}
	return nil
}

// getPolicyTier determines which tier a policy belongs to.
func (rs *MemoryRetentionService) getPolicyTier(policy *quality.RetentionPolicy) string {
	if policy.MinQuality >= 0.7 {
		return "high"
	} else if policy.MinQuality >= 0.5 {
		return "medium"
	}
	return "low"
}

// recordScore updates average quality score.
func (rs *MemoryRetentionService) recordScore(score float64) {
	rs.stats.mu.Lock()
	defer rs.stats.mu.Unlock()

	n := float64(rs.stats.totalScored)
	rs.stats.avgQualityScore = (rs.stats.avgQualityScore*n + score) / (n + 1)
	rs.stats.totalScored++
}

// recordArchival increments archival counter.
func (rs *MemoryRetentionService) recordArchival() {
	rs.stats.mu.Lock()
	defer rs.stats.mu.Unlock()
	rs.stats.totalArchived++
}

// recordDeletion increments deletion counter.
func (rs *MemoryRetentionService) recordDeletion() {
	rs.stats.mu.Lock()
	defer rs.stats.mu.Unlock()
	rs.stats.totalDeleted++
}

// recordPolicyUsage tracks which policies are used.
func (rs *MemoryRetentionService) recordPolicyUsage(tier string) {
	rs.stats.mu.Lock()
	defer rs.stats.mu.Unlock()
	rs.stats.policyBreakdown[tier]++
}

// GetStats returns current retention statistics.
func (rs *MemoryRetentionService) GetStats() map[string]interface{} {
	rs.stats.mu.RLock()
	defer rs.stats.mu.RUnlock()

	policyBreakdown := make(map[string]int)
	for k, v := range rs.stats.policyBreakdown {
		policyBreakdown[k] = v
	}

	return map[string]interface{}{
		"total_scored":      rs.stats.totalScored,
		"total_archived":    rs.stats.totalArchived,
		"total_deleted":     rs.stats.totalDeleted,
		"last_cleanup":      rs.stats.lastCleanup,
		"avg_quality_score": rs.stats.avgQualityScore,
		"policy_breakdown":  policyBreakdown,
		"running":           rs.running,
		"auto_archival":     rs.config.EnableAutoArchival,
		"cleanup_interval":  rs.config.CleanupIntervalMinutes,
	}
}

// ResetStats clears all statistics.
func (rs *MemoryRetentionService) ResetStats() {
	rs.stats.mu.Lock()
	defer rs.stats.mu.Unlock()

	rs.stats.totalScored = 0
	rs.stats.totalArchived = 0
	rs.stats.totalDeleted = 0
	rs.stats.avgQualityScore = 0
	rs.stats.policyBreakdown = make(map[string]int)
}
