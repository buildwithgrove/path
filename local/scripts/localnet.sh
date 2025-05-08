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

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

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
        echo -e "${RED}❌ Directory does not exist: $input_path${NC}"
        echo -e "${WHITE}Would you like to create it? [y/N]${NC}"
        read -p "> " create_dir
        if [[ "${create_dir,,}" == "y" || "${create_dir,,}" == "yes" ]]; then
            mkdir -p "$input_path"
            echo -e "${GREEN}✅ Created directory: $input_path${NC}"
        else
            echo -e "${YELLOW}❌ Using remote helm charts instead${NC}"
            return 1
        fi
    fi
    
    # Set the environment variable with the normalized path
    export LOCAL_HELM_CHARTS_PATH="$input_path"
    echo -e "  ✅ Using local helm charts from: ${BLUE}$LOCAL_HELM_CHARTS_PATH${NC}"
    return 0
}

# Function to run a spinner with a timeout
run_spinner() {
    local timeout=$1
    local check_url=$2
    local message=$3
    local success_message=$4
    
    echo -e "\n${WHITE}⏳ $message${NC}"
    
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
    echo -e "${RED}❌ Timed out waiting for $message${NC}"
    return 1
}

# Function to show a simple spinner for a process
show_spinner() {
    local message=$1
    local cmd=$2
    local success_message=$3
    
    echo -e "\n${WHITE}⏳ $message${NC}"
    
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

# Function to check container status
check_container_status() {
    # Wait briefly to allow container to start and run validation
    echo -e "  ${WHITE}🔍 Starting container and checking validation...${NC}"
    sleep 1
    
    # Check if container is running
    if docker ps --format '{{.Names}}' | grep -q "^path-localnet$"; then
        # Container is running, validation passed
        echo -e "  ${GREEN}✓ Container validation passed${NC}"
        return 0
    else
        # Container exited, check the exit code
        exit_code=$(docker inspect path-localnet --format='{{.State.ExitCode}}')
        if [ "$exit_code" != "0" ]; then
            echo -e "${RED}❌ PATH Localnet container exited with code $exit_code. See logs below:${NC}"
            docker logs path-localnet
            exit 1
        fi
        # If we get here, container exited with code 0, which is unexpected
        echo -e "${YELLOW}⚠️ Container exited with code 0 unexpectedly${NC}"
        return 1
    fi
}

# Function to start Docker container with local helm charts
run_with_local_helm_charts() {
    local helm_charts_path=$1
    
    echo -e "  ${WHITE}📦 Mounting local helm charts from ${helm_charts_path}${NC}"
    
    if ! docker run \
        --name path-localnet \
        -v "$(pwd)":/app \
        -p 10350:10350 \
        -p 3070:3070 \
        --privileged \
        -v "${helm_charts_path}":/helm-charts \
        -e LOCAL_HELM_CHARTS_PATH=/helm-charts \
        -d \
        ${DOCKER_IMAGE}; then
        echo -e "${RED}❌ Failed to start Docker container. Check if ports 10350 and 3070 are available.${NC}"
        exit 1
    fi

    # Check container status
    check_container_status
}

# Function to start Docker container without local helm charts
run_without_local_helm_charts() {
    echo -e "  ${WHITE}📡 Using remote helm charts${NC}"
    
    if ! docker run \
        --name path-localnet \
        -v "$(pwd)":/app \
        -p 10350:10350 \
        -p 3070:3070 \
        --privileged \
        -e LOCAL_HELM_CHARTS_PATH= \
        -d \
        ${DOCKER_IMAGE}; then
        echo -e "${RED}❌ Failed to start Docker container. Check if ports 10350 and 3070 are available.${NC}"
        exit 1
    fi

    # Check container status
    check_container_status
}

# Function to start up PATH Localnet
start_localnet() {
    # Start the container based on whether we should use local helm charts
    echo -e "${BLUE}🍃 Starting PATH Localnet ...${NC}"
    
    if [ "$USE_LOCAL_HELM" = true ]; then
        echo -e "${WHITE}🔍 Running PATH Localnet with local helm charts ...${NC}"
        if prompt_for_local_helm_charts; then
            if [ ! -d "${LOCAL_HELM_CHARTS_PATH}" ]; then
                echo -e "${RED}❌ Error: LOCAL_HELM_CHARTS_PATH directory does not exist: ${LOCAL_HELM_CHARTS_PATH}${NC}"
                exit 1
            fi
            run_with_local_helm_charts "${LOCAL_HELM_CHARTS_PATH}"
        else
            echo -e "${YELLOW}⚠️ Failed to set up local helm charts, reverting to remote charts${NC}"
            run_without_local_helm_charts
        fi
    else
        run_without_local_helm_charts
    fi

    # First, wait for Tilt UI to become available
    if run_spinner 180 "http://localhost:10350" "Waiting for Tilt UI to become available..." "Tilt UI is now available!"; then
        echo -e "  ${WHITE}✅ Access Tilt UI at: ${CYAN}http://localhost:10350${NC}"
        
        # Next, wait for the healthz endpoint
        if run_spinner 300 "http://localhost:3070/healthz" "Waiting for PATH API to be ready..." "PATH API is ready!"; then
            echo -e "\n${GREEN}🌿 PATH Localnet started successfully.${NC}"
            echo -e "  ${WHITE}🚀 Send relay requests to: ${CYAN}http://localhost:3070/v1${NC}"
            exit 0
        else
            echo -e "  ${YELLOW}Check logs with: docker logs path-localnet${NC}"
            exit 1
        fi
    else
        echo -e "  ${YELLOW}Check logs with: docker logs path-localnet${NC}"
        exit 1
    fi
}

# Function to stop PATH Localnet
stop_localnet() {
    # Check if Docker is installed
    if ! command -v docker >/dev/null 2>&1; then
        echo -e "${RED}❌ Docker is not installed. Make sure you review README.md before continuing${NC}"
        exit 1
    fi

    # Stop and remove the container with spinner
    if show_spinner "Stopping path-localnet container" "docker stop path-localnet > /dev/null 2>&1 || true" "Container stopped"; then
        if show_spinner "Removing path-localnet container" "docker rm path-localnet > /dev/null 2>&1 || true" "Container removed"; then
            echo -e "\n${GREEN}✅ PATH Localnet has been successfully stopped and removed.${NC}"
            exit 0
        fi
    fi
    
    echo -e "${YELLOW}⚠️ There might have been issues stopping PATH Localnet.${NC}"
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
            echo -e "${RED}❌ Unknown option: $1${NC}"
            echo -e "${WHITE}Usage: $0 [up|down] [--use-local-helm]${NC}"
            exit 1
            ;;
    esac
    shift
done

# Check if container already exists
if docker ps -a --format '{{.Names}}' | grep -q "^path-localnet$"; then
    # Check if container is running
    if docker ps --format '{{.Names}}' | grep -q "^path-localnet$"; then
        echo -e "${RED}❌ Error: path-localnet is already running.${NC}"
        echo -e "${WHITE}To stop it, run: ${BLUE}make path_down${NC}"
        exit 1
    else
        echo -e "${YELLOW}🧹 Removing stopped path-localnet container...${NC}"
        docker rm path-localnet > /dev/null 2>&1 || true
    fi
fi

# Check for required config file
if [ ! -f "./local/path/.config.yaml" ]; then
    echo -e "\n${RED}❌ Error: ./local/path/.config.yaml not found. Ensure you have a valid config YAML file at this location.${NC}\n"
    echo -e "  💡 For information about the PATH config YAML file, see the documentation at: "
    echo -e "       ${CYAN}https://path.grove.city/develop/path/configurations_path${NC} "
    echo -e "\n  🌿 Grove employees: you may find a valid ${BLUE}.config.yaml${NC} file on 1Password in the note called ${BLUE}\"PATH Localnet Config\"${NC}\n"
    exit 1
fi

# Check for optional values file
if [ ! -f "./local/path/.values.yaml" ]; then
    echo -e "\n${BLUE}ℹ️  Info: ./local/path/.values.yaml not found. PATH Localnet will use the default Helm Chart values.${NC}\n"
    echo -e "  💡 For information about the PATH values YAML file, see the documentation at: "
    echo -e "       ${CYAN}https://path.grove.city/develop/path/configurations_helm${NC}  "
    echo -e "\n  🌿 Grove employees: you may find a valid ${BLUE}.values.yaml${NC} file on 1Password in the note called ${BLUE}\"PATH Localnet Config\"${NC}\n"
fi

# Main script logic to handle arguments
case "${COMMAND}" in
    up)
        start_localnet
        ;;
    down)
        stop_localnet
        ;;
    *)
        echo -e "${RED}❌ Invalid argument: $COMMAND${NC}"
        echo -e "${WHITE}Usage: $0 [up|down] [--use-local-helm]${NC}"
        exit 1
        ;;
esac
