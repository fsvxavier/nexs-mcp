# NEXS MCP Application Layer

**Version:** 1.4.0
**Last Updated:** January 4, 2026
**Status:** Production

---

## Table of Contents

- [Introduction](#introduction)
- [Application Layer Purpose](#application-layer-purpose)
- [Services Overview](#services-overview)
- [Use Cases](#use-cases)
- [EnsembleExecutor Service](#ensembleexecutor-service)
- [EnsembleMonitor Service](#ensemblemonitor-service)
- [EnsembleAggregation Service](#ensembleaggregation-service)
- [MetricsCollector Service](#metricscollector-service)
- [StatisticsService](#statisticsservice)
- [Memory Consolidation Services](#memory-consolidation-services-new-in-v130)
  - [DuplicateDetection Service](#duplicatedetection-service)
  - [Clustering Service](#clustering-service)
  - [KnowledgeGraphExtractor Service](#knowledgegraphextractor-service)
  - [MemoryConsolidation Service](#memoryconsolidation-service)
  - [HybridSearch Service](#hybridsearch-service)
  - [MemoryRetention Service](#memoryretention-service)
  - [SemanticSearch Service](#semanticsearch-service)
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
├── onnx_bert_provider.go         # ⚡ NEW: ONNX BERT/DistilBERT provider (v1.4.0)
├── onnx_bert_provider_stub.go    # ⚡ NEW: Stub for noonnx builds (v1.4.0)
├── enhanced_entity_extractor.go  # ⚡ NEW: Transformer-based entity extraction (v1.4.0)
├── sentiment_analyzer.go         # ⚡ NEW: Multilingual sentiment analysis (v1.4.0)
├── topic_modeler.go              # ⚡ NEW: LDA/NMF topic modeling (v1.4.0)
├── duplicate_detection.go        # HNSW-based duplicate detection (v1.3.0)
├── clustering.go                 # DBSCAN/K-means clustering (v1.3.0)
├── knowledge_graph_extractor.go  # NLP entity extraction (v1.3.0)
├── memory_consolidation.go       # Consolidation orchestration (v1.3.0)
├── hybrid_search.go              # HNSW hybrid search (v1.3.0)
├── memory_retention.go           # Quality-based retention (v1.3.0)
├── semantic_search.go            # Semantic search service (v1.3.0)
├── *_test.go                     # Unit tests (295+ tests, 76.4% coverage)
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

## Services Overview

The Application Layer provides **24 services** organized into 5 categories:

### Core Services (5 services)
- **EnsembleExecutor** - Multi-agent orchestration with sequential/parallel/hybrid modes
- **EnsembleMonitor** - Real-time execution progress tracking
- **EnsembleAggregation** - Result aggregation strategies (first, last, consensus, voting, merge)
- **MetricsCollector** - Tool usage metrics and analytics
- **StatisticsService** - Usage statistics generation

### NLP & Analytics Services (4 services) ⚡ NEW in v1.4.0
- **ONNXBERTProvider** - ONNX Runtime provider for BERT/DistilBERT models (641 LOC)
- **EnhancedEntityExtractor** - Transformer-based entity extraction with 9 types + 10 relationships (432 LOC)
- **SentimentAnalyzer** - Multilingual sentiment with emotional dimensions (418 LOC)
- **TopicModeler** - LDA/NMF topic modeling with coherence scoring (653 LOC)

### Memory Consolidation Services (7 services) ⚡ v1.3.0
- **DuplicateDetection** - HNSW-based similarity detection and merging
- **Clustering** - DBSCAN and K-means clustering algorithms
- **KnowledgeGraphExtractor** - NLP-based entity and relationship extraction
- **MemoryConsolidation** - End-to-end consolidation workflow orchestration
- **HybridSearch** - HNSW/linear search with automatic mode selection
- **MemoryRetention** - Quality-based retention policies and cleanup
- **SemanticSearch** - Semantic indexing and search across element types

### Token Optimization Services (8 services)
- **PromptCompression** - Redundancy removal and whitespace compression
- **AdaptiveCache** - Dynamic TTL adjustment based on access patterns
- **Summarization** - Extractive and truncation-based summarization
- **Streaming** - Chunked response streaming with throttling

### Temporal & Working Memory Services (1 service)
- **WorkingMemoryService** - Short-term memory with auto-promotion

**Total Services:** 24 (21 + 3 NLP services + 1 ONNX provider)
**Total Tests:** 295+ tests (100% passing, includes 15 NLP unit tests + 7 integration tests)
**Code Coverage:** 76.4% (application layer)
**LOC:** ~15,300 lines (12,450 + 2,850 NLP services)

---

## NLP & Analytics Services ⚡ NEW in v1.4.0

Sprint 18 introduced advanced NLP capabilities through transformer-based models (BERT, DistilBERT) with fallback mechanisms for portable deployments.

### Architecture Overview

```
┌───────────────────────────────────────────────────────────────┐
│                    NLP Services Architecture                   │
└───────────────────────────────────────────────────────────────┘
                                │
                                ▼
              ┌──────────────────────────────────┐
              │      ONNXBERTProvider            │
              │   (BERT NER + DistilBERT Sent)   │
              │   Thread-safe, BIO tokenization  │
              └──────────────────────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
                ▼               ▼               ▼
    ┌───────────────┐  ┌──────────────┐  ┌───────────────┐
    │   Enhanced    │  │  Sentiment   │  │     Topic     │
    │    Entity     │  │   Analyzer   │  │   Modeler     │
    │  Extractor    │  │ (POSITIVE/   │  │ (LDA/NMF)     │
    │ (9 types +    │  │  NEGATIVE)   │  │               │
    │ 10 relations) │  │              │  │               │
    └───────────────┘  └──────────────┘  └───────────────┘
            │                  │                  │
            └──────────────────┼──────────────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │    Fallback        │
                    │ (Rule-based/       │
                    │  Lexicon-based)    │
                    └────────────────────┘
```

### ONNXBERTProvider

**Purpose:** Unified ONNX provider for transformer-based NLP models with thread-safe inference.

**Location:** `internal/application/onnx_bert_provider.go` (641 LOC)

**Key Features:**
- **Dual Model Support:**
  - BERT NER: protectai/bert-base-NER-onnx (411 MB, 3 inputs: input_ids, attention_mask, token_type_ids)
  - DistilBERT Sentiment: lxyuan/distilbert-multilingual-sentiment (516 MB, 2 inputs: input_ids, attention_mask)
- **BIO Format Tokenization:** CoNLL-2003 standard (B-PER, I-PER, B-ORG, I-ORG, B-LOC, I-LOC, B-MISC, I-MISC, O)
- **Thread Safety:** sync.RWMutex protection for concurrent access
- **Batch Processing:** Configurable batch size (default: 16)
- **GPU Acceleration:** CUDA/ROCm support via NEXS_NLP_USE_GPU=true
- **Build Tags:** Portable builds without ONNX (noonnx tag) with stub implementation

**Interface:**
```go
type ONNXModelProvider interface {
    ExtractEntities(text string) ([]EnhancedEntity, error)
    ExtractEntitiesBatch(texts []string) ([][]EnhancedEntity, error)
    AnalyzeSentiment(text string) (*SentimentResult, error)
    IsAvailable() bool
}
```

**Performance:**
- Entity extraction: 100-200ms (CPU), 15-30ms (GPU)
- Sentiment analysis: 50-100ms (CPU), 10-20ms (GPU)
- Tokenization: 3.5µs/op
- Accuracy: 93%+ (entity), 91%+ (sentiment)

### EnhancedEntityExtractor Service

**Purpose:** Extract entities and relationships from text using transformer models with fallback.

**Location:** `internal/application/enhanced_entity_extractor.go` (432 LOC)

**Key Features:**
- **9 Entity Types:**
  - PERSON, ORGANIZATION, LOCATION, DATE
  - EVENT, PRODUCT, TECHNOLOGY, CONCEPT, OTHER
- **10 Relationship Types:**
  - WORKS_AT, FOUNDED, LOCATED_IN, BORN_IN, LIVES_IN
  - HEADQUARTERED_IN, DEVELOPED_BY, USED_BY, AFFILIATED_WITH, RELATED_TO
- **Confidence Scoring:** 0.0-1.0 with configurable threshold (default: 0.7)
- **BIO Parsing:** Multi-token entity aggregation with confidence averaging
- **Fallback Mechanism:** Rule-based regex extraction (confidence=0.5)
- **Relationship Inference:** Co-occurrence-based with evidence tracking
- **Bidirectional Storage:** Relationships stored in both directions

**Dependencies:**
- ONNXBERTProvider (primary)
- ElementRepository (storage)

**Configuration:**
```go
type EntityExtractionConfig struct {
    Enabled             bool
    ModelPath           string
    ConfidenceMin       float64  // default: 0.7
    MaxPerDocument      int      // default: 100
    EnableDisambiguation bool    // future: entity linking
}
```

### SentimentAnalyzer Service

**Purpose:** Analyze sentiment and emotional dimensions using multilingual transformer models.

**Location:** `internal/application/sentiment_analyzer.go` (418 LOC)

**Key Features:**
- **4 Sentiment Labels:** POSITIVE, NEGATIVE, NEUTRAL, MIXED (threshold: 0.6)
- **6 Emotional Dimensions:** joy, sadness, anger, fear, surprise, disgust (0.0-1.0 scores)
- **Trend Analysis:** 5-point moving average for sentiment tracking
- **Shift Detection:** Configurable threshold for emotional changes
- **Sentiment Summary:** Aggregate statistics (dominant, distribution, avg confidence)
- **Subjectivity Scoring:** 0.0-1.0 objectivity/subjectivity measure
- **Fallback Mechanism:** Lexicon-based (positive/negative word lists)

**Dependencies:**
- ONNXBERTProvider (primary)
- ElementRepository (memory storage)

**Configuration:**
```go
type SentimentConfig struct {
    Enabled    bool
    ModelPath  string
    Threshold  float64  // default: 0.6
}
```

**Performance:**
- Latency: 50-100ms (CPU), 10-20ms (GPU)
- Accuracy: 91%+ (SST-2 benchmark)
- Throughput: ~10-20 inferences/second (CPU)

### TopicModeler Service

**Purpose:** Extract topics from document collections using classical algorithms (LDA/NMF).

**Location:** `internal/application/topic_modeler.go` (653 LOC)

**Key Features:**
- **LDA Algorithm:** Latent Dirichlet Allocation with Gibbs sampling
- **NMF Algorithm:** Non-negative Matrix Factorization (faster than LDA)
- **Coherence Scoring:** Keyword co-occurrence quality metric
- **Diversity Scoring:** Keyword uniqueness metric
- **Pure Go Implementation:** No ONNX dependency (portable)
- **Configurable Parameters:** 14 params (algorithm, num_topics, iterations, alpha, beta, etc.)

**Dependencies:**
- ElementRepository (memory access)
- No ONNX dependency (classical algorithms)

**Configuration:**
```go
type TopicModelConfig struct {
    Algorithm      string  // "lda" or "nmf"
    NumTopics      int     // default: 5
    MaxIterations  int     // LDA: 100, NMF: 50
    Alpha          float64 // LDA prior (default: 0.1)
    Beta           float64 // LDA prior (default: 0.01)
    MinDocFreq     int     // default: 2
    MaxDocFreq     float64 // default: 0.95
    // ... 7 more params
}
```

**Performance:**
- LDA: 1-5s for 100 documents (CPU)
- NMF: 0.5-2s for 100 documents (CPU, faster)
- Coherence: 0.8-0.9 (quality)
- Diversity: 0.7-0.8 (uniqueness)

### Integration Points

**NLP services integrate with:**
1. **MCP Layer** - 6 NLP tools registered (`internal/mcp/nlp_tools.go`)
2. **Config System** - 14 NLP parameters with env vars (NEXS_NLP_*)
3. **Server Initialization** - ONNXBERTProvider created and injected
4. **Fallback Chain** - ONNX → Classical → Rule-based

**Build Strategy:**
- **Portable Build** (`make build`): noonnx tag, fallback only
- **Full Build** (`make build-onnx`): ONNX Runtime included, transformer models
- **Multi-Platform** (`make build-all`): Portable builds for all targets

---

## Memory Consolidation Services ⚡ v1.3.0

Sprint 14 introduced advanced memory management capabilities through 7 new services that implement state-of-the-art algorithms for duplicate detection, clustering, knowledge extraction, and quality management.

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                Memory Consolidation Workflow                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
        ┌──────────────────────────────────────┐
        │   MemoryConsolidation Orchestrator   │
        │   (workflow coordination)            │
        └──────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
          ▼                   ▼                   ▼
┌─────────────────┐  ┌─────────────────┐  ┌──────────────────┐
│ DuplicateDetect │  │   Clustering    │  │ KnowledgeGraph   │
│ (HNSW-based)    │  │ (DBSCAN/K-means)│  │ (NLP extraction) │
└─────────────────┘  └─────────────────┘  └──────────────────┘
          │                   │                   │
          └───────────────────┼───────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │ MemoryRetention  │
                    │ (Quality scoring)│
                    └──────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
          ▼                   ▼                   ▼
┌─────────────────┐  ┌─────────────────┐  ┌──────────────────┐
│  HybridSearch   │  │ SemanticSearch  │  │ ContextEnrichment│
│ (HNSW/linear)   │  │ (Multi-type)    │  │ (Relationships)  │
└─────────────────┘  └─────────────────┘  └──────────────────┘
```

### DuplicateDetection Service

**Purpose:** Detect and merge duplicate or highly similar elements using HNSW-based vector similarity.

**Structure:**
```go
type DuplicateDetectionService struct {
    repository        domain.ElementRepository
    embeddingProvider embeddings.Provider
    vectorStore       *vectorstore.VectorStore
    config            DuplicateDetectionConfig
    logger            *slog.Logger
}

type DuplicateDetectionConfig struct {
    Enabled             bool
    SimilarityThreshold float32  // 0.95 default
    MinContentLength    int      // 20 chars default
    MaxResults          int      // 100 default
}
```

**Key Features:**
- HNSW-based similarity search (O(log n) complexity)
- Configurable similarity thresholds (0.0-1.0)
- Automatic merging with metadata preservation
- Group detection (multiple duplicates of same content)
- Content length filtering

**Algorithm:**
1. Extract embeddings for all elements
2. Build HNSW index for fast similarity search
3. For each element, find neighbors within threshold
4. Group similar elements together
5. Select "best" element to keep (earliest timestamp, most metadata)
6. Optionally merge duplicates

**Example Usage:**
```go
detector := NewDuplicateDetectionService(repo, provider, config, logger)
duplicates, err := detector.DetectDuplicates(ctx, "memory", 0.95)
// Returns groups of similar elements

merged, err := detector.MergeDuplicates(ctx, duplicates, false) // dry_run=false
// Merges duplicate groups
```

**Test Coverage:** 15 tests, 100% passing

---

### Clustering Service

**Purpose:** Group related memories using DBSCAN (density-based) or K-means (centroid-based) clustering.

**Structure:**
```go
type ClusteringService struct {
    repository        domain.ElementRepository
    embeddingProvider embeddings.Provider
    vectorStore       *vectorstore.VectorStore
    config            ClusteringConfig
    logger            *slog.Logger
}

type ClusteringConfig struct {
    Enabled         bool
    Algorithm       string   // "dbscan" or "kmeans"
    MinClusterSize  int      // DBSCAN: 3 default
    EpsilonDistance float32  // DBSCAN: 0.15 default
    NumClusters     int      // K-means: 10 default
    MaxIterations   int      // K-means: 100 default
}
```

**Algorithms:**

**DBSCAN (Density-Based Spatial Clustering of Applications with Noise)**
- Discovers clusters of arbitrary shape
- Identifies outliers automatically
- Requires epsilon (neighborhood radius) and minPts parameters
- Time complexity: O(n log n) with spatial indexing

**K-means**
- Partitions data into K predefined clusters
- Minimizes within-cluster variance
- Requires number of clusters (K) parameter
- Time complexity: O(n * k * i) where i is iterations

**Key Features:**
- Two clustering algorithms for different use cases
- Automatic outlier detection (DBSCAN)
- Cluster quality metrics (silhouette score)
- Keyword extraction per cluster
- Temporal range tracking

**Example Usage:**
```go
clusterer := NewClusteringService(repo, provider, config, logger)

// DBSCAN clustering
clusters, outliers, err := clusterer.ClusterDBSCAN(ctx, "memory", 3, 0.15)

// K-means clustering
clusters, err := clusterer.ClusterKMeans(ctx, "memory", 10, 100)
```

**Test Coverage:** 13 tests, 100% passing

---

### KnowledgeGraphExtractor Service

**Purpose:** Extract entities (people, organizations, URLs, emails, concepts) and relationships from content using NLP.

**Structure:**
```go
type KnowledgeGraphExtractorService struct {
    repository domain.ElementRepository
    config     KnowledgeGraphConfig
    logger     *slog.Logger
}

type KnowledgeGraphConfig struct {
    Enabled                  bool
    ExtractPeople           bool  // true default
    ExtractOrganizations    bool  // true default
    ExtractURLs             bool  // true default
    ExtractEmails           bool  // true default
    ExtractConcepts         bool  // true default
    ExtractKeywords         bool  // true default
    MaxKeywords             int   // 10 default
    ExtractRelationships    bool  // true default
    MaxRelationships        int   // 20 default
}
```

**Entity Types:**
- **People:** Person names (John Smith, Jane Doe)
- **Organizations:** Company/org names (Acme Corp)
- **URLs:** Web links (https://github.com/example)
- **Emails:** Email addresses (john@example.com)
- **Concepts:** Key topics (machine learning, neural networks)
- **Keywords:** Important terms with TF-IDF scoring

**Relationship Types:**
- **works_on:** Person → Project
- **collaborates_with:** Person → Person
- **belongs_to:** Person → Organization
- **related_to:** Generic relationship
- **mentions:** Entity mention in context

**NLP Techniques:**
- Regular expressions for structured data (URLs, emails)
- Name entity recognition for people/orgs
- TF-IDF for keyword extraction
- Co-occurrence analysis for relationships
- Context window analysis

**Example Usage:**
```go
extractor := NewKnowledgeGraphExtractorService(repo, config, logger)
graph, err := extractor.ExtractKnowledgeGraph(ctx, elementIDs)

// Access extracted entities
for _, person := range graph.Entities.People {
    fmt.Printf("Person: %s (mentions: %d)\n", person.Name, person.Mentions)
}

// Access relationships
for _, rel := range graph.Relationships {
    fmt.Printf("%s -[%s]-> %s (strength: %.2f)\n",
        rel.FromEntity, rel.Type, rel.ToEntity, rel.Strength)
}
```

**Test Coverage:** 20 tests, 100% passing

---

### MemoryConsolidation Service

**Purpose:** Orchestrate complete consolidation workflow: duplicate detection, clustering, knowledge extraction, and quality scoring.

**Structure:**
```go
type MemoryConsolidationService struct {
    duplicateDetection *DuplicateDetectionService
    clustering         *ClusteringService
    knowledgeExtractor *KnowledgeGraphExtractorService
    qualityScorer      *MemoryRetentionService
    config             MemoryConsolidationConfig
    logger             *slog.Logger
}

type MemoryConsolidationConfig struct {
    Enabled                     bool
    AutoConsolidate            bool          // false default
    ConsolidationInterval      time.Duration // 24h default
    MinMemoriesForConsolidation int          // 10 default
    EnableDuplicateDetection   bool          // true default
    EnableClustering           bool          // true default
    EnableKnowledgeExtraction  bool          // true default
    EnableQualityScoring       bool          // true default
}
```

**Workflow Steps:**
1. **Duplicate Detection** - Find and merge similar memories
2. **Clustering** - Group related memories by topic
3. **Knowledge Extraction** - Extract entities and relationships
4. **Quality Scoring** - Calculate quality scores
5. **Recommendations** - Generate actionable insights

**Key Features:**
- End-to-end orchestration with progress tracking
- Configurable workflow steps
- Dry-run mode for preview
- Detailed recommendations
- Performance metrics per step

**Example Usage:**
```go
consolidator := NewMemoryConsolidationService(
    duplicateDetector,
    clusterer,
    extractor,
    scorer,
    config,
    logger,
)

result, err := consolidator.ConsolidateMemories(ctx, ConsolidationRequest{
    ElementType:                "memory",
    MinQuality:                 0.5,
    EnableDuplicateDetection:  true,
    EnableClustering:          true,
    EnableKnowledgeExtraction: true,
    EnableQualityScoring:      true,
    DryRun:                    false,
})

// Access results
fmt.Printf("Workflow ID: %s\n", result.WorkflowID)
fmt.Printf("Duration: %v\n", result.Duration)
fmt.Printf("Duplicates removed: %d\n", result.Summary.DuplicatesRemoved)
fmt.Printf("Clusters created: %d\n", result.Summary.NewClusters)
```

**Test Coverage:** 20 tests, 100% passing

---

### HybridSearch Service

**Purpose:** Intelligent search with automatic HNSW/linear mode selection based on dataset size.

**Structure:**
```go
type HybridSearchService struct {
    repository        domain.ElementRepository
    embeddingProvider embeddings.Provider
    hnswIndex         *hnsw.Index
    config            HybridSearchConfig
    logger            *slog.Logger
}

type HybridSearchConfig struct {
    Enabled             bool
    Mode                string   // "auto", "hnsw", "linear"
    SimilarityThreshold float32  // 0.7 default
    MaxResults          int      // 10 default
    AutoSwitchThreshold int      // 100 vectors default
    IndexPersistence    bool     // true default
    IndexPath           string   // "data/hnsw-index" default
}
```

**Search Modes:**
- **auto:** Automatically switch between HNSW and linear based on dataset size
- **hnsw:** Force HNSW mode (fast for large datasets)
- **linear:** Force linear mode (accurate for small datasets)

**Key Features:**
- HNSW approximate nearest neighbor search (O(log n))
- Linear exhaustive search fallback (O(n))
- Automatic mode switching at threshold
- Index persistence to disk
- Configurable similarity thresholds

**Performance:**
```
Dataset Size    Mode        Search Time    Accuracy
< 100          linear       ~5ms          100%
100-1000       auto         ~15ms         95%
> 1000         hnsw         ~20ms         90%
```

**Example Usage:**
```go
search := NewHybridSearchService(repo, provider, config, logger)

results, err := search.Search(ctx, SearchRequest{
    Query:               "machine learning implementation",
    ElementType:         "memory",
    Mode:                "auto",
    SimilarityThreshold: 0.7,
    MaxResults:          10,
})

// Access results
for _, result := range results.Results {
    fmt.Printf("Element: %s (similarity: %.2f)\n",
        result.Name, result.Similarity)
}
```

**Test Coverage:** 20 tests, 100% passing

---

### MemoryRetention Service

**Purpose:** Quality-based memory retention with configurable retention periods and automatic cleanup.

**Structure:**
```go
type MemoryRetentionService struct {
    repository domain.ElementRepository
    config     MemoryRetentionConfig
    logger     *slog.Logger
}

type MemoryRetentionConfig struct {
    Enabled                    bool
    QualityThreshold          float32       // 0.5 default
    HighQualityRetentionDays  int          // 365 default
    MediumQualityRetentionDays int         // 180 default
    LowQualityRetentionDays   int          // 90 default
    AutoCleanup               bool         // false default
    CleanupInterval           time.Duration // 24h default
}
```

**Quality Scoring Factors:**
1. **Content Quality (40%)**
   - Length and structure
   - Presence of title, tags, keywords
   - Formatting and completeness

2. **Recency Score (20%)**
   - Age of memory
   - Last access timestamp
   - Update frequency

3. **Relationship Score (20%)**
   - Number of relationships
   - Relationship strength
   - Cluster membership

4. **Access Score (20%)**
   - Access frequency
   - Recent access count
   - Usage patterns

**Retention Policies:**
- **High Quality (≥ 0.7):** Retain for 365 days
- **Medium Quality (0.5-0.7):** Retain for 180 days
- **Low Quality (< 0.5):** Retain for 90 days or delete

**Example Usage:**
```go
retention := NewMemoryRetentionService(repo, config, logger)

// Score memories
scores, err := retention.ScoreMemories(ctx, "memory", 0.3)

// Apply retention policy
result, err := retention.ApplyRetentionPolicy(ctx, RetentionRequest{
    QualityThreshold:          0.5,
    HighQualityRetentionDays:  365,
    MediumQualityRetentionDays: 180,
    LowQualityRetentionDays:   90,
    DryRun:                    true, // Preview first
})

// Review and apply
if result.Summary.LowQualityRemoved > 0 {
    // Apply with DryRun: false
}
```

**Test Coverage:** 15 tests, 100% passing

---

### SemanticSearch Service

**Purpose:** Semantic indexing and search across all element types with metadata filtering.

**Structure:**
```go
type SemanticSearchService struct {
    repository        domain.ElementRepository
    embeddingProvider embeddings.Provider
    indexes           map[string]*SemanticIndex
    config            SemanticSearchConfig
    logger            *slog.Logger
}

type SemanticSearchConfig struct {
    Enabled             bool
    SimilarityThreshold float32 // 0.7 default
    MaxResults          int     // 10 default
    IndexRefreshInterval time.Duration // 1h default
}
```

**Key Features:**
- Multi-type indexing (memory, agent, persona, skill)
- Metadata-aware search with filtering
- Automatic index refresh
- Incremental updates
- Tag and date range filtering

**Index Management:**
```go
// Index all elements of a type
err := search.IndexElements(ctx, "memory")

// Reindex specific element
err := search.ReindexElement(ctx, elementID)

// Search with filters
results, err := search.Search(ctx, SearchRequest{
    Query:       "project planning",
    ElementType: "memory",
    FilterTags:  []string{"sprint", "planning"},
    DateFrom:    time.Now().AddDate(0, -1, 0),
    DateTo:      time.Now(),
})
```

**Test Coverage:** 20 tests, 100% passing

---

### Configuration

All consolidation services are configurable via environment variables:

```bash
# Memory Consolidation
NEXS_MEMORY_CONSOLIDATION_ENABLED=true
NEXS_MEMORY_CONSOLIDATION_AUTO=false
NEXS_MEMORY_CONSOLIDATION_INTERVAL=24h

# Duplicate Detection
NEXS_DUPLICATE_DETECTION_ENABLED=true
NEXS_DUPLICATE_DETECTION_THRESHOLD=0.95

# Clustering
NEXS_CLUSTERING_ENABLED=true
NEXS_CLUSTERING_ALGORITHM=dbscan
NEXS_CLUSTERING_MIN_SIZE=3
NEXS_CLUSTERING_EPSILON=0.15

# Knowledge Graph
NEXS_KNOWLEDGE_GRAPH_ENABLED=true
NEXS_KNOWLEDGE_GRAPH_EXTRACT_PEOPLE=true
NEXS_KNOWLEDGE_GRAPH_EXTRACT_KEYWORDS=true

# Hybrid Search
NEXS_HYBRID_SEARCH_ENABLED=true
NEXS_HYBRID_SEARCH_MODE=auto
NEXS_HYBRID_SEARCH_THRESHOLD=0.7

# Memory Retention
NEXS_MEMORY_RETENTION_ENABLED=true
NEXS_MEMORY_RETENTION_THRESHOLD=0.5
NEXS_MEMORY_RETENTION_HIGH_DAYS=365
```

See [Configuration Reference](../api/CLI.md) for complete details.

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
