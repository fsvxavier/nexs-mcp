package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// --- Memory Search Input/Output structures ---

// SearchMemoryInput defines input for search_memory tool
type SearchMemoryInput struct {
	Query      string `json:"query" jsonschema:"search query text"`
	Author     string `json:"author,omitempty" jsonschema:"filter by author"`
	DateFrom   string `json:"date_from,omitempty" jsonschema:"filter by date from (YYYY-MM-DD)"`
	DateTo     string `json:"date_to,omitempty" jsonschema:"filter by date to (YYYY-MM-DD)"`
	Limit      int    `json:"limit,omitempty" jsonschema:"maximum number of results (default: 10)"`
	IncludeAll bool   `json:"include_all,omitempty" jsonschema:"include inactive memories (default: false)"`
	User       string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// SearchMemoryOutput defines output for search_memory tool
type SearchMemoryOutput struct {
	Memories []MemorySummary `json:"memories" jsonschema:"list of matching memories"`
	Total    int             `json:"total" jsonschema:"total number of results"`
	Query    string          `json:"query" jsonschema:"the search query used"`
}

// MemorySummary represents a memory search result
type MemorySummary struct {
	ID          string `json:"id" jsonschema:"memory ID"`
	Name        string `json:"name" jsonschema:"memory name"`
	Content     string `json:"content" jsonschema:"memory content"`
	DateCreated string `json:"date_created" jsonschema:"creation date"`
	Author      string `json:"author" jsonschema:"author"`
	IsActive    bool   `json:"is_active" jsonschema:"active status"`
}

// --- Memory Summarization Input/Output structures ---

// SummarizeMemoriesInput defines input for summarize_memories tool
type SummarizeMemoriesInput struct {
	Author   string `json:"author,omitempty" jsonschema:"filter by author"`
	DateFrom string `json:"date_from,omitempty" jsonschema:"filter by date from (YYYY-MM-DD)"`
	DateTo   string `json:"date_to,omitempty" jsonschema:"filter by date to (YYYY-MM-DD)"`
	MaxItems int    `json:"max_items,omitempty" jsonschema:"maximum memories to include (default: 50)"`
	User     string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// SummarizeMemoriesOutput defines output for summarize_memories tool
type SummarizeMemoriesOutput struct {
	Summary      string           `json:"summary" jsonschema:"text summary of memories"`
	TotalCount   int              `json:"total_count" jsonschema:"total number of memories"`
	DateRange    string           `json:"date_range" jsonschema:"date range covered"`
	TopAuthors   []string         `json:"top_authors" jsonschema:"most frequent authors"`
	Statistics   MemoryStatistics `json:"statistics" jsonschema:"memory statistics"`
	RecentMemory *MemorySummary   `json:"recent_memory,omitempty" jsonschema:"most recent memory"`
}

// MemoryStatistics represents statistics about memories
type MemoryStatistics struct {
	TotalMemories  int     `json:"total_memories" jsonschema:"total number of memories"`
	ActiveMemories int     `json:"active_memories" jsonschema:"number of active memories"`
	TotalSize      int     `json:"total_size" jsonschema:"total content size in bytes"`
	AverageSize    float64 `json:"average_size" jsonschema:"average content size"`
}

// --- Memory Update Input/Output structures ---

// UpdateMemoryInput defines input for update_memory tool
type UpdateMemoryInput struct {
	ID          string            `json:"id" jsonschema:"memory ID to update"`
	Content     string            `json:"content,omitempty" jsonschema:"new content"`
	Name        string            `json:"name,omitempty" jsonschema:"new name"`
	Description string            `json:"description,omitempty" jsonschema:"new description"`
	Tags        []string          `json:"tags,omitempty" jsonschema:"new tags"`
	Metadata    map[string]string `json:"metadata,omitempty" jsonschema:"additional metadata"`
	User        string            `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// UpdateMemoryOutput defines output for update_memory tool
type UpdateMemoryOutput struct {
	Memory MemorySummary `json:"memory" jsonschema:"updated memory details"`
}

// --- Memory Delete Input/Output structures ---

// DeleteMemoryInput defines input for delete_memory tool
type DeleteMemoryInput struct {
	ID   string `json:"id" jsonschema:"memory ID to delete"`
	User string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// DeleteMemoryOutput defines output for delete_memory tool
type DeleteMemoryOutput struct {
	Success bool   `json:"success" jsonschema:"deletion success status"`
	Message string `json:"message" jsonschema:"deletion result message"`
	ID      string `json:"id" jsonschema:"deleted memory ID"`
}

// --- Clear Memories Input/Output structures ---

// ClearMemoriesInput defines input for clear_memories tool
type ClearMemoriesInput struct {
	Author     string `json:"author,omitempty" jsonschema:"clear only memories by this author"`
	DateBefore string `json:"date_before,omitempty" jsonschema:"clear memories before this date (YYYY-MM-DD)"`
	Confirm    bool   `json:"confirm" jsonschema:"confirmation flag (must be true to proceed)"`
	User       string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// ClearMemoriesOutput defines output for clear_memories tool
type ClearMemoriesOutput struct {
	Success      bool   `json:"success" jsonschema:"operation success status"`
	DeletedCount int    `json:"deleted_count" jsonschema:"number of memories deleted"`
	Message      string `json:"message" jsonschema:"operation result message"`
}

// --- Tool handlers ---

// handleSearchMemory handles the search_memory tool
func (s *MCPServer) handleSearchMemory(ctx context.Context, req *sdk.CallToolRequest, input SearchMemoryInput) (*sdk.CallToolResult, SearchMemoryOutput, error) {
	// Validate input
	if input.Query == "" {
		return nil, SearchMemoryOutput{}, fmt.Errorf("query is required")
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 10
	}

	// Build filter
	filter := domain.ElementFilter{
		Type: func() *domain.ElementType { t := domain.MemoryElement; return &t }(),
	}

	if !input.IncludeAll {
		active := true
		filter.IsActive = &active
	}

	// Get all memories
	elements, err := s.repo.List(filter)
	if err != nil {
		return nil, SearchMemoryOutput{}, fmt.Errorf("failed to list memories: %w", err)
	}

	// Filter and score memories
	var results []struct {
		memory *domain.Memory
		score  int
	}

	queryLower := strings.ToLower(input.Query)
	queryWords := strings.Fields(queryLower)

	for _, elem := range elements {
		memory, ok := elem.(*domain.Memory)
		if !ok {
			continue
		}

		// Filter by author
		if input.Author != "" && memory.GetMetadata().Author != input.Author {
			continue
		}

		// Filter by date range
		if input.DateFrom != "" && memory.DateCreated < input.DateFrom {
			continue
		}
		if input.DateTo != "" && memory.DateCreated > input.DateTo {
			continue
		}

		// Calculate relevance score
		score := 0
		contentLower := strings.ToLower(memory.Content)
		nameLower := strings.ToLower(memory.GetMetadata().Name)

		// Exact match in name = highest score
		if nameLower == queryLower {
			score += 100
		} else if strings.Contains(nameLower, queryLower) {
			score += 50
		}

		// Word matches
		for _, word := range queryWords {
			// Count occurrences in content
			contentCount := strings.Count(contentLower, word)
			score += contentCount * 5

			// Word in name is worth more
			nameCount := strings.Count(nameLower, word)
			score += nameCount * 25
		}

		// Search index matches
		for _, indexTerm := range memory.SearchIndex {
			if strings.Contains(strings.ToLower(indexTerm), queryLower) {
				score += 15
			}
		}

		if score > 0 {
			results = append(results, struct {
				memory *domain.Memory
				score  int
			}{memory, score})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		if results[i].score == results[j].score {
			// Secondary sort by date (newer first)
			return results[i].memory.DateCreated > results[j].memory.DateCreated
		}
		return results[i].score > results[j].score
	})

	// Apply limit
	if len(results) > input.Limit {
		results = results[:input.Limit]
	}

	// Convert to output
	memories := make([]MemorySummary, len(results))
	for i, r := range results {
		meta := r.memory.GetMetadata()
		memories[i] = MemorySummary{
			ID:          meta.ID,
			Name:        meta.Name,
			Content:     r.memory.Content,
			DateCreated: r.memory.DateCreated,
			Author:      meta.Author,
			IsActive:    meta.IsActive,
		}
	}

	output := SearchMemoryOutput{
		Memories: memories,
		Total:    len(memories),
		Query:    input.Query,
	}

	return nil, output, nil
}

// handleSummarizeMemories handles the summarize_memories tool
func (s *MCPServer) handleSummarizeMemories(ctx context.Context, req *sdk.CallToolRequest, input SummarizeMemoriesInput) (*sdk.CallToolResult, SummarizeMemoriesOutput, error) {
	// Set defaults
	if input.MaxItems <= 0 {
		input.MaxItems = 50
	}

	// Build filter
	filter := domain.ElementFilter{
		Type: func() *domain.ElementType { t := domain.MemoryElement; return &t }(),
	}
	active := true
	filter.IsActive = &active

	// Get all active memories
	elements, err := s.repo.List(filter)
	if err != nil {
		return nil, SummarizeMemoriesOutput{}, fmt.Errorf("failed to list memories: %w", err)
	}

	var memories []*domain.Memory
	for _, elem := range elements {
		if memory, ok := elem.(*domain.Memory); ok {
			// Filter by author
			if input.Author != "" && memory.GetMetadata().Author != input.Author {
				continue
			}

			// Filter by date range
			if input.DateFrom != "" && memory.DateCreated < input.DateFrom {
				continue
			}
			if input.DateTo != "" && memory.DateCreated > input.DateTo {
				continue
			}

			memories = append(memories, memory)
		}
	}

	// Sort by date (newest first)
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].DateCreated > memories[j].DateCreated
	})

	// Apply limit
	if len(memories) > input.MaxItems {
		memories = memories[:input.MaxItems]
	}

	// Calculate statistics
	stats := MemoryStatistics{
		TotalMemories:  len(memories),
		ActiveMemories: len(memories),
	}

	authorCount := make(map[string]int)
	var minDate, maxDate string
	totalSize := 0

	for _, memory := range memories {
		authorCount[memory.GetMetadata().Author]++
		totalSize += len(memory.Content)

		if minDate == "" || memory.DateCreated < minDate {
			minDate = memory.DateCreated
		}
		if maxDate == "" || memory.DateCreated > maxDate {
			maxDate = memory.DateCreated
		}
	}

	stats.TotalSize = totalSize
	if len(memories) > 0 {
		stats.AverageSize = float64(totalSize) / float64(len(memories))
	}

	// Get top authors
	type authorScore struct {
		author string
		count  int
	}
	var authorScores []authorScore
	for author, count := range authorCount {
		authorScores = append(authorScores, authorScore{author, count})
	}
	sort.Slice(authorScores, func(i, j int) bool {
		return authorScores[i].count > authorScores[j].count
	})

	topAuthors := make([]string, 0, min(3, len(authorScores)))
	for i := 0; i < min(3, len(authorScores)); i++ {
		topAuthors = append(topAuthors, authorScores[i].author)
	}

	// Create summary text
	summary := fmt.Sprintf("Found %d active memories", len(memories))
	if input.Author != "" {
		summary += fmt.Sprintf(" by author '%s'", input.Author)
	}
	if minDate != "" && maxDate != "" {
		summary += fmt.Sprintf(", spanning from %s to %s", minDate, maxDate)
	}
	summary += fmt.Sprintf(". Total content size: %d bytes (avg: %.0f bytes/memory).", totalSize, stats.AverageSize)

	// Get most recent memory
	var recentMemory *MemorySummary
	if len(memories) > 0 {
		recent := memories[0]
		meta := recent.GetMetadata()
		recentMemory = &MemorySummary{
			ID:          meta.ID,
			Name:        meta.Name,
			Content:     recent.Content,
			DateCreated: recent.DateCreated,
			Author:      meta.Author,
			IsActive:    meta.IsActive,
		}
	}

	output := SummarizeMemoriesOutput{
		Summary:      summary,
		TotalCount:   len(memories),
		DateRange:    fmt.Sprintf("%s to %s", minDate, maxDate),
		TopAuthors:   topAuthors,
		Statistics:   stats,
		RecentMemory: recentMemory,
	}

	return nil, output, nil
}

// handleUpdateMemory handles the update_memory tool
func (s *MCPServer) handleUpdateMemory(ctx context.Context, req *sdk.CallToolRequest, input UpdateMemoryInput) (*sdk.CallToolResult, UpdateMemoryOutput, error) {
	// Validate input
	if input.ID == "" {
		return nil, UpdateMemoryOutput{}, fmt.Errorf("id is required")
	}

	// Get memory
	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		return nil, UpdateMemoryOutput{}, fmt.Errorf("memory not found: %w", err)
	}

	memory, ok := element.(*domain.Memory)
	if !ok {
		return nil, UpdateMemoryOutput{}, fmt.Errorf("element is not a memory")
	}

	// Update fields
	updated := false

	if input.Content != "" {
		memory.Content = input.Content
		memory.ComputeHash()
		updated = true
	}

	if input.Metadata != nil {
		if memory.Metadata == nil {
			memory.Metadata = make(map[string]string)
		}
		for k, v := range input.Metadata {
			memory.Metadata[k] = v
		}
		updated = true
	}

	meta := memory.GetMetadata()

	if input.Name != "" {
		meta.Name = input.Name
		updated = true
	}

	if input.Description != "" {
		meta.Description = input.Description
		updated = true
	}

	if input.Tags != nil {
		meta.Tags = input.Tags
		updated = true
	}

	if updated {
		meta.UpdatedAt = time.Now()
		memory.SetMetadata(meta)

		if err := s.repo.Update(memory); err != nil {
			return nil, UpdateMemoryOutput{}, fmt.Errorf("failed to update memory: %w", err)
		}
	}

	// Return updated memory
	output := UpdateMemoryOutput{
		Memory: MemorySummary{
			ID:          meta.ID,
			Name:        meta.Name,
			Content:     memory.Content,
			DateCreated: memory.DateCreated,
			Author:      meta.Author,
			IsActive:    meta.IsActive,
		},
	}

	return nil, output, nil
}

// handleDeleteMemory handles the delete_memory tool
func (s *MCPServer) handleDeleteMemory(ctx context.Context, req *sdk.CallToolRequest, input DeleteMemoryInput) (*sdk.CallToolResult, DeleteMemoryOutput, error) {
	// Validate input
	if input.ID == "" {
		return nil, DeleteMemoryOutput{}, fmt.Errorf("id is required")
	}

	// Check if memory exists
	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		return nil, DeleteMemoryOutput{}, fmt.Errorf("memory not found: %w", err)
	}

	if element.GetType() != domain.MemoryElement {
		return nil, DeleteMemoryOutput{}, fmt.Errorf("element is not a memory")
	}

	// Delete memory
	if err := s.repo.Delete(input.ID); err != nil {
		return nil, DeleteMemoryOutput{}, fmt.Errorf("failed to delete memory: %w", err)
	}

	output := DeleteMemoryOutput{
		Success: true,
		Message: fmt.Sprintf("Memory '%s' deleted successfully", element.GetMetadata().Name),
		ID:      input.ID,
	}

	return nil, output, nil
}

// handleClearMemories handles the clear_memories tool
func (s *MCPServer) handleClearMemories(ctx context.Context, req *sdk.CallToolRequest, input ClearMemoriesInput) (*sdk.CallToolResult, ClearMemoriesOutput, error) {
	// Require confirmation
	if !input.Confirm {
		return nil, ClearMemoriesOutput{}, fmt.Errorf("confirmation required: set confirm=true to proceed")
	}

	// Build filter
	filter := domain.ElementFilter{
		Type: func() *domain.ElementType { t := domain.MemoryElement; return &t }(),
	}

	// Get all memories
	elements, err := s.repo.List(filter)
	if err != nil {
		return nil, ClearMemoriesOutput{}, fmt.Errorf("failed to list memories: %w", err)
	}

	// Filter and delete
	deletedCount := 0
	for _, elem := range elements {
		memory, ok := elem.(*domain.Memory)
		if !ok {
			continue
		}

		meta := memory.GetMetadata()

		// Apply filters
		if input.Author != "" && meta.Author != input.Author {
			continue
		}

		if input.DateBefore != "" && memory.DateCreated >= input.DateBefore {
			continue
		}

		// Delete memory
		if err := s.repo.Delete(meta.ID); err != nil {
			// Log error but continue
			continue
		}
		deletedCount++
	}

	message := fmt.Sprintf("Cleared %d memories", deletedCount)
	if input.Author != "" {
		message += fmt.Sprintf(" by author '%s'", input.Author)
	}
	if input.DateBefore != "" {
		message += fmt.Sprintf(" before %s", input.DateBefore)
	}

	output := ClearMemoriesOutput{
		Success:      true,
		DeletedCount: deletedCount,
		Message:      message,
	}

	return nil, output, nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
