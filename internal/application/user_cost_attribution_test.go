package application

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestUserCostAttributionService_NewService(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}

	if service.storageDir == "" {
		t.Error("Storage directory should be set")
	}

	// Clean up
	service.Stop()
}

func TestUserCostAttributionService_RecordOperation(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	service.RecordUserOperation(ctx, "user1", "session1", "test_tool", 100.0, 1000, 800, true)

	record, exists := service.GetUserCostRecord("user1")
	if !exists {
		t.Fatal("Expected user record to exist")
	}

	if record.TotalOperations != 1 {
		t.Errorf("Expected 1 operation, got %d", record.TotalOperations)
	}
	if record.TotalDuration != 100.0 {
		t.Errorf("Expected duration 100, got %f", record.TotalDuration)
	}
	if record.TotalTokens != 1000 {
		t.Errorf("Expected 1000 tokens, got %d", record.TotalTokens)
	}
	if record.TotalOptimizedTokens != 800 {
		t.Errorf("Expected 800 optimized tokens, got %d", record.TotalOptimizedTokens)
	}
	if record.ErrorCount != 0 {
		t.Errorf("Expected 0 errors, got %d", record.ErrorCount)
	}
}

func TestUserCostAttributionService_RecordError(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	service.RecordUserOperation(ctx, "user1", "", "tool1", 100, 100, 80, false) // Error

	record, _ := service.GetUserCostRecord("user1")
	if record.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", record.ErrorCount)
	}
}

func TestUserCostAttributionService_MultipleOperations(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	for range 10 {
		service.RecordUserOperation(ctx, "user1", "", "tool1", 50.0, 100, 80, true)
	}

	record, _ := service.GetUserCostRecord("user1")
	if record.TotalOperations != 10 {
		t.Errorf("Expected 10 operations, got %d", record.TotalOperations)
	}
	if record.TotalDuration != 500.0 {
		t.Errorf("Expected total duration 500, got %f", record.TotalDuration)
	}
	if record.TotalTokens != 1000 {
		t.Errorf("Expected 1000 total tokens, got %d", record.TotalTokens)
	}
}

func TestUserCostAttributionService_MultipleTools(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	service.RecordUserOperation(ctx, "user1", "", "tool1", 100, 500, 400, true)
	service.RecordUserOperation(ctx, "user1", "", "tool2", 200, 1000, 800, true)

	record, _ := service.GetUserCostRecord("user1")
	if len(record.OperationsByTool) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(record.OperationsByTool))
	}
	if record.OperationsByTool["tool1"] != 1 {
		t.Errorf("Expected 1 operation for tool1, got %d", record.OperationsByTool["tool1"])
	}
	if record.DurationByTool["tool2"] != 200 {
		t.Errorf("Expected duration 200 for tool2, got %f", record.DurationByTool["tool2"])
	}
}

func TestUserCostAttributionService_GetUserCostSummary(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	for range 20 {
		service.RecordUserOperation(ctx, "user1", "", "tool1", 100, 1000, 700, true)
	}

	summary, err := service.GetUserCostSummary("user1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if summary.UserID != "user1" {
		t.Errorf("Expected user ID 'user1', got '%s'", summary.UserID)
	}
	if summary.TotalOperations != 20 {
		t.Errorf("Expected 20 operations, got %d", summary.TotalOperations)
	}
	if summary.TokenSavings != 6000 { // 20 * (1000 - 700)
		t.Errorf("Expected 6000 token savings, got %d", summary.TokenSavings)
	}
	if summary.TokenSavingsPercent != 30.0 {
		t.Errorf("Expected 30%% savings, got %f", summary.TokenSavingsPercent)
	}
	if summary.CostScore < 0 || summary.CostScore > 100 {
		t.Errorf("Cost score should be 0-100, got %f", summary.CostScore)
	}
}

func TestUserCostAttributionService_GetUserCostSummary_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	_, err := service.GetUserCostSummary("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent user")
	}
}

func TestUserCostAttributionService_GetAllUsers(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	service.RecordUserOperation(ctx, "user1", "", "tool1", 100, 100, 80, true)
	service.RecordUserOperation(ctx, "user2", "", "tool1", 100, 100, 80, true)
	service.RecordUserOperation(ctx, "user3", "", "tool1", 100, 100, 80, true)

	users := service.GetAllUsers()
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}

func TestUserCostAttributionService_GetTopUsers(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	// User 1: low cost
	service.RecordUserOperation(ctx, "user1", "", "tool1", 50, 100, 80, true)

	// User 2: high cost
	for range 100 {
		service.RecordUserOperation(ctx, "user2", "", "tool1", 500, 10000, 8000, true)
	}

	// User 3: medium cost
	for range 10 {
		service.RecordUserOperation(ctx, "user3", "", "tool1", 200, 1000, 800, true)
	}

	topUsers, err := service.GetTopUsers(2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(topUsers) != 2 {
		t.Errorf("Expected 2 top users, got %d", len(topUsers))
	}

	// Should be sorted by cost score
	if topUsers[0].CostScore < topUsers[1].CostScore {
		t.Error("Top users should be sorted by cost score descending")
	}

	// User2 should be top (highest cost - 100 operations with 10000 tokens each = 1M tokens total)
	if topUsers[0].UserID != "user2" {
		t.Errorf("Expected user2 as top user, got %s", topUsers[0].UserID)
	}
}

func TestUserCostAttributionService_GetUsersByDateRange(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	now := time.Now()

	service.RecordUserOperation(ctx, "user1", "", "tool1", 100, 100, 80, true)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	service.RecordUserOperation(ctx, "user2", "", "tool1", 100, 100, 80, true)

	users := service.GetUsersByDateRange(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if len(users) != 2 {
		t.Errorf("Expected 2 users in range, got %d", len(users))
	}
}

func TestUserCostAttributionService_ClearUser(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	service.RecordUserOperation(ctx, "user1", "", "tool1", 100, 100, 80, true)

	err := service.ClearUser("user1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, exists := service.GetUserCostRecord("user1")
	if exists {
		t.Error("User record should be deleted")
	}
}

func TestUserCostAttributionService_ClearUser_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	err := service.ClearUser("nonexistent")
	if err == nil {
		t.Error("Expected error when clearing nonexistent user")
	}
}

func TestUserCostAttributionService_UpdateMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	service.RecordUserOperation(ctx, "user1", "", "tool1", 100, 100, 80, true)

	metadata := map[string]string{
		"team": "engineering",
		"role": "developer",
	}

	err := service.UpdateUserMetadata("user1", metadata)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	record, _ := service.GetUserCostRecord("user1")
	if record.Metadata["team"] != "engineering" {
		t.Error("Metadata not updated correctly")
	}
}

func TestUserCostAttributionService_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create service and add data
	service1 := NewUserCostAttributionService(tmpDir, true)
	ctx := context.Background()
	service1.RecordUserOperation(ctx, "user1", "", "tool1", 100, 1000, 800, true)
	service1.RecordUserOperation(ctx, "user2", "", "tool2", 200, 2000, 1600, true)

	// Save and stop
	err := service1.Stop()
	if err != nil {
		t.Fatalf("Error stopping service: %v", err)
	}

	// Create new service and verify data loaded
	service2 := NewUserCostAttributionService(tmpDir, false)
	defer service2.Stop()

	record1, exists1 := service2.GetUserCostRecord("user1")
	if !exists1 {
		t.Error("User1 should exist after reload")
	}
	if record1.TotalTokens != 1000 {
		t.Errorf("Expected 1000 tokens for user1, got %d", record1.TotalTokens)
	}

	record2, exists2 := service2.GetUserCostRecord("user2")
	if !exists2 {
		t.Error("User2 should exist after reload")
	}
	if record2.TotalDuration != 200.0 {
		t.Errorf("Expected duration 200 for user2, got %f", record2.TotalDuration)
	}
}

func TestUserCostAttributionService_SessionFallback(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	// Record with empty userID, should use sessionID as key
	service.RecordUserOperation(ctx, "", "session123", "tool1", 100, 100, 80, true)

	record, exists := service.GetUserCostRecord("session123")
	if !exists {
		t.Error("Session record should exist")
	}
	if record.SessionID != "session123" {
		t.Errorf("Expected session ID 'session123', got '%s'", record.SessionID)
	}
}

func TestUserCostAttributionService_NoIdentifier(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	// Record with both empty - should be ignored
	service.RecordUserOperation(ctx, "", "", "tool1", 100, 100, 80, true)

	users := service.GetAllUsers()
	if len(users) != 0 {
		t.Error("No users should be recorded without identifier")
	}
}

// Test concurrent access.
func TestUserCostAttributionService_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Multiple goroutines recording operations
	for i := range 10 {
		wg.Add(1)
		go func(userNum int) {
			defer wg.Done()
			for range 100 {
				service.RecordUserOperation(
					ctx,
					"user1",
					"",
					"tool1",
					100,
					100,
					80,
					true,
				)
			}
		}(i)
	}

	wg.Wait()

	record, _ := service.GetUserCostRecord("user1")
	if record.TotalOperations != 1000 {
		t.Errorf("Expected 1000 operations from concurrent access, got %d", record.TotalOperations)
	}
}

// Test race conditions with -race flag.
func TestUserCostAttributionService_RaceConditions(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewUserCostAttributionService(tmpDir, false)
	defer service.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent writes
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service.RecordUserOperation(ctx, "user1", "", "tool1", 100, 100, 80, true)
		}()
	}

	// Concurrent reads
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service.GetUserCostRecord("user1")
		}()
	}

	// Concurrent summaries
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service.GetUserCostSummary("user1")
		}()
	}

	wg.Wait()
}
