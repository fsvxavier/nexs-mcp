package domain

import (
	"errors"
	"math"
	"time"
)

// ConfidenceDecay manages time-based confidence decay for relationships and elements
type ConfidenceDecay struct {
	HalfLife          time.Duration          `json:"half_life"`          // Time for confidence to decay to 50%
	MinimumConfidence float64                `json:"minimum_confidence"` // Floor value (confidence doesn't decay below this)
	DecayFunction     DecayFunction          `json:"decay_function"`     // Type of decay curve
	ReferenceTime     time.Time              `json:"reference_time"`     // Time to calculate decay from (default: now)
	ReinforcementMap  map[string]int         `json:"reinforcement_map"`  // Track reinforcements per relationship/element ID
	Config            *ConfidenceDecayConfig `json:"config"`             // Advanced configuration
}

// DecayFunction defines the mathematical function for decay
type DecayFunction string

const (
	// DecayExponential - Standard exponential decay (most common)
	DecayExponential DecayFunction = "exponential"
	// DecayLinear - Linear decay over time
	DecayLinear DecayFunction = "linear"
	// DecayLogarithmic - Logarithmic decay (slower at first, then accelerates)
	DecayLogarithmic DecayFunction = "logarithmic"
	// DecayStep - Step function decay (discrete confidence levels)
	DecayStep DecayFunction = "step"
)

// ConfidenceDecayConfig provides advanced configuration for decay behavior
type ConfidenceDecayConfig struct {
	// EnableReinforcement - whether to apply reinforcement learning
	EnableReinforcement bool `json:"enable_reinforcement"`
	// ReinforcementBonus - confidence boost per reinforcement (0.0-1.0)
	ReinforcementBonus float64 `json:"reinforcement_bonus"`
	// ReinforcementDecay - how fast reinforcement effects decay
	ReinforcementDecay float64 `json:"reinforcement_decay"`
	// MaxConfidence - maximum confidence cap (default: 1.0)
	MaxConfidence float64 `json:"max_confidence"`
	// PreserveCritical - don't decay critical relationships
	PreserveCritical bool `json:"preserve_critical"`
	// CriticalThreshold - confidence threshold for "critical" (default: 0.9)
	CriticalThreshold float64 `json:"critical_threshold"`
	// StepIntervals - for step decay, define discrete intervals
	StepIntervals []time.Duration `json:"step_intervals,omitempty"`
	// StepValues - corresponding confidence values for each step
	StepValues []float64 `json:"step_values,omitempty"`
}

// DefaultConfidenceDecayConfig returns sensible defaults
func DefaultConfidenceDecayConfig() *ConfidenceDecayConfig {
	return &ConfidenceDecayConfig{
		EnableReinforcement: true,
		ReinforcementBonus:  0.1,  // 10% boost per reinforcement
		ReinforcementDecay:  0.95, // Reinforcement effects decay at 95% rate
		MaxConfidence:       1.0,
		PreserveCritical:    true,
		CriticalThreshold:   0.9,
		StepIntervals: []time.Duration{
			7 * 24 * time.Hour,  // 1 week
			30 * 24 * time.Hour, // 1 month
			90 * 24 * time.Hour, // 3 months
		},
		StepValues: []float64{0.9, 0.7, 0.5, 0.3},
	}
}

// NewConfidenceDecay creates a new confidence decay manager
func NewConfidenceDecay(halfLife time.Duration, minimumConfidence float64) *ConfidenceDecay {
	return &ConfidenceDecay{
		HalfLife:          halfLife,
		MinimumConfidence: minimumConfidence,
		DecayFunction:     DecayExponential, // Default to exponential
		ReferenceTime:     time.Now(),
		ReinforcementMap:  make(map[string]int),
		Config:            DefaultConfidenceDecayConfig(),
	}
}

// DefaultConfidenceDecay returns a confidence decay with sensible defaults
// Half-life: 30 days, Minimum: 0.1 (10%)
func DefaultConfidenceDecay() *ConfidenceDecay {
	return NewConfidenceDecay(30*24*time.Hour, 0.1)
}

// CalculateDecay computes the decayed confidence value for a given initial confidence and time
func (cd *ConfidenceDecay) CalculateDecay(initialConfidence float64, createdAt time.Time) (float64, error) {
	if initialConfidence < 0 || initialConfidence > 1 {
		return 0, errors.New("initial confidence must be between 0 and 1")
	}

	// Check if critical and preservation is enabled
	if cd.Config.PreserveCritical && initialConfidence >= cd.Config.CriticalThreshold {
		return initialConfidence, nil
	}

	// Calculate time elapsed
	elapsed := cd.ReferenceTime.Sub(createdAt)
	if elapsed < 0 {
		return initialConfidence, nil // Future date, no decay
	}

	// Apply decay function
	var decayedConfidence float64
	switch cd.DecayFunction {
	case DecayExponential:
		decayedConfidence = cd.exponentialDecay(initialConfidence, elapsed)
	case DecayLinear:
		decayedConfidence = cd.linearDecay(initialConfidence, elapsed)
	case DecayLogarithmic:
		decayedConfidence = cd.logarithmicDecay(initialConfidence, elapsed)
	case DecayStep:
		decayedConfidence = cd.stepDecay(initialConfidence, elapsed)
	default:
		decayedConfidence = cd.exponentialDecay(initialConfidence, elapsed)
	}

	// Apply minimum confidence floor
	if decayedConfidence < cd.MinimumConfidence {
		decayedConfidence = cd.MinimumConfidence
	}

	// Cap at maximum
	if decayedConfidence > cd.Config.MaxConfidence {
		decayedConfidence = cd.Config.MaxConfidence
	}

	return decayedConfidence, nil
}

// CalculateDecayWithReinforcement calculates decay with reinforcement learning applied
func (cd *ConfidenceDecay) CalculateDecayWithReinforcement(
	relationshipID string,
	initialConfidence float64,
	createdAt time.Time,
) (float64, error) {
	// First calculate base decay
	decayed, err := cd.CalculateDecay(initialConfidence, createdAt)
	if err != nil {
		return 0, err
	}

	// Apply reinforcement if enabled
	if !cd.Config.EnableReinforcement {
		return decayed, nil
	}

	reinforcements, exists := cd.ReinforcementMap[relationshipID]
	if !exists || reinforcements == 0 {
		return decayed, nil
	}

	// Calculate reinforcement boost
	// Each reinforcement adds ReinforcementBonus, but with diminishing returns
	totalBoost := 0.0
	for i := 0; i < reinforcements; i++ {
		// Diminishing returns: each reinforcement is worth less than the previous
		boost := cd.Config.ReinforcementBonus * math.Pow(cd.Config.ReinforcementDecay, float64(i))
		totalBoost += boost
	}

	// Apply boost to decayed confidence
	reinforcedConfidence := decayed + totalBoost

	// Cap at maximum
	if reinforcedConfidence > cd.Config.MaxConfidence {
		reinforcedConfidence = cd.Config.MaxConfidence
	}

	return reinforcedConfidence, nil
}

// Reinforce increases confidence for a relationship when it's reaffirmed
func (cd *ConfidenceDecay) Reinforce(relationshipID string) {
	if !cd.Config.EnableReinforcement {
		return
	}
	cd.ReinforcementMap[relationshipID]++
}

// GetReinforcementCount returns the number of reinforcements for a relationship
func (cd *ConfidenceDecay) GetReinforcementCount(relationshipID string) int {
	return cd.ReinforcementMap[relationshipID]
}

// ResetReinforcement resets reinforcement counter for a relationship
func (cd *ConfidenceDecay) ResetReinforcement(relationshipID string) {
	delete(cd.ReinforcementMap, relationshipID)
}

// exponentialDecay implements the standard exponential decay formula
// C(t) = C0 * (1/2)^(t/half_life)
func (cd *ConfidenceDecay) exponentialDecay(initialConfidence float64, elapsed time.Duration) float64 {
	if cd.HalfLife == 0 {
		return initialConfidence
	}

	// Convert to hours for calculation
	elapsedHours := elapsed.Hours()
	halfLifeHours := cd.HalfLife.Hours()

	// Calculate decay
	decayFactor := math.Pow(0.5, elapsedHours/halfLifeHours)
	return initialConfidence * decayFactor
}

// linearDecay implements linear decay over time
// C(t) = C0 * (1 - t/total_decay_time)
func (cd *ConfidenceDecay) linearDecay(initialConfidence float64, elapsed time.Duration) float64 {
	// Linear decay to minimum over 4x half-life
	totalDecayTime := cd.HalfLife * 4

	if elapsed >= totalDecayTime {
		return cd.MinimumConfidence
	}

	decayFraction := float64(elapsed) / float64(totalDecayTime)
	decayed := initialConfidence * (1.0 - decayFraction)

	if decayed < cd.MinimumConfidence {
		return cd.MinimumConfidence
	}

	return decayed
}

// logarithmicDecay implements logarithmic decay (slow at first, then faster)
// C(t) = C0 * (1 - log(1 + t/half_life) / log(N))
func (cd *ConfidenceDecay) logarithmicDecay(initialConfidence float64, elapsed time.Duration) float64 {
	if cd.HalfLife == 0 {
		return initialConfidence
	}

	// N determines the steepness (higher = slower decay)
	N := 10.0

	elapsedHours := elapsed.Hours()
	halfLifeHours := cd.HalfLife.Hours()

	// Logarithmic decay
	decayFactor := 1.0 - (math.Log(1.0+elapsedHours/halfLifeHours) / math.Log(N))
	if decayFactor < 0 {
		decayFactor = 0
	}

	return initialConfidence * decayFactor
}

// stepDecay implements step function decay (discrete levels)
func (cd *ConfidenceDecay) stepDecay(initialConfidence float64, elapsed time.Duration) float64 {
	if len(cd.Config.StepIntervals) == 0 || len(cd.Config.StepValues) == 0 {
		// Fallback to exponential if steps not configured
		return cd.exponentialDecay(initialConfidence, elapsed)
	}

	// Find which step we're in
	for i, interval := range cd.Config.StepIntervals {
		if elapsed < interval {
			if i < len(cd.Config.StepValues) {
				return initialConfidence * cd.Config.StepValues[i]
			}
		}
	}

	// Past all intervals, use last step value
	lastIndex := len(cd.Config.StepValues) - 1
	return initialConfidence * cd.Config.StepValues[lastIndex]
}

// ProjectFutureConfidence projects what confidence will be at a future time
func (cd *ConfidenceDecay) ProjectFutureConfidence(
	initialConfidence float64,
	createdAt time.Time,
	futureTime time.Time,
) (float64, error) {
	if futureTime.Before(cd.ReferenceTime) {
		return 0, errors.New("future time must be after reference time")
	}

	// Temporarily adjust reference time
	originalRef := cd.ReferenceTime
	cd.ReferenceTime = futureTime
	defer func() { cd.ReferenceTime = originalRef }()

	return cd.CalculateDecay(initialConfidence, createdAt)
}

// GetHalfLifeRemaining calculates how much time until confidence reaches 50% of current value
func (cd *ConfidenceDecay) GetHalfLifeRemaining(currentConfidence float64, createdAt time.Time) time.Duration {
	elapsed := cd.ReferenceTime.Sub(createdAt)

	switch cd.DecayFunction {
	case DecayExponential:
		// For exponential decay, half-life is constant
		return cd.HalfLife

	case DecayLinear:
		// Calculate remaining time to decay to 50% of current
		totalDecayTime := cd.HalfLife * 4
		remaining := totalDecayTime - elapsed
		if remaining < 0 {
			return 0
		}
		return remaining / 2

	default:
		// For other functions, use half-life as approximation
		return cd.HalfLife
	}
}

// ConfidenceStats provides statistics about decay behavior
type ConfidenceStats struct {
	TotalRelationships      int                    `json:"total_relationships"`
	ReinforcedRelationships int                    `json:"reinforced_relationships"`
	AverageReinforcements   float64                `json:"average_reinforcements"`
	HighConfidenceCount     int                    `json:"high_confidence_count"`   // >= 0.8
	MediumConfidenceCount   int                    `json:"medium_confidence_count"` // 0.5-0.8
	LowConfidenceCount      int                    `json:"low_confidence_count"`    // < 0.5
	TotalReinforcementBonus float64                `json:"total_reinforcement_bonus"`
	DecayConfiguration      map[string]interface{} `json:"decay_configuration"`
}

// GetStats returns statistics about confidence decay state
func (cd *ConfidenceDecay) GetStats() *ConfidenceStats {
	totalReinforcements := 0
	reinforcedCount := 0

	for _, count := range cd.ReinforcementMap {
		if count > 0 {
			reinforcedCount++
			totalReinforcements += count
		}
	}

	avgReinforcements := 0.0
	if reinforcedCount > 0 {
		avgReinforcements = float64(totalReinforcements) / float64(reinforcedCount)
	}

	return &ConfidenceStats{
		TotalRelationships:      len(cd.ReinforcementMap),
		ReinforcedRelationships: reinforcedCount,
		AverageReinforcements:   avgReinforcements,
		DecayConfiguration: map[string]interface{}{
			"half_life":          cd.HalfLife.String(),
			"minimum_confidence": cd.MinimumConfidence,
			"decay_function":     cd.DecayFunction,
			"reference_time":     cd.ReferenceTime,
			"config":             cd.Config,
		},
	}
}

// BatchCalculateDecay calculates decay for multiple items efficiently
func (cd *ConfidenceDecay) BatchCalculateDecay(items []DecayInput) ([]DecayOutput, error) {
	results := make([]DecayOutput, len(items))

	for i, item := range items {
		var decayed float64
		var err error

		if item.RelationshipID != "" && cd.Config.EnableReinforcement {
			decayed, err = cd.CalculateDecayWithReinforcement(
				item.RelationshipID,
				item.InitialConfidence,
				item.CreatedAt,
			)
		} else {
			decayed, err = cd.CalculateDecay(item.InitialConfidence, item.CreatedAt)
		}

		if err != nil {
			return nil, err
		}

		results[i] = DecayOutput{
			RelationshipID:    item.RelationshipID,
			InitialConfidence: item.InitialConfidence,
			DecayedConfidence: decayed,
			DecayAmount:       item.InitialConfidence - decayed,
			DecayPercentage:   ((item.InitialConfidence - decayed) / item.InitialConfidence) * 100,
			CreatedAt:         item.CreatedAt,
			ReferenceTime:     cd.ReferenceTime,
		}
	}

	return results, nil
}

// DecayInput represents input for batch decay calculation
type DecayInput struct {
	RelationshipID    string    `json:"relationship_id"`
	InitialConfidence float64   `json:"initial_confidence"`
	CreatedAt         time.Time `json:"created_at"`
}

// DecayOutput represents output of decay calculation
type DecayOutput struct {
	RelationshipID    string    `json:"relationship_id"`
	InitialConfidence float64   `json:"initial_confidence"`
	DecayedConfidence float64   `json:"decayed_confidence"`
	DecayAmount       float64   `json:"decay_amount"`
	DecayPercentage   float64   `json:"decay_percentage"`
	CreatedAt         time.Time `json:"created_at"`
	ReferenceTime     time.Time `json:"reference_time"`
}
