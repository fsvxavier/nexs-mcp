# NEXS MCP Application Layer

**Version:** 1.0.0  
**Last Updated:** December 20, 2025  
**Status:** Production

---

## Table of Contents

- [Introduction](#introduction)
- [Application Layer Purpose](#application-layer-purpose)
- [Use Cases](#use-cases)
- [EnsembleExecutor Service](#ensembleexecutor-service)
- [EnsembleMonitor Service](#ensemblemonitor-service)
- [EnsembleAggregation Service](#ensembleaggregation-service)
- [MetricsCollector Service](#metricscollector-service)
- [StatisticsService](#statisticsservice)
- [Orchestration Patterns](#orchestration-patterns)
- [Error Handling](#error-handling)
- [Performance Considerations](#performance-considerations)
- [Testing Strategies](#testing-strategies)
- [Best Practices](#best-practices)

---

## Introduction

The **Application Layer** contains use cases, services, and orchestration logic. It sits between the pure Domain Layer and the Infrastructure Layer, coordinating domain entities to fulfill business workflows.

### Application Layer Location

```
internal/application/
├── ensemble_executor.go          # Ensemble execution orchestration
├── ensemble_monitor.go           # Real-time execution monitoring
├── ensemble_aggregation.go       # Result aggregation strategies
├── statistics.go                 # Usage statistics and metrics
├── *_test.go                     # Unit tests
```

### Dependencies

```
Application Layer
      │
      ├─→ Domain Layer (interfaces only)
      │
      └─→ Standard Library (time, context, sync, etc.)
```

**Depends On:**
- Domain interfaces (ElementRepository, Element)
- Standard library packages
- **NOT** Infrastructure implementations

---

## Application Layer Purpose

### What Application Layer Does

✅ **Orchestrate Domain Entities**
- Coordinate multiple domain objects
- Execute complex workflows
- Manage transactions

✅ **Implement Use Cases**
- Execute ensemble (sequential, parallel, hybrid)
- Aggregate agent results
- Collect and analyze metrics

✅ **Apply Cross-Cutting Concerns**
- Logging (via injected logger)
- Metrics collection
- Performance monitoring

✅ **Handle Business Workflows**
- Multi-step processes
- Error recovery strategies
- Result aggregation

### What Application Layer Does NOT Do

❌ **Infrastructure Details**
- File I/O operations
- HTTP requests
- Database queries
- External service calls

❌ **Domain Logic**
- Validation rules
- Business invariants
- Entity behavior

❌ **Protocol Details**
- MCP message formatting
- HTTP handling
- CLI parsing

---

## Use Cases

The Application Layer implements these key use cases:

### 1. Ensemble Execution

**Use Case:** Execute ensemble of agents with orchestration

```go
// Execute ensemble with specific mode
func (e *EnsembleExecutor) Execute(
    ctx context.Context, 
    req ExecutionRequest,
) (*ExecutionResult, error)
```

**Modes:**
- **Sequential** - One agent at a time
- **Parallel** - All agents concurrently
- **Hybrid** - Mix of sequential and parallel

### 2. Result Aggregation

**Use Case:** Combine agent results using strategies

```go
// Aggregate results using specified strategy
func (e *EnsembleExecutor) aggregateResults(
    results []AgentResult,
    strategy string,
) (interface{}, error)
```

**Strategies:**
- **first** - Return first successful result
- **last** - Return last result
- **consensus** - Majority voting with threshold
- **voting** - Weighted voting system
- **all** - Return all results
- **merge** - Merge all results

### 3. Execution Monitoring

**Use Case:** Track ensemble execution progress

```go
// Monitor execution with callbacks
monitor := NewExecutionMonitor(executionID, ensembleID, totalAgents)
monitor.RegisterProgressCallback(callback)
```

### 4. Metrics Collection

**Use Case:** Record and analyze tool usage

```go
// Record tool call metrics
collector.RecordToolCall(ToolCallMetric{
    ToolName:  "create_element",
    Timestamp: time.Now(),
    Duration:  duration,
    Success:   true,
})
```

### 5. Statistics Generation

**Use Case:** Generate usage statistics

```go
// Get statistics for time period
stats := collector.GetStatistics(period)
```

---

## EnsembleExecutor Service

### Overview

**EnsembleExecutor** orchestrates multi-agent execution with support for sequential, parallel, and hybrid modes.

### Structure

```go
type EnsembleExecutor struct {
    repository domain.ElementRepository  // Domain interface
    logger     *slog.Logger
}

func NewEnsembleExecutor(
    repository domain.ElementRepository,
    log *slog.Logger,
) *EnsembleExecutor {
    if log == nil {
        log = logger.Get()
    }
    return &EnsembleExecutor{
        repository: repository,
        logger:     log,
    }
}
```

### Execution Request

```go
type ExecutionRequest struct {
    EnsembleID string                 `json:"ensemble_id"`
    Input      map[string]interface{} `json:"input"`
    Options    ExecutionOptions       `json:"options,omitempty"`
}

type ExecutionOptions struct {
    Timeout          time.Duration `json:"timeout,omitempty"`
    MaxRetries       int           `json:"max_retries,omitempty"`
    FailFast         bool          `json:"fail_fast,omitempty"`
    CollectAll       bool          `json:"collect_all,omitempty"`
    EnableMonitoring bool          `json:"enable_monitoring,omitempty"`
}
```

### Execution Result

```go
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
```

### Sequential Execution

Agents run one after another, passing context:

```go
func (e *EnsembleExecutor) executeSequential(
    ctx context.Context,
    ensemble *domain.Ensemble,
    req ExecutionRequest,
) (*ExecutionResult, error) {
    results := make([]AgentResult, 0, len(ensemble.Members))
    sharedContext := req.Input
    
    // Execute agents in order
    for _, member := range ensemble.Members {
        // Check context cancellation
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        // Load agent
        agent, err := e.loadAgent(member.AgentID)
        if err != nil {
            if req.Options.FailFast {
                return nil, fmt.Errorf("failed to load agent %s: %w", member.AgentID, err)
            }
            results = append(results, AgentResult{
                AgentID: member.AgentID,
                Role:    member.Role,
                Status:  "failed",
                Error:   err.Error(),
            })
            continue
        }
        
        // Execute agent
        agentResult := e.executeAgent(ctx, agent, member, sharedContext, req.Options)
        results = append(results, agentResult)
        
        // Update shared context with result
        if agentResult.Status == "success" && agentResult.Result != nil {
            sharedContext[member.Role] = agentResult.Result
        }
        
        // Fail fast on error
        if req.Options.FailFast && agentResult.Status == "failed" {
            break
        }
    }
    
    return e.buildExecutionResult(ensemble.GetID(), results, req), nil
}
```

**Flow Diagram:**

```
Agent 1 → Result 1
          │
          ├─ Update Context
          │
          ▼
Agent 2 → Result 2
          │
          ├─ Update Context
          │
          ▼
Agent 3 → Result 3
          │
          ▼
    Aggregate
```

### Parallel Execution

All agents run concurrently:

```go
func (e *EnsembleExecutor) executeParallel(
    ctx context.Context,
    ensemble *domain.Ensemble,
    req ExecutionRequest,
) (*ExecutionResult, error) {
    results := make([]AgentResult, len(ensemble.Members))
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    // Execute all agents concurrently
    for i, member := range ensemble.Members {
        wg.Add(1)
        go func(index int, m domain.EnsembleMember) {
            defer wg.Done()
            
            // Load agent
            agent, err := e.loadAgent(m.AgentID)
            if err != nil {
                mu.Lock()
                results[index] = AgentResult{
                    AgentID: m.AgentID,
                    Role:    m.Role,
                    Status:  "failed",
                    Error:   err.Error(),
                }
                mu.Unlock()
                return
            }
            
            // Execute agent
            result := e.executeAgent(ctx, agent, m, req.Input, req.Options)
            
            mu.Lock()
            results[index] = result
            mu.Unlock()
        }(i, member)
    }
    
    // Wait for all agents to complete
    wg.Wait()
    
    return e.buildExecutionResult(ensemble.GetID(), results, req), nil
}
```

**Flow Diagram:**

```
         ┌─ Agent 1 → Result 1
         │
Input ───┼─ Agent 2 → Result 2
         │
         └─ Agent 3 → Result 3
                │
                ▼
            Aggregate
```

### Hybrid Execution

Mix sequential and parallel based on priorities:

```go
func (e *EnsembleExecutor) executeHybrid(
    ctx context.Context,
    ensemble *domain.Ensemble,
    req ExecutionRequest,
) (*ExecutionResult, error) {
    // Group agents by priority
    priorityGroups := e.groupByPriority(ensemble.Members)
    
    allResults := make([]AgentResult, 0)
    sharedContext := req.Input
    
    // Execute each priority group sequentially
    for _, priority := range e.sortedPriorities(priorityGroups) {
        group := priorityGroups[priority]
        
        // Execute agents in group concurrently
        groupResults := e.executeGroup(ctx, group, sharedContext, req.Options)
        allResults = append(allResults, groupResults...)
        
        // Update shared context with successful results
        for _, result := range groupResults {
            if result.Status == "success" && result.Result != nil {
                sharedContext[result.Role] = result.Result
            }
        }
        
        // Check fail fast
        if req.Options.FailFast && e.hasFailures(groupResults) {
            break
        }
    }
    
    return e.buildExecutionResult(ensemble.GetID(), allResults, req), nil
}
```

**Flow Diagram:**

```
Priority 10: Agent A ──┐
Priority 10: Agent B ──┼─ Parallel ─→ Results 1
Priority 10: Agent C ──┘              │
                                      ▼
Priority 5:  Agent D ──┐       Update Context
Priority 5:  Agent E ──┼─ Parallel ─→ Results 2
                                      │
                                      ▼
Priority 1:  Agent F ─────────────→ Result 3
                                      │
                                      ▼
                                  Aggregate
```

### Retry Logic

```go
func (e *EnsembleExecutor) executeAgent(
    ctx context.Context,
    agent *domain.Agent,
    member domain.EnsembleMember,
    input map[string]interface{},
    options ExecutionOptions,
) AgentResult {
    startTime := time.Now()
    
    var lastErr error
    maxRetries := options.MaxRetries
    if maxRetries <= 0 {
        maxRetries = 1
    }
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        // Execute agent logic
        result, err := e.runAgent(ctx, agent, input)
        
        if err == nil {
            return AgentResult{
                AgentID:       member.AgentID,
                Role:          member.Role,
                Status:        "success",
                Result:        result,
                ExecutionTime: time.Since(startTime),
                StartedAt:     startTime,
                FinishedAt:    time.Now(),
            }
        }
        
        lastErr = err
        
        // Wait before retry (exponential backoff)
        if attempt < maxRetries-1 {
            backoff := time.Duration(attempt+1) * 100 * time.Millisecond
            time.Sleep(backoff)
        }
    }
    
    // All retries failed
    return AgentResult{
        AgentID:       member.AgentID,
        Role:          member.Role,
        Status:        "failed",
        Error:         lastErr.Error(),
        ExecutionTime: time.Since(startTime),
        StartedAt:     startTime,
        FinishedAt:    time.Now(),
    }
}
```

### Timeout Handling

```go
func (e *EnsembleExecutor) Execute(
    ctx context.Context,
    req ExecutionRequest,
) (*ExecutionResult, error) {
    // Setup context with timeout
    execCtx := ctx
    if req.Options.Timeout > 0 {
        var cancel context.CancelFunc
        execCtx, cancel = context.WithTimeout(ctx, req.Options.Timeout)
        defer cancel()
    }
    
    // Execute with timeout context
    result, err := e.executeWithMode(execCtx, ensemble, req)
    
    // Check if timeout occurred
    if errors.Is(err, context.DeadlineExceeded) {
        return &ExecutionResult{
            EnsembleID: req.EnsembleID,
            Status:     "failed",
            Errors:     []string{"execution timeout"},
        }, err
    }
    
    return result, err
}
```

---

## EnsembleMonitor Service

### Overview

**EnsembleMonitor** provides real-time tracking of ensemble execution with progress updates and callbacks.

### Structure

```go
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
```

### Progress Callbacks

```go
type ProgressCallback func(monitor *ExecutionMonitor)
type StateCallback func(monitor *ExecutionMonitor, oldState, newState string)

// Register callback
monitor := NewExecutionMonitor(executionID, ensembleID, 5)
monitor.RegisterProgressCallback(func(m *ExecutionMonitor) {
    update := m.GetProgressUpdate()
    logger.Info("Progress update",
        "ensemble", update.EnsembleID,
        "progress", update.Progress,
        "completed", update.CompletedAgents,
        "total", update.TotalAgents,
    )
})
```

### Agent Tracking

```go
// Start agent execution
monitor.StartAgent("agent-123", "analyzer")

// Update progress (0.0 to 1.0)
monitor.UpdateAgentProgress("agent-123", 0.5, map[string]interface{}{
    "current_step": "analyzing files",
})

// Complete agent
monitor.CompleteAgent("agent-123")

// Or fail agent
monitor.FailAgent("agent-123", "timeout exceeded")
```

### Progress Update

```go
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

// Get current progress
update := monitor.GetProgressUpdate()
```

### Status Transitions

```
initializing → running → completing → completed
                  │
                  └────→ failed
```

### Example Usage

```go
// Create monitor
monitor := NewExecutionMonitor("exec-123", "ensemble-456", 3)
monitor.SetStatus("running")
monitor.SetPhase("initialization")

// Register callback
monitor.RegisterProgressCallback(func(m *ExecutionMonitor) {
    progress := m.GetProgressUpdate()
    fmt.Printf("Progress: %.1f%%\n", progress.Progress*100)
})

// Track agents
monitor.StartAgent("agent-1", "analyzer")
monitor.UpdateAgentProgress("agent-1", 0.5, nil)
monitor.CompleteAgent("agent-1")

monitor.StartAgent("agent-2", "reviewer")
monitor.UpdateAgentProgress("agent-2", 0.3, nil)
monitor.CompleteAgent("agent-2")

monitor.SetStatus("completed")
```

---

## EnsembleAggregation Service

### Overview

**EnsembleAggregation** implements sophisticated result aggregation strategies including consensus and voting algorithms.

### Aggregation Strategies

| Strategy | Description | Use Case |
|----------|-------------|----------|
| **first** | Return first successful result | Fast response, single winner |
| **last** | Return last result | Final override |
| **consensus** | Majority agreement with threshold | Democratic decision |
| **voting** | Weighted voting system | Priority-based decision |
| **all** | Return all results | Complete information |
| **merge** | Merge all results | Combined output |

### First Strategy

```go
func (e *EnsembleExecutor) aggregateByFirst(results []AgentResult) interface{} {
    for _, result := range results {
        if result.Status == "success" && result.Result != nil {
            return result.Result
        }
    }
    return nil
}
```

### Last Strategy

```go
func (e *EnsembleExecutor) aggregateByLast(results []AgentResult) interface{} {
    for i := len(results) - 1; i >= 0; i-- {
        if results[i].Status == "success" && results[i].Result != nil {
            return results[i].Result
        }
    }
    return nil
}
```

### Consensus Strategy

Advanced consensus with configurable threshold:

```go
type ConsensusConfig struct {
    Threshold      float64 // Minimum agreement (0.0 to 1.0)
    RequireQuorum  bool    // Require minimum participants
    QuorumSize     int     // Minimum participants
    WeightedVoting bool    // Use agent priority as weight
}

type ConsensusResult struct {
    Value            interface{}            `json:"value"`
    AgreementLevel   float64                `json:"agreement_level"`
    Participants     int                    `json:"participants"`
    Supporting       []string               `json:"supporting"`
    Alternative      map[string]interface{} `json:"alternative,omitempty"`
    ReachedConsensus bool                   `json:"reached_consensus"`
}

func (e *EnsembleExecutor) aggregateByConsensus(
    results []AgentResult,
    config ConsensusConfig,
) (*ConsensusResult, error) {
    // Filter successful results
    successResults := filterSuccessful(results)
    
    // Check quorum
    if config.RequireQuorum && len(successResults) < config.QuorumSize {
        return nil, fmt.Errorf("quorum not met")
    }
    
    // Group similar results
    groups := groupSimilarResults(successResults)
    
    // Find largest group
    largestGroup := findLargestGroup(groups)
    
    // Calculate agreement level
    agreementLevel := calculateAgreement(
        largestGroup,
        successResults,
        config.WeightedVoting,
    )
    
    return &ConsensusResult{
        Value:            largestGroup[0].Result,
        AgreementLevel:   agreementLevel,
        Participants:     len(successResults),
        Supporting:       extractAgentIDs(largestGroup),
        ReachedConsensus: agreementLevel >= config.Threshold,
    }, nil
}
```

**Example:**

```
3 agents with results:
Agent A (priority 10): "Option X"
Agent B (priority 8):  "Option X"
Agent C (priority 5):  "Option Y"

Without weighting: 2/3 = 66.7% agreement
With weighting: (10+8)/(10+8+5) = 78.3% agreement

If threshold = 0.6 → consensus reached on "Option X"
```

### Voting Strategy

```go
type VotingConfig struct {
    WeightByPriority   bool               // Use priority as weight
    WeightByConfidence bool               // Use confidence scores
    MinimumVotes       int                // Minimum votes required
    TieBreaker         string             // "first", "random", "highest_priority"
    CustomWeights      map[string]float64 // Custom weights per agent
}

type VotingResult struct {
    Winner      interface{}        `json:"winner"`
    TotalVotes  float64            `json:"total_votes"`
    WinnerVotes float64            `json:"winner_votes"`
    Percentage  float64            `json:"percentage"`
    Voters      []string           `json:"voters"`
    Breakdown   map[string]float64 `json:"breakdown"`
    TieBreaker  bool               `json:"tie_breaker,omitempty"`
}

func (e *EnsembleExecutor) aggregateByVoting(
    results []AgentResult,
    config VotingConfig,
) (*VotingResult, error) {
    // Count votes with weights
    voteCount := make(map[string]float64)
    votersByOption := make(map[string][]string)
    
    for _, result := range results {
        if result.Status != "success" || result.Result == nil {
            continue
        }
        
        weight := 1.0
        if config.WeightByPriority {
            if priority, ok := result.Metadata["priority"].(int); ok {
                weight = float64(priority) / 10.0
            }
        }
        
        option := fmt.Sprintf("%v", result.Result)
        voteCount[option] += weight
        votersByOption[option] = append(votersByOption[option], result.AgentID)
    }
    
    // Find winner
    var winner string
    var maxVotes float64
    for option, votes := range voteCount {
        if votes > maxVotes {
            winner = option
            maxVotes = votes
        }
    }
    
    // Calculate total votes
    var totalVotes float64
    for _, votes := range voteCount {
        totalVotes += votes
    }
    
    return &VotingResult{
        Winner:      winner,
        TotalVotes:  totalVotes,
        WinnerVotes: maxVotes,
        Percentage:  maxVotes / totalVotes,
        Voters:      votersByOption[winner],
        Breakdown:   voteCount,
    }, nil
}
```

### Merge Strategy

```go
func (e *EnsembleExecutor) aggregateByMerge(results []AgentResult) map[string]interface{} {
    merged := make(map[string]interface{})
    
    for _, result := range results {
        if result.Status == "success" && result.Result != nil {
            // Add result with role as key
            merged[result.Role] = result.Result
        }
    }
    
    return merged
}
```

**Example:**

```json
{
  "analyzer": {"code_quality": 8.5, "issues": 3},
  "reviewer": {"style_score": 9.0, "suggestions": ["improve naming"]},
  "tester": {"coverage": 85.2, "passed": 142, "failed": 3}
}
```

---

## MetricsCollector Service

### Overview

**MetricsCollector** records and aggregates tool usage metrics for analytics.

### Structure

```go
type MetricsCollector struct {
    mu           sync.RWMutex
    metrics      []ToolCallMetric
    maxMetrics   int
    storageDir   string
    autoSave     bool
    saveInterval time.Duration
    lastSaveTime time.Time
}

type ToolCallMetric struct {
    ToolName      string        `json:"tool_name"`
    Timestamp     time.Time     `json:"timestamp"`
    Duration      time.Duration `json:"duration_ms"`
    Success       bool          `json:"success"`
    ErrorMessage  string        `json:"error_message,omitempty"`
    User          string        `json:"user,omitempty"`
    RequestParams interface{}   `json:"request_params,omitempty"`
}
```

### Recording Metrics

```go
func (mc *MetricsCollector) RecordToolCall(metric ToolCallMetric) {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    
    // Add metric
    mc.metrics = append(mc.metrics, metric)
    
    // Trim if exceeds max
    if len(mc.metrics) > mc.maxMetrics {
        mc.metrics = mc.metrics[len(mc.metrics)-mc.maxMetrics:]
    }
    
    // Auto-save if enabled
    if mc.autoSave && time.Since(mc.lastSaveTime) > mc.saveInterval {
        mc.saveMetrics()
        mc.lastSaveTime = time.Now()
    }
}
```

### Usage Example

```go
// Create collector
collector := NewMetricsCollector("~/.nexs-mcp/metrics")

// Record tool call
startTime := time.Now()
err := createElement(...)
collector.RecordToolCall(ToolCallMetric{
    ToolName:  "create_element",
    Timestamp: startTime,
    Duration:  time.Since(startTime),
    Success:   err == nil,
    ErrorMessage: errorToString(err),
    User:      currentUser,
})
```

---

## StatisticsService

### Overview

Aggregates metrics into actionable statistics.

### Statistics Structure

```go
type UsageStatistics struct {
    TotalOperations    int                `json:"total_operations"`
    SuccessfulOps      int                `json:"successful_ops"`
    FailedOps          int                `json:"failed_ops"`
    SuccessRate        float64            `json:"success_rate"`
    OperationsByTool   map[string]int     `json:"operations_by_tool"`
    ErrorsByTool       map[string]int     `json:"errors_by_tool"`
    AvgDurationByTool  map[string]float64 `json:"avg_duration_by_tool_ms"`
    MostUsedTools      []ToolUsageStat    `json:"most_used_tools"`
    SlowestOperations  []ToolCallMetric   `json:"slowest_operations"`
    RecentErrors       []ToolCallMetric   `json:"recent_errors"`
    ActiveUsers        []string           `json:"active_users"`
    OperationsByPeriod map[string]int     `json:"operations_by_period"`
    Period             string             `json:"period"`
    StartTime          time.Time          `json:"start_time"`
    EndTime            time.Time          `json:"end_time"`
}
```

### Generating Statistics

```go
func (mc *MetricsCollector) GetStatistics(period string) UsageStatistics {
    mc.mu.RLock()
    defer mc.mu.RUnlock()
    
    // Filter metrics by period
    periodMetrics := mc.filterByPeriod(period)
    
    stats := UsageStatistics{
        TotalOperations:    len(periodMetrics),
        OperationsByTool:   make(map[string]int),
        ErrorsByTool:       make(map[string]int),
        AvgDurationByTool:  make(map[string]float64),
        Period:             period,
    }
    
    // Aggregate metrics
    durationSums := make(map[string]time.Duration)
    durationCounts := make(map[string]int)
    
    for _, metric := range periodMetrics {
        stats.OperationsByTool[metric.ToolName]++
        
        if metric.Success {
            stats.SuccessfulOps++
        } else {
            stats.FailedOps++
            stats.ErrorsByTool[metric.ToolName]++
        }
        
        durationSums[metric.ToolName] += metric.Duration
        durationCounts[metric.ToolName]++
    }
    
    // Calculate averages
    for tool, sum := range durationSums {
        count := durationCounts[tool]
        stats.AvgDurationByTool[tool] = float64(sum.Milliseconds()) / float64(count)
    }
    
    // Calculate success rate
    if stats.TotalOperations > 0 {
        stats.SuccessRate = float64(stats.SuccessfulOps) / float64(stats.TotalOperations)
    }
    
    return stats
}
```

---

## Orchestration Patterns

### Pattern 1: Pipeline

Sequential processing with data transformation:

```go
func (e *EnsembleExecutor) executePipeline(
    ctx context.Context,
    stages []Agent,
    input interface{},
) (interface{}, error) {
    result := input
    
    for _, agent := range stages {
        result, err = agent.Process(ctx, result)
        if err != nil {
            return nil, err
        }
    }
    
    return result, nil
}
```

### Pattern 2: Fan-Out/Fan-In

Parallel processing with aggregation:

```go
func (e *EnsembleExecutor) executeFanOut(
    ctx context.Context,
    workers []Agent,
    input interface{},
) ([]interface{}, error) {
    results := make([]interface{}, len(workers))
    var wg sync.WaitGroup
    errChan := make(chan error, len(workers))
    
    // Fan out
    for i, worker := range workers {
        wg.Add(1)
        go func(index int, w Agent) {
            defer wg.Done()
            result, err := w.Process(ctx, input)
            if err != nil {
                errChan <- err
                return
            }
            results[index] = result
        }(i, worker)
    }
    
    wg.Wait()
    close(errChan)
    
    // Check errors
    if len(errChan) > 0 {
        return nil, <-errChan
    }
    
    return results, nil
}
```

### Pattern 3: Circuit Breaker

Fail fast with fallback:

```go
type CircuitBreaker struct {
    failures     int
    threshold    int
    timeout      time.Duration
    lastFailTime time.Time
    state        string // closed, open, half-open
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    if cb.state == "open" {
        if time.Since(cb.lastFailTime) > cb.timeout {
            cb.state = "half-open"
        } else {
            return fmt.Errorf("circuit breaker open")
        }
    }
    
    err := fn()
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        if cb.failures >= cb.threshold {
            cb.state = "open"
        }
        return err
    }
    
    cb.failures = 0
    cb.state = "closed"
    return nil
}
```

---

## Error Handling

### Error Propagation

```go
// Application error wraps domain errors
func (e *EnsembleExecutor) Execute(...) (*ExecutionResult, error) {
    ensemble, err := e.loadEnsemble(req.EnsembleID)
    if errors.Is(err, domain.ErrElementNotFound) {
        return nil, fmt.Errorf("ensemble not found: %w", err)
    }
    
    // Continue processing
}
```

### Partial Success

```go
func (e *EnsembleExecutor) buildExecutionResult(
    ensembleID string,
    results []AgentResult,
    req ExecutionRequest,
) *ExecutionResult {
    successCount := 0
    failures := make([]string, 0)
    
    for _, result := range results {
        if result.Status == "success" {
            successCount++
        } else {
            failures = append(failures, fmt.Sprintf("%s: %s", result.AgentID, result.Error))
        }
    }
    
    status := "success"
    if successCount == 0 {
        status = "failed"
    } else if len(failures) > 0 {
        status = "partial_success"
    }
    
    return &ExecutionResult{
        EnsembleID: ensembleID,
        Status:     status,
        Results:    results,
        Errors:     failures,
    }
}
```

---

## Performance Considerations

### Concurrency Control

```go
// Limit concurrent agents
const maxConcurrent = 10

func (e *EnsembleExecutor) executeParallelLimited(
    ctx context.Context,
    agents []Agent,
) []Result {
    semaphore := make(chan struct{}, maxConcurrent)
    results := make([]Result, len(agents))
    var wg sync.WaitGroup
    
    for i, agent := range agents {
        wg.Add(1)
        go func(index int, a Agent) {
            defer wg.Done()
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release
            
            results[index] = a.Execute(ctx)
        }(i, agent)
    }
    
    wg.Wait()
    return results
}
```

### Memory Management

```go
// Stream large results
func (mc *MetricsCollector) StreamMetrics(
    writer io.Writer,
    period string,
) error {
    mc.mu.RLock()
    defer mc.mu.RUnlock()
    
    encoder := json.NewEncoder(writer)
    for _, metric := range mc.metrics {
        if mc.matchesPeriod(metric, period) {
            if err := encoder.Encode(metric); err != nil {
                return err
            }
        }
    }
    return nil
}
```

---

## Testing Strategies

### Mocking Repository

```go
type mockRepository struct {
    elements map[string]domain.Element
}

func (m *mockRepository) GetByID(id string) (domain.Element, error) {
    if elem, ok := m.elements[id]; ok {
        return elem, nil
    }
    return nil, domain.ErrElementNotFound
}

func TestEnsembleExecutor(t *testing.T) {
    repo := &mockRepository{
        elements: map[string]domain.Element{
            "ensemble-1": testEnsemble,
            "agent-1": testAgent1,
        },
    }
    
    executor := NewEnsembleExecutor(repo, nil)
    result, err := executor.Execute(context.Background(), req)
    
    assert.NoError(t, err)
    assert.Equal(t, "success", result.Status)
}
```

---

## Best Practices

### 1. Use Dependency Injection

```go
// ✅ Good: Dependencies injected
func NewEnsembleExecutor(
    repo domain.ElementRepository,
    logger *slog.Logger,
) *EnsembleExecutor
```

### 2. Handle Context Cancellation

```go
// ✅ Good: Check context
select {
case <-ctx.Done():
    return ctx.Err()
default:
    // Continue
}
```

### 3. Avoid Business Logic

```go
// ❌ Bad: Business validation in application
if len(ensemble.Members) == 0 {
    return fmt.Errorf("no members")
}

// ✅ Good: Use domain validation
if err := ensemble.Validate(); err != nil {
    return err
}
```

### 4. Log at Boundaries

```go
func (e *EnsembleExecutor) Execute(...) {
    e.logger.Info("Starting ensemble execution",
        "ensemble_id", req.EnsembleID,
        "mode", ensemble.ExecutionMode)
    // ...
}
```

---

## Conclusion

The Application Layer orchestrates domain entities to fulfill complex business workflows. Through services like EnsembleExecutor, EnsembleMonitor, and MetricsCollector, it provides sophisticated features while maintaining clean separation from infrastructure concerns.

**Key Points:**
- Orchestrate don't dictate
- Use domain interfaces only
- Handle cross-cutting concerns
- Test with mocks
- Monitor performance

---

**Document Version:** 1.0.0  
**Total Lines:** 1054  
**Last Updated:** December 20, 2025
