package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockElementRepository implementa domain.ElementRepository para testes.
type mockElementRepository struct {
	elements map[string]domain.Element
	mu       sync.RWMutex
	delay    time.Duration
	failIDs  map[string]error
}

func newMockRepo() *mockElementRepository {
	return &mockElementRepository{
		elements: make(map[string]domain.Element),
		failIDs:  make(map[string]error),
	}
}

func (m *mockElementRepository) GetByID(id string) (domain.Element, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.failIDs[id]; ok {
		return nil, err
	}

	elem, ok := m.elements[id]
	if !ok {
		return nil, errors.New("element not found")
	}

	return elem, nil
}

func (m *mockElementRepository) List(filter domain.ElementFilter) ([]domain.Element, error) {
	return nil, errors.New("not implemented")
}

func (m *mockElementRepository) Create(elem domain.Element) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.elements[elem.GetID()] = elem
	return nil
}

func (m *mockElementRepository) Update(elem domain.Element) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.elements[elem.GetID()] = elem
	return nil
}

func (m *mockElementRepository) Exists(id string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.elements[id]
	return ok, nil
}

func (m *mockElementRepository) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.elements, id)
	return nil
}

func (m *mockElementRepository) addElement(elem domain.Element) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.elements[elem.GetID()] = elem
}

func (m *mockElementRepository) addFailure(id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failIDs[id] = err
}

// Helper functions for creating test elements with specific IDs
// Usando nomes únicos para garantir IDs previsíveis.
func createTestMemoryForEnrichment(uniqueName string, relatedTo string) *domain.Memory {
	memory := domain.NewMemory(uniqueName, "Test description", "1.0.0", "test")
	if relatedTo != "" {
		memory.Metadata["related_to"] = relatedTo
	}
	return memory
}

func createTestPersonaForEnrichment(uniqueName string) *domain.Persona {
	return domain.NewPersona(uniqueName, "A test persona", "1.0.0", "test")
}

func createTestSkillForEnrichment(uniqueName string) *domain.Skill {
	return domain.NewSkill(uniqueName, "A test skill", "1.0.0", "test")
}

func createTestAgentForEnrichment(uniqueName string) *domain.Agent {
	return domain.NewAgent(uniqueName, "A test agent", "1.0.0", "test")
}

func createTestTemplateForEnrichment(uniqueName string) *domain.Template {
	return domain.NewTemplate(uniqueName, "A test template", "1.0.0", "test")
}

// Helpers para criar elementos e obter seus IDs.
func createPersonaWithID(repo *mockElementRepository, name string) string {
	persona := createTestPersonaForEnrichment(name)
	repo.addElement(persona)
	return persona.GetID()
}

func createSkillWithID(repo *mockElementRepository, name string) string {
	skill := createTestSkillForEnrichment(name)
	repo.addElement(skill)
	return skill.GetID()
}

func createAgentWithID(repo *mockElementRepository, name string) string {
	agent := createTestAgentForEnrichment(name)
	repo.addElement(agent)
	return agent.GetID()
}

func createTemplateWithID(repo *mockElementRepository, name string) string {
	template := createTestTemplateForEnrichment(name)
	repo.addElement(template)
	return template.GetID()
}

// Test Cases

func TestExpandMemoryContext_Success(t *testing.T) {
	t.Run("expand with persona and skill", func(t *testing.T) {
		repo := newMockRepo()

		// Criar elementos e obter IDs reais
		personaID := createPersonaWithID(repo, "TestPersona1")
		skillID := createSkillWithID(repo, "TestSkill1")

		// Criar memory com IDs reais
		memory := createTestMemoryForEnrichment("TestMemory1", personaID+","+skillID)

		options := DefaultExpandOptions()
		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		require.NoError(t, err)
		assert.Equal(t, 2, enriched.GetElementCount())
		assert.False(t, enriched.HasErrors())
		assert.GreaterOrEqual(t, enriched.TotalTokensSaved, 275)
		assert.Greater(t, enriched.FetchDuration, time.Duration(0))
	})

	t.Run("no related elements", func(t *testing.T) {
		repo := newMockRepo()
		memory := createTestMemoryForEnrichment("TestMemory2", "")

		options := DefaultExpandOptions()
		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		require.NoError(t, err)
		assert.Equal(t, 0, enriched.GetElementCount())
		assert.False(t, enriched.HasErrors())
		assert.Equal(t, 0, enriched.TotalTokensSaved)
	})

	t.Run("single related element", func(t *testing.T) {
		repo := newMockRepo()

		agentID := createAgentWithID(repo, "TestAgent1")
		memory := createTestMemoryForEnrichment("TestMemory3", agentID)

		options := DefaultExpandOptions()
		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		require.NoError(t, err)
		assert.Equal(t, 1, enriched.GetElementCount())
		assert.False(t, enriched.HasErrors())
		assert.GreaterOrEqual(t, enriched.TotalTokensSaved, 100)
	})

	t.Run("multiple element types", func(t *testing.T) {
		repo := newMockRepo()

		personaID := createPersonaWithID(repo, "TestPersona2")
		skillID := createSkillWithID(repo, "TestSkill2")
		agentID := createAgentWithID(repo, "TestAgent2")
		templateID := createTemplateWithID(repo, "TestTemplate2")

		relatedIDs := strings.Join([]string{personaID, skillID, agentID, templateID}, ",")
		memory := createTestMemoryForEnrichment("TestMemory4", relatedIDs)

		options := DefaultExpandOptions()
		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		require.NoError(t, err)
		assert.Equal(t, 4, enriched.GetElementCount())
		assert.False(t, enriched.HasErrors())
		assert.GreaterOrEqual(t, enriched.TotalTokensSaved, 500)
	})
}

func TestExpandMemoryContext_WithErrors(t *testing.T) {
	t.Run("missing element with error", func(t *testing.T) {
		repo := newMockRepo()

		personaID := createPersonaWithID(repo, "TestPersona3")
		repo.addFailure("missing-id", errors.New("not found"))

		memory := createTestMemoryForEnrichment("TestMemory5", personaID+",missing-id")

		options := DefaultExpandOptions()
		options.IgnoreErrors = false

		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		assert.Error(t, err)
		assert.Equal(t, 1, enriched.GetElementCount()) // persona should succeed
	})

	t.Run("missing element ignore errors", func(t *testing.T) {
		repo := newMockRepo()

		personaID := createPersonaWithID(repo, "TestPersona4")
		repo.addFailure("missing-id", errors.New("not found"))

		memory := createTestMemoryForEnrichment("TestMemory6", personaID+",missing-id")

		options := DefaultExpandOptions()
		options.IgnoreErrors = true

		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		require.NoError(t, err)
		assert.Equal(t, 1, enriched.GetElementCount())
		assert.True(t, enriched.HasErrors())
		assert.Equal(t, 1, enriched.GetErrorCount())
	})

	t.Run("all elements missing", func(t *testing.T) {
		repo := newMockRepo()

		repo.addFailure("missing-1", errors.New("not found"))
		repo.addFailure("missing-2", errors.New("not found"))

		memory := createTestMemoryForEnrichment("TestMemory7", "missing-1,missing-2")

		options := DefaultExpandOptions()
		options.IgnoreErrors = false

		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		assert.Error(t, err)
		assert.Equal(t, 0, enriched.GetElementCount())
		assert.True(t, enriched.HasErrors())
	})
}

func TestExpandMemoryContext_TypeFilters(t *testing.T) {
	t.Run("include only personas", func(t *testing.T) {
		repo := newMockRepo()

		personaID := createPersonaWithID(repo, "TestPersona5")
		skillID := createSkillWithID(repo, "TestSkill3")
		agentID := createAgentWithID(repo, "TestAgent3")

		relatedIDs := strings.Join([]string{personaID, skillID, agentID}, ",")
		memory := createTestMemoryForEnrichment("TestMemory8", relatedIDs)

		options := DefaultExpandOptions()
		options.IncludeTypes = []domain.ElementType{domain.PersonaElement}

		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		require.NoError(t, err)
		assert.Equal(t, 1, enriched.GetElementCount())

		personas := enriched.GetElementsByType(domain.PersonaElement)
		assert.Len(t, personas, 1)
	})

	t.Run("exclude skills", func(t *testing.T) {
		repo := newMockRepo()

		personaID := createPersonaWithID(repo, "TestPersona6")
		skillID := createSkillWithID(repo, "TestSkill4")
		agentID := createAgentWithID(repo, "TestAgent4")

		relatedIDs := strings.Join([]string{personaID, skillID, agentID}, ",")
		memory := createTestMemoryForEnrichment("TestMemory9", relatedIDs)

		options := DefaultExpandOptions()
		options.ExcludeTypes = []domain.ElementType{domain.SkillElement}

		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		require.NoError(t, err)
		assert.Equal(t, 2, enriched.GetElementCount())

		skills := enriched.GetElementsByType(domain.SkillElement)
		assert.Len(t, skills, 0)
	})

	t.Run("include personas and agents", func(t *testing.T) {
		repo := newMockRepo()

		personaID := createPersonaWithID(repo, "TestPersona7")
		skillID := createSkillWithID(repo, "TestSkill5")
		agentID := createAgentWithID(repo, "TestAgent5")

		relatedIDs := strings.Join([]string{personaID, skillID, agentID}, ",")
		memory := createTestMemoryForEnrichment("TestMemory10", relatedIDs)

		options := DefaultExpandOptions()
		options.IncludeTypes = []domain.ElementType{domain.PersonaElement, domain.AgentElement}

		enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

		require.NoError(t, err)
		assert.Equal(t, 2, enriched.GetElementCount())

		personas := enriched.GetElementsByType(domain.PersonaElement)
		assert.Len(t, personas, 1)

		agents := enriched.GetElementsByType(domain.AgentElement)
		assert.Len(t, agents, 1)
	})
}

func TestExpandMemoryContext_MaxElements(t *testing.T) {
	repo := newMockRepo()

	// Create 30 personas
	relatedIDs := make([]string, 30)
	for i := range 30 {
		relatedIDs[i] = createPersonaWithID(repo, fmt.Sprintf("TestPersona%d", i+100))
	}

	memory := createTestMemoryForEnrichment("TestMemory11", strings.Join(relatedIDs, ","))

	options := DefaultExpandOptions()
	options.MaxElements = 10

	enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

	require.NoError(t, err)
	assert.Equal(t, 10, enriched.GetElementCount(), "should limit to MaxElements")
}

func TestExpandMemoryContext_Parallel(t *testing.T) {
	repo := newMockRepo()
	repo.delay = 10 * time.Millisecond // Simulate network delay

	personaID := createPersonaWithID(repo, "TestPersona8")
	skillID := createSkillWithID(repo, "TestSkill6")
	agentID := createAgentWithID(repo, "TestAgent6")

	relatedIDs := strings.Join([]string{personaID, skillID, agentID}, ",")
	memory := createTestMemoryForEnrichment("TestMemory12", relatedIDs)

	options := DefaultExpandOptions()
	options.FetchStrategy = "parallel"

	start := time.Now()
	enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 3, enriched.GetElementCount())

	// Parallel should be faster than sequential (3 * 10ms = 30ms)
	// Allow margin for overhead
	assert.Less(t, duration, 25*time.Millisecond, "parallel fetch should be faster")
}

func TestExpandMemoryContext_Sequential(t *testing.T) {
	repo := newMockRepo()
	repo.delay = 5 * time.Millisecond

	personaID := createPersonaWithID(repo, "TestPersona9")
	skillID := createSkillWithID(repo, "TestSkill7")

	relatedIDs := strings.Join([]string{personaID, skillID}, ",")
	memory := createTestMemoryForEnrichment("TestMemory13", relatedIDs)

	options := DefaultExpandOptions()
	options.FetchStrategy = "sequential"

	start := time.Now()
	enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)
	require.NoError(t, err)
	duration := time.Since(start)

	assert.Equal(t, 2, enriched.GetElementCount())
	assert.GreaterOrEqual(t, duration, 10*time.Millisecond, "sequential should take at least 2*5ms")
}

func TestExpandMemoryContext_Timeout(t *testing.T) {
	repo := newMockRepo()
	repo.delay = 100 * time.Millisecond

	personaID := createPersonaWithID(repo, "TestPersona10")
	skillID := createSkillWithID(repo, "TestSkill8")

	relatedIDs := strings.Join([]string{personaID, skillID}, ",")
	memory := createTestMemoryForEnrichment("TestMemory14", relatedIDs)

	options := DefaultExpandOptions()
	options.Timeout = 50 * time.Millisecond
	options.IgnoreErrors = true

	enriched, err := ExpandMemoryContext(context.Background(), memory, repo, options)

	require.NoError(t, err)
	// May have partial results depending on timing
	assert.LessOrEqual(t, enriched.GetElementCount(), 2)
}

func TestEnrichedContext_HelperMethods(t *testing.T) {
	repo := newMockRepo()

	personaID := createPersonaWithID(repo, "TestPersona11")
	skillID := createSkillWithID(repo, "TestSkill9")

	relatedIDs := strings.Join([]string{personaID, skillID}, ",")
	memory := createTestMemoryForEnrichment("TestMemory15", relatedIDs)

	enriched, err := ExpandMemoryContext(context.Background(), memory, repo, DefaultExpandOptions())
	require.NoError(t, err)

	t.Run("GetElementByID", func(t *testing.T) {
		elem, ok := enriched.GetElementByID(personaID)
		assert.True(t, ok)
		assert.Equal(t, domain.PersonaElement, elem.GetType())

		elem, ok = enriched.GetElementByID("missing-id")
		assert.False(t, ok)
		assert.Nil(t, elem)
	})

	t.Run("GetElementsByType", func(t *testing.T) {
		personas := enriched.GetElementsByType(domain.PersonaElement)
		assert.Len(t, personas, 1)

		skills := enriched.GetElementsByType(domain.SkillElement)
		assert.Len(t, skills, 1)

		agents := enriched.GetElementsByType(domain.AgentElement)
		assert.Len(t, agents, 0)
	})

	t.Run("HasErrors", func(t *testing.T) {
		assert.False(t, enriched.HasErrors())
		assert.Equal(t, 0, enriched.GetErrorCount())
	})
}

func TestParseRelatedIDs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple IDs",
			input: "id1,id2,id3",
			want:  []string{"id1", "id2", "id3"},
		},
		{
			name:  "with spaces",
			input: "id1, id2 , id3",
			want:  []string{"id1", "id2", "id3"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "single ID",
			input: "id1",
			want:  []string{"id1"},
		},
		{
			name:  "trailing comma",
			input: "id1,id2,",
			want:  []string{"id1", "id2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRelatedIDs(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShouldIncludeElement(t *testing.T) {
	persona := createTestPersonaForEnrichment("TestPersonaShouldInclude")
	skill := createTestSkillForEnrichment("TestSkillShouldInclude")

	tests := []struct {
		name    string
		element domain.Element
		options ExpandOptions
		want    bool
	}{
		{
			name:    "no filters",
			element: persona,
			options: ExpandOptions{},
			want:    true,
		},
		{
			name:    "include match",
			element: persona,
			options: ExpandOptions{
				IncludeTypes: []domain.ElementType{domain.PersonaElement},
			},
			want: true,
		},
		{
			name:    "include no match",
			element: skill,
			options: ExpandOptions{
				IncludeTypes: []domain.ElementType{domain.PersonaElement},
			},
			want: false,
		},
		{
			name:    "exclude match",
			element: persona,
			options: ExpandOptions{
				ExcludeTypes: []domain.ElementType{domain.PersonaElement},
			},
			want: false,
		},
		{
			name:    "exclude no match",
			element: skill,
			options: ExpandOptions{
				ExcludeTypes: []domain.ElementType{domain.PersonaElement},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldIncludeElement(tt.element, tt.options)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateTokenSavings(t *testing.T) {
	tests := []struct {
		name        string
		elemCount   int
		wantMinimum int
	}{
		{
			name:        "no elements",
			elemCount:   0,
			wantMinimum: 0,
		},
		{
			name:        "one element",
			elemCount:   1,
			wantMinimum: 100, // (100-25) + 50
		},
		{
			name:        "five elements",
			elemCount:   5,
			wantMinimum: 700, // (500-25) + 250
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enriched := &EnrichedContext{
				RelatedElements: make(map[string]domain.Element),
			}

			for i := range tt.elemCount {
				persona := createTestPersonaForEnrichment(fmt.Sprintf("TestPersonaTokenSavings%d", i))
				enriched.RelatedElements[persona.GetID()] = persona
			}

			saved := calculateTokenSavings(enriched)
			assert.GreaterOrEqual(t, saved, tt.wantMinimum)
		})
	}
}
