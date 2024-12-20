#!/bin/bash

# URL of the gateway-endpoints.example.yaml file in the PADS repo
URL="https://raw.githubusercontent.com/buildwithgrove/path-auth-data-server/refs/heads/main/yaml/testdata/gateway-endpoints.example.yaml"

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Destination file path
GATEWAY_CONFIG_FILE_NAME=".gateway-endpoints.yaml"
BASE_ENVOY_PATH=$(realpath "$SCRIPT_DIR/../../local/path/envoy")
GATEWAY_CONFIG_FILE="$BASE_ENVOY_PATH/$GATEWAY_CONFIG_FILE_NAME"

# Check if the gateway-endpoints.yaml file already exists
if [ -f "$GATEWAY_CONFIG_FILE" ]; then
    echo "💡 $GATEWAY_CONFIG_FILE already exists, not overwriting."
    exit 0
fi

# Download the file using wget or PowerShell
if command -v wget &>/dev/null; then
    wget -O "$GATEWAY_CONFIG_FILE" "$URL"
elif command -v powershell &>/dev/null; then
    powershell -Command "Invoke-WebRequest -Uri '$URL' -OutFile '$GATEWAY_CONFIG_FILE'"
else
    echo "Please install wget or use PowerShell to run this script."
    exit 1
fi

echo "✅ $GATEWAY_CONFIG_FILE has been created"
echo "📄 README: Please update this file with your own data."
