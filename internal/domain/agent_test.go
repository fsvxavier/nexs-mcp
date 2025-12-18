package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	agent := NewAgent("test-agent", "Test Agent", "1.0", "author")

	assert.NotNil(t, agent)
	assert.Equal(t, "test-agent", agent.metadata.Name)
	assert.Equal(t, AgentElement, agent.metadata.Type)
	assert.Equal(t, 10, agent.MaxIterations)
	assert.True(t, agent.metadata.IsActive)
}

func TestAgent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Agent
		wantErr bool
	}{
		{
			name: "valid agent",
			setup: func() *Agent {
				agent := NewAgent("valid", "Valid Agent", "1.0", "author")
				agent.Goals = []string{"goal1"}
				agent.Actions = []AgentAction{{Name: "action1", Type: "tool"}}
				return agent
			},
			wantErr: false,
		},
		{
			name: "no goals",
			setup: func() *Agent {
				agent := NewAgent("invalid", "Invalid Agent", "1.0", "author")
				agent.Actions = []AgentAction{{Name: "action1", Type: "tool"}}
				return agent
			},
			wantErr: true,
		},
		{
			name: "no actions",
			setup: func() *Agent {
				agent := NewAgent("invalid", "Invalid Agent", "1.0", "author")
				agent.Goals = []string{"goal1"}
				return agent
			},
			wantErr: true,
		},
		{
			name: "invalid max_iterations",
			setup: func() *Agent {
				agent := NewAgent("invalid", "Invalid Agent", "1.0", "author")
				agent.Goals = []string{"goal1"}
				agent.Actions = []AgentAction{{Name: "action1", Type: "tool"}}
				agent.MaxIterations = 200
				return agent
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := tt.setup()
			err := agent.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAgent_ActivateDeactivate(t *testing.T) {
	agent := NewAgent("test", "Test", "1.0", "author")

	assert.True(t, agent.IsActive())

	err := agent.Deactivate()
	assert.NoError(t, err)
	assert.False(t, agent.IsActive())

	err = agent.Activate()
	assert.NoError(t, err)
	assert.True(t, agent.IsActive())
}
