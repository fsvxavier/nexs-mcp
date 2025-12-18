# Especificação das Ferramentas MCP

**Versão:** 1.0  
**Data:** 18 de Dezembro de 2025

## Visão Geral

Este documento especifica as 41 ferramentas MCP implementadas usando:
- **MCP SDK Oficial:** `github.com/modelcontextprotocol/go-sdk` para protocol compliance
- **Auto Schema Generation:** `invopop/jsonschema` gera JSON Schema via reflection de structs Go
- **Validation Tags:** `go-playground/validator/v10` para validação automática
- **Transportes:** Stdio (padrão), SSE, HTTP disponíveis via SDK

Cada tool utiliza struct tags para definir schema e validação simultaneamente:
```go
type ToolInput struct {
    Field string `json:"field" jsonschema:"required,minLength=3" validate:"required,min=3"`
}
// JSON Schema gerado automaticamente via reflection
// Validação automática via struct tags
```

## Índice
1. [Element Management Tools (12)](#element-management-tools)
2. [Private Personas Tools (8)](#private-personas-tools)
3. [Collection Tools (5)](#collection-tools)
4. [Portfolio Tools (8)](#portfolio-tools)
5. [Search Tools (4)](#search-tools)
6. [Configuration Tools (6)](#configuration-tools)
7. [Security Tools (4)](#security-tools)
8. [Capability Index Tools (2)](#capability-index-tools)

**Total:** 49 ferramentas MCP

---

## Model Configuration Tools (2)

### 0. `get_supported_models`

**Descrição:** Lista todos os modelos de IA suportados pelo servidor

**Entrada:**
```json
{}
```

**Saída:**
```json
{
  "models": [
    {
      "id": "auto",
      "name": "Auto-select",
      "provider": "system",
      "description": "Seleciona automaticamente o melhor modelo para a solicitação",
      "capabilities": ["auto-selection"],
      "available": true
    },
    {
      "id": "claude-sonnet-4.5",
      "name": "Claude Sonnet 4.5",
      "provider": "anthropic",
      "description": "Equilíbrio entre capacidade e velocidade",
      "capabilities": ["general", "code", "reasoning"],
      "available": true
    },
    {
      "id": "gpt-5.1-codex-max",
      "name": "GPT-5.1 Codex Max",
      "provider": "openai",
      "description": "Máxima capacidade para geração de código",
      "capabilities": ["code", "advanced"],
      "available": true
    }
  ],
  "total": 20,
  "default_model": "auto"
}
```

**Implementação Go:**
```go
package tools

import (
    "context"
)

// GetSupportedModelsInput - Schema gerado automaticamente
type GetSupportedModelsInput struct{}

// GetSupportedModelsOutput - Schema gerado automaticamente
type GetSupportedModelsOutput struct {
    Models       []ModelInfo `json:"models" jsonschema:"required,description=List of supported AI models"`
    Total        int         `json:"total" jsonschema:"required,description=Total count of models"`
    DefaultModel string      `json:"default_model" jsonschema:"required,description=Default model when not specified"`
}

type ModelInfo struct {
    ID           string   `json:"id" jsonschema:"required"`
    Name         string   `json:"name" jsonschema:"required"`
    Provider     string   `json:"provider" jsonschema:"required,enum=anthropic|google|openai|xai|oswe|system"`
    Description  string   `json:"description" jsonschema:"required"`
    Capabilities []string `json:"capabilities" jsonschema:"required"`
    Available    bool     `json:"available" jsonschema:"required"`
}

var SupportedModels = []ModelInfo{
    {ID: "auto", Name: "Auto-select", Provider: "system", Description: "Seleciona automaticamente o melhor modelo", Capabilities: []string{"auto-selection"}, Available: true},
    {ID: "claude-sonnet-4.5", Name: "Claude Sonnet 4.5", Provider: "anthropic", Description: "Equilíbrio entre capacidade e velocidade", Capabilities: []string{"general", "code", "reasoning"}, Available: true},
    {ID: "claude-haiku-4.5", Name: "Claude Haiku 4.5", Provider: "anthropic", Description: "Respostas rápidas e eficientes", Capabilities: []string{"general", "fast"}, Available: true},
    {ID: "claude-opus-4.5", Name: "Claude Opus 4.5", Provider: "anthropic", Description: "Máxima capacidade de raciocínio", Capabilities: []string{"general", "code", "reasoning", "advanced"}, Available: true},
    {ID: "claude-sonnet-4", Name: "Claude Sonnet 4", Provider: "anthropic", Description: "Versão estável anterior", Capabilities: []string{"general", "code", "reasoning"}, Available: true},
    {ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: "google", Description: "Gemini Pro versão 2.5", Capabilities: []string{"general", "code", "reasoning"}, Available: true},
    {ID: "gemini-3-flash-preview", Name: "Gemini 3 Flash Preview", Provider: "google", Description: "Preview de alta velocidade", Capabilities: []string{"general", "fast"}, Available: true},
    {ID: "gemini-3-pro-preview", Name: "Gemini 3 Pro Preview", Provider: "google", Description: "Preview avançado", Capabilities: []string{"general", "code", "reasoning", "advanced"}, Available: true},
    {ID: "gpt-4.1", Name: "GPT-4.1", Provider: "openai", Description: "GPT-4.1 base", Capabilities: []string{"general", "code"}, Available: true},
    {ID: "gpt-4o", Name: "GPT-4o", Provider: "openai", Description: "GPT-4 otimizado", Capabilities: []string{"general", "code"}, Available: true},
    {ID: "gpt-5", Name: "GPT-5", Provider: "openai", Description: "GPT-5 base", Capabilities: []string{"general", "code", "reasoning"}, Available: true},
    {ID: "gpt-5-mini", Name: "GPT-5 Mini", Provider: "openai", Description: "Versão compacta e eficiente", Capabilities: []string{"general", "fast"}, Available: true},
    {ID: "gpt-5-codex", Name: "GPT-5 Codex", Provider: "openai", Description: "Especializado em código", Capabilities: []string{"code"}, Available: true},
    {ID: "gpt-5.1", Name: "GPT-5.1", Provider: "openai", Description: "GPT-5.1 base", Capabilities: []string{"general", "code", "reasoning"}, Available: true},
    {ID: "gpt-5.1-codex", Name: "GPT-5.1 Codex", Provider: "openai", Description: "Codex versão 5.1", Capabilities: []string{"code", "advanced"}, Available: true},
    {ID: "gpt-5.1-codex-max", Name: "GPT-5.1 Codex Max", Provider: "openai", Description: "Máxima capacidade para código", Capabilities: []string{"code", "advanced"}, Available: true},
    {ID: "gpt-5.1-codex-mini", Name: "GPT-5.1 Codex Mini", Provider: "openai", Description: "Versão compacta do Codex", Capabilities: []string{"code", "fast"}, Available: true},
    {ID: "gpt-5.2", Name: "GPT-5.2", Provider: "openai", Description: "Última versão GPT-5", Capabilities: []string{"general", "code", "reasoning", "advanced"}, Available: true},
    {ID: "grok-code-fast-1", Name: "Grok Code Fast 1", Provider: "xai", Description: "Otimizado para geração rápida de código", Capabilities: []string{"code", "fast"}, Available: true},
    {ID: "oswe-vscode-prim", Name: "OSWE VSCode Prim", Provider: "oswe", Description: "Integração especializada para VSCode", Capabilities: []string{"code", "vscode"}, Available: true},
}

func (h *ModelHandler) GetSupportedModels(ctx context.Context, input GetSupportedModelsInput) (*GetSupportedModelsOutput, error) {
    return &GetSupportedModelsOutput{
        Models:       SupportedModels,
        Total:        len(SupportedModels),
        DefaultModel: "auto",
    }, nil
}
```

---

### 1. `set_default_model`

**Descrição:** Define o modelo padrão a ser usado nas solicitações

**Entrada:**
```json
{
  "model_id": "claude-sonnet-4.5"
}
```

**Saída:**
```json
{
  "success": true,
  "previous_model": "auto",
  "new_model": "claude-sonnet-4.5",
  "message": "Modelo padrão atualizado com sucesso"
}
```

**Implementação Go:**
```go
// SetDefaultModelInput - Schema gerado automaticamente
type SetDefaultModelInput struct {
    ModelID string `json:"model_id" jsonschema:"required,description=ID of the model to set as default" validate:"required,oneof=auto claude-sonnet-4.5 claude-haiku-4.5 claude-opus-4.5 claude-sonnet-4 gemini-2.5-pro gemini-3-flash-preview gemini-3-pro-preview gpt-4.1 gpt-4o gpt-5 gpt-5-mini gpt-5-codex gpt-5.1 gpt-5.1-codex gpt-5.1-codex-max gpt-5.1-codex-mini gpt-5.2 grok-code-fast-1 oswe-vscode-prim"`
}

// SetDefaultModelOutput - Schema gerado automaticamente
type SetDefaultModelOutput struct {
    Success       bool   `json:"success" jsonschema:"required"`
    PreviousModel string `json:"previous_model" jsonschema:"required"`
    NewModel      string `json:"new_model" jsonschema:"required"`
    Message       string `json:"message" jsonschema:"required"`
}

func (h *ModelHandler) SetDefaultModel(ctx context.Context, input SetDefaultModelInput) (*SetDefaultModelOutput, error) {
    // Validação automática via struct tags já executada
    
    // Get current default
    previous := h.config.DefaultModel
    
    // Update config
    h.config.DefaultModel = input.ModelID
    
    // Persist config
    if err := h.config.Save(); err != nil {
        return nil, fmt.Errorf("failed to save config: %w", err)
    }
    
    return &SetDefaultModelOutput{
        Success:       true,
        PreviousModel: previous,
        NewModel:      input.ModelID,
        Message:       "Modelo padrão atualizado com sucesso",
    }, nil
}
```

---

## Element Management Tools

### 1. `list_elements`

**Descrição:** Lista elementos disponíveis por tipo

**Entrada:**
```json
{
  "type": "personas|skills|templates|agents|memories|ensembles"
}
```

**Saída:**
```json
{
  "elements": [
    {
      "id": "persona_creative-writer_alice_20251218-120000",
      "name": "creative-writer",
      "type": "personas",
      "version": "1.0.0",
      "author": "alice",
      "description": "Imaginative storyteller for engaging narratives",
      "created_at": "2025-12-18T12:00:00Z"
    }
  ],
  "total": 42
}
```

**Implementação Go:**
```go
package tools

import (
    "context"
    "fmt"
    "github.com/fsvxavier/mcp-server/internal/domain"
)

// ListElementsInput - Schema gerado automaticamente via reflection
type ListElementsInput struct {
    Type string `json:"type" jsonschema:"required,enum=personas|skills|templates|agents|memories|ensembles,description=Element type to list" validate:"required,oneof=personas skills templates agents memories ensembles"`
}

// ListElementsOutput - Schema gerado automaticamente
type ListElementsOutput struct {
    Elements []ElementSummary `json:"elements" jsonschema:"required,description=List of elements"`
    Total    int              `json:"total" jsonschema:"required,description=Total count"`
}

type ElementSummary struct {
    ID          string `json:"id" jsonschema:"required"`
    Name        string `json:"name" jsonschema:"required"`
    Type        string `json:"type" jsonschema:"required"`
    Version     string `json:"version" jsonschema:"required"`
    Author      string `json:"author" jsonschema:"required"`
    Description string `json:"description" jsonschema:"required"`
    CreatedAt   string `json:"created_at" jsonschema:"required,format=date-time"`
}

// Handler registrado automaticamente com SDK
func (h *ElementHandler) ListElements(ctx context.Context, input ListElementsInput) (*ListElementsOutput, error) {
    // 1. Validation automática via struct tags (já executada pelo SDK)
    
    // 2. Get elements from repository
    elements, err := h.repo.FindByType(ctx, domain.ElementType(input.Type))
    if err != nil {
        return nil, fmt.Errorf("failed to list elements: %w", err)
    }
    
    // 3. Map to summaries
    summaries := make([]ElementSummary, len(elements))
    for i, elem := range elements {
        summaries[i] = ElementSummary{
            ID:          elem.ID,
            Name:        elem.Name,
            Type:        string(elem.Type),
            Version:     elem.Version,
            Author:      elem.Author,
            Description: elem.Description,
            CreatedAt:   elem.CreatedAt.Format("2006-01-02T15:04:05Z"),
        }
    }
    
    return &ListElementsOutput{
        Elements: summaries,
        Total:    len(summaries),
    }, nil
}

// JSON Schema gerado automaticamente:
// {
//   "type": "object",
//   "properties": {
//     "type": {
//       "type": "string",
//       "enum": ["personas", "skills", "templates", "agents", "memories", "ensembles"],
//       "description": "Element type to list"
//     }
//   },
//   "required": ["type"]
// }
```

---

### 2. `create_element`

**Descrição:** Cria um novo elemento

**Entrada:**
```json
{
  "type": "personas",
  "name": "technical-writer",
  "author": "bob",
  "description": "Expert in technical documentation",
  "content": "# Technical Writer\n\nBehavioral traits:\n- Clarity\n- Precision...",
  "metadata": {
    "behavioral_traits": ["clarity", "precision"],
    "category": "writing"
  }
}
```

**Saída:**
```json
{
  "element": {
    "id": "persona_technical-writer_bob_20251218-120530",
    "name": "technical-writer",
    "type": "personas",
    "version": "1.0.0",
    "created_at": "2025-12-18T12:05:30Z"
  },
  "path": "/home/user/.dollhouse/portfolio/personas/technical-writer.md"
}
```

**Validações (300+ regras):**
```go
func (v *ElementValidator) Validate(elem *domain.Element) ValidationResult {
    result := ValidationResult{Valid: true}
    
    // Name validation
    if !validNamePattern.MatchString(elem.Name) {
        result.AddError("invalid name: must be kebab-case")
    }
    
    // Size validation
    if len(elem.Content) > MaxContentSize {
        result.AddError("content exceeds 1MB limit")
    }
    
    // Security validation
    if containsDangerousPatterns(elem.Content) {
        result.AddError("content contains dangerous patterns")
    }
    
    // YAML bomb detection
    if isYAMLBomb(elem.Content) {
        result.AddError("potential YAML bomb detected")
    }
    
    // Prototype pollution check
    if containsPrototypePollution(elem.Metadata) {
        result.AddError("metadata contains prototype pollution attempt")
    }
    
    // Type-specific validation
    switch elem.Type {
    case domain.PersonaElement:
        result.Merge(v.validatePersona(elem))
    case domain.SkillElement:
        result.Merge(v.validateSkill(elem))
    // ...
    }
    
    return result
}
```

---

### 3. `edit_element`

**Descrição:** Edita um elemento existente

**Entrada:**
```json
{
  "id": "persona_technical-writer_bob_20251218-120530",
  "fields": {
    "description": "Updated description",
    "metadata.category": "documentation"
  }
}
```

**Saída:**
```json
{
  "element": {
    "id": "persona_technical-writer_bob_20251218-120530",
    "version": "1.0.1",
    "updated_at": "2025-12-18T13:00:00Z"
  }
}
```

---

### 4. `delete_element`

**Descrição:** Remove um elemento

**Entrada:**
```json
{
  "id": "persona_technical-writer_bob_20251218-120530",
  "type": "personas"
}
```

**Saída:**
```json
{
  "deleted": true,
  "message": "Element deleted successfully"
}
```

---

### 5. `activate_element`

**Descrição:** Ativa um elemento (carrega no contexto)

**Entrada:**
```json
{
  "id": "persona_creative-writer_alice_20251218-120000",
  "type": "personas"
}
```

**Saída:**
```json
{
  "activated": true,
  "element": {
    "id": "persona_creative-writer_alice_20251218-120000",
    "name": "creative-writer",
    "status": "active"
  }
}
```

**Implementação com Token Budget:**
```go
type ElementActivator struct {
    maxTokens int // e.g., 10000
    active    map[string]*domain.Element
}

func (a *ElementActivator) Activate(ctx context.Context, elem *domain.Element) error {
    // 1. Calculate token cost
    tokens := a.estimateTokens(elem)
    
    // 2. Check budget
    currentUsage := a.getCurrentUsage()
    if currentUsage+tokens > a.maxTokens {
        return errors.New("token budget exceeded")
    }
    
    // 3. Activate
    a.active[elem.ID] = elem
    
    // 4. Emit event
    a.eventBus.Publish(ElementActivatedEvent{Element: elem})
    
    return nil
}
```

---

### 6. `deactivate_element`

**Descrição:** Desativa um elemento

**Entrada:**
```json
{
  "id": "persona_creative-writer_alice_20251218-120000",
  "type": "personas"
}
```

---

### 7. `get_active_elements`

**Descrição:** Lista elementos atualmente ativos

**Entrada:**
```json
{
  "type": "personas"  // opcional
}
```

**Saída:**
```json
{
  "active_elements": [
    {
      "id": "persona_creative-writer_alice_20251218-120000",
      "name": "creative-writer",
      "type": "personas",
      "tokens_used": 2500
    }
  ],
  "total_tokens": 2500,
  "budget": 10000
}
```

---

### 8. `get_element_details`

**Descrição:** Obtém detalhes completos de um elemento

**Entrada:**
```json
{
  "id": "persona_creative-writer_alice_20251218-120000",
  "type": "personas"
}
```

**Saída:**
```json
{
  "element": {
    "id": "persona_creative-writer_alice_20251218-120000",
    "name": "creative-writer",
    "type": "personas",
    "version": "1.0.0",
    "author": "alice",
    "description": "Imaginative storyteller",
    "content": "# Creative Writer\n\n...",
    "metadata": {
      "behavioral_traits": ["imaginative", "engaging"],
      "category": "writing"
    },
    "created_at": "2025-12-18T12:00:00Z",
    "updated_at": "2025-12-18T12:00:00Z",
    "file_path": "/home/user/.dollhouse/portfolio/personas/creative-writer.md",
    "file_size": 2048,
    "token_estimate": 2500
  }
}
```

---

### 9. `validate_element`

**Descrição:** Valida um elemento sem salvá-lo

**Entrada:**
```json
{
  "type": "personas",
  "name": "test-persona",
  "content": "# Test\n\n...",
  "strict": true
}
```

**Saída:**
```json
{
  "valid": true,
  "errors": [],
  "warnings": [
    "Description is empty"
  ],
  "checks_passed": 287,
  "checks_total": 300
}
```

---

### 10. `reload_elements`

**Descrição:** Recarrega elementos do filesystem

**Entrada:**
```json
{
  "type": "personas"  // opcional
}
```

---

### 11. `render_template`

**Descrição:** Renderiza um template com variáveis

**Entrada:**
```json
{
  "id": "template_project-proposal_alice_20251218-120000",
  "variables": {
    "project_name": "MCP Server Go",
    "budget": "$50000",
    "timeline": "18 weeks"
  }
}
```

**Saída:**
```json
{
  "rendered": "# Project Proposal: MCP Server Go\n\nBudget: $50000\n..."
}
```

---

### 12. `execute_agent`

**Descrição:** Executa um agente com um objetivo

**Entrada:**
```json
{
  "id": "agent_code-reviewer_bob_20251218-120000",
  "goal": "Review pull request #42 for code quality",
  "context": {
    "pr_url": "https://github.com/...",
    "files_changed": 5
  }
}
```

**Saída:**
```json
{
  "result": {
    "completed": true,
    "actions_taken": [
      "Analyzed 5 files",
      "Found 3 issues",
      "Generated review comments"
    ],
    "output": "# Code Review\n\n..."
  }
}
```

---

## Private Personas Tools

### 13. `create_private_persona`

**Descrição:** Cria persona privada no diretório do usuário

**Entrada:**
```json
{
  "name": "work-assistant",
  "description": "Professional work helper",
  "content": "# Work Assistant\n\nExpertise in project management...",
  "behavioral_traits": ["professional", "organized"],
  "expertise_areas": ["project-management", "communication"],
  "privacy_level": "private",
  "template_id": "developer-template" // optional
}
```

**Saída:**
```json
{
  "id": "persona_work-assistant_alice_20251218-120000",
  "path": "personas/private-alice/work-assistant.md",
  "owner": "alice",
  "created_from_template": true
}
```

**Implementação Go:**
```go
type CreatePrivatePersonaInput struct {
    Name             string   `json:"name" jsonschema:"required,minLength=3,maxLength=50" validate:"required,min=3,max=50"`
    Description      string   `json:"description" jsonschema:"required" validate:"required"`
    Content          string   `json:"content" jsonschema:"required" validate:"required"`
    BehavioralTraits []string `json:"behavioral_traits" jsonschema:"description=Behavioral traits"`
    ExpertiseAreas   []string `json:"expertise_areas" jsonschema:"description=Areas of expertise"`
    PrivacyLevel     string   `json:"privacy_level" jsonschema:"enum=private|shared,default=private" validate:"oneof=private shared"`
    TemplateID       string   `json:"template_id" jsonschema:"description=Optional template to use"`
}

func (h *PrivatePersonaHandler) CreatePrivatePersona(ctx context.Context, input CreatePrivatePersonaInput) (*CreatePersonaOutput, error) {
    // 1. Get current user from context
    user := auth.UserFromContext(ctx)
    
    // 2. If template specified, load and merge
    var content string
    if input.TemplateID != "" {
        template, err := h.templateRepo.FindByID(ctx, input.TemplateID)
        if err != nil {
            return nil, fmt.Errorf("template not found: %w", err)
        }
        content = h.mergeTemplate(template, input)
    } else {
        content = input.Content
    }
    
    // 3. Create persona in user's private directory
    persona := &domain.Persona{
        Element: domain.Element{
            ID:          generateID("persona", input.Name, user.Username),
            Type:        domain.PersonaElement,
            Name:        input.Name,
            Description: input.Description,
            Content:     content,
            Author:      user.Username,
        },
        BehavioralTraits: input.BehavioralTraits,
        ExpertiseAreas:   input.ExpertiseAreas,
        PrivacyLevel:     domain.PrivacyLevel(input.PrivacyLevel),
        Owner:            user.Username,
        TemplateID:       input.TemplateID,
    }
    
    // 4. Save to private directory
    path := fmt.Sprintf("personas/private-%s/%s.md", user.Username, input.Name)
    err := h.repo.Save(ctx, persona, path)
    if err != nil {
        return nil, fmt.Errorf("failed to save: %w", err)
    }
    
    return &CreatePersonaOutput{
        ID:                  persona.ID,
        Path:                path,
        Owner:               user.Username,
        CreatedFromTemplate: input.TemplateID != "",
    }, nil
}
```

---

### 14. `share_persona`

**Descrição:** Compartilha persona privada com outros usuários

**Entrada:**
```json
{
  "persona_id": "persona_work-assistant_alice_20251218-120000",
  "shared_with": ["bob", "charlie"],
  "permissions": {"read": true, "fork": true, "edit": false}
}
```

**Saída:**
```json
{
  "share_url": "mcp://personas/shared/work-assistant-alice",
  "shared_with": ["bob", "charlie"],
  "permissions": {"read": true, "fork": true, "edit": false}
}
```

---

### 15. `fork_persona`

**Descrição:** Cria cópia privada de persona compartilhada

**Entrada:**
```json
{
  "source_persona_id": "persona_creative-writer_alice_20251218-120000",
  "new_name": "my-creative-writer",
  "customizations": {
    "behavioral_traits": ["creative", "technical"],
    "expertise_areas": ["coding", "writing"]
  }
}
```

**Saída:**
```json
{
  "id": "persona_my-creative-writer_bob_20251218-130000",
  "path": "personas/private-bob/my-creative-writer.md",
  "forked_from": "persona_creative-writer_alice_20251218-120000",
  "fork_attribution": "Forked from alice/creative-writer"
}
```

**Implementação Go:**
```go
type ForkPersonaInput struct {
    SourcePersonaID string                 `json:"source_persona_id" jsonschema:"required" validate:"required"`
    NewName         string                 `json:"new_name" jsonschema:"required" validate:"required"`
    Customizations  map[string]interface{} `json:"customizations" jsonschema:"description=Custom fields to override"`
}

func (h *PrivatePersonaHandler) ForkPersona(ctx context.Context, input ForkPersonaInput) (*ForkPersonaOutput, error) {
    user := auth.UserFromContext(ctx)
    
    // 1. Load source persona (check read permission)
    source, err := h.repo.FindByID(ctx, input.SourcePersonaID)
    if err != nil {
        return nil, fmt.Errorf("source not found: %w", err)
    }
    
    // 2. Check fork permission
    if !h.authz.CanFork(user, source) {
        return nil, fmt.Errorf("no fork permission")
    }
    
    // 3. Create fork with customizations
    fork := source.Clone()
    fork.ID = generateID("persona", input.NewName, user.Username)
    fork.Name = input.NewName
    fork.Owner = user.Username
    fork.ForkedFrom = input.SourcePersonaID
    fork.PrivacyLevel = domain.PrivacyPrivate
    
    // Apply customizations
    h.applyCustomizations(fork, input.Customizations)
    
    // 4. Save to user's private directory
    path := fmt.Sprintf("personas/private-%s/%s.md", user.Username, input.NewName)
    err = h.repo.Save(ctx, fork, path)
    if err != nil {
        return nil, err
    }
    
    return &ForkPersonaOutput{
        ID:              fork.ID,
        Path:            path,
        ForkedFrom:      input.SourcePersonaID,
        ForkAttribution: fmt.Sprintf("Forked from %s/%s", source.Owner, source.Name),
    }, nil
}
```

---

### 16. `bulk_import_personas`

**Descrição:** Importa múltiplas personas de CSV/JSON

**Entrada:**
```json
{
  "format": "csv",
  "data": "name,description,behavioral_traits\ndev-helper,Development assistant,technical|precise\nwriter,Content creator,creative|empathetic",
  "privacy_level": "private",
  "duplicate_handling": "skip" // skip, overwrite, rename
}
```

**Saída:**
```json
{
  "imported": 2,
  "skipped": 0,
  "errors": [],
  "personas": [
    {"id": "persona_dev-helper_alice_20251218-120000", "status": "created"},
    {"id": "persona_writer_alice_20251218-120001", "status": "created"}
  ]
}
```

---

### 17. `bulk_update_personas`

**Descrição:** Atualiza múltiplas personas baseado em filtro

**Entrada:**
```json
{
  "filter": {
    "owner": "alice",
    "tags": ["work"]
  },
  "updates": {
    "behavioral_traits": ["professional", "organized"],
    "metadata.department": "engineering"
  }
}
```

**Saída:**
```json
{
  "updated": 5,
  "personas": [
    "persona_work-assistant_alice_20251218-120000",
    "persona_project-manager_alice_20251218-110000"
  ]
}
```

---

### 18. `search_personas_advanced`

**Descrição:** Busca avançada com múltiplos critérios

**Entrada:**
```json
{
  "query": "developer assistant",
  "filters": {
    "owner": "alice",
    "tags": ["technical", "coding"],
    "date_range": {"from": "2025-12-01", "to": "2025-12-31"},
    "privacy_level": "private"
  },
  "search_mode": "fuzzy", // exact, fuzzy, regex
  "sort": "relevance", // relevance, date, name
  "limit": 20
}
```

**Saída:**
```json
{
  "results": [
    {
      "id": "persona_dev-helper_alice_20251218-120000",
      "name": "dev-helper",
      "relevance_score": 0.95,
      "match_reason": "Fuzzy match on 'developer' (Levenshtein distance: 1)"
    }
  ],
  "total": 1,
  "search_time_ms": 15
}
```

**Implementação Go:**
```go
type SearchPersonasAdvancedInput struct {
    Query   string        `json:"query" jsonschema:"required" validate:"required"`
    Filters SearchFilters `json:"filters"`
    SearchMode string     `json:"search_mode" jsonschema:"enum=exact|fuzzy|regex,default=fuzzy" validate:"oneof=exact fuzzy regex"`
    Sort    string        `json:"sort" jsonschema:"enum=relevance|date|name,default=relevance" validate:"oneof=relevance date name"`
    Limit   int           `json:"limit" jsonschema:"minimum=1,maximum=100,default=20" validate:"min=1,max=100"`
}

type SearchFilters struct {
    Owner        string    `json:"owner"`
    Tags         []string  `json:"tags"`
    DateRange    DateRange `json:"date_range"`
    PrivacyLevel string    `json:"privacy_level" validate:"omitempty,oneof=public private shared"`
}

func (h *PrivatePersonaHandler) SearchPersonasAdvanced(ctx context.Context, input SearchPersonasAdvancedInput) (*SearchResults, error) {
    user := auth.UserFromContext(ctx)
    
    // 1. Build search query with filters
    query := h.searchBuilder.NewQuery()
    query.Text(input.Query)
    
    // Apply filters
    if input.Filters.Owner != "" {
        query.Filter("owner", input.Filters.Owner)
    }
    if len(input.Filters.Tags) > 0 {
        query.Filter("tags", input.Filters.Tags)
    }
    
    // 2. Execute search based on mode
    var results []SearchResult
    switch input.SearchMode {
    case "fuzzy":
        results = h.fuzzySearch(ctx, query, 2) // Levenshtein distance ≤ 2
    case "regex":
        results = h.regexSearch(ctx, query)
    default:
        results = h.exactSearch(ctx, query)
    }
    
    // 3. Filter by permissions (user can only see own private + shared)
    results = h.filterByPermissions(user, results)
    
    // 4. Score and sort
    results = h.scoreResults(results, input.Query)
    results = h.sortResults(results, input.Sort)
    
    // 5. Limit
    if len(results) > input.Limit {
        results = results[:input.Limit]
    }
    
    return &SearchResults{
        Results:      results,
        Total:        len(results),
        SearchTimeMs: time.Since(start).Milliseconds(),
    }, nil
}
```

---

### 19. `list_persona_versions`

**Descrição:** Lista histórico de versões de uma persona

**Entrada:**
```json
{
  "persona_id": "persona_work-assistant_alice_20251218-120000"
}
```

**Saída:**
```json
{
  "versions": [
    {
      "version": 3,
      "timestamp": "2025-12-18T14:30:00Z",
      "author": "alice",
      "message": "Added project management skills",
      "content_hash": "abc123..."
    },
    {
      "version": 2,
      "timestamp": "2025-12-18T12:15:00Z",
      "author": "alice",
      "message": "Updated behavioral traits",
      "content_hash": "def456..."
    }
  ],
  "current_version": 3
}
```

---

### 20. `diff_persona_versions`

**Descrição:** Compara duas versões de uma persona

**Entrada:**
```json
{
  "persona_id": "persona_work-assistant_alice_20251218-120000",
  "version1": 2,
  "version2": 3
}
```

**Saída:**
```json
{
  "diff": {
    "behavioral_traits": {
      "added": ["organized"],
      "removed": ["casual"]
    },
    "expertise_areas": {
      "added": ["project-management"],
      "removed": []
    },
    "content": "+ Added section on project management\n- Removed casual tone references"
  },
  "change_summary": "2 fields modified, 3 additions, 1 deletion"
}
```

---

## Collection Tools

### 13. `browse_collection`

**Descrição:** Navega pela coleção da comunidade

**Entrada:**
```json
{
  "section": "library|showcase|catalog",
  "type": "personas|skills|templates|agents|memories"
}
```

**Saída:**
```json
{
  "items": [
    {
      "id": "persona_debug-detective",
      "name": "Debug Detective",
      "author": "community",
      "description": "Systematic problem-solver",
      "rating": 4.5,
      "downloads": 1250,
      "category": "debugging"
    }
  ],
  "total": 500
}
```

---

### 14. `search_collection`

**Descrição:** Busca na coleção da comunidade

**Entrada:**
```json
{
  "query": "python debugging",
  "type": "skills",
  "filters": {
    "min_rating": 4.0,
    "category": "debugging"
  }
}
```

---

### 15. `install_collection_content`

**Descrição:** Instala conteúdo da coleção

**Entrada:**
```json
{
  "path": "personas/debug-detective.md",
  "overwrite": false
}
```

---

### 16. `get_collection_content`

**Descrição:** Obtém detalhes de conteúdo da coleção

**Entrada:**
```json
{
  "path": "personas/debug-detective.md"
}
```

---

### 17. `submit_to_collection`

**Descrição:** Envia elemento para a coleção

**Entrada:**
```json
{
  "id": "persona_custom-writer_alice_20251218-120000",
  "category": "writing",
  "tags": ["creative", "storytelling"]
}
```

---

## Portfolio Tools

### 18. `portfolio_status`

**Descrição:** Status do portfolio

**Entrada:** `{}`

**Saída:**
```json
{
  "location": "/home/user/.dollhouse/portfolio",
  "github_connected": true,
  "github_repo": "alice/dollhouse-portfolio",
  "last_sync": "2025-12-18T11:00:00Z",
  "elements": {
    "personas": 12,
    "skills": 25,
    "templates": 8,
    "agents": 5,
    "memories": 150,
    "ensembles": 2
  },
  "total_elements": 202,
  "storage_used": "15.2 MB"
}
```

---

### 19. `sync_portfolio`

**Descrição:** Sincroniza portfolio com GitHub

**Entrada:**
```json
{
  "direction": "push|pull|bidirectional",
  "force": false
}
```

**Saída:**
```json
{
  "synced": true,
  "changes": {
    "pushed": 5,
    "pulled": 2,
    "conflicts": 0
  },
  "duration": "2.3s"
}
```

**Implementação com Conflict Resolution:**
```go
type PortfolioSyncer struct {
    local  PortfolioRepository
    remote PortfolioRepository
}

func (s *PortfolioSyncer) Sync(ctx context.Context, direction SyncDirection) (*SyncResult, error) {
    switch direction {
    case Bidirectional:
        return s.bidirectionalSync(ctx)
    case Push:
        return s.push(ctx)
    case Pull:
        return s.pull(ctx)
    }
}

func (s *PortfolioSyncer) bidirectionalSync(ctx context.Context) (*SyncResult, error) {
    // 1. Get local changes
    localChanges, err := s.local.GetChanges(ctx)
    if err != nil {
        return nil, err
    }
    
    // 2. Get remote changes
    remoteChanges, err := s.remote.GetChanges(ctx)
    if err != nil {
        return nil, err
    }
    
    // 3. Detect conflicts
    conflicts := s.detectConflicts(localChanges, remoteChanges)
    if len(conflicts) > 0 {
        return nil, fmt.Errorf("conflicts detected: %v", conflicts)
    }
    
    // 4. Push local changes
    if err := s.remote.Apply(ctx, localChanges); err != nil {
        return nil, err
    }
    
    // 5. Pull remote changes
    if err := s.local.Apply(ctx, remoteChanges); err != nil {
        return nil, err
    }
    
    return &SyncResult{
        Pushed:  len(localChanges),
        Pulled:  len(remoteChanges),
        Conflicts: 0,
    }, nil
}
```

---

### 20. `search_portfolio`

**Descrição:** Busca no portfolio local

**Entrada:**
```json
{
  "query": "debugging",
  "type": "skills",
  "filters": {
    "author": "alice"
  }
}
```

---

### 21. `backup_portfolio`

**Descrição:** Cria backup do portfolio

**Entrada:**
```json
{
  "destination": "/backup/dollhouse-2025-12-18.tar.gz",
  "compress": true
}
```

---

### 22. `restore_portfolio`

**Descrição:** Restaura portfolio de backup

---

### 23. `export_element`

**Descrição:** Exporta elemento para arquivo

---

### 24. `import_element`

**Descrição:** Importa elemento de arquivo

---

### 25. `portfolio_element_manager`

**Descrição:** Gerencia sincronização de elementos individuais

---

## Search Tools

### 26. `unified_search`

**Descrição:** Busca unificada em todas as fontes

**Entrada:**
```json
{
  "query": "code review automation",
  "sources": ["local", "github", "collection"],
  "limit": 20
}
```

**Saída:**
```json
{
  "results": [
    {
      "id": "skill_code-review_bob_20251218-120000",
      "name": "Code Review Automation",
      "source": "local",
      "score": 0.95,
      "type": "skills",
      "snippet": "Automated code quality checks..."
    }
  ],
  "total": 42,
  "search_time": "15ms"
}
```

**Implementação com 3-Tier Index:**
```go
type UnifiedSearchEngine struct {
    localIndex      *InvertedIndex
    githubIndex     *InvertedIndex
    collectionIndex *InvertedIndex
    nlpScorer       *NLPScorer
}

func (e *UnifiedSearchEngine) Search(ctx context.Context, query string, sources []string) ([]SearchResult, error) {
    // 1. Tokenize query
    tokens := e.tokenize(query)
    
    // 2. Search each source in parallel
    var wg sync.WaitGroup
    resultsChan := make(chan []SearchResult, len(sources))
    
    for _, source := range sources {
        wg.Add(1)
        go func(src string) {
            defer wg.Done()
            
            var results []SearchResult
            switch src {
            case "local":
                results = e.localIndex.Search(tokens)
            case "github":
                results = e.githubIndex.Search(tokens)
            case "collection":
                results = e.collectionIndex.Search(tokens)
            }
            
            resultsChan <- results
        }(source)
    }
    
    // 3. Wait for all searches
    go func() {
        wg.Wait()
        close(resultsChan)
    }()
    
    // 4. Merge and rank results
    var allResults []SearchResult
    for results := range resultsChan {
        allResults = append(allResults, results...)
    }
    
    // 5. Score with NLP (Jaccard + Shannon Entropy)
    for i := range allResults {
        allResults[i].Score = e.nlpScorer.Score(query, allResults[i].Content)
    }
    
    // 6. Sort by score
    sort.Slice(allResults, func(i, j int) bool {
        return allResults[i].Score > allResults[j].Score
    })
    
    return allResults, nil
}
```

---

### 27. `build_search_index`

**Descrição:** Reconstrói índice de busca

---

### 28. `search_statistics`

**Descrição:** Estatísticas do índice de busca

---

### 29. `optimize_index`

**Descrição:** Otimiza índice de busca

---

## Configuration Tools

### 30. `configure_indicator`

**Descrição:** Configura indicadores de elementos ativos

---

### 31. `dollhouse_config`

**Descrição:** Gerencia configurações gerais

**Entrada:**
```json
{
  "operation": "get|set|reset",
  "key": "source_priority",
  "value": ["local", "github", "collection"]
}
```

---

### 32. `set_portfolio_location`

**Descrição:** Define localização do portfolio

---

### 33. `configure_telemetry`

**Descrição:** Configura telemetria

---

### 34. `get_config`

**Descrição:** Obtém configuração atual

---

### 35. `set_config`

**Descrição:** Define configuração

---

## Security Tools

### 36. `validate_security`

**Descrição:** Valida elemento contra 300+ regras de segurança

**Entrada:**
```json
{
  "id": "persona_test_alice_20251218-120000",
  "checks": ["injection", "traversal", "yaml_bomb", "prototype_pollution"]
}
```

**Saída:**
```json
{
  "secure": true,
  "vulnerabilities": [],
  "checks_passed": 300,
  "checks_total": 300,
  "report": {
    "path_traversal": "PASS",
    "command_injection": "PASS",
    "yaml_bomb": "PASS",
    "prototype_pollution": "PASS"
  }
}
```

---

### 37. `scan_portfolio_security`

**Descrição:** Escaneia todo portfolio por vulnerabilidades

---

### 38. `encrypt_sensitive_data`

**Descrição:** Encripta dados sensíveis (tokens, etc.)

---

### 39. `audit_log`

**Descrição:** Registra eventos de auditoria

---

## Capability Index Tools

### 40. `get_capability_index`

**Descrição:** Obtém índice de capacidades

**Entrada:**
```json
{
  "variant": "summary|full|stats"
}
```

**Saída (summary):**
```json
{
  "capabilities": [
    {
      "id": "debugging",
      "triggers": ["debug", "error", "bug"],
      "elements": {
        "personas": ["debug-detective"],
        "skills": ["error-analysis", "stack-trace-reader"]
      },
      "confidence": 0.95,
      "token_cost": 2500
    }
  ],
  "total_capabilities": 42,
  "token_estimate": 3500
}
```

---

### 41. `refresh_capability_index`

**Descrição:** Atualiza índice de capacidades

---

## Resumo de Implementação

### Priorização (MVP)

**Fase 1 - Core (Semanas 1-4):**
1. list_elements
2. create_element
3. get_element_details
4. activate_element
5. deactivate_element

**Fase 2 - CRUD (Semanas 5-6):**
6. edit_element
7. delete_element
8. validate_element

**Fase 3 - Portfolio (Semanas 7-8):**
9. portfolio_status
10. sync_portfolio
11. search_portfolio

**Fase 4 - Collection (Semanas 9-10):**
12. browse_collection
13. install_collection_content

**Fase 5 - Advanced (Semanas 11-14):**
14. unified_search
15. capability_index
16. security_scan

**Fase 6 - Polish (Semanas 15-18):**
17. All remaining tools
18. Performance tuning
19. Documentation

---

**Próximo Documento:** [Plano de Testes](./TESTING_PLAN.md)
