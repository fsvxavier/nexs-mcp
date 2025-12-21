package application

import (
	"sync"
	"time"
)

// ExecutionMonitor tracks ensemble execution progress in real-time.
type ExecutionMonitor struct {
	mu                sync.RWMutex
	executionID       string
	ensembleID        string
	totalAgents       int
	completedAgents   int
	failedAgents      int
	startTime         time.Time
	status            string
	currentPhase      string
	agentProgress     map[string]*AgentProgress
	progressCallbacks []ProgressCallback
	stateCallbacks    []StateCallback
}

// AgentProgress tracks individual agent progress.
type AgentProgress struct {
	AgentID    string
	Role       string
	Status     string // queued, running, completed, failed
	Progress   float64
	StartTime  time.Time
	LastUpdate time.Time
	Error      string
	Metadata   map[string]interface{}
}

// ProgressCallback is called when progress is updated.
type ProgressCallback func(monitor *ExecutionMonitor)

// StateCallback is called when execution state changes.
type StateCallback func(monitor *ExecutionMonitor, oldState, newState string)

// ProgressUpdate represents a progress update event.
type ProgressUpdate struct {
	ExecutionID        string                    `json:"execution_id"`
	EnsembleID         string                    `json:"ensemble_id"`
	Status             string                    `json:"status"`
	Phase              string                    `json:"phase"`
	TotalAgents        int                       `json:"total_agents"`
	CompletedAgents    int                       `json:"completed_agents"`
	FailedAgents       int                       `json:"failed_agents"`
	Progress           float64                   `json:"progress"` // 0.0 to 1.0
	ElapsedTime        time.Duration             `json:"elapsed_time"`
	EstimatedRemaining time.Duration             `json:"estimated_remaining,omitempty"`
	Timestamp          time.Time                 `json:"timestamp"`
	AgentProgress      map[string]*AgentProgress `json:"agent_progress,omitempty"`
}

// NewExecutionMonitor creates a new execution monitor.
func NewExecutionMonitor(executionID, ensembleID string, totalAgents int) *ExecutionMonitor {
	return &ExecutionMonitor{
		executionID:   executionID,
		ensembleID:    ensembleID,
		totalAgents:   totalAgents,
		startTime:     time.Now(),
		status:        "initializing",
		agentProgress: make(map[string]*AgentProgress),
	}
}

// RegisterProgressCallback adds a callback for progress updates.
func (m *ExecutionMonitor) RegisterProgressCallback(callback ProgressCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.progressCallbacks = append(m.progressCallbacks, callback)
}

// RegisterStateCallback adds a callback for state changes.
func (m *ExecutionMonitor) RegisterStateCallback(callback StateCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stateCallbacks = append(m.stateCallbacks, callback)
}

// StartAgent marks an agent as started.
func (m *ExecutionMonitor) StartAgent(agentID, role string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.agentProgress[agentID] = &AgentProgress{
		AgentID:    agentID,
		Role:       role,
		Status:     "running",
		Progress:   0.0,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	m.notifyProgress()
}

// UpdateAgentProgress updates an agent's progress.
func (m *ExecutionMonitor) UpdateAgentProgress(agentID string, progress float64, metadata map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ap, exists := m.agentProgress[agentID]; exists {
		ap.Progress = progress
		ap.LastUpdate = time.Now()
		if metadata != nil {
			for k, v := range metadata {
				ap.Metadata[k] = v
			}
		}
		m.notifyProgress()
	}
}

// CompleteAgent marks an agent as completed successfully.
func (m *ExecutionMonitor) CompleteAgent(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ap, exists := m.agentProgress[agentID]; exists {
		ap.Status = "completed"
		ap.Progress = 1.0
		ap.LastUpdate = time.Now()
		m.completedAgents++
		m.notifyProgress()
	}
}

// FailAgent marks an agent as failed.
func (m *ExecutionMonitor) FailAgent(agentID, errorMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ap, exists := m.agentProgress[agentID]; exists {
		ap.Status = "failed"
		ap.Error = errorMsg
		ap.LastUpdate = time.Now()
		m.failedAgents++
		m.notifyProgress()
	}
}

// SetPhase updates the current execution phase.
func (m *ExecutionMonitor) SetPhase(phase string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentPhase = phase
	m.notifyProgress()
}

// SetStatus updates the execution status.
func (m *ExecutionMonitor) SetStatus(newStatus string) {
	m.mu.Lock()
	oldStatus := m.status
	m.status = newStatus
	m.mu.Unlock()

	// Notify state callbacks
	for _, callback := range m.stateCallbacks {
		callback(m, oldStatus, newStatus)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifyProgress()
}

// GetProgress returns the current overall progress (0.0 to 1.0).
func (m *ExecutionMonitor) GetProgress() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.totalAgents == 0 {
		return 0.0
	}

	return float64(m.completedAgents+m.failedAgents) / float64(m.totalAgents)
}

// GetProgressUpdate returns a snapshot of current progress.
func (m *ExecutionMonitor) GetProgressUpdate() *ProgressUpdate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	elapsed := time.Since(m.startTime)
	progress := m.GetProgress()

	var estimated time.Duration
	if progress > 0 && progress < 1.0 {
		estimated = time.Duration(float64(elapsed) / progress * (1.0 - progress))
	}

	// Copy agent progress
	agentProgressCopy := make(map[string]*AgentProgress)
	for k, v := range m.agentProgress {
		metaCopy := make(map[string]interface{})
		for mk, mv := range v.Metadata {
			metaCopy[mk] = mv
		}
		agentProgressCopy[k] = &AgentProgress{
			AgentID:    v.AgentID,
			Role:       v.Role,
			Status:     v.Status,
			Progress:   v.Progress,
			StartTime:  v.StartTime,
			LastUpdate: v.LastUpdate,
			Error:      v.Error,
			Metadata:   metaCopy,
		}
	}

	return &ProgressUpdate{
		ExecutionID:        m.executionID,
		EnsembleID:         m.ensembleID,
		Status:             m.status,
		Phase:              m.currentPhase,
		TotalAgents:        m.totalAgents,
		CompletedAgents:    m.completedAgents,
		FailedAgents:       m.failedAgents,
		Progress:           progress,
		ElapsedTime:        elapsed,
		EstimatedRemaining: estimated,
		Timestamp:          time.Now(),
		AgentProgress:      agentProgressCopy,
	}
}

// notifyProgress calls all registered progress callbacks (must be called with lock held).
func (m *ExecutionMonitor) notifyProgress() {
	for _, callback := range m.progressCallbacks {
		go callback(m) // Run in goroutine to avoid blocking
	}
}

// GetExecutionID returns the execution ID.
func (m *ExecutionMonitor) GetExecutionID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.executionID
}

// GetStatus returns the current status.
func (m *ExecutionMonitor) GetStatus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// GetAgentProgress returns progress for a specific agent.
func (m *ExecutionMonitor) GetAgentProgress(agentID string) (*AgentProgress, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ap, exists := m.agentProgress[agentID]
	return ap, exists
}
