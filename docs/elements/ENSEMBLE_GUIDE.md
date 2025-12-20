# Ensemble Execution Guide

## Table of Contents
- [Introduction](#introduction)
- [What is an Ensemble?](#what-is-an-ensemble)
- [Execution Modes](#execution-modes)
- [Aggregation Strategies](#aggregation-strategies)
- [Real-time Monitoring](#real-time-monitoring)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Introduction

Ensembles are powerful constructs that allow you to orchestrate multiple AI agents to work together on complex tasks. Think of an ensemble as a team where each agent has a specific role, and the team's combined output is greater than the sum of its parts.

## What is an Ensemble?

An **Ensemble** is a coordinated group of agents that work together to solve problems. Each ensemble has:

- **Members**: Individual agents with specific roles
- **Execution Mode**: How agents are coordinated (sequential, parallel, or hybrid)
- **Aggregation Strategy**: How results are combined (consensus, voting, merge, etc.)
- **Shared Context**: Information passed between agents
- **Fallback Chain**: Backup plans if agents fail

### When to Use Ensembles

Use ensembles when you need to:
- **Collaborate**: Multiple perspectives on the same problem
- **Divide and Conquer**: Break complex tasks into parallel subtasks
- **Verify**: Cross-check results from multiple sources
- **Enhance Quality**: Combine multiple approaches for better results

## Execution Modes

### 1. Sequential Execution

Agents execute one after another, with each agent building on the previous agent's results.

**Use when:**
- Order matters (e.g., plan → execute → review)
- Later agents need results from earlier agents
- You want incremental refinement

**Example:**
```json
{
  "execution_mode": "sequential",
  "members": [
    {"agent_id": "planner", "role": "planning", "priority": 10},
    {"agent_id": "executor", "role": "execution", "priority": 9},
    {"agent_id": "reviewer", "role": "review", "priority": 8}
  ]
}
```

**Workflow:**
```
Input → Planner → Executor → Reviewer → Final Result
          ↓          ↓           ↓
      (plan)    (implementation) (quality check)
```

### 2. Parallel Execution

All agents execute simultaneously and independently.

**Use when:**
- Tasks are independent
- Speed is critical
- You want diverse perspectives
- No inter-agent dependencies

**Example:**
```json
{
  "execution_mode": "parallel",
  "members": [
    {"agent_id": "analyst_1", "role": "data_analysis", "priority": 5},
    {"agent_id": "analyst_2", "role": "data_analysis", "priority": 5},
    {"agent_id": "analyst_3", "role": "data_analysis", "priority": 5}
  ]
}
```

**Workflow:**
```
           ┌─→ Analyst 1 ─┐
Input ─────┼─→ Analyst 2 ─┼──→ Aggregation → Final Result
           └─→ Analyst 3 ─┘
```

### 3. Hybrid Execution

Combines sequential and parallel execution using priority groups. Agents with the same priority execute in parallel, but different priority levels execute sequentially.

**Use when:**
- You have phases with parallel subtasks
- Some tasks depend on others, but some are independent
- You want optimal resource utilization

**Example:**
```json
{
  "execution_mode": "hybrid",
  "members": [
    {"agent_id": "researcher_1", "role": "research", "priority": 10},
    {"agent_id": "researcher_2", "role": "research", "priority": 10},
    {"agent_id": "synthesizer", "role": "synthesis", "priority": 9},
    {"agent_id": "writer_1", "role": "writing", "priority": 8},
    {"agent_id": "writer_2", "role": "writing", "priority": 8}
  ]
}
```

**Workflow:**
```
Priority 10: Research1 & Research2 (parallel)
                    ↓
Priority 9:  Synthesizer (sequential)
                    ↓
Priority 8:  Writer1 & Writer2 (parallel)
```

## Aggregation Strategies

### 1. First

Returns the result from the first successful agent.

**Use for:**
- Quick answers where first response is good enough
- Racing multiple approaches (fastest wins)

```json
{"aggregation_strategy": "first"}
```

### 2. Last

Returns the result from the last successful agent.

**Use for:**
- Sequential refinement where final result is best
- Iterative improvement workflows

```json
{"aggregation_strategy": "last"}
```

### 3. All

Returns all successful results as an array.

**Use for:**
- Comparing different approaches
- Presenting multiple options to users
- Collecting diverse perspectives

```json
{"aggregation_strategy": "all"}
```

### 4. Consensus

Uses advanced consensus algorithm to find the most agreed-upon result.

**Features:**
- Agreement threshold (default 70%)
- Weighted voting by agent priority
- Quorum requirements
- Alternative results if no consensus

**Use for:**
- Critical decisions requiring agreement
- Reducing noise from outlier results
- Democratic decision-making

```json
{
  "aggregation_strategy": "consensus",
  "config": {
    "threshold": 0.7,
    "weighted_voting": true,
    "require_quorum": true,
    "quorum_size": 3
  }
}
```

**Result structure:**
```json
{
  "value": "agreed_result",
  "agreement_level": 0.85,
  "participants": 5,
  "supporting": ["agent-1", "agent-2", "agent-3", "agent-4"],
  "reached_consensus": true,
  "alternative": {
    "other_option": "minority_result"
  }
}
```

### 5. Voting

Implements weighted voting with multiple tie-breaking strategies.

**Features:**
- Priority-based vote weights
- Confidence-based vote weights
- Custom weight per agent
- Tie-breaker strategies (first, highest_priority, random)

**Use for:**
- Choosing between discrete options
- Leveraging agent expertise (weights)
- Democratic decisions with expert input

```json
{
  "aggregation_strategy": "voting",
  "config": {
    "weight_by_priority": true,
    "weight_by_confidence": true,
    "minimum_votes": 3,
    "tie_breaker": "highest_priority",
    "custom_weights": {
      "expert_agent": 2.0,
      "junior_agent": 0.5
    }
  }
}
```

**Result structure:**
```json
{
  "winner": "option_a",
  "total_votes": 10.0,
  "winner_votes": 6.5,
  "percentage": 65.0,
  "voters": ["agent-1", "agent-2", "agent-3"],
  "breakdown": {
    "option_a": 6.5,
    "option_b": 3.5
  },
  "tie_breaker": false
}
```

### 6. Weighted Consensus

Consensus with confidence scores from agents.

**Use for:**
- Decisions where confidence matters
- Quality-weighted results
- Expert systems with uncertainty

```json
{"aggregation_strategy": "weighted_consensus"}
```

### 7. Threshold Consensus

Requires minimum agreement threshold with quorum.

**Use for:**
- High-stakes decisions requiring strong agreement
- Ensuring minimum participation
- Critical validations

```json
{
  "aggregation_strategy": "threshold_consensus",
  "threshold": 0.8,
  "quorum": 5
}
```

### 8. Merge

Combines all results into a single map.

**Use for:**
- Collecting complementary information
- Parallel data gathering
- Comprehensive reports

```json
{"aggregation_strategy": "merge"}
```

## Real-time Monitoring

Enable real-time monitoring to track long-running ensemble executions.

### Enabling Monitoring

```json
{
  "ensemble_id": "my-ensemble",
  "options": {
    "enable_monitoring": true,
    "timeout": 300
  }
}
```

### Monitor Data

When monitoring is enabled, the execution result includes:

```json
{
  "metadata": {
    "monitoring": {
      "execution_id": "exec-my-ensemble-1703001234",
      "ensemble_id": "my-ensemble",
      "status": "running",
      "phase": "execution-parallel",
      "total_agents": 5,
      "completed_agents": 3,
      "failed_agents": 0,
      "progress": 0.6,
      "elapsed_time": "45s",
      "estimated_remaining": "30s",
      "agent_progress": {
        "agent-1": {
          "status": "completed",
          "progress": 1.0,
          "role": "analyzer"
        },
        "agent-2": {
          "status": "running",
          "progress": 0.7,
          "role": "processor"
        }
      }
    }
  }
}
```

### Progress Callbacks

Register callbacks to receive real-time updates:

```go
monitor.RegisterProgressCallback(func(m *ExecutionMonitor) {
    update := m.GetProgressUpdate()
    fmt.Printf("Progress: %.1f%% (%d/%d agents)\n", 
        update.Progress*100, 
        update.CompletedAgents, 
        update.TotalAgents)
})
```

## Best Practices

### 1. Choose the Right Execution Mode

- **Sequential** for dependent tasks
- **Parallel** for independent tasks
- **Hybrid** for complex multi-phase workflows

### 2. Set Appropriate Priorities

- Higher priority (10) = executes first in hybrid mode
- Use priorities to control execution order
- Group related tasks at same priority

### 3. Configure Timeouts

```json
{
  "options": {
    "timeout": 300,  // 5 minutes
    "max_retries": 2,
    "fail_fast": false
  }
}
```

### 4. Use Fail-Fast Wisely

```json
{
  "options": {
    "fail_fast": true  // Stop on first error
  }
}
```

Enable `fail_fast` when:
- Early failure invalidates everything
- You want quick feedback
- Errors are unrecoverable

Disable `fail_fast` when:
- Partial results are valuable
- Some failures are acceptable
- You want comprehensive results

### 5. Leverage Shared Context

```json
{
  "shared_context": {
    "project_name": "MyProject",
    "requirements": ["fast", "accurate"],
    "constraints": {
      "max_tokens": 1000
    }
  }
}
```

Shared context allows agents to:
- Access common configuration
- Share discoveries
- Maintain consistency

### 6. Design Effective Fallback Chains

```json
{
  "fallback_chain": [
    {"agent_id": "primary_agent", "condition": "timeout"},
    {"agent_id": "backup_agent", "condition": "error"},
    {"agent_id": "simple_agent", "condition": "any"}
  ]
}
```

## Examples

See the `examples/ensembles/` directory for complete examples:

- **simple_sequential.yaml**: Basic sequential workflow
- **parallel_analysis.yaml**: Parallel data analysis
- **hybrid_research.yaml**: Multi-phase research project
- **consensus_voting.yaml**: Decision-making with consensus
- **code_review.yaml**: Code review ensemble
- **content_generation.yaml**: Content creation pipeline

## Troubleshooting

### No Results Returned

**Problem**: Ensemble completes but no aggregated result.

**Solutions**:
- Check that at least one agent succeeds
- Verify aggregation strategy supports your result types
- Enable monitoring to see which agents fail

### Timeout Issues

**Problem**: Ensemble times out before completion.

**Solutions**:
- Increase timeout value
- Use parallel execution for independent tasks
- Optimize individual agent performance
- Enable monitoring to identify slow agents

### Consensus Not Reached

**Problem**: Consensus strategy returns no result.

**Solutions**:
- Lower threshold (e.g., 0.5 instead of 0.7)
- Increase agent count for better agreement
- Check that agents produce comparable results
- Review agent implementations for consistency

### Agent Failures

**Problem**: Multiple agents failing.

**Solutions**:
- Check agent configurations
- Verify input format
- Review error messages in results
- Test agents individually first
- Enable retries (`max_retries: 2`)

### Performance Issues

**Problem**: Ensemble is slow.

**Solutions**:
- Use parallel execution when possible
- Optimize agent performance
- Set reasonable timeouts
- Use hybrid mode for large ensembles
- Monitor agent-by-agent progress

## Advanced Topics

### Custom Aggregation

For complex aggregation logic, consider implementing custom aggregation strategies in your application layer.

### Dynamic Ensembles

Ensembles can be created programmatically:

```go
ensemble := &domain.Ensemble{
    ID: "dynamic-ensemble",
    Name: "Dynamic Team",
    ExecutionMode: "parallel",
    AggregationStrategy: "voting",
    Members: buildMembersFromContext(ctx),
}
```

### Nested Ensembles

Agents in an ensemble can themselves trigger other ensembles for hierarchical processing.

### Real-world Use Cases

1. **Code Review**: Sequential analysis → parallel reviewers → consensus
2. **Research**: Parallel research → synthesis → parallel writing
3. **Decision Making**: Parallel analysis → voting → validation
4. **Content Creation**: Planning → parallel generation → selection
5. **Data Processing**: Parallel extraction → merge → validation

---

For more information, see:
- [Ensemble Element Documentation](./ENSEMBLE.md)
- [Agent Documentation](./AGENT.md)
- [API Reference](../api/MCP_TOOLS.md)
