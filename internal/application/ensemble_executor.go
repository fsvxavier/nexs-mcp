package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/common"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

// EnsembleExecutor orchestrates multi-agent execution.
type EnsembleExecutor struct {
	repository domain.ElementRepository
	logger     *slog.Logger
}

// NewEnsembleExecutor creates a new ensemble executor.
func NewEnsembleExecutor(repository domain.ElementRepository, log *slog.Logger) *EnsembleExecutor {
	if log == nil {
		log = logger.Get()
	}
	return &EnsembleExecutor{
		repository: repository,
		logger:     log,
	}
}

// ExecutionRequest represents an ensemble execution request.
type ExecutionRequest struct {
	EnsembleID string                 `json:"ensemble_id"`
	Input      map[string]interface{} `json:"input"`
	Options    ExecutionOptions       `json:"options,omitempty"`
}

// ExecutionOptions configures execution behavior.
type ExecutionOptions struct {
	Timeout          time.Duration `json:"timeout,omitempty"`           // Max execution time
	MaxRetries       int           `json:"max_retries,omitempty"`       // Max retries per agent
	FailFast         bool          `json:"fail_fast,omitempty"`         // Stop on first error
	CollectAll       bool          `json:"collect_all,omitempty"`       // Collect all results even after first success
	EnableMonitoring bool          `json:"enable_monitoring,omitempty"` // Enable execution monitoring
}

// ExecutionResult represents the result of ensemble execution.
type ExecutionResult struct {
	EnsembleID       string                 `json:"ensemble_id"`
	Status           string                 `json:"status"` // success, partial_success, failed
	Results          []AgentResult          `json:"results"`
	AggregatedResult interface{}            `json:"aggregated_result,omitempty"`
	ExecutionTime    time.Duration          `json:"execution_time"`
	StartedAt        time.Time              `json:"started_at"`
	FinishedAt       time.Time              `json:"finished_at"`
	Errors           []string               `json:"errors,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// AgentResult represents the result from a single agent.
type AgentResult struct {
	AgentID       string                 `json:"agent_id"`
	Role          string                 `json:"role"`
	Status        string                 `json:"status"` // success, failed, skipped
	Result        interface{}            `json:"result,omitempty"`
	Error         string                 `json:"error,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time"`
	StartedAt     time.Time              `json:"started_at"`
	FinishedAt    time.Time              `json:"finished_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Execute runs the ensemble with the specified execution mode.
func (e *EnsembleExecutor) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
	startTime := time.Now()

	// Load ensemble
	ensemble, err := e.loadEnsemble(req.EnsembleID)
	if err != nil {
		return nil, fmt.Errorf("failed to load ensemble: %w", err)
	}

	// Validate ensemble
	if len(ensemble.Members) == 0 {
		return nil, errors.New("ensemble has no members")
	}

	// Create execution monitor if enabled
	var monitor *ExecutionMonitor
	if req.Options.EnableMonitoring {
		monitor = NewExecutionMonitor(
			fmt.Sprintf("exec-%s-%d", req.EnsembleID, time.Now().Unix()),
			req.EnsembleID,
			len(ensemble.Members),
		)
		monitor.SetStatus("running")
		monitor.SetPhase("initialization")
	}

	// Setup context with timeout
	execCtx := ctx
	if req.Options.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, req.Options.Timeout)
		defer cancel()
	}

	// Execute based on mode
	var results []AgentResult
	var execErr error

	if monitor != nil {
		monitor.SetPhase("execution-" + ensemble.ExecutionMode)
	}

	switch ensemble.ExecutionMode {
	case "sequential":
		results, execErr = e.executeSequential(execCtx, ensemble, req, monitor)
	case "parallel":
		results, execErr = e.executeParallel(execCtx, ensemble, req, monitor)
	case "hybrid":
		results, execErr = e.executeHybrid(execCtx, ensemble, req, monitor)
	default:
		execErr = fmt.Errorf("unsupported execution mode: %s", ensemble.ExecutionMode)
		results = []AgentResult{}
	}

	if monitor != nil {
		monitor.SetPhase("aggregation")
	}

	// Calculate execution time
	executionTime := time.Since(startTime)

	// Build execution result
	result := &ExecutionResult{
		EnsembleID:    req.EnsembleID,
		Results:       results,
		ExecutionTime: executionTime,
		StartedAt:     startTime,
		FinishedAt:    time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// Add monitoring data to metadata
	if monitor != nil {
		progressUpdate := monitor.GetProgressUpdate()
		result.Metadata["monitoring"] = progressUpdate
		monitor.SetPhase("completed")
	}

	// Aggregate results
	if execErr == nil {
		aggregated, err := e.aggregateResults(ensemble, results)
		if err != nil {
			e.logger.Error("Failed to aggregate results", "error", err)
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.AggregatedResult = aggregated
		}
	}

	// Determine status
	result.Status = e.determineStatus(results, execErr)
	if execErr != nil {
		result.Errors = append(result.Errors, execErr.Error())
	}

	return result, nil
}

// executeSequential executes agents one after another.
func (e *EnsembleExecutor) executeSequential(ctx context.Context, ensemble *domain.Ensemble, req ExecutionRequest, monitor *ExecutionMonitor) ([]AgentResult, error) {
	results := make([]AgentResult, 0, len(ensemble.Members))
	sharedContext := e.initializeSharedContext(ensemble, req.Input)

	for _, member := range ensemble.Members {
		select {
		case <-ctx.Done():
			return results, errors.New("execution timeout or cancelled")
		default:
		}

		if monitor != nil {
			monitor.StartAgent(member.AgentID, member.Role)
		}

		result := e.executeAgent(ctx, member, sharedContext, req.Options)
		results = append(results, result)

		if monitor != nil {
			if result.Status == common.StatusSuccess {
				monitor.CompleteAgent(member.AgentID)
			} else {
				monitor.FailAgent(member.AgentID, result.Error)
			}
		}

		// Update shared context with result
		if result.Status == common.StatusSuccess && result.Result != nil {
			sharedContext[fmt.Sprintf("agent_%s_result", member.AgentID)] = result.Result
		}

		// Check fail-fast
		if req.Options.FailFast && result.Status == common.StatusFailed {
			return results, fmt.Errorf("execution failed at agent %s: %s", member.AgentID, result.Error)
		}

		// Try fallback chain if current agent failed
		if result.Status == common.StatusFailed && len(ensemble.FallbackChain) > 0 {
			fallbackResult := e.tryFallbackChain(ctx, ensemble.FallbackChain, sharedContext, req.Options)
			if fallbackResult.Status == "success" {
				results = append(results, fallbackResult)
				break // Success with fallback, stop sequential execution
			}
		}
	}

	return results, nil
}

// executeParallel executes all agents concurrently.
func (e *EnsembleExecutor) executeParallel(ctx context.Context, ensemble *domain.Ensemble, req ExecutionRequest, monitor *ExecutionMonitor) ([]AgentResult, error) {
	results := make([]AgentResult, len(ensemble.Members))
	sharedContext := e.initializeSharedContext(ensemble, req.Input)

	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(ensemble.Members))

	for i, member := range ensemble.Members {
		if monitor != nil {
			monitor.StartAgent(member.AgentID, member.Role)
		}

		wg.Add(1)
		go func(idx int, m domain.EnsembleMember) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				mu.Lock()
				results[idx] = AgentResult{
					AgentID: m.AgentID,
					Role:    m.Role,
					Status:  common.StatusFailed,
					Error:   "execution timeout or cancelled",
				}
				mu.Unlock()
				if monitor != nil {
					monitor.FailAgent(m.AgentID, "execution timeout or cancelled")
				}
				return
			default:
			}

			result := e.executeAgent(ctx, m, sharedContext, req.Options)

			if monitor != nil {
				if result.Status == common.StatusSuccess {
					monitor.CompleteAgent(m.AgentID)
				} else {
					monitor.FailAgent(m.AgentID, result.Error)
				}
			}

			mu.Lock()
			results[idx] = result
			mu.Unlock()

			if result.Status == common.StatusFailed && req.Options.FailFast {
				errChan <- fmt.Errorf("agent %s failed: %s", m.AgentID, result.Error)
			}
		}(i, member)
	}

	// Wait for all agents to complete
	wg.Wait()
	close(errChan)

	// Check for fail-fast errors
	if req.Options.FailFast {
		select {
		case err := <-errChan:
			if err != nil {
				return results, err
			}
		default:
		}
	}

	return results, nil
}

// executeHybrid combines sequential and parallel execution.
func (e *EnsembleExecutor) executeHybrid(ctx context.Context, ensemble *domain.Ensemble, req ExecutionRequest, monitor *ExecutionMonitor) ([]AgentResult, error) {
	results := make([]AgentResult, 0, len(ensemble.Members))
	sharedContext := e.initializeSharedContext(ensemble, req.Input)

	// Group members by priority
	priorityGroups := e.groupByPriority(ensemble.Members)

	// Execute each priority group sequentially, but members within group in parallel
	for priority := 10; priority >= 1; priority-- {
		members, exists := priorityGroups[priority]
		if !exists {
			continue
		}

		if monitor != nil {
			monitor.SetPhase(fmt.Sprintf("priority-group-%d", priority))
		}

		groupResults := make([]AgentResult, len(members))
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i, member := range members {
			if monitor != nil {
				monitor.StartAgent(member.AgentID, member.Role)
			}

			wg.Add(1)
			go func(idx int, m domain.EnsembleMember) {
				defer wg.Done()

				result := e.executeAgent(ctx, m, sharedContext, req.Options)

				mu.Lock()
				groupResults[idx] = result
				// Update shared context
				if result.Status == common.StatusSuccess && result.Result != nil {
					sharedContext[fmt.Sprintf("agent_%s_result", m.AgentID)] = result.Result
				}
				mu.Unlock()

				if monitor != nil {
					if result.Status == common.StatusSuccess {
						monitor.CompleteAgent(m.AgentID)
					} else {
						monitor.FailAgent(m.AgentID, result.Error)
					}
				}
			}(i, member)
		}

		wg.Wait()
		results = append(results, groupResults...)

		// Check if we should continue to next priority group
		if req.Options.FailFast && e.hasFailures(groupResults) {
			return results, fmt.Errorf("execution failed in priority group %d", priority)
		}
	}

	return results, nil
}

// executeAgent executes a single agent.
func (e *EnsembleExecutor) executeAgent(ctx context.Context, member domain.EnsembleMember, sharedContext map[string]interface{}, options ExecutionOptions) AgentResult {
	startTime := time.Now()

	result := AgentResult{
		AgentID:   member.AgentID,
		Role:      member.Role,
		StartedAt: startTime,
		Metadata:  make(map[string]interface{}),
	}

	// Load agent
	agent, err := e.loadAgent(member.AgentID)
	if err != nil {
		result.Status = common.StatusFailed
		result.Error = fmt.Sprintf("failed to load agent: %v", err)
		result.FinishedAt = time.Now()
		result.ExecutionTime = time.Since(startTime)
		return result
	}

	// Execute agent with retries
	var execResult interface{}
	var execErr error
	maxRetries := options.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Note: This is a placeholder for actual agent execution
		// In a real implementation, you would:
		// 1. Prepare agent input from sharedContext
		// 2. Execute agent's skill or workflow
		// 3. Capture the result
		execResult, execErr = e.executeAgentLogic(ctx, agent, sharedContext)

		if execErr == nil {
			break
		}

		if attempt < maxRetries {
			e.logger.Warn("Agent execution failed, retrying",
				"agent_id", member.AgentID,
				"attempt", attempt,
				"error", execErr)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}
	}

	result.FinishedAt = time.Now()
	result.ExecutionTime = time.Since(startTime)

	if execErr != nil {
		result.Status = common.StatusFailed
		result.Error = execErr.Error()
	} else {
		result.Status = common.StatusSuccess
		result.Result = execResult
	}

	return result
}

// executeAgentLogic is a placeholder for actual agent execution logic.
func (e *EnsembleExecutor) executeAgentLogic(ctx context.Context, agent *domain.Agent, input map[string]interface{}) (interface{}, error) {
	// Note: This is a simplified placeholder implementation
	// In a real implementation, this would:
	// 1. Load the agent's goals and actions
	// 2. Execute each action according to the agent's decision tree
	// 3. Process results and handle fallback strategies
	// 4. Return aggregated output

	e.logger.Info("Executing agent", "agent_id", agent.GetID())

	// For now, return a simple execution result
	return map[string]interface{}{
		"agent_id": agent.GetID(),
		"status":   "executed",
		"input":    input,
		"goals":    agent.Goals,
		"actions":  len(agent.Actions),
	}, nil
}

// loadEnsemble loads an ensemble from repository.
func (e *EnsembleExecutor) loadEnsemble(ensembleID string) (*domain.Ensemble, error) {
	element, err := e.repository.GetByID(ensembleID)
	if err != nil {
		return nil, err
	}

	ensemble, ok := element.(*domain.Ensemble)
	if !ok {
		return nil, fmt.Errorf("element %s is not an ensemble", ensembleID)
	}

	return ensemble, nil
}

// loadAgent loads an agent from repository.
func (e *EnsembleExecutor) loadAgent(agentID string) (*domain.Agent, error) {
	element, err := e.repository.GetByID(agentID)
	if err != nil {
		return nil, err
	}

	agent, ok := element.(*domain.Agent)
	if !ok {
		return nil, fmt.Errorf("element %s is not an agent", agentID)
	}

	return agent, nil
}

// initializeSharedContext creates the initial shared context.
func (e *EnsembleExecutor) initializeSharedContext(ensemble *domain.Ensemble, input map[string]interface{}) map[string]interface{} {
	ctx := make(map[string]interface{})

	// Copy ensemble's shared context
	for k, v := range ensemble.SharedContext {
		ctx[k] = v
	}

	// Add input
	ctx["input"] = input
	ctx["ensemble_id"] = ensemble.GetID()

	return ctx
}

// tryFallbackChain tries to execute fallback agents.
func (e *EnsembleExecutor) tryFallbackChain(ctx context.Context, fallbackChain []string, sharedContext map[string]interface{}, options ExecutionOptions) AgentResult {
	for _, agentID := range fallbackChain {
		member := domain.EnsembleMember{
			AgentID:  agentID,
			Role:     "fallback",
			Priority: 1,
		}

		result := e.executeAgent(ctx, member, sharedContext, options)
		if result.Status == common.StatusSuccess {
			return result
		}
	}

	return AgentResult{
		AgentID: "fallback_chain",
		Role:    "fallback",
		Status:  common.StatusFailed,
		Error:   "all fallback agents failed",
	}
}

// groupByPriority groups members by their priority.
func (e *EnsembleExecutor) groupByPriority(members []domain.EnsembleMember) map[int][]domain.EnsembleMember {
	groups := make(map[int][]domain.EnsembleMember)
	for _, member := range members {
		groups[member.Priority] = append(groups[member.Priority], member)
	}
	return groups
}

// hasFailures checks if any result failed.
func (e *EnsembleExecutor) hasFailures(results []AgentResult) bool {
	for _, result := range results {
		if result.Status == common.StatusFailed {
			return true
		}
	}
	return false
}

// aggregateResults aggregates agent results based on strategy.
func (e *EnsembleExecutor) aggregateResults(ensemble *domain.Ensemble, results []AgentResult) (interface{}, error) {
	successResults := make([]interface{}, 0)
	for _, result := range results {
		if result.Status == common.StatusSuccess && result.Result != nil {
			successResults = append(successResults, result.Result)
		}
	}

	if len(successResults) == 0 {
		return nil, errors.New("no successful results to aggregate")
	}

	switch ensemble.AggregationStrategy {
	case "first":
		return successResults[0], nil

	case "last":
		return successResults[len(successResults)-1], nil

	case common.SelectorAll:
		return successResults, nil

	case "consensus":
		// Advanced consensus with 70% agreement threshold
		config := ConsensusConfig{
			Threshold:      0.7,
			WeightedVoting: true,
		}
		return e.aggregateByConsensus(results, config)

	case "voting":
		// Advanced voting with priority-based weights
		config := VotingConfig{
			WeightByPriority: true,
			MinimumVotes:     1,
			TieBreaker:       "highest_priority",
		}
		return e.aggregateByVoting(results, config)

	case "weighted_consensus":
		// Weighted consensus with confidence scores
		return e.aggregateByWeightedConsensus(results, 0.6)

	case "threshold_consensus":
		// Threshold consensus requiring 80% agreement
		quorum := len(results) / 2 // Require at least half participation
		return e.aggregateByThresholdConsensus(results, 0.8, quorum)

	case "merge":
		// Merge all results into a single map
		merged := make(map[string]interface{})
		for i, result := range successResults {
			if resultMap, ok := result.(map[string]interface{}); ok {
				for k, v := range resultMap {
					merged[fmt.Sprintf("agent_%d_%s", i, k)] = v
				}
			}
		}
		return merged, nil

	default:
		return nil, fmt.Errorf("unsupported aggregation strategy: %s", ensemble.AggregationStrategy)
	}
}

// determineStatus determines overall execution status.
func (e *EnsembleExecutor) determineStatus(results []AgentResult, execErr error) string {
	if execErr != nil {
		return common.StatusFailed
	}

	successCount := 0
	failedCount := 0

	for _, result := range results {
		switch result.Status {
		case common.StatusSuccess:
			successCount++
		case common.StatusFailed:
			failedCount++
		}
	}

	if failedCount == 0 {
		return common.StatusSuccess
	}

	if successCount == 0 {
		return common.StatusFailed
	}

	return "partial_success"
}
