#!/bin/bash
# Script to manage PATH Localnet in Docker
# Usage: ./localnet.sh [up|down]

set -e

# Configuration
DOCKER_IMAGE="ghcr.io/buildwithgrove/path-localnet-env:latest"

# Define spinner animation frames
FRAMES=('⠋' '⠙' '⠚' '⠞' '⠖' '⠦' '⠴' '⠲' '⠳' '⠓')
FRAMES_COUNT=${#FRAMES[@]}

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

    # Start the container
    echo "📡 Starting PATH Localnet ..."
    if ! docker run \
        --name path-localnet \
        -v "$(pwd)":/app \
        -p 10350:10350 \
        -p 3070:3070 \
        --privileged \
        -d \
        ${DOCKER_IMAGE}; then
        
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

# Main script logic to handle arguments
case "${1:-up}" in
    up)
        start_localnet
        ;;
    down)
        stop_localnet
        ;;
    *)
        echo "❌ Invalid argument: $1"
        echo "Usage: $0 [up|down]"
        exit 1
        ;;
esac
