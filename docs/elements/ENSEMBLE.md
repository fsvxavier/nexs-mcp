# Ensemble Element

## Overview

Ensembles orchestrate multiple agents for complex multi-step workflows with parallel execution and result aggregation.

## Key Features

- Multi-agent coordination
- Three execution modes: sequential, parallel, hybrid
- Result aggregation strategies
- Fallback chains
- Shared context between agents

## Examples

### Code Review Ensemble
```json
{
  "name": "Comprehensive Code Review",
  "version": "1.0.0",
  "author": "qa-team",
  "members": [
    {"agent_id": "security-agent", "role": "security_reviewer", "priority": 1},
    {"agent_id": "performance-agent", "role": "performance_reviewer", "priority": 2},
    {"agent_id": "style-agent", "role": "style_checker", "priority": 3}
  ],
  "execution_mode": "parallel",
  "aggregation_strategy": "merge_all_findings",
  "fallback_chain": ["security-agent", "manual-review"]
}
```

### Research Team Ensemble
```json
{
  "name": "Research Analysis Team",
  "version": "1.0.0",
  "author": "research-lead",
  "members": [
    {"agent_id": "data-collector", "role": "collector", "priority": 1},
    {"agent_id": "data-analyzer", "role": "analyzer", "priority": 2},
    {"agent_id": "report-writer", "role": "writer", "priority": 3}
  ],
  "execution_mode": "sequential",
  "aggregation_strategy": "final_report",
  "shared_context": {
    "project": "Market Analysis Q4 2025",
    "deadline": "2025-12-31"
  }
}
```

### Multi-perspective Analysis
```json
{
  "name": "Product Decision Ensemble",
  "version": "1.0.0",
  "author": "product-team",
  "members": [
    {"agent_id": "technical-analyst", "role": "technical_review", "priority": 1},
    {"agent_id": "business-analyst", "role": "business_review", "priority": 1},
    {"agent_id": "user-experience-analyst", "role": "ux_review", "priority": 1}
  ],
  "execution_mode": "parallel",
  "aggregation_strategy": "weighted_vote",
  "fallback_chain": ["escalate_to_pm"]
}
```

## Execution Modes

- **Sequential**: Agents execute one after another, passing results forward
- **Parallel**: All agents execute simultaneously for speed
- **Hybrid**: Mix of sequential and parallel based on dependencies

## Aggregation Strategies

- **merge_all_findings**: Combine all agent outputs
- **vote**: Majority consensus from agent results
- **weighted_vote**: Weighted by agent priority
- **final_report**: Last agent output is final result

## Usage
```javascript
{
  "tool": "create_ensemble",
  "arguments": {
    "name": "Testing Ensemble",
    "members": [
      {"agent_id": "unit-tester", "role": "unit", "priority": 1},
      {"agent_id": "integration-tester", "role": "integration", "priority": 2}
    ],
    "execution_mode": "sequential",
    "aggregation_strategy": "merge_all_findings"
  }
}
```

## See Also

- [Agent Element](AGENT.md)
- [Skill Element](SKILL.md)
