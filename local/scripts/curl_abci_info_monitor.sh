#!/bin/bash

# Monitor Cosmos RPC abci_info endpoint at regular intervals

show_help() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS]

Monitor Cosmos RPC abci_info endpoint and display results with timing information.

OPTIONS:
    --block          Track only block height changes from abci_info response
    --node <URL>     Override default RPC endpoint (default: https://shannon-grove-rpc.mainnet.poktroll.com)
    --help           Show this help message and exit

EXAMPLES:
    $(basename "$0")
        Monitor full abci_info response using default endpoint

    $(basename "$0") --block
        Monitor only block height changes

    $(basename "$0") --node http://localhost:26657
        Use custom RPC endpoint

    $(basename "$0") --block --node http://localhost:26657
        Monitor block height using custom endpoint

EOF
    exit 0
}

# Parse command line arguments
MONITOR_BLOCK=false
RPC_ENDPOINT="https://shannon-grove-rpc.mainnet.poktroll.com"

while [[ $# -gt 0 ]]; do
    case $1 in
        --block)
            MONITOR_BLOCK=true
            shift
            ;;
        --node)
            if [ -z "$2" ] || [[ "$2" == --* ]]; then
                echo "Error: --node requires a URL argument"
                exit 1
            fi
            RPC_ENDPOINT="$2"
            shift 2
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
if [ "$MONITOR_BLOCK" = true ]; then
    echo "📊 MODE: BLOCK HEIGHT MONITORING"
else
    echo "🔍 MODE: ABCI INFO MONITORING"
fi
echo "🌐 ENDPOINT: $RPC_ENDPOINT"
echo ""

echo "Starting Cosmos RPC abci_info monitor..."
echo "Press Ctrl+C to stop"
echo ""

# Initialize variables for delta calculation
last_request_time=0
last_block_height=""

while true; do
    # Get current timestamp for display and delta calculation
    current_time=$(date +%s.%N)
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    # Calculate delta from last request
    if [ "$last_request_time" != "0" ]; then
        delta=$(echo "$current_time - $last_request_time" | bc -l)
        delta_formatted=$(printf "%.2f" "$delta")

        # Color code based on delta (assuming 0.5s target interval)
        # Green: <= 0.7s (quick), Yellow: 0.7-1.5s (moderate), Red: > 1.5s (slow)
        if (( $(echo "$delta <= 0.7" | bc -l) )); then
            delta_color="\033[32m"  # Green
            delta_status="⚡"
        elif (( $(echo "$delta <= 1.5" | bc -l) )); then
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

    # Make the curl request to abci_info endpoint
    response=$(curl -s "$RPC_ENDPOINT/abci_info" 2>&1)

    # Check if curl command was successful
    curl_exit_code=$?

    if [ $curl_exit_code -ne 0 ]; then
        echo -e "❌ [$timestamp] ERROR: curl failed with exit code $curl_exit_code: $response$delta_display"
    else
        # Check if response contains an error
        if echo "$response" | grep -q '"error"'; then
            echo -e "❌ [$timestamp] API ERROR: $response$delta_display"
        else
            if [ "$MONITOR_BLOCK" = true ]; then
                # Extract block height from abci_info response
                block_height=$(echo "$response" | jq -r '.result.response.last_block_height // empty' 2>/dev/null)

                if [ -z "$block_height" ] || [ "$block_height" = "null" ]; then
                    echo -e "❌ [$timestamp] PARSE ERROR: Could not extract block height from response$delta_display"
                else
                    # Check if block height changed
                    if [ "$last_block_height" != "" ] && [ "$block_height" != "$last_block_height" ]; then
                        height_change="📈"
                        height_diff=$((block_height - last_block_height))
                        height_info=" (↑$height_diff)"
                    else
                        height_change="📊"
                        height_info=""
                    fi

                    echo -e "✅ [$timestamp] $height_change Block Height: $block_height$height_info$delta_display"
                    last_block_height="$block_height"
                fi
            else
                # Display full abci_info response details
                app_name=$(echo "$response" | jq -r '.result.response.data // "N/A"' 2>/dev/null)
                version=$(echo "$response" | jq -r '.result.response.version // "N/A"' 2>/dev/null)
                block_height=$(echo "$response" | jq -r '.result.response.last_block_height // "N/A"' 2>/dev/null)

                if [ "$app_name" = "null" ] || [ "$app_name" = "" ]; then
                    app_name="N/A"
                fi
                if [ "$version" = "null" ] || [ "$version" = "" ]; then
                    version="N/A"
                fi
                if [ "$block_height" = "null" ] || [ "$block_height" = "" ]; then
                    block_height="N/A"
                fi

                echo -e "✅ [$timestamp] ABCI Info - App: $app_name, Version: $version, Height: $block_height$delta_display"
            fi
        fi
    fi

    # Update last request time for next iteration
    last_request_time="$current_time"

    # Wait 0.5 seconds
    sleep 0.5
done
