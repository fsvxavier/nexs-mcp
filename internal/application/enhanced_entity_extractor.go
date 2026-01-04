package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// EntityType represents the type of extracted entity.
type EntityType string

const (
	EntityTypePerson       EntityType = "PERSON"
	EntityTypeOrganization EntityType = "ORGANIZATION"
	EntityTypeLocation     EntityType = "LOCATION"
	EntityTypeDate         EntityType = "DATE"
	EntityTypeEvent        EntityType = "EVENT"
	EntityTypeProduct      EntityType = "PRODUCT"
	EntityTypeTechnology   EntityType = "TECHNOLOGY"
	EntityTypeConcept      EntityType = "CONCEPT"
	EntityTypeOther        EntityType = "OTHER"
)

// EnhancedEntity represents an entity extracted using transformer models.
type EnhancedEntity struct {
	Type       EntityType        `json:"type"`
	Value      string            `json:"value"`
	Confidence float64           `json:"confidence"` // 0.0-1.0 confidence score
	StartPos   int               `json:"start_pos"`  // Character position in text
	EndPos     int               `json:"end_pos"`    // Character position in text
	Context    string            `json:"context"`    // Surrounding text for disambiguation
	Metadata   map[string]string `json:"metadata"`   // Additional attributes
}

// EntityExtractionResult contains the extraction results with metadata.
type EntityExtractionResult struct {
	Entities       []EnhancedEntity     `json:"entities"`
	ProcessingTime float64              `json:"processing_time_ms"`
	ModelUsed      string               `json:"model_used"`
	Confidence     float64              `json:"avg_confidence"`
	Relationships  []EntityRelationship `json:"relationships"`
}

// EntityRelationship represents a relationship between two entities.
type EntityRelationship struct {
	SourceEntity string  `json:"source_entity"`
	TargetEntity string  `json:"target_entity"`
	RelationType string  `json:"relation_type"`
	Confidence   float64 `json:"confidence"`
	Evidence     string  `json:"evidence"` // Text that supports this relationship
}

// EnhancedEntityExtractor provides advanced NER using transformer models via ONNX.
type EnhancedEntityExtractor struct {
	config           EnhancedNLPConfig
	repository       ElementRepository
	modelProvider    ONNXModelProvider // Interface for ONNX model inference
	fallbackEnabled  bool
	classicExtractor *KnowledgeGraphExtractor
}

// EnhancedNLPConfig holds configuration for enhanced NLP features.
type EnhancedNLPConfig struct {
	// Entity Extraction
	EntityModel          string  // ONNX model path for entity extraction (e.g., bert-base-ner)
	EntityConfidenceMin  float64 // Minimum confidence threshold (default: 0.7)
	EntityMaxPerDoc      int     // Maximum entities to extract per document (default: 100)
	EnableDisambiguation bool    // Enable entity disambiguation (default: true)

	// Sentiment Analysis
	SentimentModel     string  // ONNX model for sentiment (e.g., distilbert-sentiment)
	SentimentThreshold float64 // Confidence threshold for sentiment (default: 0.6)

	// Topic Modeling
	TopicModel string // Model for topic extraction
	TopicCount int    // Number of topics to extract (default: 5)

	// Performance
	BatchSize int  // Batch size for processing (default: 16)
	MaxLength int  // Maximum sequence length for tokenization (default: 512)
	UseGPU    bool // Enable GPU acceleration if available

	// Fallback
	EnableFallback bool // Fall back to rule-based if transformer fails
}

// DefaultEnhancedNLPConfig returns default configuration.
func DefaultEnhancedNLPConfig() EnhancedNLPConfig {
	return EnhancedNLPConfig{
		EntityModel:          "models/bert-base-ner/model.onnx",
		EntityConfidenceMin:  0.7,
		EntityMaxPerDoc:      100,
		EnableDisambiguation: true,

		SentimentModel:     "models/distilbert-sentiment/model.onnx",
		SentimentThreshold: 0.6,

		TopicCount: 5,

		BatchSize: 16,
		MaxLength: 512,
		UseGPU:    false,

		EnableFallback: true,
	}
}

// Topic represents a discovered topic with keywords.
type Topic struct {
	ID       string   `json:"id"`
	Keywords []string `json:"keywords"`
	Weight   float64  `json:"weight"`
}

// ONNXModelProvider defines interface for ONNX inference.
type ONNXModelProvider interface {
	// ExtractEntities runs NER model on text
	ExtractEntities(ctx context.Context, text string) ([]EnhancedEntity, error)

	// ExtractEntitiesBatch runs NER on multiple texts
	ExtractEntitiesBatch(ctx context.Context, texts []string) ([][]EnhancedEntity, error)

	// AnalyzeSentiment runs sentiment analysis model
	AnalyzeSentiment(ctx context.Context, text string) (*SentimentResult, error)

	// ExtractTopics runs topic modeling
	ExtractTopics(ctx context.Context, texts []string, numTopics int) ([]Topic, error)

	// IsAvailable checks if ONNX runtime and models are available
	IsAvailable() bool
}

// NewEnhancedEntityExtractor creates a new enhanced entity extractor.
func NewEnhancedEntityExtractor(
	config EnhancedNLPConfig,
	repository ElementRepository,
	modelProvider ONNXModelProvider,
) *EnhancedEntityExtractor {
	// Create fallback extractor
	var fallbackExtractor *KnowledgeGraphExtractor
	if config.EnableFallback {
		fallbackExtractor = NewKnowledgeGraphExtractor(repository)
	}

	return &EnhancedEntityExtractor{
		config:           config,
		repository:       repository,
		modelProvider:    modelProvider,
		fallbackEnabled:  config.EnableFallback,
		classicExtractor: fallbackExtractor,
	}
}

// ExtractFromMemory extracts entities from a memory using transformer models.
func (e *EnhancedEntityExtractor) ExtractFromMemory(ctx context.Context, memoryID string) (*EntityExtractionResult, error) {
	// Get memory
	element, err := e.repository.GetByID(memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	memory, ok := element.(*domain.Memory)
	if !ok {
		return nil, fmt.Errorf("element %s is not a memory", memoryID)
	}

	return e.ExtractFromText(ctx, memory.Content)
}

// ExtractFromText extracts entities from raw text.
func (e *EnhancedEntityExtractor) ExtractFromText(ctx context.Context, text string) (*EntityExtractionResult, error) {
	// Check if ONNX is available
	if !e.modelProvider.IsAvailable() {
		if e.fallbackEnabled && e.classicExtractor != nil {
			return e.fallbackToClassicExtraction(ctx, text)
		}
		return nil, errors.New("ONNX runtime not available and fallback disabled")
	}

	// Extract entities using transformer model
	entities, err := e.modelProvider.ExtractEntities(ctx, text)
	if err != nil {
		if e.fallbackEnabled && e.classicExtractor != nil {
			return e.fallbackToClassicExtraction(ctx, text)
		}
		return nil, fmt.Errorf("entity extraction failed: %w", err)
	}

	// Filter by confidence threshold
	filteredEntities := e.filterByConfidence(entities)

	// Extract relationships between entities
	relationships := e.extractRelationships(text, filteredEntities)

	// Calculate average confidence
	avgConfidence := e.calculateAverageConfidence(filteredEntities)

	result := &EntityExtractionResult{
		Entities:       filteredEntities,
		ProcessingTime: 0, // TODO: Add timing
		ModelUsed:      e.config.EntityModel,
		Confidence:     avgConfidence,
		Relationships:  relationships,
	}

	return result, nil
}

// ExtractFromMemoryBatch extracts entities from multiple memories in batch.
func (e *EnhancedEntityExtractor) ExtractFromMemoryBatch(ctx context.Context, memoryIDs []string) ([]*EntityExtractionResult, error) {
	// Get all memories
	texts := make([]string, 0, len(memoryIDs))
	validIndices := make([]int, 0, len(memoryIDs))

	for i, memoryID := range memoryIDs {
		element, err := e.repository.GetByID(memoryID)
		if err != nil {
			continue // Skip missing memories
		}

		if memory, ok := element.(*domain.Memory); ok {
			texts = append(texts, memory.Content)
			validIndices = append(validIndices, i)
		}
	}

	if len(texts) == 0 {
		return nil, errors.New("no valid memories found")
	}

	// Batch extraction
	batchResults, err := e.modelProvider.ExtractEntitiesBatch(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("batch entity extraction failed: %w", err)
	}

	// Convert to results
	results := make([]*EntityExtractionResult, len(memoryIDs))
	for i, entities := range batchResults {
		idx := validIndices[i]

		filteredEntities := e.filterByConfidence(entities)
		relationships := e.extractRelationships(texts[i], filteredEntities)
		avgConfidence := e.calculateAverageConfidence(filteredEntities)

		results[idx] = &EntityExtractionResult{
			Entities:       filteredEntities,
			ProcessingTime: 0,
			ModelUsed:      e.config.EntityModel,
			Confidence:     avgConfidence,
			Relationships:  relationships,
		}
	}

	return results, nil
}

// filterByConfidence filters entities based on confidence threshold.
func (e *EnhancedEntityExtractor) filterByConfidence(entities []EnhancedEntity) []EnhancedEntity {
	filtered := make([]EnhancedEntity, 0, len(entities))

	for _, entity := range entities {
		if entity.Confidence >= e.config.EntityConfidenceMin {
			filtered = append(filtered, entity)
		}
	}

	// Limit to max entities
	if len(filtered) > e.config.EntityMaxPerDoc {
		filtered = filtered[:e.config.EntityMaxPerDoc]
	}

	return filtered
}

// extractRelationships extracts relationships between entities.
func (e *EnhancedEntityExtractor) extractRelationships(text string, entities []EnhancedEntity) []EntityRelationship {
	relationships := make([]EntityRelationship, 0)

	// Simple co-occurrence based relationship extraction
	// TODO: Enhance with dependency parsing and semantic role labeling

	for i := range entities {
		for j := i + 1; j < len(entities); j++ {
			entity1 := entities[i]
			entity2 := entities[j]

			// Check if entities are mentioned in close proximity (within 50 characters)
			distance := entity2.StartPos - entity1.EndPos
			if distance > 0 && distance < 50 {
				// Extract evidence text between entities
				evidence := text[entity1.EndPos:entity2.StartPos]
				evidence = strings.TrimSpace(evidence)

				// Infer relationship type
				relationType := e.inferRelationType(entity1.Type, entity2.Type, evidence)

				if relationType != "" {
					relationships = append(relationships, EntityRelationship{
						SourceEntity: entity1.Value,
						TargetEntity: entity2.Value,
						RelationType: relationType,
						Confidence:   (entity1.Confidence + entity2.Confidence) / 2,
						Evidence:     evidence,
					})
				}
			}
		}
	}

	return relationships
}

// inferRelationType infers the relationship type based on entity types and context.
func (e *EnhancedEntityExtractor) inferRelationType(type1, type2 EntityType, evidence string) string {
	evidence = strings.ToLower(evidence)

	// Person-Organization relationships
	if type1 == EntityTypePerson && type2 == EntityTypeOrganization {
		if strings.Contains(evidence, "work") || strings.Contains(evidence, "employ") {
			return "WORKS_AT"
		}
		if strings.Contains(evidence, "found") || strings.Contains(evidence, "ceo") {
			return "FOUNDED"
		}
		return "AFFILIATED_WITH"
	}

	// Person-Location relationships
	if type1 == EntityTypePerson && type2 == EntityTypeLocation {
		if strings.Contains(evidence, "born") || strings.Contains(evidence, "from") {
			return "BORN_IN"
		}
		if strings.Contains(evidence, "live") || strings.Contains(evidence, "resid") {
			return "LIVES_IN"
		}
		return "LOCATED_IN"
	}

	// Organization-Location relationships
	if type1 == EntityTypeOrganization && type2 == EntityTypeLocation {
		if strings.Contains(evidence, "headquarter") || strings.Contains(evidence, "based") {
			return "HEADQUARTERED_IN"
		}
		return "LOCATED_IN"
	}

	// Technology-Organization relationships
	if type1 == EntityTypeTechnology && type2 == EntityTypeOrganization {
		if strings.Contains(evidence, "develop") || strings.Contains(evidence, "created") {
			return "DEVELOPED_BY"
		}
		return "USED_BY"
	}

	// Generic relationship
	return "RELATED_TO"
}

// calculateAverageConfidence calculates the average confidence score.
func (e *EnhancedEntityExtractor) calculateAverageConfidence(entities []EnhancedEntity) float64 {
	if len(entities) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, entity := range entities {
		sum += entity.Confidence
	}

	return sum / float64(len(entities))
}

// fallbackToClassicExtraction falls back to rule-based extraction.
func (e *EnhancedEntityExtractor) fallbackToClassicExtraction(ctx context.Context, text string) (*EntityExtractionResult, error) {
	graph, err := e.classicExtractor.extractFromContent(text)
	if err != nil {
		return nil, fmt.Errorf("fallback extraction failed: %w", err)
	}

	// Convert classic entities to enhanced format
	enhancedEntities := make([]EnhancedEntity, len(graph.Entities))
	for i, entity := range graph.Entities {
		enhancedEntities[i] = EnhancedEntity{
			Type:       EntityType(strings.ToUpper(entity.Type)),
			Value:      entity.Value,
			Confidence: 0.5, // Lower confidence for rule-based
			StartPos:   -1,  // Position not available
			EndPos:     -1,
			Context:    "",
			Metadata:   map[string]string{"source": "fallback"},
		}
	}

	// Convert relationships
	relationships := make([]EntityRelationship, len(graph.Relationships))
	for i, rel := range graph.Relationships {
		relationships[i] = EntityRelationship{
			SourceEntity: rel.SourceID,
			TargetEntity: rel.TargetID,
			RelationType: string(rel.Type),
			Confidence:   0.5,
			Evidence:     "",
		}
	}

	return &EntityExtractionResult{
		Entities:       enhancedEntities,
		ProcessingTime: 0,
		ModelUsed:      "rule-based-fallback",
		Confidence:     0.5,
		Relationships:  relationships,
	}, nil
}

// DisambiguateEntity attempts to disambiguate an entity against a knowledge base.
func (e *EnhancedEntityExtractor) DisambiguateEntity(ctx context.Context, entity EnhancedEntity, context string) (*EnhancedEntity, error) {
	if !e.config.EnableDisambiguation {
		return &entity, nil
	}

	// TODO: Implement entity disambiguation using:
	// 1. Context similarity
	// 2. Knowledge base lookup
	// 3. Entity linking

	// For now, return entity as-is
	return &entity, nil
}
