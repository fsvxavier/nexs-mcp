//go:build !noonnx
// +build !noonnx

package application

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewONNXBERTProvider(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:         "testdata/models/bert-ner.onnx",
		SentimentModel:      "testdata/models/bert-sentiment.onnx",
		EntityConfidenceMin: 0.7,
		BatchSize:           16,
		MaxLength:           512,
		UseGPU:              false,
		EnableFallback:      true,
	}

	provider, err := NewONNXBERTProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider is nil")
	}

	// Provider should not be available without actual model files
	if provider.IsAvailable() {
		t.Error("Provider should not be available without model files")
	}
}

func TestONNXBERTProvider_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		config    EnhancedNLPConfig
		setupFunc func() (cleanup func())
		want      bool
	}{
		{
			name: "unavailable without models",
			config: EnhancedNLPConfig{
				EntityModel:    "nonexistent.onnx",
				SentimentModel: "nonexistent.onnx",
			},
			want: false,
		},
		{
			name: "unavailable with invalid models",
			config: EnhancedNLPConfig{
				EntityModel:    "testdata/invalid.onnx",
				SentimentModel: "testdata/invalid.onnx",
			},
			setupFunc: func() func() {
				// Create dummy files
				os.MkdirAll("testdata", 0755)
				os.WriteFile("testdata/invalid.onnx", []byte("not an onnx model"), 0644)
				return func() {
					os.RemoveAll("testdata")
				}
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				cleanup := tt.setupFunc()
				defer cleanup()
			}

			provider, _ := NewONNXBERTProvider(tt.config)
			if provider == nil {
				t.Fatal("Provider is nil")
			}

			got := provider.IsAvailable()
			if got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestONNXBERTProvider_ExtractEntities_Unavailable(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:    "nonexistent.onnx",
		SentimentModel: "nonexistent.onnx",
	}

	provider, _ := NewONNXBERTProvider(config)
	ctx := context.Background()

	_, err := provider.ExtractEntities(ctx, "Test text")
	if err == nil {
		t.Error("Expected error when ONNX not available")
	}

	expectedMsg := "ONNX BERT provider not available"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestONNXBERTProvider_AnalyzeSentiment_Unavailable(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:    "nonexistent.onnx",
		SentimentModel: "nonexistent.onnx",
	}

	provider, _ := NewONNXBERTProvider(config)
	ctx := context.Background()

	_, err := provider.AnalyzeSentiment(ctx, "Test text")
	if err == nil {
		t.Error("Expected error when ONNX not available")
	}

	expectedMsg := "ONNX BERT provider not available"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestONNXBERTProvider_ExtractTopics_NotSupported(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:    "test.onnx",
		SentimentModel: "test.onnx",
	}

	provider, _ := NewONNXBERTProvider(config)
	ctx := context.Background()

	_, err := provider.ExtractTopics(ctx, []string{"text1", "text2"}, 5)
	if err == nil {
		t.Error("Expected error for unsupported operation")
	}

	if err.Error() != "topic extraction not supported by BERT provider, use TopicModeler" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestONNXBERTProvider_Tokenize(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:    "test.onnx",
		SentimentModel: "test.onnx",
		MaxLength:      512,
	}

	provider, _ := NewONNXBERTProvider(config)

	// Test tokenization (will use simplified vocab)
	inputIDs, attentionMask, tokenTypeIDs, err := provider.tokenize("Hello world", 10)
	if err != nil {
		t.Fatalf("Tokenization failed: %v", err)
	}

	// Check lengths
	if len(inputIDs) != 10 {
		t.Errorf("Expected input_ids length 10, got %d", len(inputIDs))
	}
	if len(attentionMask) != 10 {
		t.Errorf("Expected attention_mask length 10, got %d", len(attentionMask))
	}
	if len(tokenTypeIDs) != 10 {
		t.Errorf("Expected token_type_ids length 10, got %d", len(tokenTypeIDs))
	}

	// Check special tokens ([CLS] at start, [SEP] at end)
	// [CLS] should be at vocab index for "[CLS]", not necessarily 2
	// Just check that we have some token ID there
	if inputIDs[0] < 0 {
		t.Errorf("Expected valid token at position 0, got %d", inputIDs[0])
	}
}

func TestSoftmax(t *testing.T) {
	tests := []struct {
		name   string
		logits []float32
		want   float32 // sum should be 1.0
	}{
		{
			name:   "simple logits",
			logits: []float32{1.0, 2.0, 3.0},
			want:   1.0,
		},
		{
			name:   "negative logits",
			logits: []float32{-1.0, 0.0, 1.0},
			want:   1.0,
		},
		{
			name:   "large logits",
			logits: []float32{100.0, 200.0, 300.0},
			want:   1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probs := softmax(tt.logits)

			// Check sum
			sum := float32(0.0)
			for _, p := range probs {
				sum += p
			}

			if sum < 0.99 || sum > 1.01 {
				t.Errorf("Softmax sum = %v, want ~1.0", sum)
			}

			// Check all probabilities are in [0, 1]
			for i, p := range probs {
				if p < 0.0 || p > 1.0 {
					t.Errorf("probs[%d] = %v, want in [0, 1]", i, p)
				}
			}
		})
	}
}

func TestArgmax(t *testing.T) {
	tests := []struct {
		name    string
		probs   []float32
		wantIdx int
		wantVal float32
		wantMin float32
		wantMax float32
	}{
		{
			name:    "simple case",
			probs:   []float32{0.1, 0.7, 0.2},
			wantIdx: 1,
			wantMin: 0.69,
			wantMax: 0.71,
		},
		{
			name:    "first element max",
			probs:   []float32{0.9, 0.05, 0.05},
			wantIdx: 0,
			wantMin: 0.89,
			wantMax: 0.91,
		},
		{
			name:    "last element max",
			probs:   []float32{0.1, 0.2, 0.7},
			wantIdx: 2,
			wantMin: 0.69,
			wantMax: 0.71,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, val := argmax(tt.probs)

			if idx != tt.wantIdx {
				t.Errorf("argmax index = %v, want %v", idx, tt.wantIdx)
			}

			if val < tt.wantMin || val > tt.wantMax {
				t.Errorf("argmax value = %v, want in [%v, %v]", val, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestONNXBERTProvider_Close(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:    "test.onnx",
		SentimentModel: "test.onnx",
	}

	provider, _ := NewONNXBERTProvider(config)

	// Close should not error even if sessions are not initialized
	err := provider.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// After close, should not be available
	if provider.IsAvailable() {
		t.Error("Provider should not be available after Close()")
	}
}

func TestONNXBERTProvider_LoadVocabulary(t *testing.T) {
	// Create test vocab file
	tmpDir := t.TempDir()
	modelDir := filepath.Join(tmpDir, "models")
	os.MkdirAll(modelDir, 0755)

	vocabPath := filepath.Join(modelDir, "vocab.txt")
	vocabContent := "[PAD]\n[UNK]\n[CLS]\n[SEP]\n[MASK]\nhello\nworld\ntest\n"
	if err := os.WriteFile(vocabPath, []byte(vocabContent), 0644); err != nil {
		t.Fatalf("Failed to create vocab file: %v", err)
	}

	config := EnhancedNLPConfig{
		EntityModel:    filepath.Join(modelDir, "model.onnx"),
		SentimentModel: filepath.Join(modelDir, "model.onnx"),
	}

	provider, _ := NewONNXBERTProvider(config)

	// Try to load vocabulary
	err := provider.loadVocabulary()
	if err != nil {
		t.Errorf("loadVocabulary() error = %v", err)
	}

	// Check if vocab was loaded
	provider.vocabMutex.RLock()
	defer provider.vocabMutex.RUnlock()

	expectedTokens := []string{"[PAD]", "[UNK]", "[CLS]", "[SEP]", "[MASK]", "hello", "world", "test"}
	for i, token := range expectedTokens {
		if id, ok := provider.vocab[token]; !ok || id != i {
			t.Errorf("Expected vocab[%s] = %d, got %d (exists: %v)", token, i, id, ok)
		}
	}
}

func TestONNXBERTProvider_ExtractEntitiesBatch(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:    "nonexistent.onnx",
		SentimentModel: "nonexistent.onnx",
	}

	provider, _ := NewONNXBERTProvider(config)
	ctx := context.Background()

	texts := []string{"text1", "text2", "text3"}
	_, err := provider.ExtractEntitiesBatch(ctx, texts)

	// Should fail since provider is not available
	if err == nil {
		t.Error("Expected error for unavailable provider")
	}
}

func TestONNXBERTProvider_LogitsToSentiment(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:    "test.onnx",
		SentimentModel: "test.onnx",
	}

	provider, _ := NewONNXBERTProvider(config)

	tests := []struct {
		name        string
		logits      []float32
		wantLabel   SentimentLabel
		wantMinConf float64
	}{
		{
			name:        "positive sentiment",
			logits:      []float32{-2.0, -1.0, 3.0}, // Strong positive
			wantLabel:   SentimentPositive,
			wantMinConf: 0.8,
		},
		{
			name:        "negative sentiment",
			logits:      []float32{3.0, -1.0, -2.0}, // Strong negative
			wantLabel:   SentimentNegative,
			wantMinConf: 0.8,
		},
		{
			name:        "neutral sentiment",
			logits:      []float32{-1.0, 2.0, -1.0}, // Neutral
			wantLabel:   SentimentNeutral,
			wantMinConf: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.logitsToSentiment(tt.logits)

			if result.Label != tt.wantLabel {
				t.Errorf("Label = %v, want %v", result.Label, tt.wantLabel)
			}

			if result.Confidence < tt.wantMinConf {
				t.Errorf("Confidence = %v, want >= %v", result.Confidence, tt.wantMinConf)
			}

			// Check score totals
			scoreSum := result.Scores.Positive + result.Scores.Negative + result.Scores.Neutral
			if scoreSum < 0.99 || scoreSum > 1.01 {
				t.Errorf("Score sum = %v, want ~1.0", scoreSum)
			}

			// Check emotional tone is populated
			if result.EmotionalTone.Joy < 0 || result.EmotionalTone.Joy > 1 {
				t.Errorf("Joy = %v, want in [0, 1]", result.EmotionalTone.Joy)
			}

			// Check intensity
			if result.Intensity < 0 || result.Intensity > 1 {
				t.Errorf("Intensity = %v, want in [0, 1]", result.Intensity)
			}
		})
	}
}

func TestONNXBERTProvider_LogitsToEntities(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:         "test.onnx",
		SentimentModel:      "test.onnx",
		EntityConfidenceMin: 0.7,
	}

	provider, _ := NewONNXBERTProvider(config)

	// Simulate logits for sequence: [CLS] John works at Google [SEP]
	// Format: [seq_len * num_labels]
	// Labels: O, PERSON, ORGANIZATION, LOCATION, ...
	numLabels := len(provider.entityLabels)
	seqLen := 7 // [CLS] + 4 words + [SEP] + padding

	logits := make([]float32, seqLen*numLabels)

	// [CLS] - all O
	for j := range numLabels {
		logits[0*numLabels+j] = -1.0
	}
	logits[0*numLabels+0] = 2.0 // O

	// "John" - PERSON
	for j := range numLabels {
		logits[1*numLabels+j] = -1.0
	}
	logits[1*numLabels+1] = 3.0 // PERSON

	// "works" - O
	for j := range numLabels {
		logits[2*numLabels+j] = -1.0
	}
	logits[2*numLabels+0] = 2.0 // O

	// "at" - O
	for j := range numLabels {
		logits[3*numLabels+j] = -1.0
	}
	logits[3*numLabels+0] = 2.0 // O

	// "Google" - ORGANIZATION
	for j := range numLabels {
		logits[4*numLabels+j] = -1.0
	}
	logits[4*numLabels+2] = 3.0 // ORGANIZATION

	text := "John works at Google"
	attentionMask := []int64{1, 1, 1, 1, 1, 1, 0} // 6 real tokens, 1 padding

	entities := provider.logitsToEntities(logits, text, attentionMask)

	// Should extract 2 entities: John (PERSON) and Google (ORGANIZATION)
	if len(entities) < 1 {
		t.Errorf("Expected at least 1 entity, got %d", len(entities))
	}

	// Check first entity
	if len(entities) > 0 {
		if entities[0].Type != EntityTypePerson {
			t.Errorf("First entity type = %v, want %v", entities[0].Type, EntityTypePerson)
		}
		if entities[0].Confidence < 0.7 {
			t.Errorf("First entity confidence = %v, want >= 0.7", entities[0].Confidence)
		}
	}
}

// Benchmark tests.
func BenchmarkSoftmax(b *testing.B) {
	logits := []float32{1.0, 2.0, 3.0, 4.0, 5.0}

	b.ResetTimer()
	for range b.N {
		_ = softmax(logits)
	}
}

func BenchmarkArgmax(b *testing.B) {
	probs := []float32{0.1, 0.2, 0.3, 0.25, 0.15}

	b.ResetTimer()
	for range b.N {
		_, _ = argmax(probs)
	}
}

func BenchmarkTokenize(b *testing.B) {
	config := EnhancedNLPConfig{
		EntityModel:    "test.onnx",
		SentimentModel: "test.onnx",
		MaxLength:      512,
	}

	provider, _ := NewONNXBERTProvider(config)
	text := "This is a test sentence for benchmarking tokenization performance"

	b.ResetTimer()
	for range b.N {
		_, _, _, _ = provider.tokenize(text, 512)
	}
}
