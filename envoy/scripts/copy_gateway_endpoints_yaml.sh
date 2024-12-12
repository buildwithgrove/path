#!/bin/bash

URL="https://raw.githubusercontent.com/buildwithgrove/path-auth-data-server/refs/heads/main/yaml/testdata/gateway-endpoints.example.yaml"

if command -v wget &> /dev/null; then
    wget -O ./local/path/envoy/.gateway-endpoints.yaml "$URL"
elif command -v powershell &> /dev/null; then
    powershell -Command "Invoke-WebRequest -Uri '$URL' -OutFile './local/path/envoy/.gateway-endpoints.yaml'"
else
    echo "Please install wget or use PowerShell to run this script."
    exit 1
fi
