package application

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// RecommendationEngine provides intelligent element recommendations based on relationships.
type RecommendationEngine struct {
	repo  domain.ElementRepository
	index *RelationshipIndex
	mu    sync.RWMutex
}

// Recommendation represents a recommended element with score.
type Recommendation struct {
	ElementID   string
	ElementType domain.ElementType
	ElementName string
	Score       float64
	Reasons     []string // Why this was recommended
}

// RecommendationOptions configures recommendation behavior.
type RecommendationOptions struct {
	ElementType    *domain.ElementType // Filter by type
	ExcludeIDs     []string            // IDs to exclude
	MinScore       float64             // Minimum score threshold (0-1)
	MaxResults     int                 // Maximum recommendations (default: 10)
	IncludeReasons bool                // Include explanation of why recommended
}

// NewRecommendationEngine creates a new recommendation engine.
func NewRecommendationEngine(repo domain.ElementRepository, index *RelationshipIndex) *RecommendationEngine {
	return &RecommendationEngine{
		repo:  repo,
		index: index,
	}
}

// RecommendForElement suggests related elements based on existing relationships and patterns.
func (e *RecommendationEngine) RecommendForElement(
	ctx context.Context,
	elementID string,
	options RecommendationOptions,
) ([]*Recommendation, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Set defaults
	if options.MaxResults == 0 {
		options.MaxResults = 10
	}
	if options.MinScore == 0 {
		options.MinScore = 0.1
	}

	// Get source element
	sourceElem, err := e.repo.GetByID(elementID)
	if err != nil {
		return nil, err
	}

	// Calculate scores for all candidates
	candidates := make(map[string]*Recommendation)

	// 1. Direct relationships (highest score)
	if err := e.addDirectRelationships(ctx, sourceElem, candidates); err != nil {
		return nil, err
	}

	// 2. Co-occurrence patterns (related memories in common)
	if err := e.addCoOccurrenceRecommendations(ctx, elementID, candidates, options); err != nil {
		return nil, err
	}

	// 3. Tag similarity
	if err := e.addTagSimilarityRecommendations(ctx, sourceElem, candidates, options); err != nil {
		return nil, err
	}

	// 4. Type-based recommendations
	if err := e.addTypeBasedRecommendations(ctx, sourceElem, candidates, options); err != nil {
		return nil, err
	}

	// Filter and sort
	results := e.filterAndSort(candidates, options)

	return results, nil
}

// addDirectRelationships adds elements directly related to source (score: 1.0).
func (e *RecommendationEngine) addDirectRelationships(
	ctx context.Context,
	sourceElem domain.Element,
	candidates map[string]*Recommendation,
) error {
	var relatedIDs []string

	// Extract related IDs based on element type
	switch elem := sourceElem.(type) {
	case *domain.Persona:
		relatedIDs = append(relatedIDs, elem.RelatedSkills...)
		relatedIDs = append(relatedIDs, elem.RelatedTemplates...)
	case *domain.Agent:
		if elem.PersonaID != "" {
			relatedIDs = append(relatedIDs, elem.PersonaID)
		}
		relatedIDs = append(relatedIDs, elem.RelatedSkills...)
		relatedIDs = append(relatedIDs, elem.RelatedTemplates...)
	case *domain.Template:
		relatedIDs = append(relatedIDs, elem.RelatedSkills...)
	case *domain.Memory:
		// Parse related_to metadata
		if relatedTo, ok := elem.Metadata["related_to"]; ok {
			relatedIDs = parseRelatedIDsFromString(relatedTo)
		}
	}

	// Add each related element as high-score candidate
	for _, id := range relatedIDs {
		if id == "" {
			continue
		}

		elem, err := e.repo.GetByID(id)
		if err != nil {
			continue // Skip missing elements
		}

		meta := elem.GetMetadata()
		if _, exists := candidates[id]; !exists {
			candidates[id] = &Recommendation{
				ElementID:   id,
				ElementType: meta.Type,
				ElementName: meta.Name,
				Score:       1.0,
				Reasons:     []string{"directly related"},
			}
		}
	}

	return nil
}

// addCoOccurrenceRecommendations finds elements that frequently appear together.
func (e *RecommendationEngine) addCoOccurrenceRecommendations(
	ctx context.Context,
	elementID string,
	candidates map[string]*Recommendation,
	options RecommendationOptions,
) error {
	// Get memories related to this element
	memories, err := GetMemoriesRelatedTo(ctx, elementID, e.repo, e.index)
	if err != nil || len(memories) == 0 {
		return nil // No error, just no co-occurrences
	}

	// Count co-occurring elements
	coOccurrence := make(map[string]int)
	totalMemories := len(memories)

	for _, memory := range memories {
		if relatedTo, ok := memory.Metadata["related_to"]; ok {
			relatedIDs := parseRelatedIDsFromString(relatedTo)
			for _, id := range relatedIDs {
				if id != elementID && id != "" {
					coOccurrence[id]++
				}
			}
		}
	}

	// Convert counts to scores (frequency-based)
	for id, count := range coOccurrence {
		if count < 2 {
			continue // Require at least 2 co-occurrences
		}

		elem, err := e.repo.GetByID(id)
		if err != nil {
			continue
		}

		meta := elem.GetMetadata()
		score := float64(count) / float64(totalMemories) * 0.8 // Max 0.8 score

		if existing, exists := candidates[id]; exists {
			existing.Score += score
			existing.Reasons = append(existing.Reasons, "frequently co-occurs")
		} else {
			candidates[id] = &Recommendation{
				ElementID:   id,
				ElementType: meta.Type,
				ElementName: meta.Name,
				Score:       score,
				Reasons:     []string{"frequently co-occurs"},
			}
		}
	}

	return nil
}

// addTagSimilarityRecommendations finds elements with similar tags.
func (e *RecommendationEngine) addTagSimilarityRecommendations(
	ctx context.Context,
	sourceElem domain.Element,
	candidates map[string]*Recommendation,
	options RecommendationOptions,
) error {
	sourceMeta := sourceElem.GetMetadata()
	if len(sourceMeta.Tags) == 0 {
		return nil
	}

	// Get all elements of the same or compatible type
	filter := domain.ElementFilter{}
	if options.ElementType != nil {
		filter.Type = options.ElementType
	}

	allElements, err := e.repo.List(filter)
	if err != nil {
		return err
	}

	// Calculate Jaccard similarity for tags
	for _, elem := range allElements {
		meta := elem.GetMetadata()
		if meta.ID == sourceMeta.ID {
			continue // Skip self
		}

		similarity := calculateTagSimilarity(sourceMeta.Tags, meta.Tags)
		if similarity < 0.3 {
			continue // Require at least 30% similarity
		}

		score := similarity * 0.6 // Max 0.6 score from tags

		if existing, exists := candidates[meta.ID]; exists {
			existing.Score += score
			existing.Reasons = append(existing.Reasons, "similar tags")
		} else {
			candidates[meta.ID] = &Recommendation{
				ElementID:   meta.ID,
				ElementType: meta.Type,
				ElementName: meta.Name,
				Score:       score,
				Reasons:     []string{"similar tags"},
			}
		}
	}

	return nil
}

// addTypeBasedRecommendations adds recommendations based on element type patterns.
func (e *RecommendationEngine) addTypeBasedRecommendations(
	ctx context.Context,
	sourceElem domain.Element,
	candidates map[string]*Recommendation,
	options RecommendationOptions,
) error {
	sourceMeta := sourceElem.GetMetadata()

	// Type-specific patterns
	var targetType *domain.ElementType

	switch sourceMeta.Type {
	case domain.PersonaElement:
		// Personas often use Skills and Templates
		skillType := domain.SkillElement
		targetType = &skillType
	case domain.AgentElement:
		// Agents often use Personas
		personaType := domain.PersonaElement
		targetType = &personaType
	case domain.SkillElement:
		// Skills often used by Personas
		personaType := domain.PersonaElement
		targetType = &personaType
	default:
		return nil
	}

	// Get elements of target type
	filter := domain.ElementFilter{Type: targetType}
	elements, err := e.repo.List(filter)
	if err != nil {
		return err
	}

	// Add with low base score (can be boosted by other signals)
	for _, elem := range elements {
		meta := elem.GetMetadata()
		if _, exists := candidates[meta.ID]; !exists {
			candidates[meta.ID] = &Recommendation{
				ElementID:   meta.ID,
				ElementType: meta.Type,
				ElementName: meta.Name,
				Score:       0.2, // Low base score
				Reasons:     []string{"commonly related type"},
			}
		}
	}

	return nil
}

// filterAndSort filters by options and sorts by score.
func (e *RecommendationEngine) filterAndSort(
	candidates map[string]*Recommendation,
	options RecommendationOptions,
) []*Recommendation {
	results := make([]*Recommendation, 0, len(candidates))

	// Convert map to slice and filter
	for _, rec := range candidates {
		// Filter by type
		if options.ElementType != nil && rec.ElementType != *options.ElementType {
			continue
		}

		// Filter by excluded IDs
		if contains(options.ExcludeIDs, rec.ElementID) {
			continue
		}

		// Filter by minimum score
		if rec.Score < options.MinScore {
			continue
		}

		// Normalize score to 0-1 range (can exceed 1.0 due to multiple signals)
		if rec.Score > 1.0 {
			rec.Score = 1.0
		}

		// Remove duplicate reasons
		rec.Reasons = uniqueStrings(rec.Reasons)

		results = append(results, rec)
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if len(results) > options.MaxResults {
		results = results[:options.MaxResults]
	}

	return results
}

// calculateTagSimilarity calculates Jaccard similarity between two tag sets.
func calculateTagSimilarity(tags1, tags2 []string) float64 {
	if len(tags1) == 0 || len(tags2) == 0 {
		return 0.0
	}

	// Convert to sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, tag := range tags1 {
		set1[strings.ToLower(tag)] = true
	}
	for _, tag := range tags2 {
		set2[strings.ToLower(tag)] = true
	}

	// Calculate intersection and union
	intersection := 0
	for tag := range set1 {
		if set2[tag] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}
