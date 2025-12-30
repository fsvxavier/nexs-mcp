package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a temporary test directory.
func createTestDir(t *testing.T) string {
	dir := t.TempDir()
	return dir
}

// Helper to create a test element.
func createTestElement(t *testing.T, elementType domain.ElementType, name string) domain.Element {
	id := domain.GenerateElementID(elementType, name)
	now := time.Now().Truncate(time.Second)

	return &testElement{
		metadata: domain.ElementMetadata{
			ID:          id,
			Type:        elementType,
			Name:        name,
			Description: "Test description",
			Version:     "1.0.0",
			Author:      "Test Author",
			Tags:        []string{"test"},
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

type testElement struct {
	metadata domain.ElementMetadata
}

func (e *testElement) GetMetadata() domain.ElementMetadata { return e.metadata }
func (e *testElement) Validate() error                     { return nil }
func (e *testElement) GetType() domain.ElementType         { return e.metadata.Type }
func (e *testElement) GetID() string                       { return e.metadata.ID }
func (e *testElement) IsActive() bool                      { return e.metadata.IsActive }
func (e *testElement) Activate() error {
	e.metadata.IsActive = true
	return nil
}

func (e *testElement) Deactivate() error {
	e.metadata.IsActive = false
	return nil
}

func TestNewFileElementRepository(t *testing.T) {
	t.Run("create with default directory", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)

		require.NoError(t, err)
		assert.NotNil(t, repo)
		assert.Equal(t, dir, repo.baseDir)
		assert.NotNil(t, repo.cache)

		// Verify directory was created
		_, err = os.Stat(dir)
		assert.NoError(t, err)
	})

	t.Run("create with empty directory", func(t *testing.T) {
		dir := createTestDir(t)
		subdir := filepath.Join(dir, "custom")
		repo, err := NewFileElementRepository(subdir)

		require.NoError(t, err)
		assert.Equal(t, subdir, repo.baseDir)

		// Verify directory was created
		_, err = os.Stat(subdir)
		assert.NoError(t, err)
	})
}

func TestFileElementRepository_Create(t *testing.T) {
	t.Run("create new element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		element := createTestElement(t, domain.PersonaElement, "Test Persona")
		err = repo.Create(element)

		require.NoError(t, err)

		// Verify element is in cache
		exists, err := repo.Exists(element.GetID())
		require.NoError(t, err)
		assert.True(t, exists)

		// Verify file was created
		metadata := element.GetMetadata()
		filePath := repo.getFilePath(metadata)
		_, err = os.Stat(filePath)
		assert.NoError(t, err)
	})

	t.Run("create duplicate element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		element := createTestElement(t, domain.SkillElement, "Test Skill")
		err = repo.Create(element)
		require.NoError(t, err)

		// Try to create again
		err = repo.Create(element)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("create nil element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		err = repo.Create(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestFileElementRepository_GetByID(t *testing.T) {
	t.Run("get existing element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		original := createTestElement(t, domain.TemplateElement, "Test Template")
		err = repo.Create(original)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(original.GetID())

		require.NoError(t, err)
		assert.Equal(t, original.GetID(), retrieved.GetID())
		assert.Equal(t, original.GetType(), retrieved.GetType())
		assert.Equal(t, original.GetMetadata().Name, retrieved.GetMetadata().Name)
	})

	t.Run("get non-existing element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		_, err = repo.GetByID("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestFileElementRepository_Update(t *testing.T) {
	t.Run("update existing element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		element := createTestElement(t, domain.AgentElement, "Test Agent")
		err = repo.Create(element)
		require.NoError(t, err)

		// Update element
		metadata := element.GetMetadata()
		metadata.Name = "Updated Agent"
		metadata.UpdatedAt = time.Now()
		updated := &testElement{metadata: metadata}

		err = repo.Update(updated)

		require.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetByID(element.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Updated Agent", retrieved.GetMetadata().Name)
	})

	t.Run("update non-existing element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		element := createTestElement(t, domain.MemoryElement, "Test Memory")
		err = repo.Update(element)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("update nil element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		err = repo.Update(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestFileElementRepository_Delete(t *testing.T) {
	t.Run("delete existing element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		element := createTestElement(t, domain.EnsembleElement, "Test Ensemble")
		err = repo.Create(element)
		require.NoError(t, err)

		// Get file path before deletion
		filePath := repo.getFilePath(element.GetMetadata())

		err = repo.Delete(element.GetID())

		require.NoError(t, err)

		// Verify element is removed from cache
		exists, err := repo.Exists(element.GetID())
		require.NoError(t, err)
		assert.False(t, exists)

		// Verify file is deleted
		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("delete non-existing element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		err = repo.Delete("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestFileElementRepository_List(t *testing.T) {
	t.Run("list all elements", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		// Create test elements
		persona := createTestElement(t, domain.PersonaElement, "Persona 1")
		skill := createTestElement(t, domain.SkillElement, "Skill 1")

		repo.Create(persona)
		repo.Create(skill)

		elements, err := repo.List(domain.ElementFilter{})

		require.NoError(t, err)
		assert.Len(t, elements, 2)
	})

	t.Run("filter by type", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		persona := createTestElement(t, domain.PersonaElement, "Persona 1")
		skill := createTestElement(t, domain.SkillElement, "Skill 1")

		repo.Create(persona)
		repo.Create(skill)

		personaType := domain.PersonaElement
		elements, err := repo.List(domain.ElementFilter{Type: &personaType})

		require.NoError(t, err)
		assert.Len(t, elements, 1)
		assert.Equal(t, domain.PersonaElement, elements[0].GetType())
	})

	t.Run("filter by active status", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		active := createTestElement(t, domain.TemplateElement, "Active Template")
		inactive := createTestElement(t, domain.TemplateElement, "Inactive Template")
		inactive.(*testElement).metadata.IsActive = false

		repo.Create(active)
		repo.Create(inactive)

		isActive := true
		elements, err := repo.List(domain.ElementFilter{IsActive: &isActive})

		require.NoError(t, err)
		assert.Len(t, elements, 1)
		assert.True(t, elements[0].IsActive())
	})

	t.Run("filter by tags", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		element1 := createTestElement(t, domain.AgentElement, "Agent 1")
		element2 := createTestElement(t, domain.AgentElement, "Agent 2")
		element2.(*testElement).metadata.Tags = []string{"special"}

		repo.Create(element1)
		repo.Create(element2)

		elements, err := repo.List(domain.ElementFilter{Tags: []string{"special"}})

		require.NoError(t, err)
		assert.Len(t, elements, 1)
	})

	t.Run("pagination", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		for i := 1; i <= 5; i++ {
			element := createTestElement(t, domain.MemoryElement, "Memory "+string(rune('0'+i)))
			repo.Create(element)
		}

		elements, err := repo.List(domain.ElementFilter{Limit: 2, Offset: 1})

		require.NoError(t, err)
		assert.Len(t, elements, 2)
	})
}

func TestFileElementRepository_Exists(t *testing.T) {
	t.Run("existing element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		element := createTestElement(t, domain.PersonaElement, "Test Persona")
		repo.Create(element)

		exists, err := repo.Exists(element.GetID())

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existing element", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		exists, err := repo.Exists("nonexistent")

		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestFileElementRepository_Persistence(t *testing.T) {
	t.Run("reload cache from disk", func(t *testing.T) {
		dir := createTestDir(t)

		// Create repository and add elements
		repo1, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		element1 := createTestElement(t, domain.PersonaElement, "Persona 1")
		element2 := createTestElement(t, domain.SkillElement, "Skill 1")
		repo1.Create(element1)
		repo1.Create(element2)

		// Create new repository instance (simulates restart)
		repo2, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		// Verify elements are loaded from disk
		elements, err := repo2.List(domain.ElementFilter{})
		require.NoError(t, err)
		assert.Len(t, elements, 2)

		exists1, _ := repo2.Exists(element1.GetID())
		exists2, _ := repo2.Exists(element2.GetID())
		assert.True(t, exists1)
		assert.True(t, exists2)
	})
}

func TestFileElementRepository_FileStructure(t *testing.T) {
	t.Run("correct file path structure", func(t *testing.T) {
		dir := createTestDir(t)
		repo, err := NewFileElementRepository(dir)
		require.NoError(t, err)

		element := createTestElement(t, domain.TemplateElement, "Test Template")
		err = repo.Create(element)
		require.NoError(t, err)

		metadata := element.GetMetadata()
		expectedDate := metadata.CreatedAt.Format("2006-01-02")
		// Directory structure is type/date/, not date/type/
		actualPath := repo.getFilePath(metadata)

		// Expect filename to be sanitized snake_case
		expectedFilename := sanitizeFileName(metadata.ID) + ".yaml"
		expectedPath := filepath.Join(dir, "template", expectedDate, expectedFilename)

		assert.Equal(t, expectedPath, actualPath)

		// Verify file exists at expected path
		_, err = os.Stat(expectedPath)
		assert.NoError(t, err)
	})
}

func TestFileElementRepository_MigrateIDs(t *testing.T) {
	dir := createTestDir(t)

	// Create an old-style file with unsanitized ID
	oldID := "persona_Test Persona_1234567890"
	now := time.Now().Truncate(time.Second)
	stored := &StoredElement{
		Metadata: domain.ElementMetadata{
			ID:        oldID,
			Type:      domain.PersonaElement,
			Name:      "Test Persona",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Data: map[string]interface{}{
			"related_skills": []interface{}{},
		},
	}

	// Write file to old path
	dateDir := now.Format("2006-01-02")
	oldPath := filepath.Join(dir, string(domain.PersonaElement), dateDir)
	require.NoError(t, os.MkdirAll(oldPath, 0o755))
	oldFile := filepath.Join(oldPath, oldID+".yaml")
	b, err := yaml.Marshal(stored)
	require.NoError(t, err)

	// Sanity check: ensure marshaled YAML contains metadata name
	t.Logf("marshaled YAML:\n%s", string(b))
	require.Contains(t, string(b), "Test Persona")

	require.NoError(t, os.WriteFile(oldFile, b, 0o644))

	// Initialize repository which should migrate the ID
	repo, err := NewFileElementRepository(dir)
	require.NoError(t, err)

	// New ID should be composed of type + sanitized name + timestamp
	elements, err := repo.List(domain.ElementFilter{})
	require.NoError(t, err)
	require.Len(t, elements, 1)
	loadedEl := elements[0]
	t.Logf("loaded ID: %s", loadedEl.GetID())
	assert.Contains(t, loadedEl.GetID(), "test_persona")

	// Old file should be removed
	_, err = os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err))

	// New file should exist and contain updated metadata.ID
	newPath := repo.getFilePath(domain.ElementMetadata{ID: loadedEl.GetID(), Type: domain.PersonaElement, CreatedAt: now})
	// Debug: list files under repo base
	var files []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	t.Logf("files on disk after migration: %v", files)

	data, err := os.ReadFile(newPath)
	require.NoError(t, err)
	var loaded StoredElement
	require.NoError(t, yaml.Unmarshal(data, &loaded))
	assert.Contains(t, loaded.Metadata.ID, "test_persona")
}
