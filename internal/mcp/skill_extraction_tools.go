package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// ExtractSkillsFromPersonaInput defines input for extract_skills_from_persona tool.
type ExtractSkillsFromPersonaInput struct {
	PersonaID string `json:"persona_id" jsonschema:"the persona ID to extract skills from"`
}

// ExtractSkillsFromPersonaOutput defines output for extract_skills_from_persona tool.
type ExtractSkillsFromPersonaOutput struct {
	SkillsCreated    int      `json:"skills_created"    jsonschema:"number of skills created"`
	SkillIDs         []string `json:"skill_ids"         jsonschema:"IDs of created/found skills"`
	PersonaUpdated   bool     `json:"persona_updated"   jsonschema:"whether persona was updated with skill references"`
	SkippedDuplicate int      `json:"skipped_duplicate" jsonschema:"number of duplicate skills skipped"`
	Errors           []string `json:"errors,omitempty"  jsonschema:"any errors encountered"`
	Message          string   `json:"message"           jsonschema:"summary message"`
}

// handleExtractSkillsFromPersona handles extract_skills_from_persona tool calls.
func (s *MCPServer) handleExtractSkillsFromPersona(ctx context.Context, req *sdk.CallToolRequest, input ExtractSkillsFromPersonaInput) (*sdk.CallToolResult, ExtractSkillsFromPersonaOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "extract_skills_from_persona",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	// Validate input
	if input.PersonaID == "" {
		handlerErr = errors.New("persona_id is required")
		return nil, ExtractSkillsFromPersonaOutput{}, handlerErr
	}

	// Create skill extractor
	extractor := application.NewSkillExtractor(s.repo)

	// Extract skills
	result, err := extractor.ExtractSkillsFromPersona(ctx, input.PersonaID)
	if err != nil {
		handlerErr = fmt.Errorf("failed to extract skills: %w", err)
		return nil, ExtractSkillsFromPersonaOutput{}, handlerErr
	}

	// Build message
	message := fmt.Sprintf("Extracted %d skills from persona. ", result.SkillsCreated)
	if result.SkippedDuplicate > 0 {
		message += fmt.Sprintf("Skipped %d duplicates. ", result.SkippedDuplicate)
	}
	if result.PersonaUpdated {
		message += "Persona updated with skill references."
	}
	if len(result.Errors) > 0 {
		message += fmt.Sprintf(" Encountered %d errors.", len(result.Errors))
	}

	output := ExtractSkillsFromPersonaOutput{
		SkillsCreated:    result.SkillsCreated,
		SkillIDs:         result.SkillIDs,
		PersonaUpdated:   result.PersonaUpdated,
		SkippedDuplicate: result.SkippedDuplicate,
		Errors:           result.Errors,
		Message:          message,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "extract_skills_from_persona", output)

	return nil, output, nil
}

// BatchExtractSkillsInput defines input for batch_extract_skills tool.
type BatchExtractSkillsInput struct {
	PersonaIDs []string `json:"persona_ids,omitempty" jsonschema:"list of persona IDs (if empty, extracts from all personas)"`
}

// BatchExtractSkillsOutput defines output for batch_extract_skills tool.
type BatchExtractSkillsOutput struct {
	TotalPersonasProcessed int                                       `json:"total_personas_processed" jsonschema:"number of personas processed"`
	TotalSkillsCreated     int                                       `json:"total_skills_created"     jsonschema:"total number of skills created"`
	TotalSkillsSkipped     int                                       `json:"total_skills_skipped"     jsonschema:"total number of duplicate skills skipped"`
	PersonasUpdated        int                                       `json:"personas_updated"         jsonschema:"number of personas updated"`
	Results                map[string]ExtractSkillsFromPersonaOutput `json:"results"                  jsonschema:"results per persona"`
	Message                string                                    `json:"message"                  jsonschema:"summary message"`
}

// handleBatchExtractSkills handles batch_extract_skills tool calls.
func (s *MCPServer) handleBatchExtractSkills(ctx context.Context, req *sdk.CallToolRequest, input BatchExtractSkillsInput) (*sdk.CallToolResult, BatchExtractSkillsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "batch_extract_skills",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	output := BatchExtractSkillsOutput{
		Results: make(map[string]ExtractSkillsFromPersonaOutput),
	}

	// Get persona IDs
	personaIDs := input.PersonaIDs
	if len(personaIDs) == 0 {
		// Get all personas
		personaType := domain.PersonaElement
		personas, err := s.repo.List(domain.ElementFilter{Type: &personaType})
		if err != nil {
			handlerErr = fmt.Errorf("failed to list personas: %w", err)
			return nil, output, handlerErr
		}
		for _, p := range personas {
			personaIDs = append(personaIDs, p.GetID())
		}
	}

	// Create extractor
	extractor := application.NewSkillExtractor(s.repo)

	// Process each persona
	for _, personaID := range personaIDs {
		result, err := extractor.ExtractSkillsFromPersona(ctx, personaID)
		if err != nil {
			output.Results[personaID] = ExtractSkillsFromPersonaOutput{
				Errors:  []string{err.Error()},
				Message: fmt.Sprintf("Failed: %v", err),
			}
			continue
		}

		output.TotalPersonasProcessed++
		output.TotalSkillsCreated += result.SkillsCreated
		output.TotalSkillsSkipped += result.SkippedDuplicate
		if result.PersonaUpdated {
			output.PersonasUpdated++
		}

		// Build message
		message := fmt.Sprintf("Created %d skills", result.SkillsCreated)
		if result.SkippedDuplicate > 0 {
			message += fmt.Sprintf(", skipped %d duplicates", result.SkippedDuplicate)
		}

		output.Results[personaID] = ExtractSkillsFromPersonaOutput{
			SkillsCreated:    result.SkillsCreated,
			SkillIDs:         result.SkillIDs,
			PersonaUpdated:   result.PersonaUpdated,
			SkippedDuplicate: result.SkippedDuplicate,
			Errors:           result.Errors,
			Message:          message,
		}
	}

	// Build summary message
	output.Message = fmt.Sprintf(
		"Processed %d personas. Created %d skills, skipped %d duplicates. Updated %d personas.",
		output.TotalPersonasProcessed,
		output.TotalSkillsCreated,
		output.TotalSkillsSkipped,
		output.PersonasUpdated,
	)

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "batch_extract_skills", output)

	return nil, output, nil
}
