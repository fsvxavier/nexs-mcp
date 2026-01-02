package mcp

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

// --- Persona Type-Specific Handlers ---

// CreatePersonaInput defines input for create_persona tool.
type CreatePersonaInput struct {
	Name             string                   `json:"name"                        jsonschema:"persona name (3-100 characters)"`
	Description      string                   `json:"description,omitempty"       jsonschema:"persona description (max 500 characters)"`
	Version          string                   `json:"version"                     jsonschema:"persona version (semver)"`
	Author           string                   `json:"author"                      jsonschema:"persona author"`
	Tags             []string                 `json:"tags,omitempty"              jsonschema:"persona tags"`
	SystemPrompt     string                   `json:"system_prompt"               jsonschema:"system prompt (10-2000 characters)"`
	BehavioralTraits []domain.BehavioralTrait `json:"behavioral_traits,omitempty" jsonschema:"behavioral traits with intensity 1-10"`
	ExpertiseAreas   []domain.ExpertiseArea   `json:"expertise_areas,omitempty"   jsonschema:"expertise areas with skill levels"`
	ResponseStyle    *domain.ResponseStyle    `json:"response_style,omitempty"    jsonschema:"response style configuration"`
	PrivacyLevel     string                   `json:"privacy_level,omitempty"     jsonschema:"privacy level: public, private, shared"`
}

// handleCreatePersona handles create_persona tool calls.
func (s *MCPServer) handleCreatePersona(ctx context.Context, req *sdk.CallToolRequest, input CreatePersonaInput) (*sdk.CallToolResult, CreateElementOutput, error) {
	// Validate required fields
	if input.Name == "" || len(input.Name) < 3 || len(input.Name) > 100 {
		return nil, CreateElementOutput{}, errors.New("name must be between 3 and 100 characters")
	}
	if input.Version == "" {
		return nil, CreateElementOutput{}, errors.New("version is required")
	}
	if input.Author == "" {
		return nil, CreateElementOutput{}, errors.New("author is required")
	}
	if input.SystemPrompt == "" || len(input.SystemPrompt) < 10 || len(input.SystemPrompt) > 2000 {
		return nil, CreateElementOutput{}, errors.New("system_prompt must be between 10 and 2000 characters")
	}

	// Create Persona
	persona := domain.NewPersona(input.Name, input.Description, input.Version, input.Author)

	// Set system prompt
	if err := persona.SetSystemPrompt(input.SystemPrompt); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("invalid system_prompt: %w", err)
	}

	// Add behavioral traits
	for _, trait := range input.BehavioralTraits {
		if err := persona.AddBehavioralTrait(trait); err != nil {
			return nil, CreateElementOutput{}, fmt.Errorf("invalid behavioral_trait: %w", err)
		}
	}

	// Add expertise areas
	for _, area := range input.ExpertiseAreas {
		if err := persona.AddExpertiseArea(area); err != nil {
			return nil, CreateElementOutput{}, fmt.Errorf("invalid expertise_area: %w", err)
		}
	}

	// Set response style
	if input.ResponseStyle != nil {
		if err := persona.SetResponseStyle(*input.ResponseStyle); err != nil {
			return nil, CreateElementOutput{}, fmt.Errorf("invalid response_style: %w", err)
		}
	}

	// Set privacy level
	if input.PrivacyLevel != "" {
		privacyLevel := domain.PersonaPrivacyLevel(input.PrivacyLevel)
		if err := persona.SetPrivacyLevel(privacyLevel); err != nil {
			return nil, CreateElementOutput{}, fmt.Errorf("invalid privacy_level: %w", err)
		}
	}

	// Set tags if provided
	if len(input.Tags) > 0 {
		metadata := persona.GetMetadata()
		metadata.Tags = input.Tags
		persona.SetMetadata(metadata)
	}

	// Validate complete persona
	if err := persona.Validate(); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("persona validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(persona); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("failed to create persona: %w", err)
	}

	// Prepare output
	output := CreateElementOutput{
		ID:      persona.GetID(),
		Element: persona.GetMetadata().ToMap(),
	}

	// Auto-extract skills if configured (synchronous for immediate feedback)
	if s.cfg != nil && s.cfg.SkillExtraction.Enabled && s.cfg.SkillExtraction.AutoExtractOnCreate {
		personaID := persona.GetID()
		extractor := application.NewSkillExtractor(s.repo)
		res, err := extractor.ExtractSkillsFromPersona(ctx, personaID)
		if err != nil {
			logger.Error("Skill extraction failed", "error", err, "persona", personaID)
			// Non-fatal: persona was created successfully, skills failed
			output.Element["skill_extraction_error"] = err.Error()
		} else {
			logger.Info("Skill extraction completed",
				"skills_created", res.SkillsCreated,
				"skills_skipped", res.SkippedDuplicate,
				"persona", personaID)
			output.Element["skills_created"] = res.SkillsCreated
			output.Element["skill_ids"] = res.SkillIDs
			if len(res.Errors) > 0 {
				output.Element["skill_extraction_warnings"] = res.Errors
			}
		}
	}

	return nil, output, nil
}

// --- Skill Type-Specific Handlers ---

// CreateSkillInput defines input for create_skill tool.
type CreateSkillInput struct {
	Name         string                   `json:"name"                   jsonschema:"skill name (3-100 characters)"`
	Description  string                   `json:"description,omitempty"  jsonschema:"skill description (max 500 characters)"`
	Version      string                   `json:"version"                jsonschema:"skill version (semver)"`
	Author       string                   `json:"author"                 jsonschema:"skill author"`
	Tags         []string                 `json:"tags,omitempty"         jsonschema:"skill tags"`
	Triggers     []domain.SkillTrigger    `json:"triggers"               jsonschema:"skill triggers (at least 1 required)"`
	Procedures   []domain.SkillProcedure  `json:"procedures"             jsonschema:"skill procedures (at least 1 required)"`
	Dependencies []domain.SkillDependency `json:"dependencies,omitempty" jsonschema:"skill dependencies"`
}

// handleCreateSkill handles create_skill tool calls.
func (s *MCPServer) handleCreateSkill(ctx context.Context, req *sdk.CallToolRequest, input CreateSkillInput) (*sdk.CallToolResult, CreateElementOutput, error) {
	// Validate required fields
	if input.Name == "" || len(input.Name) < 3 || len(input.Name) > 100 {
		return nil, CreateElementOutput{}, errors.New("name must be between 3 and 100 characters")
	}
	if input.Version == "" {
		return nil, CreateElementOutput{}, errors.New("version is required")
	}
	if input.Author == "" {
		return nil, CreateElementOutput{}, errors.New("author is required")
	}
	if len(input.Triggers) == 0 {
		return nil, CreateElementOutput{}, errors.New("at least one trigger is required")
	}
	if len(input.Procedures) == 0 {
		return nil, CreateElementOutput{}, errors.New("at least one procedure is required")
	}

	// Create Skill
	skill := domain.NewSkill(input.Name, input.Description, input.Version, input.Author)

	// Add triggers
	for _, trigger := range input.Triggers {
		if err := skill.AddTrigger(trigger); err != nil {
			return nil, CreateElementOutput{}, fmt.Errorf("invalid trigger: %w", err)
		}
	}

	// Add procedures
	for _, procedure := range input.Procedures {
		if err := skill.AddProcedure(procedure); err != nil {
			return nil, CreateElementOutput{}, fmt.Errorf("invalid procedure: %w", err)
		}
	}

	// Add dependencies
	for _, dep := range input.Dependencies {
		if err := skill.AddDependency(dep); err != nil {
			return nil, CreateElementOutput{}, fmt.Errorf("invalid dependency: %w", err)
		}
	}

	// Set tags
	if len(input.Tags) > 0 {
		metadata := skill.GetMetadata()
		metadata.Tags = input.Tags
		skill.SetMetadata(metadata)
	}

	// Validate
	if err := skill.Validate(); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("skill validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(skill); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("failed to create skill: %w", err)
	}

	output := CreateElementOutput{
		ID:      skill.GetID(),
		Element: skill.GetMetadata().ToMap(),
	}

	return nil, output, nil
}

// --- Template Type-Specific Handlers ---

// CreateTemplateInput defines input for create_template tool.
type CreateTemplateInput struct {
	Name        string                    `json:"name"                  jsonschema:"template name (3-100 characters)"`
	Description string                    `json:"description,omitempty" jsonschema:"template description (max 500 characters)"`
	Version     string                    `json:"version"               jsonschema:"template version (semver)"`
	Author      string                    `json:"author"                jsonschema:"template author"`
	Tags        []string                  `json:"tags,omitempty"        jsonschema:"template tags"`
	Content     string                    `json:"content"               jsonschema:"template content with {{variables}}"`
	Format      string                    `json:"format"                jsonschema:"format: markdown, yaml, json, text"`
	Variables   []domain.TemplateVariable `json:"variables,omitempty"   jsonschema:"template variables"`
}

// handleCreateTemplate handles create_template tool calls.
func (s *MCPServer) handleCreateTemplate(ctx context.Context, req *sdk.CallToolRequest, input CreateTemplateInput) (*sdk.CallToolResult, CreateElementOutput, error) {
	// Validate required fields
	if input.Name == "" || len(input.Name) < 3 || len(input.Name) > 100 {
		return nil, CreateElementOutput{}, errors.New("name must be between 3 and 100 characters")
	}
	if input.Version == "" {
		return nil, CreateElementOutput{}, errors.New("version is required")
	}
	if input.Author == "" {
		return nil, CreateElementOutput{}, errors.New("author is required")
	}
	if input.Content == "" {
		return nil, CreateElementOutput{}, errors.New("content is required")
	}

	// Create Template
	template := domain.NewTemplate(input.Name, input.Description, input.Version, input.Author)

	// Set content and format
	template.Content = input.Content
	if input.Format != "" {
		template.Format = input.Format
	}

	// Set variables
	template.Variables = input.Variables

	// Set tags
	if len(input.Tags) > 0 {
		metadata := template.GetMetadata()
		metadata.Tags = input.Tags
		template.SetMetadata(metadata)
	}

	// Validate
	if err := template.Validate(); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("template validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(template); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("failed to create template: %w", err)
	}

	output := CreateElementOutput{
		ID:      template.GetID(),
		Element: template.GetMetadata().ToMap(),
	}

	return nil, output, nil
}

// --- Agent Type-Specific Handlers ---

// CreateAgentInput defines input for create_agent tool.
type CreateAgentInput struct {
	Name             string               `json:"name"                        jsonschema:"agent name (3-100 characters)"`
	Description      string               `json:"description,omitempty"       jsonschema:"agent description (max 500 characters)"`
	Version          string               `json:"version"                     jsonschema:"agent version (semver)"`
	Author           string               `json:"author"                      jsonschema:"agent author"`
	Tags             []string             `json:"tags,omitempty"              jsonschema:"agent tags"`
	Goals            []string             `json:"goals"                       jsonschema:"agent goals (at least 1 required)"`
	Actions          []domain.AgentAction `json:"actions"                     jsonschema:"agent actions (at least 1 required)"`
	FallbackStrategy string               `json:"fallback_strategy,omitempty" jsonschema:"fallback strategy"`
	MaxIterations    int                  `json:"max_iterations,omitempty"    jsonschema:"max iterations (1-100, default: 10)"`
}

// handleCreateAgent handles create_agent tool calls.
func (s *MCPServer) handleCreateAgent(ctx context.Context, req *sdk.CallToolRequest, input CreateAgentInput) (*sdk.CallToolResult, CreateElementOutput, error) {
	// Validate required fields
	if input.Name == "" || len(input.Name) < 3 || len(input.Name) > 100 {
		return nil, CreateElementOutput{}, errors.New("name must be between 3 and 100 characters")
	}
	if input.Version == "" {
		return nil, CreateElementOutput{}, errors.New("version is required")
	}
	if input.Author == "" {
		return nil, CreateElementOutput{}, errors.New("author is required")
	}
	if len(input.Goals) == 0 {
		return nil, CreateElementOutput{}, errors.New("at least one goal is required")
	}
	if len(input.Actions) == 0 {
		return nil, CreateElementOutput{}, errors.New("at least one action is required")
	}

	// Create Agent
	agent := domain.NewAgent(input.Name, input.Description, input.Version, input.Author)

	// Set goals and actions
	agent.Goals = input.Goals
	agent.Actions = input.Actions

	// Set fallback strategy
	if input.FallbackStrategy != "" {
		agent.FallbackStrategy = input.FallbackStrategy
	}

	// Set max iterations
	if input.MaxIterations > 0 {
		agent.MaxIterations = input.MaxIterations
	}

	// Set tags
	if len(input.Tags) > 0 {
		metadata := agent.GetMetadata()
		metadata.Tags = input.Tags
		agent.SetMetadata(metadata)
	}

	// Validate
	if err := agent.Validate(); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("agent validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(agent); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("failed to create agent: %w", err)
	}

	output := CreateElementOutput{
		ID:      agent.GetID(),
		Element: agent.GetMetadata().ToMap(),
	}

	return nil, output, nil
}

// --- Memory Type-Specific Handlers ---

// CreateMemoryInput defines input for create_memory tool.
type CreateMemoryInput struct {
	Name        string   `json:"name"                  jsonschema:"memory name (3-100 characters)"`
	Description string   `json:"description,omitempty" jsonschema:"memory description (max 500 characters)"`
	Version     string   `json:"version"               jsonschema:"memory version (semver)"`
	Author      string   `json:"author"                jsonschema:"memory author"`
	Tags        []string `json:"tags,omitempty"        jsonschema:"memory tags"`
	Content     string   `json:"content"               jsonschema:"memory content"`
}

// handleCreateMemory handles create_memory tool calls.
func (s *MCPServer) handleCreateMemory(ctx context.Context, req *sdk.CallToolRequest, input CreateMemoryInput) (*sdk.CallToolResult, CreateElementOutput, error) {
	// Validate required fields
	if input.Name == "" || len(input.Name) < 3 || len(input.Name) > 100 {
		return nil, CreateElementOutput{}, errors.New("name must be between 3 and 100 characters")
	}
	if input.Version == "" {
		return nil, CreateElementOutput{}, errors.New("version is required")
	}
	if input.Author == "" {
		return nil, CreateElementOutput{}, errors.New("author is required")
	}
	if input.Content == "" {
		return nil, CreateElementOutput{}, errors.New("content is required")
	}

	// Create Memory
	memory := domain.NewMemory(input.Name, input.Description, input.Version, input.Author)

	// Set content
	memory.Content = input.Content

	// Compute hash for deduplication
	memory.ComputeHash()

	// Set tags
	if len(input.Tags) > 0 {
		metadata := memory.GetMetadata()
		metadata.Tags = input.Tags
		memory.SetMetadata(metadata)
	}

	// Validate
	if err := memory.Validate(); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("memory validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(memory); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("failed to create memory: %w", err)
	}

	output := CreateElementOutput{
		ID:      memory.GetID(),
		Element: memory.GetMetadata().ToMap(),
	}

	return nil, output, nil
}

// --- Ensemble Type-Specific Handlers ---

// CreateEnsembleInput defines input for create_ensemble tool.
type CreateEnsembleInput struct {
	Name                string                  `json:"name"                     jsonschema:"ensemble name (3-100 characters)"`
	Description         string                  `json:"description,omitempty"    jsonschema:"ensemble description (max 500 characters)"`
	Version             string                  `json:"version"                  jsonschema:"ensemble version (semver)"`
	Author              string                  `json:"author"                   jsonschema:"ensemble author"`
	Tags                []string                `json:"tags,omitempty"           jsonschema:"ensemble tags"`
	Members             []domain.EnsembleMember `json:"members"                  jsonschema:"ensemble members (at least 1 required)"`
	ExecutionMode       string                  `json:"execution_mode,omitempty" jsonschema:"execution mode: sequential, parallel, hybrid"`
	AggregationStrategy string                  `json:"aggregation_strategy"     jsonschema:"aggregation strategy"`
}

// handleCreateEnsemble handles create_ensemble tool calls.
func (s *MCPServer) handleCreateEnsemble(ctx context.Context, req *sdk.CallToolRequest, input CreateEnsembleInput) (*sdk.CallToolResult, CreateElementOutput, error) {
	// Validate required fields
	if input.Name == "" || len(input.Name) < 3 || len(input.Name) > 100 {
		return nil, CreateElementOutput{}, errors.New("name must be between 3 and 100 characters")
	}
	if input.Version == "" {
		return nil, CreateElementOutput{}, errors.New("version is required")
	}
	if input.Author == "" {
		return nil, CreateElementOutput{}, errors.New("author is required")
	}
	if len(input.Members) == 0 {
		return nil, CreateElementOutput{}, errors.New("at least one member is required")
	}
	if input.AggregationStrategy == "" {
		return nil, CreateElementOutput{}, errors.New("aggregation_strategy is required")
	}

	// Create Ensemble
	ensemble := domain.NewEnsemble(input.Name, input.Description, input.Version, input.Author)

	// Set members
	ensemble.Members = input.Members

	// Set execution mode
	if input.ExecutionMode != "" {
		ensemble.ExecutionMode = input.ExecutionMode
	}

	// Set aggregation strategy
	ensemble.AggregationStrategy = input.AggregationStrategy

	// Set tags
	if len(input.Tags) > 0 {
		metadata := ensemble.GetMetadata()
		metadata.Tags = input.Tags
		ensemble.SetMetadata(metadata)
	}

	// Validate
	if err := ensemble.Validate(); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("ensemble validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(ensemble); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("failed to create ensemble: %w", err)
	}

	output := CreateElementOutput{
		ID:      ensemble.GetID(),
		Element: ensemble.GetMetadata().ToMap(),
	}

	return nil, output, nil
}
