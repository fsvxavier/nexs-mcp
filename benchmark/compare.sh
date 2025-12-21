#!/bin/bash
#
# NEXS-MCP Benchmark Comparison Script
# Compares NEXS-MCP performance with baseline/competitors
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "NEXS-MCP Performance Benchmark Suite"
echo "========================================="
echo ""

# Configuration
BENCHMARK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_DIR="${BENCHMARK_DIR}/results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_FILE="${RESULTS_DIR}/benchmark_${TIMESTAMP}.txt"
COMPARISON_FILE="${RESULTS_DIR}/comparison_${TIMESTAMP}.md"

# Create results directory
mkdir -p "${RESULTS_DIR}"

echo "Configuration:"
echo "  Results Directory: ${RESULTS_DIR}"
echo "  Results File: ${RESULTS_FILE}"
echo "  Comparison File: ${COMPARISON_FILE}"
echo ""

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to run benchmarks
run_benchmarks() {
    local name=$1
    local cmd=$2
    
    print_status "${YELLOW}" "Running ${name} benchmarks..."
    
    # Run Go benchmarks
    if [[ "$cmd" == "go" ]]; then
        go test -bench=. -benchmem -benchtime=5s ./benchmark/... | tee -a "${RESULTS_FILE}"
    fi
    
    echo "" | tee -a "${RESULTS_FILE}"
}

# Function to extract benchmark results
extract_results() {
    local results_file=$1
    
    echo "## Benchmark Results" > "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
    echo "**Date:** $(date)" >> "${COMPARISON_FILE}"
    echo "**Go Version:** $(go version)" >> "${COMPARISON_FILE}"
    echo "**OS:** $(uname -s) $(uname -m)" >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
    
    # Extract key metrics
    echo "### Performance Metrics" >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
    echo "| Operation | ns/op | B/op | allocs/op |" >> "${COMPARISON_FILE}"
    echo "|-----------|-------|------|-----------|" >> "${COMPARISON_FILE}"
    
    # Parse benchmark results
    grep "^Benchmark" "${results_file}" | while read -r line; do
        benchmark=$(echo "$line" | awk '{print $1}' | sed 's/Benchmark//')
        nsop=$(echo "$line" | awk '{print $3}')
        bop=$(echo "$line" | awk '{print $5}')
        allocsop=$(echo "$line" | awk '{print $7}')
        echo "| ${benchmark} | ${nsop} | ${bop} | ${allocsop} |" >> "${COMPARISON_FILE}"
    done
    
    echo "" >> "${COMPARISON_FILE}"
}

# Function to calculate statistics
calculate_stats() {
    echo "### Summary Statistics" >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
    
    # Average operation time
    avg_time=$(grep "^Benchmark" "${RESULTS_FILE}" | awk '{sum+=$3; count++} END {if(count>0) print sum/count; else print 0}')
    echo "- **Average Operation Time:** ${avg_time} ns/op" >> "${COMPARISON_FILE}"
    
    # Total allocations
    total_allocs=$(grep "^Benchmark" "${RESULTS_FILE}" | awk '{sum+=$7; count++} END {if(count>0) print sum; else print 0}')
    echo "- **Total Allocations:** ${total_allocs}" >> "${COMPARISON_FILE}"
    
    echo "" >> "${COMPARISON_FILE}"
}

# Function to generate comparison charts (ASCII)
generate_charts() {
    echo "### Performance Charts" >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
    echo "\`\`\`" >> "${COMPARISON_FILE}"
    echo "Operation Performance (ns/op)" >> "${COMPARISON_FILE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >> "${COMPARISON_FILE}"
    
    # Generate ASCII chart
    grep "^Benchmark" "${RESULTS_FILE}" | head -n 10 | while read -r line; do
        benchmark=$(echo "$line" | awk '{print $1}' | sed 's/Benchmark//' | cut -c1-20)
        nsop=$(echo "$line" | awk '{print $3}')
        # Normalize to 50 chars max
        bars=$(echo "$nsop" | awk '{printf "%d", $1/1000}')
        if [ "$bars" -gt 50 ]; then
            bars=50
        fi
        bar=$(printf '█%.0s' $(seq 1 $bars))
        printf "%-20s ▕%-50s▏ %s ns/op\n" "$benchmark" "$bar" "$nsop" >> "${COMPARISON_FILE}"
    done
    
    echo "\`\`\`" >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
}

# Function to add recommendations
add_recommendations() {
    echo "### Performance Recommendations" >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
    
    # Analyze results and provide recommendations
    slow_ops=$(grep "^Benchmark" "${RESULTS_FILE}" | awk '$3 > 1000000 {print $1}')
    
    if [ -n "$slow_ops" ]; then
        echo "#### Slow Operations (>1ms)" >> "${COMPARISON_FILE}"
        echo "" >> "${COMPARISON_FILE}"
        echo "$slow_ops" | sed 's/Benchmark//' | while read -r op; do
            echo "- **${op}**: Consider optimization or caching" >> "${COMPARISON_FILE}"
        done
        echo "" >> "${COMPARISON_FILE}"
    fi
    
    # Memory recommendations
    high_mem=$(grep "^Benchmark" "${RESULTS_FILE}" | awk '$5 > 10000 {print $1}')
    
    if [ -n "$high_mem" ]; then
        echo "#### High Memory Operations (>10KB)" >> "${COMPARISON_FILE}"
        echo "" >> "${COMPARISON_FILE}"
        echo "$high_mem" | sed 's/Benchmark//' | while read -r op; do
            echo "- **${op}**: Review memory allocation patterns" >> "${COMPARISON_FILE}"
        done
        echo "" >> "${COMPARISON_FILE}"
    fi
    
    # General recommendations
    echo "#### General Recommendations" >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
    echo "1. **Caching**: Implement caching for frequently accessed elements" >> "${COMPARISON_FILE}"
    echo "2. **Concurrency**: Leverage Go's concurrency for parallel operations" >> "${COMPARISON_FILE}"
    echo "3. **Memory Pooling**: Use sync.Pool for frequently allocated objects" >> "${COMPARISON_FILE}"
    echo "4. **Indexing**: Add indexing for search operations" >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
}

# Main execution
print_status "${GREEN}" "Starting NEXS-MCP benchmarks..."
echo "=========================================" | tee "${RESULTS_FILE}"
echo "NEXS-MCP Performance Benchmark Results" | tee -a "${RESULTS_FILE}"
echo "Date: $(date)" | tee -a "${RESULTS_FILE}"
echo "=========================================" | tee -a "${RESULTS_FILE}"
echo "" | tee -a "${RESULTS_FILE}"

# Run benchmarks
run_benchmarks "NEXS-MCP" "go"

# Extract and analyze results
print_status "${YELLOW}" "Analyzing results..."
extract_results "${RESULTS_FILE}"
calculate_stats
generate_charts
add_recommendations

# Display summary
print_status "${GREEN}" "✓ Benchmarks completed!"
echo ""
echo "Results saved to:"
echo "  - Raw results: ${RESULTS_FILE}"
echo "  - Comparison report: ${COMPARISON_FILE}"
echo ""

# Display quick summary
echo "Quick Summary:"
grep "^Benchmark" "${RESULTS_FILE}" | wc -l | xargs -I {} echo "  - Total benchmarks run: {}"
avg_time=$(grep "^Benchmark" "${RESULTS_FILE}" | awk '{sum+=$3; count++} END {if(count>0) printf "%.2f", sum/count; else print 0}')
echo "  - Average operation time: ${avg_time} ns/op"
echo ""

print_status "${YELLOW}" "View full comparison report:"
echo "  cat ${COMPARISON_FILE}"
echo ""

# Optional: Compare with previous results
if [ -f "${RESULTS_DIR}/baseline.txt" ]; then
    print_status "${YELLOW}" "Comparing with baseline..."
    echo ""
    echo "### Comparison with Baseline" >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
    echo "See ${RESULTS_DIR}/baseline.txt for baseline results." >> "${COMPARISON_FILE}"
    echo "" >> "${COMPARISON_FILE}"
fi

exit 0
