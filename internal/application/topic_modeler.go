package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// TopicModel represents a topic with keywords and weights.
type TopicModel struct {
	ID        string            `json:"id"`
	Label     string            `json:"label"`     // Human-readable label
	Keywords  []TopicKeyword    `json:"keywords"`  // Top keywords for this topic
	Documents []string          `json:"documents"` // Memory IDs belonging to this topic
	Coherence float64           `json:"coherence"` // Topic coherence score (0.0-1.0)
	Diversity float64           `json:"diversity"` // Keyword diversity score (0.0-1.0)
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// TopicKeyword represents a keyword within a topic.
type TopicKeyword struct {
	Word   string  `json:"word"`
	Weight float64 `json:"weight"` // Importance within topic (0.0-1.0)
}

// TopicDistribution represents how strongly a document belongs to each topic.
type TopicDistribution struct {
	MemoryID      string             `json:"memory_id"`
	TopicScores   map[string]float64 `json:"topic_scores"` // topic_id -> score
	DominantTopic string             `json:"dominant_topic"`
	Confidence    float64            `json:"confidence"`
}

// TopicModelingConfig configures topic modeling parameters.
type TopicModelingConfig struct {
	Algorithm        string  // "lda" or "nmf"
	NumTopics        int     // Number of topics to extract
	MaxIterations    int     // Maximum iterations for convergence
	MinWordFrequency int     // Minimum word frequency to include
	MaxWordFrequency float64 // Maximum word frequency (as percentage)
	TopKeywords      int     // Number of keywords per topic
	RandomSeed       int64   // For reproducibility
	Alpha            float64 // LDA hyperparameter
	Beta             float64 // LDA hyperparameter
	UseONNX          bool    // Use ONNX models if available
}

// DefaultTopicModelingConfig returns sensible defaults.
func DefaultTopicModelingConfig() TopicModelingConfig {
	return TopicModelingConfig{
		Algorithm:        "lda",
		NumTopics:        5,
		MaxIterations:    100,
		MinWordFrequency: 2,
		MaxWordFrequency: 0.8,
		TopKeywords:      10,
		RandomSeed:       42,
		Alpha:            0.1,
		Beta:             0.01,
		UseONNX:          true,
	}
}

// TopicModeler performs topic modeling on document collections.
type TopicModeler struct {
	config        TopicModelingConfig
	repository    ElementRepository
	modelProvider ONNXModelProvider
}

// NewTopicModeler creates a new topic modeler.
func NewTopicModeler(
	config TopicModelingConfig,
	repository ElementRepository,
	modelProvider ONNXModelProvider,
) *TopicModeler {
	return &TopicModeler{
		config:        config,
		repository:    repository,
		modelProvider: modelProvider,
	}
}

// ExtractTopics discovers topics from a collection of memories.
func (t *TopicModeler) ExtractTopics(ctx context.Context, memoryIDs []string) ([]TopicModel, error) {
	// Fetch memory contents
	texts := make([]string, 0, len(memoryIDs))
	for _, memoryID := range memoryIDs {
		element, err := t.repository.GetByID(memoryID)
		if err != nil {
			continue
		}

		memory, ok := element.(*domain.Memory)
		if !ok {
			continue
		}

		texts = append(texts, memory.Content)
	}

	if len(texts) == 0 {
		return nil, errors.New("no valid memories found")
	}

	// Try ONNX-based topic modeling if available
	if t.config.UseONNX && t.modelProvider.IsAvailable() {
		return t.extractWithONNX(ctx, texts, memoryIDs)
	}

	// Fallback to classical algorithms
	return t.extractWithClassical(texts, memoryIDs)
}

// extractWithONNX uses ONNX transformer models for topic extraction.
func (t *TopicModeler) extractWithONNX(ctx context.Context, texts []string, memoryIDs []string) ([]TopicModel, error) {
	topics, err := t.modelProvider.ExtractTopics(ctx, texts, t.config.NumTopics)
	if err != nil {
		return nil, fmt.Errorf("ONNX topic extraction failed: %w", err)
	}

	// Convert to TopicModel format
	models := make([]TopicModel, len(topics))
	for i, topic := range topics {
		keywords := make([]TopicKeyword, len(topic.Keywords))
		for j, kw := range topic.Keywords {
			keywords[j] = TopicKeyword{
				Word:   kw,
				Weight: topic.Weight, // Simplified - would need per-keyword weights
			}
		}

		models[i] = TopicModel{
			ID:        topic.ID,
			Label:     generateTopicLabel(topic.Keywords),
			Keywords:  keywords,
			Documents: []string{}, // Would need document assignment
			Coherence: 0.8,        // Placeholder
			Diversity: 0.7,        // Placeholder
		}
	}

	return models, nil
}

// extractWithClassical uses LDA or NMF for topic modeling.
func (t *TopicModeler) extractWithClassical(texts []string, memoryIDs []string) ([]TopicModel, error) {
	// Build vocabulary and document-term matrix
	vocab, dtm := t.buildDocumentTermMatrix(texts)

	if len(vocab) == 0 {
		return nil, errors.New("empty vocabulary after preprocessing")
	}

	// Run LDA or NMF
	var topicWordDist [][]float64
	var docTopicDist [][]float64
	var err error

	switch t.config.Algorithm {
	case "lda":
		topicWordDist, docTopicDist, err = t.runLDA(dtm, len(vocab))
	case "nmf":
		topicWordDist, docTopicDist, err = t.runNMF(dtm, len(vocab))
	default:
		return nil, fmt.Errorf("unknown algorithm: %s", t.config.Algorithm)
	}

	if err != nil {
		return nil, err
	}

	// Build topic models
	topics := make([]TopicModel, t.config.NumTopics)
	for i := range t.config.NumTopics {
		keywords := t.getTopKeywords(topicWordDist[i], vocab)
		docs := t.getTopDocuments(docTopicDist, i, memoryIDs)

		topics[i] = TopicModel{
			ID:        fmt.Sprintf("topic_%d", i+1),
			Label:     generateTopicLabel(extractWords(keywords)),
			Keywords:  keywords,
			Documents: docs,
			Coherence: t.calculateCoherence(keywords, dtm, vocab),
			Diversity: t.calculateDiversity(keywords),
		}
	}

	return topics, nil
}

// buildDocumentTermMatrix creates a vocabulary and document-term matrix.
func (t *TopicModeler) buildDocumentTermMatrix(texts []string) ([]string, [][]int) {
	// Tokenize and count words
	wordCounts := make(map[string]int)
	docWords := make([][]string, len(texts))

	for i, text := range texts {
		words := t.tokenize(text)
		docWords[i] = words

		for _, word := range words {
			wordCounts[word]++
		}
	}

	// Filter by frequency
	vocab := make([]string, 0)
	maxFreq := int(float64(len(texts)) * t.config.MaxWordFrequency)

	for word, count := range wordCounts {
		if count >= t.config.MinWordFrequency && count <= maxFreq {
			vocab = append(vocab, word)
		}
	}

	sort.Strings(vocab) // For consistent ordering

	// Build word index
	wordIndex := make(map[string]int)
	for i, word := range vocab {
		wordIndex[word] = i
	}

	// Create document-term matrix
	dtm := make([][]int, len(texts))
	for i, words := range docWords {
		dtm[i] = make([]int, len(vocab))
		for _, word := range words {
			if idx, ok := wordIndex[word]; ok {
				dtm[i][idx]++
			}
		}
	}

	return vocab, dtm
}

// tokenize performs simple word tokenization.
func (t *TopicModeler) tokenize(text string) []string {
	text = strings.ToLower(text)

	// Remove punctuation
	text = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || r == ' ' {
			return r
		}
		return -1
	}, text)

	// Split and filter
	words := strings.Fields(text)
	filtered := make([]string, 0, len(words))

	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"was": true, "are": true, "were": true, "be": true, "been": true,
		"as": true, "it": true, "this": true, "that": true, "from": true,
	}

	for _, word := range words {
		if len(word) > 2 && !stopWords[word] {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// runLDA performs Latent Dirichlet Allocation.
func (t *TopicModeler) runLDA(dtm [][]int, vocabSize int) ([][]float64, [][]float64, error) {
	numDocs := len(dtm)
	numTopics := t.config.NumTopics

	// Initialize matrices randomly
	topicWordDist := t.initializeMatrix(numTopics, vocabSize)
	docTopicDist := t.initializeMatrix(numDocs, numTopics)

	// Simplified LDA using Gibbs sampling approximation
	for range t.config.MaxIterations {
		for d := range numDocs {
			for w := range vocabSize {
				count := dtm[d][w]
				if count == 0 {
					continue
				}

				// Update topic assignments (simplified)
				for k := range numTopics {
					score := (docTopicDist[d][k] + t.config.Alpha) *
						(topicWordDist[k][w] + t.config.Beta) /
						(sum(topicWordDist[k]) + float64(vocabSize)*t.config.Beta)

					// Update distributions
					docTopicDist[d][k] += score * float64(count)
					topicWordDist[k][w] += score * float64(count)
				}
			}
		}

		// Normalize
		t.normalizeMatrix(topicWordDist)
		t.normalizeMatrix(docTopicDist)
	}

	return topicWordDist, docTopicDist, nil
}

// runNMF performs Non-negative Matrix Factorization.
func (t *TopicModeler) runNMF(dtm [][]int, vocabSize int) ([][]float64, [][]float64, error) {
	numDocs := len(dtm)
	numTopics := t.config.NumTopics

	// Convert DTM to float
	X := make([][]float64, numDocs)
	for i := range X {
		X[i] = make([]float64, vocabSize)
		for j := range dtm[i] {
			X[i][j] = float64(dtm[i][j])
		}
	}

	// Initialize W (doc-topic) and H (topic-word) matrices
	W := t.initializeMatrix(numDocs, numTopics)
	H := t.initializeMatrix(numTopics, vocabSize)

	// Multiplicative update rules
	for range t.config.MaxIterations {
		// Update H
		numerator := t.matrixMultiply(t.transpose(W), X)
		denominator := t.matrixMultiply(t.matrixMultiply(t.transpose(W), W), H)
		H = t.elementwiseDivide(t.elementwiseMultiply(H, numerator), denominator)

		// Update W
		numerator = t.matrixMultiply(X, t.transpose(H))
		denominator = t.matrixMultiply(W, t.matrixMultiply(H, t.transpose(H)))
		W = t.elementwiseDivide(t.elementwiseMultiply(W, numerator), denominator)
	}

	// Normalize
	t.normalizeMatrix(H)
	t.normalizeMatrix(W)

	return H, W, nil
}

// getTopKeywords extracts top keywords from topic distribution.
func (t *TopicModeler) getTopKeywords(dist []float64, vocab []string) []TopicKeyword {
	type wordScore struct {
		word  string
		score float64
	}

	scores := make([]wordScore, len(vocab))
	for i, word := range vocab {
		scores[i] = wordScore{word: word, score: dist[i]}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Take top N
	n := t.config.TopKeywords
	if n > len(scores) {
		n = len(scores)
	}

	keywords := make([]TopicKeyword, n)
	for i := range n {
		keywords[i] = TopicKeyword{
			Word:   scores[i].word,
			Weight: scores[i].score,
		}
	}

	return keywords
}

// getTopDocuments finds documents most strongly associated with a topic.
func (t *TopicModeler) getTopDocuments(docTopicDist [][]float64, topicIdx int, memoryIDs []string) []string {
	type docScore struct {
		id    string
		score float64
	}

	scores := make([]docScore, len(docTopicDist))
	for i := range docTopicDist {
		scores[i] = docScore{
			id:    memoryIDs[i],
			score: docTopicDist[i][topicIdx],
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return top 50% of documents
	threshold := 0.5
	docs := make([]string, 0)
	for _, s := range scores {
		if s.score >= threshold {
			docs = append(docs, s.id)
		}
	}

	return docs
}

// Helper functions for matrix operations

func (t *TopicModeler) initializeMatrix(rows, cols int) [][]float64 {
	matrix := make([][]float64, rows)
	for i := range matrix {
		matrix[i] = make([]float64, cols)
		for j := range matrix[i] {
			matrix[i][j] = 0.1 + float64(i+j)*0.01 // Simple initialization
		}
	}
	return matrix
}

func (t *TopicModeler) normalizeMatrix(matrix [][]float64) {
	for i := range matrix {
		total := sum(matrix[i])
		if total > 0 {
			for j := range matrix[i] {
				matrix[i][j] /= total
			}
		}
	}
}

func (t *TopicModeler) matrixMultiply(a, b [][]float64) [][]float64 {
	if len(a) == 0 || len(b) == 0 || len(a[0]) != len(b) {
		return nil
	}

	result := make([][]float64, len(a))
	for i := range result {
		result[i] = make([]float64, len(b[0]))
		for j := range result[i] {
			for k := range a[i] {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return result
}

func (t *TopicModeler) transpose(matrix [][]float64) [][]float64 {
	if len(matrix) == 0 {
		return nil
	}

	result := make([][]float64, len(matrix[0]))
	for i := range result {
		result[i] = make([]float64, len(matrix))
		for j := range result[i] {
			result[i][j] = matrix[j][i]
		}
	}
	return result
}

func (t *TopicModeler) elementwiseMultiply(a, b [][]float64) [][]float64 {
	result := make([][]float64, len(a))
	for i := range result {
		result[i] = make([]float64, len(a[i]))
		for j := range result[i] {
			result[i][j] = a[i][j] * b[i][j]
		}
	}
	return result
}

func (t *TopicModeler) elementwiseDivide(a, b [][]float64) [][]float64 {
	result := make([][]float64, len(a))
	for i := range result {
		result[i] = make([]float64, len(a[i]))
		for j := range result[i] {
			if b[i][j] != 0 {
				result[i][j] = a[i][j] / b[i][j]
			} else {
				result[i][j] = 0
			}
		}
	}
	return result
}

func (t *TopicModeler) calculateCoherence(keywords []TopicKeyword, dtm [][]int, vocab []string) float64 {
	// Simplified coherence calculation
	// Measures co-occurrence of keyword pairs

	wordIndex := make(map[string]int)
	for i, word := range vocab {
		wordIndex[word] = i
	}

	var totalCoherence float64
	count := 0

	for i := range len(keywords) - 1 {
		for j := i + 1; j < len(keywords); j++ {
			idx1, ok1 := wordIndex[keywords[i].Word]
			idx2, ok2 := wordIndex[keywords[j].Word]

			if !ok1 || !ok2 {
				continue
			}

			// Count co-occurrences
			cooccur := 0
			for _, doc := range dtm {
				if doc[idx1] > 0 && doc[idx2] > 0 {
					cooccur++
				}
			}

			totalCoherence += float64(cooccur) / float64(len(dtm))
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return totalCoherence / float64(count)
}

func (t *TopicModeler) calculateDiversity(keywords []TopicKeyword) float64 {
	// Measure lexical diversity using unique prefixes
	prefixes := make(map[string]bool)

	for _, kw := range keywords {
		if len(kw.Word) >= 3 {
			prefixes[kw.Word[:3]] = true
		}
	}

	return float64(len(prefixes)) / float64(len(keywords))
}

func sum(arr []float64) float64 {
	total := 0.0
	for _, v := range arr {
		total += v
	}
	return total
}

func generateTopicLabel(keywords []string) string {
	if len(keywords) == 0 {
		return "Unnamed Topic"
	}

	// Take first 3 keywords
	n := 3
	if len(keywords) < n {
		n = len(keywords)
	}

	return strings.Join(keywords[:n], ", ")
}

func extractWords(keywords []TopicKeyword) []string {
	words := make([]string, len(keywords))
	for i, kw := range keywords {
		words[i] = kw.Word
	}
	return words
}

// AnalyzeTopicTrends analyzes how topics evolve over time.
func (t *TopicModeler) AnalyzeTopicTrends(ctx context.Context, memoryIDs []string) (map[string][]float64, error) {
	// Group memories by time period
	// Extract topics for each period
	// Track topic prevalence over time

	// This is a placeholder for future implementation
	return nil, errors.New("topic trends analysis not yet implemented")
}

// AssignTopics assigns topics to new documents.
func (t *TopicModeler) AssignTopics(ctx context.Context, models []TopicModel, text string) (*TopicDistribution, error) {
	// Tokenize text
	words := t.tokenize(text)

	// Calculate similarity to each topic
	scores := make(map[string]float64)

	for _, topic := range models {
		// Count keyword matches
		matches := 0
		weightSum := 0.0

		for _, kw := range topic.Keywords {
			for _, word := range words {
				if word == kw.Word {
					matches++
					weightSum += kw.Weight
					break
				}
			}
		}

		// Score based on matches and weights
		if matches > 0 {
			scores[topic.ID] = weightSum / float64(len(topic.Keywords))
		}
	}

	// Find dominant topic
	var dominantTopic string
	var maxScore float64

	for id, score := range scores {
		if score > maxScore {
			maxScore = score
			dominantTopic = id
		}
	}

	// Normalize scores
	total := 0.0
	for _, score := range scores {
		total += score
	}

	if total > 0 {
		for id := range scores {
			scores[id] /= total
		}
	}

	return &TopicDistribution{
		MemoryID:      "",
		TopicScores:   scores,
		DominantTopic: dominantTopic,
		Confidence:    maxScore,
	}, nil
}
