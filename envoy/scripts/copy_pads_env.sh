#!/bin/bash

URL="https://raw.githubusercontent.com/buildwithgrove/path-auth-data-server/refs/heads/main/.env.example"

if command -v wget &> /dev/null; then
    wget -O ./local/path/envoy/.env.pads "$URL"
elif command -v powershell &> /dev/null; then
    powershell -Command "Invoke-WebRequest -Uri '$URL' -OutFile './local/path/envoy/.env.pads'"
else
    echo "Please install wget or use PowerShell to run this script."
    exit 1
fi
