#!/usr/bin/env bash
set -e
set -o nounset

# IMPORTANT: This script will modify the `./e2e/config/.shannon.config.yaml` file,
# which is used in GitHub actions to enable fallback endpoints for ETH service
# when testing with external fallback providers.
#
# It used in GitHub actions to run E2E tests against external fallback providers.
# If run locally:
# 1. The `SHANNON_ETH_FALLBACK_URL` environment variable must be set to an eth fallback URL.
# 2. The `./e2e/config/.shannon.config.yaml` file will be modified to send all traffic to the `SHANNON_ETH_FALLBACK_URL`

# This script configures ETH fallback settings in the Shannon E2E config file.
# It is used in GitHub actions to enable fallback endpoints for ETH service
# when testing with external fallback providers.

# Set the current working directory to e2e/config directory.
cd "$(dirname "$0")/../config" || exit 1

update_shannon_eth_fallback_config() {
    check_env_vars "SHANNON_ETH_FALLBACK_URL"

    local CONFIG_FILE="./.shannon.config.yaml"
    if [[ ! -f $CONFIG_FILE ]]; then
        echo "config file" $CONFIG_FILE "not found in" $PWD
        return 1
    fi

    # Configure ETH service fallback settings
    yq -i '
	(.shannon_config.gateway_config.service_fallback[] | select(.service_id == "eth")).send_all_traffic = true |
	(.shannon_config.gateway_config.service_fallback[] | select(.service_id == "eth")).fallback_endpoints[0].default_url = env(SHANNON_ETH_FALLBACK_URL)
    ' $CONFIG_FILE
    
    echo "Successfully configured ETH fallback settings in $CONFIG_FILE"
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

update_shannon_eth_fallback_config "$@"
