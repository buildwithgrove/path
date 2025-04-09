#!/usr/bin/env bash
set -e
set -o nounset

# This script updates the Morse E2E config file from environment variables.
# It is used in GitHub actions to run the CI, and the environment variables
# are populated from repo's secrets.

# Set the current working directory to e2e config directory.
cd "$(dirname "$0")/.." || exit 1

update_morse_config_from_env() {
    check_env_vars "MORSE_GATEWAY_SIGNING_KEY" "MORSE_FULLNODE_URL" "MORSE_AATS"

    local CONFIG_FILE="./.morse.config.yaml"
    if [[ ! -f $CONFIG_FILE ]]; then
        echo "config file" $CONFIG_FILE "not found in" $PWD
	return 1
    fi

    # Update the PATH Shannon config to reflect secrets on GitHub.
    # Also set the hydrator config to run 20 checks to ensure invalid endpoints are sanctioned.
    yq -i '
    .hydrator_config.bootstrap_initial_qos_data_checks = 20 |
	.morse_config.full_node_config.relay_signing_key = env(MORSE_GATEWAY_SIGNING_KEY) |
	.morse_config.full_node_config.url = env(MORSE_FULLNODE_URL) |
	.morse_config.signed_aats = env(MORSE_AATS)
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

update_morse_config_from_env "$@"
