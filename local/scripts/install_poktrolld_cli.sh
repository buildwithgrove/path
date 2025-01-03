#!/usr/bin/env bash

# This script installs the poktrolld binary if not already installed.
# It logs each step and provides a basic explanation of how functions work via comments.

# Function to check if a command exists on the system
# This function takes a single argument (the command name), checks if it's available, and returns 0 if found, 1 if not.
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install poktrolld if not present
# This function checks if poktrolld is installed. If not, it downloads the correct binary, extracts it, makes it executable, and verifies with 'poktrolld version'.
install_poktrolld() {
    if command_exists poktrolld; then
        echo "$(date) - poktrolld already installed." >> install.log
    else
        echo "$(date) - Installing poktrolld..." >> install.log
        OS=$(uname | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
        TARBALL="poktroll_${OS}_${ARCH}.tar.gz"
        curl -LO "https://github.com/pokt-network/poktroll/releases/latest/download/${TARBALL}"
        sudo tar -zxf "${TARBALL}" -C /usr/local/bin
        sudo chmod +x /usr/local/bin/poktrolld
        echo "$(date) - poktrolld installation complete. Checking version..." >> install.log
        poktrolld version
        echo "$(date) - poktrolld version check complete." >> install.log
    fi
}

# Main execution starts here
echo "$(date) - Starting poktrolld installation script..." >> install.log
install_poktrolld
echo "$(date) - poktrolld installation script completed." >> install.log