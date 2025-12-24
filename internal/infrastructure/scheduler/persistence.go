package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TaskPersistence handles saving and loading tasks from disk.
type TaskPersistence struct {
	filePath string
	mu       sync.RWMutex
}

// NewTaskPersistence creates a new task persistence manager.
func NewTaskPersistence(filePath string) *TaskPersistence {
	return &TaskPersistence{
		filePath: filePath,
	}
}

// SerializableTask represents a task that can be serialized to JSON.
type SerializableTask struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Schedule     SerializableSchedule `json:"schedule"`
	Enabled      bool                 `json:"enabled"`
	Priority     int                  `json:"priority"`
	Dependencies []string             `json:"dependencies"`
	LastRun      time.Time            `json:"last_run,omitempty"`
	NextRun      time.Time            `json:"next_run,omitempty"`
	RunCount     int64                `json:"run_count"`
	ErrorCount   int64                `json:"error_count"`
}

// SerializableSchedule represents a schedule that can be serialized.
type SerializableSchedule struct {
	Type     string      `json:"type"`
	Interval string      `json:"interval,omitempty"` // Duration as string
	CronSpec string      `json:"cron_spec,omitempty"`
	Times    []time.Time `json:"times,omitempty"`
}

// Save saves tasks to disk.
func (tp *TaskPersistence) Save(tasks map[string]*Task) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Convert tasks to serializable format
	serializable := make([]SerializableTask, 0, len(tasks))
	for _, task := range tasks {
		st := SerializableTask{
			ID:           task.ID,
			Name:         task.Name,
			Description:  task.Description,
			Enabled:      task.Enabled,
			Priority:     task.Priority,
			Dependencies: task.Dependencies,
			LastRun:      task.LastRun,
			NextRun:      task.NextRun,
			RunCount:     task.RunCount,
			ErrorCount:   task.ErrorCount,
			Schedule: SerializableSchedule{
				Type:     string(task.Schedule.Type),
				CronSpec: task.Schedule.CronSpec,
				Times:    task.Schedule.Times,
			},
		}

		if task.Schedule.Interval > 0 {
			st.Schedule.Interval = task.Schedule.Interval.String()
		}

		serializable = append(serializable, st)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(serializable, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(tp.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to temporary file first
	tempFile := tp.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename to final file (atomic operation)
	if err := os.Rename(tempFile, tp.filePath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// Load loads tasks from disk.
func (tp *TaskPersistence) Load() ([]SerializableTask, error) {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(tp.filePath); os.IsNotExist(err) {
		return []SerializableTask{}, nil
	}

	// Read file
	data, err := os.ReadFile(tp.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal JSON
	var tasks []SerializableTask
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	return tasks, nil
}

// Delete removes the persistence file.
func (tp *TaskPersistence) Delete() error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if err := os.Remove(tp.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ConvertToTask converts a SerializableTask back to a Task
// Note: TaskHandler must be registered separately as functions cannot be serialized.
func ConvertToTask(st SerializableTask) (*Task, error) {
	task := &Task{
		ID:           st.ID,
		Name:         st.Name,
		Description:  st.Description,
		Enabled:      st.Enabled,
		Priority:     st.Priority,
		Dependencies: st.Dependencies,
		LastRun:      st.LastRun,
		NextRun:      st.NextRun,
		RunCount:     st.RunCount,
		ErrorCount:   st.ErrorCount,
		Schedule: Schedule{
			Type:     ScheduleType(st.Schedule.Type),
			CronSpec: st.Schedule.CronSpec,
			Times:    st.Schedule.Times,
		},
	}

	// Parse interval if present
	if st.Schedule.Interval != "" {
		interval, err := time.ParseDuration(st.Schedule.Interval)
		if err != nil {
			return nil, fmt.Errorf("failed to parse interval: %w", err)
		}
		task.Schedule.Interval = interval
	}

	return task, nil
}
