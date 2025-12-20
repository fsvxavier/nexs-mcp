# Extending Validation: Complete Guide

## Table of Contents

- [Introduction](#introduction)
- [Validation Architecture Overview](#validation-architecture-overview)
- [Types of Validation](#types-of-validation)
- [Step 1: Identify Validation Requirement](#step-1-identify-validation-requirement)
- [Step 2: Choose Validator Class](#step-2-choose-validator-class)
- [Step 3: Implement Validation Logic](#step-3-implement-validation-logic)
- [Step 4: Add Error/Warning Messages](#step-4-add-errorwarning-messages)
- [Step 5: Write Comprehensive Tests](#step-5-write-comprehensive-tests)
- [Step 6: Consider Backward Compatibility](#step-6-consider-backward-compatibility)
- [Examples](#examples)
- [Testing Validation Rules](#testing-validation-rules)
- [Validation Levels](#validation-levels)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Troubleshooting](#troubleshooting)

---

## Introduction

NEXS-MCP uses a sophisticated multi-level validation system to ensure data quality while maintaining flexibility. This guide teaches you how to extend and customize validation rules.

**What You'll Learn:**
- Validation architecture and design
- How to add new validation rules
- Different validation levels and when to use them
- Testing validation logic
- Backward compatibility considerations

**Prerequisites:**
- Understanding of NEXS-MCP domain models
- Familiarity with Go
- Basic knowledge of validation concepts

---

## Validation Architecture Overview

### Validation Flow

```
┌─────────────────────────────────────────────────────────────┐
│                     Element Creation/Update                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
                  ┌──────────────┐
                  │ Basic Check  │ element.Validate()
                  │  (Quick)     │ - Required fields
                  │              │ - Data types
                  └──────┬───────┘ - Basic structure
                         │
                         ▼
                  ┌──────────────┐
                  │ Validator    │ validator.Validate(element, level)
                  │ (Detailed)   │
                  └──────┬───────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
        ▼                ▼                ▼
  ┌──────────┐   ┌──────────────┐  ┌──────────┐
  │  Basic   │   │ Comprehensive│  │  Strict  │
  │  Level   │   │    Level     │  │  Level   │
  └──────────┘   └──────────────┘  └──────────┘
   Structure      + Business logic   + Best practices
   only           + Relationships    + Performance
                  + Cross-field      + Security
```

### Validation Components

```go
// 1. Element.Validate() - Quick structural validation
type Element interface {
    Validate() error  // Returns first error encountered
}

// 2. Validator - Comprehensive validation
type Validator interface {
    Validate(element Element, level ValidationLevel) ValidationResult
}

// 3. ValidationResult - Detailed results
type ValidationResult struct {
    IsValid        bool
    Errors         []ValidationIssue  // Blocking issues
    Warnings       []ValidationIssue  // Non-blocking issues
    Infos          []ValidationIssue  // Informational
    ValidationTime int64
    ElementType    string
    ElementID      string
}

// 4. ValidationIssue - Individual problem
type ValidationIssue struct {
    Severity   ValidationSeverity  // error, warning, info
    Field      string              // Which field has the issue
    Message    string              // Human-readable description
    Line       int                 // Line number (if applicable)
    Suggestion string              // How to fix it
    Code       string              // Machine-readable code
}
```

### Validation Layers

1. **Structural Layer** (`Element.Validate()`):
   - Fast, synchronous
   - Checks required fields, types, formats
   - Returns on first error
   - No external dependencies

2. **Business Logic Layer** (`Validator.Validate()`):
   - More thorough
   - Checks relationships, references
   - Collects all issues
   - May query repository

3. **Best Practices Layer** (Strict level):
   - Optional checks
   - Performance suggestions
   - Security warnings
   - Documentation recommendations

---

## Types of Validation

### 1. Structural Validation

Checks basic structure and data types:

```go
// Required fields
if persona.Role == "" {
    return fmt.Errorf("role is required")
}

// String length
if len(persona.Name) < 3 || len(persona.Name) > 100 {
    return fmt.Errorf("name must be between 3 and 100 characters")
}

// Numeric ranges
if workflow.MaxDuration < 1 {
    return fmt.Errorf("max_duration must be at least 1 second")
}

// Enum values
validTones := []string{"professional", "casual", "formal", "friendly"}
if !contains(validTones, persona.Tone) {
    return fmt.Errorf("invalid tone: must be one of %v", validTones)
}

// Array constraints
if len(skill.Examples) < 1 {
    return fmt.Errorf("at least one example is required")
}
```

### 2. Business Rules Validation

Checks domain-specific logic:

```go
// Cross-field validation
if agent.MaxIterations * agent.Timeout > 3600 {
    result.AddWarning("agent", 
        "Total possible execution time exceeds 1 hour", 
        "EXECUTION_TIME_HIGH")
}

// Reference validation
for _, skillID := range agent.SkillIDs {
    if _, err := s.repo.Get(skillID); err != nil {
        result.AddError(fmt.Sprintf("skill_ids[%d]", i),
            fmt.Sprintf("Referenced skill '%s' not found", skillID),
            "SKILL_NOT_FOUND")
    }
}

// State consistency
if ensemble.Strategy == "consensus" && ensemble.MinConsensus == 0 {
    result.AddError("min_consensus",
        "min_consensus is required for consensus strategy",
        "CONSENSUS_VALUE_REQUIRED")
}

// Circular dependency detection
if hasCycle(workflow.Steps) {
    result.AddError("steps",
        "Circular dependency detected in workflow steps",
        "CIRCULAR_DEPENDENCY")
}
```

### 3. Format Validation

Validates specific formats:

```go
// Email
emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
if !emailRegex.MatchString(persona.ContactEmail) {
    result.AddError("contact_email", 
        "Invalid email format", 
        "EMAIL_INVALID")
}

// Semantic version
semverRegex := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`)
if !semverRegex.MatchString(element.Version) {
    result.AddError("version",
        "Version must be semantic version (e.g., 1.0.0)",
        "VERSION_INVALID")
}

// Cron expression
if trigger.Type == "schedule" {
    if !isValidCron(trigger.Schedule) {
        result.AddWarning("schedule",
            "Schedule may not be a valid cron expression",
            "CRON_INVALID")
    }
}

// URL
if _, err := url.Parse(webhook.URL); err != nil {
    result.AddError("url", "Invalid URL format", "URL_INVALID")
}

// Date ranges
if endDate.Before(startDate) {
    result.AddError("end_date",
        "End date must be after start date",
        "DATE_RANGE_INVALID")
}
```

### 4. Content Quality Validation

Checks content quality and best practices:

```go
// Length recommendations
if len(persona.Context) < 100 {
    result.AddWarning("context",
        "Context is quite short, consider adding more detail",
        "CONTEXT_SHORT")
}

if len(persona.Context) > 5000 {
    result.AddWarning("context",
        "Very long context may cause performance issues",
        "CONTEXT_LONG")
}

// Complexity analysis
if countWords(template.Content) > 1000 {
    result.AddInfo("content",
        "Large template, consider breaking into partials",
        "TEMPLATE_LARGE")
}

// Readability
if readabilityScore(persona.Context) < 50 {
    result.AddInfo("context",
        "Context may be difficult to understand",
        "READABILITY_LOW")
}

// Consistency checks
if containsInconsistentTone(persona.Context, persona.Tone) {
    result.AddWarning("context",
        "Context tone doesn't match declared tone",
        "TONE_INCONSISTENT")
}
```

---

## Step 1: Identify Validation Requirement

### Questions to Ask

1. **What are you validating?**
   - A single field value?
   - Relationship between fields?
   - External references?
   - Best practices?

2. **When should it fail?**
   - Always (error)?
   - Usually (warning)?
   - Optionally (info)?

3. **What level?**
   - Basic (structure only)?
   - Comprehensive (business logic)?
   - Strict (best practices)?

4. **Performance impact?**
   - Fast check (< 1ms)?
   - May need I/O (repository lookup)?
   - Complex computation?

### Example Scenarios

**Scenario 1: Email Validation**
- What: Email format in persona contact field
- When: Always validate format (warning)
- Level: Comprehensive
- Performance: Fast (regex)

**Scenario 2: Referenced Element Exists**
- What: Agent references skills by ID
- When: Fail if skill doesn't exist (error)
- Level: Comprehensive
- Performance: Slow (repository lookup)

**Scenario 3: Template Size Recommendation**
- What: Large templates should be split
- When: Recommend only (info)
- Level: Strict
- Performance: Fast (string length)

---

## Step 2: Choose Validator Class

Validation can be added in different places:

### Option 1: Element.Validate() Method

**Use for:**
- Required field checks
- Basic type validation
- Fast, synchronous checks
- No external dependencies

**Example:**
```go
func (p *Persona) Validate() error {
    if err := p.Metadata.Validate(); err != nil {
        return err
    }
    
    if p.Role == "" {
        return fmt.Errorf("role is required")
    }
    
    if len(p.Expertise) == 0 {
        return fmt.Errorf("at least one expertise area required")
    }
    
    return nil
}
```

### Option 2: Dedicated Validator

**Use for:**
- Complex validation logic
- Multiple validation levels
- Collecting all issues (not just first)
- Repository lookups
- Best practices checks

**Example:**
```go
type PersonaValidator struct {
    repo domain.ElementRepository
}

func (v *PersonaValidator) Validate(element domain.Element, level ValidationLevel) ValidationResult {
    // Comprehensive validation with detailed results
}
```

### Option 3: Shared Validation Functions

**Use for:**
- Common validation patterns
- Reusable across validators
- Utility functions

**Example:**
```go
// internal/validation/common.go
func ValidateEmail(email string) bool {
    re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return re.MatchString(email)
}

func ValidateSemver(version string) bool {
    re := regexp.MustCompile(`^\d+\.\d+\.\d+`)
    return re.MatchString(version)
}

func ValidateURL(urlStr string) bool {
    _, err := url.Parse(urlStr)
    return err == nil
}
```

---

## Step 3: Implement Validation Logic

### Adding to Element.Validate()

**Location:** `internal/domain/{element_type}.go`

```go
func (w *Workflow) Validate() error {
    // 1. Validate metadata
    if err := w.Metadata.Validate(); err != nil {
        return fmt.Errorf("metadata validation failed: %w", err)
    }

    // 2. Validate required fields
    if len(w.Steps) == 0 {
        return fmt.Errorf("workflow must have at least one step")
    }

    if len(w.Triggers) == 0 {
        return fmt.Errorf("workflow must have at least one trigger")
    }

    // 3. Validate field constraints
    if w.MaxDuration < 1 {
        return fmt.Errorf("max_duration must be at least 1 second")
    }

    // 4. Validate nested structures
    stepIDs := make(map[string]bool)
    for i, step := range w.Steps {
        if step.ID == "" {
            return fmt.Errorf("step %d: ID is required", i)
        }
        
        if stepIDs[step.ID] {
            return fmt.Errorf("step %d: duplicate ID '%s'", i, step.ID)
        }
        stepIDs[step.ID] = true

        if step.Type == "" {
            return fmt.Errorf("step %d: type is required", i)
        }
    }

    // 5. Validate conditional requirements
    if w.RetryPolicy != nil {
        if err := w.validateRetryPolicy(); err != nil {
            return fmt.Errorf("retry_policy: %w", err)
        }
    }

    return nil
}

func (w *Workflow) validateRetryPolicy() error {
    if w.RetryPolicy.MaxAttempts < 1 || w.RetryPolicy.MaxAttempts > 10 {
        return fmt.Errorf("max_attempts must be between 1 and 10")
    }
    
    if w.RetryPolicy.InitialDelay < 1 {
        return fmt.Errorf("initial_delay must be at least 1 second")
    }
    
    return nil
}
```

### Adding to Validator

**Location:** `internal/validation/{element_type}_validator.go`

```go
func (v *WorkflowValidator) Validate(element domain.Element, level ValidationLevel) ValidationResult {
    workflow, ok := element.(*domain.Workflow)
    if !ok {
        return invalidTypeResult()
    }

    result := ValidationResult{
        IsValid:     true,
        ElementType: "workflow",
        ElementID:   workflow.GetID(),
    }

    start := time.Now()
    defer func() {
        result.ValidationTime = time.Since(start).Milliseconds()
    }()

    // Basic validation
    v.validateBasic(workflow, &result)

    // Comprehensive validation
    if level == ComprehensiveLevel || level == StrictLevel {
        v.validateComprehensive(workflow, &result)
    }

    // Strict validation
    if level == StrictLevel {
        v.validateStrict(workflow, &result)
    }

    return result
}

func (v *WorkflowValidator) validateBasic(workflow *domain.Workflow, result *ValidationResult) {
    // Structure and required fields
    if err := workflow.Metadata.Validate(); err != nil {
        result.AddError("metadata", err.Error(), "METADATA_INVALID")
    }

    if len(workflow.Steps) == 0 {
        result.AddError("steps", "At least one step required", "STEPS_REQUIRED")
    }

    // Validate each step
    for i, step := range workflow.Steps {
        v.validateStep(step, i, result)
    }
}

func (v *WorkflowValidator) validateStep(step domain.WorkflowStep, index int, result *ValidationResult) {
    prefix := fmt.Sprintf("steps[%d]", index)

    if step.ID == "" {
        result.AddError(fmt.Sprintf("%s.id", prefix), 
            "Step ID is required", "STEP_ID_REQUIRED")
    }

    if step.Name == "" {
        result.AddError(fmt.Sprintf("%s.name", prefix),
            "Step name is required", "STEP_NAME_REQUIRED")
    }

    if step.Type == "" {
        result.AddError(fmt.Sprintf("%s.type", prefix),
            "Step type is required", "STEP_TYPE_REQUIRED")
    } else {
        validTypes := []string{"action", "condition", "loop", "parallel"}
        if !contains(validTypes, step.Type) {
            result.AddError(fmt.Sprintf("%s.type", prefix),
                fmt.Sprintf("Invalid type: %s (must be %s)", step.Type, strings.Join(validTypes, ", ")),
                "STEP_TYPE_INVALID")
        }
    }
}

func (v *WorkflowValidator) validateComprehensive(workflow *domain.Workflow, result *ValidationResult) {
    // Business logic validation
    v.checkStepDependencies(workflow, result)
    v.checkReachability(workflow, result)
    v.checkTimeouts(workflow, result)
    v.checkTriggers(workflow, result)
}

func (v *WorkflowValidator) validateStrict(workflow *domain.Workflow, result *ValidationResult) {
    // Best practices
    v.checkNamingConventions(workflow, result)
    v.checkDocumentation(workflow, result)
    v.checkComplexity(workflow, result)
    v.checkErrorHandling(workflow, result)
}
```

---

## Step 4: Add Error/Warning Messages

### Message Guidelines

1. **Be Specific**: Say exactly what's wrong
2. **Be Helpful**: Suggest how to fix it
3. **Be Consistent**: Use similar patterns
4. **Provide Context**: Include field names, values

### Message Patterns

```go
// Bad messages
result.AddError("field", "Invalid", "ERROR")
result.AddWarning("value", "Wrong", "WARNING")

// Good messages
result.AddError("email", 
    "Email 'user@invalid' has invalid format",
    "EMAIL_INVALID")

result.AddWarning("steps", 
    "Workflow has 25 steps, consider breaking into smaller workflows for maintainability",
    "WORKFLOW_COMPLEX")

result.AddErrorWithSuggestion("skill_ids[2]",
    "Referenced skill 'skill-123' not found",
    "SKILL_NOT_FOUND",
    "Check that skill ID exists or create it first")
```

### Error Codes

Use consistent naming:

```go
const (
    // Field-level errors
    "FIELD_REQUIRED"
    "FIELD_INVALID"
    "FIELD_TOO_SHORT"
    "FIELD_TOO_LONG"
    
    // Type-level errors
    "TYPE_INVALID"
    "FORMAT_INVALID"
    "RANGE_INVALID"
    
    // Reference errors
    "REFERENCE_NOT_FOUND"
    "CIRCULAR_DEPENDENCY"
    
    // Business logic errors
    "CONSTRAINT_VIOLATED"
    "STATE_INVALID"
    
    // Best practice warnings
    "COMPLEXITY_HIGH"
    "DOCUMENTATION_MISSING"
    "PERFORMANCE_CONCERN"
)
```

### Severity Levels

```go
// ERROR: Blocks creation/update
result.AddError("field", "message", "CODE")
// Use when: Data is invalid, operation must fail

// WARNING: Allows creation but flags issue
result.AddWarning("field", "message", "CODE")
// Use when: Data is valid but suboptimal

// INFO: Informational only
result.AddInfo("field", "message", "CODE")
// Use when: Suggestion or recommendation
```

---

## Step 5: Write Comprehensive Tests

### Test Structure

```go
// internal/validation/{type}_validator_test.go

func TestWorkflowValidator_Validate(t *testing.T) {
    validator := NewWorkflowValidator()

    tests := []struct {
        name           string
        workflow       *domain.Workflow
        level          ValidationLevel
        expectValid    bool
        expectErrors   int
        expectWarnings int
        expectCode     string // Specific error code to check
    }{
        {
            name:         "valid workflow",
            workflow:     createValidWorkflow(),
            level:        BasicLevel,
            expectValid:  true,
            expectErrors: 0,
        },
        {
            name:         "missing steps",
            workflow:     workflowWithoutSteps(),
            level:        BasicLevel,
            expectValid:  false,
            expectErrors: 1,
            expectCode:   "STEPS_REQUIRED",
        },
        {
            name:           "circular dependency warning",
            workflow:       workflowWithCircularDep(),
            level:          ComprehensiveLevel,
            expectValid:    true,
            expectErrors:   0,
            expectWarnings: 1,
            expectCode:     "CIRCULAR_DEPENDENCY",
        },
        {
            name:           "complexity warning in strict mode",
            workflow:       complexWorkflow(),
            level:          StrictLevel,
            expectValid:    true,
            expectWarnings: 1,
            expectCode:     "WORKFLOW_COMPLEX",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validator.Validate(tt.workflow, tt.level)

            assert.Equal(t, tt.expectValid, result.IsValid)
            assert.Equal(t, tt.expectErrors, len(result.Errors))

            if tt.expectWarnings > 0 {
                assert.GreaterOrEqual(t, len(result.Warnings), tt.expectWarnings)
            }

            if tt.expectCode != "" {
                found := false
                for _, err := range result.Errors {
                    if err.Code == tt.expectCode {
                        found = true
                        break
                    }
                }
                for _, warn := range result.Warnings {
                    if warn.Code == tt.expectCode {
                        found = true
                        break
                    }
                }
                assert.True(t, found, "Expected error code %s not found", tt.expectCode)
            }
        })
    }
}

// Test individual validation functions
func TestValidateStep(t *testing.T) {
    validator := NewWorkflowValidator()
    result := ValidationResult{IsValid: true}

    tests := []struct {
        name        string
        step        domain.WorkflowStep
        expectError bool
        errorField  string
    }{
        {
            name: "valid step",
            step: domain.WorkflowStep{
                ID:   "step1",
                Name: "Test Step",
                Type: "action",
            },
            expectError: false,
        },
        {
            name: "missing ID",
            step: domain.WorkflowStep{
                Name: "Test",
                Type: "action",
            },
            expectError: true,
            errorField:  "steps[0].id",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ValidationResult{IsValid: true}
            validator.validateStep(tt.step, 0, &result)

            if tt.expectError {
                assert.False(t, result.IsValid)
                assert.Greater(t, len(result.Errors), 0)
                assert.Contains(t, result.Errors[0].Field, tt.errorField)
            } else {
                assert.True(t, result.IsValid)
            }
        })
    }
}

// Test helper functions
func createValidWorkflow() *domain.Workflow {
    return &domain.Workflow{
        Metadata: validMetadata(),
        Steps: []domain.WorkflowStep{
            {ID: "step1", Name: "Step 1", Type: "action"},
        },
        Triggers: []domain.WorkflowTrigger{
            {Type: "manual"},
        },
        MaxDuration: 3600,
    }
}
```

### Edge Cases to Test

```go
// Empty/nil values
func TestValidation_EmptyValues(t *testing.T) {
    tests := []struct {
        name string
        workflow *domain.Workflow
    }{
        {"empty steps", &domain.Workflow{Steps: []domain.WorkflowStep{}}},
        {"nil steps", &domain.Workflow{Steps: nil}},
        {"empty triggers", &domain.Workflow{Triggers: []domain.WorkflowTrigger{}}},
    }
    // Test each case
}

// Boundary values
func TestValidation_BoundaryValues(t *testing.T) {
    // Test min/max values
    // Test exactly at boundaries
}

// Invalid combinations
func TestValidation_InvalidCombinations(t *testing.T) {
    // Test incompatible field combinations
}
```

---

## Step 6: Consider Backward Compatibility

### Backward Compatibility Strategies

#### 1. Additive Changes Only

```go
// ✓ Good: Adding optional field
type Workflow struct {
    Steps []WorkflowStep
    // New field with default
    RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`
}

// ✗ Bad: Making existing field required
type Workflow struct {
    Steps []WorkflowStep
    // This breaks existing workflows!
    RetryPolicy *RetryPolicy `json:"retry_policy"` // Required now
}
```

#### 2. Graceful Degradation

```go
func (v *WorkflowValidator) validateNewFeature(workflow *domain.Workflow, result *ValidationResult) {
    // Only validate if new field is present
    if workflow.RetryPolicy != nil {
        // Validate retry policy
    }
    // If not present, don't fail validation
}
```

#### 3. Version-Aware Validation

```go
func (v *WorkflowValidator) Validate(element domain.Element, level ValidationLevel) ValidationResult {
    workflow := element.(*domain.Workflow)
    
    // Check version
    version := parseVersion(workflow.Metadata.Version)
    
    // Apply version-appropriate validation
    if version.Major >= 2 {
        v.validateV2Features(workflow, &result)
    } else {
        v.validateV1Features(workflow, &result)
    }
}
```

#### 4. Deprecation Warnings

```go
func (v *WorkflowValidator) validateStrict(workflow *domain.Workflow, result *ValidationResult) {
    // Warn about deprecated features
    if workflow.LegacyField != "" {
        result.AddWarning("legacy_field",
            "Field 'legacy_field' is deprecated and will be removed in v3.0. Use 'new_field' instead",
            "FIELD_DEPRECATED")
    }
}
```

### Migration Strategy

```go
// Support both old and new formats
func workflowFromMap(data map[string]interface{}) (*domain.Workflow, error) {
    workflow := &domain.Workflow{}
    
    // Try new format first
    if retryPolicy, ok := data["retry_policy"]; ok {
        workflow.RetryPolicy = parseRetryPolicy(retryPolicy)
    } else if legacyRetry, ok := data["legacy_retry_config"]; ok {
        // Migrate from old format
        workflow.RetryPolicy = migrateLegacyRetry(legacyRetry)
    }
    
    return workflow, nil
}
```

---

## Examples

### Example 1: Email Validation

```go
func (v *PersonaValidator) validateEmailFormat(persona *domain.Persona, result *ValidationResult) {
    if persona.ContactEmail == "" {
        return // Email is optional
    }
    
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(persona.ContactEmail) {
        result.AddWarning("contact_email",
            fmt.Sprintf("Email '%s' may be invalid", persona.ContactEmail),
            "EMAIL_INVALID")
    }
}
```

### Example 2: Date Range Validation

```go
func (v *MemoryValidator) validateDateRange(memory *domain.Memory, result *ValidationResult) {
    now := time.Now()
    
    // Check if timestamp is in the past
    if memory.Timestamp.After(now) {
        result.AddWarning("timestamp",
            "Timestamp is in the future",
            "TIMESTAMP_FUTURE")
    }
    
    // Check if timestamp is too old
    oneYearAgo := now.AddDate(-1, 0, 0)
    if memory.Timestamp.Before(oneYearAgo) {
        result.AddInfo("timestamp",
            "Memory is over 1 year old, consider archiving",
            "TIMESTAMP_OLD")
    }
}
```

### Example 3: Reference Validation

```go
func (v *EnsembleValidator) validateElementReferences(ensemble *domain.Ensemble, result *ValidationResult) {
    if v.repo == nil {
        return // Can't validate without repository
    }
    
    for i, elementID := range ensemble.ElementIDs {
        element, err := v.repo.Get(elementID)
        if err != nil {
            result.AddError(fmt.Sprintf("element_ids[%d]", i),
                fmt.Sprintf("Referenced element '%s' not found", elementID),
                "ELEMENT_NOT_FOUND")
            continue
        }
        
        // Check if element is active
        if !element.IsActive() {
            result.AddWarning(fmt.Sprintf("element_ids[%d]", i),
                fmt.Sprintf("Referenced element '%s' is inactive", elementID),
                "ELEMENT_INACTIVE")
        }
        
        // Check element type compatibility
        if !v.isCompatibleType(element.GetType(), ensemble.Strategy) {
            result.AddWarning(fmt.Sprintf("element_ids[%d]", i),
                fmt.Sprintf("Element type '%s' may not work well with strategy '%s'",
                    element.GetType(), ensemble.Strategy),
                "TYPE_INCOMPATIBLE")
        }
    }
}
```

### Example 4: Complexity Analysis

```go
func (v *TemplateValidator) validateComplexity(template *domain.Template, result *ValidationResult) {
    content := template.Content
    
    // Count variables
    varCount := countVariables(content)
    if varCount > 20 {
        result.AddWarning("content",
            fmt.Sprintf("Template has %d variables, consider breaking into smaller templates", varCount),
            "VARIABLE_COUNT_HIGH")
    }
    
    // Check nesting depth
    maxDepth := calculateNestingDepth(content)
    if maxDepth > 4 {
        result.AddWarning("content",
            fmt.Sprintf("Template nesting depth is %d, consider simplifying", maxDepth),
            "NESTING_DEEP")
    }
    
    // Check size
    if len(content) > 10000 {
        result.AddInfo("content",
            "Large template (>10KB), may impact performance",
            "TEMPLATE_LARGE")
    }
}

func countVariables(content string) int {
    re := regexp.MustCompile(`\{\{[^}]+\}\}`)
    return len(re.FindAllString(content, -1))
}

func calculateNestingDepth(content string) int {
    maxDepth := 0
    currentDepth := 0
    
    for _, line := range strings.Split(content, "\n") {
        if strings.Contains(line, "{{#") {
            currentDepth++
            if currentDepth > maxDepth {
                maxDepth = currentDepth
            }
        } else if strings.Contains(line, "{{/") {
            currentDepth--
        }
    }
    
    return maxDepth
}
```

### Example 5: Security Validation

```go
func (v *TemplateValidator) validateSecurity(template *domain.Template, result *ValidationResult) {
    content := template.Content
    
    // Check for potentially unsafe patterns
    unsafePatterns := map[string]string{
        `\{\{\{.*\}\}\}`:       "Unescaped HTML may cause XSS vulnerabilities",
        `eval\s*\(`:            "eval() is dangerous and should be avoided",
        `innerHTML`:            "innerHTML may cause XSS, use textContent",
        `<script`:              "Inline scripts should be avoided",
    }
    
    for pattern, message := range unsafePatterns {
        re := regexp.MustCompile(pattern)
        if re.MatchString(content) {
            result.AddWarning("content", message, "SECURITY_RISK")
        }
    }
}
```

---

## Testing Validation Rules

### Unit Test Pattern

```go
func TestValidationRule_Scenario(t *testing.T) {
    // Arrange
    validator := NewValidator()
    element := createTestElement()
    
    // Act
    result := validator.Validate(element, ComprehensiveLevel)
    
    // Assert
    assert.True(t, result.IsValid)
    assert.Equal(t, 0, len(result.Errors))
    assert.Equal(t, 1, len(result.Warnings))
    assert.Equal(t, "WARNING_CODE", result.Warnings[0].Code)
}
```

### Table-Driven Tests

```go
func TestValidation_TableDriven(t *testing.T) {
    tests := []struct {
        name    string
        input   *domain.Element
        level   ValidationLevel
        want    ValidationResult
    }{
        {
            name:  "valid basic",
            input: validElement(),
            level: BasicLevel,
            want: ValidationResult{
                IsValid: true,
                Errors:  []ValidationIssue{},
            },
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := validator.Validate(tt.input, tt.level)
            assert.Equal(t, tt.want.IsValid, got.IsValid)
            // More assertions...
        })
    }
}
```

### Regression Tests

```go
// Test that previously failing inputs now pass
func TestValidation_Regression(t *testing.T) {
    // Issue #123: Circular dependency not detected
    t.Run("issue-123", func(t *testing.T) {
        workflow := createWorkflowWithCircularDep()
        result := validator.Validate(workflow, ComprehensiveLevel)
        
        // Should now detect circular dependency
        hasCircularError := false
        for _, err := range result.Errors {
            if err.Code == "CIRCULAR_DEPENDENCY" {
                hasCircularError = true
            }
        }
        assert.True(t, hasCircularError, "Should detect circular dependency")
    })
}
```

---

## Validation Levels

### Basic Level

**Use for:**
- Quick checks during creation
- API endpoints with strict performance requirements
- Bulk operations

**Includes:**
- Required fields
- Data types
- Format validation
- Basic constraints

**Typical time:** < 1ms

```go
if level == BasicLevel {
    return result
}
```

### Comprehensive Level (Default)

**Use for:**
- Standard validation
- User-facing tools
- Most operations

**Includes:**
- Everything in Basic
- Business rules
- Reference checking
- Cross-field validation
- Relationship consistency

**Typical time:** 5-50ms

```go
if level == ComprehensiveLevel || level == StrictLevel {
    v.validateComprehensive(element, &result)
}
```

### Strict Level

**Use for:**
- Quality assurance
- Pre-deployment checks
- Manual review triggers

**Includes:**
- Everything in Comprehensive
- Best practices
- Performance recommendations
- Security warnings
- Documentation completeness

**Typical time:** 50-200ms

```go
if level == StrictLevel {
    v.validateStrict(element, &result)
}
```

---

## Best Practices

### 1. Fail Fast in Element.Validate()

```go
func (e *Element) Validate() error {
    if e.Name == "" {
        return fmt.Errorf("name is required")
    }
    // Don't continue if basic validation fails
    return nil
}
```

### 2. Collect All Issues in Validators

```go
func (v *Validator) Validate(element Element, level ValidationLevel) ValidationResult {
    result := ValidationResult{IsValid: true}
    
    // Don't return early, collect all issues
    v.checkField1(element, &result)
    v.checkField2(element, &result)
    v.checkField3(element, &result)
    
    return result
}
```

### 3. Provide Context in Error Messages

```go
// Bad
result.AddError("field", "Invalid", "ERROR")

// Good
result.AddError("steps[3].action",
    "Action 'invalid_action' is not recognized. Available actions: create, update, delete",
    "ACTION_INVALID")
```

### 4. Use Appropriate Severity

```go
// Error: Blocks operation
result.AddError("field", "Required field missing", "FIELD_REQUIRED")

// Warning: Flags potential issue
result.AddWarning("field", "Value may cause performance issues", "PERFORMANCE_CONCERN")

// Info: Helpful suggestion
result.AddInfo("field", "Consider adding description for better documentation", "DOCUMENTATION_SUGGESTED")
```

### 5. Make Validation Deterministic

```go
// Bad: Time-dependent validation
if element.CreatedAt.After(time.Now()) {
    result.AddError("created_at", "Cannot be in future", "TIME_INVALID")
}

// Good: Deterministic validation
if element.CreatedAt.After(element.UpdatedAt) {
    result.AddError("created_at", "Cannot be after updated_at", "TIME_ORDER_INVALID")
}
```

### 6. Document Validation Rules

```go
// validateEmailFormat checks if the contact email has valid format.
// Returns warning (not error) since email is optional.
// RFC 5322 compliant regex pattern.
func (v *PersonaValidator) validateEmailFormat(persona *domain.Persona, result *ValidationResult) {
    // Implementation
}
```

---

## Common Patterns

### Pattern 1: Repository Lookup

```go
func (v *Validator) validateReferences(element *domain.Element, result *ValidationResult) {
    if v.repo == nil {
        // Can't validate without repository, skip
        return
    }
    
    for i, refID := range element.References {
        if _, err := v.repo.Get(refID); err != nil {
            result.AddError(fmt.Sprintf("references[%d]", i),
                fmt.Sprintf("Reference '%s' not found", refID),
                "REFERENCE_NOT_FOUND")
        }
    }
}
```

### Pattern 2: Conditional Validation

```go
func (v *Validator) validateConditional(element *domain.Element, result *ValidationResult) {
    // Only validate if condition met
    if element.Type == "specific_type" {
        v.validateSpecificFields(element, result)
    }
    
    // Validate required dependencies
    if element.UsesFeature {
        if element.FeatureConfig == nil {
            result.AddError("feature_config",
                "Configuration required when feature is enabled",
                "CONFIG_REQUIRED")
        }
    }
}
```

### Pattern 3: Aggregated Validation

```go
func (v *Validator) validateCollection(elements []*domain.Element, result *ValidationResult) {
    // Check collection-level constraints
    if len(elements) < 2 {
        result.AddError("elements",
            "At least 2 elements required",
            "COLLECTION_TOO_SMALL")
        return
    }
    
    // Check uniqueness
    seen := make(map[string]bool)
    for i, elem := range elements {
        if seen[elem.ID] {
            result.AddError(fmt.Sprintf("elements[%d]", i),
                fmt.Sprintf("Duplicate element ID: %s", elem.ID),
                "DUPLICATE_ID")
        }
        seen[elem.ID] = true
    }
    
    // Validate each element
    for i, elem := range elements {
        if err := elem.Validate(); err != nil {
            result.AddError(fmt.Sprintf("elements[%d]", i),
                err.Error(),
                "ELEMENT_INVALID")
        }
    }
}
```

---

## Troubleshooting

### Validation Too Slow

**Symptom:** Validation takes > 100ms

**Solutions:**
1. Move expensive checks to comprehensive/strict levels
2. Cache repository lookups
3. Optimize regex patterns
4. Use goroutines for independent checks

```go
// Parallel validation
func (v *Validator) validateParallel(element *domain.Element, result *ValidationResult) {
    var wg sync.WaitGroup
    results := make(chan ValidationIssue, 10)
    
    // Start multiple validations concurrently
    wg.Add(3)
    go func() { defer wg.Done(); v.check1(element, results) }()
    go func() { defer wg.Done(); v.check2(element, results) }()
    go func() { defer wg.Done(); v.check3(element, results) }()
    
    // Close results channel when all done
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    for issue := range results {
        result.Errors = append(result.Errors, issue)
    }
}
```

### Validation Not Triggered

**Symptom:** Invalid data passes validation

**Check:**
1. Is validation called? (Add logging)
2. Is validator registered?
3. Is validation level correct?
4. Is result checked?

```go
// Add debug logging
logger.Debug("Validating element",
    "type", element.GetType(),
    "id", element.GetID(),
    "level", level)
```

### Too Many False Positives

**Symptom:** Valid data flagged as invalid

**Solutions:**
1. Review validation rules
2. Make strict rules warnings instead of errors
3. Add exceptions for edge cases
4. Gather user feedback

---

## Conclusion

You now understand NEXS-MCP's validation system and how to extend it. Key takeaways:

1. **Multiple Layers**: Element.Validate() for structure, Validator for business logic
2. **Three Levels**: Basic (fast), Comprehensive (default), Strict (thorough)
3. **Clear Messages**: Specific, helpful, with suggestions
4. **Test Thoroughly**: Unit tests, edge cases, regressions
5. **Backward Compatible**: Don't break existing data

**Next Steps:**
- Review existing validators in `internal/validation/`
- Add validation rules for your element types
- Write comprehensive tests
- Update documentation

**Related Guides:**
- [CODE_TOUR.md](./CODE_TOUR.md) - Architecture overview
- [ADDING_ELEMENT_TYPE.md](./ADDING_ELEMENT_TYPE.md) - Creating elements
- [ADDING_MCP_TOOL.md](./ADDING_MCP_TOOL.md) - Creating tools

Developer tutorials created successfully
