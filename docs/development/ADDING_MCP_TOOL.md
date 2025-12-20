# Adding a New MCP Tool: Complete Tutorial

## Table of Contents

- [Introduction](#introduction)
- [Understanding MCP Tools](#understanding-mcp-tools)
- [Step 1: Define Input/Output Schemas](#step-1-define-inputoutput-schemas)
- [Step 2: Create Handler Function](#step-2-create-handler-function)
- [Step 3: Register with SDK](#step-3-register-with-sdk)
- [Step 4: Implement Business Logic](#step-4-implement-business-logic)
- [Step 5: Error Handling](#step-5-error-handling)
- [Step 6: Add Metrics Tracking](#step-6-add-metrics-tracking)
- [Step 7: Write Tests](#step-7-write-tests)
- [Step 8: Update Documentation](#step-8-update-documentation)
- [Complete Example: validate_template Tool](#complete-example-validate_template-tool)
- [Best Practices](#best-practices)
- [Performance Considerations](#performance-considerations)
- [Common Patterns](#common-patterns)
- [Troubleshooting](#troubleshooting)

---

## Introduction

This tutorial teaches you how to create a new MCP tool for NEXS-MCP using the official MCP Go SDK. We'll build a complete working example: a `validate_template` tool that validates template syntax and variables.

**Prerequisites:**
- Understanding of MCP protocol
- Familiarity with Go
- Knowledge of NEXS-MCP architecture (see [CODE_TOUR.md](./CODE_TOUR.md))
- Official MCP Go SDK (`github.com/modelcontextprotocol/go-sdk/mcp`)

**Time Required:** 1-2 hours

---

## Understanding MCP Tools

### What is an MCP Tool?

An MCP tool is a function that:
1. Accepts structured input (JSON)
2. Performs an operation
3. Returns structured output (JSON)

**Communication Flow:**
```
┌─────────────┐                           ┌──────────────┐
│ MCP Client  │  JSON-RPC over stdio      │  MCP Server  │
│  (Claude)   │ ───────────────────────>  │   (NEXS)     │
│             │                           │              │
│             │  <────────────────────────│              │
│             │      JSON Response        │              │
└─────────────┘                           └──────────────┘
```

### Tool Structure (Official SDK)

Using the official MCP Go SDK:

```go
import sdk "github.com/modelcontextprotocol/go-sdk/mcp"

// 1. Define the tool metadata
tool := &sdk.Tool{
    Name:        "my_tool",
    Description: "What the tool does",
    // InputSchema is auto-generated from handler parameter type
}

// 2. Create handler function
handler := func(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    // Implementation
}

// 3. Register with server
sdk.AddTool(server, tool, handler)
```

### Tool Categories in NEXS-MCP

1. **CRUD Tools**: Create, read, update, delete elements
2. **Search Tools**: Find elements by criteria
3. **Validation Tools**: Validate element structure/content
4. **Execution Tools**: Execute templates, ensembles, agents
5. **Management Tools**: Backup, restore, statistics
6. **Integration Tools**: GitHub, collections, portfolio

---

## Step 1: Define Input/Output Schemas

Create clear, well-documented schemas using Go structs with JSON schema tags.

### Input Schema

**Location:** `internal/mcp/tools.go` or new file `internal/mcp/validation_tools.go`

```go
package mcp

// ValidateTemplateInput defines input for validate_template tool
type ValidateTemplateInput struct {
    // Template content to validate
    Content string `json:"content" jsonschema:"required,template content to validate (Handlebars syntax)"`
    
    // Template type (optional, defaults to handlebars)
    TemplateType string `json:"template_type,omitempty" jsonschema:"template type (handlebars, jinja2, go-template) default: handlebars"`
    
    // Variables to check against
    Variables []TemplateVariable `json:"variables,omitempty" jsonschema:"expected variables for validation"`
    
    // Validation level (optional)
    ValidationLevel string `json:"validation_level,omitempty" jsonschema:"validation depth (basic, comprehensive, strict) default: comprehensive"`
    
    // Sample data for test rendering (optional)
    SampleData map[string]interface{} `json:"sample_data,omitempty" jsonschema:"sample data to test rendering"`
}

// TemplateVariable represents an expected variable
type TemplateVariable struct {
    Name        string `json:"name" jsonschema:"required,variable name"`
    Type        string `json:"type" jsonschema:"variable type (string, number, boolean, array, object)"`
    Required    bool   `json:"required" jsonschema:"whether variable is required"`
    Description string `json:"description,omitempty" jsonschema:"variable description"`
}
```

### Output Schema

```go
// ValidateTemplateOutput defines output for validate_template tool
type ValidateTemplateOutput struct {
    // Overall validation result
    IsValid bool `json:"is_valid" jsonschema:"whether template is valid"`
    
    // Validation errors
    Errors []ValidationIssue `json:"errors" jsonschema:"validation errors"`
    
    // Validation warnings
    Warnings []ValidationIssue `json:"warnings" jsonschema:"validation warnings"`
    
    // Informational messages
    Infos []ValidationIssue `json:"infos,omitempty" jsonschema:"informational messages"`
    
    // Detected variables
    DetectedVariables []string `json:"detected_variables" jsonschema:"variables found in template"`
    
    // Missing variables (expected but not found)
    MissingVariables []string `json:"missing_variables,omitempty" jsonschema:"variables expected but not found"`
    
    // Unknown variables (found but not expected)
    UnknownVariables []string `json:"unknown_variables,omitempty" jsonschema:"variables found but not expected"`
    
    // Test render result (if sample data provided)
    TestRenderResult string `json:"test_render_result,omitempty" jsonschema:"result of test rendering"`
    
    // Validation time in milliseconds
    ValidationTime int64 `json:"validation_time_ms" jsonschema:"validation duration in milliseconds"`
    
    // Suggestions for improvement
    Suggestions []string `json:"suggestions,omitempty" jsonschema:"suggestions for improvement"`
}

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
    Severity   string `json:"severity" jsonschema:"issue severity (error, warning, info)"`
    Line       int    `json:"line,omitempty" jsonschema:"line number where issue occurs"`
    Column     int    `json:"column,omitempty" jsonschema:"column number where issue occurs"`
    Message    string `json:"message" jsonschema:"issue description"`
    Code       string `json:"code" jsonschema:"issue code (e.g., SYNTAX_ERROR)"`
    Suggestion string `json:"suggestion,omitempty" jsonschema:"suggested fix"`
}
```

**Schema Design Principles:**

1. **Required vs Optional**: Use `omitempty` for optional fields
2. **Defaults**: Document in jsonschema description
3. **Validation Hints**: Use jsonschema tags to guide clients
4. **Clear Names**: Use descriptive field names
5. **Rich Errors**: Include line numbers, suggestions, codes

---

## Step 2: Create Handler Function

Create the handler function that implements the tool's logic.

**Location:** `internal/mcp/validation_tools.go` (new file)

```go
package mcp

import (
    "context"
    "fmt"
    "regexp"
    "strings"
    "time"

    sdk "github.com/modelcontextprotocol/go-sdk/mcp"

    "github.com/aymerick/raymond"
    "github.com/fsvxavier/nexs-mcp/internal/logger"
    "github.com/fsvxavier/nexs-mcp/internal/validation"
)

// handleValidateTemplate handles the validate_template tool
func (s *MCPServer) handleValidateTemplate(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    // Step 1: Start timing for metrics
    start := time.Now()
    var err error
    
    // Step 2: Defer metrics recording
    defer func() {
        duration := time.Since(start)
        s.metrics.RecordOperation("validate_template", duration, err == nil)
        s.perfMetrics.Record(logger.PerformanceEntry{
            Timestamp: start,
            Operation: "validate_template",
            Duration:  duration,
            Success:   err == nil,
            Metadata: map[string]interface{}{
                "template_length": len(input.Content),
                "validation_level": input.ValidationLevel,
            },
        })
    }()

    // Step 3: Parse and validate input
    var input ValidateTemplateInput
    if err = parseArguments(arguments, &input); err != nil {
        logger.Error("Failed to parse arguments", "error", err)
        return errorResult("Invalid arguments: %v", err), nil
    }

    // Step 4: Validate input
    if input.Content == "" {
        err = fmt.Errorf("content is required")
        return errorResult("Template content is required"), nil
    }

    // Step 5: Set defaults
    if input.TemplateType == "" {
        input.TemplateType = "handlebars"
    }
    if input.ValidationLevel == "" {
        input.ValidationLevel = "comprehensive"
    }

    // Step 6: Log operation
    logger.Info("Validating template",
        "type", input.TemplateType,
        "level", input.ValidationLevel,
        "content_length", len(input.Content),
        "has_sample_data", input.SampleData != nil)

    // Step 7: Perform validation (delegate to business logic)
    result := s.validateTemplateContent(input)

    // Step 8: Return success response
    return successResult(result)
}
```

**Handler Function Signature:**

```go
func (s *MCPServer) handleToolName(
    ctx context.Context,        // Context for cancellation, timeouts
    arguments interface{},      // Raw arguments from client
) (*sdk.CallToolResult, error) // Result or error
```

**Key Points:**

1. **Context**: Always accept `context.Context` for cancellation
2. **Arguments**: SDK passes as `interface{}`, you parse to your input type
3. **Return Type**: Always return `*sdk.CallToolResult` and `error`
4. **Error Handling**: Return errors via `errorResult()`, not Go error
5. **Logging**: Log important operations and errors

---

## Step 3: Register with SDK

Register the tool with the MCP server using `sdk.AddTool()`.

**Location:** `internal/mcp/server.go` in `registerTools()` method

```go
func (s *MCPServer) registerTools() {
    // ... existing tool registrations ...

    // Validation tools
    sdk.AddTool(s.server, &sdk.Tool{
        Name: "validate_template",
        Description: `Validates template syntax and structure.

Supports multiple template types (Handlebars, Jinja2, Go templates) and 
performs comprehensive validation including:
- Syntax checking
- Variable detection
- Type validation
- Best practices
- Optional test rendering with sample data

Use this before saving a template to ensure it's valid and follows best practices.`,
    }, s.handleValidateTemplate)

    // ... more tool registrations ...
}
```

**Registration Best Practices:**

1. **Clear Name**: Use snake_case, descriptive names
2. **Rich Description**: Explain what it does, when to use it, what it validates
3. **Multi-line Descriptions**: Use backticks for better formatting
4. **Handler Reference**: Pass method reference `s.handleValidateTemplate`

**The SDK automatically:**
- Generates JSON Schema from input struct
- Validates input against schema
- Marshals/unmarshals JSON
- Handles JSON-RPC protocol details

---

## Step 4: Implement Business Logic

Separate business logic from handler for better testing and reusability.

```go
// validateTemplateContent performs the actual template validation
func (s *MCPServer) validateTemplateContent(input ValidateTemplateInput) ValidateTemplateOutput {
    start := time.Now()
    
    output := ValidateTemplateOutput{
        IsValid:           true,
        Errors:            []ValidationIssue{},
        Warnings:          []ValidationIssue{},
        Infos:             []ValidationIssue{},
        DetectedVariables: []string{},
        MissingVariables:  []string{},
        UnknownVariables:  []string{},
        Suggestions:       []string{},
    }
    
    defer func() {
        output.ValidationTime = time.Since(start).Milliseconds()
    }()

    // Validate based on template type
    switch input.TemplateType {
    case "handlebars":
        s.validateHandlebarsTemplate(input, &output)
    case "jinja2":
        s.validateJinja2Template(input, &output)
    case "go-template":
        s.validateGoTemplate(input, &output)
    default:
        output.IsValid = false
        output.Errors = append(output.Errors, ValidationIssue{
            Severity: "error",
            Message:  fmt.Sprintf("Unsupported template type: %s", input.TemplateType),
            Code:     "UNSUPPORTED_TYPE",
        })
        return output
    }

    // Detect variables in template
    output.DetectedVariables = s.detectTemplateVariables(input.Content, input.TemplateType)

    // Check expected vs detected variables
    if len(input.Variables) > 0 {
        s.checkVariableConsistency(input, &output)
    }

    // Test rendering if sample data provided
    if input.SampleData != nil && output.IsValid {
        s.testTemplateRendering(input, &output)
    }

    // Add suggestions based on findings
    s.addValidationSuggestions(input, &output)

    return output
}

// validateHandlebarsTemplate validates Handlebars syntax
func (s *MCPServer) validateHandlebarsTemplate(input ValidateTemplateInput, output *ValidateTemplateOutput) {
    // Try to parse the template
    _, err := raymond.Parse(input.Content)
    if err != nil {
        output.IsValid = false
        
        // Parse error to extract line number if possible
        line, col := extractErrorLocation(err.Error())
        
        output.Errors = append(output.Errors, ValidationIssue{
            Severity:   "error",
            Line:       line,
            Column:     col,
            Message:    fmt.Sprintf("Syntax error: %v", err),
            Code:       "SYNTAX_ERROR",
            Suggestion: "Check for unclosed tags, mismatched braces, or invalid helper calls",
        })
        return
    }

    // Additional Handlebars-specific validation
    s.validateHandlebarsSyntaxRules(input.Content, output)
}

// validateHandlebarsSyntaxRules checks Handlebars-specific rules
func (s *MCPServer) validateHandlebarsSyntaxRules(content string, output *ValidateTemplateOutput) {
    lines := strings.Split(content, "\n")
    
    for i, line := range lines {
        lineNum := i + 1
        
        // Check for common issues
        
        // 1. Unescaped variables (security risk)
        if strings.Contains(line, "{{{") && !strings.Contains(line, "}}}") {
            output.Warnings = append(output.Warnings, ValidationIssue{
                Severity:   "warning",
                Line:       lineNum,
                Message:    "Unclosed unescaped variable block",
                Code:       "UNCLOSED_UNESCAPED_BLOCK",
                Suggestion: "Ensure all {{{ have matching }}}",
            })
        }
        
        // 2. Empty blocks
        if regexp.MustCompile(`\{\{\s*\}\}`).MatchString(line) {
            output.Warnings = append(output.Warnings, ValidationIssue{
                Severity:   "warning",
                Line:       lineNum,
                Message:    "Empty template block",
                Code:       "EMPTY_BLOCK",
                Suggestion: "Remove empty blocks or add variable name",
            })
        }
        
        // 3. Potentially unsafe helpers
        if strings.Contains(line, "{{eval") {
            output.Warnings = append(output.Warnings, ValidationIssue{
                Severity:   "warning",
                Line:       lineNum,
                Message:    "Use of 'eval' helper may be unsafe",
                Code:       "UNSAFE_HELPER",
                Suggestion: "Avoid dynamic evaluation, use safer alternatives",
            })
        }
    }
}

// detectTemplateVariables extracts variable names from template
func (s *MCPServer) detectTemplateVariables(content string, templateType string) []string {
    variables := make(map[string]bool) // Use map to deduplicate
    
    switch templateType {
    case "handlebars":
        // Match {{variable}} and {{{variable}}}
        re := regexp.MustCompile(`\{\{+\s*([a-zA-Z_][a-zA-Z0-9_\.]*)\s*\}+`)
        matches := re.FindAllStringSubmatch(content, -1)
        for _, match := range matches {
            if len(match) > 1 {
                varName := match[1]
                // Skip helpers (e.g., {{#if}}, {{#each}})
                if !strings.HasPrefix(varName, "#") && 
                   !strings.HasPrefix(varName, "/") &&
                   !strings.HasPrefix(varName, "!") {
                    variables[varName] = true
                }
            }
        }
    
    case "jinja2":
        // Match {{ variable }} and {% variable %}
        re := regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_\.]*)\s*\}\}`)
        matches := re.FindAllStringSubmatch(content, -1)
        for _, match := range matches {
            if len(match) > 1 {
                variables[match[1]] = true
            }
        }
    
    case "go-template":
        // Match {{.Variable}}
        re := regexp.MustCompile(`\{\{\s*\.([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)
        matches := re.FindAllStringSubmatch(content, -1)
        for _, match := range matches {
            if len(match) > 1 {
                variables[match[1]] = true
            }
        }
    }
    
    // Convert map to sorted slice
    result := make([]string, 0, len(variables))
    for v := range variables {
        result = append(result, v)
    }
    sort.Strings(result)
    
    return result
}

// checkVariableConsistency checks expected vs detected variables
func (s *MCPServer) checkVariableConsistency(input ValidateTemplateInput, output *ValidateTemplateOutput) {
    expectedVars := make(map[string]bool)
    for _, v := range input.Variables {
        expectedVars[v.Name] = v.Required
    }
    
    detectedVars := make(map[string]bool)
    for _, v := range output.DetectedVariables {
        detectedVars[v] = true
    }
    
    // Find missing required variables
    for _, v := range input.Variables {
        if v.Required && !detectedVars[v.Name] {
            output.MissingVariables = append(output.MissingVariables, v.Name)
            output.Warnings = append(output.Warnings, ValidationIssue{
                Severity:   "warning",
                Message:    fmt.Sprintf("Required variable '%s' not found in template", v.Name),
                Code:       "MISSING_REQUIRED_VARIABLE",
                Suggestion: fmt.Sprintf("Add {{%s}} to the template", v.Name),
            })
        }
    }
    
    // Find unknown variables
    for _, v := range output.DetectedVariables {
        if !expectedVars[v] {
            output.UnknownVariables = append(output.UnknownVariables, v)
            output.Infos = append(output.Infos, ValidationIssue{
                Severity: "info",
                Message:  fmt.Sprintf("Variable '%s' found but not in expected list", v),
                Code:     "UNKNOWN_VARIABLE",
            })
        }
    }
}

// testTemplateRendering tests rendering with sample data
func (s *MCPServer) testTemplateRendering(input ValidateTemplateInput, output *ValidateTemplateOutput) {
    switch input.TemplateType {
    case "handlebars":
        result, err := raymond.Render(input.Content, input.SampleData)
        if err != nil {
            output.Errors = append(output.Errors, ValidationIssue{
                Severity:   "error",
                Message:    fmt.Sprintf("Test rendering failed: %v", err),
                Code:       "RENDER_ERROR",
                Suggestion: "Check that sample data matches template variables",
            })
            output.IsValid = false
        } else {
            output.TestRenderResult = result
            output.Infos = append(output.Infos, ValidationIssue{
                Severity: "info",
                Message:  "Test rendering successful",
                Code:     "RENDER_SUCCESS",
            })
        }
    }
}

// addValidationSuggestions adds general suggestions
func (s *MCPServer) addValidationSuggestions(input ValidateTemplateInput, output *ValidateTemplateOutput) {
    // Suggest comprehensive validation if basic was used
    if input.ValidationLevel == "basic" {
        output.Suggestions = append(output.Suggestions, 
            "Run with validation_level='comprehensive' for more detailed checks")
    }
    
    // Suggest test data if not provided
    if input.SampleData == nil {
        output.Suggestions = append(output.Suggestions, 
            "Provide sample_data to test template rendering")
    }
    
    // Suggest documenting variables
    if len(output.DetectedVariables) > 0 && len(input.Variables) == 0 {
        output.Suggestions = append(output.Suggestions, 
            "Consider documenting expected variables for better validation")
    }
}

// Helper function to extract error location
func extractErrorLocation(errMsg string) (line int, col int) {
    // Try to parse error message for line/column
    // Example: "Parse error on line 5, column 10"
    lineRe := regexp.MustCompile(`line\s+(\d+)`)
    colRe := regexp.MustCompile(`column\s+(\d+)`)
    
    if matches := lineRe.FindStringSubmatch(errMsg); len(matches) > 1 {
        fmt.Sscanf(matches[1], "%d", &line)
    }
    
    if matches := colRe.FindStringSubmatch(errMsg); len(matches) > 1 {
        fmt.Sscanf(matches[1], "%d", &col)
    }
    
    return
}
```

**Business Logic Best Practices:**

1. **Separate Concerns**: Handler parses input, business logic validates
2. **Reusable**: Business logic can be called from tests
3. **Structured Output**: Build output incrementally
4. **Error Recovery**: Don't panic, collect all errors
5. **Performance**: Time operations, optimize hot paths

---

## Step 5: Error Handling

Proper error handling is crucial for good user experience.

### Error Patterns

```go
// 1. Invalid Input
if input.Content == "" {
    return errorResult("Template content is required"), nil
}

// 2. Business Logic Error
if err := validate(); err != nil {
    return errorResult("Validation failed: %v", err), nil
}

// 3. Internal Error
if err := s.repo.Get(id); err != nil {
    logger.Error("Repository error", "error", err)
    return errorResult("Internal error: failed to retrieve data"), nil
}

// 4. Partial Failure (still return success with warnings)
result := validateTemplate(input)
// result.IsValid might be false, but we return the result
return successResult(result)
```

### Error Result Helper

```go
// errorResult creates an error result
func errorResult(format string, args ...interface{}) *sdk.CallToolResult {
    message := fmt.Sprintf(format, args...)
    
    return &sdk.CallToolResult{
        IsError: true,
        Content: []sdk.Content{{
            Type: "text",
            Text: message,
        }},
    }
}
```

### Success Result Helper

```go
// successResult creates a success result
func successResult(data interface{}) (*sdk.CallToolResult, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return errorResult("Failed to marshal result: %v", err), nil
    }

    return &sdk.CallToolResult{
        Content: []sdk.Content{{
            Type: "text",
            Text: string(jsonData),
        }},
    }, nil
}
```

### Error Handling Levels

1. **User Errors**: Invalid input, validation failures → Return clear message
2. **Business Errors**: Resource not found, permission denied → Return descriptive error
3. **System Errors**: Database failure, network error → Log detailed error, return generic message
4. **Panics**: Shouldn't happen, but recover and log if they do

```go
func (s *MCPServer) handleToolWithRecovery(ctx context.Context, arguments interface{}) (result *sdk.CallToolResult, err error) {
    // Recover from panics
    defer func() {
        if r := recover(); r != nil {
            logger.Error("Panic in tool handler", "panic", r, "stack", debug.Stack())
            result = errorResult("Internal error occurred")
            err = nil
        }
    }()
    
    // Normal handling
    return s.handleTool(ctx, arguments)
}
```

---

## Step 6: Add Metrics Tracking

Track tool usage and performance for monitoring and optimization.

### Basic Metrics

```go
func (s *MCPServer) handleValidateTemplate(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    start := time.Now()
    var err error
    
    defer func() {
        duration := time.Since(start)
        // Record operation metrics
        s.metrics.RecordOperation("validate_template", duration, err == nil)
    }()
    
    // ... implementation ...
}
```

### Detailed Metrics

```go
defer func() {
    duration := time.Since(start)
    
    // Record in metrics collector
    s.metrics.RecordOperation("validate_template", duration, err == nil)
    
    // Record in performance metrics
    s.perfMetrics.Record(logger.PerformanceEntry{
        Timestamp: start,
        Operation: "validate_template",
        Duration:  duration,
        Success:   err == nil,
        ErrorMsg:  getErrorMessage(err),
        Metadata: map[string]interface{}{
            "template_type":    input.TemplateType,
            "content_length":   len(input.Content),
            "validation_level": input.ValidationLevel,
            "has_sample_data":  input.SampleData != nil,
            "variables_count":  len(input.Variables),
        },
    })
    
    // Log slow operations
    if duration > 1*time.Second {
        logger.Warn("Slow validation operation",
            "duration_ms", duration.Milliseconds(),
            "content_length", len(input.Content))
    }
}()
```

### Metrics Types

1. **Operation Count**: How many times tool was called
2. **Success Rate**: Percentage of successful calls
3. **Duration**: How long operations take (avg, p50, p95, p99)
4. **Error Rate**: Types and frequency of errors
5. **Custom Metrics**: Tool-specific measurements

---

## Step 7: Write Tests

Comprehensive testing ensures reliability and makes refactoring safe.

### Unit Tests

**Location:** `internal/mcp/validation_tools_test.go`

```go
package mcp

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/fsvxavier/nexs-mcp/internal/config"
    "github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

func TestHandleValidateTemplate(t *testing.T) {
    tests := []struct {
        name        string
        input       ValidateTemplateInput
        expectValid bool
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid handlebars template",
            input: ValidateTemplateInput{
                Content:      "Hello {{name}}!",
                TemplateType: "handlebars",
            },
            expectValid: true,
            expectError: false,
        },
        {
            name: "invalid syntax",
            input: ValidateTemplateInput{
                Content:      "Hello {{name}",
                TemplateType: "handlebars",
            },
            expectValid: false,
            expectError: false, // Not a tool error, validation just fails
        },
        {
            name: "missing content",
            input: ValidateTemplateInput{
                Content: "",
            },
            expectValid: false,
            expectError: true,
            errorMsg:    "content is required",
        },
        {
            name: "with sample data",
            input: ValidateTemplateInput{
                Content:      "Hello {{name}}!",
                TemplateType: "handlebars",
                SampleData: map[string]interface{}{
                    "name": "World",
                },
            },
            expectValid: true,
            expectError: false,
        },
        {
            name: "missing required variable",
            input: ValidateTemplateInput{
                Content:      "Hello {{name}}!",
                TemplateType: "handlebars",
                Variables: []TemplateVariable{
                    {Name: "name", Required: true},
                    {Name: "email", Required: true},
                },
            },
            expectValid: true, // Syntax is valid
            expectError: false,
            // Should have warning about missing 'email'
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            repo := infrastructure.NewInMemoryElementRepository()
            cfg := &config.Config{}
            server := NewMCPServer("test", "1.0.0", repo, cfg)

            // Execute
            result, err := server.handleValidateTemplate(context.Background(), tt.input)

            // Assert
            if tt.expectError {
                require.NoError(t, err) // Handler error (not tool error)
                assert.True(t, result.IsError)
                if tt.errorMsg != "" {
                    assert.Contains(t, result.Content[0].Text, tt.errorMsg)
                }
            } else {
                require.NoError(t, err)
                assert.False(t, result.IsError)
                
                // Parse output
                var output ValidateTemplateOutput
                err = json.Unmarshal([]byte(result.Content[0].Text), &output)
                require.NoError(t, err)
                
                assert.Equal(t, tt.expectValid, output.IsValid)
            }
        })
    }
}

func TestValidateTemplateContent(t *testing.T) {
    server := setupTestServer()

    t.Run("detects variables", func(t *testing.T) {
        input := ValidateTemplateInput{
            Content:      "Hello {{name}}, your email is {{email}}",
            TemplateType: "handlebars",
        }

        result := server.validateTemplateContent(input)

        assert.True(t, result.IsValid)
        assert.Contains(t, result.DetectedVariables, "name")
        assert.Contains(t, result.DetectedVariables, "email")
        assert.Equal(t, 2, len(result.DetectedVariables))
    })

    t.Run("finds missing variables", func(t *testing.T) {
        input := ValidateTemplateInput{
            Content:      "Hello {{name}}",
            TemplateType: "handlebars",
            Variables: []TemplateVariable{
                {Name: "name", Required: true},
                {Name: "email", Required: true},
            },
        }

        result := server.validateTemplateContent(input)

        assert.Contains(t, result.MissingVariables, "email")
        assert.True(t, len(result.Warnings) > 0)
    })

    t.Run("test rendering works", func(t *testing.T) {
        input := ValidateTemplateInput{
            Content:      "Hello {{name}}!",
            TemplateType: "handlebars",
            SampleData: map[string]interface{}{
                "name": "World",
            },
        }

        result := server.validateTemplateContent(input)

        assert.True(t, result.IsValid)
        assert.Equal(t, "Hello World!", result.TestRenderResult)
    })
}

func TestDetectTemplateVariables(t *testing.T) {
    server := setupTestServer()

    tests := []struct {
        name         string
        content      string
        templateType string
        expected     []string
    }{
        {
            name:         "simple handlebars",
            content:      "{{name}}",
            templateType: "handlebars",
            expected:     []string{"name"},
        },
        {
            name:         "multiple variables",
            content:      "{{first}} {{last}} {{email}}",
            templateType: "handlebars",
            expected:     []string{"email", "first", "last"}, // Sorted
        },
        {
            name:         "with helpers (should ignore)",
            content:      "{{#if condition}}{{name}}{{/if}}",
            templateType: "handlebars",
            expected:     []string{"name"}, // 'if' is ignored
        },
        {
            name:         "unescaped variables",
            content:      "{{{html}}}",
            templateType: "handlebars",
            expected:     []string{"html"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := server.detectTemplateVariables(tt.content, tt.templateType)
            assert.Equal(t, tt.expected, result)
        })
    }
}

// Test helper
func setupTestServer() *MCPServer {
    repo := infrastructure.NewInMemoryElementRepository()
    cfg := &config.Config{}
    return NewMCPServer("test", "1.0.0", repo, cfg)
}
```

### Integration Tests

```go
func TestValidateTemplateIntegration(t *testing.T) {
    // Setup real server with file repository
    tmpDir := t.TempDir()
    repo, err := infrastructure.NewFileElementRepository(tmpDir)
    require.NoError(t, err)
    
    cfg := &config.Config{}
    server := NewMCPServer("test", "1.0.0", repo, cfg)

    // Test full flow: create template, then validate it
    t.Run("full workflow", func(t *testing.T) {
        // 1. Create a template
        template := createTestTemplate()
        err := repo.Create(template)
        require.NoError(t, err)

        // 2. Validate the template
        input := ValidateTemplateInput{
            Content:      template.Content,
            TemplateType: template.TemplateType,
            Variables:    convertVariables(template.Variables),
        }

        result, err := server.handleValidateTemplate(context.Background(), input)
        require.NoError(t, err)
        assert.False(t, result.IsError)
    })
}
```

### Benchmark Tests

```go
func BenchmarkValidateTemplate(b *testing.B) {
    server := setupTestServer()
    input := ValidateTemplateInput{
        Content:      strings.Repeat("Hello {{name}} ", 100), // Large template
        TemplateType: "handlebars",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        server.validateTemplateContent(input)
    }
}
```

---

## Step 8: Update Documentation

Document the new tool in user-facing documentation.

**Location:** `docs/api/MCP_TOOLS.md`

```markdown
### validate_template

Validates template syntax and structure. Supports Handlebars, Jinja2, and Go templates.

**When to use:**
- Before saving a new template
- After editing template content
- To check variable consistency
- To test rendering with sample data

**Input:**
```json
{
  "content": "Hello {{name}}! Your order #{{order_id}} is {{status}}.",
  "template_type": "handlebars",
  "variables": [
    {
      "name": "name",
      "type": "string",
      "required": true,
      "description": "Customer name"
    },
    {
      "name": "order_id",
      "type": "number",
      "required": true,
      "description": "Order ID"
    },
    {
      "name": "status",
      "type": "string",
      "required": true,
      "description": "Order status"
    }
  ],
  "validation_level": "comprehensive",
  "sample_data": {
    "name": "John Doe",
    "order_id": 12345,
    "status": "shipped"
  }
}
```

**Output:**
```json
{
  "is_valid": true,
  "errors": [],
  "warnings": [],
  "infos": [
    {
      "severity": "info",
      "message": "Test rendering successful",
      "code": "RENDER_SUCCESS"
    }
  ],
  "detected_variables": ["name", "order_id", "status"],
  "missing_variables": [],
  "unknown_variables": [],
  "test_render_result": "Hello John Doe! Your order #12345 is shipped.",
  "validation_time_ms": 15,
  "suggestions": []
}
```

**Validation Levels:**
- `basic`: Syntax checking only (fast)
- `comprehensive`: Syntax + variable checking + best practices (default)
- `strict`: All checks + performance + security warnings

**Common Errors:**
- `SYNTAX_ERROR`: Template has syntax errors
- `MISSING_REQUIRED_VARIABLE`: Required variable not found in template
- `RENDER_ERROR`: Test rendering failed with sample data
- `UNSUPPORTED_TYPE`: Template type not supported

**Tips:**
1. Always validate before saving templates
2. Use `sample_data` to test rendering
3. Document expected variables for better validation
4. Use `comprehensive` level for production templates
```

---

## Complete Example: validate_template Tool

Here's the complete implementation you can use as a reference:

**Files created:**
1. `internal/mcp/validation_tools.go` - Handler and business logic
2. `internal/mcp/validation_tools_test.go` - Tests
3. Updated `internal/mcp/server.go` - Tool registration
4. Updated `docs/api/MCP_TOOLS.md` - Documentation

**Total Lines:** ~700 lines across all files

**The tool provides:**
- ✓ Syntax validation for multiple template types
- ✓ Variable detection and consistency checking  
- ✓ Test rendering with sample data
- ✓ Detailed error messages with line numbers
- ✓ Warnings and suggestions
- ✓ Performance metrics
- ✓ Comprehensive tests
- ✓ Full documentation

---

## Best Practices

### 1. Input Validation

```go
// Always validate required fields
if input.Content == "" {
    return errorResult("Content is required"), nil
}

// Validate enums
validTypes := []string{"handlebars", "jinja2", "go-template"}
if !contains(validTypes, input.TemplateType) {
    return errorResult("Invalid template_type: %s", input.TemplateType), nil
}

// Validate ranges
if input.MaxLength > 0 && len(input.Content) > input.MaxLength {
    return errorResult("Content exceeds max length"), nil
}
```

### 2. Provide Defaults

```go
// Set sensible defaults
if input.TemplateType == "" {
    input.TemplateType = "handlebars"
}
if input.ValidationLevel == "" {
    input.ValidationLevel = "comprehensive"
}
```

### 3. Rich Error Messages

```go
// Bad
return errorResult("Validation failed")

// Good
return errorResult("Validation failed at line %d, column %d: %s. Suggestion: %s",
    line, col, message, suggestion)
```

### 4. Structured Logging

```go
logger.Info("Tool executed",
    "tool", "validate_template",
    "duration_ms", duration.Milliseconds(),
    "content_length", len(input.Content),
    "result", output.IsValid)
```

### 5. Graceful Degradation

```go
// Don't fail completely if non-critical feature fails
result := validateSyntax(input)
if testRenderingErr != nil {
    result.Warnings = append(result.Warnings, Warning{
        Message: "Could not test rendering",
    })
    // Continue, don't return error
}
```

### 6. Performance Optimization

```go
// Cache expensive operations
var templateCache = make(map[string]*raymond.Template)

func (s *MCPServer) parseTemplate(content string) (*raymond.Template, error) {
    hash := hashContent(content)
    if cached, ok := templateCache[hash]; ok {
        return cached, nil
    }
    
    template, err := raymond.Parse(content)
    if err == nil {
        templateCache[hash] = template
    }
    return template, err
}
```

### 7. Test Coverage

Aim for:
- 80%+ code coverage
- Happy path tests
- Error cases
- Edge cases (empty, max size, etc.)
- Integration tests
- Benchmarks for critical paths

---

## Performance Considerations

### 1. Time Complexity

```go
// Bad: O(n²) nested loops
for _, item1 := range items {
    for _, item2 := range items {
        if item1.ID == item2.ID {
            // ...
        }
    }
}

// Good: O(n) with map
itemMap := make(map[string]Item)
for _, item := range items {
    itemMap[item.ID] = item
}
```

### 2. Memory Usage

```go
// Bad: Loads entire file into memory
content, _ := ioutil.ReadFile(filename)

// Good: Stream large files
file, _ := os.Open(filename)
defer file.Close()
scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    // Process line by line
}
```

### 3. Timeouts

```go
func (s *MCPServer) handleTool(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    // Set timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Check context periodically in long operations
    for _, item := range largeList {
        select {
        case <-ctx.Done():
            return errorResult("Operation timed out"), nil
        default:
            // Process item
        }
    }
}
```

### 4. Concurrency

```go
// Process multiple items concurrently
results := make(chan Result, len(items))
errChan := make(chan error, len(items))

for _, item := range items {
    go func(item Item) {
        result, err := process(item)
        if err != nil {
            errChan <- err
            return
        }
        results <- result
    }(item)
}

// Collect results
for i := 0; i < len(items); i++ {
    select {
    case result := <-results:
        allResults = append(allResults, result)
    case err := <-errChan:
        return errorResult("Processing failed: %v", err), nil
    }
}
```

---

## Common Patterns

### 1. Pagination

```go
type ListInput struct {
    Limit  int `json:"limit,omitempty"`
    Offset int `json:"offset,omitempty"`
}

type ListOutput struct {
    Items      []Item `json:"items"`
    Total      int    `json:"total"`
    Limit      int    `json:"limit"`
    Offset     int    `json:"offset"`
    HasMore    bool   `json:"has_more"`
}

func (s *MCPServer) handleList(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    var input ListInput
    parseArguments(arguments, &input)
    
    // Set defaults
    if input.Limit == 0 {
        input.Limit = 20
    }
    if input.Limit > 100 {
        input.Limit = 100 // Max limit
    }
    
    // Get items
    allItems, _ := s.repo.List()
    total := len(allItems)
    
    // Paginate
    start := input.Offset
    end := start + input.Limit
    if end > total {
        end = total
    }
    
    items := allItems[start:end]
    
    return successResult(ListOutput{
        Items:   items,
        Total:   total,
        Limit:   input.Limit,
        Offset:  input.Offset,
        HasMore: end < total,
    })
}
```

### 2. Filtering

```go
type FilterInput struct {
    Type      string   `json:"type,omitempty"`
    Tags      []string `json:"tags,omitempty"`
    Author    string   `json:"author,omitempty"`
    IsActive  *bool    `json:"is_active,omitempty"`
}

func applyFilters(items []Element, filter FilterInput) []Element {
    result := items
    
    if filter.Type != "" {
        result = filterByType(result, filter.Type)
    }
    
    if len(filter.Tags) > 0 {
        result = filterByTags(result, filter.Tags)
    }
    
    if filter.Author != "" {
        result = filterByAuthor(result, filter.Author)
    }
    
    if filter.IsActive != nil {
        result = filterByActive(result, *filter.IsActive)
    }
    
    return result
}
```

### 3. Batch Operations

```go
type BatchInput struct {
    Items []ItemInput `json:"items"`
}

type BatchOutput struct {
    Results  []ItemResult `json:"results"`
    Success  int          `json:"success"`
    Failed   int          `json:"failed"`
    Errors   []string     `json:"errors,omitempty"`
}

func (s *MCPServer) handleBatch(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    var input BatchInput
    parseArguments(arguments, &input)
    
    output := BatchOutput{
        Results: make([]ItemResult, len(input.Items)),
    }
    
    for i, item := range input.Items {
        result, err := s.processItem(item)
        if err != nil {
            output.Failed++
            output.Errors = append(output.Errors, 
                fmt.Sprintf("Item %d: %v", i, err))
            output.Results[i] = ItemResult{Success: false}
        } else {
            output.Success++
            output.Results[i] = result
        }
    }
    
    return successResult(output)
}
```

---

## Troubleshooting

### Tool Not Appearing in Client

**Cause:** Not registered or registration failed

**Fix:**
1. Check `registerTools()` has `sdk.AddTool()` call
2. Verify server starts without errors
3. Check tool name is unique
4. Rebuild: `go build ./cmd/nexs-mcp`

### Arguments Not Parsing

**Cause:** Input schema mismatch

**Fix:**
```go
// Add debug logging
var input MyInput
if err := parseArguments(arguments, &input); err != nil {
    logger.Error("Failed to parse arguments", 
        "error", err, 
        "arguments", arguments) // Log raw arguments
    return errorResult("Invalid arguments: %v", err), nil
}
```

### Tool Times Out

**Cause:** Long-running operation

**Fix:**
1. Add timeout context
2. Make operation async if needed
3. Optimize algorithm
4. Add progress reporting

### Memory Leaks

**Cause:** Not closing resources

**Fix:**
```go
func (s *MCPServer) handleTool(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    file, err := os.Open(filename)
    if err != nil {
        return errorResult("Failed to open file"), nil
    }
    defer file.Close() // Always close
    
    // ... use file ...
}
```

---

## Conclusion

You've learned how to create a complete MCP tool! Key points:

1. **Use Official SDK**: `github.com/modelcontextprotocol/go-sdk/mcp`
2. **Clear Schemas**: Well-documented input/output
3. **Separation**: Handler vs business logic
4. **Error Handling**: User-friendly messages
5. **Metrics**: Track usage and performance
6. **Testing**: Comprehensive test coverage
7. **Documentation**: Keep docs updated

**Next Steps:**
- Read [EXTENDING_VALIDATION.md](./EXTENDING_VALIDATION.md) for validation patterns
- Check [CODE_TOUR.md](./CODE_TOUR.md) for architecture details
- Look at existing tools for more examples

**Resources:**
- MCP SDK: https://github.com/modelcontextprotocol/go-sdk
- NEXS Tools: `internal/mcp/*_tools.go`
- API Docs: `docs/api/MCP_TOOLS.md`

Developer tutorials created successfully
