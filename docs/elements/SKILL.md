# Skill Element

## Overview

A **Skill** represents a procedural capability that can be triggered by specific conditions and executed as a series of steps. Skills are composable and can depend on other skills or tools.

## Key Features

- **Trigger-based Activation**: Keyword, pattern, context, or manual triggers
- **Step-by-Step Procedures**: Defined sequence of actions
- **Tool Integration**: Can invoke external tools
- **Composable**: Skills can depend on other skills
- **Reusable**: Same skill can be used in multiple contexts

## Schema

```json
{
  "name": "string (3-100 chars)",
  "description": "string (max 500 chars)",
  "version": "semver string",
  "author": "string",
  "tags": ["array", "of", "strings"],
  "triggers": [
    {
      "type": "keyword|pattern|context|manual",
      "keywords": ["array", "for", "keyword", "type"],
      "pattern": "regex pattern for pattern type",
      "context": "context description for context type"
    }
  ],
  "procedures": [
    {
      "step": "integer",
      "action": "action description",
      "description": "optional details",
      "tool_name": "optional tool to invoke",
      "parameters": {"key": "value"}
    }
  ],
  "dependencies": [
    {
      "skill_id": "uuid",
      "required": true
    }
  ]
}
```

## Examples

### 1. Code Review Skill

```json
{
  "name": "Comprehensive Code Review",
  "description": "Performs thorough code review with security, performance, and style checks",
  "version": "1.0.0",
  "author": "dev-team",
  "tags": ["code-review", "quality", "security"],
  "triggers": [
    {
      "type": "keyword",
      "keywords": ["review this code", "code review", "check this code"]
    },
    {
      "type": "pattern",
      "pattern": "review.*\\.\\w+$"
    }
  ],
  "procedures": [
    {
      "step": 1,
      "action": "Analyze code structure and organization",
      "description": "Check for proper separation of concerns and modularity"
    },
    {
      "step": 2,
      "action": "Review for security vulnerabilities",
      "description": "Check for SQL injection, XSS, auth issues, etc.",
      "tool_name": "security_scanner"
    },
    {
      "step": 3,
      "action": "Evaluate performance implications",
      "description": "Identify potential bottlenecks and optimization opportunities"
    },
    {
      "step": 4,
      "action": "Check code style and best practices",
      "description": "Verify adherence to coding standards",
      "tool_name": "linter"
    },
    {
      "step": 5,
      "action": "Provide actionable feedback",
      "description": "Summarize findings with specific recommendations"
    }
  ]
}
```

### 2. Data Pipeline Skill

```json
{
  "name": "ETL Data Pipeline",
  "description": "Extract, transform, and load data from multiple sources",
  "version": "2.0.0",
  "author": "data-team",
  "tags": ["etl", "data", "pipeline"],
  "triggers": [
    {
      "type": "manual"
    }
  ],
  "procedures": [
    {
      "step": 1,
      "action": "Extract data from sources",
      "description": "Connect to databases, APIs, and files",
      "tool_name": "data_extractor",
      "parameters": {"sources": ["db1", "api1", "file1"]}
    },
    {
      "step": 2,
      "action": "Validate data quality",
      "description": "Check for nulls, duplicates, and format issues",
      "tool_name": "data_validator"
    },
    {
      "step": 3,
      "action": "Transform data",
      "description": "Apply business rules and data transformations",
      "tool_name": "data_transformer",
      "parameters": {"rules": "transformation_rules.yaml"}
    },
    {
      "step": 4,
      "action": "Load to destination",
      "description": "Write transformed data to target systems",
      "tool_name": "data_loader",
      "parameters": {"target": "warehouse"}
    }
  ],
  "dependencies": [
    {
      "skill_id": "data-validation-skill-uuid",
      "required": true
    }
  ]
}
```

### 3. Research Summarization Skill

```json
{
  "name": "Academic Paper Summarizer",
  "description": "Extracts key insights from academic papers",
  "version": "1.1.0",
  "author": "research-team",
  "tags": ["research", "summarization", "academic"],
  "triggers": [
    {
      "type": "keyword",
      "keywords": ["summarize paper", "paper summary", "extract insights"]
    },
    {
      "type": "context",
      "context": "user provides academic paper URL or text"
    }
  ],
  "procedures": [
    {
      "step": 1,
      "action": "Extract paper content",
      "description": "Download and parse PDF or fetch from URL"
    },
    {
      "step": 2,
      "action": "Identify key sections",
      "description": "Find abstract, methodology, results, conclusion"
    },
    {
      "step": 3,
      "action": "Extract main findings",
      "description": "Identify core contributions and results"
    },
    {
      "step": 4,
      "action": "Summarize methodology",
      "description": "Describe approach and methods used"
    },
    {
      "step": 5,
      "action": "Generate structured summary",
      "description": "Create formatted summary with key points",
      "tool_name": "template_renderer",
      "parameters": {"template": "paper-summary-template"}
    }
  ]
}
```

## Usage with MCP

```javascript
{
  "tool": "create_skill",
  "arguments": {
    "name": "Bug Triage Skill",
    "description": "Analyzes and prioritizes bug reports",
    "version": "1.0.0",
    "author": "qa-team",
    "triggers": [
      {
        "type": "keyword",
        "keywords": ["triage bug", "prioritize issue"]
      }
    ],
    "procedures": [
      {
        "step": 1,
        "action": "Extract bug details",
        "description": "Parse bug report for symptoms and impact"
      },
      {
        "step": 2,
        "action": "Assess severity",
        "description": "Determine critical/high/medium/low severity"
      },
      {
        "step": 3,
        "action": "Assign priority",
        "description": "Set P0/P1/P2/P3 based on severity and impact"
      }
    ]
  }
}
```

## Best Practices

1. **Trigger Design**: Use multiple trigger types for flexibility
2. **Procedural Steps**: Keep steps atomic and well-defined
3. **Tool Integration**: Specify tool names for external integrations
4. **Dependencies**: Document skill dependencies clearly
5. **Versioning**: Update version when changing procedures

## See Also

- [Agent Element](AGENT.md) - Uses skills for task execution
- [Template Element](TEMPLATE.md) - Can be used within skill procedures
