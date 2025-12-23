package application

import (
	"context"
	"fmt"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
	"github.com/fsvxavier/nexs-mcp/internal/vectorstore"
)

// ElementRepository defines the interface for element storage operations
type ElementRepository interface {
	GetByID(id string) (domain.Element, error)
	List(filter domain.ElementFilter) ([]domain.Element, error)
}

// SemanticSearchService provides semantic search capabilities across all elements
type SemanticSearchService struct {
	store      *vectorstore.Store
	provider   embeddings.Provider
	repository ElementRepository
}

// NewSemanticSearchService creates a new semantic search service
func NewSemanticSearchService(provider embeddings.Provider, repository ElementRepository) *SemanticSearchService {
	return &SemanticSearchService{
		store:      vectorstore.NewStore(provider),
		provider:   provider,
		repository: repository,
	}
}

// IndexElement adds an element to the semantic search index
func (s *SemanticSearchService) IndexElement(ctx context.Context, element domain.Element) error {
	// Create searchable text from element
	text := s.createSearchableText(element)

	metadata := map[string]interface{}{
		"type": string(element.GetType()),
		"id":   element.GetID(),
		"name": element.GetMetadata().Name,
	}

	return s.store.Add(ctx, element.GetID(), text, metadata)
}

// IndexAllElements indexes all elements from the repository
func (s *SemanticSearchService) IndexAllElements(ctx context.Context) error {
	types := []domain.ElementType{
		domain.PersonaElement,
		domain.SkillElement,
		domain.AgentElement,
		domain.MemoryElement,
		domain.TemplateElement,
		domain.EnsembleElement,
	}

	for _, elemType := range types {
		filter := domain.ElementFilter{Type: &elemType}
		elements, err := s.repository.List(filter)
		if err != nil {
			return fmt.Errorf("failed to list %s elements: %w", elemType, err)
		}

		for _, elem := range elements {
			if err := s.IndexElement(ctx, elem); err != nil {
				return fmt.Errorf("failed to index element %s: %w", elem.GetID(), err)
			}
		}
	}

	return nil
}

// Search performs semantic search across indexed elements
func (s *SemanticSearchService) Search(ctx context.Context, query string, limit int, elementType string) ([]embeddings.Result, error) {
	filters := make(map[string]interface{})
	if elementType != "" {
		filters["type"] = elementType
	}

	return s.store.Search(ctx, query, limit, filters)
}

// FindSimilarMemories finds memories semantically similar to a query
func (s *SemanticSearchService) FindSimilarMemories(ctx context.Context, query string, limit int) ([]*domain.Memory, error) {
	filters := map[string]interface{}{
		"type": string(domain.MemoryElement),
	}

	results, err := s.store.Search(ctx, query, limit, filters)
	if err != nil {
		return nil, err
	}

	// Fetch full memory objects
	var memories []*domain.Memory
	for _, result := range results {
		elem, err := s.repository.GetByID(result.ID)
		if err != nil {
			continue // Skip if not found
		}

		memory, ok := elem.(*domain.Memory)
		if ok {
			memories = append(memories, memory)
		}
	}

	return memories, nil
}

// FindSimilarElements finds elements of any type similar to a query
func (s *SemanticSearchService) FindSimilarElements(ctx context.Context, query string, elemType domain.ElementType, limit int) ([]domain.Element, error) {
	filters := make(map[string]interface{})
	if elemType != "" {
		filters["type"] = string(elemType)
	}

	results, err := s.store.Search(ctx, query, limit, filters)
	if err != nil {
		return nil, err
	}

	// Fetch full element objects
	var elements []domain.Element
	for _, result := range results {
		elem, err := s.repository.GetByID(result.ID)
		if err != nil {
			continue
		}

		elements = append(elements, elem)
	}

	return elements, nil
}

// GetIndexStats returns statistics about the search index
func (s *SemanticSearchService) GetIndexStats() map[string]interface{} {
	return map[string]interface{}{
		"total_vectors": s.store.Size(),
		"provider":      s.provider.Name(),
		"dimensions":    s.provider.Dimensions(),
		"provider_cost": s.provider.Cost(),
	}
}

// ClearIndex removes all vectors from the search index
func (s *SemanticSearchService) ClearIndex() {
	s.store.Clear()
}

// RemoveElement removes an element from the search index
func (s *SemanticSearchService) RemoveElement(id string) error {
	return s.store.Delete(id)
}

// createSearchableText creates a searchable text representation of an element
func (s *SemanticSearchService) createSearchableText(element domain.Element) string {
	meta := element.GetMetadata()
	text := fmt.Sprintf("%s: %s", meta.Name, meta.Description)

	// Add tags if present
	if len(meta.Tags) > 0 {
		text += fmt.Sprintf(" Tags: %v", meta.Tags)
	}

	// Add type-specific information
	text += fmt.Sprintf(" Type: %s", meta.Type)

	return text
}
