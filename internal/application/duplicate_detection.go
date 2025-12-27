package application

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
	"github.com/fsvxavier/nexs-mcp/internal/vectorstore"
)

// MemoryDuplicateGroup represents a group of duplicate or near-duplicate memories.
type MemoryDuplicateGroup struct {
	Representative *domain.Memory   `json:"representative"`
	Duplicates     []*domain.Memory `json:"duplicates"`
	Similarity     float32          `json:"similarity"`
	Count          int              `json:"count"`
}

// DuplicateDetectionConfig contains configuration for duplicate detection.
type DuplicateDetectionConfig struct {
	SimilarityThreshold float32 // Default: 0.95 (95% similar)
	MinContentLength    int     // Minimum content length to consider (skip short memories)
	MaxResults          int     // Maximum number of duplicate groups to return
}

// DuplicateDetectionService detects duplicate and near-duplicate memories using HNSW.
type DuplicateDetectionService struct {
	store      *vectorstore.Store
	provider   embeddings.Provider
	repository ElementRepository
	config     DuplicateDetectionConfig
}

// NewDuplicateDetectionService creates a new duplicate detection service.
func NewDuplicateDetectionService(provider embeddings.Provider, repository ElementRepository, config DuplicateDetectionConfig) *DuplicateDetectionService {
	if config.SimilarityThreshold == 0 {
		config.SimilarityThreshold = 0.95 // 95% similar by default
	}
	if config.MinContentLength == 0 {
		config.MinContentLength = 20 // Skip very short content
	}
	if config.MaxResults == 0 {
		config.MaxResults = 100 // Maximum 100 duplicate groups
	}

	return &DuplicateDetectionService{
		store:      vectorstore.NewStore(provider),
		provider:   provider,
		repository: repository,
		config:     config,
	}
}

// DetectDuplicates finds all duplicate memory groups.
func (d *DuplicateDetectionService) DetectDuplicates(ctx context.Context) ([]MemoryDuplicateGroup, error) {
	// Get all memories
	memories, err := d.getMemories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories: %w", err)
	}

	if len(memories) < 2 {
		return []DuplicateGroup{}, nil // No duplicates possible
	}

	// Index all memories in vector store
	for _, memory := range memories {
		if len(memory.Content) < d.config.MinContentLength {
			continue // Skip short content
		}
		if err := d.store.Add(ctx, memory.GetID(), memory.Content, map[string]interface{}{
			"id":           memory.GetID(),
			"name":         memory.GetMetadata().Name,
			"date_created": memory.DateCreated,
			"content_hash": memory.ContentHash,
		}); err != nil {
			return nil, fmt.Errorf("failed to index memory %s: %w", memory.GetID(), err)
		}
	}

	// Find duplicates using HNSW similarity search
	duplicateMap := make(map[string]*MemoryDuplicateGroup)
	processed := make(map[string]bool)

	for _, memory := range memories {
		if processed[memory.GetID()] {
			continue
		}

		if len(memory.Content) < d.config.MinContentLength {
			continue
		}

		// Search for similar memories
		results, err := d.store.Search(ctx, memory.Content, 20, nil) // Top 20 similar
		if err != nil {
			return nil, fmt.Errorf("failed to search similar memories for %s: %w", memory.GetID(), err)
		}

		// Filter by similarity threshold
		var duplicates []*domain.Memory
		for _, result := range results {
			if result.ID == memory.GetID() {
				continue // Skip self
			}
			if float32(result.Score) >= d.config.SimilarityThreshold {
				// Find the duplicate memory
				for _, m := range memories {
					if m.GetID() == result.ID {
						duplicates = append(duplicates, m)
						processed[result.ID] = true
						break
					}
				}
			}
		}

		// Create duplicate group if we found duplicates
		if len(duplicates) > 0 {
			group := &MemoryDuplicateGroup{
				Representative: memory,
				Duplicates:     duplicates,
				Similarity:     d.config.SimilarityThreshold,
				Count:          len(duplicates) + 1,
			}
			duplicateMap[memory.GetID()] = group
			processed[memory.GetID()] = true
		}
	}

	// Convert map to slice
	groups := make([]MemoryDuplicateGroup, 0, len(duplicateMap))
	for _, group := range duplicateMap {
		groups = append(groups, *group)
		if len(groups) >= d.config.MaxResults {
			break
		}
	}

	return groups, nil
}

// FindDuplicatesForMemory finds duplicates for a specific memory.
func (d *DuplicateDetectionService) FindDuplicatesForMemory(ctx context.Context, memoryID string) ([]domain.Memory, error) {
	// Get the target memory
	element, err := d.repository.GetByID(memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	memory, ok := element.(*domain.Memory)
	if !ok {
		return nil, fmt.Errorf("element %s is not a memory", memoryID)
	}

	// Index all memories
	memories, err := d.getMemories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories: %w", err)
	}

	for _, m := range memories {
		if len(m.Content) < d.config.MinContentLength {
			continue
		}
		if err := d.store.Add(ctx, m.GetID(), m.Content, map[string]interface{}{
			"id": m.GetID(),
		}); err != nil {
			return nil, fmt.Errorf("failed to index memory: %w", err)
		}
	}

	// Search for similar memories
	results, err := d.store.Search(ctx, memory.Content, 10, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar memories: %w", err)
	}

	// Filter by threshold
	var duplicates []domain.Memory
	for _, result := range results {
		if result.ID == memoryID {
			continue // Skip self
		}
		if result.Score >= d.config.SimilarityThreshold {
			element, err := d.repository.GetByID(result.ID)
			if err != nil {
				continue
			}
			if mem, ok := element.(*domain.Memory); ok {
				duplicates = append(duplicates, *mem)
			}
		}
	}

	return duplicates, nil
}

// MergeDuplicates merges duplicate memories into a single memory.
func (d *DuplicateDetectionService) MergeDuplicates(ctx context.Context, representativeID string, duplicateIDs []string) (*domain.Memory, error) {
	// Get representative memory
	element, err := d.repository.GetByID(representativeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get representative memory: %w", err)
	}

	representative, ok := element.(*domain.Memory)
	if !ok {
		return nil, fmt.Errorf("representative %s is not a memory", representativeID)
	}

	// Create merged memory
	merged := domain.NewMemory(
		representative.GetMetadata().Name,
		fmt.Sprintf("Merged from %d memories", len(duplicateIDs)+1),
		representative.GetMetadata().Version,
		representative.GetMetadata().Author,
	)

	// Combine content with timestamps
	mergedContent := representative.Content + "\n\n--- Merged Content ---\n"
	metadata := make(map[string]string)
	metadata["merged_from"] = representativeID
	metadata["merged_count"] = fmt.Sprintf("%d", len(duplicateIDs))
	metadata["merged_at"] = time.Now().Format(time.RFC3339)

	for i, duplicateID := range duplicateIDs {
		element, err := d.repository.GetByID(duplicateID)
		if err != nil {
			continue
		}
		if dup, ok := element.(*domain.Memory); ok {
			mergedContent += fmt.Sprintf("\n\n--- Source %d (ID: %s, Created: %s) ---\n", i+1, dup.GetID(), dup.DateCreated)
			mergedContent += dup.Content
			metadata[fmt.Sprintf("source_%d_id", i+1)] = dup.GetID()
		}
	}

	merged.Content = mergedContent
	merged.Metadata = metadata
	merged.ComputeHash()

	return merged, nil
}

// ComputeSimilarity computes cosine similarity between two memories.
func (d *DuplicateDetectionService) ComputeSimilarity(ctx context.Context, memoryID1, memoryID2 string) (float32, error) {
	// Get embeddings for both memories
	element1, err := d.repository.GetByID(memoryID1)
	if err != nil {
		return 0, fmt.Errorf("failed to get memory 1: %w", err)
	}
	memory1, ok := element1.(*domain.Memory)
	if !ok {
		return 0, fmt.Errorf("element %s is not a memory", memoryID1)
	}

	element2, err := d.repository.GetByID(memoryID2)
	if err != nil {
		return 0, fmt.Errorf("failed to get memory 2: %w", err)
	}
	memory2, ok := element2.(*domain.Memory)
	if !ok {
		return 0, fmt.Errorf("element %s is not a memory", memoryID2)
	}

	// Generate embeddings
	embedding1, err := d.provider.Embed(ctx, memory1.Content)
	if err != nil {
		return 0, fmt.Errorf("failed to embed memory 1: %w", err)
	}

	embedding2, err := d.provider.Embed(ctx, memory2.Content)
	if err != nil {
		return 0, fmt.Errorf("failed to embed memory 2: %w", err)
	}

	// Compute cosine similarity
	if len(embedding1) != len(embedding2) {
		return 0, fmt.Errorf("embedding dimensions mismatch")
	}

	var dotProduct, norm1, norm2 float64
	for i := 0; i < len(embedding1); i++ {
		dotProduct += float64(embedding1[i]) * float64(embedding2[i])
		norm1 += float64(embedding1[i]) * float64(embedding1[i])
		norm2 += float64(embedding2[i]) * float64(embedding2[i])
	}

	if norm1 == 0 || norm2 == 0 {
		return 0, nil
	}

	similarity := dotProduct / (sqrt(norm1) * sqrt(norm2))
	return float32(similarity), nil
}

// getMemories retrieves all memories from the repository.
func (d *DuplicateDetectionService) getMemories(ctx context.Context) ([]*domain.Memory, error) {
	elements, err := d.repository.List(domain.ElementFilter{
		Type: domain.MemoryElement,
	})
	if err != nil {
		return nil, err
	}

	memories := make([]*domain.Memory, 0, len(elements))
	for _, element := range elements {
		if memory, ok := element.(*domain.Memory); ok {
			memories = append(memories, memory)
		}
	}

	return memories, nil
}

// sqrt computes square root (simple implementation).
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
