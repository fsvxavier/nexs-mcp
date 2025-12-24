package quality

import (
	"context"
	"testing"
	"time"
)

func TestImplicitScorerBasic(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})

	ctx := context.Background()
	if scorer == nil {
		t.Fatal("NewImplicitScorer returned nil")
	}

	if scorer.Name() != "implicit" {
		t.Errorf("Expected name 'implicit', got '%s'", scorer.Name())
	}

	if !scorer.IsAvailable(ctx) {
		t.Error("Implicit scorer should always be available")
	}

	if scorer.Cost() != 0.0 {
		t.Errorf("Expected cost 0.0, got %f", scorer.Cost())
	}
}

func TestImplicitScorerDefaultScore(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	score, err := scorer.Score(ctx, "test content")
	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}

	if score == nil {
		t.Fatal("Score returned nil")
	}

	if score.Value != 0.4 {
		t.Errorf("Default score value should be 0.4, got %f", score.Value)
	}

	if score.Confidence != 0.3 {
		t.Errorf("Default confidence should be 0.3, got %f", score.Confidence)
	}

	if score.Method != "implicit" {
		t.Errorf("Expected method 'implicit', got '%s'", score.Method)
	}

	if score.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestImplicitScorerWithSignals(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	tests := []struct {
		name          string
		signals       ImplicitSignals
		minScore      float64
		maxScore      float64
		minConfidence float64
		maxConfidence float64
	}{
		{
			name: "High activity signals",
			signals: ImplicitSignals{
				AccessCount:    100,
				ReferenceCount: 50,
				AgeInDays:      10,
				LastAccessDays: 1,
				UserRating:     1.0, // 0-1 range
				ContentLength:  1000,
				TagCount:       5,
				IsPromoted:     true,
			},
			minScore:      0.7,
			maxScore:      1.0,
			minConfidence: 0.65,
			maxConfidence: 0.8,
		},
		{
			name: "Medium activity signals",
			signals: ImplicitSignals{
				AccessCount:    20,
				ReferenceCount: 10,
				AgeInDays:      100,
				LastAccessDays: 7,
				UserRating:     0.6, // 0-1 range
				ContentLength:  500,
				TagCount:       2,
				IsPromoted:     false,
			},
			minScore:      0.5,
			maxScore:      0.85,
			minConfidence: 0.5,
			maxConfidence: 0.85,
		},
		{
			name: "Low activity signals",
			signals: ImplicitSignals{
				AccessCount:    1,
				ReferenceCount: 0,
				AgeInDays:      300,
				LastAccessDays: 100,
				UserRating:     0.2, // 0-1 range
				ContentLength:  100,
				TagCount:       0,
				IsPromoted:     false,
			},
			minScore:      0.0,
			maxScore:      0.4,
			minConfidence: 0.4,
			maxConfidence: 0.85,
		},
		{
			name: "Recently created with high engagement",
			signals: ImplicitSignals{
				AccessCount:    50,
				ReferenceCount: 25,
				AgeInDays:      1,
				LastAccessDays: 0,
				UserRating:     0.9, // 0-1 range
				ContentLength:  800,
				TagCount:       3,
				IsPromoted:     false,
			},
			minScore:      0.6,
			maxScore:      1.0,
			minConfidence: 0.6,
			maxConfidence: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := scorer.ScoreWithSignals(ctx, "content", tt.signals)
			if err != nil {
				t.Fatalf("ScoreWithSignals failed: %v", err)
			}

			if score.Value < tt.minScore || score.Value > tt.maxScore {
				t.Errorf("Score %f out of expected range [%f, %f]", score.Value, tt.minScore, tt.maxScore)
			}

			if score.Confidence < tt.minConfidence || score.Confidence > tt.maxConfidence {
				t.Errorf("Confidence %f out of expected range [%f, %f]", score.Confidence, tt.minConfidence, tt.maxConfidence)
			}

			if score.Value < 0 || score.Value > 1 {
				t.Errorf("Score %f out of valid range [0, 1]", score.Value)
			}

			if score.Confidence < 0 || score.Confidence > 1 {
				t.Errorf("Confidence %f out of valid range [0, 1]", score.Confidence)
			}

			if score.Method != "implicit" {
				t.Errorf("Expected method 'implicit', got '%s'", score.Method)
			}
		})
	}
}

func TestImplicitScorerAccessCountWeight(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	// Test access count impact (weight: 0.3)
	lowAccess := ImplicitSignals{AccessCount: 1}
	highAccess := ImplicitSignals{AccessCount: 200}

	scoreLow, _ := scorer.ScoreWithSignals(ctx, "content", lowAccess)
	scoreHigh, _ := scorer.ScoreWithSignals(ctx, "content", highAccess)

	if scoreHigh.Value <= scoreLow.Value {
		t.Error("Higher access count should result in higher score")
	}

	// High access should have noticeable impact given 0.3 weight
	diff := scoreHigh.Value - scoreLow.Value
	if diff < 0.1 {
		t.Errorf("Access count impact too small: %f", diff)
	}
}

func TestImplicitScorerReferenceCountWeight(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	// Test reference count impact (weight: 0.25)
	lowRefs := ImplicitSignals{ReferenceCount: 0}
	highRefs := ImplicitSignals{ReferenceCount: 100}

	scoreLow, _ := scorer.ScoreWithSignals(ctx, "content", lowRefs)
	scoreHigh, _ := scorer.ScoreWithSignals(ctx, "content", highRefs)

	if scoreHigh.Value <= scoreLow.Value {
		t.Error("Higher reference count should result in higher score")
	}

	diff := scoreHigh.Value - scoreLow.Value
	if diff < 0.1 {
		t.Errorf("Reference count impact too small: %f", diff)
	}
}

func TestImplicitScorerRecencyWeight(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	// Test age recency impact (weight: 0.2)
	recent := ImplicitSignals{AgeInDays: 1}
	old := ImplicitSignals{AgeInDays: 500}

	scoreRecent, _ := scorer.ScoreWithSignals(ctx, "content", recent)
	scoreOld, _ := scorer.ScoreWithSignals(ctx, "content", old)

	if scoreRecent.Value <= scoreOld.Value {
		t.Error("Recent content should score higher than old content")
	}
}

func TestImplicitScorerLastAccessWeight(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	// Test last access recency impact (weight: 0.15)
	recentAccess := ImplicitSignals{LastAccessDays: 0}
	oldAccess := ImplicitSignals{LastAccessDays: 200}

	scoreRecentAccess, _ := scorer.ScoreWithSignals(ctx, "content", recentAccess)
	scoreOldAccess, _ := scorer.ScoreWithSignals(ctx, "content", oldAccess)

	if scoreRecentAccess.Value <= scoreOldAccess.Value {
		t.Error("Recently accessed content should score higher")
	}
}

func TestImplicitScorerUserRatingWeight(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	// Test user rating impact (weight: 0.1) - rating is 0-1 scale
	lowRating := ImplicitSignals{UserRating: 0.0}
	highRating := ImplicitSignals{UserRating: 1.0}

	scoreLow, _ := scorer.ScoreWithSignals(ctx, "content", lowRating)
	scoreHigh, _ := scorer.ScoreWithSignals(ctx, "content", highRating)

	// Rating alone has 0.1 weight, so difference should be ~0.1
	diff := scoreHigh.Value - scoreLow.Value
	if diff < 0.05 || diff > 0.15 {
		t.Errorf("User rating impact should be ~0.1, got difference %f (low=%f, high=%f)", diff, scoreLow.Value, scoreHigh.Value)
	}
}

func TestImplicitScorerPromotionBonus(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	// Test promotion bonus (weight: 0.05)
	notPromoted := ImplicitSignals{IsPromoted: false}
	promoted := ImplicitSignals{IsPromoted: true}

	scoreNot, _ := scorer.ScoreWithSignals(ctx, "content", notPromoted)
	scorePro, _ := scorer.ScoreWithSignals(ctx, "content", promoted)

	if scorePro.Value <= scoreNot.Value {
		t.Error("Promoted content should score higher")
	}

	// Promotion bonus should be exactly 0.05
	diff := scorePro.Value - scoreNot.Value
	if diff < 0.04 || diff > 0.06 {
		t.Errorf("Promotion bonus should be ~0.05, got %f", diff)
	}
}

func TestImplicitScorerConfidenceCalculation(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	tests := []struct {
		name            string
		signals         ImplicitSignals
		expectedMinConf float64
	}{
		{
			name:            "No signals - base confidence",
			signals:         ImplicitSignals{},
			expectedMinConf: 0.5,
		},
		{
			name: "All signals present",
			signals: ImplicitSignals{
				AccessCount:    50,
				ReferenceCount: 25,
				AgeInDays:      10,
				LastAccessDays: 1,
				UserRating:     0.8, // 0-1 range
				ContentLength:  500,
				TagCount:       3,
				IsPromoted:     true,
			},
			expectedMinConf: 0.65,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, _ := scorer.ScoreWithSignals(ctx, "content", tt.signals)

			if score.Confidence < tt.expectedMinConf {
				t.Errorf("Confidence %f below expected minimum %f", score.Confidence, tt.expectedMinConf)
			}

			// Confidence should be capped at 0.8 for implicit scorer
			if score.Confidence > 0.8 {
				t.Errorf("Confidence %f exceeds maximum 0.8", score.Confidence)
			}
		})
	}
}

func TestImplicitScorerBatch(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	contents := []string{"content1", "content2", "content3"}
	scores, err := scorer.ScoreBatch(ctx, contents)

	if err != nil {
		t.Fatalf("ScoreBatch failed: %v", err)
	}

	if len(scores) != len(contents) {
		t.Errorf("Expected %d scores, got %d", len(contents), len(scores))
	}

	for i, score := range scores {
		if score == nil {
			t.Errorf("Score %d is nil", i)
			continue
		}

		if score.Value != 0.4 {
			t.Errorf("Batch score %d: expected value 0.4, got %f", i, score.Value)
		}

		if score.Confidence != 0.3 {
			t.Errorf("Batch score %d: expected confidence 0.3, got %f", i, score.Confidence)
		}

		if score.Method != "implicit" {
			t.Errorf("Batch score %d: expected method 'implicit', got '%s'", i, score.Method)
		}
	}
}

func TestImplicitScorerClose(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})

	err := scorer.Close()
	if err != nil {
		t.Errorf("Close should not return error, got: %v", err)
	}

	// Should be safe to close multiple times
	err = scorer.Close()
	if err != nil {
		t.Errorf("Second Close should not return error, got: %v", err)
	}
}

func TestImplicitScorerMetadata(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	signals := ImplicitSignals{
		AccessCount:    25,
		ReferenceCount: 10,
		UserRating:     0.8, // 0-1 range
	}

	score, err := scorer.ScoreWithSignals(ctx, "content", signals)
	if err != nil {
		t.Fatalf("ScoreWithSignals failed: %v", err)
	}

	// Metadata should exist
	if score.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}

	// Check that metadata contains expected fields (structure may vary)
	if _, ok := score.Metadata["access_count"]; ok {
		t.Logf("Metadata includes access_count: %v", score.Metadata["access_count"])
	}

	if _, ok := score.Metadata["reference_count"]; ok {
		t.Logf("Metadata includes reference_count: %v", score.Metadata["reference_count"])
	}
}

func TestImplicitScorerEmptyContent(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	score, err := scorer.Score(ctx, "")
	if err != nil {
		t.Fatalf("Score with empty content failed: %v", err)
	}

	if score == nil {
		t.Fatal("Score should not be nil for empty content")
	}

	// Should still return default score
	if score.Value != 0.4 {
		t.Errorf("Empty content should get default score 0.4, got %f", score.Value)
	}
}

func TestImplicitScorerTimestamp(t *testing.T) {
	scorer := NewImplicitScorer(&Config{})
	ctx := context.Background()

	before := time.Now()
	score, err := scorer.Score(ctx, "content")
	after := time.Now()

	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}

	if score.Timestamp.Before(before) || score.Timestamp.After(after) {
		t.Error("Timestamp should be between before and after time")
	}
}
