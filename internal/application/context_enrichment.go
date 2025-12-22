package application

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// EnrichedContext representa um contexto de memória enriquecido com elementos relacionados.
type EnrichedContext struct {
	Memory           *domain.Memory
	RelatedElements  map[string]domain.Element
	RelationshipMap  domain.RelationshipMap
	TotalTokensSaved int
	FetchErrors      []error
	FetchDuration    time.Duration
}

// ExpandOptions configura o comportamento da expansão de contexto.
type ExpandOptions struct {
	// MaxDepth profundidade de expansão (0 = apenas diretos, -1 = ilimitado)
	MaxDepth int

	// IncludeTypes filtra tipos de elementos a incluir
	IncludeTypes []domain.ElementType

	// ExcludeTypes filtra tipos de elementos a excluir
	ExcludeTypes []domain.ElementType

	// IgnoreErrors continua expansão mesmo com erros
	IgnoreErrors bool

	// FetchStrategy estratégia de fetch: "parallel" ou "sequential"
	FetchStrategy string

	// MaxElements limite máximo de elementos relacionados (default: 20)
	MaxElements int

	// Timeout timeout para fetch de cada elemento
	Timeout time.Duration
}

// DefaultExpandOptions retorna opções padrão sensatas.
func DefaultExpandOptions() ExpandOptions {
	return ExpandOptions{
		MaxDepth:      0,
		IncludeTypes:  nil,
		ExcludeTypes:  nil,
		IgnoreErrors:  false,
		FetchStrategy: "parallel",
		MaxElements:   20,
		Timeout:       5 * time.Second,
	}
}

// ExpandMemoryContext enriquece uma Memory com seus elementos relacionados.
func ExpandMemoryContext(
	ctx context.Context,
	memory *domain.Memory,
	repo domain.ElementRepository,
	options ExpandOptions,
) (*EnrichedContext, error) {
	startTime := time.Now()

	enriched := &EnrichedContext{
		Memory:          memory,
		RelatedElements: make(map[string]domain.Element),
		RelationshipMap: make(domain.RelationshipMap),
		FetchErrors:     []error{},
	}

	// Parse related_to metadata
	relatedStr, ok := memory.Metadata["related_to"]
	if !ok || relatedStr == "" {
		enriched.FetchDuration = time.Since(startTime)
		return enriched, nil
	}

	// Parse IDs
	relatedIDs := parseRelatedIDs(relatedStr)
	if len(relatedIDs) == 0 {
		enriched.FetchDuration = time.Since(startTime)
		return enriched, nil
	}

	// Apply MaxElements limit
	if options.MaxElements > 0 && len(relatedIDs) > options.MaxElements {
		relatedIDs = relatedIDs[:options.MaxElements]
	}

	// Fetch elements
	var err error
	if options.FetchStrategy == "sequential" {
		err = fetchSequential(ctx, relatedIDs, repo, enriched, options)
	} else {
		err = fetchParallel(ctx, relatedIDs, repo, enriched, options)
	}

	enriched.FetchDuration = time.Since(startTime)

	// Calculate token savings
	enriched.TotalTokensSaved = calculateTokenSavings(enriched)

	if err != nil && !options.IgnoreErrors {
		return enriched, err
	}

	return enriched, nil
}

// parseRelatedIDs extrai e limpa IDs da string related_to.
func parseRelatedIDs(relatedStr string) []string {
	parts := strings.Split(relatedStr, ",")
	ids := make([]string, 0, len(parts))

	for _, id := range parts {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			ids = append(ids, trimmed)
		}
	}

	return ids
}

// fetchParallel busca elementos em paralelo usando goroutines.
func fetchParallel(
	ctx context.Context,
	relatedIDs []string,
	repo domain.ElementRepository,
	enriched *EnrichedContext,
	options ExpandOptions,
) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errorCount atomic.Int32
	errChan := make(chan error, len(relatedIDs))

	for _, id := range relatedIDs {
		wg.Add(1)
		go func(elemID string) {
			defer wg.Done()

			// Create timeout context for each fetch
			// Note: fetchCtx preparado para implementações futuras que suportem context
			var cancel context.CancelFunc
			if options.Timeout > 0 {
				_, cancel = context.WithTimeout(ctx, options.Timeout)
				defer cancel()
			}

			// Fetch element
			elem, err := repo.GetByID(elemID)
			if err != nil {
				errorCount.Add(1)
				relErr := domain.NewRelationshipError(elemID, domain.RelationshipRelatedTo, err)
				errChan <- relErr
				return
			}

			// Apply type filters
			if !shouldIncludeElement(elem, options) {
				return
			}

			// Add to enriched context
			mu.Lock()
			enriched.RelatedElements[elemID] = elem
			enriched.RelationshipMap.Add(elemID, domain.RelationshipRelatedTo)
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	for err := range errChan {
		enriched.FetchErrors = append(enriched.FetchErrors, err)
	}

	if len(enriched.FetchErrors) > 0 && !options.IgnoreErrors {
		return fmt.Errorf("failed to fetch %d elements: %w", len(enriched.FetchErrors), enriched.FetchErrors[0])
	}

	return nil
}

// fetchSequential busca elementos sequencialmente.
func fetchSequential(
	ctx context.Context,
	relatedIDs []string,
	repo domain.ElementRepository,
	enriched *EnrichedContext,
	options ExpandOptions,
) error {
	for _, elemID := range relatedIDs {
		// Create timeout context
		// Note: fetchCtx preparado para implementações futuras que suportem context
		var cancel context.CancelFunc
		if options.Timeout > 0 {
			_, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}

		// Fetch element
		elem, err := repo.GetByID(elemID)
		if err != nil {
			relErr := domain.NewRelationshipError(elemID, domain.RelationshipRelatedTo, err)
			enriched.FetchErrors = append(enriched.FetchErrors, relErr)

			if !options.IgnoreErrors {
				return relErr
			}
			continue
		}

		// Apply type filters
		if !shouldIncludeElement(elem, options) {
			continue
		}

		// Add to enriched context
		enriched.RelatedElements[elemID] = elem
		enriched.RelationshipMap.Add(elemID, domain.RelationshipRelatedTo)
	}

	return nil
}

// shouldIncludeElement verifica se elemento deve ser incluído com base nos filtros.
func shouldIncludeElement(elem domain.Element, options ExpandOptions) bool {
	elemType := elem.GetType()

	// Check exclude list
	for _, excludeType := range options.ExcludeTypes {
		if elemType == excludeType {
			return false
		}
	}

	// Check include list (if specified)
	if len(options.IncludeTypes) > 0 {
		found := false
		for _, includeType := range options.IncludeTypes {
			if elemType == includeType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// calculateTokenSavings estima economia de tokens vs chamadas individuais.
func calculateTokenSavings(ctx *EnrichedContext) int {
	if len(ctx.RelatedElements) == 0 {
		return 0
	}

	// Estimativa conservadora:
	// - Cada request individual: ~100 tokens overhead (headers, metadata, etc)
	// - Agregação em single response: ~25 tokens overhead total
	// - Economia: ~75% do overhead

	baseTokensPerRequest := 100
	totalIndividualTokens := len(ctx.RelatedElements) * baseTokensPerRequest
	aggregatedTokens := 25

	saved := totalIndividualTokens - aggregatedTokens

	// Adicionar economia de contexto compartilhado
	// Elementos relacionados compartilham contexto, economizando mais tokens
	contextSavingsPerElement := 50
	contextSavings := len(ctx.RelatedElements) * contextSavingsPerElement

	return saved + contextSavings
}

// GetElementByID helper para buscar elemento já carregado.
func (ec *EnrichedContext) GetElementByID(id string) (domain.Element, bool) {
	elem, ok := ec.RelatedElements[id]
	return elem, ok
}

// HasErrors retorna true se houver erros de fetch.
func (ec *EnrichedContext) HasErrors() bool {
	return len(ec.FetchErrors) > 0
}

// GetErrorCount retorna número de erros.
func (ec *EnrichedContext) GetErrorCount() int {
	return len(ec.FetchErrors)
}

// GetElementCount retorna número de elementos carregados.
func (ec *EnrichedContext) GetElementCount() int {
	return len(ec.RelatedElements)
}

// GetElementsByType retorna elementos filtrados por tipo.
func (ec *EnrichedContext) GetElementsByType(elemType domain.ElementType) []domain.Element {
	elements := make([]domain.Element, 0)

	for _, elem := range ec.RelatedElements {
		if elem.GetType() == elemType {
			elements = append(elements, elem)
		}
	}

	return elements
}
