#!/bin/bash
# Script to manage PATH Localnet in Docker
# Usage: ./localnet.sh [up|down]
#        ./localnet.sh up --use-local-helm

set -e

# Configuration
DOCKER_IMAGE="ghcr.io/buildwithgrove/path-localnet-env:latest"
USE_LOCAL_HELM=false

# Define spinner animation frames
FRAMES=('⠋' '⠙' '⠚' '⠞' '⠖' '⠦' '⠴' '⠲' '⠳' '⠓')
FRAMES_COUNT=${#FRAMES[@]}

# Function to prompt for local helm charts
prompt_for_local_helm_charts() {
    # Return if LOCAL_HELM_CHARTS_PATH is already set
    if [ -n "${LOCAL_HELM_CHARTS_PATH}" ]; then
        return 0
    fi

    # Default path for local helm charts
    DEFAULT_HELM_PATH="../helm-charts"
    
    read -p "  📂 Enter the path to the local helm charts repository [press enter for default: ${DEFAULT_HELM_PATH}]: " input_path
    input_path=${input_path:-$DEFAULT_HELM_PATH}
    
    # Expand path if it starts with ~
    if [[ "$input_path" == ~* ]]; then
        input_path="${HOME}${input_path:1}"
    fi
    
    # Convert relative path to absolute path
    if [[ ! "$input_path" = /* ]]; then
        input_path="$(pwd)/$input_path"
    fi
    
    # Normalize the path (resolve ../ and ./ segments)
    input_path=$(cd "$(dirname "$input_path")" 2>/dev/null && pwd -P)/$(basename "$input_path")
    
    # Check if the path exists
    if [ ! -d "$input_path" ]; then
        echo "❌ Directory does not exist: $input_path"
        echo "Would you like to create it? [y/N]"
        read -p "> " create_dir
        if [[ "${create_dir,,}" == "y" || "${create_dir,,}" == "yes" ]]; then
            mkdir -p "$input_path"
            echo "✅ Created directory: $input_path"
        else
            echo "❌ Using remote helm charts instead"
            return 1
        fi
    fi
    
    # Set the environment variable with the normalized path
    export LOCAL_HELM_CHARTS_PATH="$input_path"
    echo "  ✅ Using local helm charts from: $LOCAL_HELM_CHARTS_PATH"
    return 0
}

# Function to run a spinner with a timeout
run_spinner() {
    local timeout=$1
    local check_url=$2
    local message=$3
    local success_message=$4
    
    echo -e "\n⏳ $message"
    
    local spinner_counter=0
    local check_interval=5  # Check connection every 5 spinner iterations (approx 1 second)
    
    local start_time=$(date +%s)
    local end_time=$((start_time + timeout))
    
    while [ $(date +%s) -lt $end_time ]; do
        # Update spinner
        local frame_idx=$((spinner_counter % FRAMES_COUNT))
        local frame="${FRAMES[$frame_idx]}"
        spinner_counter=$((spinner_counter + 1))
        
        # Display spinner
        printf "\r  %s Loading... " "$frame"
        
        # Check endpoint every ~1 second
        if [ $((spinner_counter % check_interval)) -eq 0 ]; then
            local status=$(curl -s -o /dev/null -w "%{http_code}" "$check_url" 2>/dev/null || echo "000")
            if [[ "$status" == "200" ]]; then
                printf "\r  ✨ $success_message                               \n"
                return 0
            fi
        fi
        
        sleep 0.2
    done
    
    echo ""
    echo "❌ Timed out waiting for $message"
    return 1
}

# Function to show a simple spinner for a process
show_spinner() {
    local message=$1
    local cmd=$2
    local success_message=$3
    
    echo -e "\n⏳ $message"
    
    local spinner_counter=0
    
    # Start the command in the background
    eval "$cmd" &
    local cmd_pid=$!
    
    # Display spinner while the command is running
    while kill -0 $cmd_pid 2>/dev/null; do
        local frame_idx=$((spinner_counter % FRAMES_COUNT))
        local frame="${FRAMES[$frame_idx]}"
        spinner_counter=$((spinner_counter + 1))
        
        printf "\r  %s Processing... " "$frame"
        sleep 0.2
    done
    
    # Wait for the command to finish and get its exit status
    wait $cmd_pid
    local exit_status=$?
    
    if [ $exit_status -eq 0 ]; then
        printf "\r  ✨ $success_message                               \n"
        return 0
    else
        printf "\r  ❌ Failed to $message                             \n"
        return 1
    fi
}

# Function to start up PATH Localnet
start_localnet() {
    # Check if container already exists
    if docker ps -a --format '{{.Names}}' | grep -q "^path-localnet$"; then
        # Check if container is running
        if docker ps --format '{{.Names}}' | grep -q "^path-localnet$"; then
            echo "❌ Error: path-localnet is already running."
            echo "To stop it, run: ./local/scripts/localnet.sh down"
            exit 1
        else
            echo "🧹 Removing stopped path-localnet container..."
            docker rm path-localnet > /dev/null 2>&1 || true
        fi
    fi

    # Check if we should use local helm charts
    if [ "$USE_LOCAL_HELM" = true ]; then
        echo "🔍 Running PATH Localnet with local helm charts ..."
        if ! prompt_for_local_helm_charts; then
            echo "⚠️ Failed to set up local helm charts, reverting to remote charts"
            unset LOCAL_HELM_CHARTS_PATH
        fi
    else
        echo "  📡 Starting PATH Localnet with remote helm charts"
    fi

    # Start the container
    echo "📡 Starting PATH Localnet ..."
    
    # Set up docker run command with base parameters
    DOCKER_CMD="docker run \
        --name path-localnet \
        -v \"$(pwd)\":/app \
        -p 10350:10350 \
        -p 3070:3070 \
        --privileged"
    
    # Add helm charts volume mount if LOCAL_HELM_CHARTS_PATH is set
    if [ -n "${LOCAL_HELM_CHARTS_PATH}" ]; then
        if [ ! -d "${LOCAL_HELM_CHARTS_PATH}" ]; then
            echo "❌ Error: LOCAL_HELM_CHARTS_PATH directory does not exist: ${LOCAL_HELM_CHARTS_PATH}"
            exit 1
        fi
        echo "  📦 Mounting local helm charts from ${LOCAL_HELM_CHARTS_PATH}"
        DOCKER_CMD="${DOCKER_CMD} \
        -v \"${LOCAL_HELM_CHARTS_PATH}\":/helm-charts \
        -e LOCAL_HELM_CHARTS_PATH=/helm-charts"
    fi
    
    # Complete the docker command
    DOCKER_CMD="${DOCKER_CMD} \
        -d \
        ${DOCKER_IMAGE}"
    
    # Run the container
    if ! eval ${DOCKER_CMD}; then
        echo "❌ Failed to start Docker container. Check if ports 10350 and 3070 are available."
        exit 1
    fi

    # First, wait for Tilt UI to become available
    if run_spinner 180 "http://localhost:10350" "Waiting for Tilt UI to become available..." "Tilt UI is now available!"; then
        echo "  ✅ Access Tilt UI at: http://localhost:10350"
        
        # Next, wait for the healthz endpoint
        if run_spinner 300 "http://localhost:3070/healthz" "Waiting for PATH API to be ready..." "PATH API is ready!"; then
            echo -e "\n🌿 PATH Localnet started successfully."
            echo "  🚀 Send relay requests to: http://localhost:3070/v1"
            exit 0
        else
            echo "  Check logs with: docker logs path-localnet"
            exit 1
        fi
    else
        echo "  Check logs with: docker logs path-localnet"
        exit 1
    fi
}

# Function to stop PATH Localnet
stop_localnet() {
    # Check if Docker is installed
    if ! command -v docker >/dev/null 2>&1; then
        echo "❌ Docker is not installed. Make sure you review README.md before continuing"
        exit 1
    fi

    # Stop and remove the container with spinner
    if show_spinner "Stopping path-localnet container" "docker stop path-localnet > /dev/null 2>&1 || true" "Container stopped"; then
        if show_spinner "Removing path-localnet container" "docker rm path-localnet > /dev/null 2>&1 || true" "Container removed"; then
            echo -e "\n✅ PATH Localnet has been successfully stopped and removed."
            exit 0
        fi
    fi
    
    echo "⚠️ There might have been issues stopping PATH Localnet."
    exit 1
}

# Parse command-line arguments
COMMAND=${1:-up}
shift || true

# Parse additional arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --use-local-helm)
            USE_LOCAL_HELM=true
            ;;
        *)
            echo "❌ Unknown option: $1"
            echo "Usage: $0 [up|down] [--use-local-helm]"
            exit 1
            ;;
    esac
    shift
done

# Main script logic to handle arguments
case "${COMMAND}" in
    up)
        start_localnet
        ;;
    down)
        stop_localnet
        ;;
    *)
        echo "❌ Invalid argument: $COMMAND"
        echo "Usage: $0 [up|down] [--use-local-helm]"
        exit 1
        ;;
esac
