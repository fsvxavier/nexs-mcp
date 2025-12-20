# NEXS MCP Domain Layer

**Version:** 1.0.0  
**Last Updated:** December 20, 2025  
**Status:** Production

---

## Table of Contents

- [Introduction](#introduction)
- [Domain-Driven Design Principles](#domain-driven-design-principles)
- [Element Type System](#element-type-system)
- [Element Metadata](#element-metadata)
- [Persona Element](#persona-element)
- [Skill Element](#skill-element)
- [Template Element](#template-element)
- [Agent Element](#agent-element)
- [Memory Element](#memory-element)
- [Ensemble Element](#ensemble-element)
- [Repository Interface](#repository-interface)
- [Domain Rules & Invariants](#domain-rules--invariants)
- [Validation System](#validation-system)
- [Access Control](#access-control)
- [Domain Events](#domain-events)
- [Best Practices](#best-practices)

---

## Introduction

The **Domain Layer** is the heart of NEXS MCP Server. It contains all business logic, rules, and entities with **zero external dependencies**. This layer defines WHAT the system does, not HOW it does it.

### Domain Layer Location

```
internal/domain/
├── element.go              # Base element types and interfaces
├── persona.go             # Persona entity
├── skill.go               # Skill entity
├── template.go            # Template entity
├── agent.go               # Agent entity
├── memory.go              # Memory entity
├── ensemble.go            # Ensemble entity
├── access_control.go      # Access control logic
└── *_test.go              # Unit tests (79.2% coverage)
```

### Domain Purity

**Zero External Dependencies:**

```go
package domain

import (
    "errors"      // ✅ Standard library
    "fmt"         // ✅ Standard library
    "time"        // ✅ Standard library
    "crypto/sha256" // ✅ Standard library
    
    // ❌ NO external packages
    // ❌ NO infrastructure imports
    // ❌ NO framework dependencies
)
```

This purity enables:
- **Unit Testing** - Test without infrastructure
- **Portability** - Run anywhere Go runs
- **Longevity** - Business logic outlives frameworks
- **Clarity** - Focus on business rules

---

## Domain-Driven Design Principles

### Ubiquitous Language

The domain uses consistent terminology across the entire system:

| Term | Definition | Example |
|------|------------|---------|
| **Element** | Any manageable AI component | Persona, Skill, Agent |
| **Persona** | AI personality and behavior profile | "Technical Expert", "Creative Writer" |
| **Skill** | Specialized capability or procedure | "Code Review", "Data Analysis" |
| **Template** | Reusable content structure | Prompt template, response format |
| **Agent** | Autonomous task executor | Task planner, decision maker |
| **Memory** | Persistent context storage | Conversation history, facts |
| **Ensemble** | Multi-agent orchestration | Sequential pipeline, parallel team |
| **Repository** | Storage abstraction | File system, database |
| **Metadata** | Common element attributes | ID, name, version, tags |

### Entities vs Value Objects

**Entities** (Have Identity):
- Persona, Skill, Template, Agent, Memory, Ensemble
- Identified by unique ID
- Mutable over time
- Equality based on ID

**Value Objects** (No Identity):
- ElementMetadata
- BehavioralTrait
- ExpertiseArea
- ResponseStyle
- SkillProcedure
- Equality based on attributes

```go
// Entity - has identity
type Persona struct {
    metadata ElementMetadata  // Contains ID
    // ...
}

// Value Object - no identity
type BehavioralTrait struct {
    Name        string
    Description string
    Intensity   int
}
```

### Aggregates

Each element type is an **Aggregate Root**:

```
Persona (Aggregate Root)
├── ElementMetadata
├── []BehavioralTrait (Value Objects)
├── []ExpertiseArea (Value Objects)
└── ResponseStyle (Value Object)
```

**Aggregate Rules:**
1. External objects can only hold references to the aggregate root
2. Modifications go through the aggregate root
3. Invariants are maintained by the aggregate root

---

## Element Type System

### Element Hierarchy

```
                    ┌─────────────────┐
                    │    Element      │
                    │   (Interface)   │
                    └────────┬────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
    ┌───────▼──────┐  ┌─────▼─────┐  ┌──────▼──────┐
    │   Persona    │  │   Skill   │  │  Template   │
    └──────────────┘  └───────────┘  └─────────────┘
            │                │                │
    ┌───────▼──────┐  ┌─────▼─────┐  ┌──────▼──────┐
    │    Agent     │  │   Memory  │  │  Ensemble   │
    └──────────────┘  └───────────┘  └─────────────┘
```

### Element Interface

All elements implement this interface:

```go
// Element is the base interface for all element types
type Element interface {
    // GetMetadata returns the element's metadata
    GetMetadata() ElementMetadata

    // Validate checks if the element is valid
    Validate() error

    // GetType returns the element type
    GetType() ElementType

    // GetID returns the element ID
    GetID() string

    // IsActive returns whether the element is active
    IsActive() bool

    // Activate activates the element
    Activate() error

    // Deactivate deactivates the element
    Deactivate() error
}
```

### Element Types

```go
type ElementType string

const (
    PersonaElement  ElementType = "persona"
    SkillElement    ElementType = "skill"
    TemplateElement ElementType = "template"
    AgentElement    ElementType = "agent"
    MemoryElement   ElementType = "memory"
    EnsembleElement ElementType = "ensemble"
)

func ValidateElementType(t ElementType) bool {
    switch t {
    case PersonaElement, SkillElement, TemplateElement,
         AgentElement, MemoryElement, EnsembleElement:
        return true
    default:
        return false
    }
}
```

### Common Domain Errors

```go
var (
    ErrInvalidElementType = errors.New("invalid element type")
    ErrInvalidElementID   = errors.New("invalid element ID")
    ErrElementNotFound    = errors.New("element not found")
    ErrValidationFailed   = errors.New("validation failed")
)
```

---

## Element Metadata

All elements share common metadata:

```go
// ElementMetadata contains common metadata for all elements
type ElementMetadata struct {
    ID          string                 `json:"id" validate:"required"`
    Type        ElementType            `json:"type" validate:"required,oneof=persona skill template agent memory ensemble"`
    Name        string                 `json:"name" validate:"required,min=3,max=100"`
    Description string                 `json:"description" validate:"max=500"`
    Version     string                 `json:"version" validate:"required,semver"`
    Author      string                 `json:"author" validate:"required"`
    Tags        []string               `json:"tags,omitempty"`
    IsActive    bool                   `json:"is_active"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
    Custom      map[string]interface{} `json:"custom,omitempty"`
}
```

### Field Descriptions

| Field | Type | Required | Validation | Purpose |
|-------|------|----------|------------|---------|
| **ID** | string | Yes | Unique identifier | Element identification |
| **Type** | ElementType | Yes | One of 6 types | Element classification |
| **Name** | string | Yes | 3-100 chars | Human-readable name |
| **Description** | string | No | Max 500 chars | Element purpose |
| **Version** | string | Yes | Semantic version | Version tracking |
| **Author** | string | Yes | - | Creator identification |
| **Tags** | []string | No | - | Categorization |
| **IsActive** | bool | Yes | - | Activation state |
| **CreatedAt** | time.Time | Yes | - | Creation timestamp |
| **UpdatedAt** | time.Time | Yes | - | Last modification |
| **Custom** | map | No | - | Extension point |

### Metadata Validation

```go
func (m ElementMetadata) Validate() error {
    if m.ID == "" {
        return ErrInvalidElementID
    }
    if !ValidateElementType(m.Type) {
        return ErrInvalidElementType
    }
    if len(m.Name) < 3 || len(m.Name) > 100 {
        return fmt.Errorf("name must be between 3 and 100 characters")
    }
    if len(m.Description) > 500 {
        return fmt.Errorf("description must not exceed 500 characters")
    }
    if m.Version == "" {
        return fmt.Errorf("version is required")
    }
    if m.Author == "" {
        return fmt.Errorf("author is required")
    }
    return nil
}
```

### ID Generation

```go
func GenerateElementID(elementType ElementType, name string) string {
    // Normalize name: lowercase, replace spaces with hyphens
    normalized := strings.ToLower(strings.TrimSpace(name))
    normalized = strings.ReplaceAll(normalized, " ", "-")
    
    // Remove special characters
    normalized = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(normalized, "")
    
    // Generate ID: type-name-timestamp
    timestamp := time.Now().Unix()
    return fmt.Sprintf("%s-%s-%d", elementType, normalized, timestamp)
}

// Example: "persona-technical-expert-1703088000"
```

---

## Persona Element

### Definition

A **Persona** represents an AI personality profile with behavioral traits, expertise areas, and response style.

### Structure

```go
type Persona struct {
    metadata         ElementMetadata
    BehavioralTraits []BehavioralTrait   `json:"behavioral_traits" yaml:"behavioral_traits" validate:"required,min=1,dive"`
    ExpertiseAreas   []ExpertiseArea     `json:"expertise_areas" yaml:"expertise_areas" validate:"required,min=1,dive"`
    ResponseStyle    ResponseStyle       `json:"response_style" yaml:"response_style" validate:"required"`
    SystemPrompt     string              `json:"system_prompt" yaml:"system_prompt" validate:"required,min=10,max=2000"`
    PrivacyLevel     PersonaPrivacyLevel `json:"privacy_level" yaml:"privacy_level" validate:"required,oneof=public private shared"`
    Owner            string              `json:"owner,omitempty" yaml:"owner,omitempty"`
    SharedWith       []string            `json:"shared_with,omitempty" yaml:"shared_with,omitempty"`
    HotSwappable     bool                `json:"hot_swappable" yaml:"hot_swappable"`
}
```

### Privacy Levels

```go
type PersonaPrivacyLevel string

const (
    PrivacyPublic  PersonaPrivacyLevel = "public"   // Visible to all
    PrivacyPrivate PersonaPrivacyLevel = "private"  // Visible to owner only
    PrivacyShared  PersonaPrivacyLevel = "shared"   // Visible to owner + shared_with
)
```

### Behavioral Traits

```go
type BehavioralTrait struct {
    Name        string `json:"name" yaml:"name" validate:"required,min=2,max=50"`
    Description string `json:"description,omitempty" yaml:"description,omitempty"`
    Intensity   int    `json:"intensity" yaml:"intensity" validate:"min=1,max=10"`
}
```

**Examples:**
```yaml
behavioral_traits:
  - name: "Analytical"
    description: "Systematic problem-solving approach"
    intensity: 9
  - name: "Patient"
    description: "Takes time to explain concepts"
    intensity: 8
  - name: "Detail-Oriented"
    description: "Focuses on precision and accuracy"
    intensity: 10
```

### Expertise Areas

```go
type ExpertiseArea struct {
    Domain      string   `json:"domain" yaml:"domain" validate:"required,min=2,max=100"`
    Level       string   `json:"level" yaml:"level" validate:"required,oneof=beginner intermediate advanced expert"`
    Keywords    []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`
    Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}
```

**Examples:**
```yaml
expertise_areas:
  - domain: "Software Architecture"
    level: "expert"
    keywords: ["clean architecture", "DDD", "microservices"]
    description: "Design and implementation of scalable systems"
    
  - domain: "Go Programming"
    level: "expert"
    keywords: ["concurrency", "performance", "best practices"]
    description: "Expert-level Go development"
```

### Response Style

```go
type ResponseStyle struct {
    Tone            string   `json:"tone" yaml:"tone" validate:"required,min=2,max=50"`
    Formality       string   `json:"formality" yaml:"formality" validate:"required,oneof=casual neutral formal"`
    Verbosity       string   `json:"verbosity" yaml:"verbosity" validate:"required,oneof=concise balanced verbose"`
    Perspective     string   `json:"perspective,omitempty" yaml:"perspective,omitempty"`
    Characteristics []string `json:"characteristics,omitempty" yaml:"characteristics,omitempty"`
}
```

**Examples:**
```yaml
response_style:
  tone: "Professional and encouraging"
  formality: "neutral"
  verbosity: "balanced"
  perspective: "First-person, collaborative"
  characteristics:
    - "Clear explanations"
    - "Code examples"
    - "Best practices"
```

### Constructor

```go
func NewPersona(name, description, version, author string) *Persona {
    now := time.Now()
    return &Persona{
        metadata: ElementMetadata{
            ID:          GenerateElementID(PersonaElement, name),
            Type:        PersonaElement,
            Name:        name,
            Description: description,
            Version:     version,
            Author:      author,
            Tags:        []string{},
            IsActive:    true,
            CreatedAt:   now,
            UpdatedAt:   now,
        },
        BehavioralTraits: []BehavioralTrait{},
        ExpertiseAreas:   []ExpertiseArea{},
        ResponseStyle:    ResponseStyle{},
        PrivacyLevel:     PrivacyPublic,
        HotSwappable:     true,
    }
}
```

### Validation

```go
func (p *Persona) Validate() error {
    // Validate metadata
    if err := p.metadata.Validate(); err != nil {
        return fmt.Errorf("metadata validation failed: %w", err)
    }
    
    // At least one behavioral trait required
    if len(p.BehavioralTraits) == 0 {
        return fmt.Errorf("at least one behavioral trait is required")
    }
    
    // Validate each trait
    for i, trait := range p.BehavioralTraits {
        if trait.Name == "" {
            return fmt.Errorf("behavioral trait %d: name is required", i)
        }
        if trait.Intensity < 1 || trait.Intensity > 10 {
            return fmt.Errorf("behavioral trait %d: intensity must be 1-10", i)
        }
    }
    
    // At least one expertise area required
    if len(p.ExpertiseAreas) == 0 {
        return fmt.Errorf("at least one expertise area is required")
    }
    
    // Validate each expertise area
    for i, area := range p.ExpertiseAreas {
        if area.Domain == "" {
            return fmt.Errorf("expertise area %d: domain is required", i)
        }
        validLevels := map[string]bool{
            "beginner": true, "intermediate": true, 
            "advanced": true, "expert": true,
        }
        if !validLevels[area.Level] {
            return fmt.Errorf("expertise area %d: invalid level %s", i, area.Level)
        }
    }
    
    // System prompt required
    if p.SystemPrompt == "" {
        return fmt.Errorf("system_prompt is required")
    }
    if len(p.SystemPrompt) < 10 || len(p.SystemPrompt) > 2000 {
        return fmt.Errorf("system_prompt must be 10-2000 characters")
    }
    
    // Validate privacy level
    validPrivacy := map[PersonaPrivacyLevel]bool{
        PrivacyPublic: true, PrivacyPrivate: true, PrivacyShared: true,
    }
    if !validPrivacy[p.PrivacyLevel] {
        return fmt.Errorf("invalid privacy level: %s", p.PrivacyLevel)
    }
    
    return nil
}
```

### Example Persona

```yaml
metadata:
  id: "persona-senior-architect-1703088000"
  type: "persona"
  name: "Senior Software Architect"
  description: "Expert in system design and clean architecture"
  version: "1.0.0"
  author: "team@company.com"
  tags: ["architecture", "expert", "mentor"]
  is_active: true
  created_at: "2025-12-20T10:00:00Z"
  updated_at: "2025-12-20T10:00:00Z"

behavioral_traits:
  - name: "Analytical"
    description: "Systematic approach to problem-solving"
    intensity: 9
  - name: "Patient"
    description: "Takes time to explain complex concepts"
    intensity: 8
  - name: "Detail-Oriented"
    intensity: 10

expertise_areas:
  - domain: "Software Architecture"
    level: "expert"
    keywords: ["clean architecture", "DDD", "microservices", "event-driven"]
    description: "25+ years designing scalable systems"
  - domain: "Go Programming"
    level: "expert"
    keywords: ["concurrency", "performance", "best practices"]

response_style:
  tone: "Professional, encouraging, mentoring"
  formality: "neutral"
  verbosity: "balanced"
  perspective: "First-person, collaborative"
  characteristics:
    - "Explains rationale behind decisions"
    - "Provides code examples"
    - "References best practices and patterns"
    - "Encourages learning and growth"

system_prompt: |
  You are a Senior Software Architect with 25+ years of experience...
  
privacy_level: "public"
hot_swappable: true
```

---

## Skill Element

### Definition

A **Skill** represents a specialized capability with triggers, procedures, and dependencies.

### Structure

```go
type Skill struct {
    metadata      ElementMetadata
    Triggers      []SkillTrigger    `json:"triggers" yaml:"triggers" validate:"required,min=1,dive"`
    Procedures    []SkillProcedure  `json:"procedures" yaml:"procedures" validate:"required,min=1,dive"`
    Dependencies  []SkillDependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
    ToolsRequired []string          `json:"tools_required,omitempty" yaml:"tools_required,omitempty"`
    Inputs        map[string]string `json:"inputs,omitempty" yaml:"inputs,omitempty"`
    Outputs       map[string]string `json:"outputs,omitempty" yaml:"outputs,omitempty"`
    Composable    bool              `json:"composable" yaml:"composable"`
}
```

### Skill Triggers

```go
type SkillTrigger struct {
    Type     string   `json:"type" yaml:"type" validate:"required,oneof=keyword pattern context manual"`
    Keywords []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`
    Pattern  string   `json:"pattern,omitempty" yaml:"pattern,omitempty"`
    Context  string   `json:"context,omitempty" yaml:"context,omitempty"`
}
```

**Trigger Types:**

| Type | Description | Example |
|------|-------------|---------|
| **keyword** | Activated by specific keywords | "review", "analyze", "optimize" |
| **pattern** | Activated by regex pattern | `^review\s+code` |
| **context** | Activated by contextual analysis | "code quality discussion" |
| **manual** | Explicitly invoked | User command |

**Examples:**
```yaml
triggers:
  - type: "keyword"
    keywords: ["review code", "code review", "analyze code"]
  - type: "pattern"
    pattern: "^(review|analyze|audit)\\s+(code|implementation)"
  - type: "context"
    context: "Discussion about code quality or improvements"
```

### Skill Procedures

```go
type SkillProcedure struct {
    Step        int      `json:"step" yaml:"step" validate:"required,min=1"`
    Action      string   `json:"action" yaml:"action" validate:"required"`
    Description string   `json:"description,omitempty" yaml:"description,omitempty"`
    ToolsUsed   []string `json:"tools_used,omitempty" yaml:"tools_used,omitempty"`
    Validation  string   `json:"validation,omitempty" yaml:"validation,omitempty"`
}
```

**Examples:**
```yaml
procedures:
  - step: 1
    action: "Analyze code structure"
    description: "Review overall architecture and organization"
    tools_used: ["list_files", "read_file"]
    validation: "All relevant files identified"
    
  - step: 2
    action: "Check code style"
    description: "Verify adherence to style guidelines"
    tools_used: ["read_file", "search"]
    validation: "Style issues documented"
    
  - step: 3
    action: "Assess test coverage"
    description: "Evaluate test completeness"
    tools_used: ["list_files", "read_file"]
    validation: "Coverage metrics calculated"
```

### Skill Dependencies

```go
type SkillDependency struct {
    SkillID  string `json:"skill_id" yaml:"skill_id" validate:"required"`
    Required bool   `json:"required" yaml:"required"`
    Version  string `json:"version,omitempty" yaml:"version,omitempty"`
}
```

**Examples:**
```yaml
dependencies:
  - skill_id: "skill-read-files-1703088000"
    required: true
    version: ">=1.0.0"
  - skill_id: "skill-analyze-syntax-1703088001"
    required: false
    version: "^2.0.0"
```

### Constructor

```go
func NewSkill(name, description, version, author string) *Skill {
    now := time.Now()
    return &Skill{
        metadata: ElementMetadata{
            ID:          GenerateElementID(SkillElement, name),
            Type:        SkillElement,
            Name:        name,
            Description: description,
            Version:     version,
            Author:      author,
            Tags:        []string{},
            IsActive:    true,
            CreatedAt:   now,
            UpdatedAt:   now,
        },
        Triggers:      []SkillTrigger{},
        Procedures:    []SkillProcedure{},
        Dependencies:  []SkillDependency{},
        ToolsRequired: []string{},
        Inputs:        make(map[string]string),
        Outputs:       make(map[string]string),
        Composable:    true,
    }
}
```

### Validation

```go
func (s *Skill) Validate() error {
    if err := s.metadata.Validate(); err != nil {
        return fmt.Errorf("metadata validation failed: %w", err)
    }
    
    // At least one trigger required
    if len(s.Triggers) == 0 {
        return fmt.Errorf("at least one trigger is required")
    }
    
    // Validate triggers
    for i, trigger := range s.Triggers {
        validTypes := map[string]bool{
            "keyword": true, "pattern": true, "context": true, "manual": true,
        }
        if !validTypes[trigger.Type] {
            return fmt.Errorf("trigger %d: invalid type %s", i, trigger.Type)
        }
    }
    
    // At least one procedure required
    if len(s.Procedures) == 0 {
        return fmt.Errorf("at least one procedure is required")
    }
    
    // Validate procedures
    for i, proc := range s.Procedures {
        if proc.Step < 1 {
            return fmt.Errorf("procedure %d: step must be >= 1", i)
        }
        if proc.Action == "" {
            return fmt.Errorf("procedure %d: action is required", i)
        }
    }
    
    return nil
}
```

### Example Skill

```yaml
metadata:
  id: "skill-code-review-1703088000"
  type: "skill"
  name: "Code Review"
  description: "Comprehensive code review skill"
  version: "2.1.0"
  author: "engineering@company.com"
  tags: ["code-quality", "review", "analysis"]
  is_active: true

triggers:
  - type: "keyword"
    keywords: ["review code", "code review", "analyze code"]
  - type: "pattern"
    pattern: "^review\\s+\\w+"

procedures:
  - step: 1
    action: "Read and understand code"
    description: "Analyze structure and logic"
    tools_used: ["list_files", "read_file"]
  - step: 2
    action: "Check style and conventions"
    description: "Verify adherence to standards"
    tools_used: ["search", "read_file"]
  - step: 3
    action: "Identify potential issues"
    tools_used: ["search", "read_file"]
  - step: 4
    action: "Generate review report"
    description: "Summarize findings"

dependencies:
  - skill_id: "skill-syntax-analysis-1703088001"
    required: true
    version: ">=1.0.0"

tools_required:
  - "read_file"
  - "list_files"
  - "search"

inputs:
  file_path: "Path to file or directory to review"
  focus_areas: "Specific areas to focus on (optional)"

outputs:
  review_report: "Detailed review findings"
  severity_level: "Overall code quality assessment"

composable: true
```

---

## Template Element

### Definition

A **Template** represents reusable content with variables and validation rules.

### Structure

```go
type Template struct {
    metadata        ElementMetadata
    Content         string             `json:"content" yaml:"content" validate:"required"`
    Format          string             `json:"format" yaml:"format" validate:"required,oneof=markdown yaml json text"`
    Variables       []TemplateVariable `json:"variables" yaml:"variables" validate:"dive"`
    ValidationRules map[string]string  `json:"validation_rules,omitempty" yaml:"validation_rules,omitempty"`
}
```

### Template Variables

```go
type TemplateVariable struct {
    Name        string `json:"name" yaml:"name" validate:"required"`
    Type        string `json:"type" yaml:"type" validate:"required,oneof=string number boolean array object"`
    Required    bool   `json:"required" yaml:"required"`
    Default     string `json:"default,omitempty" yaml:"default,omitempty"`
    Description string `json:"description,omitempty" yaml:"description,omitempty"`
}
```

### Rendering

```go
func (t *Template) Render(values map[string]string) (string, error) {
    result := t.Content
    
    // Replace each variable
    for _, v := range t.Variables {
        val, ok := values[v.Name]
        if !ok {
            if v.Required {
                return "", fmt.Errorf("required variable %s not provided", v.Name)
            }
            val = v.Default
        }
        
        // Replace {{variable}} with value
        result = strings.ReplaceAll(result, "{{"+v.Name+"}}", val)
    }
    
    return result, nil
}
```

### Example Template

```yaml
metadata:
  id: "template-code-review-prompt-1703088000"
  type: "template"
  name: "Code Review Prompt"
  description: "Standard prompt for code reviews"
  version: "1.2.0"
  author: "engineering@company.com"
  tags: ["prompt", "code-review"]
  is_active: true

content: |
  # Code Review: {{project_name}}
  
  ## Review Scope
  {{review_scope}}
  
  ## Focus Areas
  {{focus_areas}}
  
  ## Instructions
  Please review the code with attention to:
  - Code quality and maintainability
  - Performance considerations
  - Security vulnerabilities
  - Best practices adherence
  
  ## Additional Context
  {{additional_context}}

format: "markdown"

variables:
  - name: "project_name"
    type: "string"
    required: true
    description: "Name of the project being reviewed"
    
  - name: "review_scope"
    type: "string"
    required: true
    description: "What is being reviewed"
    
  - name: "focus_areas"
    type: "string"
    required: false
    default: "All aspects"
    description: "Specific areas to focus on"
    
  - name: "additional_context"
    type: "string"
    required: false
    default: "None"
    description: "Any additional context"

validation_rules:
  project_name: "min_length:3,max_length:100"
  review_scope: "min_length:10"
```

---

## Agent Element

### Definition

An **Agent** is an autonomous task executor with goals, actions, and decision-making.

### Structure

```go
type Agent struct {
    metadata         ElementMetadata
    Goals            []string               `json:"goals" yaml:"goals" validate:"required,min=1"`
    Actions          []AgentAction          `json:"actions" yaml:"actions" validate:"required,min=1,dive"`
    DecisionTree     map[string]interface{} `json:"decision_tree,omitempty" yaml:"decision_tree,omitempty"`
    FallbackStrategy string                 `json:"fallback_strategy,omitempty" yaml:"fallback_strategy,omitempty"`
    MaxIterations    int                    `json:"max_iterations" yaml:"max_iterations" validate:"min=1,max=100"`
    Context          map[string]interface{} `json:"context,omitempty" yaml:"context,omitempty"`
}
```

### Agent Actions

```go
type AgentAction struct {
    Name       string            `json:"name" yaml:"name" validate:"required"`
    Type       string            `json:"type" yaml:"type" validate:"required,oneof=tool skill decision loop"`
    Parameters map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
    OnSuccess  string            `json:"on_success,omitempty" yaml:"on_success,omitempty"`
    OnFailure  string            `json:"on_failure,omitempty" yaml:"on_failure,omitempty"`
}
```

**Action Types:**

| Type | Description | Example |
|------|-------------|---------|
| **tool** | Execute an MCP tool | `create_element`, `search` |
| **skill** | Execute a skill | Code review, analysis |
| **decision** | Make a decision | Branch logic |
| **loop** | Repeat actions | Iterate over items |

### Example Agent

```yaml
metadata:
  id: "agent-task-planner-1703088000"
  type: "agent"
  name: "Task Planning Agent"
  description: "Plans and breaks down complex tasks"
  version: "1.0.0"
  author: "team@company.com"
  tags: ["planning", "automation"]
  is_active: true

goals:
  - "Analyze task requirements"
  - "Break down into subtasks"
  - "Prioritize execution order"
  - "Identify dependencies"

actions:
  - name: "analyze_task"
    type: "skill"
    parameters:
      skill_id: "skill-task-analysis-1703088000"
    on_success: "break_down_task"
    on_failure: "request_clarification"
    
  - name: "break_down_task"
    type: "tool"
    parameters:
      tool: "create_element"
      type: "agent"
    on_success: "prioritize"
    
  - name: "prioritize"
    type: "decision"
    parameters:
      criteria: "complexity,dependencies,urgency"

max_iterations: 10
fallback_strategy: "request_human_assistance"
```

---

## Memory Element

### Definition

A **Memory** stores persistent context with deduplication via content hashing.

### Structure

```go
type Memory struct {
    metadata    ElementMetadata
    Content     string            `json:"content" yaml:"content" validate:"required"`
    DateCreated string            `json:"date_created" yaml:"date_created"`
    ContentHash string            `json:"content_hash" yaml:"content_hash"`
    SearchIndex []string          `json:"search_index,omitempty" yaml:"search_index,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
```

### Content Hashing

```go
func (m *Memory) ComputeHash() {
    hash := sha256.Sum256([]byte(m.Content))
    m.ContentHash = fmt.Sprintf("%x", hash)
}
```

Prevents duplicate memories by comparing hashes.

### Example Memory

```yaml
metadata:
  id: "memory-project-context-1703088000"
  type: "memory"
  name: "NEXS MCP Project Context"
  description: "Key information about the project"
  version: "1.0.0"
  author: "team@company.com"
  tags: ["project", "context"]
  is_active: true

content: |
  NEXS MCP is a Model Context Protocol server built in Go.
  Uses Clean Architecture with 4 layers: Domain, Application, Infrastructure, MCP.
  72.2% test coverage. 55 MCP tools. 6 element types.
  
date_created: "2025-12-20"
content_hash: "3f79bb7b..."
search_index:
  - "nexs"
  - "mcp"
  - "clean architecture"
  - "domain driven"
metadata:
  project: "nexs-mcp"
  category: "technical"
```

---

## Ensemble Element

### Definition

An **Ensemble** orchestrates multiple agents with execution strategies.

### Structure

```go
type Ensemble struct {
    metadata            ElementMetadata
    Members             []EnsembleMember       `json:"members" yaml:"members" validate:"required,min=1,dive"`
    ExecutionMode       string                 `json:"execution_mode" yaml:"execution_mode" validate:"required,oneof=sequential parallel hybrid"`
    AggregationStrategy string                 `json:"aggregation_strategy" yaml:"aggregation_strategy" validate:"required"`
    FallbackChain       []string               `json:"fallback_chain,omitempty" yaml:"fallback_chain,omitempty"`
    SharedContext       map[string]interface{} `json:"shared_context,omitempty" yaml:"shared_context,omitempty"`
}
```

### Ensemble Members

```go
type EnsembleMember struct {
    AgentID  string `json:"agent_id" yaml:"agent_id" validate:"required"`
    Role     string `json:"role" yaml:"role" validate:"required"`
    Priority int    `json:"priority" yaml:"priority" validate:"min=1,max=10"`
}
```

### Execution Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| **sequential** | Agents run in order | Pipeline processing |
| **parallel** | Agents run concurrently | Independent tasks |
| **hybrid** | Mix of sequential and parallel | Complex workflows |

### Aggregation Strategies

| Strategy | Description |
|----------|-------------|
| **first** | Return first successful result |
| **last** | Return last result |
| **consensus** | Majority voting |
| **all** | Return all results |
| **merge** | Merge all results |

### Example Ensemble

```yaml
metadata:
  id: "ensemble-code-review-team-1703088000"
  type: "ensemble"
  name: "Code Review Team"
  description: "Multi-agent code review"
  version: "1.0.0"
  author: "team@company.com"
  tags: ["code-review", "team"]
  is_active: true

members:
  - agent_id: "agent-style-reviewer-1703088000"
    role: "style_checker"
    priority: 8
  - agent_id: "agent-security-reviewer-1703088001"
    role: "security_checker"
    priority: 10
  - agent_id: "agent-performance-reviewer-1703088002"
    role: "performance_checker"
    priority: 7

execution_mode: "parallel"
aggregation_strategy: "merge"
fallback_chain:
  - "agent-general-reviewer-1703088003"
```

---

## Repository Interface

### Definition

The repository interface abstracts storage operations:

```go
type ElementRepository interface {
    Create(element Element) error
    GetByID(id string) (Element, error)
    Update(element Element) error
    Delete(id string) error
    List(filter ElementFilter) ([]Element, error)
    Exists(id string) (bool, error)
}
```

### Element Filter

```go
type ElementFilter struct {
    Type     *ElementType `json:"type,omitempty"`
    IsActive *bool        `json:"is_active,omitempty"`
    Tags     []string     `json:"tags,omitempty"`
    Limit    int          `json:"limit,omitempty"`
    Offset   int          `json:"offset,omitempty"`
}
```

### Usage Example

```go
// Domain code uses repository interface
func SavePersona(repo ElementRepository, persona *Persona) error {
    // Validate first
    if err := persona.Validate(); err != nil {
        return err
    }
    
    // Save via repository
    return repo.Create(persona)
}

// Works with any implementation
fileRepo := infrastructure.NewFileElementRepository("data/elements")
memRepo := infrastructure.NewInMemoryElementRepository()

SavePersona(fileRepo, persona)  // Saves to files
SavePersona(memRepo, persona)   // Saves to memory
```

---

## Domain Rules & Invariants

### Element Invariants

1. **Unique IDs** - No two elements can have the same ID
2. **Valid Type** - Type must be one of 6 valid types
3. **Name Length** - Name must be 3-100 characters
4. **Version Format** - Must follow semantic versioning
5. **Timestamps** - UpdatedAt >= CreatedAt

### Persona Invariants

1. **Behavioral Traits** - At least one trait required
2. **Expertise Areas** - At least one area required
3. **System Prompt** - 10-2000 characters
4. **Privacy Level** - Must be valid level
5. **Shared Access** - If shared, owner + shared_with must exist

### Skill Invariants

1. **Triggers** - At least one trigger required
2. **Procedures** - At least one procedure required
3. **Step Ordering** - Steps must be sequential (1, 2, 3...)
4. **Dependencies** - Referenced skills must exist

### Agent Invariants

1. **Goals** - At least one goal required
2. **Actions** - At least one action required
3. **Max Iterations** - Must be 1-100
4. **Action References** - on_success/on_failure must reference valid actions

### Ensemble Invariants

1. **Members** - At least one member required
2. **Execution Mode** - Must be valid mode
3. **Aggregation Strategy** - Must be valid strategy
4. **Member References** - All agent_ids must exist
5. **Priority Range** - Priorities must be 1-10

---

## Validation System

### Validation Approach

NEXS MCP uses three levels of validation:

1. **Struct Tags** - Compile-time validation
2. **Validate() Method** - Domain-level validation
3. **Business Rules** - Complex invariants

### Example: Multi-Level Validation

```go
type Persona struct {
    metadata         ElementMetadata
    // Level 1: Struct tags
    BehavioralTraits []BehavioralTrait `validate:"required,min=1,dive"`
    SystemPrompt     string            `validate:"required,min=10,max=2000"`
}

func (p *Persona) Validate() error {
    // Level 2: Metadata validation
    if err := p.metadata.Validate(); err != nil {
        return err
    }
    
    // Level 3: Business rules
    if len(p.BehavioralTraits) == 0 {
        return fmt.Errorf("at least one behavioral trait is required")
    }
    
    // Complex validation
    for i, trait := range p.BehavioralTraits {
        if trait.Intensity < 1 || trait.Intensity > 10 {
            return fmt.Errorf("trait %d: intensity must be 1-10", i)
        }
    }
    
    return nil
}
```

---

## Access Control

### Access Control Model

```go
type AccessControl struct {
    Owner      string   `json:"owner"`
    Privacy    string   `json:"privacy"`
    SharedWith []string `json:"shared_with"`
}

func (ac *AccessControl) CanAccess(userID string) bool {
    // Owner always has access
    if ac.Owner == userID {
        return true
    }
    
    // Public elements
    if ac.Privacy == "public" {
        return true
    }
    
    // Shared elements
    if ac.Privacy == "shared" {
        for _, shared := range ac.SharedWith {
            if shared == userID {
                return true
            }
        }
    }
    
    return false
}

func (ac *AccessControl) CanModify(userID string) bool {
    // Only owner can modify
    return ac.Owner == userID
}
```

---

## Domain Events

### Event System (Future Enhancement)

```go
// Domain events for future implementation
type DomainEvent interface {
    EventName() string
    OccurredAt() time.Time
    AggregateID() string
}

type ElementCreated struct {
    elementID string
    elementType ElementType
    occurredAt time.Time
}

type ElementActivated struct {
    elementID string
    occurredAt time.Time
}

type ElementDeactivated struct {
    elementID string
    occurredAt time.Time
}
```

---

## Best Practices

### 1. Keep Domain Pure

❌ **Bad:**
```go
func (p *Persona) SaveToFile(path string) error {
    // Infrastructure concern in domain
}
```

✅ **Good:**
```go
func (p *Persona) Validate() error {
    // Pure domain logic
}
```

### 2. Validate Early

```go
func NewPersona(...) *Persona {
    p := &Persona{...}
    // Don't validate in constructor
    return p
}

// Validate before persistence
if err := persona.Validate(); err != nil {
    return err
}
repo.Create(persona)
```

### 3. Use Value Objects

```go
// ✅ Value object - immutable
type BehavioralTrait struct {
    Name      string
    Intensity int
}

// ❌ Not a value object - mutable state
type BehavioralTraitManager struct {
    currentIntensity int
}
```

### 4. Explicit State Changes

```go
// ✅ Explicit
func (p *Persona) Activate() error {
    p.metadata.IsActive = true
    p.metadata.UpdatedAt = time.Now()
    return nil
}

// ❌ Implicit
func (p *Persona) SetActive() {
    // What happens? Unclear
}
```

### 5. Return Errors, Don't Panic

```go
// ✅ Good
func (p *Persona) Validate() error {
    if p.SystemPrompt == "" {
        return fmt.Errorf("system_prompt is required")
    }
    return nil
}

// ❌ Bad
func (p *Persona) Validate() {
    if p.SystemPrompt == "" {
        panic("system_prompt is required")
    }
}
```

---

## Conclusion

The Domain Layer is the foundation of NEXS MCP Server. Its purity, clarity, and focus on business logic make the system:

- **Testable** - 79.2% coverage without infrastructure
- **Maintainable** - Business rules clearly expressed
- **Portable** - Works with any storage/framework
- **Evolvable** - Business logic changes independently

**Key Principles:**
1. Zero external dependencies
2. Rich domain model
3. Explicit validation
4. Clear invariants
5. Interface-based design

---

**Document Version:** 1.0.0  
**Total Lines:** 1203  
**Last Updated:** December 20, 2025
