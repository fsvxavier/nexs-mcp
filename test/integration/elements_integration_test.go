package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

func setupIntegrationTest(t *testing.T) *infrastructure.InMemoryElementRepository {
	return infrastructure.NewInMemoryElementRepository()
}

// TestSkillWithTemplate demonstrates a Skill referencing a Template for output rendering
func TestSkillWithTemplate(t *testing.T) {
	repo := setupIntegrationTest(t)

	// Create a Template for email responses
	template := domain.NewTemplate("Email Response", "Customer email template", "1.0.0", "test")
	template.Content = "Dear {{name}}, Thank you for {{action}}."
	template.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
		{Name: "action", Type: "string", Required: true},
	}
	require.NoError(t, repo.Create(template))

	// Create a Skill that uses the Template
	skill := domain.NewSkill("Response Skill", "Generate responses", "1.0.0", "test")
	skill.AddTrigger(domain.SkillTrigger{Type: "keyword", Keywords: []string{"respond"}})
	skill.AddProcedure(domain.SkillProcedure{
		Step:        1,
		Action:      "Render email using template",
		Description: "Use template " + template.GetID(),
		ToolsUsed:   []string{"template_renderer"},
	})
	require.NoError(t, repo.Create(skill))

	// Verify Template can render correctly
	rendered, err := template.Render(map[string]string{"name": "John", "action": "contacting us"})
	require.NoError(t, err)
	assert.Contains(t, rendered, "Dear John")
	assert.Contains(t, rendered, "Thank you for contacting us")

	// Verify Skill references the Template
	retrievedSkill, err := repo.GetByID(skill.GetID())
	require.NoError(t, err)
	assert.Len(t, retrievedSkill.(*domain.Skill).Procedures, 1)
	assert.Contains(t, retrievedSkill.(*domain.Skill).Procedures[0].Description, template.GetID())

	t.Log("✓ Skill successfully references Template for rendering")
}

// TestAgentExecutingSkills demonstrates an Agent orchestrating multiple Skills
func TestAgentExecutingSkills(t *testing.T) {
	repo := setupIntegrationTest(t)

	// Create Skills for different tasks
	skill1 := domain.NewSkill("Analyze Data", "Data analysis", "1.0.0", "test")
	skill1.AddTrigger(domain.SkillTrigger{Type: "manual"})
	skill1.AddProcedure(domain.SkillProcedure{Step: 1, Action: "Analyze dataset"})
	require.NoError(t, repo.Create(skill1))

	skill2 := domain.NewSkill("Generate Report", "Report generation", "1.0.0", "test")
	skill2.AddTrigger(domain.SkillTrigger{Type: "manual"})
	skill2.AddProcedure(domain.SkillProcedure{Step: 1, Action: "Create report"})
	require.NoError(t, repo.Create(skill2))

	// Create an Agent that orchestrates the Skills
	agent := domain.NewAgent("Analysis Agent", "Data analysis workflow", "1.0.0", "test")
	agent.Goals = []string{"analyze data", "generate insights"}
	agent.Actions = []domain.AgentAction{
		{Name: "analyze", Type: "skill", Parameters: map[string]string{"skill_id": skill1.GetID()}},
		{Name: "report", Type: "skill", Parameters: map[string]string{"skill_id": skill2.GetID()}},
	}
	require.NoError(t, repo.Create(agent))

	// Verify Agent references both Skills
	retrievedAgent, err := repo.GetByID(agent.GetID())
	require.NoError(t, err)
	assert.Len(t, retrievedAgent.(*domain.Agent).Actions, 2)
	assert.Equal(t, "skill", retrievedAgent.(*domain.Agent).Actions[0].Type)
	assert.Equal(t, "skill", retrievedAgent.(*domain.Agent).Actions[1].Type)

	t.Logf("✓ Agent orchestrates %d Skills in a workflow", len(agent.Actions))
}

// TestEnsembleCoordinatingAgents demonstrates an Ensemble coordinating multiple Agents
func TestEnsembleCoordinatingAgents(t *testing.T) {
	repo := setupIntegrationTest(t)

	// Create Agents with different specializations
	agent1 := domain.NewAgent("Security Agent", "Security review", "1.0.0", "test")
	agent1.Goals = []string{"check security vulnerabilities"}
	agent1.Actions = []domain.AgentAction{{Name: "scan", Type: "tool"}}
	require.NoError(t, repo.Create(agent1))

	agent2 := domain.NewAgent("Performance Agent", "Performance review", "1.0.0", "test")
	agent2.Goals = []string{"optimize performance"}
	agent2.Actions = []domain.AgentAction{{Name: "profile", Type: "tool"}}
	require.NoError(t, repo.Create(agent2))

	// Create an Ensemble to coordinate the Agents
	ensemble := domain.NewEnsemble("Code Review Ensemble", "Multi-agent code review", "1.0.0", "test")
	ensemble.Members = []domain.EnsembleMember{
		{AgentID: agent1.GetID(), Role: "security", Priority: 1},
		{AgentID: agent2.GetID(), Role: "performance", Priority: 2},
	}
	ensemble.ExecutionMode = "parallel"
	ensemble.AggregationStrategy = "merge"
	require.NoError(t, repo.Create(ensemble))

	// Verify Ensemble references both Agents
	assert.Len(t, ensemble.Members, 2)
	for _, member := range ensemble.Members {
		agent, err := repo.GetByID(member.AgentID)
		require.NoError(t, err)
		assert.Equal(t, domain.AgentElement, agent.GetType())
	}

	t.Logf("✓ Ensemble coordinates %d Agents in %s mode", len(ensemble.Members), ensemble.ExecutionMode)
}

// TestMemoryDeduplication verifies Memory deduplication via SHA-256 hashing
func TestMemoryDeduplication(t *testing.T) {
	repo := setupIntegrationTest(t)

	content := "Important information that should be deduplicated"

	// Create first Memory
	memory1 := domain.NewMemory("Notes 1", "First entry", "1.0.0", "test")
	memory1.Content = content
	memory1.ComputeHash()
	require.NoError(t, repo.Create(memory1))

	// Create second Memory with same content
	memory2 := domain.NewMemory("Notes 2", "Second entry", "1.0.0", "test")
	memory2.Content = content
	memory2.ComputeHash()
	require.NoError(t, repo.Create(memory2))

	// Verify both have the same hash (can be deduplicated)
	assert.Equal(t, memory1.ContentHash, memory2.ContentHash)
	assert.Len(t, memory1.ContentHash, 64) // SHA-256 produces 64 hex characters

	// Create third Memory with different content
	memory3 := domain.NewMemory("Notes 3", "Different entry", "1.0.0", "test")
	memory3.Content = "Different content entirely"
	memory3.ComputeHash()
	require.NoError(t, repo.Create(memory3))

	// Verify hash is different
	assert.NotEqual(t, memory1.ContentHash, memory3.ContentHash)

	t.Log("✓ Memory deduplication working correctly via SHA-256 hashing")
}

// TestPersonaHotSwap verifies Persona can be activated/deactivated at runtime
func TestPersonaHotSwap(t *testing.T) {
	repo := setupIntegrationTest(t)

	// Create first Persona (active)
	persona1 := domain.NewPersona("Technical Expert", "Expert assistant", "1.0.0", "test")
	persona1.SetSystemPrompt("You are a technical expert")
	persona1.AddBehavioralTrait(domain.BehavioralTrait{Name: "analytical", Intensity: 9})
	persona1.AddExpertiseArea(domain.ExpertiseArea{Domain: "technology", Level: "expert"})
	persona1.SetResponseStyle(domain.ResponseStyle{Tone: "professional", Formality: "formal", Verbosity: "detailed"})
	require.NoError(t, repo.Create(persona1))
	assert.True(t, persona1.IsActive())

	// Create second Persona (inactive)
	persona2 := domain.NewPersona("Creative Writer", "Creative assistant", "1.0.0", "test")
	persona2.SetSystemPrompt("You are a creative writer")
	persona2.AddBehavioralTrait(domain.BehavioralTrait{Name: "imaginative", Intensity: 10})
	persona2.AddExpertiseArea(domain.ExpertiseArea{Domain: "writing", Level: "expert"})
	persona2.SetResponseStyle(domain.ResponseStyle{Tone: "warm", Formality: "casual", Verbosity: "balanced"})
	persona2.Deactivate()
	require.NoError(t, repo.Create(persona2))
	assert.False(t, persona2.IsActive())

	// Swap: deactivate persona1, activate persona2
	persona1.Deactivate()
	require.NoError(t, repo.Update(persona1))

	persona2.Activate()
	require.NoError(t, repo.Update(persona2))

	// Verify the swap
	updated1, err := repo.GetByID(persona1.GetID())
	require.NoError(t, err)
	assert.False(t, updated1.IsActive())

	updated2, err := repo.GetByID(persona2.GetID())
	require.NoError(t, err)
	assert.True(t, updated2.IsActive())

	t.Log("✓ Persona hot-swap successful: switched from Technical Expert to Creative Writer")
}

// TestE2EAllElementTypes creates and verifies all 6 element types working together
func TestE2EAllElementTypes(t *testing.T) {
	repo := setupIntegrationTest(t)

	// Create Persona
	persona := domain.NewPersona("E2E Persona", "Test assistant", "1.0.0", "e2e")
	persona.SetSystemPrompt("You are a helpful assistant")
	persona.SetResponseStyle(domain.ResponseStyle{Tone: "friendly", Formality: "neutral", Verbosity: "balanced"})
	require.NoError(t, repo.Create(persona))

	// Create Template
	template := domain.NewTemplate("E2E Template", "Test template", "1.0.0", "e2e")
	template.Content = "Status: {{status}}"
	template.Format = "markdown"
	require.NoError(t, repo.Create(template))

	// Create Skill
	skill := domain.NewSkill("E2E Skill", "Test skill", "1.0.0", "e2e")
	skill.AddTrigger(domain.SkillTrigger{Type: "manual"})
	skill.AddProcedure(domain.SkillProcedure{Step: 1, Action: "Execute test"})
	require.NoError(t, repo.Create(skill))

	// Create Agent
	agent := domain.NewAgent("E2E Agent", "Test agent", "1.0.0", "e2e")
	agent.Goals = []string{"test workflow"}
	agent.Actions = []domain.AgentAction{{Name: "test", Type: "skill"}}
	require.NoError(t, repo.Create(agent))

	// Create Memory
	memory := domain.NewMemory("E2E Memory", "Test log", "1.0.0", "e2e")
	memory.Content = "Test execution log"
	memory.ComputeHash()
	require.NoError(t, repo.Create(memory))

	// Create Ensemble
	ensemble := domain.NewEnsemble("E2E Ensemble", "Test coordination", "1.0.0", "e2e")
	ensemble.Members = []domain.EnsembleMember{{AgentID: agent.GetID(), Role: "tester", Priority: 1}}
	ensemble.ExecutionMode = "sequential"
	ensemble.AggregationStrategy = "merge"
	require.NoError(t, repo.Create(ensemble))

	// Verify all elements were created
	allIDs := []string{
		persona.GetID(),
		template.GetID(),
		skill.GetID(),
		agent.GetID(),
		memory.GetID(),
		ensemble.GetID(),
	}

	for _, id := range allIDs {
		elem, err := repo.GetByID(id)
		require.NoError(t, err)
		assert.NotNil(t, elem)
	}

	// Verify counts by type
	allElems, err := repo.List(domain.ElementFilter{})
	require.NoError(t, err)

	counts := make(map[domain.ElementType]int)
	for _, e := range allElems {
		counts[e.GetType()]++
	}

	assert.Equal(t, 1, counts[domain.PersonaElement])
	assert.Equal(t, 1, counts[domain.SkillElement])
	assert.Equal(t, 1, counts[domain.TemplateElement])
	assert.Equal(t, 1, counts[domain.AgentElement])
	assert.Equal(t, 1, counts[domain.MemoryElement])
	assert.Equal(t, 1, counts[domain.EnsembleElement])

	t.Logf("✓ E2E Test PASSED: All 6 element types created and verified (%d total elements)", len(allElems))
}
