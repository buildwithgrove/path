#!/usr/bin/env bash

# This script installs the pocketd binary if not already installed.
# It logs each step and provides a basic explanation of how functions work via comments.

# Function to check if a command exists on the system
# This function takes a single argument (the command name), checks if it's available, and returns 0 if found, 1 if not.
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install pocketd if not present
# This function checks if pocketd is installed. If not, it downloads the correct binary, extracts it, makes it executable, and verifies with 'pocketd version'.
install_pocketd() {
    if command_exists pocketd; then
        echo "pocketd already installed." | tee -a install.log
    else
        echo "Installing pocketd..." | tee -a install.log
        OS=$(uname | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
        TARBALL="poktroll_${OS}_${ARCH}.tar.gz"
        curl -LO "https://github.com/pokt-network/poktroll/releases/latest/download/${TARBALL}"
        sudo tar -zxf "${TARBALL}" -C /usr/local/bin
        sudo chmod +x /usr/local/bin/pocketd
        echo "pocketd installation complete. Checking version..." | tee -a install.log
        pocketd version
        echo "pocketd version check complete." | tee -a install.log
    fi
}

# Main execution starts here
echo "Starting pocketd installation script..." | tee -a install.log
install_pocketd
echo "pocketd installation script completed." | tee -a install.log
