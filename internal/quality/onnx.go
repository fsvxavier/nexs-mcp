//go:build !noonnx
// +build !noonnx

package quality

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	ort "github.com/yalue/onnxruntime_go"
)

var (
	onnxEnvInitialized bool
	onnxEnvMutex       sync.Mutex
)

// initializeONNXEnvironment ensures ONNX environment is initialized only once.
func initializeONNXEnvironment() error {
	onnxEnvMutex.Lock()
	defer onnxEnvMutex.Unlock()

	if onnxEnvInitialized {
		return nil
	}

	if err := ort.InitializeEnvironment(); err != nil {
		return fmt.Errorf("failed to initialize ONNX environment: %w", err)
	}

	onnxEnvInitialized = true
	return nil
}

// ONNXScorer uses ONNX model for quality scoring.
type ONNXScorer struct {
	config       *Config
	session      *ort.DynamicAdvancedSession // Changed to Dynamic to avoid pre-allocated tensors
	mu           sync.RWMutex
	initialized  bool
	modelPath    string
	modelType    string // "reranker" or "embedder"
	outputName   string // "logits", "last_hidden_state", etc.
	embeddingDim int    // Output dimension for embedders (384, 512, 768, etc.)
	// Input names dynamically determined based on model requirements
	inputNames []string
}

// NewONNXScorer creates a new ONNX-based quality scorer.
func NewONNXScorer(config *Config) (*ONNXScorer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Set defaults for model type and output if not specified
	modelType := config.ONNXModelType
	if modelType == "" {
		modelType = ModelTypeReranker // Default to reranker for backward compatibility
	}

	outputName := config.ONNXOutputName
	if outputName == "" {
		if modelType == ModelTypeReranker {
			outputName = "logits"
		} else {
			outputName = "last_hidden_state" // Sentence transformers output token embeddings
		}
	}

	// Determine embedding dimension from output shape
	embeddingDim := 1 // Default for reranker
	switch {
	case modelType == ModelTypeEmbedder && len(config.ONNXOutputShape) >= 3:
		// For embedders with shape [batch, seq_len, hidden_dim], use hidden_dim (index 2)
		embeddingDim = int(config.ONNXOutputShape[2])
	case len(config.ONNXOutputShape) >= 2:
		// For rerankers or other models, use second dimension
		embeddingDim = int(config.ONNXOutputShape[1])
	case modelType == ModelTypeEmbedder:
		embeddingDim = 384 // Default for embedder
	}

	scorer := &ONNXScorer{
		config:       config,
		modelPath:    config.ONNXModelPath,
		modelType:    modelType,
		outputName:   outputName,
		embeddingDim: embeddingDim,
	}

	// Initialize ONNX runtime
	if err := scorer.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize ONNX scorer: %w", err)
	}

	return scorer, nil
}

// initialize loads the ONNX model and creates session.
func (s *ONNXScorer) initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	// Check if model file exists
	if _, err := os.Stat(s.modelPath); os.IsNotExist(err) {
		return fmt.Errorf("ONNX model not found at %s", s.modelPath)
	}

	// Initialize ONNX runtime library (singleton)
	if err := initializeONNXEnvironment(); err != nil {
		return err
	}

	// Determine input names based on model requirements
	var inputNames []string
	inputNames = append(inputNames, "input_ids", "attention_mask")
	if s.config.RequiresTokenTypeIds {
		inputNames = append(inputNames, "token_type_ids")
	}
	s.inputNames = inputNames

	// Create ONNX session using DynamicAdvancedSession (no pre-allocated tensors)
	if s.config.DebugMode {
		fmt.Printf("\n[DEBUG] Creating DynamicAdvancedSession for: %s\n", s.modelPath)
		fmt.Printf("[DEBUG] Input names: %v\n", inputNames)
		fmt.Printf("[DEBUG] Output name: %s\n", s.outputName)
		fmt.Printf("[DEBUG] Model type: %s\n", s.modelType)
		fmt.Printf("[DEBUG] Embedding dim: %d\n", s.embeddingDim)
	}

	session, err := ort.NewDynamicAdvancedSession(
		s.modelPath,
		inputNames,             // Dynamic input names
		[]string{s.outputName}, // Dynamic output name
		nil,                    // Use default session options
	)
	if err != nil {
		return fmt.Errorf("failed to create ONNX session: %w", err)
	}

	if s.config.DebugMode {
		fmt.Printf("[DEBUG] DynamicAdvancedSession created successfully!\n\n")
	}

	s.session = session
	s.initialized = true

	return nil
}

// Score calculates quality using ONNX model.
func (s *ONNXScorer) Score(ctx context.Context, content string) (*Score, error) {
	s.mu.RLock()
	if !s.initialized {
		s.mu.RUnlock()
		return nil, errors.New("ONNX scorer not initialized")
	}
	s.mu.RUnlock()

	// Tokenize and encode the content
	tokenIDs, err := s.encodeContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to encode content: %w", err)
	}

	// Run inference
	startTime := time.Now()
	qualityScore, confidence, err := s.runInference(tokenIDs)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}
	latency := time.Since(startTime)

	return &Score{
		Value:      qualityScore,
		Confidence: confidence,
		Method:     "onnx",
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"model":       "ms-marco-MiniLM-L-6-v2",
			"latency_ms":  latency.Milliseconds(),
			"content_len": len(content),
		},
	}, nil
}

// ScoreBatch scores multiple contents efficiently.
func (s *ONNXScorer) ScoreBatch(ctx context.Context, contents []string) ([]*Score, error) {
	s.mu.RLock()
	if !s.initialized {
		s.mu.RUnlock()
		return nil, errors.New("ONNX scorer not initialized")
	}
	s.mu.RUnlock()

	scores := make([]*Score, len(contents))
	for i, content := range contents {
		score, err := s.Score(ctx, content)
		if err != nil {
			return nil, fmt.Errorf("failed to score content %d: %w", i, err)
		}
		scores[i] = score
	}

	return scores, nil
}

// encodeContent converts text to token IDs for the model.
func (s *ONNXScorer) encodeContent(content string) ([]int64, error) {
	// Simple tokenization: convert to token IDs
	// In production, use proper tokenizer (e.g., BPE, WordPiece)
	// For now, use basic character-level encoding

	maxLength := 512
	tokenIDs := make([]int64, maxLength)

	// Convert content to token IDs (character codes)
	runes := []rune(content)
	for i := 0; i < len(runes) && i < maxLength; i++ {
		// Use character code as token ID
		tokenIDs[i] = int64(runes[i])
	}

	return tokenIDs, nil
}

// runInference executes the ONNX model with dynamically created tensors.
func (s *ONNXScorer) runInference(tokenIDs []int64) (float64, float64, error) {
	// Create input tensors for this inference call
	inputShape := ort.NewShape(1, 512)

	// Create input_ids tensor
	inputTensor, err := ort.NewTensor(inputShape, tokenIDs)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer func() { _ = inputTensor.Destroy() }()

	// Create attention_mask tensor (1 where there's content, 0 for padding)
	maskData := make([]int64, 512)
	for i := 0; i < len(tokenIDs) && i < 512; i++ {
		if tokenIDs[i] != 0 {
			maskData[i] = 1
		} else {
			maskData[i] = 0
		}
	}
	attentionMask, err := ort.NewTensor(inputShape, maskData)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create attention mask: %w", err)
	}
	defer func() { _ = attentionMask.Destroy() }()

	// Prepare input values
	var inputs []ort.Value
	inputs = append(inputs, inputTensor, attentionMask)

	// Add token_type_ids if required (BERT models)
	if s.config.RequiresTokenTypeIds {
		typeIDsData := make([]int64, 512)
		for i := range typeIDsData {
			typeIDsData[i] = 0
		}
		tokenTypeIDs, err := ort.NewTensor(inputShape, typeIDsData)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to create token type IDs: %w", err)
		}
		defer func() { _ = tokenTypeIDs.Destroy() }()
		inputs = append(inputs, tokenTypeIDs)
	}

	// Create output tensor with dynamic shape
	var outputShape ort.Shape
	if s.modelType == ModelTypeReranker {
		outputShape = ort.NewShape(1, 1)
	} else {
		// Embedder: last_hidden_state has shape [batch, seq_len, hidden_dim]
		// We'll extract [CLS] token (first token) embedding after inference
		outputShape = ort.NewShape(1, 512, int64(s.embeddingDim))
	}
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer func() { _ = outputTensor.Destroy() }()

	// Run inference with dynamic tensors
	err = s.session.Run(inputs, []ort.Value{outputTensor})
	if err != nil {
		return 0, 0, fmt.Errorf("inference execution failed: %w", err)
	}

	// Get output data
	outputData := outputTensor.GetData()
	if len(outputData) == 0 {
		return 0, 0, errors.New("empty output data")
	}

	var qualityScore float64

	// Process output based on model type
	switch s.modelType {
	case "reranker":
		// Cross-encoder: single score output
		// MS MARCO model outputs logits, apply sigmoid to get 0-1 probability
		rawScore := float64(outputData[0])
		qualityScore = 1.0 / (1.0 + math.Exp(-rawScore/10.0))

	case "embedder":
		// Sentence transformer: last_hidden_state output (token embeddings)
		// Shape: [batch=1, seq_len=512, hidden_dim]
		// Extract [CLS] token embedding (first token, index 0)
		// [CLS] is at position 0, so we extract outputData[0:embeddingDim]
		if len(outputData) < s.embeddingDim {
			return 0, 0, fmt.Errorf("insufficient output data: got %d, need %d", len(outputData), s.embeddingDim)
		}

		// Extract [CLS] token embedding (first hidden_dim values)
		clsEmbedding := outputData[0:s.embeddingDim]

		// Calculate L2 norm of [CLS] embedding
		var sumSquares float64
		for _, val := range clsEmbedding {
			sumSquares += float64(val) * float64(val)
		}
		magnitude := math.Sqrt(sumSquares)

		// Normalize to 0-1 range (typical embedding magnitude is 1-10)
		qualityScore = magnitude / 10.0
		if qualityScore > 1.0 {
			qualityScore = 1.0
		}

	default:
		return 0, 0, fmt.Errorf("unknown model type: %s", s.modelType)
	}

	// Clamp to valid range
	if qualityScore < 0 {
		qualityScore = 0
	}
	if qualityScore > 1 {
		qualityScore = 1
	}

	// ONNX model has high confidence (0.85-0.95)
	confidence := 0.9

	// Adjust confidence based on score extremes
	if qualityScore < 0.1 || qualityScore > 0.9 {
		confidence = 0.95 // Very confident on extreme values
	}

	return qualityScore, confidence, nil
}

// computeEmbedding generates embedding vector for sentence transformers.
func (s *ONNXScorer) computeEmbedding(tokenIDs []int64) ([]float32, error) {
	// Create input tensors dynamically
	inputShape := ort.NewShape(1, 512)

	inputTensor, err := ort.NewTensor(inputShape, tokenIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer func() { _ = inputTensor.Destroy() }()

	// Create attention mask
	maskData := make([]int64, 512)
	for i := 0; i < len(tokenIDs) && i < 512; i++ {
		if tokenIDs[i] != 0 {
			maskData[i] = 1
		} else {
			maskData[i] = 0
		}
	}
	attentionMask, err := ort.NewTensor(inputShape, maskData)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention mask: %w", err)
	}
	defer func() { _ = attentionMask.Destroy() }()

	// Prepare inputs
	var inputs []ort.Value
	inputs = append(inputs, inputTensor, attentionMask)

	// Add token_type_ids if required
	if s.config.RequiresTokenTypeIds {
		typeIDsData := make([]int64, 512)
		for i := range typeIDsData {
			typeIDsData[i] = 0
		}
		tokenTypeIDs, err := ort.NewTensor(inputShape, typeIDsData)
		if err != nil {
			return nil, fmt.Errorf("failed to create token type IDs: %w", err)
		}
		defer func() { _ = tokenTypeIDs.Destroy() }()
		inputs = append(inputs, tokenTypeIDs)
	}

	// Create output tensor for last_hidden_state [batch, seq_len, hidden_dim]
	outputShape := ort.NewShape(1, 512, int64(s.embeddingDim))
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer func() { _ = outputTensor.Destroy() }()

	// Run inference
	err = s.session.Run(inputs, []ort.Value{outputTensor})
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// Get output embedding and extract [CLS] token (first token)
	outputData := outputTensor.GetData()
	if len(outputData) < s.embeddingDim {
		return nil, fmt.Errorf("insufficient output data: got %d, need %d", len(outputData), s.embeddingDim)
	}

	// Extract [CLS] embedding (first hidden_dim values)
	embedding := make([]float32, s.embeddingDim)
	copy(embedding, outputData[0:s.embeddingDim])

	return embedding, nil
}

// ScoreWithQuery calculates similarity between query and passage (for sentence transformers).
func (s *ONNXScorer) ScoreWithQuery(ctx context.Context, query, passage string) (*Score, error) {
	s.mu.RLock()
	if !s.initialized {
		s.mu.RUnlock()
		return nil, errors.New("ONNX scorer not initialized")
	}
	s.mu.RUnlock()

	// For reranker models, use standard Score method (concatenated text)
	if s.modelType == "reranker" {
		return s.Score(ctx, passage)
	}

	// For embedder models, compute separate embeddings and calculate similarity
	queryTokens, err := s.encodeContent(query)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	passageTokens, err := s.encodeContent(passage)
	if err != nil {
		return nil, fmt.Errorf("failed to encode passage: %w", err)
	}

	// Generate embeddings
	queryEmb, err := s.computeEmbedding(queryTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to compute query embedding: %w", err)
	}

	passageEmb, err := s.computeEmbedding(passageTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to compute passage embedding: %w", err)
	}

	// Calculate cosine similarity
	similarity := cosineSimilarity(queryEmb, passageEmb)

	// Normalize similarity from [-1, 1] to [0, 1]
	score := (similarity + 1.0) / 2.0

	return &Score{
		Value:      score,
		Confidence: 0.9,
		Method:     "onnx-embedder",
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"model_type": s.modelType,
			"similarity": similarity,
		},
	}, nil
}

// cosineSimilarity computes cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Name returns the scorer identifier.
func (s *ONNXScorer) Name() string {
	return "onnx"
}

// IsAvailable checks if ONNX scorer is available.
func (s *ONNXScorer) IsAvailable(ctx context.Context) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return false
	}

	// Check if model file still exists
	if _, err := os.Stat(s.modelPath); os.IsNotExist(err) {
		return false
	}

	return true
}

// Cost returns the computational cost (CPU time in arbitrary units).
func (s *ONNXScorer) Cost() float64 {
	// Estimated cost based on inference time
	// CPU: ~75ms average, GPU: ~15ms average
	// Using CPU estimate as baseline
	return 0.075 // Cost in seconds
}

// Close releases ONNX resources.
func (s *ONNXScorer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clean up ONNX session
	if s.session != nil {
		_ = s.session.Destroy()
		s.session = nil
	}

	s.initialized = false
	return nil
}
