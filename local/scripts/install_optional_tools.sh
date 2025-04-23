#!/usr/bin/env bash

# This script installs optional tools for PATH development:
# - Relay Util: Load testing tool for sending configurable batches of relays concurrently
# - Graphviz: Required for generating profiling & debugging performance
# - Uber Mockgen: Mock interface generator for testing
# It detects the OS and architecture to download the correct binaries.

set -e

# Terminal colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

# Function to log messages to file and console
log() {
    local level="$1"
    local message="$2"
    local color="$RESET"
    
    case "$level" in
        "INFO") color="$BLUE" ;;
        "SUCCESS") color="$GREEN" ;;
        "WARNING") color="$YELLOW" ;;
        "ERROR") color="$RED" ;;
    esac
    
    echo -e "${color}${message}${RESET}"
}

# Function to check if a command exists on the system
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to detect system architecture and OS
detect_system() {
    # Detect OS type
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    # Detect architecture
    ARCH="$(uname -m)"
    
    # Normalize architecture naming
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac
    
    # Set system type
    if [ "$OS" = "darwin" ] && [ "$ARCH" = "amd64" ]; then
        SYSTEM="mac_x86"
    elif [ "$OS" = "darwin" ] && [ "$ARCH" = "arm64" ]; then
        SYSTEM="mac_arm"
    elif [ "$OS" = "linux" ] && [ "$ARCH" = "amd64" ]; then
        SYSTEM="linux_x86"
    elif [ "$OS" = "linux" ] && [ "$ARCH" = "arm64" ]; then
        SYSTEM="linux_arm"
    else
        SYSTEM="unknown"
        log "WARNING" "Unsupported system: $OS $ARCH"
    fi
    
    log "INFO" "Detected system: $OS $ARCH (System type: $SYSTEM)"
}

# Function to install Relay Util if not present
install_relay_util() {
    if command_exists relay-util; then
        log "INFO" "🚚 Relay Util already installed."
        return
    fi
    
    if ! command_exists go; then
        log "WARNING" "🚨 Go is not installed. In order to install Relay Util, please install Go from https://go.dev/doc/install"
        return
    fi
    
    log "INFO" "🚚 Installing Relay Util..."
    
    go install github.com/commoddity/relay-util/v2@latest
    
    log "SUCCESS" "✅ Relay Util installed successfully."
}

# Function to install Graphviz if not present
install_graphviz() {
    if command_exists dot; then
        log "INFO" "📊 Graphviz already installed."
        dot -V
        return
    fi
    
    log "INFO" "📊 Installing Graphviz..."
    
    case "$SYSTEM" in
        mac_x86|mac_arm)
            if ! command_exists brew; then
                log "WARNING" "🚨 Homebrew is missing. Please install Homebrew first or install Graphviz manually: https://graphviz.org/download/"
                return 1
            fi
            brew install graphviz
            ;;
        linux_x86|linux_arm)
            sudo apt-get update
            sudo apt-get install -y graphviz
            ;;
        *)
            log "ERROR" "Unsupported system for Graphviz installation. Please install manually: https://graphviz.org/download/"
            return 1
            ;;
    esac
    
    log "SUCCESS" "✅ Graphviz installed successfully."
    dot -V
}

# Function to install Uber Mockgen if not present
install_mockgen() {
    if command_exists mockgen; then
        log "INFO" "🧪 Mockgen already installed."
        mockgen -version
        return
    fi
    
    if ! command_exists go; then
        log "WARNING" "🚨 Go is not installed. In order to install Mockgen, please install Go from https://go.dev/doc/install"
        return
    fi
    
    log "INFO" "🧪 Installing Uber Mockgen..."
    
    go install github.com/uber-go/mock/mockgen@latest
    
    log "SUCCESS" "✅ Mockgen installed successfully."
    mockgen -version
}

# Function to prompt user for confirmation
prompt_user() {
    local message="$1"
    echo -e "${BLUE}${message}${RESET}"
    echo -n "> "
    read -r answer
    answer=$(echo "$answer" | tr '[:upper:]' '[:lower:]')
    if [[ "$answer" =~ ^(y|yes)$ ]]; then
        return 0
    else
        return 1
    fi
}

# Main execution starts here
log "INFO" "🔍 Starting optional tools installation script..."

# Detect system architecture and OS
detect_system

# Check for missing dependencies
MISSING_DEPS=()

for cmd in relay-util dot mockgen; do
    if ! command_exists "$cmd"; then
        case "$cmd" in
            relay-util) MISSING_DEPS+=("🚚 Relay Util: Simple load-testing tool for relays") ;;
            dot) MISSING_DEPS+=("📊 Graphviz (dot): Required for generating profiling & debugging performance") ;;
            mockgen) MISSING_DEPS+=("🧪 Uber Mockgen: Mock interface generator for testing") ;;
        esac
    fi
done

if [ ${#MISSING_DEPS[@]} -eq 0 ]; then
    log "SUCCESS" "✅ All optional dependencies are already installed."
    exit 0
fi

# Display missing dependencies
log "WARNING" "🚨 The following optional dependencies are missing:"
for dep in "${MISSING_DEPS[@]}"; do
    echo -e "${YELLOW}${dep}${RESET}"
done

# Prompt user to install
if ! prompt_user "❔ Would you like to install these optional dependencies? (y/n):"; then
    log "WARNING" "Installation aborted by user"
    exit 1
fi

# Install missing dependencies
if ! command_exists relay-util; then
    install_relay_util
fi

if ! command_exists dot; then
    install_graphviz
fi

if ! command_exists mockgen; then
    install_mockgen
fi

log "SUCCESS" "✅ Optional tools installation script completed."
