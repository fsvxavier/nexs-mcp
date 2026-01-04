# Build Instructions

## Quick Start

### Native Build (Recommended)

#### Portable Build (without ONNX Runtime)
```bash
make build
```
This builds a portable binary that works on any system without ONNX dependencies. NLP features will use fallback mechanisms (rule-based entity extraction, lexicon-based sentiment).

#### Full Build (with ONNX Runtime)
```bash
make build-onnx
```
This builds with full NLP capabilities including transformer models (BERT NER, DistilBERT Sentiment). Requires ONNX Runtime library installed.

## Build Tags

### noonnx Tag (Portable Builds)
- **Purpose**: Create portable binaries without ONNX dependencies
- **Usage**: `go build -tags noonnx ./cmd/nexs-mcp`
- **Features**: All features except transformer-based NLP
- **Fallbacks**: Rule-based entity extraction, lexicon-based sentiment, classical topic modeling (LDA/NMF)
- **Size**: Smaller binaries (~20-30 MB)
- **Default**: Used by `make build` and `make build-all`

### Default Build (Full ONNX Support)
- **Purpose**: Full NLP capabilities with transformer models
- **Usage**: `go build ./cmd/nexs-mcp`
- **Features**: All features including BERT NER, DistilBERT Sentiment
- **Requirements**: ONNX Runtime library (libonn xruntime.so/dylib/dll)
- **Size**: Larger binaries (~50-60 MB + model files)
- **Models**: BERT NER (411 MB), DistilBERT Sentiment (516 MB)
- **Performance**: 100-200ms entity extraction, 50-100ms sentiment (CPU)
- **Usage**: `make build-onnx`

## ONNX Runtime Dependencies

### Linux (Ubuntu/Debian)
```bash
# Install ONNX Runtime
wget https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-x64-1.23.2.tgz
tar -xzf onnxruntime-linux-x64-1.23.2.tgz
sudo cp onnxruntime-linux-x64-1.23.2/lib/* /usr/local/lib/
sudo ldconfig

# Verify
ldconfig -p | grep onnxruntime
```

### macOS
```bash
# Install ONNX Runtime
brew install onnxruntime

# Or manual:
wget https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-osx-universal-1.23.2.tgz
tar -xzf onnxruntime-osx-universal-1.23.2.tgz
sudo cp onnxruntime-osx-universal-1.23.2/lib/* /usr/local/lib/
```

### Windows
```powershell
# Download from https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-win-x64-1.23.2.zip
# Extract and add to PATH
```

## NLP Model Setup

### Download BERT NER Model (411 MB)
```bash
mkdir -p models/bert-base-ner
cd models/bert-base-ner
wget https://huggingface.co/protectai/bert-base-NER-onnx/resolve/main/model.onnx
wget https://huggingface.co/protectai/bert-base-NER-onnx/resolve/main/vocab.txt
cd ../..
```

### Download DistilBERT Sentiment Model (516 MB)
```bash
mkdir -p models/distilbert-sentiment
cd models/distilbert-sentiment
wget https://huggingface.co/lxyuan/distilbert-base-multilingual-cased-sentiments-student/resolve/main/model.onnx
wget https://huggingface.co/lxyuan/distilbert-base-multilingual-cased-sentiments-student/resolve/main/vocab.txt
cd ../..
```

See [DOWNLOAD_NLP_MODELS.md](docs/DOWNLOAD_NLP_MODELS.md) for complete instructions.

### Cross-Platform Builds

**⚠️ Known Issue**: Cross-compilation with `make build-all` currently fails due to HNSW library dependencies.

#### Workaround: Build on Target Platform

**For Linux:**
```bash
go build -o nexs-mcp ./cmd/nexs-mcp
```

**For macOS:**
```bash
go build -o nexs-mcp ./cmd/nexs-mcp
```

**For Windows:**
```bash
go build -o nexs-mcp.exe ./cmd/nexs-mcp
```

## Technical Details

The HNSW vector index (added in v1.3.0) uses `github.com/TFMV/hnsw` which depends on `github.com/google/renameio`. This library has platform-specific code that doesn't work well with `CGO_ENABLED=0` cross-compilation.

### Why This Happens

The `renameio` package uses:
- `renameio.TempFile()` which has different implementations per OS
- Cross-compilation with `CGO_ENABLED=0` can't resolve these platform-specific symbols

### Solution Options

1. **Current**: Build natively on each target platform ✅ (Works perfectly)
2. **Future**: Switch to a pure Go HNSW implementation
3. **Alternative**: Use build tags to disable HNSW persistence features in cross-builds

## Testing

All tests work regardless of build method:
```bash
make test        # Run all tests
make test-short  # Run quick tests
make bench       # Run benchmarks
```

## Performance

Native builds achieve full performance:
- 3000x faster vector search (vs linear)
- Sub-50µs queries on 10k vectors
- All HNSW optimizations active

## Support

If you encounter build issues:
1. Use native build: `make build`
2. Check Go version: requires Go 1.21+
3. Ensure dependencies: `go mod tidy`
4. See issues: https://github.com/fsvxavier/nexs-mcp/issues
