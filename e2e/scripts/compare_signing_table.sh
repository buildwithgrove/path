#!/bin/bash

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BENCHMARKS=(
  "BenchmarkShannonKeyOperations"
  "BenchmarkShannonSigningDirect"
  "BenchmarkShannonCompleteSigningPipeline"
)

BENCHTIME="${1:-10s}"
OUTPUT_DIR="${PROJECT_ROOT}/bench_results"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

mkdir -p "$OUTPUT_DIR"

# Function to run benchmark silently
run_benchmark() {
    local build_tags="$1"
    local bench_name="$2"
    local output_file="$3"

    cd "$PROJECT_ROOT"

    local cmd="go test -bench=^$bench_name$ -benchtime=$BENCHTIME -tags=\"bench"
    if [[ -n "$build_tags" ]]; then
        cmd="$cmd,$build_tags"
    fi
    cmd="$cmd\" ./e2e/ -benchmem"

    eval "$cmd" 2>&1 > "$output_file" || echo "FAILED" >> "$output_file"
}

# Function to extract metrics
extract_metrics() {
    local output_file="$1"
    local bench_name="$2"

    local line=$(grep -E "^${bench_name}-[0-9]+\s+" "$output_file" | tail -1 || true)
    if [[ -n "$line" ]]; then
        local ns_per_op=$(echo "$line" | awk '{print $3}')
        echo "$ns_per_op"
    else
        echo "N/A"
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

sanitize_name() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/_/g'
}

# Run benchmarks
for bench_name in "${BENCHMARKS[@]}"; do
    name_sanitized=$(sanitize_name "$bench_name")
    WITHOUT_TAG_OUTPUT="${OUTPUT_DIR}/signing_${name_sanitized}_without_tag_${TIMESTAMP}.txt"
    WITH_TAG_OUTPUT="${OUTPUT_DIR}/signing_${name_sanitized}_with_tag_${TIMESTAMP}.txt"

    run_benchmark "" "$bench_name" "$WITHOUT_TAG_OUTPUT"
    run_benchmark "ethereum_secp256k1" "$bench_name" "$WITH_TAG_OUTPUT"
done

# Print table header
echo ""
echo "Shannon SDK Signing Performance Comparison"
echo "==========================================="
echo ""
printf "%-40s %12s %12s %10s\n" "Benchmark" "Decred" "Ethereum" "Speedup"
printf "%-40s %12s %12s %10s\n" "----------------------------------------" "------------" "------------" "----------"

# Print results for each benchmark
for bench_name in "${BENCHMARKS[@]}"; do
    name_sanitized=$(sanitize_name "$bench_name")
    without_output="${OUTPUT_DIR}/signing_${name_sanitized}_without_tag_${TIMESTAMP}.txt"
    with_output="${OUTPUT_DIR}/signing_${name_sanitized}_with_tag_${TIMESTAMP}.txt"

    ns_without=$(extract_metrics "$without_output" "$bench_name")
    ns_with=$(extract_metrics "$with_output" "$bench_name")

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
echo "Note: BenchmarkShannonKeyOperations only tests key creation from hex,"
echo "      which doesn't involve the actual signing operations where the"
echo "      optimization has the most impact."