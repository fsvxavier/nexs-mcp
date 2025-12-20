package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregateByConsensus(t *testing.T) {
	executor := &EnsembleExecutor{}

	tests := []struct {
		name            string
		results         []AgentResult
		config          ConsensusConfig
		expectError     bool
		expectConsensus bool
		minAgreement    float64
	}{
		{
			name: "unanimous_consensus",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "option_a"},
				{AgentID: "agent-2", Status: "success", Result: "option_a"},
				{AgentID: "agent-3", Status: "success", Result: "option_a"},
			},
			config: ConsensusConfig{
				Threshold: 0.7,
			},
			expectConsensus: true,
			minAgreement:    1.0,
		},
		{
			name: "majority_consensus",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "option_a"},
				{AgentID: "agent-2", Status: "success", Result: "option_a"},
				{AgentID: "agent-3", Status: "success", Result: "option_b"},
			},
			config: ConsensusConfig{
				Threshold: 0.6,
			},
			expectConsensus: true,
			minAgreement:    0.6,
		},
		{
			name: "no_consensus",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "option_a"},
				{AgentID: "agent-2", Status: "success", Result: "option_b"},
				{AgentID: "agent-3", Status: "success", Result: "option_c"},
			},
			config: ConsensusConfig{
				Threshold: 0.6,
			},
			expectConsensus: false,
		},
		{
			name: "quorum_not_met",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "option_a"},
				{AgentID: "agent-2", Status: "success", Result: "option_a"},
			},
			config: ConsensusConfig{
				Threshold:     0.7,
				RequireQuorum: true,
				QuorumSize:    3,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.aggregateByConsensus(tt.results, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectConsensus, result.ReachedConsensus)

			if tt.expectConsensus {
				assert.GreaterOrEqual(t, result.AgreementLevel, tt.minAgreement)
				assert.NotNil(t, result.Value)
			}
		})
	}
}

func TestAggregateByVoting(t *testing.T) {
	executor := &EnsembleExecutor{}

	tests := []struct {
		name           string
		results        []AgentResult
		config         VotingConfig
		expectError    bool
		expectedWinner interface{}
		minPercentage  float64
	}{
		{
			name: "simple_majority",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "option_a"},
				{AgentID: "agent-2", Status: "success", Result: "option_a"},
				{AgentID: "agent-3", Status: "success", Result: "option_b"},
			},
			config: VotingConfig{
				MinimumVotes: 2,
			},
			expectedWinner: "option_a",
			minPercentage:  60.0,
		},
		{
			name: "weighted_by_priority",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "option_a", Metadata: map[string]interface{}{"priority": 10}},
				{AgentID: "agent-2", Status: "success", Result: "option_b", Metadata: map[string]interface{}{"priority": 5}},
				{AgentID: "agent-3", Status: "success", Result: "option_b", Metadata: map[string]interface{}{"priority": 5}},
			},
			config: VotingConfig{
				WeightByPriority: true,
				MinimumVotes:     2,
			},
			expectedWinner: "option_a", // Higher priority agent wins even with fewer votes
		},
		{
			name: "insufficient_votes",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "option_a"},
			},
			config: VotingConfig{
				MinimumVotes: 2,
			},
			expectError: true,
		},
		{
			name: "custom_weights",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "option_a"},
				{AgentID: "agent-2", Status: "success", Result: "option_b"},
			},
			config: VotingConfig{
				MinimumVotes: 1,
				CustomWeights: map[string]float64{
					"agent-1": 2.0,
					"agent-2": 1.0,
				},
			},
			expectedWinner: "option_a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.aggregateByVoting(tt.results, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result.Winner)

			if tt.expectedWinner != nil {
				assert.Equal(t, tt.expectedWinner, result.Winner)
			}

			if tt.minPercentage > 0 {
				assert.GreaterOrEqual(t, result.Percentage, tt.minPercentage)
			}
		})
	}
}

func TestWeightedConsensus(t *testing.T) {
	executor := &EnsembleExecutor{}

	results := []AgentResult{
		{AgentID: "agent-1", Status: "success", Result: "answer_a", Metadata: map[string]interface{}{"confidence": 0.9}},
		{AgentID: "agent-2", Status: "success", Result: "answer_a", Metadata: map[string]interface{}{"confidence": 0.8}},
		{AgentID: "agent-3", Status: "success", Result: "answer_b", Metadata: map[string]interface{}{"confidence": 0.5}},
	}

	result, err := executor.aggregateByWeightedConsensus(results, 0.6)

	require.NoError(t, err)
	assert.True(t, result.ReachedConsensus)
	assert.Equal(t, "answer_a", result.Value)
	assert.NotNil(t, result.ConfidenceScores)
	assert.Len(t, result.ConfidenceScores, 3)
}

func TestThresholdConsensus(t *testing.T) {
	executor := &EnsembleExecutor{}

	tests := []struct {
		name            string
		results         []AgentResult
		threshold       float64
		quorum          int
		expectThreshold bool
	}{
		{
			name: "threshold_met",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "result_a"},
				{AgentID: "agent-2", Status: "success", Result: "result_a"},
				{AgentID: "agent-3", Status: "success", Result: "result_a"},
				{AgentID: "agent-4", Status: "success", Result: "result_b"},
			},
			threshold:       0.7,
			quorum:          2,
			expectThreshold: true,
		},
		{
			name: "threshold_not_met",
			results: []AgentResult{
				{AgentID: "agent-1", Status: "success", Result: "result_a"},
				{AgentID: "agent-2", Status: "success", Result: "result_b"},
				{AgentID: "agent-3", Status: "success", Result: "result_c"},
			},
			threshold:       0.7,
			quorum:          2,
			expectThreshold: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := executor.aggregateByThresholdConsensus(tt.results, tt.threshold, tt.quorum)

			assert.Equal(t, tt.expectThreshold, result.ThresholdMet)
			assert.Equal(t, tt.threshold, result.RequiredThreshold)
		})
	}
}

func TestGroupSimilarResults(t *testing.T) {
	executor := &EnsembleExecutor{}

	results := []AgentResult{
		{AgentID: "agent-1", Result: "answer_a"},
		{AgentID: "agent-2", Result: "answer_a"},
		{AgentID: "agent-3", Result: "answer_b"},
		{AgentID: "agent-4", Result: map[string]interface{}{"key": "value"}},
		{AgentID: "agent-5", Result: map[string]interface{}{"key": "value"}},
	}

	groups := executor.groupSimilarResults(results)

	// Should have 3 groups: "answer_a", "answer_b", and the map
	assert.Len(t, groups, 3)

	// Find group with "answer_a"
	var foundGroupA bool
	for _, group := range groups {
		if len(group) == 2 && group[0].Result == "answer_a" {
			foundGroupA = true
			break
		}
	}
	assert.True(t, foundGroupA, "Should find group with 2 'answer_a' results")
}

func TestResultToKey(t *testing.T) {
	executor := &EnsembleExecutor{}

	tests := []struct {
		name     string
		result   interface{}
		expected string
	}{
		{
			name:     "string",
			result:   "test",
			expected: `"test"`,
		},
		{
			name:     "number",
			result:   42,
			expected: "42",
		},
		{
			name:     "map",
			result:   map[string]interface{}{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "array",
			result:   []string{"a", "b", "c"},
			expected: `["a","b","c"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := executor.resultToKey(tt.result)
			assert.Equal(t, tt.expected, key)
		})
	}
}

func TestBreakTie(t *testing.T) {
	executor := &EnsembleExecutor{}

	// Note: resultToKey converts results to JSON strings
	votes := map[string]float64{
		`"result_a"`: 2.0,
		`"result_b"`: 2.0,
	}

	results := map[string]interface{}{
		`"result_a"`: "result_a",
		`"result_b"`: "result_b",
	}

	agentResults := []AgentResult{
		{AgentID: "agent-1", Result: "result_a", Metadata: map[string]interface{}{"priority": 8}},
		{AgentID: "agent-2", Result: "result_b", Metadata: map[string]interface{}{"priority": 5}},
	}

	tests := []struct {
		name     string
		strategy string
	}{
		{name: "first_strategy", strategy: "first"},
		{name: "highest_priority", strategy: "highest_priority"},
		{name: "random", strategy: "random"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winner := executor.breakTie(votes, results, agentResults, tt.strategy)
			assert.NotEmpty(t, winner)
			assert.Contains(t, []string{`"result_a"`, `"result_b"`}, winner)
		})
	}
}

func TestVotingTieBreaker(t *testing.T) {
	executor := &EnsembleExecutor{}

	results := []AgentResult{
		{AgentID: "agent-1", Status: "success", Result: "option_a", Metadata: map[string]interface{}{"priority": 9}},
		{AgentID: "agent-2", Status: "success", Result: "option_b", Metadata: map[string]interface{}{"priority": 7}},
	}

	config := VotingConfig{
		MinimumVotes: 1,
		TieBreaker:   "highest_priority",
	}

	result, err := executor.aggregateByVoting(results, config)

	require.NoError(t, err)
	assert.True(t, result.TieBreaker)
	assert.Equal(t, "option_a", result.Winner) // Higher priority should win
}

func TestConsensusWithNoSuccessfulResults(t *testing.T) {
	executor := &EnsembleExecutor{}

	results := []AgentResult{
		{AgentID: "agent-1", Status: "failed", Error: "error1"},
		{AgentID: "agent-2", Status: "failed", Error: "error2"},
	}

	config := ConsensusConfig{
		Threshold: 0.7,
	}

	_, err := executor.aggregateByConsensus(results, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no successful results")
}

func TestVotingWithConfidenceWeight(t *testing.T) {
	executor := &EnsembleExecutor{}

	results := []AgentResult{
		{AgentID: "agent-1", Status: "success", Result: "option_a", Metadata: map[string]interface{}{"confidence": 0.9}},
		{AgentID: "agent-2", Status: "success", Result: "option_b", Metadata: map[string]interface{}{"confidence": 0.3}},
		{AgentID: "agent-3", Status: "success", Result: "option_b", Metadata: map[string]interface{}{"confidence": 0.3}},
	}

	config := VotingConfig{
		WeightByConfidence: true,
		MinimumVotes:       1,
	}

	result, err := executor.aggregateByVoting(results, config)

	require.NoError(t, err)
	// Agent-1 with high confidence should outweigh the other two
	assert.Equal(t, "option_a", result.Winner)
}

func TestResultEquals(t *testing.T) {
	executor := &EnsembleExecutor{}

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "equal_strings",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "different_strings",
			a:        "test1",
			b:        "test2",
			expected: false,
		},
		{
			name:     "equal_maps",
			a:        map[string]interface{}{"key": "value"},
			b:        map[string]interface{}{"key": "value"},
			expected: true,
		},
		{
			name:     "different_maps",
			a:        map[string]interface{}{"key": "value1"},
			b:        map[string]interface{}{"key": "value2"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.resultEquals(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
