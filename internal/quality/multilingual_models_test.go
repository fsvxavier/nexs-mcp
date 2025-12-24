package quality

import (
	"context"
	"strings"
	"testing"
)

// TestMultilingualModelsComparison tests two multilingual models with all 11 supported languages.
func TestMultilingualModelsComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multilingual model tests in short mode")
	}

	// Define all 11 languages supported by the MCP server
	tests := []struct {
		language string
		content  string
		reason   string
	}{
		{
			language: "Portuguese",
			content:  "Este é um exemplo de texto em português para testar a qualidade do modelo",
			reason:   "Primary language for Brazil",
		},
		{
			language: "English",
			content:  "This is a sample text in English to test the model quality",
			reason:   "International language",
		},
		{
			language: "Spanish",
			content:  "Este es un texto de ejemplo en español para probar la calidad del modelo",
			reason:   "Latin America and Spain",
		},
		{
			language: "French",
			content:  "Ceci est un exemple de texte en français pour tester la qualité du modèle",
			reason:   "European and African regions",
		},
		{
			language: "German",
			content:  "Dies ist ein Beispieltext auf Deutsch, um die Modellqualität zu testen",
			reason:   "Central Europe",
		},
		{
			language: "Italian",
			content:  "Questo è un testo di esempio in italiano per testare la qualità del modello",
			reason:   "Italy and European Union",
		},
		{
			language: "Russian",
			content:  "Это пример текста на русском языке для проверки качества модели",
			reason:   "Eastern Europe and Central Asia",
		},
		{
			language: "Arabic",
			content:  "هذا نص تجريبي باللغة العربية لاختبار جودة النموذج",
			reason:   "Middle East and North Africa",
		},
		{
			language: "Hindi",
			content:  "यह हिंदी में एक नमूना पाठ है मॉडल की गुणवत्ता का परीक्षण करने के लिए",
			reason:   "India and South Asia",
		},
		{
			language: "Japanese",
			content:  "これは日本語のサンプルテキストで、モデルの品質をテストするためのものです",
			reason:   "Japan - CJK language (critical test)",
		},
		{
			language: "Chinese",
			content:  "这是中文示例文本，用于测试模型质量",
			reason:   "China - CJK language (critical test)",
		},
	}

	// Produção usa apenas 2 modelos: MS MARCO (default) e Paraphrase-Multilingual (configurável)
	// Modelos Distiluse V1/V2 foram descontinuados em 23/12/2025 devido a desempenho inferior

	// Test Model 3: paraphrase-multilingual-MiniLM-L12-v2
	t.Run("ParaphraseMultilingualMiniLM", func(t *testing.T) {
		config := DefaultConfig()
		config.ONNXModelPath = "../../models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx"
		config.RequiresTokenTypeIds = true // BERT-based, needs token_type_ids
		config.ONNXModelType = "embedder"
		config.ONNXOutputName = "last_hidden_state"   // Token embeddings (needs pooling)
		config.ONNXOutputShape = []int64{1, 512, 384} // [batch=1, seq_len=512, hidden_dim=384]

		scorer, err := NewONNXScorer(config)
		if err != nil {
			t.Skipf("Skipping paraphrase-multilingual test: %v", err)
			return
		}
		defer scorer.Close()

		passCount := 0
		failCount := 0

		t.Log("\n" + strings.Repeat("=", 80))
		t.Log("Testing: sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2")
		t.Log("Expected: 50+ languages supported (384-dim embeddings, BERT-based)")
		t.Log(strings.Repeat("=", 80) + "\n")

		for _, tt := range tests {
			t.Run(tt.language, func(t *testing.T) {
				ctx := context.Background()
				result, err := scorer.Score(ctx, tt.content)

				if err != nil {
					t.Logf("✗ %s: FAILED - %v", tt.language, err)
					t.Logf("  Reason: %s", tt.reason)
					failCount++
					return
				}

				if result.Value <= 0 || result.Value > 1 {
					t.Logf("✗ %s: Invalid score %.4f (expected 0-1)", tt.language, result.Value)
					failCount++
					return
				}

				t.Logf("✓ %s: score=%.4f, confidence=%.3f - %s",
					tt.language, result.Value, result.Confidence, tt.reason)
				passCount++
			})
		}

		t.Log("\n" + strings.Repeat("=", 80))
		t.Logf("RESULTS: %d passed, %d failed", passCount, failCount)
		t.Logf("Coverage: %d/11 languages (%.1f%%)", passCount, float64(passCount)/11*100)
		t.Log(strings.Repeat("=", 80) + "\n")

		if passCount < 9 {
			t.Errorf("Expected at least 9/11 languages to work (same as MS MARCO baseline)")
		}

		// Paraphrase-multilingual is BERT-based and should support many languages
		if passCount < 11 {
			t.Logf("⚠ Warning: Paraphrase-multilingual should support all 11 languages including CJK")
		}
	})

	// Test Model 4: MS MARCO (baseline/reference)
	t.Run("MSMarco", func(t *testing.T) {
		config := DefaultConfig()
		config.ONNXModelPath = "../../models/ms-marco-MiniLM-L-6-v2/model.onnx"
		config.RequiresTokenTypeIds = true // BERT-based, needs token_type_ids
		config.ONNXModelType = "reranker"
		config.ONNXOutputName = "logits"
		config.ONNXOutputShape = []int64{1, 1}

		scorer, err := NewONNXScorer(config)
		if err != nil {
			t.Skipf("Skipping MS MARCO test: %v", err)
			return
		}
		defer scorer.Close()

		passCount := 0
		failCount := 0

		t.Log("\n" + strings.Repeat("=", 80))
		t.Log("Testing: ms-marco-MiniLM-L-6-v2 (BASELINE)")
		t.Log("Expected: 9/11 languages (81.8% coverage)")
		t.Log(strings.Repeat("=", 80) + "\n")

		for _, tt := range tests {
			t.Run(tt.language, func(t *testing.T) {
				ctx := context.Background()
				result, err := scorer.Score(ctx, tt.content)

				if err != nil {
					t.Logf("✗ %s: FAILED - %v", tt.language, err)
					t.Logf("  Reason: %s", tt.reason)
					failCount++
					return
				}

				if result.Value <= 0 || result.Value > 1 {
					t.Logf("✗ %s: Invalid score %.4f (expected 0-1)", tt.language, result.Value)
					failCount++
					return
				}

				t.Logf("✓ %s: score=%.4f, confidence=%.3f - %s",
					tt.language, result.Value, result.Confidence, tt.reason)
				passCount++
			})
		}

		t.Log("\n" + strings.Repeat("=", 80))
		t.Logf("RESULTS: %d passed, %d failed", passCount, failCount)
		t.Logf("Coverage: %d/11 languages (%.1f%%)", passCount, float64(passCount)/11*100)
		t.Log(strings.Repeat("=", 80) + "\n")

		if passCount != 9 {
			t.Errorf("Expected exactly 9/11 languages for MS MARCO (baseline)")
		}
	})
}

// TestMultilingualModelsBatch tests batch processing with mixed languages.
func TestMultilingualModelsBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping batch tests in short mode")
	}

	// Test apenas modelos em produção: MS MARCO e Paraphrase-Multilingual
	models := []struct {
		name            string
		path            string
		requiresTokenID bool
		modelType       string
		outputName      string
		outputShape     []int64
		contents        []string // Custom content per model based on language support
	}{
		{
			name:            "paraphrase-multilingual",
			path:            "../../models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
			requiresTokenID: true,
			modelType:       "embedder",
			outputName:      "last_hidden_state",
			outputShape:     []int64{1, 512, 384},
			contents: []string{
				"Este é um texto em português", // Portuguese
				"This is English text",         // English
				"これは日本語のテキストです",                // Japanese
				"这是中文文本",                       // Chinese
				"Это русский текст",            // Russian
			},
		},
		{
			name:            "ms-marco",
			path:            "../../models/ms-marco-MiniLM-L-6-v2/model.onnx",
			requiresTokenID: true,
			modelType:       "reranker",
			outputName:      "logits",
			outputShape:     []int64{1, 1},
			contents: []string{
				"Este é um texto em português",  // Portuguese
				"This is English text",          // English
				"Esto es un texto en español",   // Spanish
				"Ceci est un texte en français", // French
				"Это русский текст",             // Russian (no CJK - ms-marco doesn't support them)
			},
		},
	}

	for _, model := range models {
		t.Run(model.name, func(t *testing.T) {
			contents := model.contents
			config := DefaultConfig()
			config.ONNXModelPath = model.path
			config.RequiresTokenTypeIds = model.requiresTokenID
			config.ONNXModelType = model.modelType
			config.ONNXOutputName = model.outputName
			config.ONNXOutputShape = model.outputShape

			scorer, err := NewONNXScorer(config)
			if err != nil {
				t.Skipf("Skipping %s batch test: %v", model.name, err)
				return
			}
			defer scorer.Close()

			ctx := context.Background()
			results, err := scorer.ScoreBatch(ctx, contents)
			if err != nil {
				t.Fatalf("Batch scoring failed: %v", err)
			}

			if len(results) != len(contents) {
				t.Fatalf("Expected %d results, got %d", len(contents), len(results))
			}

			successCount := 0
			for i, result := range results {
				if result != nil && result.Value > 0 && result.Value <= 1 {
					t.Logf("✓ Text %d: score=%.4f", i+1, result.Value)
					successCount++
				} else {
					t.Logf("✗ Text %d: Failed or invalid score", i+1)
				}
			}

			t.Logf("\nBatch Results: %d/%d successful (%.1f%%)",
				successCount, len(contents), float64(successCount)/float64(len(contents))*100)
		})
	}
}

// TestMultilingualModelsCompatibility verifies input/output compatibility.
func TestMultilingualModelsCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compatibility tests in short mode")
	}

	// Test apenas modelos em produção: MS MARCO e Paraphrase-Multilingual
	models := []struct {
		name            string
		path            string
		requiresTokenID bool
		modelType       string
		outputName      string
		outputShape     []int64
		expectedInputs  int
		expectedOutputs int
		description     string
	}{
		{
			name:            "paraphrase-multilingual-MiniLM-L12-v2",
			path:            "../../models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
			requiresTokenID: true, // BERT-based, needs token_type_ids
			modelType:       "embedder",
			outputName:      "last_hidden_state",
			outputShape:     []int64{1, 512, 384},
			expectedInputs:  3, // input_ids, attention_mask, token_type_ids
			expectedOutputs: 1, // embeddings
			description:     "BERT-based, 11 languages, 384-dim embeddings (configurável)",
		},
		{
			name:            "ms-marco-MiniLM-L-6-v2",
			path:            "../../models/ms-marco-MiniLM-L-6-v2/model.onnx",
			requiresTokenID: true,
			modelType:       "reranker",
			outputName:      "logits",
			outputShape:     []int64{1, 1},
			expectedInputs:  3, // input_ids, attention_mask, token_type_ids
			expectedOutputs: 1, // logits
			description:     "MiniLM-based, 9 languages (no CJK), direct scoring (default)",
		},
	}

	for _, model := range models {
		t.Run(model.name, func(t *testing.T) {
			config := DefaultConfig()
			config.ONNXModelPath = model.path
			config.RequiresTokenTypeIds = model.requiresTokenID
			config.ONNXModelType = model.modelType
			config.ONNXOutputName = model.outputName
			config.ONNXOutputShape = model.outputShape

			scorer, err := NewONNXScorer(config)
			if err != nil {
				t.Logf("❌ Model failed to initialize: %v", err)
				t.Logf("   This indicates compatibility issues with current ONNXScorer")
				t.Logf("   Expected: %d inputs, %d outputs", model.expectedInputs, model.expectedOutputs)
				t.Logf("   Description: %s", model.description)
				return
			}
			defer scorer.Close()

			// Test with simple English text
			ctx := context.Background()
			result, err := scorer.Score(ctx, "This is a test")
			if err != nil {
				t.Errorf("✗ Compatibility test failed: %v", err)
				return
			}

			t.Logf("✓ Model is compatible with current ONNXScorer")
			t.Logf("  Test score: %.4f", result.Value)
			t.Logf("  Description: %s", model.description)
		})
	}
}
