package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// Entity represents an extracted entity from memory content.
type Entity struct {
	Type  string `json:"type"`  // person, organization, location, concept, etc.
	Value string `json:"value"` // The entity text
	Count int    `json:"count"` // Frequency in the content
}

// KnowledgeGraph represents extracted knowledge from memories.
type KnowledgeGraph struct {
	Entities      []Entity              `json:"entities"`
	Relationships []domain.Relationship `json:"relationships"`
	Concepts      map[string]int        `json:"concepts"` // Concept → frequency
	Keywords      []string              `json:"keywords"`
	Summary       string                `json:"summary,omitempty"`
}

// KnowledgeGraphExtractor extracts entities and relationships from memory content.
type KnowledgeGraphExtractor struct {
	repository ElementRepository
}

// NewKnowledgeGraphExtractor creates a new knowledge graph extractor.
func NewKnowledgeGraphExtractor(repository ElementRepository) *KnowledgeGraphExtractor {
	return &KnowledgeGraphExtractor{
		repository: repository,
	}
}

// ExtractFromMemory extracts knowledge graph from a single memory.
func (k *KnowledgeGraphExtractor) ExtractFromMemory(ctx context.Context, memoryID string) (*KnowledgeGraph, error) {
	element, err := k.repository.GetByID(memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	memory, ok := element.(*domain.Memory)
	if !ok {
		return nil, fmt.Errorf("element %s is not a memory", memoryID)
	}

	return k.extractFromContent(memory.Content)
}

// ExtractFromCluster extracts knowledge graph from a cluster of memories.
func (k *KnowledgeGraphExtractor) ExtractFromCluster(ctx context.Context, cluster *Cluster) (*KnowledgeGraph, error) {
	if cluster.Size == 0 {
		return &KnowledgeGraph{
			Entities:      []Entity{},
			Relationships: []domain.Relationship{},
			Concepts:      make(map[string]int),
			Keywords:      []string{},
		}, nil
	}

	// Combine all memory content
	var combinedContent strings.Builder
	for _, memory := range cluster.Members {
		combinedContent.WriteString(memory.Content)
		combinedContent.WriteString("\n")
	}

	graph, err := k.extractFromContent(combinedContent.String())
	if err != nil {
		return nil, err
	}

	// Generate summary for cluster
	graph.Summary = fmt.Sprintf("Knowledge extracted from cluster %d with %d memories", cluster.ID, cluster.Size)

	return graph, nil
}

// ExtractFromMultipleMemories extracts knowledge graph from multiple memories.
func (k *KnowledgeGraphExtractor) ExtractFromMultipleMemories(ctx context.Context, memoryIDs []string) (*KnowledgeGraph, error) {
	var combinedContent strings.Builder

	for _, memoryID := range memoryIDs {
		element, err := k.repository.GetByID(memoryID)
		if err != nil {
			continue // Skip missing memories
		}

		if memory, ok := element.(*domain.Memory); ok {
			combinedContent.WriteString(memory.Content)
			combinedContent.WriteString("\n")
		}
	}

	return k.extractFromContent(combinedContent.String())
}

// extractFromContent performs the actual extraction from text content.
func (k *KnowledgeGraphExtractor) extractFromContent(content string) (*KnowledgeGraph, error) {
	graph := &KnowledgeGraph{
		Entities:      []Entity{},
		Relationships: []domain.Relationship{},
		Concepts:      make(map[string]int),
		Keywords:      []string{},
	}

	// Extract entities
	graph.Entities = k.extractEntities(content)

	// Extract relationships (simple heuristic-based)
	graph.Relationships = k.extractRelationships(content, graph.Entities)

	// Extract concepts and keywords
	graph.Concepts = k.extractConcepts(content)
	graph.Keywords = k.extractKeywords(content)

	return graph, nil
}

// extractEntities extracts named entities using regex patterns.
func (k *KnowledgeGraphExtractor) extractEntities(content string) []Entity {
	entities := make(map[string]*Entity)

	// Pattern for capitalized words (potential proper nouns)
	properNounPattern := regexp.MustCompile(`\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\b`)
	matches := properNounPattern.FindAllString(content, -1)

	for _, match := range matches {
		// Skip common words
		if isCommonWord(match) {
			continue
		}

		entityType := inferEntityType(match)
		key := fmt.Sprintf("%s:%s", entityType, match)

		if entity, exists := entities[key]; exists {
			entity.Count++
		} else {
			entities[key] = &Entity{
				Type:  entityType,
				Value: match,
				Count: 1,
			}
		}
	}

	// Extract URLs
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlPattern.FindAllString(content, -1)
	for _, url := range urls {
		key := fmt.Sprintf("url:%s", url)
		if entity, exists := entities[key]; exists {
			entity.Count++
		} else {
			entities[key] = &Entity{
				Type:  "url",
				Value: url,
				Count: 1,
			}
		}
	}

	// Extract email addresses
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailPattern.FindAllString(content, -1)
	for _, email := range emails {
		key := fmt.Sprintf("email:%s", email)
		if entity, exists := entities[key]; exists {
			entity.Count++
		} else {
			entities[key] = &Entity{
				Type:  "email",
				Value: email,
				Count: 1,
			}
		}
	}

	// Convert map to slice
	result := make([]Entity, 0, len(entities))
	for _, entity := range entities {
		result = append(result, *entity)
	}

	return result
}

// extractRelationships extracts relationships using simple patterns.
func (k *KnowledgeGraphExtractor) extractRelationships(content string, entities []Entity) []domain.Relationship {
	relationships := []domain.Relationship{}

	// Simple pattern-based relationship extraction
	// Pattern: "X is Y" or "X has Y" or "X uses Y"
	patterns := []struct {
		regex   string
		relType string
	}{
		{`(\w+)\s+is\s+(\w+)`, "is_a"},
		{`(\w+)\s+has\s+(\w+)`, "has"},
		{`(\w+)\s+uses\s+(\w+)`, "uses"},
		{`(\w+)\s+works with\s+(\w+)`, "works_with"},
		{`(\w+)\s+depends on\s+(\w+)`, "depends_on"},
		{`(\w+)\s+related to\s+(\w+)`, "related_to"},
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern.regex)
		matches := re.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) >= 3 {
				rel := domain.Relationship{
					SourceID:   fmt.Sprintf("entity:%s", match[1]),
					TargetID:   fmt.Sprintf("entity:%s", match[2]),
					Type:       domain.RelationshipType(pattern.relType),
					Properties: make(map[string]string),
				}
				relationships = append(relationships, rel)
			}
		}
	}

	return relationships
}

// extractConcepts extracts key concepts and their frequency.
func (k *KnowledgeGraphExtractor) extractConcepts(content string) map[string]int {
	concepts := make(map[string]int)

	// Convert to lowercase and split into words
	words := strings.Fields(strings.ToLower(content))

	for _, word := range words {
		// Clean word (remove punctuation)
		word = cleanWord(word)
		if len(word) < 3 {
			continue // Skip short words
		}

		// Skip stop words
		if isStopWord(word) {
			continue
		}

		concepts[word]++
	}

	// Keep only concepts with frequency > 1
	filtered := make(map[string]int)
	for concept, count := range concepts {
		if count > 1 {
			filtered[concept] = count
		}
	}

	return filtered
}

// extractKeywords extracts important keywords (top N concepts).
func (k *KnowledgeGraphExtractor) extractKeywords(content string) []string {
	concepts := k.extractConcepts(content)

	// Sort by frequency
	type kv struct {
		Key   string
		Value int
	}

	var sorted []kv
	for k, v := range concepts {
		sorted = append(sorted, kv{k, v})
	}

	// Sort descending
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Value > sorted[i].Value {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Take top 10
	keywords := []string{}
	max := 10
	if len(sorted) < max {
		max = len(sorted)
	}
	for i := 0; i < max; i++ {
		keywords = append(keywords, sorted[i].Key)
	}

	return keywords
}

// inferEntityType infers the type of entity based on patterns.
func inferEntityType(text string) string {
	// Check for common patterns
	if strings.Contains(strings.ToLower(text), "inc") ||
		strings.Contains(strings.ToLower(text), "corp") ||
		strings.Contains(strings.ToLower(text), "ltd") {
		return "organization"
	}

	// Check for person names (simple heuristic)
	words := strings.Fields(text)
	if len(words) >= 2 && len(words[0]) > 0 && len(words[1]) > 0 {
		return "person"
	}

	// Check for locations (very simple)
	locations := []string{"city", "country", "state", "province"}
	for _, loc := range locations {
		if strings.Contains(strings.ToLower(text), loc) {
			return "location"
		}
	}

	return "concept" // Default
}

// isCommonWord checks if a word is too common to be an entity.
func isCommonWord(word string) bool {
	common := []string{
		"The", "This", "That", "These", "Those",
		"I", "You", "He", "She", "It", "We", "They",
		"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday",
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	for _, c := range common {
		if word == c {
			return true
		}
	}
	return false
}

// cleanWord removes punctuation from a word.
func cleanWord(word string) string {
	// Remove punctuation
	cleaned := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(word, "")
	return cleaned
}

// isStopWord checks if a word is a stop word.
func isStopWord(word string) bool {
	stopWords := []string{
		"the", "be", "to", "of", "and", "a", "in", "that", "have", "i",
		"it", "for", "not", "on", "with", "he", "as", "you", "do", "at",
		"this", "but", "his", "by", "from", "they", "we", "say", "her", "she",
		"or", "an", "will", "my", "one", "all", "would", "there", "their",
		"what", "so", "up", "out", "if", "about", "who", "get", "which", "go",
		"me", "when", "make", "can", "like", "time", "no", "just", "him", "know",
		"take", "into", "year", "your", "good", "some", "could", "them", "see",
		"other", "than", "then", "now", "look", "only", "come", "its", "over",
		"think", "also", "back", "after", "use", "two", "how", "our", "work",
		"first", "well", "way", "even", "new", "want", "because", "any", "these",
		"give", "day", "most", "us", "is", "was", "are", "been", "has", "had",
		"were", "said", "did", "having", "may", "should", "could", "might",
	}

	for _, sw := range stopWords {
		if word == sw {
			return true
		}
	}
	return false
}
