package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

// TaskPriority defines task execution priority.
type TaskPriority int

const (
	PriorityLow    TaskPriority = 0
	PriorityMedium TaskPriority = 50
	PriorityHigh   TaskPriority = 100
)

// Task represents a scheduled task.
type Task struct {
	ID           string
	Name         string
	Description  string
	Handler      TaskHandler
	Schedule     Schedule
	Enabled      bool
	Priority     int      // Higher values = higher priority
	Dependencies []string // IDs of tasks that must complete first
	LastRun      time.Time
	NextRun      time.Time
	RunCount     int64
	ErrorCount   int64
	LastError    error
}

// TaskHandler is the function signature for task execution.
type TaskHandler func(ctx context.Context) error

// Schedule defines when a task should run.
type Schedule struct {
	Type     ScheduleType
	Interval time.Duration // For interval-based scheduling
	CronSpec string        // For cron-like scheduling
	Times    []time.Time   // For one-time or specific times
}

// ScheduleType defines the type of scheduling.
type ScheduleType string

const (
	ScheduleTypeInterval ScheduleType = "interval" // Run every X duration
	ScheduleTypeOnce     ScheduleType = "once"     // Run once at specific time
	ScheduleTypeCron     ScheduleType = "cron"     // Cron-like scheduling
)

// Scheduler manages and executes scheduled tasks.
type Scheduler struct {
	tasks        map[string]*Task
	handlers     map[string]TaskHandler // Separate handler storage for persistence
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	logger       *slog.Logger
	maxRetries   int
	retryDelay   time.Duration
	persistence  *TaskPersistence
	runningTasks map[string]bool // Track currently running tasks
	runningMu    sync.RWMutex
}

// NewScheduler creates a new task scheduler.
func NewScheduler(logger *slog.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		tasks:        make(map[string]*Task),
		handlers:     make(map[string]TaskHandler),
		ctx:          ctx,
		cancel:       cancel,
		logger:       logger,
		maxRetries:   3,
		retryDelay:   5 * time.Second,
		runningTasks: make(map[string]bool),
	}
}

// SetPersistence enables task persistence.
func (s *Scheduler) SetPersistence(filePath string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.persistence = NewTaskPersistence(filePath)
}

// RegisterHandler registers a handler for a task ID (needed for persistence).
func (s *Scheduler) RegisterHandler(taskID string, handler TaskHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handlers[taskID] = handler
}

// SaveTasks saves current tasks to disk.
func (s *Scheduler) SaveTasks() error {
	if s.persistence == nil {
		return errors.New("persistence not enabled")
	}

	s.mu.RLock()
	tasks := make(map[string]*Task, len(s.tasks))
	for k, v := range s.tasks {
		tasks[k] = v
	}
	s.mu.RUnlock()

	return s.persistence.Save(tasks)
}

// LoadTasks loads tasks from disk and restores handlers.
func (s *Scheduler) LoadTasks() error {
	if s.persistence == nil {
		return errors.New("persistence not enabled")
	}

	serializedTasks, err := s.persistence.Load()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, st := range serializedTasks {
		task, err := ConvertToTask(st)
		if err != nil {
			s.logger.Warn("Failed to convert task", "task_id", st.ID, "error", err)
			continue
		}

		// Restore handler if registered
		if handler, exists := s.handlers[task.ID]; exists {
			task.Handler = handler
			s.tasks[task.ID] = task

			s.logger.Info("Task loaded from persistence",
				"task_id", task.ID,
				"name", task.Name,
				"enabled", task.Enabled,
			)
		} else {
			s.logger.Warn("Handler not found for task", "task_id", task.ID)
		}
	}

	return nil
}

// AddTask adds a new task to the scheduler.
func (s *Scheduler) AddTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task.ID == "" {
		return errors.New("task ID cannot be empty")
	}

	if task.Handler == nil {
		return errors.New("task handler cannot be nil")
	}

	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	// Validate dependencies
	if err := s.validateDependencies(task.Dependencies); err != nil {
		return fmt.Errorf("invalid dependencies: %w", err)
	}

	// Calculate next run time - run immediately for interval tasks on first schedule
	now := time.Now()
	if task.Schedule.Type == ScheduleTypeInterval {
		task.NextRun = now // Run immediately
	} else {
		nextRun, err := s.calculateNextRunForSchedule(task.Schedule, now)
		if err != nil {
			return fmt.Errorf("failed to calculate next run: %w", err)
		}
		task.NextRun = nextRun
	}

	s.tasks[task.ID] = task
	s.handlers[task.ID] = task.Handler

	s.logger.Info("Task added to scheduler",
		"task_id", task.ID,
		"name", task.Name,
		"schedule_type", task.Schedule.Type,
		"priority", task.Priority,
		"next_run", task.NextRun,
	)

	// Auto-save if persistence enabled
	if s.persistence != nil {
		go func() {
			if err := s.SaveTasks(); err != nil {
				s.logger.Error("Failed to auto-save tasks", "error", err)
			}
		}()
	}

	return nil
}

// validateDependencies checks if all dependencies exist.
func (s *Scheduler) validateDependencies(deps []string) error {
	for _, depID := range deps {
		if _, exists := s.tasks[depID]; !exists {
			return fmt.Errorf("dependency task not found: %s", depID)
		}
	}
	return nil
}

// RemoveTask removes a task from the scheduler.
func (s *Scheduler) RemoveTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[taskID]; !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// Check if other tasks depend on this one
	for _, task := range s.tasks {
		for _, dep := range task.Dependencies {
			if dep == taskID {
				return fmt.Errorf("cannot remove task: %s depends on it", task.ID)
			}
		}
	}

	delete(s.tasks, taskID)
	delete(s.handlers, taskID)

	s.logger.Info("Task removed from scheduler", "task_id", taskID)

	// Auto-save if persistence enabled
	if s.persistence != nil {
		go func() {
			if err := s.SaveTasks(); err != nil {
				s.logger.Error("Failed to auto-save tasks", "error", err)
			}
		}()
	}

	return nil
}

// EnableTask enables a task.
func (s *Scheduler) EnableTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	task.Enabled = true
	s.logger.Info("Task enabled", "task_id", taskID)

	return nil
}

// DisableTask disables a task.
func (s *Scheduler) DisableTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	task.Enabled = false
	s.logger.Info("Task disabled", "task_id", taskID)

	return nil
}

// GetTask returns task information.
func (s *Scheduler) GetTask(taskID string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	// Return a copy to avoid external modifications
	taskCopy := *task
	taskCopy.Dependencies = make([]string, len(task.Dependencies))
	copy(taskCopy.Dependencies, task.Dependencies)
	return &taskCopy, nil
}

// ListTasks returns all tasks.
func (s *Scheduler) ListTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		taskCopy := *task
		taskCopy.Dependencies = make([]string, len(task.Dependencies))
		copy(taskCopy.Dependencies, task.Dependencies)
		tasks = append(tasks, &taskCopy)
	}

	return tasks
}

// Start begins the scheduler.
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()

	s.logger.Info("Scheduler started", "task_count", len(s.tasks))
}

// Stop stops the scheduler gracefully.
func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()

	s.logger.Info("Scheduler stopped")

	// Final save if persistence enabled
	if s.persistence != nil {
		if err := s.SaveTasks(); err != nil {
			s.logger.Error("Failed to save tasks on shutdown", "error", err)
		}
	}
}

// run is the main scheduler loop.
func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case now := <-ticker.C:
			s.checkAndRunTasks(now)
		}
	}
}

// checkAndRunTasks checks for tasks that need to run.
func (s *Scheduler) checkAndRunTasks(now time.Time) {
	s.mu.Lock()

	// Find tasks that are ready to run
	var readyTasks []*Task
	for _, task := range s.tasks {
		if !task.Enabled || task.NextRun.After(now) {
			continue
		}

		// Check if task is already running
		s.runningMu.RLock()
		isRunning := s.runningTasks[task.ID]
		s.runningMu.RUnlock()

		if isRunning {
			continue
		}

		// Check if all dependencies have completed successfully
		if s.areDependenciesMet(task) {
			readyTasks = append(readyTasks, task)
			// Mark as running
			task.NextRun = now.Add(365 * 24 * time.Hour) // Far future
		}
	}

	// Sort by priority (higher priority first)
	sort.Slice(readyTasks, func(i, j int) bool {
		return readyTasks[i].Priority > readyTasks[j].Priority
	})

	s.mu.Unlock()

	// Execute tasks in priority order
	for _, task := range readyTasks {
		s.runningMu.Lock()
		s.runningTasks[task.ID] = true
		s.runningMu.Unlock()

		s.wg.Add(1)
		go s.executeTask(task)
	}
}

// areDependenciesMet checks if all task dependencies have completed successfully.
func (s *Scheduler) areDependenciesMet(task *Task) bool {
	if len(task.Dependencies) == 0 {
		return true
	}

	for _, depID := range task.Dependencies {
		depTask, exists := s.tasks[depID]
		if !exists {
			return false
		}

		// Dependency must have run at least once and not have errors
		if depTask.RunCount == 0 || depTask.LastError != nil {
			return false
		}

		// Check if dependency is currently running
		s.runningMu.RLock()
		isRunning := s.runningTasks[depID]
		s.runningMu.RUnlock()

		if isRunning {
			return false
		}
	}

	return true
}

// executeTask executes a task with retry logic.
func (s *Scheduler) executeTask(task *Task) {
	defer s.wg.Done()
	defer func() {
		s.runningMu.Lock()
		delete(s.runningTasks, task.ID)
		s.runningMu.Unlock()
	}()

	s.logger.Debug("Executing task",
		"task_id", task.ID,
		"name", task.Name,
		"priority", task.Priority,
	)

	startTime := time.Now()
	var lastErr error

	// Retry logic
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			s.logger.Warn("Retrying task",
				"task_id", task.ID,
				"attempt", attempt,
				"max_retries", s.maxRetries,
			)

			select {
			case <-s.ctx.Done():
				return
			case <-time.After(s.retryDelay):
			}
		}

		// Execute task
		err := task.Handler(s.ctx)
		if err == nil {
			// Success
			runCount := s.updateTaskAfterRun(task, nil, startTime)

			s.logger.Info("Task completed successfully",
				"task_id", task.ID,
				"priority", task.Priority,
				"duration", time.Since(startTime),
				"run_count", runCount,
			)
			return
		}

		lastErr = err

		// Check if context was cancelled
		if s.ctx.Err() != nil {
			return
		}
	}

	// All retries failed
	s.updateTaskAfterRun(task, lastErr, startTime)

	s.logger.Error("Task failed after retries",
		"task_id", task.ID,
		"error", lastErr,
		"attempts", s.maxRetries+1,
		"duration", time.Since(startTime),
	)
}

// updateTaskAfterRun updates task statistics after execution.
func (s *Scheduler) updateTaskAfterRun(task *Task, err error, startTime time.Time) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.LastRun = startTime
	task.RunCount++

	if err != nil {
		task.ErrorCount++
		task.LastError = err
	} else {
		task.LastError = nil
	}

	// Calculate next run time
	if task.Schedule.Type == ScheduleTypeOnce {
		// One-time tasks get disabled after running
		task.Enabled = false
	} else {
		nextRun, calcErr := s.calculateNextRunForSchedule(task.Schedule, time.Now())
		if calcErr != nil {
			s.logger.Error("Failed to calculate next run",
				"task_id", task.ID,
				"error", calcErr,
			)
			task.NextRun = time.Now().Add(1 * time.Hour) // Default fallback
		} else {
			task.NextRun = nextRun
		}
	}

	return task.RunCount
}

// calculateNextRunForSchedule calculates when a task should run next.
func (s *Scheduler) calculateNextRunForSchedule(schedule Schedule, from time.Time) (time.Time, error) {
	switch schedule.Type {
	case ScheduleTypeInterval:
		if schedule.Interval <= 0 {
			return time.Time{}, errors.New("invalid interval: must be > 0")
		}
		return from.Add(schedule.Interval), nil

	case ScheduleTypeOnce:
		if len(schedule.Times) > 0 {
			return schedule.Times[0], nil
		}
		return from, nil

	case ScheduleTypeCron:
		if schedule.CronSpec == "" {
			return time.Time{}, errors.New("empty cron specification")
		}

		cronSchedule, err := ParseCron(schedule.CronSpec)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid cron spec: %w", err)
		}

		return cronSchedule.Next(from), nil

	default:
		return time.Time{}, fmt.Errorf("unknown schedule type: %s", schedule.Type)
	}
}

// GetStats returns scheduler statistics.
func (s *Scheduler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalTasks := len(s.tasks)
	enabledTasks := 0
	totalRuns := int64(0)
	totalErrors := int64(0)

	s.runningMu.RLock()
	runningCount := len(s.runningTasks)
	s.runningMu.RUnlock()

	for _, task := range s.tasks {
		if task.Enabled {
			enabledTasks++
		}
		totalRuns += task.RunCount
		totalErrors += task.ErrorCount
	}

	return map[string]interface{}{
		"total_tasks":   totalTasks,
		"enabled_tasks": enabledTasks,
		"running_tasks": runningCount,
		"total_runs":    totalRuns,
		"total_errors":  totalErrors,
	}
}

// SetMaxRetries sets the maximum number of retries for failed tasks.
func (s *Scheduler) SetMaxRetries(maxRetries int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.maxRetries = maxRetries
}

// SetRetryDelay sets the delay between retry attempts.
func (s *Scheduler) SetRetryDelay(delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.retryDelay = delay
}
