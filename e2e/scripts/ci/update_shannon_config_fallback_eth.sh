#!/usr/bin/env bash
set -e
set -o nounset

# Modifies ./e2e/config/.shannon.config.yaml for ETH fallback testing
# PRIMARY USE: GitHub Actions E2E tests with external fallback providers
# KEY FUNCTIONS:
# • Configures ETH fallback settings in Shannon config
# • Enables fallback endpoints for ETH service
# • Routes traffic to external fallback providers
# LOCAL USAGE REQUIREMENTS:
# • Set SHANNON_ETH_FALLBACK_URL environment variable
# • Config file will redirect all traffic to fallback URL

# Set the current working directory to e2e/config directory.
cd "$(dirname "$0")/../../config" || exit 1

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
