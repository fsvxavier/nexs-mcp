package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// SaveConversationContextInput defines input for save_conversation_context tool.
type SaveConversationContextInput struct {
	Context    string   `json:"context"              jsonschema:"conversation context to save as memory"`
	Summary    string   `json:"summary,omitempty"    jsonschema:"brief summary of the context"`
	Tags       []string `json:"tags,omitempty"       jsonschema:"tags for categorization"`
	Importance string   `json:"importance,omitempty" jsonschema:"importance level: low, medium, high, critical"`
	RelatedTo  []string `json:"related_to,omitempty" jsonschema:"IDs of related elements"`
}

// SaveConversationContextOutput defines output for save_conversation_context tool.
type SaveConversationContextOutput struct {
	MemoryID string `json:"memory_id"`
	Saved    bool   `json:"saved"`
	Message  string `json:"message"`
}

// handleSaveConversationContext handles automatic saving of conversation context.
func (s *MCPServer) handleSaveConversationContext(ctx context.Context, req *sdk.CallToolRequest, input SaveConversationContextInput) (*sdk.CallToolResult, SaveConversationContextOutput, error) {
	// Check if auto-save is enabled
	if !s.cfg.AutoSaveMemories {
		return nil, SaveConversationContextOutput{
			Saved:   false,
			Message: "Auto-save memories is disabled",
		}, nil
	}

	// Validate input
	if input.Context == "" || len(input.Context) < 10 {
		return nil, SaveConversationContextOutput{}, errors.New("context must be at least 10 characters")
	}

	// Generate name from summary or first line of context
	name := input.Summary
	if name == "" {
		lines := strings.Split(input.Context, "\n")
		name = lines[0]
		if len(name) > 80 {
			name = name[:80] + "..."
		}
	}

	// Create timestamp-based name
	timestamp := time.Now().Format("2006-01-02 15:04")
	memoryName := "Conversation Context - " + timestamp
	if name != "" {
		memoryName = name
	}

	// Create memory
	memory := domain.NewMemory(memoryName, input.Summary, "1.0.0", "auto-save")
	memory.Content = input.Context
	memory.ComputeHash()

	// Set tags
	tags := input.Tags
	if tags == nil {
		tags = []string{"auto-save", "conversation"}
	} else {
		tags = append(tags, "auto-save", "conversation")
	}

	// Add importance tag if specified
	if input.Importance != "" {
		tags = append(tags, "importance:"+input.Importance)
	}

	// Add search index based on content
	searchTerms := extractKeywords(input.Context, 10)
	memory.SearchIndex = searchTerms

	// Add metadata
	memory.Metadata = map[string]string{
		"auto_saved": "true",
		"saved_at":   time.Now().Format(time.RFC3339),
		"importance": input.Importance,
	}

	// Add related elements if specified
	if len(input.RelatedTo) > 0 {
		memory.Metadata["related_to"] = strings.Join(input.RelatedTo, ",")
	}

	// Set tags
	metadata := memory.GetMetadata()
	metadata.Tags = tags
	memory.SetMetadata(metadata)

	// Validate
	if err := memory.Validate(); err != nil {
		return nil, SaveConversationContextOutput{}, fmt.Errorf("memory validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(memory); err != nil {
		return nil, SaveConversationContextOutput{}, fmt.Errorf("failed to save conversation context: %w", err)
	}

	output := SaveConversationContextOutput{
		MemoryID: memory.GetID(),
		Saved:    true,
		Message:  "Conversation context saved successfully",
	}

	return nil, output, nil
}

// extractKeywords extracts relevant keywords from text for search indexing.
func extractKeywords(text string, maxKeywords int) []string {
	// Simple keyword extraction - can be enhanced with NLP
	words := strings.Fields(strings.ToLower(text))

	// Common words to filter out (stop words)
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
		"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
		"that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
		// Portuguese stop words
		"o": true, "os": true, "um": true, "uma": true,
		"de": true, "da": true, "do": true, "dos": true, "das": true, "em": true,
		"no": true, "na": true, "nos": true, "nas": true, "para": true, "pelo": true,
		"pela": true, "com": true, "sem": true, "por": true, "ao": true, "à": true,
		"foi": true, "ser": true, "está": true, "são": true, "essa": true, "esse": true,
	}

	// Count word frequency (excluding stop words)
	wordFreq := make(map[string]int)
	for _, word := range words {
		// Clean word (remove punctuation)
		word = strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(word) < 3 || stopWords[word] {
			continue
		}
		wordFreq[word]++
	}

	// Get top keywords
	type wordCount struct {
		word  string
		count int
	}
	var counts []wordCount
	for word, count := range wordFreq {
		counts = append(counts, wordCount{word, count})
	}

	// Sort by frequency (simple bubble sort for small lists)
	for i := 0; i < len(counts); i++ {
		for j := i + 1; j < len(counts); j++ {
			if counts[j].count > counts[i].count {
				counts[i], counts[j] = counts[j], counts[i]
			}
		}
	}

	// Extract top keywords
	keywords := []string{}
	limit := maxKeywords
	if len(counts) < limit {
		limit = len(counts)
	}
	for i := range limit {
		keywords = append(keywords, counts[i].word)
	}

	return keywords
}
