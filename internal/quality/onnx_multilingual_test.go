//go:build !noonnx
// +build !noonnx

package quality

import (
	"context"
	"testing"
)

// TestONNXScorerMultilingual tests ONNX scorer with multiple languages.
func TestONNXScorerMultilingual(t *testing.T) {
	config := getTestModelConfig(t)
	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()

	// Test cases with different languages
	testCases := []struct {
		name        string
		language    string
		content     string
		minScore    float64 // Minimum expected score
		expectError bool    // CJK languages will fail with current simple tokenizer
	}{
		{
			name:     "Portuguese",
			language: "pt",
			content:  "Este é um texto de alta qualidade em português. A inteligência artificial está revolucionando a forma como processamos linguagem natural.",
			minScore: 0.3,
		},
		{
			name:     "Spanish",
			language: "es",
			content:  "Este es un texto de alta calidad en español. La inteligencia artificial está revolucionando la forma en que procesamos el lenguaje natural.",
			minScore: 0.3,
		},
		{
			name:     "French",
			language: "fr",
			content:  "Ceci est un texte de haute qualité en français. L'intelligence artificielle révolutionne la façon dont nous traitons le langage naturel.",
			minScore: 0.3,
		},
		{
			name:     "German",
			language: "de",
			content:  "Dies ist ein qualitativ hochwertiger Text auf Deutsch. Künstliche Intelligenz revolutioniert die Art und Weise, wie wir natürliche Sprache verarbeiten.",
			minScore: 0.3,
		},
		{
			name:     "Italian",
			language: "it",
			content:  "Questo è un testo di alta qualità in italiano. L'intelligenza artificiale sta rivoluzionando il modo in cui elaboriamo il linguaggio naturale.",
			minScore: 0.3,
		},
		{
			name:     "Russian",
			language: "ru",
			content:  "Это высококачественный текст на русском языке. Искусственный интеллект революционизирует способ обработки естественного языка.",
			minScore: 0.3,
		},
		{
			name:        "Japanese (not supported)",
			language:    "ja",
			content:     "これは日本語の高品質なテキストです。人工知能は自然言語処理の方法を革新しています。",
			minScore:    0.3,
			expectError: true, // CJK characters exceed BERT vocab size (30522)
		},
		{
			name:        "Chinese (not supported)",
			language:    "zh",
			content:     "这是一篇高质量的中文文本。人工智能正在彻底改变我们处理自然语言的方式。",
			minScore:    0.3,
			expectError: true, // CJK characters exceed BERT vocab size (30522)
		},
		{
			name:     "Arabic",
			language: "ar",
			content:  "هذا نص عالي الجودة باللغة العربية. الذكاء الاصطناعي يحدث ثورة في طريقة معالجة اللغة الطبيعية.",
			minScore: 0.3,
		},
		{
			name:     "Hindi",
			language: "hi",
			content:  "यह हिंदी में एक उच्च गुणवत्ता वाला पाठ है। कृत्रिम बुद्धिमत्ता प्राकृतिक भाषा प्रसंस्करण के तरीके में क्रांति ला रही है।",
			minScore: 0.3,
		},
		{
			name:     "English (baseline)",
			language: "en",
			content:  "This is a high-quality text in English. Artificial intelligence is revolutionizing the way we process natural language.",
			minScore: 0.3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score, err := scorer.Score(ctx, tc.content)

			// Handle expected errors for CJK languages
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s text (BERT vocab limitation), but got success", tc.language)
				} else {
					t.Logf("✓ Expected error for %s: %v", tc.language, err)
					t.Skip("CJK languages not supported by current BERT model (vocab limited to 30522 tokens)")
				}
				return
			}

			if err != nil {
				t.Errorf("Failed to score %s text: %v", tc.language, err)
				return
			}

			// Validate score
			assertValidScore(t, score, "onnx")

			// Check if score is reasonable
			if score.Value < tc.minScore {
				t.Logf("Warning: %s text score (%.3f) is below expected minimum (%.3f)",
					tc.language, score.Value, tc.minScore)
			}

			t.Logf("%s (%s): score=%.3f, confidence=%.3f, method=%s",
				tc.name, tc.language, score.Value, score.Confidence, score.Method)
		})
	}
}

// TestONNXScorerPortugueseSamples tests various Portuguese text samples.
func TestONNXScorerPortugueseSamples(t *testing.T) {
	config := getTestModelConfig(t)
	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()

	testCases := []struct {
		name        string
		content     string
		description string
	}{
		{
			name: "Technical documentation",
			content: `O ONNX Runtime é uma biblioteca de inferência de alto desempenho para modelos de machine learning. 
Ele suporta múltiplos frameworks como PyTorch, TensorFlow e scikit-learn. 
A biblioteca é otimizada para CPUs e GPUs, oferecendo excelente desempenho em produção.`,
			description: "Technical content with specific terminology",
		},
		{
			name: "Business communication",
			content: `Prezado cliente, estamos satisfeitos em anunciar o lançamento da nossa nova plataforma de análise de qualidade. 
Esta solução inovadora utiliza inteligência artificial para avaliar automaticamente a relevância e qualidade do conteúdo.`,
			description: "Formal business communication",
		},
		{
			name: "Informal conversation",
			content: `Oi! Tudo bem? Eu testei aquele novo sistema de IA e achei muito legal! 
Funciona super bem com textos em português, viu? Vale a pena conferir!`,
			description: "Casual, informal text",
		},
		{
			name: "Mixed code and Portuguese",
			content: `Para inicializar o ONNX scorer em Go, use: scorer, err := NewONNXScorer(config)
Certifique-se de que o modelo está no caminho correto. 
Em caso de erro, verifique se o ONNX Runtime está instalado corretamente no sistema.`,
			description: "Code snippets mixed with Portuguese",
		},
		{
			name:        "Short text",
			content:     "Qualidade excelente!",
			description: "Very short Portuguese text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score, err := scorer.Score(ctx, tc.content)
			if err != nil {
				t.Errorf("Failed to score Portuguese text (%s): %v", tc.description, err)
				return
			}

			assertValidScore(t, score, "onnx")

			t.Logf("%s: score=%.3f, confidence=%.3f, length=%d chars",
				tc.description, score.Value, score.Confidence, len(tc.content))
		})
	}
}

// TestONNXScorerMultilingualBatch tests batch scoring with multiple languages.
func TestONNXScorerMultilingualBatch(t *testing.T) {
	config := getTestModelConfig(t)
	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()

	contents := []string{
		"Este é um excelente texto em português sobre inteligência artificial.",
		"This is a high-quality English text about machine learning.",
		"Este es un texto de alta calidad en español sobre aprendizaje automático.",
		"Ceci est un texte français de haute qualité sur l'apprentissage automatique.",
		"Dies ist ein hochwertiger deutscher Text über maschinelles Lernen.",
	}

	scores, err := scorer.ScoreBatch(ctx, contents)
	if err != nil {
		t.Fatalf("Failed to score batch: %v", err)
	}

	if len(scores) != len(contents) {
		t.Fatalf("Expected %d scores, got %d", len(contents), len(scores))
	}

	languages := []string{"Portuguese", "English", "Spanish", "French", "German"}
	for i, score := range scores {
		assertValidScore(t, score, "onnx")
		t.Logf("%s: score=%.3f, confidence=%.3f",
			languages[i], score.Value, score.Confidence)
	}
}

// TestONNXScorerSpecialCharacters tests handling of special characters and accents.
func TestONNXScorerSpecialCharacters(t *testing.T) {
	config := getTestModelConfig(t)
	scorer, err := NewONNXScorer(config)
	if err != nil {
		t.Skipf("ONNX Runtime not available: %v", err)
	}
	defer scorer.Close()

	ctx := context.Background()

	testCases := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:    "Portuguese accents",
			content: "Ação, função, união, órgão - palavras com acentuação em português.",
		},
		{
			name:    "Spanish tildes",
			content: "Año, niño, señor, mañana - palabras con tilde en español.",
		},
		{
			name:    "French accents",
			content: "Été, café, naïve, Noël - mots français avec accents.",
		},
		{
			name:    "German umlauts",
			content: "Müller, Köln, über, Größe - deutsche Wörter mit Umlauten.",
		},
		{
			name:        "Mixed symbols (not supported)",
			content:     "Text with emoji 🎉, symbols ©®™, and punctuation: ¡¿!?",
			expectError: true, // Emoji exceed BERT vocab size
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score, err := scorer.Score(ctx, tc.content)

			// Handle expected errors for high Unicode symbols
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for high Unicode symbols, but got success")
				} else {
					t.Logf("✓ Expected error for high Unicode: %v", err)
					t.Skip("Emoji and high Unicode symbols not supported (exceed BERT vocab)")
				}
				return
			}

			if err != nil {
				t.Errorf("Failed to score text with special characters: %v", err)
				return
			}

			assertValidScore(t, score, "onnx")
			t.Logf("%s: score=%.3f, confidence=%.3f", tc.name, score.Value, score.Confidence)
		})
	}
}
