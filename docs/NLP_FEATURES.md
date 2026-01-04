# NLP Features (Sprint 18)

## Overview

The NEXS-MCP system includes advanced Natural Language Processing (NLP) capabilities for enhanced memory analysis, entity extraction, sentiment tracking, and topic modeling. These features leverage both state-of-the-art transformer models (via ONNX) and classical algorithms with robust fallback mechanisms.

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
export NEXS_NLP_ENTITY_MODEL="models/bert-base-ner.onnx"

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
- Entity Extraction: `models/bert-base-ner.onnx` (BERT/RoBERTa NER)
- Sentiment Analysis: `models/distilbert-sentiment.onnx` (DistilBERT)

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

### ONNX Models Not Loading

**Problem:** Entity extraction or sentiment analysis fails with "ONNX provider unavailable"

**Solution:**
1. Check model files exist in configured paths
2. Verify ONNX Runtime is installed
3. Enable fallback: `NEXS_NLP_ENABLE_FALLBACK=true`

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
