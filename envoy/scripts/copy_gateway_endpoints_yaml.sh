#!/bin/bash

# URL of the gateway-endpoints.example.yaml file in the PADS repo
URL="https://raw.githubusercontent.com/buildwithgrove/path-auth-data-server/refs/heads/main/yaml/testdata/gateway-endpoints.example.yaml"

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Destination file path
DESTINATION=$(realpath "$SCRIPT_DIR/../../local/path/envoy/.gateway-endpoints.yaml")

# Check if the gateway-endpoints.yaml file already exists
if [ -f "$DESTINATION" ]; then
    echo "💡 $DESTINATION already exists, not overwriting."
    exit 0
fi

# Download the file using wget or PowerShell
if command -v wget &> /dev/null; then
    wget -O "$DESTINATION" "$URL"
elif command -v powershell &> /dev/null; then
    powershell -Command "Invoke-WebRequest -Uri '$URL' -OutFile '$DESTINATION'"
else
    echo "Please install wget or use PowerShell to run this script."
    exit 1
fi

echo "✅ $DESTINATION has been created"
echo "📄 README: Please update this file with your own data."
