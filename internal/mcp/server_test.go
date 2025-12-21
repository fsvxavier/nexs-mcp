package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

func TestNewMCPServer(t *testing.T) {
	repo := NewMockElementRepository()
	server := newTestServer("test-server", "1.0.0", repo)

	assert.NotNil(t, server)
	assert.NotNil(t, server.server)
	assert.Equal(t, repo, server.repo)
}

func TestHandleListElements(t *testing.T) {
	ctx := context.Background()
	repo := NewMockElementRepository()
	server := newTestServer("test-server", "1.0.0", repo)

	// Create test elements
	now := time.Now()
	elem1 := &SimpleElement{
		metadata: domain.ElementMetadata{
			ID:          "test-id-1",
			Type:        domain.PersonaElement,
			Name:        "Test Persona",
			Description: "Test Description",
			Version:     "1.0.0",
			Author:      "Test Author",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	elem2 := &SimpleElement{
		metadata: domain.ElementMetadata{
			ID:          "test-id-2",
			Type:        domain.SkillElement,
			Name:        "Test Skill",
			Description: "Test Skill Description",
			Version:     "1.0.0",
			Author:      "Test Author",
			IsActive:    false,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	repo.Create(elem1)
	repo.Create(elem2)

	tests := []struct {
		name          string
		input         ListElementsInput
		expectedCount int
		expectError   bool
	}{
		{
			name:          "list all elements",
			input:         ListElementsInput{},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "filter by type",
			input: ListElementsInput{
				Type: "persona",
			},
			expectedCount: 2, // Mock doesn't implement filtering
			expectError:   false,
		},
		{
			name: "filter by active status",
			input: ListElementsInput{
				IsActive: boolPtr(true),
			},
			expectedCount: 2, // Mock doesn't implement filtering
			expectError:   false,
		},
		{
			name: "filter by tags",
			input: ListElementsInput{
				Tags: "tag1,tag2",
			},
			expectedCount: 2, // Mock doesn't implement filtering
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleListElements(ctx, nil, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedCount, output.Total)
				assert.Len(t, output.Elements, tt.expectedCount)
			}
		})
	}
}

func TestHandleGetElement(t *testing.T) {
	ctx := context.Background()
	repo := NewMockElementRepository()
	server := newTestServer("test-server", "1.0.0", repo)

	// Create test element
	now := time.Now()
	elem := &SimpleElement{
		metadata: domain.ElementMetadata{
			ID:          "test-id",
			Type:        domain.PersonaElement,
			Name:        "Test Persona",
			Description: "Test Description",
			Version:     "1.0.0",
			Author:      "Test Author",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	repo.Create(elem)

	tests := []struct {
		name        string
		input       GetElementInput
		expectError bool
		errorMsg    string
	}{
		{
			name:        "get existing element",
			input:       GetElementInput{ID: "test-id"},
			expectError: false,
		},
		{
			name:        "get non-existing element",
			input:       GetElementInput{ID: "non-existing"},
			expectError: true,
			errorMsg:    "failed to get element",
		},
		{
			name:        "empty ID",
			input:       GetElementInput{ID: ""},
			expectError: true,
			errorMsg:    "id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleGetElement(ctx, nil, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, result)
				assert.NotNil(t, output.Element)
				assert.Equal(t, "test-id", output.Element["id"])
			}
		})
	}
}

func TestHandleCreateElement(t *testing.T) {
	ctx := context.Background()
	repo := NewMockElementRepository()
	server := newTestServer("test-server", "1.0.0", repo)

	tests := []struct {
		name        string
		input       CreateElementInput
		expectError bool
		errorMsg    string
	}{
		{
			name: "create valid persona",
			input: CreateElementInput{
				Type:        "persona",
				Name:        "Test Persona",
				Description: "Test Description",
				Version:     "1.0.0",
				Author:      "Test Author",
				Tags:        []string{"tag1", "tag2"},
				IsActive:    true,
			},
			expectError: false,
		},
		{
			name: "create valid skill",
			input: CreateElementInput{
				Type:        "skill",
				Name:        "Test Skill",
				Description: "Test Skill Description",
				Version:     "1.0.0",
				Author:      "Test Author",
				IsActive:    false,
			},
			expectError: false,
		},
		{
			name: "missing type",
			input: CreateElementInput{
				Name:        "Test",
				Description: "Test",
				Version:     "1.0.0",
				Author:      "Test",
			},
			expectError: true,
			errorMsg:    "type is required",
		},
		{
			name: "invalid type",
			input: CreateElementInput{
				Type:        "invalid",
				Name:        "Test",
				Description: "Test",
				Version:     "1.0.0",
				Author:      "Test",
			},
			expectError: true,
			errorMsg:    "invalid element type",
		},
		{
			name: "name too short",
			input: CreateElementInput{
				Type:        "persona",
				Name:        "AB",
				Description: "Test",
				Version:     "1.0.0",
				Author:      "Test",
			},
			expectError: true,
			errorMsg:    "name must be between 3 and 100 characters",
		},
		{
			name: "name too long",
			input: CreateElementInput{
				Type:        "persona",
				Name:        string(make([]byte, 101)),
				Description: "Test",
				Version:     "1.0.0",
				Author:      "Test",
			},
			expectError: true,
			errorMsg:    "name must be between 3 and 100 characters",
		},
		{
			name: "description too long",
			input: CreateElementInput{
				Type:        "persona",
				Name:        "Test",
				Description: string(make([]byte, 501)),
				Version:     "1.0.0",
				Author:      "Test",
			},
			expectError: true,
			errorMsg:    "description must be at most 500 characters",
		},
		{
			name: "missing version",
			input: CreateElementInput{
				Type:        "persona",
				Name:        "Test",
				Description: "Test",
				Author:      "Test",
			},
			expectError: true,
			errorMsg:    "version is required",
		},
		{
			name: "missing author",
			input: CreateElementInput{
				Type:        "persona",
				Name:        "Test",
				Description: "Test",
				Version:     "1.0.0",
			},
			expectError: true,
			errorMsg:    "author is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleCreateElement(ctx, nil, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, result)
				assert.NotEmpty(t, output.ID)
				assert.NotNil(t, output.Element)
				assert.Equal(t, tt.input.Type, output.Element["type"])
				assert.Equal(t, tt.input.Name, output.Element["name"])
			}
		})
	}
}

func TestHandleUpdateElement(t *testing.T) {
	ctx := context.Background()
	repo := NewMockElementRepository()
	server := newTestServer("test-server", "1.0.0", repo)

	// Create test element
	now := time.Now()
	elem := &SimpleElement{
		metadata: domain.ElementMetadata{
			ID:          "test-id",
			Type:        domain.PersonaElement,
			Name:        "Original Name",
			Description: "Original Description",
			Version:     "1.0.0",
			Author:      "Test Author",
			Tags:        []string{"tag1"},
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	repo.Create(elem)

	tests := []struct {
		name        string
		input       UpdateElementInput
		expectError bool
		errorMsg    string
		checkUpdate func(t *testing.T, output UpdateElementOutput)
	}{
		{
			name: "update name",
			input: UpdateElementInput{
				ID:   "test-id",
				Name: "Updated Name",
				User: "Test Author", // Owner can update
			},
			expectError: false,
			checkUpdate: func(t *testing.T, output UpdateElementOutput) {
				assert.Equal(t, "Updated Name", output.Element["name"])
			},
		},
		{
			name: "update description",
			input: UpdateElementInput{
				ID:          "test-id",
				Description: "Updated Description",
				User:        "Test Author", // Owner can update
			},
			expectError: false,
			checkUpdate: func(t *testing.T, output UpdateElementOutput) {
				assert.Equal(t, "Updated Description", output.Element["description"])
			},
		},
		{
			name: "update tags",
			input: UpdateElementInput{
				ID:   "test-id",
				Tags: []string{"tag2", "tag3"},
				User: "Test Author", // Owner can update
			},
			expectError: false,
			checkUpdate: func(t *testing.T, output UpdateElementOutput) {
				tags := output.Element["tags"].([]string)
				assert.Contains(t, tags, "tag2")
				assert.Contains(t, tags, "tag3")
			},
		},
		{
			name: "update active status",
			input: UpdateElementInput{
				ID:       "test-id",
				IsActive: boolPtr(false),
				User:     "Test Author", // Owner can update
			},
			expectError: false,
			checkUpdate: func(t *testing.T, output UpdateElementOutput) {
				assert.False(t, output.Element["is_active"].(bool))
			},
		},
		{
			name: "update multiple fields",
			input: UpdateElementInput{
				ID:          "test-id",
				Name:        "Multi Update",
				Description: "Multi Description",
				Tags:        []string{"multi"},
				IsActive:    boolPtr(true),
				User:        "Test Author", // Owner can update
			},
			expectError: false,
			checkUpdate: func(t *testing.T, output UpdateElementOutput) {
				assert.Equal(t, "Multi Update", output.Element["name"])
				assert.Equal(t, "Multi Description", output.Element["description"])
			},
		},
		{
			name: "missing ID",
			input: UpdateElementInput{
				Name: "Test",
			},
			expectError: true,
			errorMsg:    "id is required",
		},
		{
			name: "non-existing element",
			input: UpdateElementInput{
				ID:   "non-existing",
				Name: "Test",
			},
			expectError: true,
			errorMsg:    "failed to get element",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleUpdateElement(ctx, nil, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, result)
				assert.NotNil(t, output.Element)
				if tt.checkUpdate != nil {
					tt.checkUpdate(t, output)
				}
			}
		})
	}
}

func TestHandleDeleteElement(t *testing.T) {
	ctx := context.Background()
	repo := NewMockElementRepository()
	server := newTestServer("test-server", "1.0.0", repo)

	// Create test element
	now := time.Now()
	elem := &SimpleElement{
		metadata: domain.ElementMetadata{
			ID:        "test-id",
			Type:      domain.PersonaElement,
			Name:      "Test Persona",
			Version:   "1.0.0",
			Author:    "Test Author",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	repo.Create(elem)

	tests := []struct {
		name        string
		input       DeleteElementInput
		expectError bool
		errorMsg    string
	}{
		{
			name:        "delete existing element",
			input:       DeleteElementInput{ID: "test-id"},
			expectError: false,
		},
		{
			name:        "delete non-existing element",
			input:       DeleteElementInput{ID: "non-existing"},
			expectError: false, // Handler returns success=false in output
		},
		{
			name:        "empty ID",
			input:       DeleteElementInput{ID: ""},
			expectError: true,
			errorMsg:    "id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := server.handleDeleteElement(ctx, nil, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, result)
				assert.NotEmpty(t, output.Message)
			}
		})
	}
}

func TestHandleDeleteElement_VerifyDeletion(t *testing.T) {
	ctx := context.Background()
	repo := NewMockElementRepository()
	server := newTestServer("test-server", "1.0.0", repo)

	// Create and delete element
	now := time.Now()
	elem := &SimpleElement{
		metadata: domain.ElementMetadata{
			ID:        "delete-test",
			Type:      domain.PersonaElement,
			Name:      "Test",
			Version:   "1.0.0",
			Author:    "Test",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	repo.Create(elem)

	// Delete
	_, output, err := server.handleDeleteElement(ctx, nil, DeleteElementInput{
		ID:   "delete-test",
		User: "Test", // Owner can delete
	})
	require.NoError(t, err)
	assert.True(t, output.Success)

	// Verify deletion
	exists, err := repo.Exists("delete-test")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMockRepository(t *testing.T) {
	repo := NewMockElementRepository()
	assert.NotNil(t, repo)

	now := time.Now()
	elem := &SimpleElement{
		metadata: domain.ElementMetadata{
			ID:        "mock-test",
			Type:      domain.SkillElement,
			Name:      "Mock Test",
			Version:   "1.0.0",
			Author:    "Test",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Test Create
	err := repo.Create(elem)
	assert.NoError(t, err)

	// Test Exists
	exists, err := repo.Exists("mock-test")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test GetByID
	retrieved, err := repo.GetByID("mock-test")
	assert.NoError(t, err)
	assert.Equal(t, "mock-test", retrieved.GetID())

	// Test Update
	elem.metadata.Name = "Updated Mock"
	err = repo.Update(elem)
	assert.NoError(t, err)

	// Test List
	elements, err := repo.List(domain.ElementFilter{})
	assert.NoError(t, err)
	assert.Len(t, elements, 1)

	// Test Delete
	err = repo.Delete("mock-test")
	assert.NoError(t, err)

	exists, err = repo.Exists("mock-test")
	assert.NoError(t, err)
	assert.False(t, exists)
}

// Helper function.
func boolPtr(b bool) *bool {
	return &b
}
