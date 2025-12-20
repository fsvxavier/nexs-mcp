package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSkill(t *testing.T) {
	skill := NewSkill("Test Skill", "A test skill", "1.0.0", "Test Author")

	assert.NotEmpty(t, skill.GetID())
	assert.Equal(t, SkillElement, skill.GetType())
	assert.True(t, skill.IsActive())
	assert.True(t, skill.Composable)
}

func TestSkill_AddTrigger(t *testing.T) {
	skill := NewSkill("Test", "Test", "1.0.0", "Test")

	tests := []struct {
		name        string
		trigger     SkillTrigger
		expectError bool
	}{
		{
			name: "keyword trigger",
			trigger: SkillTrigger{
				Type:     "keyword",
				Keywords: []string{"test", "example"},
			},
			expectError: false,
		},
		{
			name: "pattern trigger",
			trigger: SkillTrigger{
				Type:    "pattern",
				Pattern: "^test.*",
			},
			expectError: false,
		},
		{
			name: "invalid type",
			trigger: SkillTrigger{
				Type: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := skill.AddTrigger(tt.trigger)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSkill_Validate(t *testing.T) {
	t.Run("valid skill", func(t *testing.T) {
		skill := NewSkill("Test", "Test", "1.0.0", "Test")
		skill.AddTrigger(SkillTrigger{Type: "manual"})
		skill.AddProcedure(SkillProcedure{Step: 1, Action: "Do something"})

		err := skill.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing triggers", func(t *testing.T) {
		skill := NewSkill("Test", "Test", "1.0.0", "Test")
		skill.AddProcedure(SkillProcedure{Step: 1, Action: "Do something"})

		err := skill.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trigger")
	})

	t.Run("missing procedures", func(t *testing.T) {
		skill := NewSkill("Test", "Test", "1.0.0", "Test")
		skill.AddTrigger(SkillTrigger{Type: "manual"})

		err := skill.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "procedure")
	})
}
