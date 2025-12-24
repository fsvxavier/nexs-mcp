package quality

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkONNXModels compares all 4 ONNX models for speed.
func BenchmarkONNXModels(b *testing.B) {
	testSamples := []struct {
		language string
		content  string
	}{
		{"Portuguese", "Este é um exemplo de texto em português com boa qualidade e estrutura clara"},
		{"English", "This is a high quality example text with clear structure and good content"},
		{"Spanish", "Este es un texto de ejemplo en español con buena calidad y estructura clara"},
		{"French", "Ceci est un exemple de texte en français de haute qualité avec une structure claire"},
		{"German", "Dies ist ein hochwertiger Beispieltext mit klarer Struktur und gutem Inhalt"},
		{"Italian", "Questo è un testo di esempio in italiano di alta qualità con struttura chiara"},
		{"Russian", "Это высококачественный пример текста с четкой структурой и хорошим содержанием"},
		{"Arabic", "هذا نص تجريبي عالي الجودة ببنية واضحة ومحتوى جيد"},
		{"Hindi", "यह स्पष्ट संरचना और अच्छी सामग्री के साथ एक उच्च गुणवत्ता वाला उदाहरण पाठ है"},
		{"Japanese", "これは明確な構造と良い内容を持つ高品質なサンプルテキストです"},
		{"Chinese", "这是一个具有清晰结构和良好内容的高质量示例文本"},
	}

	models := []struct {
		name   string
		config *Config
	}{
		{
			name: "ParaphraseMultilingual",
			config: &Config{
				ONNXModelPath:        "../../models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
				RequiresTokenTypeIds: true,
				ONNXModelType:        "embedder",
				ONNXOutputName:       "last_hidden_state",
				ONNXOutputShape:      []int64{1, 512, 384},
			},
		},
		{
			name: "MSMarco",
			config: &Config{
				ONNXModelPath:        "../../models/ms-marco-MiniLM-L-6-v2/model.onnx",
				RequiresTokenTypeIds: true,
				ONNXModelType:        "reranker",
				ONNXOutputName:       "logits",
				ONNXOutputShape:      []int64{1, 1},
			},
		},
	}

	for _, model := range models {
		b.Run(model.name, func(b *testing.B) {
			scorer, err := NewONNXScorer(model.config)
			if err != nil {
				b.Skipf("Failed to initialize %s: %v", model.name, err)
				return
			}
			defer scorer.Close()

			b.ResetTimer()
			ctx := context.Background()
			for i := range b.N {
				sample := testSamples[i%len(testSamples)]
				_, err := scorer.Score(ctx, sample.content)
				if err != nil {
					b.Fatalf("Inference failed for %s: %v", sample.language, err)
				}
			}
		})
	}
}

// BenchmarkONNXModelsParallel tests concurrent performance.
func BenchmarkONNXModelsParallel(b *testing.B) {
	testSamples := []struct {
		language string
		content  string
	}{
		{"Portuguese", "Este é um exemplo de texto em português com boa qualidade e estrutura clara"},
		{"English", "This is a high quality example text with clear structure and good content"},
		{"Spanish", "Este es un texto de ejemplo en español con buena calidad y estructura clara"},
		{"French", "Ceci est un exemple de texte en français de haute qualité avec une structure claire"},
		{"German", "Dies ist ein hochwertiger Beispieltext mit klarer Struktur und gutem Inhalt"},
	}

	models := []struct {
		name   string
		config *Config
	}{
		{
			name: "ParaphraseMultilingual",
			config: &Config{
				ONNXModelPath:        "../../models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
				RequiresTokenTypeIds: true,
				ONNXModelType:        "embedder",
				ONNXOutputName:       "last_hidden_state",
				ONNXOutputShape:      []int64{1, 512, 384},
			},
		},
		{
			name: "MSMarco",
			config: &Config{
				ONNXModelPath:        "../../models/ms-marco-MiniLM-L-6-v2/model.onnx",
				RequiresTokenTypeIds: true,
				ONNXModelType:        "reranker",
				ONNXOutputName:       "logits",
				ONNXOutputShape:      []int64{1, 1},
			},
		},
	}

	for _, model := range models {
		b.Run(model.name, func(b *testing.B) {
			scorer, err := NewONNXScorer(model.config)
			if err != nil {
				b.Skipf("Failed to initialize %s: %v", model.name, err)
				return
			}
			defer scorer.Close()

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				ctx := context.Background()
				i := 0
				for pb.Next() {
					sample := testSamples[i%len(testSamples)]
					_, err := scorer.Score(ctx, sample.content)
					if err != nil {
						b.Fatalf("Inference failed: %v", err)
					}
					i++
				}
			})
		})
	}
}

// TestONNXModelsEffectiveness compares effectiveness across languages.
func TestONNXModelsEffectiveness(t *testing.T) {
	testSamples := []struct {
		language string
		content  string
	}{
		{"Portuguese", "Este é um exemplo de texto em português com boa qualidade e estrutura clara"},
		{"English", "This is a high quality example text with clear structure and good content"},
		{"Spanish", "Este es un texto de ejemplo en español con buena calidad y estructura clara"},
		{"French", "Ceci est un exemple de texte en français de haute qualité avec une structure claire"},
		{"German", "Dies ist ein hochwertiger Beispieltext mit klarer Struktur und gutem Inhalt"},
		{"Italian", "Questo è un testo di esempio in italiano di alta qualità con struttura chiara"},
		{"Russian", "Это высококачественный пример текста с четкой структурой и хорошим содержанием"},
		{"Arabic", "هذا نص تجريبي عالي الجودة ببنية واضحة ومحتوى جيد"},
		{"Hindi", "यह स्पष्ट संरचना और अच्छी सामग्री के साथ एक उच्च गुणवत्ता वाला उदाहरण पाठ है"},
		{"Japanese", "これは明確な構造と良い内容を持つ高品質なサンプルテキストです"},
		{"Chinese", "这是一个具有清晰结构和良好内容的高质量示例文本"},
	}

	models := []struct {
		name   string
		config *Config
	}{
		{
			name: "MSMarco",
			config: &Config{
				ONNXModelPath:        "../../models/ms-marco-MiniLM-L-6-v2/model.onnx",
				RequiresTokenTypeIds: true,
				ONNXModelType:        "reranker",
				ONNXOutputName:       "logits",
				ONNXOutputShape:      []int64{1, 1},
			},
		},
		{
			name: "ParaphraseMultilingual",
			config: &Config{
				ONNXModelPath:        "../../models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
				RequiresTokenTypeIds: true,
				ONNXModelType:        "embedder",
				ONNXOutputName:       "last_hidden_state",
				ONNXOutputShape:      []int64{1, 512, 384},
			},
		},
	}

	results := make(map[string]map[string]float64)
	times := make(map[string]time.Duration)

	ctx := context.Background()

	for _, model := range models {
		t.Run(model.name, func(t *testing.T) {
			scorer, err := NewONNXScorer(model.config)
			if err != nil {
				t.Skipf("Failed to initialize %s: %v", model.name, err)
				return
			}
			defer scorer.Close()

			results[model.name] = make(map[string]float64)
			passed := 0
			totalTime := time.Duration(0)
			tested := 0

			for _, sample := range testSamples {
				// Skip CJK languages for MS MARCO (vocabulary limitations)
				if model.name == "MSMarco" && (sample.language == "Japanese" || sample.language == "Chinese") {
					t.Logf("  ⊘ %s: SKIPPED - CJK not supported by MS MARCO", sample.language)
					continue
				}

				tested++
				start := time.Now()
				result, err := scorer.Score(ctx, sample.content)
				if err != nil {
					t.Logf("  ✗ %s: inference error: %v", sample.language, err)
					continue
				}
				totalTime += time.Since(start)
				results[model.name][sample.language] = result.Value
				elapsed := time.Since(start)

				if result.Value > 0.1 {
					passed++
					t.Logf("  ✓ %s: score=%.4f (%.2fms)", sample.language, result.Value, float64(elapsed.Microseconds())/1000.0)
				} else {
					t.Logf("  ✗ %s: score=%.4f (too low)", sample.language, result.Value)
				}
			}

			times[model.name] = totalTime
			avgTime := totalTime / time.Duration(tested)

			t.Logf("\n  Results: %d/%d languages passed (%.1f%%)", passed, tested, float64(passed)/float64(tested)*100)
			t.Logf("  Average time: %.2fms per inference", float64(avgTime.Microseconds())/1000.0)
		})
	}

	t.Run("Summary", func(t *testing.T) {
		t.Log("\n================================================================================")
		t.Log("EFFECTIVENESS & PERFORMANCE COMPARISON")
		t.Log("================================================================================\n")

		for modelName, langScores := range results {
			if len(langScores) == 0 {
				continue
			}

			var total float64
			for _, score := range langScores {
				total += score
			}
			avgScore := total / float64(len(langScores))
			avgTime := times[modelName] / time.Duration(len(langScores))

			t.Logf("%-25s | Avg Score: %.4f | Avg Time: %.2fms | Languages: %d/%d", modelName, avgScore, float64(avgTime.Microseconds())/1000.0, len(langScores), len(testSamples))
		}

		t.Log("\n================================================================================")
	})
}

// BenchmarkONNXModelsByTextLength tests performance with different text lengths.
func BenchmarkONNXModelsByTextLength(b *testing.B) {
	textLengths := []struct {
		name string
		text string
	}{
		{"Short", "This is a short text."},
		{"Medium", "This is a medium length text with more content to process and evaluate for quality assessment."},
		{"Long", "This is a much longer text that contains significantly more content and information. It includes multiple sentences and covers various topics to provide a comprehensive example for testing the performance of different ONNX models with varying input lengths. The goal is to understand how each model handles longer sequences and whether there are significant performance differences based on input size."},
	}

	models := []struct {
		name   string
		config *Config
	}{
		{
			name: "ParaphraseMultilingual",
			config: &Config{
				ONNXModelPath:        "../../models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
				RequiresTokenTypeIds: true,
				ONNXModelType:        "embedder",
				ONNXOutputName:       "last_hidden_state",
				ONNXOutputShape:      []int64{1, 512, 384},
			},
		},
		{
			name: "MSMarco",
			config: &Config{
				ONNXModelPath:        "../../models/ms-marco-MiniLM-L-6-v2/model.onnx",
				RequiresTokenTypeIds: true,
				ONNXModelType:        "reranker",
				ONNXOutputName:       "logits",
				ONNXOutputShape:      []int64{1, 1},
			},
		},
	}

	for _, model := range models {
		for _, length := range textLengths {
			testName := fmt.Sprintf("%s/%s", model.name, length.name)
			b.Run(testName, func(b *testing.B) {
				scorer, err := NewONNXScorer(model.config)
				if err != nil {
					b.Skipf("Failed to initialize %s: %v", model.name, err)
					return
				}
				defer scorer.Close()

				b.ResetTimer()
				ctx := context.Background()
				for range b.N {
					_, err := scorer.Score(ctx, length.text)
					if err != nil {
						b.Fatalf("Inference failed: %v", err)
					}
				}
			})
		}
	}
}
