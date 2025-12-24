package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_AddTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	task := &Task{
		ID:      "test-task",
		Name:    "Test Task",
		Handler: func(ctx context.Context) error { return nil },
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

	// Try to add duplicate
	err = scheduler.AddTask(task)
	if err == nil {
		t.Fatal("Expected error when adding duplicate task")
	}
}

func TestScheduler_AddTask_Validation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{
			name: "missing ID",
			task: &Task{
				Name:    "Test",
				Handler: func(ctx context.Context) error { return nil },
				Schedule: Schedule{
					Type:     ScheduleTypeInterval,
					Interval: 1 * time.Minute,
				},
			},
			wantErr: true,
		},
		{
			name: "missing handler",
			task: &Task{
				ID:   "test",
				Name: "Test",
				Schedule: Schedule{
					Type:     ScheduleTypeInterval,
					Interval: 1 * time.Minute,
				},
			},
			wantErr: true,
		},
		{
			name: "valid task",
			task: &Task{
				ID:      "test",
				Name:    "Test",
				Handler: func(ctx context.Context) error { return nil },
				Schedule: Schedule{
					Type:     ScheduleTypeInterval,
					Interval: 1 * time.Minute,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scheduler.AddTask(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScheduler_RemoveTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	task := &Task{
		ID:      "test-task",
		Name:    "Test Task",
		Handler: func(ctx context.Context) error { return nil },
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

	err = scheduler.RemoveTask("test-task")
	if err != nil {
		t.Fatalf("Failed to remove task: %v", err)
	}

	// Try to remove non-existent task
	err = scheduler.RemoveTask("non-existent")
	if err == nil {
		t.Fatal("Expected error when removing non-existent task")
	}
}

func TestScheduler_EnableDisableTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	task := &Task{
		ID:      "test-task",
		Name:    "Test Task",
		Handler: func(ctx context.Context) error { return nil },
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

	// Disable task
	err = scheduler.DisableTask("test-task")
	if err != nil {
		t.Fatalf("Failed to disable task: %v", err)
	}

	retrievedTask, _ := scheduler.GetTask("test-task")
	if retrievedTask.Enabled {
		t.Fatal("Task should be disabled")
	}

	// Enable task
	err = scheduler.EnableTask("test-task")
	if err != nil {
		t.Fatalf("Failed to enable task: %v", err)
	}

	retrievedTask, _ = scheduler.GetTask("test-task")
	if !retrievedTask.Enabled {
		t.Fatal("Task should be enabled")
	}
}

func TestScheduler_GetTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	task := &Task{
		ID:      "test-task",
		Name:    "Test Task",
		Handler: func(ctx context.Context) error { return nil },
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

	retrievedTask, err := scheduler.GetTask("test-task")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if retrievedTask.ID != task.ID {
		t.Fatalf("Expected task ID %s, got %s", task.ID, retrievedTask.ID)
	}

	// Try to get non-existent task
	_, err = scheduler.GetTask("non-existent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent task")
	}
}

func TestScheduler_ListTasks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	tasks := []*Task{
		{
			ID:      "task-1",
			Name:    "Task 1",
			Handler: func(ctx context.Context) error { return nil },
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 1 * time.Minute,
			},
			Enabled: true,
		},
		{
			ID:      "task-2",
			Name:    "Task 2",
			Handler: func(ctx context.Context) error { return nil },
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 2 * time.Minute,
			},
			Enabled: false,
		},
	}

	for _, task := range tasks {
		err := scheduler.AddTask(task)
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	listedTasks := scheduler.ListTasks()
	if len(listedTasks) != len(tasks) {
		t.Fatalf("Expected %d tasks, got %d", len(tasks), len(listedTasks))
	}
}

func TestScheduler_IntervalExecution(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	var counter atomic.Int32

	task := &Task{
		ID:   "interval-task",
		Name: "Interval Task",
		Handler: func(ctx context.Context) error {
			counter.Add(1)
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 100 * time.Millisecond,
		},
		Enabled: true,
	}

	err := scheduler.AddTask(task)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()

	// Wait for task to run multiple times
	time.Sleep(350 * time.Millisecond)

	count := counter.Load()
	if count < 2 {
		t.Fatalf("Expected task to run at least 2 times, ran %d times", count)
	}
}

func TestScheduler_OnceExecution(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	var counter atomic.Int32

	task := &Task{
		ID:   "once-task",
		Name: "Once Task",
		Handler: func(ctx context.Context) error {
			counter.Add(1)
			return nil
		},
		Schedule: Schedule{
			Type:  ScheduleTypeOnce,
			Times: []time.Time{time.Now().Add(50 * time.Millisecond)},
		},
		Enabled: true,
	}

	err := scheduler.AddTask(task)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()

	// Wait for task to run
	time.Sleep(200 * time.Millisecond)

	count := counter.Load()
	if count != 1 {
		t.Fatalf("Expected task to run exactly once, ran %d times", count)
	}

	// Verify task was disabled
	retrievedTask, _ := scheduler.GetTask("once-task")
	if retrievedTask.Enabled {
		t.Fatal("Once task should be disabled after execution")
	}
}

func TestScheduler_TaskRetry(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)
	scheduler.SetMaxRetries(2)
	scheduler.SetRetryDelay(50 * time.Millisecond)

	var attempts atomic.Int32

	task := &Task{
		ID:   "failing-task",
		Name: "Failing Task",
		Handler: func(ctx context.Context) error {
			count := attempts.Add(1)
			if count < 3 {
				return errors.New("simulated error")
			}
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 1 * time.Minute, // Long interval so it only runs once during test
		},
		Enabled: true,
	}

	err := scheduler.AddTask(task)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()

	// Wait for task to run with retries
	time.Sleep(300 * time.Millisecond)

	count := attempts.Load()
	if count != 3 {
		t.Fatalf("Expected 3 attempts (1 initial + 2 retries), got %d", count)
	}

	// Verify task stats
	retrievedTask, _ := scheduler.GetTask("failing-task")
	if retrievedTask.RunCount != 1 {
		t.Fatalf("Expected RunCount=1, got %d", retrievedTask.RunCount)
	}
	if retrievedTask.ErrorCount != 0 {
		t.Fatalf("Expected ErrorCount=0 (succeeded on retry), got %d", retrievedTask.ErrorCount)
	}
}

func TestScheduler_TaskFailureAfterRetries(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)
	scheduler.SetMaxRetries(2)
	scheduler.SetRetryDelay(50 * time.Millisecond)

	var attempts atomic.Int32

	task := &Task{
		ID:   "always-failing-task",
		Name: "Always Failing Task",
		Handler: func(ctx context.Context) error {
			attempts.Add(1)
			return errors.New("permanent error")
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

	// Wait for task to run with retries
	time.Sleep(300 * time.Millisecond)

	count := attempts.Load()
	if count != 3 {
		t.Fatalf("Expected 3 attempts (1 initial + 2 retries), got %d", count)
	}

	// Verify task recorded the error
	retrievedTask, _ := scheduler.GetTask("always-failing-task")
	if retrievedTask.ErrorCount != 1 {
		t.Fatalf("Expected ErrorCount=1, got %d", retrievedTask.ErrorCount)
	}
	if retrievedTask.LastError == nil {
		t.Fatal("Expected LastError to be set")
	}
}

func TestScheduler_GetStats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	task1 := &Task{
		ID:      "task-1",
		Name:    "Task 1",
		Handler: func(ctx context.Context) error { return nil },
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 1 * time.Minute,
		},
		Enabled: true,
	}

	task2 := &Task{
		ID:      "task-2",
		Name:    "Task 2",
		Handler: func(ctx context.Context) error { return errors.New("error") },
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 2 * time.Minute,
		},
		Enabled: false,
	}

	scheduler.AddTask(task1)
	scheduler.AddTask(task2)

	stats := scheduler.GetStats()

	if stats["total_tasks"].(int) != 2 {
		t.Fatalf("Expected 2 total tasks, got %v", stats["total_tasks"])
	}

	if stats["enabled_tasks"].(int) != 1 {
		t.Fatalf("Expected 1 enabled task, got %v", stats["enabled_tasks"])
	}
}

func TestScheduler_GracefulShutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)

	var started atomic.Bool
	var completed atomic.Bool

	task := &Task{
		ID:   "long-task",
		Name: "Long Task",
		Handler: func(ctx context.Context) error {
			started.Store(true)
			time.Sleep(200 * time.Millisecond)
			completed.Store(true)
			return nil
		},
		Schedule: Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 50 * time.Millisecond,
		},
		Enabled: true,
	}

	scheduler.AddTask(task)
	scheduler.Start()

	// Wait for task to start
	time.Sleep(100 * time.Millisecond)

	// Stop scheduler (should wait for running task)
	scheduler.Stop()

	if !started.Load() {
		t.Fatal("Task should have started")
	}

	if !completed.Load() {
		t.Fatal("Task should have completed before shutdown")
	}
}

func TestScheduler_ConcurrentModification(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	scheduler := NewScheduler(logger)
	scheduler.Start()
	defer scheduler.Stop()

	// Add tasks concurrently
	done := make(chan bool)
	for i := range 10 {
		go func(id int) {
			task := &Task{
				ID:   string(rune('a' + id)),
				Name: "Concurrent Task",
				Handler: func(ctx context.Context) error {
					return nil
				},
				Schedule: Schedule{
					Type:     ScheduleTypeInterval,
					Interval: 1 * time.Second,
				},
				Enabled: true,
			}
			scheduler.AddTask(task)
			done <- true
		}(i)
	}

	// Wait for all adds to complete
	for range 10 {
		<-done
	}

	// List tasks (concurrent read)
	tasks := scheduler.ListTasks()
	if len(tasks) != 10 {
		t.Fatalf("Expected 10 tasks, got %d", len(tasks))
	}
}
