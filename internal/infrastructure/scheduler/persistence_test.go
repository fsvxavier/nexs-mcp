package scheduler

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskPersistence(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.json")
	persistence := NewTaskPersistence(tmpFile)

	require.NotNil(t, persistence)
	assert.Equal(t, tmpFile, persistence.filePath)
}

func TestTaskPersistence_SaveLoad(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "tasks.json")
	persistence := NewTaskPersistence(tmpFile)

	// Create test tasks
	now := time.Now().Round(time.Second)
	tasks := map[string]*Task{
		"task1": {
			ID:           "task1",
			Name:         "Test Task 1",
			Description:  "First test task",
			Enabled:      true,
			Priority:     10,
			Dependencies: []string{"dep1", "dep2"},
			LastRun:      now.Add(-1 * time.Hour),
			NextRun:      now.Add(1 * time.Hour),
			RunCount:     5,
			ErrorCount:   1,
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 15 * time.Minute,
			},
		},
		"task2": {
			ID:          "task2",
			Name:        "Test Task 2",
			Description: "Second test task",
			Enabled:     false,
			Priority:    20,
			RunCount:    10,
			ErrorCount:  0,
			Schedule: Schedule{
				Type:     ScheduleTypeCron,
				CronSpec: "0 */5 * * *",
			},
		},
	}

	// Save tasks
	err := persistence.Save(tasks)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(tmpFile)
	require.NoError(t, err)

	// Load tasks
	loaded, err := persistence.Load()
	require.NoError(t, err)
	assert.Len(t, loaded, 2)

	// Verify task1
	var task1 *SerializableTask
	for i := range loaded {
		if loaded[i].ID == "task1" {
			task1 = &loaded[i]
			break
		}
	}
	require.NotNil(t, task1)
	assert.Equal(t, "Test Task 1", task1.Name)
	assert.Equal(t, "First test task", task1.Description)
	assert.True(t, task1.Enabled)
	assert.Equal(t, 10, task1.Priority)
	assert.Equal(t, []string{"dep1", "dep2"}, task1.Dependencies)
	assert.Equal(t, int64(5), task1.RunCount)
	assert.Equal(t, int64(1), task1.ErrorCount)
	assert.Equal(t, "interval", task1.Schedule.Type)
	assert.Equal(t, "15m0s", task1.Schedule.Interval)

	// Verify task2
	var task2 *SerializableTask
	for i := range loaded {
		if loaded[i].ID == "task2" {
			task2 = &loaded[i]
			break
		}
	}
	require.NotNil(t, task2)
	assert.Equal(t, "Test Task 2", task2.Name)
	assert.False(t, task2.Enabled)
	assert.Equal(t, "cron", task2.Schedule.Type)
	assert.Equal(t, "0 */5 * * *", task2.Schedule.CronSpec)
}

func TestTaskPersistence_LoadNonExistent(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "nonexistent.json")
	persistence := NewTaskPersistence(tmpFile)

	// Load from non-existent file should return empty slice
	loaded, err := persistence.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestTaskPersistence_SaveEmptyTasks(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.json")
	persistence := NewTaskPersistence(tmpFile)

	// Save empty tasks map
	err := persistence.Save(map[string]*Task{})
	require.NoError(t, err)

	// Load should return empty slice
	loaded, err := persistence.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestTaskPersistence_WithDependencies(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "dependencies.json")
	persistence := NewTaskPersistence(tmpFile)

	tasks := map[string]*Task{
		"main": {
			ID:           "main",
			Name:         "Main Task",
			Enabled:      true,
			Dependencies: []string{"dep1", "dep2", "dep3"},
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 1 * time.Hour,
			},
		},
	}

	err := persistence.Save(tasks)
	require.NoError(t, err)

	loaded, err := persistence.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 1)

	assert.Equal(t, []string{"dep1", "dep2", "dep3"}, loaded[0].Dependencies)
}

func TestTaskPersistence_CronSchedule(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "cron.json")
	persistence := NewTaskPersistence(tmpFile)

	tasks := map[string]*Task{
		"cron-task": {
			ID:      "cron-task",
			Name:    "Cron Task",
			Enabled: true,
			Schedule: Schedule{
				Type:     ScheduleTypeCron,
				CronSpec: "0 0 * * *",
			},
		},
	}

	err := persistence.Save(tasks)
	require.NoError(t, err)

	loaded, err := persistence.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 1)

	assert.Equal(t, "cron", loaded[0].Schedule.Type)
	assert.Equal(t, "0 0 * * *", loaded[0].Schedule.CronSpec)
	assert.Empty(t, loaded[0].Schedule.Interval)
}

func TestTaskPersistence_OnceSchedule(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "once.json")
	persistence := NewTaskPersistence(tmpFile)

	runTime := time.Now().Add(24 * time.Hour).Round(time.Second)

	tasks := map[string]*Task{
		"once-task": {
			ID:      "once-task",
			Name:    "Once Task",
			Enabled: true,
			Schedule: Schedule{
				Type:  ScheduleTypeOnce,
				Times: []time.Time{runTime},
			},
		},
	}

	err := persistence.Save(tasks)
	require.NoError(t, err)

	loaded, err := persistence.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 1)

	assert.Equal(t, "once", loaded[0].Schedule.Type)
	assert.Len(t, loaded[0].Schedule.Times, 1)
	assert.WithinDuration(t, runTime, loaded[0].Schedule.Times[0], 1*time.Second)
}

func TestTaskPersistence_MultipleSaves(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "multiple.json")
	persistence := NewTaskPersistence(tmpFile)

	// First save
	tasks1 := map[string]*Task{
		"task1": {
			ID:      "task1",
			Name:    "Task 1",
			Enabled: true,
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 1 * time.Minute,
			},
		},
	}
	err := persistence.Save(tasks1)
	require.NoError(t, err)

	// Second save (should overwrite)
	tasks2 := map[string]*Task{
		"task2": {
			ID:      "task2",
			Name:    "Task 2",
			Enabled: false,
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 5 * time.Minute,
			},
		},
	}
	err = persistence.Save(tasks2)
	require.NoError(t, err)

	// Load should return only task2
	loaded, err := persistence.Load()
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, "task2", loaded[0].ID)
}

func TestTaskPersistence_InvalidDirectory(t *testing.T) {
	// Try to save to an invalid path (file exists where directory should be)
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	_, err := os.Create(tmpFile)
	require.NoError(t, err)

	// Try to create tasks file inside the file (should fail)
	invalidPath := filepath.Join(tmpFile, "tasks.json")
	persistence := NewTaskPersistence(invalidPath)

	tasks := map[string]*Task{
		"task": {
			ID:   "task",
			Name: "Test",
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 1 * time.Minute,
			},
		},
	}

	err = persistence.Save(tasks)
	assert.Error(t, err)
}

func TestTaskPersistence_RunCounts(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "counts.json")
	persistence := NewTaskPersistence(tmpFile)

	tasks := map[string]*Task{
		"task": {
			ID:         "task",
			Name:       "Task with counts",
			RunCount:   100,
			ErrorCount: 5,
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 1 * time.Minute,
			},
		},
	}

	err := persistence.Save(tasks)
	require.NoError(t, err)

	loaded, err := persistence.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 1)

	assert.Equal(t, int64(100), loaded[0].RunCount)
	assert.Equal(t, int64(5), loaded[0].ErrorCount)
}

func TestTaskPersistence_TimePreservation(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "times.json")
	persistence := NewTaskPersistence(tmpFile)

	now := time.Now().Round(time.Second)
	lastRun := now.Add(-1 * time.Hour)
	nextRun := now.Add(1 * time.Hour)

	tasks := map[string]*Task{
		"task": {
			ID:      "task",
			Name:    "Task with times",
			LastRun: lastRun,
			NextRun: nextRun,
			Schedule: Schedule{
				Type:     ScheduleTypeInterval,
				Interval: 1 * time.Hour,
			},
		},
	}

	err := persistence.Save(tasks)
	require.NoError(t, err)

	loaded, err := persistence.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 1)

	// Times should be preserved within a second
	assert.WithinDuration(t, lastRun, loaded[0].LastRun, 1*time.Second)
	assert.WithinDuration(t, nextRun, loaded[0].NextRun, 1*time.Second)
}

func TestSerializableTask_Structure(t *testing.T) {
	// Verify that SerializableTask has all expected fields
	task := SerializableTask{
		ID:           "test-id",
		Name:         "test-name",
		Description:  "test-description",
		Enabled:      true,
		Priority:     50,
		Dependencies: []string{"dep1"},
		LastRun:      time.Now(),
		NextRun:      time.Now(),
		RunCount:     10,
		ErrorCount:   2,
		Schedule: SerializableSchedule{
			Type:     "interval",
			Interval: "5m",
			CronSpec: "0 * * * *",
			Times:    []time.Time{time.Now()},
		},
	}

	// Just verify the structure compiles and fields are accessible
	assert.Equal(t, "test-id", task.ID)
	assert.Equal(t, "test-name", task.Name)
	assert.True(t, task.Enabled)
	assert.Equal(t, 50, task.Priority)
	assert.Equal(t, int64(10), task.RunCount)
}
