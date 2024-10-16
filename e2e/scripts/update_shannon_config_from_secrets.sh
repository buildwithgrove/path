#!/usr/bin/env bash

# This script updates the Shannon E2E config file from environment variables.
# It is used in GitHub actions to run the CI, and the environment variables
# are populated from repo's secrets.

# Set the current working directory to e2e config directory.
cd "$(dirname "$0")/.." || exit 1

update_shannon_config_from_env() {
    if [[ -z "$GATEWAY_PRIVATE_KEY" ]]; then
        echo " GATEWAY_PRIVATE_KEY environment variable not set"
        return 1
    fi

    local CONFIG_FILE="./.shannon.config.yaml"
    if [[ ! -f $CONFIG_FILE ]]; then
        echo "config file" $CONFIG_FILE "not found in" $PWD
	return 1
    fi

    sed  -i 's/gateway_private_key:.*/gateway_private_key: '"$GATEWAY_PRIVATE_KEY"'/' $CONFIG_FILE
}

update_shannon_config_from_env "$@"
