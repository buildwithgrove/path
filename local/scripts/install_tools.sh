#!/usr/bin/env bash

# This script installs Docker, Kind, Kubectl, Helm, Tilt, and pocketd if they are not already installed.
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

# Function to install Docker if not present
install_docker() {
    if command_exists docker; then
        log "INFO" "🐳 Docker already installed."
        docker --version
        return
    fi

    log "INFO" "🐳 Installing Docker..."

    case "$SYSTEM" in
        mac_x86|mac_arm)
            if ! command_exists brew; then
                log "WARNING" "🚨 Docker not found and Homebrew is missing. Please install Docker Desktop manually from https://www.docker.com/products/docker-desktop"
                return 1
            fi
            brew install --cask docker
            ;;
        linux_x86|linux_arm)
            curl -fsSL https://get.docker.com -o get-docker.sh
            sudo sh get-docker.sh
            rm get-docker.sh
            if [ -e "/var/run/docker.sock" ]; then
                sudo chmod 666 /var/run/docker.sock
            fi
            # Check if Docker daemon is running
            if ! pgrep -x dockerd >/dev/null; then
                log "WARNING" "Docker daemon not running. Attempting to start dockerd..."
                sudo systemctl start docker || sudo dockerd &
                sleep 3
                if ! pgrep -x dockerd >/dev/null; then
                    log "ERROR" "Docker daemon did not start correctly"
                    return 1
                fi
                log "SUCCESS" "✅ Docker daemon started successfully."
            fi
            ;;
        *)
            log "ERROR" "Unsupported system for Docker installation"
            return 1
            ;;
    esac

    log "SUCCESS" "✅ Docker installed successfully."
    docker --version
}

# Function to install Kind if not present
install_kind() {
    if command_exists kind; then
        log "INFO" "🌀 Kind already installed."
        kind --version
        return
    fi

    log "INFO" "🌀 Installing Kind..."

    # Try to get the latest version
    KIND_VERSION=$(curl -s https://api.github.com/repos/kubernetes-sigs/kind/releases/latest | grep tag_name | cut -d '"' -f4)
    if [ -z "$KIND_VERSION" ]; then
        KIND_VERSION="v0.27.0"  # Fallback to a known version
    fi

    # Create the binary name based on OS and architecture
    BINARY_NAME="kind-${OS}-${ARCH}"
    KIND_URL="https://kind.sigs.k8s.io/dl/${KIND_VERSION}/${BINARY_NAME}"

    curl -Lo /tmp/kind "$KIND_URL"
    chmod +x /tmp/kind
    sudo mv /tmp/kind /usr/local/bin/kind

    log "SUCCESS" "✅ Kind installed successfully."
    kind --version
}

# Function to install kubectl if not present
install_kubectl() {
    if command_exists kubectl; then
        log "INFO" "🔧 kubectl already installed."
        kubectl version --client
        return
    fi

    log "INFO" "🔧 Installing kubectl..."

    # Get stable kubectl version
    KUBECTL_VERSION=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
    if [ -z "$KUBECTL_VERSION" ]; then
        KUBECTL_VERSION="latest"
    fi

    curl -LO "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/${OS}/${ARCH}/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/kubectl

    log "SUCCESS" "✅ kubectl installed successfully."
    kubectl version --client
}

# Function to install Helm if not present
install_helm() {
    if command_exists helm; then
        log "INFO" "⛵ Helm already installed."
        helm version --short
        return
    fi

    log "INFO" "⛵ Installing Helm..."

    curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

    log "SUCCESS" "✅ Helm installed successfully."
    helm version --short
}

# Function to install Tilt if not present
install_tilt() {
    if command_exists tilt; then
        log "INFO" "🚀 Tilt already installed."
        tilt version
        return
    fi

    log "INFO" "🚀 Installing Tilt..."

    # Create ~/.local/bin if it doesn't exist
    mkdir -p "$HOME/.local/bin"

    curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash

    log "SUCCESS" "✅ Tilt installed successfully."
    tilt version
}

# Function to install pocketd (always install to ensure latest version)
install_pocketd() {
    if command_exists pocketd; then
        log "WARNING" "🌟 pocketd already installed. Overwriting with latest version..."
        pocketd version
    else
        log "INFO" "🌟 Installing pocketd..."
    fi

    # Run the official pocketd installation script
    curl -sSL https://raw.githubusercontent.com/pokt-network/poktroll/main/tools/scripts/pocketd-install.sh | bash -s -- --upgrade

    if command_exists pocketd; then
        log "SUCCESS" "✅ pocketd installed successfully."
        pocketd version
    else
        log "WARNING" "pocketd installation completed but pocketd command not found. You may need to restart your terminal or update your PATH."
    fi
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
log "INFO" "🔍 Starting installation script..."

# Detect system architecture and OS
detect_system

# Check for missing dependencies (pocketd is always installed)
MISSING_DEPS=()

for cmd in docker kind kubectl helm tilt pocketd; do
    if ! command_exists "$cmd"; then
        case "$cmd" in
            docker) MISSING_DEPS+=("🐳 Docker: Container engine for running applications in containers") ;;
            kind) MISSING_DEPS+=("🌀 Kind: Tool for running local Kubernetes clusters using Docker") ;;
            kubectl) MISSING_DEPS+=("🔧 kubectl: CLI tool for controlling Kubernetes clusters") ;;
            helm) MISSING_DEPS+=("⛵ Helm: Package manager for Kubernetes") ;;
            tilt) MISSING_DEPS+=("🚀 Tilt: Tool for development on Kubernetes") ;;
        esac
    fi
done

# Always check pocketd status for user information
if command_exists pocketd; then
    POCKETD_STATUS="🌟 pocketd: Pocket Network Protocol daemon (will be updated to latest version)"
else
    POCKETD_STATUS="🌟 pocketd: Pocket Network Protocol daemon (will be installed)"
fi

if [ ${#MISSING_DEPS[@]} -eq 0 ]; then
    log "SUCCESS" "✅ All core dependencies are installed."
    log "INFO" "$POCKETD_STATUS"
else
    # Display missing dependencies
    log "WARNING" "🚨 The following required dependencies are missing:"
    for dep in "${MISSING_DEPS[@]}"; do
        echo -e "${YELLOW}${dep}${RESET}"
    done
    log "INFO" "$POCKETD_STATUS"
fi

# Always prompt for installation if there are missing deps OR for pocketd update
if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
    PROMPT_MESSAGE="❔ Would you like to install the missing dependencies and update pocketd? (y/n):"
else
    PROMPT_MESSAGE="❔ Would you like to update pocketd to the latest version? (y/n):"
fi

if ! prompt_user "$PROMPT_MESSAGE"; then
    log "WARNING" "Installation aborted by user"
    exit 1
fi

# Install missing dependencies
install_docker
install_kind
install_kubectl
install_helm
install_tilt
install_pocketd

log "SUCCESS" "✅ Installation script completed."