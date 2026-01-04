package application

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock ONNX provider for testing
type mockONNXProvider struct {
	available bool
	entities  []EnhancedEntity
	err       error
}

func (m *mockONNXProvider) ExtractEntities(ctx context.Context, text string) ([]EnhancedEntity, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.entities, nil
}

func (m *mockONNXProvider) ExtractEntitiesBatch(ctx context.Context, texts []string) ([][]EnhancedEntity, error) {
	if m.err != nil {
		return nil, m.err
	}
	results := make([][]EnhancedEntity, len(texts))
	for i := range texts {
		results[i] = m.entities
	}
	return results, nil
}

func (m *mockONNXProvider) AnalyzeSentiment(ctx context.Context, text string) (*SentimentResult, error) {
	return nil, nil
}

func (m *mockONNXProvider) ExtractTopics(ctx context.Context, texts []string, numTopics int) ([]Topic, error) {
	return nil, nil
}

func (m *mockONNXProvider) IsAvailable() bool {
	return m.available
}

func TestNewEnhancedEntityExtractor(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockONNXProvider{available: true}

	extractor := NewEnhancedEntityExtractor(config, repo, mockProvider)

	assert.NotNil(t, extractor)
	assert.Equal(t, config, extractor.config)
}

func TestExtractEntitiesFromText(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	config := DefaultEnhancedNLPConfig()

	t.Run("successful_extraction", func(t *testing.T) {
		mockProvider := &mockONNXProvider{
			available: true,
			entities: []EnhancedEntity{
				{
					Type:       EntityTypePerson,
					Value:      "John Smith",
					Confidence: 0.95,
					StartPos:   0,
					EndPos:     10,
				},
				{
					Type:       EntityTypeOrganization,
					Value:      "OpenAI",
					Confidence: 0.92,
					StartPos:   25,
					EndPos:     31,
				},
				{
					Type:       EntityTypeLocation,
					Value:      "San Francisco",
					Confidence: 0.88,
					StartPos:   40,
					EndPos:     53,
				},
			},
		}

		extractor := NewEnhancedEntityExtractor(config, repo, mockProvider)
		result, err := extractor.ExtractFromText(context.Background(), "John Smith works for OpenAI in San Francisco")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result.Entities))
		assert.Equal(t, EntityTypePerson, result.Entities[0].Type)
		assert.Equal(t, "John Smith", result.Entities[0].Value)
		assert.GreaterOrEqual(t, result.Entities[0].Confidence, 0.9)
	})

	t.Run("onnx_unavailable_fallback", func(t *testing.T) {
		mockProvider := &mockONNXProvider{available: false}

		extractor := NewEnhancedEntityExtractor(config, repo, mockProvider)
		result, err := extractor.ExtractFromText(context.Background(), "Test text")

		// Should use fallback
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result.ModelUsed, "fallback")
	})
}

func TestExtractEntitiesFromMemory(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create test memory
	memory := domain.NewMemory(
		"test-entity",
		"Test memory for entities",
		"v1.0.0",
		"test-user",
	)
	memory.Content = "Apple Inc. was founded by Steve Jobs in Cupertino, California."
	require.NoError(t, repo.Create(memory))

	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockONNXProvider{
		available: true,
		entities: []EnhancedEntity{
			{
				Type:       EntityTypeOrganization,
				Value:      "Apple Inc.",
				Confidence: 0.96,
			},
			{
				Type:       EntityTypePerson,
				Value:      "Steve Jobs",
				Confidence: 0.98,
			},
			{
				Type:       EntityTypeLocation,
				Value:      "Cupertino, California",
				Confidence: 0.94,
			},
		},
	}

	extractor := NewEnhancedEntityExtractor(config, repo, mockProvider)
	result, err := extractor.ExtractFromMemory(ctx, memory.GetID())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Entities), 3)
}

func TestExtractEntitiesFromMultipleMemories(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create test memories
	memoryIDs := make([]string, 0)
	for i := 0; i < 3; i++ {
		memory := domain.NewMemory(
			"batch-entity-"+string(rune('a'+i)),
			"Batch entity test",
			"v1.0.0",
			"test-user",
		)
		memory.Content = "Test content with entities"
		require.NoError(t, repo.Create(memory))
		memoryIDs = append(memoryIDs, memory.GetID())
	}

	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockONNXProvider{
		available: true,
		entities: []EnhancedEntity{
			{
				Type:       EntityTypeTechnology,
				Value:      "Python",
				Confidence: 0.89,
			},
		},
	}

	extractor := NewEnhancedEntityExtractor(config, repo, mockProvider)
	results, err := extractor.ExtractFromMemoryBatch(ctx, memoryIDs)

	require.NoError(t, err)
	assert.Equal(t, 3, len(results))
}

func TestFindRelationships(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create test memory with entities that have relationships
	memory := domain.NewMemory(
		"test-relationships",
		"Test memory for relationships",
		"v1.0.0",
		"test-user",
	)
	memory.Content = "Steve Jobs founded Apple Inc. in Cupertino."
	require.NoError(t, repo.Create(memory))

	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockONNXProvider{
		available: true,
		entities: []EnhancedEntity{
			{
				Type:       EntityTypePerson,
				Value:      "Steve Jobs",
				Confidence: 0.95,
			},
			{
				Type:       EntityTypeOrganization,
				Value:      "Apple Inc.",
				Confidence: 0.92,
			},
		},
	}

	extractor := NewEnhancedEntityExtractor(config, repo, mockProvider)
	result, err := extractor.ExtractFromMemory(ctx, memory.GetID())

	require.NoError(t, err)
	assert.NotNil(t, result)
	// Relationships are extracted automatically as part of ExtractFromMemory
	assert.NotNil(t, result.Relationships)
}

func TestEnhancedEntityExtractorWithInvalidMemory(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	config := DefaultEnhancedNLPConfig()
	mockProvider := &mockONNXProvider{available: true}

	extractor := NewEnhancedEntityExtractor(config, repo, mockProvider)
	_, err := extractor.ExtractFromMemory(context.Background(), "non-existent-id")

	assert.Error(t, err)
}

func TestDefaultEnhancedNLPConfig(t *testing.T) {
	config := DefaultEnhancedNLPConfig()

	assert.Greater(t, config.EntityConfidenceMin, 0.0)
	assert.Greater(t, config.EntityMaxPerDoc, 0)
	assert.True(t, config.EnableFallback)
	assert.Greater(t, config.BatchSize, 0)
	assert.Greater(t, config.MaxLength, 0)
}
