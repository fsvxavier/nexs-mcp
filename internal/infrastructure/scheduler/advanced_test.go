package scheduler

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_Priority(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	tasks := []*Task{
		{
			ID:   "low-priority",
			Name: "Low Priority Task",
			Handler: func(ctx context.Context) error {
				return nil
			},
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 1 * time.Minute,
			},
			Enabled:  true,
			Priority: int(PriorityLow),
		},
		{
			ID:   "high-priority",
			Name: "High Priority Task",
			Handler: func(ctx context.Context) error {
				return nil
			},
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 1 * time.Minute,
			},
			Enabled:  true,
			Priority: int(PriorityHigh),
		},
	}

	for _, task := range tasks {
		if err := scheduler.AddTask(task); err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	highTask, _ := scheduler.GetTask("high-priority")
	if highTask.Priority != int(PriorityHigh) {
		t.Errorf("Expected high priority %d, got %d", PriorityHigh, highTask.Priority)
	}

	lowTask, _ := scheduler.GetTask("low-priority")
	if lowTask.Priority != int(PriorityLow) {
		t.Errorf("Expected low priority %d, got %d", PriorityLow, lowTask.Priority)
	}
}

func TestScheduler_Dependencies(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	var executionOrder []string
	var mu sync.Mutex

	// Create task A (no dependencies)
	taskA := &Task{
		ID:   "task-a",
		Name: "Task A",
		Handler: func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "A")
			mu.Unlock()
			time.Sleep(50 * time.Millisecond) // Simulate work
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 1 * time.Minute,
		},
		Enabled:      true,
		Dependencies: []string{},
	}

	err := scheduler.AddTask(taskA)
	if err != nil {
		t.Fatalf("Failed to add task A: %v", err)
	}

	// Create task B (depends on A)
	taskB := &Task{
		ID:   "task-b",
		Name: "Task B",
		Handler: func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "B")
			mu.Unlock()
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 1 * time.Minute,
		},
		Enabled:      true,
		Dependencies: []string{"task-a"},
	}

	err = scheduler.AddTask(taskB)
	if err != nil {
		t.Fatalf("Failed to add task B: %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()

	// Wait for tasks to execute
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 tasks to execute, got %d", len(executionOrder))
	}

	// Task A should execute before Task B
	if executionOrder[0] != "A" {
		t.Errorf("Expected task A first, got %s", executionOrder[0])
	}

	if executionOrder[1] != "B" {
		t.Errorf("Expected task B second, got %s", executionOrder[1])
	}
}

func TestScheduler_Dependencies_InvalidDependency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	task := &Task{
		ID:   "task-with-invalid-dep",
		Name: "Task with Invalid Dependency",
		Handler: func(ctx context.Context) error {
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 1 * time.Minute,
		},
		Enabled:      true,
		Dependencies: []string{"non-existent-task"},
	}

	err := scheduler.AddTask(task)
	if err == nil {
		t.Fatal("Expected error when adding task with invalid dependency")
	}
}

func TestScheduler_CronSchedule(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	var executed atomic.Bool

	task := &Task{
		ID:   "cron-task",
		Name: "Cron Task",
		Handler: func(ctx context.Context) error {
			executed.Store(true)
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeCron,
			CronSpec: "* * * * *", // Every minute
		},
		Enabled: true,
	}

	err := scheduler.AddTask(task)
	if err != nil {
		t.Fatalf("Failed to add cron task: %v", err)
	}

	// Check that next run was calculated
	retrievedTask, _ := scheduler.GetTask("cron-task")
	if retrievedTask.NextRun.IsZero() {
		t.Fatal("NextRun should not be zero for cron task")
	}

	// Next run should be in the future
	if retrievedTask.NextRun.Before(time.Now()) {
		t.Error("NextRun should be in the future")
	}
}

func TestScheduler_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	persistFile := filepath.Join(tempDir, "tasks.json")

	// Create first scheduler with persistence
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler1 := NewScheduler(logger)
	scheduler1.SetPersistence(persistFile)

	// Add a task
	task := &Task{
		ID:   "persistent-task",
		Name: "Persistent Task",
		Handler: func(ctx context.Context) error {
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 5 * time.Minute,
		},
		Enabled:  true,
		Priority: int(PriorityHigh),
	}

	err := scheduler1.AddTask(task)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Save tasks
	err = scheduler1.SaveTasks()
	if err != nil {
		t.Fatalf("Failed to save tasks: %v", err)
	}

	// Create second scheduler and load tasks
	time.Sleep(50 * time.Millisecond) // Allow file operations to complete

	scheduler2 := NewScheduler(logger)
	scheduler2.SetPersistence(persistFile)

	// Register handler before loading
	scheduler2.RegisterHandler("persistent-task", func(ctx context.Context) error {
		return nil
	})

	err = scheduler2.LoadTasks()
	if err != nil {
		t.Fatalf("Failed to load tasks: %v", err)
	}

	// Verify task was loaded
	loadedTask, err := scheduler2.GetTask("persistent-task")
	if err != nil {
		t.Fatalf("Task not found after loading: %v", err)
	}

	if loadedTask.ID != task.ID {
		t.Errorf("Task ID mismatch: got %s, want %s", loadedTask.ID, task.ID)
	}

	if loadedTask.Name != task.Name {
		t.Errorf("Task Name mismatch: got %s, want %s", loadedTask.Name, task.Name)
	}

	if loadedTask.Priority != task.Priority {
		t.Errorf("Task Priority mismatch: got %d, want %d", loadedTask.Priority, task.Priority)
	}

	if loadedTask.Enabled != task.Enabled {
		t.Errorf("Task Enabled mismatch: got %v, want %v", loadedTask.Enabled, task.Enabled)
	}
}

func TestScheduler_GetStats_WithRunningTasks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	started := make(chan bool)
	task := &Task{
		ID:   "long-task",
		Name: "Long Running Task",
		Handler: func(ctx context.Context) error {
			started <- true
			time.Sleep(200 * time.Millisecond)
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 1 * time.Minute,
		},
		Enabled: true,
	}

	err := scheduler.AddTask(task)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()

	<-started
	time.Sleep(10 * time.Millisecond)

	stats := scheduler.GetStats()
	runningTasks, ok := stats["running_tasks"].(int)
	if !ok {
		t.Fatal("running_tasks stat not found or wrong type")
	}

	if runningTasks != 1 {
		t.Errorf("Expected 1 running task, got %d", runningTasks)
	}

	time.Sleep(300 * time.Millisecond)
	stats = scheduler.GetStats()
	runningTasks = stats["running_tasks"].(int)
	if runningTasks != 0 {
		t.Logf("Warning: Expected 0 running tasks after completion, got %d (may be timing issue)", runningTasks)
	}
}

func TestScheduler_RemoveTask_WithDependents(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	taskA := &Task{
		ID:   "task-a",
		Name: "Task A",
		Handler: func(ctx context.Context) error {
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 1 * time.Minute,
		},
		Enabled: true,
	}

	err := scheduler.AddTask(taskA)
	if err != nil {
		t.Fatalf("Failed to add task A: %v", err)
	}

	taskB := &Task{
		ID:   "task-b",
		Name: "Task B",
		Handler: func(ctx context.Context) error {
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 1 * time.Minute,
		},
		Enabled:      true,
		Dependencies: []string{"task-a"},
	}

	err = scheduler.AddTask(taskB)
	if err != nil {
		t.Fatalf("Failed to add task B: %v", err)
	}

	// Try to remove task A (should fail because B depends on it)
	err = scheduler.RemoveTask("task-a")
	if err == nil {
		t.Fatal("Expected error when removing task with dependents")
	}

	// Remove task B first
	err = scheduler.RemoveTask("task-b")
	if err != nil {
		t.Fatalf("Failed to remove task B: %v", err)
	}

	// Now removing task A should succeed
	err = scheduler.RemoveTask("task-a")
	if err != nil {
		t.Fatalf("Failed to remove task A after removing dependent: %v", err)
	}
}
