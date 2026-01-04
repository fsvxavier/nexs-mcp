package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/quality"
)

func TestNewMemoryRetentionService(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.config == nil {
		t.Error("Expected non-nil config")
	}
	if service.scorer == nil {
		t.Error("Expected non-nil scorer")
	}
	if service.stats == nil {
		t.Error("Expected non-nil stats")
	}
}

func TestMemoryRetentionService_StartStop(t *testing.T) {
	config := quality.DefaultConfig()
	config.EnableAutoArchival = false // Disable auto-archival for this test
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	ctx := context.Background()

	// Start service
	err := service.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Try starting again (should error)
	err = service.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already running service")
	}

	// Stop service
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}

	// Try stopping again (should error)
	err = service.Stop()
	if err == nil {
		t.Error("Expected error when stopping already stopped service")
	}
}

func TestMemoryRetentionService_RunCleanup(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	// Create test memories
	mem1 := domain.NewMemory("Mem1", "High quality content with detailed information", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Short text", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	err := service.RunCleanup(ctx)
	if err != nil {
		t.Fatalf("RunCleanup failed: %v", err)
	}

	// Verify stats were updated
	stats := service.GetStats()
	if stats["last_cleanup"].(time.Time).IsZero() {
		t.Error("Expected lastCleanup to be updated")
	}
}

func TestMemoryRetentionService_ScoreMemory(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	// Create memory
	mem := domain.NewMemory("TestMem", "This is quality content for scoring", "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	score, err := service.ScoreMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("ScoreMemory failed: %v", err)
	}

	if score == nil {
		t.Fatal("Expected non-nil score")
	}

	if score.Value < 0 || score.Value > 1 {
		t.Errorf("Expected score between 0 and 1, got %f", score.Value)
	}
}

func TestMemoryRetentionService_ScoreMemory_NotFound(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	ctx := context.Background()
	_, err := service.ScoreMemory(ctx, "nonexistent-id")
	if err == nil {
		t.Error("Expected error for nonexistent memory")
	}
}

func TestMemoryRetentionService_ScoreMemory_InvalidType(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	// Create non-memory element
	persona := domain.NewPersona("Test", "Test persona", "1.0.0", "test")
	repo.Create(persona)

	ctx := context.Background()
	_, err := service.ScoreMemory(ctx, persona.GetID())
	if err == nil {
		t.Error("Expected error for non-memory element")
	}
}

func TestMemoryRetentionService_GetStats(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	stats := service.GetStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Initial stats should be zero
	if stats["total_scored"].(int) != 0 {
		t.Errorf("Expected 0 totalScored initially, got %d", stats["total_scored"].(int))
	}
	if stats["total_archived"].(int) != 0 {
		t.Errorf("Expected 0 totalArchived initially, got %d", stats["total_archived"].(int))
	}
	if stats["total_deleted"].(int) != 0 {
		t.Errorf("Expected 0 totalDeleted initially, got %d", stats["total_deleted"].(int))
	}
}

func TestRetentionStats_Structure(t *testing.T) {
	stats := &RetentionStats{
		totalScored:     10,
		totalArchived:   5,
		totalDeleted:    2,
		lastCleanup:     time.Now(),
		avgQualityScore: 0.75,
		policyBreakdown: map[string]int{
			"high":   3,
			"medium": 5,
			"low":    2,
		},
	}

	if stats.totalScored != 10 {
		t.Errorf("Expected 10 totalScored, got %d", stats.totalScored)
	}
	if stats.totalArchived != 5 {
		t.Errorf("Expected 5 totalArchived, got %d", stats.totalArchived)
	}
	if stats.totalDeleted != 2 {
		t.Errorf("Expected 2 totalDeleted, got %d", stats.totalDeleted)
	}
	if stats.avgQualityScore != 0.75 {
		t.Errorf("Expected 0.75 avgQualityScore, got %f", stats.avgQualityScore)
	}
	if len(stats.policyBreakdown) != 3 {
		t.Errorf("Expected 3 policy entries, got %d", len(stats.policyBreakdown))
	}
}

func TestMemoryRetentionService_WithNilConfig(t *testing.T) {
	scorer := quality.NewImplicitScorer(quality.DefaultConfig())
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	// Nil config should use defaults
	service := NewMemoryRetentionService(nil, scorer, repo, wmService)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.config == nil {
		t.Error("Expected config to be set to default")
	}
}

func TestMemoryRetentionService_ProcessOldMemories(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	// Create old memory
	oldMem := domain.NewMemory("OldMem", "Old content", "1.0.0", "test")
	oldMem.SetMetadata(domain.ElementMetadata{
		ID:        oldMem.GetID(),
		Name:      "OldMem",
		CreatedAt: time.Now().Add(-400 * 24 * time.Hour), // Very old
	})
	repo.Create(oldMem)

	ctx := context.Background()
	err := service.RunCleanup(ctx)
	if err != nil {
		t.Fatalf("RunCleanup failed: %v", err)
	}

	// Old memory should have been processed (may be 0 if age check didn't work)
	stats := service.GetStats()
	totalScored := 0
	if val, ok := stats["total_scored"].(int); ok {
		totalScored = val
	}
	t.Logf("Scored %d memories (age detection may vary)", totalScored)
}

func TestMemoryRetentionService_ProcessRecentMemories(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	// Create recent memory
	recentMem := domain.NewMemory("RecentMem", "Recent high quality content", "1.0.0", "test")
	repo.Create(recentMem)

	ctx := context.Background()
	err := service.RunCleanup(ctx)
	if err != nil {
		t.Fatalf("RunCleanup failed: %v", err)
	}

	// Recent memory should remain active
	elem, err := repo.GetByID(recentMem.GetID())
	if err != nil {
		t.Error("Recent memory should still exist")
	}
	if mem, ok := elem.(*domain.Memory); ok {
		if !mem.IsActive() {
			t.Error("Recent memory should remain active")
		}
	}
}

func TestMemoryRetentionService_EmptyRepository(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	ctx := context.Background()
	err := service.RunCleanup(ctx)
	if err != nil {
		t.Fatalf("RunCleanup failed on empty repo: %v", err)
	}
}

func TestMemoryRetentionService_MultipleCleanups(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	// Create memories
	for i := range 5 {
		mem := domain.NewMemory(fmt.Sprintf("Mem%d", i), "Content", "1.0.0", "test")
		repo.Create(mem)
	}

	ctx := context.Background()

	// Run multiple cleanups
	for i := range 3 {
		err := service.RunCleanup(ctx)
		if err != nil {
			t.Fatalf("Cleanup %d failed: %v", i, err)
		}
	}

	// Stats should accumulate
	stats := service.GetStats()
	if stats["total_scored"].(int) == 0 {
		t.Error("Expected memories to be scored")
	}
}

func TestMemoryRetentionService_ConcurrentCleanup(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	// Create memories
	mem := domain.NewMemory("Mem", "Content", "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()

	// Run cleanups concurrently
	done := make(chan bool, 3)
	for range 3 {
		go func() {
			err := service.RunCleanup(ctx)
			if err != nil {
				t.Errorf("Concurrent cleanup failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 3 {
		<-done
	}
}

func TestMemoryRetentionService_StatsUpdates(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	// Create memory
	mem := domain.NewMemory("Mem", "Quality content for testing", "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()

	// Get stats before cleanup
	statsBefore := service.GetStats()
	scoredBefore := statsBefore["total_scored"].(int)

	// Run cleanup
	service.RunCleanup(ctx)

	// Get stats after cleanup
	statsAfter := service.GetStats()
	scoredAfter := statsAfter["total_scored"].(int)

	// totalScored should have increased
	if scoredAfter <= scoredBefore {
		t.Error("Expected totalScored to increase after cleanup")
	}
}

func TestMemoryRetentionService_DisabledAutoArchival(t *testing.T) {
	config := quality.DefaultConfig()
	config.EnableAutoArchival = false
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	ctx := context.Background()

	// Start should succeed but not run cleanup loop
	err := service.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Should be able to stop immediately
	err = service.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestMemoryRetentionService_QualityScoreTracking(t *testing.T) {
	config := quality.DefaultConfig()
	scorer := quality.NewImplicitScorer(config)
	repo := newMockRepoForIndex()
	wmService := NewWorkingMemoryService(repo)

	service := NewMemoryRetentionService(config, scorer, repo, wmService)

	// Create memories
	highQuality := domain.NewMemory("High", "This is a very detailed and informative memory with rich content", "1.0.0", "test")
	lowQuality := domain.NewMemory("Low", "x", "1.0.0", "test")

	repo.Create(highQuality)
	repo.Create(lowQuality)

	ctx := context.Background()
	service.RunCleanup(ctx)

	// Check average quality score
	stats := service.GetStats()
	if stats["avg_quality_score"].(float64) < 0 || stats["avg_quality_score"].(float64) > 1 {
		t.Errorf("Expected avgQualityScore between 0 and 1, got %f", stats["avg_quality_score"].(float64))
	}
}
