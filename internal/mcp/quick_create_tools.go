package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/indexing"
)

// QuickCreatePersonaInput defines simplified input for quick persona creation.
type QuickCreatePersonaInput struct {
	Name        string   `json:"name"                  jsonschema:"persona name (required)"`
	Description string   `json:"description,omitempty" jsonschema:"brief description"`
	Expertise   []string `json:"expertise,omitempty"   jsonschema:"areas of expertise"`
	Template    string   `json:"template,omitempty"    jsonschema:"template: technical, creative, business, support (applies defaults)"`
}

// QuickCreateSkillInput defines simplified input for quick skill creation.
type QuickCreateSkillInput struct {
	Name        string `json:"name"                  jsonschema:"skill name (required)"`
	Description string `json:"description,omitempty" jsonschema:"brief description"`
	Trigger     string `json:"trigger,omitempty"     jsonschema:"when to activate this skill"`
	Template    string `json:"template,omitempty"    jsonschema:"template: api, data, automation, analysis (applies defaults)"`
}

// QuickCreateMemoryInput defines simplified input for quick memory creation.
type QuickCreateMemoryInput struct {
	Name       string   `json:"name"                 jsonschema:"memory name (required)"`
	Content    string   `json:"content"              jsonschema:"memory content (required)"`
	Tags       []string `json:"tags,omitempty"       jsonschema:"tags for categorization"`
	Importance string   `json:"importance,omitempty" jsonschema:"importance: low, medium, high, critical"`
}

// QuickCreateTemplateInput defines simplified input for quick template creation.
type QuickCreateTemplateInput struct {
	Name        string   `json:"name"                  jsonschema:"template name (required)"`
	Content     string   `json:"content"               jsonschema:"template content (required)"`
	Description string   `json:"description,omitempty" jsonschema:"brief description"`
	Variables   []string `json:"variables,omitempty"   jsonschema:"variable names to extract from content"`
}

// QuickCreateAgentInput defines simplified input for quick agent creation.
type QuickCreateAgentInput struct {
	Name        string   `json:"name"                  jsonschema:"agent name (required)"`
	Goal        string   `json:"goal"                  jsonschema:"agent goal (required)"`
	Description string   `json:"description,omitempty" jsonschema:"brief description"`
	Actions     []string `json:"actions,omitempty"     jsonschema:"actions the agent can perform"`
}

// QuickCreateEnsembleInput defines simplified input for quick ensemble creation.
type QuickCreateEnsembleInput struct {
	Name          string   `json:"name"                     jsonschema:"ensemble name (required)"`
	Description   string   `json:"description,omitempty"    jsonschema:"brief description"`
	AgentIDs      []string `json:"agent_ids,omitempty"      jsonschema:"IDs of agents to include"`
	ExecutionMode string   `json:"execution_mode,omitempty" jsonschema:"sequential, parallel, or hybrid (default: sequential)"`
}

// handleQuickCreatePersona handles quick persona creation with minimal input.
func (s *MCPServer) handleQuickCreatePersona(ctx context.Context, req *sdk.CallToolRequest, input QuickCreatePersonaInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	// Apply template defaults
	template := s.getPersonaTemplate(input.Template)

	// Create persona with defaults
	persona := domain.NewPersona(
		input.Name,
		orDefault(input.Description, template.Description),
		"1.0.0",
		getCurrentUserFromContext(ctx),
	)

	// Set behavioral traits from template
	for name, desc := range template.BehavioralTraits {
		persona.BehavioralTraits = append(persona.BehavioralTraits, domain.BehavioralTrait{
			Name:        name,
			Description: desc,
			Intensity:   7, // Default moderate-high intensity
		})
	}

	// Merge expertise
	expertise := input.Expertise
	if len(expertise) == 0 {
		expertise = template.Expertise
	}
	for _, exp := range expertise {
		persona.ExpertiseAreas = append(persona.ExpertiseAreas, domain.ExpertiseArea{
			Domain: exp,
			Level:  "intermediate", // Default level
		})
	}

	// Set response style from template
	persona.ResponseStyle = domain.ResponseStyle{
		Tone:      template.CommunicationTone,
		Formality: template.CommunicationFormality,
		Verbosity: "balanced",
	}

	// Set system prompt
	persona.SystemPrompt = fmt.Sprintf("You are %s, a %s. %s", input.Name, template.Description, orDefault(input.Description, ""))

	// Set default tags
	tags := []string{"quick-create", template.Name}
	metadata := persona.GetMetadata()
	metadata.Tags = tags
	persona.SetMetadata(metadata)

	// Validate
	if err := persona.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save
	if err := s.repo.Create(persona); err != nil {
		return nil, nil, fmt.Errorf("failed to create persona: %w", err)
	}

	// Update index
	s.mu.Lock()
	metadata = persona.GetMetadata()
	s.index.AddDocument(&indexing.Document{
		ID:      metadata.ID,
		Content: metadata.Name + " " + metadata.Description,
	})
	s.mu.Unlock()

	output := map[string]interface{}{
		"id":        metadata.ID,
		"name":      metadata.Name,
		"type":      persona.GetType(),
		"template":  template.Name,
		"message":   "Persona created successfully (quick mode)",
		"file_path": fmt.Sprintf("data/elements/persona/%s/%s.yaml", time.Now().Format("2006-01-02"), persona.GetID()),
	}

	return nil, output, nil
}

// handleQuickCreateSkill handles quick skill creation with minimal input.
func (s *MCPServer) handleQuickCreateSkill(ctx context.Context, req *sdk.CallToolRequest, input QuickCreateSkillInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	// Apply template defaults
	template := s.getSkillTemplate(input.Template)

	// Create skill with defaults
	skill := domain.NewSkill(
		input.Name,
		orDefault(input.Description, template.Description),
		"1.0.0",
		getCurrentUserFromContext(ctx),
	)

	// Set triggers from template or input
	if input.Trigger != "" {
		skill.Triggers = []domain.SkillTrigger{
			{
				Type:     "keyword",
				Pattern:  input.Trigger,
				Keywords: []string{input.Trigger}, // Add keywords for validation
			},
		}
	} else {
		for _, t := range template.Triggers {
			skill.Triggers = append(skill.Triggers, domain.SkillTrigger{
				Type:     "keyword",
				Pattern:  t,
				Keywords: []string{t}, // Add keywords for validation
			})
		}
	}

	// Set procedures from template
	for i, proc := range template.Procedures {
		skill.Procedures = append(skill.Procedures, domain.SkillProcedure{
			Step:        i + 1,
			Action:      proc, // Action is required
			Description: proc,
		})
	}

	// Set default tags
	tags := []string{"quick-create", template.Name}
	metadata := skill.GetMetadata()
	metadata.Tags = tags
	skill.SetMetadata(metadata)

	// Validate
	if err := skill.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save
	if err := s.repo.Create(skill); err != nil {
		return nil, nil, fmt.Errorf("failed to create skill: %w", err)
	}

	// Update index
	s.mu.Lock()
	skillMetadata := skill.GetMetadata()
	s.index.AddDocument(&indexing.Document{
		ID:      skillMetadata.ID,
		Content: skillMetadata.Name + " " + skillMetadata.Description,
	})
	s.mu.Unlock()

	output := map[string]interface{}{
		"id":        skillMetadata.ID,
		"name":      skillMetadata.Name,
		"type":      skill.GetType(),
		"template":  template.Name,
		"message":   "Skill created successfully (quick mode)",
		"file_path": fmt.Sprintf("data/elements/skill/%s/%s.yaml", time.Now().Format("2006-01-02"), skill.GetID()),
	}

	return nil, output, nil
}

// handleQuickCreateMemory handles quick memory creation with minimal input.
func (s *MCPServer) handleQuickCreateMemory(ctx context.Context, req *sdk.CallToolRequest, input QuickCreateMemoryInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	// Create memory
	memory := domain.NewMemory(
		input.Name,
		"", // No description needed for quick create
		"1.0.0",
		getCurrentUserFromContext(ctx),
	)

	// Set content
	memory.Content = input.Content
	memory.ComputeHash()

	// Set tags
	tags := input.Tags
	if tags == nil {
		tags = []string{"quick-create"}
	} else {
		tags = append(tags, "quick-create")
	}

	// Add importance tag
	if input.Importance != "" {
		tags = append(tags, "importance:"+input.Importance)
	}

	metadata := memory.GetMetadata()
	metadata.Tags = tags
	memory.SetMetadata(metadata)

	// Validate
	if err := memory.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save
	if err := s.repo.Create(memory); err != nil {
		return nil, nil, fmt.Errorf("failed to create memory: %w", err)
	}

	// Update index
	s.mu.Lock()
	memoryMetadata := memory.GetMetadata()
	s.index.AddDocument(&indexing.Document{
		ID:      memoryMetadata.ID,
		Content: memoryMetadata.Name + " " + input.Content,
	})
	s.mu.Unlock()

	output := map[string]interface{}{
		"id":        memoryMetadata.ID,
		"name":      memoryMetadata.Name,
		"type":      memory.GetType(),
		"hash":      memory.ContentHash,
		"message":   "Memory created successfully (quick mode)",
		"file_path": fmt.Sprintf("data/elements/memory/%s/%s.yaml", time.Now().Format("2006-01-02"), memory.GetID()),
	}

	return nil, output, nil
}

// handleQuickCreateTemplate handles quick template creation with minimal input.
func (s *MCPServer) handleQuickCreateTemplate(ctx context.Context, req *sdk.CallToolRequest, input QuickCreateTemplateInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	// Create template
	template := domain.NewTemplate(
		input.Name,
		orDefault(input.Description, "Quick-created template"),
		"1.0.0",
		getCurrentUserFromContext(ctx),
	)

	// Set content
	template.Content = input.Content

	// Auto-extract variables from content if not provided
	variables := input.Variables
	if len(variables) == 0 {
		// Simple extraction: look for {{variable}} patterns
		variables = extractVariablesFromContent(input.Content)
	}

	// Set variables
	for _, varName := range variables {
		template.Variables = append(template.Variables, domain.TemplateVariable{
			Name:        varName,
			Description: "Auto-extracted variable",
			Required:    true,
		})
	}

	// Set default tags
	tags := []string{"quick-create"}
	metadata := template.GetMetadata()
	metadata.Tags = tags
	template.SetMetadata(metadata)

	// Validate
	if err := template.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save
	if err := s.repo.Create(template); err != nil {
		return nil, nil, fmt.Errorf("failed to create template: %w", err)
	}

	// Update index
	s.mu.Lock()
	templateMetadata := template.GetMetadata()
	s.index.AddDocument(&indexing.Document{
		ID:      templateMetadata.ID,
		Content: templateMetadata.Name + " " + input.Content,
	})
	s.mu.Unlock()

	output := map[string]interface{}{
		"id":        templateMetadata.ID,
		"name":      templateMetadata.Name,
		"type":      template.GetType(),
		"variables": len(template.Variables),
		"message":   "Template created successfully (quick mode)",
		"file_path": fmt.Sprintf("data/elements/template/%s/%s.yaml", time.Now().Format("2006-01-02"), template.GetID()),
	}

	return nil, output, nil
}

// handleQuickCreateAgent handles quick agent creation with minimal input.
func (s *MCPServer) handleQuickCreateAgent(ctx context.Context, req *sdk.CallToolRequest, input QuickCreateAgentInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	// Create agent
	agent := domain.NewAgent(
		input.Name,
		orDefault(input.Description, "Quick-created agent"),
		"1.0.0",
		getCurrentUserFromContext(ctx),
	)

	// Set goal
	agent.Goals = []string{input.Goal}

	// Set actions
	if len(input.Actions) > 0 {
		for _, action := range input.Actions {
			agent.Actions = append(agent.Actions, domain.AgentAction{
				Name: action,
				Type: "tool",
			})
		}
	} else {
		// Default action
		agent.Actions = []domain.AgentAction{
			{
				Name: "execute",
				Type: "tool",
			},
		}
	}

	// Set max iterations
	agent.MaxIterations = 10

	// Set default tags
	tags := []string{"quick-create"}
	metadata := agent.GetMetadata()
	metadata.Tags = tags
	agent.SetMetadata(metadata)

	// Validate
	if err := agent.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save
	if err := s.repo.Create(agent); err != nil {
		return nil, nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Update index
	s.mu.Lock()
	agentMetadata := agent.GetMetadata()
	s.index.AddDocument(&indexing.Document{
		ID:      agentMetadata.ID,
		Content: agentMetadata.Name + " " + input.Goal,
	})
	s.mu.Unlock()

	output := map[string]interface{}{
		"id":        agentMetadata.ID,
		"name":      agentMetadata.Name,
		"type":      agent.GetType(),
		"goal":      input.Goal,
		"actions":   len(agent.Actions),
		"message":   "Agent created successfully (quick mode)",
		"file_path": fmt.Sprintf("data/elements/agent/%s/%s.yaml", time.Now().Format("2006-01-02"), agent.GetID()),
	}

	return nil, output, nil
}

// handleQuickCreateEnsemble handles quick ensemble creation with minimal input.
func (s *MCPServer) handleQuickCreateEnsemble(ctx context.Context, req *sdk.CallToolRequest, input QuickCreateEnsembleInput) (*sdk.CallToolResult, map[string]interface{}, error) {
	// Create ensemble
	ensemble := domain.NewEnsemble(
		input.Name,
		orDefault(input.Description, "Quick-created ensemble"),
		"1.0.0",
		getCurrentUserFromContext(ctx),
	)

	// Set execution mode
	if input.ExecutionMode != "" {
		ensemble.ExecutionMode = input.ExecutionMode
	}

	// Add members if agent IDs provided
	if len(input.AgentIDs) > 0 {
		for i, agentID := range input.AgentIDs {
			ensemble.Members = append(ensemble.Members, domain.EnsembleMember{
				AgentID:  agentID,
				Role:     fmt.Sprintf("agent-%d", i+1),
				Priority: i + 1,
			})
		}
	} else {
		// Need at least one member for validation
		ensemble.Members = []domain.EnsembleMember{
			{
				AgentID:  "placeholder-agent",
				Role:     "primary",
				Priority: 1,
			},
		}
	}

	// Set aggregation strategy
	ensemble.AggregationStrategy = "consensus"

	// Set default tags
	tags := []string{"quick-create"}
	metadata := ensemble.GetMetadata()
	metadata.Tags = tags
	ensemble.SetMetadata(metadata)

	// Validate
	if err := ensemble.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save
	if err := s.repo.Create(ensemble); err != nil {
		return nil, nil, fmt.Errorf("failed to create ensemble: %w", err)
	}

	// Update index
	s.mu.Lock()
	ensembleMetadata := ensemble.GetMetadata()
	s.index.AddDocument(&indexing.Document{
		ID:      ensembleMetadata.ID,
		Content: ensembleMetadata.Name + " " + ensembleMetadata.Description,
	})
	s.mu.Unlock()

	output := map[string]interface{}{
		"id":             ensembleMetadata.ID,
		"name":           ensembleMetadata.Name,
		"type":           ensemble.GetType(),
		"execution_mode": ensemble.ExecutionMode,
		"members":        len(ensemble.Members),
		"message":        "Ensemble created successfully (quick mode)",
		"file_path":      fmt.Sprintf("data/elements/ensemble/%s/%s.yaml", time.Now().Format("2006-01-02"), ensemble.GetID()),
	}

	return nil, output, nil
}

// Template definitions.
type personaTemplate struct {
	Name                   string
	Description            string
	BehavioralTraits       map[string]string
	Expertise              []string
	CommunicationTone      string
	CommunicationFormality string
}

type skillTemplate struct {
	Name        string
	Description string
	Triggers    []string
	Procedures  []string
}

func (s *MCPServer) getPersonaTemplate(name string) personaTemplate {
	templates := map[string]personaTemplate{
		"technical": {
			Name:        "technical",
			Description: "Technical expert with deep knowledge in software engineering",
			BehavioralTraits: map[string]string{
				"analytical":      "Breaks down complex problems systematically",
				"detail-oriented": "Focuses on precision and accuracy",
				"logical":         "Uses data-driven reasoning",
			},
			Expertise:              []string{"software-engineering", "architecture", "best-practices"},
			CommunicationTone:      "professional",
			CommunicationFormality: "formal",
		},
		"creative": {
			Name:        "creative",
			Description: "Creative thinker focused on innovation and user experience",
			BehavioralTraits: map[string]string{
				"innovative":  "Generates novel solutions",
				"empathetic":  "Considers user perspective",
				"open-minded": "Explores multiple approaches",
			},
			Expertise:              []string{"design-thinking", "user-experience", "innovation"},
			CommunicationTone:      "engaging",
			CommunicationFormality: "casual",
		},
		"business": {
			Name:        "business",
			Description: "Business-oriented professional focused on value and ROI",
			BehavioralTraits: map[string]string{
				"pragmatic":      "Focuses on practical outcomes",
				"results-driven": "Prioritizes measurable impact",
				"strategic":      "Considers long-term implications",
			},
			Expertise:              []string{"business-strategy", "roi-analysis", "stakeholder-management"},
			CommunicationTone:      "business",
			CommunicationFormality: "formal",
		},
		"support": {
			Name:        "support",
			Description: "Supportive assistant focused on helping and teaching",
			BehavioralTraits: map[string]string{
				"patient":     "Takes time to explain thoroughly",
				"encouraging": "Provides positive reinforcement",
				"educational": "Focuses on teaching and learning",
			},
			Expertise:              []string{"customer-support", "education", "problem-solving"},
			CommunicationTone:      "friendly",
			CommunicationFormality: "casual",
		},
	}

	// Default template if not found
	if template, ok := templates[name]; ok {
		return template
	}
	return templates["technical"]
}

func (s *MCPServer) getSkillTemplate(name string) skillTemplate {
	templates := map[string]skillTemplate{
		"api": {
			Name:        "api",
			Description: "API integration and consumption skill",
			Triggers:    []string{"api integration needed", "external service call", "REST API"},
			Procedures: []string{
				"Analyze API documentation",
				"Design request/response handling",
				"Implement error handling",
				"Add rate limiting and retries",
				"Write integration tests",
			},
		},
		"data": {
			Name:        "data",
			Description: "Data processing and transformation skill",
			Triggers:    []string{"data processing needed", "ETL", "data transformation"},
			Procedures: []string{
				"Analyze data structure and format",
				"Design transformation pipeline",
				"Implement data validation",
				"Handle edge cases and errors",
				"Add monitoring and logging",
			},
		},
		"automation": {
			Name:        "automation",
			Description: "Task automation and scripting skill",
			Triggers:    []string{"repetitive task", "automation opportunity", "workflow"},
			Procedures: []string{
				"Identify automation opportunity",
				"Design workflow steps",
				"Implement script/automation",
				"Add error handling and recovery",
				"Test and validate automation",
			},
		},
		"analysis": {
			Name:        "analysis",
			Description: "Code and system analysis skill",
			Triggers:    []string{"code review", "performance analysis", "debugging"},
			Procedures: []string{
				"Gather context and requirements",
				"Analyze code/system structure",
				"Identify issues or improvements",
				"Provide recommendations",
				"Document findings",
			},
		},
	}

	// Default template if not found
	if template, ok := templates[name]; ok {
		return template
	}
	return templates["analysis"]
}

// Helper functions.
func orDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const userContextKey contextKey = "user"

func getCurrentUserFromContext(ctx context.Context) string {
	// Try to get from context
	if user, ok := ctx.Value(userContextKey).(string); ok && user != "" {
		return user
	}
	return "system"
}

func extractVariablesFromContent(content string) []string {
	// Simple extraction of {{variable}} patterns
	var variables []string
	seen := make(map[string]bool)

	// Look for patterns like {{var}}, ${var}, {var}
	patterns := []string{
		"{{", "}}", // Mustache/Handlebars style
		"${", "}", // Shell/ES6 style
	}

	for i := 0; i < len(patterns); i += 2 {
		start := patterns[i]
		end := patterns[i+1]

		searchPos := 0
		for {
			startPos := -1
			for j := searchPos; j < len(content)-len(start); j++ {
				if content[j:j+len(start)] == start {
					startPos = j
					break
				}
			}

			if startPos == -1 {
				break
			}

			endPos := -1
			for j := startPos + len(start); j < len(content)-len(end); j++ {
				if content[j:j+len(end)] == end {
					endPos = j
					break
				}
			}

			if endPos == -1 {
				break
			}

			varName := content[startPos+len(start) : endPos]
			// Clean up whitespace
			varName = strings.TrimSpace(varName)

			if varName != "" && !seen[varName] {
				variables = append(variables, varName)
				seen[varName] = true
			}

			searchPos = endPos + len(end)
		}
	}

	return variables
}
