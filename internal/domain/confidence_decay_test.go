package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfidenceDecay(t *testing.T) {
	halfLife := 30 * 24 * time.Hour
	minConfidence := 0.1

	cd := NewConfidenceDecay(halfLife, minConfidence)

	assert.NotNil(t, cd)
	assert.Equal(t, halfLife, cd.HalfLife)
	assert.Equal(t, minConfidence, cd.MinimumConfidence)
	assert.Equal(t, DecayExponential, cd.DecayFunction)
	assert.NotNil(t, cd.ReinforcementMap)
	assert.NotNil(t, cd.Config)
}

func TestDefaultConfidenceDecay(t *testing.T) {
	cd := DefaultConfidenceDecay()

	assert.NotNil(t, cd)
	assert.Equal(t, 30*24*time.Hour, cd.HalfLife)
	assert.Equal(t, 0.1, cd.MinimumConfidence)
	assert.Equal(t, DecayExponential, cd.DecayFunction)
}

func TestConfidenceDecay_ExponentialDecay(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.Config.PreserveCritical = false // Disable to test pure decay
	cd.ReferenceTime = time.Now()

	tests := []struct {
		name              string
		initialConfidence float64
		elapsed           time.Duration
		minExpected       float64
		maxExpected       float64
	}{
		{
			name:              "no decay at t=0",
			initialConfidence: 1.0,
			elapsed:           0,
			minExpected:       0.99,
			maxExpected:       1.0,
		},
		{
			name:              "half decay at half-life",
			initialConfidence: 1.0,
			elapsed:           30 * 24 * time.Hour,
			minExpected:       0.49,
			maxExpected:       0.51,
		},
		{
			name:              "quarter decay at 2x half-life",
			initialConfidence: 1.0,
			elapsed:           60 * 24 * time.Hour,
			minExpected:       0.24,
			maxExpected:       0.26,
		},
		{
			name:              "respect minimum confidence",
			initialConfidence: 1.0,
			elapsed:           365 * 24 * time.Hour, // 1 year
			minExpected:       0.1,
			maxExpected:       0.11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdAt := cd.ReferenceTime.Add(-tt.elapsed)
			decayed, err := cd.CalculateDecay(tt.initialConfidence, createdAt)

			require.NoError(t, err)
			assert.GreaterOrEqual(t, decayed, tt.minExpected)
			assert.LessOrEqual(t, decayed, tt.maxExpected)
		})
	}
}

func TestConfidenceDecay_LinearDecay(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.Config.PreserveCritical = false // Disable to test pure decay
	cd.DecayFunction = DecayLinear
	cd.ReferenceTime = time.Now()

	initialConfidence := 1.0
	halfLife := cd.HalfLife

	// Linear decay over 4x half-life
	createdAt := cd.ReferenceTime.Add(-2 * halfLife) // Halfway through decay period
	decayed, err := cd.CalculateDecay(initialConfidence, createdAt)

	require.NoError(t, err)
	assert.Greater(t, decayed, 0.4) // Should be around 50% decayed
	assert.Less(t, decayed, 0.6)
}

func TestConfidenceDecay_LogarithmicDecay(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.Config.PreserveCritical = false // Disable to test pure decay
	cd.DecayFunction = DecayLogarithmic
	cd.ReferenceTime = time.Now()

	initialConfidence := 1.0

	// Test that logarithmic decay is slower initially
	earlyTime := cd.ReferenceTime.Add(-7 * 24 * time.Hour) // 1 week
	earlyDecay, err := cd.CalculateDecay(initialConfidence, earlyTime)
	require.NoError(t, err)

	// Should still be relatively high (slow initial decay)
	assert.Greater(t, earlyDecay, 0.8)
}

func TestConfidenceDecay_StepDecay(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.Config.PreserveCritical = false // Disable to test pure decay
	cd.DecayFunction = DecayStep
	cd.ReferenceTime = time.Now()

	// Default config has step intervals
	initialConfidence := 1.0

	// At 3 days (before first interval of 7 days)
	early := cd.ReferenceTime.Add(-3 * 24 * time.Hour)
	earlyDecay, err := cd.CalculateDecay(initialConfidence, early)
	require.NoError(t, err)
	assert.Greater(t, earlyDecay, 0.85) // First step value is 0.9

	// At 15 days (after first interval, before second of 30 days)
	mid := cd.ReferenceTime.Add(-15 * 24 * time.Hour)
	midDecay, err := cd.CalculateDecay(initialConfidence, mid)
	require.NoError(t, err)
	assert.Greater(t, midDecay, 0.65) // Should be at second step
	assert.Less(t, midDecay, 0.75)
}

func TestConfidenceDecay_PreserveCritical(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.Config.PreserveCritical = true
	cd.Config.CriticalThreshold = 0.9
	cd.ReferenceTime = time.Now()

	// Critical confidence should not decay
	criticalConfidence := 0.95
	oldTime := cd.ReferenceTime.Add(-365 * 24 * time.Hour) // 1 year ago

	decayed, err := cd.CalculateDecay(criticalConfidence, oldTime)
	require.NoError(t, err)
	assert.Equal(t, criticalConfidence, decayed)

	// Non-critical should decay
	normalConfidence := 0.8
	decayed2, err := cd.CalculateDecay(normalConfidence, oldTime)
	require.NoError(t, err)
	assert.Less(t, decayed2, normalConfidence)
}

func TestConfidenceDecay_Reinforce(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.ReferenceTime = time.Now()

	relationshipID := "rel-123"

	// Initially no reinforcement
	assert.Equal(t, 0, cd.GetReinforcementCount(relationshipID))

	// Add reinforcements
	cd.Reinforce(relationshipID)
	assert.Equal(t, 1, cd.GetReinforcementCount(relationshipID))

	cd.Reinforce(relationshipID)
	cd.Reinforce(relationshipID)
	assert.Equal(t, 3, cd.GetReinforcementCount(relationshipID))

	// Reset
	cd.ResetReinforcement(relationshipID)
	assert.Equal(t, 0, cd.GetReinforcementCount(relationshipID))
}

func TestConfidenceDecay_CalculateDecayWithReinforcement(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.ReferenceTime = time.Now()

	relationshipID := "rel-123"
	initialConfidence := 0.8
	createdAt := cd.ReferenceTime.Add(-30 * 24 * time.Hour) // At half-life

	// Without reinforcement
	decayedNoReinforce, err := cd.CalculateDecayWithReinforcement(relationshipID, initialConfidence, createdAt)
	require.NoError(t, err)

	// With reinforcement
	cd.Reinforce(relationshipID)
	cd.Reinforce(relationshipID)
	decayedWithReinforce, err := cd.CalculateDecayWithReinforcement(relationshipID, initialConfidence, createdAt)
	require.NoError(t, err)

	// Reinforced confidence should be higher
	assert.Greater(t, decayedWithReinforce, decayedNoReinforce)
	assert.LessOrEqual(t, decayedWithReinforce, cd.Config.MaxConfidence)
}

func TestConfidenceDecay_ReinforcementDiminishingReturns(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.ReferenceTime = time.Now()

	relationshipID := "rel-123"
	initialConfidence := 0.5
	createdAt := cd.ReferenceTime.Add(-60 * 24 * time.Hour) // 2x half-life

	// Add multiple reinforcements
	var previousBoost float64
	for i := 1; i <= 5; i++ {
		cd.Reinforce(relationshipID)
		reinforced, err := cd.CalculateDecayWithReinforcement(relationshipID, initialConfidence, createdAt)
		require.NoError(t, err)

		baseDecay, _ := cd.CalculateDecay(initialConfidence, createdAt)
		currentBoost := reinforced - baseDecay

		if i > 1 {
			// Each additional reinforcement should provide less boost (diminishing returns)
			assert.Less(t, currentBoost-previousBoost, previousBoost)
		}
		previousBoost = currentBoost
	}
}

func TestConfidenceDecay_BatchCalculateDecay(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.Config.PreserveCritical = false // Disable to test decay
	cd.ReferenceTime = time.Now()

	// Prepare batch inputs
	inputs := []DecayInput{
		{
			RelationshipID:    "rel-1",
			InitialConfidence: 1.0,
			CreatedAt:         cd.ReferenceTime.Add(-30 * 24 * time.Hour),
		},
		{
			RelationshipID:    "rel-2",
			InitialConfidence: 0.8,
			CreatedAt:         cd.ReferenceTime.Add(-60 * 24 * time.Hour),
		},
		{
			RelationshipID:    "rel-3",
			InitialConfidence: 0.6,
			CreatedAt:         cd.ReferenceTime.Add(-10 * 24 * time.Hour),
		},
	}

	// Add some reinforcements
	cd.Reinforce("rel-1")
	cd.Reinforce("rel-1")

	outputs, err := cd.BatchCalculateDecay(inputs)
	require.NoError(t, err)
	assert.Len(t, outputs, 3)

	// Verify each output
	for i, output := range outputs {
		assert.Equal(t, inputs[i].RelationshipID, output.RelationshipID)
		assert.Equal(t, inputs[i].InitialConfidence, output.InitialConfidence)
		assert.Less(t, output.DecayedConfidence, inputs[i].InitialConfidence)
		assert.GreaterOrEqual(t, output.DecayAmount, 0.0)
		assert.GreaterOrEqual(t, output.DecayPercentage, 0.0)
	}

	// rel-1 should have reinforcement boost
	assert.Greater(t, outputs[0].DecayedConfidence, 0.5)
}

func TestConfidenceDecay_ProjectFutureConfidence(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.Config.PreserveCritical = false // Disable to test decay
	cd.ReferenceTime = time.Now()

	initialConfidence := 1.0
	createdAt := cd.ReferenceTime.Add(-15 * 24 * time.Hour)

	// Project 15 days into future (total 30 days = 1 half-life)
	futureTime := cd.ReferenceTime.Add(15 * 24 * time.Hour)
	projected, err := cd.ProjectFutureConfidence(initialConfidence, createdAt, futureTime)

	require.NoError(t, err)
	assert.Greater(t, projected, 0.49)
	assert.Less(t, projected, 0.51) // Should be around 0.5
}

func TestConfidenceDecay_GetHalfLifeRemaining(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.ReferenceTime = time.Now()

	currentConfidence := 0.8
	createdAt := cd.ReferenceTime.Add(-15 * 24 * time.Hour)

	remaining := cd.GetHalfLifeRemaining(currentConfidence, createdAt)

	// For exponential decay, remaining half-life is constant
	assert.Equal(t, cd.HalfLife, remaining)
}

func TestConfidenceDecay_GetStats(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)

	// Add some reinforcements
	cd.Reinforce("rel-1")
	cd.Reinforce("rel-1")
	cd.Reinforce("rel-2")
	cd.Reinforce("rel-3")
	cd.Reinforce("rel-3")
	cd.Reinforce("rel-3")

	stats := cd.GetStats()

	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.TotalRelationships)
	assert.Equal(t, 3, stats.ReinforcedRelationships)
	assert.Greater(t, stats.AverageReinforcements, 1.0)
	assert.NotNil(t, stats.DecayConfiguration)
}

func TestConfidenceDecay_InvalidInputs(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.ReferenceTime = time.Now()

	tests := []struct {
		name              string
		initialConfidence float64
		wantErr           bool
	}{
		{"valid confidence 0.5", 0.5, false},
		{"valid confidence 1.0", 1.0, false},
		{"valid confidence 0.0", 0.0, false},
		{"invalid confidence -0.1", -0.1, true},
		{"invalid confidence 1.5", 1.5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cd.CalculateDecay(tt.initialConfidence, time.Now())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfidenceDecay_FutureDate(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.ReferenceTime = time.Now()

	// Future date should not decay
	futureDate := cd.ReferenceTime.Add(30 * 24 * time.Hour)
	decayed, err := cd.CalculateDecay(1.0, futureDate)

	require.NoError(t, err)
	assert.Equal(t, 1.0, decayed)
}

func TestDefaultConfidenceDecayConfig(t *testing.T) {
	config := DefaultConfidenceDecayConfig()

	assert.NotNil(t, config)
	assert.True(t, config.EnableReinforcement)
	assert.Equal(t, 0.1, config.ReinforcementBonus)
	assert.Equal(t, 0.95, config.ReinforcementDecay)
	assert.Equal(t, 1.0, config.MaxConfidence)
	assert.True(t, config.PreserveCritical)
	assert.Equal(t, 0.9, config.CriticalThreshold)
	assert.NotEmpty(t, config.StepIntervals)
	assert.NotEmpty(t, config.StepValues)
}

func TestDecayFunctions(t *testing.T) {
	assert.Equal(t, DecayFunction("exponential"), DecayExponential)
	assert.Equal(t, DecayFunction("linear"), DecayLinear)
	assert.Equal(t, DecayFunction("logarithmic"), DecayLogarithmic)
	assert.Equal(t, DecayFunction("step"), DecayStep)
}

func TestConfidenceDecay_MinimumConfidenceFloor(t *testing.T) {
	cd := NewConfidenceDecay(1*time.Hour, 0.3) // High minimum
	cd.Config.PreserveCritical = false         // Disable critical preservation to test floor
	cd.ReferenceTime = time.Now()

	// Very old confidence should decay but respect the floor
	veryOld := cd.ReferenceTime.Add(-365 * 24 * time.Hour)
	decayed, err := cd.CalculateDecay(1.0, veryOld)

	require.NoError(t, err)
	// Should be at or very close to the minimum floor
	assert.GreaterOrEqual(t, decayed, cd.MinimumConfidence)
	assert.LessOrEqual(t, decayed, cd.MinimumConfidence+0.01) // Allow small margin
}

func TestConfidenceDecay_MaxConfidenceCap(t *testing.T) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.Config.MaxConfidence = 0.9 // Lower cap
	cd.ReferenceTime = time.Now()

	relationshipID := "rel-123"

	// Add many reinforcements
	for i := 0; i < 20; i++ {
		cd.Reinforce(relationshipID)
	}

	// Recent relationship with lots of reinforcement
	recent := cd.ReferenceTime.Add(-1 * 24 * time.Hour)
	decayed, err := cd.CalculateDecayWithReinforcement(relationshipID, 0.8, recent)

	require.NoError(t, err)
	assert.LessOrEqual(t, decayed, cd.Config.MaxConfidence)
}

func BenchmarkConfidenceDecay_ExponentialDecay(b *testing.B) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.ReferenceTime = time.Now()
	createdAt := cd.ReferenceTime.Add(-15 * 24 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cd.CalculateDecay(0.8, createdAt)
	}
}

func BenchmarkConfidenceDecay_WithReinforcement(b *testing.B) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.ReferenceTime = time.Now()
	createdAt := cd.ReferenceTime.Add(-15 * 24 * time.Hour)
	cd.Reinforce("rel-1")
	cd.Reinforce("rel-1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cd.CalculateDecayWithReinforcement("rel-1", 0.8, createdAt)
	}
}

func BenchmarkConfidenceDecay_BatchCalculate(b *testing.B) {
	cd := NewConfidenceDecay(30*24*time.Hour, 0.1)
	cd.ReferenceTime = time.Now()

	inputs := make([]DecayInput, 100)
	for i := 0; i < 100; i++ {
		inputs[i] = DecayInput{
			RelationshipID:    "rel-" + string(rune(i)),
			InitialConfidence: 0.8,
			CreatedAt:         cd.ReferenceTime.Add(-time.Duration(i) * 24 * time.Hour),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cd.BatchCalculateDecay(inputs)
	}
}
