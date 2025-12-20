package mcp

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// BatchElementInput defines input for batch element creation
type BatchElementInput struct {
	Type        string                 `json:"type" jsonschema:"required,enum=persona,enum=skill,enum=memory,enum=template,enum=agent,enum=ensemble"`
	Name        string                 `json:"name" jsonschema:"required"`
	Description string                 `json:"description,omitempty"`
	Template    string                 `json:"template,omitempty" jsonschema:"template name for quick creation"`
	Data        map[string]interface{} `json:"data,omitempty" jsonschema:"type-specific data"`
}

// BatchCreateElementsInput defines input for batch creation
type BatchCreateElementsInput struct {
	Elements []BatchElementInput `json:"elements" jsonschema:"required,min=1,max=50"`
	Confirm  bool                `json:"confirm,omitempty" jsonschema:"set to true to skip preview"`
}

// BatchCreateElementsOutput defines output for batch creation
type BatchCreateElementsOutput struct {
	Created    int                  `json:"created"`
	Failed     int                  `json:"failed"`
	Total      int                  `json:"total"`
	Results    []BatchElementResult `json:"results"`
	Summary    string               `json:"summary"`
	DurationMs int64                `json:"duration_ms"`
}

// BatchElementResult defines result for each element
type BatchElementResult struct {
	Index   int                    `json:"index"`
	Type    string                 `json:"type"`
	Name    string                 `json:"name"`
	ID      string                 `json:"id,omitempty"`
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// handleBatchCreateElements handles batch creation of multiple elements
func (s *MCPServer) handleBatchCreateElements(ctx context.Context, req *sdk.CallToolRequest, input BatchCreateElementsInput) (*sdk.CallToolResult, BatchCreateElementsOutput, error) {
	startTime := time.Now()

	// Validate batch size
	if len(input.Elements) == 0 {
		return nil, BatchCreateElementsOutput{}, fmt.Errorf("at least one element required")
	}
	if len(input.Elements) > 50 {
		return nil, BatchCreateElementsOutput{}, fmt.Errorf("maximum 50 elements per batch")
	}

	results := make([]BatchElementResult, 0, len(input.Elements))
	created := 0
	failed := 0

	// Process each element
	for i, elem := range input.Elements {
		result := BatchElementResult{
			Index: i,
			Type:  elem.Type,
			Name:  elem.Name,
		}

		var elementID string
		var err error

		// Create based on type
		switch elem.Type {
		case "persona":
			elementID, err = s.batchCreatePersona(ctx, elem)
		case "skill":
			elementID, err = s.batchCreateSkill(ctx, elem)
		case "memory":
			elementID, err = s.batchCreateMemory(ctx, elem)
		case "template":
			elementID, err = s.batchCreateTemplate(ctx, elem)
		case "agent":
			elementID, err = s.batchCreateAgent(ctx, elem)
		case "ensemble":
			elementID, err = s.batchCreateEnsemble(ctx, elem)
		default:
			err = fmt.Errorf("unsupported element type: %s", elem.Type)
		}

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			failed++
		} else {
			result.Success = true
			result.ID = elementID
			result.Data = map[string]interface{}{
				"file_path": fmt.Sprintf("data/elements/%s/%s/%s.yaml",
					elem.Type,
					time.Now().Format("2006-01-02"),
					elementID),
			}
			created++
		}

		results = append(results, result)
	}

	duration := time.Since(startTime)

	output := BatchCreateElementsOutput{
		Created:    created,
		Failed:     failed,
		Total:      len(input.Elements),
		Results:    results,
		Summary:    fmt.Sprintf("Batch complete: %d created, %d failed out of %d total", created, failed, len(input.Elements)),
		DurationMs: duration.Milliseconds(),
	}

	return nil, output, nil
}

// Helper functions for batch creation

func (s *MCPServer) batchCreatePersona(ctx context.Context, input BatchElementInput) (string, error) {
	// Use quick create if template specified
	if input.Template != "" {
		quickInput := QuickCreatePersonaInput{
			Name:        input.Name,
			Description: input.Description,
			Template:    input.Template,
		}

		// Extract expertise from data if provided
		if expertise, ok := input.Data["expertise"].([]interface{}); ok {
			for _, exp := range expertise {
				if expStr, ok := exp.(string); ok {
					quickInput.Expertise = append(quickInput.Expertise, expStr)
				}
			}
		}

		_, output, err := s.handleQuickCreatePersona(ctx, nil, quickInput)
		if err != nil {
			return "", err
		}
		return output["id"].(string), nil
	}

	// Fallback to standard creation
	persona := domain.NewPersona(input.Name, input.Description, "1.0.0", getCurrentUserFromContext(ctx))

	if err := persona.Validate(); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repo.Create(persona); err != nil {
		return "", fmt.Errorf("failed to create: %w", err)
	}

	return persona.GetMetadata().ID, nil
}

func (s *MCPServer) batchCreateSkill(ctx context.Context, input BatchElementInput) (string, error) {
	// Use quick create if template specified
	if input.Template != "" {
		quickInput := QuickCreateSkillInput{
			Name:        input.Name,
			Description: input.Description,
			Template:    input.Template,
		}

		// Extract trigger from data if provided
		if trigger, ok := input.Data["trigger"].(string); ok {
			quickInput.Trigger = trigger
		}

		_, output, err := s.handleQuickCreateSkill(ctx, nil, quickInput)
		if err != nil {
			return "", err
		}
		return output["id"].(string), nil
	}

	// Fallback to standard creation
	skill := domain.NewSkill(input.Name, input.Description, "1.0.0", getCurrentUserFromContext(ctx))

	if err := skill.Validate(); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repo.Create(skill); err != nil {
		return "", fmt.Errorf("failed to create: %w", err)
	}

	return skill.GetMetadata().ID, nil
}

func (s *MCPServer) batchCreateMemory(ctx context.Context, input BatchElementInput) (string, error) {
	// Use quick create
	quickInput := QuickCreateMemoryInput{
		Name: input.Name,
	}

	// Extract content from data
	if content, ok := input.Data["content"].(string); ok {
		quickInput.Content = content
	} else {
		return "", fmt.Errorf("content required for memory")
	}

	// Extract tags from data if provided
	if tags, ok := input.Data["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				quickInput.Tags = append(quickInput.Tags, tagStr)
			}
		}
	}

	// Extract importance from data if provided
	if importance, ok := input.Data["importance"].(string); ok {
		quickInput.Importance = importance
	}

	_, output, err := s.handleQuickCreateMemory(ctx, nil, quickInput)
	if err != nil {
		return "", err
	}
	return output["id"].(string), nil
}

func (s *MCPServer) batchCreateTemplate(ctx context.Context, input BatchElementInput) (string, error) {
	template := domain.NewTemplate(input.Name, input.Description, "1.0.0", getCurrentUserFromContext(ctx))

	// Extract content from data
	if content, ok := input.Data["content"].(string); ok {
		template.Content = content
	} else {
		return "", fmt.Errorf("content required for template")
	}

	// Extract variables from data if provided
	if variables, ok := input.Data["variables"].(map[string]interface{}); ok {
		for k, v := range variables {
			if vStr, ok := v.(string); ok {
				template.Variables = append(template.Variables, domain.TemplateVariable{
					Name:        k,
					Description: vStr,
				})
			}
		}
	}

	if err := template.Validate(); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repo.Create(template); err != nil {
		return "", fmt.Errorf("failed to create: %w", err)
	}

	return template.GetMetadata().ID, nil
}

func (s *MCPServer) batchCreateAgent(ctx context.Context, input BatchElementInput) (string, error) {
	agent := domain.NewAgent(input.Name, input.Description, "1.0.0", getCurrentUserFromContext(ctx))

	// Extract goal from data
	if goal, ok := input.Data["goal"].(string); ok {
		agent.Goals = []string{goal}
	} else {
		return "", fmt.Errorf("goal required for agent")
	}

	if err := agent.Validate(); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repo.Create(agent); err != nil {
		return "", fmt.Errorf("failed to create: %w", err)
	}

	return agent.GetMetadata().ID, nil
}

func (s *MCPServer) batchCreateEnsemble(ctx context.Context, input BatchElementInput) (string, error) {
	ensemble := domain.NewEnsemble(input.Name, input.Description, "1.0.0", getCurrentUserFromContext(ctx))

	// Extract purpose from data and set execution mode
	if purpose, ok := input.Data["purpose"].(string); ok {
		ensemble.AggregationStrategy = purpose // Use aggregation strategy
	}

	// Default to sequential if no execution mode specified
	if mode, ok := input.Data["execution_mode"].(string); ok {
		ensemble.ExecutionMode = mode
	}

	if err := ensemble.Validate(); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repo.Create(ensemble); err != nil {
		return "", fmt.Errorf("failed to create: %w", err)
	}

	return ensemble.GetMetadata().ID, nil
}
