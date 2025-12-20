package validation

import (
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// MemoryValidator validates Memory elements
type MemoryValidator struct{}

func NewMemoryValidator() *MemoryValidator {
	return &MemoryValidator{}
}

func (v *MemoryValidator) SupportedType() domain.ElementType {
	return domain.MemoryElement
}

func (v *MemoryValidator) Validate(element domain.Element, level ValidationLevel) (*ValidationResult, error) {
	startTime := time.Now()

	memory, ok := element.(*domain.Memory)
	if !ok {
		return nil, fmt.Errorf("element is not a Memory type")
	}

	result := &ValidationResult{
		IsValid:     true,
		Errors:      []ValidationIssue{},
		Warnings:    []ValidationIssue{},
		Infos:       []ValidationIssue{},
		ElementType: string(domain.MemoryElement),
		ElementID:   memory.GetID(),
	}

	v.validateBasic(memory, result)

	result.ValidationTime = time.Since(startTime).Milliseconds()
	return result, nil
}

func (v *MemoryValidator) validateBasic(memory *domain.Memory, result *ValidationResult) {
	// Validate content hash exists
	if memory.ContentHash == "" {
		result.AddInfo("content_hash", "Consider computing content hash for deduplication", "MEMORY_NO_HASH")
	}

	// Validate date format
	if memory.DateCreated == "" {
		result.AddWarning("date_created", "DateCreated should be set", "MEMORY_NO_DATE")
	}

	// Validate search index
	if len(memory.SearchIndex) == 0 {
		result.AddInfo("search_index", "Consider adding search index entries for better discoverability", "MEMORY_NO_INDEX")
	}
}
