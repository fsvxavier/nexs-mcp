package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// SkillExtractor extracts skills from persona elements and creates them as separate skill elements.
type SkillExtractor struct {
	repo domain.ElementRepository
}

// NewSkillExtractor creates a new skill extractor.
func NewSkillExtractor(repo domain.ElementRepository) *SkillExtractor {
	return &SkillExtractor{
		repo: repo,
	}
}

// ExtractionResult contains the results of skill extraction.
type ExtractionResult struct {
	SkillsCreated    int      `json:"skills_created"`
	SkillIDs         []string `json:"skill_ids"`
	PersonaUpdated   bool     `json:"persona_updated"`
	Errors           []string `json:"errors,omitempty"`
	SkippedDuplicate int      `json:"skipped_duplicate"`
}

// ExtractSkillsFromPersona extracts skills from a persona and creates them as separate elements.
func (e *SkillExtractor) ExtractSkillsFromPersona(ctx context.Context, personaID string) (*ExtractionResult, error) {
	result := &ExtractionResult{
		SkillIDs: []string{},
		Errors:   []string{},
	}

	// Get the persona
	elem, err := e.repo.GetByID(personaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get persona: %w", err)
	}

	persona, ok := elem.(*domain.Persona)
	if !ok {
		return nil, errors.New("element is not a persona")
	}

	// Get raw JSON data to access custom fields
	rawData, err := e.getPersonaRawData(personaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get persona raw data: %w", err)
	}

	// Extract skills from various sources
	skills := e.extractSkillsFromRawData(rawData, persona)

	// Create skill elements
	for idxSkill := range skills {
		// Check if skill already exists by name
		existing := e.findExistingSkill(skills[idxSkill].Name)
		if existing != nil {
			result.SkippedDuplicate++
			result.SkillIDs = append(result.SkillIDs, existing.GetID())
			continue
		}

		// Create new skill
		skill := domain.NewSkill(
			skills[idxSkill].Name,
			skills[idxSkill].Description,
			"1.0.0",
			persona.GetMetadata().Author,
		)

		// Add triggers
		for idxTrigger := range skills[idxSkill].Triggers {
			if err := skill.AddTrigger(skills[idxSkill].Triggers[idxTrigger]); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to add trigger to skill '%s': %v", skills[idxSkill].Name, err))
			}
		}

		// Add procedures
		for idxProcedure := range skills[idxSkill].Procedures {
			if err := skill.AddProcedure(skills[idxSkill].Procedures[idxProcedure]); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to add procedure to skill '%s': %v", skills[idxSkill].Name, err))
			}
		}

		// Set metadata tags
		metadata := skill.GetMetadata()
		metadata.Tags = skills[idxSkill].Tags
		skill.SetMetadata(metadata)

		// Validate skill
		if err := skill.Validate(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("skill '%s' validation failed: %v", skills[idxSkill].Name, err))
			continue
		}

		// Save skill
		if err := e.repo.Create(skill); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create skill '%s': %v", skills[idxSkill].Name, err))
			continue
		}

		result.SkillsCreated++
		result.SkillIDs = append(result.SkillIDs, skill.GetID())
	}

	// Update persona with related skills
	if len(result.SkillIDs) > 0 {
		for idxSkillID := range result.SkillIDs {
			persona.AddRelatedSkill(result.SkillIDs[idxSkillID])
		}

		if err := e.repo.Update(persona); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to update persona: %v", err))
		} else {
			result.PersonaUpdated = true
		}
	}

	return result, nil
}

// skillSpec represents a skill to be created.
type skillSpec struct {
	Name        string
	Description string
	Triggers    []domain.SkillTrigger
	Procedures  []domain.SkillProcedure
	Tags        []string
}

// extractSkillsFromRawData extracts skill specifications from persona raw data.
func (e *SkillExtractor) extractSkillsFromRawData(rawData map[string]interface{}, persona *domain.Persona) []skillSpec {
	var skills []skillSpec

	// Extract from expertise_areas (standard field)
	if len(persona.ExpertiseAreas) > 0 {
		for _, area := range persona.ExpertiseAreas {
			skills = append(skills, skillSpec{
				Name:        area.Domain,
				Description: fmt.Sprintf("%s expertise at %s level", area.Domain, area.Level),
				Triggers: []domain.SkillTrigger{
					{
						Type:     "keyword",
						Keywords: append([]string{strings.ToLower(area.Domain)}, area.Keywords...),
					},
				},
				Procedures: []domain.SkillProcedure{
					{
						Step:        1,
						Action:      fmt.Sprintf("Apply %s expertise", area.Domain),
						Description: area.Description,
					},
				},
				Tags: []string{"auto-extracted", "expertise", strings.ToLower(area.Level)},
			})
		}
	}

	// Extract from custom technical_skills field if present
	if techSkills, ok := rawData["technical_skills"].(map[string]interface{}); ok {
		// Extract from core_expertise
		if coreExpertise, ok := techSkills["core_expertise"].([]interface{}); ok {
			for _, skill := range coreExpertise {
				if skillName, ok := skill.(string); ok {
					skills = append(skills, e.createSkillFromName(skillName, "core-expertise"))
				}
			}
		}

		// Extract from architecture_patterns
		if archPatterns, ok := techSkills["architecture_patterns"].([]interface{}); ok {
			for _, pattern := range archPatterns {
				if patternName, ok := pattern.(string); ok {
					skills = append(skills, e.createSkillFromName(patternName, "architecture"))
				}
			}
		}

		// Extract from design_patterns
		if designPatterns, ok := techSkills["design_patterns"].([]interface{}); ok {
			for _, pattern := range designPatterns {
				if patternName, ok := pattern.(string); ok {
					skills = append(skills, e.createSkillFromName(patternName, "design-pattern"))
				}
			}
		}

		// Extract from go_expertise
		if goExpertise, ok := techSkills["go_expertise"].([]interface{}); ok {
			for _, skill := range goExpertise {
				if skillName, ok := skill.(string); ok {
					skills = append(skills, e.createSkillFromName(skillName, "golang"))
				}
			}
		}

		// Extract from security
		if security, ok := techSkills["security"].([]interface{}); ok {
			for _, skill := range security {
				if skillName, ok := skill.(string); ok {
					skills = append(skills, e.createSkillFromName(skillName, "security"))
				}
			}
		}
	}

	return skills
}

// createSkillFromName creates a skill specification from a name and category.
func (e *SkillExtractor) createSkillFromName(name, category string) skillSpec {
	keywords := e.generateKeywords(name)

	return skillSpec{
		Name:        name,
		Description: fmt.Sprintf("%s capability in %s", name, category),
		Triggers: []domain.SkillTrigger{
			{
				Type:     "keyword",
				Keywords: keywords,
			},
		},
		Procedures: []domain.SkillProcedure{
			{
				Step:        1,
				Action:      fmt.Sprintf("Apply %s knowledge and best practices", name),
				Description: fmt.Sprintf("Utilize %s expertise to solve problems", name),
			},
		},
		Tags: []string{"auto-extracted", category},
	}
}

// generateKeywords generates keywords from a skill name.
func (e *SkillExtractor) generateKeywords(name string) []string {
	keywords := []string{strings.ToLower(name)}

	// Add variations
	words := strings.Fields(name)
	if len(words) > 1 {
		for _, word := range words {
			if len(word) > 3 { // Skip short words
				keywords = append(keywords, strings.ToLower(word))
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	unique := []string{}
	for _, kw := range keywords {
		if !seen[kw] {
			seen[kw] = true
			unique = append(unique, kw)
		}
	}

	return unique
}

// findExistingSkill checks if a skill with the given name already exists.
func (e *SkillExtractor) findExistingSkill(name string) domain.Element {
	// List all skills
	skillType := domain.SkillElement
	filter := domain.ElementFilter{
		Type: &skillType,
	}

	elements, err := e.repo.List(filter)
	if err != nil {
		return nil
	}

	// Check for matching name
	nameLower := strings.ToLower(strings.TrimSpace(name))
	for _, elem := range elements {
		elemNameLower := strings.ToLower(strings.TrimSpace(elem.GetMetadata().Name))
		if elemNameLower == nameLower {
			return elem
		}
	}

	return nil
}

// getPersonaRawData retrieves the raw JSON data for a persona.
func (e *SkillExtractor) getPersonaRawData(personaID string) (map[string]interface{}, error) {
	// Get element from repository
	elem, err := e.repo.GetByID(personaID)
	if err != nil {
		return nil, err
	}

	// Convert to map using JSON marshaling
	data, err := json.Marshal(elem)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal persona: %w", err)
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal persona: %w", err)
	}

	return rawData, nil
}
