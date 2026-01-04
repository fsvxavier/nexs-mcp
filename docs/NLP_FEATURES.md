# NLP Features (Sprint 18 - v1.4.0)

**Release Date:** January 4, 2026
**Status:** Production

## Overview

The NEXS-MCP system includes advanced Natural Language Processing (NLP) capabilities for enhanced memory analysis, entity extraction, sentiment tracking, and topic modeling. These features leverage both state-of-the-art transformer models (via ONNX) and classical algorithms with robust fallback mechanisms.

**Sprint 18 Additions:**
- 4 new NLP services: ONNXBERTProvider, EnhancedEntityExtractor, SentimentAnalyzer, TopicModeler
- 6 new MCP tools for NLP operations
- 2,499 LOC implementation + 2,350 LOC tests
- CPU performance: 100-200ms entity extraction, 50-100ms sentiment analysis
- ONNX models: BERT NER (411 MB), DistilBERT Sentiment (516 MB)

## Features

### 1. Enhanced Entity Extraction

Extract named entities with confidence scores and relationship detection.

**Entity Types:**
- `PERSON` - People and characters
- `ORGANIZATION` - Companies, institutions, groups
- `LOCATION` - Geographic locations
- `DATE` - Temporal expressions
- `EVENT` - Named events
- `PRODUCT` - Products and services
- `TECHNOLOGY` - Technologies and tools
- `CONCEPT` - Abstract concepts
- `OTHER` - Unclassified entities

**Relationship Types:**
- `WORKS_AT` - Employment relationships
- `FOUNDED` - Founding relationships
- `LOCATED_IN` - Geographic relationships
- `BORN_IN` - Birth location
- `LIVES_IN` - Residence
- `HEADQUARTERED_IN` - Organization location
- `DEVELOPED_BY` - Technology development
- `USED_BY` - Technology usage
- `AFFILIATED_WITH` - Affiliations
- `RELATED_TO` - General relationships

**Configuration:**
```bash
# Enable entity extraction (requires ONNX models)
export NEXS_NLP_ENTITY_EXTRACTION_ENABLED=false

# Model path (default: models/bert-base-ner.onnx)
export NEXS_NLP_ENTITY_MODEL="models/bert-base-ner/model.onnx"

# Minimum confidence threshold (0.0-1.0)
export NEXS_NLP_ENTITY_CONFIDENCE_MIN=0.7

# Maximum entities per document
export NEXS_NLP_ENTITY_MAX_PER_DOC=100
```

**MCP Tool:**
```json
{
  "tool": "extract_entities_advanced",
  "input": {
    "text": "John Smith works at Google in Mountain View",
    "memory_ids": ["mem-123"]
  }
}
```

**Response:**
```json
{
  "entities": [
    {
      "type": "PERSON",
      "value": "John Smith",
      "confidence": 0.95,
      "start_pos": 0,
      "end_pos": 10,
      "context": "John Smith works at Google"
    },
    {
      "type": "ORGANIZATION",
      "value": "Google",
      "confidence": 0.92,
      "start_pos": 21,
      "end_pos": 27
    }
  ],
  "relationships": [
    {
      "source_entity": "John Smith",
      "target_entity": "Google",
      "relation_type": "WORKS_AT",
      "confidence": 0.88,
      "evidence": "works at"
    }
  ],
  "processing_time": 0.145,
  "model_used": "bert-base-ner",
  "confidence": 0.935
}
```

**Fallback Mechanism:**
When ONNX models are unavailable, the system falls back to rule-based extraction using regex patterns for basic entity recognition with confidence=0.5.

### 2. Sentiment Analysis

Analyze sentiment with emotional tone tracking and trend detection.

**Sentiment Labels:**
- `POSITIVE` - Positive sentiment
- `NEGATIVE` - Negative sentiment
- `NEUTRAL` - Neutral sentiment
- `MIXED` - Mixed sentiment

**Emotional Dimensions:**
- Joy
- Sadness
- Anger
- Fear
- Surprise
- Disgust

**Configuration:**
```bash
# Enable sentiment analysis (requires ONNX models)
export NEXS_NLP_SENTIMENT_ENABLED=false

# Model path (default: models/distilbert-sentiment.onnx)
export NEXS_NLP_SENTIMENT_MODEL="models/distilbert-sentiment.onnx"

# Confidence threshold (0.0-1.0)
export NEXS_NLP_SENTIMENT_THRESHOLD=0.6
```

**MCP Tools:**

#### analyze_sentiment
```json
{
  "tool": "analyze_sentiment",
  "input": {
    "text": "This is an amazing product! I love it!",
    "memory_id": "mem-123"
  }
}
```

**Response:**
```json
{
  "label": "POSITIVE",
  "confidence": 0.96,
  "scores": {
    "positive": 0.96,
    "negative": 0.02,
    "neutral": 0.02
  },
  "intensity": 0.94,
  "emotional_tone": {
    "joy": 0.85,
    "sadness": 0.05,
    "anger": 0.02,
    "fear": 0.01,
    "surprise": 0.05,
    "disgust": 0.02
  },
  "subjectivity_score": 0.88,
  "processing_time": 0.098,
  "model_used": "distilbert-sentiment"
}
```

#### analyze_sentiment_trend
Track sentiment evolution over time with moving averages.

```json
{
  "tool": "analyze_sentiment_trend",
  "input": {
    "memory_ids": ["mem-1", "mem-2", "mem-3", "mem-4", "mem-5"]
  }
}
```

**Response:**
```json
{
  "trend": [
    {
      "timestamp": "2024-01-15T10:00:00Z",
      "sentiment": "POSITIVE",
      "confidence": 0.85,
      "moving_avg": 0.85
    },
    {
      "timestamp": "2024-01-15T11:00:00Z",
      "sentiment": "NEUTRAL",
      "confidence": 0.72,
      "moving_avg": 0.785
    }
  ],
  "overall_trend": "stable"
}
```

#### detect_emotional_shifts
Detect significant emotional changes between memories.

```json
{
  "tool": "detect_emotional_shifts",
  "input": {
    "memory_ids": ["mem-1", "mem-2", "mem-3"],
    "threshold": 0.3
  }
}
```

**Response:**
```json
{
  "shifts": [
    {
      "timestamp": "2024-01-15T11:00:00Z",
      "from_sentiment": "POSITIVE",
      "to_sentiment": "NEGATIVE",
      "magnitude": 0.45,
      "direction": "negative"
    }
  ],
  "total_shifts": 1
}
```

#### summarize_sentiment
Generate aggregate sentiment statistics.

```json
{
  "tool": "summarize_sentiment",
  "input": {
    "memory_ids": ["mem-1", "mem-2", "mem-3", "mem-4", "mem-5"]
  }
}
```

**Response:**
```json
{
  "total_memories": 5,
  "positive_count": 3,
  "negative_count": 1,
  "neutral_count": 1,
  "average_intensity": 0.72,
  "dominant_sentiment": "POSITIVE",
  "sentiment_score": 0.58,
  "emotional_profile": {
    "joy": 0.65,
    "sadness": 0.15,
    "anger": 0.05,
    "fear": 0.03,
    "surprise": 0.08,
    "disgust": 0.04
  }
}
```

**Fallback Mechanism:**
When ONNX models are unavailable, the system uses lexicon-based sentiment analysis with positive/negative word lists, returning confidence scores of 0.4-0.5.

### 3. Topic Modeling

Extract topics using LDA (Latent Dirichlet Allocation) or NMF (Non-negative Matrix Factorization) algorithms.

**Configuration:**
```bash
# Enable topic modeling (works without ONNX)
export NEXS_NLP_TOPIC_MODELING_ENABLED=true

# Number of topics to extract
export NEXS_NLP_TOPIC_COUNT=5

# Algorithm: "lda" or "nmf"
export NEXS_NLP_TOPIC_ALGORITHM=lda
```

**MCP Tool:**
```json
{
  "tool": "extract_topics",
  "input": {
    "memory_ids": ["mem-1", "mem-2", "mem-3"],
    "num_topics": 5,
    "algorithm": "lda"
  }
}
```

**Response:**
```json
{
  "topics": [
    {
      "id": "topic-0",
      "label": "Technology & Development",
      "keywords": [
        {"word": "software", "weight": 0.15},
        {"word": "development", "weight": 0.12},
        {"word": "code", "weight": 0.10},
        {"word": "programming", "weight": 0.09},
        {"word": "system", "weight": 0.08}
      ],
      "documents": ["mem-1", "mem-3"],
      "coherence": 0.72,
      "diversity": 0.85,
      "metadata": {
        "algorithm": "lda",
        "iterations": 100
      }
    }
  ],
  "success": true
}
```

**Algorithms:**

- **LDA (Latent Dirichlet Allocation):**
  - Probabilistic generative model
  - Uses Gibbs sampling approximation
  - Hyperparameters: Alpha (document-topic density), Beta (topic-word density)
  - Default: Alpha=0.1, Beta=0.01

- **NMF (Non-negative Matrix Factorization):**
  - Matrix factorization approach
  - Uses multiplicative update rules
  - Deterministic results
  - Better for interpretability

**Scoring Metrics:**

- **Coherence:** Measures co-occurrence of topic keywords across documents (0.0-1.0)
- **Diversity:** Measures uniqueness of keywords using prefix overlap (0.0-1.0)

**Topic Assignment:**
After extracting topics, you can assign new documents to topics based on keyword matching.

## General Configuration

**Common Settings:**
```bash
# Enable GPU acceleration (requires CUDA/ROCm)
export NEXS_NLP_USE_GPU=false

# Enable fallback to classical methods
export NEXS_NLP_ENABLE_FALLBACK=true

# Batch size for processing
export NEXS_NLP_BATCH_SIZE=16

# Maximum token length
export NEXS_NLP_MAX_LENGTH=512
```

## Architecture

### ONNX Models

The system supports ONNX (Open Neural Network Exchange) models for transformer-based NLP:

**Required Models:**
- Entity Extraction: `models/bert-base-ner/model.onnx` (BERT/RoBERTa NER)
- Sentiment Analysis: `models/distilbert-sentiment/model.onnx` (DistilBERT)

**Model Provider Interface:**
```go
type ONNXModelProvider interface {
    ExtractEntities(ctx context.Context, text string) ([]EnhancedEntity, error)
    ExtractEntitiesBatch(ctx context.Context, texts []string) ([][]EnhancedEntity, error)
    AnalyzeSentiment(ctx context.Context, text string) (*SentimentResult, error)
    ExtractTopics(ctx context.Context, texts []string, numTopics int) ([]Topic, error)
    IsAvailable() bool
}
```

### ONNX Runtime Integration

The system integrates ONNX Runtime for high-performance transformer model inference with automatic fallback mechanisms.

**Installation:**

1. **Install ONNX Runtime Library:**
```bash
# Ubuntu/Debian
wget https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-linux-x64-1.16.3.tgz
tar -xzf onnxruntime-linux-x64-1.16.3.tgz
sudo cp onnxruntime-linux-x64-1.16.3/lib/* /usr/local/lib/
sudo ldconfig

# macOS
brew install onnxruntime
```

2. **Install Go Binding:**
```bash
go get github.com/yalue/onnxruntime_go
```

3. **Compile with ONNX Support:**
```bash
# Full ONNX support
go build -o nexs-mcp ./cmd/nexs-mcp

# Without ONNX (stub mode)
go build -tags noonnx -o nexs-mcp ./cmd/nexs-mcp
```

**Model Requirements:**

Both models should be exported from PyTorch/TensorFlow to ONNX format with the following specifications:

**Entity Extraction Model (`bert-base-ner.onnx`):**
- Architecture: BERT/RoBERTa fine-tuned for Named Entity Recognition
- Input tensors:
  - `input_ids`: [batch_size, max_length] INT64
  - `attention_mask`: [batch_size, max_length] INT64
  - `token_type_ids`: [batch_size, max_length] INT64 (optional)
- Output tensor:
  - `logits`: [batch_size, max_length, num_labels] FLOAT32
- Labels: BIO tagging scheme (B-PERSON, I-PERSON, B-ORG, I-ORG, B-LOC, I-LOC, etc.)
- Vocabulary: `vocab.txt` in model directory

**Sentiment Analysis Model (`distilbert-sentiment.onnx`):**
- Architecture: DistilBERT fine-tuned for sentiment classification
- Input tensors:
  - `input_ids`: [batch_size, max_length] INT64
  - `attention_mask`: [batch_size, max_length] INT64
- Output tensor:
  - `logits`: [batch_size, num_classes] FLOAT32
- Classes: [NEGATIVE, NEUTRAL, POSITIVE] or [NEGATIVE, POSITIVE]
- Vocabulary: `vocab.txt` in model directory

**Exporting Models to ONNX:**

```python
# Example: Export PyTorch model to ONNX
import torch
from transformers import BertForTokenClassification, BertTokenizer

# Load model
model = BertForTokenClassification.from_pretrained("dslim/bert-base-NER")
tokenizer = BertTokenizer.from_pretrained("dslim/bert-base-NER")

# Save vocabulary
tokenizer.save_vocabulary("./models/bert-base-ner/")

# Create dummy input
dummy_input = {
    "input_ids": torch.randint(0, 28996, (1, 512)),
    "attention_mask": torch.ones(1, 512, dtype=torch.long),
    "token_type_ids": torch.zeros(1, 512, dtype=torch.long)
}

# Export to ONNX
torch.onnx.export(
    model,
    (dummy_input["input_ids"], dummy_input["attention_mask"], dummy_input["token_type_ids"]),
    "./models/bert-base-ner/model.onnx",
    input_names=["input_ids", "attention_mask", "token_type_ids"],
    output_names=["logits"],
    dynamic_axes={
        "input_ids": {0: "batch", 1: "sequence"},
        "attention_mask": {0: "batch", 1: "sequence"},
        "token_type_ids": {0: "batch", 1: "sequence"},
        "logits": {0: "batch", 1: "sequence"}
    },
    opset_version=14
)
```

**Configuration:**
```bash
# Entity model path
export NEXS_NLP_ENTITY_MODEL="models/bert-base-ner"

# Sentiment model path
export NEXS_NLP_SENTIMENT_MODEL="models/distilbert-sentiment"

# Enable GPU (requires CUDA/ROCm)
export NEXS_NLP_USE_GPU=false

# Maximum sequence length
export NEXS_NLP_MAX_LENGTH=512

# Batch size for inference
export NEXS_NLP_BATCH_SIZE=16
```

**Performance Benchmarks:**

| Operation | CPU (ms) | GPU (ms) | Accuracy |
|-----------|----------|----------|----------|
| Entity Extraction | 100-200 | 15-30 | 93%+ |
| Sentiment Analysis | 50-100 | 10-20 | 91%+ |
| Tokenization | 3.5 | - | 100% |
| Softmax | 0.0001 | - | - |
| Argmax | 0.000003 | - | - |

**Tokenization Details:**

The system includes a simplified WordPiece/BPE tokenizer with:
- Automatic vocabulary loading from `vocab.txt`
- Special token support: [CLS], [SEP], [PAD], [UNK], [MASK]
- Fallback vocabulary for testing (101 tokens)
- Maximum sequence length: 512 tokens
- Truncation and padding strategies

**Thread Safety:**

All ONNX sessions are protected by `sync.RWMutex` for safe concurrent access:
```go
type ONNXBERTProvider struct {
    entitySession    *ort.DynamicAdvancedSession
    sentimentSession *ort.DynamicAdvancedSession
    mutex            sync.RWMutex
    // ...
}
```

**Error Handling:**

The provider implements graceful degradation:
1. Check ONNX availability at initialization
2. Log warnings if models not found
3. Fall back to rule-based methods if configured
4. Return clear error messages for debugging

**Monitoring:**

The provider logs:
- Initialization status (available/unavailable)
- Model loading success/failure
- Inference errors
- Fallback activation

**Testing:**

Comprehensive test suite with 15 unit tests:
- Initialization and availability checks
- Entity extraction with various inputs
- Sentiment analysis (positive/negative/neutral)
- Batch processing
- Utility functions (softmax, argmax, tokenize)
- Error handling and unavailability scenarios
- Resource cleanup

**Build Tags:**

Use `noonnx` tag to compile without ONNX Runtime dependency:
```bash
# With ONNX support (default)
go build ./cmd/nexs-mcp

# Without ONNX (stub mode)
go build -tags noonnx ./cmd/nexs-mcp
```

The stub implementation returns `"ONNX support not enabled"` errors for all operations and always reports `IsAvailable() == false`.

### Fallback Mechanisms

All NLP features include robust fallback mechanisms:

1. **Entity Extraction:** Falls back to regex-based extraction with confidence=0.5
2. **Sentiment Analysis:** Falls back to lexicon-based analysis with word lists
3. **Topic Modeling:** Uses pure LDA/NMF (no ONNX required)

**Configuration:**
```bash
# Disable fallback (fail if ONNX unavailable)
export NEXS_NLP_ENABLE_FALLBACK=false
```

## Usage Examples

### Example 1: Extract Entities from Memory

```bash
# Using MCP tool
mcp call extract_entities_advanced '{
  "memory_id": "mem-abc123"
}'
```

### Example 2: Analyze Sentiment Trend

```bash
# Track sentiment over time
mcp call analyze_sentiment_trend '{
  "memory_ids": ["mem-1", "mem-2", "mem-3", "mem-4", "mem-5"]
}'
```

### Example 3: Extract Topics

```bash
# Extract 3 topics using NMF
mcp call extract_topics '{
  "memory_ids": ["mem-1", "mem-2", "mem-3"],
  "num_topics": 3,
  "algorithm": "nmf"
}'
```

### Example 4: Detect Emotional Shifts

```bash
# Find emotional changes with threshold 0.3
mcp call detect_emotional_shifts '{
  "memory_ids": ["mem-1", "mem-2", "mem-3"],
  "threshold": 0.3
}'
```

## Performance Considerations

**Batch Processing:**
- Use batch methods (`extract_entities_batch`, `analyze_memory_batch`) for better performance
- Default batch size: 16 (configurable via `NEXS_NLP_BATCH_SIZE`)

**GPU Acceleration:**
- Enable GPU with `NEXS_NLP_USE_GPU=true`
- Requires CUDA/ROCm support
- Significant speedup for transformer models

**Memory Usage:**
- Topic modeling: Memory scales with vocabulary size
- Sentiment analysis: Minimal memory footprint
- Entity extraction: Memory scales with text length

**Processing Times (CPU, typical):**
- Entity extraction: 100-200ms per document
- Sentiment analysis: 50-100ms per document
- Topic modeling: 1-5s for 100 documents (LDA)

## Troubleshooting

### ONNX Runtime Installation Issues

**Problem:** `error while loading shared libraries: libonnxruntime.so: cannot open shared object file`

**Solution:**
```bash
# Add ONNX Runtime to library path
export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH
sudo ldconfig

# Or install system-wide
sudo cp /path/to/onnxruntime/lib/* /usr/local/lib/
sudo ldconfig
```

**Problem:** Compilation fails with ONNX binding errors

**Solution:**
```bash
# Build without ONNX support
go build -tags noonnx -o nexs-mcp ./cmd/nexs-mcp

# Or install ONNX Runtime properly
go get github.com/yalue/onnxruntime_go
```

### ONNX Models Not Loading

**Problem:** Entity extraction or sentiment analysis fails with "ONNX provider unavailable"

**Solution:**
1. Check model files exist in configured paths:
```bash
ls -la models/bert-base-ner/model.onnx
ls -la models/bert-base-ner/vocab.txt
ls -la models/distilbert-sentiment/model.onnx
ls -la models/distilbert-sentiment/vocab.txt
```

2. Verify ONNX Runtime is installed:
```bash
ldconfig -p | grep onnx
```

3. Enable fallback for graceful degradation:
```bash
export NEXS_NLP_ENABLE_FALLBACK=true
```

4. Check server logs for initialization errors:
```bash
grep "ONNX" ~/.nexs-mcp/logs/nexs-mcp.log
```

**Problem:** Model loads but inference fails

**Solution:**
1. Verify model input/output shapes match specification
2. Check vocabulary file format (one token per line)
3. Ensure model opset version is compatible (opset >= 11)
4. Test model with ONNX Runtime directly:
```python
import onnxruntime as ort
session = ort.InferenceSession("model.onnx")
print(session.get_inputs())  # Verify input names/shapes
print(session.get_outputs())  # Verify output names/shapes
```

### GPU Acceleration Issues

**Problem:** GPU not utilized despite `NEXS_NLP_USE_GPU=true`

**Solution:**
1. Verify CUDA/ROCm installation:
```bash
nvidia-smi  # For NVIDIA GPUs
rocm-smi    # For AMD GPUs
```

2. Install ONNX Runtime with GPU support:
```bash
# Download GPU-enabled version
wget https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-linux-x64-gpu-1.16.3.tgz
```

3. Check GPU availability in logs:
```bash
grep "GPU" ~/.nexs-mcp/logs/nexs-mcp.log
```

### Low Confidence Scores

**Problem:** Entity or sentiment confidence below threshold

**Solution:**
1. Lower threshold: `NEXS_NLP_ENTITY_CONFIDENCE_MIN=0.5`
2. Use fallback methods (generally lower confidence)
3. Improve input text quality

### Topic Modeling Low Coherence

**Problem:** Topics have low coherence scores

**Solution:**
1. Increase topic count: `NEXS_NLP_TOPIC_COUNT=10`
2. Switch algorithms: Try NMF if using LDA
3. Increase document count (minimum 10-20 documents)
4. Filter stopwords (automatically handled)

### Memory Usage High

**Problem:** Topic modeling consumes too much memory

**Solution:**
1. Reduce max word frequency: `NEXS_NLP_TOPIC_MAX_WORD_FREQ=0.6`
2. Increase min word frequency: `NEXS_NLP_TOPIC_MIN_WORD_FREQ=3`
3. Reduce max iterations: `NEXS_NLP_TOPIC_MAX_ITERATIONS=50`

## Future Enhancements

- [ ] Support for multilingual models
- [ ] Custom fine-tuned models
- [ ] Advanced entity disambiguation
- [ ] Cross-document coreference resolution
- [ ] Aspect-based sentiment analysis
- [ ] Dynamic topic modeling (evolution over time)
- [ ] Named entity linking to knowledge bases
- [ ] Emotion intensity estimation
- [ ] Sarcasm and irony detection

## References

- BERT: Devlin et al. (2018) - "BERT: Pre-training of Deep Bidirectional Transformers"
- DistilBERT: Sanh et al. (2019) - "DistilBERT, a distilled version of BERT"
- LDA: Blei et al. (2003) - "Latent Dirichlet Allocation"
- NMF: Lee & Seung (1999) - "Learning the parts of objects by non-negative matrix factorization"
- ONNX: https://onnx.ai/

## See Also

- [MCP Tools API](api/MCP_TOOLS.md)
- [Configuration Reference](VSCODE_SETTINGS_REFERENCE.md)
- [Performance Benchmarks](benchmarks/RESULTS.md)
