# Ensemble Examples

This directory contains practical examples of ensemble configurations for various use cases.

## Available Examples

### 1. Simple Sequential (`simple_sequential.yaml`)
**Complexity**: Beginner  
**Execution Mode**: Sequential  
**Use Case**: Basic document processing pipeline

A straightforward example showing how to chain agents sequentially where each agent builds on the previous one's work. Perfect for learning the basics.

**Workflow**: Analyzer → Processor → Refiner

### 2. Parallel Analysis (`parallel_analysis.yaml`)
**Complexity**: Intermediate  
**Execution Mode**: Parallel  
**Use Case**: Multi-perspective data analysis

Demonstrates parallel execution with multiple analysts providing diverse perspectives on the same dataset. Uses consensus aggregation to find agreed-upon insights.

**Key Features**:
- Weighted voting by confidence
- Quorum requirements
- Consensus threshold
- Multiple analytical approaches

### 3. Hybrid Research (`hybrid_research.yaml`)
**Complexity**: Advanced  
**Execution Mode**: Hybrid  
**Use Case**: Comprehensive research and writing

A complex multi-phase workflow combining parallel and sequential execution. Shows how to orchestrate a complete research project from data gathering to final report.

**Phases**:
1. **Research** (parallel): 3 researchers gather information
2. **Synthesis** (sequential): 1 synthesizer consolidates findings
3. **Writing** (parallel): 3 writers create content
4. **Review** (sequential): 1 editor polishes the result

### 4. Code Review (`code_review.yaml`)
**Complexity**: Intermediate  
**Execution Mode**: Parallel  
**Use Case**: Automated code review

Demonstrates weighted voting with specialized reviewers. Each reviewer has expertise in a specific area (security, performance, quality) and their votes are weighted according to importance.

**Reviewers**:
- Security (weight: 2.0) - Critical
- Code Quality (weight: 1.5) - Important  
- Performance (weight: 1.2) - Significant
- Testing (weight: 1.0) - Necessary
- Documentation (weight: 0.8) - Nice to have

## Using These Examples

### Via MCP Tool

```json
{
  "tool": "create_ensemble",
  "input": {
    "id": "my-ensemble",
    "name": "My Custom Ensemble",
    "execution_mode": "parallel",
    "aggregation_strategy": "consensus",
    "members": [
      // ... copy from examples
    ]
  }
}
```

### Via CLI

```bash
# Load an example
nexs-mcp load examples/ensembles/parallel_analysis.yaml

# Execute with custom input
nexs-mcp execute-ensemble parallel-data-analysis \
  --input '{"dataset": "my_data.csv"}'
```

### Customizing Examples

1. **Copy** the example file
2. **Modify** the configuration:
   - Change execution mode
   - Adjust priorities
   - Add/remove members
   - Update aggregation strategy
3. **Test** with your data
4. **Iterate** based on results

## Configuration Guide

### Execution Modes

```yaml
execution_mode: sequential  # One after another
execution_mode: parallel    # All at once
execution_mode: hybrid      # Priority-based groups
```

### Aggregation Strategies

```yaml
aggregation_strategy: first      # First result
aggregation_strategy: last       # Last result
aggregation_strategy: all        # All results
aggregation_strategy: consensus  # Agreement-based
aggregation_strategy: voting     # Vote-based
aggregation_strategy: merge      # Combine all
```

### Execution Options

```yaml
execution_options:
  timeout: 300           # Max execution time (seconds)
  max_retries: 2         # Retry failed agents
  fail_fast: false       # Stop on first error
  enable_monitoring: true # Track progress
```

### Consensus Configuration

```yaml
consensus_config:
  threshold: 0.7         # 70% agreement required
  weighted_voting: true  # Weight by priority/confidence
  require_quorum: true   # Minimum participants
  quorum_size: 3         # Minimum 3 agents
```

### Voting Configuration

```yaml
voting_config:
  weight_by_priority: true      # Use agent priority
  weight_by_confidence: true    # Use confidence scores
  minimum_votes: 3              # Minimum votes required
  tie_breaker: highest_priority # How to break ties
  custom_weights:               # Custom weights per agent
    expert-agent: 2.0
    junior-agent: 0.5
```

## Best Practices

### 1. Start Simple
Begin with `simple_sequential.yaml` to understand the basics before moving to complex examples.

### 2. Match Mode to Task
- **Sequential**: When order matters
- **Parallel**: When speed matters
- **Hybrid**: When both matter

### 3. Choose Right Aggregation
- **Consensus**: For agreement on important decisions
- **Voting**: For choosing between options
- **Last**: For iterative refinement
- **All**: For comprehensive results

### 4. Set Realistic Timeouts
- Simple tasks: 30-60 seconds
- Medium tasks: 2-5 minutes
- Complex tasks: 5-10 minutes

### 5. Use Monitoring
Always enable monitoring for long-running ensembles to track progress.

### 6. Handle Failures Gracefully
- Set `fail_fast: false` for partial results
- Configure fallback chains
- Enable retries for transient failures

## Creating Your Own Ensemble

### Step 1: Define Purpose
What problem are you solving? What outcome do you need?

### Step 2: Choose Agents
Which agents (roles) do you need? What expertise is required?

### Step 3: Select Execution Mode
Should agents work sequentially, in parallel, or in phases?

### Step 4: Configure Aggregation
How should results be combined? Consensus? Voting? Last result?

### Step 5: Test and Iterate
Run with test data, monitor execution, refine configuration.

## Troubleshooting

### Problem: Agents Not Executing
**Check**: Agent IDs are correct and agents exist

### Problem: No Consensus Reached
**Solution**: Lower threshold or increase agent count

### Problem: Slow Execution
**Solution**: Use parallel mode or optimize agents

### Problem: Inconsistent Results
**Solution**: Improve agent implementations or shared context

## Further Reading

- [Ensemble Execution Guide](../../docs/elements/ENSEMBLE_GUIDE.md) - Complete guide
- [Ensemble Documentation](../../docs/elements/ENSEMBLE.md) - Reference
- [Agent Documentation](../../docs/elements/AGENT.md) - Agent creation
- [API Reference](../../docs/api/MCP_TOOLS.md) - MCP tools

## Contributing Examples

Have a great ensemble example? Submit a PR with:
1. YAML configuration file
2. Description in this README
3. Example input/output
4. Use case explanation

---

**Last Updated**: December 20, 2025  
**Version**: 1.0.0
