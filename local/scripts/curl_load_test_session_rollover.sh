#!/bin/bash

# Monitor XRPL EVM testnet block number

show_help() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS]

Monitor XRPL EVM testnet block numbers via JSON-RPC eth_blockNumber calls.

OPTIONS:
    --local     Use localhost endpoint (http://localhost:3070/v1) instead of remote
    --help      Show this help message and exit

EXAMPLES:
    $(basename "$0")
        Monitor using remote Grove endpoint

    $(basename "$0") --local
        Monitor using local PATH gateway

EOF
    exit 0
}

# Parse command line arguments
USE_LOCAL=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --local)
            USE_LOCAL=true
            shift
            ;;
        --help)
            show_help
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Print configuration mode at the beginning
if [ "$USE_LOCAL" = true ]; then
    echo "🔧 MODE: LOCAL"
else
    echo "🌐 MODE: REMOTE"
fi
echo ""

# Configure endpoint based on mode
if [ "$USE_LOCAL" = true ]; then
    ENDPOINT="http://localhost:3070/v1"
    echo "Starting XRPL EVM testnet block number monitor (LOCAL)..."
else
    ENDPOINT="https://xrpl-evm-testnet.rpc.grove.city/v1/6c5de5ff"
    echo "Starting XRPL EVM testnet block number monitor (REMOTE)..."
fi

echo "Press Ctrl+C to stop"
echo ""

# Initialize variables for delta calculation
last_request_time=0

while true; do
    # Get current timestamp for display and delta calculation
    current_time=$(date +%s.%N)
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    # Calculate delta from last request
    if [ "$last_request_time" != "0" ]; then
        delta=$(echo "$current_time - $last_request_time" | bc -l)
        delta_formatted=$(printf "%.2f" "$delta")

        # Color code based on delta (assuming 0.5s target interval)
        # Green: <= 1s (quick), Yellow: 1-3s (moderate), Red: > 3s (slow)
        if (( $(echo "$delta <= 1" | bc -l) )); then
            delta_color="\033[32m"  # Green
            delta_status="⚡"
        elif (( $(echo "$delta <= 3" | bc -l) )); then
            delta_color="\033[33m"  # Yellow
            delta_status="⏱️"
        else
            delta_color="\033[31m"  # Red
            delta_status="🐌"
        fi
        reset_color="\033[0m"
        delta_display=" ${delta_color}[Δ${delta_formatted}s ${delta_status}]${reset_color}"
    else
        delta_display=""
    fi

    # Execute RPC call to get current block number
    if [ "$USE_LOCAL" = true ]; then
        response=$(curl -s "$ENDPOINT" \
            -H "Target-Service-Id: xrplevm" \
            -H "Authorization: test_api_key" \
            -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }' 2>&1)
    else
        response=$(curl -s "$ENDPOINT" \
            -X POST \
            -H "Content-Type: application/json" \
            --data '{"method":"eth_blockNumber","params":[],"id":1,"jsonrpc":"2.0"}' 2>&1)
    fi

    # Check if curl command was successful
    curl_exit_code=$?

    if [ $curl_exit_code -ne 0 ]; then
        echo -e "❌ [$timestamp] ERROR: curl failed with exit code $curl_exit_code: $response$delta_display"
    else
        # Check if response contains an error
        if echo "$response" | grep -q '"error"'; then
            echo -e "❌ [$timestamp] API ERROR: $response$delta_display"
        else
            # Extract block number from response (remove 0x prefix and convert from hex)
            block_hex=$(echo "$response" | grep -o '"result":"0x[^"]*"' | cut -d'"' -f4)

            if [ -z "$block_hex" ]; then
                echo -e "❌ [$timestamp] PARSE ERROR: Could not extract block number from response: $response$delta_display"
            else
                # Convert hex to decimal, handle potential conversion errors
                if block_decimal=$((16#${block_hex#0x})) 2>/dev/null; then
                    echo -e "✅ [$timestamp] Block: $block_decimal (hex: $block_hex)$delta_display"
                else
                    echo -e "❌ [$timestamp] CONVERSION ERROR: Could not convert hex $block_hex to decimal$delta_display"
                fi
            fi
        fi
    fi

    # Update last request time for next iteration
    last_request_time="$current_time"

    # Wait 0.5 seconds
    sleep 0.5
done