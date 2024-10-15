#!/bin/bash

# Check and prompt for AUTH0_DOMAIN if not set
if [ -z "$AUTH0_DOMAIN" ]; then
    read -p "Enter AUTH0_DOMAIN (eg. 'auth.example.com'): " AUTH0_DOMAIN
    export AUTH0_DOMAIN
fi

# Check and prompt for AUTH0_AUDIENCE if not set
if [ -z "$AUTH0_AUDIENCE" ]; then
    read -p "Enter AUTH0_AUDIENCE (eg. 'https://auth.example.com/oauth/token'): " AUTH0_AUDIENCE
    export AUTH0_AUDIENCE
fi

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Use envsubst to substitute variables in the template
envsubst < "$SCRIPT_DIR/../envoy.template.yaml" > "$SCRIPT_DIR/../envoy.yaml"
