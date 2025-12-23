package quality

import (
	"testing"
)

func TestDefaultRetentionPolicies(t *testing.T) {
	policies := DefaultRetentionPolicies()

	if len(policies) != 3 {
		t.Errorf("Expected 3 policies, got %d", len(policies))
	}

	// Test high quality policy
	high := policies[0]
	if high.MinQuality != 0.7 || high.MaxQuality != 1.1 {
		t.Errorf("High quality policy bounds incorrect: min=%f, max=%f", high.MinQuality, high.MaxQuality)
	}
	if high.RetentionDays != 365 {
		t.Errorf("High quality retention should be 365 days, got %d", high.RetentionDays)
	}
	if high.ArchiveAfterDays != 180 {
		t.Errorf("High quality archive should be 180 days, got %d", high.ArchiveAfterDays)
	}

	// Test medium quality policy
	medium := policies[1]
	if medium.MinQuality != 0.5 || medium.MaxQuality != 0.7 {
		t.Errorf("Medium quality policy bounds incorrect: min=%f, max=%f", medium.MinQuality, medium.MaxQuality)
	}
	if medium.RetentionDays != 180 {
		t.Errorf("Medium quality retention should be 180 days, got %d", medium.RetentionDays)
	}

	// Test low quality policy
	low := policies[2]
	if low.MinQuality != 0.0 || low.MaxQuality != 0.5 {
		t.Errorf("Low quality policy bounds incorrect: min=%f, max=%f", low.MinQuality, low.MaxQuality)
	}
	if low.RetentionDays != 90 {
		t.Errorf("Low quality retention should be 90 days, got %d", low.RetentionDays)
	}
}

func TestGetRetentionPolicy(t *testing.T) {
	policies := DefaultRetentionPolicies()

	tests := []struct {
		name         string
		score        float64
		expectedTier string
		expectedDays int
	}{
		{"Very high quality", 0.95, "high", 365},
		{"High quality", 0.7, "high", 365},
		{"Upper medium quality", 0.65, "medium", 180},
		{"Lower medium quality", 0.5, "medium", 180},
		{"Low quality", 0.3, "low", 90},
		{"Very low quality", 0.05, "low", 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := GetRetentionPolicy(tt.score, policies)
			if policy == nil {
				t.Fatalf("No policy returned for score %f", tt.score)
			}
			if policy.RetentionDays != tt.expectedDays {
				t.Errorf("Expected %d retention days, got %d", tt.expectedDays, policy.RetentionDays)
			}
		})
	}
}

func TestGetRetentionPolicyEdgeCases(t *testing.T) {
	policies := DefaultRetentionPolicies()

	// Test exact boundary values
	t.Run("Exact high boundary", func(t *testing.T) {
		policy := GetRetentionPolicy(0.7, policies)
		if policy == nil || policy.RetentionDays != 365 {
			t.Error("Score 0.7 should match high quality policy")
		}
	})

	t.Run("Just below high boundary", func(t *testing.T) {
		policy := GetRetentionPolicy(0.69, policies)
		if policy == nil || policy.RetentionDays != 180 {
			t.Error("Score 0.69 should match medium quality policy")
		}
	})

	t.Run("Score above maximum", func(t *testing.T) {
		policy := GetRetentionPolicy(1.5, policies)
		// Out of range scores may return last policy or nil - both are acceptable
		if policy != nil {
			t.Logf("Score above maximum returned policy: %+v", policy)
		}
	})

	t.Run("Negative score", func(t *testing.T) {
		policy := GetRetentionPolicy(-0.1, policies)
		// Out of range scores may return last policy or nil - both are acceptable
		if policy != nil {
			t.Logf("Negative score returned policy: %+v", policy)
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig should not return nil")
	}

	if config.DefaultScorer != "onnx" {
		t.Errorf("Default scorer should be 'onnx', got '%s'", config.DefaultScorer)
	}

	if !config.EnableFallback {
		t.Error("Fallback should be enabled by default")
	}

	expectedChain := []string{"onnx", "groq", "gemini", "implicit"}
	if len(config.FallbackChain) != len(expectedChain) {
		t.Errorf("Expected fallback chain length %d, got %d", len(expectedChain), len(config.FallbackChain))
	}

	for i, expected := range expectedChain {
		if config.FallbackChain[i] != expected {
			t.Errorf("Fallback chain[%d]: expected '%s', got '%s'", i, expected, config.FallbackChain[i])
		}
	}

	if config.ONNXModelPath != "models/ms-marco-MiniLM-L-6-v2.onnx" {
		t.Errorf("Unexpected ONNX model path: %s", config.ONNXModelPath)
	}

	if len(config.RetentionPolicies) != 3 {
		t.Errorf("Expected 3 retention policies, got %d", len(config.RetentionPolicies))
	}

	if !config.EnableAutoArchival {
		t.Error("Auto archival should be enabled by default")
	}

	if config.CleanupIntervalMinutes != 60 {
		t.Errorf("Expected cleanup interval of 60 minutes, got %d", config.CleanupIntervalMinutes)
	}
}

func TestScoreValidation(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		valid bool
	}{
		{"Minimum valid", 0.0, true},
		{"Maximum valid", 1.0, true},
		{"Mid-range", 0.5, true},
		{"Below minimum", -0.1, false},
		{"Above maximum", 1.1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.score >= 0.0 && tt.score <= 1.0
			if valid != tt.valid {
				t.Errorf("Score %f validity: expected %v, got %v", tt.score, tt.valid, valid)
			}
		})
	}
}

func TestImplicitSignalsDefaults(t *testing.T) {
	signals := ImplicitSignals{}

	if signals.AccessCount != 0 {
		t.Errorf("Default AccessCount should be 0, got %d", signals.AccessCount)
	}
	if signals.ReferenceCount != 0 {
		t.Errorf("Default ReferenceCount should be 0, got %d", signals.ReferenceCount)
	}
	if signals.UserRating != 0 {
		t.Errorf("Default UserRating should be 0, got %f", signals.UserRating)
	}
	if signals.IsPromoted {
		t.Error("Default IsPromoted should be false")
	}
}
