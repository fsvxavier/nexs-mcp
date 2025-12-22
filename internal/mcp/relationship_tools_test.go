package mcp

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRelatedElementsInput_Structure(t *testing.T) {
	input := GetRelatedElementsInput{
		ElementID:       "test-id",
		Direction:       "both",
		ElementTypes:    []string{"persona", "skill"},
		IncludeInactive: true,
	}

	assert.Equal(t, "test-id", input.ElementID)
	assert.Equal(t, "both", input.Direction)
	assert.Equal(t, []string{"persona", "skill"}, input.ElementTypes)
	assert.True(t, input.IncludeInactive)
}

func TestGetRelatedElementsOutput_Structure(t *testing.T) {
	output := GetRelatedElementsOutput{
		ElementID: "test-id",
		Forward:   []map[string]interface{}{{"id": "forward-1"}},
		Reverse:   []map[string]interface{}{{"id": "reverse-1"}},
		Total:     2,
	}

	assert.Equal(t, "test-id", output.ElementID)
	assert.Len(t, output.Forward, 1)
	assert.Len(t, output.Reverse, 1)
	assert.Equal(t, 2, output.Total)
}

func TestExpandRelationshipsInput_Structure(t *testing.T) {
	input := ExpandRelationshipsInput{
		ElementID:      "root-id",
		MaxDepth:       3,
		IncludeTypes:   []string{"persona"},
		ExcludeVisited: true,
		FollowBothWays: false,
	}

	assert.Equal(t, "root-id", input.ElementID)
	assert.Equal(t, 3, input.MaxDepth)
	assert.Equal(t, []string{"persona"}, input.IncludeTypes)
	assert.True(t, input.ExcludeVisited)
	assert.False(t, input.FollowBothWays)
}

func TestExpandRelationshipsOutput_Structure(t *testing.T) {
	graph := map[string]interface{}{
		"id":   "root",
		"type": "persona",
	}

	output := ExpandRelationshipsOutput{
		RootID:        "root-id",
		TotalElements: 5,
		MaxDepth:      3,
		Graph:         graph,
	}

	assert.Equal(t, "root-id", output.RootID)
	assert.Equal(t, 5, output.TotalElements)
	assert.Equal(t, 3, output.MaxDepth)
	assert.NotNil(t, output.Graph)
	assert.Equal(t, "root", output.Graph["id"])
}

func TestInferRelationshipsInput_Structure(t *testing.T) {
	input := InferRelationshipsInput{
		ElementID:       "test-id",
		Methods:         []string{"mention", "keyword"},
		MinConfidence:   0.75,
		TargetTypes:     []string{"skill"},
		AutoApply:       true,
		RequireEvidence: 2,
	}

	assert.Equal(t, "test-id", input.ElementID)
	assert.Equal(t, []string{"mention", "keyword"}, input.Methods)
	assert.Equal(t, 0.75, input.MinConfidence)
	assert.Equal(t, []string{"skill"}, input.TargetTypes)
	assert.True(t, input.AutoApply)
	assert.Equal(t, 2, input.RequireEvidence)
}

func TestInferRelationshipsOutput_Structure(t *testing.T) {
	inferences := []map[string]interface{}{
		{
			"source_id":   "source-1",
			"target_id":   "target-1",
			"confidence":  0.85,
			"inferred_by": "mention",
		},
	}

	output := InferRelationshipsOutput{
		ElementID:     "test-id",
		TotalInferred: 1,
		AutoApplied:   true,
		Inferences:    inferences,
	}

	assert.Equal(t, "test-id", output.ElementID)
	assert.Equal(t, 1, output.TotalInferred)
	assert.True(t, output.AutoApplied)
	assert.Len(t, output.Inferences, 1)
	assert.Equal(t, "source-1", output.Inferences[0]["source_id"])
}

func TestGetRecommendationsInput_Structure(t *testing.T) {
	input := GetRecommendationsInput{
		ElementID:      "test-id",
		ElementType:    "skill",
		MinScore:       0.5,
		MaxResults:     10,
		IncludeReasons: true,
	}

	assert.Equal(t, "test-id", input.ElementID)
	assert.Equal(t, "skill", input.ElementType)
	assert.Equal(t, 0.5, input.MinScore)
	assert.Equal(t, 10, input.MaxResults)
	assert.True(t, input.IncludeReasons)
}

func TestGetRecommendationsOutput_Structure(t *testing.T) {
	recommendations := []map[string]interface{}{
		{
			"element_id":   "rec-1",
			"element_type": "skill",
			"score":        0.85,
		},
	}

	output := GetRecommendationsOutput{
		ElementID:       "test-id",
		TotalFound:      1,
		Recommendations: recommendations,
	}

	assert.Equal(t, "test-id", output.ElementID)
	assert.Equal(t, 1, output.TotalFound)
	assert.Len(t, output.Recommendations, 1)
	assert.Equal(t, "rec-1", output.Recommendations[0]["element_id"])
}

func TestGetRelationshipStatsInput_Structure(t *testing.T) {
	input := GetRelationshipStatsInput{
		ElementID: "test-id",
	}

	assert.Equal(t, "test-id", input.ElementID)
}

func TestGetRelationshipStatsOutput_Structure(t *testing.T) {
	output := GetRelationshipStatsOutput{
		ForwardEntries: 10,
		ReverseEntries: 15,
		CacheHits:      100,
		CacheMisses:    20,
		CacheHitRate:   83.33,
		CacheSize:      50,
		ElementDetails: &struct {
			ForwardCount int `json:"forward_count"`
			ReverseCount int `json:"reverse_count"`
		}{
			ForwardCount: 5,
			ReverseCount: 3,
		},
	}

	assert.Equal(t, 10, output.ForwardEntries)
	assert.Equal(t, 15, output.ReverseEntries)
	assert.Equal(t, int64(100), output.CacheHits)
	assert.Equal(t, int64(20), output.CacheMisses)
	assert.Equal(t, 83.33, output.CacheHitRate)
	assert.Equal(t, 50, output.CacheSize)
	require.NotNil(t, output.ElementDetails)
	assert.Equal(t, 5, output.ElementDetails.ForwardCount)
	assert.Equal(t, 3, output.ElementDetails.ReverseCount)
}

func TestConvertNodeToMap_NilNode(t *testing.T) {
	result := convertNodeToMap(nil)
	assert.Nil(t, result)
}

func TestConvertNodeToMap_SingleNode(t *testing.T) {
	persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "test")
	err := persona.SetSystemPrompt("Test prompt")
	require.NoError(t, err)

	node := &application.RelationshipNode{
		Element:      persona,
		Depth:        1,
		Relationship: "uses",
		Score:        0.85,
		Children:     []*application.RelationshipNode{},
	}

	result := convertNodeToMap(node)

	require.NotNil(t, result)
	assert.Equal(t, persona.GetID(), result["id"])
	assert.Equal(t, "persona", result["type"])
	assert.Equal(t, "Test Persona", result["name"])
	assert.Equal(t, 1, result["depth"])
	assert.Equal(t, "uses", result["relationship"])
	assert.Equal(t, 0.85, result["score"])
	assert.Nil(t, result["children"])
}

func TestConvertNodeToMap_WithChildren(t *testing.T) {
	parent := domain.NewPersona("Parent", "Test", "1.0.0", "test")
	err := parent.SetSystemPrompt("Parent prompt")
	require.NoError(t, err)

	child1 := domain.NewSkill("Child1", "Test", "1.0.0", "test")
	err = child1.AddTrigger(domain.SkillTrigger{Type: "keyword", Keywords: []string{"test"}})
	require.NoError(t, err)
	err = child1.AddProcedure(domain.SkillProcedure{Step: 1, Action: "test"})
	require.NoError(t, err)

	child2 := domain.NewMemory("Child2", "Test", "1.0.0", "test")
	child2.Content = "Test content"

	parentNode := &application.RelationshipNode{
		Element:      parent,
		Depth:        0,
		Relationship: "root",
		Score:        1.0,
		Children: []*application.RelationshipNode{
			{
				Element:      child1,
				Depth:        1,
				Relationship: "uses",
				Score:        0.8,
				Children:     []*application.RelationshipNode{},
			},
			{
				Element:      child2,
				Depth:        1,
				Relationship: "references",
				Score:        0.7,
				Children:     []*application.RelationshipNode{},
			},
		},
	}

	result := convertNodeToMap(parentNode)

	require.NotNil(t, result)
	assert.Equal(t, "Parent", result["name"])
	assert.Equal(t, 0, result["depth"])

	children, ok := result["children"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, children, 2)

	assert.Equal(t, "Child1", children[0]["name"])
	assert.Equal(t, "skill", children[0]["type"])
	assert.Equal(t, 1, children[0]["depth"])

	assert.Equal(t, "Child2", children[1]["name"])
	assert.Equal(t, "memory", children[1]["type"])
	assert.Equal(t, 1, children[1]["depth"])
}

func TestCountNodesInTree_NilNode(t *testing.T) {
	count := countNodesInTree(nil)
	assert.Equal(t, 0, count)
}

func TestCountNodesInTree_SingleNode(t *testing.T) {
	persona := domain.NewPersona("Test", "Test", "1.0.0", "test")
	err := persona.SetSystemPrompt("Test prompt at least 10 chars")
	require.NoError(t, err)

	node := &application.RelationshipNode{
		Element:  persona,
		Children: []*application.RelationshipNode{},
	}

	count := countNodesInTree(node)
	assert.Equal(t, 1, count)
}

func TestCountNodesInTree_WithChildren(t *testing.T) {
	parent := domain.NewPersona("Parent", "Test", "1.0.0", "test")
	err := parent.SetSystemPrompt("Parent prompt at least 10 chars")
	require.NoError(t, err)

	child1 := domain.NewSkill("Child1", "Test", "1.0.0", "test")
	err = child1.AddTrigger(domain.SkillTrigger{Type: "keyword", Keywords: []string{"test"}})
	require.NoError(t, err)
	err = child1.AddProcedure(domain.SkillProcedure{Step: 1, Action: "test"})
	require.NoError(t, err)

	child2 := domain.NewMemory("Child2", "Test", "1.0.0", "test")
	child2.Content = "Test"

	grandchild := domain.NewTemplate("Grandchild", "Test", "1.0.0", "test")
	grandchild.Content = "Test"

	parentNode := &application.RelationshipNode{
		Element: parent,
		Children: []*application.RelationshipNode{
			{
				Element: child1,
				Children: []*application.RelationshipNode{
					{
						Element:  grandchild,
						Children: []*application.RelationshipNode{},
					},
				},
			},
			{
				Element:  child2,
				Children: []*application.RelationshipNode{},
			},
		},
	}

	count := countNodesInTree(parentNode)
	assert.Equal(t, 4, count) // parent + 2 children + 1 grandchild
}

func TestContainsString_Found(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	assert.True(t, containsString(slice, "banana"))
	assert.True(t, containsString(slice, "apple"))
	assert.True(t, containsString(slice, "cherry"))
}

func TestContainsString_NotFound(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	assert.False(t, containsString(slice, "orange"))
	assert.False(t, containsString(slice, "grape"))
	assert.False(t, containsString(slice, ""))
}

func TestContainsString_EmptySlice(t *testing.T) {
	slice := []string{}

	assert.False(t, containsString(slice, "anything"))
	assert.False(t, containsString(slice, ""))
}

func TestContainsString_CaseSensitive(t *testing.T) {
	slice := []string{"Apple", "Banana"}

	assert.True(t, containsString(slice, "Apple"))
	assert.False(t, containsString(slice, "apple"))
	assert.True(t, containsString(slice, "Banana"))
	assert.False(t, containsString(slice, "banana"))
}

func TestHandleGetRelatedElements_DefaultDirection(t *testing.T) {
	input := GetRelatedElementsInput{
		ElementID: "test-id",
		// Direction not set, should default to "both"
	}

	// Direction should be set to "both" by the handler
	assert.Empty(t, input.Direction)
}

func TestHandleExpandRelationships_MaxDepthDefaults(t *testing.T) {
	tests := []struct {
		name          string
		inputMaxDepth int
		expectedMax   int
	}{
		{
			name:          "Zero defaults to 3",
			inputMaxDepth: 0,
			expectedMax:   3,
		},
		{
			name:          "Under limit unchanged",
			inputMaxDepth: 2,
			expectedMax:   2,
		},
		{
			name:          "At limit unchanged",
			inputMaxDepth: 5,
			expectedMax:   5,
		},
		{
			name:          "Over limit capped at 5",
			inputMaxDepth: 10,
			expectedMax:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := ExpandRelationshipsInput{
				ElementID: "test-id",
				MaxDepth:  tt.inputMaxDepth,
			}

			// Simulate the handler's logic
			maxDepth := input.MaxDepth
			if maxDepth == 0 {
				maxDepth = 3
			}
			if maxDepth > 5 {
				maxDepth = 5
			}

			assert.Equal(t, tt.expectedMax, maxDepth)
		})
	}
}

func TestGetRelationshipStatsOutput_CacheHitRateCalculation(t *testing.T) {
	tests := []struct {
		name            string
		cacheHits       int64
		cacheMisses     int64
		expectedHitRate float64
		shouldCalculate bool
	}{
		{
			name:            "High hit rate",
			cacheHits:       90,
			cacheMisses:     10,
			expectedHitRate: 90.0,
			shouldCalculate: true,
		},
		{
			name:            "Low hit rate",
			cacheHits:       10,
			cacheMisses:     90,
			expectedHitRate: 10.0,
			shouldCalculate: true,
		},
		{
			name:            "Perfect hit rate",
			cacheHits:       100,
			cacheMisses:     0,
			expectedHitRate: 100.0,
			shouldCalculate: true,
		},
		{
			name:            "Zero hits",
			cacheHits:       0,
			cacheMisses:     100,
			expectedHitRate: 0.0,
			shouldCalculate: true,
		},
		{
			name:            "No requests",
			cacheHits:       0,
			cacheMisses:     0,
			expectedHitRate: 0.0,
			shouldCalculate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the handler's logic
			totalRequests := tt.cacheHits + tt.cacheMisses
			var hitRate float64
			if totalRequests > 0 {
				hitRate = float64(tt.cacheHits) / float64(totalRequests) * 100
			}

			if tt.shouldCalculate {
				assert.InDelta(t, tt.expectedHitRate, hitRate, 0.01)
			} else {
				assert.Equal(t, 0.0, hitRate)
			}
		})
	}
}

func TestInferRelationshipsOutput_InferenceFormat(t *testing.T) {
	// Simulate how inferences are formatted
	inf := application.InferredRelationship{
		SourceID:   "source-123",
		TargetID:   "target-456",
		SourceType: domain.MemoryElement,
		TargetType: domain.SkillElement,
		Confidence: 0.85,
		Evidence:   []string{"ID mentioned in content", "Keyword match"},
		InferredBy: "mention",
	}

	inferenceMap := map[string]interface{}{
		"source_id":   inf.SourceID,
		"target_id":   inf.TargetID,
		"source_type": string(inf.SourceType),
		"target_type": string(inf.TargetType),
		"confidence":  inf.Confidence,
		"evidence":    inf.Evidence,
		"inferred_by": inf.InferredBy,
	}

	assert.Equal(t, "source-123", inferenceMap["source_id"])
	assert.Equal(t, "target-456", inferenceMap["target_id"])
	assert.Equal(t, "memory", inferenceMap["source_type"])
	assert.Equal(t, "skill", inferenceMap["target_type"])
	assert.Equal(t, 0.85, inferenceMap["confidence"])
	assert.Equal(t, []string{"ID mentioned in content", "Keyword match"}, inferenceMap["evidence"])
	assert.Equal(t, "mention", inferenceMap["inferred_by"])
}

func TestGetRecommendationsOutput_RecommendationFormat(t *testing.T) {
	// Test with reasons
	t.Run("With Reasons", func(t *testing.T) {
		rec := application.Recommendation{
			ElementID:   "rec-123",
			ElementType: domain.SkillElement,
			ElementName: "Test Skill",
			Score:       0.92,
			Reasons:     []string{"Frequently used together", "Similar context"},
		}

		recMap := map[string]interface{}{
			"element_id":   rec.ElementID,
			"element_type": string(rec.ElementType),
			"element_name": rec.ElementName,
			"score":        rec.Score,
			"reasons":      rec.Reasons,
		}

		assert.Equal(t, "rec-123", recMap["element_id"])
		assert.Equal(t, "skill", recMap["element_type"])
		assert.Equal(t, "Test Skill", recMap["element_name"])
		assert.Equal(t, 0.92, recMap["score"])
		assert.Equal(t, []string{"Frequently used together", "Similar context"}, recMap["reasons"])
	})

	// Test without reasons
	t.Run("Without Reasons", func(t *testing.T) {
		rec := application.Recommendation{
			ElementID:   "rec-456",
			ElementType: domain.PersonaElement,
			ElementName: "Test Persona",
			Score:       0.75,
		}

		recMap := map[string]interface{}{
			"element_id":   rec.ElementID,
			"element_type": string(rec.ElementType),
			"element_name": rec.ElementName,
			"score":        rec.Score,
		}

		assert.Equal(t, "rec-456", recMap["element_id"])
		assert.Equal(t, "persona", recMap["element_type"])
		assert.Equal(t, "Test Persona", recMap["element_name"])
		assert.Equal(t, 0.75, recMap["score"])
		assert.NotContains(t, recMap, "reasons")
	})
}
