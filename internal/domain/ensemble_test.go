package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEnsemble(t *testing.T) {
	ensemble := NewEnsemble("test-ensemble", "Test Ensemble", "1.0", "author")

	assert.NotNil(t, ensemble)
	assert.Equal(t, "test-ensemble", ensemble.metadata.Name)
	assert.Equal(t, EnsembleElement, ensemble.metadata.Type)
	assert.Equal(t, "sequential", ensemble.ExecutionMode)
	assert.True(t, ensemble.metadata.IsActive)
}

func TestEnsemble_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Ensemble
		wantErr bool
	}{
		{
			name: "valid ensemble",
			setup: func() *Ensemble {
				ens := NewEnsemble("valid", "Valid Ensemble", "1.0", "author")
				ens.Members = []EnsembleMember{{AgentID: "agent1", Role: "leader", Priority: 1}}
				ens.AggregationStrategy = "vote"
				return ens
			},
			wantErr: false,
		},
		{
			name: "no members",
			setup: func() *Ensemble {
				ens := NewEnsemble("invalid", "Invalid Ensemble", "1.0", "author")
				ens.AggregationStrategy = "vote"
				return ens
			},
			wantErr: true,
		},
		{
			name: "invalid execution mode",
			setup: func() *Ensemble {
				ens := NewEnsemble("invalid", "Invalid Ensemble", "1.0", "author")
				ens.Members = []EnsembleMember{{AgentID: "agent1", Role: "leader", Priority: 1}}
				ens.ExecutionMode = "invalid"
				ens.AggregationStrategy = "vote"
				return ens
			},
			wantErr: true,
		},
		{
			name: "missing aggregation strategy",
			setup: func() *Ensemble {
				ens := NewEnsemble("invalid", "Invalid Ensemble", "1.0", "author")
				ens.Members = []EnsembleMember{{AgentID: "agent1", Role: "leader", Priority: 1}}
				ens.AggregationStrategy = ""
				return ens
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ens := tt.setup()
			err := ens.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnsemble_ActivateDeactivate(t *testing.T) {
	ens := NewEnsemble("test", "Test", "1.0", "author")

	assert.True(t, ens.IsActive())

	err := ens.Deactivate()
	assert.NoError(t, err)
	assert.False(t, ens.IsActive())

	err = ens.Activate()
	assert.NoError(t, err)
	assert.True(t, ens.IsActive())
}
