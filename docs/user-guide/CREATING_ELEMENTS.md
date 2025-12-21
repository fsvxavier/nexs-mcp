# Creating Elements - Step-by-Step Guide

This guide walks you through creating each type of element in NEXS-MCP with practical examples and best practices.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Personas](#creating-personas)
3. [Skills](#creating-skills)
4. [Templates](#creating-templates)
5. [Agents](#creating-agents)
6. [Memories](#creating-memories)
7. [Ensembles](#creating-ensembles)
8. [Best Practices](#best-practices)
9. [Validation](#validation)
10. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Using Quick Create Tools

NEXS-MCP provides `quick_create_*` tools for rapid element creation:

```bash
# Via Claude Desktop or MCP client
quick_create_persona {
  "name": "my-assistant",
  "description": "A helpful AI assistant",
  "traits": ["helpful", "creative", "analytical"]
}
```

### Manual YAML Creation

For more control, create YAML files in `data/elements/<type>/`:

```yaml
# data/elements/personas/my-assistant.yaml
name: my-assistant
type: persona
version: "1.0.0"
description: A helpful AI assistant
traits:
  - helpful
  - creative
  - analytical
```

---

## Creating Personas

### What is a Persona?

A Persona defines the character, tone, and behavior of an AI assistant. It includes personality traits, communication style, and behavioral guidelines.

### Step 1: Choose Your Persona Type

Common persona types:
- **Professional**: Business analyst, technical architect, project manager
- **Creative**: Writer, designer, content creator
- **Educational**: Teacher, tutor, explainer
- **Specialized**: Domain expert (medical, legal, financial)

### Step 2: Define Core Attributes

```yaml
name: technical-architect
type: persona
version: "1.0.0"
author: your-name
description: >
  Expert system architect with 15+ years of experience in 
  distributed systems, microservices, and cloud architecture.

# Personality traits (3-7 recommended)
traits:
  - analytical
  - detail-oriented
  - pragmatic
  - experienced
  - systematic

# Communication style
communication_style:
  tone: professional
  formality: formal
  verbosity: detailed
  technical_level: expert
```

### Step 3: Add Behavioral Guidelines

```yaml
# Behavioral patterns
behavior:
  - "Always consider scalability and maintainability"
  - "Provide architectural trade-offs and alternatives"
  - "Use industry-standard terminology"
  - "Reference established patterns and practices"
  - "Ask clarifying questions when requirements are ambiguous"

# Expertise areas
expertise:
  - microservices architecture
  - cloud platforms (AWS, Azure, GCP)
  - distributed systems
  - API design
  - system scalability
```

### Step 4: Define Constraints (Optional)

```yaml
# What the persona should avoid
constraints:
  - "Avoid over-engineering solutions"
  - "Don't recommend proprietary solutions without justification"
  - "Never compromise security for convenience"

# Response format preferences
response_format:
  structure: "markdown"
  include_diagrams: true
  code_examples: true
```

### Step 5: Add Metadata and Tags

```yaml
metadata:
  category: professional
  industry: technology
  experience_level: senior
  language: english

tags:
  - architecture
  - cloud
  - distributed-systems
  - enterprise
```

### Complete Persona Example

```yaml
name: senior-data-scientist
type: persona
version: "1.0.0"
author: nexs-team
description: >
  Senior data scientist with expertise in machine learning, 
  statistical analysis, and data visualization.

traits:
  - analytical
  - detail-oriented
  - curious
  - methodical
  - data-driven

communication_style:
  tone: professional
  formality: balanced
  verbosity: comprehensive
  technical_level: expert

behavior:
  - "Start with data exploration and understanding"
  - "Explain statistical concepts clearly"
  - "Validate assumptions before modeling"
  - "Consider both statistical and practical significance"
  - "Provide visualizations to support findings"

expertise:
  - machine learning algorithms
  - statistical analysis
  - data visualization
  - feature engineering
  - model evaluation

constraints:
  - "Never p-hack or cherry-pick results"
  - "Always check for data leakage"
  - "Consider ethical implications of models"

response_format:
  structure: "markdown"
  include_visualizations: true
  code_examples: true
  statistical_tests: true

metadata:
  category: professional
  industry: data-science
  experience_level: senior
  language: english

tags:
  - machine-learning
  - statistics
  - data-analysis
  - visualization
  - python
```

### Using the Persona

```bash
# Create via quick_create
quick_create_persona {
  "name": "senior-data-scientist",
  "description": "Expert data scientist",
  "traits": ["analytical", "detail-oriented", "curious"],
  "expertise": ["machine-learning", "statistics", "python"]
}

# Or manually save YAML to data/elements/personas/
# Then reload:
reload_elements {"types": ["persona"]}
```

---

## Creating Skills

### What is a Skill?

A Skill is a specific capability or task that can be performed. Skills are modular, reusable, and can be combined in Agents and Ensembles.

### Step 1: Identify the Skill Purpose

Questions to ask:
- What specific task does this skill perform?
- What inputs does it need?
- What outputs does it produce?
- Are there any prerequisites?

### Step 2: Define Skill Structure

```yaml
name: code-review-expert
type: skill
version: "1.0.0"
author: nexs-team
description: >
  Expert code review skill that analyzes code quality, security, 
  performance, and best practices.

# Core capability
capability: code_review

# Input requirements
inputs:
  - name: code
    type: string
    required: true
    description: "Source code to review"
  
  - name: language
    type: string
    required: true
    description: "Programming language (e.g., python, javascript)"
  
  - name: context
    type: string
    required: false
    description: "Additional context about the code"
```

### Step 3: Define Implementation

```yaml
# How the skill works
implementation:
  type: analysis
  method: multi-pass
  
  steps:
    - name: syntax_check
      description: "Verify code syntax"
      
    - name: security_scan
      description: "Check for security vulnerabilities"
      
    - name: performance_analysis
      description: "Identify performance bottlenecks"
      
    - name: best_practices
      description: "Evaluate adherence to best practices"
```

### Step 4: Define Output Format

```yaml
# What the skill returns
outputs:
  - name: review_report
    type: object
    description: "Comprehensive code review report"
    schema:
      issues: array
      recommendations: array
      score: number
      summary: string

# Quality metrics
quality_metrics:
  - security_score
  - performance_score
  - maintainability_score
  - overall_score
```

### Complete Skill Example

```yaml
name: api-design-validator
type: skill
version: "1.0.0"
author: nexs-team
description: >
  Validates REST API designs for best practices, consistency,
  and RESTful principles.

capability: api_validation

inputs:
  - name: api_spec
    type: string
    required: true
    description: "OpenAPI/Swagger specification"
  
  - name: standards
    type: array
    required: false
    description: "Specific standards to validate against"

implementation:
  type: validation
  method: rule-based
  
  rules:
    - check_http_methods
    - validate_resource_naming
    - verify_status_codes
    - check_authentication
    - validate_response_formats
    - check_pagination
    - verify_versioning

  steps:
    - name: parse_spec
      description: "Parse OpenAPI specification"
      
    - name: apply_rules
      description: "Apply validation rules"
      
    - name: generate_report
      description: "Create validation report"

outputs:
  - name: validation_result
    type: object
    description: "API validation results"
    schema:
      valid: boolean
      issues: array
      warnings: array
      suggestions: array
      compliance_score: number

quality_metrics:
  - restful_compliance
  - naming_consistency
  - documentation_quality
  - security_score

dependencies:
  - openapi-parser
  - rest-validator

metadata:
  category: validation
  domain: api-design
  complexity: medium

tags:
  - rest-api
  - openapi
  - validation
  - best-practices
```

---

## Creating Templates

### What is a Template?

Templates are reusable text structures with variables that can be filled in. They support Handlebars syntax for dynamic content generation.

### Step 1: Define Template Purpose

```yaml
name: technical-report
type: template
version: "1.0.0"
author: nexs-team
description: "Professional technical report template"

format: markdown
category: documentation
```

### Step 2: Define Variables

```yaml
# Variables that users will provide
variables:
  - name: title
    type: string
    required: true
    description: "Report title"
    
  - name: author
    type: string
    required: true
    description: "Report author"
    
  - name: date
    type: string
    required: false
    default: "{{currentDate}}"
    description: "Report date"
    
  - name: sections
    type: array
    required: true
    description: "Report sections with content"
```

### Step 3: Create Template Content

```yaml
content: |
  # {{title}}
  
  **Author:** {{author}}  
  **Date:** {{date}}
  
  ---
  
  ## Executive Summary
  
  {{executive_summary}}
  
  ---
  
  ## Table of Contents
  
  {{#each sections}}
  - [{{this.title}}](#{{this.anchor}})
  {{/each}}
  
  ---
  
  {{#each sections}}
  ## {{this.title}}
  
  {{this.content}}
  
  {{/each}}
  
  ---
  
  ## Conclusion
  
  {{conclusion}}
  
  {{#if appendices}}
  ---
  
  ## Appendices
  
  {{#each appendices}}
  ### {{this.title}}
  
  {{this.content}}
  
  {{/each}}
  {{/if}}
```

### Complete Template Example

```yaml
name: meeting-minutes
type: template
version: "1.0.0"
author: nexs-team
description: "Structured meeting minutes template"

format: markdown
category: documentation

variables:
  - name: meeting_title
    type: string
    required: true
    
  - name: date
    type: string
    required: true
    
  - name: attendees
    type: array
    required: true
    
  - name: agenda_items
    type: array
    required: true
    
  - name: action_items
    type: array
    required: false
    
  - name: next_meeting
    type: string
    required: false

content: |
  # {{meeting_title}}
  
  **Date:** {{date}}  
  **Duration:** {{duration}}
  
  ## Attendees
  
  {{#each attendees}}
  - {{this}}
  {{/each}}
  
  ## Agenda
  
  {{#each agenda_items}}
  ### {{@index}}. {{this.topic}}
  
  **Presenter:** {{this.presenter}}
  
  {{this.discussion}}
  
  **Decision:** {{this.decision}}
  
  {{/each}}
  
  ## Action Items
  
  {{#each action_items}}
  - [ ] **{{this.task}}** - Assigned to: {{this.owner}} - Due: {{this.due_date}}
  {{/each}}
  
  {{#if next_meeting}}
  ## Next Meeting
  
  **Scheduled for:** {{next_meeting}}
  {{/if}}

metadata:
  category: meeting
  format: markdown
  use_case: team_collaboration

tags:
  - meetings
  - minutes
  - documentation
  - team
```

---

## Creating Agents

### What is an Agent?

An Agent combines a Persona (who) with Skills (what) to create an autonomous actor that can perform tasks.

### Step 1: Design Agent Purpose

```yaml
name: ci-automation-agent
type: agent
version: "1.0.0"
author: nexs-team
description: >
  Automated CI/CD agent that monitors builds, runs tests,
  and manages deployment pipelines.
```

### Step 2: Assign Persona and Skills

```yaml
# Who is this agent?
persona_id: devops-engineer

# What can this agent do?
skills:
  - skill_id: ci-monitor
    priority: 1
    
  - skill_id: test-runner
    priority: 2
    
  - skill_id: deploy-manager
    priority: 3
```

### Step 3: Configure Behavior

```yaml
# How does the agent operate?
behavior:
  mode: autonomous
  trigger: on_commit
  
  rules:
    - "Run tests on every commit"
    - "Block deployments if tests fail"
    - "Notify team on build failures"
    - "Auto-deploy to staging on success"

# Decision making
decision_strategy: rule-based

# Interaction preferences
interaction:
  notifications: enabled
  approval_required: false
  escalation_on_failure: true
```

### Complete Agent Example

```yaml
name: security-audit-agent
type: agent
version: "1.0.0"
author: nexs-team
description: >
  Autonomous security agent that performs continuous security
  audits, vulnerability scans, and compliance checks.

persona_id: security-specialist

skills:
  - skill_id: vulnerability-scanner
    priority: 1
    config:
      scan_depth: comprehensive
      
  - skill_id: dependency-checker
    priority: 2
    config:
      check_licenses: true
      
  - skill_id: code-security-analyzer
    priority: 3
    
  - skill_id: compliance-validator
    priority: 4

behavior:
  mode: scheduled
  schedule: "0 2 * * *"  # Daily at 2 AM
  
  rules:
    - "Scan all repositories daily"
    - "Flag critical vulnerabilities immediately"
    - "Generate weekly compliance reports"
    - "Auto-create tickets for findings"

decision_strategy: risk-based

thresholds:
  critical_severity: immediate_action
  high_severity: notify_within_24h
  medium_severity: weekly_report
  low_severity: monthly_report

interaction:
  notifications: enabled
  channels:
    - email
    - slack
  approval_required: false
  escalation_on_failure: true

outputs:
  - security_report
  - vulnerability_list
  - compliance_status

metadata:
  category: security
  automation_level: high
  criticality: high

tags:
  - security
  - automation
  - compliance
  - vulnerabilities
```

---

## Creating Memories

### What is a Memory?

Memories store contextual information, learned patterns, and historical data that can be referenced by agents and ensembles.

### Step 1: Define Memory Type

Types of memories:
- **Context Memory**: Current conversation/project context
- **Knowledge Memory**: Learned facts and patterns
- **Historical Memory**: Past interactions and outcomes
- **Preference Memory**: User preferences and settings

```yaml
name: project-context
type: memory
version: "1.0.0"
author: nexs-team
description: "Current project context and requirements"

memory_type: context
scope: project
persistence: session
```

### Step 2: Structure Memory Data

```yaml
# What information is stored
structure:
  project_info:
    - project_name
    - description
    - timeline
    - stakeholders
    
  technical_stack:
    - languages
    - frameworks
    - tools
    - infrastructure
    
  requirements:
    - functional
    - non_functional
    - constraints
```

### Step 3: Configure Access and Retention

```yaml
# Who can access this memory
access:
  visibility: private
  shared_with: []
  
# How long to keep the memory
retention:
  duration: 30d
  auto_cleanup: true
  archive_after: 90d

# How to update the memory
update_strategy: append
conflict_resolution: merge
```

### Complete Memory Example

```yaml
name: conversation-history
type: memory
version: "1.0.0"
author: nexs-team
description: >
  Maintains conversation history with context, decisions,
  and learned preferences.

memory_type: historical
scope: user
persistence: permanent

structure:
  conversations:
    - timestamp
    - topic
    - summary
    - key_points
    - decisions_made
    
  learned_preferences:
    - communication_style
    - preferred_formats
    - topics_of_interest
    - expertise_areas
    
  interaction_patterns:
    - common_questions
    - typical_workflows
    - preferred_tools

access:
  visibility: private
  shared_with: []
  encryption: enabled

retention:
  duration: 365d
  auto_cleanup: true
  archive_after: 730d
  max_entries: 1000

update_strategy: append
conflict_resolution: timestamp
compression: enabled

indexing:
  searchable_fields:
    - topic
    - key_points
    - decisions_made
  full_text_search: enabled

metadata:
  category: historical
  sensitivity: medium
  backup_enabled: true

tags:
  - conversation
  - history
  - context
  - user-preferences
```

---

## Creating Ensembles

### What is an Ensemble?

An Ensemble coordinates multiple agents working together to solve complex problems through collaboration.

### Step 1: Design Ensemble Strategy

```yaml
name: code-review-team
type: ensemble
version: "1.0.0"
author: nexs-team
description: >
  Collaborative code review ensemble with security,
  performance, and quality specialists.

execution_mode: parallel
aggregation_strategy: voting
```

### Step 2: Add Team Members

```yaml
# Who is in the ensemble?
members:
  - agent_id: security-reviewer
    role: security_analyst
    priority: 1
    weight: 1.5  # Security has higher weight
    
  - agent_id: performance-reviewer
    role: performance_analyst
    priority: 1
    weight: 1.0
    
  - agent_id: quality-reviewer
    role: quality_analyst
    priority: 1
    weight: 1.0
    
  - agent_id: style-reviewer
    role: style_checker
    priority: 2
    weight: 0.5  # Style has lower weight
```

### Step 3: Configure Collaboration

```yaml
# How do members work together?
collaboration:
  communication: enabled
  shared_context: true
  conflict_resolution: weighted_voting
  
# Decision making
decision:
  threshold: 0.7  # 70% agreement needed
  fallback_chain:
    - senior-reviewer
    - tech-lead
```

### Complete Ensemble Example

```yaml
name: research-analysis-team
type: ensemble
version: "1.0.0"
author: nexs-team
description: >
  Research ensemble that analyzes topics from multiple perspectives:
  technical, business, and user experience.

execution_mode: sequential
aggregation_strategy: merge

members:
  - agent_id: technical-analyst
    role: technical_research
    priority: 1
    weight: 1.0
    config:
      focus: technical_feasibility
      depth: comprehensive
    
  - agent_id: business-analyst
    role: business_research
    priority: 2
    weight: 1.0
    config:
      focus: market_viability
      depth: comprehensive
    
  - agent_id: ux-researcher
    role: user_research
    priority: 3
    weight: 0.8
    config:
      focus: user_needs
      depth: detailed
    
  - agent_id: synthesizer
    role: synthesis
    priority: 4
    weight: 1.5
    config:
      combine_perspectives: true
      generate_recommendations: true

collaboration:
  communication: enabled
  shared_context: true
  context_passing: sequential
  conflict_resolution: consensus

decision:
  threshold: 0.75
  require_unanimous: false
  fallback_chain:
    - synthesizer
    - research-lead

workflow:
  steps:
    - phase: research
      parallel: true
      members: [technical-analyst, business-analyst, ux-researcher]
      
    - phase: synthesis
      parallel: false
      members: [synthesizer]
      inputs: [research_results]
      
  timeout: 30m
  retry_on_failure: 2

outputs:
  - research_report
  - recommendations
  - risk_analysis
  - implementation_plan

metadata:
  category: research
  complexity: high
  use_case: strategic_planning

tags:
  - research
  - analysis
  - collaboration
  - multi-perspective
```

---

## Best Practices

### General Guidelines

1. **Naming Conventions**
   - Use lowercase with hyphens: `my-element-name`
   - Be descriptive but concise
   - Include type indicator when helpful: `api-validator-skill`

2. **Version Management**
   - Follow semantic versioning: `MAJOR.MINOR.PATCH`
   - Increment MAJOR for breaking changes
   - Increment MINOR for new features
   - Increment PATCH for bug fixes

3. **Documentation**
   - Write clear, comprehensive descriptions
   - Document all inputs and outputs
   - Provide usage examples
   - Include prerequisites and dependencies

4. **Metadata and Tags**
   - Use consistent tag taxonomy
   - Add relevant category information
   - Include search-friendly keywords
   - Specify author and version

### Element-Specific Tips

**Personas:**
- Keep traits focused (3-7 traits ideal)
- Make communication style explicit
- Define clear behavioral guidelines
- Specify expertise areas precisely

**Skills:**
- Single responsibility principle
- Clear input/output contracts
- Proper error handling
- Well-defined prerequisites

**Templates:**
- Use semantic variable names
- Provide sensible defaults
- Include usage examples
- Document Handlebars syntax

**Agents:**
- Match persona to purpose
- Choose relevant skills
- Configure appropriate behavior
- Set realistic thresholds

**Memories:**
- Define clear structure
- Set appropriate retention
- Configure proper access control
- Enable search when needed

**Ensembles:**
- Choose appropriate execution mode
- Balance member weights thoughtfully
- Set reasonable decision thresholds
- Define clear fallback chains

---

## Validation

### Validate Before Using

Always validate elements before using them:

```bash
# Validate single element
validate_element {
  "element_id": "my-persona",
  "type": "persona",
  "strict": true
}

# Validate all elements of a type
validate_element {
  "type": "skill",
  "validate_all": true
}
```

### Common Validation Issues

1. **Missing Required Fields**
   ```yaml
   # ❌ Invalid - missing type
   name: my-persona
   description: "A persona"
   
   # ✅ Valid
   name: my-persona
   type: persona
   description: "A persona"
   ```

2. **Invalid Version Format**
   ```yaml
   # ❌ Invalid
   version: "1.0"
   
   # ✅ Valid
   version: "1.0.0"
   ```

3. **Wrong Variable Types**
   ```yaml
   # ❌ Invalid - traits should be array
   traits: "helpful"
   
   # ✅ Valid
   traits:
     - helpful
     - creative
   ```

---

## Troubleshooting

### Element Not Loading

**Problem:** Element doesn't appear after creation

**Solutions:**
1. Check file location: `data/elements/<type>/`
2. Verify YAML syntax: `yamllint my-file.yaml`
3. Reload elements: `reload_elements {"types": ["persona"]}`
4. Check logs for errors

### Validation Failures

**Problem:** Element fails validation

**Solutions:**
1. Run strict validation: `validate_element` with `strict: true`
2. Check error messages carefully
3. Compare with working examples
4. Verify all required fields present

### Template Not Rendering

**Problem:** Template variables not substituting

**Solutions:**
1. Check variable names match exactly
2. Verify Handlebars syntax: `{{variable}}`
3. Ensure all required variables provided
4. Test with simple template first

### Ensemble Not Executing

**Problem:** Ensemble members not running

**Solutions:**
1. Verify all member agent IDs exist
2. Check execution mode configuration
3. Verify required skills are available
4. Check logs for member failures

---

## Next Steps

- [Quick Start Guide](QUICK_START.md) - Get started quickly
- [MCP Tools Reference](../api/MCP_TOOLS.md) - All available tools
- [Element Types Reference](../elements/README.md) - Detailed element specs
- [Examples](../../examples/) - Real-world examples

---

**Need Help?**

- Check [Troubleshooting Guide](TROUBLESHOOTING.md)
- Read [Getting Started](GETTING_STARTED.md)
- See [Examples](../../examples/basic/)
- Open an [Issue](https://github.com/fsvxavier/nexs-mcp/issues)
