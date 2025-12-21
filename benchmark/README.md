# NEXS-MCP Benchmark Suite

Comprehensive performance benchmarking for NEXS-MCP operations.

## Overview

This benchmark suite measures the performance of core NEXS-MCP operations:

- **Element CRUD Operations**: Create, Read, Update, Delete
- **Search Operations**: By type, tags, full-text
- **Validation**: Element validation performance
- **Serialization**: JSON marshaling/unmarshaling
- **MCP Tools**: Tool invocation latency
- **Memory Usage**: Allocation patterns
- **Concurrency**: Parallel operation performance
- **Startup Time**: Application initialization

## Running Benchmarks

### Quick Start

Run all benchmarks:

```bash
cd benchmark
./compare.sh
```

### Run Specific Benchmarks

```bash
# All benchmarks
go test -bench=. -benchmem ./benchmark/...

# Specific benchmark
go test -bench=BenchmarkElementCreate -benchmem ./benchmark/...

# With longer benchmark time
go test -bench=. -benchmem -benchtime=10s ./benchmark/...

# With CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof ./benchmark/...

# With memory profiling
go test -bench=. -benchmem -memprofile=mem.prof ./benchmark/...
```

## Benchmark Categories

### Element CRUD Operations

```bash
go test -bench=BenchmarkElement* -benchmem ./benchmark/...
```

Measures:
- Create operations per second
- Read latency
- Update throughput
- Delete performance
- List operations with varying dataset sizes

### Search Performance

```bash
go test -bench=BenchmarkSearch* -benchmem ./benchmark/...
```

Measures:
- Search by type
- Search by tags
- Search by metadata
- Full-text search (if implemented)

### Validation Performance

```bash
go test -bench=BenchmarkValidation -benchmem ./benchmark/...
```

Measures:
- Element validation latency
- Schema validation overhead
- Cross-field validation

### Serialization

```bash
go test -bench=BenchmarkSerialize* -benchmem ./benchmark/...
```

Measures:
- JSON serialization speed
- JSON deserialization speed
- Memory allocations during serialization

### Concurrency

```bash
go test -bench=BenchmarkConcurrent* -benchmem ./benchmark/...
```

Measures:
- Concurrent read performance
- Concurrent write performance
- Lock contention
- Scalability with multiple goroutines

## Understanding Results

### Output Format

```
BenchmarkElementCreate/persona-12    5000    234567 ns/op    1024 B/op    15 allocs/op
```

- `BenchmarkElementCreate/persona-12`: Benchmark name and GOMAXPROCS
- `5000`: Number of iterations run
- `234567 ns/op`: Nanoseconds per operation
- `1024 B/op`: Bytes allocated per operation
- `15 allocs/op`: Number of allocations per operation

### Performance Targets

| Operation | Target | Acceptable |
|-----------|--------|------------|
| Element Create | <100µs | <500µs |
| Element Read | <10µs | <50µs |
| Element Update | <100µs | <500µs |
| Element Delete | <50µs | <200µs |
| List (100 items) | <1ms | <5ms |
| Search by type | <1ms | <5ms |
| Validation | <10µs | <100µs |
| JSON Serialize | <50µs | <200µs |
| JSON Deserialize | <100µs | <500µs |
| MCP Tool Latency | <10ms | <50ms |
| Startup Time | <100ms | <500ms |

## Comparison with DollHouseMCP

To generate a comparison report:

```bash
./compare.sh
```

This will:
1. Run NEXS-MCP benchmarks
2. Generate detailed performance report
3. Create comparison charts
4. Provide optimization recommendations

Results are saved to `results/` directory:
- `benchmark_TIMESTAMP.txt`: Raw benchmark output
- `comparison_TIMESTAMP.md`: Formatted comparison report

## Analyzing Results

### Using pprof

CPU profiling:
```bash
go test -bench=. -cpuprofile=cpu.prof ./benchmark/...
go tool pprof cpu.prof
```

Memory profiling:
```bash
go test -bench=. -memprofile=mem.prof ./benchmark/...
go tool pprof mem.prof
```

### Useful pprof Commands

```
(pprof) top10        # Show top 10 functions
(pprof) list funcName # Show source code for function
(pprof) web          # Generate call graph (requires graphviz)
(pprof) pdf          # Generate PDF report
```

## Continuous Benchmarking

### In CI/CD

Add to GitHub Actions:

```yaml
- name: Run Benchmarks
  run: |
    cd benchmark
    ./compare.sh
    
- name: Upload Results
  uses: actions/upload-artifact@v3
  with:
    name: benchmark-results
    path: benchmark/results/
```

### Regression Detection

Compare with baseline:

```bash
# Save baseline
go test -bench=. ./benchmark/... > results/baseline.txt

# After changes
go test -bench=. ./benchmark/... > results/current.txt

# Compare
benchstat results/baseline.txt results/current.txt
```

## Test Data

Benchmarks use synthetic test data located in `testdata/` directory.

To generate test data:

```bash
# Create test elements
for i in {1..1000}; do
  # Generate test elements
done
```

## Performance Tips

### When to Benchmark

- Before and after performance optimizations
- When adding new features
- When refactoring core code
- During release preparation
- When investigating performance issues

### Interpreting Results

- **ns/op**: Lower is better
- **B/op**: Lower is better (less memory)
- **allocs/op**: Lower is better (less GC pressure)
- Compare against baseline, not absolute numbers
- Focus on relative improvements
- Consider real-world usage patterns

### Common Issues

1. **High Variance**: Run with `-benchtime=10s` for more stable results
2. **Cache Effects**: First run may be slower due to cold cache
3. **CPU Throttling**: Disable CPU frequency scaling for consistent results
4. **Background Processes**: Close unnecessary applications
5. **Dataset Size**: Ensure test data is representative

## Contributing

To add new benchmarks:

1. Add benchmark function to `performance_test.go`
2. Follow naming convention: `BenchmarkOperationName`
3. Use `b.ResetTimer()` after setup
4. Use `b.ReportAllocs()` to track allocations
5. Document expected performance in comments

Example:

```go
func BenchmarkMyOperation(b *testing.B) {
    // Setup
    setup := doSetup()
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        myOperation(setup)
    }
}
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Blog: Profiling Go Programs](https://go.dev/blog/pprof)
- [Benchstat Tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [NEXS-MCP Performance Documentation](../docs/benchmarks/RESULTS.md)

## Support

For questions about benchmarking:
- Open an issue
- Check [Troubleshooting Guide](../docs/user-guide/TROUBLESHOOTING.md)
- See [SUPPORT.md](../SUPPORT.md)

---

**Last Updated:** December 20, 2025
