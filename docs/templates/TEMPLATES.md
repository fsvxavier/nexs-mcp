# NEXS MCP Template System

The NEXS MCP Template System provides a powerful, flexible way to create reusable element templates with advanced features including variable substitution, conditional logic, loops, and 20+ helper functions.

## Table of Contents

1. [Overview](#overview)
2. [Template Structure](#template-structure)
3. [Variable System](#variable-system)
4. [Handlebars Syntax](#handlebars-syntax)
5. [Helper Functions](#helper-functions)
6. [Standard Library](#standard-library)
7. [MCP Tools](#mcp-tools)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Overview

Templates in NEXS MCP use the **Handlebars** templating language, providing:

- **Variable Substitution**: Replace placeholders with actual values
- **Conditional Logic**: Show/hide content based on conditions
- **Iteration**: Loop over arrays and objects
- **Helpers**: 20+ built-in functions for data transformation
- **Partials**: Reusable template fragments
- **Validation**: Syntax and variable checking

### Quick Example

```yaml
name: {{name}}
description: {{description}}
expertise_areas:
{{#each areas}}
  - {{this}}
{{/each}}
traits:
  primary: {{upper primary_trait}}
  {{#if secondary_trait}}
  secondary: {{secondary_trait}}
  {{/if}}
```

## Template Structure

A template is a YAML element with the following structure:

```yaml
id: my-template-id
type: template
name: My Template
description: Template description
version: 1.0.0
author: author-name
tags:
  - category-tag
  - feature-tag
content: |
  # Template content goes here
  {{variable_name}}
format: yaml  # or markdown, json, text
variables:
  - name: variable_name
    type: string  # string, number, boolean, array, object
    required: true
    default: default_value
    description: Variable description
```

### Template Formats

Supported output formats:

- **yaml**: YAML element definition (most common)
- **markdown**: Markdown documentation
- **json**: JSON configuration
- **text**: Plain text

## Variable System

### Variable Types

Templates support five variable types:

1. **string**: Text values
2. **number**: Numeric values (integers or floats)
3. **boolean**: true/false values
4. **array**: Lists of values
5. **object**: Key-value maps

### Variable Definition

```yaml
variables:
  # Required string
  - name: username
    type: string
    required: true
    description: User's display name
  
  # Optional number with default
  - name: max_connections
    type: number
    required: false
    default: "10"
    description: Maximum connections
  
  # Boolean flag
  - name: enabled
    type: boolean
    required: true
    description: Enable this feature
  
  # Array of strings
  - name: tags
    type: array
    required: false
    description: List of tags
  
  # Object/map
  - name: metadata
    type: object
    required: false
    description: Additional metadata
```

### Default Values

Variables can specify default values used when no value is provided:

```yaml
variables:
  - name: role
    type: string
    default: "user"
    
  - name: timeout
    type: number
    default: "30"
```

## Handlebars Syntax

### Basic Variables

```handlebars
Hello {{name}}!
Your email is {{email}}.
```

### Conditionals

#### Simple If

```handlebars
{{#if premium}}
You have premium access!
{{/if}}
```

#### If-Else

```handlebars
{{#if authenticated}}
Welcome back!
{{else}}
Please log in.
{{/if}}
```

#### Nested Conditionals

```handlebars
{{#if role}}
  {{#if (eq role "admin")}}
  Admin Dashboard
  {{else}}
  User Dashboard
  {{/if}}
{{/if}}
```

### Iterations

#### Each Loop

```handlebars
Skills:
{{#each skills}}
  - {{this}}
{{/each}}
```

#### Loop with Index

```handlebars
{{#each items}}
{{@index}}. {{this}}
{{/each}}
```

#### Loop over Objects

```handlebars
{{#each user}}
{{@key}}: {{this}}
{{/each}}
```

#### Nested Loops

```handlebars
{{#each categories}}
Category: {{name}}
  {{#each items}}
  - {{this}}
  {{/each}}
{{/each}}
```

### Partials

Partials are reusable template fragments:

```handlebars
{{> header}}

Main content here

{{> footer}}
```

### Comments

```handlebars
{{!-- This is a comment --}}
{{! This is also a comment }}
```

## Helper Functions

NEXS MCP provides 20+ helper functions organized by category.

### String Helpers

#### upper
Convert to uppercase:
```handlebars
{{upper name}}
<!-- "alice" → "ALICE" -->
```

#### lower
Convert to lowercase:
```handlebars
{{lower name}}
<!-- "ALICE" → "alice" -->
```

#### title
Title case (first letter uppercase):
```handlebars
{{title name}}
<!-- "alice smith" → "Alice Smith" -->
```

#### trim
Remove leading/trailing whitespace:
```handlebars
{{trim description}}
```

#### replace
Replace substring:
```handlebars
{{replace text "old" "new"}}
<!-- "old value" → "new value" -->
```

### Comparison Helpers

#### eq
Equal to:
```handlebars
{{#if (eq status "active")}}
Active!
{{/if}}
```

#### ne
Not equal to:
```handlebars
{{#if (ne role "guest")}}
Authenticated user
{{/if}}
```

#### gt / gte
Greater than / Greater than or equal:
```handlebars
{{#if (gt age 18)}}
Adult
{{/if}}

{{#if (gte count 10)}}
Ten or more
{{/if}}
```

#### lt / lte
Less than / Less than or equal:
```handlebars
{{#if (lt remaining 5)}}
Running low
{{/if}}

{{#if (lte score 0)}}
No points
{{/if}}
```

### Logical Helpers

#### and
Logical AND:
```handlebars
{{#if (and enabled verified)}}
Fully activated
{{/if}}
```

#### or
Logical OR:
```handlebars
{{#if (or isAdmin isModerator)}}
Has permissions
{{/if}}
```

#### not
Logical NOT:
```handlebars
{{#if (not disabled)}}
Enabled
{{/if}}
```

### Formatting Helpers

#### json
Format as JSON:
```handlebars
{{{json object}}}
<!-- {"key": "value"} -->
```

#### default
Provide fallback value:
```handlebars
{{default name "Anonymous"}}
<!-- Shows "Anonymous" if name is empty -->
```

### Array Helpers

#### join
Join array with separator:
```handlebars
{{join tags ", "}}
<!-- ["go", "mcp"] → "go, mcp" -->
```

#### length
Array/string length:
```handlebars
Count: {{length items}}
<!-- [1,2,3] → "Count: 3" -->
```

### Complete Helper Reference

| Helper | Category | Description | Example |
|--------|----------|-------------|---------|
| `upper` | String | Convert to uppercase | `{{upper "hello"}}` → `HELLO` |
| `lower` | String | Convert to lowercase | `{{lower "WORLD"}}` → `world` |
| `title` | String | Title case | `{{title "john doe"}}` → `John Doe` |
| `trim` | String | Remove whitespace | `{{trim " text "}}` → `text` |
| `replace` | String | Replace substring | `{{replace text "a" "b"}}` |
| `eq` | Comparison | Equal to | `{{#if (eq x 5)}}` |
| `ne` | Comparison | Not equal | `{{#if (ne x 0)}}` |
| `gt` | Comparison | Greater than | `{{#if (gt x 10)}}` |
| `gte` | Comparison | Greater/equal | `{{#if (gte x 5)}}` |
| `lt` | Comparison | Less than | `{{#if (lt x 100)}}` |
| `lte` | Comparison | Less/equal | `{{#if (lte x 50)}}` |
| `and` | Logical | Logical AND | `{{#if (and a b)}}` |
| `or` | Logical | Logical OR | `{{#if (or a b)}}` |
| `not` | Logical | Logical NOT | `{{#if (not flag)}}` |
| `json` | Format | Format as JSON | `{{{json obj}}}` |
| `default` | Format | Default value | `{{default x "N/A"}}` |
| `join` | Array | Join array | `{{join arr ", "}}` |
| `length` | Array | Array length | `{{length arr}}` |

## Standard Library

NEXS MCP includes 30+ built-in templates across all element types.

### Persona Templates

- `stdlib-persona-assistant`: General AI assistant
- `stdlib-persona-analyst`: Data analyst
- `stdlib-persona-developer`: Software developer
- `stdlib-persona-writer`: Content writer
- `stdlib-persona-researcher`: Research specialist

### Skill Templates

- `stdlib-skill-api-call`: HTTP API integration
- `stdlib-skill-data-transform`: Data transformation
- `stdlib-skill-validation`: Input validation
- `stdlib-skill-search`: Search operation
- `stdlib-skill-aggregation`: Data aggregation

### Agent Templates

- `stdlib-agent-simple`: Basic agent
- `stdlib-agent-multi-step`: Multi-step workflow
- `stdlib-agent-decision-tree`: Decision-based agent
- `stdlib-agent-reactive`: Event-driven agent

### Memory Templates

- `stdlib-memory-conversation`: Conversation memory
- `stdlib-memory-knowledge`: Knowledge base entry
- `stdlib-memory-context`: Context snapshot
- `stdlib-memory-task`: Task record

### Ensemble Templates

- `stdlib-ensemble-pipeline`: Linear pipeline
- `stdlib-ensemble-parallel`: Parallel execution
- `stdlib-ensemble-hierarchical`: Hierarchical structure

### Template Templates

- `stdlib-template-basic`: Basic template structure
- `stdlib-template-advanced`: Advanced features
- `stdlib-template-documentation`: Documentation template

## MCP Tools

### list_templates

List available templates with filtering:

```json
{
  "tool": "list_templates",
  "arguments": {
    "category": "persona",
    "tags": ["ai", "assistant"],
    "element_type": "persona",
    "include_builtin": true,
    "page": 1,
    "per_page": 20
  }
}
```

**Returns:**
- Array of template summaries
- Total count
- Pagination info

### get_template

Retrieve template details:

```json
{
  "tool": "get_template",
  "arguments": {
    "id": "stdlib-persona-assistant"
  }
}
```

**Returns:**
- Complete template definition
- Variable specifications
- Available helpers list

### instantiate_template

Create element from template:

```json
{
  "tool": "instantiate_template",
  "arguments": {
    "template_id": "stdlib-persona-assistant",
    "variables": {
      "name": "Customer Support Bot",
      "expertise": "customer service",
      "tone": "friendly and helpful"
    },
    "save_as": "my-support-bot",
    "dry_run": false
  }
}
```

**Returns:**
- Instantiated content
- Warnings (if any)
- Element ID (if saved)
- Used helpers list

### validate_template

Validate template syntax:

```json
{
  "tool": "validate_template",
  "arguments": {
    "template_id": "my-custom-template",
    "variables": {
      "test_var": "test_value"
    }
  }
}
```

**Returns:**
- Validation status (valid/invalid)
- Error messages
- Warning messages
- Fix suggestions

## Best Practices

### 1. Use Descriptive Variable Names

**Good:**
```yaml
variables:
  - name: user_full_name
  - name: max_retry_count
  - name: enable_logging
```

**Avoid:**
```yaml
variables:
  - name: x
  - name: val
  - name: flag
```

### 2. Always Provide Descriptions

```yaml
variables:
  - name: timeout_seconds
    type: number
    description: Request timeout in seconds (1-300)
    default: "30"
```

### 3. Set Sensible Defaults

```yaml
variables:
  - name: log_level
    type: string
    default: "info"
    description: Logging level (debug, info, warn, error)
```

### 4. Mark Required Variables

```yaml
variables:
  - name: api_key
    type: string
    required: true
    description: API authentication key (required)
```

### 5. Validate Before Instantiate

Always validate templates before using them in production:

```javascript
// 1. Validate syntax
validate_template(template_id)

// 2. Test with sample variables
instantiate_template(template_id, test_vars, dry_run=true)

// 3. Use in production
instantiate_template(template_id, prod_vars, save_as=new_id)
```

### 6. Use Conditionals for Optional Sections

```handlebars
{{#if description}}
description: {{description}}
{{/if}}

{{#if tags}}
tags:
{{#each tags}}
  - {{this}}
{{/each}}
{{/if}}
```

### 7. Leverage Helpers for Consistency

```handlebars
# Always uppercase status
status: {{upper status}}

# Trim whitespace from user input
name: {{trim name}}

# Provide defaults for optional fields
role: {{default role "user"}}
```

### 8. Keep Templates Focused

- One template = One purpose
- Break complex templates into partials
- Use standard library as building blocks

### 9. Version Your Templates

```yaml
id: my-template-v2
version: 2.0.0
description: Updated template with new features
```

### 10. Test Edge Cases

Test templates with:
- Empty variables
- Maximum length values
- Special characters
- Missing optional variables
- All combinations of conditionals

## Troubleshooting

### Unbalanced Delimiters

**Error:** `template syntax invalid: unbalanced delimiters`

**Cause:** Mismatched `{{` and `}}`

**Fix:**
```handlebars
<!-- Wrong -->
{{name}

<!-- Correct -->
{{name}}
```

### Undefined Variables

**Warning:** `undefined variable: xyz`

**Cause:** Variable used in template but not declared

**Fix:**
```yaml
# Add to variables section
variables:
  - name: xyz
    type: string
```

### Missing Required Variables

**Error:** `required variable 'name' not provided`

**Cause:** Required variable not provided during instantiation

**Fix:**
```json
{
  "variables": {
    "name": "value"  // Provide all required variables
  }
}
```

### Type Mismatch

**Error:** `variable 'count' expects number, got string`

**Cause:** Wrong variable type provided

**Fix:**
```json
{
  "variables": {
    "count": 42  // Use number, not "42"
  }
}
```

### Helper Not Found

**Error:** `unknown helper: xyz`

**Cause:** Helper doesn't exist or typo

**Fix:**
- Check [Helper Reference](#complete-helper-reference)
- Verify spelling: `{{upper name}}` not `{{uppercase name}}`

### Template Too Large

**Error:** `template exceeds maximum size (1MB)`

**Cause:** Template content > 1MB

**Fix:**
- Split into multiple templates
- Use partials for shared content
- Remove unnecessary whitespace

### Too Many Variables

**Error:** `template exceeds variable limit (100)`

**Cause:** More than 100 variables declared

**Fix:**
- Use objects to group related variables
- Split template into multiple templates
- Review if all variables are necessary

### Conditional Not Working

**Problem:** `{{#if variable}}` always false

**Cause:** Variable is string "false" not boolean

**Fix:**
```handlebars
<!-- Wrong -->
{{#if "false"}}  <!-- String is truthy -->

<!-- Correct -->
{{#if (eq enabled true)}}
{{#if enabled}}  <!-- Boolean variable -->
```

### Loop Not Iterating

**Problem:** `{{#each items}}` produces no output

**Cause:** Variable is not an array

**Fix:**
```yaml
# Ensure variable type is array
variables:
  - name: items
    type: array
```

```json
// Provide array value
{
  "items": ["a", "b", "c"]
}
```

## Advanced Examples

### Persona with Conditional Expertise

```yaml
name: {{name}}
description: {{description}}
expertise_areas:
{{#each expertise}}
  - {{this}}
{{/each}}
traits:
  communication_style: {{default style "professional"}}
  {{#if tone}}
  tone: {{tone}}
  {{/if}}
  {{#if (gt experience_years 5)}}
  seniority: senior
  {{else}}
  seniority: junior
  {{/if}}
response_patterns:
  {{#if (and formal concise)}}
  - formal_brief
  {{/if}}
  {{#if (or creative analytical)}}
  - balanced_approach
  {{/if}}
```

### Skill with Dynamic Triggers

```yaml
name: {{name}}
description: {{description}}
triggers:
{{#each trigger_patterns}}
  - pattern: "{{this}}"
    type: {{default trigger_type "keyword"}}
{{/each}}
parameters:
  {{#each parameters}}
  {{@key}}: {{json this}}
  {{/each}}
```

### Agent with Decision Tree

```yaml
name: {{name}}
goals:
{{#each goals}}
  - {{this}}
{{/each}}
decision_tree:
  {{#if (eq strategy "sequential")}}
  type: sequential
  {{else if (eq strategy "parallel")}}
  type: parallel
  {{else}}
  type: adaptive
  {{/if}}
  max_depth: {{default max_depth "5"}}
actions:
{{#each actions}}
  - id: {{id}}
    type: {{type}}
    {{#if condition}}
    condition: {{condition}}
    {{/if}}
{{/each}}
```

---

**Version:** 1.0.0  
**Last Updated:** M0.9 Release  
**Related:** See [ADR-009](../adr/ADR-009-element-template-system.md) for architecture details
