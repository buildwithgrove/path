#!/usr/bin/env bash
set -e
set -o nounset

# This script updates the Shannon E2E config file from environment variables.
# It is used in GitHub actions to run the CI, and the environment variables
# are populated from repo's secrets.

# Set the current working directory to e2e/config directory.
cd "$(dirname "$0")/../config" || exit 1

update_shannon_config_from_env() {
    check_env_vars "SHANNON_GATEWAY_PRIVATE_KEY" "SHANNON_OWNED_APPS_PRIVATE_KEYS"

    local CONFIG_FILE="./.shannon.config.yaml"
    if [[ ! -f $CONFIG_FILE ]]; then
        echo "config file" $CONFIG_FILE "not found in" $PWD
        return 1
    fi

    # Update the PATH Shannon config to reflect secrets on GitHub.
    yq -i '
	.shannon_config.gateway_config.gateway_private_key_hex = env(SHANNON_GATEWAY_PRIVATE_KEY) |
	.shannon_config.gateway_config.owned_apps_private_keys_hex = (env(SHANNON_OWNED_APPS_PRIVATE_KEYS) | split(",")) |
    .shannon_config.full_node_config.lazy_mode = true
    ' $CONFIG_FILE
}

# check_env_vars verifies that all the input arguments are environment variables with non-empty values.
check_env_vars() {
    for env_var in "$@"; do
        if [[ -z "${!env_var}" ]]; then
            echo " $env_var environment variable not set"
            return 1
        fi
    done
}

update_shannon_config_from_env "$@"
