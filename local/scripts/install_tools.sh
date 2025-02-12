#!/usr/bin/env bash

# This script installs Docker, Kind, Helm, and Tilt if they are not already installed.
# It logs each step and provides a basic explanation of how functions work via comments.

# Function to check if a command exists on the system
# This function takes a single argument (the command name), checks if it's available, and returns 0 if found, 1 if not.
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Docker if not present
# This function checks if Docker is installed. If not, it downloads and runs the official installation script.
install_docker() {
    if command_exists docker; then
        echo "$(date) - Docker already installed." >>install.log
    else
        echo "$(date) - Installing Docker..." >>install.log
        curl -fsSL https://get.docker.com -o get-docker.sh
        sh get-docker.sh
        rm -f get-docker.sh
        echo "$(date) - Docker installation complete." >>install.log
    fi
}

# Function to install Kind if not present
# This function checks if Kind is installed. If not, it downloads the binary and moves it to /usr/local/bin.
install_kind() {
    if command_exists kind; then
        echo "$(date) - Kind already installed." >>install.log
    else
        echo "$(date) - Installing Kind..." >>install.log
        KIND_VERSION=$(curl -s https://api.github.com/repos/kubernetes-sigs/kind/releases/latest | grep tag_name | cut -d '"' -f4)
        curl -Lo ./kind "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-amd64"
        chmod +x kind
        mv kind /usr/local/bin/kind
        echo "$(date) - Kind installation complete." >>install.log
    fi
}

# Function to install kubectl if not present
install_kubectl() {
    if command_exists kubectl; then
        echo "$(date) - kubectl already installed." >>install.log
    else
        echo "$(date) - Installing kubectl..." >>install.log
        KUBECTL_VERSION=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
        curl -LO "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
        chmod +x kubectl
        mv kubectl /usr/local/bin/kubectl
        echo "$(date) - kubectl installation complete." >>install.log
    fi
}

# Function to install Helm if not present
# This function checks if Helm is installed. If not, it uses the Helm install script to get the latest version.
install_helm() {
    if command_exists helm; then
        echo "$(date) - Helm already installed." >>install.log
    else
        echo "$(date) - Installing Helm..." >>install.log
        curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
        echo "$(date) - Helm installation complete." >>install.log
    fi
}

# Function to install Tilt if not present
# This function checks if Tilt is installed. If not, it runs the Tilt install script.
install_tilt() {
    if command_exists tilt; then
        echo "$(date) - Tilt already installed." >>install.log
    else
        echo "$(date) - Installing Tilt..." >>install.log
        curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
        echo "$(date) - Tilt installation complete." >>install.log
    fi
}

# Function to install Graphviz if not present
# This function checks if Graphviz is installed. If not, it installs it using the package manager.
install_graphviz() {
    if command_exists dot; then
        echo "$(date) - Graphviz already installed." >>install.log
    else
        echo "$(date) - Visit this link to install Graphviz manually: https://graphviz.org"
        echo "$(date) - This is optional and only needed for debugging purposes."
    fi
}

# Main execution starts here
echo "$(date) - Starting installation script..." >>install.log

install_docker
install_kind
install_kubectl
install_helm
install_tilt
install_graphviz

echo "$(date) - Installation script completed." >>install.log
