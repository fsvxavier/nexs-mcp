package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/config"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const userContextKey contextKey = "user"

func TestQuickCreatePersona(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	cfg := &config.Config{
		DataDir:     tmpDir,
		StorageType: "file",
	}

	server := NewMCPServer("test-server", "1.0.0", repo, cfg)

	tests := []struct {
		name        string
		input       QuickCreatePersonaInput
		wantErr     bool
		errContains string
		validate    func(t *testing.T, output map[string]interface{})
	}{
		{
			name: "create with technical template",
			input: QuickCreatePersonaInput{
				Name:     "DevOps Expert",
				Template: "technical",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				assert.Equal(t, "DevOps Expert", output["name"])
				assert.Equal(t, domain.PersonaElement, output["type"])
				assert.Equal(t, "technical", output["template"])
				assert.NotEmpty(t, output["id"])

				// Verify persona was created in repo
				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				persona, ok := elem.(*domain.Persona)
				require.True(t, ok)
				assert.NotEmpty(t, persona.BehavioralTraits)
				assert.NotEmpty(t, persona.ExpertiseAreas)
				assert.NotEmpty(t, persona.SystemPrompt)
			},
		},
		{
			name: "create with custom expertise",
			input: QuickCreatePersonaInput{
				Name:        "Cloud Architect",
				Description: "Expert in cloud infrastructure",
				Expertise:   []string{"aws", "kubernetes", "terraform"},
				Template:    "technical",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				persona, ok := elem.(*domain.Persona)
				require.True(t, ok)
				assert.Len(t, persona.ExpertiseAreas, 3)
				assert.Equal(t, "aws", persona.ExpertiseAreas[0].Domain)
			},
		},
		{
			name: "create with creative template",
			input: QuickCreatePersonaInput{
				Name:     "UX Designer",
				Template: "creative",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				assert.Equal(t, "creative", output["template"])

				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				persona, ok := elem.(*domain.Persona)
				require.True(t, ok)
				assert.Equal(t, "casual", persona.ResponseStyle.Formality)
			},
		},
		{
			name: "create with business template",
			input: QuickCreatePersonaInput{
				Name:     "Product Manager",
				Template: "business",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				persona, ok := elem.(*domain.Persona)
				require.True(t, ok)
				assert.Equal(t, "formal", persona.ResponseStyle.Formality)
			},
		},
		{
			name: "create with support template",
			input: QuickCreatePersonaInput{
				Name:     "Support Agent",
				Template: "support",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				persona, ok := elem.(*domain.Persona)
				require.True(t, ok)
				assert.Equal(t, "casual", persona.ResponseStyle.Formality)
				assert.NotEmpty(t, persona.BehavioralTraits)
			},
		},
		{
			name: "create with unknown template defaults to technical",
			input: QuickCreatePersonaInput{
				Name:     "Generic Expert",
				Template: "unknown",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				persona, ok := elem.(*domain.Persona)
				require.True(t, ok)
				assert.NotEmpty(t, persona.BehavioralTraits)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			_, output, err := server.handleQuickCreatePersona(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestQuickCreateSkill(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	cfg := &config.Config{
		DataDir:     tmpDir,
		StorageType: "file",
	}

	server := NewMCPServer("test-server", "1.0.0", repo, cfg)

	tests := []struct {
		name     string
		input    QuickCreateSkillInput
		wantErr  bool
		validate func(t *testing.T, output map[string]interface{})
	}{
		{
			name: "create with api template",
			input: QuickCreateSkillInput{
				Name:     "REST API Integration",
				Template: "api",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				assert.Equal(t, "REST API Integration", output["name"])
				assert.Equal(t, domain.SkillElement, output["type"])
				assert.Equal(t, "api", output["template"])

				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				skill, ok := elem.(*domain.Skill)
				require.True(t, ok)
				assert.NotEmpty(t, skill.Triggers)
				assert.NotEmpty(t, skill.Procedures)
			},
		},
		{
			name: "create with custom trigger",
			input: QuickCreateSkillInput{
				Name:        "Custom Skill",
				Description: "A custom skill",
				Trigger:     "when user needs help",
				Template:    "automation",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				skill, ok := elem.(*domain.Skill)
				require.True(t, ok)
				assert.Len(t, skill.Triggers, 1)
				assert.Equal(t, "when user needs help", skill.Triggers[0].Pattern)
			},
		},
		{
			name: "create with data template",
			input: QuickCreateSkillInput{
				Name:     "ETL Pipeline",
				Template: "data",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				assert.Equal(t, "data", output["template"])

				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				skill, ok := elem.(*domain.Skill)
				require.True(t, ok)
				assert.NotEmpty(t, skill.Procedures)
			},
		},
		{
			name: "create with analysis template",
			input: QuickCreateSkillInput{
				Name:     "Code Review",
				Template: "analysis",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				skill, ok := elem.(*domain.Skill)
				require.True(t, ok)
				assert.NotEmpty(t, skill.Triggers)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			_, output, err := server.handleQuickCreateSkill(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestQuickCreateMemory(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	cfg := &config.Config{
		DataDir:     tmpDir,
		StorageType: "file",
	}

	server := NewMCPServer("test-server", "1.0.0", repo, cfg)

	tests := []struct {
		name     string
		input    QuickCreateMemoryInput
		wantErr  bool
		validate func(t *testing.T, output map[string]interface{})
	}{
		{
			name: "create basic memory",
			input: QuickCreateMemoryInput{
				Name:    "Meeting Notes",
				Content: "Discussed project timeline and deliverables",
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				assert.Equal(t, "Meeting Notes", output["name"])
				assert.Equal(t, domain.MemoryElement, output["type"])
				assert.NotEmpty(t, output["hash"])

				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				memory, ok := elem.(*domain.Memory)
				require.True(t, ok)
				assert.Equal(t, "Discussed project timeline and deliverables", memory.Content)
				assert.NotEmpty(t, memory.ContentHash)
			},
		},
		{
			name: "create with tags",
			input: QuickCreateMemoryInput{
				Name:    "Bug Fix Notes",
				Content: "Fixed memory persistence issue by using type-specific handlers",
				Tags:    []string{"bug", "fix", "persistence"},
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				memory, ok := elem.(*domain.Memory)
				require.True(t, ok)

				metadata := memory.GetMetadata()
				assert.Contains(t, metadata.Tags, "bug")
				assert.Contains(t, metadata.Tags, "quick-create")
			},
		},
		{
			name: "create with importance",
			input: QuickCreateMemoryInput{
				Name:       "Critical Security Issue",
				Content:    "SQL injection vulnerability found in user input handler",
				Importance: "critical",
				Tags:       []string{"security", "vulnerability"},
			},
			wantErr: false,
			validate: func(t *testing.T, output map[string]interface{}) {
				id := output["id"].(string)
				elem, err := repo.GetByID(id)
				require.NoError(t, err)

				memory, ok := elem.(*domain.Memory)
				require.True(t, ok)

				metadata := memory.GetMetadata()
				assert.Contains(t, metadata.Tags, "importance:critical")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			_, output, err := server.handleQuickCreateMemory(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestPersonaTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	cfg := &config.Config{
		DataDir:     tmpDir,
		StorageType: "file",
	}

	server := NewMCPServer("test-server", "1.0.0", repo, cfg)

	templates := []string{"technical", "creative", "business", "support", "unknown"}

	for _, tmpl := range templates {
		t.Run("template_"+tmpl, func(t *testing.T) {
			template := server.getPersonaTemplate(tmpl)

			assert.NotEmpty(t, template.Name)
			assert.NotEmpty(t, template.Description)
			assert.NotEmpty(t, template.BehavioralTraits)
			assert.NotEmpty(t, template.Expertise)
			assert.NotEmpty(t, template.CommunicationTone)
			assert.NotEmpty(t, template.CommunicationFormality)
		})
	}
}

func TestSkillTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	cfg := &config.Config{
		DataDir:     tmpDir,
		StorageType: "file",
	}

	server := NewMCPServer("test-server", "1.0.0", repo, cfg)

	templates := []string{"api", "data", "automation", "analysis", "unknown"}

	for _, tmpl := range templates {
		t.Run("template_"+tmpl, func(t *testing.T) {
			template := server.getSkillTemplate(tmpl)

			assert.NotEmpty(t, template.Name)
			assert.NotEmpty(t, template.Description)
			assert.NotEmpty(t, template.Triggers)
			assert.NotEmpty(t, template.Procedures)
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("orDefault with value", func(t *testing.T) {
		result := orDefault("value", "default")
		assert.Equal(t, "value", result)
	})

	t.Run("orDefault with empty", func(t *testing.T) {
		result := orDefault("", "default")
		assert.Equal(t, "default", result)
	})

	t.Run("getCurrentUserFromContext with user", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userContextKey, "testuser")
		user := getCurrentUserFromContext(ctx)
		assert.Equal(t, "testuser", user)
	})

	t.Run("getCurrentUserFromContext without user", func(t *testing.T) {
		ctx := context.Background()
		user := getCurrentUserFromContext(ctx)
		assert.Equal(t, "system", user)
	})
}

func TestQuickCreateIndexing(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	cfg := &config.Config{
		DataDir:     tmpDir,
		StorageType: "file",
	}

	server := NewMCPServer("test-server", "1.0.0", repo, cfg)
	ctx := context.Background()

	// Create persona
	_, personaOutput, err := server.handleQuickCreatePersona(ctx, nil, QuickCreatePersonaInput{
		Name:     "Indexing Test Persona",
		Template: "technical",
	})
	require.NoError(t, err)
	personaID := personaOutput["id"].(string)

	// Create skill
	_, skillOutput, err := server.handleQuickCreateSkill(ctx, nil, QuickCreateSkillInput{
		Name:     "Indexing Test Skill",
		Template: "api",
	})
	require.NoError(t, err)
	skillID := skillOutput["id"].(string)

	// Wait for indexing
	time.Sleep(50 * time.Millisecond)

	// Verify items were created (indexing is verified indirectly)
	persona, err := repo.GetByID(personaID)
	require.NoError(t, err)
	assert.Equal(t, "Indexing Test Persona", persona.GetMetadata().Name)

	skill, err := repo.GetByID(skillID)
	require.NoError(t, err)
	assert.Equal(t, "Indexing Test Skill", skill.GetMetadata().Name)
}
