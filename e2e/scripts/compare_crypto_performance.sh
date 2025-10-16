#!/bin/bash

# compare_crypto_performance.sh
#
# This script runs the Shannon crypto performance benchmark with and without
# the ethereum_secp256k1 build tag to compare the performance impact.
#
# Usage:
#   ./e2e/scripts/compare_crypto_performance.sh [mode] [benchtime]
#
# Modes:
#   quiet    - Table output only (default)
#   verbose  - Full detailed output with progress
#
# Examples:
#   ./e2e/scripts/compare_crypto_performance.sh              # Quiet mode, 10s
#   ./e2e/scripts/compare_crypto_performance.sh quiet 30s    # Quiet mode, 30s
#   ./e2e/scripts/compare_crypto_performance.sh verbose      # Verbose mode, 10s

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BENCHMARKS=(
  "BenchmarkShannonKeyOperations"
  "BenchmarkShannonSigningDirect"
  "BenchmarkShannonCompleteSigningPipeline"
)

# Parse arguments
MODE="${1:-quiet}"
if [[ "$MODE" =~ ^[0-9]+s$ ]]; then
    # First arg is benchtime, not mode
    BENCHTIME="$MODE"
    MODE="quiet"
else
    BENCHTIME="${2:-10s}"
fi

# Handle help
if [[ "$MODE" == "-h" ]] || [[ "$MODE" == "--help" ]]; then
    echo "Usage: $0 [mode] [benchtime]"
    echo ""
    echo "Modes:"
    echo "  quiet    - Table output only (default)"
    echo "  verbose  - Full detailed output with progress"
    echo ""
    echo "Examples:"
    echo "  $0                    # Quiet mode with 10s benchtime"
    echo "  $0 quiet 30s         # Quiet mode with 30s benchtime"
    echo "  $0 verbose           # Verbose mode with 10s benchtime"
    echo "  $0 60s               # Quiet mode with 60s benchtime"
    echo ""
    echo "Environment variables:"
    echo "  BENCHMEM=true        # Include memory benchmarks (for verbose mode)"
    exit 0
fi

# Validate mode
if [[ "$MODE" != "quiet" ]] && [[ "$MODE" != "verbose" ]]; then
    echo "Error: Invalid mode '$MODE'. Use 'quiet' or 'verbose'."
    exit 1
fi

BENCHMEM="${BENCHMEM:-false}"

# Colors for output (only used in verbose mode)
if [[ "$MODE" == "verbose" ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    BLUE='\033[0;34m'
    YELLOW='\033[1;33m'
    CYAN='\033[0;36m'
    BOLD='\033[1m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    BLUE=''
    YELLOW=''
    CYAN=''
    BOLD=''
    NC=''
fi

# Output files
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
OUTPUT_DIR="${PROJECT_ROOT}/bench_results"
COMPARISON_OUTPUT="${OUTPUT_DIR}/crypto_comparison_${TIMESTAMP}.txt"

# Create output directory
mkdir -p "$OUTPUT_DIR" 2>/dev/null

sanitize_name() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/_/g'
}

# Function to run benchmark
run_benchmark() {
    local description="$1"
    local build_tags="$2"
    local bench_name="$3"
    local output_file="$4"

    if [[ "$MODE" == "verbose" ]]; then
        echo -e "${BLUE}🏃 Running benchmark: ${BOLD}$description${NC}"
        echo -e "${CYAN}   Build tags: $build_tags${NC}"
        echo -e "${CYAN}   Output: $(basename "$output_file")${NC}"
        echo -e "${CYAN}   Benchmark: $bench_name${NC}"
    fi

    cd "$PROJECT_ROOT"

    # Build the benchmark command
    local cmd="go test -bench=^$bench_name$ -benchtime=$BENCHTIME -tags=\"bench"
    if [[ -n "$build_tags" ]]; then
        cmd="$cmd,$build_tags"
    fi
    cmd="$cmd\" ./e2e/"

    if [[ "$BENCHMEM" == "true" ]]; then
        cmd="$cmd -benchmem"
    fi

    if [[ "$MODE" == "verbose" ]]; then
        echo -e "${YELLOW}   Command: $cmd${NC}"
        echo ""
        # Run with output
        if eval "$cmd" 2>&1 | tee "$output_file"; then
            echo -e "${GREEN}✅ Benchmark completed successfully${NC}"
        else
            echo -e "${RED}❌ Benchmark failed (continuing)${NC}"
            echo "FAILED" >> "$output_file" || true
        fi
        echo ""
    else
        # Run silently
        eval "$cmd" > "$output_file" 2>&1 || echo "FAILED" >> "$output_file"
    fi
}

# Function to extract metrics
extract_metrics() {
    local output_file="$1"
    local bench_name="$2"

    local line=$(grep -E "^${bench_name}-[0-9]+\s+" "$output_file" 2>/dev/null | tail -1 || true)
    if [[ -n "$line" ]]; then
        local iterations=$(echo "$line" | awk '{print $2}')
        local ns_per_op=$(echo "$line" | awk '{print $3}')
        echo "${iterations}|${ns_per_op}"
    else
        echo "N/A|N/A"
    fi
}

# Function to format time
format_time() {
    local ns="$1"
    if [[ "$ns" == "N/A" || ! "$ns" =~ ^[0-9]+$ ]]; then
        echo "N/A"
    elif (( ns >= 1000000000 )); then
        echo "$(echo "scale=2; $ns / 1000000000" | bc)s"
    elif (( ns >= 1000000 )); then
        echo "$(echo "scale=2; $ns / 1000000" | bc)ms"
    elif (( ns >= 1000 )); then
        echo "$(echo "scale=1; $ns / 1000" | bc)μs"
    else
        echo "${ns}ns"
    fi
}

# Function to format iterations
format_iterations() {
    local iterations="$1"
    if [[ "$iterations" == "N/A" || ! "$iterations" =~ ^[0-9]+$ ]]; then
        echo "N/A"
    elif (( iterations >= 1000000 )); then
        echo "$(echo "scale=1; $iterations / 1000000" | bc)M"
    elif (( iterations >= 1000 )); then
        echo "$(echo "scale=1; $iterations / 1000" | bc)K"
    else
        echo "$iterations"
    fi
}

# Function to calculate speedup
calc_speedup() {
    local ns1="$1"
    local ns2="$2"
    if [[ "$ns1" != "N/A" && "$ns2" != "N/A" && "$ns1" =~ ^[0-9]+$ && "$ns2" =~ ^[0-9]+$ && "$ns1" -gt 0 ]]; then
        echo "scale=2; $ns1 / $ns2" | bc
    else
        echo "N/A"
    fi
}

# Function to create verbose comparison report
create_verbose_report() {
    echo -e "${BOLD}📊 Creating comparison report...${NC}"

    cat > "$COMPARISON_OUTPUT" << EOF
🔬 Shannon SDK Signing Performance Benchmark
══════════════════════════════════════════════════════════════════

Test Duration: $BENCHTIME per benchmark
Environment: Shannon SDK E2E crypto pipeline

EOF

    for bench_name in "${BENCHMARKS[@]}"; do
        local name_sanitized=$(sanitize_name "$bench_name")
        local without_output="${OUTPUT_DIR}/crypto_${name_sanitized}_without_tag_${TIMESTAMP}.txt"
        local with_output="${OUTPUT_DIR}/crypto_${name_sanitized}_with_tag_${TIMESTAMP}.txt"

        # Extract metrics
        local without_metrics=$(extract_metrics "$without_output" "$bench_name")
        local with_metrics=$(extract_metrics "$with_output" "$bench_name")

        IFS='|' read -r iter1 ns1 <<< "$without_metrics"
        IFS='|' read -r iter2 ns2 <<< "$with_metrics"

        local time1=$(format_time "$ns1")
        local time2=$(format_time "$ns2")
        local fiter1=$(format_iterations "$iter1")
        local fiter2=$(format_iterations "$iter2")

        # Determine winner
        local winner=""
        local performance_change=""
        if [[ "$ns1" != "N/A" && "$ns2" != "N/A" && "$ns1" =~ ^[0-9]+$ && "$ns2" =~ ^[0-9]+$ ]]; then
            if (( ns2 < ns1 )); then
                winner="🥇"
                local improvement=$(echo "scale=1; ($ns1 - $ns2) / $ns1 * 100" | bc -l 2>/dev/null || echo "0")
                performance_change="$(echo $improvement | sed 's/^\./0./')% faster"
            elif (( ns1 < ns2 )); then
                winner=""
                local degradation=$(echo "scale=1; ($ns2 - $ns1) / $ns1 * 100" | bc -l 2>/dev/null || echo "0")
                performance_change="$(echo $degradation | sed 's/^\./0./')% slower"
            else
                winner="🤝"
                performance_change="equivalent performance"
            fi
        fi

        {
            echo ""
            echo "📊 ${bench_name}:"
            echo "┌──────────────────────────┬─────────────┬─────────────┬────────────┐"
            echo "│ Implementation           │ Time/op     │ Iterations  │ Result     │"
            echo "├──────────────────────────┼─────────────┼─────────────┼────────────┤"
            printf "│ %-24s │ %-11s │ %-11s │            │\n" "Decred (default)" "$time1" "$fiter1"
            printf "│ %-24s │ %-11s │ %-11s │ %s          │\n" "Ethereum (libsecp256k1)" "$time2" "$fiter2" "$winner"
            echo "└──────────────────────────┴─────────────┴─────────────┴────────────┘"
            echo ""
            echo "   • Ethereum (libsecp256k1): $performance_change vs Decred (default)"
        } >> "$COMPARISON_OUTPUT"
    done

    cat >> "$COMPARISON_OUTPUT" << EOF

Legend:
  🥇 = Winner    🤝 = Equivalent

Build Tag Configuration:
├─ ethereum_secp256k1: Enables libsecp256k1 C library (fastest, requires CGO)
├─ default (no tag): Uses Decred pure Go implementation (portable)
└─ Impact: Core ECDSA operations (key generation, crypto, verification)
EOF

    echo -e "${GREEN}✅ Comparison report created: $(basename "$COMPARISON_OUTPUT")${NC}"
}

# Main execution
main() {
    if [[ "$MODE" == "verbose" ]]; then
        echo -e "${BOLD}🔐 Shannon Signing Performance Comparison${NC}"
        echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        BENCHMARK_LIST=$(printf "%s, " "${BENCHMARKS[@]}"); BENCHMARK_LIST=${BENCHMARK_LIST%, } || true
        echo -e "${YELLOW}📊 Benchmarks:${NC} $BENCHMARK_LIST"
        echo -e "${YELLOW}⏱️  Duration:${NC} $BENCHTIME"
        echo -e "${YELLOW}📁 Output:${NC} $OUTPUT_DIR"
        echo ""
        echo -e "${CYAN}🔍 Checking prerequisites...${NC}"
    fi

    # Check prerequisites
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        if [[ "$MODE" == "verbose" ]]; then
            echo -e "${RED}❌ Error: Please run this script from the PATH project root${NC}"
        else
            echo "Error: Invalid project root"
        fi
        exit 1
    fi

    if ! grep -q "shannon-sdk" "$PROJECT_ROOT/go.mod" 2>/dev/null; then
        if [[ "$MODE" == "verbose" ]]; then
            echo -e "${RED}❌ Error: shannon-sdk not found in go.mod${NC}"
        else
            echo "Error: shannon-sdk not found"
        fi
        exit 1
    fi

    if [[ "$MODE" == "verbose" ]]; then
        echo -e "${GREEN}✅ Prerequisites check passed${NC}"
        echo ""
    fi

    # Run all benchmarks
    for bench_name in "${BENCHMARKS[@]}"; do
        name_sanitized=$(sanitize_name "$bench_name")
        WITHOUT_TAG_OUTPUT="${OUTPUT_DIR}/crypto_${name_sanitized}_without_tag_${TIMESTAMP}.txt"
        WITH_TAG_OUTPUT="${OUTPUT_DIR}/crypto_${name_sanitized}_with_tag_${TIMESTAMP}.txt"

        run_benchmark "Without ethereum_secp256k1 optimization" "" "$bench_name" "$WITHOUT_TAG_OUTPUT"
        run_benchmark "With ethereum_secp256k1 optimization" "ethereum_secp256k1" "$bench_name" "$WITH_TAG_OUTPUT"
    done

    if [[ "$MODE" == "verbose" ]]; then
        # Create detailed report
        create_verbose_report
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
    else
        # Print clean table only
        echo ""
        echo "Shannon SDK Signing Performance Comparison"
        echo "==========================================="
        echo ""
        printf "%-40s %12s %12s %10s\n" "Benchmark" "Decred" "Ethereum" "Speedup"
        printf "%-40s %12s %12s %10s\n" "----------------------------------------" "------------" "------------" "----------"

        for bench_name in "${BENCHMARKS[@]}"; do
            name_sanitized=$(sanitize_name "$bench_name")
            without_output="${OUTPUT_DIR}/crypto_${name_sanitized}_without_tag_${TIMESTAMP}.txt"
            with_output="${OUTPUT_DIR}/crypto_${name_sanitized}_with_tag_${TIMESTAMP}.txt"

            metrics_without=$(extract_metrics "$without_output" "$bench_name")
            metrics_with=$(extract_metrics "$with_output" "$bench_name")

            IFS='|' read -r _ ns_without <<< "$metrics_without"
            IFS='|' read -r _ ns_with <<< "$metrics_with"

            time_without=$(format_time "$ns_without")
            time_with=$(format_time "$ns_with")
            speedup=$(calc_speedup "$ns_without" "$ns_with")

            if [[ "$speedup" != "N/A" ]]; then
                speedup="${speedup}x"
            fi

            # Shorten benchmark name for display
            display_name="${bench_name#Benchmark}"
            display_name="${display_name#Shannon}"

            printf "%-40s %12s %12s %10s\n" "$display_name" "$time_without" "$time_with" "$speedup"
        done

        echo ""
    fi
}

# Execute main function
main