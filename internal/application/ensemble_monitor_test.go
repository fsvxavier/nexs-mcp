package application

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutionMonitor(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 3)

	assert.Equal(t, "exec-123", monitor.GetExecutionID())
	assert.Equal(t, "ensemble-456", monitor.ensembleID)
	assert.Equal(t, 3, monitor.totalAgents)
	assert.Equal(t, "initializing", monitor.GetStatus())
	assert.Equal(t, 0, monitor.completedAgents)
	assert.Equal(t, 0, monitor.failedAgents)
	assert.NotNil(t, monitor.agentProgress)
}

func TestExecutionMonitor_StartAgent(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 2)

	monitor.StartAgent("agent-1", "analyzer")

	progress, exists := monitor.GetAgentProgress("agent-1")
	require.True(t, exists)
	assert.Equal(t, "agent-1", progress.AgentID)
	assert.Equal(t, "analyzer", progress.Role)
	assert.Equal(t, "running", progress.Status)
	assert.Equal(t, 0.0, progress.Progress)
	assert.NotZero(t, progress.StartTime)
}

func TestExecutionMonitor_UpdateAgentProgress(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 1)
	monitor.StartAgent("agent-1", "processor")

	// Update progress
	metadata := map[string]interface{}{
		"step":  "processing",
		"count": 50,
	}
	monitor.UpdateAgentProgress("agent-1", 0.5, metadata)

	progress, exists := monitor.GetAgentProgress("agent-1")
	require.True(t, exists)
	assert.Equal(t, 0.5, progress.Progress)
	assert.Equal(t, "processing", progress.Metadata["step"])
	assert.Equal(t, 50, progress.Metadata["count"])
}

func TestExecutionMonitor_CompleteAgent(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 2)
	monitor.StartAgent("agent-1", "worker")

	monitor.CompleteAgent("agent-1")

	progress, exists := monitor.GetAgentProgress("agent-1")
	require.True(t, exists)
	assert.Equal(t, "completed", progress.Status)
	assert.Equal(t, 1.0, progress.Progress)
	assert.Equal(t, 1, monitor.completedAgents)
	assert.Equal(t, 0, monitor.failedAgents)
}

func TestExecutionMonitor_FailAgent(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 2)
	monitor.StartAgent("agent-1", "worker")

	monitor.FailAgent("agent-1", "connection timeout")

	progress, exists := monitor.GetAgentProgress("agent-1")
	require.True(t, exists)
	assert.Equal(t, "failed", progress.Status)
	assert.Equal(t, "connection timeout", progress.Error)
	assert.Equal(t, 0, monitor.completedAgents)
	assert.Equal(t, 1, monitor.failedAgents)
}

func TestExecutionMonitor_GetProgress(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 4)

	// Initially 0%
	assert.Equal(t, 0.0, monitor.GetProgress())

	// Start agents
	monitor.StartAgent("agent-1", "worker")
	monitor.StartAgent("agent-2", "worker")
	monitor.StartAgent("agent-3", "worker")
	monitor.StartAgent("agent-4", "worker")

	// Complete 2 agents = 50%
	monitor.CompleteAgent("agent-1")
	monitor.CompleteAgent("agent-2")
	assert.Equal(t, 0.5, monitor.GetProgress())

	// Fail 1 agent = 75% (3 done out of 4)
	monitor.FailAgent("agent-3", "error")
	assert.Equal(t, 0.75, monitor.GetProgress())

	// Complete last agent = 100%
	monitor.CompleteAgent("agent-4")
	assert.Equal(t, 1.0, monitor.GetProgress())
}

func TestExecutionMonitor_SetPhase(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 1)

	monitor.SetPhase("initialization")
	assert.Equal(t, "initialization", monitor.currentPhase)

	monitor.SetPhase("execution")
	assert.Equal(t, "execution", monitor.currentPhase)

	monitor.SetPhase("aggregation")
	assert.Equal(t, "aggregation", monitor.currentPhase)
}

func TestExecutionMonitor_SetStatus(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 1)

	monitor.SetStatus("running")
	assert.Equal(t, "running", monitor.GetStatus())

	monitor.SetStatus("completed")
	assert.Equal(t, "completed", monitor.GetStatus())

	monitor.SetStatus("failed")
	assert.Equal(t, "failed", monitor.GetStatus())
}

func TestExecutionMonitor_ProgressCallbacks(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 2)

	// Track callback invocations
	callbackCount := 0
	monitor.RegisterProgressCallback(func(m *ExecutionMonitor) {
		callbackCount++
	})

	// Start agent should trigger callback
	monitor.StartAgent("agent-1", "worker")
	time.Sleep(10 * time.Millisecond) // Give goroutine time to execute

	// Update progress should trigger callback
	monitor.UpdateAgentProgress("agent-1", 0.5, nil)
	time.Sleep(10 * time.Millisecond)

	// Complete agent should trigger callback
	monitor.CompleteAgent("agent-1")
	time.Sleep(10 * time.Millisecond)

	assert.Greater(t, callbackCount, 0, "Callbacks should have been invoked")
}

func TestExecutionMonitor_StateCallbacks(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 1)

	// Track state transitions
	var transitions []struct {
		old, new string
	}

	monitor.RegisterStateCallback(func(m *ExecutionMonitor, oldState, newState string) {
		transitions = append(transitions, struct{ old, new string }{oldState, newState})
	})

	monitor.SetStatus("running")
	monitor.SetStatus("completed")

	require.Len(t, transitions, 2)
	assert.Equal(t, "initializing", transitions[0].old)
	assert.Equal(t, "running", transitions[0].new)
	assert.Equal(t, "running", transitions[1].old)
	assert.Equal(t, "completed", transitions[1].new)
}

func TestExecutionMonitor_GetProgressUpdate(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 3)
	monitor.SetStatus("running")
	monitor.SetPhase("execution")

	monitor.StartAgent("agent-1", "analyzer")
	monitor.StartAgent("agent-2", "processor")
	monitor.StartAgent("agent-3", "validator")

	monitor.UpdateAgentProgress("agent-1", 0.5, map[string]interface{}{"step": "analysis"})
	monitor.CompleteAgent("agent-2")

	update := monitor.GetProgressUpdate()

	assert.Equal(t, "exec-123", update.ExecutionID)
	assert.Equal(t, "ensemble-456", update.EnsembleID)
	assert.Equal(t, "running", update.Status)
	assert.Equal(t, "execution", update.Phase)
	assert.Equal(t, 3, update.TotalAgents)
	assert.Equal(t, 1, update.CompletedAgents)
	assert.Equal(t, 0, update.FailedAgents)
	assert.InDelta(t, 0.33, update.Progress, 0.01)
	assert.Greater(t, update.ElapsedTime, time.Duration(0))
	assert.NotZero(t, update.Timestamp)

	// Check agent progress is included
	require.Len(t, update.AgentProgress, 3)
	assert.Equal(t, "running", update.AgentProgress["agent-1"].Status)
	assert.Equal(t, 0.5, update.AgentProgress["agent-1"].Progress)
	assert.Equal(t, "completed", update.AgentProgress["agent-2"].Status)
	assert.Equal(t, 1.0, update.AgentProgress["agent-2"].Progress)
}

func TestExecutionMonitor_EstimatedRemaining(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 4)

	// Set start time in the past
	monitor.startTime = time.Now().Add(-10 * time.Second)

	// Complete 2 out of 4 agents (50%)
	monitor.StartAgent("agent-1", "worker")
	monitor.StartAgent("agent-2", "worker")
	monitor.CompleteAgent("agent-1")
	monitor.CompleteAgent("agent-2")

	update := monitor.GetProgressUpdate()

	// If 50% took 10s, remaining 50% should take ~10s
	assert.InDelta(t, 10*time.Second, update.EstimatedRemaining, float64(2*time.Second))
}

func TestExecutionMonitor_MultipleCallbacks(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 1)

	callback1Count := 0
	callback2Count := 0

	monitor.RegisterProgressCallback(func(m *ExecutionMonitor) {
		callback1Count++
	})

	monitor.RegisterProgressCallback(func(m *ExecutionMonitor) {
		callback2Count++
	})

	monitor.StartAgent("agent-1", "worker")
	time.Sleep(10 * time.Millisecond)

	assert.Greater(t, callback1Count, 0)
	assert.Greater(t, callback2Count, 0)
}

func TestExecutionMonitor_ConcurrentAccess(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 10)

	// Simulate concurrent access from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		agentID := "agent-" + string(rune('0'+i))
		go func(id string) {
			monitor.StartAgent(id, "worker")
			monitor.UpdateAgentProgress(id, 0.5, nil)
			monitor.CompleteAgent(id)
			done <- true
		}(agentID)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// All agents should be completed
	assert.Equal(t, 1.0, monitor.GetProgress())
	assert.Equal(t, 10, monitor.completedAgents)
}

func TestExecutionMonitor_AgentProgressIsolation(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 2)

	monitor.StartAgent("agent-1", "worker")
	monitor.UpdateAgentProgress("agent-1", 0.3, map[string]interface{}{"data": "test1"})

	monitor.StartAgent("agent-2", "processor")
	monitor.UpdateAgentProgress("agent-2", 0.7, map[string]interface{}{"data": "test2"})

	// Each agent should maintain independent progress
	progress1, _ := monitor.GetAgentProgress("agent-1")
	progress2, _ := monitor.GetAgentProgress("agent-2")

	assert.Equal(t, 0.3, progress1.Progress)
	assert.Equal(t, "test1", progress1.Metadata["data"])
	assert.Equal(t, 0.7, progress2.Progress)
	assert.Equal(t, "test2", progress2.Metadata["data"])
}

func TestExecutionMonitor_ZeroAgents(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 0)

	// Progress should be 0 when there are no agents
	assert.Equal(t, 0.0, monitor.GetProgress())

	update := monitor.GetProgressUpdate()
	assert.Equal(t, 0, update.TotalAgents)
	assert.Equal(t, 0.0, update.Progress)
}

func TestExecutionMonitor_ProgressUpdateCopy(t *testing.T) {
	monitor := NewExecutionMonitor("exec-123", "ensemble-456", 1)
	monitor.StartAgent("agent-1", "worker")
	monitor.UpdateAgentProgress("agent-1", 0.5, map[string]interface{}{"key": "value1"})

	// Get first update
	update1 := monitor.GetProgressUpdate()

	// Modify agent progress
	monitor.UpdateAgentProgress("agent-1", 0.8, map[string]interface{}{"key": "value2"})

	// Get second update
	update2 := monitor.GetProgressUpdate()

	// First update should not be affected by later changes
	assert.Equal(t, 0.5, update1.AgentProgress["agent-1"].Progress)
	assert.Equal(t, "value1", update1.AgentProgress["agent-1"].Metadata["key"])
	assert.Equal(t, 0.8, update2.AgentProgress["agent-1"].Progress)
	assert.Equal(t, "value2", update2.AgentProgress["agent-1"].Metadata["key"])
}
