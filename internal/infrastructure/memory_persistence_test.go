package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestMemoryContentPersistence verifica se o conteúdo da Memory é persistido corretamente.
func TestMemoryContentPersistence(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	repo, err := NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	// Criar Memory com conteúdo
	memory := domain.NewMemory("Test Memory", "Memory for testing persistence", "1.0.0", "test-author")
	memory.Content = "This is the memory content that MUST be persisted in YAML!"
	memory.ComputeHash()
	memory.SearchIndex = []string{"test", "persistence", "memory"}
	memory.Metadata = map[string]string{"priority": "high", "category": "test"}

	// Salvar
	err = repo.Create(memory)
	require.NoError(t, err, "Failed to create memory")

	// Verificar arquivo YAML diretamente
	typeDir := string(memory.GetMetadata().Type)
	dateDir := memory.GetMetadata().CreatedAt.Format("2006-01-02")
	yamlPath := filepath.Join(tmpDir, typeDir, dateDir, sanitizeFileName(memory.GetID())+".yaml")

	t.Logf("Checking YAML file at: %s", yamlPath)
	require.FileExists(t, yamlPath, "YAML file should exist")

	// Ler arquivo YAML
	yamlData, err := os.ReadFile(yamlPath)
	require.NoError(t, err, "Failed to read YAML file")

	t.Logf("YAML Content:\n%s", string(yamlData))

	// Parse YAML
	var stored StoredElement
	err = yaml.Unmarshal(yamlData, &stored)
	require.NoError(t, err, "Failed to unmarshal YAML")

	// Verificar que Data não está vazio
	assert.NotNil(t, stored.Data, "StoredElement.Data should not be nil")
	assert.NotEmpty(t, stored.Data, "StoredElement.Data should not be empty")

	// Verificar campos específicos em Data
	assert.Contains(t, stored.Data, "content", "Data should contain 'content' field")
	assert.Contains(t, stored.Data, "content_hash", "Data should contain 'content_hash' field")
	assert.Contains(t, stored.Data, "search_index", "Data should contain 'search_index' field")
	assert.Contains(t, stored.Data, "metadata", "Data should contain 'metadata' field")

	// Verificar valores
	assert.Equal(t, memory.Content, stored.Data["content"], "Content should match")
	assert.Equal(t, memory.ContentHash, stored.Data["content_hash"], "ContentHash should match")

	// Recuperar do repositório
	retrieved, err := repo.GetByID(memory.GetID())
	require.NoError(t, err, "Failed to retrieve memory")

	// Verificar tipo
	retrievedMemory, ok := retrieved.(*domain.Memory)
	require.True(t, ok, "Retrieved element should be *domain.Memory")

	// Verificar conteúdo recuperado
	assert.Equal(t, memory.Content, retrievedMemory.Content, "Retrieved content should match original")
	assert.Equal(t, memory.ContentHash, retrievedMemory.ContentHash, "Retrieved hash should match original")
	assert.Equal(t, memory.SearchIndex, retrievedMemory.SearchIndex, "Retrieved search index should match original")
	assert.Equal(t, memory.Metadata, retrievedMemory.Metadata, "Retrieved metadata should match original")
}

// TestPersonaContentPersistence verifica se dados da Persona são persistidos.
func TestPersonaContentPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	// Criar Persona
	persona := domain.NewPersona("Test Persona", "Persona for testing", "1.0.0", "test-author")
	persona.SetSystemPrompt("You are a test persona for validation purposes.")

	trait := domain.BehavioralTrait{Name: "analytical", Intensity: 9}
	err = persona.AddBehavioralTrait(trait)
	require.NoError(t, err)

	// Salvar
	err = repo.Create(persona)
	require.NoError(t, err)

	// Ler arquivo YAML
	typeDir := string(persona.GetMetadata().Type)
	dateDir := persona.GetMetadata().CreatedAt.Format("2006-01-02")
	yamlPath := filepath.Join(tmpDir, typeDir, dateDir, sanitizeFileName(persona.GetID())+".yaml")
	yamlData, err := os.ReadFile(yamlPath)
	require.NoError(t, err)

	t.Logf("Persona YAML:\n%s", string(yamlData))

	var stored StoredElement
	err = yaml.Unmarshal(yamlData, &stored)
	require.NoError(t, err)

	// Verificar Data
	assert.NotNil(t, stored.Data, "Persona Data should not be nil")
	assert.Contains(t, stored.Data, "system_prompt", "Data should contain 'system_prompt'")
	assert.Contains(t, stored.Data, "behavioral_traits", "Data should contain 'behavioral_traits'")

	assert.Equal(t, persona.SystemPrompt, stored.Data["system_prompt"], "SystemPrompt should match")
}

// TestSkillContentPersistence verifica se dados da Skill são persistidos.
func TestSkillContentPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	// Criar Skill
	skill := domain.NewSkill("Test Skill", "Skill for testing", "1.0.0", "test-author")

	trigger := domain.SkillTrigger{
		Type:     "keyword",
		Keywords: []string{"test"},
	}
	err = skill.AddTrigger(trigger)
	require.NoError(t, err)

	procedure := domain.SkillProcedure{
		Step:        1,
		Action:      "test_action",
		Description: "Test procedure",
	}
	err = skill.AddProcedure(procedure)
	require.NoError(t, err)

	// Salvar
	err = repo.Create(skill)
	require.NoError(t, err)

	typeDir := string(skill.GetMetadata().Type)
	dateDir := skill.GetMetadata().CreatedAt.Format("2006-01-02")
	yamlPath := filepath.Join(tmpDir, typeDir, dateDir, sanitizeFileName(skill.GetID())+".yaml")
	yamlData, err := os.ReadFile(yamlPath)
	require.NoError(t, err)

	t.Logf("Skill YAML:\n%s", string(yamlData))

	var stored StoredElement
	err = yaml.Unmarshal(yamlData, &stored)
	require.NoError(t, err)

	// Verificar Data
	assert.NotNil(t, stored.Data, "Skill Data should not be nil")
	assert.Contains(t, stored.Data, "triggers", "Data should contain 'triggers'")
	assert.Contains(t, stored.Data, "procedures", "Data should contain 'procedures'")
}
