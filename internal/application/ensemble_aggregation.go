package application

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// ConsensusConfig configures consensus behavior
type ConsensusConfig struct {
	Threshold      float64 // Minimum agreement percentage (0.0 to 1.0)
	RequireQuorum  bool    // Require minimum number of participants
	QuorumSize     int     // Minimum number of participants if RequireQuorum is true
	WeightedVoting bool    // Use agent priority as weight
}

// VotingConfig configures voting behavior
type VotingConfig struct {
	WeightByPriority   bool               // Use agent priority as vote weight
	WeightByConfidence bool               // Use result confidence scores
	MinimumVotes       int                // Minimum votes required
	TieBreaker         string             // "first", "random", "highest_priority"
	CustomWeights      map[string]float64 // Custom weights per agent ID
}

// ConsensusResult represents the result of consensus algorithm
type ConsensusResult struct {
	Value            interface{}            `json:"value"`
	AgreementLevel   float64                `json:"agreement_level"` // 0.0 to 1.0
	Participants     int                    `json:"participants"`
	Supporting       []string               `json:"supporting"`            // Agent IDs that support this result
	Alternative      map[string]interface{} `json:"alternative,omitempty"` // Alternative results if no strong consensus
	ReachedConsensus bool                   `json:"reached_consensus"`
}

// VotingResult represents the result of voting algorithm
type VotingResult struct {
	Winner      interface{}        `json:"winner"`
	TotalVotes  float64            `json:"total_votes"`
	WinnerVotes float64            `json:"winner_votes"`
	Percentage  float64            `json:"percentage"`
	Voters      []string           `json:"voters"`
	Breakdown   map[string]float64 `json:"breakdown"` // votes per option
	TieBreaker  bool               `json:"tie_breaker,omitempty"`
}

// aggregateByConsensus implements advanced consensus algorithm
func (e *EnsembleExecutor) aggregateByConsensus(results []AgentResult, config ConsensusConfig) (*ConsensusResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results for consensus")
	}

	// Filter successful results
	successResults := make([]AgentResult, 0)
	for _, r := range results {
		if r.Status == "success" && r.Result != nil {
			successResults = append(successResults, r)
		}
	}

	if len(successResults) == 0 {
		return nil, fmt.Errorf("no successful results for consensus")
	}

	// Check quorum
	if config.RequireQuorum && len(successResults) < config.QuorumSize {
		return &ConsensusResult{
			Participants:     len(successResults),
			ReachedConsensus: false,
		}, fmt.Errorf("quorum not met: got %d, required %d", len(successResults), config.QuorumSize)
	}

	// Group results by similarity
	resultGroups := e.groupSimilarResults(successResults)

	// Find largest group
	var largestGroup []AgentResult
	var largestGroupKey string
	for key, group := range resultGroups {
		if len(group) > len(largestGroup) {
			largestGroup = group
			largestGroupKey = key
		}
	}

	// Calculate agreement level
	var totalWeight float64
	var supportingWeight float64
	supportingAgents := make([]string, 0)

	for _, result := range successResults {
		weight := 1.0
		if config.WeightedVoting {
			// Use agent metadata priority as weight (default 1.0 if not set)
			if priority, ok := result.Metadata["priority"].(int); ok {
				weight = float64(priority) / 10.0 // Normalize priority to 0.1-1.0
				if weight <= 0 {
					weight = 0.1
				}
			}
		}
		totalWeight += weight
	}

	for _, result := range largestGroup {
		weight := 1.0
		if config.WeightedVoting {
			if priority, ok := result.Metadata["priority"].(int); ok {
				weight = float64(priority) / 10.0
				if weight <= 0 {
					weight = 0.1
				}
			}
		}
		supportingWeight += weight
		supportingAgents = append(supportingAgents, result.AgentID)
	}

	agreementLevel := supportingWeight / totalWeight
	reachedConsensus := agreementLevel >= config.Threshold

	consensusResult := &ConsensusResult{
		Value:            largestGroup[0].Result,
		AgreementLevel:   agreementLevel,
		Participants:     len(successResults),
		Supporting:       supportingAgents,
		ReachedConsensus: reachedConsensus,
	}

	// Add alternatives if consensus not reached
	if !reachedConsensus && len(resultGroups) > 1 {
		consensusResult.Alternative = make(map[string]interface{})
		for key, group := range resultGroups {
			if key != largestGroupKey {
				consensusResult.Alternative[key] = group[0].Result
			}
		}
	}

	return consensusResult, nil
}

// aggregateByVoting implements advanced voting algorithm
func (e *EnsembleExecutor) aggregateByVoting(results []AgentResult, config VotingConfig) (*VotingResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results for voting")
	}

	// Filter successful results
	successResults := make([]AgentResult, 0)
	for _, r := range results {
		if r.Status == "success" && r.Result != nil {
			successResults = append(successResults, r)
		}
	}

	if len(successResults) == 0 {
		return nil, fmt.Errorf("no successful results for voting")
	}

	if len(successResults) < config.MinimumVotes {
		return nil, fmt.Errorf("insufficient votes: got %d, required %d", len(successResults), config.MinimumVotes)
	}

	// Group results and calculate votes
	voteCounts := make(map[string]float64)
	voteResults := make(map[string]interface{})
	voters := make([]string, 0)
	var totalVotes float64

	for _, result := range successResults {
		// Calculate vote weight
		weight := 1.0

		// Apply custom weight if defined
		if customWeight, exists := config.CustomWeights[result.AgentID]; exists {
			weight = customWeight
		} else {
			// Apply priority-based weight
			if config.WeightByPriority {
				if priority, ok := result.Metadata["priority"].(int); ok {
					weight = float64(priority) / 10.0
					if weight <= 0 {
						weight = 0.1
					}
				}
			}

			// Apply confidence-based weight
			if config.WeightByConfidence {
				if confidence, ok := result.Metadata["confidence"].(float64); ok {
					weight *= confidence
				}
			}
		}

		// Convert result to comparable key
		resultKey := e.resultToKey(result.Result)
		voteCounts[resultKey] += weight
		voteResults[resultKey] = result.Result
		voters = append(voters, result.AgentID)
		totalVotes += weight
	}

	// Find winner
	var winnerKey string
	var winnerVotes float64
	tied := false

	for key, votes := range voteCounts {
		if votes > winnerVotes {
			winnerKey = key
			winnerVotes = votes
			tied = false
		} else if votes == winnerVotes {
			tied = true
		}
	}

	// Handle tie
	if tied {
		winnerKey = e.breakTie(voteCounts, voteResults, successResults, config.TieBreaker)
	}

	return &VotingResult{
		Winner:      voteResults[winnerKey],
		TotalVotes:  totalVotes,
		WinnerVotes: winnerVotes,
		Percentage:  (winnerVotes / totalVotes) * 100,
		Voters:      voters,
		Breakdown:   voteCounts,
		TieBreaker:  tied,
	}, nil
}

// groupSimilarResults groups results that are similar or equal
func (e *EnsembleExecutor) groupSimilarResults(results []AgentResult) map[string][]AgentResult {
	groups := make(map[string][]AgentResult)

	for _, result := range results {
		key := e.resultToKey(result.Result)
		groups[key] = append(groups[key], result)
	}

	return groups
}

// resultToKey converts a result to a comparable string key
func (e *EnsembleExecutor) resultToKey(result interface{}) string {
	// Try JSON serialization for complex types
	if jsonBytes, err := json.Marshal(result); err == nil {
		return string(jsonBytes)
	}

	// Fallback to string representation
	return fmt.Sprintf("%v", result)
}

// breakTie resolves voting ties based on strategy
func (e *EnsembleExecutor) breakTie(votes map[string]float64, results map[string]interface{}, agentResults []AgentResult, strategy string) string {
	switch strategy {
	case "first":
		// Return first result in original order
		if len(agentResults) > 0 {
			return e.resultToKey(agentResults[0].Result)
		}

	case "highest_priority":
		// Return result from highest priority agent
		highestPriority := -1
		var winnerKey string

		for _, result := range agentResults {
			priority := 5 // default priority
			if p, ok := result.Metadata["priority"].(int); ok {
				priority = p
			}

			if priority > highestPriority {
				highestPriority = priority
				winnerKey = e.resultToKey(result.Result)
			}
		}

		if winnerKey != "" {
			return winnerKey
		}

	case "random":
		// Return random tied result
		for key, voteCount := range votes {
			maxVotes := 0.0
			for _, v := range votes {
				if v > maxVotes {
					maxVotes = v
				}
			}
			if voteCount == maxVotes {
				return key // Return first found (pseudo-random in map iteration)
			}
		}
	}

	// Default: return first key
	for key := range votes {
		return key
	}

	return ""
}

// WeightedConsensusResult represents weighted consensus with confidence scores
type WeightedConsensusResult struct {
	ConsensusResult
	WeightedAgreement float64            `json:"weighted_agreement"`
	ConfidenceScores  map[string]float64 `json:"confidence_scores"`
}

// aggregateByWeightedConsensus implements consensus with confidence weighting
func (e *EnsembleExecutor) aggregateByWeightedConsensus(results []AgentResult, threshold float64) (*WeightedConsensusResult, error) {
	config := ConsensusConfig{
		Threshold:      threshold,
		WeightedVoting: true,
	}

	baseConsensus, err := e.aggregateByConsensus(results, config)
	if err != nil {
		return nil, err
	}

	// Calculate confidence scores
	confidenceScores := make(map[string]float64)
	for _, result := range results {
		if result.Status == "success" {
			confidence := 1.0
			if conf, ok := result.Metadata["confidence"].(float64); ok {
				confidence = conf
			}
			confidenceScores[result.AgentID] = confidence
		}
	}

	return &WeightedConsensusResult{
		ConsensusResult:   *baseConsensus,
		WeightedAgreement: baseConsensus.AgreementLevel,
		ConfidenceScores:  confidenceScores,
	}, nil
}

// ThresholdConsensusResult represents consensus requiring minimum threshold
type ThresholdConsensusResult struct {
	ConsensusResult
	ThresholdMet      bool    `json:"threshold_met"`
	RequiredThreshold float64 `json:"required_threshold"`
}

// aggregateByThresholdConsensus requires minimum agreement threshold
func (e *EnsembleExecutor) aggregateByThresholdConsensus(results []AgentResult, threshold float64, quorum int) (*ThresholdConsensusResult, error) {
	config := ConsensusConfig{
		Threshold:     threshold,
		RequireQuorum: quorum > 0,
		QuorumSize:    quorum,
	}

	baseConsensus, err := e.aggregateByConsensus(results, config)

	thresholdMet := err == nil && baseConsensus.ReachedConsensus

	result := &ThresholdConsensusResult{
		ThresholdMet:      thresholdMet,
		RequiredThreshold: threshold,
	}

	if baseConsensus != nil {
		result.ConsensusResult = *baseConsensus
	}

	return result, err
}

// resultEquals checks if two results are equal
func (e *EnsembleExecutor) resultEquals(a, b interface{}) bool {
	// Try direct equality
	if reflect.DeepEqual(a, b) {
		return true
	}

	// Try JSON comparison for complex types
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)

	if aErr == nil && bErr == nil {
		return string(aJSON) == string(bJSON)
	}

	return false
}
