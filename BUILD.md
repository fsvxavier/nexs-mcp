# Build Instructions

## Quick Start

### Native Build (Recommended)
```bash
make build
```
This builds for your current platform and works perfectly.

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
