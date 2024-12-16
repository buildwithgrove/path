#!/bin/bash

# This script is used to copy the envoy.template.yaml file to the envoy.yaml file
# and replace the sensitive auth variables with the values provided by the user.

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source file names
ENVOY_TEMPLATE_FILE_NAME="envoy.template.yaml"
ENVOY_RATELIMIT_TEMPLATE_FILE_NAME="ratelimit.template.yaml"

# Destination file names
ENVOY_FILE_NAME=".envoy.yaml"
ENVOY_RATELIMIT_FILE_NAME=".ratelimit.yaml"

# Define the absolute path for envoy.yaml
ENVOY_CONFIG_PATH="$SCRIPT_DIR/../../local/path/envoy/$ENVOY_FILE_NAME"

# Check if envoy.yaml exists and throw an error if it does
if [ -f "$ENVOY_CONFIG_PATH" ]; then
    echo "Error: $ENVOY_CONFIG_PATH already exists."
    exit 1
fi

# Function to prompt the user for OAuth usage
prompt_oauth_usage() {
    while true; do
        echo "Configure JWT Authentication:"
        echo "1. Enable JWT Auth (requires an OAuth provider like Auth0, along with domain and audience details)"
        echo "2. Disable JWT Auth"
        echo "Select an option (1 or 2): " 
        read -p "> " USE_OAUTH
        if [[ "$USE_OAUTH" == "1" ]]; then
            return 0
        elif [[ "$USE_OAUTH" == "2" ]]; then
            return 1
        else
            echo "Invalid selection. Please enter '1' or '2'."
        fi
    done
}

# Prompt the user if they wish to use an OAuth provider
if prompt_oauth_usage; then
    # Prompt for AUTH_DOMAIN
    echo "Enter AUTH_DOMAIN: This is the domain of your OAuth provider, where the authorization server is hosted."
    echo "Example: 'auth.example.com'"
    read -p "> " AUTH_DOMAIN

    # Prompt for AUTH_AUDIENCE
    echo "Enter AUTH_AUDIENCE: This is the intended audience for the token, typically the identifier of the API or service that will consume the token."
    echo "Example: 'https://auth.example.com/oauth/token'"
    read -p "> " AUTH_AUDIENCE

    # Substitute sensitive variables manually using bash parameter expansion
    sed -e "s|\${AUTH_DOMAIN}|$AUTH_DOMAIN|g" \
        -e "s|\${AUTH_AUDIENCE}|$AUTH_AUDIENCE|g" \
        "$SCRIPT_DIR/../$ENVOY_TEMPLATE_FILE_NAME" > "$ENVOY_CONFIG_PATH"
else
    # Just copy the file without substitution if the user does not wish to use JWT authorization
    cp "$SCRIPT_DIR/../$ENVOY_TEMPLATE_FILE_NAME" "$ENVOY_CONFIG_PATH"

    # Use yq to remove specific YAML blocks related to JWT authentication
    yq eval 'del(.static_resources.clusters[] | select(.name == "auth_jwks_cluster"))' -i "$ENVOY_CONFIG_PATH"
    yq eval 'del(.static_resources.listeners[].filter_chains[].filters[].typed_config.http_filters[] | select(.name == "envoy.filters.http.jwt_authn"))' -i "$ENVOY_CONFIG_PATH"
    yq eval 'del(.static_resources.listeners[].filter_chains[].filters[].typed_config.http_filters[] | select(.name == "envoy.filters.http.header_mutation"))' -i "$ENVOY_CONFIG_PATH"
fi

echo "$ENVOY_FILE_NAME has been created at $ENVOY_CONFIG_PATH"

# Define the absolute path for ratelimit.yaml
RATELIMIT_CONFIG_PATH="$SCRIPT_DIR/../../local/path/envoy/$ENVOY_RATELIMIT_FILE_NAME"

cp "$SCRIPT_DIR/../$ENVOY_RATELIMIT_TEMPLATE_FILE_NAME" "$RATELIMIT_CONFIG_PATH"

echo "$ENVOY_RATELIMIT_FILE_NAME has been created at $RATELIMIT_CONFIG_PATH"
