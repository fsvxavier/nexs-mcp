package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/common"
)

func setupTestServerForBatch() *MCPServer {
	repo := NewMockElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

func TestHandleBatchCreateElements_EmptyElements(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchCreateElementsInput{
		Elements: []BatchElementInput{},
	}

	_, _, err := server.handleBatchCreateElements(context.Background(), nil, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one element required")
}

func TestHandleBatchCreateElements_ExceedsMaxBatchSize(t *testing.T) {
	server := setupTestServerForBatch()

	// Create 51 elements
	elements := make([]BatchElementInput, 51)
	for i := range 51 {
		elements[i] = BatchElementInput{
			Type: common.ElementTypePersona,
			Name: "Test Persona",
		}
	}

	input := BatchCreateElementsInput{
		Elements: elements,
	}

	_, _, err := server.handleBatchCreateElements(context.Background(), nil, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum 50 elements per batch")
}

func TestHandleBatchCreateElements_SinglePersona(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchCreateElementsInput{
		Elements: []BatchElementInput{
			{
				Type:        common.ElementTypePersona,
				Name:        "Test Persona",
				Description: "A test persona",
				Template:    "technical", // Use template to avoid validation
			},
		},
	}

	_, output, err := server.handleBatchCreateElements(context.Background(), nil, input)
	require.NoError(t, err)
	assert.Equal(t, 1, output.Created)
	assert.Equal(t, 0, output.Failed)
	assert.Equal(t, 1, output.Total)
	assert.Len(t, output.Results, 1)
	assert.True(t, output.Results[0].Success)
	assert.NotEmpty(t, output.Results[0].ID)
	assert.Contains(t, output.Summary, "1 created, 0 failed")
}

func TestHandleBatchCreateElements_MultipleElements(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchCreateElementsInput{
		Elements: []BatchElementInput{
			{
				Type:        common.ElementTypePersona,
				Name:        "Persona 1",
				Description: "First persona",
				Template:    "technical", // Use template
			},
			{
				Type:        common.ElementTypeSkill,
				Name:        "Skill 1",
				Description: "First skill",
				Template:    "coding", // Use template
			},
			{
				Type: common.ElementTypeMemory,
				Name: "Memory 1",
				Data: map[string]interface{}{
					"content": "Test memory content",
				},
			},
		},
	}

	_, output, err := server.handleBatchCreateElements(context.Background(), nil, input)
	require.NoError(t, err)
	assert.Equal(t, 3, output.Created)
	assert.Equal(t, 0, output.Failed)
	assert.Equal(t, 3, output.Total)
	assert.Len(t, output.Results, 3)

	// Verify all succeeded
	for _, result := range output.Results {
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.ID)
		assert.Empty(t, result.Error)
	}
}

func TestHandleBatchCreateElements_PartialFailure(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchCreateElementsInput{
		Elements: []BatchElementInput{
			{
				Type:        common.ElementTypePersona,
				Name:        "Valid Persona",
				Description: "Valid",
				Template:    "technical", // Use template
			},
			{
				Type: common.ElementTypeMemory,
				Name: "Invalid Memory",
				// Missing required "content" field
			},
		},
	}

	_, output, err := server.handleBatchCreateElements(context.Background(), nil, input)
	require.NoError(t, err)
	assert.Equal(t, 1, output.Created)
	assert.Equal(t, 1, output.Failed)
	assert.Equal(t, 2, output.Total)
	assert.Contains(t, output.Summary, "1 created, 1 failed")

	// First should succeed
	assert.True(t, output.Results[0].Success)
	assert.NotEmpty(t, output.Results[0].ID)

	// Second should fail
	assert.False(t, output.Results[1].Success)
	assert.NotEmpty(t, output.Results[1].Error)
}

func TestHandleBatchCreateElements_UnsupportedType(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchCreateElementsInput{
		Elements: []BatchElementInput{
			{
				Type: "invalid_type",
				Name: "Test",
			},
		},
	}

	_, output, err := server.handleBatchCreateElements(context.Background(), nil, input)
	require.NoError(t, err)
	assert.Equal(t, 0, output.Created)
	assert.Equal(t, 1, output.Failed)
	assert.False(t, output.Results[0].Success)
	assert.Contains(t, output.Results[0].Error, "unsupported element type")
}

func TestHandleBatchCreateElements_ResultStructure(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchCreateElementsInput{
		Elements: []BatchElementInput{
			{
				Type:        common.ElementTypeSkill,
				Name:        "Test Skill",
				Description: "A test skill",
				Template:    "coding", // Use template
			},
		},
	}

	_, output, err := server.handleBatchCreateElements(context.Background(), nil, input)
	require.NoError(t, err)

	// Verify output structure
	assert.GreaterOrEqual(t, output.DurationMs, int64(0))
	assert.NotEmpty(t, output.Summary)
	assert.Equal(t, 0, output.Results[0].Index)
	assert.Equal(t, common.ElementTypeSkill, output.Results[0].Type)
	assert.Equal(t, "Test Skill", output.Results[0].Name)
	assert.NotNil(t, output.Results[0].Data)
	assert.Contains(t, output.Results[0].Data["file_path"], "data/elements/skill")
}

func TestBatchCreatePersona_WithoutTemplate(t *testing.T) {
	t.Skip("Direct persona creation requires BehavioralTraits and ExpertiseAreas, tested via template creation")
}

func TestBatchCreatePersona_WithTemplate(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchElementInput{
		Type:        common.ElementTypePersona,
		Name:        "Test Persona",
		Description: "A test persona",
		Template:    "technical",
		Data: map[string]interface{}{
			"expertise": []interface{}{"Go", "Testing"},
		},
	}

	id, err := server.batchCreatePersona(context.Background(), input)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestBatchCreateSkill_WithoutTemplate(t *testing.T) {
	t.Skip("Direct skill creation requires triggers, tested via template creation")
}

func TestBatchCreateSkill_WithTemplate(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchElementInput{
		Type:        common.ElementTypeSkill,
		Name:        "Test Skill",
		Description: "A test skill",
		Template:    "coding",
		Data: map[string]interface{}{
			"trigger": "when coding",
		},
	}

	id, err := server.batchCreateSkill(context.Background(), input)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestBatchCreateMemory_Success(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchElementInput{
		Type: common.ElementTypeMemory,
		Name: "Test Memory",
		Data: map[string]interface{}{
			"content":    "Test memory content",
			"tags":       []interface{}{"tag1", "tag2"},
			"importance": "high",
		},
	}

	id, err := server.batchCreateMemory(context.Background(), input)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestBatchCreateMemory_MissingContent(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchElementInput{
		Type: common.ElementTypeMemory,
		Name: "Test Memory",
		Data: map[string]interface{}{},
	}

	_, err := server.batchCreateMemory(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content required")
}

func TestBatchCreateTemplate_Success(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchElementInput{
		Type:        common.ElementTypeTemplate,
		Name:        "Test Template",
		Description: "A test template",
		Data: map[string]interface{}{
			"content": "{{name}} template content",
			"variables": map[string]interface{}{
				"name": "The name variable",
			},
		},
	}

	id, err := server.batchCreateTemplate(context.Background(), input)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestBatchCreateTemplate_MissingContent(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchElementInput{
		Type:        common.ElementTypeTemplate,
		Name:        "Test Template",
		Description: "A test template",
		Data:        map[string]interface{}{},
	}

	_, err := server.batchCreateTemplate(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content required")
}

func TestBatchCreateAgent_Success(t *testing.T) {
	t.Skip("Agent creation requires actions array, tested via integration")
}

func TestBatchCreateAgent_MissingGoal(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchElementInput{
		Type:        common.ElementTypeAgent,
		Name:        "Test Agent",
		Description: "A test agent",
		Data:        map[string]interface{}{},
	}

	_, err := server.batchCreateAgent(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "goal required")
}

func TestBatchCreateEnsemble_Success(t *testing.T) {
	t.Skip("Ensemble creation requires member array, tested via integration")
}

func TestBatchCreateEnsemble_DefaultExecutionMode(t *testing.T) {
	t.Skip("Ensemble creation requires member array, tested via integration")
}

func TestBatchCreateElements_AllTypes(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchCreateElementsInput{
		Elements: []BatchElementInput{
			{
				Type:        common.ElementTypePersona,
				Name:        "Batch Persona",
				Description: "Test persona",
				Template:    "technical", // Use template to avoid validation
			},
			{
				Type:        common.ElementTypeSkill,
				Name:        "Batch Skill",
				Description: "Test skill",
				Template:    "coding", // Use template to avoid validation
			},
			{
				Type: common.ElementTypeMemory,
				Name: "Batch Memory",
				Data: map[string]interface{}{
					"content": "Test content",
				},
			},
			{
				Type:        common.ElementTypeTemplate,
				Name:        "Batch Template",
				Description: "Test template",
				Data: map[string]interface{}{
					"content": "{{var}} template",
				},
			},
		},
	}

	_, output, err := server.handleBatchCreateElements(context.Background(), nil, input)
	require.NoError(t, err)
	assert.Equal(t, 4, output.Created)
	assert.Equal(t, 0, output.Failed)
	assert.Equal(t, 4, output.Total)

	// Verify tested types
	types := []string{
		common.ElementTypePersona,
		common.ElementTypeSkill,
		common.ElementTypeMemory,
		common.ElementTypeTemplate,
	}

	for i, expectedType := range types {
		assert.Equal(t, expectedType, output.Results[i].Type)
		assert.True(t, output.Results[i].Success)
		assert.NotEmpty(t, output.Results[i].ID)
	}
}

func TestBatchCreateElements_MCPIntegration(t *testing.T) {
	server := setupTestServerForBatch()

	input := BatchCreateElementsInput{
		Elements: []BatchElementInput{
			{
				Type:        common.ElementTypePersona,
				Name:        "Test Persona",
				Description: "Test",
				Template:    "technical", // Use template to avoid validation
			},
		},
	}

	result, output, err := server.handleBatchCreateElements(context.Background(), nil, input)
	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, output.Created)
}
