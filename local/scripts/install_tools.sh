#!/usr/bin/env bash

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

install_pocketd() {
    if command_exists pocketd; then
        log "INFO" "Pocketd already installed."
        pocketd version
        return
    fi

    log "INFO" "Installing Pocketd..."

    curl -sSL https://raw.githubusercontent.com/pokt-network/poktroll/main/tools/scripts/pocketd-install.sh | bash

    log "SUCCESS" "Pocketd installed successfully."
    pocketd version

    log "Fore more information, see https://dev.poktroll.com/explore/account_management/pocketd_cli"
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

# Check for missing dependencies
MISSING_DEPS=()

for cmd in docker pocketd; do
    if ! command_exists "$cmd"; then
        case "$cmd" in
            docker) MISSING_DEPS+=("🐳 Docker: Container engine for running applications in containers") ;;
            pocketd) MISSING_DEPS+=("🔧 pocketd: CLI tool for interacting with Pocket Network") ;;
        esac
    fi
done

if [ ${#MISSING_DEPS[@]} -eq 0 ]; then
    log "SUCCESS" "✅ All dependencies are installed."
    exit 0
fi

# Display missing dependencies
log "WARNING" "🚨 The following required dependencies are missing:"
for dep in "${MISSING_DEPS[@]}"; do
    echo -e "${YELLOW}${dep}${RESET}"
done

# Prompt user to install
if ! prompt_user "❔ Would you like to install these dependencies? (y/n):"; then
    log "WARNING" "Installation aborted by user"
    exit 1
fi

# Install missing dependencies
install_docker
install_pocketd

log "SUCCESS" "✅ Installation script completed."
