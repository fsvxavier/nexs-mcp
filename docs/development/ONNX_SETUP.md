# ONNX Setup Guide

This guide explains how to build and run nexs-mcp with ONNX support for local ML-based quality scoring.

## Overview

nexs-mcp supports two build modes:

1. **With ONNX** (local ML inference, requires ONNX Runtime)
2. **Without ONNX** (cloud APIs only: Groq → Gemini → Implicit fallback)

## Quick Start

### Option 1: Build Without ONNX (Recommended for most users)

```bash
# Portable build, no dependencies required
make build-noonnx

# Or with CGO disabled for cross-compilation
CGO_ENABLED=0 go build -tags noonnx -o bin/nexs-mcp ./cmd/nexs-mcp
```

**Pros:**
- ✅ No system dependencies
- ✅ Cross-platform compilation works
- ✅ Smaller binary size
- ✅ Falls back to cloud APIs (Groq, Gemini) or implicit signals

**Cons:**
- ❌ No local ONNX inference (relies on APIs or implicit scoring)

### Option 2: Build With ONNX (For local ML inference)

```bash
# Install ONNX Runtime (cross-platform, requires sudo on Linux/macOS)
make install-onnx

# Build with ONNX support
make build-onnx
```

**Supported Platforms:**
- ✅ Linux (x64) - automatic installation
- ✅ macOS (universal2) - automatic installation
- ⚠️ Windows (x64) - manual installation required

**Pros:**
- ✅ Local ML inference (ms-marco-MiniLM-L-6-v2 model)
- ✅ Low latency (50-100ms CPU, 10-20ms GPU)
- ✅ No API costs
- ✅ Privacy (data stays local)

**Cons:**
- ❌ Requires ONNX Runtime installed
- ❌ Requires CGO (no cross-compilation)
- ❌ Platform-specific builds

## Automatic Installation (Linux/macOS)

The `make install-onnx` command automatically detects your platform and installs ONNX Runtime v1.23.2:

```bash
make install-onnx
```

This works on:
- **Linux**: Uses wget to download and install to /usr/local
- **macOS**: Uses curl to download and install to /usr/local
- **Windows**: Shows manual installation instructions

## Manual ONNX Runtime Installation

### Linux (Ubuntu/Debian)

```bash
cd /tmp
wget https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-x64-1.23.2.tgz
tar -xzf onnxruntime-linux-x64-1.23.2.tgz
sudo cp -r onnxruntime-linux-x64-1.23.2/lib/* /usr/local/lib/
sudo cp -r onnxruntime-linux-x64-1.23.2/include/* /usr/local/include/
sudo ldconfig

# Verify installation
ldconfig -p | grep onnxruntime
```

### macOS

```bash
cd /tmp
curl -LO https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-osx-universal2-1.23.2.tgz
tar -xzf onnxruntime-osx-universal2-1.23.2.tgz
sudo cp -r onnxruntime-osx-universal2-1.23.2/lib/* /usr/local/lib/
sudo cp -r onnxruntime-osx-universal2-1.23.2/include/* /usr/local/include/
sudo update_dyld_shared_cache 2>/dev/null || true

# Verify installation
ls -la /usr/local/lib/libonnxruntime*
```

### Windows

1. Download: https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-win-x64-1.23.2.zip
2. Extract to `C:\onnxruntime`
3. Add to PATH: `C:\onnxruntime\lib`
4. Set environment variables:
   ```cmd
   set CGO_CFLAGS=-IC:\onnxruntime\include
   set CGO_LDFLAGS=-LC:\onnxruntime\lib -lonnxruntime
   ```

**Alternative: Use PowerShell**

```powershell
# Download
$url = "https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-win-x64-1.23.2.zip"
$output = "$env:TEMP\onnxruntime-win-x64-1.23.2.zip"
Invoke-WebRequest -Uri $url -OutFile $output

# Extract
Expand-Archive -Path $output -DestinationPath "C:\onnxruntime" -Force

# Add to PATH (requires admin)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\onnxruntime\lib", [System.EnvironmentVariableTarget]::Machine)

# Verify
Get-ChildItem "C:\onnxruntime\lib\onnxruntime.dll"
```

## Building

### With ONNX Support

```bash
# Using Makefile (sets CGO flags automatically)
make build-onnx

# Or manually
CGO_ENABLED=1 \
  CGO_CFLAGS="-I/usr/local/include" \
  CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime" \
  go build -o bin/nexs-mcp ./cmd/nexs-mcp
```

### Without ONNX Support

```bash
# Using Makefile
make build-noonnx

# Or manually
CGO_ENABLED=0 go build -tags noonnx -o bin/nexs-mcp ./cmd/nexs-mcp
```

### Cross-Platform Builds

```bash
# Builds for all platforms (automatically uses noonnx tag)
make build-all

# Creates release archives
make release
```

## Verifying Your Build

Check which scorer is active:

```bash
# With ONNX
./bin/nexs-mcp --version
# Should show: scorer=onnx (with fallback chain)

# Without ONNX
./bin/nexs-mcp --version
# Should show: scorer=fallback (groq→gemini→implicit)
```

## Fallback Chain

The quality scoring system uses a **multi-tier fallback chain**:

```
1. ONNX (local ML, 50-100ms)
   ↓ if unavailable or fails
2. Groq API (cloud, fast, requires API key)
   ↓ if unavailable or fails
3. Gemini API (cloud, reliable, requires API key)
   ↓ if unavailable or fails
4. Implicit Signals (always available, free, lower accuracy)
```

### Configuring Fallback

In your config file:

```json
{
  "quality": {
    "default_scorer": "onnx",
    "enable_fallback": true,
    "fallback_chain": ["onnx", "groq", "gemini", "implicit"],
    "onnx_model_path": "models/ms-marco-MiniLM-L-6-v2.onnx",
    "groq_api_key": "your-groq-key",
    "gemini_api_key": "your-gemini-key"
  }
}
```

## Troubleshooting

### Error: "ONNX scorer not available"

**Cause:** ONNX Runtime not installed or model file missing.

**Solution:**
```bash
# Install ONNX Runtime
make install-onnx

# Download MS MARCO MiniLM-L-6-v2 model (~23MB)
mkdir -p models
wget -O models/ms-marco-MiniLM-L-6-v2.onnx \
  https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx

# Alternative: use curl
curl -L -o models/ms-marco-MiniLM-L-6-v2.onnx \
  https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx
```

### Error: "undefined reference to onnxruntime"

**Cause:** CGO not finding ONNX Runtime libraries.

**Solution:**
```bash
# Set CGO flags explicitly
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime"
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"

make build-onnx
```

### Error: "build constraints exclude all Go files"

**Cause:** Building with `noonnx` tag but no stub files.

**Solution:** This should not happen. Both `onnx.go` and `onnx_stub.go` exist. Check:
```bash
ls -la internal/quality/onnx*.go
```

### Build works but runtime error: "cannot open shared object file"

**Cause:** ONNX Runtime library not in library path.

**Solution:**
```bash
# Temporary fix
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"

# Permanent fix
sudo ldconfig
```

## Docker Build

The Dockerfile includes ONNX Runtime:

```dockerfile
FROM golang:1.25-bullseye

# Install ONNX Runtime
RUN wget https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-x64-1.23.2.tgz && \
    tar -xzf onnxruntime-linux-x64-1.23.2.tgz && \
    cp -r onnxruntime-linux-x64-1.23.2/lib/* /usr/local/lib/ && \
    cp -r onnxruntime-linux-x64-1.23.2/include/* /usr/local/include/ && \
    ldconfig

WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 go build -o nexs-mcp ./cmd/nexs-mcp

CMD ["./nexs-mcp"]
```

Build with:
```bash
make docker-build
```

## Performance Comparison

| Scorer | Latency | Cost | Accuracy | Availability |
|--------|---------|------|----------|--------------|
| ONNX | 50-100ms (CPU)<br>10-20ms (GPU) | Free | High | Requires setup |
| Groq | 100-200ms | $0.10/1M tokens | Very High | Requires API key |
| Gemini | 200-500ms | $0.075/1M tokens | Very High | Requires API key |
| Implicit | <1ms | Free | Medium | Always available |

## Recommendations

- **Production servers**: Use ONNX build + API fallbacks
- **Development**: Use noonnx build for simplicity
- **CI/CD pipelines**: Use noonnx build for portability
- **Edge devices**: Use ONNX build for privacy/offline
- **Quick testing**: Use noonnx build + implicit scorer

## ONNX Model Information

### MS MARCO MiniLM-L-6-v2 (Recommended)

This is the **official model** used for quality scoring in nexs-mcp:

- **Name**: MS MARCO MiniLM-L-6-v2
- **Size**: ~23MB
- **Purpose**: Passage ranking and semantic similarity
- **Training**: Microsoft MS MARCO dataset (question-answer pairs)
- **Architecture**: MiniLM (distilled BERT)
- **Layers**: 6 transformer layers
- **Parameters**: ~23M

#### Download Model

```bash
# Primary source (HuggingFace)
mkdir -p models
wget -O models/ms-marco-MiniLM-L-6-v2.onnx \
  https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx

# Alternative: curl
curl -L -o models/ms-marco-MiniLM-L-6-v2.onnx \
  https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx
```

#### Model Specifications

- **Input Shape**: `(batch_size, sequence_length)` where sequence_length ≤ 512
- **Input Type**: `int64` token IDs
- **Output Shape**: `(batch_size, 1)`
- **Output Type**: `float32` quality scores (0-1 range)
- **Optimization**: Quantized for faster inference
- **Framework**: ONNX Runtime v1.23.2+

#### Model Sources

- **Primary (Recommended)**: [HuggingFace - Xenova/ms-marco-MiniLM-L-6-v2](https://huggingface.co/Xenova/ms-marco-MiniLM-L-6-v2)
- **Cross-Encoder Version**: [cross-encoder/ms-marco-MiniLM-L6-v2](https://huggingface.co/cross-encoder/ms-marco-MiniLM-L6-v2)
- **ONNX Model Zoo**: [github.com/onnx/models](https://github.com/onnx/models)
- **ONNX Runtime Models**: [onnxruntime.ai/models](https://onnxruntime.ai/models)
- **OpenSearch Models**: [opensearch.org pretrained models](https://docs.opensearch.org/latest/ml-commons-plugin/pretrained-models/)

#### Model Performance

- **CPU Inference**: 50-100ms per document
- **GPU Inference**: 10-20ms per document
- **Batch Inference**: 20-40ms for 10 documents
- **Memory Usage**: ~150MB (model + runtime)
- **Throughput**: ~10-20 docs/sec (single thread)

#### Why MS MARCO MiniLM?

✅ **Optimized for text quality**: Trained specifically for passage ranking  
✅ **Balanced size/performance**: Small enough for edge devices, accurate enough for production  
✅ **Well-supported**: Active maintenance, wide adoption  
✅ **ONNX native**: Optimized ONNX conversion from PyTorch  
✅ **Open source**: MIT license, commercially friendly  

## Further Reading

- [ONNX Runtime Documentation](https://onnxruntime.ai/docs/)
- [MS MARCO Model](https://github.com/microsoft/MSMARCO-Passage-Ranking)
- [Build Tags in Go](https://pkg.go.dev/cmd/go#hdr-Build_constraints)
- [CGO Documentation](https://pkg.go.dev/cmd/cgo)
