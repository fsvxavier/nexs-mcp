package infrastructure

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// TestExtractElementData_Persona tests data extraction from Persona elements.
func TestExtractElementData_Persona(t *testing.T) {
	persona := domain.NewPersona("P1", "Test Persona", "1.0.0", "tester")
	persona.BehavioralTraits = []domain.BehavioralTrait{
		{Name: "helpful", Intensity: 8},
		{Name: "quick", Intensity: 7},
	}
	persona.ExpertiseAreas = []domain.ExpertiseArea{
		{Domain: "testing", Level: "expert", Keywords: []string{"qa", "automation"}},
	}
	persona.ResponseStyle = domain.ResponseStyle{
		Tone:      "professional",
		Formality: "formal",
		Verbosity: "balanced",
	}
	persona.SystemPrompt = "You are a test assistant"
	persona.PrivacyLevel = "private"
	persona.Owner = "tester"
	persona.SharedWith = []string{"team1", "team2"}
	persona.HotSwappable = true

	data := extractElementData(persona)

	// Verify behavioral traits
	traits, ok := data["behavioral_traits"].([]domain.BehavioralTrait)
	if !ok {
		t.Fatalf("Expected behavioral_traits to be []BehavioralTrait")
	}
	if len(traits) != 2 {
		t.Errorf("Expected 2 behavioral traits, got %d", len(traits))
	}
	if traits[0].Name != "helpful" || traits[0].Intensity != 8 {
		t.Errorf("Unexpected behavioral trait: %+v", traits[0])
	}

	// Verify expertise areas
	areas, ok := data["expertise_areas"].([]domain.ExpertiseArea)
	if !ok {
		t.Fatalf("Expected expertise_areas to be []ExpertiseArea")
	}
	if len(areas) != 1 {
		t.Errorf("Expected 1 expertise area, got %d", len(areas))
	}
	if areas[0].Domain != "testing" {
		t.Errorf("Expected domain 'testing', got %s", areas[0].Domain)
	}

	// Verify response style
	style, ok := data["response_style"].(domain.ResponseStyle)
	if !ok {
		t.Fatalf("Expected response_style to be ResponseStyle")
	}
	if style.Tone != "professional" {
		t.Errorf("Expected tone 'professional', got %s", style.Tone)
	}

	// Verify system prompt
	prompt, ok := data["system_prompt"].(string)
	if !ok || prompt != "You are a test assistant" {
		t.Errorf("Expected system_prompt 'You are a test assistant', got %v", data["system_prompt"])
	}

	// Verify other fields
	if privLevel, ok := data["privacy_level"].(domain.PersonaPrivacyLevel); !ok || string(privLevel) != "private" {
		t.Errorf("Expected privacy_level 'private', got %v", data["privacy_level"])
	}
	if data["owner"] != "tester" {
		t.Errorf("Expected owner 'tester', got %v", data["owner"])
	}
	if data["hot_swappable"] != true {
		t.Errorf("Expected hot_swappable true, got %v", data["hot_swappable"])
	}
}

// TestExtractElementData_Skill tests data extraction from Skill elements.
func TestExtractElementData_Skill(t *testing.T) {
	skill := domain.NewSkill("S1", "Test Skill", "1.0.0", "tester")
	skill.Triggers = []domain.SkillTrigger{
		{Type: "keyword", Keywords: []string{"test", "qa"}},
	}
	skill.Procedures = []domain.SkillProcedure{
		{Step: 1, Action: "Run tests", Description: "Execute test suite"},
	}
	skill.Dependencies = []domain.SkillDependency{
		{SkillID: "skill_dep1", Required: true, Version: "1.0.0"},
	}
	skill.ToolsRequired = []string{"git", "docker"}
	skill.Composable = true

	data := extractElementData(skill)

	// Verify triggers
	triggers, ok := data["triggers"].([]domain.SkillTrigger)
	if !ok {
		t.Fatalf("Expected triggers to be []SkillTrigger")
	}
	if len(triggers) != 1 {
		t.Errorf("Expected 1 trigger, got %d", len(triggers))
	}
	if triggers[0].Type != "keyword" {
		t.Errorf("Expected trigger type 'keyword', got %s", triggers[0].Type)
	}

	// Verify procedures
	procedures, ok := data["procedures"].([]domain.SkillProcedure)
	if !ok {
		t.Fatalf("Expected procedures to be []SkillProcedure")
	}
	if len(procedures) != 1 {
		t.Errorf("Expected 1 procedure, got %d", len(procedures))
	}
	if procedures[0].Action != "Run tests" {
		t.Errorf("Expected action 'Run tests', got %s", procedures[0].Action)
	}

	// Verify dependencies
	deps, ok := data["dependencies"].([]domain.SkillDependency)
	if !ok {
		t.Fatalf("Expected dependencies to be []SkillDependency")
	}
	if len(deps) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(deps))
	}

	// Verify composable
	if data["composable"] != true {
		t.Errorf("Expected composable true, got %v", data["composable"])
	}
}

// TestExtractElementData_Template tests data extraction from Template elements.
func TestExtractElementData_Template(t *testing.T) {
	template := domain.NewTemplate("T1", "Test Template", "1.0.0", "tester")
	template.Content = "Hello {{name}}"
	template.Format = "text"
	template.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true, Default: "World"},
	}
	template.ValidationRules = map[string]string{
		"name": "required|string",
	}

	data := extractElementData(template)

	if data["content"] != "Hello {{name}}" {
		t.Errorf("Expected content 'Hello {{name}}', got %v", data["content"])
	}
	if data["format"] != "text" {
		t.Errorf("Expected format 'text', got %v", data["format"])
	}

	variables, ok := data["variables"].([]domain.TemplateVariable)
	if !ok {
		t.Fatalf("Expected variables to be []TemplateVariable")
	}
	if len(variables) != 1 || variables[0].Name != "name" {
		t.Errorf("Unexpected variables: %+v", variables)
	}
}

// TestRestoreElementData_Persona tests restoration of Persona data.
func TestRestoreElementData_Persona(t *testing.T) {
	// Create original persona
	original := domain.NewPersona("P1", "Test Persona", "1.0.0", "tester")
	original.BehavioralTraits = []domain.BehavioralTrait{
		{Name: "helpful", Intensity: 8},
	}
	original.ExpertiseAreas = []domain.ExpertiseArea{
		{Domain: "testing", Level: "expert", Keywords: []string{"qa"}},
	}
	original.ResponseStyle = domain.ResponseStyle{
		Tone:      "professional",
		Formality: "formal",
		Verbosity: "balanced",
	}
	original.SystemPrompt = "You are a test assistant"
	original.PrivacyLevel = "private"

	// Extract and restore
	data := extractElementData(original)
	restored := domain.NewPersona("P1", "Test Persona", "1.0.0", "tester")
	restoreElementData(restored, data)

	// Verify restoration
	if len(restored.BehavioralTraits) != 1 {
		t.Errorf("Expected 1 behavioral trait, got %d", len(restored.BehavioralTraits))
	}
	if restored.BehavioralTraits[0].Name != "helpful" {
		t.Errorf("Expected trait 'helpful', got %s", restored.BehavioralTraits[0].Name)
	}
	if len(restored.ExpertiseAreas) != 1 {
		t.Errorf("Expected 1 expertise area, got %d", len(restored.ExpertiseAreas))
	}
	if restored.ResponseStyle.Tone != "professional" {
		t.Errorf("Expected tone 'professional', got %s", restored.ResponseStyle.Tone)
	}
	if restored.SystemPrompt != "You are a test assistant" {
		t.Errorf("Expected system prompt 'You are a test assistant', got %s", restored.SystemPrompt)
	}
	if string(restored.PrivacyLevel) != "private" {
		t.Errorf("Expected privacy level 'private', got %s", restored.PrivacyLevel)
	}
}

// TestRestoreElementData_Skill tests restoration of Skill data.
func TestRestoreElementData_Skill(t *testing.T) {
	// Create original skill
	original := domain.NewSkill("S1", "Test Skill", "1.0.0", "tester")
	original.Triggers = []domain.SkillTrigger{
		{Type: "keyword", Keywords: []string{"test"}},
	}
	original.Procedures = []domain.SkillProcedure{
		{Step: 1, Action: "Run tests", Description: "Execute test suite"},
	}
	original.Composable = true

	// Extract and restore
	data := extractElementData(original)
	restored := domain.NewSkill("S1", "Test Skill", "1.0.0", "tester")
	restoreElementData(restored, data)

	// Verify restoration
	if len(restored.Triggers) != 1 {
		t.Errorf("Expected 1 trigger, got %d", len(restored.Triggers))
	}
	if restored.Triggers[0].Type != "keyword" {
		t.Errorf("Expected trigger type 'keyword', got %s", restored.Triggers[0].Type)
	}
	if len(restored.Procedures) != 1 {
		t.Errorf("Expected 1 procedure, got %d", len(restored.Procedures))
	}
	if restored.Composable != true {
		t.Errorf("Expected composable true, got %v", restored.Composable)
	}
}

// TestRestoreElementData_Template tests restoration of Template data.
func TestRestoreElementData_Template(t *testing.T) {
	// Create original template
	original := domain.NewTemplate("T1", "Test Template", "1.0.0", "tester")
	original.Content = "Hello {{name}}"
	original.Format = "text"
	original.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
	}

	// Extract and restore
	data := extractElementData(original)
	restored := domain.NewTemplate("T1", "Test Template", "1.0.0", "tester")
	restoreElementData(restored, data)

	// Verify restoration
	if restored.Content != "Hello {{name}}" {
		t.Errorf("Expected content 'Hello {{name}}', got %s", restored.Content)
	}
	if restored.Format != "text" {
		t.Errorf("Expected format 'text', got %s", restored.Format)
	}
	if len(restored.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(restored.Variables))
	}
}

// TestUnmarshalBehavioralTraits tests the unmarshal function for behavioral traits.
func TestUnmarshalBehavioralTraits(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []domain.BehavioralTrait
	}{
		{
			name: "Valid traits with intensity",
			input: []interface{}{
				map[string]interface{}{"name": "helpful", "intensity": 8.0},
				map[string]interface{}{"name": "quick", "intensity": 7.0},
			},
			expected: []domain.BehavioralTrait{
				{Name: "helpful", Intensity: 8},
				{Name: "quick", Intensity: 7},
			},
		},
		{
			name: "Trait with integer intensity",
			input: []interface{}{
				map[string]interface{}{"name": "helpful", "intensity": 8},
			},
			expected: []domain.BehavioralTrait{
				{Name: "helpful", Intensity: 8},
			},
		},
		{
			name:     "Empty array",
			input:    []interface{}{},
			expected: []domain.BehavioralTrait{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unmarshalBehavioralTraits(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d traits, got %d", len(tt.expected), len(result))
				return
			}
			for i, trait := range result {
				if trait.Name != tt.expected[i].Name {
					t.Errorf("Expected name %s, got %s", tt.expected[i].Name, trait.Name)
				}
				if trait.Intensity != tt.expected[i].Intensity {
					t.Errorf("Expected intensity %d, got %d", tt.expected[i].Intensity, trait.Intensity)
				}
			}
		})
	}
}

// TestUnmarshalExpertiseAreas tests the unmarshal function for expertise areas.
func TestUnmarshalExpertiseAreas(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []domain.ExpertiseArea
	}{
		{
			name: "Valid expertise areas",
			input: []interface{}{
				map[string]interface{}{
					"domain":   "testing",
					"level":    "expert",
					"keywords": []interface{}{"qa", "automation"},
				},
			},
			expected: []domain.ExpertiseArea{
				{Domain: "testing", Level: "expert", Keywords: []string{"qa", "automation"}},
			},
		},
		{
			name:     "Empty array",
			input:    []interface{}{},
			expected: []domain.ExpertiseArea{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unmarshalExpertiseAreas(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d areas, got %d", len(tt.expected), len(result))
				return
			}
			for i, area := range result {
				if area.Domain != tt.expected[i].Domain {
					t.Errorf("Expected domain %s, got %s", tt.expected[i].Domain, area.Domain)
				}
				if area.Level != tt.expected[i].Level {
					t.Errorf("Expected level %s, got %s", tt.expected[i].Level, area.Level)
				}
			}
		})
	}
}

// TestUnmarshalResponseStyle tests the unmarshal function for response style.
func TestUnmarshalResponseStyle(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected domain.ResponseStyle
	}{
		{
			name: "Valid response style",
			input: map[string]interface{}{
				"tone":      "professional",
				"formality": "formal",
				"verbosity": "balanced",
			},
			expected: domain.ResponseStyle{
				Tone:      "professional",
				Formality: "formal",
				Verbosity: "balanced",
			},
		},
		{
			name:     "Empty style",
			input:    map[string]interface{}{},
			expected: domain.ResponseStyle{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unmarshalResponseStyle(tt.input)
			if result.Tone != tt.expected.Tone {
				t.Errorf("Expected tone %s, got %s", tt.expected.Tone, result.Tone)
			}
			if result.Formality != tt.expected.Formality {
				t.Errorf("Expected formality %s, got %s", tt.expected.Formality, result.Formality)
			}
			if result.Verbosity != tt.expected.Verbosity {
				t.Errorf("Expected verbosity %s, got %s", tt.expected.Verbosity, result.Verbosity)
			}
		})
	}
}

// TestUnmarshalSkillTriggers tests the unmarshal function for skill triggers.
func TestUnmarshalSkillTriggers(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []domain.SkillTrigger
	}{
		{
			name: "Valid triggers",
			input: []interface{}{
				map[string]interface{}{
					"type":     "keyword",
					"keywords": []interface{}{"test", "qa"},
				},
			},
			expected: []domain.SkillTrigger{
				{Type: "keyword", Keywords: []string{"test", "qa"}},
			},
		},
		{
			name:     "Empty array",
			input:    []interface{}{},
			expected: []domain.SkillTrigger{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unmarshalSkillTriggers(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d triggers, got %d", len(tt.expected), len(result))
				return
			}
			for i, trigger := range result {
				if trigger.Type != tt.expected[i].Type {
					t.Errorf("Expected type %s, got %s", tt.expected[i].Type, trigger.Type)
				}
			}
		})
	}
}

// TestUnmarshalSkillProcedures tests the unmarshal function for skill procedures.
func TestUnmarshalSkillProcedures(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []domain.SkillProcedure
	}{
		{
			name: "Valid procedures",
			input: []interface{}{
				map[string]interface{}{
					"step":        1.0,
					"action":      "Run tests",
					"description": "Execute test suite",
				},
			},
			expected: []domain.SkillProcedure{
				{Step: 1, Action: "Run tests", Description: "Execute test suite"},
			},
		},
		{
			name:     "Empty array",
			input:    []interface{}{},
			expected: []domain.SkillProcedure{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unmarshalSkillProcedures(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d procedures, got %d", len(tt.expected), len(result))
				return
			}
			for i, proc := range result {
				if proc.Action != tt.expected[i].Action {
					t.Errorf("Expected action %s, got %s", tt.expected[i].Action, proc.Action)
				}
			}
		})
	}
}

// TestUnmarshalTemplateVariables tests the unmarshal function for template variables.
func TestUnmarshalTemplateVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []domain.TemplateVariable
	}{
		{
			name: "Valid variables",
			input: []interface{}{
				map[string]interface{}{
					"name":     "username",
					"type":     "string",
					"required": true,
					"default":  "anonymous",
				},
			},
			expected: []domain.TemplateVariable{
				{Name: "username", Type: "string", Required: true, Default: "anonymous"},
			},
		},
		{
			name:     "Empty array",
			input:    []interface{}{},
			expected: []domain.TemplateVariable{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unmarshalTemplateVariables(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d variables, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v.Name != tt.expected[i].Name {
					t.Errorf("Expected name %s, got %s", tt.expected[i].Name, v.Name)
				}
				if v.Type != tt.expected[i].Type {
					t.Errorf("Expected type %s, got %s", tt.expected[i].Type, v.Type)
				}
			}
		})
	}
}

// TestRestoreElementData_Agent tests restoration of Agent data.
func TestRestoreElementData_Agent(t *testing.T) {
	original := domain.NewAgent("A1", "Test Agent", "1.0.0", "tester")
	original.Goals = []string{"goal1", "goal2"}
	original.Actions = []domain.AgentAction{
		{Name: "action1", Type: "tool", Parameters: map[string]string{"cmd": "test"}},
	}
	original.MaxIterations = 10

	data := extractElementData(original)
	restored := domain.NewAgent("A1", "Test Agent", "1.0.0", "tester")
	restoreElementData(restored, data)

	if len(restored.Goals) != 2 {
		t.Errorf("Expected 2 goals, got %d", len(restored.Goals))
	}
	if len(restored.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(restored.Actions))
	}
	if restored.MaxIterations != 10 {
		t.Errorf("Expected max_iterations 10, got %d", restored.MaxIterations)
	}
}

// TestRestoreElementData_Memory tests restoration of Memory data.
func TestRestoreElementData_Memory(t *testing.T) {
	original := domain.NewMemory("M1", "Test Memory", "1.0.0", "tester")
	original.Content = "Test content"
	original.ContentHash = "abc123"
	original.DateCreated = "2025-12-19"
	original.SearchIndex = []string{"test", "content"}

	data := extractElementData(original)
	restored := domain.NewMemory("M1", "Test Memory", "1.0.0", "tester")
	restoreElementData(restored, data)

	if restored.Content != "Test content" {
		t.Errorf("Expected content 'Test content', got %s", restored.Content)
	}
	if restored.ContentHash != "abc123" {
		t.Errorf("Expected content_hash 'abc123', got %s", restored.ContentHash)
	}
	if len(restored.SearchIndex) != 2 {
		t.Errorf("Expected 2 search index items, got %d", len(restored.SearchIndex))
	}
}

// TestRestoreElementData_Ensemble tests restoration of Ensemble data.
func TestRestoreElementData_Ensemble(t *testing.T) {
	original := domain.NewEnsemble("E1", "Test Ensemble", "1.0.0", "tester")
	original.Members = []domain.EnsembleMember{
		{AgentID: "agent1", Role: "leader", Priority: 1},
	}
	original.ExecutionMode = "parallel"
	original.AggregationStrategy = "consensus"

	data := extractElementData(original)
	restored := domain.NewEnsemble("E1", "Test Ensemble", "1.0.0", "tester")
	restoreElementData(restored, data)

	if len(restored.Members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(restored.Members))
	}
	if restored.ExecutionMode != "parallel" {
		t.Errorf("Expected execution_mode 'parallel', got %s", restored.ExecutionMode)
	}
	if restored.AggregationStrategy != "consensus" {
		t.Errorf("Expected aggregation_strategy 'consensus', got %s", restored.AggregationStrategy)
	}
}
