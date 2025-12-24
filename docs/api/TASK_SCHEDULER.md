# Background Task Scheduler

## Overview

The task scheduler provides a robust system for managing background tasks with support for:

- **Interval-based scheduling**: Run tasks at fixed intervals
- **One-time scheduling**: Run tasks once at specific times
- **Retry logic**: Automatic retry with configurable delays
- **Task management**: Enable/disable tasks dynamically
- **Statistics and monitoring**: Track task executions and failures
- **Graceful shutdown**: Wait for running tasks before stopping

## Architecture

### Components

1. **Scheduler**: Core scheduler managing task lifecycle
2. **Task**: Individual schedulable unit with handler and schedule
3. **Schedule**: Defines when a task should run (interval, once, cron)
4. **TaskHandler**: Function type for task execution

### Features

- Thread-safe operations with RWMutex
- Automatic retry with exponential backoff
- Task isolation - failures don't affect other tasks
- Race-condition free (tested with -race)
- Concurrent task execution

## Usage

### Creating a Scheduler

```go
import (
    "log/slog"
    "github.com/fsvxavier/nexs-mcp/internal/infrastructure/scheduler"
)

logger := slog.Default()
sched := scheduler.NewScheduler(logger)
```

### Adding Tasks

#### Interval-Based Task

```go
task := &scheduler.Task{
    ID:   "cleanup-task",
    Name: "Periodic Cleanup",
    Handler: func(ctx context.Context) error {
        // Task logic here
        return nil
    },
    Schedule: scheduler.Schedule{
        Type:     scheduler.ScheduleTypeInterval,
        Interval: 5 * time.Minute,
    },
    Enabled: true,
}

err := sched.AddTask(task)
```

#### One-Time Task

```go
task := &scheduler.Task{
    ID:   "one-time-task",
    Name: "Run Once",
    Handler: func(ctx context.Context) error {
        // Task logic here
        return nil
    },
    Schedule: scheduler.Schedule{
        Type:  scheduler.ScheduleTypeOnce,
        Times: []time.Time{time.Now().Add(1 * time.Hour)},
    },
    Enabled: true,
}

err := sched.AddTask(task)
```

### Starting and Stopping

```go
// Start the scheduler
sched.Start()

// Stop gracefully (waits for running tasks)
defer sched.Stop()
```

### Task Management

```go
// Disable a task
err := sched.DisableTask("cleanup-task")

// Enable a task
err := sched.EnableTask("cleanup-task")

// Remove a task
err := sched.RemoveTask("cleanup-task")

// Get task info
task, err := sched.GetTask("cleanup-task")

// List all tasks
tasks := sched.ListTasks()
```

### Configuration

```go
// Set maximum retries (default: 3)
sched.SetMaxRetries(5)

// Set retry delay (default: 5s)
sched.SetRetryDelay(10 * time.Second)
```

### Monitoring

```go
// Get statistics
stats := sched.GetStats()
fmt.Printf("Total tasks: %d\n", stats["total_tasks"])
fmt.Printf("Enabled tasks: %d\n", stats["enabled_tasks"])
fmt.Printf("Total runs: %d\n", stats["total_runs"])
fmt.Printf("Total errors: %d\n", stats["total_errors"])

// Check individual task stats
task, _ := sched.GetTask("cleanup-task")
fmt.Printf("Run count: %d\n", task.RunCount)
fmt.Printf("Error count: %d\n", task.ErrorCount)
fmt.Printf("Last error: %v\n", task.LastError)
fmt.Printf("Last run: %s\n", task.LastRun)
fmt.Printf("Next run: %s\n", task.NextRun)
```

## Implementation Details

### Scheduling Algorithm

- **Ticker-based checking**: Every 100ms the scheduler checks for tasks to run
- **Immediate first run**: Tasks run immediately upon registration (if enabled)
- **Collision prevention**: Tasks mark next run as far future while executing
- **Interval calculation**: Next run calculated after task completion

### Retry Mechanism

```
Initial attempt
  ↓ (fails)
Wait retry_delay
  ↓
Retry 1
  ↓ (fails)
Wait retry_delay
  ↓
Retry 2
  ↓ (fails)
Wait retry_delay
  ↓
Retry 3
  ↓
All retries exhausted - mark as failed
```

### Thread Safety

All operations are protected by RWMutex:
- Read operations: `GetTask`, `ListTasks`, `GetStats`
- Write operations: `AddTask`, `RemoveTask`, `EnableTask`, `DisableTask`
- Execution: Tasks execute outside locks to prevent blocking

### Graceful Shutdown

```go
scheduler.Stop() // Calls:
  1. cancel() - Signal context cancellation
  2. wg.Wait() - Wait for all running tasks to complete
```

## Testing

The scheduler includes comprehensive tests:

- **Basic operations**: Add, remove, enable, disable tasks
- **Scheduling**: Interval and one-time execution
- **Retry logic**: Success after retries and permanent failures
- **Concurrency**: Race detection with `-race` flag
- **Graceful shutdown**: Task completion before stop

Run tests:

```bash
go test -race -v ./internal/infrastructure/scheduler/
```

## Examples

### Working Memory Cleanup

```go
cleanupTask := &scheduler.Task{
    ID:   "wm-cleanup",
    Name: "Working Memory Cleanup",
    Description: "Remove expired memories and inactive sessions",
    Handler: func(ctx context.Context) error {
        return workingMemoryService.Cleanup(ctx)
    },
    Schedule: scheduler.Schedule{
        Type:     scheduler.ScheduleTypeInterval,
        Interval: 5 * time.Minute,
    },
    Enabled: true,
}
```

### Confidence Decay Recalculation

```go
decayTask := &scheduler.Task{
    ID:   "confidence-decay",
    Name: "Confidence Decay Calculation",
    Description: "Recalculate confidence values based on time decay",
    Handler: func(ctx context.Context) error {
        return temporalService.RecalculateDecay(ctx)
    },
    Schedule: scheduler.Schedule{
        Type:     scheduler.ScheduleTypeInterval,
        Interval: 1 * time.Hour,
    },
    Enabled: true,
}
```

### Backup Task

```go
backupTask := &scheduler.Task{
    ID:   "daily-backup",
    Name: "Daily Backup",
    Description: "Create daily backup of data",
    Handler: func(ctx context.Context) error {
        return backupService.CreateBackup(ctx)
    },
    Schedule: scheduler.Schedule{
        Type:  scheduler.ScheduleTypeOnce,
        Times: []time.Time{time.Now().Add(24 * time.Hour)},
    },
    Enabled: true,
}
```

## Advanced Features

### Cron-like Scheduling ✅

Support for cron expressions in "minute hour day month weekday" format:

```go
task := &scheduler.Task{
    ID:   "daily-task",
    Name: "Daily Midnight Job",
    Handler: func(ctx context.Context) error {
        return doWork()
    },
    Schedule: scheduler.Schedule{
        Type: scheduler.ScheduleTypeCron,
        Cron: "0 0 * * *", // Daily at midnight
    },
    Enabled: true,
}
```

**Supported Cron Syntax:**
- Wildcards: `* * * * *` (every minute)
- Ranges: `0 9-17 * * *` (every hour from 9am to 5pm)
- Steps: `*/5 * * * *` (every 5 minutes)
- Lists: `0 8,12,18 * * *` (at 8am, noon, and 6pm)
- Combinations: `0 9-17 * * 1-5` (business hours, Mon-Fri)

**Examples:**
- `0 0 * * *` - Daily at midnight
- `*/15 * * * *` - Every 15 minutes
- `0 9-17 * * 1-5` - Every hour during business hours (Mon-Fri)
- `0 0 1 * *` - First day of every month at midnight
- `30 2 * * 0` - Sundays at 2:30am

### Priority-based Execution ✅

Tasks with higher priority execute first when multiple tasks are ready:

```go
import "github.com/fsvxavier/nexs-mcp/internal/infrastructure/scheduler"

highPriorityTask := &scheduler.Task{
    ID:       "critical-task",
    Name:     "Critical Operation",
    Priority: scheduler.PriorityHigh, // 100
    Handler:  func(ctx context.Context) error { return doCritical() },
    Schedule: scheduler.Schedule{
        Type:     scheduler.ScheduleTypeInterval,
        Interval: 1 * time.Minute,
    },
    Enabled: true,
}

lowPriorityTask := &scheduler.Task{
    ID:       "background-task",
    Name:     "Background Work",
    Priority: scheduler.PriorityLow, // 0
    Handler:  func(ctx context.Context) error { return doBackground() },
    Schedule: scheduler.Schedule{
        Type:     scheduler.ScheduleTypeInterval,
        Interval: 1 * time.Minute,
    },
    Enabled: true,
}
```

**Priority Levels:**
- `PriorityLow`: 0 (background tasks)
- `PriorityMedium`: 50 (default, regular tasks)
- `PriorityHigh`: 100 (critical operations)
- Custom values: Any integer (higher = higher priority)

### Task Dependencies ✅

Define task execution order by specifying dependencies:

```go
// Task A: Prepare data
taskA := &scheduler.Task{
    ID:      "prepare-data",
    Name:    "Data Preparation",
    Handler: func(ctx context.Context) error { return prepareData() },
    Schedule: scheduler.Schedule{
        Type:     scheduler.ScheduleTypeInterval,
        Interval: 10 * time.Minute,
    },
    Enabled: true,
}

// Task B: Process data (depends on A)
taskB := &scheduler.Task{
    ID:           "process-data",
    Name:         "Data Processing",
    Dependencies: []string{"prepare-data"}, // Wait for taskA
    Handler:      func(ctx context.Context) error { return processData() },
    Schedule: scheduler.Schedule{
        Type:     scheduler.ScheduleTypeInterval,
        Interval: 10 * time.Minute,
    },
    Enabled: true,
}

sched.AddTask(taskA)
sched.AddTask(taskB) // Will wait for taskA to complete
```

**Dependency Rules:**
- Tasks only run after ALL dependencies complete successfully
- Circular dependencies are detected and rejected
- Dependencies are validated when adding tasks
- If a dependency fails, dependent tasks are skipped for that cycle

### Persistent Task Storage ✅

Tasks can be saved to JSON and restored on restart:

```go
import "github.com/fsvxavier/nexs-mcp/internal/infrastructure/scheduler"

// Setup persistence
persistence := scheduler.NewTaskPersistence("/var/lib/nexs-mcp/tasks.json")
sched.SetPersistence(persistence)

// Register handler functions (handlers can't be serialized)
sched.RegisterHandler("cleanup-task", func(ctx context.Context) error {
    return cleanupService.Cleanup(ctx)
})

sched.RegisterHandler("backup-task", func(ctx context.Context) error {
    return backupService.Backup(ctx)
})

// Load tasks from previous session
if err := sched.LoadTasks(); err != nil {
    logger.Warn("Failed to load tasks", "error", err)
}

// Start scheduler
sched.Start()

// Tasks are automatically saved on:
// - AddTask()
// - RemoveTask()
// - EnableTask()
// - DisableTask()

// Manual save
if err := sched.SaveTasks(); err != nil {
    logger.Error("Failed to save tasks", "error", err)
}
```

**Persistence Features:**
- JSON serialization with atomic writes (temp file + rename)
- Separate handler registration (handlers can't be serialized)
- Automatic save on task modifications
- Restore task state: schedule, priority, dependencies, enabled status
- Thread-safe save/load operations

## Future Enhancements

- [ ] Web UI for task management
- [ ] Metrics export (Prometheus)
- [ ] Task execution history/audit log
- [ ] Distributed task scheduling (multiple instances)
- [ ] Task result persistence
- [ ] Advanced cron features (last day of month, nth weekday)
- [ ] Task hooks (before/after execution callbacks)

## Performance Characteristics

- **Overhead**: ~100ms scheduling precision (ticker interval)
- **Concurrency**: Unlimited concurrent tasks (one goroutine per task)
- **Memory**: O(n) where n = number of registered tasks
- **CPU**: Minimal - only ticker and mutex operations

## Best Practices

1. **Task handlers should be idempotent**: Tasks may run multiple times if retried
2. **Keep handlers short**: Long-running tasks block goroutines
3. **Use context for cancellation**: Respect context cancellation in handlers
4. **Log appropriately**: Task execution is logged automatically
5. **Handle errors gracefully**: Return errors for retry, nil for success
6. **Avoid blocking operations**: Use timeouts and cancellation

## Related Documentation

- [Working Memory](../user-guide/WORKING_MEMORY.md)
- [Temporal Features](./TEMPORAL_FEATURES.md)
- [Background Tasks in Working Memory](../architecture/APPLICATION.md#background-tasks)
