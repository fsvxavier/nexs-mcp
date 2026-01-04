package mcp

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/common"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// FindRelatedMemoriesInput defines input for find_related_memories tool.
type FindRelatedMemoriesInput struct {
	ElementID   string   `json:"element_id"             jsonschema:"required"                                                                   jsonschema_description:"Element ID to find related memories for"`
	IncludeTags []string `json:"include_tags,omitempty" jsonschema_description:"Filter by tags (AND logic)"`
	ExcludeTags []string `json:"exclude_tags,omitempty" jsonschema_description:"Exclude memories with these tags"`
	Author      string   `json:"author,omitempty"       jsonschema_description:"Filter by author"`
	FromDate    string   `json:"from_date,omitempty"    jsonschema_description:"Filter from date (YYYY-MM-DD)"`
	ToDate      string   `json:"to_date,omitempty"      jsonschema_description:"Filter to date (YYYY-MM-DD)"`
	SortBy      string   `json:"sort_by,omitempty"      jsonschema_description:"Sort field: created_at, updated_at, name (default: updated_at)"`
	SortOrder   string   `json:"sort_order,omitempty"   jsonschema_description:"Sort order: asc, desc (default: desc)"`
	Limit       int      `json:"limit,omitempty"        jsonschema_description:"Maximum number of memories to return (default: 50)"`
}

// FindRelatedMemoriesOutput defines output for find_related_memories tool.
type FindRelatedMemoriesOutput struct {
	ElementID      string                   `json:"element_id"      jsonschema_description:"Element ID that was searched"`
	ElementType    string                   `json:"element_type"    jsonschema_description:"Type of the element"`
	ElementName    string                   `json:"element_name"    jsonschema_description:"Name of the element"`
	TotalMemories  int                      `json:"total_memories"  jsonschema_description:"Number of memories returned"`
	Memories       []map[string]interface{} `json:"memories"        jsonschema_description:"Array of related memories with metadata"`
	IndexStats     map[string]interface{}   `json:"index_stats"     jsonschema_description:"Relationship index statistics"`
	SearchDuration int64                    `json:"search_duration" jsonschema_description:"Time taken to search (milliseconds)"`
}

// handleFindRelatedMemories handles bidirectional search for memories.
func (s *MCPServer) handleFindRelatedMemories(
	ctx context.Context,
	req *sdk.CallToolRequest,
	input FindRelatedMemoriesInput,
) (*sdk.CallToolResult, FindRelatedMemoriesOutput, error) {
	startTime := time.Now()
	var err error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "find_related_memories",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   err == nil,
			ErrorMessage: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		})
	}()

	// Validate input
	if input.ElementID == "" {
		err = errors.New("element_id is required")
		return nil, FindRelatedMemoriesOutput{}, err
	}

	// Get element to verify it exists
	elem, err := s.repo.GetByID(input.ElementID)
	if err != nil {
		return nil, FindRelatedMemoriesOutput{}, fmt.Errorf("element not found: %w", err)
	}

	metadata := elem.GetMetadata()

	// Get related memories using index
	memories, err := application.GetMemoriesRelatedTo(ctx, input.ElementID, s.repo, s.relationshipIndex)
	if err != nil {
		return nil, FindRelatedMemoriesOutput{}, fmt.Errorf("failed to get related memories: %w", err)
	}

	// Apply filters
	filtered := applyMemoryFilters(memories, input)

	// Apply sorting
	sortMemories(filtered, input.SortBy, input.SortOrder)

	// Apply limit
	limit := input.Limit
	if limit == 0 {
		limit = 50 // Default limit
	}
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Convert memories to maps
	memoriesMaps := make([]map[string]interface{}, len(filtered))
	for i, memory := range filtered {
		memoriesMaps[i] = convertMemoryToMap(memory)
	}

	// Get index stats
	stats := s.relationshipIndex.Stats()
	indexStats := map[string]interface{}{
		"forward_entries": stats.ForwardEntries,
		"reverse_entries": stats.ReverseEntries,
		"cache_hits":      stats.CacheHits,
		"cache_misses":    stats.CacheMisses,
		"cache_size":      stats.CacheSize,
	}

	// Build output
	output := FindRelatedMemoriesOutput{
		ElementID:      input.ElementID,
		ElementType:    string(metadata.Type),
		ElementName:    metadata.Name,
		TotalMemories:  len(memoriesMaps),
		Memories:       memoriesMaps,
		IndexStats:     indexStats,
		SearchDuration: time.Since(startTime).Milliseconds(),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "find_related_memories", output)

	return nil, output, nil
}

// applyMemoryFilters applies filters to memory list.
func applyMemoryFilters(memories []*domain.Memory, input FindRelatedMemoriesInput) []*domain.Memory {
	if len(memories) == 0 {
		return memories
	}

	filtered := make([]*domain.Memory, 0, len(memories))

	for _, memory := range memories {
		metadata := memory.GetMetadata()

		// Filter by author
		if input.Author != "" && metadata.Author != input.Author {
			continue
		}

		// Filter by include tags (AND logic)
		if len(input.IncludeTags) > 0 {
			if !hasAllTags(metadata.Tags, input.IncludeTags) {
				continue
			}
		}

		// Filter by exclude tags
		if len(input.ExcludeTags) > 0 {
			if hasAnyTag(metadata.Tags, input.ExcludeTags) {
				continue
			}
		}

		// Filter by date range
		if input.FromDate != "" {
			fromDate, err := time.Parse("2006-01-02", input.FromDate)
			if err == nil && metadata.CreatedAt.Before(fromDate) {
				continue
			}
		}

		if input.ToDate != "" {
			toDate, err := time.Parse("2006-01-02", input.ToDate)
			if err == nil && metadata.CreatedAt.After(toDate.Add(24*time.Hour)) {
				continue
			}
		}

		filtered = append(filtered, memory)
	}

	return filtered
}

// sortMemories sorts memories by specified field and order.
func sortMemories(memories []*domain.Memory, sortBy, sortOrder string) {
	if len(memories) == 0 {
		return
	}

	// Default: sort by updated_at desc
	if sortBy == "" {
		sortBy = "updated_at"
	}
	if sortOrder == "" {
		sortOrder = common.SortOrderDesc
	}

	sort.Slice(memories, func(i, j int) bool {
		meta1 := memories[i].GetMetadata()
		meta2 := memories[j].GetMetadata()

		var less bool
		switch sortBy {
		case "created_at":
			less = meta1.CreatedAt.Before(meta2.CreatedAt)
		case "updated_at":
			less = meta1.UpdatedAt.Before(meta2.UpdatedAt)
		case "name":
			less = meta1.Name < meta2.Name
		default:
			less = meta1.UpdatedAt.Before(meta2.UpdatedAt)
		}

		if sortOrder == common.SortOrderDesc {
			return !less
		}
		return less
	})
}

// hasAllTags checks if slice contains all required tags.
func hasAllTags(tags []string, required []string) bool {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	for _, req := range required {
		if !tagSet[req] {
			return false
		}
	}

	return true
}

// hasAnyTag checks if slice contains any of the excluded tags.
func hasAnyTag(tags []string, excluded []string) bool {
	excludeSet := make(map[string]bool)
	for _, tag := range excluded {
		excludeSet[tag] = true
	}

	for _, tag := range tags {
		if excludeSet[tag] {
			return true
		}
	}

	return false
}
