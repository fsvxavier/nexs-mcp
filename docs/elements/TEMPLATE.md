# Template Element

## Overview

Templates enable variable substitution and multi-format content generation. Support Markdown, YAML, JSON, and plain text.

## Key Features

- Variable substitution with `{{variable}}` syntax
- Multiple format support
- Required/optional variables with defaults
- Validation rules

## Examples

### Email Template
```json
{
  "name": "Welcome Email",
  "version": "1.0.0",
  "author": "marketing",
  "content": "Hello {{name}},\n\nWelcome to {{product}}! Your account {{account_id}} is ready.\n\nBest regards,\n{{sender}}",
  "format": "markdown",
  "variables": [
    {"name": "name", "type": "string", "required": true},
    {"name": "product", "type": "string", "required": true},
    {"name": "account_id", "type": "string", "required": true},
    {"name": "sender", "type": "string", "required": false, "default": "Support Team"}
  ]
}
```

### API Response Template
```json
{
  "name": "Error Response",
  "version": "1.0.0",
  "author": "api-team",
  "content": "{\"error\": \"{{error_type}}\", \"message\": \"{{message}}\", \"code\": {{code}}}",
  "format": "json",
  "variables": [
    {"name": "error_type", "type": "string", "required": true},
    {"name": "message", "type": "string", "required": true},
    {"name": "code", "type": "number", "required": true}
  ]
}
```

## Usage
```javascript
// Render template
template.render({
  "name": "John",
  "product": "NEXS MCP",
  "account_id": "ACC-12345"
})
```
