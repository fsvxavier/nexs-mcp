//go:build !noonnx
// +build !noonnx

package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	ort "github.com/yalue/onnxruntime_go"
)

var (
	bertEnvInitialized bool
	bertEnvMutex       sync.Mutex
)

// ONNXBERTProvider implements ONNXModelProvider using ONNX Runtime for BERT-based models.
type ONNXBERTProvider struct {
	config EnhancedNLPConfig

	// ONNX sessions for different models
	entitySession    *ort.DynamicAdvancedSession
	sentimentSession *ort.DynamicAdvancedSession

	// Model metadata
	entityLabels    []string
	sentimentLabels []SentimentLabel

	// Tokenizer (simplified BPE/WordPiece)
	vocab      map[string]int
	vocabMutex sync.RWMutex

	// State
	initialized bool
	mu          sync.RWMutex
}

// NewONNXBERTProvider creates a new ONNX BERT provider.
func NewONNXBERTProvider(config EnhancedNLPConfig) (*ONNXBERTProvider, error) {
	provider := &ONNXBERTProvider{
		config: config,
		// CoNLL-2003 NER labels (BIO format)
		entityLabels: []string{
			"O",      // Outside
			"B-PER",  // Beginning of person
			"I-PER",  // Inside person
			"B-ORG",  // Beginning of organization
			"I-ORG",  // Inside organization
			"B-LOC",  // Beginning of location
			"I-LOC",  // Inside location
			"B-MISC", // Beginning of miscellaneous
			"I-MISC", // Inside miscellaneous
		},
		sentimentLabels: []SentimentLabel{
			SentimentNegative,
			SentimentNeutral,
			SentimentPositive,
		},
		vocab: make(map[string]int),
	}

	if err := provider.initialize(); err != nil {
		// If initialization fails, provider will fallback to unavailable
		return provider, nil
	}

	return provider, nil
}

// initialize sets up ONNX Runtime environment and loads models.
func (p *ONNXBERTProvider) initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return nil
	}

	// Initialize ONNX environment (once globally)
	bertEnvMutex.Lock()
	if !bertEnvInitialized {
		if err := ort.InitializeEnvironment(); err != nil {
			bertEnvMutex.Unlock()
			return fmt.Errorf("failed to initialize ONNX environment: %w", err)
		}
		bertEnvInitialized = true
	}
	bertEnvMutex.Unlock()

	// Load entity extraction model
	if err := p.loadEntityModel(); err != nil {
		return fmt.Errorf("failed to load entity model: %w", err)
	}

	// Load sentiment analysis model
	if err := p.loadSentimentModel(); err != nil {
		return fmt.Errorf("failed to load sentiment model: %w", err)
	}

	// Load tokenizer vocabulary
	if err := p.loadVocabulary(); err != nil {
		return fmt.Errorf("failed to load vocabulary: %w", err)
	}

	p.initialized = true
	return nil
}

// loadEntityModel loads the BERT NER model.
func (p *ONNXBERTProvider) loadEntityModel() error {
	if p.config.EntityModel == "" {
		return errors.New("entity model path not configured")
	}

	// Check if model file exists
	if _, err := os.Stat(p.config.EntityModel); err != nil {
		return fmt.Errorf("entity model not found: %w", err)
	}

	// Create ONNX session with dynamic input shapes
	// BERT NER requires input_ids, attention_mask, and token_type_ids
	var err error
	p.entitySession, err = ort.NewDynamicAdvancedSession(
		p.config.EntityModel,
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"logits"},
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create entity session: %w", err)
	}

	return nil
}

// loadSentimentModel loads the BERT sentiment analysis model.
func (p *ONNXBERTProvider) loadSentimentModel() error {
	if p.config.SentimentModel == "" {
		return errors.New("sentiment model path not configured")
	}

	// Check if model file exists
	if _, err := os.Stat(p.config.SentimentModel); err != nil {
		return fmt.Errorf("sentiment model not found: %w", err)
	}

	// Create ONNX session
	// DistilBERT doesn't use token_type_ids
	var err error
	p.sentimentSession, err = ort.NewDynamicAdvancedSession(
		p.config.SentimentModel,
		[]string{"input_ids", "attention_mask"},
		[]string{"logits"},
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create sentiment session: %w", err)
	}

	return nil
}

// loadVocabulary loads the tokenizer vocabulary.
func (p *ONNXBERTProvider) loadVocabulary() error {
	// Load vocab.txt from model directory
	modelDir := filepath.Dir(p.config.EntityModel)
	vocabPath := filepath.Join(modelDir, "vocab.txt")

	// Check if vocab file exists
	if _, err := os.Stat(vocabPath); err != nil {
		// Use simplified vocabulary if file not found
		return p.loadSimplifiedVocabulary()
	}

	// Read vocabulary file
	data, err := os.ReadFile(vocabPath)
	if err != nil {
		return fmt.Errorf("failed to read vocabulary: %w", err)
	}

	// Parse vocabulary (one token per line)
	lines := strings.Split(string(data), "\n")
	p.vocabMutex.Lock()
	defer p.vocabMutex.Unlock()

	for idx, token := range lines {
		token = strings.TrimSpace(token)
		if token != "" {
			p.vocab[token] = idx
		}
	}

	return nil
}

// loadSimplifiedVocabulary creates a minimal vocabulary for testing.
func (p *ONNXBERTProvider) loadSimplifiedVocabulary() error {
	// Special tokens
	specialTokens := []string{
		"[PAD]", "[UNK]", "[CLS]", "[SEP]", "[MASK]",
	}

	p.vocabMutex.Lock()
	defer p.vocabMutex.Unlock()

	idx := 0
	for _, token := range specialTokens {
		p.vocab[token] = idx
		idx++
	}

	// Add common words (simplified for fallback)
	// In production, this should be loaded from the model's vocab.txt
	return nil
}

// tokenize converts text to token IDs.
func (p *ONNXBERTProvider) tokenize(text string, maxLength int) ([]int64, []int64, []int64, error) {
	p.vocabMutex.RLock()
	defer p.vocabMutex.RUnlock()

	// Simple whitespace tokenization (simplified for basic support)
	// In production, use proper WordPiece/BPE tokenizer
	words := strings.Fields(strings.ToLower(text))

	// Get special token IDs
	clsID := int64(p.vocab["[CLS]"])
	sepID := int64(p.vocab["[SEP]"])
	padID := int64(p.vocab["[PAD]"])
	unkID := int64(p.vocab["[UNK]"])

	// Build token sequence: [CLS] + tokens + [SEP]
	inputIDs := []int64{clsID}
	for _, word := range words {
		if len(inputIDs) >= maxLength-1 {
			break
		}
		if id, ok := p.vocab[word]; ok {
			inputIDs = append(inputIDs, int64(id))
		} else {
			inputIDs = append(inputIDs, unkID)
		}
	}
	inputIDs = append(inputIDs, sepID)

	// Pad to maxLength
	seqLen := len(inputIDs)
	for len(inputIDs) < maxLength {
		inputIDs = append(inputIDs, padID)
	}

	// Create attention mask (1 for real tokens, 0 for padding)
	attentionMask := make([]int64, maxLength)
	for i := range seqLen {
		attentionMask[i] = 1
	}

	// Create token type IDs (all 0 for single sequence)
	tokenTypeIDs := make([]int64, maxLength)

	return inputIDs, attentionMask, tokenTypeIDs, nil
}

// ExtractEntities runs NER model on text.
func (p *ONNXBERTProvider) ExtractEntities(ctx context.Context, text string) ([]EnhancedEntity, error) {
	if !p.IsAvailable() {
		return nil, errors.New("ONNX BERT provider not available")
	}

	startTime := time.Now()

	// Tokenize input
	inputIDs, attentionMask, tokenTypeIDs, err := p.tokenize(text, p.config.MaxLength)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}

	// Prepare input tensors
	inputShape := ort.NewShape(1, int64(len(inputIDs))) // Batch size 1
	inputTensor, err := ort.NewTensor(inputShape, inputIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer func() { _ = inputTensor.Destroy() }()

	attentionTensor, err := ort.NewTensor(inputShape, attentionMask)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention tensor: %w", err)
	}
	defer func() { _ = attentionTensor.Destroy() }()

	tokenTypeTensor, err := ort.NewTensor(inputShape, tokenTypeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create token type tensor: %w", err)
	}
	defer func() { _ = tokenTypeTensor.Destroy() }()

	// Create output tensor
	outputShape := ort.NewShape(1, int64(len(inputIDs)), int64(len(p.entityLabels)))
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer func() { _ = outputTensor.Destroy() }()

	// Run inference with all three inputs
	p.mu.RLock()
	err = p.entitySession.Run(
		[]ort.Value{inputTensor, attentionTensor, tokenTypeTensor},
		[]ort.Value{outputTensor},
	)
	p.mu.RUnlock()

	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// Extract logits from output
	logits := outputTensor.GetData()

	// Convert logits to entities
	entities := p.logitsToEntities(logits, text, attentionMask)

	// Add processing time metadata
	processingTime := time.Since(startTime).Milliseconds()
	for i := range entities {
		if entities[i].Metadata == nil {
			entities[i].Metadata = make(map[string]string)
		}
		entities[i].Metadata["processing_time_ms"] = strconv.FormatInt(processingTime, 10)
		entities[i].Metadata["model"] = "onnx-bert"
	}

	return entities, nil
}

// ExtractEntitiesBatch processes multiple texts in batch.
func (p *ONNXBERTProvider) ExtractEntitiesBatch(ctx context.Context, texts []string) ([][]EnhancedEntity, error) {
	if !p.IsAvailable() {
		return nil, errors.New("ONNX BERT provider not available")
	}

	// TODO: Implement true batch inference for better performance
	// For now, process sequentially
	results := make([][]EnhancedEntity, len(texts))
	for i, text := range texts {
		entities, err := p.ExtractEntities(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to process text %d: %w", i, err)
		}
		results[i] = entities
	}

	return results, nil
}

// AnalyzeSentiment runs sentiment analysis model.
func (p *ONNXBERTProvider) AnalyzeSentiment(ctx context.Context, text string) (*SentimentResult, error) {
	if !p.IsAvailable() {
		return nil, errors.New("ONNX BERT provider not available")
	}

	startTime := time.Now()

	// Tokenize input (DistilBERT sentiment doesn't use token_type_ids)
	inputIDs, attentionMask, _, err := p.tokenize(text, p.config.MaxLength)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}

	// Prepare input tensors
	inputShape := ort.NewShape(1, int64(len(inputIDs)))
	inputTensor, err := ort.NewTensor(inputShape, inputIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer func() { _ = inputTensor.Destroy() }()

	attentionTensor, err := ort.NewTensor(inputShape, attentionMask)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention tensor: %w", err)
	}
	defer func() { _ = attentionTensor.Destroy() }()

	// Create output tensor
	outputShape := ort.NewShape(1, int64(len(p.sentimentLabels)))
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer func() { _ = outputTensor.Destroy() }()

	// Run inference (without token_type_ids for DistilBERT compatibility)
	p.mu.RLock()
	err = p.sentimentSession.Run(
		[]ort.Value{inputTensor, attentionTensor},
		[]ort.Value{outputTensor},
	)
	p.mu.RUnlock()

	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// Extract logits
	logits := outputTensor.GetData()

	// Convert logits to sentiment result
	result := p.logitsToSentiment(logits)
	result.ProcessingTime = float64(time.Since(startTime).Milliseconds())
	result.ModelUsed = "onnx-bert-sentiment"

	return result, nil
}

// ExtractTopics is not implemented for BERT (use TopicModeler instead).
func (p *ONNXBERTProvider) ExtractTopics(ctx context.Context, texts []string, numTopics int) ([]Topic, error) {
	return nil, errors.New("topic extraction not supported by BERT provider, use TopicModeler")
}

// IsAvailable checks if ONNX runtime and models are loaded.
func (p *ONNXBERTProvider) IsAvailable() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.initialized && p.entitySession != nil && p.sentimentSession != nil
}

// logitsToEntities converts model logits to entity objects using BIO format.
func (p *ONNXBERTProvider) logitsToEntities(logits []float32, text string, attentionMask []int64) []EnhancedEntity {
	// Logits shape: [seq_len, num_labels]
	// Process BIO format: B-TYPE (beginning), I-TYPE (inside), O (outside)

	words := strings.Fields(text)
	entities := make([]EnhancedEntity, 0)

	numLabels := len(p.entityLabels)
	seqLen := len(logits) / numLabels

	var currentEntity *EnhancedEntity
	currentPos := 0

	for i := 1; i < seqLen-1 && i-1 < len(words); i++ { // Skip [CLS] and [SEP]
		if attentionMask[i] == 0 {
			break // Reached padding
		}

		// Get logits for this token
		tokenLogits := logits[i*numLabels : (i+1)*numLabels]

		// Apply softmax and get predicted label
		probs := softmax(tokenLogits)
		labelIdx, confidence := argmax(probs)

		labelStr := p.entityLabels[labelIdx]

		// Skip if confidence too low or label is "O" (outside)
		if float64(confidence) < p.config.EntityConfidenceMin || labelStr == "O" {
			if currentEntity != nil {
				entities = append(entities, *currentEntity)
				currentEntity = nil
			}
			currentPos += len(words[i-1]) + 1
			continue
		}

		// Parse BIO label: B-PER, I-PER, etc.
		var prefix string // B or I
		var entityTypeStr string
		if len(labelStr) >= 3 && labelStr[1] == '-' {
			prefix = string(labelStr[0])
			entityTypeStr = labelStr[2:]
		} else {
			// Fallback for non-BIO format
			prefix = "B"
			entityTypeStr = labelStr
		}

		// Map CoNLL labels to our EntityType
		var entityType EntityType
		switch entityTypeStr {
		case "PER":
			entityType = EntityTypePerson
		case "ORG":
			entityType = EntityTypeOrganization
		case "LOC":
			entityType = EntityTypeLocation
		case "MISC":
			entityType = EntityTypeConcept
		default:
			entityType = EntityTypeConcept
		}

		// Handle BIO logic
		if prefix == "B" || currentEntity == nil || currentEntity.Type != entityType {
			// Start new entity (B- tag or type change)
			if currentEntity != nil {
				entities = append(entities, *currentEntity)
			}
			currentEntity = &EnhancedEntity{
				Type:       entityType,
				Value:      words[i-1],
				Confidence: float64(confidence),
				StartPos:   currentPos,
				EndPos:     currentPos + len(words[i-1]),
				Context:    text,
				Metadata:   make(map[string]string),
			}
		} else if prefix == "I" && currentEntity != nil && currentEntity.Type == entityType {
			// Continue entity (I- tag with matching type)
			currentEntity.Value += " " + words[i-1]
			currentEntity.EndPos = currentPos + len(words[i-1])
			currentEntity.Confidence = (currentEntity.Confidence + float64(confidence)) / 2.0
		}

		currentPos += len(words[i-1]) + 1
	}

	// Add final entity
	if currentEntity != nil {
		entities = append(entities, *currentEntity)
	}

	return entities
}

// logitsToSentiment converts sentiment logits to SentimentResult.
func (p *ONNXBERTProvider) logitsToSentiment(logits []float32) *SentimentResult {
	// Apply softmax to get probabilities
	probs := softmax(logits[:len(p.sentimentLabels)])

	// Get predicted label
	labelIdx, confidence := argmax(probs)
	label := p.sentimentLabels[labelIdx]

	// Build sentiment scores
	scores := SentimentScores{
		Negative: float64(probs[0]),
		Neutral:  float64(probs[1]),
		Positive: float64(probs[2]),
	}

	// Calculate intensity (distance from neutral)
	intensity := math.Abs(float64(probs[2]) - float64(probs[0]))

	// Estimate emotional tone (simplified heuristics)
	emotionalTone := EmotionalTone{
		Joy:      float64(probs[2]) * 0.8, // High for positive
		Sadness:  float64(probs[0]) * 0.6, // High for negative
		Anger:    float64(probs[0]) * 0.3,
		Fear:     float64(probs[0]) * 0.2,
		Surprise: intensity * 0.3,
		Disgust:  float64(probs[0]) * 0.2,
	}

	return &SentimentResult{
		Label:             label,
		Confidence:        float64(confidence),
		Scores:            scores,
		Intensity:         intensity,
		EmotionalTone:     emotionalTone,
		SubjectivityScore: intensity, // Simplified: higher intensity = more subjective
	}
}

// softmax applies softmax function to logits.
func softmax(logits []float32) []float32 {
	// Find max for numerical stability
	maxLogit := logits[0]
	for _, l := range logits[1:] {
		if l > maxLogit {
			maxLogit = l
		}
	}

	// Compute exp(x - max)
	expSum := float32(0.0)
	probs := make([]float32, len(logits))
	for i, l := range logits {
		probs[i] = float32(math.Exp(float64(l - maxLogit)))
		expSum += probs[i]
	}

	// Normalize
	for i := range probs {
		probs[i] /= expSum
	}

	return probs
}

// argmax returns index and value of maximum element.
func argmax(probs []float32) (int, float32) {
	maxIdx := 0
	maxVal := probs[0]
	for i, p := range probs[1:] {
		if p > maxVal {
			maxIdx = i + 1
			maxVal = p
		}
	}
	return maxIdx, maxVal
}

// Close releases ONNX resources.
func (p *ONNXBERTProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error

	if p.entitySession != nil {
		if err := p.entitySession.Destroy(); err != nil {
			errs = append(errs, fmt.Errorf("failed to destroy entity session: %w", err))
		}
		p.entitySession = nil
	}

	if p.sentimentSession != nil {
		if err := p.sentimentSession.Destroy(); err != nil {
			errs = append(errs, fmt.Errorf("failed to destroy sentiment session: %w", err))
		}
		p.sentimentSession = nil
	}

	p.initialized = false

	if len(errs) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errs)
	}

	return nil
}
