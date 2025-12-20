# Adding a New Element Type: Complete Tutorial

## Table of Contents

- [Introduction](#introduction)
- [Overview](#overview)
- [Step 1: Define Domain Model](#step-1-define-domain-model)
- [Step 2: Add to ElementType Enum](#step-2-add-to-elementtype-enum)
- [Step 3: Implement Element Interface](#step-3-implement-element-interface)
- [Step 4: Create Validator](#step-4-create-validator)
- [Step 5: Update Repository](#step-5-update-repository)
- [Step 6: Add MCP Tools](#step-6-add-mcp-tools)
- [Step 7: Register Tools](#step-7-register-tools)
- [Step 8: Write Tests](#step-8-write-tests)
- [Step 9: Update Documentation](#step-9-update-documentation)
- [Complete Working Example: Workflow Element](#complete-working-example-workflow-element)
- [Common Pitfalls](#common-pitfalls)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

---

## Introduction

This tutorial guides you through adding a new element type to NEXS-MCP. We'll use a complete working example: adding a "Workflow" element type that represents automated multi-step processes.

**Prerequisites:**
- Go 1.25+ installed
- Familiarity with Go interfaces
- Understanding of NEXS-MCP architecture (see [CODE_TOUR.md](./CODE_TOUR.md))
- Official MCP Go SDK knowledge

**Time Required:** 2-4 hours for a complete implementation

---

## Overview

Adding a new element type involves these steps:

```
1. Define Domain Model ────────────> internal/domain/workflow.go
                                     
2. Add to ElementType Enum ───────> internal/domain/element.go
                                     
3. Implement Element Interface ───> workflow.go (methods)
                                     
4. Create Validator ──────────────> internal/validation/workflow_validator.go
                                     
5. Update Repository ─────────────> internal/infrastructure/*_repository.go
                                     
6. Add MCP Tools ─────────────────> internal/mcp/workflow_tools.go
                                     
7. Register Tools ────────────────> internal/mcp/server.go
                                     
8. Write Tests ───────────────────> *_test.go files
                                     
9. Update Documentation ──────────> docs/api/MCP_TOOLS.md
```

---

## Step 1: Define Domain Model

Create `internal/domain/workflow.go`:

```go
package domain

import (
    "fmt"
    "time"
)

// Workflow represents an automated multi-step process
type Workflow struct {
    Metadata ElementMetadata `json:"metadata"`
    
    // Workflow-specific fields
    Steps          []WorkflowStep    `json:"steps" validate:"required,min=1"`
    Triggers       []WorkflowTrigger `json:"triggers" validate:"required,min=1"`
    MaxDuration    int               `json:"max_duration" validate:"required,min=1"` // seconds
    RetryPolicy    *RetryPolicy      `json:"retry_policy,omitempty"`
    OnSuccess      *WorkflowAction   `json:"on_success,omitempty"`
    OnFailure      *WorkflowAction   `json:"on_failure,omitempty"`
    Variables      map[string]string `json:"variables,omitempty"`
    AccessControl  *AccessControl    `json:"access_control,omitempty"`
}

// WorkflowStep represents a single step in the workflow
type WorkflowStep struct {
    ID          string                 `json:"id" validate:"required"`
    Name        string                 `json:"name" validate:"required"`
    Type        string                 `json:"type" validate:"required,oneof=action condition loop parallel"`
    Action      string                 `json:"action,omitempty"` // Tool name or script
    Input       map[string]interface{} `json:"input,omitempty"`
    Condition   string                 `json:"condition,omitempty"` // For conditional steps
    Dependencies []string              `json:"dependencies,omitempty"` // Step IDs that must complete first
    Timeout     int                    `json:"timeout,omitempty"` // Step-specific timeout in seconds
    OnError     string                 `json:"on_error,omitempty" validate:"omitempty,oneof=fail retry skip"`
}

// WorkflowTrigger defines when the workflow should execute
type WorkflowTrigger struct {
    Type      string                 `json:"type" validate:"required,oneof=manual schedule event webhook"`
    Schedule  string                 `json:"schedule,omitempty"` // Cron expression
    Event     string                 `json:"event,omitempty"` // Event name
    Condition string                 `json:"condition,omitempty"`
    Config    map[string]interface{} `json:"config,omitempty"`
}

// RetryPolicy defines retry behavior for failed steps
type RetryPolicy struct {
    MaxAttempts  int   `json:"max_attempts" validate:"required,min=1,max=10"`
    InitialDelay int   `json:"initial_delay" validate:"required,min=1"` // seconds
    MaxDelay     int   `json:"max_delay" validate:"required,min=1"` // seconds
    Multiplier   float64 `json:"multiplier" validate:"required,min=1"`
}

// WorkflowAction defines an action to take on workflow completion
type WorkflowAction struct {
    Type   string                 `json:"type" validate:"required,oneof=notify webhook create_memory"`
    Target string                 `json:"target,omitempty"`
    Config map[string]interface{} `json:"config,omitempty"`
}

// NewWorkflow creates a new workflow with default values
func NewWorkflow(name, author string) *Workflow {
    now := time.Now()
    return &Workflow{
        Metadata: ElementMetadata{
            ID:          generateID("workflow"),
            Type:        WorkflowElement,
            Name:        name,
            Description: "",
            Version:     "1.0.0",
            Author:      author,
            Tags:        []string{},
            IsActive:    true,
            CreatedAt:   now,
            UpdatedAt:   now,
            Custom:      make(map[string]interface{}),
        },
        Steps:       []WorkflowStep{},
        Triggers:    []WorkflowTrigger{},
        MaxDuration: 3600, // 1 hour default
        Variables:   make(map[string]string),
    }
}

// GetMetadata returns the workflow's metadata
func (w *Workflow) GetMetadata() ElementMetadata {
    return w.Metadata
}

// GetType returns the element type
func (w *Workflow) GetType() ElementType {
    return WorkflowElement
}

// GetID returns the workflow ID
func (w *Workflow) GetID() string {
    return w.Metadata.ID
}

// IsActive returns whether the workflow is active
func (w *Workflow) IsActive() bool {
    return w.Metadata.IsActive
}

// Activate activates the workflow
func (w *Workflow) Activate() error {
    w.Metadata.IsActive = true
    w.Metadata.UpdatedAt = time.Now()
    return nil
}

// Deactivate deactivates the workflow
func (w *Workflow) Deactivate() error {
    w.Metadata.IsActive = false
    w.Metadata.UpdatedAt = time.Now()
    return nil
}

// Validate validates the workflow
func (w *Workflow) Validate() error {
    // Validate metadata
    if err := w.Metadata.Validate(); err != nil {
        return fmt.Errorf("metadata validation failed: %w", err)
    }

    // Validate steps
    if len(w.Steps) == 0 {
        return fmt.Errorf("workflow must have at least one step")
    }

    stepIDs := make(map[string]bool)
    for i, step := range w.Steps {
        if step.ID == "" {
            return fmt.Errorf("step %d: ID is required", i)
        }
        
        // Check for duplicate step IDs
        if stepIDs[step.ID] {
            return fmt.Errorf("step %d: duplicate step ID '%s'", i, step.ID)
        }
        stepIDs[step.ID] = true

        if step.Name == "" {
            return fmt.Errorf("step %d: name is required", i)
        }

        if step.Type == "" {
            return fmt.Errorf("step %d: type is required", i)
        }

        // Validate dependencies
        for _, depID := range step.Dependencies {
            if !stepIDs[depID] && depID != step.ID {
                // Note: this only checks previously defined steps
                // Full validation happens in the validator
            }
        }
    }

    // Validate triggers
    if len(w.Triggers) == 0 {
        return fmt.Errorf("workflow must have at least one trigger")
    }

    for i, trigger := range w.Triggers {
        if trigger.Type == "" {
            return fmt.Errorf("trigger %d: type is required", i)
        }
        
        // Validate trigger-specific fields
        switch trigger.Type {
        case "schedule":
            if trigger.Schedule == "" {
                return fmt.Errorf("trigger %d: schedule is required for schedule trigger", i)
            }
        case "event":
            if trigger.Event == "" {
                return fmt.Errorf("trigger %d: event is required for event trigger", i)
            }
        }
    }

    // Validate max duration
    if w.MaxDuration < 1 {
        return fmt.Errorf("max_duration must be at least 1 second")
    }

    // Validate retry policy if present
    if w.RetryPolicy != nil {
        if w.RetryPolicy.MaxAttempts < 1 || w.RetryPolicy.MaxAttempts > 10 {
            return fmt.Errorf("retry_policy: max_attempts must be between 1 and 10")
        }
        if w.RetryPolicy.InitialDelay < 1 {
            return fmt.Errorf("retry_policy: initial_delay must be at least 1 second")
        }
        if w.RetryPolicy.MaxDelay < w.RetryPolicy.InitialDelay {
            return fmt.Errorf("retry_policy: max_delay must be >= initial_delay")
        }
        if w.RetryPolicy.Multiplier < 1 {
            return fmt.Errorf("retry_policy: multiplier must be >= 1")
        }
    }

    return nil
}

// ToMap converts the workflow to a map for JSON serialization
func (w *Workflow) ToMap() map[string]interface{} {
    m := w.Metadata.ToMap()
    m["steps"] = w.Steps
    m["triggers"] = w.Triggers
    m["max_duration"] = w.MaxDuration
    m["retry_policy"] = w.RetryPolicy
    m["on_success"] = w.OnSuccess
    m["on_failure"] = w.OnFailure
    m["variables"] = w.Variables
    m["access_control"] = w.AccessControl
    return m
}

// Clone creates a deep copy of the workflow
func (w *Workflow) Clone() *Workflow {
    clone := *w
    clone.Metadata = w.Metadata
    clone.Metadata.ID = generateID("workflow")
    clone.Metadata.CreatedAt = time.Now()
    clone.Metadata.UpdatedAt = time.Now()
    
    // Deep copy slices and maps
    clone.Steps = make([]WorkflowStep, len(w.Steps))
    copy(clone.Steps, w.Steps)
    
    clone.Triggers = make([]WorkflowTrigger, len(w.Triggers))
    copy(clone.Triggers, w.Triggers)
    
    clone.Variables = make(map[string]string)
    for k, v := range w.Variables {
        clone.Variables[k] = v
    }
    
    return &clone
}
```

**Key Design Decisions:**

1. **Struct Tags**: Use `json` tags for serialization and `validate` tags for validation hints
2. **Pointers for Optional Fields**: Use pointers (e.g., `*RetryPolicy`) for optional complex types
3. **Validation**: Both basic validation in `Validate()` and comprehensive in validator
4. **Helper Methods**: Provide `ToMap()`, `Clone()`, etc. for common operations

---

## Step 2: Add to ElementType Enum

Update `internal/domain/element.go`:

```go
// ElementType represents the type of an element
type ElementType string

const (
    PersonaElement  ElementType = "persona"
    SkillElement    ElementType = "skill"
    TemplateElement ElementType = "template"
    AgentElement    ElementType = "agent"
    MemoryElement   ElementType = "memory"
    EnsembleElement ElementType = "ensemble"
    WorkflowElement ElementType = "workflow"  // ADD THIS LINE
)

// ValidateElementType checks if an element type is valid
func ValidateElementType(t ElementType) bool {
    switch t {
    case PersonaElement, SkillElement, TemplateElement,
         AgentElement, MemoryElement, EnsembleElement,
         WorkflowElement:  // ADD THIS CASE
        return true
    default:
        return false
    }
}

// AllElementTypes returns all valid element types
func AllElementTypes() []ElementType {
    return []ElementType{
        PersonaElement,
        SkillElement,
        TemplateElement,
        AgentElement,
        MemoryElement,
        EnsembleElement,
        WorkflowElement,  // ADD THIS LINE
    }
}
```

---

## Step 3: Implement Element Interface

The interface methods are already implemented in Step 1. Verify all required methods:

```go
type Element interface {
    GetMetadata() ElementMetadata  // ✓ Implemented
    Validate() error               // ✓ Implemented
    GetType() ElementType          // ✓ Implemented
    GetID() string                 // ✓ Implemented
    IsActive() bool                // ✓ Implemented
    Activate() error               // ✓ Implemented
    Deactivate() error             // ✓ Implemented
}
```

**Checklist:**
- [ ] All interface methods implemented
- [ ] Metadata properly initialized
- [ ] Validation covers all required fields
- [ ] Helper methods provided (ToMap, Clone, etc.)

---

## Step 4: Create Validator

Create `internal/validation/workflow_validator.go`:

```go
package validation

import (
    "fmt"
    "regexp"
    "strings"
    "time"

    "github.com/fsvxavier/nexs-mcp/internal/domain"
)

// WorkflowValidator validates workflow elements
type WorkflowValidator struct{}

// NewWorkflowValidator creates a new workflow validator
func NewWorkflowValidator() *WorkflowValidator {
    return &WorkflowValidator{}
}

// Validate validates a workflow element
func (v *WorkflowValidator) Validate(element domain.Element, level ValidationLevel) ValidationResult {
    workflow, ok := element.(*domain.Workflow)
    if !ok {
        return ValidationResult{
            IsValid:     false,
            ElementType: "workflow",
            Errors: []ValidationIssue{{
                Severity: ErrorSeverity,
                Field:    "type",
                Message:  "Element is not a workflow",
                Code:     "INVALID_TYPE",
            }},
        }
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

    // Basic validation: structure and required fields
    v.validateBasic(workflow, &result)

    // Comprehensive validation: business rules and relationships
    if level == ComprehensiveLevel || level == StrictLevel {
        v.validateComprehensive(workflow, &result)
    }

    // Strict validation: best practices and optimizations
    if level == StrictLevel {
        v.validateStrict(workflow, &result)
    }

    return result
}

// validateBasic performs basic structural validation
func (v *WorkflowValidator) validateBasic(workflow *domain.Workflow, result *ValidationResult) {
    // Validate metadata
    if err := workflow.Metadata.Validate(); err != nil {
        result.AddError("metadata", err.Error(), "METADATA_INVALID")
    }

    // Validate steps
    if len(workflow.Steps) == 0 {
        result.AddError("steps", "Workflow must have at least one step", "STEPS_REQUIRED")
    }

    stepIDs := make(map[string]bool)
    for i, step := range workflow.Steps {
        stepPrefix := fmt.Sprintf("steps[%d]", i)

        if step.ID == "" {
            result.AddError(fmt.Sprintf("%s.id", stepPrefix), "Step ID is required", "STEP_ID_REQUIRED")
        } else {
            if stepIDs[step.ID] {
                result.AddError(fmt.Sprintf("%s.id", stepPrefix), 
                    fmt.Sprintf("Duplicate step ID: %s", step.ID), "STEP_ID_DUPLICATE")
            }
            stepIDs[step.ID] = true
        }

        if step.Name == "" {
            result.AddError(fmt.Sprintf("%s.name", stepPrefix), "Step name is required", "STEP_NAME_REQUIRED")
        }

        if step.Type == "" {
            result.AddError(fmt.Sprintf("%s.type", stepPrefix), "Step type is required", "STEP_TYPE_REQUIRED")
        } else {
            validTypes := []string{"action", "condition", "loop", "parallel"}
            if !contains(validTypes, step.Type) {
                result.AddError(fmt.Sprintf("%s.type", stepPrefix), 
                    fmt.Sprintf("Invalid step type: %s (must be one of: %s)", 
                        step.Type, strings.Join(validTypes, ", ")), 
                    "STEP_TYPE_INVALID")
            }
        }

        // Validate action for action-type steps
        if step.Type == "action" && step.Action == "" {
            result.AddError(fmt.Sprintf("%s.action", stepPrefix), 
                "Action is required for action-type steps", "STEP_ACTION_REQUIRED")
        }

        // Validate condition for condition-type steps
        if step.Type == "condition" && step.Condition == "" {
            result.AddError(fmt.Sprintf("%s.condition", stepPrefix), 
                "Condition is required for condition-type steps", "STEP_CONDITION_REQUIRED")
        }
    }

    // Validate triggers
    if len(workflow.Triggers) == 0 {
        result.AddError("triggers", "Workflow must have at least one trigger", "TRIGGERS_REQUIRED")
    }

    for i, trigger := range workflow.Triggers {
        triggerPrefix := fmt.Sprintf("triggers[%d]", i)

        if trigger.Type == "" {
            result.AddError(fmt.Sprintf("%s.type", triggerPrefix), 
                "Trigger type is required", "TRIGGER_TYPE_REQUIRED")
        }

        switch trigger.Type {
        case "schedule":
            if trigger.Schedule == "" {
                result.AddError(fmt.Sprintf("%s.schedule", triggerPrefix), 
                    "Schedule is required for schedule trigger", "TRIGGER_SCHEDULE_REQUIRED")
            } else if !v.isValidCronExpression(trigger.Schedule) {
                result.AddWarning(fmt.Sprintf("%s.schedule", triggerPrefix), 
                    "Schedule may not be a valid cron expression", "TRIGGER_SCHEDULE_INVALID")
            }
        case "event":
            if trigger.Event == "" {
                result.AddError(fmt.Sprintf("%s.event", triggerPrefix), 
                    "Event is required for event trigger", "TRIGGER_EVENT_REQUIRED")
            }
        case "webhook":
            // Webhook-specific validation
            if trigger.Config == nil || trigger.Config["url"] == nil {
                result.AddWarning(fmt.Sprintf("%s.config", triggerPrefix), 
                    "Webhook trigger should have a URL in config", "TRIGGER_WEBHOOK_CONFIG")
            }
        }
    }

    // Validate max_duration
    if workflow.MaxDuration < 1 {
        result.AddError("max_duration", 
            "Maximum duration must be at least 1 second", "MAX_DURATION_INVALID")
    }
}

// validateComprehensive performs business rules and relationship validation
func (v *WorkflowValidator) validateComprehensive(workflow *domain.Workflow, result *ValidationResult) {
    // Build step ID map for dependency checking
    stepIDs := make(map[string]bool)
    for _, step := range workflow.Steps {
        stepIDs[step.ID] = true
    }

    // Validate step dependencies
    for i, step := range workflow.Steps {
        for j, depID := range step.Dependencies {
            if !stepIDs[depID] {
                result.AddError(fmt.Sprintf("steps[%d].dependencies[%d]", i, j), 
                    fmt.Sprintf("Referenced step '%s' does not exist", depID), 
                    "STEP_DEPENDENCY_NOT_FOUND")
            }

            // Check for circular dependencies (simplified check)
            if depID == step.ID {
                result.AddError(fmt.Sprintf("steps[%d].dependencies[%d]", i, j), 
                    "Step cannot depend on itself", "STEP_CIRCULAR_DEPENDENCY")
            }
        }
    }

    // Check for unreachable steps
    reachable := v.findReachableSteps(workflow)
    for i, step := range workflow.Steps {
        if !reachable[step.ID] {
            result.AddWarning(fmt.Sprintf("steps[%d]", i), 
                fmt.Sprintf("Step '%s' may be unreachable", step.Name), "STEP_UNREACHABLE")
        }
    }

    // Validate retry policy
    if workflow.RetryPolicy != nil {
        if workflow.RetryPolicy.MaxAttempts > 5 {
            result.AddWarning("retry_policy.max_attempts", 
                "High retry count may cause long execution times", "RETRY_HIGH_ATTEMPTS")
        }

        if workflow.RetryPolicy.MaxDelay > 300 {
            result.AddWarning("retry_policy.max_delay", 
                "Maximum delay > 5 minutes may cause timeout", "RETRY_HIGH_DELAY")
        }
    }

    // Validate total workflow timeout
    totalStepTimeout := 0
    for _, step := range workflow.Steps {
        if step.Timeout > 0 {
            totalStepTimeout += step.Timeout
        }
    }
    if totalStepTimeout > workflow.MaxDuration {
        result.AddWarning("max_duration", 
            "Sum of step timeouts exceeds workflow max duration", "WORKFLOW_TIMEOUT_INSUFFICIENT")
    }

    // Validate trigger combinations
    hasSchedule := false
    hasManual := false
    for _, trigger := range workflow.Triggers {
        if trigger.Type == "schedule" {
            hasSchedule = true
        }
        if trigger.Type == "manual" {
            hasManual = true
        }
    }
    if hasSchedule && !hasManual {
        result.AddInfo("triggers", 
            "Consider adding manual trigger for testing", "TRIGGER_MANUAL_RECOMMENDED")
    }
}

// validateStrict performs best practices validation
func (v *WorkflowValidator) validateStrict(workflow *domain.Workflow, result *ValidationResult) {
    // Check for meaningful names
    for i, step := range workflow.Steps {
        if len(step.Name) < 5 {
            result.AddWarning(fmt.Sprintf("steps[%d].name", i), 
                "Step name is very short, consider being more descriptive", "STEP_NAME_SHORT")
        }
    }

    // Check for error handling
    hasErrorHandling := false
    for _, step := range workflow.Steps {
        if step.OnError != "" {
            hasErrorHandling = true
            break
        }
    }
    if !hasErrorHandling && workflow.RetryPolicy == nil {
        result.AddWarning("error_handling", 
            "Consider adding error handling (step.on_error or retry_policy)", "ERROR_HANDLING_MISSING")
    }

    // Check for monitoring/notifications
    if workflow.OnSuccess == nil && workflow.OnFailure == nil {
        result.AddInfo("notifications", 
            "Consider adding on_success or on_failure actions for monitoring", "MONITORING_RECOMMENDED")
    }

    // Check for documentation
    if workflow.Metadata.Description == "" {
        result.AddInfo("metadata.description", 
            "Consider adding a description for documentation", "DESCRIPTION_MISSING")
    }

    // Check for reasonable complexity
    if len(workflow.Steps) > 20 {
        result.AddWarning("steps", 
            "Workflow has many steps, consider breaking into smaller workflows", "WORKFLOW_COMPLEX")
    }

    // Check for variable usage
    if len(workflow.Variables) == 0 {
        result.AddInfo("variables", 
            "Consider using variables for configuration values", "VARIABLES_RECOMMENDED")
    }
}

// Helper methods

func (v *WorkflowValidator) isValidCronExpression(expr string) bool {
    // Basic cron validation (5 or 6 fields)
    fields := strings.Fields(expr)
    return len(fields) == 5 || len(fields) == 6
}

func (v *WorkflowValidator) findReachableSteps(workflow *domain.Workflow) map[string]bool {
    reachable := make(map[string]bool)
    
    // For simplicity, assume all steps with no dependencies are entry points
    queue := []string{}
    for _, step := range workflow.Steps {
        if len(step.Dependencies) == 0 {
            queue = append(queue, step.ID)
        }
    }

    // BFS to find all reachable steps
    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        
        if reachable[current] {
            continue
        }
        reachable[current] = true

        // Add steps that depend on this one
        for _, step := range workflow.Steps {
            for _, depID := range step.Dependencies {
                if depID == current && !reachable[step.ID] {
                    queue = append(queue, step.ID)
                }
            }
        }
    }

    return reachable
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

**Validation Levels:**

1. **Basic**: Structure, required fields, data types
2. **Comprehensive**: Business rules, relationships, dependencies
3. **Strict**: Best practices, performance, documentation

---

## Step 5: Update Repository

The repository implementations are generic and should handle new element types automatically, but you may need to add helper methods.

Update `internal/infrastructure/file_repository.go`:

```go
// GetWorkflows returns all workflow elements
func (r *FileElementRepository) GetWorkflows() ([]*domain.Workflow, error) {
    elements, err := r.GetByType(domain.WorkflowElement)
    if err != nil {
        return nil, err
    }

    workflows := make([]*domain.Workflow, 0, len(elements))
    for _, elem := range elements {
        if workflow, ok := elem.(*domain.Workflow); ok {
            workflows = append(workflows, workflow)
        }
    }

    return workflows, nil
}
```

**Important:** The repository's serialization code needs to handle the new type. Check the `elementFromMap` function:

```go
func elementFromMap(data map[string]interface{}) (domain.Element, error) {
    metadata, err := metadataFromMap(data)
    if err != nil {
        return nil, err
    }

    switch metadata.Type {
    case domain.PersonaElement:
        return personaFromMap(data)
    case domain.SkillElement:
        return skillFromMap(data)
    // ... other cases ...
    case domain.WorkflowElement:
        return workflowFromMap(data)  // ADD THIS CASE
    default:
        return nil, fmt.Errorf("unknown element type: %s", metadata.Type)
    }
}

// ADD THIS FUNCTION
func workflowFromMap(data map[string]interface{}) (*domain.Workflow, error) {
    metadata, err := metadataFromMap(data)
    if err != nil {
        return nil, err
    }

    workflow := &domain.Workflow{
        Metadata: metadata,
    }

    // Parse steps
    if stepsData, ok := data["steps"].([]interface{}); ok {
        workflow.Steps = make([]domain.WorkflowStep, len(stepsData))
        for i, stepData := range stepsData {
            stepMap, ok := stepData.(map[string]interface{})
            if !ok {
                return nil, fmt.Errorf("invalid step data at index %d", i)
            }
            
            step := domain.WorkflowStep{
                ID:   getStringOrEmpty(stepMap, "id"),
                Name: getStringOrEmpty(stepMap, "name"),
                Type: getStringOrEmpty(stepMap, "type"),
                // ... map other fields ...
            }
            workflow.Steps[i] = step
        }
    }

    // Parse other fields similarly...

    return workflow, nil
}
```

---

## Step 6: Add MCP Tools

Create `internal/mcp/workflow_tools.go`:

```go
package mcp

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    sdk "github.com/modelcontextprotocol/go-sdk/mcp"

    "github.com/fsvxavier/nexs-mcp/internal/domain"
    "github.com/fsvxavier/nexs-mcp/internal/validation"
)

// --- Input/Output Structures ---

// CreateWorkflowInput defines input for create_workflow tool
type CreateWorkflowInput struct {
    Name          string                      `json:"name" jsonschema:"workflow name (3-100 characters)"`
    Description   string                      `json:"description,omitempty" jsonschema:"workflow description (max 500 characters)"`
    Version       string                      `json:"version" jsonschema:"workflow version (semver)"`
    Author        string                      `json:"author" jsonschema:"workflow author"`
    Tags          []string                    `json:"tags,omitempty" jsonschema:"workflow tags"`
    Steps         []domain.WorkflowStep       `json:"steps" jsonschema:"workflow steps (minimum 1)"`
    Triggers      []domain.WorkflowTrigger    `json:"triggers" jsonschema:"workflow triggers (minimum 1)"`
    MaxDuration   int                         `json:"max_duration" jsonschema:"maximum workflow duration in seconds"`
    RetryPolicy   *domain.RetryPolicy         `json:"retry_policy,omitempty" jsonschema:"retry policy for failed steps"`
    OnSuccess     *domain.WorkflowAction      `json:"on_success,omitempty" jsonschema:"action on workflow success"`
    OnFailure     *domain.WorkflowAction      `json:"on_failure,omitempty" jsonschema:"action on workflow failure"`
    Variables     map[string]string           `json:"variables,omitempty" jsonschema:"workflow variables"`
    IsActive      bool                        `json:"is_active,omitempty" jsonschema:"active status (default: true)"`
    User          string                      `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// QuickCreateWorkflowInput defines input for quick_create_workflow tool
type QuickCreateWorkflowInput struct {
    Name        string   `json:"name" jsonschema:"workflow name"`
    Author      string   `json:"author" jsonschema:"workflow author"`
    Description string   `json:"description,omitempty" jsonschema:"workflow description"`
    StepActions []string `json:"step_actions" jsonschema:"list of action names for sequential steps"`
    TriggerType string   `json:"trigger_type,omitempty" jsonschema:"trigger type (manual, schedule, event) default: manual"`
    Schedule    string   `json:"schedule,omitempty" jsonschema:"cron schedule (required if trigger_type=schedule)"`
    User        string   `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// --- Tool Handlers ---

// handleCreateWorkflow handles the create_workflow tool
func (s *MCPServer) handleCreateWorkflow(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    start := time.Now()
    var err error
    defer func() {
        duration := time.Since(start)
        s.metrics.RecordOperation("create_workflow", duration, err == nil)
        s.perfMetrics.Record(logger.PerformanceEntry{
            Timestamp: start,
            Operation: "create_workflow",
            Duration:  duration,
            Success:   err == nil,
        })
    }()

    // Parse input
    var input CreateWorkflowInput
    if err = parseArguments(arguments, &input); err != nil {
        return errorResult("Invalid arguments: %v", err), nil
    }

    // Create workflow
    now := time.Now()
    workflow := &domain.Workflow{
        Metadata: domain.ElementMetadata{
            ID:          fmt.Sprintf("workflow-%s", uuid.New().String()[:8]),
            Type:        domain.WorkflowElement,
            Name:        input.Name,
            Description: input.Description,
            Version:     input.Version,
            Author:      input.Author,
            Tags:        input.Tags,
            IsActive:    input.IsActive,
            CreatedAt:   now,
            UpdatedAt:   now,
            Custom:      make(map[string]interface{}),
        },
        Steps:       input.Steps,
        Triggers:    input.Triggers,
        MaxDuration: input.MaxDuration,
        RetryPolicy: input.RetryPolicy,
        OnSuccess:   input.OnSuccess,
        OnFailure:   input.OnFailure,
        Variables:   input.Variables,
    }

    // Set default active status
    if !input.IsActive {
        workflow.Metadata.IsActive = true
    }

    // Validate workflow
    if err = workflow.Validate(); err != nil {
        return errorResult("Workflow validation failed: %v", err), nil
    }

    // Comprehensive validation
    validator := validation.NewWorkflowValidator()
    validationResult := validator.Validate(workflow, validation.ComprehensiveLevel)
    if !validationResult.IsValid {
        errors := make([]string, len(validationResult.Errors))
        for i, e := range validationResult.Errors {
            errors[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
        }
        return errorResult("Workflow validation failed:\n- %s", 
            strings.Join(errors, "\n- ")), nil
    }

    // Set access control if user provided
    if input.User != "" {
        workflow.AccessControl = &domain.AccessControl{
            Owner:        input.User,
            Permissions:  []string{"read", "write", "delete"},
            AllowedUsers: []string{input.User},
        }
    }

    // Store workflow
    if err = s.repo.Create(workflow); err != nil {
        return errorResult("Failed to create workflow: %v", err), nil
    }

    // Update index
    content := fmt.Sprintf("%s %s %s", workflow.Metadata.Name, 
        workflow.Metadata.Description, strings.Join(workflow.Metadata.Tags, " "))
    s.index.AddDocument(workflow.GetID(), content, "workflow")

    // Log creation
    logger.Info("Workflow created",
        "id", workflow.GetID(),
        "name", workflow.Metadata.Name,
        "steps", len(workflow.Steps),
        "triggers", len(workflow.Triggers))

    // Return success
    return successResult(CreateElementOutput{
        ID:      workflow.GetID(),
        Element: workflow.ToMap(),
    })
}

// handleQuickCreateWorkflow handles the quick_create_workflow tool
func (s *MCPServer) handleQuickCreateWorkflow(ctx context.Context, arguments interface{}) (*sdk.CallToolResult, error) {
    start := time.Now()
    var err error
    defer func() {
        duration := time.Since(start)
        s.metrics.RecordOperation("quick_create_workflow", duration, err == nil)
    }()

    // Parse input
    var input QuickCreateWorkflowInput
    if err = parseArguments(arguments, &input); err != nil {
        return errorResult("Invalid arguments: %v", err), nil
    }

    // Validate required fields
    if input.Name == "" {
        return errorResult("Name is required"), nil
    }
    if input.Author == "" {
        return errorResult("Author is required"), nil
    }
    if len(input.StepActions) == 0 {
        return errorResult("At least one step action is required"), nil
    }

    // Build steps
    steps := make([]domain.WorkflowStep, len(input.StepActions))
    for i, action := range input.StepActions {
        steps[i] = domain.WorkflowStep{
            ID:      fmt.Sprintf("step-%d", i+1),
            Name:    fmt.Sprintf("Step %d: %s", i+1, action),
            Type:    "action",
            Action:  action,
            Timeout: 300, // 5 minutes default
            OnError: "fail",
        }
        
        // Add dependency on previous step (sequential execution)
        if i > 0 {
            steps[i].Dependencies = []string{steps[i-1].ID}
        }
    }

    // Build trigger
    triggerType := input.TriggerType
    if triggerType == "" {
        triggerType = "manual"
    }

    trigger := domain.WorkflowTrigger{
        Type: triggerType,
    }

    if triggerType == "schedule" {
        if input.Schedule == "" {
            return errorResult("Schedule is required for schedule trigger"), nil
        }
        trigger.Schedule = input.Schedule
    }

    // Create full input for standard handler
    createInput := CreateWorkflowInput{
        Name:        input.Name,
        Description: input.Description,
        Version:     "1.0.0",
        Author:      input.Author,
        Tags:        []string{"auto-generated"},
        Steps:       steps,
        Triggers:    []domain.WorkflowTrigger{trigger},
        MaxDuration: 3600, // 1 hour
        IsActive:    true,
        User:        input.User,
    }

    // Delegate to standard create handler
    return s.handleCreateWorkflow(ctx, createInput)
}
```

---

## Step 7: Register Tools

Update `internal/mcp/server.go` in the `registerTools()` method:

```go
func (s *MCPServer) registerTools() {
    // ... existing tool registrations ...

    // Workflow tools
    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "create_workflow",
        Description: "Create a new workflow element with full configuration",
    }, s.handleCreateWorkflow)

    sdk.AddTool(s.server, &sdk.Tool{
        Name:        "quick_create_workflow",
        Description: "Quickly create a workflow with sequential steps",
    }, s.handleQuickCreateWorkflow)

    // ... rest of tools ...
}
```

**Don't forget:** The generic tools (`list_elements`, `get_element`, `update_element`, `delete_element`) already support workflows!

---

## Step 8: Write Tests

Create `internal/domain/workflow_test.go`:

```go
package domain

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestWorkflow_Validate(t *testing.T) {
    tests := []struct {
        name        string
        workflow    *Workflow
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid workflow",
            workflow: &Workflow{
                Metadata: validWorkflowMetadata(),
                Steps: []WorkflowStep{
                    {ID: "step1", Name: "First Step", Type: "action", Action: "test_action"},
                },
                Triggers: []WorkflowTrigger{
                    {Type: "manual"},
                },
                MaxDuration: 3600,
            },
            expectError: false,
        },
        {
            name: "missing steps",
            workflow: &Workflow{
                Metadata:    validWorkflowMetadata(),
                Steps:       []WorkflowStep{},
                Triggers:    []WorkflowTrigger{{Type: "manual"}},
                MaxDuration: 3600,
            },
            expectError: true,
            errorMsg:    "at least one step",
        },
        {
            name: "missing triggers",
            workflow: &Workflow{
                Metadata:    validWorkflowMetadata(),
                Steps:       []WorkflowStep{{ID: "step1", Name: "Step", Type: "action"}},
                Triggers:    []WorkflowTrigger{},
                MaxDuration: 3600,
            },
            expectError: true,
            errorMsg:    "at least one trigger",
        },
        {
            name: "duplicate step IDs",
            workflow: &Workflow{
                Metadata: validWorkflowMetadata(),
                Steps: []WorkflowStep{
                    {ID: "step1", Name: "First", Type: "action"},
                    {ID: "step1", Name: "Second", Type: "action"},
                },
                Triggers:    []WorkflowTrigger{{Type: "manual"}},
                MaxDuration: 3600,
            },
            expectError: true,
            errorMsg:    "duplicate",
        },
        {
            name: "schedule trigger without schedule",
            workflow: &Workflow{
                Metadata:    validWorkflowMetadata(),
                Steps:       []WorkflowStep{{ID: "step1", Name: "Step", Type: "action"}},
                Triggers:    []WorkflowTrigger{{Type: "schedule"}},
                MaxDuration: 3600,
            },
            expectError: true,
            errorMsg:    "schedule is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.workflow.Validate()
            if tt.expectError {
                assert.Error(t, err)
                if tt.errorMsg != "" {
                    assert.Contains(t, err.Error(), tt.errorMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestWorkflow_ToMap(t *testing.T) {
    workflow := &Workflow{
        Metadata: validWorkflowMetadata(),
        Steps: []WorkflowStep{
            {ID: "step1", Name: "Test Step", Type: "action", Action: "test"},
        },
        Triggers: []WorkflowTrigger{
            {Type: "manual"},
        },
        MaxDuration: 3600,
        Variables:   map[string]string{"key": "value"},
    }

    m := workflow.ToMap()
    
    assert.Equal(t, workflow.Metadata.ID, m["id"])
    assert.Equal(t, "workflow", m["type"])
    assert.Equal(t, workflow.Metadata.Name, m["name"])
    assert.NotNil(t, m["steps"])
    assert.NotNil(t, m["triggers"])
    assert.Equal(t, 3600, m["max_duration"])
}

func TestWorkflow_Clone(t *testing.T) {
    original := &Workflow{
        Metadata: validWorkflowMetadata(),
        Steps: []WorkflowStep{
            {ID: "step1", Name: "Step 1", Type: "action"},
        },
        Triggers: []WorkflowTrigger{
            {Type: "manual"},
        },
        MaxDuration: 3600,
        Variables:   map[string]string{"key": "value"},
    }

    clone := original.Clone()

    // IDs should be different
    assert.NotEqual(t, original.GetID(), clone.GetID())
    
    // Other fields should match
    assert.Equal(t, original.Metadata.Name, clone.Metadata.Name)
    assert.Equal(t, len(original.Steps), len(clone.Steps))
    assert.Equal(t, len(original.Triggers), len(clone.Triggers))
    
    // Modifying clone shouldn't affect original
    clone.Variables["key2"] = "value2"
    assert.NotContains(t, original.Variables, "key2")
}

func validWorkflowMetadata() ElementMetadata {
    return ElementMetadata{
        ID:          "workflow-test-001",
        Type:        WorkflowElement,
        Name:        "Test Workflow",
        Description: "A test workflow",
        Version:     "1.0.0",
        Author:      "test@example.com",
        Tags:        []string{"test"},
        IsActive:    true,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
        Custom:      make(map[string]interface{}),
    }
}
```

Create `internal/validation/workflow_validator_test.go`:

```go
package validation

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    
    "github.com/fsvxavier/nexs-mcp/internal/domain"
)

func TestWorkflowValidator_Validate(t *testing.T) {
    validator := NewWorkflowValidator()

    tests := []struct {
        name           string
        workflow       *domain.Workflow
        level          ValidationLevel
        expectValid    bool
        expectErrors   int
        expectWarnings int
    }{
        {
            name:         "valid simple workflow",
            workflow:     createValidWorkflow(),
            level:        BasicLevel,
            expectValid:  true,
            expectErrors: 0,
        },
        {
            name:         "missing steps",
            workflow:     createWorkflowWithoutSteps(),
            level:        BasicLevel,
            expectValid:  false,
            expectErrors: 1,
        },
        {
            name:           "unreachable step warning",
            workflow:       createWorkflowWithUnreachableStep(),
            level:          ComprehensiveLevel,
            expectValid:    true,
            expectWarnings: 1,
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
        })
    }
}

func createValidWorkflow() *domain.Workflow {
    return &domain.Workflow{
        Metadata: domain.ElementMetadata{
            ID:          "workflow-001",
            Type:        domain.WorkflowElement,
            Name:        "Test Workflow",
            Description: "Test",
            Version:     "1.0.0",
            Author:      "test@example.com",
            IsActive:    true,
            CreatedAt:   time.Now(),
            UpdatedAt:   time.Now(),
        },
        Steps: []domain.WorkflowStep{
            {ID: "step1", Name: "First Step", Type: "action", Action: "test"},
        },
        Triggers: []domain.WorkflowTrigger{
            {Type: "manual"},
        },
        MaxDuration: 3600,
    }
}

// More test helpers...
```

Create `internal/mcp/workflow_tools_test.go`:

```go
package mcp

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/fsvxavier/nexs-mcp/internal/config"
    "github.com/fsvxavier/nexs-mcp/internal/domain"
    "github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

func TestHandleCreateWorkflow(t *testing.T) {
    // Setup
    repo := infrastructure.NewInMemoryElementRepository()
    cfg := &config.Config{}
    server := NewMCPServer("test", "1.0.0", repo, cfg)

    // Test input
    input := CreateWorkflowInput{
        Name:    "Test Workflow",
        Version: "1.0.0",
        Author:  "test@example.com",
        Steps: []domain.WorkflowStep{
            {
                ID:     "step1",
                Name:   "Test Step",
                Type:   "action",
                Action: "test_action",
            },
        },
        Triggers: []domain.WorkflowTrigger{
            {Type: "manual"},
        },
        MaxDuration: 3600,
        IsActive:    true,
    }

    // Execute
    result, err := server.handleCreateWorkflow(context.Background(), input)

    // Assert
    require.NoError(t, err)
    assert.False(t, result.IsError)
    
    // Verify workflow was created
    workflows, err := repo.GetByType(domain.WorkflowElement)
    require.NoError(t, err)
    assert.Equal(t, 1, len(workflows))
}

func TestHandleQuickCreateWorkflow(t *testing.T) {
    // Setup
    repo := infrastructure.NewInMemoryElementRepository()
    cfg := &config.Config{}
    server := NewMCPServer("test", "1.0.0", repo, cfg)

    // Test input
    input := QuickCreateWorkflowInput{
        Name:        "Quick Workflow",
        Author:      "test@example.com",
        StepActions: []string{"action1", "action2", "action3"},
        TriggerType: "manual",
    }

    // Execute
    result, err := server.handleQuickCreateWorkflow(context.Background(), input)

    // Assert
    require.NoError(t, err)
    assert.False(t, result.IsError)
    
    // Verify workflow has 3 steps
    workflows, err := repo.GetByType(domain.WorkflowElement)
    require.NoError(t, err)
    require.Equal(t, 1, len(workflows))
    
    workflow := workflows[0].(*domain.Workflow)
    assert.Equal(t, 3, len(workflow.Steps))
    assert.Equal(t, "action1", workflow.Steps[0].Action)
}
```

---

## Step 9: Update Documentation

Update `docs/api/MCP_TOOLS.md`:

```markdown
### create_workflow

Create a new workflow element with full configuration.

**Input:**
```json
{
  "name": "Data Processing Workflow",
  "description": "Automated data processing pipeline",
  "version": "1.0.0",
  "author": "team@example.com",
  "tags": ["automation", "data"],
  "steps": [
    {
      "id": "step1",
      "name": "Fetch Data",
      "type": "action",
      "action": "fetch_data",
      "timeout": 300
    },
    {
      "id": "step2",
      "name": "Process Data",
      "type": "action",
      "action": "process_data",
      "dependencies": ["step1"],
      "timeout": 600
    }
  ],
  "triggers": [
    {
      "type": "schedule",
      "schedule": "0 0 * * *"
    }
  ],
  "max_duration": 3600,
  "retry_policy": {
    "max_attempts": 3,
    "initial_delay": 10,
    "max_delay": 300,
    "multiplier": 2.0
  }
}
```

**Output:**
```json
{
  "id": "workflow-abc123",
  "element": { ... }
}
```

### quick_create_workflow

Quickly create a workflow with sequential steps.

**Input:**
```json
{
  "name": "Simple Workflow",
  "author": "user@example.com",
  "description": "A simple sequential workflow",
  "step_actions": ["action1", "action2", "action3"],
  "trigger_type": "manual"
}
```
```

---

## Complete Working Example: Workflow Element

This example shows the complete implementation. All code above combines to create a fully functional workflow element type.

**Test the implementation:**

```bash
# Run tests
go test ./internal/domain -v -run TestWorkflow
go test ./internal/validation -v -run TestWorkflowValidator
go test ./internal/mcp -v -run TestHandleCreateWorkflow

# Build
go build -o nexs-mcp ./cmd/nexs-mcp

# Test with MCP client
echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "quick_create_workflow",
    "arguments": {
      "name": "Test Workflow",
      "author": "test@example.com",
      "step_actions": ["step1", "step2"],
      "trigger_type": "manual"
    }
  }
}' | ./nexs-mcp
```

---

## Common Pitfalls

### 1. Forgetting to Update ElementType Enum

**Symptom:** "unknown element type" errors

**Solution:** Always add new type to:
- `const` block
- `ValidateElementType()` function
- `AllElementTypes()` function

### 2. Not Implementing All Interface Methods

**Symptom:** Compilation errors

**Solution:** Verify all `Element` interface methods are implemented:
```bash
# Check interface compliance
go build ./internal/domain
```

### 3. Repository Serialization Issues

**Symptom:** Elements not persisting correctly

**Solution:** 
- Add case in `elementFromMap()`
- Implement `workflowFromMap()` function
- Add case in `elementToMap()`

### 4. Missing Validator Registration

**Symptom:** Validation not working

**Solution:** Validators are created per-use, no registration needed. Just ensure validator implements `Validator` interface.

### 5. Tool Registration Forgotten

**Symptom:** Tools not available to clients

**Solution:** Always add `sdk.AddTool()` call in `registerTools()` method.

### 6. Incomplete Test Coverage

**Symptom:** Bugs slip through

**Solution:** Test at multiple levels:
- Unit tests for domain model
- Validation tests
- Integration tests for tools
- End-to-end tests

### 7. Validation Too Strict in Basic Level

**Symptom:** Users can't create elements easily

**Solution:** Keep basic validation minimal. Put business rules in comprehensive level.

### 8. Not Handling Backward Compatibility

**Symptom:** Existing elements fail to load

**Solution:** Use optional fields and provide defaults when loading old data.

---

## Troubleshooting

### Error: "Element with ID X already exists"

**Cause:** Duplicate ID generation or cached element

**Fix:**
```go
// Ensure unique ID generation
id := fmt.Sprintf("workflow-%s", uuid.New().String()[:8])
```

### Error: "Failed to unmarshal file"

**Cause:** YAML structure mismatch

**Fix:** 
1. Check YAML tags match struct fields
2. Verify `workflowFromMap()` handles all fields
3. Test serialization roundtrip

### Validation Errors Not Showing

**Cause:** Validator not returning result correctly

**Fix:**
```go
// Always set IsValid = false when adding errors
result.AddError(field, message, code)  // This sets IsValid = false internally
```

### Tool Returns Empty Response

**Cause:** Not using `successResult()` helper

**Fix:**
```go
// Always return via helper
return successResult(CreateElementOutput{
    ID: workflow.GetID(),
    Element: workflow.ToMap(),
})
```

---

## Best Practices

### 1. Start Simple, Add Complexity Incrementally

```go
// Phase 1: Basic structure
type Workflow struct {
    Metadata ElementMetadata
    Steps    []WorkflowStep
}

// Phase 2: Add features
type Workflow struct {
    Metadata    ElementMetadata
    Steps       []WorkflowStep
    Triggers    []WorkflowTrigger  // New
    MaxDuration int                // New
}
```

### 2. Use Validation Levels Appropriately

- **Basic:** Structure only (fast)
- **Comprehensive:** Business logic (default for tools)
- **Strict:** Best practices (optional user request)

### 3. Provide Helper Methods

```go
// Good: Convenient constructors
func NewWorkflow(name, author string) *Workflow { ... }

// Good: Common operations
func (w *Workflow) Clone() *Workflow { ... }
func (w *Workflow) ToMap() map[string]interface{} { ... }
```

### 4. Write Descriptive Error Messages

```go
// Bad
return fmt.Errorf("invalid step")

// Good
return fmt.Errorf("step %d: type '%s' is invalid, must be one of: action, condition, loop, parallel", 
    i, step.Type)
```

### 5. Test Edge Cases

```go
// Test with:
- Empty/nil values
- Boundary values (0, 1, max)
- Invalid combinations
- Circular dependencies
- Very large inputs
```

### 6. Document Your Design Decisions

```go
// WorkflowStep uses string IDs instead of numeric indices because:
// 1. More readable in logs and errors
// 2. Stable across insertions/deletions
// 3. Can be referenced by name in conditions
type WorkflowStep struct {
    ID string `json:"id"`
    // ...
}
```

### 7. Keep Related Code Together

```
internal/domain/
  workflow.go           # Domain model
  workflow_test.go      # Domain tests

internal/validation/
  workflow_validator.go     # Validation logic
  workflow_validator_test.go # Validation tests

internal/mcp/
  workflow_tools.go     # MCP tools
  workflow_tools_test.go # Tool tests
```

### 8. Use Constants for Magic Strings

```go
const (
    StepTypeAction    = "action"
    StepTypeCondition = "condition"
    StepTypeLoop      = "loop"
    StepTypeParallel  = "parallel"
)
```

---

## Conclusion

You've successfully added a new element type! Key takeaways:

1. **Follow the pattern**: Domain model → Interface → Validation → Tools
2. **Test thoroughly**: Unit, integration, and end-to-end tests
3. **Validate at multiple levels**: Basic, comprehensive, strict
4. **Use the MCP SDK**: `sdk.AddTool()` for registration
5. **Document well**: Update API docs and examples

**Next Steps:**
- Read [ADDING_MCP_TOOL.md](./ADDING_MCP_TOOL.md) for adding custom tools
- Read [EXTENDING_VALIDATION.md](./EXTENDING_VALIDATION.md) for advanced validation

**Need Help?**
- Check [CODE_TOUR.md](./CODE_TOUR.md) for architecture overview
- Look at existing element types for examples
- Run tests to verify your implementation
