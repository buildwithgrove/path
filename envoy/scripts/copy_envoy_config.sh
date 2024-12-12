#!/bin/bash

# This script is used to copy the envoy.template.yaml file to the envoy.yaml file
# and replace the sensitive auth variables with the values provided by the user.

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Define the absolute path for envoy.yaml
ENVOY_CONFIG_PATH="$SCRIPT_DIR/../../local/path/envoy/.envoy.yaml"

# Check if envoy.yaml exists and throw an error if it does
if [ -f "$ENVOY_CONFIG_PATH" ]; then
    echo "Error: $ENVOY_CONFIG_PATH already exists."
    exit 1
fi

# Prompt for AUTH_DOMAIN
read -p "Enter AUTH_DOMAIN (eg. 'auth.example.com'): " AUTH_DOMAIN

# Prompt for AUTH_AUDIENCE
read -p "Enter AUTH_AUDIENCE (eg. 'https://auth.example.com/oauth/token'): " AUTH_AUDIENCE

# Substitute sensitive variables manually using bash parameter expansion
sed -e "s|\${AUTH_DOMAIN}|$AUTH_DOMAIN|g" \
    -e "s|\${AUTH_AUDIENCE}|$AUTH_AUDIENCE|g" \
    "$SCRIPT_DIR/../envoy.template.yaml" > "$ENVOY_CONFIG_PATH"

echo "envoy.yaml has been created at $ENVOY_CONFIG_PATH"

# Define the absolute path for ratelimit.yaml
RATELIMIT_CONFIG_PATH="$SCRIPT_DIR/../../local/path/envoy/.ratelimit.yaml"

cp "$SCRIPT_DIR/../ratelimit.yaml" "$RATELIMIT_CONFIG_PATH"