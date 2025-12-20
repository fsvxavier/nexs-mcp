package infrastructure

import (
	"path/filepath"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLRUCache(t *testing.T) {
	t.Run("Put and Get", func(t *testing.T) {
		cache := NewLRUCache(2)

		elem1 := &StoredElement{Metadata: domain.ElementMetadata{ID: "1", Name: "Element 1"}}
		elem2 := &StoredElement{Metadata: domain.ElementMetadata{ID: "2", Name: "Element 2"}}

		cache.Put("1", elem1)
		cache.Put("2", elem2)

		got, found := cache.Get("1")
		assert.True(t, found)
		assert.Equal(t, "Element 1", got.Metadata.Name)
	})

	t.Run("Eviction", func(t *testing.T) {
		cache := NewLRUCache(2)

		elem1 := &StoredElement{Metadata: domain.ElementMetadata{ID: "1", Name: "Element 1"}}
		elem2 := &StoredElement{Metadata: domain.ElementMetadata{ID: "2", Name: "Element 2"}}
		elem3 := &StoredElement{Metadata: domain.ElementMetadata{ID: "3", Name: "Element 3"}}

		cache.Put("1", elem1)
		cache.Put("2", elem2)
		cache.Put("3", elem3) // Should evict elem1

		_, found := cache.Get("1")
		assert.False(t, found, "Element 1 should have been evicted")

		got, found := cache.Get("2")
		assert.True(t, found)
		assert.Equal(t, "Element 2", got.Metadata.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		cache := NewLRUCache(2)

		elem1 := &StoredElement{Metadata: domain.ElementMetadata{ID: "1", Name: "Element 1"}}
		cache.Put("1", elem1)

		cache.Delete("1")

		_, found := cache.Get("1")
		assert.False(t, found)
	})

	t.Run("Clear", func(t *testing.T) {
		cache := NewLRUCache(5)

		for i := 1; i <= 3; i++ {
			elem := &StoredElement{Metadata: domain.ElementMetadata{ID: string(rune(i)), Name: "Element"}}
			cache.Put(string(rune(i)), elem)
		}

		cache.Clear()

		for i := 1; i <= 3; i++ {
			_, found := cache.Get(string(rune(i)))
			assert.False(t, found)
		}
	})
}

func TestSearchIndex(t *testing.T) {
	t.Run("Index and Search", func(t *testing.T) {
		index := NewSearchIndex()

		persona := domain.NewPersona("Tech Expert", "Expert in technology", "1.0.0", "test")
		skill := domain.NewSkill("Code Review", "Review code quality", "1.0.0", "test")

		index.Index(persona)
		index.Index(skill)

		// Search for "tech"
		results := index.Search("tech")
		assert.Contains(t, results, persona.GetID())
		assert.NotContains(t, results, skill.GetID())

		// Search for "code"
		results = index.Search("code")
		assert.Contains(t, results, skill.GetID())
		assert.NotContains(t, results, persona.GetID())

		// Search for "review"
		results = index.Search("review")
		assert.Contains(t, results, skill.GetID())
	})

	t.Run("Remove from Index", func(t *testing.T) {
		index := NewSearchIndex()

		persona := domain.NewPersona("Tech Expert", "Expert in technology", "1.0.0", "test")
		index.Index(persona)

		results := index.Search("tech")
		assert.Contains(t, results, persona.GetID())

		index.Remove(persona.GetID())

		results = index.Search("tech")
		assert.NotContains(t, results, persona.GetID())
	})

	t.Run("Multi-word Search", func(t *testing.T) {
		index := NewSearchIndex()

		persona := domain.NewPersona("Senior Engineer", "Experienced software engineer", "1.0.0", "test")
		index.Index(persona)

		results := index.Search("software engineer")
		assert.Contains(t, results, persona.GetID())
	})
}

func TestEnhancedFileElementRepository(t *testing.T) {
	// Create temporary directory for tests
	tmpDir := t.TempDir()

	t.Run("Create and GetByID with LRU cache", func(t *testing.T) {
		repo, err := NewEnhancedFileElementRepository(filepath.Join(tmpDir, "create-get"), 10)
		require.NoError(t, err)

		persona := domain.NewPersona("Test Persona", "Test description", "1.0.0", "testuser")

		err = repo.Create(persona)
		require.NoError(t, err)

		// First get - should load into cache
		retrieved, err := repo.GetByID(persona.GetID())
		require.NoError(t, err)
		assert.Equal(t, persona.GetID(), retrieved.GetID())
		assert.Equal(t, domain.PersonaElement, retrieved.GetType())

		// Second get - should hit cache
		retrieved2, err := repo.GetByID(persona.GetID())
		require.NoError(t, err)
		assert.Equal(t, persona.GetID(), retrieved2.GetID())
	})

	t.Run("Directory structure", func(t *testing.T) {
		repo, err := NewEnhancedFileElementRepository(filepath.Join(tmpDir, "structure"), 10)
		require.NoError(t, err)

		persona := domain.NewPersona("Test Persona", "Test", "1.0.0", "testuser")
		require.NoError(t, repo.Create(persona))

		// Verify path structure: baseDir/type/date/filename.yaml
		metadata := persona.GetMetadata()
		expectedDate := metadata.CreatedAt.Format("2006-01-02")
		path := repo.getFilePath(metadata)

		assert.Contains(t, path, "persona", "Path should contain type directory")
		assert.Contains(t, path, expectedDate, "Path should contain date directory")
		assert.Contains(t, path, metadata.ID+".yaml", "Path should contain ID-based filename")

		// Verify structure: type comes before date
		personaIdx := filepath.Dir(filepath.Dir(path))
		assert.Contains(t, personaIdx, "structure", "Base directory should be present")
	})

	t.Run("Atomic updates", func(t *testing.T) {
		repo, err := NewEnhancedFileElementRepository(filepath.Join(tmpDir, "atomic"), 10)
		require.NoError(t, err)

		skill := domain.NewSkill("Original Skill", "Original", "1.0.0", "test")
		require.NoError(t, repo.Create(skill))

		// Create updated skill
		updatedSkill := domain.NewSkill("Updated Skill", "Updated description", "1.0.0", "test")
		metadata := skill.GetMetadata()
		metadata.Description = "Updated description"
		updatedSkill.SetMetadata(metadata)

		require.NoError(t, repo.Update(updatedSkill))

		// Verify update
		retrieved, err := repo.GetByID(skill.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Updated description", retrieved.GetMetadata().Description)

		// Verify no temp files left
		files, _ := filepath.Glob(filepath.Join(tmpDir, "atomic/**/*.tmp.*"))
		assert.Empty(t, files, "No temp files should remain")
	})

	t.Run("Full-text Search", func(t *testing.T) {
		repo, err := NewEnhancedFileElementRepository(filepath.Join(tmpDir, "search"), 10)
		require.NoError(t, err)

		persona1 := domain.NewPersona("AI Expert", "Expert in artificial intelligence", "1.0.0", "test")
		persona2 := domain.NewPersona("Data Scientist", "Expert in data science", "1.0.0", "test")
		skill := domain.NewSkill("Machine Learning", "ML algorithms", "1.0.0", "test")

		require.NoError(t, repo.Create(persona1))
		require.NoError(t, repo.Create(persona2))
		require.NoError(t, repo.Create(skill))

		// Search for "expert"
		results, err := repo.Search("expert", domain.ElementFilter{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)

		// Search for "machine learning"
		results, err = repo.Search("machine learning", domain.ElementFilter{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		// Search with type filter
		personaType := domain.PersonaElement
		results, err = repo.Search("expert", domain.ElementFilter{Type: &personaType})
		require.NoError(t, err)
		for _, r := range results {
			assert.Equal(t, domain.PersonaElement, r.GetType())
		}
	})

	t.Run("Backup and Restore", func(t *testing.T) {
		dataDir := filepath.Join(tmpDir, "backup-test")
		backupDir := filepath.Join(tmpDir, "backups")

		repo, err := NewEnhancedFileElementRepository(dataDir, 10)
		require.NoError(t, err)

		// Create some elements
		persona := domain.NewPersona("Backup Test", "Test persona", "1.0.0", "test")
		skill := domain.NewSkill("Backup Skill", "Test skill", "1.0.0", "test")

		require.NoError(t, repo.Create(persona))
		require.NoError(t, repo.Create(skill))

		// Create backup
		require.NoError(t, repo.Backup(backupDir))

		// Verify backup exists
		backups, err := filepath.Glob(filepath.Join(backupDir, "backup-*"))
		require.NoError(t, err)
		assert.NotEmpty(t, backups)

		// Delete an element
		require.NoError(t, repo.Delete(skill.GetID()))

		// Verify deleted
		_, err = repo.GetByID(skill.GetID())
		assert.Error(t, err)

		// Restore from backup
		require.NoError(t, repo.Restore(backups[0]))

		// Verify restored
		retrieved, err := repo.GetByID(skill.GetID())
		require.NoError(t, err)
		assert.Equal(t, skill.GetID(), retrieved.GetID())
	})

	t.Run("Delete removes from cache and index", func(t *testing.T) {
		repo, err := NewEnhancedFileElementRepository(filepath.Join(tmpDir, "delete"), 10)
		require.NoError(t, err)

		template := domain.NewTemplate("Test Template", "Test", "1.0.0", "test")
		require.NoError(t, repo.Create(template))

		// Verify in cache
		_, found := repo.lruCache.Get(template.GetID())
		assert.True(t, found)

		// Delete
		require.NoError(t, repo.Delete(template.GetID()))

		// Verify removed from cache
		_, found = repo.lruCache.Get(template.GetID())
		assert.False(t, found)

		// Verify removed from index
		_, exists := repo.index[template.GetID()]
		assert.False(t, exists)

		// Verify not found
		_, err = repo.GetByID(template.GetID())
		assert.Error(t, err)
	})

	t.Run("List with pagination", func(t *testing.T) {
		repo, err := NewEnhancedFileElementRepository(filepath.Join(tmpDir, "pagination"), 10)
		require.NoError(t, err)

		// Create 10 personas
		for i := 0; i < 10; i++ {
			persona := domain.NewPersona("Persona "+string(rune(i+'0')), "Test", "1.0.0", "test")
			require.NoError(t, repo.Create(persona))
		}

		// Get first page
		results, err := repo.List(domain.ElementFilter{Limit: 5, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, results, 5)

		// Get second page
		results, err = repo.List(domain.ElementFilter{Limit: 5, Offset: 5})
		require.NoError(t, err)
		assert.Len(t, results, 5)

		// Get beyond available
		results, err = repo.List(domain.ElementFilter{Limit: 5, Offset: 10})
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestEnhancedFileElementRepository_TypeConversion(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := NewEnhancedFileElementRepository(filepath.Join(tmpDir, "types"), 10)
	require.NoError(t, err)

	tests := []struct {
		name         string
		element      domain.Element
		expectedType domain.ElementType
	}{
		{"Persona", domain.NewPersona("Test", "desc", "1.0", "test"), domain.PersonaElement},
		{"Skill", domain.NewSkill("Test", "desc", "1.0", "test"), domain.SkillElement},
		{"Template", domain.NewTemplate("Test", "desc", "1.0", "test"), domain.TemplateElement},
		{"Agent", domain.NewAgent("Test", "desc", "1.0", "test"), domain.AgentElement},
		{"Memory", domain.NewMemory("Test", "desc", "1.0", "test"), domain.MemoryElement},
		{"Ensemble", domain.NewEnsemble("Test", "desc", "1.0", "test"), domain.EnsembleElement},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, repo.Create(tt.element))

			retrieved, err := repo.GetByID(tt.element.GetID())
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, retrieved.GetType())
			assert.Equal(t, tt.element.GetID(), retrieved.GetID())
		})
	}
}
