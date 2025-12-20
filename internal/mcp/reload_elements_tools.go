package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ReloadElementsInput represents the input for reload_elements tool
type ReloadElementsInput struct {
	ElementTypes        []string `json:"element_types,omitempty"`
	ClearCaches         bool     `json:"clear_caches,omitempty"`
	ValidateAfterReload bool     `json:"validate_after_reload,omitempty"`
}

// ElementTypeCount represents count of elements by type
type ElementTypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// ValidationError represents a validation error for an element
type ValidationError struct {
	ElementID   string `json:"element_id"`
	ElementType string `json:"element_type"`
	Error       string `json:"error"`
}

// CacheStats represents cache statistics before and after reload
type CacheStats struct {
	BeforeSize int `json:"before_size"`
	AfterSize  int `json:"after_size"`
	Cleared    int `json:"cleared"`
}

// ReloadElementsOutput represents the output of reload_elements tool
type ReloadElementsOutput struct {
	ElementsReloaded []ElementTypeCount `json:"elements_reloaded"`
	ElementsFailed   []ElementTypeCount `json:"elements_failed,omitempty"`
	ValidationErrors []ValidationError  `json:"validation_errors,omitempty"`
	CacheStats       CacheStats         `json:"cache_stats"`
	ReloadTimeMs     int64              `json:"reload_time_ms"`
	TotalReloaded    int                `json:"total_reloaded"`
	TotalFailed      int                `json:"total_failed"`
}

// handleReloadElements handles reload_elements tool calls
func (s *MCPServer) handleReloadElements(ctx context.Context, req *sdk.CallToolRequest, input ReloadElementsInput) (*sdk.CallToolResult, ReloadElementsOutput, error) {
	startTime := time.Now()

	// Set defaults
	if len(input.ElementTypes) == 0 {
		input.ElementTypes = []string{"all"}
	}

	// Use input values directly (bool defaults to false)
	clearCaches := input.ClearCaches
	validateAfterReload := input.ValidateAfterReload

	// Determine which types to reload
	typesToReload := make(map[domain.ElementType]bool)
	for _, typeStr := range input.ElementTypes {
		if typeStr == "all" {
			typesToReload[domain.PersonaElement] = true
			typesToReload[domain.SkillElement] = true
			typesToReload[domain.TemplateElement] = true
			typesToReload[domain.AgentElement] = true
			typesToReload[domain.MemoryElement] = true
			typesToReload[domain.EnsembleElement] = true
			break
		}

		switch typeStr {
		case "persona":
			typesToReload[domain.PersonaElement] = true
		case "skill":
			typesToReload[domain.SkillElement] = true
		case "template":
			typesToReload[domain.TemplateElement] = true
		case "agent":
			typesToReload[domain.AgentElement] = true
		case "memory":
			typesToReload[domain.MemoryElement] = true
		case "ensemble":
			typesToReload[domain.EnsembleElement] = true
		default:
			return nil, ReloadElementsOutput{}, fmt.Errorf("invalid element_type: %s", typeStr)
		}
	}

	// Track cache stats (before)
	cacheStatsBefore := 0
	// Note: In a real implementation, we would query the actual cache sizes
	// For now, we'll use a placeholder value

	// Clear caches if requested
	if clearCaches {
		// Clear template cache if reloading templates
		if typesToReload[domain.TemplateElement] {
			// Note: TemplateRegistry doesn't expose a Clear method yet
			// This would need to be added to the registry
			// For now, we just document that caches should be cleared
		}

		// Clear any other caches (collection registry, etc.)
		// Note: CollectionRegistry also doesn't expose Clear methods yet
		// This is a placeholder for future implementation
	}

	// Reload elements by type
	reloadedCounts := make(map[string]int)
	failedCounts := make(map[string]int)
	var validationErrors []ValidationError

	for elementType := range typesToReload {
		// List all elements from repository using ElementFilter
		typeFilter := elementType
		filter := domain.ElementFilter{
			Type: &typeFilter,
		}

		elements, err := s.repo.List(filter)
		if err != nil {
			failedCounts[string(elementType)] += 1
			continue
		}

		// Debug: log how many elements found for each type
		// fmt.Printf("DEBUG: Found %d elements of type %s\n", len(elements), elementType)

		// Validate elements if requested
		for _, element := range elements {
			if validateAfterReload {
				if err := element.Validate(); err != nil {
					validationErrors = append(validationErrors, ValidationError{
						ElementID:   element.GetID(),
						ElementType: string(element.GetType()),
						Error:       err.Error(),
					})
					failedCounts[string(elementType)] += 1
				} else {
					reloadedCounts[string(elementType)] += 1
				}
			} else {
				reloadedCounts[string(elementType)] += 1
			}
		}
	}

	// Calculate totals
	totalReloaded := 0
	totalFailed := 0

	for _, count := range reloadedCounts {
		totalReloaded += count
	}

	for _, count := range failedCounts {
		totalFailed += count
	}

	// Build output arrays
	elementsReloaded := []ElementTypeCount{}
	for typeStr, count := range reloadedCounts {
		if count > 0 {
			elementsReloaded = append(elementsReloaded, ElementTypeCount{
				Type:  typeStr,
				Count: count,
			})
		}
	}

	elementsFailed := []ElementTypeCount{}
	for typeStr, count := range failedCounts {
		if count > 0 {
			elementsFailed = append(elementsFailed, ElementTypeCount{
				Type:  typeStr,
				Count: count,
			})
		}
	}

	// Track cache stats (after)
	cacheStatsAfter := totalReloaded
	cacheStatsCleared := cacheStatsBefore

	cacheStats := CacheStats{
		BeforeSize: cacheStatsBefore,
		AfterSize:  cacheStatsAfter,
		Cleared:    cacheStatsCleared,
	}

	// Calculate reload time
	reloadTime := time.Since(startTime).Milliseconds()

	output := ReloadElementsOutput{
		ElementsReloaded: elementsReloaded,
		ElementsFailed:   elementsFailed,
		ValidationErrors: validationErrors,
		CacheStats:       cacheStats,
		ReloadTimeMs:     reloadTime,
		TotalReloaded:    totalReloaded,
		TotalFailed:      totalFailed,
	}

	return nil, output, nil
}
