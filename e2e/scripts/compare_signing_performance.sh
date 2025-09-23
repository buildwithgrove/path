#!/bin/bash

# compare_signing_performance.sh
#
# This script runs the Shannon signing performance benchmark with and without
# the ethereum_secp256k1 build tag to compare the performance impact.
#
# Usage:
#   ./e2e/scripts/compare_signing_performance.sh
#   ./e2e/scripts/compare_signing_performance.sh 60s  # Custom benchtime

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BENCHMARK_NAME="BenchmarkShannonSigningDirect"

# Default benchmark parameters
BENCHTIME="${1:-5s}"
BENCHMEM="${BENCHMEM:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Output files
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
OUTPUT_DIR="${PROJECT_ROOT}/bench_results"
WITHOUT_TAG_OUTPUT="${OUTPUT_DIR}/signing_bench_without_tag_${TIMESTAMP}.txt"
WITH_TAG_OUTPUT="${OUTPUT_DIR}/signing_bench_with_tag_${TIMESTAMP}.txt"
COMPARISON_OUTPUT="${OUTPUT_DIR}/signing_comparison_${TIMESTAMP}.txt"

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo -e "${BOLD}🔐 Shannon Signing Performance Comparison${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}📊 Benchmark:${NC} $BENCHMARK_NAME"
echo -e "${YELLOW}⏱️  Duration:${NC} $BENCHTIME"
echo -e "${YELLOW}📁 Output:${NC} $OUTPUT_DIR"
echo ""

# Function to run benchmark
run_benchmark() {
    local description="$1"
    local build_tags="$2"
    local output_file="$3"

    echo -e "${BLUE}🏃 Running benchmark: ${BOLD}$description${NC}"
    echo -e "${CYAN}   Build tags: $build_tags${NC}"
    echo -e "${CYAN}   Output: $(basename "$output_file")${NC}"

    cd "$PROJECT_ROOT"

    # Build the benchmark command - use e2e directory benchmark
    local cmd="go test -bench=$BENCHMARK_NAME -benchtime=$BENCHTIME -tags=\"bench"
    if [[ -n "$build_tags" ]]; then
        cmd="$cmd,$build_tags"
    fi
    cmd="$cmd\" ./e2e/"

    if [[ "$BENCHMEM" == "true" ]]; then
        cmd="$cmd -benchmem"
    fi

    echo -e "${YELLOW}   Command: $cmd${NC}"
    echo ""

    # Run the benchmark and capture output
    if eval "$cmd" 2>&1 | tee "$output_file"; then
        echo -e "${GREEN}✅ Benchmark completed successfully${NC}"
    else
        echo -e "${RED}❌ Benchmark failed${NC}"
        return 1
    fi
    echo ""
}

# Function to extract metrics from benchmark output
extract_metrics() {
    local output_file="$1"
    local description="$2"

    # Extract Go benchmark results (iterations and time per operation)
    # Handle case where benchmark line might be corrupted by log output
    # Look for the pattern: numbers + "ns/op"
    local iterations=$(grep -E "^\s*[0-9]+\s+[0-9]+\s+ns/op" "$output_file" | awk '{print $1}' | tail -1 || echo "N/A")
    local ns_per_op=$(grep -E "^\s*[0-9]+\s+[0-9]+\s+ns/op" "$output_file" | awk '{print $2}' | tail -1 || echo "N/A")

    # Convert ns to more readable format
    local time_per_op="N/A"
    if [[ "$ns_per_op" != "N/A" && "$ns_per_op" =~ ^[0-9]+$ ]]; then
        if (( ns_per_op >= 1000000 )); then
            time_per_op="$(echo "scale=1; $ns_per_op / 1000000" | bc)ms"
        elif (( ns_per_op >= 1000 )); then
            time_per_op="$(echo "scale=1; $ns_per_op / 1000" | bc)μs"
        else
            time_per_op="${ns_per_op}ns"
        fi
    fi

    # Format iterations with appropriate suffix
    local formatted_iterations="N/A"
    if [[ "$iterations" != "N/A" && "$iterations" =~ ^[0-9]+$ ]]; then
        if (( iterations >= 1000000 )); then
            formatted_iterations="$(echo "scale=1; $iterations / 1000000" | bc)M"
        elif (( iterations >= 1000 )); then
            formatted_iterations="$(echo "scale=1; $iterations / 1000" | bc)K"
        else
            formatted_iterations="$iterations"
        fi
    fi

    echo "$description|$time_per_op|$formatted_iterations|$ns_per_op"
}

# Function to create comparison report
create_comparison_report() {
    echo -e "${BOLD}📊 Creating comparison report...${NC}"

    # Extract metrics from both runs
    local without_metrics=$(extract_metrics "$WITHOUT_TAG_OUTPUT" "Decred (default)")
    local with_metrics=$(extract_metrics "$WITH_TAG_OUTPUT" "Ethereum (libsecp256k1)")

    # Parse metrics
    IFS='|' read -r desc1 time1 iter1 ns1 <<< "$without_metrics"
    IFS='|' read -r desc2 time2 iter2 ns2 <<< "$with_metrics"

    # Determine winner
    local winner=""
    local performance_change=""
    if [[ "$ns1" != "N/A" && "$ns2" != "N/A" && "$ns1" =~ ^[0-9]+$ && "$ns2" =~ ^[0-9]+$ ]]; then
        if (( ns2 < ns1 )); then
            winner="🥇"
            local improvement=$(echo "scale=1; ($ns1 - $ns2) / $ns1 * 100" | bc -l 2>/dev/null || echo "0")
            performance_change="$(echo $improvement | sed 's/^\./0./')% faster"
        elif (( ns1 < ns2 )); then
            winner="🥈"
            local degradation=$(echo "scale=1; ($ns2 - $ns1) / $ns1 * 100" | bc -l 2>/dev/null || echo "0")
            performance_change="$(echo $degradation | sed 's/^\./0./')% slower"
        else
            winner="🤝"
            performance_change="equivalent performance"
        fi
    fi

    cat > "$COMPARISON_OUTPUT" << EOF
🔬 Shannon SDK Signing Performance Benchmark
══════════════════════════════════════════════════════════════════

📊 SIGNING PERFORMANCE COMPARISON:
┌──────────────────────────┬─────────────┬─────────────┬────────────┐
│ Implementation           │ Time/op     │ Iterations  │ Result     │
├──────────────────────────┼─────────────┼─────────────┼────────────┤
│ $(printf "%-24s" "$desc1") │ $(printf "%-11s" "$time1") │ $(printf "%-11s" "$iter1") │            │
│ $(printf "%-24s" "$desc2") │ $(printf "%-11s" "$time2") │ $(printf "%-11s" "$iter2") │ $winner          │
└──────────────────────────┴─────────────┴─────────────┴────────────┘

💡 Performance Summary:
   🥇 = Winner    🥈 = Second place    🤝 = Equivalent

   • $desc2: $performance_change vs $desc1
   • Test Duration: $BENCHTIME per implementation
   • Environment: Shannon SDK E2E signing pipeline

Build Tag Configuration:
├─ ethereum_secp256k1: Enables libsecp256k1 C library (fastest, requires CGO)
├─ default (no tag): Uses Decred pure Go implementation (portable)
└─ Impact: Core ECDSA operations (key generation, signing, verification)
EOF

    echo -e "${GREEN}✅ Comparison report created: $(basename "$COMPARISON_OUTPUT")${NC}"
}

# Main execution
main() {
    echo -e "${CYAN}🔍 Checking prerequisites...${NC}"

    # Check if we're in the correct directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        echo -e "${RED}❌ Error: Please run this script from the PATH project root${NC}"
        exit 1
    fi

    # Check if shannon-sdk is configured
    if ! grep -q "shannon-sdk" "$PROJECT_ROOT/go.mod"; then
        echo -e "${RED}❌ Error: shannon-sdk not found in go.mod${NC}"
        exit 1
    fi

    echo -e "${GREEN}✅ Prerequisites check passed${NC}"
    echo ""

    # Run benchmark without ethereum_secp256k1 tag
    run_benchmark "Without ethereum_secp256k1 optimization" "" "$WITHOUT_TAG_OUTPUT"

    # Run benchmark with ethereum_secp256k1 tag
    run_benchmark "With ethereum_secp256k1 optimization" "ethereum_secp256k1" "$WITH_TAG_OUTPUT"

    # Create comparison report
    create_comparison_report

    echo -e "${BOLD}🎉 Benchmark comparison completed!${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}📊 Results available in: $OUTPUT_DIR${NC}"
    echo -e "${GREEN}📈 Comparison report: $(basename "$COMPARISON_OUTPUT")${NC}"
    echo ""
    echo -e "${YELLOW}📋 Benchmark Results:${NC}"
    echo ""
    cat "$COMPARISON_OUTPUT"
    echo ""
    echo -e "${CYAN}💾 Full report saved to: $COMPARISON_OUTPUT${NC}"
}

# Handle script arguments
case "${1:-}" in
    -h|--help)
        echo "Usage: $0 [benchtime]"
        echo ""
        echo "Examples:"
        echo "  $0                    # Run with default 30s benchtime"
        echo "  $0 60s               # Run with 60s benchtime"
        echo ""
        echo "Environment variables:"
        echo "  BENCHMEM=true        # Include memory benchmarks"
        exit 0
        ;;
esac

# Execute main function
main