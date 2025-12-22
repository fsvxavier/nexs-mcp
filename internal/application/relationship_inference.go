package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/indexing"
)

// RelationshipInferenceEngine automatically infers relationships between elements.
type RelationshipInferenceEngine struct {
	repo       domain.ElementRepository
	index      *RelationshipIndex
	tfidfIndex *indexing.TFIDFIndex
}

// InferredRelationship represents a relationship discovered from content analysis.
type InferredRelationship struct {
	SourceID   string
	TargetID   string
	SourceType domain.ElementType
	TargetType domain.ElementType
	Confidence float64  // 0.0-1.0
	Evidence   []string // Reasons why this relationship was inferred
	InferredBy string   // Method used: "mention", "keyword", "semantic", "pattern"
}

// InferenceOptions controls the inference behavior.
type InferenceOptions struct {
	MinConfidence   float64              // Minimum confidence threshold (default: 0.5)
	Methods         []string             // Methods to use: "mention", "keyword", "semantic", "pattern"
	TargetTypes     []domain.ElementType // Limit inference to these types
	AutoApply       bool                 // Automatically add inferred relationships
	RequireEvidence int                  // Minimum evidence count (default: 1)
}

// NewRelationshipInferenceEngine creates a new inference engine.
func NewRelationshipInferenceEngine(
	repo domain.ElementRepository,
	index *RelationshipIndex,
	tfidfIndex *indexing.TFIDFIndex,
) *RelationshipInferenceEngine {
	return &RelationshipInferenceEngine{
		repo:       repo,
		index:      index,
		tfidfIndex: tfidfIndex,
	}
}

// InferRelationshipsForElement analyzes an element and infers relationships.
func (e *RelationshipInferenceEngine) InferRelationshipsForElement(
	ctx context.Context,
	elementID string,
	opts InferenceOptions,
) ([]*InferredRelationship, error) {
	// Set defaults
	if opts.MinConfidence == 0 {
		opts.MinConfidence = 0.5
	}
	if len(opts.Methods) == 0 {
		opts.Methods = []string{"mention", "keyword", "semantic"}
	}
	if opts.RequireEvidence == 0 {
		opts.RequireEvidence = 1
	}

	// Get source element
	sourceElem, err := e.repo.GetByID(elementID)
	if err != nil {
		return nil, fmt.Errorf("source element not found: %w", err)
	}

	var allInferences []*InferredRelationship

	// Apply each inference method
	for _, method := range opts.Methods {
		var inferences []*InferredRelationship

		switch method {
		case "mention":
			inferences, err = e.inferByMentions(ctx, sourceElem, opts)
		case "keyword":
			inferences, err = e.inferByKeywords(ctx, sourceElem, opts)
		case "semantic":
			inferences, err = e.inferBySemantic(ctx, sourceElem, opts)
		case "pattern":
			inferences, err = e.inferByPatterns(ctx, sourceElem, opts)
		default:
			continue
		}

		if err != nil {
			// Log error but continue with other methods
			continue
		}

		allInferences = append(allInferences, inferences...)
	}

	// Merge and aggregate inferences for same target
	merged := e.mergeInferences(allInferences)

	// Filter by confidence and evidence
	filtered := e.filterInferences(merged, opts)

	// Auto-apply if requested
	if opts.AutoApply {
		for _, inf := range filtered {
			e.applyInference(ctx, inf)
		}
	}

	return filtered, nil
}

// inferByMentions detects explicit mentions of element IDs or names in content.
func (e *RelationshipInferenceEngine) inferByMentions(
	ctx context.Context,
	sourceElem domain.Element,
	opts InferenceOptions,
) ([]*InferredRelationship, error) {
	var inferences []*InferredRelationship

	// Extract searchable content from element
	content := e.extractContentForAnalysis(sourceElem)
	if content == "" {
		return nil, nil
	}

	// Get all elements to check for mentions
	filter := domain.ElementFilter{}
	allElements, err := e.repo.List(filter)
	if err != nil {
		return nil, err
	}

	sourceID := sourceElem.GetID()
	contentLower := strings.ToLower(content)

	for _, targetElem := range allElements {
		targetID := targetElem.GetID()
		if targetID == sourceID {
			continue // Skip self
		}

		// Apply type filter
		if len(opts.TargetTypes) > 0 && !containsType(opts.TargetTypes, targetElem.GetType()) {
			continue
		}

		targetMeta := targetElem.GetMetadata()
		targetNameLower := strings.ToLower(targetMeta.Name)

		// Check for ID mention
		idMentioned := strings.Contains(content, targetID)

		// Check for name mention (whole word match)
		nameMentioned := false
		if len(targetNameLower) >= 3 {
			// Use word boundaries to avoid false positives
			pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(targetNameLower) + `\b`)
			nameMentioned = pattern.MatchString(contentLower)
		}

		if !idMentioned && !nameMentioned {
			continue
		}

		// Calculate confidence based on mention context
		confidence := 0.7
		evidence := []string{}

		if idMentioned {
			confidence = 0.9
			evidence = append(evidence, "ID explicitly mentioned")
		}
		if nameMentioned {
			confidence = max(confidence, 0.8)
			evidence = append(evidence, fmt.Sprintf("Name '%s' mentioned", targetMeta.Name))
		}

		inferences = append(inferences, &InferredRelationship{
			SourceID:   sourceID,
			TargetID:   targetID,
			SourceType: sourceElem.GetType(),
			TargetType: targetElem.GetType(),
			Confidence: confidence,
			Evidence:   evidence,
			InferredBy: "mention",
		})
	}

	return inferences, nil
}

// inferByKeywords detects relationships based on shared keywords/tags.
func (e *RelationshipInferenceEngine) inferByKeywords(
	ctx context.Context,
	sourceElem domain.Element,
	opts InferenceOptions,
) ([]*InferredRelationship, error) {
	var inferences []*InferredRelationship

	sourceMeta := sourceElem.GetMetadata()
	sourceTags := sourceMeta.Tags

	if len(sourceTags) == 0 {
		return nil, nil
	}

	// Get all elements
	filter := domain.ElementFilter{}
	allElements, err := e.repo.List(filter)
	if err != nil {
		return nil, err
	}

	sourceID := sourceElem.GetID()

	for _, targetElem := range allElements {
		targetID := targetElem.GetID()
		if targetID == sourceID {
			continue
		}

		// Apply type filter
		if len(opts.TargetTypes) > 0 && !containsType(opts.TargetTypes, targetElem.GetType()) {
			continue
		}

		targetMeta := targetElem.GetMetadata()
		targetTags := targetMeta.Tags

		if len(targetTags) == 0 {
			continue
		}

		// Calculate tag overlap
		sharedTags := intersectStrings(sourceTags, targetTags)
		if len(sharedTags) == 0 {
			continue
		}

		// Calculate Jaccard similarity
		similarity := float64(len(sharedTags)) / float64(len(unionStrings(sourceTags, targetTags)))

		if similarity < 0.3 {
			continue // Require at least 30% similarity
		}

		confidence := similarity * 0.8 // Max 0.8 confidence from keywords
		evidence := []string{
			fmt.Sprintf("%d shared tags: %s", len(sharedTags), strings.Join(sharedTags, ", ")),
		}

		inferences = append(inferences, &InferredRelationship{
			SourceID:   sourceID,
			TargetID:   targetID,
			SourceType: sourceElem.GetType(),
			TargetType: targetElem.GetType(),
			Confidence: confidence,
			Evidence:   evidence,
			InferredBy: "keyword",
		})
	}

	return inferences, nil
}

// inferBySemantic uses TF-IDF to find semantically similar elements.
func (e *RelationshipInferenceEngine) inferBySemantic(
	ctx context.Context,
	sourceElem domain.Element,
	opts InferenceOptions,
) ([]*InferredRelationship, error) {
	if e.tfidfIndex == nil {
		return nil, nil // TF-IDF not available
	}

	var inferences []*InferredRelationship

	// Extract content for semantic analysis
	content := e.extractContentForAnalysis(sourceElem)
	if content == "" {
		return nil, nil
	}

	// Find similar documents using TF-IDF
	results := e.tfidfIndex.FindSimilar(content, 20) // Top 20 similar

	sourceID := sourceElem.GetID()

	for _, result := range results {
		if result.DocumentID == sourceID {
			continue // Skip self
		}

		if result.Score < 0.3 {
			continue // Require at least 30% similarity
		}

		// Get target element
		targetElem, err := e.repo.GetByID(result.DocumentID)
		if err != nil {
			continue
		}

		// Apply type filter
		if len(opts.TargetTypes) > 0 && !containsType(opts.TargetTypes, targetElem.GetType()) {
			continue
		}

		confidence := result.Score * 0.9 // Max 0.9 confidence from semantic
		evidence := []string{
			fmt.Sprintf("semantic similarity: %.2f", result.Score),
		}

		inferences = append(inferences, &InferredRelationship{
			SourceID:   sourceID,
			TargetID:   result.DocumentID,
			SourceType: sourceElem.GetType(),
			TargetType: targetElem.GetType(),
			Confidence: confidence,
			Evidence:   evidence,
			InferredBy: "semantic",
		})
	}

	return inferences, nil
}

// inferByPatterns detects domain-specific relationship patterns.
func (e *RelationshipInferenceEngine) inferByPatterns(
	ctx context.Context,
	sourceElem domain.Element,
	opts InferenceOptions,
) ([]*InferredRelationship, error) {
	var inferences []*InferredRelationship

	// Pattern 1: Agent should reference a Persona
	if sourceElem.GetType() == domain.AgentElement {
		agent, ok := sourceElem.(*domain.Agent)
		if ok && agent.PersonaID == "" {
			// Suggest personas with similar goals
			personas, err := e.findSimilarPersonas(ctx, agent)
			if err == nil {
				for _, persona := range personas {
					inferences = append(inferences, &InferredRelationship{
						SourceID:   agent.GetID(),
						TargetID:   persona.GetID(),
						SourceType: domain.AgentElement,
						TargetType: domain.PersonaElement,
						Confidence: 0.6,
						Evidence:   []string{"agents typically reference a persona"},
						InferredBy: "pattern",
					})
				}
			}
		}
	}

	// Pattern 2: Templates and Skills with matching keywords
	if sourceElem.GetType() == domain.TemplateElement {
		template, ok := sourceElem.(*domain.Template)
		if ok {
			skills, err := e.findRelatedSkills(ctx, template)
			if err == nil {
				for _, skill := range skills {
					inferences = append(inferences, &InferredRelationship{
						SourceID:   template.GetID(),
						TargetID:   skill.GetID(),
						SourceType: domain.TemplateElement,
						TargetType: domain.SkillElement,
						Confidence: 0.5,
						Evidence:   []string{"template and skill have matching keywords"},
						InferredBy: "pattern",
					})
				}
			}
		}
	}

	return inferences, nil
}

// Helper functions

func (e *RelationshipInferenceEngine) extractContentForAnalysis(elem domain.Element) string {
	var parts []string

	meta := elem.GetMetadata()
	parts = append(parts, meta.Name, meta.Description)
	parts = append(parts, meta.Tags...)

	// Type-specific content extraction
	switch e := elem.(type) {
	case *domain.Persona:
		parts = append(parts, e.SystemPrompt)
		for _, trait := range e.BehavioralTraits {
			parts = append(parts, trait.Name, trait.Description)
		}
		for _, exp := range e.ExpertiseAreas {
			parts = append(parts, exp.Domain)
		}
	case *domain.Skill:
		// Extract trigger keywords
		for _, trigger := range e.Triggers {
			parts = append(parts, trigger.Keywords...)
			if trigger.Pattern != "" {
				parts = append(parts, trigger.Pattern)
			}
			if trigger.Context != "" {
				parts = append(parts, trigger.Context)
			}
		}
		// Extract procedure descriptions
		for _, proc := range e.Procedures {
			parts = append(parts, proc.Action, proc.Description)
		}
	case *domain.Agent:
		// Extract goals and action descriptions
		parts = append(parts, e.Goals...)
		for _, action := range e.Actions {
			parts = append(parts, action.Name)
		}
	case *domain.Template:
		parts = append(parts, e.Content)
	case *domain.Memory:
		parts = append(parts, e.Content)
	}

	return strings.Join(parts, " ")
}

func (e *RelationshipInferenceEngine) mergeInferences(inferences []*InferredRelationship) []*InferredRelationship {
	merged := make(map[string]*InferredRelationship)

	for _, inf := range inferences {
		key := inf.SourceID + ":" + inf.TargetID

		if existing, ok := merged[key]; ok {
			// Combine confidences (weighted average)
			existing.Confidence = (existing.Confidence + inf.Confidence) / 2
			existing.Evidence = append(existing.Evidence, inf.Evidence...)
			existing.InferredBy = existing.InferredBy + "," + inf.InferredBy
		} else {
			merged[key] = inf
		}
	}

	result := make([]*InferredRelationship, 0, len(merged))
	for _, inf := range merged {
		result = append(result, inf)
	}

	return result
}

func (e *RelationshipInferenceEngine) filterInferences(
	inferences []*InferredRelationship,
	opts InferenceOptions,
) []*InferredRelationship {
	filtered := make([]*InferredRelationship, 0)

	for _, inf := range inferences {
		if inf.Confidence < opts.MinConfidence {
			continue
		}
		if len(inf.Evidence) < opts.RequireEvidence {
			continue
		}
		filtered = append(filtered, inf)
	}

	return filtered
}

func (e *RelationshipInferenceEngine) applyInference(ctx context.Context, inf *InferredRelationship) error {
	// Add to relationship index
	e.index.Add(inf.SourceID, []string{inf.TargetID})
	return nil
}

func (e *RelationshipInferenceEngine) findSimilarPersonas(ctx context.Context, agent *domain.Agent) ([]*domain.Persona, error) {
	personaType := domain.PersonaElement
	filter := domain.ElementFilter{Type: &personaType}

	elements, err := e.repo.List(filter)
	if err != nil {
		return nil, err
	}

	personas := make([]*domain.Persona, 0)
	for _, elem := range elements {
		if persona, ok := elem.(*domain.Persona); ok {
			personas = append(personas, persona)
		}
	}

	return personas, nil
}

func (e *RelationshipInferenceEngine) findRelatedSkills(ctx context.Context, template *domain.Template) ([]*domain.Skill, error) {
	skillType := domain.SkillElement
	filter := domain.ElementFilter{Type: &skillType}

	elements, err := e.repo.List(filter)
	if err != nil {
		return nil, err
	}

	skills := make([]*domain.Skill, 0)
	for _, elem := range elements {
		if skill, ok := elem.(*domain.Skill); ok {
			skills = append(skills, skill)
		}
	}

	return skills, nil
}

// Utility functions

func intersectStrings(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}

	var result []string
	for _, item := range b {
		if m[item] {
			result = append(result, item)
		}
	}

	return result
}

func unionStrings(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}
	for _, item := range b {
		m[item] = true
	}

	result := make([]string, 0, len(m))
	for item := range m {
		result = append(result, item)
	}

	return result
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
