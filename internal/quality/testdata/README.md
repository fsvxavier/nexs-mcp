# Test Models for ONNX Quality Scorer

## Overview

The ONNX quality scorer tests require an ONNX model file to run. The tests will automatically try to:

1. Use the production model at `../../models/ms-marco-MiniLM-L-6-v2.onnx` if available
2. Download the MS MARCO MiniLM L-6 v2 model (~23MB) from HuggingFace
3. Skip tests if no model is available and downloads are disabled

## Recommended Model: MS MARCO MiniLM-L-6-v2

This is the **recommended model** for text quality scoring:

- **Size**: ~23MB
- **Purpose**: Semantic similarity and passage ranking
- **Performance**: 50-100ms CPU inference, 10-20ms GPU
- **Input**: Text sequences (max 512 tokens)
- **Output**: Quality score (0-1 range)

## Downloading Models Manually

### Option 1: Use Production Model (Recommended)

Download the MS MARCO MiniLM-L-6-v2 model from HuggingFace:

```bash
mkdir -p ../../models
cd ../../models
wget https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx \
  -O ms-marco-MiniLM-L-6-v2.onnx
```

Or use curl:

```bash
curl -L -o ../../models/ms-marco-MiniLM-L-6-v2.onnx \
  https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx
```

### Option 2: Use Test Model in testdata

Download directly to test directory:

```bash
mkdir -p testdata/models
cd testdata/models
wget https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx \
  -O ms-marco-MiniLM-L-6-v2.onnx
```

### Option 3: Automatic Download

The tests will automatically attempt to download the model when you run them:

```bash
go test ./internal/quality/... -run TestONNX -v
```

The model will be cached in `testdata/models/` for future test runs.

## Disabling Automatic Downloads

Set environment variables to prevent automatic downloads:

```bash
export SKIP_DOWNLOAD=1
go test ./internal/quality/... -run TestONNX -v
```

Or in CI:

```bash
CI=true go test ./internal/quality/... -run TestONNX -v
```

## Model Sources

### Primary Sources (Recommended)

- **HuggingFace MS MARCO**: https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2
  - Direct ONNX: https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2/blob/main/onnx/model.onnx
- **Cross-Encoder Version**: https://huggingface.co/cross-encoder/ms-marco-MiniLM-L6-v2

### Alternative Sources

- **ONNX Model Zoo**: https://github.com/onnx/models
- **ONNX Runtime Models**: https://onnxruntime.ai/models
- **OpenSearch Models**: https://docs.opensearch.org/latest/ml-commons-plugin/pretrained-models/

## Model Specifications

The MS MARCO MiniLM-L-6-v2 model expects:

- **Input Shape**: `(batch_size, sequence_length)` where sequence_length ≤ 512
- **Input Type**: `int64` token IDs
- **Output Shape**: `(batch_size, 1)`
- **Output Type**: `float32` quality scores
- **Framework**: ONNX Runtime v1.16.3+
- **Optimization**: Quantized for faster inference

## Notes

- The MS MARCO MiniLM model is specifically trained for passage ranking and quality assessment
- Model is optimized for semantic similarity scoring
- Works best with English text content
- Supports batch inference for better throughput
- For production use, ensure ONNX Runtime is properly installed (see docs/development/ONNX_SETUP.md)
- Model cache improves test execution time (tests skip download after first run)
